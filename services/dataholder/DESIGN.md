# Mock data holder: design

This document gives the specification and the requirements for the mock data holder. It does not describe the implementation.

The [overview of the specification](design/README.md) gives the summary, the context, the areas, the goals and the non-goals. Each area file in `design/` gives the rules of its area.

## Online bank site

### Accounts, balances and transactions

- The accounts page shows each account of the customer of the session: first the open accounts, then the closed accounts in a group of their own. For each account, it shows the display name, the product name, the label of the product category, the masked number, the current balance and the available balance. (See Customers and accounts for the masked number)
- When the customer has no account, or no open account, the accounts page says so. It also says that the customer cannot share data under the Consumer Data Right until the customer has an open account, and links to the product list, where the consumer can open an account. ([CDR Rules][cdr-rules], rule 1.10B(1)(b); see Consent)
- The page of an account shows the fields of the accounts page, the credit limit, the status, the date of creation and the transactions of the account.
- The transactions are those of Get Transactions For Account with no filter: from 90 days before the time of the request to the time of the request, in the order of Get Transactions For Account, with pages. ([CDS: Get Transactions For Account][cds-get-transactions]: the default `oldest-time` is `newest-time` minus 90 days; see Transactions and Decisions)
- For each transaction, the page shows the effective date, the description, the amount with its sign, and "Pending" for a pending transaction. The other fields of the transaction that have a value show in the detail of the row, with a label in plain language for each value of a CDS enumeration. ([CDS: BankingTransactionV2][cds-schema-transaction]; [design system DESIGN.md: Content][ui-content])
- The site shows the same values as the APIs of a consent for the same account at the same time: the fields of Get Accounts, the balance of Get Account Balance and the transactions of Get Transactions For Account. Only the `accountId` and the `transactionId` are different ([the site API uses the functions of the APIs of an arrangement](design/online-bank-site-api.md#dh-1751); see Decisions).
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
- The data holder closes the account with the rules of a close in the operator console: the close sends the arrangements that the confirmation named, and when the close now ends different arrangements, the data holder changes nothing. The dialog then says that the accounts changed, and the site shows the current state. ([console API: a close that ends different arrangements](design/operator-console-api.md#dh-1651); see Decisions)
- A close has the same effects as a close by the operator: the end of each arrangement for the end of eligibility at the close of the last open account, at the time of the close, the revocation of its tokens, and the notice to the data recipient. (See End of eligibility and Notice to the data recipient)
- A close has no rule for the balance or for pending transactions. The closed account keeps its balance and its transactions. (See Decisions)
- After a close, the dialog closes, a status message names the closed account, and the site shows the accounts page with the focus on its heading. The account is in the group of the closed accounts. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- When a close fails for a reason that is not a changed set of arrangements, the dialog stays open and shows the error. The site never sends a close again by itself.

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Paths of the brand assets | On the console listener: `/brand/logo.svg`, `/brand/mark.svg` and `/brand/console.css`, which sets only the roles of the brand that the console uses. The paths of the public listener are in [the paths of the brand files](design/authorisation.md#dh-2090). | Judgement: the console has the same paths for the same files, and a different name for a stylesheet with different content. |
| Account number on the online bank site | Masked, as on each page for the consumer | The CDS display rule for a masked number ([CDS: Common Field Types][cds-common-field-types]) and the rule of Customers and accounts: a different person can see the screen of a consumer. |
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
| Balance at a close | No rule for the balance or for pending transactions. The dialog shows the balance and says that the bank pays out no balance. | Neither the CDS nor the CDR Rules give a rule for the close of an account. The operator close has no such rule. Judgement: payments are out of scope, so a rule of a zero balance stops the close of an account with transactions on the site. |
| Channel of a close | No record of the channel (site or operator) and no new event | Rule 9.3 requires records of authorisations, their withdrawals and the notices, and no record of how an account closed ([CDR Rules][cdr-rules]). [ADR 0021][adr-0021]: a field needs a rule or a reader that uses it. The end of eligibility has a reason, not a method. |

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
