package telemetry

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

func runTraced(t *testing.T, fn func(context.Context) error) ([]tracetest.SpanStub, error) {
	t.Helper()

	spans := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans))

	err := runSubcommand(t.Context(), tp, "/usr/local/bin/register", "seed", fn)

	return spans.GetSpans(), err
}

func TestRunSubcommandRecordsACLISpan(t *testing.T) {
	spans, err := runTraced(t, func(context.Context) error { return nil })

	require.NoError(t, err)
	require.Len(t, spans, 1)
	span := spans[0]
	assert.Equal(t, "register seed", span.Name)
	assert.Equal(t, trace.SpanKindInternal, span.SpanKind)
	assert.Equal(t, codes.Unset, span.Status.Code)
	assert.Equal(t, map[string]string{
		string(semconv.ProcessExecutableNameKey): "register",
		string(semconv.ProcessPIDKey):            strconv.Itoa(os.Getpid()),
		string(semconv.ProcessExitCodeKey):       "0",
	}, attrs(span))
}

func TestRunSubcommandGivesTheSpanToTheSubcommand(t *testing.T) {
	var child trace.SpanContext

	spans, _ := runTraced(t, func(ctx context.Context) error {
		child = trace.SpanContextFromContext(ctx)
		return nil
	})

	require.Len(t, spans, 1)
	assert.Equal(t, spans[0].SpanContext, child)
}

func TestRunSubcommandRecordsAnErrorWithoutItsMessage(t *testing.T) {
	cmdErr := errors.New("secret-argument-value")

	spans, err := runTraced(t, func(context.Context) error { return cmdErr })

	require.ErrorIs(t, err, cmdErr)
	require.Len(t, spans, 1)
	span := spans[0]
	assert.Equal(t, codes.Error, span.Status.Code)
	assert.Empty(t, span.Status.Description)
	assert.Equal(t, "1", attrs(span)[string(semconv.ProcessExitCodeKey)])
	assert.Equal(t, "*errors.errorString", attrs(span)[string(semconv.ErrorTypeKey)])
	assert.Empty(t, span.Events)
}
