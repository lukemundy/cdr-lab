package ca_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// RSA key generation is slow, so the tests share a few keys.
var keys = sync.OnceValue(func() []*rsa.PrivateKey {
	var ks []*rsa.PrivateKey

	for range 4 {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}

		ks = append(ks, k)
	}

	return ks
})

func key(i int) *rsa.PrivateKey { return keys()[i] }

func cdrSubject(cn string) pkix.Name {
	return pkix.Name{
		CommonName:         cn,
		Organization:       []string{"Holder Bank"},
		OrganizationalUnit: []string{"Consumer Data Right"},
		Country:            []string{"AU"},
	}
}

func csr(t *testing.T, tmpl *x509.CertificateRequest) *x509.CertificateRequest {
	t.Helper()
	return csrSignedBy(t, key(3), tmpl)
}

func csrSignedBy(t *testing.T, signer crypto.Signer, tmpl *x509.CertificateRequest) *x509.CertificateRequest {
	t.Helper()

	der, err := x509.CreateCertificateRequest(rand.Reader, tmpl, signer)
	require.NoError(t, err)
	parsed, err := x509.ParseCertificateRequest(der)
	require.NoError(t, err)

	return parsed
}
