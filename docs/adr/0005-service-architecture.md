# 0005. Use a hexagonal architecture for each service

- Date: 2026-10-04
- Superseded by [ADR-0013](0013-hexagonal-architecture-with-an-internal-composition-root.md)

## Context

Each service has business rules, and each service also does I/O: HTTP APIs, PostgreSQL, calls to other services and background jobs.

We want to test the business rules with no database and no network. Event sourcing also needs this separation: handle and apply are pure functions ([ADR 0003](0003-event-sourcing.md)).

The repository has more than one service, and people and agents write them. If each service has its own structure, each one is harder to read and to review. A structure that a tool can check stays correct.

## Decision

This decision applies to services only. Each service uses a hexagonal (ports and adapters) architecture, with this layout:

```
services/<name>/
├── DESIGN.md            # specification and requirements; no implementation details
├── Dockerfile           # container image
├── cmd/<command>/       # one main package for each binary: composition root only
└── internal/
    ├── domain/          # entities and business rules
    ├── app/             # use cases and ports
    ├── adapter/<name>/  # one package for each adapter
    └── migrations/      # goose SQL migrations, if the service keeps state (ADR 0002)
```

- `domain` contains the entities and the business rules. It imports only the standard library, and no I/O packages (`context`, `os`, `net`, `database/sql`, `log`). It does not call `time.Now` or `math/rand`. The app layer gives it the time, random values and IDs. It contains no telemetry.
- `app` contains the use cases, and it defines the ports (interfaces) that the use cases consume. It does not import adapters.
- `adapter` has one package for each adapter. Driving adapters (HTTP handlers, River workers) call use cases. Driven adapters (repositories, HTTP clients) implement ports.
- `cmd` has one directory for each binary of the service. Most services have one binary, with the name of the service. A service can have more than one. Each `main` package is a composition root only: it reads the configuration, makes the adapters and connects them to the use cases.
- golangci-lint enforces the import rules: one rule for `domain` in all services, and one rule for the `app` layer of each service.
- Participant state is event sourced. [ADR 0003](0003-event-sourcing.md) gives the design and the terms. Handle and apply are in `domain`. The event store, the projections and the read model queries are adapters.

## Consequences

- Tests of `domain` need no database, no network and no clock.
- A storage technology or an external API can change in one adapter, with no change to `domain` or `app`.
- A small service has the same number of layers as a large one, and data must go through more types between the layers. We accept this for a structure that is the same in all services.
- The linter cannot give the `app` rule for all services in one rule. Each new service needs its own rule.
