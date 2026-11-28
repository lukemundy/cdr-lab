package branddir_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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

func noProductCheck([][]byte) error { return nil }

func load(t *testing.T, dir fs.FS) (branddir.Brand, error) {
	t.Helper()

	return branddir.Load(dir, noProductCheck)
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

// Verifies: DH-0321
func TestLoadRefusesAFileThatTheContractDoesNotNameAndAMissingFile(t *testing.T) {
	dir := validDir(t)
	dir["favicon.ico"] = &fstest.MapFile{Data: []byte("x")}
	dir["extra/notes.txt"] = &fstest.MapFile{Data: []byte("x")}
	delete(dir, "mark.svg")
	delete(dir, "products.json")

	_, err := load(t, dir)

	require.Error(t, err)

	message := err.Error()
	assert.Contains(t, message, "favicon.ico: the brand directory contract has no file of this name")
	assert.Contains(t, message, "extra: the brand directory contract has no file of this name")
	assert.Contains(t, message, "mark.svg: the file is missing")
	assert.Contains(t, message, "products.json: the file is missing")
}

// Verifies: DH-0331
func TestTheBrandFileHasOnlyTheFieldsOfTheDataHolder(t *testing.T) {
	for name, tc := range map[string]struct{ file, want string }{
		"unknown field": {
			`{"brandName": "Coastal Bank", "bsb": "376100", "layout": "meadow", "logo": "x"}`,
			`brand.json: json: unknown field "logo"`,
		},
		"brand name": {
			`{"brandName": " ", "bsb": "376100", "layout": "meadow"}`,
			"brand.json: brandName: give 1 to 200 characters",
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := validDir(t)
			dir["brand.json"] = &fstest.MapFile{Data: []byte(tc.file)}

			_, err := load(t, dir)

			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestLoadGivesTheSharedPartsOfTheBrand(t *testing.T) {
	dir := validDir(t)

	brand, err := load(t, dir)

	require.NoError(t, err)
	assert.Equal(t, dir["logo.png"].Data, brand.LogoPNG)
	assert.NotEmpty(t, brand.Colours["brand"])
}

// The bsb has the rule of a bsb of an account in the operator interface.
// Verifies: DH-0334
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

func TestLoadGivesTheProductsToTheProductCheck(t *testing.T) {
	failure := errors.New(`product 1 ("everyday"): name is mandatory`)

	var gotProducts [][]byte

	_, err := branddir.Load(validDir(t), func(products [][]byte) error {
		gotProducts = products
		return failure
	})

	assert.Len(t, gotProducts, 1)
	require.ErrorContains(t, err, `products.json: product 1 ("everyday"): name is mandatory`)
}

func TestLoadRefusesAProductsFileThatIsNotAProductSet(t *testing.T) {
	dir := validDir(t)
	dir["products.json"] = &fstest.MapFile{Data: []byte(`{"product": []}`)}

	_, err := load(t, dir)

	require.ErrorContains(t, err, "products.json: ")
}
