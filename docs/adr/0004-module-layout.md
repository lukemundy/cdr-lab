# 0004. Use one Go module for all components, in one repository

- Date: 2026-10-04

## Context

This repository contains more than one standalone component, and libraries that the components share. We need patterns for how to manage them in one repository: where each kind of component goes, what a component can import, and how components share code.

Without these patterns, components become coupled through imports. A change to one component can then break a different component, and no component can be understood or built on its own.

The principle is that each component is standalone. The purpose of the principle is to keep the code of each component contained. It is acceptable that a component needs some steps before it can move to its own repository.

We compared two module layouts:

1. One Go module for each component, with a Go workspace (`go.work`) that joins them for local development.
   - The module boundary keeps the dependencies of each component apart. A component can import only the modules that its `go.mod` names.
   - `go mod tidy` ignores `go.work` ([golang/go#50750](https://github.com/golang/go/issues/50750)). Thus a module that imports a shared library needs a `require` and a local `replace` directive for it.
   - Each `go.mod` must list the dependencies that the module needs for its build, with versions that are not lower than its dependencies require ([graph pruning](https://go.dev/ref/mod#graph-pruning)). Thus a version change in a shared library makes a change necessary in the `go.mod` of each module that imports the library. A tool that updates one `go.mod` does not make that change.
   - The workspace selects one version of a dependency for all modules, but a build with no workspace uses only the `go.mod` of one module. Thus two builds of the same component can use different versions of the same dependency.
2. One Go module for the repository.
   - There is one `go.mod`, so there is one list of dependency versions. A tool updates it in one step.
   - The module boundary does not keep components apart. Only `internal/` directories and review do this.
   - A component that leaves the repository needs a new `go.mod`.

In the two layouts, a component that leaves the repository must vendor the shared libraries that it uses.

## Decision

- The repository is one Go module, with one `go.mod` at its root. There is no Go workspace and no `replace` directive. The containment that the principle needs comes from the directory rules below, not from module boundaries, and one list of dependency versions is less work to keep correct.
- `services/` contains services: programs that are deployed and that run continuously. Each directory in `services/` is one service.
- All the code of a service is in `internal/`, so the compiler prevents an import of one service by another.
- `pkg/` contains shared libraries. Each directory in `pkg/` is one library.
- Services share code only through `pkg/` libraries. Code goes in `pkg/` only when two or more services need it. An exception is permitted for code that we know ahead of time the services will share, for example telemetry. `pkg/` never contains the business logic of one service.
- The `go.mod` gives the Go language version and has no `toolchain` line. `mise.toml` pins the exact Go version. The linter rejects a `toolchain` line.

## Consequences

- All components use the same version of each dependency, in each build. A component cannot keep an older version of a dependency than the other components use.
- A dependency update for one component is an update for all components. The tests of all components show its effect.
- The `go.mod` lists the dependencies of all components, and does not show which component uses a dependency. A binary still contains only the packages that it imports.
- The compiler does not prevent an import of a package that is not in `internal/` and not in `pkg/`. For such packages, the rule that services share code only through `pkg/` depends on review.
- A component that leaves the repository needs its own `go.mod`, and it must vendor the `pkg/` libraries that it uses. It is not likely that the shared libraries will be published in a form that a `require` directive can use.
- A new service needs changes in more than one file: the linter configuration and the commit scopes.
- Generic code that only one service uses stays in that service until a second service needs it. To move it to `pkg/` is later work.
