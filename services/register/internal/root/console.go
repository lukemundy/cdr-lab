package root

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
)

func registerConsole(internal *http.ServeMux, cfg consolehttp.Config, ui fs.FS, ports console.Ports) error {
	if !cfg.Enabled {
		return nil
	}

	if err := console.Register(internal, ui, cfg.AllowedHosts, ports); err != nil {
		return fmt.Errorf("serve the operator console: %w", err)
	}

	return nil
}
