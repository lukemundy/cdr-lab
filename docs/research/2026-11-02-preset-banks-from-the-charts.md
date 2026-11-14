# Four preset banks from the charts: technical research

- Research date: 2026-11-02
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 2 of the project "The data holder brands deploy to Kubernetes from published artifacts": "The four preset banks deploy with configuration only. The Register chart loads a participant seed file, and the repository has the seed file and the values for the four preset banks." Acceptance criteria of the project for this milestone: "A Register that the chart deploys with the seed file has the four preset brands at start-up, with their endpoints", and "The chart README gives each value and each Secret that a deployment of a brand needs".
- Out of scope (project): the deployment to a cluster, the Secrets of a deployment, a PostgreSQL server in the charts.
- Related notes: [Data holder image and chart](2026-11-02-data-holder-image-and-chart.md) (milestone 1), [Four preset banks in the lab](2026-11-02-four-preset-banks-in-the-lab.md) (the lab seed records). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Kubernetes: ConfigMaps (the data of a ConfigMap is 1 MiB at most; `data` and `binaryData`) | https://kubernetes.io/docs/concepts/configuration/configmap/ |
| S2 | Kubernetes: a ConfigMap volume (each key is a file in the mount directory; an update does not restart the pod) | https://kubernetes.io/docs/concepts/storage/volumes/#configmap |
| S3 | Kubernetes: init containers run in sequence, and each must complete before the next starts | https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ |
| S4 | Kubernetes: `kubectl create configmap --from-file=<key>=<path>` | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_configmap/ |
| S5 | Kubernetes: Secrets (one Secret is 1 MiB at most; Helm keeps each release in a Secret by default, see S7) | https://kubernetes.io/docs/concepts/configuration/secret/#restriction-data-size |
| S6 | Helm: values files, `--values` and `--set-file` (a list in a later values file replaces the full list) | https://helm.sh/docs/chart_template_guide/values_files/ and https://helm.sh/docs/helm/helm_install/ |
| S7 | Helm: storage backends (the release, with its values, is in a Secret) | https://helm.sh/docs/topics/advanced/#storage-backends |
| S8 | Helm: `.Files.Get` and `tpl` | https://helm.sh/docs/chart_template_guide/accessing_files/ and https://helm.sh/docs/howto/charts_tips_and_tricks/#using-the-tpl-function |
| S9 | Helm: `.helmignore` (without it, `helm package` includes each file of the chart directory) | https://helm.sh/docs/chart_template_guide/helm_ignore_file/ |
| S10 | bjw-s common library chart 5.x: `controllers.<c>.initContainers.<n>.order`, `persistence.<n>.advancedMounts` | https://github.com/bjw-s-labs/helm-charts/tree/main/charts/library/common |
| S11 | CDS 1.36.0: Get Data Holder Brands (`RegisterDataHolderBrand`, `RegisterDataHolderBrandServiceEndpoint`) and Get Data Holder Brands Summary | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-apis |
| S12 | RFC 6761, section 6.2: `.test` is reserved for tests | https://www.rfc-editor.org/rfc/rfc6761#section-6.2 |
| S13 | Argo CD: Helm sources (`valueFiles`, `values`, `valuesObject`, `fileParameters`, multiple sources) | https://argo-cd.readthedocs.io/en/stable/user-guide/helm/ |
| S14 | kubeconform | https://github.com/yannh/kubeconform |

## 1. How the Register chart loads a seed file

### Facts: the Register chart and the seed command

- The Register chart runs one init container, `migrate`, as the owner user, and then the main container `serve` (`charts/register/templates/common.yaml:66-76`, `:77-80`). It has no `seed` init container and no seed value (`charts/register/values.yaml`). A Register from the chart starts with no participants.
- The chart test requires exactly one init container (`charts/register/chart_test.go:37`). A seed init container that is on by default breaks this test.
- `register seed [-if-empty] <file.json>` reads the file, then loads it into a Register that has no participants. With `-if-empty`, a Register that has participants is not an error (`services/register/internal/root/subcommands.go:21-61`). The load is all or nothing (`services/register/DESIGN.md:526`).
- The seed command needs only the database and the telemetry settings (`services/register/internal/root/run.go:42-60`). It applies the migrations only when `DATABASE_MIGRATE_ON_START` is `true` (`subcommands.go:44`, `:67-74`).
- In the lab, `register-seed` runs `seed -if-empty /seed/example.json` as the service user `register`, after `postgres-grants`, and the Register starts after it (`compose.yaml:116-124`, `:57-60`, `:101-102`). Thus the service user has the rights for the load.
- The file decoder refuses unknown fields and data after the participants (`services/register/internal/adapter/seedfile/seedfile.go:12-27`).
- The Register DESIGN.md makes the load a lab-only feature, separate from the Register start: "the Register start does not need a file" (`services/register/DESIGN.md:521-528`).
- The data holder chart already has the same pattern: an init container `seed -if-empty` (order 2, after `migrate` order 1) as the service user, and an existing ConfigMap mounted as a directory into `seed` and `main` with `advancedMounts` (`charts/dataholder/templates/common.yaml:79-98`, `:183-196`). Milestone 1 decision D8 chose "an existing ConfigMap name, mounted as a directory" over a ConfigMap from values ([Data holder image and chart](2026-11-02-data-holder-image-and-chart.md), section 6).
- `charttest.Volume` already decodes `configMap.name` (`charts/internal/charttest/charttest.go:47-55`).

### Facts: size

- The four preset banks of the lab seed, with `"dataRecipients": []`, are 2,722 bytes as compact JSON (measured with `jq -c` on `services/register/seed/example.json`). The full lab seed is 55,655 bytes.
- A ConfigMap holds at most 1 MiB [S1]. A Helm release, with its values, is in one Secret [S7], and a Secret holds at most 1 MiB [S5]. Thus either place holds the seed with a margin of more than 300 times.

### Options for the source of the file

- **A. An existing ConfigMap that the operator makes.** Value `seed.configMap` (optional, default empty). When it is set, the chart adds an init container `seed` (after `migrate`) with `args: [seed, -if-empty, /seed/participants.json]` as the service user, and mounts the ConfigMap at `/seed` in that init container only. A fixed key name (`participants.json`) keeps one value. The operator makes it with `kubectl create configmap <name> --from-file=participants.json=<file>` [S4], or with a manifest in the deployment repository.
  - For: the same pattern as `brand.configMap` (D8). The chart does not know the content of the file. No seed is the default, so a Register with no lab data stays possible, as the DESIGN.md requires. A deployment already makes Secrets beside the chart, so one more manifest is no new mechanism.
  - Against: one more object for the operator to make.
- **B. The file in the values.** Value `seed.participants` (an object); the chart renders a ConfigMap with `toJson`. With `helm install --set-file` [S6] or an Argo CD `fileParameters` entry [S13], the operator can give the JSON file as it is (check that `fileParameters` reads from a second source before you depend on it).
  - For: one values file per deployment, no extra manifest. A change of the values can add a checksum annotation.
  - Against: a second pattern beside D8. A list in a values file replaces the full list [S6], so a deployment copies the full seed to change one host. `-if-empty` loads only once, so a later change of the values has no effect, which a reader does not expect from values.
- **C. A seed preset in the chart.** Value `seed.preset: preset-banks` and a map of base URLs for each brand. The chart has the four records in a file that it reads with `.Files.Get` and renders with `tpl` [S8].
  - For: the deployment gives only host names ("configuration only" in the strongest sense). One copy of the identifiers in the repository.
  - Against: the Register chart then knows the presets of the data holder, and a template language over JSON is harder to review and to test. A fifth brand needs a chart change.

Recommendation: option A. It is the simplest, it reuses the D8 pattern, and it keeps the chart deployment-agnostic with no knowledge of the participants. Option C is the only option that removes the copy of the file in a deployment, but its cost is high for four records that change rarely.

### Options for "the seed file for the four preset banks" with no host of one deployment

Fact: each URI field of a participant is an absolute `https` URI with a host (`services/register/DESIGN.md:70`; `services/register/internal/domain/dataholder.go:192-208`). Thus a seed with placeholders (for example `${WATTLE_BANK_HOST}`) does not load, and the distroless image has no `envsubst`.

Fact: the public and the secure base URIs of a brand are values of the data holder chart (`publicBaseURL`, `secureBaseURL`, `charts/dataholder/values.yaml:21-23`). The secure listener needs TCP pass-through (`charts/dataholder/README.md`, Operation), so a deployment can give the secure API a different port or a different host from the public API. Thus a replacement of one domain is not always sufficient.

- **1. An example seed with `.test` hosts** [S12], for example `https://wattle-bank.test` and `https://wattle-bank.test:8083`. It loads with the Register loader, and a test can prove it. A deployment copies it and changes the URIs.
- **2. A template with placeholders.** It does not load and cannot be tested without a substitution step.
- **3. Option C above.** No copy, but a chart that knows the presets.

Recommendation: option 1, with each URI of a brand on the host `<brand>.test`, so that one `sed` or `jq` replacement of the domain is sufficient for a deployment with the same shape. The README of the Register chart gives the rule that the URIs of each brand agree with `publicBaseURL` and `secureBaseURL` of its data holder release.

Where: `services/register/seed/preset-banks.json`, beside `example.json`. The format belongs to the Register, and the seed file test (`seedfile_test.go:45-50`) already opens files of that directory. An alternative is `charts/register/examples/`; the file is then in the package of the chart unless a `.helmignore` excludes it [S9].

## 2. What the seed holds for each brand

### Facts

- A data holder in the seed file is a legal entity with `legalEntityId`, `legalEntityName` (36 characters at most), `logoUri`, `status` and `dataHolderBrands` (`services/register/seed/example.json`, the first four entries; `services/register/DESIGN.md:61`).
- A brand has `dataHolderBrandId`, `brandName`, `industries`, `logoUri`, `status`, `endpointDetail` and `jwksEndpoint` (`example.json`). The Register requires `logoUri`, `endpointDetail.publicBaseUri` and `industries`. With a `dataHolderBrandId`, it also requires `endpointDetail.version`, `resourceBaseUri`, `infosecBaseUri`, `websiteUri` and `jwksEndpoint`. `productBaseUri` and `extensionBaseUri` are optional (`services/register/internal/domain/dataholder.go:45-53`, `:192-208`) [S11].
- The four lab records:

| Brand | `legalEntityName` | `legalEntityId` | `dataHolderBrandId` | Lab host | `websiteUri` |
|---|---|---|---|---|---|
| Wattle Bank | Wattle Bank Limited | `dd4d6bc6-…` | `d7385786-d8b1-4d4b-bb80-b21616df1a4d` | `wattle-bank` | `https://wattlebank.test` |
| Redgum Bank | Redgum Bank Limited | `830186b8-…` | `4967dd80-2cd2-43c3-8b13-f6faafb684ad` | `redgum-bank` | `https://redgumbank.test` |
| Harbour Bank | Harbour Bank Limited | `00b12e81-…` | `366f7b1c-93c7-4b84-9c92-ab2456614095` | `harbour-bank` | `https://harbourbank.test` |
| Kestrel Bank | Kestrel Bank Limited | `0ed9945f-…` | `1ee54312-624b-4679-bf33-cf4af113f934` | `kestrel-bank` | `https://kestrelbank.test` |

  All are `ACTIVE`, with the industry `banking`, `version` `v1`, the public, product and infosec base URIs on port 8081, the resource base URI on port 8083, the logo `<public>/brand/logo.svg` and the JWKS `<public>/jwks` (`jq` on `example.json`).
- The data holder serves `GET /brand/logo.svg` and `GET /jwks` on its public listener (`services/dataholder/internal/adapter/branddir/handler.go:30`, `pkg/authserver/dataholder.go:21`, `:405`). The `jwksEndpoint` of the brand is the `jwks_uri` of the data holder (`services/dataholder/DESIGN.md:280`).
- The `brandName` of each preset brand directory is the `brandName` of the record (`services/dataholder/brands/*/brand.json`; `compose.yaml:136`).
- The seed format has no certificate field for a data holder brand (`example.json`, brand keys). The Register does not register a server certificate. It issues one from a CSR for a brand that exists, and records only the thumbprint (`services/register/DESIGN.md:260-265`). The O of the CSR is the brand name or the legal entity name (`services/register/internal/domain/servercertificateissue.go:86-112`).
- The lab seed has data recipients only for the tests. Their hosts are `conformance`, `openid-runner` or `.example` hosts (`jq` on `example.json`).
- A data holder becomes ready after one successful request for the software product statuses. An empty list is a successful response (`services/dataholder/internal/adapter/registerstatus/registerstatus.go:83`; [Data holder image and chart](2026-11-02-data-holder-image-and-chart.md), section 2).

### Recommendations

- The deployment seed has the four legal entities, each with one brand, with the identifiers, names and statuses of the lab records, and `"dataRecipients": []`. The test participants of `example.json` are not in it: they are fixtures of the conformance suite with hosts that do not exist outside the lab.
- No certificate at seed: a data holder brand needs none on the Register. The seed is a precondition of the server certificate issue through the operator interface, so the README can give this order: install the Register with the seed, issue a server certificate for each brand, then make the Secrets of the data holders. The Secrets stay out of scope.
- `websiteUri`: keep the `.test` value of the lab (a fixed name, not a deployment value), or the online bank site of the brand when `bankSite` is on. Recommendation: keep `https://<brand>bank.test`; the site URL is a deployment value.

## 3. The values for the four preset banks

### Facts

- A brand release needs `brand.preset` or `brand.configMap`, `brand.id`, `publicBaseURL`, `secureBaseURL`, `register.publicBaseURL`, `register.statusInterval`, `mail.smtpAddress`, `mail.from`, `operator.allowedHosts`, `telemetry.otlpEndpoint`, the database and the Secrets (`charts/dataholder/README.md`, Values and Secrets).
- The values that differ for each preset brand are `brand.preset`, `brand.id`, `publicBaseURL`, `secureBaseURL` and `mail.from` (lab: `no-reply@wattlebank.test`, `compose.yaml:142`). All other values are values of the deployment, the same for all brands or chosen by the operator (for example Secret names and allowed hosts).
- `ci/` of each chart has only values for the tests and `lint:chart`, "not for a deployment" (`charts/dataholder/ci/test-values.yaml:1`). The repository has no `examples/` directory and no `.helmignore` in a chart.
- `lint:chart` lints and validates each chart with `ci/test-values.yaml` and `ci/lint-values.yaml` (`mise.toml:75-85`) [S14].

### Options for the place

- **a. `charts/dataholder/examples/<brand>.yaml`**, one file for each brand, and the seed in `services/register/seed/`. A new convention (AGENTS.md: propose it before use).
- **b. One file `charts/dataholder/examples/preset-banks.yaml`** with four documents. Helm reads one document from each values file, so the operator must split it. Not recommended.
- **c. A table in the chart README** with the five values of each brand. No file to test with `helm template`.

Recommendation: option a. Each file has only the five values that differ for the brand, and a comment that the deployment adds the other values. Its URLs use the same `.test` hosts as the example seed. Also add a `.helmignore` with `ci/` and `examples/`, or accept them in the package (they are small). Decision needed.

### How a test proves them

- A chart test of the data holder renders `ci/test-values.yaml` with each example file on top, and checks that the chart renders, that `DATA_HOLDER_BRAND` is the preset, and that the brand ID and base URLs agree with the record of that brand in the example seed: `dataHolderBrandId`, `publicBaseUri` and `resourceBaseUri`. The test decodes the seed with `encoding/json` into a small local type, because `charts/dataholder` cannot import `services/register/internal` (Go internal packages; ADR 0004). This one test keeps the two copies of the identifiers and the URLs in agreement.
- `charttest.Render` takes only `--set` values and always `ci/test-values.yaml` (`charttest.go:120-157`). A test of an example file needs a small extension: an extra `--values` file.
- `lint:chart` can also render each example file. Optional; the Go test already renders them.

## 4. Tests and verification

- Chart test (Register): with `seed.configMap` set, the pod has two init containers, `migrate` and then `seed`. `seed` has `args: [seed, -if-empty, /seed/participants.json]`, the service user, a read-only mount of the ConfigMap, and no owner credentials. The main container has no seed mount. With no `seed.configMap`, the pod has one init container, as now (`chart_test.go:34-52`). `ci/lint-values.yaml` sets `seed.configMap`, so that kubeconform checks the seed parts.
- Loader test (Register): in `seedfile_test.go`, `preset-banks.json` decodes, `Validate()` passes, each data holder imports (as in `TestTheExampleFileIsValidAndHasEachStatus`, `seedfile_test.go:45-60`), and it has exactly the four preset brands, each `ACTIVE`, `banking`, with each required endpoint. The test also checks that the identifiers and names agree with the four records of `example.json`, so that a change of one file without the other fails.
- Optional: the brand test of the data holder checks that each `brandName` of the seed is the `brandName` of the preset directory.
- Cluster verification. Unit tests prove the wiring and the file, but not "a Register that the chart deploys with the seed file has the four preset brands at start-up". The project constraint permits a temporary verification with temporary resources only (standing authorisation of 2026-11-01). Minimum steps, a subset of the milestone 1 steps (option A of that note):
  1. A temporary namespace, a one-instance CloudNativePG cluster with the database `register`, its users and the managed role `register_app`.
  2. Throwaway CAs and Secrets from `pki init` in a scratch directory.
  3. `kubectl create configmap register-seed --from-file=participants.json=services/register/seed/preset-banks.json`.
  4. `helm install` the published Register chart with `seed.configMap=register-seed` and `console.enabled=true`. Wait for Ready; the `seed` init container log says "participants loaded".
  5. Through `kubectl port-forward`: Get Data Holder Brands Summary (`GET /cdr-register/v1/banking/data-holders/brands/summary`, `x-v: 2`, public TLS, no token) gives the four brands with their `publicBaseUri` [S11]. The full `endpointDetail` is on Get Data Holder Brands, which needs mTLS and a token of a software product, and the seed has none. The console API `GET /api/data-holders/{legalEntityId}` gives the full record with no token (`services/register/internal/adapter/console/console.go:54-55`).
  6. Restart the pod: the `seed` init container logs "participants not loaded" and the Register has the same four brands.
  7. Remove the namespace and the scratch directory.
- Optional: one data holder release with an example values file against this Register becomes Ready. It proves that `brand.id` agrees with the seed on a cluster, but milestone 1 already proved the release, and the chart test proves the agreement.

## 5. Risks and decisions for the requirements pull request

Where the requirements go:

- Fact: `services/register/DESIGN.md` does not mention the chart (`grep -n -i chart`: no lines). The Register chart has its rules in `charts/register/README.md`, `values.yaml` and its tests. The data holder DESIGN.md has a section "Deployment from published artifacts", and its chart README links to it (`services/dataholder/DESIGN.md`, `charts/dataholder/README.md:3`).
- Recommendation: add a short section "Deployment from published artifacts" to `services/register/DESIGN.md` with the seed requirements only (an optional existing ConfigMap, the load before the start with `-if-empty`, the service user, no seed by default), and a bullet in "The chart" of `services/dataholder/DESIGN.md` for the example values of the four presets and their agreement with the seed. A full copy of the earlier Register chart rules into the DESIGN.md is a separate change. Decision needed.

Decisions needed:

- E1. The source of the seed: an existing ConfigMap (A, recommended), the file in the values (B), or a preset in the chart (C).
- E2. Seed on or off by default: off, so that the default Register has no lab data (recommended; the DESIGN.md says that the start needs no file), or required.
- E3. The deployment seed: the four banks only, with no data recipients (recommended), or also some test participants of `example.json`.
- E4. The hosts of the example seed and the example values: `<brand>.test` (recommended) or a different `.test` pattern.
- E5. The place of the files: seed in `services/register/seed/`, values in `charts/dataholder/examples/` (new convention, recommended), and a `.helmignore` or not.
- E6. `websiteUri`: the fixed `.test` site of the lab (recommended) or a deployment value.
- E7. The cluster verification for the milestone: a temporary Register only (recommended), or a Register and one or four data holders.

Risks:

- Two copies of the identifiers and the URLs: the seed and the values of each data holder release. A difference makes a data holder that is ready (an empty status list is a success) but that data recipients cannot find, or that the Register lists with incorrect URLs. The cross-check test covers the repository files only; a deployment can still differ. The README must say it.
- `-if-empty` loads only into an empty Register. A change of the ConfigMap after the first start has no effect. A change needs the operator interface or a new database. The README must say it.
- The seed init container runs at each start. It needs the grant of `register_app` to the service user, as the main container does; with no grant it fails with `permission denied` (milestone 1).
- The CDR server certificate of each brand has the deployment host as its CN, and the Register issues it only for a brand that exists. Thus the seed comes before the data holder Secrets. This is an order for the deployment, not for the chart.
- With no data recipient in the seed, no consent is possible on a deployment until an operator adds one. This is correct for the scope, but a reader can expect a working flow.
- The chart-publish job failed one time with "Tag not found". A Register chart change publishes a new Register chart version.

## Main findings

- The Register chart runs no seed today. A Register from the chart starts with no participants, and its test requires exactly one init container.
- The data holder chart already has the pattern that the Register seed needs: a `seed -if-empty` init container after `migrate`, as the service user, with an existing ConfigMap mounted as a directory.
- The seed of the four preset banks is 2,722 bytes, far below the 1 MiB limit of a ConfigMap or of a Helm release Secret.
- A seed with placeholders does not load, because each URI of a participant must be an absolute `https` URI with a host. An example seed with `.test` hosts loads and can be tested.
- The public and the secure base URIs of a brand can have different hosts or ports in a deployment, so the endpoints of the seed are deployment values, the same as the base URLs of the data holder release.
- A data holder brand needs no certificate on the Register at seed. The seed is a precondition of the issue of its CDR server certificate by the Register.
- Get Data Holder Brands Summary proves the four brands without a token, but only with the public base URI. The full endpoints need Get Data Holder Brands (mTLS and a software product token) or the console API.
- A data holder is ready with an empty status list, so a Register with no data recipients is sufficient for the four banks.
