// Package removaljobs cleans up a REMOVED software product as a River job.
package removaljobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

type Remover interface {
	RemoveSoftwareProduct(ctx context.Context, softwareID string, removedAt time.Time) error
}

type Args struct {
	SoftwareProductID string `json:"software_product_id"`
	// RemovedAt is the time at which the data holder got the status, which is the time of the end.
	RemovedAt time.Time `json:"removed_at"`
}

func (Args) Kind() string { return "removed_software_product" }

func (Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

func Hook(inserter *river.Client[pgx.Tx]) func(context.Context, pgx.Tx, string, time.Time) error {
	return func(ctx context.Context, tx pgx.Tx, softwareProductID string, removedAt time.Time) error {
		_, err := inserter.InsertTx(ctx, tx, Args{SoftwareProductID: softwareProductID, RemovedAt: removedAt}, nil)

		return err
	}
}

type Worker struct {
	river.WorkerDefaults[Args]

	Products Remover
}

func (w Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	if err := w.Products.RemoveSoftwareProduct(ctx, job.Args.SoftwareProductID, job.Args.RemovedAt); err != nil {
		// River keeps the text in the job and in telemetry, and a database error can contain query arguments.
		slog.ErrorContext(ctx, "removed software product job failed", "error.type", fmt.Sprintf("%T", err))

		return errors.New("internal error")
	}

	return nil
}
