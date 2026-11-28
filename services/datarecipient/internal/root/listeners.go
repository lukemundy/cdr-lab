package root

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/adapter/webapp"
)

func newListeners(ctx context.Context, pool *pgxpool.Pool, cfg config) ([]httpserver.Listener, error) {
	cert, err := cfg.webAppCert.watch(ctx, "web app")
	if err != nil {
		return nil, err
	}

	return listeners(pool, cfg, cert, webapp.UI())
}

// listeners has the internal listener, which serves only the health probes, and the web app listener.
func listeners(
	pool *pgxpool.Pool, cfg config, cert *cds.ServerCertificate, ui fs.FS,
) ([]httpserver.Listener, error) {
	internal := http.NewServeMux()
	httpserver.RegisterProbes(internal, pool.Ping)

	web := http.NewServeMux()
	if err := webapp.Register(web, ui, cfg.webAppHosts, webapp.Ports{
		BrandFiles: func(routes *httproute.Routes) { cfg.brand.Register(routes) },
	}); err != nil {
		return nil, fmt.Errorf("serve the web app: %w", err)
	}

	return []httpserver.Listener{
		{Port: internalPort, Handler: telemetry.NewHTTPHandler(internal)},
		// The web app has the TLS rules of a public listener, with a certificate from the public CA.
		{Port: webAppPort, Handler: telemetry.NewHTTPHandler(web), TLSConfig: cds.TLSConfig(cert)},
	}, nil
}
