# Mock data holder: Banking APIs

This area gives the customers and accounts of the brand, and the APIs that give the CDR data of an arrangement: Get Customer, the accounts, balances and transactions, UserInfo, and the record of disclosures.

## Customers and accounts

The [operator interface](operator-interface.md) gives the fields of a request that creates customer data. [Consumer authentication](../DESIGN.md#customers-and-accounts) gives the user identifiers of a customer.

- <a id="dh-1000"></a>**DH-1000** The data holder holds only the customer, account and transaction data that the authentication, the authorisation and the banking APIs need. For example, it holds no address and no phone number of a customer. ([Judgement: the fields of a customer, an account and a transaction](operator-interface.md#dh-0771))
- <a id="dh-1001"></a>**DH-1001** A customer is a person. ([CDS: Get Customer][cds-get-customer]; [Judgement: only persons with their own accounts](#dh-1150))
- <a id="dh-1002"></a>**DH-1002** A customer has a customer number, a last name, an optional first name, an email address, the time of its creation and the time of its last change. ([Judgement: the fields of a customer, an account and a transaction](operator-interface.md#dh-0771))
- <a id="dh-1003"></a>**DH-1003** The lab operator [creates a customer with the operator interface](operator-interface.md#dh-0720), or the consumer creates one with a [sign-up on the online bank site](../DESIGN.md#sign-up). Both use the same rules. ([Judgement: the fields of a customer, an account and a transaction](operator-interface.md#dh-0771))
- <a id="dh-1004"></a>**DH-1004** An account has one customer, who is its only owner. ([CDS: BankingAccountV3][cds-schema-account]; [Judgement: only persons with their own accounts](#dh-1150))
- <a id="dh-1005"></a>**DH-1005** An account has these fields. ([CDS: Product Categories][cds-product-categories]; [CDS: BankingAccountV3][cds-schema-account]; [Judgement: the fields of a customer, an account and a transaction](operator-interface.md#dh-0771))
  - a display name
  - a product category of `BankingProductCategoryV2`
  - a product name
  - a BSB of 6 digits
  - an account number of 6 to 10 digits
  - an optional credit limit
  - the time of its creation
  - a status, open or closed
  - the time of its close, when it is closed
- <a id="dh-1006"></a>**DH-1006** An account is open when the lab operator creates it, or when the consumer opens it on the online bank site. ([the opening of an account](../DESIGN.md#opening-of-an-account); [Judgement: only the close of an account changes the customer data](operator-interface.md#dh-0772))
- <a id="dh-1007"></a>**DH-1007** The lab operator, or the consumer on the online bank site, can close an open account. A closed account cannot open again. ([Judgement: only the close of an account changes the customer data](operator-interface.md#dh-0772); [Judgement: a closed account stays in its arrangements](#dh-1160))
- <a id="dh-1008"></a>**DH-1008** A customer is eligible while the customer has an open account. When the customer has no open account, [the arrangements of the customer end](end-of-an-arrangement.md#dh-0918). ([CDR Rules][cdr-rules], rule 1.10B(1)(b))
- <a id="dh-1009"></a>**DH-1009** Each account can be accessed online. ([CDR Rules][cdr-rules], Schedule 3, clause 2.1)
- <a id="dh-1010"></a>**DH-1010** The data holder creates an account identifier for each account: a random RFC 4122 UUID. It identifies the account only in the data holder. A data recipient does not get it. ([Judgement: a random account identifier with no meaning](#dh-1151))
- <a id="dh-1011"></a>**DH-1011** A transaction belongs to one account. It has the fields of `BankingTransactionV2` that [the operator interface gives](operator-interface.md#dh-0748), and optionally the `extendedData` of `BankingTransactionDetailV3`. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail]; [Judgement: the fields of a customer, an account and a transaction](operator-interface.md#dh-0771))
- <a id="dh-1012"></a>**DH-1012** A transaction does not change after the lab operator creates it. To show a posting, the lab operator creates a posted transaction, and the pending transaction stays. ([CDS: BankingTransactionV2][cds-schema-transaction]; [Judgement: no link between a pending and a posted transaction](#dh-1152))
- <a id="dh-1013"></a>**DH-1013** The current balance of an account is the sum of the amounts of its posted transactions. For example, posted amounts of `100.00` and `-30.00` give `70.00`. ([CDS: BankingBalance][cds-schema-balance]; [Judgement: balances from the transactions](#dh-1153))
- <a id="dh-1014"></a>**DH-1014** The available balance of an account is the credit limit, plus the current balance, plus the negative amounts of its pending transactions. When this sum is less than zero, the available balance is zero. ([CDS: BankingBalance][cds-schema-balance]; [Judgement: balances from the transactions](#dh-1153))
- <a id="dh-1015"></a>**DH-1015** The consumer UI shows an account number only as a masked number: `xxx-xxx` for the BSB, a space, and the account number with each digit except the last 4 as the letter `x`. For example, `xxx-xxx xxxx5678`. ([CDS: Common Field Types][cds-common-field-types])
- <a id="dh-1016"></a>**DH-1016** All customer and account data is invented. It does not copy the data of a real person. ([Judgement: no personal information in the lab](#dh-1154))

## Access to the data

The rules of this group are for the APIs on the resource base URI and for the UserInfo endpoint. Registration gives the responses for [a software product that is not ACTIVE](registration.md#dh-0695).

- <a id="dh-1020"></a>**DH-1020** A request has the access token in the `Authorization` header with the scheme `Bearer`. The data holder ignores an access token in the query or in the body. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource]; [RFC 6750, section 2.1][rfc6750-header])
- <a id="dh-1021"></a>**DH-1021** The data holder accepts an access token of an arrangement only when all of these are true. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource]; [CDS: Holder of Key Mechanism][cds-hok]; [the token is valid only with its client certificate](../../../pkg/authserver/design/token-endpoint.md#auth-0055))
  - The token is not expired.
  - The token is not revoked.
  - The arrangement of the token did not end.
  - The token is bound to the client certificate of the request.
- <a id="dh-1022"></a>**DH-1022** A client credentials token gives no access to the CDR data of an arrangement. ([CDS: Authorisation Scopes][cds-authorisation-scopes])
- <a id="dh-1023"></a>**DH-1023** The data holder returns only the data of the customer of the arrangement, of the accounts in the arrangement, and of the scopes of the token. For example, a token with only `bank:accounts.basic:read` gets no transactions. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource])
- <a id="dh-1024"></a>**DH-1024** The data holder checks the access token before the other checks of the request. For example, a request with a refused token and no `x-v` header gets 401. ([CDR Support: Error scenarios and responses][support-error-scenarios]; [the status check after the token check](registration.md#dh-0699))
- <a id="dh-1025"></a>**DH-1025** When the data holder refuses a token, the response is one of the table below. The responses follow RFC 6750. ([RFC 6750, section 3.1][rfc6750-errors])

| ID | Condition | Response | Source |
|---|---|---|---|
| <a id="dh-1026"></a>**DH-1026** | The request has no bearer token | 401 with `WWW-Authenticate: Bearer` | [RFC 6750, section 3.1][rfc6750-errors] |
| <a id="dh-1027"></a>**DH-1027** | The token is not known, is expired, is revoked, is of an arrangement that ended, is not bound to the client certificate, or is a client credentials token | 401 with `WWW-Authenticate: Bearer error="invalid_token"` | [RFC 6750, section 3.1][rfc6750-errors]; [Judgement: invalid_token after the end of an arrangement](#dh-1156) |
| <a id="dh-1028"></a>**DH-1028** | On the resource base URI, the token does not have the scope of the API | 403 with `WWW-Authenticate: Bearer error="insufficient_scope"` and the scope of the API. The body has the CDS error `urn:au-cds:error:cds-all:Authorisation/InvalidConsent`. | [RFC 6750, section 3.1][rfc6750-errors]; [CDS: HTTP Response Codes][cds-http-response-codes]; [CDS: Invalid Consent][cds-error-403-invalid-consent] |
| <a id="dh-1029"></a>**DH-1029** | On the UserInfo endpoint, the token does not have the scope `openid` | 403 with `WWW-Authenticate: Bearer error="insufficient_scope"` and the scope `openid`, with no body | [RFC 6750, section 3.1][rfc6750-errors]; [Judgement: no body for a refused token](#dh-1155) |

- <a id="dh-1030"></a>**DH-1030** A 401 response for a refused token has no body. ([RFC 6750, section 3.1][rfc6750-errors]; [Judgement: no body for a refused token](#dh-1155))
- <a id="dh-1031"></a>**DH-1031** Each successful response on the resource base URI has [the response envelope](../../../pkg/cds/DESIGN.md#cdsl-0034), with [a `links.self`](../../../pkg/cds/DESIGN.md#cdsl-0035). ([CDS: Get Accounts][cds-get-accounts])
- <a id="dh-1032"></a>**DH-1032** A list on the resource base URI uses [the CDS pagination rules](../../../pkg/cds/DESIGN.md#cdsl-0017). For example, Get Accounts. ([CDS: Get Accounts][cds-get-accounts])
- <a id="dh-1033"></a>**DH-1033** The data holder logs the `x-fapi-interaction-id` of each request on the resource base URI and on the UserInfo endpoint. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource]; [the values that telemetry does not contain](non-functional-requirements.md#dh-0008))

## Resource identifiers

- <a id="dh-1040"></a>**DH-1040** An `accountId` or a `transactionId` is a version 8 UUID from a SHA-256 hash of three values: the software product ID, the customer number and the identifier of the account or the transaction in the data holder. ([RFC 9562, section 5.8][rfc9562-v8]; [CDS: ID Permanence][cds-id-permanence]; [Judgement: a hash with no key for the resource identifiers](#dh-1157))
- <a id="dh-1041"></a>**DH-1041** An `accountId` or a `transactionId` has no meaning. It is not an identifier of the operator interface. ([CDS: ID Permanence][cds-id-permanence])
- <a id="dh-1042"></a>**DH-1042** A software product gets the same identifier for an account or a transaction in each arrangement of the customer, also after it registers again. ([CDS: ID Permanence][cds-id-permanence]; [CDR Support: ID Permanence and PPID][support-ppid])
- <a id="dh-1043"></a>**DH-1043** A different software product or a different customer gets a different identifier for the same account or transaction. ([CDS: ID Permanence][cds-id-permanence]; [CDR Support: ID Permanence and PPID][support-ppid])

## Get Customer

- <a id="dh-1050"></a>**DH-1050** Get Customer returns the customer of the arrangement with `customerUType` `person` and a `CommonPerson` object. ([CDS: Get Customer][cds-get-customer])
- <a id="dh-1051"></a>**DH-1051** The person has `lastName`, `firstName` when the customer has one, `middleNames` as an empty list, and `lastUpdateTime`. The `lastUpdateTime` is the time of the last change to the customer. ([CDS: CommonPerson][cds-schema-person]; [Judgement: the fields of a person](#dh-1158))
- <a id="dh-1052"></a>**DH-1052** The person has no `prefix`, no `suffix` and no occupation. ([CDS: CommonPerson][cds-schema-person]; [Judgement: the fields of a person](#dh-1158))

## Accounts

- <a id="dh-1060"></a>**DH-1060** Get Accounts returns the accounts of the arrangement as `BankingAccountV3` records, in the sequence of their creation. ([CDS: Get Accounts][cds-get-accounts])
- <a id="dh-1061"></a>**DH-1061** Get Accounts supports the filters `product-category`, `open-status` (default `ALL`) and `is-owned`. An account must match each filter in the request. For example, `open-status=CLOSED` gives only the closed accounts. ([CDS: Get Accounts][cds-get-accounts])
- <a id="dh-1062"></a>**DH-1062** An account record has these fields: `accountId`, `displayName`, `maskedNumber`, `productCategory`, `productName`, `openStatus`, `isOwned` with the value `true`, `accountOwnership` with the value `ONE_PARTY`, and `creationDate`. ([CDS: BankingAccountV3][cds-schema-account]; [Judgement: the optional fields of an account](#dh-1159))
- <a id="dh-1063"></a>**DH-1063** A closed account stays in the CDR data of an arrangement that has it, with `openStatus` `CLOSED` and its transactions. ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(7); [Judgement: a closed account stays in its arrangements](#dh-1160))
- <a id="dh-1064"></a>**DH-1064** Get Account Detail returns one account of the arrangement as a `BankingAccountDetailV5` record. It adds `bsb` and `accountNumber`, with digits only. It has no other optional field. ([CDS: Get Account Detail][cds-get-account-detail]; [Judgement: the optional fields of an account](#dh-1159))

## Balances

- <a id="dh-1070"></a>**DH-1070** Get Account Balance returns the balance of one account of the arrangement. ([CDS: Get Account Balance][cds-get-account-balance])
- <a id="dh-1071"></a>**DH-1071** Get Bulk Balances returns the balances of the accounts of the arrangement, with [the filters of Get Accounts](#dh-1061). ([CDS: Get Bulk Balances][cds-get-bulk-balances])
- <a id="dh-1072"></a>**DH-1072** Get Balances For Specific Accounts returns the balances of the accounts in `data.accountIds`. ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific])
- <a id="dh-1073"></a>**DH-1073** A balance has `accountId`, `currentBalance`, `availableBalance` and `creditLimit`, from [the balances of the account](#dh-1013). It has no `amortisedLimit`, no `currency` and no `purses`, so the currency is the default `AUD`. ([CDS: BankingBalance][cds-schema-balance])
- <a id="dh-1074"></a>**DH-1074** Get Balances For Specific Accounts gives one balance for each different `accountId` in `data.accountIds`, in the sequence of Get Accounts, with pagination. For example, an `accountId` that is in the list two times gives one balance. ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]; [Judgement: the accounts of Get Balances For Specific Accounts](#dh-1161))
- <a id="dh-1075"></a>**DH-1075** When `data.accountIds` is empty, Get Balances For Specific Accounts gives an empty list. ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]; [Judgement: the accounts of Get Balances For Specific Accounts](#dh-1161))

## Transactions

- <a id="dh-1080"></a>**DH-1080** Get Transactions For Account returns the transactions of one account of the arrangement as `BankingTransactionV2` records. ([CDS: Get Transactions For Account][cds-get-transactions])
- <a id="dh-1081"></a>**DH-1081** The response of Get Transactions For Account has [`meta.isQueryParamUnsupported` with the value `false`](../../../pkg/cds/DESIGN.md#cdsl-0032). ([CDS: Get Transactions For Account][cds-get-transactions])
- <a id="dh-1082"></a>**DH-1082** The effective time of a transaction is its `postingDateTime` when it has one, and its `executionDateTime` when it does not. ([CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: a required execution time for a pending transaction](#dh-1162))
- <a id="dh-1083"></a>**DH-1083** The list is in descending order of the effective time. Transactions with the same effective time are in descending order of their creation. ([CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: the creation as the second order of the transactions](#dh-1163))
- <a id="dh-1084"></a>**DH-1084** The filters are `oldest-time`, `newest-time`, `min-amount`, `max-amount` and `text`. A transaction must match each filter in the request. ([CDS: Get Transactions For Account][cds-get-transactions])
- <a id="dh-1085"></a>**DH-1085** When `newest-time` is absent, it is the time of the request. When `oldest-time` is absent, it is `newest-time` minus 90 days. ([CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: the time of the request as the default newest-time](#dh-1164))
- <a id="dh-1086"></a>**DH-1086** The time filters include their limits, and they use the effective time. ([CDS: Get Transactions For Account][cds-get-transactions])
- <a id="dh-1087"></a>**DH-1087** The `text` filter gives the transactions whose `description` or `reference` contains the value, with no difference between upper case and lower case. For example, `text=coffee` gives a transaction with the description "Corner Coffee". ([CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: support for the text filter](#dh-1165))
- <a id="dh-1088"></a>**DH-1088** The data holder gives only the transactions whose effective time is not more than 2 years before the time of the request. ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(6)(a); [Judgement: transactions of 2 years at most](#dh-1166))
- <a id="dh-1089"></a>**DH-1089** An `oldest-time` more than 2 years before the time of the request is not an error. ([CDR Support: Error scenarios and responses][support-error-scenarios]; [Judgement: transactions of 2 years at most](#dh-1166))
- <a id="dh-1090"></a>**DH-1090** A transaction has `isDetailAvailable` with the value `true` only when it has `extendedData`. ([CDS: BankingTransactionV2][cds-schema-transaction]; [Judgement: detail only for a transaction with extended data](#dh-1167))
- <a id="dh-1091"></a>**DH-1091** Get Transaction Detail returns one transaction of the account as a `BankingTransactionDetailV3` record, with its `extendedData`. [The 2-year limit](#dh-1088) applies. ([CDS: Get Transaction Detail][cds-get-transaction-detail]; [the error for a transaction with no detail](api-standards.md#dh-0448))

## UserInfo

- <a id="dh-1100"></a>**DH-1100** The UserInfo endpoint accepts `GET` and `POST`. ([CDS: UserInfo endpoint][cds-userinfo-endpoint]; [OpenID Connect Core, section 5.3][oidc-core-userinfo]; [Judgement: a UserInfo endpoint with plain JSON](#dh-1168))
- <a id="dh-1101"></a>**DH-1101** The UserInfo response has `sub`. The `sub` is the `sub` of the ID token of the arrangement. ([OpenID Connect Core, section 5.3][oidc-core-userinfo]; [CDS: Scopes and Claims][cds-scopes-and-claims])
- <a id="dh-1102"></a>**DH-1102** For a token with the scope `profile`, the UserInfo response also has `name`, `family_name` and `updated_at`, and `given_name` when the customer has a first name. ([CDS: Scopes and Claims][cds-scopes-and-claims])
- <a id="dh-1103"></a>**DH-1103** The UserInfo response has no `auth_time`. ([CDS: Scopes and Claims][cds-scopes-and-claims])
- <a id="dh-1104"></a>**DH-1104** The UserInfo response is JSON. It is not signed and not encrypted. ([OpenID Connect Core, section 5.3.2][oidc-core-userinfo-response]; [Judgement: a UserInfo endpoint with plain JSON](#dh-1168))
- <a id="dh-1105"></a>**DH-1105** The UserInfo endpoint needs a token with the scope `openid`. ([OpenID Connect Core, section 5.3][oidc-core-userinfo])

## Record of disclosures

- <a id="dh-1110"></a>**DH-1110** The data holder records each disclosure of CDR data. A disclosure is a response with a 2xx status of an API on the resource base URI or of the UserInfo endpoint, for an access token of an arrangement. ([CDR Rules][cdr-rules], rule 7.9(1) and rule 9.3(1)(d); [Judgement: one record for each response](#dh-1170))
- <a id="dh-1111"></a>**DH-1111** A record has the `cdr_arrangement_id`, the data cluster and the time of the response. ([CDR Rules][cdr-rules], rule 9.3(1)(d); [Judgement: one record for each response](#dh-1170))
- <a id="dh-1112"></a>**DH-1112** The record identifies the data cluster with the scope of the API. The data cluster is the data cluster of the scope in [the scope table of the authorisation page](authorisation.md#dh-0805). ([CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters]; [Judgement: the scope of the API as the data cluster](#dh-1171))
- <a id="dh-1113"></a>**DH-1113** For an API with the scope `bank:accounts.detail:read`, the data cluster depends on the scopes of the arrangement, as on the authorisation page. For example, with `bank:accounts.basic:read` it is [Account numbers and features](authorisation.md#dh-0807). ([CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters]; [Judgement: the scope of the API as the data cluster](#dh-1171))
- <a id="dh-1114"></a>**DH-1114** For the UserInfo endpoint, the data cluster is [the data cluster of `profile`](authorisation.md#dh-0810). ([CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters])
- <a id="dh-1115"></a>**DH-1115** A UserInfo response for a token with no `profile` scope gives no record. ([CDS: Data Language Standards][cds-data-language]; [Judgement: no record for a UserInfo response with no profile scope](#dh-1174))
- <a id="dh-1116"></a>**DH-1116** A 2xx response with no records is a disclosure. For example, a response with no transactions in the period. ([CDR Rules][cdr-rules], rule 9.3(1)(d); [Judgement: a response with no records is a disclosure](#dh-1173))
- <a id="dh-1117"></a>**DH-1117** The data holder writes the record before it sends a byte of the response. ([CDR Rules][cdr-rules], rule 9.3(1)(d); [Judgement: the record before the response](#dh-1172))
- <a id="dh-1118"></a>**DH-1118** When the data holder cannot write the record, the response is 500 with no CDR data. ([CDR Rules][cdr-rules], rule 9.3(1)(d); [Judgement: the record before the response](#dh-1172))
- <a id="dh-1119"></a>**DH-1119** A response that is not 2xx gives no record. ([CDR Rules][cdr-rules], rule 9.3(1)(d))
- <a id="dh-1120"></a>**DH-1120** A response to `HEAD` gives no record, because it has no body. ([RFC 9110, section 9.3.2][rfc9110-head])
- <a id="dh-1121"></a>**DH-1121** The [consumer dashboard](consumer-dashboard.md#dh-1919) shows a disclosure at the next read after the response. ([CDR Rules][cdr-rules], rule 7.9(1))
- <a id="dh-1122"></a>**DH-1122** The data holder does not remove a record of a disclosure. ([CDR Rules][cdr-rules], rule 9.3(5); [Judgement: no removal of the records of disclosures](#dh-1175))

## Judgements

### <a id="dh-1150"></a>DH-1150 Only persons with their own accounts

- Kind: choice
- Basis: Step 3.3. The CDS also gives organisations, joint accounts, secondary users and nominated representatives ([CDS: Authorisation Standards][cds-authorisation-standards]).
- Reason: One consumer with their own accounts proves the authorisation flow. The other kinds add rules for profiles, approvals and notices, and they do not change the protocol.

### <a id="dh-1151"></a>DH-1151 A random account identifier with no meaning

- Kind: choice
- Basis: Step 3.3. No source gives the identifier of an account in a data holder.
- Reason: The authorisation page and the arrangement need a value for each account, and an identifier with no meaning keeps the account number out of a form and an arrangement. It is not the `accountId` of the banking APIs, which follows [the resource identifier rules](#dh-1040). The operator interface gives it, so that a lab user can find the accounts of an arrangement.

### <a id="dh-1152"></a>DH-1152 No link between a pending and a posted transaction

- Kind: choice
- Basis: Step 3.3. The CDS does not require a link between a pending transaction and its posted transaction ([CDS: BankingTransactionV2][cds-schema-transaction]).
- Reason: It is the least complex model.

### <a id="dh-1153"></a>DH-1153 Balances from the transactions

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for a balance. It says that the balance is the same as the balance of the other channels of the data holder ([CDS: Data Latency][cds-data-latency]).
- Reason: Balances that agree with the transactions let a data recipient test both. The lab operator does not have to keep two values the same.

### <a id="dh-1154"></a>DH-1154 No personal information in the lab

- Kind: choice
- Basis: Step 3.1. The CDR protects the data of real consumers, and the lab has no real consumers.
- Reason: The lab must hold no personal information. Invented data has no privacy risk.

### <a id="dh-1155"></a>DH-1155 No body for a refused token

- Kind: choice
- Basis: Step 3.2. RFC 6750 gives the `WWW-Authenticate` values and no body ([RFC 6750, section 3.1][rfc6750-errors]). The CDS requires an error body only for a 403 for a scope that is not correct ([CDS: HTTP Response Codes][cds-http-response-codes]).
- Reason: The header gives the client all the information. Registration management has the same 401 responses.

### <a id="dh-1156"></a>DH-1156 invalid_token after the end of an arrangement

- Kind: conflict
- Basis: Step 1. The CDS lists 403 `Authorisation/RevokedConsent` for an arrangement that ended, as a SHOULD ([CDS: Consent Is Revoked][cds-error-403-revoked-consent]). The CDS also requires that the end revokes the tokens ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]), and FAPI requires the RFC 6750 errors, which give 401 `invalid_token` for a revoked token ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource]; [CDR Support: Error scenarios and responses][support-error-scenarios]).
- Reason: The token of an ended arrangement is a revoked token, so the MUST rules apply.

### <a id="dh-1157"></a>DH-1157 A hash with no key for the resource identifiers

- Kind: choice
- Basis: Step 3.3. The CDS forbids an identifier with meaning and an identifier that two software products share, and it names a hash for a `transactionId` ([CDS: ID Permanence][cds-id-permanence]; [CDS: BankingTransactionV2][cds-schema-transaction]). The ACCC suggests an identifier for the subject, or for the subject and the client ([CDR Support: ID Permanence and PPID][support-ppid]).
- Reason: A hash needs no stored value for each pair of resource and software product, and one method for accounts and transactions is one rule to keep correct. The internal identifiers are random UUIDs that no data recipient gets, so a party cannot predict or reverse the hash, and a key adds nothing. A new registration keeps the software product ID, so the identifiers stay the same.

### <a id="dh-1158"></a>DH-1158 The fields of a person

- Kind: choice
- Basis: Step 3.3. Only `lastName` and `middleNames` are required, and the occupation is optional ([CDS: CommonPerson][cds-schema-person]).
- Reason: The lab needs no more personal data.

### <a id="dh-1159"></a>DH-1159 The optional fields of an account

- Kind: choice
- Basis: Step 3.3. The CDS marks the other fields of an account and its detail as optional ([CDS: BankingAccountV3][cds-schema-account]; [CDS: Get Account Detail][cds-get-account-detail]).
- Reason: The account has no other data. The rates, fees and features of a product are in the product reference data.

### <a id="dh-1160"></a>DH-1160 A closed account stays in its arrangements

- Kind: choice
- Basis: Step 3.3. The data of a closed account is voluntary consumer data ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(7); [CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]).
- Reason: A data recipient can then test `openStatus` and its filter. A consumer selects an account so that its data is shared, and the arrangements end when the customer has no open account.

### <a id="dh-1161"></a>DH-1161 The accounts of Get Balances For Specific Accounts

- Kind: choice
- Basis: Step 3.3. The schema permits an empty `accountIds`, and it gives no rule for a duplicate or for the sequence ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]).
- Reason: One record for each account keeps `totalRecords` correct. One sequence for the three balance APIs gives stable pages.

### <a id="dh-1162"></a>DH-1162 A required execution time for a pending transaction

- Kind: choice
- Basis: Step 3.3. The CDS uses the posting time, then the execution time, then a time that the data holder selects ([CDS: Get Transactions For Account][cds-get-transactions]).
- Reason: The lab operator must give the `executionDateTime` of a pending transaction, so the third case does not occur.

### <a id="dh-1163"></a>DH-1163 The creation as the second order of the transactions

- Kind: choice
- Basis: Step 3.3. The CDS gives only the order of the effective time ([CDS: Get Transactions For Account][cds-get-transactions]).
- Reason: Pagination needs a stable order.

### <a id="dh-1164"></a>DH-1164 The time of the request as the default newest-time

- Kind: choice
- Basis: Step 3.3. The CDS gives "today" as the default, and it does not give the time ([CDS: Get Transactions For Account][cds-get-transactions]).
- Reason: The time of the request includes each transaction until then, and it needs no time zone.

### <a id="dh-1165"></a>DH-1165 Support for the text filter

- Kind: choice
- Basis: Step 3.3. The CDS makes the `text` filter optional for a data holder ([CDS: Get Transactions For Account][cds-get-transactions]).
- Reason: A substring match is simple, and a data recipient can then test the filter.

### <a id="dh-1166"></a>DH-1166 Transactions of 2 years at most

- Kind: choice
- Basis: Step 3.3. Older transactions are not required consumer data ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(6)(a)). The ACCC permits an error for an `oldest-time` that is older, and does not require it ([CDR Support: Error scenarios and responses][support-error-scenarios]).
- Reason: The data holder shares only required data. A refusal of an older `oldest-time` gives a data recipient no benefit.

### <a id="dh-1167"></a>DH-1167 Detail only for a transaction with extended data

- Kind: choice
- Basis: Step 3.3. The detail object is required, and a transaction with no detail gets 404 ([CDR Support: Error scenarios and responses][support-error-scenarios]).
- Reason: The lab operator can then make each case that a data recipient must handle.

### <a id="dh-1168"></a>DH-1168 A UserInfo endpoint with plain JSON

- Kind: conflict
- Basis: Step 1. The CDS says that data holders MUST support a UserInfo endpoint ([CDS: UserInfo endpoint][cds-userinfo-endpoint]). Its list of configuration fields says that `userinfo_endpoint` is necessary only for the hybrid flow ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]).
- Reason: The first rule is a direct requirement, and the hybrid flow is retired. OpenID Connect permits plain JSON when the registration does not ask for a signed response ([OpenID Connect Core, section 5.3.2][oidc-core-userinfo-response]).

### <a id="dh-1170"></a>DH-1170 One record for each response

- Kind: choice
- Basis: Step 3.3. The dashboard must show what CDR data was disclosed and when, and the data holder must keep records of disclosures ([CDR Rules][cdr-rules], rule 1.15(3)(f), rule 7.9(1) and rule 9.3(1)(d)). The ACCC says that the dashboard shows the first and the last date for each data cluster ([CDR Support: Privacy Safeguard 5 - dates on the consumer dashboards][support-dashboard-dates]; [CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters]).
- Reason: One record for each response meets the three rules, and the dashboard gets the first and the last date from the records. A disclosure does not change the state of an arrangement, so it is a record and not an event, as a transaction is ([ADR 0003][adr-0003]).

### <a id="dh-1171"></a>DH-1171 The scope of the API as the data cluster

- Kind: choice
- Basis: Step 3.3. The ACCC says that the dashboard shows the sharing for each data cluster ([CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters]), and gives no form of the record.
- Reason: The scope is stable, and for an arrangement it selects one data cluster. The language of a data cluster can change with the CDS.

### <a id="dh-1172"></a>DH-1172 The record before the response

- Kind: choice
- Basis: Step 3.3. Rule 9.3(1)(d) requires a record of each disclosure ([CDR Rules][cdr-rules]), and no source gives the time of the record.
- Reason: A response that left before its record can have no record. A response is at most one CDS page, so to keep it in memory is a small cost.

### <a id="dh-1173"></a>DH-1173 A response with no records is a disclosure

- Kind: choice
- Basis: Step 3.3. Rule 9.3(1)(d) does not say if an empty response is a disclosure ([CDR Rules][cdr-rules]).
- Reason: The data recipient learns that the data holder has no data of that data cluster for the period. The data holder cannot know which responses the data recipient uses.

### <a id="dh-1174"></a>DH-1174 No record for a UserInfo response with no profile scope

- Kind: choice
- Basis: Step 3.3. The `sub` is in no data cluster of the CDS ([CDS: Data Language Standards][cds-data-language]).
- Reason: The dashboard has no data cluster to show for the response.

### <a id="dh-1175"></a>DH-1175 No removal of the records of disclosures

- Kind: choice
- Basis: Step 3.3. Rule 9.3(5) requires that the data holder keeps the records for 6 years ([CDR Rules][cdr-rules]).
- Reason: A record that stays longer meets the rule, and the other records of the data holder also stay. A production service can operate with it, so it is not a production gap.

[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-authorisation-scopes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-scopes
[cds-authorisation-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-standards
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-data-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-language-standards-common
[cds-data-latency]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-latency
[cds-error-403-invalid-consent]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-invalid-consent
[cds-error-403-revoked-consent]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-revoked-consent
[cds-get-account-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-balance
[cds-get-account-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-detail
[cds-get-accounts]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-accounts
[cds-get-balances-specific]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-balances-for-specific-accounts
[cds-get-bulk-balances]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-bulk-balances
[cds-get-customer]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-customer
[cds-get-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transaction-detail
[cds-get-transactions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[cds-http-response-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-response-codes
[cds-id-permanence]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#id-permanence
[cds-oidc-config]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#openid-provider-configuration-endpoint
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[cds-schema-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingaccountv3
[cds-schema-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingbalance
[cds-schema-person]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_schemas_tocScommonperson
[cds-schema-transaction]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactionv2
[cds-schema-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingtransactiondetailv3
[cds-scopes-and-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#scopes-and-claims
[cds-userinfo-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#userinfo-endpoint
[fapi-baseline-resource]: https://openid.net/specs/openid-financial-api-part-1-1_0.html#protected-resources-provisions
[oidc-core-userinfo]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
[oidc-core-userinfo-response]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc6750-header]: https://www.rfc-editor.org/rfc/rfc6750#section-2.1
[rfc9110-head]: https://www.rfc-editor.org/rfc/rfc9110#section-9.3.2
[rfc9562-v8]: https://www.rfc-editor.org/rfc/rfc9562#section-5.8
[support-closed-accounts]: https://cdr-support.zendesk.com/hc/en-us/articles/14519397304079
[support-dashboard-clusters]: https://cdr-support.zendesk.com/hc/en-us/articles/4920717853327
[support-dashboard-dates]: https://cdr-support.zendesk.com/hc/en-us/articles/5333587630095
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[support-ppid]: https://cdr-support.zendesk.com/hc/en-us/articles/5505351487631
