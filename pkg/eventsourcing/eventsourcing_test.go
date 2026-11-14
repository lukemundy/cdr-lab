package eventsourcing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
)

func TestExecuteAppendsTheEventsOfANewStreamWithExpectedVersionZero(t *testing.T) {
	store := newMemoryStore()

	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", increment{By: 3}))

	require.Len(t, store.appends, 1)
	require.Len(t, store.appends[0], 1)
	a := store.appends[0][0]
	assert.Equal(t, "counter", a.StreamType)
	assert.Equal(t, "c-1", a.StreamID)
	assert.Equal(t, int64(0), a.ExpectedVersion)
	require.Len(t, a.Events, 1)
	assert.Equal(t, "Incremented", a.Events[0].EventType)
	assert.JSONEq(t, `{"by": 3}`, string(a.Events[0].Payload))
}

func TestExecuteHandlesTheCommandWithTheLoadedState(t *testing.T) {
	store := newMemoryStore()
	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", increment{By: 6}))

	err := newCounterType().Execute(t.Context(), store, "c-1", increment{By: 6})

	require.ErrorIs(t, err, errLimit)
	assert.Len(t, store.appends, 1, "a rejected command appends nothing")
}

func TestExecuteAppendsWithTheLoadedStreamVersion(t *testing.T) {
	store := newMemoryStore()
	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", increment{By: 1}))
	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", increment{By: 1}))

	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", resetCounter{}))

	assert.Equal(t, int64(2), store.appends[2][0].ExpectedVersion)
}

func TestExecuteAppendsNothingWhenHandleGivesNoEvents(t *testing.T) {
	store := newMemoryStore()

	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", resetCounter{}))

	assert.Empty(t, store.appends)
}

func TestExecuteReturnsTheStoreError(t *testing.T) {
	store := newMemoryStore()
	store.err = eventsourcing.ErrVersionConflict

	err := newCounterType().Execute(t.Context(), store, "c-1", increment{By: 1})

	assert.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
}

func TestLoadAppliesTheEventsOfTheStreamInOrder(t *testing.T) {
	store := newMemoryStore()
	for _, c := range []counterCommand{increment{By: 4}, resetCounter{}, increment{By: 2}} {
		require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", c))
	}

	state, version, err := newCounterType().Load(t.Context(), store, "c-1")

	require.NoError(t, err)
	assert.Equal(t, 2, state)
	assert.Equal(t, int64(3), version)
}

func TestLoadGivesTheInitialStateForAStreamThatDoesNotExist(t *testing.T) {
	counter := newCounterType()
	counter.Initial = 5

	state, version, err := counter.Load(t.Context(), newMemoryStore(), "c-1")

	require.NoError(t, err)
	assert.Equal(t, 5, state)
	assert.Equal(t, int64(0), version)
}

func TestLoadFailsForAnEventThatTheCodecCannotDecode(t *testing.T) {
	store := newMemoryStore()
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: "counter",
		StreamID:   "c-1",
		Events:     []eventsourcing.EventData{{EventType: "Unknown", SchemaVersion: 1, Payload: []byte(`{}`)}},
	}))

	_, _, err := newCounterType().Load(t.Context(), store, "c-1")

	assert.ErrorContains(t, err, "counter c-1 version 1")
}

func TestFoldAppliesRecordedEventsWithNoStore(t *testing.T) {
	recorded := []eventsourcing.RecordedEvent{
		{EventType: "Incremented", SchemaVersion: 1, Payload: []byte(`{"by": 4}`), StreamVersion: 1},
		{EventType: "Incremented", SchemaVersion: 1, Payload: []byte(`{"by": 2}`), StreamVersion: 2},
	}

	state, version, err := newCounterType().Fold("c-1", recorded)

	require.NoError(t, err)
	assert.Equal(t, 6, state)
	assert.Equal(t, int64(2), version)
}

func TestExecuteAtAppendsWhenTheStreamIsAtTheExpectedVersion(t *testing.T) {
	store := newMemoryStore()
	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", increment{By: 1}))

	require.NoError(t, newCounterType().ExecuteAt(t.Context(), store, "c-1", 1, increment{By: 1}))

	require.Len(t, store.appends, 2)
	assert.Equal(t, int64(1), store.appends[1][0].ExpectedVersion)
}

func TestExecuteAtRefusesADifferentVersionBeforeItHandlesTheCommand(t *testing.T) {
	store := newMemoryStore()
	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", increment{By: 1}))
	require.NoError(t, newCounterType().Execute(t.Context(), store, "c-1", increment{By: 1}))

	for _, version := range []int64{1, 3} {
		// The command also breaks the limit, so the error shows that Handle did not run.
		err := newCounterType().ExecuteAt(t.Context(), store, "c-1", version, increment{By: counterLimit})

		require.ErrorIs(t, err, eventsourcing.ErrVersionConflict, "version %d", version)
	}

	assert.Len(t, store.appends, 2)
}

func TestExecuteAtHandlesTheCommandOnTheInitialStateForAStreamThatDoesNotExist(t *testing.T) {
	store := newMemoryStore()

	err := newCounterType().ExecuteAt(t.Context(), store, "c-1", 4, increment{By: counterLimit + 1})

	require.ErrorIs(t, err, errLimit)
	assert.Empty(t, store.appends)
}
