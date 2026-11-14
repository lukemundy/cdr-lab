package consumerpages_test

import (
	"flag"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/brands"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/branddir"
)

// Golden files: a test writes its output to testdata, and fails when the output is different from the file. Run the
// test with -update to write the files again, and review the change in the diff. The browser tests of the design
// system (web/packages/ui/src/consumer-pages.test.ts) check the accessibility and the reflow of these pages.
var update = flag.Bool("update", false, "write the golden files in testdata")

func golden(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", name)

	if *update {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o600))

		return
	}

	want, err := fs.ReadFile(os.DirFS("testdata"), name)
	require.NoError(t, err, "run the test with -update to write the golden file")
	assert.Equal(t, string(want), got, "run the test with -update, and review the change")
}

// stableBody replaces the random form token, so that a golden file is the same in each run.
func stableBody(body string) string {
	return hiddenToken.ReplaceAllString(body, `<input type="hidden" name="csrf_token" value="form-token">`)
}

// goldenPages gives each page and each state of a page that looks different.
func goldenPages(t *testing.T) map[string]page {
	t.Helper()

	pages := eachPage(t, bankName)

	b := newBrowser(t)
	b.identify(knownCustomer)
	pages["one-time password incorrect"] = b.enter("000000")

	b = newBrowser(t)
	b.identify(knownCustomer)
	b.act("resend")
	pages["one-time password no resend"] = b.act("resend")

	b = newBrowser(t)
	b.sharingDuration = 365 * 24 * time.Hour
	b.scopes = []string{"openid", "bank:accounts.basic:read", "bank:transactions:read", "common:customer.basic:read"}
	b.consent()
	pages["consent no account selected"] = b.confirm()

	b = newBrowser(t)
	b.customers.accounts = nil
	pages["consent not eligible"] = b.consent()

	b = newBrowser(t)
	b.scopes = []string{"openid", "common:customer.basic:read"}
	pages["consent no accounts"] = b.consent()

	return pages
}

func TestThePagesAreTheGoldenFiles(t *testing.T) {
	var names []string

	for name, p := range goldenPages(t) {
		file := strings.ReplaceAll(name, " ", "-") + ".html"
		names = append(names, file)

		t.Run(name, func(t *testing.T) {
			golden(t, "pages/"+file, stableBody(p.body))
		})
	}

	noStaleGoldenFiles(t, "pages", names)
}

// noStaleGoldenFiles fails for a golden file in the directory of testdata that no test writes. -update removes it.
func noStaleGoldenFiles(t *testing.T, dir string, names []string) {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("testdata", dir))
	require.NoError(t, err)

	for _, entry := range entries {
		if slices.Contains(names, entry.Name()) {
			continue
		}

		if *update {
			require.NoError(t, os.Remove(filepath.Join("testdata", dir, entry.Name())))
			continue
		}

		t.Errorf("testdata/%s/%s is a golden file that no test writes: run the test with -update", dir, entry.Name())
	}
}

// The browser tests show the pages with the styles of each preset brand.
func TestTheStylesOfEachPresetBrandAreTheGoldenFiles(t *testing.T) {
	var names []string

	for _, name := range brands.Names() {
		names = append(names, name+".css")

		t.Run(name, func(t *testing.T) {
			dir, err := brands.Preset(name)
			require.NoError(t, err)

			brand, err := branddir.Load(dir, func(string, [][]byte) error { return nil })
			require.NoError(t, err)

			mux := http.NewServeMux()
			brand.Register(mux)

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/brand/styles.css", nil))
			require.Equal(t, http.StatusOK, rec.Code)

			golden(t, "brands/"+name+".css", rec.Body.String())
		})
	}

	noStaleGoldenFiles(t, "brands", names)
}
