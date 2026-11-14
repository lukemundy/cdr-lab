// Package jobs runs the durable background jobs of a service with River.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"github.com/riverqueue/rivercontrib/otelriver"
)

const (
	// A job that needs more time sets its own Timeout.
	jobTimeout = 30 * time.Second
	// A job that runs at a hard stop of the process gets its next attempt after this time, which River requires to be
	// longer than jobTimeout.
	rescueStuckJobsAfter = time.Minute
	// A hard stop cancels each running job at once, so the client stops quickly.
	stopTimeout = 2 * time.Second
	maxWorkers  = 10
)

// NewWorkerClient needs the River schema in the database of pool.
func NewWorkerClient(
	pool *pgxpool.Pool,
	workers *river.Workers,
	periodicJobs ...*river.PeriodicJob,
) (*river.Client[pgx.Tx], error) {
	return river.NewClient(riverpgxv5.New(pool), &river.Config{
		JobTimeout:           jobTimeout,
		Logger:               slog.Default(),
		Middleware:           middleware(),
		PeriodicJobs:         periodicJobs,
		Queues:               map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: maxWorkers}},
		RescueStuckJobsAfter: rescueStuckJobsAfter,
		Workers:              workers,
	})
}

// NewInsertOnlyClient is for a process that works no jobs, or that must insert a job before its workers exist.
func NewInsertOnlyClient(pool *pgxpool.Pool) (*river.Client[pgx.Tx], error) {
	return river.NewClient(riverpgxv5.New(pool), &river.Config{
		Logger:     slog.Default(),
		Middleware: middleware(),
	})
}

func middleware() []rivertype.Middleware {
	return []rivertype.Middleware{otelriver.NewMiddleware(&otelriver.MiddlewareConfig{EnableTracePropagation: true})}
}

// Run gives client the context of serve, so that the two stop at the same time and their stop times do not add up.
func Run(ctx context.Context, client *river.Client[pgx.Tx], serve func(context.Context) error) error {
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("start the job client: %w", err)
	}

	err := serve(ctx)

	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stopTimeout)
	defer cancel()

	if stopErr := client.StopAndCancel(stopCtx); stopErr != nil {
		err = errors.Join(err, fmt.Errorf("stop the job client: %w", stopErr))
	}

	return err
}
