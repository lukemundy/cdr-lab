package root

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/seedfile"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

func seed(ctx context.Context, pool *pgxpool.Pool, args []string) error {
	flags := flag.NewFlagSet("seed", flag.ContinueOnError)

	ifEmpty := flags.Bool("if-empty", false, "exit with no error when the Register already has participants")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() != 1 {
		return errors.New("usage: register seed [-if-empty] <file.json>")
	}

	f, err := os.Open(flags.Arg(0))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	participants, err := seedfile.Read(f)
	if err != nil {
		return fmt.Errorf("%s: %w", flags.Arg(0), err)
	}

	if err := migrateOnStart(ctx, pool); err != nil {
		return err
	}

	err = app.Seed(ctx, newEventStore(pool), participants)
	if errors.Is(err, app.ErrNotEmpty) && *ifEmpty {
		slog.InfoContext(ctx, "participants not loaded: the Register already has participants")
		return nil
	}

	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "participants loaded")

	return nil
}

func newEventStore(pool *pgxpool.Pool) *eventstore.Store {
	return eventstore.New(pool, readmodel.Projections()...)
}

func migrateOnStart(ctx context.Context, pool *pgxpool.Pool) error {
	onStart, err := database.MigrateOnStart()
	if err != nil || !onStart {
		return err
	}

	return migrate(ctx, pool)
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := database.Migrate(ctx, pool, migrations.FS); err != nil {
		return err
	}

	// A migration can add a read model, which then has no rows for the streams that exist.
	if err := newEventStore(pool).ProjectAll(ctx); err != nil {
		return fmt.Errorf("project all streams: %w", err)
	}

	return nil
}
