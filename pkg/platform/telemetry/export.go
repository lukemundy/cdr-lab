package telemetry

import (
	"os"
	"strings"
)

const metricExportIntervalMillis = "10000"

// The signal names in the OpenTelemetry environment variables.
const (
	traces  = "TRACES"
	metrics = "METRICS"
	logs    = "LOGS"
)

func setDefaultMetricExportInterval() error {
	const key = "OTEL_METRIC_EXPORT_INTERVAL"
	if os.Getenv(key) != "" {
		return nil
	}

	return os.Setenv(key, metricExportIntervalMillis)
}

// exportConfigured is false when nothing is configured, because the OpenTelemetry
// default (OTLP to localhost) gives an error for each export.
func exportConfigured(signal string) bool {
	for _, key := range []string{
		"OTEL_" + signal + "_EXPORTER",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_" + signal + "_ENDPOINT",
	} {
		if os.Getenv(key) != "" {
			return true
		}
	}

	return false
}

func signalsWithNoExport() []string {
	var off []string

	for _, signal := range []string{traces, metrics, logs} {
		if !exportConfigured(signal) {
			off = append(off, strings.ToLower(signal))
		}
	}

	return off
}
