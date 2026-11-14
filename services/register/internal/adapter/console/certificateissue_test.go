package console_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type issueRoute struct {
	path, kind, legalEntityID, recordID, notFound string
}

var issueRoutes = map[string]issueRoute{
	"client": {
		"/api/data-recipients/dr-1/software-products/sp-1/certificates/issue", "client", "dr-1", "sp-1",
		"The Register has no data recipient with this legal entity ID, " +
			"or the data recipient has no software product with this ID.",
	},
	"server": {
		"/api/data-holders/dh-1/brands/dhb-1/certificates/issue", "server", "dh-1", "dhb-1",
		"The Register has no data holder with this legal entity ID, or the data holder has no brand with this ID.",
	},
}

type fakeIssuer struct {
	kind, legalEntityID, recordID string
	version                       int64
	csr                           *x509.CertificateRequest
	chain                         []*x509.Certificate
	err                           error
}

func (f *fakeIssuer) IssueClientAt(
	_ context.Context,
	legalEntityID string,
	version int64,
	softwareProductID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	f.kind, f.legalEntityID, f.version, f.recordID, f.csr = "client", legalEntityID, version, softwareProductID, csr
	return f.chain, f.err
}

func (f *fakeIssuer) IssueServerAt(
	_ context.Context,
	legalEntityID string,
	version int64,
	brandID string,
	csr *x509.CertificateRequest,
) ([]*x509.Certificate, error) {
	f.kind, f.legalEntityID, f.version, f.recordID, f.csr = "server", legalEntityID, version, brandID, csr
	return f.chain, f.err
}

func csrPEM(t *testing.T) (*x509.CertificateRequest, string) {
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

	return csr, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der}))
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

func statusRuleError(t *testing.T) error {
	t.Helper()

	dr := recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientSurrendered)
	state := domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{DataRecipient: dr})
	_, err := domain.HandleDataRecipient(state, domain.ReinstateAccreditation{})
	require.ErrorAs(t, err, new(domain.StatusRuleError))

	return err
}

func issueRequest(t *testing.T, version int64, csr string) string {
	t.Helper()

	body, err := json.Marshal(map[string]any{"version": version, "csr": csr})
	require.NoError(t, err)

	return string(body)
}

func postIssue(t *testing.T, issuer *fakeIssuer, target, body string) response {
	t.Helper()

	return post(t, console.Ports{Issuer: issuer}, target, body)
}

func TestAnIssueGivesTheChainAndTheThumbprintOfTheCertificate(t *testing.T) {
	for name, route := range issueRoutes {
		t.Run(name, func(t *testing.T) {
			csr, csrText := csrPEM(t)
			leaf, intermediate := selfSigned(t, "App"), selfSigned(t, "Intermediate")
			issuer := &fakeIssuer{chain: []*x509.Certificate{leaf, intermediate}}

			resp := postIssue(t, issuer, route.path, issueRequest(t, 3, csrText))

			require.Equal(t, http.StatusOK, resp.status, resp.body)
			assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))

			var body struct {
				CertificateChain string `json:"certificateChain"`
				Thumbprint       string `json:"thumbprint"`
			}
			require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

			sum := sha256.Sum256(leaf.Raw)
			assert.Equal(t, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf.Raw}))+
				string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: intermediate.Raw})),
				body.CertificateChain)
			assert.Equal(t, base64.RawURLEncoding.EncodeToString(sum[:]), body.Thumbprint)

			assert.Equal(t, route.kind, issuer.kind)
			assert.Equal(t, route.legalEntityID, issuer.legalEntityID)
			assert.Equal(t, route.recordID, issuer.recordID)
			assert.Equal(t, int64(3), issuer.version)
			assert.Equal(t, csr.Raw, issuer.csr.Raw)
		})
	}
}

func TestAnIssueRefusesABodyThatIsNotTheVersionAndTheCSR(t *testing.T) {
	_, csrText := csrPEM(t)

	const want = "The request body must be a JSON object with only the version of the participant, " +
		"an integer of 1 or more, and the PEM CSR as a string."

	for routeName, route := range issueRoutes {
		for name, body := range map[string]string{
			"empty":         "",
			"not JSON":      "csr",
			"no version":    `{"csr": "x"}`,
			"version 0":     issueRequest(t, 0, csrText),
			"no CSR":        `{"version": 1}`,
			"unknown field": `{"version": 1, "csr": "x", "other": 1}`,
			"too large":     issueRequest(t, 1, csrText+strings.Repeat(" ", 32<<10)),
		} {
			t.Run(routeName+"/"+name, func(t *testing.T) {
				issuer := &fakeIssuer{}

				resp := postIssue(t, issuer, route.path, body)

				require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
				assert.Equal(t, want, problemDetail(t, resp))
				assert.Nil(t, issuer.csr, "the issuer gets no request")
			})
		}
	}
}

func TestAnIssueRefusesACSRThatIsNotOnePEMCSR(t *testing.T) {
	_, one := csrPEM(t)
	certificate, _ := certificatePEM(t)

	for routeName, route := range issueRoutes {
		for name, csr := range map[string]string{
			"empty":       "",
			"text":        "not a CSR",
			"two CSRs":    one + one,
			"certificate": certificate,
		} {
			t.Run(routeName+"/"+name, func(t *testing.T) {
				issuer := &fakeIssuer{}

				resp := postIssue(t, issuer, route.path, issueRequest(t, 1, csr))

				require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
				assert.Equal(t, "The CSR must be one PEM CSR, with no other text.", problemDetail(t, resp))
				assert.Nil(t, issuer.csr)
			})
		}
	}
}

func TestAnIssueGivesTheRuleThatRefusesIt(t *testing.T) {
	rulesErr := domain.CSRRulesError{Rules: []string{"subject C is required", `subject O must be the brand name "App"`}}
	ruleErr := statusRuleError(t)

	tests := map[string]struct {
		err        error
		status     int
		wantDetail string
	}{
		"CSR rules":   {err: rulesErr, status: http.StatusBadRequest, wantDetail: rulesErr.Error()},
		"status rule": {err: ruleErr, status: http.StatusConflict, wantDetail: ruleErr.Error()},
		"stale version": {
			err:        eventsourcing.ErrVersionConflict,
			status:     http.StatusConflict,
			wantDetail: "The participant changed after the version in the request. The Register changed nothing.",
		},
	}
	for routeName, route := range issueRoutes {
		for name, tt := range tests {
			t.Run(routeName+"/"+name, func(t *testing.T) {
				_, csrText := csrPEM(t)

				resp := postIssue(t, &fakeIssuer{err: tt.err}, route.path, issueRequest(t, 1, csrText))

				require.Equal(t, tt.status, resp.status, resp.body)
				assert.Equal(t, tt.wantDetail, problemDetail(t, resp))
			})
		}
	}
}

func TestAnIssueForARecordThatDoesNotExistGivesNoIdentifier(t *testing.T) {
	for name, route := range issueRoutes {
		t.Run(name, func(t *testing.T) {
			_, csrText := csrPEM(t)
			issuer := &fakeIssuer{err: fmt.Errorf("record %q %w", route.recordID, domain.ErrNotFound)}

			resp := postIssue(t, issuer, route.path, issueRequest(t, 1, csrText))

			require.Equal(t, http.StatusNotFound, resp.status, resp.body)
			assert.Equal(t, route.notFound, problemDetail(t, resp))
		})
	}
}

func TestAnIssueThatFailsForAnUnknownReasonGivesNoDetailOfTheCause(t *testing.T) {
	for name, route := range issueRoutes {
		t.Run(name, func(t *testing.T) {
			_, csrText := csrPEM(t)

			resp := postIssue(t, &fakeIssuer{err: errors.New("disk full")}, route.path, issueRequest(t, 1, csrText))

			require.Equal(t, http.StatusInternalServerError, resp.status, resp.body)
			assert.NotContains(t, resp.body, "disk full")
		})
	}
}
