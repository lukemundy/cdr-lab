package operatorapi_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// issueRoute is the route of one certificate kind, with the IDs of its path.
type issueRoute struct {
	path, kind, legalEntityID, recordID, participant string
}

var issueRoutes = map[string]issueRoute{
	"client": {
		"/v1/data-recipients/dr-1/software-products/sp-1/certificates/issue", "client", "dr-1", "sp-1",
		"data recipient",
	},
	"server": {"/v1/data-holders/dh-1/brands/dhb-1/certificates/issue", "server", "dh-1", "dhb-1", "data holder"},
}

type fakeIssuer struct {
	kind, legalEntityID, recordID string
	csr                           *x509.CertificateRequest
	chain                         []*x509.Certificate
	err                           error
}

func (f *fakeIssuer) IssueClient(
	_ context.Context,
	legalEntityID, softwareProductID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	f.kind, f.legalEntityID, f.recordID, f.csr = "client", legalEntityID, softwareProductID, csr
	return f.chain, f.err
}

func (f *fakeIssuer) IssueServer(
	_ context.Context,
	legalEntityID, brandID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	f.kind, f.legalEntityID, f.recordID, f.csr = "server", legalEntityID, brandID, csr
	return f.chain, f.err
}

func issue(t *testing.T, issuer *fakeIssuer, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, &fakeOperator{}, issuer, &fakeRevoker{}, nil, &fakeKeys{}, access)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, bytes.NewReader(body))
	req.Header = authorised()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func csrPEM(t *testing.T) (*x509.CertificateRequest, []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.CreateCertificateRequest(
		rand.Reader,
		&x509.CertificateRequest{Subject: pkix.Name{CommonName: "App"}},
		key,
	)
	require.NoError(t, err)
	csr, err := x509.ParseCertificateRequest(der)
	require.NoError(t, err)

	return csr, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der})
}

func selfSigned(t *testing.T, cn string) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: cn}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return cert
}

// Verifies: REG-0612
func TestAnIssueGivesTheCertificateAndTheIntermediateInPEM(t *testing.T) {
	for name, route := range issueRoutes {
		t.Run(name, func(t *testing.T) {
			csr, body := csrPEM(t)
			leaf, intermediate := selfSigned(t, "App"), selfSigned(t, "Intermediate")
			issuer := &fakeIssuer{chain: []*x509.Certificate{leaf, intermediate}}

			rec := issue(t, issuer, route.path, body)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "application/pem-certificate-chain", rec.Header().Get("Content-Type"))
			assert.Equal(t, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf.Raw}))+
				string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: intermediate.Raw})), rec.Body.String())
			assert.Equal(t, route.kind, issuer.kind)
			assert.Equal(t, route.legalEntityID, issuer.legalEntityID)
			assert.Equal(t, route.recordID, issuer.recordID)
			assert.Equal(t, csr.Raw, issuer.csr.Raw)
		})
	}
}

// Verifies: REG-0602
func TestAnIssueRefusesABodyThatIsNotOnePEMCSR(t *testing.T) {
	_, one := csrPEM(t)

	for routeName, route := range issueRoutes {
		for name, body := range map[string][]byte{
			"empty":     nil,
			"two CSRs":  append(append([]byte{}, one...), one...),
			"too large": append(append([]byte{}, one...), bytes.Repeat([]byte(" "), 16<<10)...),
		} {
			t.Run(routeName+"/"+name, func(t *testing.T) {
				issuer := &fakeIssuer{}

				p := requireProblem(t, issue(t, issuer, route.path, body), http.StatusBadRequest)

				assert.Equal(t, "The request body must be one PEM CSR.", p.Detail)
				assert.Nil(t, issuer.csr, "the issuer gets no request")
			})
		}
	}
}

// Verifies: REG-0605
func TestAnIssueGivesTheRuleThatRefusesIt(t *testing.T) {
	ruleErr := statusRuleError(t)
	rulesErr := domain.CSRRulesError{Rules: []string{"subject C is required", `subject O must be the brand name "App"`}}

	tests := map[string]struct {
		err        error
		status     int
		wantDetail string
	}{
		"CSR rules":   {rulesErr, http.StatusBadRequest, rulesErr.Error()},
		"status rule": {ruleErr, http.StatusConflict, ruleErr.Error()},
		"not found": {
			fmt.Errorf("software product %w", domain.ErrNotFound),
			http.StatusNotFound,
			"software product does not exist",
		},
	}
	for routeName, route := range issueRoutes {
		for name, tt := range tests {
			t.Run(routeName+"/"+name, func(t *testing.T) {
				_, body := csrPEM(t)

				p := requireProblem(t, issue(t, &fakeIssuer{err: tt.err}, route.path, body), tt.status)

				assert.Equal(t, tt.wantDetail, p.Detail)
			})
		}

		t.Run(routeName+"/version conflict", func(t *testing.T) {
			_, body := csrPEM(t)

			p := requireProblem(
				t,
				issue(t, &fakeIssuer{err: eventsourcing.ErrVersionConflict}, route.path, body),
				http.StatusConflict,
			)

			assert.Equal(t, "The "+route.participant+" changed during the request. Send the request again.", p.Detail)
		})
	}
}

func TestAnIssueWithoutTheOperatorTokenIsUnauthorised(t *testing.T) {
	for name, route := range issueRoutes {
		t.Run(name, func(t *testing.T) {
			_, body := csrPEM(t)
			issuer := &fakeIssuer{}
			mux := http.NewServeMux()
			operatorapi.Register(mux, &fakeOperator{}, issuer, &fakeRevoker{}, nil, &fakeKeys{}, access)

			req := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				route.path,
				strings.NewReader(string(body)),
			)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Nil(t, issuer.csr)
		})
	}
}
