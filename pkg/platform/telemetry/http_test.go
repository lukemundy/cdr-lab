package telemetry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	route  = "/accounts/{accountId}"
	secret = "secret-account-id"
)

func serve(t *testing.T, target string) (tracetest.SpanStub, metricdata.ResourceMetrics) {
	t.Helper()

	spans := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(urlRedactor{}),
		sdktrace.WithSyncer(spans),
	)
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithView(metricViews()...))

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+route, func(http.ResponseWriter, *http.Request) {})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	req.Host = "attacker-controlled.example"
	newHTTPHandler(mux, tp, mp).ServeHTTP(httptest.NewRecorder(), req)

	got := spans.GetSpans()
	require.Len(t, got, 1)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))

	return got[0], rm
}

func attrs(span tracetest.SpanStub) map[string]string {
	m := make(map[string]string, len(span.Attributes))
	for _, a := range span.Attributes {
		m[string(a.Key)] = a.Value.String()
	}

	return m
}

func assertNoSecret(t *testing.T, span tracetest.SpanStub) {
	t.Helper()

	for key, value := range attrs(span) {
		assert.NotContains(t, value, secret, "attribute %s contains the raw URL", key)
	}

	if query, found := attrs(span)[string(semconv.URLQueryKey)]; found {
		assert.Equal(t, redacted, query)
	}
}

// Verifies: TELE-0005
func TestHTTPHandlerRecordsRouteNotRawPath(t *testing.T) {
	span, _ := serve(t, "/accounts/"+secret+"?page="+secret)

	assert.Equal(t, "GET "+route, span.Name)
	assert.Equal(t, route, attrs(span)[string(semconv.HTTPRouteKey)])
	assert.Equal(t, route, attrs(span)[string(semconv.URLPathKey)])
	assertNoSecret(t, span)
}

// Verifies: TELE-0006
func TestHTTPHandlerRedactsPathWithNoRoute(t *testing.T) {
	span, _ := serve(t, "/unknown/"+secret)

	assert.Equal(t, "GET", span.Name)
	assert.Equal(t, redacted, attrs(span)[string(semconv.URLPathKey)])
	assertNoSecret(t, span)
}

// otelhttp v0.72.0 does not record url.query, so the test gives the span the attribute itself.
// Verifies: TELE-0007
func TestTheRedactorReplacesTheQueryString(t *testing.T) {
	spans := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(urlRedactor{}), sdktrace.WithSyncer(spans))

	_, span := tp.Tracer("test").Start(t.Context(), "GET", trace.WithAttributes(semconv.URLQuery("q="+secret)))
	span.End()

	got := spans.GetSpans()
	require.Len(t, got, 1)
	assert.Equal(t, redacted, attrs(got[0])[string(semconv.URLQueryKey)])
}

func TestHTTPHandlerRecordsRequestDurationByRouteWithoutHostHeader(t *testing.T) {
	_, rm := serve(t, "/accounts/"+secret)

	var routes []string

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http.server.request.duration" {
				continue
			}

			for _, dp := range m.Data.(metricdata.Histogram[float64]).DataPoints {
				v, _ := dp.Attributes.Value(semconv.HTTPRouteKey)
				routes = append(routes, v.AsString())

				assert.False(t, dp.Attributes.HasValue(semconv.ServerAddressKey), "metric has server.address")
				assert.False(t, dp.Attributes.HasValue(semconv.ServerPortKey), "metric has server.port")
			}
		}
	}

	assert.Equal(t, []string{route}, routes)
}
