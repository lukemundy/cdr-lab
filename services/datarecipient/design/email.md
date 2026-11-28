# Mock data recipient: Email

This area gives the emails that the data recipient sends to a consumer, their content and their presentation.

## Sending

- <a id="dr-0500"></a>**DR-0500** The data recipient sends each email through the SMTP server of the setting `SMTP_ADDRESS`, from the address of the setting `MAIL_FROM`. Both settings are required, and they have no default. ([RFC 5321][rfc5321]; [Judgement: the mail settings of the data holder](#dr-0550))
- <a id="dr-0501"></a>**DR-0501** The `From` header has the brand name as its display name and the address of `MAIL_FROM` as its address. The header is valid for each brand name that the brand directory permits. ([RFC 5322, section 3.4][rfc5322-3-4]; [the `From` header of the data holder](../../dataholder/design/consumer-authentication.md#dh-1245))
- <a id="dr-0502"></a>**DR-0502** When the data recipient cannot send the email of a sign-up, the sign-up gives 503 and keeps no pending sign-up. Each branch of a sign-up sends an email, so the 503 does not show if a member has the address. ([Judgement: no change without its email](#dr-0551))
- <a id="dr-0514"></a>**DR-0514** When the data recipient cannot send a different email, the response of the request does not change, and the span of the request records the failure. For example, a reset request for an address with a member gives the same response when the mail server is down. ([OWASP ASVS 5.0, 6.3.8][asvs-v6])
- <a id="dr-0503"></a>**DR-0503** The response to the request that sends an email does not say if the data recipient sent an email. ([OWASP ASVS 5.0, 6.3.8][asvs-v6]; [the same response for each address](sign-up.md#dr-0307))

## Presentation

- <a id="dr-0504"></a>**DR-0504** Each email has a header with `logo.png` of the brand, on the header colour of the brand. ([the logo of a brand directory](brand.md#dr-0113))
- <a id="dr-0505"></a>**DR-0505** Each email has a plain text part and an HTML part with the same information, as the OTP email of the data holder has. ([the parts of the OTP email](../../dataholder/design/consumer-authentication.md#dh-1247); [the logo in a `multipart/related` body](../../dataholder/design/consumer-authentication.md#dh-1248); [the transfer encoding](../../dataholder/design/consumer-authentication.md#dh-1250))
- <a id="dr-0506"></a>**DR-0506** Each email has the password statement and the demonstration statement at its end. ([the password statement](web-app.md#dr-0224); [the demonstration statement](web-app.md#dr-0225))
- <a id="dr-0507"></a>**DR-0507** The text of each email is in plain language. ([plain language in a consumer UI](../../../web/packages/ui/design/content.md#ui-0142))
- <a id="dr-0508"></a>**DR-0508** Each link of an email starts with the base URI of the web app. ([the base URI of the web app](web-app.md#dr-0207))
- <a id="dr-0509"></a>**DR-0509** The subject of an email does not contain a token or a link. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])

## Emails

- <a id="dr-0510"></a>**DR-0510** The verification email has the verification link and the time for which it is valid. It says that the consumer can ignore the email when they did not sign up. ([the verification link](sign-up.md#dr-0330))
- <a id="dr-0511"></a>**DR-0511** The email for an address that has a member says that someone tried to sign up with the address. It has a link to the sign-in page and a link to the password reset request page. ([OWASP Authentication Cheat Sheet][owasp-authentication])
- <a id="dr-0512"></a>**DR-0512** The password reset email has the reset link and the time for which it is valid. It says that the password does not change when the member ignores the email. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0513"></a>**DR-0513** The email for a changed password says that the password of the membership changed, and that the member must reset the password again when they did not change it. ([OWASP ASVS 5.0, 6.3.7][asvs-v6])

## Judgements

### <a id="dr-0550"></a>DR-0550 The mail settings of the data holder

- Kind: choice
- Basis: Step 3.3. The data holder sends its OTP email with the settings `SMTP_ADDRESS` and `MAIL_FROM` ([the email of the data holder](../../dataholder/design/consumer-authentication.md#dh-1244)).
- Reason: The same settings let a lab user configure the mail of each kind of instance in the same way.

### <a id="dr-0551"></a>DR-0551 No change without its email

- Kind: choice
- Basis: Step 3.3. A pending sign-up that the consumer does not get has no use.
- Reason: A refusal with 503 tells the consumer to try again. A change with no email has no use and is not visible to the consumer.

[asvs-v6]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x15-V6-Authentication.md
[owasp-authentication]: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
[owasp-forgot-password]: https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
[rfc5321]: https://www.rfc-editor.org/rfc/rfc5321
[rfc5322-3-4]: https://www.rfc-editor.org/rfc/rfc5322#section-3.4
