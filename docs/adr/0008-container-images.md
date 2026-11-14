# 0008. Build service images on distroless, and pin all images by digest

- Date: 2026-10-04
- Superseded by [ADR-0018](0018-container-images-with-a-web-ui-build-stage.md)

## Context

Each service runs as a container image. Services handle sensitive data and private keys. Thus an image must contain as little as possible, and must not run as root.

An image tag can move to different content. A build that uses tags can then change with no change in the repository, and a compromised tag gets in with no review.

Each service is a Go binary. With no cgo, it needs no C library and no shell.

## Decision

- Each service has a `Dockerfile` in its directory. The build context is the repository root, because the build needs the `go.mod` and the `pkg/` libraries ([ADR 0004](0004-module-layout.md)). `.dockerignore` is an allowlist: only the `pkg/` and `services/` source goes into the context.
- The build has two stages. The build stage compiles the binary with `CGO_ENABLED=0`. The final image contains only the binary.
- Final images use a [distroless](https://github.com/GoogleContainerTools/distroless) base image, `static` and `nonroot` variants.
- Pin each image that the repository uses by digest, with its tag. This includes base images and third-party images.
- The build sets the build information (commit, build time and build ID) in the binary from build arguments ([ADR 0006](0006-telemetry.md)).

## Consequences

- A final image has no shell and no package manager. You cannot start a shell in a running service container. Use the telemetry and the service APIs to diagnose a problem.
- A container health check cannot run a command from the image, for example `curl`.
- A service cannot use a library that needs cgo.
- A service cannot write to paths that the `nonroot` user does not own, and cannot listen on a port below 1024 without more configuration.
- An image changes only through a change in the repository. Security fixes arrive only when a change updates the digest, so the digests must be updated regularly.
