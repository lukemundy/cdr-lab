package root

import (
	"crypto/tls"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
)

// bankSiteListeners gives the site the certificate and the TLS rules of the public listener.
func bankSiteListeners(
	cfg consolehttp.Config,
	cert tls.Certificate,
	ui fs.FS,
	ports banksite.Ports,
) ([]httpserver.Listener, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	mux := http.NewServeMux()
	if err := banksite.Register(mux, ui, cfg.AllowedHosts, ports); err != nil {
		return nil, fmt.Errorf("serve the online bank site: %w", err)
	}

	return []httpserver.Listener{
		{Port: bankSitePort, Handler: telemetry.NewHTTPHandler(mux), TLSConfig: cds.TLSConfig(cert)},
	}, nil
}
