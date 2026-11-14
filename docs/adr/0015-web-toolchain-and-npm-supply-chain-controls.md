# 0015. Build the web UIs in one pnpm workspace, with explicit npm supply-chain controls

- Date: 2026-10-22

## Context

The web UIs are React single-page applications that Node builds ([ADR 0014](0014-web-uis-as-embedded-react-single-page-applications.md)). They need a package manager, a place in the repository, a compiler, a bundler, lint and format tools, and test tools. Several UIs share a design system, so they share packages.

npm packages are the largest new risk. In 2025 and 2026, attackers published malicious versions of popular packages many times. The research is in [Web front ends](../research/2026-10-22-web-front-ends.md), section 6. Examples:

- Shai-Hulud and its later waves ran from `preinstall` or `postinstall` scripts, stole tokens and published more infected packages ([CISA alert](https://www.cisa.gov/news-events/alerts/2025/09/23/widespread-supply-chain-compromise-impacting-npm-ecosystem), [Microsoft](https://www.microsoft.com/en-us/security/blog/2025/12/09/shai-hulud-2-0-guidance-for-detecting-investigating-and-defending-against-the-supply-chain-attack/)).
- The `debug` and `chalk` compromise had no install script. Its code ran in browser bundles ([GHSA-4x49-vf9v-38px](https://github.com/advisories/GHSA-4x49-vf9v-38px)). The malicious versions were removed on the same day.
- The TanStack compromise published versions with valid provenance, through the real trusted publisher, after CI cache poisoning. Its payload came from a git subdependency ([GHSA-g7cv-rxg3-hmpx](https://github.com/advisories/GHSA-g7cv-rxg3-hmpx)).

Script blocking and a minimum release age together stop each of these, except a payload that stays dormant for longer than the age. Provenance did not stop the TanStack compromise.

The package managers in October 2026:

- pnpm 11 and 12 block dependency build scripts and fail the install when a dependency has a build script that nobody approved (`strictDepBuilds`). They do not install a version that is newer than `minimumReleaseAge`. They block git and tarball subdependencies (`blockExoticSubdeps`). They can fail when the trust evidence of a package drops (`trustPolicy: no-downgrade`, which is off by default). pnpm 12 fails on a setting that it does not know ([pnpm settings](https://pnpm.io/settings/dependency-resolution), [pnpm 12.0.0](https://github.com/pnpm/pnpm/releases/tag/v12.0.0)). Gitea, Teleport, Nomad and Zitadel use pnpm.
- npm 12 also blocks scripts and git dependencies, but Node 26 and Node 24 still bundle npm 11 ([Node.js releases](https://nodejs.org/dist/index.json)). npm has no trust-downgrade check.
- Bun and Deno block scripts by default, but each is a second JavaScript runtime.

Renovate does not apply its minimum release age to lock file maintenance. Its preset tells the user to make the package manager do its own check ([Renovate security presets](https://github.com/renovatebot/renovate/blob/main/lib/config/presets/internal/security.preset.ts)). This repository uses `security:minimumReleaseAgeNpm`, which is 3 days.

Go does not ignore a `node_modules` directory. We measured with Go 1.27.2: a `.go` file in `node_modules/` makes `go list ./...` fail. The `go.mod` `ignore` directive removes a directory from `./...` ([Go modules reference](https://go.dev/ref/mod#go-mod-file-ignore)). Corepack is not in Node 25 and later ([Node.js 25 changelog](https://github.com/nodejs/node/blob/main/doc/changelogs/CHANGELOG_V25.md)), and mise installs Node with checksum and signature verification ([mise: Node](https://mise.jdx.dev/lang/node.html)).

TypeScript 7.0 is the native Go port, with strict mode on by default. It has no compiler API until 7.1, so a tool that needs the API must use TypeScript 6 through the npm alias `@typescript/typescript6` ([Announcing TypeScript 7.0](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/)). For lint and format, we compared:

| | oxlint and oxfmt | Biome | ESLint, typescript-eslint and Prettier |
|---|---|---|---|
| Type-aware rules | 59 of the 61 typescript-eslint type-aware rules, with the TypeScript 7 checker ([oxlint](https://oxc.rs/docs/guide/usage/linter/type-aware.html)) | Its own type inference, rules in the nursery group | Full, but typescript-eslint does not accept TypeScript 7 yet |
| Stability | oxlint 1.x; oxfmt is 0.x | 2.x | Stable |
| Dependencies | None at runtime | None at runtime | Many |

For component tests, Vitest Browser Mode runs the tests in a real browser, and it is stable since Vitest 4 ([Vitest 4](https://vitest.dev/blog/vitest-4)). jsdom and happy-dom do no layout and no real focus management.

## Decision

- All JavaScript and TypeScript code is in one pnpm workspace in `web/`. Each app is in `web/apps/<name>/`, and each shared package is in `web/packages/<name>/`. `go.mod` ignores `./web`. There is no build orchestrator (for example Nx or Turborepo). The workspace commands and the mise tasks give the order of the builds. The reason for one workspace is one lock file and one set of controls for all JavaScript code.
- `mise.toml` pins Node (an LTS version) and pnpm. Nothing else pins them, so there is one source of truth.
- `pnpm-workspace.yaml` sets each supply-chain control explicitly, also when the pnpm default is the same, so that a change of a default cannot weaken it:
  - a minimum release age that is the same as the Renovate minimum release age for npm, with the strict setting, so that a lock file maintenance update also waits;
  - `trustPolicy: no-downgrade`;
  - `blockExoticSubdeps: true` and `strictDepBuilds: true`;
  - `allowBuilds` as an allowlist. An entry needs a reason in a comment.
- CI installs from the lock file with no changes, checks the registry signatures (`pnpm audit signatures`), and scans the pnpm and Go lock files with OSV-Scanner, which also reports known malicious packages. npm updates have the same auto-merge policy as the other dependencies: a person reviews only a major update. The reason is that a person cannot usefully review the code of a minor update and of its transitive dependencies, and the controls above apply to each update.
- The code is TypeScript 7 in strict mode, built with Vite. A tool that needs the TypeScript compiler API uses `@typescript/typescript6` in its own package only, until it supports TypeScript 7.
- oxlint, with its type-aware rules, lints the code, and oxfmt formats it. The reason is the type-aware rules (for example floating promises) without a TypeScript 6 dependency.
- Component tests use Vitest Browser Mode with the Playwright provider, not jsdom or happy-dom. Tests stub the network with MSW, find elements by role and label, and do automated accessibility checks with axe-core. A small number of Playwright tests run user journeys against the lab deployment. The mise `fmt`, `lint` and `test` tasks include the workspace.
- The commit scope of a change in one app or package is the name of its directory. The scope `web` is for a change to the full workspace ([ADR 0007](0007-conventional-commits.md)).

## Consequences

- Development and CI need Node and pnpm, which `mise install` gives. CI also downloads a browser for the tests.
- A new version of a package can be installed only after the release age. A security fix of a dependency also waits, unless a change adds an exclusion for it.
- A new dependency with a build script fails the install until a change adds it to `allowBuilds`, with a reason.
- npm updates merge without a review by a person, except major updates. A malicious version that stays dormant for longer than the release age, and that the scans do not find, can get into the lock file.
- `go list ./...`, `go test ./...` and `go mod tidy` do not see the workspace.
- oxfmt is 0.x, so a new version can change the format of existing code.
- The TypeScript 6 alias stays until each tool that needs it supports TypeScript 7.
- Browser tests are slower than tests in jsdom, but a pass means that the behaviour works in a real browser.
- A UI is not in the directory of the service that serves it. The service build copies the UI build into its own package.
