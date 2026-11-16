package console_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

const (
	checkRevocation = "/api/cdr-certificates/check-revocation"
	revoke          = "/api/cdr-certificates/revoke"
	badRevocation   = "The request body must be a JSON object with only the PEM certificate and the reason, " +
		"as strings."
)

type fakeCA struct {
	cert        *x509.Certificate
	reason      ca.Reason
	revoked     bool
	checked     bool
	query       app.RevocationPageQuery
	revocations []app.Revocation
	total       int
	err         error
}

func (f *fakeCA) CheckRevocable(_ context.Context, cert *x509.Certificate) (app.RevocableCertificate, error) {
	f.cert, f.checked = cert, true
	if f.err != nil {
		return app.RevocableCertificate{}, f.err
	}

	return app.RevocableCertificate{
		Serial:     big.NewInt(0x0a1b),
		Subject:    "CN=App,O=Brand,C=AU",
		Thumbprint: "thumbprint",
		NotAfter:   time.Date(2027, 10, 27, 1, 2, 3, 0, time.UTC),
	}, nil
}

func (f *fakeCA) Revoke(_ context.Context, cert *x509.Certificate, reason ca.Reason) error {
	f.cert, f.reason = cert, reason
	if f.err != nil {
		return f.err
	}

	f.revoked = true

	return nil
}

func (f *fakeCA) Revocations(_ context.Context, query app.RevocationPageQuery) ([]app.Revocation, int, error) {
	f.query = query
	return f.revocations, f.total, f.err
}

func revocationRequest(t *testing.T, certificate, reason string) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{"certificate": certificate, "reason": reason})
	require.NoError(t, err)

	return string(body)
}

type fieldErrorsBody struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
	Errors []struct {
		Pointer string `json:"pointer"`
		Detail  string `json:"detail"`
	} `json:"errors"`
}

func fieldErrorsOf(t *testing.T, resp response) fieldErrorsBody {
	t.Helper()

	var body fieldErrorsBody
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body), resp.body)

	return body
}

// Verifies: REG-1313, REG-1337
func TestTheCABundleIsThePEMOfEachCertificateOfTheBundle(t *testing.T) {
	intermediate, root := selfSigned(t, "Intermediate"), selfSigned(t, "Root")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/cdr-ca-bundle", nil)

	resp := do(t, console.Ports{CABundle: []*x509.Certificate{intermediate, root}}, req)

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))

	var body struct {
		CABundle string `json:"caBundle"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	assert.Equal(t, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: intermediate.Raw}))+
		string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: root.Raw})), body.CABundle)
}

// Verifies: REG-1315, REG-1343
func TestACheckOfARevocationNamesTheFirstCertificateAndRevokesNothing(t *testing.T) {
	leaf, intermediate := selfSigned(t, "App"), selfSigned(t, "Intermediate")
	chain := string(ca.EncodeCertificates(leaf, intermediate))
	authority := &fakeCA{}

	resp := post(t, console.Ports{CertificateAuthority: authority}, checkRevocation,
		revocationRequest(t, chain, "keyCompromise"))

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{
		"serialNumber": "0A1B",
		"subject": "CN=App,O=Brand,C=AU",
		"thumbprint": "thumbprint",
		"notAfter": "2027-10-27T01:02:03Z"
	}`, resp.body)
	assert.Equal(t, leaf.Raw, authority.cert.Raw)
	assert.False(t, authority.revoked)
}

// Verifies: REG-1316, REG-1344
func TestARevocationRevokesTheFirstCertificateWithTheReason(t *testing.T) {
	leaf, intermediate := selfSigned(t, "App"), selfSigned(t, "Intermediate")
	authority := &fakeCA{}

	resp := post(t, console.Ports{CertificateAuthority: authority}, revoke,
		revocationRequest(t, string(ca.EncodeCertificates(leaf, intermediate)), "privilegeWithdrawn"))

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)
	assert.True(t, authority.revoked)
	assert.Equal(t, leaf.Raw, authority.cert.Raw)
	assert.Equal(t, ca.PrivilegeWithdrawn, authority.reason)
}

// Verifies: REG-1342
func TestARevocationAndItsCheckRefuseABodyThatIsNotACertificateAndAReason(t *testing.T) {
	cert, _ := certificatePEM(t)

	for _, path := range []string{checkRevocation, revoke} {
		for name, body := range map[string]string{
			"empty":          "",
			"not JSON":       "certificate",
			"a number":       `{"certificate": 1, "reason": "superseded"}`,
			"unknown field":  `{"certificate": "x", "reason": "superseded", "version": 1}`,
			"too large":      revocationRequest(t, cert+strings.Repeat(" ", 32<<10), "superseded"),
			"a JSON array":   `["superseded"]`,
			"a JSON string":  `"superseded"`,
			"two JSON items": `{} {}`,
		} {
			t.Run(path+"/"+name, func(t *testing.T) {
				authority := &fakeCA{}

				resp := post(t, console.Ports{CertificateAuthority: authority}, path, body)

				require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
				assert.Equal(t, badRevocation, problemDetail(t, resp))
				assert.Nil(t, authority.cert, "the CA gets no request")
			})
		}
	}
}

// Verifies: REG-1345
func TestARevocationAndItsCheckGiveAFieldErrorForEachValueThatIsNotValid(t *testing.T) {
	for _, path := range []string{checkRevocation, revoke} {
		t.Run(path, func(t *testing.T) {
			authority := &fakeCA{}

			resp := post(t, console.Ports{CertificateAuthority: authority}, path,
				revocationRequest(t, "not a certificate", "unknown"))

			require.Equal(t, http.StatusBadRequest, resp.status, resp.body)

			body := fieldErrorsOf(t, resp)
			require.Len(t, body.Errors, 2)
			assert.Equal(t, "#/certificate", body.Errors[0].Pointer)
			assert.Equal(t, app.NotACertificate, body.Errors[0].Detail)
			assert.Equal(t, "#/reason", body.Errors[1].Pointer)
			assert.Contains(t, body.Errors[1].Detail, "keyCompromise")
			assert.Equal(t, "One or more values in the request are not valid.", body.Detail)
			assert.Nil(t, authority.cert)

			resp = post(t, console.Ports{CertificateAuthority: authority}, path, `{"reason": "superseded"}`)

			require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
			body = fieldErrorsOf(t, resp)
			require.Len(t, body.Errors, 1)
			assert.Equal(t, "#/certificate", body.Errors[0].Pointer)
		})
	}
}

// Verifies: REG-1345, REG-1346
func TestARevocationAndItsCheckGiveTheRefusalOfTheCA(t *testing.T) {
	cert, _ := certificatePEM(t)

	for _, path := range []string{checkRevocation, revoke} {
		t.Run(path+"/not the intermediate", func(t *testing.T) {
			resp := post(t, console.Ports{CertificateAuthority: &fakeCA{err: app.ErrNotIssuedByTheIntermediate}},
				path, revocationRequest(t, cert, "superseded"))

			require.Equal(t, http.StatusBadRequest, resp.status, resp.body)

			body := fieldErrorsOf(t, resp)
			require.Len(t, body.Errors, 1)
			assert.Equal(t, "#/certificate", body.Errors[0].Pointer)
			assert.Equal(t, "The mock CDR intermediate did not sign the certificate.", body.Errors[0].Detail)
		})
		t.Run(path+"/already revoked", func(t *testing.T) {
			authority := &fakeCA{err: app.AlreadyRevokedError{Reason: ca.KeyCompromise}}

			resp := post(t, console.Ports{CertificateAuthority: authority}, path,
				revocationRequest(t, cert, "superseded"))

			require.Equal(t, http.StatusConflict, resp.status, resp.body)
			assert.NotContains(t, resp.body, `"errors"`)
			assert.Equal(t,
				"The certificate is already revoked, with the reason keyCompromise. The Register changed nothing.",
				problemDetail(t, resp))
		})
		t.Run(path+"/unknown", func(t *testing.T) {
			resp := post(t, console.Ports{CertificateAuthority: &fakeCA{err: errors.New("disk full")}}, path,
				revocationRequest(t, cert, "superseded"))

			require.Equal(t, http.StatusInternalServerError, resp.status, resp.body)
			assert.NotContains(t, resp.body, "disk full")
		})
	}
}

// Verifies: REG-1314, REG-1338, REG-1339, REG-1341
func TestTheRevocationsAreAPageWithTheTotal(t *testing.T) {
	at := time.Date(2026, 10, 27, 1, 2, 3, 0, time.FixedZone("AWST", 8*60*60))
	authority := &fakeCA{total: 30, revocations: []app.Revocation{{
		Serial:     big.NewInt(0xab),
		Subject:    "CN=App,O=Brand,C=AU",
		Thumbprint: "thumbprint",
		Time:       at,
		Reason:     ca.CessationOfOperation,
	}}}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet,
		"/api/cdr-certificates/revocations?page=2&pageSize=10", nil)

	resp := do(t, console.Ports{CertificateAuthority: authority}, req)

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
	assert.JSONEq(t, `{"totalRecords": 30, "data": [{
		"serialNumber": "AB",
		"subject": "CN=App,O=Brand,C=AU",
		"thumbprint": "thumbprint",
		"revokedAt": "2026-10-26T17:02:03Z",
		"reason": "cessationOfOperation"
	}]}`, resp.body)
	assert.Equal(t, app.RevocationPageQuery{Page: 2, PageSize: 10}, authority.query)
}

func TestTheRevocationsRefuseAPageThatIsNotValid(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/cdr-certificates/revocations?page=0", nil)

	resp := do(t, console.Ports{CertificateAuthority: &fakeCA{}}, req)

	assert.Equal(t, http.StatusBadRequest, resp.status, resp.body)
}

func TestTheRevocationsGiveNoDetailOfAnUnknownError(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/cdr-certificates/revocations", nil)

	resp := do(t, console.Ports{CertificateAuthority: &fakeCA{err: errors.New("disk full")}}, req)

	require.Equal(t, http.StatusInternalServerError, resp.status, resp.body)
	assert.NotContains(t, resp.body, "disk full")
}
