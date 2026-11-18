// Package root is the composition root of the register binary.
package root

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
)

const (
	serviceName  = "register"
	internalPort = 8080
	apiPort      = 8081
	securePort   = 8083
	// With app.CRLRenewalAge, less than 24 hours (ACCC CP, section 4.9.7). The job compares wall-clock times, so the
	// first run after a suspend of the host renews the CRL.
	crlRenewalInterval = 10 * time.Minute
	// With the server shutdown, less than 10 seconds: the default wait before a container stop sends SIGKILL.
	telemetryShutdownTimeout = 3 * time.Second
)

// Run runs the subcommand of the arguments, which are the command-line arguments with no program name.
func Run(ctx context.Context, args []string) error {
	shutdownTelemetry, err := telemetry.Setup(ctx, serviceName, os.Stdout)
	if err != nil {
		return fmt.Errorf("set up telemetry: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), telemetryShutdownTimeout)
		defer cancel()

		shutdownTelemetry(shutdownCtx)
	}()

	pool, err := database.Open(ctx, "")
	if err != nil {
		return fmt.Errorf("open the database: %w", err)
	}
	defer pool.Close()

	if m, found := modeOf(args); found {
		return serve(ctx, pool, m)
	}

	switch {
	case len(args) == 1 && args[0] == "migrate":
		return telemetry.RunSubcommand(ctx, "migrate", func(ctx context.Context) error {
			return migrate(ctx, pool)
		})
	case args[0] == "seed":
		return telemetry.RunSubcommand(ctx, "seed", func(ctx context.Context) error {
			return seed(ctx, pool, args[1:])
		})
	default:
		return fmt.Errorf(
			"unknown arguments %q; usage: register [serve | api | worker | migrate | seed [-if-empty] <file.json>]",
			args,
		)
	}
}

// mode gives the parts of the Register that one process runs.
type mode struct{ api, workers bool }

func modeOf(args []string) (mode, bool) {
	if len(args) == 0 {
		args = []string{"serve"}
	}

	if len(args) != 1 {
		return mode{}, false
	}

	m, found := map[string]mode{
		"serve":  {api: true, workers: true},
		"api":    {api: true},
		"worker": {workers: true},
	}[args[0]]

	return m, found
}
