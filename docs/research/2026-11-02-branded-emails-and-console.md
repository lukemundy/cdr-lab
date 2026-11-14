# Branded emails and operator console: technical research

- Research date: 2026-11-02
- Consumer Data Standards (CDS) version: 1.36.0. The live site says "These standards represent version 1.36.0". [S1]
- CDR Rules: Compilation No. 10 (https://www.legislation.gov.au/F2020L00094/2025-03-04/text). [S2]
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 2 of the project "Each data holder instance is a distinct, branded bank": "The OTP and notice emails to the consumer, and the operator console of the instance, use the brand of the instance."
- Related notes: [Branded data holder instances](2026-11-01-branded-data-holder-instances.md) (brand directory, brand stylesheet), [Preset bank brands: design](2026-11-01-preset-bank-brand-design.md) (palettes, logos). This note adds to them and does not replace them.

Check the sources again before you depend on a fact; the rules, the standards and the email clients change.

## Legend

- **Fact**: a source, the code or a test in the lab says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site) | https://consumerdatastandardsaustralia.github.io/standards/ |
| S2 | CDR Rules, Compilation No. 10 | https://www.legislation.gov.au/F2020L00094/2025-03-04/text |
| S3 | DSB CX Guidelines: Redirect to Web with One Time Password | https://cx.dsb.gov.au/authenticate/redirect-with-one-time-password/cx-guidelines |
| S4 | RFC 2046, MIME Part Two: Media Types, section 5.1.4 (multipart/alternative) | https://www.rfc-editor.org/rfc/rfc2046#section-5.1.4 |
| S5 | RFC 2387, The MIME Multipart/Related Content-type | https://www.rfc-editor.org/rfc/rfc2387 |
| S6 | RFC 2392, Content-ID and Message-ID Uniform Resource Locators | https://www.rfc-editor.org/rfc/rfc2392 |
| S7 | RFC 5322, Internet Message Format, sections 2.1.1, 3.2.3, 3.2.5, 3.4, 3.6.2 | https://www.rfc-editor.org/rfc/rfc5322 |
| S8 | RFC 2047, MIME Part Three: Message Header Extensions for Non-ASCII Text, section 5 | https://www.rfc-editor.org/rfc/rfc2047#section-5 |
| S9 | Go `net/mail`: `Address.String`, `ParseAddress` | https://pkg.go.dev/net/mail |
| S10 | caniemail: SVG image format (`_features/image-svg.md`, last test 2026-09-16) | https://www.caniemail.com/features/image-svg/ |
| S11 | caniemail: PNG image format (last test 2020-02-06) | https://www.caniemail.com/features/image-png/ |
| S12 | caniemail: Base64 data URI images (last test 2020-02-06) | https://www.caniemail.com/features/image-base64/ |
| S13 | caniemail: `<style>` element (last test 2023-07-27) | https://www.caniemail.com/features/html-style/ |
| S14 | caniemail: lch(), oklch(), lab(), oklab() (last test 2023-01-18) | https://www.caniemail.com/features/css-modern-color/ |
| S15 | caniemail: CSS Variables (last test 2020-02-25) | https://www.caniemail.com/features/css-variables/ |
| S16 | Mailpit v1.32.0 source, `server/apiv1/message.go` (the part handler) | https://github.com/axllent/mailpit/blob/v1.32.0/server/apiv1/message.go |
| S17 | Gmail Help: Turn images on or off in Gmail | https://support.google.com/mail/answer/145919 |
| S18 | GitHub repository metadata (`pushed_at`, licence) of `srwiley/oksvg`, `srwiley/rasterx`, `tdewolff/canvas`, `kanrichan/resvg-go`, `linebender/resvg`, and their `go.mod` files | https://github.com/srwiley/oksvg and the others |
| S19 | WCAG 2.2 | https://www.w3.org/TR/WCAG22/ |

## What the data holder does now

- The data holder sends one kind of email: the OTP email. `app.Authentication` calls `Mail.Send` with `domain.OTPEmail(otp, recipientBrand)`. No other code calls the mailer. (`internal/app/authentication.go:53`; `grep` for `Mailer` and `.Send(` in `internal/app`)
- The "notice" of the data holder is an HTTP request to the data recipient (`POST <recipient_base_uri>/arrangements/revoke`), not an email. (`DESIGN.md:638-649`; `internal/app/notices.go:64`)
- `domain.Email` has a `Subject` and a plain text `Body` only. The OTP email has the OTP, the brand name of the data recipient and the validity time, and says that the OTP is only for CDR sharing. (`internal/domain/otp.go:88-108`)
- The mailer writes the `From` header as the raw value of `MAIL_FROM`, and one `text/plain; charset=utf-8` part with `Content-Transfer-Encoding: 8bit`. Only the subject is encoded. (`internal/adapter/smtpmail/smtpmail.go:90-101`) The configuration checks only that `MAIL_FROM` is not empty. (`internal/root/config.go:142`)
- The lab sets `MAIL_FROM: no-reply@wattlebank.test` and `PUBLIC_BASE_URL: https://wattle-bank:8081`. (`compose.yaml:110`, `:128`) The host name `wattle-bank` resolves only on the compose network.
- The brand name of `brand.json` has 1 to 200 characters, and is not only white space. No rule refuses a control character, for example a line break. (`internal/adapter/branddir/branddir.go:184-196`)
- The brand directory has exactly five files: `brand.json`, `tokens.json`, `logo.svg`, `mark.svg` and `products.json`. ADR 0023 says: "A later file, for example a PNG logo for email ... is a change of the contract". (`docs/adr/0023-...md:55`)
- The four preset logos (`services/dataholder/brands/*/logo.svg`, 4 to 6 KB) have no `<text>`, no `<style>`, no `class` and no gradient. The test preset `fixture` and the old lab logo used `<text>`. (`git show` of each file)
- The brand routes `/brand/logo.svg`, `/brand/mark.svg` and `/brand/styles.css` are on the public TLS listener only. (`internal/root/api.go:86`; `internal/adapter/branddir/handler.go`) The console listener has its own mux with the console routes only. (`internal/root/console.go:23-24`)
- The brand stylesheet sets each `--ds-*` property of the brand on `:root:root`. (`internal/adapter/branddir/tokens.go:68`) Its colours are OKLCH. (ADR 0023, Tokens)
- The console gets the brand name from `GET /api/brand` and shows it as text in the header of `AppShell` and in the tab title. Until the response it shows "Data holder". (`internal/adapter/console/console.go:49`; `web/apps/dataholder-console/src/router.tsx:35-50`; `DESIGN.md:992`) `AppShell` is shared with the Register console and has a `name` prop only. (`web/packages/ui/src/components/app-shell.tsx:9-36`)
- The console CSP is `default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; ...`. (`pkg/platform/consolehttp/consolehttp.go:22`)
- The console "uses the neutral lab brand". (`DESIGN.md:762`; `web/packages/ui/DESIGN.md:55`: "The lab brand is neutral. The consoles use it.")
- `index.html` of the console is static and the same for each brand. (`web/apps/dataholder-console/index.html`)
- The conformance suite reads the `Text` field of each message from the Mailpit API, and finds the OTP with the pattern `one-time password is ([0-9]{6})\.`. (`conformance/dataholder/otp_test.go:23-77`)
- The e2e test checks that the banner has exactly the text of the brand name. (`web/apps/dataholder-console/e2e/customers.spec.ts:13`)

## 1. The scope: "notice emails"

### Facts

- The CDR Rules require the data holder to notify the **accredited person** when an authorisation is withdrawn or expires (rule 4.26A), and to update the consumer dashboard (rule 4.27). [S2, p. 82]
- The only data holder duty in rules 4.25 to 4.28 to notify a **consumer** is rule 4.28: the data holder notifies the account holder "through its ordinary means of contacting the account holder" when a secondary user amends or withdraws an authorisation. [S2, rule 4.28(2)] Secondary users are out of scope. (`DESIGN.md:40`)
- The CDS notification standards are CDR receipts and 90-day notifications, both "provided by a data recipient", and joint account alerts. [S1, `#consumer-experience_notification-standards`] Joint accounts are out of scope. (`DESIGN.md:40`)

### Analysis

No source requires an email from the data holder to the consumer when a consent ends. The data holder sends no such email now. The word "notice" in the milestone thus has no existing email to give a brand to.

- **Option A: remove "notice emails" from milestone 2.** The milestone is then the OTP email and the console. Change the Scope line (`DESIGN.md:47`) and the milestone text.
- **Option B: add a new email to the consumer when a consent ends** (for example "Your consent to share data with <recipient> ended"). This is new behaviour from judgement, not a CDR requirement. It needs a customer email address at the end of each event of the end table, a template, and a decision on each end event (the end of the sharing, a withdrawal, the end of eligibility, a REMOVED software product).

**Recommendation: option A.** It agrees with "the least complex solution" of the README source process, and no rule gives option B. Option B can be a later ticket.

## 2. A branded transactional email

### Facts: MIME structure

- In `multipart/alternative`, "each of the body parts is an 'alternative' version of the same information", in the order of "increasing faithfulness to the original content", so the best part is last. [S4, 5.1.4]
- `multipart/related` joins a root part (for example HTML) with the parts that it refers to (for example images). [S5]
- A `cid:` URL refers to a body part by its `Content-ID`. [S6]
- A line of a message "MUST be no more than 998 characters" (excluding CRLF). [S7, 2.1.1] A generated HTML part can have longer lines unless it is quoted-printable or base64.
- The usual structure is `multipart/alternative` { `text/plain`, `multipart/related` { `text/html`, `image/png` with `Content-ID` } }. The Go standard library has `mime/multipart` (with `Writer.SetBoundary` for a fixed boundary in a test) and `mime/quotedprintable`.

### Facts: email client support

- `<style>` in the head has partial support. Gmail does not support it inside `<body>`, not "with non Google accounts" in its apps, and limits it to 16 KB. Outlook for Windows 2007 to 2016 is "buggy". [S13] Inline `style` attributes are the reliable method.
- CSS variables: 27 "y" and 36 "n" entries. [S15] OKLCH and the other modern colour functions: 8 "y" and 23 "n" entries; Gmail "n" on each platform. [S14] Thus an email cannot use the brand stylesheet or OKLCH values. It needs literal sRGB values, for example hex.
- SVG as a linked image is now usually supported: Gmail renders it "but rasterized in PNG" (2026-09 tests), and Apple Mail, Outlook.com, Outlook for Mac, Yahoo, Thunderbird and Samsung Email support it. Outlook for Windows 2007 to 2016, and Mail.ru, do not. [S10] The project brief said that most clients do not render SVG. That is no longer correct for linked SVG.
- caniemail tests SVG only "as a linked file". [S10] It has no data for an SVG in a `cid:` part.
- PNG: 100 % support (test of 2020-02-06). [S11]
- Base64 `data:` images: 80.95 % estimated support; Outlook for Windows does not render a base64 GIF; Orange, SFR and LaPoste webmail remove the `src`. The test is from 2020-02-06. [S12]
- Gmail can "Ask before displaying external images". [S17]

### Facts: the lab mail sink (Mailpit v1.32.0)

A test email in the running lab, sent through the Mailpit send API (`POST :8025/api/v1/send`), on 2026-11-02:

- An inline `image/svg+xml` part with a `Content-ID`: the HTML view refers to it as `/api/v1/message/<ID>/part/2`, and that URL responds with `Content-Type: application/octet-stream`. A browser does not show an `img` with that type. The source forces `application/octet-stream` and `attachment` for `text/html`, `image/svg+xml` and the XML types. [S16, lines 158-166]
- An inline `image/png` part responds with `Content-Type: image/png`. (Same test.)
- The HTML view keeps a remote `img` URL as it is (`https://localhost:8181/brand/logo.svg` in the test). The browser of the person who reads the sink loads it. The view CSP has `img-src * data: blob:`. (Same test.)
- The message API gives `From.Name`, `From.Address`, `Text`, `HTML`, and `Inline[]` with `ContentType` and `ContentID` of each inline part. `/api/v1/message/<ID>/html-check` gives a score from the caniemail data, and a warning for each feature with partial support (the test email got "SVG image format", `<body>`, `background`, `height`, `width`, `padding`). (Same test.)
- Mailpit encodes a display name with specials as a quoted string: `"Wattle Bank, \"Test\" Ltd" <no-reply@bank.test>`. (Same test.)

### Options for the logo

| | A. Remote URL (`<public base URI>/brand/logo.svg`) | B. `cid:` part with `logo.svg` | C. `cid:` part with a PNG from the brand directory | D. `cid:` part with a PNG that the data holder renders from `logo.svg` |
|---|---|---|---|---|
| Lab (Mailpit) | Fails: `https://wattle-bank:8081` does not resolve on the host, and the browser does not trust the lab public CA | Fails: Mailpit serves the part as `application/octet-stream` [S16] | Works (lab test) | Works, if the render is correct |
| Real clients | Linked SVG mostly works [S10]; blocked until the reader permits images in some clients [S17] | No data [S10] | PNG 100 % [S11] | PNG 100 % [S11] |
| Privacy | Each open is a request to the bank: an open tracker | No request | No request | No request |
| Contract change (ADR 0023) | None | None | New file in the directory | None |
| New dependency | None | None | None (`image/png` in the standard library can check the file) | Yes |
| Fidelity | Same file as the pages | Same file | The author keeps two files in step | Depends on the SVG features that the renderer supports |

Facts for option D [S18]:

- Go has no SVG rasteriser in the standard library.
- `srwiley/oksvg` (BSD-3-Clause, "Partial implementation of SVG 2.0", last push 2023-05-29) and `srwiley/rasterx` (BSD-3-Clause, last push 2022-07-30) are pure Go. They need `golang.org/x/image` and `golang.org/x/net`. The oksvg README says that "many SVG elements will not be read". The project did not test the preset logos with it (the lab command was not permitted in this research).
- `tdewolff/canvas` (MIT, active) is pure Go, but its `go.mod` has approximately 25 direct requirements, for example Fyne, Gio, OpenGL and go-chart.
- `kanrichan/resvg-go` runs resvg as WASM with no cgo, but it is GPL-3.0. The repository has no licence yet ("All rights reserved"), so a GPL dependency in the binary is not acceptable without a decision.
- `librsvg` through cgo needs C libraries, and the final image is distroless with only the binary (ADR 0018).

**Recommendation: option C**, with a required `logo.png` in the brand directory, as ADR 0023 foresaw. It is the only option that works in the lab sink and in all clients with no new dependency. Its cost is one more file for each brand author, and a PNG that can differ from the SVG. Rules that a start-up check can make: a valid PNG (`image.DecodeConfig`), 100 KiB or less, a fixed pixel height (for example 3 times the displayed height), and the same aspect ratio as the `viewBox` of `logo.svg`, with a small tolerance. Legibility on the brand colour stays a duty of the author, as for `logo.svg`. For the presets, a one-time conversion with a tool outside the repository is sufficient; the gallery review of the user checks the result.

Option D is a valid fallback only if the user does not accept a contract change. Limit it to `oksvg` and accept that a user SVG with unsupported features renders badly, or refuse such an SVG at start-up (which is a contract change too).

### Other parts of the brand in the email

- **Colours.** The email uses literal sRGB values of the brand tokens [S14, S15]. The data holder can convert each OKLCH value of `tokens.json` to hex at start-up; the gamut check already makes sure that each colour is in sRGB (ADR 0023). Use the pairs that the contrast test already checks: `brand`/`fg-on-brand` for the header band, `surface`/`fg`, `heading`, `primary`/`fg-on-primary` if the email has a highlighted OTP box. A link colour is not necessary, because the email has no link (see section 4).
- **Font.** The brand fonts are web fonts that the data holder serves; most clients do not load a web font. Recommendation: a system font stack (for example `Arial, Helvetica, sans-serif`), and the logo gives the brand look. Judgement.
- **Go templates.** `html/template` refuses a value with `(` or `)` in a CSS context and writes `ZgotmplZ`. Thus OKLCH values from data fail, and hex values pass. (Go `html/template` CSS value filter; check the behaviour in a test.)
- **Dark mode.** Some clients change the colours of an email in dark mode. No caniemail data was checked for this. A logo on a coloured header cell stays legible only if the client keeps the cell colour. This is a risk, not a requirement.

## 3. The sender

### Facts

- `display-name = phrase`, and `name-addr = [display-name] angle-addr`. [S7, 3.4] A phrase is words; the specials (for example `,`, `.`, `"`, `@`) are not in an atom, so a name with them needs a quoted-string. [S7, 3.2.3, 3.2.5]
- A non-ASCII display name uses an RFC 2047 encoded-word in the phrase. [S8, section 5(3)]
- Go `mail.Address.String` "formats the address as a valid RFC 5322 address", and renders a non-ASCII name "according to RFC 2047". `mail.ParseAddress` parses one address, also one with a display name. [S9]
- The "From:" field "SHOULD NOT contain any mailbox that does not belong to the author(s) of the message". [S7, 3.6.2]
- The Decision "Sender address of the email" says that "A display name from the brand is a later change, with the brand on the emails". (`DESIGN.md:1023`)

### Recommendation

- `From` is `mail.Address{Name: <brand name>, Address: <MAIL_FROM>}.String()`. A test gives a brand name with a comma, a quote, a non-ASCII character and a line break, and checks that the header is one valid line.
- The data holder parses `MAIL_FROM` with `mail.ParseAddress` at start-up, and refuses a value with a display name. Judgement: one source for the name, and an error at start-up and not at the first OTP.
- The brand directory still has no email address (`DESIGN.md:1023` stays).

## 4. Content rules

### Facts

- "Data holders and data recipients MUST state in consumer-facing interactions and communications that third parties do not need consumer passwords to access CDR data." [S1, `#consumer-experience_common-authentication-standards`, "Authentication: Passwords"; S3, 2AU.00.31] The OTP email is a consumer-facing communication. The current OTP email does not have this statement. (`internal/domain/otp.go:96-108`)
- Unwarranted friction includes "Offering additional or alternative services" and "Referencing or including other documents". [S1, "Authentication: Friction"]
- In OTP delivery, "data holders should avoid including links that could [cause] unnecessary friction or trigger phishing concerns". [S3, 2AU.00.36, about an invitation to install an app]
- The data holder "should use the Brand Name of the data recipient wherever the data recipient is referenced", and should not show the Software Product Name. [S3, 2AU.00.16]
- The OTP "MUST only be used for authentication for CDR based sharing". [S1, One Time Password Credential Requirements] The CDS gives no rule for the brand of the data holder in the OTP message. [S3]
- The footer text of the consumer pages: "<brand name> is not a real bank, this is just a demonstration. This project has no affiliation with the ACCC, DSB, any CDR participant or any real bank." The data holder gives it, a brand cannot change it, and it has no link. (`DESIGN.md:486`)

### Recommendations

- The email keeps its current facts and adds the password statement in both parts (for example "<recipient> does not need your password to get your data. Do not give your password to anyone.").
- The email has no link, no button, no offer and no reference to another document. The only image is the logo.
- The email has the same footer text as the pages, as text, with no link. The Text part has it too.
- The subject stays as it is, with no OTP. A brand name in the subject is optional; the `From` display name already gives the brand.
- The Text and HTML parts give the same information [S4]. The OTP is text in the HTML part, not an image, so that a consumer can copy it.
- Accessibility (judgement, from WCAG 2.2 [S19] for web content): `lang="en-AU"` on `html`, a `title`, `role="presentation"` on each layout table, the brand name as the `alt` of the logo (SC 1.1.1), contrast of 4.5:1 for each text pair (SC 1.4.3, with the pairs that the brand contrast test checks), no information in colour only (SC 1.4.1), and a readable size for the OTP. The plain text part is an alternative for clients that do not show HTML.
- Telemetry: no change. The email body and the addresses do not go into telemetry (`DESIGN.md:838`).

## 5. The operator console

### Facts

- The console is a staff tool (`DESIGN.md:713-716`). It uses the neutral lab brand and compact density (`DESIGN.md:762`; `web/packages/ui/DESIGN.md:55`).
- The console CSP permits same-origin images and styles only. (`pkg/platform/consolehttp/consolehttp.go:22`) The brand files are on a different origin (the public TLS listener), so the console cannot load them now. (`internal/root/api.go:86`)
- The SVG response CSP (`default-src 'none'; style-src 'unsafe-inline'; sandbox`) has no effect on an SVG in an `img` element of the console. (`internal/adapter/branddir/handler.go:9-11`; ADR 0023, Assets)
- The header of a consumer page shows the logo on the brand colour, and each logo is legible on its brand colour. (`web/packages/ui/DESIGN.md:119`)
- A brand changes colours, the font, the radius and the logo. (`web/packages/ui/DESIGN.md:54`) The preset brands have pill radii and, for one brand, a yellow primary colour with a contrast of 1.51:1 on white. (Preset design note)

### Options

- **A. The full brand stylesheet.** The console links `/brand/styles.css`. Simple (one `link`), but each colour, the font and the radius of the bank apply to a dense staff tool, and the console looks like a consumer page of the bank.
- **B. The brand in the header only.** The header shows the logo on the brand colour, and the rest of the console stays in the lab brand. The console needs the logo, and the values of the `brand` and `fg-on-brand` roles (and the accent line, if it is used).
- **C. The colours from `GET /api/brand`, applied by JavaScript through the CSSOM.** The CSP permits it (`web/packages/ui/DESIGN.md:142`), but the earlier research rejected tokens in the console API, and it adds a second path for the same values.

### Recommendation

Option B:

- The console listener serves `/brand/logo.svg`, `/brand/mark.svg` and a small stylesheet that sets only the header roles (for example `/brand/console.css`), through the console routes, so that the `Host` allowlist and the response headers of the console apply. The values come from the same tokens and generator as `/brand/styles.css`.
- `index.html` gets static links to that stylesheet and to `/brand/mark.svg` as the icon of the tab. The HTML stays the same for each brand, so the Decision "Source of the brand name in the UI" (`DESIGN.md:992`) stays true. Check that Vite keeps the absolute links as they are (for example with the `vite-ignore` attribute). This is a build risk to check, not a fact.
- `AppShell` gets an optional logo, which is a change to the design system and its tests. The Register console does not use it.
- The header shows the logo with the brand name as its `alt`, and a label that says that this is the operator console, so that the operator does not mistake it for a consumer page. The tab title keeps the brand name. With four instances (milestone 3), the header colour and the tab icon show which bank a tab is.
- The body, the tables, the controls and the status colours stay in the lab brand. Judgement: legibility of dense data, one look for all staff tools, and the text "The lab brand is neutral. The consoles use it." (`web/packages/ui/DESIGN.md:55`) changes only for the header.
- The console still shows no OTP and no email. The mail sink has its own UI (`DESIGN.md:733`).

## 6. Tests

| Item | Test | Source of the idiom |
|---|---|---|
| Content of the email | Domain tests: each fact is in the Text and the HTML part (OTP, recipient brand name, validity, CDR only, password statement, footer) | `internal/domain/otp_test.go` |
| MIME structure and headers | `smtpmail` tests with the fake SMTP server: the structure of section 2, `From` with the display name, a brand name with specials and line breaks, quoted-printable HTML with no line longer than 998 | `internal/adapter/smtpmail/smtpmail_test.go:33`, `:137` |
| Look of the email | Golden files of the HTML part for each preset brand, with `-update`; a browser test in `web/packages/ui` runs the axe checks on them, as for the consumer pages | `internal/adapter/consumerpages/golden_test.go`; `web/packages/ui/DESIGN.md:229` |
| `logo.png` rules (option C) | `branddir` tests with testdata directories, as for the SVG rules | `internal/adapter/branddir/branddir_test.go` |
| CDR content of the OTP email | No change: the conformance suite reads `Text` and the OTP pattern still matches. The brand is not CDR behaviour, so the suite does not test it | `conformance/dataholder/otp_test.go:70`; ADR 0019 |
| Console listener | Go tests: the brand routes on the console listener, the `Host` allowlist and the headers | `pkg/platform/consolehttp/consolehttp_test.go` |
| Console UI | Unit tests with MSW (header with logo and `alt`); `AppShell` component test with the logo; the e2e test of the header (`customers.spec.ts:13` changes, because the banner text then has the console label) with the axe checks, in `test:lab-web` | `web/apps/dataholder-console/src/router.test.tsx`; `web/apps/dataholder-console/e2e/customers.spec.ts` |
| Lab verification | Mailpit API: `From.Name` is the brand name, `Inline[0].ContentType` is `image/png`, the HTML has the `cid:` reference; the Mailpit page shows the email; `html-check` for information only | Section 2, lab test |

## 7. Risks and open questions for the requirements PR

| Item | Type |
|---|---|
| Remove "notice emails" from milestone 2, or add a new consumer email at the end of a consent (section 1) | Decision for the user |
| The logo in the email: option C (`logo.png` in the contract, ADR 0023 change) or option D (render in Go, new dependency) | Decision for the user |
| ADR 0023: change the file list in place, or a new ADR that supersedes that part | Decision; use the ADR skill |
| The console: header only (option B) or the full brand (option A) | Decision for the user |
| The password statement in the OTP email: a change to the current requirement at `DESIGN.md:425` | Decision; the CDS text supports it |
| Refuse a `MAIL_FROM` with a display name | Decision |
| A brand name with a line break: refuse it in `brand.json` too, or only encode it in the header | Decision; refusal is a contract change |
| Vite and the static absolute links in `index.html` | Build risk to check in the first ticket |
| Dark mode in clients can change the header cell colour, and the logo can then be less legible | Risk; the lab sink does not change colours |
| A PNG that does not agree with the SVG (option C) | Risk; the gallery review of the presets, and the aspect-ratio check |
| Mailpit update to v1.33.0 | Check that the part handler of S16 does not change the PNG behaviour |
| Size of each OTP email with the PNG (approximately the size of the PNG, base64 adds approximately 33 %) | Low risk; a limit of 100 KiB for the PNG keeps it small |
