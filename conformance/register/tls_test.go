//go:build conformance

package register

import (
	"testing"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// Verifies: REG-0346
func TestThePublicAPIsAcceptOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t *testing.T) {
	labtest.AssertAcceptsOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t, labtest.Env(t, "REGISTER_PUBLIC_URL"))
}

// Verifies: REG-0349
func TestThePublicAPIsPresentACertificateFromThePublicCAOnly(t *testing.T) {
	labtest.AssertPresentsACertificateFromThePublicCAOnly(t, labtest.Env(t, "REGISTER_PUBLIC_URL"))
}

// Verifies: REG-0350
func TestThePublicAPIsDoNotRequestAClientCertificate(t *testing.T) {
	labtest.AssertDoesNotRequestAClientCertificate(t, labtest.Env(t, "REGISTER_PUBLIC_URL"))
}
