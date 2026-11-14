// Package noticejobs sends the notices to data recipients as River jobs.
package noticejobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	maxAttempts = 3
	// The send has a timeout of 10 seconds, and the attempt also reads and writes the database.
	attemptTimeout = 20 * time.Second
)

// retryWaits are the waits after the first and the second attempt.
var retryWaits = [maxAttempts - 1]time.Duration{10 * time.Second, time.Minute}

type Notices interface {
	Attempt(ctx context.Context, arrangementID string, attempt int) (app.NoticeAttemptResult, error)
	RecordAttempt(ctx context.Context, arrangementID string, attempt domain.NoticeAttempt) error
	Record(ctx context.Context, arrangementID string, status domain.NoticeStatus, last domain.NoticeAttempt) error
}

type Args struct {
	ArrangementID string `json:"arrangement_id"`
}

func (Args) Kind() string { return "arrangement_notice" }

func (Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: maxAttempts, UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

// Projection inserts the job of a pending notice in the transaction of the event that ends the arrangement. A new
// projection of a stream inserts no second job, because the job is unique and a recorded notice is not pending.
func Projection(inserter *river.Client[pgx.Tx]) eventstore.Projection {
	return eventstore.Projection{
		StreamType: app.ArrangementType.StreamType,
		Project: func(ctx context.Context, tx pgx.Tx, id string, recorded []eventsourcing.RecordedEvent) error {
			state, _, err := app.ArrangementType.Fold(id, recorded)
			if err != nil || state.NoticeStatus() != domain.NoticePending {
				return err
			}

			_, err = inserter.InsertTx(ctx, tx, Args{ArrangementID: id}, nil)

			return err
		},
	}
}

type Worker struct {
	river.WorkerDefaults[Args]

	Notices Notices
}

func (w Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	id := job.Args.ArrangementID

	result, err := w.Notices.Attempt(ctx, id, job.Attempt)
	if errors.Is(err, domain.ErrNoNoticePending) {
		return nil
	}

	if err != nil {
		return internalError(ctx, err)
	}

	switch {
	case result.Outcome == domain.NoticeDelivered:
		if err := w.Notices.Record(ctx, id, domain.NoticeComplete, result.NoticeAttempt); err != nil {
			return internalError(ctx, err)
		}

		return nil
	case result.Outcome == domain.NoticeTryAgain && job.Attempt < job.MaxAttempts:
		if err := w.Notices.RecordAttempt(ctx, id, result.NoticeAttempt); err != nil {
			return internalError(ctx, err)
		}

		// River keeps the text in the job. It contains no value of the response.
		return errors.New(result.Result)
	default:
		if err := w.Notices.Record(ctx, id, domain.NoticeFailed, result.NoticeAttempt); err != nil {
			return internalError(ctx, err)
		}

		return river.JobCancel(fmt.Errorf("the notice failed: %s", result.Result))
	}
}

// internalError gives River an error with no text of err, because River keeps the text in the job and in telemetry,
// and a database error can contain query arguments.
func internalError(ctx context.Context, err error) error {
	slog.ErrorContext(ctx, "notice job failed", "error.type", fmt.Sprintf("%T", err))

	return errors.New("internal error")
}

func (Worker) NextRetry(job *river.Job[Args]) time.Time {
	return time.Now().Add(retryWaits[min(job.Attempt, len(retryWaits))-1])
}

func (Worker) Timeout(*river.Job[Args]) time.Duration {
	return attemptTimeout
}
