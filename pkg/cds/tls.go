package cds

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// TLSConfig permits the TLS 1.2 suites that the CDS permits, and TLS 1.3, whose suites Go does not let a server choose.
func TLSConfig(cert *ServerCertificate) *tls.Config {
	return &tls.Config{
		GetCertificate: cert.GetCertificate,
		MinVersion:     tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}
}

const revocationCheckTimeout = 5 * time.Second

// MutualTLSConfig stops the revocation checks that are in progress at the end
// of ctx, because a TLS handshake has no context of its own.
func MutualTLSConfig(
	ctx context.Context,
	cert *ServerCertificate,
	clientCAs *x509.CertPool,
	checkRevocation func(context.Context, *x509.Certificate) error,
) *tls.Config {
	cfg := TLSConfig(cert)
	cfg.ClientAuth = tls.RequireAndVerifyClientCert
	cfg.ClientCAs = clientCAs
	// VerifyConnection also runs when a session resumes, so a certificate that is revoked later cannot resume.
	cfg.VerifyConnection = func(state tls.ConnectionState) error {
		ctx, cancel := context.WithTimeout(ctx, revocationCheckTimeout)
		defer cancel()

		return checkRevocation(ctx, state.PeerCertificates[0])
	}

	return cfg
}

// LoadCABundle reads the intermediate and the root: clients can send only their leaf certificate.
func LoadCABundle(file string) ([]*x509.Certificate, error) {
	data, err := os.ReadFile(file) //nolint:gosec // G304: the operator of the service gives the file name
	if err != nil {
		return nil, fmt.Errorf("load the CDR CA bundle: %w", err)
	}

	var certs []*x509.Certificate

	for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("load the CDR CA bundle: %w", err)
		}

		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("load the CDR CA bundle: %s has no PEM certificate", file)
	}

	return certs, nil
}
