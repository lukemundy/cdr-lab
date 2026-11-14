package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportConfigured(t *testing.T) {
	tests := map[string]struct {
		env  map[string]string
		want bool
	}{
		"nothing set": {want: false},
		"common OTLP endpoint": {
			env:  map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://lgtm:4318"},
			want: true,
		},
		"signal OTLP endpoint": {
			env:  map[string]string{"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://lgtm:4318/v1/traces"},
			want: true,
		},
		"signal exporter": {env: map[string]string{"OTEL_TRACES_EXPORTER": "console"}, want: true},
		"other signal endpoint": {
			env:  map[string]string{"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT": "http://lgtm:4318/v1/logs"},
			want: false,
		},
		"other signal exporter": {env: map[string]string{"OTEL_METRICS_EXPORTER": "console"}, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for _, key := range []string{
				"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "OTEL_TRACES_EXPORTER",
				"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "OTEL_METRICS_EXPORTER",
			} {
				t.Setenv(key, tt.env[key])
			}

			assert.Equal(t, tt.want, exportConfigured("TRACES"))
		})
	}
}
