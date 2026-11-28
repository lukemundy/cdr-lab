# Data recipient, milestone 1: sign-up and sign-in research

- Research date: 2026-11-28
- Applies to: CDS 1.36.0, CDR Rules Compilation No. 10 (F2025C00572), Privacy Act 1988 Compilation No. 104 (C2026C00227), OWASP ASVS 5.0, NIST SP 800-63B-4, RFC 6265, draft-ietf-httpbis-rfc6265bis-22, WCAG 2.2, Telecommunications Numbering Plan 2025 (F2025L00409)
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 1 of the project "A consumer can connect their bank accounts to a mock data recipient in the lab" (P-CDR-13): "A consumer signs up, verifies their email address, gives their details and signs in to the default data recipient brand." The project risk for this milestone: "Can the session cookie of the data recipient safely share the host `localhost`?" The requirements PR resolves it.
- Related notes: [Web front ends](2026-10-22-web-front-ends.md), [Online bank: landing page and products](2026-11-02-online-bank-landing-and-products.md), [Online bank: sign-up, sign-in and accounts](2026-11-02-online-bank-sign-in-and-accounts.md), [Branded data holder instances](2026-11-01-branded-data-holder-instances.md). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | OWASP ASVS 5.0: V6 Authentication, V7 Session Management, V11 Cryptography, Appendix C | https://github.com/OWASP/ASVS/tree/master/5.0/en |
| S2 | NIST SP 800-63B-4, section 3.1.1.2 (passwords), section 3.2.2 (rate limits), section 3.1.3.1 (out-of-band) | https://pages.nist.gov/800-63-4/sp800-63b/authenticators/ |
| S3 | OWASP Password Storage Cheat Sheet, "Argon2id" | https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html |
| S4 | OWASP Forgot Password Cheat Sheet | https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html |
| S5 | OWASP Authentication Cheat Sheet, "Authentication and Error Messages", "Implement Proper Password Strength Controls" | https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html |
| S6 | RFC 6265, section 8.5 | https://www.rfc-editor.org/rfc/rfc6265#section-8.5 |
| S7 | draft-ietf-httpbis-rfc6265bis-22, sections 1 and 4.1.3.2 | https://datatracker.ietf.org/doc/draft-ietf-httpbis-rfc6265bis/ |
| S8 | WHATWG HTML: same site (7.1.1.1), autofill tokens (4.10.19.7) | https://html.spec.whatwg.org/ |
| S9 | Telecommunications Numbering Plan 2025, Schedule 4 | https://www.legislation.gov.au/F2025L00409/latest/text |
| S10 | Australia Post, addressing guidelines | https://auspost.com.au/business/marketing-and-communications/access-data-and-insights/address-data/addressing-guidelines |
| S11 | W3C, "Personal names around the world" | https://www.w3.org/International/questions/qa-personal-names |
| S12 | Privacy Act 1988: APP 1, APP 3, APP 5, APP 11 | https://www.legislation.gov.au/C2004A03712/latest/text |
| S13 | CDR Rules: rule 1.10B, rule 7.2, Schedule 2 clause 2.2 | https://www.legislation.gov.au/F2020L00094/latest/text |
| S14 | CDS 1.36.0 CX standards (git tag `v1.36.0`, `slate/source/includes/cx_standards/`) | https://github.com/ConsumerDataStandardsAustralia/standards |
| S15 | WCAG 2.2: SC 1.3.5, 2.2.1, 3.3.7, 3.3.8 | https://www.w3.org/TR/WCAG22/ |
| S16 | `golang.org/x/crypto/argon2` v0.59.0 | https://pkg.go.dev/golang.org/x/crypto/argon2 |

## 1. What the repository has now

### Facts

- A service has `main` → `lifecycle.Run` → `root.Run`, with the modes `serve`, `api` and `worker`, and the subcommands `migrate` and `seed` (`services/dataholder/internal/root/run.go:38-95`).
- [ADR 0025](../adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md) names the listeners internal, public, mTLS and "online bank site listener, on a data holder only" (lines 45-49). It names no listener for the web app of a data recipient.
- [ADR 0014](../adr/0014-web-uis-as-embedded-react-single-page-applications.md) permits a session cookie for a consumer UI with a sign-in (line 32). It says that a data recipient UI uses its service as a backend for frontend (line 34).
- [ADR 0003](../adr/0003-event-sourcing.md) gives event sourcing for participant state. The users of a data recipient and their details are consumer data, not participant state.
- The design system has one light theme. A dark theme is a non-goal (`web/packages/ui/design/README.md:51`). [ADR 0016](../adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md) says that a dark theme needs a second set of semantic values and contrast checks (lines 39 and 50).
- The data holder brand directory has generic parts with no data holder types: the tokens to `--ds-*` CSS, the SVG and PNG checks, the strict JSON decoder and the `/brand/*` handler (`services/dataholder/internal/adapter/branddir/`). The other parts are data holder parts: `bsb`, `layout`, `products.json` and the photos. Web tests read the preset directory by path (`web/packages/tokens/src/presets.node.test.ts:10`, `web/packages/ui/src/otp-email.test.ts:7-24`).
- The online bank site has a session cookie `__Host-bank-session-<brand ID>` with `Path=/`, `Secure`, `HttpOnly` and `SameSite=Strict`, a new session ID at each sign-in, and an in-memory store of 10,000 sessions for 24 hours (`services/dataholder/internal/adapter/banksite/session.go:40-106`, `sessions.go:11-158`). Its sign-in checks no password.
- `pkg/platform/consolehttp` gives the `Host` allowlist, `http.CrossOriginProtection`, the CSP and the SPA fallback (`consolehttp.go:23-143`). Its documentation says that the UI has no credential and that a production service has no such UI (line 35).
- The SMTP mailer of the data holder builds a `multipart/alternative` message with an inline `logo.png` (`services/dataholder/internal/adapter/smtpmail/smtpmail.go:159-230`). Its text, template and demonstration notice are data holder text.
- No browser test reads the mail sink. Only `conformance/dataholder/otp_test.go` reads it, through `MAIL_SINK_URL`.
- No code in the repository hashes a password. `golang.org/x/crypto` is not in `go.mod`.
- "Add a service" in `AGENTS.md` does not list these files, which a new service also changes: `cli/cdrlab/cmd/cdrlab/lab.go:47-50` (the release images), `.github/image-inputs.json`, `lab/certificates.json` and the service lists in `.github/workflows/ci.yml` (lines 29-30, 93-105, 230-237, 509-513, 531 and 649-654).
- The glossary gives "data recipient" and forbids "ADR" (`docs/spec/glossary.md:58-62`). It has no term for the record of a consumer at a data recipient, the counterpart of "customer" at a data holder.

## 2. Passwords

### Facts

- ASVS 5.0 (S1):
  - 6.2.1: 8 characters at least, and 15 characters strongly recommended.
  - 6.2.9: at least 64 characters permitted.
  - 6.2.4: a check against the top 3000 passwords at least. 6.2.12: a check against breached passwords.
  - 6.2.5: no composition rules. 6.2.7: paste and password managers permitted. 6.2.10: no forced periodic change.
  - 6.3.3 (L2): MFA, or a documented reason with other controls.
  - 11.4.2 and Appendix C: an approved password hash. The argon2id minimums are t=1 with m≥46 MiB, t=2 with m≥19 MiB, or t≥3 with m≥12 MiB, all with p=1.
- NIST SP 800-63B-4 (S2):
  - A password that is the single factor SHALL have 15 characters at least.
  - The maximum length SHOULD be 64 characters or more.
  - Each Unicode code point counts as one character. NFC normalisation SHOULD occur before the hash.
  - No composition rules and no periodic change. A change SHALL occur when there is evidence of compromise.
  - The verifier SHALL compare the full password with a blocklist and give the reason for a refusal.
  - Rate limits SHALL permit 100 consecutive failures at most (section 3.2.2).
  - "Email SHALL NOT be used for out-of-band authentication" (section 3.1.3.1).
- The OWASP Password Storage Cheat Sheet gives argon2id with m=19 MiB, t=2 and p=1 at least (S3).
- `golang.org/x/crypto/argon2` gives only `IDKey`. It gives no PHC string format and no verify function (S16).
- No maintained Go library for an offline blocklist was found. The OWASP Authentication Cheat Sheet names the Pwned Passwords download and the SecLists lists as offline sources (S5). The licences of these lists were not confirmed.
- Schedule 2, clause 2.2, control (h) of the CDR Rules: "Strong authentication mechanisms are enforced prior to allowing users to access systems within the CDR data environment, including, but not limited to, general security requirements relating to password complexity, account lockout, password history, and password ageing." (S13)

### Recommendations

- A minimum of 15 characters and a maximum of 128 characters, after NFC normalisation, because the password is the single factor.
- argon2id with m=19 MiB, t=2 and p=1, a random salt of 16 bytes and a PHC string that the service writes and reads.
- A blocklist in the binary: a list of common passwords with a licence that permits distribution, and context words (the brand name, "cdr", "lab", the local part of the email address).
- A new dependency on `golang.org/x/crypto`, with the usual review. `crypto/pbkdf2` in the standard library is the alternative, but ASVS and OWASP prefer argon2id.
- Control (h) is about the users of the CDR data environment, in our interpretation: staff, not consumers. Its "complexity" and "ageing" disagree with ASVS and NIST. A judgement in the specification records that the data recipient follows ASVS and NIST for consumers.
- No MFA is a lab shortcut against ASVS 6.3.3. It goes into `docs/production-gaps.md`.

## 3. Email verification and password reset

### Facts

- The OWASP Forgot Password Cheat Sheet (S4):
  - The same message, at the same time, for an account that exists and for an account that does not.
  - A token from a CSPRNG, for one user, single use, with an expiry, stored securely.
  - No reset URL from the `Host` header. `Referrer-Policy: no-referrer` on the reset page.
  - No sign-in after the reset. The other sessions end, or the user chooses.
  - No account lock after a reset request.
- The OWASP Authentication Cheat Sheet gives the correct sign-up message: "A link to activate your account has been emailed to the address provided." The incorrect message: "This user ID is already in use." (S5)
- ASVS 5.0 (S1): 6.4.1 (activation codes expire after a short time or after use), 6.5.1 (single use), 6.5.2 (a hash of a secret of 112 bits or more), 6.5.3 (CSPRNG), 6.5.5 (an out-of-band code is valid for 10 minutes at most), 7.4.3 (an option to end the other sessions after a change of the password) and 6.3.7 (a notice to the user after a reset).

### Recommendations

- A link with a token of 256 bits, stored as a SHA-256 hash. The verification link is valid for 24 hours. The reset link is valid for 30 minutes. Each link is single use. A new link makes the earlier links of the same kind invalid.
- A sign-up with an email address that has an account gives the same response. The email to that address says that an account exists and gives a reset link.
- A reset ends all sessions of the user, does not sign the user in, and sends an email that the password changed.
- The public base URI of the brand comes from the configuration, not from the request.

## 4. Session cookie on the shared host

### Facts

- "Cookies do not provide isolation by port." A server "SHOULD NOT" run mutually distrusting services on different ports of one host and keep security-sensitive information in cookies (S6, section 8.5).
- A `__Host-` cookie has `Secure`, `Path=/` and no `Domain`. Ports are the only part of the origin that `__Host-` ignores (S7, section 4.1.3.2). The draft is in the RFC Editor queue.
- A site is the scheme and the host. "Same site" ignores the port (S8, 7.1.1.1). Thus `SameSite` does not separate the lab UIs on `localhost`.
- ASVS 3.5.4 (L2): separate applications have different host names (S1).
- The production gaps already have an entry for the shared host `localhost` (`docs/production-gaps.md`, "The online bank sites, the consoles and the CDR pages of the lab share the host `localhost`").

### Recommendations

- A `__Host-` cookie with a name for each brand, for example `__Host-recipient-session-<brand>`, with `Secure`, `HttpOnly`, `SameSite=Lax` and `Path=/`. `Lax`, not `Strict`, because the redirect back from the bank in milestone 2 is a top-level navigation from a different port, and `Strict` is not a separation on one host.
- The session state stays on the server. `http.CrossOriginProtection` refuses requests from a different origin, also from a different port.
- The existing production gap gets the data recipient in its text. A cluster deployment with a host name for each app meets ASVS 3.5.4.

## 5. User details

### Facts

- Australian mobile numbers start with 04 or 05 and have 10 digits (S9, Schedule 4). The allocation of 05 numbers was not confirmed.
- Australia Post publishes no table of postcodes for each state. It assigns postcodes on operational grounds and can change them (S10). Some postcodes cross a state border.
- Australia Post: line 2 of an address is "the street number and name, or PO Box". Line 3 is "locality or suburb, state and postcode" (S10).
- An eligible CDR consumer is "an individual who is 18 years of age or older" (S13, rule 1.10B(1)(a)(i)). No source gives a time zone for the age.
- "Don't require that people supply a family name." Names can contain spaces, hyphens and apostrophes (S11).

### Recommendations

- Mobile: accept `04`, `05`, `+614` and `+615` forms with spaces. Store the E.164 form.
- Address: unit (optional), street number, street name, suburb, state (one of 8), postcode (4 digits). No PO Box. A postcode that does not agree with the first digit of the state gives a warning, not a refusal.
- Date of birth: compare dates, not times, in `Australia/Sydney`. A birthday on 29 February is 1 March in a year that is not a leap year.
- Name: given names (required), family name (optional with "I have only one name"), Unicode, a large maximum length.

## 6. Privacy and CDR duties of the sign-up

### Facts

- APP 3.2: an organisation "must not collect personal information ... unless the information is reasonably necessary for one or more of the entity's functions or activities." (S12)
- APP 5.1: at or before the collection, the entity takes reasonable steps to notify the individual of the matters of APP 5.2, for example the purposes and the consequences if it does not collect the information. (S12)
- APP 1.3: a "clearly expressed and up-to-date" APP privacy policy. (S12)
- Rule 7.2(2): the CDR policy is a document that is "distinct from any of the CDR entity's privacy policies". Rule 7.2(8): it is "readily available through each online service" in which the entity deals with CDR consumers. (S13)
- No CDR rule was found for the user accounts of a data recipient or its sign-up.
- CDS 1.36.0, CX standards, "Authentication: Passwords": "Data holders and data recipients MUST state in consumer-facing interactions and communications that third parties do not need consumer passwords to access CDR data." (S14)
- CDS 1.36.0, CX standards, "Consent: Redirection": data recipients "MUST notify consumers of redirection prior to authentication." This applies in milestone 2. (S14)

### Recommendations

- A collection notice on the details form (APP 5). A privacy policy and a CDR policy, as two documents, linked from each page of the app.
- The date of birth has a purpose: the check of rule 1.10B.
- The password statement of the CDS on the landing page, the sign-up page and each email.

## 7. Accessibility

### Facts

- SC 1.3.5: the purpose of each input is programmatically determinable. The tokens are `email`, `new-password`, `current-password`, `one-time-code`, `given-name`, `family-name`, `bday`, `tel`, `address-line1`, `address-line2`, `address-level2`, `address-level1` and `postal-code` (S8, S15).
- SC 2.2.1: the user can turn off, adjust or extend a time limit, unless the limit is essential or longer than 20 hours (S15).
- SC 3.3.7: the user does not enter the same information again in one process, except for security (S15).
- SC 3.3.8: no cognitive function test without an alternative. Password managers and copy and paste are such mechanisms (S15).

### Recommendations

- No block of paste in any field. A link, not a code, is the main path for verification.
- A warning before the session ends, with an extension.

## 8. Risks and decisions for the requirements PR

| Item | Options | Recommendation |
|---|---|---|
| The dark brand `basalt` and the light-only design system | A dark theme in the design system (a change to ADR 0016). A light `basalt` with dark accents. A different second brand. | A decision of the repository owner before milestone 3. Milestone 1 needs only `sorrel`. |
| The listener of the web app | A new listener in ADR 0025 for the web app of a data recipient. The web app on the public listener. | A change to ADR 0025 that adds a web app listener on a data recipient. The CDR endpoints of later milestones (JWKS, arrangement revocation) then have no session cookie and no SPA fallback on their listener. |
| The brand directory | Copy the generic parts. Move them to `pkg/`. | Move all generic parts to `pkg/` at the first user that is not the data holder, with the web tests that read the presets. |
| Users and details outside the event store | Plain tables. Events with crypto-shredding. | Plain tables, with a short ADR or a judgement that ADR 0003 is for participant state and consents only. |
| A glossary term for the record of a consumer at a data recipient | "User" (now a lab term only). "Member". "Account" (conflicts with a bank account). | Ask the repository owner. |
| `consolehttp` for a UI with a credential | Use it as it is. Change its documentation. A new package. | Change its documentation, because the mechanisms (host allowlist, CSRF, CSP) are the same. |
| A browser test that reads the mail sink | A Playwright helper on the Mailpit API. | Add it in milestone 1, because sign-up and reset need it. |

## Main findings

- No CDR rule applies to the sign-up of a data recipient. The Privacy Act APPs 1, 3, 5 and 11 apply, and rule 7.2 requires a separate CDR policy in each online service.
- NIST SP 800-63B-4 requires 15 characters at least for a password that is the single factor. ASVS 5.0 requires 8 and strongly recommends 15.
- Schedule 2 control (h) of the CDR Rules names password complexity and ageing. ASVS 5.0 and NIST SP 800-63B-4 forbid both.
- The repository hashes no passwords. argon2id needs `golang.org/x/crypto`.
- No maintained Go library gives an offline blocklist of passwords.
- `SameSite` and `__Host-` do not separate the lab UIs on `localhost`, because a site ignores the port.
- The design system has no dark theme, so a dark brand is a change to ADR 0016.
- ADR 0025 has no listener for the web app of a data recipient.
- The generic parts of the data holder brand directory can move to `pkg/`.
- No browser test reads the mail sink.
- The CDS requires the statement that third parties do not need consumer passwords also in the app of a data recipient.
