//go:build conformance

package register

import (
	"crypto/tls"
	"testing"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

func secureClient(t *testing.T) tls.Certificate {
	t.Helper()

	return labtest.ClientCertificate(t, "PKI_CLIENT_CERT", "PKI_CLIENT_KEY")
}

func TestTheSecureAPIsAcceptAClientCertificateFromTheCDRCA(t *testing.T) {
	labtest.AssertAcceptsAClientCertificateFromTheCDRCA(t, labtest.Env(t, "REGISTER_SECURE_URL"), secureClient(t))
}

func TestTheSecureAPIsRefuseAClientWithoutAValidCDRCertificate(t *testing.T) {
	labtest.AssertRefusesAClientWithoutAValidCDRCertificate(t, labtest.Env(t, "REGISTER_SECURE_URL"))
}

func TestTheSecureAPIsPresentACertificateFromTheCDRCAOnly(t *testing.T) {
	labtest.AssertPresentsACertificateFromTheCDRCAOnly(t, labtest.Env(t, "REGISTER_SECURE_URL"), secureClient(t))
}

func TestTheSecureAPIsAcceptOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t *testing.T) {
	labtest.AssertMutualTLSAcceptsOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(
		t,
		labtest.Env(t, "REGISTER_SECURE_URL"),
		secureClient(t),
	)
}
