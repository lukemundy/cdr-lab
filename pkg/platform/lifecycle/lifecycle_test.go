package lifecycle_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/lifecycle"
)

func TestRunCancelsContextOnSIGTERM(t *testing.T) {
	err := lifecycle.Run(t.Context(), func(ctx context.Context) error {
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return errors.New("context not cancelled after SIGTERM")
		}
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want %v", err, context.Canceled)
	}
}

func TestRunReturnsErrorFromFunc(t *testing.T) {
	want := errors.New("boom")

	err := lifecycle.Run(t.Context(), func(context.Context) error { return want })

	if !errors.Is(err, want) {
		t.Fatalf("got error %v, want %v", err, want)
	}
}

func TestRunCancelsContextWhenParentIsCancelled(t *testing.T) {
	parent, cancel := context.WithCancel(t.Context())
	cancel()

	err := lifecycle.Run(parent, func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want %v", err, context.Canceled)
	}
}
