# 0020. Let the migrations of a service create its roles, and let the operator create the users

- Date: 2026-10-24

## Context

[ADR 0002](0002-data-store.md) gives one PostgreSQL database for each service. It also gives a role model for later: an owner role that owns the objects and runs migrations, and an application role with data access only. Until now, the lab did not use this model:

- The Register database was the default database of the PostgreSQL image, and the Register connected as the superuser of the image.
- The data holder database came from an init script in the service directory, with a lab-only password. A file in a service directory can become part of a real deployment by mistake.
- No service defined a role or a permission in its own database.

Two parties know different facts. The service knows which objects it has, and which access the running service needs. The operator of a deployment knows the users, the passwords and the databases.

These PostgreSQL facts limit the options:

- Roles are global to the server, not to a database ([PostgreSQL: Database Roles](https://www.postgresql.org/docs/18/user-manag.html)). Two databases of the same service on one server see the same role names. PostgreSQL has no `CREATE ROLE IF NOT EXISTS`.
- A role with `CREATEROLE` can create roles. Since PostgreSQL 16, it can change only the roles that it has `ADMIN OPTION` on, and it gets that option on each role that it creates ([PostgreSQL: Role Attributes](https://www.postgresql.org/docs/18/role-attributes.html)).
- A user must exist before it owns a database or runs a migration. A role that a migration creates exists only after the migration. Thus the operator can make a user a member of that role only after the migrations.
- `GRANT ... ON ALL TABLES` changes only the objects that exist. `ALTER DEFAULT PRIVILEGES` gives the same access to the objects that the same role creates later ([PostgreSQL: ALTER DEFAULT PRIVILEGES](https://www.postgresql.org/docs/18/sql-alterdefaultprivileges.html)).
- `LOCK TABLE` in a mode stronger than `ROW EXCLUSIVE` needs `UPDATE`, `DELETE`, `TRUNCATE` or `MAINTAIN` ([PostgreSQL: LOCK](https://www.postgresql.org/docs/18/sql-lock.html)). The product store locks its table when it loads products. The event store locks the `events` table in `SHARE` mode only when it projects all streams again, so that no append changes a stream between the read and the projection. Only the `migrate` subcommand projects all streams.
- The events of a stream are the record of each change ([ADR 0003](0003-event-sourcing.md)). The event store only reads and appends them.
- `REINDEX` needs `MAINTAIN` on the table ([PostgreSQL: REINDEX](https://www.postgresql.org/docs/18/sql-reindex.html)). The River reindexer runs `REINDEX INDEX CONCURRENTLY` on the indexes of `river_job` ([ADR 0010](0010-background-jobs-with-river.md)).

The options for who creates the application role:

| | The migrations of the service | The operator, before the migrations |
|---|---|---|
| Where the role and its access are | In the service, with the objects that they apply to | Divided: the role in the deployment, the access in the service |
| Owner user | Needs `CREATEROLE` | Needs no role attribute |
| Membership of the service user | A step after the migrations | In the same step as the users |
| A new deployment | Knows only the role names | Must copy the role definitions of each service |

This ADR supersedes these parts of [ADR 0002](0002-data-store.md): the privileged lab user ("Now, the lab services use a privileged user"), the lab setting of `DATABASE_MIGRATE_ON_START` to `true`, and the consequence that a service must use two roles before it goes outside the lab. The other parts of ADR 0002 do not change.

## Decision

This decision applies to each service that has a database.

- The migrations of a service create its application role, with no login, and give it all the access that the running service needs, and no more. The access is data access (`SELECT`, `INSERT`, `UPDATE`, `DELETE`) to each table, except the event store tables, and `USAGE` on each sequence, of the service schema. The default privileges give the same access to each table and sequence that a later migration creates. A service that needs more access, for example `MAINTAIN` for River, grants it in a migration, with the reason.
- The application role has no access to the goose version table. Only the owner changes the migration state.
- The application role has only `SELECT` and `INSERT` on the event store tables. A running service cannot change or remove an event, also after a security problem in the service. The projection of all streams, with its table lock, runs as the owner in the `migrate` subcommand. Thus the table lock stays, and it needs no more access for the application role.
- The role name starts with the name of the service, for example `register_app`. A migration creates it only when it does not exist, because the role is global to the server, and the migrations of a second database of the service, or a test database, find it.
- The operator of a deployment creates each database, the users and their passwords. For each database, the operator creates an owner user, which owns the database and runs the migrations, and a service user, which the running service connects as. The owner user has `CREATEROLE`, so that the migrations can create the roles. After the migrations, the operator makes the service user a member of the application role.
- The operator also controls which users can connect to each database. A database gives no privileges to `PUBLIC`. Because the role names are global, two databases of one service share the application role. The `CONNECT` privilege of the service user is the boundary between the two databases.
- A service contains no user, no password and no database set-up. The lab set-up is in the lab directory of the repository, not in a service directory.
- The lab runs the `migrate` subcommand of each service in a one-shot container as the owner user, then makes the memberships, then starts the service as the service user. The lab does not set `DATABASE_MIGRATE_ON_START`, because a service user cannot migrate. The variable stays for a deployment that runs a service as its owner user.
- Tests connect as the superuser of the lab PostgreSQL. A test of the migrations of a service checks that its application role has the access of this decision, and no more.

## Consequences

- A security problem in a running service gives access to the data of its database only. The service cannot change the schema, the migration state, the roles or the events.
- A running service cannot project all streams again, because it cannot lock the `events` table. The attempt fails with a permission error. A change that must project all streams outside the `migrate` subcommand needs a different lock, for example a transaction-level advisory lock that each append also takes.
- A correction of an event is a new event. A person who must remove an event, for example for a legal reason, uses the owner role.
- The owner user can create roles, also roles that can log in. Its credentials need the same protection as a superuser password in a deployment where other services share the server.
- The lab has a step after the migrations that makes the memberships. The PostgreSQL image runs its init scripts only on an empty volume, thus a new database or user needs `mise run reset`. The membership step runs at each start.
- A new service with a database adds its database and two users to the lab set-up, its memberships to the lab membership step, and a migrate container to the lab deployment.
- A migration that adds an object in a different schema, or that needs more than data access, must also give the application role its access. The test of the migrations finds a table or a sequence with no access.
- Tests do not run as the application role, except where a test chooses it. A query that needs more than data access fails first in the lab deployment.
- Roles that the migrations create in test databases stay on the lab server after the test databases are removed. They give no access, because their databases do not exist.
- When a River reindex fails, River tries to remove the index that the failure leaves. Only the owner can remove an index, so a person must remove it.
