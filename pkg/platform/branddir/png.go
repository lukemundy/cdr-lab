package branddir

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"math"
	"strconv"
)

const (
	// logoPNGHeight is 2 times the height of the logo in the email, for a screen with a pixel ratio of 2.
	logoPNGHeight = 96
	// maxLogoPNGWidth limits the memory that the decode of a logo can use.
	maxLogoPNGWidth = 8 * logoPNGHeight
	// aspectRatioTolerance permits the rounding of the width to a whole pixel.
	aspectRatioTolerance = 0.02
)

// checkPNG checks the size in the header first, because a small file can declare an image that needs more memory than
// the process has.
func checkPNG(data []byte, svgAspectRatio float64) error {
	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return notValidPNG(err)
	}

	if err := checkPNGSize(config.Width, config.Height, svgAspectRatio); err != nil {
		return err
	}

	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		return notValidPNG(err)
	}

	return nil
}

func checkPNGSize(width, height int, svgAspectRatio float64) error {
	switch {
	case height != logoPNGHeight:
		return fmt.Errorf("the image is %d pixels high: give an image exactly %d pixels high", height, logoPNGHeight)
	case width > maxLogoPNGWidth:
		return fmt.Errorf("the image is %d pixels wide: give an image %d pixels wide or less", width, maxLogoPNGWidth)
	}

	aspectRatio := float64(width) / float64(height)
	if math.Abs(aspectRatio/svgAspectRatio-1) > aspectRatioTolerance {
		return fmt.Errorf(
			"the aspect ratio of the image is %s, and the aspect ratio of %s is %s: give the same logo, with a "+
				"tolerance of 2 %%",
			ratio(aspectRatio), LogoFile, ratio(svgAspectRatio),
		)
	}

	return nil
}

func notValidPNG(err error) error {
	return errors.New("the file is not a valid PNG image: " + err.Error())
}

func ratio(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}
