# Mock data recipient: Sign-up

This area gives the sign-up of a member, the verification of the email address, the password rules and the member details.

## Sign-up

- <a id="dr-0300"></a>**DR-0300** The sign-up page asks for the email address, and for nothing else. ([Judgement: the password at the verification](#dr-0351))
- <a id="dr-0301"></a>**DR-0301** The email address field has `type="email"` and `autocomplete="email"`. ([WCAG 2.2, SC 1.3.5][wcag-1-3-5])
- <a id="dr-0303"></a>**DR-0303** The consumer can paste into each field of the sign-up and of the verification, and a password manager can fill each field. ([WCAG 2.2, SC 3.3.8][wcag-3-3-8]; [OWASP ASVS 5.0, 6.2.7][asvs-v6])
- <a id="dr-0304"></a>**DR-0304** The sign-up page gives an example email address on the reserved domain `.test`. For example, `name@example.test`. The data recipient does not refuse an email address on a different domain. ([RFC 6761, section 6.2][rfc6761-test]; [Judgement: a demonstration with invented details only](web-app.md#dr-0255))
- <a id="dr-0305"></a>**DR-0305** The data recipient refuses an email address that is not an `addr-spec` of RFC 5322 with a domain, or that has more than 254 characters. ([RFC 5322, section 3.4.1][rfc5322-addr-spec]; [RFC 5321, section 4.5.3.1.3][rfc5321-path])
- <a id="dr-0306"></a>**DR-0306** Two email addresses that are different only in upper and lower case are the same address. For example, `Name@Example.test` and `name@example.test` are one address. ([Judgement: one member for each address, with no difference of case](#dr-0354))
- <a id="dr-0307"></a>**DR-0307** When the email address obeys the rules, the response of the sign-up is the same for each email address, in its status and its body. The web app then shows a page that tells the consumer to check their email. ([OWASP Authentication Cheat Sheet][owasp-authentication]; [Judgement: the time of a response that sends an email](sign-in.md#dr-0456))
- <a id="dr-0308"></a>**DR-0308** When no member has the email address, a sign-up keeps a pending sign-up with the email address. It sends [the verification email](email.md#dr-0510) to the address. ([Judgement: the password at the verification](#dr-0351))
- <a id="dr-0309"></a>**DR-0309** When a member has the email address, a sign-up keeps no pending sign-up. It sends [the email for an address that has a member](email.md#dr-0511) to the address. ([OWASP Authentication Cheat Sheet][owasp-authentication])
- <a id="dr-0310"></a>**DR-0310** A pending sign-up is not a member, and it has no password. The consumer cannot sign in with it. ([Judgement: the password at the verification](#dr-0351))
- <a id="dr-0311"></a>**DR-0311** An email address can have more than one pending sign-up. Each pending sign-up has its own verification link. ([Judgement: the password at the verification](#dr-0351))

## Passwords

- <a id="dr-0302"></a>**DR-0302** Each field for a new password has `type="password"`, `autocomplete="new-password"` and a control that shows the password as text. ([WCAG 2.2, SC 1.3.5][wcag-1-3-5]; [NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords])
- <a id="dr-0312"></a>**DR-0312** A password has 15 to 128 characters. The data recipient counts each Unicode code point as one character, after it normalises the password to NFC. ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords]; [Judgement: a password of 15 characters or more](#dr-0353))
- <a id="dr-0313"></a>**DR-0313** The data recipient accepts each Unicode character in a password, also a space. Other than the normalisation, it does not change, cut or trim a password. ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords]; [OWASP ASVS 5.0, 6.2.8][asvs-v6])
- <a id="dr-0314"></a>**DR-0314** A password has no composition rule. For example, the data recipient does not ask for a digit or a symbol. ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords]; [OWASP ASVS 5.0, 6.2.5][asvs-v6]; [Judgement: a password of 15 characters or more](#dr-0353))
- <a id="dr-0315"></a>**DR-0315** The data recipient refuses a password that, with no difference of case, is one of these values. ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords]; [OWASP ASVS 5.0, 6.2.4 and 6.2.11][asvs-v6]; [Judgement: a blocklist in the data recipient](#dr-0352))
  - a password of the blocklist
  - the brand name, the email address, or the part of the email address before the `@`
  - a password that repeats one character, for example `aaaaaaaaaaaaaaa`
- <a id="dr-0316"></a>**DR-0316** The blocklist has at least the 10,000 most common passwords of 15 characters or more, from a published list whose licence permits distribution. The data recipient contains the blocklist. ([OWASP ASVS 5.0, 6.2.4][asvs-v6]; [Judgement: a blocklist in the data recipient](#dr-0352))
- <a id="dr-0317"></a>**DR-0317** When the data recipient refuses a password, the field error says why. For example: "This password is too common. Choose a different password". ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords])
- <a id="dr-0318"></a>**DR-0318** The data recipient keeps only a hash of each password. The hash is Argon2id with a memory of 19 MiB, 2 iterations, 1 degree of parallelism and a random salt of 16 bytes. ([OWASP Password Storage Cheat Sheet][owasp-password-storage]; [OWASP ASVS 5.0, 11.4.2 and Appendix C][asvs-v11])
- <a id="dr-0319"></a>**DR-0319** The hash of a password includes its parameters, so that a change of the parameters does not make the earlier hashes unusable. ([OWASP Password Storage Cheat Sheet][owasp-password-storage])

## Verification of the email address

- <a id="dr-0330"></a>**DR-0330** The verification link has a token of 256 bits from a cryptographic random source, in the fragment of the URL. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password]; [Judgement: a token in the fragment and a form on the page of the link](#dr-0355))
- <a id="dr-0331"></a>**DR-0331** The data recipient keeps only the SHA-256 hash of the token. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0332"></a>**DR-0332** A verification link is valid for 24 hours after the sign-up. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password]; [Judgement: a verification link of 24 hours and a reset link of 30 minutes](#dr-0356))
- <a id="dr-0333"></a>**DR-0333** The page of a verification link asks for a new password, with the password rules. The page sends no request before the consumer submits the form. ([Judgement: the password at the verification](#dr-0351); [Judgement: a token in the fragment and a form on the page of the link](#dr-0355))
- <a id="dr-0334"></a>**DR-0334** When the consumer submits a valid link with a password that obeys the rules, and no member has the email address, the data recipient makes a member with the email address and the password. ([Judgement: the password at the verification](#dr-0351))
- <a id="dr-0335"></a>**DR-0335** When a link makes a member, each other pending sign-up of the email address ends, and its link is not valid. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0336"></a>**DR-0336** After the member is made, the web app shows the sign-in page with a message that the email address is verified. The verification starts no session. ([Judgement: a token in the fragment and a form on the page of the link](#dr-0355))
- <a id="dr-0337"></a>**DR-0337** When a link is not valid, has expired or was used, the page says that the link does not work, and that the consumer can sign up again or sign in. The data recipient gives the same response for each of these three cases. ([OWASP ASVS 5.0, 6.3.8][asvs-v6])
- <a id="dr-0338"></a>**DR-0338** A pending sign-up ends 24 hours after the sign-up. The data recipient deletes each pending sign-up within 24 hours after it ends. ([Privacy Act 1988][privacy-act], APP 11.2)

## Member details

- <a id="dr-0320"></a>**DR-0320** The member details page asks for these details, and for nothing else. ([Judgement: member details as a finance app asks for them](#dr-0350))
  - the given names, and the family name
  - the date of birth
  - the mobile number
  - the residential address: the unit (optional), the street number, the street name, the suburb, the state and the postcode
- <a id="dr-0321"></a>**DR-0321** Until the member gives the member details, each page that needs a session shows the member details page. ([Judgement: member details as a finance app asks for them](#dr-0350))
- <a id="dr-0322"></a>**DR-0322** The member can correct the member details on the member details page, with the same rules. ([Privacy Act 1988][privacy-act], APP 13)
- <a id="dr-0323"></a>**DR-0323** The member details page has a collection notice before the fields. It says who collects the details, why, what occurs when the member does not give them, and where the privacy policy is. ([Privacy Act 1988][privacy-act], APP 5.2)
- <a id="dr-0324"></a>**DR-0324** The given names and the family name each have 1 to 200 characters, are not only white space and have no control character. The data recipient accepts each other Unicode character. ([W3C: Personal names around the world][w3c-names]; [Judgement: lengths that hold each real value](#dr-0361))
- <a id="dr-0325"></a>**DR-0325** The page has an option "I have only one name". When the member selects it, the page hides the family name, and the family name is empty. ([W3C: Personal names around the world][w3c-names]; [Judgement: no forced family name](#dr-0358))
- <a id="dr-0326"></a>**DR-0326** The member is 18 years of age or older on the date of the request in the time zone `Australia/Sydney`. When the member is younger, the data recipient refuses the details with an error at the date of birth. ([Judgement: members of 18 or older](#dr-0362); [Judgement: the age on the date in Sydney](#dr-0357))
- <a id="dr-0327"></a>**DR-0327** A member who was born on 29 February has a birthday on 1 March in a year that is not a leap year. ([Judgement: the age on the date in Sydney](#dr-0357))
- <a id="dr-0328"></a>**DR-0328** The data recipient refuses a date of birth that is in the future or before 1 January 1900. ([Judgement: the age on the date in Sydney](#dr-0357))
- <a id="dr-0329"></a>**DR-0329** The mobile number is an Australian mobile number in one of these forms, with or without spaces. The data recipient keeps it in the E.164 form, for example `+61412345678`. ([Telecommunications Numbering Plan 2025][numbering-plan]; [ITU-T E.164][itu-e164])
  - 10 digits that start with `04`
  - `+61` and 9 digits that start with `4`
- <a id="dr-0339"></a>**DR-0339** The unit has 20 characters or less. The street number has 1 to 20 characters. The street name and the suburb each have 1 to 100 characters. ([Judgement: lengths that hold each real value](#dr-0361))
- <a id="dr-0340"></a>**DR-0340** The state is one of `ACT`, `NSW`, `NT`, `QLD`, `SA`, `TAS`, `VIC` and `WA`. The postcode has 4 digits. ([Australia Post: addressing guidelines][auspost-addressing])
- <a id="dr-0341"></a>**DR-0341** The data recipient refuses a residential address whose street number or street name starts with "PO Box", "GPO Box", "Locked Bag" or "Private Bag", with no difference of case. ([Judgement: no postal delivery address as the residential address](#dr-0363))
- <a id="dr-0342"></a>**DR-0342** When the first digit of the postcode is not a first digit of the state in the table that follows, the page shows a warning at the postcode. The member can save the details with the warning. ([Judgement: a warning, not a refusal, for the postcode](#dr-0359))

  | State | First digits |
  |---|---|
  | `ACT` | 0, 2 |
  | `NSW` | 1, 2, 3, 4 |
  | `NT` | 0 |
  | `QLD` | 4, 9 |
  | `SA` | 5 |
  | `TAS` | 7 |
  | `VIC` | 3, 8 |
  | `WA` | 6 |

- <a id="dr-0343"></a>**DR-0343** Each field of the member details has the `autocomplete` token of its purpose. ([WCAG 2.2, SC 1.3.5][wcag-1-3-5])

  | Field | Token |
  |---|---|
  | Given names | `given-name` |
  | Family name | `family-name` |
  | Date of birth | `bday` |
  | Mobile number | `tel` |
  | Unit | `address-line2` |
  | Street number and street name | `address-line1` |
  | Suburb | `address-level2` |
  | State | `address-level1` |
  | Postcode | `postal-code` |

- <a id="dr-0344"></a>**DR-0344** When the data recipient refuses a value of a form, the page shows the error at its field and in an error summary. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])

## Storage

- <a id="dr-0345"></a>**DR-0345** The data recipient keeps the members, the pending sign-ups and the member details as rows that it can change and delete. It records no event for them. ([Judgement: members in tables, not in events](#dr-0360))

## Judgements

### <a id="dr-0350"></a>DR-0350 Member details as a finance app asks for them

- Kind: choice
- Basis: Step 3.3. APP 3.2 permits a collection that is reasonably necessary for a function of the entity ([Privacy Act 1988][privacy-act]). A consumer finance app usually asks for these details after the sign-up.
- Reason: The data recipient is a realistic demonstration and a reference implementation, so it asks for the details that a real app asks for. The lab collects invented details only, and it collects them only from a person who controls the email address.

### <a id="dr-0351"></a>DR-0351 The password at the verification

- Kind: choice
- Basis: Step 3.2. A sign-up must not show which addresses have a member ([OWASP Authentication Cheat Sheet][owasp-authentication]). A membership that a different person makes for an address before its owner is a known attack.
- Reason: When the consumer sets the password on the page of the link, only the person who reads the email can make a member, and only with their own password. A pending sign-up with no password gives a different person nothing to use. The sign-up also calculates no hash, so its time does not change with the address.

### <a id="dr-0352"></a>DR-0352 A blocklist in the data recipient

- Kind: choice
- Basis: Step 3.2. NIST and ASVS require a blocklist of common passwords ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords]; [OWASP ASVS 5.0, 6.2.4][asvs-v6]). The lab has no access to an external service in each deployment.
- Reason: A list in the binary works in each deployment and in CI with no network call. The length rule already refuses a shorter common password, so the list has only passwords of 15 characters or more.

### <a id="dr-0353"></a>DR-0353 A password of 15 characters or more

- Kind: conflict
- Basis: Step 3.2. Schedule 2, clause 2.2, control (h) of the [CDR Rules][cdr-rules] names password complexity, account lockout, password history and password ageing. NIST requires 15 characters or more for a password that is the single factor, and forbids composition rules and periodic changes ([NIST SP 800-63B-4, section 3.1.1.2][nist-63b-passwords]).
- Reason: Control (h) is about the users of the CDR data environment, in our interpretation: the staff of the data recipient, not its consumers. For the consumers, we follow NIST and ASVS, also for lockout and history, because they are newer and more specific.

### <a id="dr-0354"></a>DR-0354 One member for each address, with no difference of case

- Kind: choice
- Basis: Step 3.2. RFC 5321 permits a local part that is case-sensitive, but says that a receiver must not depend on it ([RFC 5321, section 2.4][rfc5321-case]).
- Reason: Almost no mail system has two mailboxes that are different only in case. A consumer who types a capital letter must find their membership.

### <a id="dr-0355"></a>DR-0355 A token in the fragment and a form on the page of the link

- Kind: choice
- Basis: Step 3.2. A browser does not send the fragment of a URL to the server ([RFC 3986, section 3.5][rfc3986-fragment]). Mail systems can open each link of an email to check it.
- Reason: The token is then not in a log or a trace of the request for the page. A check of a link by a mail system loads the page, but it does not use the token.

### <a id="dr-0356"></a>DR-0356 A verification link of 24 hours and a reset link of 30 minutes

- Kind: choice
- Basis: Step 3.2. OWASP requires that a reset token expires ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password]). It gives no time.
- Reason: A new member can open the verification email later on the same day. A reset link gives access to the membership, so its time is short.

### <a id="dr-0357"></a>DR-0357 The age on the date in Sydney

- Kind: choice
- Basis: Step 3.3. No source gives a time zone for the age of a consumer.
- Reason: A date of birth is a date, not a time, so the data recipient compares two dates. One fixed Australian time zone gives the same result on each server.

### <a id="dr-0358"></a>DR-0358 No forced family name

- Kind: choice
- Basis: Step 3.2. The W3C says that a form must not require a family name ([W3C: Personal names around the world][w3c-names]).
- Reason: Some people have only one name. A form that forces a family name makes them type an invented one.

### <a id="dr-0359"></a>DR-0359 A warning, not a refusal, for the postcode

- Kind: choice
- Basis: Step 3.3. Some postcodes are for places on both sides of a state border, for example a postcode of a town on the border of New South Wales and Victoria.
- Reason: A strict check refuses some real addresses. A warning helps a member who made a mistake.

### <a id="dr-0360"></a>DR-0360 Members in tables, not in events

- Kind: choice
- Basis: Step 3.3. The services use event sourcing for participant state ([ADR 0003][adr-0003]). A member must be able to correct personal information (APP 13 of the [Privacy Act 1988][privacy-act]), and the data recipient must destroy personal information that it no longer needs (APP 11.2). An event does not change.
- Reason: A member is consumer data, not participant state. Rows can change and be deleted, and no rule needs a history of the member details.

### <a id="dr-0361"></a>DR-0361 Lengths that hold each real value

- Kind: choice
- Basis: Step 3.3. No source gives a maximum length for a name or an address field. The W3C says that a form must not have short length limits for names ([W3C: Personal names around the world][w3c-names]).
- Reason: The limits are much larger than each real value, and they limit the size of a request.

### <a id="dr-0362"></a>DR-0362 Members of 18 or older

- Kind: choice
- Basis: Step 3.3. Only an individual of 18 years of age or older is an eligible CDR consumer of a data holder ([CDR Rules][cdr-rules], rule 1.10B(1)(a)(i)). No rule requires the data recipient to check the age.
- Reason: A younger member can connect no bank, so a membership has no use for them. A check at the member details tells them before they continue.

### <a id="dr-0363"></a>DR-0363 No postal delivery address as the residential address

- Kind: choice
- Basis: Step 3.3. A PO Box, a GPO Box, a Locked Bag and a Private Bag are postal delivery addresses ([Australia Post: addressing guidelines][auspost-addressing]). They are not a place where a person lives.
- Reason: The field asks for the residential address, so a postal delivery address is a mistake. A refusal with a plain-language error tells the member at once.

[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[asvs-v6]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x15-V6-Authentication.md
[asvs-v11]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x20-V11-Cryptography.md
[auspost-addressing]: https://auspost.com.au/business/marketing-and-communications/access-data-and-insights/address-data/addressing-guidelines
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[itu-e164]: https://www.itu.int/rec/T-REC-E.164
[nist-63b-passwords]: https://pages.nist.gov/800-63-4/sp800-63b/authenticators/#passwordver
[numbering-plan]: https://www.legislation.gov.au/F2025L00409/latest/text
[owasp-authentication]: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
[owasp-forgot-password]: https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
[owasp-password-storage]: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
[privacy-act]: https://www.legislation.gov.au/C2004A03712/latest/text
[rfc3986-fragment]: https://www.rfc-editor.org/rfc/rfc3986#section-3.5
[rfc5321-case]: https://www.rfc-editor.org/rfc/rfc5321#section-2.4
[rfc5321-path]: https://www.rfc-editor.org/rfc/rfc5321#section-4.5.3.1.3
[rfc5322-addr-spec]: https://www.rfc-editor.org/rfc/rfc5322#section-3.4.1
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[w3c-names]: https://www.w3.org/International/questions/qa-personal-names
[wcag-1-3-5]: https://www.w3.org/TR/WCAG22/#identify-input-purpose
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-8]: https://www.w3.org/TR/WCAG22/#accessible-authentication-minimum
