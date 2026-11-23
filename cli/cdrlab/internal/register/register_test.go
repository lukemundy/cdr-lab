package register_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register/registertest"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func csr(t *testing.T, kind ca.Kind, names ...string) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, ca.LeafKeyBits)
	require.NoError(t, err)

	data, err := ca.NewCSR(key, kind, "wattle-bank", "Wattle Bank", names)
	require.NoError(t, err)

	return data
}

func TestCABundleGivesTheIntermediateAndTheRoot(t *testing.T) {
	fake := registertest.New(t)

	bundle, err := register.New(fake.URL, registertest.Token, fake.Client()).CABundle(t.Context())

	require.NoError(t, err)
	require.Len(t, bundle, 2)
	assert.Equal(t, fake.Intermediate.Certificate.Raw, bundle[0].Raw)
	assert.Equal(t, fake.Root.Certificate.Raw, bundle[1].Raw)
}

// Verifies: CLI-0030
func TestIssueServerSendsTheCSRToTheIssuePathOfTheBrand(t *testing.T) {
	fake := registertest.New(t)
	request := csr(t, ca.Server, "wattle-bank")

	chain, err := register.New(fake.URL, registertest.Token, fake.Client()).
		IssueServer(t.Context(), "le-1", "brand-1", request)

	require.NoError(t, err)
	require.Len(t, chain, 2)
	assert.Equal(t, "wattle-bank", chain[0].Subject.CommonName)
	assert.Equal(t, []string{"POST /v1/data-holders/le-1/brands/brand-1/certificates/issue"}, fake.Issues())
	assert.Equal(t, request, fake.LastCSR())
}

// Verifies: CLI-0030
func TestIssueClientSendsTheCSRToTheIssuePathOfTheSoftwareProduct(t *testing.T) {
	fake := registertest.New(t)

	chain, err := register.New(fake.URL, registertest.Token, fake.Client()).
		IssueClient(t.Context(), "le-2", "sp-2", csr(t, ca.Client))

	require.NoError(t, err)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, chain[0].ExtKeyUsage)
	assert.Equal(t, []string{"POST /v1/data-recipients/le-2/software-products/sp-2/certificates/issue"}, fake.Issues())
}

// Verifies: CLI-0038
func TestARefusalGivesTheStatusAndTheDetailOfTheProblem(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(
			w,
			`{"type":"about:blank","title":"Bad Request","status":400,"detail":"O must be Wattle Bank."}`,
		)
	}))
	t.Cleanup(srv.Close)

	_, err := register.New(srv.URL, registertest.Token, srv.Client()).
		IssueServer(t.Context(), "le-1", "brand-1", csr(t, ca.Server))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "400 Bad Request")
	assert.Contains(t, err.Error(), "O must be Wattle Bank.")
}

// Verifies: CLI-0060
func TestAnErrorDoesNotGiveTheToken(t *testing.T) {
	fake := registertest.New(t)

	_, err := register.New(fake.URL, "wrong-token", fake.Client()).CABundle(t.Context())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.NotContains(t, err.Error(), "wrong-token")
}

func TestAResponseThatIsNotACertificateChainIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "<html>not ready</html>")
	}))
	t.Cleanup(srv.Close)

	_, err := register.New(srv.URL, registertest.Token, srv.Client()).CABundle(t.Context())

	require.Error(t, err)
}
