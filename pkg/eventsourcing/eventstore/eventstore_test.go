package eventstore_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	_, err := pool.Exec(t.Context(), eventstore.Schema)
	require.NoError(t, err)

	return pool
}

func newStore(t *testing.T) (*eventstore.Store, *pgxpool.Pool) {
	t.Helper()
	pool := newPool(t)

	return eventstore.New(pool), pool
}

func event(eventType, payload string) eventsourcing.EventData {
	return eventsourcing.EventData{EventType: eventType, SchemaVersion: 1, Payload: []byte(payload)}
}

func appendTo(streamID string, expected int64, events ...eventsourcing.EventData) eventsourcing.Append {
	return eventsourcing.Append{StreamType: "counter", StreamID: streamID, ExpectedVersion: expected, Events: events}
}

func countRows(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()

	var n int
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT count(*) FROM events").Scan(&n))

	return n
}

func TestReadGivesTheAppendedEventsInOrder(t *testing.T) {
	store, _ := newStore(t)
	require.NoError(
		t,
		store.Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{"by": 1}`), event("Reset", `{}`))),
	)
	require.NoError(t, store.Append(t.Context(), appendTo("c-1", 2, event("Incremented", `{"by": 5}`))))

	got, err := store.Read(t.Context(), "counter", "c-1")

	require.NoError(t, err)
	require.Len(t, got, 3)

	for i, want := range []eventsourcing.EventData{
		event("Incremented", `{"by": 1}`),
		event("Reset", `{}`),
		event("Incremented", `{"by": 5}`),
	} {
		assert.Equal(t, want.EventType, got[i].EventType)
		assert.Equal(t, want.SchemaVersion, got[i].SchemaVersion)
		assert.JSONEq(t, string(want.Payload), string(got[i].Payload))
		assert.Equal(t, int64(i+1), got[i].StreamVersion)
		assert.False(t, got[i].RecordedAt.IsZero())
	}

	assert.Less(t, got[0].GlobalPosition, got[1].GlobalPosition)
	assert.Less(t, got[1].GlobalPosition, got[2].GlobalPosition)
}

func TestReadGivesNoEventsForAStreamThatDoesNotExist(t *testing.T) {
	store, _ := newStore(t)

	got, err := store.Read(t.Context(), "counter", "c-1")

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestStreamsWithTheSameIDAndADifferentStreamTypeAreSeparate(t *testing.T) {
	store, _ := newStore(t)
	require.NoError(t, store.Append(t.Context(), appendTo("x-1", 0, event("Incremented", `{"by": 1}`))))

	other := eventsourcing.Append{
		StreamType: "other",
		StreamID:   "x-1",
		Events:     []eventsourcing.EventData{event("Created", `{}`)},
	}
	require.NoError(t, store.Append(t.Context(), other))

	got, err := store.Read(t.Context(), "other", "x-1")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Created", got[0].EventType)
}

func TestAppendFailsWhenTheExpectedVersionIsNotTheStreamVersion(t *testing.T) {
	for name, expected := range map[string]int64{"lower": 0, "higher": 2} {
		t.Run(name, func(t *testing.T) {
			store, pool := newStore(t)
			require.NoError(t, store.Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{"by": 1}`))))

			err := store.Append(t.Context(), appendTo("c-1", expected, event("Reset", `{}`)))

			require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
			assert.Equal(t, 1, countRows(t, pool))
		})
	}
}

func TestAppendAddsNoEventsWhenOneStreamHasAConflict(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, store.Append(t.Context(), appendTo("c-2", 0, event("Incremented", `{"by": 1}`))))

	err := store.Append(t.Context(),
		appendTo("c-1", 0, event("Incremented", `{"by": 1}`)),
		appendTo("c-2", 0, event("Incremented", `{"by": 1}`)),
	)

	require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
	assert.Equal(t, 1, countRows(t, pool))
}

func TestOnlyOneOfTwoConcurrentAppendsWithTheSameExpectedVersionSucceeds(t *testing.T) {
	store, pool := newStore(t)

	const writers = 2

	errs := make([]error, writers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := range writers {
		wg.Go(func() {
			<-start

			errs[i] = store.Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{"by": 1}`)))
		})
	}

	close(start)
	wg.Wait()

	var conflicts int

	for _, err := range errs {
		if err != nil {
			require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)

			conflicts++
		}
	}

	assert.Equal(t, 1, conflicts)
	assert.Equal(t, 1, countRows(t, pool))
}

func TestAppendRecordsTheTraceContextInTheMetadata(t *testing.T) {
	store, pool := newStore(t)
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{
			0x0a,
			0xf7,
			0x65,
			0x19,
			0x16,
			0xcd,
			0x43,
			0xdd,
			0x84,
			0x48,
			0xeb,
			0x21,
			0x1c,
			0x80,
			0x31,
			0x9c,
		},
		SpanID:     trace.SpanID{0xb7, 0xad, 0x6b, 0x71, 0x69, 0x20, 0x33, 0x31},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(t.Context(), sc)

	require.NoError(t, store.Append(ctx, appendTo("c-1", 0, event("Incremented", `{"by": 1}`))))

	var metadata string
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT metadata::text FROM events").Scan(&metadata))
	assert.JSONEq(t, `{"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"}`, metadata)
}

func TestAppendRecordsEmptyMetadataWithNoTraceContext(t *testing.T) {
	store, pool := newStore(t)

	require.NoError(t, store.Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{"by": 1}`))))

	var metadata string
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT metadata::text FROM events").Scan(&metadata))
	assert.JSONEq(t, `{}`, metadata)
}

func TestIsEmptyIsTrueOnlyWithNoStreams(t *testing.T) {
	store, _ := newStore(t)

	empty, err := store.IsEmpty(t.Context())
	require.NoError(t, err)
	assert.True(t, empty)

	require.NoError(t, store.Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{"by": 1}`))))

	empty, err = store.IsEmpty(t.Context())
	require.NoError(t, err)
	assert.False(t, empty)
}

// projectedVersions records the stream versions that a projection gets,
// in a table, so that a test can see the transaction.
func projectedVersions(t *testing.T, pool *pgxpool.Pool, err error) eventstore.Projection {
	t.Helper()
	_, execErr := pool.Exec(t.Context(), "CREATE TABLE projected (stream_id text, stream_version bigint)")
	require.NoError(t, execErr)

	return eventstore.Projection{
		StreamType: "counter",
		Project: func(ctx context.Context, tx pgx.Tx, streamID string, recorded []eventsourcing.RecordedEvent) error {
			for _, r := range recorded {
				if _, err := tx.Exec(
					ctx,
					"INSERT INTO projected VALUES ($1, $2)",
					streamID,
					r.StreamVersion,
				); err != nil {
					return err
				}
			}

			return err
		},
	}
}

func projected(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	rows, err := pool.Query(t.Context(), "SELECT stream_id || '/' || stream_version FROM projected ORDER BY 1")
	require.NoError(t, err)
	got, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	return got
}

func TestAppendProjectsAllTheEventsOfEachChangedStream(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, projectedVersions(t, pool, nil))
	require.NoError(t, store.Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{"by": 1}`))))
	_, err := pool.Exec(t.Context(), "TRUNCATE projected")
	require.NoError(t, err)

	require.NoError(t, store.Append(t.Context(),
		appendTo("c-1", 1, event("Incremented", `{"by": 1}`)),
		appendTo("c-2", 0, event("Incremented", `{"by": 1}`)),
	))

	assert.Equal(t, []string{"c-1/1", "c-1/2", "c-2/1"}, projected(t, pool))
}

func TestAppendDoesNotProjectOtherStreamTypes(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, projectedVersions(t, pool, nil))

	other := eventsourcing.Append{
		StreamType: "other",
		StreamID:   "x-1",
		Events:     []eventsourcing.EventData{event("Created", `{}`)},
	}
	require.NoError(t, store.Append(t.Context(), other))

	assert.Empty(t, projected(t, pool))
}

func TestAppendAddsNothingWhenAProjectionFails(t *testing.T) {
	pool := newPool(t)

	errProjection := errors.New("projection failed")
	store := eventstore.New(pool, projectedVersions(t, pool, errProjection))

	err := store.Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{"by": 1}`)))

	require.ErrorIs(t, err, errProjection)
	assert.Equal(t, 0, countRows(t, pool))
	assert.Empty(t, projected(t, pool))
}

func TestProjectAllProjectsAllTheEventsOfEachStreamOfTheStreamType(t *testing.T) {
	pool := newPool(t)
	require.NoError(t, eventstore.New(pool).Append(t.Context(),
		appendTo("c-1", 0, event("Incremented", `{"by": 1}`), event("Incremented", `{"by": 1}`)),
		appendTo("c-2", 0, event("Incremented", `{"by": 1}`)),
		eventsourcing.Append{
			StreamType: "other",
			StreamID:   "x-1",
			Events:     []eventsourcing.EventData{event("Created", `{}`)},
		},
	))

	store := eventstore.New(pool, projectedVersions(t, pool, nil))

	require.NoError(t, store.ProjectAll(t.Context()))

	assert.Equal(t, []string{"c-1/1", "c-1/2", "c-2/1"}, projected(t, pool))
	assert.Equal(t, 4, countRows(t, pool))
}

func TestProjectAllChangesNoReadModelWhenAProjectionFails(t *testing.T) {
	pool := newPool(t)
	require.NoError(t, eventstore.New(pool).Append(t.Context(),
		appendTo("c-1", 0, event("Incremented", `{"by": 1}`)),
		appendTo("c-2", 0, event("Incremented", `{"by": 1}`)),
	))

	errProjection := errors.New("projection failed")
	store := eventstore.New(pool, projectedVersions(t, pool, errProjection))

	err := store.ProjectAll(t.Context())

	require.ErrorIs(t, err, errProjection)
	assert.Empty(t, projected(t, pool))
}

// replacedVersions replaces the projected rows of a stream with its versions, as a read model does. Its first call
// closes started, then waits for release.
func replacedVersions(t *testing.T, pool *pgxpool.Pool, started, release chan struct{}) eventstore.Projection {
	t.Helper()
	_, err := pool.Exec(t.Context(), "CREATE TABLE projected (stream_id text, stream_version bigint)")
	require.NoError(t, err)

	var called atomic.Bool

	return eventstore.Projection{
		StreamType: "counter",
		Project: func(ctx context.Context, tx pgx.Tx, streamID string, recorded []eventsourcing.RecordedEvent) error {
			if called.CompareAndSwap(false, true) {
				close(started)
				<-release
			}

			if _, err := tx.Exec(ctx, "DELETE FROM projected WHERE stream_id = $1", streamID); err != nil {
				return err
			}

			for _, r := range recorded {
				if _, err := tx.Exec(
					ctx,
					"INSERT INTO projected VALUES ($1, $2)",
					streamID,
					r.StreamVersion,
				); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func waitForALockWaitOnEvents(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	require.Eventually(t, func() bool {
		var waiting bool

		err := pool.QueryRow(t.Context(),
			"SELECT EXISTS (SELECT FROM pg_locks WHERE relation = 'events'::regclass AND NOT granted)").Scan(&waiting)

		return err == nil && waiting
	}, 5*time.Second, 10*time.Millisecond)
}

// Without the lock, the append commits while ProjectAll waits, and ProjectAll then replaces its read model rows with
// the older events.
func TestAnAppendWaitsUntilProjectAllCommitsSoThatProjectAllDoesNotLoseIt(t *testing.T) {
	pool := newPool(t)
	require.NoError(t, eventstore.New(pool).Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{}`))))

	started, release := make(chan struct{}), make(chan struct{})
	store := eventstore.New(pool, replacedVersions(t, pool, started, release))

	var wg sync.WaitGroup

	// The pool closes only after each transaction ends, also when the test fails before the release.
	releaseOnce := sync.OnceFunc(func() { close(release) })

	t.Cleanup(func() {
		releaseOnce()
		wg.Wait()
	})

	var projectAllErr, appendErr error

	wg.Go(func() { projectAllErr = store.ProjectAll(t.Context()) })
	<-started
	wg.Go(func() { appendErr = store.Append(t.Context(), appendTo("c-1", 1, event("Incremented", `{}`))) })

	waitForALockWaitOnEvents(t, pool)
	releaseOnce()
	wg.Wait()

	require.NoError(t, projectAllErr)
	require.NoError(t, appendErr)
	assert.Equal(t, []string{"c-1/1", "c-1/2"}, projected(t, pool))
}

func recordSpans(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()

	spans := tracetest.NewInMemoryExporter()
	previous := otel.GetTracerProvider()

	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans)))
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	return spans
}

// projectAllSpan gives the one span of ProjectAll; the recorder also gets the database spans.
func projectAllSpan(t *testing.T, spans *tracetest.InMemoryExporter) tracetest.SpanStub {
	t.Helper()

	var found []tracetest.SpanStub

	for _, span := range spans.GetSpans() {
		if span.Name == "project all streams" {
			found = append(found, span)
		}
	}

	require.Len(t, found, 1)

	return found[0]
}

func TestProjectAllMakesOneSpanWithTheNumberOfStreamsAndEvents(t *testing.T) {
	pool := newPool(t)
	require.NoError(t, eventstore.New(pool).Append(t.Context(),
		appendTo("c-1", 0, event("Incremented", `{"by": 1}`), event("Incremented", `{"by": 1}`)),
		appendTo("c-2", 0, event("Incremented", `{"by": 1}`)),
	))

	store := eventstore.New(pool, projectedVersions(t, pool, nil))
	spans := recordSpans(t)

	require.NoError(t, store.ProjectAll(t.Context()))

	got := projectAllSpan(t, spans)
	assert.Equal(t, codes.Unset, got.Status.Code)
	assert.ElementsMatch(t, []attribute.KeyValue{
		attribute.Int("app.projection.stream.count", 2),
		attribute.Int("app.projection.event.count", 3),
	}, got.Attributes)
}

func TestTheProjectAllSpanOfAFailureHasTheErrorTypeAndNoMessage(t *testing.T) {
	pool := newPool(t)
	require.NoError(t, eventstore.New(pool).Append(t.Context(), appendTo("c-1", 0, event("Incremented", `{}`))))

	store := eventstore.New(pool, projectedVersions(t, pool, errors.New("projection failed for secret-value")))
	spans := recordSpans(t)

	require.Error(t, store.ProjectAll(t.Context()))

	got := projectAllSpan(t, spans)
	assert.Equal(t, codes.Error, got.Status.Code)
	assert.Empty(t, got.Status.Description)
	assert.Equal(t, []attribute.KeyValue{attribute.String("error.type", "*errors.errorString")}, got.Attributes)
}
