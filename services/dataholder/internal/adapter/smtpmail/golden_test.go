package smtpmail_test

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/brands"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/branddir"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/smtpmail"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Golden files: the test writes the HTML part of the OTP email of each preset brand to testdata, and fails when the
// output is different from the file. Run the test with -update to write the files again, and review the change in the
// diff. The browser tests of the design system (web/packages/ui/src/otp-email.test.ts) check the accessibility of
// these emails.
var update = flag.Bool("update", false, "write the golden files in testdata")

func TestTheOTPEmailOfEachPresetBrandIsTheGoldenFile(t *testing.T) {
	var names []string

	for _, name := range brands.Names() {
		names = append(names, name+".html")

		t.Run(name, func(t *testing.T) {
			dir, err := brands.Preset(name)
			require.NoError(t, err)

			preset, err := branddir.Load(dir, func(string, [][]byte) error { return nil })
			require.NoError(t, err)

			colours := smtpmail.RoleColours(preset.Colours)
			for _, colour := range []string{
				colours.Brand, colours.Page, colours.Surface, colours.Border,
				colours.Text, colours.Heading, colours.Muted,
			} {
				assert.Regexp(t, `^#[0-9a-f]{6}$`, colour, "each colour of the email is a colour of the brand")
			}

			brand := smtpmail.Brand{Name: preset.Name, Logo: preset.LogoPNG, Colours: colours}
			html := send(t, brand, domain.OTPEmail("042137", "Budget App")).html

			golden(t, name+".html", html)
		})
	}

	noStaleGoldenFiles(t, names)
}

func golden(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", "emails", name)

	if *update {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o600))

		return
	}

	want, err := fs.ReadFile(os.DirFS("testdata/emails"), name)
	require.NoError(t, err, "run the test with -update to write the golden file")
	assert.Equal(t, string(want), got, "run the test with -update, and review the change")
}

// noStaleGoldenFiles fails for a golden file that no test writes. -update removes it.
func noStaleGoldenFiles(t *testing.T, names []string) {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("testdata", "emails"))
	require.NoError(t, err)

	for _, entry := range entries {
		if slices.Contains(names, entry.Name()) {
			continue
		}

		if *update {
			require.NoError(t, os.Remove(filepath.Join("testdata", "emails", entry.Name())))
			continue
		}

		t.Errorf("testdata/emails/%s is a golden file that no test writes: run the test with -update", entry.Name())
	}
}
