# CDS banking and common APIs: rules for a mock banking data holder

- Research date: 2026-10-06
- Consumer Data Standards (CDS) version: 1.36.0 (released 04/12/2025). This is the current version at the research date. The live site states "version 1.36.0", the newest release tag in the standards repository is `v1.36.0`, and the 1.36.0 archive URL redirects to the live site. No release is newer. [S1 #version] [S2 releases]
- CDR Rules version: Competition and Consumer (Consumer Data Right) Rules 2020, Compilation No. 10, compilation date 4 March 2025. This is the latest compilation on the Federal Register of Legislation at the research date. This note uses the rules only for the data that is in scope, because the CDS does not give those rules.
- Purpose: source-cited facts for requirements (functional and non-functional) for the first mock banking data holder: product reference data, the common discovery APIs, and the customer, account, balance and transaction APIs. These notes do not design an implementation. They do not cover the information security profile, consent or Dynamic Client Registration.

Check the sources again before you depend on a fact; the rules and standards change.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site; the archive URL redirects to it while 1.36.0 is current) | https://consumerdatastandardsaustralia.github.io/standards/ and https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/ |
| S2 | CDS source repository at tag `v1.36.0` (OpenAPI files `cds_banking.json`, `cds_common.json`, `cds_register.json`; release list) | https://github.com/ConsumerDataStandardsAustralia/standards/tree/v1.36.0/docs/includes/swagger and https://github.com/ConsumerDataStandardsAustralia/standards/releases |
| S3 | CDS Endpoint Version Schedule | https://consumerdatastandardsaustralia.github.io/standards/includes/endpoint-version-schedule/ |
| S4 | CDS release notes 1.35.0 and 1.36.0 | https://consumerdatastandardsaustralia.github.io/standards/includes/releasenotes/releasenotes.1.35.0.html and https://consumerdatastandardsaustralia.github.io/standards/includes/releasenotes/releasenotes.1.36.0.html |
| S5 | CDS pages for superseded endpoint versions (for example Get Accounts v2) | https://consumerdatastandardsaustralia.github.io/standards/includes/obsolete/get-accounts-v2.html (the other pages are listed in section 12) |
| S6 | Decision 376, White label brand arrangements (28 November 2025): issue and decision PDF | https://github.com/ConsumerDataStandardsAustralia/standards/issues/376 and https://github.com/user-attachments/files/23812777/Decision.376.-.White.label.brand.arrangements.-.Data.Standards.pdf |
| S7 | CDR Rules 2020, Compilation No. 10 | https://www.legislation.gov.au/F2020L00094/latest/text |
| S8 | CDR Support Portal articles (IDs given for each fact; read through the help centre API) | https://cdr-support.zendesk.com/hc/en-us/articles/<id> |
| S9 | Standards maintenance issues (numbers given for each fact) | https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/<number> |
| S10 | ACCC Developer Portal: Register API Change History and Delivery Dates | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/42827777 and https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081043 |

Notation:
- "MUST", "SHOULD" and "MAY" in quotes are normative words from the source.
- **UNVERIFIED** means that no primary source confirms the fact. Do not depend on it.
- **ARCHIVED** means that the source is old. A newer source replaces it, or the fact may be out of date.
- A citation `[S2 cds_banking.json <path>]` gives a `jq` path in the OpenAPI file. A citation `[S1 #anchor]` gives an anchor on the live site.
- The source order is: CDS text and OpenAPI, then first-party ACCC and DSB guidance. S8 is guidance. When S8 does not agree with S1 or S2, section 10 records the conflict.

---

## 1. Versions and dates that affect the project

### 1.1 Changes after CDS 1.36.0

- There are none. CDS 1.36.0 is the current version. [S1 #version] [S2 releases]
- CDS 1.36.0 contains only Decision 376 (white label brand arrangements). For the endpoints in scope, it adds the optional `brandGroup` field to Get Products and Get Product Detail. It does not change the endpoint versions. [S4 1.36.0] [S6]
- The standards repository has no published decision that is later than Decision 376. Consultation Draft 374 (minimum baseline authentication) is an open placeholder. [S2 issues] The content of the next release is **UNVERIFIED**.

### 1.2 Future dated obligations in CDS 1.36.0

CDS 1.36.0 gives endpoint versions that are not binding at the research date. This table shows the endpoints in scope. [S1 #future-dated-obligations] [S3]

| Endpoint | Version | Data holders "MUST implement" by | Previous version | Data holders "MAY retire" the previous version from |
|---|---|---|---|---|
| Get Products | 5 | 13 July 2026 (past) | 4 | 10 August 2026 (past) |
| Get Product Detail | 7 | 13 July 2026 (past) | 6 | 10 August 2026 (past) |
| Get Accounts | 3 | 9 November 2026 | 2 | 7 December 2026 |
| Get Account Detail | 5 | 9 November 2026 | 4 | 7 December 2026 |
| Get Bulk Balances | 2 | 9 November 2026 | 1 | 7 December 2026 |
| Get Transactions For Account | 2 | 9 November 2026 | 1 | 7 December 2026 |
| Get Transaction Detail | 3 | 9 November 2026 | 2 | 7 December 2026 |

- At the research date, the binding versions of the five authenticated endpoints above are the previous versions. The OpenAPI files at `v1.36.0` describe only the new versions. The CDS calls the previous versions "obsolete" and keeps one page for each. [S2 cds_banking.json `.paths[].get["x-version"]`] [S5]
- Between 9 November 2026 and 7 December 2026, a data holder must support the two versions of each endpoint. After 7 December 2026, the data holder can support only the new version. [S1 #future-dated-obligations]
- CDS 1.35.0 made these versions. The changes came from Decision 367 (March 2025 Rules, non-bank lending and buy now pay later). The release notes say: "Updated affected Banking endpoints to accommodate BNPL products and features, and added '400 - Missing Required Header' error response". [S4 1.35.0]
- The superseded Get Accounts v2 page uses the schema `BankingProductCategory` and does not contain the value `BUY_NOW_PAY_LATER`. [S5 get-accounts-v2] The full field differences between each previous version and its new version are **UNVERIFIED**. I did not compare the schemas.

---

## 2. Base paths and transport

### 2.1 URI structure

- The URI structure "MUST" be: `https://<holder-path>/cds-au/<version>/(<industry>|<HID>)/<resource>`. `<version>` is `v1`, the major version of the standards. It is not the endpoint version. [S1 #uri-structure]
- `<industry>` values: `banking`, `energy`, `telco`, `common`. `banking` is for banking and non-bank lending. "Unless noted otherwise", CDS statements about "banking", "bank account" and "bank transaction" apply to the two sectors. [S1 #uri-structure] [S1 #high-level-standards_definitions_sectors]
- The base path is the URL up to and including `<industry>`. [S1 #uri-base-path]
- "While all authenticated endpoints must be accessible under the same holder path the data holder may stipulate a different holder path for unauthenticated endpoints." [S1 #uri-structure]
- The discovery paths in the OpenAPI file are `/discovery/status` and `/discovery/outages`, below `/cds-au/v1`. They do not have the `common` industry component. The customer paths are `/common/customer` and `/common/customer/detail`. [S2 cds_common.json `.paths`]

### 2.2 Base URIs and the Register fields

A data holder brand registers its base URIs with the Register. [S1 #participant-endpoints]

| Base URI (CDS) | Register field | Transport | Endpoints in scope |
|---|---|---|---|
| PublicBaseUri | `publicBaseUri` (required) | TLS | Get Status, Get Outages |
| ProductBaseUri | `productBaseUri` (optional) | TLS | Get Products, Get Product Detail |
| ResourceBaseUri | `resourceBaseUri` (required) | MTLS | Get Customer, Get Customer Detail, all authenticated banking endpoints |

[S1 #participant-endpoints] [S2 cds_register.json `.components.schemas.RegisterDataHolderBrandServiceEndpointV2`]

- PublicBaseUri: "This should encompass endpoints not requiring authentication that do not have a dedicated base URI." [S1 #participant-endpoints]
- ProductBaseUri: "Base URI for the Consumer Data Standard product reference data endpoints. MAY be the same as the publicBaseUri." Decision 376 added this row in CDS 1.36.0. [S1 #participant-endpoints] [S6 section 3.1.2]
- Register `productBaseUri`: "If `null` or not present, indicates that PRD endpoints are not available." The field is also in Get Data Holder Brands Summary, which is public. [S2 cds_register.json `.components.schemas.DataHolderBrandSummaryV2`]
- ResourceBaseUri "MUST encompass all CDS resource endpoints requiring authentication". [S1 #participant-endpoints]
- The registered value is the holder path only. It does not include `cds-au`, `v1`, the industry or the resource. Get Products is at `{productBaseUri}/cds-au/v1/banking/products`. [S8: 16505555811855]
- The ACCC Register added the Decision 376 changes to Get Data Holder Brands V3 and Get Data Holder Brands Summary V2 on 11 March 2026. [S10 Change History] The date comes from a machine summary of that page, and the Delivery Dates page does not give it: **UNVERIFIED**.
- Changes to Register APIs and certificates for Decision 376 "do not have FDOs specified in the standards". The ACCC manages their delivery. [S6 section 2]

### 2.3 Transport rules

- "All HTTP calls MUST be made using HTTPS incorporating TLS >= 1.2." [S1 #transaction-security]
- TLS endpoints "MUST be configured with a certificate issued by a public CA accepted by major web browsers". Public endpoints and TLS endpoints "MUST NOT use" MTLS. [S1 #participant-endpoints] [S1 #transaction-security]
- MTLS endpoints use certificates from the CDR CA for client and server. "resource requests MUST be validated to ensure the client certificate and access token match." [S1 #transaction-security] [S1 #holder-of-key-mechanism]
- In the OpenAPI files, Get Products, Get Product Detail, Get Status and Get Outages have the server `https://tls.dh.example.com/cds-au/v1` (TLS). All other operations use the default server `https://mtls.dh.example.com/cds-au/v1` (MTLS). [S2 cds_banking.json and cds_common.json `.servers`, `.paths[].get.servers`]

### 2.4 CORS

- "Cross-origin resource sharing (CORS) protections must be disabled (i.e. Access-Control-Allow-Origin set to *) for all unauthenticated endpoints unless specifically stated otherwise in these standards or in a normative reference." [S1 #cors]
- DSB guidance for unauthenticated endpoints: [S8: 900003054706]
  - Support the `OPTIONS` method for the preflight request.
  - The `OPTIONS` response has at a minimum `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods: GET` and `Access-Control-Allow-Headers: x-min-v, x-v, Cache-Control, x-fapi-interaction-id`. It can have `Access-Control-Max-Age`.
  - The `GET` response has at a minimum `x-v`, `Access-Control-Allow-Origin: *` and `Access-Control-Expose-Headers: x-v, Retry-After, x-fapi-interaction-id`.

---

## 3. Product reference data (PRD)

### 3.1 Endpoints

| Endpoint | Method and path | Version | Superseded versions (retirement date) | Transport | Auth | Scope |
|---|---|---|---|---|---|---|
| Get Products | `GET /banking/products` | 5 | 4 (2026-08-10), 3 (2026-05-11), 2 (2021-05-31), 1 (2020-09-29) | TLS | none | none |
| Get Product Detail | `GET /banking/products/{productId}` | 7 | 6 (2026-08-10), 5 (2026-05-11), 4 (2025-11-10), 3 (2023-02-28), 2 (2021-05-31), 1 (2020-09-29) | TLS | none | none |

[S2 cds_banking.json `.paths["/banking/products"].get`, `.paths["/banking/products/{productId}"].get`] [S3]

- The two operations have no `x-scopes` value. "Access to PRD does not require authentication." [S2] [S8: 6260497948431]
- Request headers: `x-v` (mandatory) and `x-min-v` (optional) only. Response headers in the specification: `x-v` only. [S2 cds_banking.json `.paths["/banking/products"].get.parameters`, `.responses["200"].headers`]

### 3.2 Get Products query parameters

| Parameter | Type | Default | Rule |
|---|---|---|---|
| `effective` | Enum `ALL`, `CURRENT`, `FUTURE` | `CURRENT` | Filters on whether the current time is in the period from `effectiveFrom` to `effectiveTo`. |
| `updated-since` | DateTimeString | none (all products) | "Only include products that have been updated after the specified date and time." |
| `brand` | string | none | "Filter results based on a specific brand." |
| `product-category` | `BankingProductCategoryV2` | none (all) | One value only. |
| `page` | PositiveInteger | 1 | Standard pagination. |
| `page-size` | PositiveInteger | 25 | Standard pagination. Maximum 1000 (section 7.4). |

[S2 cds_banking.json `.paths["/banking/products"].get.parameters`]

- "the results returned by this endpoint are expected to be ordered in descending order according to _lastUpdated_". [S2 cds_banking.json `.paths["/banking/products"].get.description`]
- Get Products describes "products that are currently openly offered to the market". [S2]
- The CDS does not say how `brand` is compared with the `brand` field (case, exact or partial match). This is **UNVERIFIED**.
- Get Product Detail has one path parameter, `productId`. [S2]

### 3.3 Product categories

`BankingProductCategoryV2` has 13 values. Accounts and products use the same enumeration. [S2 cds_banking.json `.components.schemas.BankingProductCategoryV2`] [S1 #product-categories]

| Group | Values |
|---|---|
| Deposit products | `REGULATED_TRUST_ACCOUNTS`, `TERM_DEPOSITS`, `TRANS_AND_SAVINGS_ACCOUNTS`, `TRAVEL_CARDS` |
| Lending products | `BUSINESS_LOANS`, `BUY_NOW_PAY_LATER`, `CRED_AND_CHRG_CARDS`, `LEASES`, `MARGIN_LOANS`, `OVERDRAFTS`, `PERS_LOANS`, `RESIDENTIAL_MORTGAGES`, `TRADE_FINANCE` |

### 3.4 Schema structure

Get Products returns `ResponseBankingProductListV4`: `data.products[]` (`BankingProductV6`), `links` (`LinksPaginated`), `meta` (`MetaPaginated`). All three root fields are required. Get Product Detail returns `ResponseBankingProductByIdV7`: `data` (`BankingProductDetailV7`) and `links` are required, `meta` is optional. [S2 cds_banking.json `.components.schemas`]

`BankingProductV6`: [S2 cds_banking.json `.components.schemas.BankingProductV6`]

| Field | Type | Required | Notes |
|---|---|:-:|---|
| `productId` | ASCIIString | yes | "must be unique to a product but does not otherwise need to adhere to ID permanence guidelines" |
| `lastUpdated` | DateTimeString | yes | Last change, or the creation date |
| `productCategory` | Enum | yes | Section 3.3 |
| `name`, `description` | string | yes | |
| `brand` | string | yes | "A label of the brand for the product. Able to be used for filtering. For data holders with single brands this value is still required." |
| `isTailored` | Boolean | yes | The data holder expects to negotiate fees and prices |
| `effectiveFrom`, `effectiveTo` | DateTimeString | no | The period in which the product is available |
| `brandName` | string | no | "An optional display name of the brand." |
| `brandGroup` | ASCIIString | no | Added in CDS 1.36.0. "The _brandGroup_ of a consumer data sharing brand that the product is associated with." |
| `applicationUri` | URIString | no | |
| `additionalInformation` | object | no | `overviewUri`, `termsUri`, `eligibilityUri`, `feesAndPricingUri`, `bundleUri`, and an `additional...Uris` array for each. Each primary URI is mandatory if its array has one or more items. |
| `cardArt[]` | array | no | `cardScheme` (`AMEX`, `EFTPOS`, `MASTERCARD`, `VISA`, `OTHER`), `cardType` (`CHARGE`, `CREDIT`, `DEBIT`) and `imageUri` are required in each item |

`BankingProductDetailV7` is `BankingProductV6` plus these fields. All are optional. [S2 cds_banking.json `.components.schemas.BankingProductDetailV7_allOf`]

| Field | Item schema | Required fields of each item | Type enumeration |
|---|---|---|---|
| `bundles[]` | `BankingProductBundle` | `name`, `description` | none |
| `features[]` | `BankingProductFeatureV4` | `featureType` | 31 values, for example `CARD_ACCESS`, `OFFSET`, `REDRAW`, `OTHER` |
| `constraints[]` | `BankingProductConstraintV3` | `constraintType` | `MAX_BALANCE`, `MAX_LIMIT`, `MAX_LVR`, `MIN_BALANCE`, `MIN_LIMIT`, `MIN_LVR`, `OPENING_BALANCE`, `OTHER` |
| `eligibility[]` | `BankingProductEligibilityV2` | `eligibilityType` | 12 values, for example `NATURAL_PERSON`, `MIN_AGE`, `OTHER` |
| `fees[]` | `BankingProductFeeV2` | `name`, `feeType`, `feeMethodUType` | `feeType` has 17 values. `feeMethodUType` is `fixedAmount`, `rateBased` or `variable`, and the object with the same name is then mandatory. |
| `depositRates[]` | `BankingProductDepositRateV2` | `depositRateType`, `rate`, `applicationType` | `BONUS`, `BUNDLE_BONUS`, `FIXED`, `FLOATING`, `INTRODUCTORY`, `MARKET_LINKED`, `VARIABLE` |
| `lendingRates[]` | `BankingProductLendingRateV3` | `lendingRateType`, `rate`, `applicationType`, `repaymentType`, `loanPurpose` | 12 values, for example `FIXED`, `VARIABLE`, `PURCHASE`, `CASH_ADVANCE` |
| `instalments` | `BankingProductInstalments` | `minimumSplit`, `maximumSplit` | none |

Conventions for these arrays: [S1 #cdr-banking-api_get-products] [S2 cds_banking.json `.paths["/banking/products"].get.description`]
- Each item has a type field. An item of the same type can occur more than one time.
- `additionalValue` is a generic string. Its content, and whether it is mandatory, depends on the type. The tables in "Product & Account Components" give the rule for each type. [S1 #product-amp-account-components]
- `additionalInfo` is display text. It is mandatory when the type is `OTHER` for features, constraints, eligibility, fees and rate conditions. [S2 cds_banking.json `.components.schemas.BankingProductFeatureV4.properties.additionalInfo`]
- Rates are RateString values. Rate tiers (`BankingProductRateTierV4`) require `name`, `unitOfMeasure` (`DAY`, `DOLLAR`, `MONTH`, `PERCENT`) and `minimumValue`. Frequencies are ISO 8601 durations. [S2]
- No ID links an account to a product. `productCategory` is the only field that accounts and products share. [S1 #cdr-banking-api_get-products]
- The full field lists are in the schema section of the Banking APIs. [S1 #cdr-banking-api_schemas_tocSbankingproductv6]

### 3.5 Rules that are not in the schema

What a data holder must publish:
- Required product data is CDR data with no CDR consumer, that is product specific data about a covered product, and "that is held by the data holder in a digital form". It excludes foreign currency accounts, consumer leases, reverse mortgages, margin loans and non-standard vehicle asset finance. Data about those products is voluntary product data. [S7 Schedule 3 clause 3.1]
- A data holder must disclose requested required product data "through its product data request service" and "in accordance with the data standards". It must include required product data that is on its website or in a disclosure document for the product. [S7 rule 2.4(3)]
- "Optional fields are not considered optionally implementable by a Data Holder." For payloads with no customer, such as product reference data, "there is more discretion for the Data Holder". [S1 #payload-conventions]
- PRD "should contain only products that are actively offered on market". Products that are not offered to new customers (grandfathered products) "do not appear in Product Reference Data". [S8: 360004255656]
- Products that are not offered now can stay in the set with `effectiveTo` in the past "for a few weeks before they are then removed". The dates remove the need to show deleted products. [S1 #cdr-banking-api_get-products]
- A data holder does not need to be on the Register to share product data. [S8: 6260497948431]
- The data holder must take "reasonable steps to ensure that the product data is ... accurate, up to date and complete". [S1 #data-quality]
- The ACCC considers that a data holder that gives the PRD location on its website meets the "publicly available" requirement. [S8: 900005460183]

Brands and white label:
- "The product reference data brand feature in the Get Products API is intended only for sub-brands of a master brand and should not be overloaded with products belonging to any other brand hosted on the Register." [S8: 900005881263]
- Each data holder brand on the Register has its own PRD endpoint, and its own Get Status and Get Outages endpoints. Products of a distribution-only brand or a portfolio extension go on the PRD endpoint of the brand that holds them. [S8: 900003938166 Tables for scenarios 1 to 6]
- `brandGroup` is for brand owners that have arrangements with more than one data holder. All the related data holder brands "should use the same brand group", and the value "should be immediately recognisable to the consumer". [S8: 900003938166]
- If a data holder offers a product for another data holder that makes the contracts with consumers, the first data holder is not required to disclose the required product data, unless the two agree in writing that it does. [S7 rule 2.4(4)–(5)]

---

## 4. Common discovery APIs

| Endpoint | Method and path | Version | Superseded versions | Transport | Auth | Performance tier |
|---|---|---|---|---|---|---|
| Get Status | `GET /discovery/status` | 1 | none | TLS | none | High Priority (1000 ms) |
| Get Outages | `GET /discovery/outages` | 1 | none | TLS | none | High Priority (1000 ms) |

[S2 cds_common.json `.paths["/discovery/status"].get`, `.paths["/discovery/outages"].get`] [S3] [S1 #performance-requirements] [S8: 900002974746]

- Request headers: `x-v` (mandatory), `x-min-v` (optional). Response header: `x-v`. No query parameters. No pagination. [S2]
- Responses: 200, 400 (Invalid Field, Missing Required Field, Invalid Version), 406 (Unsupported Version). [S2 cds_common.json `.paths["/discovery/status"].get.responses`]

Get Status response (`ResponseCommonDiscoveryStatus`): `data` and `links` are required. [S2 cds_common.json `.components.schemas.ResponseCommonDiscoveryStatus_data`]

| Field | Type | Required | Rule |
|---|---|:-:|---|
| `status` | Enum | yes | See the next table |
| `updateTime` | DateTimeString | yes | "The date and time that this status was last updated by the Data Holder." |
| `explanation` | string | conditional | "Mandatory if the status property is any value other than `OK`." Text that an end customer can see. |
| `detectionTime` | DateTimeString | no | "Should only be present if the status property is `PARTIAL_FAILURE` or `UNAVAILABLE`." |
| `expectedResolutionTime` | DateTimeString | no | "Should not be present if the status property has a value of `OK`." |

| `status` value | Meaning in the CDS |
|---|---|
| `OK` | "implementation is fully functional" |
| `PARTIAL_FAILURE` | "one or more endpoints are unexpectedly unavailable" |
| `UNAVAILABLE` | "the full implementation is unexpectedly unavailable" |
| `SCHEDULED_OUTAGE` | "an advertised outage is in effect" |

Get Outages response (`ResponseDiscoveryOutagesList`): `data.outages[]` is required and can be empty. Each `DiscoveryOutage` has: [S2 cds_common.json `.components.schemas.DiscoveryOutage`]

| Field | Type | Required | Rule |
|---|---|:-:|---|
| `outageTime` | DateTimeString | yes | Scheduled start |
| `duration` | ISO 8601 duration | yes | Planned duration |
| `explanation` | string | yes | Text that an end customer can see |
| `isPartial` | Boolean | no | `true` means that only some endpoints are affected |

Guidance:
- The status is the state of the full CDR implementation, "including all APIs and Information Security components". One endpoint down is `PARTIAL_FAILURE`. [S8: 900003123023]
- The two APIs "represent your CDR implementation as a whole", PRD and authenticated APIs together. [S8: 900005350723]
- `updateTime` is "the last time the information being presented was either updated, modified or verified". `detectionTime` is when the issue was found. [S8: 900005173683]
- Get Outages gives future planned outages. Get Status gives the outage that is in progress. The data holder does not add unplanned outages to Get Outages after they occur. [S8: 6783248192143]
- A data holder that shares only PRD must also have Get Status and Get Outages. [S8: 15490454471695]

---

## 5. Authenticated APIs in scope

### 5.1 Endpoint table

All these endpoints use MTLS and a certificate-bound access token. Paths are below `<ResourceBaseUri>/cds-au/v1`. [S2 `.servers`] [S1 #participant-endpoints]

| Endpoint | Method and path | Version | Superseded versions (retirement date) | Scope |
|---|---|---|---|---|
| Get Customer | `GET /common/customer` | 1 | none | `common:customer.basic:read` |
| Get Customer Detail | `GET /common/customer/detail` | 2 | 1 (2023-02-28) | `common:customer.detail:read` |
| Get Accounts | `GET /banking/accounts` | 3 | 2 (2026-12-07), 1 (2024-03-11) | `bank:accounts.basic:read` |
| Get Account Detail | `GET /banking/accounts/{accountId}` | 5 | 4 (2026-12-07), 3 (2026-05-11), 2 (2024-03-11), 1 (2023-02-28) | `bank:accounts.detail:read` |
| Get Account Balance | `GET /banking/accounts/{accountId}/balance` | 1 | none | `bank:accounts.basic:read` |
| Get Bulk Balances | `GET /banking/accounts/balances` | 2 | 1 (2026-12-07) | `bank:accounts.basic:read` |
| Get Balances For Specific Accounts | `POST /banking/accounts/balances` | 1 | none | `bank:accounts.basic:read` |
| Get Transactions For Account | `GET /banking/accounts/{accountId}/transactions` | 2 | 1 (2026-12-07) | `bank:transactions:read` |
| Get Transaction Detail | `GET /banking/accounts/{accountId}/transactions/{transactionId}` | 3 | 2 (2026-12-07), 1 (2025-11-10) | `bank:transactions:read` |

[S2 cds_common.json and cds_banking.json `.paths[][]["x-version"]`, `["x-scopes"]`] [S3]

- The versions 3, 5, 2, 2 and 3 of the five banking endpoints with a 2026-12-07 date become binding on 9 November 2026 (section 1.2).

### 5.2 Request headers

- Each endpoint in section 5.1 has the same header parameters: `x-v` (mandatory), `x-min-v`, `x-fapi-interaction-id`, `x-fapi-auth-date`, `x-fapi-customer-ip-address` and `x-cds-client-headers`. Sections 7.1 and 7.2 give the rules. [S2 cds_banking.json `.components.parameters`]
- `Content-Type` is "Mandatory for PUT and POST calls", and the media type must be `application/json`. This applies to Get Balances For Specific Accounts. [S1 #request-headers]
- `Accept` is optional. If present, it must be `application/json`. If it is absent or `*/*`, the default is `application/json`. An unacceptable value gets 406. [S1 #request-headers]
- The data holder must process `Content-Type` and `Accept` as RFC 7231 gives, for example case-insensitive and with a `charset` parameter. [S1 #request-headers]
- The access token is a bearer token. The customer comes from the token: Get Customer has no customer ID in the path, and the data holder uses the `sub` of the token. [S8: 900003476506]

### 5.3 Query parameters, defaults and limits

| Endpoint | Parameters |
|---|---|
| Get Customer, Get Customer Detail | none |
| Get Accounts, Get Bulk Balances | `product-category` (`BankingProductCategoryV2`; absent = all), `open-status` (`OPEN`, `CLOSED`, `ALL`; default `ALL`), `is-owned` (Boolean; absent = all), `page` (default 1), `page-size` (default 25) |
| Get Account Detail, Get Account Balance | path `accountId` only |
| Get Balances For Specific Accounts | `page`, `page-size`; body `{"data": {"accountIds": [...]}, "meta": {}}`. `data` and `data.accountIds` are required. |
| Get Transactions For Account | path `accountId`; `oldest-time`, `newest-time`, `min-amount`, `max-amount`, `text`, `page`, `page-size` |
| Get Transaction Detail | path `accountId`, `transactionId` |

[S2 cds_banking.json `.paths[][].parameters`, `.components.schemas.RequestAccountIdListV1`]

Get Transactions For Account parameters: [S2 cds_banking.json `.components.parameters`]

| Parameter | Type | Default | Rule |
|---|---|---|---|
| `oldest-time` | DateTimeString | `newest-time` minus 90 days | Effective time "at or after" |
| `newest-time` | DateTimeString | "today" | Effective time "at or before" |
| `min-amount` | AmountString | none | Amounts "higher than or equal to" |
| `max-amount` | AmountString | none | Amounts "less than or equal to" |
| `text` | string | none | Substring of `reference` or `description`. Optional for the data holder. |

- If the data holder does not implement `text`, it sends the usual response with no text filter, and `meta.isQueryParamUnsupported` is `true`, "whether the text parameter is supplied or not". [S2 cds_banking.json `.components.parameters.QueryBankingTransactionText`] [S8: 4899818390159]
- The effective time of a transaction, for filter and order, is the posted time if available, then the execution time if available, then "A reasonable date/time nominated by the data holder". [S2 cds_banking.json `.paths["/banking/accounts/{accountId}/transactions"].get.description`]
- Transactions "should be ordered according to effective date in descending order". [S2, same path]
- A negative amount reduces the available balance. A positive amount increases it. [S2, same path]
- If `oldest-time` or `newest-time` is out of the supported range, not possible, or the two conflict, the data holder "can respond with 400 Bad Request, error code urn:au-cds:error:cds-all:Field/InvalidDateTime". [S8: 5729181240591]
- `page-size` has no maximum in the OpenAPI schema. The maximum of 1000 is in the pagination text. [S2 cds_banking.json `.components.parameters.QueryPageSize`] [S1 #pagination]

### 5.4 Responses and error codes of each endpoint

The CDS says that these error codes "MUST be supported" for each endpoint. [S2 cds_banking.json and cds_common.json `.paths[][].responses`]

| Endpoint | 400 | 404 | 406 | 422 |
|---|---|---|---|---|
| Get Customer, Get Customer Detail | Invalid Field, Missing Required Field, Invalid Version | | Unsupported Version | |
| Get Accounts, Get Bulk Balances | Invalid Field, Missing Required Field, Missing Required Header, Invalid Version, Invalid Page Size | | Unsupported Version | Invalid Page |
| Get Balances For Specific Accounts | as Get Accounts | | Unsupported Version | Invalid Page, Unavailable Banking Account, Invalid Banking Account |
| Get Account Detail, Get Account Balance | Invalid Field, Missing Required Field, Missing Required Header, Invalid Version | Unavailable Banking Account, Invalid Banking Account | Unsupported Version | |
| Get Transactions For Account | as Get Accounts, and Invalid Date | Unavailable Banking Account, Invalid Banking Account | Unsupported Version | Invalid Page |
| Get Transaction Detail | Invalid Field, Missing Required Field, Missing Required Header, Invalid Version, Invalid Date | Unavailable Banking Account, Invalid Banking Account, Unavailable Resource, Invalid Resource | Unsupported Version | |
| Get Products | Invalid Field, Missing Required Field, Missing Required Header, Invalid Version, Invalid Date, Invalid Page Size | | Unsupported Version | Invalid Page |
| Get Product Detail | Invalid Field, Missing Required Field, Missing Required Header, Invalid Version | Unavailable Resource, Invalid Resource | Unsupported Version | |

- Each response of an authenticated endpoint, success or error, has the `x-fapi-interaction-id` header. The 200 response also has `x-v`. [S2 `.responses[].headers`]
- The OpenAPI files do not give 401, 403, 405, 415, 429 or 5xx responses for these endpoints. The general table gives them (section 7.6). [S2] [S1 #http-response-codes]

### 5.5 Main schemas

Customer: [S2 cds_common.json `.components.schemas`]
- `ResponseCommonCustomer.data`: `customerUType` (`person` or `organisation`, required) and the object with the same name.
- `CommonPerson` requires `lastName` and `middleNames` (the array can be empty). A person with one name has it in `lastName`. If the data holder cannot divide the given names, it "MAY" put them all in `firstName`. [S1 #cdr-common-api_get-customer_conventions]
- `CommonOrganisation` requires `agentLastName`, `agentRole` ("Default to "Unspecified" if the role is not known"), `businessName` and `organisationType` (`COMPANY`, `GOVERNMENT_ENTITY`, `OTHER`, `PARTNERSHIP`, `SOLE_TRADER`, `TRUST`).
- Get Customer Detail adds `phoneNumbers[]`, `emailAddresses[]` and `physicalAddresses[]` for a person (all required, can be empty), and `physicalAddresses[]` for an organisation. One and only one address has the purpose `REGISTERED`. Zero or one has the purpose `MAIL`. An address is `simple` or `paf` (`addressUType`).

Accounts (`BankingAccountV3`): [S2 cds_banking.json `.components.schemas.BankingAccountV3`]

| Field | Type | Required | Rule |
|---|---|:-:|---|
| `accountId` | ASCIIString | yes | ID permanence rules (section 7.5) |
| `displayName` | string | yes | Must not contain an account number or PAN, unless masked |
| `maskedNumber` | MaskedAccountString | yes | |
| `productCategory` | Enum | yes | Section 3.3 |
| `productName` | string | yes | |
| `accountOwnership` | Enum | yes | `UNKNOWN`, `ONE_PARTY`, `TWO_PARTY`, `MANY_PARTY`, `OTHER` |
| `openStatus` | Enum | no | `OPEN` or `CLOSED`. "If not present then `OPEN` is assumed." |
| `isOwned` | Boolean | no | "If not present then `true` is assumed." |
| `creationDate` | DateString | no | |
| `nickname` | string | no | |
| `isInstalmentDetailAvailable` | Boolean | no | Default `false` |

- Get Account Detail (`BankingAccountDetailV5`) adds only optional fields: `bsb`, `accountNumber` (both not masked, digits only), `bundleName`, `instalments`, `specificAccountUType` (`termDeposit`, `creditCard`, `loan`) with the object of the same name, `depositRate`, `lendingRate`, `depositRates[]`, `lendingRates[]`, `features[]`, `fees[]`, `addresses[]`. The rate, feature and fee items use the PRD structures. [S2 cds_banking.json `.components.schemas.BankingAccountDetailV5_allOf`]
- `accountNumber` "Should not be supplied if the account number is a PAN requiring PCI compliance". [S2, same schema]
- Get Account Detail does not depend on Get Product Detail. It returns the data of the account of the customer, also for a product that is not in PRD. [S8: 900004104506] [S8: 360004255656]

Balances (`BankingBalance`): `accountId`, `currentBalance` and `availableBalance` are required. `currentBalance` is "Assumed to be negative if the customer has money owing". `availableBalance` is "Assumed to be zero or positive". Optional: `creditLimit` (default `0.00`), `amortisedLimit` (default `0.00`), `currency` (default `AUD`), `purses[]`. [S2 cds_banking.json `.components.schemas.BankingBalance`]

Transactions (`BankingTransactionV2`): [S2 cds_banking.json `.components.schemas.BankingTransactionV2`]

| Field | Type | Required | Rule |
|---|---|:-:|---|
| `accountId` | ASCIIString | yes | |
| `transactionId` | ASCIIString | conditional | "mandatory (through hashing if necessary) unless there are specific and justifiable technical reasons". Mandatory if `isDetailAvailable` is `true`. |
| `isDetailAvailable` | Boolean | yes | `true` if Get Transaction Detail has extended data |
| `type` | Enum | yes | `DIRECT_DEBIT`, `FEE`, `INTEREST_CHARGED`, `INTEREST_PAID`, `OTHER`, `PAYMENT`, `TRANSFER_INCOMING`, `TRANSFER_OUTGOING` |
| `status` | Enum | yes | `PENDING` or `POSTED` |
| `description` | string | yes | |
| `amount` | AmountString | yes | |
| `reference` | string | yes | "Empty string if no data provided." |
| `postingDateTime` | DateTimeString | conditional | Mandatory if `status` is `POSTED` |
| `valueDateTime`, `executionDateTime` | DateTimeString | no | |
| `currency` | CurrencyString | no | Default `AUD` |
| `merchantName`, `merchantCategoryCode`, `billerCode`, `billerName`, `crn`, `apcaNumber`, `instalmentPlanId` | | no | |

- The list response uses `MetaPaginatedTransaction`: `totalRecords`, `totalPages` and the optional `isQueryParamUnsupported` (default `false`). [S2 cds_banking.json `.components.schemas.MetaPaginatedTransaction_allOf`]
- Get Transaction Detail (`BankingTransactionDetailV3`) adds the required object `extendedData`: `payer` (mandatory for an inbound payment), `payee` (mandatory for an outbound payment), `extensionUType` (`nppPayload` only) and `nppPayload` (`service` and `serviceVersion` required; `service` is `X2P1`, `IFTI`, `BSCT` or `CATSCT`). [S2 cds_banking.json `.components.schemas.BankingTransactionDetailV3_allOf_extendedData`]
- The CDS does not require a link between a pending transaction and its posted transaction. [S2 `BankingTransactionV2.properties.status`] [S8: 6471943957263]
- The data holder decides the `type` of each transaction. `PENDING` can change. `POSTED` is assumed to be final. [S8: 10994916162319]

---

## 6. The other banking endpoints (not in the first project)

| Endpoint | Method and path | Version | Scope |
|---|---|---|---|
| Get Direct Debits For Account | `GET /banking/accounts/{accountId}/direct-debits` | 1 | `bank:regular_payments:read` |
| Get Bulk Direct Debits | `GET /banking/accounts/direct-debits` | 2 | `bank:regular_payments:read` |
| Get Direct Debits For Specific Accounts | `POST /banking/accounts/direct-debits` | 1 | `bank:regular_payments:read` |
| Get Scheduled Payments for Account | `GET /banking/accounts/{accountId}/payments/scheduled` | 2 | `bank:regular_payments:read` |
| Get Scheduled Payments Bulk | `GET /banking/payments/scheduled` | 3 | `bank:regular_payments:read` |
| Get Scheduled Payments For Specific Accounts | `POST /banking/payments/scheduled` | 2 | `bank:regular_payments:read` |
| Get Instalment Plans for Account | `GET /banking/accounts/{accountId}/payments/plans` | 1 | `bank:regular_payments:read` |
| Get Instalment Plans Bulk | `GET /banking/accounts/payments/plans` | 1 | `bank:regular_payments:read` |
| Get Payees | `GET /banking/payees` | 2 | `bank:payees:read` |
| Get Payee Detail | `GET /banking/payees/{payeeId}` | 2 | `bank:payees:read` |

[S2 cds_banking.json `.paths`]

- An endpoint that is in the standards but not implemented returns 404 with `urn:au-cds:error:cds-all:Resource/NotImplemented`. This code "SHOULD be supported for unimplemented APIs". [S1 #error-404-resource-not-implemented]

---

## 7. Cross-cutting rules for data holder resource APIs

### 7.1 Endpoint versioning

- Each endpoint has its own version. The client asks for a version with a header. [S1 #endpoint-versioning]
- `x-v`: "Must be set to a positive integer. The holder should respond with the highest supported version between x-min-v and x-v." It is mandatory. [S1 #request-headers]
- `x-min-v`: optional. "If the value of x-min-v is equal to or higher than the value of x-v then the x-min-v header should be treated as absent." [S1 #request-headers]
- "If all versions requested are not supported then the holder must respond with a 406 Not Acceptable." The code is `urn:au-cds:error:cds-all:Header/UnsupportedVersion`. [S1 #request-headers] [S1 #error-406-header-unsupported-version]
- A version value that is not a positive integer gets 400 with `urn:au-cds:error:cds-all:Header/InvalidVersion`. This applies to `x-v`, `x-min-v` and `x-<HID>-v`. [S1 #error-400-header-invalid-version]
- The response header `x-v` is mandatory. It gives "The payload version that the endpoint has responded with." [S1 #response-headers]
- `x-<HID>-v` is an optional header for the version of holder-specific extension fields. It "Should not be used in conjunction with x-min-v." `<HID>` is the holder identifier prefix, by preference the ASX symbol. A holder can add fields, query parameters, headers and endpoints with this prefix. It "MUST NOT" change existing fields, add enumeration values or make a mandatory field optional. [S1 #request-headers] [S1 #extensibility]
- Headers are strings. A value with leading zeros, for example `x-v: 0001`, is best treated as valid, with the response `x-v: 1`. [S8: 900004807083]
- `x-v` does not apply to an error response that has no payload version. [S8: 5729181240591]

### 7.2 FAPI and CDS headers

| Header | Rule |
|---|---|
| `x-fapi-interaction-id` (request) | "An optional [RFC4122] UUID used as a correlation id." "Not required for unauthenticated calls." |
| `x-fapi-interaction-id` (response) | Mandatory. The data holder sets it "to the value received from the corresponding request header or to a new [RFC4122] UUID value if the request header was not provided. This header MUST be responded for success and error responses for authenticated APIs." |
| `x-fapi-auth-date` | "The time when the customer last logged in to the Data Recipient Software Product as described in [FAPI-1.0-Baseline]. Required for all resource calls (customer present and unattended). Not required for unauthenticated calls." |
| `x-fapi-customer-ip-address` | "The customer's original IP address if the customer is currently logged in". "The presence of this header indicates that the API is being called in a customer present context. Not to be included for unauthenticated calls." |
| `x-cds-client-headers` | The original HTTP headers of the customer, Base64 encoded, with the `User-Agent` header. "Mandatory for customer present calls. Not required for unattended or unauthenticated calls." It does not need security headers or proprietary headers. |

[S1 #request-headers] [S1 #response-headers]

- The format of `x-fapi-auth-date` is an HTTP-date (RFC 7231 section 7.1.1.1), for example `Tue, 11 Sep 2012 19:43:31 GMT`. [S8: 4405773134863]
- "Technically a data holder will define an API request as "Customer Present" if, and only if, the x-fapi-customer-ip-address header is populated with a valid IP address of the end customer's device." All other authenticated requests are unattended. [S1 #definitions]
- In the OpenAPI files, `x-fapi-auth-date`, `x-fapi-customer-ip-address` and `x-cds-client-headers` have `required: false`. `x-fapi-auth-date` and `x-cds-client-headers` have `x-conditional: true`. [S2 cds_banking.json `.components.parameters.HeaderXFAPIAuthDate`]
- DSB guidance: "Use of the client x-fapi headers is at the discretion of the DH", except that the data holder must return `x-fapi-interaction-id`. "DHs must validate that the mandatory headers are provided or return an error code." [S8: 4410664891919]
- DSB guidance: the `x-cds-client-headers` rule is a requirement on the data recipient. "A request should not be rejected based on the syntax of this header." [S8: 4404953530767]
- A missing required header "should" get 400 with `urn:au-cds:error:cds-all:Header/Missing`. An invalid header value has the code `Header/Invalid`. The two codes "SHOULD be supported". [S1 #error-400-header-missing] [S1 #error-400-header-invalid] [S8: 5729181240591]

### 7.3 Payload conventions

- A request payload "MUST" have a root object with a `data` object. It has `meta` only if the endpoint requires it. [S1 #request-payload-structure]
- A 200 response root object "MUST contain a data object", "MUST contain a links object" and "MAY contain a meta object if REQUIRED by the definition of the specific endpoint". [S1 #response-payload-structure]
- "The links object MUST contain a field named self that will have the fully qualified URI to the current request as a value." [S1 #response-payload-structure]
- A response that is not 200 "MAY contain an errors object". [S1 #response-payload-structure]
- Field names are case sensitive and use camel case. The characters are `a-z`, `A-Z`, `0-9`, and `-`, `_`, `$` not at the start or end. [S1 #field-naming-conventions]
- Mandatory fields "MUST be present and have a non-null value". Optional fields can be absent or `null`. A conditional field is mandatory when its condition is true. [S1 #field-property-conventions]
- An absent field is equal to a `null` field. An empty string is not equal to `null`. `false` is not equal to `null`. [S1 #field-property-conventions]
- A mandatory array field is always an array, also when it is empty or has one item. [S1 #array-conventions]
- A union object has a field with the suffix `UType` that names the object that is present, for example `customerUType`. [S1 #object-conventions]
- A field with the name `id` is never used. Each ID has a meaningful name, for example `accountId`. [S1 #id-permanence]

Common field types: [S1 #common-field-types]

| Type | Rule |
|---|---|
| String | UTF-8, any Unicode character |
| ASCIIString | UTF-8, ASCII characters only |
| Boolean | JSON boolean |
| Enum | String. All values are given. "Values MUST be in all caps", with `_` for a space, ASCII only. |
| NaturalNumber, PositiveInteger, NegativeInteger, Integer, Number | 0 or more; 1 or more; 0 or less; integer; integer or decimal |
| Base64 | RFC 4648 |
| DateTimeString | RFC 3339 `date-time`. "times MUST be offset relative to UTC." Example `2007-05-01T15:43:00.12345Z` |
| DateString | RFC 3339 `full-date`, for example `2007-05-01` |
| TimeString | RFC 3339 `full-time` |
| CurrencyString | ISO 4217, 3 characters |
| RateString | A percentage as a decimal string: `"1"` is 100%, `"0.2"` is 20%. 1 to 16 digits before the point, a maximum of 16 after. No format characters. |
| AmountString | Positive, zero or negative, with `-` for negative. No currency symbol. 1 to 16 digits before the point. "Minimum 2 digits following a decimal point". No format characters. AUD unless specified. Examples `"0.01"`, `"-1001.23"`, `"1.999"` |
| MaskedPANString | Lower case `x` for each digit but the last four, for example `"xxxx xxxx xxxx 1234"` |
| MaskedAccountString | The number as shown for display, with the format, "but with all digits except the last four replaced with a lower case x", for example `"xxx-xxx xxxxx1234"` |
| URIString | A valid URI |
| ExternalRef | A format from an external reference, for example an ISO standard |

- Mask characters other than lower case `x` are not permitted. [S8: 5527282549135]

### 7.4 Pagination

- Query parameters: `page` (first page is 1; default 1) and `page-size` (default 25). [S1 #pagination]
- The holder "MUST" give, in `links`: `first` and `prev` ("Mandatory if this response is not the first page"), `next` and `last` ("Mandatory if this response is not the last page"). [S1 #pagination]
- The holder "MUST" give, in `meta`: `totalRecords` and `totalPages`. "If totalRecords is 0 totalPages MUST be 0." [S1 #pagination]
- "A maximum page size of 1000 records is assumed for all endpoints (unless otherwise stipulated in the endpoint definition). If a page size greater than this maximum is requested then an Invalid Page Size error SHOULD be returned." The code is 400 `urn:au-cds:error:cds-all:Field/InvalidPageSize`. [S1 #additional-pagination-rules] [S1 #error-400-field-invalid-page-size]
- A page that is out of range gets 422 `urn:au-cds:error:cds-all:Field/InvalidPage`. "The error detail SHOULD be the maximum number of pages that are available." [S1 #error-422-field-invalid-page]
- Holders "are not expected to implement pagination with transaction isolation". A record can be on more than one page. [S1 #additional-pagination-rules]
- Cursors are permitted in the link URIs. The link and `meta` fields are then still required. [S1 #cursor-support]
- With one page of results, `first` and `last` are not mandatory, and `prev` and `next` are absent. [S8: 4874672614927]

### 7.5 ID permanence

Resource IDs are "REQUIRED to comply" with these rules: [S1 #id-permanence]
- An ID that the standards specify "is mandatory and MUST be supplied".
- "the ID value MUST be entirely arbitrary and have no inherent meaning". It must not be a combination of other fields or a string that gives information when parsed.
- "IDs SHOULD be unique", possibly in a bounded context only.
- "IDs MUST be immutable across sessions and consents but MUST NOT be transferable across Data Recipient Software Products." Two software products get different IDs for the same account, also with the same customer.
- "IDs MUST NOT be transferable between different customers for the same Data Recipient Software Product." A software product gets a different ID for a joint account from each account holder.

Guidance:
- `accountId`, `transactionId`, `scheduledPaymentId` and `payeeId` follow the ID permanence rules. [S8: 6140375237647]
- If one subject shares the same account in two arrangements, "the same ID must be provided". The DSB suggests an ID that is keyed on the subject, or on the subject and the client ID. [S8: 6140375237647]
- `productId` does not follow the ID permanence rules. [S2 cds_banking.json `.components.schemas.BankingProductId`]
- A data recipient can call Get Transaction Detail many times in the life of a consent. The transaction ID must stay the same. [S8: 5729181240591]

### 7.6 HTTP response codes and the error catalogue

Error response structure: [S1 #error-codes] [S2 cds_banking.json `.components.schemas.ErrorV2`]
- The root has `errors`, an array. Each item has `code`, `title` and `detail`, which "MUST be present". `meta.urn` is optional.
- `title` "MUST NOT change from occurrence to occurrence" for one code. [S2 `ErrorV2.properties.title`]
- URN structure: `urn:au-cds:error:<sub-type>:<error-category>/<error-code>`. Sub-types: `cds-all`, `cds-register`, `cds-banking`, `cds-energy`. [S1 #error-codes]
- Participants "SHOULD use the standard error codes". An application-specific code is permitted in `code`. The response then "MUST" have the standard URN in `meta.urn`. [S1 #error-codes] [S1 #extensibility-and-application-specific-errors]
- For more than one problem, "the most generally applicable HTTP error code SHOULD be used". The server "MAY" stop at the first problem or return all problems. [S1 #error-codes]

Standard codes that apply to the endpoints in scope: [S1 #standard-error-codes]

| HTTP | Title | URN (after `urn:au-cds:error:`) | Support | Use |
|---|---|---|---|---|
| 4xx | Expected Error Encountered | `cds-all:GeneralError/Expected` | | An expected error with no more specific code |
| 5xx | Unexpected Error Encountered | `cds-all:GeneralError/Unexpected` | | |
| 503 | Service Unavailable | `cds-all:Service/Unavailable` | | Outage |
| 400 | Missing Required Field | `cds-all:Field/Missing` | MUST | Missing query parameter or body field. Detail: the field name. |
| 400 | Missing Required Header | `cds-all:Header/Missing` | SHOULD | Detail: the header name |
| 400 | Invalid Field | `cds-all:Field/Invalid` | MUST | Wrong type or constraint. Detail: the field name. |
| 400 | Invalid Header | `cds-all:Header/Invalid` | SHOULD | Detail: the header name |
| 400 | Invalid Date | `cds-all:Field/InvalidDateTime` | MUST | DateTimeString, DateString, TimeString |
| 400 | Invalid Page Size | `cds-all:Field/InvalidPageSize` | MUST | `page-size` more than 1000 |
| 400 | Invalid Version | `cds-all:Header/InvalidVersion` | MUST | Version is not a positive integer |
| 403 | ADR Status Is Not Active | `cds-all:Authorisation/AdrStatusNotActive` | | Detail: the status of the software product |
| 403 | Consent Is Revoked | `cds-all:Authorisation/RevokedConsent` | SHOULD | Consent revoked or expired |
| 403 | Consent Is Invalid | `cds-all:Authorisation/InvalidConsent` | SHOULD | Consent not associated with the resource, or not sufficient |
| 404 | Resource Not Implemented | `cds-all:Resource/NotImplemented` | SHOULD | Endpoint is in the CDS but not implemented |
| 404 | Resource Not Found | `cds-all:Resource/NotFound` | | URL is not a CDS endpoint |
| 404 | Invalid Resource | `cds-all:Resource/Invalid` | MUST | Resource ID in the URI, permanently unavailable |
| 404 | Unavailable Resource | `cds-all:Resource/Unavailable` | MUST | Resource ID in the URI, temporarily unavailable |
| 404 | Invalid Banking Account | `cds-banking:Authorisation/InvalidBankingAccount` | MUST | Account ID in the URI, permanently unavailable |
| 404 | Unavailable Banking Account | `cds-banking:Authorisation/UnavailableBankingAccount` | MUST | Account ID in the URI, temporarily unavailable |
| 406 | Unsupported Version | `cds-all:Header/UnsupportedVersion` | MUST | Version lower or higher than the supported range |
| 422 | Invalid Resource | `cds-all:Resource/Invalid` | MUST | Resource ID in the request body |
| 422 | Unavailable Resource | `cds-all:Resource/Unavailable` | MUST | Resource ID in the request body |
| 422 | Invalid Banking Account | `cds-banking:Authorisation/InvalidBankingAccount` | MUST | Account ID in the request body |
| 422 | Unavailable Banking Account | `cds-banking:Authorisation/UnavailableBankingAccount` | MUST | Account ID in the request body |
| 422 | Invalid Page | `cds-all:Field/InvalidPage` | MUST | Page out of range |

- For the account and resource codes, "The error detail is the account ID of the resource being requested" (or the resource ID). [S1 #error-404-authorisation-invalid-banking-account]
- Invalid means: "permanently unavailable. No subsequent request for the account will be successful." Unavailable means: "temporarily unavailable. Subsequent requests for the account may be successful." [S1 #error-404-authorisation-invalid-banking-account] [S1 #error-404-authorisation-unavailable-banking-account]
- 404 "MUST NOT be used for resources presented in the body of the request". [S1 #http-response-codes]

General HTTP codes: [S1 #http-response-codes]

| Situation | Code |
|---|---|
| Malformed or missing body or parameters | 400 |
| "Authorization header missing or invalid token" | 401 |
| "Token has incorrect scope or a security policy was violated." An error payload is expected. | 403 |
| Method not supported | 405 |
| `Accept` not permitted, character set not UTF-8, or version not supported | 406 |
| Payload format not supported (POST) | 415 |
| Well formed, but business logic cannot process it | 422 |
| Too many requests. The holder "should include a Retry-After header". | 429 |
| Server failure; unavailable; timeout with retry possible | 500; 503; 504 |

Rules for an account that is not in the consent or not valid:
- Get Accounts returns only the accounts that can be shared. An account that cannot be shared "must not be returned in the accounts list". [S8: 5729181240591]
- For an `accountId` that cannot be shared: in the URL, the status is 404. In the request body, the status is 422. [S8: 5729181240591] [S8: 14519397304079]
- A request with a list of accounts gets an error and no data if one or more accounts cannot be shared. The CDS does not permit a partial response. The 422 response names the accounts. [S8: 5729181240591] [S8: 5809287867151]
- For a filter endpoint (for example Get Bulk Balances), the data holder can omit accounts that are not in the consent. This is not a refusal. [S8: 5729181240591 "Non-disclosure due to a bad request is not refusal"]
- The data holder selects Invalid or Unavailable by "whether future calls to the same resource are likely to receive a response". For fraud or vulnerability, the data holder selects the code and the detail text: 403, 404 (ID in the path) or 422. [S8: 5729181240591]
- If one source fails in a bulk response, the data holder "should return a 500 Internal Server Error". [S8: 5809287867151]
- Get Transaction Detail for a transaction with no extended data returns 404 `urn:au-cds:error:cds-all:Resource/Invalid`, and the list has `isDetailAvailable: false` for it. [S8: 5729181240591]
- A client certificate that is absent, expired or not from the CDR CA: "the DH can return a 401". A withdrawn consent with a revoked token gets 401 with an OAuth error response. [S8: 5729181240591]
- Errors from OAuth, OIDC and FAPI take precedence over CDS error codes. [S8: 5809287867151]

---

## 8. Accounts and transactions that are in scope

The CDS does not give these rules. The CDR Rules and ACCC guidance give them.

| Topic | Rule | Source |
|---|---|---|
| Relevant accounts | An account for a covered product that the consumer holds alone, a joint account, a partnership account, or an account for which the consumer is a secondary user, "whether or not the account can be accessed online" | S7 Schedule 3 clause 3.2(1) |
| Excluded products | Foreign currency accounts, consumer leases, reverse mortgages, margin loans and non-standard vehicle asset finance are not required consumer data | S7 Schedule 3 clause 3.2(3) |
| Open accounts, transactions | Transaction data for "a transaction that occurred more than 2 years before that time" is not required consumer data. It is voluntary consumer data. | S7 Schedule 3 clause 3.2(6)(a) |
| Open accounts, direct debits | A direct debit that "occurred more than 13 months before that time" is not required | S7 Schedule 3 clause 3.2(6)(b) |
| Closed accounts | "if a relevant account is closed at a particular time, CDR data relating to that account is not required consumer data at that time." It is voluntary consumer data. | S7 Schedule 3 clause 3.2(7) |
| Other persons | Customer data about an account holder or secondary user other than the requester is not shared | S7 Schedule 3 clause 3.2(5)(b) |
| Minors on joint accounts | Data of a joint or partnership account with a holder below 18 years is not shared | S7 Schedule 3 clause 3.2(5)(a) |
| Customer data | Does not include the date of birth of an individual | S7 Schedule 3 clause 1.3 |

- For a consent given on or after 4 March 2025, closed account data is voluntary. Get Accounts returns "the list of accounts that can be shared, including those they have elected to share voluntarily". [S8: 14519397304079]
- For a consent given before 4 March 2025, the previous rules continue: for an account closed 2 years ago or less, transactions from the 12 months before closure; for an account closed more than 2 years ago, nothing. [S8: 14519397304079]
- When a consumer closes all accounts with a data holder, the consumer is not eligible and the authorisations expire. [S8: 14519397304079] [S8: 6437909148943]
- Eligibility: the consumer has an open account with the data holder that "is set up so it can be accessed online", and is 18 years or over, or is not an individual, or is a partner. [S8: 14519397304079, citing S7 rule 1.10B and Schedule 3 clause 2.1]
- `open-status` filters on `openStatus`. The default `ALL` includes closed accounts that the data holder shares. [S2 cds_banking.json `.components.parameters.QueryBankingAccountOpenStatus`]
- The default transaction period is the 90 days before `newest-time`. The CDS gives no limit for `oldest-time`. [S2 cds_banking.json `.components.parameters.QueryOldestTime`]
- The transaction APIs "should present the same list as can be seen by the customer in their internet banking experience". The data holder decides which transactions are pending. [S8: 6471943957263]
- `isOwned`: "Flag indicating that the customer associated with the authorisation is an owner of the account. Does not indicate sole ownership". [S2 cds_banking.json `.components.schemas.BankingAccountV3.properties.isOwned`] An additional card holder that is not a primary account holder has `false`. [S8: 900005202163] For a business consumer, the flag is about the organisation, not the nominated representative. [S8: 5080119008015]
- `accountOwnership` gives the number of owners "according to the data holder's definition of account ownership". [S2 `BankingAccountV3.properties.accountOwnership`]
- Joint accounts (pointer only): Part 4A of the CDR Rules gives the disclosure options. The default is the pre-approval option. [S7 Part 4A, rule 4A.4] If a joint account stops being shareable, the status is 404 for an ID in the path and 422 for an ID in the body. [S8: 6437916008719]
- An archived article says that banking transaction data is required for seven years. The current rules replace it. (ARCHIVED) [S8: 10085678783759]

---

## 9. Non-functional requirements

The CDS NFRs apply to data holders. [S1 #non-functional-requirements]

Availability: [S1 #availability-requirements]
- "99.5% per month", for authenticated and unauthenticated endpoints.
- Unavailable means that "any of the API endpoints defined in the standard is unable to reliably provide a successful response to an appropriately constructed request".
- Planned outages are excluded. They should have "at least one week lead time for normal outages". An outage for a critical service or security issue can occur with no notice.

Performance: "95% of calls per hour responded to within a nominated threshold". The data holder measures as near to the data recipient as practicable. [S1 #performance-requirements]

| Tier | Threshold | Endpoints in scope |
|---|---|---|
| High Priority | 1000 ms | Get Status, Get Outages (all calls). Customer present calls to Get Accounts, Get Customer, Get Customer Detail. |
| Low Priority | 1500 ms | Customer present calls to Get Account Detail, Get Account Balance, Get Bulk Balances, Get Balances For Specific Accounts, Get Transactions For Account, Get Transaction Detail |
| Unauthenticated | 1500 ms | Get Products, Get Product Detail ("All Unauthenticated endpoints not otherwise specified") |
| Unattended | 4000 ms | Unattended calls to all the authenticated endpoints in scope |
| Large Payload | 6000 ms | None in scope (Get Bulk Direct Debits, Get Direct Debits For Specific Accounts) |

Traffic thresholds. The data holder can throttle or reject calls above these values. [S1 #traffic-thresholds]

| Traffic | Threshold |
|---|---|
| Customer present and authorisation | Unlimited sessions each day; 10 TPS for each customer; 50 TPS for each software product |
| Unattended, low traffic period (midnight to 6 am) | 20 sessions each day for each customer and software product; 100 calls for each session; 5 TPS for each session; 50 TPS for each software product |
| Unattended, high traffic period (6 am to midnight) | "only best effort support is required" |
| Secure traffic, total | 150 peak TPS for 0 to 10,000 active authorisations, then 50 more for each 10,000, to 450 peak TPS for more than 60,000 |
| Public traffic, total | 300 TPS, "additive to secure traffic" |

- A session is "the life span of a unique Access Token". [S1 #definitions]
- A call above a threshold gets 429. The response should have `Retry-After`. [S1 #http-response-codes] [S8: 5729181240591]
- The data holder can also return 429 for a denial of service attack or a software product that behaves incorrectly, and 403 when it finds "the potential for physical or financial harm or abuse". [S1 #exemptions-to-protect-service]

Data latency and quality:
- There is "no specific requirement" for latency. Data "should be commensurate to data presented via other primary digital channels". A balance from the API "should be the same as the balance presented through the mobile application". [S1 #data-latency]
- Product data: "reasonable steps to ensure that the product data is ... accurate, up to date and complete". [S1 #data-quality]

---

## 10. Conflicts and gaps found

1. PRD base URI. CDS 1.36.0 has a separate ProductBaseUri that "MAY be the same as the publicBaseUri" [S1 #participant-endpoints]. A support article of 2026 agrees [S8: 16505555811855]. An older support article says that a different base URI for PRD and for Get Status and Get Outages is "not acceptable" [S8: 900005460183]. The CDS wins.
2. Get Product Detail path. The OpenAPI file has `/banking/products/{productId}` [S2]. The Endpoint Version Schedule has `/banking/product/{productId}`. The schedule also has a row "Get Account Balances" with `/banking/accounts/balances`, version 1 and no retirement date, in addition to the two Get Bulk Balances rows [S3].
3. Error URN sub-type. The catalogue gives `urn:au-cds:error:cds-banking:Authorisation/UnavailableBankingAccount`. The non-normative 422 example gives `urn:au-cds:error:cds-all:Authorisation/UnavailableBankingAccount`. The transition examples use `cdr-all` [S1 #standard-error-codes] [S1 #error-codes].
4. `detail` field. `ErrorV2` requires `detail` [S2]. The non-normative 404 examples for `Resource/NotFound` and `Resource/NotImplemented` have no `detail` [S1 #standard-error-codes].
5. Invalid version value. The CDS gives 400 `Header/InvalidVersion` for a version that is not a positive integer [S1 #error-400-header-invalid-version]. The DSB error guide says that 406 "is a suitable response" for "a negative or non-integer value" [S8: 5729181240591].
6. Account not in the consent. The CDS and one part of the DSB guide give 404 for an ID in the URL and 422 for an ID in the body. A different part of the same guide ("Consent and account relationship") and the "Error responses" article give 422 for an ID "in the URI or as a filter" [S1 #standard-error-codes] [S8: 5729181240591] [S8: 5809287867151].
7. Transaction not available. The DSB guide says that 422 is correct for a `transactionId` that was valid and is not available now [S8: 5729181240591]. Get Transaction Detail has no 422 response, and the CDS gives 404 `Resource/Unavailable` for an ID in the URI [S2].
8. Page out of range. The CDS says that 422 `Field/InvalidPage` "MUST be supported" [S1 #error-422-field-invalid-page]. The DSB guide permits 200 with no records, 400 or 422 [S8: 5729181240591].
9. Page size above 1000. The pagination text says that the error "SHOULD be returned". The error catalogue says that the code "MUST be supported" [S1 #additional-pagination-rules] [S1 #error-400-field-invalid-page-size].
10. `x-fapi-auth-date`. The CDS text says "Required for all resource calls" and marks the header conditional. The OpenAPI file has `required: false`. DSB guidance says that use of the header is "at the discretion of the DH", and also that data holders "must validate that the mandatory headers are provided". Get Customer and Get Customer Detail do not list 400 Missing Required Header, and the banking endpoints do [S1 #request-headers] [S2] [S8: 4410664891919] [S9: 532].
11. `x-fapi-interaction-id` on public endpoints. The CDS says "Not required for unauthenticated calls", and the public operations do not define the header [S1 #request-headers] [S2]. The CORS guidance lists the header for public endpoints "where supported by the Data Holder" [S8: 900003054706].
12. `brandGroup` with no version change. CDS 1.36.0 adds `brandGroup` to `BankingProductV6`, but Get Products stays at version 5 and Get Product Detail at version 7. The schema name does not change [S4 1.36.0] [S6 section 5.1].
13. Closed accounts. The grandfathered products article says that closed account data is required for 24 months [S8: 360004255656]. The rules of 4 March 2025 and a newer article say that it is voluntary [S7 Schedule 3 clause 3.2(7)] [S8: 14519397304079].
14. Version dates. The OpenAPI files at `v1.36.0` describe versions that are not binding until 9 November 2026. The binding versions at the research date are only on the "obsolete" pages [S2] [S5] [S1 #future-dated-obligations].
15. `newest-time` default. The default is "today". The CDS does not say if this is the current time or the end of the day, or which time zone applies [S2 cds_banking.json `.components.parameters.QueryNewestTime`]. Issue 173 (open since 2020) records a related problem for posting times with a date only [S9: 173]. Issue 714 (open) proposes to clarify transaction order [S9: 714].
16. Field errors. Issue 727 (open) says that the difference between `Field/Missing` and `Field/Invalid` is not clear for path, query and body parameters [S9: 727]. Issue 728 (open) says that "per customer" in the unattended session threshold is not clear [S9: 728].
17. No 401 or 403 in the endpoint definitions. The OpenAPI files give no 401, 403 or 429 response for the authenticated endpoints. The general table and the error catalogue give them [S2] [S1 #http-response-codes].
18. Stale text. The `product-category` parameter of Get Products says "applicable to accounts". It is a copy of the account parameter [S2 cds_banking.json `.components.parameters.QueryBankingProductCategory`]. The PRD support article names `BankingProductDetailV4`. The current schema is `BankingProductDetailV7` [S8: 900004104506].

Items marked **UNVERIFIED** in this note:
- The content of a CDS release after 1.36.0 (section 1.1).
- The full field differences between each superseded endpoint version and its new version (section 1.2).
- How the `brand` query parameter is compared with the `brand` field (section 3.2).

I did not read these documents, and this note does not depend on them: the ACCC "Compliance guide for data holders in the banking and non-bank lenders sectors", the CX Guidelines, and the support article "ID Permanence and PPID" (5505351487631; the help centre API returned no content).

## 11. Open questions for requirements

- Which endpoint versions does the mock support: only the versions in the `v1.36.0` OpenAPI files, or also the versions that are binding until 9 November 2026 and can be retired from 7 December 2026?
- Does the mock use one base URI for PublicBaseUri and ProductBaseUri, or two? The Register must then hold `productBaseUri` for the brand.
- Does the mock return `x-fapi-interaction-id` on public endpoints?
- Does the mock reject a request with no `x-fapi-auth-date` (400 `Header/Missing`), or accept it?
- Which response does the mock give for a version value that is not a positive integer: 400 (CDS) or 406 (DSB guide)?
- Which response does the mock give for a page that is out of range: 422 (CDS "MUST be supported") or one of the other permitted responses?
- For an `accountId` in the path that is not in the consent: 404 with Invalid Banking Account or Unavailable Banking Account? The two are permitted. The selection can show if an account exists.
- Does the mock share closed accounts (voluntary data)? If yes, for which period?
- Which transaction history period does the mock hold: the required 2 years, or more as voluntary data? What does it return for `oldest-time` before that period: an empty list or 400 `Field/InvalidDateTime`?
- Does the mock implement the `text` filter, or set `isQueryParamUnsupported`?
- Does the mock implement `transactionId` and Get Transaction Detail extended data (NPP payload) for some transactions, or return `isDetailAvailable: false` for all?
- How does the mock make `accountId` and `transactionId` values that are stable for one customer and one software product, and different for other customers and software products?
- How does the mock interpret "today" for the `newest-time` default?
- Does the mock return 404 `Resource/NotImplemented` for the banking endpoints in section 6?
- Does the mock enforce the traffic thresholds (429 with `Retry-After`)? Which performance targets apply to a lab deployment?
- Which product categories and how many brands does the seed data have, and does it use `brandGroup`?
- Does the mock support holder extensions (`x-<HID>-v`)? The CDS makes them optional.

## 12. Source URLs for superseded versions (S5) and support articles (S8)

Superseded endpoint version pages (S5), base URL https://consumerdatastandardsaustralia.github.io/standards/includes/obsolete/:
`get-accounts-v2.html`; `get-account-detail-v4.html`; `get-bulk-balances-v1.html`; `get-transactions-for-account-v1.html`; `get-transaction-detail-v2.html`; `get-products-v4.html`; `get-product-detail-v6.html`. Each page gives the retirement date of its version.

CDR Support Portal articles used (S8), base URL https://cdr-support.zendesk.com/hc/en-us/articles/:
16505555811855 Understanding Product Base URI and URI Structure for Product Reference Data; 900005460183 PRD endpoints and public endpoint URI discovery; 6260497948431 CDS Guide, Product Reference Data (PRD); 900004104506 Product Reference Data; 360004255656 Grandfathered products; 900005881263 Brands in the CDR ecosystem; 900003938166 White Labelled brands in the CDR; 900003123023 Definition of 'status' and how health check status can be determined; 900005173683 Get Status Payload Fields; 900005350723 Get Status and Get Outages when different vendors supply PRD and Consumer Data; 900002974746 Get Status and Get Outages Performance Tier; 6783248192143 Outages; 15490454471695 ACCC expectations for NBL compliance with Get Metrics, Get Outages and Get Status endpoints; 900003054706 CORS support; 900004807083 Allowed values for x-v and x-min-v; 4405773134863 Format for x-fapi-auth-date; 4410664891919 FAPI header obligations for Data Holders; 4404953530767 DH requirement for x-cds-client-headers; 4874672614927 LinksPaginated properties; 5729181240591 Error scenarios and responses; 5809287867151 Error responses; 6140375237647 Guidance on ID permanence; 5527282549135 Guidance on masking for CDR data; 900003476506 Identifying the customer in Get Customer and Get Customer Detail calls; 14519397304079 Closed accounts in the banking and non-bank lenders sectors; 6437909148943 Guidance on closed accounts; 10085678783759 Historical transaction data: Earliest Holding Day (ARCHIVED); 900005202163 Setting the isOwned attribute for a shared account; 5080119008015 isOwned flag in the context of business data sharing; 6437916008719 Guidance on Joint Accounts; 6471943957263 Pending and Posted Transactions; 10994916162319 Guidance on BankingTransaction; 4899818390159 Response when text filter not implemented.
