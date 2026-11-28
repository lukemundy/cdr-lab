// Package branddir reads the brand directory of a lab service, and serves the logo, the mark and the styles of the
// brand. It validates the files that each brand directory has. The service gives its own files and their rules in a
// Spec.
//
// The specification is pkg/platform/branddir/DESIGN.md.
package branddir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	BrandFile       = "brand.json"
	TokensFile      = "tokens.json"
	LogoFile        = "logo.svg"
	LogoPNGFile     = "logo.png"
	MarkFile        = "mark.svg"
	AttributionFile = "ATTRIBUTION.md"

	maxBrandNameLength = 200
)

// Limit is the size limit of a file. Text gives the limit in an error, for example "1 MiB".
type Limit struct {
	Bytes int64
	Text  string
}

// JSONLimit is the limit of each JSON file. It and the limit of the full directory are the size of a ConfigMap, which
// can hold a brand directory.
var JSONLimit = Limit{Bytes: 1 << 20, Text: "1 MiB"}

var (
	imageLimit       = Limit{Bytes: 100 << 10, Text: "100 KiB"}
	directoryLimit   = JSONLimit
	attributionLimit = Limit{Bytes: 16 << 10, Text: "16 KiB"}
)

var sharedFiles = map[string]Limit{
	BrandFile: JSONLimit, TokensFile: JSONLimit, LogoFile: imageLimit, LogoPNGFile: imageLimit, MarkFile: imageLimit,
}

// BrandFields are the fields of brand.json that each brand directory has. The brand file of a service embeds them.
type BrandFields struct {
	// BrandName is the brandName of the brand on the Register.
	BrandName string `json:"brandName"`
}

func (f *BrandFields) brandFields() *BrandFields { return f }

// Fields is a pointer to the brand file of a service: a struct that embeds BrandFields.
type Fields interface {
	brandFields() *BrandFields
}

// Spec gives the files of a service and their rules. Each field is optional.
type Spec struct {
	// BrandFile gets the fields of brand.json. Load refuses a field that it does not have. When it is nil, brand.json
	// has only the fields of BrandFields.
	BrandFile Fields
	// CheckBrandFile gives the error of a rule of the service for the fields of BrandFile. Load calls it only when the
	// shared rules of brand.json pass.
	CheckBrandFile func() error
	// Files has each required file of the service, with its size limit.
	Files map[string]Limit
	// Optional gives the size limit of an optional file of the service, and false for a name that the service does
	// not give.
	Optional func(name string) (Limit, bool)
	// Check gives an error for each rule of the files of the service that fails. Make each error with FileError.
	Check func(d Directory) []error
}

type Directory struct {
	// Name is the brandName of the brand on the Register.
	Name string
	// LogoPNG is the logo for clients that do not show SVG, for example email clients.
	LogoPNG []byte
	// Colours has each opaque colour role of the tokens, in hex notation, for example "#1d4ed8".
	Colours map[string]string
	// Files has the content of each file of the service that the directory has, by file name.
	Files map[string][]byte

	logo         []byte
	mark         []byte
	declarations []declaration
}

// Load validates each file of the directory at the root of dir, and gives one error that names each file and rule that
// fails.
func Load(dir fs.FS, spec Spec) (Directory, error) {
	files, errs := readFiles(dir, spec)

	var d Directory

	if data, found := files[BrandFile]; found {
		name, err := parseBrandFile(data, spec)
		errs = appendFileError(errs, BrandFile, err)
		d.Name = name
	}

	if data, found := files[TokensFile]; found {
		declarations, tokenErrs := parseTokens(data)
		for _, err := range tokenErrs {
			errs = appendFileError(errs, TokensFile, err)
		}

		d.declarations = declarations

		if len(tokenErrs) == 0 {
			d.Colours = hexColours(data)
		}
	}

	if data, found := files[LogoFile]; found {
		aspectRatio, err := checkSVG(data, false)
		errs = appendFileError(errs, LogoFile, err)
		d.logo = data

		if logoPNG, found := files[LogoPNGFile]; found && err == nil {
			errs = appendFileError(errs, LogoPNGFile, checkPNG(logoPNG, aspectRatio))
			d.LogoPNG = logoPNG
		}
	}

	if data, found := files[MarkFile]; found {
		_, err := checkSVG(data, true)
		errs = appendFileError(errs, MarkFile, err)
		d.mark = data
	}

	d.Files = map[string][]byte{}

	for name, data := range files {
		if _, shared := sharedFiles[name]; !shared && name != AttributionFile {
			d.Files[name] = data
		}
	}

	if spec.Check != nil {
		errs = append(errs, spec.Check(d)...)
	}

	if err := errors.Join(errs...); err != nil {
		return Directory{}, err
	}

	return d, nil
}

// FileError gives err with the name of the file, or nil when err is nil.
func FileError(file string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", file, err)
}

func appendFileError(errs []error, file string, err error) []error {
	if err := FileError(file, err); err != nil {
		return append(errs, err)
	}

	return errs
}

// readFiles reads each file of the contract, and gives an error for each file that is missing or that the contract does
// not name.
func readFiles(dir fs.FS, spec Spec) (map[string][]byte, []error) {
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, []error{fmt.Errorf("read the brand directory: %w", err)}
	}

	required := maps.Clone(sharedFiles)
	maps.Copy(required, spec.Files)

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

		limit, found := required[name]
		if !found {
			limit, found = optionalLimit(name, spec)
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

	for _, name := range slices.Sorted(maps.Keys(required)) {
		if !present[name] {
			errs = append(errs, fmt.Errorf("%s: the file is missing", name))
		}
	}

	if total > directoryLimit.Bytes {
		errs = append(errs, errors.New(
			"the files of the brand directory are more than "+directoryLimit.Text+" in all: give "+
				directoryLimit.Text+" or less",
		))
	}

	return files, errs
}

func optionalLimit(name string, spec Spec) (Limit, bool) {
	if name == AttributionFile {
		return attributionLimit, true
	}

	if spec.Optional == nil {
		return Limit{}, false
	}

	return spec.Optional(name)
}

// readFile accepts a symbolic link to a file, which a ConfigMap volume gives for each file.
func readFile(dir fs.FS, name string, limit Limit) ([]byte, error) {
	info, err := fs.Stat(dir, name)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, errors.New("the entry is not a file")
	}

	if info.Size() > limit.Bytes {
		return nil, errors.New("the file is more than " + limit.Text)
	}

	return fs.ReadFile(dir, name)
}

// parseBrandFile gives the brand name when each rule of the brand file passes.
func parseBrandFile(data []byte, spec Spec) (string, error) {
	file := spec.BrandFile
	if file == nil {
		file = &BrandFields{}
	}

	if err := decodeStrict(data, file); err != nil {
		return "", err
	}

	name := file.brandFields().BrandName
	if err := checkBrandName(name); err != nil {
		return "", err
	}

	if spec.CheckBrandFile != nil {
		if err := spec.CheckBrandFile(); err != nil {
			return "", err
		}
	}

	return name, nil
}

func checkBrandName(name string) error {
	if utf8.RuneCountInString(name) > maxBrandNameLength || strings.TrimFunc(name, unicode.IsSpace) == "" {
		return fmt.Errorf("brandName: give 1 to %d characters, and not only white space", maxBrandNameLength)
	}

	// The name goes into the From header of an email, where a line break can start a new header field.
	if strings.ContainsFunc(name, unicode.IsControl) {
		return errors.New("brandName: give no control characters")
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
