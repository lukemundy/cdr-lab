# cdr-lab

A standards-compliant simulation of Australia's Consumer Data Right (CDR), written in Go. It runs a mock CDR Register and mock banks, so you can test your own CDR software, for example an accredited data recipient (ADR) client, without the real CDR.

Everything in this repository is a mock. Real CDR software lives in other repositories and uses the lab as a test environment.

This project is not affiliated with the ACCC, the Data Standards Body (DSB) or any CDR participant.

## Status

Early development. Three services run in the lab, and a black-box conformance suite tests the Register and the data holder:

| Service | What it is | Specification |
|---|---|---|
| Mock CDR Register | The CDR Register | [design/README.md](services/register/design/README.md) |
| Mock data holder | A bank, with one of four preset brands for each instance | [design/README.md](services/dataholder/design/README.md) |
| Mock data recipient | A data recipient brand with a consumer web app, with no CDR function yet | [design/README.md](services/datarecipient/design/README.md) |

The lab CLI, `cdrlab`, sets up and runs the lab ([specification](cli/cdrlab/design/README.md)).

The lab is not a production service. [`docs/production-gaps.md`](docs/production-gaps.md) lists the shortcuts that are acceptable only in the lab.

## Quick start

To run the lab with Docker Compose from release `vX.Y.Z` (the [install guide](docs/guides/install.md) also shows how to verify the download):

```sh
gh release download vX.Y.Z --repo lukemundy/cdr-lab --pattern 'cdrlab-vX.Y.Z-linux-amd64.tar.gz'
tar -xzf cdrlab-vX.Y.Z-linux-amd64.tar.gz cdrlab
./cdrlab up
```

To install the umbrella chart in a Kubernetes namespace, first do the [prerequisites](charts/cdr-lab/README.md#before-the-install), then:

```sh
helm install cdr-lab oci://ghcr.io/lukemundy/cdr-lab/charts/cdr-lab --version X.Y.Z --namespace lab --values my-values.yaml --wait --timeout 10m
kubectl get pods,jobs --namespace lab
```

The [guides](docs/guides/) cover install, upgrade, ADR and data holder development, and CI.

## Target standards

| Item | Version |
|---|---|
| Consumer Data Standards | 1.36.0 |
| Security profile | FAPI 1.0 Advanced, with PAR, PKCE and JARM |

## Layout

```
cdr-lab/
├── mise.toml            # tool versions and tasks
├── compose.yaml         # the lab deployment
├── composeproject.go    # embeds the compose project in cdrlab, for the release mode (ADR 0038)
├── charts/              # Helm charts: one for each service, and the umbrella chart
├── cli/                 # the lab CLI, cdrlab
├── conformance/         # black-box tests of the deployed services, and the pinned CDS OpenAPI specifications
├── docs/
│   ├── adr/             # architecture decision records
│   ├── guides/          # guides for people who use the lab
│   ├── production-gaps.md # shortcuts that are acceptable in the lab only
│   ├── spec/            # specification conventions and glossary
│   └── research/        # dated research notes (background only)
├── lab/                 # lab-only set-up files, for example the PostgreSQL databases and users
├── pkg/                 # shared libraries, one directory each
├── pki/                 # the specification of the mock CDR PKI
├── services/            # one directory and one binary for each service
├── tools/               # repository tools that are not part of the lab
└── web/                 # the pnpm workspace: web UIs in apps/, shared packages in packages/
```

The repository is one Go module. Each directory in `pkg/`, `services/` and `cli/` is one component, and a service or a CLI must not import a service ([ADR 0004](docs/adr/0004-module-layout.md)). All JavaScript and TypeScript code is in the pnpm workspace in `web/` ([ADR 0015](docs/adr/0015-web-toolchain-and-npm-supply-chain-controls.md)). Each service has a hexagonal architecture and event-sourced participant state ([ADR 0013](docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md), [ADR 0003](docs/adr/0003-event-sourcing.md)).

## Specifications

Each component has a specification: a `DESIGN.md`, or a `design/` directory for a large one. A specification holds requirements only, each one rule with an identifier and a source ([ADR 0024](docs/adr/0024-write-specifications-as-identified-single-requirements-split-by-area.md)). The [specification conventions](docs/spec/conventions.md) give the templates, the formats and the [sources for requirements](docs/spec/conventions.md#sources-for-requirements). The CDS is always the first source. Specifications and ADRs have no status field: on `main` means accepted.

## Images, releases and telemetry

- CI publishes the tested image and chart of each component from `main` to `ghcr.io/lukemundy/cdr-lab/`, tagged with the commit ID. The images use distroless base images pinned by digest, and are built for `linux/amd64` and `linux/arm64` ([ADR 0018](docs/adr/0018-container-images-with-a-web-ui-build-stage.md), [ADR 0030](docs/adr/0030-build-each-service-image-for-linux-amd64-and-linux-arm64-with-a-cross-compile.md)).
- A release gives one version, `vX.Y.Z`, to all images, charts and `cdrlab` binaries. It promotes the tested images without a rebuild, and signs everything with keyless cosign, with provenance and an SBOM ([ADR 0029](docs/adr/0029-promote-the-tested-image-of-each-service-to-a-release-with-no-rebuild.md), [ADR 0031](docs/adr/0031-sign-releases-with-keyless-cosign-and-attest-provenance-and-an-sbom.md), [release specification](tools/release/DESIGN.md)).
- The services send OpenTelemetry traces, metrics and logs. CDR data and personal information never go in telemetry, even when the data is mocked ([ADR 0006](docs/adr/0006-telemetry.md)). Metrics show only the state of each process, and traces show each request ([ADR 0026](docs/adr/0026-send-metrics-only-for-the-state-of-a-process.md)).

## Development

Install [mise](https://mise.jdx.dev), then run these from the repository root:

```sh
mise install          # install the pinned tools
mise run hooks        # install the git hooks (once, after you clone)
mise run tidy         # go mod tidy
mise run fmt          # format the code
mise run generate     # generate the committed files
mise run lint         # all linters and the specification check
mise run test         # all tests except conformance (needs the lab PostgreSQL)
mise run audit        # scan the Go and npm dependencies
mise run build        # build the binaries and the web workspace
mise run build:images # build the lab images for this host
mise run conformance  # black-box tests against the running lab
mise run test:lab-web # browser tests against the running lab
```

`mise tasks` lists all tasks. CI runs each job only when its inputs change (`.github/workflows/ci.yml`).

## Lab deployment

The lab runs on Docker Compose (`compose.yaml`): all the services, PostgreSQL, a [Mailpit](https://mailpit.axllent.org) mail sink and a [`grafana/otel-lgtm`](https://github.com/grafana/docker-otel-lgtm) telemetry backend. It needs Docker, or Podman with its Docker-compatible API (`DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock`).

```sh
mise run up         # build and start the lab, and wait until it is healthy
mise run down       # stop the lab and keep its data
mise run reset      # stop the lab and delete its data and its CAs
```

`mise run up --no-telemetry` starts the lab without telemetry, as CI does. By default the lab runs one bank, `wattle-bank`, and one data recipient, `sorrel`. `mise run up --all-banks` also starts the other three preset banks.

All ports bind to `127.0.0.1`. The operator interfaces need the bearer token `lab-operator-token`.

| Port | Service | Listener |
|---|---|---|
| 8080 | `register` | Internal (HTTP): health, operator interface and console, CRL |
| 8081 | `register` | Public APIs (TLS) |
| 8083 | `register` | Secure APIs (mTLS) |
| 8180 | `wattle-bank` | Internal (HTTP): health, operator interface and console |
| 8181 | `wattle-bank` | Public APIs and authorisation pages (TLS) |
| 8183 | `wattle-bank` | Secure APIs (mTLS) |
| 8185 | `wattle-bank` | Online bank site (TLS, lab only) |
| 8280–8285 | `redgum-bank` | As 8180–8185 (`--all-banks`) |
| 8380–8385 | `harbour-bank` | As 8180–8185 (`--all-banks`) |
| 8480–8485 | `kestrel-bank` | As 8180–8185 (`--all-banks`) |
| 8580 | `sorrel` | Internal (HTTP): health |
| 8585 | `sorrel` | Web app (TLS) |
| 8025 | `mail` | Mail sink |
| 8443 | `openid` | OpenID conformance suite (profile `openid`) |
| 5432 | `postgres` | PostgreSQL, one database for each service |
| 3000 | `lgtm` | Grafana |
| 3100, 3200, 9090 | `lgtm` | Loki, Tempo and Prometheus APIs |
| 4317, 4318 | `lgtm` | OTLP gRPC and HTTP |

For example, to get the products of `wattle-bank` with [HTTPie](https://httpie.io):

```sh
http --verify .lab/pki/public/root.crt https://localhost:8181/cds-au/v1/banking/products x-v:5
```

### Certificates

`cdrlab up` creates two lab CAs in `.lab/pki/`: a CDR CA and a separate public CA. TLS clients trust `public/root.crt`, and mTLS clients and servers trust `cdr/ca-bundle.pem`. In a running lab the Register is the CDR intermediate CA, and it issues every CDR certificate ([ADR 0022](docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md)). [`lab/certificates.json`](lab/certificates.json) lists the lab certificates. See [pki/DESIGN.md](pki/DESIGN.md) and the [cdrlab specification](cli/cdrlab/design/README.md).

### Data

- Each service has its own database and runs its own migrations ([ADR 0020](docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md)).
- The lab loads its participants from [`services/register/seed/example.json`](services/register/seed/example.json), and each bank loads the product catalogue of its brand ([brands](services/dataholder/brands/README.md)).
- `cdrlab participants add` and `cdrlab customers add` load your own participants and customers ([onboarding](cli/cdrlab/design/onboarding.md)). Use invented data only.

### Conformance tests

The tests in `conformance/` are black-box tests of the CDR behaviour of the running services ([conformance/DESIGN.md](conformance/DESIGN.md)). Run them on a new lab:

```sh
mise run reset && mise run up && mise run conformance
```

CI also runs the suite against the umbrella chart in a kind cluster ([ADR 0036](docs/adr/0036-test-the-umbrella-chart-in-a-kind-cluster-in-ci-with-the-conformance-suite-as-a-job-in-the-cluster.md)). To do the same locally:

```sh
mise run cluster:up && mise run cluster:conformance; mise run cluster:down
```

`mise run openid-conformance` runs the FAPI 1.0 Advanced test plan of the [OpenID conformance suite](https://gitlab.com/openid/conformance-suite), with the Australian CDR profile, against the data holder ([ADR 0001](docs/adr/0001-auth-server-library.md)). It needs a new lab, and it does not run in CI.

## Cluster deployment

The umbrella chart [`charts/cdr-lab`](charts/cdr-lab/README.md) deploys the Register and the preset banks to one Kubernetes namespace ([specification](charts/cdr-lab/DESIGN.md)). Its README gives the prerequisites, an example values file and the operator access.

## Licence

Copyright 2026 Luke Mundy. Licensed under the [Apache License 2.0](LICENSE). See [NOTICE](NOTICE) for the attribution and for the parts of the repository that have a different licence.
