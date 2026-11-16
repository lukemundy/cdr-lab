# Mock CDR Register: Operator console actions

This area gives the actions of the operator console, the rules for each action, and the actions on CDR certificates. The [forms](operator-console-forms.md) give the creates, the adds and the changes. The [change API](operator-console-change-api.md) and the [console API](operator-console-api.md) give the requests of the actions.

## Actions

- <a id="reg-1200"></a>**REG-1200** The console does the actions of the operator interface, with the same rules. For example, a suspension in the console has the cascades of a suspension in the operator interface. ([the rules of a status action](operator-interface.md#reg-0805))
- <a id="reg-1201"></a>**REG-1201** The lab operator can suspend, revoke, surrender and reinstate a data recipient legal entity. ([the transitions of a data recipient legal entity](participants.md#reg-0222))
- <a id="reg-1202"></a>**REG-1202** The lab operator can activate, deactivate and remove a data recipient brand, a software product and a data holder brand. ([the transitions of a software product](participants.md#reg-0223); [the transitions of a brand](participants.md#reg-0225))
- <a id="reg-1203"></a>**REG-1203** The lab operator can remove a data holder legal entity. ([the transitions of a data holder legal entity](participants.md#reg-0224))
- <a id="reg-1204"></a>**REG-1204** The lab operator can register a client certificate of a software product. The lab operator gives one PEM certificate as text or as a file. ([the registration of a client certificate](operator-interface.md#reg-0834); [one PEM certificate in a registration](operator-interface.md#reg-0835))
- <a id="reg-1205"></a>**REG-1205** The lab operator can remove a client certificate of a software product. The lab operator selects one of the registered thumbprints. ([the removal of a client certificate](operator-interface.md#reg-0834))
- <a id="reg-1206"></a>**REG-1206** The lab operator can issue a client certificate for a software product. The lab operator gives one PEM CSR as text or as a file. ([the issue of a client certificate](certificate-management.md#reg-0601))
- <a id="reg-1207"></a>**REG-1207** The lab operator can issue a server certificate for a data holder brand. The lab operator gives one PEM CSR as text or as a file. ([the issue of a server certificate](certificate-management.md#reg-0618))
- <a id="reg-1208"></a>**REG-1208** The lab operator can save the CDR CA bundle as a file. ([the CDR CA bundle for the lab operator](certificate-management.md#reg-0615))
- <a id="reg-1209"></a>**REG-1209** The lab operator can revoke a certificate of the mock CDR intermediate. The lab operator gives one PEM certificate, or the PEM chain of an issue, as text or as a file, and selects the reason. ([the revocation of a certificate](certificate-revocation.md#reg-0700); [the first certificate of a PEM chain](certificate-revocation.md#reg-0703))
- <a id="reg-1210"></a>**REG-1210** The lab operator can publish the next key of a key set, and activate it. ([the two steps of a key change](authorisation-server.md#reg-0468))

## Confirmation

- <a id="reg-1211"></a>**REG-1211** Before each action of this area, the console asks the lab operator to confirm. The registration of a client certificate and the issue of a certificate have no confirmation. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]; [Judgement: a confirmation for each action except a registration and an issue](#reg-1280))
- <a id="reg-1212"></a>**REG-1212** The confirmation names the action, its record and each other record that the action also changes, with the status before and after. For example, the confirmation of a suspension names each software product that becomes INACTIVE. ([the changes of a permitted status action](operator-console-api.md#reg-1134); [Judgement: the confirmation names each record that a cascade changes](#reg-1281))
- <a id="reg-1213"></a>**REG-1213** When the action or its cascade gives a final status (REMOVED or SURRENDERED), the confirmation says that the action cannot be reversed. For example, after a revocation the data recipient can become ACTIVE again, but its software products cannot. ([CDS: Participant Statuses][cds-participant-statuses]; [the cascade of a revocation](participants.md#reg-0228))
- <a id="reg-1214"></a>**REG-1214** The confirmation of a reinstatement says that the software products do not change, and that the lab operator must activate each one. ([CDS: Status Mapping][cds-status-mapping])
- <a id="reg-1215"></a>**REG-1215** The confirmation of a key publish names the previous key that leaves the JWKS, if there is one. ([when the previous key leaves the JWKS](authorisation-server.md#reg-0469); [Judgement: a key step cannot be reversed](#reg-1282))
- <a id="reg-1216"></a>**REG-1216** The confirmation of a key activation names the key that starts to sign and the key that becomes the previous key. It says that a participant with an old copy of the JWKS cannot verify the new signatures. ([the two steps of a key change](authorisation-server.md#reg-0468); [Judgement: a key step cannot be reversed](#reg-1282))
- <a id="reg-1217"></a>**REG-1217** The confirmation of each key step says that the step cannot be reversed, also when the step changes no key that exists. For example, a publish with no previous key has this text. ([Judgement: a key step cannot be reversed](#reg-1282))
- <a id="reg-1218"></a>**REG-1218** The confirmation of a client certificate removal says that the lab operator must register the PEM certificate again to trust it again. It does not say that the removal cannot be reversed. ([Judgement: a confirmation for each action except a registration and an issue](#reg-1280))

## Results and refusals

- <a id="reg-1219"></a>**REG-1219** The console offers only the actions that the console API permits for the current state of the record. The UI has no copy of the rules. ([the permitted status actions](operator-console-api.md#reg-1132); [the permitted certificate actions](operator-console-api.md#reg-1136); [Judgement: the domain rules give the permitted actions and their changes](operator-console-api.md#reg-1181))
- <a id="reg-1220"></a>**REG-1220** When the Register refuses an action, the console shows the `detail` of the problem in the dialog of the action. For example, a stale version shows its `detail` in the confirmation dialog. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="reg-1221"></a>**REG-1221** When the Register refuses a PEM certificate, the console shows the `detail` as the error of the certificate field. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="reg-1222"></a>**REG-1222** After an action, also after a refused action, each view shows the current state with no reload of the page. ([Judgement: no body after an action](operator-console-api.md#reg-1183); [Judgement: each view gets the current state after a refusal](#reg-1283))
- <a id="reg-1223"></a>**REG-1223** The console never sends an action again by itself. ([Judgement: no automatic second request](#reg-1284))
- <a id="reg-1224"></a>**REG-1224** The console shows the result of a successful action in a toast. The view also shows the new state, so the toast is never the only place of the result. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="reg-1225"></a>**REG-1225** For a certificate file or a CSR file, the console puts the text of the file in the text field of the certificate or the CSR. The lab operator can see the text before the action. ([Judgement: the console puts the text of a file in its field](#reg-1285))
- <a id="reg-1226"></a>**REG-1226** The console does not read a certificate or a CSR, and does not compare a CSR with the names of the record. The Register checks them. ([the CSR rules of the Register](certificate-management.md#reg-0604); [Judgement: the console puts the text of a file in its field](#reg-1285))

## Certificate issue

- <a id="reg-1227"></a>**REG-1227** The software product card has the button "Issue certificate" after "Register certificate", in its action row. ([the action row of a record](../../../web/packages/ui/design/components.md#ui-0246); [Judgement: an issue on the record card, and other certificate actions on one page](#reg-1286))
- <a id="reg-1228"></a>**REG-1228** The data holder brand card has the button "Issue certificate" after its status actions, in its action row. ([the action row of a record](../../../web/packages/ui/design/components.md#ui-0246); [Judgement: an issue on the record card, and other certificate actions on one page](#reg-1286))
- <a id="reg-1229"></a>**REG-1229** The console shows each "Issue certificate" button only when the console API permits the action. For example, a REMOVED software product has no button. ([the permitted certificate actions](operator-console-api.md#reg-1136); [the server certificate actions of a brand](operator-console-api.md#reg-1137))
- <a id="reg-1230"></a>**REG-1230** The side navigation has the section "CDR certificates" after "Signing keys", with its own page. ([the structure of a console](../../../web/packages/ui/design/layout.md#ui-0071); [Judgement: an issue on the record card, and other certificate actions on one page](#reg-1286))
- <a id="reg-1231"></a>**REG-1231** The CDR certificates page shows the CDR CA bundle with a "Save as file" button, then a "Revoke certificate" button in an action row, then the list of the revocations. ([the action row of a record](../../../web/packages/ui/design/components.md#ui-0246); [Judgement: an issue on the record card, and other certificate actions on one page](#reg-1286))
- <a id="reg-1232"></a>**REG-1232** An issue is a dialog with one PEM CSR field and a CSR file input. The field and the input have the rules of the registration of a client certificate. ([the request body of an issue](certificate-management.md#reg-0602))
- <a id="reg-1233"></a>**REG-1233** The help text of the CSR field says that the Register needs only the CSR, and never the private key. ([ACCC CPS, section 3.2.1][accc-cps]; [the Register never gets the private key](certificate-management.md#reg-0603))
- <a id="reg-1234"></a>**REG-1234** For a client certificate, the help text of the CSR field gives the subject that the Register requires, with the names of the record. The CN is the name of the software product, and the O is the name of its brand. ([CDS: Certificate Signing Request Profile][cds-csr-profile]; [WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [Judgement: the help text of a CSR field gives the subject that the Register requires](#reg-1287))
- <a id="reg-1235"></a>**REG-1235** For a server certificate, the help text of the CSR field gives the subject that the Register requires, with the names of the records. The CN is the primary DNS name, and the O is the name of the brand or of its legal entity. ([CDS: Certificate Signing Request Profile][cds-csr-profile]; [WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [Judgement: the help text of a CSR field gives the subject that the Register requires](#reg-1287))
- <a id="reg-1236"></a>**REG-1236** When the Register refuses the CSR (400), the console shows the `detail` of the problem as the error of the CSR field. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [Judgement: the `detail` as the CSR error, and field errors for a revocation](#reg-1291))
- <a id="reg-1237"></a>**REG-1237** When the Register refuses an issue for a different reason, the console shows the `detail` in an alert in the dialog. For example, a REMOVED record or a stale version (409) shows in the alert. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="reg-1238"></a>**REG-1238** After a successful issue, the dialog stays open and shows the result in place of the form. The console shows no toast for an issue. ([Judgement: the result of an issue shows in its dialog, with no toast](#reg-1288))
- <a id="reg-1239"></a>**REG-1239** The result of an issue has these parts, in this sequence. ([Judgement: the result of an issue shows in its dialog, with no toast](#reg-1288))
  - a heading that names the certificate type and the record
  - the thumbprint of the certificate
  - the PEM chain, in a text area that the lab operator cannot change
  - a "Save as file" button
  - a "Close" button
- <a id="reg-1240"></a>**REG-1240** When the result shows, the focus moves to the heading of the result. ([WAI-ARIA APG: Dialog][apg-dialog])
- <a id="reg-1241"></a>**REG-1241** The result says that the chain is the certificate and then the intermediate certificate. ([the response of an issue](certificate-management.md#reg-0612); [WCAG 2.2, SC 3.3.2][wcag-3-3-2])
- <a id="reg-1242"></a>**REG-1242** The result says that the Register does not keep the certificate, so the console cannot show it again. It tells the lab operator to save the chain before they close the dialog. ([the Register keeps only the thumbprint](certificate-management.md#reg-0614); [WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [Judgement: an issue gives 200 and the certificate chain with no root](certificate-management.md#reg-0644))
- <a id="reg-1243"></a>**REG-1243** The result of a client certificate issue says that the Register registered the certificate for the software product, and that the token endpoint accepts it at once. ([the registration of an issued client certificate](certificate-management.md#reg-0610); [the token endpoint accepts the certificate at once](certificate-management.md#reg-0611))
- <a id="reg-1244"></a>**REG-1244** The result of a server certificate issue says that the Register does not register a server certificate, and that a TLS client that trusts the CDR CA bundle accepts it. ([no registration of a server certificate](certificate-management.md#reg-0629))
- <a id="reg-1245"></a>**REG-1245** The lab operator closes the result with "Close" or the Escape key, with no confirmation, also when they did not save the chain. ([WAI-ARIA APG: Dialog][apg-dialog]; [Judgement: the result of an issue closes with no confirmation](#reg-1289))
- <a id="reg-1246"></a>**REG-1246** While the result shows, the views get the current state after the issue. For example, the software product card shows the new thumbprint. ([Judgement: no body after an action](operator-console-api.md#reg-1183))
- <a id="reg-1247"></a>**REG-1247** "Save as file" saves the PEM text that the Register gave, with no change, as a file with the extension `.pem`. ([HTML: Downloading resources][html-download]; [Judgement: the browser saves the PEM text, with a file name from the ID](#reg-1290))
- <a id="reg-1248"></a>**REG-1248** The name of a saved file is one of these. ([Judgement: the browser saves the PEM text, with a file name from the ID](#reg-1290))
  - `client-certificate-{softwareProductId}.pem`
  - `server-certificate-{dataHolderBrandId}.pem`
  - `cdr-ca-bundle.pem`
- <a id="reg-1249"></a>**REG-1249** A save sends no request to the Register, and needs no change to the Content Security Policy of the console. ([the Content Security Policy of the web UI](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007); [Judgement: the browser saves the PEM text, with a file name from the ID](#reg-1290))

## Certificate revocation

- <a id="reg-1250"></a>**REG-1250** "Revoke certificate" opens a dialog with a PEM certificate field, a certificate file input and a reason. ([the request body of a revocation](certificate-revocation.md#reg-0702))
- <a id="reg-1251"></a>**REG-1251** The reason is a native select with no selected value first. Its values are the reasons of the operator interface in words. For example, "Key compromise" is the word for `keyCompromise`. ([the request body of a revocation](certificate-revocation.md#reg-0702); [Judgement: no reason selected first](#reg-1293))
- <a id="reg-1252"></a>**REG-1252** The help text of the certificate field says that the field accepts one PEM certificate or the file of an issue, and that the Register revokes the first certificate. ([the first certificate of a PEM chain](certificate-revocation.md#reg-0703); [WCAG 2.2, SC 3.3.2][wcag-3-3-2])
- <a id="reg-1253"></a>**REG-1253** The submit button of the revocation dialog is "Continue". It sends [the check of a revocation](operator-console-change-api.md#reg-1343), which changes nothing and gives the certificate. ([Judgement: a check by the Register before the confirmation of a revocation](#reg-1292))
- <a id="reg-1254"></a>**REG-1254** When the Register refuses the check, the dialog shows each field error at its field, with the error summary at its top, as a form does. A refusal that names no field shows in the summary. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [Judgement: the `detail` as the CSR error, and field errors for a revocation](#reg-1291))
- <a id="reg-1255"></a>**REG-1255** A certificate that is already revoked is a refusal with no field (409). ([a second revocation gives 409](certificate-revocation.md#reg-0706))
- <a id="reg-1256"></a>**REG-1256** After a successful check, the console asks the lab operator to confirm. The confirmation names the certificate with its subject, its serial number, its thumbprint and the end of its validity, and it names the reason. ([Judgement: a check by the Register before the confirmation of a revocation](#reg-1292))
- <a id="reg-1257"></a>**REG-1257** The confirmation of a revocation says these four facts. ([a revocation takes effect at the Register at once](certificate-revocation.md#reg-0725); [a revocation does not remove a registration](certificate-revocation.md#reg-0709); [Judgement: a check by the Register before the confirmation of a revocation](#reg-1292))
  - The CRL lists the certificate.
  - The Register refuses the certificate at once, and other relying parties within 1 minute.
  - A registration of the certificate for a software product does not change.
  - The revocation cannot be reversed.
- <a id="reg-1258"></a>**REG-1258** After a successful revocation, the console shows the result in a toast, and the list of the revocations shows the revocation first. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [the revocations on the CDR certificates page](operator-console.md#reg-1046))

## Judgements

### <a id="reg-1280"></a>REG-1280 A confirmation for each action except a registration and an issue

- Kind: choice
- Basis: Step 3.2. [WCAG 2.2, SC 3.3.4][wcag-3-3-4] accepts a change that is reversible, or a check of the values that the lab operator can correct.
- Reason: The lab operator can remove a registered certificate, and can reverse an issue with a removal or a revocation. A revocation, a key step and a status change cannot be reversed, or change records that the lab operator did not select. The Register does not keep the PEM file of a removed certificate, so only a copy of it can reverse a removal.

### <a id="reg-1281"></a>REG-1281 The confirmation names each record that a cascade changes

- Kind: choice
- Basis: Step 3.2. [WCAG 2.2, SC 3.3.4][wcag-3-3-4] asks for a check or a confirmation for changes to data. The cascades come from the status rules ([CDS: Status Mapping][cds-status-mapping]).
- Reason: A cascade changes records that the lab operator did not select. The confirmation shows each of them before the change.

### <a id="reg-1282"></a>REG-1282 A key step cannot be reversed

- Kind: choice
- Basis: Step 3.3. The operator interface has no step that removes a published key or activates a previous key again ([the two steps of a key change](authorisation-server.md#reg-0468)).
- Reason: An activation changes the key that participants must trust, and a publish deletes the previous key. A publish with no previous key is also a step that the lab operator cannot reverse.

### <a id="reg-1283"></a>REG-1283 Each view gets the current state after a refusal

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: A refusal often means that the view is not current. For example, a stale version means that a different tab changed the participant.

### <a id="reg-1284"></a>REG-1284 No automatic second request

- Kind: choice
- Basis: Step 3.3. A console action is a `POST`, which is not idempotent ([RFC 9110, section 9.2.2][rfc9110-idempotent]).
- Reason: An automatic second request can apply a change that the lab operator did not confirm.

### <a id="reg-1285"></a>REG-1285 The console puts the text of a file in its field

- Kind: choice
- Basis: Step 3.3. The Register checks each certificate and each CSR ([the CSR rules of the Register](certificate-management.md#reg-0604)).
- Reason: The console then has one value to send, and the rules are only in the Register.

### <a id="reg-1286"></a>REG-1286 An issue on the record card, and other certificate actions on one page

- Kind: choice
- Basis: Step 3.3. The design system puts an action on one record in the action row of that record ([the action row of a record](../../../web/packages/ui/design/components.md#ui-0246)).
- Reason: An issue is an action on one record. The CA bundle and a revocation are not of a participant, because [a revocation can be of a certificate of the lab set-up](certificate-revocation.md#reg-0701), so they get their own page, as the key sets do. A revocation needs the certificate, which the Register does not keep, so a revocation from a registered thumbprint is not possible ([Judgement: the certificate identifies the certificate to revoke](certificate-revocation.md#reg-0740)).

### <a id="reg-1287"></a>REG-1287 The help text of a CSR field gives the subject that the Register requires

- Kind: choice
- Basis: Step 3.2. [WCAG 2.2, SC 3.3.2][wcag-3-3-2] asks for instructions when the content needs input from the user. The CDS gives the CN and the O ([CDS: Certificate Signing Request Profile][cds-csr-profile]).
- Reason: A CN or an O that does not agree is the most probable refusal. The text gives the CDS rule and the names that the view already shows. It is not a check, because the Register still checks the CSR.

### <a id="reg-1288"></a>REG-1288 The result of an issue shows in its dialog, with no toast

- Kind: choice
- Basis: Step 3.3. The response is the only copy of the certificate ([Judgement: an issue gives 200 and the certificate chain with no root](certificate-management.md#reg-0644)). A toast closes by itself ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]; [only a success shows in a toast](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0336)).
- Reason: The text area lets the lab operator see and copy the chain. The console does not read the certificate, so the Register gives the thumbprint, which the lab operator finds again in the software product card and in the history.

### <a id="reg-1289"></a>REG-1289 The result of an issue closes with no confirmation

- Kind: choice
- Basis: Step 3.3. The Escape key closes a dialog in the [WAI-ARIA APG: Dialog][apg-dialog] pattern.
- Reason: The lab operator can issue a new certificate from the same CSR at once, and remove the registration of a lost client certificate. A lost certificate is in no component, so it gives no access. A confirmation at each close adds a step to each issue.

### <a id="reg-1290"></a>REG-1290 The browser saves the PEM text, with a file name from the ID

- Kind: choice
- Basis: Step 3.3. The Register does not keep the certificate, so a second request cannot get it again ([the Register keeps only the thumbprint](certificate-management.md#reg-0614)).
- Reason: One way to save is the least complex for the three files. An ID is unique and has only characters that are safe in a file name, but a name can have other characters. The design system button and text area do the work, so the design system gets no new component for one UI ([design system overview](../../../web/packages/ui/design/README.md)).

### <a id="reg-1291"></a>REG-1291 The `detail` as the CSR error, and field errors for a revocation

- Kind: choice
- Basis: Step 3.3. [The field errors of the problem details library](../../../pkg/platform/problem/DESIGN.md#prob-0003) apply to a request with more than one field.
- Reason: An issue has one field, so the `detail` is its error, as for the registration of a client certificate. A revocation has two fields, so it uses field errors.

### <a id="reg-1292"></a>REG-1292 A check by the Register before the confirmation of a revocation

- Kind: choice
- Basis: Step 3.3. The confirmation names its record ([the confirmation of an action](#reg-1212)), but the console does not read the certificate ([the console reads no certificate](#reg-1226)).
- Reason: Only the Register can name the certificate, and a revocation of the wrong certificate, for example the server certificate of the Register, cannot be reversed with no host access. The check also shows each refusal at its field before the confirmation. The revocation does each check again, so a change between the check and the revocation is safe.

### <a id="reg-1293"></a>REG-1293 No reason selected first

- Kind: choice
- Basis: Step 3.3. RFC 5280 gives no usual reason for a revocation ([the request body of a revocation](certificate-revocation.md#reg-0702)).
- Reason: A value that the console selects can go to the Register with no decision of the lab operator. A revocation cannot be reversed, so the lab operator chooses the reason.

[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[apg-dialog]: https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/
[html-download]: https://html.spec.whatwg.org/multipage/links.html#downloading-resources
[rfc9110-idempotent]: https://www.rfc-editor.org/rfc/rfc9110#name-idempotent-methods
