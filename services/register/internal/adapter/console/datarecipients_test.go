package console_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/openapi"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

var ui = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Register console</title>")}}

func loadSpec(t *testing.T) *openapi.Spec {
	t.Helper()

	data, err := os.ReadFile("openapi.yaml")
	require.NoError(t, err)
	spec, err := openapi.Load(data, "")
	require.NoError(t, err)

	return spec
}

type response struct {
	status int
	header http.Header
	body   string
}

func get(t *testing.T, readers console.Readers, target string) response {
	t.Helper()

	return serve(t, readers, nil, target)
}

func serve(t *testing.T, readers console.Readers, keys app.KeyReader, target string) response {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)

	return do(t, console.Ports{Readers: readers, Keys: keys}, req)
}

// do validates the response against the OpenAPI document.
func do(t *testing.T, ports console.Ports, req *http.Request) response {
	t.Helper()

	mux := http.NewServeMux()
	require.NoError(t, console.Register(mux, ui, []string{"example.com"}, ports))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	resp.Request = req
	require.NoError(t, loadSpec(t).ValidateResponse(resp))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{status: resp.StatusCode, header: resp.Header, body: string(body)}
}

func product(id string, status domain.Status) domain.SoftwareProduct {
	return domain.SoftwareProduct{
		SoftwareProductID: id, SoftwareProductName: "Product " + id, SoftwareProductDescription: "A product",
		LogoURI: "https://recipient.example/product.png", Status: status, ClientURI: "https://recipient.example",
		RedirectURIs: []string{"https://recipient.example/callback"}, JWKSURI: "https://recipient.example/jwks",
		RevocationURI: "https://recipient.example/revoke", RecipientBaseURI: "https://cdr.recipient.example",
		Scopes: []string{"openid"},
	}
}

func recipient(
	id, name string,
	status domain.DataRecipientStatus,
	brands ...[]domain.SoftwareProduct,
) domain.DataRecipient {
	dr := domain.DataRecipient{
		LegalEntityID: id, LegalEntityName: name, AccreditationNumber: "ADR-" + id,
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png", Status: status,
	}
	for i, products := range brands {
		dr.Brands = append(dr.Brands, domain.DataRecipientBrand{
			DataRecipientBrandID: fmt.Sprintf("%s-b%d", id, i+1), BrandName: "Brand",
			LogoURI: "https://recipient.example/brand.png", Status: domain.Active, SoftwareProducts: products,
		})
	}

	return dr
}

func seed(t *testing.T, recipients ...domain.DataRecipient) *readmodel.Queries {
	t.Helper()

	return seedParticipants(t, domain.Participants{DataRecipients: recipients})
}

func seedParticipants(t *testing.T, participants domain.Participants) *readmodel.Queries {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, participants))

	return readmodel.NewQueries(pool)
}

func TestTheDocumentClosesEachObjectSchema(t *testing.T) {
	assert.NoError(t, loadSpec(t).CheckClosed())
}

func TestDataRecipientsGiveOnePageWithTheCountsAndTheTotal(t *testing.T) {
	queries := seed(t,
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Removed)},
			[]domain.SoftwareProduct{product("sp-3", domain.Inactive)}),
		recipient("dr-2", "Beta Pty Ltd", domain.DataRecipientSuspended),
		recipient("dr-3", "Gamma Pty Ltd", domain.DataRecipientActive),
	)

	resp := get(t, queries, "/api/data-recipients?pageSize=2")

	assert.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
	assert.JSONEq(t, `{"totalRecords": 3, "data": [
		{"legalEntityId": "dr-1", "legalEntityName": "Alpha Pty Ltd", "accreditationNumber": "ADR-dr-1",
		 "status": "ACTIVE", "brandCount": 2, "softwareProductCount": 3},
		{"legalEntityId": "dr-2", "legalEntityName": "Beta Pty Ltd", "accreditationNumber": "ADR-dr-2",
		 "status": "SUSPENDED", "brandCount": 0, "softwareProductCount": 0}
	]}`, resp.body)
}

type queryRecorder struct {
	got       app.DataRecipientPageQuery
	gotHolder app.DataHolderPageQuery
}

func (r *queryRecorder) DataRecipientPage(
	_ context.Context,
	query app.DataRecipientPageQuery,
) ([]app.DataRecipientItem, int, error) {
	r.got = query

	return nil, 0, nil
}

func (r *queryRecorder) DataHolderPage(
	_ context.Context,
	query app.DataHolderPageQuery,
) ([]app.DataHolderItem, int, error) {
	r.gotHolder = query

	return nil, 0, nil
}

func TestDataRecipientsGiveTheQueryParametersToTheReadModel(t *testing.T) {
	suspended := domain.DataRecipientSuspended

	tests := map[string]struct {
		query string
		want  app.DataRecipientPageQuery
	}{
		"defaults": {"", app.DataRecipientPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25}},
		"each parameter": {
			"?status=SUSPENDED&q=Mock%20B&sort=accreditationNumber&order=desc&page=3&pageSize=100",
			app.DataRecipientPageQuery{
				Status: &suspended, NameContains: "Mock B", Sort: app.SortByAccreditationNumber,
				Descending: true, Page: 3, PageSize: 100,
			},
		},
		"sort by status ascending": {
			"?sort=status&order=asc",
			app.DataRecipientPageQuery{Sort: app.SortByStatus, Page: 1, PageSize: 25},
		},
		"sort by brand count": {
			"?sort=brandCount",
			app.DataRecipientPageQuery{Sort: app.SortByBrandCount, Page: 1, PageSize: 25},
		},
		"sort by software product count descending": {
			"?sort=softwareProductCount&order=desc",
			app.DataRecipientPageQuery{Sort: app.SortBySoftwareProductCount, Descending: true, Page: 1, PageSize: 25},
		},
		"empty values are the defaults": {
			"?status=&q=&sort=&order=&page=&pageSize=",
			app.DataRecipientPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25},
		},
		"an unknown parameter is ignored": {
			"?unknown=1", app.DataRecipientPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := &queryRecorder{}

			resp := get(t, reader, "/api/data-recipients"+tt.query)

			assert.Equal(t, http.StatusOK, resp.status)
			assert.JSONEq(t, `{"totalRecords": 0, "data": []}`, resp.body)
			assert.Equal(t, tt.want, reader.got)
		})
	}
}

func TestDataRecipientsRefuseAParameterThatIsNotValid(t *testing.T) {
	tests := map[string]string{
		"page 0":                   "?page=0",
		"page not an integer":      "?page=one",
		"page size 0":              "?pageSize=0",
		"page size more than 100":  "?pageSize=101",
		"page size not an integer": "?pageSize=1.5",
		"unknown status":           "?status=DELETED",
		"status in lower case":     "?status=active",
		"unknown sort":             "?sort=lastUpdated",
		"unknown order":            "?order=up",
	}
	for name, query := range tests {
		t.Run(name, func(t *testing.T) {
			reader := &queryRecorder{}

			resp := get(t, reader, "/api/data-recipients"+query)

			assert.Equal(t, http.StatusBadRequest, resp.status)
			assert.Equal(t, "application/problem+json", resp.header.Get("Content-Type"))
			assert.Zero(t, reader.got, "the read model gets no query")
		})
	}
}

type failingReader struct{}

func (failingReader) DataRecipientPage(
	context.Context,
	app.DataRecipientPageQuery,
) ([]app.DataRecipientItem, int, error) {
	return nil, 0, errors.New("a message that contains Mock Budgeting")
}

func (failingReader) DataHolderPage(context.Context, app.DataHolderPageQuery) ([]app.DataHolderItem, int, error) {
	return nil, 0, errors.New("a message that contains Mock Holdings")
}

func TestDataRecipientsGiveAProblemWithNoErrorMessageWhenTheReadFails(t *testing.T) {
	resp := get(t, failingReader{}, "/api/data-recipients")

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Budgeting")
}

// The page size bounds the response, so the first page does not depend on the number of data recipients.
func TestTheFirstPageOf1000DataRecipientsHas25Records(t *testing.T) {
	recipients := make([]domain.DataRecipient, 0, 1000)

	for i := range 1000 {
		id := fmt.Sprintf("dr-%04d", i)
		recipients = append(recipients, recipient(id, "Recipient "+id, domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-"+id, domain.Active)}))
	}

	queries := seed(t, recipients...)

	start := time.Now()
	resp := get(t, queries, "/api/data-recipients?q=recipient&sort=status&order=desc")
	elapsed := time.Since(start)

	var list struct {
		Data         []json.RawMessage `json:"data"`
		TotalRecords int               `json:"totalRecords"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &list))
	assert.Len(t, list.Data, 25)
	assert.Equal(t, 1000, list.TotalRecords)
	assert.Less(t, elapsed, time.Second)
	t.Logf("first page of 1,000 data recipients: %s", elapsed)
}
