package branddir

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

const svgNamespace = "http://www.w3.org/2000/svg"

// checkSVG checks the structure only, and gives the aspect ratio of the viewBox. The data holder serves the file in a
// sandbox, so a script in it cannot run.
func checkSVG(data []byte, square bool) (float64, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	var root *xml.StartElement

	depth := 0

	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return 0, errors.New("the file is not well-formed XML: " + err.Error())
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				if root != nil {
					return 0, errors.New("the file has more than one root element")
				}

				root = &t
			}

			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.TrimSpace(t)) > 0 {
				return 0, errors.New("the file has text outside the root element")
			}
		}
	}

	if root == nil {
		return 0, errors.New("the file has no root element")
	}

	if root.Name.Space != svgNamespace || root.Name.Local != "svg" {
		return 0, errors.New("the root element is not svg in the SVG namespace " + svgNamespace)
	}

	return checkViewBox(*root, square)
}

func checkViewBox(svg xml.StartElement, square bool) (float64, error) {
	var viewBox string

	found := false

	for _, attr := range svg.Attr {
		if attr.Name.Space == "" && attr.Name.Local == "viewBox" {
			viewBox, found = attr.Value, true
		}
	}

	if !found {
		return 0, errors.New("the svg element has no viewBox")
	}

	fields := strings.FieldsFunc(
		viewBox,
		func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r' },
	)
	if len(fields) != 4 {
		return 0, errors.New("the viewBox is not 4 finite numbers")
	}

	numbers := make([]float64, 4)

	for i, field := range fields {
		n, err := strconv.ParseFloat(field, 64)
		if err != nil || math.IsInf(n, 0) || math.IsNaN(n) {
			return 0, errors.New("the viewBox is not 4 finite numbers")
		}

		numbers[i] = n
	}

	width, height := numbers[2], numbers[3]

	switch {
	case width <= 0 || height <= 0:
		return 0, errors.New("the viewBox has a width or a height of 0 or less")
	case square && width != height:
		return 0, errors.New("the viewBox is not square: the mark is a square symbol")
	}

	return width / height, nil
}
