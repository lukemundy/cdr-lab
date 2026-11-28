package branddir

import (
	"net/http"
	"slices"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/platform/asset"
)

// svgPolicy stops scripts and other resources when a browser opens the URL of an SVG file directly. A page shows the
// file only in an img element, where a script cannot run.
const svgPolicy = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

type Routes interface {
	Handle(pattern string, handler http.Handler)
}

// Register adds the logo, the mark and the styles of the brand. They contain no CDR data, so they need no
// authentication.
func (d Directory) Register(routes Routes) {
	d.RegisterImages(routes)
	routes.Handle("GET /brand/styles.css", asset.Handler("text/css; charset=utf-8", d.Styles(), ""))
}

// RegisterImages adds only the logo and the mark.
func (d Directory) RegisterImages(routes Routes) {
	routes.Handle("GET /brand/logo.svg", asset.Handler("image/svg+xml", d.logo, svgPolicy))
	routes.Handle("GET /brand/mark.svg", asset.Handler("image/svg+xml", d.mark, svgPolicy))
}

// Styles gives the stylesheet that sets each custom property of the design system with the value of the brand.
func (d Directory) Styles() []byte {
	return stylesheet(d.declarations)
}

// StylesOf gives the stylesheet with only the given custom properties, for example "--ds-color-brand".
func (d Directory) StylesOf(properties []string) []byte {
	return stylesheet(slices.DeleteFunc(slices.Clone(d.declarations), func(dec declaration) bool {
		return !slices.Contains(properties, dec.property)
	}))
}

// stylesheet uses the selector :root:root, which has a higher specificity than the :root of the design system. Thus
// the order of the stylesheets on a page has no effect.
func stylesheet(declarations []declaration) []byte {
	var css strings.Builder

	css.WriteString(":root:root {\n")

	for _, dec := range declarations {
		css.WriteString("  " + dec.property + ": " + dec.value + ";\n")
	}

	css.WriteString("}\n")

	return []byte(css.String())
}
