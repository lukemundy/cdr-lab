// Command register runs the mock CDR Register.
//
// Usage:
//
//	register                                serve the Register
//	register migrate                        apply the database migrations, then stop
//	register seed [-if-empty] <file.json>   load a full set of participants into an empty Register, then stop
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lukemundy/cdr-lab/pkg/platform/lifecycle"
	"github.com/lukemundy/cdr-lab/services/register/internal/root"
)

func main() {
	if err := lifecycle.Run(
		context.Background(),
		func(ctx context.Context) error { return root.Run(ctx, os.Args[1:]) },
	); err != nil {
		slog.Error("register stopped with error", "error", err)
		os.Exit(1)
	}
}
