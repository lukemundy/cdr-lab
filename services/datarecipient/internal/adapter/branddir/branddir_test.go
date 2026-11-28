package branddir_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/datarecipient/brands"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/adapter/branddir"
)

// validDir gives a copy of testdata/valid, which each test can change.
func validDir(t *testing.T) fstest.MapFS {
	t.Helper()

	dir := fstest.MapFS{}

	entries, err := os.ReadDir("testdata/valid")
	require.NoError(t, err)

	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join("testdata/valid", entry.Name()))
		require.NoError(t, err)

		dir[entry.Name()] = &fstest.MapFile{Data: data}
	}

	return dir
}

// lossyWebP gives the header of a lossy WebP file (RFC 9649) with the size of the image, and size bytes in all.
func lossyWebP(width, height, size int) []byte {
	data := []byte("RIFF\x00\x00\x00\x00WEBPVP8 \x0a\x00\x00\x00\x00\x00\x00\x9d\x01\x2a")
	data = binary.LittleEndian.AppendUint16(data, uint16(width))
	data = binary.LittleEndian.AppendUint16(data, uint16(height))
	data = append(data, bytes.Repeat([]byte{0}, max(0, size-len(data)))...)
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)-8))

	return data
}

// Verifies: DR-0105
func TestLoadGivesTheBrandNameOfAValidDirectory(t *testing.T) {
	brand, err := branddir.Load(validDir(t))

	require.NoError(t, err)
	assert.Equal(t, "Fernleaf", brand.Name)
	assert.Empty(t, brand.Photos())
}

// Verifies: DR-0113
func TestLoadRefusesAFileOfTheDataHolderAndAMissingFile(t *testing.T) {
	dir := validDir(t)
	dir["products.json"] = &fstest.MapFile{Data: []byte(`{"products": []}`)}
	dir["category-term-deposits.webp"] = &fstest.MapFile{Data: lossyWebP(1000, 750, 100)}
	delete(dir, "logo.png")

	_, err := branddir.Load(dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "products.json: the brand directory contract has no file of this name")
	assert.Contains(
		t,
		err.Error(),
		"category-term-deposits.webp: the brand directory contract has no file of this name",
	)
	assert.Contains(t, err.Error(), "logo.png: the file is missing")
}

// Verifies: DR-0114
func TestTheBrandFileHasOnlyTheBrandName(t *testing.T) {
	for name, tc := range map[string]struct{ file, want string }{
		"a field of the data holder": {
			`{"brandName": "Fernleaf", "layout": "meadow"}`, `brand.json: json: unknown field "layout"`,
		},
		"no brand name": {`{}`, "brand.json: brandName: give 1 to 200 characters"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["brand.json"] = &fstest.MapFile{Data: []byte(tc.file)}

			_, err := branddir.Load(dir)

			require.ErrorContains(t, err, tc.want)
		})
	}
}

// Verifies: DR-0113, DR-0118
func TestTheHeroPhotoIsAWebPImageOf4By3(t *testing.T) {
	tests := map[string]struct {
		data []byte
		want string
	}{
		"too narrow": {
			lossyWebP(1000, 750, 100),
			"hero.webp: the image is 1000 pixels wide: give an image 1200 to 2400 pixels wide",
		},
		"too wide": {
			lossyWebP(2800, 2100, 100),
			"hero.webp: the image is 2800 pixels wide: give an image 1200 to 2400 pixels wide",
		},
		"aspect ratio": {
			lossyWebP(1600, 900, 100),
			"hero.webp: the aspect ratio of the image is 1.78, and the hero photo needs 1.33, with a tolerance of 2 %",
		},
		"size": {lossyWebP(1600, 1200, 256<<10+1), "hero.webp: the file is more than 256 KiB"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["hero.webp"] = &fstest.MapFile{Data: tc.data}

			_, err := branddir.Load(dir)

			require.ErrorContains(t, err, tc.want)
		})
	}
}

// Verifies: DR-0118
func TestTheHeroPhotoCanBeAtItsLimits(t *testing.T) {
	for name, data := range map[string][]byte{
		"min": lossyWebP(1200, 900, 100),
		"max": lossyWebP(2400, 1800, 256<<10),
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["hero.webp"] = &fstest.MapFile{Data: data}
			dir["ATTRIBUTION.md"] = &fstest.MapFile{Data: []byte("# Attribution\n")}

			brand, err := branddir.Load(dir)

			require.NoError(t, err)
			assert.Equal(t, []string{"hero.webp"}, brand.Photos())
		})
	}
}

// Verifies: DR-0113, DR-0127
func TestTheAccessPhotoIsAWebPImageOf4By5(t *testing.T) {
	tests := map[string]struct {
		data []byte
		want string
	}{
		"too narrow": {
			lossyWebP(640, 800, 100),
			"access.webp: the image is 640 pixels wide: give an image 800 to 1600 pixels wide",
		},
		"too wide": {
			lossyWebP(2000, 2500, 100),
			"access.webp: the image is 2000 pixels wide: give an image 800 to 1600 pixels wide",
		},
		"aspect ratio": {
			lossyWebP(1200, 900, 100),
			"access.webp: the aspect ratio of the image is 1.33, and the access photo needs 0.8, " +
				"with a tolerance of 2 %",
		},
		"size": {lossyWebP(1000, 1250, 160<<10+1), "access.webp: the file is more than 160 KiB"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["access.webp"] = &fstest.MapFile{Data: tc.data}

			_, err := branddir.Load(dir)

			require.ErrorContains(t, err, tc.want)
		})
	}
}

// Verifies: DR-0127
func TestTheAccessPhotoCanBeAtItsLimits(t *testing.T) {
	for name, data := range map[string][]byte{
		"min": lossyWebP(800, 1000, 100),
		"max": lossyWebP(1600, 2000, 160<<10),
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["access.webp"] = &fstest.MapFile{Data: data}
			dir["hero.webp"] = &fstest.MapFile{Data: lossyWebP(1600, 1200, 100)}

			brand, err := branddir.Load(dir)

			require.NoError(t, err)
			assert.Equal(t, []string{"access.webp", "hero.webp"}, brand.Photos())
		})
	}
}

// Verifies: DR-0111
func TestEachPresetBrandIsValid(t *testing.T) {
	require.Contains(t, brands.Names(), "sorrel")

	for _, name := range brands.Names() {
		t.Run(name, func(t *testing.T) {
			dir, err := brands.Preset(name)
			require.NoError(t, err)

			brand, err := branddir.Load(dir)

			require.NoError(t, err)
			assert.NotEmpty(t, brand.Name)
		})
	}
}

// Verifies: DR-0121
func TestThePresetSorrelHasTheBrandNameSorrel(t *testing.T) {
	dir, err := brands.Preset("sorrel")
	require.NoError(t, err)

	brand, err := branddir.Load(dir)

	require.NoError(t, err)
	assert.Equal(t, "Sorrel", brand.Name)
	assert.Equal(t, []string{"access.webp", "hero.webp"}, brand.Photos())
}
