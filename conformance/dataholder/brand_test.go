//go:build conformance

package dataholder

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// A data recipient shows the logo of the bank that the consumer selects. A script in the SVG file must not run, also
// when a browser opens its URL directly.
func TestTheLogoURIOfTheBrandOnTheRegisterGivesAnSVGLogoThatCannotRunAScript(t *testing.T) {
	logoURI := brandOnTheRegister(t).LogoURI
	require.NotEmpty(t, logoURI)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, logoURI, nil)
	require.NoError(t, err)

	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "image/svg+xml", resp.Header.Get("Content-Type"))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "sandbox")
	assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "default-src 'none'")
}

var (
	stylesheetLink = regexp.MustCompile(`<link rel="stylesheet" href="([^"]+)">`)
	fontURL        = regexp.MustCompile(`url\(([^)]+\.woff2)\)`)
)

// getAsset requires 200 and the content type for the URL, and gives the body.
func getAsset(t *testing.T, target *url.URL, contentType string) string {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target.String(), nil)
	require.NoError(t, err)

	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, target.String())
	assert.Equal(t, contentType, resp.Header.Get("Content-Type"), target.String())

	return string(body)
}

// A data holder with no styles or fonts still works, so only a request for each file shows that the image has them.
func TestTheAuthenticationPageLoadsItsStylesheetsAndAFontFromItsOrigin(t *testing.T) {
	_, first := registered(t, registrationTestProduct).startFlow(t)

	links := stylesheetLink.FindAllStringSubmatch(first.body, -1)
	require.NotEmpty(t, links, "the page links a stylesheet")

	var fonts []*url.URL

	for _, link := range links {
		stylesheet := first.url.ResolveReference(&url.URL{Path: link[1]})
		assert.Equal(t, first.url.Host, stylesheet.Host, "the stylesheet is on the origin of the page")

		css := getAsset(t, stylesheet, "text/css; charset=utf-8")
		for _, font := range fontURL.FindAllStringSubmatch(css, -1) {
			fonts = append(fonts, stylesheet.ResolveReference(&url.URL{Path: font[1]}))
		}
	}

	require.NotEmpty(t, fonts, "a stylesheet refers to a font")

	getAsset(t, fonts[0], "font/woff2")
}
