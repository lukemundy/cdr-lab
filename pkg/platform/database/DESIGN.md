# Database library

The database library connects a lab service to its PostgreSQL database, and it applies the migrations of the service. It traces the database calls with no query arguments and no error messages, because they can contain CDR data and personal information.

## Context

```mermaid
flowchart LR
  service["Lab service<br>[External system]"]
  library["Database library<br>[System]"]
  postgres["PostgreSQL<br>[External system]"]

  service -- "Opens its database and applies its migrations<br>Go calls" --> library
  library -- "Runs the queries and the migrations<br>PostgreSQL protocol" --> postgres

  classDef system fill:#1168bd,color:#fff,stroke:#0b4884
  classDef external fill:#999,color:#fff,stroke:#6b6b6b
  class library system
  class service,postgres external
```

## Library specifications

None.

## Goals

- One way for each lab service to connect to its database and to apply its migrations.
- No CDR data and no personal information in the database spans.

## Non-goals

- The migrations, the roles and the tables of a service. The service gives them ([ADR 0020][adr-0020]).
- The databases and the users of a deployment. The operator of the deployment gives them ([ADR 0020][adr-0020]).

## Connection

- <a id="dbase-0001"></a>**DBASE-0001** A service gets its database connection from the libpq environment variables. For example, `PGHOST` gives the server, and `PGDATABASE` gives the database. ([Judgement: one PostgreSQL database for each service (ADR 0002)](../../../docs/adr/0002-data-store.md))

## Migrations

- <a id="dbase-0002"></a>**DBASE-0002** A service applies its migrations at its start only when `DATABASE_MIGRATE_ON_START` is `true`. When the variable is not set, or is `false`, the service does not migrate. ([Judgement: one PostgreSQL database for each service (ADR 0002)](../../../docs/adr/0002-data-store.md); [Judgement: the migrations of a service create its roles (ADR 0020)](../../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md))
- <a id="dbase-0003"></a>**DBASE-0003** The migrations hold a PostgreSQL advisory lock while they run. For example, when two processes migrate the same database at the same time, the second process waits, and each migration applies one time. ([Judgement: one PostgreSQL database for each service (ADR 0002)](../../../docs/adr/0002-data-store.md))

## Telemetry

- <a id="dbase-0004"></a>**DBASE-0004** A database span has the SQL text with its placeholders, and no query arguments. ([Judgement: one PostgreSQL database for each service (ADR 0002)](../../../docs/adr/0002-data-store.md); [Judgement: OpenTelemetry for traces, metrics and logs (ADR 0006)](../../../docs/adr/0006-telemetry.md))
- <a id="dbase-0005"></a>**DBASE-0005** The span of a query that fails has the SQLSTATE of the error, and not its message. For example, a value that is not a UUID gives `22P02` in the span, with no value. ([Judgement: one PostgreSQL database for each service (ADR 0002)](../../../docs/adr/0002-data-store.md); [Judgement: OpenTelemetry for traces, metrics and logs (ADR 0006)](../../../docs/adr/0006-telemetry.md))

## Judgements

None.

## Withdrawn

None.

[adr-0020]: ../../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
