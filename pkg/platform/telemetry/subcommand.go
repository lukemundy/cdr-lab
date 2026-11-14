package telemetry

import (
	"context"
	"os"
	"path/filepath"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// RunSubcommand runs fn in a span and records exit code 1 when fn fails, so the caller must then exit with 1.
func RunSubcommand(ctx context.Context, subcommand string, fn func(context.Context) error) error {
	return runSubcommand(ctx, otel.GetTracerProvider(), os.Args[0], subcommand, fn)
}

func runSubcommand(
	ctx context.Context,
	tp trace.TracerProvider,
	executable, subcommand string,
	fn func(context.Context) error,
) error {
	name := filepath.Base(executable)

	ctx, span := tp.Tracer(scope).Start(ctx, name+" "+subcommand,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(semconv.ProcessExecutableName(name), semconv.ProcessPID(os.Getpid())))
	defer span.End()

	err := fn(ctx)
	if err != nil {
		// No description: errors can contain query arguments.
		span.SetStatus(codes.Error, "")
		span.SetAttributes(semconv.ProcessExitCode(1), semconv.ErrorType(err))

		return err
	}

	span.SetAttributes(semconv.ProcessExitCode(0))

	return nil
}
