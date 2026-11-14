package branddir

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	heroFile       = "hero.webp"
	categoryPrefix = "category-"
	photoSuffix    = ".webp"
)

var (
	heroLimit     = sizeLimit{bytes: 256 << 10, text: "256 KiB"}
	categoryLimit = sizeLimit{bytes: 96 << 10, text: "96 KiB"}
)

type photoRule struct {
	minWidth, maxWidth int
	aspectRatio        float64
	// ratioOwner names what sets the aspect ratio, for the error.
	ratioOwner string
}

// The widths give a sharp photo on a screen with a pixel ratio of 2 at the sizes of the layouts.
var (
	heroWidths   = [2]int{1200, 2400}
	categoryRule = photoRule{minWidth: 800, maxWidth: 1200, aspectRatio: 4.0 / 3, ratioOwner: "a category photo"}
	layoutNames  = []string{"meadow", "ledger", "horizon", "updraft"}
	heroRatios   = map[string]float64{"meadow": 4.0 / 3, "ledger": 21.0 / 9, "horizon": 3, "updraft": 1}
	errNotWebP   = errors.New("the file is not a WebP image")
	errAnimated  = errors.New("give an image that is not animated")
)

// photoLimit gives the size limit of a photo of the contract, and false for a name that the contract does not give.
func photoLimit(name string) (sizeLimit, bool) {
	if name == heroFile {
		return heroLimit, true
	}

	_, found := photoCategory(name)

	return categoryLimit, found
}

// photoCategory gives the CDS product category of a category photo, for example TERM_DEPOSITS for
// category-term-deposits.webp.
func photoCategory(name string) (domain.ProductCategory, bool) {
	anchor, found := strings.CutPrefix(name, categoryPrefix)
	if !found {
		return "", false
	}

	anchor, found = strings.CutSuffix(anchor, photoSuffix)
	if !found || strings.Contains(anchor, "_") {
		return "", false
	}

	return domain.ParseProductCategory(strings.ToUpper(strings.ReplaceAll(anchor, "-", "_")))
}

// checkPhoto has no aspect ratio check for a hero photo when the layout is not valid.
func checkPhoto(name string, data []byte, layout string) error {
	rule := categoryRule
	if name == heroFile {
		rule = photoRule{
			minWidth: heroWidths[0], maxWidth: heroWidths[1], aspectRatio: heroRatios[layout],
			ratioOwner: "the layout " + layout,
		}
	}

	width, height, err := webpSize(data)
	if err != nil {
		return err
	}

	if width < rule.minWidth || width > rule.maxWidth {
		return fmt.Errorf(
			"the image is %d pixels wide: give an image %d to %d pixels wide", width, rule.minWidth, rule.maxWidth,
		)
	}

	aspectRatio := float64(width) / float64(height)
	if rule.aspectRatio != 0 && math.Abs(aspectRatio/rule.aspectRatio-1) > aspectRatioTolerance {
		return fmt.Errorf(
			"the aspect ratio of the image is %s, and %s needs %s, with a tolerance of 2 %%",
			ratio(aspectRatio), rule.ratioOwner, ratio(rule.aspectRatio),
		)
	}

	return nil
}

// webpSize reads the size of the image from the first chunk of the file (RFC 9649), with no decode.
func webpSize(data []byte) (width, height int, err error) {
	const chunkData = 20

	if len(data) < chunkData+10 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return 0, 0, errNotWebP
	}

	payload := data[chunkData:]

	switch string(data[12:16]) {
	case "VP8 ":
		// A frame tag of 3 bytes and the start code 9d 01 2a, then the width and the height in 14 bits each.
		if payload[3] != 0x9d || payload[4] != 0x01 || payload[5] != 0x2a {
			return 0, 0, errNotWebP
		}

		return int(binary.LittleEndian.Uint16(payload[6:]) & 0x3fff),
			int(binary.LittleEndian.Uint16(payload[8:]) & 0x3fff), nil
	case "VP8L":
		if payload[0] != 0x2f {
			return 0, 0, errNotWebP
		}

		bits := binary.LittleEndian.Uint32(payload[1:])

		return int(bits&0x3fff) + 1, int(bits>>14&0x3fff) + 1, nil
	case "VP8X":
		const animationFlag = 0x02
		if payload[0]&animationFlag != 0 {
			return 0, 0, errAnimated
		}

		return uint24(payload[4:]) + 1, uint24(payload[7:]) + 1, nil
	default:
		return 0, 0, errNotWebP
	}
}

func uint24(b []byte) int {
	return int(b[0]) | int(b[1])<<8 | int(b[2])<<16
}
