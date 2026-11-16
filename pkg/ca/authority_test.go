package ca_test

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

var now = time.Date(2026, 10, 2, 6, 30, 0, 0, time.UTC)

func newLab(t *testing.T) ca.Lab {
	t.Helper()

	lab, err := ca.NewLab(ca.LabKeys{CDRRoot: key(0), CDRIntermediate: key(1), PublicRoot: key(2)}, now)
	require.NoError(t, err)

	return lab
}

func verify(
	t *testing.T,
	cert *x509.Certificate,
	root *x509.Certificate,
	intermediates []*x509.Certificate,
	usage x509.ExtKeyUsage,
) error {
	t.Helper()

	roots, pool := x509.NewCertPool(), x509.NewCertPool()
	roots.AddCert(root)

	for _, c := range intermediates {
		pool.AddCert(c)
	}

	_, err := cert.Verify(
		x509.VerifyOptions{Roots: roots, Intermediates: pool, CurrentTime: now, KeyUsages: []x509.ExtKeyUsage{usage}},
	)

	return err
}

// Verifies: CA-0003, CA-0005, CA-0007
func TestTheCDRRootIsASelfSignedCAForTwentyYears(t *testing.T) {
	root := newLab(t).CDRRoot.Certificate

	assert.Equal(t, "CN=Mock CDR Root CA,OU=Consumer Data Right,O=cdr-lab,C=AU", root.Subject.String())
	assert.Equal(t, root.Subject.String(), root.Issuer.String())
	require.NoError(t, root.CheckSignatureFrom(root))
	assert.True(t, root.IsCA)
	assert.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, root.KeyUsage)
	assert.Equal(t, now, root.NotBefore)
	assert.Equal(t, now.AddDate(20, 0, 0), root.NotAfter)
}

// Verifies: CA-0003, CA-0004, CA-0006, CA-0007
func TestTheCDRIntermediateIssuesOnlyLeafCertificatesForTenYears(t *testing.T) {
	lab := newLab(t)
	intermediate := lab.CDRIntermediate.Certificate

	assert.Equal(t, "CN=Mock CDR Intermediate CA,OU=Consumer Data Right,O=cdr-lab,C=AU", intermediate.Subject.String())
	require.NoError(t, intermediate.CheckSignatureFrom(lab.CDRRoot.Certificate))
	assert.True(t, intermediate.IsCA)
	assert.True(t, intermediate.MaxPathLenZero)
	assert.Zero(t, intermediate.MaxPathLen)
	assert.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, intermediate.KeyUsage)
	assert.Equal(t, now.AddDate(10, 0, 0), intermediate.NotAfter)
}

// Verifies: CA-0009
func TestThePublicRootIsSeparateFromTheCDRChain(t *testing.T) {
	lab := newLab(t)
	public := lab.PublicRoot.Certificate

	assert.Equal(t, "CN=Mock Public Root CA,O=cdr-lab,C=AU", public.Subject.String())
	require.NoError(t, public.CheckSignatureFrom(public))
	assert.True(t, public.IsCA)
	assert.Error(t, verify(t, lab.CDRIntermediate.Certificate, public, nil, x509.ExtKeyUsageAny))
}

// Verifies: CA-0025, CA-0026, CA-0028, CA-0029, CA-0030, CA-0033, CA-0036
func TestAServerCertificateHasTheProfileOfAProductionCDRServerCertificate(t *testing.T) {
	lab := newLab(t)
	r := csr(t, serverCSR())

	cert, err := lab.CDRIntermediate.Issue(r, ca.Server, now)

	require.NoError(t, err)
	assert.Equal(t, r.Subject.String(), cert.Subject.String())
	assert.Equal(
		t,
		[]string{"mtls.holder.example", "api.holder.example"},
		cert.DNSNames,
		"Go checks host names only in the SAN",
	)
	assert.Equal(t, x509.KeyUsageDigitalSignature, cert.KeyUsage)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage)
	assert.True(t, cert.BasicConstraintsValid)
	assert.False(t, cert.IsCA)
	assert.Equal(t, now, cert.NotBefore)
	assert.Equal(t, now.AddDate(0, 13, 0), cert.NotAfter)
	assert.Equal(t, r.PublicKey, cert.PublicKey)
	assert.Greater(t, cert.SerialNumber.BitLen(), 64)
	require.NoError(
		t,
		verify(
			t,
			cert,
			lab.CDRRoot.Certificate,
			[]*x509.Certificate{lab.CDRIntermediate.Certificate},
			x509.ExtKeyUsageServerAuth,
		),
	)
	assert.NoError(t, cert.VerifyHostname("mtls.holder.example"))
}

// Verifies: CA-0033
func TestAServerCertificateHasEachNameOneTime(t *testing.T) {
	tmpl := serverCSR()
	tmpl.DNSNames = []string{"mtls.holder.example", "api.holder.example"}

	cert, err := newLab(t).CDRIntermediate.Issue(csr(t, tmpl), ca.Server, now)

	require.NoError(t, err)
	assert.Equal(t, []string{"mtls.holder.example", "api.holder.example"}, cert.DNSNames)
}

// Verifies: CA-0031
func TestAClientCertificateIsOnlyForClientAuthentication(t *testing.T) {
	lab := newLab(t)

	cert, err := lab.CDRIntermediate.Issue(csr(t, clientCSR()), ca.Client, now)

	require.NoError(t, err)
	assert.Equal(t, "CN=Recipient App Web,OU=Consumer Data Right,O=Holder Bank,C=AU", cert.Subject.String())
	assert.Empty(t, cert.DNSNames)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, cert.ExtKeyUsage)

	chain := []*x509.Certificate{lab.CDRIntermediate.Certificate}
	require.NoError(t, verify(t, cert, lab.CDRRoot.Certificate, chain, x509.ExtKeyUsageClientAuth))
	assert.Error(t, verify(t, cert, lab.CDRRoot.Certificate, chain, x509.ExtKeyUsageServerAuth))
}

// Verifies: CA-0009
func TestACDRCertificateDoesNotChainToThePublicRoot(t *testing.T) {
	lab := newLab(t)

	cert, err := lab.CDRIntermediate.Issue(csr(t, clientCSR()), ca.Client, now)

	require.NoError(t, err)
	assert.Error(
		t,
		verify(
			t,
			cert,
			lab.PublicRoot.Certificate,
			[]*x509.Certificate{lab.CDRIntermediate.Certificate},
			x509.ExtKeyUsageClientAuth,
		),
	)
}

// Verifies: CA-0010
func TestIssueRefusesACSRThatBreaksTheProfile(t *testing.T) {
	tmpl := clientCSR()
	tmpl.Subject.OrganizationalUnit = nil

	_, err := newLab(t).CDRIntermediate.Issue(csr(t, tmpl), ca.Client, now)

	assert.EqualError(t, err, `subject OU must be "Consumer Data Right"`)
}

// Verifies: CA-0027
func TestACertificateDoesNotExpireAfterItsIssuer(t *testing.T) {
	lab := newLab(t)
	late := lab.CDRIntermediate.Certificate.NotAfter.AddDate(0, -1, 0)

	cert, err := lab.CDRIntermediate.Issue(csr(t, clientCSR()), ca.Client, late)

	require.NoError(t, err)
	assert.Equal(t, lab.CDRIntermediate.Certificate.NotAfter, cert.NotAfter)
}

func TestEachCertificateHasADifferentSerialNumber(t *testing.T) {
	lab := newLab(t)

	first, err := lab.CDRIntermediate.Issue(csr(t, clientCSR()), ca.Client, now)
	require.NoError(t, err)
	second, err := lab.CDRIntermediate.Issue(csr(t, clientCSR()), ca.Client, now)
	require.NoError(t, err)

	assert.NotEqual(t, first.SerialNumber, second.SerialNumber)
}

// Verifies: CA-0009, CA-0038, CA-0039
func TestAPublicServerCertificateChainsOnlyToThePublicRoot(t *testing.T) {
	lab := newLab(t)

	cert, err := lab.PublicRoot.IssuePublicServer(&key(3).PublicKey, []string{"register", "localhost"}, now)

	require.NoError(t, err)
	assert.Equal(t, "CN=register", cert.Subject.String())
	assert.Equal(t, []string{"register", "localhost"}, cert.DNSNames)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage)
	assert.Equal(t, now.AddDate(0, 13, 0), cert.NotAfter)
	require.NoError(t, verify(t, cert, lab.PublicRoot.Certificate, nil, x509.ExtKeyUsageServerAuth))
	assert.Error(
		t,
		verify(
			t,
			cert,
			lab.CDRRoot.Certificate,
			[]*x509.Certificate{lab.CDRIntermediate.Certificate},
			x509.ExtKeyUsageServerAuth,
		),
	)
}

// Verifies: CA-0040
func TestAPublicServerCertificateNeedsValidDNSNames(t *testing.T) {
	lab := newLab(t)

	_, err := lab.PublicRoot.IssuePublicServer(&key(3).PublicKey, nil, now)
	require.EqualError(t, err, "at least one DNS name is required")
	_, err = lab.PublicRoot.IssuePublicServer(&key(3).PublicKey, []string{"not a name"}, now)
	require.ErrorContains(t, err, `DNS name "not a name" is not a host name`)
	_, err = lab.PublicRoot.IssuePublicServer(&key(3).PublicKey, []string{"localhost", "127.0.0.1"}, now)
	assert.EqualError(t, err, `DNS name "127.0.0.1" is an IP address, not a host name`)
}
