# Online bank of a data holder, milestone 1: research

- Research date: 2026-11-02
- Applies to: CDS 1.36.0, CDR Rules Compilation No. 10, WCAG 2.2, RFC 6265 and RFC 6265bis, W3C Secure Contexts
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 1 of the project "A consumer uses the online bank of each data holder": "The branded landing page and the product list of the brand, with the same data as Get Products and Get Product Detail." Milestones 2 to 4 add sign-up and sign-in, accounts, the open and close of accounts, and the consumer dashboard of rule 1.15. This note gives the facts that milestone 1 must not block for them.
- Related notes: [Web front ends](2026-10-22-web-front-ends.md), [Branded data holder instances](2026-11-01-branded-data-holder-instances.md), [Preset bank brands: design](2026-11-01-preset-bank-brand-design.md), [Four preset banks in the lab](2026-11-02-four-preset-banks-in-the-lab.md). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | CDS 1.36.0: Get Products | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-products |
| S2 | CDS 1.36.0: Get Product Detail | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-product-detail |
| S3 | CDS 1.36.0: BankingProductDetailV7 | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductdetailv7 |
| S4 | CDS 1.36.0: Common Field Types (RateString, AmountString, DurationString) | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types |
| S5 | CDS 1.36.0: Product Categories | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories |
| S6 | CDS 1.36.0: Common Authentication Standards (CX), "Authentication: Passwords" | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards |
| S7 | CDS 1.36.0: Restricted Credentials | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#restricted-credentials |
| S8 | CDS 1.36.0: CORS | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors |
| S9 | CDR Rules, rule 1.15 (consumer dashboard) | https://www.legislation.gov.au/F2020L00094/latest/text |
| S10 | RFC 6265, section 8.5 (Weak Confidentiality) | https://www.rfc-editor.org/rfc/rfc6265#section-8.5 |
| S11 | RFC 6265bis draft, section 4.1.3.2 (the `__Host-` prefix) | https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-rfc6265bis#section-4.1.3.2 |
| S12 | W3C Secure Contexts, section 3.1 (potentially trustworthy origins) | https://www.w3.org/TR/secure-contexts/#is-origin-trustworthy |
| S13 | MDN: `Crypto.randomUUID()` (secure context only) | https://developer.mozilla.org/en-US/docs/Web/API/Crypto/randomUUID |
| S14 | OWASP ASVS 5.0, V3 Web Frontend Security (3.5.4: different host names for different applications) | https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md |
| S15 | WCAG 2.2, SC 1.4.10 Reflow | https://www.w3.org/TR/WCAG22/#reflow |
| S16 | RFC 10017, section 6.1 and 7.1 (BFF, cookie sessions) | https://www.rfc-editor.org/rfc/rfc10017#section-6.1 |
| S17 | MDN: `Intl.DurationFormat` | https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/DurationFormat |

## 1. The project and what the repository has now

Facts:

- The project puts in scope a landing page, the product list, sign-up, sign-in "with a customer number and a password field in which any password works", accounts with balances and transactions, the open and close of accounts, and the consumer dashboard. Its constraints: the CDR authentication pages never ask for a password, "the consumer must not confuse the two", the site is lab-only with an entry in `docs/production-gaps.md`, the visual quality matches the design guide, and no customer number or customer name goes into telemetry. Its first acceptance criterion: "The product list shows each current product of the brand, with the same data as Get Products and Get Product Detail." Its last: "Each page shows the brand of the instance and the footer that says that the bank is a simulation."
- The data holder `DESIGN.md` puts "A consumer dashboard" out of scope (`services/dataholder/DESIGN.md:47`). The console "is not a consumer dashboard" (`:737`).
- The design guide expects a React consumer UI: "The consumer pages of the data holder have the components of the next section. The first React consumer UI adds them to this package as components, with the same look." (`web/packages/ui/DESIGN.md:111`)
- ADR 0014 names "later consumer UIs (mobile), for example a data holder consumer dashboard" in its context (`docs/adr/0014-…md:7`). Its decision is "Each web UI is a React SPA with client rendering only", with one exception: the data holder authorisation pages (`:25`). It also says "The browser gets no credential" (`:31`).
- The authorisation pages are Go `html/template` pages (`services/dataholder/internal/adapter/consumerpages/templates/`). Their CSP is `default-src 'none'; style-src 'self'; font-src 'self'; img-src 'self'; frame-ancestors 'none'; base-uri 'none'` (`consumerpages.go:458-459`). Their flow cookie is `__Host-cdr-flow`, `Path=/`, `Secure`, `HttpOnly`, `SameSite=Strict` (`consumerpages.go:24-26`, `:428-435`).
- The public listener (container port 8081, host 8181 for `wattle-bank`) serves the CDS public APIs, the OpenID Provider configuration, the authorisation endpoint and its pages, the brand files and the consumer pages stylesheet on one mux (`internal/root/api.go:91-96`). It wraps all responses with `Access-Control-Allow-Origin: *`, except the authorisation path and the paths below it (`internal/adapter/httpapi/httpapi.go:65-76`; `api.go:171-174`). Its mux sends each unknown path, also `/`, to a CDS `404 Resource/NotFound` JSON error (`httpapi.go:27`, `:61-63`). It has no `Host` allowlist.
- The console listener uses `pkg/platform/consolehttp`: a SPA from an embedded `dist`, a `/api/` prefix, the CSP of the Register console, `nosniff`, `Referrer-Policy: no-referrer`, COOP, the `Host` allowlist and `http.CrossOriginProtection` (`pkg/platform/consolehttp/consolehttp.go:20-23`, `:73-117`). The listener is off by default (`CONSOLE_ENABLED`, `CONSOLE_ALLOWED_HOSTS`, `:29-34`), and it is plain HTTP (`internal/root/console.go:18-28`).
- The console listener serves its own brand files (`/brand/console.css`, `/brand/logo.svg`, `/brand/mark.svg`) because the CSP permits only the same origin (`internal/adapter/branddir/handler.go:24-31`; `DESIGN.md`, "Listener and security").
- The public TLS certificate of each bank has the SANs `<host>` and `localhost` (`mise.toml:216`).
- The data holder image builds the console in a Node stage and copies `web/apps/dataholder-console/dist/` and `web/packages/ui/dist/consumer-pages/` into the Go embed directories (`services/dataholder/Dockerfile:29-30`).
- `crypto.randomUUID()` is available only in a secure context [S13]. `http://localhost` and `http://127.0.0.1` are potentially trustworthy origins, so they are secure contexts also with no TLS [S12]. In the repository, only the Register console calls it (`web/apps/register-console/src/…/Create*Page.tsx`, `Add*Page.tsx`). The data holder console does not.

## 2. Architecture of the site

Options:

| | A. React SPA (ADR 0014) | B. Go pages with no script | C. Mix: Go pages for M1, SPA for the signed-in area |
|---|---|---|---|
| Agrees with ADRs | Yes. ADR 0014 names the consumer dashboard as a SPA. The design guide expects the first React consumer UI (`ui/DESIGN.md:111`). | No. Needs a new ADR that extends the exception of ADR 0014. | Partly. Two renderers in one site. |
| Separation from the CDR pages (project constraint) | Strong. Different listener, different CSP (scripts on), different layout (navigation, no hero with the data recipient). | Weak. The same renderer and the same look as the CDR pages: a password form in the same template set as the no-password pages. | Mixed. |
| M2 to M4: sign-in, forms, confirmation dialog of a withdrawal, tables of transactions | The design system has Form, Error summary, Confirmation dialog, Data table, Toast (`ui/DESIGN.md:76-111`). React Hook Form and Zod (ADR 0016). | Each component is new CSS in `consumer-pages.css` and new templates. A confirmation is a full page POST and redirect. No component tests. | The cost of A and B. |
| Tests | Vitest browser tests with axe and `openapi-msw` handlers; Playwright journeys against the lab (`test:lab-web`). The same as the consoles. | Go golden files, and the axe checks of `consumer-pages.test.ts` on the golden files for each preset brand. | Both. |
| Supply chain (ADR 0015) | No new dependency: the console already uses React, TanStack Router and Query, `openapi-fetch`. One more app in the workspace. | No npm code. | As A. |
| Works with no JavaScript | No. ADR 0014 accepts this. | Yes. | Landing page only. |
| Cost now | A new app, an OpenAPI document, an embed directory, a Dockerfile `COPY` line, `.gitignore` entries, a Playwright project and a CI step. | New templates and CSS only. | Highest. |

Recommendation: **A, a React SPA in a new app in `web/apps/`**, served by the data holder, with a site API under ADR 0017 (a hand-written OpenAPI document in the data holder directory, `openapi-typescript`, `openapi-fetch`, `openapi-msw`). Reasons: ADR 0014 already decides it, the design guide expects it, and milestones 2 to 4 need the form and dialog components that only the React package has. Option B moves the bank sign-in close to the CDR pages, which the project tells us to keep apart. The cost of A is mostly one-time set-up, which a copy of `web/apps/dataholder-console` and `pkg/platform/consolehttp` gives.

M1 adds the consumer components of `ui/DESIGN.md` "Consumer pages" (header, card, buttons, footer) to `web/packages/ui` as React components, as line 111 requires. The site does not use the hero with the data recipient, the step indicator or the "Secure" label of the CDR pages (see section 7).

## 3. Listener and origin

Facts:

- Cookies "do not provide isolation by port" [S10]. In the lab all ports are on `localhost` (or `127.0.0.1`), so a cookie of one port goes to each other port. A separate port does not isolate the site cookies from the CDR pages; a separate host name does [S14, 3.5.4].
- A `__Host-` cookie must be `Secure`, have no `Domain` and have `Path=/` [S11]. Thus a `__Host-` session cookie on the public origin always goes to `/authorize` too. A cookie path is not a security boundary [S10].
- RFC 10017 permits a server-side cookie session for an application on one domain with no OAuth [S16, section 7.1].
- The public listener is the `publicBaseUri`, `productBaseUri` and issuer of the brand on the Register (`DESIGN.md`, Brand). Its CORS rule gives `*` to each path except the authorisation path (`httpapi.go:65-76`). The CDS requires CORS on the unauthenticated APIs [S8]; it gives no rule for a bank web site.

Options:

| | P. The public listener (8x81, TLS) | N. A new listener for the site |
|---|---|---|
| CDR surface | The site shares the mux, the CORS wrapper and the `/` catch-all of the CDR APIs. Each needs an exception for the site paths. | The CDR listener does not change. |
| Lab-only rule | A lab-only feature on the CDR listener, which a real component uses. | Off by default, as the console. A production data holder has its own site. |
| Security headers | A third CSP on one origin (APIs, CDR pages with no script, SPA with script). | The console pattern: one CSP, `Host` allowlist, `CrossOriginProtection`. |
| Cookies (M2) | The session cookie goes to the CDR pages. | The same in the lab (one host), but not in a deployment with a host name for each listener. |
| Telemetry | Site spans mix with CDR spans; `server.port` does not separate them. | Separate by `server.port` on spans. |
| Ports and README | No change. | One more port for each bank: the block of five becomes six (`DESIGN.md:863`). One README row. |

Recommendation: **N, a new listener for each instance, with TLS and the public certificate of the bank**, and the rules of the console listener (embedded SPA, `/api/`, CSP of `consolehttp`, `Host` allowlist, `CrossOriginProtection`, off by default). Port 8x85: 8185, 8285, 8385, 8485 are free in `compose.yaml`. Reasons for TLS and not plain HTTP: a `__Host-` session cookie in M2 needs `Secure` [S11]; a real online bank is HTTPS; the certificate already has the SAN `localhost`, so the PKI does not change; and a page on a host other than localhost is then still a secure context [S12, S13]. The cost: Playwright Chromium does not read `NODE_EXTRA_CA_CERTS`, so the lab-web project needs the lab root CA or `ignoreHTTPSErrors` (`web/playwright.config.ts` has HTTP projects only now).

The `consolehttp` package is plain HTTP and its names say "console". To reuse it, a ticket can give it a TLS config, or move the shared parts to a more general package. This is an implementation choice, not a requirement.

The lab cannot isolate the site cookies from the CDR pages, because all ports share `localhost`. Use different cookie names. A deployment with a host name for each listener isolates them. The site is lab-only and out of scope for production, so this is not a production gap by the test of `docs/production-gaps.md` ("A function that a project put out of scope is not a gap"). The "any password" sign-in of M2 is the gap that the project names.

## 4. The product list

Facts:

- The data holder holds the required fields of `BankingProductDetailV7` and only `effectiveFrom`, `effectiveTo`, `features`, `fees`, `depositRates` and `lendingRates` (`DESIGN.md:114-115`; `internal/domain/product.go:16-82`). It does not hold `eligibility`, `constraints`, `applicationUri`, `additionalInformation`, `cardArt`, `bundles` or rate tiers [S3].
- `app.ProductReader` gives a page of products for a query (`Effective`, `At`, `UpdatedSince`, `Brand`, `Category`, `Page`, `PageSize`) and one product by ID (`internal/app/products.go:27-47`). Get Products and Get Product Detail call it (`internal/adapter/httpapi/products.go:52-74`, `:120-137`).
- Get Product Detail writes `domain.Product` with its JSON tags (`products.go:135`). Get Products maps the fields of `BankingProductV6` with a struct in `httpapi` (`products.go:15-49`). ADR 0013 forbids one adapter to import another adapter.
- A product is current when the time is at or after `effectiveFrom` and before `effectiveTo` (`DESIGN.md:126`). Get Products gives current products by default [S1].
- Each preset catalogue has a future product and a product whose `effectiveTo` is in the past (`DESIGN.md:129`). Example, `wattle-bank`: 10 current products in 7 categories, "Roam Travel Card" from 2027-03-01, "Blossom Christmas Saver" to 2026-06-30 (`services/dataholder/brands/wattle-bank/products.json`). `redgum-bank` has business products: "Ledger Business Account", three `BUSINESS_LOANS` (`brands/redgum-bank/products.json`).
- `productenums.go` has 9 enumeration types with 97 values (`internal/domain/productenums.go`). A rate is a RateString (a decimal fraction, for example `0.0125` is 1.25 %), an amount is an AmountString, and `applicationFrequency` and many `additionalValue` values are ISO 8601 durations [S4]. The meaning of `additionalValue` depends on the `featureType` or the rate type [S3]: in the presets it is a card scheme ("Visa Debit"), a count ("99"), a duration ("P55D"), a points total ("60000") or text.
- The design guide formats dates, times and amounts with `Intl` and `en-AU`, at display only (`ui/DESIGN.md`, Content). `Temporal` is not in Safari (`ui/DESIGN.md`, Browser support). `Intl.DurationFormat` formats a duration object but does not parse an ISO 8601 string [S17].

Options for the data:

| | Q. The site API calls `app.ProductReader` | H. The site calls the CDS API over HTTP |
|---|---|---|
| Same data | By construction: the same port, the same store, the same `At` and the same current rule. | By construction, with the CDS shapes. |
| Origin | Same origin, no CORS. | A cross-origin call from the browser to 8x81, with `x-v` headers, or a BFF call in the server. |
| Labels and pages | The site API gives the shapes that the UI needs (one list with no pages). | The UI pages through Get Products and maps the CDS shapes. |
| Hexagonal rules | A new driving adapter on an existing port. | An adapter that calls its own service over the network. |

Recommendation: **Q**. The site API gives the products with the CDS field names and values (no new names, no conversion in Go), so that a test can compare them field by field with Get Products and Get Product Detail. The site API gives all current products of the brand in one response (the catalogue is small: a brand file is 1 MiB or less, `DESIGN.md:104`) and the detail of one current product.

Recommendations for the content:

- **List:** each current product, in groups by `productCategory`, in the order of the 13 categories of the CDS [S5], with the name, the description, and a short headline: the highest base deposit rate or the lowest lending rate, with its type, when the product has one. Only categories with a current product show. Future and past products do not show. A realistic bank lists only products that it sells, and the acceptance criterion names current products only.
- **Detail:** name, category, description, `isTailored` ("Rates and fees can be different for you" when true), each deposit rate and lending rate (type, rate as a percentage, comparison rate, application frequency, repayment type, loan purpose, additional value and info), each fee (name, type, amount or rate, frequency, additional info), each feature (type, additional value and info), and `effectiveFrom` or `effectiveTo` when present. A product that is not current, or not known, gives the "not found" page of the site. Eligibility is not held; the page shows nothing for it.
- **Enumeration labels:** the UI gives a plain-language label for each value, in a TypeScript `Record` of each enumeration type, so that the compiler finds a missing label. The OpenAPI document of the site gives each enumeration, as the console document gives `ProductCategory` (`internal/adapter/console/openapi.yaml:1068`). A Go test compares each enumeration of the document with the domain values. The labels come from the CDS description of each value [S3, S5], in plain language at year 7 level (`ui/DESIGN.md`, Content).
- **Values:** the UI formats each RateString as a percentage with `Intl.NumberFormat` (`style: "percent"`), each AmountString as an amount, and each ISO 8601 duration with a small parser and plain words ("each month", "55 days"). The UI formats `additionalValue` by its type for the types that the presets use, and shows the text as it is for the others.

## 5. The landing page

Facts:

- ADR 0023 fixes the files and fields of a brand directory. A new field, for example a tagline, is a change of the contract, and the data holder refuses an unknown field (`docs/adr/0023-…md`, Decision).
- `domain.DemonstrationNotice(bank)` gives the footer text of the CDR pages and the email: "<brand name> is not a real bank, this is just a demonstration. This project has no affiliation with the ACCC, DSB, any CDR participant or any real bank." (`internal/domain/otp.go:113-114`; `DESIGN.md:508`). A brand directory cannot change it.
- "Data holders and data recipients MUST state in consumer-facing interactions and communications that third parties do not need consumer passwords to access CDR data." [S6] The site is a consumer-facing interaction of the data holder.
- The CDR pages have no link in the footer because the CDS forbids links to other documents and services on the authentication pages [S6; `DESIGN.md:508`]. That rule is for the authentication flow, not for a bank web site.

Recommendations:

- **Header:** the logo of the brand on the brand colour (32 px, brand name as the text alternative), and a "Products" link. In M2, a "Sign in" button. No "Secure" label and no hero with the data recipient, so that the site does not look like a CDR page.
- **Hero:** a band in the hero colour with the mark of the brand, an h1 with text from the data holder that names the brand (for example "Banking with <brand name>"), one sentence, and a button to the products. No claims about rates, awards or customers, and no copy from a real bank.
- **Categories:** one card for each category that has a current product, with a Lucide icon, the plain-language category name and the number of products. Each card links to its group in the product list.
- **CDR band:** a short statement that the bank shares data under the Consumer Data Right only with the consent of the consumer, and that third parties do not need the password of the consumer [S6]. In M4, a link to the dashboard.
- **Footer:** the text of `domain.DemonstrationNotice` on each page. The site API gives the brand name and this text, so that the text has one source. The footer can have links to the site pages, but not to other sites.
- No content of the brand directory changes in M1. A tagline or hero text for each brand is a later change of the ADR 0023 contract, if it is necessary.

## 6. Telemetry, accessibility and the brand of each instance

Facts:

- ADR 0006: no CDR data and no personal information in telemetry, route patterns and not raw URLs (`docs/adr/0006-telemetry.md`, Sensitive data). The data holder telemetry has no customer number, name, email address or account data (`DESIGN.md:876`). The project adds: no customer number and no name.
- ADR 0014: the UIs send no browser telemetry. The server spans of the API show each request.
- A consumer UI targets WCAG 2.2 AA, is designed for 360 px and reflows at 320 px [S15], has controls of 44 px or more, and body text of 16 px or more (`ui/DESIGN.md`, Layout, Typography, Accessibility). The axe checks of the CDR pages run for each preset brand at 320 px (`ui/DESIGN.md`, Accessibility; `web/packages/ui/src/consumer-pages.test.ts`).
- Each brand sets the custom properties of the design system at run time with a brand stylesheet (`branddir/handler.go:18-26`; ADR 0023, Assets). The font of a brand is one of a fixed set (`DESIGN.md:100`).

Recommendations:

- M1 has only public product data, so its telemetry risk is low. Set the rules now for M2 to M4: the site API puts no query string, no path value and no request body in telemetry (the console rule, `DESIGN.md`, Console API); no path of the site API contains a customer number, because the session gives the customer (for example `/api/me/accounts`); the session cookie value never goes into telemetry.
- One SPA build for all brands. The site listener serves the full brand stylesheet, the logo and the mark on its own origin, as the console listener does. The build contains each font of the fixed set.
- The component and page tests load the brand stylesheet of each preset brand (the golden files in `consumerpages/testdata/brands/`) and run axe at 320 px, as `consumer-pages.test.ts` does. The lab-web tests run against the default brand only, as CI does now.

## 7. Separation from the CDR authentication (input for M2)

Facts:

- The CDR pages "never ask for a password" and say that the data recipient does not need the password (`DESIGN.md`, Consumer authentication; [S7]).
- The CDR pages have the header "Secure" label, a hero with the initial of the data recipient and the mark of the bank, a step indicator and one card (`ui/DESIGN.md:113-131`).

Recommendations for M1, so that M2 has a clear difference:

- A different listener and origin (section 3).
- A different layout: navigation in the header, no "Secure" label, no data recipient, no steps.
- The page title of each site page starts with the page name and ends with "<brand> online banking". The CDR pages keep their titles.
- M2 decides the text of the sign-in page. A candidate: "This is <brand> online banking. To share your data with an app under the Consumer Data Right, you do not sign in here and you never give your password."

## 8. Tests

Facts:

- ADR 0019: the conformance suite tests only behaviour that the CDR defines. Tests in the service test each interface that the CDR does not define, in the process, with `mise run test`. For a console API, these tests validate each response against its OpenAPI document with code in `pkg/`. Browser tests check the main journeys against the lab, and are not part of the conformance suite (`docs/adr/0019-…md:25-27`).
- The lab-web job of the data holder in CI runs with no conformance data (`.github/workflows/ci.yml:511`). The product catalogue is in each lab, because a seed container loads it (README, Data).

Recommendations:

- The site is not in the conformance suite (ADR 0019). The site is not CDR behaviour.
- Data holder tests (Go): the site listener (headers, CSP, `Host` allowlist, cross-origin refusal, TLS); each site API response against its OpenAPI document; **the equality test**: for each preset brand, load the catalogue into a test store, and compare the site list with Get Products (`effective=CURRENT`) and each site detail with Get Product Detail, field by field; the site gives no future or past product; the enumerations of the OpenAPI document equal the domain values.
- UI tests (Vitest browser): each page with `openapi-msw` handlers, each preset brand stylesheet, axe at 320 px; each label `Record` (compiler); the formatters of rates, amounts and durations.
- Lab-web (Playwright): one journey: open the landing page, go to a category, open a product, and compare the values with Get Products and Get Product Detail of the lab on 8181.

## 9. Risks and open questions for the requirements PR

| Item | Notes |
|---|---|
| ADR 0014 says "The browser gets no credential". M2 needs a session cookie. | Decide in M2 with a change to ADR 0014 or a new ADR. M1 chooses TLS for the listener so that a `__Host-` cookie works. |
| Names of the new app, listener, settings and port (new conventions). | Proposal: app `web/apps/dataholder-bank`, settings `BANK_SITE_ENABLED` and `BANK_SITE_ALLOWED_HOSTS`, port 8x85. Needs approval before use (AGENTS.md). |
| The block of five ports becomes six (`DESIGN.md:863`), and the README port table gets one row. | A README change that the "Keep the README a summary" test permits: a new port. |
| `consolehttp` is plain HTTP, and its names are for a console. | Give it a TLS option, or move the shared parts to a general package. |
| Cookies of all lab ports are shared on `localhost`. | Not fixable in the lab without host names. Different cookie names. Not a production gap (out of scope). |
| Playwright and the lab TLS certificate. | Add the lab root CA to the browser of the test, or `ignoreHTTPSErrors` for the site project only. |
| Fields that a realistic product page needs and that the data holder does not hold: eligibility, constraints, `applicationUri`, `additionalInformation`. | Out of scope for M1. Adding them changes the `products.json` contract (ADR 0023) and the CDS validation test of each preset. |
| `redgum-bank` sells business products, but each customer is a person (`DESIGN.md`, Customers and accounts). | Not a problem for M1. M3 ("open an account from a product") must decide who can open a business product. |
| An account holds `productCategory` and a free-text `productName`, not a `productId` (`DESIGN.md:139`). | M3 needs a link from an account to a product, or a copy of the name at the open. M1 does not block either. |
| Labels of 97 enumeration values, and the format of `additionalValue` by type. | Wording and format for each type are a large part of the M1 work. Decide which types get a special format. |
| Future products. | Recommendation: do not show. Alternative: a "Coming soon" group. |
| The CDR policy of the data holder on its web site. | Not researched in this note. A real data holder publishes a CDR policy. Out of scope for M1 unless the requirements PR adds it. |

## Main findings

- ADR 0014 and the design guide already make a data holder consumer UI a React SPA. Server-rendered Go pages for the site need a new ADR, and they put the bank sign-in close to the CDR pages.
- The public listener gives CORS `*` to each path except the authorisation path, and sends each unknown path to a CDS JSON error. A site on it needs exceptions in the CDR listener. A separate, lab-only listener with the console rules needs none.
- Cookies are not isolated by port, and a `__Host-` cookie always has `Path=/`. Thus neither a different port nor a path separates a site session cookie from the CDR pages in the lab.
- The public certificate of each bank has the SAN `localhost`, so a new TLS listener for the site needs no PKI change. `http://localhost` is a secure context, but a host other than localhost needs HTTPS for `crypto.randomUUID()`.
- A site API on `app.ProductReader` gives the same data as Get Products and Get Product Detail by construction, and a test in the data holder can prove it for each preset brand. The conformance suite does not test the site (ADR 0019).
- The data holder holds no eligibility, constraints or application URI for a product, so the detail page cannot show them.
- The CDS requires each consumer-facing interaction of a data holder to state that third parties do not need consumer passwords. The site must have this statement.
- The product enumerations have 97 values, and `additionalValue` has a different meaning for each type. The labels and the formats are the largest content task of milestone 1.
