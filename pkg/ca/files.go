package ca

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
)

func LoadAuthority(certFile, keyFile, crlURL string) (Authority, error) {
	if err := CheckCRLURL(crlURL); err != nil {
		return Authority{}, err
	}

	certs, err := ReadCertificates(certFile)
	if err != nil {
		return Authority{}, err
	}

	cert := certs[0]
	if !cert.IsCA || cert.KeyUsage&x509.KeyUsageCertSign == 0 {
		return Authority{}, fmt.Errorf("%s is not a CA certificate", certFile)
	}

	key, err := ReadKey(keyFile)
	if err != nil {
		return Authority{}, err
	}

	if !publicKeysEqual(cert.PublicKey, key.Public()) {
		return Authority{}, fmt.Errorf("%s is not the key of the CA certificate %s", keyFile, certFile)
	}

	return Authority{Certificate: cert, Key: key, CRLURL: crlURL}, nil
}

func publicKeysEqual(a, b crypto.PublicKey) bool {
	k, ok := a.(interface{ Equal(x crypto.PublicKey) bool })
	return ok && k.Equal(b)
}

func CheckCRLURL(value string) error {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("the CRL URL %q is not an absolute http or https URL", value)
	}

	return nil
}

func ReadCertificates(path string) ([]*x509.Certificate, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	blocks := pemBlocks(data, "CERTIFICATE")
	if len(blocks) == 0 {
		return nil, fmt.Errorf("%s: no PEM CERTIFICATE", path)
	}

	certs := make([]*x509.Certificate, 0, len(blocks))
	for _, b := range blocks {
		c, err := x509.ParseCertificate(b.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}

		certs = append(certs, c)
	}

	return certs, nil
}

func ReadKey(path string) (crypto.Signer, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	key, err := ParseKey(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return key, nil
}

// ParseKey reads the first PEM PKCS #8 private key of data. Its error does not contain the text.
func ParseKey(data []byte) (crypto.Signer, error) {
	blocks := pemBlocks(data, "PRIVATE KEY")
	if len(blocks) == 0 {
		return nil, errors.New("no PEM PRIVATE KEY")
	}

	key, err := x509.ParsePKCS8PrivateKey(blocks[0].Bytes)
	if err != nil {
		return nil, errors.New("the PEM PRIVATE KEY is not a PKCS #8 key")
	}

	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New("the key cannot sign")
	}

	return signer, nil
}

// EncodeKey gives the key as a PEM PKCS #8 private key.
func EncodeKey(key crypto.Signer) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

// ParseCSR refuses a second PEM block or other text, and its error does not contain the text.
func ParseCSR(data []byte) (*x509.CertificateRequest, error) {
	notCSR := errors.New("the text is not one PEM CSR")

	block, rest := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE REQUEST" || len(bytes.TrimSpace(rest)) != 0 {
		return nil, notCSR
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, notCSR
	}

	return csr, nil
}

// ParseCertificates refuses a PEM block that is not a certificate, and other text. Its error does not contain the text.
func ParseCertificates(data []byte) ([]*x509.Certificate, error) {
	notCertificates := errors.New("the text is not PEM certificates")

	var certs []*x509.Certificate

	for rest := bytes.TrimSpace(data); len(rest) > 0; {
		block, after := pem.Decode(rest)
		if block == nil || block.Type != "CERTIFICATE" || !bytes.HasPrefix(rest, []byte("-----BEGIN ")) {
			return nil, notCertificates
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, notCertificates
		}

		certs = append(certs, cert)
		rest = bytes.TrimSpace(after)
	}

	if len(certs) == 0 {
		return nil, notCertificates
	}

	return certs, nil
}

func EncodeCertificates(certs ...*x509.Certificate) []byte {
	var data []byte
	for _, c := range certs {
		data = append(data, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})...)
	}

	return data
}

func readFile(path string) ([]byte, error) {
	data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return data, nil
}

func pemBlocks(data []byte, blockType string) []*pem.Block {
	var blocks []*pem.Block

	for {
		var b *pem.Block

		b, data = pem.Decode(data)
		if b == nil {
			return blocks
		}

		if b.Type == blockType {
			blocks = append(blocks, b)
		}
	}
}
