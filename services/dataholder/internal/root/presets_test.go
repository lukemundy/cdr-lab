package root

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/openapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/brands"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/branddir"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// These tests check the rules of design/brand.md and design/product-reference-data.md for each preset brand. The
// contrast of each preset is a test of the design system (web/packages/tokens).

func presetBrands(t *testing.T) map[string]branddir.Brand {
	t.Helper()

	presets := map[string]branddir.Brand{}

	for _, name := range brands.Names() {
		dir, err := brands.Preset(name)
		require.NoError(t, err)

		brand, err := branddir.Load(dir, checkProducts)
		require.NoError(t, err, "preset %s", name)

		presets[name] = brand
	}

	require.NotEmpty(t, presets)

	return presets
}

func catalogue(t *testing.T, brand branddir.Brand) []domain.Product {
	t.Helper()

	products, err := app.CheckProducts(brand.Name, brand.Products)
	require.NoError(t, err)

	return products
}

// blockedWords make a name look like a test fixture, not a bank. A preset has no host name or email address, so the
// test top-level domain does not occur in these names.
var blockedWords = regexp.MustCompile(`(?i)\b(mock|example|sample|demo|test|fake|dummy)\b`)

// Verifies: DH-0361
func TestNoNameOfAPresetBrandHasABlockedWord(t *testing.T) {
	for name, brand := range presetBrands(t) {
		t.Run(name, func(t *testing.T) {
			assert.NotRegexp(t, blockedWords, brand.Name)

			for _, p := range catalogue(t, brand) {
				assert.NotRegexp(t, blockedWords, p.Name, "product %s", p.ProductID)
			}
		})
	}
}

// DESIGN.md, Decisions: the BSB of an account that the consumer opens. Each preset has a BSB of its own.
func TestEachPresetBrandHasTheBSBOfTheDecision(t *testing.T) {
	bsbs := map[string]string{}
	for name, brand := range presetBrands(t) {
		bsbs[name] = brand.BSB
	}

	assert.Equal(t, map[string]string{
		"wattle-bank": "372100", "redgum-bank": "373100", "harbour-bank": "374100", "kestrel-bank": "375100",
	}, bsbs)
}

// Verifies: DH-0222, DH-0223
func TestTheCatalogueOfEachPresetBrandHasFiveCategoriesAFutureProductAndAPastProduct(t *testing.T) {
	now := time.Now()

	for name, brand := range presetBrands(t) {
		t.Run(name, func(t *testing.T) {
			var categories []domain.ProductCategory

			var future, past bool

			for _, p := range catalogue(t, brand) {
				if !slices.Contains(categories, p.ProductCategory) {
					categories = append(categories, p.ProductCategory)
				}

				future = future || (p.EffectiveFrom != nil && p.EffectiveFrom.After(now))
				past = past || (p.EffectiveTo != nil && p.EffectiveTo.Before(now))
			}

			assert.GreaterOrEqual(t, len(categories), 5, "the product categories: %v", categories)
			assert.True(t, future, "no product has an effectiveFrom in the future")
			assert.True(t, past, "no product has an effectiveTo in the past")
		})
	}
}

// DESIGN.md, Online bank site: Presentation. Each preset has a layout of its own, a hero photo and a photo for each
// category that has a current product.
func TestEachPresetBrandHasALayoutOfItsOwnAndAPhotoForEachCurrentCategory(t *testing.T) {
	now := time.Now()
	layouts := map[string]string{}

	for name, brand := range presetBrands(t) {
		layouts[name] = brand.Layout

		t.Run(name, func(t *testing.T) {
			want := []string{"hero.webp"}

			for _, p := range catalogue(t, brand) {
				current := (p.EffectiveFrom == nil || !p.EffectiveFrom.After(now)) &&
					(p.EffectiveTo == nil || p.EffectiveTo.After(now))
				photo := "category-" + strings.ReplaceAll(
					strings.ToLower(string(p.ProductCategory)),
					"_",
					"-",
				) + ".webp"

				if current && !slices.Contains(want, photo) {
					want = append(want, photo)
				}
			}

			assert.Subset(t, brand.Photos(), want)
		})
	}

	assert.Equal(t, map[string]string{
		"wattle-bank": "meadow", "redgum-bank": "ledger", "harbour-bank": "horizon", "kestrel-bank": "updraft",
	}, layouts)
}

type catalogueReader struct{ products []domain.Product }

func (c catalogueReader) Products(context.Context, app.ProductQuery) ([]domain.Product, int, error) {
	return c.products, len(c.products), nil
}

func (c catalogueReader) Product(_ context.Context, id string) (domain.Product, error) {
	for _, p := range c.products {
		if p.ProductID == id {
			return p, nil
		}
	}

	return domain.Product{}, app.ErrProductNotFound
}

// Verifies: DH-0225
func TestTheProductResponsesOfEachPresetBrandAreValidAgainstTheCDSSpecification(t *testing.T) {
	specJSON, err := os.ReadFile("../../../../conformance/dataholder/spec/cds_banking.json")
	require.NoError(t, err)

	spec, err := openapi.Load(specJSON, "/cds-au/v1")
	require.NoError(t, err)

	for name, brand := range presetBrands(t) {
		t.Run(name, func(t *testing.T) {
			products := catalogue(t, brand)

			mux := http.NewServeMux()
			httpapi.Register(mux, "https://bank.test", catalogueReader{products}, time.Now)

			for _, effective := range []string{"CURRENT", "FUTURE", "ALL"} {
				validate(t, spec, mux, "/cds-au/v1/banking/products?page-size=1000&effective="+effective, "5")
			}

			for _, p := range products {
				validate(t, spec, mux, "/cds-au/v1/banking/products/"+p.ProductID, "7")
			}
		})
	}
}

func validate(t *testing.T, spec *openapi.Spec, mux *http.ServeMux, target, version string) {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://bank.test"+target, nil)
	req.Header.Set("x-v", version)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	resp := rec.Result()
	resp.Request = req

	require.Equal(t, http.StatusOK, resp.StatusCode, "%s: %s", target, rec.Body.String())
	assert.NoError(t, spec.ValidateResponse(resp), target)
}
