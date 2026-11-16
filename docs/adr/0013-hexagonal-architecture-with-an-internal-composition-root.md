# 0013. Use a hexagonal architecture for each service, with the composition root in an internal package

- Date: 2026-10-21
- Amended by [ADR-0024](0024-write-specifications-as-identified-single-requirements-split-by-area.md): the place and the structure of the specification of a service

## Context

This ADR supersedes [ADR 0005](0005-service-architecture.md). It keeps the hexagonal architecture of ADR 0005, and changes where the composition root of a binary is.

Each service has business rules, and each service also does I/O: HTTP APIs, PostgreSQL, calls to other services and background jobs. We want to test the business rules with no database and no network. Event sourcing also needs this separation: handle and apply are pure functions ([ADR 0003](0003-event-sourcing.md)).

The repository has more than one service, and people and agents write them. If each service has its own structure, each one is harder to read and to review. A structure that a tool can check stays correct.

ADR 0005 put the composition root of each binary in its `main` package. The composition root reads the configuration, selects the subcommand, makes the adapters and connects them to the use cases. As a service gets more adapters, its composition root gets larger. On 2026-10-21, the `main.go` of the data holder had 631 lines: 30 adapters, 4 listeners, 4 background workers, 5 subcommands and the configuration. Its tests (configuration, modes, and the restart of the projections) were also in `cmd/`, because they test `package main`. The composition root also made HTTP clients with the timeouts and redirect rules of external protocols. For example, the client that gets the Register statuses must not follow a redirect.

The Go guidance for a server project is to keep "the Go packages implementing the server's logic in the `internal` directory" and the commands in a `cmd` directory, where each command has a `main.go` ([Go: Organizing a Go module](https://go.dev/doc/modules/layout)). It does not say what the composition root is.

We compared two options:

1. Keep the composition root in `package main`, in more than one file.
   - This is the smallest change, and the layout of ADR 0005 stays correct.
   - `cmd/<binary>/` has many files, and the tests of the composition root stay there.
   - The linter cannot make a rule that applies only to the composition root, because the root and the entry point are one package.
2. Move the composition root to a package in `internal/`, and keep only the entry point in `main`.
   - `main` becomes approximately 15 lines, and the composition root is a normal package with its own tests.
   - Each service has one more package.
   - The linter can make a rule for the composition root, separate from `main`.

## Decision

This decision applies to services only. Each service uses a hexagonal (ports and adapters) architecture, with this layout:

```
services/<name>/
├── DESIGN.md            # specification and requirements; no implementation details
├── Dockerfile           # container image
├── cmd/<binary>/main.go # one entry point for each binary
└── internal/
    ├── root/            # the composition root
    ├── domain/          # entities and business rules
    ├── app/             # use cases and ports
    ├── adapter/<name>/  # one package for each adapter
    └── migrations/      # goose SQL migrations, if the service keeps state (ADR 0002)
```

- `domain` contains the entities and the business rules. It imports only the standard library, and no I/O packages (`context`, `os`, `net`, `database/sql`, `log`). It does not call `time.Now` or `math/rand`. The app layer gives it the time, random values and IDs. It contains no telemetry.
- `app` contains the use cases, and it defines the ports (interfaces) that the use cases consume. It does not import adapters.
- `adapter` has one package for each adapter. Driving adapters (HTTP handlers, River workers) call use cases. Driven adapters (repositories, HTTP clients) implement ports. An adapter does not import a different adapter.
- `root` is the composition root. It reads the configuration, selects the subcommand or mode, makes the adapters and connects them to the use cases. It contains no business rules, no HTTP handlers and no SQL. The tests of the configuration, the modes and the wiring are in `root`. Only `main` imports `root`.
- `cmd` has one directory for each binary of the service, and each directory contains only `main.go`. Most services have one binary, with the name of the service. A service can have more than one, and then each binary has its own composition root in `internal/root/<binary>/`. `main` starts the process lifecycle, calls the composition root with the command-line arguments, and logs a returned error. It does nothing else, so that it needs no tests.
- The composition root does not set the transport settings of a call: a timeout, a redirect rule or a TLS setting. The adapter, or the `pkg/` library, that makes the call sets them. The composition root gives only configuration values, for example a base URL. Thus a rule of an external protocol is next to the code that obeys it, and a test of that code can check it.
- golangci-lint enforces the import rules: one rule for `domain` in all services, and two rules for each service. The first rule permits only the composition root, and the tests of adapters, to import an adapter. It thus also keeps adapters out of `app` and out of other adapters. The second rule permits only `main` to import the composition root.
- Participant state is event sourced. [ADR 0003](0003-event-sourcing.md) gives the design and the terms. Handle and apply are in `domain`. The event store, the projections and the read model queries are adapters.

## Consequences

- Tests of `domain` need no database, no network and no clock.
- A storage technology or an external API can change in one adapter, with no change to `domain` or `app`.
- A small service has the same number of layers as a large one, and data must go through more types between the layers. We accept this for a structure that is the same in all services.
- The linter cannot give the import rules of a service for all services in one rule. Each new service needs its own two rules.
- `main` has no tests. A fault in it shows only when the binary starts, for example in the lab deployment or in CI.
- The composition root of a large service is still large. It can have more than one file, and the file names show its parts.
- A file count is not an import rule, so the linter does not check that `cmd/<binary>/` contains only `main.go`. Review checks it.
- An adapter that makes its own HTTP client has fewer settings that a test can change. Its tests use a test server that the client can reach, or a constructor parameter for the parts that a test must replace.
