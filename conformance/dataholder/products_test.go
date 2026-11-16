//go:build conformance

package dataholder

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const productsPath = "/cds-au/v1/banking/products"

type listedProduct struct {
	ID, Category               string
	LastUpdated                time.Time
	EffectiveFrom, EffectiveTo *time.Time
}

func (p listedProduct) current(now time.Time) bool {
	return (p.EffectiveFrom == nil || !p.EffectiveFrom.After(now)) &&
		(p.EffectiveTo == nil || now.Before(*p.EffectiveTo))
}

func (p listedProduct) future(now time.Time) bool {
	return p.EffectiveFrom != nil && p.EffectiveFrom.After(now)
}

// catalogue gives the products through the API, so that the tests do not depend on the catalogue of a brand.
func catalogue(t *testing.T) []listedProduct {
	t.Helper()

	page := products(t, productsPath+"?effective=ALL&page-size=1000")
	require.Equal(t, 1, page.Meta.TotalPages)

	all := make([]listedProduct, 0, len(page.Data.Products))
	for _, product := range page.Data.Products {
		all = append(all, listedProduct{
			ID:            product["productId"].(string),
			Category:      product["productCategory"].(string),
			LastUpdated:   dateTime(t, product["lastUpdated"]),
			EffectiveFrom: optionalDateTime(t, product["effectiveFrom"]),
			EffectiveTo:   optionalDateTime(t, product["effectiveTo"]),
		})
	}

	return all
}

func dateTime(t *testing.T, value any) time.Time {
	t.Helper()

	text, ok := value.(string)
	require.True(t, ok, "%v is not a string", value)

	parsed, err := time.Parse(time.RFC3339, text)
	require.NoError(t, err)

	return parsed
}

func optionalDateTime(t *testing.T, value any) *time.Time {
	t.Helper()

	if value == nil {
		return nil
	}

	parsed := dateTime(t, value)

	return &parsed
}

func idsOf(products []listedProduct) []string {
	return idsWhere(products, func(listedProduct) bool { return true })
}

func idsWhere(products []listedProduct, keep func(listedProduct) bool) []string {
	result := []string{}

	for _, p := range products {
		if keep(p) {
			result = append(result, p.ID)
		}
	}

	return result
}

func currentIDs(t *testing.T) []string {
	t.Helper()

	now := time.Now()
	current := idsWhere(catalogue(t), func(p listedProduct) bool { return p.current(now) })
	require.NotEmpty(t, current)

	return current
}

func productBaseURI(t *testing.T) string {
	t.Helper()

	return registeredEndpoints(t).ProductBaseURI
}

// banking validates each response against the pinned CDS banking OpenAPI specification.
func banking(t *testing.T, target string, header http.Header) response {
	t.Helper()

	if !strings.HasPrefix(target, "https://") {
		target = productBaseURI(t) + target
	}

	resp := get(t, target, header)

	spec, err := LoadBankingSpec()
	require.NoError(t, err)

	resp.Body = io.NopCloser(bytes.NewReader(resp.body))
	require.NoError(t, spec.ValidateResponse(resp.Response))

	return resp
}

type productPage struct {
	Data struct {
		Products []map[string]any `json:"products"`
	} `json:"data"`
	Links map[string]string `json:"links"`
	Meta  struct {
		TotalRecords int `json:"totalRecords"`
		TotalPages   int `json:"totalPages"`
	} `json:"meta"`
}

func (p productPage) ids() []string {
	ids := make([]string, 0, len(p.Data.Products))
	for _, product := range p.Data.Products {
		ids = append(ids, product["productId"].(string))
	}

	return ids
}

func products(t *testing.T, target string) productPage {
	t.Helper()

	resp := banking(t, target, xv("5"))
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	assert.Equal(t, "5", resp.Header.Get("x-v"))

	var page productPage
	require.NoError(t, json.Unmarshal(resp.body, &page))

	return page
}

func requireCDSError(t *testing.T, resp response, status int, code, title string) string {
	t.Helper()

	require.Equal(t, status, resp.StatusCode, string(resp.body))

	var body struct {
		Errors []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, code, body.Errors[0].Code)
	assert.Equal(t, title, body.Errors[0].Title)

	return body.Errors[0].Detail
}

func TestGetProductsGivesEachProductWithTheNewestLastUpdatedFirstAndThenInTheOrderOfItsID(t *testing.T) {
	all := catalogue(t)

	assert.True(t, slices.IsSortedFunc(all, func(a, b listedProduct) int {
		if c := b.LastUpdated.Compare(a.LastUpdated); c != 0 {
			return c
		}

		return strings.Compare(a.ID, b.ID)
	}), "the order of the products: %v", idsOf(all))
}

func TestGetProductsGivesTheCurrentProductsOfTheBrand(t *testing.T) {
	brand := bankName(t)
	current := currentIDs(t)
	page := products(t, productsPath)

	assert.Equal(t, current, page.ids())
	assert.Equal(t, len(current), page.Meta.TotalRecords)
	assert.Equal(t, 1, page.Meta.TotalPages)
	assert.Equal(t, productBaseURI(t)+productsPath+"?page=1&page-size=25", page.Links["self"])

	listFields := []string{
		"productId", "effectiveFrom", "effectiveTo", "lastUpdated", "productCategory", "name", "description",
		"brand", "isTailored",
	}

	for _, product := range page.Data.Products {
		for field := range product {
			assert.Contains(t, listFields, field, "%s has a field that is not in the list", product["productId"])
		}

		assert.Equal(t, brand, product["brand"])
	}
}

func TestGetProductsEffectiveGivesTheCurrentTheFutureOrAllProducts(t *testing.T) {
	now := time.Now()
	all := catalogue(t)
	future := idsWhere(all, func(p listedProduct) bool { return p.future(now) })
	past := idsWhere(all, func(p listedProduct) bool { return p.EffectiveTo != nil && !now.Before(*p.EffectiveTo) })

	// The lab brand has a future product and a product that is no longer offered (data holder DH-0223).
	require.NotEmpty(t, future)
	require.NotEmpty(t, past)

	tests := map[string][]string{
		"CURRENT": idsWhere(all, func(p listedProduct) bool { return p.current(now) }),
		"FUTURE":  future,
		"ALL":     idsOf(all),
	}
	for effective, want := range tests {
		t.Run(effective, func(t *testing.T) {
			page := products(t, productsPath+"?effective="+effective)

			assert.Equal(t, want, page.ids())
			assert.Equal(t, len(want), page.Meta.TotalRecords)
			assert.Contains(t, page.Links["self"], "effective="+effective)
		})
	}
}

// middleLastUpdated gives the middle value of the lastUpdated values of the current products, so that an updated-since
// filter with it has products on each side.
func middleLastUpdated(t *testing.T, current []listedProduct) time.Time {
	t.Helper()

	var times []time.Time

	for _, p := range current {
		if !slices.ContainsFunc(times, p.LastUpdated.Equal) {
			times = append(times, p.LastUpdated)
		}
	}

	require.GreaterOrEqual(t, len(times), 2, "the current products need two lastUpdated values")

	return times[len(times)/2]
}

func TestGetProductsGivesOnlyTheProductsThatMatchEachFilterOfTheRequest(t *testing.T) {
	brand := bankName(t)
	now := time.Now()
	all := catalogue(t)

	var current []listedProduct

	for _, p := range all {
		if p.current(now) {
			current = append(current, p)
		}
	}

	since := middleLastUpdated(t, current)
	category := current[0].Category
	currentList := idsOf(current)

	tests := map[string]struct {
		query url.Values
		want  []string
	}{
		"updated-since gives a later lastUpdated only": {
			url.Values{"updated-since": {since.Format(time.RFC3339)}},
			idsWhere(current, func(p listedProduct) bool { return p.LastUpdated.After(since) }),
		},
		"updated-since with a time zone offset": {
			url.Values{"updated-since": {since.Add(-time.Second).In(time.FixedZone("", 8*60*60)).Format(time.RFC3339)}},
			idsWhere(current, func(p listedProduct) bool { return !p.LastUpdated.Before(since) }),
		},
		"brand":                     {url.Values{"brand": {brand}}, currentList},
		"brand in a different case": {url.Values{"brand": {swapCase(brand)}}, []string{}},
		"product-category": {
			url.Values{"product-category": {category}},
			idsWhere(current, func(p listedProduct) bool { return p.Category == category }),
		},
		"each filter": {
			url.Values{
				"effective":        {"ALL"},
				"updated-since":    {since.Format(time.RFC3339)},
				"brand":            {brand},
				"product-category": {category},
			},
			idsWhere(all, func(p listedProduct) bool { return p.Category == category && p.LastUpdated.After(since) }),
		},
		"empty values are the same as no parameters": {
			url.Values{"effective": {""}, "updated-since": {""}, "brand": {""}, "product-category": {""}},
			currentList,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			page := products(t, productsPath+"?"+tt.query.Encode())

			assert.Equal(t, tt.want, page.ids())
			assert.Equal(t, len(tt.want), page.Meta.TotalRecords)

			if len(tt.want) == 0 {
				assert.Equal(t, 0, page.Meta.TotalPages, "totalPages is 0 when totalRecords is 0")
				assert.Equal(t, []string{"self"}, slices.Collect(maps.Keys(page.Links)))
			}
		})
	}
}

func TestGetProductsPagesFollowTheCDSPaginationRules(t *testing.T) {
	all := idsOf(catalogue(t))
	require.GreaterOrEqual(t, len(all), 5, "the pages need a first, a middle and a last page")

	lastPage := (len(all) + 1) / 2
	base := productBaseURI(t)

	var listed []string

	target := productsPath + "?effective=ALL&page-size=2"
	for number := 1; target != ""; number++ {
		require.LessOrEqual(t, number, lastPage, "the next links stop at the last page")

		page := products(t, target)

		assert.Equal(t, len(all), page.Meta.TotalRecords)
		assert.Equal(t, lastPage, page.Meta.TotalPages)

		for name, link := range page.Links {
			assert.True(t, strings.HasPrefix(link, base+productsPath+"?"), "%s is a product URL: %s", name, link)
			assert.Contains(t, link, "page-size=2", name)
			assert.Contains(t, link, "effective=ALL", "%s has the filters of the request", name)
		}

		assert.Contains(t, page.Links["self"], "page="+strconv.Itoa(number))
		assert.Equal(t, number > 1, page.Links["prev"] != "", "prev is present only after the first page")
		assert.Equal(t, number < lastPage, page.Links["next"] != "", "next is present only before the last page")
		assert.NotEmpty(t, page.Links["first"])
		assert.NotEmpty(t, page.Links["last"])

		listed = append(listed, page.ids()...)
		target = page.Links["next"]
	}

	assert.Equal(t, all, listed)
}

func TestGetProductsAcceptsTheMaximumPageSize(t *testing.T) {
	page := products(t, productsPath+"?page-size=1000")

	assert.Equal(t, currentIDs(t), page.ids())
}

func TestGetProductsRefusesAParameterThatIsNotValid(t *testing.T) {
	currentCount := len(currentIDs(t))
	require.Less(t, currentCount, 1000, "page 1000 is out of range")

	const (
		invalid         = "urn:au-cds:error:cds-all:Field/Invalid"
		invalidDate     = "urn:au-cds:error:cds-all:Field/InvalidDateTime"
		invalidPageSize = "urn:au-cds:error:cds-all:Field/InvalidPageSize"
		invalidPage     = "urn:au-cds:error:cds-all:Field/InvalidPage"
	)

	tests := map[string]struct {
		query               string
		status              int
		code, title, detail string
	}{
		"effective not in the enumeration": {"effective=PAST", 400, invalid, "Invalid Field", "effective"},
		"product-category not in the enumeration": {
			"product-category=SAVINGS", 400, invalid, "Invalid Field", "product-category",
		},
		"page zero":                {"page=0", 400, invalid, "Invalid Field", "page"},
		"page not a number":        {"page=one", 400, invalid, "Invalid Field", "page"},
		"updated-since not a date": {"updated-since=yesterday", 400, invalidDate, "Invalid Date", "updated-since"},
		"page-size zero":           {"page-size=0", 400, invalidPageSize, "Invalid Page Size", "page-size"},
		"page-size above 1000":     {"page-size=1001", 400, invalidPageSize, "Invalid Page Size", "page-size"},
		"page out of range": {
			"page=1000&page-size=1",
			422,
			invalidPage,
			"Invalid Page",
			strconv.Itoa(currentCount),
		},
		"page 2 of an empty list": {"brand=None&page=2", 422, invalidPage, "Invalid Page", "0"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := banking(t, productsPath+"?"+tt.query, xv("5"))

			detail := requireCDSError(t, resp, tt.status, tt.code, tt.title)
			assert.Equal(t, tt.detail, detail)
			assert.Equal(t, "5", resp.Header.Get("x-v"), "an error after the version negotiation has x-v")
		})
	}
}

type productDetail struct {
	Data  map[string]json.RawMessage `json:"data"`
	Links map[string]string          `json:"links"`
}

func detail(t *testing.T, productID string) productDetail {
	t.Helper()

	resp := banking(t, productsPath+"/"+productID, xv("7"))
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	assert.Equal(t, "7", resp.Header.Get("x-v"))

	var body productDetail
	require.NoError(t, json.Unmarshal(resp.body, &body))

	return body
}

func TestGetProductDetailGivesEachProductOfTheListAlsoWhenItIsNotCurrent(t *testing.T) {
	list := products(t, productsPath+"?effective=ALL&page-size=1000")
	require.NotEmpty(t, list.ids())

	for _, listed := range list.Data.Products {
		id := listed["productId"].(string)

		t.Run(id, func(t *testing.T) {
			body := detail(t, id)

			assert.Equal(t, productBaseURI(t)+productsPath+"/"+id, body.Links["self"])

			for field, want := range listed {
				wantJSON, err := json.Marshal(want)
				require.NoError(t, err)
				assert.JSONEq(t, string(wantJSON), string(body.Data[field]), "%s is the same as in the list", field)
			}
		})
	}
}

func TestGetProductDetailGivesTheFeaturesFeesAndRatesOfTheProducts(t *testing.T) {
	given := map[string]bool{}

	for _, p := range catalogue(t) {
		body := detail(t, p.ID)

		for _, field := range []string{"features", "fees", "depositRates", "lendingRates"} {
			if raw, ok := body.Data[field]; ok {
				var components []map[string]any
				require.NoError(t, json.Unmarshal(raw, &components), "%s of %s", field, p.ID)
				given[field] = given[field] || len(components) > 0
			}
		}
	}

	// The catalogue of a retail bank has a deposit product and a lending product (data holder DH-0222).
	for _, field := range []string{"features", "fees", "depositRates", "lendingRates"} {
		assert.True(t, given[field], "no product has %s", field)
	}
}

func TestGetProductDetailRefusesAProductIDThatTheBrandDoesNotHave(t *testing.T) {
	resp := banking(t, productsPath+"/no-such-product", xv("7"))

	detail := requireCDSError(
		t, resp, http.StatusNotFound, "urn:au-cds:error:cds-all:Resource/Invalid", "Invalid Resource",
	)
	assert.Equal(t, "no-such-product", detail)
	assert.Equal(t, "7", resp.Header.Get("x-v"))
}

type productAPI struct {
	path       string
	version    string
	deprecated string
}

func productAPIs(t *testing.T) map[string]productAPI {
	t.Helper()

	return map[string]productAPI{
		"Get Products":       {path: productsPath, version: "5", deprecated: "4"},
		"Get Product Detail": {path: productsPath + "/" + currentIDs(t)[0], version: "7", deprecated: "6"},
	}
}

func TestTheProductAPIsServeOnlyTheirCurrentVersion(t *testing.T) {
	for name, api := range productAPIs(t) {
		served := map[string]http.Header{
			"x-v is the version":                xv(api.version),
			"x-v higher, x-min-v lower":         {"X-V": {"99"}, "X-Min-V": {"1"}},
			"x-min-v higher than x-v, ignored":  {"X-V": {api.version}, "X-Min-V": {"99"}},
			"x-v higher, x-min-v the version":   {"X-V": {"99"}, "X-Min-V": {api.version}},
			"x-v with leading zeros":            xv("00" + api.version),
			"x-v higher, x-min-v is deprecated": {"X-V": {"99"}, "X-Min-V": {api.deprecated}},
		}
		for scenario, header := range served {
			t.Run(name+" "+scenario, func(t *testing.T) {
				resp := banking(t, api.path, header)

				assert.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
				assert.Equal(t, api.version, resp.Header.Get("x-v"))
			})
		}

		refused := map[string]struct {
			header      http.Header
			status      int
			code, title string
		}{
			"no x-v": {
				http.Header{},
				http.StatusBadRequest,
				"urn:au-cds:error:cds-all:Header/Missing", "Missing Required Header",
			},
			"x-v is not a positive integer": {
				xv("0"), http.StatusBadRequest,
				"urn:au-cds:error:cds-all:Header/InvalidVersion", "Invalid Version",
			},
			"x-min-v is not a positive integer": {
				http.Header{"X-V": {api.version}, "X-Min-V": {"a"}},
				http.StatusBadRequest,
				"urn:au-cds:error:cds-all:Header/InvalidVersion", "Invalid Version",
			},
			"a deprecated version": {
				xv(api.deprecated), http.StatusNotAcceptable,
				"urn:au-cds:error:cds-all:Header/UnsupportedVersion", "Unsupported Version",
			},
			"x-min-v higher than the version": {
				http.Header{"X-V": {"99"}, "X-Min-V": {"98"}},
				http.StatusNotAcceptable,
				"urn:au-cds:error:cds-all:Header/UnsupportedVersion", "Unsupported Version",
			},
		}
		for scenario, tt := range refused {
			t.Run(name+" "+scenario, func(t *testing.T) {
				resp := banking(t, api.path, tt.header)

				detail := requireCDSError(t, resp, tt.status, tt.code, tt.title)
				assert.NotEmpty(t, detail)
				assert.Empty(t, resp.Header.Get("x-v"))
			})
		}
	}
}

func TestTheVersionOfAProductRequestIsCheckedBeforeItsParameters(t *testing.T) {
	resp := banking(t, productsPath+"?page=0", http.Header{})

	requireCDSError(
		t, resp, http.StatusBadRequest, "urn:au-cds:error:cds-all:Header/Missing", "Missing Required Header",
	)
}

func TestTheProductAPIsHaveAnInteractionIDAllowAnyOriginAndExposeTheCDSHeaders(t *testing.T) {
	const interactionID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	for name, api := range productAPIs(t) {
		tests := map[string]http.Header{
			"success": {"X-V": {api.version}},
			"error":   {},
		}
		for scenario, header := range tests {
			t.Run(name+" "+scenario, func(t *testing.T) {
				header.Set("Origin", "https://app.example")
				header.Set("x-fapi-interaction-id", interactionID)

				resp := banking(t, api.path, header)

				assert.Equal(t, interactionID, resp.Header.Get("x-fapi-interaction-id"))
				assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
				assert.Equal(
					t,
					"x-v, Retry-After, x-fapi-interaction-id",
					resp.Header.Get("Access-Control-Expose-Headers"),
				)
			})
		}
	}
}

func TestACORSPreflightRequestToAProductAPIIsAllowed(t *testing.T) {
	for name, api := range productAPIs(t) {
		t.Run(name, func(t *testing.T) {
			resp := send(t, http.MethodOptions, productBaseURI(t)+api.path, http.Header{
				"Origin":                         {"https://app.example"},
				"Access-Control-Request-Method":  {"GET"},
				"Access-Control-Request-Headers": {"x-v,x-min-v"},
			})

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET", resp.Header.Get("Access-Control-Allow-Methods"))
			assert.Equal(t, "x-v,x-min-v", resp.Header.Get("Access-Control-Allow-Headers"))
		})
	}
}

func TestTheLastUpdatedOfEachProductIsAnRFC3339DateTime(t *testing.T) {
	for _, product := range products(t, productsPath+"?effective=ALL").Data.Products {
		_, err := time.Parse(time.RFC3339, product["lastUpdated"].(string))
		assert.NoError(t, err, product["productId"])
	}
}

// swapCase changes the case of each letter, so the result differs from the name in each letter.
func swapCase(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsUpper(r) {
			return unicode.ToLower(r)
		}

		return unicode.ToUpper(r)
	}, name)
}
