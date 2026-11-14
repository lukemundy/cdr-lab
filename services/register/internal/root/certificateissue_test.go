package root

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

func clientCSRPEM(t *testing.T, cn, o string) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{Subject: pkix.Name{
		CommonName:         cn,
		Organization:       []string{o},
		OrganizationalUnit: []string{"Consumer Data Right"},
		Country:            []string{"AU"},
	}}, key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der})
}

// Data holder brands of seed/example.json.
const (
	inactiveHolderBrand = "1d344684-ca92-463d-af1e-4239ed438c18" // Mock Lending
	interimHolderBrand  = "d9530e60-cdc4-408c-a6f4-aba39e4158b6" // Mock Mutual, with no dataHolderBrandId
)

func holderBrandIssue(legalEntityID, brandID string) string {
	return "/v1/data-holders/" + legalEntityID + "/brands/" + brandID + "/certificates/issue"
}

func serverCSRPEM(t *testing.T, o, cn string, dnsNames ...string) (*rsa.PrivateKey, []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         cn,
			Organization:       []string{o},
			OrganizationalUnit: []string{"Consumer Data Right"},
			Country:            []string{"AU"},
		},
		DNSNames: dnsNames,
	}, key)
	require.NoError(t, err)

	return key, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der})
}

func TestTheOperatorGetsAServerCertificateThatATLSClientAcceptsForEachNameOfTheCSR(t *testing.T) {
	handlers, pool := labWithPool(t)
	names := []string{"mtls.mock-energy.example", "mock-energy", "localhost"}
	key, csr := serverCSRPEM(t, "Mock Energy", names[0], names[1:]...)

	rec := serveOperator(t, handlers, holderBrandIssue(seedDataHolder, seedHolderBrand), authorisedOperator(), csr)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/pem-certificate-chain", rec.Header().Get("Content-Type"))

	chain, err := ca.ReadCertificates(writeTemp(t, rec.Body.Bytes()))
	require.NoError(t, err)
	require.Len(t, chain, 2)

	bundle, err := ca.ReadCertificates(os.Getenv("CDR_CA_BUNDLE_FILE"))
	require.NoError(t, err)
	assert.Equal(t, bundle[0].Raw, chain[1].Raw, "the intermediate follows the certificate")
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, chain[0].ExtKeyUsage)
	assert.Equal(t, names, chain[0].DNSNames)
	assert.Equal(t, []string{testCRLURL}, chain[0].CRLDistributionPoints)

	roots := x509.NewCertPool()
	for _, c := range bundle {
		roots.AddCert(c)
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.TLS = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{chain[0].Raw, chain[1].Raw}, PrivateKey: key}},
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	for _, name := range names {
		dialer := &tls.Dialer{Config: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots, ServerName: name}}
		conn, err := dialer.DialContext(t.Context(), "tcp", server.Listener.Addr().String())
		require.NoError(t, err, name)
		require.NoError(t, conn.Close())
	}

	entries, _, err := readmodel.NewQueries(pool).History(t.Context(), "data-holder", app.HistoryPageQuery{
		LegalEntityID: seedDataHolder, Page: 1, PageSize: 1,
	})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "ServerCertificateIssued", entries[0].EventType)
	assert.Equal(t, &app.HistoryBrand{ID: seedHolderBrand, Name: "Mock Energy"}, entries[0].Brand)

	sum := sha256.Sum256(chain[0].Raw)
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(sum[:]), entries[0].Thumbprint)
}

func TestTheOperatorGetsAServerCertificateWithTheLegalEntityNameOrForAnInactiveBrand(t *testing.T) {
	handlers := lab(t)

	for brand, o := range map[string]string{seedHolderBrand: "Mock Holdings Ltd", inactiveHolderBrand: "Mock Lending"} {
		_, csr := serverCSRPEM(t, o, "mtls.mock-energy.example")

		rec := serveOperator(t, handlers, holderBrandIssue(seedDataHolder, brand), authorisedOperator(), csr)

		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestTheOperatorCannotGetAServerCertificateForACSRThatBreaksARuleOrForABrandThatCannotUseIt(t *testing.T) {
	handlers := lab(t)
	_, csr := serverCSRPEM(t, "Mock Energy", "mtls.mock-energy.example")
	_, badNames := serverCSRPEM(t, "Mock Group", "*.mock-energy.example", "10.0.0.1")

	tests := map[string]struct {
		holder, brand string
		body          []byte
		wantStatus    int
		wantDetail    string
	}{
		"not a CSR": {seedDataHolder, seedHolderBrand, []byte("not a CSR"), http.StatusBadRequest, "one PEM CSR"},
		"names that are not host names, and other O": {
			seedDataHolder, seedHolderBrand, badNames, http.StatusBadRequest,
			`subject CN "*.mock-energy.example" is not a host name: use only letters, digits and hyphens in labels ` +
				`of 1 to 63 characters, with a dot between two labels and no hyphen at the start or end of a label; ` +
				`DNS name "10.0.0.1" is an IP address, not a host name; ` +
				`subject O must be the brand name "Mock Energy" or the legal entity name "Mock Holdings Ltd"`,
		},
		"unknown brand":   {seedDataHolder, unknownID, csr, http.StatusNotFound, "does not exist"},
		"unknown holder":  {unknownID, seedHolderBrand, csr, http.StatusNotFound, "does not exist"},
		"removed brand":   {removedDataHolder, removedHolderBrand, csr, http.StatusConflict, "REMOVED is final"},
		"interim ID only": {seedDataHolder, interimHolderBrand, csr, http.StatusConflict, "no dataHolderBrandId"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			status, p := operate(t, handlers, holderBrandIssue(tt.holder, tt.brand), tt.body)

			assert.Equal(t, tt.wantStatus, status)
			assert.Contains(t, p.Detail, tt.wantDetail)
		})
	}
}

func TestTheOperatorGetsARegisteredClientCertificateFromACSR(t *testing.T) {
	handlers, pool := labWithPool(t)

	rec := serveOperator(
		t,
		handlers,
		productAction(mockBudgetRecipient, mockBudgetWeb, "certificates/issue"),
		authorisedOperator(),
		clientCSRPEM(t, "Mock Budget Web", "Mock Budget"),
	)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/pem-certificate-chain", rec.Header().Get("Content-Type"))

	chain, err := ca.ReadCertificates(writeTemp(t, rec.Body.Bytes()))
	require.NoError(t, err)
	require.Len(t, chain, 2)

	bundle, err := ca.ReadCertificates(os.Getenv("CDR_CA_BUNDLE_FILE"))
	require.NoError(t, err)
	assert.Equal(t, bundle[0].Raw, chain[1].Raw, "the intermediate follows the certificate")

	roots := x509.NewCertPool()
	roots.AddCert(bundle[0])
	_, err = chain[0].Verify(
		x509.VerifyOptions{Roots: roots, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}},
	)
	require.NoError(t, err)
	assert.Equal(t, []string{testCRLURL}, chain[0].CRLDistributionPoints)

	client, err := readmodel.NewQueries(pool).SoftwareProductClient(t.Context(), mockBudgetWeb)
	require.NoError(t, err)

	sum := sha256.Sum256(chain[0].Raw)
	assert.Contains(t, client.ClientCertificates, base64.RawURLEncoding.EncodeToString(sum[:]),
		"the token endpoint accepts the certificate")
}

func TestTheOperatorCannotGetACertificateForACSRThatBreaksARuleOrForARecordThatCannotUseIt(t *testing.T) {
	handlers := lab(t)
	csr := clientCSRPEM(t, "Mock Budget Web", "Mock Budget")

	tests := map[string]struct {
		product    string
		body       []byte
		wantStatus int
		wantDetail string
	}{
		"not a CSR": {mockBudgetWeb, []byte("not a CSR"), http.StatusBadRequest, "one PEM CSR"},
		"other names": {
			mockBudgetWeb,
			clientCSRPEM(t, "Mock Budget Mobile", "Mock Budget Pty Ltd"),
			http.StatusBadRequest,
			`subject CN must be the software product name "Mock Budget Web"; ` +
				`subject O must be the brand name "Mock Budget"`,
		},
		"unknown software product": {unknownID, csr, http.StatusNotFound, "does not exist"},
		"removed software product": {mockBudgetBeta, csr, http.StatusConflict, "REMOVED is final"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			status, p := operate(
				t,
				handlers,
				productAction(mockBudgetRecipient, tt.product, "certificates/issue"),
				tt.body,
			)

			assert.Equal(t, tt.wantStatus, status)
			assert.Contains(t, p.Detail, tt.wantDetail)
		})
	}
}

func TestTheRegisterDoesNotStartWithoutTheMockCDRIntermediate(t *testing.T) {
	for _, name := range []string{
		"CDR_INTERMEDIATE_CERT_FILE",
		"CDR_INTERMEDIATE_KEY_FILE",
		"CDR_INTERMEDIATE_CRL_URL",
	} {
		t.Run(name, func(t *testing.T) {
			setLabSettings(t)
			t.Setenv(name, "")

			_, err := loadConfig(serveMode)

			assert.ErrorContains(t, err, name)
		})
	}
}

func TestTheRegisterDoesNotStartWithAnIntermediateThatIsNotInTheCDRCABundle(t *testing.T) {
	setLabSettings(t)

	otherCert, otherKey := tlsFiles(t)
	t.Setenv("CDR_INTERMEDIATE_CERT_FILE", otherCert)
	t.Setenv("CDR_INTERMEDIATE_KEY_FILE", otherKey)

	_, err := loadConfig(serveMode)

	assert.ErrorContains(t, err, "is not in the CDR CA bundle")
}

func writeTemp(t *testing.T, data []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(path, data, 0o600))

	return path
}
