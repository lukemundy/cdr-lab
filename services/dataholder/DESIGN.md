# Mock data holder: design

This document gives the specification and the requirements for the mock data holder. It does not describe the implementation.

## Purpose

The mock data holder implements a CDR data holder brand as the CDS specifies it. Data recipients and public clients in this lab use it the same way they use a production data holder. A "real" component must be able to change from the mock data holder to a production data holder with a configuration change only.

## Sources

Target: Consumer Data Standards (CDS) 1.36.0. CDS links go to the 1.36.0 archive URL. While 1.36.0 is the current version, that URL redirects to the live site.

The requirements follow the process in [README: Sources for requirements][readme-sources]. Each requirement links to its source. "Judgement" marks a requirement that the CDS and the ACCC do not state, and gives the reason.

Background research, with the full detail and more sources, is in [`docs/research/`](../../docs/research/README.md).

## Scope

In scope:

- One data holder brand of the industry `banking`.
- The brand of the instance: a brand directory that the configuration selects, and preset brands in the data holder. The name, logo and colours of the brand on the consumer authentication and consent pages, on the OTP email and in the shell of the operator console, and its product catalogue.
- The public (TLS) APIs: product reference data, Get Status and Get Outages.
- The OpenID Provider configuration and the JWKS of the brand.
- Dynamic Client Registration with an SSA, and registration management.
- The token endpoint, for the client credentials, authorisation code and refresh token grants.
- The PAR endpoint, the authorisation endpoint, the introspection endpoint, the UserInfo endpoint, the token revocation endpoint and the CDR arrangement revocation endpoint.
- The authentication of a consumer with a one-time password, and the pages on which the consumer gives consent.
- The APIs that give the data of a consent: Get Customer, Get Accounts, Get Account Detail, Get Account Balance, Get Bulk Balances, Get Balances For Specific Accounts, Get Transactions For Account and Get Transaction Detail.
- The end of a consent: a revocation by the data recipient, a withdrawal by the consumer, the end of the sharing period and the end of the eligibility of the consumer, and the notice to the data recipient.
- The reaction to the status of a software product on the Register.
- The customers, accounts and transactions of the brand, and the lab-only operator interface that creates them, closes accounts and withdraws consents for a consumer.
- The lab-only feature to load product reference data.
- The lab-only operator console: a web UI for the staff of the bank.
- The lab-only online bank site of the brand, for consumers: a landing page, the list and the detail of the current products, sign-up, sign-in and sign-out, the accounts of the consumer with their balances and transactions, the opening and the close of an account by the consumer, and the consumer dashboard, on which the consumer sees their CDR consents and withdraws a consent (see Online bank site).
- The four preset brands in the lab at the same time, each as its own instance, behind a lab-only compose profile.
- The publication of the data holder image and of a Helm chart from `main` by CI, so that an operator deploys a brand to Kubernetes with configuration only (see Deployment from published artifacts). The deployment itself is out of scope.

Out of scope for now:

- The other banking APIs: payees, direct debits, scheduled payments and instalment plans. Also Get Customer Detail. (See Decisions)
- The amendment of a consent (see Decisions).
- Customers that are organisations, joint accounts, secondary users and nominated representatives (see Decisions).
- Redirect to App and decoupled authentication. The data holder has no app (see Decisions).
- A change of the signing key of the data holder.
- Encryption of ID tokens and of authorisation responses ([Judgement: no encryption of ID tokens or authorisation responses](../../pkg/authserver/design/configuration.md#auth-0017)).
- The admin APIs: Get Metrics and Metadata Update.
- The traffic thresholds of the CDS, and the 429 response for a request above them.
- A CDR policy of the data holder. Rule 7.2(8) requires that each online service of the data holder makes it readily available, also the online bank site (see Decisions).
- Real authentication of a consumer on the online bank site. The site checks no password (see Online bank site).
- An email from the data holder to the consumer, other than the OTP email, for example at the end of a consent. The data holder notifies the data recipient, not the consumer (see Notice to the data recipient and Decisions).
- The brand in the content of the operator console. Only the shell and the link and primary action colours show the brand (see Operator console).
- A run of the conformance suite or of the OpenID conformance suite against a preset brand other than the default preset brand in CI.
- A change of the brand of a running instance. A change needs a restart.
- Other industries.
- Deprecated endpoint versions. The mock serves only the current version of each endpoint (see Decisions).
- Changes to the outcome of Get Status and Get Outages by the operator (see Decisions).
- The mock CDR PKI. It is a separate component with its own design ([PKI DESIGN.md](../../pki/DESIGN.md)).

## Actors

| Actor | Uses |
|---|---|
| Public client | Get Products, Get Product Detail, Get Status, Get Outages |
| Data recipient software product | OpenID Provider configuration, JWKS, registration, token endpoint, registration management, PAR endpoint, introspection endpoint, UserInfo endpoint, token revocation endpoint, CDR arrangement revocation endpoint, the customer and banking APIs of a consent |
| Consumer | Authorisation endpoint, and the authentication and consent pages, in a browser. The online bank site, in a browser (lab-only) |
| Lab operator | Operator interface, operator console, lab-only features |

The data holder calls two other participants: the CDR Register, for its JWKS and the statuses of software products, and the CDR arrangement revocation endpoint of a software product.

Source: [CDS: Banking APIs][cds-get-products], [CDS: Common APIs][cds-get-status], [CDS: Security Endpoints][cds-security-endpoints], [CDS: Consent][cds-consent], [CDS: Metadata Cache Management][cds-metadata-cache].

## Brand

- One instance of the data holder is one data holder brand. Each data holder brand on the Register has its own product reference data, Get Status and Get Outages endpoints. ([CDR Support: White Labelled brands in the CDR][support-white-label])
- The instance gets the identity of its brand and its base URIs from its configuration. Judgement: the lab can then run more than one brand, and a brand can change its host names with no change to the data holder.
- The identity of the brand includes its brand name: the `brandName` of the brand on the Register. The brand directory gives it, and no other setting gives it. The pages for the consumer and for the operator, and the OTP email, show it. (See Brand directory and Decisions)
- The identity of the brand includes its data holder brand ID: the `dataHolderBrandId` of the brand on the Register. It is the client ID of the brand when it calls a software product. ([CDS: Data Holders calling Data Recipients][cds-dh-calling-dr])
- The brand has a record on the Register, with a `publicBaseUri` and a `productBaseUri` that are the base URIs of the instance. A brand with no `productBaseUri` on the Register has no product reference data endpoints. ([CDS: Participant Endpoints][cds-participant-endpoints], [CDS: Register API schemas][cds-register-schemas])
- The `infosecBaseUri` of the brand on the Register is the issuer of the instance. A software product finds the OpenID Provider configuration with it. ([CDS: Participant Endpoints][cds-participant-endpoints])
- The brand has its own issuer. A second brand is a second instance with a different issuer. ([CDS: Multi-Brand Support][cds-multi-brand])
- The `logoUri` of the brand on the Register is the URL of the logo that the instance serves on its public base URI. Judgement: the CDS makes `logoUri` mandatory for a data holder brand ([CDS: Register API schemas][cds-register-schemas]), and a data recipient can then show the logo of the bank that the consumer selects.

### Brand directory

A brand directory gives the parts of the brand that are the same in each deployment. The brand ID, the base URIs, the keys and certificates, and the sender address of the email stay in the configuration of the deployment. (See Decisions)

- The configuration selects the brand with exactly one of two settings: the name of a preset brand, or the path of a brand directory. Neither setting has a default. The data holder does not start when both or neither are present. (See Decisions)
- The preset brands are brand directories that the data holder contains. Each instance can run each preset brand with no file outside the data holder. ([ADR 0018][adr-0018]: the final image contains only the binary)
- A user adds a brand with a brand directory and configuration only, with no change to the code.
- A brand directory has these files, and no other files. Only the photos and the attribution file are optional:

| File | Content |
|---|---|
| `brand.json` | The brand name (`brandName`): 1 to 200 characters, not only white space, and no control characters ([CDS: Register API schemas][cds-register-schemas]: `brandName` has `maxLength: 200`; see Decisions). The BSB (`bsb`) of the accounts that a consumer opens on the online bank site: 6 digits, the rule of a `bsb` in the operator interface (see Online bank site and Decisions). The layout of the online bank site (`layout`): `meadow`, `ledger`, `horizon` or `updraft` (see Online bank site: Presentation and Decisions) |
| `tokens.json` | The values of the brand for the semantic colour roles, the body font, the heading font and the radius of the design system, in the token format of the design system, with literal values only ([design system DESIGN.md: Brands][ui-brands]) |
| `logo.svg` | The logo, legible on the brand colour of the header of the pages and on white ([design system DESIGN.md: Consumer pages][ui-consumer-pages]) |
| `logo.png` | The logo of `logo.svg` as a PNG image, for the email, legible on the brand colour of the header of the email (see The OTP email) |
| `mark.svg` | A square symbol of the brand, for small sizes |
| `products.json` | The product catalogue of the brand, in the format of the product load file (see Lab-only features) |
| `hero.webp` | Optional. The hero photo of the landing page of the online bank site, with the aspect ratio of the layout (see Online bank site: Presentation) |
| `category-<anchor>.webp` | Optional. The photo of a CDS product category on the online bank site, with the anchor of the category, for example `category-term-deposits.webp` ([CDS: Product Categories][cds-product-categories]) |
| `ATTRIBUTION.md` | Optional. The source, the author and the licence of each photo, 16 KiB or less. The data holder does not read or serve it. |

- A JSON file of the directory has only the fields of the contract. The data holder refuses an unknown field and a file that the contract does not name. The contract has no format version. (See Decisions)
- The body font and the heading font of a brand are each one of a fixed set of fonts that the data holder contains. A brand with a different font is a change to the design system, not only a new directory. (See Decisions)
- Each colour of `tokens.json` is in the sRGB gamut. Each pair of text and background roles has a contrast of 4.5:1 or more, and each focus ring and control boundary has a contrast of 3:1 or more against its background. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3], [SC 1.4.11][wcag-1-4-11]; [design system DESIGN.md: Tokens][ui-design])
- Each SVG file is well-formed XML with an `svg` root element and a `viewBox`, and is 100 KiB or less. Judgement: a page shows a logo at different sizes, and a logo needs no more.
- `logo.png` is a valid PNG image of 100 KiB or less, exactly 96 pixels high and 768 pixels wide or less. The data holder checks the size in the header of the file before it decodes the image. Its aspect ratio (width divided by height) is the aspect ratio of the `viewBox` of `logo.svg`, with a tolerance of 2 %. (See Decisions)
- Each JSON file is 1 MiB or less. Judgement: a Kubernetes ConfigMap, which can hold a brand directory, is at most 1 MiB ([Kubernetes: ConfigMaps][k8s-configmap]).
- Each photo is a WebP image that is not animated. `hero.webp` is 256 KiB or less, 1200 to 2400 pixels wide, with the aspect ratio of the layout within 2 %. Each category photo is 96 KiB or less, 800 to 1200 pixels wide, with an aspect ratio of 4:3 within 2 %. The data holder reads the size and the animation flag in the header of the file before it decodes the image. ([RFC 9649][rfc9649]; see Decisions)
- The name of a category photo has the anchor of one of the CDS product categories. The data holder refuses a different name. A photo of a category with no current product is valid. (See Decisions)
- The brand directory is 1 MiB or less in all. Judgement: the limit of each file does not keep a directory with photos in one ConfigMap ([Kubernetes: ConfigMaps][k8s-configmap]; see Decisions).
- The `brand` of each product of the catalogue is the brand name. ([CDR Support: Brands in the CDR ecosystem][support-brands])
- The process that runs the APIs, and the load of products, validate the brand directory at start-up: each file and field of the contract, the gamut of each colour, the fonts, the layout, the SVG files, the PNG file, the photos and the size of the directory. When a rule fails, the process does not start, and its error names the file and the rule. The process that runs only the workers needs no brand setting. (See Non-functional requirements: Least privilege)
- The data holder does not check the contrast of a brand at start-up. A test checks it for each preset brand. (See Decisions)
- The brand name, the product names, the email addresses and the host names of each preset brand and of the lab configuration do not contain "Mock", "Example" or a similar word. An invented domain uses the top-level domain `.test`. ([RFC 6761, section 6.2][rfc6761-test]: `.test` is "for use in testing"; see Decisions)
- By default, the lab runs one preset brand: the default preset brand. It replaces the brand "Mock Bank" of the lab, also on the Register, where the brand has its own legal entity with a realistic name. (See Decisions)
- The lab can also run the four preset brands at the same time. Each is its own instance of the same data holder image, with its own issuer, and the instances differ only in their configuration. ([CDS: Multi-Brand Support][cds-multi-brand]; see Lab-only features)

## Product reference data

- A product has the required fields of the `BankingProductDetailV7` schema, and these optional fields: `effectiveFrom`, `effectiveTo`, `features`, `fees`, `depositRates` and `lendingRates`. The data holder does not hold the other optional fields of the schema now. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; see Decisions)
- A feature has `featureType`, `additionalValue` and `additionalInfo`. A fee has `name`, `feeType`, `feeMethodUType`, `fixedAmount`, `rateBased`, `variable`, `additionalValue` and `additionalInfo`. A deposit rate has `depositRateType`, `rate`, `applicationType`, `applicationFrequency`, `additionalValue` and `additionalInfo`. A lending rate has the same fields, with `lendingRateType`, and also `comparisonRate`, `repaymentType` and `loanPurpose`.
- A field with a CDS enumeration accepts each value of that enumeration.
- A rate is a CDS RateString, and an amount is a CDS AmountString. ([CDS: Common Field Types][cds-common-field-types])
- A product has one of the 13 categories of `BankingProductCategoryV2`. ([CDS: Product Categories][cds-product-categories])
- A `productId` is unique in the brand and has only ASCII characters. It does not follow the ID permanence rules. ([CDS: BankingProductV6][cds-schema-product])
- A product obeys these conditional rules of the schema ([CDS: BankingProductDetailV7][cds-schema-product-detail]):
  - A fee has the object that its `feeMethodUType` names.
  - `additionalInfo` is mandatory when a feature or a fee has the type `OTHER`.
  - A deposit rate or a lending rate with the application type `PERIODIC` has an `applicationFrequency`.
- The `lastUpdated`, `effectiveFrom` and `effectiveTo` of a product are RFC 3339 date-times. ([CDS: Common Field Types][cds-common-field-types])
- The `brand` of a product is the brand of the instance or one of its sub-brands. It is not a different data holder brand on the Register. ([CDR Support: Brands in the CDR ecosystem][support-brands])
- A product is current when the time of the request is at or after its `effectiveFrom` and before its `effectiveTo`. A date that is absent has no limit. A product is future when its `effectiveFrom` is after the time of the request. Judgement: the CDS says only that the filter uses "whether the current time is within the period of time defined as effective" ([CDS: Get Products][cds-get-products]).
- A `productId` that is empty is not valid. Judgement: the schema requires the field, and an empty value cannot identify a product.
- A product `name` has 100 characters or fewer. Judgement: the CDS gives no maximum ([CDS: BankingProductV6][cds-schema-product]), but the `productName` of an account has the rule of 1 to 100 characters (see Operator interface), and an account that a consumer opens gets the name of its product (see Online bank site). The load then refuses each name that an open cannot use. (See Decisions)
- All product data is invented. It does not copy the products of a real data holder. Judgement: the lab is not affiliated with a CDR participant.
- The catalogue of each preset brand is realistic for an Australian retail bank. It has products in five or more product categories, with the fields that the data holder holds now. It also has one future product and one product whose `effectiveTo` is in the past. ([CDS: Product Categories][cds-product-categories]; see Decisions)
- The Get Products response and the Get Product Detail response of each product of each preset brand are valid against the CDS 1.36.0 OpenAPI specification. A test checks each preset brand, also a preset that the lab does not run. ([CDS OpenAPI: banking][cds-openapi-banking]; see Decisions)

## Customers and accounts

The data holder holds only the customer, account and transaction data that the authentication, the consent and the APIs in scope need. (See Decisions)

- A customer is a person. A customer has a customer number, a last name, an optional first name, an email address, the time of its creation and the time of its last change. (See Decisions)
- The operator creates a customer with the operator interface, or the consumer creates one with a sign-up on the online bank site. Both use the same rules. (See Operator interface and Online bank site)
- The customer number and the email address are each a user identifier of the consumer. The consumer can give either of them. ([CDS: User Identifiers][cds-user-identifiers]: an identifier "that can uniquely identify the customer and that is already known by the customer"; see Decisions)
- The data holder creates the customer number. It has 8 decimal digits, and it is unique in the brand.
- The email address is unique in the brand, with no difference for upper and lower case. The data holder keeps the address as the consumer or the operator gives it, and finds it with no difference for upper and lower case. A different brand can have a customer with the same address. (See Decisions)
- The email address is also the channel for the one-time password. ([CDS: One-Time Password Credential Requirements][cds-otp]; [CDR Support: One Time Password (OTP)][support-otp])
- A user identifier that contains `@` is an email address. A different user identifier is a customer number. The data holder removes spaces at the start and at the end of a user identifier. (See Decisions)
- An account has one customer, who is its only owner. An account has a display name, a product category of `BankingProductCategoryV2`, a product name, a BSB of 6 digits, an account number of 6 to 10 digits, an optional credit limit, the time of its creation, a status that is open or closed, and the time of its close when it is closed. ([CDS: Product Categories][cds-product-categories], [CDS: BankingAccountV3][cds-schema-account]; see Decisions)
- An account is open when the operator creates it, or when the consumer opens it on the online bank site. The operator, or the consumer on the site, can close an open account. A closed account cannot open again. (See Decisions)
- A customer is eligible while the customer has an open account. Each account of the lab can be accessed online. ([CDR Rules][cdr-rules], rule 1.10B(1)(b) and Schedule 3, clause 2.1; see The end of a consent)
- The data holder creates an account identifier for each account: a random RFC 4122 UUID. It identifies the account only in the data holder, and a data recipient does not get it. (See Decisions)
- A transaction belongs to one account. It has the fields of `BankingTransactionV2` that the operator interface gives (see Operator interface), and optionally the `extendedData` of `BankingTransactionDetailV3`. A transaction does not change after the operator creates it. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail]; see Decisions)
- The current balance of an account is the sum of the amounts of its posted transactions. The available balance is the credit limit, plus the current balance, plus the amounts of its pending transactions that are negative. When this sum is less than zero, the available balance is zero. ([CDS: BankingBalance][cds-schema-balance]: `currentBalance` is "negative if the customer has money owing" and `availableBalance` is "zero or positive"; see Decisions)
- The pages for the consumer show an account number only as a masked number: the BSB and the account number in their usual format, with each digit except the last 4 as the letter `x`, for example `xxx-xxx xxxx5678`. ([CDS: Common Field Types][cds-common-field-types]: MaskedAccountString)
- All customer and account data is invented. It does not copy the data of a real person. Judgement: the lab must hold no personal information.

## APIs

### Endpoints

| API | Method and path | Version | Base URI | Transport | Auth |
|---|---|---|---|---|---|
| [Get Products][cds-get-products] | `GET /cds-au/v1/banking/products` | 5 | Product | TLS | none |
| [Get Product Detail][cds-get-product-detail] | `GET /cds-au/v1/banking/products/{productId}` | 7 | Product | TLS | none |
| [Get Status][cds-get-status] | `GET /cds-au/v1/discovery/status` | 1 | Public | TLS | none |
| [Get Outages][cds-get-outages] | `GET /cds-au/v1/discovery/outages` | 1 | Public | TLS | none |
| [Get Customer][cds-get-customer] | `GET /cds-au/v1/common/customer` | 1 | Resource | mTLS | Bearer, `common:customer.basic:read`, holder of key |
| [Get Accounts][cds-get-accounts] | `GET /cds-au/v1/banking/accounts` | 3 | Resource | mTLS | Bearer, `bank:accounts.basic:read`, holder of key |
| [Get Account Detail][cds-get-account-detail] | `GET /cds-au/v1/banking/accounts/{accountId}` | 5 | Resource | mTLS | Bearer, `bank:accounts.detail:read`, holder of key |
| [Get Account Balance][cds-get-account-balance] | `GET /cds-au/v1/banking/accounts/{accountId}/balance` | 1 | Resource | mTLS | Bearer, `bank:accounts.basic:read`, holder of key |
| [Get Bulk Balances][cds-get-bulk-balances] | `GET /cds-au/v1/banking/accounts/balances` | 2 | Resource | mTLS | Bearer, `bank:accounts.basic:read`, holder of key |
| [Get Balances For Specific Accounts][cds-get-balances-specific] | `POST /cds-au/v1/banking/accounts/balances` | 1 | Resource | mTLS | Bearer, `bank:accounts.basic:read`, holder of key |
| [Get Transactions For Account][cds-get-transactions] | `GET /cds-au/v1/banking/accounts/{accountId}/transactions` | 2 | Resource | mTLS | Bearer, `bank:transactions:read`, holder of key |
| [Get Transaction Detail][cds-get-transaction-detail] | `GET /cds-au/v1/banking/accounts/{accountId}/transactions/{transactionId}` | 3 | Resource | mTLS | Bearer, `bank:transactions:read`, holder of key |

The security endpoints have no fixed path, except the OpenID Provider configuration. A client gets the URL of each one from the OpenID Provider configuration. ([CDS: Participant Endpoints][cds-participant-endpoints]; see Decisions)

| Endpoint | Method and URL | Transport | Auth |
|---|---|---|---|
| [OpenID Provider configuration][cds-oidc-config] | `GET <issuer>/.well-known/openid-configuration` | TLS | none |
| [JWKS][cds-jwks-endpoint] | `GET`, `jwks_uri` | TLS | none |
| [Token][cds-token-endpoint] | `POST`, `token_endpoint` | mTLS | `private_key_jwt` |
| [PAR][cds-par-endpoint] | `POST`, `pushed_authorization_request_endpoint` | mTLS | `private_key_jwt` |
| [Authorisation][cds-authorisation-endpoint] | `GET`, `authorization_endpoint` | TLS | none. The consumer authenticates in the flow. |
| [Introspection][cds-introspection-endpoint] | `POST`, `introspection_endpoint` | mTLS | `private_key_jwt` |
| [Token revocation][cds-token-revocation] | `POST`, `revocation_endpoint` | mTLS | `private_key_jwt` |
| [CDR arrangement revocation][cds-arrangement-revocation] | `POST`, `cdr_arrangement_revocation_endpoint` | mTLS | `private_key_jwt` |
| [UserInfo][cds-userinfo-endpoint] | `GET` or `POST`, `userinfo_endpoint` | mTLS | Bearer, holder of key |
| [Register a client][cds-dcr-endpoints] | `POST`, `registration_endpoint` | mTLS | none. The request is a signed JWT with an SSA. |
| [Read, update or delete a registration][cds-dcr-endpoints] | `GET`, `PUT` or `DELETE`, `registration_endpoint` and `/{ClientId}` | mTLS | Bearer, `cdr:registration`, holder of key |

- Request and response bodies obey the CDS 1.36.0 OpenAPI specifications ([`cds_banking.json`][cds-openapi-banking], [`cds_common.json`][cds-openapi-common] and [`cds_dcr.json`][cds-openapi-dcr] at v1.36.0), except where a requirement below says otherwise.
- The path of each API follows the CDS URI structure, after the base URI that the brand has on the Register. ([CDS: URI Structure][cds-uri-structure], [CDR Support: Product Base URI][support-product-base-uri])
- The product base URI and the public base URI are the same. ([CDS: Participant Endpoints][cds-participant-endpoints]: the product base URI "MAY be the same as the publicBaseUri"; see Decisions)
- The resource base URI is the base URL of the mTLS listener, and the `resourceBaseUri` of the brand on the Register. ([CDS: Participant Endpoints][cds-participant-endpoints])
- A request on the resource base URI to a CDS banking API that is out of scope, or to Get Customer Detail, gives 404 with `urn:au-cds:error:cds-all:Resource/NotImplemented`, with or without a token. ([CDS: Resource Not Implemented][cds-error-404-not-implemented]: the code "SHOULD be supported for unimplemented APIs"; see Decisions)

### Versioning and headers

- The security endpoints have no version. They do not use `x-v` and `x-min-v`, and the rules for these two headers below do not apply to them. ([CDS: DCR APIs][cds-dcr-apis])
- Each API in the first table of Endpoints uses [the CDS version negotiation](../../pkg/cds/DESIGN.md#cdsl-0001), with the version of the table as its lowest and highest version.
- Each response of the public listener and the mTLS listener has [the `x-fapi-interaction-id` of the request, or a new UUID](../../pkg/cds/DESIGN.md#cdsl-0010).
- The APIs on the resource base URI require `x-fapi-auth-date`, as an HTTP-date. If it is missing, the response is 400 with `urn:au-cds:error:cds-all:Header/Missing`. If it is not an HTTP-date, the response is 400 with `urn:au-cds:error:cds-all:Header/Invalid`. In both cases, the `detail` is the name of the header. ([CDS: HTTP Headers][cds-http-headers]: "Required for all resource calls"; [CDR Support: Format for x-fapi-auth-date][support-auth-date]; see Decisions)
- The data holder accepts `x-fapi-customer-ip-address` with an IPv4 or IPv6 address. A different value gives 400 with `urn:au-cds:error:cds-all:Header/Invalid`. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource]: the server "shall not reject requests" with a valid address; [CDS: Invalid Header][cds-error-400-header-invalid])
- The data holder does not validate `x-cds-client-headers`. ([CDR Support: DH requirement for x-cds-client-headers][support-client-headers]: "A request should not be rejected based on the syntax of this header")
- Each response of an API on the resource base URI has a `Date` header. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource])

### Errors

- The security endpoints and the operator interface have their own error formats (see Authorisation server, Authorisation request, Consumer authentication, Client registration, Registration management and Operator interface). The rules below are for all other requests.
- Error responses use [the CDS error format](../../pkg/cds/DESIGN.md#cdsl-0012).
- A query parameter with a value that is not valid gives 400 with `urn:au-cds:error:cds-all:Field/Invalid`, and the `detail` is the name of the parameter. This includes an `effective` or a `product-category` that is not in its enumeration, and [a `page` that is not a positive integer](../../pkg/cds/DESIGN.md#cdsl-0020). ([CDS: Invalid Field][cds-error-400-field-invalid])
- An `updated-since`, `oldest-time` or `newest-time` that is not an RFC 3339 date-time with an offset gives 400 with `urn:au-cds:error:cds-all:Field/InvalidDateTime`. ([CDS: Invalid Date][cds-error-400-invalid-date-time], [CDS: Common Field Types][cds-common-field-types]: times "MUST be offset relative to UTC")
- A `page-size` that is not valid, and a page out of range, give [the errors of CDS pagination](../../pkg/cds/DESIGN.md#cdsl-0021).
- A query parameter with an empty value is the same as a parameter that is absent. For `page` and `page-size`, [the CDS HTTP library gives this rule](../../pkg/cds/DESIGN.md#cdsl-0022). (See Decisions)
- A `productId` that the brand does not have gives 404 with `urn:au-cds:error:cds-all:Resource/Invalid`, and the `detail` is the `productId`. ([CDS: Invalid Resource][cds-error-404-resource-invalid])
- An `accountId` in the path that the software product cannot get with the token gives 404 with `urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount`, and the `detail` is the `accountId`. This includes an `accountId` that does not exist, an `accountId` of a different customer or software product, and an `accountId` of an account that is not in the arrangement. ([CDS: Invalid Banking Account][cds-error-404-invalid-banking-account]; see Decisions)
- An `accountId` in a request body that the software product cannot get with the token gives 422 with `urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount`. The response has one error for each such `accountId`, with the `accountId` as the `detail`, and no data. ([CDS: Invalid Banking Account][cds-error-422-invalid-banking-account]; [CDR Support: Error scenarios and responses][support-error-scenarios]: no partial response)
- A `transactionId` in the path that is not a transaction of the account, or a transaction with no detail, gives 404 with `urn:au-cds:error:cds-all:Resource/Invalid`, and the `detail` is the `transactionId`. ([CDS: Invalid Resource][cds-error-404-resource-invalid]; [CDR Support: Error scenarios and responses][support-error-scenarios])
- A request body that is not a JSON object with the structure of its schema gives 400 with `urn:au-cds:error:cds-all:Field/Invalid`. A required field that is absent gives 400 with `urn:au-cds:error:cds-all:Field/Missing`. The `detail` is the name of the field, or `data` when the body is not JSON. The data holder ignores a field that the schema does not give. ([CDS: Missing Required Field][cds-error-400-field-missing], [CDS: Invalid Field][cds-error-400-field-invalid]; see Decisions)
- A request body of more than 64 KiB gives 413 with `urn:au-cds:error:cds-all:GeneralError/Expected`. (See Decisions)
- A `POST` request with a `Content-Type` that is not `application/json` gives 415 with `urn:au-cds:error:cds-all:GeneralError/Expected`. ([CDS: HTTP Headers][cds-http-headers]; [CDS: HTTP Response Codes][cds-http-response-codes]; see Decisions)
- An unknown path gives 404 with `urn:au-cds:error:cds-all:Resource/NotFound`. ([CDS: Resource Not Found][cds-error-404-not-found])
- An unexpected server error gives [500 with no cause in the `detail`](../../pkg/cds/DESIGN.md#cdsl-0015).

### Get Products and Get Product Detail

- Get Products returns the products of the brand as `BankingProductV6` records. It supports the filters `effective` (default `CURRENT`), `updated-since`, `brand` and `product-category`. A product must match each filter in the request. ([CDS: Get Products][cds-get-products])
- `effective` gives the current products for `CURRENT`, the future products for `FUTURE`, and all products for `ALL`. `ALL` includes a product whose `effectiveTo` is in the past. ([CDS: Get Products][cds-get-products])
- `updated-since` gives the products with a `lastUpdated` after that time. ([CDS: Get Products][cds-get-products])
- `brand` gives the products whose `brand` is equal to the value. The comparison is exact and case sensitive. (See Decisions)
- Get Products returns the products in descending order of `lastUpdated`. Products with the same `lastUpdated` are in the order of their `productId`. ([CDS: Get Products][cds-get-products]; the second order is judgement: pagination needs a stable order)
- Get Products supports `page` and `page-size` with [the CDS pagination rules](../../pkg/cds/DESIGN.md#cdsl-0017). Each link of a page keeps [the filters of the request](../../pkg/cds/DESIGN.md#cdsl-0029).
- Get Product Detail returns one product as a `BankingProductDetailV7` record, in [the response envelope](../../pkg/cds/DESIGN.md#cdsl-0034). It returns a product that is not current. Judgement: Get Products with `effective=ALL` or `FUTURE` lists such a product, and a client must be able to read the detail of each product that the list gives.

### Get Status and Get Outages

- Get Status returns the status `OK`. The `updateTime` is the time of the request. ([CDS: Get Status][cds-get-status]; see Decisions)
- Get Outages returns an empty list of outages. ([CDS: Get Outages][cds-get-outages]; see Decisions)
- The two responses use [the response envelope with `links.self` and an empty `meta`](../../pkg/cds/DESIGN.md#cdsl-0034).

### CORS

- The TLS endpoints use [the CORS rules](../../pkg/cds/DESIGN.md#cdsl-0037), also for [a preflight request](../../pkg/cds/DESIGN.md#cdsl-0039), except the authorisation endpoint and the pages of its flow. This includes the OpenID Provider configuration and the JWKS, which the CDS says do "not require CORS". ([CDS: CORS][cds-cors], [CDS: OpenID Provider Configuration endpoint][cds-oidc-config]; see Decisions)
- The mTLS endpoints do not permit cross-origin requests. Judgement: they are authenticated endpoints, and a web page in a browser has no CDR client certificate.
- The authorisation endpoint and the pages of its flow do not permit cross-origin requests. Judgement: they are pages for a consumer and not an API, and a different web site must not read them. ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]: the authorisation endpoint does "not require CORS")
- The CORS rules of the TLS endpoints [expose the headers](../../pkg/cds/DESIGN.md#cdsl-0038) `x-v`, `Retry-After` and `x-fapi-interaction-id`, so that a web page can read these headers. ([CDR Support: CORS support][support-cors])

## Authorisation server

The data holder serves the data holder authorisation server of the [authorisation server library](../../pkg/authserver/design/README.md). This section gives only the choices of the data holder.

### OpenID Provider configuration and JWKS

- The issuer is an `https` URL with no query and no fragment. ([OpenID Connect Discovery, section 4.3][oidc-discovery-validation]; see Decisions)
- The OpenID Provider configuration is [below the issuer](../../pkg/authserver/design/configuration.md#auth-0001), and it has [the fields and values of the data holder configuration](../../pkg/authserver/design/configuration.md#auth-0004).
- The supported scopes are `openid`, `profile`, `cdr:registration`, `common:customer.basic:read`, `bank:accounts.basic:read`, `bank:accounts.detail:read` and `bank:transactions:read`. ([CDS: Scopes and Claims][cds-scopes-and-claims], [CDS: Authorisation Scopes][cds-authorisation-scopes]; see Decisions)
- The JWKS has [the public key of each signing key](../../pkg/authserver/design/configuration.md#auth-0006). The data holder gives [each key a different key ID](../../pkg/authserver/design/configuration.md#auth-0008).
- The `jwksEndpoint` of the brand on the Register is the `jwks_uri` of the configuration. Judgement: the CDS permits two URLs, and one key set is the least complex.

### Signing keys

- The data holder signs [with PS256 only](../../pkg/authserver/design/configuration.md#auth-0009), and its keys are [RSA keys of 2048 bits or longer](../../pkg/authserver/design/configuration.md#auth-0012).
- The data holder creates its signing key when it has none. (See Decisions)

### Client authentication

- A software product authenticates with [`private_key_jwt`](../../pkg/authserver/design/client-authentication.md#auth-0020). Its client ID is the client ID that the data holder issued at the registration.
- The data holder gives the `jwks_uri` of the registration to [validate the assertion signature](../../pkg/authserver/design/client-authentication.md#auth-0025).
- The data holder refuses an assertion that breaks a rule of the library, for example [a `jti` that it accepted before](../../pkg/authserver/design/client-authentication.md#auth-0030) or [an `exp` more than 10 minutes ahead](../../pkg/authserver/design/client-authentication.md#auth-0031).

### Token endpoint

- The token endpoint accepts [three grant types](../../pkg/authserver/design/token-endpoint.md#auth-0035). The authorisation server library gives [the rules of the `authorization_code` and `refresh_token` grants](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0201).
- A client credentials token has [only the scope `cdr:registration`](../../pkg/authserver/design/token-endpoint.md#auth-0038), and it [expires 5 minutes after it is issued](../../pkg/authserver/design/token-endpoint.md#auth-0043).
- Each access token is [bound to the client certificate of the token request](../../pkg/authserver/design/token-endpoint.md#auth-0045).
- Errors have [the format and the codes of RFC 6749, section 5.2](../../pkg/authserver/design/token-endpoint.md#auth-0048). A failed client authentication gives [401 `invalid_client`](../../pkg/authserver/design/client-authentication.md#auth-0032).
- The data holder [does not compare the client certificate with the Register data](../../pkg/authserver/design/token-endpoint.md#auth-0054).

## Client registration

### Registration request

- A software product registers with `POST` to the `registration_endpoint`. The body is one compact JWT, and the content type is `application/jwt`. A different content type gives 415 with no body. ([CDS: Registration Request using JWT][cds-registration-request]; [CDS: HTTP Response Codes][cds-http-response-codes]; see Decisions)
- The data holder accepts a request JWT that is signed with PS256 or ES256. ([CDS: Registration Request using JWT][cds-registration-request], [FAPI 1.0 Advanced, section 8.6][fapi-algorithms])
- The request has the claims `iss`, `iat`, `exp`, `jti`, `aud`, `token_endpoint_auth_method`, `token_endpoint_auth_signing_alg`, `grant_types`, `response_types`, `id_token_signed_response_alg`, `authorization_signed_response_alg` and `software_statement`. ([CDS: Registration Request using JWT][cds-registration-request]: `authorization_signed_response_alg` is "Required if response_type of code is registered")
- The `aud` is the issuer, or a list that contains the issuer. The `iss` is the `software_id` of the SSA. ([CDS: Registration Request using JWT][cds-registration-request])
- The data holder refuses a request JWT that is expired, a request JWT whose `exp` is more than 10 minutes after the time of the request, and a request JWT whose `jti` it accepted before. (See Decisions)

### Validation of the SSA

- The data holder verifies the SSA signature with the key at the Register JWKS endpoint that has the `kid` of the SSA. The SSA algorithm is PS256. ([CDS: Registration Validation][cds-registration-validation], [CDS: SSA][cds-ssa])
- The data holder gets the Register JWKS from a URL in its configuration, and trusts only the public CAs in its configuration for that connection. When the SSA has a `kid` that the data holder does not have, the data holder gets the JWKS again before it refuses the SSA, but not more than one time each second. Judgement: the Register can change its key, and a registration must then not need a restart of the data holder. (See Decisions for the limit)
- The SSA `iss` is `cdr-register`, the SSA is not expired, and the SSA has each claim that the SSA definition marks as required. The `software_roles` is `data-recipient-software-product`. ([CDS: SSA Definition][cds-ssa-definition])
- The data holder does not keep the `jti` of an SSA. (See Decisions)
- The data holder does not check the status of the software product on the Register. (See Decisions)

### Validation of the request

- The data holder verifies the signature of the request JWT with the keys at the `jwks_uri` of the SSA. ([CDS: Registration Validation][cds-registration-validation])
- The `jwks_uri` is an `https` URL. The data holder connects to any host, and trusts only the public CAs in its configuration for that connection. It has no allowlist of data recipient hosts. ([CDR Support: Egress policy for calls from DH to ADRs][support-egress]; [CDS: Transaction Security][cds-transaction-security])
- `token_endpoint_auth_method` is `private_key_jwt`. `token_endpoint_auth_signing_alg` and `request_object_signing_alg` are `PS256` or `ES256`. `id_token_signed_response_alg` and `authorization_signed_response_alg` are `PS256`. `application_type`, when present, is `web`. ([CDS: Registration Request using JWT][cds-registration-request]; [Judgement: the signing algorithms](../../pkg/authserver/design/configuration.md#auth-0013))
- Each value of `grant_types` is `client_credentials`, `authorization_code` or `refresh_token`. ([CDS: Registration Request using JWT][cds-registration-request])
- `response_types` contains `code`. The data holder registers only `code`, also when the request has `code id_token`. (See Decisions)
- `redirect_uris`, when present, has only values that are in the `redirect_uris` of the SSA. ([CDS: Registration Request using JWT][cds-registration-request])
- The data holder permits one registration for each `software_id`. ([CDS: Registration Validation][cds-registration-validation])

### Registration

- The data holder issues a new client ID for each registration. It is a random RFC 4122 UUID, and it is not the `software_id`. The data holder does not issue the client ID of a deleted registration again. ([CDS: Registration Response][cds-registration-response]: "dynamically generated identifier"; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed])
- The registration takes these values from the SSA: `client_name`, `client_description`, `client_uri`, `org_id`, `org_name`, `redirect_uris`, `logo_uri`, `jwks_uri`, `revocation_uri`, `recipient_base_uri`, `software_id`, `software_roles` and `scope`, and also `legal_entity_id`, `legal_entity_name`, `sector_identifier_uri`, `tos_uri` and `policy_uri` when the SSA has them. A claim of the request with one of these names has no effect, except `redirect_uris`. ([CDS: Registration Response][cds-registration-response]; [RFC 7591, section 2.3][rfc7591-software-statement]: the value in the software statement "MUST take precedence")
- The `scope` of the registration has the scopes of the SSA that the data holder supports, in the sequence of the SSA. ([CDS: Registration Validation][cds-registration-validation])
- The `redirect_uris` of the registration are those of the request. When the request has none, they are those of the SSA.
- When the request has no `request_object_signing_alg`, the registration has `PS256`. When it has no `application_type`, the registration has `web`. ([CDS: Registration Request using JWT][cds-registration-request]: "data holders are expected to respond with an appropriate default value")
- The data holder ignores each claim of the request that it does not support, with no error. This includes the claims for the encryption of ID tokens and of authorisation responses. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]; [Judgement: no encryption of ID tokens or authorisation responses](../../pkg/authserver/design/configuration.md#auth-0017))

### Registration response

- A successful registration gives 201 with the registration as `application/json`. ([CDS: Registration Response][cds-registration-response], [RFC 7591, section 3.2.1][rfc7591-response])
- The response has `client_id`, `client_id_issued_at`, each value of the registration, and `software_statement` with the SSA of the request, not changed. It has no claim that the data holder ignored. ([CDS: Registration Response][cds-registration-response])
- The response has no `registration_access_token` and no `registration_client_uri`. ([CDS: RegistrationProperties][cds-schema-registration-properties])

### Registration errors

- A registration request that is not valid gives 400 with `application/json` and the members `error` and `error_description`. The `error_description` names the claim or the rule that fails. ([CDS: Registration Errors][cds-registration-errors], [RFC 7591, section 3.2.2][rfc7591-errors])
- The data holder uses these `error` values (see Decisions):

| Condition | `error` |
|---|---|
| The SSA is absent or is not a JWT, its signature is not valid, its `kid` is not in the Register JWKS, it is expired, or one of its claims is absent or not valid | `invalid_software_statement` |
| The `software_id` of the SSA has a registration | `invalid_software_statement` |
| A value of `redirect_uris` is not in the SSA | `invalid_redirect_uri` |
| The body is not a JWT or is larger than 64 KiB, or the request JWT has a signature, an `aud`, an `iss`, an `exp`, a `jti` or a client metadata value that is not valid, or a required claim is absent | `invalid_client_metadata` |
| The data holder cannot get the keys at the `jwks_uri` of the SSA | `invalid_client_metadata` |

- The data holder does not use `unapproved_software_statement`. (See Decisions)
- When the data holder cannot get the Register JWKS, the response is 500 with `server_error`. Judgement: the cause is not the request, and RFC 7591 gives no code for it.

## Registration management

- `GET`, `PUT` and `DELETE` at the `registration_endpoint` and `/{ClientId}` read, replace and delete a registration. ([CDS: Registration Management][cds-registration-management])
- Each request needs an access token as a bearer token ([RFC 6750][rfc6750]). The token is active, has the scope `cdr:registration`, is bound to the client certificate of the request, and is a token of the client in the path. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints], [CDS: Holder of Key Mechanism][cds-hok])
- The responses for a request that the data holder refuses have no body ([`cds_dcr.json`][cds-openapi-dcr]; see Decisions):

| Condition | Response |
|---|---|
| No bearer token | 401 with `WWW-Authenticate: Bearer` |
| The token is not known, is expired, is revoked, is not bound to the client certificate, or is a token of a registration that is deleted | 401 with `WWW-Authenticate: Bearer error="invalid_token"` |
| The path has a client ID that does not exist | 401 with `WWW-Authenticate: Bearer error="invalid_token"` |
| The token does not have the scope `cdr:registration` | 403 with `WWW-Authenticate: Bearer error="insufficient_scope"` |
| The path has the client ID of a different registration | 403 |

- `GET` gives 200 with the registration, in the format of the registration response. The `software_statement` is the SSA of the last request that the data holder accepted. ([CDS: RegistrationProperties][cds-schema-registration-properties])
- `PUT` has the same body as a registration request, and the same validation rules and errors, with a new SSA. The `software_id` of the SSA is that of the registration. If it is not, the response is 400 `invalid_software_statement`. ([CDS: PUT Operations][cds-put-operations])
- `PUT` replaces all values of the registration with the values that the request and its SSA give. The client ID and `client_id_issued_at` do not change. It gives 200 with the registration. ([CDS: PUT Operations][cds-put-operations])
- `DELETE` deletes the registration and gives 204. Client authentication with its client ID then fails. The software product can register again, and it then gets a new client ID. ([CDS: DCR APIs][cds-dcr-apis]; see Decisions)

## Authorisation request

### Pushed authorisation request

- A software product starts an authorisation with `POST` to the PAR endpoint. The body is a form with the client assertion parameters and a `request` parameter, which is a signed request object. ([CDS: Pushed Authorisation endpoint][cds-par-endpoint], [RFC 9126, section 3][rfc9126-request-object])
- The data holder uses only the parameters in the request object. It ignores each other authorisation parameter of the form. ([FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 10)
- A request with a `request_uri` parameter is refused. ([RFC 9126, section 2.1][rfc9126-request])
- The registration of the client has the grant type `authorization_code`. If it does not, the response is 400 `invalid_grant`. (See Decisions)
- A successful request gives 201 with `request_uri` and `expires_in`, and `Cache-Control: no-store`. ([RFC 9126, section 2.2][rfc9126-response])
- The `request_uri` expires 60 seconds after the data holder issues it, and `expires_in` is 60. ([CDS: Request Object Submission][cds-request-object-submission]: "between 10 seconds and 90 seconds"; see Decisions)
- The `request_uri` has a random part that a client cannot guess. ([RFC 9126, section 7.1][rfc9126-guessing])
- Errors follow RFC 9126, section 2.3: 401 `invalid_client` when the client authentication fails, and 400 with `error` and `error_description` for each other error. For a 400 response, the `error_description` names the claim or the rule that fails, and it does not repeat a value of the request object. For a 401 response, it does not give the cause. ([RFC 9126, section 2.3][rfc9126-errors])

### Request object

- The request object is a JWT that the software product signs with the `request_object_signing_alg` of the registration, which is PS256 or ES256. The data holder verifies the signature with the keys at the `jwks_uri` of the registration. ([CDS: Request Object][cds-request-object], [FAPI 1.0 Advanced, section 8.6][fapi-algorithms])
- The `iss` and the `client_id` are the client ID of the authenticated client. The `aud` is the issuer, or a list that contains the issuer. ([FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 15; the CDS example has a different `aud`, and it is not normative)
- The request object has an `nbf` that is not more than 60 minutes before the time of the request, and an `exp` that is not more than 60 minutes after the `nbf`. A request object that is expired is refused. ([CDS: Request Object, Data Holders][cds-request-object-dh], [FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], items 13 and 17)
- `response_type` is `code`. ([CDS: Baseline Security Provisions][cds-baseline-security])
- `response_mode` is `jwt`. (See Decisions)
- `redirect_uri` is present, and it is equal to one of the `redirect_uris` of the registration. ([FAPI 1.0 Baseline, section 5.2.2][fapi-baseline-server], items 9 and 10)
- `scope` contains `openid`. ([CDS: Scopes and Claims][cds-scopes-and-claims]: `openid` "MUST be present on each authentication request")
- Each scope is in the `scope` of the registration, and it is not `cdr:registration`. If not, the response is 400 `invalid_scope`. ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]; [CDS: Authorisation Scopes][cds-authorisation-scopes])
- `nonce` is present. `state` is optional. ([FAPI 1.0 Baseline, section 5.2.2.2][fapi-baseline-nonce])
- `code_challenge` is present, and `code_challenge_method` is `S256`. ([CDS: Request Object, Data Holders][cds-request-object-dh], [RFC 7636, section 4.2][rfc7636-challenge])
- `claims.sharing_duration`, when present, is an integer of seconds that is 0 or more. A negative value or a value that is not an integer is refused. The data holder uses one year (31,536,000 seconds) for a larger value. ([CDS: Requesting Sharing Duration][cds-sharing-duration])
- A request object with `claims.cdr_arrangement_id` is refused. (See Decisions)
- When `claims.id_token.acr` is essential and gives values, one of the values is `urn:cds.au:cdr:2`. If not, the request is refused. ([CDS: Levels of Assurance][cds-loa]; [CDR Support: One Time Password (OTP)][support-otp]; see Decisions)
- The data holder ignores each other member of `claims`, and each other parameter that this document does not name. ([CDS: Claims][cds-claims]: other standard claims "MUST be ignored and not authorised")
- A fault of the JWT (its signature, `iss`, `client_id`, `aud`, `nbf` or `exp`) and a fault of a member of `claims` give 400 `invalid_request_object`. A fault of `response_type`, `response_mode`, `redirect_uri`, `nonce`, `code_challenge` or `code_challenge_method`, and a `scope` with no `openid`, give 400 `invalid_request`. (See Decisions)

### Authorisation endpoint

- The software product sends the browser of the consumer to the authorisation endpoint with `client_id` and `request_uri` in the query. The data holder needs no other parameter, and it ignores each other parameter. ([RFC 9126, section 4][rfc9126-authorisation]; [CDR Support: Pushed Authorization Requests redirect_uri parameter][support-par-redirect-uri])
- A request with a `request` parameter, or with no `request_uri`, is refused. ([CDS: Request Object, Data Holders][cds-request-object-dh])
- The data holder accepts a `request_uri` one time only. It refuses a `request_uri` that it does not know, that is expired, or that it accepted before. ([CDS: Request Object, Data Holders][cds-request-object-dh]: "MUST reject the reuse of request_uri values"; [RFC 9126, section 4][rfc9126-authorisation])
- A `request_uri` of a different client is refused with 401. ([CDR Support: Error scenarios and responses][support-error-scenarios])
- When the data holder refuses a request to the authorisation endpoint, it does not redirect the browser. It shows an error page with the status 400, or with 401 for the condition above and for a client ID that is not registered. The page does not say which rule failed. ([RFC 6749, section 4.1.2.1][rfc6749-authorisation-errors]; see Decisions)
- A flow that ended does not continue: a request for one of its pages is refused in the same way.

## Consumer authentication

The only authentication flow is Redirect to Web with a one-time password (OTP). ([CDS: Fallback Authentication Flows][cds-fallback-flows]; see Decisions)

- The first page asks the consumer for the email address or the customer number, in one field, and for nothing else. ([CDS: User Identifiers][cds-user-identifiers]; see Customers and accounts)
- The page says that the data recipient does not need the password of the consumer to get CDR data, and that the consumer must not give a password. ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Passwords)
- The pages never ask for a password. Judgement: the CDS does not permit a memorised secret as the single factor to achieve "LoA 2" ([CDS: Restricted Credentials][cds-restricted-credentials]), and the sign-in of the online bank site asks for a password, so a page that never asks for one is easy to tell apart (see Online bank site).
- The pages have no link for forgotten details, no offer of other products or services, and no reference to other documents. ([CDS: Redirect to Web with One Time Password][cds-cx-otp], [CDS: Common Authentication Standards][cds-cx-authentication])
- The pages name the data recipient with its brand name, which is the `org_name` of the registration. ([CX Guidelines: Redirect to Web with One Time Password][cx-otp])
- When the consumer gives the email address or the customer number of a customer, the data holder makes an OTP and sends it by email to the email address of that customer. The next page asks for the OTP. A correct OTP authenticates that customer, and the consent is for that customer. ([CDS: One-Time Password Credential Requirements][cds-otp]; see Decisions)
- The email has the OTP, the brand name of the data recipient and the time for which the OTP is valid. It says that the OTP is only for the sharing of CDR data, and that the data recipient does not need the password of the consumer to get CDR data. The subject of the email does not have the OTP, because a device can show a subject on a locked screen (judgement). ([CDS: One-Time Password Credential Requirements][cds-otp], [CDS: Common Authentication Standards][cds-cx-authentication], [CX Guidelines: Redirect to Web with One Time Password][cx-otp]; see The OTP email)
- For an email address or a customer number that no customer has, the data holder sends no email and shows the same OTP page. No OTP is then correct. Judgement: the page then does not show which email addresses and customer numbers exist ([CDS: One-Time Password Credential Requirements][cds-otp]: the data holder "SHOULD implement additional controls to minimise the risk of enumeration attacks").
- An OTP has 6 decimal digits from a cryptographic random source, or is a fixed OTP of a lab-only feature (see Lab-only features). ([CDS: One-Time Password Credential Requirements][cds-otp]: "numeric digits and be between 4 and 6 digits in length")
- An OTP is valid for 5 minutes, for one use, and only in the authorisation flow that made it. The OTP page shows this time. ([CDS: One-Time Password Credential Requirements][cds-otp], [CDS: Redirect to Web with One Time Password][cds-cx-otp]; see Decisions)
- The consumer can ask for a new OTP on the OTP page. A new OTP makes the previous OTP not valid. The data holder sends a maximum of 3 OTPs in one authorisation flow. ([CX Guidelines: Redirect to Web with One Time Password][cx-otp]; see Decisions)
- After 3 incorrect OTP entries in one authorisation flow, the flow ends with the error `access_denied`. ([CDR Support: One Time Password (OTP)][support-otp]; [CDR Support: Error scenarios and responses][support-error-scenarios]; see Decisions)
- Each page has a control to cancel. A cancel ends the flow with the error `access_denied`. ([CDR Support: Error scenarios and responses][support-error-scenarios])
- When the data holder cannot send the email, the flow ends with the error `temporarily_unavailable`. ([RFC 6749, section 4.1.2.1][rfc6749-authorisation-errors])
- An authorisation flow that the consumer does not complete within 10 minutes of the authorisation request ends. The data holder then shows an error page and does not redirect the browser. (See Decisions)
- A correct OTP gives the level of assurance `urn:cds.au:cdr:2`. The time of that entry is the authentication time. ([CDS: Levels of Assurance][cds-loa])

### The OTP email

These rules apply to each OTP email. The data holder sends no other email (see Scope).

- The `From` header has the brand name as its display name and the sender address of the configuration as its address. The header is one valid header line for each brand name that the brand directory permits, also for a name with special characters or characters that are not ASCII. ([RFC 5322, section 3.4][rfc5322-3-4], [RFC 2047, section 5][rfc2047-5]; see Decisions)
- The sender address of the configuration is an address only. The data holder does not start when the setting has a display name, and its error names the setting. Judgement: the brand directory is the only source of the name (see Brand), and an error at start-up is better than an error at the first OTP.
- The email has a plain text part and an HTML part, which give the same information, in a `multipart/alternative` body with the plain text part first. The HTML part and the logo are in a `multipart/related` body, and the HTML part refers to the logo with a `cid:` URL. ([RFC 2046, section 5.1.4][rfc2046-5-1-4], [RFC 2387][rfc2387], [RFC 2392][rfc2392]; see Decisions)
- No line of the message is longer than 998 characters. ([RFC 5322, section 2.1.1][rfc5322-2-1-1])
- The only image is `logo.png` of the brand directory, in the email. The email loads no content from a different location. Judgement: a remote image is a request to the bank each time that the consumer opens the email, and some clients do not show a remote image until the reader permits it ([Gmail Help: Turn images on or off][gmail-images]).
- The header of the HTML part shows the logo on the brand colour. The text alternative of the logo is the brand name. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- The HTML part uses only inline styles, with literal sRGB colours in hex notation from the tokens of the brand, and a system font stack. Each pair of text and background colours is a pair that the contrast test of the brand checks. ([caniemail: style element][caniemail-style], [caniemail: CSS variables][caniemail-css-variables], [caniemail: modern colour functions][caniemail-modern-colour]; [WCAG 2.2, SC 1.4.3][wcag-1-4-3]; see Decisions)
- The OTP is text in the HTML part, not an image, so that the consumer can copy it. Judgement: a consumer who can copy the OTP does not have to transcribe it ([WCAG 2.2, SC 3.3.8][wcag-3-3-8]).
- The email has no link, no button, no offer of other products or services, and no reference to other documents. ([CDS: Common Authentication Standards][cds-cx-authentication], [CX Guidelines: Redirect to Web with One Time Password][cx-otp])
- Both parts have the footer text of the consumer pages, with no link (see Presentation of the consumer pages). Judgement: the email copies the brand of a bank, and it must not deceive a person, as for the pages.
- The HTML part has the language `en-AU` on its `html` element and a title, and each table for layout has the role `presentation`. Judgement: the HTML part is a document that a screen reader reads ([WCAG 2.2, SC 3.1.1][wcag-3-1-1], [SC 1.3.1][wcag-1-3-1]).
- Tests in the data holder check the MIME structure and the headers of the email, also a `From` header with a brand name that has special characters, and the HTML part of each preset brand with the automated accessibility checks. The conformance suite reads only the plain text part, and it does not test the brand. ([ADR 0019][adr-0019])

## Consent

After the authentication, one page asks the consumer for the authorisation to disclose data. ([CDR Rules][cdr-rules], rule 4.5 and rule 4.22)

- The page shows the name of the accredited person: the `legal_entity_name` of the registration, or the `org_name` when the registration has no `legal_entity_name`. It also shows the brand name (`org_name`) and the name of the software product (`client_name`). ([CDR Rules][cdr-rules], rule 4.23(1)(a); [CX Guidelines: Authorisation to disclose][cx-authorise]; [CDS: SSA Definition][cds-ssa-definition]; see Decisions)
- The page shows the types of data. For each scope of the request, it shows the data cluster language and the permission language that the CDS gives, word for word ([CDS: Data Language Standards][cds-data-language], [CDS: Banking Language][cds-banking-language], [CDS: Customer Language][cds-customer-language]):

| Scope | Data cluster | Permissions |
|---|---|---|
| `common:customer.basic:read` | Name and occupation | Name; Occupation |
| `bank:accounts.basic:read` | Account name, type and balance | Name of account; Type of account; Account balance |
| `bank:accounts.detail:read`, with `bank:accounts.basic:read` | Account numbers and features | Account number; Interest rates; Fees; Discounts; Account terms; Account mail address |
| `bank:accounts.detail:read`, without `bank:accounts.basic:read` | Account balance and details | Name of account; Type of account; Account balance; Account number; Interest rates; Fees; Discounts; Account terms; Account mail address |
| `bank:transactions:read` | Transaction details | Incoming and outgoing transactions; Amounts; Dates; Descriptions of transactions; Who you have sent money to and received money from |
| `profile` | Name | Full name and title(s) |

- The page shows no data cluster for `openid`.
- The page says if the disclosure is on one occasion or for a period. For a period, it shows the duration and the date on which the sharing ends. A `sharing_duration` that is absent, 0, or not more than 24 hours is one occasion. ([CDR Rules][cdr-rules], rule 4.23(1)(d) and (e); [CDS: Requesting Sharing Duration][cds-sharing-duration]; [CX Guidelines: Authorisation to disclose][cx-authorise])
- The page shows the duration in days and hours, and as one year for the longest duration. It shows the date in UTC, and it says so. (See Decisions)
- The page says that the data includes the data from before the consent that the data holder holds for the selected accounts. ([CDR Rules][cdr-rules], rule 4.23(1)(b); see Decisions)
- The page says that the consumer can withdraw the authorisation at any time, and how. The text is "You can stop the sharing at any time. Sign in to <brand name> online banking and go to Data sharing, or ask <brand name>." It has no link. ([CDR Rules][cdr-rules], rule 4.23(1)(f) and (g), rule 4.24(d) and rule 4.25(1); see Decisions)
- When the request has a scope with the prefix `bank:`, the page lists each open account of the customer with its display name and its masked number, and the consumer selects one or more accounts. No account is selected before the consumer selects it. The consumer cannot confirm with no account selected. ([CDS: Authorisation Standards][cds-authorisation-standards]: account selection and account confirmation)
- When the request has no scope with the prefix `bank:`, the page has no account selection. ([CDS: Authorisation Standards][cds-authorisation-standards])
- When the customer has no open account, the page says that the customer has no open account at the bank and cannot share data, and the consumer can only cancel. This applies to each request, also a request with no scope with the prefix `bank:`. ([CDR Rules][cdr-rules], rule 4.5(1)(c); [CX Guidelines: Authorisation to disclose][cx-authorise]; see Decisions)
- The page has one control to confirm and one control to cancel. A cancel ends the flow with the error `access_denied`. ([CDR Support: Error scenarios and responses][support-error-scenarios])
- A confirmation with no account, with an account that the customer does not have or that is closed, or of a customer with no open account, creates no arrangement. The data holder shows the page again with the status 400, and the flow continues. (See Decisions)
- The page asks for no other information, offers no other products or services, and has no reference to other documents. ([CDR Rules][cdr-rules], rule 4.24)
- A confirmation creates a CDR arrangement (see CDR arrangement) and ends the flow with an authorisation code.
- All pages of the flow work with no JavaScript, load no content from other hosts, and protect each form against cross-site request forgery. A form that does not come from a page of the flow in the browser that started the flow gives 403, and the flow continues. Their text and controls follow the WCAG success criteria that the CDS names. ([CDS: Accessibility Standards][cds-accessibility]; the first three rules are judgement: the conformance tests and assistive technology can then use the pages, and the pages are an authentication surface)

### Authorisation response

- At the end of a flow, the data holder redirects the browser to the `redirect_uri` of the request object, with one query parameter, `response`. It is a JWT that the data holder signs with PS256. ([CDS: Authorization Code Flow][cds-code-flow], [JARM, section 2.3.1][jarm-query]; see Decisions)
- The JWT has `iss` (the issuer), `aud` (the client ID) and `exp`, which is 10 minutes after the time of the response. For a confirmation it has `code`. For an error it has `error` and `error_description`. It has `state` when the request object has one. ([JARM, section 2.1][jarm-response]: a lifetime of 10 minutes at most is recommended)
- The response is not encrypted. ([CDS: Authorization Code Flow][cds-code-flow]; [Judgement: no encryption of ID tokens or authorisation responses](../../pkg/authserver/design/configuration.md#auth-0017))
- The authorisation code is valid for 60 seconds and for one use. ([RFC 6749, section 4.1.2][rfc6749-code]: 10 minutes at most; see Decisions)

### CDR arrangement

- An arrangement records the customer, the client, the scopes of the request, the account identifier of each selected account, the time of the confirmation and the sharing duration. ([CDS: CDR Arrangement ID][cds-arrangement-id])
- An arrangement also records the names of the data recipient that the consent page showed: the `client_name`, the `org_name` and, when the registration has one, the `legal_entity_name` of the registration at the confirmation. A later change of the registration does not change them. ([CDR Rules][cdr-rules], rule 9.3(1)(a); see Decisions)
- The `cdr_arrangement_id` is a random RFC 4122 UUID. It does not identify the consumer. ([CDS: CDR Arrangement ID][cds-arrangement-id]: "non-guessable and MUST NOT identify a consumer")
- Each confirmation creates a new arrangement. A consumer can have more than one arrangement with one software product. ([CDR Support: Concurrent consent and arrangement ID requirements][support-concurrent-consent])
- The sharing of an arrangement ends at the time of the confirmation plus the sharing duration. An arrangement can end before that time (see The end of a consent).

### Presentation of the consumer pages

These rules apply to each page of the authentication and the consent.

- Each page shows the logo and the name of the brand, and uses the colours of the brand. The text alternative of the logo is the brand name. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]; the brand on the page is judgement: the CDS and the CX Guidelines give no rule for the brand of the data holder on these pages, and a consumer expects to see the bank that they selected)
- Each page names the data holder with its brand name. Where this document says that a page names "the bank", the page gives the brand name. Judgement: the consumer can then see which bank shares the data.
- Each page has a footer that names the brand, says that the bank is not a real bank and that the pages are a demonstration, and says that CDR Lab has no affiliation with a real bank, the ACCC, the DSB or a real CDR participant. The text is "<brand name> is not a real bank and this is just a demonstration. CDR Lab has no affiliation with any real bank nor the ACCC, DSB, or any real CDR participant." The data holder gives the text of the footer, and a brand directory cannot change or remove it. The footer has no link. (Judgement: a realistic copy of a bank page must not deceive a person. [CDS: Common Authentication Standards][cds-cx-authentication]: no reference to other documents and no other services)
- The pages follow the design guide of the design system in `web/packages/ui/`. The outcomes that a test can check are:
  - Each page meets WCAG 2.2 Level AA, and the automated accessibility checks pass. ([ADR 0016][adr-0016])
  - Each page is usable at a viewport width of 320 CSS px with no horizontal scroll. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10])
  - Each page shows the brand and the footer, as the rules above say.
- The pages load styles, fonts and images only from the same origin, and run no script. Judgement: the pages are an authentication surface, and the rule "load no content from other hosts" (see Consent) stays true.
- The data holder serves the logo, the mark and the styles of the brand on the public listener, with no authentication. They contain no CDR data. A response with an SVG file does not permit scripts, also when a browser opens its URL directly. ([MDN: CSP sandbox][mdn-csp-sandbox]; judgement: a user brand directory can contain an SVG file with a script)

## Tokens of a consent

- The authorisation server library gives the tokens of an arrangement: [the authorisation code grant](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0201), [the token response](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0205), [the refresh token grant](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0210) and [the ID token](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0214).
- The data holder gives the library [the subject identifier of each authorisation](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0216). Subject identifier gives its rules.

### Subject identifier

- The `sub` is a pairwise identifier: a random RFC 4122 version 4 UUID that the data holder creates and keeps for each pair of customer and sector identifier. ([CDS: Sub Claim][cds-sub-claim], [OpenID Connect Core, section 8.1][oidc-core-pairwise])
- The sector identifier is the host of the `sector_identifier_uri` of the registration. When the registration has none, it is the host of the `redirect_uris`. When the `redirect_uris` then have more than one host, it is the `software_id`. ([OpenID Connect Core, section 8.1][oidc-core-pairwise]; see Decisions)
- The `sub` of a customer is the same in each arrangement with the same sector identifier, also after the software product registers again. ([CDR Support: ID Permanence and PPID][support-ppid])

## Introspection

- The authorisation server library gives [the rules of the introspection endpoint](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0218).

## Data of a consent

### Access to the data

These rules are for the APIs on the resource base URI and for the UserInfo endpoint.

- A request has the access token in the `Authorization` header with the scheme `Bearer`. The data holder ignores an access token in the query or in the body. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource], [RFC 6750, section 2.1][rfc6750-header])
- The data holder accepts an access token of an arrangement only when the token is not expired and not revoked, its arrangement did not end, and it is bound to the client certificate of the request. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource], [CDS: Holder of Key Mechanism][cds-hok], [RFC 8705, section 3][rfc8705-binding])
- The data holder gives data only to a software product that is ACTIVE on the Register (see Register statuses).
- A client credentials token does not give access to the data of a consent. ([CDS: Authorisation Scopes][cds-authorisation-scopes])
- The data holder returns only the data of the customer of the arrangement, of the accounts in the arrangement, and of the scopes of the token. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource]: "shall only return the resource identified by the combination of the entity implicit in the access and the granted scope")
- The responses for a token that the data holder refuses follow RFC 6750, section 3.1, and the token checks come before the other checks of the request ([CDR Support: Error scenarios and responses][support-error-scenarios]: errors of OAuth "take precedence"; see Decisions):

| Condition | Response |
|---|---|
| No bearer token | 401 with `WWW-Authenticate: Bearer` |
| The token is not known, is expired, is revoked, is of an arrangement that ended, is not bound to the client certificate, or is a client credentials token | 401 with `WWW-Authenticate: Bearer error="invalid_token"` |
| The software product is not ACTIVE on the Register | On the resource base URI, 403 with the CDS error `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`. On the UserInfo endpoint, 401 with `WWW-Authenticate: Bearer error="invalid_token"`. (See Register statuses) |
| The token does not have the scope of the API | 403 with `WWW-Authenticate: Bearer error="insufficient_scope"`. On the resource base URI, the body has the CDS error `urn:au-cds:error:cds-all:Authorisation/InvalidConsent`. |

- The 401 responses, and each response of the UserInfo endpoint for a refused token, have no body.

- Each response uses [the response envelope](../../pkg/cds/DESIGN.md#cdsl-0034), with [a `links.self` on the resource base URI](../../pkg/cds/DESIGN.md#cdsl-0035). A list uses [the CDS pagination rules](../../pkg/cds/DESIGN.md#cdsl-0017).
- The data holder logs the `x-fapi-interaction-id` of each request. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource])

### Resource identifiers

- An `accountId` or a `transactionId` is a version 8 UUID ([RFC 9562, section 5.8][rfc9562-v8]) from a SHA-256 hash of the `software_id`, the customer number and the identifier of the account or the transaction in the data holder. Neither has meaning, and neither is an identifier of the operator interface. ([CDS: ID Permanence][cds-id-permanence]; see Decisions)
- The identifiers are for each pair of customer and software product. The software product gets the same identifier in each arrangement of the customer, also after it registers again. A different software product or a different customer gets a different identifier for the same account or transaction. ([CDS: ID Permanence][cds-id-permanence]: IDs "MUST be immutable across sessions and consents but MUST NOT be transferable across Data Recipient Software Products"; [CDR Support: ID Permanence and PPID][support-ppid])

### Get Customer

- Get Customer returns the customer of the arrangement with `customerUType` `person` and a `CommonPerson` object. ([CDS: Get Customer][cds-get-customer])
- The person has `lastName`, `firstName` when the customer has one, `middleNames` as an empty list, and `lastUpdateTime`, which is the time of the last change to the customer. It has no `prefix`, `suffix` or occupation. ([CDS: CommonPerson][cds-schema-person]; see Decisions)

### Get Accounts and Get Account Detail

- Get Accounts returns the accounts of the arrangement as `BankingAccountV3` records, in the sequence of their creation. It supports the filters `product-category`, `open-status` (default `ALL`) and `is-owned`. An account must match each filter in the request. ([CDS: Get Accounts][cds-get-accounts])
- An account has `accountId`, `displayName`, `maskedNumber`, `productCategory`, `productName`, `openStatus`, `isOwned` with the value `true`, `accountOwnership` with the value `ONE_PARTY`, and `creationDate`. ([CDS: BankingAccountV3][cds-schema-account])
- A closed account stays in the data of an arrangement that has it, with `openStatus` `CLOSED`. (See Decisions)
- Get Account Detail returns one account of the arrangement as a `BankingAccountDetailV5` record. It adds `bsb` and `accountNumber`, with digits only. It has no other optional field. ([CDS: Get Account Detail][cds-get-account-detail]; see Decisions)

### Balances

- Get Account Balance returns the balance of one account of the arrangement. Get Bulk Balances returns the balances of the accounts of the arrangement, with the filters of Get Accounts. Get Balances For Specific Accounts returns the balances of the accounts in `data.accountIds`. ([CDS: Get Account Balance][cds-get-account-balance], [CDS: Get Bulk Balances][cds-get-bulk-balances], [CDS: Get Balances For Specific Accounts][cds-get-balances-specific])
- A balance has `accountId`, `currentBalance`, `availableBalance` and `creditLimit`, from Customers and accounts. It has no `amortisedLimit`, `currency` or `purses`. ([CDS: BankingBalance][cds-schema-balance]: the currency default is `AUD`)
- Get Balances For Specific Accounts gives one balance for each different `accountId` in `data.accountIds`, in the sequence of Get Accounts, with pagination. An empty `data.accountIds` gives an empty list. ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]; see Decisions)

### Transactions

- Get Transactions For Account returns the transactions of one account of the arrangement as `BankingTransactionV2` records. The response has [`meta.isQueryParamUnsupported` with the value `false`](../../pkg/cds/DESIGN.md#cdsl-0032). ([CDS: Get Transactions For Account][cds-get-transactions])
- The effective time of a transaction is its `postingDateTime` when it has one, and its `executionDateTime` when it does not. The list is in descending order of the effective time. Transactions with the same effective time are in descending order of their creation. ([CDS: Get Transactions For Account][cds-get-transactions]; the second order is judgement: pagination needs a stable order)
- The filters are `oldest-time` (default: `newest-time` minus 90 days), `newest-time` (default: the time of the request), `min-amount`, `max-amount` and `text`. A transaction must match each filter in the request. The time filters include their limits, and they use the effective time. `text` gives the transactions whose `description` or `reference` contains the value, with no difference between upper case and lower case. ([CDS: Get Transactions For Account][cds-get-transactions]; see Decisions)
- The data holder gives only the transactions whose effective time is not more than 2 years before the time of the request. An `oldest-time` before that time is not an error. ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(6)(a); see Decisions)
- A transaction has `isDetailAvailable` with the value `true` only when it has `extendedData`. ([CDS: BankingTransactionV2][cds-schema-transaction])
- Get Transaction Detail returns one transaction of the account as a `BankingTransactionDetailV3` record, with its `extendedData`. The 2-year rule applies. ([CDS: Get Transaction Detail][cds-get-transaction-detail])

### UserInfo

- The UserInfo endpoint accepts `GET` and `POST`. It returns `sub`, and for a token with the scope `profile` also `name`, `given_name` (when the customer has a first name), `family_name` and `updated_at`. The `sub` is that of the ID token of the arrangement. It does not return `auth_time`. The response is JSON, and it is not signed or encrypted. ([CDS: UserInfo endpoint][cds-userinfo-endpoint], [CDS: Scopes and Claims][cds-scopes-and-claims], [OpenID Connect Core, section 5.3][oidc-core-userinfo]; see Decisions)
- The token needs the scope `openid`. ([OpenID Connect Core, section 5.3][oidc-core-userinfo])

### Record of disclosures

- The data holder records each disclosure of CDR data: one record for each response with a 2xx status of an API on the resource base URI or of the UserInfo endpoint, for an access token of an arrangement. ([CDR Rules][cdr-rules], rule 7.9(1) and rule 9.3(1)(d); see Decisions)
- A record has the `cdr_arrangement_id`, the data cluster and the time of the response. The record identifies the data cluster with the scope of the API, and the data cluster is that of the scope in the table of Consent. For an API with the scope `bank:accounts.detail:read`, the data cluster depends on the scopes of the arrangement, as on the consent page. For the UserInfo endpoint, the data cluster is that of `profile`. A response of the UserInfo endpoint for a token with no `profile` scope gives no record, because it discloses only the `sub`. ([CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters]; see Decisions)
- A 2xx response with no records, for example no transactions in the period, is a disclosure. (See Decisions)
- The data holder writes the record before it sends a byte of the response. When it cannot write the record, the response is 500 with no CDR data. ([CDR Rules][cdr-rules], rule 9.3(1)(d); see Decisions)
- A response that is not 2xx gives no record. A response to `HEAD` gives no record, because it has no body. ([RFC 9110, section 9.3.2][rfc9110-head])
- The consumer dashboard shows a disclosure at the next read after the response. ([CDR Rules][cdr-rules], rule 7.9(1): "as soon as practicable")
- A disclosure is not an event. The data holder does not remove a record. (See Decisions)

## The end of a consent

### End of an arrangement

- An arrangement ends at the earliest of these events ([CDR Rules][cdr-rules], rule 4.26(1) and (2); [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation] for the notice):

| Event | Time of the end | Notice to the data recipient |
|---|---|---|
| The sharing of the arrangement ends (see CDR arrangement) | The end of the sharing | No |
| The software product revokes the arrangement | The time of the request | No |
| The consumer withdraws the consent on the consumer dashboard, or the operator withdraws it for the consumer | The time of the request | Yes |
| The customer has no open account | The close of the last open account | Yes |
| The software product becomes REMOVED on the Register | The time at which the data holder gets the status | No |

- The sharing duration is one year at most, so the sharing of an arrangement always ends before the limit of 12 months of rule 4.26(1)(e). ([CDS: Requesting Sharing Duration][cds-sharing-duration])
- An arrangement that ended does not start again.
- When an arrangement ends, its refresh token and its access tokens are revoked at the same time. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]: "Consent revocation MUST also revoke associated refresh and/or access tokens")
- The data holder records the time and the event of each end before the end of the sharing. ([CDR Rules][cdr-rules], rule 9.3(1))

### Token revocation

- The authorisation server library gives [the rules of the token revocation endpoint](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0222). A token revocation [does not end the arrangement](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0223).

### CDR arrangement revocation

- The authorisation server library gives [the rules of the CDR arrangement revocation endpoint](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0228).
- The data holder [ends the arrangement and revokes its tokens before the response](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0229) (see End of an arrangement).
- An arrangement with the sharing duration 0 ended at the confirmation. Its revocation thus gives [422 with `Authorisation/InvalidArrangement`](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0230).

### Withdrawal by the consumer

- The consumer can withdraw a consent with two methods ([CDR Rules][cdr-rules], rule 4.25(1); [CDR Support: Withdrawal of authorisation to disclose CDR data and notification][support-withdrawal-notification]: a dashboard and a different method "must always" be available; see Decisions):
  - On the consumer dashboard of the online bank site (see Consumer dashboard). ([CDR Rules][cdr-rules], rule 4.25(1)(a))
  - When they ask the bank. The operator then withdraws it with the operator interface or the operator console (see Operator interface and Operator console). This is the "simple alternative method of communication". ([CDR Rules][cdr-rules], rule 4.25(1)(b))
- A withdrawal has its effect at the time of the request: of the consumer on the dashboard, or of the operator. ([CDR Rules][cdr-rules], rule 4.25(2), rule 4.26(1)(a) and (b))
- The two methods have the same effects: the end of the arrangement, the revocation of its tokens, and the notice to the data recipient (see End of an arrangement and Notice to the data recipient).
- The data holder records the method of each withdrawal by the consumer. A withdrawal on the consumer dashboard has the method `CONSUMER_DASHBOARD`. A withdrawal through the operator interface or the operator console has the method `ALTERNATIVE_METHOD`. This includes each withdrawal before the data holder recorded the method, because each one was through the operator interface. ([CDR Rules][cdr-rules], rule 9.3(1)(b); [CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method]: "Both DHs and ADRs are expected to record how the withdrawal was requested by the consumer"; see Decisions)

### End of eligibility

- When the operator, or the consumer on the online bank site, closes the last open account of a customer, each arrangement of the customer that did not end, ends. This includes an arrangement that has no account. ([CDR Rules][cdr-rules], rule 4.26(1)(c); [CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts])
- The close of an account does not end an arrangement while the customer has a different open account. ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]: "existing authorisations with the data holder remain in place")
- The arrangements end before the response to the close. Their time of the end is the time of the close.
- Two closes at the same time of the last two open accounts of a customer end the arrangements of the customer.
- When the data holder stops after the close and before the arrangements end, the arrangements end after the next start, with the time of the close. (See Decisions)

### Notice to the data recipient

- For the events of the table in End of an arrangement that need a notice, the data holder notifies the software product of the arrangement. ([CDR Rules][cdr-rules], rule 4.26A; [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation])
- The notice is a `POST` to `<recipient_base_uri>/arrangements/revoke`, where `<recipient_base_uri>` is the value of the current registration of the client of the arrangement. When the client has no registration, it is the value of the registration at the confirmation. (See Decisions) The connection is TLS. The data holder trusts only the public CAs in its configuration for it, and it connects to any host, as for a `jwks_uri`. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation], [CDR Support: Egress policy for calls from DH to ADRs][support-egress])
- The request has `Authorization: Bearer` with a self-signed JWT that the data holder signs with PS256 and a key of its JWKS. The JWT has `iss` and `sub` with the data holder brand ID, `aud` with the URL of the endpoint, a `jti` that is a new random RFC 4122 UUID, `iat`, and `exp` 5 minutes after `iat`. ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt]; see Decisions)
- The body is a form with only `cdr_arrangement_jwt`: a JWT with `cdr_arrangement_id` and the claims of the bearer JWT, with a different `jti`, signed in the same way. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; see Decisions)
- A response of 204 or 422 completes the notice. The data holder tries again after a 5xx response, after no response within 10 seconds, and when it cannot connect. Each other response makes the notice fail, with no new attempt. (See Decisions)
- The data holder makes a maximum of 3 attempts. The second attempt is 10 seconds after the first, and the third is 60 seconds after the second. After the third attempt, the notice fails. ([CDR Support: Data Recipient unavailability back-off pattern][support-backoff]; see Decisions)
- The arrangement ends before the first attempt. The notice does not change the end. ([CDR Rules][cdr-rules], rule 4.26(1))
- A notice that is not complete and did not fail when the data holder stops continues its attempts after the next start. The total stays at 3 attempts. (See Decisions)
- The data holder records each notice, its outcome and time, and the time and the result of each attempt. A result is the status of the response, `timeout`, `no connection`, `tls error`, or `internal error` when the data holder cannot make the request, and it contains no value of the response. ([CDR Rules][cdr-rules], rule 9.3(1); see Decisions)

## Register statuses

- The data holder gets the statuses of the software products from [Get Software Products Statuses][cds-get-sp-statuses] of the Register, with the `{industry}` `all`. It gets the base URI of the public APIs of the Register from its configuration, and it trusts only the public CAs in its configuration for that connection. ([CDS: Metadata Cache Management][cds-metadata-cache]; see Decisions)
- The software product of a registration is the one with the `software_id` of the registration.
- The data holder gets the statuses at an interval from its configuration. The interval is 4 minutes at most, so that a status change on the Register has its effect at the data holder within 5 minutes. The request has `If-None-Match` with the `ETag` of the last response. ([CDS: Metadata Cache Management][cds-metadata-cache])
- When the Register does not respond, or its response has no status or a status that is not known for a software product, the data holder keeps the previous status, with no time limit. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- A software product that has no status at the data holder gets the responses for INACTIVE. The readiness probe fails until the data holder gets the statuses one time. (See Decisions)
- The data holder reacts to the status of a software product as follows. The status of the data recipient does not change these reactions. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; see Decisions)

| Request or item | ACTIVE | INACTIVE |
|---|---|---|
| APIs on the resource base URI | Permitted | 403 with `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`, and the `detail` is the status of the software product |
| UserInfo endpoint | Permitted | 401 with `WWW-Authenticate: Bearer error="invalid_token"` and no body |
| PAR endpoint | Permitted | 400 `unauthorized_client` |
| Authorisation endpoint, and the confirmation of a consent | Permitted | The flow ends with the error `unauthorized_client` (see Authorisation response), and the data holder creates no arrangement |
| Token endpoint, `authorization_code` and `refresh_token` grants | Permitted | 400 `unauthorized_client`. The refresh token does not change. |
| Token endpoint, `client_credentials` grant; registration management; introspection endpoint; token revocation endpoint; CDR arrangement revocation endpoint | Permitted | Permitted |
| Arrangements | Do not change | Do not change. Sharing starts again when the software product is ACTIVE again. |

- On the resource base URI and the UserInfo endpoint, the status check comes after the check of the token, and before the other checks of the request. ([CDR Support: Error scenarios and responses][support-error-scenarios]: "ensure the Access Token (AT) is valid before validating consent")
- At the PAR and token endpoints, the status check comes after the client authentication. At these endpoints and at the authorisation endpoint, other checks of the request can come before the status check. (See Decisions)
- When the data holder gets the status REMOVED for a software product, it ends each arrangement of each registration of that software product at once, and deletes each such registration. It sends no notice. Client authentication with the client ID then fails, and the tokens of the client are not accepted, as after a `DELETE` of the registration. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]; see Decisions)
- A software product that is REMOVED cannot register again, because the Register gives an SSA only to an ACTIVE software product. ([CDS: SSA][cds-ssa])

## Transport security

- All CDR endpoints use [the TLS versions](../../pkg/cds/DESIGN.md#cdsl-0041) and [the TLS 1.2 cipher suites](../../pkg/cds/DESIGN.md#cdsl-0042) of the CDS HTTP library.
- The TLS endpoints present a certificate from a mock public CA, not from the mock CDR CA. They [do not request client certificates](../../pkg/cds/DESIGN.md#cdsl-0043). ([CDS: Participant Endpoints][cds-participant-endpoints], [CDS: Transaction Security][cds-transaction-security])
- The mTLS endpoints are on a listener that is separate from the TLS endpoints. They present a server certificate from the mock CDR CA, and they accept only [client certificates that chain to the CDR CA bundle](../../pkg/cds/DESIGN.md#cdsl-0044). ([CDS: Certificate Trust Model][cds-cert-trust-model])
- The data holder checks the revocation status of each client certificate with [the CRL check of the CDS HTTP library](../../pkg/cds/DESIGN.md#cdsl-0048), and [refuses the certificate when the check fails](../../pkg/cds/DESIGN.md#cdsl-0047).
- The [maximum CRL age](../../pkg/cds/DESIGN.md#cdsl-0052) is a setting of the data holder. The setting is optional, and [its default is 1 minute](../../pkg/cds/DESIGN.md#cdsl-0053). (See Decisions)
- The data holder gets the base URL of the mTLS listener, its certificate and key, and the CA bundle for client certificates from its configuration.

## Operator interface

A production data holder gets its customers, accounts and transactions from its banking systems. The operator interface creates them in the lab. Real components must not depend on it.

- The operator can create a customer, create an account for a customer, close an account, create a transaction for an account, read a customer with its accounts, read the transactions of an account, read the arrangements of a customer, and withdraw a consent for a customer.
- The operator interface is on a separate listener from the CDR APIs. It does not permit cross-origin requests.
- Each request must have the operator token as a bearer token ([RFC 6750][rfc6750]). This includes a request for an unknown path. A request with no bearer token gets 401 with `WWW-Authenticate: Bearer`. A request with a different bearer token gets 401 with `WWW-Authenticate: Bearer error="invalid_token"`. ([RFC 6750, section 3.1][rfc6750-errors])
- The operator interface refuses each request, with each method and path, whose `Host` header is not in an allowlist. The allowlist is a setting with the values of the `Host` header: a host and a port, for example `127.0.0.1:8182`, or a host with no port for a client that uses the default port of its scheme, for example through a TLS proxy ([RFC 9110, section 7.2][rfc9110-host]). The setting is required, and has no default. The comparison of the host name has no difference for upper and lower case. The response is 403 with a problem details body, and the interface does not examine the token. ([Vite advisory GHSA-vg6x-rcgg-rjx6][vite-dns-rebinding]: a plain HTTP server with no check of the `Host` header is open to DNS rebinding. See Decisions.)
- The data holder gets the operator token from its configuration.
- Request and response bodies are JSON. A body that is not valid gives 400, and the `detail` names each field and rule that fails. This includes a body with a field that the table below does not give. The `detail` does not repeat a value of the request. A customer that does not exist gives 404. Error responses are [RFC 9457][rfc9457] problem details with the type `about:blank`.
- A customer, an account and a transaction cannot be removed. A customer and a transaction cannot be changed. The only change to an account is to close it. The operator cannot create an arrangement, and the only change to an arrangement is a withdrawal. (See Decisions)

| Action | Method and path | Request body | Success |
|---|---|---|---|
| Create a customer | `POST /v1/customers` | `lastName`, `emailAddress`, and optionally `firstName` | 201 with the customer, which has its `customerNumber` |
| Read a customer | `GET /v1/customers/{customerNumber}` | none | 200 with the customer and its `accounts`, in the sequence of their creation |
| Create an account | `POST /v1/customers/{customerNumber}/accounts` | `displayName`, `productCategory`, `productName`, `bsb`, `accountNumber`, and optionally `creditLimit` | 201 with the account, which has its `accountId` |
| Close an account | `POST /v1/customers/{customerNumber}/accounts/{accountId}/close` | none | 200 with the account |
| Create a transaction | `POST /v1/customers/{customerNumber}/accounts/{accountId}/transactions` | `type`, `status`, `description`, `amount`, and optionally `reference`, `postingDateTime`, `executionDateTime`, `valueDateTime`, `merchantName`, `merchantCategoryCode`, `billerCode`, `billerName`, `crn`, `apcaNumber` and `extendedData` | 201 with the transaction, which has its `transactionId` |
| Read the transactions of an account | `GET /v1/customers/{customerNumber}/accounts/{accountId}/transactions` | none | 200 with the transactions, in the sequence of their creation |
| Read the arrangements of a customer | `GET /v1/customers/{customerNumber}/arrangements` | none | 200 with the arrangements, in the sequence of their creation |
| Withdraw a consent | `POST /v1/customers/{customerNumber}/arrangements/{arrangementId}/withdraw` | none | 200 with the arrangement |

- A `lastName`, a `displayName` and a `productName` have 1 to 100 characters. A `firstName` has 100 characters at most. An `emailAddress` is one address of [RFC 5321][rfc5321-mailbox]. A customer with an `emailAddress` that a different customer of the brand has, with no difference for upper and lower case, gives 409, and the `detail` does not contain the address (see Customers and accounts). A `productCategory` is a value of `BankingProductCategoryV2`. A `bsb` has 6 digits, and an `accountNumber` has 6 to 10 digits. The pair of `bsb` and `accountNumber` is unique in the brand. A second account with the same pair gives 409. A `creditLimit` is an AmountString with exactly 2 decimal digits that is zero or more. When it is absent, it is `0.00`. (See Decisions)
- The transaction fields have the names, types and enumerations of `BankingTransactionV2` and `BankingTransactionDetailV3`. A `description` has 1 to 500 characters. A `reference` that is absent is an empty string. A `POSTED` transaction has a `postingDateTime`. A `PENDING` transaction has an `executionDateTime`. `extendedData` has `service` and `serviceVersion`, `payer` for a positive amount, `payee` for a negative amount, and `extendedDescription`. ([CDS: BankingTransactionDetailV3][cds-schema-transaction-detail]; see Decisions)
- An account or a transaction of a different customer gives 404. A transaction for a closed account, and a close of a closed account, give 409. The `transactionId` of the operator interface is not the `transactionId` of the banking APIs.
- An arrangement has `arrangementId` (its `cdr_arrangement_id`), `softwareId`, `clientName` (the `client_name` of the registration at the confirmation; see Decisions), `scopes`, `accountIds` (the account identifiers of the operator interface), `confirmedAt` and `sharingEndsAt`. An arrangement that ended before `sharingEndsAt` also has `endedAt` and `endReason`: `REVOKED_BY_DATA_RECIPIENT`, `WITHDRAWN_BY_CONSUMER`, `CONSUMER_NOT_ELIGIBLE` or `SOFTWARE_PRODUCT_REMOVED`. An arrangement with the end reason `WITHDRAWN_BY_CONSUMER` also has `withdrawalMethod`: `CONSUMER_DASHBOARD` or `ALTERNATIVE_METHOD`. An arrangement with a notice also has `noticeStatus`: `PENDING`, `COMPLETE` or `FAILED`.
- An arrangement of a different customer gives 404. A withdrawal of an arrangement that ended gives 409.

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

The console does the actions of the operator interface, with the same rules ([Operator interface](#operator-interface)):

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
- The title of the browser tab gives the brand name, and the icon of the tab is the mark of the brand (see Brand).
- The search, the sort and the page of a list are in the URL. A search replaces the history entry of the list, and a sort and a page add a history entry. The detail of a customer, the transactions of an account and the detail of a registration each have their own URL. The page of the arrangements of a registration is in the URL of the detail of the registration. Judgement: the same as the Register console.
- The console shows each time in the time zone of the browser, with its offset from UTC (see the format of a time in the [design system DESIGN.md][ui-design]). The API gives each time in UTC. ([ADR 0012][adr-0012])
- The console shows each URI of a registration as text with a copy button, not as a link, and loads no file from it. ([OWASP ASVS 5.0, V3.6.1][asvs-v3])
- An amount shows with its sign and 2 decimal digits. A pending transaction shows as pending.
- A BSB shows as two groups of three digits with a hyphen, for example `062-000`. Its copy button copies the 6 digits. (See Decisions.)
- The console has the accessibility level of an operator console in the design system: the automated checks pass, with no other target.

### Listener and security

The console listener has the rules of the Register console listener ([Register DESIGN.md: Listener and security][register-console-listener]): its own listener, off by default, plain HTTP, no login and no credential in the browser, the refusal of cross-origin changes, the `Host` allowlist, no CORS headers, the response headers, the cache headers and the HTML page for a path of the UI.

- The console listener serves the logo and the mark of the brand, and a stylesheet that sets only the roles of the brand that the console uses, with the `Host` allowlist and the response headers of the console. Judgement: the Content Security Policy of the console permits images and styles from the same origin only, and the brand files of the public listener are on a different origin. (See Decisions.)
- The data holder starts the console listener only in the process that runs the APIs. A process that runs only the workers needs no console setting. (See Non-functional requirements: Least privilege.)

### Console API

The console API has the rules of the Register console API ([Register DESIGN.md: Console API][register-console-api]) for: the `/api/` prefix, the OpenAPI document and its tests, the problem details and their status codes, an unknown path, the pages and their sizes, a query parameter that is not known or empty, no version in the path, a body that is not valid, the field errors with a JSON Pointer and their problem type, and no request body in telemetry. These rules are different:

- The OpenAPI document is in the data holder directory, and it defines its own problem type for field errors. (See Decisions.)
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
- When the data holder refuses a value of a create, the response is 400, as in the operator interface, and it has a field error for each value that the rules of the [Operator interface](#operator-interface) refuse. This includes a value with no valid format, for example a `creditLimit` or an `amount` that is not an AmountString, or a time that is not an RFC 3339 date-time. The `detail` of each field error is the text that the `detail` of the operator interface gives for that rule. A rule of two fields is at the field that is absent. (See Decisions.)
- A create of an account with the `bsb` and the `accountNumber` of a different account gives 409, as in the operator interface, with one field error at `accountNumber`. The `detail` of the field error is the `detail` of the operator interface. The data holder changes nothing. The operator interface gives this 409 with no field error. (See Decisions.)
- An unknown customer, account, arrangement or registration gives 404, and the `detail` does not contain its identifier.
- The fields have the names of the operator interface. The client registrations have the names of the registration metadata ([CDS: Dynamic Client Registration][cds-dcr-endpoints]).
- The overview gives the counts and the 10 newest changes in one response, from one read of the database, so that the counts and the changes agree with each other. The changes are newest first. Each change has `time` (UTC), `kind`, and, when the kind has them, `customerNumber`, `clientId`, `clientName`, `clientRegistered` (if a registration has the client ID now), `productName`, `sharingEndsAt`, `endReason` and `withdrawalMethod`. The response has no name or email address of a customer, no display name of an account and no account number. (See Decisions.)
- The console API puts no query string and no path value in telemetry. A path with an identifier uses its route pattern. (See Non-functional requirements: Telemetry.)

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

- Telemetry: the console API has the same server telemetry as the other listeners, and the rules of the data holder telemetry apply to it. A search of the customers has a customer number, a name or an email address in its query string, so no span, metric or log contains the query string. The UI sends no telemetry. ([ADR 0014][adr-0014], [ADR 0006][adr-0006])
- Performance: on a development workstation, a search of 1,000 customers shows within 1 second, and the first page of 10,000 transactions of an account shows within 1 second. Judgement: the same limits as the Register console.
- Tests: tests in the data holder check the listener, the console API, and that the telemetry of a search and of each path contains no customer number, name or email address. A small number of browser tests check the main operator journeys against the lab. A browser test creates records through the operator interface and checks that each count of the overview changes by the number of records that it creates, so that it passes in a new lab and in a lab with other data. The conformance suite does not test the console. ([ADR 0019][adr-0019])

## Online bank site

The online bank site is the web site of the bank for its consumers, in the brand of the instance. Now it has a landing page, the list and the detail of the current products of the brand, sign-up, sign-in and sign-out, the accounts of the consumer with their balances and transactions, the opening and the close of an account by the consumer, and the consumer dashboard. It is a lab-only feature: real components must not depend on it, and it is not the web site of a production bank. The architecture is in [ADR 0014][adr-0014], the design system in the [design system DESIGN.md][ui-design], and the contract of its API in [ADR 0017][adr-0017].

The site is not part of the CDR authentication flow. The authentication and consent pages stay on the public listener, and the site does not change them (see Consumer authentication, Presentation of the consumer pages and Separation from the CDR authentication).

### Pages

- The landing page has:
  - A heading that names the brand, one sentence about the bank, and a link to the product list. The data holder gives this text, with the brand name in it. A brand directory does not change it. (See Decisions)
  - One link for each product category that has a current product, with the name of the category and the number of its current products. Each link goes to the group of that category in the product list. Judgement: a consumer looks for a type of product first.
  - A short statement that the bank shares the data of a consumer under the Consumer Data Right only with the consent of the consumer. Judgement: the site is where a consumer can learn about the CDR sharing of the bank, and the consumer dashboard is on the site.
- The text of the site makes no claim about rates, awards or customers, and copies no text of a real bank. Judgement: the bank is invented, and the text must not deceive a person (see Presentation of the consumer pages).
- The product list shows each current product of the brand, and no future product and no product whose `effectiveTo` is in the past. ([CDS: Get Products][cds-get-products]: `effective` is `CURRENT` by default; see Product reference data for a current product, and Decisions)
- The product list puts the products in groups by `productCategory`, in the order of the categories in the CDS. Only a category with a current product has a group. In a group, the products are in the order of their names. ([CDS: Product Categories][cds-product-categories]; the order in a group is judgement: a consumer finds a product by its name)
- For each product, the product list shows its name and its description, and links to the detail of the product. Judgement: the name and the description are the required fields that a consumer reads ([CDS: BankingProductV6][cds-schema-product]).
- The detail of a product shows each field that the data holder holds for the product: the name, the category, the description, `isTailored`, `effectiveFrom` and `effectiveTo` when present, each deposit rate and each lending rate, each fee and each feature, with each of their fields that has a value. When `isTailored` is true, the page says that the rates and the fees can be different for the consumer. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
- The detail of a product that is not current, or that does not exist, is the "not found" page of the site. Judgement: the site shows only current products.
- The site shows the same products and the same values as Get Products and Get Product Detail of the brand at the same time. (See Site API)
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
  - `bsb`: the `bsb` of the brand directory (see Brand directory).
  - `accountNumber`: 8 random digits. When the pair of the BSB and the account number is in use, the data holder makes a new number, with at most 3 attempts in all, as for a customer number. When the 3 attempts fail, the open fails, and the data holder changes nothing.
  - `creditLimit`: the credit limit of the consumer for revolving credit, and `0.00` for a deposit product and a loan.
- The open of a loan also adds one transaction to the account, in the same write as the account: the drawdown. When the write of the transaction fails, the data holder opens no account. The drawdown has the `type` `OTHER`, the `status` `POSTED`, the `description` "Loan drawdown", the `amount` minus the loan amount, an empty `reference`, and the time of the open as its `postingDateTime` and its `executionDateTime`. The account then has the current balance minus the loan amount and the available balance `0.00`, by the balance rule of Customers and accounts. ([CDS: BankingTransactionV2][cds-schema-transaction], [CDS: BankingBalance][cds-schema-balance]: `currentBalance` is "negative if the customer has money owing"; see Decisions)
- The values of the APIs of a consent for a lending account that the consumer opens are the same as for an account that the operator creates with the same credit limit and, for a loan, the same transaction.
- The account has no link to its product. A later change to the product, or its end, does not change the account. (See Decisions)
- The new account is not in an arrangement that exists. The accounts of an arrangement do not change after the confirmation (see Scope: the amendment of a consent). The account is available for the next consent of the customer.
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
- The page shows no accreditation number, no CDR receipt and no amendment. (See Scope and Decisions)
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

The site listener has the rules of the console listener ([Listener and security](#listener-and-security)), and so the rules of the Register console listener: its own listener, off by default, the refusal of cross-origin changes, the `Host` allowlist, no CORS headers, the response headers, the cache headers and the HTML page for a path of the UI. These rules are different:

- The settings of the site have the prefix `BANK_SITE_`: `BANK_SITE_ENABLED` turns the listener on, and `BANK_SITE_ALLOWED_HOSTS` is the `Host` allowlist. The allowlist is required when the listener is on, and has no default. (See Decisions)
- The listener uses TLS, with the certificate and the key of the public listener and the TLS versions and cipher suites of the CDR endpoints. It does not request a client certificate. (See Transport security and Decisions)
- The responses have no `Strict-Transport-Security` header. ([RFC 6797, Appendix B][rfc6797-b]: a browser applies an HSTS policy to each port of the host; see Decisions)
- The site sets one cookie, the session cookie (see Session). The browser gets no token. ([ADR 0014][adr-0014]; see Decisions)
- The site listener serves the brand stylesheet, the logo and the mark of the brand at the paths of the public listener (`/brand/styles.css`, `/brand/logo.svg` and `/brand/mark.svg`), the photos of the brand at `/brand/<file name>` (for example `/brand/hero.webp`), and the fonts of the fixed set, with the `Host` allowlist and the response headers of the site. Judgement: the Content Security Policy permits styles, fonts and images from the same origin only.
- The data holder starts the site listener only in the process that runs the APIs. A process that runs only the workers needs no site setting. (See Non-functional requirements: Least privilege)

### Site API

The site API has the rules of the console API of the data holder ([Console API](#console-api)) for: the `/api/` prefix, the OpenAPI document in the data holder directory and its tests, the problem details and their status codes, an unknown path, no version in the path, a body that is not valid, a body with a field that the OpenAPI document does not give, the field errors with a JSON Pointer and their problem type, and no request body in telemetry. These rules are different:

- The brand gives the `layout` of the brand directory, and the name of each photo that the brand directory has. The site asks only for a photo in the list. Judgement: the site then never shows a broken image (see Presentation).
- The fields of a product have the names and the values of the CDS product schemas. ([CDS: BankingProductV6][cds-schema-product], [CDS: BankingProductDetailV7][cds-schema-product-detail]; judgement: a test can then compare each field with Get Products and Get Product Detail, with no conversion)
- The list of the products gives each current product of the brand in one response, with no pages. Judgement: a catalogue file is 1 MiB or less (see Brand directory), and the site shows all current products on one page.
- The detail of a product gives one current product. A product that is not current, or that does not exist, gives 404, and the `detail` does not contain its identifier.
- The site API and Get Products use the same rule for a current product and the same time of the request. For each preset brand, the list gives the products and the values of Get Products with `effective=CURRENT`, and each detail gives the values of Get Product Detail. A test in the data holder checks this. ([CDS: Get Products][cds-get-products], [CDS: Get Product Detail][cds-get-product-detail]; see Decisions)
- The OpenAPI document gives the values of each CDS enumeration of a product. A test checks that each list is the same as the values that the data holder accepts. (See Decisions)
- No path of the site API contains a customer number or a different identifier of a consumer. Judgement: the session gives the consumer of a request, and a path is in telemetry as a route pattern only.
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

- Telemetry: the site API has the same server telemetry as the other listeners, and the rules of the data holder telemetry apply to it. No span, metric or log contains a query string, a path value, a request body, a cookie value, a session ID or a password. No log of a sign-up, a sign-in, a sign-out or the end of a session contains the customer number or the email address, also for a refused sign-up or sign-in. No log of an open or a close contains the customer number, the product ID, the BSB, the account number, the display name, the credit limit, the loan amount or an arrangement ID. No span, metric or log of the consent list, of a withdrawal on the site or of the record of a disclosure contains the customer number, an arrangement ID, the name of the software product, the brand name or the legal entity name of a data recipient, a data cluster or a time of a disclosure. Judgement: with the session, a product ID tells which product a consumer has, and the name of a data recipient tells with which app a consumer shares data. The UI sends no telemetry. ([ADR 0014][adr-0014], [ADR 0006][adr-0006]; see Non-functional requirements: Telemetry)
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

## Lab-only features

These features do not exist in a production data holder. Real components must not depend on them.

- The data holder sends each OTP email through an SMTP server that its configuration gives, with a sender address from its configuration. In the lab, that server is a mail sink: it keeps each email and delivers none. A lab user and the conformance tests read the OTP from the mail sink. (See Decisions)
- The configuration can give fixed OTPs: a list of email addresses, each with an OTP of 6 digits. Each OTP of a customer with one of these addresses then has the fixed value, and not a random value. The case of the address has no effect. All the other OTP rules stay the same, and the email still goes to the customer. When the configuration gives no list, which is the default, each OTP is random. The data holder does not start when an entry is not valid. Judgement: the browser of the OpenID conformance suite cannot read the mail sink, and it must authenticate with no manual step. The other customers keep a random OTP.
- The operator can load a full set of products into a data holder that has no products: the product catalogue of the brand, or the products of a file. The load is separate from the data holder start, and the data holder start does not need a file.
- The load validates the full set against the product rules in "Product reference data". If one product is not valid, the load changes nothing, and it names each product and rule that fails. The load refuses a product with a field that the data holder does not hold, and a product whose `brand` is not the brand name.
- The file uses the field names of the CDS product schemas. Judgement: the operator then sees the same names in the file and in the API responses.
- The lab loads the product catalogue of the default preset brand (see Product reference data). Judgement: a client can then test each filter of Get Products against the lab with realistic products.
- The online bank site is a lab-only feature (see Online bank site). The lab turns it on for each instance that runs. A production data holder has its own web site.

### The four preset brands in the lab

- With no compose profile, the lab runs only the default preset brand, `wattle-bank`, and CI runs only that brand. Judgement: the lab and CI then need no more memory or time than with one brand.
- With the compose profile `all-banks`, the lab also runs `redgum-bank`, `harbour-bank` and `kestrel-bank` from the same data holder image. Each instance is healthy, and its Get Products response gives the product catalogue of its own preset brand. (See Decisions)
- Each instance has its own database, and a service user that can connect only to that database. ([ADR 0002][adr-0002], [ADR 0020][adr-0020])
- Each instance has its own TLS certificate from the mock public CA, and its own CDR server certificate. The O of the CDR server certificate is the legal entity name of the brand, and its CN and a SAN are the host name of the instance. ([PKI DESIGN.md](../../pki/DESIGN.md); see Transport security)
- Each instance has its own sender address for the OTP email, `no-reply@<name>bank.test`, and the OTP email shows its own brand. (See The OTP email and Decisions)
- On the host, the ports of an instance are a block of six, with the same last digit for each listener as for the default preset brand: 0 health probes, 1 public APIs, 2 operator interface, 3 secure APIs, 4 operator console and 5 online bank site. `wattle-bank` uses 8180 to 8185, `redgum-bank` 8280 to 8285, `harbour-bank` 8380 to 8385 and `kestrel-bank` 8480 to 8485. Judgement: an operator then finds each listener of each instance from one pattern.
- The databases, the users and the certificates of the four instances exist in each lab, also with no profile. A lab that turns the profile on later needs no reset. ([PostgreSQL image][postgres-image-init]: the initialisation scripts run only when the data directory is empty; see Decisions)
- The Register seed of the lab has a record for each of the four preset brands, each with its own legal entity. The legal entity and the brand have the status `ACTIVE`, also with no profile. Each record has the brand name of its preset brand and the base URIs of its instance. (See Brand and Decisions)
- The lab starts with no profile. A service with no profile does not depend on a service of the profile. ([Compose: profiles][compose-profiles])
- CI checks that the compose file is valid with the profile on, and it does not start the profile. The conformance suite runs against the default preset brand only. (See Decisions)
- After `mise run down` or `mise run reset`, no container stays of a lab service that the compose file no longer defines, for example a renamed service. ([Compose: down][compose-down]; see Decisions)

## Deployment from published artifacts

An operator deploys a data holder brand to Kubernetes from the image and the Helm chart that CI publishes. These requirements apply to the artifacts and to CI, not to a deployment.

### Publication

- CI publishes the data holder image from `main` to `ghcr.io/lukemundy/cdr-lab/dataholder`, with the full and the short commit ID as tags, as for the Register image ([README: Container images][readme-conventions]). Judgement: all images of the repository then have one tag pattern.
- CI publishes the image of a commit only after the lint, test, conformance, web and data holder image jobs of that commit pass. A job that fails or that is cancelled stops the publication. A job that CI skips because its inputs did not change does not stop it. ([GitHub Actions: `needs`][gha-needs], [GitHub Actions: status check functions][gha-status-checks]; see Decisions)
- Each commit on `main` that changes an input of the image gets an image. The inputs are the Go code of the data holder and of `pkg/`, `go.mod`, `go.sum`, the web workspace and the `Dockerfile`. (See Decisions)
- The data holder in a published image identifies its commit: in its telemetry, `service.version` is the short commit ID of the image tag ([ADR 0006][adr-0006]).
- CI publishes the chart from `main` when the chart changes, as `oci://ghcr.io/lukemundy/cdr-lab/charts/dataholder` with the version `0.0.0-g<short commit>`, as for the Register chart ([Register chart][register-chart]). CI publishes the chart of a commit only after the tests of the chart of that commit pass. (See Decisions)

### The chart

- One image and one chart serve all brands. Brands differ only in the values of the chart and in the brand directory. Judgement: one artifact for each brand would be a second copy of the same code, and the brand directory already selects the brand ([ADR 0023][adr-0023]).
- No default value of the chart is a value of one deployment. A value that each deployment must give has no default, and the chart does not render when the value is absent. The error names the value. ([Helm: `required`][helm-required]; Judgement: a default that is correct for one deployment is not correct for other deployments, and a deployment that omits the value then starts with a value that is not correct)
- The chart README gives each value and each Secret that a deployment of a brand needs. The chart makes no Secrets.
- The values give exactly one brand: a preset brand by its name, or the name of an existing ConfigMap that has a brand directory. The chart does not render when the values give no brand or two brands. The data holder gets the files of the ConfigMap as one directory. ([Kubernetes: ConfigMap volume][k8s-configmap-volume]; [ADR 0023][adr-0023]; see Decisions)
- Before the data holder starts, the chart applies the migrations as the owner user, and then loads the product catalogue of the brand as the service user when the data holder has no products. The load starts only after the migrations complete. ([Kubernetes: init containers][k8s-init-containers]; [ADR 0020][adr-0020]; see Lab-only features and Decisions)
- The chart runs one pod, and an upgrade stops the old pod before it starts the new pod. Judgement: the migrations take no lock, and the authorisation sessions and the sessions of the online bank site are in the memory of the process (see Non-functional requirements: Restart).
- The chart sets the Register JWKS URL from the base URI of the public APIs of the Register and the CDS path `/cdr-register/v1/jwks`. ([CDS: Register APIs][cds-register-apis]; see Decisions)
- The interval of the Register statuses is a value with no default (see Register statuses and Decisions).
- The chart has no value for the fixed OTPs (see Lab-only features and Decisions).
- The operator console and the online bank site are off by default (see Decisions).
- A deployment of one brand from the published image and chart passes its readiness probe, and Get Products and the OpenID Provider configuration answer on its public listener. The data holder is ready only after it gets the statuses from the Register (see Register statuses).
- The repository has example values for each of the four preset brands, one set for each brand. Each set has only the values of its brand: the preset, the brand ID, the base URLs of the public and the secure APIs, and the sender address of the OTP email. The brand ID and the base URLs agree with the record of the brand in the example seed of the Register ([Register: Deployment from published artifacts][register-deployment]). A deployment adds the other values. (See Decisions)

## Non-functional requirements

These requirements are for the lab. The CDS non-functional requirements for data holders ([CDS: Non-functional Requirements][cds-nfrs]) are context only.

- Performance: on a development workstation, 95% of API calls complete within 200 ms at 50 requests each second. The CDS thresholds are 1000 ms for Get Status and Get Outages, 1500 ms for the product reference data APIs, and 1000 ms to 4000 ms for the APIs of a consent ([CDS: Performance Requirements][cds-performance]).
- Restart: the products, the registrations, the signing keys, the customers, the accounts, the transactions, the arrangements and their ends, the records of disclosures, the notices with their attempts and their outcomes, the subject identifiers, the resource identifiers, the refresh tokens and the revoked access tokens stay the same after a restart. An authorisation flow that is in progress at a restart can end. Each session of the online bank site ends at a restart, and the consumer signs in again. The data holder gets the statuses of software products again after a restart.
- Telemetry: the data holder sends traces, metrics and logs with OpenTelemetry. Telemetry uses route patterns, and not raw URLs. It contains no SSA, no request JWT, no client assertion, no request object, no `request_uri`, no authorisation code, no token, no self-signed JWT and no `cdr_arrangement_jwt`. It contains no customer number, no name, no email address, no OTP, no `sub`, no account data, no transaction data, no `accountId` or `transactionId`, no session ID of the online bank site and no password. It can contain the `x-fapi-interaction-id`. A test in the data holder checks that the telemetry of the first page of the consumer authentication, with a known and an unknown email address and with an email address in a different case, contains no email address.
- Secrets: the repository contains no private keys. Keys and certificates are created when the lab is set up.
- Least privilege: a process that runs only the APIs, or only the workers ([ADR 0010](../../docs/adr/0010-background-jobs-with-river.md)), needs only the settings that it uses. A process that runs only the workers needs no TLS key of the APIs, no operator token and no SMTP settings. A process does not start when a setting that it uses is not present, and its error names the setting. The usage text of the command gives the settings of each subcommand. Judgement: each deployment that has a secret that it does not use is one more location where the secret can be disclosed.
- Portability: a real client changes between the mock data holder and a production data holder with configuration only: base URIs, trust anchors, client certificate, signing keys and identifiers.
- Conformance: a black-box test suite checks each requirement in this document that the CDR defines, through the APIs only ([ADR 0019](../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md)). It finds the base URIs and the brand name of the brand through the Register, as a client does, and it does not depend on the name of a preset brand. For the registration requirements, it gets each SSA from the Register, as a software product does. For the authentication and consent requirements, it acts as the browser of the consumer: it reads each page and sends each form. For the data of a consent, it gets tokens through that flow, and it creates the data with the operator interface. For the end of a consent, it serves the CDR arrangement revocation endpoint of a software product to get the notice, and it changes the statuses of software products with the operator interface of the Register.
- Local run: the data holder runs with local Docker (Compose), with its dependencies.
- Health: the data holder serves a liveness probe (`GET /livez`) and a readiness probe (`GET /readyz`) on a plain HTTP listener that is separate from the CDR API listeners. Judgement: these are the probe paths of the [Kubernetes API server][k8s-health], and all lab services then have the same probes.

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Deprecated versions | Not served | A request for a deprecated version gives 406 on the mock. Data holders can retire version 4 of Get Products and version 6 of Get Product Detail from 10 August 2026 ([CDS: Get Products][cds-get-products], [CDS: Get Product Detail][cds-get-product-detail]). |
| Product base URI | The same as the public base URI | The CDS permits it ([CDS: Participant Endpoints][cds-participant-endpoints]). Judgement: one base URI for all public endpoints is the least complex. |
| Comparison of the `brand` filter | Exact and case sensitive | Judgement: the CDS says only "Filter results based on a specific brand" ([CDS: Get Products][cds-get-products]). A client gets the value from the `brand` field of a product, so an exact comparison is sufficient and the simplest. |
| Query parameter with an empty value | The same as an absent parameter | Judgement: the CDS does not specify it. Each parameter of Get Products is optional, and a client that builds a query from a form can send a name with no value. For `page` and `page-size`, the CDS HTTP library gives the same rule. The mock Register does the same for `updated-since`. |
| `updated-since` comparison | `lastUpdated` after `updated-since` | The CDS says "updated after the specified date and time" ([CDS: Get Products][cds-get-products]). |
| 404 Unavailable Resource on Get Product Detail | Not used | The CDS requires support for the code ([CDS: Get Product Detail][cds-get-product-detail]). A product in the mock is never temporarily unavailable, so the condition cannot occur. |
| Outcome of Get Status and Get Outages | `OK`, and no outages | Judgement: the mock has no monitor of its own endpoints and no planned outages. A data holder that responds is functional. An operator feature to set a status or an outage can be added when a client needs to test it. |
| `updateTime` of Get Status | The time of the request | It is the last time that the status was "updated, modified or verified" ([CDR Support: Get Status Payload Fields][support-status-fields]), and the data holder verifies the status at each request. |
| Sequence of request checks | Version headers first, then the path and query parameters. | Judgement: the CDS does not give a sequence. |
| `additionalValue` of a product component | The load does not check it | The CDS gives a rule for each type value in a table for each component ([CDS: Product & Account Components][cds-product-components]), and not in the schema. Judgement: the load is a lab-only guard for the operator, and a check of each type value does not justify its cost now. The catalogues of the preset brands and of the lab brand obey the tables. |
| Optional fields of a product | Only the effective dates, features, fees, deposit rates and lending rates | The schema marks all the other fields as optional, and the CDS gives a data holder more discretion for payloads with no customer ([CDS: Payload Conventions][cds-payload-conventions]). Judgement: these fields are sufficient for the most common products. Add a field when the lab needs it. The fields that are out are `brandName`, `brandGroup`, `applicationUri`, `additionalInformation`, `cardArt`, `bundles`, `constraints`, `eligibility` and `instalments`, and in the components, fee discounts, rate tiers and rate conditions. |
| `isTailored` that is absent in a product file | `false` | Judgement: the response always has the field, and `false` is correct for a product with standard prices. |
| Load into a data holder that has products | Refused | Judgement: a merge of two sets can leave a `lastUpdated` that is not correct. A reset is simple in the lab. |
| Issuer and hosts | The issuer is the `infosecBaseUri`, on the TLS listener. The mTLS endpoints are on a second listener. | OpenID Connect Discovery requires that the `issuer` is the URL prefix of the configuration ([section 4.3][oidc-discovery-validation]), and the configuration endpoint is a TLS endpoint ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]). The CDS example has an issuer on the mTLS host, which does not obey that rule and is not normative. |
| Paths of the security endpoints | Not specified. A client gets each URL from the configuration. | The CDS gives a path only for the configuration ([CDS: Participant Endpoints][cds-participant-endpoints]). The DCR OpenAPI specification has the server URL `https://mtls.dh.example.com/cds-au/v1`, and the security profile examples use `/register`. A client that reads the configuration works with each. |
| Supported scopes | `openid`, `profile`, `cdr:registration`, and the scopes of the banking and common APIs that are in the plan for this data holder | The CDS requires `openid` and `profile` ([CDS: Scopes and Claims][cds-scopes-and-claims]). Judgement: a scope with no API gives a consent that discloses nothing. Add a scope with the first API that needs it. |
| `response_types` with `code id_token` | Accepted when `code` is also present. Only `code` is registered. | The CDS schema still lists `code id_token`, but data recipients "SHALL only use" `code` from 12 May 2025 ([CDS: Registration Request using JWT][cds-registration-request]). An authorisation server can replace a requested value ([RFC 7591, section 2][rfc7591-metadata]). Judgement: a data recipient that still sends the two values can then register, and the response shows what is registered. |
| Signing key storage | In the data holder database. The data holder creates the key when it has none. | Judgement: the key must stay the same after a restart, and the repository must contain no keys. The mock Register does the same. The lab database is not encrypted. |
| Status of the software product at registration | Not checked | The CDS does not require the check. The Register issues an SSA only for an ACTIVE software product of an ACTIVE data recipient, and the SSA is valid for 10 minutes ([CDS: SSA][cds-ssa]). Register statuses gives the reaction at the other endpoints. |
| Second visit with a `request_uri`, before the consumer completes the flow | Refused with the error page | The CDS requires that the data holder "MUST reject the reuse of request_uri values" ([CDS: Request Object, Data Holders][cds-request-object-dh]). FAPI recommends one-time use at the completion of the flow, and the OpenID conformance suite tests that recommendation. The CDS wins. |
| TLS probes of the OpenID conformance suite | The failures of the conditions `DisallowInsecureCipher` and `CheckForBCP195InsecureFAPICiphers` are accepted | The suite probes TLS 1.2 on the resource endpoint with no client certificate. The mTLS listener ends that handshake with the alert `handshake_failure`, as it must for a client with no certificate ([CDS: Transaction Security][cds-transaction-security]), and the suite reports the closed connection as a failure. A handshake with a cipher suite that the CDS does not permit also gets `handshake_failure`. |
| `sector_identifier_uri` | Registered and returned as the SSA gives it. The data holder does not get the document. It uses only the host of the URL, for the pairwise `sub`. | The DSB article that requires the document at registration is under review. See Open questions. |
| Lifetime and reuse of a registration request JWT | `exp` not more than 10 minutes after the request. A `jti` is accepted one time, across all software products, for 1 hour. | The CDS says only that the `jti` is "used to prevent reuse of the token" ([CDS: Registration Request using JWT][cds-registration-request]), and gives no lifetime. Judgement: a short limit keeps the window for a stolen request small, and a request is not valid for longer than the limit, so a longer memory adds nothing. The mock Register uses the same limits for client assertions. |
| Reuse of an SSA | Permitted until the SSA expires | The SSA definition says that its `jti` is "used to prevent reuse of the SSA" ([CDS: SSA Definition][cds-ssa-definition]), but no rule requires that a data holder refuses a second use. Judgement: a software product can send one SSA in a `POST` and then in a `PUT`. The duplicate rule and the 10 minutes of the SSA limit a replay. |
| Error code of each registration failure | See Registration errors | The CDS gives only the four codes of RFC 7591 and no table of conditions ([CDS: Registration Errors][cds-registration-errors]). Judgement: a fault in the SSA gives `invalid_software_statement`, a redirect URI that the SSA does not have gives `invalid_redirect_uri`, and each other fault of the request gives `invalid_client_metadata`. |
| Error code of a duplicate registration | 400 `invalid_software_statement` | The CDS requires 400 and gives no code ([CDS: Registration Errors][cds-registration-errors]). Its example, which is not normative, uses this code. |
| Body of the 415 response | None | The CDS gives the status code for a media type that is not supported ([CDS: HTTP Response Codes][cds-http-response-codes]) and no body for it. Judgement: no registration error code of RFC 7591 fits, and the status code is sufficient for a client. |
| Limit on new requests for the Register JWKS | One time each second | Judgement: an SSA with an invented `kid` must not make the data holder call the Register for each request. The Register publishes a new key before it signs with it, so a data holder that has the JWKS from after that time has the key. |
| Size of a registration request | 64 KiB at most | Judgement: an SSA and the client metadata are a few kilobytes. A limit protects the data holder from a large body. |
| `unapproved_software_statement` | Not used | Judgement: the data holder approves each valid SSA of the Register, so the condition cannot occur. |
| Errors of registration management | 401 and 403 with no body, and the `WWW-Authenticate` values of RFC 6750 | The DCR OpenAPI specification gives 401 for an "unknown or invalid Client or invalid access token" and 403 when the client "does not have permission", with no body schema ([`cds_dcr.json`][cds-openapi-dcr]). The error values are those of [RFC 6750, section 3.1][rfc6750-errors]. A certificate that does not match gives `invalid_token` ([RFC 8705, section 3][rfc8705-binding]). |
| `DELETE` of a registration | Supported | The CDS marks it as optional ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]). Judgement: a lab user and the conformance tests can then register a software product again with no reset of the lab. |
| CORS on the OpenID Provider configuration and the JWKS | `Access-Control-Allow-Origin: *` | The CDS says that these endpoints do "not require CORS" ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]), which does not forbid the header. Judgement: one behaviour for all TLS endpoints is the least complex. |
| Maximum CRL age in the lab | The lab sets 0: the data holder loads the CRL at each check. | Judgement: a conformance test revokes a certificate and then checks the refusal at once, because the suite does not wait ([conformance DESIGN.md][conformance-design]). The CRL server of the lab is in the lab network, so a load at each check costs little. |
| Health probe paths | `/livez` and `/readyz` | Judgement: see Non-functional requirements. |
| Authentication flow | Redirect to Web with OTP only | A data holder with no app must support this flow ([CDS: Fallback Authentication Flows][cds-fallback-flows]). Redirect to App is for a data holder that has an app, from 10 May 2027. |
| Delivery of the OTP | By email, through an SMTP server from the configuration. In the lab, the server is a mail sink. | The CDS requires "an existing channel or mechanism", at the discretion of the data holder ([CDS: One-Time Password Credential Requirements][cds-otp]), and an OTP does not have to go by SMS ([CDR Support: One Time Password (OTP)][support-otp]). No source gives a method for a test system. Judgement: email is a channel that a mock can have with the behaviour of a production data holder, and a mail sink makes sure that the lab sends no email to a real address. |
| OTP limits | 6 digits, valid for 5 minutes, 3 OTPs and 3 incorrect entries for each flow | The CDS requires 4 to 6 digits and an expiry, and leaves the period to the data holder ([CDS: One-Time Password Credential Requirements][cds-otp]). The DSB says that a limit on failed entries is good practice ([CDR Support: One Time Password (OTP)][support-otp]). Judgement: 5 minutes is sufficient to read an email, and 3 entries of a 6-digit OTP give a guess a chance of 3 in 1,000,000. |
| End of a flow after incorrect OTPs or a cancel | Redirect with `access_denied` | The data holder "should respond to the ADR (client) with an error outcome" for a cancel ([CDR Support: Error scenarios and responses][support-error-scenarios]). The CDS gives no error value. `access_denied` is the value of RFC 6749 for a resource owner that denies the request ([section 4.1.2.1][rfc6749-authorisation-errors]). |
| Time limit of an authorisation flow | 10 minutes | Judgement: the CDS gives no limit. The limit gives time for two OTPs, and it removes flows that a consumer left. After the limit, the data holder cannot be sure that the browser is that of the consumer, so it does not redirect. |
| Levels of assurance | `urn:cds.au:cdr:2` only. A request that needs a different value as essential is refused. | An OTP is one factor ([CDS: Levels of Assurance][cds-loa]). All CDR data APIs are read operations, which need LoA 2. When a data holder cannot reach an essential LoA, "the request should be rejected" ([CDR Support: One Time Password (OTP)][support-otp]). |
| Kinds of customer | Persons only, each with their own accounts | Judgement: one consumer with their own accounts proves the consent flow. Organisations, joint accounts, secondary users and nominated representatives add rules for profiles, approvals and notifications ([CDS: Authorisation Standards][cds-authorisation-standards]), and they do not change the protocol. |
| Fields of a customer, an account and a transaction | The required fields of the CDS schemas, and only the optional fields that this document names | The CDS marks the other fields as optional. Judgement: these fields are sufficient for a data recipient to test each API in scope. Add a field when the lab needs it. The lab has no data to migrate, because a reset is simple. |
| Change or removal of a customer, an account or a transaction | Only the close of an account | Judgement: a closed account lets a data recipient test `openStatus` and its filter. No other requirement needs a change, and a reset of the lab is simple. |
| User identifiers of a consumer | The customer number or the email address, in one field. A value with `@` is an email address, and a different value is a customer number. Spaces at the start and at the end are removed. | The CDS requires an identifier "that can uniquely identify the customer and that is already known by the customer" ([CDS: User Identifiers][cds-user-identifiers]). A consumer knows the email address better than a number of 8 digits that the bank makes. A customer number has only digits, and an email address always has `@`, so one field needs no selection of the type. Judgement: a value that a consumer pastes can have spaces. |
| Uniqueness of the email address | Unique in the brand, with no difference for upper and lower case. The address stays as given. An address in use gives 409 in the operator interface, the console API and the site API. | An identifier must identify one customer ([CDS: User Identifiers][cds-user-identifiers]). Judgement: most mail systems do not make a difference for the case of an address, and a consumer can type it in a different case. 409 is the status for a conflict with the current state of the resource ([RFC 9110, section 15.5.10][rfc9110-409]), as for an account with the same BSB and account number. The data holder of each brand has its own database, so a different brand can have the same address. A refusal of a sign-up then shows that the address is in use, and no rule can prevent this. |
| Identifier of an account | A random UUID that the data holder creates. The operator interface gives it as `accountId`. | The consent page needs a value for each account that the consumer can select, and an arrangement must record the selection. Judgement: with an identifier that has no meaning, a form and an arrangement contain no account number. It is not the `accountId` of the banking APIs, which follows the ID permanence rules (see Resource identifiers). The operator gets the identifier, so that a lab user can find the accounts of an arrangement. |
| Versions of the APIs of a consent | Only the versions of the CDS 1.36.0 OpenAPI specification | Data holders must implement these versions by 9 November 2026, and they can retire the earlier versions from 7 December 2026 ([CDS: Future Dated Obligations][cds-future-dated]). Judgement: the lab has no data recipient that needs the earlier versions, and the earlier versions are only on pages that the CDS calls obsolete. |
| APIs of a consent that are out of scope | 404 `Resource/NotImplemented` | The CDS says that the code "SHOULD be supported for unimplemented APIs" ([CDS: Resource Not Implemented][cds-error-404-not-implemented]). Judgement: payees, direct debits, scheduled payments and instalment plans add data types and no new behaviour. Get Customer Detail needs addresses and phone numbers, and the lab needs no more personal data. |
| `x-fapi-auth-date` that is missing | 400 `Header/Missing` | The CDS text says that the header is "Required for all resource calls" ([CDS: HTTP Headers][cds-http-headers]). The OpenAPI specification marks it as not required, and a DSB article says that its use is "at the discretion of the DH" but that data holders "must validate that the mandatory headers are provided" ([CDR Support: FAPI header obligations for Data Holders][support-fapi-headers]). The CDS text is the first source. |
| Account that is not in the arrangement | The same response as an account that does not exist: `InvalidBankingAccount` | The CDS permits `InvalidBankingAccount` or `UnavailableBankingAccount` ([CDS: Invalid Banking Account][cds-error-404-invalid-banking-account]). For this token, the account is never available. Judgement: one response for the two conditions does not show which accounts exist. |
| Token checks before the other checks | A refused token gives 401 or 403, also when the request has other faults | The DSB says that errors of OAuth, OIDC and FAPI take precedence over CDS error codes ([CDR Support: Error scenarios and responses][support-error-scenarios]). |
| Token for an API that is out of scope | No token check: the response is 404 `Resource/NotImplemented` for each request | An API that the data holder does not implement has no scope to check. Judgement: the response tells a caller with no token only that the API is not implemented, and a token check first adds code with no benefit. |
| Body of the token errors | No body for 401. For 403 on the resource base URI, the CDS error `Authorisation/InvalidConsent`. | RFC 6750 gives the `WWW-Authenticate` values ([section 3.1][rfc6750-errors]). The CDS says that a 403 for a scope that is not correct has an error payload ([CDS: HTTP Response Codes][cds-http-response-codes]), and `InvalidConsent` is for a consent that "is not sufficient" ([CDS: Invalid Consent][cds-error-403-invalid-consent]). The same responses as registration management for 401. |
| Response for a `POST` with a content type that is not JSON | 415 with `GeneralError/Expected` | The CDS gives 415 ([CDS: HTTP Response Codes][cds-http-response-codes]) and no error code for it. `GeneralError/Expected` is the CDS code for an expected error with no more specific code ([CDS: Error Codes][cds-error-codes]). |
| Size of a request body on the resource base URI | 64 KiB at most. A larger body gives 413 with `GeneralError/Expected`. | The CDS gives no limit. Judgement: 1000 `accountId` values are approximately 40 KB, and the other request bodies of the data holder have the same limit. 413 is the HTTP status for this condition ([RFC 9110, section 15.5.14][rfc9110-413]), and the CDS has no error code for it, as for 415. |
| `detail` of a request body that is not JSON | `data` | The CDS `detail` of a body error is the name of the field ([CDS: Invalid Field][cds-error-400-field-invalid]). Judgement: a body that is not JSON has no field, and `data` is the top-level field of each request schema. |
| Fields of a request body that the schema does not give | Ignored | The request schemas do not forbid other properties, and `meta` is an open object ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]). A refusal would refuse a request that the schema accepts. |
| Accounts of Get Balances For Specific Accounts | One balance for each different `accountId`, in the sequence of Get Accounts. An empty list gives an empty list. | The schema permits an empty `accountIds` and gives no rule for a duplicate or for the sequence ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]). Judgement: one record for each account keeps `totalRecords` correct, and one sequence for the three balance APIs gives stable pages. |
| Resource identifiers | For an account and for a transaction, a SHA-256 hash with no key of the `software_id`, the customer number and the identifier of the account or the transaction in the data holder. | The CDS forbids an identifier with meaning and an identifier that is the same for two software products ([CDS: ID Permanence][cds-id-permanence]), and it names a hash for a `transactionId` ([CDS: BankingTransactionV2][cds-schema-transaction]: "through hashing if necessary"). The DSB suggests an identifier for the subject, or for the subject and the client ([CDR Support: ID Permanence and PPID][support-ppid]). Judgement: a stored value for each resource and software product grows with the two, and each consent or read of a page must write it. One method for the two types is one rule to keep correct. The identifier of an account or a transaction in the data holder is a random UUID that no data recipient gets, so a party without it cannot predict or reverse the hash, and a key adds nothing. The two internal identifiers are different random UUIDs, so an account and a transaction do not get the same identifier. A data holder whose internal identifiers a party can guess must use a keyed hash. A new registration keeps the `software_id`, so the identifiers stay the same. The input of the hash must not change. |
| Balances | The data holder calculates them from the transactions and the credit limit | The CDS gives no rule. The balance "should be the same as the balance presented through the mobile application" ([CDS: Data Latency][cds-data-latency]). Judgement: balances that agree with the transactions let a data recipient test both, and the operator does not have to keep two values the same. |
| Change of a transaction | Not supported. To show a posting, the operator creates a posted transaction, and the pending transaction stays. | The CDS does not require a link between a pending transaction and its posted transaction ([CDS: BankingTransactionV2][cds-schema-transaction]). Judgement: the least complex model. |
| Effective time of a pending transaction | The `executionDateTime`, which the operator must give | The CDS uses the posting time, then the execution time, then "A reasonable date/time nominated by the data holder" ([CDS: Get Transactions For Account][cds-get-transactions]). Judgement: a required execution time removes the third case. |
| Transactions with detail | Only the transactions to which the operator gives `extendedData` | The detail object is required, and a transaction with no detail gives 404 ([CDR Support: Error scenarios and responses][support-error-scenarios]). Judgement: the operator can then make each case that a data recipient must handle. |
| `text` filter | Supported | The CDS makes it optional for a data holder ([CDS: Get Transactions For Account][cds-get-transactions]). Judgement: a substring match is simple, and a data recipient can then test the filter. |
| Default of `newest-time` | The time of the request | The CDS says "today" and does not say which time ([CDS: Get Transactions For Account][cds-get-transactions]). Judgement: the time of the request includes each transaction until then, and it needs no time zone. |
| Period of the transactions | 2 years before the request, and no older transactions | Older transactions are not required consumer data ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(6)(a)). Judgement: the data holder shares only required data, and it does not refuse an `oldest-time` that is older, because the CDS only permits the error ([CDR Support: Error scenarios and responses][support-error-scenarios]). |
| Closed accounts | The operator can close an account. An arrangement that has the account keeps it, with `openStatus` `CLOSED` and its transactions. The consent page offers only open accounts. | Data of a closed account is voluntary consumer data for a consent from 4 March 2025 ([CDR Rules][cdr-rules], Schedule 3, clause 3.2(7); [CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). Judgement: a data recipient can then test `openStatus` and its filter. A consumer gives consent for an account so that its data is shared, and eligibility depends on open accounts. When the customer has no open account, the arrangements end (see End of eligibility). |
| Decimal digits of an amount from the operator | Exactly 2 | An AmountString has 2 decimal digits or more, and more "only if required" ([CDS: Common Field Types][cds-common-field-types]). The data holder holds AUD only, which never needs more. Judgement: whole cents make the balance arithmetic exact, with no decimal library. |
| Optional fields of an account and its detail | `openStatus`, `isOwned`, `creationDate`, `bsb` and `accountNumber` only | The CDS marks the other fields as optional. Judgement: the account has no other data. Rates, fees and features of a product are in product reference data. |
| Fields of a person | `lastName`, `firstName`, an empty `middleNames` and `lastUpdateTime` | `lastName` and `middleNames` are required ([CDS: CommonPerson][cds-schema-person]). Judgement: the lab needs no more personal data. Occupation is optional. |
| UserInfo endpoint | Supported, with plain JSON | The CDS says "Data Holders MUST support a UserInfo endpoint" ([CDS: UserInfo endpoint][cds-userinfo-endpoint]). Its list of configuration fields says that `userinfo_endpoint` is necessary "only if OIDC Hybrid Flow is supported" ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]). The first rule is a direct requirement, and the hybrid flow is retired. OpenID Connect permits plain JSON when the registration does not ask for a signed response ([section 5.3.2][oidc-core-userinfo-response]). |
| Traffic thresholds | Not enforced | Judgement: the lab has no such load, and a 429 response would make the conformance tests slow and not reliable. |
| Format of a masked number | `xxx-xxx` for the BSB, a space, and the account number with only its last 4 digits | The CDS requires the number "as the full account number would normally be represented for display (including formatting) but with all digits except the last four replaced with a lower case x", and gives the example `xxx-xxx xxxxx1234` ([CDS: Common Field Types][cds-common-field-types]). |
| Times, and the sharing period on the consent page | All times are UTC in the data holder. The consent page shows the end date in UTC with the label "UTC", and the duration in days and hours. | No source gives a time zone or a format for the page. Judgement: a time zone other than UTC is only for display, and a brand time zone is one more setting that the lab does not need. A duration in days and hours is exact for each value that a software product can ask for, to the hour. |
| Confirmation that breaks an account rule | The consent page again, with the status 400. The flow continues. | The CDS requires that the consumer selects accounts and gives no response for a form that breaks the rule ([CDS: Authorisation Standards][cds-authorisation-standards]). Judgement: a browser sends such a form only when the consumer selects nothing or when the form is changed, and the consumer can correct the first condition. An account of a different customer gets the same response, so the response does not show which identifiers exist. |
| Account number in the operator interface | Not masked | The mask is a rule for data that a consumer or a data recipient sees ([CDS: Common Field Types][cds-common-field-types]). Judgement: the operator sent the number, so a mask hides nothing, and the operator must be able to find an account by its number. |
| Error codes of the PAR endpoint | The codes of the authorisation server library: `invalid_request_object` for a fault of the JWT or of `claims`, `invalid_request` for a fault of an authorisation parameter, `invalid_scope` for a scope that is not permitted, and `invalid_grant` for a client with no `authorization_code` grant type | The CDS gives no codes. RFC 9126 permits each code of RFC 6749, sections 5.2 and 4.1.2.1 ([section 2.3][rfc9126-errors]), and OpenID Connect adds `invalid_request_object` ([OpenID Connect Core, section 3.1.2.6][oidc-core-auth-error]). Judgement: the behaviour of the library is the least complex, and a data recipient reads the `error_description` to find its fault. |
| Lifetime of a `request_uri` | 60 seconds | The CDS permits 10 to 90 seconds ([CDS: Request Object Submission][cds-request-object-submission]). Judgement: the software product redirects the browser at once, and 60 seconds permits a slow browser. The CDS example has 3600, and it is not normative. |
| `response_mode` | Only `jwt` | FAPI 1.0 Advanced requires the value `jwt` with the response type `code` ([section 5.2.2.2][fapi-advanced-jarm]). For that response type, `jwt` gives the response in the query ([JARM, section 2.3.4][jarm-jwt]). The CDS does not name the other JARM modes. Judgement: one mode is the least complex. |
| Refused request at the authorisation endpoint | An error page, and no redirect | RFC 6749 forbids a redirect when the client or the redirect URI is not valid ([section 4.1.2.1][rfc6749-authorisation-errors]). Without a valid `request_uri`, the data holder has no request, and thus no redirect URI that it can trust. |
| `cdr_arrangement_id` in a request object | Refused at the PAR endpoint with `invalid_request_object` | The CDS requires support for the claim, for the amendment of a consent ([CDS: Specifying an existing arrangement][cds-existing-arrangement]). This is a gap: amendment is out of scope now. Judgement: a refusal before the consumer authenticates is better than a new arrangement that the software product did not ask for. |
| Name of the accredited person on the consent page | The legal entity name, with the brand name and the software product name | Rule 4.23(1)(a) requires the name of the accredited person ([CDR Rules][cdr-rules]), and the CX Guidelines say that this is the legal entity name ([CX Guidelines: Authorisation to disclose][cx-authorise]). The SSA definition says that `client_name` and `org_name` are for the consumer "during authorization" ([CDS: SSA Definition][cds-ssa-definition]). Judgement: all three names meet each source. The names come from the registration, because the data holder does not read participant data from the Register now. |
| Period of the data on the consent page | A statement that the data includes data from before the consent | Rule 4.23(1)(b) requires "the period of time to which the CDR data that is the subject of the request relates" ([CDR Rules][cdr-rules]). An authorisation request has no such period: the software product gives a date range only in each data request. Judgement: the statement is correct for each request that the consent permits. |
| `clientName` of an arrangement | The `client_name` of the registration at the confirmation. A later change of the registration does not change it. | Rule 9.3 requires records of each consent ([CDR Rules][cdr-rules]), and the consent page showed this name to the consumer. Judgement: a record that changes with the registration is not a record of the consent, and it also stays after a `DELETE` of the registration. |
| Instructions to withdraw | Text with no link: "Sign in to <brand name> online banking and go to Data sharing, or ask <brand name>." | Rule 4.23(1)(g) requires instructions, and rule 4.24(d) forbids a reference to other documents in the authorisation process ([CDR Rules][cdr-rules]). The CDS friction standard also forbids other documents and services in the flow ([CDS: Common Authentication Standards][cds-cx-authentication]). The text names the two methods of rule 4.25(1). Judgement: a description of the dashboard is the instruction that rule 4.23(1)(g) asks for, not a different document or service, and the site has no link to the consent page (see Separation from the CDR authentication). |
| Time at which an arrangement starts | The confirmation of the consumer | When the software product does not exchange the code, what the data holder does "is at the discretion of the DH" ([CDR Support: Failure to obtain token within Authorisation Code expiry period][support-code-expiry]). Judgement: the arrangement then has no token and no party can use it, so a second step that removes it adds nothing. |
| Lifetime of an authorisation code and of the authorisation response | 60 seconds, and 10 minutes | The CDS gives no lifetime. RFC 6749 recommends 10 minutes at most for a code ([section 4.1.2][rfc6749-code]), and JARM recommends 10 minutes at most for the response ([section 2.1][jarm-response]). Judgement: a software product exchanges the code at once. The 10 minutes of the response is the value of the authorisation server library, which is the least complex. |
| Sector identifier when the registration has no `sector_identifier_uri` | The host of the redirect URIs, or the `software_id` when they have more than one host | OpenID Connect requires a `sector_identifier_uri` for redirect URIs with more than one host ([section 8.1][oidc-core-pairwise]), but the CDS marks the SSA claim as optional ([CDS: SSA Definition][cds-ssa-definition]). Judgement: the `software_id` gives one software product a `sub` that stays the same and that no other software product gets. |
| Response to a data request after the end of an arrangement | 401 with `invalid_token`. The data holder does not use `Authorisation/RevokedConsent`. | The CDS lists 403 `RevokedConsent` for a consent that is "no longer authorised", as a SHOULD ([CDS: Consent Is Revoked][cds-error-403-revoked-consent]). The CDS also requires that a revocation revokes the tokens ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]), and FAPI requires the RFC 6750 errors, which give 401 `invalid_token` for a revoked token ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource], [RFC 6750, section 3.1][rfc6750-errors]). The DSB gives 401 "with a corresponding oAuth error response" ([CDR Support: Error scenarios and responses][support-error-scenarios]). Judgement: the token of an ended arrangement is a revoked token, so the MUST rules apply. |
| Disclosure on a single occasion | The arrangement ends at the end of its sharing, or when the software product revokes it. The data holder does not find when the data "has been disclosed". | Rule 4.26(1)(f) expires the authorisation "after the CDR data has been disclosed" ([CDR Rules][cdr-rules]). A data holder cannot know when a data recipient has all the data that it needs. The CDS requires that the data recipient calls the revocation endpoint "after successful collection" for a sharing duration of 24 hours or less ([CDS: Requesting Sharing Duration][cds-sharing-duration]). |
| Method of the consumer to withdraw | Two methods: the consumer dashboard on the online bank site, and a request to the bank, after which the operator withdraws the consent with the operator interface or the operator console. Only the dashboard shows the messages of the CDS Withdrawal Standards. | Rule 4.25(1) gives the two methods ([CDR Rules][cdr-rules]), and the ACCC says that a data holder "must always have two forms of mechanisms available" ([CDR Support: Withdrawal of authorisation to disclose CDR data and notification][support-withdrawal-notification]). The CDS Withdrawal Standards are for the withdrawal process with the consumer ([CDS: Withdrawal Standards][cds-withdrawal-standards]). In the lab, the operator is the bank, and the operator interface and the operator console are not for the consumer. |
| End of eligibility | When the last open account of a customer closes, all the arrangements of the customer end, and the data holder sends a notice for each one. | Rule 4.26(1)(c) expires the authorisations of a consumer who is not eligible ([CDR Rules][cdr-rules]). The ACCC says that a consumer who closes all accounts is not eligible ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). This end is not one of the exceptions to the notice ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]). |
| Consent of a customer with no open account | The consumer can only cancel, also for a request with no bank scope. | Rule 4.5(1)(c) asks the data holder for an authorisation only for a request "on behalf of an eligible CDR consumer", and rule 4.26(1)(c) expires an authorisation when the consumer is not eligible ([CDR Rules][cdr-rules]). Judgement: the CX Guidelines show unavailable accounts, not a consumer who is not eligible ([CX Guidelines: Authorisation to disclose][cx-authorise]), so the page uses the same pattern as a request with a bank scope and no open account. |
| End of eligibility after a stop | The close of the last open account also records, in the same transaction, the work to end the arrangements. A job does this work 1 minute after the close, and ends each arrangement that did not end with the time of the close. | A stop between the close and the end leaves consents of a consumer who is not eligible, and a second close is refused. Judgement: the job records the work in the transaction of its cause ([ADR 0010](../../docs/adr/0010-background-jobs-with-river.md)). The delay lets the request end the arrangements first. |
| `aud` of the JWTs of a notice | The URL of the endpoint | The CDS says that "The Resource Path for the endpoint being accessed MUST be used" from 31 July 2022 ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt]). A DSB article says that the `aud` is the `RecipientBaseUri` ([CDR Support: Guidance on revocation or withdrawal of consent][support-revocation-guidance]). The CDS is the first source. |
| Lifetime of the JWTs of a notice | 5 minutes | Judgement: the CDS requires `exp` and gives no lifetime ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt]). Each attempt has new JWTs, so a short lifetime is sufficient. The CDS example has 5 minutes. |
| Body of a notice | Only `cdr_arrangement_jwt` | The CDS requires the JWT method, and the duplicate form parameter is a MAY ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]). Judgement: one form of the value is the least complex, and each data recipient must accept it. |
| Outcome of a notice | 204 and 422 complete it. 5xx, no response and no connection give a new attempt. Each other response makes it fail. | The CDS gives 204 for success and 422 for an arrangement that the data recipient does not have or that it revoked ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]). In both cases, the data recipient knows that the arrangement ended. Judgement: a new attempt can help only when the cause is temporary. |
| Attempts of a notice | 3 attempts, 10 seconds and then 60 seconds apart. A notice that is not complete at a stop continues its attempts after the next start, and the total stays at 3. | The DSB asks for "reasonable effort", for example a wait of "a few seconds upon first failure", and not an exhaustive retry ([CDR Support: Data Recipient unavailability back-off pattern][support-backoff]). Judgement: the notice is a civil penalty provision ([CDR Rules][cdr-rules], rule 4.26A), so a restart of the data holder must not lose it. A restart does not give new attempts, so that a data holder that restarts again and again does not send a notice with no limit. |
| Target of a notice | The `recipient_base_uri` of the current registration of the client. A client with no registration gets the value of the registration at the confirmation, which the arrangement keeps. | The CDS names the `recipient_base_uri` of the registration ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]), and a software product changes it with a `PUT` after a change on the Register. Judgement: the current value follows that change, and the value at the confirmation gives a target after a `DELETE` of the registration. |
| Record of a notice | The outcome, with the time and the result of each attempt | Rule 9.3(1) requires a record of each notice ([CDR Rules][cdr-rules]). Judgement: the attempts show what the data holder did when a notice fails. A result has no value of the response, because the response comes from a different participant. |
| Time of the record of an attempt | The data holder records each attempt when it occurs, also when the notice stays pending. The record of the outcome then contains each attempt. | Judgement: the operator and the developer of the data recipient must see the attempts of a notice that is in progress, and a restart must not lose them. A record at the outcome only shows no attempt until the third attempt, which is more than 70 seconds after the first. |
| Statuses that the data holder gets from the Register | The statuses of software products only | The CDS says that the data holder "SHOULD" use the statuses of the data recipients and of the software products ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). In each row of its table, the reactions are the same for each data recipient status with the same software product status, because the Register changes the status of each software product when the status of its data recipient changes ([CDS: Status Mapping][cds-status-mapping]). Judgement: one API gives the same behaviour with one cache. |
| Software product with no status | The responses for INACTIVE. The readiness probe fails until the first statuses. | The CDS says to keep the previous status and gives no rule for a software product with no previous status ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). Judgement: data must not go to a software product whose status is not known, and INACTIVE stops the sharing and changes no arrangement. A registration needs an SSA from the Register, so each registered software product has a status after the next request to the Register. |
| Error for a software product that is not ACTIVE at the PAR, authorisation and token endpoints | `unauthorized_client` | The CDS says only that the data holder does not "Facilitate Consent Authorisation" or disclose data ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). The DSB says that the authorisation endpoint redirects with an error, and gives no code ([CDR Support: Error scenarios and responses][support-error-scenarios]). RFC 6749 gives `unauthorized_client` for a client that is not permitted to do the request ([section 4.1.2.1][rfc6749-authorisation-errors], [section 5.2][rfc6749-errors]). Judgement: `invalid_grant` at the token endpoint would tell the data recipient that the consent ended, but an INACTIVE software product keeps its consents. |
| Sequence of the status check at the PAR, authorisation and token endpoints | After the client authentication at the PAR and token endpoints. The sequence of the other checks is not specified. | The CDS and the ACCC give no sequence for these endpoints. The ACCC sequence is for an access token ([CDR Support: Error scenarios and responses][support-error-scenarios]). Judgement: each sequence refuses the request, and the Register API gives the status to each person, so the sequence discloses nothing. The authorisation server library does its checks of a request before the data holder can do its status check, and a different sequence would need a second client authentication. |
| Time of the clean-up for a REMOVED software product | When the data holder gets the status | The CDS permits a batch, for example each night ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]), and the ACCC also permits it "every time the CDR Register cache is refreshed" ([CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]). Judgement: one step at the time of the status is simpler than a separate batch, and a lab test does not have to wait for it. |
| A check of the `Host` header on the operator interface | An allowlist from a required setting, with no default, as for the Register console. The lab gives the service name for the clients on the compose network, and `127.0.0.1` and `localhost` with the port on the host. | The operator token is public in `compose.yaml`, and the port binding to `127.0.0.1` does not stop DNS rebinding. With DNS rebinding, a web page that the operator opens can send a same-origin request with the token ([Vite advisory GHSA-vg6x-rcgg-rjx6][vite-dns-rebinding]). Judgement: the same check as the Register console, so that one rule protects each plain HTTP interface of the lab. A default would be a value of one deployment, so the setting is required. |
| Method of a withdrawal | A field of the withdrawal event with two values: `CONSUMER_DASHBOARD` and `ALTERNATIVE_METHOD`. No new event type. The use case of the withdrawal gets the method from its adapter. | Rule 9.3(1)(b) requires a record of each withdrawal, and the ACCC expects the record to say how the consumer asked for it ([CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method]). The names come from rule 4.25(1)(a) and (b) ([CDR Rules][cdr-rules]). Each rule of the data holder acts in the same way on the two methods: the end at the time of the request, the revocation of the tokens and the notice. Only the operator sees the method. A field whose values all rules and readers handle in the same way is permitted ([ADR 0021][adr-0021]). |
| Method of a withdrawal before the data holder recorded the method | A withdrawal records its method in a new event. A withdrawal that ended an arrangement before this event existed has the method `ALTERNATIVE_METHOD`. The data holder does not change the events that exist. | Each field of an event type applies to each event of that type ([ADR 0021][adr-0021]), and a method applies only to a withdrawal, not to each end of an arrangement. The events that exist do not change ([ADR 0003][adr-0003]). Each withdrawal before the record of the method was through the operator interface, which is the alternative method (see Withdrawal by the consumer). Judgement: the earlier records give the correct method. |
| Search of the customers in the console | A part of the customer number, the name or the email address | Judgement: a consumer who asks the staff for their records gives one of these values ([CDR Rules][cdr-rules], rule 9.5). The query string then has personal data, so the telemetry rule of the console API is explicit. |
| Account numbers in the console | In full | Judgement: the operator is the staff of the bank, and the operator interface gives the account number in full. The pages for the consumer mask it, because a different person can see the screen of a consumer. |
| Version of a record in the console API | No version. The state of the record refuses an action that is not permitted now. A close also sends the arrangements that its confirmation named. A withdrawal sends nothing. | The customers, the accounts and the transactions have no history and no version (see Operator interface), and a transaction and a customer cannot change. A close and a withdrawal change only a record that is open, and a second close or withdrawal gives 409 ([Operator interface](#operator-interface)). A close of a different account of the customer can make a close the last close, which then ends arrangements that its confirmation did not name. Judgement: the arrangements in the request protect against this change, and a version of the customer would also refuse changes that do not change the effects of the close. A withdrawal ends only the arrangement of its path, so no other change makes its effect different from its confirmation, and a withdrawal of an arrangement that ended gives 409. |
| Field of a rule with two fields | The field that is absent: the `postingDateTime` of a `POSTED` transaction, the `executionDateTime` of a `PENDING` transaction, and `extendedData.payee` or `extendedData.payer` for the sign of the amount | Judgement: the operator corrects the rule with a value in that field. The status and the amount are usually correct, and the error of the field must show at the field that the operator changes ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]). |
| Field of a `bsb` and an `accountNumber` that a different account has | `accountNumber`, with the status 409 of the operator interface | Judgement: the operator corrects the pair with a different account number, because the accounts of a brand usually have the same few BSBs. The Register console API also gives an identifier that a different record has as 409 with a field error ([Register DESIGN.md: Console API][register-console-api]). |
| Location of a new transaction in the console API | `/api/customers/{customerNumber}/accounts/{accountId}/transactions/{transactionId}`, a path with no `GET` | [RFC 9110, section 15.3.2][rfc9110-201]: the `Location` of a 201 identifies the new resource. Judgement: the path of the list does not identify one transaction. The list of the transactions gives each field, so a detail of a transaction adds a second view of the same data, and the UI does not read it. The `transactionId` in the path is the identifier of the operator interface, and no telemetry has a response header. |
| Permitted create of a transaction | An action of the account (`create-transaction`), with the rules of the operator interface: only an open account | The console API gives the actions that the rules permit now (see Rules for the actions), and the UI has no copy of the rules. Judgement: the state of the account permits the create, so the action is with the other actions of the account, as the close is. The form is a link on the transactions of the account, not a confirmation, because a create changes no other record. |
| Group of the extended data in the form of a transaction | A group of fields with a legend that is always open. The console sends `extendedData` only when a field of the group has a value. | Judgement: the design system has no group of fields that opens and closes ([design system DESIGN.md][ui-design]: only a card of a record collapses). A collapsed group also hides its field errors and the target of the links of the error summary. The values of the group are the only sign that the operator wants it, so an empty group sends nothing and a group with one value sends each field, and the data holder refuses each value that is missing. |
| Time of a transaction in the form | A native date-time control in the time zone of the browser, with seconds. The console sends the instant in UTC. | [ADR 0012][adr-0012]: the API gets each time in UTC, and the UI shows each time in the time zone of the browser. Judgement: the operator thinks of a time of a test in local time, and the control of the browser gives a valid local time or no value. The seconds make each time of the CDS available. |
| Problem type of field errors | One type for each console API, in its own OpenAPI document | Judgement: each OpenAPI document defines the extension members of its problem types ([RFC 9457, section 3.2][rfc9457-3-2]), and a type in one service document does not make a service depend on a different service. |
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
| Selection of the brand | Two settings, the name of a preset brand or the path of a brand directory, and exactly one of them. No default. | Judgement: a preset needs no file outside the data holder, so the image runs each preset with one setting ([ADR 0018][adr-0018]), and a user brand needs only a directory. One loader reads both forms. A default would be the brand of one deployment. |
| Source of the brand name | The brand directory only. The data holder has no setting for the brand name. | Judgement: a second source of the same name can disagree with the directory. The Register record of the brand has the same name. |
| Format version of the brand directory | None. The data holder refuses an unknown field and an unknown file. | Judgement: one data holder reads the presets that it contains, so no version is necessary now, and a strict check finds an error in a user directory at start-up. A format version is a new convention. An ADR records the contract. |
| Contrast check of a brand | A test for each preset brand. At start-up, the data holder checks the structure only. | Judgement: the contract requires WCAG 2.2 AA for the pairs of roles ([WCAG 2.2, SC 1.4.3][wcag-1-4-3], [SC 1.4.11][wcag-1-4-11]). A start-up check needs a colour conversion in the data holder only for user brands. |
| Styles of the consumer pages | A stylesheet of the design system that the data holder contains, and the styles of the brand. The pages permit styles, fonts and images of the same origin only. | Judgement: the pages had no styles, and a brand needs the components of the design system. [ADR 0016][adr-0016] changes, because it did not change the data holder pages. The pages stay Go pages with no script ([ADR 0014][adr-0014]). |
| Paths of the styles and the fonts of the consumer pages | `/static/consumer.css` and `/static/fonts/` on the public listener, with no authentication. A browser asks again each time, and an ETag makes the answer small. | Judgement: the files contain no CDR data, and the pages are on the public listener. The file names have no content hash, so a browser must not keep a file after a new version starts. |
| Content Security Policy of the consumer pages | `default-src 'none'; style-src 'self'; font-src 'self'; img-src 'self'; frame-ancestors 'none'; base-uri 'none'`. No `form-action`. | Judgement: the pages need only their stylesheets, fonts and the logo, from the same origin, and no script ([MDN: CSP][mdn-csp]). Some browsers apply `form-action` also to the redirect after a form ([MDN: form-action][mdn-csp-form-action]), and the last form of the flow redirects to the data recipient. |
| Text of the pages when they refuse a request | "<brand name> cannot continue" and "<brand name> cannot use this form" | Judgement: the pages name the bank with its brand name (see Presentation of the consumer pages). |
| Name of the data holder in the page text | The brand name. The conformance suite gets the name from the Register. The data holder gives the footer, not the brand directory. | Judgement: the consumer sees which bank shares the data. A user brand cannot remove the footer. |
| Product fields of the preset catalogues | The fields that the data holder holds now. No eligibility, constraints or rate tiers. | Judgement: five categories are realistic with the fields of today. More fields are a separate change, when a client needs them. |
| Schema check of the preset catalogues | A test of the data holder validates the Get Products response of each preset brand against the pinned CDS OpenAPI specification. | Judgement: the conformance suite runs against the lab brand only, and the other presets must also be valid ([CDS OpenAPI: banking][cds-openapi-banking]). |
| Register record of the default preset brand | The default preset brand has its own legal entity on the Register, with one brand. It replaces the brand "Mock Bank". The other brands of the seed stay with their legal entity. | Judgement: each Australian bank is an authorised deposit-taking institution and a legal entity of its own, and a legal entity with a "Mock" name is not a realistic bank. Each later preset adds its own legal entity and brand. |
| Sender address of the email | A setting of the deployment, with an address only. The brand directory has no email address. | Judgement: the address depends on the mail domain of the deployment. |
| Sender display name of the email | The brand name, from the brand directory. The data holder refuses a sender address setting with a display name. | Judgement: the consumer sees which bank sent the OTP. The brand directory stays the one source of the name (see Source of the brand name), and a display name in the setting can disagree with it. |
| Control characters in the brand name | Refused in `brand.json` | Judgement: the brand name goes into the `From` header, and a line break in a header field can start a new header field ([RFC 5322, section 2.2][rfc5322-2-2]). No real brand name has a control character. A refusal at start-up is simpler than a different rule for each place that shows the name. |
| Notice email to the consumer | None. The OTP email is the only email. | The CDR Rules require a notice to the accredited person when an authorisation ends ([CDR Rules][cdr-rules], rule 4.26A), and an update of the consumer dashboard (rule 4.27). The only notice to a consumer in rules 4.25 to 4.28 is for a secondary user (rule 4.28), who is out of scope. Judgement: the least complex solution. A new email to the consumer can be a later change. |
| Logo of the email | `logo.png`, a required file of the brand directory, in a `cid:` part of the email. No remote image and no conversion of `logo.svg` in the data holder. | PNG has full support in email clients ([caniemail: PNG][caniemail-png]). Some clients do not show SVG ([caniemail: SVG][caniemail-svg]), and the lab mail sink serves an SVG part as `application/octet-stream` ([Mailpit v1.32.0: message.go][mailpit-part-handler]). Judgement: the Go standard library has no SVG rasteriser, a rasteriser is a new dependency, and a remote image works neither in the lab nor before the reader permits images. The cost is one more file for each brand, which ADR 0023 foresaw ([ADR 0023][adr-0023]). The PNG files of the preset brands are made one time from their `logo.svg`, with a tool outside the build. |
| Rules of `logo.png` | A valid PNG, 100 KiB or less, exactly 96 pixels high and 768 pixels wide or less, with the aspect ratio of `logo.svg` within 2 % | Judgement: the email shows the logo 48 CSS px high, and 96 pixels gives a sharp image on a screen with a pixel ratio of 2. The size limit is the limit of the SVG files. A small file can declare an image that needs more memory than the process has, so the data holder checks the size in the header first; 768 pixels (an aspect ratio of 8) is more than each preset logo needs. The aspect ratio check finds a PNG that is not the same logo as the SVG, and the tolerance permits the rounding of the width to a whole pixel. The author keeps the PNG legible on the brand colour, as for `logo.svg`. |
| Format of the email | `multipart/alternative` with a plain text part and a `multipart/related` part with the HTML part and the logo. The HTML part is quoted-printable. Inline styles only, hex colours from the tokens, and a system font stack. | [RFC 2046, section 5.1.4][rfc2046-5-1-4]: the best part is last. [RFC 5322, section 2.1.1][rfc5322-2-1-1]: a line has 998 characters or less, and quoted-printable makes sure of it for a generated HTML part. Many clients do not support a `<style>` element, CSS variables or OKLCH colours ([caniemail: style element][caniemail-style], [caniemail: CSS variables][caniemail-css-variables], [caniemail: modern colour functions][caniemail-modern-colour]). Judgement: most clients do not load a web font, so the logo gives the look of the brand. |
| Password statement in the OTP email | The OTP email says that the data recipient does not need the password of the consumer | The CDS requires the statement in "consumer-facing interactions and communications" ([CDS: Common Authentication Standards][cds-cx-authentication]), and the OTP email is a communication to the consumer. |
| Brand in the console | The shell of the design system with the brand colours: the logo on the brand colour with an accent rule, the side navigation on the hero colour, the page header band and the heading colour. The link and the primary action use the brand. The font, the radius, the other colours and the density stay in the neutral lab brand. The mark is the icon of the tab. | Judgement: the operator asked for consoles that show more of the brand and are less white and flat, but stay dense staff tools. The brand colour is white for three preset brands, so the side navigation uses the hero colour. A different font changes the width of each table column. The label "Operator console" stops an operator from mistaking the console for a consumer page. ([design system DESIGN.md: Brands][ui-brands]) |
| Content of the overview | Counts and the 10 newest changes. The customer number, but no name, email address, display name of an account or account number. | Judgement: a screenshot of the overview then shows no personal information. The customer number is the identifier that the operator needs to find the customer, and it identifies a person only with access to the console. A display name is text that a consumer can choose. |
| Changes in the overview | The creation of a customer, the creation and the close of an account, the confirmation and the end of an arrangement, and the registration, the change and the deletion of a client registration. Not the transactions and not the notices. | Judgement: a lab can have thousands of transactions, which push out all other changes. The transactions of an account and the notice of an arrangement show on their own pages. |
| Counts of the overview | One endpoint for the counts and the changes, from one read | Judgement: two of the counts from the lists need each page of the customers, and the counts of one read agree with each other. |
| Fonts of a brand | One of a fixed set that the data holder contains. A new font is a change to the design system. | Judgement: the fonts are self-hosted and need a licence that permits it ([ADR 0016][adr-0016]). A font file in a user directory needs a licence check and adds size. |
| Layout of the online bank site | A required `layout` in `brand.json`, one of four layouts in the site: `meadow`, `ledger`, `horizon` and `updraft`. The presets: `wattle-bank` `meadow`, `redgum-bank` `ledger`, `harbour-bank` `horizon`, `kestrel-bank` `updraft`. | Judgement: sites with the same layout and different colours look like the same bank. A layout for each preset in the code gives a preset a feature that a user brand cannot have ([ADR 0023][adr-0023]). Required and with no default, as each token is required. |
| Heading font of a brand | A required token `font.heading`, with the rules of the body font. The presets: Bricolage Grotesque (`wattle-bank`), Source Serif 4 (`redgum-bank`), Geist (`harbour-bank`) and Unbounded (`kestrel-bank`). | Judgement: a heading font changes the character of a brand at a small cost. A brand that wants one font gives the body font again. |
| Photos of the brand | Optional files `hero.webp` and `category-<anchor>.webp`, flat in the brand directory | A ConfigMap key cannot contain `/`, so a ConfigMap cannot hold a subdirectory ([Kubernetes: ConfigMaps][k8s-configmap]). Judgement: optional, so that a brand with no photos is valid, and one file name for each category gives each category one photo with no index file. |
| Format and size of the photos | WebP only, not animated. Hero: 256 KiB or less, 1200 to 2400 pixels wide, with the aspect ratio of the layout. Category: 96 KiB or less, 800 to 1200 pixels wide, 4:3. | Each current browser shows WebP, and its header gives the size of the image ([RFC 9649][rfc9649]). Judgement: one format is the simplest check. The widths give a sharp photo on a screen with a pixel ratio of 2 at the size of the layouts, and the limits keep a preset with seven photos in one ConfigMap. |
| Size of a brand directory | 1 MiB or less in all | A brand directory can be one ConfigMap, which is at most 1 MiB ([Kubernetes: ConfigMaps][k8s-configmap]). Judgement: the limits of each file allow more than 1 MiB in all when the directory has photos. |
| Text alternative of the photos | Empty. No alt text field in the brand directory. | Judgement: each photo is decoration, and the text of the page gives its meaning ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]). A field for the text is a change of the contract that gives no information. |
| Accent colour on the site | Shapes only, no text | Judgement: the accent of some presets has a contrast lower than 4.5:1 on white or on the hero colour, and the contrast test of the presets checks only the pairs that the layouts use as text ([WCAG 2.2, SC 1.4.3][wcag-1-4-3]). |
| Names of the brand settings | `DATA_HOLDER_BRAND` (the name of a preset brand) and `DATA_HOLDER_BRAND_DIR` (the path of a brand directory). `api`, `serve` and `seed` need one. `worker` needs neither. | Judgement: the names of the other brand settings start with `DATA_HOLDER_BRAND`. `seed` needs the brand name to check the `brand` of each product, also for a file. ([ADR 0023][adr-0023]) |
| Fields of `brand.json` | `brandName` only | Judgement: the field name of the Register ([CDS: Register API schemas][cds-register-schemas]), as the product file uses the names of the CDS. |
| Tokens of `tokens.json` | Each token of a brand of the design system, with literal values. A missing token is an error. | Judgement: a role that falls back to a different brand gives colours that the contrast test did not check. An alias needs the primitive tokens, which only the web build has. ([ADR 0023][adr-0023]) |
| Extra SVG rules | The `svg` root element is in the SVG namespace, and the `viewBox` of `mark.svg` is square | Judgement: a browser does not show an SVG file with no SVG namespace as an image. DESIGN.md requires a square mark. |
| Entries of a directory that start with `..` | Ignored | Judgement: a Kubernetes ConfigMap volume adds `..data` and a `..<timestamp>` directory ([Kubernetes: atomic_writer.go][k8s-atomic-writer]). Without this rule, a ConfigMap cannot hold a user brand. No file of the contract has such a name. |
| Paths of the brand assets | `/brand/logo.svg`, `/brand/mark.svg` and `/brand/styles.css` on the public listener. The stylesheet sets the custom properties of the design system under `:root:root`. On the console listener: `/brand/logo.svg`, `/brand/mark.svg` and `/brand/console.css`, which sets only the roles of the brand that the console uses. | Judgement: the `logoUri` on the Register needs a fixed URL on the public base URI. The selector has a higher specificity than the `:root` of the design system tokens, so the order of the stylesheets on a page has no effect. The console has the same paths for the same files, and a different name for a stylesheet with different content. |
| Catalogue with no products | Valid. The `products.json` of a brand can be `{"products": []}`. | Judgement: a user brand that tests only the authentication and the consent needs no products, and a brand on the Register with no `productBaseUri` has no product reference data endpoints. |
| Preset brands | `wattle-bank` ("Wattle Bank", the default preset brand), `redgum-bank`, `harbour-bank` and `kestrel-bank`. In the lab, the compose service, the host name, the database and the certificate directory of an instance have the name of its preset. | Judgement: one pattern for the names of each instance, so that a second brand in the lab uses the same pattern. |
| Register records of the preset brands | The legal entity name is the brand name with "Limited", for example "Wattle Bank Limited". The `websiteUri` is `https://<name>bank.test`, and the sender address of the OTP email is `no-reply@<name>bank.test`. | Judgement: an Australian bank is a public company, and its name ends with "Limited". Each name is shorter than the limit of 36 characters of `legalEntityName` ([CDS: Register API schemas][cds-register-schemas]). |
| Product data in the conformance suite | The suite gets the products from Get Products, and checks the order, the filters and the details against them. It has no list of the product IDs of a brand. | Judgement: the suite does not depend on the name of a preset brand (see Non-functional requirements), and a change of a catalogue then needs no change of the suite. |
| Profile of the other preset brands | `all-banks`. With no profile, the lab and CI run only the default preset brand. | Judgement: the lab and CI then need no more memory or time than with one brand. The name says what the profile adds, as the names `openid` and `telemetry` do. |
| Ports of the other preset brands | `redgum-bank` 828x, `harbour-bank` 838x, `kestrel-bank` 848x, with the last digits of `wattle-bank` (818x) | Judgement: ports 8280 to 8484 are not used in the lab, and one pattern for each instance is the simplest to find. |
| Databases, users and certificates with no profile | Created in each lab for all four instances | Judgement: the PostgreSQL image runs its initialisation scripts only when the data directory is empty ([PostgreSQL image][postgres-image-init]), so a database that only the profile creates is not there when a lab turns the profile on later. An empty database and three more certificates cost little. The grants of the application role then also do not depend on a service of the profile ([Compose: profiles][compose-profiles]). |
| Register records of the brands that do not run | Present and `ACTIVE`, in the one seed file, also with no profile | Judgement: the status of a brand on the Register is a fact about its registration, not about whether its instance runs. To a client, a brand that does not run is the same as a bank with an outage, which a data recipient must handle. A seed for each profile needs a second seed file and a second load mode. |
| CI and the profile | No CI job starts the profile. One CI step checks that the compose file is valid with the profile on. The lab deployment proves the four instances after the merge. | Judgement: the conformance suite runs against the default preset brand only, so a CI job that starts the profile proves only that three more containers start, at the cost of three more instances on a runner with one node. |
| Orphan containers | `mise run down` and `mise run reset` remove the containers of services that the compose file no longer defines | Judgement: after the rename of `mock-bank` to `wattle-bank`, the old containers stayed and kept port 8180. Compose removes them only with `--remove-orphans` ([Compose: down][compose-down]). |
| Invented domains | `.test` | `.example` is reserved too ([RFC 2606, section 2][rfc2606-tlds]), but the names must not contain "Example". `.test` is reserved "for use in testing" ([RFC 6761, section 6.2][rfc6761-test]). A realistic domain such as `.com.au` can belong to a real business. |
| Architecture of the online bank site | A React SPA that the data holder embeds and serves, with a site API in an OpenAPI document | [ADR 0014][adr-0014] names a data holder consumer dashboard as a SPA, and the design guide expects the first React consumer UI to add the consumer components to the design system ([design system DESIGN.md: Consumer pages][ui-consumer-pages]). Judgement: the later forms and dialogs of the site need the components that only the React package has, and Go pages put the bank sign-in near the CDR pages. |
| Listener of the online bank site | A lab-only listener of its own on each instance, off by default, with TLS. Port 8x85 on the host: 8185, 8285, 8385 and 8485. | Judgement: on the public listener, the site needs exceptions in the CORS rule and the not-found rule of the CDR APIs, and a lab-only feature goes into the listener of the CDR endpoints. The port is the next number of the block of each instance (see The four preset brands in the lab). |
| Settings of the online bank site | `BANK_SITE_ENABLED` and `BANK_SITE_ALLOWED_HOSTS` | Judgement: the same pattern as `CONSOLE_ENABLED` and `CONSOLE_ALLOWED_HOSTS`. `BANK_SITE` says what the listener serves, and it is different from the CDR names "public" and "secure". |
| TLS of the online bank site | The certificate and the key of the public listener, with the TLS rules of the CDR endpoints | Judgement: a real online bank uses HTTPS, and a session cookie of a later sign-in needs the `Secure` attribute, which the `__Host-` prefix requires ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]). The public certificate of each instance already has the host name `localhost`, so the PKI does not change. One TLS configuration for all TLS listeners is the simplest. |
| HSTS on the online bank site | No `Strict-Transport-Security` header | A browser applies an HSTS policy to each port of the host ([RFC 6797, Appendix B][rfc6797-b]). In the lab, all listeners are on `localhost`, and the consoles and the operator interfaces are plain HTTP, so an HSTS policy from the site stops them in the browser. The site is lab-only. |
| Cookies of the online bank site | One session cookie, `__Host-bank-session-<DATA_HOLDER_BRAND_ID>`, with `Secure`, `HttpOnly`, `SameSite=Strict` and `Path=/`, and no `Domain`, `Max-Age` or `Expires` | The `__Host-` prefix and `SameSite=Strict` are the recommendation for a session ID ([OWASP Session Management Cheat Sheet][owasp-session], [RFC 6265bis, section 4.1.3.2][rfc6265bis-host]). Cookies are not isolated by port ([RFC 6265, section 8.5][rfc6265-8-5]), so in the lab the cookies of each site go to the CDR pages, to the consoles and to the sites of the other banks. Judgement: a name that the CDR pages do not use keeps the two apart, and the brand ID in the name stops a sign-in at one bank from replacing the session of a different bank in the same browser. A UUID has only characters that a cookie name permits. With no expiry, the browser removes the cookie when it closes, and the server gives the lifetime. The cookie is a credential, but it names a session on the server and no script can read it, so the reason of [ADR 0014][adr-0014] for no token in the browser stays true. ADR 0014 gets an amendment for this cookie on the site only. The consoles still get no credential. |
| Source of the products of the site | The same products of the brand as Get Products and Get Product Detail, with the same rule for a current product. A test compares them for each preset brand. | Judgement: a consumer and a data recipient then see the same products, and one source of the products gives it by construction. A call from the site to the CDS API adds a cross-origin call or a call to its own service over the network. |
| Products that the site shows | The current products only. No future product and no product whose `effectiveTo` is in the past. | Get Products gives the current products by default ([CDS: Get Products][cds-get-products]). Judgement: a bank web site lists the products that the bank sells now. |
| Labels of the CDS enumerations | A label in plain language for each value of each enumeration of a product. The build of the site fails when a value has no label. The OpenAPI document of the site gives each enumeration, and a test compares each list with the values that the data holder accepts. | Judgement: a consumer does not read a CDS code. With the two checks, a new value in the data holder or in the CDS cannot show with no label. |
| Text of the landing page | Text of the data holder, the same for each brand, with the brand name. The site API gives the brand name and the text of the footer. No new field in the brand directory. | Judgement: a new field changes the contract of the brand directory ([ADR 0023][adr-0023]), and the site needs no text for each brand now. The footer then has one source, as on the consumer pages. |
| Password statement on the site | In the footer of each page | The CDS requires the statement "in consumer-facing interactions and communications" ([CDS: Common Authentication Standards][cds-cx-authentication]), and the site is a consumer-facing interaction of the data holder. Judgement: the footer is on each page, and a consumer can open each page first. |
| Tests of the online bank site | Tests in the data holder, component tests with axe for each preset brand at 320 CSS px, and one browser test against the lab. Not in the conformance suite. | [ADR 0019][adr-0019]: the conformance suite tests only behaviour that the CDR defines, and a bank web site is not CDR behaviour. |
| Sign-up on the online bank site | The use case of a create of a customer by the operator, with no change. No new event, and no record of the channel of the sign-up. | Customers and accounts are records in tables, not events. The same use case gives the rules of the operator by construction. Rule 9.3 requires records of consents and their withdrawals ([CDR Rules][cdr-rules]), and no rule requires a record of how a person became a customer. [ADR 0021][adr-0021]: a field needs a rule or a reader that uses it, and none uses the channel. |
| Account at sign-up | None. The consumer opens an account on the site, or the operator opens one. | Judgement: the consumer selects the product, as at a real bank. The consequence: a new customer of the site cannot give a CDR consent until the customer has an open account ([CDR Rules][cdr-rules], rule 1.10B(1)(b)). |
| Email address domain at sign-up | Only the text of the page asks for an invented address, with an example on `.test`. No rule of the data holder refuses a different domain. | Judgement: the operator interface has no such rule, and the mail sink of the lab delivers no email. A rule only for the site makes the two flows different. [RFC 6761, section 6.2][rfc6761-test] reserves `.test` for tests. |
| Field errors of the site API | A `detail` in plain language for the consumer, from the site API | Judgement: the `detail` of the operator interface names the fields in camelCase, which a consumer does not read ([design system DESIGN.md: Content][ui-content]). The UI then has no copy of the rules, as in the operator console. |
| Session after a sign-up | A sign-up starts a session | Judgement: one step less for the consumer. The page with the new customer number shows first, so the consumer still sees the number. |
| Password on the sign-in of the site | The page requires a value but does not send it. The site API takes only the user identifier and refuses a `password` field. | The project requires a sign-in in which any password works. Judgement: a person can type a real password from habit. When the UI does not send it, no password goes to the data holder, its database or its telemetry, and a test can prove that the API refuses it. An API that accepts and ignores the field cannot prove this. |
| Unknown user identifier at the sign-in of the site | A field error that says that the bank could not sign the consumer in, the same for an email address and a customer number | Judgement: with any password, each sign-in with the user identifier only shows which email addresses and customer numbers exist. The site is lab-only and its data is invented. The CDR authentication keeps its rule that hides which email addresses and customer numbers exist ([CDS: One-Time Password Credential Requirements][cds-otp]; see Consumer authentication). |
| Session store of the online bank site | On the server, in the memory of the process, with at most 10,000 sessions. A full store refuses a new session with 503. A restart ends each session. | A server-side session ends at a sign-out with no revocation list ([OWASP ASVS 5.0, 7.4.1][asvs-v7]), and [RFC 10017, section 7.1][rfc10017-cookie-session] permits a cookie session for an application on one domain with no OAuth. Judgement: the lab runs one instance of each bank, so a store in memory is the simplest, with the same limit as the authorisation sessions. A full store refuses a new session, so that many sign-ins cannot end the sessions of other consumers. A database store needs a migration and a removal job for no lab need. |
| Session ID of the online bank site | A random value with at least 128 bits of entropy. A new ID at each sign-in and sign-up. | [OWASP ASVS 5.0, 7.2.3 and 7.2.4][asvs-v7]. A new ID at each authentication stops session fixation ([OWASP Session Management Cheat Sheet][owasp-session]). |
| Lifetime of a session of the online bank site | 24 hours from the start. No inactivity timeout. No limit to the number of sessions of a customer. | A time limit for security is a time limit of the content, and a limit that is longer than 20 hours needs no warning or extension ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]). Judgement: the lab has no value to protect, and a short timeout only stops a demonstration. [OWASP ASVS 5.0, 7.3.1][asvs-v7] requires an inactivity timeout, so the site does not meet it. The production gap of the sign-in gives this. |
| Sign-out of the online bank site | The data holder ends the session, and the response expires the cookie. No `Clear-Site-Data` header. | [OWASP ASVS 5.0, 7.4.1][asvs-v7]. `Clear-Site-Data: "cookies"` affects the entire registered domain ([MDN: Clear-Site-Data][mdn-clear-site-data]). In the lab, all UIs are on `localhost`, so it also removes the cookie of a CDR authentication flow in progress. |
| CSRF protection of the online bank site | The refusal of cross-origin changes of the listener. No CSRF token. | Fetch Metadata is the primary protection against CSRF for modern browsers, with the check of `Origin` as the fallback ([OWASP CSRF Prevention Cheat Sheet][owasp-csrf]), and `http.CrossOriginProtection` does both ([Go: CrossOriginProtection][go-crossoriginprotection]). The design system supports modern browsers only ([design system DESIGN.md][ui-design]). The check also applies to the sign-in, so login CSRF needs no session before the sign-in. A different lab port is a different origin, so the check also refuses a change from a different lab UI. |
| Response of the site API with no valid session | 403 | Judgement: a 401 response must have a `WWW-Authenticate` challenge ([RFC 9110, section 15.5.2][rfc9110-401]), and a cookie session has no HTTP authentication scheme. 403 is the response when the server refuses the request ([RFC 9110, section 15.5.4][rfc9110-403]). The UI then goes to the sign-in page. Only the server knows if a session cookie names a session that ended, so a problem type of its own tells the UI, also after a reload, a restart or the end of the 24 hours. A sign-out needs no session: the consumer is signed out after it in each case. |
| Separation of the sign-in of the site from the CDR authentication | A note on the sign-in page of the site, different titles and headings, a different origin, no links in either direction, and no change to the authentication and consent pages | The friction standard forbids a reference to a different service in the authentication flow ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Friction), so the note goes on the site. The authentication and consent pages already say that the bank does not ask for a password on the page. The CDS gives the phrasing of the password statement to the data holder ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Passwords). |
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
| Maximum length of a product name | 100 characters, a product rule | The CDS gives no maximum for the `name` of a product ([CDS: BankingProductV6][cds-schema-product]). Judgement: the account of an open gets the product name, which has 1 to 100 characters. With the rule in the product load, each product can open, and an open cannot fail for its name. The longest name of the presets has 31 characters. |
| Confirmation of an open | The open page, with "Open account" and "Cancel". No second dialog. | An account cannot be removed, so the open cannot be reversed. A page that names the product and the effect before the submission is the "Confirmed" method ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]). |
| Second request of an open | The busy state of the button only. No idempotency key. | Judgement: a second account after two requests is what the consumer asked for, and it has no effect on a different person in the lab. The UI never sends an action again by itself ([design system DESIGN.md][ui-design]). A key adds a store for a case that the busy state already stops. |
| Close on the site | The close use case of the operator console, with the `ends` of the confirmation. The site API gives each open account its `close` action with the arrangements that end. | Judgement: one use case gives the same rules and the same end of eligibility by construction. A consent that starts between the confirmation and the close then gives 409, and does not end with no warning to the consumer. The consumer sees the name of each app, and not a general text. |
| Balance at a close | No rule for the balance or for pending transactions. The dialog shows the balance and says that the bank pays out no balance. | Neither the CDS nor the CDR Rules give a rule for the close of an account. The operator close has no such rule. Judgement: payments are out of scope, so a rule of a zero balance stops the close of an account with transactions on the site. |
| Channel of a close | No record of the channel (site or operator) and no new event | Rule 9.3 requires records of authorisations, their withdrawals and the notices, and no record of how an account closed ([CDR Rules][cdr-rules]). [ADR 0021][adr-0021]: a field needs a rule or a reader that uses it. The end of eligibility has a reason, not a method. |
| Categories of revolving credit and of a loan in the UI | The OpenAPI document of the site API gives the two lists, and a test compares each with the rule of the data holder. The product responses keep the CDS fields only. | Judgement: the product responses then stay the same as Get Products and Get Product Detail with no conversion, and the UI gets the list from one source, as for the labels of the CDS enumerations. |
| Response to an open | 201 with the new account, and no `Location` header | Judgement: the site goes to the page of the account, and it needs the `accountId` from the response. The site API has no `GET` of one account. The sign-up of the site also gives its record in the body with no `Location` header. |
| Record of disclosures | One record in a table for each response with a 2xx status of an API of a consent or of the UserInfo endpoint, with the arrangement, the data cluster and the time. Not an event. | Rule 1.15(3)(f) and rule 7.9(1) require the dashboard to show what CDR data was disclosed and when, and rule 9.3(1)(d) requires records of disclosures ([CDR Rules][cdr-rules]). The ACCC says that the dashboard shows the first and the final date for each data cluster, and that the dates of the consent are not sufficient ([CDR Support: Privacy Safeguard 5 - dates on the consumer dashboards][support-dashboard-dates], [CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters]). Judgement: one record for each response meets the three rules, and the dashboard gets the first and the last date from the records. No rule of the data holder acts on a disclosure, and a disclosure does not change the state of an arrangement, so it is a row, as a transaction is ([ADR 0003][adr-0003]). |
| Data cluster of a record of disclosures | The scope of the API | Judgement: the scope is stable, and for an arrangement it selects one data cluster of the table of Consent, also for `bank:accounts.detail:read` with and without `bank:accounts.basic:read`. The language of a data cluster can change with the CDS. |
| Failure of the record of a disclosure | The data holder keeps the response until it writes the record, and gives 500 with no CDR data when the write fails | Rule 9.3(1)(d) requires a record of each disclosure ([CDR Rules][cdr-rules]). Judgement: a response that left before its record can have no record. A response of an API of a consent is at most one page of the CDS, so to keep it in memory is a small cost. |
| Disclosure with no records | A 2xx response with an empty list is a disclosure | Judgement: the data recipient learns that the data holder has no data of that cluster for the period, and the data holder cannot know which responses the data recipient uses. |
| UserInfo response with no `profile` scope | No record | Judgement: the response gives only the `sub`, which is in no data cluster of the CDS ([CDS: Data Language Standards][cds-data-language]), so the dashboard has no data cluster to show for it. |
| Removal of the records of disclosures | None, as for the other records of the data holder. No entry in `docs/production-gaps.md`. | Rule 9.3(5) requires that the data holder keeps the records for 6 years ([CDR Rules][cdr-rules]). Judgement: a record that stays longer meets the rule, and a production service can operate with it, so the choice does not meet the test of the production gaps file. |
| Names of the data recipient in an arrangement | The `org_name` and the `legal_entity_name` of the registration at the confirmation, in the event that creates the arrangement, with the `client_name` | The DSB recommends the brand name, the software product name and the legal entity name of the data recipient on the dashboard of a data holder ([CDR Support: Dashboards][support-dashboards]). Rule 7.9(1)(c) requires that the dashboard identifies the data recipient ([CDR Rules][cdr-rules]). Judgement: the reason of the decision "`clientName` of an arrangement": the dashboard then shows the names that the consent page showed, also after a change or a `DELETE` of the registration. |
| Accreditation number on the dashboard | Not shown | The DSB only recommends it ([CDR Support: Dashboards][support-dashboards]). The data holder does not read participant data from the Register (see Decisions: Name of the accredited person on the consent page). No data standard or Register entry is specified for rule 1.15(1)(bb) or rule 7.9(1)(c) in the sources of this document. |
| Ended consents on the dashboard | Shown in a group of their own, with the date of the end, and with no reason and no method. No time limit and no pages. | Rule 1.15(3)(e) requires the time of the end of an authorisation that is not current ([CDR Rules][cdr-rules]). The ACCC says that the dashboard need not show the method or the reason ([CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method], [CDR Support: Dashboards][support-dashboards]). Judgement: the accounts page shows closed accounts in the same way, and a customer has few arrangements. |
| Dashboard of a customer with no open account | The same dashboard, with no special case | The ACCC says that the obligation to give a dashboard ends with the eligibility of the consumer ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). No rule forbids a dashboard after that time. Judgement: the least complex solution, and the consumer sees the consents that ended at the close. |
| Text of the CDS messages | The suggested text of the DSB, word for word, with the brand name of the data recipient | The CDS requires the messages, and permits the data holder to use or paraphrase the suggested text ([CDS: Dashboard Standards][cds-dashboard-standards], [CDS: Withdrawal Standards][cds-withdrawal-standards]). Judgement: the text of the DSB needs no review, and a test can check it. The brand name is the name of the data recipient that the consumer knows best. |
| Confirmation of a withdrawal on the dashboard | Two steps: the action "Stop sharing", then a dialog with the CDS messages and the actions "Stop sharing" and "Cancel" | A withdrawal cannot be reversed ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]), and the CDS messages are "part of the withdrawal process" ([CDS: Withdrawal Standards][cds-withdrawal-standards]). Rule 1.15(1)(e): no more complicated than the authorisation, which has a customer number, an OTP and a confirmation ([CDR Rules][cdr-rules]). The CX Guidelines suggest positive friction and the words "Stop sharing" ([CX Guidelines: Withdrawal][cx-withdrawal]). |
| Withdrawal on the site | The use case of a withdrawal by the operator, with the customer of the session and the method `CONSUMER_DASHBOARD`. The `arrangementId` is in the path. | Judgement: one use case gives the same effects as a withdrawal by the operator by construction. A `cdr_arrangement_id` does not identify the consumer ([CDS: CDR Arrangement ID][cds-arrangement-id]), so the rule that no path of the site API contains an identifier of a consumer stays true. |
| Tests of the consumer dashboard | A test in the data holder, in one process, is the proof in CI. The browser test uses an arrangement of the conformance suite and skips when the lab has none. | Judgement: a consent needs the full consent flow of a registered software product with a client certificate, which a browser test cannot do with little code. The lab-web job in CI has no conformance data. The test in one process uses the real token verifier, the event store and the site API, as the test of the end of eligibility through the site does. |
| CDR policy of the data holder | Out of scope | Rule 7.2(8) requires that the CDR policy is readily available through each online service of the data holder ([CDR Rules][cdr-rules]), and so through the online bank site. Judgement: a CDR policy is a legal document of a real entity, and the footer says that the bank is not real. |
| Publication gate of the image | The lint, test, conformance, web and data holder image jobs of the commit. A job that CI skips does not stop the publication. The Register image gets the same gate. | A job with `needs` is skipped when a job that it needs fails, unless its condition has a status check function ([GitHub Actions: `needs`][gha-needs], [GitHub Actions: status check functions][gha-status-checks]). Judgement: CI skips a job only when its inputs did not change, so the result of the previous commit applies. "Tested image" then has one meaning for the two images. One job builds the images of the commit one time, and the conformance job and the image jobs test these same images. |
| Path filter of the image job | All inputs of the image, for pull requests and for `main`. The Register image job gets the same rule. | Judgement: with the inputs of the browser tests only, a commit that changes only Go code outside the UI gets no image. One filter for pull requests and `main` is the simplest. |
| Publication gate of the chart | The tests of the chart only | Judgement: the chart does not refer to an image of its own commit, because the image tag is a required value. A failure of the tests of the service does not make the chart incorrect. |
| Product load in the chart | Always, before each start, with the rule "only when the data holder has no products" | Judgement: each preset brand and each brand directory has a product catalogue, and the load changes nothing when products exist (see Lab-only features). The data holder is a mock, so its chart can run a lab-only feature. The lab does the same. |
| Register JWKS URL in the chart | Derived from the base URI of the public APIs of the Register | The CDS gives the JWKS of the Register at `/cdr-register/v1/jwks` on that base URI ([CDS: Register APIs][cds-register-apis]). Judgement: one value fewer, and the path cannot be different. |
| Interval of the Register statuses in the chart | Required, with no default | Judgement: the data holder has no default, and a default in the chart would be a value that one deployment selected. |
| Fixed OTPs in the chart | No value | Judgement: a fixed OTP is a test feature for the OpenID conformance suite in the lab (see Lab-only features), and a deployment on a cluster does not run that suite. An operator cannot set by mistake a value that the chart does not have. |
| Operator console and online bank site in the chart | Off by default, each with a value that turns it on | Judgement: the same as the Register chart for its console. Both are lab-only features, and a deployment turns on only the features that it uses. |
| Brand directory in the chart | The name of an existing ConfigMap, given to the data holder as a directory, and not file by file | Helm reads only the files in the chart, so a published chart cannot make a ConfigMap from a brand directory of the repository ([Helm: chart files][helm-files]). A ConfigMap volume mount with `subPath` gets no updates ([Kubernetes: ConfigMap volume][k8s-configmap-volume]). Judgement: a brand directory in base64 in the values is difficult to review. The operator makes the ConfigMap with `kubectl create configmap --from-file` from the directory. |
| Example values of the preset brands | One set for each brand, with only the values of the brand, and with the brand ID and the base URLs of its record in the example seed of the Register | Judgement: the Register gives data recipients the endpoints of the record of a brand. A release with different base URLs becomes ready, because an empty list of statuses is a success, but data recipients cannot find it. The other values are values of the deployment, and no value of one deployment is a default (see The chart). One set for each brand gives each release one file. |

## Open questions

| Question | Detail |
|---|---|
| `brandGroup` of the products | Decision 376 added the optional `brandGroup` field for brand owners that have arrangements with more than one data holder ([CDS: BankingProductV6][cds-schema-product]). The lab has one data holder, so the example products do not have the field. Does the load check `brandGroup` against the brand group of the brand on the Register? |
| `sector_identifier_uri` at registration | DSB guidance says that a data holder gets the document at `POST` and `PUT`, and refuses the request when it cannot ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]). OpenID Connect requires that the document contains each registered redirect URI ([OpenID Connect Core, section 8.1][oidc-core-pairwise]). An older article with the same procedure is under review. The data holder does not get the document now. |
| Pairwise `sub` for two software products with one sector identifier | The CDS rule gives two software products with the same `sector_identifier_uri` host the same `sub` for one customer ([CDS: Sub Claim][cds-sub-claim]), and the DSB has an open issue about it ([standards-maintenance issue 480][dsb-issue-480]). The data holder follows the CDS rule. |
| Permitted authenticator types | The CDS permits only the authenticator types of the Digital ID accreditation standards ([CDS: Baseline Credential Requirements][cds-baseline-credentials]). This document does not use that instrument as a source. An OTP by email is the flow that the CDS names, but its authenticator type is not confirmed against that instrument. |
| Closed accounts on a new consent | The ACCC says that a data holder that does not share a closed account "should show these accounts in the account selection step, but they must not be selectable", with the reason ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). The consent page offers only open accounts and does not show closed accounts (see Decisions: Closed accounts). The close on the online bank site makes closed accounts usual in the lab. |
| Arrangements of a deleted registration | The CDS gives no rule for the arrangements of a registration that the software product deletes with `DELETE` ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]). Their tokens are then not accepted, but the arrangements do not end, and the operator can withdraw them. The notice then goes to the `recipient_base_uri` of the registration at the confirmation (see Decisions). |

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
[cds-request-object]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-object
[cds-request-object-dh]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#security-profile_request-object_data-holders
[cds-request-object-submission]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-object-submission
[cds-existing-arrangement]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#specifying-an-existing-arrangement
[cds-baseline-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#baseline-security-provisions
[cds-code-flow]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorization-code-flow
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
[cds-arrangement-id]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-id
[cds-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#claims
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
[support-par-redirect-uri]: https://cdr-support.zendesk.com/hc/en-us/articles/900004240446
[support-concurrent-consent]: https://cdr-support.zendesk.com/hc/en-us/articles/4407560834959
[support-code-expiry]: https://cdr-support.zendesk.com/hc/en-us/articles/360004665496
[support-ppid]: https://cdr-support.zendesk.com/hc/en-us/articles/5505351487631
[dsb-issue-480]: https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/480
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cx-otp]: https://cx.dsb.gov.au/authenticate/redirect-with-one-time-password
[cx-authorise]: https://cx.dsb.gov.au/authorise/authorisation-to-disclose
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[fapi-advanced-server]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#authorization-server
[fapi-advanced-jarm]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#jarm
[fapi-baseline-server]: https://openid.net/specs/openid-financial-api-part-1-1_0.html#authorization-server
[fapi-baseline-nonce]: https://openid.net/specs/openid-financial-api-part-1-1_0.html#client-requesting-openid-scope
[jarm-response]: https://openid.net/specs/oauth-v2-jarm.html#section-2.1
[jarm-query]: https://openid.net/specs/oauth-v2-jarm.html#section-2.3.1
[jarm-jwt]: https://openid.net/specs/oauth-v2-jarm.html#section-2.3.4
[oidc-core-auth-error]: https://openid.net/specs/openid-connect-core-1_0.html#AuthError
[oidc-discovery-validation]: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationValidation
[oidc-core-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[rfc6749-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
[rfc6749-code]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2
[rfc6749-authorisation-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2.1
[rfc6750]: https://www.rfc-editor.org/rfc/rfc6750
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc5321-mailbox]: https://www.rfc-editor.org/rfc/rfc5321#section-4.1.2
[rfc7636-challenge]: https://www.rfc-editor.org/rfc/rfc7636#section-4.2
[rfc9126-request]: https://www.rfc-editor.org/rfc/rfc9126#section-2.1
[rfc9126-response]: https://www.rfc-editor.org/rfc/rfc9126#section-2.2
[rfc9126-errors]: https://www.rfc-editor.org/rfc/rfc9126#section-2.3
[rfc9126-request-object]: https://www.rfc-editor.org/rfc/rfc9126#section-3
[rfc9126-authorisation]: https://www.rfc-editor.org/rfc/rfc9126#section-4
[rfc9126-guessing]: https://www.rfc-editor.org/rfc/rfc9126#section-7.1
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-host]: https://www.rfc-editor.org/rfc/rfc9110#section-7.2
[rfc9110-413]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.14
[rfc9110-head]: https://www.rfc-editor.org/rfc/rfc9110#section-9.3.2
[rfc7591-metadata]: https://www.rfc-editor.org/rfc/rfc7591#section-2
[rfc7591-software-statement]: https://www.rfc-editor.org/rfc/rfc7591#section-2.3
[rfc7591-response]: https://www.rfc-editor.org/rfc/rfc7591#section-3.2.1
[rfc7591-errors]: https://www.rfc-editor.org/rfc/rfc7591#section-3.2.2
[rfc8705-binding]: https://www.rfc-editor.org/rfc/rfc8705#section-3
[k8s-health]: https://kubernetes.io/docs/reference/using-api/health-checks/
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
[vite-dns-rebinding]: https://github.com/vitejs/vite/security/advisories/GHSA-vg6x-rcgg-rjx6
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
[rfc9457-3-2]: https://www.rfc-editor.org/rfc/rfc9457#section-3.2
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
