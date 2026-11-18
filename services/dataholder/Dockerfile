# Keep this stage the same in each service Dockerfile, so that BuildKit builds it one time for all images.
# Keep the Node version the same as in mise.toml; mise run lint checks it.
FROM docker.io/library/node:26.13.0-trixie-slim@sha256:e4db4c914bf8ad241bd24804fc25d272d89bad97e68cdc8a1414862790a02c39 AS web

# mise.toml pins pnpm, and compose.yaml gives its version from the environment that mise sets.
ARG PNPM_VERSION
RUN test -n "$PNPM_VERSION" || { echo "Set the build argument PNPM_VERSION: build through mise, for example mise run up." >&2; exit 1; }
# Without its scripts, the pnpm package runs through Node, not through its native binary.
RUN npm install --global --ignore-scripts "pnpm@${PNPM_VERSION}"

WORKDIR /src/web
COPY web/ ./
# pnpm applies the supply-chain controls of pnpm-workspace.yaml here, as in CI.
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --filter @cdr-lab/register-console... --filter @cdr-lab/dataholder-console... \
      --filter @cdr-lab/dataholder-bank... --filter @cdr-lab/ui... && \
    pnpm --filter @cdr-lab/register-console... --filter @cdr-lab/dataholder-console... \
      --filter @cdr-lab/dataholder-bank... --filter @cdr-lab/ui... build

# Keep the Go version the same as in mise.toml; mise run lint checks it.
FROM docker.io/library/golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS build

ENV CGO_ENABLED=0
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY pkg/ pkg/
COPY services/dataholder/ services/dataholder/
COPY --from=web /src/web/apps/dataholder-console/dist/ services/dataholder/internal/adapter/console/dist/
COPY --from=web /src/web/apps/dataholder-bank/dist/ services/dataholder/internal/adapter/banksite/dist/
COPY --from=web /src/web/packages/ui/dist/consumer-pages/ services/dataholder/internal/adapter/consumerpages/static/

ARG BUILD_COMMIT=unknown BUILD_COMMIT_SHORT=unknown BUILD_TIME=unknown BUILD_ID=local
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    buildinfo=github.com/lukemundy/cdr-lab/pkg/platform/buildinfo && \
    go build -trimpath -o /out/dataholder -ldflags="-s -w \
      -X ${buildinfo}.commit=${BUILD_COMMIT} \
      -X ${buildinfo}.commitShort=${BUILD_COMMIT_SHORT} \
      -X ${buildinfo}.buildTime=${BUILD_TIME} \
      -X ${buildinfo}.id=${BUILD_ID}" \
      ./services/dataholder/cmd/dataholder

FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329

LABEL org.opencontainers.image.source=https://github.com/lukemundy/cdr-lab
COPY --from=build /out/dataholder /dataholder
# The SIL Open Font License requires its text with each copy of the fonts of the web UIs.
COPY --from=web /src/web/packages/tokens/dist/fonts/*-LICENSE.txt /licenses/fonts/
EXPOSE 8080 8081 8083 8085
ENTRYPOINT ["/dataholder"]
