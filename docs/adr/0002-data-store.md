# 0002. Use PostgreSQL, with goose for schema changes

- Date: 2026-09-28
- Superseded in part by [ADR-0020](0020-service-migrations-create-roles-and-the-operator-creates-users.md)

## Context

The Register must keep participant data, statuses and signing keys after a restart ([Register DESIGN.md: Non-functional requirements](../../services/register/DESIGN.md#non-functional-requirements)). Later services (data holders, data recipients) also keep state. The lab has no data store.

We chose PostgreSQL as the data store at the start of the project, but we did not record the decision until now. The reasons are:

- It is used very widely.
- It is the SQL database that the maintainer knows best.
- In the maintainer's judgement, it is the best SQL database that is available.

We chose River for background jobs for similar reasons: we know it, and it works well with PostgreSQL. River supports only PostgreSQL, so River depends on this decision. This decision does not depend on River.

We use the current stable major version of PostgreSQL, 18. Version 19 is a beta.

Each service that keeps state needs a connection, schema changes, telemetry and a readiness check. If each service does this in its own way, the behaviour is different in each service. Thus this code goes in `pkg/platform/database`.

### Schema change tools

We compared goose, golang-migrate and tern. Usage, from GitHub and pkg.go.dev on 2026-09-28:

| Tool | GitHub stars | pkg.go.dev known importers |
|---|---|---|
| golang-migrate/migrate v4.20.1 | 18,944 | 5,219 |
| pressly/goose v3.28.0 | 11,509 | 1,725 |
| jackc/tern v2.4.3 | 1,329 | 50 |

Dependencies, measured in a test module. The baseline imports pgx v5.11.0 (`pgxpool` and `stdlib`) and otelpgx v0.12.0. Each other row adds the tool to the baseline. "Module graph" is `go list -m all`. "Compiled" is `go list -deps` for the packages that the program imports. The binary is built with `CGO_ENABLED=0 -trimpath -ldflags='-s -w'`.

| Case | Module graph | Compiled modules | Compiled packages | Binary |
|---|---|---|---|---|
| Baseline | 30 | 14 | 58 | 6.6 MB |
| + goose (`goose`, `goose/lock`) | 98 | 18 | 74 | 7.0 MB |
| + golang-migrate (`database/pgx/v5`, `source/iofs`) | 215 | 16 | 67 | 6.8 MB |
| + tern (`migrate`) | 51 | 26 | 75 | 7.6 MB |

The module graph of golang-migrate is large because its `go.mod` has the dependencies of all its database drivers. It compiles fewer modules than goose. The dependencies do not make a big difference between goose and golang-migrate.

The difference is what occurs when a migration fails:

- goose runs each migration, and the insert into its version table, in one transaction (`provider_run.go`, `runIndividually`). When the migration fails, the two roll back together. The next run tries the same migration again. No person must do an action. The migrations before it stay applied.
- A goose migration with the `-- +goose NO TRANSACTION` annotation does not run in a transaction. `CREATE INDEX CONCURRENTLY` needs this. When such a migration fails, some of its changes can stay, and goose does not record the version. The next run executes the file again. Thus these migrations must be idempotent.
- golang-migrate writes the version with `dirty = true` before it runs a migration, and it sets `dirty = false` only after the migration succeeds (`migrate.go`, `runMigrations`). After a failure, each subsequent run stops with `ErrDirty` until a person runs `migrate force`. PostgreSQL runs a multi-statement query as one implicit transaction ([PostgreSQL: Multiple Statements in a Simple Query](https://www.postgresql.org/docs/18/protocol-flow.html#PROTOCOL-FLOW-MULTI-STATEMENT)). Thus the schema usually rolls back correctly, but the dirty flag stays. When a service migrates at startup, the service then restarts again and again until a person clears the flag.

### Test databases

testcontainers-go v0.44.0 (`modules/postgres`) adds 58 modules to the module graph, 38 compiled modules and 101 compiled packages to the baseline above. These are in test binaries only. Its main advantage is isolation between parallel test runs, for example from parallel agents in worktrees. Parallel runs are rare in this repository. A test database for each test, on the lab PostgreSQL, also gives isolation, and it adds no dependencies.

### One database or one schema for each service

| | One database for each service | One schema for each service, in one database |
|---|---|---|
| Isolation | Strong. "It is not possible to access more than one database per connection" ([PostgreSQL: Overview](https://www.postgresql.org/docs/18/manage-ag-overview.html)). A query cannot join the tables of two services. | Weak. "Unlike databases, schemas are not rigidly separated" ([PostgreSQL: Schemas](https://www.postgresql.org/docs/18/ddl-schemas.html)). Isolation depends on correct grants. |
| Default privileges | PUBLIC gets `CONNECT` and `TEMPORARY` on each new database ([PostgreSQL: Privileges](https://www.postgresql.org/docs/18/ddl-priv.html)). Each database needs a `REVOKE`. | PUBLIC gets no privileges on a new schema (same source). |
| Name resolution | The `public` schema. No configuration. | Each role needs a `search_path`, or each name needs its schema. The goose version table and River (`Config.Schema`) need configuration. |
| Extensions | One version for each service. | One version for all services. |
| Move a service to its own server | `pg_dump` of one database. | `pg_dump -n` of one schema, if no object refers to another schema. |
| Create for a new service in the lab | The image init scripts run only on an empty volume. A new service needs `mise run reset`, or a provisioning step that is idempotent. PostgreSQL has no `CREATE DATABASE IF NOT EXISTS`. | The first migration of the service can do `CREATE SCHEMA IF NOT EXISTS`, if its role has `CREATE` on the database. |
| Operations | One connection for each service. | One `psql` session shows all services. |

The role model is the same for both: an owner role that owns the objects and runs migrations, and an application role with data access only.

## Decision

- One PostgreSQL server in the lab, with one database for each service. This agrees with the rule that each service can move to its own repository, and PostgreSQL enforces the boundary.
- Services connect with pgx and `pgxpool`. The connection configuration comes from the standard libpq environment variables (`PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`, `PGSSLMODE`), as the `OTEL_*` variables configure telemetry.
- otelpgx traces and measures database calls. It does not record query parameters. Parameters will contain participant data, and in data holders, CDR data.
- PostgreSQL error messages and pgx encode errors can contain argument values, for example `invalid input syntax for type uuid: "<value>"`. otelpgx records the error message in the span. Thus `pkg/platform/database` removes the message from query errors before otelpgx records them. The span keeps the SQLSTATE.
- goose applies schema changes. Each service embeds its migrations as SQL files. goose holds a PostgreSQL advisory lock while it migrates, so two instances cannot migrate at the same time.
- A service migrates at startup only when `DATABASE_MIGRATE_ON_START` is `true`. If the variable is not set, or is `false`, the service does not migrate, and an operator runs the `migrate` subcommand of the service binary. This supports an init container that runs the same image. The lab sets the variable to `true`.
- Tests of database adapters use the lab PostgreSQL. Each test gets a new database, which the test removes when it ends.
- Now, the lab services use a privileged user, and the password is in `compose.yaml`. It is a lab-only password.

## Consequences

- `mise run test` needs the lab PostgreSQL. Run `mise run up` first. CI starts PostgreSQL before the tests.
- The test user must be able to create databases.
- The image creates one database. Each more database comes from an init script, which the image runs only on an empty volume. Thus a lab that already has data needs a reset when a service gets a new database.
- Before a service goes outside the lab, it must use two roles: an owner role for migrations, and an application role with data access only. The `migrate` subcommand runs as the owner role.
- A migration with `-- +goose NO TRANSACTION` must be idempotent.
- The error that a query returns to the caller still contains the full message. Services must not log it as it is.
- The Register gets its first migration, the `DATABASE_MIGRATE_ON_START` variable and the `migrate` subcommand with its first table. goose does not start with no migrations.
