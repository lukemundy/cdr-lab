# Mock data holder: Operator console

This area gives the lab-only operator console: its functions, its presentation, its listener and security. The [actions and create forms](operator-console-create.md) of the console are a separate area. The [console API](operator-console-api.md) gives the data and the actions of the console.

## Scope

- <a id="dh-1400"></a>**DH-1400** The operator console is a web UI for the lab operator. It shows the customers, their accounts, their transactions and their arrangements, the notices to the data recipients, the registrations and the participant statuses of the software products. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1401"></a>**DH-1401** The operator console is a lab-only feature. A real component does not depend on it. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1402"></a>**DH-1402** The operator console is not a consumer dashboard. It does not meet rule 1.15, because the consumer dashboard is an online service for the consumer. The [consumer dashboard](consumer-dashboard.md#dh-1900) is on the online bank site. ([CDR Rules][cdr-rules], rule 1.15(1))

## Overview

- <a id="dh-1403"></a>**DH-1403** The overview is the first page of the console. The path `/` is the overview. ([Judgement: the content of the overview](#dh-1500))
- <a id="dh-1404"></a>**DH-1404** The overview gives these counts. ([Judgement: the content of the overview](#dh-1500))
  - the customers
  - the open accounts, and the closed accounts
  - the current arrangements, and the ended arrangements
  - the registrations
  - the registered software products whose participant status is not `ACTIVE`
- <a id="dh-1405"></a>**DH-1405** Each count of the overview links to its list, when the console has one. For example, the count of the customers links to the list of the customers. ([Judgement: the content of the overview](#dh-1500))
- <a id="dh-1406"></a>**DH-1406** The overview gives the 10 newest changes, newest first. ([Judgement: the content of the overview](#dh-1500))
- <a id="dh-1407"></a>**DH-1407** A change of the overview is one of these kinds. Transactions and notices are not changes of the overview. ([Judgement: the changes in the overview](#dh-1501))
  - the creation of a customer
  - the creation of an account, and the close of an account
  - the confirmation of an arrangement, and the end of an arrangement
  - the registration, the change and the deletion of a registration
- <a id="dh-1408"></a>**DH-1408** Each change of the overview gives its time and its kind. ([Judgement: the changes in the overview](#dh-1501))
- <a id="dh-1409"></a>**DH-1409** A change of a customer, an account or an arrangement gives the customer number. The customer number links to the detail of the customer. ([Judgement: the content of the overview](#dh-1500))
- <a id="dh-1410"></a>**DH-1410** A change of an arrangement or a registration gives the `client_name` of the registration. When the registration exists, the `client_name` links to it. ([Judgement: the link from an arrangement to its registration](#dh-1508))
- <a id="dh-1411"></a>**DH-1411** When the kind of a change has a detail, the change gives it. ([Judgement: the changes in the overview](#dh-1501))
  - the product name, for an account
  - the end of the sharing, for the confirmation of an arrangement
  - the reason of the end, for the end of an arrangement, and the [method of a withdrawal](end-of-an-arrangement.md#dh-0917)
- <a id="dh-1412"></a>**DH-1412** The overview does not show the name or the email address of a customer, the display name of an account or an account number. ([Judgement: the content of the overview](#dh-1500))

## Customers and accounts

- <a id="dh-1413"></a>**DH-1413** The list of the customers gives, for each customer, the customer number, the name, the email address, the number of open accounts and the number of arrangements that did not end. ([Judgement: the records of one customer](#dh-1503))
- <a id="dh-1414"></a>**DH-1414** An arrangement that did not end has no event that ended it, and the end of its sharing is after the time of the request. For example, an arrangement whose sharing ended one minute ago with no event is not counted. ([the end of an arrangement](end-of-an-arrangement.md#dh-0900); [Judgement: the count of the arrangements that did not end](#dh-1502))
- <a id="dh-1415"></a>**DH-1415** The lab operator can search the list of the customers by a part of the customer number, the name or the email address. For example, "Smi" finds the customer "Jane Smith". ([CDR Rules][cdr-rules], rule 9.5; [Judgement: the records of one customer](#dh-1503))
- <a id="dh-1416"></a>**DH-1416** The detail of a customer gives the customer number, the first name, the last name, the email address and the time of the last change. ([CDR Rules][cdr-rules], rule 9.5; [Judgement: the records of one customer](#dh-1503))
- <a id="dh-1417"></a>**DH-1417** The detail of a customer gives each account and each arrangement of the customer. ([CDR Rules][cdr-rules], rule 9.5; [Judgement: the records of one customer](#dh-1503))
- <a id="dh-1418"></a>**DH-1418** The detail of a customer gives these fields of each account. ([the fields of an account](operator-interface.md#dh-0722); [Judgement: no mask of the account number for the lab operator](operator-interface.md#dh-0774))
  - the display name, the product category and the product name
  - the BSB, and the account number in full
  - the credit limit, the current balance and the available balance
  - the status and the date of creation
- <a id="dh-1419"></a>**DH-1419** The date of creation of an account is its `creationDate`, the date in UTC. ([CDS: BankingAccountV3][cds-schema-account]; [Judgement: the date of creation of an account](#dh-1504))
- <a id="dh-1420"></a>**DH-1420** The transactions of an account show newest first, with the fields that the [operator interface](operator-interface.md#dh-0724) gives. ([Judgement: the columns of the transactions](#dh-1506))
- <a id="dh-1421"></a>**DH-1421** The transactions of an account include each transaction of the account, with no time limit. ([Judgement: all the transactions for the lab operator](#dh-1505))
- <a id="dh-1422"></a>**DH-1422** The time, the description, the type, the status and the amount of a transaction are columns. A details column gives each other field that has a value. ([Judgement: the columns of the transactions](#dh-1506))
- <a id="dh-1423"></a>**DH-1423** Each row of a list links to the detail of its record. For example, a row of the list of the customers links to the detail of the customer. ([Judgement: the records of one customer](#dh-1503))

## Arrangements

- <a id="dh-1424"></a>**DH-1424** The detail of a customer gives these fields of each arrangement. ([CDR Rules][cdr-rules], rule 9.3(1)(a) to (c))
  - the `cdr_arrangement_id`
  - the software product: its software product ID and its `clientName`
  - the scopes and the accounts
  - the time of the confirmation and the end of the sharing
  - for an arrangement that an event ended: the time and the reason of the end
  - for a withdrawal: the [method of the withdrawal](end-of-an-arrangement.md#dh-0917)
- <a id="dh-1425"></a>**DH-1425** When an arrangement has no account, its accounts show as "None". For example, an arrangement with no `bank:` scope has no account selection. ([the account selection on the authorisation page](authorisation.md#dh-0819))
- <a id="dh-1426"></a>**DH-1426** For an arrangement with a notice, the detail gives the status of the notice, and each attempt with its time and its result. ([CDR Rules][cdr-rules], rule 4.26A and rule 9.3(1); [the record of a notice](end-of-an-arrangement.md#dh-0939); [the record of an attempt](end-of-an-arrangement.md#dh-0940))
- <a id="dh-1427"></a>**DH-1427** When a registration has the client ID of an arrangement, the arrangement links to that registration. Otherwise, the arrangement has no link. ([Judgement: the link from an arrangement to its registration](#dh-1508))

## Registrations

- <a id="dh-1428"></a>**DH-1428** The list of the registrations gives, for each registration, these values. ([Judgement: the counts of the arrangements of a registration](#dh-1509))
  - the client ID, the software product ID and the `client_name`
  - the time of the registration and the time of the last change
  - the participant status of the software product
  - the number of arrangements of the client that did not end, and the number of all the arrangements of the client
- <a id="dh-1429"></a>**DH-1429** The time of the registration is the time of the last registration of the software product. ([Judgement: the times of a registration](#dh-1510))
- <a id="dh-1430"></a>**DH-1430** The time of the last change is the time of the last `PUT` of the registration. Before a `PUT`, it is the time of the registration. ([Judgement: the times of a registration](#dh-1510))
- <a id="dh-1431"></a>**DH-1431** The detail of a registration gives its metadata, with each redirect URI, the `recipient_base_uri`, the `jwks_uri` and the scopes. ([CDS: Dynamic Client Registration][cds-dcr-endpoints])
- <a id="dh-1432"></a>**DH-1432** The console shows each URI of a registration as text with a copy button, not as a link. It loads no file from the URI. ([OWASP ASVS 5.0, V3.6.1][asvs-v3])
- <a id="dh-1433"></a>**DH-1433** The detail of a registration gives each arrangement of its client ID, also an arrangement that ended. The arrangements are in a table with pages, newest first. ([Judgement: the arrangements of a registration in pages](#dh-1511))
- <a id="dh-1434"></a>**DH-1434** Each row of the arrangements of a registration gives the `cdr_arrangement_id`, the name and the customer number of the customer, the status, the time of the confirmation and the end of the sharing. It gives no email address. ([Judgement: the customer in the arrangements of a registration](#dh-1512))
- <a id="dh-1435"></a>**DH-1435** The name of the customer in the arrangements of a registration links to the detail of the customer. ([Judgement: the customer in the arrangements of a registration](#dh-1512))

## Participant statuses

- <a id="dh-1436"></a>**DH-1436** The participant statuses page gives each software product that the data holder has a status for, with its name and its status. ([CDS: Metadata Cache Management][cds-metadata-cache]; [the statuses from the Register](registration.md#dh-0679))
- <a id="dh-1437"></a>**DH-1437** The name of a software product is the `client_name` of its registration. A software product with no registration shows "Not registered". ([Judgement: the name of a software product](#dh-1513))
- <a id="dh-1438"></a>**DH-1438** The page gives the time of the last response of Get Software Products Statuses. A response with no change (304) also sets the time, and a request that fails does not change it. ([CDS: Metadata Cache Management][cds-metadata-cache]; [Judgement: the time of the last response of the Register](#dh-1514))
- <a id="dh-1439"></a>**DH-1439** Before the data holder gets the statuses one time, the page gives no time. ([Judgement: the time of the last response of the Register](#dh-1514))
- <a id="dh-1440"></a>**DH-1440** The lab operator can filter the list by the registration: all, registered or not registered. ([Judgement: the name of a software product](#dh-1513))

## Views that are not in the console

- <a id="dh-1441"></a>**DH-1441** The console does not show the one-time passwords, the signing keys, the tokens or the authorisation codes. ([Judgement: the views that are not in the console](#dh-1515))
- <a id="dh-1442"></a>**DH-1442** The console has no view of the products and no history of events. ([Judgement: the views that are not in the console](#dh-1515))

## Presentation

- <a id="dh-1465"></a>**DH-1465** The console is for desktop browsers, with the minimum width of a console of the design system. ([the minimum width of a console](../../../web/packages/ui/design/layout.md#ui-0070); [Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1466"></a>**DH-1466** The header of the console shows the logo of the brand on the brand colour, and the label "Operator console". The text alternative of the logo is the brand name. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]; [Judgement: the brand in the console](#dh-1523))
- <a id="dh-1467"></a>**DH-1467** The console gets the brand name from the console API. Until the response, and when the request fails, the header and the tab title say "Data holder". ([one source of the brand name](brand.md#dh-0305); [Judgement: the source of the brand name in the UI](#dh-1524))
- <a id="dh-1468"></a>**DH-1468** The console has the shell of a console of the design system, with the roles of the brand that the console of a participant uses. ([the roles of the brand in the console of a participant](../../../web/packages/ui/design/tokens-and-brands.md#ui-0035); [Judgement: the brand in the console](#dh-1523))
- <a id="dh-1469"></a>**DH-1469** The rest of the console uses the lab brand, with its fonts and its radius, and the compact density. ([the lab brand for the other roles](../../../web/packages/ui/design/tokens-and-brands.md#ui-0034); [the fonts and the radius of the lab brand](../../../web/packages/ui/design/tokens-and-brands.md#ui-0036); [the compact density of a console](../../../web/packages/ui/design/tokens-and-brands.md#ui-0028); [Judgement: the brand in the console](#dh-1523))
- <a id="dh-1470"></a>**DH-1470** The side navigation of each preset brand is not white. With the logo hidden, the colours of the shell tell the console of each preset brand from the console of each other preset brand and from the Register console. ([Judgement: consoles that differ in one look](#dh-1525))
- <a id="dh-1471"></a>**DH-1471** "Overview" is the first item of the side navigation. ([Judgement: the content of the overview](#dh-1500))
- <a id="dh-1472"></a>**DH-1472** The title of the browser tab is the brand name. The icon of the tab is [the mark of the brand](brand.md#dh-0326). ([Judgement: the brand in the console](#dh-1523))
- <a id="dh-1473"></a>**DH-1473** The search, the sort and the page of a list are in the URL. For example, a reload of a list shows the same search, sort and page. ([Judgement: the same action and form rules in each console](operator-console-create.md#dh-1517))
- <a id="dh-1474"></a>**DH-1474** A search replaces the history entry of the list. A sort and a page add a history entry. ([Judgement: the same action and form rules in each console](operator-console-create.md#dh-1517))
- <a id="dh-1475"></a>**DH-1475** The detail of a customer, the transactions of an account and the detail of a registration each have their own URL. The page of the arrangements of a registration is in the URL of the detail of the registration. ([Judgement: the same action and form rules in each console](operator-console-create.md#dh-1517))
- <a id="dh-1476"></a>**DH-1476** The console shows each time in the time zone of the browser, as the design system gives it. ([Judgement: UTC for all times (ADR 0012)][adr-0012]; [the time zone and the offset of a time](../../../web/packages/ui/design/content.md#ui-0147))
- <a id="dh-1477"></a>**DH-1477** An amount shows in the format of the design system. A pending transaction shows as pending. ([the format of an amount](../../../web/packages/ui/design/content.md#ui-0157); [no accounting format in a console](../../../web/packages/ui/design/content.md#ui-0159); [CDS: BankingTransactionDetailV3][cds-schema-transaction-detail])
- <a id="dh-1478"></a>**DH-1478** A BSB shows as two groups of three digits with a hyphen. For example, `062-000`. Its copy button copies the 6 digits. ([Judgement: the format of a BSB](#dh-1526))
- <a id="dh-1479"></a>**DH-1479** The console has the accessibility level of an operator console. ([the accessibility level of an operator console](../../../web/packages/ui/design/accessibility.md#ui-0111))

## Listener and security

The console is a web UI. The [web UI library](../../../pkg/platform/consolehttp/DESIGN.md) gives the rules of its listener, its browser security, its cache headers and its routes.

- <a id="dh-1480"></a>**DH-1480** The console is on its own listener, separate from the CDR APIs and from the operator interface. ([the web UI and its API on one listener](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0001); [Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1481"></a>**DH-1481** The setting `CONSOLE_ENABLED` [turns the console on, and the console is off by default](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0002). ([Judgement: the web UI is off by default](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0018))
- <a id="dh-1482"></a>**DH-1482** The setting `CONSOLE_ALLOWED_HOSTS` is [the `Host` allowlist of the console](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0004). Its values are [the values of the `Host` header](../../../pkg/platform/hostallowlist/DESIGN.md#host-0001), with [no difference for upper and lower case](../../../pkg/platform/hostallowlist/DESIGN.md#host-0002). For example, `127.0.0.1:8184`. ([Judgement: a Host allowlist with no default](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0019))
- <a id="dh-1483"></a>**DH-1483** The console listener is plain HTTP. ([Judgement: plain HTTP for the console](#dh-1527))
- <a id="dh-1484"></a>**DH-1484** The console has no sign-in, and the browser gets no credential. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1485"></a>**DH-1485** The console listener serves the logo at `/brand/logo.svg`, the mark at `/brand/mark.svg` and a stylesheet at `/brand/console.css`. They have [the `Host` allowlist](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0004) and [the response headers](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007) of the console. ([Judgement: the brand files on the console listener](#dh-1528))
- <a id="dh-1486"></a>**DH-1486** The stylesheet of the console sets only the roles of the brand that the console uses. ([Judgement: the brand files on the console listener](#dh-1528))
- <a id="dh-1487"></a>**DH-1487** The data holder starts the console listener only in the process that runs the APIs. A process that runs only the workers needs no console setting. ([a process needs only the settings that it uses](non-functional-requirements.md#dh-0010))

## Non-functional requirements

- <a id="dh-1488"></a>**DH-1488** Telemetry: the console API has the same server telemetry as the other listeners, and [the telemetry rules of the data holder](non-functional-requirements.md#dh-0008) apply to it. ([Judgement: telemetry with OpenTelemetry (ADR 0006)][adr-0006])
- <a id="dh-1489"></a>**DH-1489** Telemetry: no log of the console contains the query string of a request, and [no span contains it](../../../pkg/platform/telemetry/DESIGN.md#tele-0007). For example, a search for a customer name puts no name in telemetry. ([no personal information in telemetry](non-functional-requirements.md#dh-0008); [Judgement: the records of one customer](#dh-1503))
- <a id="dh-1490"></a>**DH-1490** Telemetry: the UI sends no telemetry. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1491"></a>**DH-1491** Performance: on a development workstation, a search of 1,000 customers shows within 1 second. ([Judgement: the performance limits of the console](#dh-1529))
- <a id="dh-1492"></a>**DH-1492** Performance: on a development workstation, the first page of 10,000 transactions of an account shows within 1 second. ([Judgement: the performance limits of the console](#dh-1529))

## Judgements

### <a id="dh-1500"></a>DH-1500 The content of the overview

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: A screenshot of the overview shows no personal information. The customer number is the identifier that the lab operator needs to find the customer, and it identifies a person only with access to the console. A display name is text that a consumer can choose.

### <a id="dh-1501"></a>DH-1501 The changes in the overview

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: A lab can have thousands of transactions, which push out all other changes. The transactions of an account and the notice of an arrangement show on their own pages.

### <a id="dh-1502"></a>DH-1502 The count of the arrangements that did not end

- Kind: choice
- Basis: Step 3.3. An arrangement ends at the end of its sharing also with no event ([the end of an arrangement](end-of-an-arrangement.md#dh-0900)).
- Reason: The count uses the same rule as the end.

### <a id="dh-1503"></a>DH-1503 The records of one customer

- Kind: choice
- Basis: Step 3.1. A consumer can ask the data holder for copies of their records "in any form" ([CDR Rules][cdr-rules], rule 9.5; [CDR Support: Rule 9.5 – Requests from CDR consumers for copies of records][support-copies-of-records]).
- Reason: The lab operator then finds the records of one customer, by a value that the consumer gives: the customer number, the name or the email address. The query string of a search then has personal information, so the telemetry does not record it.

### <a id="dh-1504"></a>DH-1504 The date of creation of an account

- Kind: choice
- Basis: Step 3.3. The banking APIs give the `creationDate` as a date ([CDS: BankingAccountV3][cds-schema-account]).
- Reason: The lab operator sees the value that the data recipient gets.

### <a id="dh-1505"></a>DH-1505 All the transactions for the lab operator

- Kind: choice
- Basis: Step 3.3. The banking APIs give a data recipient only 2 years of transactions ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(6)(a)).
- Reason: The 2-year limit is for the data recipient. The lab operator acts for the brand, and the brand has all the records of the customer.

### <a id="dh-1506"></a>DH-1506 The columns of the transactions

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The table stays readable at 1280 CSS px. It still shows each field that the operator interface gives.

### <a id="dh-1508"></a>DH-1508 The link from an arrangement to its registration

- Kind: choice
- Basis: Step 3.3. A new registration of a software product after a `DELETE` has a different client ID ([no client ID again](registration.md#dh-0634)).
- Reason: The registration with the client ID of the arrangement is the one that the arrangement uses. A link to a new registration shows values that the arrangement did not use.

### <a id="dh-1509"></a>DH-1509 The counts of the arrangements of a registration

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The arrangements that did not end are the current use of the software product. All the arrangements show that a client had arrangements that ended.

### <a id="dh-1510"></a>DH-1510 The times of a registration

- Kind: choice
- Basis: Step 3.3. The registration metadata has no time of the registration ([CDS: Dynamic Client Registration][cds-dcr-endpoints]).
- Reason: The events of the registration give the times. A registration after a `DELETE` has a new client ID, so it is a new registration with its own time.

### <a id="dh-1511"></a>DH-1511 The arrangements of a registration in pages

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: A software product can have many arrangements, and one card for each arrangement does not scale. The lab operator looks for the newest use of a software product first.

### <a id="dh-1512"></a>DH-1512 The customer in the arrangements of a registration

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The name lets the lab operator find the consumer, and the customer number is the identifier of the link. The detail of the customer gives the email address. One more column makes the table wider than 1280 CSS px.

### <a id="dh-1513"></a>DH-1513 The name of a software product

- Kind: choice
- Basis: Step 3.3. The registration metadata gives `client_name` as the name of the software product ([CDS: Dynamic Client Registration][cds-dcr-endpoints]), and the data holder has one registration for each software product.
- Reason: The data holder does not read participant data from the Register. Get Data Recipients adds a second Register client only for the names.

### <a id="dh-1514"></a>DH-1514 The time of the last response of the Register

- Kind: choice
- Basis: Step 3.3. A 304 response tells the data holder that its statuses are current ([CDS: Metadata Cache Management][cds-metadata-cache]).
- Reason: The lab operator sees that the data holder still gets the statuses. An old time shows that the data holder cannot get the statuses.

### <a id="dh-1515"></a>DH-1515 The views that are not in the console

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: A one-time password lets the lab operator authenticate as the consumer, and the mail sink and the telemetry backend have their own UIs. Get Products gives the products, and the data holder has one signing key. Only the arrangements and the registrations have events, and the detail of an arrangement gives its end and its notice.

### <a id="dh-1523"></a>DH-1523 The brand in the console

- Kind: choice
- Basis: Step 3.3. The design system gives the roles of the brand in the console of a participant ([the roles of the brand in the console of a participant](../../../web/packages/ui/design/tokens-and-brands.md#ui-0035)).
- Reason: The console shows the brand and stays a dense tool for the lab operator. A different font changes the width of each table column. The label "Operator console" stops the lab operator from mistaking the console for a consumer page.

### <a id="dh-1524"></a>DH-1524 The source of the brand name in the UI

- Kind: choice
- Basis: Step 3.3. The console API is the only source of data for the UI ([ADR 0014][adr-0014]).
- Reason: A change of the HTML page at runtime adds a second contract. "Data holder" is true for each brand.

### <a id="dh-1525"></a>DH-1525 Consoles that differ in one look

- Kind: choice
- Basis: Step 3.3. The brand colour is white for three preset brands, so the side navigation uses the hero colour ([the side navigation on the hero colour](../../../web/packages/ui/design/layout.md#ui-0076)).
- Reason: A lab operator who has more than one console open finds the tab of a brand in one look.

### <a id="dh-1526"></a>DH-1526 The format of a BSB

- Kind: choice
- Basis: Step 3.3. The CDS gives a BSB as 6 digits ([CDS: BankingAccountV3][cds-schema-account]).
- Reason: The two groups are the usual format of a BSB. The create form, the operator interface and the banking APIs use the 6 digits, so the copy gives a value that each of them accepts.

### <a id="dh-1527"></a>DH-1527 Plain HTTP for the console

- Kind: choice
- Basis: Step 3.3. All lab ports bind to `127.0.0.1`, and the console sets no cookie ([ADR 0014][adr-0014]).
- Reason: The operator interface is also plain HTTP. With no cookie, the console needs no `Secure` attribute.

### <a id="dh-1528"></a>DH-1528 The brand files on the console listener

- Kind: choice
- Basis: Step 3.3. The Content Security Policy of the console permits images and styles from the same origin only ([a strict CSP](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0020)).
- Reason: The brand files of the public listener are on a different origin. The console has the same paths for the same files, and a different name for a stylesheet with different content.

### <a id="dh-1529"></a>DH-1529 The performance limits of the console

- Kind: choice
- Basis: Step 3.3. The CDS performance requirements are for the CDR APIs, not for a tool of the lab operator.
- Reason: The limits are the same as for the Register console. They find a page that loads all records at once.

[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[support-copies-of-records]: https://cdr-support.zendesk.com/hc/en-us/articles/900004438886
[cds-schema-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingaccountv3
[cds-schema-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactiondetailv3
[cds-dcr-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dynamic-client-registration-endpoints
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[adr-0006]: ../../../docs/adr/0006-telemetry.md
[adr-0012]: ../../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
