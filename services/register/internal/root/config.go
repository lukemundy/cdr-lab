package root

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
)

type config struct {
	publicBaseURL string
	secureBaseURL string
	operator      operatorhttp.Config
	publicCert    tls.Certificate
	secureCert    tls.Certificate
	cdrCAs        []*x509.Certificate
	intermediate  ca.Authority
	console       consolehttp.Config
	consoleUI     fs.FS
}

// loadConfig loads only the CA settings for a process with no APIs, because the workers only sign CRLs.
func loadConfig(m mode) (config, error) {
	if !m.api {
		var cfg config

		err := loadCA(&cfg)

		return cfg, err
	}

	cfg := config{
		publicBaseURL: os.Getenv("PUBLIC_BASE_URL"),
		secureBaseURL: os.Getenv("SECURE_BASE_URL"),
	}
	if cfg.publicBaseURL == "" {
		return config{}, errors.New("set PUBLIC_BASE_URL to the base URL that clients use for the public APIs")
	}

	if cfg.secureBaseURL == "" {
		return config{}, errors.New("set SECURE_BASE_URL to the base URL that clients use for the secure APIs")
	}

	var err error
	if cfg.operator, err = operatorhttp.LoadConfig(operatorhttp.AllowedHostsEnv); err != nil {
		return config{}, err
	}

	publicCertFile, publicKeyFile := os.Getenv("PUBLIC_TLS_CERT_FILE"), os.Getenv("PUBLIC_TLS_KEY_FILE")
	if publicCertFile == "" || publicKeyFile == "" {
		return config{}, errors.New(
			"set PUBLIC_TLS_CERT_FILE and PUBLIC_TLS_KEY_FILE to the TLS certificate and key of the public APIs",
		)
	}

	if cfg.publicCert, err = tls.LoadX509KeyPair(publicCertFile, publicKeyFile); err != nil {
		return config{}, fmt.Errorf("load the TLS certificate of the public APIs: %w", err)
	}

	secureCertFile, secureKeyFile := os.Getenv("SECURE_TLS_CERT_FILE"), os.Getenv("SECURE_TLS_KEY_FILE")
	if secureCertFile == "" || secureKeyFile == "" {
		return config{}, errors.New(
			"set SECURE_TLS_CERT_FILE and SECURE_TLS_KEY_FILE to the TLS certificate and key of the secure APIs",
		)
	}

	if cfg.secureCert, err = tls.LoadX509KeyPair(secureCertFile, secureKeyFile); err != nil {
		return config{}, fmt.Errorf("load the TLS certificate of the secure APIs: %w", err)
	}

	if err := loadCA(&cfg); err != nil {
		return config{}, err
	}

	if cfg.console, err = consolehttp.Console.LoadConfig(); err != nil {
		return config{}, err
	}

	cfg.consoleUI = console.UI()

	return cfg, nil
}

func loadCA(cfg *config) error {
	caBundleFile := os.Getenv("CDR_CA_BUNDLE_FILE")
	if caBundleFile == "" {
		return errors.New(
			"set CDR_CA_BUNDLE_FILE to the CDR CA bundle that client certificates of the secure APIs must chain to",
		)
	}

	var err error
	if cfg.cdrCAs, err = cds.LoadCABundle(caBundleFile); err != nil {
		return err
	}

	cfg.intermediate, err = loadIntermediate(cfg.cdrCAs)

	return err
}

func loadIntermediate(cdrCAs []*x509.Certificate) (ca.Authority, error) {
	certFile, keyFile := os.Getenv("CDR_INTERMEDIATE_CERT_FILE"), os.Getenv("CDR_INTERMEDIATE_KEY_FILE")
	crlURL := os.Getenv("CDR_INTERMEDIATE_CRL_URL")

	switch {
	case certFile == "" || keyFile == "":
		return ca.Authority{}, errors.New(
			"set CDR_INTERMEDIATE_CERT_FILE and CDR_INTERMEDIATE_KEY_FILE to the certificate and key of the mock " +
				"CDR intermediate CA",
		)
	case crlURL == "":
		return ca.Authority{}, errors.New(
			"set CDR_INTERMEDIATE_CRL_URL to the URL of the CRL of the mock CDR intermediate CA, which each " +
				"certificate that the Register issues gives",
		)
	}

	authority, err := ca.LoadAuthority(certFile, keyFile, crlURL)
	if err != nil {
		return ca.Authority{}, fmt.Errorf("load the mock CDR intermediate CA: %w", err)
	}

	if !slices.ContainsFunc(cdrCAs, authority.Certificate.Equal) {
		return ca.Authority{}, fmt.Errorf("%s is not in the CDR CA bundle CDR_CA_BUNDLE_FILE", certFile)
	}

	return authority, nil
}
