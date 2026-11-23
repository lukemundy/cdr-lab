package app_test

import (
	"flag"
	"testing"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventsourcingtest"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var update = flag.Bool("update", false, "write the event golden file in testdata")

// Verifies: RLS-0080, RLS-0081, RLS-0082, RLS-0083, RLS-0084, RLS-0085, RLS-0086, RLS-0087, RLS-0088
func TestTheEncodedSampleOfEachEventTypeIsTheEventGoldenFile(t *testing.T) {
	golden := eventsourcingtest.NewGolden("testdata/events.golden.json")
	eventsourcingtest.Add(golden, app.DataHolderType, dataHolderSamples()...)
	eventsourcingtest.Add(golden, app.DataRecipientType, dataRecipientSamples()...)
	golden.Check(t, *update)
}

func dataHolderSamples() []domain.DataHolderEvent {
	holder := domain.DataHolder{
		LegalEntityID:   "le-holder",
		LegalEntityName: "Sample Holder Pty Ltd",
		LogoURI:         "https://holder.example.com/logo.png",
		Status:          domain.DataHolderActive,
		Brands:          []domain.DataHolderBrand{sampleDataHolderBrand()},
	}

	return []domain.DataHolderEvent{
		domain.DataHolderImported{DataHolder: holder},
		domain.DataHolderCreated{DataHolder: holder},
		domain.DataHolderDetailsChanged{
			LegalEntityName: "Sample Holder Group Pty Ltd",
			LogoURI:         "https://holder.example.com/logo-2.png",
		},
		domain.DataHolderLegalEntityRemoved{},
		domain.DataHolderBrandAdded{Brand: sampleDataHolderBrand()},
		domain.DataHolderBrandDetailsChanged{
			BrandID:        "dhb-1",
			BrandName:      "Sample Bank Online",
			BrandGroup:     "Sample Group",
			Industries:     []domain.Industry{domain.Banking},
			LogoURI:        "https://bank.example.com/logo-2.png",
			EndpointDetail: sampleEndpointDetail(),
			JWKSEndpoint:   "https://bank.example.com/jwks",
		},
		domain.DataHolderBrandActivated{BrandID: "dhb-1"},
		domain.DataHolderBrandDeactivated{BrandID: "dhb-1"},
		domain.DataHolderBrandRemoved{BrandID: "dhb-1"},
		domain.ServerCertificateIssued{BrandID: "dhb-1", Thumbprint: "sample-server-thumbprint"},
	}
}

func sampleDataHolderBrand() domain.DataHolderBrand {
	return domain.DataHolderBrand{
		DataHolderBrandID: "dhb-1",
		InterimID:         "interim-1",
		BrandName:         "Sample Bank",
		BrandGroup:        "Sample Group",
		Industries:        []domain.Industry{domain.Banking},
		LogoURI:           "https://bank.example.com/logo.png",
		Status:            domain.Active,
		EndpointDetail:    sampleEndpointDetail(),
		JWKSEndpoint:      "https://bank.example.com/jwks",
	}
}

func sampleEndpointDetail() domain.EndpointDetail {
	return domain.EndpointDetail{
		Version:          "1",
		PublicBaseURI:    "https://bank.example.com",
		ProductBaseURI:   "https://products.bank.example.com",
		ResourceBaseURI:  "https://api.bank.example.com",
		InfosecBaseURI:   "https://auth.bank.example.com",
		ExtensionBaseURI: "https://extension.bank.example.com",
		WebsiteURI:       "https://www.bank.example.com",
	}
}

func dataRecipientSamples() []domain.DataRecipientEvent {
	recipient := domain.DataRecipient{
		LegalEntityID:       "le-recipient",
		LegalEntityName:     "Sample Recipient Pty Ltd",
		AccreditationNumber: "ADR-000001",
		AccreditationLevel:  domain.Unrestricted,
		LogoURI:             "https://recipient.example.com/logo.png",
		Status:              domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "drb-1",
			BrandName:            "Sample App",
			LogoURI:              "https://app.example.com/logo.png",
			Status:               domain.Active,
			SoftwareProducts:     []domain.SoftwareProduct{sampleSoftwareProduct()},
		}},
	}

	return []domain.DataRecipientEvent{
		domain.DataRecipientImported{DataRecipient: recipient},
		domain.DataRecipientCreated{DataRecipient: recipient},
		domain.DataRecipientDetailsChanged{
			LegalEntityName:     "Sample Recipient Group Pty Ltd",
			AccreditationNumber: "ADR-000002",
			AccreditationLevel:  domain.Sponsored,
			LogoURI:             "https://recipient.example.com/logo-2.png",
		},
		domain.AccreditationSuspended{},
		domain.AccreditationRevoked{},
		domain.AccreditationSurrendered{},
		domain.AccreditationReinstated{},
		domain.DataRecipientBrandAdded{Brand: recipient.Brands[0]},
		domain.DataRecipientBrandDetailsChanged{
			BrandID:   "drb-1",
			BrandName: "Sample App Plus",
			LogoURI:   "https://app.example.com/logo-2.png",
		},
		domain.DataRecipientBrandActivated{BrandID: "drb-1"},
		domain.DataRecipientBrandDeactivated{BrandID: "drb-1"},
		domain.DataRecipientBrandRemoved{BrandID: "drb-1"},
		domain.SoftwareProductAdded{BrandID: "drb-1", SoftwareProduct: sampleSoftwareProduct()},
		domain.SoftwareProductDetailsChanged{
			SoftwareProductID:          "sp-1",
			SoftwareProductName:        "Sample Budget",
			SoftwareProductDescription: "A sample budget product.",
			LogoURI:                    "https://app.example.com/budget-2.png",
			ClientURI:                  "https://app.example.com",
			SectorIdentifierURI:        "https://app.example.com/sector",
			RedirectURIs:               []string{"https://app.example.com/callback"},
			TOSURI:                     "https://app.example.com/terms",
			PolicyURI:                  "https://app.example.com/policy",
			JWKSURI:                    "https://app.example.com/jwks",
			RevocationURI:              "https://app.example.com/revoke",
			RecipientBaseURI:           "https://api.app.example.com",
			Scopes:                     []string{"openid"},
		},
		domain.SoftwareProductActivated{SoftwareProductID: "sp-1"},
		domain.SoftwareProductDeactivated{SoftwareProductID: "sp-1"},
		domain.SoftwareProductRemoved{SoftwareProductID: "sp-1"},
		domain.ClientCertificateRegistered{SoftwareProductID: "sp-1", Thumbprint: "sample-client-thumbprint"},
		domain.ClientCertificateRemoved{SoftwareProductID: "sp-1", Thumbprint: "sample-client-thumbprint"},
		domain.ClientCertificateIssued{SoftwareProductID: "sp-1", Thumbprint: "sample-client-thumbprint"},
	}
}

func sampleSoftwareProduct() domain.SoftwareProduct {
	return domain.SoftwareProduct{
		SoftwareProductID:          "sp-1",
		SoftwareProductName:        "Sample Budget",
		SoftwareProductDescription: "A sample budget product.",
		LogoURI:                    "https://app.example.com/budget.png",
		Status:                     domain.Active,
		ClientURI:                  "https://app.example.com",
		SectorIdentifierURI:        "https://app.example.com/sector",
		RedirectURIs:               []string{"https://app.example.com/callback"},
		TOSURI:                     "https://app.example.com/terms",
		PolicyURI:                  "https://app.example.com/policy",
		JWKSURI:                    "https://app.example.com/jwks",
		RevocationURI:              "https://app.example.com/revoke",
		RecipientBaseURI:           "https://api.app.example.com",
		Scopes:                     []string{"openid", "bank:accounts.basic:read"},
		ClientCertificates:         []string{"sample-client-thumbprint"},
	}
}
