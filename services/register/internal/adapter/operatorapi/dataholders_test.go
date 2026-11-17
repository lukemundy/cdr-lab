package operatorapi_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const dataHolderJSON = `{
	"legalEntityId": "dh-1", "legalEntityName": "Holder", "logoUri": "https://holder.example/logo.png",
	"status": "ACTIVE",
	"dataHolderBrands": [` + brandJSON + `]
}`

const brandJSON = `{
	"interimId": "dhb-1", "brandName": "Holder Bank", "industries": ["banking"],
	"logoUri": "https://holder.example/brand.png", "status": "ACTIVE",
	"endpointDetail": {"publicBaseUri": "https://holder.example"}, "jwksEndpoint": ""
}`

func dataHolderBrand() domain.DataHolderBrand {
	return domain.DataHolderBrand{
		InterimID: "dhb-1", BrandName: "Holder Bank", Industries: []domain.Industry{domain.Banking},
		LogoURI: "https://holder.example/brand.png", Status: domain.Active,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://holder.example"},
	}
}

// Verifies: REG-0809, REG-0815, REG-0816, REG-0874
func TestACreateOfADataHolderExecutesItsCommandAndGivesTheLocation(t *testing.T) {
	operator := &fakeOperator{}

	rec := serveBody(t, operator, "/v1/data-holders", []byte(dataHolderJSON))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "/v1/data-holders/dh-1", rec.Header().Get("Location"))
	assert.Empty(t, rec.Body.String())
	assert.Equal(t, "dh-1", operator.legalEntityID)
	assert.Equal(t, domain.CreateDataHolder{DataHolder: domain.DataHolder{
		LegalEntityID: "dh-1", LegalEntityName: "Holder", LogoURI: "https://holder.example/logo.png",
		Status: domain.DataHolderActive, Brands: []domain.DataHolderBrand{dataHolderBrand()},
	}}, operator.dataHolderCommand)
}

// Verifies: REG-0811, REG-0877
func TestAnAddOfADataHolderBrandExecutesItsCommandAndGivesTheLocation(t *testing.T) {
	operator := &fakeOperator{}

	rec := serveBody(t, operator, "/v1/data-holders/dh-1/brands", []byte(brandJSON))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "/v1/data-holders/dh-1/brands/dhb-1", rec.Header().Get("Location"))
	assert.Equal(t, "dh-1", operator.legalEntityID)
	assert.Equal(t, domain.AddDataHolderBrand{Brand: dataHolderBrand()}, operator.dataHolderCommand)
}

// Verifies: REG-0803, REG-0804, REG-0817, REG-0875, REG-0876, REG-0878, REG-0879
func TestEachDataHolderChangeExecutesItsCommand(t *testing.T) {
	const brand = "/v1/data-holders/dh-1/brands/dhb-1"

	tests := map[string]struct {
		method, path, body string
		want               domain.DataHolderCommand
	}{
		"change the details": {
			http.MethodPut, "/v1/data-holders/dh-1",
			`{"legalEntityName": "Holder", "logoUri": "https://holder.example/logo.png"}`,
			domain.ChangeDataHolderDetails{Details: domain.DataHolderDetails{
				LegalEntityName: "Holder", LogoURI: "https://holder.example/logo.png",
			}},
		},
		"remove": {http.MethodPost, "/v1/data-holders/dh-1/remove", "", domain.RemoveDataHolder{}},
		"change the details of a brand": {
			http.MethodPut, brand,
			`{"brandName": "Bank", "industries": ["banking"], "logoUri": "https://holder.example/b.png",
			  "endpointDetail": {"publicBaseUri": "https://holder.example"}, "jwksEndpoint": ""}`,
			domain.ChangeDataHolderBrandDetails{BrandID: "dhb-1", Details: domain.DataHolderBrandDetails{
				BrandName: "Bank", Industries: []domain.Industry{domain.Banking},
				LogoURI:        "https://holder.example/b.png",
				EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://holder.example"},
			}},
		},
		"activate a brand": {
			http.MethodPost,
			brand + "/activate",
			"",
			domain.ActivateDataHolderBrand{BrandID: "dhb-1"},
		},
		"deactivate a brand": {
			http.MethodPost,
			brand + "/deactivate",
			"",
			domain.DeactivateDataHolderBrand{BrandID: "dhb-1"},
		},
		"remove a brand": {http.MethodPost, brand + "/remove", "", domain.RemoveDataHolderBrand{BrandID: "dhb-1"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			operator := &fakeOperator{}

			rec := serveMethodBody(t, operator, tt.method, tt.path, []byte(tt.body))

			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			assert.Empty(t, rec.Body.String())
			assert.Equal(t, "dh-1", operator.legalEntityID)
			assert.Equal(t, tt.want, operator.dataHolderCommand)
		})
	}
}

// Verifies: REG-0852, REG-0853
func TestABodyThatIsNotOneRecordOfTheActionIsABadRequest(t *testing.T) {
	tests := map[string]struct{ method, path, body, wantDetail string }{
		"not JSON": {http.MethodPost, "/v1/data-holders", "{", "The request body is not valid JSON"},
		"two objects": {
			http.MethodPost, "/v1/data-holders", dataHolderJSON + dataHolderJSON, "more than one JSON value",
		},
		"a field that the action does not accept": {
			http.MethodPut, "/v1/data-holders/dh-1",
			`{"legalEntityName": "Holder", "logoUri": "https://holder.example/logo.png", "status": "REMOVED"}`,
			`unknown field "status"`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			operator := &fakeOperator{}

			rec := serveMethodBody(t, operator, tt.method, tt.path, []byte(tt.body))

			p := requireProblem(t, rec, http.StatusBadRequest)
			assert.Contains(t, p.Detail, tt.wantDetail)
			assert.Nil(t, operator.dataHolderCommand)
		})
	}
}

// Verifies: REG-0854
func TestADataHolderBodyLargerThan1MiBGives413(t *testing.T) {
	operator := &fakeOperator{}
	body := `{"legalEntityName": "` + strings.Repeat("a", 1<<20) + `"}`

	rec := serveMethodBody(t, operator, http.MethodPost, "/v1/data-holders", []byte(body))

	p := requireProblem(t, rec, http.StatusRequestEntityTooLarge)
	assert.Equal(t, "The request body is larger than 1 MiB.", p.Detail)
	assert.Nil(t, operator.dataHolderCommand)
}

// Verifies: REG-0810, REG-0855, REG-0858, REG-0861
func TestDataHolderErrors(t *testing.T) {
	_, notValid := domain.HandleDataHolder(domain.DataHolderState{}, domain.CreateDataHolder{})

	tests := map[string]struct {
		err        error
		status     int
		wantDetail string
	}{
		"not valid": {notValid, http.StatusBadRequest, notValid.Error()},
		"exists": {
			fmt.Errorf("data holder %w", domain.ErrExists), http.StatusConflict, "data holder already exists",
		},
		"ID not unique": {
			fmt.Errorf("project data-holder dh-1: %w", domain.IDNotUniqueError{ID: "dhb-1"}),
			http.StatusConflict,
			`the ID "dhb-1" is not unique: a different record has it`,
		},
		"not found": {
			fmt.Errorf("data holder %w", domain.ErrNotFound), http.StatusNotFound, "data holder does not exist",
		},
		"version conflict": {
			eventsourcing.ErrVersionConflict,
			http.StatusConflict,
			"The data holder changed during the request. Send the request again.",
		},
		"other": {
			errors.New("connection to 10.0.0.1 failed"),
			http.StatusInternalServerError,
			"The Register could not complete the request.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serveBody(t, &fakeOperator{err: tt.err}, "/v1/data-holders", []byte(dataHolderJSON))

			p := requireProblem(t, rec, tt.status)
			assert.Equal(t, tt.wantDetail, p.Detail)
		})
	}
}
