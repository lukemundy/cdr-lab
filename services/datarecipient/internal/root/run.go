// Package root is the composition root of the datarecipient binary.
package root

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/migrations"
)

const (
	serviceName  = "datarecipient"
	internalPort = 8080
	// Each listener has the same last digit in each service: the web app has the digit of the online bank site.
	webAppPort = 8085
	// With the server shutdown, less than 10 seconds: the default wait before a container stop sends SIGKILL.
	telemetryShutdownTimeout = 3 * time.Second
)

// Run runs the subcommand of the arguments, which are the command-line arguments with no program name.
func Run(ctx context.Context, args []string) error {
	serving := len(args) == 0 || (len(args) == 1 && args[0] == "serve")

	process := telemetry.OneShot
	if serving {
		process = telemetry.LongRunning
	}

	shutdownTelemetry, err := telemetry.Setup(ctx, serviceName, process, os.Stdout)
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

	switch {
	case serving:
		return serve(ctx, pool)
	case len(args) == 1 && args[0] == "migrate":
		return telemetry.RunSubcommand(ctx, "migrate", func(ctx context.Context) error {
			return migrate(ctx, pool)
		})
	default:
		return fmt.Errorf("unknown arguments %q; usage: datarecipient [serve | migrate]", args)
	}
}

func serve(ctx context.Context, pool *pgxpool.Pool) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := migrateOnStart(ctx, pool); err != nil {
		return fmt.Errorf("apply the database migrations: %w", err)
	}

	listeners, err := newListeners(ctx, pool, cfg)
	if err != nil {
		return err
	}

	return httpserver.ListenAndServe(ctx, listeners)
}

func migrateOnStart(ctx context.Context, pool *pgxpool.Pool) error {
	onStart, err := database.MigrateOnStart()
	if err != nil || !onStart {
		return err
	}

	return migrate(ctx, pool)
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return database.Migrate(ctx, pool, migrations.FS)
}
