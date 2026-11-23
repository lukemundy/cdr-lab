# 0035. Give the lab one CLI, cdrlab, in cli/, with an image that follows the rules of a service image

- Date: 2026-11-23

## Context

This ADR amends [ADR-0018](0018-container-images-with-a-web-ui-build-stage.md), [ADR-0029](0029-promote-the-tested-image-of-each-service-to-a-release-with-no-rebuild.md) and [ADR-0030](0030-build-each-service-image-for-linux-amd64-and-linux-arm64-with-a-cross-compile.md). Their rules for the image of a service apply also to the image of the lab CLI.

The lab has one command for its set-up, `pki`, in `pki/cmd/pki`. It makes the CAs and issues certificates offline. The mise tasks run it with `go run`, so the host needs Go and a copy of the repository. [ADR 0004](0004-module-layout.md) gives `services/` for programs that run continuously and `pkg/` for shared libraries, but it gives no place for a command that a person or a Job runs and that then ends.

A cluster deployment needs commands that run in the cluster: the set-up of the mock CDR CA and the issue of CDR certificates through the Register ([ADR 0032](0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md)). A Job runs a command from an image. A team that runs the lab from a release also needs the commands on a host with no Go.

The real options were:

1. **More commands in `pki`, and a new image for it.** The name `pki` does not fit the commands that onboard participants and start the compose lab.
2. **A separate binary for each task.** Each binary needs its own image, release asset and documentation.
3. **One CLI with sub-commands for all lab tasks.** One binary and one image. A team learns one tool.

## Decision

The lab has one CLI, `cdrlab` (option 3). It replaces the `pki` command: when `cdrlab` gives all functions of `pki`, the repository removes `pki/cmd/pki`. The repository then has no other CLI for the lab. The repository tools in `tools/` are not lab CLIs and stay.

- `cli/` contains the CLIs of the lab. Each directory in `cli/` is one CLI, with its code in `internal/`. A CLI shares code with the services only through `pkg/` libraries, and it never imports a service ([ADR 0004](0004-module-layout.md)).
- A CLI uses the network interfaces of the services, for example the Register operator interface, as a lab operator does.
- The `cdrlab` image has the rules of the image of a service. Its `Dockerfile` is in its directory, with a distroless base image and the licence notices of its third-party code (ADR 0018). CI publishes it from `main` as an OCI image index for `linux/amd64` and `linux/arm64` (ADR 0030). A release promotes the tested image with no rebuild (ADR 0029).
- The image is for the Jobs of a cluster deployment. A person on a host uses a release binary or `go install`.

## Consequences

- A Job and a person use the same commands, from the same code.
- The release has one more image to check and to promote. The image commit rules apply to it.
- A change to `pkg/` is an input of the `cdrlab` image, as it is of each service image. Thus most changes to `pkg/` give a new `cdrlab` image.
- The binary in an image from `main` has the commit, not the version of a release, for the same reason as a service binary.
- The linter rules, the commit scopes and the `.dockerignore` allowlist get one more entry for `cli/`.
- Until `cdrlab` gives all functions of `pki`, the repository has two lab commands.
