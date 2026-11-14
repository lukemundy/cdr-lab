# 0018. Build service images on distroless, with a build stage for an embedded web UI, and pin all images by digest

- Date: 2026-10-22

## Context

This ADR supersedes [ADR 0008](0008-container-images.md). It keeps the decisions of ADR 0008, and adds a build stage for a service that embeds a web UI.

Each service runs as a container image. Services handle sensitive data and private keys. Thus an image must contain as little as possible, and must not run as root.

An image tag can move to different content. A build that uses tags can then change with no change in the repository, and a compromised tag gets in with no review.

Each service is a Go binary. With no cgo, it needs no C library and no shell.

A service can now embed the build of a web UI in its binary ([ADR 0014](0014-web-uis-as-embedded-react-single-page-applications.md)). Node and pnpm build the UI from the workspace in `web/` ([ADR 0015](0015-web-toolchain-and-npm-supply-chain-controls.md)). ADR 0008 gave two stages and a build context with only the Go source, so it has no place for this build.

## Decision

- Each service has a `Dockerfile` in its directory. The build context is the repository root, because the build needs the `go.mod` and the `pkg/` libraries ([ADR 0004](0004-module-layout.md)). `.dockerignore` is an allowlist: only the `pkg/` and `services/` source, and the `web/` source, go into the context. The `web/` source does not include `node_modules` or build output.
- The build stage compiles the binary with `CGO_ENABLED=0`. The final image contains only the binary.
- A service that embeds a web UI has one more stage before the build stage. It installs the workspace dependencies from the lock file and builds the UI with the Node and pnpm versions that `mise.toml` pins. The Go build stage copies only the UI build output.
- Final images use a [distroless](https://github.com/GoogleContainerTools/distroless) base image, `static` and `nonroot` variants.
- Pin each image that the repository uses by digest, with its tag. This includes base images, build images and third-party images.
- The build sets the build information (commit, build time and build ID) in the binary from build arguments ([ADR 0006](0006-telemetry.md)).

## Consequences

- A final image has no shell, no package manager and no Node. You cannot start a shell in a running service container. Use the telemetry and the service APIs to diagnose a problem.
- A container health check cannot run a command from the image, for example `curl`.
- A service cannot use a library that needs cgo.
- A service cannot write to paths that the `nonroot` user does not own, and cannot listen on a port below 1024 without more configuration.
- An image changes only through a change in the repository. Security fixes arrive only when a change updates the digest, so the digests must be updated regularly.
- The Node versions in `mise.toml` and in the Node build image must agree. Renovate updates the two in one pull request (`renovate.json`, group `node`), and `mise run lint:versions` fails when they do not agree. The Go versions in `mise.toml` and in the Go build image have the same rule (group `go`).
- A service build with a UI is slower, and the install of npm packages runs in the image build. The pnpm controls of ADR 0015 apply there as in CI.
- A change in `web/` invalidates the build cache of each service image that copies the workspace.
