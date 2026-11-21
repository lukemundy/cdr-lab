// Package database connects services to PostgreSQL, and applies schema changes.
package database

import (
	"context"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
)

// Open returns a traced pool, configured from the PG* environment variables
// and then connString, that connects only when used.
func Open(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	return open(ctx, connString, otel.GetTracerProvider(), telemetry.ProcessMeterProvider())
}

func open(
	ctx context.Context,
	connString string,
	tp trace.TracerProvider,
	mp metric.MeterProvider,
) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	// Do not enable query parameters: they contain participant data and CDR data.
	cfg.ConnConfig.Tracer = redactingTracer{otelpgx.NewTracer(
		otelpgx.WithTracerProvider(tp),
		// mp is only for the pool statistics, because the spans give the duration of each query.
		otelpgx.WithMeterProvider(noop.NewMeterProvider()),
	)}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := otelpgx.RecordStats(pool, otelpgx.WithStatsMeterProvider(mp)); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
