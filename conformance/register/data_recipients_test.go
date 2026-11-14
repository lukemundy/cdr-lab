//go:build conformance

package register

import (
	"encoding/json"
	"maps"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The specification permits other fields, so response validation does not
// find SSA metadata that must not be in the response.
var (
	dataRecipientFields = []string{
		"legalEntityId", "legalEntityName", "accreditationNumber", "accreditationLevel", "logoUri",
		"dataRecipientBrands", "status", "lastUpdated",
	}
	dataRecipientBrandFields = []string{"dataRecipientBrandId", "brandName", "logoUri", "softwareProducts", "status"}
	softwareProductFields    = []string{
		"softwareProductId",
		"softwareProductName",
		"softwareProductDescription",
		"logoUri",
		"status",
	}
)

type object map[string]json.RawMessage

func (o object) string(t *testing.T, field string) string {
	t.Helper()

	var s string
	require.NoError(t, json.Unmarshal(o[field], &s), field)

	return s
}

func (o object) objects(t *testing.T, field string) []object {
	t.Helper()

	var list []object
	if raw, ok := o[field]; ok {
		require.NoError(t, json.Unmarshal(raw, &list), field)
	}

	return list
}

func dataRecipients(t *testing.T) []object {
	t.Helper()
	resp := call(t, dataRecipientsPath, xv("4"))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "4", resp.Header.Get("x-v"))

	var body struct {
		Data []object `json:"data"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &body))
	require.NotEmpty(t, body.Data)

	return body.Data
}

func statuses(t *testing.T, path, idField string) map[string]string {
	t.Helper()
	resp := call(t, path, v3())
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Data []map[string]string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &body))

	got := map[string]string{}
	for _, item := range body.Data {
		got[item[idField]] = item["status"]
	}

	return got
}

func TestDataRecipientsHaveEachDataRecipientAndSoftwareProductOfTheStatusLists(t *testing.T) {
	recipients := map[string]string{}
	products := map[string]string{}

	for _, dr := range dataRecipients(t) {
		recipients[dr.string(t, "legalEntityId")] = dr.string(t, "status")
		for _, brand := range dr.objects(t, "dataRecipientBrands") {
			for _, sp := range brand.objects(t, "softwareProducts") {
				products[sp.string(t, "softwareProductId")] = sp.string(t, "status")
			}
		}
	}

	assert.Equal(t, statuses(t, dataRecipientsStatusPath, "legalEntityId"), recipients)
	assert.Equal(t, statuses(t, softwareProductsStatusPath, "softwareProductId"), products)
}

func TestDataRecipientsHaveOnlyTheFieldsOfTheSchema(t *testing.T) {
	requireOnly := func(t *testing.T, o object, fields []string) {
		t.Helper()

		for field := range maps.Keys(o) {
			assert.Contains(t, fields, field)
		}
	}

	var sawProduct bool

	for _, dr := range dataRecipients(t) {
		requireOnly(t, dr, dataRecipientFields)

		for _, brand := range dr.objects(t, "dataRecipientBrands") {
			requireOnly(t, brand, dataRecipientBrandFields)

			for _, sp := range brand.objects(t, "softwareProducts") {
				sawProduct = true

				requireOnly(t, sp, softwareProductFields)
			}
		}
	}

	assert.True(t, sawProduct, "no data recipient has a software product")
}

func TestDataRecipientsHaveALastUpdatedDateTime(t *testing.T) {
	for _, dr := range dataRecipients(t) {
		lastUpdated := dr.string(t, "lastUpdated")

		// CDS DateTimeString is an RFC 3339 date-time.
		_, err := time.Parse(time.RFC3339, lastUpdated)
		assert.NoError(t, err, "data recipient %s: lastUpdated %q", dr.string(t, "legalEntityId"), lastUpdated)
	}
}

func TestDataRecipientsDoNotServeVersion3(t *testing.T) {
	resp := call(t, dataRecipientsPath, v3())

	requireCDSError(
		t,
		resp,
		http.StatusNotAcceptable,
		"urn:au-cds:error:cds-all:Header/UnsupportedVersion",
		"Unsupported Version",
	)
}
