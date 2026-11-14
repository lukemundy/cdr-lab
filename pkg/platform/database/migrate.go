package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// MigrateOnStartEnv is the environment variable that tells a service to migrate when it starts.
const MigrateOnStartEnv = "DATABASE_MIGRATE_ON_START"

// MigrateOnStart reports whether MigrateOnStartEnv is true; an empty value is false.
func MigrateOnStart() (bool, error) {
	v := os.Getenv(MigrateOnStartEnv)
	if v == "" {
		return false, nil
	}

	on, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s: %w", MigrateOnStartEnv, err)
	}

	return on, nil
}

// Migrate applies the goose SQL migrations in the root of migrations, under an advisory lock.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrations fs.FS) error {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations, goose.WithSessionLocker(locker))
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}

	if len(results) > 0 {
		slog.InfoContext(ctx, "database migrations applied")
	}

	return nil
}
