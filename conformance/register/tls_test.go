//go:build conformance

package register

import (
	"testing"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

func TestThePublicAPIsAcceptOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t *testing.T) {
	labtest.AssertAcceptsOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t, labtest.Env(t, "REGISTER_PUBLIC_URL"))
}

func TestThePublicAPIsPresentACertificateFromThePublicCAOnly(t *testing.T) {
	labtest.AssertPresentsACertificateFromThePublicCAOnly(t, labtest.Env(t, "REGISTER_PUBLIC_URL"))
}

func TestThePublicAPIsDoNotRequestAClientCertificate(t *testing.T) {
	labtest.AssertDoesNotRequestAClientCertificate(t, labtest.Env(t, "REGISTER_PUBLIC_URL"))
}
