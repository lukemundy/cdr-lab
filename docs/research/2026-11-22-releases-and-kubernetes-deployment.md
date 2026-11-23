# Releases and a long-lived Kubernetes deployment: technical research

- Research date: 2026-11-22
- File and line references are to the repository on the research date (`main` at `1a45e0b`).
- Purpose: the facts for the requirements step of the project "Teams run cdr-lab in their own Kubernetes clusters from versioned releases". The project decided these items: one version `vX.Y.Z` for all artifacts; a release is a manual action and pushes no commits to `main`; no data guarantee before 1.0.0; one umbrella chart whose values define the seed data; cert-manager gives the public TLS certificates, and the Register gives the CDR certificates; one CLI, `cdrlab`, replaces the `pki` CLI; a Kubernetes development environment is a supported environment, together with compose and CI.
- Related notes: [Data holder image and chart](2026-11-02-data-holder-image-and-chart.md), [Preset banks from the charts](2026-11-02-preset-banks-from-the-charts.md). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source, a test or the code says it. Each fact has a citation. "Tested" means that the research ran the command on a scratch clone, not on the repository.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.
- Each part has its own sources table. The source IDs start with A (part 1), B (part 2), C (part 3) and D (part 4).

## Summary of the recommendations

Releases and the changelog (part 1):

- Use cocogitto in tag-only mode. Add `tag_prefix = "v"` to `cog.toml`: without it, cog does not see `vX.Y.Z` tags (tested). `cog bump --auto --dry-run` gives the next version with no commit, and `cog changelog <prev>..HEAD -t <template>` gives the release notes before the tag exists (tested). On 0.x, a breaking change gives a minor bump, and cog has no setting to change this ([section 1.1](#11-cocogitto)).
- Put the changelog in the GitHub Release notes only, and optionally attach the full changelog as a release asset. A committed `CHANGELOG.md` needs a commit to `main`. Add a tag ruleset for `v*`, and turn on immutable releases ([section 1.3](#13-changelogmd-or-github-release-notes-only)).
- Squash merges keep only the PR title, so commit footers do not reach `main`. Find "needs a fresh deployment" from the diff: a changed, renamed or deleted migration file, or a changed entry of the golden file of the event payloads ([section 1.4](#14-mark-this-release-needs-a-fresh-deployment)).
- For each service, promote the image of the last commit that had an image build, with `crane tag` on its digest. `docker buildx imagetools create` changes the digest of a single-platform image unless you add `--prefer-index=false` ([section 1.5](#15-promote-a-tested-image-without-a-rebuild)).
- `helm package --version` does not change the dependency versions of an umbrella chart. Set the versions of all `Chart.yaml` files in the CI workspace, and use `file://` dependencies ([section 1.6](#16-helm-charts)).
- Build the CLI binaries with a plain `go build` loop and `gh release create`. `go install github.com/lukemundy/cdr-lab/<path>@vX.Y.Z` works with a root tag ([section 1.7](#17-go-cli-release-binaries)).

Supply chain (part 2):

- All three pinned base images are multi-arch indexes with arm64. Cross-compile on the current amd64 runner with `--platform=$BUILDPLATFORM`. Estimate: 20 to 40 seconds more on the `images` job (not measured). The arm64 image is not tested ([section 2.1](#21-multi-arch-images-linuxamd64-and-linuxarm64)).
- The release workflow must sign: a release chart has a new digest, and the CLI binaries are built at release ([section 2.2](#22-signing-with-cosign-keyless)).
- Use `actions/attest` for provenance (SLSA Build Level 2) and the SBOM, and also `cosign sign` for the images and charts, because Flux verifies cosign signatures, not GitHub attestations ([section 2.3](#23-github-artifact-attestations)).
- A Syft scan of the image finds the Go modules but not the npm packages, which are bundled in the binary ([section 2.4](#24-sbom-generation)).
- Helm does not verify cosign signatures. Consumers run `cosign verify` on the chart digest, then install by digest ([section 2.5](#25-helm-chart-verification)).

Certificates (part 3):

- Do not use cert-manager for CDR certificates. The CA issuer needs a second copy of the intermediate key and skips the checks, the events and the client registration of the Register. An external issuer needs CRDs, a controller, cluster-scoped RBAC and a cluster-wide approval change, and it gives no path for compose ([section 3.1](#31-can-the-cert-manager-ca-issuer-make-a-certificate-with-the-cdr-profile), [section 3.2](#32-a-cert-manager-external-issuer-for-cdr-lab)).
- One `cdrlab` command sends a CSR to the existing Register issue path and writes files (compose) or a Secret (cluster). An install Job runs it one time, and a CronJob renews. The operator API needs no new endpoint ([section 3.3](#33-request-and-renew-cdr-certificates-from-the-register-in-a-cluster-with-no-host-step-and-no-cert-manager), [section 3.6](#36-recommended-design)).
- The services load their certificates only at start. A reload in `pkg/cds` is necessary, so that a renewed certificate takes effect with no restart ([section 3.3](#33-request-and-renew-cdr-certificates-from-the-register-in-a-cluster-with-no-host-step-and-no-cert-manager)).
- An idempotent install Job runs `cdrlab pki init` and writes the CA Secrets only when they are absent ([section 3.4](#34-the-lab-ca-roots-in-a-cluster)).

Exposure, operator access, CI and PostgreSQL (part 4):

- `TLSRoute` is in the Gateway API Standard channel since v1.5.0, and the common implementations pass its passthrough tests. ingress-nginx is retired. The umbrella chart gives optional `TLSRoute` templates and one `HTTPRoute` for the CRL path only ([section 4.1](#41-exposure-of-the-tls-and-mtls-listeners)).
- Interim operator access: `kubectl port-forward` with Kubernetes RBAC is the default. oauth2-proxy is a documented option for a browser URL to the console. It cannot pass the static operator token, so the operator API stays on port-forward ([section 4.2](#42-interim-auth-for-the-operator-console-and-the-operator-interface)).
- CI: kind on the standard runner. The services call back to the conformance suite, so the suite runs as a Job in the cluster. This changes CONF-0010 and CONF-0037 ([section 4.3](#43-a-kubernetes-cluster-in-github-actions-ci)).
- PostgreSQL stays outside the chart. The docs give CloudNativePG, and CI uses a plain PostgreSQL with the lab SQL files. Do not use Bitnami ([section 4.4](#44-postgresql-for-the-umbrella-chart)).

## Items for the requirements step

Decided after the research:

1. A release finds "needs a fresh deployment" from the diff only: a changed, renamed or deleted migration, or a changed entry of an event golden file. The golden file is a new convention, and it is approved. The release check compares only the entries that the earlier file already had, so a new event type does not mark a release as fresh ([section 1.4](#14-mark-this-release-needs-a-fresh-deployment)). The squash settings stay as they are.
2. A release uses, for each service, the image of the last commit that had an image build ([section 1.5](#15-promote-a-tested-image-without-a-rebuild)).
3. `service.version` of an image stays the short commit. Only the CLI gets the release version ([section 1.8](#18-stamp-the-semver-version-into-binaries-and-images)).
4. Compose moves from the offline `pki issue` to the Register issue path, so that the Register is the only issuer of CDR certificates in both environments ([section 3.6](#36-recommended-design)).

These items have no decision yet:

1. Whether `image.tag` gets `.Chart.AppVersion` as its default in the release charts ([section 1.6](#16-helm-charts)).
2. Whether the SBOM must list the npm packages in version 1 ([section 2.4](#24-sbom-generation)).
3. Whether the Register issues its own secure certificate at start ([section 3.6](#36-recommended-design)).
4. Whether the umbrella chart puts all releases in one namespace ([section 3.4](#34-the-lab-ca-roots-in-a-cluster)).
5. The host names and ports of the base URLs behind a shared Gateway ([section 4.1](#41-exposure-of-the-tls-and-mtls-listeners)).

## 1. Versions, changelog and release artifacts

Date: 2026-11-22. Repository state: `main` at `1a45e0b`, no tags (`gh api repos/lukemundy/cdr-lab/tags` gives 0).

Tool versions that the research checked:

| Tool | Version | How |
|---|---|---|
| cocogitto (`cog`) | 7.0.0 (latest release, 2026-03-04) | `mise.toml:5`, `cog --version`, tests on a scratch clone |
| Helm | v4.3.2 | `mise.toml:11`, `helm version`, tests in scratch |
| Go | 1.27.4 | `go version`, test module in scratch |
| git-cliff | v2.14.2 (latest release) | `gh release view`, docs only |
| release-please | v17.11.2 (latest release) | `gh release view`, docs only |
| semantic-release | v26.0.1 (latest release) | `gh release view`, docs only |
| svu | v3.4.4 (latest release) | `gh release view`, docs only |
| GoReleaser | v2.20.3 (latest release) | `gh release view`, docs only |
| crane (go-containerregistry) | v0.22.1 (latest release) | `gh release view`, docs only |
| oras | v1.4.0 (latest release) | `gh release view`, docs only |
| docker buildx | v0.39.1 (latest release) | `gh release view`, docs only |

"Tested" in this note means: the research ran the command on a scratch clone in the scratchpad, with test tags and empty commits. The research did not change the repository.

### Sources of part 1

| ID | Source | URL |
|---|---|---|
| A1 | cocogitto: configuration reference | https://docs.cocogitto.io/reference/config.html |
| A2 | cocogitto: template reference | https://docs.cocogitto.io/reference/template.html |
| A3 | cocogitto 7.0.0 source: `conventional/bump.rs` | https://github.com/cocogitto/cocogitto/blob/7.0.0/crates/cocogitto/src/conventional/bump.rs |
| A4 | cocogitto 7.0.0 source: `changelog/serde.rs`, `changelog/release.rs`, `template/simple.tera` | https://github.com/cocogitto/cocogitto/tree/7.0.0/crates/cocogitto/src/conventional/changelog |
| A5 | git-cliff: bump configuration | https://git-cliff.org/docs/configuration/bump |
| A6 | release-please: README and manifest releaser docs | https://github.com/googleapis/release-please |
| A7 | semantic-release: plugins | https://semantic-release.gitbook.io/semantic-release/usage/plugins |
| A8 | svu: README | https://github.com/caarlos0/svu |
| A9 | Docker: `docker buildx imagetools create` | https://docs.docker.com/reference/cli/docker/buildx/imagetools/create/ |
| A10 | crane: `crane tag` | https://github.com/google/go-containerregistry/blob/main/cmd/crane/doc/crane_tag.md |
| A11 | oras: `oras tag` | https://oras.land/docs/commands/oras_tag/ |
| A12 | Docker: attestation storage | https://docs.docker.com/build/metadata/attestations/attestation-storage/ |
| A13 | Sigstore: signing containers | https://docs.sigstore.dev/cosign/signing/signing_with_containers/ |
| A14 | Go modules reference (`go install`, module subdirectories, `+dirty`) | https://go.dev/ref/mod |
| A15 | GoReleaser Pro features | https://goreleaser.com/pro/ |
| A16 | GoReleaser: release customization (`--release-notes`, `mode`, `draft`) | https://goreleaser.com/customization/release/ |
| A17 | GitHub: `GITHUB_TOKEN` | https://docs.github.com/en/actions/concepts/security/github_token |
| A18 | GitHub: immutable releases | https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases |
| A19 | Helm 4.3.2 CLI help (`helm package --help`, `helm dependency update --help`) | local `helm` binary |
| A20 | cocogitto 7.0.0 CLI help (`cog bump --help`, `cog changelog --help`) | local `cog` binary |

### 1.1. cocogitto

#### Facts

- `cog bump` has `--auto`, `--dry-run` ("print the target version. No action taken"), `--disable-bump-commit`, `--package` (monorepo) and `--pre` (A20).
- Tested: `cog bump --auto --dry-run` computes the next version from tags and commits only. It makes no commit, no tag and no file change. It reads no version file.
- Tested: the repository `cog.toml` has no `tag_prefix`. Without it, cog does not see a `v0.1.0` tag as a version: `cog get-version` stays at "No version yet". With `tag_prefix = "v"` (A1), cog reads `vX.Y.Z` tags and prints `v0.2.0` style output.
- Tested: with no tags, `cog bump --auto --dry-run` prints `v0.1.0` (with the prefix). This agrees with "start at 0.1.0" with no extra configuration.
- Tested: `--dry-run` fails when the working tree has changes ("Untracked files"). In CI, the checkout is clean, so this is not a problem. `--skip-untracked` exists (A20).
- Tested: when the range has only `docs`, `chore` or `refactor` commits, `cog bump --auto --dry-run` exits with code 0 and prints a message on stdout ("No conventional commits for your repository that required a bump..."), not a version. A workflow must check the output against a SemVer pattern.
- Tested: a commit with a scope that is not in `scopes` (`cog.toml`) is ignored for the bump, with an "Errored commit" warning. When it is the only commit, the bump fails.
- 0.x behaviour, from the source: `is_major_bump = self.version.major != 0 && commits.iter().any(Commit::is_major_bump)` (A3, line 258). Thus on 0.x a breaking change gives a minor bump.
- Tested on 0.x: `feat(register)!:` gives 0.1.0 → 0.2.0. `fix(register): x` with a `BREAKING CHANGE:` footer also counts as breaking. `feat` gives a minor bump (0.2.0 → 0.3.0). `fix` gives a patch bump (0.2.1 → 0.2.2).
- Tested on 1.x: `feat(register)!:` gives 1.0.0 → 2.0.0.
- The 0.x rule is not configurable. The settings struct has no key for it (A3; `settings/mod.rs` at 7.0.0; A1 lists none). Per commit type, `bump_minor` and `bump_patch` are configurable (`commit_types`, A1).
- To go to 1.0.0, use `cog bump --major` or `--version 1.0.0` (A20). With `--dry-run` this only prints the version.
- `cog changelog [PATTERN]` takes a git range, and `--at <tag>` gives the changelog of one tag (A20). Tested: `cog changelog v0.1.0..v0.2.0` and `cog changelog --at v0.2.1` print one release section. `cog changelog v1.0.0..HEAD` prints an "Unreleased (abc..def)" section. Thus the workflow can make the release notes before the tag exists.
- Built-in templates: `default`, `remote`, `full_hash`, or a path to a Tera template (A20, A2). `remote` needs `--remote github.com --owner lukemundy --repository cdr-lab`, or the `changelog.*` keys in `cog.toml` (A1). Tested: `remote` gives commit links and a compare link. Observed: for a `--at` range the compare link starts at a commit ID, not at the earlier tag.
- The default template groups commits by type, then by scope, and shows the scope in bold (A4 `simple.tera`). A breaking commit gets a red "BREAKING" HTML `<span>` (tested). The HTML shows on GitHub, but is noise in plain text.
- Template context per commit: `id`, `author`, `signature`, `type`, `date`, `scope`, `summary`, `body`, `type_order`, `breaking_change`, `footers` (A4 `serde.rs` lines 45–58). Note: A2 names the field `footer`, but 7.0.0 serializes `footers`. Each footer is an enum: `{"footer": {"token", "content"}}`, `github_co_authored_by` or `github_closes` (A4 `release.rs` line 128).
- Tested: this Tera condition finds a custom footer: `{% if f.footer is defined and f.footer.token == "Lab-Reset" %}`.
- Monorepo mode: `[packages]` with `path`, `generate_mono_repository_global_tag`, `generate_mono_repository_package_tags`, `monorepo_version_separator` (A1, `settings/mod.rs`). It gives one version and one tag for each package (for example `register-0.3.0`).
- The release must use full history and tags: `actions/checkout` must use `fetch-depth: 0`. Unconfirmed in this research: the research did not fetch the checkout docs; the repository workflow uses the default today (`.github/workflows/ci.yml`, checkout steps have no `fetch-depth`).
- `cog check` on `main` today: "No errored commits". `cog changelog` with no tags gives 73 entries.
- `main` uses squash merges only, with `squash_merge_commit_title: PR_TITLE` and `squash_merge_commit_message: BLANK` (`gh api repos/lukemundy/cdr-lab`). Thus the commit on `main` is the PR title plus trailers only (`git log`: only `Co-authored-by`). A `BREAKING CHANGE:` footer or a custom footer in a branch commit does not reach `main`. Only `!` in the PR title does.

#### Options

- **cog in tag-only mode.** Add `tag_prefix = "v"` to `cog.toml`. The workflow runs `cog bump --auto --dry-run` to get the version, `cog changelog HEAD-range -t <template>` for the notes, then creates the tag through the GitHub release. It never runs `cog bump` without `--dry-run`.
- **cog monorepo mode.** One version for each package. This does not agree with the decision of one version for all artifacts. It also adds package tags and package changelogs that we do not need. Not necessary.

#### Recommendation

Use cog in tag-only mode. Add `tag_prefix = "v"` and `[changelog] remote/owner/repository` to `cog.toml`. Do not use monorepo mode: the scopes already give the component in each changelog line. Add a small custom template without the HTML badge, which also renders the "needs a fresh lab" section (section 1.4). Exact commands:

```sh
version=$(cog bump --auto --dry-run)
[[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "No release: $version" >&2; exit 1; }
prev=$(cog get-version --tag 2>/dev/null || true)
cog changelog "${prev:+$prev..}HEAD" -t .github/release-notes.tera > "$RUNNER_TEMP/notes.md"
```

Unconfirmed: the exact output of `cog get-version --tag` with `tag_prefix`, and how `cog changelog ..HEAD` behaves with no earlier tag. Test both in the implementation ticket.

### 1.2. Alternatives

#### Facts

- **git-cliff** (Apache-2.0 per the GitHub API): `git-cliff --bumped-version` prints the next version. `[bump] features_always_bump_minor`, `breaking_always_bump_major` (set false for "breaking gives minor on 0.x") and `initial_tag = "0.1.0"` (A5). It makes no commits. It is a second Conventional Commits parser next to cog, with its own config (`cliff.toml`).
- **release-please**: keeps a standing release PR. On the merge of that PR, it updates `CHANGELOG.md` and the version files, then tags (A6). It needs a manifest file (`.release-please-manifest.json`) and a config file in the repository (A6). Thus each release is a commit to `main` through a bot PR.
- **semantic-release**: the default plugins are `commit-analyzer`, `release-notes-generator`, `npm` and `github`. It does not commit by default; `@semantic-release/git` adds commits (A7). It is a Node tool with an npm plugin tree. That needs the supply-chain controls of ADR 0015 outside `web/`. Unconfirmed: its 0.x behaviour (A7 does not say).
- **svu**: prints the next version (`svu next`) from the git log. It makes no changelog and no commits (A8). It has a `v0` config option; A8 does not explain it (unconfirmed).

#### Options

| Tool | Commits or bot PRs to `main` | Changelog | Fit with cog in the repo |
|---|---|---|---|
| cog (in use) | None in dry-run and changelog mode | Yes, Tera templates | Already the commit linter and hook |
| git-cliff | None | Yes, Tera templates | Second parser and config; 0.x rule configurable |
| release-please | Yes: release PR merge commit, manifest file | Yes, committed `CHANGELOG.md` | Does not agree with "no commits to main" |
| semantic-release | None by default | Yes | Node plus npm plugins outside the pnpm workspace |
| svu | None | No | Version only; needs a second tool for notes |

#### Recommendation

Use cog. It meets all decisions (tags only, automatic notes, 0.1.0 start, breaking gives minor on 0.x). It is already pinned in `mise.toml` and already validates the commit format. git-cliff is the fallback only if a configurable 0.x rule becomes necessary.

### 1.3. CHANGELOG.md or GitHub Release notes only

#### Facts

- A release must not push commits to `main` (decision). The ruleset "Protect Default Branch" has the rules `creation`, `deletion`, `non_fast_forward`, `pull_request` and `required_status_checks`, with no bypass actors (`gh api .../rulesets`). Thus a workflow cannot push `CHANGELOG.md` to `main`.
- `cog changelog` makes the full changelog again from the tags at any time (A20, tested). The tags are the source of truth, so a committed file is a copy.
- GitHub immutable releases lock the tag and the assets after publication. The title and the notes stay editable. GitHub recommends: create a draft, attach the assets, then publish (A18).
- No tag ruleset exists today (`gh api .../rulesets` lists only the branch ruleset).

#### Options

1. **Release notes only.** The workflow writes the notes into the GitHub Release. No file in the repository. Simple; nothing to keep in sync. A reader offline does not have a changelog.
2. **A release PR before the tag.** A person or a workflow opens a PR `chore(release): vX.Y.Z` that updates `CHANGELOG.md`. After the merge, the workflow tags the merge commit. Two steps for each release, and the PR itself changes `main`, so the tag is not on the commit that CI tested last.
3. **A follow-up PR after the release.** The release workflow opens a PR with the new `CHANGELOG.md`. Auto-merge is off on the repository, so each release needs a manual merge. The file lags the tag.
4. **Changelog as a release asset.** The workflow attaches `CHANGELOG.md` (the full history from `cog changelog`) to each release. No commit. With immutable releases, the asset is locked with the release.

#### Recommendation

Option 1, optionally with option 4. Add one README line that links to the Releases page. A committed file adds a step or lags the tags, and it gives no information that the tags do not give. Also add a tag ruleset that limits the creation of `v*` tags to the release workflow, and turn on immutable releases.

### 1.4. Mark "this release needs a fresh deployment"

#### Facts

- Migrations are goose SQL files in `services/register/internal/migrations/` and `services/dataholder/internal/migrations/`, embedded with `//go:embed *.sql` (`services/register/internal/migrations/migrations.go:6`). A `<service>-migrate` container applies new files (README, Data).
- In the visible history, no migration file was changed or deleted after it was added: `git log --diff-filter=MD -- 'services/*/internal/migrations/*.sql'` gives nothing. A new migration file applies on an existing lab and needs no reset.
- The event codec accepts only schema version 1 (`pkg/eventsourcing/codec.go:12`, `:69`) and refuses unknown fields (`codec.go:75`, `DisallowUnknownFields`). `docs/production-gaps.md:19` says: "a change to an event payload needs a lab reset".
- Event types are registered by name with `eventsourcing.RegisterJSON[...]` in `services/register/internal/app/aggregates.go`, `services/dataholder/internal/app/consent.go` and `services/dataholder/internal/app/registration.go`. The payload Go types are in `services/*/internal/domain/*.go`, in the same files as domain logic (for example `services/dataholder/internal/domain/arrangement.go`). The codec lists the registered names: `EventTypes()` (`pkg/eventsourcing/codec.go:42`).
- Example: commit `1a87d10` (CDR-94) changed event types in `domain/arrangement.go` and `app/consent.go`. A path check on `internal/domain/` would also flag each change of domain logic.
- Commit footers do not reach `main` (section 1.1: squash merge with a blank message). A `Lab-Reset:` footer works in cog templates (tested), but only if the merger types it into the squash message.
- Other state that a reset deletes: the lab CAs (README, Certificates). Out of scope here.

#### Options

1. **Commit footer** (`Lab-Reset: required`). Rendered by the cog template (tested). It depends on a person, and the squash settings drop it. Not reliable.
2. **`!` in the PR title.** It survives the squash. But on 0.x `!` gives only a minor bump, and `!` also marks API breaks that need no reset. It is a mixed signal.
3. **Path check on migrations.** In the release workflow: `git diff --name-only --diff-filter=MDR "$prev"..HEAD -- 'services/*/internal/migrations/*.sql'`. Not empty means a fresh deployment. Exact and cheap. It misses event payload changes.
4. **Golden file of event payloads.** A test in each service encodes one sample of each registered event type (the list from `EventTypes()`) and compares it with a committed file, for example `services/<service>/internal/app/testdata/events.golden.json`. A payload change then must change the golden file. The release check: `git diff --quiet "$prev"..HEAD -- 'services/*/internal/app/testdata/events.golden.json'`. An added event type also changes the file; the check can treat each change as "fresh" (conservative and simple), or compare only the entries that the earlier file had (one `jq` expression).

#### Recommendation

Use options 3 and 4 together, in the release workflow, as one rule: "A release needs a fresh deployment when, since the earlier tag, a committed migration file was changed, renamed or deleted, or an entry of an event golden file that the earlier file had changed or was removed." The workflow writes a fixed "Needs a fresh deployment" section at the top of the notes, with the list of files. The rule comes from the diff, not from a person.

The golden test of each service:

- Gets the registered type names from `EventTypes()`, and has a sample value for each type. The test fails when a registered type has no sample, so each new event type gets one.
- Encodes each sample with the real codec, and compares the result with the committed file. A field rename, a field removal, a type change or a type rename then fails the test. The `-update` flag writes the file again.
- Uses sample JSON, not a schema of field names and types, because sample JSON is simpler.

Limit: the test finds a change of the shape of a payload, not of its meaning, for example a field that keeps its type but changes its units. The author of such a change writes it in the release notes.

### 1.5. Promote a tested image without a rebuild

#### Facts

- Today CI does `docker save` in the test job and `docker load` plus `docker push --all-tags` in the publish job, with tags = full and short commit (`.github/workflows/ci.yml:411–422`, `:442–464`, `:602–620`). The images are single-platform. CI has no signing, no SBOM and no provenance attestation (`grep cosign|attest|provenance|platform` finds nothing in `ci.yml`, `compose.yaml`, `mise.toml`).
- `crane tag`: "Tag remote image without downloading it". It skips the layer checks "because we know the manifest already exists" (A10). A10 does not state the digest; because only a tag is added to the existing manifest, the digest stays the same (inference; unconfirmed in A10).
- `oras tag`: "Tag a manifest in a registry", by tag or by digest (A11). A11 does not state the digest; same inference (unconfirmed).
- `docker buildx imagetools create`: with one source that is an index, it makes "a carbon copy". With one source that is not an index, the output is a manifest list by default; `--prefer-index=false` keeps the source format (A9). Thus for a single-platform image without `--prefer-index=false`, the digest changes. That is a trap.
- buildx attestations are manifests inside the image index, linked by the `vnd.docker.reference.digest` annotation (A12). A tag on the same index digest keeps them.
- cosign signatures use the OCI 1.1 referrer model (A13), and refer to the image by digest. A new tag on the same digest keeps them. A13 does not state this binding in words (partly unconfirmed).
- Permissions: the publish jobs use `packages: write` (`ci.yml:449`, `:609`). The `GITHUB_TOKEN` is limited to the repository (A17). A release job needs `packages: write` for ghcr and `contents: write` to create the tag and the release (contents permission: unconfirmed in A17, standard GitHub behaviour).
- Events that the `GITHUB_TOKEN` causes do not start new workflow runs, except `workflow_dispatch` and `repository_dispatch` (A17). Thus a tag that the release workflow creates does not start a `push: tags` workflow. The release workflow must do all steps itself.

#### Options

| Tool | Digest kept | Attestations and signatures kept | Notes |
|---|---|---|---|
| `crane tag` | Yes (inference) | Yes, same digest | One static binary; no Docker daemon |
| `oras tag` | Yes (inference) | Yes, same digest | One static binary |
| `docker buildx imagetools create` | Yes for an index; no for a single manifest unless `--prefer-index=false` | Yes for an index | Already on GitHub runners |

#### Recommendation

Use `crane tag ghcr.io/lukemundy/cdr-lab/<service>@sha256:<digest> vX.Y.Z` (also `X.Y.Z` if wanted). Resolve the digest from the full-commit tag first (`crane digest ...:<full commit>`), and check after the tag that the digest of `vX.Y.Z` is the same. Pin crane by version in `mise.toml`. A release job needs `contents: write` and `packages: write`. The release must refuse a commit that has no tested image of each service: today an image exists only for commits where the image inputs changed (path filters in `ci.yml`).

For each service, release the image of the last commit that had an image build: walk the first-parent history of `main` back from the release commit, and take the first commit that has a full-commit tag (`crane digest`). That image is equal to a build of the release commit, because the path filter skipped the build only when no image input changed. Two guards stop a release with an old image when a publish job failed or was cancelled:

1. The release refuses a commit whose CI run on `main` did not pass.
2. For each service, the image commit must be `git log -1 --format=%H <release> -- <image inputs>`. The inputs are the paths of the image filter (`ci.yml:94–96`). Keep these paths in one file that the CI filter and the release script both read, so that the two lists cannot differ.

The release notes give the image commit of each service, because the services can have different commits and `service.version` is the commit (section 1.8).

### 1.6. Helm charts

#### Facts

- `helm package` has `--version` and `--app-version`, and `-u, --dependency-update` (A19). CI already uses `--version "0.0.0-g$commit" --app-version "$commit"` (`ci.yml:513`). These flags change only the packaged chart, not the files in the repository.
- `Chart.yaml` of each chart has `version: 0.0.0` and a comment that CI sets the version (`charts/register/Chart.yaml`, `charts/dataholder/Chart.yaml`).
- The charts do not use `.Chart.AppVersion`. `image.tag` is required and has no default (`charts/register/values.yaml:7`, `charts/register/templates/common.yaml:38`). Thus the `appVersion` of a release chart does not select the image today.
- Tested (Helm 4.3.2): an umbrella chart with a `file://../sub` dependency at `version: 0.1.0` fails in `helm dependency update` when `sub/Chart.yaml` has `version: 0.0.0`. After the workspace change of `sub/Chart.yaml` to `0.1.0`, the update succeeds. The packaged umbrella contains the subchart as a directory (`umb/charts/sub/`), with `version` and `appVersion` from the subchart `Chart.yaml` in the workspace. `helm package sub --version` does not change `sub/Chart.yaml`.
- Thus `helm package --version` on the umbrella does not change the versions of its dependencies. The workflow must change the `Chart.yaml` files in the CI workspace (no commit) before `helm dependency update`.
- The charts depend on `common` 5.3.0 from `oci://ghcr.io/bjw-s-labs/helm`, locked in `Chart.lock` (`charts/register/Chart.lock`). `mise run chart:deps` runs `helm dependency build --skip-refresh` (`mise.toml:64–69`).

#### Options

1. **file:// dependencies.** In the workspace, set `version` and `appVersion` in `charts/{register,dataholder}/Chart.yaml` and the dependency versions in the umbrella `Chart.yaml` (for example with `yq`), then `helm dependency update charts/<umbrella>` and `helm package`. The umbrella contains copies of the subcharts. One step, no registry round trip. The `Chart.lock` of the umbrella is made in CI and not committed.
2. **OCI dependencies.** The umbrella `Chart.yaml` refers to `oci://ghcr.io/lukemundy/cdr-lab/charts`. The workflow first pushes the two subcharts at `X.Y.Z`, then sets the dependency versions in the workspace, runs `helm dependency update` and packages the umbrella. The umbrella then contains the exact published subchart packages. It needs order and registry access, and a lint of the umbrella in pull requests needs a published version or a local override.

#### Recommendation

Option 1. Use `version: 0.0.0` in the committed umbrella dependencies, the same as the subcharts. The release (and the `main` publish) job sets one version in all three `Chart.yaml` files of the workspace, then packages with `--version X.Y.Z --app-version X.Y.Z`. Consider a default of `image.tag` to `.Chart.AppVersion` for the release charts, so that a release chart deploys the release images; this changes the chart rule "image.tag is required" and needs its own decision.

### 1.7. Go CLI release binaries

#### Facts

- `go install pkg@version` rules: the module `go.mod` must not contain `replace` or `exclude` directives; the module must not require a higher version of itself (A14). The repository `go.mod` has no `replace` and no `exclude`; it has `ignore ./web` (`go.mod:6`). Unconfirmed: whether the `ignore` directive has an effect on `go install @version`. A14 lists only `replace` and `exclude`.
- A module at the repository root uses plain tags `vX.Y.Z`. Only a module in a subdirectory needs a prefix (A14). The repository is one module, `github.com/lukemundy/cdr-lab`, at the root (`go.mod:1`). Thus yes: `go install github.com/lukemundy/cdr-lab/cmd/cdrlab@v0.1.0` works with the tag `v0.1.0` (A14). The package path of the CLI is still open (`cmd/cdrlab` or `tools/...`).
- `go install` downloads the full module zip. The Go module zip has a size limit (unconfirmed in this research: the research did not check the limit or the size of the repository zip; the `web/` directory is in the zip).
- GoReleaser OSS is MIT (GitHub API). Pro-only features include monorepo support, the split prepare/publish workflow, nightlies and custom template variables (A15). One root module with one tag needs none of these.
- GoReleaser takes release notes from a file: `--release-notes=FILE` (A16). `release.mode` is `keep-existing` (default), `append`, `prepend` or `replace`; `draft: true` and `use_existing_draft` exist (A16). It expects a tagged commit; you can create the tag locally and not push it (A16).
- GoReleaser makes archives, `checksums.txt` and the GitHub Release (common knowledge of the tool; not fetched in this research beyond A16).

#### Options

1. **GoReleaser OSS.** One `.goreleaser.yaml` with `builds` (`goos: linux, darwin, windows`, `goarch: amd64, arm64`, `CGO_ENABLED=0`, `-trimpath`), archives, checksums, and `--release-notes` from cog. It creates the draft release. Use it only for the CLI, not for the images. A new tool to pin and a new config file.
2. **Plain `go build` matrix.** A shell loop over six `GOOS/GOARCH` pairs, `sha256sum` into `checksums.txt`, and `gh release create vX.Y.Z --draft --target <sha> --notes-file notes.md dist/*`. No new tool. More script to keep: archive names, Windows `.exe`, and later extras (SBOM, signing) by hand.

#### Recommendation

A plain `go build` loop in a mise task, with `gh release create`. Six targets and one checksum file do not need GoReleaser, and the repository already prefers mise tasks and pinned tools. Move to GoReleaser OSS only when archives, SBOMs or signing are added.

### 1.8. Stamp the SemVer version into binaries and images

#### Facts

- Today `pkg/platform/buildinfo` has four `-X` variables: `commit`, `commitShort`, `buildTime`, `id` (`pkg/platform/buildinfo/buildinfo.go`). The Dockerfile and `mise run build` set them (`services/register/Dockerfile`, `go build ... -ldflags`; `mise.toml:23–26`, `:195`).
- `service.version` is `build.CommitShort` (`pkg/platform/telemetry/resource.go:30`), and `vcs.ref.head.revision` is `build.Commit` (`resource.go:32`). The attributes come last, so `OTEL_RESOURCE_ATTRIBUTES` cannot change them (`resource.go:26`).
- Tested (Go 1.27.4): `go build` in a git checkout at a tag sets `debug.ReadBuildInfo().Main.Version` to `v0.1.0`, with `vcs.revision`. After one more commit it gives a pseudo-version `v0.1.1-0.<time>-<commit>`, with `+dirty` when the tree has changes (A14). `go run` gives `(devel)`.
- The Dockerfile copies no `.git` directory (`services/register/Dockerfile`: `COPY go.mod go.sum`, `pkg/`, `services/register/`). Thus the VCS stamp is not available in the image build, and `-X` stays necessary there.
- A promoted image (section 1.5) is built before the tag exists. Its binary cannot contain `vX.Y.Z`.

#### Options

1. **Images keep the commit; only the CLI gets the SemVer.** `service.version` stays the short commit. The release notes and the image tags map `vX.Y.Z` to the commit. The CLI is built at release time, so it gets `-X buildinfo.version=vX.Y.Z` plus the commit. Simple; agrees with "promote without rebuild". Telemetry does not show the release name.
2. **Add a `version` variable, set to `git describe --tags` at each build.** On `main` between releases this gives `v0.1.0-3-gabc1234`. The image of the release commit is built before the tag, so it says `v0.0.0-...` or the earlier tag, not the new version. Misleading for the release image.
3. **Rebuild the images at release time with `version=vX.Y.Z`.** The released image is then not the tested image. This goes against the promotion decision.

#### Recommendation

Option 1. Add one `version` variable to `pkg/platform/buildinfo` (default `"dev"`, set with `-X` from a `BUILD_VERSION` variable). Use it in the CLI `--version` output with the commit. Keep `service.version` as the short commit for the images, because the commit identifies the tested build and the release tag points to it. For `go install @vX.Y.Z` (no ldflags), read `debug.ReadBuildInfo().Main.Version` as the fallback when `version` is `"dev"`; this gives `vX.Y.Z` without any flags (tested behaviour of Go, A14). Do not set a CLI default to a lab value.

## 2. Supply chain: multi-arch, signatures, attestations and SBOM

Date: 2026-11-22. Repository state: `main` at `1a45e0b`.

Versions that the research checked:

- Local tools: Docker Buildx 0.39.0, gh 2.102.0. The research did not have `crane` or `cosign`. The research inspected the base images with `docker buildx imagetools inspect`.
- GitHub `ubuntu-24.04` runner image 20261116.376.1: Docker Client and Server 28.0.4, Buildx 0.39.0, Compose 2.38.2 (B3).
- cosign v3.1.4, released 2026-10-20 (B6). `sigstore/cosign-installer@v4.0.0` in the Sigstore CI quickstart (B8).
- Helm docs show version 4.3.1 (B12).
- `actions/attest` v4. `actions/attest-build-provenance` v4 is a wrapper of `actions/attest` (B4, B5).

### Sources of part 2

| ID | Source | URL |
|---|---|---|
| B1 | GitHub Docs: GitHub-hosted runners reference | https://docs.github.com/en/actions/reference/runners/github-hosted-runners |
| B2 | Docker Docs: multi-platform builds | https://docs.docker.com/build/building/multi-platform/ |
| B3 | actions/runner-images: Ubuntu 24.04 readme | https://github.com/actions/runner-images/blob/main/images/ubuntu/Ubuntu2404-Readme.md |
| B4 | actions/attest-build-provenance repo | https://github.com/actions/attest-build-provenance |
| B5 | actions/attest repo | https://github.com/actions/attest |
| B6 | sigstore/cosign releases | https://github.com/sigstore/cosign/releases |
| B7 | cosign `verify` reference | https://raw.githubusercontent.com/sigstore/cosign/main/doc/cosign_verify.md |
| B8 | Sigstore CI quickstart | https://docs.sigstore.dev/quickstart/quickstart-ci/ |
| B9 | cosign signature specification | https://raw.githubusercontent.com/sigstore/cosign/main/specs/SIGNATURE_SPEC.md |
| B10 | Sigstore: signing containers | https://docs.sigstore.dev/cosign/signing/signing_with_containers/ |
| B11 | Sigstore: verify (blobs and bundles) | https://docs.sigstore.dev/cosign/verifying/verify/ |
| B12 | Helm: provenance and integrity | https://helm.sh/docs/topics/provenance/ |
| B13 | GitHub Docs: artifact attestations (concepts) | https://docs.github.com/en/actions/concepts/security/artifact-attestations |
| B14 | GitHub Docs: use artifact attestations | https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations |
| B15 | GitHub CLI manual: `gh attestation verify` | https://cli.github.com/manual/gh_attestation_verify |
| B16 | Docker Docs: SBOM attestations | https://docs.docker.com/build/metadata/attestations/sbom/ |
| B17 | Anchore: Syft package capabilities | https://oss.anchore.com/docs/capabilities/all-packages/ |
| B18 | anchore/sbom-action repo | https://github.com/anchore/sbom-action |
| B19 | Flux: HelmChart source (verify) | https://fluxcd.io/flux/components/source/helmcharts/ |
| B20 | Flux: OCIRepository source (verify) | https://fluxcd.io/flux/components/source/ocirepositories/ |
| B21 | Argo CD: source integrity | https://argo-cd.readthedocs.io/en/latest/user-guide/source-integrity/ |
| B22 | Docker Docs: `docker image save` | https://docs.docker.com/reference/cli/docker/image/save/ |
| B23 | Docker Docs: multi-platform builds in GitHub Actions | https://docs.docker.com/build/ci/github-actions/multi-platform/ |

### 2.1. Multi-arch images (linux/amd64 and linux/arm64)

#### Facts

- The `images` job builds all lab images with `mise run build:images` (`docker-compose build`), then `docker save` puts them in one tar artifact (`.github/workflows/ci.yml:201-230`, `mise.toml:183-188`). The lab jobs `docker load` the tar (`ci.yml:268-276`).
- The `register-image` and `dataholder-image` jobs tag the image of the running, tested container with the full and short commit, `docker save` it, and upload it (`ci.yml:411-429`, `ci.yml:572-590`). The publish jobs `docker load` and `docker push --all-tags` with `packages: write` only (`ci.yml:442-465`, `ci.yml:602-625`).
- Recent `images` job time on `main`: 99 s to 136 s (5 successful runs, from `gh run view`). The publish jobs take 10 s to 27 s.
- Each Dockerfile has three stages: `web` (Node), `build` (Go) and the final distroless stage (`services/register/Dockerfile:3,21,47`, `services/dataholder/Dockerfile:3,21,49`). No stage uses `--platform`. The `build` stage sets `CGO_ENABLED=0` (`services/register/Dockerfile:23`).
- All three pinned digests are multi-arch OCI image indexes (`application/vnd.oci.image.index.v1+json`), checked with `docker buildx imagetools inspect` on 2026-11-22:
  - `node:26.13.0-trixie-slim@sha256:e4db4c91…`: linux/amd64, linux/arm64/v8 and more, plus attestation manifests.
  - `golang:1.27.4-trixie@sha256:b54b2639…`: linux/386, amd64, arm64/v8, arm/v7, ppc64le, s390x.
  - `distroless/static-debian13:nonroot@sha256:25ac8f0c…`: linux/amd64, arm64/v8, arm/v7, s390x, ppc64le, riscv64.
  - Thus no digest changes are necessary for arm64.
- The Docker strategies are QEMU emulation, native nodes for each architecture, and cross-compilation with `FROM --platform=$BUILDPLATFORM` and `TARGETOS`/`TARGETARCH` passed to `GOOS`/`GOARCH` (B2). B2: "Emulation with QEMU can be much slower than native builds", mostly for compilation and compression.
- `ubuntu-24.04-arm` (also `ubuntu-22.04-arm`, `ubuntu-26.04-arm`) is a standard runner for public repositories: 4 CPU, 16 GB RAM, arm64. "Use of the standard GitHub-hosted runners is free and unlimited on public repositories" (B1).
- A multi-platform image needs an image store that supports manifest lists. Docker Engine 29.0+ uses the containerd image store by default. The `docker-container` driver does not load into the Engine store, so you push its output directly (B2). The runner has Docker Engine 28.0.4 (B3). Thus the runner daemon uses the classic store by default (inference from B2 and B3, unconfirmed on a runner). It cannot hold a two-platform index.
- `docker save --platform` (API 1.48+) saves only the named platform. It fails when that platform is not in the local store (B22).
- Docker documents a reusable workflow `docker/github-builder/.github/workflows/build.yml@v1` that distributes platforms over runners and creates the manifest. Its other examples use `docker/setup-qemu-action@v4`, `docker/setup-buildx-action@v4` and `docker/build-push-action@v7` (B23).
- Arch dependence of the Go notices step: `tools/notices` runs `go list -deps -json` in a subprocess (`tools/notices/main.go:41-42`). `go list` applies `GOOS`/`GOARCH` build constraints. The research compared the module sets of `./services/register/cmd/register` and `./services/dataholder/cmd/dataholder` for `GOARCH=amd64` and `GOARCH=arm64`. They are the same (74 modules each, on 2026-11-22). Thus the Go notices are the same for the two architectures today, but this can change with a new dependency.
- Pitfall: the Dockerfile runs the tool with `go run ./tools/notices` (`services/register/Dockerfile:34-35`). If the stage sets `ENV GOARCH=$TARGETARCH`, `go run` compiles the tool for the target and cannot run it on the build host (standard Go behaviour, not checked in a build). Set `GOARCH` only on the `go build` command, or build the tool for the host and run it with `GOARCH=$TARGETARCH` in its environment.
- Arch dependence of the web stage: the lock file has platform-specific native packages for build tools, for example `lightningcss-linux-arm64-gnu` (`web/pnpm-lock.yaml:1680`). These are build tools. The output is JavaScript, CSS, fonts and notice text. Thus the web output does not depend on the target (inference, unconfirmed by a build). With `FROM --platform=$BUILDPLATFORM`, the stage runs one time natively, and the two images share it.
- BuildKit SBOM and provenance attestations go into the image index on push (B16). With the classic store and `docker save`/`docker load`, they are not kept (inference, unconfirmed).

#### Options

| Option | How | CI time (estimate, unconfirmed) | Notes |
|---|---|---|---|
| A. Cross-compile on one amd64 runner | `web` and `build` stages `FROM --platform=$BUILDPLATFORM`; `go build` with `GOOS=$TARGETOS GOARCH=$TARGETARCH`; final stage per target | Web stage one time; Go compile two times. Approximately +20–40 s on the `images` job. | No emulation, because the distroless final stage has no `RUN`. One runner. |
| B. QEMU on one amd64 runner | `docker/setup-qemu-action`, no Dockerfile change | The arm64 Node install and build and the Go build run emulated: likely many minutes | Slowest. Also the known "Bus error" in the web build on hosted runners (`ci.yml:233`) can get worse (unconfirmed). |
| C. Native arm64 runner and manifest merge | Matrix of `ubuntu-latest` and `ubuntu-24.04-arm`, push by digest, merge with `docker buildx imagetools create` | Same wall time as today in parallel; one more merge job (approximately 10–20 s) | Free for public repos (B1). More jobs and more YAML. Gives a native arm64 test only if the lab jobs also run on arm64. |

Change to the artifact flow, for all options:

- The lab jobs test amd64 only. They must test the same amd64 image that goes into the index.
- Approach 1 (simplest with Engine 28): build amd64 with `--load` as today for the lab and tests. Build arm64 separately to an OCI tar (`--output type=oci`). In the publish job, push the tested amd64 image and the arm64 image by digest, then make the index with `docker buildx imagetools create`. Risk: the arm64 image is not tested, and it comes from a different build invocation (the build arguments `BUILD_TIME` and `BUILD_ID` must be the same).
- Approach 2: enable the containerd image store on the runners (for example with `docker/setup-docker-action` and the `containerd-snapshotter` feature, not checked), build one index with `docker buildx build --platform linux/amd64,linux/arm64 --load`, and `docker save` it. The lab runs the amd64 variant from the same index. The publish job pushes the whole index, and the BuildKit attestations can stay. Unconfirmed: that compose, `docker save` and `docker load` keep the index intact on Engine 28 with the containerd store.

#### Recommendation

Option A (cross-compile on the current amd64 runner), with Approach 1 for the artifacts. It adds the least CI time (estimate: under 1 minute on the `images` job) and needs no new runner type. Put `--platform=$BUILDPLATFORM` on the `web` and `build` stages, and keep the `web` stage the same in each Dockerfile so that BuildKit shares it (`services/register/Dockerfile:1`). Add a smoke start of the arm64 image on `ubuntu-24.04-arm` only if an arm64 consumer needs it. Measure the time cost with approximately 5 runs before you claim it.

### 2.2. Signing with cosign keyless

#### Facts

- Keyless signing gets a short-lived Fulcio certificate for an OIDC identity and logs the signing event in Rekor (B10, B11 overview).
- GitHub Actions signing needs `id-token: write`. The quickstart uses `sigstore/cosign-installer@v4.0.0` and `cosign sign --yes ${images}` (B8).
- Verify command (B8, B7):

  ```sh
  cosign verify ghcr.io/lukemundy/cdr-lab/register@sha256:<digest> \
    --certificate-identity=https://github.com/lukemundy/cdr-lab/.github/workflows/<workflow>.yml@<ref> \
    --certificate-oidc-issuer=https://token.actions.githubusercontent.com
  ```

  `<ref>` is `refs/heads/main` when `ci.yml` on `main` signs. It is `refs/tags/vX.Y.Z` when a release workflow on a tag push signs. `--certificate-identity-regexp` is the regular-expression alternative. `--certificate-github-workflow-repository`, `-ref`, `-trigger` and `-sha` check the claims of the GitHub OIDC token (B7).
- Blobs: `cosign verify-blob <file> --bundle <file>.sigstore.json --certificate-identity=… --certificate-oidc-issuer=…` (B11). Signing produces a bundle by default in cosign v3 (B6).
- cosign v3: bundle format is the default for signing (B6). B10 says "Signatures use the OCI 1.1 referrer specification". The older specification stores the signature at the tag `sha256-<digest>.sig` (B9). Which mode cosign v3 uses by default against ghcr.io: unconfirmed.
- Signatures are keyed by digest, not tag. A tag that moves to a new digest has no signature; the signature of the old digest stays (B9). Thus a retag of the same digest keeps the signature. B9 says the design "does not mutate the signed object".
- Helm charts on an OCI registry are OCI artifacts with a digest. cosign signs an OCI reference by digest, so the same command applies to `ghcr.io/lukemundy/cdr-lab/charts/<name>@sha256:…`. The research found no official cosign page that names Helm charts (unconfirmed, but the mechanism is generic).

#### Options

- Sign at the `main` build (in the publish jobs, after the push) and only retag at release. The signature stays valid (B9). The certificate identity is `…/ci.yml@refs/heads/main`. Consumers verify that identity for each release. Weakness: the identity does not show that the digest is a release. The tag shows it, and tags are mutable.
- Sign again in the release workflow. A digest can have more than one signature (B10). The second signature has the identity `…/release.yml@refs/tags/vX.Y.Z`, so a consumer can require "signed by the release workflow on a tag". The release job needs `id-token: write`.
- Charts: today, `chart-publish` makes a new chart version `0.0.0-g<commit>` for each commit (`ci.yml:501-514`). A SemVer chart version (for example `1.2.0`) changes `Chart.yaml` content, thus the release chart has a new digest. A retag cannot promote a chart. The release workflow must package and push it, and so it must sign it.
- CLI binaries: built at release, so the release workflow signs them (or their checksums file) with `cosign sign-blob --bundle`.

#### Recommendation

Sign in the release workflow, for all three artifact types: images (sign the promoted digest again), charts (new digest at release) and the checksums file of the CLI binaries. Signing only at the `main` build is valid for images, but it does not work for charts and binaries. Then the release workflow must sign anyway, and one identity (`release.yml@refs/tags/v*`) is simpler for consumers. Whether to also sign each `main` image is a choice: it is not necessary if the release step attests and signs. See section 2.3, because attestations can replace cosign signatures.

### 2.3. GitHub artifact attestations

#### Facts

- `actions/attest-build-provenance` v4 is a wrapper of `actions/attest`. "New implementations should use `actions/attest` instead" (B4).
- Availability: public repositories on all current plans. Free, Pro and Team plans get them for public repositories only (B4). Public repositories use the Sigstore public-good instance, and the bundle goes to the public transparency log (B13). Cost: free on public repos (B1 for runners; no fee is stated for attestations, unconfirmed).
- "Artifact attestations by itself provides SLSA v1.0 Build Level 2." A reusable workflow can give isolation for Build Level 3 (B13).
- Inputs: `subject-name` (image name with no tag), `subject-digest` (`sha256:…`), `push-to-registry` (default false; needs one fully-qualified OCI image reference with a digest), `sbom-path` (SPDX or CycloneDX JSON, maximum 16 MB) (B5, B14).
- Permissions: `id-token: write`, `attestations: write`, `contents: read`, and `packages: write` for images. `artifact-metadata: write` makes the storage record on the linked artifacts page (B5, B14).
- Storage: the GitHub attestations API, linked to the repository. With `push-to-registry`, also the image registry (B5). The exact registry format (referrers or tag) is not stated (unconfirmed).
- Verify (B14, B15):

  ```sh
  gh attestation verify oci://ghcr.io/lukemundy/cdr-lab/register:v1.2.3 \
    --repo lukemundy/cdr-lab \
    --signer-workflow lukemundy/cdr-lab/.github/workflows/release.yml
  # SBOM:
  gh attestation verify oci://ghcr.io/lukemundy/cdr-lab/register:v1.2.3 \
    --repo lukemundy/cdr-lab --predicate-type https://spdx.dev/Document/v2.3
  # Binary:
  gh attestation verify ./cdrlab_linux_amd64 --repo lukemundy/cdr-lab
  ```

  The default predicate is SLSA provenance v1. `--source-ref` checks the git ref. `--bundle-from-oci` reads the bundle from the registry, not the GitHub API. `--deny-self-hosted-runners` rejects self-hosted builds (B15). The consumer must log in to the registry first (B14).
- An attestation is a signed in-toto statement with a Sigstore certificate (B13). Thus it is a signature on the digest, and it also gives provenance.
- How `gh` finds attestations (by digest through the API) is not stated on B15 (unconfirmed). The subject is the digest (B5), so a retag keeps it (inference).
- `cosign verify-attestation` against attestations from `actions/attest`: no official source found (unconfirmed).

#### Options

- Attestations only (provenance + SBOM). Consumers use `gh` (needs a GitHub login for the API, or `--bundle-from-oci`).
- Attestations plus a separate `cosign sign`. Gives the plain `cosign verify` path, which Flux (`provider: cosign`) and admission controllers use (B19, B20). An `actions/attest` attestation is not a plain cosign signature, so `cosign verify` does not find it (inference; unconfirmed).

#### Recommendation

Minimum set:

1. `actions/attest` build provenance for each image digest (with `push-to-registry: true`), each chart digest, and the CLI binaries. SLSA Build Level 2.
2. `actions/attest` with `sbom-path` for each image.
3. `cosign sign` of each image and chart digest, in the same release job, because Flux and other cluster tools verify cosign signatures, not GitHub attestations. For the CLI binaries, the attestation is sufficient. Add `cosign sign-blob` only if a consumer has no `gh`.

To get SLSA Build Level 3, move the build and attest steps into a reusable workflow (B13). That is a bigger change. Record it as a choice, not a requirement.

### 2.4. SBOM generation

#### Facts

- BuildKit SBOM: uses the BuildKit Syft scanner (Anchore Syft). Format: SPDX JSON in an in-toto predicate. By default it scans only the final stage. `BUILDKIT_SBOM_SCAN_STAGE` and `BUILDKIT_SBOM_SCAN_CONTEXT` add stages or the context; they must be `ARG`s in the Dockerfile (B16).
- The B16 example SBOM lists Go modules "from go module information: bin/server". Thus the scan finds Go modules from the build information of a Go binary (B16, B17 `go-module-binary-cataloger`). The binary is built with `-trimpath -s -w` (`services/register/Dockerfile:40`); `-s -w` strips the symbol table, not the build information (standard Go behaviour, unconfirmed with a scan).
- The npm packages are only in the bundled JavaScript inside the Go binary, and in the notice text in `/licenses/web/` (`services/register/Dockerfile:31,55`). The final image has no `package.json` and no lock file. Thus a scan of the final image does not list the npm packages (inference, unconfirmed with a scan). A scan of the `web` stage lists all workspace packages, also development dependencies and build tools that are not in the image.
- Syft catalogs `pnpm-lock.yaml` (`javascript-lock-cataloger`) (B17).
- `anchore/sbom-action@v0`: inputs `image`, `path`, `file`, `format` (`spdx`, `spdx-json`, `cyclonedx`, `cyclonedx-json`; default `spdx-json`), `output-file`, `upload-artifact`, `upload-release-assets` (B18).
- `actions/attest` accepts SPDX or CycloneDX JSON (B5).
- The BuildKit attestations are not kept by the current `docker save`/`docker load` flow (section 2.1).

#### Options

- BuildKit `--sbom=true`: no extra tool, but it does not survive the current artifact flow, and it misses the npm packages.
- Syft (`anchore/sbom-action`) on the pushed image digest, in the release job: gives the Go modules. Add the npm packages with a second scan of `web/pnpm-lock.yaml` filtered to the production packages of the apps that each image embeds (filtering: unconfirmed).
- Format: SPDX JSON is the default of both tools and has a fixed `--predicate-type` for `gh` (`https://spdx.dev/Document/v2.3`, B14). CycloneDX is also accepted.

#### Recommendation

Syft through `anchore/sbom-action` on the image digest, in SPDX JSON, attested with `actions/attest` `sbom-path`. Accept that version 1 lists the Go modules only, and record the npm gap, or merge a lock-file scan. Do not use BuildKit `--sbom` while the images go through `docker save`/`docker load`.

### 2.5. Helm chart verification

#### Facts

- `helm install --verify` checks a PGP provenance file (`.prov`) with a keyring. If verification fails, Helm stops the install before it renders the chart (B12).
- For OCI registries, the Helm docs refer to the `helm-sigstore` plugin to publish provenance to Rekor (B12). The page says that it "has not yet been updated for Helm 4" (B12, version selector 4.3.1).
- The research found no official source that says Helm 3 or Helm 4 verifies cosign signatures on OCI charts natively (unconfirmed; probably not). Helm 4 signs and verifies plugins, not charts (search result, not a primary source).
- Consumer verify before install, with cosign (mechanism of section 2.2):

  ```sh
  cosign verify ghcr.io/lukemundy/cdr-lab/charts/register@sha256:<digest> \
    --certificate-identity=https://github.com/lukemundy/cdr-lab/.github/workflows/release.yml@refs/tags/v1.2.3 \
    --certificate-oidc-issuer=https://token.actions.githubusercontent.com
  helm install register oci://ghcr.io/lukemundy/cdr-lab/charts/register@sha256:<digest>
  ```

  Install by digest so that the verified artifact is the installed artifact. Or `gh attestation verify oci://ghcr.io/lukemundy/cdr-lab/charts/register:1.2.3 --repo lukemundy/cdr-lab`.
- Flux: `HelmChart` `.spec.verify` with `provider: cosign` or `notation`, only for charts from an OCI registry. Keyless with `matchOIDCIdentity` (`issuer` and `subject` regexps); keyless is experimental, and checks against public Fulcio and Rekor (B19). `OCIRepository` has the same `verify` (B20).
- Argo CD: source integrity supports only GnuPG verification of Git commits. No cosign or Notation for Helm or OCI sources (B21, latest docs).

#### Recommendation

Sign the chart digest with cosign keyless (section 2.2). Give consumers the `cosign verify` command and the Flux `matchOIDCIdentity` values in the chart README. Do not use PGP `.prov` files: they need a long-lived key, and keyless cosign does not.

### 2.6. GitHub Actions permissions

#### Facts

| Permission | Why | Source |
|---|---|---|
| `contents: read` | Checkout | B14 |
| `contents: write` | Create the GitHub Release and upload the CLI binaries (standard GitHub API rule, not checked on a page) | unconfirmed |
| `packages: write` | Push and retag images and charts on ghcr.io; push attestations and cosign signatures to the registry | B14, `ci.yml:448-449` |
| `id-token: write` | OIDC token for the Fulcio certificate (cosign and `actions/attest`) | B5, B8 |
| `attestations: write` | Store attestations in the GitHub attestations API | B5, B14 |
| `artifact-metadata: write` | Storage record on the linked artifacts page (optional) | B5, B14 |

- The workflow default is `contents: read` (`ci.yml:17-18`), and only the publish jobs get `packages: write` (`ci.yml:448`, `ci.yml:489-491`).

#### Recommendation

Give `id-token: write`, `attestations: write` and `packages: write` only to the release job that signs and attests, and `contents: write` only to the step or job that creates the GitHub Release. Keep the workflow default `contents: read`. Do not give `id-token: write` to a job that runs on a pull request.

## 3. CDR certificates and public TLS in a cluster

Date: 2026-11-22. Versions checked: cert-manager docs "latest" (newest release in the docs menu: v1.21; v1.19 and v1.20 release notes also read). cert-manager source: `master` branch on 2026-11-22. issuer-lib: `main` branch, no tagged release shown. Argo CD docs: `stable`. Reloader: `master` README, version not shown.

### Sources of part 3

| ID | Source | URL |
|---|---|---|
| C1 | cert-manager API reference (Certificate, CAIssuer, X509Subject, CertificatePrivateKey, KeyUsage) | https://cert-manager.io/docs/reference/api-docs/ |
| C2 | cert-manager docs: CA issuer | https://cert-manager.io/docs/configuration/ca/ |
| C3 | cert-manager source: certificate template from a CSR | https://github.com/cert-manager/cert-manager/blob/master/pkg/util/pki/certificatetemplate.go |
| C4 | cert-manager docs: external issuers | https://cert-manager.io/docs/contributing/external-issuers/ |
| C5 | cert-manager issuer-lib | https://github.com/cert-manager/issuer-lib |
| C6 | cert-manager docs: CertificateRequest (approval) | https://cert-manager.io/docs/usage/certificaterequest/ |
| C7 | cert-manager release notes 1.19 / 1.20 | https://cert-manager.io/docs/releases/release-notes/release-notes-1.19 , https://cert-manager.io/docs/releases/release-notes/release-notes-1.20 |
| C8 | Helm docs: chart tips (`helm.sh/resource-policy`) | https://helm.sh/docs/howto/charts_tips_and_tricks/ |
| C9 | Argo CD docs: Helm (hooks, resource policy, random data) | https://argo-cd.readthedocs.io/en/stable/user-guide/helm/ |
| C10 | Argo CD docs: resource tracking | https://argo-cd.readthedocs.io/en/stable/user-guide/resource_tracking/ |
| C11 | Argo CD issue 5202: Helm `lookup` not supported | https://github.com/argoproj/argo-cd/issues/5202 |
| C12 | stakater Reloader README | https://github.com/stakater/Reloader |
| C13 | Kubernetes docs: Secrets as files (mounted Secrets are updated, not with `subPath`) | https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod |

Repository files are cited as `path:line`.

### 3.0. Corrections to the research questions

The research question said that the CDR profile has "a certificate policies extension with specific OIDs". It does not. `pkg/ca/DESIGN.md:107` (CA-0008) says that the mock CDR CA certificates have **no** ACCC certificate policy OID, and no rule in the certificate profile (`pkg/ca/DESIGN.md:134-155`) requires a certificate policies extension. Thus cert-manager's lack of a policies field is not a gap.

Also: today the lab issues CDR certificates **offline**. `mise.toml:198-266` runs `pki issue` with the intermediate key on the host. It does not call the Register. The Register operator API issues only after the lab starts (`services/register/design/certificate-management.md:53-55`).

### 3.1. Can the cert-manager CA issuer make a certificate with the CDR profile?

#### Facts

The CDR profile (`pkg/ca/DESIGN.md:136-155`), with the cert-manager feature for each rule:

| Rule | cert-manager CA issuer | Result |
|---|---|---|
| CA-0025: subject is the CSR subject | The template copies `Subject` and `RawSubject` from the CSR (C3). `subject.organizations`, `organizationalUnits`, `countries` and `commonName` exist (C1). `literalSubject` gives an exact DN (C1). | Yes |
| CA-0026: valid 13 months from issue | `duration` is a Go duration (C1). `NotAfter = NotBefore + duration` (C3). `pkg/ca` uses `AddDate(0, 13, 0)` (`pkg/ca/authority.go:173`). | Approximately. 13 calendar months is 395 or 396 days. A fixed `duration: 9504h` (396 days) is near, but not the same rule. |
| CA-0027: never after the issuer expires | `pkg/ca` truncates `NotAfter` to the issuer (`pkg/ca/authority.go:174-175`). The C2 docs say that a leaf can outlive the CA. | No (per C2). Low effect: the intermediate is valid 10 years (CA-0006). |
| CA-0028: key usage `digitalSignature`, critical | `usages: ["digital signature", ...]` (C1). The default adds `key encipherment` (C1), so set `usages` explicitly. Go `crypto/x509` marks key usage critical (unconfirmed for the cert-manager path). | Yes, with explicit `usages` |
| CA-0029: not a CA | `isCA: false` sets basic constraints with CA false (C3). | Yes |
| CA-0030/0031/0032: one EKU, `serverAuth` or `clientAuth` | `usages: ["digital signature", "server auth"]` or `"client auth"` (C1). | Yes |
| CA-0033: SAN has the CN and each DNS name | `dnsNames` (C1). The CN is not added to the SAN automatically (unconfirmed). Give the CN also in `dnsNames`. | Yes, by configuration |
| CA-0034: one CRL distribution point | `spec.ca.crlDistributionPoints` on the Issuer (C1, C2). cert-manager does not make the CRL (C2). | Yes |
| CA-0035: no OCSP, no CA issuers URL | `ocspServers` and `issuingCertificateURLs` are optional (C1). | Yes |
| CA-0036: random serial over 64 bits | Not in C3. Unconfirmed. | Unconfirmed |
| CA-0056: SKI = SHA-1 of the public key | Not in C3. Unconfirmed. | Unconfirmed |
| CSR profile: RSA 2048, SHA-256 with RSA | `privateKey.algorithm: RSA`, `size: 2048` (C1); `signatureAlgorithm` field exists (C1). | Yes |
| No certificate policies | No field exists (C1). | Yes (none required) |

Rules of the Register that a cert-manager CA issuer cannot do:

- The participant name checks: O of a server CSR is the brand or legal entity name (REG-0623, `services/register/design/certificate-management.md:39`); CN and O of a client CSR (REG-0606, REG-0607).
- No certificate for a REMOVED brand or software product (REG-0609, REG-0626).
- The event of a server certificate issue (REG-0630).
- The registration of a client certificate at issue, so that the token endpoint accepts it (REG-0610, REG-0611).
- The CA issuer needs the intermediate key in a Secret: in the namespace of an `Issuer`, or in the cluster resource namespace (default `cert-manager`) for a `ClusterIssuer` (C2). That is a second holder of the intermediate key. ADR 0022 puts the CA in the Register (`docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md:32-40`), and the project decides that CDR certificates come only from the Register.
- The Register can still revoke a certificate that it did not issue (REG-0701, `services/register/design/certificate-revocation.md:8`), so the CRL is not a blocker.

#### Options

1. CA issuer with the intermediate Secret. Technically near the profile. It breaks the decision "CDR certificates come only from the Register" and skips the Register checks and the client registration.
2. External issuer that calls the Register (section 3.2).
3. No cert-manager for CDR certificates (section 3.3).

#### Recommendation

Do not use the cert-manager CA issuer for CDR certificates. The profile gaps are small (13 months as a fixed duration, the issuer expiry cap, two unconfirmed items). The real blocker is the design: a second copy of the intermediate key, and no Register checks, events or client registration.

### 3.2. A cert-manager external issuer for cdr-lab

#### Facts

- An external issuer is a controller with its own Issuer CRD. It watches CertificateRequests, checks that `issuerRef` points to its issuer, signs the PEM CSR, writes the certificate into the CertificateRequest status, and sets the Ready condition (C4).
- It must sign only an Approved request. It does not approve requests itself (C4).
- cert-manager auto-approves only `cert-manager.io` issuers by default. For an external issuer, the cluster operator adds its signer names to the cert-manager Helm value `approveSignerNames`, or a separate approver (for example approver-policy) approves. An approver needs the `approve` verb on `signers` in the `cert-manager.io` group, cluster scoped (C6).
- On success, the status has the signed certificate and the CA, if available (C6).
- issuer-lib gives the controllers, the status handling, retries, and approval handling. The author implements `Check` (Issuer controller) and `Sign` (CertificateRequest controller) (C5).
- issuer-lib says: "This library's API is still subject to change", and recommends experiments only until there are 5+ open-source issuers (C5). No tagged release was shown (C5).
- The Register issue paths need the participant IDs in the path: `POST /v1/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue` and `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue` (`services/register/design/certificate-management.md:53,55`). The body is one PEM CSR, and the response is the leaf and the intermediate in PEM (REG-0602, REG-0612). The operator interface needs the bearer token (`services/register/design/operator-interface.md:72`).
- The CA bundle (intermediate and root) is `GET /v1/cdr-ca-bundle` (REG-0633).

#### How it would call the Register

1. cert-manager makes the key and the CSR from a `Certificate` (subject, `dnsNames`, `usages`, `privateKey`).
2. The issuer controller gets an Approved CertificateRequest for its Issuer kind.
3. The Issuer spec gives the Register internal URL, a Secret reference for the operator token, the kind (server or client), and the participant IDs (one Issuer for each brand or software product; or IDs on annotations of the Certificate, which reach the CertificateRequest: unconfirmed).
4. `Sign` sends the CSR to the Register issue path. It puts the leaf in the certificate field and the CDR CA bundle in the CA field.
5. cert-manager writes the Secret (`tls.crt`, `tls.key`, `ca.crt`) and renews it by `renewBefore` (C1).

#### Size estimate (components and moving parts)

New in the repository:

1. A Go API package with the Issuer and ClusterIssuer CRD types, and generated CRD manifests.
2. A controller binary (issuer-lib, controller-runtime) with `Check` and `Sign`, and a Register client.
3. A container image, its CI job and its publication.
4. A Helm chart: Deployment, ServiceAccount, ClusterRole and bindings (CertificateRequests and their status, the Issuer CRDs, Events, Secrets for the token, a Lease for leader election), and the CRDs.
5. A specification, an ADR, and tests (envtest or a kind cluster in CI).

Outside the repository, in each cluster:

6. cert-manager installed (exists in the target clusters).
7. A change to the cluster-wide cert-manager configuration (`approveSignerNames`), or approver-policy with a policy, or an approver with cluster-scoped `approve` on `signers` (C6).
8. Cluster-scoped CRDs. A shared development cluster can need an administrator for this.

Then the app still needs a reload of a renewed certificate (section 3.3).

Assessment: not easy. It is a new long-running component with CRDs, cluster-scoped RBAC, a cluster-wide approval change, and a pre-v1 library (C5). It also does not help compose, so compose needs a second path. This disagrees with the "one shared issue path" decision.

#### Options

- Build the external issuer (above).
- Do not use cert-manager for CDR certificates (section 3.3, recommended).

#### Recommendation

Do not build an external issuer now. Record it as a possible later change if a cluster needs CDR certificates for many workloads that already use cert-manager.

### 3.3. Request and renew CDR certificates from the Register in a cluster, with no host step and no cert-manager

#### Facts

- The Register and the data holder load each TLS certificate one time at start, with `tls.LoadX509KeyPair` (`services/register/internal/root/config.go:65,76`; `services/dataholder/internal/root/config.go:176,187`). They put it in `tls.Config.Certificates` (`pkg/cds/tls.go:14-23`). No `GetCertificate`, no file watch. **A changed certificate file needs a restart today.**
- The charts mount each Secret as a directory, not with `subPath` (`charts/dataholder/templates/common.yaml:22-30,166-171`). Kubernetes updates the files of a mounted Secret after a change, but not for a `subPath` mount (C13). Thus an in-app reload would see a renewed Secret.
- The charts make no Secrets (`charts/dataholder/README.md:73`, DH-1344 at `services/dataholder/design/deployment.md:80`; `charts/register/README.md:13`).
- The Register never gets the private key of a certificate that it issues (REG-0603, `services/register/design/certificate-management.md:13`).
- The Register also needs a CDR server certificate for its own secure listener (`charts/register/values.yaml:55`). Today the offline `pki issue` makes it (`mise.toml:222-223`).
- Component rule: real components must not contain mock-specific code (README "Component types"). The Register and the data holder are both mocks. ADR 0022 says issue and revocation are lab-only features of the mock Register, and "Real components must not depend on them" (`docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md:44`).
- Reloader restarts a workload when a referenced Secret changes. Its `annotations` strategy is for GitOps tools; the default `env-vars` strategy can cause drift in Argo CD or Flux (C12).
- Argo CD maps Helm `pre-install` and `pre-upgrade` both to PreSync, so a "pre-install" hook runs at each sync (C9).
- Argo CD tracks resources by its tracking annotation (default) or label (C10). A Secret that a Job makes at run time with no tracking metadata is not in the manifests. That it is then outside the sync status and not pruned is inferred from C10, not stated: unconfirmed.

#### Options

**(a) CronJob with the `cdrlab` image.** It reads the current Secret. When the certificate is absent or near expiry (for example less than 30 days), it makes a new RSA 2048 key and CSR, sends the CSR to the Register issue path, gets the CA bundle, and writes a `kubernetes.io/tls` Secret.

- Moving parts: CronJob, ServiceAccount, Role and RoleBinding, the operator token Secret, the same image as compose.
- RBAC: `get`, `update`, `patch` on the named Secret (`resourceNames`), and `create` on Secrets (Kubernetes cannot limit `create` by name).
- First issue: the data holder pod cannot start until its Secret exists. Also run the same command as a Job at install (or a CronJob run at once), after the Register is ready.
- New certificate in the pod: in-app reload (needs code: `GetCertificate` that reads the files again), or Reloader (an extra cluster component), or the Job runs `kubectl rollout restart` (needs `patch` on the Deployment).
- GitOps: the Secret is not in the manifests, so no drift (unconfirmed, see facts). A plain Job is immutable in Argo CD; a CronJob avoids this.
- Failures: Register down or token wrong → the job fails, the old certificate stays valid; retry at the next run. Months of margin with a 13-month certificate. A failed run is visible as a failed Job only.
- Component rule: no code in either service. The issue logic is in the CLI. Good fit.

**(b) The data holder requests its own CDR server certificate** at start and before expiry.

- Moving parts: a Register client and a timer in the data holder, and the operator token in the data holder (or a new Register endpoint with a different credential).
- Key: in memory. Each restart makes a new key and a new certificate, and each issue is a Register event (REG-0630), so the history fills with issues.
- RBAC: none. Reload: in-app, by design.
- GitOps: no Secret, no drift.
- Failures: the data holder cannot start while the Register is down. Today a Register outage only makes the pod not ready (`charts/dataholder/README.md:90`). The Register must have the brand before the data holder starts (`charts/register/README.md:32`).
- Component rule: allowed, because the data holder is a mock. But it couples a mock data holder to the Register operator API, and the data holder then holds an operator token of the Register. A real third-party data holder in the lab cannot use this path.

**(c) The Register writes Secrets itself.**

- Moving parts: a Kubernetes client in the Register, cluster-wide or multi-namespace RBAC on Secrets.
- It must make the private key, which breaks REG-0603. Compose has no Kubernetes API, so compose needs a different path. Kubernetes-specific code in the Register.
- Not a fit.

**(d) Helm hook Job at install.**

- Moving parts: a Job with a hook annotation, RBAC as (a).
- Issues only at install. No renewal after 13 months unless an upgrade runs it again. In Argo CD it runs at each sync (C9), so it must be idempotent: check the Secret first.
- Good only as the first-issue step of (a).

**The Register's own secure certificate.** The Register has the intermediate key. Options: it issues its own CDR server certificate at start in-process, with a key in memory (no REG-0603 issue: it is its own key, not a participant key); or the init step issues it offline with the intermediate key, as `mise.toml` does now.

#### Recommendation

Option (a) with (d) as its first run: one `cdrlab` command that issues from the Register and writes a Secret, run by a CronJob (renewal) and by an idempotent install Job or hook (first issue). Add an in-app certificate reload to the shared TLS code (`pkg/cds/tls.go`), so that a renewed Secret takes effect without a restart in clusters and in compose. Do not depend on Reloader. It is an extra cluster component, and the target clusters may not have it.

### 3.4. The lab CA roots in a cluster

#### Facts

- `pki init` makes three RSA 4096 keys (CDR root, CDR intermediate, public root) and writes the lab directory if it is not complete (`pki/cmd/pki/main.go:87-126`).
- The Register reads the intermediate certificate and key from files (`CDR_INTERMEDIATE_KEY_FILE`, `compose.yaml:80-81`; `services/register/internal/root/config.go:113-128`). It never reads a root key (PKI-0076). In the chart, the intermediate is a `kubernetes.io/tls` Secret (`charts/register/values.yaml:56-57`).
- The CDR CA bundle is a Secret with `ca.crt` for each service (`charts/register/values.yaml:58-59`; `charts/dataholder/README.md:81`).
- The root key is needed only to make the intermediate. The Register never needs it (`docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md:38`).
- A Secret that a Job makes is not part of the Helm release, so `helm uninstall` does not delete it. A chart resource with `helm.sh/resource-policy: keep` is not deleted, but becomes orphaned, and can conflict with `helm install --replace` (C8). Argo CD maps `keep` to `Delete=false` (C9).
- A chart cannot make the CA with `lookup` under Argo CD: Argo CD renders with `helm template`, where `lookup` returns nothing (C11). Random values in templates keep the app OutOfSync (C9).
- CRL numbers come from the time (PKI-0066, `pki/DESIGN.md:83`), so a new Register database with an old CA still gives higher CRL numbers.

#### Options

1. **Install Job: `cdrlab pki init` to Secrets if absent.** The Job writes the intermediate Secret (cert and key), the CDR CA bundle Secret, and, for a cluster with no cert-manager, the public root. It does not keep the root key (or keeps it in a separate Secret that no workload mounts). RBAC: `get`/`create` on Secrets in the namespace. Idempotent: if the Secrets exist, it does nothing. Secrets stay after uninstall; a reinstall reuses them.
2. **The CLI writes Secrets through the Kubernetes API from a workstation.** It is a host step, which the decision excludes for clusters.
3. **User-provided Secrets.** The current chart contract (`charts/register/README.md:13`). Keep it as the override: when the values give Secret names, the Job does not run.

Persistence cases:

- Uninstall and reinstall, Secrets kept, database kept: no change.
- Secrets kept, database new: valid. Old certificates stay valid; the revocations are lost (PKI-0065 effect).
- Secrets deleted, database kept: a new CA. Each issued certificate fails, and the database has revocations of the old intermediate. Needs a full reset of the lab.

#### Recommendation

Option 1, with option 3 as the override. Write all CA Secrets from one Job in the Register namespace. Data holder releases in other namespaces need a copy of the CA bundle Secret: the umbrella chart puts all releases in one namespace, or the Job writes the bundle to each namespace (cross-namespace RBAC). The one-namespace choice is the less complex.

### 3.5. Public TLS with cert-manager

#### Facts

- The data holder terminates TLS on its public listener (8081) and the bank site (8085) with `pki.publicTLSSecret` (`charts/dataholder/README.md:79`, `charts/dataholder/README.md:93`). The Register terminates TLS on its public listener (`charts/register/values.yaml:53-54`).
- The charts expect `kubernetes.io/tls` Secrets with `tls.crt` and `tls.key` (`charts/dataholder/values.yaml:76-79`). A cert-manager Certificate writes a Secret named by `secretName` with these keys, and `ca.crt` (C1, C2).
- The data holder trusts the Register and software products through `pki.publicRootSecret` (key `ca.crt`), or the system roots (`charts/dataholder/README.md:83`).
- The public lab CA profile (CA-0038, CA-0039) applies to the mock public CA only. A public CA of a cluster has its own profile.
- The `pki issue-public` command makes the key and signs with the public root (`pki/cmd/pki/main.go:175-214`).

#### Options

1. The umbrella chart makes a `Certificate` for each public listener when `certManager.issuerRef` is set, with `secretName` equal to the value of `pki.publicTLSSecret`. The DNS names come from the host names of `publicBaseURL`. Renewal: cert-manager renews; the app reload of section 3.3 takes it.
2. When the cluster CA is not publicly trusted, give its root as `pki.publicRootSecret` to the data holder and the Register (trust distribution, for example trust-manager: not checked).
3. Compose and clusters with no cert-manager: `cdrlab issue-public` writes the same Secret (cluster) or files (compose), from the public root of section 3.4.

#### Recommendation

Option 1 for clusters with cert-manager, and option 3 for the rest. The chart keeps `pki.publicTLSSecret` as the one input, so the services do not know the source. The Certificate is a chart resource in git, so GitOps has no drift; the Secret is owned by cert-manager.

### 3.6. Recommended design

#### Cluster

1. **CA set-up:** an idempotent install Job (`cdrlab pki init`) writes the intermediate Secret, the CDR CA bundle Secret and, with no cert-manager, the public root Secret, only when they are absent. User-provided Secrets override it.
2. **Register secure certificate:** the Register issues its own CDR server certificate from its intermediate at start, with a key in memory, and before expiry. This removes the chicken-and-egg step. (Alternative: the install Job issues it offline, as `mise.toml:222-223` does now.)
3. **Data holder CDR server certificates and data recipient client certificates:** one `cdrlab` command sends a CSR to the Register operator API and writes a `kubernetes.io/tls` Secret. A CronJob runs it each day; it renews only when the certificate is absent or near expiry. An install Job (or Argo CD-safe hook) runs it once after the Register is ready.
4. **Public TLS:** cert-manager `Certificate` resources that write the existing `pki.publicTLSSecret` Secrets; without cert-manager, `cdrlab issue-public` in the same Jobs.
5. **Reload:** the shared TLS code reads the certificate files again when they change (`GetCertificate`), for the public and the secure listeners of both services. No restart and no Reloader.

#### Compose (the shared path)

The same `cdrlab` commands, with a file output instead of a Secret output: `cdrlab pki init` before the Register starts; then a one-shot container, after the Register is healthy, issues each data holder CDR server certificate and each test client certificate through the Register operator API. This replaces the offline `pki issue` of CDR certificates in `mise.toml:209-263`, so the Register is the only issuer in both environments (ADR 0022).

#### Reasons

- The Register stays the only CDR issuer, with its checks, its events and the client registration (REG-0606 to REG-0611, REG-0623 to REG-0630).
- One issue path for compose and clusters, as the decision requires.
- No code in the data holder (a mock that must not need the Register operator token). No Kubernetes code in the Register. REG-0603 holds: the key is made where the certificate is used.
- No cluster-wide change: no CRDs, no cert-manager approval change, no Reloader.
- GitOps: the Jobs and the CronJob are in git; the generated Secrets are not (no drift, unconfirmed for Argo CD in C10).

#### What the Register operator API needs

- **Nothing new for issue and renewal.** The existing issue paths take a CSR and give the chain (REG-0632, REG-0634), and the CA bundle path exists (REG-0633). A renewal is a new issue: CA-0055 permits the same subject.
- Possible small additions (decide in the plan):
  - The CLI must know the legal entity name for the O of a server CSR (REG-0623). It can read it from an existing participant read path of the operator interface, or take it as a flag. Not checked which read paths exist.
  - An option to revoke the previous certificate with the reason `superseded` after a renewal. The existing revoke path does this already (REG-0731); the CLI can call it. No API change.
- New Register requirements (not API): self-issue of its own secure certificate (step 2), and the TLS reload (step 5, in `pkg/cds`).

#### Open items

- Unconfirmed: cert-manager serial size, SKI, critical key usage, CN in SAN (section 3.1). Not needed if cert-manager is not used for CDR certificates.
- Unconfirmed: Argo CD behaviour for a run-time Secret with no tracking metadata.
- The CRL URL in each certificate (`charts/register/values.yaml:15-16`) must be reachable by each relying party in the cluster. Not researched here.

## 4. Exposure, operator access, a cluster in CI and PostgreSQL

Date checked: 2026-11-22. The repository facts come from `main` at commit `1a45e0b`.

### Sources of part 4

| ID | Source | URL |
|---|---|---|
| D1 | Gateway API: TLSRoute reference | https://gateway-api.sigs.k8s.io/reference/api-types/tlsroute/ |
| D2 | Kubernetes blog: Gateway API v1.5 (2026-04-21) | https://kubernetes.io/blog/2026/04/21/gateway-api-v1-5/ |
| D3 | Gateway API releases (v1.7.0 latest, 29 Oct) | https://github.com/kubernetes-sigs/gateway-api/releases |
| D4 | Gateway API conformance reports, v1.6 directory | https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/v1.6 |
| D5 | Gateway API implementations page | https://gateway-api.sigs.k8s.io/implementations/ |
| D6 | Kubernetes blog: Ingress NGINX retirement (2025-11-11) | https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/ |
| D7 | kubernetes/ingress-nginx repository (archived) | https://github.com/kubernetes/ingress-nginx |
| D8 | Kubernetes docs: Ingress | https://kubernetes.io/docs/concepts/services-networking/ingress/ |
| D9 | oauth2-proxy: configuration overview (v7.15.x) | https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview |
| D10 | oauth2-proxy: providers | https://oauth2-proxy.github.io/oauth2-proxy/configuration/providers/ |
| D11 | Envoy Gateway: OIDC task (v1.10) | https://gateway.envoyproxy.io/docs/tasks/security/oidc/ |
| D12 | Envoy Gateway: API extension types (`passThroughAuthHeader`) | https://gateway.envoyproxy.io/docs/api/extension_types/ |
| D13 | Pomerium: Gateway API | https://www.pomerium.com/docs/deploy/k8s/gateway-api |
| D14 | Istio: RequestAuthentication | https://istio.io/latest/docs/reference/config/security/request_authentication/ |
| D15 | Tailscale: Kubernetes operator cluster ingress | https://tailscale.com/kb/1439/kubernetes-operator-cluster-ingress |
| D16 | Kubernetes docs: RBAC | https://kubernetes.io/docs/reference/access-authn-authz/rbac/ |
| D17 | GitHub docs: GitHub-hosted runners | https://docs.github.com/en/actions/reference/runners/github-hosted-runners |
| D18 | kind: quick start | https://kind.sigs.k8s.io/docs/user/quick-start/ |
| D19 | kind: configuration | https://kind.sigs.k8s.io/docs/user/configuration/ |
| D20 | kind releases (v0.33.0) | https://github.com/kubernetes-sigs/kind/releases |
| D21 | cloud-provider-kind README (v0.12.0) | https://github.com/kubernetes-sigs/cloud-provider-kind |
| D22 | helm/kind-action | https://github.com/helm/kind-action |
| D23 | k3s: networking services | https://docs.k3s.io/networking/networking-services |
| D24 | k3d releases (v5.9.0) | https://github.com/k3d-io/k3d/releases |
| D25 | Bitnami: catalog changes notice (bitnami/containers #83267) | https://github.com/bitnami/containers/issues/83267 |
| D26 | CloudNativePG repository (v1.30.2, Apache-2.0, CNCF Incubating) | https://github.com/cloudnative-pg/cloudnative-pg |
| D27 | Envoy Gateway releases (v1.10.0, 2026-11-06) | https://github.com/envoyproxy/gateway/releases |

Versions that the research checked with the GitHub API on 2026-11-22: Gateway API v1.7.0; Envoy Gateway v1.10.0; oauth2-proxy v7.15.6; CloudNativePG v1.30.2 (chart `cloudnative-pg-v0.29.2`); kind v0.33.0 (Kubernetes 1.37.0); cloud-provider-kind v0.12.0; k3d v5.9.0; helm/kind-action v1.15.1; bjw-s `app-template-5.3.0` (the common library version itself is not checked here).

### 4.1. Exposure of the TLS and mTLS listeners

#### Facts

- TLSRoute is GA. It is in the Standard channel since Gateway API v1.5.0 ([D1], [D2]). v1.5 was released on 2026-02-27, and v1.5 also moved ListenerSet, ReferenceGrant and client certificate validation to Standard ([D2]). The latest release is v1.7.0 ([D3]).
- TLSRoute has two modes: `Passthrough` (the Gateway does not decrypt) and `Terminate` ([D1]). `hostnames` is required, each is an FQDN, and the Gateway matches them against the SNI of the TLS handshake ([D1]). `TLSRouteModeMixed` (Passthrough and Terminate on one port) is an optional feature ([D1]); v1.7.0-rc.3 marks it Standard ([D3]).
- Conformance reports for Gateway API v1.6.x, profile `GATEWAY-TLS` (the core tests are the passthrough TLSRoute tests). Each of these reports says "Core tests succeeded" ([D4]):

  | Implementation | Version in report | Gateway API | `TLSRouteModeTerminate` / `TLSRouteModeMixed` |
  |---|---|---|---|
  | Envoy Gateway | v1.9.0 | v1.6.1 | supported |
  | Istio | 1.31 | v1.6.0 | supported |
  | Cilium | v1.20.0 (standard and experimental) | v1.6.1 | not supported |
  | NGINX Gateway Fabric | v2.8.0 | v1.6.3 | supported |
  | Traefik | v3.7.10 | v1.6.1 | supported (its HTTP profile result is "partial") |
  | kgateway | v2.4.5 | v1.6.1 | supported |

  Envoy Gateway v1.10.0 (2026-11-06) is newer than its report ([D27]). The reports for Gateway API v1.7 are unconfirmed.
- ingress-nginx is retired. Best-effort maintenance stopped in March 2026, with "no further releases, no bugfixes, and no updates to resolve any security vulnerabilities" ([D6]). The repository is archived since 2026-03-24 ([D7]). Its `ssl-passthrough` annotation is therefore not a valid recommendation.
- The Ingress API is frozen, and the Kubernetes project recommends Gateway ([D8]). Ingress "only supports rules for directing HTTP(S) traffic", and other protocols use NodePort or LoadBalancer Services ([D8]). Ingress has no standard passthrough field (passthrough is only a controller annotation).
- The charts make ClusterIP Services only, and no ingress (`charts/register/README.md:20-21`, `charts/dataholder/README.md:56-57`). The secure listener needs mTLS end to end, so only TCP/TLS passthrough can be in front of 8083 (`charts/dataholder/README.md:57`). Behind a proxy on 443, `allowedHosts` gets the host name with no port (`charts/register/README.md:20`).
- The Register advertises the base URIs of each brand from its seed. The example seed gives `https://wattle-bank:8081` (public, product, infosec) and `https://wattle-bank:8083` (resource) (`services/register/seed/example.json:19-22`). The base URIs must agree with `publicBaseURL` and `secureBaseURL` of the data holder release (`charts/register/README.md:32`). The Register has its own `PUBLIC_BASE_URL` and `SECURE_BASE_URL` (`compose.yaml:72-75`).
- The lab certificates have the service name and `localhost` as SANs, for example `DNS:register,DNS:localhost` (`mise.toml:223`, `mise.toml:233-236`).
- The CRL URL is a plain HTTP URL in each issued certificate, from `CDR_INTERMEDIATE_CRL_URL` (`compose.yaml:82`, `services/register/internal/root/config.go:113`). Only the path `/crl/cdr-intermediate.crl` of the internal listener goes outside the cluster ([ADR 0025](../adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md), `charts/register/README.md:22`).
- The CDS gives a separate public base URI and resource base URI for each brand, and RFC 8705 gives separate mTLS endpoint aliases (ADR 0025, context).

#### Options

1. **Gateway API, one Gateway, port 443, SNI routing.** One `TLS` listener in `Passthrough` mode on 443 with a wildcard host, and one TLSRoute for each of these: Register public, Register secure, each brand public, each brand secure, each bank site. Port 443 has only SNI to tell routes apart, so the public and the secure listener of one service need **different host names**, for example `register.lab.example` and `mtls.register.lab.example`. The base URLs then have no port. The certificate of each listener needs its external host name as a SAN, and also the Service name if in-cluster clients use the Service name.
2. **Gateway API, one Gateway, ports 8081 and 8083.** Two passthrough listeners, one on each port, and one host name for each service. The base URLs keep the ports, as in the lab. This needs fewer host names and certificates, but a shared cluster Gateway that the team does not own often has only 443 open (unconfirmed for each team; it is a platform choice).
3. **The CRL and the console over HTTP routes.** An HTTPRoute with an exact path match for `/crl/cdr-intermediate.crl` to the port `internal`, on an HTTP listener (port 80). The CRL URL in the certificates must then be the external URL. HTTPRoute cannot be used for 8081/8083, because the services terminate TLS themselves.
4. **LoadBalancer Services.** One LoadBalancer Service for each instance with 8081 and 8083 (and 8085). This needs no Gateway controller. It needs one external IP for each instance, and a cloud or MetalLB. It is the fallback for clusters with no TLSRoute support.
5. **Documentation only.** The chart gives no route, and the README gives the rules and an example.

Also note: with passthrough, the service sees the IP address of the gateway as the client IP, not the IP of the participant. PROXY protocol support differs for each implementation (unconfirmed).

In-cluster callers (the data holder calls the Register, the Register calls the data holder brands and the software products) use the advertised base URLs. If the base URLs use external host names, the pods must resolve these names, which can mean hairpin traffic through the Gateway, or split-horizon DNS (unconfirmed for each cluster). A simpler choice for a dev cluster: in-cluster callers use the Service name (as `REGISTER_PUBLIC_BASE_URL` allows today), and the certificates have both names.

#### Recommendation

- The umbrella chart offers **optional** Gateway API templates, off by default: one TLSRoute (`gateway.networking.k8s.io/v1`, Passthrough) for each TLS listener, and one HTTPRoute for the CRL path only. The values give the `parentRefs` (an existing Gateway of the team) and the host names. The chart does not make a Gateway, because the Gateway belongs to the platform team in a shared cluster.
- Default the docs to option 1 (443 and SNI, separate host names for public and secure). Permit option 2 by values.
- Document LoadBalancer Services as the fallback, with values of the bjw-s `service` map, and no new template.
- Do not mention ingress-nginx, except to say that it is retired. Do not make Ingress templates.
- Never route the internal listener with an HTTPRoute except the CRL path, and the console through the auth proxy of section 4.2.

### 4.2. Interim auth for the operator console and the operator interface

#### Facts

- The internal listener has these route groups (`services/register/internal/root/serve.go:65-157`):
  - `/livez`, `/readyz`: no `Host` check and no auth (`pkg/platform/httpserver/probes.go:12-15`, ADR 0025).
  - `/crl/cdr-intermediate.crl` on the Register: no `Host` check and no auth (ADR 0025).
  - The operator interface `/v1/...`: one static bearer token in `Authorization: Bearer` (`pkg/platform/operatorhttp/operatorhttp.go:62-73`, `services/register/internal/adapter/operatorapi/operatorapi.go:101`).
  - The console: a single-page app at `/` and a same-origin API at `/api/` with no credential and no login (`pkg/platform/consolehttp/consolehttp.go:104-110`, ADR 0014). The data holder console shows names, email addresses and customer numbers (`docs/production-gaps.md:26`).
  - The operator interface and the console share the `Host` allowlist `INTERNAL_ALLOWED_HOSTS` (ADR 0025). The console uses `http.CrossOriginProtection` (ADR 0014).
- The repository has no operator CLI (`tools/` has `notices` and `speccheck` only). The clients of the operator API today are the conformance suite, the web tests and HTTPie (`compose.yaml:397-402`, `mise.toml:158`).
- oauth2-proxy (v7.15.6): providers include GitHub, Google, Microsoft Entra ID, Keycloak OIDC and generic OIDC ([D10]). `skip_auth_routes` takes `method=path_regex` rules ([D9]). `skip_jwt_bearer_tokens` passes requests with a verified JWT bearer token whose `aud` matches ([D9]). It sends `X-Forwarded-User`, `X-Forwarded-Email` and related headers to the upstream by default (`pass_user_headers`) ([D9]).
- Envoy Gateway (v1.10): a SecurityPolicy with `oidc` targets a Gateway or an HTTPRoute ([D11]). `passThroughAuthHeader` skips OIDC when the request has a header that the JWT filter extracts, and it needs a `jwt` provider ([D12]). Thus the static operator token, which is not a JWT, cannot pass through the OIDC policy. Envoy Gateway has a separate external authorization feature (detail not confirmed).
- Pomerium: experimental Gateway API support; a `PolicyFilter` CRD on each HTTPRoute; it denies each route with no explicit policy; it uses session cookies ([D13]). Its CLI and bearer support are unconfirmed.
- Istio `RequestAuthentication` only validates JWTs. It has no browser login. A request with no token passes unless an AuthorizationPolicy refuses it ([D14]). Istio needs oauth2-proxy (or similar) for the browser login.
- Tailscale operator: an Ingress with `ingressClassName: tailscale`, or a Service annotation `tailscale.com/expose: "true"`, puts a Service on the tailnet with a Let's Encrypt certificate ([D15]). Identity headers are not described on that page (unconfirmed).
- Cloudflare Access with `cloudflared`: not checked in this research (unconfirmed).
- `kubectl port-forward` uses the Kubernetes API, so Kubernetes RBAC controls it. The rule is usually `create` on `pods/portforward` ([D16]; the research did not read the exact text, unconfirmed). With port-forward, the `Host` header is `localhost:<local port>`, so `INTERNAL_ALLOWED_HOSTS` must have that value, as the lab has `localhost:8180` (`compose.yaml:146`).

#### Options

| Option | Protects | Moving parts | Service change | How a CLI reaches `/v1/` |
|---|---|---|---|---|
| A. `kubectl port-forward` + RBAC | Console and API: only people with `pods/portforward` in the namespace | None in the cluster | None. Add `localhost:<port>` to `INTERNAL_ALLOWED_HOSTS` | Port-forward, then the static bearer token |
| B. oauth2-proxy in front of the internal listener (sidecar or Deployment), with an HTTPRoute to it | Browser console (OIDC session) | oauth2-proxy, an OIDC client, a cookie secret, a route, a host name | None if the proxy sends the original `Host` and the external host is in `INTERNAL_ALLOWED_HOSTS`. Do not trust `X-Forwarded-User` until the service has operator identity | `skip_auth_routes` for `/v1/` (the static token then guards it alone, over the route), or port-forward (option A) for the API |
| C. Envoy Gateway `SecurityPolicy` OIDC on an HTTPRoute | Browser console | Needs Envoy Gateway as the team's Gateway; an OIDC client | As B | A second HTTPRoute for `/v1/` with no OIDC policy, or port-forward. `passThroughAuthHeader` does not fit the static token |
| D. Pomerium | Browser console | Pomerium controller, IdP | As B | Unconfirmed |
| E. Istio `RequestAuthentication` + oauth2-proxy | As B | Istio mesh, oauth2-proxy | As B | As B |
| F. Tailscale or Cloudflare Access (outside the cluster) | Console and API: only tailnet or Access users | Operator or `cloudflared`, a vendor account | `Host` allowlist gets the tailnet or Access host name | Over the tailnet, with the static token |

Each option leaves `/livez` and `/readyz` reachable by the kubelet on the pod IP, because none of them change the Service. The CRL route is separate from the console route (section 4.1), so the auth proxy does not apply to it.

#### Recommendation

- **Default to document: option A** (`kubectl port-forward` with Kubernetes RBAC). It needs no new component, no IdP, and no service change. It protects the console and the operator API the same way. The access control is the RBAC that each team already has. It agrees with the chart rule "do not expose the internal listener outside the cluster" (`charts/dataholder/README.md:54`).
- **Document option B** as the choice for a team that wants a browser URL, with an example values file for the oauth2-proxy Helm chart (an upstream chart), not a template in the umbrella chart. Route only the console paths through it, and keep `/v1/` on port-forward.
- **Do not add an oauth2-proxy sidecar to the chart now.** The bjw-s common library can already add a sidecar container from values, so a team can do it with no chart change (bjw-s sidecar support is from general knowledge, unconfirmed here). A sidecar adds an IdP, secrets and a cookie domain to each brand release.
- Update the operator interface and console rows of `docs/production-gaps.md` when the shared-cluster deployment exists: the "port binds to `127.0.0.1`" reason does not apply in a cluster.

### 4.3. A Kubernetes cluster in GitHub Actions CI

#### Facts

- Standard hosted Linux runner for a public repository: 4 vCPU, 16 GB RAM, 14 GB SSD, x64 (`ubuntu-latest`, `ubuntu-24.04`, `ubuntu-22.04`, `ubuntu-26.04`); arm64 is the same size ([D17]). They are free and have no limit for public repositories ([D17]). All CI jobs of the repository use `ubuntu-latest` (`.github/workflows/ci.yml:35-647`).
- kind v0.33.0, default Kubernetes 1.37.0 ([D20]). `kind load docker-image` and `kind load image-archive <tar>` copy local images into the nodes ([D18]). With a `:latest` tag or no tag, Kubernetes pulls again, so use a fixed tag or `imagePullPolicy: IfNotPresent` ([D18]). `extraPortMappings` maps node ports to the host, and a NodePort must equal the `containerPort` ([D19]). `helm/kind-action` v1.15.1 installs kind v0.33.0 and takes `config`, `cluster_name` and `node_image` ([D22]).
- cloud-provider-kind v0.12.0 gives LoadBalancer Services, and a Gateway API controller (standard channel by default) for `Gateway` and `HTTPRoute` only. It does not list TLSRoute ([D21]).
- k3d v5.9.0 runs k3s in Docker ([D24]). k3s deploys Traefik and ServiceLB by default; `--disable=traefik` and `--disable=servicelb` turn them off ([D23]).
- minikube: not checked (unconfirmed).
- Cluster create time on a hosted runner: no primary source. A typical kind single-node create is approximately 30 to 90 seconds (unconfirmed estimate).
- The CI `images` job already saves the compose-built images to a tar (`.github/workflows/ci.yml:224`). `kind load image-archive` can load that tar.
- The conformance suite today runs as the `conformance` compose service: `go test -tags conformance` against the mounted repository, with the source tree and the `.lab/pki` files mounted read-only (`compose.yaml:386-437`). It uses the service DNS names `register`, `wattle-bank` and `mail` (`compose.yaml:397-404`).
- The data holder and the Register **call back to the suite**: the seed gives `https://conformance:8443` and `https://conformance:8444` as JWKS hosts and a recipient base URI (`compose.yaml:423-427`). Thus the suite must be reachable by the name `conformance` from the pods, with a certificate for that name (`.lab/pki/conformance/jwks.crt`).
- The suite needs the mail sink (`MAIL_SINK_URL: http://mail:8025`, `compose.yaml:404`). CONF-0010 says the suite runs in the compose service on the lab network (`conformance/DESIGN.md`, CONF-0010, CONF-0037). A cluster run changes this requirement.
- The base URLs, the seed URIs and the certificate SANs must use the same names that the suite uses (`services/register/seed/example.json:19-25`, `mise.toml:223-236`). The chart keeps the full release name as the Service name (`charts/register/README.md:21`), so the release names must be `register` and `wattle-bank`, or the seed and the certificates change.
- Memory: compose sets no limits, and the charts set no resources by default (`charts/*/values.yaml:2`). No lab was running, so the research could not measure. Estimate (unconfirmed): kind control plane approximately 0.6 to 1 GB; PostgreSQL approximately 100 to 200 MB; each Go service approximately 50 to 150 MB; a Go `go test` build of the suite approximately 1 to 2 GB at peak; a Gateway controller (if used) approximately 100 to 300 MB. The total fits in 16 GB with a large margin.

#### Options

1. **kind + the suite as an in-cluster Job (recommended).** Build the images (as today), `kind load image-archive`, install PostgreSQL, Mailpit, the umbrella chart with release names `register` and `wattle-bank`, then run a Job named for a Service `conformance` (ports 8443, 8444). The Job image is the pinned `golang` image, and it gets the repository and `.lab/pki` by a kind `extraMounts` host path (kind feature, details unconfirmed), or a test binary from `go test -c` in a small image. No change to the test code: the names are the same as in compose.
2. **kind + the suite on the runner, with port-forward or NodePorts.** The suite reaches the services, but the services cannot reach the suite on the runner by the name `conformance` without extra DNS and routing. The base URLs and SANs would also need `localhost` ports. Not practical.
3. **k3d.** Similar to kind, with Traefik and ServiceLB that must be turned off to stay neutral. No clear advantage.
4. **A Gateway in kind.** Only needed to test the optional TLSRoute templates. cloud-provider-kind does not do TLSRoute ([D21]); Envoy Gateway or Istio would add install time. Rendering the templates in `lint:chart` (helm lint, kubeconform) can be enough at first.

#### Recommendation

- Use **kind** with `helm/kind-action` on `ubuntu-latest`, in one new CI job that runs when the charts or the services change.
- Run the conformance suite **as a Job in the cluster** with a Service `conformance`, so the names, the seed and the certificate SANs stay the same as in compose. Update CONF-0010 and CONF-0037 to permit the cluster environment.
- Install Mailpit as a plain Deployment in CI only.
- Do not install a Gateway in CI in the first change. Check the route templates with `lint:chart`. Add a Gateway test only if a template defect occurs.
- Measure the memory and time over approximately five runs before any claim (the practice of this repository).

### 4.4. PostgreSQL for the umbrella chart

#### Facts

- The charts keep PostgreSQL outside. An init container runs the migrations as the owner user. The service connects as the service user. The operator grants `register_app` or `dataholder_app` to the service user. The charts give CloudNativePG managed-role instructions (`charts/register/README.md:15`, `charts/dataholder/README.md:51`).
- The lab uses `postgres:18.6-trixie`, pinned by digest, with `lab/postgres/init.sql` (`compose.yaml:548-567`), and `lab/postgres/grants.sql` for the grants (README, Development).
- CloudNativePG v1.30.2 (2026-11-17), Apache-2.0, a CNCF Incubating project ([D26]). Helm chart `cloudnative-pg-v0.29.2`. Its install command and its kind notes: not read (unconfirmed).
- Bitnami: since 2025-08-28 `docker.io/bitnami` has only a limited set of `latest` tags for development. Versioned images moved to `docker.io/bitnamilegacy`, which gets "no further updates or support" ([D25]). The charts at `docker.io/bitnamicharts` get no updates, and "will not work out-of-the-box" without image overrides ([D25]). The chart source stays Apache 2.0 ([D25]). The supported path is the paid Bitnami Secure Images ([D25]).

#### Options

1. **CloudNativePG in kind for CI and quick start.** It matches the production advice of the charts, and it tests the managed-role and `inRoles` instructions. It adds an operator install and CRDs to CI.
2. **A plain postgres StatefulSet (or Deployment) in CI only**, with the same pinned image and the same `init.sql` and `grants.sql` as the lab, from a ConfigMap. Fewest moving parts; it reuses the lab files.
3. **A Bitnami PostgreSQL chart.** Not supported for free since 2025; needs `bitnamilegacy` images with no updates ([D25]). Reject.
4. **A PostgreSQL subchart in the umbrella chart.** It goes against "PostgreSQL is outside the chart" and against the own-database rules of ADR 0002 and ADR 0020 for each team. Reject.

#### Recommendation

- The umbrella chart has no PostgreSQL. Keep "bring your own" and document CloudNativePG as the quick start, with one example `Cluster` manifest that declares the roles, the databases and the grants for the Register and each brand.
- CI uses option 2 (a plain postgres with the lab `init.sql` and `grants.sql`), because it reuses tested files and needs no operator. If a later defect shows that the CloudNativePG instructions are wrong, move CI to option 1.
- Do not use Bitnami charts or images.

### 4.5. Unconfirmed items

- Gateway API v1.7 conformance reports for each implementation.
- Pomerium CLI and bearer support; Cloudflare Access; Tailscale identity headers; minikube.
- The exact RBAC text for `pods/portforward`.
- kind create time and the memory figures (estimates only).
- kind `extraMounts` details; bjw-s sidecar support; CloudNativePG install commands.
- PROXY protocol support for passthrough in each implementation.
