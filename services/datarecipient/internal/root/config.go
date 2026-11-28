package root

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/datarecipient/brands"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/adapter/branddir"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/adapter/webapp"
)

// config has only the settings that serve uses. Compose and the charts also give settings that only a later function
// uses, for example SMTP_ADDRESS.
type config struct {
	brand branddir.Brand
	// webAppHosts are the Host header values of the web app, in lower case.
	webAppHosts []string
	webAppCert  certificatePaths
}

// The listener, not the configuration, reads the files, because the reload lasts as long as the process.
type certificatePaths struct{ cert, key string }

func (f certificatePaths) watch(ctx context.Context, name string) (*cds.ServerCertificate, error) {
	return cds.WatchCertificateFiles(ctx, name, f.cert, f.key)
}

func loadConfig() (config, error) {
	var (
		cfg config
		err error
	)

	if cfg.brand, err = loadBrand(); err != nil {
		return config{}, err
	}

	if cfg.webAppHosts, err = webapp.Settings.LoadAllowedHosts(); err != nil {
		return config{}, err
	}

	cfg.webAppCert = certificatePaths{cert: os.Getenv("WEB_APP_TLS_CERT_FILE"), key: os.Getenv("WEB_APP_TLS_KEY_FILE")}
	if cfg.webAppCert.cert == "" || cfg.webAppCert.key == "" {
		return config{}, errors.New(
			"set WEB_APP_TLS_CERT_FILE and WEB_APP_TLS_KEY_FILE to the TLS certificate and key of the web app, " +
				"from the public CA of the deployment",
		)
	}

	return cfg, nil
}

// loadBrand reads the brand from a preset of the data recipient or from a brand directory, and validates it.
func loadBrand() (branddir.Brand, error) {
	preset, dir := os.Getenv("DATA_RECIPIENT_BRAND"), os.Getenv("DATA_RECIPIENT_BRAND_DIR")

	var (
		source  string
		brandFS fs.FS
	)

	switch {
	case preset != "" && dir != "":
		return branddir.Brand{}, errors.New("set only one of DATA_RECIPIENT_BRAND and DATA_RECIPIENT_BRAND_DIR")
	case preset != "":
		source = "DATA_RECIPIENT_BRAND " + preset

		var err error
		if brandFS, err = brands.Preset(preset); err != nil {
			return branddir.Brand{}, fmt.Errorf("DATA_RECIPIENT_BRAND: %w", err)
		}
	case dir != "":
		source = "DATA_RECIPIENT_BRAND_DIR " + dir

		// A root refuses a symbolic link out of the directory, and permits the relative links of a ConfigMap volume.
		root, err := os.OpenRoot(dir)
		if err != nil {
			return branddir.Brand{}, fmt.Errorf("%s: open the brand directory: %w", source, err)
		}
		defer func() { _ = root.Close() }()

		brandFS = root.FS()
	default:
		return branddir.Brand{}, fmt.Errorf(
			"set DATA_RECIPIENT_BRAND to the name of a preset brand (%s), "+
				"or DATA_RECIPIENT_BRAND_DIR to the path of a brand directory",
			strings.Join(brands.Names(), ", "),
		)
	}

	brand, err := branddir.Load(brandFS)
	if err != nil {
		return branddir.Brand{}, fmt.Errorf("%s: the brand is not valid:\n%w", source, err)
	}

	return brand, nil
}
