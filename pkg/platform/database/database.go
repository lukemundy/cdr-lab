// Package database connects services to PostgreSQL, and applies schema changes.
package database

import (
	"context"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Open returns a traced pool, configured from the PG* environment variables
// and then connString, that connects only when used.
func Open(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	return open(ctx, connString, otel.GetTracerProvider(), otel.GetMeterProvider())
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
		otelpgx.WithMeterProvider(mp),
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
