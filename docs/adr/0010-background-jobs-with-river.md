# 0010. Run background jobs with River, and enqueue each job in the transaction of its cause

- Date: 2026-10-17

## Context

Some work of a service must occur after a request, and must occur also when the service stops or the other party is not available. An example is a notice to a different participant that a rule requires. This work needs a record that stays after a restart, a limit of attempts, a time between attempts, and a timeout. It must also be safe when more than one instance of a service runs. An in-memory queue or a goroutine meets none of these after a restart.

[ADR 0002](0002-data-store.md) chose River for background jobs. River keeps its jobs in PostgreSQL. It gives these functions:

- A maximum of attempts for each job, a time to the next attempt for each worker (`NextRetry`), a timeout for each job, and an error that ends the job with no new attempt (`JobCancel`).
- Each instance gets jobs with `FOR UPDATE SKIP LOCKED`, so one job runs on one instance at a time.
- An insert in a transaction of the caller (`InsertTx`), so a job exists only when the change that causes it commits.
- Unique jobs (`UniqueOpts`), which stop a second insert of the same job while the first is not complete.
- A rescue of a job that was running when its process stopped. The job gets its next attempt after `RescueStuckJobsAfter`.

River has its own schema. It gives two methods to apply it:

| | `rivermigrate` in the code | SQL files from `river migrate-get`, applied by goose |
|---|---|---|
| Tools | Two migration tools, each with its own version table and lock | One tool, one version table and one lock |
| River upgrade | The upgrade applies the new River migrations | A person adds a goose file with the new River migrations |
| Review | The SQL is in the River module | The SQL is in the repository |

The River tool gives the SQL with `--exclude-version 1` for this use. Version 1 is the River migration table, which goose replaces.

A job row is not a permanent record. River deletes completed jobs after 24 hours, and cancelled and discarded jobs after 7 days, by default. The `errors` of a job contain the text of each error that its worker returns.

A service that uses event sourcing changes its state only when it appends events ([ADR 0003](0003-event-sourcing.md)). A job that the service inserts after the append can be lost when the process stops between the two. The event store calls each projection in the transaction of the append.

## Decision

- A service runs its background jobs with River, through `pkg/platform/jobs`. The package sets the telemetry, the logger, the queue, the timeout and the rescue time the same for each service. A stop of the service stops River at the same time as the HTTP servers, and cancels the running jobs. River keeps each cancelled job for its next attempt.
- A service with workers can run them in the process of its APIs, or in a separate process. It has three subcommands:
  - `serve` runs the APIs and the workers. With no subcommand, the service runs `serve`. The lab uses this model.
  - `api` runs only the APIs. It inserts jobs but does not work them.
  - `worker` runs only the workers, and serves only the health probes.

  A production deployment can run `api` and `worker` as separate deployments, and manage and scale each one separately.
- The River schema is in goose SQL files from `river migrate-get`, with no version 1, in the migrations of each service that uses River. A River upgrade that adds a migration adds a goose file. We do not edit these files. The reason is one migration tool and one `migrate` command for each service.
- A job that an event causes is inserted in the transaction of the append of that event, with `InsertTx` from a projection. Each such job is unique by its arguments. The projection inserts the job only when the state of the stream needs it, so a new projection of all streams does not start a job that ended.
- The events are the record of what a job did. When a job ends, its worker appends an event with the outcome and with the history of its attempts, if a rule needs the record. The job row is not a record.
- A worker returns errors that it classifies, for example `status 503` or `timeout`. An error does not contain a value from a request or a response. River keeps the text in the job row, and the telemetry middleware puts it in the span status.

## Consequences

- A job survives a restart, and more than one instance can run with no duplicate work, except after a rescue.
- A job can run more than one time: after a rescue, or when its process stops after the work and before River records the result. Each job must be safe to run again.
- A cancelled job keeps the number of its attempts. A stop does not give a job more attempts than its maximum.
- A River upgrade needs a check of the River changelog for new migrations, and a new goose file for each one.
- Each service that uses River has the River tables in its database.
- The River maintenance (leader election, cleaner, rescuer) runs in each instance that works jobs. A deployment that has only `api` instances inserts jobs that no process works.
- A `worker` instance is ready when it can connect to the database.
- A test of a job that waits between attempts controls the time in-process. A black-box test cannot wait for the time between attempts.
