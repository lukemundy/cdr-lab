package root

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

type claimsSigner struct{ claims []domain.SSAClaims }

func (s *claimsSigner) SignSSA(_ context.Context, claims domain.SSAClaims) (string, error) {
	s.claims = append(s.claims, claims)
	return "signed-ssa", nil
}

// The lab browser test cannot get an SSA, because only the conformance suite serves a JWKS that the Register can get.
func TestTheNextSSAOfASoftwareProductHasTheRedirectURIsOfAChangeInTheConsole(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{{
		LegalEntityID: "dr-1", LegalEntityName: "Recipient Pty Ltd", AccreditationNumber: "ADR-1",
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
		Status: domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "brand-1", BrandName: "Brand", LogoURI: "https://recipient.example/brand.png",
			Status: domain.Active, SoftwareProducts: []domain.SoftwareProduct{{
				SoftwareProductID: "product-1", SoftwareProductName: "Product", SoftwareProductDescription: "A product",
				LogoURI: "https://recipient.example/product.png", Status: domain.Active,
				ClientURI: "https://recipient.example", RedirectURIs: []string{"https://recipient.example/callback"},
				JWKSURI: "https://recipient.example/jwks", RevocationURI: "https://recipient.example/revoke",
				RecipientBaseURI: "https://cdr.recipient.example", Scopes: []string{"openid", "profile"},
			}},
		}},
	}}}))

	queries := readmodel.NewQueries(pool)
	listeners, err := consoleListeners(testConsole, ui, console.Ports{
		Readers: queries, Operator: app.NewOperator(store),
	})
	require.NoError(t, err)

	body := `{"version": 1, "softwareProductName": "Product", "softwareProductDescription": "A product",
		"logoUri": "https://recipient.example/product.png", "clientUri": "https://recipient.example",
		"redirectUris": ["https://recipient.example/new", "https://recipient.example/second"],
		"jwksUri": "https://recipient.example/jwks", "revocationUri": "https://recipient.example/revoke",
		"recipientBaseUri": "https://cdr.recipient.example", "scopes": ["openid", "profile"]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost,
		"/api/data-recipients/dr-1/software-products/product-1/change", strings.NewReader(body))
	listeners[0].Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	signer := &claimsSigner{}
	issuer := app.SSAIssuer{
		Owners: queries, Signer: signer, Now: time.Now, NewID: func() string { return "jti-1" },
	}
	_, err = issuer.Issue(t.Context(), "product-1", "brand-1")
	require.NoError(t, err)

	require.Len(t, signer.claims, 1)
	assert.Equal(t, []string{"https://recipient.example/new", "https://recipient.example/second"},
		signer.claims[0].RedirectURIs)
}

func TestASoftwareProductThatTheOperatorAddsInTheConsoleGetsAnSSA(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	queries := readmodel.NewQueries(pool)
	listeners, err := consoleListeners(testConsole, ui, console.Ports{
		Readers: queries, Operator: app.NewOperator(store),
	})
	require.NoError(t, err)

	for _, r := range []struct{ target, body string }{
		{target: "/api/data-recipients", body: `{"legalEntityId": "dr-1", "legalEntityName": "Recipient Pty Ltd",
			"accreditationNumber": "ADR-1", "accreditationLevel": "UNRESTRICTED",
			"logoUri": "https://recipient.example/logo.png", "status": "ACTIVE"}`},
		{target: "/api/data-recipients/dr-1/brands", body: `{"version": 1, "dataRecipientBrandId": "brand-1",
			"brandName": "Brand", "logoUri": "https://recipient.example/brand.png", "status": "ACTIVE"}`},
		{target: "/api/data-recipients/dr-1/brands/brand-1/software-products", body: `{"version": 2,
			"softwareProductId": "product-1", "softwareProductName": "Product",
			"softwareProductDescription": "A product", "logoUri": "https://recipient.example/product.png",
			"clientUri": "https://recipient.example", "redirectUris": ["https://recipient.example/callback"],
			"jwksUri": "https://recipient.example/jwks", "revocationUri": "https://recipient.example/revoke",
			"recipientBaseUri": "https://cdr.recipient.example", "scopes": ["openid", "profile"], "status": "ACTIVE"}`},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, r.target, strings.NewReader(r.body))
		listeners[0].Handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code, "%s: %s", r.target, rec.Body.String())
	}

	signer := &claimsSigner{}
	issuer := app.SSAIssuer{
		Owners: queries, Signer: signer, Now: time.Now, NewID: func() string { return "jti-1" },
	}
	_, err = issuer.Issue(t.Context(), "product-1", "brand-1")
	require.NoError(t, err)

	require.Len(t, signer.claims, 1)
	assert.Equal(t, "product-1", signer.claims[0].SoftwareID)
	assert.Equal(t, "brand-1", signer.claims[0].OrgID)
	assert.Equal(t, []string{"https://recipient.example/callback"}, signer.claims[0].RedirectURIs)
	assert.Equal(t, "https://recipient.example/jwks", signer.claims[0].JWKSURI)
}

// callerToken gives each token to the software product, with the scope of the Register APIs.
type callerToken string

func (c callerToken) VerifyAccessToken(context.Context, string, *x509.Certificate) (authserver.AccessToken, error) {
	return authserver.AccessToken{ClientID: string(c), Scopes: []string{"cdr-register:read"}}, nil
}

// The lab browser test cannot call Get Data Holder Brands, because only the conformance suite serves a JWKS that the
// Register can get for the token.
type holderBrandsTest struct {
	console http.Handler
	secure  *http.ServeMux
}

func newHolderBrandsTest(t *testing.T) holderBrandsTest {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{{
		LegalEntityID: "dr-1", LegalEntityName: "Recipient Pty Ltd", AccreditationNumber: "ADR-1",
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
		Status: domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "brand-1", BrandName: "Brand", LogoURI: "https://recipient.example/brand.png",
			Status: domain.Active, SoftwareProducts: []domain.SoftwareProduct{{
				SoftwareProductID: "product-1", SoftwareProductName: "Product", SoftwareProductDescription: "A product",
				LogoURI: "https://recipient.example/product.png", Status: domain.Active,
				ClientURI: "https://recipient.example", RedirectURIs: []string{"https://recipient.example/callback"},
				JWKSURI: "https://recipient.example/jwks", RevocationURI: "https://recipient.example/revoke",
				RecipientBaseURI: "https://cdr.recipient.example", Scopes: []string{"openid", "profile"},
			}},
		}},
	}}}))

	queries := readmodel.NewQueries(pool)
	listeners, err := consoleListeners(testConsole, ui, console.Ports{
		Readers: queries, Operator: app.NewOperator(store),
	})
	require.NoError(t, err)

	secure := http.NewServeMux()
	httpapi.RegisterSecure(secure, queries, callerToken("product-1"), nil, "https://secure.register.example")

	return holderBrandsTest{console: listeners[0].Handler, secure: secure}
}

const holderEndpoints = `"endpointDetail": {
	"version": "1", "publicBaseUri": "https://api.holder.example", "resourceBaseUri": "https://mtls.holder.example",
	"infosecBaseUri": "https://id.holder.example", "websiteUri": "https://holder.example"
}, "jwksEndpoint": "https://id.holder.example/jwks"`

func (h holderBrandsTest) post(t *testing.T, target, body string, status int) {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, target, strings.NewReader(body))
	h.console.ServeHTTP(rec, req)
	require.Equal(t, status, rec.Code, "%s: %s", target, rec.Body.String())
}

func (h holderBrandsTest) createHolderWithBrand(t *testing.T) {
	t.Helper()

	h.post(t, "/api/data-holders", `{"legalEntityId": "dh-1", "legalEntityName": "Holder Ltd",
		"logoUri": "https://holder.example/logo.png", "status": "ACTIVE"}`, http.StatusCreated)
	h.post(t, "/api/data-holders/dh-1/brands", `{"version": 1, "dataHolderBrandId": "dh-brand-1",
		"brandName": "Holder Bank", "industries": ["banking"], "logoUri": "https://holder.example/brand.png",
		"status": "ACTIVE", `+holderEndpoints+`}`, http.StatusCreated)
}

type registerHolderBrand struct {
	DataHolderBrandID string `json:"dataHolderBrandId"`
	BrandName         string `json:"brandName"`
	Status            string `json:"status"`
	LegalEntity       struct {
		LegalEntityID string `json:"legalEntityId"`
	} `json:"legalEntity"`
}

func (h holderBrandsTest) dataHolderBrands(t *testing.T) []registerHolderBrand {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/cdr-register/v1/banking/data-holders/brands",
		nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("x-v", "3")
	h.secure.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Data []registerHolderBrand `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	return body.Data
}

func TestADataHolderBrandThatTheOperatorAddsAndChangesInTheConsoleIsOnGetDataHolderBrands(t *testing.T) {
	h := newHolderBrandsTest(t)
	h.createHolderWithBrand(t)
	h.post(t, "/api/data-holders/dh-1/brands/dh-brand-1/change", `{"version": 2,
		"brandName": "Renamed Bank", "industries": ["banking"], "logoUri": "https://holder.example/brand.png", `+
		holderEndpoints+`}`, http.StatusNoContent)

	brands := h.dataHolderBrands(t)
	require.Len(t, brands, 1)
	assert.Equal(t, "dh-brand-1", brands[0].DataHolderBrandID)
	assert.Equal(t, "Renamed Bank", brands[0].BrandName)
	assert.Equal(t, "dh-1", brands[0].LegalEntity.LegalEntityID)
}

func TestGetDataHolderBrandsGivesTheStatusAfterEachConsoleAction(t *testing.T) {
	h := newHolderBrandsTest(t)
	h.createHolderWithBrand(t)

	h.post(t, "/api/data-holders/dh-1/brands/dh-brand-1/deactivate", `{"version": 2}`, http.StatusNoContent)

	brands := h.dataHolderBrands(t)
	require.Len(t, brands, 1)
	assert.Equal(t, "INACTIVE", brands[0].Status)

	// Get Data Holder Brands also gives REMOVED brands.
	h.post(t, "/api/data-holders/dh-1/remove", `{"version": 3}`, http.StatusNoContent)

	brands = h.dataHolderBrands(t)
	require.Len(t, brands, 1)
	assert.Equal(t, "REMOVED", brands[0].Status)
}
