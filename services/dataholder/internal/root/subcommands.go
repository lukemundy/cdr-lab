package root

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/branddir"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/noticejobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

const seedUsage = "usage: dataholder seed [-if-empty] [<file.json>]"

// seed loads the product catalogue of the brand, or the products of a file of the same format.
func seed(ctx context.Context, pool *pgxpool.Pool, args []string) error {
	flags := flag.NewFlagSet("seed", flag.ContinueOnError)

	ifEmpty := flags.Bool("if-empty", false, "exit with no error when the data holder already has products")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() > 1 {
		return errors.New(seedUsage)
	}

	brand, err := loadBrand()
	if err != nil {
		return err
	}

	products := brand.Products
	if flags.NArg() == 1 {
		if products, err = readProductFile(flags.Arg(0)); err != nil {
			return err
		}
	}

	if err := migrateOnStart(ctx, pool); err != nil {
		return err
	}

	err = app.Seed(ctx, productstore.New(pool), products)
	if errors.Is(err, app.ErrNotEmpty) && *ifEmpty {
		slog.InfoContext(ctx, "products not loaded: the data holder already has products")
		return nil
	}

	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "products loaded", "app.product.count", len(products))

	return nil
}

func readProductFile(name string) ([][]byte, error) {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	products, err := branddir.ReadProducts(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	return products, nil
}

func migrateOnStart(ctx context.Context, pool *pgxpool.Pool) error {
	onStart, err := database.MigrateOnStart()
	if err != nil || !onStart {
		return err
	}

	return migrate(ctx, pool)
}

// newEventStore needs the inserter of the notice jobs, which projects them in the transaction of their event.
func newEventStore(pool *pgxpool.Pool, inserter *river.Client[pgx.Tx]) *eventstore.Store {
	projections := append(registrationstore.Projections(), arrangementstore.Projections()...)

	return eventstore.New(pool, append(projections, noticejobs.Projection(inserter))...)
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := database.Migrate(ctx, pool, migrations.FS); err != nil {
		return err
	}

	inserter, err := jobs.NewInsertOnlyClient(pool)
	if err != nil {
		return fmt.Errorf("create the job inserter: %w", err)
	}

	// A migration can add a read model, which then has no rows for the streams that exist.
	if err := newEventStore(pool, inserter).ProjectAll(ctx); err != nil {
		return fmt.Errorf("project all streams: %w", err)
	}

	return nil
}
