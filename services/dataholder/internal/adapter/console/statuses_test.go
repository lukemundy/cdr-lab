package console_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

type fixedStatusList struct {
	list  app.StatusList
	err   error
	query *app.StatusListQuery
}

func (f fixedStatusList) StatusList(_ context.Context, query app.StatusListQuery) (app.StatusList, error) {
	if f.query != nil {
		*f.query = query
	}

	return f.list, f.err
}

func TestTheRegisterStatusesGiveEachStatusItsRegistrationAndTheTimeOfTheLastResponse(t *testing.T) {
	resp := get(t, console.Ports{Statuses: fixedStatusList{list: app.StatusList{
		Products: []app.ProductStatus{
			{
				SoftwareProductID: "software-1", Status: domain.SoftwareProductActive,
				ClientID: "client-1", ClientName: "Recipient Web",
			},
			{SoftwareProductID: "software-2", Status: domain.SoftwareProductRemoved},
		},
		ReceivedAt: time.Date(2026, 10, 15, 13, 30, 0, 123456000, time.UTC),
	}}}, "/api/software-product-statuses")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{
		"data": [
			{"softwareProductId": "software-1", "status": "ACTIVE", "clientId": "client-1", "name": "Recipient Web"},
			{"softwareProductId": "software-2", "status": "REMOVED"}
		],
		"receivedAt": "2026-10-15T13:30:00.123456Z"
	}`, resp.body)
}

func TestTheRegisterStatusesAreFilteredAndSortedByTheQuery(t *testing.T) {
	registered, notRegistered := true, false

	for target, want := range map[string]app.StatusListQuery{
		"/api/software-product-statuses": {Sort: app.SortStatusesByName},
		"/api/software-product-statuses?registration=registered": {
			Sort: app.SortStatusesByName, Registered: &registered,
		},
		"/api/software-product-statuses?registration=notRegistered&sort=status&order=desc": {
			Sort: app.SortStatusesByStatus, Registered: &notRegistered, Descending: true,
		},
		"/api/software-product-statuses?sort=name&order=asc": {Sort: app.SortStatusesByName},
		// An empty value is no value.
		"/api/software-product-statuses?registration=&sort=&order=": {Sort: app.SortStatusesByName},
	} {
		var got app.StatusListQuery

		resp := get(t, console.Ports{Statuses: fixedStatusList{query: &got}}, target)

		require.Equal(t, http.StatusOK, resp.status, target)
		assert.Equal(t, want, got, target)
	}
}

func TestAStatusQueryParameterThatIsNotValidGivesBadRequestWithNoValue(t *testing.T) {
	for _, target := range []string{
		"/api/software-product-statuses?registration=zebracorn",
		"/api/software-product-statuses?sort=zebracorn",
		"/api/software-product-statuses?order=zebracorn",
	} {
		resp := get(t, console.Ports{Statuses: fixedStatusList{}}, target)

		require.Equal(t, http.StatusBadRequest, resp.status, target)
		assert.NotContains(t, resp.body, "zebracorn", target)
	}
}

func TestTheRegisterStatusesHaveNoTimeBeforeTheFirstResponse(t *testing.T) {
	resp := get(t, console.Ports{Statuses: fixedStatusList{}}, "/api/software-product-statuses")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"data": []}`, resp.body)
}

func TestAFailureToReadTheRegisterStatusesGives500(t *testing.T) {
	resp := get(t, console.Ports{Statuses: fixedStatusList{err: errors.New("database failure")}},
		"/api/software-product-statuses")

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "database failure")
}

type registerStatuses struct {
	statuses map[string]string
	etag     string
}

func (r *registerStatuses) SoftwareProductStatuses(_ context.Context, etag string) (app.StatusResponse, error) {
	if etag == r.etag {
		return app.StatusResponse{NotModified: true}, nil
	}

	return app.StatusResponse{Statuses: r.statuses, ETag: r.etag}, nil
}

type statusBody struct {
	Data []struct {
		SoftwareProductID string `json:"softwareProductId"`
		Status            string `json:"status"`
	} `json:"data"`
	ReceivedAt time.Time `json:"receivedAt"`
}

func readStatuses(t *testing.T, ports console.Ports) statusBody {
	t.Helper()

	resp := get(t, ports, "/api/software-product-statuses")
	require.Equal(t, http.StatusOK, resp.status)

	var body statusBody
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	return body
}

// The interval of the refresh is at most 4 minutes (root config), so each refresh here stands for one interval.
func TestAStatusChangeOnTheRegisterShowsAfterTheNextRefresh(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	store := statusstore.New(pool, nil)
	register := &registerStatuses{statuses: map[string]string{"software-1": "INACTIVE"}, etag: `"v1"`}
	refresh := app.Statuses{Source: register, Store: store}
	ports := console.Ports{Statuses: store}

	require.NoError(t, refresh.Refresh(t.Context()))

	first := readStatuses(t, ports)
	require.Len(t, first.Data, 1)
	assert.Equal(t, "INACTIVE", first.Data[0].Status)

	require.NoError(t, refresh.Refresh(t.Context()))

	notModified := readStatuses(t, ports)
	assert.True(t, notModified.ReceivedAt.After(first.ReceivedAt), "a response with no change also changes the time")

	register.statuses, register.etag = map[string]string{"software-1": "ACTIVE"}, `"v2"`

	require.NoError(t, refresh.Refresh(t.Context()))

	changed := readStatuses(t, ports)
	require.Len(t, changed.Data, 1)
	assert.Equal(t, "ACTIVE", changed.Data[0].Status)
	assert.True(t, changed.ReceivedAt.After(notModified.ReceivedAt))
}
