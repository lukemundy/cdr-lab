package dataholder_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/lukemundy/cdr-lab/charts/internal/charttest"
)

// The chart cannot import the Register, so the test decodes only the fields of the seed that the values must agree
// with.
type seedBrand struct {
	DataHolderBrandID string `json:"dataHolderBrandId"`
	BrandName         string `json:"brandName"`
	EndpointDetail    struct {
		PublicBaseURI   string `json:"publicBaseUri"`
		ResourceBaseURI string `json:"resourceBaseUri"`
		WebsiteURI      string `json:"websiteUri"`
	} `json:"endpointDetail"`
}

// exampleValues has only the values of a brand, so that a decode of an example file with other values fails.
type exampleValues struct {
	Brand struct {
		Preset string `yaml:"preset"`
		ID     string `yaml:"id"`
	} `yaml:"brand"`
	PublicBaseURL string `yaml:"publicBaseURL"`
	SecureBaseURL string `yaml:"secureBaseURL"`
	Mail          struct {
		From string `yaml:"from"`
	} `yaml:"mail"`
}

func presetBanksSeed(t *testing.T) map[string]seedBrand {
	t.Helper()

	data, err := os.ReadFile("../../services/register/seed/preset-banks.json")
	require.NoError(t, err)

	var seed struct {
		DataHolders []struct {
			DataHolderBrands []seedBrand `json:"dataHolderBrands"`
		} `json:"dataHolders"`
	}
	require.NoError(t, json.Unmarshal(data, &seed))

	brands := map[string]seedBrand{}

	for _, dh := range seed.DataHolders {
		for _, b := range dh.DataHolderBrands {
			brands[b.DataHolderBrandID] = b
		}
	}

	return brands
}

func decodeExample(t *testing.T, file string) exampleValues {
	t.Helper()

	data, err := fs.ReadFile(os.DirFS("examples"), file)
	require.NoError(t, err)

	var v exampleValues

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	require.NoError(t, dec.Decode(&v), file)

	return v
}

// Verifies: DH-1356, DH-1357, DH-1358
func TestEachExampleAgreesWithItsBrandInThePresetBanksSeed(t *testing.T) {
	seed := presetBanksSeed(t)

	files, err := fs.Glob(os.DirFS("examples"), "*.yaml")
	require.NoError(t, err)
	require.Len(t, files, len(seed), "one example for each brand of the seed")

	ids := map[string]bool{}

	for _, file := range files {
		values := decodeExample(t, file)
		preset := strings.TrimSuffix(file, ".yaml")
		env := charttest.Main(t, charttest.MustRenderValues(t, "examples/"+file)).Vars()

		brand, ok := seed[env["DATA_HOLDER_BRAND_ID"]]
		require.True(t, ok, "%s: brand.id is a brand of the seed", file)

		ids[brand.DataHolderBrandID] = true

		data, err := fs.ReadFile(os.DirFS("../../services/dataholder/brands"), preset+"/brand.json")
		require.NoError(t, err, file)

		var presetDir struct {
			BrandName string `json:"brandName"`
		}
		require.NoError(t, json.Unmarshal(data, &presetDir), file)

		website, err := url.Parse(brand.EndpointDetail.WebsiteURI)
		require.NoError(t, err, file)

		assert.Equal(t, preset, values.Brand.Preset, file)
		assert.Equal(t, preset, env["DATA_HOLDER_BRAND"], file)
		assert.Equal(t, brand.BrandName, presetDir.BrandName, file)
		assert.Equal(t, brand.EndpointDetail.PublicBaseURI, env["PUBLIC_BASE_URL"], file)
		assert.Equal(t, brand.EndpointDetail.ResourceBaseURI, env["SECURE_BASE_URL"], file)
		assert.Equal(t, "no-reply@"+website.Host, env["MAIL_FROM"], file)
	}

	assert.Len(t, ids, len(seed), "each example has its own brand")
}
