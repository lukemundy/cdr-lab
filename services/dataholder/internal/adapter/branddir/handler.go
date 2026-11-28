package branddir

import (
	"github.com/lukemundy/cdr-lab/pkg/platform/asset"
	shared "github.com/lukemundy/cdr-lab/pkg/platform/branddir"
)

// consoleProperties are the only roles of the brand in the console: the roles of the console shell, the link and the
// primary action (web/packages/ui/design/tokens-and-brands.md, UI-0035). The console uses the lab brand for the rest.
var consoleProperties = []string{
	"--ds-color-brand", "--ds-color-fg-on-brand", "--ds-color-hero", "--ds-color-fg-on-hero", "--ds-color-accent",
	"--ds-color-heading", "--ds-color-link", "--ds-color-primary", "--ds-color-primary-hover",
	"--ds-color-primary-pressed", "--ds-color-fg-on-primary",
}

// RegisterSite adds the assets of the pages and the photos of the online bank site.
func (b Brand) RegisterSite(routes shared.Routes) {
	b.Register(routes)

	for name, data := range b.photos {
		routes.Handle("GET /brand/"+name, asset.Handler("image/webp", data, ""))
	}
}

// RegisterConsole adds the brand assets of the console, because the console has a different origin from the pages.
func (b Brand) RegisterConsole(routes shared.Routes) {
	b.RegisterImages(routes)
	routes.Handle("GET /brand/console.css", asset.Handler("text/css; charset=utf-8", b.StylesOf(consoleProperties), ""))
}
