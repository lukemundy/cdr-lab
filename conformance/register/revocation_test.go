//go:build conformance

package register

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// The Register signs the CRL, so a revocation needs no wait for a CRL that a relying party keeps.
func TestARevokedCertificateIsOnTheCRLAndTheRegisterRefusesItAtOnce(t *testing.T) {
	baseURL := labtest.Env(t, "REGISTER_SECURE_URL")
	client := labtest.ClientCertificate(t, "PKI_REVOCATION_CLIENT_CERT", "PKI_REVOCATION_CLIENT_KEY")
	labtest.AssertAcceptsAClientCertificateFromTheCDRCA(t, baseURL, client)

	labtest.Revoke(t, "PKI_REVOCATION_CLIENT_CERT", "cessationOfOperation")

	labtest.AssertRefusesTheClientCertificate(t, baseURL, client)

	cert := leaf(t, client)
	entry := entryOf(currentCRL(t, cert), cert)
	require.NotNil(t, entry)
	assert.Equal(t, cessationOfOperation, entry.ReasonCode)
}
