package brands_test

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/brands"
)

func TestEachPresetIsADirectoryWithItsFilesAtTheRoot(t *testing.T) {
	require.NotEmpty(t, brands.Names())

	for _, name := range brands.Names() {
		dir, err := brands.Preset(name)
		require.NoError(t, err)

		_, err = fs.Stat(dir, "brand.json")
		assert.NoError(t, err, "preset %s", name)
	}
}

func TestTheAttributionFileOfEachPresetNamesEachPhoto(t *testing.T) {
	for _, name := range brands.Names() {
		dir, err := brands.Preset(name)
		require.NoError(t, err)

		photos, err := fs.Glob(dir, "*.webp")
		require.NoError(t, err)

		if len(photos) == 0 {
			continue
		}

		attribution, err := fs.ReadFile(dir, "ATTRIBUTION.md")
		require.NoError(t, err, "preset %s", name)

		for _, photo := range photos {
			assert.Contains(t, string(attribution), "`"+photo+"`", "preset %s", name)
		}
	}
}

func TestAnUnknownPresetNamesThePresets(t *testing.T) {
	for _, name := range []string{"nope", "../brands", ".", ""} {
		_, err := brands.Preset(name)

		require.Error(t, err)
		assert.ErrorContains(t, err, "is not a preset brand: give one of "+brands.Names()[0])
	}
}
