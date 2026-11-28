package branddir

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// PhotoRule is the size of a photo that a service gives.
type PhotoRule struct {
	MinWidth, MaxWidth int
	// AspectRatio is the width divided by the height. 0 means no aspect ratio check.
	AspectRatio float64
	// RatioOwner names what sets the aspect ratio, for the error. For example, "the hero photo".
	RatioOwner string
}

var (
	errNotWebP  = errors.New("the file is not a WebP image")
	errAnimated = errors.New("give an image that is not animated")
)

// CheckPhoto checks that the photo is a WebP image that is not animated, with the width and the aspect ratio of the
// rule. It reads the header only, and does not decode the image.
func CheckPhoto(data []byte, rule PhotoRule) error {
	width, height, err := webpSize(data)
	if err != nil {
		return err
	}

	if width < rule.MinWidth || width > rule.MaxWidth {
		return fmt.Errorf(
			"the image is %d pixels wide: give an image %d to %d pixels wide", width, rule.MinWidth, rule.MaxWidth,
		)
	}

	aspectRatio := float64(width) / float64(height)
	if rule.AspectRatio != 0 && math.Abs(aspectRatio/rule.AspectRatio-1) > aspectRatioTolerance {
		return fmt.Errorf(
			"the aspect ratio of the image is %s, and %s needs %s, with a tolerance of 2 %%",
			ratio(aspectRatio), rule.RatioOwner, ratio(rule.AspectRatio),
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
