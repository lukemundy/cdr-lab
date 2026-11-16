package branddir_test

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type webpKind int

const (
	lossy webpKind = iota
	lossless
	extended
	animated
)

// webp gives the header of a WebP file (RFC 9649) with the size of the image, and size bytes in all. The data holder
// reads only the header.
func webp(kind webpKind, width, height, size int) []byte {
	var chunk []byte

	switch kind {
	case lossy:
		chunk = []byte("VP8 \x0a\x00\x00\x00\x00\x00\x00\x9d\x01\x2a")
		chunk = binary.LittleEndian.AppendUint16(chunk, uint16(width))
		chunk = binary.LittleEndian.AppendUint16(chunk, uint16(height))
	case lossless:
		chunk = []byte("VP8L\x05\x00\x00\x00\x2f")
		chunk = binary.LittleEndian.AppendUint32(chunk, uint32(width-1)|uint32(height-1)<<14)
	case extended, animated:
		flags := byte(0)
		if kind == animated {
			flags = 0x02
		}

		chunk = []byte{'V', 'P', '8', 'X', 10, 0, 0, 0, flags, 0, 0, 0}
		chunk = append(chunk, uint24(width-1)...)
		chunk = append(chunk, uint24(height-1)...)
	}

	data := []byte("RIFF\x00\x00\x00\x00WEBP")
	data = append(data, chunk...)
	data = append(data, bytes.Repeat([]byte{0}, max(0, size-len(data)))...)
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)-8))

	return data
}

func uint24(v int) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16)}
}

func withPhotos(t *testing.T, photos map[string][]byte) fstest.MapFS {
	t.Helper()

	dir := validDir(t)
	for name, data := range photos {
		dir[name] = &fstest.MapFile{Data: data}
	}

	return dir
}

// testdata/valid has the layout meadow, which has a hero photo of 4:3.
func TestLoadGivesTheLayoutAndThePhotosOfTheBrand(t *testing.T) {
	dir := withPhotos(t, map[string][]byte{
		"hero.webp":                   webp(extended, 1600, 1200, 1000),
		"category-term-deposits.webp": webp(lossy, 1000, 750, 1000),
		"category-margin-loans.webp":  webp(lossless, 800, 600, 1000),
		"category-trade-finance.webp": webp(extended, 1200, 900, 1000),
		"..2026_11_03/hero.webp":      webp(animated, 1, 1, 1000),
	})

	brand, err := load(t, dir)

	require.NoError(t, err)
	assert.Equal(t, "meadow", brand.Layout)
	assert.Equal(t, []string{
		"category-margin-loans.webp", "category-term-deposits.webp", "category-trade-finance.webp", "hero.webp",
	}, brand.Photos())
}

func TestAValidDirectoryWithNoPhotosHasNoPhotos(t *testing.T) {
	brand, err := load(t, validDir(t))

	require.NoError(t, err)
	assert.Empty(t, brand.Photos())
}

// Verifies: DH-0335
func TestTheBrandFileHasALayoutOfTheSite(t *testing.T) {
	for name, file := range map[string]string{
		"missing": `{"brandName": "Coastal Bank", "bsb": "376100"}`,
		"unknown": `{"brandName": "Coastal Bank", "bsb": "376100", "layout": "Meadow"}`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["brand.json"] = &fstest.MapFile{Data: []byte(file)}

			_, err := load(t, dir)

			require.ErrorContains(t, err, `brand.json: layout: give one of "meadow", "ledger", "horizon", "updraft"`)
		})
	}
}

func TestEachLayoutHasTheAspectRatioOfItsHeroPhoto(t *testing.T) {
	for layout, size := range map[string][2]int{
		"meadow": {1600, 1200}, "ledger": {2100, 900}, "horizon": {2400, 800}, "updraft": {1200, 1200},
	} {
		t.Run(layout, func(t *testing.T) {
			dir := withPhotos(t, map[string][]byte{"hero.webp": webp(lossy, size[0], size[1], 1000)})
			dir["brand.json"] = &fstest.MapFile{
				Data: []byte(`{"brandName": "Coastal Bank", "bsb": "376100", "layout": "` + layout + `"}`),
			}

			brand, err := load(t, dir)

			require.NoError(t, err)
			assert.Equal(t, layout, brand.Layout)
		})
	}
}

// Verifies: DH-0351, DH-0352, DH-0353
func TestEachPhotoIsAWebPImageOfItsSizeAndAspectRatio(t *testing.T) {
	tests := map[string]struct {
		file string
		data []byte
		want string
	}{
		"not WebP": {
			"hero.webp", append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 40)...),
			"hero.webp: the file is not a WebP image",
		},
		"short": {"hero.webp", []byte("RIFF\x04\x00\x00\x00WEBP"), "hero.webp: the file is not a WebP image"},
		"animated": {
			"hero.webp", webp(animated, 1600, 1200, 100), "hero.webp: give an image that is not animated",
		},
		"hero too narrow": {
			"hero.webp", webp(lossy, 1000, 750, 100),
			"hero.webp: the image is 1000 pixels wide: give an image 1200 to 2400 pixels wide",
		},
		"hero too wide": {
			"hero.webp", webp(lossless, 2800, 2100, 100),
			"hero.webp: the image is 2800 pixels wide: give an image 1200 to 2400 pixels wide",
		},
		"hero aspect ratio": {
			"hero.webp", webp(lossy, 1600, 1000, 100),
			"hero.webp: the aspect ratio of the image is 1.6, and the layout meadow needs 1.33, " +
				"with a tolerance of 2 %",
		},
		"category too narrow": {
			"category-term-deposits.webp", webp(lossy, 640, 480, 100),
			"category-term-deposits.webp: the image is 640 pixels wide: give an image 800 to 1200 pixels wide",
		},
		"category aspect ratio": {
			"category-term-deposits.webp", webp(extended, 1000, 1000, 100),
			"category-term-deposits.webp: the aspect ratio of the image is 1, and a category photo needs 1.33, " +
				"with a tolerance of 2 %",
		},
		"hero size": {
			"hero.webp", webp(lossy, 1600, 1200, 256<<10+1), "hero.webp: the file is more than 256 KiB",
		},
		"category size": {
			"category-term-deposits.webp", webp(lossy, 1000, 750, 96<<10+1),
			"category-term-deposits.webp: the file is more than 96 KiB",
		},
		"unknown category": {
			"category-pets.webp", webp(lossy, 1000, 750, 100),
			"category-pets.webp: the brand directory contract has no file of this name",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := load(t, withPhotos(t, map[string][]byte{tc.file: tc.data}))

			require.ErrorContains(t, err, tc.want)
		})
	}
}

// Verifies: DH-0352, DH-0353
func TestAPhotoCanBeAtItsLimits(t *testing.T) {
	_, err := load(t, withPhotos(t, map[string][]byte{
		"hero.webp":                   webp(lossy, 2400, 1800, 256<<10),
		"category-term-deposits.webp": webp(lossy, 800, 612, 96<<10),
		"category-overdrafts.webp":    webp(lossy, 1200, 883, 96<<10),
	}))

	require.NoError(t, err)
}

// A brand directory can be one ConfigMap, which is at most 1 MiB.
// Verifies: DH-0357
func TestTheBrandDirectoryIs1MiBOrLessInAll(t *testing.T) {
	dir := validDir(t)
	tokens := dir["tokens.json"].Data
	dir["tokens.json"] = &fstest.MapFile{Data: append(tokens, bytes.Repeat([]byte(" "), 900<<10-len(tokens))...)}
	dir["hero.webp"] = &fstest.MapFile{Data: webp(lossy, 1600, 1200, 200<<10)}

	_, err := load(t, dir)

	require.ErrorContains(t, err, "the files of the brand directory are more than 1 MiB in all: give 1 MiB or less")
}

// Verifies: DH-0341
func TestTheTokensHaveAHeadingFontOfTheDataHolder(t *testing.T) {
	dir := validDir(t)
	changeTokens(t, dir, func(tokens map[string]map[string]any) { setValue(tokens, "font", "heading", "Papyrus") })

	_, err := load(t, dir)

	require.ErrorContains(t, err, `tokens.json: font.heading: "Papyrus" is not a font of the data holder`)
}

// Verifies: DH-2071, DH-2072
func TestTheSiteGetsEachPhotoOfTheBrand(t *testing.T) {
	hero := webp(lossy, 1600, 1200, 1000)

	brand, err := load(t, withPhotos(t, map[string][]byte{"hero.webp": hero}))
	require.NoError(t, err)

	mux := http.NewServeMux()
	brand.RegisterSite(mux)

	rec := get(t, mux, "/brand/hero.webp", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/webp", rec.Header().Get("Content-Type"))
	assert.Equal(t, hero, rec.Body.Bytes())
	assert.Equal(t, http.StatusOK, get(t, mux, "/brand/styles.css", nil).Code)
	assert.Equal(t, http.StatusNotFound, get(t, mux, "/brand/category-term-deposits.webp", nil).Code)
}

func TestThePublicListenerGetsNoPhoto(t *testing.T) {
	rec := get(t, serveBrand(t, withPhotos(t, map[string][]byte{"hero.webp": webp(lossy, 1600, 1200, 1000)})),
		"/brand/hero.webp", nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
