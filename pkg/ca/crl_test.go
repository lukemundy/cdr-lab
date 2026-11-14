package ca_test

import (
	"crypto/x509"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func TestTheCRLIsSignedByTheIntermediateAndValidForSevenDays(t *testing.T) {
	lab := newLab(t)
	revoked := []ca.Revocation{
		{Serial: big.NewInt(42), Time: now.Add(-time.Hour), Reason: ca.KeyCompromise},
		{Serial: big.NewInt(7), Time: now.Add(-time.Minute), Reason: ca.CessationOfOperation},
	}

	der, err := lab.CDRIntermediate.CRL(revoked, big.NewInt(3), now)

	require.NoError(t, err)
	crl, err := x509.ParseRevocationList(der)
	require.NoError(t, err)
	require.NoError(t, crl.CheckSignatureFrom(lab.CDRIntermediate.Certificate))
	assert.Equal(t, x509.SHA256WithRSA, crl.SignatureAlgorithm)
	assert.Equal(t, big.NewInt(3), crl.Number)
	assert.Equal(t, now, crl.ThisUpdate)
	assert.Equal(t, now.AddDate(0, 0, 7), crl.NextUpdate)
	assert.Equal(t, lab.CDRIntermediate.Certificate.SubjectKeyId, crl.AuthorityKeyId)
	require.Len(t, crl.RevokedCertificateEntries, 2)
	assert.Equal(t, big.NewInt(42), crl.RevokedCertificateEntries[0].SerialNumber)
	assert.Equal(t, now.Add(-time.Hour), crl.RevokedCertificateEntries[0].RevocationTime)
	assert.Equal(t, int(ca.KeyCompromise), crl.RevokedCertificateEntries[0].ReasonCode)
	assert.Equal(t, int(ca.CessationOfOperation), crl.RevokedCertificateEntries[1].ReasonCode)
}

func TestAnEmptyCRLIsValid(t *testing.T) {
	lab := newLab(t)

	der, err := lab.CDRIntermediate.CRL(nil, big.NewInt(1), now)

	require.NoError(t, err)
	crl, err := x509.ParseRevocationList(der)
	require.NoError(t, err)
	assert.Empty(t, crl.RevokedCertificateEntries)
}

func TestParseReasonAcceptsOnlyTheReasonsForALeafCertificate(t *testing.T) {
	for name, want := range map[string]ca.Reason{
		"unspecified":          ca.Unspecified,
		"keyCompromise":        ca.KeyCompromise,
		"affiliationChanged":   ca.AffiliationChanged,
		"superseded":           ca.Superseded,
		"cessationOfOperation": ca.CessationOfOperation,
		"privilegeWithdrawn":   ca.PrivilegeWithdrawn,
	} {
		got, err := ca.ParseReason(name)

		require.NoError(t, err, name)
		assert.Equal(t, want, got, name)
	}

	for _, name := range []string{"cACompromise", "certificateHold", "removeFromCRL", "aACompromise", "lost", ""} {
		_, err := ca.ParseReason(name)

		assert.Error(t, err, name)
	}
}

func TestACDRCertificateHasTheCRLURLOfItsIssuer(t *testing.T) {
	lab := newLab(t)
	lab.CDRIntermediate.CRLURL = "http://register:8085/crl/cdr-intermediate.crl"

	cert, err := lab.CDRIntermediate.Issue(csr(t, clientCSR()), ca.Client, now)

	require.NoError(t, err)
	assert.Equal(t, []string{"http://register:8085/crl/cdr-intermediate.crl"}, cert.CRLDistributionPoints)
}

func TestACertificateHasNoCRLURLWhenItsIssuerHasNone(t *testing.T) {
	lab := newLab(t)

	cert, err := lab.PublicRoot.IssuePublicServer(&key(3).PublicKey, []string{"register"}, now)

	require.NoError(t, err)
	assert.Empty(t, cert.CRLDistributionPoints)
}

func TestAReasonGivesItsRFC5280Name(t *testing.T) {
	for _, name := range []string{
		"unspecified", "keyCompromise", "affiliationChanged", "superseded",
		"cessationOfOperation", "privilegeWithdrawn",
	} {
		r, err := ca.ParseReason(name)

		require.NoError(t, err)
		assert.Equal(t, name, r.String())
	}
}

func TestAnUnspecifiedRevocationHasNoReasonCode(t *testing.T) {
	lab := newLab(t)

	der, err := lab.CDRIntermediate.CRL([]ca.Revocation{{Serial: big.NewInt(9), Time: now, Reason: ca.Unspecified}},
		big.NewInt(1), now)

	require.NoError(t, err)
	crl, err := x509.ParseRevocationList(der)
	require.NoError(t, err)
	require.Len(t, crl.RevokedCertificateEntries, 1)

	for _, ext := range crl.RevokedCertificateEntries[0].Extensions {
		assert.False(t, ext.Id.Equal([]int{2, 5, 29, 21}), "the entry has a reason code extension")
	}
}
