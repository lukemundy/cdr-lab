package root

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/cds/revocation"
	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/brands"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/branddir"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// config has the settings of only the parts that the mode runs.
type config struct {
	api     apiConfig
	workers workerConfig
}

type apiConfig struct {
	brand branddir.Brand
	// publicBaseURL is also the issuer of the authorisation server.
	publicBaseURL string
	secureBaseURL string
	// registerJWKSURL is the URL of the keys that verify an SSA.
	registerJWKSURL string
	operator        operatorhttp.Config
	// smtpAddress is the host and port of the SMTP server that takes the emails to customers.
	smtpAddress string
	// mailFrom is the sender address of those emails.
	mailFrom   string
	fixedOTPs  domain.FixedOTPs
	publicCert tls.Certificate
	secureCert tls.Certificate
	cdrCAs     []*x509.Certificate
	crlMaxAge  time.Duration
	console    consolehttp.Config
	bankSite   consolehttp.Config
	// brandID is the dataHolderBrandId of the brand on the Register. Only the online bank site uses it.
	brandID string
}

type workerConfig struct {
	// brandID is the dataHolderBrandId of the brand on the Register.
	brandID               string
	registerPublicBaseURL string
	statusInterval        time.Duration
}

func loadConfig(m mode) (config, error) {
	var cfg config

	if m.api {
		api, err := loadAPIConfig()
		if err != nil {
			return config{}, err
		}

		cfg.api = api
	}

	if m.workers {
		workers, err := loadWorkerConfig()
		if err != nil {
			return config{}, err
		}

		cfg.workers = workers
	}

	return cfg, nil
}

func loadWorkerConfig() (workerConfig, error) {
	cfg := workerConfig{
		brandID:               os.Getenv("DATA_HOLDER_BRAND_ID"),
		registerPublicBaseURL: os.Getenv("REGISTER_PUBLIC_BASE_URL"),
	}
	if cfg.brandID == "" {
		return workerConfig{}, errors.New(
			"set DATA_HOLDER_BRAND_ID to the dataHolderBrandId of the brand on the Register",
		)
	}

	if cfg.registerPublicBaseURL == "" {
		return workerConfig{}, errors.New(
			"set REGISTER_PUBLIC_BASE_URL to the base URI of the public APIs of the Register, which give the statuses",
		)
	}

	interval, err := time.ParseDuration(os.Getenv("REGISTER_STATUS_INTERVAL"))
	if err != nil || interval <= 0 || interval > maxStatusInterval {
		return workerConfig{}, fmt.Errorf(
			"set REGISTER_STATUS_INTERVAL to the interval at which the data holder gets the statuses, "+
				"more than 0 and at most %s, for example 1m", maxStatusInterval,
		)
	}

	cfg.statusInterval = interval

	return cfg, nil
}

func loadAPIConfig() (apiConfig, error) {
	cfg := apiConfig{
		publicBaseURL:   os.Getenv("PUBLIC_BASE_URL"),
		secureBaseURL:   os.Getenv("SECURE_BASE_URL"),
		registerJWKSURL: os.Getenv("REGISTER_JWKS_URL"),
		smtpAddress:     os.Getenv("SMTP_ADDRESS"),
		mailFrom:        os.Getenv("MAIL_FROM"),
	}

	var err error
	if cfg.brand, err = loadBrand(); err != nil {
		return apiConfig{}, err
	}

	if cfg.publicBaseURL == "" {
		return apiConfig{}, errors.New("set PUBLIC_BASE_URL to the base URL that clients use for the public APIs")
	}

	if cfg.secureBaseURL == "" {
		return apiConfig{}, errors.New("set SECURE_BASE_URL to the base URL that clients use for the secure APIs")
	}

	if cfg.registerJWKSURL == "" {
		return apiConfig{}, errors.New("set REGISTER_JWKS_URL to the URL of the Register JWKS, which verifies each SSA")
	}

	if cfg.smtpAddress == "" {
		return apiConfig{}, errors.New(
			"set SMTP_ADDRESS to the host and port of the SMTP server that takes the emails to customers",
		)
	}

	if cfg.mailFrom == "" {
		return apiConfig{}, errors.New("set MAIL_FROM to the sender address of the emails to customers")
	}

	// The brand name is the display name of the sender, and the brand directory is its only source.
	if from, err := mail.ParseAddress(cfg.mailFrom); err != nil || from.Name != "" || from.Address != cfg.mailFrom {
		return apiConfig{}, errors.New(
			"MAIL_FROM: give only the sender address, with no display name, for example no-reply@bank.test",
		)
	}

	if cfg.operator, err = operatorhttp.LoadConfig("OPERATOR_ALLOWED_HOSTS"); err != nil {
		return apiConfig{}, err
	}

	if cfg.fixedOTPs, err = domain.ParseFixedOTPs(os.Getenv("FIXED_OTPS")); err != nil {
		return apiConfig{}, fmt.Errorf("FIXED_OTPS: %w", err)
	}

	publicCertFile, publicKeyFile := os.Getenv("PUBLIC_TLS_CERT_FILE"), os.Getenv("PUBLIC_TLS_KEY_FILE")
	if publicCertFile == "" || publicKeyFile == "" {
		return apiConfig{}, errors.New(
			"set PUBLIC_TLS_CERT_FILE and PUBLIC_TLS_KEY_FILE to the TLS certificate and key of the public APIs",
		)
	}

	if cfg.publicCert, err = tls.LoadX509KeyPair(publicCertFile, publicKeyFile); err != nil {
		return apiConfig{}, fmt.Errorf("load the TLS certificate of the public APIs: %w", err)
	}

	secureCertFile, secureKeyFile := os.Getenv("SECURE_TLS_CERT_FILE"), os.Getenv("SECURE_TLS_KEY_FILE")
	if secureCertFile == "" || secureKeyFile == "" {
		return apiConfig{}, errors.New(
			"set SECURE_TLS_CERT_FILE and SECURE_TLS_KEY_FILE to the TLS certificate and key of the secure APIs",
		)
	}

	if cfg.secureCert, err = tls.LoadX509KeyPair(secureCertFile, secureKeyFile); err != nil {
		return apiConfig{}, fmt.Errorf("load the TLS certificate of the secure APIs: %w", err)
	}

	caBundleFile := os.Getenv("CDR_CA_BUNDLE_FILE")
	if caBundleFile == "" {
		return apiConfig{}, errors.New(
			"set CDR_CA_BUNDLE_FILE to the CDR CA bundle that client certificates of the secure APIs must chain to",
		)
	}

	if cfg.cdrCAs, err = cds.LoadCABundle(caBundleFile); err != nil {
		return apiConfig{}, err
	}

	if cfg.crlMaxAge, err = loadCRLMaxAge(); err != nil {
		return apiConfig{}, err
	}

	if cfg.console, err = console.Settings.LoadConfig(); err != nil {
		return apiConfig{}, err
	}

	if cfg.bankSite, err = banksite.Settings.LoadConfig(); err != nil {
		return apiConfig{}, err
	}

	if cfg.bankSite.Enabled {
		if cfg.brandID = os.Getenv("DATA_HOLDER_BRAND_ID"); !cookieToken.MatchString(cfg.brandID) {
			return apiConfig{}, errors.New(
				"set DATA_HOLDER_BRAND_ID to the dataHolderBrandId of the brand on the Register, " +
					"which names the session cookie of the online bank site, with only the characters of a cookie name",
			)
		}
	}

	return cfg, nil
}

// cookieToken is the token of RFC 6265, section 4.1.1, which is the syntax of a cookie name.
var cookieToken = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")

// loadBrand reads the brand from a preset of the data holder or from a brand directory, and validates it.
func loadBrand() (branddir.Brand, error) {
	preset, dir := os.Getenv("DATA_HOLDER_BRAND"), os.Getenv("DATA_HOLDER_BRAND_DIR")

	var (
		source  string
		brandFS fs.FS
	)

	switch {
	case preset != "" && dir != "":
		return branddir.Brand{}, errors.New("set only one of DATA_HOLDER_BRAND and DATA_HOLDER_BRAND_DIR")
	case preset != "":
		source = "DATA_HOLDER_BRAND " + preset

		var err error
		if brandFS, err = brands.Preset(preset); err != nil {
			return branddir.Brand{}, fmt.Errorf("DATA_HOLDER_BRAND: %w", err)
		}
	case dir != "":
		source = "DATA_HOLDER_BRAND_DIR " + dir

		// A root refuses a symbolic link out of the directory, and permits the relative links of a ConfigMap volume.
		root, err := os.OpenRoot(dir)
		if err != nil {
			return branddir.Brand{}, fmt.Errorf("%s: open the brand directory: %w", source, err)
		}
		defer func() { _ = root.Close() }()

		brandFS = root.FS()
	default:
		return branddir.Brand{}, fmt.Errorf(
			"set DATA_HOLDER_BRAND to the name of a preset brand (%s), "+
				"or DATA_HOLDER_BRAND_DIR to the path of a brand directory",
			strings.Join(brands.Names(), ", "),
		)
	}

	brand, err := branddir.Load(brandFS, checkProducts)
	if err != nil {
		return branddir.Brand{}, fmt.Errorf("%s: the brand is not valid:\n%w", source, err)
	}

	return brand, nil
}

func checkProducts(products [][]byte) error {
	_, err := app.CheckProducts(products)
	return err
}

func loadCRLMaxAge() (time.Duration, error) {
	value := os.Getenv("CRL_MAX_AGE")
	if value == "" {
		return revocation.DefaultMaxAge, nil
	}

	maxAge, err := time.ParseDuration(value)
	if err != nil || maxAge < 0 {
		return 0, errors.New(
			"set CRL_MAX_AGE to the time that the data holder keeps a CRL, 0 or more, for example 1m, or leave it out",
		)
	}

	return maxAge, nil
}
