package operatorapi_test

import (
	"crypto/x509"
	"encoding/pem"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/operatorapi"
)

const caBundlePath = "/v1/cdr-ca-bundle"

func getCABundle(
	t *testing.T,
	bundle []*x509.Certificate,
	method string,
	header http.Header,
) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, &fakeOperator{}, &fakeIssuer{}, &fakeRevoker{}, bundle, &fakeKeys{}, access)

	req := httptest.NewRequestWithContext(t.Context(), method, caBundlePath, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// Verifies: REG-0616
func TestTheCABundleGivesEachCertificateOfTheBundleInPEMInTheSameOrder(t *testing.T) {
	intermediate, root := selfSigned(t, "Intermediate"), selfSigned(t, "Root")

	rec := getCABundle(t, []*x509.Certificate{intermediate, root}, http.MethodGet, authorised())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/pem-certificate-chain", rec.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: intermediate.Raw}))+
		string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: root.Raw})), rec.Body.String())
}

func TestTheCABundleNeedsTheOperatorToken(t *testing.T) {
	rec := getCABundle(t, []*x509.Certificate{selfSigned(t, "Root")}, http.MethodGet, nil)

	requireProblem(t, rec, http.StatusUnauthorized)
}

func TestTheCABundleRefusesAMethodThatIsNotGet(t *testing.T) {
	rec := getCABundle(t, []*x509.Certificate{selfSigned(t, "Root")}, http.MethodPost, authorised())

	requireProblem(t, rec, http.StatusNotFound)
}
