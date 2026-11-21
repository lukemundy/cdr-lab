// Package telemetry sets up OpenTelemetry traces, metrics and logs for a service.
package telemetry

import (
	"context"
	"io"
	"log/slog"
	"slices"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const scope = "github.com/lukemundy/cdr-lab/pkg/platform/telemetry"

type Process int

const (
	// LongRunning is a server or a worker. It sends the metrics of its process state.
	LongRunning Process = iota
	// OneShot stops when its work is done, for example migrate. It sends no metrics.
	OneShot
)

var processMeterProvider metric.MeterProvider = noop.NewMeterProvider()

// ProcessMeterProvider is the only provider that makes metrics, so that an instrumentation adds no series by default.
func ProcessMeterProvider() metric.MeterProvider {
	return processMeterProvider
}

func Setup(
	ctx context.Context,
	serviceName string,
	process Process,
	stdout io.Writer,
) (shutdown func(context.Context), err error) {
	var shutdowns []func(context.Context) error

	shutdown = func(ctx context.Context) {
		for _, fn := range slices.Backward(shutdowns) {
			if err := fn(ctx); err != nil {
				otel.Handle(err)
			}
		}
	}
	fail := func(err error) (func(context.Context), error) {
		shutdown(ctx)
		return nil, err
	}

	text := slog.NewTextHandler(stdout, nil)
	// Stdout only, because a failed log export must not send more logs to the exporter.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.New(text).ErrorContext(ctx, "telemetry error", "error", err)
	}))
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator(propagation.TraceContext{}))

	if err := setDefaultMetricExportInterval(); err != nil {
		return nil, err
	}

	res := newResource(ctx, serviceName)

	tp, err := newTracerProvider(ctx, res)
	if err != nil {
		return fail(err)
	}

	shutdowns = append(shutdowns, tp.Shutdown)
	otel.SetTracerProvider(tp)

	otel.SetMeterProvider(noop.NewMeterProvider())
	processMeterProvider = noop.NewMeterProvider()

	if process == LongRunning {
		mp, err := newMeterProvider(ctx, res)
		if err != nil {
			return fail(err)
		}

		shutdowns = append(shutdowns, mp.Shutdown)
		processMeterProvider = mp

		if err := startRuntimeMetrics(mp); err != nil {
			return fail(err)
		}
	}

	lp, err := newLoggerProvider(ctx, res)
	if err != nil {
		return fail(err)
	}

	shutdowns = append(shutdowns, lp.Shutdown)
	otel.SetLoggerProvider(lp)

	slog.SetDefault(slog.New(slog.NewMultiHandler(
		text,
		otelslog.NewHandler(scope, otelslog.WithLoggerProvider(lp)),
	)))

	if off := signalsWithNoExport(); len(off) > 0 {
		slog.InfoContext(
			ctx,
			"telemetry export is off",
			slog.Any("signals", off),
			slog.String("reason", "no exporter or OTLP endpoint is set"),
		)
	}

	return shutdown, nil
}

func newTracerProvider(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res), sdktrace.WithSpanProcessor(urlRedactor{})}
	if !exportConfigured(traces) {
		return sdktrace.NewTracerProvider(opts...), nil
	}

	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(append(opts, sdktrace.WithBatcher(exporter))...), nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	opts := []sdkmetric.Option{sdkmetric.WithResource(res)}
	if !exportConfigured(metrics) {
		return sdkmetric.NewMeterProvider(opts...), nil
	}

	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, err
	}

	return sdkmetric.NewMeterProvider(append(opts, sdkmetric.WithReader(reader))...), nil
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	opts := []sdklog.LoggerProviderOption{sdklog.WithResource(res)}
	if !exportConfigured(logs) {
		return sdklog.NewLoggerProvider(opts...), nil
	}

	exporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, err
	}

	return sdklog.NewLoggerProvider(append(opts, sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)))...), nil
}

func startRuntimeMetrics(mp metric.MeterProvider) error {
	return runtime.Start(runtime.WithMeterProvider(mp))
}
