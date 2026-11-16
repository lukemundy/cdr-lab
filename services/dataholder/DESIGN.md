# Mock data holder: design

This document gives the specification and the requirements for the mock data holder. It does not describe the implementation.

The [overview of the specification](design/README.md) gives the summary, the context, the areas, the goals and the non-goals. Each area file in `design/` gives the rules of its area.

## Operator console

The operator console is a web UI for the lab operator, who acts as the staff of the bank. It shows the customers, their accounts, transactions and arrangements, the notices to data recipients, the client registrations and the Register statuses of software products. It does the actions of the operator interface. It is a lab-only feature: real components must not depend on it. The architecture is in [ADR 0014][adr-0014], the design system in the [design system DESIGN.md][ui-design], and the contract of its API in [ADR 0017][adr-0017].

The console is a staff tool. It is not a consumer dashboard, and it does not satisfy rule 1.15: the dashboard is "an online service" for the consumer ([CDR Rules][cdr-rules], rule 1.15(1)). The consumer dashboard is on the online bank site (see Consumer dashboard).

### Functions

The console shows:

- An overview, which is the first page of the console. It gives the number of customers, the number of open accounts and of closed accounts, the number of current arrangements and of ended arrangements, the number of client registrations and the number of registered software products that do not have the Register status `ACTIVE`. It also gives the 10 newest changes (see the next list). (See Decisions.)
- A list of the customers, with the customer number, the name, the email address, the number of open accounts and the number of arrangements that did not end. The operator can search the list by a part of the customer number, the name or the email address. (See Decisions.)
- The detail of a customer: the customer (customer number, first name, last name, email address and the time of the last change), each account, and each arrangement. Judgement: a consumer can ask the staff for copies of their records "in any form", and the staff then find the records of one customer ([CDR Rules][cdr-rules], rule 9.5; [CDR Support: Rule 9.5 – Requests from CDR consumers for copies of records][support-copies-of-records]).
- For each account: the display name, the product category, the product name, the BSB and the account number in full, the credit limit, the current balance, the available balance, the status and the date of creation. (See Decisions.)
- The transactions of an account, newest first, with the fields that the operator interface gives.
- For each arrangement: the `cdr_arrangement_id`, the software product (`softwareId` and `clientName`), the scopes, the accounts ("None" when the consent has no accounts, because a consent with no `bank:` scope has no account selection), the time of the confirmation, the end of the sharing, and when it ended, the time and the reason of the end, and for a withdrawal, the method. ([CDR Rules][cdr-rules], rule 9.3(1)(a) to (c))
- For each arrangement with a notice: the status of the notice, and each attempt with its time and its result. ([CDR Rules][cdr-rules], rule 4.26A and rule 9.3(1); see Notice to the data recipient)
- A list of the client registrations, with the client ID, the software ID, the `client_name`, the time of the registration, the time of the last change, the Register status of the software product, the number of arrangements of the client that did not end and the number of all the arrangements of the client. The detail of a registration gives its metadata, with each redirect URI, the `recipient_base_uri`, the `jwks_uri` and the scopes. (See Decisions.)
- The arrangements of the client ID of a registration, also the arrangements that ended, in a table with pages, newest first: the `cdr_arrangement_id`, the name and the customer number of the customer, the status, the time of the confirmation and the end of the sharing. The name links to the customer. (See Decisions.)
- The Register statuses: each software product that the data holder has a status for, with its name, its status, and the time at which the data holder last got the statuses from the Register. The name is the `client_name` of the registration of the software product. A software product with no registration shows "Not registered". The operator can filter the list by the registration: all, registered or not registered. ([CDS: Metadata Cache Management][cds-metadata-cache]; see Register statuses and Decisions)

The changes of the overview are the creation of a customer, the creation and the close of an account, the confirmation and the end of an arrangement, and the registration, the change and the deletion of a client registration. Each change has:

- its time and its kind;
- the customer number, for a change of a customer, an account or an arrangement;
- the `client_name` of the registration, for a change of an arrangement or a registration;
- a detail when the kind has one: the product name of an account, the end of the sharing of a confirmed arrangement, and the reason of the end of an arrangement, with the method for a withdrawal.

The overview does not show the name or the email address of a customer, the display name of an account or an account number. (See Decisions.) The customer number links to the customer, and the `client_name` links to the registration when the registration exists. Each count links to its list when the console has one.

Each row of a list links to the detail of its record. An arrangement links to the registration of its client, when the registration exists.

The console does not show the one-time passwords, the signing keys, the tokens or the authorisation codes. Judgement: a one-time password in a staff tool lets the staff authenticate as the consumer. The mail sink and the telemetry backend have their own UIs.

The console does the actions of the operator interface, with the same rules ([operator interface](design/operator-interface.md)):

- Create a customer. Create an account for a customer. Create a transaction for an open account.
- Close an account.
- Withdraw a consent for a customer.

Rules for the actions:

- Before a close and a withdrawal, the console asks the operator to confirm. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]: neither can be reversed)
- The confirmation of a close of the last open account of a customer names each arrangement that ends, and says that the data holder sends a notice for each one. (See End of eligibility. Judgement: the close changes records that the operator did not select.)
- The confirmation of a withdrawal names the arrangement and its software product, says that it cannot be reversed, and says that the data holder sends a notice to the data recipient.
- The console offers only the actions that the rules permit now for the record. The console API gives these actions. The UI has no copy of the rules.
- A create is a form on its own page, with its own URL. The forms follow the rules of a create in the Register console ([Register DESIGN.md: Operator console][register-console]): the label of an optional field, white space, no checks of values in the UI, the field errors, the error summary, the busy submit button, the toast and the focus after a success, and "Cancel".
- The list of the customers has the action "Create customer". After a create of a customer, the console goes to the detail of the new customer, and the toast gives the new customer number. (Judgement: the data holder creates the customer number, and the operator needs it to find the customer again.)
- The detail of a customer has the action "Create account", also when the customer has no open account. After a create of an account, the console goes to the detail of the customer, which shows the new account, and the toast gives the display name of the new account. (Judgement: each customer can get a new account, and a closed account cannot open again. The data holder creates the account ID, and the operator does not need it to find the account.)
- The form of an account gives the values of `productCategory` as a list. The values are the values of `BankingProductCategoryV2` in the target CDS version, and the OpenAPI document of the console API gives them. ([CDS: Product Categories][cds-product-categories])
- The transactions of an account have the action "Create transaction" when the console API permits it. After a create of a transaction, the console goes to the first page of the transactions of the account, and the toast gives the description of the new transaction. (Judgement: the console API has no detail of a transaction, and the list shows each field.)
- The form of a transaction has a select for each field with a CDS enumeration (`type`, `status`, `extendedData.extensionUType` and `extendedData.nppPayload.service`), with no value selected first. The fields of `extendedData` are in one group, and the console sends `extendedData` only when a field of the group has a value. (See Decisions.)
- The operator types each time of a transaction in the time zone of the browser, with seconds. The console sends each time in UTC. ([ADR 0012][adr-0012])
- When the data holder refuses an action, the console shows the `detail` of the problem in the dialog or in the form of the action. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- After an action, also after a refused action, each view shows the current state without a reload of the page.
- The console never sends an action again by itself.
- The console shows the result of a successful action in a toast. The view also shows the new state. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])

### Presentation

- The console is for desktop browsers. It is usable at a viewport width of 1280 CSS px and more. ([ADR 0014][adr-0014])
- The header of the console shows the logo of the brand on the brand colour, with the brand name as the text alternative of the logo, and the label "Operator console". ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]; see Decisions)
- The console has the shell of a console of the design system, with the roles of the brand that the design system gives to the console of a participant: the shell colours, the link and the primary action. The rest of the console uses the neutral lab brand, its font and its radius, and the compact density. ([design system DESIGN.md: Brands][ui-brands], [Layout][ui-layout]; see Decisions)
- The side navigation of each preset brand is not white. With the logo hidden, an operator can tell the console of each preset brand from the console of each other preset brand and from the Register console by the colours of the shell. Judgement: an operator who has more than one console open finds the tab of a bank in one look.
- "Overview" is the first item of the side navigation. The path `/` is the overview.
- The title of the browser tab gives the brand name, and the icon of the tab is [the mark of the brand](design/brand.md#dh-0326).
- The search, the sort and the page of a list are in the URL. A search replaces the history entry of the list, and a sort and a page add a history entry. The detail of a customer, the transactions of an account and the detail of a registration each have their own URL. The page of the arrangements of a registration is in the URL of the detail of the registration. Judgement: the same as the Register console.
- The console shows each time in the time zone of the browser, with its offset from UTC (see the format of a time in the [design system DESIGN.md][ui-design]). The API gives each time in UTC. ([ADR 0012][adr-0012])
- The console shows each URI of a registration as text with a copy button, not as a link, and loads no file from it. ([OWASP ASVS 5.0, V3.6.1][asvs-v3])
- An amount shows with its sign and 2 decimal digits. A pending transaction shows as pending.
- A BSB shows as two groups of three digits with a hyphen, for example `062-000`. Its copy button copies the 6 digits. (See Decisions.)
- The console has the accessibility level of an operator console in the design system: the automated checks pass, with no other target.

### Listener and security

The console is a web UI. The [web UI library specification](../../pkg/platform/consolehttp/DESIGN.md) gives the rules of its listener, its browser security, its cache headers and its routes. These rules are the same as in the Register console ([Register DESIGN.md: Listener and security][register-console-listener]): its own listener, the setting `CONSOLE_ENABLED`, the setting `CONSOLE_ALLOWED_HOSTS` and its values, plain HTTP, and no login and no credential in the browser.

- The console listener serves the logo and the mark of the brand, and a stylesheet that sets only the roles of the brand that the console uses, with [the `Host` allowlist](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0004) and [the response headers](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007) of the console. Judgement: the Content Security Policy of the console permits images and styles from the same origin only, and the brand files of the public listener are on a different origin. (See Decisions.)
- The data holder starts the console listener only in the process that runs the APIs. A process that runs only the workers needs no console setting ([a process needs only the settings that it uses](design/non-functional-requirements.md#dh-0010)).

### Console API

The console API has [the `/api/` prefix](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0011), [404 for a path that is not an operation](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0012), [the page parameters](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0014), [the page sizes](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0015) and [400 for a page that is not valid](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0016) of the web UI library. It has the rules of the Register console API ([Register DESIGN.md: Console API][register-console-api]) for: the OpenAPI document and its tests, the problem details and their status codes, a query parameter that is not known or empty, no version in the path, a body that is not valid, the field errors, and no request body in telemetry. The field errors have [the field-errors type](../../pkg/platform/problem/DESIGN.md#prob-0002) and [the format](../../pkg/platform/problem/DESIGN.md#prob-0003) of the problem details library. These rules are different:

- The OpenAPI document is in the data holder directory.
- The list of customers is searched by a part of the customer number, the first name, the last name or the email address, with no difference for upper and lower case. It is sorted by the last name, the customer number, the time of the last change, the number of open accounts or the number of arrangements that did not end, and then by the customer number. (See Decisions.)
- The list of customers, the transactions of an account and the arrangements of a registration are in pages. The transactions are newest first. The accounts and the arrangements of a customer, the client registrations and the Register statuses are each one response. Judgement: a customer has few accounts and arrangements, and the lab has few registrations, but the lab can have many customers, an account can have many transactions, and a software product can have many arrangements.
- The client registrations are sorted by the `client_name` (the default), the time of the registration, the number of arrangements that did not end or the number of all the arrangements, and then by the client ID. (See Decisions.)
- The Register statuses are filtered by the registration: all (the default), registered or not registered. They are sorted by the name (the default) or the status, and then by the software ID. In a sort by the name, each software product with no registration comes after the registered software products, in both directions. (See Decisions.)
- The arrangements of a registration are each arrangement of its client ID, also an arrangement that ended, with its `arrangementId`, the customer number, the `firstName` and the `lastName` of the customer, the status, the time of the confirmation and the end of the sharing. They are newest first by the time of the confirmation, and then by the `arrangementId`. The detail of a registration does not give its arrangements. An unknown registration gives 404, as the detail does. (See Decisions.)
- The detail of a customer gives, for each arrangement, the client ID, and if a registration has that client ID now. (See Decisions.)
- The detail of a customer gives, for each account and each arrangement, the actions that the rules permit now. The action that closes the last open account gives each arrangement that ends. The data holder finds these with the same rules as the change itself. An open account also has the action `create-transaction`, which ends no arrangement. (See Decisions.)
- An action and a create send no version. When the record changed and the action is not permitted now, for example a close of a closed account, the response is 409, and the data holder changes nothing. (See Decisions.)
- The withdraw action of an arrangement gives the arrangement as the one arrangement that ends. A withdrawal sends no body, because it ends only the arrangement of its path. (See Decisions.)
- A close sends the `arrangementId` of each arrangement that its confirmation named, and an empty list when it named none. When the close now ends different arrangements, the response is 409, its `detail` says that the customer changed and that the operator must look at the new state, and the data holder changes nothing. The data holder finds the arrangements that the close ends with the same rules as the end itself, in the same transaction as the close. (See Decisions.)
- A create of a customer, an account or a transaction gives 201 with no body, and a `Location` header with the path of the new record in the console API. The path of a new transaction is `/api/customers/{customerNumber}/accounts/{accountId}/transactions/{transactionId}`, and the console API has no `GET` for it. (See Decisions.) A close and a withdrawal give 204 with no body.
- When the data holder refuses a value of a create, the response is 400, as in the operator interface, and it has a field error for each value that the rules of the [operator interface](design/operator-interface.md) refuse. This includes a value with no valid format, for example a `creditLimit` or an `amount` that is not an AmountString, or a time that is not an RFC 3339 date-time. The `detail` of each field error is the text that the `detail` of the operator interface gives for that rule. A rule of two fields is at the field that is absent. (See Decisions.)
- A create of an account with the `bsb` and the `accountNumber` of a different account gives 409, as in the operator interface, with one field error at `accountNumber`. The `detail` of the field error is the `detail` of the operator interface. The data holder changes nothing. The operator interface gives this 409 with no field error. (See Decisions.)
- An unknown customer, account, arrangement or registration gives 404, and the `detail` does not contain its identifier.
- The fields have the names of the operator interface. The client registrations have the names of the registration metadata ([CDS: Dynamic Client Registration][cds-dcr-endpoints]).
- The overview gives the counts and the 10 newest changes in one response, from one read of the database, so that the counts and the changes agree with each other. The changes are newest first. Each change has `time` (UTC), `kind`, and, when the kind has them, `customerNumber`, `clientId`, `clientName`, `clientRegistered` (if a registration has the client ID now), `productName`, `sharingEndsAt`, `endReason` and `withdrawalMethod`. The response has no name or email address of a customer, no display name of an account and no account number. (See Decisions.)
- In telemetry, a path of the console API is [its route pattern](../../pkg/platform/telemetry/DESIGN.md#tele-0005), and [its query string is `REDACTED`](../../pkg/platform/telemetry/DESIGN.md#tele-0007).

| Action | Method and path | Success |
|---|---|---|
| Read the brand | `GET /api/brand` | 200 |
| Read the overview | `GET /api/overview` | 200 |
| List the customers | `GET /api/customers` | 200 |
| Create a customer | `POST /api/customers` | 201 |
| Read a customer | `GET /api/customers/{customerNumber}` | 200 |
| Create an account | `POST /api/customers/{customerNumber}/accounts` | 201 |
| Close an account | `POST /api/customers/{customerNumber}/accounts/{accountId}/close` | 204 |
| List the transactions of an account | `GET /api/customers/{customerNumber}/accounts/{accountId}/transactions` | 200 |
| Create a transaction | `POST /api/customers/{customerNumber}/accounts/{accountId}/transactions` | 201 |
| Withdraw a consent | `POST /api/customers/{customerNumber}/arrangements/{arrangementId}/withdraw` | 204 |
| List the client registrations | `GET /api/registrations` | 200 |
| Read a client registration | `GET /api/registrations/{clientId}` | 200 |
| List the arrangements of a client registration | `GET /api/registrations/{clientId}/arrangements` | 200 |
| Read the Register statuses | `GET /api/software-product-statuses` | 200 |

### Non-functional requirements

- Telemetry: the console API has the same server telemetry as the other listeners, and the rules of the data holder telemetry apply to it. A search of the customers has a customer number, a name or an email address in its query string. Thus no log contains the query string, and [the spans do not contain it](../../pkg/platform/telemetry/DESIGN.md#tele-0007). The UI sends no telemetry. ([ADR 0014][adr-0014], [ADR 0006][adr-0006])
- Performance: on a development workstation, a search of 1,000 customers shows within 1 second, and the first page of 10,000 transactions of an account shows within 1 second. Judgement: the same limits as the Register console.
- Tests: tests in the data holder check the listener, the console API, and that the telemetry of a search and of each path contains no customer number, name or email address. A small number of browser tests check the main operator journeys against the lab. A browser test creates records through the operator interface and checks that each count of the overview changes by the number of records that it creates, so that it passes in a new lab and in a lab with other data. The conformance suite does not test the console. ([ADR 0019][adr-0019])

## Online bank site

The online bank site is the web site of the bank for its consumers, in the brand of the instance. Now it has a landing page, the list and the detail of the current products of the brand, sign-up, sign-in and sign-out, the accounts of the consumer with their balances and transactions, the opening and the close of an account by the consumer, and the consumer dashboard. It is a lab-only feature: real components must not depend on it, and it is not the web site of a production bank. The architecture is in [ADR 0014][adr-0014], the design system in the [design system DESIGN.md][ui-design], and the contract of its API in [ADR 0017][adr-0017].

The site is not part of the CDR authentication flow. The authentication and consent pages stay on the public listener, and the site does not change them (see [consumer authentication](design/consumer-authentication.md), Presentation of the consumer pages and Separation from the CDR authentication).

### Pages

- The landing page has:
  - A heading that names the brand, one sentence about the bank, and a link to the product list. The data holder gives this text, with the brand name in it. A brand directory does not change it. (See Decisions)
  - One link for each product category that has a current product, with the name of the category and the number of its current products. Each link goes to the group of that category in the product list. Judgement: a consumer looks for a type of product first.
  - A short statement that the bank shares the data of a consumer under the Consumer Data Right only with the consent of the consumer. Judgement: the site is where a consumer can learn about the CDR sharing of the bank, and the consumer dashboard is on the site.
- The text of the site makes no claim about rates, awards or customers, and copies no text of a real bank. Judgement: the bank is invented, and the text must not deceive a person (see Presentation of the consumer pages).
- The product list shows each current product of the brand, and no future product and no product whose `effectiveTo` is in the past. ([CDS: Get Products][cds-get-products]: `effective` is `CURRENT` by default; [a current product](design/product-reference-data.md#dh-0219); see Decisions)
- The product list puts the products in groups by `productCategory`, in the order of the categories in the CDS. Only a category with a current product has a group. In a group, the products are in the order of their names. ([CDS: Product Categories][cds-product-categories]; the order in a group is judgement: a consumer finds a product by its name)
- For each product, the product list shows its name and its description, and links to the detail of the product. Judgement: the name and the description are the required fields that a consumer reads ([CDS: BankingProductV6][cds-schema-product]).
- The detail of a product shows each field that the data holder holds for the product: the name, the category, the description, `isTailored`, `effectiveFrom` and `effectiveTo` when present, each deposit rate and each lending rate, each fee and each feature, with each of their fields that has a value. When `isTailored` is true, the page says that the rates and the fees can be different for the consumer. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
- The detail of a product that is not current, or that does not exist, is the "not found" page of the site. Judgement: the site shows only current products.
- The site shows the same products and the same values as Get Products and Get Product Detail of the brand at the same time. (See Site API)
- The site shows a label in plain language for each value of a CDS enumeration, for example `TRANS_AND_SAVINGS_ACCOUNTS` and `MONTHLY_ACCOUNT_FEE`. The label gives the meaning that the CDS gives for the value. ([CDS: BankingProductDetailV7][cds-schema-product-detail], [CDS: Product Categories][cds-product-categories]; [design system DESIGN.md: Content][ui-content]: plain language; see Decisions)
- The site shows each RateString as a percentage, for example `0.0125` as "1.25%", each AmountString as an amount, and each ISO 8601 duration in words, for example `P1M` as "each month" for an `applicationFrequency`, and `P55D` as "55 days". It shows an `additionalValue` in the format that the CDS gives for its type, and as text when the type gives no format. ([CDS: Common Field Types][cds-common-field-types], [CDS: BankingProductDetailV7][cds-schema-product-detail]; [design system DESIGN.md: Content][ui-content])
- A path that the site does not know shows the "not found" page of the site. Judgement: the UI routes in the browser, as the consoles do.

### Accounts, balances and transactions

- The accounts page shows each account of the customer of the session: first the open accounts, then the closed accounts in a group of their own. For each account, it shows the display name, the product name, the label of the product category, the masked number, the current balance and the available balance. (See Customers and accounts for the masked number)
- When the customer has no account, or no open account, the accounts page says so. It also says that the customer cannot share data under the Consumer Data Right until the customer has an open account, and links to the product list, where the consumer can open an account. ([CDR Rules][cdr-rules], rule 1.10B(1)(b); see Consent)
- The page of an account shows the fields of the accounts page, the credit limit, the status, the date of creation and the transactions of the account.
- The transactions are those of Get Transactions For Account with no filter: from 90 days before the time of the request to the time of the request, in the order of Get Transactions For Account, with pages. ([CDS: Get Transactions For Account][cds-get-transactions]: the default `oldest-time` is `newest-time` minus 90 days; see Transactions and Decisions)
- For each transaction, the page shows the effective date, the description, the amount with its sign, and "Pending" for a pending transaction. The other fields of the transaction that have a value show in the detail of the row, with a label in plain language for each value of a CDS enumeration. ([CDS: BankingTransactionV2][cds-schema-transaction]; [design system DESIGN.md: Content][ui-content])
- The site shows the same values as the APIs of a consent for the same account at the same time: the fields of Get Accounts, the balance of Get Account Balance and the transactions of Get Transactions For Account. Only the `accountId` and the `transactionId` are different. (See Site API and Decisions)
- An account of a different customer, or an account that does not exist, shows the "not found" page of the site.

### Opening of an account

- The consumer can open an account from each current product of the brand, in each CDS product category. The bank does no credit assessment, and it approves each open immediately. ([CDS: Product Categories][cds-product-categories]; see Decisions)
- Each current product can open, also a product for business and a product for a young person. The data holder has no rule for the age of the consumer or for a business. (See Decisions)
- The CDS lending categories are of two kinds (see Decisions):
  - Revolving credit: `CRED_AND_CHRG_CARDS`, `OVERDRAFTS`, `MARGIN_LOANS` and `BUY_NOW_PAY_LATER`. The consumer gives the credit limit.
  - Loans: `PERS_LOANS`, `BUSINESS_LOANS`, `RESIDENTIAL_MORTGAGES`, `LEASES` and `TRADE_FINANCE`. The consumer gives the loan amount.
- The operator still opens an account of each category, also a lending account with a credit limit (see Operator interface).
- The detail of each product has the action "Open an account", which goes to the open page of the product.
- The open page needs a session (see Sign-in and sign-out). It names the product and the label of its category, and it says that the bank opens an account of this product in the name of the consumer. It has the actions "Open account" and "Cancel". ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]: an account cannot be removed, so the page is the confirmation; see Decisions)
- The open page of a lending product has one required field: "Credit limit" for revolving credit, and "Loan amount" for a loan. The page says in plain language that this demonstration bank does no credit check, and that it approves the account immediately. The field accepts an amount in dollars as a person usually types it, for example `5000`, `5,000` or `$5,000.50`, and the page sends it as an AmountString. Text that is not an amount gives a field error in plain language on the page, and the page sends no open. (See Decisions)
- The credit limit and the loan amount are more than `0.00`. A value that is not valid gives a field error in plain language, and the data holder opens no account. The open of a deposit product has no credit limit and no loan amount.
- The data holder opens the account with the rules of a create of an account in the operator interface. The account gets these values (see Decisions):
  - `productCategory`: the `productCategory` of the product.
  - `productName` and `displayName`: the `name` of the product.
  - `bsb`: [the `bsb` of the brand file](design/brand.md#dh-0334).
  - `accountNumber`: 8 random digits. When the pair of the BSB and the account number is in use, the data holder makes a new number, with at most 3 attempts in all, as for a customer number. When the 3 attempts fail, the open fails, and the data holder changes nothing.
  - `creditLimit`: the credit limit of the consumer for revolving credit, and `0.00` for a deposit product and a loan.
- The open of a loan also adds one transaction to the account, in the same write as the account: the drawdown. When the write of the transaction fails, the data holder opens no account. The drawdown has the `type` `OTHER`, the `status` `POSTED`, the `description` "Loan drawdown", the `amount` minus the loan amount, an empty `reference`, and the time of the open as its `postingDateTime` and its `executionDateTime`. The account then has the current balance minus the loan amount and the available balance `0.00`, by the balance rule of Customers and accounts. ([CDS: BankingTransactionV2][cds-schema-transaction], [CDS: BankingBalance][cds-schema-balance]: `currentBalance` is "negative if the customer has money owing"; see Decisions)
- The values of the APIs of a consent for a lending account that the consumer opens are the same as for an account that the operator creates with the same credit limit and, for a loan, the same transaction.
- The account has no link to its product. A later change to the product, or its end, does not change the account. (See Decisions)
- The new account is not in an arrangement that exists. The accounts of an arrangement do not change after the confirmation ([non-goals: the amendment of an authorisation](design/README.md#non-goals)). The account is available for the next consent of the customer.
- The data holder decides the product with the same rule for a current product and the same products as the site API. A product that is not current and a product that does not exist give a field error in plain language, and the data holder opens no account.
- After an open, the site shows the page of the new account, with a status message that names the product, and the focus on the heading of the page. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- The busy state of the button stops a second open while the first is in progress. The site never sends an open again by itself. The data holder has no idempotency key for an open. (See Decisions)

### Close of an account

- The page of an open account has the action "Close account". A closed account has no action.
- Before a close, the site asks the consumer to confirm in a dialog. The dialog gives the display name, the masked number and the current balance of the account. It says that the close cannot be reversed, and that this demonstration bank pays out no balance. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4])
- When the account is the last open account of the customer, the dialog also says that the consumer then cannot share data under the Consumer Data Right, names the app (the `clientName`) of each consent that ends, and says that the bank tells each app. (See End of eligibility)
- The data holder closes the account with the rules of a close in the operator console: the close sends the arrangements that the confirmation named, and when the close now ends different arrangements, the data holder changes nothing. The dialog then says that the accounts changed, and the site shows the current state. (See Console API and Decisions)
- A close has the same effects as a close by the operator: the end of each arrangement for the end of eligibility at the close of the last open account, at the time of the close, the revocation of its tokens, and the notice to the data recipient. (See End of eligibility and Notice to the data recipient)
- A close has no rule for the balance or for pending transactions. The closed account keeps its balance and its transactions. (See Decisions)
- After a close, the dialog closes, a status message names the closed account, and the site shows the accounts page with the focus on its heading. The account is in the group of the closed accounts. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- When a close fails for a reason that is not a changed set of arrangements, the dialog stays open and shows the error. The site never sends a close again by itself.

### Consumer dashboard

The consumer dashboard is the online service of rule 1.15 of the CDR Rules. On it, the consumer sees each CDR consent that they gave to the bank, and withdraws a consent. ([CDR Rules][cdr-rules], rule 1.15(1))

- The dashboard is the page "Data sharing". It needs a session (see Sign-in and sign-out). Its title is "Data sharing - <brand name> online banking".
- The page shows each arrangement of the customer of the session: first the current arrangements, then the ended arrangements in a group of their own. In each group, the newest confirmation is first. An arrangement is ended after an event ended it, and at and after the end of its sharing, as in the console. An arrangement with a sharing duration of 0 is thus in the ended group at once. ([CDR Rules][cdr-rules], rule 1.15(3)(d) and (e); [CDR Support: Dashboards][support-dashboards]; see Decisions)
- When the customer has no arrangement, the page says that the customer shares no data under the Consumer Data Right. A customer with no open account sees the same page. (See Decisions)
- For each arrangement, the page shows ([CDR Rules][cdr-rules], rule 1.15(3) and rule 7.9(1); see Decisions):
  - The data recipient: the name of the software product (the `clientName`), the brand name (the `org_name`) and, when the arrangement has one, the legal entity name, from the arrangement (see CDR arrangement).
  - The data that the consumer authorised: each data cluster of its scopes, with its permissions, in the language of the table of Consent, word for word. ([CDS: Data Language Standards][cds-data-language]: the language applies also when the consumer reviews or withdraws an authorisation)
  - The accounts of the arrangement, with the display name and the masked number. A closed account shows its closed status. (See Decisions: Closed accounts)
  - When the consumer gave the authorisation: the date of the confirmation.
  - The period: one occasion, or the duration, as on the consent page (see Consent).
  - For a current arrangement, the date on which the sharing ends. For an ended arrangement, the date on which it ended: the time of the event that ended it, or the end of its sharing. The page does not show the reason or the method of the end. ([CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method]: "the CDR Rules do not require the method of withdrawal to be shown on the dashboard"; [CDR Support: Dashboards][support-dashboards])
  - For each data cluster, the date of the first and of the last disclosure, from the Record of disclosures. A data cluster with no disclosure shows "No data shared". This applies also to an ended arrangement. ([CDR Support: Privacy Safeguard 5 - dates on the consumer dashboards][support-dashboard-dates]: the first and the final disclosure date, "for both active and expired sharing arrangements"; [CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters])
  - The message of the CDS about the handling of the data, with the brand name of the data recipient in place of each name of the data recipient: "You should check with <data recipient brand name> for more information on how they are handling your data, and for any other permissions you may have given them. See <data recipient brand name>'s CDR policy or their Dashboard for more information." ([CDS: Dashboard Standards][cds-dashboard-standards]: Data recipient handling details; see Decisions)
- The page shows no accreditation number, no CDR receipt and no amendment. ([non-goals](design/README.md#non-goals); see Decisions)
- Each current arrangement has the action "Stop sharing". An ended arrangement has no action. ([CDR Rules][cdr-rules], rule 1.15(1)(c); [CX Guidelines: Withdrawal][cx-withdrawal]: "Stop sharing")
- Before a withdrawal, the site asks the consumer to confirm in a dialog. The dialog names the software product and the brand name of the data recipient. It says that the sharing stops at once, that the withdrawal cannot be reversed, and that the bank tells the data recipient. It has the actions "Stop sharing" and "Cancel". ([CDR Rules][cdr-rules], rule 1.15(1)(e); [WCAG 2.2, SC 3.3.4][wcag-3-3-4]; see Decisions)
- The dialog shows the two messages of the CDS, with the brand name of the data recipient in place of the name of the data recipient ([CDS: Withdrawal Standards][cds-withdrawal-standards]: Withdrawing authorisation: Consequences and Redundant data; [CDR Rules][cdr-rules], rule 1.15(1)(d); see Decisions):
  - "You should check with <data recipient brand name> before you stop sharing to understand the consequences."
  - "CDR data is either deleted or de-identified when it is no longer required."
- The data holder withdraws the arrangement with the use case of a withdrawal by the operator, with the customer of the session and the method `CONSUMER_DASHBOARD` (see Withdrawal by the consumer).
- After a withdrawal, the dialog closes, a status message names the software product, and the site shows the dashboard with the focus on its heading. The arrangement is in the group of the ended arrangements. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [CX Guidelines: Withdrawal][cx-withdrawal]: a message that the withdrawal was successful; [CDR Rules][cdr-rules], rule 4.27)
- When the arrangement ended between the read and the confirmation, the dialog says that the consent changed, and the site shows the current state. When a withdrawal fails for a different reason, the dialog stays open and shows the error. The site never sends a withdrawal again by itself.
- The data holder sends no email and no CDR receipt to the consumer after a withdrawal. (See Decisions: Notice email to the consumer)

### Presentation

- The site uses the tokens and the components of the design system, with the colours, the fonts and the radius of the brand. It has a shell of its own (header, main area and footer), not the shell of the operator console. ([design system DESIGN.md: Brands][ui-brands])
- Each brand has one of four layouts, which the `layout` of the brand directory selects. A layout changes the arrangement and the shapes of the landing page, of the account cards and of the other pages. It does not change the text, the content, the order of the content in the product list or the accessibility of a page. Judgement: with the logo and the brand name hidden, a person must see which bank a page is from, and colours alone do not do this. (See Decisions)

  | Layout | Hero photo | Product categories on the landing page | Account cards |
  |---|---|---|---|
  | `meadow` | 4:3, in a frame with round corners, with the text on a panel in the brand colour that overlaps it | Cards with a photo and a count | Cards in the brand colour |
  | `ledger` | 21:9, full width, with the text on a panel that overlaps its lower edge | A ruled list with a square photo, the name and the count | Rows in the hero colour, with the balances in columns |
  | `horizon` | 3:1, full width, above a band in the hero colour with the text | Cards with a 3:2 photo and a count | Cards in the hero colour, like a ticket |
  | `updraft` | 1:1, in a panel in the hero colour with the text, cut at an angle | A grid of tiles with a photo and a label on a solid panel. The first tile is large and is `TRANS_AND_SAVINGS_ACCOUNTS` when the brand has a current product in it; the other tiles are in the order of the CDS | Cards in the hero colour |

- The landing page shows the hero photo, and each link of a product category shows the photo of its category. The product list shows the photo of each category at its group, and the detail of a product shows the photo of its category. The sign-up and sign-in pages show the hero photo.
- Each photo is decoration, with an empty text alternative. Text on the page gives the meaning of each photo. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]; see Decisions)
- When the brand has no photo for a place, the page shows a shape in the brand colour with the mark, and never a broken image.
- No text is on a photo. A text that is in front of a photo is on a solid panel, so that its contrast does not depend on the photo. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3])
- The accent colour of a brand is only for shapes, not for text. Each pair of colours that a layout uses for text, or for the boundary of a control, is in the contrast test of the design system for each preset brand. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3], [SC 1.4.11][wcag-1-4-11]; see Decisions)
- Headings use the heading font of the brand, and the other text uses the body font.
- Each page shows the logo of the brand on the brand colour, with the brand name as its text alternative, and a link to the product list. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- The pages of the site do not look like the authentication and consent pages: no "Secure" label, no hero with the data recipient and no step indicator. The title of each page is the name of the page, then "<brand name> online banking". Judgement: the consumer must not confuse the site with the CDR authentication. The sign-in of the site asks for a password, and the CDR authentication never does (see Separation from the CDR authentication).
- Each page has a footer with two statements, in this sequence:
  - Third parties do not need the password of the consumer to get CDR data. ([CDS: Common Authentication Standards][cds-cx-authentication]: the statement is necessary "in consumer-facing interactions and communications"; the footer is judgement: a consumer can open each page first)
  - The text of the footer of the consumer pages: "<brand name> is not a real bank and this is just a demonstration. CDR Lab has no affiliation with any real bank nor the ACCC, DSB, or any real CDR participant." The site API gives this text. A brand directory cannot change or remove it. (See Presentation of the consumer pages)
- The footer has no link. ([design system DESIGN.md: Consumer pages][ui-consumer-pages])
- Until the site API gives the brand, and when it fails, the pages say "Bank" in place of the brand name. Judgement: the same rule as the operator console, and "Bank" is true for each brand.
- The site meets WCAG 2.2 Level AA, and the automated accessibility checks pass on each page, with the brand of each preset brand. ([ADR 0016][adr-0016]; [design system DESIGN.md: Accessibility][ui-accessibility])
- Each page is designed for a viewport width of 360 CSS px, and is usable at 320 CSS px with no horizontal scroll. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10]; [design system DESIGN.md: Layout][ui-layout])
- The site shows dates and amounts with the rules of the design system: Australian English, `Intl` and the locale `en-AU`. ([design system DESIGN.md: Content][ui-content])

### Site API

The site API has [the `/api/` prefix](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0011) and [404 for a path that is not an operation](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0012) of the web UI library. It has the rules of the console API of the data holder ([Console API](#console-api)) for: the OpenAPI document in the data holder directory and its tests, the problem details and their status codes, no version in the path, a body that is not valid, [a body with a field that the OpenAPI document does not give](../../pkg/platform/jsonhttp/DESIGN.md#body-0001), the field errors with [the field-errors type](../../pkg/platform/problem/DESIGN.md#prob-0002), and no request body in telemetry. These rules are different:

- The brand gives the `layout` of the brand directory, and the name of each photo that the brand directory has. The site asks only for a photo in the list. Judgement: the site then never shows a broken image (see Presentation).
- The fields of a product have the names and the values of the CDS product schemas. ([CDS: BankingProductV6][cds-schema-product], [CDS: BankingProductDetailV7][cds-schema-product-detail]; judgement: a test can then compare each field with Get Products and Get Product Detail, with no conversion)
- The list of the products gives each current product of the brand in one response, with no pages. Judgement: a catalogue file is 1 MiB or less ([the size of a JSON file of the brand directory](design/brand.md#dh-0350)), and the site shows all current products on one page.
- The detail of a product gives one current product. A product that is not current, or that does not exist, gives 404, and the `detail` does not contain its identifier.
- The site API and Get Products use the same rule for a current product and the same time of the request. For each preset brand, the list gives the products and the values of Get Products with `effective=CURRENT`, and each detail gives the values of Get Product Detail. A test in the data holder checks this. ([CDS: Get Products][cds-get-products], [CDS: Get Product Detail][cds-get-product-detail]; see Decisions)
- The OpenAPI document gives the values of each CDS enumeration of a product. A test checks that each list is the same as the values that the data holder accepts. (See Decisions)
- No path of the site API contains a customer number or a different identifier of a consumer. Judgement: the session gives the consumer of a request, and a path is in telemetry [as a route pattern only](../../pkg/platform/telemetry/DESIGN.md#tele-0005).
- The sign-up has the fields `firstName` (optional), `lastName` and `emailAddress`, with the rules of the operator interface. When the data holder refuses a value, the response is 400 with a field error for each value. The `detail` of each field error is a text in plain language for the consumer, which names the field with its label on the page. Judgement: the UI then has no copy of the rules, as in the operator console, and a consumer does not read the names of the fields. (See Decisions)
- A sign-up with an email address that a different customer of the brand has gives 409 with a field error at `/emailAddress`. (See Decisions)
- The sign-in has the field `identifier` only: the email address or the customer number. A body with a `password` field gives 400. A value that no customer has gives 400 with a field error at `/identifier`. (See Decisions)
- A sign-up and a sign-in give the session cookie in the response. A sign-up gives 201 with the `customerNumber`, the `firstName` and the `lastName` of the new customer, and with no `Location` header. Judgement: no path of the site API contains a customer number.
- A request for the session, the accounts, a balance, the transactions, an open, a close, the consents or a withdrawal with no valid session gives 403. When the request has a session cookie, the problem has the type `https://github.com/lukemundy/cdr-lab/blob/main/services/dataholder/DESIGN.md#session`, and the response expires the cookie. The site then says that the bank signed the consumer out. (See Decisions)
- The fields of an account have the names and the values of `BankingAccountV3` in Get Accounts: `displayName`, `maskedNumber`, `productCategory`, `productName`, `openStatus` and `creationDate`, and `accountId`, which is the account identifier of the data holder. The fields of a balance have the names and the values of `BankingBalance` in Get Account Balance: `currentBalance`, `availableBalance` and `creditLimit`. The fields of a transaction have the names and the values of `BankingTransactionV2` in Get Transactions For Account, and `transactionId` is the transaction identifier of the data holder. The site API gives no `bsb`, no `accountNumber` and no `extendedData`. ([CDS: BankingAccountV3][cds-schema-account], [CDS: BankingBalance][cds-schema-balance], [CDS: BankingTransactionV2][cds-schema-transaction]; judgement: a test can then compare each field with the APIs of a consent, with no conversion)
- Each account in the list of the accounts has a `balance` with the fields of a balance, so that the accounts page needs one request. Judgement: the balances come from one query of the totals, as in Get Bulk Balances ([CDS: Get Bulk Balances][cds-get-bulk-balances]).
- The data holder gives the accounts, the balances and the transactions of the site with the same functions as the APIs of a consent, but for the customer of the session in place of an arrangement. The accounts are in the order of Get Accounts. An account of a different customer, or an account that does not exist, gives 404, and the `detail` does not contain its identifier. (See Decisions)
- An open has the field `productId`, the field `creditLimit` only for revolving credit, and the field `loanAmount` only for a loan. The product ID is in the body, not in the path. The credit limit and the loan amount are AmountStrings, with the rule of the operator interface for a `creditLimit`, and they are more than `0.00`. A product that is not current and a product that does not exist give 400 with a field error at `/productId`. An amount that is absent, not valid, or for a product of a different kind gives 400 with a field error at its field. A successful open gives 201 with the new account, with the fields of an account in the list of the accounts, and with no `Location` header. (See Decisions)
- The OpenAPI document gives the categories of revolving credit and of a loan. The UI selects the field of the open page from these lists. A test checks that each list is the same as the rule of the data holder. (See Decisions)
- Each open account in the list of the accounts has the action `close`, with each arrangement that the close ends: its `arrangementId` and its `clientName`, in the sequence of their confirmation. The list is empty when the account is not the last open account. A closed account has no action. The data holder finds these arrangements with the same rules as the end itself. (See Console API)
- A close has the field `ends`: the `arrangementId` of each arrangement that the confirmation named, and an empty list when it named none. A successful close gives 204. A close of a closed account, and a close that now ends different arrangements, give 409, and the data holder changes nothing. An account of a different customer, or an account that does not exist, gives 404, and the `detail` does not contain its identifier.
- The list of the consents gives each arrangement of the customer of the session in one response, with no pages: the `arrangementId`, the `clientName`, the `brandName`, the `legalEntityName` when the arrangement has one, the `status` (`CURRENT` or `ENDED`, as in the console API), `confirmedAt`, `sharingDurationSeconds`, `sharingEndsAt`, `endedAt` for an ended arrangement (the time of the event that ended it, or the end of its sharing), the `accounts` with the fields of an account in the list of the accounts and no `balance` and no `actions`, and the `dataClusters`. Each data cluster has its `name`, its `permissions`, and `firstDisclosedAt` and `lastDisclosedAt` when it has a disclosure. The data holder gives the language of the data clusters, so that the UI has no copy of the rules. (See Decisions)
- A withdrawal sends no body, because it ends only the arrangement of its path. A successful withdrawal gives 204. A withdrawal of an arrangement that ended gives 409, and the data holder changes nothing. An arrangement of a different customer, or an arrangement that does not exist, gives 404, and the `detail` does not contain its identifier. (See Console API)
- The arrangement IDs, the names of the data recipients and the disclosure times of the consents and of the close action go only to the customer of the session.
- The transactions of an account take only `page` and `page-size`, with [the defaults and the maximum of CDS pagination](../../pkg/cds/DESIGN.md#cdsl-0017). The response gives the number of records and the number of pages. ([CDS: Pagination][cds-pagination])

| Action | Method and path | Success |
|---|---|---|
| Read the brand | `GET /api/brand` | 200 with the brand name, the text of the footer, the layout and the photos of the brand |
| List the current products | `GET /api/products` | 200 |
| Read a current product | `GET /api/products/{productId}` | 200 |
| Sign up | `POST /api/customers` | 201 with the new customer, and the session cookie |
| Sign in | `POST /api/session` | 204 with the session cookie |
| Read the customer of the session | `GET /api/session` | 200 with the `customerNumber`, the `firstName` and the `lastName` |
| Sign out | `DELETE /api/session` | 204, and the cookie expires, also with no valid session |
| List the accounts | `GET /api/accounts` | 200 |
| Read the balance of an account | `GET /api/accounts/{accountId}/balance` | 200 |
| List the transactions of an account | `GET /api/accounts/{accountId}/transactions` | 200 |
| Open an account | `POST /api/accounts` | 201 with the new account |
| Close an account | `POST /api/accounts/{accountId}/close` | 204 |
| List the consents | `GET /api/consents` | 200 |
| Withdraw a consent | `POST /api/consents/{arrangementId}/withdraw` | 204 |

### Non-functional requirements

- Telemetry: the site API has the same server telemetry as the other listeners, and the rules of the data holder telemetry apply to it. The spans contain [no path value](../../pkg/platform/telemetry/DESIGN.md#tele-0005) and [no query string](../../pkg/platform/telemetry/DESIGN.md#tele-0007). No span, metric or log contains a request body, a cookie value, a session ID or a password. No log contains a query string or a path value. No log of a sign-up, a sign-in, a sign-out or the end of a session contains the customer number or the email address, also for a refused sign-up or sign-in. No log of an open or a close contains the customer number, the product ID, the BSB, the account number, the display name, the credit limit, the loan amount or an arrangement ID. No span, metric or log of the consent list, of a withdrawal on the site or of the record of a disclosure contains the customer number, an arrangement ID, the name of the software product, the brand name or the legal entity name of a data recipient, a data cluster or a time of a disclosure. Judgement: with the session, a product ID tells which product a consumer has, and the name of a data recipient tells with which app a consumer shares data. The UI sends no telemetry. ([ADR 0014][adr-0014], [ADR 0006][adr-0006]; [no secret values in telemetry](design/non-functional-requirements.md#dh-0007), [no personal information in telemetry](design/non-functional-requirements.md#dh-0008))
- Performance: on a development workstation, the landing page and the product list of the largest preset catalogue show within 1 second, and the first page of 10,000 transactions of an account shows within 1 second. Judgement: the same limits as the consoles.
- Tests: tests in the data holder check the listener and the site API, and that the site gives the same products as Get Products and Get Product Detail for each preset brand. Component tests check each page with the brand of each preset brand at 320 CSS px. One browser test goes through the landing page, a category and the detail of a product against the lab. The conformance suite does not test the site. ([ADR 0019][adr-0019])
- Tests of an open and a close: tests in the data holder check that an open of a product that is not current, of an unknown product, of a lending product with no valid amount and of a deposit product with an amount gives 400 and opens no account, that a failure of the write of the drawdown opens no account, that a close of a closed account gives 409, that a close of an account of a different customer gives 404, and that an open and a close need a session.
- Tests of the session: tests in the data holder check the sign-up, the sign-in, the refusal of a `password` field, a new session ID at each sign-in and sign-up, the sign-out, the end of a session after 24 hours, the limit of 10,000 sessions, the attributes of the cookie, and the 404 for an account of a different customer.
- Tests of the account data: a test in the data holder, in one process, creates a customer, accounts and transactions, and an arrangement for the accounts. It calls the APIs of a consent with a token verifier that gives that arrangement, and the site API with a session of the same customer. It checks that the site gives the same values as Get Accounts, Get Account Balance and Get Transactions For Account, field by field, except `accountId` and `transactionId`. It needs no CDR authentication flow. (See Decisions)
- Tests of an account that the consumer opens: a test in the data holder, in one process, signs up, opens two accounts from two deposit products with the site API, and creates an arrangement for both. It checks that Get Accounts gives the same `displayName`, `productCategory`, `productName`, `maskedNumber`, `openStatus` and `creationDate` as the site, with `productCategory` and `productName` from the product. It then closes one account with the site API, and checks `openStatus` `CLOSED` in Get Accounts. A test in the data holder, in one process, opens a credit card with a credit limit and a home loan with a loan amount with the site API, and checks Get Account Balance and Get Transactions For Account of the next consent: the card has the credit limit and no transaction, and the loan owes the loan amount and has the drawdown.
- Tests of the end of eligibility through the site: a test in the data holder, in one process, with no waits, closes the last open account of a customer with the site API and the arrangements of its close action. It checks that each arrangement of the customer that did not end has the end reason `CONSUMER_NOT_ELIGIBLE` at the time of the close, that the data holder refuses its tokens, and that its notice is `PENDING`. It also checks that a consent that starts after the read of the close action gives 409, and that the account stays open.
- Tests of the consumer dashboard: a test in the data holder, in one process, with no waits, creates a registration, a customer with an account, and an arrangement with a grant. It calls an API of a consent with the access token, and then reads the consents with the site API: the arrangement shows with its data clusters and the time of the first and the last disclosure, its accounts, and the end of its sharing. It then withdraws the arrangement with the site API. It checks that the arrangement has the end reason `WITHDRAWN_BY_CONSUMER` and the method `CONSUMER_DASHBOARD` in the read model and in the operator interface, that the token verifier of the data holder refuses the access token, that the grant of the refresh token is revoked, and that its notice is `PENDING`. A test of the authorisation server proves that a refresh with a revoked grant gives `invalid_grant`. (See Decisions)
- Tests of the withdrawal on the site: tests in the data holder check that a withdrawal of an arrangement of a different customer and of an unknown arrangement gives 404 with no identifier in the `detail`, that a withdrawal of an ended arrangement gives 409, that the consents and a withdrawal need a session, and that a withdrawal by the operator still records the method `ALTERNATIVE_METHOD`.
- Tests of the record of disclosures: tests in the data holder check that a 2xx response of each API of a consent and of the UserInfo endpoint with the scope `profile` gives a record with its data cluster, and that a response that is not 2xx and a UserInfo response with no `profile` scope give no record.
- Component tests check the dashboard and the dialog of a withdrawal with the brand of each preset brand at 320 CSS px, with the exact text of the CDS messages.
- Tests of the telemetry: a test in the data holder does a sign-up, a sign-up with an email address in use, a sign-in with a password, a sign-in with the email address, a refused sign-in with an email address, reads the accounts, a balance and the transactions, and does a sign-out. It also opens an account and closes the last open account of a customer with an arrangement. It also calls an API of a consent, reads the consents and withdraws a consent on the site. It checks that no span, metric or log contains the customer number, a name, the email address, the masked number, the BSB, the account number, an account ID, an arrangement ID, the name of the software product, the brand name or the legal entity name of the data recipient, the description of a transaction, the session ID or the password.
- Browser tests: one browser test against the lab signs up, reads the customer number, signs out, signs in with the number and a password, and sees no account. It then creates an account and a transaction with the operator interface, and sees them on the site. Last, it signs out, and signs in with the email address in a different case. It makes its own data. One browser test signs up, opens an account from the detail of a deposit product, sees it with a balance of 0, closes it with the confirmation, and sees it in the group of the closed accounts with the text that the customer cannot share data. It makes its own data and gives no consent. One browser test signs up, opens revolving credit with a credit limit and a loan with a loan amount, and sees the balances and the drawdown. It makes its own data. One browser test checks that the sign-in page of the site and the customer number page of the CDR authentication have different titles and headings, and that only the sign-in page of the site has a password field. One browser test signs up, opens Data sharing from the header, and sees that the customer shares no data. It makes its own data. One browser test finds a customer with a current arrangement with the console API, signs in to the site as that customer, opens Data sharing, stops the sharing with the confirmation, and sees the consent in the group of the ended consents. A consent needs the full consent flow of a registered software product, so the test uses an arrangement that the conformance suite makes, as the withdrawal test of the console does. It skips when the lab has no current arrangement, and so it skips in CI. The test of the consumer dashboard in one process is the proof in CI. (See Decisions)

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Search of the customers in the console | A part of the customer number, the name or the email address | Judgement: a consumer who asks the staff for their records gives one of these values ([CDR Rules][cdr-rules], rule 9.5). The query string then has personal data, and [the telemetry library removes the query string from each span](../../pkg/platform/telemetry/DESIGN.md#tele-0007). |
| Account numbers in the console | In full | Judgement: the operator is the staff of the bank, and the operator interface gives the account number in full. The pages for the consumer mask it, because a different person can see the screen of a consumer. |
| Version of a record in the console API | No version. The state of the record refuses an action that is not permitted now. A close also sends the arrangements that its confirmation named. A withdrawal sends nothing. | The customers, the accounts and the transactions have no history and no version (see Operator interface), and a transaction and a customer cannot change. A close and a withdrawal change only a record that is open, and a second close or withdrawal gives 409 ([operator interface](design/operator-interface.md)). A close of a different account of the customer can make a close the last close, which then ends arrangements that its confirmation did not name. Judgement: the arrangements in the request protect against this change, and a version of the customer would also refuse changes that do not change the effects of the close. A withdrawal ends only the arrangement of its path, so no other change makes its effect different from its confirmation, and a withdrawal of an arrangement that ended gives 409. |
| Field of a rule with two fields | The field that is absent: the `postingDateTime` of a `POSTED` transaction, the `executionDateTime` of a `PENDING` transaction, and `extendedData.payee` or `extendedData.payer` for the sign of the amount | Judgement: the operator corrects the rule with a value in that field. The status and the amount are usually correct, and the error of the field must show at the field that the operator changes ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]). |
| Field of a `bsb` and an `accountNumber` that a different account has | `accountNumber`, with the status 409 of the operator interface | Judgement: the operator corrects the pair with a different account number, because the accounts of a brand usually have the same few BSBs. The Register console API also gives an identifier that a different record has as 409 with a field error ([Register DESIGN.md: Console API][register-console-api]). |
| Location of a new transaction in the console API | `/api/customers/{customerNumber}/accounts/{accountId}/transactions/{transactionId}`, a path with no `GET` | [RFC 9110, section 15.3.2][rfc9110-201]: the `Location` of a 201 identifies the new resource. Judgement: the path of the list does not identify one transaction. The list of the transactions gives each field, so a detail of a transaction adds a second view of the same data, and the UI does not read it. The `transactionId` in the path is the identifier of the operator interface, and no telemetry has a response header. |
| Permitted create of a transaction | An action of the account (`create-transaction`), with the rules of the operator interface: only an open account | The console API gives the actions that the rules permit now (see Rules for the actions), and the UI has no copy of the rules. Judgement: the state of the account permits the create, so the action is with the other actions of the account, as the close is. The form is a link on the transactions of the account, not a confirmation, because a create changes no other record. |
| Group of the extended data in the form of a transaction | A group of fields with a legend that is always open. The console sends `extendedData` only when a field of the group has a value. | Judgement: the design system has no group of fields that opens and closes ([design system DESIGN.md][ui-design]: only a card of a record collapses). A collapsed group also hides its field errors and the target of the links of the error summary. The values of the group are the only sign that the operator wants it, so an empty group sends nothing and a group with one value sends each field, and the data holder refuses each value that is missing. |
| Time of a transaction in the form | A native date-time control in the time zone of the browser, with seconds. The console sends the instant in UTC. | [ADR 0012][adr-0012]: the API gets each time in UTC, and the UI shows each time in the time zone of the browser. Judgement: the operator thinks of a time of a test in local time, and the control of the browser gives a valid local time or no value. The seconds make each time of the CDS available. |
| Source of the brand name in the UI | The console API gives it (`GET /api/brand`). Until the response, and when it fails, the header and the tab title say "Data holder". | Judgement: the console API is the only source of data for the UI ([ADR 0014][adr-0014]), and a change of the HTML page at runtime adds a second contract. "Data holder" is true for each brand. |
| Count of the arrangements of a customer in the list | The arrangements that did not end at the time of the request: no end, and the end of the sharing is after the time of the request. | Judgement: an arrangement ends at the end of its sharing also with no event (see The end of a consent), so the count uses the same rule as the end. |
| Views that are not in the console | The products, the signing keys, the one-time passwords and a history of events. The overview gives only the 10 newest changes. | Judgement: Get Products gives the products. The data holder has one signing key, and no change of the key is in scope. A one-time password lets the staff authenticate as the consumer. Only the arrangements and the client registrations have events, and the detail of an arrangement already gives its end and its notice. |
| Status of an arrangement in the console | The console API gives `CURRENT` or `ENDED`. An arrangement is `ENDED` after an event ended it, and at and after the end of its sharing. `endedAt` and `endReason` are only for an event. | Judgement: the title bar of a card shows the status of the record ([design system DESIGN.md][ui-design]), and the UI has no copy of the rules. The operator interface gives no reason for the end of the sharing, so the console API does not invent one. |
| Date of creation of an account in the console | `creationDate`, the date in UTC | Judgement: it is the same value as the `creationDate` of the banking APIs ([CDS: BankingAccountV3][cds-schema-account]), so the operator sees what the data recipient gets. |
| Period of the transactions in the console | All the transactions of the account, with no time limit | Judgement: the banking APIs give a data recipient only 2 years ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(6)(a); see Period of the transactions), but the staff of the bank see all the records of the customer. |
| Transaction fields in the console | The time, the description, the type, the status and the amount are columns. A details column gives each other field that has a value. | Judgement: the table stays readable at 1280 CSS px, and it still shows each field that the operator interface gives. |
| Time of the last response of the Register in the console | The time of the last response of Get Software Products Statuses, also a response with no change (304). A request that fails does not change it. The console API gives it as `receivedAt`, and does not give it until the data holder gets the statuses one time. | Judgement: a 304 tells the data holder that its statuses are current at that time, so the operator can see that the data holder still gets the statuses. A failed request leaves the time as it was, so an old time shows that the data holder cannot get the statuses. |
| Client registrations in the console API | A registration has `registeredAt`, `lastUpdated`, `softwareProductStatus` when the data holder has a status, and `metadata`: each value of the registration with the names of the registration metadata, except `software_statement`. The list gives the same object for each registration. | The CDS gives the names of the metadata ([CDS: Dynamic Client Registration][cds-dcr-endpoints]). Judgement: the console values are not metadata, so they stay out of the object with the CDS names. The SSA is a signed token that the Register issued, and the console shows no tokens. One object for the list and the detail is the least complex, and the lab has few registrations. |
| Times of a registration in the console | `registeredAt` is the time of the last registration of the software product. `lastUpdated` is the time of the last `PUT`, or `registeredAt` before a `PUT`. | Judgement: the events of the registration give the times. A registration after a `DELETE` has a new client ID, so it is a new registration with its own time. |
| Order of the client registrations in the console | The `client_name` by default, then the client ID | Judgement: the operator finds a software product by its name. The client ID makes the order stable. |
| Sorts of the lists in the console | The sorts of the Register console: text in the byte order, a name with no difference for case, and the identifier of the record last ([Register DESIGN.md: Decisions][register-decisions]). The customer list and the registrations list also sort by their counts. | The operator asked for the count sorts, as on the data recipient list of the Register. Judgement: the same rules in both consoles. The number of all the arrangements of a client is a column of the read model, as are the counts of the Register. A count of the arrangements that did not end depends on the time of the request (see Count of the arrangements of a customer in the list), so the data holder calculates it at the request from the arrangements of the read model. |
| Count of the arrangements of a registration in the list | The arrangements of its client ID that did not end, with the rule of the count of a customer, and all the arrangements of its client ID | Judgement: the operator asked for both. The first is the current use of the software product. The second shows that a client had arrangements that ended. |
| Name of a software product in the Register statuses | The `client_name` of its registration. A software product with no registration has no name. | The data holder permits one registration for each `software_id` (see Registration), and the registration metadata gives `client_name` as the "Human-readable string name of the software product" ([CDS: Dynamic Client Registration][cds-dcr-endpoints]). Judgement: the data holder does not read participant data from the Register now, and Get Data Recipients would add a second Register client only for names. |
| Software products with no registration in a sort by name | After the registered software products, in both directions | Judgement: the operator asked for it. The registered software products are the ones that the operator looks for, and a name that does not exist has no place in the order of the names. |
| Format of a BSB in the console | Two groups of three digits with a hyphen. The copy gives the 6 digits. | Judgement: the operator asked for the usual format of a BSB. The create form, the operator interface and the banking APIs use the 6 digits, so the copy gives a value that each of them accepts. |
| Link from an arrangement to a registration | By the client ID of the arrangement. When no registration has that client ID, the arrangement has no link. | Judgement: the registration of the client at the consent is the one that the arrangement uses. After a `DELETE`, a new registration of the same software product has a different client ID and a different set of values, so a link to it shows the operator values that the arrangement did not use. |
| Arrangements of a registration in the console | A table with pages, apart from the detail of the registration, with the page rules and sizes of the transactions of an account. Newest first by the time of the confirmation, then by the `arrangementId`. | Judgement: the operator asked for a table with pages, because a software product can have many arrangements, and one response or a card for each arrangement does not scale. The operator looks for the newest use of a software product first. The `arrangementId` makes the order stable. A separate path keeps the detail of the registration one response with one size. |
| Customer in the arrangements of a registration | The customer number and the name of the customer. No email address. | Judgement: the operator asked for the customer of each arrangement. The name lets the operator find the consumer, and the customer number is the identifier of the link. The detail of the customer gives the email address, and one more column makes the table wider than the 1280 CSS px of the console (see Presentation). |
| Brand in the console | The shell of the design system with the brand colours: the logo on the brand colour with an accent rule, the side navigation on the hero colour, the page header band and the heading colour. The link and the primary action use the brand. The font, the radius, the other colours and the density stay in the neutral lab brand. The mark is the icon of the tab. | Judgement: the operator asked for consoles that show more of the brand and are less white and flat, but stay dense staff tools. The brand colour is white for three preset brands, so the side navigation uses the hero colour. A different font changes the width of each table column. The label "Operator console" stops an operator from mistaking the console for a consumer page. ([design system DESIGN.md: Brands][ui-brands]) |
| Content of the overview | Counts and the 10 newest changes. The customer number, but no name, email address, display name of an account or account number. | Judgement: a screenshot of the overview then shows no personal information. The customer number is the identifier that the operator needs to find the customer, and it identifies a person only with access to the console. A display name is text that a consumer can choose. |
| Changes in the overview | The creation of a customer, the creation and the close of an account, the confirmation and the end of an arrangement, and the registration, the change and the deletion of a client registration. Not the transactions and not the notices. | Judgement: a lab can have thousands of transactions, which push out all other changes. The transactions of an account and the notice of an arrangement show on their own pages. |
| Counts of the overview | One endpoint for the counts and the changes, from one read | Judgement: two of the counts from the lists need each page of the customers, and the counts of one read agree with each other. |
| Text alternative of the photos | Empty. No alt text field in the brand directory. | Judgement: each photo is decoration, and the text of the page gives its meaning ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]). A field for the text is a change of the contract that gives no information. |
| Accent colour on the site | Shapes only, no text | Judgement: the accent of some presets has a contrast lower than 4.5:1 on white or on the hero colour, and the contrast test of the presets checks only the pairs that the layouts use as text ([WCAG 2.2, SC 1.4.3][wcag-1-4-3]). |
| Paths of the brand assets | On the console listener: `/brand/logo.svg`, `/brand/mark.svg` and `/brand/console.css`, which sets only the roles of the brand that the console uses. The paths of the public listener are in [the paths of the brand files](design/authorisation.md#dh-2090). | Judgement: the console has the same paths for the same files, and a different name for a stylesheet with different content. |
| Architecture of the online bank site | A React SPA that the data holder embeds and serves, with a site API in an OpenAPI document | [ADR 0014][adr-0014] names a data holder consumer dashboard as a SPA, and the design guide expects the first React consumer UI to add the consumer components to the design system ([design system DESIGN.md: Consumer pages][ui-consumer-pages]). Judgement: the later forms and dialogs of the site need the components that only the React package has, and Go pages put the bank sign-in near the CDR pages. |
| Source of the products of the site | The same products of the brand as Get Products and Get Product Detail, with the same rule for a current product. A test compares them for each preset brand. | Judgement: a consumer and a data recipient then see the same products, and one source of the products gives it by construction. A call from the site to the CDS API adds a cross-origin call or a call to its own service over the network. |
| Products that the site shows | The current products only. No future product and no product whose `effectiveTo` is in the past. | Get Products gives the current products by default ([CDS: Get Products][cds-get-products]). Judgement: a bank web site lists the products that the bank sells now. |
| Labels of the CDS enumerations | A label in plain language for each value of each enumeration of a product. The build of the site fails when a value has no label. The OpenAPI document of the site gives each enumeration, and a test compares each list with the values that the data holder accepts. | Judgement: a consumer does not read a CDS code. With the two checks, a new value in the data holder or in the CDS cannot show with no label. |
| Text of the landing page | Text of the data holder, the same for each brand, with the brand name. The site API gives the brand name and the text of the footer. No new field in the brand directory. | Judgement: a new field changes the contract of the brand directory ([ADR 0023][adr-0023]), and the site needs no text for each brand now. The footer then has one source, as on the consumer pages. |
| Password statement on the site | In the footer of each page | The CDS requires the statement "in consumer-facing interactions and communications" ([CDS: Common Authentication Standards][cds-cx-authentication]), and the site is a consumer-facing interaction of the data holder. Judgement: the footer is on each page, and a consumer can open each page first. |
| Tests of the online bank site | Tests in the data holder, component tests with axe for each preset brand at 320 CSS px, and one browser test against the lab. Not in the conformance suite. | [ADR 0019][adr-0019]: the conformance suite tests only behaviour that the CDR defines, and a bank web site is not CDR behaviour. |
| Field errors of the site API | A `detail` in plain language for the consumer, from the site API | Judgement: the `detail` of the operator interface names the fields in camelCase, which a consumer does not read ([design system DESIGN.md: Content][ui-content]). The UI then has no copy of the rules, as in the operator console. |
| Response of the site API with no valid session | 403 | Judgement: a 401 response must have a `WWW-Authenticate` challenge ([RFC 9110, section 15.5.2][rfc9110-401]), and a cookie session has no HTTP authentication scheme. 403 is the response when the server refuses the request ([RFC 9110, section 15.5.4][rfc9110-403]). The UI then goes to the sign-in page. Only the server knows if a session cookie names a session that ended, so a problem type of its own tells the UI, also after a reload, a restart or the end of the 24 hours. A sign-out needs no session: the consumer is signed out after it in each case. |
| Account data of the online bank site | The functions of the APIs of a consent: the balance of Customers and accounts, and the transaction query of Get Transactions For Account. The fields and the values of the CDS schemas, and the account and transaction identifiers of the data holder. | Judgement: one source of the data gives the same values as the APIs of a consent by construction, as for the products. The `accountId` and the `transactionId` of a consent are for each software product (see Resource identifiers), and the site has no software product. |
| Account number on the online bank site | Masked, as on each page for the consumer | The CDS display rule for a masked number ([CDS: Common Field Types][cds-common-field-types]) and the rule of Customers and accounts: a different person can see the screen of a consumer. |
| Transaction period on the online bank site | The default period of Get Transactions For Account: 90 days to the time of the request. No filters. Pages with the defaults of the CDS. | [CDS: Get Transactions For Account][cds-get-transactions]. Judgement: with no filter, the site gives the same list as Get Transactions For Account with no filter, and a test compares the two with no conversion. |
| Test of the account data of the online bank site | A test in the data holder, in one process, with a token verifier that gives an arrangement. No CDR authentication flow. | Judgement: the test then compares the values with no browser and no mail sink. The conformance suite does not test the site ([ADR 0019][adr-0019]), and the lab-web job in CI has no conformance data, so the browser test makes its own data with the operator interface. |
| Products that a consumer opens on the site | Each current product, in each CDS product category. The bank does no credit assessment, and it approves each open immediately. | The CDS and the ACCC give no rule for how a data holder opens an account. Judgement, from the spirit of the CDR: the lab is a development environment in which engineers build and test the apps of data recipients, and it has no real consumers. A data recipient must be able to test lending data from a consent that a consumer gives, and a product that the site lists but does not open has no value. A credit provider assesses a credit contract with a consumer before it enters the contract ([National Consumer Credit Protection Act 2009][nccp-act], Chapter 3), but the lab enters no credit contract with a real consumer, so the Act does not apply. The site is lab-only, and a production data holder has its own site, so this is not a production gap. |
| Value of the open of a lending product | Revolving credit: the consumer gives the credit limit, which the account gets. A loan: the consumer gives the loan amount, the account gets a credit limit of `0.00` and a posted drawdown of minus the loan amount, in one write. The split uses the CDS category only. | The CDS gives no rule. `creditLimit` is "the maximum amount of credit that is available for this account" ([CDS: BankingBalance][cds-schema-balance]), which is the limit of a card or an overdraft. A loan pays out its amount when it opens, and then the customer owes it, so `currentBalance` is negative ([CDS: BankingBalance][cds-schema-balance]). Judgement: data recipients test with the balances of loans, and a loan with a balance of 0 and the amount available is not realistic lending data. One posted transaction gives that balance with the current balance rule, and the data holder holds no new field. The CDS has no transaction type for a drawdown ([CDS: BankingTransactionV2][cds-schema-transaction]), so the type is `OTHER`. The data holder gives no `loan` or `creditCard` detail in Get Account Detail, because these fields are optional and the responses are correct without them. |
| Amount on the open page | The page accepts an amount as a person usually types it, and sends a strict AmountString. The site API and the operator interface keep the AmountString rule. | Judgement: a consumer does not type a CDS AmountString, and the API keeps one amount rule with the operator interface. The page converts only the format; the data holder checks the rules of the value, so the UI has no copy of them. |
| Business products and products for a young person | Each current product can open. No rule for the age of the consumer or for a business. | A customer is a person, and a sole trader holds a business account as a person. Eligibility needs a consumer who is 18 years or older ([CDR Rules][cdr-rules], rule 1.10B(1)(a)), but a customer has no date of birth. Judgement: each customer of the lab is an adult, and a rule with no data to check is not a rule. |
| Identity and credit checks at an open | None, and no entry in `docs/production-gaps.md` | A bank identifies a customer before it gives a designated service ([Anti-Money Laundering and Counter-Terrorism Financing Act 2006][aml-ctf-act], section 32), and a credit provider assesses a credit contract ([National Consumer Credit Protection Act 2009][nccp-act], Chapter 3). The lab has no real consumers (see Products that a consumer opens on the site). The online bank site is lab-only, and a production data holder has its own site. The project puts the security of a production online bank out of scope, and the production gaps file excludes a function that the project puts out of scope. |
| BSB of an account that the consumer opens | A required `bsb` field in `brand.json`. The presets: `wattle-bank` `372100`, `redgum-bank` `373100`, `harbour-bank` `374100`, `kestrel-bank` `375100`. The operator still gives each BSB of an account of the operator. | Judgement: a BSB is a fact of the bank, as its name is, and the accounts of a bank usually have the same few BSBs. The field changes the contract of the brand directory ([ADR 0023][adr-0023]), and ADR 0023 gives the field. A setting of the deployment puts a fact of the brand in each deployment. On 1 November 2026, the AusPayNet BSB directory had no BSB that starts with `372` to `375` ([AusPayNet: BSB directory][auspaynet-bsb]). The lab makes no payments, so a BSB that AusPayNet gives later has no effect. |
| Account number of an account that the consumer opens | 8 random digits, with at most 3 attempts in all when the pair is in use | The rule of the operator interface permits 6 to 10 digits. Judgement: the same method as the customer number. |
| Link from an account to its product | None. The open copies the category and the name of the product into the account. The display name is the product name. | No CDS account schema has a product ID ([CDS: BankingAccountV3][cds-schema-account], [CDS: Get Account Detail][cds-get-account-detail]). Judgement: customers and accounts are rows in tables, not events, and no rule or reader uses a link. An account of the operator has no product, so a link has no value for it. The account keeps its name when the product changes. |
| Confirmation of an open | The open page, with "Open account" and "Cancel". No second dialog. | An account cannot be removed, so the open cannot be reversed. A page that names the product and the effect before the submission is the "Confirmed" method ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]). |
| Second request of an open | The busy state of the button only. No idempotency key. | Judgement: a second account after two requests is what the consumer asked for, and it has no effect on a different person in the lab. The UI never sends an action again by itself ([design system DESIGN.md][ui-design]). A key adds a store for a case that the busy state already stops. |
| Close on the site | The close use case of the operator console, with the `ends` of the confirmation. The site API gives each open account its `close` action with the arrangements that end. | Judgement: one use case gives the same rules and the same end of eligibility by construction. A consent that starts between the confirmation and the close then gives 409, and does not end with no warning to the consumer. The consumer sees the name of each app, and not a general text. |
| Balance at a close | No rule for the balance or for pending transactions. The dialog shows the balance and says that the bank pays out no balance. | Neither the CDS nor the CDR Rules give a rule for the close of an account. The operator close has no such rule. Judgement: payments are out of scope, so a rule of a zero balance stops the close of an account with transactions on the site. |
| Channel of a close | No record of the channel (site or operator) and no new event | Rule 9.3 requires records of authorisations, their withdrawals and the notices, and no record of how an account closed ([CDR Rules][cdr-rules]). [ADR 0021][adr-0021]: a field needs a rule or a reader that uses it. The end of eligibility has a reason, not a method. |
| Categories of revolving credit and of a loan in the UI | The OpenAPI document of the site API gives the two lists, and a test compares each with the rule of the data holder. The product responses keep the CDS fields only. | Judgement: the product responses then stay the same as Get Products and Get Product Detail with no conversion, and the UI gets the list from one source, as for the labels of the CDS enumerations. |
| Response to an open | 201 with the new account, and no `Location` header | Judgement: the site goes to the page of the account, and it needs the `accountId` from the response. The site API has no `GET` of one account. The sign-up of the site also gives its record in the body with no `Location` header. |
| Accreditation number on the dashboard | Not shown | The DSB only recommends it ([CDR Support: Dashboards][support-dashboards]). The data holder does not read participant data from the Register (see Decisions: Name of the accredited person on the consent page). No data standard or Register entry is specified for rule 1.15(1)(bb) or rule 7.9(1)(c) in the sources of this document. |
| Ended consents on the dashboard | Shown in a group of their own, with the date of the end, and with no reason and no method. No time limit and no pages. | Rule 1.15(3)(e) requires the time of the end of an authorisation that is not current ([CDR Rules][cdr-rules]). The ACCC says that the dashboard need not show the method or the reason ([CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method], [CDR Support: Dashboards][support-dashboards]). Judgement: the accounts page shows closed accounts in the same way, and a customer has few arrangements. |
| Dashboard of a customer with no open account | The same dashboard, with no special case | The ACCC says that the obligation to give a dashboard ends with the eligibility of the consumer ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). No rule forbids a dashboard after that time. Judgement: the least complex solution, and the consumer sees the consents that ended at the close. |
| Text of the CDS messages | The suggested text of the DSB, word for word, with the brand name of the data recipient | The CDS requires the messages, and permits the data holder to use or paraphrase the suggested text ([CDS: Dashboard Standards][cds-dashboard-standards], [CDS: Withdrawal Standards][cds-withdrawal-standards]). Judgement: the text of the DSB needs no review, and a test can check it. The brand name is the name of the data recipient that the consumer knows best. |
| Confirmation of a withdrawal on the dashboard | Two steps: the action "Stop sharing", then a dialog with the CDS messages and the actions "Stop sharing" and "Cancel" | A withdrawal cannot be reversed ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]), and the CDS messages are "part of the withdrawal process" ([CDS: Withdrawal Standards][cds-withdrawal-standards]). Rule 1.15(1)(e): no more complicated than the authorisation, which has a customer number, an OTP and a confirmation ([CDR Rules][cdr-rules]). The CX Guidelines suggest positive friction and the words "Stop sharing" ([CX Guidelines: Withdrawal][cx-withdrawal]). |
| Withdrawal on the site | The use case of a withdrawal by the operator, with the customer of the session and the method `CONSUMER_DASHBOARD`. The `arrangementId` is in the path. | Judgement: one use case gives the same effects as a withdrawal by the operator by construction. A `cdr_arrangement_id` does not identify the consumer ([CDS: CDR Arrangement ID][cds-arrangement-id]), so the rule that no path of the site API contains an identifier of a consumer stays true. |
| Tests of the consumer dashboard | A test in the data holder, in one process, is the proof in CI. The browser test uses an arrangement of the conformance suite and skips when the lab has none. | Judgement: a consent needs the full consent flow of a registered software product with a client certificate, which a browser test cannot do with little code. The lab-web job in CI has no conformance data. The test in one process uses the real token verifier, the event store and the site API, as the test of the end of eligibility through the site does. |
| CDR policy of the data holder | Out of scope | Rule 7.2(8) requires that the CDR policy is readily available through each online service of the data holder ([CDR Rules][cdr-rules]), and so through the online bank site. Judgement: a CDR policy is a legal document of a real entity, and the footer says that the bank is not real. |

## Open questions

| Question | Detail |
|---|---|

[readme-sources]: ../../README.md#sources-for-requirements
[cds-get-products]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-products
[cds-get-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-product-detail
[cds-get-status]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-status
[cds-get-outages]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-outages
[cds-schema-product]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductv6
[cds-schema-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductdetailv7
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-id-permanence]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#id-permanence
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[cds-product-components]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-amp-account-components
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-uri-structure]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#uri-structure
[cds-openapi-banking]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_banking.json
[cds-openapi-common]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_common.json
[cds-http-headers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-headers
[cds-payload-conventions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#payload-conventions
[cds-pagination]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination
[cds-cors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[cds-error-400-field-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid
[cds-error-400-invalid-date-time]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid-date-time
[cds-error-404-resource-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-resource-invalid
[cds-error-404-not-found]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-resource-not-found
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-security-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#security-endpoints
[cds-oidc-config]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#openid-provider-configuration-endpoint
[cds-jwks-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#json-web-key-set-endpoint
[cds-token-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#token-endpoint
[cds-dcr-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dynamic-client-registration-endpoints
[cds-dcr-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dcr-apis
[cds-multi-brand]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#multi-brand-support-separate-issuers-for-data-holder-brands
[cds-scopes-and-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#scopes-and-claims
[cds-authorisation-scopes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-scopes
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-registration-request]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-request-using-jwt
[cds-registration-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-validation
[cds-registration-response]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-response
[cds-registration-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-errors
[cds-registration-management]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-management
[cds-put-operations]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#put-operations
[cds-schema-registration-properties]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-dynamic-client-registration-api_schemas_tocSregistrationproperties
[cds-http-response-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-response-codes
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-cert-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-openapi-dcr]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_dcr.json
[cds-consent]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consent
[cds-par-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pushed-authorisation-endpoint
[cds-authorisation-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-endpoint
[cds-introspection-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#introspection-endpoint
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
[cds-arrangement-id]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-id
[cds-sub-claim]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#sub-claim
[cds-loa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#levels-of-assurance-loas
[cds-fallback-flows]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#fallback-authentication-flows
[cds-user-identifiers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#user-identifiers
[cds-otp]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#one-time-password-credential-requirements
[cds-restricted-credentials]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#restricted-credentials
[cds-baseline-credentials]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#baseline-credential-requirements
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[cds-cx-otp]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_redirect-to-web-with-one-time-password
[cds-authorisation-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-standards
[cds-data-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-language-standards-common
[cds-banking-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#banking-and-non-bank-lending-language
[cds-customer-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#customer-language-common
[cds-accessibility]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#accessibility-standards
[cds-nfrs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#non-functional-requirements
[cds-performance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#performance-requirements
[support-white-label]: https://cdr-support.zendesk.com/hc/en-us/articles/900003938166
[support-brands]: https://cdr-support.zendesk.com/hc/en-us/articles/900005881263
[support-product-base-uri]: https://cdr-support.zendesk.com/hc/en-us/articles/16505555811855
[support-cors]: https://cdr-support.zendesk.com/hc/en-us/articles/900003054706
[support-status-fields]: https://cdr-support.zendesk.com/hc/en-us/articles/900005173683
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[support-dcr]: https://cdr-support.zendesk.com/hc/en-us/articles/6828251436431
[support-egress]: https://cdr-support.zendesk.com/hc/en-us/articles/4403543156751
[support-removed]: https://cdr-support.zendesk.com/hc/en-us/articles/900001657906
[support-otp]: https://cdr-support.zendesk.com/hc/en-us/articles/6140553082255
[support-concurrent-consent]: https://cdr-support.zendesk.com/hc/en-us/articles/4407560834959
[support-code-expiry]: https://cdr-support.zendesk.com/hc/en-us/articles/360004665496
[support-ppid]: https://cdr-support.zendesk.com/hc/en-us/articles/5505351487631
[dsb-issue-480]: https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/480
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cx-otp]: https://cx.dsb.gov.au/authenticate/redirect-with-one-time-password
[cx-authorise]: https://cx.dsb.gov.au/authorise/authorisation-to-disclose
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[oidc-discovery-validation]: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationValidation
[oidc-core-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[rfc6749-authorisation-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2.1
[rfc6750]: https://www.rfc-editor.org/rfc/rfc6750
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc5321-mailbox]: https://www.rfc-editor.org/rfc/rfc5321#section-4.1.2
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-413]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.14
[rfc9110-head]: https://www.rfc-editor.org/rfc/rfc9110#section-9.3.2
[rfc7591-metadata]: https://www.rfc-editor.org/rfc/rfc7591#section-2
[rfc7591-software-statement]: https://www.rfc-editor.org/rfc/rfc7591#section-2.3
[rfc7591-response]: https://www.rfc-editor.org/rfc/rfc7591#section-3.2.1
[rfc7591-errors]: https://www.rfc-editor.org/rfc/rfc7591#section-3.2.2
[rfc8705-binding]: https://www.rfc-editor.org/rfc/rfc8705#section-3
[cds-get-customer]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-customer
[cds-get-accounts]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-accounts
[cds-get-account-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-detail
[cds-get-account-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-balance
[cds-get-bulk-balances]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-bulk-balances
[cds-get-balances-specific]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-balances-for-specific-accounts
[cds-get-transactions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account
[cds-get-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transaction-detail
[cds-schema-person]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_schemas_tocScommonperson
[cds-schema-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingaccountv3
[cds-schema-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingbalance
[rfc9562-v8]: https://www.rfc-editor.org/rfc/rfc9562#section-5.8
[cds-schema-transaction]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactionv2
[cds-schema-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactiondetailv3
[cds-userinfo-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#userinfo-endpoint
[cds-future-dated]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#future-dated-obligations
[cds-data-latency]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-latency
[cds-error-404-not-implemented]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-resource-not-implemented
[cds-error-404-invalid-banking-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-authorisation-invalid-banking-account
[cds-error-422-invalid-banking-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-422-authorisation-invalid-banking-account
[cds-error-400-header-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-header-invalid
[cds-error-400-field-missing]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-missing
[cds-error-403-invalid-consent]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-invalid-consent
[support-auth-date]: https://cdr-support.zendesk.com/hc/en-us/articles/4405773134863
[support-client-headers]: https://cdr-support.zendesk.com/hc/en-us/articles/4404953530767
[support-fapi-headers]: https://cdr-support.zendesk.com/hc/en-us/articles/4410664891919
[support-closed-accounts]: https://cdr-support.zendesk.com/hc/en-us/articles/14519397304079
[nccp-act]: https://www.legislation.gov.au/C2009A00134/latest/text
[aml-ctf-act]: https://www.legislation.gov.au/C2006A00169/latest/text
[auspaynet-bsb]: https://bsb.auspaynet.com.au/
[fapi-baseline-resource]: https://openid.net/specs/openid-financial-api-part-1-1_0.html#protected-resources-provisions
[oidc-core-userinfo]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
[oidc-core-userinfo-response]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse
[rfc6750-header]: https://www.rfc-editor.org/rfc/rfc6750#section-2.1
[cds-token-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#token-revocation-endpoint
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-error-403-revoked-consent]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-revoked-consent
[cds-get-sp-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-products-statuses
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-self-signed-jwt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#self-signed-jwt-client-authentication
[cds-dh-calling-dr]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holders-calling-data-recipients
[cds-withdrawal-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#withdrawal-standards
[support-revocation-guidance]: https://cdr-support.zendesk.com/hc/en-us/articles/6828747049359
[support-backoff]: https://cdr-support.zendesk.com/hc/en-us/articles/360004056455
[adr-0003]: ../../docs/adr/0003-event-sourcing.md
[adr-0021]: ../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0006]: ../../docs/adr/0006-telemetry.md
[conformance-design]: ../../conformance/DESIGN.md
[adr-0012]: ../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0017]: ../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0019]: ../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[ui-design]: ../../web/packages/ui/DESIGN.md
[register-console]: ../register/DESIGN.md#operator-console
[register-decisions]: ../register/DESIGN.md#decisions
[register-console-listener]: ../register/DESIGN.md#listener-and-security
[register-console-api]: ../register/DESIGN.md#console-api
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[support-copies-of-records]: https://cdr-support.zendesk.com/hc/en-us/articles/900004438886
[support-withdrawal-method]: https://cdr-support.zendesk.com/hc/en-us/articles/4411292647823
[support-withdrawal-notification]: https://cdr-support.zendesk.com/hc/en-us/articles/900001658546
[support-dashboards]: https://cdr-support.zendesk.com/hc/en-us/articles/6140622825231
[support-dashboard-clusters]: https://cdr-support.zendesk.com/hc/en-us/articles/4920717853327
[support-dashboard-dates]: https://cdr-support.zendesk.com/hc/en-us/articles/5333587630095
[cds-dashboard-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_dashboard-standards_data-holder-dashboards
[cx-withdrawal]: https://cx.dsb.gov.au/consent-management/consent-management-data-holder/withdrawal/cx-guidelines
[adr-0016]: ../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[adr-0018]: ../../docs/adr/0018-container-images-with-a-web-ui-build-stage.md
[adr-0023]: ../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md
[adr-0002]: ../../docs/adr/0002-data-store.md
[adr-0020]: ../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[compose-profiles]: https://docs.docker.com/reference/compose-file/profiles/
[compose-down]: https://docs.docker.com/reference/cli/docker/compose/down/
[postgres-image-init]: https://hub.docker.com/_/postgres
[k8s-configmap]: https://kubernetes.io/docs/concepts/configuration/configmap/
[rfc9649]: https://www.rfc-editor.org/rfc/rfc9649
[k8s-atomic-writer]: https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/util/atomic_writer.go
[ui-brands]: ../../web/packages/ui/DESIGN.md#brands
[ui-consumer-pages]: ../../web/packages/ui/DESIGN.md#consumer-pages
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-4-3]: https://www.w3.org/TR/WCAG22/#contrast-minimum
[wcag-1-4-10]: https://www.w3.org/TR/WCAG22/#reflow
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
[rfc2606-tlds]: https://www.rfc-editor.org/rfc/rfc2606#section-2
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[mdn-csp-sandbox]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/sandbox
[mdn-csp]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy
[mdn-csp-form-action]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/form-action
[rfc5322-2-1-1]: https://www.rfc-editor.org/rfc/rfc5322#section-2.1.1
[rfc5322-2-2]: https://www.rfc-editor.org/rfc/rfc5322#section-2.2
[rfc5322-3-4]: https://www.rfc-editor.org/rfc/rfc5322#section-3.4
[rfc2047-5]: https://www.rfc-editor.org/rfc/rfc2047#section-5
[rfc2046-5-1-4]: https://www.rfc-editor.org/rfc/rfc2046#section-5.1.4
[rfc2387]: https://www.rfc-editor.org/rfc/rfc2387
[rfc2392]: https://www.rfc-editor.org/rfc/rfc2392
[gmail-images]: https://support.google.com/mail/answer/145919
[caniemail-png]: https://www.caniemail.com/features/image-png/
[caniemail-svg]: https://www.caniemail.com/features/image-svg/
[caniemail-style]: https://www.caniemail.com/features/html-style/
[caniemail-css-variables]: https://www.caniemail.com/features/css-variables/
[caniemail-modern-colour]: https://www.caniemail.com/features/css-modern-color/
[mailpit-part-handler]: https://github.com/axllent/mailpit/blob/v1.32.0/server/apiv1/message.go
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-3-1-1]: https://www.w3.org/TR/WCAG22/#language-of-page
[wcag-3-3-8]: https://www.w3.org/TR/WCAG22/#accessible-authentication-minimum
[ui-content]: ../../web/packages/ui/DESIGN.md#content
[ui-accessibility]: ../../web/packages/ui/DESIGN.md#accessibility
[ui-layout]: ../../web/packages/ui/DESIGN.md#layout
[rfc6797-b]: https://www.rfc-editor.org/rfc/rfc6797#appendix-B
[rfc6265-8-5]: https://www.rfc-editor.org/rfc/rfc6265#section-8.5
[rfc6265bis-host]: https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-rfc6265bis#section-4.1.3.2
[rfc9110-401]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.2
[rfc9110-403]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.4
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.10
[rfc10017-cookie-session]: https://www.rfc-editor.org/rfc/rfc10017#section-7.1
[asvs-v7]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x16-V7-Session-Management.md
[owasp-session]: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
[owasp-csrf]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
[go-crossoriginprotection]: https://pkg.go.dev/net/http#CrossOriginProtection
[mdn-clear-site-data]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Clear-Site-Data
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[wcag-1-3-5]: https://www.w3.org/TR/WCAG22/#identify-input-purpose
[readme-conventions]: ../../README.md#other-conventions
[register-chart]: ../../charts/register/README.md
[gha-needs]: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idneeds
[gha-status-checks]: https://docs.github.com/en/actions/reference/workflows-and-actions/expressions#status-check-functions
[helm-required]: https://helm.sh/docs/howto/charts_tips_and_tricks/#using-the-required-function
[helm-files]: https://helm.sh/docs/chart_template_guide/accessing_files/
[k8s-configmap-volume]: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
[k8s-init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-apis
[register-deployment]: ../register/DESIGN.md#deployment-from-published-artifacts
