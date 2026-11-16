# Mock data holder: Consumer authentication

This area gives the authentication of a consumer with a one-time password, the OTP email and the user identifiers of a customer. The [authorisation](authorisation.md) gives the page that follows the authentication. The [authorisation server library](../../../pkg/authserver/design/authorisation-request.md) gives the authorisation request that starts the flow.

## User identifiers

- <a id="dh-1200"></a>**DH-1200** The customer number and the email address of a customer are each a user identifier of the consumer. The consumer can give either of them. ([CDS: User Identifiers][cds-user-identifiers]; [Judgement: two user identifiers in one field](#dh-1270))
- <a id="dh-1201"></a>**DH-1201** The data holder creates the customer number of each customer. The customer number has 8 decimal digits, and it is unique in the brand. For example, `40213307`. ([CDS: User Identifiers][cds-user-identifiers])
- <a id="dh-1202"></a>**DH-1202** The email address of a customer is unique in the brand, with no difference for upper and lower case. For example, when a customer has `Jo@example.test`, no other customer of the brand can have `jo@example.test`. ([CDS: User Identifiers][cds-user-identifiers]; [Judgement: no case difference in an email address](#dh-1271))
- <a id="dh-1203"></a>**DH-1203** The data holder keeps the email address as the consumer or the lab operator gives it. It finds a customer by the email address with no difference for upper and lower case. ([Judgement: no case difference in an email address](#dh-1271))
- <a id="dh-1204"></a>**DH-1204** A customer of a different brand can have the same email address. ([Judgement: no case difference in an email address](#dh-1271))
- <a id="dh-1205"></a>**DH-1205** The email address of the customer is the channel for the one-time password (OTP). ([CDS: One-Time Password Credential Requirements][cds-otp]; [CDR Support: One Time Password (OTP)][support-otp]; [Judgement: the OTP by email](#dh-1276))
- <a id="dh-1206"></a>**DH-1206** When a user identifier contains `@`, the data holder takes it as an email address. When it does not, the data holder takes it as a customer number. ([Judgement: two user identifiers in one field](#dh-1270))
- <a id="dh-1207"></a>**DH-1207** The data holder removes the spaces at the start and at the end of a user identifier before it finds the customer. For example, ` 40213307 ` finds the customer `40213307`. ([Judgement: two user identifiers in one field](#dh-1270))

## Authentication pages

- <a id="dh-1210"></a>**DH-1210** The only authentication flow is Redirect to Web with an OTP. ([CDS: Fallback Authentication Flows][cds-fallback-flows])
- <a id="dh-1211"></a>**DH-1211** The first page of the authentication asks for the email address or the customer number in one field. It asks for nothing else. ([CDS: User Identifiers][cds-user-identifiers]; [Judgement: two user identifiers in one field](#dh-1270))
- <a id="dh-1212"></a>**DH-1212** The first page says that the data recipient does not need the password of the consumer to get CDR data. It also says that the consumer must not give a password. ([CDS: Common Authentication Standards][cds-cx-authentication])
- <a id="dh-1213"></a>**DH-1213** The authentication pages never ask for a password. ([Judgement: no password on the authentication pages](#dh-1273))
- <a id="dh-1214"></a>**DH-1214** The authentication pages have no link for forgotten details. They offer no other products or services, and they have no reference to other documents. ([CDS: Redirect to Web with One Time Password][cds-cx-otp]; [CDS: Common Authentication Standards][cds-cx-authentication])
- <a id="dh-1215"></a>**DH-1215** The authentication pages name the data recipient with its brand name. The brand name is the `org_name` of the registration. ([CX Guidelines: Redirect to Web with One Time Password][cx-otp])
- <a id="dh-1216"></a>**DH-1216** Each authentication page has a control to cancel. When the consumer cancels, the authorisation flow ends with the error `access_denied`. ([CDR Support: Error scenarios and responses][support-error-scenarios]; [Judgement: `access_denied` at the end of an authentication](#dh-1275))

## One-time password

- <a id="dh-1220"></a>**DH-1220** When the consumer gives the user identifier of a customer, the data holder makes an OTP. It sends the OTP by email to the email address of that customer. ([CDS: One-Time Password Credential Requirements][cds-otp]; [Judgement: the OTP by email](#dh-1276))
- <a id="dh-1221"></a>**DH-1221** After the consumer gives a user identifier, the next page asks for the OTP. ([CDS: Redirect to Web with One Time Password][cds-cx-otp])
- <a id="dh-1222"></a>**DH-1222** When the consumer enters the correct OTP, the data holder authenticates the customer of that OTP. The authorisation is then for that customer. ([CDS: One-Time Password Credential Requirements][cds-otp])
- <a id="dh-1223"></a>**DH-1223** When no customer has the user identifier, the data holder sends no email and shows the same OTP page. No OTP is then correct. ([CDS: One-Time Password Credential Requirements][cds-otp]; [Judgement: no enumeration of user identifiers](#dh-1274))
- <a id="dh-1224"></a>**DH-1224** An OTP has 6 decimal digits from a cryptographic random source. When a [fixed OTP](#dh-1264) applies, the OTP has the fixed value. ([CDS: One-Time Password Credential Requirements][cds-otp]; [Judgement: the limits of an OTP](#dh-1277))
- <a id="dh-1225"></a>**DH-1225** An OTP is valid for 5 minutes. ([CDS: One-Time Password Credential Requirements][cds-otp]; [Judgement: the limits of an OTP](#dh-1277))
- <a id="dh-1226"></a>**DH-1226** An OTP is valid for one use, and only in the authorisation flow that made it. ([CDS: One-Time Password Credential Requirements][cds-otp])
- <a id="dh-1227"></a>**DH-1227** The OTP page shows the time for which the OTP is valid. For example, the page says that the OTP is valid for 5 minutes. ([CDS: Redirect to Web with One Time Password][cds-cx-otp])
- <a id="dh-1228"></a>**DH-1228** The consumer can ask for a new OTP on the OTP page. A new OTP makes the previous OTP not valid. ([CX Guidelines: Redirect to Web with One Time Password][cx-otp])
- <a id="dh-1229"></a>**DH-1229** The data holder sends a maximum of 3 OTPs in one authorisation flow. ([CX Guidelines: Redirect to Web with One Time Password][cx-otp]; [Judgement: the limits of an OTP](#dh-1277))
- <a id="dh-1230"></a>**DH-1230** After 3 incorrect OTP entries in one authorisation flow, the flow ends with the error `access_denied`. ([CDR Support: One Time Password (OTP)][support-otp]; [CDR Support: Error scenarios and responses][support-error-scenarios]; [Judgement: the limits of an OTP](#dh-1277); [Judgement: `access_denied` at the end of an authentication](#dh-1275))
- <a id="dh-1231"></a>**DH-1231** When the data holder cannot send the OTP email, the authorisation flow ends with the error `temporarily_unavailable`. ([RFC 6749, section 4.1.2.1][rfc6749-authorisation-errors])
- <a id="dh-1232"></a>**DH-1232** When the consumer does not complete the authorisation flow, the flow [ends 10 minutes after the authorisation request](../../../pkg/authserver/design/authorisation-request.md#auth-0103). The data holder then shows an error page. ([the end of an authorisation flow](../../../pkg/authserver/design/authorisation-request.md#auth-0103))
- <a id="dh-1233"></a>**DH-1233** A correct OTP entry gives the level of assurance `urn:cds.au:cdr:2`. The time of that entry is the authentication time. ([CDS: Levels of Assurance][cds-loa]; [Judgement: one level of assurance](../../../pkg/authserver/design/authorisation-request.md#auth-0121))

## OTP email

- <a id="dh-1240"></a>**DH-1240** The OTP email has the OTP, the brand name of the data recipient and the time for which the OTP is valid. ([CDS: One-Time Password Credential Requirements][cds-otp]; [CX Guidelines: Redirect to Web with One Time Password][cx-otp])
- <a id="dh-1241"></a>**DH-1241** The OTP email says that the OTP is only for the sharing of CDR data. ([CDS: One-Time Password Credential Requirements][cds-otp]; [CX Guidelines: Redirect to Web with One Time Password][cx-otp])
- <a id="dh-1242"></a>**DH-1242** The OTP email says that the data recipient does not need the password of the consumer to get CDR data. ([CDS: Common Authentication Standards][cds-cx-authentication])
- <a id="dh-1243"></a>**DH-1243** The subject of the OTP email does not contain the OTP. For example, "Your one-time password to share data with Acme Budget". ([Judgement: no OTP in the subject](#dh-1280))
- <a id="dh-1244"></a>**DH-1244** The `From` header has the brand name as its display name and the sender address of the configuration as its address. ([RFC 5322, section 3.4][rfc5322-3-4]; [one source of the brand name](brand.md#dh-0305); [Judgement: the brand name as the sender name](#dh-1281))
- <a id="dh-1245"></a>**DH-1245** The `From` header is one valid header line for each brand name that the brand directory permits. This includes a name with special characters or with characters that are not ASCII. ([RFC 5322, section 3.4][rfc5322-3-4]; [RFC 2047, section 5][rfc2047-5]; [no control character in the brand name](brand.md#dh-0333))
- <a id="dh-1246"></a>**DH-1246** When the sender address setting has a display name, the data holder does not start. The error names the setting. For example, the value `Wattle Bank <no-reply@wattlebank.test>` stops the start. ([Judgement: the brand name as the sender name](#dh-1281))
- <a id="dh-1247"></a>**DH-1247** The OTP email has a plain text part and an HTML part, which give the same information. They are in a `multipart/alternative` body, with the plain text part first. ([RFC 2046, section 5.1.4][rfc2046-5-1-4]; [Judgement: an email format for email clients](#dh-1282))
- <a id="dh-1248"></a>**DH-1248** The HTML part and the logo are in a `multipart/related` body. The HTML part refers to the logo with a `cid:` URL. ([RFC 2387][rfc2387]; [RFC 2392][rfc2392]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-1249"></a>**DH-1249** No line of the message is longer than 998 characters. ([RFC 5322, section 2.1.1][rfc5322-2-1-1])
- <a id="dh-1250"></a>**DH-1250** The plain text part and the HTML part have the transfer encoding `quoted-printable`. ([RFC 5322, section 2.1.1][rfc5322-2-1-1]; [Judgement: an email format for email clients](#dh-1282))
- <a id="dh-1251"></a>**DH-1251** The only image of the OTP email is `logo.png` of the brand directory, in the email. The email loads no content from a different location. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-1252"></a>**DH-1252** The header of the HTML part shows the logo on the brand colour. The text alternative of the logo is the brand name. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="dh-1253"></a>**DH-1253** The HTML part uses only inline styles, with literal sRGB colours in hex notation from the tokens of the brand, and a system font stack. For example, `color: #1a3d2b`. ([caniemail: style element][caniemail-style]; [caniemail: CSS variables][caniemail-css-variables]; [caniemail: modern colour functions][caniemail-modern-colour]; [Judgement: an email format for email clients](#dh-1282))
- <a id="dh-1254"></a>**DH-1254** Each pair of text and background colours in the HTML part is a [pair of text and background roles of the brand](brand.md#dh-0343). For example, the text colour on the brand colour. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3])
- <a id="dh-1255"></a>**DH-1255** The OTP is text in the HTML part, not an image, so that the consumer can copy it. ([WCAG 2.2, SC 3.3.8][wcag-3-3-8]; [Judgement: an OTP that the consumer can copy](#dh-1284))
- <a id="dh-1256"></a>**DH-1256** The OTP email has no link and no button. It offers no other products or services, and it has no reference to other documents. ([CDS: Common Authentication Standards][cds-cx-authentication]; [CX Guidelines: Redirect to Web with One Time Password][cx-otp])
- <a id="dh-1257"></a>**DH-1257** Both parts of the OTP email have the [footer text of the consumer pages](authorisation.md#dh-0837), with no link. ([Judgement: the footer text in the email](#dh-1285))
- <a id="dh-1258"></a>**DH-1258** The HTML part has the language `en-AU` on its `html` element, and it has a title. ([WCAG 2.2, SC 3.1.1][wcag-3-1-1]; [Judgement: the HTML part as a document for a screen reader](#dh-1286))
- <a id="dh-1259"></a>**DH-1259** Each table for layout in the HTML part has the role `presentation`. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; [Judgement: the HTML part as a document for a screen reader](#dh-1286))
- <a id="dh-1260"></a>**DH-1260** The data holder sends no email other than the OTP email. For example, it sends no email to the consumer at the end of an arrangement. ([CDR Rules][cdr-rules], rules 4.26A to 4.28; [Judgement: no email other than the OTP email](#dh-1287))

## Mail delivery and fixed OTPs

These rules are lab-only. A production data holder has its own channel for the OTP.

- <a id="dh-1262"></a>**DH-1262** The data holder sends each OTP email through the SMTP server that its configuration gives. ([CDS: One-Time Password Credential Requirements][cds-otp]; [Judgement: the OTP by email](#dh-1276))
- <a id="dh-1263"></a>**DH-1263** In the lab, the SMTP server is a mail sink. It keeps each email and delivers none. A lab user and the conformance tests read the OTP from the mail sink. ([Judgement: the OTP by email](#dh-1276))
- <a id="dh-1264"></a>**DH-1264** The configuration can give fixed OTPs: a list of email addresses, each with an OTP of 6 digits. When a customer has one of these addresses, each OTP of that customer has the fixed value. ([Judgement: fixed OTPs for the OpenID conformance suite](#dh-1288))
- <a id="dh-1265"></a>**DH-1265** The case of the address has no effect on a fixed OTP. For example, the entry for `openid@mail.test` also applies to `OpenID@mail.test`. ([Judgement: fixed OTPs for the OpenID conformance suite](#dh-1288))
- <a id="dh-1266"></a>**DH-1266** A fixed OTP changes only the value of the OTP. All the other OTP rules stay the same, and the email still goes to the customer. ([Judgement: fixed OTPs for the OpenID conformance suite](#dh-1288))
- <a id="dh-1267"></a>**DH-1267** When the configuration gives no list of fixed OTPs, each OTP is random. The default is no list. ([Judgement: fixed OTPs for the OpenID conformance suite](#dh-1288))
- <a id="dh-1268"></a>**DH-1268** When an entry of the list of fixed OTPs is not valid, the data holder does not start. An entry is not valid when it has no valid email address, when its OTP does not have 6 digits, or when it has the address of an earlier entry. ([Judgement: fixed OTPs for the OpenID conformance suite](#dh-1288))

## Judgements

### <a id="dh-1270"></a>DH-1270 Two user identifiers in one field

- Kind: choice
- Basis: Step 3.3. The CDS requires an identifier "that can uniquely identify the customer and that is already known by the customer" ([CDS: User Identifiers][cds-user-identifiers]). It does not say how many identifiers a data holder accepts.
- Reason: A consumer knows the email address better than a number of 8 digits that the data holder makes. A customer number has only digits, and an email address always has `@`, so one field needs no selection of the type. A value that a consumer pastes can have spaces.

### <a id="dh-1271"></a>DH-1271 No case difference in an email address

- Kind: choice
- Basis: Step 3.3. An identifier must identify one customer ([CDS: User Identifiers][cds-user-identifiers]). The CDS gives no rule for the case of an email address.
- Reason: Most mail systems make no difference for the case of an address, and a consumer can type it in a different case. The data holder of each brand has its own database, so a different brand can have the same address.

### <a id="dh-1273"></a>DH-1273 No password on the authentication pages

- Kind: choice
- Basis: Step 3.3. The CDS does not permit a memorised secret as the single factor to achieve "LoA 2" ([CDS: Restricted Credentials][cds-restricted-credentials]).
- Reason: The OTP is the only factor, so the pages need no password. The sign-in of the online bank site asks for a password, so a page that never asks for one is easy to tell apart.

### <a id="dh-1274"></a>DH-1274 No enumeration of user identifiers

- Kind: choice
- Basis: Step 3.3. The data holder "SHOULD implement additional controls to minimise the risk of enumeration attacks" ([CDS: One-Time Password Credential Requirements][cds-otp]). The CDS does not say which controls.
- Reason: The same OTP page for each user identifier does not show which email addresses and customer numbers exist.

### <a id="dh-1275"></a>DH-1275 `access_denied` at the end of an authentication

- Kind: choice
- Basis: Step 3.2. The ACCC says that the data holder "should respond to the ADR (client) with an error outcome" for a cancel ([CDR Support: Error scenarios and responses][support-error-scenarios]). The CDS gives no error value.
- Reason: `access_denied` is the value of RFC 6749 for a resource owner that denies the request ([RFC 6749, section 4.1.2.1][rfc6749-authorisation-errors]). A flow that ends after incorrect OTPs gets the same value, because the consumer did not authenticate.

### <a id="dh-1276"></a>DH-1276 The OTP by email

- Kind: choice
- Basis: Step 3.3. The CDS requires "an existing channel or mechanism", at the discretion of the data holder ([CDS: One-Time Password Credential Requirements][cds-otp]). An OTP does not have to go by SMS ([CDR Support: One Time Password (OTP)][support-otp]).
- Reason: Email is a channel that a mock can have with the behaviour of a production data holder. A mail sink makes sure that the lab sends no email to a real address.

### <a id="dh-1277"></a>DH-1277 The limits of an OTP

- Kind: choice
- Basis: Step 3.3. The CDS requires 4 to 6 digits and an expiry, and leaves the period to the data holder ([CDS: One-Time Password Credential Requirements][cds-otp]). The ACCC says that a limit on failed entries is good practice ([CDR Support: One Time Password (OTP)][support-otp]).
- Reason: 5 minutes is sufficient to read an email. 3 entries of a 6-digit OTP give a guess a chance of 3 in 1,000,000.

### <a id="dh-1280"></a>DH-1280 No OTP in the subject

- Kind: choice
- Basis: Step 3.3. No source gives a rule for the subject of an OTP email.
- Reason: A device can show the subject of an email on a locked screen. Then a different person can read the OTP.

### <a id="dh-1281"></a>DH-1281 The brand name as the sender name

- Kind: choice
- Basis: Step 3.3. No source gives the sender of an OTP email.
- Reason: The consumer sees which brand sent the OTP. The brand directory stays the one source of the name, and a display name in the setting can disagree with it. An error at start-up is better than an error at the first OTP.

### <a id="dh-1282"></a>DH-1282 An email format for email clients

- Kind: choice
- Basis: Step 3.2. The best part of a `multipart/alternative` body is last ([RFC 2046, section 5.1.4][rfc2046-5-1-4]). Many clients do not support a `<style>` element, CSS variables or modern colour functions ([caniemail: style element][caniemail-style], [caniemail: CSS variables][caniemail-css-variables], [caniemail: modern colour functions][caniemail-modern-colour]).
- Reason: Quoted-printable keeps each line of a generated part in the limit of 998 characters. Most clients do not load a web font, so the logo gives the look of the brand.

### <a id="dh-1284"></a>DH-1284 An OTP that the consumer can copy

- Kind: choice
- Basis: Step 3.2. An authentication must not require a cognitive function test, for example a transcription, unless a mechanism helps ([WCAG 2.2, SC 3.3.8][wcag-3-3-8]).
- Reason: A consumer who can copy the OTP does not have to transcribe it.

### <a id="dh-1285"></a>DH-1285 The footer text in the email

- Kind: choice
- Basis: Step 3.3. No source requires a footer in an OTP email. The [footer of the consumer pages](authorisation.md#dh-0860) has the same reason.
- Reason: The email copies the brand of a bank, and it must not deceive a person, as for the pages.

### <a id="dh-1286"></a>DH-1286 The HTML part as a document for a screen reader

- Kind: choice
- Basis: Step 3.2. WCAG requires the language of a page and the structure of its content ([WCAG 2.2, SC 3.1.1][wcag-3-1-1], [SC 1.3.1][wcag-1-3-1]). WCAG is for web content, and an email is not a web page.
- Reason: A screen reader reads the HTML part as a document, so the rules of a web page apply.

### <a id="dh-1287"></a>DH-1287 No email other than the OTP email

- Kind: choice
- Basis: Step 3.3. The CDR Rules require a notice to the accredited person when an authorisation ends (rule 4.26A), and an update of the consumer dashboard (rule 4.27). The only notice to a consumer in rules 4.25 to 4.28 is for a secondary user (rule 4.28), who is out of scope ([CDR Rules][cdr-rules]).
- Reason: It is the least complex solution. An email to the consumer can be a later change.

### <a id="dh-1288"></a>DH-1288 Fixed OTPs for the OpenID conformance suite

- Kind: choice
- Basis: Step 3.3. No source gives a test feature for the authentication of a mock data holder.
- Reason: The browser of the OpenID conformance suite cannot read the mail sink, and it must authenticate with no manual step. The other customers keep a random OTP.

[adr-0023]: ../../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md
[caniemail-css-variables]: https://www.caniemail.com/features/css-variables/
[caniemail-modern-colour]: https://www.caniemail.com/features/css-modern-color/
[caniemail-style]: https://www.caniemail.com/features/html-style/
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[cds-cx-otp]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_redirect-to-web-with-one-time-password
[cds-fallback-flows]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#fallback-authentication-flows
[cds-loa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#levels-of-assurance-loas
[cds-otp]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#one-time-password-credential-requirements
[cds-restricted-credentials]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#restricted-credentials
[cds-user-identifiers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#user-identifiers
[cx-otp]: https://cx.dsb.gov.au/authenticate/redirect-with-one-time-password
[rfc2046-5-1-4]: https://www.rfc-editor.org/rfc/rfc2046#section-5.1.4
[rfc2047-5]: https://www.rfc-editor.org/rfc/rfc2047#section-5
[rfc2387]: https://www.rfc-editor.org/rfc/rfc2387
[rfc2392]: https://www.rfc-editor.org/rfc/rfc2392
[rfc5322-2-1-1]: https://www.rfc-editor.org/rfc/rfc5322#section-2.1.1
[rfc5322-3-4]: https://www.rfc-editor.org/rfc/rfc5322#section-3.4
[rfc6749-authorisation-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2.1
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[support-otp]: https://cdr-support.zendesk.com/hc/en-us/articles/6140553082255
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-1-4-3]: https://www.w3.org/TR/WCAG22/#contrast-minimum
[wcag-3-1-1]: https://www.w3.org/TR/WCAG22/#language-of-page
[wcag-3-3-8]: https://www.w3.org/TR/WCAG22/#accessible-authentication-minimum
