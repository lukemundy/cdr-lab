package console_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func brand(id, name string, status domain.Status, industries ...domain.Industry) domain.DataHolderBrand {
	return domain.DataHolderBrand{
		DataHolderBrandID: id, BrandName: name, Industries: industries, LogoURI: "https://holder.example/brand.png",
		Status: status, JWKSEndpoint: "https://holder.example/jwks",
		EndpointDetail: domain.EndpointDetail{
			Version: "1", PublicBaseURI: "https://api.holder.example", ResourceBaseURI: "https://mtls.holder.example",
			InfosecBaseURI: "https://id.holder.example", WebsiteURI: "https://holder.example",
		},
	}
}

func dataHolder(id, name string, status domain.DataHolderStatus, brands ...domain.DataHolderBrand) domain.DataHolder {
	return domain.DataHolder{
		LegalEntityID: id, LegalEntityName: name, LogoURI: "https://holder.example/logo.png", Status: status,
		Brands: brands,
	}
}

// Verifies: REG-1116, REG-1126
func TestDataHoldersGiveOnePageWithTheBrandCountTheIndustriesAndTheTotal(t *testing.T) {
	queries := seedParticipants(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive,
			brand("dh-1-b1", "Alpha Energy", domain.Active, domain.Telco, domain.Energy),
			brand("dh-1-b2", "Alpha Bank", domain.Inactive, domain.Banking)),
		dataHolder("dh-2", "Beta Ltd", domain.DataHolderActive),
		dataHolder("dh-3", "Gamma Ltd", domain.DataHolderRemoved,
			brand("dh-3-b1", "Gamma Bank", domain.Removed, domain.Banking)),
	}})

	resp := get(t, queries, "/api/data-holders?pageSize=2")

	assert.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
	assert.JSONEq(t, `{"totalRecords": 3, "data": [
		{"legalEntityId": "dh-1", "legalEntityName": "Alpha Ltd", "status": "ACTIVE", "brandCount": 2,
		 "industries": ["banking", "energy", "telco"]},
		{"legalEntityId": "dh-2", "legalEntityName": "Beta Ltd", "status": "ACTIVE", "brandCount": 0,
		 "industries": []}
	]}`, resp.body)
}

// Verifies: REG-1124
func TestDataHoldersGiveTheDataHolderOfABrandName(t *testing.T) {
	queries := seedParticipants(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Mock Holdings Ltd", domain.DataHolderActive,
			brand("dh-1-b1", "Mock Mutual", domain.Active, domain.Banking)),
		dataHolder("dh-2", "Other Ltd", domain.DataHolderActive),
	}})

	resp := get(t, queries, "/api/data-holders?q=MUTUAL")

	assert.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 1, "data": [
		{"legalEntityId": "dh-1", "legalEntityName": "Mock Holdings Ltd", "status": "ACTIVE", "brandCount": 1,
		 "industries": ["banking"]}
	]}`, resp.body)
}

func TestDataHoldersGiveTheQueryParametersToTheReadModel(t *testing.T) {
	removed := domain.DataHolderRemoved

	tests := map[string]struct {
		query string
		want  app.DataHolderPageQuery
	}{
		"defaults": {"", app.DataHolderPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25}},
		"each parameter": {
			"?status=REMOVED&q=Mock%20B&sort=status&order=desc&page=3&pageSize=100",
			app.DataHolderPageQuery{
				Status: &removed, NameContains: "Mock B", Sort: app.SortByStatus, Descending: true, Page: 3,
				PageSize: 100,
			},
		},
		"sort by name ascending": {
			"?sort=legalEntityName&order=asc",
			app.DataHolderPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25},
		},
		"sort by brand count descending": {
			"?sort=brandCount&order=desc",
			app.DataHolderPageQuery{Sort: app.SortByBrandCount, Descending: true, Page: 1, PageSize: 25},
		},
		"empty values are the defaults": {
			"?status=&q=&sort=&order=&page=&pageSize=",
			app.DataHolderPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25},
		},
		"an unknown parameter is ignored": {
			"?unknown=1", app.DataHolderPageQuery{Sort: app.SortByName, Page: 1, PageSize: 25},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := &queryRecorder{}

			resp := get(t, reader, "/api/data-holders"+tt.query)

			assert.Equal(t, http.StatusOK, resp.status)
			assert.JSONEq(t, `{"totalRecords": 0, "data": []}`, resp.body)
			assert.Equal(t, tt.want, reader.gotHolder)
		})
	}
}

func TestDataHoldersRefuseAParameterThatIsNotValid(t *testing.T) {
	tests := map[string]string{
		"page 0":                       "?page=0",
		"page not an integer":          "?page=one",
		"page size 0":                  "?pageSize=0",
		"page size more than 100":      "?pageSize=101",
		"status of a data recipient":   "?status=SUSPENDED",
		"status in lower case":         "?status=active",
		"sort of the data recipients":  "?sort=accreditationNumber",
		"count of the data recipients": "?sort=softwareProductCount",
		"unknown sort":                 "?sort=lastUpdated",
		"unknown order":                "?order=up",
	}
	for name, query := range tests {
		t.Run(name, func(t *testing.T) {
			reader := &queryRecorder{}

			resp := get(t, reader, "/api/data-holders"+query)

			assert.Equal(t, http.StatusBadRequest, resp.status)
			assert.Equal(t, "application/problem+json", resp.header.Get("Content-Type"))
			assert.Zero(t, reader.gotHolder, "the read model gets no query")
		})
	}
}

func TestDataHoldersGiveAProblemWithNoErrorMessageWhenTheReadFails(t *testing.T) {
	resp := get(t, failingReader{}, "/api/data-holders")

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Holdings")
}

// The page size bounds the response, so the first page does not depend on the number of data holders.
func TestTheFirstPageOf1000DataHoldersHas25Records(t *testing.T) {
	holders := make([]domain.DataHolder, 0, 1000)

	for i := range 1000 {
		id := fmt.Sprintf("dh-%04d", i)
		holders = append(holders, dataHolder(id, "Holder "+id, domain.DataHolderActive,
			brand(id+"-b1", "Brand "+id+" Bank", domain.Active, domain.Banking),
			brand(id+"-b2", "Brand "+id+" Energy", domain.Active, domain.Energy),
			brand(id+"-b3", "Brand "+id+" Mobile", domain.Inactive, domain.Telco)))
	}

	queries := seedParticipants(t, domain.Participants{DataHolders: holders})

	start := time.Now()
	resp := get(t, queries, "/api/data-holders?q=mobile&sort=status&order=desc")
	elapsed := time.Since(start)

	var list struct {
		Data         []json.RawMessage `json:"data"`
		TotalRecords int               `json:"totalRecords"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &list))
	assert.Len(t, list.Data, 25)
	assert.Equal(t, 1000, list.TotalRecords)
	assert.Less(t, elapsed, time.Second)
	t.Logf("first page of 1,000 data holders: %s", elapsed)
}
