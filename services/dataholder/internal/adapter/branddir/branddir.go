// Package branddir reads a brand directory, and serves the logo, the mark and the styles of the brand to the pages and
// to the operator console.
//
// ADR 0023 (docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md) gives the contract.
package branddir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	brandFile       = "brand.json"
	tokensFile      = "tokens.json"
	logoFile        = "logo.svg"
	logoPNGFile     = "logo.png"
	markFile        = "mark.svg"
	productsFile    = "products.json"
	attributionFile = "ATTRIBUTION.md"

	maxBrandNameLength = 200
)

type sizeLimit struct {
	bytes int64
	text  string
}

// The JSON limit and the limit of the full directory are the size of a ConfigMap, which can hold a brand directory.
var (
	imageLimit       = sizeLimit{bytes: 100 << 10, text: "100 KiB"}
	jsonLimit        = sizeLimit{bytes: 1 << 20, text: "1 MiB"}
	directoryLimit   = jsonLimit
	attributionLimit = sizeLimit{bytes: 16 << 10, text: "16 KiB"}
)

var contractFiles = map[string]sizeLimit{
	brandFile: jsonLimit, tokensFile: jsonLimit, logoFile: imageLimit, logoPNGFile: imageLimit, markFile: imageLimit,
	productsFile: jsonLimit,
}

type Brand struct {
	// Name is the brandName of the brand on the Register.
	Name string
	// BSB is the bsb of the accounts that a consumer opens on the online bank site.
	BSB string
	// Layout is the layout of the online bank site.
	Layout string
	// Products is the product catalogue of the brand, each product as JSON text.
	Products [][]byte
	// LogoPNG is the logo for clients that do not show SVG, for example email clients.
	LogoPNG []byte
	// Colours has each opaque colour role of the tokens, in hex notation, for example "#1d4ed8".
	Colours map[string]string

	logo          []byte
	mark          []byte
	styles        []byte
	consoleStyles []byte
	// photos has the optional photos of the online bank site, by file name.
	photos map[string][]byte
}

// Photos gives the file name of each photo of the brand, in order.
func (b Brand) Photos() []string {
	return slices.Sorted(maps.Keys(b.photos))
}

// ProductCheck validates the products of the catalogue. Its error names each product and rule that fails.
type ProductCheck func(products [][]byte) error

// Load validates each file of the directory at the root of dir, and gives one error that names each file and rule that
// fails.
func Load(dir fs.FS, checkProducts ProductCheck) (Brand, error) {
	files, errs := readFiles(dir)

	var b Brand

	if data, found := files[brandFile]; found {
		fields, err := parseBrandFile(data)
		errs = appendFileError(errs, brandFile, err)
		b.Name, b.BSB, b.Layout = fields.BrandName, fields.BSB, fields.Layout
	}

	b.photos = map[string][]byte{}

	for _, name := range slices.Sorted(maps.Keys(files)) {
		if _, isPhoto := photoLimit(name); isPhoto {
			errs = appendFileError(errs, name, checkPhoto(name, files[name], b.Layout))
			b.photos[name] = files[name]
		}
	}

	if data, found := files[tokensFile]; found {
		styles, consoleStyles, tokenErrs := parseTokens(data)
		for _, err := range tokenErrs {
			errs = appendFileError(errs, tokensFile, err)
		}

		b.styles = styles
		b.consoleStyles = consoleStyles

		if len(tokenErrs) == 0 {
			b.Colours = hexColours(data)
		}
	}

	if data, found := files[logoFile]; found {
		aspectRatio, err := checkSVG(data, false)
		errs = appendFileError(errs, logoFile, err)
		b.logo = data

		if logoPNG, found := files[logoPNGFile]; found && err == nil {
			errs = appendFileError(errs, logoPNGFile, checkPNG(logoPNG, aspectRatio))
			b.LogoPNG = logoPNG
		}
	}

	if data, found := files[markFile]; found {
		_, err := checkSVG(data, true)
		errs = appendFileError(errs, markFile, err)
		b.mark = data
	}

	if data, found := files[productsFile]; found {
		products, err := ReadProducts(bytes.NewReader(data))
		if err == nil {
			err = checkProducts(products)
		}

		errs = appendFileError(errs, productsFile, err)
		b.Products = products
	}

	if err := errors.Join(errs...); err != nil {
		return Brand{}, err
	}

	return b, nil
}

// readFiles reads each file of the contract, and gives an error for each file that is missing or that the contract does
// not name.
func readFiles(dir fs.FS) (map[string][]byte, []error) {
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, []error{fmt.Errorf("read the brand directory: %w", err)}
	}

	var errs []error

	files := map[string][]byte{}
	present := map[string]bool{}

	var total int64

	for _, entry := range entries {
		name := entry.Name()
		// A Kubernetes ConfigMap volume adds entries with these names, and they are not files of the brand.
		if strings.HasPrefix(name, "..") {
			continue
		}

		limit, found := contractFiles[name]
		if !found {
			limit, found = optionalLimit(name)
		}

		if !found {
			errs = append(errs, fmt.Errorf("%s: the brand directory contract has no file of this name", name))
			continue
		}

		present[name] = true

		data, err := readFile(dir, name, limit)
		if err != nil {
			errs = appendFileError(errs, name, err)
			continue
		}

		files[name] = data
		total += int64(len(data))
	}

	for _, name := range slices.Sorted(maps.Keys(contractFiles)) {
		if !present[name] {
			errs = append(errs, fmt.Errorf("%s: the file is missing", name))
		}
	}

	if total > directoryLimit.bytes {
		errs = append(errs, errors.New(
			"the files of the brand directory are more than "+directoryLimit.text+" in all: give "+
				directoryLimit.text+" or less",
		))
	}

	return files, errs
}

func optionalLimit(name string) (sizeLimit, bool) {
	if name == attributionFile {
		return attributionLimit, true
	}

	return photoLimit(name)
}

// readFile accepts a symbolic link to a file, which a ConfigMap volume gives for each file.
func readFile(dir fs.FS, name string, limit sizeLimit) ([]byte, error) {
	info, err := fs.Stat(dir, name)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, errors.New("the entry is not a file")
	}

	if info.Size() > limit.bytes {
		return nil, errors.New("the file is more than " + limit.text)
	}

	return fs.ReadFile(dir, name)
}

func appendFileError(errs []error, file string, err error) []error {
	if err == nil {
		return errs
	}

	return append(errs, fmt.Errorf("%s: %w", file, err))
}

type brandJSON struct {
	BrandName string `json:"brandName"`
	BSB       string `json:"bsb"`
	Layout    string `json:"layout"`
}

func parseBrandFile(data []byte) (brandJSON, error) {
	var b brandJSON
	if err := decodeStrict(data, &b); err != nil {
		return brandJSON{}, err
	}

	if err := checkBrandFields(b); err != nil {
		return brandJSON{}, err
	}

	return b, nil
}

func checkBrandFields(b brandJSON) error {
	if utf8.RuneCountInString(b.BrandName) > maxBrandNameLength ||
		strings.TrimFunc(b.BrandName, unicode.IsSpace) == "" {
		return fmt.Errorf("brandName: give 1 to %d characters, and not only white space", maxBrandNameLength)
	}

	// The name goes into the From header of the email, where a line break can start a new header field.
	if strings.ContainsFunc(b.BrandName, unicode.IsControl) {
		return errors.New("brandName: give no control characters")
	}

	if !domain.ValidBSB(b.BSB) {
		return errors.New("bsb: give 6 digits")
	}

	if !slices.Contains(layoutNames, b.Layout) {
		quoted := make([]string, len(layoutNames))
		for i, name := range layoutNames {
			quoted[i] = strconv.Quote(name)
		}

		return errors.New("layout: give one of " + strings.Join(quoted, ", "))
	}

	return nil
}

// decodeStrict refuses a field that v does not have, and data after the JSON value.
func decodeStrict(data []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return err
	}

	if decoder.More() {
		return errors.New("the file has data after the JSON value")
	}

	return nil
}
