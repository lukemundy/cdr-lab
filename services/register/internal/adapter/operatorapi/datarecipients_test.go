package operatorapi_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const softwareProductJSON = `{
	"softwareProductId": "sp-1", "softwareProductName": "Web", "softwareProductDescription": "A web app",
	"logoUri": "https://recipient.example/sp.png", "status": "ACTIVE", "clientUri": "https://recipient.example",
	"redirectUris": ["https://recipient.example/cb"], "jwksUri": "https://recipient.example/jwks",
	"revocationUri": "https://recipient.example/revoke", "recipientBaseUri": "https://cdr.recipient.example",
	"scopes": ["openid"]
}`

const recipientBrandJSON = `{
	"dataRecipientBrandId": "drb-1", "brandName": "App", "logoUri": "https://recipient.example/b.png",
	"status": "ACTIVE", "softwareProducts": [` + softwareProductJSON + `]
}`

func softwareProduct() domain.SoftwareProduct {
	return domain.SoftwareProduct{
		SoftwareProductID: "sp-1", SoftwareProductName: "Web", SoftwareProductDescription: "A web app",
		LogoURI: "https://recipient.example/sp.png", Status: domain.Active, ClientURI: "https://recipient.example",
		RedirectURIs: []string{"https://recipient.example/cb"}, JWKSURI: "https://recipient.example/jwks",
		RevocationURI: "https://recipient.example/revoke", RecipientBaseURI: "https://cdr.recipient.example",
		Scopes: []string{"openid"},
	}
}

func recipientBrand() domain.DataRecipientBrand {
	return domain.DataRecipientBrand{
		DataRecipientBrandID: "drb-1", BrandName: "App", LogoURI: "https://recipient.example/b.png",
		Status: domain.Active, SoftwareProducts: []domain.SoftwareProduct{softwareProduct()},
	}
}

// Verifies: REG-0809, REG-0811, REG-0815, REG-0816, REG-0867, REG-0869, REG-0872
func TestEachCreateAndAddOfADataRecipientExecutesItsCommandAndGivesTheLocation(t *testing.T) {
	tests := map[string]struct {
		path, body, location string
		want                 domain.DataRecipientCommand
	}{
		"create a data recipient": {
			"/v1/data-recipients",
			`{"legalEntityId": "dr-1", "legalEntityName": "Recipient", "accreditationNumber": "ADR1",
			  "accreditationLevel": "SPONSORED", "logoUri": "https://recipient.example/logo.png", "status": "ACTIVE",
			  "dataRecipientBrands": [` + recipientBrandJSON + `]}`,
			"/v1/data-recipients/dr-1",
			domain.CreateDataRecipient{DataRecipient: domain.DataRecipient{
				LegalEntityID: "dr-1", LegalEntityName: "Recipient", AccreditationNumber: "ADR1",
				AccreditationLevel: domain.Sponsored, LogoURI: "https://recipient.example/logo.png",
				Status: domain.DataRecipientActive, Brands: []domain.DataRecipientBrand{recipientBrand()},
			}},
		},
		"add a brand": {
			"/v1/data-recipients/dr-1/brands", recipientBrandJSON, "/v1/data-recipients/dr-1/brands/drb-1",
			domain.AddDataRecipientBrand{Brand: recipientBrand()},
		},
		"add a software product": {
			"/v1/data-recipients/dr-1/brands/drb-1/software-products", softwareProductJSON,
			"/v1/data-recipients/dr-1/software-products/sp-1",
			domain.AddSoftwareProduct{BrandID: "drb-1", SoftwareProduct: softwareProduct()},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			operator := &fakeOperator{}

			rec := serveBody(t, operator, tt.path, []byte(tt.body))

			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, tt.location, rec.Header().Get("Location"))
			assert.Equal(t, "dr-1", operator.legalEntityID)
			assert.Equal(t, tt.want, operator.command)
		})
	}
}

// Verifies: REG-0817, REG-0868, REG-0870, REG-0871, REG-0873
func TestEachDataRecipientChangeExecutesItsCommand(t *testing.T) {
	const brand = "/v1/data-recipients/dr-1/brands/drb-1"

	tests := map[string]struct {
		method, path, body string
		want               domain.DataRecipientCommand
	}{
		"change the details": {
			http.MethodPut, "/v1/data-recipients/dr-1",
			`{"legalEntityName": "R", "accreditationNumber": "ADR1", "accreditationLevel": "UNRESTRICTED",
			  "logoUri": "https://recipient.example/logo.png"}`,
			domain.ChangeDataRecipientDetails{Details: domain.DataRecipientDetails{
				LegalEntityName: "R", AccreditationNumber: "ADR1", AccreditationLevel: domain.Unrestricted,
				LogoURI: "https://recipient.example/logo.png",
			}},
		},
		"change the details of a brand": {
			http.MethodPut, brand, `{"brandName": "Pay", "logoUri": "https://recipient.example/pay.png"}`,
			domain.ChangeDataRecipientBrandDetails{BrandID: "drb-1", Details: domain.DataRecipientBrandDetails{
				BrandName: "Pay", LogoURI: "https://recipient.example/pay.png",
			}},
		},
		"activate a brand": {
			http.MethodPost, brand + "/activate", "", domain.ActivateDataRecipientBrand{BrandID: "drb-1"},
		},
		"deactivate a brand": {
			http.MethodPost, brand + "/deactivate", "", domain.DeactivateDataRecipientBrand{BrandID: "drb-1"},
		},
		"remove a brand": {http.MethodPost, brand + "/remove", "", domain.RemoveDataRecipientBrand{BrandID: "drb-1"}},
		"change the details of a software product": {
			http.MethodPut, "/v1/data-recipients/dr-1/software-products/sp-1",
			`{"softwareProductName": "Web", "softwareProductDescription": "A web app",
			  "logoUri": "https://recipient.example/sp.png", "clientUri": "https://recipient.example",
			  "redirectUris": ["https://recipient.example/cb"], "jwksUri": "https://recipient.example/jwks",
			  "revocationUri": "https://recipient.example/revoke", "recipientBaseUri": "https://cdr.recipient.example",
			  "scopes": ["openid"], "tosUri": "https://recipient.example/tos"}`,
			domain.ChangeSoftwareProductDetails{SoftwareProductID: "sp-1", Details: domain.SoftwareProductDetails{
				SoftwareProductName: "Web", SoftwareProductDescription: "A web app",
				LogoURI: "https://recipient.example/sp.png", ClientURI: "https://recipient.example",
				RedirectURIs: []string{"https://recipient.example/cb"}, JWKSURI: "https://recipient.example/jwks",
				RevocationURI: "https://recipient.example/revoke", RecipientBaseURI: "https://cdr.recipient.example",
				Scopes: []string{"openid"}, TOSURI: "https://recipient.example/tos",
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			operator := &fakeOperator{}

			rec := serveMethodBody(t, operator, tt.method, tt.path, []byte(tt.body))

			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			assert.Equal(t, "dr-1", operator.legalEntityID)
			assert.Equal(t, tt.want, operator.command)
		})
	}
}

// Verifies: REG-0818
func TestAChangeOfASoftwareProductRefusesItsStatusAndCertificates(t *testing.T) {
	for _, field := range []string{`"status": "ACTIVE"`, `"clientCertificates": []`, `"softwareProductId": "sp-2"`} {
		t.Run(field, func(t *testing.T) {
			operator := &fakeOperator{}

			rec := serveMethodBody(t, operator, http.MethodPut, "/v1/data-recipients/dr-1/software-products/sp-1",
				[]byte(`{"softwareProductName": "Web", `+field+`}`))

			requireProblem(t, rec, http.StatusBadRequest)
			assert.Nil(t, operator.command)
		})
	}
}
