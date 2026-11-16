package banksite_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/openapi"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const host = "localhost:8185"

var (
	ui  = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Bank</title>")}}
	now = time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
)

// products is a ProductReader that gives its products for each query, and keeps the last query.
type products struct {
	list  []domain.Product
	query app.ProductQuery
	err   error
}

func (p *products) Products(_ context.Context, query app.ProductQuery) ([]domain.Product, int, error) {
	p.query = query
	found := []domain.Product{}

	for _, product := range p.list {
		if query.ProductID == "" || product.ProductID == query.ProductID {
			found = append(found, product)
		}
	}

	return found, len(found), p.err
}

func (p *products) Product(context.Context, string) (domain.Product, error) {
	return domain.Product{}, errors.New("the site does not read a product without the current rule")
}

func saver() domain.Product {
	return domain.Product{
		ProductID: "saver", LastUpdated: now.Add(-time.Hour), ProductCategory: domain.CategoryTransAndSavingsAccounts,
		Name: "Saver", Description: "A savings account.", Brand: "Wattle Bank",
		Features: []domain.Feature{{FeatureType: domain.FeatureNppPayid}},
		DepositRates: []domain.DepositRate{{
			DepositRateType: domain.DepositRateVariable, Rate: "0.0125", ApplicationType: domain.ApplicationPeriodic,
			ApplicationFrequency: "P1M",
		}},
	}
}

// The document loads once, because a test of the session limit sends 10,000 requests.
var spec = sync.OnceValues(func() (*openapi.Spec, error) {
	data, err := os.ReadFile("openapi.yaml")
	if err != nil {
		return nil, err
	}

	return openapi.Load(data, "")
})

func loadSpec(t *testing.T) *openapi.Spec {
	t.Helper()

	loaded, err := spec()
	require.NoError(t, err)

	return loaded
}

type response struct {
	status int
	header http.Header
	body   string
}

// get validates the response against the OpenAPI document.
func get(t *testing.T, ports banksite.Ports, target string) response {
	t.Helper()

	if ports.Now == nil {
		ports.Now = func() time.Time { return now }
	}

	if ports.Products == nil {
		ports.Products = &products{}
	}

	mux := http.NewServeMux()
	require.NoError(t, banksite.Register(mux, ui, []string{host}, ports))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+host+target, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	if strings.HasPrefix(target, "/api/") {
		resp.Request = req
		require.NoError(t, loadSpec(t).ValidateResponse(resp))
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{status: resp.StatusCode, header: resp.Header, body: string(body)}
}

// Verifies: DH-1715
func TestTheDocumentClosesEachObjectSchema(t *testing.T) {
	assert.NoError(t, loadSpec(t).CheckClosed())
}

// Verifies: DH-1724
func TestTheBrandGivesTheBrandNameTheTextOfTheFooterTheLayoutAndThePhotos(t *testing.T) {
	ports := banksite.Ports{
		BrandName: "Wattle Bank", BrandLayout: "meadow", BrandPhotos: []string{"category-overdrafts.webp", "hero.webp"},
	}

	resp := get(t, ports, "/api/brand")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{
		"brandName": "Wattle Bank",
		"layout": "meadow",
		"photos": ["category-overdrafts.webp", "hero.webp"],
		"demonstrationNotice": "Wattle Bank is not a real bank and this is just a demonstration. `+
		`CDR Lab has no affiliation with any real bank nor the ACCC, DSB, or any real CDR participant."
	}`, resp.body)
}

// The site then shows no broken image for a brand with no photos.
func TestABrandWithNoPhotosGivesAnEmptyList(t *testing.T) {
	resp := get(t, banksite.Ports{BrandName: "Wattle Bank", BrandLayout: "ledger"}, "/api/brand")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Contains(t, resp.body, `"photos":[]`)
}

// Verifies: DH-1726, DH-1730
func TestTheProductsAreEachCurrentProductAtTheTimeOfTheRequestOnOnePage(t *testing.T) {
	reader := &products{list: []domain.Product{saver()}}

	resp := get(t, banksite.Ports{Products: reader}, "/api/products")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, app.ProductQuery{Effective: app.EffectiveCurrent, At: now, Page: 1}, reader.query)
	assert.JSONEq(t, `{"products": [{
		"productId": "saver", "lastUpdated": "2026-11-01T23:00:00Z", "productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
		"name": "Saver", "description": "A savings account.", "brand": "Wattle Bank", "isTailored": false
	}]}`, resp.body)
}

func TestNoCurrentProductGivesAnEmptyList(t *testing.T) {
	resp := get(t, banksite.Ports{}, "/api/products")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"products": []}`, resp.body)
}

// Verifies: DH-1727
func TestTheDetailGivesTheCurrentProductWithTheID(t *testing.T) {
	reader := &products{list: []domain.Product{saver()}}

	resp := get(t, banksite.Ports{Products: reader}, "/api/products/saver")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Equal(
		t,
		app.ProductQuery{Effective: app.EffectiveCurrent, At: now, ProductID: "saver", Page: 1},
		reader.query,
	)

	want, err := json.Marshal(saver())
	require.NoError(t, err)
	assert.JSONEq(t, string(want), resp.body)
}

// Verifies: DH-1727
func TestTheDetailOfAProductThatIsNotCurrentIsNotFound(t *testing.T) {
	resp := get(t, banksite.Ports{Products: &products{list: []domain.Product{saver()}}}, "/api/products/ended-saver")

	require.Equal(t, http.StatusNotFound, resp.status)
	assert.NotContains(t, resp.body, "ended-saver")
}

func TestAFailureOfTheStoreIsAServerError(t *testing.T) {
	for _, target := range []string{"/api/products", "/api/products/saver"} {
		t.Run(target, func(t *testing.T) {
			resp := get(t, banksite.Ports{Products: &products{err: errors.New("no database")}}, target)

			assert.Equal(t, http.StatusInternalServerError, resp.status)
			assert.NotContains(t, resp.body, "no database")
		})
	}
}

// The CSP of the site permits styles and images from the same origin only.
//
// Verifies: DH-2071, DH-2074
func TestTheSiteServesTheBrandFilesWithTheResponseHeadersOfTheSite(t *testing.T) {
	ports := banksite.Ports{BrandFiles: func(routes *httproute.Routes) {
		routes.Handle("GET /brand/styles.css", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, ":root:root {}\n")
		}))
	}}

	resp := get(t, ports, "/brand/styles.css")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, ":root:root {}\n", resp.body)
	assert.True(t, strings.HasPrefix(resp.header.Get("Content-Security-Policy"), "default-src 'self';"))
}

// Verifies: DH-2069
func TestTheSiteHasNoHSTSAndThePagesOfTheProductsSetNoCookie(t *testing.T) {
	for _, target := range []string{"/", "/products", "/api/brand", "/api/products"} {
		t.Run(target, func(t *testing.T) {
			resp := get(t, banksite.Ports{BrandName: "Wattle Bank", BrandLayout: "meadow"}, target)

			assert.Equal(t, http.StatusOK, resp.status)
			assert.Empty(t, resp.header.Get("Strict-Transport-Security"))
			assert.Empty(t, resp.header.Values("Set-Cookie"))
		})
	}
}

// The UI gives a label for each value of the document, so each list must be the list of the data holder and the CDS.
// Verifies: DH-1731
func TestEachEnumerationOfTheDocumentHasTheValuesOfTheDataHolderAndTheCDS(t *testing.T) {
	document, err := openapi3.NewLoader().LoadFromFile("openapi.yaml")
	require.NoError(t, err)

	cds := cdsEnumerations(t)
	enums := domain.ProductEnumerations()
	maps.Copy(enums, domain.AccountEnumerations())
	require.Len(t, cds, len(enums))

	for name, want := range enums {
		t.Run(name, func(t *testing.T) {
			schema := document.Components.Schemas[name]
			require.NotNil(t, schema, "the document has no schema %s", name)

			var got []string
			for _, value := range schema.Value.Enum {
				got = append(got, value.(string))
			}

			assert.ElementsMatch(t, want, got)
			assert.ElementsMatch(t, cds[name], got)
		})
	}
}

// cdsEnumerations reads each enumeration of a product, an account and a transaction from the CDS OpenAPI document,
// by the name of the domain type.
func cdsEnumerations(t *testing.T) map[string][]string {
	t.Helper()

	data, err := os.ReadFile("../../../../../conformance/dataholder/spec/cds_banking.json")
	require.NoError(t, err)

	var banking struct {
		Components struct {
			Schemas map[string]struct {
				Enum       []string `json:"enum"`
				Properties map[string]struct {
					Enum []string `json:"enum"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	require.NoError(t, json.Unmarshal(data, &banking))

	schemas := banking.Components.Schemas
	property := func(schema, name string) []string {
		values := schemas[schema].Properties[name].Enum
		require.NotEmpty(t, values, "%s.%s", schema, name)

		return values
	}

	return map[string][]string{
		"ProductCategory":   schemas["BankingProductCategoryV2"].Enum,
		"FeatureType":       property("BankingProductFeatureV4", "featureType"),
		"FeeType":           property("BankingProductFeeV2", "feeType"),
		"FeeMethodUType":    property("BankingProductFeeV2", "feeMethodUType"),
		"FeeRateType":       property("BankingFeeRate", "rateType"),
		"DepositRateType":   property("BankingProductDepositRateV2", "depositRateType"),
		"LendingRateType":   property("BankingProductLendingRateV3", "lendingRateType"),
		"ApplicationType":   property("BankingProductLendingRateV3", "applicationType"),
		"RepaymentType":     property("BankingProductLendingRateV3", "repaymentType"),
		"LoanPurpose":       property("BankingProductLendingRateV3", "loanPurpose"),
		"OpenStatus":        property("BankingAccountV3", "openStatus"),
		"TransactionType":   property("BankingTransactionV2", "type"),
		"TransactionStatus": property("BankingTransactionV2", "status"),
	}
}
