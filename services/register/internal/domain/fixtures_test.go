package domain_test

import "github.com/lukemundy/cdr-lab/services/register/internal/domain"

func validDataHolder() domain.DataHolder {
	return domain.DataHolder{
		LegalEntityID:   "dh-1",
		LegalEntityName: "Holder Pty Ltd",
		LogoURI:         "https://holder.example/logo.png",
		Status:          domain.DataHolderActive,
		Brands: []domain.DataHolderBrand{{
			DataHolderBrandID: "dhb-1",
			BrandName:         "Holder Bank",
			Industries:        []domain.Industry{domain.Banking},
			LogoURI:           "https://holder.example/brand.png",
			Status:            domain.Active,
			EndpointDetail: domain.EndpointDetail{
				Version:         "v1",
				PublicBaseURI:   "https://api.holder.example",
				ResourceBaseURI: "https://mtls.holder.example",
				InfosecBaseURI:  "https://id.holder.example",
				WebsiteURI:      "https://holder.example",
			},
			JWKSEndpoint: "https://holder.example/jwks",
		}},
	}
}

func validDataRecipient() domain.DataRecipient {
	return domain.DataRecipient{
		LegalEntityID:       "dr-1",
		LegalEntityName:     "Recipient Pty Ltd",
		AccreditationNumber: "ADR000001",
		AccreditationLevel:  domain.Unrestricted,
		LogoURI:             "https://recipient.example/logo.png",
		Status:              domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "drb-1",
			BrandName:            "Recipient App",
			LogoURI:              "https://recipient.example/brand.png",
			Status:               domain.Active,
			SoftwareProducts: []domain.SoftwareProduct{{
				SoftwareProductID:          "sp-1",
				SoftwareProductName:        "Recipient App Web",
				SoftwareProductDescription: "A web app",
				LogoURI:                    "https://recipient.example/product.png",
				Status:                     domain.Active,
				ClientURI:                  "https://recipient.example",
				RedirectURIs:               []string{"https://recipient.example/callback"},
				JWKSURI:                    "https://recipient.example/jwks",
				RevocationURI:              "https://recipient.example/revoke",
				RecipientBaseURI:           "https://cdr.recipient.example",
				Scopes:                     []string{"openid", "profile", "bank:accounts.basic:read"},
			}},
		}},
	}
}
