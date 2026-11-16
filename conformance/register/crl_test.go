//go:build conformance

package register

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// RFC 5280 reason codes.
const (
	keyCompromise        = 1
	cessationOfOperation = 5
)

func intermediate(t *testing.T) *x509.Certificate {
	t.Helper()

	data, err := os.ReadFile(labtest.Env(t, "PKI_CA_BUNDLE"))
	require.NoError(t, err)

	block, _ := pem.Decode(data)
	require.NotNil(t, block, "the CA bundle starts with the intermediate")
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	return cert
}

func leaf(t *testing.T, cert tls.Certificate) *x509.Certificate {
	t.Helper()

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)

	return parsed
}

// currentCRL gets the CRL at the one CRL URL of cert, as a relying party does (CDS: Certificate Validation).
func currentCRL(t *testing.T, cert *x509.Certificate) *x509.RevocationList {
	t.Helper()

	require.Len(t, cert.CRLDistributionPoints, 1, "a CDR certificate gives one CRL URL")

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, cert.CRLDistributionPoints[0], nil)
	require.NoError(t, err)

	resp, err := (&http.Client{Timeout: labtest.RequestTimeout}).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	der, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/pkix-crl", resp.Header.Get("Content-Type"))

	crl, err := x509.ParseRevocationList(der)
	require.NoError(t, err, "the CRL is DER")
	require.NoError(t, crl.CheckSignatureFrom(intermediate(t)))
	assert.Positive(t, crl.Number.Sign(), "the CRL has a CRL number")
	assert.Equal(t, 7*24*time.Hour, crl.NextUpdate.Sub(crl.ThisUpdate))
	assert.True(t, time.Now().Before(crl.NextUpdate), "the CRL is current")

	return crl
}

func entryOf(crl *x509.RevocationList, cert *x509.Certificate) *x509.RevocationListEntry {
	for i, entry := range crl.RevokedCertificateEntries {
		if entry.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			return &crl.RevokedCertificateEntries[i]
		}
	}

	return nil
}

// Verifies: PKI-0040, PKI-0056, PKI-0058, PKI-0060, PKI-0064
func TestTheCRLURLOfEachCDRCertificateGivesACurrentCRLOfTheIntermediate(t *testing.T) {
	for _, variables := range [][2]string{
		{"PKI_CLIENT_CERT", "PKI_CLIENT_KEY"},
		{"PKI_STATUS_CLIENT_CERT", "PKI_STATUS_CLIENT_KEY"},
		{"PKI_REVOKED_CLIENT_CERT", "PKI_REVOKED_CLIENT_KEY"},
	} {
		t.Run(variables[0], func(t *testing.T) {
			currentCRL(t, leaf(t, labtest.ClientCertificate(t, variables[0], variables[1])))
		})
	}
}

// Verifies: PKI-0053
func TestTheCRLListsTheRevokedFixtureWithItsReason(t *testing.T) {
	revoked := leaf(t, labtest.RevokedClientCertificate(t))

	entry := entryOf(currentCRL(t, revoked), revoked)

	require.NotNil(t, entry)
	assert.Equal(t, keyCompromise, entry.ReasonCode)
}
