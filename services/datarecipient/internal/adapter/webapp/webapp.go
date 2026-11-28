// Package webapp serves the web app of the data recipient: the embedded web UI and its API, on the web app listener.
//
// The specification is services/datarecipient/design/web-app.md.
package webapp

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
)

// The data recipient image copies the build of web/apps/datarecipient-app into dist, so a plain Go build has no UI.
//
//go:embed all:dist
var dist embed.FS

// Settings has no EnabledEnv, because the web app is always on. Use LoadAllowedHosts, not LoadConfig.
var Settings = consolehttp.Settings{
	Name:            "web app",
	AllowedHostsEnv: "WEB_APP_ALLOWED_HOSTS",
	ExampleHost:     "localhost:8585",
}

func UI() fs.FS {
	ui, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}

	return ui
}

type Ports struct {
	// BrandFiles adds the stylesheet, the logo and the mark of the brand; a test can leave it nil.
	BrandFiles func(*httproute.Routes)
}

// Register returns an error when ui has no build of the UI or allowedHosts is empty.
func Register(mux *http.ServeMux, ui fs.FS, allowedHosts []string, ports Ports) error {
	return Settings.Register(mux, ui, allowedHosts, func(routes *httproute.Routes) {
		if ports.BrandFiles != nil {
			ports.BrandFiles(routes)
		}
	})
}
