// Package crljobs renews the CRL of the mock CDR intermediate as a periodic River job.
package crljobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type Renewer interface {
	RenewCRL(ctx context.Context) error
}

type Args struct{}

func (Args) Kind() string { return "cdr_crl_renewal" }

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

	CRLs Renewer
}

func (w Worker) Work(ctx context.Context, _ *river.Job[Args]) error {
	if err := w.CRLs.RenewCRL(ctx); err != nil {
		// River keeps the text in the job and in telemetry, and a database error can contain connection details.
		slog.ErrorContext(ctx, "the CRL was not renewed", "error.type", fmt.Sprintf("%T", err))

		return errors.New("the CRL was not renewed")
	}

	return nil
}
