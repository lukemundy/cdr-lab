package branddir

import (
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/asset"
)

// svgPolicy stops scripts and other resources when a browser opens the URL of an SVG file directly. A page shows the
// file only in an img element, where a script cannot run.
const svgPolicy = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

type Routes interface {
	Handle(pattern string, handler http.Handler)
}

// Register adds the assets of the brand. They contain no CDR data, so they need no authentication.
func (b Brand) Register(routes Routes) {
	b.registerImages(routes)
	routes.Handle("GET /brand/styles.css", asset.Handler("text/css; charset=utf-8", b.styles, ""))
}

// RegisterSite adds the assets of the pages and the photos of the online bank site.
func (b Brand) RegisterSite(routes Routes) {
	b.Register(routes)

	for name, data := range b.photos {
		routes.Handle("GET /brand/"+name, asset.Handler("image/webp", data, ""))
	}
}

// RegisterConsole adds the brand assets of the console, because the console has a different origin from the pages.
func (b Brand) RegisterConsole(routes Routes) {
	b.registerImages(routes)
	routes.Handle("GET /brand/console.css", asset.Handler("text/css; charset=utf-8", b.consoleStyles, ""))
}

func (b Brand) registerImages(routes Routes) {
	routes.Handle("GET /brand/logo.svg", asset.Handler("image/svg+xml", b.logo, svgPolicy))
	routes.Handle("GET /brand/mark.svg", asset.Handler("image/svg+xml", b.mark, svgPolicy))
}
