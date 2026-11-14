package telemetry

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

const redacted = "REDACTED"

func NewHTTPHandler(mux *http.ServeMux) http.Handler {
	return newHTTPHandler(mux, otel.GetTracerProvider(), otel.GetMeterProvider())
}

func newHTTPHandler(mux *http.ServeMux, tp trace.TracerProvider, mp metric.MeterProvider) http.Handler {
	instrumented := otelhttp.NewHandler(mux, "", otelhttp.WithTracerProvider(tp), otelhttp.WithMeterProvider(mp))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// otelhttp reads r.Pattern when it starts the span, before mux sets it. The copy keeps the caller's request.
		r = r.WithContext(r.Context())
		_, r.Pattern = mux.Handler(r)
		instrumented.ServeHTTP(w, r)
	})
}

// urlRedactor also replaces a query string, which can contain personal data, because a later otelhttp can record it.
type urlRedactor struct{}

func (urlRedactor) OnStart(_ context.Context, s sdktrace.ReadWriteSpan) {
	path, hasPath, hasQuery := redacted, false, false

	for _, a := range s.Attributes() {
		switch a.Key {
		case semconv.URLPathKey:
			hasPath = true
		case semconv.URLQueryKey:
			hasQuery = true
		case semconv.HTTPRouteKey:
			path = a.Value.AsString()
		default:
		}
	}

	if hasPath {
		s.SetAttributes(semconv.URLPath(path))
	}

	if hasQuery {
		s.SetAttributes(semconv.URLQuery(redacted))
	}
}

func (urlRedactor) OnEnd(sdktrace.ReadOnlySpan)      {}
func (urlRedactor) Shutdown(context.Context) error   { return nil }
func (urlRedactor) ForceFlush(context.Context) error { return nil }

func metricViews() []sdkmetric.View {
	return []sdkmetric.View{sdkmetric.NewView(
		sdkmetric.Instrument{Name: "http.server.*"},
		sdkmetric.Stream{AttributeFilter: attribute.NewDenyKeysFilter(semconv.ServerAddressKey, semconv.ServerPortKey)},
	)}
}
