// Package lifecycle controls the start and stop of a service process.
package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Run calls fn with a context that is cancelled when the process receives
// SIGINT or SIGTERM, or when ctx is cancelled. After the first signal, Run
// restores the default signal behaviour, so a second signal stops the
// process immediately.
func Run(ctx context.Context, fn func(context.Context) error) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	context.AfterFunc(ctx, stop)

	return fn(ctx)
}
