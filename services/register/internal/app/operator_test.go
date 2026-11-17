package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func seededWithProduct(t *testing.T) *memoryStore {
	t.Helper()

	p := participants()
	p.DataRecipients[0].Brands = []domain.DataRecipientBrand{
		{
			DataRecipientBrandID: "drb-1",
			BrandName:            "App",
			LogoURI:              "https://recipient.example/brand.png",
			Status:               domain.Active,
			SoftwareProducts: []domain.SoftwareProduct{
				{
					SoftwareProductID:          "sp-1",
					SoftwareProductName:        "App",
					SoftwareProductDescription: "An app",
					LogoURI:                    "https://recipient.example/product.png",
					Status:                     domain.Active,
					ClientURI:                  "https://recipient.example",
					RedirectURIs:               []string{"https://recipient.example/callback"},
					JWKSURI:                    "https://recipient.example/jwks",
					RevocationURI:              "https://recipient.example/revoke",
					RecipientBaseURI:           "https://cdr.recipient.example",
					Scopes:                     []string{"openid", "profile"},
				},
			},
		},
	}
	store := &memoryStore{}
	require.NoError(t, app.Seed(t.Context(), store, p))

	return store
}

func TestOperatorAppendsTheEventsOfTheCommandAndTheCascade(t *testing.T) {
	store := seededWithProduct(t)

	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	require.Len(t, store.appends, 3)
	a := store.appends[2]
	assert.Equal(t, "data-recipient", a.StreamType)
	assert.Equal(t, "dr-1", a.StreamID)
	assert.Equal(t, int64(1), a.ExpectedVersion)
	require.Len(t, a.Events, 2)
	assert.Equal(t, "AccreditationSuspended", a.Events[0].EventType)
	assert.Equal(t, "SoftwareProductDeactivated", a.Events[1].EventType)
}

func TestEachStatusEventLoads(t *testing.T) {
	store := seededWithProduct(t)

	operator := app.NewOperator(store)
	for _, command := range []domain.DataRecipientCommand{
		domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-1"},
		domain.ActivateSoftwareProduct{SoftwareProductID: "sp-1"},
		domain.SuspendAccreditation{},
		domain.ReinstateAccreditation{},
		domain.RevokeAccreditation{},
		domain.ReinstateAccreditation{},
		domain.SurrenderAccreditation{},
	} {
		require.NoError(t, operator.Execute(t.Context(), "dr-1", command), "%T", command)
	}

	state, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")

	require.NoError(t, err)
	assert.Equal(t, domain.DataRecipientSurrendered, state.DataRecipient.Status)
	assert.Equal(t, domain.Removed, state.DataRecipient.Brands[0].SoftwareProducts[0].Status)
}

func TestOperatorExecutesAtTheVersionThatTheOperatorSaw(t *testing.T) {
	store := seededWithProduct(t)
	operator := app.NewOperator(store)

	require.NoError(t, operator.ExecuteAt(t.Context(), "dr-1", 1, domain.SuspendAccreditation{}))

	err := operator.ExecuteAt(t.Context(), "dr-1", 1, domain.ReinstateAccreditation{})

	require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
	require.Len(t, store.appends, 3)
}

func TestOperatorExecuteAtGivesNotFoundForAnUnknownDataRecipient(t *testing.T) {
	err := app.NewOperator(seededWithProduct(t)).ExecuteAt(t.Context(), "dr-unknown", 1, domain.SuspendAccreditation{})

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOperatorExecutesADataHolderCommandAtTheVersionThatTheOperatorSaw(t *testing.T) {
	store := &memoryStore{}
	require.NoError(t, app.Seed(t.Context(), store, participants()))
	operator := app.NewOperator(store)
	dh := participants().DataHolders[0]
	change := domain.ChangeDataHolderDetails{
		Details: domain.DataHolderDetails{LegalEntityName: "Changed", LogoURI: dh.LogoURI},
	}

	require.NoError(t, operator.ExecuteDataHolderAt(t.Context(), dh.LegalEntityID, 1, change))

	err := operator.ExecuteDataHolderAt(t.Context(), dh.LegalEntityID, 1, change)

	require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
	_, version, err := app.DataHolderType.Load(t.Context(), store, dh.LegalEntityID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), version)
}

func TestARemovedClientCertificateLoadsWithNoThumbprint(t *testing.T) {
	const thumbprint = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	store := seededWithProduct(t)
	operator := app.NewOperator(store)
	require.NoError(t, operator.Execute(t.Context(), "dr-1",
		domain.RegisterClientCertificate{SoftwareProductID: "sp-1", Thumbprint: thumbprint}))

	require.NoError(t, operator.Execute(t.Context(), "dr-1",
		domain.RemoveClientCertificate{SoftwareProductID: "sp-1", Thumbprint: thumbprint}))

	require.Len(t, store.appends, 4)
	assert.Equal(t, "ClientCertificateRemoved", store.appends[3].Events[0].EventType)
	state, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")
	require.NoError(t, err)
	assert.Empty(t, state.DataRecipient.Brands[0].SoftwareProducts[0].ClientCertificates)
}

func TestEachDataHolderEventLoads(t *testing.T) {
	store := &memoryStore{}
	operator := app.NewOperator(store)
	brand := domain.DataHolderBrand{
		InterimID:      "dhb-2",
		BrandName:      "Second",
		Industries:     []domain.Industry{domain.Energy},
		LogoURI:        "https://holder.example/brand.png",
		Status:         domain.Active,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://holder.example"},
	}
	first := brand
	first.InterimID, first.BrandName = "dhb-1", "First"
	dh := participants().DataHolders[0]
	dh.Brands = []domain.DataHolderBrand{first}

	for _, command := range []domain.DataHolderCommand{
		domain.CreateDataHolder{DataHolder: dh},
		domain.ChangeDataHolderDetails{
			Details: domain.DataHolderDetails{LegalEntityName: "Changed", LogoURI: dh.LogoURI},
		},
		domain.AddDataHolderBrand{Brand: brand},
		domain.DeactivateDataHolderBrand{BrandID: "dhb-2"},
		domain.ActivateDataHolderBrand{BrandID: "dhb-2"},
		domain.ChangeDataHolderBrandDetails{BrandID: "dhb-2", Details: domain.DataHolderBrandDetails{
			BrandName: "Renamed", Industries: brand.Industries, LogoURI: brand.LogoURI,
			EndpointDetail: brand.EndpointDetail,
		}},
		domain.RemoveDataHolderBrand{BrandID: "dhb-2"},
		domain.RemoveDataHolder{},
	} {
		require.NoError(t, operator.ExecuteDataHolder(t.Context(), dh.LegalEntityID, command), "%T", command)
	}

	state, version, err := app.DataHolderType.Load(t.Context(), store, dh.LegalEntityID)

	require.NoError(t, err)
	assert.Equal(t, int64(9), version, "the removal of the data holder also removes its first brand")
	assert.Equal(t, "Changed", state.DataHolder.LegalEntityName)
	assert.Equal(t, domain.DataHolderRemoved, state.DataHolder.Status)
	assert.Equal(t, "Renamed", state.DataHolder.Brands[1].BrandName)

	for _, b := range state.DataHolder.Brands {
		assert.Equal(t, domain.Removed, b.Status)
	}
}

func TestEachDataRecipientChangeEventLoads(t *testing.T) {
	store := seededWithProduct(t)
	operator := app.NewOperator(store)
	state, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")
	require.NoError(t, err)

	product := state.DataRecipient.Brands[0].SoftwareProducts[0]
	product.SoftwareProductID = "sp-2"
	brand := state.DataRecipient.Brands[0]
	brand.DataRecipientBrandID, brand.SoftwareProducts = "drb-2", nil

	for _, command := range []domain.DataRecipientCommand{
		domain.ChangeDataRecipientDetails{Details: domain.DataRecipientDetails{
			LegalEntityName: "Changed", AccreditationNumber: "ADR2", AccreditationLevel: domain.Unrestricted,
			LogoURI: "https://recipient.example/logo.png",
		}},
		domain.AddDataRecipientBrand{Brand: brand},
		domain.ChangeDataRecipientBrandDetails{BrandID: "drb-2", Details: domain.DataRecipientBrandDetails{
			BrandName: "Pay", LogoURI: brand.LogoURI,
		}},
		domain.AddSoftwareProduct{BrandID: "drb-2", SoftwareProduct: product},
		domain.ChangeSoftwareProductDetails{SoftwareProductID: "sp-2", Details: domain.SoftwareProductDetails{
			SoftwareProductName: "Pay Web", SoftwareProductDescription: "Pay", LogoURI: product.LogoURI,
			ClientURI: product.ClientURI, RedirectURIs: product.RedirectURIs, JWKSURI: product.JWKSURI,
			RevocationURI: product.RevocationURI, RecipientBaseURI: product.RecipientBaseURI, Scopes: product.Scopes,
		}},
		domain.DeactivateDataRecipientBrand{BrandID: "drb-2"},
		domain.ActivateDataRecipientBrand{BrandID: "drb-2"},
		domain.RemoveDataRecipientBrand{BrandID: "drb-2"},
	} {
		require.NoError(t, operator.Execute(t.Context(), "dr-1", command), "%T", command)
	}

	state, _, err = app.DataRecipientType.Load(t.Context(), store, "dr-1")

	require.NoError(t, err)
	assert.Equal(t, "Changed", state.DataRecipient.LegalEntityName)
	assert.Equal(t, "Pay", state.DataRecipient.Brands[1].BrandName)
	assert.Equal(t, domain.Removed, state.DataRecipient.Brands[1].Status)
	assert.Equal(t, "Pay Web", state.DataRecipient.Brands[1].SoftwareProducts[0].SoftwareProductName)
	assert.Equal(t, domain.Removed, state.DataRecipient.Brands[1].SoftwareProducts[0].Status)
}
