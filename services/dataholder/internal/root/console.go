package root

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
)

func consoleListeners(
	cfg consolehttp.Config,
	ui fs.FS,
	ports console.Ports,
) ([]httpserver.Listener, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	mux := http.NewServeMux()
	if err := console.Register(mux, ui, cfg.AllowedHosts, ports); err != nil {
		return nil, fmt.Errorf("serve the operator console: %w", err)
	}

	return []httpserver.Listener{{Port: consolePort, Handler: telemetry.NewHTTPHandler(mux)}}, nil
}
