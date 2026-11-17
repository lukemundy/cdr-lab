package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	baseURL               = "https://register.example"
	recipientStatusesPath = "/cdr-register/v1/all/data-recipients/status"
	productStatusesPath   = "/cdr-register/v1/all/data-recipients/brands/software-products/status"
	holderStatusesPath    = "/cdr-register/v1/all/data-holders/status"
	brandsSummaryPath     = "/cdr-register/v1/all/data-holders/brands/summary"
	dataRecipientsPath    = "/cdr-register/v1/all/data-recipients"
	errorContentType      = "application/json"
)

func noKeys(context.Context) ([]authserver.SigningKey, error) { return nil, nil }

type fakeReader struct {
	brands         []app.DataHolderBrandItem
	brandIndustry  *domain.Industry
	holders        []app.DataHolderStatusItem
	holderIndustry *domain.Industry
	recipients     []app.DataRecipientStatusItem
	products       []app.SoftwareProductStatusItem
	dataRecipients []app.DataRecipientItem
	err            error
}

func (f *fakeReader) DataRecipients(context.Context) ([]app.DataRecipientItem, error) {
	return f.dataRecipients, f.err
}

func (f *fakeReader) DataHolderBrands(_ context.Context, industry *domain.Industry) ([]app.DataHolderBrandItem, error) {
	f.brandIndustry = industry
	return f.brands, f.err
}

func (f *fakeReader) DataHolderStatuses(
	_ context.Context,
	industry *domain.Industry,
) ([]app.DataHolderStatusItem, error) {
	f.holderIndustry = industry
	return f.holders, f.err
}

func (f *fakeReader) DataRecipientStatuses(context.Context) ([]app.DataRecipientStatusItem, error) {
	return f.recipients, f.err
}

func (f *fakeReader) SoftwareProductStatuses(context.Context) ([]app.SoftwareProductStatusItem, error) {
	return f.products, f.err
}

func newReader() *fakeReader {
	return &fakeReader{
		brands: []app.DataHolderBrandItem{
			{
				Brand: domain.DataHolderBrand{
					DataHolderBrandID: "dhb-1",
					InterimID:         "dhi-1",
					BrandName:         "Holder Bank",
					BrandGroup:        "Holder Group",
					Industries: []domain.Industry{
						domain.Banking,
						domain.Energy,
					},
					LogoURI:      "https://holder.example/logo.png",
					Status:       domain.Active,
					JWKSEndpoint: "https://holder.example/jwks",
					EndpointDetail: domain.EndpointDetail{
						Version:          "1",
						PublicBaseURI:    "https://api.holder.example",
						ProductBaseURI:   "https://prd.holder.example",
						ResourceBaseURI:  "https://mtls.holder.example",
						InfosecBaseURI:   "https://id.holder.example",
						ExtensionBaseURI: "https://ext.holder.example",
						WebsiteURI:       "https://holder.example",
					},
				},
				LastUpdated: time.Date(2026, 9, 30, 10, 15, 30, 0, time.FixedZone("AWST", 8*60*60)),
			},
			{
				Brand: domain.DataHolderBrand{
					InterimID: "dhi-2", BrandName: "Holder Energy", Industries: []domain.Industry{domain.Energy},
					LogoURI: "https://holder.example/energy.png", Status: domain.Inactive,
					EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://prd.energy.example"},
				},
				LastUpdated: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		holders: []app.DataHolderStatusItem{
			{LegalEntityID: "dh-1", Status: domain.DataHolderActive},
			{LegalEntityID: "dh-2", Status: domain.DataHolderRemoved},
		},
		recipients: []app.DataRecipientStatusItem{
			{LegalEntityID: "dr-1", Status: domain.DataRecipientActive},
			{LegalEntityID: "dr-2", Status: domain.DataRecipientSuspended},
		},
		products: []app.SoftwareProductStatusItem{
			{SoftwareProductID: "sp-1", Status: domain.Active},
			{SoftwareProductID: "sp-2", Status: domain.Removed},
		},
		dataRecipients: []app.DataRecipientItem{
			{
				DataRecipient: domain.DataRecipient{
					LegalEntityID: "dr-1", LegalEntityName: "Recipient Pty Ltd", AccreditationNumber: "ADR000001",
					AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
					Status: domain.DataRecipientActive,
					Brands: []domain.DataRecipientBrand{
						{
							DataRecipientBrandID: "drb-1",
							BrandName:            "Recipient App",
							LogoURI:              "https://recipient.example/brand.png",
							Status:               domain.Active,
							SoftwareProducts: []domain.SoftwareProduct{
								{
									SoftwareProductID:          "sp-1",
									SoftwareProductName:        "Recipient App Web",
									SoftwareProductDescription: "A web app",
									LogoURI:                    "https://recipient.example/product.png",
									Status:                     domain.Active,
									ClientURI:                  "https://recipient.example",
									SectorIdentifierURI:        "https://recipient.example/sector",
									TOSURI:                     "https://recipient.example/tos",
									PolicyURI:                  "https://recipient.example/policy",
									RedirectURIs:               []string{"https://recipient.example/callback"},
									JWKSURI:                    "https://recipient.example/jwks",
									RevocationURI:              "https://recipient.example/revoke",
									RecipientBaseURI:           "https://cdr.recipient.example",
									Scopes:                     []string{"openid", "profile"},
								},
							},
						},
						{
							DataRecipientBrandID: "drb-2",
							BrandName:            "Old App",
							LogoURI:              "https://recipient.example/old.png",
							Status:               domain.Removed,
						},
					},
				},
				LastUpdated: time.Date(2026, 9, 30, 10, 15, 30, 123456000, time.FixedZone("AWST", 8*60*60)),
			},
			{
				DataRecipient: domain.DataRecipient{
					LegalEntityID: "dr-2", LegalEntityName: "Former Recipient Ltd", AccreditationNumber: "ADR000002",
					AccreditationLevel: domain.Sponsored, LogoURI: "https://former.example/logo.png",
					Status: domain.DataRecipientRevoked,
				},
				LastUpdated: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func serve(t *testing.T, reader *fakeReader, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, reader, http.MethodGet, path, header)
}

func send(t *testing.T, reader *fakeReader, method, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	httpapi.Register(mux, reader, noKeys, baseURL)

	req := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	httpapi.AllowAnyOrigin(cds.InteractionID(mux)).ServeHTTP(rec, req)

	return rec
}

func v3() http.Header {
	return http.Header{"X-V": {"3"}}
}

func v4() http.Header {
	return http.Header{"X-V": {"4"}}
}

var lists = map[string]struct {
	path    string
	version string
}{
	"data recipient statuses":    {recipientStatusesPath, "3"},
	"software product statuses":  {productStatusesPath, "3"},
	"data recipients":            {dataRecipientsPath, "4"},
	"data holder statuses":       {holderStatusesPath, "2"},
	"data holder brands summary": {brandsSummaryPath, "2"},
}

// Verifies: REG-0306
func TestDataRecipientsGiveEachDataRecipientWithOnlyTheFieldsOfTheSchema(t *testing.T) {
	rec := serve(t, newReader(), dataRecipientsPath, v4())

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"data":[{
			"legalEntityId":"dr-1","legalEntityName":"Recipient Pty Ltd","accreditationNumber":"ADR000001",
			"accreditationLevel":"UNRESTRICTED","logoUri":"https://recipient.example/logo.png","status":"ACTIVE",
			"lastUpdated":"2026-09-30T02:15:30Z",
			"dataRecipientBrands":[{
				"dataRecipientBrandId":"drb-1","brandName":"Recipient App",
				"logoUri":"https://recipient.example/brand.png",
				"status":"ACTIVE",
				"softwareProducts":[{
					"softwareProductId":"sp-1","softwareProductName":"Recipient App Web",
					"softwareProductDescription":"A web app","logoUri":"https://recipient.example/product.png",
					"status":"ACTIVE"
				}]
			},{
				"dataRecipientBrandId":"drb-2","brandName":"Old App","logoUri":"https://recipient.example/old.png",
				"status":"REMOVED","softwareProducts":[]
			}]
		},{
			"legalEntityId":"dr-2","legalEntityName":"Former Recipient Ltd","accreditationNumber":"ADR000002",
			"accreditationLevel":"SPONSORED","logoUri":"https://former.example/logo.png","status":"REVOKED",
			"lastUpdated":"2026-09-01T00:00:00Z","dataRecipientBrands":[]
		}],
		"links":{"self":"https://register.example/cdr-register/v1/all/data-recipients"},"meta":{}}`, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "4", rec.Header().Get("X-V"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Regexp(t, `^"[0-9a-f]{32}"$`, rec.Header().Get("ETag"))
}

// Verifies: REG-0305, REG-0313
func TestDataRecipientsServeOnlyVersion4(t *testing.T) {
	tests := map[string]struct {
		header http.Header
		status int
	}{
		"x-v 4":                      {http.Header{"X-V": {"4"}}, http.StatusOK},
		"x-v higher, x-min-v 4":      {http.Header{"X-V": {"5"}, "X-Min-V": {"4"}}, http.StatusOK},
		"x-v 3":                      {http.Header{"X-V": {"3"}}, http.StatusNotAcceptable},
		"x-v higher, x-min-v higher": {http.Header{"X-V": {"6"}, "X-Min-V": {"5"}}, http.StatusNotAcceptable},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, newReader(), dataRecipientsPath, tt.header)

			require.Equal(t, tt.status, rec.Code, rec.Body.String())

			if tt.status == http.StatusOK {
				assert.Equal(t, "4", rec.Header().Get("X-V"))
			}
		})
	}
}

func TestTheDataRecipientsETagChangesWithAStatus(t *testing.T) {
	reader := newReader()
	first := serve(t, reader, dataRecipientsPath, v4()).Header().Get("ETag")
	reader.dataRecipients[0].DataRecipient.Brands[0].SoftwareProducts[0].Status = domain.Inactive
	changed := serve(t, reader, dataRecipientsPath, v4()).Header().Get("ETag")

	assert.NotEqual(t, first, changed)
}

type cdsError struct {
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func requireError(t *testing.T, rec *httptest.ResponseRecorder, status int, code, title string) cdsError {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())
	assert.Equal(t, errorContentType, rec.Header().Get("Content-Type"))

	var body struct {
		Errors []cdsError `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, code, body.Errors[0].Code)
	assert.Equal(t, title, body.Errors[0].Title)
	assert.NotEmpty(t, body.Errors[0].Detail)

	return body.Errors[0]
}

// Verifies: REG-0338
func TestStatusListsGiveEachParticipant(t *testing.T) {
	tests := map[string]struct {
		path string
		want string
	}{
		"data recipients": {
			path: recipientStatusesPath,
			want: `{"data":[{"legalEntityId":"dr-1","status":"ACTIVE"},` +
				`{"legalEntityId":"dr-2","status":"SUSPENDED"}],` +
				`"links":{"self":"https://register.example/cdr-register/v1/all/data-recipients/status"},"meta":{}}`,
		},
		"software products": {
			path: productStatusesPath,
			want: `{"data":[{"softwareProductId":"sp-1","status":"ACTIVE"},` +
				`{"softwareProductId":"sp-2","status":"REMOVED"}],` +
				`"links":{"self":"https://register.example` +
				`/cdr-register/v1/all/data-recipients/brands/software-products/status"},"meta":{}}`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, newReader(), tt.path, v3())

			require.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, tt.want, rec.Body.String())
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.Equal(t, "3", rec.Header().Get("X-V"))
			assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
			assert.Regexp(t, `^"[0-9a-f]{32}"$`, rec.Header().Get("ETag"))
		})
	}
}

func TestListsGiveAnEmptyListWithNoParticipants(t *testing.T) {
	for name, l := range lists {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, &fakeReader{}, l.path, http.Header{"X-V": {l.version}})

			require.Equal(t, http.StatusOK, rec.Code)

			var body map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.JSONEq(t, `[]`, string(body["data"]))
		})
	}
}

// Verifies: REG-0328
func TestTheETagIsTheSameForTheSameDataAndChangesWithTheData(t *testing.T) {
	reader := newReader()
	first := serve(t, reader, recipientStatusesPath, v3()).Header().Get("ETag")
	again := serve(t, reader, recipientStatusesPath, v3()).Header().Get("ETag")
	reader.recipients[1].Status = domain.DataRecipientRevoked
	changed := serve(t, reader, recipientStatusesPath, v3()).Header().Get("ETag")

	assert.Equal(t, first, again)
	assert.NotEqual(t, first, changed)
}

// Verifies: REG-0326, REG-0327
func TestIfNoneMatchGivesNotModifiedWhenAnEntityTagMatches(t *testing.T) {
	etag := serve(t, newReader(), recipientStatusesPath, v3()).Header().Get("ETag")

	tests := map[string]string{
		"same tag":         etag,
		"weak tag":         "W/" + etag,
		"tag in a list":    `"other", ` + etag,
		"any":              "*",
		"no space in list": `"other",` + etag,
	}
	for name, ifNoneMatch := range tests {
		t.Run(name, func(t *testing.T) {
			header := v3()
			header.Set("If-None-Match", ifNoneMatch)

			rec := serve(t, newReader(), recipientStatusesPath, header)

			assert.Equal(t, http.StatusNotModified, rec.Code)
			assert.Empty(t, rec.Body.String())
			assert.Equal(t, etag, rec.Header().Get("ETag"))
			assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		})
	}
}

func TestIfNoneMatchGivesTheListWhenNoEntityTagMatches(t *testing.T) {
	etag := serve(t, newReader(), recipientStatusesPath, v3()).Header().Get("ETag")
	for name, ifNoneMatch := range map[string]string{
		"other tag":          `"other"`,
		"tag with no quotes": etag[1 : len(etag)-1],
	} {
		t.Run(name, func(t *testing.T) {
			header := v3()
			header.Set("If-None-Match", ifNoneMatch)

			rec := serve(t, newReader(), recipientStatusesPath, header)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, etag, rec.Header().Get("ETag"))
		})
	}
}

func TestVersionNegotiation(t *testing.T) {
	tests := map[string]struct {
		xv, minV string
	}{
		"x-v is the supported version":           {xv: "3"},
		"x-v is higher, x-min-v permits 3":       {xv: "99", minV: "2"},
		"x-min-v is equal to x-v and is ignored": {xv: "3", minV: "3"},
		"x-min-v is higher than x-v":             {xv: "3", minV: "5"},
		"x-v is too large for an integer":        {xv: "99999999999999999999", minV: "1"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			header := http.Header{"X-V": {tt.xv}}
			if tt.minV != "" {
				header.Set("X-Min-V", tt.minV)
			}

			rec := serve(t, newReader(), recipientStatusesPath, header)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "3", rec.Header().Get("X-V"))
		})
	}
}

func TestVersionErrors(t *testing.T) {
	const (
		missing     = "urn:au-cds:error:cds-all:Header/Missing"
		invalid     = "urn:au-cds:error:cds-all:Header/InvalidVersion"
		unsupported = "urn:au-cds:error:cds-all:Header/UnsupportedVersion"
	)

	tests := map[string]struct {
		header http.Header
		status int
		code   string
		title  string
	}{
		"no x-v":                {http.Header{}, http.StatusBadRequest, missing, "Missing Required Header"},
		"x-v is not an integer": {http.Header{"X-V": {"abc"}}, http.StatusBadRequest, invalid, "Invalid Version"},
		"x-v is zero":           {http.Header{"X-V": {"0"}}, http.StatusBadRequest, invalid, "Invalid Version"},
		"x-v is negative":       {http.Header{"X-V": {"-3"}}, http.StatusBadRequest, invalid, "Invalid Version"},
		"x-v has a sign":        {http.Header{"X-V": {"+3"}}, http.StatusBadRequest, invalid, "Invalid Version"},
		"x-min-v is not valid": {
			http.Header{"X-V": {"3"}, "X-Min-V": {"abc"}},
			http.StatusBadRequest,
			invalid,
			"Invalid Version",
		},
		"x-v is higher": {
			http.Header{"X-V": {"99"}},
			http.StatusNotAcceptable,
			unsupported,
			"Unsupported Version",
		},
		"x-v is lower": {
			http.Header{"X-V": {"2"}},
			http.StatusNotAcceptable,
			unsupported,
			"Unsupported Version",
		},
		"range is lower": {
			http.Header{"X-V": {"2"}, "X-Min-V": {"1"}},
			http.StatusNotAcceptable,
			unsupported,
			"Unsupported Version",
		},
		"ignored x-min-v, x-v 4": {
			http.Header{"X-V": {"4"}, "X-Min-V": {"4"}},
			http.StatusNotAcceptable,
			unsupported,
			"Unsupported Version",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for _, path := range []string{recipientStatusesPath, productStatusesPath} {
				rec := serve(t, newReader(), path, tt.header)

				requireError(t, rec, tt.status, tt.code, tt.title)
				assert.Empty(t, rec.Header().Get("X-V"))
			}
		})
	}
}

// Verifies: REG-1494
func TestAListRefusesAnAcceptHeaderThatPermitsNoJSON(t *testing.T) {
	for name, l := range lists {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, newReader(), l.path, http.Header{"X-V": {l.version}, "Accept": {"application/xml"}})

			e := requireError(
				t, rec, http.StatusNotAcceptable, "urn:au-cds:error:cds-all:Header/Invalid", "Invalid Header",
			)
			assert.Equal(t, "Accept", e.Detail)
		})
	}
}

func TestMissingVersionDetailIsTheHeaderName(t *testing.T) {
	rec := serve(t, newReader(), recipientStatusesPath, http.Header{})

	e := requireError(
		t,
		rec,
		http.StatusBadRequest,
		"urn:au-cds:error:cds-all:Header/Missing",
		"Missing Required Header",
	)
	assert.Equal(t, "x-v", e.Detail)
}

// Verifies: REG-0310, REG-0311
func TestDataHolderStatusesGiveEachDataHolderOfTheIndustry(t *testing.T) {
	for _, industry := range []string{"all", "banking", "energy", "non-bank-lending", "telco"} {
		t.Run(industry, func(t *testing.T) {
			reader := newReader()
			path := "/cdr-register/v1/" + industry + "/data-holders/status"

			rec := serve(t, reader, path, http.Header{"X-V": {"2"}})

			require.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(
				t,
				`{"data":[{"legalEntityId":"dh-1","status":"ACTIVE"},{"legalEntityId":"dh-2","status":"REMOVED"}],`+
					`"links":{"self":"https://register.example`+path+`"},"meta":{}}`,
				rec.Body.String(),
			)
			assert.Equal(t, "2", rec.Header().Get("X-V"))
			assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
			assert.Regexp(t, `^"[0-9a-f]{32}"$`, rec.Header().Get("ETag"))

			if industry == "all" {
				assert.Nil(t, reader.holderIndustry)
			} else {
				require.NotNil(t, reader.holderIndustry)
				assert.Equal(t, domain.Industry(industry), *reader.holderIndustry)
			}
		})
	}
}

// Verifies: REG-0302, REG-0313
func TestDataHolderStatusesServeOnlyVersion2(t *testing.T) {
	rec := serve(t, newReader(), holderStatusesPath, http.Header{"X-V": {"1"}})

	requireError(
		t,
		rec,
		http.StatusNotAcceptable,
		"urn:au-cds:error:cds-all:Header/UnsupportedVersion",
		"Unsupported Version",
	)
}

// Verifies: REG-0316
func TestAnIndustryThatIsNotValidOnDataHolderStatuses(t *testing.T) {
	for _, industry := range []string{"foo", "Banking", "ALL"} {
		t.Run(industry, func(t *testing.T) {
			rec := serve(t, newReader(), "/cdr-register/v1/"+industry+"/data-holders/status", http.Header{"X-V": {"2"}})

			e := requireError(t, rec, http.StatusBadRequest, "urn:au-cds:error:cds-all:Field/Invalid", "Invalid Field")
			assert.Equal(t, "industry", e.Detail)
			assert.Equal(t, "2", rec.Header().Get("X-V"), "the version is negotiated before the industry is checked")
		})
	}
}

// Verifies: REG-0309, REG-0316
func TestAnIndustryOtherThanAllIsNotValidOnDataRecipientEndpoints(t *testing.T) {
	for _, industry := range []string{"banking", "energy", "foo"} {
		for name, l := range lists {
			if !strings.Contains(l.path, "/data-recipients") {
				continue
			}

			path := strings.Replace(l.path, "/all/", "/"+industry+"/", 1)
			t.Run(name+" "+industry, func(t *testing.T) {
				rec := serve(t, newReader(), path, http.Header{"X-V": {l.version}})

				e := requireError(
					t,
					rec,
					http.StatusBadRequest,
					"urn:au-cds:error:cds-all:Field/Invalid",
					"Invalid Field",
				)
				assert.Equal(t, "industry", e.Detail)
				assert.Equal(
					t,
					l.version,
					rec.Header().Get("X-V"),
					"the version is negotiated before the industry is checked",
				)
			})
		}
	}
}

// Verifies: REG-0312
func TestTheVersionIsCheckedBeforeTheIndustry(t *testing.T) {
	rec := serve(t, newReader(), "/cdr-register/v1/banking/data-recipients/status", http.Header{})

	requireError(t, rec, http.StatusBadRequest, "urn:au-cds:error:cds-all:Header/Missing", "Missing Required Header")
}

// Verifies: REG-0318
func TestAnUnknownPathIsNotFound(t *testing.T) {
	for _, path := range []string{
		"/",
		"/cdr-register/v1/all/data-recipients/brands",
		"/cdr-register/v1/all/data-recipients/status/extra",
	} {
		t.Run(path, func(t *testing.T) {
			rec := serve(t, newReader(), path, v3())

			requireError(
				t,
				rec,
				http.StatusNotFound,
				"urn:au-cds:error:cds-all:Resource/NotFound",
				"Resource Not Found",
			)
		})
	}
}

// Verifies: REG-1495
func TestAnotherMethodOnAListPathIsNotAllowed(t *testing.T) {
	for name, l := range lists {
		t.Run(name, func(t *testing.T) {
			rec := send(t, newReader(), http.MethodPost, l.path, http.Header{"X-V": {l.version}})

			e := requireError(
				t, rec, http.StatusMethodNotAllowed,
				"urn:au-cds:error:cds-all:GeneralError/Expected", "Expected Error Encountered",
			)
			assert.Equal(t, "The endpoint does not support the method POST.", e.Detail)
			assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
		})
	}
}

func TestAReadErrorIsAnUnexpectedErrorThatDoesNotShowTheCause(t *testing.T) {
	reader := &fakeReader{err: errors.New("connection to 10.0.0.1 failed")}

	rec := serve(t, reader, recipientStatusesPath, v3())

	e := requireError(
		t,
		rec,
		http.StatusInternalServerError,
		"urn:au-cds:error:cds-all:GeneralError/Unexpected",
		"Unexpected Error Encountered",
	)
	assert.NotContains(t, e.Detail, "10.0.0.1")
}

func TestAllResponsesPermitAnyOrigin(t *testing.T) {
	for name, header := range map[string]http.Header{"success": v3(), "error": {}} {
		t.Run(name, func(t *testing.T) {
			for _, path := range []string{recipientStatusesPath, "/unknown"} {
				rec := serve(t, newReader(), path, header)

				assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

// Verifies: REG-0314
func TestTheResponseGivesTheRequestInteractionID(t *testing.T) {
	for name, header := range map[string]http.Header{"success": v3(), "error": {}} {
		t.Run(name, func(t *testing.T) {
			header.Set("x-fapi-interaction-id", "6ba7b810-9dad-11d1-80b4-00c04fd430c8")

			rec := serve(t, newReader(), recipientStatusesPath, header)

			assert.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", rec.Header().Get("x-fapi-interaction-id"))
		})
	}
}

func TestTheResponseGivesANewInteractionIDWhenTheRequestHasNone(t *testing.T) {
	first := serve(t, newReader(), recipientStatusesPath, v3()).Header().Get("x-fapi-interaction-id")
	second := serve(t, newReader(), "/unknown", v3()).Header().Get("x-fapi-interaction-id")

	for _, id := range []string{first, second} {
		parsed, err := uuid.Parse(id)
		require.NoError(t, err)
		assert.Equal(t, uuid.RFC4122, parsed.Variant())
	}

	assert.NotEqual(t, first, second)
}

func TestSelfLinkIgnoresATrailingSlashInTheBaseURL(t *testing.T) {
	mux := http.NewServeMux()
	reader := newReader()
	httpapi.Register(mux, reader, noKeys, baseURL+"/")

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, recipientStatusesPath, nil)
	req.Header.Set("x-v", "3")

	mux.ServeHTTP(rec, req)

	var body struct {
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, baseURL+recipientStatusesPath, body.Links.Self)
}

// Verifies: REG-0340, REG-0341, REG-0342
func TestTheBrandsSummaryGivesThePublicDetailsOfEachBrandOfTheIndustry(t *testing.T) {
	for _, industry := range []string{"all", "banking", "energy", "non-bank-lending", "telco"} {
		t.Run(industry, func(t *testing.T) {
			reader := newReader()
			path := "/cdr-register/v1/" + industry + "/data-holders/brands/summary"

			rec := serve(t, reader, path, http.Header{"X-V": {"2"}})

			require.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, `{"data":[{
					"dataHolderBrandId":"dhb-1","interimId":"dhi-1","brandName":"Holder Bank",
					"brandGroup":"Holder Group",
					"publicBaseUri":"https://api.holder.example","productBaseUri":"https://prd.holder.example",
					"logoUri":"https://holder.example/logo.png","industries":["banking","energy"],
					"lastUpdated":"2026-09-30T02:15:30Z"
				},{
					"interimId":"dhi-2","brandName":"Holder Energy","publicBaseUri":"https://prd.energy.example",
					"logoUri":"https://holder.example/energy.png","industries":["energy"],
					"lastUpdated":"2026-09-01T00:00:00Z"
				}],
				"links":{"self":"https://register.example`+path+`"},"meta":{}}`, rec.Body.String())
			assert.Equal(t, "2", rec.Header().Get("X-V"))
			assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
			assert.Regexp(t, `^"[0-9a-f]{32}"$`, rec.Header().Get("ETag"))

			if industry == "all" {
				assert.Nil(t, reader.brandIndustry)
			} else {
				require.NotNil(t, reader.brandIndustry)
				assert.Equal(t, domain.Industry(industry), *reader.brandIndustry)
			}
		})
	}
}

// Verifies: REG-0315
func TestAnIndustryThatIsNotValidOnTheBrandsSummaryIsNotFound(t *testing.T) {
	for _, industry := range []string{"foo", "Banking", "ALL"} {
		t.Run(industry, func(t *testing.T) {
			rec := serve(
				t,
				newReader(),
				"/cdr-register/v1/"+industry+"/data-holders/brands/summary",
				http.Header{"X-V": {"2"}},
			)

			e := requireError(
				t,
				rec,
				http.StatusNotFound,
				"urn:au-cds:error:cds-register:Field/InvalidIndustry",
				"Invalid Industry",
			)
			assert.Equal(t, "industry", e.Detail)
			assert.Equal(t, "2", rec.Header().Get("X-V"), "the version is negotiated before the industry is checked")
		})
	}
}

func preflight(t *testing.T, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, newReader(), http.MethodOptions, path, header)
}

// Verifies: REG-0344
func TestAPreflightRequestPermitsGetWithTheRequestedHeadersFromAnyOrigin(t *testing.T) {
	for name, l := range lists {
		t.Run(name, func(t *testing.T) {
			rec := preflight(t, l.path, http.Header{
				"Origin":                         {"https://comparison.example"},
				"Access-Control-Request-Method":  {"GET"},
				"Access-Control-Request-Headers": {"x-v,x-min-v,if-none-match,x-fapi-interaction-id"},
			})

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET", rec.Header().Get("Access-Control-Allow-Methods"))
			assert.Equal(
				t,
				"x-v,x-min-v,if-none-match,x-fapi-interaction-id",
				rec.Header().Get("Access-Control-Allow-Headers"),
			)
			assert.Empty(t, rec.Body.String())
		})
	}
}

func TestAPreflightRequestWithNoRequestedHeadersPermitsNoHeaders(t *testing.T) {
	rec := preflight(t, recipientStatusesPath, http.Header{"Access-Control-Request-Method": {"GET"}})

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.NotContains(t, rec.Header(), "Access-Control-Allow-Headers")
}

func TestAnOptionsRequestThatIsNotAPreflightIsNotAPreflightResponse(t *testing.T) {
	rec := preflight(t, recipientStatusesPath, http.Header{})

	assert.NotEqual(t, http.StatusNoContent, rec.Code)
	assert.NotContains(t, rec.Header(), "Access-Control-Allow-Methods")
}

// Verifies: REG-0345
func TestResponsesExposeTheCDSHeadersToAnyOrigin(t *testing.T) {
	tests := map[string]struct {
		path   string
		header http.Header
	}{
		"success":      {brandsSummaryPath, http.Header{"X-V": {"2"}}},
		"error":        {brandsSummaryPath, http.Header{}},
		"unknown path": {"/unknown", http.Header{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, newReader(), tt.path, tt.header)

			assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "x-v, etag, x-fapi-interaction-id", rec.Header().Get("Access-Control-Expose-Headers"))
		})
	}
}
