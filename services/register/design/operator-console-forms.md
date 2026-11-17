# Mock CDR Register: Operator console forms

This area gives the forms of the operator console that create, add and change participant records. The [operator console](operator-console.md) gives the views, and the [change API](operator-console-change-api.md) gives the requests of the forms.

## Forms

- <a id="reg-1400"></a>**REG-1400** The console does the creates, the adds and the changes of the operator interface, with the same rules. For example, an add in the console has the status rules of an add in the operator interface. ([the creates, the adds and the changes of the operator interface](operator-interface.md#reg-0809); [Judgement: the console creates and changes participants](#reg-1480))
- <a id="reg-1401"></a>**REG-1401** A create of a data recipient or of a data holder makes the legal entity with no brands. For example, the lab operator adds each brand after the create. ([Judgement: a create of the legal entity only](#reg-1481))
- <a id="reg-1402"></a>**REG-1402** The lab operator can add a brand to a legal entity, and a software product to a data recipient brand. ([the add of a brand or a software product](operator-interface.md#reg-0811))
- <a id="reg-1403"></a>**REG-1403** The lab operator can change the details of a legal entity, a brand or a software product. A change form has the fields that a change in the operator interface replaces. ([the body of a change in the operator interface](operator-interface.md#reg-0818))
- <a id="reg-1404"></a>**REG-1404** Each create, add and change is a form on its own page, with its own URL. It is not a dialog. ([Judgement: forms on their own pages](#reg-1482))
- <a id="reg-1405"></a>**REG-1405** A URL of a form is never the URL of a record. For example, a record with the ID `new` keeps its detail. ([Judgement: forms on their own pages](#reg-1482))
- <a id="reg-1406"></a>**REG-1406** Before a create, an add or a change, the console does not ask the lab operator to confirm. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]; [Judgement: no confirmation of a form](#reg-1483))

## Links to the forms

- <a id="reg-1407"></a>**REG-1407** The list of the data recipients has the action "Create data recipient". The list of the data holders has the action "Create data holder". ([Judgement: the console creates and changes participants](#reg-1480))
- <a id="reg-1408"></a>**REG-1408** The detail of a participant has these links to the forms. ([Judgement: the console creates and changes participants](#reg-1480))
  - "Change details" and "Add brand" for the legal entity
  - "Change brand" for each brand
  - "Add software product" for each data recipient brand
  - "Change software product" for each software product
- <a id="reg-1409"></a>**REG-1409** The console shows each link to a form for each record in each status. For example, a REMOVED brand has "Change brand". ([Judgement: links to the forms in each status](#reg-1484))
- <a id="reg-1410"></a>**REG-1410** Each link to a form is directly after the heading of its record. ([the links of a section](../../../web/packages/ui/design/components.md#ui-0231))
- <a id="reg-1411"></a>**REG-1411** "Add software product" is at the heading "Software products" of its brand. ([the links of a section](../../../web/packages/ui/design/components.md#ui-0231))

## Fields

- <a id="reg-1415"></a>**REG-1415** A create form and an add form give each field of one record in the format of the load file, with the identifiers and the status. ([the body of a create](operator-console-change-api.md#reg-1325); [the body of an add](operator-console-change-api.md#reg-1326))
- <a id="reg-1416"></a>**REG-1416** A create form and an add form have no fields for a child record or a client certificate. ([the body of a create](operator-console-change-api.md#reg-1325); [Judgement: a create of the legal entity only](#reg-1481))
- <a id="reg-1417"></a>**REG-1417** For each identifier of a create or an add, the form shows a new random UUID in a field that the lab operator can change. For example, the lab operator gives the brand ID from the configuration of a lab service. ([RFC 9562, section 5.4][rfc9562-v4]; [Judgement: a form shows a new random UUID for each identifier](#reg-1485))
- <a id="reg-1418"></a>**REG-1418** The add form of a data holder brand shows the UUID in the `dataHolderBrandId` field. Its `interimId` field is empty. ([the identifiers of a data holder brand](participants.md#reg-0208); [Judgement: a form shows a new random UUID for each identifier](#reg-1485))
- <a id="reg-1419"></a>**REG-1419** The status of a create or an add is a field. Its values are the status values of the record, and ACTIVE is selected first. ([the participant records](participants.md#reg-0200); [a status in a create or an add](operator-interface.md#reg-0814); [Judgement: ACTIVE selected first](#reg-1486))
- <a id="reg-1420"></a>**REG-1420** The status rules apply to the status field. For example, a software product of an INACTIVE brand must be INACTIVE or REMOVED. ([the Register checks the full legal entity](operator-interface.md#reg-0823); [Judgement: ACTIVE selected first](#reg-1486))
- <a id="reg-1421"></a>**REG-1421** A change form shows the current value of each field. ([a field that is not in the body of a change](operator-console-change-api.md#reg-1324))
- <a id="reg-1422"></a>**REG-1422** A change form shows the identifiers of the record as text that the lab operator cannot change. ([the identifiers of a record do not change](operator-interface.md#reg-0820))
- <a id="reg-1423"></a>**REG-1423** A choice of one value from a fixed set is a [select](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0308). For example, the accreditation level and the status are selects. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; [Judgement: text areas for lists, and selects and checkboxes for fixed sets](#reg-1487))
- <a id="reg-1424"></a>**REG-1424** A choice of more than one value from a fixed set is a [checkbox group](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0314). For example, the industries of a data holder brand are checkboxes. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; [Judgement: text areas for lists, and selects and checkboxes for fixed sets](#reg-1487))
- <a id="reg-1425"></a>**REG-1425** A list of values that the lab operator writes is a [text area with one value on each line](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0310). For example, the redirect URIs and the scopes are text areas. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; [Judgement: text areas for lists, and selects and checkboxes for fixed sets](#reg-1487))
- <a id="reg-1426"></a>**REG-1426** The endpoints of a data holder brand are a [group of related fields](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0317). ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="reg-1427"></a>**REG-1427** In a create, a select of a required value that has no usual value has [no selected value](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0309). For example, the accreditation level of a data recipient has no selected value. ([Judgement: no accreditation level selected first](#reg-1488))
- <a id="reg-1428"></a>**REG-1428** Each field that the console API schema does not require is an optional field, as the design system gives it. ([the label of an optional field](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0326))
- <a id="reg-1429"></a>**REG-1429** The help text of the identifier fields of a data holder brand says that a brand with no `dataHolderBrandId` needs only its public details. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [Judgement: only the public details for a brand of product data](participants.md#reg-0261))

## Values that the console sends

- <a id="reg-1430"></a>**REG-1430** The console removes the white space at the start and at the end of each value, and of each line of a list. ([Judgement: the console removes the white space at the ends of each value](#reg-1489))
- <a id="reg-1431"></a>**REG-1431** The console ignores an empty line of a list. For example, an empty last line of the redirect URIs gives no value. ([Judgement: the console removes the white space at the ends of each value](#reg-1489))
- <a id="reg-1432"></a>**REG-1432** The console sends an optional field that has no value as no field. For example, a change with no policy URI sends no `policyUri`. ([a field that is not in the body of a change](operator-console-change-api.md#reg-1324))
- <a id="reg-1433"></a>**REG-1433** A change form sends each field of the change, also the fields that the lab operator did not change. ([the body of a change](operator-console-change-api.md#reg-1323); [a field that is not in the body of a change](operator-console-change-api.md#reg-1324))
- <a id="reg-1434"></a>**REG-1434** A change form and an add form send the version of the participant that the form loaded. A create form sends no version. ([the version in the body of an action](operator-console-api.md#reg-1139); [Judgement: a version in the body, and 409 for a stale view](operator-console-api.md#reg-1180))
- <a id="reg-1435"></a>**REG-1435** The console does not check the values. The Register checks them with the rules of the participant model and the status rules, and the UI has no copy of the rules. ([the Register checks the full legal entity](operator-interface.md#reg-0823); [Judgement: hexagonal architecture (ADR 0013)][adr-0013])

## Refusals

- <a id="reg-1436"></a>**REG-1436** When the Register refuses the value of a field, the console shows the `detail` of the field error as the error message of that field. ([the field errors of the console API](operator-console-change-api.md#reg-1328); [WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [SC 3.3.3][wcag-3-3-3])
- <a id="reg-1437"></a>**REG-1437** The error of an item of a list shows at the field of the list. For example, a refused redirect URI shows at the redirect URIs. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="reg-1438"></a>**REG-1438** A refusal that names no field shows in the error summary of the form. For example, a stale version shows in the error summary. ([a refusal that names no field](operator-console-change-api.md#reg-1330); [WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="reg-1439"></a>**REG-1439** After a refusal, the form shows the error summary of the design system. ([the error summary of a form](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0327); [the focus on the error summary](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0331); [Judgement: a form shows an error summary, also for a refusal with no field](#reg-1490))
- <a id="reg-1440"></a>**REG-1440** The error summary gives the `detail` of the problem, and a link to each field that has an error. ([the links of the error summary](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0330); [Judgement: a form shows an error summary, also for a refusal with no field](#reg-1490))
- <a id="reg-1441"></a>**REG-1441** After a refusal, the form keeps each value that the lab operator gave, and the version that it loaded. ([Judgement: a form keeps its values and its version after a refusal](#reg-1491))
- <a id="reg-1442"></a>**REG-1442** When the version is stale, the lab operator opens the form again to see the current values. The form does not get the current values by itself. ([Judgement: a form keeps its values and its version after a refusal](#reg-1491))

## Submit, success and cancel

- <a id="reg-1443"></a>**REG-1443** While the request runs, the form is busy, as the design system gives it. ([the busy state of a form](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0323); [Judgement: no automatic second request](operator-console-actions.md#reg-1284))
- <a id="reg-1444"></a>**REG-1444** The console never sends a create, an add or a change again by itself. ([Judgement: no automatic second request](operator-console-actions.md#reg-1284))
- <a id="reg-1445"></a>**REG-1445** After a success, the console goes to the detail of the participant and shows the result in a toast. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="reg-1446"></a>**REG-1446** After a success, the focus moves to the heading of the detail, because the page of the form no longer exists. ([WCAG 2.2, SC 2.4.3][wcag-2-4-3])
- <a id="reg-1447"></a>**REG-1447** After a success, the card of the record that the form changed or added is open. For example, after an add of a software product, the card of the new software product is open. ([the cards of a detail](operator-console.md#reg-1413); [the cards that are open after a form](../../../web/packages/ui/design/components.md#ui-0242))
- <a id="reg-1448"></a>**REG-1448** Each form has a "Cancel" link back to the page that opened it. ([the Cancel link of a form](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0321); [Judgement: no confirmation of a cancel](#reg-1492))
- <a id="reg-1449"></a>**REG-1449** A cancel sends no request. The console does not ask the lab operator to confirm a cancel, and gives no warning about values that it did not send. ([Judgement: no confirmation of a cancel](#reg-1492))

## Judgements

### <a id="reg-1480"></a>REG-1480 The console creates and changes participants

- Kind: choice
- Basis: Step 3.3. The CDS gives no API to create participants, and the operator interface creates and changes each record ([the creates, the adds and the changes of the operator interface](operator-interface.md#reg-0809)).
- Reason: A lab user who tests their own data recipient software must register it first. An HTTP client with the operator token is a barrier to the most frequent set-up tasks. The console API and the operator interface stay separate adapters on the same use cases ([ADR 0014][adr-0014]).

### <a id="reg-1481"></a>REG-1481 A create of the legal entity only

- Kind: choice
- Basis: Step 3.3. The operator interface can create a legal entity with no brands ([the body of a create in the operator interface](operator-interface.md#reg-0809)).
- Reason: Each form then has the fields of one record, and the add forms do the work of the child records. Each field error then has a short pointer.

### <a id="reg-1482"></a>REG-1482 Forms on their own pages

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator. The console uses a dialog only for the confirmation of an action ([the confirmation of an action](operator-console-actions.md#reg-1211)).
- Reason: The add form of a software product has 14 fields, and the add form of a data holder brand has 15, which is too much for a dialog. A page has its own URL and history entry, and Back leaves it as the lab operator expects. A URL of a form that is never a URL of a record keeps the detail of each record available.

### <a id="reg-1483"></a>REG-1483 No confirmation of a form

- Kind: choice
- Basis: Step 3.2. [WCAG 2.2, SC 3.3.4][wcag-3-3-4] is met when the Register checks the values and the lab operator can correct them, or when the change is reversible.
- Reason: A change is reversible by a second change. A create and an add change no other record.

### <a id="reg-1484"></a>REG-1484 Links to the forms in each status

- Kind: choice
- Basis: Step 3.3. The domain permits a change and an add in each status, and refuses only a value or a status that breaks a rule ([the Register checks the full legal entity](operator-interface.md#reg-0823)).
- Reason: A list of the permitted changes in the console API would always give each change.

### <a id="reg-1485"></a>REG-1485 A form shows a new random UUID for each identifier

- Kind: choice
- Basis: Step 3.3. The Register accepts an identifier in each format ([an identifier in each format](participants.md#reg-0220)), and the lab operator gives the identifiers ([Judgement: the lab operator gives the identifiers](operator-interface.md#reg-0889)).
- Reason: A UUID is the usual value, so the lab operator types nothing. A lab service gets its identifiers in its configuration, so the lab operator must be able to give a known value. A brand with a `dataHolderBrandId` is the usual brand.

### <a id="reg-1486"></a>REG-1486 ACTIVE selected first

- Kind: choice
- Basis: Step 3.3. The operator interface permits each status in a create or an add ([a status in a create or an add](operator-interface.md#reg-0814); [Judgement: a create or an add can give each status](operator-interface.md#reg-0890)).
- Reason: ACTIVE is the usual status. With a fixed status of ACTIVE, an add under a parent that is not ACTIVE always fails, for example a software product of an INACTIVE brand. The field needs no rule in the UI, because the Register refuses a status that breaks a rule at the field of the status.

### <a id="reg-1487"></a>REG-1487 Text areas for lists, and selects and checkboxes for fixed sets

- Kind: choice
- Basis: Step 3.2. [WCAG 2.2, SC 1.3.1][wcag-1-3-1] requires that the structure of a form is available to assistive technology.
- Reason: A text area is one control with one label and one error, and the lab operator can paste a list into it. A list of text fields needs add and remove buttons and a focus rule for each, and gives no better access. The checkboxes and the select show each value of the fixed set, with the keyboard support of the browser.

### <a id="reg-1488"></a>REG-1488 No accreditation level selected first

- Kind: choice
- Basis: Step 3.3. The CDS gives no usual accreditation level ([the accreditation of a data recipient legal entity](participants.md#reg-0211)).
- Reason: A value that the console selects can go to the Register with no decision of the lab operator. The Register then refuses a create with no level at its field.

### <a id="reg-1489"></a>REG-1489 The console removes the white space at the ends of each value

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: A value that the lab operator pastes often has white space or an empty last line. A space at the end of a URI or an identifier is never intended. This is not a rule of the Register, which still checks each value.

### <a id="reg-1490"></a>REG-1490 A form shows an error summary, also for a refusal with no field

- Kind: choice
- Basis: Step 3.2. [WCAG 2.2, SC 3.3.1][wcag-3-3-1] needs only the error at the field. The [GOV.UK Design System: Error summary][govuk-error-summary] shows a summary with a link to each field, and moves the focus to it.
- Reason: A software product form is longer than one screen, so an error at a field far down the page is not seen. The summary also has the refusals that name no field, so the form has one place for errors.

### <a id="reg-1491"></a>REG-1491 A form keeps its values and its version after a refusal

- Kind: choice
- Basis: Step 3.3. The version check stops a change that is based on an old view ([Judgement: a version in the body, and 409 for a stale view](operator-console-api.md#reg-1180)).
- Reason: The lab operator does not lose their work. A form that takes a new version by itself would replace the change of a different person or tab with no decision of the lab operator.

### <a id="reg-1492"></a>REG-1492 No confirmation of a cancel

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: A lab form is quick to fill again. A warning before each navigation adds a rule for each route.

[adr-0013]: ../../../docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[rfc9562-v4]: https://www.rfc-editor.org/rfc/rfc9562#section-5.4
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-3-3-3]: https://www.w3.org/TR/WCAG22/#error-suggestion
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[govuk-error-summary]: https://design-system.service.gov.uk/components/error-summary/
