# Mock data holder: Operator interface

This area gives the lab-only operator interface, which creates customers, accounts and transactions, closes accounts and withdraws authorisations. A production data holder gets its customers, accounts and transactions from its banking systems. In the lab, the operator interface creates them, and a real component does not depend on it.

## Operations

| ID | Action | Method and path | Request body | Success | Source |
|---|---|---|---|---|---|
| <a id="dh-0720"></a>**DH-0720** | Create a customer | `POST /v1/customers` | `lastName`, `emailAddress`, and optionally `firstName` | 201 with the customer, which has its `customerNumber` | [Judgement: an operator interface for the customer data of the lab](#dh-0770) |
| <a id="dh-0721"></a>**DH-0721** | Read a customer | `GET /v1/customers/{customerNumber}` | None | 200 with the customer and its `accounts`, in the sequence of their creation | [Judgement: an operator interface for the customer data of the lab](#dh-0770) |
| <a id="dh-0722"></a>**DH-0722** | Create an account | `POST /v1/customers/{customerNumber}/accounts` | `displayName`, `productCategory`, `productName`, `bsb`, `accountNumber`, and optionally `creditLimit` | 201 with the account, which has its `accountId` | [Judgement: an operator interface for the customer data of the lab](#dh-0770) |
| <a id="dh-0723"></a>**DH-0723** | Close an account | `POST /v1/customers/{customerNumber}/accounts/{accountId}/close` | None | 200 with the account | [Judgement: an operator interface for the customer data of the lab](#dh-0770) |
| <a id="dh-0724"></a>**DH-0724** | Create a transaction | `POST /v1/customers/{customerNumber}/accounts/{accountId}/transactions` | `type`, `status`, `description`, `amount`, and optionally `reference`, `postingDateTime`, `executionDateTime`, `valueDateTime`, `merchantName`, `merchantCategoryCode`, `billerCode`, `billerName`, `crn`, `apcaNumber` and `extendedData` | 201 with the transaction, which has its `transactionId` | [Judgement: an operator interface for the customer data of the lab](#dh-0770) |
| <a id="dh-0725"></a>**DH-0725** | Read the transactions of an account | `GET /v1/customers/{customerNumber}/accounts/{accountId}/transactions` | None | 200 with the transactions, in the sequence of their creation | [Judgement: an operator interface for the customer data of the lab](#dh-0770) |
| <a id="dh-0726"></a>**DH-0726** | Read the arrangements of a customer | `GET /v1/customers/{customerNumber}/arrangements` | None | 200 with the arrangements, in the sequence of their creation | [CDR Rules][cdr-rules], rule 9.3; [Judgement: an operator interface for the customer data of the lab](#dh-0770) |
| <a id="dh-0727"></a>**DH-0727** | Withdraw an authorisation | `POST /v1/customers/{customerNumber}/arrangements/{arrangementId}/withdraw` | None | 200 with the arrangement | [CDR Rules][cdr-rules], rule 4.25(1)(b); [CDS: Withdrawal Standards][cds-withdrawal-standards] |

## Listener and access

The operator interface obeys the [operator interface library](../../../pkg/platform/operatorhttp/DESIGN.md): [each request has the operator token](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0001), [the operator token comes from the configuration](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0004), [a `Host` allowlist with no default](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0005) and [no cross-origin requests](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0007).

- <a id="dh-0728"></a>**DH-0728** The operator interface is on a listener that is separate from the CDR APIs. ([Judgement: an operator interface for the customer data of the lab](#dh-0770))

## Bodies and errors

- <a id="dh-0729"></a>**DH-0729** The request bodies and the response bodies are JSON. ([Judgement: an operator interface for the customer data of the lab](#dh-0770))
- <a id="dh-0730"></a>**DH-0730** When a request body is not valid, the response is 400, and the `detail` names each field and rule that fails. For example, a body with [a field that the table does not give](../../../pkg/platform/jsonhttp/DESIGN.md#body-0001) is not valid. ([RFC 9457, section 3.1.4][rfc9457-detail])
- <a id="dh-0731"></a>**DH-0731** The `detail` of an error response does not repeat a value of the request. ([RFC 9457, section 3.1.4][rfc9457-detail])
- <a id="dh-0732"></a>**DH-0732** A request for a customer that does not exist gets 404. ([RFC 9110, section 15.5.5][rfc9110-404])
- <a id="dh-0733"></a>**DH-0733** An error response is a [problem details](../../../docs/spec/glossary.md#problem-details) body with [the type `about:blank`](../../../pkg/platform/problem/DESIGN.md#prob-0001). ([RFC 9457][rfc9457])

## Customers and accounts

- <a id="dh-0734"></a>**DH-0734** A `lastName` has 1 to 100 characters. A `firstName` has 100 characters at most. ([Judgement: only the required CDS fields and the named optional fields](#dh-0771))
- <a id="dh-0735"></a>**DH-0735** An `emailAddress` is one address of RFC 5321. ([RFC 5321, section 4.1.2][rfc5321-mailbox])
- <a id="dh-0736"></a>**DH-0736** When a different customer of the brand has the `emailAddress`, with no difference for upper and lower case, the response is 409. The `detail` does not contain the address. ([RFC 9110, section 15.5.10][rfc9110-409]; [the uniqueness of the email address](consumer-authentication.md#dh-1202))
- <a id="dh-0737"></a>**DH-0737** A `displayName` and a `productName` have 1 to 100 characters. ([Judgement: only the required CDS fields and the named optional fields](#dh-0771))
- <a id="dh-0738"></a>**DH-0738** A `productCategory` is a value of `BankingProductCategoryV2`. ([CDS: Product Categories][cds-product-categories])
- <a id="dh-0739"></a>**DH-0739** A `bsb` has 6 digits, and an `accountNumber` has 6 to 10 digits. ([Judgement: only the required CDS fields and the named optional fields](#dh-0771))
- <a id="dh-0740"></a>**DH-0740** The pair of `bsb` and `accountNumber` is unique in the brand. A second account with the same pair gets 409. ([RFC 9110, section 15.5.10][rfc9110-409])
- <a id="dh-0741"></a>**DH-0741** A `creditLimit` is an AmountString with exactly 2 decimal digits, and it is zero or more. For example, `500.00`. ([CDS: Common Field Types][cds-common-field-types]; [Judgement: two decimal digits in an amount from the lab operator](#dh-0773))
- <a id="dh-0742"></a>**DH-0742** When the `creditLimit` is absent, it is `0.00`. ([Judgement: only the required CDS fields and the named optional fields](#dh-0771))
- <a id="dh-0743"></a>**DH-0743** The operator interface gives the account number in full, with no mask. ([CDS: Common Field Types][cds-common-field-types]; [Judgement: no mask of the account number for the lab operator](#dh-0774))
- <a id="dh-0744"></a>**DH-0744** The lab operator cannot remove a customer, an account or a transaction. ([Judgement: only the close of an account changes the customer data](#dh-0772))
- <a id="dh-0745"></a>**DH-0745** The lab operator cannot change a customer or a transaction. The only change to an account is its close. ([Judgement: only the close of an account changes the customer data](#dh-0772))
- <a id="dh-0746"></a>**DH-0746** A request for an account of a different customer gets 404. ([RFC 9110, section 15.5.5][rfc9110-404])
- <a id="dh-0747"></a>**DH-0747** A close of an account that is closed gets 409. ([RFC 9110, section 15.5.10][rfc9110-409])

## Transactions

- <a id="dh-0748"></a>**DH-0748** The transaction fields have the names, the types and the enumerations of `BankingTransactionV2` and `BankingTransactionDetailV3`. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail]; [Judgement: only the required CDS fields and the named optional fields](#dh-0771))
- <a id="dh-0749"></a>**DH-0749** An `amount` is an AmountString with exactly 2 decimal digits. For example, `-12.50`. ([CDS: Common Field Types][cds-common-field-types]; [Judgement: two decimal digits in an amount from the lab operator](#dh-0773))
- <a id="dh-0750"></a>**DH-0750** A `description` has 1 to 500 characters. ([Judgement: only the required CDS fields and the named optional fields](#dh-0771))
- <a id="dh-0751"></a>**DH-0751** When the `reference` is absent, it is an empty string. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail])
- <a id="dh-0752"></a>**DH-0752** A `POSTED` transaction has a `postingDateTime`. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail])
- <a id="dh-0753"></a>**DH-0753** A `PENDING` transaction has an `executionDateTime`. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail])
- <a id="dh-0754"></a>**DH-0754** The `extendedData` has these fields. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail])
  - `service` and `serviceVersion`
  - `payer`, for a positive amount
  - `payee`, for a negative amount
  - `extendedDescription`
- <a id="dh-0755"></a>**DH-0755** A request for a transaction of an account of a different customer gets 404. ([RFC 9110, section 15.5.5][rfc9110-404])
- <a id="dh-0756"></a>**DH-0756** A create of a transaction for a closed account gets 409. ([RFC 9110, section 15.5.10][rfc9110-409])
- <a id="dh-0757"></a>**DH-0757** The `transactionId` of the operator interface is not the `transactionId` of the banking APIs. ([the resource identifiers of the banking APIs](banking-apis.md#dh-1041))

## Arrangements

- <a id="dh-0758"></a>**DH-0758** An arrangement has these fields. ([CDS: CDR Arrangement ID][cds-arrangement-id]; [CDR Rules][cdr-rules], rule 9.3)
  - `arrangementId`: its `cdr_arrangement_id`
  - `softwareId`
  - `clientName`
  - `scopes`
  - `accountIds`: the account identifiers of the operator interface
  - `confirmedAt` and `sharingEndsAt`
- <a id="dh-0759"></a>**DH-0759** The `clientName` of an arrangement is the `client_name` of the registration at the confirmation. A later change of the registration does not change it. ([CDR Rules][cdr-rules], rule 9.3; [Judgement: the client name of an arrangement at the confirmation](#dh-0775))
- <a id="dh-0760"></a>**DH-0760** An arrangement that ended before `sharingEndsAt` also has `endedAt` and `endReason`. The `endReason` is one of these values. ([the end of an arrangement](end-of-an-arrangement.md#dh-0900))
  - `REVOKED_BY_DATA_RECIPIENT`
  - `WITHDRAWN_BY_CONSUMER`
  - `CONSUMER_NOT_ELIGIBLE`
  - `SOFTWARE_PRODUCT_REMOVED`
- <a id="dh-0761"></a>**DH-0761** An arrangement with the end reason `WITHDRAWN_BY_CONSUMER` also has `withdrawalMethod`: `CONSUMER_DASHBOARD` or `ALTERNATIVE_METHOD`. ([CDR Rules][cdr-rules], rule 4.25(1))
- <a id="dh-0762"></a>**DH-0762** An arrangement with a notice also has `noticeStatus`: `PENDING`, `COMPLETE` or `FAILED`. ([the notice to the data recipient](end-of-an-arrangement.md#dh-0924))
- <a id="dh-0763"></a>**DH-0763** A request for an arrangement of a different customer gets 404. ([RFC 9110, section 15.5.5][rfc9110-404])
- <a id="dh-0764"></a>**DH-0764** A withdrawal of an arrangement that ended gets 409. ([RFC 9110, section 15.5.10][rfc9110-409])
- <a id="dh-0765"></a>**DH-0765** The lab operator cannot create an arrangement. The only change to an arrangement is a withdrawal. ([Judgement: only the close of an account changes the customer data](#dh-0772))

## Judgements

### <a id="dh-0770"></a>DH-0770 An operator interface for the customer data of the lab

- Kind: choice
- Basis: Step 3.3. The CDS gives no interface for the customer data of a data holder. A production data holder gets the data from its banking systems.
- Reason: The lab needs invented customers, accounts and transactions for each test. A JSON API on its own listener lets a lab user and the conformance suite create them, with no access to the CDR APIs.

### <a id="dh-0771"></a>DH-0771 Only the required CDS fields and the named optional fields

- Kind: choice
- Basis: Step 3.3. The CDS marks many fields of the schemas as optional ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail]).
- Reason: The required fields of the CDS schemas, and the optional fields that this specification names, are sufficient for a data recipient to test each API in scope. Add a field when the lab needs it. The lab has no data to migrate, because a reset is simple.

### <a id="dh-0772"></a>DH-0772 Only the close of an account changes the customer data

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the change of customer data.
- Reason: A closed account lets a data recipient test `openStatus` and its filter. No other requirement needs a change, and a reset of the lab is simple.

### <a id="dh-0773"></a>DH-0773 Two decimal digits in an amount from the lab operator

- Kind: choice
- Basis: Step 3.3. An AmountString has 2 decimal digits or more, and more only when necessary ([CDS: Common Field Types][cds-common-field-types]).
- Reason: The data holder holds AUD only, which never needs more than 2 digits. Whole cents make the balance arithmetic exact, with no decimal library.

### <a id="dh-0774"></a>DH-0774 No mask of the account number for the lab operator

- Kind: choice
- Basis: Step 3.3. The mask is a rule for data that a consumer or a data recipient sees ([CDS: Common Field Types][cds-common-field-types]).
- Reason: The lab operator sent the number, so a mask hides nothing. The lab operator must be able to find an account by its number.

### <a id="dh-0775"></a>DH-0775 The client name of an arrangement at the confirmation

- Kind: choice
- Basis: Step 3.1. Rule 9.3 requires a record of each authorisation ([CDR Rules][cdr-rules]), and the authorisation page showed this name to the consumer.
- Reason: A record that changes with the registration is not a record of the authorisation. The name also stays after a `DELETE` of the registration.

[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-withdrawal-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#withdrawal-standards
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-schema-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactiondetailv3
[cds-arrangement-id]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-id
[rfc5321-mailbox]: https://www.rfc-editor.org/rfc/rfc5321#section-4.1.2
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.10
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[rfc9457-detail]: https://www.rfc-editor.org/rfc/rfc9457#section-3.1.4
