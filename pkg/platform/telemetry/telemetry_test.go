package telemetry_test

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
)

func clearExportEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_TRACES_EXPORTER", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_METRICS_EXPORTER", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_LOGS_EXPORTER", "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		t.Setenv(key, "")
	}
}

func setup(t *testing.T) *bytes.Buffer {
	t.Helper()

	defaultLogger := slog.Default()

	t.Cleanup(func() { slog.SetDefault(defaultLogger) })

	var stdout bytes.Buffer

	shutdown, err := telemetry.Setup(t.Context(), "test", telemetry.LongRunning, &stdout)
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t.Context()) })

	return &stdout
}

func TestSetupWritesLogsToStdoutAsText(t *testing.T) {
	clearExportEnv(t)
	stdout := setup(t)

	slog.InfoContext(t.Context(), "service started", "server.port", 8080)

	assert.Contains(t, stdout.String(), "level=INFO")
	assert.Contains(t, stdout.String(), `msg="service started"`)
	assert.Contains(t, stdout.String(), "server.port=8080")
}

// Verifies: TELE-0002
func TestSetupWithNoExportConfiguredDropsTelemetryWithoutErrors(t *testing.T) {
	clearExportEnv(t)

	defaultLogger := slog.Default()

	t.Cleanup(func() { slog.SetDefault(defaultLogger) })

	var stdout bytes.Buffer

	shutdown, err := telemetry.Setup(t.Context(), "test", telemetry.LongRunning, &stdout)
	require.NoError(t, err)

	slog.InfoContext(t.Context(), "a log record")
	shutdown(t.Context())

	assert.Contains(t, stdout.String(), `msg="telemetry export is off"`)
	assert.NotContains(t, stdout.String(), "level=ERROR")
}

func TestSetupPropagatesTraceContextButNotBaggage(t *testing.T) {
	clearExportEnv(t)
	t.Setenv("OTEL_PROPAGATORS", "")

	setup(t)

	assert.ElementsMatch(t, []string{"traceparent", "tracestate"}, otel.GetTextMapPropagator().Fields())
}

func TestSetupDefaultsMetricExportIntervalTo10Seconds(t *testing.T) {
	clearExportEnv(t)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "")
	require.NoError(t, os.Unsetenv("OTEL_METRIC_EXPORT_INTERVAL"))

	setup(t)

	assert.Equal(t, "10000", os.Getenv("OTEL_METRIC_EXPORT_INTERVAL"))
}

func TestSetupKeepsMetricExportIntervalFromEnvironment(t *testing.T) {
	clearExportEnv(t)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "5000")

	setup(t)

	assert.Equal(t, "5000", os.Getenv("OTEL_METRIC_EXPORT_INTERVAL"))
}

// Verifies: TELE-0010
func TestSetupGivesAGlobalMeterProviderThatMakesNoMetrics(t *testing.T) {
	clearExportEnv(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "console")
	setup(t)

	assert.IsType(t, noop.MeterProvider{}, otel.GetMeterProvider())
	assert.NotEqual(t, noop.MeterProvider{}, telemetry.ProcessMeterProvider())
}

// Verifies: TELE-0011
func TestAOneShotProcessMakesNoMetrics(t *testing.T) {
	clearExportEnv(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "console")

	defaultLogger := slog.Default()

	t.Cleanup(func() { slog.SetDefault(defaultLogger) })

	shutdown, err := telemetry.Setup(t.Context(), "test", telemetry.OneShot, &bytes.Buffer{})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t.Context()) })

	assert.IsType(t, noop.MeterProvider{}, otel.GetMeterProvider())
	assert.IsType(t, noop.MeterProvider{}, telemetry.ProcessMeterProvider())
}
