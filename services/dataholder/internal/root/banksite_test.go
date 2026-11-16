package root

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var testBankSite = consolehttp.Config{Enabled: true, AllowedHosts: []string{"example.com"}}

func TestWithoutTheSettingTheDataHolderOpensNoBankSiteListener(t *testing.T) {
	listeners, err := bankSiteListeners(consolehttp.Config{}, tls.Certificate{}, ui, banksite.Ports{})

	require.NoError(t, err)
	assert.Empty(t, listeners)
}

func TestWithTheSettingTheBankSitePortGivesTheHTMLPageWithTheTLSOfThePublicListener(t *testing.T) {
	cert := tls.Certificate{Certificate: [][]byte{[]byte("public certificate")}}

	listeners, err := bankSiteListeners(testBankSite, cert, ui, banksite.Ports{})
	require.NoError(t, err)
	require.Len(t, listeners, 1)
	assert.Equal(t, 8085, listeners[0].Port)
	require.NotNil(t, listeners[0].TLSConfig)
	assert.Equal(t, []tls.Certificate{cert}, listeners[0].TLSConfig.Certificates)
	assert.Equal(t, tls.NoClientCert, listeners[0].TLSConfig.ClientAuth)
	assert.Equal(t, uint16(tls.VersionTLS12), listeners[0].TLSConfig.MinVersion)

	rec := httptest.NewRecorder()
	listeners[0].Handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/products", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
}

func TestTheBankSitePortRefusesAHostThatIsNotAllowed(t *testing.T) {
	listeners, err := bankSiteListeners(testBankSite, tls.Certificate{}, ui, banksite.Ports{})
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://attacker.example:8185/", nil)
	rec := httptest.NewRecorder()
	listeners[0].Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTheAPIProcessOpensNoBankSiteListenerWithoutTheSetting(t *testing.T) {
	lab := newOperatorLab(t)

	assert.NotContains(t, lab.handlers, bankSitePort)
}

func TestTheAPIsLoadTheBankSiteSettings(t *testing.T) {
	setOnly(t, apiMode)
	t.Setenv("BANK_SITE_ENABLED", "true")
	t.Setenv("BANK_SITE_ALLOWED_HOSTS", "localhost:8185")
	t.Setenv("DATA_HOLDER_BRAND_ID", "d7385786-d8b1-4d4b-bb80-b21616df1a4d")

	cfg, err := loadConfig(apiMode)

	require.NoError(t, err)
	assert.Equal(t, consolehttp.Config{Enabled: true, AllowedHosts: []string{"localhost:8185"}}, cfg.api.bankSite)
	assert.Equal(t, "d7385786-d8b1-4d4b-bb80-b21616df1a4d", cfg.api.brandID)

	for _, notACookieName := range []string{"", "brand id", "brand;id", "brand=id", "bränd"} {
		t.Setenv("DATA_HOLDER_BRAND_ID", notACookieName)

		_, err = loadConfig(apiMode)
		require.ErrorContains(t, err, "DATA_HOLDER_BRAND_ID", "value %q", notACookieName)
	}

	t.Setenv("DATA_HOLDER_BRAND_ID", "d7385786-d8b1-4d4b-bb80-b21616df1a4d")

	t.Setenv("BANK_SITE_ALLOWED_HOSTS", "")

	_, err = loadConfig(apiMode)
	assert.ErrorContains(t, err, "BANK_SITE_ALLOWED_HOSTS")
}

func TestTheWorkersNeedNoBankSiteSetting(t *testing.T) {
	setOnly(t, workerMode)
	t.Setenv("BANK_SITE_ENABLED", "true")
	t.Setenv("BANK_SITE_ALLOWED_HOSTS", "")

	_, err := loadConfig(workerMode)

	require.NoError(t, err)
}

// The site and Get Products read the same store at the same time.
//
// Verifies: DH-1728, DH-1729
func TestTheBankSiteGivesTheProductsOfGetProductsAndGetProductDetailForEachPresetBrand(t *testing.T) {
	now := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }

	for name, brand := range presetBrands(t) {
		t.Run(name, func(t *testing.T) {
			pool := databasetest.New(t)
			require.NoError(t, migrate(t.Context(), pool))

			products := productstore.New(pool)
			all := catalogue(t, brand)
			require.NoError(t, products.ImportProducts(t.Context(), all))
			requireFutureAndPastProducts(t, all, now)

			cds := http.NewServeMux()
			httpapi.Register(cds, "https://bank.test", products, clock)

			site := http.NewServeMux()
			require.NoError(t, banksite.Register(site, ui, []string{"bank.test"}, banksite.Ports{
				BrandName: brand.Name, Products: products, Now: clock,
			}))

			var getProducts struct {
				Data struct {
					Products []json.RawMessage `json:"products"`
				} `json:"data"`
			}
			decode(t, cds, "/cds-au/v1/banking/products?effective=CURRENT&page-size=1000", "5", &getProducts)

			var siteList struct {
				Products []json.RawMessage `json:"products"`
			}
			decode(t, site, "/api/products", "", &siteList)

			assert.ElementsMatch(t, canonical(t, getProducts.Data.Products), canonical(t, siteList.Products))

			current := map[string]bool{}

			for _, p := range siteList.Products {
				var id struct {
					ProductID string `json:"productId"`
				}
				require.NoError(t, json.Unmarshal(p, &id))
				current[id.ProductID] = true
			}

			for _, p := range all {
				target := "/api/products/" + p.ProductID
				if !current[p.ProductID] {
					assert.Equal(t, http.StatusNotFound, serveProduct(t, site, target, "").Code, target)
					continue
				}

				var detail struct {
					Data json.RawMessage `json:"data"`
				}
				decode(t, cds, "/cds-au/v1/banking/products/"+p.ProductID, "7", &detail)

				var siteDetail json.RawMessage
				decode(t, site, target, "", &siteDetail)

				assert.Equal(
					t,
					canonical(t, []json.RawMessage{detail.Data}),
					canonical(t, []json.RawMessage{siteDetail}),
				)
			}
		})
	}
}

func serveProduct(t *testing.T, mux *http.ServeMux, target, version string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://bank.test"+target, nil)
	if version != "" {
		req.Header.Set("x-v", version)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func decode(t *testing.T, mux *http.ServeMux, target, version string, into any) {
	t.Helper()

	rec := serveProduct(t, mux, target, version)
	require.Equal(t, http.StatusOK, rec.Code, "%s: %s", target, rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), into))
}

// canonical gives each JSON value with sorted keys and no white space, so that equal values are equal strings.
func canonical(t *testing.T, values []json.RawMessage) []string {
	t.Helper()

	list := make([]string, 0, len(values))
	for _, v := range values {
		var decoded any
		require.NoError(t, json.Unmarshal(v, &decoded))

		encoded, err := json.Marshal(decoded)
		require.NoError(t, err)

		list = append(list, string(encoded))
	}

	return list
}

func TestTheBankSitePutsNoQueryStringAndNoPathValueInTheSpansMetricsOrLogs(t *testing.T) {
	spans, metrics, logs := recordTelemetry(t)

	listeners, err := bankSiteListeners(testBankSite, tls.Certificate{}, ui, banksite.Ports{
		BrandName: "Coastal Bank", Products: catalogueReader{}, Now: time.Now,
	})
	require.NoError(t, err)

	for _, target := range []string{"/api/products/kookaburra-saver?q=quokka", "/api/products?wombat=numbat"} {
		rec := httptest.NewRecorder()
		listeners[0].Handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
		require.NotEqual(t, http.StatusForbidden, rec.Code, rec.Body.String())
	}

	var text strings.Builder
	for _, span := range spans.Ended() {
		text.WriteString(spanText(span))
	}

	text.WriteString(metricText(t, metrics))
	text.WriteString(strings.ToLower(logs.String()))

	require.NotEmpty(t, spans.Ended())

	for _, value := range []string{"kookaburra", "quokka", "wombat", "numbat"} {
		assert.NotContains(t, text.String(), value)
	}
}

// Each preset has a future and a past product, so that the test checks both reasons for a 404 of the site.
func requireFutureAndPastProducts(t *testing.T, products []domain.Product, now time.Time) {
	t.Helper()

	var future, past int

	for _, p := range products {
		if p.EffectiveFrom != nil && p.EffectiveFrom.After(now) {
			future++
		}

		if p.EffectiveTo != nil && !p.EffectiveTo.After(now) {
			past++
		}
	}

	require.Positive(t, future, "the catalogue has no future product at %s", now)
	require.Positive(t, past, "the catalogue has no past product at %s", now)
}
