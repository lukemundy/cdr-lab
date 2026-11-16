# Mock data holder: Online bank site accounts

This area gives the accounts of a consumer on the online bank site: the balances and transactions, the opening of an account and the close of an account. The [site API](online-bank-site-api.md) gives the requests of these pages.

## Accounts and transactions

- <a id="dh-2200"></a>**DH-2200** The accounts page shows each account of the customer of the session. The open accounts are first. The closed accounts follow, in a group of their own. ([CDS: Get Accounts][cds-get-accounts]: `open-status` is `ALL` by default; [Judgement: the account data of the APIs of an arrangement](online-bank-site-api.md#dh-1790))
- <a id="dh-2201"></a>**DH-2201** For each account, the accounts page shows the display name, the product name, the label of the product category, the masked number, the current balance and the available balance. ([CDS: BankingAccountV3][cds-schema-account]; [CDS: BankingBalance][cds-schema-balance]; [the masked number in the consumer UI](banking-apis.md#dh-1015))
- <a id="dh-2202"></a>**DH-2202** When the customer has no open account, the accounts page says so. It says that the customer cannot share data under the Consumer Data Right until the customer has an open account. It links to the product list, where the consumer can open an account. ([CDR Rules][cdr-rules], rule 1.10B(1)(b); [the authorisation page of a customer with no open account](authorisation.md#dh-0823))
- <a id="dh-2203"></a>**DH-2203** The page of an account shows the fields of the accounts page, the credit limit, the status, the date of creation and the transactions of the account. ([CDS: BankingAccountV3][cds-schema-account]; [CDS: BankingBalance][cds-schema-balance])
- <a id="dh-2204"></a>**DH-2204** The transactions on the page of an account are those of Get Transactions For Account with no filter. For example, the page shows the transactions from 90 days before the time of the request to the time of the request. ([CDS: Get Transactions For Account][cds-get-transactions]; [the default period of Get Transactions For Account](banking-apis.md#dh-1085); [Judgement: the transactions with no filter](online-bank-site-api.md#dh-1795))
- <a id="dh-2205"></a>**DH-2205** The transactions on the page of an account are in the order of Get Transactions For Account, with pages. ([CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: the transactions with no filter](online-bank-site-api.md#dh-1795))
- <a id="dh-2206"></a>**DH-2206** For each transaction, the page shows the effective date, the description and the amount with its sign. A pending transaction also shows "Pending". ([CDS: BankingTransactionV2][cds-schema-transaction])
- <a id="dh-2207"></a>**DH-2207** The detail of a transaction row shows each other field of the transaction that has a value. Each value of a CDS enumeration has a label in plain language. ([CDS: BankingTransactionV2][cds-schema-transaction]; [plain language in a consumer UI](../../../web/packages/ui/design/content.md#ui-0142))
- <a id="dh-2208"></a>**DH-2208** The site shows the same values as the banking APIs for the same account at the same time. These are the fields of Get Accounts, the balance of Get Account Balance and the transactions of Get Transactions For Account. Only the `accountId` and the `transactionId` are different. ([the resource identifiers](banking-apis.md#dh-1040); [CDS: Get Accounts][cds-get-accounts]; [CDS: Get Account Balance][cds-get-account-balance]; [CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: the account data of the APIs of an arrangement](online-bank-site-api.md#dh-1790))
- <a id="dh-2209"></a>**DH-2209** The page of an account of a different customer, or of an account that does not exist, is the "not found" page of the site. ([RFC 9110, section 15.5.5][rfc9110-404]; [Judgement: the account data of the APIs of an arrangement](online-bank-site-api.md#dh-1790))

## Open of an account

- <a id="dh-2210"></a>**DH-2210** The consumer can open an account of each current product of the brand, in each CDS product category. For example, a product for a business and a product for a young person can open. ([CDS: Product Categories][cds-product-categories]; [a current product](product-reference-data.md#dh-0219); [Judgement: an open of each current product](#dh-2300))
- <a id="dh-2211"></a>**DH-2211** The data holder has no rule for the age of the consumer or for a business at an open. ([Judgement: an open of each current product](#dh-2300))
- <a id="dh-2212"></a>**DH-2212** The data holder does no identity check and no credit assessment at an open. It approves each open immediately. ([Judgement: no identity check and no credit assessment at an open](#dh-2301))
- <a id="dh-2213"></a>**DH-2213** The product category alone gives the kind of a lending product. The revolving credit categories are these. The consumer gives the credit limit for them. ([CDS: Product Categories][cds-product-categories]; [Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
  - `CRED_AND_CHRG_CARDS`
  - `OVERDRAFTS`
  - `MARGIN_LOANS`
  - `BUY_NOW_PAY_LATER`
- <a id="dh-2214"></a>**DH-2214** The loan categories are these. The consumer gives the loan amount for them. ([CDS: Product Categories][cds-product-categories]; [Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
  - `PERS_LOANS`
  - `BUSINESS_LOANS`
  - `RESIDENTIAL_MORTGAGES`
  - `LEASES`
  - `TRADE_FINANCE`
- <a id="dh-2215"></a>**DH-2215** The detail of each product has the action "Open an account". The action goes to the open page of the product. ([Judgement: an open of each current product](#dh-2300))
- <a id="dh-2216"></a>**DH-2216** The open page needs a session. ([a page that needs a session](online-bank-site-sign-in.md#dh-2027))
- <a id="dh-2217"></a>**DH-2217** The open page names the product and the label of its category. It says that the data holder opens an account of this product in the name of the consumer. It has the actions "Open account" and "Cancel". ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]: an account cannot be removed, so the page is the confirmation)
- <a id="dh-2218"></a>**DH-2218** The open page of a lending product has one required field. The field is "Credit limit" for revolving credit, and "Loan amount" for a loan. ([Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
- <a id="dh-2219"></a>**DH-2219** The open page of a lending product says in plain language that the data holder does no credit check and approves the account immediately. ([plain language in a consumer UI](../../../web/packages/ui/design/content.md#ui-0142); [Judgement: no identity check and no credit assessment at an open](#dh-2301))
- <a id="dh-2220"></a>**DH-2220** The amount field accepts an amount in dollars as a person usually types it. The page sends it as an AmountString. For example, `5000`, `5,000` and `$5,000.50` are valid. ([CDS: Common Field Types][cds-common-field-types]; [Judgement: an amount as a person types it on the open page](#dh-2303))
- <a id="dh-2221"></a>**DH-2221** When the text of the amount field is not an amount, the page shows a field error in plain language and sends no open. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [Judgement: an amount as a person types it on the open page](#dh-2303))
- <a id="dh-2222"></a>**DH-2222** The credit limit and the loan amount are more than `0.00`. When a value is not valid, the site shows a field error in plain language, and the data holder opens no account. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
- <a id="dh-2223"></a>**DH-2223** The open of a deposit product has no credit limit and no loan amount. ([Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
- <a id="dh-2224"></a>**DH-2224** When the product is not current or does not exist, the site shows a field error in plain language, and the data holder opens no account. The data holder uses the same rule for a current product and the same products as the site API. ([a current product](product-reference-data.md#dh-0219); [the current products of the site API](online-bank-site-api.md#dh-1728))
- <a id="dh-2225"></a>**DH-2225** The data holder opens the account with the rules of [a create of an account in the operator interface](operator-interface.md#dh-0722). The account gets these values. ([the fields of an account](operator-interface.md#dh-0737); [Judgement: no link from an account to its product](#dh-2306); [Judgement: the BSB of the brand for an account that the consumer opens](#dh-2304); [Judgement: 8 random digits for the account number](#dh-2305))
  - `productCategory`: the `productCategory` of the product.
  - `productName` and `displayName`: the `name` of the product.
  - `bsb`: [the `bsb` of the brand file](brand.md#dh-0334).
  - `accountNumber`: 8 random digits.
  - `creditLimit`: the credit limit of the consumer for revolving credit, and `0.00` for a deposit product and a loan.
- <a id="dh-2226"></a>**DH-2226** When [the pair of the BSB and the account number is in use](operator-interface.md#dh-0740), the data holder makes a new account number. It makes at most 3 attempts in all. When the 3 attempts fail, the open fails, and the data holder changes nothing. ([Judgement: 8 random digits for the account number](#dh-2305))
- <a id="dh-2227"></a>**DH-2227** The open of a loan adds one transaction to the account: the drawdown. The data holder writes the account and the drawdown in the same write. When the write of the drawdown fails, the data holder opens no account. ([Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
- <a id="dh-2228"></a>**DH-2228** The drawdown has these values. ([CDS: BankingTransactionV2][cds-schema-transaction]; [Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
  - `type`: `OTHER`
  - `status`: `POSTED`
  - `description`: "Loan drawdown"
  - `amount`: minus the loan amount
  - `reference`: an empty string
  - `postingDateTime` and `executionDateTime`: the time of the open
- <a id="dh-2229"></a>**DH-2229** After the open of a loan, the current balance is minus the loan amount and the available balance is `0.00`, by [the current balance rule](banking-apis.md#dh-1013) and [the available balance rule](banking-apis.md#dh-1014). ([CDS: BankingBalance][cds-schema-balance]: `currentBalance` is negative when the customer owes money; [Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
- <a id="dh-2230"></a>**DH-2230** The banking APIs give the same values for a lending account that the consumer opens as for an account that the operator creates with the same credit limit. For a loan, the operator account also has the same transaction. ([CDS: Get Account Balance][cds-get-account-balance]; [Judgement: a credit limit for revolving credit and a drawdown for a loan](#dh-2302))
- <a id="dh-2231"></a>**DH-2231** The account has no link to its product. A later change to the product, or its end, does not change the account. ([CDS: BankingAccountV3][cds-schema-account]; [Judgement: no link from an account to its product](#dh-2306))
- <a id="dh-2232"></a>**DH-2232** An account that the consumer opens is not in an arrangement that exists. The next authorisation of the customer can include it. ([the authorisation page lists each open account](authorisation.md#dh-0819); [the amendment of an authorisation is a non-goal](README.md))
- <a id="dh-2233"></a>**DH-2233** After an open, the site shows the page of the new account, with a status message that names the product. The focus is on the heading of the page. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="dh-2234"></a>**DH-2234** While an open is in progress, the busy state of the button stops a second open. The site never sends an open again by itself. ([design system DESIGN.md][ui-design]; [Judgement: no idempotency key for an open](#dh-2307))
- <a id="dh-2235"></a>**DH-2235** The data holder has no idempotency key for an open. For example, two requests open two accounts. ([Judgement: no idempotency key for an open](#dh-2307))

## Close of an account

- <a id="dh-2240"></a>**DH-2240** The page of an open account has the action "Close account". A closed account has no action. ([the close of an account by the operator](operator-interface.md#dh-0723); [Judgement: the close of the operator console](online-bank-site-api.md#dh-1796))
- <a id="dh-2241"></a>**DH-2241** Before a close, the site asks the consumer to confirm in a dialog. The dialog gives the display name, the masked number and the current balance of the account. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4])
- <a id="dh-2242"></a>**DH-2242** The dialog says that the close cannot be reversed, and that this demonstration bank pays out no balance. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]; [Judgement: no balance rule at a close](#dh-2308))
- <a id="dh-2243"></a>**DH-2243** When the account is the last open account of the customer, the dialog also says that the consumer then cannot share data under the Consumer Data Right. It names the app (the `clientName`) of each arrangement that ends, and says that the data holder tells each app. ([the end of eligibility](end-of-an-arrangement.md#dh-0918); [the notice to the data recipient](end-of-an-arrangement.md#dh-0924); [Judgement: the close of the operator console](online-bank-site-api.md#dh-1796))
- <a id="dh-2244"></a>**DH-2244** The data holder closes the account with the rules of a close in the operator console. The close sends the arrangements that the confirmation named. ([a close sends the arrangements of its confirmation](operator-console-api.md#dh-1650); [Judgement: the close of the operator console](online-bank-site-api.md#dh-1796))
- <a id="dh-2245"></a>**DH-2245** When the close now ends different arrangements, the data holder changes nothing. The dialog then says that the accounts changed, and the site shows the current state. ([a close that ends different arrangements](operator-console-api.md#dh-1651); [Judgement: the close of the operator console](online-bank-site-api.md#dh-1796))
- <a id="dh-2246"></a>**DH-2246** A close has no rule for the balance or for pending transactions. The closed account keeps its balance and its transactions. ([Judgement: no balance rule at a close](#dh-2308))
- <a id="dh-2247"></a>**DH-2247** The data holder records no channel of a close. For example, a close on the site and a close by the operator give the same record. ([CDR Rules][cdr-rules], rule 9.3; [Judgement: no record of the channel of a close](#dh-2309))
- <a id="dh-2248"></a>**DH-2248** After a close, the dialog closes and a status message names the closed account. The site shows the accounts page with the focus on its heading, and the account is in the group of the closed accounts. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="dh-2249"></a>**DH-2249** When a close fails for a reason that is not a changed set of arrangements, the dialog stays open and shows the error. The site never sends a close again by itself. ([design system DESIGN.md][ui-design])

## Judgements

### <a id="dh-2300"></a>DH-2300 An open of each current product

- Kind: choice
- Basis: Step 3.1. The CDS and the ACCC give no rule for how a data holder opens an account. Eligibility needs a consumer who is 18 years or older ([CDR Rules][cdr-rules], rule 1.10B(1)(a)).
- Reason: The lab has no real consumers, and a data recipient must test lending data from an authorisation that a consumer gives. A product that the site lists but cannot open has no value. A customer has no date of birth, and a sole trader holds a business account as a person, so a rule for the age or for a business has no data to check.

### <a id="dh-2301"></a>DH-2301 No identity check and no credit assessment at an open

- Kind: choice
- Basis: Step 3.1. A bank identifies a customer before it gives a designated service ([Anti-Money Laundering and Counter-Terrorism Financing Act 2006][aml-ctf-act], section 32). A credit provider assesses a credit contract with a consumer ([National Consumer Credit Protection Act 2009][nccp-act], Chapter 3).
- Reason: The lab has no real consumers and enters no credit contract, so these Acts do not apply. The online bank site is lab-only, and a production data holder has its own site. Thus this is not a [production gap][production-gaps].

### <a id="dh-2302"></a>DH-2302 A credit limit for revolving credit and a drawdown for a loan

- Kind: choice
- Basis: Step 3.2. `creditLimit` is the maximum credit that is available, and `currentBalance` is negative when the customer owes money ([CDS: BankingBalance][cds-schema-balance]). The CDS has no transaction type for a drawdown, so the type is `OTHER` ([CDS: BankingTransactionV2][cds-schema-transaction]).
- Reason: Data recipients test with the balances of loans, and a loan with a balance of 0 is not realistic lending data. One posted transaction gives the balance of a loan with the balance rule, and the data holder holds no new field. The kind comes from the CDS category only, so each product of a category opens in the same way.

### <a id="dh-2303"></a>DH-2303 An amount as a person types it on the open page

- Kind: choice
- Basis: Step 3.3. An amount in the CDS is an AmountString ([CDS: Common Field Types][cds-common-field-types]).
- Reason: A consumer does not type a CDS AmountString. The site API and the operator interface keep one amount rule. The page converts only the format, and the data holder checks the value, so the UI has no copy of the value rules.

### <a id="dh-2304"></a>DH-2304 The BSB of the brand for an account that the consumer opens

- Kind: choice
- Basis: Step 3.3. [ADR 0023][adr-0023] gives the `bsb` field of the brand directory. Observed on 2026-11-01: the [AusPayNet BSB directory][auspaynet-bsb] had no BSB that starts with `372` to `375`, the BSBs of the preset brands.
- Reason: A BSB is a fact of the brand, as its name is, and the accounts of a brand usually have the same few BSBs. A setting of the deployment puts a fact of the brand in each deployment. The lab makes no payments, so a BSB that AusPayNet gives later has no effect.

### <a id="dh-2305"></a>DH-2305 8 random digits for the account number

- Kind: choice
- Basis: Step 3.3. The operator interface permits [6 to 10 digits](operator-interface.md#dh-0739).
- Reason: The open uses the same method as [the customer number](consumer-authentication.md#dh-1201): random digits, and a new value when the value is in use.

### <a id="dh-2306"></a>DH-2306 No link from an account to its product

- Kind: choice
- Basis: Step 3.3. No CDS account schema has a product ID ([CDS: BankingAccountV3][cds-schema-account], [CDS: Get Account Detail][cds-get-account-detail]).
- Reason: No rule or reader uses a link, and an account that the operator creates has no product. The open copies the category and the name of the product into the account. The account keeps its name when the product changes.

### <a id="dh-2307"></a>DH-2307 No idempotency key for an open

- Kind: choice
- Basis: Step 3.3. The UI never sends an action again by itself ([design system DESIGN.md][ui-design]).
- Reason: A second account after two requests is what the consumer asked for, and it has no effect on a different person in the lab. A key adds a store for a case that the busy state already stops.

### <a id="dh-2308"></a>DH-2308 No balance rule at a close

- Kind: choice
- Basis: Step 3.3. The CDS and the [CDR Rules][cdr-rules] give no rule for the close of an account. The close by the operator has no balance rule.
- Reason: Payments are out of scope. Thus a rule of a zero balance stops the close of each account with transactions on the site.

### <a id="dh-2309"></a>DH-2309 No record of the channel of a close

- Kind: choice
- Basis: Step 3.3. Rule 9.3 requires records of the authorisations, their withdrawals and the notices, and no record of how an account closed ([CDR Rules][cdr-rules]).
- Reason: A field needs a rule or a reader that uses it ([ADR 0021][adr-0021]). The end of eligibility has a reason, not a method.

[adr-0021]: ../../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0023]: ../../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md
[production-gaps]: ../../../docs/production-gaps.md
[ui-design]: ../../../web/packages/ui/DESIGN.md
[aml-ctf-act]: https://www.legislation.gov.au/C2006A00169/latest/text
[auspaynet-bsb]: https://bsb.auspaynet.com.au/
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[nccp-act]: https://www.legislation.gov.au/C2009A00134/latest/text
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-get-account-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-balance
[cds-get-account-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-detail
[cds-get-accounts]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-accounts
[cds-get-transactions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[cds-schema-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingaccountv3
[cds-schema-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingbalance
[cds-schema-transaction]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactionv2
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
