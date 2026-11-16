# Mock data holder: design

This document gives the specification and the requirements for the mock data holder. It does not describe the implementation.

The [overview of the specification](design/README.md) gives the summary, the context, the areas, the goals and the non-goals. Each area file in `design/` gives the rules of its area.

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
- The site shows the same products and the same values as Get Products and Get Product Detail of the brand at the same time ([the site API gives the values of Get Products](design/online-bank-site-api.md#dh-1729)).
- The site shows a label in plain language for each value of a CDS enumeration, for example `TRANS_AND_SAVINGS_ACCOUNTS` and `MONTHLY_ACCOUNT_FEE`. The label gives the meaning that the CDS gives for the value. ([CDS: BankingProductDetailV7][cds-schema-product-detail], [CDS: Product Categories][cds-product-categories]; [design system DESIGN.md: Content][ui-content]: plain language; see Decisions)
- The site shows each RateString as a percentage, for example `0.0125` as "1.25%", each AmountString as an amount, and each ISO 8601 duration in words, for example `P1M` as "each month" for an `applicationFrequency`, and `P55D` as "55 days". It shows an `additionalValue` in the format that the CDS gives for its type, and as text when the type gives no format. ([CDS: Common Field Types][cds-common-field-types], [CDS: BankingProductDetailV7][cds-schema-product-detail]; [design system DESIGN.md: Content][ui-content])
- A path that the site does not know shows the "not found" page of the site. Judgement: the UI routes in the browser, as the consoles do.

### Sign-up

- The sign-up page asks for the first name (optional), the last name and the email address, and for nothing else. It asks for no password and no identity document, and the data holder makes no identity check. (See Decisions)
- The data holder creates the customer with the rules of a create of a customer in the operator interface: the same rules for each value, and a new customer number from the rules of Customers and accounts. The operator interface and the operator console show the new customer, as they show a customer of the operator. (See Operator interface and Decisions)
- The page says that the bank is a demonstration, and that the consumer must give only invented details. It gives an example email address on a reserved domain, for example `name@example.test`. The data holder does not refuse an address on a different domain. ([RFC 6761, section 6.2][rfc6761-test]; see Decisions)
- When the data holder refuses a value, the page shows the error at its field and in an error summary. The text of each error is in plain language, and the site API gives it. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [design system DESIGN.md: Content][ui-content]; see Site API)
- An email address that a different customer of the brand has is refused at its field. The text says that the address is in use, and that the consumer can sign in or give a different address. (See Customers and accounts and Decisions)
- A successful sign-up starts a session (see Session). The site then shows the new customer number on a page of its own, with a copy button. The page says that the consumer signs in to the site with this number or with the email address. It also says that an app that gets the data of the consumer under the Consumer Data Right asks for the email address or this number, and then for a one-time password that the bank sends to the email address. ([CDS: User Identifiers][cds-user-identifiers]: an identifier "that is already known by the customer")
- A new customer has no account. The data holder does not open an account at sign-up. The consumer opens an account on the site (see Opening of an account), or the operator opens one with the operator interface or the operator console. (See Decisions)

### Sign-in and sign-out

- The sign-in page asks for the email address or the customer number, in one field, and a password. That field has `autocomplete="username"`. The password field has `type="password"` and `autocomplete="current-password"`. The consumer can paste into each field, and a password manager can fill each field. ([WCAG 2.2, SC 3.3.8][wcag-3-3-8], [WCAG 2.2, SC 1.3.5][wcag-1-3-5])
- Any password works. The page requires a value in the password field, but it does not send the password to the data holder. The site API takes only the user identifier, and it refuses a body with a different field, also `password`. Judgement: a person can type a real password from habit, and then no password goes to the data holder, its database or its telemetry. (See Decisions)
- For the email address or the customer number of a customer, the data holder starts a session for that customer (see Session). For a value that no customer has, the data holder refuses the sign-in, and the page says "We could not sign you in. Check your email address or customer number." The sign-in then shows which email addresses and customer numbers exist. (See Decisions)
- The pages that show the data of a customer need a session. With no valid session, the site goes to the sign-in page. When a session ended, the sign-in page says that the bank signed the consumer out.
- Each page of the site shows if the consumer is signed in. When the consumer is signed in, the header has a link to the accounts page, a link "Data sharing" to the consumer dashboard, and a control to sign out. ([CDR Rules][cdr-rules], rule 1.15(1)(f): "prominently displayed and readily accessible") After a sign-in, the site shows the accounts page. A sign-out ends the session, and the site then shows the sign-in page.

### Session

- The data holder keeps each session on the server. The session cookie has only the session ID, and the browser gets no token and no customer number. ([RFC 10017, section 7.1][rfc10017-cookie-session]; see Decisions)
- The session ID is a random value from a cryptographic random source, with at least 128 bits of entropy. ([OWASP ASVS 5.0, 7.2.3][asvs-v7])
- Each sign-in and each sign-up starts a new session with a new session ID. When the request has a session cookie, that session ends. ([OWASP ASVS 5.0, 7.2.4][asvs-v7]; [OWASP Session Management Cheat Sheet][owasp-session])
- A session ends 24 hours after its start, at a sign-out, or at a restart of the data holder. A session has no inactivity timeout. A customer can have more than one session at the same time, with no limit. ([OWASP ASVS 5.0, 7.1.1 and 7.1.2][asvs-v7]: the document gives these values; [WCAG 2.2, SC 2.2.1][wcag-2-2-1]: the exception for a time limit that is "longer than 20 hours"; see Decisions)
- The data holder keeps at most 10,000 sessions that did not end. When it has 10,000, it refuses a sign-in and a sign-up with 503, and it creates no customer. (See Decisions)
- After a session ends, its session ID gives no access. ([OWASP ASVS 5.0, 7.4.1][asvs-v7])
- A session of a customer that the data holder no longer has is not valid. The data holder ends it at the next request of the session, also at a read of the session. Judgement: the rule of the account routes, for each route.
- The session cookie has the name `__Host-bank-session-<DATA_HOLDER_BRAND_ID>`, with the `dataHolderBrandId` of the instance, and the attributes `Secure`, `HttpOnly`, `SameSite=Strict` and `Path=/`. It has no `Domain`, no `Max-Age` and no `Expires`. ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]; [OWASP Session Management Cheat Sheet][owasp-session]; see Decisions)
- The response to a sign-out expires the cookie with `Max-Age=0`. It has no `Clear-Site-Data` header. ([MDN: Clear-Site-Data][mdn-clear-site-data]: `"cookies"` affects the entire registered domain; see Decisions)
- The refusal of cross-origin changes of the listener (see Listener and security) is the only protection against cross-site request forgery. The site has no CSRF token. The sign-up, the sign-in and the sign-out are not `GET` requests, so the refusal also applies to them. ([OWASP CSRF Prevention Cheat Sheet][owasp-csrf]: Fetch Metadata; [Go: CrossOriginProtection][go-crossoriginprotection]; see Decisions)

### Separation from the CDR authentication

The consumer must not confuse the sign-in of the site, which asks for a password, with the CDR authentication, which does not. These rules keep the two apart:

- The sign-in page shows a note above the form, in a style that is different from the lock note of the consumer pages. The note says: "This is <brand name> online banking. This is a demonstration bank: any password works here, and you must not use a real password. When an app asks for your <brand name> data under the Consumer Data Right, you do not sign in here. <brand name> then asks only for your email address or customer number and a one-time password, and never for your password." ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Passwords, "The exact phrasing of this is at the discretion of the data holder"; the rest is judgement)
- The title of the sign-in page is "Sign in - <brand name> online banking", and its heading is "Sign in to online banking". The authentication and consent pages keep their titles and headings. (See Presentation)
- The site and the authentication and consent pages are on different origins (see Listener and security).
- The site has no link to the authentication and consent pages, and those pages have no link to the site. ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Friction forbids "Offering additional or alternative services" and "Referencing or including other documents" in the authentication flow; the link from the site is judgement: the data recipient starts the flow)
- The authentication and consent pages do not change. The CDR authentication does not read a session of the site: a session of the site does not replace a step of the authentication, and a CDR authentication does not start a session of the site. Judgement: the CDR authentication always asks for the email address or the customer number, and the OTP (see Consumer authentication).

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

### Listener and security

The site is a web UI. The [web UI library specification](../../pkg/platform/consolehttp/DESIGN.md) gives the rules of its listener, its browser security, its cache headers and its routes. The site is on its own listener, as the console is ([Listener and security](#listener-and-security)). These rules are different:

- The settings of the site have the prefix `BANK_SITE_`: `BANK_SITE_ENABLED` [turns the listener on](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0002), and `BANK_SITE_ALLOWED_HOSTS` is [the `Host` allowlist](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0003), with the values of the console allowlist. (See Decisions)
- The listener uses TLS, with the certificate and the key of the public listener and the TLS versions and cipher suites of the CDR endpoints. It does not request a client certificate. (See Transport security and Decisions)
- The responses have no `Strict-Transport-Security` header. ([RFC 6797, Appendix B][rfc6797-b]: a browser applies an HSTS policy to each port of the host; see Decisions)
- The site sets one cookie, the session cookie (see Session). The browser gets no token. ([ADR 0014][adr-0014]; see Decisions)
- The site listener serves the brand stylesheet, the logo and the mark of the brand at the paths of the public listener (`/brand/styles.css`, `/brand/logo.svg` and `/brand/mark.svg`), the photos of the brand at `/brand/<file name>` (for example `/brand/hero.webp`), and the fonts of the fixed set, with [the `Host` allowlist](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0004) and [the response headers](../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007) of the site. Judgement: the Content Security Policy permits styles, fonts and images from the same origin only.
- The data holder starts the site listener only in the process that runs the APIs. A process that runs only the workers needs no site setting ([a process needs only the settings that it uses](design/non-functional-requirements.md#dh-0010)).

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
| Text alternative of the photos | Empty. No alt text field in the brand directory. | Judgement: each photo is decoration, and the text of the page gives its meaning ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]). A field for the text is a change of the contract that gives no information. |
| Accent colour on the site | Shapes only, no text | Judgement: the accent of some presets has a contrast lower than 4.5:1 on white or on the hero colour, and the contrast test of the presets checks only the pairs that the layouts use as text ([WCAG 2.2, SC 1.4.3][wcag-1-4-3]). |
| Paths of the brand assets | `/brand/logo.svg`, `/brand/mark.svg` and `/brand/styles.css` on the public listener. The stylesheet sets the custom properties of the design system under `:root:root`. On the console listener: `/brand/logo.svg`, `/brand/mark.svg` and `/brand/console.css`, which sets only the roles of the brand that the console uses. | Judgement: the `logoUri` on the Register needs a fixed URL on the public base URI. The selector has a higher specificity than the `:root` of the design system tokens, so the order of the stylesheets on a page has no effect. The console has the same paths for the same files, and a different name for a stylesheet with different content. |
| Architecture of the online bank site | A React SPA that the data holder embeds and serves, with a site API in an OpenAPI document | [ADR 0014][adr-0014] names a data holder consumer dashboard as a SPA, and the design guide expects the first React consumer UI to add the consumer components to the design system ([design system DESIGN.md: Consumer pages][ui-consumer-pages]). Judgement: the later forms and dialogs of the site need the components that only the React package has, and Go pages put the bank sign-in near the CDR pages. |
| Listener of the online bank site | A lab-only listener of its own on each instance, off by default, with TLS. Port 8x85 on the host: 8185, 8285, 8385 and 8485. | Judgement: on the public listener, the site needs exceptions in the CORS rule and the not-found rule of the CDR APIs, and a lab-only feature goes into the listener of the CDR endpoints. The port is the next number of the block of each instance (see The four preset brands in the lab). |
| Settings of the online bank site | `BANK_SITE_ENABLED` and `BANK_SITE_ALLOWED_HOSTS` | Judgement: the same pattern as `CONSOLE_ENABLED` and `CONSOLE_ALLOWED_HOSTS`. `BANK_SITE` says what the listener serves, and it is different from the CDR names "public" and "secure". |
| TLS of the online bank site | The certificate and the key of the public listener, with the TLS rules of the CDR endpoints | Judgement: a real online bank uses HTTPS, and a session cookie of a later sign-in needs the `Secure` attribute, which the `__Host-` prefix requires ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]). The public certificate of each instance already has the host name `localhost`, so the PKI does not change. One TLS configuration for all TLS listeners is the simplest. |
| HSTS on the online bank site | No `Strict-Transport-Security` header | A browser applies an HSTS policy to each port of the host ([RFC 6797, Appendix B][rfc6797-b]). In the lab, all listeners are on `localhost`, and the consoles and the operator interfaces are plain HTTP, so an HSTS policy from the site stops them in the browser. The site is lab-only. |
| Cookies of the online bank site | One session cookie, `__Host-bank-session-<DATA_HOLDER_BRAND_ID>`, with `Secure`, `HttpOnly`, `SameSite=Strict` and `Path=/`, and no `Domain`, `Max-Age` or `Expires` | The `__Host-` prefix and `SameSite=Strict` are the recommendation for a session ID ([OWASP Session Management Cheat Sheet][owasp-session], [RFC 6265bis, section 4.1.3.2][rfc6265bis-host]). Cookies are not isolated by port ([RFC 6265, section 8.5][rfc6265-8-5]), so in the lab the cookies of each site go to the CDR pages, to the consoles and to the sites of the other banks. Judgement: a name that the CDR pages do not use keeps the two apart, and the brand ID in the name stops a sign-in at one bank from replacing the session of a different bank in the same browser. A UUID has only characters that a cookie name permits. With no expiry, the browser removes the cookie when it closes, and the server gives the lifetime. The cookie is a credential, but it names a session on the server and no script can read it, so the reason of [ADR 0014][adr-0014] for no token in the browser stays true. ADR 0014 gets an amendment for this cookie on the site only. The consoles still get no credential. |
| Text of the landing page | Text of the data holder, the same for each brand, with the brand name. The site API gives the brand name and the text of the footer. No new field in the brand directory. | Judgement: a new field changes the contract of the brand directory ([ADR 0023][adr-0023]), and the site needs no text for each brand now. The footer then has one source, as on the consumer pages. |
| Password statement on the site | In the footer of each page | The CDS requires the statement "in consumer-facing interactions and communications" ([CDS: Common Authentication Standards][cds-cx-authentication]), and the site is a consumer-facing interaction of the data holder. Judgement: the footer is on each page, and a consumer can open each page first. |
| Tests of the online bank site | Tests in the data holder, component tests with axe for each preset brand at 320 CSS px, and one browser test against the lab. Not in the conformance suite. | [ADR 0019][adr-0019]: the conformance suite tests only behaviour that the CDR defines, and a bank web site is not CDR behaviour. |
| Sign-up on the online bank site | The use case of a create of a customer by the operator, with no change. No new event, and no record of the channel of the sign-up. | Customers and accounts are records in tables, not events. The same use case gives the rules of the operator by construction. Rule 9.3 requires records of consents and their withdrawals ([CDR Rules][cdr-rules]), and no rule requires a record of how a person became a customer. [ADR 0021][adr-0021]: a field needs a rule or a reader that uses it, and none uses the channel. |
| Account at sign-up | None. The consumer opens an account on the site, or the operator opens one. | Judgement: the consumer selects the product, as at a real bank. The consequence: a new customer of the site cannot give a CDR consent until the customer has an open account ([CDR Rules][cdr-rules], rule 1.10B(1)(b)). |
| Email address domain at sign-up | Only the text of the page asks for an invented address, with an example on `.test`. No rule of the data holder refuses a different domain. | Judgement: the operator interface has no such rule, and the mail sink of the lab delivers no email. A rule only for the site makes the two flows different. [RFC 6761, section 6.2][rfc6761-test] reserves `.test` for tests. |
| Session after a sign-up | A sign-up starts a session | Judgement: one step less for the consumer. The page with the new customer number shows first, so the consumer still sees the number. |
| Password on the sign-in of the site | The page requires a value but does not send it. The site API takes only the user identifier and refuses a `password` field. | The project requires a sign-in in which any password works. Judgement: a person can type a real password from habit. When the UI does not send it, no password goes to the data holder, its database or its telemetry, and a test can prove that the API refuses it. An API that accepts and ignores the field cannot prove this. |
| Unknown user identifier at the sign-in of the site | A field error that says that the bank could not sign the consumer in, the same for an email address and a customer number | Judgement: with any password, each sign-in with the user identifier only shows which email addresses and customer numbers exist. The site is lab-only and its data is invented. The CDR authentication keeps its rule that hides which email addresses and customer numbers exist ([CDS: One-Time Password Credential Requirements][cds-otp]; see Consumer authentication). |
| Session store of the online bank site | On the server, in the memory of the process, with at most 10,000 sessions. A full store refuses a new session with 503. A restart ends each session. | A server-side session ends at a sign-out with no revocation list ([OWASP ASVS 5.0, 7.4.1][asvs-v7]), and [RFC 10017, section 7.1][rfc10017-cookie-session] permits a cookie session for an application on one domain with no OAuth. Judgement: the lab runs one instance of each bank, so a store in memory is the simplest, with the same limit as the authorisation sessions. A full store refuses a new session, so that many sign-ins cannot end the sessions of other consumers. A database store needs a migration and a removal job for no lab need. |
| Session ID of the online bank site | A random value with at least 128 bits of entropy. A new ID at each sign-in and sign-up. | [OWASP ASVS 5.0, 7.2.3 and 7.2.4][asvs-v7]. A new ID at each authentication stops session fixation ([OWASP Session Management Cheat Sheet][owasp-session]). |
| Lifetime of a session of the online bank site | 24 hours from the start. No inactivity timeout. No limit to the number of sessions of a customer. | A time limit for security is a time limit of the content, and a limit that is longer than 20 hours needs no warning or extension ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]). Judgement: the lab has no value to protect, and a short timeout only stops a demonstration. [OWASP ASVS 5.0, 7.3.1][asvs-v7] requires an inactivity timeout, so the site does not meet it. The production gap of the sign-in gives this. |
| Sign-out of the online bank site | The data holder ends the session, and the response expires the cookie. No `Clear-Site-Data` header. | [OWASP ASVS 5.0, 7.4.1][asvs-v7]. `Clear-Site-Data: "cookies"` affects the entire registered domain ([MDN: Clear-Site-Data][mdn-clear-site-data]). In the lab, all UIs are on `localhost`, so it also removes the cookie of a CDR authentication flow in progress. |
| CSRF protection of the online bank site | The refusal of cross-origin changes of the listener. No CSRF token. | Fetch Metadata is the primary protection against CSRF for modern browsers, with the check of `Origin` as the fallback ([OWASP CSRF Prevention Cheat Sheet][owasp-csrf]), and `http.CrossOriginProtection` does both ([Go: CrossOriginProtection][go-crossoriginprotection]). The design system supports modern browsers only ([design system DESIGN.md][ui-design]). The check also applies to the sign-in, so login CSRF needs no session before the sign-in. A different lab port is a different origin, so the check also refuses a change from a different lab UI. |
| Separation of the sign-in of the site from the CDR authentication | A note on the sign-in page of the site, different titles and headings, a different origin, no links in either direction, and no change to the authentication and consent pages | The friction standard forbids a reference to a different service in the authentication flow ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Friction), so the note goes on the site. The authentication and consent pages already say that the bank does not ask for a password on the page. The CDS gives the phrasing of the password statement to the data holder ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Passwords). |
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
| Accreditation number on the dashboard | Not shown | The DSB only recommends it ([CDR Support: Dashboards][support-dashboards]). The data holder does not read participant data from the Register (see Decisions: Name of the accredited person on the consent page). No data standard or Register entry is specified for rule 1.15(1)(bb) or rule 7.9(1)(c) in the sources of this document. |
| Ended consents on the dashboard | Shown in a group of their own, with the date of the end, and with no reason and no method. No time limit and no pages. | Rule 1.15(3)(e) requires the time of the end of an authorisation that is not current ([CDR Rules][cdr-rules]). The ACCC says that the dashboard need not show the method or the reason ([CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method], [CDR Support: Dashboards][support-dashboards]). Judgement: the accounts page shows closed accounts in the same way, and a customer has few arrangements. |
| Dashboard of a customer with no open account | The same dashboard, with no special case | The ACCC says that the obligation to give a dashboard ends with the eligibility of the consumer ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). No rule forbids a dashboard after that time. Judgement: the least complex solution, and the consumer sees the consents that ended at the close. |
| Text of the CDS messages | The suggested text of the DSB, word for word, with the brand name of the data recipient | The CDS requires the messages, and permits the data holder to use or paraphrase the suggested text ([CDS: Dashboard Standards][cds-dashboard-standards], [CDS: Withdrawal Standards][cds-withdrawal-standards]). Judgement: the text of the DSB needs no review, and a test can check it. The brand name is the name of the data recipient that the consumer knows best. |
| Confirmation of a withdrawal on the dashboard | Two steps: the action "Stop sharing", then a dialog with the CDS messages and the actions "Stop sharing" and "Cancel" | A withdrawal cannot be reversed ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]), and the CDS messages are "part of the withdrawal process" ([CDS: Withdrawal Standards][cds-withdrawal-standards]). Rule 1.15(1)(e): no more complicated than the authorisation, which has a customer number, an OTP and a confirmation ([CDR Rules][cdr-rules]). The CX Guidelines suggest positive friction and the words "Stop sharing" ([CX Guidelines: Withdrawal][cx-withdrawal]). |
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
