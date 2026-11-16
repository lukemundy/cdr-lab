package ca_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func serverCSR() *x509.CertificateRequest {
	return &x509.CertificateRequest{
		Subject:  cdrSubject("mtls.holder.example"),
		DNSNames: []string{"api.holder.example"},
	}
}

func clientCSR() *x509.CertificateRequest {
	return &x509.CertificateRequest{Subject: cdrSubject("Recipient App Web")}
}

func TestACSRThatFollowsTheCDSProfileIsValid(t *testing.T) {
	assert.NoError(t, ca.ValidateCSR(csr(t, serverCSR()), ca.Server))
	assert.NoError(t, ca.ValidateCSR(csr(t, clientCSR()), ca.Client))
}

// Verifies: CA-0023
func TestASingleLabelServerNameIsValid(t *testing.T) {
	tmpl := serverCSR()
	tmpl.Subject.CommonName = "register"
	tmpl.DNSNames = []string{"localhost"}

	assert.NoError(t, ca.ValidateCSR(csr(t, tmpl), ca.Server))
}

const notAHostName = "is not a host name: use only letters, digits and hyphens in labels of 1 to 63 characters, " +
	"with a dot between two labels and no hyphen at the start or end of a label"

// longHostName gives a valid host name of n characters, with labels of 63 characters or fewer.
func longHostName(n int) string {
	name := strings.Repeat(strings.Repeat("a", 62)+".", n/63) + strings.Repeat("a", n%63)

	return strings.TrimSuffix(name, ".")
}

// Verifies: CA-0021, CA-0022
func TestAServerNameAtEachLengthLimitIsValid(t *testing.T) {
	tmpl := serverCSR()
	tmpl.Subject.CommonName = longHostName(64)
	tmpl.DNSNames = []string{longHostName(253)}

	assert.NoError(t, ca.ValidateCSR(csr(t, tmpl), ca.Server))
}

// Verifies: CA-0013, CA-0015, CA-0017, CA-0018, CA-0019, CA-0021, CA-0022, CA-0024
func TestACSRThatBreaksTheCDSProfileIsRefused(t *testing.T) {
	tests := map[string]struct {
		kind    ca.Kind
		change  func(*x509.CertificateRequest)
		wantErr string
	}{
		"no common name": {
			ca.Client,
			func(r *x509.CertificateRequest) { r.Subject.CommonName = "" },
			"subject CN is required",
		},
		"no organisation": {
			ca.Client,
			func(r *x509.CertificateRequest) { r.Subject.Organization = nil },
			"subject O is required",
		},
		"no organisational unit": {
			ca.Client, func(r *x509.CertificateRequest) { r.Subject.OrganizationalUnit = nil },
			`subject OU must be "Consumer Data Right"`,
		},
		"other organisational unit": {
			ca.Client, func(r *x509.CertificateRequest) { r.Subject.OrganizationalUnit = []string{"IT"} },
			`subject OU must be "Consumer Data Right"`,
		},
		"no country": {
			ca.Client,
			func(r *x509.CertificateRequest) { r.Subject.Country = nil },
			"subject C is required",
		},
		"country is not a two-letter code": {
			ca.Client, func(r *x509.CertificateRequest) { r.Subject.Country = []string{"Australia"} },
			`subject C "Australia" is not a two-letter country code`,
		},
		"client with a DNS name": {
			ca.Client, func(r *x509.CertificateRequest) { r.DNSNames = []string{"app.recipient.example"} },
			"a client CSR must not have subject alternative names",
		},
		"client with an email address": {
			ca.Client, func(r *x509.CertificateRequest) { r.EmailAddresses = []string{"it@recipient.example"} },
			"a client CSR must not have subject alternative names",
		},
		"server common name is not a DNS name": {
			ca.Server,
			func(r *x509.CertificateRequest) { r.Subject.CommonName = "Holder Bank API" },
			`subject CN "Holder Bank API" ` + notAHostName,
		},
		"server wildcard name": {
			ca.Server,
			func(r *x509.CertificateRequest) { r.DNSNames = []string{"*.holder.example"} },
			`DNS name "*.holder.example" ` + notAHostName,
		},
		"server IPv4 address as the common name": {
			ca.Server, func(r *x509.CertificateRequest) { r.Subject.CommonName = "10.0.0.1" },
			`subject CN "10.0.0.1" is an IP address, not a host name`,
		},
		"server IPv4 address as a DNS name": {
			ca.Server, func(r *x509.CertificateRequest) { r.DNSNames = []string{"192.168.1.20"} },
			`DNS name "192.168.1.20" is an IP address, not a host name`,
		},
		"server common name of more than 64 characters": {
			ca.Server, func(r *x509.CertificateRequest) { r.Subject.CommonName = longHostName(65) },
			`subject CN "` + longHostName(65) + `" has more than 64 characters`,
		},
		"server DNS name of more than 253 characters": {
			ca.Server, func(r *x509.CertificateRequest) { r.DNSNames = []string{longHostName(254)} },
			`DNS name "` + longHostName(254) + `" has more than 253 characters`,
		},
		"server IP address": {
			ca.Server, func(r *x509.CertificateRequest) { r.IPAddresses = []net.IP{net.ParseIP("10.0.0.1")} },
			"a server CSR can have only DNS names as subject alternative names",
		},
		"server URI": {
			ca.Server,
			func(r *x509.CertificateRequest) { r.URIs = []*url.URL{{Scheme: "https", Host: "holder.example"}} },
			"a server CSR can have only DNS names as subject alternative names",
		},
		"SHA-384 signature": {
			ca.Client, func(r *x509.CertificateRequest) { r.SignatureAlgorithm = x509.SHA384WithRSA },
			"the signature algorithm must be SHA256-RSA, not SHA384-RSA",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tmpl := clientCSR()
			if tt.kind == ca.Server {
				tmpl = serverCSR()
			}

			tt.change(tmpl)

			assert.EqualError(t, ca.ValidateCSR(csr(t, tmpl), tt.kind), tt.wantErr)
		})
	}
}

// Verifies: CA-0012
func TestACSRWithAKeyThatIsNotRSA2048IsRefused(t *testing.T) {
	rsa3072, err := rsa.GenerateKey(rand.Reader, 3072)
	require.NoError(t, err)
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	require.ErrorContains(
		t,
		ca.ValidateCSR(csrSignedBy(t, rsa3072, clientCSR()), ca.Client),
		"the key must be RSA 2048, not RSA 3072",
	)
	assert.ErrorContains(
		t,
		ca.ValidateCSR(csrSignedBy(t, ec, clientCSR()), ca.Client),
		"the key must be RSA 2048, not ECDSA",
	)
}

// Verifies: CA-0014
func TestACSRWithASignatureThatIsNotValidIsRefused(t *testing.T) {
	r := csr(t, clientCSR())
	r.Signature[0] ^= 0xff

	assert.ErrorContains(t, ca.ValidateCSR(r, ca.Client), "the CSR signature is not valid")
}

// Verifies: CA-0011
func TestACSRNamesEachRuleThatItBreaks(t *testing.T) {
	tmpl := clientCSR()
	tmpl.Subject.CommonName = ""
	tmpl.Subject.Country = nil

	assert.EqualError(t, ca.ValidateCSR(csr(t, tmpl), ca.Client), "subject CN is required\nsubject C is required")
}
