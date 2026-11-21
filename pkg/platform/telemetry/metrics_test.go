package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// Verifies: TELE-0010
func TestTheRuntimeMetricsAreAFixedSet(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	require.NoError(t, startRuntimeMetrics(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))))

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))

	series := map[string]int{}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			series[m.Name] = dataPoints(m.Data)
		}
	}

	assert.Equal(t, map[string]int{
		"go.config.gogc":        1,
		"go.goroutine.count":    1,
		"go.memory.allocated":   1,
		"go.memory.allocations": 1,
		"go.memory.gc.goal":     1,
		"go.memory.used":        2,
		"go.processor.limit":    1,
	}, series)
}

func dataPoints(data metricdata.Aggregation) int {
	switch d := data.(type) {
	case metricdata.Sum[int64]:
		return len(d.DataPoints)
	case metricdata.Gauge[int64]:
		return len(d.DataPoints)
	default:
		return -1
	}
}
