package branddir_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serveBrand(t *testing.T, dir fstest.MapFS) *http.ServeMux {
	t.Helper()

	brand, err := load(t, dir)
	require.NoError(t, err)

	mux := http.NewServeMux()
	brand.Register(mux)

	return mux
}

func get(t *testing.T, mux *http.ServeMux, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// The pages show a logo only in an img element, where a script cannot run. The sandbox stops a script in an SVG
// file when a browser opens its URL directly.
// Verifies: BRAND-0024
func TestTheLogoAndTheMarkAreSVGFilesThatCannotRunAScript(t *testing.T) {
	dir := validDir(t)
	mux := serveBrand(t, dir)

	for path, file := range map[string]string{"/brand/logo.svg": "logo.svg", "/brand/mark.svg": "mark.svg"} {
		t.Run(path, func(t *testing.T) {
			rec := get(t, mux, path, nil)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "image/svg+xml", rec.Header().Get("Content-Type"))
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "default-src 'none'; style-src 'unsafe-inline'; sandbox",
				rec.Header().Get("Content-Security-Policy"))
			assert.Equal(t, string(dir[file].Data), rec.Body.String())
		})
	}
}

// Verifies: BRAND-0023
func TestTheStylesSetTheCustomPropertiesOfTheDesignSystemForEachTokenOfTheBrand(t *testing.T) {
	rec := get(t, serveBrand(t, validDir(t)), "/brand/styles.css", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/css; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))

	css := rec.Body.String()
	// The selector has a higher specificity than the :root of tokens.css, so the order of the stylesheets has no
	// effect.
	assert.Contains(t, css, ":root:root {\n")
	assert.Contains(t, css, "  --ds-color-page: oklch(0.985 0.002 260);\n")
	assert.Contains(t, css, "  --ds-color-backdrop: oklch(0.15 0.01 260 / 0.4);\n")
	assert.Contains(t, css, `  --ds-font-body: "Inter Variable", system-ui, sans-serif;`+"\n")
	assert.Contains(t, css, `  --ds-font-code: "JetBrains Mono Variable", ui-monospace, monospace;`+"\n")
	assert.Contains(t, css, "  --ds-radius-md: 6px;\n")
	assert.Contains(t, css, "  --ds-radius-full: 9999px;\n")
}

func TestABrowserCanUseItsCopyOfAnAssetUntilTheAssetChanges(t *testing.T) {
	mux := serveBrand(t, validDir(t))

	first := get(t, mux, "/brand/styles.css", nil)
	etag := first.Header().Get("ETag")

	require.NotEmpty(t, etag)
	assert.Equal(t, "no-cache", first.Header().Get("Cache-Control"))

	again := get(t, mux, "/brand/styles.css", http.Header{"If-None-Match": {etag}})

	assert.Equal(t, http.StatusNotModified, again.Code)
}

func TestTheBrandHasNoOtherAsset(t *testing.T) {
	rec := get(t, serveBrand(t, validDir(t)), "/brand/brand.json", nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRegisterImagesAddsOnlyTheLogoAndTheMark(t *testing.T) {
	brand, err := load(t, validDir(t))
	require.NoError(t, err)

	mux := http.NewServeMux()
	brand.RegisterImages(mux)

	assert.Equal(t, http.StatusOK, get(t, mux, "/brand/logo.svg", nil).Code)
	assert.Equal(t, http.StatusOK, get(t, mux, "/brand/mark.svg", nil).Code)
	assert.Equal(t, http.StatusNotFound, get(t, mux, "/brand/styles.css", nil).Code)
}

// A service can serve a stylesheet with only some roles of the brand, for example for a console.
// Verifies: BRAND-0023
func TestStylesOfSetsOnlyTheGivenPropertiesInTheOrderOfTheStylesheet(t *testing.T) {
	brand, err := load(t, validDir(t))
	require.NoError(t, err)

	css := string(brand.StylesOf([]string{"--ds-radius-md", "--ds-color-brand", "--ds-color-unknown"}))

	assert.Equal(t, ":root:root {\n  --ds-color-brand: oklch(1 0 0);\n  --ds-radius-md: 6px;\n}\n", css)
}
