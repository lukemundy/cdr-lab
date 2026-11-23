# The rules of a service image apply to this image. It has no web stage, because cdrlab has no web UI.
# Keep the Go version the same as in mise.toml; mise run lint checks it.
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS build

ENV CGO_ENABLED=0
WORKDIR /src
ARG TARGETOS TARGETARCH

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY pkg/ pkg/
COPY cli/cdrlab/ cli/cdrlab/

COPY tools/notices/ tools/notices/
# The notices tool runs on the build platform, but it lists the dependencies of the target platform.
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go build -o /tmp/notices ./tools/notices && mkdir -p /out && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH /tmp/notices ./cli/cdrlab/cmd/cdrlab > /out/third-party-notices.txt

ARG BUILD_COMMIT=unknown BUILD_COMMIT_SHORT=unknown BUILD_TIME=unknown BUILD_ID=local
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    buildinfo=github.com/lukemundy/cdr-lab/pkg/platform/buildinfo && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/cdrlab -ldflags="-s -w \
      -X ${buildinfo}.commit=${BUILD_COMMIT} \
      -X ${buildinfo}.commitShort=${BUILD_COMMIT_SHORT} \
      -X ${buildinfo}.buildTime=${BUILD_TIME} \
      -X ${buildinfo}.id=${BUILD_ID}" \
      ./cli/cdrlab/cmd/cdrlab

FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329

LABEL org.opencontainers.image.source=https://github.com/lukemundy/cdr-lab
COPY --from=build /out/cdrlab /cdrlab
# The licences of the Go modules require their notices with each copy of their code.
COPY --from=build /out/third-party-notices.txt /licenses/go/
ENTRYPOINT ["/cdrlab"]
