# Online bank of a data holder, milestone 2: research

- Research date: 2026-11-02
- Applies to: CDS 1.36.0 (with the CX standards), CDR Rules Compilation No. 10, OWASP ASVS 5.0, OWASP Cheat Sheet Series (Session Management, CSRF Prevention), RFC 6265 and RFC 6265bis, WCAG 2.2
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 2 of the project "A consumer uses the online bank of each data holder": "Sign-up of a mock customer, sign-in in which any password works, and the accounts with balances and transactions." The project risk for this milestone: "How the consumer tells the sign-in with a password from the CDR authentication with no password." The requirements PR resolves it.
- Related notes: [Online bank: landing page and products](2026-11-02-online-bank-landing-and-products.md), [Web front ends](2026-10-22-web-front-ends.md), [CDS banking and common APIs](2026-10-06-cds-banking-and-common-apis.md), [CDS data holder InfoSec and consent](2026-10-06-cds-data-holder-infosec-and-consent.md). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | CDS 1.36.0: Consumer Experience, Common Authentication Standards ("Authentication: Passwords", "Authentication: Friction") | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards |
| S2 | CDS 1.36.0: Consumer Experience, Redirect to Web with One Time Password ("Authentication: Password Link", "Authentication: OTP Expiry") | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience |
| S3 | CDS 1.36.0: Credential Requirements: User Identifiers, One Time Password Credential Requirements, Restricted Credentials | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#restricted-credentials |
| S4 | CDS 1.36.0: Get Account Balance | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-balance |
| S5 | CDS 1.36.0: Get Transactions For Account | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account |
| S6 | CDR Rules, rules 1.10B, 4.5, 4.24, 9.3 | https://www.legislation.gov.au/F2020L00094/latest/text |
| S7 | OWASP ASVS 5.0, V7 Session Management (7.1.1, 7.1.2, 7.2.3, 7.2.4, 7.3.1, 7.3.2, 7.4.1) | https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x16-V7-Session-Management.md |
| S8 | OWASP Session Management Cheat Sheet | https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html |
| S9 | OWASP CSRF Prevention Cheat Sheet | https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html |
| S10 | RFC 6265, section 8.5 (Weak Confidentiality) | https://www.rfc-editor.org/rfc/rfc6265#section-8.5 |
| S11 | RFC 6265bis draft, section 4.1.3.2 (the `__Host-` prefix) | https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-rfc6265bis#section-4.1.3.2 |
| S12 | Go `net/http`: `CrossOriginProtection` | https://pkg.go.dev/net/http#CrossOriginProtection |
| S13 | MDN: `Clear-Site-Data` | https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Clear-Site-Data |
| S14 | WCAG 2.2: SC 2.2.1 Timing Adjustable, and its Understanding document | https://www.w3.org/WAI/WCAG22/Understanding/timing-adjustable.html |
| S15 | WCAG 2.2: SC 3.3.8 Accessible Authentication (Minimum), and its Understanding document | https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum.html |
| S16 | WCAG 2.2: SC 1.3.5 Identify Input Purpose | https://www.w3.org/TR/WCAG22/#identify-input-purpose |
| S17 | RFC 10017, section 6.1 and 7.1 (BFF, cookie sessions) | https://www.rfc-editor.org/rfc/rfc10017#section-6.1 |

## 1. What the repository has now

Facts:

- The data holder `DESIGN.md` puts "Sign-up, sign-in and the accounts of a consumer on the online bank site" out of scope now, and says that later changes add them (`services/dataholder/DESIGN.md:49`).
- The site rules now: "The site sets no cookie, and the browser gets no credential." (`DESIGN.md:891`). The decision "Cookies of the online bank site" says "None now. A later sign-in decides the session cookie, with a name that the CDR pages do not use." (`DESIGN.md:1167`).
- ADR 0014: "The browser gets no credential. In the lab, a console listener has no login." (`docs/adr/0014-…md:31`). It also says that a UI that is an OAuth client uses its service as a BFF, and that "A production service needs an identity for each operator, with a session" (`:32`, `:42`).
- The site listener uses `consolehttp.Settings.Register` (`internal/adapter/banksite/banksite.go:26-31`, `:53-63`). This gives the CSP, `nosniff`, `Referrer-Policy: no-referrer`, COOP, the `Host` allowlist, `http.CrossOriginProtection` and `Cache-Control: no-store` on `/api/` (`pkg/platform/consolehttp/consolehttp.go:22-25`, `:88-122`, `:135-142`). The package comment says "with no credential" (`consolehttp.go:1-2`). The listener uses TLS with the public certificate (`internal/root/banksite.go:33`). Container port 8085 (`internal/root/run.go:25`); host ports 8185, 8285, 8385, 8485 (`compose.yaml:147`, `:210`, `:273`, `:336`).
- The site API has only `GET /api/brand`, `GET /api/products` and `GET /api/products/{productId}` (`banksite.go:59-61`).
- No path of the site API contains a customer number or a different identifier of a consumer, and no span, metric or log contains a query string, a path value, a request body or a cookie value (`DESIGN.md`, Site API and Online bank site: Non-functional requirements).
- `pkg/platform/telemetry` records no HTTP request headers: its code that is not a test has no match for "header" or "cookie".

## 2. Sign-up

### Facts

- A customer is a person with a customer number, a last name, an optional first name, an email address and the time of its last change (`DESIGN.md:138`). "All customer and account data is invented. It does not copy the data of a real person." (`DESIGN.md`, Customers and accounts).
- The customer number is the user identifier of the consumer. The data holder creates it. It has 8 decimal digits, and it is unique in the brand (`DESIGN.md:139`).
- The CDS: a data holder "MUST request a user identifier that can uniquely identify the customer" and "MUST request a user identifier that is already known by the customer in the redirected page" [S3, User Identifiers].
- The email address is the channel for the OTP (`DESIGN.md:140`). The CDS: the OTP goes "through an existing channel or mechanism", which "MUST align to existing and preferred channels for the customer" [S3].
- `app.Customers.Create` validates the customer, sets the time, and tries a maximum of 3 random customer numbers (`internal/app/customers.go:54-56`, `:69-91`). `domain.NewCustomerNumber` makes 8 digits from a random value (`internal/domain/customer.go:16-24`). `Customer.Validate` checks `firstName` (100 characters at most), `lastName` (1 to 100) and the email address (`customer.go:104-115`). The operator interface and the console call this use case (`internal/adapter/console/creates.go:19`, `console.go:57`).
- Customers and accounts are rows in tables, not events (`internal/migrations/00006_create_customers_and_accounts.sql`; `internal/adapter/customerstore`). Only the arrangements and the client registrations have events (`DESIGN.md`, Decisions, "Views that are not in the console").
- ADR 0021: "Each field of an event type applies to each event of that type", and a field "whose values all rules and readers handle in the same way is permitted" (`docs/adr/0021-…md`, Decision).
- Rule 9.3 requires records of consents, of their withdrawals and of the method of a withdrawal [S6]. No rule that we found requires a record of how a customer became a customer of the bank.
- A customer is eligible only while the customer has an open account (`DESIGN.md`, Customers and accounts; rule 1.10B [S6]). When the customer has no open account, the consent page says that the customer cannot share data, and the consumer can only cancel (`DESIGN.md:483`).
- The lab configures fixed OTPs for two addresses, `openid@mail.test` and `fixed-otp@mail.test` (`compose.yaml:144`). The mail sink keeps each email and delivers none (`DESIGN.md`, Lab-only features).

### Options for the event model

| | R. The use case of the operator (`Customers.Create`) | E. A new event that records the channel |
|---|---|---|
| Same rules as the operator flow | By construction | Only when the event handler calls the same validation and number rules |
| Storage | The existing table | The first customer event, so a new aggregate type, stream and projection for customers (ADR 0003) |
| Channel of the sign-up | Not recorded | Recorded |
| Rule that reads the channel | None | None now |
| Cost | A new driving adapter on an existing use case | A large change to the customer model for a fact that no rule reads |

Recommendation: **R**. Reuse `app.Customers.Create` with no change. Record no channel. No rule or reader needs it now (ADR 0021 permits a field only when a reader uses it, and it gives no reason to add one). The console then shows each new customer, as for a customer of the operator.

### Options for an account at sign-up

| | N. No account | A. One account at sign-up |
|---|---|---|
| Scope | M3 opens accounts from products. M2 stays separate. | Takes part of M3: an account number, a BSB, and a product of the brand. |
| CDR eligibility of a new customer | Not eligible until the operator or M3 opens an account (`DESIGN.md:483`) | Eligible at once |
| Equality test and browser test | Accounts from the operator interface | Accounts from sign-up |

Recommendation: **N**. The accounts page of a customer with no account says so. The operator creates accounts for a customer of the site in the same way as now. M3 adds "Open an account". The consequence: a new customer cannot give a CDR consent until the customer has an open account. The requirements PR must say this, so that a demonstration starts with an account from the operator.

### Recommendations for the content

- The sign-up form asks for the first name (optional), the last name and the email address: the fields of the operator flow, with the same rules. It asks for no password: the bank keeps no password. It asks for no identity document and makes no identity check.
- The page says that the bank is a demonstration and that the consumer must give invented details only. It gives an example email address on a reserved domain, for example `name@example.test`.
- Decision for the requirements PR: refuse an email address that is not on a reserved domain (RFC 2606 and RFC 6761: `.test`, `.example`, `example.com`). This stops real addresses in the lab database. The cost: one more rule that the operator flow does not have, or a change of the operator flow too. Our recommendation: only the text on the page. The mail sink delivers no email, and the operator flow has no such rule.
- The site API gives field errors with a JSON Pointer, as the console API does. The site shows a message in plain language for each rule (design system: Content). Decision: the site API gives the plain-language text, or the UI maps each rule to its text. Recommendation: the site API gives it, so that the UI has no copy of the rules (the console rule, `DESIGN.md`, Operator console).
- After the sign-up, the site shows the new customer number on a page of its own, with a copy button and two statements: the customer signs in with this number, and an app that gets data under the CDR asks for this number and then for a one-time password. This makes the number "already known by the customer" [S3].
- Decision: the sign-up signs the customer in, or the customer then signs in. Recommendation: the sign-up starts a session (a new session ID, see section 3). One step less, and the page with the number still shows first.

## 3. Sign-in and the session

### Facts

- OWASP: "Session identifiers must have at least 64 bits of entropy" [S8]. ASVS 5.0 7.2.3: reference tokens are made by a CSPRNG and "possess at least 128 bits of entropy" [S7]. Go `crypto/rand.Text()` gives 128 bits; the CDR pages use `rand.Text() + rand.Text()` (`internal/adapter/consumerpages/consumerpages.go:426`).
- ASVS 7.2.4: a new session token at each authentication [S7]. OWASP: "The session ID regeneration is mandatory to prevent session fixation attacks" [S8].
- ASVS 7.1.1 and 7.1.2: document the inactivity timeout, the absolute maximum lifetime and the number of concurrent sessions [S7]. ASVS 7.3.1 and 7.3.2 require both timeouts [S7]. OWASP: "Common idle timeouts ranges are 2-5 minutes for high-value applications and 15-30 minutes for low risk applications" and "All sessions should implement an absolute timeout" [S8].
- ASVS 7.4.1: after a sign-out or an expiry, "the application disallows any further use of the session". Stateless tokens need more measures, for example a revocation list [S7]. OWASP: "the web application must invalidate the session at least on server side" [S8].
- OWASP marks the `__Host-` prefix "Recommended for session IDs", and says "Session cookies must explicitly set SameSite=Strict (preferred) or SameSite=Lax" [S8]. A `__Host-` cookie has `Secure`, no `Domain` and `Path=/` [S11].
- "Cookies do not provide isolation by port" [S10]. All lab ports are on `localhost` or `127.0.0.1`. Thus the session cookie of a site goes to the CDR pages (8x81), to the consoles (8x84) and to the sites of the other banks (8185, 8285, 8385, 8485). A cookie with the same name from a different port replaces it, because the name, the host and the path are the same.
- The CDR flow cookie is `__Host-cdr-flow`, `Secure`, `HttpOnly`, `SameSite=Strict`, with no expiry (`consumerpages.go:23-26`, `:420-438`). Its value is a browser token that the page keeps when it is present, so the four banks can share it.
- Each instance has a unique setting `DATA_HOLDER_BRAND_ID`, the `dataHolderBrandId` of the brand on the Register (`internal/root/config.go:85`; `compose.yaml:135`, `:203`, `:266`, `:329`).
- `Clear-Site-Data: "cookies"` removes "all cookies for the origin of the response URL" and "affects the entire registered domain" [S13]. In the lab, it removes the cookies of each lab UI on `localhost`.
- The authorisation server keeps its sessions in memory with a limit of 10,000 (`pkg/authserver/sessions.go:12-27`; `docs/production-gaps.md`, Shared libraries).
- RFC 10017 permits a server-side cookie session for an application on one domain with no OAuth [S17, section 7.1].

### Options for the session

| | M. Server-side, in memory | D. Server-side, in PostgreSQL | C. Signed cookie with no server state |
|---|---|---|---|
| Sign-out (ASVS 7.4.1) | Removes the entry | Removes the row | Needs a revocation list, so server state again |
| After a restart | The consumer signs in again | The session stays | The session stays |
| Cost | A map with a lock, a limit and an expiry | A migration, a store adapter and a removal job | A key, its storage, and a revocation list |
| Content in the browser | A random ID only | A random ID only | The customer number, signed (encrypted only with more code) |
| Fit with the lab | One instance of each bank | More than the lab needs | The customer number in a cookie that all lab ports get |

Recommendation: **M**. A random session ID of 128 bits or more, from `crypto/rand`, that names an entry in memory with the customer number, the time of the sign-in and the time of the last request. A limit of 10,000 sessions, as the authorisation server has. A full store refuses a new sign-in with 503. The Restart requirement (`DESIGN.md:949`) adds: "A session of the online bank site can end at a restart."

### Recommendations for the cookie

- Name: `__Host-bank-session-<DATA_HOLDER_BRAND_ID>`, for example `__Host-bank-session-d7385786-d8b1-4d4b-bb80-b21616df1a4d`. Reasons: the `__Host-` prefix [S8, S11]; a name that the CDR pages never use; and one name for each instance, so that a sign-in at one bank does not replace the session of a different bank in the same browser. A UUID has only characters that a cookie name permits. Alternative: one fixed name, `__Host-bank-session`. Then a consumer who uses two banks in one browser is signed out of the first. Decision for the requirements PR.
- Attributes: `Secure`, `HttpOnly`, `SameSite=Strict`, `Path=/`, no `Domain`, and no `Max-Age` or `Expires` (a cookie for the browser session; the server gives the timeouts). `SameSite=Strict` costs nothing for a SPA: the browser sends the cookie on each `fetch` of the same origin. In the lab, `SameSite` does not separate the lab ports, because they are the same site.
- The value is the session ID only. It is not the customer number.
- The CDR pages do not read the site session. A site session gives no step of the CDR authentication, and a CDR authentication gives no site session. Judgement: the CDR flow always asks for the customer number and the OTP (`DESIGN.md`, Consumer authentication).

### Recommendations for CSRF

- Keep `http.CrossOriginProtection` of `consolehttp` as the only CSRF control. It refuses a request that is not safe when `Sec-Fetch-Site` is not `same-origin` or `none`, and it compares `Origin` with `Host` when the browser sends no `Sec-Fetch-Site` [S12]. OWASP calls Fetch Metadata "the primary signal for CSRF protection" for modern browsers, with a fallback to the origin check [S9]. The browser support of the design system is modern browsers only (`web/packages/ui/DESIGN.md`, Browser support).
- No CSRF token. The same check covers the sign-in form, so login CSRF needs no pre-session [S9]. It also refuses a change from a different lab port, because a different port is a different origin.
- Defence in depth, optional: refuse a `POST` body that is not `application/json`. A cross-origin JSON request then needs a CORS preflight, which the site does not permit. `pkg/platform/jsonhttp` does not check the request `Content-Type` now.
- Sign-out and sign-in are `POST` (or `DELETE`) requests, not `GET`.

### Recommendations for the lifetime and the sign-out

- Inactivity timeout: 15 minutes. Absolute lifetime: 12 hours. Unlimited concurrent sessions for one customer. Record these values in the `DESIGN.md` (ASVS 7.1.1, 7.1.2 [S7]).
- WCAG 2.2 SC 2.2.1 applies: "if a time limit is included in order to address security concerns, it would be considered to have been set by the content" [S14]. The site meets SC 2.2.1 with "Extend": a dialog at least 1 minute before the inactivity timeout, with "Stay signed in", which the consumer can use at least ten times [S14]. The absolute lifetime of 12 hours does not meet the 20-hour exception. Options: an absolute lifetime of more than 20 hours (the 20-hour exception), or the "Essential" exception for a security limit, which the Understanding document permits only for "certain time limits" [S14]. Recommendation: an absolute lifetime of 24 hours, so that only the inactivity timeout needs the dialog. Decision for the requirements PR.
- Sign-out removes the session on the server and sends the cookie with `Max-Age=0`. Do not send `Clear-Site-Data: "cookies"`: in the lab it removes the cookies of each lab UI, also a CDR flow in progress [S13].
- After an expiry, an API request gives 401. The site goes to the sign-in page and says that the bank signed the consumer out.

### Recommendations for the "any password" rule

- The sign-in form has two fields: "Customer number" (`autocomplete="username"`, `inputmode="numeric"`) and "Password" (`type="password"`, `autocomplete="current-password"`). The password must not be empty. Paste and password managers work [S15, S16].
- The UI does not send the password. The site API takes the customer number only, and refuses a body with a field that it does not name (the console API rule). Thus no password of a person goes to the data holder, to its telemetry or to its database. A test checks the refusal. Alternative: the API accepts and ignores a `password` field. Decision for the requirements PR. Our recommendation is the first: a person can type a real password from habit.
- An unknown customer number gives a general message: "We could not sign you in. Check your customer number." Note: with any password, this message shows which numbers exist, as each sign-in form with no password check does. The CDR pages keep their rule to hide this (`DESIGN.md`, Consumer authentication). The site cannot hide it and still sign in with the number only.
- The sign-in page shows a note above the form, in a status style that is not the lock note of the CDR pages. Candidate text: "This is a demonstration bank. Any password works here. Do not use a real password." (see section 4 for the full text).
- No rate limit. The existing gap "No endpoint has a rate limit" (`docs/production-gaps.md:28`) covers the sign-in and the sign-up.

### Changes to ADR 0014

- ADR 0014 line 31 ("The browser gets no credential") applies to all web UIs. The site now gets a credential: an `HttpOnly` session cookie that names a server-side session. No token goes to a script, so the BFF reason of the ADR stays true [S17].
- Recommendation: amend ADR 0014 in place (a small edit), with the `write-adr` skill: "An operator console gets no credential. The online bank site of the data holder gets an `HttpOnly` cookie that names a session on the server, and no token. Its sign-in is lab-only." Also change the package comment of `consolehttp` ("with no credential") or move the session code to `banksite` only.

## 4. The sign-in and the CDR authentication

### Facts

- "Data holders and data recipients MUST state in consumer-facing interactions and communications that third parties do not need consumer passwords to access CDR data. The exact phrasing of this is at the discretion of the data holder and data recipient." [S1, Authentication: Passwords]
- "Data holders MUST NOT introduce unwarranted friction into the authentication process." Friction includes "Offering additional or alternative services" and "Referencing or including other documents" [S1, Authentication: Friction; rule 4.24 [S6]].
- "Data holders MUST NOT include forgotten details links in redirect screens." "Data holders MUST communicate the expiry period of the OTP" [S2].
- Restricted Credentials in CDS 1.36.0: authenticators "MUST NOT permit Memorised Secrets defined by [DigitalID-Accreditation] to achieve 'LoA 2' (as a single factor of authentication only)" [S3]. The text does not say "never ask for a password". The rule of the data holder that the CDR pages never ask for a password is stricter (`DESIGN.md`, Consumer authentication). The project cites Restricted Credentials for "never ask for a password". The requirements PR must cite S1 and S3 for what each says.
- The CDR pages have the title "Share your data - <bank>" (`consumerpages/templates/customer-number.html:1`, `otp.html:1`, `consent.html:1`), a "Secure" label, the data recipient in a hero, a step indicator, and a lock note: "<recipient> does not need your password to get your data. <bank> does not ask for your password on this page. Do not give your password to <recipient> or to a different person." (`templates/layout.html:11`, `:36`; `web/packages/ui/DESIGN.md:113`, Consumer pages).
- The site titles are "<page> - <brand> online banking", with no "Secure" label, no data recipient and no steps (`DESIGN.md`, Online bank site: Presentation). The site is on a different port (8x85) from the CDR pages (8x81).
- The CDS has no rule for the online bank site of a data holder. The CDS has no rule that a data holder must link to or from the CDR flow. The data recipient starts the flow with PAR [S3].

### Options for the design measures

| Measure | On the site | On the CDR pages |
|---|---|---|
| Different origin | Yes (port 8x85) | No change |
| Different page title | "Sign in - <brand> online banking" | No change ("Share your data - <bank>") |
| Different heading | "Sign in to online banking" | No change ("Share your data with <recipient>") |
| Different layout | Site header with navigation, no steps, no data recipient | No change |
| A note on the sign-in page | Yes: the demonstration, any password, and that a CDR app never needs it | — |
| A new note on the CDR pages about online banking | — | No. It refers to a different service, which the friction standard forbids [S1] |
| A link between the two | No | No. The flow has no links to other services [S1, S2] |

Recommendation:

- Keep the CDR pages as they are. Their lock note already says that the bank does not ask for a password on the page.
- Add a note to the sign-in page, above the form. Candidate text: "This is <brand> online banking. This is a demonstration bank: any password works here, and you must not use a real password. When an app asks for your <brand> data under the Consumer Data Right, you do not sign in here. <brand> asks only for your customer number and a one-time password, and never for your password."
- The site footer keeps the password statement of M1 on each page [S1].
- No link from the site to the CDR flow, and no link from the CDR pages to the site. M4 shows the consents on the dashboard, which is not a link into the flow.
- A browser test checks that the site sign-in page and the CDR customer number page have different titles and headings, and that only the site page has a password field.

## 5. Accounts, balances and transactions

### Facts

- The CDS gives `currentBalance`, `availableBalance` and `creditLimit` for Get Account Balance, and the data holder gives no `amortisedLimit`, `currency` or `purses` (`DESIGN.md`, Balances; [S4]).
- One domain function calculates a balance: `Account.Balance(TransactionTotals)` (`internal/domain/balance.go:15-21`). Get Account Balance (`internal/app/balances.go:78-98`) and the console (`internal/app/customerdetail.go:63-65`) both call it with `TransactionStore.Totals`.
- Get Transactions For Account: default `oldest-time` is `newest-time` minus 90 days, default `newest-time` is the time of the request, and the data holder gives no transaction more than 2 years old (`internal/app/consentdata.go:24-30`, `:125-162`; `DESIGN.md`, Transactions; [S5]). The console gives each transaction with no time limit (`customerdetail.go:95-122`).
- The `accountId` and `transactionId` of a consent are hashes for each software product (`DESIGN.md`, Resource identifiers). The internal account identifier is a random UUID that the operator interface and the console give (`DESIGN.md`, Customers and accounts).
- The pages for the consumer show an account number only as a masked number (`DESIGN.md:147`). The console shows it in full, because "The pages for the consumer mask it, because a different person can see the screen of a consumer" (`DESIGN.md`, Decisions, "Account numbers in the console"). Get Account Detail gives `bsb` and `accountNumber` to a data recipient (`DESIGN.md`, Get Accounts and Get Account Detail).
- `httpapi.RegisterSecure` takes a `ConsentTokenVerifier` port (`internal/adapter/httpapi/httpapi.go:47`; `consent.go:27-34`). Tests make an arrangement with `app.ArrangementType.Execute(..., domain.CreateArrangement{...})` (`internal/root/operator_test.go:505`). The M1 equality test calls the CDS mux and the site mux in one process (`internal/root/banksite_test.go:96`).
- ADR 0019 and the M1 rules: the conformance suite does not test the site (`DESIGN.md`, Online bank site: Non-functional requirements). The lab-web job in CI has no conformance data, so a browser test makes its own data.

### Recommendations

- **Same data by construction.** The site API reads through the app layer, with the same functions as the consent: `Account.Balance` with `TransactionStore.Totals` for a balance, and one function for the transaction query that `ConsentData.Transactions` and the site both use (the 2-year limit and the order). A new use case takes the customer number of the session in place of an arrangement ID. It gives `ErrAccountNotFound` for an account of a different customer.
- **Paths.** `GET /api/accounts`, `GET /api/accounts/{accountId}` and `GET /api/accounts/{accountId}/transactions`, with the internal account ID. The session gives the customer, so no path has a customer number (the M1 rule).
- **Field names.** The CDS names (`currentBalance`, `availableBalance`, `creditLimit`, and the fields of `BankingTransactionV2`), as for the products. A test then compares the values with no conversion.
- **Accounts page.** For each account: the display name, the product name, the product category label, the masked number, the current balance and the available balance. Open accounts first, then a "Closed accounts" group. Decision: the account detail shows the BSB and the account number in full, or masked. Recommendation: in full on the detail only, so that the customer can give the number for a payment, as Get Account Detail gives it. This changes the scope of the rule at `DESIGN.md:147` to the CDR pages.
- **Transactions.** The period that a consent can share: from 2 years before the time of the request to the time of the request, newest first, with pages (for example 25). For each transaction: the effective date, the description, the amount with its sign, and "Pending" for a pending transaction. The other fields that have a value (type label, reference, merchant, biller, `extendedData`) show in the detail of the row. Alternative: the CDS default of 90 days, with a control to show older transactions. Decision for the requirements PR.
- **Equality test (Go, in `internal/root`).** Create a customer, two accounts (one with a credit limit) and transactions of each type and status, also one older than 2 years and one with `extendedData`, with the stores. Make an arrangement for one account with `CreateArrangement`. Give `httpapi.RegisterSecure` a stub `ConsentTokenVerifier` that gives that arrangement. Sign in to the site API in the same process. Compare Get Account Balance with the balance of the site, and Get Transactions For Account (`oldest-time` 2 years before the fixed time, all pages) with the transactions of the site, field by field, except `accountId` and `transactionId`, which are different by design. No full CDR flow is necessary.
- **Browser test (lab-web).** Sign up, see the customer number, sign out, sign in with the number and any password, and see "no accounts". Then create an account and a transaction with the operator interface, and see them on the site. No consent in the browser test.

## 6. Telemetry

Facts:

- ADR 0006: no CDR data and no personal information in telemetry, also mocked data; route patterns, not raw URLs (`docs/adr/0006-telemetry.md`, Sensitive data). The data holder telemetry has no customer number, name, email address, account data or transaction data (`DESIGN.md:950`). The project adds the customer number and the name.
- The site rules of M1 already forbid a query string, a path value, a request body and a cookie value in telemetry (`DESIGN.md`, Online bank site: Non-functional requirements).

Recommendations:

- Add the session ID to the list of `DESIGN.md:950`: no session ID and no password in telemetry. Also no customer number in a log line of the sign-in, the sign-up or a session expiry.
- The site API responses contain the customer number, the name and the account data. The response bodies are not in telemetry, so this needs no new rule.
- Extend `TestTheBankSitePutsNoQueryStringAndNoPathValueInTheSpansMetricsOrLogs` (`internal/root/banksite_test.go:207`) to a sign-up, a sign-in with a password, the accounts, the transactions and a sign-out. Check that no span, metric or log contains the customer number, the names, the email address, the BSB, the account number, the account ID, a description of a transaction, the session ID or the password. This is the pattern of `consoletelemetry_test.go:206`.

## 7. Lab-only gating and the production gap

Facts:

- `BANK_SITE_ENABLED` turns the listener on, off by default (`banksite.go:26-31`). The site is lab-only, and real components must not depend on it (`DESIGN.md`, Online bank site; Lab-only features).
- The test of `docs/production-gaps.md`: "An entry is a choice that we made because it is easy, and that a production service could not operate with. A function that a project put out of scope is not a gap." The project puts "Real authentication of a consumer, and the security of a production online bank" out of scope, but it also requires an entry for the any-password sign-in.

Recommendations:

- No new setting. The sign-in is part of the site, which `BANK_SITE_ENABLED` already gates. A second setting for the password rule adds a mode that no deployment needs.
- The entry meets the test if it states the shortcut as a choice: a password field that checks nothing is a choice for ease, and a production bank cannot operate with it. Candidate entry for the Data holder table:

| Shortcut | Why it is acceptable in the lab | What a production service needs | Where |
|---|---|---|---|
| The online bank site signs in a consumer with the customer number only. Its password field accepts any value, and the site does not send the password to the data holder. The sessions are in the memory of the process. | The site is off by default, all customer data of the lab is invented, the sign-in page says that any password works, and the session cookie is `HttpOnly` and names a server-side session. | Real authentication of the consumer: a credential that the bank verifies, multi-factor authentication, limits on failed sign-ins, and sessions that all instances share and that stay after a restart. | `adapter/banksite` of the data holder |

- The tension between the project and the test of the file is for the requirements PR to resolve. If the entry stays, its text says why it is a gap and not only out of scope.

## 8. Risks and decisions for the requirements PR

| Item | Notes | Recommendation |
|---|---|---|
| Event model of a sign-up | Customers are rows, not events. No rule reads a channel. | Reuse `Customers.Create`. No event, no channel field. |
| Account at sign-up | M3 opens accounts. A customer with no open account cannot consent. | No account. The operator adds accounts until M3. Say so in the requirements. |
| Email domain at sign-up | A person can type a real address. The mail sink delivers nothing. | Text on the page only. |
| Plain-language field errors | The operator texts name the fields in camelCase. | The site API gives the consumer text. |
| Sign-in after sign-up | One step less, or the consumer practises the number. | The sign-up starts a session, then the number page shows. |
| Session store | Memory, PostgreSQL or a signed cookie. | Memory, limit 10,000, lost at a restart. |
| Cookie name | All lab ports share `localhost`, so the same name at four banks replaces the session. | `__Host-bank-session-<DATA_HOLDER_BRAND_ID>`. |
| Password in the request | Sent and ignored, or not sent. | Not sent. The API refuses a `password` field. |
| Timeouts and SC 2.2.1 | A security time limit is a content time limit. | 15 minutes inactivity with a "Stay signed in" dialog; 24 hours absolute. |
| CSRF | Fetch Metadata and the origin check are in `consolehttp`. | No token. Optional `Content-Type` check. |
| ADR 0014 "The browser gets no credential" | The site needs a session cookie. | Amend ADR 0014 in place. |
| Restricted Credentials citation | The CDS text forbids a memorised secret as a single factor for LoA 2. It does not say "never ask for a password". | Cite S1 and S3 for what each says. |
| Note on the CDR pages | A sentence about online banking refers to a different service. | No change to the CDR pages. A note on the site sign-in page. |
| Full account number on the site | `DESIGN.md:147` masks it on each consumer page. | Full on the account detail only; masked elsewhere. |
| Transaction period on the site | 90 days (CDS default) or 2 years (the shared period). | 2 years, newest first, with pages. |
| Enumeration on the site sign-in | With any password, the sign-in shows which numbers exist. | Accept for the lab. The CDR pages keep their rule. |
| Production-gaps test | The project requires an entry, and the file excludes functions out of scope. | Add the entry, with the text above. |
| New names | Cookie name, API paths, timeouts. | Propose them in the requirements PR before use (AGENTS.md). |

## Main findings

- Customers and accounts are rows in tables, not events. A sign-up can reuse `app.Customers.Create` and get the customer number rules of the operator flow by construction. No rule needs a record of the channel.
- A customer with no open account cannot give a CDR consent. With no account at sign-up, a new customer of the site needs an account from the operator until M3.
- Cookies are not isolated by port. A session cookie with one name at all four banks on `localhost` replaces the session of a different bank. A name with the brand ID of the instance stops this.
- A server-side session in memory meets ASVS 7.2.3, 7.2.4 and 7.4.1 with the least code. A signed cookie needs a revocation list for a sign-out.
- `http.CrossOriginProtection`, which the site listener already has, is the CSRF control that OWASP calls primary for modern browsers. It also covers login CSRF. No token is necessary.
- WCAG 2.2 SC 2.2.1 applies to a session timeout for security. The inactivity timeout needs a warning with a "Stay signed in" action.
- `Clear-Site-Data: "cookies"` removes the cookies of each lab UI on `localhost`, also a CDR flow in progress. The sign-out must not send it.
- The CDS 1.36.0 Restricted Credentials text forbids a memorised secret as the single factor for LoA 2. It does not say "never ask for a password". The no-password rule of the CDR pages is a stricter rule of the data holder.
- The friction standard forbids references to other services on the CDR pages. The separation goes on the site sign-in page, not on the CDR pages.
- ADR 0014 says that the browser gets no credential. The site session cookie needs an amendment of ADR 0014.
- One domain function calculates each balance, and one query gives the transactions of a consent. A site use case on the same functions gives the same data by construction. A Go test with a stub consent token verifier proves the equality with no CDR flow.
- If the UI does not send the password, no password of a person reaches the data holder or its telemetry.
