// Package statusjobs refreshes the statuses of the Register as a periodic River job.
package statusjobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type Refresher interface {
	Refresh(ctx context.Context) error
}

type Args struct{}

func (Args) Kind() string { return "register_statuses" }

// InsertOpts has no retry and no unique completed state, so that the next interval runs the job again.
func (Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: 1, UniqueOpts: river.UniqueOpts{ByArgs: true, ByState: []rivertype.JobState{
		rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning,
		rivertype.JobStateRetryable, rivertype.JobStateScheduled,
	}}}
}

func PeriodicJob(interval time.Duration) *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(interval),
		func() (river.JobArgs, *river.InsertOpts) { return Args{}, nil },
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}

type Worker struct {
	river.WorkerDefaults[Args]

	Statuses Refresher
}

func (w Worker) Work(ctx context.Context, _ *river.Job[Args]) error {
	if err := w.Statuses.Refresh(ctx); err != nil {
		// River keeps the text in the job and in telemetry, and an error can contain a part of the response.
		slog.ErrorContext(ctx, "the statuses were not refreshed", "error.type", fmt.Sprintf("%T", err))

		return errors.New("the statuses were not refreshed")
	}

	return nil
}
