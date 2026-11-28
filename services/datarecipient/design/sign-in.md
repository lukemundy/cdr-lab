# Mock data recipient: Sign-in

This area gives the sign-in, the sign-out and the session of a member, and the password reset.

## Sign-in and sign-out

- <a id="dr-0400"></a>**DR-0400** The sign-in page asks for the email address and the password. ([WCAG 2.2, SC 3.3.8][wcag-3-3-8])
- <a id="dr-0401"></a>**DR-0401** The email address field has `autocomplete="username"`. The password field has `type="password"` and `autocomplete="current-password"`, and a control that shows the password as text. ([WCAG 2.2, SC 1.3.5][wcag-1-3-5]; [NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords])
- <a id="dr-0402"></a>**DR-0402** The consumer can paste into each field of the sign-in page, and a password manager can fill each field. ([WCAG 2.2, SC 3.3.8][wcag-3-3-8])
- <a id="dr-0403"></a>**DR-0403** The data recipient normalises the password to NFC before it compares it with the hash. ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords])
- <a id="dr-0404"></a>**DR-0404** When the email address is the address of a member and the password agrees with the hash of the member, the data recipient starts a session for that member. ([OWASP ASVS 5.0, 6.3.1][asvs-v6])
- <a id="dr-0405"></a>**DR-0405** When no member has the email address, or the password does not agree, the data recipient refuses the sign-in. The page says "We could not sign you in. Check your email address and password. If you signed up, confirm your email address first". ([OWASP Authentication Cheat Sheet][owasp-authentication]; [Judgement: one refusal that hides the members](#dr-0450))
- <a id="dr-0406"></a>**DR-0406** The refusal for an email address with no member is the same as the refusal for an incorrect password, in its status, its body and its time. For example, the data recipient calculates a hash also when no member has the address. ([OWASP ASVS 5.0, 6.3.8][asvs-v6]; [Judgement: one refusal that hides the members](#dr-0450))
- <a id="dr-0407"></a>**DR-0407** The data recipient does not lock a membership after failed sign-ins. ([OWASP Authentication Cheat Sheet][owasp-authentication]; [Judgement: no limit of failed sign-ins in the lab](#dr-0451))
- <a id="dr-0408"></a>**DR-0408** The sign-in page has a link to the password reset request page. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0409"></a>**DR-0409** After a successful sign-in, the web app shows the home page, or the member details page when the member has no member details. ([the member details before each other page](sign-up.md#dr-0321))
- <a id="dr-0410"></a>**DR-0410** A sign-out ends the session. After a sign-out, the web app shows the landing page. ([OWASP ASVS 5.0, 7.4.1][asvs-v7])
- <a id="dr-0411"></a>**DR-0411** Each operation of the API that changes state is not a `GET` request. For example, the sign-in and the save of the member details are not `GET` requests. ([the refusal of cross-origin changes](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0005); [OWASP CSRF Prevention Cheat Sheet][owasp-csrf])
- <a id="dr-0412"></a>**DR-0412** Each page and each operation of the API that shows or changes the data of a member needs a session. When the browser has no valid session, the web app shows the sign-in page. ([OWASP ASVS 5.0, 8.3.1][asvs-v8])
- <a id="dr-0413"></a>**DR-0413** When a session ended, the sign-in page says that the data recipient signed the member out. ([the session-ended type](../../../pkg/platform/problem/DESIGN.md#prob-0010))

## Session

- <a id="dr-0420"></a>**DR-0420** The data recipient keeps each session on the server. The session cookie has only the session ID. ([RFC 10017, section 7.1][rfc10017-cookie-session]; [Judgement: one session cookie for a consumer UI with a sign-in (ADR 0014)][adr-0014])
- <a id="dr-0421"></a>**DR-0421** The session ID is a random value from a cryptographic random source, with at least 128 bits of entropy. ([OWASP ASVS 5.0, 7.2.3][asvs-v7])
- <a id="dr-0422"></a>**DR-0422** Each sign-in starts a new session with a new session ID. When the request has a session cookie, the session of that cookie ends. A sign-in that fails keeps that session. ([OWASP ASVS 5.0, 7.2.4][asvs-v7]; [OWASP Session Management Cheat Sheet][owasp-session])
- <a id="dr-0423"></a>**DR-0423** A session ends 24 hours after its start. It has no inactivity timeout. ([OWASP ASVS 5.0, 7.1.1 and 7.1.2][asvs-v7]; [WCAG 2.2, SC 2.2.1][wcag-2-2-1]; [Judgement: a session of 24 hours with no inactivity timeout](#dr-0452))
- <a id="dr-0424"></a>**DR-0424** A member can have more than one session at the same time, with no limit. ([Judgement: a session of 24 hours with no inactivity timeout](#dr-0452))
- <a id="dr-0425"></a>**DR-0425** The data recipient keeps the sessions in the memory of the process, and at most 10,000 sessions that did not end. A restart ends each session. ([Judgement: sessions in the memory of the process](#dr-0453))
- <a id="dr-0426"></a>**DR-0426** When the data recipient has 10,000 sessions that did not end, it refuses a sign-in with 503. ([Judgement: sessions in the memory of the process](#dr-0453))
- <a id="dr-0428"></a>**DR-0428** When the data recipient has 10,000 sessions that did not end, the place of the valid session of a sign-in request is free for that request only. ([the free place of the data holder](../../dataholder/design/online-bank-site-sign-in.md#dh-2043); [Judgement: sessions in the memory of the process](#dr-0453))
- <a id="dr-0427"></a>**DR-0427** After a session ends, its session ID gives no access. ([OWASP ASVS 5.0, 7.4.1][asvs-v7])

## Session cookie

- <a id="dr-0430"></a>**DR-0430** The name of the session cookie is `__Host-member-session-<DATA_RECIPIENT_BRAND_ID>`, with the brand ID of the instance. ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]; [Judgement: a cookie name with the brand ID](#dr-0454))
- <a id="dr-0431"></a>**DR-0431** The session cookie has these attributes. ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]; [OWASP Session Management Cheat Sheet][owasp-session]; [Judgement: `SameSite=Lax` for the return from a data holder](#dr-0455))
  - `Secure`
  - `HttpOnly`
  - `SameSite=Lax`
  - `Path=/`
- <a id="dr-0432"></a>**DR-0432** The session cookie has no `Domain`, no `Max-Age` and no `Expires`. ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host])
- <a id="dr-0433"></a>**DR-0433** The response to a sign-out expires the session cookie with `Max-Age=0`. It has no `Clear-Site-Data` header. ([OWASP ASVS 5.0, 7.4.1][asvs-v7]; [no `Clear-Site-Data` at a sign-out of the data holder](../../dataholder/design/online-bank-site-sign-in.md#dh-2050))

## Password reset

- <a id="dr-0440"></a>**DR-0440** The password reset request page asks for the email address only. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0441"></a>**DR-0441** The response to a reset request is the same for each email address, in its status and its body. The page says that the member gets an email with a link when the address has a member. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password]; [Judgement: the time of a response that sends an email](#dr-0456))
- <a id="dr-0442"></a>**DR-0442** When a member has the email address, a reset request sends [the password reset email](email.md#dr-0512). When no member has the address, it sends no email. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0443"></a>**DR-0443** A reset request does not change the password and does not lock the membership. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0444"></a>**DR-0444** The reset link has a token of 256 bits from a cryptographic random source, in the fragment of the URL. The data recipient keeps only the SHA-256 hash of the token. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password]; [Judgement: a token in the fragment and a form on the page of the link](sign-up.md#dr-0355))
- <a id="dr-0445"></a>**DR-0445** A reset link is valid for 30 minutes and for one use. A new reset request makes each earlier reset link of the member not valid. The data recipient deletes each reset token within 24 hours after it ends. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password]; [Judgement: a verification link of 24 hours and a reset link of 30 minutes](sign-up.md#dr-0356))
- <a id="dr-0446"></a>**DR-0446** The page of a reset link asks for a new password. The new password obeys the password rules of a sign-up. ([the password rules](sign-up.md#dr-0312); [the blocklist](sign-up.md#dr-0315))
- <a id="dr-0448"></a>**DR-0448** A successful reset changes the password, ends each session of the member, and sends [the email for a changed password](email.md#dr-0513). It starts no session. The web app then shows the sign-in page. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password]; [OWASP ASVS 5.0, 6.3.7 and 7.4.3][asvs-v7])
- <a id="dr-0449"></a>**DR-0449** When a reset link is not valid, has expired or was used, the page says that the link does not work, and that the member can ask for a new link. ([OWASP ASVS 5.0, 6.3.8][asvs-v6])

## Judgements

### <a id="dr-0450"></a>DR-0450 One refusal that hides the members

- Kind: choice
- Basis: Step 3.2. OWASP requires one generic refusal for an unknown user and an incorrect password ([OWASP Authentication Cheat Sheet][owasp-authentication]). The data holder shows which identifiers exist, because its sign-in checks no password ([the refusal of the data holder](../../dataholder/design/online-bank-site-sign-in.md#dh-2025)).
- Reason: The sign-up and the reset also hide the members, so a refusal that shows them gives the information back. The data recipient has a real password check, so the reason of the data holder does not apply.

### <a id="dr-0451"></a>DR-0451 No limit of failed sign-ins in the lab

- Kind: gap
- Basis: Step 3.2. NIST requires a limit of 100 consecutive failed attempts or fewer for each account ([NIST SP 800-63B-4, section 3.2.2][nist-63b-rate-limit]). The lab services have no rate limit ([production gaps](../../../docs/production-gaps.md)).
- Reason: The lab has invented members only. A limit is part of the rate limits that a production service needs.

### <a id="dr-0452"></a>DR-0452 A session of 24 hours with no inactivity timeout

- Kind: gap
- Basis: Step 3.2. ASVS lets the risk set the inactivity timeout ([OWASP ASVS 5.0, 7.3.1][asvs-v7]). WCAG requires a warning and an extension before a time limit of 20 hours or less ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]). The data holder has the same rule ([no inactivity timeout of the data holder](../../dataholder/design/online-bank-site-sign-in.md#dh-2107)).
- Reason: ASVS requires an inactivity timeout, and the data recipient does not meet it. A time limit of more than 20 hours meets WCAG with no warning, and the lab has invented members only.

### <a id="dr-0453"></a>DR-0453 Sessions in the memory of the process

- Kind: choice
- Basis: Step 3.3. The data holder keeps the sessions of its online bank site in the memory of the process ([sessions in memory of the data holder](../../dataholder/design/online-bank-site-sign-in.md#dh-2106)).
- Reason: The lab runs one instance of each brand, and a member who signs in again after a restart is acceptable. A store in the memory needs no table.

### <a id="dr-0454"></a>DR-0454 A cookie name with the brand ID

- Kind: choice
- Basis: Step 3.2. Cookies are not isolated by port ([RFC 6265, section 8.5][rfc6265-ports]), and the lab UIs share the host `localhost`.
- Reason: A different name for each brand stops one brand from replacing the session of a different brand in one browser. The name does not stop a different lab app from reading the cookie, so the lab accepts the shared host as a production gap.

### <a id="dr-0455"></a>DR-0455 `SameSite=Lax` for the return from a data holder

- Kind: choice
- Basis: Step 3.2. A browser does not send a `Strict` cookie on a navigation from a different site ([RFC 6265bis, section 5.6.7.1][rfc6265bis-samesite]). The authorisation flow returns the consumer from the data holder to the data recipient with a navigation.
- Reason: With `Strict`, a deployment with a different host for each participant loses the session at the return from the data holder. `Lax` sends the cookie only on a navigation with a safe method, and the changes need a method that is not safe ([the requests that are not `GET`](#dr-0411)).

### <a id="dr-0456"></a>DR-0456 The time of a response that sends an email

- Kind: gap
- Basis: Step 3.2. ASVS Level 3 requires that the time of a response does not show which accounts exist ([OWASP ASVS 5.0, 6.3.8][asvs-v6]). The data recipient sends an email in the request.
- Reason: A reset request for an address with a member sends an email, and a request for a different address sends none, so the times are different. A sign-up sends one of two emails, so its times are almost the same. The lab accepts the difference.

[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[asvs-v6]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x15-V6-Authentication.md
[asvs-v7]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x16-V7-Session-Management.md
[asvs-v8]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x17-V8-Authorization.md
[nist-63b-passwords]: https://pages.nist.gov/800-63-4/sp800-63b/authenticators/#passwordver
[nist-63b-rate-limit]: https://pages.nist.gov/800-63-4/sp800-63b/authenticators/#throttle
[owasp-authentication]: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
[owasp-csrf]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
[owasp-forgot-password]: https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
[owasp-session]: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
[rfc10017-cookie-session]: https://www.rfc-editor.org/rfc/rfc10017#section-7.1
[rfc6265-ports]: https://www.rfc-editor.org/rfc/rfc6265#section-8.5
[rfc6265bis-host]: https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-rfc6265bis#section-4.1.3.2
[rfc6265bis-samesite]: https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-rfc6265bis#section-5.6.7.1
[wcag-1-3-5]: https://www.w3.org/TR/WCAG22/#identify-input-purpose
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[wcag-3-3-8]: https://www.w3.org/TR/WCAG22/#accessible-authentication-minimum
