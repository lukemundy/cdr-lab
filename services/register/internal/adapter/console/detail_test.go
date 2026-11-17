package console_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func (*queryRecorder) DataRecipient(context.Context, string) (app.DataRecipientItem, error) {
	return app.DataRecipientItem{}, errors.New("not used")
}

func (*queryRecorder) DataHolder(context.Context, string) (app.DataHolderItem, error) {
	return app.DataHolderItem{}, errors.New("not used")
}

func (failingReader) DataRecipient(context.Context, string) (app.DataRecipientItem, error) {
	return app.DataRecipientItem{}, errors.New("a message that contains Mock Budgeting")
}

func (failingReader) DataHolder(context.Context, string) (app.DataHolderItem, error) {
	return app.DataHolderItem{}, errors.New("a message that contains Mock Holdings")
}

// withoutLastUpdated checks that lastUpdated is a UTC time near now, and gives the body with no lastUpdated.
func withoutLastUpdated(t *testing.T, body string) string {
	t.Helper()

	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &detail))

	lastUpdated, ok := detail["lastUpdated"].(string)
	require.True(t, ok, "lastUpdated is a string")
	assert.Regexp(t, `Z$`, lastUpdated, "lastUpdated is in UTC")

	at, err := time.Parse(time.RFC3339Nano, lastUpdated)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), at, time.Minute)

	delete(detail, "lastUpdated")

	rest, err := json.Marshal(detail)
	require.NoError(t, err)

	return string(rest)
}

const (
	thumbprint1 = "fj5bZBu5UoSrAua59pRyfSczfxvH2Jo9mn-XiLUWZ_M"
	thumbprint2 = "K1mHUV9VotBbECiNHlOgxTqXzkRHAR0KmgmBU-ggd_E"
)

// Verifies: REG-1127, REG-1128
func TestADataRecipientDetailGivesEachBrandAndSoftwareProductInTheOrderOfTheRecord(t *testing.T) {
	full := product("sp-2", domain.Inactive)
	full.SectorIdentifierURI = "https://recipient.example/sector"
	full.TOSURI = "https://recipient.example/terms"
	full.PolicyURI = "https://recipient.example/policy"
	full.RedirectURIs = []string{"https://recipient.example/b", "https://recipient.example/a"}
	full.Scopes = []string{"openid", "profile", "bank:accounts.basic:read"}
	full.ClientCertificates = []string{thumbprint2, thumbprint1}

	dr := recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
		[]domain.SoftwareProduct{product("sp-1", domain.Active), full},
		[]domain.SoftwareProduct{})
	dr.AccreditationLevel = domain.Sponsored
	dr.Brands[1].Status = domain.Removed
	queries := seed(t, dr, recipient("dr-2", "Beta Pty Ltd", domain.DataRecipientActive))

	resp := get(t, queries, "/api/data-recipients/dr-1")

	assert.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
	assert.JSONEq(t, `{
		"legalEntityId": "dr-1", "legalEntityName": "Alpha Pty Ltd", "accreditationNumber": "ADR-dr-1",
		"accreditationLevel": "SPONSORED", "logoUri": "https://recipient.example/logo.png", "status": "ACTIVE",
		"dataRecipientBrands": [
			{"dataRecipientBrandId": "dr-1-b1", "brandName": "Brand", "logoUri": "https://recipient.example/brand.png",
			 "status": "ACTIVE", "softwareProducts": [
				{"softwareProductId": "sp-1", "softwareProductName": "Product sp-1",
				 "softwareProductDescription": "A product", "logoUri": "https://recipient.example/product.png",
				 "status": "ACTIVE", "clientUri": "https://recipient.example",
				 "redirectUris": ["https://recipient.example/callback"], "jwksUri": "https://recipient.example/jwks",
				 "revocationUri": "https://recipient.example/revoke",
				 "recipientBaseUri": "https://cdr.recipient.example", "scopes": ["openid", "profile"],
				 "clientCertificateThumbprints": []},
				{"softwareProductId": "sp-2", "softwareProductName": "Product sp-2",
				 "softwareProductDescription": "A product", "logoUri": "https://recipient.example/product.png",
				 "status": "INACTIVE", "clientUri": "https://recipient.example",
				 "sectorIdentifierUri": "https://recipient.example/sector",
				 "redirectUris": ["https://recipient.example/b", "https://recipient.example/a"],
				 "tosUri": "https://recipient.example/terms", "policyUri": "https://recipient.example/policy",
				 "jwksUri": "https://recipient.example/jwks", "revocationUri": "https://recipient.example/revoke",
				 "recipientBaseUri": "https://cdr.recipient.example",
				 "scopes": ["openid", "profile", "bank:accounts.basic:read"],
				 "clientCertificateThumbprints": ["`+thumbprint2+`", "`+thumbprint1+`"]}
			]},
			{"dataRecipientBrandId": "dr-1-b2", "brandName": "Brand", "logoUri": "https://recipient.example/brand.png",
			 "status": "REMOVED", "softwareProducts": []}
		]
	}`, withoutActions(t, withoutLastUpdated(t, resp.body)))
}

// Verifies: REG-1130
func TestADataRecipientDetailWithNoBrandsHasAnEmptyList(t *testing.T) {
	queries := seed(t, recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive))

	resp := get(t, queries, "/api/data-recipients/dr-1")

	assert.Equal(t, http.StatusOK, resp.status)
	assert.Contains(t, resp.body, `"dataRecipientBrands":[]`)
}

// Verifies: REG-1127, REG-1128
func TestADataHolderDetailGivesEachBrandInTheOrderOfTheRecord(t *testing.T) {
	registered := brand("dh-1-b1", "Alpha Bank", domain.Active, domain.Telco, domain.Banking)
	registered.InterimID = "dh-1-i1"
	registered.BrandGroup = "Alpha Group"
	registered.EndpointDetail.ProductBaseURI = "https://api.holder.example/products"
	registered.EndpointDetail.ExtensionBaseURI = "https://ext.holder.example"

	// A brand that shares only product data has only its public details.
	interim := domain.DataHolderBrand{
		InterimID: "dh-1-i2", BrandName: "Alpha Mutual", Industries: []domain.Industry{domain.Banking},
		LogoURI: "https://holder.example/mutual.png", Status: domain.Inactive,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://api.mutual.example"},
	}

	queries := seedParticipants(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive, registered, interim),
		dataHolder("dh-2", "Beta Ltd", domain.DataHolderRemoved),
	}})

	resp := get(t, queries, "/api/data-holders/dh-1")

	assert.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
	assert.JSONEq(t, `{
		"legalEntityId": "dh-1", "legalEntityName": "Alpha Ltd", "logoUri": "https://holder.example/logo.png",
		"status": "ACTIVE", "version": 1,
		"dataHolderBrands": [
			{"dataHolderBrandId": "dh-1-b1", "interimId": "dh-1-i1", "brandName": "Alpha Bank",
			 "brandGroup": "Alpha Group", "industries": ["telco", "banking"],
			 "logoUri": "https://holder.example/brand.png", "status": "ACTIVE",
			 "endpointDetail": {"version": "1", "publicBaseUri": "https://api.holder.example",
			  "productBaseUri": "https://api.holder.example/products",
			  "resourceBaseUri": "https://mtls.holder.example", "infosecBaseUri": "https://id.holder.example",
			  "extensionBaseUri": "https://ext.holder.example", "websiteUri": "https://holder.example"},
			 "jwksEndpoint": "https://holder.example/jwks"},
			{"interimId": "dh-1-i2", "brandName": "Alpha Mutual", "industries": ["banking"],
			 "logoUri": "https://holder.example/mutual.png", "status": "INACTIVE",
			 "endpointDetail": {"publicBaseUri": "https://api.mutual.example"}}
		]
	}`, withoutHolderActions(t, withoutLastUpdated(t, resp.body)))
}

// Verifies: REG-1130
func TestADataHolderDetailWithNoBrandsHasAnEmptyList(t *testing.T) {
	queries := seedParticipants(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive),
	}})

	resp := get(t, queries, "/api/data-holders/dh-1")

	assert.Equal(t, http.StatusOK, resp.status)
	assert.Contains(t, resp.body, `"dataHolderBrands":[]`)
}

// Verifies: REG-1111
func TestADetailOfAnUnknownLegalEntityGivesNotFoundWithNoIdentifier(t *testing.T) {
	queries := seedParticipants(t, domain.Participants{
		DataRecipients: []domain.DataRecipient{recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive)},
		DataHolders:    []domain.DataHolder{dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive)},
	})

	tests := map[string]struct{ path, detail string }{
		"unknown data recipient": {
			"/api/data-recipients/unknown-id", "The Register has no data recipient with this legal entity ID.",
		},
		"data holder as a data recipient": {
			"/api/data-recipients/dh-1", "The Register has no data recipient with this legal entity ID.",
		},
		"unknown data holder": {
			"/api/data-holders/unknown-id", "The Register has no data holder with this legal entity ID.",
		},
		"data recipient as a data holder": {
			"/api/data-holders/dr-1", "The Register has no data holder with this legal entity ID.",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := get(t, queries, tt.path)

			assert.Equal(t, http.StatusNotFound, resp.status)
			assert.Equal(t, "application/problem+json", resp.header.Get("Content-Type"))
			assert.JSONEq(t, `{"type": "about:blank", "title": "Not Found", "status": 404, "detail": "`+tt.detail+`"}`,
				resp.body)
		})
	}
}

func TestADetailGivesAProblemWithNoErrorMessageWhenTheReadFails(t *testing.T) {
	for _, path := range []string{"/api/data-recipients/dr-1", "/api/data-holders/dh-1"} {
		resp := get(t, failingReader{}, path)

		assert.Equal(t, http.StatusInternalServerError, resp.status, path)
		assert.NotContains(t, resp.body, "Mock", path)
	}
}
