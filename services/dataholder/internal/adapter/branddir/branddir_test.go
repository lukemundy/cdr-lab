package branddir_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/branddir"
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

func noProductCheck(string, [][]byte) error { return nil }

func load(t *testing.T, dir fs.FS) (branddir.Brand, error) {
	t.Helper()

	return branddir.Load(dir, noProductCheck)
}

// changeTokens decodes tokens.json, applies change and encodes the file again.
func changeTokens(t *testing.T, dir fstest.MapFS, change func(tokens map[string]map[string]any)) {
	t.Helper()

	var tokens map[string]map[string]any
	require.NoError(t, json.Unmarshal(dir["tokens.json"].Data, &tokens))

	change(tokens)

	data, err := json.Marshal(tokens)
	require.NoError(t, err)

	dir["tokens.json"] = &fstest.MapFile{Data: data}
}

func setValue(tokens map[string]map[string]any, group, name string, value any) {
	tokens[group][name] = map[string]any{"$value": value}
}

func TestLoadGivesTheBrandNameAndTheProductsOfAValidDirectory(t *testing.T) {
	brand, err := load(t, validDir(t))

	require.NoError(t, err)
	assert.Equal(t, "Coastal Bank", brand.Name)
	assert.Equal(t, "376100", brand.BSB)
	require.Len(t, brand.Products, 1)
	assert.Contains(t, string(brand.Products[0]), `"productId": "everyday"`)
}

func TestLoadReadsADirectoryOnDisk(t *testing.T) {
	brand, err := load(t, os.DirFS("testdata/valid"))

	require.NoError(t, err)
	assert.Equal(t, "Coastal Bank", brand.Name)
}

// The contract has the token names of the design system, so a brand directory gives each role that the UIs use.
func TestTheTokensOfTheContractAreTheTokensOfABrandOfTheDesignSystem(t *testing.T) {
	designSystem, err := os.ReadFile("../../../../../web/packages/tokens/tokens/brand/lab.json")
	require.NoError(t, err)

	brand := validDir(t)["tokens.json"].Data

	assert.Equal(t, tokenNames(t, designSystem), tokenNames(t, brand))
}

func tokenNames(t *testing.T, data []byte) []string {
	t.Helper()

	var groups map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &groups))

	var names []string

	for group, tokens := range groups {
		for name := range tokens {
			if !strings.HasPrefix(name, "$") {
				names = append(names, group+"."+name)
			}
		}
	}

	slices.Sort(names)

	return names
}

func TestLoadRefusesAFileThatTheContractDoesNotNameAndAMissingFile(t *testing.T) {
	dir := validDir(t)
	dir["favicon.ico"] = &fstest.MapFile{Data: []byte("x")}
	dir["extra/notes.txt"] = &fstest.MapFile{Data: []byte("x")}
	delete(dir, "mark.svg")

	_, err := load(t, dir)

	require.Error(t, err)

	message := err.Error()
	assert.Contains(t, message, "favicon.ico: the brand directory contract has no file of this name")
	assert.Contains(t, message, "extra: the brand directory contract has no file of this name")
	assert.Contains(t, message, "mark.svg: the file is missing")
}

func TestLoadAcceptsAnAttributionFileOf16KiBOrLess(t *testing.T) {
	dir := validDir(t)
	dir["ATTRIBUTION.md"] = &fstest.MapFile{Data: make([]byte, 16<<10)}

	_, err := load(t, dir)

	require.NoError(t, err)

	dir["ATTRIBUTION.md"] = &fstest.MapFile{Data: make([]byte, 16<<10+1)}

	_, err = load(t, dir)

	require.ErrorContains(t, err, "ATTRIBUTION.md: the file is more than 16 KiB")
}

// A Kubernetes ConfigMap volume adds ..data and a ..<timestamp> directory, and each file is a symbolic link.
func TestLoadIgnoresTheEntriesOfAConfigMapVolume(t *testing.T) {
	dir := validDir(t)
	dir["..data/brand.json"] = &fstest.MapFile{Data: []byte("x")}
	dir["..2026_11_01_00_00_00.123/brand.json"] = &fstest.MapFile{Data: []byte("x")}

	_, err := load(t, dir)

	require.NoError(t, err)
}

// configMapDir writes the files as a ConfigMap volume does: in ..data, with a relative symbolic link for each file.
func configMapDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "..data"), 0o700))

	for name, file := range validDir(t) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "..data", name), file.Data, 0o600))
		require.NoError(t, os.Symlink(filepath.Join("..data", name), filepath.Join(dir, name)))
	}

	return dir
}

func openRoot(t *testing.T, dir string) fs.FS {
	t.Helper()

	root, err := os.OpenRoot(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = root.Close() })

	return root.FS()
}

func TestLoadAcceptsARelativeSymbolicLinkToAFileInTheDirectory(t *testing.T) {
	brand, err := load(t, openRoot(t, configMapDir(t)))

	require.NoError(t, err)
	assert.Equal(t, "Coastal Bank", brand.Name)
}

func TestLoadRefusesAFileOfTheContractThatIsADirectory(t *testing.T) {
	dir := validDir(t)
	delete(dir, "logo.svg")
	dir["logo.svg/logo.svg"] = &fstest.MapFile{Data: []byte("x")}

	_, err := load(t, dir)

	require.ErrorContains(t, err, "logo.svg: the entry is not a file")
}

func TestLoadRefusesASymbolicLinkToAnEntryThatIsNotAFile(t *testing.T) {
	dir := configMapDir(t)
	require.NoError(t, os.Remove(filepath.Join(dir, "mark.svg")))
	require.NoError(t, os.Symlink("..data", filepath.Join(dir, "mark.svg")))

	_, err := load(t, openRoot(t, dir))

	require.ErrorContains(t, err, "mark.svg: the entry is not a file")
}

type readRecorder struct {
	fstest.MapFS

	read []string
}

func (r *readRecorder) ReadFile(name string) ([]byte, error) {
	r.read = append(r.read, name)
	return r.MapFS.ReadFile(name)
}

func TestLoadRefusesAFileAboveItsSizeLimitBeforeItReadsTheFile(t *testing.T) {
	dir := validDir(t)
	dir["tokens.json"] = &fstest.MapFile{Data: make([]byte, 1<<20+1)}
	dir["products.json"] = &fstest.MapFile{Data: make([]byte, 1<<20+1)}
	dir["brand.json"] = &fstest.MapFile{Data: make([]byte, 1<<20+1)}
	dir["logo.svg"] = &fstest.MapFile{Data: []byte(paddedSVG(100<<10 + 1))}
	recorder := &readRecorder{MapFS: dir}

	_, err := load(t, recorder)

	require.Error(t, err)

	message := err.Error()
	assert.Contains(t, message, "tokens.json: the file is more than 1 MiB")
	assert.Contains(t, message, "products.json: the file is more than 1 MiB")
	assert.Contains(t, message, "brand.json: the file is more than 1 MiB")
	assert.Contains(t, message, "logo.svg: the file is more than 100 KiB")
	assert.ElementsMatch(t, []string{"logo.png", "mark.svg"}, recorder.read)
}

// The directory is then also at its limit of 1 MiB in all.
func TestAFileCanBeAtItsSizeLimit(t *testing.T) {
	dir := validDir(t)
	others := 0

	for name, file := range dir {
		if name != "tokens.json" {
			others += len(file.Data)
		}
	}

	tokens := dir["tokens.json"].Data
	dir["tokens.json"] = &fstest.MapFile{
		Data: append(tokens, bytes.Repeat([]byte(" "), 1<<20-others-len(tokens))...),
	}

	_, err := load(t, dir)

	require.NoError(t, err)
}

func TestLoadNamesEachFileAndRuleThatFailsInOneError(t *testing.T) {
	dir := validDir(t)
	dir["brand.json"] = &fstest.MapFile{Data: []byte(`{"brandName": " "}`)}
	dir["mark.svg"] = &fstest.MapFile{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`)}

	_, err := load(t, dir)

	require.Error(t, err)

	message := err.Error()
	assert.Contains(t, message, "brand.json: brandName: give 1 to 200 characters, and not only white space")
	assert.Contains(t, message, "mark.svg: the svg element has no viewBox")
}

func TestABrandNameOf200CharactersIsValid(t *testing.T) {
	name := strings.Repeat("é", 200)
	dir := validDir(t)
	dir["brand.json"] = &fstest.MapFile{
		Data: []byte(`{"brandName": "` + name + `", "bsb": "376100", "layout": "meadow"}`),
	}

	brand, err := branddir.Load(dir, noProductCheck)

	require.NoError(t, err)
	assert.Equal(t, name, brand.Name)
}

func TestTheBrandFileHasABrandNameOf1To200CharactersAndNotOnlyWhiteSpace(t *testing.T) {
	for name, file := range map[string]string{
		"201 characters":      `{"brandName": "` + strings.Repeat("a", 201) + `"}`,
		"empty":               `{"brandName": ""}`,
		"white space":         `{"brandName": " \t"}`,
		"absent":              `{}`,
		"unknown field":       `{"brandName": "Coastal Bank", "shortName": "coastal"}`,
		"data after the JSON": `{"brandName": "Coastal Bank"} {}`,
		"not JSON":            `brandName: Coastal Bank`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["brand.json"] = &fstest.MapFile{Data: []byte(file)}

			_, err := branddir.Load(dir, noProductCheck)

			require.ErrorContains(t, err, "brand.json: ")
		})
	}
}

// The bsb has the rule of a bsb of an account in the operator interface.
func TestTheBrandFileHasABSBOf6Digits(t *testing.T) {
	for name, file := range map[string]string{
		"absent":     `{"brandName": "Coastal Bank"}`,
		"empty":      `{"brandName": "Coastal Bank", "bsb": ""}`,
		"5 digits":   `{"brandName": "Coastal Bank", "bsb": "37610"}`,
		"7 digits":   `{"brandName": "Coastal Bank", "bsb": "3761000"}`,
		"hyphen":     `{"brandName": "Coastal Bank", "bsb": "376-100"}`,
		"not digits": `{"brandName": "Coastal Bank", "bsb": "37610a"}`,
		"a number":   `{"brandName": "Coastal Bank", "bsb": 376100}`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["brand.json"] = &fstest.MapFile{Data: []byte(file)}

			_, err := branddir.Load(dir, noProductCheck)

			require.ErrorContains(t, err, "brand.json: ")
		})
	}
}

func TestABSBThatIsNotValidNamesItsRule(t *testing.T) {
	dir := validDir(t)
	dir["brand.json"] = &fstest.MapFile{Data: []byte(`{"brandName": "Coastal Bank", "bsb": "376-100"}`)}

	_, err := branddir.Load(dir, noProductCheck)

	require.ErrorContains(t, err, "brand.json: bsb: give 6 digits")
}

func TestLoadGivesTheBrandNameAndTheProductsToTheProductCheck(t *testing.T) {
	failure := errors.New(`product 1 ("everyday"): brand "Other" is not the brand name`)

	var gotName string

	var gotProducts [][]byte

	_, err := branddir.Load(validDir(t), func(name string, products [][]byte) error {
		gotName, gotProducts = name, products
		return failure
	})

	assert.Equal(t, "Coastal Bank", gotName)
	assert.Len(t, gotProducts, 1)
	require.ErrorContains(t, err, `products.json: product 1 ("everyday"): brand "Other" is not the brand name`)
}

func TestLoadRefusesAProductsFileThatIsNotAProductSet(t *testing.T) {
	dir := validDir(t)
	dir["products.json"] = &fstest.MapFile{Data: []byte(`{"product": []}`)}

	_, err := load(t, dir)

	require.ErrorContains(t, err, "products.json: ")
}

func TestLoadRefusesTokensThatAreNotTheTokensOfTheContract(t *testing.T) {
	tests := map[string]struct {
		change func(tokens map[string]map[string]any)
		want   string
	}{
		"missing role": {
			func(tokens map[string]map[string]any) { delete(tokens["color"], "primary") },
			"tokens.json: color.primary: the token is missing",
		},
		"unknown role": {
			func(tokens map[string]map[string]any) {
				setValue(
					tokens,
					"color",
					"highlight",
					map[string]any{"colorSpace": "oklch", "components": []any{0.5, 0.1, 10}},
				)
			},
			"tokens.json: color.highlight: the contract has no token of this name",
		},
		"unknown group": {
			func(tokens map[string]map[string]any) { tokens["space"] = map[string]any{"$type": "dimension"} },
			"tokens.json: space: the contract has no group of this name",
		},
		"missing group": {
			func(tokens map[string]map[string]any) { delete(tokens, "radius") },
			"tokens.json: radius: the group is missing",
		},
		"wrong type": {
			func(tokens map[string]map[string]any) { tokens["radius"]["$type"] = "number" },
			`tokens.json: radius: $type is not "dimension"`,
		},
		"unknown field of a token": {
			func(tokens map[string]map[string]any) {
				tokens["color"]["page"] = map[string]any{
					"$value":      map[string]any{"colorSpace": "oklch", "components": []any{0.98, 0, 0}},
					"$extensions": map[string]any{},
				}
			},
			"tokens.json: color.page: ",
		},
		"alias": {
			func(tokens map[string]map[string]any) {
				setValue(tokens, "color", "page", "{primitive.color.neutral.50}")
			},
			"tokens.json: color.page: $value is not a literal OKLCH colour",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			changeTokens(t, dir, tc.change)

			_, err := load(t, dir)

			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestATokenAndAGroupCanHaveADescription(t *testing.T) {
	dir := validDir(t)
	changeTokens(t, dir, func(tokens map[string]map[string]any) {
		tokens["color"]["$description"] = "The colour roles."
		tokens["color"]["page"] = map[string]any{
			"$value":       map[string]any{"colorSpace": "oklch", "components": []any{0.98, 0, 0}},
			"$description": "The page background.",
		}
	})

	_, err := load(t, dir)

	require.NoError(t, err)
}

func TestEachColourIsAnOKLCHColourInTheSRGBGamut(t *testing.T) {
	tests := map[string]struct {
		value any
		want  string
	}{
		"other colour space": {
			map[string]any{"colorSpace": "srgb", "components": []any{1, 1, 1}},
			"the colour space is not oklch",
		},
		"hex": {
			map[string]any{"colorSpace": "oklch", "components": []any{1, 0, 0}, "hex": "#ffffff"},
			"$value is not a literal OKLCH colour",
		},
		"none component": {
			map[string]any{"colorSpace": "oklch", "components": []any{0.5, 0, "none"}},
			"$value is not a literal OKLCH colour",
		},
		"two components": {
			map[string]any{"colorSpace": "oklch", "components": []any{0.5, 0}},
			"give 3 components",
		},
		"lightness above 1": {
			map[string]any{"colorSpace": "oklch", "components": []any{1.2, 0, 0}},
			"the lightness is not from 0 to 1",
		},
		"negative chroma": {
			map[string]any{"colorSpace": "oklch", "components": []any{0.5, -0.1, 0}},
			"the chroma is less than 0",
		},
		"hue of 360": {
			map[string]any{"colorSpace": "oklch", "components": []any{0.5, 0.1, 360}},
			"the hue is not from 0 to less than 360",
		},
		"alpha above 1": {
			map[string]any{"colorSpace": "oklch", "components": []any{0.5, 0.1, 10}, "alpha": 1.5},
			"the alpha is not from 0 to 1",
		},
		"outside sRGB": {
			map[string]any{"colorSpace": "oklch", "components": []any{0.7, 0.4, 150}},
			"the colour is not in the sRGB gamut",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			changeTokens(
				t,
				dir,
				func(tokens map[string]map[string]any) { setValue(tokens, "color", "primary", tc.value) },
			)

			_, err := load(t, dir)

			require.ErrorContains(t, err, "tokens.json: color.primary: "+tc.want)
		})
	}
}

func TestTheGamutCheckAcceptsTheEdgesOfTheSRGBGamut(t *testing.T) {
	// White, black and the sRGB primary blue, which is at the edge of the gamut.
	for name, components := range map[string][]any{
		"white": {1, 0, 0},
		"black": {0, 0, 0},
		"blue":  {0.452014, 0.313214, 264.052},
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			changeTokens(t, dir, func(tokens map[string]map[string]any) {
				setValue(tokens, "color", "primary", map[string]any{"colorSpace": "oklch", "components": components})
			})

			_, err := load(t, dir)

			require.NoError(t, err)
		})
	}
}

func TestEachFontIsAFontOfTheDataHolderWithOnlyGenericFallbacks(t *testing.T) {
	tests := map[string]struct {
		value any
		want  string
	}{
		"font that is not there": {
			[]any{"Comic Sans MS", "sans-serif"},
			`"Comic Sans MS" is not a font of the data holder`,
		},
		"named fallback": {[]any{"Inter Variable", "Arial"}, `"Arial" is not a generic font family`},
		"generic only":   {[]any{"sans-serif"}, `"sans-serif" is not a font of the data holder`},
		"empty list":     {[]any{}, "give one font or more"},
		"number":         {12, "$value is not a font family"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			changeTokens(t, dir, func(tokens map[string]map[string]any) { setValue(tokens, "font", "body", tc.value) })

			_, err := load(t, dir)

			require.ErrorContains(t, err, "tokens.json: font.body: "+tc.want)
		})
	}
}

// The design system builds the font files of the families in fonts.json, so the data holder serves each font that a
// brand can give.
func TestTheFontsOfTheDataHolderAreTheFontsOfTheDesignSystem(t *testing.T) {
	data, err := os.ReadFile("../../../../../web/packages/tokens/fonts.json")
	require.NoError(t, err)

	var designSystem []struct {
		Family string `json:"family"`
	}
	require.NoError(t, json.Unmarshal(data, &designSystem))

	families := make([]string, len(designSystem))
	for i, font := range designSystem {
		families[i] = strconv.Quote(font.Family)
	}

	dir := validDir(t)
	changeTokens(t, dir, func(tokens map[string]map[string]any) { setValue(tokens, "font", "body", "Unknown") })

	_, err = load(t, dir)

	require.Error(t, err)
	assert.Regexp(t, `(?m)give one of `+regexp.QuoteMeta(strings.Join(families, ", "))+`$`, err.Error())
}

func TestAFontIsOneFontOrAFontWithGenericFallbacks(t *testing.T) {
	for name, value := range map[string]any{
		"one font":         "Inter Variable",
		"font and generic": []any{"JetBrains Mono Variable", "ui-monospace", "monospace"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			changeTokens(t, dir, func(tokens map[string]map[string]any) { setValue(tokens, "font", "body", value) })

			_, err := load(t, dir)

			require.NoError(t, err)
		})
	}
}

func TestEachRadiusIsAPixelOrRemDimensionOfZeroOrMore(t *testing.T) {
	tests := map[string]struct {
		value any
		want  string
	}{
		"negative":      {map[string]any{"value": -1, "unit": "px"}, "the value is less than 0"},
		"other unit":    {map[string]any{"value": 1, "unit": "em"}, `the unit is not "px" or "rem"`},
		"no value":      {map[string]any{"unit": "px"}, "$value is not a dimension"},
		"unknown field": {map[string]any{"value": 1, "unit": "px", "scale": 2}, "$value is not a dimension"},
		"string":        {"4px", "$value is not a dimension"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			changeTokens(t, dir, func(tokens map[string]map[string]any) { setValue(tokens, "radius", "sm", tc.value) })

			_, err := load(t, dir)

			require.ErrorContains(t, err, "tokens.json: radius.sm: "+tc.want)
		})
	}
}

func TestARadiusCanBeInRemAndCanBeZero(t *testing.T) {
	for name, value := range map[string]any{
		"rem":  map[string]any{"value": 0.5, "unit": "rem"},
		"zero": map[string]any{"value": 0, "unit": "px"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			changeTokens(t, dir, func(tokens map[string]map[string]any) { setValue(tokens, "radius", "sm", value) })

			_, err := load(t, dir)

			require.NoError(t, err)
		})
	}
}

func TestEachSVGFileIsAWellFormedSVGWithAViewBoxAnd100KiBOrLess(t *testing.T) {
	const ns = `xmlns="http://www.w3.org/2000/svg"`

	tests := map[string]struct {
		svg  string
		want string
	}{
		"not well-formed": {`<svg ` + ns + ` viewBox="0 0 10 2"><g></svg>`, "the file is not well-formed XML"},
		"other root": {
			`<html ` + ns + ` viewBox="0 0 10 2"/>`,
			"the root element is not svg in the SVG namespace",
		},
		"no namespace": {`<svg viewBox="0 0 10 2"/>`, "the root element is not svg in the SVG namespace"},
		"no viewBox":   {`<svg ` + ns + `/>`, "the svg element has no viewBox"},
		"viewBox of 3": {`<svg ` + ns + ` viewBox="0 0 10"/>`, "the viewBox is not 4 finite numbers"},
		"infinite width": {
			`<svg ` + ns + ` viewBox="0 0 Inf 2"/>`,
			"the viewBox is not 4 finite numbers",
		},
		"NaN height": {`<svg ` + ns + ` viewBox="0 0 10 NaN"/>`, "the viewBox is not 4 finite numbers"},
		"zero width": {
			`<svg ` + ns + ` viewBox="0 0 0 2"/>`,
			"the viewBox has a width or a height of 0 or less",
		},
		"two roots": {
			`<svg ` + ns + ` viewBox="0 0 10 2"/><svg ` + ns + ` viewBox="0 0 10 2"/>`,
			"the file has more than one root element",
		},
		"text after root": {`<svg ` + ns + ` viewBox="0 0 10 2"/>text`, "the file has text outside the root element"},
		"empty":           {``, "the file has no root element"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["logo.svg"] = &fstest.MapFile{Data: []byte(tc.svg)}

			_, err := load(t, dir)

			require.ErrorContains(t, err, "logo.svg: "+tc.want)
		})
	}
}

func TestAnSVGFileCanHaveADoctypeAndCanBe100KiB(t *testing.T) {
	for name, svg := range map[string]string{
		"with a doctype": `<?xml version="1.0"?><!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" ` +
			`"http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">` +
			`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0,0,10,2"/>`,
		"100 KiB": paddedSVG(100 << 10),
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["logo.svg"] = &fstest.MapFile{Data: []byte(svg)}

			_, err := load(t, dir)

			require.NoError(t, err)
		})
	}
}

// paddedSVG gives a valid SVG file of size bytes.
func paddedSVG(size int) string {
	start, end := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 2">`, `</svg>`

	return start + strings.Repeat(" ", size-len(start)-len(end)) + end
}

func TestTheMarkIsSquare(t *testing.T) {
	dir := validDir(t)
	dir["mark.svg"] = &fstest.MapFile{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 30"/>`)}

	_, err := load(t, dir)

	require.ErrorContains(t, err, "mark.svg: the viewBox is not square")
}

func TestTheDirectoryCanBeTheRootOfAnEmbeddedFileSystem(t *testing.T) {
	dir := fstest.MapFS{}
	for name, file := range validDir(t) {
		dir["presets/coastal/"+name] = file
	}

	sub, err := fs.Sub(dir, "presets/coastal")
	require.NoError(t, err)

	brand, err := load(t, sub)

	require.NoError(t, err)
	assert.Equal(t, "Coastal Bank", brand.Name)
}

// A line break in the brand name can start a new header field of an email.
func TestTheBrandNameHasNoControlCharacters(t *testing.T) {
	for name, brandName := range map[string]string{
		"line feed":       `Coastal\nBcc: x@example.com`,
		"carriage return": `Coastal\rBank`,
		"tab":             `Coastal\tBank`,
		"null":            `Coastal\u0000Bank`,
		"delete":          `Coastal\u007fBank`,
		"C1 next line":    `Coastal\u0085Bank`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["brand.json"] = &fstest.MapFile{
				Data: []byte(`{"brandName": "` + brandName + `", "bsb": "376100", "layout": "meadow"}`),
			}

			_, err := load(t, dir)

			require.ErrorContains(t, err, "brand.json: brandName: give no control characters")
		})
	}
}

// pngOf gives a valid PNG image of the size.
func pngOf(t *testing.T, width, height int) []byte {
	t.Helper()

	var data bytes.Buffer
	require.NoError(t, png.Encode(&data, image.NewNRGBA(image.Rect(0, 0, width, height))))

	return data.Bytes()
}

func TestLoadGivesTheLogoPNG(t *testing.T) {
	dir := validDir(t)

	brand, err := load(t, dir)

	require.NoError(t, err)
	assert.Equal(t, dir["logo.png"].Data, brand.LogoPNG)
}

// The viewBox of testdata/valid/logo.svg is 160 by 32: an aspect ratio of 5.
func TestTheLogoPNGIsAValidPNG96PixelsHighWithTheAspectRatioOfTheSVGLogo(t *testing.T) {
	valid := pngOf(t, 480, 96)

	tests := map[string]struct {
		data []byte
		want string
	}{
		"not a PNG":      {[]byte("GIF89a"), "the file is not a valid PNG image"},
		"truncated":      {valid[:len(valid)/2], "the file is not a valid PNG image"},
		"48 pixels high": {pngOf(t, 240, 48), "the image is 48 pixels high: give an image exactly 96 pixels high"},
		"too narrow": {
			pngOf(t, 460, 96),
			"the aspect ratio of the image is 4.79, and the aspect ratio of logo.svg is 5",
		},
		"too wide": {
			pngOf(t, 500, 96),
			"the aspect ratio of the image is 5.21, and the aspect ratio of logo.svg is 5",
		},
		"square, as mark": {
			pngOf(t, 96, 96),
			"the aspect ratio of the image is 1, and the aspect ratio of logo.svg is 5",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["logo.png"] = &fstest.MapFile{Data: tc.data}

			_, err := load(t, dir)

			require.ErrorContains(t, err, "logo.png: "+tc.want)
		})
	}
}

// The tolerance of 2 % permits the rounding of the width to a whole pixel.
func TestTheAspectRatioOfTheLogoPNGCanDifferBy2Percent(t *testing.T) {
	for _, width := range []int{471, 489} {
		dir := validDir(t)
		dir["logo.png"] = &fstest.MapFile{Data: pngOf(t, width, 96)}

		_, err := load(t, dir)

		require.NoError(t, err, "width %d", width)
	}
}

func TestTheLogoPNGIs100KiBOrLess(t *testing.T) {
	dir := validDir(t)
	dir["logo.png"] = &fstest.MapFile{Data: make([]byte, 100<<10+1)}

	_, err := load(t, dir)

	require.ErrorContains(t, err, "logo.png: the file is more than 100 KiB")
}

// An SVG logo that is not valid has no aspect ratio, so its error is the only error.
func TestTheLogoPNGHasNoAspectRatioErrorWhenTheSVGLogoIsNotValid(t *testing.T) {
	dir := validDir(t)
	dir["logo.svg"] = &fstest.MapFile{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`)}

	_, err := load(t, dir)

	require.ErrorContains(t, err, "logo.svg: the svg element has no viewBox")
	assert.NotContains(t, err.Error(), "logo.png")
}

func TestLoadGivesEachOpaqueColourInHexNotation(t *testing.T) {
	dir := validDir(t)
	changeTokens(t, dir, func(tokens map[string]map[string]any) {
		setValue(tokens, "color", "brand", map[string]any{"colorSpace": "oklch", "components": []float64{1, 0, 0}})
		setValue(tokens, "color", "fg", map[string]any{"colorSpace": "oklch", "components": []float64{0, 0, 0}})
		setValue(tokens, "color", "danger-fg", map[string]any{
			"colorSpace": "oklch", "components": []float64{0.62796, 0.25768, 29.234},
		})
		setValue(tokens, "color", "backdrop", map[string]any{
			"colorSpace": "oklch", "components": []float64{0, 0, 0}, "alpha": 0.4,
		})
	})

	brand, err := load(t, dir)

	require.NoError(t, err)
	assert.Equal(t, "#ffffff", brand.Colours["brand"])
	assert.Equal(t, "#000000", brand.Colours["fg"])
	assert.Equal(t, "#ff0000", brand.Colours["danger-fg"])
	assert.NotContains(t, brand.Colours, "backdrop", "a colour with transparency has no hex value")

	for role, hex := range brand.Colours {
		assert.Regexp(t, `^#[0-9a-f]{6}$`, hex, role)
	}
}

// declaredSize gives the PNG with a different size in its header and a valid checksum, but with the same image data.
func declaredSize(t *testing.T, data []byte, width, height uint32) []byte {
	t.Helper()

	const ihdrStart, sizeStart, crcStart = 12, 16, 29

	changed := slices.Clone(data)
	binary.BigEndian.PutUint32(changed[sizeStart:], width)
	binary.BigEndian.PutUint32(changed[sizeStart+4:], height)
	binary.BigEndian.PutUint32(changed[crcStart:], crc32.ChecksumIEEE(changed[ihdrStart:crcStart]))

	return changed
}

// Each file declares an image that needs more memory than the test has, so a decode before the check fails the test.
func TestTheLogoPNGIsRefusedFromTheSizeInItsHeaderBeforeItIsDecoded(t *testing.T) {
	logo := pngOf(t, 480, 96)

	tests := map[string]struct {
		data []byte
		want string
	}{
		"huge height": {
			declaredSize(t, logo, 100_000, 100_000),
			"the image is 100000 pixels high: give an image exactly 96 pixels high",
		},
		"huge width": {
			declaredSize(t, logo, 1_000_000_000, 96),
			"the image is 1000000000 pixels wide: give an image 768 pixels wide or less",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["logo.png"] = &fstest.MapFile{Data: tc.data}

			_, err := load(t, dir)

			require.ErrorContains(t, err, "logo.png: "+tc.want)
		})
	}
}

func TestTheLogoPNGCanBe768PixelsWide(t *testing.T) {
	dir := validDir(t)
	dir["logo.svg"] = &fstest.MapFile{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 1"/>`)}
	dir["logo.png"] = &fstest.MapFile{Data: pngOf(t, 768, 96)}

	_, err := load(t, dir)

	require.NoError(t, err)
}
