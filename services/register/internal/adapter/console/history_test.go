package console_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
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

func (*queryRecorder) History(context.Context, string, app.HistoryPageQuery) ([]app.HistoryEntry, int, error) {
	return nil, 0, errors.New("not used")
}

func (failingReader) History(context.Context, string, app.HistoryPageQuery) ([]app.HistoryEntry, int, error) {
	return nil, 0, errors.New("a message that contains Product sp-1")
}

func seedStore(t *testing.T, participants domain.Participants) (*eventstore.Store, *readmodel.Queries) {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	store := eventstore.New(pool, readmodel.Projections()...)
	require.NoError(t, app.Seed(t.Context(), store, participants))

	return store, readmodel.NewQueries(pool)
}

// withoutRecordedAt checks that each recordedAt is a UTC time near now, and gives the body with no recordedAt.
func withoutRecordedAt(t *testing.T, body string) string {
	t.Helper()

	var history struct {
		Data         []map[string]any `json:"data"`
		TotalRecords int              `json:"totalRecords"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &history))

	for _, entry := range history.Data {
		recordedAt, ok := entry["recordedAt"].(string)
		require.True(t, ok, "recordedAt is a string")
		assert.Regexp(t, `Z$`, recordedAt, "recordedAt is in UTC")

		at, err := time.Parse(time.RFC3339Nano, recordedAt)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), at, time.Minute)

		delete(entry, "recordedAt")
	}

	rest, err := json.Marshal(history)
	require.NoError(t, err)

	return string(rest)
}

// Verifies: REG-1016, REG-1142
func TestTheHistoryOfASuspendedDataRecipientGivesTheCascadeToEachActiveSoftwareProduct(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Inactive)},
			[]domain.SoftwareProduct{product("sp-3", domain.Active)}),
	}})
	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	resp := get(t, queries, "/api/data-recipients/dr-1/history")

	require.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
	assert.JSONEq(t, `{"totalRecords": 4, "data": [
		{"streamVersion": 4, "eventType": "SoftwareProductDeactivated",
		 "softwareProductId": "sp-3", "softwareProductName": "Product sp-3"},
		{"streamVersion": 3, "eventType": "SoftwareProductDeactivated",
		 "softwareProductId": "sp-1", "softwareProductName": "Product sp-1"},
		{"streamVersion": 2, "eventType": "AccreditationSuspended"},
		{"streamVersion": 1, "eventType": "DataRecipientImported"}
	]}`, withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1037
func TestTheHistoryOfADataRecipientGivesTheThumbprintOfARegisteredClientCertificate(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}})
	data, err := app.DataRecipientType.Encode([]domain.DataRecipientEvent{
		domain.ClientCertificateRegistered{SoftwareProductID: "sp-1", Thumbprint: thumbprint1},
	})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: "dr-1", ExpectedVersion: 1, Events: data,
	}))

	resp := get(t, queries, "/api/data-recipients/dr-1/history?pageSize=1")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 2, "data": [
		{"streamVersion": 2, "eventType": "ClientCertificateRegistered",
		 "softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "thumbprint": "`+thumbprint1+`"}
	]}`, withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1039
func TestTheHistoryOfADataRecipientGivesTheThumbprintOfARemovedClientCertificate(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}})
	data, err := app.DataRecipientType.Encode([]domain.DataRecipientEvent{
		domain.ClientCertificateRegistered{SoftwareProductID: "sp-1", Thumbprint: thumbprint1},
		domain.ClientCertificateRemoved{SoftwareProductID: "sp-1", Thumbprint: thumbprint1},
	})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: "dr-1", ExpectedVersion: 1, Events: data,
	}))

	resp := get(t, queries, "/api/data-recipients/dr-1/history?pageSize=1")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 3, "data": [
		{"streamVersion": 3, "eventType": "ClientCertificateRemoved",
		 "softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "thumbprint": "`+thumbprint1+`"}
	]}`, withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1038
func TestTheHistoryOfADataRecipientGivesTheThumbprintOfAnIssuedClientCertificate(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}})
	data, err := app.DataRecipientType.Encode([]domain.DataRecipientEvent{
		domain.ClientCertificateIssued{SoftwareProductID: "sp-1", Thumbprint: thumbprint1},
	})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: "dr-1", ExpectedVersion: 1, Events: data,
	}))

	resp := get(t, queries, "/api/data-recipients/dr-1/history?pageSize=1")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 2, "data": [
		{"streamVersion": 2, "eventType": "ClientCertificateIssued",
		 "softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "thumbprint": "`+thumbprint1+`"}
	]}`, withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1019
func TestTheHistoryOfADataHolderGivesItsImport(t *testing.T) {
	_, queries := seedStore(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive),
	}})

	resp := get(t, queries, "/api/data-holders/dh-1/history")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 1, "data": [{"streamVersion": 1, "eventType": "DataHolderImported"}]}`,
		withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1146
func TestTheHistoryOfADataHolderGivesTheBrandOfABrandEvent(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive,
			brand("dh-1-b1", "Alpha Bank", domain.Active, domain.Banking)),
	}})
	require.NoError(t, app.NewOperator(store).ExecuteDataHolder(t.Context(), "dh-1", domain.RemoveDataHolder{}))

	resp := get(t, queries, "/api/data-holders/dh-1/history")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 3, "data": [
		{"streamVersion": 3, "eventType": "DataHolderBrandRemoved", "brandId": "dh-1-b1", "brandName": "Alpha Bank"},
		{"streamVersion": 2, "eventType": "DataHolderLegalEntityRemoved"},
		{"streamVersion": 1, "eventType": "DataHolderImported"}
	]}`, withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1040, REG-1146
func TestTheHistoryOfADataHolderGivesTheBrandAndTheThumbprintOfAnIssuedServerCertificate(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive,
			brand("dh-1-b1", "Alpha Bank", domain.Active, domain.Banking)),
	}})
	data, err := app.DataHolderType.Encode([]domain.DataHolderEvent{
		domain.ServerCertificateIssued{BrandID: "dh-1-b1", Thumbprint: thumbprint1},
	})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataHolderType.StreamType, StreamID: "dh-1", ExpectedVersion: 1, Events: data,
	}))

	resp := get(t, queries, "/api/data-holders/dh-1/history?pageSize=1")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 2, "data": [
		{"streamVersion": 2, "eventType": "ServerCertificateIssued",
		 "brandId": "dh-1-b1", "brandName": "Alpha Bank", "thumbprint": "`+thumbprint1+`"}
	]}`, withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1146
func TestTheHistoryOfADataRecipientGivesTheBrandOfABrandEvent(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}})
	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1",
		domain.DeactivateDataRecipientBrand{BrandID: "dr-1-b1"}))

	resp := get(t, queries, "/api/data-recipients/dr-1/history?pageSize=2")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"totalRecords": 3, "data": [
		{"streamVersion": 3, "eventType": "SoftwareProductDeactivated",
		 "softwareProductId": "sp-1", "softwareProductName": "Product sp-1"},
		{"streamVersion": 2, "eventType": "DataRecipientBrandDeactivated", "brandId": "dr-1-b1", "brandName": "Brand"}
	]}`, withoutRecordedAt(t, resp.body))
}

// Verifies: REG-1115, REG-1116
func TestAHistoryGivesOnePageAndTheTotal(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}})
	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	for target, versions := range map[string][]int{
		"/api/data-recipients/dr-1/history?pageSize=2":        {3, 2},
		"/api/data-recipients/dr-1/history?pageSize=2&page=2": {1},
		"/api/data-recipients/dr-1/history?pageSize=2&page=3": {},
		"/api/data-recipients/dr-1/history?page=&pageSize=":   {3, 2, 1},
		"/api/data-recipients/dr-1/history?unknown=x&page=1":  {3, 2, 1},
	} {
		resp := get(t, queries, target)
		require.Equal(t, http.StatusOK, resp.status, target)

		var history struct {
			Data []struct {
				StreamVersion int `json:"streamVersion"`
			} `json:"data"`
			TotalRecords int `json:"totalRecords"`
		}
		require.NoError(t, json.Unmarshal([]byte(resp.body), &history))

		got := []int{}
		for _, e := range history.Data {
			got = append(got, e.StreamVersion)
		}

		assert.Equal(t, versions, got, target)
		assert.Equal(t, 3, history.TotalRecords, target)
	}
}

func TestAHistoryRefusesAPageThatIsNotValid(t *testing.T) {
	_, queries := seedStore(t, domain.Participants{DataHolders: []domain.DataHolder{
		dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive),
	}})

	for query, detail := range map[string]string{
		"page=0":       "page must be an integer of 1 or more",
		"page=x":       "page must be an integer of 1 or more",
		"pageSize=0":   "pageSize must be an integer of 1 or more",
		"pageSize=101": "pageSize must be 100 or less",
	} {
		resp := get(t, queries, "/api/data-holders/dh-1/history?"+query)

		assert.Equal(t, http.StatusBadRequest, resp.status, query)
		assert.Contains(t, resp.body, detail, query)
	}
}

// Verifies: REG-1111
func TestAHistoryOfAnUnknownLegalEntityGivesNotFoundWithNoIdentifier(t *testing.T) {
	_, queries := seedStore(t, domain.Participants{
		DataRecipients: []domain.DataRecipient{recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive)},
		DataHolders:    []domain.DataHolder{dataHolder("dh-1", "Alpha Ltd", domain.DataHolderActive)},
	})

	recipientDetail := "The Register has no data recipient with this legal entity ID."
	holderDetail := "The Register has no data holder with this legal entity ID."

	for path, detail := range map[string]string{
		"/api/data-recipients/unknown-id/history": recipientDetail,
		"/api/data-recipients/dh-1/history":       recipientDetail,
		"/api/data-holders/unknown-id/history":    holderDetail,
		"/api/data-holders/dr-1/history":          holderDetail,
	} {
		resp := get(t, queries, path)

		assert.Equal(t, http.StatusNotFound, resp.status, path)
		assert.JSONEq(t, `{"type": "about:blank", "title": "Not Found", "status": 404, "detail": "`+detail+`"}`,
			resp.body, path)
	}
}

func TestAHistoryGivesAProblemWithNoErrorMessageWhenTheReadFails(t *testing.T) {
	for _, path := range []string{"/api/data-recipients/dr-1/history", "/api/data-holders/dh-1/history"} {
		resp := get(t, failingReader{}, path)

		assert.Equal(t, http.StatusInternalServerError, resp.status, path)
		assert.NotContains(t, resp.body, "Product", path)
	}
}

// The check of the response is the check of the field rule: the schema of each event type is closed.
// Verifies: REG-1145
func TestTheDocumentRefusesAFieldThatTheTableDoesNotGiveForTheEventType(t *testing.T) {
	for name, entry := range map[string]string{
		"thumbprint of a suspension": `{"streamVersion": 2, "eventType": "AccreditationSuspended",
			"recordedAt": "2026-10-23T01:02:03Z", "thumbprint": "abc"}`,
		"thumbprint of a deactivation": `{"streamVersion": 2, "eventType": "SoftwareProductDeactivated",
			"recordedAt": "2026-10-23T01:02:03Z", "softwareProductId": "sp-1", "softwareProductName": "A",
			"thumbprint": "abc"}`,
		"no name of a software product": `{"streamVersion": 2, "eventType": "SoftwareProductRemoved",
			"recordedAt": "2026-10-23T01:02:03Z", "softwareProductId": "sp-1"}`,
		"payload": `{"streamVersion": 1, "eventType": "DataRecipientImported",
			"recordedAt": "2026-10-23T01:02:03Z", "payload": {}}`,
		"unknown event type": `{"streamVersion": 1, "eventType": "Unknown", "recordedAt": "2026-10-23T01:02:03Z"}`,
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/data-recipients/dr-1/history", nil)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"totalRecords": 1, "data": [` + entry + `]}`)),
				Request:    req,
			}

			assert.Error(t, loadSpec(t).ValidateResponse(resp))
		})
	}
}

func loadDocument(t *testing.T) *openapi3.T {
	t.Helper()

	data, err := os.ReadFile("openapi.yaml")
	require.NoError(t, err)
	doc, err := openapi3.NewLoader().LoadFromData(data)
	require.NoError(t, err)

	return doc
}

func historyEventTypes(t *testing.T, names ...string) []string {
	t.Helper()

	doc := loadDocument(t)

	var types []string

	for _, name := range names {
		schema, ok := doc.Components.Schemas[name]
		require.True(t, ok, name)

		for _, v := range schema.Value.Properties["eventType"].Value.Enum {
			eventType, ok := v.(string)
			require.True(t, ok)

			types = append(types, eventType)
		}
	}

	slices.Sort(types)

	return types
}

func TestTheDocumentGivesEachEventTypeOfTheCodecs(t *testing.T) {
	recipientCodec, ok := app.DataRecipientType.Codec.(*eventsourcing.JSONCodec[domain.DataRecipientEvent])
	require.True(t, ok)
	holderCodec, ok := app.DataHolderType.Codec.(*eventsourcing.JSONCodec[domain.DataHolderEvent])
	require.True(t, ok)

	assert.Equal(t, recipientCodec.EventTypes(), historyEventTypes(t,
		"LegalEntityHistoryEntry", "SoftwareProductHistoryEntry", "ClientCertificateHistoryEntry",
		"DataRecipientBrandHistoryEntry"))
	assert.Equal(t, holderCodec.EventTypes(), historyEventTypes(t,
		"DataHolderLegalEntityHistoryEntry", "DataHolderBrandHistoryEntry", "ServerCertificateHistoryEntry"))

	assert.Equal(
		t,
		recipientCodec.EventTypes(),
		slices.Sorted(
			maps.Keys(loadDocument(t).Components.Schemas["DataRecipientHistoryEntry"].Value.Discriminator.Mapping),
		),
		"the discriminator maps each event type",
	)
	assert.Equal(
		t,
		holderCodec.EventTypes(),
		slices.Sorted(
			maps.Keys(loadDocument(t).Components.Schemas["DataHolderHistoryEntry"].Value.Discriminator.Mapping),
		),
		"the discriminator maps each data holder event type",
	)
}

// The page size bounds the response, so the first page does not depend on the number of events.
func TestTheFirstPageOfAHistoryOf1000EventsHas25Entries(t *testing.T) {
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}})

	events := make([]domain.DataRecipientEvent, 0, 999)

	for i := range 999 {
		if i%2 == 0 {
			events = append(events, domain.AccreditationSuspended{})
		} else {
			events = append(events, domain.AccreditationReinstated{})
		}
	}

	data, err := app.DataRecipientType.Encode(events)
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: "dr-1", ExpectedVersion: 1, Events: data,
	}))

	start := time.Now()
	resp := get(t, queries, "/api/data-recipients/dr-1/history")
	elapsed := time.Since(start)

	var history struct {
		Data []struct {
			StreamVersion int `json:"streamVersion"`
		} `json:"data"`
		TotalRecords int `json:"totalRecords"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &history))
	require.Len(t, history.Data, 25)
	assert.Equal(t, 1000, history.Data[0].StreamVersion)
	assert.Equal(t, 1000, history.TotalRecords)
	assert.Less(t, elapsed, time.Second)
	t.Logf("first page of a history of 1,000 events: %s", elapsed)
}
