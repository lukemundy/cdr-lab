# Branded data holder instances: technical research

- Research date: 2026-11-01
- Consumer Data Standards (CDS) version: 1.36.0. It is still the current version: the live site says "These standards represent version 1.36.0 of the high level standards". [S1]
- File and line references are to the repository on the research date.
- Purpose: the technical facts for milestone 1 of the project "Each data holder instance is a distinct, branded bank": the brand directory, brand tokens at run time, the rename of "Mock Bank", the product catalogue, and the multi-brand rules. The names, palettes, logos and visual direction are not in this note. The note [Preset bank brands: design](2026-11-01-preset-bank-brand-design.md) has them.
- Related notes: [CDS banking and common APIs](2026-10-06-cds-banking-and-common-apis.md) (product reference data), [CDS Register APIs](2026-09-25-cds-register-apis.md) (brand records), [Web front ends](2026-10-22-web-front-ends.md) (tokens). This note adds to them and does not replace them.

Check the sources again before you depend on a fact; the rules and standards change.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site; the archive URL redirects to it while 1.36.0 is current) | https://consumerdatastandardsaustralia.github.io/standards/ |
| S2 | CDS 1.36.0 OpenAPI, banking (pinned copy) | `conformance/dataholder/spec/cds_banking.json` |
| S3 | CDS 1.36.0 OpenAPI, Register (pinned copy) | `conformance/register/spec/cds_register.json` |
| S4 | DSB CX Guidelines: Redirect to Web with One Time Password | https://cx.dsb.gov.au/authenticate/redirect-with-one-time-password/cx-guidelines |
| S5 | CDR Support Portal: Brands in the CDR ecosystem (900005881263) | https://cdr-support.zendesk.com/hc/en-us/articles/900005881263 |
| S6 | RFC 2606, Reserved Top Level DNS Names, section 2 | https://www.rfc-editor.org/rfc/rfc2606#section-2 |
| S7 | RFC 6761, Special-Use Domain Names, section 6.2 | https://www.rfc-editor.org/rfc/rfc6761#section-6.2 |
| S8 | Tailwind CSS v4: Theme variables, "Referencing other variables" (`@theme inline`) | https://tailwindcss.com/docs/theme#referencing-other-variables |
| S9 | MDN: SVG as an image | https://developer.mozilla.org/en-US/docs/Web/SVG/Guides/SVG_as_an_image |
| S10 | MDN: CSP `sandbox` directive | https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/sandbox |
| S11 | Kubernetes: ConfigMaps ("cannot exceed 1 MiB") | https://kubernetes.io/docs/concepts/configuration/configmap/ |
| S12 | Australian Business Register: ABN format (11 digits with a check algorithm) | https://abr.business.gov.au/Help/AbnFormat |

## 1. The data holder now

### 1.1 Brand identity and configuration

- The API process requires `DATA_HOLDER_BRAND_NAME`, the `brandName` of the brand on the Register. The worker process requires `DATA_HOLDER_BRAND_ID`. Neither has a default. (`services/dataholder/internal/root/config.go:78`, `:108`; `services/dataholder/DESIGN.md:68-69`)
- The API process also requires `MAIL_FROM`, and accepts `FIXED_OTPS`. (`config.go:113`, `:146`)
- The console API gives the brand name at `GET /api/brand`. The console shows "Data holder" until the response. A Decision says that a change of the HTML page at run time "adds a second contract". (`DESIGN.md:943`; `web/apps/dataholder-console/src/router.tsx:35-47`)
- The console uses "the neutral lab brand". (`DESIGN.md:713`)

### 1.2 Consumer authentication and consent pages

- The pages are Go `html/template` files, embedded with `go:embed`. (`services/dataholder/internal/adapter/consumerpages/consumerpages.go:34-37`)
- The pages have no CSS, no images and no data holder name. The layout is a bare `head` and `main`. (`consumerpages/templates/layout.html`)
- The CSP is `default-src 'none'; frame-ancestors 'none'; base-uri 'none'`. (`consumerpages.go:431`) Thus a stylesheet, an image or a font cannot load now.
- The field `view.Brand` is the brand name of the **data recipient**, not of the data holder. (`consumerpages.go:94-95`, `:161`)
- The text says "this bank" in many places, for example "This bank will share your data one time." The conformance suite and the unit tests assert this text. (`conformance/dataholder/consent_test.go:29`, `:107`, `:170`, `:269`, `:364`; `consumerpages/consent_test.go:43-256`)
- The conformance suite asserts only that the CSP contains `default-src 'none'`. (`conformance/dataholder/authorise_test.go:251`)
- A DESIGN rule says that the pages "load no content from other hosts". (`DESIGN.md:429`)

### 1.3 The web UIs and the tokens

- The tokens are DTCG 2025.10 files with a resolver. The `brand` modifier has one context, `lab`. It sets the colour roles, the font families and the radius. (`web/packages/tokens/tokens/resolver.json`, `tokens/brand/lab.json`)
- Terrazzo builds the tokens at build time into `tokens.css`, as `--ds-*` custom properties on `:root`. (`web/packages/tokens/terrazzo.config.ts`; `src/tokens.ts`)
- `theme.css` maps each Tailwind theme value to a `--ds-*` variable with `@theme inline`. (`web/packages/tokens/src/theme.css:2-31`) With `inline`, a utility class uses the value as written, here `var(--ds-color-primary)`. [S8] Thus a stylesheet that sets `--ds-*` again after `tokens.css` changes the colours of a built UI, with no new build.
- The design system lets a brand change only "Colours, font family, radius and logo". (`web/packages/ui/DESIGN.md:55`, `:170`) A test checks the contrast of each pair of roles for each brand. (`ui/DESIGN.md:48`; `tokens/src/contrast.ts`)
- Fonts are self-hosted from `@fontsource-variable` packages. Only Inter and JetBrains Mono are in the build. (`web/packages/tokens/package.json`; ADR 0016, line 39)
- The console CSP is `default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; ...`. (`pkg/platform/consolehttp/consolehttp.go:22`) Thus a same-origin stylesheet and a same-origin SVG can load.
- The console `index.html` is the same for all brands. (`web/apps/dataholder-console/index.html`)

### 1.4 Image and build

- The final image contains only the binary. (ADR 0018, line 20; `services/dataholder/Dockerfile`) Each file that the service reads at run time is in the binary (`go:embed`) or comes from a mount.
- The build context copies all of `services/dataholder/`. (`Dockerfile`: `COPY services/dataholder/ services/dataholder/`) Thus a directory under `services/dataholder/` can be embedded with no change to `.dockerignore`.
- `go:embed` cannot read a path outside the directory of the Go package, or a parent directory. Thus embedded presets must be under `services/dataholder/`.

### 1.5 Products and their loader

- `dataholder seed [-if-empty] <file.json>` loads a full set of products into the database. It is separate from the start of the service. (`services/dataholder/internal/root/subcommands.go:25-63`; `DESIGN.md:778-781`)
- The file is `{"products": [...]}`, with the CDS field names. The reader refuses unknown top-level fields. `app.Seed` validates each product. (`internal/adapter/seedfile/seedfile.go`)
- The example file has 6 products in 4 categories, each with `"brand": "Mock Bank"`: a savings account, a credit card, a personal loan, two home loans (one future) and one savings account that is no longer offered. (`services/dataholder/seed/example.json`) A unit test requires the 4 categories, a future product and a past product. (`internal/adapter/seedfile/example_test.go:45-69`)
- The conformance suite asserts the literal brand "Mock Bank" and filters by it. (`conformance/dataholder/products_test.go:129`, `:163`, `:173`)
- The data holder holds only these optional fields: the effective dates, `features`, `fees`, `depositRates` and `lendingRates`. It refuses `brandName`, `brandGroup`, `applicationUri`, `additionalInformation`, `cardArt`, `bundles`, `constraints`, `eligibility`, `instalments`, fee discounts, rate tiers and rate conditions. (`DESIGN.md:816`)

## 2. Brand directory contract (question 1)

### 2.1 What a brand needs, by consumer of the data

| Item | Used by | Source of the rule |
|---|---|---|
| Brand name | Consumer pages, emails, console, product `brand`, the Register `brandName` | `DESIGN.md:68`; CDS: `brand` is "A label of the brand for the product ... For data holders with single brands this value is still required." [S2, `BankingProductV6.brand`] |
| Legal entity name | Register record; O of the server certificate | `legalEntityName` has `maxLength: 36` [S3, `LegalEntityDetail`]. The Register accepts an O that is the brand name or the legal entity name (`services/register/internal/root/certificateissue_test.go:155`). |
| Colour roles, font family, radius | Consumer pages, console, later the consumer site | `ui/DESIGN.md:55` |
| Logo (wordmark) and a small mark | Pages, console, favicon, Register `logoUri` | Register `logoUri` is mandatory: "Brand logo URI." [S1 #cdr-register-api_registerdataholderbrandv2_properties] |
| Product catalogue | `seed` subcommand | `DESIGN.md:778-781` |
| Email texts | Milestone 2 | The OTP email content is fixed by the CDS and `DESIGN.md:390`. Only the brand name and the look change. |

- Fact: the Register gives `abn`, `acn` and `arbn` only as optional fields. [S3, `LegalEntityDetail`] An ABN has 11 digits and a check algorithm. [S12] An invented ABN that passes the check can be the ABN of a real business. Recommendation: the brand directory has no ABN or ACN.
- Fact: the Register assigns `dataHolderBrandId`, and the base URIs are the host names of a deployment. (`DESIGN.md:67-71`) Recommendation: these stay in the configuration, not in the brand directory. The brand directory holds only the identity that is the same in each deployment.

### 2.2 Proposed directory layout

This is a proposal for the ADR. The names are not a convention until the requirements accept them.

```
<brand>/
├── brand.json     # name, legal entity name, short name for the host and the slug, and the version of the format
├── tokens.json    # one DTCG 2025.10 context of the brand modifier: color.*, font.*, radius.*, literal values only
├── logo.svg       # the wordmark, for headers
├── mark.svg       # a square symbol, for small sizes and the favicon
└── products.json  # the seed format of today: {"products": [...]}
```

- `tokens.json` has the same token names as `web/packages/tokens/tokens/brand/lab.json`, but with literal OKLCH values in place of aliases such as `{primitive.color.neutral.50}`. Reason: an alias needs the primitive set, which only the web build has. Literal values let the Go service convert the file to CSS with no Terrazzo.
- A role that the brand file does not give keeps the lab value, because the brand stylesheet loads after `tokens.css`. Thus an old brand directory still works after a new role is added.
- `font.*` can name only a font that the web build includes. A brand that needs a new font is a change to `web/packages/tokens`, not only a new directory. This limit is a decision for the requirements (see section 7).

### 2.3 Options for selection and load

| | A. Presets embedded in the binary, and an optional directory | B. Directory only, presets copied into the image | C. Directory only, presets mounted from the repository |
|---|---|---|---|
| Settings | `DATA_HOLDER_BRAND=<preset>` or `DATA_HOLDER_BRAND_DIR=<path>`, exactly one | `DATA_HOLDER_BRAND_DIR=<path>`; presets at a fixed path in the image, for example `/brands/<name>` | `DATA_HOLDER_BRAND_DIR=<path>` |
| Code paths | One loader on `fs.FS`: `embed.FS` or `os.DirFS` | One | One |
| Distroless image | Works: no file outside the binary | Needs a `COPY` of files into the final image, against ADR 0018 ("The final image contains only the binary") | Works, but the image alone cannot run a preset |
| Helm chart | A preset needs one value. A user brand needs a ConfigMap mount. | A preset needs one value. The path in the image becomes a contract. | Each brand needs a ConfigMap with its files, also the presets. The chart cannot read files from the repository after it is published. |
| User brand (no code change) | Yes, with the directory setting | Yes | Yes |
| Existing idiom | Templates, migrations and the console build are already `go:embed` (`consumerpages.go:34`, `migrations/migrations.go:6`, `console/console.go:18`) | None | The seed files are mounts (`compose.yaml:83`, `:170`) |

- Fact: a ConfigMap "cannot exceed 1 MiB". [S11] SVG logos, a JSON catalogue and a tokens file are much smaller than this.
- **Recommendation: option A.** It keeps ADR 0018 as it is, it follows the `go:embed` idiom of the service, and the published image runs each preset with one setting. The presets are in `services/dataholder/brands/<name>/`. The two settings are the only cost. A single setting with two forms (a name, or a path that starts with `/`) is a smaller alternative. The requirements choose.

### 2.4 Validation at start-up

Recommendation. The API process and the `seed` subcommand load and validate the brand. The worker process does not need it (`DESIGN.md:791`, least privilege).

- `brand.json`: strict JSON decode (unknown fields refused, as `seedfile.go` does), each required field present and not empty, the name at most 200 characters, the legal entity name at most 36 characters [S3].
- `tokens.json`: valid DTCG JSON. Each token is in `color`, `font` or `radius`. Each colour is a literal OKLCH value in the sRGB gamut, as the Terrazzo lint requires for the lab brand (`terrazzo.config.ts`: `core/colorspace`, `core/max-gamut`).
- `logo.svg`, `mark.svg`: well-formed XML with an `svg` root and a `viewBox`, and a size limit.
- `products.json`: the validation of `app.Seed` of today, plus: the `brand` of each product is the brand name.
- The error names the file and the rule, as the other settings do (`config.go:80-139`).
- The contrast check stays in the web tests (`contrast.ts`), for each preset directory. A Go copy of the pairs list would be a second source of the same rule. A user brand with a low contrast then starts. This is a question for the requirements (section 7).

## 3. Brand tokens and logos at run time (question 2)

### 3.1 Options

| | A. Build-time permutations | B. A brand stylesheet that the Go service makes | C. Tokens in the console API, applied by JavaScript |
|---|---|---|---|
| How | Terrazzo builds each preset into `tokens.css` under a selector, for example `[data-brand=x]` | At start-up, the service converts `tokens.json` to `:root { --ds-...: oklch(...) }` and serves it at a fixed path, for example `/brand/tokens.css` | `GET /api/brand` gives the tokens. The SPA sets them on `document.documentElement.style` |
| User brand with no code change | No: a new web build and image | Yes | Yes, for the SPAs only |
| Consumer pages (no JavaScript) | Yes | Yes | No |
| Flash of the lab brand | No | No: the stylesheet is in the HTML | Yes, until the response |
| CSP change | None | `style-src 'self'` on the consumer pages | None |

- **Recommendation: option B.** One stylesheet serves the consumer pages, the console, and later the consumer site. It needs `@theme inline` (already in place [S8], `theme.css:2`).
- The brand rule must win over `tokens.css` in each load order. A selector with a higher specificity, for example `:root:root`, does this. Then the position of the `link` element in the Vite output has no effect.
- The console `index.html` gets a static `link` to the brand stylesheet and a static `link` to the favicon. The HTML is then the same for each brand, so the Decision at `DESIGN.md:943` stays true. The brand name still comes from `GET /api/brand`.

### 3.2 Logos

- Fact: an SVG that a page loads with `img` cannot run scripts or load other resources. [S9] A user can still open the SVG URL directly. Then a script in the SVG runs in the origin of the data holder.
- Fact: the CSP `sandbox` directive puts a response in a unique origin and stops scripts unless it allows them. [S10]
- Recommendation: serve each logo with `Content-Type: image/svg+xml`, `X-Content-Type-Options: nosniff` and `Content-Security-Policy: default-src 'none'; style-src 'unsafe-inline'; sandbox`. Pages use the logo only in `img`. No SVG goes inline into a page. This is sufficient for the lab, and it needs no SVG sanitiser.

### 3.3 The consumer pages need more than tokens

- The pages have no CSS now (section 1.2). Tokens alone give no layout or components. The pages need a stylesheet of the design system.
- Option: a CSS-only build in the web workspace. Tailwind scans the Go templates (`@source` to `services/dataholder/internal/adapter/consumerpages/templates`), and the data holder embeds the output, as it embeds the console build (`Dockerfile`: `COPY --from=web ... console/dist/`). ADR 0014 gave this consequence: "two ways to render UI: React components, and Go templates with the same CSS".
- ADR 0016 says: "It does not change the data holder authorisation pages now" (line 33). Milestone 1 changes this. The ADR needs an edit. ADR 0014 does not change: the pages stay Go pages with no JavaScript.
- The CSP of the consumer pages becomes, for example, `default-src 'none'; style-src 'self'; img-src 'self'; font-src 'self'; frame-ancestors 'none'; base-uri 'none'`. The conformance assertion at `authorise_test.go:251` stays true. The rule "load no content from other hosts" (`DESIGN.md:429`) stays true.
- The brand routes (`/brand/*`) are on the public TLS listener (port 8181, which serves the pages) and on the console listener. They hold no CDR data.
- Fact: the SSA `logo_uri` is "URL string that references a logo for the client software product. The server SHOULD display this image to the end-user during approval." [S1 #ssa-definition] The data holder does not show it now, because of the rule "load no content from other hosts". This is a separate question, not part of this project.

## 4. What the rename touches (question 3)

`git grep` for `Mock Bank`, `mock-bank`, `mock_bank`, `MAIL_FROM`, `consumer.example`, `FIXED_OTPS` and `DATAHOLDER_*`, without `docs/research/` and the lock file, finds 40 files. The groups:

| Area | Files and lines | What changes |
|---|---|---|
| Compose | `compose.yaml:89-170` (service, build anchors, `PGUSER`/`PGDATABASE`, brand ID and name, base URLs, `OPERATOR_ALLOWED_HOSTS`, `MAIL_FROM`, `FIXED_OTPS`, TLS mount, migrate and seed containers), `:189-251` (conformance and openid runners), `:268`, `:366` | Service name, host names, database users, brand settings, mail sender, test addresses |
| Lab database | `lab/postgres/init.sql:10-14`, `lab/postgres/grants.sql:4` | Database and user names. A new user needs `mise run reset` (ADR 0020). |
| PKI | `mise.toml:189-213` | `.lab/pki/mock-bank/` directory, the TLS host names, the O of the server certificate (`"Mock Bank"`) |
| Register seed | `services/register/seed/example.json:3-26` | Brand name, legal entity, `brandGroup`, `logoUri`, base URIs, `websiteUri`, JWKS URL |
| Register tests | `services/register/internal/root/certificateissue_test.go:75-155`, `dataholderoperator_test.go:15-16`, `operator_test.go:96` (they load the seed file), `conformance/register/data_holders_test.go:16`, `:25`, `:39-49`, `:97-108` | The Register tests assert the full sets of seed brands and legal entities for each industry. Each new preset record changes these sets. |
| Register console tests | `web/apps/register-console/e2e/*.spec.ts` (`Mock Holdings Ltd`, the brand `Mock Bank`, `https://mock-bank:8083`) | Names and URIs, if the seed record changes |
| Data holder conformance | `conformance/dataholder/products_test.go:129-173` (literal brand), `consent_test.go` ("this bank"), `otp_test.go:91`, `:232-233`, `customer_test.go:161` (`consumer.example`), `conformance/openid/openid_test.go:27-29` | Get the brand name from the Register (`labtest.RegisteredDataHolderBrand`, `conformance/internal/labtest/dataholder.go:69-78`) in place of a literal |
| Data holder unit tests | `httpapi/products_test.go`, `productstore/*_test.go`, `domain/product_test.go`, `app/seed_test.go`, `console/customers_test.go`, `root/config_test.go`, `pkg/cds/page_test.go` | Test data only. A unit test can keep an invented name. The AC is only about the preset configuration. |
| Web e2e | `web/apps/dataholder-console/e2e/customers.spec.ts:6-7` | Brand name |
| CI | `.github/workflows/ci.yml:490`, `:495` | The compose service name in the image check job |
| Docs | `README.md:136-140` (ports table), `services/dataholder/cmd/dataholder/main.go:19-20` (usage), `pki/DESIGN.md:114`, `services/dataholder/DESIGN.md:781` (example file) | Names, settings and the example file rule |

Facts for the decisions:

- The brand "Mock Bank" shares the legal entity "Mock Holdings Ltd" with "Mock Energy" (energy and telco), "Mock Lending" and "Mock Mutual". (`services/register/seed/example.json:3-60`) These records are fixtures of the Register tests, for example a legal entity with brands in more than one industry. A real major bank is its own legal entity.
- The Register seed loads one file. (`services/register/internal/root/subcommands.go:29-30`: `usage: register seed [-if-empty] <file.json>`)
- `.example` is reserved by RFC 2606 [S6]. `.test` is also reserved, "for use in testing", and DNS servers must not resolve it to real hosts. [S7] The AC "No ... email address or host name in the preset configuration contains 'Mock' or 'Example'" forbids `.example`. A realistic domain such as `<name>.com.au` can belong to a real business.
- **Recommendation:** use `.test` for each invented domain in the preset configuration: mail sender, `websiteUri`, customer email addresses. Container host names stay single labels (the compose service name), as now.
- **Recommendation:** the compose service, the PKI directory and the database names use the short name of the brand (`brand.json`). In milestone 3, each brand then has its own names with one pattern. A neutral name such as `bank` would need a second pattern in milestone 3.

## 5. Product catalogue (question 4)

### 5.1 What the data holder serves

- Get Products version 5 (`BankingProductV6`) and Get Product Detail version 7 (`BankingProductDetailV7`). (`services/dataholder/internal/adapter/httpapi/httpapi.go:21-22`; `DESIGN.md:114-115`)
- Required fields of `BankingProductV6`: `productId`, `lastUpdated`, `productCategory`, `name`, `description`, `brand`, `isTailored`. [S2]
- `BankingProductDetailV7` adds only optional arrays: `bundles`, `constraints`, `depositRates`, `eligibility`, `features`, `fees`, `instalments`, `lendingRates`. [S2]
- The 13 values of `BankingProductCategoryV2`: `BUSINESS_LOANS`, `BUY_NOW_PAY_LATER`, `CRED_AND_CHRG_CARDS`, `LEASES`, `MARGIN_LOANS`, `OVERDRAFTS`, `PERS_LOANS`, `REGULATED_TRUST_ACCOUNTS`, `RESIDENTIAL_MORTGAGES`, `TERM_DEPOSITS`, `TRADE_FINANCE`, `TRANS_AND_SAVINGS_ACCOUNTS`, `TRAVEL_CARDS`. [S2; S1 #product-categories]
- CDS definitions that matter for a retail bank [S1 #product-categories]:
  - `TRANS_AND_SAVINGS_ACCOUNTS`: "all accounts where cash is deposited in the account and is accessible to the customer when they choose", for example transaction, savings and cash accounts.
  - `TERM_DEPOSITS`: "cash is deposited in the account for a set time period with restrictions on when funds can be withdrawn".
  - `TRAVEL_CARDS`: "prepaid cards with multi-currency capabilities".
  - `OVERDRAFTS`: lending that "allows for the loan amount to be withdrawn, repaid, and redrawn again in any manner and any number of times".
  - `PERS_LOANS`: "all lending for personal purposes that is not a residential mortgage, credit card or margin lending", including car loans.

### 5.2 A catalogue with five or more categories

Recommendation for each preset, with the fields that the data holder holds now:

| Category | Products | Fields that make it realistic |
|---|---|---|
| `TRANS_AND_SAVINGS_ACCOUNTS` | An everyday account, a bonus saver, a high-interest saver for young customers | `fees` (monthly fee 0 or with a waiver in `additionalInfo`), `depositRates` `VARIABLE` and `BONUS`, `features` `NPP_PAYID`, `DIGITAL_WALLET`, `CARD_ACCESS` |
| `TERM_DEPOSITS` | One term deposit | `depositRates` `FIXED` with `additionalValue` as an ISO 8601 duration for each term, `fees` `EXIT` (early withdrawal) |
| `CRED_AND_CHRG_CARDS` | A low-rate card and a rewards card | `lendingRates` `PURCHASE`, `CASH_ADVANCE`, `BALANCE_TRANSFER`; `fees` `PERIODIC` (annual), `LATE_PAYMENT`; `features` `INTEREST_FREE`, `LOYALTY_PROGRAM` |
| `PERS_LOANS` | An unsecured personal loan and a secured car loan | `lendingRates` `FIXED` with `comparisonRate`, `repaymentType` `PRINCIPAL_AND_INTEREST`; `fees` `UPFRONT`, `PERIODIC` |
| `RESIDENTIAL_MORTGAGES` | A variable owner-occupier loan, a fixed loan, an investment loan | `lendingRates` with `loanPurpose` `OWNER_OCCUPIED` or `INVESTMENT`, `repaymentType`, `comparisonRate`; `features` `OFFSET`, `REDRAW`, `EXTRA_REPAYMENTS` |
| `OVERDRAFTS` or `TRAVEL_CARDS` (optional sixth) | A personal overdraft, or a multi-currency travel card | `lendingRates` `VARIABLE`; or `fees` `TRANSACTION` with `rateBased` |

- Each preset also keeps one future product and one product whose `effectiveTo` is in the past. The conformance filter tests need them (`DESIGN.md:781`; `seedfile/example_test.go:58-69`). A realistic catalogue has both: an offer that starts soon, and an offer that closed.
- The enumerations that limit the values are in the pinned schema [S2]: `featureType` (31 values), `feeType` (17), `depositRateType` (7), `lendingRateType` (12), `repaymentType` and `loanPurpose` (4 each).
- Realism gaps with the fields of today (`DESIGN.md:816`): no `tiers` (a home loan rate in real product data usually changes with the loan-to-value ratio), no `eligibility` (minimum age, residency), no `constraints` (minimum and maximum LVR, opening balance), no `cardArt`, no `applicationUri`. A rate for each LVR band can be written now as separate lending rates with `additionalValue` text, but a client cannot read it as a tier. Whether to add `eligibility`, `constraints` and `tiers` is a question for the requirements (section 7).
- The product data must not copy the names, fees or rates of real products (`DESIGN.md:90`). Four catalogues of 10 or more products each are a large amount of hand-written JSON. A test must check each preset catalogue (see section 7).

### 5.3 Schema validity of each preset

- The AC requires a Get Products response "valid against the CDS schema" for each preset. Now only the conformance suite validates responses against the pinned OpenAPI (`conformance/dataholder/openapi.go`), and it runs against one data holder.
- `app.Seed` checks the DESIGN product rules, not the OpenAPI schema. A unit test can load each preset through `app.Seed`. For the schema of the other three presets, milestone 3 can run the products conformance tests against each brand in the compose profile, or a data holder test can validate the response of each preset against the pinned file. The second option reads a file of the conformance component from a service test. This is a question for the requirements.

## 6. CDS multi-brand rules (question 5)

- "For the purposes of this standard a single designated organisation MAY be represented via the Register as multiple separate Data Holders to support multiple brands or market identities." [S1 #multi-brand-support-separate-issuers-for-data-holder-brands, preceding paragraph]
- "Where a Data Holder has multiple brands, each brand MUST have a separate issuer." (from 4 July 2022) [S1 #multi-brand-support-separate-issuers-for-data-holder-brands] One instance for each brand meets this, because each instance has its own issuer (`DESIGN.md:72`).
- The product `brand` feature "is intended only for sub-brands of a master brand and should not be overloaded with products belonging to any other brand hosted on the Register." [S5] Thus the catalogue of one preset must not contain the products of another preset.
- Each Register brand record has a mandatory `brandName`, `industries`, `logoUri`, `legalEntity`, `status`, `endpointDetail` and `authDetails`; `brandGroup` is optional. [S1 #cdr-register-api_registerdataholderbrandv2_properties; S3]
- `legalEntityName` has `maxLength: 36`. [S3, `LegalEntityDetail`] The names that the design work chooses for the legal entities must fit this limit.
- The client ID of a data holder when it calls a software product is the "Data Holder Brand ID" that the Register issues. [S1 #data-holders-calling-data-recipients] Each preset needs its own `dataHolderBrandId`, and the compose configuration and the conformance runner use it (`compose.yaml:107`, `:194`, `:251`).
- Recommendation for the Register seed: each preset is its own legal entity with one brand, as a major bank is. Its `logoUri` is the URL of the logo on the instance, for example `https://<host>:8081/brand/logo.svg`, so that a data recipient can show the real logo. Its `websiteUri` uses `.test`.
- The Helm charts need a participant seed file with the four presets for the Register chart. The Register seed reads one file, and the Register tests depend on the fixtures of `example.json`. Recommendation: keep the preset records in the same file for milestone 1, and let the chart work decide if it needs a second file.

## 7. Risks and open questions (question 6)

### Risks

| Risk | Detail |
|---|---|
| The Register tests list each seed record | `conformance/register/data_holders_test.go:39-49` and `:97-108` assert the full set of data holders and brands for each industry. Each preset that the seed adds changes these tests, and milestone 3 adds three more. A change of the tests to check only their own fixtures removes this coupling. |
| Page text in tests | The conformance suite asserts "this bank". A change to the brand name in the text changes these tests. The suite must get the name from the Register, not from a literal. |
| A copy of the contrast rule in Go | If the start-up check repeats `contrast.ts`, two sources of one rule can disagree. |
| SVG from a user directory | A script in an SVG runs when a user opens its URL. The sandbox CSP of section 3.2 stops it. |
| Fonts | Fonts are build-time only. Four realistic banks with one font look less different. A new font per preset needs a licence that permits self-hosting (ADR 0016 rule). |
| Invented identifiers that are real | An invented ABN or domain can belong to a real business. Section 2.1 and section 4 avoid both. |
| Two sources of the brand name | `DATA_HOLDER_BRAND_NAME` and `brand.json` can disagree. The Register `brandName` is a third. |
| Catalogue maintenance | Approximately 40 hand-written products. Rates go out of date when the market changes. A static `lastUpdated` is correct for a mock, but the rates can look wrong later. |

### Questions for the requirements

1. Brand selection: option A (embedded presets and an optional directory), and one or two settings (section 2.3).
2. The source of the brand name: `brand.json` only, with `DATA_HOLDER_BRAND_NAME` removed, or both with a start-up check that they are equal. The Register seed must use the same name; a test can check the seed against the presets.
3. The fields of `brand.json`, and if the format has a version field. A format version is a new convention.
4. Contrast: a start-up check in Go, or only a web test for presets and a documented check command for a user brand.
5. The consumer page CSS: a Tailwind build that scans the Go templates, and the edit to ADR 0016 (line 33). The new CSP of the pages.
6. The text of the pages: the brand name in place of "this bank", and the footer. The footer text must be fixed in the templates, not in the brand directory, so that a user brand cannot remove it.
7. The product fields: keep the fields of today, or add `eligibility`, `constraints` and rate `tiers` for realism (`DESIGN.md:816`).
8. The schema check of the presets that the lab does not run by default (section 5.3).
9. The Register seed: a new legal entity and a new brand ID for the preset, or a rename of the "Mock Bank" record in place. A new legal entity is more realistic. It changes more Register tests, because "Mock Holdings Ltd" then has no banking brand except "Mock Mutual".
10. `MAIL_FROM`: configuration, as now, with a `.test` address, or a field of `brand.json`. A display name in the `From` header can come from the brand name.
11. The compose service name: the short name of the brand (recommended) or a neutral name.

### Needs of later projects that the contract must not block

- Branded emails: emails use the brand. Many mail clients do not show SVG. A PNG logo for email can be a later, optional file of the directory. Unknown files in the directory must then not be an error, or the format version must change.
- Four instances: each needs its own ports in `compose.yaml`; the README ports table now gives one block (8180-8184) for `mock-bank`.
- The consumer site: the site is a React SPA. Option B of section 3 works for it with no change. The landing page can need more brand content, for example a tagline or images. The format must accept new optional fields in a later version.
- The Helm chart: with option A, a preset needs one value. A user brand needs a ConfigMap, within 1 MiB [S11].

## Main findings

- The CDS has no rule for the brand of the data holder on its own authentication pages. The DSB CX Guidelines speak only of the brand name of the data recipient (2AU.00.16, 2AU.00.37) and of links that "could ... trigger phishing concerns" (2AU.00.36). [S4] The branding of the pages is judgement.
- One instance for each brand meets the CDS rule of a separate issuer for each brand. [S1]
- `theme.css` uses `@theme inline`, so a stylesheet that the Go service makes at run time can change the colours of an embedded SPA with no new build.
- The consumer pages have no CSS and a CSP of `default-src 'none'` now. Branding needs a CSS build for Go templates and a CSP change, and an edit of ADR 0016.
- Embedded presets keep ADR 0018 ("The final image contains only the binary") true and give the Helm chart one value for each preset.
- `legalEntityName` has a maximum length of 36 in the Register schema.
- The preset configuration cannot use `.example`. `.test` is reserved for tests by RFC 6761 and has no "Example".
- The Register conformance tests list each seed record, so each preset record changes them.
- The current catalogue has 4 categories. The data holder fields of today support 5 or more, but not rate tiers, eligibility or constraints.
