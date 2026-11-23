# 0030. Build each service image for linux/amd64 and linux/arm64, with a cross-compile

- Date: 2026-11-23

## Context

This ADR amends [ADR-0018](0018-container-images-with-a-web-ui-build-stage.md). It adds the platforms of each image. All other decisions of ADR 0018 stay.

The images are only for `linux/amd64`. Teams run the released images in their own Kubernetes clusters ([ADR 0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md)), and some of these clusters have `linux/arm64` nodes. An image with no variant for the platform of a node does not run on that node.

These facts constrain a multi-platform build:

- The pinned Node, Go and distroless base images are OCI image indexes, with a `linux/amd64` and a `linux/arm64` variant. Thus the build needs no new base image.
- The build stage compiles with `CGO_ENABLED=0` ([ADR 0018](0018-container-images-with-a-web-ui-build-stage.md)). The Go toolchain then compiles for a different platform with no C toolchain. The final stage has no `RUN` step, so it runs no program of the target platform.
- The web UI stage makes JavaScript, CSS, fonts and notice text. The output does not depend on the target platform.
- CI uses hosted runners only. The lab jobs run on `linux/amd64`.

Docker gives three strategies ([Docker: multi-platform builds](https://docs.docker.com/build/building/multi-platform/)):

1. **Emulation with QEMU** on one runner. No Dockerfile change. "Emulation with QEMU can be much slower than native builds", mostly for compilation. The Node install and the web build then also run emulated.
2. **A native runner for each platform**, and a merge of the two images into one index. GitHub gives `linux/arm64` hosted runners for public repositories at no cost ([GitHub: hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)). It needs more jobs and a merge step.
3. **A cross-compile on one runner.** The web and build stages run on the platform of the runner (`--platform=$BUILDPLATFORM`), and the Go build compiles for the target platform. The web stage runs one time for the two images.

## Decision

Each image of a service that CI publishes is an OCI image index with a `linux/amd64` and a `linux/arm64` image.

- The build cross-compiles on one hosted runner. The web stage and the Go build stage run on the platform of the runner, and the Go build sets the target platform. We use the cross-compile because it adds the least CI time and needs no new runner type or merge job.
- The build makes the licence notices of each image from the dependencies of its own platform, as ADR 0018 requires.
- The `linux/amd64` image of the index is the image that the lab jobs tested.
- CI does not test the `linux/arm64` image. It comes from the same source and the same build arguments as the tested image.

## Consequences

- The images run natively on `linux/amd64` and `linux/arm64` nodes. A cluster selects the variant for each node from the index.
- The digest that a team pins is the digest of the index, not of one platform.
- An error that occurs only on `linux/arm64` is not found in CI. A team that runs on `linux/arm64` finds it first.
- The Go build runs two times, so the image job takes more time. We did not measure the increase.
- A dependency that needs cgo cannot go into an image. ADR 0018 already excludes it, and the cross-compile depends on that rule.
- A Dockerfile that runs a Go tool in the build stage must build that tool for the runner, not for the target platform.
