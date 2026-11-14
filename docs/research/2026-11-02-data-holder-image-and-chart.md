# Data holder image and chart: technical research

- Research date: 2026-11-02
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 1 of the project "The data holder brands deploy to Kubernetes from published artifacts": "One data holder brand runs on a cluster, from the data holder image and Helm chart that CI publishes from main." Acceptance criteria of the project for this milestone: CI publishes the image and the chart from `main` only after the tests of that commit pass, and a temporary deployment of one brand from the published artifacts passes its readiness probe, and its public APIs answer.
- Related notes: [Branded data holder instances](2026-11-01-branded-data-holder-instances.md) (option A, presets in the binary, and the ConfigMap limit), [Four preset banks in the lab](2026-11-02-four-preset-banks-in-the-lab.md). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | GitHub Actions: `jobs.<job_id>.needs` (a skipped or failed need skips the job, unless a status check function is in `if`) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idneeds |
| S2 | GitHub Actions: status check functions (`failure()` is true when an ancestor job fails; `cancelled()`) | https://docs.github.com/en/actions/reference/workflows-and-actions/expressions#status-check-functions |
| S3 | GitHub Packages: the visibility of a new package is private | https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility |
| S4 | GitHub Packages: publish to the Container registry with `GITHUB_TOKEN` and `packages: write` | https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry |
| S5 | Helm: chart dependencies, `helm dependency build` | https://helm.sh/docs/helm/helm_dependency_build/ |
| S6 | Helm: OCI registries, `helm push` | https://helm.sh/docs/topics/registries/ |
| S7 | Helm: the `required` and `fail` template functions | https://helm.sh/docs/howto/charts_tips_and_tricks/#using-the-required-function |
| S8 | bjw-s common library chart (5.x): values schema, `controllers`, `service`, `persistence`, `forceRename` | https://github.com/bjw-s-labs/helm-charts/tree/main/charts/library/common |
| S9 | Kubernetes: ConfigMaps (1 MiB limit, `binaryData`) | https://kubernetes.io/docs/concepts/configuration/configmap/ |
| S10 | Kubernetes: a ConfigMap volume, and a `subPath` mount gets no updates | https://kubernetes.io/docs/concepts/storage/volumes/#configmap |
| S11 | Kubernetes: `kubectl create configmap --from-file` with a directory | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_configmap/ |
| S12 | Kubernetes: init containers run in sequence, and each must complete before the app containers start | https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ |
| S13 | Kubernetes: liveness, readiness and startup probes | https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/ |
| S14 | Kubernetes: Secret types (`kubernetes.io/tls`, `kubernetes.io/basic-auth`) | https://kubernetes.io/docs/concepts/configuration/secret/#secret-types |
| S15 | CloudNativePG: declarative role management (`managed.roles`, `inRoles`) | https://cloudnative-pg.io/documentation/current/declarative_role_management/ |
| S16 | kubeconform | https://github.com/yannh/kubeconform |
| S17 | Renovate: `groupName` and package rules | https://docs.renovatebot.com/configuration-options/#groupname |
| S18 | CDS 1.36.0: Register APIs (Get JWKS at `/cdr-register/v1/jwks` on the Register base URI) | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-apis |

## 1. Image publication

### Facts: the Register image

- The `register-image` job builds and starts the Register with `mise run up register`, checks that the image has no Node, and runs the browser tests of the console (`.github/workflows/ci.yml:308-359`). `mise run up` exports `BUILD_COMMIT`, `BUILD_COMMIT_SHORT`, `BUILD_TIME` and `BUILD_ID` before the build (`mise.toml:22-27`, `mise.toml:272-278`).
- Only on a push to `main`, the job tags the image of the running container with the full and the short commit (`git describe --always --abbrev=40 --exclude='*'` and `git describe --always --exclude='*'`), saves it with `docker save`, and uploads it as an artifact with a retention of 1 day (`ci.yml:361-381`).
- The `register-image-publish` job needs only `register-image`. It runs on `ubuntu-latest`, has `packages: write`, loads the artifact and runs `docker push --all-tags` to `ghcr.io/${{ github.repository }}/register` (`ci.yml:25`, `ci.yml:391-412`).
- `register-image-publish` does not need `lint`, `test`, `conformance` or `web`. Thus the Register image of a commit is published when the browser tests of the console pass, also when the Go tests or the conformance tests of that commit fail (`ci.yml:391-393`).
- A job whose need is skipped or failed is skipped, unless its `if` has a status check function [S1]. `failure()` is true when an ancestor job failed [S2].
- The path filter of `register-image` is the inputs of the console browser tests, not the inputs of the image (`ci.yml:79-82`). The Register `Dockerfile` copies `web/`, `go.mod`, `go.sum`, `pkg/` and `services/register/` (`services/register/Dockerfile:11-28`). Thus a commit that changes only, for example, `services/register/internal/app` or `pkg/cds` gets no published image. The published tags show that only some commits have an image (anonymous `tags/list` on `ghcr.io/lukemundy/cdr-lab/register`: 64 tags, 32 commits).
- The Register image and chart packages are public: GHCR gives an anonymous pull token and the tag list for `lukemundy/cdr-lab/register` and `lukemundy/cdr-lab/charts/register`. The repository is private (`gh repo view`).

### Facts: the data holder image

- The `dataholder-image` job exists. It builds and starts `wattle-bank` with `mise run pki && mise exec -- docker-compose up --detach --build --wait --wait-timeout 120 wattle-bank`, checks that the image has no Node, and runs the browser tests `dataholder-console` and `dataholder-bank` (`ci.yml:461-522`). It does not save, upload or publish the image.
- That step does not use `mise run up`, so it does not export the build variables. Compose then gives `BUILD_COMMIT=unknown`, `BUILD_COMMIT_SHORT=unknown` and `BUILD_ID=local` (`compose.yaml:7-11`, `services/dataholder/Dockerfile:34-41`). An image from this job reports `service.version` `unknown` in its telemetry. No commit message gives the reason for the direct compose call or the 120 s wait.
- Compose starts the Register with `wattle-bank`, because the data holder is ready only after it gets the statuses of the software products from the Register (`compose.yaml:159-164`, `ci.yml:499-500`).
- The path filter of `dataholder-image` is also the inputs of the browser tests (`ci.yml:83-91`). It does not include `services/dataholder/cmd`, most adapters (for example `statusstore`, `registerstatus`, `smtpmail`), most of `pkg/`, `go.mod` or `go.sum`. The data holder `Dockerfile` copies `web/`, `go.mod`, `go.sum`, `pkg/` and `services/dataholder/` (`services/dataholder/Dockerfile:12`, `25-32`).
- The GHCR name follows the Register: `ghcr.io/lukemundy/cdr-lab/dataholder`. No package with that name, or with `cdr-lab/charts/dataholder`, exists now (GitHub API: 404).
- A new package is private when it is first published [S3]. A workflow with `packages: write` can publish to `ghcr.io/OWNER/...` with `GITHUB_TOKEN` [S4].

### Options for "only after the tests of that commit pass"

- **A. The same gate as the Register now.** `dataholder-image-publish` needs only `dataholder-image`. The image is tested only by the browser tests. This does not meet the acceptance criterion, because the Go and conformance tests can fail.
- **B. Need all test jobs.** `dataholder-image-publish` needs `lint`, `test`, `conformance`, `web` and `dataholder-image`, with `if: github.ref == 'refs/heads/main' && github.event_name != 'pull_request' && !cancelled() && !failure() && needs.dataholder-image.result == 'success'` [S1, S2]. A need that the path filter skipped does not stop the publish. A need that fails stops it.

Recommendation: option B. Use the same gate for `register-image-publish` in the same change, because "tested image" then has one meaning. A skipped job means that its inputs did not change, so the result of the previous commit applies.

Note: the conformance job builds its own images (`mise run up`, `ci.yml:221-224`). The published image is the image of the `dataholder-image` job, from the same commit and the same `Dockerfile`, but not the same build. The builds are not proved to be byte-identical. Recommendation: accept this for the lab; it is also true for the Register now.

### Options for the path filter

- **A. Keep the browser-test filter.** Some commits on `main` get no image.
- **B. On `main`, run the image job when any input of the image changes:** the shared inputs, `go.mod`, `go.sum`, `pkg`, `services/dataholder`, `web`, `.dockerignore`, `compose.yaml` and `lab`. Pull requests can keep the narrow filter, or use the same filter for simplicity.

Recommendation: option B with one filter for pull requests and `main`, for the data holder and the Register. Decision needed: the Register change is outside the milestone, but the same rule applies.

### Recommendation for the job

- Change the start step of `dataholder-image` to `mise run up wattle-bank` with `BUILD_ID: ${{ github.run_id }}`, as the Register job does, so that the image reports its commit. If 60 s (`mise.toml:277`) is not sufficient for the wait, add a timeout setting to the `up` task. Do not keep the direct compose call.
- Add the save and upload steps (artifact `dataholder-image`) and a `dataholder-image-publish` job on `ubuntu-latest` with only `packages: write`, the same as `register-image-publish` (`ci.yml:361-412`). Add `DATAHOLDER_IMAGE: ghcr.io/${{ github.repository }}/dataholder` beside `REGISTER_IMAGE` (`ci.yml:25`).
- Escalation point: after the first publish, the user must make the packages `cdr-lab/dataholder` and `cdr-lab/charts/dataholder` public [S3], as for the Register. Until then, an anonymous pull from the cluster fails.

## 2. The chart

### Facts: the Register chart

- `charts/register` depends on the bjw-s `common` library 5.3.0 from `oci://ghcr.io/bjw-s-labs/helm` (`charts/register/Chart.yaml:7-10`, `Chart.lock`). Git ignores the dependency archive (`.gitignore:45`). `mise run chart:deps` runs `helm dependency build --skip-refresh charts/register` (`mise.toml:65-67`) [S5].
- `values.yaml` has only the values of the Register. An empty value with no default is required (`charts/register/values.yaml:1-2`). One template converts these values into the values of `common`, merges them with `mustMergeOverwrite`, and calls `bjw-s.common.loader.all` (`charts/register/templates/common.yaml:33-152`).
- `register.required` fails when a value is empty, because Helm `required` accepts an empty list (`common.yaml:1-7`) [S7].
- One Deployment, one replica, strategy `Recreate`, because the migrations take no lock and the Register is the issuing CA (`common.yaml:62-66`). The production gap of this is in `docs/production-gaps.md:35`.
- The migrations run in an init container (`args: [migrate]`) as the owner user. The main container (`args: [serve]`) gets only the service user (`common.yaml:67-83`). The Register chart chose an init container over a Helm hook, because the order is the same in Argo CD and in plain Helm, and one replica has no race.
- The pod has no service account token, runs as 65532, with seccomp `RuntimeDefault`. Each container has no privilege escalation, a read-only root filesystem and no capabilities (`common.yaml:16-21`, `53-60`).
- Probes: liveness `/livez` and readiness `/readyz` on port 8080. The startup probe is off (`common.yaml:112-124`).
- One ClusterIP Service with the API ports. Port 8080 is not on the Service. The console port is on the Service only when the console is on (`common.yaml:127-139`).
- `forceRename` is a literal template string, because the library can render the full name only after it loads its defaults (`common.yaml:130-131`). Without it, the library adds `-main` to the name when the values add a second Service; the certificates and the CRL URL give that name (`charts/register/README.md:16`, test `chart_test.go:344-354`).
- Each Secret is an existing Secret, mounted read-only: `kubernetes.io/tls` Secrets for TLS and the intermediate CA, `ca.crt` for the CA bundle and the optional public root, `token` for the operator token, `username`/`password` (basic-auth, for example the Secrets of CloudNativePG) for the database users (`charts/register/values.yaml:32-53`, `common.yaml:9-31`, `141-148`) [S14].
- `OTEL_EXPORTER_OTLP_ENDPOINT` is a required value (`common.yaml:49`). `image.repository` has a default; `image.tag` is required; `image.digest` is optional (`values.yaml:4-9`).
- The chart requirements are in `charts/register/README.md`, `values.yaml` and the tests `charts/register/chart_test.go`. They are not in `services/register/DESIGN.md` (no mention of the chart there). A ticket gave the behaviour.
- `charts/register/README.md:14` gives the role grant: the Register can start only after the operator grants `register_app` to the service user. With CloudNativePG, declare the role as a managed role and give it in `inRoles`, because CloudNativePG does not try a grant again when the role did not exist [S15].
- The verification of the Register chart ran the chart in a temporary namespace with a CloudNativePG cluster and Secrets from `mise run pki`. The Register failed with `permission denied` until the grant.

### Facts: what the data holder reads

The data holder has one binary with the subcommands `serve` (APIs and workers), `api`, `worker`, `migrate` and `seed [-if-empty] [<file.json>]` (`services/dataholder/cmd/dataholder/main.go:1-32`, `internal/root/run.go:54-89`). The default is `serve`.

| Setting | Read by | Required | Source |
|---|---|---|---|
| `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE` | all | yes (libpq) | `run.go:44`, README Data |
| `DATABASE_MIGRATE_ON_START` | serve, api, worker, seed | no | `subcommands.go:89-95` |
| `OTEL_*` | all | no; with no endpoint, export is off | `pkg/platform/telemetry/export.go:26-40` |
| `DATA_HOLDER_BRAND` or `DATA_HOLDER_BRAND_DIR` (exactly one) | api, serve, seed | yes | `config.go:228-271` |
| `PUBLIC_BASE_URL`, `PUBLIC_TLS_CERT_FILE`, `PUBLIC_TLS_KEY_FILE` | api | yes | `config.go:118-131`, `167-176` |
| `SECURE_BASE_URL`, `SECURE_TLS_CERT_FILE`, `SECURE_TLS_KEY_FILE`, `CDR_CA_BUNDLE_FILE` | api | yes | `config.go:134-136`, `178-201` |
| `REGISTER_JWKS_URL` | api | yes | `config.go:137-139` |
| `SMTP_ADDRESS`, `MAIL_FROM` (address only) | api | yes | `config.go:141-156` |
| `OPERATOR_TOKEN`, `OPERATOR_ALLOWED_HOSTS` | api | yes | `pkg/platform/operatorhttp/operatorhttp.go:28-33` |
| `FIXED_OTPS` | api | no (lab-only) | `config.go:162-164` |
| `CRL_MAX_AGE` | api | no; default 1 minute | `config.go:278-292` |
| `CONSOLE_ENABLED`, `CONSOLE_ALLOWED_HOSTS` | api | no; hosts required when on | `pkg/platform/consolehttp/consolehttp.go:43-72` |
| `BANK_SITE_ENABLED`, `BANK_SITE_ALLOWED_HOSTS` | api | no; hosts required when on (lab-only) | `internal/adapter/banksite/banksite.go:26-31` |
| `DATA_HOLDER_BRAND_ID` | worker, serve; api when the bank site is on | yes | `config.go:88-93`, `212-220` |
| `REGISTER_PUBLIC_BASE_URL` | worker, serve | yes | `config.go:96-100` |
| `REGISTER_STATUS_INTERVAL` | worker, serve | yes, more than 0 and at most 4m, no default | `config.go:103-110`, `run.go:28` |
| `SSL_CERT_FILE` | all (Go system roots) | no | `compose.yaml:44-46` |

Other facts:

- Listeners: probes 8080 (HTTP), public 8081 (TLS), operator 8082 (HTTP), secure 8083 (mTLS), console 8084 (HTTP), online bank site 8085 (TLS, with the public certificate) (`run.go:18-25`, `services/dataholder/DESIGN.md` Online bank site: Listener and security). The `Dockerfile` exposes 8080 to 8085 (`Dockerfile:47`).
- The readiness probe fails until the database answers and, in a process that serves the APIs, until the data holder has a saved response of the software product statuses from the Register (`internal/root/process.go:114-135`). An empty list from the Register is a saved response (`statusstore.go:49-72`). The status job runs at start and then at the interval (`statusjobs.go:31-37`). Thus a data holder becomes ready only when it can reach the public APIs of a Register and trusts its TLS certificate.
- The signing keys are created in the database at start (`process.go:41-44`). The chart needs no signing key Secret. The production gap "private signing keys in the database" applies (`docs/production-gaps.md:26`).
- The application role is `dataholder_app` (`internal/migrations/00019_create_application_role.sql:6`). A role is global to the server, so all brands on one PostgreSQL server share it, and the `CONNECT` privilege of each service user is the boundary (ADR 0020, Decision).
- `seed -if-empty` loads the catalogue of the brand, and exits with no error when products exist (`subcommands.go:31-72`). The lab runs it as the service user, after the grants (`compose.yaml:179-191`). The product load is a lab-only feature, separate from the start of the data holder (`DESIGN.md` Lab-only features).
- The authorisation sessions and the online bank site sessions are in the memory of the process (`docs/production-gaps.md:16`, `:46`). One process serves one brand correctly; two replicas do not share a flow.
- SMTP is dialled only when an email is sent (`internal/adapter/smtpmail/smtpmail.go:105`). The setting is required at start, but the server is not needed for the readiness probe.
- The brand directory: a preset needs one setting; a user brand is a mounted directory. The loader opens the directory with `os.OpenRoot`, which permits the relative links of a ConfigMap volume and refuses a link out of the directory (`config.go:248-256`). The data holder ignores entries that start with `..` (ADR 0023, Decision: Files). Each preset directory is approximately 44 KB (`du` of `services/dataholder/brands/wattle-bank`).
- A ConfigMap holds at most 1 MiB, and `binaryData` holds data that is not UTF-8 [S9]. `kubectl create configmap --from-file=<dir>` adds each regular file of the directory as a key [S11]. A ConfigMap volume gives each key as a symbolic link through `..data` [S10]. A `subPath` mount gets no updates [S10].
- Helm `.Files` reads only files in the chart, so a published chart cannot make a ConfigMap from a brand directory in the repository [S5, S6]. (The 2026-11-01 note gives the same conclusion.)

### The values, mapped to the Register chart

Recommendation (each "required" has no default, and the render fails with its name):

| Value | Required | Register equivalent | Data holder setting |
|---|---|---|---|
| `image.repository` (default `ghcr.io/lukemundy/cdr-lab/dataholder`), `image.tag`, `image.digest` | tag | same | image |
| `brand.preset` or `brand.configMap` (exactly one) | one of two | none | `DATA_HOLDER_BRAND`, or `DATA_HOLDER_BRAND_DIR` at a fixed mount path |
| `brandID` | yes | none | `DATA_HOLDER_BRAND_ID` |
| `publicBaseURL`, `secureBaseURL` | yes | same | `PUBLIC_BASE_URL`, `SECURE_BASE_URL` |
| `register.publicBaseURL` | yes | none | `REGISTER_PUBLIC_BASE_URL`, and `REGISTER_JWKS_URL` (see decision D4) |
| `register.statusInterval` | yes (see decision D5) | none | `REGISTER_STATUS_INTERVAL` |
| `crlMaxAge` | no | none | `CRL_MAX_AGE` |
| `mail.smtpAddress`, `mail.from` | yes | none | `SMTP_ADDRESS`, `MAIL_FROM` |
| `operator.allowedHosts`, `operator.tokenSecret` | yes | same | `OPERATOR_*` |
| `console.enabled` (default `false`), `console.allowedHosts` | hosts when on | same | `CONSOLE_*` |
| `bankSite.enabled` (default `false`), `bankSite.allowedHosts` | hosts when on | none | `BANK_SITE_*` |
| `telemetry.otlpEndpoint` | yes | same | `OTEL_EXPORTER_OTLP_ENDPOINT` |
| `database.host`, `.name`, `.port` (5432), `.sslMode` (`require`), `.ownerSecret`, `.serviceSecret` | host, name, Secrets | same | `PG*` |
| `pki.publicTLSSecret`, `pki.secureTLSSecret`, `pki.caBundleSecret` | yes | same | the three TLS settings and `CDR_CA_BUNDLE_FILE` |
| `pki.publicRootSecret` | no | same | `SSL_CERT_FILE` |

The chart makes no Secrets. It needs: two basic-auth database Secrets, two `kubernetes.io/tls` Secrets (public, which the bank site also uses, and secure), one CA bundle Secret (`ca.crt`), one operator token Secret (`token`), and optionally the public root (`ca.crt`) and the brand ConfigMap. It needs no intermediate CA and no signing key.

Templates, the same as the Register:

- One Deployment, one replica, `Recreate`, with the least-privilege pod and container settings.
- Init containers in sequence [S12]: `migrate` as the owner user, with no mounts; then `seed -if-empty` as the service user, with the brand setting and mount only (see decision D3).
- The main container `serve` (APIs and workers in one process, as in compose). Not `api` and `worker` in two Deployments: the APIs must be one replica because of the in-memory sessions, and a second Deployment adds a second set of settings with no gain in the lab.
- Probes `/livez` and `/readyz` on 8080. Recommendation: no startup probe, as the Register, but the README says that the pod is not ready until it reaches the Register.
- One ClusterIP Service with public 8081, operator 8082, secure 8083, and console 8084 and bank site 8085 only when on. The same `forceRename` workaround, because the certificates and the base URLs give the Service name.
- A production-gap entry for the data holder chart: one pod and `Recreate`, because `dataholder migrate` takes no lock and the sessions are in memory; each upgrade stops the brand.
- The chart README gives the role grant of `dataholder_app`, with the CloudNativePG managed role advice of the Register chart, and that all brands on one server share the role.
- Behind a TLS-terminating proxy on port 443, `operator.allowedHosts`, `console.allowedHosts` and `bankSite.allowedHosts` are host names with no port (`charts/register/README.md:16`). Note: the public and secure listeners terminate TLS in the data holder, and the secure listener needs mTLS end to end, so a proxy cannot terminate TLS for 8083.

## 3. The ADR 0018 correction

- Fact: ADR 0018 says "The Node versions in `mise.toml` and in the digest of the Node build image must agree. Renovate updates each one separately." (`docs/adr/0018-container-images-with-a-web-ui-build-stage.md:33`).
- Fact: `renovate.json:48-53` puts `node` and `docker.io/library/node` in one group `node` for major, minor and patch updates, so that one pull request updates `mise.toml` and the image together (`lint:versions`). `renovate.json:38-41` gives `node`, `pnpm` and the Node image the same minimum release age of 3 days, so the two are available to the group at the same time. `renovate.json:43-47` does the same for Go (`go` and `docker.io/library/golang`). A digest-only update of an image goes to the weekly group `docker digests` (`renovate.json:26-31`); it does not change a version [S17].
- Correction: "Renovate updates the Node version of `mise.toml` and of the Node image in one pull request (`renovate.json`, group `node`), and `mise run lint:versions` fails when they do not agree." The same is true for Go. Use the `write-adr` skill before the edit; a small in-place edit is the accepted practice for this kind of correction.

## 4. CI for the chart

- Fact: `lint:chart` runs `helm lint --strict` and `helm template ... | kubeconform -strict -summary -kubernetes-version 1.36.0` for `charts/register` only (`mise.toml:69-78`) [S16]. `chart:deps` builds the dependency of `charts/register` only (`mise.toml:65-67`). `test:go` depends on `chart:deps` and runs all Go tests, also the chart tests (`mise.toml:114-118`).
- Fact: the `chart` job runs `mise run lint:chart` and `go test ./charts/...` when `charts`, `go.mod`, `go.sum` or the shared inputs change (`ci.yml:93`, `414-430`).
- Fact: `chart-publish` needs only `chart`, runs on `ubuntu-latest` with `contents: read` and `packages: write`, packages `charts/register` as `0.0.0-g<short commit>` with the short commit as the appVersion, and pushes to `oci://ghcr.io/lukemundy/cdr-lab/charts` (`ci.yml:432-459`).
- Fact: `chart-publish` failed one time with `failed to perform "Tag" … not found`, and passed when it ran again. This research found no primary source for the cause. If it fails again, run `helm push --debug`.
- Recommendations:
  - `chart:deps` and `lint:chart` loop over each directory in `charts/`, with the `ci/test-values.yaml` of each chart, and a set of `--set` values that turns on each optional part (console, bank site, public root, brand ConfigMap).
  - `chart-publish` packages and pushes each chart in `charts/`. Option: push only the charts whose directory changed. Recommendation: push each chart; it is simpler, and a version of each chart for each chart commit is harmless.
  - The chart filter stays `charts go.mod go.sum` and the shared inputs; the new chart is under `charts/`.
  - The chart publish does not need the image tests: the chart does not point to an image of its own commit (`image.tag` is required). Recommendation: keep it so.
  - Renovate updates the `common` dependency of each `Chart.yaml` separately. Option: a group for `common`, so the two charts use one version. Recommendation: add it; the two charts then have one set of library behaviour.

## 5. The temporary verification on a Kubernetes cluster

The minimum proof: the published chart, at the published version, with the published image, makes a pod that becomes Ready, and Get Products and the OpenID Provider configuration answer on the public listener.

Facts that limit the steps:

- The pod becomes Ready only after it gets the statuses from a Register that it trusts (section 2). Thus the verification needs a reachable Register.
- The data holder checks only that its keys and certificates load and that the CA bundle parses (`config.go:167-201`). Throwaway CAs from `go run ./pki/cmd/pki init -dir <tmp>` and certificates from `pki issue-public` and `pki issue -type server` are sufficient (`pki/cmd/pki/main.go:23-25`, `mise.toml:181-199`).
- The chart has no PostgreSQL. The verification of the Register chart used a CloudNativePG cluster in a temporary namespace.
- The image and the chart must be pullable with no credential (section 1).

The Register is a temporary Register from the published Register chart in the same namespace. It is fully isolated. It needs a second database and the intermediate CA Secret from the throwaway `pki` directory. The verification of the Register chart already proved it.

Steps:

1. Make sure that the image `ghcr.io/lukemundy/cdr-lab/dataholder:<short commit>` and the chart `oci://ghcr.io/lukemundy/cdr-lab/charts/dataholder:0.0.0-g<short commit>` exist and are public. If GHCR refuses an anonymous pull, stop and ask the user to make the packages public.
2. Create a temporary namespace, for example `p-cdr-10-verify`.
3. Create a CloudNativePG `Cluster` with one instance, or one `postgres` pod with an init script. Create the databases `register` and `wattle_bank`, an owner user with `CREATEROLE` and a service user for each, and the managed roles `register_app` and `dataholder_app` with `inRoles` for the service users [S15]. A migration accepts a role that exists (ADR 0020).
4. In a scratch directory outside the repository: `pki init`, then `issue-public` for the public names of the two releases (the Service DNS names in the namespace), `issue -type server` for their secure certificates, and the CA bundle. Create the TLS, CA bundle, public root, intermediate and operator token Secrets in the temporary namespace only.
5. `helm install` the published Register chart (latest published version, with a tag that has an image), with the throwaway Secrets and the temporary OTLP endpoint value. Wait until it is Ready. (The Register has no participants; the data holder needs only an empty status list.)
6. `helm install` the published data holder chart with `brand.preset=wattle-bank`, a brand ID (any UUID; the Register record is milestone 2), `register.publicBaseURL` of the temporary Register, `pki.publicRootSecret` (the throwaway public root), a placeholder SMTP address, and the Secrets.
7. Wait for the pod condition `Ready` (`kubectl wait --for=condition=Ready`). Record the init container logs: migrate and seed completed.
8. Through `kubectl port-forward` to 8081: `GET /cds-au/v1/banking/products` with `x-v` gives 200 and the products of the brand, and `GET /.well-known/openid-configuration` gives the issuer of `publicBaseURL`. Use the throwaway public root as the trust anchor.
9. Optional: check the image commit in the pod logs (`service.version`).
10. `helm uninstall` both releases, and delete the namespace. Delete the scratch `pki` directory. Make sure that nothing stays (`kubectl get ns`).

Note: OTLP export with no reachable collector logs errors but does not stop the readiness probe. A test endpoint in the namespace is not necessary.

## 6. Risks and decisions for the requirements pull request

Where the requirements go:

- Fact: the Register chart has its requirements in `charts/register/README.md`, `values.yaml` and its tests; `services/register/DESIGN.md` does not mention the chart. The project acceptance criterion says "The chart README gives each value and each Secret that a deployment of a brand needs."
- Recommendation: put the milestone requirements in a new `charts/dataholder/README.md` (with the CI requirements that it needs), and add one line to `services/dataholder/DESIGN.md` (Non-functional requirements: Local run) that the data holder also deploys from a published image and chart. Decision needed: AGENTS.md says the requirements go in the `DESIGN.md` of the component. A chart is a component under `charts/`, not a service. Option: a `charts/dataholder/DESIGN.md`. This is a new convention, so propose it before use.

Decisions needed:

- D1. The publish gate (section 1): need all test jobs (recommended), also for the Register image.
- D2. The image path filter on `main` (section 1): all image inputs (recommended), also for the Register image.
- D3. The product seed: an init container `seed -if-empty` always (recommended, because each brand directory has a catalogue and `-if-empty` makes it safe at each start), or a value that turns it on, or no seed in the chart. The seed is a lab-only feature, and the data holder is a mock, so a mock chart can run it.
- D4. `REGISTER_JWKS_URL`: derive it from `register.publicBaseURL` and the CDS path `/cdr-register/v1/jwks` [S18] (recommended: one value, and the path is fixed by the CDS), or a second value.
- D5. `REGISTER_STATUS_INTERVAL`: required with no default (recommended: the service has no default, and the memory rule says "make the setting required"), or a chart default of `1m`.
- D6. `FIXED_OTPS`: not a chart value (recommended: a fixed OTP is a test backdoor, and only the OpenID suite in the lab uses it), or an optional value.
- D7. The bank site: a value with default `false` (recommended, the same as the console), so that milestone 2 can turn it on for the four presets.
- D8. The brand ConfigMap: an existing ConfigMap name, mounted as a directory, not with `subPath` (recommended), or the chart makes the ConfigMap from values (base64 in values is hard to review).

Risks:

- The data holder is not ready without a Register (section 2). A deployment order or a Register outage makes the pod not ready. The README must say it.
- The role grant of `dataholder_app` must exist before the main container and the seed start, or they fail with `permission denied` (the verification of the Register chart).
- The secure listener needs mTLS end to end. An ingress or a proxy that terminates TLS cannot front 8083; the deployment needs a TCP pass-through or a LoadBalancer Service. This is out of scope for the chart (ClusterIP only) but the README must say it.
- The data holder image job starts the Register as well; the save step must save only the data holder image.
- The new GHCR packages are private until the user changes them (escalation point).
- The flaky "Tag not found" of `chart-publish` can occur on the first push of the new chart.

## Main findings

- The Register image is published when only the browser tests of its console pass. The Go and conformance tests do not gate it, and its path filter skips commits that change only Go code outside the console.
- The data holder image job builds with no commit in the binary, because it does not use `mise run up`.
- A data holder becomes ready only after it gets the software product statuses from a Register that it trusts. A verification on a cluster needs a Register.
- ADR 0018 is not correct about Renovate: one group updates the Node of `mise.toml` and the Node image together.
