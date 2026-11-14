// Package databasetest gives each test its own PostgreSQL database.
package databasetest

import (
	"context"
	"crypto/rand"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
)

// New returns a pool for a new database on the PG* server, which it removes when the test ends.
func New(t testing.TB) *pgxpool.Pool {
	t.Helper()

	if os.Getenv("PGHOST") == "" {
		t.Fatal("PGHOST is not set. Run mise run up, then mise run test.")
	}

	name := "test_" + strings.ToLower(rand.Text())
	ident := pgx.Identifier{name}.Sanitize()
	admin(t, "CREATE DATABASE "+ident)

	pool, err := database.Open(t.Context(), "dbname="+name)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pool.Close()
		admin(t, "DROP DATABASE "+ident+" WITH (FORCE)")
	})

	return pool
}

// admin runs statement on the default database, without the test context, which is cancelled before cleanup.
func admin(t testing.TB, statement string) {
	t.Helper()

	ctx := context.WithoutCancel(t.Context())

	conn, err := pgx.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close(ctx) }()

	if _, err := conn.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
}

// AsRole returns a pool on the database of pool whose connections act as role (SET ROLE).
func AsRole(t testing.TB, pool *pgxpool.Pool, role string) *pgxpool.Pool {
	t.Helper()

	cfg := pool.Config()
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET ROLE "+pgx.Identifier{role}.Sanitize())
		return err
	}

	rolePool, err := pgxpool.NewWithConfig(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(rolePool.Close)

	return rolePool
}

const userObjects = `
FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname NOT IN ('pg_catalog', 'information_schema') AND n.nspname NOT LIKE 'pg\_toast%'
  AND n.nspname NOT LIKE 'pg\_temp\_%'`

var roleChecks = []struct {
	problem string
	query   string
}{
	{
		problem: "can log in, or has a privileged attribute (LOGIN, SUPERUSER, CREATEDB, CREATEROLE or BYPASSRLS)",
		query: `SELECT '' FROM pg_roles
WHERE rolname = $1 AND (rolcanlogin OR rolsuper OR rolcreatedb OR rolcreaterole OR rolbypassrls)`,
	},
	{
		problem: "owns ",
		query: `SELECT format('%I.%I', n.nspname, c.relname)` + userObjects + `
  AND c.relkind IN ('r', 'p', 'v', 'm', 'f') AND c.relowner = (SELECT oid FROM pg_roles WHERE rolname = $1)
ORDER BY 1`,
	},
	{
		problem: "has CREATE on schema ",
		query: `SELECT format('%I', nspname) FROM pg_namespace
WHERE nspname NOT IN ('pg_catalog', 'information_schema') AND nspname NOT LIKE 'pg\_toast%'
  AND nspname NOT LIKE 'pg\_temp\_%' AND has_schema_privilege($1, oid, 'CREATE')
ORDER BY 1`,
	},
}

// For a sequence, data access and append access are USAGE. Each privilege function is in a CASE on relkind, because
// PostgreSQL does not evaluate the conditions in a fixed sequence, and a sequence function fails on a table.
const relationAccessQuery = `SELECT format('%I.%I', n.nspname, c.relname),
  c.relname IN ('goose_db_version', 'goose_db_version_id_seq'),
  coalesce(format('%I.%I', n.nspname, c.relname) = ANY($2::text[]), false),
  CASE c.relkind WHEN 'S' THEN has_sequence_privilege($1, c.oid, 'USAGE')
    ELSE has_table_privilege($1, c.oid, 'SELECT') AND has_table_privilege($1, c.oid, 'INSERT')
      AND has_table_privilege($1, c.oid, 'UPDATE') AND has_table_privilege($1, c.oid, 'DELETE') END,
  CASE c.relkind WHEN 'S' THEN has_sequence_privilege($1, c.oid, 'USAGE')
    ELSE has_table_privilege($1, c.oid, 'SELECT') AND has_table_privilege($1, c.oid, 'INSERT') END,
  CASE c.relkind WHEN 'S' THEN false ELSE has_table_privilege($1, c.oid, 'UPDATE, DELETE, TRUNCATE') END,
  CASE c.relkind WHEN 'S' THEN has_sequence_privilege($1, c.oid, 'USAGE, SELECT, UPDATE')
    ELSE has_table_privilege($1, c.oid, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, MAINTAIN') END
` + userObjects + `
  AND c.relkind IN ('r', 'p', 'S')
ORDER BY 1`

type relationAccess struct {
	Name         string
	Goose        bool
	AppendOnly   bool
	DataAccess   bool
	AppendAccess bool
	ChangeAccess bool
	AnyAccess    bool
}

// ApplicationRoleProblems gives each way in which role is not an application role: data access to each table and
// sequence except those of goose, also to those that the pool user creates later, and no other access. An
// append-only table, for example "public.events", gives SELECT and INSERT only.
func ApplicationRoleProblems(t testing.TB, pool *pgxpool.Pool, role string, appendOnly ...string) []string {
	t.Helper()

	ctx := t.Context()

	var exists bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM pg_roles WHERE rolname = $1)", role).
		Scan(&exists); err != nil {
		t.Fatal(err)
	}

	if !exists {
		return []string{"role " + role + " does not exist"}
	}

	var problems []string

	for _, check := range roleChecks {
		objects, err := queryStrings(ctx, pool, check.query, role)
		if err != nil {
			t.Fatal(err)
		}

		for _, object := range objects {
			problems = append(problems, "role "+role+" "+check.problem+object)
		}
	}

	rows, err := pool.Query(ctx, relationAccessQuery, role, appendOnly)
	if err != nil {
		t.Fatal(err)
	}

	relations, err := pgx.CollectRows(rows, pgx.RowToStructByPos[relationAccess])
	if err != nil {
		t.Fatal(err)
	}

	for _, relation := range relations {
		if problem := relationProblem(relation); problem != "" {
			problems = append(problems, "role "+role+" "+problem+relation.Name)
		}
	}

	if !newTableAccess(t, pool, role) {
		problems = append(problems, "role "+role+" has no data access to a new table and its sequence: "+
			"set the default privileges")
	}

	return problems
}

func relationProblem(relation relationAccess) string {
	switch {
	case relation.Goose && relation.AnyAccess:
		return "has a privilege on "
	case relation.Goose:
		return ""
	case relation.AppendOnly && relation.ChangeAccess:
		return "can change or remove the rows of append-only table "
	case relation.AppendOnly && !relation.AppendAccess:
		return "has no SELECT and INSERT on append-only table "
	case !relation.AppendOnly && !relation.DataAccess:
		return "has no data access to "
	}

	return ""
}

func queryStrings(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]string, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// newTableAccess creates a table in a transaction that it rolls back, and reports whether role has data access to it.
func newTableAccess(t testing.TB, pool *pgxpool.Pool, role string) bool {
	t.Helper()

	ctx := t.Context()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "CREATE TABLE databasetest_probe (id bigint GENERATED ALWAYS AS IDENTITY)"); err != nil {
		t.Fatal(err)
	}

	var access bool
	if err := tx.QueryRow(ctx, `SELECT has_table_privilege($1, 'databasetest_probe', 'SELECT')
  AND has_table_privilege($1, 'databasetest_probe', 'INSERT')
  AND has_table_privilege($1, 'databasetest_probe', 'UPDATE')
  AND has_table_privilege($1, 'databasetest_probe', 'DELETE')
  AND has_sequence_privilege($1, 'databasetest_probe_id_seq', 'USAGE')`, role).Scan(&access); err != nil {
		t.Fatal(err)
	}

	return access
}
