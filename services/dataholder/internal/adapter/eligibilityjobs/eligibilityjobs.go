// Package eligibilityjobs ends, as a River job, the arrangements of a customer whose last open account closed, when
// the data holder stops between the close and the end.
package eligibilityjobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

// backstopDelay lets the request of the close end the arrangements first, so that the job and the request do not
// append to the same streams at the same time.
const backstopDelay = time.Minute

type Args struct {
	CustomerNumber string    `json:"customer_number"`
	ClosedAt       time.Time `json:"closed_at"`
}

func (Args) Kind() string { return "ineligible_customer" }

func (Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

// Hook inserts the job in the transaction of the close.
func Hook(inserter *river.Client[pgx.Tx]) func(context.Context, pgx.Tx, string, time.Time) error {
	return func(ctx context.Context, tx pgx.Tx, customerNumber string, closedAt time.Time) error {
		_, err := inserter.InsertTx(
			ctx, tx, Args{CustomerNumber: customerNumber, ClosedAt: closedAt},
			&river.InsertOpts{ScheduledAt: closedAt.Add(backstopDelay)},
		)

		return err
	}
}

type Worker struct {
	river.WorkerDefaults[Args]

	Arrangements app.ArrangementEnder
}

func (w Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	err := w.Arrangements.EndForIneligibleCustomer(ctx, job.Args.CustomerNumber, job.Args.ClosedAt)
	if err == nil {
		return nil
	}

	// River keeps the text in the job and in telemetry, and a database error can contain the customer number.
	slog.ErrorContext(ctx, "ineligible customer job failed", "error.type", fmt.Sprintf("%T", err))

	return errors.New("internal error")
}
