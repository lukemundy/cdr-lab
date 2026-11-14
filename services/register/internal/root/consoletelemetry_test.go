package root

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/idp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

const searchText = "zebracorn"

// recordTelemetry has no URL redactor, so the test is stricter than the telemetry of the Register.
func recordTelemetry(t *testing.T) (*tracetest.SpanRecorder, *bytes.Buffer) {
	t.Helper()

	spans := tracetest.NewSpanRecorder()
	previousTracer := otel.GetTracerProvider()

	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)))

	var logs bytes.Buffer

	previousLogger := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))

	t.Cleanup(func() {
		otel.SetTracerProvider(previousTracer)
		slog.SetDefault(previousLogger)
	})

	return spans, &logs
}

func spanText(span sdktrace.ReadOnlySpan) string {
	parts := []string{span.Name(), span.Status().Description}
	for _, a := range span.Attributes() {
		parts = append(parts, a.Value.String())
	}

	for _, e := range span.Events() {
		parts = append(parts, e.Name)
		for _, a := range e.Attributes {
			parts = append(parts, a.Value.String())
		}
	}

	return strings.Join(parts, "\n")
}

func TestASearchOfTheConsolePutsNoSearchTextInTheSpansOrLogs(t *testing.T) {
	spans, logs := recordTelemetry(t)

	// The pool gets the tracer provider when it opens.
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	require.NoError(t, app.Seed(t.Context(), eventstore.New(pool, readmodel.Projections()...),
		domain.Participants{DataRecipients: []domain.DataRecipient{{
			LegalEntityID: "dr-1", LegalEntityName: "Zebracorn Pty Ltd", AccreditationNumber: "ADR-1",
			AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
			Status: domain.DataRecipientActive,
		}}, DataHolders: []domain.DataHolder{{
			LegalEntityID: "dh-1", LegalEntityName: "Mock Holdings Ltd", LogoURI: "https://holder.example/logo.png",
			Status: domain.DataHolderActive, Brands: []domain.DataHolderBrand{
				{
					DataHolderBrandID: "dh-1-b1",
					BrandName:         "Zebracorn Bank",
					Industries:        []domain.Industry{domain.Banking},
					LogoURI:           "https://holder.example/brand.png",
					Status:            domain.Active,
					JWKSEndpoint:      "https://holder.example/jwks",
					EndpointDetail: domain.EndpointDetail{
						Version: "1", PublicBaseURI: "https://api.holder.example",
						ResourceBaseURI: "https://mtls.holder.example", InfosecBaseURI: "https://id.holder.example",
						WebsiteURI: "https://holder.example",
					},
				},
			},
		}}}))

	listeners, err := consoleListeners(testConsole, ui, console.Ports{Readers: readmodel.NewQueries(pool)})
	require.NoError(t, err)

	// The spans of the seed also contain the table names, so only the spans of the requests are counted.
	seedSpans := len(spans.Ended())

	var targets []string
	for _, path := range []string{"/api/data-recipients", "/api/data-holders"} {
		targets = append(targets,
			path+"?q="+searchText,
			path+"?status=ACTIVE&q="+searchText+"&page=1",
			path+"?q="+searchText+"&page=0",
		)
	}

	for _, target := range targets {
		rec := httptest.NewRecorder()
		listeners[0].Handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
		require.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, rec.Code)
	}

	ended := spans.Ended()
	routes, queries := map[string]int{}, map[string]int{}

	for i, span := range ended {
		text := strings.ToLower(spanText(span))
		assert.NotContains(t, text, searchText, "span %q", span.Name())

		switch {
		case i < seedSpans:
		case strings.HasPrefix(span.Name(), "GET /api/"):
			routes[span.Name()]++
		case strings.Contains(text, "from data_recipients"):
			queries["data_recipients"]++
		case strings.Contains(text, "from data_holders"):
			queries["data_holders"]++
		}
	}

	assert.Equal(t, map[string]int{"GET /api/data-recipients": 3, "GET /api/data-holders": 3}, routes,
		"the test sees the server span of each request")
	assert.Positive(t, queries["data_recipients"], "the test sees the database spans of the data recipients")
	assert.Positive(t, queries["data_holders"], "the test sees the database spans of the data holders")
	assert.NotContains(t, strings.ToLower(logs.String()), searchText)
}

func TestTheKeySetsOfTheConsolePutNoKeyMaterialInTheSpansOrLogs(t *testing.T) {
	spans, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := idp.NewStore(pool)
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.AccessTokenKeys))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))

	listeners, err := consoleListeners(testConsole, ui, console.Ports{Keys: store, KeyChanger: store})
	require.NoError(t, err)

	var material []string

	// The material of each key before each step, because a publish deletes the previous key.
	collect := func() {
		for _, set := range []app.KeySet{app.AccessTokenKeys, app.SSAKeys} {
			keys, err := store.SigningKeys(set)(t.Context())
			require.NoError(t, err)

			for _, jwk := range authserver.PublicJWKS(keys).Keys {
				material = append(material, jwk.N)
			}

			for _, k := range keys {
				der, err := k.MarshalPKCS8()
				require.NoError(t, err)

				material = append(material, base64.StdEncoding.EncodeToString(der), hex.EncodeToString(der))
			}
		}
	}

	requests := []struct {
		method, target, body string
		status               int
	}{
		{method: http.MethodGet, target: "/api/key-sets", status: http.StatusOK},
		{
			method: http.MethodPost, target: "/api/key-sets/ssa/activate-next", body: `{"version": 2}`,
			status: http.StatusNoContent,
		},
		{
			method: http.MethodPost, target: "/api/key-sets/ssa/publish-next", body: `{"version": 3}`,
			status: http.StatusNoContent,
		},
		{
			method: http.MethodPost, target: "/api/key-sets/ssa/publish-next", body: `{"version": 4}`,
			status: http.StatusConflict,
		},
		{
			method: http.MethodPost, target: "/api/key-sets/ssa/activate-next", body: `{"version": 1}`,
			status: http.StatusConflict,
		},
		{method: http.MethodGet, target: "/api/key-sets", status: http.StatusOK},
	}
	for _, r := range requests {
		collect()

		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), r.method, r.target, strings.NewReader(r.body))
		listeners[0].Handler.ServeHTTP(rec, req)
		require.Equal(t, r.status, rec.Code, "%s %s: %s", r.method, r.target, rec.Body.String())
	}

	collect()

	routes := map[string]int{}

	for _, span := range spans.Ended() {
		text := spanText(span)
		for _, m := range material {
			assert.NotContains(t, text, m, "span %q", span.Name())
		}

		routes[span.Name()]++
	}

	assert.Equal(t, 2, routes["GET /api/key-sets"], "the test sees the server span of each request")
	assert.Equal(t, 2, routes["POST /api/key-sets/{keySet}/publish-next"])
	assert.Equal(t, 2, routes["POST /api/key-sets/{keySet}/activate-next"])

	for _, m := range material {
		assert.NotContains(t, logs.String(), m)
	}
}

// The test does not check url.path, because recordTelemetry has no URL redactor to set it to the route.
func TestADetailOfTheConsolePutsNoLegalEntityIDInTheSpansOrLogs(t *testing.T) {
	const recipientID, holderID, unknownID = "zebracorn-recipient", "zebracorn-holder", "zebracorn-unknown"

	spans, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	require.NoError(t, app.Seed(t.Context(), eventstore.New(pool, readmodel.Projections()...),
		domain.Participants{DataRecipients: []domain.DataRecipient{{
			LegalEntityID: recipientID, LegalEntityName: "Mock Budgeting Pty Ltd", AccreditationNumber: "ADR-1",
			AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
			Status: domain.DataRecipientActive,
		}}, DataHolders: []domain.DataHolder{{
			LegalEntityID: holderID, LegalEntityName: "Mock Holdings Ltd", LogoURI: "https://holder.example/logo.png",
			Status: domain.DataHolderActive,
		}}}))

	listeners, err := consoleListeners(testConsole, ui, console.Ports{Readers: readmodel.NewQueries(pool)})
	require.NoError(t, err)

	seedSpans := len(spans.Ended())

	targets := map[string]int{
		"/api/data-recipients/" + recipientID: http.StatusOK,
		"/api/data-recipients/" + unknownID:   http.StatusNotFound,
		"/api/data-holders/" + holderID:       http.StatusOK,
		"/api/data-holders/" + unknownID:      http.StatusNotFound,
	}
	for target, status := range targets {
		rec := httptest.NewRecorder()
		listeners[0].Handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
		require.Equal(t, status, rec.Code, target)
	}

	routes := map[string]int{}

	for _, span := range spans.Ended()[seedSpans:] {
		parts := []string{span.Name(), span.Status().Description}

		for _, a := range span.Attributes() {
			switch a.Key {
			case semconv.URLPathKey:
			case semconv.HTTPRouteKey:
				routes[a.Value.AsString()]++
			default:
				parts = append(parts, a.Value.String())
			}
		}

		for _, e := range span.Events() {
			parts = append(parts, e.Name)
			for _, a := range e.Attributes {
				parts = append(parts, a.Value.String())
			}
		}

		assert.NotContains(t, strings.Join(parts, "\n"), "zebracorn", "span %q", span.Name())
	}

	assert.Equal(t, map[string]int{
		"/api/data-recipients/{legalEntityId}": 2,
		"/api/data-holders/{legalEntityId}":    2,
	}, routes, "each server span has the route pattern")
	assert.NotContains(t, logs.String(), "zebracorn")
}

// The payloads of the events contain "zebracorn" in the IDs, the names and the thumbprint.
func TestAHistoryOfTheConsolePutsNoLegalEntityIDAndNoPayloadInTheSpansOrLogs(t *testing.T) {
	const recipientID, holderID, unknownID = "zebracorn-recipient", "zebracorn-holder", "zebracorn-unknown"

	spans, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{{
		LegalEntityID: recipientID, LegalEntityName: "Zebracorn Pty Ltd", AccreditationNumber: "ADR-1",
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://zebracorn.example/logo.png",
		Status: domain.DataRecipientActive, Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "zebracorn-brand", BrandName: "Zebracorn", LogoURI: "https://zebracorn.example/b.png",
			Status: domain.Active, SoftwareProducts: []domain.SoftwareProduct{{
				SoftwareProductID: "zebracorn-product", SoftwareProductName: "Zebracorn app",
				SoftwareProductDescription: "Zebracorn", LogoURI: "https://zebracorn.example/p.png",
				Status: domain.Active, ClientURI: "https://zebracorn.example",
				RedirectURIs: []string{"https://zebracorn.example/cb"}, JWKSURI: "https://zebracorn.example/jwks",
				RevocationURI: "https://zebracorn.example/revoke", RecipientBaseURI: "https://zebracorn.example",
				Scopes: []string{"openid"},
			}},
		}},
	}}, DataHolders: []domain.DataHolder{{
		LegalEntityID: holderID, LegalEntityName: "Zebracorn Holdings", LogoURI: "https://zebracorn.example/h.png",
		Status: domain.DataHolderActive,
	}}}))

	data, err := app.DataRecipientType.Encode([]domain.DataRecipientEvent{
		domain.ClientCertificateRegistered{SoftwareProductID: "zebracorn-product", Thumbprint: "zebracorn-thumbprint"},
	})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: recipientID, ExpectedVersion: 1, Events: data,
	}))

	listeners, err := consoleListeners(testConsole, ui, console.Ports{Readers: readmodel.NewQueries(pool)})
	require.NoError(t, err)

	seedSpans := len(spans.Ended())

	targets := map[string]int{
		"/api/data-recipients/" + recipientID + "/history":          http.StatusOK,
		"/api/data-recipients/" + recipientID + "/history?page=0":   http.StatusBadRequest,
		"/api/data-recipients/" + unknownID + "/history":            http.StatusNotFound,
		"/api/data-holders/" + holderID + "/history":                http.StatusOK,
		"/api/data-holders/" + unknownID + "/history?pageSize=1000": http.StatusBadRequest,
	}
	for target, status := range targets {
		rec := httptest.NewRecorder()
		listeners[0].Handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
		require.Equal(t, status, rec.Code, target)
	}

	routes, historyQueries := map[string]int{}, 0

	for _, span := range spans.Ended()[seedSpans:] {
		parts := []string{span.Name(), span.Status().Description}

		for _, a := range span.Attributes() {
			switch a.Key {
			case semconv.URLPathKey:
			case semconv.HTTPRouteKey:
				routes[a.Value.AsString()]++
			default:
				parts = append(parts, a.Value.String())
			}
		}

		for _, e := range span.Events() {
			parts = append(parts, e.Name)
			for _, a := range e.Attributes {
				parts = append(parts, a.Value.String())
			}
		}

		text := strings.ToLower(strings.Join(parts, "\n"))
		assert.NotContains(t, text, "zebracorn", "span %q", span.Name())

		if strings.Contains(text, "from participant_history") {
			historyQueries++
		}
	}

	assert.Equal(t, map[string]int{
		"/api/data-recipients/{legalEntityId}/history": 3,
		"/api/data-holders/{legalEntityId}/history":    2,
	}, routes, "each server span has the route pattern")
	assert.Positive(t, historyQueries, "the test sees the database spans of the history")
	assert.NotContains(t, strings.ToLower(logs.String()), "zebracorn")
}

// The test does not check url.path, because recordTelemetry has no URL redactor to set it to the route.
func TestAStatusActionOfTheConsolePutsNoIdentifierAndNoRequestBodyInTheSpansOrLogs(t *testing.T) {
	const recipientID, productID = "zebracorn-recipient", "zebracorn-product"

	spans, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{{
		LegalEntityID: recipientID, LegalEntityName: "Zebracorn Pty Ltd", AccreditationNumber: "ADR-1",
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
		Status: domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "zebracorn-brand", BrandName: "Zebracorn", LogoURI: "https://recipient.example/b.png",
			Status: domain.Active,
			SoftwareProducts: []domain.SoftwareProduct{{
				SoftwareProductID: productID, SoftwareProductName: "Zebracorn Web",
				SoftwareProductDescription: "A product", LogoURI: "https://recipient.example/p.png",
				Status: domain.Active, ClientURI: "https://recipient.example",
				RedirectURIs: []string{"https://recipient.example/callback"}, JWKSURI: "https://recipient.example/jwks",
				RevocationURI: "https://recipient.example/revoke", RecipientBaseURI: "https://cdr.recipient.example",
				Scopes: []string{"openid"},
			}},
		}},
	}}}))

	listeners, err := consoleListeners(testConsole, ui, console.Ports{
		Readers: readmodel.NewQueries(pool), Operator: app.NewOperator(store),
	})
	require.NoError(t, err)

	seedSpans := len(spans.Ended())

	recipientPath := "/api/data-recipients/" + recipientID
	productPath := recipientPath + "/software-products/" + productID

	requests := []struct {
		target, body string
		status       int
	}{
		{target: recipientPath + "/suspend", body: `{"version": 1}`, status: http.StatusNoContent},
		{target: recipientPath + "/revoke", body: `{"version": 1}`, status: http.StatusConflict},
		// The status rules refuse it, and the rule names the software product.
		{target: productPath + "/activate", body: `{"version": 3}`, status: http.StatusConflict},
		{
			target: recipientPath + "/software-products/zebracorn-unknown/remove", body: `{"version": 3}`,
			status: http.StatusNotFound,
		},
		{
			target: recipientPath + "/reinstate",
			body:   `{"version": 3, "zebracorn": true}`,
			status: http.StatusBadRequest,
		},
	}
	for _, r := range requests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, r.target, strings.NewReader(r.body))
		listeners[0].Handler.ServeHTTP(rec, req)
		require.Equal(t, r.status, rec.Code, "%s: %s", r.target, rec.Body.String())
	}

	routes := map[string]int{}

	for _, span := range spans.Ended()[seedSpans:] {
		parts := []string{span.Name(), span.Status().Description}

		for _, a := range span.Attributes() {
			switch a.Key {
			case semconv.URLPathKey:
			case semconv.HTTPRouteKey:
				routes[a.Value.AsString()]++
			default:
				parts = append(parts, a.Value.String())
			}
		}

		for _, e := range span.Events() {
			parts = append(parts, e.Name)
			for _, a := range e.Attributes {
				parts = append(parts, a.Value.String())
			}
		}

		assert.NotContains(t, strings.ToLower(strings.Join(parts, "\n")), "zebracorn", "span %q", span.Name())
	}

	productRoute := "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/"
	assert.Equal(t, map[string]int{
		"/api/data-recipients/{legalEntityId}/suspend":   1,
		"/api/data-recipients/{legalEntityId}/revoke":    1,
		"/api/data-recipients/{legalEntityId}/reinstate": 1,
		productRoute + "activate":                        1,
		productRoute + "remove":                          1,
	}, routes, "each server span has the route pattern")
	assert.NotContains(t, strings.ToLower(logs.String()), "zebracorn")
}

func TestAChangeACreateOrAnAddOfTheConsolePutsNoRequestBodyInTheSpansOrLogs(t *testing.T) {
	const recipientID = "zebracorn-recipient"

	spans, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{{
		LegalEntityID: recipientID, LegalEntityName: "Zebracorn Pty Ltd", AccreditationNumber: "ADR-1",
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
		Status: domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "brand-1", BrandName: "Brand", LogoURI: "https://recipient.example/brand.png",
			Status: domain.Active, SoftwareProducts: []domain.SoftwareProduct{{
				SoftwareProductID: "product-1", SoftwareProductName: "Product", SoftwareProductDescription: "A product",
				LogoURI: "https://recipient.example/product.png", Status: domain.Active,
				ClientURI: "https://recipient.example", RedirectURIs: []string{"https://recipient.example/callback"},
				JWKSURI: "https://recipient.example/jwks", RevocationURI: "https://recipient.example/revoke",
				RecipientBaseURI: "https://cdr.recipient.example", Scopes: []string{"openid"},
			}},
		}},
	}}}))

	listeners, err := consoleListeners(testConsole, ui, console.Ports{
		Readers: readmodel.NewQueries(pool), Operator: app.NewOperator(store),
	})
	require.NoError(t, err)

	seedSpans := len(spans.Ended())
	path := "/api/data-recipients/" + recipientID
	details := `"legalEntityName": "Zebracorn Group", "accreditationNumber": "ADR-zebracorn",
		"accreditationLevel": "SPONSORED", "logoUri": "https://zebracorn.example/logo.png"`
	product := `"softwareProductName": "Zebracorn Group", "softwareProductDescription": "ADR-zebracorn",
		"logoUri": "https://zebracorn.example/logo.png", "clientUri": "https://zebracorn.example",
		"redirectUris": ["https://zebracorn.example/cb"], "jwksUri": "https://zebracorn.example/jwks",
		"revocationUri": "https://zebracorn.example/revoke", "recipientBaseUri": "https://zebracorn.example",
		"scopes": []`
	create := `{"legalEntityId": "zebracorn-created", "status": "ACTIVE", ` + details + `}`
	brand := `{"version": 3, "dataRecipientBrandId": "zebracorn-brand", "brandName": "Zebracorn Group",
		"logoUri": "https://zebracorn.example/logo.png", "status": "ACTIVE"}`
	// The test checks no span for the legal entity ID of the data holder, because the path of its change has it.
	holder := `{"legalEntityId": "zebracorn-holder", "legalEntityName": "Zebracorn Group",
		"logoUri": "https://zebracorn.example/logo.png", "status": "ACTIVE"}`
	// The test checks no span for the interim ID of the brand, because the path of its change has it.
	holderBrand := func(version string) string {
		return `{"version": ` + version + `, "interimId": "zebracorn-interim", "brandName": "Zebracorn Group",
			"industries": ["banking"], "logoUri": "https://zebracorn.example/logo.png", "status": "ACTIVE",
			"endpointDetail": {"publicBaseUri": "https://zebracorn.example"}}`
	}
	// The path has brand-1, because the test checks the path of each span for the values of the bodies.
	added := func(version string) string {
		return `{"version": ` + version + `, "softwareProductId": "zebracorn-product", "status": "ACTIVE", ` +
			product + `}`
	}

	for _, r := range []struct {
		target, body string
		status       int
	}{
		{target: path + "/change", body: `{"version": 1, ` + details + `}`, status: http.StatusNoContent},
		{target: path + "/change", body: `{"version": 1, ` + details + `}`, status: http.StatusConflict},
		{
			target: path + "/change",
			body:   `{"version": 2, "legalEntityName": "Zebracorn Group", "accreditationLevel": "ZEBRACORN"}`,
			status: http.StatusBadRequest,
		},
		{
			target: path + "/brands/brand-1/change",
			body:   `{"version": 2, "brandName": "Zebracorn Group", "logoUri": "https://zebracorn.example/logo.png"}`,
			status: http.StatusNoContent,
		},
		{
			target: path + "/software-products/product-1/change",
			body:   `{"version": 3, ` + product + `}`,
			status: http.StatusBadRequest,
		},
		{target: "/api/data-recipients", body: create, status: http.StatusCreated},
		{target: "/api/data-recipients", body: create, status: http.StatusConflict},
		{target: path + "/brands", body: brand, status: http.StatusCreated},
		{target: path + "/brands/brand-1/software-products", body: added("4"), status: http.StatusBadRequest},
		{target: "/api/data-holders", body: holder, status: http.StatusCreated},
		{target: "/api/data-holders", body: holder, status: http.StatusConflict},
		{
			target: "/api/data-holders/zebracorn-holder/change",
			body: `{"version": 1, "legalEntityName": "Zebracorn Group",
				"logoUri": "https://zebracorn.example/logo.png"}`,
			status: http.StatusNoContent,
		},
		{target: "/api/data-holders/zebracorn-holder/brands", body: holderBrand("2"), status: http.StatusCreated},
		{target: "/api/data-holders/zebracorn-holder/brands", body: holderBrand("3"), status: http.StatusConflict},
		{
			target: "/api/data-holders/zebracorn-holder/brands/zebracorn-interim/change",
			body: `{"version": 3, "brandName": "Zebracorn Group", "industries": ["banking"],
				"logoUri": "https://zebracorn.example/logo.png", "endpointDetail": {"publicBaseUri": ""}}`,
			status: http.StatusBadRequest,
		},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, r.target, strings.NewReader(r.body))
		listeners[0].Handler.ServeHTTP(rec, req)
		require.Equal(t, r.status, rec.Code, rec.Body.String())
	}

	for _, span := range spans.Ended()[seedSpans:] {
		text := strings.ToLower(spanText(span))
		// The test has no URL redactor, so only the values of the body are checked.
		for _, value := range []string{
			"zebracorn group", "adr-zebracorn", "zebracorn.example", "zebracorn-created",
			"zebracorn-brand", "zebracorn-product",
		} {
			assert.NotContains(t, text, value, "span %q", span.Name())
		}
	}

	assert.NotContains(t, strings.ToLower(logs.String()), "zebracorn")
}
