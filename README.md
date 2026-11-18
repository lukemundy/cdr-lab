# cdr-lab

A standards-compliant simulation of Australia's Consumer Data Right (CDR) in Go. It has a mock CDR Register and mock data holders.

This project is not affiliated with the ACCC, the Data Standards Body (DSB) or any CDR participant.

## Status

Early development. Two services run in the lab deployment, and a black-box conformance suite checks them:

| Service | Summary | Specification |
|---|---|---|
| Mock CDR Register | Participants and their statuses, SSAs, and the Register APIs | [design/README.md](services/register/design/README.md) |
| Mock data holder | One banking brand for each instance, from four preset brands: product data, the authorisation server, authorisations, and the banking APIs of an arrangement | [design/README.md](services/dataholder/design/README.md) |

## Target standards

| Item | Version |
|---|---|
| Consumer Data Standards | 1.36.0 |
| Security profile | FAPI 1.0 Advanced, with PAR, PKCE and JARM |

Change these versions only on purpose, in a separate change.

## Component types

- **Mock** components copy the behaviour of their production equivalents. They are not for production use. They can include test features, for example seed data.
- **Real** components use the mocks now. To connect them to the real CDR, you change only their configuration: Register URL, CA bundles, certificates, keys and registration IDs. Real components must not contain mock-specific code or test backdoors.

## Layout

```
cdr-lab/
├── mise.toml            # tool versions and tasks
├── compose.yaml         # the lab deployment
├── charts/              # Helm charts, for deployments outside the lab
├── conformance/         # black-box tests of deployed components, and the pinned CDS OpenAPI specifications
├── docs/
│   ├── adr/             # architecture decision records
│   ├── production-gaps.md # shortcuts that are acceptable in the lab only
│   ├── spec/            # the conventions and the glossary of the specifications
│   └── research/        # dated research notes with sources (background only)
├── lab/                 # lab-only set-up files, for example the databases and users of the lab PostgreSQL
├── pkg/                 # shared libraries, one directory each
├── pki/                 # mock CDR PKI: creates the lab CAs and issues certificates
├── services/            # one directory and one binary for each service
├── tools/               # repository tools that are not components of the lab
└── web/                 # the pnpm workspace: web UIs in apps/, shared packages in packages/
```

The repository is one Go module. Each directory in `pkg/` and `services/` is one component. A service must not import another service ([ADR 0004](docs/adr/0004-module-layout.md)). All JavaScript and TypeScript code is in one pnpm workspace in `web/`, which `go.mod` ignores ([ADR 0015](docs/adr/0015-web-toolchain-and-npm-supply-chain-controls.md)). Each service has a hexagonal architecture, and participant state is event sourced ([ADR 0013](docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md), [ADR 0003](docs/adr/0003-event-sourcing.md)).

To add a service, copy `services/register`, except its `design/` directory: write a new specification. Then change the import paths, and add the service to `.golangci.yml` (depguard rules), `cog.toml` (scope), its `Dockerfile`, `compose.yaml` (with `<<: *lab-dns`, as each compose service has) and the layout above. A service that keeps state gets its own database ([ADR 0002](docs/adr/0002-data-store.md)): its migrations create its application role, and the lab adds the database, an owner user and a service user to `lab/postgres/init.sql`, the membership of the role to `lab/postgres/grants.sql`, and a `<service>-migrate` container to `compose.yaml` ([ADR 0020](docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md)).

## Conventions

### Specifications

- Each service, `pkg/` library and other component has a specification: a `DESIGN.md`, or a `design/` directory when it has more than 5,000 words. It gives requirements only, also non-functional requirements, and does not describe the implementation ([ADR 0024](docs/adr/0024-write-specifications-as-identified-single-requirements-split-by-area.md)). [`docs/spec/conventions.md`](docs/spec/conventions.md) gives the templates and the formats.
- Each requirement is one rule, with an identifier and a link to its primary source or to a judgement in the same file. Refer to a rule with a descriptive link, never with an identifier alone.
- A rule that a `pkg/` library implements is in the specification of that library. A service gives only its differences.
- The research notes in `docs/research/` are background. Do not cite them as a source, and do not cite through them.
- Specifications and ADRs have no status field. On `main` means accepted. Anything else is a draft.

### Sources for requirements

Use this process to decide each requirement, and to decide when sources are not clear or do not agree:

1. The CDS text and specifications as written, for the target CDS version. This is always the first source.
2. When the CDS is not clear or says nothing: first-party ACCC guidance from official sources, for example the ACCC Developer Portal or the CDR Support Portal. The ACCC mock implementations are not a source.
3. When the ACCC guidance is also not clear or says nothing: our own judgement. Balance these items, in approximately this order:
   1. The spirit and intent of the Consumer Data Right.
   2. Relevant technical standards, protocols, RFCs, frameworks and industry benchmarks.
   3. The least complex solution.

Examples:

- The CDS does not give an HTTP response code for a condition. Use the code that is correct for the HTTP protocol and usual in the industry, if it does not go against the spirit or intent of the CDR.
- The CDS is not clear or does not give enough information, but ACCC supporting documentation gives guidance or a clarification. Use the ACCC guidance.
- The CDS and the ACCC say nothing about a detail, and more than one solution meets industry norms. Use the solution that is the simplest to understand, implement and maintain.
- The CDS does not agree with an ACCC support article or another official ACCC document that is not a specification. Follow the CDS.
- The CDS text does not agree with the CDS OpenAPI specification, and the ACCC has not clarified it. Use the option that agrees more with the spirit of the CDR or with industry norms.

Record each choice that this process makes as a judgement next to its requirements, with its basis and its reason.

### Other conventions

- **Production gaps:** [`docs/production-gaps.md`](docs/production-gaps.md) lists the shortcuts that are acceptable in the lab only. A change that takes such a shortcut adds an entry, and a change that closes a gap removes its entry.
- **Code:** write tests first (TDD). Fix lint issues; do not suppress them. If you must use `//nolint`, name the linter and give the reason.
- **Telemetry:** OpenTelemetry through `pkg/platform/telemetry`. CDR data and personally identifiable information must never appear in telemetry, also when the data is mocked. The client IP address is the only exception ([ADR 0006](docs/adr/0006-telemetry.md)).
- **Commits:** Conventional Commits, with the component directory name as the scope ([ADR 0007](docs/adr/0007-conventional-commits.md), `cog.toml`).
- **npm dependencies:** `web/pnpm-workspace.yaml` sets the supply-chain controls of [ADR 0015](docs/adr/0015-web-toolchain-and-npm-supply-chain-controls.md). Do not weaken them. A dependency with a build script needs an `allowBuilds` entry with a reason. A person reviews each major update.
- **Container images:** distroless base images, pinned by digest ([ADR 0018](docs/adr/0018-container-images-with-a-web-ui-build-stage.md)). CI publishes the tested image of each service from `main` to `ghcr.io/lukemundy/cdr-lab/<service>` (the service directory name), tagged with the full and the short commit ID. When a chart changes, CI also publishes the charts from `main` ([charts/register](charts/register/README.md), [charts/dataholder](charts/dataholder/README.md)).
- **Writing:** write all prose (comments, docs, commit messages) in ASD-STE100 Simplified Technical English. Use inclusive language, for example allowlist/blocklist and primary/replica.

## Development

Install [mise](https://mise.jdx.dev), then run from the repository root:

```sh
mise install         # install the pinned tools, also Node and pnpm
mise run hooks       # install the git hooks (one time, after you clone)
mise run tidy        # go mod tidy (CI fails if this changes files)
mise run fmt         # format the code
mise run generate    # generate the committed files, for example the console API types (CI fails if this changes files)
mise run lint        # golangci-lint (linters, formatting, architecture rules); tsc, oxlint and oxfmt for web/; Go and Node image versions; helm lint and kubeconform for charts/; the specification check (ADR 0024)
mise run test        # all tests except the conformance tests (needs the lab PostgreSQL; web tests run in Chromium)
mise run audit       # OSV-Scanner on go.mod and web/pnpm-lock.yaml
mise run build       # all service binaries into bin/, and the web workspace (build:go, build:web)
mise run build:images # the images of the lab deployment (CI builds them one time and the lab jobs use up --no-build)
mise run conformance # black-box tests against the running lab deployment
mise run test:lab-web # browser tests of the web UIs against the running lab deployment
mise run openid-conformance # the OpenID FAPI 1.0 Advanced test plan against the data holder in the running lab
```

`build`, `fmt`, `lint` and `test` run a `:go` and a `:web` task, for example `mise run test:web`. `lint` also runs `lint:versions`, `lint:chart` and `lint:spec`. `lint:spec` checks the structure of the specifications and the `// Verifies:` comments of the tests ([conventions](docs/spec/conventions.md#the-specification-check)). The web tasks install the workspace from the lock file first. CI runs each job only when its inputs change (`.github/workflows/ci.yml`).

## Lab deployment

The lab deployment is Docker Compose (`compose.yaml`). It runs all services, PostgreSQL, a [Mailpit](https://mailpit.axllent.org) mail sink and a [`grafana/otel-lgtm`](https://github.com/grafana/docker-otel-lgtm) telemetry backend. It needs Docker, or Podman with its Docker-compatible API (`DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock`).

```sh
mise run up         # create the lab CAs if necessary, build and start the lab (or only the named services), and wait until it is healthy
mise run down       # stop the lab; data stays
mise run reset      # stop the lab and delete all its data and its CAs
```

To start the lab without telemetry, set `COMPOSE_PROFILES` and `LAB_OTLP_ENDPOINT` to empty values, as CI does.

The lab and CI run one bank, `wattle-bank`. To also start the three other preset banks (compose profile `all-banks`), run `mise run up --all-banks`. Each bank is its own instance, with its own Register record and its own block of ports.

All ports bind to `127.0.0.1`. `compose.yaml` gives the configuration of each service. The operator interfaces need the bearer token `lab-operator-token`.

| Port | Service | Use |
|---|---|---|
| 8080 | `register` | Internal (HTTP): health probes, operator interface, operator console (no login) and the CRL of the mock CDR intermediate |
| 8081 | `register` | Public APIs (TLS) |
| 8083 | `register` | Secure APIs (mTLS) |
| 8180 | `wattle-bank` | Health probes |
| 8181 | `wattle-bank` | Public APIs, OpenID Provider configuration and authorisation pages (TLS) |
| 8182 | `wattle-bank` | Operator interface (HTTP) |
| 8183 | `wattle-bank` | Secure APIs: authorisation server endpoints and resource APIs (mTLS) |
| 8184 | `wattle-bank` | Operator console: web UI and its API (HTTP, no login) |
| 8185 | `wattle-bank` | Online bank site: web UI and its API (TLS, lab-only) |
| 8280–8285 | `redgum-bank` | As 8180–8185 (profile `all-banks`) |
| 8380–8385 | `harbour-bank` | As 8180–8185 (profile `all-banks`) |
| 8480–8485 | `kestrel-bank` | As 8180–8185 (profile `all-banks`) |
| 8025 | `mail` | Mail sink: web page and API |
| 8443 | `openid` | OpenID conformance suite: web page and API (profile `openid`) |
| 5432 | `postgres` | PostgreSQL: one database for each service |
| 3000 | `lgtm` | Grafana |
| 3100, 3200, 9090 | `lgtm` | Loki, Tempo (also its MCP server at `/api/mcp`) and Prometheus APIs |
| 4317, 4318 | `lgtm` | OTLP gRPC and HTTP |

Example calls from the host ([HTTPie](https://httpie.io)):

```sh
http --verify .lab/pki/public/root.crt https://localhost:8181/cds-au/v1/banking/products x-v:5
http --verify .lab/pki/cdr/ca-bundle.pem --cert .lab/pki/conformance/client.crt \
  --cert-key .lab/pki/conformance/client.key https://localhost:8083/...
```

### Certificates

`mise run pki` creates the lab CAs (a CDR CA and a separate public CA) and the certificates of the lab services and tests in `.lab/pki/`, which git ignores. `mise run up` runs it first, and `mise run reset` deletes the CAs. TLS clients trust `public/root.crt`, and mTLS clients and servers trust `cdr/ca-bundle.pem`. The `pki` command also issues CDR certificates from a CSR. See [pki/DESIGN.md](pki/DESIGN.md) and `go run ./pki/cmd/pki -h`. In a running lab, the Register is the CDR intermediate CA: its operator interface issues CDR client and server certificates from a CSR and revokes CDR certificates, with no host access, and the Register signs and serves the CRL ([ADR 0022](docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md)).

### Data

- **Database:** each service reads its connection from the libpq variables (`PG*`). A service applies its migrations at startup only when `DATABASE_MIGRATE_ON_START` is `true`; otherwise, run its `migrate` subcommand. Migrations are goose SQL files. In the lab, a one-shot `<service>-migrate` container runs the migrations as the owner user, and the service connects as a service user with the application role only ([ADR 0020](docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md)). Tests connect as the superuser `postgres`, and tests of database adapters create a new database with `databasetest.New(t)`.
- **Seed data:** one-shot compose services load [`services/register/seed/example.json`](services/register/seed/example.json) (participants) and the product catalogue of the data holder brand before the services start. To load a different set, change the file, then `mise run reset` and `mise run up`.
- **Data holder brand:** each data holder instance gets its name, logo, colours and product catalogue from a preset brand or a mounted brand directory ([services/dataholder/brands](services/dataholder/brands/README.md), [ADR 0023](docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md)).
- **Customer data:** the data holder operator interface creates customers, accounts and transactions ([data holder: operator interface](services/dataholder/design/operator-interface.md)). Use invented data only.
- **Register operations:** the Register operator interface creates and changes participants, registers client certificates and changes statuses ([Register: operator interface](services/register/design/operator-interface.md#reg-0800)).

### Conformance tests

The tests in `conformance/` are black-box tests of the CDR behaviour of the deployed components. [conformance/DESIGN.md](conformance/DESIGN.md) gives the rules for each test. The suite passes one time after a new lab:

```sh
mise run reset && mise run up && mise run conformance
```

### OpenID conformance suite

`mise run openid-conformance` runs the FAPI 1.0 Advanced test plan of the [OpenID conformance suite](https://gitlab.com/openid/conformance-suite), with the Australian CDR profile, against the data holder authorisation server ([ADR 0001](docs/adr/0001-auth-server-library.md)). It starts the suite (compose profile `openid`), registers its two clients with the data holder, and fails when a test module fails or has a warning. The output gives a link to the log of each module on port 8443. It needs a new lab, because the data holder refuses a second registration of the clients. It does not run in CI.

## Licence

Copyright © Luke Mundy. All rights reserved. Not yet licensed.
