package database

import (
	"os"
	"testing"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const secret = "secret-argument-value"

// query runs sql with arg, in a recorded parent span, and returns the spans.
func query(t *testing.T, sql string, arg any) (tracetest.SpanStubs, error) {
	t.Helper()

	if os.Getenv("PGHOST") == "" {
		t.Fatal("PGHOST is not set. Run mise run up, then mise run test.")
	}

	spans := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans))
	pool, err := open(t.Context(), "", tp, sdkmetric.NewMeterProvider())
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	ctx, parent := tp.Tracer("test").Start(t.Context(), "test")

	var out string

	queryErr := pool.QueryRow(ctx, sql, arg).Scan(&out)

	parent.End()

	return spans.GetSpans(), queryErr
}

// querySpan returns the span of the query. The span of its prepare has a statement name.
func querySpan(t *testing.T, spans tracetest.SpanStubs, sql string) tracetest.SpanStub {
	t.Helper()

	for _, s := range spans {
		attrs := make(map[attribute.Key]string, len(s.Attributes))
		for _, a := range s.Attributes {
			attrs[a.Key] = a.Value.String()
		}

		if _, prepare := attrs["pgx.prepare_stmt.name"]; !prepare && attrs["db.query.text"] == sql {
			return s
		}
	}

	require.FailNow(t, "no span has the query text", sql)

	return tracetest.SpanStub{}
}

func assertNoSecret(t *testing.T, spans tracetest.SpanStubs) {
	t.Helper()

	for _, s := range spans {
		assert.NotContains(t, s.Status.Description, secret, "status of span %s", s.Name)

		for _, a := range s.Attributes {
			assert.NotContains(t, a.Value.String(), secret, "attribute %s of span %s", a.Key, s.Name)
		}

		for _, e := range s.Events {
			for _, a := range e.Attributes {
				assert.NotContains(
					t,
					a.Value.String(),
					secret,
					"event %s attribute %s of span %s",
					e.Name,
					a.Key,
					s.Name,
				)
			}
		}
	}
}

func TestQuerySpanHasQueryTextButNoArguments(t *testing.T) {
	const sql = "SELECT $1::text"

	spans, err := query(t, sql, secret)
	require.NoError(t, err)

	span := querySpan(t, spans, sql)
	assert.Equal(t, trace.SpanKindClient, span.SpanKind)
	assertNoSecret(t, spans)
}

// PostgreSQL error messages and pgx encode errors can contain argument values.
func TestQueryErrorSpanHasSQLStateButNoArguments(t *testing.T) {
	tests := []struct {
		name, sql, sqlState string
		arg                 any
	}{
		{name: "PostgreSQL error", sql: "SELECT $1::uuid::text", arg: secret, sqlState: "22P02"},
		{name: "pgx encode error", sql: "SELECT $1::int::text", arg: struct{ Value string }{secret}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spans, err := query(t, tt.sql, tt.arg)
			require.ErrorContains(t, err, secret, "the caller gets the full error")

			span := querySpan(t, spans, tt.sql)
			assert.Equal(t, "Error", span.Status.Code.String())

			if tt.sqlState != "" {
				assert.Contains(t, span.Attributes, otelpgx.SQLStateKey.String(tt.sqlState))
			}

			assertNoSecret(t, spans)
		})
	}
}

func TestOpenDoesNotConnect(t *testing.T) {
	t.Setenv("PGHOST", "127.0.0.1")
	t.Setenv("PGPORT", "1")

	pool, err := Open(t.Context(), "")
	require.NoError(t, err)

	defer pool.Close()

	assert.Error(t, pool.Ping(t.Context()))
}

func TestOpenRejectsInvalidConnString(t *testing.T) {
	_, err := Open(t.Context(), "port=not-a-number")
	assert.Error(t, err)
}

var _ pgxpool.AcquireTracer = redactingTracer{}
