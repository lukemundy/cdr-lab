package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const productsPath = "/cds-au/v1/banking/products"

type fakeReader struct {
	products []domain.Product
	total    int
	err      error
	queries  []app.ProductQuery
}

func (f *fakeReader) Products(_ context.Context, q app.ProductQuery) ([]domain.Product, int, error) {
	f.queries = append(f.queries, q)
	return f.products, f.total, f.err
}

func (f *fakeReader) Product(_ context.Context, productID string) (domain.Product, error) {
	if f.err != nil {
		return domain.Product{}, f.err
	}

	for _, p := range f.products {
		if p.ProductID == productID {
			return p, nil
		}
	}

	return domain.Product{}, app.ErrProductNotFound
}

func homeLoan() domain.Product {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	return domain.Product{
		ProductID:       "home-loan",
		LastUpdated:     time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC),
		ProductCategory: domain.CategoryResidentialMortgages,
		Name:            "Home Loan",
		Description:     "A loan for a home.",
		Brand:           "Mock Bank",
		EffectiveFrom:   &from,
		EffectiveTo:     &to,
		Features:        []domain.Feature{{FeatureType: domain.FeatureRedraw}},
		Fees: []domain.Fee{{
			Name: "Monthly fee", FeeType: domain.FeeTypePeriodic, FeeMethodUType: domain.FeeFixedAmount,
			FixedAmount: &domain.FeeAmount{Amount: "8.00"}, AdditionalValue: "P1M",
		}},
		DepositRates: []domain.DepositRate{{
			DepositRateType: domain.DepositRateVariable, Rate: "0.01", ApplicationType: domain.ApplicationPeriodic,
			ApplicationFrequency: "P1M",
		}},
		LendingRates: []domain.LendingRate{{
			LendingRateType: domain.LendingRateVariable, Rate: "0.0614", ComparisonRate: "0.0618",
			ApplicationType: domain.ApplicationPeriodic, ApplicationFrequency: "P1M",
			RepaymentType: domain.RepaymentPrincipalAndInterest, LoanPurpose: domain.PurposeOwnerOccupied,
		}},
	}
}

func oneHomeLoan() *fakeReader {
	return &fakeReader{products: []domain.Product{homeLoan()}, total: 1}
}

func requireError(t *testing.T, rec *httptest.ResponseRecorder, status int, code, title, detail string) {
	t.Helper()

	require.Equal(t, status, rec.Code, rec.Body.String())

	want, err := json.Marshal(map[string]any{
		"errors": []map[string]string{{"code": code, "title": title, "detail": detail}},
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(want), rec.Body.String())
}

func TestGetProductsGivesTheListFieldsOfEachProductWithLinksAndMeta(t *testing.T) {
	rec := sendTo(t, oneHomeLoan(), http.MethodGet, productsPath, xv("5"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "5", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {"products": [{
			"productId": "home-loan",
			"effectiveFrom": "2026-01-01T00:00:00Z",
			"effectiveTo": "2027-01-01T00:00:00Z",
			"lastUpdated": "2026-09-09T00:00:00Z",
			"productCategory": "RESIDENTIAL_MORTGAGES",
			"name": "Home Loan",
			"description": "A loan for a home.",
			"brand": "Mock Bank",
			"isTailored": false
		}]},
		"links": {
			"self": "https://bank.example/cds-au/v1/banking/products?page=1&page-size=25",
			"first": "https://bank.example/cds-au/v1/banking/products?page=1&page-size=25",
			"last": "https://bank.example/cds-au/v1/banking/products?page=1&page-size=25"
		},
		"meta": {"totalRecords": 1, "totalPages": 1}
	}`, rec.Body.String())
}

func TestGetProductsGivesAnEmptyArrayWhenNoProductMatches(t *testing.T) {
	rec := sendTo(t, &fakeReader{}, http.MethodGet, productsPath, xv("5"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, `{
		"data": {"products": []},
		"links": {"self": "https://bank.example/cds-au/v1/banking/products?page=1&page-size=25"},
		"meta": {"totalRecords": 0, "totalPages": 0}
	}`, rec.Body.String())
}

// Verifies: DH-0443
func TestGetProductsAsksForTheCurrentProductsAtTheTimeOfTheRequestWhenTheRequestHasNoParameters(t *testing.T) {
	tests := map[string]string{
		"no parameters":    "",
		"empty parameters": "?effective=&updated-since=&brand=&product-category=&page=&page-size=",
	}
	for name, query := range tests {
		t.Run(name, func(t *testing.T) {
			reader := oneHomeLoan()

			rec := sendTo(t, reader, http.MethodGet, productsPath+query, xv("5"))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(
				t,
				[]app.ProductQuery{{Effective: app.EffectiveCurrent, At: requestTime, Page: 1, PageSize: 25}},
				reader.queries,
			)
		})
	}
}

func TestGetProductsAsksForTheFiltersAndThePageOfTheRequestAndGivesThemInEachLink(t *testing.T) {
	reader := oneHomeLoan()
	reader.total = 5

	rec := sendTo(
		t,
		reader,
		http.MethodGet,
		productsPath+"?effective=ALL&updated-since=2026-10-01T08:00:00%2B08:00&brand=Mock%20Bank"+
			"&product-category=RESIDENTIAL_MORTGAGES&page=2&page-size=2",
		xv("5"),
	)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Len(t, reader.queries, 1)

	query := reader.queries[0]
	require.NotNil(t, query.UpdatedSince)
	assert.True(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).Equal(*query.UpdatedSince))

	query.UpdatedSince = nil
	assert.Equal(t, app.ProductQuery{
		Effective: app.EffectiveAll, At: requestTime, Brand: "Mock Bank",
		Category: domain.CategoryResidentialMortgages, Page: 2, PageSize: 2,
	}, query)

	var body struct {
		Links map[string]string `json:"links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	link := func(page string) string {
		return "https://bank.example/cds-au/v1/banking/products?brand=Mock+Bank&effective=ALL&page=" + page +
			"&page-size=2&product-category=RESIDENTIAL_MORTGAGES&updated-since=2026-10-01T08%3A00%3A00%2B08%3A00"
	}
	assert.Equal(t, map[string]string{
		"self": link("2"), "first": link("1"), "prev": link("1"), "next": link("3"), "last": link("3"),
	}, body.Links)
}

func TestGetProductsAcceptsEachValueOfTheEffectiveFilter(t *testing.T) {
	for _, effective := range []app.Effective{app.EffectiveCurrent, app.EffectiveFuture, app.EffectiveAll} {
		t.Run(string(effective), func(t *testing.T) {
			reader := oneHomeLoan()

			rec := sendTo(t, reader, http.MethodGet, productsPath+"?effective="+string(effective), xv("5"))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, effective, reader.queries[0].Effective)
		})
	}
}

// Verifies: DH-0441, DH-0442
func TestGetProductsRefusesAParameterThatIsNotValid(t *testing.T) {
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
		"effective in lower case":          {"effective=all", 400, invalid, "Invalid Field", "effective"},
		"product-category not in the enumeration": {
			"product-category=SAVINGS", 400, invalid, "Invalid Field", "product-category",
		},
		"updated-since not a date": {"updated-since=yesterday", 400, invalidDate, "Invalid Date", "updated-since"},
		"updated-since with no time": {
			"updated-since=2026-10-01", 400, invalidDate, "Invalid Date", "updated-since",
		},
		"page zero":                   {"page=0", 400, invalid, "Invalid Field", "page"},
		"page-size above the maximum": {"page-size=1001", 400, invalidPageSize, "Invalid Page Size", "page-size"},
		"page out of range":           {"page=2", 422, invalidPage, "Invalid Page", "1"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := sendTo(t, oneHomeLoan(), http.MethodGet, productsPath+"?"+tt.query, xv("5"))

			requireError(t, rec, tt.status, tt.code, tt.title, tt.detail)
			assert.Equal(t, "5", rec.Header().Get("x-v"))
		})
	}
}

func TestGetProductDetailGivesEachFieldOfTheProduct(t *testing.T) {
	rec := sendTo(t, oneHomeLoan(), http.MethodGet, productsPath+"/home-loan", xv("7"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "7", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {
			"productId": "home-loan",
			"effectiveFrom": "2026-01-01T00:00:00Z",
			"effectiveTo": "2027-01-01T00:00:00Z",
			"lastUpdated": "2026-09-09T00:00:00Z",
			"productCategory": "RESIDENTIAL_MORTGAGES",
			"name": "Home Loan",
			"description": "A loan for a home.",
			"brand": "Mock Bank",
			"isTailored": false,
			"features": [{"featureType": "REDRAW"}],
			"fees": [{
				"name": "Monthly fee", "feeType": "PERIODIC", "feeMethodUType": "fixedAmount",
				"fixedAmount": {"amount": "8.00"}, "additionalValue": "P1M"
			}],
			"depositRates": [{
				"depositRateType": "VARIABLE", "rate": "0.01", "applicationType": "PERIODIC",
				"applicationFrequency": "P1M"
			}],
			"lendingRates": [{
				"lendingRateType": "VARIABLE", "rate": "0.0614", "comparisonRate": "0.0618",
				"applicationType": "PERIODIC", "applicationFrequency": "P1M",
				"repaymentType": "PRINCIPAL_AND_INTEREST", "loanPurpose": "OWNER_OCCUPIED"
			}]
		},
		"links": {"self": "https://bank.example/cds-au/v1/banking/products/home-loan"},
		"meta": {}
	}`, rec.Body.String())
}

func TestGetProductDetailGivesAnEscapedLinkForAProductIDThatAURLMustEscape(t *testing.T) {
	loan := homeLoan()
	loan.ProductID = "home loan/1"
	reader := &fakeReader{products: []domain.Product{loan}}

	rec := sendTo(t, reader, http.MethodGet, productsPath+"/home%20loan%2F1", xv("7"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Links map[string]string `json:"links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "https://bank.example/cds-au/v1/banking/products/home%20loan%2F1", body.Links["self"])
}

// Verifies: DH-0444
func TestGetProductDetailRefusesAProductIDThatTheBrandDoesNotHave(t *testing.T) {
	rec := sendTo(t, oneHomeLoan(), http.MethodGet, productsPath+"/car-loan", xv("7"))

	requireError(
		t, rec, http.StatusNotFound, "urn:au-cds:error:cds-all:Resource/Invalid", "Invalid Resource", "car-loan",
	)
	assert.Equal(t, "7", rec.Header().Get("x-v"))
}

// Verifies: DH-0431
func TestTheProductAPIsServeOnlyTheirCurrentVersion(t *testing.T) {
	paths := map[string]string{productsPath: "5", productsPath + "/home-loan": "7"}
	for path, version := range paths {
		t.Run(path+" serves the version", func(t *testing.T) {
			rec := sendTo(t, oneHomeLoan(), http.MethodGet, path, http.Header{"X-V": {"9"}, "X-Min-V": {"1"}})

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, version, rec.Header().Get("x-v"))
		})

		t.Run(path+" refuses a deprecated version", func(t *testing.T) {
			rec := sendTo(t, oneHomeLoan(), http.MethodGet, path, xv("4"))

			assert.Equal(t, http.StatusNotAcceptable, rec.Code)
			assert.Contains(t, rec.Body.String(), "urn:au-cds:error:cds-all:Header/UnsupportedVersion")
		})

		t.Run(path+" checks the version before the parameters", func(t *testing.T) {
			rec := sendTo(t, oneHomeLoan(), http.MethodGet, path+"?page=0", http.Header{})

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "urn:au-cds:error:cds-all:Header/Missing")
		})
	}
}

func TestAReadFailureOnTheProductAPIsIsAnUnexpectedErrorThatDoesNotShowTheCause(t *testing.T) {
	for _, path := range []string{productsPath, productsPath + "/home-loan"} {
		t.Run(path, func(t *testing.T) {
			reader := &fakeReader{err: errors.New("connection refused")}

			rec := sendTo(t, reader, http.MethodGet, path, http.Header{"X-V": {"9"}, "X-Min-V": {"1"}})

			requireError(
				t, rec, http.StatusInternalServerError,
				"urn:au-cds:error:cds-all:GeneralError/Unexpected", "Unexpected Error Encountered",
				"The data holder could not complete the request.",
			)
		})
	}
}

func TestTheProductAPIsAllowAnyOrigin(t *testing.T) {
	for _, path := range []string{productsPath, productsPath + "/home-loan"} {
		t.Run(path, func(t *testing.T) {
			rec := sendTo(t, oneHomeLoan(), http.MethodGet, path, http.Header{"Origin": {"https://app.example"}})

			assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(
				t, "x-v, Retry-After, x-fapi-interaction-id", rec.Header().Get("Access-Control-Expose-Headers"),
			)
		})
	}
}
