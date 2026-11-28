# Mock data recipient: Accounts and transactions

This area gives the pages of the web app that show the collected data: the connections of the member on the home page, the accounts and their balances, and the transactions of an account. Each page shows only the data that the data recipient collected and keeps. No page calls a bank.

## Home page

- <a id="dr-1100"></a>**DR-1100** When the member has no connection, the home page says that the member has no connected bank, and has [the action to connect a bank](consent.md#dr-0800). ([Judgement: a landing page that says what the app does](web-app.md#dr-0252))
- <a id="dr-1101"></a>**DR-1101** When the member has connections, the home page shows each connection with the logo and the brand name of its bank, in the sequence of the time of the connection, the newest first. ([Judgement: the collected data in the web app only](#dr-1150))
- <a id="dr-1102"></a>**DR-1102** For each connection, the home page shows the time of the last successful collection run and the date on which the consent expires. ([CDR Rules][cdr-rules], rule 4.14(1)(d); [a time in the time zone of the browser](../../../web/packages/ui/design/content.md#ui-0147))
- <a id="dr-1103"></a>**DR-1103** While the first run of a connection is in progress, the connection shows that the brand is getting the data. ([the loading state](../../../web/packages/ui/design/components.md#ui-0275))
- <a id="dr-1104"></a>**DR-1104** For each connection, the home page lists each collected account with its display name, the label of its product category, its current balance and its available balance. ([CDS: BankingAccountV3][cds-schema-account]; [CDS: BankingBalance][cds-schema-balance]; [the data cluster of the consent](consent.md#dr-0807))
- <a id="dr-1105"></a>**DR-1105** The home page does not show a connection whose consent expired or was replaced. ([the deletion at the end of a consent](consent.md#dr-0835))
- <a id="dr-1106"></a>**DR-1106** Each account of the home page is a link to the page of that account. ([Judgement: the collected data in the web app only](#dr-1150))

## Page of an account

- <a id="dr-1107"></a>**DR-1107** The page of an account shows the fields of the home page for that account, the brand name of its bank, and the action to get new data of [its connection](data-collection.md#dr-1004). ([Judgement: the collected data in the web app only](#dr-1150))
- <a id="dr-1108"></a>**DR-1108** When the consent of the connection has `bank:transactions:read`, the page of the account lists its collected transactions. When it does not have the scope, the page says that the member did not share the transactions of this bank. ([the transaction cluster is optional](consent.md#dr-0808))
- <a id="dr-1109"></a>**DR-1109** The transactions are in the sequence of their effective time, the newest first. The effective time is the `postingDateTime`, or the `executionDateTime` when the transaction has no `postingDateTime`. ([the effective time of a transaction](../../dataholder/design/banking-apis.md#dh-1082); [CDS: BankingTransactionV2][cds-schema-transaction]; [Judgement: the sequence of the data holder](#dr-1151))
- <a id="dr-1110"></a>**DR-1110** For each transaction, the page shows the date of its effective time, its description and its amount with its sign. A pending transaction also shows "Pending". ([CDS: BankingTransactionV2][cds-schema-transaction]; [the transactions on the online bank site](../../dataholder/design/online-bank-site-accounts.md#dh-2206))
- <a id="dr-1111"></a>**DR-1111** The page shows 25 transactions at a time, with [pagination](../../../web/packages/ui/design/components.md#ui-0266) below the list. ([CDS: Pagination][cds-pagination]; [Judgement: the sequence of the data holder](#dr-1151))
- <a id="dr-1112"></a>**DR-1112** When the account has no collected transaction, the page says that the bank gave no transactions for the period of the collection. ([the empty state](../../../web/packages/ui/design/components.md#ui-0275))
- <a id="dr-1113"></a>**DR-1113** Each amount has the format of a consumer UI. For example, "-$12.00". ([the form of an amount](../../../web/packages/ui/design/content.md#ui-0157); [the sign of a negative amount](../../../web/packages/ui/design/content.md#ui-0158))
- <a id="dr-1114"></a>**DR-1114** The page of an account of a different member, or of an account that does not exist, is the "not found" page of the web app. ([OWASP ASVS 5.0, 8.2.2][asvs-v8]; [RFC 9110, section 15.5.5][rfc9110-404])

## API

- <a id="dr-1115"></a>**DR-1115** The API identifies an account with an ID that the data recipient makes, and not with the account ID of the bank. ([Judgement: an account ID of the data recipient](#dr-1152))
- <a id="dr-1116"></a>**DR-1116** The API gives a page of transactions with the parameters `page` and `page-size`. The page size is 25 by default, and 100 at most. ([CDS: Pagination][cds-pagination]; [Judgement: the sequence of the data holder](#dr-1151))
- <a id="dr-1117"></a>**DR-1117** Each response of the API with collected data has `Cache-Control: no-store`. ([the cache headers of a web UI](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0008); [OWASP ASVS 5.0, 14.3.2][asvs-v14])

## Judgements

### <a id="dr-1150"></a>DR-1150 The collected data in the web app only

- Kind: choice
- Basis: Step 3.3. The CDS gives no page of a data recipient that shows CDR data. The use of the consent is to show the data to the member ([the use of the data](consent.md#dr-0810)).
- Reason: A home page with each connection and its accounts is the view of a personal finance app. A page that shows only the stored data has no call to a bank, so it obeys the traffic rules of the collection.

### <a id="dr-1151"></a>DR-1151 The sequence of the data holder

- Kind: choice
- Basis: Step 3.3. Get Transactions For Account gives the transactions in the order of their effective time, the newest first ([the order of the data holder](../../dataholder/design/banking-apis.md#dh-1083)), and the CDS gives a default page size of 25 ([CDS: Pagination][cds-pagination]).
- Reason: The same sequence and page size as the bank let an engineer compare a page of the web app with a response of the bank.

### <a id="dr-1152"></a>DR-1152 An account ID of the data recipient

- Kind: choice
- Basis: Step 3.2. The account ID of a bank is CDR data, and the paths of the web app are in the spans of the requests ([route patterns in place of raw paths](../../../pkg/platform/telemetry/DESIGN.md#tele-0005)).
- Reason: An ID of the data recipient in a path is not an identifier that a bank gave. The account ID of the bank stays in the tables of the collected data only.

[asvs-v8]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x17-V8-Authorization.md
[asvs-v14]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x23-V14-Data-Protection.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-pagination]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination
[cds-schema-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingaccountv3
[cds-schema-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingbalance
[cds-schema-transaction]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactionv2
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5
