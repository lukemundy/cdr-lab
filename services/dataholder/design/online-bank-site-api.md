# Mock data holder: Online bank site API

This area gives the API of the online bank site. The site API is JSON over HTTP on the listener of the site, and only the site uses it. It obeys the [web UI library](../../../pkg/platform/consolehttp/DESIGN.md): [the `/api/` prefix](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0011) and [404 for a path that is not an operation](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0012). The [online bank site](online-bank-site.md#dh-1801) gives the pages, the [sign-in and the session](online-bank-site-sign-in.md) and the listener.

## Operations

| ID | Action | Method and path | Success | Source |
|---|---|---|---|---|
| <a id="dh-1700"></a>**DH-1700** | Read the brand | `GET /api/brand` | 200 with the brand name, the text of the footer, the layout and the photos of the brand | [Judgement: the brand gives a list of its photos](#dh-1792) |
| <a id="dh-1701"></a>**DH-1701** | List the current products | `GET /api/products` | 200 | [CDS: Get Products][cds-get-products] |
| <a id="dh-1702"></a>**DH-1702** | Read a current product | `GET /api/products/{productId}` | 200 | [CDS: Get Product Detail][cds-get-product-detail] |
| <a id="dh-1703"></a>**DH-1703** | Sign up | `POST /api/customers` | 201 with the new customer, and the session cookie | [Judgement: a sign-up starts a session](online-bank-site-sign-in.md#dh-2103) |
| <a id="dh-1704"></a>**DH-1704** | Sign in | `POST /api/session` | 204 with the session cookie | [the sign-in of a customer](online-bank-site-sign-in.md#dh-2023) |
| <a id="dh-1705"></a>**DH-1705** | Read the customer of the session | `GET /api/session` | 200 with the `customerNumber`, the `firstName` and the `lastName` | [the session on the server](online-bank-site-sign-in.md#dh-2033) |
| <a id="dh-1706"></a>**DH-1706** | Sign out | `DELETE /api/session` | 204, and the cookie expires, also with no valid session | [Judgement: 403 for a request with no valid session](#dh-1780) |
| <a id="dh-1707"></a>**DH-1707** | List the accounts | `GET /api/accounts` | 200 | [CDS: Get Accounts][cds-get-accounts] |
| <a id="dh-1708"></a>**DH-1708** | Read the balance of an account | `GET /api/accounts/{accountId}/balance` | 200 | [CDS: Get Account Balance][cds-get-account-balance] |
| <a id="dh-1709"></a>**DH-1709** | List the transactions of an account | `GET /api/accounts/{accountId}/transactions` | 200 | [CDS: Get Transactions For Account][cds-get-transactions] |
| <a id="dh-1710"></a>**DH-1710** | Open an account | `POST /api/accounts` | 201 with the new account | [Judgement: the new account in the response to an open](#dh-1797) |
| <a id="dh-1711"></a>**DH-1711** | Close an account | `POST /api/accounts/{accountId}/close` | 204 | [Judgement: the close uses the use case of the operator console](#dh-1796) |
| <a id="dh-1712"></a>**DH-1712** | List the authorisations | `GET /api/consents` | 200 | [CDR Rules][cdr-rules], rule 1.15 |
| <a id="dh-1713"></a>**DH-1713** | Withdraw an authorisation | `POST /api/consents/{arrangementId}/withdraw` | 204 | [Judgement: the withdrawal uses the use case of the operator](#dh-1798) |

## Requests and errors

- <a id="dh-1714"></a>**DH-1714** An OpenAPI 3.0 document in the data holder directory defines the site API. Tests in the data holder check each response against the document. ([ADR 0017][adr-0017]; [ADR 0019][adr-0019])
- <a id="dh-1715"></a>**DH-1715** Each object schema of the OpenAPI document forbids the properties that it does not name. ([ADR 0017][adr-0017])
- <a id="dh-1716"></a>**DH-1716** The paths of the site API have no version. For example, the path of the accounts is `/api/accounts`. ([Judgement: the UI and the API ship together (ADR 0014)][adr-0014])
- <a id="dh-1717"></a>**DH-1717** An error response is a [problem details](../../../docs/spec/glossary.md#problem-details) body with [the type `about:blank`](../../../pkg/platform/problem/DESIGN.md#prob-0001), except a refusal with [field errors](../../../pkg/platform/problem/DESIGN.md#prob-0002) and a [session that ended](../../../pkg/platform/problem/DESIGN.md#prob-0010). ([RFC 9457][rfc9457])
- <a id="dh-1718"></a>**DH-1718** The status codes of an error are the codes of the console API: 400 for a request that is not valid, 404 for a record that does not exist, and 409 for a change that the rules do not permit now. ([console API: the status codes of an error](operator-console-api.md#dh-1619))
- <a id="dh-1719"></a>**DH-1719** A request body that is not valid gives 400. For example, a body with [a field that the OpenAPI document does not give](../../../pkg/platform/jsonhttp/DESIGN.md#body-0001), or with [a second JSON value](../../../pkg/platform/jsonhttp/DESIGN.md#body-0002), is not valid. ([RFC 9110, section 15.5.1][rfc9110-400])
- <a id="dh-1720"></a>**DH-1720** A request body is 4 KiB or less for a sign-up, a sign-in and an open, and 16 KiB or less for a close. A larger body gives 413. ([the status of a body that is larger than its limit](../../../pkg/platform/problem/DESIGN.md#prob-0012); [Judgement: a body limit for each operation, with 413](operator-console-api.md#dh-1681))
- <a id="dh-1721"></a>**DH-1721** Telemetry: the site API puts no request body in telemetry. ([ADR 0006][adr-0006])
- <a id="dh-1722"></a>**DH-1722** No path of the site API contains a customer number or a different identifier of a consumer. For example, the accounts of the consumer are at `/api/accounts`. ([Judgement: no identifier of a consumer in a path](#dh-1791))
- <a id="dh-1723"></a>**DH-1723** The arrangement IDs, the names of the data recipients and the disclosure times go only to the customer of the session. This applies to the list of the authorisations and to the close action. ([CDR Rules][cdr-rules], rule 1.15)

## Brand and products

- <a id="dh-1724"></a>**DH-1724** The brand gives the [`layout` of the brand directory](brand.md#dh-0335), and the name of each photo that the brand directory has. For example, the list has `hero.webp` only when the directory has [that photo](brand.md#dh-0328). ([Judgement: the brand gives a list of its photos](#dh-1792))
- <a id="dh-1725"></a>**DH-1725** The fields of a product have the names and the values of the CDS product schemas. ([CDS: BankingProductV6][cds-schema-product]; [CDS: BankingProductDetailV7][cds-schema-product-detail]; [Judgement: the site API uses the CDS schemas](#dh-1789))
- <a id="dh-1726"></a>**DH-1726** The list of the products gives each current product of the brand in one response, with no pages. ([Judgement: all current products in one response](#dh-1793))
- <a id="dh-1727"></a>**DH-1727** The detail of a product gives one current product. A product that is not current, or that does not exist, gives 404, and the `detail` does not contain its identifier. ([RFC 9110, section 15.5.5][rfc9110-404]; [Judgement: only current products on the site](#dh-1786))
- <a id="dh-1728"></a>**DH-1728** The site API and Get Products use [the same rule for a current product](product-reference-data.md#dh-0219) and the same time of the request. ([CDS: Get Products][cds-get-products]; [Judgement: the products from the same source as Get Products](#dh-1785))
- <a id="dh-1729"></a>**DH-1729** For each preset brand, the list gives the products and the values of Get Products with `effective=CURRENT`, and each detail gives the values of Get Product Detail. ([CDS: Get Products][cds-get-products]; [CDS: Get Product Detail][cds-get-product-detail]; [Judgement: the products from the same source as Get Products](#dh-1785))
- <a id="dh-1730"></a>**DH-1730** The site API gives no future product, and no product whose `effectiveTo` is in the past. ([CDS: Get Products][cds-get-products]; [Judgement: only current products on the site](#dh-1786))
- <a id="dh-1731"></a>**DH-1731** The OpenAPI document gives the values of each CDS enumeration of a product. Each list has the same values that [the data holder accepts](product-reference-data.md#dh-0206). ([Judgement: the values of each CDS enumeration in the OpenAPI document](#dh-1787))
- <a id="dh-1732"></a>**DH-1732** The OpenAPI document gives the categories of revolving credit and the categories of a loan. Each list is the same as the rule of the data holder for an open. ([Judgement: the categories of an open as lists in the OpenAPI document](#dh-1788))
- <a id="dh-1733"></a>**DH-1733** The UI selects the amount field of the open page from the categories of revolving credit and of a loan. ([Judgement: the categories of an open as lists in the OpenAPI document](#dh-1788))

## Sign-up and sign-in

- <a id="dh-1734"></a>**DH-1734** A sign-up has the fields `lastName`, `emailAddress` and, optionally, `firstName`, with the rules of the [operator interface](operator-interface.md#dh-0734). ([Judgement: a sign-up starts a session](online-bank-site-sign-in.md#dh-2103))
- <a id="dh-1735"></a>**DH-1735** When the data holder refuses a value of a sign-up, the response is 400, with a [field error](../../../pkg/platform/problem/DESIGN.md#prob-0003) for each value. ([Judgement: each field error shows at its field](../../../pkg/platform/problem/DESIGN.md#prob-0009))
- <a id="dh-1736"></a>**DH-1736** The `detail` of each field error is a text in plain language for the consumer, which names the field with its label on the page. For example, the `detail` of a last name that is too long starts with "Enter a last name". ([Judgement: a plain-language detail for each field error](#dh-1781))
- <a id="dh-1737"></a>**DH-1737** A sign-up with an email address that a different customer of the brand has gives 409, with a field error at `/emailAddress`. ([operator interface: an email address in use](operator-interface.md#dh-0736); [Judgement: a plain-language detail for each field error](#dh-1781))
- <a id="dh-1738"></a>**DH-1738** A sign-in has the field `identifier` only. Its value is the email address or the customer number. ([the user identifiers of a consumer](consumer-authentication.md#dh-1200); [Judgement: the page does not send the password](online-bank-site-sign-in.md#dh-2104))
- <a id="dh-1739"></a>**DH-1739** A sign-in with a `password` field gives 400. ([Judgement: the page does not send the password](online-bank-site-sign-in.md#dh-2104))
- <a id="dh-1740"></a>**DH-1740** A sign-in with an identifier that no customer has gives 400, with a field error at `/identifier`. The `detail` is the same for an email address and a customer number. ([Judgement: a plain-language refusal that does not hide unknown user identifiers](online-bank-site-sign-in.md#dh-2105))
- <a id="dh-1741"></a>**DH-1741** A successful sign-up and a successful sign-in give the session cookie in the response. ([Judgement: a sign-up starts a session](online-bank-site-sign-in.md#dh-2103))
- <a id="dh-1742"></a>**DH-1742** A successful sign-up gives 201 with the `customerNumber`, the `firstName` and the `lastName` of the new customer, and with no `Location` header. ([Judgement: no identifier of a consumer in a path](#dh-1791))

## Session

- <a id="dh-1743"></a>**DH-1743** A request with no valid session gives 403. This applies to the session, the accounts, a balance, the transactions, an open, a close, the authorisations and a withdrawal. ([RFC 9110, section 15.5.4][rfc9110-403]; [Judgement: 403 for a request with no valid session](#dh-1780))
- <a id="dh-1744"></a>**DH-1744** When a request with no valid session has a session cookie, the problem has [the session-ended type](../../../pkg/platform/problem/DESIGN.md#prob-0010), and the response expires the cookie. For example, a request after a restart of the data holder gets it. ([Judgement: one type for a session that ended](../../../pkg/platform/problem/DESIGN.md#prob-0011))

## Accounts

- <a id="dh-1745"></a>**DH-1745** The fields of an account are `displayName`, `maskedNumber`, `productCategory`, `productName`, `openStatus` and `creationDate`, with the names and the values of `BankingAccountV3`, and `accountId`. ([CDS: BankingAccountV3][cds-schema-account]; [Judgement: the site API uses the CDS schemas](#dh-1789))
- <a id="dh-1746"></a>**DH-1746** The `accountId` of an account and the `transactionId` of a transaction are the identifiers of the data holder, not the identifiers of an arrangement. ([Judgement: the same account functions as the APIs of an arrangement](#dh-1790))
- <a id="dh-1747"></a>**DH-1747** The fields of a balance are `currentBalance`, `availableBalance` and `creditLimit`, with the names and the values of `BankingBalance`. ([CDS: BankingBalance][cds-schema-balance]; [Judgement: the site API uses the CDS schemas](#dh-1789))
- <a id="dh-1748"></a>**DH-1748** The fields of a transaction have the names and the values of `BankingTransactionV2`. ([CDS: BankingTransactionV2][cds-schema-transaction]; [Judgement: the site API uses the CDS schemas](#dh-1789))
- <a id="dh-1749"></a>**DH-1749** The site API gives no `bsb`, no `accountNumber` and no `extendedData`. ([CDS: Common Field Types][cds-common-field-types])
- <a id="dh-1750"></a>**DH-1750** Each account in the list of the accounts has a `balance` with the fields of a balance. ([CDS: Get Bulk Balances][cds-get-bulk-balances]; [Judgement: the balance in the list of the accounts](#dh-1794))
- <a id="dh-1751"></a>**DH-1751** The data holder gives the accounts, the balances and the transactions of the site with the same functions as the APIs of an arrangement, for the customer of the session. ([Judgement: the same account functions as the APIs of an arrangement](#dh-1790))
- <a id="dh-1752"></a>**DH-1752** The accounts are in the order of [Get Accounts](banking-apis.md#dh-1060). ([Judgement: the same account functions as the APIs of an arrangement](#dh-1790))
- <a id="dh-1753"></a>**DH-1753** An account of a different customer, or an account that does not exist, gives 404, and the `detail` does not contain its identifier. This applies to each operation on an account. ([RFC 9110, section 15.5.5][rfc9110-404])
- <a id="dh-1754"></a>**DH-1754** The transactions of an account take only the query parameters `page` and `page-size`, with [the default page](../../../pkg/cds/DESIGN.md#cdsl-0017), [the default page size](../../../pkg/cds/DESIGN.md#cdsl-0018) and [the maximum page size](../../../pkg/cds/DESIGN.md#cdsl-0019) of CDS pagination. ([CDS: Pagination][cds-pagination]; [Judgement: the same transactions as Get Transactions For Account with no filter](#dh-1795))
- <a id="dh-1755"></a>**DH-1755** The response of the transactions gives the number of records and the number of pages. ([CDS: Pagination][cds-pagination])

## Open and close

- <a id="dh-1756"></a>**DH-1756** An open has the field `productId`. The product ID is in the body, not in the path. ([the open of an account of a product](online-bank-site-accounts.md#dh-2210))
- <a id="dh-1757"></a>**DH-1757** An open has the field `creditLimit` only for revolving credit, and the field `loanAmount` only for a loan. ([Judgement: a credit limit for revolving credit and a drawdown for a loan](online-bank-site-accounts.md#dh-2302))
- <a id="dh-1758"></a>**DH-1758** The `creditLimit` and the `loanAmount` of an open are AmountStrings, with [the rule of the operator interface for a `creditLimit`](operator-interface.md#dh-0741), and they are more than `0.00`. ([Judgement: an amount as a person types it on the open page](online-bank-site-accounts.md#dh-2303))
- <a id="dh-1759"></a>**DH-1759** An open of a product that is not current, or that does not exist, gives 400 with a field error at `/productId`. ([the product of an open](online-bank-site-accounts.md#dh-2224))
- <a id="dh-1760"></a>**DH-1760** An amount that is absent, not valid, or for a product of a different kind gives 400 with a field error at its field. For example, a `loanAmount` for a credit card gives a field error at `/loanAmount`. ([Judgement: each field error shows at its field](../../../pkg/platform/problem/DESIGN.md#prob-0009))
- <a id="dh-1761"></a>**DH-1761** A successful open gives 201 with the new account, with the fields of an account in the list of the accounts, and with no `Location` header. ([Judgement: the new account in the response to an open](#dh-1797))
- <a id="dh-1762"></a>**DH-1762** Each open account in the list of the accounts has the action `close`, with each arrangement that the close ends: its `arrangementId` and its `clientName`, in the sequence of their confirmation. ([the end of eligibility](end-of-an-arrangement.md#dh-0918); [Judgement: the close uses the use case of the operator console](#dh-1796))
- <a id="dh-1763"></a>**DH-1763** The list of the arrangements of the `close` action is empty when the account is not the last open account. A closed account has no action. ([the end of eligibility](end-of-an-arrangement.md#dh-0920))
- <a id="dh-1764"></a>**DH-1764** A close has the field `ends`: the `arrangementId` of each arrangement that the confirmation named, or an empty list when it named none. ([console API: a close sends the arrangements of its confirmation](operator-console-api.md#dh-1650))
- <a id="dh-1765"></a>**DH-1765** A close of a closed account, and a close that now ends different arrangements, give 409, and the data holder changes nothing. ([console API: a close that ends different arrangements](operator-console-api.md#dh-1651); [Judgement: the close uses the use case of the operator console](#dh-1796))

## Authorisations

- <a id="dh-1766"></a>**DH-1766** The list of the authorisations gives each arrangement of the customer of the session in one response, with no pages. Each arrangement has these fields. ([CDR Rules][cdr-rules], rule 1.15(3); [the consumer dashboard](consumer-dashboard.md#dh-1903))
  - `arrangementId`, `clientName` and `brandName`
  - `legalEntityName`, when the arrangement has one
  - `status`: `CURRENT` or `ENDED`, as in [the console API](operator-console-api.md#dh-1640)
  - `confirmedAt`, `sharingDurationSeconds` and `sharingEndsAt`
  - `endedAt`, for an ended arrangement: the time of the event that ended it, or the end of its sharing
  - `accounts`, with the fields of an account in the list of the accounts, and no `balance` and no `actions`
  - `dataClusters`
- <a id="dh-1767"></a>**DH-1767** Each data cluster has its `name` and its `permissions`. It also has `firstDisclosedAt` and `lastDisclosedAt` when it has a disclosure. ([CDR Rules][cdr-rules], rule 1.15(3)(f))
- <a id="dh-1768"></a>**DH-1768** The data holder gives the language of the data clusters, so that the UI has no copy of the rules. ([CDS: Data Language Standards][cds-data-language])
- <a id="dh-1769"></a>**DH-1769** A withdrawal sends no body, because it ends only the arrangement of its path. A successful withdrawal gives 204. ([Judgement: the withdrawal uses the use case of the operator](#dh-1798))
- <a id="dh-1770"></a>**DH-1770** A withdrawal of an arrangement that ended gives 409, and the data holder changes nothing. ([operator interface: a withdrawal of an ended arrangement](operator-interface.md#dh-0764))
- <a id="dh-1771"></a>**DH-1771** An arrangement of a different customer, or an arrangement that does not exist, gives 404, and the `detail` does not contain its identifier. ([RFC 9110, section 15.5.5][rfc9110-404])

## Judgements

### <a id="dh-1780"></a>DH-1780 403 for a request with no valid session

- Kind: choice
- Basis: Step 3.2. A 401 response must have a `WWW-Authenticate` challenge ([RFC 9110, section 15.5.2][rfc9110-401]), and a cookie session has no HTTP authentication scheme. 403 is the response when the server refuses the request ([RFC 9110, section 15.5.4][rfc9110-403]).
- Reason: The UI goes to the sign-in page after a 403. A sign-out needs no session, because the consumer is signed out after it in each case.

### <a id="dh-1781"></a>DH-1781 A plain-language detail for each field error

- Kind: choice
- Basis: Step 3.2. The `detail` of the operator interface names the fields in camelCase, which a consumer does not read ([plain language in a consumer UI](../../../web/packages/ui/design/content.md#ui-0142)).
- Reason: The site API gives the text, so the UI has no copy of the rules, as in the operator console.

### <a id="dh-1785"></a>DH-1785 The products from the same source as Get Products

- Kind: choice
- Basis: Step 3.3. The CDS gives the products of a brand in Get Products and Get Product Detail ([CDS: Get Products][cds-get-products]).
- Reason: A consumer and a data recipient then see the same products, by construction. A call from the site to the CDS API adds a cross-origin call, or a call of the service to itself over the network.

### <a id="dh-1786"></a>DH-1786 Only current products on the site

- Kind: choice
- Basis: Step 3.3. Get Products gives the current products by default ([CDS: Get Products][cds-get-products]).
- Reason: A bank web site lists the products that the bank sells now.

### <a id="dh-1787"></a>DH-1787 The values of each CDS enumeration in the OpenAPI document

- Kind: choice
- Basis: Step 3.3. The CDS gives the values of each enumeration, and the data holder accepts each of them ([CDS: BankingProductDetailV7][cds-schema-product-detail]).
- Reason: The site shows a label in plain language for each value, because a consumer does not read a CDS code. With the list in the OpenAPI document and the check against the data holder, a new value cannot show with no label.

### <a id="dh-1788"></a>DH-1788 The categories of an open as lists in the OpenAPI document

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for an open of an account.
- Reason: The product responses then stay the same as Get Products and Get Product Detail, with no conversion. The UI gets the lists from one source, as for the CDS enumerations.

### <a id="dh-1789"></a>DH-1789 The site API uses the CDS schemas

- Kind: choice
- Basis: Step 3.3. The CDS gives the schemas of a product, an account, a balance and a transaction ([CDS: BankingProductV6][cds-schema-product]; [CDS: BankingAccountV3][cds-schema-account]).
- Reason: A test can then compare each field with the CDS APIs, with no conversion. The site shows what a data recipient gets.

### <a id="dh-1790"></a>DH-1790 The same account functions as the APIs of an arrangement

- Kind: choice
- Basis: Step 3.3. The CDS gives the account data in the banking APIs ([CDS: Get Transactions For Account][cds-get-transactions]).
- Reason: One source of the data gives the same values as the APIs of an arrangement, by construction. The `accountId` and the `transactionId` of an arrangement are for each software product, and the site has no software product.

### <a id="dh-1791"></a>DH-1791 No identifier of a consumer in a path

- Kind: choice
- Basis: Step 3.3. A path is in telemetry [as its route pattern](../../../pkg/platform/telemetry/DESIGN.md#tele-0005).
- Reason: The session gives the consumer of each request, so a path needs no identifier of the consumer. A sign-up then gives no `Location` header, because the path of the customer is `/api/session`.

### <a id="dh-1792"></a>DH-1792 The brand gives a list of its photos

- Kind: choice
- Basis: Step 3.3. The photos of a brand directory are optional ([brand: the hero photo](brand.md#dh-0328)).
- Reason: The site asks only for a photo in the list, so it never shows a broken image.

### <a id="dh-1793"></a>DH-1793 All current products in one response

- Kind: choice
- Basis: Step 3.3. A catalogue file is 1 MiB or less ([brand: the size of a JSON file](brand.md#dh-0350)).
- Reason: The site shows all current products on one page. One response is the least complex.

### <a id="dh-1794"></a>DH-1794 The balance in the list of the accounts

- Kind: choice
- Basis: Step 3.3. Get Bulk Balances gives the balances of many accounts in one response ([CDS: Get Bulk Balances][cds-get-bulk-balances]).
- Reason: The accounts page then needs one request. The balances come from one query of the totals.

### <a id="dh-1795"></a>DH-1795 The same transactions as Get Transactions For Account with no filter

- Kind: choice
- Basis: Step 3.3. With no filter, Get Transactions For Account gives the transactions of 90 days to the time of the request ([CDS: Get Transactions For Account][cds-get-transactions]).
- Reason: The site gives the same list as Get Transactions For Account with no filter. A test then compares the two, with no conversion.

### <a id="dh-1796"></a>DH-1796 The close uses the use case of the operator console

- Kind: choice
- Basis: Step 3.3. The CDS and the CDR Rules give no rule for the close of an account.
- Reason: One use case gives the same rules and the same end of eligibility, by construction. An authorisation that starts between the confirmation and the close gives 409, so it does not end with no warning to the consumer. The consumer sees the name of each app, not a general text.

### <a id="dh-1797"></a>DH-1797 The new account in the response to an open

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the online bank site.
- Reason: The site goes to the page of the account, and it needs the `accountId` from the response. The site API has no `GET` of one account, so a `Location` header has no target.

### <a id="dh-1798"></a>DH-1798 The withdrawal uses the use case of the operator

- Kind: choice
- Basis: Step 3.3. A withdrawal on the dashboard and a withdrawal of the operator have the same effects ([the methods of withdrawal](end-of-an-arrangement.md#dh-0916)).
- Reason: One use case gives the same effects, by construction, with [the method `CONSUMER_DASHBOARD`](end-of-an-arrangement.md#dh-0917). A `cdr_arrangement_id` does not identify the consumer ([CDS: CDR Arrangement ID][cds-arrangement-id]), so the path keeps [no identifier of a consumer](#dh-1722).

[adr-0006]: ../../../docs/adr/0006-telemetry.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0017]: ../../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0019]: ../../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-arrangement-id]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-id
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-data-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-language-standards-common
[cds-get-account-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-balance
[cds-get-accounts]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-accounts
[cds-get-bulk-balances]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-bulk-balances
[cds-get-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-product-detail
[cds-get-products]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-products
[cds-get-transactions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account
[cds-otp]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#one-time-password-credential-requirements
[cds-pagination]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination
[cds-schema-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingaccountv3
[cds-schema-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingbalance
[cds-schema-product]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductv6
[cds-schema-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductdetailv7
[cds-schema-transaction]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactionv2
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-400]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1
[rfc9110-401]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.2
[rfc9110-403]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.4
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
