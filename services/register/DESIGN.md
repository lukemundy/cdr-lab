# Mock CDR Register: design

The [overview of the specification](design/README.md) gives the summary, the context, the areas, the goals and the non-goals. Each area file in `design/` gives the rules of its area.

## Operator console

### Functions

|---|---|---|

The console does the actions of the operator interface, with the same rules ([Operator interface](design/operator-interface.md#reg-0800)):

- Suspend, revoke, surrender and reinstate a data recipient legal entity.
- Activate, deactivate and remove a data recipient brand, a software product and a data holder brand.
- Remove a data holder legal entity.
- Register a client certificate of a software product: the operator gives one PEM certificate as text or as a file. Remove a client certificate: the operator selects one of the registered thumbprints.
- Issue a client certificate for a software product, and issue a server certificate for a data holder brand: the operator gives one PEM CSR as text or as a file.
- Save the CDR CA bundle as a file.
- Revoke a certificate of the mock CDR intermediate: the operator gives one PEM certificate, or the PEM chain of an issue, as text or as a file, and selects the reason.
- Publish the next key of a key set, and activate it.
- Create a data recipient or a data holder: the legal entity, with no brands. Add a brand to a legal entity, and add a software product to a data recipient brand. (See Decisions.)
- Change the details of a legal entity, a brand or a software product: the fields that a change in the operator interface replaces.

Rules for the actions:

- Before each action except a create, an add, a change, the registration of a client certificate and the issue of a certificate, the console asks the operator to confirm. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]: the registration, an issue and a change are reversible, and the Register checks each value of a create, an add, a change and an issue, and the operator can correct it. See Decisions.)
- The confirmation names the action, its record and each other record that the action also changes, with the status before and after. For example, it names each software product that a suspension makes INACTIVE. (Judgement: a cascade changes records that the operator did not select.)
- When the action or its cascade gives a final status (REMOVED, SURRENDERED), the confirmation says that this cannot be reversed. A revocation is an example: the data recipient can become ACTIVE again, but its software products cannot. ([CDS: Participant Statuses][cds-participant-statuses])
- The confirmation of a reinstatement says that the software products do not change, and that the operator must activate each one. ([CDS: Status Mapping][cds-status-mapping])
- The confirmation of a key publish names the previous key that leaves the JWKS, if there is one. The confirmation of a key activation names the key that starts to sign and the key that becomes the previous key, and says that a participant with an old copy of the JWKS cannot verify the new signatures. (Judgement: an activation cannot be reversed, and a publish deletes a key. See [the two steps of a key change](design/authorisation-server.md#reg-0468).)
- The console offers only the actions that the rules permit now for the record. The console API gives these actions (see Console API). The UI has no copy of the rules.
- When the Register refuses an action, the console shows the `detail` of the problem in the dialog or the form of the action. A refused PEM certificate shows as the error of its field. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- After an action, also after a refused action, each view shows the current state without a reload of the page. (Judgement: a refusal often means that the view is not current.)
- The console never sends an action again by itself. (Judgement: an automatic second request can apply a change that the operator did not confirm.)
- The console shows the result of a successful action in a toast. The view also shows the new state, so the toast is never the only place of the result. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]) An issue shows its result in its dialog, not in a toast (see the rules for the CDR certificates).
- For a certificate file or a CSR file, the console puts the text of the file in the text field of the certificate or the CSR, and the operator can see it before the action. The console does not read the certificate or the CSR: the Register checks it. (Judgement: one value to send, and one set of rules in the Register.)

Rules for the CDR certificates:

- The software product card has the button "Issue certificate" after "Register certificate", in its action row. The data holder brand card has the button "Issue certificate" after its status actions, in its action row. The console shows each button only when the console API permits the action (see Console API). ([Design system DESIGN.md][ui-design]: buttons that act on the record are in its action row)
- The side navigation has the section "CDR certificates", after "Signing keys", with its own page. The page shows the CDR CA bundle with a "Save as file" button, then a "Revoke certificate" button in an action row, then the list of revocations. ([Design system DESIGN.md][ui-design]; see Decisions)
- An issue is a dialog with one PEM CSR field and a CSR file input, with the same rules as the registration of a client certificate. The help text of the CSR field says that the Register needs only the CSR and never the private key. ([ACCC CPS, section 3.2.1][accc-cps]: the CA never sees the private key)
- The help text of the CSR field gives the subject that the Register requires. For a client certificate: the CN is the name of the software product, and the O is the name of its brand, with the names of the record. For a server certificate: the CN is the primary DNS name, and the O is the name of the brand or of its legal entity, with the names of the records. ([CDS: Certificate Signing Request Profile][cds-csr-profile], [WCAG 2.2, SC 3.3.2][wcag-3-3-2]; see Decisions) The console does not compare the CSR with these names: the Register checks them.
- When the Register refuses the CSR (400), the console shows the `detail` of the problem as the error of the CSR field. A different refusal, for example a REMOVED record or a stale version (409), shows in an alert in the dialog. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- After a successful issue, the dialog stays open and shows the result in place of the form: a heading that names the certificate type and the record, the thumbprint of the certificate, the PEM chain in a text area that the operator cannot change, a "Save as file" button and a "Close" button. The focus moves to the heading of the result. ([WAI-ARIA APG: Dialog][apg-dialog]; see Decisions)
- The result says that the chain is the certificate and then the intermediate certificate, and that the Register does not keep the certificate, so the console cannot show it again. It tells the operator to save the chain before they close the dialog. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [Judgement: the response of an issue](design/certificate-management.md#reg-0644))
- The result of a client certificate issue says that the Register registered the certificate for the software product, and that the token endpoint accepts it now. The result of a server certificate issue says that the Register does not register a server certificate, and that a TLS client that trusts the CDR CA bundle accepts it. ([Operator interface](design/operator-interface.md#reg-0800))
- The operator closes the result with "Close" or the Escape key, with no confirmation, also when they did not save the chain. (See Decisions.)
- The views get the current state after the issue, while the result shows: for example, the software product card shows the new thumbprint.
- "Save as file" saves the PEM text that the Register gave, with no change, as a file with the extension `.pem`: `client-certificate-{softwareProductId}.pem`, `server-certificate-{dataHolderBrandId}.pem` and `cdr-ca-bundle.pem`. The save sends no request to the Register, and needs no change to the console CSP. ([HTML: Downloading resources][html-download]; see Decisions)
- "Revoke certificate" opens a dialog with a PEM certificate field, a certificate file input and a reason. The reason is a native select with no selected value first, and the values of the [Operator interface](design/operator-interface.md#reg-0800) in words, for example "Key compromise" for `keyCompromise`. The help text of the certificate field says that the field accepts one PEM certificate, or the file of an issue, and that the Register revokes the first certificate. (Judgement: a reason has no usual value, see the rule for a select in a create.)
- The submit button of the revocation dialog is "Continue". The Register checks the request with no change and gives the certificate (see Console API). When it refuses the request, the dialog shows each field error at its field, with the error summary at its top, as a form does, and a refusal that names no field in the summary. A certificate that is already revoked is a refusal with no field (409). (See Decisions.)
- After a successful check, the console asks the operator to confirm. The confirmation names the certificate with its subject, its serial number, its thumbprint and the end of its validity, and names the reason. It says that the CRL lists the certificate, that the Register refuses it at once and other relying parties within 1 minute, that a registration of the certificate for a software product does not change, and that the revocation cannot be reversed. ([a revocation takes effect at the Register at once](design/certificate-revocation.md#reg-0725); see Decisions)
- After a successful revocation, the console shows the result in a toast, and the list of revocations shows the revocation first.

Rules for a create, an add and a change:

- Each create, add and change is a form on its own page, with its own URL. It is not a dialog. A URL of a form is never the URL of a record, so that a record with each ID, for example `new`, keeps its detail. (Judgement: the records have many fields. See Decisions.)
- The list of data recipients has the action "Create data recipient", and the list of data holders has "Create data holder". The detail of a participant has the actions "Change details" and "Add brand" for the legal entity, "Change brand" for each brand, "Add software product" for each data recipient brand, and "Change software product" for each software product. The console shows these actions for each record in each status. (See Decisions.) Each of these actions is a link directly after the heading of its record, and "Add software product" is at the heading "Software products" of its brand. The status actions and "Register certificate" are buttons after the details of their record and before its child records. ([Design system DESIGN.md][ui-design]) Each brand and each software product is a card that starts collapsed, with its status in its title bar. "Brands" and each "Software products" heading have "Show all" and "Hide all". After a form, the card of the record that the form changed or added is open. ([Design system DESIGN.md][ui-design])
- A create and an add give each field of one record in the format of the load file, with the identifiers and the status, and with no child records and no client certificates. ([Operator interface](design/operator-interface.md#reg-0800))
- For each identifier of a create or an add, the form shows a new random UUID ([RFC 9562, section 5.4][rfc9562-v4]) in a field that the operator can change. The operator can keep it, or give a known identifier, for example the brand ID in the configuration of a lab service. A data holder brand gets a UUID in its `dataHolderBrandId` field, and its `interimId` field is empty. (See Decisions.)
- The status of a create or an add is a field. Its values are the status values of the record of the [participant records](design/participants.md#reg-0200), and ACTIVE is selected first. The status rules apply: for example, a software product of an INACTIVE brand must be INACTIVE or REMOVED. (See Decisions.)
- A change form shows the current value of each field, and shows the identifiers of the record as text that the operator cannot change. The console sends each field of the change, as a change in the operator interface does: an optional field with no value has no value after the change. ([Operator interface](design/operator-interface.md#reg-0800))
- A change and an add send the version of the participant that the form loaded. A create has no version. A stale version gives 409, and the Register changes nothing (see the row "Action with a stale view" in Decisions).
- A choice of one value from a fixed set (the accreditation level, the status) is a native select. A choice from a fixed set of more than one value (the industries of a data holder brand) is a group of checkboxes. A list of values that the operator writes (the redirect URIs, the scopes) is a text area with one value on each line. The endpoints of a data holder brand are a group of fields with a legend. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; see Decisions)
- In a create, a select of a value that is required and that has no usual value (the accreditation level) has no selected value, so that the operator chooses it. (Judgement: a value that the console selects can go to the Register with no decision of the operator.)
- The label of each field that the console API schema does not require says "(optional)". The help text of the identifier fields of a data holder brand says that a brand with no `dataHolderBrandId` needs only its public details. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [GOV.UK Design System: Question pages][govuk-question-pages] marks the optional fields, not the required fields)
- The console removes the white space at the start and at the end of each value and of each line of a list, and ignores an empty line. It sends an optional field that has no value as no field. (Judgement: a value that the operator pastes often has white space, and a space at the end of a URI or an identifier is never intended.)
- The console does not check the values. The Register checks them with the rules of the participant model and the status rules. The UI has no copy of the rules. (Judgement: one set of rules, in the domain ([ADR 0013][adr-0013]).)
- When the Register refuses a value of a field, the console shows the `detail` of the field error as the error message of that field. The error of an item of a list shows at the field of the list. A refusal that names no field shows in the error summary of the form. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1], [SC 3.3.3][wcag-3-3-3])
- After a refusal, the form shows an error summary at its top, and the focus moves to the summary. The summary gives the `detail` of the problem, and a link to each field that has an error. ([GOV.UK Design System: Error summary][govuk-error-summary]; see the error summary in the [design system DESIGN.md][ui-design] and Decisions)
- After a refusal, the form keeps each value that the operator gave, and the version that it loaded. When the version is stale, the operator opens the form again to see the current values. (Judgement: the operator does not lose their work, and the console never replaces the change of a different person or tab with no decision of the operator.)
- While the request runs, the submit button is busy, and the form sends no second request. The console never sends a create, an add or a change again by itself.
- After a success, the console goes to the detail of the participant and shows the result in a toast. The focus moves to the heading of the detail. ([WCAG 2.2, SC 2.4.3][wcag-2-4-3], [SC 4.1.3][wcag-4-1-3]; judgement: the form page no longer exists)
- Each form has a "Cancel" link back to the page that opened it. A cancel sends no request, and the console does not ask the operator to confirm it. (See Decisions.)

### Console API

- A create and an add are a `POST` on the path of the list of the record. They give 201 with no body, and a `Location` header with the path of the new record in the console API, which is the start of the paths of its actions. ([RFC 9110, section 15.3.2][rfc9110-201]; judgement: the same as the operator interface)
- A change is a `POST` on the path of the record, with `/change` at the end. Its body gives the version and each field of the change, with the names and the format of the operator interface. A field that is not in the body has no value after the change. (See Decisions.)
- The body of a create gives the fields of one record, as a create in the operator interface does, with no child records and no client certificates. The body of an add gives the version and the fields of one record, with no child records and no client certificates. A body that is not one JSON object, that has a field that the action does not accept, or that is larger than 1 MiB, gives 400. ([Operator interface](design/operator-interface.md#reg-0800))
- When the Register refuses a value in the body, the problem details give [the field errors of the problem details library](../../pkg/platform/problem/DESIGN.md#prob-0003): each with [a JSON Pointer to the field](../../pkg/platform/problem/DESIGN.md#prob-0004), for example `#/redirectUris`, and [a `detail` that the UI can show at the field](../../pkg/platform/problem/DESIGN.md#prob-0005).
- These refusals have field errors: a value that does not obey the participant model (400), an identifier that a different record has (409, at the field of the identifier), a legal entity ID that a legal entity has (409, at `#/legalEntityId`), and a status in the body that breaks a status rule (409, at `#/status`). The status code is the same as in the operator interface. A refusal that names no field, for example a stale version, has [no `errors`](../../pkg/platform/problem/DESIGN.md#prob-0007).
- The `detail` of a problem with field errors gives each problem that names no field. When each problem names a field, the `detail` is [the general `detail` of field errors](../../pkg/platform/problem/DESIGN.md#prob-0006).
- The registration of a client certificate gives the PEM certificate in the JSON body, with the rules of the operator interface for the certificate.
- An issue gives the version and the PEM CSR (`csr`) in the JSON body, with the rules of the operator interface for the CSR, the record and the certificate. A successful issue gives 200 with a JSON body: the PEM chain of the operator interface (`certificateChain`) and the thumbprint of the certificate in the base64url format of RFC 8705 `x5t#S256` (`thumbprint`). A CSR that the Register refuses gives 400, and the `detail` names each rule that it breaks, as in the operator interface. (See Decisions.)
- The CDR CA bundle is a `GET` that gives 200 with a JSON body: the PEM bundle of the operator interface (`caBundle`). (See Decisions.)
- The list of revocations is a `GET` in pages, with the rules of the other lists. Each revocation gives the serial number, the subject as an [RFC 4514][rfc4514] string, the thumbprint, the time of the revocation and the reason. The list is sorted by the time of the revocation, newest first, and then by the serial number. (See Decisions.)
- The console API gives a serial number in uppercase hexadecimal, with two digits for each byte and no separators, as `openssl x509 -serial` gives it. (See Decisions.)
- The check of a revocation and the revocation give the PEM certificate (`certificate`) and the reason (`reason`) in the JSON body, with the rules of the operator interface. The body has no version. A check makes no change. A successful check gives 200 with the subject, the serial number, the thumbprint and the end of the validity of the first certificate. A successful revocation gives 204. A value that the Register refuses gives a field error at `#/certificate` or `#/reason` (400), and a certificate that is already revoked gives 409 with no field errors, and its `detail` gives the reason of the revocation that exists. (See Decisions.)

The console API has these paths for the creates, the adds, the changes and the CDR certificates.

| Action | Method and path | Success |
|---|---|---|
| Create a data recipient | `POST /api/data-recipients` | 201 |
| Change the details of a data recipient | `POST /api/data-recipients/{legalEntityId}/change` | 204 |
| Add a brand to a data recipient | `POST /api/data-recipients/{legalEntityId}/brands` | 201 |
| Change the details of a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/change` | 204 |
| Add a software product to a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/software-products` | 201 |
| Change the details of a software product | `POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/change` | 204 |
| Create a data holder | `POST /api/data-holders` | 201 |
| Change the details of a data holder | `POST /api/data-holders/{legalEntityId}/change` | 204 |
| Add a brand to a data holder | `POST /api/data-holders/{legalEntityId}/brands` | 201 |
| Change the details of a data holder brand | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/change` | 204 |
| Issue a client certificate for a software product from a CSR, and register it | `POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue` | 200 |
| Issue a server certificate for a data holder brand from a CSR | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue` | 200 |
| Get the CDR CA bundle | `GET /api/cdr-ca-bundle` | 200 |
| Get one page of the revocations of the mock CDR intermediate | `GET /api/cdr-certificates/revocations` | 200 |
| Check a revocation with no change | `POST /api/cdr-certificates/check-revocation` | 200 |
| Revoke a certificate of the mock CDR intermediate | `POST /api/cdr-certificates/revoke` | 204 |

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Confirmation of a status change | A dialog that names the cascade, and says when a change cannot be reversed | [WCAG 2.2, SC 3.3.4][wcag-3-3-4] asks for a check or a confirmation for changes to data. Judgement: a cascade changes records that the operator did not select. |
| Creation and change of participants in the console | In scope: the console creates, adds and changes each record that the operator interface does, and does the status actions of brands and data holders, with the same use cases | Judgement: a lab user who tests their own data recipient software must register it first, and an HTTP client with the operator token is a barrier to the most frequent set-up tasks. The console API and the operator interface stay separate adapters on the same use cases ([ADR 0014][adr-0014]). |
| Status codes of field errors | The status code of the operator interface: 400 for a value that does not obey the participant model, 409 for an identifier in use or a status that breaks a status rule | Judgement: one rule for the status codes in the two interfaces (see the row "Refusal of a body"). The field errors are in the body, so the UI does not need a different code. |
| Identifiers of a create or an add in the console | The form shows a new random UUID for each new identifier, in a field that the operator can change. A data holder brand gets a `dataHolderBrandId`, and no `interimId`. | Judgement: the Register issues GUIDs in production ([CDS: Register API schemas][cds-register-schemas]), so a UUID is the usual value and the operator types nothing. A lab service gets its identifiers in its configuration before it starts, for example the brand ID of the mock bank, so the operator must be able to give a known value. A brand with a `dataHolderBrandId` is the usual brand. The Register refuses an identifier that a different record has (409). |
| Status of a new record in the console | A field, with ACTIVE selected first, and the status rules apply | Judgement: the operator interface permits each status in a create or an add (see the row "Status in the body of a create or an add"). With a fixed status of ACTIVE, an add under a parent that is not ACTIVE always fails, for example a software product of an INACTIVE brand. The field needs no rule in the UI: the Register refuses a status that breaks a rule, at the field of the status. |
| Child records in a create in the console | A create makes the legal entity only. The operator adds each brand and software product after the create. | The participant model does not require a brand for a legal entity, and the operator interface creates a legal entity with no brands. Judgement: each form then has the fields of one record, the add forms do the work of the child records, and each field error has a short pointer. |
| Values of a change form | The current values, and each field of the change, with the version that the form loaded. An add also sends the version. | Judgement: a change in the operator interface replaces each field ([RFC 9110, section 9.3.4][rfc9110-put]), so the form must send the fields that the operator did not change. The version stops a change that is based on an old view (see the row "Action with a stale view"). An add changes the legal entity, so it has the same check. |
| Path and method of a change in the console API | `POST` on the path of the record with `/change` at the end, with the version and the fields in the JSON body | Judgement: the same form as each other console action (see the row "Path of a certificate removal in the console API"). A `PUT` replaces the representation that a `GET` of the same path gives ([RFC 9110, section 9.3.4][rfc9110-put]), and the `GET` of a participant gives its detail with its child records, its version and its actions, not the fields of the change. A brand and a software product have no `GET`. |
| Response of a create or an add in the console API | 201 with a `Location` header and no body | [RFC 9110, section 9.3.3][rfc9110-post]: a `POST` that makes a resource sends 201 with a `Location`. Judgement: the same as the operator interface. The UI knows the identifiers, because it sent them. |
| Forms in the console | A page with its own URL for each create, add and change. A dialog only for the confirmation of an action. | Judgement: the add form of a software product has 14 fields and the add form of a data holder brand has 15, which is too much for a dialog. A page has its own URL and history entry, and Back leaves it as the operator expects. |
| Confirmation of a create, an add or a change | None | [WCAG 2.2, SC 3.3.4][wcag-3-3-4] is met when the Register checks the values and the operator can correct them, or when the change is reversible. A change is reversible by a second change. A create and an add change no other record. |
| Lists of values in a form | A text area with one value on each line for a list that the operator writes. A group of checkboxes for a fixed set. A native select for one value of a fixed set. | Judgement: a text area is one control with one label and one error, the operator can paste a list, and the keyboard and the screen reader read it line by line. A list of text fields needs add and remove buttons and a focus rule for each, and gives no better access. The checkboxes and the select show each value of the fixed set, with the keyboard support of the browser. |
| Error summary in a form | An error summary at the top of the form after each refusal, with the focus on it | WCAG 2.2 SC 3.3.1 needs only the error at the field ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]). Judgement: a software product form is longer than one screen, so an error at a field far down the page is not seen. The [GOV.UK Design System: Error summary][govuk-error-summary] shows a summary for each error, with a link to each field, and moves the focus to it. The summary also has the refusals that name no field, so the form has one place for errors. |
| Values after a refusal | The form keeps the values of the operator and the version that it loaded. It does not get the current values by itself. | Judgement: the operator does not lose their work. A form that takes a new version by itself would replace the change of a different person or tab with no decision of the operator, which the version check stops. |
| White space in a form | The console removes it at the start and at the end of each value and of each line, and ignores an empty line | Judgement: a pasted value often has white space or an empty last line. This is not a rule of the Register: a value with a space at the end is never intended. The Register still checks each value. |
| Change and add actions in the console | Shown for each record in each status | Judgement: the domain permits a change and an add in each status, and refuses only a value or a status that breaks a rule. A list of permitted changes in the console API would always give each change. |
| Cancel of a form | A link back, with no confirmation and no warning about values that the operator did not send | Judgement: a lab form is quick to fill again, and a warning before each navigation adds a rule for each route. |
| Optional fields in a form | The label says "(optional)". A required field has no mark. | [GOV.UK Design System: Question pages][govuk-question-pages]. The console API schema gives which fields are optional, so the UI copies no rule. |
| Actions that need a confirmation in the console | Each action except the registration of a client certificate and the issue of a certificate. The confirmation of a certificate removal does not say that the removal cannot be reversed. | [WCAG 2.2, SC 3.3.4][wcag-3-3-4] accepts a change that is reversible, or a check of the values that the operator can correct. The operator can remove a registered certificate. The Register checks each rule of a CSR, and the operator can reverse an issue: remove the registration of a client certificate, or revoke the certificate with the chain that the issue gave. A revocation cannot be reversed. The Register does not keep the PEM file, so the operator can reverse a removal only with a copy of it: the confirmation says this. A key change is not reversible, and a status change can change records that the operator did not select (see the row "Confirmation of a status change"). |
| Path of a certificate removal in the console API | `POST` on `certificates/{thumbprint}/remove`, with the version in the JSON body | Judgement: each console action is a `POST` with the version in a JSON body (see the row "Action with a stale view"). The content of a `DELETE` has no defined meaning ([RFC 9110, section 9.3.5][rfc9110-delete]), so the version cannot go in the body of a `DELETE`. |
| A key step that changes no key that exists | The confirmation still says that the step cannot be reversed. | Judgement: the operator interface has no step that removes a published key or activates a previous key again, so a publish with no previous key is also a step that cannot be reversed. |
| Result of a console action | 204, then the UI gets each view again | Judgement: the same response as the operator interface. One rule for each action is simpler than a different body for each action. |
| Place of the certificate actions in the console | "Issue certificate" in the action row of the software product card and of the data holder brand card. The CDR CA bundle and "Revoke certificate" on a page of their own, "CDR certificates", in the side navigation. | Judgement: an issue is an action on one record, and the design system puts an action in the action row of its record ([Design system DESIGN.md][ui-design]). The bundle and a revocation are not of a participant: a revocation can be of a certificate of the lab set-up, which has no participant in the Register ([Judgement: the form of a revocation request](design/certificate-revocation.md#reg-0741)). Thus they get their own page, as the key sets do. A revocation from a registered thumbprint is not possible: the Register keeps no certificate, and a revocation needs the certificate ([Judgement: the certificate identifies the certificate to revoke](design/certificate-revocation.md#reg-0740)). |
| Result of an issue in the console | The dialog shows the result: the thumbprint, the PEM chain in a text area that the operator cannot change, "Save as file" and "Close". No toast. | The response is the only copy of the certificate ([Judgement: the response of an issue](design/certificate-management.md#reg-0644)). A toast closes by itself, and the design system shows only a result that the view also shows in a toast (Design system DESIGN.md, row "Errors and toasts"; [WCAG 2.2, SC 2.2.1][wcag-2-2-1]). Judgement: the text area lets the operator see and copy the chain. The console does not read the certificate, so the Register gives the thumbprint, which the operator finds again in the software product card and in the history. |
| Close of the result of an issue | With "Close" or the Escape key, with no confirmation, also when the operator did not save the chain | Judgement: the operator can issue a new certificate from the same CSR at once, and remove the registration of a lost client certificate. A lost certificate is in no component, so it gives no access. The result tells the operator to save the chain, and a confirmation at each close adds a step to each issue. The Escape key closes a dialog in the [WAI-ARIA APG: Dialog][apg-dialog] pattern. |
| Help text of a CSR field | The subject that the Register requires, with the names of the record | [WCAG 2.2, SC 3.3.2][wcag-3-3-2]: instructions when the content needs input from the user. A CN or an O that does not agree is the most probable refusal, and the CDS gives both values ([CDS: Certificate Signing Request Profile][cds-csr-profile]). Judgement: the text gives the CDS rule and the names that the view already shows. It is not a check: the Register still checks the CSR. |
| Save of a certificate or the CA bundle in the console | The UI saves the PEM text that the Register gave as a `.pem` file, with no second request. The file name has the ID of the record. | The Register does not keep the certificate, so a second request cannot get it again. Judgement: one way to save for the three files. An ID is unique and has only characters that are safe in a file name, and a name can have other characters. The design system button and text area do the work, so the design system gets no new component while only this UI saves a file ([Design system DESIGN.md][ui-design]: the components that more than one UI uses). |
| Format of an issue and of the CA bundle in the console API | JSON, with the PEM text in a field, not `application/pem-certificate-chain` | Judgement: the console API is JSON, and its OpenAPI document and the generated types describe JSON bodies ([ADR 0017][adr-0017]). The thumbprint goes in the same response as the chain. The operator interface keeps the PEM content type for scripts. |
| Confirmation of a revocation | A check with no change first, then a confirmation that names the certificate with its subject, its serial number, its thumbprint and the end of its validity | The confirmation names its record (see the rules for the actions). The console does not read the certificate, so only the Register can name it. Judgement: the revocation applies also to the certificates of the lab set-up, for example the server certificate of the Register, and a revocation of the wrong certificate cannot be reversed with no host access. The check also shows each refusal at its field before the confirmation. The revocation does each check again, so a change between the check and the revocation is safe. The row "Cascade in the confirmation" refuses a preview request because the detail already gives its data. No view gives the data of a certificate. |
| Version in a revocation in the console API | None | Judgement: a revocation is not of a participant or a key set. The only state that it depends on is whether the certificate is revoked, and a second revocation gives 409 with no change ([Judgement: 409 for a second revocation](design/certificate-revocation.md#reg-0744)). Thus a stale view cannot cause a change that the operator did not confirm. |
| Errors of a CSR and of a revocation in the console | The `detail` of a refused CSR at the CSR field. Field errors at `#/certificate` and `#/reason` for a revocation. | Judgement: an issue has one field, so the `detail` is its error, as for the registration of a client certificate. A revocation has two fields, so it uses [the field errors of the console API](../../pkg/platform/problem/DESIGN.md#prob-0003). |
| Format of a serial number in the console | Uppercase hexadecimal, with two digits for each byte and no separators | Judgement: the CDS and the ACCC give no format. The operator compares the serial number in the confirmation and in the list with the output of `openssl x509 -serial`, which has this format. The Register keeps its own format in the database: the console API changes only the format of its responses. |

[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-apis
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-federation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-federation
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-archiving]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#archiving
[cds-get-oidc]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-openid-provider-config
[cds-get-jwks]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-jwks
[cds-get-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-statement-assertion-ssa
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-get-dh-brands-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands-summary
[cds-get-dh-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-statuses
[cds-get-sp-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-products-statuses
[cds-get-dr-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients-statuses
[cds-get-drs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients
[cds-openapi]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_register.json
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[cds-error-403-adr-not-active]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-adr-status-not-active
[cds-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-errors
[cds-known-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#standardise-register-api-error-codes
[cds-dr-calling-register]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-the-cdr-register
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-dcr-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#tocSregistrationproperties
[rfc3986-2]: https://www.rfc-editor.org/rfc/rfc3986#section-2
[fapi-as]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#authorization-server
[oidc-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-cert-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-cert-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cds-nfrs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#non-functional-requirements
[cds-register-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-endpoints
[cds-error-400-field-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid
[cds-error-invalid-industry]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-error-404-field-invalid-industry
[cds-cors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors
[cds-schema-brand-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_schemas_tocSdataholderbrandsummaryv2
[rfc9110-etag]: https://www.rfc-editor.org/rfc/rfc9110#name-etag
[rfc9111-no-cache]: https://www.rfc-editor.org/rfc/rfc9111#name-no-cache-2
[rfc7517-jwks]: https://www.rfc-editor.org/rfc/rfc7517#section-5
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[readme-sources]: ../../README.md#sources-for-requirements
[accc-known-issues]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29950003
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[accc-cert-mgmt]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management
[accc-mock-ssa-test]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/API/SSA/GetSoftwareStatementAssertion_Tests.cs#L416
[accc-mock-dr-ssa]: https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.SDK/Services/Register/SsaService.cs#L46-L56
[cds-registration-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-validation
[accc-mock-dh-statuses]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/RegisterStatusRepository.cs#L65-L77
[accc-brands-summary]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982
[fetch-cors]: https://fetch.spec.whatwg.org/#http-cors-protocol
[accc-mock-certs]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CertificateManagement/README.md
[research-accc-headers]: ../../docs/research/2026-09-25-accc-mock-register.md
[support-removed-sp]: https://cdr-support.zendesk.com/hc/en-us/articles/900001657906
[support-brands]: https://cdr-support.zendesk.com/hc/en-us/articles/900005881263
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[dsb-537]: https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/537
[adr-0003]: ../../docs/adr/0003-event-sourcing.md
[adr-0006]: ../../docs/adr/0006-telemetry.md
[adr-0011]: ../../docs/adr/0011-domain-data-as-typed-go.md
[adr-0012]: ../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0017]: ../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0019]: ../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[adr-0021]: ../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0022]: ../../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[adr-0010]: ../../docs/adr/0010-background-jobs-with-river.md
[pki-design]: ../../pki/DESIGN.md
[pki-design-keys]: ../../pki/DESIGN.md#keys
[rfc5280-crl]: https://www.rfc-editor.org/rfc/rfc5280#section-5
[rfc5280-crl-number]: https://www.rfc-editor.org/rfc/rfc5280#section-5.2.3
[rfc5280-crldp]: https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.13
[rfc2585-mime]: https://www.rfc-editor.org/rfc/rfc2585#section-4.2
[rfc9110-methods]: https://www.rfc-editor.org/rfc/rfc9110#section-9.1
[rfc9110-405]: https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
[rfc9110-503]: https://www.rfc-editor.org/rfc/rfc9110#name-503-service-unavailable
[pki-design-revocation]: ../../pki/DESIGN.md#revocation
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[accc-cp]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf
[accc-portal-guide]: https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf
[cds-cert-mgmt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-management
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[rfc5280-bounds]: https://www.rfc-editor.org/rfc/rfc5280#appendix-A
[rfc8555-chain]: https://www.rfc-editor.org/rfc/rfc8555#section-9.1
[rfc8446-cert]: https://www.rfc-editor.org/rfc/rfc8446#section-4.4.2
[ui-design]: ../../web/packages/ui/DESIGN.md
[ui-brands]: ../../web/packages/ui/DESIGN.md#brands
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
[rfc9110-delete]: https://www.rfc-editor.org/rfc/rfc9110#name-delete
[rfc9110-idempotent]: https://www.rfc-editor.org/rfc/rfc9110#name-idempotent-methods
[rfc9110-if-match]: https://www.rfc-editor.org/rfc/rfc9110#name-if-match
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-put]: https://www.rfc-editor.org/rfc/rfc9110#name-put
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#name-409-conflict
[rfc7396]: https://www.rfc-editor.org/rfc/rfc7396
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[adr-0013]: ../../docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[fetch-metadata]: https://w3c.github.io/webappsec-fetch-metadata/
[rfc9562-v4]: https://www.rfc-editor.org/rfc/rfc9562#section-5.4
[rfc9110-post]: https://www.rfc-editor.org/rfc/rfc9110#name-post
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-3-3-3]: https://www.w3.org/TR/WCAG22/#error-suggestion
[govuk-error-summary]: https://design-system.service.gov.uk/components/error-summary/
[govuk-question-pages]: https://design-system.service.gov.uk/patterns/question-pages/
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[apg-dialog]: https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/
[html-download]: https://html.spec.whatwg.org/multipage/links.html#downloading-resources
[rfc4514]: https://www.rfc-editor.org/rfc/rfc4514
[k8s-configmap-volume]: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
[k8s-init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[helm-values]: https://helm.sh/docs/chart_template_guide/values_files/
[dataholder-brand-configmap]: ../dataholder/design/deployment.md#dh-1345
[dataholder-register-statuses]: ../dataholder/design/registration.md#dh-0689
[adr-0020]: ../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[register-chart]: ../../charts/register/README.md
