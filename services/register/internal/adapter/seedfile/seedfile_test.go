package seedfile_test

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/seedfile"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func TestReadDecodesParticipants(t *testing.T) {
	p, err := seedfile.Read(strings.NewReader(`{
		"dataHolders": [{
			"legalEntityId": "dh-1", "status": "ACTIVE",
			"dataHolderBrands": [{"dataHolderBrandId": "dhb-1", "industries": ["energy"]}]
		}],
		"dataRecipients": [{
			"legalEntityId": "dr-1",
			"dataRecipientBrands": [{"softwareProducts": [{"softwareProductId": "sp-1", "scopes": ["openid"]}]}]
		}]
	}`))

	require.NoError(t, err)
	assert.Equal(t, domain.DataHolderActive, p.DataHolders[0].Status)
	assert.Equal(t, []domain.Industry{domain.Energy}, p.DataHolders[0].Brands[0].Industries)
	assert.Equal(t, []string{"openid"}, p.DataRecipients[0].Brands[0].SoftwareProducts[0].Scopes)
}

func TestReadRejectsUnknownFields(t *testing.T) {
	_, err := seedfile.Read(strings.NewReader(`{"dataHolders": [{"legalEntityNumber": "dh-1"}]}`))

	assert.ErrorContains(t, err, `unknown field "legalEntityNumber"`)
}

func TestReadRejectsDataAfterTheParticipants(t *testing.T) {
	_, err := seedfile.Read(strings.NewReader(`{} {}`))

	assert.ErrorContains(t, err, "data after the participants")
}

func readSeed(t *testing.T, name string) domain.Participants {
	t.Helper()

	f, err := os.DirFS("../../../seed").Open(name)
	require.NoError(t, err)

	defer func() { _ = f.Close() }()

	p, err := seedfile.Read(f)
	require.NoError(t, err)

	require.NoError(t, p.Validate())

	return p
}

func TestTheExampleFileIsValidAndHasEachStatus(t *testing.T) {
	p := readSeed(t, "example.json")

	holders := map[domain.DataHolderStatus]bool{}
	holderBrands := map[domain.Status]bool{}

	for _, dh := range p.DataHolders {
		_, err := domain.HandleDataHolder(domain.DataHolderState{}, domain.ImportDataHolder{DataHolder: dh})
		require.NoError(t, err)

		holders[dh.Status] = true
		for _, b := range dh.Brands {
			holderBrands[b.Status] = true
		}
	}

	recipients := map[domain.DataRecipientStatus]bool{}
	recipientBrands := map[domain.Status]bool{}
	products := map[domain.Status]bool{}

	for _, dr := range p.DataRecipients {
		_, err := domain.HandleDataRecipient(domain.DataRecipientState{}, domain.ImportDataRecipient{DataRecipient: dr})
		require.NoError(t, err)

		recipients[dr.Status] = true
		for _, b := range dr.Brands {
			recipientBrands[b.Status] = true
			for _, sp := range b.SoftwareProducts {
				products[sp.Status] = true
			}
		}
	}

	all := map[domain.Status]bool{domain.Active: true, domain.Inactive: true, domain.Removed: true}

	assert.Equal(
		t,
		map[domain.DataHolderStatus]bool{domain.DataHolderActive: true, domain.DataHolderRemoved: true},
		holders,
	)
	assert.Equal(t, all, holderBrands)
	assert.Equal(t, map[domain.DataRecipientStatus]bool{
		domain.DataRecipientActive:      true,
		domain.DataRecipientSuspended:   true,
		domain.DataRecipientRevoked:     true,
		domain.DataRecipientSurrendered: true,
	}, recipients)
	assert.Equal(t, all, recipientBrands)
	assert.Equal(t, all, products)
}

func TestThePresetBanksFileHasTheFourBanksOfTheLabOnTheirOwnHosts(t *testing.T) {
	p := readSeed(t, "preset-banks.json")
	lab := map[string]domain.DataHolder{}

	for _, dh := range readSeed(t, "example.json").DataHolders {
		lab[dh.LegalEntityID] = dh
	}

	assert.Empty(t, p.DataRecipients)

	var brands []string

	for _, dh := range p.DataHolders {
		_, err := domain.HandleDataHolder(domain.DataHolderState{}, domain.ImportDataHolder{DataHolder: dh})
		require.NoError(t, err)

		labHolder, ok := lab[dh.LegalEntityID]
		require.True(t, ok, dh.LegalEntityName)
		require.Len(t, dh.Brands, 1, dh.LegalEntityName)
		require.Len(t, labHolder.Brands, 1, dh.LegalEntityName)

		assert.Equal(t, labHolder.LegalEntityName, dh.LegalEntityName)
		assert.Equal(t, domain.DataHolderActive, dh.Status, dh.LegalEntityName)
		assert.Equal(t, labHolder.Status, dh.Status, dh.LegalEntityName)

		brand, labBrand := dh.Brands[0], labHolder.Brands[0]
		assert.Equal(t, labBrand.DataHolderBrandID, brand.DataHolderBrandID, brand.BrandName)
		assert.Equal(t, labBrand.BrandName, brand.BrandName)
		assert.Equal(t, labBrand.Status, brand.Status, brand.BrandName)
		assert.Equal(t, labBrand.Industries, brand.Industries, brand.BrandName)
		assert.Equal(t, labBrand.EndpointDetail.Version, brand.EndpointDetail.Version, brand.BrandName)
		assert.Equal(t, labBrand.EndpointDetail.WebsiteURI, brand.EndpointDetail.WebsiteURI, brand.BrandName)

		website, err := url.Parse(brand.EndpointDetail.WebsiteURI)
		require.NoError(t, err)

		public := "https://" + website.Host + ":8081"
		assert.Equal(t, domain.EndpointDetail{
			Version:         labBrand.EndpointDetail.Version,
			PublicBaseURI:   public,
			ProductBaseURI:  public,
			ResourceBaseURI: "https://" + website.Host + ":8083",
			InfosecBaseURI:  public,
			WebsiteURI:      labBrand.EndpointDetail.WebsiteURI,
		}, brand.EndpointDetail, brand.BrandName)
		assert.Equal(t, public+"/brand/logo.svg", brand.LogoURI, brand.BrandName)
		assert.Equal(t, public+"/brand/logo.svg", dh.LogoURI, brand.BrandName)
		assert.Equal(t, public+"/jwks", brand.JWKSEndpoint, brand.BrandName)

		brands = append(brands, brand.BrandName)
	}

	assert.ElementsMatch(t, []string{"Wattle Bank", "Redgum Bank", "Harbour Bank", "Kestrel Bank"}, brands)
}
