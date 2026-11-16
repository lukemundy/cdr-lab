# Research: web front ends for cdr-lab

- Research date: 2026-10-22
- Scope: the technology stack, design system, repository layout, security, testing and telemetry of the web UIs of the lab: operator consoles (desktop) and consumer UIs (mobile) for the Register, the data holders and later data recipients. Also the CDR consumer experience (CX) requirements and the accessibility requirements that apply to these UIs.
- Applies to: CDS 1.36.0, CDR Rules Compilation No. 10, and the product versions in the table below.

Check the sources again before you depend on a fact. Front-end products change quickly. Some of the products in this note had a major release in the month before this note.

## Legend

- **UNVERIFIED**: I did not find a primary source. Do a check before you depend on it.
- **JUDGEMENT**: my opinion, not a fact from a source.
- **Measured**: a result from a command in this research session.

Versions, release dates and repository activity come from the GitHub REST API (`gh api repos/<owner>/<repo>/releases/latest` and `/commits`) and the npm registry (`registry.npmjs.org/<package>`), on 2026-10-22. Weekly download counts come from `api.npmjs.org/downloads/point/last-week/<package>` (week to 2026-10-21). Download counts include CI and transitive installs, so they are larger than real use.

## Versions examined

| Item | Version | Date | Source |
|---|---|---|---|
| React | 19.3.1 | 2026-10-14 | https://github.com/facebook/react/releases/tag/v19.3.1 |
| React Compiler (`babel-plugin-react-compiler`) | 1.0.0 | 2025-10-07 | https://react.dev/blog/2025/10/07/react-compiler-1 |
| Next.js | 16.4.2 | 2026-10-15 | https://github.com/vercel/next.js/releases |
| React Router | 8.5.0 | 2026-10-19 | https://github.com/remix-run/react-router/releases |
| TanStack Router / Start | 1.170.45 / 1.169.3 | 2026-10-16 | https://www.npmjs.com/package/@tanstack/react-router , https://tanstack.com/blog/tanstack-start-1-0 |
| Vue | 3.5.44 (3.6 with Vapor Mode is at RC 12) | 2026-10-13 | https://github.com/vuejs/core/releases |
| Svelte / SvelteKit | 5.57.2 / 3.1.0 | 2026-10-06 / 2026-10-14 | https://github.com/sveltejs/svelte/releases , https://github.com/sveltejs/kit/releases |
| Angular | 22.2.3 | 2026-10-14 | https://github.com/angular/angular/releases |
| Solid | 1.9.17 (2.0 is at RC 17) | 2026-10-07 | https://www.npmjs.com/package/solid-js |
| htmx | 4.0.0 (npm `latest` stays on 2.x until early 2027) | 2026-08-28 | https://github.com/bigskysoftware/htmx/releases/tag/v4.0.0 |
| templ | v0.3.1070 | 2026-10-04 | https://github.com/a-h/templ/releases |
| Datastar | v1.0.4 | 2026-09-21 | https://github.com/starfederation/datastar/releases |
| Node.js | 26.12.1 (LTS from 2026-10-28); 24.22.1 (maintenance) | 2026-10-21 | https://nodejs.org/dist/index.json , https://github.com/nodejs/Release/blob/main/schedule.json |
| pnpm | 12.15.0 | 2026-10-21 | https://github.com/pnpm/pnpm/releases/tag/v12.15.0 |
| npm CLI | 12.2.1 (Node 26 still bundles npm 11.20.0) | 2026-10-15 | https://github.com/npm/cli/releases , https://nodejs.org/dist/index.json |
| Bun / Deno | 1.4.4 / 2.9.8 | 2026-10-22 / 2026-10-21 | https://github.com/oven-sh/bun/releases , https://github.com/denoland/deno/releases |
| TypeScript | 7.0.2 (native Go port); 7.1 beta | 2026-07-08; 2026-10-20 | https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/ |
| Vite | 8.3.5 (Rolldown bundler) | 2026-10-21 | https://github.com/vitejs/vite/releases |
| oxlint / oxfmt | 1.89.0 / 0.74.0 | 2026-10-19 | https://github.com/oxc-project/oxc/releases |
| Biome | 2.6.2 | 2026-10-21 | https://github.com/biomejs/biome/releases |
| ESLint / Prettier | 10.13.0 / 3.9.10 | 2026-10-16 / 2026-10-10 | https://github.com/eslint/eslint/releases , https://github.com/prettier/prettier/releases |
| openapi-typescript / openapi-fetch | 7.13.0 / 0.17.0 | 2026-02-11 | https://github.com/openapi-ts/openapi-typescript/releases |
| orval | 8.44.0 | 2026-10-21 | https://github.com/orval-labs/orval/releases |
| Tailwind CSS | 4.3.3 | 2026-07-16 | https://github.com/tailwindlabs/tailwindcss/releases |
| shadcn CLI | 4.22.1 | 2026-10-18 | https://github.com/shadcn-ui/ui/releases |
| Base UI (`@base-ui/react`) | 1.9.0 (1.0.0 on 2025-12-11) | 2026-10-09 | https://base-ui.com/react/overview/releases |
| Radix (`radix-ui`) | 1.8.1 | 2026-10-12 | https://registry.npmjs.org/radix-ui |
| React Aria Components | 1.22.1 | 2026-10-09 | https://registry.npmjs.org/react-aria-components |
| Ark UI (`@ark-ui/react`) | 5.39.3 | 2026-10-05 | https://registry.npmjs.org/@ark-ui/react |
| Mantine | 9.7.3 | 2026-10-17 | https://github.com/mantinedev/mantine/releases |
| DTCG design tokens format | 2025.10 (stable) | 2025-10-28 | https://www.designtokens.org/technical-reports/ |
| Terrazzo CLI / Style Dictionary | 2.10.0 / 5.7.0 | 2026-10-21 / 2026-10-16 | https://registry.npmjs.org/@terrazzo/cli , https://github.com/style-dictionary/style-dictionary/releases |
| TanStack Table / Query | 9.3.1 / 5.104.5 | 2026-10-12 / 2026-10-19 | https://registry.npmjs.org/@tanstack/react-table , https://www.npmjs.com/package/@tanstack/react-query |
| React Hook Form / TanStack Form / Zod | 7.90.0 / 1.33.5 / 4.6.7 | 2026-10 | https://github.com/react-hook-form/react-hook-form/releases , https://registry.npmjs.org/@tanstack/react-form , https://github.com/colinhacks/zod/releases |
| Lucide / Recharts | 1.61.0 / 3.10.1 | 2026-10-21 / 2026-07-25 | https://github.com/lucide-icons/lucide/releases , https://registry.npmjs.org/recharts |
| Vitest | 5.0.5 | 2026-10-15 | https://github.com/vitest-dev/vitest/releases |
| Playwright | 1.64.1 | 2026-10-14 | https://github.com/microsoft/playwright/releases |
| Storybook | 10.6.3 | 2026-10-20 | https://github.com/storybookjs/storybook/releases |
| MSW / openapi-msw | 3.0.3 / 2.1.0 | 2026-10-10 / 2026-09-30 | https://github.com/mswjs/msw/releases , https://www.npmjs.com/package/openapi-msw |
| axe-core | 4.14.0 | 2026-10-05 | https://github.com/dequelabs/axe-core/releases |
| OpenTelemetry JS | 3.0.0 | 2026-10-20 | https://github.com/open-telemetry/opentelemetry-js/releases/tag/v3.0.0 |
| Grafana Faro Web SDK | 2.13.0 (npm `latest` is 2.12.1) | 2026-10-22 | https://github.com/grafana/faro-web-sdk/releases |
| `grafana/otel-lgtm` | 0.37.0 (the version in `compose.yaml`) | 2026-10-16 | https://github.com/grafana/docker-otel-lgtm/releases |

## Primary sources

- CDS 1.36.0 (Consumer Experience, Security Profile): https://consumerdatastandardsaustralia.github.io/standards/
- DSB CX Guidelines, open source assets and CX Checklist: https://cx.dsb.gov.au/overview/how-to-use-these-guidelines , https://cx.dsb.gov.au/overview/open-source-assets , https://cx.dsb.gov.au/overview/cx-checklist
- CDR Rules, Compilation No. 10: https://www.legislation.gov.au/F2020L00094/latest/versions
- WCAG 2.2: https://www.w3.org/TR/WCAG22/ , https://www.w3.org/WAI/standards-guidelines/wcag/new-in-22/
- AHRC Guidelines on equal access to digital goods and services (April 2025): https://humanrights.gov.au/our-work/disability-rights/publications/guidelines-equal-access-digital-goods-and-services
- RFC 10017 OAuth 2.0 for Browser-Based Applications: https://www.rfc-editor.org/rfc/rfc10017
- RFC 9700 OAuth 2.0 Security Best Current Practice: https://www.rfc-editor.org/rfc/rfc9700.html
- OWASP ASVS 5.0, V3 Web Frontend Security: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
- Go `net/http` `CrossOriginProtection`: https://pkg.go.dev/net/http#CrossOriginProtection
- Go modules reference, `ignore` directive: https://go.dev/ref/mod#go-mod-file-ignore
- pnpm settings: https://pnpm.io/settings/dependency-resolution , https://pnpm.io/settings/build
- Australian Government Style Manual: https://www.stylemanual.gov.au/

---

## 1. The current state of the repository

- The data holder authorisation pages (login, one-time password, account selection, consent) are server-rendered Go `html/template` pages in `services/dataholder/internal/adapter/consumerpages/`. They run no JavaScript. They send `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'`, `Referrer-Policy: no-referrer`, `Cache-Control: no-store` and `X-Content-Type-Options: nosniff`, and they check a cookie and a form token (`consumerpages.go`).
- Each operator interface is a JSON API on a separate plain-HTTP listener. It accepts one static bearer token from the configuration (`pkg/platform/operatorhttp`, `OPERATOR_TOKEN` in `compose.yaml`). It does not permit cross-origin requests ([the operator interface library sends no CORS headers](../../pkg/platform/operatorhttp/DESIGN.md#opi-0007)). Errors are RFC 9457 problem details.
- The operator interfaces have no OpenAPI document. The conformance suite already validates CDS responses against the pinned CDS OpenAPI specifications with `getkin/kin-openapi` (`conformance/internal/openapi`).
- `renovate.json` already extends `security:minimumReleaseAgeNpm` and `:maintainLockFilesWeekly`. CI saves caches only on a push to `main` (`.github/workflows/ci.yml`).
- The Go services use the W3C Trace Context propagator and `otelhttp`, and they redact URL paths (`pkg/platform/telemetry`).

## 2. Rendering architectures

| Architecture | What it is | Runtime in production |
|---|---|---|
| Client single-page application (SPA) | A static build of HTML, JS and CSS. The browser renders the UI and calls JSON APIs. | None. Node is a build tool only. |
| Server-side rendering (SSR) with a JS meta-framework | Next.js, React Router framework mode, TanStack Start, SvelteKit. A JS server renders HTML and runs server functions. | A Node server for each UI. |
| Server-rendered hypermedia from Go | Go renders HTML (`html/template` or templ). htmx or Datastar attributes add partial page updates. | None beyond Go. |

Facts:

- React Router "SPA mode" (`ssr: false`) makes no server build and prerenders only `index.html` (https://reactrouter.com/how-to/spa). TanStack Start has a similar SPA mode (https://tanstack.com/start/latest/docs/framework/react/guide/spa-mode).
- Next.js `output: 'export'` makes static files, but it removes cookies, rewrites, redirects, headers, middleware, Server Actions and dynamic routes without `generateStaticParams` (https://nextjs.org/docs/app/guides/static-exports).
- The server half of JS meta-frameworks had many critical advisories in 2025 and 2026:
  - CVE-2025-29927, a Next.js middleware authorisation bypass, CVSS 9.1 (https://github.com/vercel/next.js/security/advisories/GHSA-f82v-jwr5-mffw).
  - CVE-2025-55182 ("React2Shell"), an unauthenticated remote code execution in React Server Components (RSC), 2025-12-03 (https://github.com/facebook/react/security/advisories/GHSA-fv66-9v8q-g76r). A second critical RSC remote code execution was fixed in React 19.3.1 on 2026-10-14. That advisory says: "If your app's React code does not use a server, your app is not affected." (https://github.com/facebook/react/security/advisories/GHSA-xqc2-xf4c-x3h3)
  - Next.js has 44 repository advisories with a 2026 publication date (5 critical), which include more middleware bypasses and remote code executions (`gh api repos/vercel/next.js/security-advisories`).
  - React Router had 10 advisories in June and July 2026, and TanStack Start had a critical reflected XSS in server-function responses on 2026-09-30 (https://github.com/remix-run/react-router/security/advisories , https://github.com/TanStack/router/security/advisories/GHSA-qx66-fv34-fjm8).
- htmx 4.0.0 (2026-08-28) moved to `fetch()` and made attribute inheritance explicit (https://github.com/bigskysoftware/htmx/releases/tag/v4.0.0).
- Gitea, a large Go project, removed htmx in April 2026. The reasons in the proposal: htmx sends requests from `hx-*` attributes and does not support a CSP nonce, so injected markup with `hx-*` attributes is a risk even under a strict CSP; implicit inheritance made behaviour unclear; the code is difficult to lint (https://github.com/go-gitea/gitea/issues/35059 , https://github.com/go-gitea/gitea/pull/37224).
- Datastar 1.0 evaluates expressions with the `Function()` constructor by default, so its CSP needs `unsafe-eval`. Its "CSP Mode" removes this, but "CSP mode does not make Datastar expressions safe to use with untrusted content" (https://data-star.dev/reference/security). Some features are in the paid "Datastar Pro", and its licence does not permit Pro code in a public repository (https://data-star.dev/pro).
- JUDGEMENT: I found no accessible headless component library for htmx or Datastar that is equivalent to React Aria or Radix. Data-dense consoles (sortable tables, comboboxes, date pickers, dialogs) then need Web Components or a second client library.

## 3. Frameworks

| Framework | Status | Weekly npm downloads | Notes |
|---|---|---|---|
| React 19.3 | Stable. React Compiler 1.0 is stable. | 116 M | Governance moved to the React Foundation (Linux Foundation) on 2026-02-24 (https://react.dev/blog/2026/02/24/the-react-foundation). Create React App is deprecated. |
| Vue 3.5 | Stable. Vapor Mode is only in 3.6 RC. | 10.2 M | |
| Svelte 5 / SvelteKit 3 | Stable. SvelteKit 3.0 on 2026-10-01. | 4.2 M / 2.0 M | |
| Angular 22 | Stable. Angular Aria is stable in v22. | 3.4 M | Needs TypeScript 6 (UNVERIFIED: from secondary summaries, https://blog.ninja-squad.com/2026/06/03/what-is-new-angular-22.0). |
| Solid 1.9 | Stable. 2.0 at RC. | 4.6 M | Highest satisfaction in State of JS 2025, about 10 % use (https://2025.stateofjs.com/en-US/libraries/front-end-frameworks/). |
| htmx | 4.0 new; 2.x is npm `latest` | 0.17 M | |

Routers for a React SPA:

- TanStack Router: type-safe routes and search parameters. One maintainer group also makes TanStack Query, Table and Form.
- React Router 8 (2026-06-17): declarative, data and framework modes. Shopify owns Remix; React Router has open governance since 2025-06-05 (https://remix.run/blog).
- Supply-chain incident: on 2026-05-11, attackers published 84 malicious versions of 42 `@tanstack/*` packages through the legitimate GitHub OIDC trusted publisher, after a `pull_request_target` and CI cache poisoning chain (https://github.com/TanStack/router/security/advisories/GHSA-g7cv-rxg3-hmpx). Section 6 gives the controls.

Accessible component libraries by framework (npm, 2026-10-22):

| Framework | Libraries |
|---|---|
| React | React Aria Components (Adobe), Base UI (MUI, from the authors of Radix and Floating UI), Radix, shadcn/ui, Mantine, Fluent UI, Carbon |
| Vue | Reka UI (Radix port), PrimeVue, Ark UI |
| Svelte | Bits UI, Ark UI |
| Angular | Angular CDK, Angular Aria, Angular Material |
| Solid | Kobalte (pre-1.0), Ark UI |

## 4. A Go back end with a JS front end

### 4.1 How Go projects serve a front end

| Project | UI stack | How it ships | Source |
|---|---|---|---|
| Prometheus | React 19.3, Mantine 9, React Router 7, Vite 8 | Embedded in the Go binary (`web/ui/assets_embed.go`) | https://github.com/prometheus/prometheus/tree/main/web/ui |
| Grafana | React 19.2, own `@grafana/ui` | Go server serves the built assets | https://github.com/grafana/grafana |
| Jaeger UI | React 19.3, React Router 7, Vite 8 | Embedded (UNVERIFIED detail) | https://github.com/jaegertracing/jaeger-ui |
| Argo CD | React 19.3 | Served by the Go API server (UNVERIFIED detail) | https://github.com/argoproj/argo-cd/tree/master/ui |
| PocketBase | Plain JS, Vite | `//go:embed all:dist` with a `!no_ui` build tag | https://github.com/pocketbase/pocketbase/blob/master/ui/embed.go |
| Gitea | Vue 3.5, Vite 8, Go templates | Go pages with Vue parts | https://github.com/go-gitea/gitea |
| Temporal UI | SvelteKit, Svelte 5 | Static build (UNVERIFIED embed detail) | https://github.com/temporalio/ui |
| Zitadel | Angular console; login in Next.js 16 | The login is a separate Node app (UNVERIFIED packaging detail) | https://github.com/zitadel/zitadel |

Almost all of these projects use a client SPA, built with Vite and served by the Go binary. None of the infrastructure projects runs a Node server in production.

### 4.2 Serving options

Measured compressed image sizes (linux/amd64, 2026-10-22):

| Option | Image | Size |
|---|---|---|
| Embed the build in the Go service (`embed.FS`) | `gcr.io/distroless/static-debian13:nonroot` (the current base) | 0.82 MB |
| Separate static server | `cgr.dev/chainguard/nginx`, `caddy:2-alpine`, `nginxinc/nginx-unprivileged` | 8.1 to 57.9 MB |
| Node runtime (SSR only) | `gcr.io/distroless/nodejs24-debian13`, `node:24-alpine` | 52.9 to 59.6 MB |

- An embedded build is on the same origin as its API, so it needs no CORS. The Dockerfile needs a Node build stage before the Go build. `go build` fails if the `dist` directory does not exist, unless a build tag or a placeholder file handles it (PocketBase and Prometheus use build tags).
- A `go:embed` pattern cannot contain `..`, so a package can embed only files in or below its own directory (UNVERIFIED: https://pkg.go.dev/embed).

### 4.3 Backend for Frontend (BFF)

- RFC 10017 (BCP 212, August 2026) gives three architectures "in decreasing order of security": BFF, token-mediating backend and browser-based OAuth client. The BFF "is strongly recommended for business applications, sensitive applications, and applications that handle personal data". The browser-based client is "not recommended" for these (https://www.rfc-editor.org/rfc/rfc10017#section-6).
- A BFF keeps the tokens on the server and gives the browser a cookie session. The cookie MUST be `Secure` and `HttpOnly`, SHOULD be `SameSite=Strict` and `Path=/`, SHOULD NOT have a `Domain`, and SHOULD use the `__Host-Http-` prefix (https://www.rfc-editor.org/rfc/rfc10017#section-6.1.3.2). "The BFF MUST implement a proper CSRF defense" (section 6.1.3.3).
- Section 7.1: an application on one domain with no OAuth can use server-side cookie sessions.
- For cdr-lab:
  - A data recipient app is the OAuth client of each data holder, with mTLS and `private_key_jwt`. That client must be confidential and on a server, so the BFF is necessary for data recipient apps.
  - An operator console must never put the operator token in browser JS or browser storage. A BFF keeps the token and the browser gets a session cookie.

## 5. Toolchain and repository layout

### 5.1 Package manager

pnpm has the most supply-chain controls (https://pnpm.io/settings/dependency-resolution , https://pnpm.io/settings/build , release notes for v10.0.0, v11.0.0 and v12.0.0 at https://github.com/pnpm/pnpm/releases):

| pnpm setting | Default in v11 and v12 | Effect |
|---|---|---|
| Dependency lifecycle scripts | Blocked (since v10) | `allowBuilds` (v11 replaces `onlyBuiltDependencies`) lists the packages that can run build scripts. |
| `strictDepBuilds` | `true` (since v11) | Install fails if a dependency has a build script that nobody approved. |
| `binding.gyp` | Needs approval | pnpm treats a package with a `binding.gyp` as a build. |
| `minimumReleaseAge` | 1440 minutes (since v11) | pnpm does not install a version that is newer than this age. |
| `blockExoticSubdeps` | `true` (since v11) | Only direct dependencies can use git or tarball URLs. |
| `trustPolicy` | `off` | `no-downgrade` fails when a package's trust level drops, for example from a trusted publisher to no provenance. |
| `trustLockfile` | `false` | Install applies the age and trust checks to lockfile entries again. |
| Unknown settings | Error (since v12) | A misspelt setting no longer fails silently. |

- npm 12 (2026-07-08) also blocks dependency scripts and git dependencies by default (https://github.com/npm/cli/releases/tag/v12.0.0). But Node 26.12.1 and 24.22.1 still bundle npm 11, so these defaults do not apply unless you install npm 12 separately (https://nodejs.org/dist/index.json). npm has no `trustPolicy` equivalent (UNVERIFIED).
- Bun and Deno block scripts by default. Deno 2.9 has a 24-hour minimum age by default (https://docs.deno.com/runtime/fundamentals/node/). They are separate runtimes.
- Node 25 stopped the distribution of Corepack (https://github.com/nodejs/node/blob/main/doc/changelogs/CHANGELOG_V25.md). mise pins Node (it checks `SHASUMS256.txt` and its OpenPGP signature by default) and pnpm (aqua backend) (https://mise.jdx.dev/lang/node.html , https://github.com/jdx/mise/blob/main/registry/pnpm.toml).
- Go projects with a TS front end that use pnpm: Gitea, Teleport, Nomad, Zitadel. Grafana uses Yarn 4 and Nx.

### 5.2 TypeScript 7, Vite 8, lint and format

- TypeScript 7.0 (2026-07-08) is the native Go port. Full builds are 8 to 12 times faster. New defaults: `strict: true`, `module: esnext`, `types: []`. These options now give an error: `baseUrl`, `moduleResolution: node10`, `target: es5` (https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/).
- TypeScript 7.0 has no programmatic API. Tools that use the compiler API must use `@typescript/typescript6` through an npm alias until 7.1. Vue, Svelte and Astro editor support still needs TypeScript 6 (same source). React TSX has full support.
  - `typescript-eslint` 8.72.0 accepts only `typescript <6.1.0` (https://www.npmjs.com/package/typescript-eslint).
  - `openapi-typescript` 7.13.0 fails with TypeScript 7 (https://github.com/openapi-ts/openapi-typescript/issues/2841). `@hey-api/openapi-ts` also fails (https://github.com/hey-api/hey-api/issues/4235). orval has no `typescript` dependency.
- Vite 8 (2026-03-12) uses Rolldown. Measured: `rolldown`, `oxlint`, `@biomejs/biome` and `typescript@7` have no install scripts, so a Vite 8 and React app can install with an empty `allowBuilds`. `esbuild` still has a `postinstall` script, and Vite 8 has it only as an optional peer.
- Lint and format:

| Option | Facts |
|---|---|
| oxlint + `oxlint-tsgolint` + oxfmt | Type-aware lint with the TypeScript 7 checker: 59 of 61 typescript-eslint type-aware rules, for example `no-floating-promises` (https://oxc.rs/docs/guide/usage/linter/type-aware.html). oxfmt is Prettier-compatible but 0.x (https://oxc.rs/docs/guide/usage/formatter.html). Teleport uses both. |
| Biome 2.6 | Lint and format in one binary with no dependencies. Its type-aware rules use its own inference and are in the "nursery" group (https://biomejs.dev/linter/rules/no-floating-promises/). |
| ESLint 10 + typescript-eslint + Prettier | Largest plugin ecosystem. typescript-eslint does not accept TypeScript 7 yet. ESLint has 30 direct dependencies. |

### 5.3 Go and `node_modules`

- The `./...` pattern does not exclude `node_modules` (`go help packages`). Measured with Go 1.27.2: a `.go` file in `node_modules/evil/` makes `go list ./...` include that package and fail on its import. With `ignore ./node_modules` in `go.mod`, `go list ./...` is clean.
- The `go.mod` `ignore` directive is new in Go 1.25 (https://go.dev/doc/go1.25 , https://go.dev/ref/mod#go-mod-file-ignore). Gitea has `ignore ./node_modules`, and Teleport has `ignore ./web` (https://github.com/go-gitea/gitea/blob/main/go.mod , https://github.com/gravitational/teleport/blob/master/go.mod).

### 5.4 Layout options

| Layout | Examples | For | Against |
|---|---|---|---|
| One JS workspace in `web/` (`web/apps/<ui>`, `web/packages/<lib>`), `ignore ./web` in `go.mod` | Teleport (`web/packages`) | All JS in one directory. The Go tree and root stay clean. | The UI is not next to the service that serves it. `go:embed` needs the build output copied into the Go package. |
| Workspace at the repository root | Gitea, Nomad, Zitadel | Simple commands. | JS and Go configuration mix in the root. |
| `services/<name>/web/` for each service | — | The UI is next to its API, and `go:embed` is direct. | The shared design system is far from its users. The workspace spans `services/`. |

JUDGEMENT: Turborepo or Nx is not necessary for a few apps. `pnpm -r`, `pnpm --filter` and mise tasks are sufficient. Nx was itself the vector of the s1ngularity attack (section 6).

### 5.5 Typed API clients from OpenAPI

| Tool | Model | Status |
|---|---|---|
| openapi-typescript + openapi-fetch | Types only and a small fetch wrapper | 6 and 1 direct dependencies. Needs the TypeScript 6 alias for generation now. |
| @hey-api/openapi-ts | SDK, types and plugins | 0.x. Needs the TypeScript 6 alias now. |
| orval 8 | Clients for fetch, TanStack Query, Zod and MSW mocks | Very active. 25 direct dependencies, and it depends on esbuild. |

`openapi-msw` gives MSW handlers types from the openapi-typescript `paths` type (https://github.com/christoph-fricke/openapi-msw). On the Go side, the options are a hand-written specification (validated with kin-openapi, as the conformance suite does now), spec-first generators (oapi-codegen v2.8.0, ogen v1.24.0) or code-first generation (huma v2.39.1).

## 6. npm supply chain

| Date | Incident | How the code ran | Controls that stop it |
|---|---|---|---|
| 2025-08-26 | Nx "s1ngularity" (https://github.com/advisories/GHSA-cxm3-wv7p-598c) | `postinstall` | Script blocking, release age |
| 2025-09-08 | `debug`, `chalk` and about 18 more; a browser crypto-wallet hijacker (https://github.com/advisories/GHSA-4x49-vf9v-38px) | In the package code, active in browser bundles | Release age only |
| 2025-09-15 | Shai-Hulud worm, more than 500 packages (https://www.cisa.gov/news-events/alerts/2025/09/23/widespread-supply-chain-compromise-impacting-npm-ecosystem) | `postinstall` | Script blocking, release age |
| 2025-11-21 | Shai-Hulud 2.0, about 700 packages (https://www.microsoft.com/en-us/security/blog/2025/12/09/shai-hulud-2-0-guidance-for-detecting-investigating-and-defending-against-the-supply-chain-attack/) | `preinstall` | Script blocking, release age |
| 2026-05-11 | TanStack, 42 packages, valid provenance (https://github.com/advisories/GHSA-g7cv-rxg3-hmpx) | A `github:` git subdependency | `blockExoticSubdeps`, script blocking, release age. Provenance does not help. |
| 2026-06-01 | "Miasma", `@redhat-cloud-services`, valid SLSA provenance (https://labs.cloudsecurityalliance.org/research/csa-research-note-miasma-npm-supply-chain-redhat-20260603-cs/) | Install scripts | Script blocking, release age |
| 2026-06-03 | `binding.gyp` worm, 57 packages (https://snyk.io/blog/node-gyp-supply-chain-compromise-self-propagating-npm-worm-binding-gyp/) | Implicit `node-gyp rebuild` | pnpm `binding.gyp` approval, release age |
| 2026-08-04 | "ChainDrop", keyv, flat-cache and more than 400 packages (https://www.microsoft.com/en-us/security/blog/2026/08/04/chaindrop-supply-chain-compromise-anatomy-self-propagating-worm/) | `preinstall` | Script blocking, release age |

Lessons:

1. Script blocking and a release-age delay together stop all of these incidents except a payload that stays dormant longer than the delay. For `debug` and `chalk`, only the delay helps.
2. Provenance is not a trust signal against these attacks. TanStack, Miasma and ChainDrop had valid provenance.
3. Git and tarball subdependencies are an attack path (TanStack).
4. CI cache poisoning is real (TanStack). cdr-lab already saves caches only on a push to `main`.
5. The worms wrote `.claude/`, `.vscode/` and `.cursor/` files into repositories (TanStack, ChainDrop). This repository has a `.claude/` directory, so review changes to these paths.

Renovate does not apply `minimumReleaseAge` to lock file maintenance. The preset sets it to `null` there, and tells you to make the package manager do its own check (https://github.com/renovatebot/renovate/blob/main/lib/config/presets/internal/security.preset.ts). Thus the pnpm `minimumReleaseAge` is the only control for transitive updates. `security:minimumReleaseAgeNpm` is 3 days.

Scanners: `pnpm audit` and `pnpm audit signatures` (https://pnpm.io/cli/audit). OSV-Scanner v2 reads `pnpm-lock.yaml` and `go.mod`, and OSV.dev imports the OpenSSF malicious-packages data (`MAL-` IDs), so one scanner covers vulnerabilities and known malware in Go and npm (https://google.github.io/osv-scanner/supported-languages-and-lockfiles/ , https://github.com/google/osv.dev/blob/master/source.yaml).

## 7. Design systems and components

### 7.1 Headless (unstyled) component libraries

A headless library gives behaviour and accessibility (focus, keyboard, ARIA) with no visual style.

| | Radix | Base UI | React Aria | Ark UI | Headless UI |
|---|---|---|---|---|---|
| Owner | WorkOS | MUI | Adobe | Chakra team | Tailwind Labs |
| Licence | MIT | MIT | Apache-2.0 | MIT | MIT |
| Stable | Yes | 1.0 on 2025-12-11 | 1.0 on 2023-12-20 | Yes | Yes |
| Commits from 2026-04-22 to 2026-10-22 | 100 or more | 100 or more | 100 or more | 100 or more | 0 |
| Accessibility claim | ARIA patterns | ARIA Authoring Practices and WCAG 2.2 (https://base-ui.com/) | Tested with screen readers on devices (https://react-aria.adobe.com/) | WAI-ARIA (https://ark-ui.com/) | — |
| Internationalisation | Minimal (UNVERIFIED) | Minimal (UNVERIFIED) | 30 or more languages, 13 calendar systems, right-to-left | Some (UNVERIFIED) | Minimal |
| Frameworks | React | React | React | React, Solid, Vue, Svelte | React, Vue |

- Radix is still active in 2026. For example, it fixed iOS Safari overlay taps on 2026-10-14 (https://github.com/radix-ui/primitives/commits). The original Radix authors now make Base UI ("From the creators of Radix, Floating UI, and Material UI", https://base-ui.com/).
- Base UI 1.6 to 1.9 improved mobile drawers, swipe gestures and the virtual keyboard (https://base-ui.com/react/overview/releases).

### 7.2 Styled kits

- **shadcn/ui** copies the component source into your repository. The CLI v4 (March 2026) adds presets, `registry:base` (one install of a full design system), `--diff` and monorepo initialisation (https://ui.shadcn.com/docs/changelog/2026-03-cli-v4). Since July 2026, Base UI is the default base: "Radix is not being deprecated", "every update and new component will ship for both libraries" (https://ui.shadcn.com/docs/changelog/2026-07-base-ui-default). React Aria is also a base since July 2026 (https://ui.shadcn.com/docs/changelog/2026-07-react-aria). Licence MIT.
- **Mantine 9** (2026-03-31) has more than 100 components (dates, charts on Recharts 3, notifications) and uses CSS Modules, with no runtime CSS-in-JS (https://mantine.dev/changelog/9-0-0/ , https://mantine.dev/styles/css-modules/). Prometheus uses it.
- **MUI Material 9**: CSS theme variables. Joy UI: "development is currently on hold" (https://v7.mui.com/joy-ui/getting-started/).
- **Chakra UI v3**: on Ark UI, Panda-style tokens (https://chakra-ui.com/docs/theming/overview). Park UI has had no commits for 6 months.
- **Fluent 2** (Microsoft) and **Carbon** (IBM): strong for enterprise consoles. JUDGEMENT: their strong vendor look makes several distinct bank brands difficult.
- **Untitled UI React**: React Aria and Tailwind, MIT open-source part, a paid PRO part (https://github.com/untitleduico/react).
- **Tailwind Labs joined Shopify on 2026-09-09.** "Everything will always be MIT-licensed". Tailwind Plus is closed to new customers, so the Catalyst kit is not available (https://tailwindcss.com/blog/tailwind-is-joining-shopify).
- **Australian government design systems:** the GOLD Design System (previously the Australian Government Design System) shows no update after 2021 (https://designsystemau.org/). AgDS, the Department of Agriculture system, is active, React and MIT, but it has a government brand (https://github.com/steelthreads/agds-next). I found no current whole-of-government design system (UNVERIFIED: an absence).

### 7.3 Supporting libraries

- **Data tables:** TanStack Table v9 (2026-08-04) is headless, with client-side and server-side modes (https://tanstack.com/table/latest). AG Grid Community is MIT, Enterprise is commercial.
- **Forms:** React Hook Form 7 with Zod 4 or a Standard Schema resolver (https://github.com/react-hook-form/resolvers). TanStack Form 1 supports Standard Schema natively (https://tanstack.com/form/latest/docs/overview).
- **Icons:** Lucide 1.x, ISC licence, the shadcn default.
- **Fonts:** Inter 4.1 (OFL-1.1). JUDGEMENT: self-host fonts, so that the CSP has no third-party font host.
- **Charts:** Recharts 3 turns on its accessibility layer by default (https://github.com/recharts/recharts/blob/main/src/state/rootPropsSlice.ts).
- **Dates:** `Temporal` works in Chrome 144+ and Firefox 139+ but not in released Safari, so iOS needs a polyfill (`temporal-polyfill` 1.0.6) (https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Temporal , https://caniuse.com/temporal).
- **Workbench:** Storybook 10 (ESM only) runs stories as Vitest tests and has an axe-based accessibility addon (https://storybook.js.org/docs/writing-tests). Ladle has low activity.
- **Design tools:** Figma Starter is free (https://www.figma.com/pricing/). Penpot 2.18 is MPL-2.0 and self-hosted, with design tokens. Its token export is reported to use the Tokens Studio format, not DTCG 2025.10 (community report: https://community.penpot.app/t/token-export-is-tokens-studio-format-not-dtcg-2025-10/10544).

## 8. Styling and design tokens

### 8.1 Styling

- Tailwind CSS v4 has CSS-first configuration with `@theme`, so theme values become CSS custom properties. v4.3 (2026-05-08) adds `@container-size` and compound variants (https://tailwindcss.com/blog/tailwindcss-v4-3). The output is static CSS, so it works the same in React pages and in Go `html/template` pages.
- CSS Modules: plain CSS with local class names. Vite supports them with no plugin (UNVERIFIED).
- Build-time CSS-in-TS: vanilla-extract 1.21, Panda CSS 2.4, StyleX 0.19.
- Modern CSS that all current browsers support (MDN Baseline): `@layer` (2022), `@container` (2023), `oklch()` (2023), `:has()` (2023), `light-dark()` (2024, "newly available") (https://developer.mozilla.org/en-US/docs/Web/CSS/@layer , https://developer.mozilla.org/en-US/docs/Web/CSS/color_value/light-dark).

### 8.2 Design tokens

- The W3C Design Tokens Community Group published the first stable format, 2025.10, on 2025-10-28. It has three modules: Format, Color and Resolver. It is a "Final Community Group Report", not a W3C Standard (https://www.designtokens.org/technical-reports/ , https://www.designtokens.org/TR/2025.10/resolver/).
- The Resolver module has `sets` (token sources), `modifiers` with `contexts` (for example `theme: light | dark`) and a `resolutionOrder`. The specification recommends modifiers that are independent of each other. A `brand` modifier is a natural model for several banks.
- Terrazzo 2.0 has "Full support for the DTCG v2025.10 spec, including resolvers" (https://github.com/terrazzoapp/terrazzo/blob/main/packages/cli/CHANGELOG.md). Style Dictionary v5 does not have full support yet (https://styledictionary.com/info/dtcg/).
- Usual token tiers (industry practice, UNVERIFIED as a standard):
  1. Primitive: raw values, for example `color.blue.600`, `space.4`.
  2. Semantic: roles that components use, for example `color.bg.surface`, `color.action.primary.bg`, `color.status.danger.fg`. Each brand and theme maps the roles to primitives.
  3. Component: only where a component needs its own setting, for example `table.row.height.compact`.
- JUDGEMENT: a density setting (compact for consoles, comfortable for consumer UIs) can use the same mechanism.

## 9. CDR consumer experience requirements

### 9.1 Binding and non-binding sources

- The CX standards in the CDS are binding. The DSB CX Guidelines are not binding (https://cx.dsb.gov.au/overview/how-to-use-these-guidelines), but rules 4.10(b) and 4.22(b) say "having regard to any consumer experience guidelines".
- CDS 1.36.0 (4 December 2025) changed no CX standard text. Its only CX-related change is a `brandGroup` field for white-label brands (CDS change log).
- The latest Rules compilation is C10 (4 March 2025). The 2024 Measures No. 1 Rules added the consent bundling rule 4.10(d). I found no 2026 rule change for consent or dashboards (https://www.legislation.gov.au/F2020L00094/latest/versions).
- The DSB published Figma files for 1.36.0 on 18 March 2026, among them: 2AU Redirect to Web with One Time Password, 3AU Authorisation to disclose, 3AU1 Amending authorisations, 5CM1 DH Authorisations (dashboard), 5CM2 DH Withdrawal, 1CO Collection and use consent, 4CM1 to 4CM3 ADR dashboard, withdrawal and receipts. The DSB says these assets are only "partially conformant" to WCAG 2.1 AA (https://cx.dsb.gov.au/overview/open-source-assets).
- The CX Checklist v1.36.0.2026.03.19 is a CSV file. Its columns include Area, Checklist ref, Type (CDR Rule, CX Standard, Technical Standard, CX Guideline), Participant and Requirement level (https://cx.dsb.gov.au/overview/cx-checklist). A filter by Participant and Area gives a list of requirements for each UI.

### 9.2 The CDS accessibility standard

The text, from https://consumerdatastandardsaustralia.github.io/standards/#accessibility-standards:

> At a minimum, all CDR participants MUST seek to comply with the following accessibility guidelines throughout the Consent Model.

The standard then names WCAG 1.4, 2.1, 2.5, 3.1 and 3.3. Measured: the links in the CDS HTML go to the WCAG Understanding pages for **guidelines**, for example "WCAG 2.1" links to `https://www.w3.org/TR/UNDERSTANDING-WCAG20/keyboard-operation.html` (Guideline 2.1 Keyboard Accessible). Thus:

- The CDS does not require WCAG version 2.1 at level AA. It names five WCAG guidelines: 1.4 Distinguishable, 2.1 Keyboard Accessible, 2.5 Input Modalities, 3.1 Readable and 3.3 Input Assistance.
- It gives no conformance level and no WCAG version.
- "MUST seek to comply" is an obligation to try, not a strict conformance requirement.
- The scope is "the Consent Model": the consent, authentication, authorisation, dashboard and withdrawal flows. The operator consoles are not in scope.

### 9.3 Data holder authentication (binding)

From https://consumerdatastandardsaustralia.github.io/standards/#consumer-experience_common-authentication-standards , #consumer-experience_redirect-to-web-with-one-time-password and #one-time-password-credential-requirements:

- Data holders and data recipients MUST tell the consumer that third parties do not need their password to get CDR data.
- Data holders MUST NOT add unwarranted friction (see rule 4.24).
- Data holders MUST NOT include forgotten-details links in the redirect screens.
- Data holders MUST tell the consumer the expiry period of the one-time password (OTP).
- The OTP MUST be numeric, 4 to 6 digits. It MUST be usable only for CDR authentication. It MUST expire.
- The user identifier MUST be one that the customer already knows.
- A memorised secret alone MUST NOT be used for LoA 2 (#credential-requirements), so a password-only login is not permitted.
- CX Guidelines (not binding): let the consumer request a new OTP (2AU.03.19); show the data recipient brand name that matches the Register `brandName` (2AU.00.37, 3AU.00.30).
- Pending changes: DSB Consultation Draft 374 "Minimum Baseline Authentication" is a placeholder with no content (https://github.com/ConsumerDataStandardsAustralia/standards/issues/374).

### 9.4 Data holder authorisation, dashboard and withdrawal (binding)

From the CDS sections #authorisation-standards, #amending-authorisation-standards, #consumer-experience_dashboard-standards_data-holder-dashboards and #withdrawal-standards, and from the Rules:

- The consumer MUST be able to select accounts when the request contains account-specific data and the consumer has more than one account. The data holder MUST show which accounts the data comes from before confirmation. It MUST show visually when an account needs more actions or approvals. It SHOULD show unavailable accounts and why.
- For an amendment with a `cdr_arrangement_id`, the data holder MUST pre-select the eligible accounts that the consumer authorised before, and MUST show added data sets or a changed duration.
- Rule 4.23(1): the data holder must show the name of the accredited person, the historical period of the data, the types of data, single occasion or a period of not more than 12 months, a statement that the consumer can withdraw at any time, and instructions for withdrawal.
- Rule 4.24: the data holder must not add requirements, ask for more information, offer other services or refer to other documents.
- Rule 1.15: the data holder dashboard must show each authorisation and each amendment, allow withdrawal at any time, show a consequences message, be "simple and straightforward to use, and is no more complicated to use than the process for giving the authorisation", and be "prominently displayed and readily accessible".
- The dashboard MUST tell the consumer to check with the data recipient how it handles their data. At withdrawal, the data holder MUST tell the consumer to check the consequences with the data recipient, and MUST give information about redundant data. The CDS gives suggested text for each.
- CX Guidelines (not binding): use "stop sharing" for withdrawal (3AU.02.18), and use the legal entity name from the Register for the accredited person (3AU.02.15).

### 9.5 Data recipient consent, dashboard and notifications (binding)

- Data recipients "MUST notify consumers of redirection prior to authentication" (#consumer-experience_consent-standards).
- Rule 4.11: the consumer actively selects data types, uses, period and recipients. "An accredited person cannot infer consent, or seek to rely on an implied consent." Direct marketing and de-identification consents must not be pre-selected. The request shows the name and accreditation number of the data recipient.
- Rule 4.10(d): a consent request must not be combined with other requests, except another consent under the Rules.
- Rule 1.14: the data recipient dashboard shows each consent and each amendment, allows withdrawal at any time, allows the election to delete redundant data, and is "simple and straightforward to use".
- The withdrawal flow MUST let the consumer elect deletion before the final step, if the data recipient has no general deletion policy (#withdrawal-standards).
- CDR receipts and 90-day notifications MUST be in writing "otherwise than through the consumer dashboard" (#consumer-experience_notification-standards_cdr-receipts). In the lab, this is email to Mailpit.

### 9.6 Data language standards (binding)

Data holders and data recipients MUST use the data language standards for data clusters and permissions when CDR data is requested, reviewed or withdrawn (https://consumerdatastandardsaustralia.github.io/standards/#data-language-standards-common , #banking-and-non-bank-lending-language). For example, the scope `bank:accounts.basic:read` has the cluster "Account name, type and balance" and the permissions "Name of account; Type of account; Account balance;". Business consumers get alternative text. The exact text is easy to check in tests.

### 9.7 Identity of the data recipient

- SSA claims: `org_name` and `client_name` are "to be presented to the end user during authorization". `logo_uri` is required, and "The server SHOULD display this image to the end-user during approval" (CDS Register APIs, SSA).
- The CDS has no explicit anti-phishing standard. Its protections are the redirection notice, the "no password needed" statement, OTP instead of a password, and no forgotten-details links (UNVERIFIED that no other DSB material has one).
- RFC 9700 section 4.16: "Authorization servers MUST prevent clickjacking attacks", with CSP on the authorisation endpoint, login pages and error pages (https://www.rfc-editor.org/rfc/rfc9700.html#section-4.16).
- The DSB studies deceptive patterns in the CDR (Noting Papers 355 and 359, https://github.com/ConsumerDataStandardsAustralia/standards/issues/359). No standard comes from them yet.

## 10. Accessibility

### 10.1 WCAG 2.2

WCAG 2.2 has been a W3C Recommendation since 5 October 2023 (https://www.w3.org/WAI/standards-guidelines/wcag/new-in-22/). The new criteria that apply to the lab UIs:

| Criterion | Level | Effect on cdr-lab |
|---|---|---|
| 2.4.11 Focus Not Obscured (Minimum) | AA | A sticky "Confirm" bar must not hide the focused field. |
| 2.5.7 Dragging Movements | AA | No control that only works with a drag, for example a duration slider. |
| 2.5.8 Target Size (Minimum) | AA | Targets of 24 by 24 CSS px or more, with exceptions. |
| 3.2.6 Consistent Help | A | Help is in the same place on each page of a flow. |
| 3.3.7 Redundant Entry | A | Keep the customer ID after an OTP error. |
| 3.3.8 Accessible Authentication (Minimum) | AA | Applies directly to the customer ID and OTP. |

3.3.8 detail (https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum.html):

- "A service that requires *manual* transcription of a verification code is not compliant." The user must be able "to at least paste the code".
- A split OTP input (one box for each digit) fails when a paste fills only the first box.
- Name, email and phone number are not cognitive function tests. A customer ID is not in this list, so the field must accept autofill and paste (JUDGEMENT).
- WCAG 2.2.1 lets a time-limited authentication code use the "essential" exception, so an OTP expiry is permitted (https://www.w3.org/WAI/WCAG22/Understanding/timing-adjustable.html).

WCAG 3 is "an incomplete draft that will change" (September 2026 Working Draft) and is "not expected to be a completed W3C standard for a few more years" (https://www.w3.org/WAI/standards-guidelines/wcag/wcag3-intro/).

### 10.2 Australian context

- Disability Discrimination Act 1992, section 24, covers goods, services and facilities (https://www.legislation.gov.au/C2004A04426/latest/text).
- The AHRC guidelines (April 2025, not binding, issued under DDA s67(1)(k)): "Organisations should conform with WCAG 2.2 at a minimum Level AA and consider appropriate Level AAA success criteria". They give financial institutions and multi-factor authentication as an example.
- The DTA Digital Service Standard criterion 3 says "latest version of the Web Content Accessibility Guidelines (WCAG)", with no level. It applies to government agencies only (https://www.digital.gov.au/policy/digital-experience/digital-service-standard/criterion-3).
- The European Accessibility Act applies to consumer banking services marketed in the EU from 28 June 2025. It is not Australian law (https://commission.europa.eu/strategy-and-policy/policies/justice-and-fundamental-rights/disability/union-equality-strategy-rights-persons-disabilities-2021-2030/european-accessibility-act_en).

### 10.3 Automated checks

- axe-core 4.14 tags include `wcag2a`, `wcag2aa` and `wcag22aa`. You must give each tag that you want (https://github.com/dequelabs/axe-core/blob/develop/doc/API.md).
- Automated checks find only part of the issues: Deque says axe-core "automatically catches up to 57% of WCAG issues" (as cited at https://storybook.js.org/docs/writing-tests/accessibility-testing). Manual keyboard and screen reader checks are still necessary.

## 11. Mobile-first engineering

- Viewport units: `svh` is stable; `dvh` changes while the user scrolls and "may degrade the interface and hurt performance"; `vh` is the same as `lvh` (https://developer.mozilla.org/en-US/docs/Web/CSS/length).
- Safe areas: `env(safe-area-inset-bottom)`, Baseline since 2020 (https://developer.mozilla.org/en-US/docs/Web/CSS/env). The viewport meta `interactive-widget=resizes-content` shrinks the layout when the virtual keyboard opens (https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/meta/name/viewport).
- Touch targets: WCAG 2.5.8 AA is 24 by 24 CSS px; WCAG 2.5.5 AAA is 44 by 44; Apple's default is 44 by 44 pt (https://developer.apple.com/design/human-interface-guidelines/accessibility); Android is at least 48 by 48 dp (https://developer.android.com/guide/topics/ui/accessibility/apps).
- OTP input: `autocomplete="one-time-code"` (https://html.spec.whatwg.org/multipage/form-control-infrastructure.html) and `inputmode="numeric"`. The WebOTP API is for SMS only and is not Baseline (https://developer.mozilla.org/en-US/docs/Web/API/WebOTP_API). The lab sends OTPs by email, so WebOTP does not apply.
- Core Web Vitals at the 75th percentile: LCP 2.5 s or less, INP 200 ms or less, CLS 0.1 or less. INP replaced FID on 12 March 2024 (https://web.dev/articles/vitals , https://web.dev/blog/inp-cwv-march-12).
- I found no primary source for a performance budget for Australian mobile networks.
- en-AU formats. Measured with Node 24 `Intl`: `Intl.NumberFormat("en-AU", {style: "currency", currency: "AUD"})` gives `-$1,234.50`, and with `currencySign: "accounting"` it gives `($1,234.50)`. `dateStyle: "short"` gives a two-digit year (`22/10/26`).
- The Style Manual says: dates as "31 December 2020", times as "2 pm", and currency with 2 decimal digits (https://www.stylemanual.gov.au/grammar-punctuation-and-conventions/numbers-and-measurements/dates-and-time). It says "Writing to an Australian year 7 level makes content usable for most people" (https://www.stylemanual.gov.au/accessible-and-inclusive-content/literacy-and-access).
- CDS `DateTimeString` values are RFC 3339 with an offset from UTC, and `AmountString` is a decimal string (CDS #common-field-types). Format them for display only.

## 12. Browser security

- OWASP ASVS 5.0 V3 asks for: `Secure`, `HttpOnly` and `__Host-` cookies (3.3.1 to 3.3.4); HSTS, a CORS allowlist, CSP, `nosniff`, `Referrer-Policy`, `frame-ancestors` on each response, and COOP (3.4.1 to 3.4.8); a CSRF defence (3.5.1 to 3.5.3); different host names for different applications (3.5.4); and SRI for external assets (3.6.1).
- Content Security Policy: web.dev says to use a nonce-based CSP for pages rendered on the server, and a hash-based CSP for static pages such as SPAs (https://web.dev/articles/strict-csp). Vite inlines small assets as `data:` URIs by default (`build.assetsInlineLimit`), and the Vite docs say "Do not allow `data:` for script-src" (https://vite.dev/guide/features). `frame-ancestors` is ignored in a `<meta>` element (https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html).
- Trusted Types are now in all three browser engines: Chrome 83, Firefox 148, Safari 26 (https://github.com/mdn/browser-compat-data/blob/main/api/TrustedTypePolicyFactory.json). React 19.3 added Trusted Types integration (https://github.com/facebook/react/releases/tag/v19.3.0).
- Cookies: `__Host-Http-` is supported by Chrome 140 and Firefox 143, but not by Safari (https://github.com/mdn/browser-compat-data/blob/main/http/headers/Set-Cookie.json).
- **A lab-only problem:** "Cookies do not provide isolation by port" (https://www.rfc-editor.org/rfc/rfc6265#section-8.5). `localhost:8181` and `localhost:8182` are also the same site, so `SameSite` does not separate them. RFC 10017 says SameSite "falls short when the BFF is hosted alongside other applications within the same site" (section 6.1.3.3.1). In the lab, all UIs on `localhost` get the cookies of each other.
- Go `http.CrossOriginProtection` (Go 1.25) allows GET, HEAD and OPTIONS. For other methods, it allows `Sec-Fetch-Site: same-origin` and `none`, and refuses all other values, `same-site` included. With no `Sec-Fetch-Site`, it compares `Origin` with `Host`. Measured: the Go 1.27.2 source (`src/net/http/csrf.go`) has this logic. Thus it blocks a POST from `localhost:8181` to `localhost:8182`, which SameSite does not block.
- Browsers send Fetch Metadata headers only to potentially trustworthy URLs, for example HTTPS or localhost (https://w3c.github.io/webappsec-fetch-metadata/).

## 13. Testing

- **Vitest Browser Mode** runs component tests in a real browser through Playwright. It has been stable since Vitest 4 (https://vitest.dev/blog/vitest-4). Vitest 5.0.0 (2026-09-03) makes locators strict by default and fails a test when an async assertion is not awaited. It also moves the WebdriverIO provider out of the main repository (https://github.com/vitest-dev/vitest/releases/tag/v5.0.0).
- jsdom and happy-dom are still supported. JUDGEMENT: they do no layout, focus management or CSS, so they can pass a test that fails in a browser.
- **Testing Library** puts accessible queries first (`getByRole`, `getByLabelText`) and `getByTestId` last (https://testing-library.com/docs/queries/about). Vitest and Playwright locators use the same model.
- **Playwright 1.62** replaced the experimental component testing packages with a "stories and galleries" model in `@playwright/test`. The docs say "It is stable." (https://playwright.dev/docs/test-components).
- **MSW 3** (2026-09-28) is ESM only and also intercepts page navigations and form submissions (https://github.com/mswjs/msw/releases/tag/v3.0.0).
- **Visual regression:** Playwright `toHaveScreenshot` and Vitest `toMatchScreenshot` are free. Chromatic and Argos are hosted services. Lost Pixel is archived.
- **Testing Trophy:** static checks, unit, integration and end-to-end, with the most tests at the integration level (https://kentcdodds.com/blog/the-testing-trophy-and-testing-classifications).
- **Contract tests:** Pact needs a broker or a file exchange and a Go provider verifier (https://www.npmjs.com/package/@pact-foundation/pact). An OpenAPI document that the Go conformance suite validates, and that the TS types come from, gives the same check with tools that the repository already has.

## 14. Browser telemetry

- **OpenTelemetry JS 3.0.0 (2026-10-20)** removed `sdk-trace-base`, `sdk-trace-web` and `sdk-trace-node`, and replaced them with `@opentelemetry/sdk-trace`. `sdk-logs` is now stable. The packages in the contrib and browser repositories "will update to JS SDK 3.x releases soon" (https://github.com/open-telemetry/opentelemetry-js/blob/main/doc/3.x/migration-guide.md).
- The fetch and XHR instrumentations are deprecated. Their replacement is in the experimental `@opentelemetry/browser-instrumentation`, whose "emitted telemetry differs significantly" (https://github.com/open-telemetry/opentelemetry-js/tree/main/experimental/packages/opentelemetry-instrumentation-fetch).
- The browser semantic conventions have the status "Development" (https://opentelemetry.io/docs/specs/semconv/browser/browser-events/).
- **Grafana Faro** sends to Grafana Alloy (`faro.receiver`). Its OTLP transport is experimental (https://github.com/grafana/faro-web-sdk/blob/v2.13.0/experimental/transport-otlp-http/README.md). `grafana/otel-lgtm` runs an upstream OpenTelemetry Collector, not Alloy (https://github.com/grafana/docker-otel-lgtm/tree/v0.37.0/docker).
- **otel-lgtm CORS:** measured, the OTLP HTTP receiver of `otel-lgtm` 0.37.0 allows only the origin `http://*`, and it has no redaction processor (https://github.com/grafana/docker-otel-lgtm/blob/v0.37.0/docker/otelcol-config.yaml). Thus a page that the lab serves over HTTPS cannot send telemetry to it directly.
- **PII risks in browser telemetry for cdr-lab:**
  - The CDR redirect carries `code`, `state` and, with JARM, a `response` JWT in the URL. PAR gives a `request_uri`. Resource paths carry `accountId`. The default URL sanitisers do not remove these. Faro's `page` meta and OpenTelemetry navigation events record the URL.
  - Console capture, error messages and stack traces can contain response values.
  - The OpenTelemetry user-action instrumentation copies each `data-otel-*` attribute.
  - Session replay records the DOM.
  - The browser is not trusted, so a scrub in the browser is not sufficient. A server-side point must also remove data.
- W3C Trace Context Level 2 says that an untrusted caller can abuse trace flags, and that a system should restart traces at its front gates (https://www.w3.org/TR/trace-context-2/). Go `otelhttp.WithPublicEndpointFn` makes the server span link to the incoming context instead of using it as the parent (https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp#WithPublicEndpointFn). The repository does not set it now, and AGENTS.md depends on a client-chosen `traceparent` to find traces.

---

## 15. Options and trade-offs (JUDGEMENT)

This section is my opinion, from the facts above.

### 15.1 Challenges to the brief

1. **"One stack for all UIs" should mean one design system, not one rendering model.** The data holder authorisation pages are a few forms in the most security-critical flow of the lab. Now they run no JS, with CSP `default-src 'none'`. An SPA there adds JS, an API for each step and a supply-chain surface, and the consumer gets nothing from it. Share the tokens and the CSS, not the JS framework.
2. **Do not add a Node server.** Most of the critical JS framework advisories of 2025 and 2026 are in server code. cdr-lab already has a Go server for each service.
3. **The npm supply chain is the largest new risk.** Any JS framework brings it. Section 6 shows the controls.
4. **The operator token must not go to the browser.** A browser console for the current operator API needs a BFF, or the operator interface must get a session login.
5. **Browser telemetry is a PII risk with little value now.** The server spans of the BFF already show each request.

### 15.2 Recommended stack

| Area | Recommendation | Main reason |
|---|---|---|
| Rendering | React SPA for consoles and data recipient apps. Go-rendered pages (`html/template`) for the data holder authorisation pages. | Best component ecosystem. No JS on the authorisation pages. |
| Framework | React 19 with React Compiler, TypeScript 7, Vite 8. No RSC, no server functions, no Next.js. | Largest ecosystem; neutral governance; the pattern of Prometheus, Grafana, Jaeger and Argo CD. |
| Router and data | TanStack Router and TanStack Query, or React Router 8 in data mode. | Typed URLs and search parameters help filtered tables. |
| Serving | Embed each build in the Go service that owns its API (`go:embed`), on the current distroless base. A Node stage in the Dockerfile builds it. | No new image or runtime. Same origin, so no CORS. |
| Auth | The Go service is the BFF. A `__Host-` session cookie (`HttpOnly`, `Secure`, `SameSite=Strict`) and `http.CrossOriginProtection`. | RFC 10017. No token in JS. |
| Package manager | pnpm 12, pinned with Node in `mise.toml`. Explicit `minimumReleaseAge: 4320`, `trustPolicy: no-downgrade`, `blockExoticSubdeps: true`, `strictDepBuilds: true`, empty `allowBuilds`. | Most supply-chain controls. 4320 minutes matches Renovate. |
| CI supply chain | `pnpm install --frozen-lockfile`, `pnpm audit signatures`, OSV-Scanner for Go and npm. No auto-merge of npm updates. | |
| Layout | One pnpm workspace in `web/` (`web/apps/<ui>`, `web/packages/<lib>`), `ignore ./web` in `go.mod`. No Nx or Turborepo. | All JS in one place. The Go tree stays clean. |
| Lint and format | oxlint with type-aware rules, and oxfmt (or the Biome formatter if 0.x is not acceptable). | Type-aware rules such as `no-floating-promises` with the TypeScript 7 checker. |
| API contract | A hand-written `openapi.yaml` for each operator and BFF API, validated in the Go conformance suite. openapi-typescript and openapi-fetch for the client; openapi-msw for test handlers. Commit the generated types and check that a new generation gives no diff. | One contract, checked on both sides, with tools that the repository already has. |
| Components | shadcn/ui on Base UI (or on React Aria), Tailwind CSS v4. | You own the source. Most common stack, so examples and coding agents know it well. |
| Tokens | DTCG 2025.10 files with `brand` and `theme` modifiers, built by Terrazzo to CSS custom properties. A `[data-brand]` block for each mock bank. | One theme mechanism that works in React and in Go templates. |
| Other libraries | TanStack Table v9, React Hook Form and Zod 4 (or TanStack Form), Lucide, Recharts 3, self-hosted Inter, `Intl` for en-AU. | |
| Tests | Vitest 5 Browser Mode with MSW 3 and axe checks; a few Playwright journeys against the lab. No jsdom. Storybook and visual regression later, if necessary. | Real browser behaviour; TDD with role-based locators. |
| Accessibility | WCAG 2.2 AA for all consumer UIs, recorded as a judgement in each `DESIGN.md`. The same components for consoles, with no mobile work for consoles. | The CDS is not testable as written. The AHRC recommends 2.2 AA. |
| Browser telemetry | None in the first milestone. Later: OpenTelemetry browser packages through a server-side redaction point, never directly to otel-lgtm. | ADR 0006 and the PII risks in section 14. |
| Design tool | No Figma at the start. Use shadcn presets and a token set for each brand. Use the DSB Figma files for flow and content, not for visual design. | |

### 15.3 Other options

- **All Go (templ and htmx 4, or Datastar):** one language and no npm. Weak for data-dense consoles, with no accessible component library, and less familiar to coding agents. Choose it if the consoles stay simple CRUD screens.
- **Svelte 5 or Vue 3 SPA:** simpler to read than React, but smaller component ecosystems, and their editor support still needs TypeScript 6.
- **Mantine 9 for consoles:** the fastest path to good consoles, but its look is less suitable for several distinct bank brands. It works against one design system for all UIs.
- **A Node meta-framework (Next.js, React Router framework mode, TanStack Start):** a Node runtime in production, the largest attack surface, and the OAuth state split between Go and Node.

### 15.4 Lab-only shortcuts to record in `docs/production-gaps.md` if the lab takes them

- All UIs share `localhost`, so cookies are not isolated by port. Production needs separate host names (ASVS 3.5.4).
- One shared operator credential with no identity for each operator.
- Plain-HTTP operator listeners, if a UI uses them.

## 16. Open questions

These questions need a decision from the owner. They are not settled by this research.

1. Do the data holder authorisation pages stay server-rendered in Go, with a shared design system in CSS?
2. Router: TanStack Router or React Router 8?
3. Component base: Base UI (the shadcn default) or React Aria (stronger internationalisation, touch and screen reader quality, but a shadcn base for only 3 months)?
4. Forms: React Hook Form or TanStack Form?
5. Where does the operator console session live: on the existing operator listener (a login replaces the static bearer token), or in a separate BFF listener that keeps the token?
6. How does an operator log in to a console in the lab: a lab password, a fixed lab user, or no login (a production gap)?
7. Lint and format: oxlint and oxfmt (0.x), or oxlint and the Biome formatter?
8. API contract: hand-written OpenAPI documents, or OpenAPI generated from Go code? The second option changes the HTTP layer of each service, so it is an architecture decision.
9. Where do the mobile data recipient UIs live? They depend on a data recipient service that does not exist yet.
10. Must the consumer UIs support dark mode, and do the operator consoles need it?
11. Is Storybook necessary as a design-system catalogue in the first milestone?
12. Is a link, not a parent, for incoming browser trace context (`WithPublicEndpointFn`) necessary on consumer listeners, given that the lab tooling depends on a client-chosen `traceparent`?
