package readmodel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

func newStore(t *testing.T) (*eventstore.Store, *pgxpool.Pool) {
	t.Helper()
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return eventstore.New(pool, readmodel.Projections()...), pool
}

func product(id string, status domain.Status) domain.SoftwareProduct {
	return domain.SoftwareProduct{
		SoftwareProductID: id, SoftwareProductName: "Product " + id, SoftwareProductDescription: "A product",
		LogoURI: "https://recipient.example/product.png", Status: status, ClientURI: "https://recipient.example",
		RedirectURIs: []string{"https://recipient.example/callback"}, JWKSURI: "https://recipient.example/jwks",
		RevocationURI: "https://recipient.example/revoke", RecipientBaseURI: "https://cdr.recipient.example",
		Scopes: []string{"openid", "profile"},
	}
}

func recipient(id string, status domain.DataRecipientStatus, brands ...[]domain.SoftwareProduct) domain.DataRecipient {
	dr := domain.DataRecipient{
		LegalEntityID: id, LegalEntityName: "Recipient " + id, AccreditationNumber: "ADR-" + id,
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png", Status: status,
	}
	for i, products := range brands {
		dr.Brands = append(dr.Brands, domain.DataRecipientBrand{
			DataRecipientBrandID: fmt.Sprintf(
				"%s-b%d",
				id,
				i+1,
			),
			BrandName:        "Brand",
			LogoURI:          "https://recipient.example/brand.png",
			Status:           domain.Active,
			SoftwareProducts: products,
		})
	}

	return dr
}

func TestStatusesGiveEachDataRecipientAndSoftwareProductInTheOrderOfTheirIDs(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-2", domain.DataRecipientSuspended, []domain.SoftwareProduct{product("sp-3", domain.Inactive)}),
		recipient("dr-1", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-2", domain.Removed), product("sp-1", domain.Active)},
			[]domain.SoftwareProduct{product("sp-4", domain.Inactive)}),
		recipient("dr-3", domain.DataRecipientRevoked),
	}}))

	queries := readmodel.NewQueries(pool)

	recipients, err := queries.DataRecipientStatuses(t.Context())
	require.NoError(t, err)
	products, err := queries.SoftwareProductStatuses(t.Context())
	require.NoError(t, err)

	assert.Equal(t, []app.DataRecipientStatusItem{
		{LegalEntityID: "dr-1", Status: domain.DataRecipientActive},
		{LegalEntityID: "dr-2", Status: domain.DataRecipientSuspended},
		{LegalEntityID: "dr-3", Status: domain.DataRecipientRevoked},
	}, recipients)
	assert.Equal(t, []app.SoftwareProductStatusItem{
		{SoftwareProductID: "sp-1", Status: domain.Active},
		{SoftwareProductID: "sp-2", Status: domain.Removed},
		{SoftwareProductID: "sp-3", Status: domain.Inactive},
		{SoftwareProductID: "sp-4", Status: domain.Inactive},
	}, products)
}

func TestStatusesAreEmptyWithNoParticipants(t *testing.T) {
	_, pool := newStore(t)
	queries := readmodel.NewQueries(pool)

	recipients, err := queries.DataRecipientStatuses(t.Context())
	require.NoError(t, err)
	products, err := queries.SoftwareProductStatuses(t.Context())
	require.NoError(t, err)

	assert.Empty(t, recipients)
	assert.Empty(t, products)
}

func TestAProjectionReplacesTheRowsOfTheStreamFromItsState(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient(
			"dr-1",
			domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Active)},
		),
		recipient("dr-2", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-3", domain.Active)}),
	}}))

	// A second import event is not a valid command result. It gives a new
	// state for the stream with no other event types.
	changed := recipient(
		"dr-1",
		domain.DataRecipientSuspended,
		[]domain.SoftwareProduct{product("sp-1", domain.Inactive)},
	)
	data, err := app.DataRecipientType.Encode(
		[]domain.DataRecipientEvent{domain.DataRecipientImported{DataRecipient: changed}},
	)
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: "dr-1", ExpectedVersion: 1, Events: data,
	}))

	queries := readmodel.NewQueries(pool)
	recipients, err := queries.DataRecipientStatuses(t.Context())
	require.NoError(t, err)
	products, err := queries.SoftwareProductStatuses(t.Context())
	require.NoError(t, err)

	assert.Equal(t, []app.DataRecipientStatusItem{
		{LegalEntityID: "dr-1", Status: domain.DataRecipientSuspended},
		{LegalEntityID: "dr-2", Status: domain.DataRecipientActive},
	}, recipients)
	assert.Equal(t, []app.SoftwareProductStatusItem{
		{SoftwareProductID: "sp-1", Status: domain.Inactive},
		{SoftwareProductID: "sp-3", Status: domain.Active},
	}, products)
}

// The tests give each stream an ID that no other stream has.
func lastRecordedAt(t *testing.T, pool *pgxpool.Pool, streamID string) time.Time {
	t.Helper()

	var recordedAt time.Time
	require.NoError(t, pool.QueryRow(t.Context(),
		"SELECT max(recorded_at) FROM events WHERE stream_id = $1", streamID).Scan(&recordedAt))

	return recordedAt
}

func TestDataRecipientsGiveEachDataRecipientInTheOrderOfTheirIDs(t *testing.T) {
	store, pool := newStore(t)
	dr1 := recipient("dr-1", domain.DataRecipientActive,
		[]domain.SoftwareProduct{product("sp-1", domain.Active)},
		[]domain.SoftwareProduct{product("sp-2", domain.Inactive)})
	dr2 := recipient("dr-2", domain.DataRecipientRevoked)
	require.NoError(
		t,
		app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{dr2, dr1}}),
	)

	got, err := readmodel.NewQueries(pool).DataRecipients(t.Context())

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, dr1, got[0].DataRecipient)
	assert.True(t, got[0].LastUpdated.Equal(lastRecordedAt(t, pool, "dr-1")))
	assert.Equal(t, dr2, got[1].DataRecipient)
	assert.True(t, got[1].LastUpdated.Equal(lastRecordedAt(t, pool, "dr-2")))
}

func TestDataRecipientsAreEmptyWithNoParticipants(t *testing.T) {
	_, pool := newStore(t)

	got, err := readmodel.NewQueries(pool).DataRecipients(t.Context())

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestADataRecipientHasTheStateAndTimeOfItsLastEvent(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}}))
	first := lastRecordedAt(t, pool, "dr-1")

	// A second import event is not a valid command result. It gives a new
	// state for the stream with no other event types.
	changed := recipient(
		"dr-1",
		domain.DataRecipientSuspended,
		[]domain.SoftwareProduct{product("sp-1", domain.Inactive)},
	)
	data, err := app.DataRecipientType.Encode(
		[]domain.DataRecipientEvent{domain.DataRecipientImported{DataRecipient: changed}},
	)
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: "dr-1", ExpectedVersion: 1, Events: data,
	}))

	got, err := readmodel.NewQueries(pool).DataRecipients(t.Context())

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, changed, got[0].DataRecipient)
	assert.True(t, got[0].LastUpdated.After(first))
	assert.True(t, got[0].LastUpdated.Equal(lastRecordedAt(t, pool, "dr-1")))
}

func TestAnOperatorStatusChangeUpdatesTheReadModels(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}}))
	seeded := lastRecordedAt(t, pool, "dr-1")

	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	queries := readmodel.NewQueries(pool)
	recipients, err := queries.DataRecipientStatuses(t.Context())
	require.NoError(t, err)
	products, err := queries.SoftwareProductStatuses(t.Context())
	require.NoError(t, err)
	details, err := queries.DataRecipients(t.Context())
	require.NoError(t, err)

	assert.Equal(
		t,
		[]app.DataRecipientStatusItem{{LegalEntityID: "dr-1", Status: domain.DataRecipientSuspended}},
		recipients,
	)
	assert.Equal(t, []app.SoftwareProductStatusItem{{SoftwareProductID: "sp-1", Status: domain.Inactive}}, products)
	require.Len(t, details, 1)
	assert.Equal(t, domain.DataRecipientSuspended, details[0].DataRecipient.Status)
	assert.Equal(t, domain.Inactive, details[0].DataRecipient.Brands[0].SoftwareProducts[0].Status)
	assert.True(t, details[0].LastUpdated.After(seeded))
}

func holder(id string, status domain.DataHolderStatus, brands ...[]domain.Industry) domain.DataHolder {
	brandStatus := domain.Active
	if status == domain.DataHolderRemoved {
		brandStatus = domain.Removed
	}

	dh := domain.DataHolder{
		LegalEntityID:   id,
		LegalEntityName: "Holder " + id,
		LogoURI:         "https://holder.example/logo.png",
		Status:          status,
	}
	for i, industries := range brands {
		dh.Brands = append(dh.Brands, domain.DataHolderBrand{
			DataHolderBrandID: fmt.Sprintf("%s-b%d", id, i+1), BrandName: "Brand", Industries: industries,
			LogoURI: "https://holder.example/brand.png", Status: brandStatus,
			EndpointDetail: domain.EndpointDetail{
				Version:         "1",
				PublicBaseURI:   "https://api.holder.example",
				ResourceBaseURI: "https://mtls.holder.example",
				InfosecBaseURI:  "https://id.holder.example",
				WebsiteURI:      "https://holder.example",
			},
			JWKSEndpoint: "https://holder.example/jwks",
		})
	}

	return dh
}

func TestDataHolderStatusesGiveTheDataHoldersOfAnIndustryInTheOrderOfTheirIDs(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{
		holder("dh-3", domain.DataHolderRemoved, []domain.Industry{domain.Banking}),
		holder(
			"dh-1",
			domain.DataHolderActive,
			[]domain.Industry{domain.Banking},
			[]domain.Industry{domain.Energy, domain.Telco},
		),
		holder("dh-2", domain.DataHolderActive, []domain.Industry{domain.Energy}),
		holder("dh-4", domain.DataHolderActive),
	}}))

	queries := readmodel.NewQueries(pool)
	industry := func(i domain.Industry) *domain.Industry { return &i }

	tests := map[string]struct {
		industry *domain.Industry
		want     []app.DataHolderStatusItem
	}{
		"all": {nil, []app.DataHolderStatusItem{
			{LegalEntityID: "dh-1", Status: domain.DataHolderActive},
			{LegalEntityID: "dh-2", Status: domain.DataHolderActive},
			{LegalEntityID: "dh-3", Status: domain.DataHolderRemoved},
			{LegalEntityID: "dh-4", Status: domain.DataHolderActive},
		}},
		"banking": {industry(domain.Banking), []app.DataHolderStatusItem{
			{LegalEntityID: "dh-1", Status: domain.DataHolderActive},
			{LegalEntityID: "dh-3", Status: domain.DataHolderRemoved},
		}},
		"energy": {industry(domain.Energy), []app.DataHolderStatusItem{
			{LegalEntityID: "dh-1", Status: domain.DataHolderActive},
			{LegalEntityID: "dh-2", Status: domain.DataHolderActive},
		}},
		"telco": {industry(domain.Telco), []app.DataHolderStatusItem{
			{LegalEntityID: "dh-1", Status: domain.DataHolderActive},
		}},
		"non-bank-lending": {industry(domain.NonBankLending), []app.DataHolderStatusItem{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := queries.DataHolderStatuses(t.Context(), tt.industry)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestADataHolderProjectionReplacesTheRowOfTheStreamFromItsState(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{
		holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking}),
	}}))

	// A second import event is not a valid command result. It gives a new
	// state for the stream with no other event types.
	changed := holder("dh-1", domain.DataHolderRemoved, []domain.Industry{domain.Energy})
	data, err := app.DataHolderType.Encode([]domain.DataHolderEvent{domain.DataHolderImported{DataHolder: changed}})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataHolderType.StreamType, StreamID: "dh-1", ExpectedVersion: 1, Events: data,
	}))

	queries := readmodel.NewQueries(pool)
	banking := domain.Banking
	inBanking, err := queries.DataHolderStatuses(t.Context(), &banking)
	require.NoError(t, err)
	all, err := queries.DataHolderStatuses(t.Context(), nil)
	require.NoError(t, err)

	assert.Empty(t, inBanking)
	assert.Equal(t, []app.DataHolderStatusItem{{LegalEntityID: "dh-1", Status: domain.DataHolderRemoved}}, all)
}

func brandsOf(items []app.DataHolderBrandItem) []string {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.Brand.ID())
	}

	return ids
}

func TestDataHolderBrandsGiveTheBrandsOfAnIndustryThatAreNotRemovedInTheOrderOfTheirIDs(t *testing.T) {
	store, pool := newStore(t)
	active := holder(
		"dh-1",
		domain.DataHolderActive,
		[]domain.Industry{domain.Banking},
		[]domain.Industry{domain.Energy, domain.Telco},
	)
	active.Brands[1].Status = domain.Inactive
	interim := domain.DataHolderBrand{
		InterimID: "dh-1-a", BrandName: "Interim", Industries: []domain.Industry{domain.Energy},
		LogoURI: "https://holder.example/interim.png", Status: domain.Active,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://prd.holder.example"},
	}
	active.Brands = append(active.Brands, interim)
	removedBrand := holder("dh-2", domain.DataHolderActive, []domain.Industry{domain.Banking})
	removedBrand.Brands[0].Status = domain.Removed
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{
		holder("dh-3", domain.DataHolderRemoved, []domain.Industry{domain.Banking}),
		active,
		removedBrand,
	}}))

	queries := readmodel.NewQueries(pool)
	industry := func(i domain.Industry) *domain.Industry { return &i }

	tests := map[string]struct {
		industry *domain.Industry
		want     []string
	}{
		"all":              {nil, []string{"dh-1-a", "dh-1-b1", "dh-1-b2"}},
		"banking":          {industry(domain.Banking), []string{"dh-1-b1"}},
		"energy":           {industry(domain.Energy), []string{"dh-1-a", "dh-1-b2"}},
		"telco":            {industry(domain.Telco), []string{"dh-1-b2"}},
		"non-bank-lending": {industry(domain.NonBankLending), []string{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := queries.DataHolderBrands(t.Context(), tt.industry)

			require.NoError(t, err)
			assert.Equal(t, tt.want, brandsOf(got))
		})
	}
}

func TestADataHolderBrandHasItsFieldsAndTheTimeOfTheLastEventOfItsDataHolder(t *testing.T) {
	store, pool := newStore(t)
	dh := holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking})
	dh.Brands[0].InterimID = "dh-1-i"
	dh.Brands[0].BrandGroup = "Group"
	dh.Brands[0].EndpointDetail.ProductBaseURI = "https://prd.holder.example"
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{dh}}))

	got, err := readmodel.NewQueries(pool).DataHolderBrands(t.Context(), nil)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, dh.Brands[0], got[0].Brand)

	var recordedAt time.Time
	require.NoError(
		t,
		pool.QueryRow(t.Context(), "SELECT max(recorded_at) FROM events WHERE stream_id = 'dh-1'").Scan(&recordedAt),
	)
	assert.True(t, got[0].LastUpdated.Equal(recordedAt))
}

func TestADataHolderProjectionReplacesTheBrandsOfTheStream(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{
		holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking}, []domain.Industry{domain.Energy}),
	}}))

	// A second import event is not a valid command result. It gives a new
	// state for the stream with no other event types.
	changed := holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Telco})
	data, err := app.DataHolderType.Encode([]domain.DataHolderEvent{domain.DataHolderImported{DataHolder: changed}})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataHolderType.StreamType, StreamID: "dh-1", ExpectedVersion: 1, Events: data,
	}))

	got, err := readmodel.NewQueries(pool).DataHolderBrands(t.Context(), nil)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, changed.Brands[0], got[0].Brand)
}

func TestSoftwareProductClientGivesTheDataOfTheTokenEndpoint(t *testing.T) {
	store, pool := newStore(t)

	const thumbprint = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient(
			"dr-1",
			domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Inactive)},
		),
	}}))

	queries := readmodel.NewQueries(pool)

	before, err := queries.SoftwareProductClient(t.Context(), "sp-1")
	require.NoError(t, err)
	require.NoError(
		t,
		app.NewOperator(store).
			Execute(
				t.Context(),
				"dr-1",
				domain.RegisterClientCertificate{SoftwareProductID: "sp-1", Thumbprint: thumbprint},
			),
	)
	require.NoError(
		t,
		app.NewOperator(store).
			Execute(t.Context(), "dr-1", domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-1"}),
	)
	after, err := queries.SoftwareProductClient(t.Context(), "sp-1")
	require.NoError(t, err)

	assert.Equal(t, app.SoftwareProductClient{
		SoftwareProductID:   "sp-1",
		Status:              domain.Active,
		DataRecipientStatus: domain.DataRecipientActive,
		JWKSURI:             "https://recipient.example/jwks",
		ClientCertificates:  []string{},
	}, before)
	assert.Equal(t, app.SoftwareProductClient{
		SoftwareProductID:   "sp-1",
		Status:              domain.Inactive,
		DataRecipientStatus: domain.DataRecipientActive,
		JWKSURI:             "https://recipient.example/jwks",
		ClientCertificates:  []string{thumbprint},
	}, after)
}

// The token endpoint refuses a certificate that is not in the list, so a removal takes effect at the next token
// request.
func TestSoftwareProductClientHasNoThumbprintAfterItsRemoval(t *testing.T) {
	store, pool := newStore(t)

	const thumbprint = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}}))

	operator := app.NewOperator(store)
	require.NoError(t, operator.Execute(t.Context(), "dr-1",
		domain.RegisterClientCertificate{SoftwareProductID: "sp-1", Thumbprint: thumbprint}))
	require.NoError(t, operator.Execute(t.Context(), "dr-1",
		domain.RemoveClientCertificate{SoftwareProductID: "sp-1", Thumbprint: thumbprint}))

	got, err := readmodel.NewQueries(pool).SoftwareProductClient(t.Context(), "sp-1")

	require.NoError(t, err)
	assert.Empty(t, got.ClientCertificates)
}

func TestSoftwareProductClientGivesNotFoundForAnUnknownSoftwareProduct(t *testing.T) {
	_, pool := newStore(t)

	_, err := readmodel.NewQueries(pool).SoftwareProductClient(t.Context(), "sp-unknown")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestSoftwareProductClientHasTheStatusOfItsDataRecipient(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}}))
	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	client, err := readmodel.NewQueries(pool).SoftwareProductClient(t.Context(), "sp-1")

	require.NoError(t, err)
	assert.Equal(t, domain.DataRecipientSuspended, client.DataRecipientStatus)
	assert.Equal(t, domain.Inactive, client.Status)
}

func registeredBrandsOf(items []app.RegisteredDataHolderBrandItem) []string {
	ids := make([]string, 0, len(items))
	for _, i := range items {
		ids = append(ids, i.Brand.DataHolderBrandID)
	}

	return ids
}

// seedRegisteredBrands gives brands in each status, one interim brand, and the time before the load.
func seedRegisteredBrands(t *testing.T) (*readmodel.Queries, time.Time) {
	t.Helper()
	store, pool := newStore(t)
	active := holder(
		"dh-1",
		domain.DataHolderActive,
		[]domain.Industry{domain.Banking},
		[]domain.Industry{domain.Energy, domain.Telco},
		[]domain.Industry{domain.Banking},
	)
	active.Brands[1].Status = domain.Inactive
	active.Brands = append(active.Brands, domain.DataHolderBrand{
		InterimID: "dh-1-a", BrandName: "Interim", Industries: []domain.Industry{domain.Banking},
		LogoURI: "https://holder.example/interim.png", Status: domain.Active,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://prd.holder.example"},
	})

	var before time.Time
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT now()").Scan(&before))
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{
		holder("dh-2", domain.DataHolderRemoved, []domain.Industry{domain.Banking}),
		active,
	}}))

	return readmodel.NewQueries(pool), before
}

func TestRegisteredDataHolderBrandsGiveBrandsInEachStatusButNotInterimBrands(t *testing.T) {
	queries, _ := seedRegisteredBrands(t)
	industry := func(i domain.Industry) *domain.Industry { return &i }

	tests := map[string]struct {
		industry *domain.Industry
		want     []string
	}{
		"all":              {nil, []string{"dh-1-b1", "dh-1-b2", "dh-1-b3", "dh-2-b1"}},
		"banking":          {industry(domain.Banking), []string{"dh-1-b1", "dh-1-b3", "dh-2-b1"}},
		"telco":            {industry(domain.Telco), []string{"dh-1-b2"}},
		"non-bank-lending": {industry(domain.NonBankLending), []string{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := queries.RegisteredDataHolderBrands(
				t.Context(),
				app.DataHolderBrandQuery{Industry: tt.industry, Page: 1, PageSize: 25},
			)

			require.NoError(t, err)
			assert.Equal(t, tt.want, registeredBrandsOf(got))
			assert.Equal(t, len(tt.want), total)
		})
	}
}

func TestRegisteredDataHolderBrandsGiveOnePageAndTheTotalOfAllPages(t *testing.T) {
	queries, _ := seedRegisteredBrands(t)

	tests := map[string]struct {
		page, size int
		want       []string
	}{
		"first page":        {1, 3, []string{"dh-1-b1", "dh-1-b2", "dh-1-b3"}},
		"last page":         {2, 3, []string{"dh-2-b1"}},
		"page out of range": {3, 3, []string{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := queries.RegisteredDataHolderBrands(
				t.Context(),
				app.DataHolderBrandQuery{Page: tt.page, PageSize: tt.size},
			)

			require.NoError(t, err)
			assert.Equal(t, tt.want, registeredBrandsOf(got))
			assert.Equal(t, 4, total)
		})
	}
}

func TestARegisteredDataHolderBrandHasItsLegalEntityAndTheTimeOfItsLastEvent(t *testing.T) {
	queries, before := seedRegisteredBrands(t)

	got, _, err := queries.RegisteredDataHolderBrands(t.Context(), app.DataHolderBrandQuery{Page: 4, PageSize: 1})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, app.DataHolderLegalEntity{
		LegalEntityID:   "dh-2",
		LegalEntityName: "Holder dh-2",
		LogoURI:         "https://holder.example/logo.png",
		Status:          domain.DataHolderRemoved,
	}, got[0].LegalEntity)
	assert.Equal(t, domain.Removed, got[0].Brand.Status)
	assert.Equal(t, "https://mtls.holder.example", got[0].Brand.EndpointDetail.ResourceBaseURI)
	assert.False(t, got[0].LastUpdated.Before(before))
}

func TestRegisteredDataHolderBrandsUpdatedSinceIncludesTheSameTime(t *testing.T) {
	queries, _ := seedRegisteredBrands(t)
	all, _, err := queries.RegisteredDataHolderBrands(t.Context(), app.DataHolderBrandQuery{Page: 1, PageSize: 25})
	require.NoError(t, err)

	lastUpdated := all[0].LastUpdated
	after := lastUpdated.Add(time.Microsecond)

	same, sameTotal, err := queries.RegisteredDataHolderBrands(
		t.Context(),
		app.DataHolderBrandQuery{UpdatedSince: &lastUpdated, Page: 1, PageSize: 25},
	)
	require.NoError(t, err)
	later, laterTotal, err := queries.RegisteredDataHolderBrands(
		t.Context(),
		app.DataHolderBrandQuery{UpdatedSince: &after, Page: 1, PageSize: 25},
	)
	require.NoError(t, err)

	assert.Contains(t, registeredBrandsOf(same), all[0].Brand.DataHolderBrandID)
	assert.Positive(t, sameTotal)
	assert.NotContains(t, registeredBrandsOf(later), all[0].Brand.DataHolderBrandID)
	assert.Less(t, laterTotal, 4)
}

func TestDataRecipientOfSoftwareProductGivesTheCurrentDataRecipient(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
		recipient(
			"dr-2",
			domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-2", domain.Active), product("sp-3", domain.Active)},
		),
	}}))
	require.NoError(
		t,
		app.NewOperator(store).
			Execute(t.Context(), "dr-2", domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-3"}),
	)

	queries := readmodel.NewQueries(pool)

	dr, err := queries.DataRecipientOfSoftwareProduct(t.Context(), "sp-3")
	require.NoError(t, err)
	_, notFound := queries.DataRecipientOfSoftwareProduct(t.Context(), "sp-unknown")

	assert.Equal(t, "dr-2", dr.LegalEntityID)
	assert.Equal(t, "dr-2-b1", dr.Brands[0].DataRecipientBrandID)
	assert.Equal(t, domain.Inactive, dr.Brands[0].SoftwareProducts[1].Status)
	assert.ErrorIs(t, notFound, domain.ErrNotFound)
}

type readModels struct {
	recipientStatuses []app.DataRecipientStatusItem
	productStatuses   []app.SoftwareProductStatusItem
	recipients        []app.DataRecipientItem
	client            app.SoftwareProductClient
	holderStatuses    []app.DataHolderStatusItem
	holderBrands      []app.DataHolderBrandItem
	registeredBrands  []app.RegisteredDataHolderBrandItem
	recipientPage     []app.DataRecipientItem
	holderPage        []app.DataHolderItem
	recipientHistory  []app.HistoryEntry
	holderHistory     []app.HistoryEntry
}

func readAll(t *testing.T, pool *pgxpool.Pool) readModels {
	t.Helper()

	var (
		got     readModels
		err     error
		queries = readmodel.NewQueries(pool)
	)

	got.recipientStatuses, err = queries.DataRecipientStatuses(t.Context())
	require.NoError(t, err)
	got.productStatuses, err = queries.SoftwareProductStatuses(t.Context())
	require.NoError(t, err)
	got.recipients, err = queries.DataRecipients(t.Context())
	require.NoError(t, err)
	got.client, err = queries.SoftwareProductClient(t.Context(), "sp-1")
	require.NoError(t, err)
	got.holderStatuses, err = queries.DataHolderStatuses(t.Context(), nil)
	require.NoError(t, err)
	got.holderBrands, err = queries.DataHolderBrands(t.Context(), nil)
	require.NoError(t, err)
	got.registeredBrands, _, err = queries.RegisteredDataHolderBrands(
		t.Context(),
		app.DataHolderBrandQuery{Page: 1, PageSize: 25},
	)
	require.NoError(t, err)
	got.recipientPage, _, err = queries.DataRecipientPage(
		t.Context(),
		app.DataRecipientPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25},
	)
	require.NoError(t, err)
	got.holderPage, _, err = queries.DataHolderPage(
		t.Context(),
		app.DataHolderPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25},
	)
	require.NoError(t, err)
	got.recipientHistory, _, err = queries.History(
		t.Context(),
		app.DataRecipientType.StreamType,
		app.HistoryPageQuery{LegalEntityID: "dr-1", Page: 1, PageSize: 25},
	)
	require.NoError(t, err)
	got.holderHistory, _, err = queries.History(
		t.Context(), app.DataHolderType.StreamType, app.HistoryPageQuery{LegalEntityID: "dh-1", Page: 1, PageSize: 25},
	)
	require.NoError(t, err)

	return got
}

func TestProjectAllMakesTheReadModelsThatTheAppendsMade(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataHolders: []domain.DataHolder{
			holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking}),
			holder("dh-2", domain.DataHolderRemoved),
		},
		DataRecipients: []domain.DataRecipient{
			recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
			recipient("dr-2", domain.DataRecipientRevoked),
		},
	}))
	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	want := readAll(t, pool)
	require.Len(t, want.recipients, 2)
	require.Len(t, want.holderStatuses, 2)
	require.Len(t, want.recipientPage, 2)
	require.Len(t, want.holderPage, 2)
	require.NotEmpty(t, want.registeredBrands)
	require.Len(t, want.recipientHistory, 3)
	require.Len(t, want.holderHistory, 1)

	_, err := pool.Exec(t.Context(), `TRUNCATE data_recipient_statuses, data_recipients, software_product_statuses,
		software_product_clients, data_holder_statuses, data_holder_brands, data_holders, participant_history CASCADE`)
	require.NoError(t, err)

	require.NoError(t, store.ProjectAll(t.Context()))

	assert.Equal(t, want, readAll(t, pool))
}

func TestProjectAllDoesNotChangeReadModelsThatAreCorrect(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataHolders: []domain.DataHolder{holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking})},
		DataRecipients: []domain.DataRecipient{
			recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
		},
	}))

	want := readAll(t, pool)

	require.NoError(t, store.ProjectAll(t.Context()))

	assert.Equal(t, want, readAll(t, pool))
}
