# Mock data holder: Operator console API

This area gives the console API of the operator console. The console API is JSON over HTTP on the listener of the console, and only the console uses it. It obeys the [web UI library](../../../pkg/platform/consolehttp/DESIGN.md): [the `/api/` prefix](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0011), [404 for a path that is not an operation](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0012), [the page parameters](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0014), [the page sizes](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0015), [400 for a page that is not valid](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0016) and [an empty parameter is absent](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0017). The [operator console](../DESIGN.md#operator-console) gives the UI.

## Operations

| ID | Action | Method and path | Success | Source |
|---|---|---|---|---|
| <a id="dh-1600"></a>**DH-1600** | Read the brand | `GET /api/brand` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1601"></a>**DH-1601** | Read the overview | `GET /api/overview` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1602"></a>**DH-1602** | List the customers | `GET /api/customers` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1603"></a>**DH-1603** | Create a customer | `POST /api/customers` | 201 | [operator interface: create a customer](operator-interface.md#dh-0720) |
| <a id="dh-1604"></a>**DH-1604** | Read a customer | `GET /api/customers/{customerNumber}` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1605"></a>**DH-1605** | Create an account | `POST /api/customers/{customerNumber}/accounts` | 201 | [operator interface: create an account](operator-interface.md#dh-0722) |
| <a id="dh-1606"></a>**DH-1606** | Close an account | `POST /api/customers/{customerNumber}/accounts/{accountId}/close` | 204 | [operator interface: close an account](operator-interface.md#dh-0723) |
| <a id="dh-1607"></a>**DH-1607** | List the transactions of an account | `GET /api/customers/{customerNumber}/accounts/{accountId}/transactions` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1608"></a>**DH-1608** | Create a transaction | `POST /api/customers/{customerNumber}/accounts/{accountId}/transactions` | 201 | [operator interface: create a transaction](operator-interface.md#dh-0724) |
| <a id="dh-1609"></a>**DH-1609** | Withdraw an authorisation | `POST /api/customers/{customerNumber}/arrangements/{arrangementId}/withdraw` | 204 | [operator interface: withdraw an authorisation](operator-interface.md#dh-0727) |
| <a id="dh-1610"></a>**DH-1610** | List the registrations | `GET /api/registrations` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1611"></a>**DH-1611** | Read a registration | `GET /api/registrations/{clientId}` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1612"></a>**DH-1612** | List the arrangements of a registration | `GET /api/registrations/{clientId}/arrangements` | 200 | [Judgement: a web UI for the lab operator (ADR 0014)][adr-0014] |
| <a id="dh-1613"></a>**DH-1613** | Read the participant statuses of the software products | `GET /api/software-product-statuses` | 200 | [CDS: Metadata Cache Management][cds-metadata-cache] |

## Requests and errors

- <a id="dh-1614"></a>**DH-1614** An OpenAPI 3.0 document in the data holder directory defines the console API. Tests in the data holder check each response against the document. ([ADR 0017][adr-0017]; [ADR 0019][adr-0019])
- <a id="dh-1615"></a>**DH-1615** Each object schema of the OpenAPI document forbids the properties that it does not name. ([ADR 0017][adr-0017])
- <a id="dh-1616"></a>**DH-1616** The paths of the console API have no version. For example, the path of the overview is `/api/overview`. ([Judgement: the UI and the API ship together (ADR 0014)][adr-0014])
- <a id="dh-1617"></a>**DH-1617** The console API ignores a query parameter that it does not know. For example, `GET /api/registrations?colour=red` gives the list. ([Judgement: an unknown query parameter is ignored](#dh-1680))
- <a id="dh-1618"></a>**DH-1618** An error response is a [problem details](../../../docs/spec/glossary.md#problem-details) body with [the type `about:blank`](../../../pkg/platform/problem/DESIGN.md#prob-0001), except a refusal with field errors. ([RFC 9457][rfc9457])
- <a id="dh-1619"></a>**DH-1619** The status codes of an error are the codes of the operator interface: 400 for a request that is not valid, 404 for a record that does not exist, and 409 for a change that the rules do not permit now. ([operator interface: a body that is not valid](operator-interface.md#dh-0730); [RFC 9110, section 15.5.10][rfc9110-409])
- <a id="dh-1620"></a>**DH-1620** A request body that is not valid gives 400. For example, a body with [a field that the operation does not accept](../../../pkg/platform/jsonhttp/DESIGN.md#body-0001), or with [a second JSON value](../../../pkg/platform/jsonhttp/DESIGN.md#body-0002), is not valid. ([RFC 9110, section 15.5.1][rfc9110-400])
- <a id="dh-1621"></a>**DH-1621** A request body is 64 KiB or less for a create, and 16 KiB or less for a close. A larger body gives 400. ([Judgement: a body limit for each operation](#dh-1681))
- <a id="dh-1622"></a>**DH-1622** A customer, an account, an arrangement or a registration that does not exist gives 404. The `detail` does not contain its identifier. ([RFC 9110, section 15.5.5][rfc9110-404]; [operator interface: no value of the request in a detail](operator-interface.md#dh-0731))
- <a id="dh-1623"></a>**DH-1623** The fields of a customer, an account, a transaction and an arrangement have the names of the operator interface. For example, an account has `accountId`. ([operator interface: the fields of an arrangement](operator-interface.md#dh-0758); [Judgement: the fields of a customer, an account and a transaction](operator-interface.md#dh-0771))
- <a id="dh-1624"></a>**DH-1624** Telemetry: the console API puts no request body in telemetry. The telemetry library gives [the route pattern of a path](../../../pkg/platform/telemetry/DESIGN.md#tele-0005) and [the redacted query string](../../../pkg/platform/telemetry/DESIGN.md#tele-0007). ([ADR 0006][adr-0006])

## Lists

- <a id="dh-1625"></a>**DH-1625** The list of the customers, the transactions of an account and the arrangements of a registration are in pages. ([Judgement: pages only for the lists that grow](#dh-1682))
- <a id="dh-1626"></a>**DH-1626** The accounts and the arrangements of a customer, the registrations and the participant statuses are each one response, with no pages. ([Judgement: pages only for the lists that grow](#dh-1682))
- <a id="dh-1627"></a>**DH-1627** The transactions of an account are newest first. ([Judgement: pages only for the lists that grow](#dh-1682))
- <a id="dh-1628"></a>**DH-1628** A search of the list of the customers finds a part of the customer number, the first name, the last name or the email address, with no difference for upper and lower case. For example, `jane` finds `Jane@example.test`. ([the search of the customers](../DESIGN.md#decisions))
- <a id="dh-1629"></a>**DH-1629** The list of the customers is sorted by one of these values, and then by the customer number. The default is the last name. ([Judgement: the sorts of the console lists](#dh-1683))
  - the last name
  - the customer number
  - the time of the last change
  - the number of open accounts
  - the number of arrangements that did not end
- <a id="dh-1630"></a>**DH-1630** An arrangement did not end when it has no end, and the end of its sharing is after the time of the request. ([the count of the arrangements of a customer](../DESIGN.md#decisions))
- <a id="dh-1631"></a>**DH-1631** The registrations are sorted by one of these values, and then by the client ID. The default is the `client_name`. ([Judgement: the sorts of the console lists](#dh-1683); [Judgement: the client name as the default order of the registrations](#dh-1684))
  - the `client_name`
  - the time of the registration
  - the number of arrangements of the client ID that did not end
  - the number of all the arrangements of the client ID
- <a id="dh-1632"></a>**DH-1632** The participant statuses are filtered by the registration: all (the default), registered or not registered. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-1633"></a>**DH-1633** The participant statuses are sorted by the name (the default) or by the status, and then by the software product ID. ([Judgement: the sorts of the console lists](#dh-1683))
- <a id="dh-1634"></a>**DH-1634** In a sort of the participant statuses by the name, each software product with no registration comes after the registered software products, in both directions. ([Judgement: a software product with no registration after the names](#dh-1685))
- <a id="dh-1635"></a>**DH-1635** The arrangements of a registration are each arrangement of its client ID, also an arrangement that ended. ([the arrangements of a registration](../DESIGN.md#decisions))
- <a id="dh-1636"></a>**DH-1636** Each arrangement of a registration has its `arrangementId`, the customer number, the `firstName` and the `lastName` of the customer, the status, the time of the confirmation and the end of the sharing. ([the customer in the arrangements of a registration](../DESIGN.md#decisions))
- <a id="dh-1637"></a>**DH-1637** The arrangements of a registration are newest first by the time of the confirmation, and then by the `arrangementId`. ([the arrangements of a registration](../DESIGN.md#decisions))
- <a id="dh-1638"></a>**DH-1638** The detail of a registration does not give its arrangements. ([the arrangements of a registration](../DESIGN.md#decisions))

## Customers, arrangements and registrations

- <a id="dh-1639"></a>**DH-1639** The detail of a customer gives, for each arrangement, its client ID, and if a registration has that client ID now. ([the link from an arrangement to a registration](../DESIGN.md#decisions))
- <a id="dh-1640"></a>**DH-1640** The status of an arrangement is `ENDED` after an event ended it, and at and after the end of its sharing. Otherwise it is `CURRENT`. ([the end of an arrangement](end-of-an-arrangement.md#dh-0900); [Judgement: the status of an arrangement in the console API](#dh-1686))
- <a id="dh-1641"></a>**DH-1641** An arrangement has `endedAt` and `endReason` only when an event ended it. ([operator interface: the end of an arrangement](operator-interface.md#dh-0760); [Judgement: the status of an arrangement in the console API](#dh-1686))
- <a id="dh-1642"></a>**DH-1642** A registration has these fields. The list and the detail give the same object. ([CDS: Dynamic Client Registration][cds-dcr-endpoints]; [Judgement: the registrations in the console API](#dh-1687))
  - `registeredAt`: the time of the last registration of the software product
  - `lastUpdated`: the time of the last `PUT`, or `registeredAt` before a `PUT`
  - `softwareProductStatus`, when the data holder has a status
  - `metadata`: each value of the registration, with the names of the registration metadata, except `software_statement`

## Actions

- <a id="dh-1643"></a>**DH-1643** The detail of a customer gives, for each account and each arrangement, the actions that the rules permit now. The data holder finds them with the same rules as the change itself. ([Judgement: the console API gives the permitted actions](#dh-1688))
- <a id="dh-1644"></a>**DH-1644** The action that closes the last open account of a customer gives each arrangement that the close ends. ([the end of eligibility](end-of-an-arrangement.md#dh-0918); [Judgement: the console API gives the permitted actions](#dh-1688))
- <a id="dh-1645"></a>**DH-1645** An open account also has the action `create-transaction`, which ends no arrangement. A closed account does not have it. ([operator interface: no transaction for a closed account](operator-interface.md#dh-0756); [Judgement: the console API gives the permitted actions](#dh-1688))
- <a id="dh-1646"></a>**DH-1646** The withdraw action of an arrangement gives that arrangement as the one arrangement that ends. ([Judgement: the console API gives the permitted actions](#dh-1688))
- <a id="dh-1647"></a>**DH-1647** An action and a create send no version of the record. ([Judgement: the state of the record in place of a version](#dh-1689))
- <a id="dh-1648"></a>**DH-1648** When the state of the record does not permit an action now, the response is 409, and the data holder changes nothing. For example, a close of a closed account gives 409. ([operator interface: a close of a closed account](operator-interface.md#dh-0747); [Judgement: the state of the record in place of a version](#dh-1689))
- <a id="dh-1649"></a>**DH-1649** A withdrawal sends no body. ([Judgement: the state of the record in place of a version](#dh-1689))
- <a id="dh-1650"></a>**DH-1650** A close sends the `arrangementId` of each arrangement that its confirmation named. It sends an empty list when the confirmation named none. ([Judgement: the state of the record in place of a version](#dh-1689))
- <a id="dh-1651"></a>**DH-1651** When a close now ends different arrangements, the response is 409, and the data holder changes nothing. The `detail` says that the customer changed, and that the operator must look at the new state. ([Judgement: the state of the record in place of a version](#dh-1689))
- <a id="dh-1652"></a>**DH-1652** The data holder finds the arrangements that a close ends with the same rules as the end itself, in the same transaction as the close. ([the end of eligibility](end-of-an-arrangement.md#dh-0918))
- <a id="dh-1653"></a>**DH-1653** A create gives 201 with no body, and a `Location` header with the path of the new record in the console API. For example, a new customer has the path `/api/customers/{customerNumber}`. ([RFC 9110, section 15.3.2][rfc9110-201])
- <a id="dh-1654"></a>**DH-1654** The path of a new transaction is `/api/customers/{customerNumber}/accounts/{accountId}/transactions/{transactionId}`. The console API has no `GET` for it. ([RFC 9110, section 15.3.2][rfc9110-201]; [Judgement: the location of a new transaction](#dh-1690))
- <a id="dh-1655"></a>**DH-1655** A close and a withdrawal give 204 with no body. ([RFC 9110, section 15.3.5][rfc9110-204])
- <a id="dh-1656"></a>**DH-1656** When the data holder refuses a value of a create, the response is 400, with a [field error](../../../pkg/platform/problem/DESIGN.md#prob-0003) for each value that the rules of the [operator interface](operator-interface.md#dh-0734) refuse. ([Judgement: each field error shows at its field](../../../pkg/platform/problem/DESIGN.md#prob-0009))
- <a id="dh-1657"></a>**DH-1657** A value with no valid format also gives a field error. For example, a `creditLimit` that is not an AmountString, or a time that is not an RFC 3339 date-time, gives a field error. ([operator interface: the format of an amount](operator-interface.md#dh-0741); [CDS: Common Field Types][cds-common-field-types])
- <a id="dh-1658"></a>**DH-1658** The `detail` of each field error is the `detail` that the operator interface gives for that rule. ([operator interface: a body that is not valid](operator-interface.md#dh-0730))
- <a id="dh-1659"></a>**DH-1659** The field error of a rule of two fields is at the field that is absent. For example, a `POSTED` transaction with no `postingDateTime` has its field error at `postingDateTime`. ([Judgement: the field error of a rule of two fields](#dh-1691))
- <a id="dh-1660"></a>**DH-1660** A create of an account with the `bsb` and the `accountNumber` of a different account gives 409, with one field error at `accountNumber`. The data holder changes nothing. ([operator interface: a unique BSB and account number](operator-interface.md#dh-0740); [Judgement: the field error of a BSB and an account number in use](#dh-1692))

## Overview

- <a id="dh-1661"></a>**DH-1661** The overview gives the counts and the 10 newest changes in one response, from one read of the database. ([Judgement: one read for the overview](#dh-1693))
- <a id="dh-1662"></a>**DH-1662** The changes of the overview are newest first. ([the changes in the overview](../DESIGN.md#decisions))
- <a id="dh-1663"></a>**DH-1663** Each change of the overview has `time` (UTC) and `kind`. When the kind has them, it also has these fields. ([the changes in the overview](../DESIGN.md#decisions); [ADR 0012][adr-0012])
  - `customerNumber`
  - `clientId`, `clientName` and `clientRegistered` (if a registration has the client ID now)
  - `productName`
  - `sharingEndsAt`
  - `endReason` and `withdrawalMethod`
- <a id="dh-1664"></a>**DH-1664** The overview has no name or email address of a customer, no display name of an account and no account number. ([the content of the overview](../DESIGN.md#decisions))

## Judgements

### <a id="dh-1680"></a>DH-1680 An unknown query parameter is ignored

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the console API, and RFC 9110 gives no rule for an unknown query parameter.
- Reason: A query parameter changes no record, so a refusal protects nothing. To read only the known parameters is the least complex. The Register console API has the same rule.

### <a id="dh-1681"></a>DH-1681 A body limit for each operation

- Kind: choice
- Basis: Step 3.2. A server can refuse a request body that is larger than it permits ([RFC 9110, section 15.5.14][rfc9110-413]).
- Reason: A customer, an account or a transaction is a few kilobytes at most, and a close has a list of arrangement IDs. A small limit stops a large body before the data holder decodes it. 400 is the status of each other body that is not valid.

### <a id="dh-1682"></a>DH-1682 Pages only for the lists that grow

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the console API.
- Reason: The lab can have many customers, an account can have many transactions, and a software product can have many arrangements. A customer has few accounts and arrangements, and the lab has few registrations, so one response is the least complex for them.

### <a id="dh-1683"></a>DH-1683 The sorts of the console lists

- Kind: choice
- Basis: Step 3.3. The Register console gives the rules: text in the byte order, a name with no difference for case, and the identifier of the record last ([Register DESIGN.md: Decisions][register-decisions]).
- Reason: The same rules in both consoles. The sorts by a count let the operator find the records with the most use. The data holder calculates a count of the arrangements that did not end at the request, because it depends on the time of the request.

### <a id="dh-1684"></a>DH-1684 The client name as the default order of the registrations

- Kind: choice
- Basis: Step 3.3. The CDS gives no order of the registrations.
- Reason: The operator finds a software product by its name. The client ID makes the order stable.

### <a id="dh-1685"></a>DH-1685 A software product with no registration after the names

- Kind: choice
- Basis: Step 3.3. The CDS gives no order of the participant statuses.
- Reason: The registered software products are the ones that the operator looks for. A name that does not exist has no place in the order of the names.

### <a id="dh-1686"></a>DH-1686 The status of an arrangement in the console API

- Kind: choice
- Basis: Step 3.3. An arrangement ends at the end of its sharing also with no event ([the end of an arrangement](end-of-an-arrangement.md#dh-0900)).
- Reason: The console shows the status of each record, and the UI has no copy of the rules. The operator interface gives no reason for the end of the sharing, so the console API does not invent one.

### <a id="dh-1687"></a>DH-1687 The registrations in the console API

- Kind: choice
- Basis: Step 3.3. The CDS gives the names of the registration metadata ([CDS: Dynamic Client Registration][cds-dcr-endpoints]).
- Reason: The console values are not metadata, so they are not in the object with the CDS names. The SSA is a signed token, and the console shows no tokens. One object for the list and the detail is the least complex, because the lab has few registrations.

### <a id="dh-1688"></a>DH-1688 The console API gives the permitted actions

- Kind: choice
- Basis: Step 3.3. The rules of an action are in the domain ([ADR 0013][adr-0013]).
- Reason: A copy of the rules in the UI can offer an action, or give a confirmation, that is not true. The state of the account permits a create of a transaction, so that action is with the other actions of the account.

### <a id="dh-1689"></a>DH-1689 The state of the record in place of a version

- Kind: choice
- Basis: Step 3.3. The customers, the accounts and the transactions have no history and no version, and a customer and a transaction cannot change ([operator interface: no change of a customer or a transaction](operator-interface.md#dh-0745)).
- Reason: A close and a withdrawal change only an open record, so the state refuses a second one. A close of a different account can make a close end arrangements that its confirmation did not name, and the arrangements in the request protect against this with no version of the customer. A withdrawal ends only the arrangement of its path, so no other change makes its effect different from its confirmation.

### <a id="dh-1690"></a>DH-1690 The location of a new transaction

- Kind: choice
- Basis: Step 3.2. The `Location` of a 201 identifies the new resource ([RFC 9110, section 15.3.2][rfc9110-201]).
- Reason: The path of the list does not identify one transaction. The list gives each field, so a `GET` of one transaction adds a second view of the same data, and the UI does not read it.

### <a id="dh-1691"></a>DH-1691 The field error of a rule of two fields

- Kind: choice
- Basis: Step 3.2. A UI shows an input error at the field that has the error ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]).
- Reason: The status and the sign of the amount are usually correct. The operator corrects the rule with a value in the field that is absent, so the error shows at that field. The fields are the `postingDateTime` of a `POSTED` transaction, the `executionDateTime` of a `PENDING` transaction, and `extendedData.payee` or `extendedData.payer` for the sign of the amount.

### <a id="dh-1692"></a>DH-1692 The field error of a BSB and an account number in use

- Kind: choice
- Basis: Step 3.3. The operator interface gives this 409 with no field error ([operator interface: a unique BSB and account number](operator-interface.md#dh-0740)).
- Reason: The accounts of a brand usually have the same few BSBs, so the operator corrects the pair with a different account number. The Register console API also gives an identifier that a different record has as 409 with a field error.

### <a id="dh-1693"></a>DH-1693 One read for the overview

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the console API.
- Reason: Two of the counts from the lists need each page of the customers. The counts and the changes of one read agree with each other.

[adr-0006]: ../../../docs/adr/0006-telemetry.md
[adr-0012]: ../../../docs/adr/0012-utc-for-all-times.md
[adr-0013]: ../../../docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0017]: ../../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0019]: ../../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[register-decisions]: ../../register/DESIGN.md#decisions
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-dcr-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dynamic-client-registration-endpoints
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-204]: https://www.rfc-editor.org/rfc/rfc9110#section-15.3.5
[rfc9110-400]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.10
[rfc9110-413]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.14
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
