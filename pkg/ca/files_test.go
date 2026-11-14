package ca_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func writeFile(t *testing.T, name string, data []byte, perm os.FileMode) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, data, perm))
	require.NoError(t, os.Chmod(path, perm))

	return path
}

func keyPEM(t *testing.T, k *rsa.PrivateKey) []byte {
	t.Helper()

	der, err := x509.MarshalPKCS8PrivateKey(k)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func intermediateFiles(t *testing.T, keyPerm os.FileMode) (lab ca.Lab, certFile, keyFile string) {
	t.Helper()

	lab = newLab(t)
	certFile = writeFile(t, "intermediate.crt", ca.EncodeCertificates(lab.CDRIntermediate.Certificate), 0o644)
	keyFile = writeFile(t, "intermediate.key", keyPEM(t, key(1)), keyPerm)

	return lab, certFile, keyFile
}

func TestLoadAuthorityGivesTheIntermediateWithItsKeyAndCRLURL(t *testing.T) {
	lab, certFile, keyFile := intermediateFiles(t, 0o600)

	authority, err := ca.LoadAuthority(certFile, keyFile, "http://ca.example/crl")

	require.NoError(t, err)
	assert.Equal(t, lab.CDRIntermediate.Certificate.Raw, authority.Certificate.Raw)
	assert.True(t, key(1).PublicKey.Equal(authority.Key.Public()))
	assert.Equal(t, "http://ca.example/crl", authority.CRLURL)
}

func TestLoadAuthorityReadsAKeyFileThatAnyUserCanRead(t *testing.T) {
	_, certFile, keyFile := intermediateFiles(t, 0o644)

	_, err := ca.LoadAuthority(certFile, keyFile, "http://ca.example/crl")

	require.NoError(t, err)
}

func TestLoadAuthorityRefusesAKeyThatIsNotTheKeyOfTheCertificate(t *testing.T) {
	_, certFile, _ := intermediateFiles(t, 0o600)
	otherKey := writeFile(t, "other.key", keyPEM(t, key(2)), 0o600)

	_, err := ca.LoadAuthority(certFile, otherKey, "http://ca.example/crl")

	require.ErrorContains(t, err, "is not the key of the CA certificate")
}

func TestLoadAuthorityRefusesACertificateThatIsNotACA(t *testing.T) {
	lab := newLab(t)
	leaf, err := lab.CDRIntermediate.Issue(
		csrSignedBy(t, key(1), &x509.CertificateRequest{Subject: cdrSubject("Client")}),
		ca.Client,
		now,
	)
	require.NoError(t, err)

	certFile := writeFile(t, "leaf.crt", ca.EncodeCertificates(leaf), 0o644)
	keyFile := writeFile(t, "leaf.key", keyPEM(t, key(1)), 0o600)

	_, err = ca.LoadAuthority(certFile, keyFile, "http://ca.example/crl")

	require.ErrorContains(t, err, "is not a CA certificate")
}

func TestLoadAuthorityNeedsAnAbsoluteHTTPCRLURL(t *testing.T) {
	_, certFile, keyFile := intermediateFiles(t, 0o600)

	for _, value := range []string{"", "ftp://ca.example/crl", "/crl", "http://"} {
		_, err := ca.LoadAuthority(certFile, keyFile, value)
		require.ErrorContains(t, err, "CRL URL", value)
	}
}

func TestParseCSRGivesOnePEMCSR(t *testing.T) {
	request := csr(t, &x509.CertificateRequest{Subject: cdrSubject("Client")})
	data := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: request.Raw})

	got, err := ca.ParseCSR(data)

	require.NoError(t, err)
	assert.Equal(t, request.Raw, got.Raw)
}

func TestParseCSRRefusesATextThatIsNotOnePEMCSR(t *testing.T) {
	request := csr(t, &x509.CertificateRequest{Subject: cdrSubject("Client")})
	one := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: request.Raw})

	for name, data := range map[string][]byte{
		"empty":       nil,
		"not PEM":     []byte("not a CSR"),
		"two CSRs":    append(append([]byte{}, one...), one...),
		"other text":  append(append([]byte{}, one...), []byte("more")...),
		"certificate": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: request.Raw}),
		"not DER":     pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: []byte("x")}),
	} {
		_, err := ca.ParseCSR(data)
		assert.EqualError(t, err, "the text is not one PEM CSR", name)
	}
}

func TestEncodeCertificatesGivesPEMInTheSameSequence(t *testing.T) {
	lab := newLab(t)

	data := ca.EncodeCertificates(lab.CDRIntermediate.Certificate, lab.CDRRoot.Certificate)
	path := writeFile(t, "bundle.pem", data, 0o644)
	got, err := ca.ReadCertificates(path)

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, lab.CDRIntermediate.Certificate.Raw, got[0].Raw)
	assert.Equal(t, lab.CDRRoot.Certificate.Raw, got[1].Raw)
}

func TestParseCertificatesGivesOneCertificateOrAChainInTheSameSequence(t *testing.T) {
	lab := newLab(t)
	one := ca.EncodeCertificates(lab.CDRIntermediate.Certificate)

	got, err := ca.ParseCertificates(one)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, lab.CDRIntermediate.Certificate.Raw, got[0].Raw)

	chain := append([]byte("\n"), ca.EncodeCertificates(lab.CDRIntermediate.Certificate, lab.CDRRoot.Certificate)...)
	got, err = ca.ParseCertificates(append(chain, '\n'))
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, lab.CDRRoot.Certificate.Raw, got[1].Raw)
}

func TestParseCertificatesRefusesATextThatIsNotPEMCertificates(t *testing.T) {
	lab := newLab(t)
	one := ca.EncodeCertificates(lab.CDRIntermediate.Certificate)
	request := csr(t, &x509.CertificateRequest{Subject: cdrSubject("Client")})
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: request.Raw})

	for name, data := range map[string][]byte{
		"empty":            nil,
		"not PEM":          []byte("not a certificate"),
		"text before":      append([]byte("text\n"), one...),
		"text after":       append(append([]byte{}, one...), []byte("more")...),
		"a CSR":            csrPEM,
		"a CSR after":      append(append([]byte{}, one...), csrPEM...),
		"not DER":          pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("x")}),
		"a private key":    keyPEM(t, key(0)),
		"text between two": append(append(append([]byte{}, one...), []byte("text\n")...), one...),
	} {
		_, err := ca.ParseCertificates(data)
		assert.EqualError(t, err, "the text is not PEM certificates", name)
	}
}
