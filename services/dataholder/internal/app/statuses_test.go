package app_test

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type fakeStatusSource struct {
	response app.StatusResponse
	err      error
	etags    []string
}

func (f *fakeStatusSource) SoftwareProductStatuses(_ context.Context, etag string) (app.StatusResponse, error) {
	f.etags = append(f.etags, etag)

	return f.response, f.err
}

type memoryStatusStore struct {
	statuses    map[string]domain.SoftwareProductStatus
	etag        string
	saves       int
	notModified int
}

func (m *memoryStatusStore) ETag(context.Context) (string, error) { return m.etag, nil }

func (m *memoryStatusStore) SaveNotModified(context.Context) error {
	m.notModified++

	return nil
}

func (m *memoryStatusStore) Save(
	_ context.Context,
	statuses map[string]domain.SoftwareProductStatus,
	etag string,
) error {
	if m.statuses == nil {
		m.statuses = map[string]domain.SoftwareProductStatus{}
	}

	maps.Copy(m.statuses, statuses)

	m.etag = etag
	m.saves++

	return nil
}

// Verifies: DH-0685
func TestARefreshSavesTheStatusesAndTheETagOfTheResponse(t *testing.T) {
	source := &fakeStatusSource{response: app.StatusResponse{
		Statuses: map[string]string{"software-1": "ACTIVE", "software-2": "REMOVED"}, ETag: `"v1"`,
	}}
	store := &memoryStatusStore{etag: `"v0"`}

	require.NoError(t, app.Statuses{Source: source, Store: store}.Refresh(t.Context()))

	assert.Equal(t, []string{`"v0"`}, source.etags, "the request has the ETag of the last response")
	assert.Equal(t, map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive, "software-2": domain.SoftwareProductRemoved,
	}, store.statuses)
	assert.Equal(t, `"v1"`, store.etag)
}

func TestANotModifiedResponseKeepsTheStatuses(t *testing.T) {
	source := &fakeStatusSource{response: app.StatusResponse{NotModified: true}}
	store := &memoryStatusStore{
		statuses: map[string]domain.SoftwareProductStatus{"software-1": domain.SoftwareProductActive}, etag: `"v1"`,
	}

	require.NoError(t, app.Statuses{Source: source, Store: store}.Refresh(t.Context()))

	assert.Zero(t, store.saves)
	assert.Equal(t, 1, store.notModified, "the time of the last response changes")
	assert.Equal(t, `"v1"`, store.etag)
}

// Verifies: DH-0687
func TestAStatusThatIsNotKnownKeepsThePreviousStatus(t *testing.T) {
	source := &fakeStatusSource{response: app.StatusResponse{
		Statuses: map[string]string{"software-1": "SUSPENDED", "software-2": "INACTIVE"}, ETag: `"v2"`,
	}}
	store := &memoryStatusStore{statuses: map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive, "software-3": domain.SoftwareProductActive,
	}}

	require.NoError(t, app.Statuses{Source: source, Store: store}.Refresh(t.Context()))

	assert.Equal(t, map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive,
		"software-2": domain.SoftwareProductInactive,
		"software-3": domain.SoftwareProductActive,
	}, store.statuses, "a status that is not known or not given keeps the previous status")
}

// Verifies: DH-0686
func TestARegisterThatDoesNotRespondKeepsTheStatuses(t *testing.T) {
	source := &fakeStatusSource{err: errors.New("no connection")}
	store := &memoryStatusStore{}

	err := app.Statuses{Source: source, Store: store}.Refresh(t.Context())

	require.ErrorIs(t, err, source.err)
	assert.Zero(t, store.saves)
	assert.Zero(t, store.notModified, "the time of the last response does not change")
}

// softwareStatuses gives the status of each software product, and fails for "software-failure".
type softwareStatuses map[string]domain.SoftwareProductStatus

func (s softwareStatuses) Status(
	_ context.Context,
	softwareProductID string,
) (domain.SoftwareProductStatus, bool, error) {
	if softwareProductID == "software-failure" {
		return "", false, errors.New("database failure")
	}

	status, found := s[softwareProductID]

	return status, found, nil
}

func clientStatuses() app.ClientStatuses {
	return app.ClientStatuses{
		Registrations: fakeRegistrations{
			"client-active":   {ClientID: "client-active", SoftwareID: "software-active"},
			"client-inactive": {ClientID: "client-inactive", SoftwareID: "software-inactive"},
			"client-new":      {ClientID: "client-new", SoftwareID: "software-new"},
			"client-failure":  {ClientID: "client-failure", SoftwareID: "software-failure"},
		},
		Statuses: softwareStatuses{
			"software-active":   domain.SoftwareProductActive,
			"software-inactive": domain.SoftwareProductInactive,
		},
	}
}

// Verifies: DH-0682, DH-0688
func TestAClientHasTheStatusOfTheSoftwareProductOfItsRegistration(t *testing.T) {
	tests := map[string]struct {
		want   domain.SoftwareProductStatus
		active bool
	}{
		"client-active":   {domain.SoftwareProductActive, true},
		"client-inactive": {domain.SoftwareProductInactive, false},
		"client-new":      {domain.SoftwareProductInactive, false},
		"client-unknown":  {domain.SoftwareProductInactive, false},
	}
	for clientID, tc := range tests {
		t.Run(clientID, func(t *testing.T) {
			status, err := clientStatuses().Status(t.Context(), clientID)
			require.NoError(t, err)
			assert.Equal(t, tc.want, status, "a software product with no status gets the responses for INACTIVE")

			active, err := clientStatuses().Active(t.Context(), clientID)
			require.NoError(t, err)
			assert.Equal(t, tc.active, active)
		})
	}
}

func TestAStatusThatCannotBeReadIsAnError(t *testing.T) {
	_, err := clientStatuses().Active(t.Context(), "client-failure")

	require.Error(t, err)
}
