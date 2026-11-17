package operatorapi_test

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

const revokePath = "/v1/cdr-certificates/revoke"

type fakeRevoker struct {
	cert   *x509.Certificate
	reason ca.Reason
	calls  int
	err    error
}

func (f *fakeRevoker) Revoke(_ context.Context, cert *x509.Certificate, reason ca.Reason) error {
	f.cert, f.reason = cert, reason
	f.calls++

	return f.err
}

func revoke(t *testing.T, revoker *fakeRevoker, header http.Header, body string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, &fakeOperator{}, &fakeIssuer{}, revoker, nil, &fakeKeys{}, access)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, revokePath, strings.NewReader(body))
	req.Header = header
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func revocationBody(t *testing.T, certificate, reason string) string {
	t.Helper()

	data, err := json.Marshal(map[string]string{"certificate": certificate, "reason": reason})
	require.NoError(t, err)

	return string(data)
}

// Verifies: REG-0703, REG-0731
func TestARevocationRevokesTheFirstCertificateWithTheReason(t *testing.T) {
	leaf, intermediate := selfSigned(t, "App"), selfSigned(t, "Intermediate")

	for name, pemText := range map[string]string{
		"a certificate":           string(ca.EncodeCertificates(leaf)),
		"the chain of an issue":   string(ca.EncodeCertificates(leaf, intermediate)),
		"a chain with whitespace": "\n" + string(ca.EncodeCertificates(leaf, intermediate)) + "\n\n",
	} {
		t.Run(name, func(t *testing.T) {
			revoker := &fakeRevoker{}

			rec := revoke(t, revoker, authorised(), revocationBody(t, pemText, "superseded"))

			require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			assert.Empty(t, rec.Body.String())
			assert.Equal(t, leaf.Raw, revoker.cert.Raw)
			assert.Equal(t, ca.Superseded, revoker.reason)
		})
	}
}

// Verifies: REG-0704
func TestARevocationBodyThatIsNotValidGives400AndNamesEachFieldThatFails(t *testing.T) {
	certificate := string(ca.EncodeCertificates(selfSigned(t, "App")))
	csr := func() string { _, data := csrPEM(t); return string(data) }()

	tests := map[string]struct {
		body   string
		fields []string
	}{
		"not JSON": {body: "certificate", fields: []string{"JSON"}},
		"an unknown field": {
			body:   `{"certificate": "x", "reason": "superseded", "serial": "1"}`,
			fields: []string{"JSON"},
		},
		"no certificate": {body: `{"reason": "superseded"}`, fields: []string{"certificate"}},
		"no reason":      {body: `{"certificate": "x"}`, fields: []string{"certificate", "reason"}},
		"not PEM": {
			body:   revocationBody(t, "not a certificate", "superseded"),
			fields: []string{"certificate"},
		},
		"a CSR": {body: revocationBody(t, csr, "superseded"), fields: []string{"certificate"}},
		"a certificate and a CSR": {
			body:   revocationBody(t, certificate+csr, "superseded"),
			fields: []string{"certificate"},
		},
		"a reason for a CA": {body: revocationBody(t, certificate, "cACompromise"), fields: []string{"reason"}},
		"both fields": {
			body:   revocationBody(t, "x", "removeFromCRL"),
			fields: []string{"certificate", "reason"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			revoker := &fakeRevoker{}

			rec := revoke(t, revoker, authorised(), tc.body)

			p := requireProblem(t, rec, http.StatusBadRequest)
			for _, field := range tc.fields {
				assert.Contains(t, p.Detail, field)
			}

			assert.Zero(t, revoker.calls)
		})
	}
}

// Verifies: REG-0854
func TestARevocationBodyLargerThan1MiBGives413(t *testing.T) {
	revoker := &fakeRevoker{}

	rec := revoke(t, revoker, authorised(), revocationBody(t, strings.Repeat("x", 1<<20), "superseded"))

	p := requireProblem(t, rec, http.StatusRequestEntityTooLarge)
	assert.Equal(t, "The request body is larger than 1 MiB.", p.Detail)
	assert.Zero(t, revoker.calls)
}

// Verifies: REG-0705, REG-0706
func TestTheResultOfARevocationGivesItsStatus(t *testing.T) {
	body := revocationBody(t, string(ca.EncodeCertificates(selfSigned(t, "App"))), "keyCompromise")

	tests := map[string]struct {
		err    error
		status int
		detail string
	}{
		"not issued by the intermediate": {
			err: app.ErrNotIssuedByTheIntermediate, status: http.StatusBadRequest, detail: "certificate",
		},
		"already revoked": {
			err: app.AlreadyRevokedError{Reason: ca.CessationOfOperation}, status: http.StatusConflict,
			detail: "cessationOfOperation",
		},
		"another error": {err: errors.New("database failure"), status: http.StatusInternalServerError},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := revoke(t, &fakeRevoker{err: tc.err}, authorised(), body)

			p := requireProblem(t, rec, tc.status)
			assert.Contains(t, p.Detail, tc.detail)
			assert.NotContains(t, p.Detail, "database")
		})
	}
}

func TestARevocationNeedsTheOperatorToken(t *testing.T) {
	revoker := &fakeRevoker{}
	body := revocationBody(t, string(ca.EncodeCertificates(selfSigned(t, "App"))), "keyCompromise")

	rec := revoke(t, revoker, nil, body)

	requireProblem(t, rec, http.StatusUnauthorized)
	assert.Zero(t, revoker.calls)
}

func TestARevocationIsAPostOnly(t *testing.T) {
	mux := http.NewServeMux()
	operatorapi.Register(mux, &fakeOperator{}, &fakeIssuer{}, &fakeRevoker{}, nil, &fakeKeys{}, access)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, revokePath, bytes.NewReader(nil))
	req.Header = authorised()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	requireProblem(t, rec, http.StatusNotFound)
}
