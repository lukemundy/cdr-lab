# Online bank of a data holder, milestone 3: research

- Research date: 2026-11-02
- Applies to: CDS 1.36.0, CDR Rules Compilation No. 10, WCAG 2.2
- File and line references are to the repository on the research date. Paths in `internal/` are in `services/dataholder/`.
- Purpose: the facts for milestone 3 of the project "A consumer uses the online bank of each data holder": "The consumer opens an account from a product of the brand, and closes an account. The close of the last open account ends each consent for the end of eligibility." Project acceptance criterion: "A consumer closes their last open account, and each of their consents then ends for the end of eligibility."
- Related notes: [Online bank: sign-up, sign-in and accounts](2026-11-02-online-bank-sign-in-and-accounts.md), [Online bank: landing page and products](2026-11-02-online-bank-landing-and-products.md), [Data holder: the end of data sharing](2026-10-16-data-holder-end-of-sharing.md). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | CDS 1.36.0: Product Categories ("Deposit Products", "Lending Products"), source `slate/source/includes/banking/_product_categories.md` at tag `v1.36.0` | https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/slate/source/includes/banking/_product_categories.md |
| S2 | CDS 1.36.0 banking OpenAPI: `BankingAccountV3`, `BankingAccountDetailV5`, `BankingProductCategoryV2` (pinned copy: `conformance/dataholder/spec/cds_banking.json`, `info.version` 1.36.0) | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#tocSbankingaccountv3 |
| S3 | CDS 1.36.0: CDR Arrangement Revocation endpoint | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#end-points |
| S4 | CDR Rules, Compilation No. 10: rules 1.10B, 4.25, 4.26, 4.26A, 9.3; Schedule 3, clauses 1.3 and 2.1 | https://www.legislation.gov.au/F2020L00094/2025-03-04/text |
| S5 | CDR Support: "Closed accounts in the banking and non-bank lenders sectors" (article 14519397304079, updated 2025-12-04, read through the help centre API) | https://cdr-support.zendesk.com/hc/en-us/articles/14519397304079 |
| S6 | WCAG 2.2: SC 3.3.4 Error Prevention (Legal, Financial, Data) | https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data |
| S7 | National Consumer Credit Protection Act 2009, Chapter 3 (responsible lending: sections 128 to 133) | https://www.legislation.gov.au/C2009A00134/latest/text |
| S8 | Anti-Money Laundering and Counter-Terrorism Financing Act 2006, section 32 (customer identification before a designated service) | https://www.legislation.gov.au/C2006A00169/latest/text |

## 1. What the repository has now

Facts:

- The data holder `DESIGN.md` puts "The opening and the close of an account by the consumer on the online bank site" out of scope now (`DESIGN.md:49`). "An account is open when the operator creates it. The operator can close an open account. A closed account cannot open again." (`DESIGN.md:144`)
- Customers and accounts are rows in tables, not events (`internal/migrations/00006_create_customers_and_accounts.sql`; `internal/adapter/customerstore/customerstore.go:52-85`). ADR 0003 and ADR 0021 apply to the arrangements and the client registrations, which are events. An open or a close of an account adds no event type.
- `app.Customers.AddAccount` validates the account, sets a random `AccountID`, `OpenStatus` `OPEN` and `CreatedAt` (`internal/app/customers.go:94-112`). `Account.Validate` checks `displayName` and `productName` (1 to 100 characters), `productCategory` (a value of `BankingProductCategoryV2`), `bsb` (6 digits), `accountNumber` (6 to 10 digits) and `creditLimit` (zero or more) (`internal/domain/customer.go:128-144`). The store gives `ErrAccountExists` for a pair of BSB and account number that is in use (`customerstore.go:66-85`).
- The operator gives each account field, also the BSB and the account number. The data holder creates none of them (`DESIGN.md:727`, Operator interface table; `internal/adapter/console/creates.go:63-89`; the console form has free text fields for the product name, the BSB and the account number, `web/apps/dataholder-console/src/customers/CreateAccountPage.tsx:64-74`). The OpenID conformance test makes a random 8-digit account number with the BSB `372200` (`conformance/openid/openid_test.go:141-149`).
- The account has no link to a product: `domain.Account` has `ProductCategory` and `ProductName` and no product ID (`internal/domain/customer.go:58-68`).
- `brand.json` has only `brandName` (`brands/wattle-bank/brand.json`). ADR 0023: "The data holder refuses an unknown field and an unknown file. A change that adds a file or a field is a change of this contract." (`docs/adr/0023-…md:43`). No BSB is a setting or a brand field now.
- The site API has no change of an account. Its routes are the sign-up, the session, the accounts, a balance and the transactions (`internal/adapter/banksite/banksite.go:63-75`). The accounts page already shows the open accounts first and the closed accounts in a group of their own (`DESIGN.md`, Online bank site: Accounts, balances and transactions).
- A customer with no open account gets this text on the accounts page: the customer "cannot share data under the Consumer Data Right until the customer has an open account" (`DESIGN.md`, Accounts, balances and transactions). A new customer has no account (`DESIGN.md:880`; Decisions, "Account at sign-up", `DESIGN.md:1246`).

## 2. Opening an account

### Facts

- The CDS puts the 13 product categories in two groups. "Deposit Products": `REGULATED_TRUST_ACCOUNTS`, `TERM_DEPOSITS`, `TRANS_AND_SAVINGS_ACCOUNTS`, `TRAVEL_CARDS`. "Lending Products": `BUSINESS_LOANS`, `BUY_NOW_PAY_LATER`, `CRED_AND_CHRG_CARDS`, `LEASES`, `MARGIN_LOANS`, `OVERDRAFTS`, `PERS_LOANS`, `RESIDENTIAL_MORTGAGES`, `TRADE_FINANCE` [S1].
- `BankingAccountV3` requires `accountId`, `accountOwnership`, `displayName`, `maskedNumber`, `productCategory` and `productName`. It has no product ID. `productName` is "The unique identifier of the account as defined by the data holder (akin to model number for the account)". `displayName` is "The display name of the account as defined by the bank" [S2]. `BankingAccountDetailV5` adds `bsb`, `accountNumber`, `bundleName`, `specificAccountUType` with `termDeposit`, `creditCard` and `loan`, rates, `features`, `fees` and `addresses`, all optional, and no product ID [S2]. The data holder gives none of these optional fields except `bsb` and `accountNumber` (`DESIGN.md`, Decisions, "Optional fields of an account and its detail").
- Thus no CDS field needs a link from an account to a product. The product category and the product name of the product give the two required fields.
- The current products of each preset brand in a deposit category (`brands/*/products.json`):
  - `wattle-bank`: 3 `TRANS_AND_SAVINGS_ACCOUNTS`, 1 `TERM_DEPOSITS`. Its `TRAVEL_CARDS` product is future, and `Blossom Christmas Saver` has ended.
  - `redgum-bank`: 4 `TRANS_AND_SAVINGS_ACCOUNTS` (2 for business: "Ledger Business Account", "Sapwood Business Saver"), 1 `TERM_DEPOSITS`.
  - `harbour-bank`: 3 `TRANS_AND_SAVINGS_ACCOUNTS`, 1 `TRAVEL_CARDS`, 1 `TERM_DEPOSITS`.
  - `kestrel-bank`: 3 `TRANS_AND_SAVINGS_ACCOUNTS`, 1 `TERM_DEPOSITS`.
  - Each preset brand has 4 or more current deposit products. The longest product name of the presets has 31 characters. A product name in a user brand directory has no maximum length (`internal/domain/product.go:116`), but an account `productName` has 100 characters at most (`customer.go:136`).
- Lending in real life: a credit provider must assess that a credit contract is "not unsuitable", and make reasonable inquiries, before it enters the contract [S7, sections 128 to 131]. A bank must identify a customer before it provides a designated service, for example an account [S8, section 32]. The lab does neither, and the project puts "the security of a production online bank" out of scope.
- Eligibility needs a CDR consumer who is "an individual who is 18 years of age or older" or "a person who is not an individual", and "an account holder or a secondary user for an account with the data holder that is open" [S4, rule 1.10B(1)]. For banking, the account must be "set up in such a way that, at that time, it can be accessed online" [S4, Schedule 3, clause 2.1]. A customer of the lab is a person with no date of birth (`DESIGN.md`, Customers and accounts), and each lab account "can be accessed online" (`DESIGN.md:145`).
- Customers that are organisations are out of scope (`DESIGN.md`, Scope). `redgum-bank` sells business products, but each customer is a person. A person who is a sole trader holds a business account as an individual.
- The accounts of an arrangement do not change after the confirmation, and the data holder has no amendment (`DESIGN.md`, Scope). Thus an account that the consumer opens is not in the arrangements that exist. It is available for the next consent.
- An open after the close of the last open account makes the customer eligible again. "An arrangement that ended does not start again." (`DESIGN.md`, End of an arrangement)

### Options for the products that a consumer can open

| | D. Deposit products only | A. Each current product | T. Transaction and savings accounts only |
|---|---|---|---|
| Realism | A deposit account opens online with no credit check, as at a real bank | A loan or a card opens with no assessment [S7], and its balance is 0 with no credit limit | Realistic, but no term deposit or travel card |
| Fields | `creditLimit` `0.00`, as the operator default | A card or an overdraft needs a credit limit, so the form or a rule gives one | As D |
| Rule | One list of 4 categories from the CDS [S1] | No rule | One category |
| Lending accounts | The operator opens them, as now | The consumer opens them | The operator opens them |

Recommendation: **D**. The site offers "Open an account" only for a current product in a CDS deposit category [S1]. For a lending product, the detail page says that the bank opens it after an application, and gives no action. The operator still opens lending accounts with a credit limit. Judgement: realism without a credit model, and the rule comes from the CDS grouping, not from a list of our own.

### Options for the account number and the BSB

The site must create the BSB and the account number, because the consumer cannot type them.

| | B1. A `bsb` field in `brand.json` | B2. A required setting, for example `BANK_SITE_BSB` | B3. A random BSB for each account |
|---|---|---|---|
| Realism | Each bank has its own BSB, and its accounts share it (`DESIGN.md`, Decisions, "Field of a `bsb` and an `accountNumber` that a different account has": "the accounts of a brand usually have the same few BSBs") | As B1 | Each account has a different BSB |
| Cost | A change of the brand directory contract and of ADR 0023 (`0023-…md:43`), a value for each preset | A setting for each of 4 instances in `compose.yaml`. A setting that is required only when the site is on | No setting and no contract change |
| Owner of the fact | The brand | The deployment | None |

- Account number: 8 random digits from the random source of `app.Customers`, with a maximum of 3 attempts after `ErrAccountExists`, as `Create` does for a customer number (`internal/app/customers.go:54-91`). This is in the 6 to 10 digits of the rule.
- Recommendation: **B1**. The BSB is a fact of the bank, as the name is. Each preset gets an invented BSB. A real BSB can have the same digits; this has no effect, because the lab makes no payments. Decision for the requirements PR. When the PR wants no contract change, B2 is the next option.

### Options for a link from the account to the product

| | N. No link (copy the fields) | L. A `product_id` column |
|---|---|---|
| CDS | Gives all required fields [S2] | No CDS field uses it [S2] |
| Reader | None | A link from the account page to the product detail. No rule reads it. |
| Product reload or `effectiveTo` | The account keeps its name and category | The link can point to a product that is no longer current |
| Operator accounts | The same shape | A null value for each account of the operator, or a required product for the operator too |

Recommendation: **N**. Copy `productCategory` and `name` of the product into `ProductCategory` and `ProductName`. The display name is the product name. No domain change, no migration and no event. Judgement: no rule or reader needs the link, and the console does the same (the DESIGN rule "Fields of a customer, an account and a transaction": only the optional fields that the document names).

### Recommendations for the open

- New use case on `app.Customers`: open an account for the customer of the session from a `productId`. It reads the product with the query of the site (`banksite.go:113-115`, `currentProducts`), refuses a product that is not current or not in a deposit category, makes the account from the product, and calls `AddAccount`. Thus the account rules of the operator apply by construction.
- A product name of more than 100 characters fails `Account.Validate`. Options: refuse the open with a problem in plain language, or limit product names to 100 characters in the product rules. Recommendation: limit the name in the product rules, so that each product can open. This changes the product load rules, which is a second change. Decision for the requirements PR.
- Site API: `POST /api/accounts` with `{"productId": "..."}`, 201 with the account (the fields of the list), and a `Location` of the account page. The product ID is in the body, not in the path (the site rule: no path value in telemetry). A product that is not current, not known or not a deposit product gives 400 with a field error at `/productId` and a text in plain language. This is a proposed path and needs approval (AGENTS.md).
- The site goes to the page of the new account, with a message that names the product. The new account has a balance of 0.

## 3. Closing an account

### Facts

- `app.Customers.CloseAccount` and `CloseAccountThatEnds` read the account, call `Account.Close` (closed gives `ErrAccountClosed`), and call the store (`internal/app/customers.go:114-191`). The store locks the customer row, sets `CLOSED`, checks if an open account stays, runs the check of the caller, and when no open account stays, runs the hook in the same transaction (`customerstore.go:137-192`).
- No rule reads the balance or the pending transactions at a close. The DESIGN gives none (`DESIGN.md`, Operator interface: "A transaction for a closed account, and a close of a closed account, give 409"). The CDS and the CDR Rules give no rule for the close of an account.
- The console sends `{"ends": [<arrangement IDs>]}`, the arrangements that its confirmation named. `CloseAccountThatEnds` closes only when the close ends the same set, else `ErrEndsChanged` and 409 (`internal/adapter/console/actions.go:35-58`; `customers.go:119-148`). The console API gives the permitted actions with their ends from `domain.PermittedAccountActions` (`internal/domain/action.go:35-67`).
- The confirmation of the console names each arrangement that ends and says that the data holder sends a notice (`DESIGN.md:771`; `web/apps/dataholder-console/src/customers/AccountActions.tsx:9-20`). It cites WCAG 2.2 SC 3.3.4: a close "cannot be reversed" (`DESIGN.md:770`).
- A closed account stays in each arrangement that has it, with `openStatus` `CLOSED` and its transactions (`DESIGN.md`, Get Accounts; Decisions, "Closed accounts", `DESIGN.md:1112`). The site already shows `openStatus` and the closed group.
- Data of a closed account is voluntary consumer data for a consent from 4 March 2025. When a data holder does not share it, the account selection "should show these accounts … but they must not be selectable", with a reason [S5]. The consent page "offers only open accounts" (`DESIGN.md:1112`).
- WCAG 2.2 SC 3.3.4: for pages that "cause legal commitments or financial transactions for the user to occur, that modify or delete user-controllable data", the submission is reversible, checked, or confirmed [S6].

### Options for the confirmation and its race

| | E. Reuse `CloseAccountThatEnds` and the `ends` body | S. A close with no body |
|---|---|---|
| What the consumer sees | The name of each app (`clientName`) whose consent ends | A general text: "your consents end" |
| A consent that starts between the confirmation and the close | 409; the site shows the new state | It ends with no warning to the consumer |
| Cost | The site API gives the close action of each open account with its ends, as the console API does | Less API, and a weaker confirmation |

Recommendation: **E**. One use case, one rule. The site API gives, for each open account, a `close` action with each consent that ends: its `arrangementId` and its `clientName`. The site shows a confirmation dialog with the display name and the masked number of the account, the current balance, and, for the last open account: "This is your last open account. When you close it, you cannot share your data under the Consumer Data Right, and these consents end: <names>. <brand> tells each app." On 409 the dialog says that the accounts changed, and the page loads the new state. The arrangement IDs are visible to the consumer of the session only; milestone 4 shows them on the dashboard too.

### Recommendations for the rules of a close

- No balance rule and no rule for pending transactions. Reasons: the operator flow has none, no source gives one, and the consumer cannot make a payment (out of scope), so a zero-balance rule makes the close of an account with transactions impossible without the operator. The confirmation shows the balance, and the closed account keeps it (Get Account Balance of a closed account still gives it). Decision for the requirements PR: a text that says the demonstration bank does not pay out the balance.
- Site API: `POST /api/accounts/{accountId}/close` with `{"ends": [...]}`, 204. A closed account gives 409. An account of a different customer gives 404 with no identifier in the `detail` (the M2 rule). Proposed path; needs approval.
- After the close, the account shows in the closed group. Its page shows no close action.

## 4. End of eligibility

### Facts

- An authorisation expires "if the CDR consumer ceases to be eligible in relation to the data holder" [S4, rule 4.26(1)(c)]. "If an authorisation … is withdrawn or otherwise expires, the data holder must notify the accredited person in accordance with the data standards." [S4, rule 4.26A]
- ACCC: "Where a CDR consumer closes all of their accounts with a data holder … they will cease to be an eligible CDR consumer … any existing authorisations given by the consumer to the data holder will expire (see CDR Rule 4.26(1)(c)). The obligation to provide a consumer dashboard will also end (see CDR Rules 1.13 and 1.15(1), and Schedule 3, clause 2.3)." When the consumer closes one of several accounts, "existing authorisations with the data holder remain in place" [S5].
- The CDS exceptions to the notice are a withdrawal through the data recipient, the natural expiry, and a status change on the Register [S3]. An end of eligibility is not an exception.
- Rule 9.3(1) requires records of authorisations, of "amendments to or withdrawals of authorisations" and of notifications of withdrawals of consents [S4]. It requires no record of the channel of an account close.
- The data holder decides eligibility with one query in the transaction of the close: `NOT EXISTS (… open_status = 'OPEN')` (`customerstore.go:170-176`). The decision is the customer has no open account (`DESIGN.md:145`, `:659`).
- When no open account stays, the hook `eligibilityjobs.Hook` inserts a River job `ineligible_customer` in the transaction of the close, scheduled 1 minute later (`internal/adapter/eligibilityjobs/eligibilityjobs.go:18-43`; `internal/root/process.go:52`). After the commit, the request calls `Arrangements.EndForIneligibleCustomer` at once (`customers.go:180-188`). It ends each arrangement that did not end, with `EndReasonConsumerNotEligible` and the time of the close, and revokes the grant (`internal/app/arrangements.go:94-118`, `:184-199`). The job repeats this after a stop; it is safe to repeat (`arrangements.go:94-95`; Decisions, "End of eligibility after a stop", `DESIGN.md:1148`).
- `EndReason.NeedsNotice` is true for `WITHDRAWN_BY_CONSUMER` and `CONSUMER_NOT_ELIGIBLE` (`internal/domain/notice.go:20-22`). The event store has the projection `noticejobs.Projection`, which inserts the notice job for an end that needs a notice (`internal/root/subcommands.go:99-103`). The notice and its records follow `DESIGN.md`, Notice to the data recipient.
- The end of the arrangement has no field for the method or the channel of the close. A withdrawal has a method; an end of eligibility has a reason only (`DESIGN.md`, Operator interface: `endReason`, and `withdrawalMethod` only for `WITHDRAWN_BY_CONSUMER`).

### Analysis

- A close on the site that calls `Customers.CloseAccountThatEnds` has the same effects as a close by the operator by construction: the same store, the same hook, the same immediate end, the same job and the same notice. No new event, field or job is necessary.
- No record of the channel of the close (site or operator) is necessary: rule 9.3 asks for none [S4], and ADR 0021 permits a field only when a rule or a reader uses it.
- The ACCC fact that the dashboard obligation ends when the consumer is not eligible [S5] is for milestone 4: the dashboard of a customer with no open account.

### Tests that exist

- Domain: `TestALossOfEligibilityEndsEachArrangementThatDidNotEnd`, `TestTheCloseOfTheLastOpenAccountEndsEachArrangementThatDidNotEndAndAClosedAccountHasNoAction` (`internal/domain/action_test.go:21`, `:58`).
- App: the close of the last open account ends the arrangements at the time of the close; the close succeeds when the end fails; `CloseAccountThatEnds` refuses a changed set (`internal/app/customers_test.go:277`, `:295`, `:343`, `:360`); the end is safe to repeat (`internal/app/arrangements_test.go:285-333`).
- Store: only the last close calls the hook; two closes at the same time give one "no open account"; an error of the hook or the check undoes the close (`internal/adapter/customerstore/customerstore_test.go:226-328`).
- Job: the hook inserts one job; a rollback removes it; the job ends the arrangements after a stop (`internal/adapter/eligibilityjobs/eligibilityjobs_test.go:47-111`, `restart_test.go:24`).
- Console: the close of the last open account ends each arrangement that its confirmation named (`internal/adapter/console/actions_test.go:121-176`).
- Conformance: `TestTheCloseOfTheLastOpenAccountEndsEachArrangementOfTheCustomer` through the operator interface: tokens stop, `CONSUMER_NOT_ELIGIBLE` at the close, one notice for each arrangement (`conformance/dataholder/eligibility_test.go:48-120`).
- No test closes an account through the site, because the site has no close.

## 5. Safety

### Facts

- `http.CrossOriginProtection` of the site listener refuses a cross-origin change, and it is the only CSRF control (`DESIGN.md`, Session). An open and a close are `POST` requests, so it applies to them.
- The design system: a button has "A disabled or busy state that the screen reader announces" (`web/packages/ui/DESIGN.md:86`). "The UI never sends the action again by itself" (`web/packages/ui/DESIGN.md:213`). The design system has a `ConfirmationDialog` (`web/packages/ui/src/components/confirmation-dialog.tsx:33`), which the console uses for a close.
- The site API decodes a body with no unknown field and a size limit (`banksite/session.go:26`, `:162`).

### Recommendations

- Confirmation of a close: the `ConfirmationDialog`, as in the console. A close cannot be reversed [S6].
- Open: SC 3.3.4 applies when an open is a legal commitment [S6]. An open also cannot be reversed (an account cannot be removed, `DESIGN.md:719`). The open page names the product, its category and the effect ("<brand> opens a <product name> account in your name"), with "Open account" and "Cancel". This is the "Confirmed" method of SC 3.3.4 with no second dialog.
- Double submit of an open: the busy button stops a second click; the UI does not send the request again. No idempotency key. Judgement: a second account after two requests is what the consumer asked for, and it is harmless in the lab. An `Idempotency-Key` header is more code for a case that the busy state already covers.
- A close with a network error: the dialog stays open with the text of the design system ("Refresh the page to see the current state"). A second close then gives 409, which is safe.

## 6. Telemetry, accessibility and tests

### Facts

- The data holder telemetry has no account data, no `accountId` and no customer number (`DESIGN.md`, Non-functional requirements: Telemetry). The site telemetry has no request body and no path value (`DESIGN.md`, Online bank site: Non-functional requirements). `TestTheBankSitePutsNoCustomerDataSessionIDOrPasswordInTheSpansMetricsOrLogs` checks this for the M2 routes (`internal/root/banksitetelemetry_test.go:48`).
- The equality test of M2 compares the site accounts with Get Accounts, Get Account Balance and Get Transactions For Account in one process (`internal/root/banksiteaccounts_test.go:281`).
- The browser test of M2 signs up and sees the account that the operator opens (`web/apps/dataholder-bank/e2e/sign-in.spec.ts:14`).
- The eligibility job logs only `error.type`, because a database error can contain the customer number (`eligibilityjobs.go:57-58`).

### Recommendations

- Telemetry: no product ID, BSB, account number, display name or arrangement ID of an open or a close in a log line. Extend the telemetry test to an open and a close of the last account, and check that no span, metric or log contains the BSB, the account number, the account ID or an arrangement ID. A product ID alone is product reference data, but with the session it says which product a consumer has, so treat it as account data.
- Accessibility: the open page and the dialog meet WCAG 2.2 AA with axe for each preset brand at 320 CSS px, as the other pages of the site. After an open, focus goes to the heading of the account page. After a close, the dialog closes, a status message names the closed account, and focus goes to the heading of the accounts page.
- Go test of equality (in `internal/root`): sign up, open two accounts from two deposit products through the site API, make an arrangement for both with `CreateArrangement`, and check that Get Accounts gives the same `displayName`, `productCategory`, `productName`, `maskedNumber`, `openStatus` and `creationDate` as the site, with `productCategory` and `productName` from the product. Close one through the site, and check `openStatus` `CLOSED` in Get Accounts.
- Go test of the end of eligibility (in `internal/root`, in one process, no waits): close the last open account through the site API with the `ends` of its action. Check that each arrangement of the customer that did not end has `CONSUMER_NOT_ELIGIBLE` with the time of the close, that its tokens are refused, and that a notice is `PENDING` (the notice job exists). Also a race: a consent that starts after the read of the action gives 409 and closes nothing.
- Tests in `banksite`: an open of a lending product, of a product that is not current and of an unknown product gives 400; a close of a closed account gives 409; a close of an account of a different customer gives 404; an open and a close need a session.
- Browser test (lab-web, its own data, no consent): sign up, open an account from the detail of a deposit product, see it with a balance of 0, close it with the confirmation, see it in the closed group and the text that the customer cannot share data. The conformance suite does not test the site (ADR 0019). The conformance test of eligibility through the operator interface stays as it is.

## 7. The known defect of `GET /api/session`

### Facts

- `readSession` gives 500 when the customer of the session does not exist: it sends each error of `Customers.Customer` to `problem.WriteUnexpected` (`internal/adapter/banksite/session.go:255-271`). The account routes end the session and give the 403 of a session that ended for `ErrCustomerNotFound` (`banksite/accounts.go:130-142`; test `TestASessionOfACustomerThatTheDataHolderNoLongerHasEnds`, `accounts_test.go:232`).
- "A customer, an account and a transaction cannot be removed." (`DESIGN.md:719`) A close changes only `open_status` of the account (`customerstore.go:160`).

### Analysis

- Milestone 3 adds no removal of a customer, so it does not make the defect reachable. The defect is reachable only when the database loses the customer while the process keeps the session, for example a new database with no restart of the data holder.
- Recommendation: fix it in the first M3 ticket that changes `session.go`, with the rule of `writeAccountError`, because it is one case in a switch and a test exists for the same rule on the accounts. It is not a requirement of M3.

## 8. Changes to the `DESIGN.md`

- Scope (`DESIGN.md:49`): remove the out-of-scope line, and add the open and the close to the site line.
- Customers and accounts (`:144`): "An account is open when the operator creates it, or when the consumer opens it on the online bank site." The consumer can also close an open account.
- Get Accounts (`:590`): "in the sequence in which the operator created them" becomes "in the sequence of their creation".
- End of an arrangement table (`:626`) and End of eligibility (`:659`): "When the operator or the consumer closes the last open account".
- Online bank site: new subsections for the open and the close, the site API rows, the tests, and the change of the text "Later changes add the opening and the close" (`:852`, `:880`).
- Decisions: the products that a consumer opens, the BSB and the account number, no link to the product, no balance rule at a close, the confirmation with `ends`, no idempotency key, and the change of "Account at sign-up" (`:1246`).
- `docs/production-gaps.md`: see section 9.

## 9. Risks and decisions for the requirements PR

| Item | Notes | Recommendation |
|---|---|---|
| Products that a consumer opens | Lending needs an assessment in real life [S7]. The CDS groups deposit and lending products [S1]. | Deposit categories only. The operator opens lending accounts. |
| BSB of a new account | No BSB is a fact of the data holder now. A new `brand.json` field changes the ADR 0023 contract. | A `bsb` field in `brand.json`, with an invented BSB for each preset. Else a required setting. |
| Account number | The pair is unique in the brand. | 8 random digits, 3 attempts, as the customer number. |
| Link from the account to the product | No CDS field needs it [S2]. Accounts are rows, not events. | No link. Copy the category and the name. No domain change. |
| Product name longer than 100 characters | A user directory can have one. `Account.Validate` refuses it. | Limit product names to 100 characters in the product rules, or refuse the open with a clear text. |
| Balance at a close | No source gives a rule. The consumer cannot make payments. | No rule. The confirmation shows the balance. |
| Confirmation of a close | The console names each arrangement that ends and refuses a changed set. | Reuse `CloseAccountThatEnds` and the `ends` body. Name each app in the dialog. |
| Channel of a close | Rule 9.3 asks for no record of it. | No field, no event. |
| Double submit of an open | A second account is harmless. | Busy button only. No idempotency key. |
| Confirmation of an open | SC 3.3.4 for a commitment that cannot be reversed [S6]. | An open page that names the product, with "Open account" and "Cancel". |
| Closed accounts on a new consent | The ACCC says that a closed account that the data holder does not share shows as not selectable, with a reason [S5]. The consent page hides closed accounts (`DESIGN.md:1112`). M3 makes closed accounts usual. | Out of M3 scope. Record it as an open question of the consent page. |
| Business products of `redgum-bank` | Each customer is a person. A sole trader holds a business account as a person. | Permit them. No rule for organisations. |
| Age of the consumer | Rule 1.10B(1)(a) needs 18 or older. Customers have no date of birth. Presets have junior savers. | Each customer is an adult. Permit junior products with no age rule. |
| Production gap | The site opens an account with no identification [S8] and no assessment. The site is lab-only, and the project puts a production online bank out of scope. | No new entry: a production data holder has its own site, and the gaps file excludes functions that a project put out of scope. Decision for the PR. |
| `GET /api/session` 500 | M3 adds no removal of a customer. | Fix it in the first M3 ticket that changes `session.go`. Not a requirement. |
| New names | `POST /api/accounts`, `POST /api/accounts/{accountId}/close`, the `close` action, the `bsb` field. | Propose them in the requirements PR before use (AGENTS.md). |

## Main findings

- No CDS account schema has a product ID. `productCategory` and `productName` of the product give the required fields, so an account needs no link to a product. Customers and accounts are rows, not events, so an open needs no new event.
- The CDS groups the product categories into deposit products and lending products. Each preset brand has 4 or more current deposit products.
- The operator gives the BSB and the account number now. The site must create them, and the data holder has no BSB of its own. A BSB in `brand.json` changes the ADR 0023 contract.
- The store already decides the end of eligibility in the transaction of the close, inserts a backstop job, and the request ends the arrangements with `CONSUMER_NOT_ELIGIBLE`. The projection of the end inserts the notice. A site close through `CloseAccountThatEnds` has the same effects by construction.
- No source gives a balance rule for a close. Payments are out of scope, so a zero-balance rule would block the close of most accounts.
- The ACCC says that the dashboard obligation ends when a consumer closes all accounts. This affects milestone 4.
- The ACCC says that closed accounts show as not selectable on a new consent when the data holder does not share them. The consent page hides them now.
- M3 does not make the `GET /api/session` 500 reachable, because no flow removes a customer.
