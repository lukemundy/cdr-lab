package branddir_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/branddir"
)

type webpKind int

const (
	lossy webpKind = iota
	lossless
	extended
	animated
)

// webp gives the header of a WebP file (RFC 9649) with the size of the image, and size bytes in all. The library reads
// only the header.
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

var heroRule = branddir.PhotoRule{MinWidth: 1200, MaxWidth: 2400, AspectRatio: 4.0 / 3, RatioOwner: "the hero photo"}

func TestCheckPhotoAcceptsEachKindOfWebPImageOfTheRule(t *testing.T) {
	for name, data := range map[string][]byte{
		"lossy":     webp(lossy, 1600, 1200, 100),
		"lossless":  webp(lossless, 1200, 900, 100),
		"extended":  webp(extended, 2400, 1800, 100),
		"rounding":  webp(lossy, 1200, 883, 100),
		"min width": webp(lossy, 1200, 900, 100),
	} {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, branddir.CheckPhoto(data, heroRule))
		})
	}
}

// Verifies: BRAND-0019, BRAND-0020
func TestEachPhotoIsAWebPImageThatIsNotAnimated(t *testing.T) {
	tests := map[string]struct {
		data []byte
		want string
	}{
		"not WebP": {append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 40)...), "the file is not a WebP image"},
		"short":    {[]byte("RIFF\x04\x00\x00\x00WEBP"), "the file is not a WebP image"},
		"animated": {webp(animated, 1600, 1200, 100), "give an image that is not animated"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.EqualError(t, branddir.CheckPhoto(tc.data, heroRule), tc.want)
		})
	}
}

func TestCheckPhotoNamesTheWidthAndTheAspectRatioOfTheRule(t *testing.T) {
	tests := map[string]struct {
		data []byte
		want string
	}{
		"too narrow": {
			webp(lossy, 1000, 750, 100), "the image is 1000 pixels wide: give an image 1200 to 2400 pixels wide",
		},
		"too wide": {
			webp(lossless, 2800, 2100, 100), "the image is 2800 pixels wide: give an image 1200 to 2400 pixels wide",
		},
		"aspect ratio": {
			webp(lossy, 1600, 1000, 100),
			"the aspect ratio of the image is 1.6, and the hero photo needs 1.33, with a tolerance of 2 %",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.EqualError(t, branddir.CheckPhoto(tc.data, heroRule), tc.want)
		})
	}
}

func TestARuleWithNoAspectRatioChecksOnlyTheWidth(t *testing.T) {
	rule := branddir.PhotoRule{MinWidth: 1200, MaxWidth: 2400}

	assert.NoError(t, branddir.CheckPhoto(webp(lossy, 1600, 100, 100), rule))
}
