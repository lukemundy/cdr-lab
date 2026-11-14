# Four preset banks in the lab: technical research

- Research date: 2026-11-02
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 3 of the project "Each data holder instance is a distinct, branded bank": "All four preset banks run in the lab. Behind a compose profile, each with its own catalogue, Register seed record, database and certificates. The lab and CI run one brand by default." Also the orphan containers after a reset.
- Related notes: [Branded data holder instances](2026-11-01-branded-data-holder-instances.md), [Branded emails and operator console](2026-11-02-branded-emails-and-console.md). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Compose Specification: profiles | https://docs.docker.com/reference/compose-file/profiles/ |
| S2 | Compose Specification: fragments (YAML anchors and aliases) | https://docs.docker.com/reference/compose-file/fragments/ |
| S3 | Compose Specification: extension fields (`x-`) | https://docs.docker.com/reference/compose-file/extension/ |
| S4 | Compose Specification: `extends` | https://docs.docker.com/reference/compose-file/services/#extends |
| S5 | Compose Specification: `include` | https://docs.docker.com/reference/compose-file/include/ |
| S6 | Compose Specification: `depends_on` (long syntax, `required`) | https://docs.docker.com/reference/compose-file/services/#depends_on |
| S7 | `docker compose down` reference (`--remove-orphans`) | https://docs.docker.com/reference/cli/docker/compose/down/ |
| S8 | PostgreSQL 18: database roles are global to the cluster | https://www.postgresql.org/docs/18/database-roles.html |
| S9 | PostgreSQL Docker image: initialisation scripts run only on an empty data directory | https://hub.docker.com/_/postgres |
| S10 | CDS 1.36.0: Multi-Brand Support | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#multi-brand-support-separate-issuers-for-data-holder-brands |
| S11 | Compose Specification: the `profiles` attribute of a service | https://docs.docker.com/reference/compose-file/services/#profiles |

## 1. Compose definition of three more data holders

Facts:

- One data holder brand is three compose services: `wattle-bank` (`compose.yaml:89-152`), `wattle-bank-migrate` (`compose.yaml:154-164`) and `wattle-bank-seed` (`compose.yaml:166-175`). They already share the build with the anchor `&wattle-bank-build` (`compose.yaml:90`) and the database settings with `&wattle-bank-database-environment` (`compose.yaml:99`).
- The values that change for each brand are: the database user, password and name (`compose.yaml:102-104`), `DATA_HOLDER_BRAND_ID` (`:107`), `DATA_HOLDER_BRAND` (`:109`), `PUBLIC_BASE_URL` and `SECURE_BASE_URL` (`:110`, `:113`), `OPERATOR_ALLOWED_HOSTS` (`:125`), `MAIL_FROM` (`:129`), `CONSOLE_ALLOWED_HOSTS` (`:136`), the five host ports (`:137-142`) and the TLS volume `./.lab/pki/wattle-bank` (`:144`). Approximately 25 lines are the same for each brand.
- A service with `profiles` starts only when its profile is active. A service with no `profiles` always starts [S1]. A profile name matches the regular expression `[a-zA-Z0-9][a-zA-Z0-9_.-]+` [S11]. The file has the profiles `test`, `openid` and `telemetry` now (`compose.yaml:178`, `:232`, `:372`).
- A YAML anchor works only in one file. Compose does not resolve an anchor across files that `include` or `extends` loads [S2]. Compose ignores a top-level key that starts with `x-`, so it can hold a fragment [S3]. The file already uses `x-telemetry-environment` and `x-build-args` (`compose.yaml:4-11`).
- `extends` copies a service from the same file or a different file, and merges the local keys onto it. Sequences such as `ports` merge (are added), and mappings merge by key [S4]. Thus a brand that extends `wattle-bank` gets the Wattle Bank ports also, unless the base service has no ports.
- `include` loads a different Compose file as a sub-project. The included file cannot use the anchors of the parent [S5].
- The Register and the conformance container call each data holder by its compose service name, for example `https://wattle-bank:8081` (`services/register/seed/example.json`, Wattle Bank `endpointDetail`; `compose.yaml:110`).

Options:

| Option | Lines per brand (estimate) | Notes |
|---|---|---|
| A. `x-dataholder` extension field with the common environment, build, volumes and `depends_on`; each brand merges it with `<<:` and adds its own values | approximately 30 for 3 services | One file. Same pattern as now. The merge key `<<` does not merge sequences, so each brand gives its own `ports` and `volumes` in full. |
| B. `extends` from a base service | approximately 30 | Sequences merge, so the base cannot have ports. No gain over A, and a reader must know the merge rules of `extends`. |
| C. `include` of a generated file for each brand, or a template | approximately 0 hand lines | Needs a generator and a "generated files" check in CI. More complex. |

Recommendation: option A. Move the common keys of `wattle-bank` to `x-dataholder-environment` and similar fragments, and keep Wattle Bank and the three profile brands at the same level. Set `profiles: [banks]` (name to decide) on the nine new services only.

Ports. Fact: the README port table uses 808x for the Register and 818x for `wattle-bank` (README "Lab deployment"; `compose.yaml:47-53`, `:137-142`). Ports 8280-8284, 8380-8384 and 8480-8484 are not in use in `compose.yaml`. Recommendation: `redgum-bank` 828x, `harbour-bank` 838x, `kestrel-bank` 848x, with the same last digit as now (0 probes, 1 public, 2 operator, 3 secure, 4 console). The README gets one row ("828x, 838x, 848x: the other preset banks, profile `banks`"), not 15 rows.

Hostnames. Recommendation: the compose service name is the preset name: `redgum-bank`, `harbour-bank`, `kestrel-bank`. The preset directories already have these names (`services/dataholder/brands/`).

## 2. Databases and users

Facts:

- `lab/postgres/init.sql:10-14` creates the owner user, the service user and the database of Wattle Bank, five statements for each service. The image runs this file only when the data directory is empty [S9] (`compose.yaml:351-352`).
- `lab/postgres/grants.sql:3` grants `dataholder_app` to `wattle_bank`. This file runs at each start, after the migrations (`compose.yaml:355-369`).
- ADR 0020 (`docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md:19`): a user must exist before it owns a database or runs a migration, and the operator grants the role only after the migrations.
- A PostgreSQL role is global to the cluster [S8]. The data holder migration creates `dataholder_app` and ignores `duplicate_object` (`services/dataholder/internal/migrations/00019_create_application_role.sql:4-10`). Thus the migration of a second brand succeeds, and all brands share one application role. Each service user can connect only to its own database (`init.sql:13-14`), so the shared role does not give access to a different brand's data.
- `postgres-grants` depends on `wattle-bank-migrate` with `condition: service_completed_successfully` (`compose.yaml:366-369`). When a service refers to a service that the active profiles ignore, Compose does not enable that service, and it returns an error [S1]. When a `depends_on` entry has `required: false`, Compose only warns when the dependency is not started or not available [S6].

Options for the profile-off case:

- Always create all four databases and users in `init.sql`. Cost: approximately 7-8 MB of disk for each empty database (the size of `template1`), no memory and no process. `grants.sql` can then grant `dataholder_app` to all four users always, because Wattle Bank always runs the migration that creates the role. Recommendation: this option. It keeps the SQL static and needs no loop.
- Create the databases only when the profile is on. `init.sql` runs one time only [S9], so a lab that turns the profile on later has no databases until `mise run reset`. Not recommended.

A PL/pgSQL loop over the four names is possible, but 20 static lines are easier to read and to review than a `DO` block with `format()`. Recommendation: static SQL, one block for each brand, as now.

`postgres-grants` must not depend on the migrate services of the profile brands, or must use `required: false` [S6]. Option: one `<brand>-grants` step is not necessary; the profile brands' `-seed` services can depend on `postgres-grants` and on their own `-migrate`. But `grants.sql` runs before the profile migrations, and that is correct only because the role is global and Wattle Bank creates it. Record this order dependency in a comment.

## 3. Certificates

Facts:

- `mise run pki` issues for Wattle Bank: a public TLS certificate for `wattle-bank` and `localhost` from the public CA (`mise.toml:208-211`), and a CDR server certificate with O "Wattle Bank Limited", CN `wattle-bank` and SANs `wattle-bank` and `localhost` (`mise.toml:212-214`). It issues no CDR client certificate for a data holder; a data holder is not an mTLS client of the Register in the lab.
- Each certificate step runs only when its file does not exist (`mise.toml:201-214`). Thus a new step for a new brand runs on the next `mise run up` with no `reset`.
- The O in a server certificate is the legal entity name (`mise.toml:192`; `pki/DESIGN.md`, CSR profile).

Need: for each of the three brands, one public certificate (`<brand>`, `localhost`) and one CDR server certificate (O = legal entity name, CN and SAN = `<brand>`, SAN `localhost`), in `.lab/pki/<brand>/`. Approximately 8 lines each; a shell function or loop in the task can make it 4 lines.

Fact: all brands share `localhost` as a SAN. The host reaches each brand by its port, so that is correct.

Recommendation: issue the certificates for all four brands always. The cost is a few RSA 2048 key generations (less than one second each). Then `mise run pki` does not need to know the profile.

## 4. Register seed

Facts:

- The seed file has two top-level keys, `dataHolders` and `dataRecipients` (`services/register/seed/example.json`). Each data holder is a legal entity with `legalEntityId`, `legalEntityName`, `logoUri`, `status` and `dataHolderBrands`. Each brand has `dataHolderBrandId`, `brandName`, `industries`, `logoUri`, `status`, `endpointDetail` (`version`, `publicBaseUri`, `productBaseUri`, `resourceBaseUri`, `infosecBaseUri`, `websiteUri`) and `jwksEndpoint`. Wattle Bank Limited is the first entry.
- The preset directories give the brand names "Redgum Bank", "Harbour Bank" and "Kestrel Bank" (`services/dataholder/brands/<preset>/brand.json`).
- The Register conformance tests count the seed brands: `seedRegisteredBrands = 9` and `"banking": 2` (`conformance/register/data_holder_brands_test.go:33`, `:141`, and the asserts at `:202`, `:229`, `:264`, `:315`). Three more banking brands change these values to 12 and 5.
- `compose.yaml` loads one seed file always (`compose.yaml:78-83`). The data holder finds its own record by `DATA_HOLDER_BRAND_ID` (`compose.yaml:107`), and the conformance tests find it by `DATAHOLDER_BRAND_ID` (`conformance/internal/labtest/dataholder.go:71-80`).

Analysis of a record for a brand that does not run (profile off):

- A data recipient that lists the brands gets an ACTIVE brand whose endpoints do not answer. In production that is an outage of the data holder, which the CDR allows; a data recipient must handle it. It is not a problem for the conformance tests, which call only Wattle Bank.
- The status of a brand on the Register is a fact about its registration, not about whether it runs. A status that depends on the profile would need two seed files or a seed step for each profile.

Options:

- A. One seed file with all four banks, all ACTIVE, always. Simplest. The lab with the profile off shows three brands that do not answer. Recommendation: this option.
- B. Seed the three brands only when the profile is on, with a second seed file and a `register-seed` step in the profile. The Register seed uses `-if-empty` (`compose.yaml:81`), so a second file needs a different load mode, or the operator interface. More complex.
- C. Seed them as INACTIVE, and the operator changes the status. This tests a status cascade that nobody asked for. Not recommended.

Each brand needs a new legal entity (for example "Redgum Bank Limited"), not a second brand of Wattle Bank Limited, because the project says each is its own bank. The CDS multi-brand rules [S10] apply to brands of one legal entity, and do not apply here.

## 5. CI

Facts:

- CI sets `COMPOSE_PROFILES: ""` for all jobs (`.github/workflows/ci.yml:23`), and the conformance job runs `mise run up` and `mise run conformance` (`ci.yml:217-220`). With an empty profile list, a service with `profiles` does not start [S1]. Thus no CI job starts the three brands, with no change.
- `mise run up` starts all services that have no profile (`mise.toml:258-265`).

Analysis: the project constraint is that the lab and CI run one brand by default. The conformance suite runs against Wattle Bank only. A CI job that starts the profile proves only that the three containers become healthy, at the cost of three more image starts on a one-node runner.

Recommendation: no new CI job. Prove the profile in the lab after the merge: `COMPOSE_PROFILES=banks mise run up`, then a Get Products call to each brand and a check of each console header. Optionally, a cheap CI check: `docker compose --profile banks config --quiet`, which proves the file is valid with the profile on, with no start.

## 6. Orphan containers

Facts:

- `mise run reset` runs `docker-compose --profile '*' down --volumes` (`mise.toml:287`), and `mise run down` runs `docker-compose --profile '*' down` (`mise.toml:283`).
- `down` removes the containers of the services in the Compose file. `--remove-orphans` also removes containers of services that the file does not define now [S7]. `--profile '*'` makes all profiles active, so the profile containers are not orphans; a renamed or removed service is.
- The `mock-bank` containers stayed after the rename to `wattle-bank`, and kept port 8180.

Recommendation: add `--remove-orphans` to `reset` and also to `down`, and optionally to `up` (`docker compose up --remove-orphans`), which removes the orphan before it can hold a port. `up --remove-orphans` with no profile does not remove the containers of an inactive profile, because they are services of the file.

Other orphan risks:

- Named volumes of a removed service stay after `down --volumes` only if the file no longer declares them. Now all volumes are top-level (`compose.yaml:391-395`); a removed volume is a small disk cost only.
- One-shot containers that `docker compose run --rm` starts (`mise.toml:167`, `:277`) remove themselves. A run that the user stops with a signal can leave one; `--remove-orphans` does not remove it, because its service is in the file, but `down` does.
- Containers of a different compose project name (for example from a worktree with a different directory name) do not collide by name, because `compose.yaml:1` sets `name: cdr-lab`. Two worktrees share the project, so the ports collide only with each other, as now.

## 7. Memory and start time

Fact: the lab was not running at the research date (`docker stats --no-stream` gave no containers), so this note has no measurement.

Estimate (not a fact): a Go service with an idle database pool uses approximately 20-40 MB RSS. Three more data holders add approximately 60-120 MB, plus approximately three PostgreSQL backends each (approximately 10 MB each). The build is one image, shared by all four brands through the same `build` anchor, so the build time does not increase; Compose may tag four images, but the layers are the same. The start time increases by the migrations and seeds of three databases; these run in parallel after `postgres` is healthy. Recommendation: measure with `docker stats --no-stream` after the first lab run with the profile, and give the result in the verification of the ticket.

## 8. Risks and open questions

| Item | Kind |
|---|---|
| The name of the profile (`banks`, `all-banks` or `multi-brand`). A new convention, thus propose it. | Decision |
| Seed the three brands always (option 4A) or only with the profile (4B). 4A changes the counts in the Register conformance tests (`data_holder_brands_test.go:33`, `:141`). | Decision |
| The legal entity names of the three banks ("Redgum Bank Limited" and so on), their `websiteUri` (`.test` domains, as `wattlebank.test`) and their `MAIL_FROM` addresses. | Decision |
| Ports 828x, 838x and 848x. | Decision |
| `postgres-grants` depends on the Wattle Bank migration to create the shared role `dataholder_app`. If Wattle Bank leaves the default later, the order breaks. | Risk |
| Compose builds the same image four times, or once, depending on the `image` and `build` keys. Set one `image:` name with `pull_policy: build` on one service, or accept four tags of the same layers. Check in the lab. | Open question |
| The fixed OTPs (`compose.yaml:131`) and the operator token are the same for all brands. Acceptable in the lab; no new production gap, because the gap already exists for one brand. | Risk (low) |
| `mise run test:lab-web` uses the Wattle Bank console ports only (`mise.toml:148`). No change is necessary. | Fact |
| `docs/production-gaps.md` needs no new entry: four brands in one lab is not a shortcut that a production service could not operate with. | Recommendation |
