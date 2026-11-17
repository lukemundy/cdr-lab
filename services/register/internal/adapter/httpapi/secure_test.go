package httpapi_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	secureBaseURL = "https://secure.register.example"
	brandsPath    = "/cdr-register/v1/all/data-holders/brands"
	callerID      = "sp-1"
)

type fakeVerifier struct {
	tokens map[string]authserver.AccessToken
	err    error
	cert   *x509.Certificate
}

func (f *fakeVerifier) VerifyAccessToken(
	_ context.Context,
	token string,
	cert *x509.Certificate,
) (authserver.AccessToken, error) {
	f.cert = cert
	if f.err != nil {
		return authserver.AccessToken{}, f.err
	}

	t, ok := f.tokens[token]
	if !ok {
		return authserver.AccessToken{}, fmt.Errorf("%w: unknown", authserver.ErrInvalidToken)
	}

	return t, nil
}

type secureReader struct {
	caller    app.SoftwareProductClient
	callerErr error
	brands    []app.RegisteredDataHolderBrandItem
	total     int
	err       error
	queries   []app.DataHolderBrandQuery
}

func (f *secureReader) SoftwareProductClient(_ context.Context, id string) (app.SoftwareProductClient, error) {
	if id != f.caller.SoftwareProductID {
		return app.SoftwareProductClient{}, fmt.Errorf("software product %w", domain.ErrNotFound)
	}

	return f.caller, f.callerErr
}

func (f *secureReader) RegisteredDataHolderBrands(
	_ context.Context,
	q app.DataHolderBrandQuery,
) ([]app.RegisteredDataHolderBrandItem, int, error) {
	f.queries = append(f.queries, q)
	return f.brands, f.total, f.err
}

type secureLab struct {
	reader   *secureReader
	verifier *fakeVerifier
	ssa      *fakeSSAIssuer
	cert     *x509.Certificate
}

func newSecureLab() *secureLab {
	return &secureLab{
		reader: &secureReader{
			caller: app.SoftwareProductClient{
				SoftwareProductID:   callerID,
				Status:              domain.Active,
				DataRecipientStatus: domain.DataRecipientActive,
			},
			brands: []app.RegisteredDataHolderBrandItem{{
				Brand: domain.DataHolderBrand{
					DataHolderBrandID: "dhb-1",
					BrandName:         "Holder Bank",
					BrandGroup:        "Holder Group",
					Industries: []domain.Industry{
						domain.Banking,
						domain.Energy,
					},
					LogoURI:      "https://holder.example/brand.png",
					Status:       domain.Inactive,
					JWKSEndpoint: "https://holder.example/jwks",
					EndpointDetail: domain.EndpointDetail{
						Version:          "v1",
						PublicBaseURI:    "https://api.holder.example",
						ProductBaseURI:   "https://prd.holder.example",
						ResourceBaseURI:  "https://mtls.holder.example",
						InfosecBaseURI:   "https://id.holder.example",
						ExtensionBaseURI: "https://ext.holder.example",
						WebsiteURI:       "https://holder.example",
					},
				},
				LegalEntity: app.DataHolderLegalEntity{
					LegalEntityID:   "dh-1",
					LegalEntityName: "Holder Pty Ltd",
					LogoURI:         "https://holder.example/logo.png",
					Status:          domain.DataHolderActive,
				},
				LastUpdated: time.Date(2026, 10, 3, 1, 2, 3, 0, time.FixedZone("AWST", 8*3600)),
			}},
			total: 1,
		},
		verifier: &fakeVerifier{tokens: map[string]authserver.AccessToken{
			"valid":    {ClientID: callerID, Scopes: []string{"cdr-register:read"}},
			"no-scope": {ClientID: callerID, Scopes: []string{"other"}},
		}},
		ssa:  &fakeSSAIssuer{ssa: "header.claims.signature"},
		cert: &x509.Certificate{Raw: []byte("client certificate")},
	}
}

func (l *secureLab) get(t *testing.T, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	httpapi.RegisterSecure(mux, l.reader, l.verifier, l.ssa, secureBaseURL)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	req.Header = header
	req.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{l.cert}}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func authorisedV3() http.Header {
	return http.Header{"Authorization": {"Bearer valid"}, "X-V": {"3"}}
}

type cdsErrors struct {
	Errors []struct {
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

func requireSecureError(t *testing.T, rec *httptest.ResponseRecorder, status int, code, title string) string {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())

	var body cdsErrors
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, code, body.Errors[0].Code)
	assert.Equal(t, title, body.Errors[0].Title)

	return body.Errors[0].Detail
}

func TestGetDataHolderBrandsGivesEachBrandWithItsEndpointsAndLegalEntity(t *testing.T) {
	l := newSecureLab()

	rec := l.get(t, brandsPath, authorisedV3())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "3", rec.Header().Get("x-v"))
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Empty(t, rec.Header().Get("ETag"))
	assert.JSONEq(t, `{
		"data": [{
			"dataHolderBrandId": "dhb-1", "brandName": "Holder Bank", "brandGroup": "Holder Group",
			"industries": ["banking", "energy"],
			"logoUri": "https://holder.example/brand.png",
			"legalEntity": {
				"legalEntityId": "dh-1", "legalEntityName": "Holder Pty Ltd",
				"logoUri": "https://holder.example/logo.png", "status": "ACTIVE"
			},
			"status": "INACTIVE",
			"endpointDetail": {
				"version": "v1", "publicBaseUri": "https://api.holder.example",
				"productBaseUri": "https://prd.holder.example",
				"resourceBaseUri": "https://mtls.holder.example", "infosecBaseUri": "https://id.holder.example",
				"extensionBaseUri": "https://ext.holder.example", "websiteUri": "https://holder.example"
			},
			"authDetails": [{"registerUType": "SIGNED-JWT", "jwksEndpoint": "https://holder.example/jwks"}],
			"lastUpdated": "2026-10-02T17:02:03Z"
		}],
		"links": {
			"self": "https://secure.register.example/cdr-register/v1/all/data-holders/brands?page=1&page-size=25",
			"first": "https://secure.register.example/cdr-register/v1/all/data-holders/brands?page=1&page-size=25",
			"last": "https://secure.register.example/cdr-register/v1/all/data-holders/brands?page=1&page-size=25"
		},
		"meta": {"totalRecords": 1, "totalPages": 1}
	}`, rec.Body.String())
	assert.Equal(t, []app.DataHolderBrandQuery{{Page: 1, PageSize: 25}}, l.reader.queries)
	assert.Same(t, l.cert, l.verifier.cert, "the token is verified with the TLS client certificate")
}

func TestGetDataHolderBrandsOmitsTheOptionalFieldsThatABrandDoesNotHave(t *testing.T) {
	l := newSecureLab()
	l.reader.brands[0].Brand.BrandGroup = ""
	l.reader.brands[0].Brand.EndpointDetail.ProductBaseURI = ""
	l.reader.brands[0].Brand.EndpointDetail.ExtensionBaseURI = ""

	rec := l.get(t, brandsPath, authorisedV3())

	require.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), "brandGroup")
	assert.NotContains(t, rec.Body.String(), "productBaseUri")
	assert.NotContains(t, rec.Body.String(), "extensionBaseUri")
}

// Verifies: REG-0331
func TestGetDataHolderBrandsFiltersByIndustryAndUpdatedSince(t *testing.T) {
	l := newSecureLab()
	since := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	banking := domain.Banking

	rec := l.get(
		t,
		"/cdr-register/v1/banking/data-holders/brands?updated-since=2026-10-01T08:00:00%2B08:00&page-size=10",
		authorisedV3(),
	)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Len(t, l.reader.queries, 1)
	q := l.reader.queries[0]
	assert.Equal(t, &banking, q.Industry)
	require.NotNil(t, q.UpdatedSince)
	assert.True(t, since.Equal(*q.UpdatedSince))
	assert.Equal(t, 10, q.PageSize)

	var body struct {
		Links map[string]string `json:"links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(
		t,
		"https://secure.register.example/cdr-register/v1/banking/data-holders/brands"+
			"?page=1&page-size=10&updated-since=2026-10-01T08%3A00%3A00%2B08%3A00",
		body.Links["self"],
	)
}

// Verifies: REG-0330, REG-0332
func TestGetDataHolderBrandsGivesThePageOfTheRequestWithLinksAndMeta(t *testing.T) {
	link := func(page int) string {
		return fmt.Sprintf("%s%s?page=%d&page-size=2", secureBaseURL, brandsPath, page)
	}

	l := newSecureLab()
	l.reader.total = 5

	rec := l.get(t, brandsPath+"?page=2&page-size=2", authorisedV3())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Data  []json.RawMessage `json:"data"`
		Links map[string]string `json:"links"`
		Meta  map[string]int    `json:"meta"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Len(t, body.Data, 1)
	assert.Equal(
		t,
		map[string]string{"self": link(2), "first": link(1), "prev": link(1), "next": link(3), "last": link(3)},
		body.Links,
	)
	assert.Equal(t, map[string]int{"totalRecords": 5, "totalPages": 3}, body.Meta)
	assert.Equal(t, 2, l.reader.queries[0].Page)
	assert.Equal(t, 2, l.reader.queries[0].PageSize)
}

func TestGetDataHolderBrandsGivesAnEmptyArrayWhenNoBrandMatches(t *testing.T) {
	l := newSecureLab()
	l.reader.brands, l.reader.total = nil, 0

	rec := l.get(t, brandsPath, authorisedV3())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, `{
		"data": [],
		"links": {"self": "`+secureBaseURL+brandsPath+`?page=1&page-size=25"},
		"meta": {"totalRecords": 0, "totalPages": 0}
	}`, rec.Body.String())
}

// Verifies: REG-0317
func TestGetDataHolderBrandsRefusesQueryParametersThatAreNotValid(t *testing.T) {
	const (
		invalid         = "urn:au-cds:error:cds-all:Field/Invalid"
		invalidPageSize = "urn:au-cds:error:cds-all:Field/InvalidPageSize"
		invalidDate     = "urn:au-cds:error:cds-all:Field/InvalidDateTime"
		invalidPage     = "urn:au-cds:error:cds-all:Field/InvalidPage"
	)

	tests := map[string]struct {
		query               string
		status              int
		code, title, detail string
	}{
		"page zero":                   {"page=0", 400, invalid, "Invalid Field", "page"},
		"page-size above the maximum": {"page-size=1001", 400, invalidPageSize, "Invalid Page Size", "page-size"},
		"updated-since not a date":    {"updated-since=yesterday", 400, invalidDate, "Invalid Date", "updated-since"},
		"updated-since with no time": {
			"updated-since=2026-10-01",
			400,
			invalidDate,
			"Invalid Date",
			"updated-since",
		},
		"page out of range": {"page=2", 422, invalidPage, "Invalid Page", "1"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newSecureLab()

			rec := l.get(t, brandsPath+"?"+tt.query, authorisedV3())

			detail := requireSecureError(t, rec, tt.status, tt.code, tt.title)
			assert.Equal(t, tt.detail, detail)
			assert.Equal(t, "3", rec.Header().Get("x-v"))
		})
	}
}

// Verifies: REG-0319
func TestGetDataHolderBrandsAcceptsTheMaximumPageSizeAndAnEmptyUpdatedSince(t *testing.T) {
	l := newSecureLab()

	rec := l.get(t, brandsPath+"?page-size=1000&updated-since=", authorisedV3())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, 1000, l.reader.queries[0].PageSize)
	assert.Nil(t, l.reader.queries[0].UpdatedSince)
}

func TestGetDataHolderBrandsRefusesAnIndustryThatIsNotValid(t *testing.T) {
	l := newSecureLab()

	rec := l.get(t, "/cdr-register/v1/insurance/data-holders/brands", authorisedV3())

	detail := requireSecureError(
		t,
		rec,
		http.StatusBadRequest,
		"urn:au-cds:error:cds-all:Field/Invalid",
		"Invalid Field",
	)
	assert.Equal(t, "industry", detail)
	assert.Empty(t, l.reader.queries)
}

// Verifies: REG-0300, REG-0313
func TestGetDataHolderBrandsServesOnlyVersion3(t *testing.T) {
	tests := map[string]struct {
		header http.Header
		status int
	}{
		"x-v 3":                {http.Header{"X-V": {"3"}}, http.StatusOK},
		"x-v 4 with x-min-v 3": {http.Header{"X-V": {"4"}, "X-Min-V": {"3"}}, http.StatusOK},
		"x-v 2":                {http.Header{"X-V": {"2"}}, http.StatusNotAcceptable},
		"no x-v":               {http.Header{}, http.StatusBadRequest},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.header.Set("Authorization", "Bearer valid")

			rec := newSecureLab().get(t, brandsPath, tt.header)

			assert.Equal(t, tt.status, rec.Code, rec.Body.String())
		})
	}
}

// Verifies: REG-0320, REG-0321, REG-0322
func TestTheSecureAPIsRefuseARequestWithoutAValidToken(t *testing.T) {
	tests := map[string]struct {
		header    http.Header
		status    int
		challenge string
	}{
		"no authorization":      {http.Header{}, 401, `Bearer`},
		"not a bearer token":    {http.Header{"Authorization": {"Basic valid"}}, 401, `Bearer`},
		"empty token":           {http.Header{"Authorization": {"Bearer "}}, 401, `Bearer error="invalid_token"`},
		"token that is refused": {http.Header{"Authorization": {"Bearer other"}}, 401, `Bearer error="invalid_token"`},
		"token with no Register scope": {
			http.Header{"Authorization": {"Bearer no-scope"}},
			403,
			`Bearer error="insufficient_scope", scope="cdr-register:read"`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newSecureLab()

			// The request has no x-v, to show that the token check is first.
			rec := l.get(t, brandsPath, tt.header)

			assert.Equal(t, tt.status, rec.Code)
			assert.Equal(t, tt.challenge, rec.Header().Get("WWW-Authenticate"))
			assert.Empty(t, rec.Body.String())
			assert.Empty(t, l.reader.queries)
		})
	}
}

func TestTheBearerSchemeOfTheSecureAPIsIsNotCaseSensitive(t *testing.T) {
	rec := newSecureLab().get(t, brandsPath, http.Header{"Authorization": {"bearer valid"}, "X-V": {"3"}})

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestATokenVerificationFailureIsAnUnexpectedError(t *testing.T) {
	l := newSecureLab()
	l.verifier.err = errors.New("connection to 10.0.0.1 failed")

	rec := l.get(t, brandsPath, authorisedV3())

	detail := requireSecureError(
		t,
		rec,
		http.StatusInternalServerError,
		"urn:au-cds:error:cds-all:GeneralError/Unexpected",
		"Unexpected Error Encountered",
	)
	assert.NotContains(t, detail, "10.0.0.1")
	assert.Empty(t, rec.Header().Get("WWW-Authenticate"))
}

// Verifies: REG-0323
func TestTheSecureAPIsRefuseACallerThatIsNotActive(t *testing.T) {
	tests := map[string]struct {
		change func(c *app.SoftwareProductClient)
		detail string
	}{
		"inactive software product": {func(c *app.SoftwareProductClient) { c.Status = domain.Inactive }, "INACTIVE"},
		"removed software product":  {func(c *app.SoftwareProductClient) { c.Status = domain.Removed }, "REMOVED"},
		"suspended data recipient": {func(c *app.SoftwareProductClient) {
			c.Status, c.DataRecipientStatus = domain.Inactive, domain.DataRecipientSuspended
		}, "INACTIVE"},
		"data recipient that is not active with an active software product": {func(c *app.SoftwareProductClient) {
			c.DataRecipientStatus = domain.DataRecipientRevoked
		}, "REVOKED"},
		"software product that does not exist": {
			func(c *app.SoftwareProductClient) { c.SoftwareProductID = "sp-other" },
			"REMOVED",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newSecureLab()
			tt.change(&l.reader.caller)

			// The request has no x-v, to show that the status check is before the version check.
			rec := l.get(t, brandsPath, http.Header{"Authorization": {"Bearer valid"}})

			detail := requireSecureError(
				t,
				rec,
				http.StatusForbidden,
				"urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive",
				"ADR Status Is Not Active",
			)
			assert.Equal(t, tt.detail, detail)
			assert.Empty(t, l.reader.queries)
		})
	}
}

func TestAReadFailureOnTheSecureAPIsIsAnUnexpectedError(t *testing.T) {
	failure := errors.New("connection to 10.0.0.1 failed")

	tests := map[string]func(r *secureReader){
		"caller status": func(r *secureReader) { r.callerErr = failure },
		"brands":        func(r *secureReader) { r.err = failure },
	}
	for name, fail := range tests {
		t.Run(name, func(t *testing.T) {
			l := newSecureLab()
			fail(l.reader)

			rec := l.get(t, brandsPath, authorisedV3())

			detail := requireSecureError(
				t,
				rec,
				http.StatusInternalServerError,
				"urn:au-cds:error:cds-all:GeneralError/Unexpected",
				"Unexpected Error Encountered",
			)
			assert.NotContains(t, detail, "10.0.0.1")
		})
	}
}

func TestAnUnknownPathOnTheSecureAPIsIsNotFound(t *testing.T) {
	rec := newSecureLab().get(t, "/unknown", http.Header{})

	requireSecureError(t, rec, http.StatusNotFound, "urn:au-cds:error:cds-all:Resource/NotFound", "Resource Not Found")
}

// Verifies: REG-1495
func TestAnotherMethodOnASecureAPIPathIsNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	lab := newSecureLab()
	httpapi.RegisterSecure(mux, lab.reader, lab.verifier, lab.ssa, secureBaseURL)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, brandsPath, nil))

	requireSecureError(
		t,
		rec,
		http.StatusMethodNotAllowed,
		"urn:au-cds:error:cds-all:GeneralError/Expected",
		"Expected Error Encountered",
	)
	assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
}
