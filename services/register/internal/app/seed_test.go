package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type memoryStore struct {
	appends []eventsourcing.Append
}

func (s *memoryStore) Read(_ context.Context, streamType, streamID string) ([]eventsourcing.RecordedEvent, error) {
	var events []eventsourcing.RecordedEvent

	for _, a := range s.appends {
		if a.StreamType == streamType && a.StreamID == streamID {
			for _, e := range a.Events {
				events = append(
					events,
					eventsourcing.RecordedEvent{EventData: e, StreamVersion: int64(len(events) + 1)},
				)
			}
		}
	}

	return events, nil
}

func (s *memoryStore) Append(_ context.Context, appends ...eventsourcing.Append) error {
	s.appends = append(s.appends, appends...)
	return nil
}

func (s *memoryStore) IsEmpty(context.Context) (bool, error) {
	return len(s.appends) == 0, nil
}

func participants() domain.Participants {
	return domain.Participants{
		DataHolders: []domain.DataHolder{
			{
				LegalEntityID:   "dh-1",
				LegalEntityName: "Holder",
				LogoURI:         "https://holder.example/logo.png",
				Status:          domain.DataHolderActive,
			},
		},
		DataRecipients: []domain.DataRecipient{
			{
				LegalEntityID:       "dr-1",
				LegalEntityName:     "Recipient",
				AccreditationNumber: "ADR1",
				AccreditationLevel:  domain.Sponsored,
				LogoURI:             "https://recipient.example/logo.png",
				Status:              domain.DataRecipientActive,
			},
		},
	}
}

func TestSeedAppendsOneImportEventForEachAggregate(t *testing.T) {
	store := &memoryStore{}

	require.NoError(t, app.Seed(t.Context(), store, participants()))

	require.Len(t, store.appends, 2)
	assert.Equal(t, "data-holder", store.appends[0].StreamType)
	assert.Equal(t, "dh-1", store.appends[0].StreamID)
	assert.Equal(t, "DataHolderImported", store.appends[0].Events[0].EventType)
	assert.Equal(t, "data-recipient", store.appends[1].StreamType)
	assert.Equal(t, "dr-1", store.appends[1].StreamID)
	assert.Equal(t, "DataRecipientImported", store.appends[1].Events[0].EventType)

	for _, a := range store.appends {
		assert.Equal(t, int64(0), a.ExpectedVersion)
		assert.Len(t, a.Events, 1)
	}
}

func TestSeededParticipantsLoad(t *testing.T) {
	store := &memoryStore{}
	p := participants()
	require.NoError(t, app.Seed(t.Context(), store, p))

	state, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")

	require.NoError(t, err)
	assert.Equal(t, domain.DataRecipientState{Imported: true, DataRecipient: p.DataRecipients[0]}, state)
}

func TestSeedFailsWhenTheStoreIsNotEmpty(t *testing.T) {
	store := &memoryStore{appends: []eventsourcing.Append{{StreamType: "data-holder", StreamID: "dh-9"}}}

	err := app.Seed(t.Context(), store, participants())

	require.ErrorIs(t, err, app.ErrNotEmpty)
	assert.Len(t, store.appends, 1)
}

func TestSeedGivesAllValidationErrorsAndAppendsNothing(t *testing.T) {
	store := &memoryStore{}
	p := participants()
	p.DataHolders[0].LogoURI = ""
	p.DataRecipients[0].LegalEntityID = "dh-1"

	err := app.Seed(t.Context(), store, p)

	require.EqualError(t, err, `participants are not valid:
ID "dh-1" is not unique: it is in data holder "dh-1" and data recipient "dh-1"
data holder "dh-1": logoUri is required`)
	assert.Empty(t, store.appends)
}

func TestSeedValidatesBeforeItChecksTheStore(t *testing.T) {
	store := &memoryStore{appends: []eventsourcing.Append{{StreamType: "data-holder", StreamID: "dh-9"}}}
	p := participants()
	p.DataHolders[0].LogoURI = ""

	err := app.Seed(t.Context(), store, p)

	require.NotErrorIs(t, err, app.ErrNotEmpty)
	assert.ErrorContains(t, err, "logoUri is required")
}
