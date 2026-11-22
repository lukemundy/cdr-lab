# Mock data holder: Operator console actions and create forms

This area gives the actions of the lab-only operator console and its create forms. The [operator console](operator-console.md) gives the views, the presentation and the listener.

## Actions

The [web platform package](../../../web/packages/web-platform/DESIGN.md) gives the refusal, the refresh and the result of an action: [the `detail` of a refusal in the dialog or the form of the action](../../../web/packages/web-platform/DESIGN.md#wcon-0014), [the current state after each action](../../../web/packages/web-platform/DESIGN.md#wcon-0015) and [the result of a success in a toast](../../../web/packages/web-platform/DESIGN.md#wcon-0016). The design system gives that [the UI never sends an action again by itself](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0301).

- <a id="dh-1443"></a>**DH-1443** The console does these actions of the operator interface, with the same rules. ([Judgement: an operator interface for the customer data of the lab](operator-interface.md#dh-0770))
  - [create a customer](operator-interface.md#dh-0720)
  - [create an account for a customer](operator-interface.md#dh-0722)
  - [create a transaction for an open account](operator-interface.md#dh-0724)
  - [close an account](operator-interface.md#dh-0723)
  - [withdraw an authorisation for a customer](operator-interface.md#dh-0727)
- <a id="dh-1444"></a>**DH-1444** Before a close and a withdrawal, the console asks the lab operator to confirm. A close and a withdrawal cannot be reversed. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4])
- <a id="dh-1445"></a>**DH-1445** The confirmation of a close of the last open account of a customer names each arrangement that ends. It says that the data holder sends a notice for each one. ([the end of eligibility](end-of-an-arrangement.md#dh-0918); [Judgement: the confirmation names each arrangement that a close ends](#dh-1516))
- <a id="dh-1446"></a>**DH-1446** The confirmation of a withdrawal names the arrangement and its software product. It says that the withdrawal cannot be reversed, and that the data holder sends a notice to the data recipient. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]; [the notice to the data recipient](end-of-an-arrangement.md#dh-0924))
- <a id="dh-1447"></a>**DH-1447** The console offers only the actions that the rules permit for the current state of the record. The [console API](operator-console-api.md#dh-1643) gives these actions, and the UI has no copy of the rules. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])

## Create forms

- <a id="dh-1452"></a>**DH-1452** Each create is a form on its own page, with its own URL. A URL of a form is never the URL of a record. ([Judgement: the same action and form rules in each console](#dh-1517))
- <a id="dh-1453"></a>**DH-1453** Each create form obeys these rules. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [SC 3.3.2][wcag-3-3-2]; [SC 3.3.3][wcag-3-3-3]; [SC 2.4.3][wcag-2-4-3]; [GOV.UK Design System: Error summary][govuk-error-summary]; [GOV.UK Design System: Question pages][govuk-question-pages]; [Judgement: the same action and form rules in each console](#dh-1517))
  - Each field that the console API schema does not require is an [optional field](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0326).
  - The console [removes the white space at the ends of each value](../../../web/packages/web-platform/DESIGN.md#wcon-0008). It sends an optional field that has no value as no field.
  - The console does not check the values. The data holder checks each value.
  - When the data holder refuses the value of a field, the console shows the `detail` of the field error as the error message of that field.
  - After a refusal, the form shows the [error summary](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0327). The summary gives the `detail` of the problem. A refusal that names no field shows only in the summary.
  - After a refusal, the form keeps each value that the lab operator gave.
  - While the request runs, the form is [busy](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0323).
  - After a success, the console goes to the page that the create gives, and the focus moves to the heading of that page.
  - Each form has a ["Cancel" link](../../../web/packages/ui/design/forms-and-dialogs.md#ui-0321) back to the page that opened it. A cancel sends no request, and the console does not ask the lab operator to confirm it.
- <a id="dh-1454"></a>**DH-1454** The list of the customers has the action "Create customer". ([create a customer](operator-interface.md#dh-0720))
- <a id="dh-1455"></a>**DH-1455** After a create of a customer, the console goes to the detail of the new customer. The toast gives the new customer number. ([Judgement: the customer number in the toast](#dh-1518))
- <a id="dh-1456"></a>**DH-1456** The detail of a customer has the action "Create account", also when the customer has no open account. ([Judgement: a new account for each customer](#dh-1519))
- <a id="dh-1457"></a>**DH-1457** After a create of an account, the console goes to the detail of the customer, which shows the new account. The toast gives the display name of the new account. ([Judgement: a new account for each customer](#dh-1519))
- <a id="dh-1458"></a>**DH-1458** The form of an account gives the values of `productCategory` as a list. The values are the values of `BankingProductCategoryV2` in the target CDS version, and the OpenAPI document of the console API gives them. ([CDS: Product Categories][cds-product-categories]; [a product category of the operator interface](operator-interface.md#dh-0738))
- <a id="dh-1459"></a>**DH-1459** When the console API permits a create of a transaction for an account, the transactions of the account have the action "Create transaction". The action is a link to the form, not a confirmation. ([Judgement: the create of a transaction from its list](#dh-1520))
- <a id="dh-1460"></a>**DH-1460** After a create of a transaction, the console goes to the first page of the transactions of the account. The toast gives the description of the new transaction. ([Judgement: the create of a transaction from its list](#dh-1520))
- <a id="dh-1461"></a>**DH-1461** The form of a transaction has a select for each of these fields, with no value selected first. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail]; [Judgement: the same action and form rules in each console](#dh-1517))
  - `type`
  - `status`
  - `extendedData.extensionUType`
  - `extendedData.nppPayload.service`
- <a id="dh-1462"></a>**DH-1462** The fields of `extendedData` are in one group, with a legend. The group is always open. ([Judgement: the extended data in a group that is always open](#dh-1521))
- <a id="dh-1463"></a>**DH-1463** The console sends `extendedData` only when a field of the group has a value. ([Judgement: the extended data in a group that is always open](#dh-1521))
- <a id="dh-1464"></a>**DH-1464** The lab operator types each time of a transaction in a native date-time control, in the time zone of the browser, with seconds. The console sends each time in UTC. ([Judgement: UTC for all times (ADR 0012)][adr-0012]; [Judgement: local time with seconds in the form, sent in UTC](#dh-1522))

## Judgements

### <a id="dh-1516"></a>DH-1516 The confirmation names each arrangement that a close ends

- Kind: choice
- Basis: Step 3.3. The close of the last open account ends each arrangement of the customer ([the end of eligibility](end-of-an-arrangement.md#dh-0918)).
- Reason: The close changes records that the lab operator did not select. The lab operator must see them before the change.

### <a id="dh-1517"></a>DH-1517 The same action and form rules in each console

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator. The Register console has the same rules for its actions, its forms and its URLs.
- Reason: The lab operator learns one behaviour for both consoles. A form on its own page has space for the many fields of a record, and a link or a reload shows the same view. A form that keeps its values after a refusal does not lose the work of the lab operator.

### <a id="dh-1518"></a>DH-1518 The customer number in the toast

- Kind: choice
- Basis: Step 3.3. The data holder creates the customer number ([the customer number](consumer-authentication.md#dh-1201)).
- Reason: The lab operator did not give the customer number and needs it to find the customer again.

### <a id="dh-1519"></a>DH-1519 A new account for each customer

- Kind: choice
- Basis: Step 3.3. A closed account cannot open again ([only the close of an account changes the customer data](operator-interface.md#dh-0772)).
- Reason: Each customer can get a new account. The data holder creates the account ID, and the lab operator does not need it to find the account.

### <a id="dh-1520"></a>DH-1520 The create of a transaction from its list

- Kind: choice
- Basis: Step 3.3. Only an open account gets a transaction ([create a transaction](operator-interface.md#dh-0724)).
- Reason: A create changes no other record, so it needs no confirmation. The console API has no detail of a transaction, and the list shows each field.

### <a id="dh-1521"></a>DH-1521 The extended data in a group that is always open

- Kind: choice
- Basis: Step 3.3. The design system has no group of fields that opens and closes ([the form fields of the design system](../../../web/packages/ui/design/forms-and-dialogs.md)).
- Reason: A collapsed group hides its field errors and the targets of the links of the error summary. The values of the group are the only sign that the lab operator wants it, so an empty group sends nothing.

### <a id="dh-1522"></a>DH-1522 Local time with seconds in the form, sent in UTC

- Kind: choice
- Basis: Step 3.3. The API gets each time in UTC, and the UI shows each time in the time zone of the browser ([ADR 0012][adr-0012]).
- Reason: The lab operator thinks of a time of a test in local time, and the control of the browser gives a valid local time or no value. The seconds make each time of the CDS available.

[cds-schema-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactiondetailv3
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[adr-0012]: ../../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-3-3-3]: https://www.w3.org/TR/WCAG22/#error-suggestion
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[govuk-error-summary]: https://design-system.service.gov.uk/components/error-summary/
[govuk-question-pages]: https://design-system.service.gov.uk/patterns/question-pages/
