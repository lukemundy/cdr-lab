# Mock CDR Register: design

This document gives the specification and the requirements for the mock CDR Register. It does not describe the implementation.

## Purpose

The mock Register implements the CDR Register as the CDS specifies it. Data holders and data recipients in this lab use it the same way they use the production Register that the ACCC operates. A "real" component must be able to change from the mock Register to the production Register with a configuration change only.

## Sources

Target: Consumer Data Standards (CDS) 1.36.0. CDS links go to the 1.36.0 archive URL. While 1.36.0 is the current version, that URL redirects to the live site.

The requirements follow the process in [README: Sources for requirements][readme-sources]. Each requirement links to its source. "Judgement" marks a requirement that the CDS and the ACCC do not state, and gives the reason.

Background research, with the full detail and more sources, is in [`docs/research/`](../../docs/research/README.md).

## Scope

In scope:

- The public (TLS) Register APIs.
- The secure (mTLS) Register APIs.
- The Register authorisation server: OpenID Provider configuration, token endpoint, access token signing keys.
- The Register JWKS and Software Statement Assertion (SSA) signing.
- Participant data and status rules.
- The operator interface, and the lab-only features to load participant data.
- The mock CDR intermediate CA in a running lab: the issue of CDR client and server certificates from CSRs, the CDR CA bundle, the revocation of CDR certificates and the CRL ([ADR 0022][adr-0022]).
- The operator console: a web UI for the lab operator.
- The load of a participant seed file by the Helm chart of the Register, and an example seed file for the four preset banks of the data holder (see Deployment from published artifacts). The deployment itself is out of scope.

Out of scope for now:

- Register calls to data holders: Get Metrics and Metadata Update. The ACCC has not implemented the Metadata Update trigger ([standards-maintenance #537][dsb-537]). Add Get Metrics when a mock data holder exists.
- Deprecated endpoint versions. The mock serves only the current version of each endpoint (see Decisions).
- The Participant Portal, accreditation workflow and CDR Rules processes.
- Automatic archiving of dormant entries ([CDS: Archiving][cds-archiving]).
- The set-up of the mock CDR PKI: the CAs, and the certificates of the lab services and tests. The PKI has its own design ([PKI DESIGN.md][pki-design]), which also gives the profiles of the certificates that the Register issues. "Transport security" below gives what the Register needs from the PKI.

## Actors

| Actor | Uses |
|---|---|
| Data recipient software product | Token endpoint, Get Data Holder Brands, Get SSA, Get OpenId Provider Config, Get Data Holder Statuses, Get Data Holder Brands Summary |
| Data holder brand | Get JWKS, Get Data Recipients, Get Data Recipients Statuses, Get Software Products Statuses, Get Data Holder Brands Summary |
| Public client | Get Data Holder Brands Summary |
| Lab operator | Operator interface, operator console, lab-only features |

Source: [CDS: Register APIs][cds-register-apis].

## Participant model

| Entity | Parent | Status values |
|---|---|---|
| Data holder legal entity | none | ACTIVE, REMOVED |
| Data holder brand | Data holder legal entity | ACTIVE, INACTIVE, REMOVED |
| Data recipient legal entity | none | ACTIVE, SUSPENDED, REVOKED, SURRENDERED |
| Data recipient brand | Data recipient legal entity | ACTIVE, INACTIVE, REMOVED |
| Software product | Data recipient brand | ACTIVE, INACTIVE, REMOVED |

Source: [CDS: Register API schemas][cds-register-schemas], [CDS: CDR Federation][cds-federation].

- The name of a data holder legal entity has 36 characters or fewer. The `LegalEntityDetail` schema of Get Data Holder Brands gives this limit, which is shorter than the limit of the other name fields. The load refuses a longer name. ([CDS: Register API schemas][cds-register-schemas])
- A data holder brand has one or more industries: `banking`, `energy`, `non-bank-lending`, `telco`. ([CDS: Get Data Holder Brands][cds-get-dh-brands])
- A data holder brand has its endpoints (public, product, resource, infosec, admin and extension base URIs, website URI), its JWKS endpoint and an optional brand group. ([CDS: Participant Endpoints][cds-participant-endpoints], [CDS: Get Data Holder Brands][cds-get-dh-brands])
- A data holder brand has a `dataHolderBrandId`, an `interimId` or both. A brand that shares only product data has no `dataHolderBrandId`. ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary])
- A brand with no `dataHolderBrandId` needs only its public details: name, logo, industries, status and public base URI. Judgement: a brand that shares only product data has no secure endpoints, and Get Data Holder Brands, which shows the secure endpoints, requires `dataHolderBrandId` ([CDS: Register API schemas][cds-register-schemas]).
- A data recipient legal entity has an accreditation number and an accreditation level (`UNRESTRICTED` or `SPONSORED`). ([CDS: Get Data Recipients][cds-get-drs])
- A software product has the metadata that the SSA contains: name, description, client URI, logo URI, redirect URIs, sector identifier URI, terms of service URI, policy URI, JWKS URI, revocation URI, recipient base URI and scopes. ([CDS: SSA Definition][cds-ssa-definition])
- A software product has a list of the client certificates that the Register trusts for it. ([ACCC Developer Portal: Token Endpoint][accc-token])
- Text fields that the CDS Register schemas give a `maxLength` have that limit in characters. The load refuses a longer value. ([CDS: Register API schemas][cds-register-schemas])
- Each URI field of a participant is an absolute URI with the `https` scheme and a host, of 1000 characters or fewer. This applies to each logo URI, each endpoint and the JWKS endpoint of a data holder brand, and each URI of a software product, also each redirect URI. An optional URI field with no value is valid. The load refuses a value that is not valid. The CDS gives these fields the type `URIString`, "A valid URI" ([CDS: Common field types][cds-common-field-types], [CDS: Register API schemas][cds-register-schemas], [CDS: RegistrationProperties schema][cds-dcr-schemas]), and a URI has only the characters of [RFC 3986, section 2][rfc3986-2]. The Register schemas give a URI `maxLength: 1000`. (See Decisions for `https` and for the length.)
- All identifiers are strings of 36 characters or fewer. The Register issues identifiers as GUIDs. ([CDS: Register API schemas][cds-register-schemas]; GUIDs are judgement: they are unique without coordination and fit the 36-character limit)
- Each identifier is unique across all participant records, not only across the records of one kind. Judgement: the CDS does not say. GUIDs are unique in all cases, so the rule costs nothing, and it prevents a mix-up between records.

## Status rules

- Data recipient legal entity transitions are only these ([CDS: Participant Statuses][cds-participant-statuses], state diagram):
  - ACTIVE to SUSPENDED, REVOKED or SURRENDERED
  - SUSPENDED to ACTIVE, REVOKED or SURRENDERED
  - REVOKED to ACTIVE
  - SURRENDERED is final.
- Software product transitions are only these ([CDS: Participant Statuses][cds-participant-statuses], state diagram):
  - ACTIVE to INACTIVE or REMOVED
  - INACTIVE to ACTIVE or REMOVED
  - REMOVED is final.
- When a data recipient legal entity changes status, its software products change at the same time. SUSPENDED gives INACTIVE. REVOKED or SURRENDERED gives REMOVED. ([CDS: Status Mapping][cds-status-mapping])
- When a data recipient legal entity is SUSPENDED, REVOKED or SURRENDERED, none of its software products can be ACTIVE. ([CDS: Status Mapping][cds-status-mapping])
- A software product can change status without a change to its data recipient. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- Data holder legal entity transitions are only ACTIVE to REMOVED. REMOVED is final. Judgement: the CDS gives no state diagram for data holders ([CDS: Participant Statuses][cds-participant-statuses]). REMOVED is final in the software product diagram, and the same value has the same meaning for each record. (See Decisions.)
- Data holder brand and data recipient brand transitions are the same as the software product transitions. Judgement: the CDS gives no state diagram for brands, and a brand has the same three status values as a software product. (See Decisions.)
- When a data holder legal entity becomes REMOVED, each of its brands that is not REMOVED becomes REMOVED at the same time.
- When a data recipient brand becomes INACTIVE, each of its ACTIVE software products becomes INACTIVE at the same time. When it becomes REMOVED, each of its software products that is not REMOVED becomes REMOVED. When it becomes ACTIVE, its software products do not change. Judgement: this is the cascade of a data recipient legal entity ([CDS: Status Mapping][cds-status-mapping]), applied to the brand, and it keeps the rule below.
- A record cannot be more active than its parent. Judgement: the CDS gives no rule for brands, and a brand that is not active cannot share data. Thus:
  - A software product of an INACTIVE data recipient brand is INACTIVE or REMOVED.
  - A software product of a REMOVED data recipient brand is REMOVED.
  - A brand of a REMOVED data holder legal entity is REMOVED.
- The Register never issues a REMOVED software product identifier again. ([CDR Support: Removed software product][support-removed-sp]) The Register deletes no record: a removal is a change to the status REMOVED, and the identifier stays in use. Judgement: the same rule for each identifier is simpler, and the events of a record stay ([ADR 0003][adr-0003]).
- The Register never returns a status value that is not in the enumerations above. ([CDS: Participant Statuses][cds-participant-statuses])
- A status change is visible in all API responses immediately after the change. A suspension, revocation or surrender takes effect when it is on the Register ([CDR Rules, rule 5.22][cdr-rules]), and data holders must react within 5 minutes ([CDS: Metadata Cache Management][cds-metadata-cache]).

## APIs

### Endpoints

| API | Method and path | Version | Transport | Auth |
|---|---|---|---|---|
| [Get OpenId Provider Config][cds-get-oidc] | `GET /idp/.well-known/openid-configuration` | none | TLS | none |
| Access token signing keys | `jwks_uri` in the OpenID Provider configuration | none | TLS | none |
| [Get JWKS][cds-get-jwks] (SSA and outbound signing keys) | `GET /cdr-register/v1/jwks` | none | TLS | none |
| Token endpoint | `POST /idp/connect/token` | none | mTLS | `private_key_jwt` |
| [Get Data Holder Brands][cds-get-dh-brands] | `GET /cdr-register/v1/{industry}/data-holders/brands` | 3 | mTLS | Bearer, `cdr-register:read` |
| [Get Software Statement Assertion][cds-get-ssa] | `GET /cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa` | 4 | mTLS | Bearer, `cdr-register:read` |
| [Get Data Holder Brands Summary][cds-get-dh-brands-summary] | `GET /cdr-register/v1/{industry}/data-holders/brands/summary` | 2 | TLS | none |
| [Get Data Holder Statuses][cds-get-dh-statuses] | `GET /cdr-register/v1/{industry}/data-holders/status` | 2 | TLS | none |
| [Get Software Products Statuses][cds-get-sp-statuses] | `GET /cdr-register/v1/{industry}/data-recipients/brands/software-products/status` | 3 | TLS | none |
| [Get Data Recipients Statuses][cds-get-dr-statuses] | `GET /cdr-register/v1/{industry}/data-recipients/status` | 3 | TLS | none |
| [Get Data Recipients][cds-get-drs] | `GET /cdr-register/v1/{industry}/data-recipients` | 4 | TLS | none |

The token endpoint path comes from the example OpenID Provider configuration in [CDS: Register Endpoints][cds-register-endpoints].

- Request and response bodies obey the CDS 1.36.0 Register OpenAPI specification ([`cds_register.json` at v1.36.0][cds-openapi]), except where a requirement below says otherwise.
- The TLS APIs and the mTLS APIs have different base URLs, as `api.cdr.gov.au` and `secure.api.cdr.gov.au` do in production. ([CDS: Register APIs][cds-register-apis])
- `{industry}` accepts `banking`, `energy`, `non-bank-lending`, `telco` and `all` on data holder APIs. It accepts only `all` on data recipient APIs and the SSA API. ([CDS: Register API schemas][cds-register-schemas])
- On Get Data Holder Statuses, an `{industry}` other than `all` gives only the data holder legal entities that have a brand in that industry. The status of the brand does not matter. `all` gives all data holder legal entities. The CDS says only that the parameter is "the industry the participant is retrieving data for" ([CDS: Get Data Holder Statuses][cds-get-dh-statuses]). The ACCC mock Register filters by industry ([RegisterStatusRepository.cs][accc-mock-dh-statuses]). See Decisions for the brand status.

### Versioning and headers

- Versioned APIs require `x-v` and accept `x-min-v`. The response has the highest supported version between `x-min-v` and `x-v`. If `x-min-v` is equal to or higher than `x-v`, the Register ignores `x-min-v`. ([CDS: HTTP Headers][cds-http-headers])
- If no requested version is supported, the response is 406 with `urn:au-cds:error:cds-all:Header/UnsupportedVersion`. ([CDS: Error Codes][cds-error-406-unsupported-version])
- If `x-v` or `x-min-v` is not a positive integer, the response is 400 with `urn:au-cds:error:cds-all:Header/InvalidVersion`. ([CDS: Error Codes][cds-error-400-invalid-version])
- If `x-v` is missing, the response is 400 with `urn:au-cds:error:cds-all:Header/Missing`. ([CDS: Error Codes][cds-error-400-header-missing])
- Each versioned response has an `x-v` header with the version served. ([CDS: HTTP Headers][cds-http-headers])
- Responses from authenticated APIs have an `x-fapi-interaction-id` header. Its value is the request `x-fapi-interaction-id`, or a new RFC 4122 UUID when the request has none. ([CDS: HTTP Headers][cds-http-headers]) Public APIs do the same. Judgement: one behaviour for all APIs is simpler, and it helps tracing across participants.

### Errors

- Error responses use the CDS error envelope `{"errors": [{"code", "title", "detail"}]}`. The `title` is the title that the CDS error list gives for the code. The Register omits the optional `meta` object, because it uses only standard CDS error codes. ([CDS: Error Codes][cds-error-codes])
- An invalid `{industry}` value on Get Data Holder Brands Summary gives 404 with `urn:au-cds:error:cds-register:Field/InvalidIndustry`. The CDS error list and the OpenAPI specification agree on this. ([CDS: Invalid Industry][cds-error-invalid-industry], [CDS: Get Data Holder Brands Summary][cds-get-dh-brands-summary])
- An invalid `{industry}` value on all other endpoints gives 400 with `urn:au-cds:error:cds-all:Field/Invalid`. This includes a valid industry that the endpoint does not allow, for example `banking` on a data recipient endpoint. The CDS error list says 404 `Field/InvalidIndustry`, and the OpenAPI specification says 400 "Invalid Path Parameter". The ACCC clarifies that current versions return 400 `Field/Invalid`. ([CDS: Invalid Field][cds-error-400-field-invalid], [ACCC Developer Portal: Known Issues][accc-known-issues])
- An invalid `page-size` gives 400 with `urn:au-cds:error:cds-all:Field/InvalidPageSize`. A page out of range gives 422 with `urn:au-cds:error:cds-all:Field/InvalidPage`, and the `detail` is the number of pages. Page 1 of an empty list is in range. ([CDS: Error Codes][cds-error-codes], [CDS: Pagination][cds-pagination])
- A `page` that is not a positive integer gives 400 with `urn:au-cds:error:cds-all:Field/Invalid`. An `updated-since` that is not an RFC 3339 date-time gives 400 with `urn:au-cds:error:cds-all:Field/InvalidDateTime`. An empty `updated-since` is ignored. ([CDS: Error Codes][cds-error-codes])
- An invalid or missing bearer token gives 401 with a `WWW-Authenticate` header and no body. A request with no bearer token gets `Bearer`. A token that is not valid, or that is bound to a different client certificate, gets `Bearer error="invalid_token"`. A valid token without the scope `cdr-register:read` gets 403 with `Bearer error="insufficient_scope"`. ([CDS: Get Data Holder Brands][cds-get-dh-brands], [RFC 6750, section 3.1][rfc6750-errors])
- An unknown path gives 404 with `urn:au-cds:error:cds-all:Resource/NotFound`. ([CDS: Error Codes][cds-error-codes])
- An unexpected server error gives 500 with `urn:au-cds:error:cds-all:GeneralError/Unexpected`. The `detail` does not show the cause. ([CDS: Unexpected Error][cds-error-5xx-unexpected])
- Token endpoint errors use the RFC 6749 section 5.2 format, with only the error codes of that section. A client that the Register cannot authenticate or trust gets 401 `invalid_client`. Judgement: the CDS does not define Register token errors, and RFC 6749 is the base protocol. ([RFC 6749, section 5.2][rfc6749-errors])

### Caching, pagination, lists and CORS

- Get JWKS, Get Data Holder Brands Summary and the four status and recipient APIs return an `ETag` header. When `If-None-Match` matches the current `ETag`, the response is 304 with no body. ([CDS: Metadata Cache Management][cds-metadata-cache])
- `ETag` and `If-None-Match` follow RFC 9110: the entity tag is in quotes, and `If-None-Match` uses the weak comparison. An `ETag` changes only when the response data changes. Judgement: the CDS does not give the format, so the HTTP standard applies. ([RFC 9110: ETag][rfc9110-etag])
- Responses from these APIs and from Get OpenId Provider Config have `Cache-Control: no-cache`, so that clients check the `ETag` on each request. Judgement: the CDS does not give cache lifetimes for responses. A status change must reach data holders quickly ([CDS: Metadata Cache Management][cds-metadata-cache]), and a revalidation with `ETag` is the simplest way to make sure it does. ([RFC 9111: no-cache][rfc9111-no-cache])
- Get Data Holder Brands supports `page` (default 1), `page-size` (default 25, maximum 1000) and `updated-since`. It returns `links` and `meta` as the CDS pagination rules require. `links` has `first` and `last` when the list has records, `prev` after the first page and `next` before the last page. Each link has `page`, `page-size` and the `updated-since` of the request. ([CDS: Pagination][cds-pagination], [CDS: Get Data Holder Brands][cds-get-dh-brands])
- Get Data Holder Brands returns the brands that have a `dataHolderBrandId`, in each status (ACTIVE, INACTIVE and REMOVED), in the order of their IDs. `updated-since` gives the brands with a `lastUpdated` at or after that time. The response has no `ETag`. (See Decisions)
- The `lastUpdated` of a brand on Get Data Holder Brands is the time of the last change to its data holder, as on Get Data Holder Brands Summary.
- Other list APIs are not paginated. They return `links.self` and an empty `meta`. ([CDS: Register API schemas][cds-register-schemas])
- Status and recipient lists include entities in all statuses, not only ACTIVE. The Register must tell data holders about status changes through these APIs, so it must return the statuses that are not ACTIVE. ([CDS: Participant Statuses][cds-participant-statuses])
- Get Data Holder Brands Summary includes brands that have only an `interimId` and no `dataHolderBrandId`. ([CDS: DataHolderBrandSummaryV2][cds-schema-brand-summary], [ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary])
- Get Data Holder Brands Summary includes ACTIVE and INACTIVE brands, and not REMOVED brands. An `{industry}` other than `all` gives only the brands that have that industry. ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary]; see Decisions for the statuses)
- The `lastUpdated` of a brand on Get Data Holder Brands Summary is the time of the last change to its data holder. Judgement: the same rule as the data recipient `lastUpdated` (see Decisions).
- All responses from unauthenticated APIs, including errors, have `Access-Control-Allow-Origin: *`. ([CDS: CORS][cds-cors])
- All responses from unauthenticated APIs have `Access-Control-Expose-Headers: x-v, etag, x-fapi-interaction-id`, so that a web page can read these headers. ([Fetch: CORS protocol][fetch-cors]; judgement: the CDS requires that CORS protections are disabled ([CDS: CORS][cds-cors]), and a browser hides these headers otherwise)
- A CORS preflight request (`OPTIONS` with `Access-Control-Request-Method`) to an unauthenticated API gives 204 with `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods: GET`, and an `Access-Control-Allow-Headers` that contains the requested headers. ([Fetch: CORS protocol][fetch-cors]; judgement: see Decisions)

### Authorisation server

- Get OpenId Provider Config returns all fields that the OpenAPI schema marks as required. It advertises `client_credentials`, `private_key_jwt`, scope `cdr-register:read` and `tls_client_certificate_bound_access_tokens: true`. ([CDS: Get OpenId Provider Config][cds-get-oidc], [CDS: Register Endpoints][cds-register-endpoints])
- The token endpoint accepts only `grant_type=client_credentials` with `client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer`. ([CDS: Data Recipients calling the CDR Register][cds-dr-calling-register])
- The client assertion `iss`, `sub` and `client_id` are the software product ID. The CDS also allows the brand ID but says it will be deprecated. Judgement: support only the software product ID. ([CDS: Data Recipients calling the CDR Register][cds-dr-calling-register])
- The Register validates the client assertion signature with the keys at the software product's JWKS URI. ([CDS: Client Authentication][cds-client-auth])
- The Register accepts PS256 and ES256 client assertions. The CDS does not name the algorithms for the Register token endpoint, and it refers to FAPI 1.0 Advanced for algorithms elsewhere. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms])
- The Register accepts its issuer identifier, its token endpoint URL, or the URL of the endpoint called as the assertion `aud`. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- The Register rejects a client assertion whose `jti` it has already accepted. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- The Register issues a token only when the TLS client certificate is one of the certificates registered for the software product. It checks the full certificate, not only the common name. ([ACCC Developer Portal: Token Endpoint][accc-token]; the ACCC mock checks only the common name, [TokenController.cs][accc-mock-token])
- The Register issues a token only when the software product is ACTIVE. ([ACCC Developer Portal: Token Endpoint][accc-token])
- Access tokens are bound to the client certificate as RFC 8705 specifies: `cnf.x5t#S256` is the base64url SHA-256 thumbprint. ([CDS: Holder of Key Mechanism][cds-hok]; the ACCC mock uses a SHA-1 hex thumbprint, [TokenController.cs][accc-mock-token])
- Access tokens are JWTs signed with PS256. They expire after 5 minutes. The Register issues no refresh tokens. The CDS gives no lifetime; its example shows 7200 seconds and is not normative. ([ACCC Developer Portal: Token Endpoint][accc-token])
- Get Data Holder Brands and Get SSA reject a token when the TLS client certificate does not match the token's certificate binding. ([CDS: Holder of Key Mechanism][cds-hok])
- Get Data Holder Brands and Get SSA check the caller's status on each call. When the software product or its data recipient is not ACTIVE, the response is 403 with `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`. ([CDS: Error Codes][cds-error-403-adr-not-active])

### Software Statement Assertion

- The Register issues an SSA only for an ACTIVE software product of an ACTIVE data recipient. ([CDS: SSA][cds-ssa])
- The SSA contains all required claims in the SSA definition, with `iss` = `cdr-register`, `org_id` = the brand ID and `software_id` = the software product ID. It also has `legal_entity_id`, `legal_entity_name` and each optional URI that the software product has. It has a `jti` that the Register does not use again. ([CDS: SSA Definition][cds-ssa-definition])
- The Get SSA response body is the SSA as a compact JWT with no JSON quotes, with the content type `application/json`. It has no `ETag`. ([CDS: Get SSA][cds-get-ssa]; see Decisions)
- The SSA is signed with PS256, with a key published at Get JWKS. ([CDS: SSA][cds-ssa], [CDS: Get JWKS][cds-get-jwks])
- The SSA expires 10 minutes after it is issued. ([CDS: SSA][cds-ssa])
- When the path `softwareProductId` is not the caller's software product, the response is 404 with `urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct`. When the path `dataRecipientBrandId` is not the caller's brand, the response is 404 with `urn:au-cds:error:cds-register:Field/InvalidBrand`. The Register checks the software product first. The `detail` is the path value. ([CDS: CDR Register Errors][cds-register-errors], [CDS Known Issues: Standardise Register API Error Codes][cds-known-register-errors])
- The SSA `scope` claim contains the scopes registered for the software product, with a space between them. ([CDS: SSA Definition][cds-ssa-definition]; see Decisions)

## Signing keys

- The Register has one key set to sign access tokens (the OpenID Provider `jwks_uri`) and a different key set to sign SSAs (Get JWKS). ([CDS: Get OpenId Provider Config][cds-get-oidc])
- Each published JWK has `alg`, `e`, `key_ops`, `kid`, `kty` and `n`. Keys are RSA, for PS256. ([CDS: JWK][cds-schema-jwk]) RSA keys are 2048 bits or longer. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms])
- The Register can change the key of each key set. During a key change, the JWKS publishes the old and the new key. Judgement: participants cache the JWKS, so a new key must be published before the Register uses it. A JWKS can hold more than one key. ([RFC 7517: JWK Set][rfc7517-jwks])
- A key change has two steps. First the JWKS publishes the next key, and the Register still signs with the old key. Then the Register signs with the next key, and the JWKS still publishes the old key. An SSA or an access token that the old key signed stays valid until it expires. (See Decisions)
- A key change of one key set does not change the other key set.

## Transport security

- All CDR endpoints use TLS 1.2 or higher, with the cipher suites that the CDS permits. ([CDS: Transaction Security][cds-transaction-security], [CDS: Ciphers][cds-ciphers])
- With TLS 1.2, the only cipher suites are `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256` and `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`. TLS 1.3 is also permitted. The CDS permits only the suites that [FAPI 1.0 Advanced, section 8.5][fapi-tls] permits and that [BCP 195][bcp195] recommends, and with an RSA key these two suites are in both lists. ([CDS: Ciphers][cds-ciphers])
- mTLS endpoints present a server certificate from the mock CDR CA. They accept only client certificates that chain to the mock CDR CA. ([CDS: Certificate Trust Model][cds-cert-trust-model], [CDS: Transaction Security][cds-transaction-security])
- TLS endpoints present a certificate from a mock public CA, not from the mock CDR CA. TLS endpoints do not request client certificates. ([CDS: Participant Endpoints][cds-participant-endpoints], [CDS: Transaction Security][cds-transaction-security])
- The Register checks the revocation status of each client certificate with the newest CRL of the mock CDR intermediate, which the Register signs and which is at the URL that the certificate gives. The mock CDR root signs only the intermediate, so this CRL covers each client certificate that chains to the mock CDR CA. The Register refuses the certificate when it is revoked, when the intermediate did not sign it, and when the Register has no current CRL. A revocation takes effect at the Register at once. ([CDS: Certificate Validation][cds-cert-validation]; see Decisions)
- When a proxy terminates TLS, the Register accepts the client certificate only from that trusted proxy. A client cannot supply the certificate in a request header. Judgement: the ACCC mock trusts forwarded certificate headers, and a client can possibly forge them ([research note][research-accc-headers]).

What the Register needs from the mock PKI:

- A root CA and an issuing intermediate CA, in the same structure as production. ([ACCC Developer Portal: Certificate Management][accc-cert-mgmt])
- Participant certificates that follow the CSR profile: RSA 2048, SHA-256, OU `Consumer Data Right`. The client certificate CN is the software product name, and its O is the brand name. The server certificate CN is the primary DNS name, and its O is the legal entity name or the brand name. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- A separate mock public CA for TLS endpoints.
- The intermediate certificate and its key, which the Register reads when it starts, so that it can issue certificates, revoke certificates and sign CRLs ([ADR 0022][adr-0022]). The deployment decides who can read the key file ([PKI DESIGN.md: Keys][pki-design-keys]).

## Certificate revocation list

In a running lab, the Register is the only component that revokes a certificate of the mock CDR intermediate and signs its CRL ([ADR 0022][adr-0022]). The CRL follows the CRL format of the [PKI DESIGN.md][pki-design-revocation]. The Register serves the CRL at the URL in each CDR certificate.

- The Register signs a new CRL when it starts, at each revocation, and no later than 24 hours after the `thisUpdate` of the current CRL, also when no revocation occurred. ([ACCC CP, section 4.9.7, Table 1][accc-cp]: "CRLs shall be issued periodically, even if there are no changes", "at least once every 24 hours". See Decisions.)
- The next update of a CRL is 7 days after its issue ([PKI DESIGN.md][pki-design-revocation]). Thus each CRL that the Register serves is current for 6 days or more, and a relying party that keeps a CRL for 1 minute never keeps a CRL after its next update.
- The CRL lists each revocation of the intermediate that the Register keeps, also the revocations before a restart.
- Each new CRL has a CRL number that is higher than the CRL number of each earlier CRL of the intermediate, also after a restart. A CRL number is never used two times. ([ACCC CP, section 10, Table 17][accc-cp]: "monotonically increasing integer (never repeated)"; [RFC 5280, section 5.2.3][rfc5280-crl-number])
- The Register serves only its newest CRL. A request never gets a CRL that is only partly written. ([ACCC CP, section 4.9.7][accc-cp]: "superseded certificate status information is removed"; [ACCC CPS, section 4.10.2][accc-cps]: "the public copy of the CRL is replaced automatically")
- The CRL listener is a plain HTTP listener, separate from each other listener of the Register. It needs no authentication, and it does not check the `Host` header. ([RFC 5280, section 4.2.1.13][rfc5280-crldp]; see Decisions.)
- `GET /crl/cdr-intermediate.crl` gives 200 with the current CRL in DER and the content type `application/pkix-crl` ([RFC 2585, section 4.2][rfc2585-mime]). `HEAD` gives the same headers with no body ([RFC 9110, section 9.1][rfc9110-methods]). Another method gives 405 with an `Allow` header ([RFC 9110, section 15.5.6][rfc9110-405]). Another path gives 404.
- The URL of the CRL is a setting of the Register. The setting is required, and has no default. Each certificate that the Register issues gives this URL as its CRL distribution point ([PKI DESIGN.md: Certificate profile][pki-design-profile]). The lab set-up gives the same URL in the certificates that it issues.
- A revocation through the operator interface takes effect at the Register at once: the next TLS handshake with the certificate fails. Other relying parties see it within 1 minute ([Transport security](#transport-security)). (See Decisions.)
- The readiness probe fails while the Register has no current CRL. ([Non-functional requirements](#non-functional-requirements))
- Tests: Register tests with a controlled clock check the schedule of the CRL signatures and the CRL numbers across a restart, with no wait. The conformance suite checks the CRL at its URL, and a revocation through the operator interface. ([ADR 0019][adr-0019])

## Operator interface

In production, the CDR Registrar keeps the participant data and statuses on the Register ([CDS: Participant Statuses][cds-participant-statuses]). The operator interface makes these changes in the lab. The interface of the production Register is not public, so the lab defines its own interface. Real components must not depend on it.

- The operator can suspend, revoke, surrender and reinstate a data recipient legal entity. The operator can activate, deactivate and remove a software product. The status rules apply, including the cascades.
- A reinstatement to ACTIVE does not change the status of the software products. The operator activates each software product. ([CDS: Status Mapping][cds-status-mapping]: the cascade applies when the status changes "for any status other than active")
- The operator can create, change and remove participants, brands and software products:
  - A create makes a data recipient or a data holder. The request body is one legal entity in the format of the load file, with its brands and, for a data recipient, their software products. A create of a legal entity that exists gives 409.
  - An add makes a brand of a legal entity, or a software product of a data recipient brand. The request body is one brand or one software product in the format of the load file. A data recipient brand can have its software products in the body.
  - A create and an add give 201 with no body. The `Location` header gives the path of the new record in the operator interface. ([RFC 9110, section 15.3.2][rfc9110-201])
  - A change is a `PUT` on the path of the record. The body gives each field of the record in the format of the load file, except the identifiers, the status, the brands, the software products and the client certificates. A field that is not in the body has no value after the change. A change gives 204. ([RFC 9110, section 9.3.4][rfc9110-put])
  - The identifiers of a record do not change. A data holder brand is in a path with its `dataHolderBrandId`, or with its `interimId` when it has no `dataHolderBrandId`.
  - A removal is a status action: the operator removes a data holder legal entity, a brand or a software product. A data recipient legal entity has no REMOVED status. Its revocation or its surrender is its removal. ([CDS: Participant Statuses][cds-participant-statuses])
- The operator can activate, deactivate and remove a data holder brand or a data recipient brand, and remove a data holder legal entity. The status rules apply, including the cascades.
- After a create, an add or a change, the participant obeys the participant model and the status rules, the same as after a load. The Register checks the full legal entity, with its brands and software products.
- A body that is not one JSON object, that has a field that the action does not accept, or that is larger than 1 MiB, gives 400. A body that does not obey the participant model gives 400. A body that obeys the participant model but breaks a status rule gives 409, for example an ACTIVE software product of an INACTIVE brand. The `detail` names each record and rule that fails.
- An identifier in the body that a different record has gives 409, also when that record is REMOVED or is a record of a different kind of participant. The `detail` gives the identifier. (See Participant model.)
- A create, an add and a change are visible on the CDR APIs immediately, as a status change is ([Status rules](#status-rules)). For example:
  - A new data holder brand is on Get Data Holder Brands Summary, and also on Get Data Holder Brands when it has a `dataHolderBrandId`. A new data holder legal entity is on Get Data Holder Statuses.
  - A new data recipient is on Get Data Recipients and Get Data Recipients Statuses, and its software products are on Get Software Products Statuses.
  - The next SSA of a software product has its changed metadata. The token endpoint uses the changed JWKS URI of a software product for the next client assertion.
  - The change sets the `lastUpdated` of the legal entity, and changes the `ETag` of each API whose response data changes.
- The operator can register and remove client certificates for a software product. The request body of a registration is one PEM certificate. The Register keeps the SHA-256 thumbprint of the certificate, not the certificate. A registration of a certificate that is already registered changes nothing and gives 204. A REMOVED software product cannot get a certificate (409). A body that is not one PEM certificate gives 400.
- A removal gives the thumbprint of the certificate in the path, in the base64url format of RFC 8705 `x5t#S256`. A removal of a thumbprint that is not registered changes nothing and gives 204. A thumbprint that is not a base64url SHA-256 value gives 400. A removal is permitted in each status of the software product. (See Decisions.)
- After a removal, the token endpoint refuses the certificate. An access token that the Register issued before the removal stays valid until it expires. (See Decisions.)
- The operator can get a client certificate for a software product from the mock CDR intermediate CA ([ADR 0022][adr-0022]). The request body is one PEM CSR. The Register never gets the private key. ([ACCC CPS, section 3.2.1][accc-cps]) A body that is not one PEM CSR gives 400.
- The CSR follows the CSR profile, and the certificate follows the certificate profile, of the [PKI DESIGN.md][pki-design]. A CSR that breaks a rule gives 400, and the `detail` names each rule that it breaks.
- The CN of the CSR subject is the name of the software product, and the O is the name of its brand, with the same characters. If not, the issue gives 400, and the `detail` names each field that does not agree. ([CDS: Certificate Signing Request Profile][cds-csr-profile]; see Decisions.)
- The issue also registers the certificate for the software product, with the rules of a registration. The token endpoint accepts the certificate immediately. A REMOVED software product cannot get a certificate (409). (See Decisions.)
- The response of an issue gives the certificate and then the intermediate certificate, in PEM, with the content type `application/pem-certificate-chain`. The Register keeps only the thumbprint of the certificate, as for a registration, so the operator cannot get the certificate again. (See Decisions.)
- The operator can get the CDR CA bundle: the intermediate and then the root certificate of the mock CDR CA, in PEM, with the content type `application/pem-certificate-chain`. It is the bundle that the mTLS listeners of the Register trust. (See Decisions.)
- The operator can get a CDR server certificate for a data holder brand from the mock CDR intermediate CA ([ADR 0022][adr-0022]). The CDS gives server certificates to data holders, and a data recipient does not need one ([CDS: Certificate Management][cds-cert-mgmt]). The ACCC issues a server certificate for a brand ([CDR Support: Brands in the CDR ecosystem][support-brands], [ACCC Participant Portal User Guide, section 10.6.1][accc-portal-guide]). The Register issues no server certificate for a data holder legal entity, a data recipient or a software product. (See Decisions.)
- The request body, the CSR profile, the certificate profile and the response are the same as for the issue of a client certificate. The CSR profile and the certificate profile are those of a server certificate in the [PKI DESIGN.md][pki-design].
- The O of the CSR subject is the name of the brand or the name of its legal entity, with the same characters. If not, the issue gives 400, and the `detail` gives the two names that are permitted. ([CDS: Certificate Signing Request Profile][cds-csr-profile]; see Decisions.) The name of a data holder legal entity has 36 characters or fewer, so a brand name of more than 64 characters does not stop the issue.
- The Register does not compare the CN and the DNS names of the CSR with the endpoints of the brand. (See Decisions.)
- A REMOVED brand cannot get a server certificate (409). An INACTIVE brand can get one. A brand with no `dataHolderBrandId` cannot get a server certificate (409). (See Decisions.)
- The Register does not register a server certificate, because no API of the Register uses it. The Register records the issue, with the thumbprint of the certificate, as an event of the data holder, and it does not keep the certificate. The issue is a change to the data holder, so it sets the `lastUpdated` of each brand of the data holder, as each other event of the data holder does. (See Decisions.)
- The operator can revoke a certificate of the mock CDR intermediate ([ADR 0022][adr-0022]). This includes a certificate that the lab set-up issued, which the Register does not know. The request body is a JSON object with two fields, and each is required:
  - `certificate`: one PEM certificate, or a PEM chain with the certificate first, as an issue gives it. The Register revokes the first certificate.
  - `reason`: one of the revocation reasons of the [PKI DESIGN.md][pki-design-revocation]: `unspecified`, `keyCompromise`, `affiliationChanged`, `superseded`, `cessationOfOperation` or `privilegeWithdrawn`. ([ACCC CP, section 4.9.3][accc-cp]: a request "shall identify the certificate to be revoked, explain the reason for revocation")
- A revocation gives 204. When it completes, the CRL lists the certificate with its reason. (See Decisions.)
- A body that is not a JSON object with these two fields, a `certificate` that does not start with a PEM certificate or that has a PEM block that is not a certificate, and a `reason` that is not in the list, give 400. A certificate that the mock CDR intermediate did not sign gives 400. The `detail` names each field that fails.
- A certificate that is already revoked gives 409, with no change. The `detail` gives the reason of the revocation that exists. ([PKI DESIGN.md][pki-design-revocation]; see Decisions.)
- The Register does not check the validity period of the certificate, and it does not compare the certificate with the participants. (See Decisions.)
- A revocation does not remove a registered client certificate of a software product. It is not an event of a participant, and the history of a participant does not show it. (See Decisions.)
- The operator can publish the next key of a key set, and then activate it. A publish when the key set has a next key gives 409. An activate when the key set has no next key gives 409. An unknown key set gives 404.
- The operator interface is on a separate listener from the CDR APIs.
- Each request must have the operator token as a bearer token ([RFC 6750][rfc6750]). This includes a request for an unknown path. A request with no bearer token gets 401 with `WWW-Authenticate: Bearer`. A request with a different bearer token gets 401 with `WWW-Authenticate: Bearer error="invalid_token"`. ([RFC 6750, section 3.1][rfc6750-errors])
- The operator interface refuses each request, with each method and path, whose `Host` header is not in an allowlist. The allowlist is a setting with the values of the `Host` header: a host and a port, for example `127.0.0.1:8082`, or a host with no port for a client that uses the default port of its scheme, for example through a TLS proxy ([RFC 9110, section 7.2][rfc9110-host]). The setting is required, and has no default. The comparison of the host name has no difference for upper and lower case. The response is 403 with a problem details body, and the interface does not examine the token. ([Vite advisory GHSA-vg6x-rcgg-rjx6][vite-dns-rebinding]: a plain HTTP server with no check of the `Host` header is open to DNS rebinding. See Decisions.)
- The operator interface does not permit cross-origin requests.
- A change that the status rules do not allow gives 409, and the `detail` names the rule. A change to a record that does not exist gives 404. Error responses are [RFC 9457][rfc9457] problem details with the type `about:blank`.

| Action | Method and path | Success |
|---|---|---|
| Suspend, revoke, surrender or reinstate a data recipient | `POST /v1/data-recipients/{legalEntityId}/{suspend,revoke,surrender,reinstate}` | 204 |
| Activate, deactivate or remove a software product | `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/{activate,deactivate,remove}` | 204 |
| Register a client certificate for a software product | `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates` | 204 |
| Remove a client certificate of a software product | `DELETE /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/{thumbprint}` | 204 |
| Issue a client certificate for a software product from a CSR, and register it | `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue` | 200 |
| Get the CDR CA bundle | `GET /v1/cdr-ca-bundle` | 200 |
| Create a data recipient | `POST /v1/data-recipients` | 201 |
| Change the details of a data recipient | `PUT /v1/data-recipients/{legalEntityId}` | 204 |
| Add a brand to a data recipient | `POST /v1/data-recipients/{legalEntityId}/brands` | 201 |
| Change the details of a data recipient brand | `PUT /v1/data-recipients/{legalEntityId}/brands/{brandId}` | 204 |
| Activate, deactivate or remove a data recipient brand | `POST /v1/data-recipients/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 |
| Add a software product to a data recipient brand | `POST /v1/data-recipients/{legalEntityId}/brands/{brandId}/software-products` | 201 |
| Change the details of a software product | `PUT /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}` | 204 |
| Create a data holder | `POST /v1/data-holders` | 201 |
| Change the details of a data holder | `PUT /v1/data-holders/{legalEntityId}` | 204 |
| Remove a data holder | `POST /v1/data-holders/{legalEntityId}/remove` | 204 |
| Add a brand to a data holder | `POST /v1/data-holders/{legalEntityId}/brands` | 201 |
| Change the details of a data holder brand | `PUT /v1/data-holders/{legalEntityId}/brands/{brandId}` | 204 |
| Activate, deactivate or remove a data holder brand | `POST /v1/data-holders/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 |
| Issue a server certificate for a data holder brand from a CSR | `POST /v1/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue` | 200 |
| Revoke a certificate of the mock CDR intermediate | `POST /v1/cdr-certificates/revoke` | 204 |
| Publish the next key of a key set, or activate it | `POST /v1/key-sets/{keySet}/{publish-next,activate-next}`, where `{keySet}` is `access-token` or `ssa` | 204 |

## Operator console

The operator console is a web UI for the lab operator. It shows the participant data, the statuses and the signing keys, and it does the operator actions. It is a lab-only feature: real components must not depend on it. The architecture is in [ADR 0014][adr-0014], the design system in the [design system DESIGN.md][ui-design], and the contract of its API in [ADR 0017][adr-0017].

### Functions

The console shows:

- A list of the data recipient legal entities, with the name, the accreditation number, the status and the number of brands and software products. The operator can filter the list by status and search it by name.
- A list of the data holder legal entities, with the name, the legal entity ID, the status, the number of brands and the industries of its brands. The operator can filter the list by status, and search it by the name of the legal entity or of a brand. (See Decisions.)
- The detail of a data recipient: the legal entity (ID, name, accreditation number, accreditation level, logo URI, status and the time of the last change), each brand (ID, name, logo URI and status), and each software product with its ID, its status, the metadata of the participant model and the thumbprints of its registered client certificates.
- The detail of a data holder: the legal entity (ID, name, logo URI, status and the time of the last change), and each brand with its brand ID or interim ID (or both), name, brand group, logo URI, status, industries, endpoints and JWKS endpoint.
- The history of a participant: each event of the stream of its legal entity, newest first, with the time, the event type and the fields that the table below gives.
- The two key sets (access token and SSA), with the key ID, the state and the time of creation of each key. The states are: the key that signs, the next key, and the previous key that the JWKS still publishes.
- The CDR certificates: the CDR CA bundle in PEM, and a list of the revocations of the mock CDR intermediate, newest first, with the serial number, the subject, the thumbprint, the time and the reason of each revocation. The list includes the certificates of the lab set-up that the operator revokes. (See Decisions.)

Each row of a list links to the detail of its record, and the detail of a participant links to its history.

The history shows only these fields of each event. Judgement: the console shows only data that a domain type gives ([ADR 0011][adr-0011]), and these fields name the record that the event changes. The detail shows the current values of the other fields.

| Event type | Text in the history | Fields |
|---|---|---|
| `DataRecipientImported`, `DataHolderImported` | Loaded into the Register | None |
| `DataRecipientCreated`, `DataHolderCreated` | Created by the operator | None |
| `DataRecipientDetailsChanged`, `DataHolderDetailsChanged` | Legal entity details changed | None |
| `DataHolderLegalEntityRemoved` | Data holder removed | None |
| `DataRecipientBrandAdded`, `DataHolderBrandAdded` | Brand added | Brand ID and name |
| `DataRecipientBrandDetailsChanged`, `DataHolderBrandDetailsChanged` | Brand details changed | Brand ID and name |
| `DataRecipientBrandActivated`, `DataHolderBrandActivated` | Brand activated | Brand ID and name |
| `DataRecipientBrandDeactivated`, `DataHolderBrandDeactivated` | Brand deactivated | Brand ID and name |
| `DataRecipientBrandRemoved`, `DataHolderBrandRemoved` | Brand removed | Brand ID and name |
| `SoftwareProductAdded` | Software product added | Software product ID and name |
| `SoftwareProductDetailsChanged` | Software product details changed | Software product ID and name |
| `AccreditationSuspended` | Accreditation suspended | None |
| `AccreditationRevoked` | Accreditation revoked | None |
| `AccreditationSurrendered` | Accreditation surrendered | None |
| `AccreditationReinstated` | Accreditation reinstated | None |
| `SoftwareProductActivated` | Software product activated | Software product ID and name |
| `SoftwareProductDeactivated` | Software product deactivated | Software product ID and name |
| `SoftwareProductRemoved` | Software product removed | Software product ID and name |
| `ClientCertificateRegistered` | Client certificate registered | Software product ID and name, and the thumbprint |
| `ClientCertificateIssued` | Client certificate issued | Software product ID and name, and the thumbprint |
| `ClientCertificateRemoved` | Client certificate removed | Software product ID and name, and the thumbprint |
| `ServerCertificateIssued` | Server certificate issued | Brand ID and name, and the thumbprint |

A new event type adds a row to this table in the same change.

The console does the actions of the operator interface, with the same rules ([Operator interface](#operator-interface)):

- Suspend, revoke, surrender and reinstate a data recipient legal entity.
- Activate, deactivate and remove a data recipient brand, a software product and a data holder brand.
- Remove a data holder legal entity.
- Register a client certificate of a software product: the operator gives one PEM certificate as text or as a file. Remove a client certificate: the operator selects one of the registered thumbprints.
- Issue a client certificate for a software product, and issue a server certificate for a data holder brand: the operator gives one PEM CSR as text or as a file.
- Save the CDR CA bundle as a file.
- Revoke a certificate of the mock CDR intermediate: the operator gives one PEM certificate, or the PEM chain of an issue, as text or as a file, and selects the reason.
- Publish the next key of a key set, and activate it.
- Create a data recipient or a data holder: the legal entity, with no brands. Add a brand to a legal entity, and add a software product to a data recipient brand. (See Decisions.)
- Change the details of a legal entity, a brand or a software product: the fields that a change in the operator interface replaces.

Rules for the actions:

- Before each action except a create, an add, a change, the registration of a client certificate and the issue of a certificate, the console asks the operator to confirm. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]: the registration, an issue and a change are reversible, and the Register checks each value of a create, an add, a change and an issue, and the operator can correct it. See Decisions.)
- The confirmation names the action, its record and each other record that the action also changes, with the status before and after. For example, it names each software product that a suspension makes INACTIVE. (Judgement: a cascade changes records that the operator did not select.)
- When the action or its cascade gives a final status (REMOVED, SURRENDERED), the confirmation says that this cannot be reversed. A revocation is an example: the data recipient can become ACTIVE again, but its software products cannot. ([CDS: Participant Statuses][cds-participant-statuses])
- The confirmation of a reinstatement says that the software products do not change, and that the operator must activate each one. ([CDS: Status Mapping][cds-status-mapping])
- The confirmation of a key publish names the previous key that leaves the JWKS, if there is one. The confirmation of a key activation names the key that starts to sign and the key that becomes the previous key, and says that a participant with an old copy of the JWKS cannot verify the new signatures. (Judgement: an activation cannot be reversed, and a publish deletes a key. See [Signing keys](#signing-keys).)
- The console offers only the actions that the rules permit now for the record. The console API gives these actions (see Console API). The UI has no copy of the rules.
- When the Register refuses an action, the console shows the `detail` of the problem in the dialog or the form of the action. A refused PEM certificate shows as the error of its field. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- After an action, also after a refused action, each view shows the current state without a reload of the page. (Judgement: a refusal often means that the view is not current.)
- The console never sends an action again by itself. (Judgement: an automatic second request can apply a change that the operator did not confirm.)
- The console shows the result of a successful action in a toast. The view also shows the new state, so the toast is never the only place of the result. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]) An issue shows its result in its dialog, not in a toast (see the rules for the CDR certificates).
- For a certificate file or a CSR file, the console puts the text of the file in the text field of the certificate or the CSR, and the operator can see it before the action. The console does not read the certificate or the CSR: the Register checks it. (Judgement: one value to send, and one set of rules in the Register.)

Rules for the CDR certificates:

- The software product card has the button "Issue certificate" after "Register certificate", in its action row. The data holder brand card has the button "Issue certificate" after its status actions, in its action row. The console shows each button only when the console API permits the action (see Console API). ([Design system DESIGN.md][ui-design]: buttons that act on the record are in its action row)
- The side navigation has the section "CDR certificates", after "Signing keys", with its own page. The page shows the CDR CA bundle with a "Save as file" button, then a "Revoke certificate" button in an action row, then the list of revocations. ([Design system DESIGN.md][ui-design]; see Decisions)
- An issue is a dialog with one PEM CSR field and a CSR file input, with the same rules as the registration of a client certificate. The help text of the CSR field says that the Register needs only the CSR and never the private key. ([ACCC CPS, section 3.2.1][accc-cps]: the CA never sees the private key)
- The help text of the CSR field gives the subject that the Register requires. For a client certificate: the CN is the name of the software product, and the O is the name of its brand, with the names of the record. For a server certificate: the CN is the primary DNS name, and the O is the name of the brand or of its legal entity, with the names of the records. ([CDS: Certificate Signing Request Profile][cds-csr-profile], [WCAG 2.2, SC 3.3.2][wcag-3-3-2]; see Decisions) The console does not compare the CSR with these names: the Register checks them.
- When the Register refuses the CSR (400), the console shows the `detail` of the problem as the error of the CSR field. A different refusal, for example a REMOVED record or a stale version (409), shows in an alert in the dialog. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- After a successful issue, the dialog stays open and shows the result in place of the form: a heading that names the certificate type and the record, the thumbprint of the certificate, the PEM chain in a text area that the operator cannot change, a "Save as file" button and a "Close" button. The focus moves to the heading of the result. ([WAI-ARIA APG: Dialog][apg-dialog]; see Decisions)
- The result says that the chain is the certificate and then the intermediate certificate, and that the Register does not keep the certificate, so the console cannot show it again. It tells the operator to save the chain before they close the dialog. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2]; see the row "Response of an issue" in Decisions)
- The result of a client certificate issue says that the Register registered the certificate for the software product, and that the token endpoint accepts it now. The result of a server certificate issue says that the Register does not register a server certificate, and that a TLS client that trusts the CDR CA bundle accepts it. ([Operator interface](#operator-interface))
- The operator closes the result with "Close" or the Escape key, with no confirmation, also when they did not save the chain. (See Decisions.)
- The views get the current state after the issue, while the result shows: for example, the software product card shows the new thumbprint.
- "Save as file" saves the PEM text that the Register gave, with no change, as a file with the extension `.pem`: `client-certificate-{softwareProductId}.pem`, `server-certificate-{dataHolderBrandId}.pem` and `cdr-ca-bundle.pem`. The save sends no request to the Register, and needs no change to the console CSP. ([HTML: Downloading resources][html-download]; see Decisions)
- "Revoke certificate" opens a dialog with a PEM certificate field, a certificate file input and a reason. The reason is a native select with no selected value first, and the values of the [Operator interface](#operator-interface) in words, for example "Key compromise" for `keyCompromise`. The help text of the certificate field says that the field accepts one PEM certificate, or the file of an issue, and that the Register revokes the first certificate. (Judgement: a reason has no usual value, see the rule for a select in a create.)
- The submit button of the revocation dialog is "Continue". The Register checks the request with no change and gives the certificate (see Console API). When it refuses the request, the dialog shows each field error at its field, with the error summary at its top, as a form does, and a refusal that names no field in the summary. A certificate that is already revoked is a refusal with no field (409). (See Decisions.)
- After a successful check, the console asks the operator to confirm. The confirmation names the certificate with its subject, its serial number, its thumbprint and the end of its validity, and names the reason. It says that the CRL lists the certificate, that the Register refuses it at once and other relying parties within 1 minute, that a registration of the certificate for a software product does not change, and that the revocation cannot be reversed. ([Certificate revocation list](#certificate-revocation-list); see Decisions)
- After a successful revocation, the console shows the result in a toast, and the list of revocations shows the revocation first.

Rules for a create, an add and a change:

- Each create, add and change is a form on its own page, with its own URL. It is not a dialog. A URL of a form is never the URL of a record, so that a record with each ID, for example `new`, keeps its detail. (Judgement: the records have many fields. See Decisions.)
- The list of data recipients has the action "Create data recipient", and the list of data holders has "Create data holder". The detail of a participant has the actions "Change details" and "Add brand" for the legal entity, "Change brand" for each brand, "Add software product" for each data recipient brand, and "Change software product" for each software product. The console shows these actions for each record in each status. (See Decisions.) Each of these actions is a link directly after the heading of its record, and "Add software product" is at the heading "Software products" of its brand. The status actions and "Register certificate" are buttons after the details of their record and before its child records. ([Design system DESIGN.md][ui-design]) Each brand and each software product is a card that starts collapsed, with its status in its title bar. "Brands" and each "Software products" heading have "Show all" and "Hide all". After a form, the card of the record that the form changed or added is open. ([Design system DESIGN.md][ui-design])
- A create and an add give each field of one record in the format of the load file, with the identifiers and the status, and with no child records and no client certificates. ([Operator interface](#operator-interface))
- For each identifier of a create or an add, the form shows a new random UUID ([RFC 9562, section 5.4][rfc9562-v4]) in a field that the operator can change. The operator can keep it, or give a known identifier, for example the brand ID in the configuration of a lab service. A data holder brand gets a UUID in its `dataHolderBrandId` field, and its `interimId` field is empty. (See Decisions.)
- The status of a create or an add is a field. Its values are the status values of the record in the [Participant model](#participant-model), and ACTIVE is selected first. The status rules apply: for example, a software product of an INACTIVE brand must be INACTIVE or REMOVED. (See Decisions.)
- A change form shows the current value of each field, and shows the identifiers of the record as text that the operator cannot change. The console sends each field of the change, as a change in the operator interface does: an optional field with no value has no value after the change. ([Operator interface](#operator-interface))
- A change and an add send the version of the participant that the form loaded. A create has no version. A stale version gives 409, and the Register changes nothing (see the row "Action with a stale view" in Decisions).
- A choice of one value from a fixed set (the accreditation level, the status) is a native select. A choice from a fixed set of more than one value (the industries of a data holder brand) is a group of checkboxes. A list of values that the operator writes (the redirect URIs, the scopes) is a text area with one value on each line. The endpoints of a data holder brand are a group of fields with a legend. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; see Decisions)
- In a create, a select of a value that is required and that has no usual value (the accreditation level) has no selected value, so that the operator chooses it. (Judgement: a value that the console selects can go to the Register with no decision of the operator.)
- The label of each field that the console API schema does not require says "(optional)". The help text of the identifier fields of a data holder brand says that a brand with no `dataHolderBrandId` needs only its public details. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [GOV.UK Design System: Question pages][govuk-question-pages] marks the optional fields, not the required fields)
- The console removes the white space at the start and at the end of each value and of each line of a list, and ignores an empty line. It sends an optional field that has no value as no field. (Judgement: a value that the operator pastes often has white space, and a space at the end of a URI or an identifier is never intended.)
- The console does not check the values. The Register checks them with the rules of the participant model and the status rules. The UI has no copy of the rules. (Judgement: one set of rules, in the domain ([ADR 0013][adr-0013]).)
- When the Register refuses a value of a field, the console shows the `detail` of the field error as the error message of that field. The error of an item of a list shows at the field of the list. A refusal that names no field shows in the error summary of the form. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1], [SC 3.3.3][wcag-3-3-3])
- After a refusal, the form shows an error summary at its top, and the focus moves to the summary. The summary gives the `detail` of the problem, and a link to each field that has an error. ([GOV.UK Design System: Error summary][govuk-error-summary]; see the error summary in the [design system DESIGN.md][ui-design] and Decisions)
- After a refusal, the form keeps each value that the operator gave, and the version that it loaded. When the version is stale, the operator opens the form again to see the current values. (Judgement: the operator does not lose their work, and the console never replaces the change of a different person or tab with no decision of the operator.)
- While the request runs, the submit button is busy, and the form sends no second request. The console never sends a create, an add or a change again by itself.
- After a success, the console goes to the detail of the participant and shows the result in a toast. The focus moves to the heading of the detail. ([WCAG 2.2, SC 2.4.3][wcag-2-4-3], [SC 4.1.3][wcag-4-1-3]; judgement: the form page no longer exists)
- Each form has a "Cancel" link back to the page that opened it. A cancel sends no request, and the console does not ask the operator to confirm it. (See Decisions.)

### Presentation

- The console is for desktop browsers. It is usable at a viewport width of 1280 CSS px and more. Judgement: an operator uses a workstation, and [ADR 0014][adr-0014] gives desktop as the target for consoles.
- The console has the shell of a console of the design system, with the neutral lab brand for each role, and the compact density. ([design system DESIGN.md: Brands][ui-brands])
- The side navigation marks the section of each page as current, also when the URL of the page is not under the URL of the section. For example, each form of a data recipient, a brand or a software product marks "Data recipients". (Judgement: the operator sees where they are in the console, and a form is part of the section of its record.)
- The header and the title of the browser tab give the name "CDR Register". Judgement: the operator asked for it. The operator manages the Register, and "console" is the name of the component, not of what the operator sees.
- The filter, the sort and the page of a list are in the URL, so that a link or a reload shows the same list. Judgement: the operator can share or go back to a view.
- The search of a list applies while the operator types (see the filter bar in the [design system DESIGN.md][ui-design]). A search replaces the history entry of the list. A filter, a sort and a page add a history entry. Judgement: Back does not go through each pause in the typing.
- The detail and the history of a participant each have their own URL, and the page of the history is in the URL. Judgement: the same reason as for the lists. (See Decisions.)
- The CDR certificates page has its own URL, and the page of the list of revocations is in the URL. Judgement: the same reason as for the lists.
- The console shows each time in the time zone of the browser, with its offset from UTC, for example "2:05 pm UTC+8" (see the format of a time in the [design system DESIGN.md][ui-design]). The API gives each time in UTC. ([ADR 0012][adr-0012])
- The console loads no image from a participant. It shows a logo URI as text. ([OWASP ASVS 5.0, V3.6.1][asvs-v3]: the UI loads no file from another origin, and the console CSP refuses it)
- The console shows each URI of a participant as text with a copy button, not as a link. (See Decisions.)
- The history shows the event type and its text from the table in Functions. It does not group the events of one action. (See Decisions.)
- The console has the accessibility level of an operator console in the design system: the automated checks pass, with no other target.

### Listener and security

- The console is on its own listener, separate from the CDR APIs and from the operator interface. The listener serves the UI and the console API. ([ADR 0014][adr-0014])
- The console is off by default, and a setting turns it on. Judgement: a production Register has no console, so a deployment must turn it on explicitly.
- The listener is plain HTTP. Judgement: the same as the operator interface (see Decisions), and the console sets no cookie, so it needs no `Secure` attribute.
- The console has no login and the browser gets no credential. ([ADR 0014][adr-0014])
- The console refuses each request that is not GET, HEAD or OPTIONS, when the browser says that the request is cross-origin (`Sec-Fetch-Site` other than `same-origin` or `none`, or an `Origin` that is not the host). The response is 403 with a problem details body. ([Go: CrossOriginProtection][go-cop], [OWASP: CSRF Prevention][owasp-csrf])
- The console refuses each request, with each method, whose `Host` header is not in an allowlist. The allowlist is a setting with the values of the `Host` header: a host and a port, for example `127.0.0.1:8084`, or a host with no port for a client that uses the default port of its scheme, for example through a TLS proxy ([RFC 9110, section 7.2][rfc9110-host]). The setting is required when the console is on, and has no default. The comparison of the host name has no difference for upper and lower case. The response is 403 with a problem details body. ([Vite advisory GHSA-vg6x-rcgg-rjx6][vite-dns-rebinding]: a plain HTTP server with no check of the `Host` header is open to DNS rebinding. See Decisions.)
- The console permits no cross-origin requests: its responses have no CORS headers. ([OWASP ASVS 5.0, V3.4.2][asvs-v3])
- Each response has these headers ([OWASP ASVS 5.0, V3.4][asvs-v3]; [RFC 9700, section 4.16][rfc9700-clickjacking] for framing):
  - `Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'; require-trusted-types-for 'script'`
  - `X-Content-Type-Options: nosniff`
  - `Referrer-Policy: no-referrer`
  - `Cross-Origin-Opener-Policy: same-origin`
- The UI loads no file from another origin. ([OWASP ASVS 5.0, V3.6.1][asvs-v3])
- The responses of the console API have `Cache-Control: no-store`. The HTML page has `Cache-Control: no-cache`. A file whose name contains a hash of its content has `Cache-Control: public, max-age=31536000, immutable`. ([RFC 9111][rfc9111]; judgement: the operator always sees current data, and a new build of the UI loads at once)
- A GET for a path that is not an API path and not a file of the UI gives the HTML page of the UI, so that the UI shows its own route or its "not found" page. Judgement: the UI routes in the browser.

### Console API

- The console API is under the path `/api/`. It is JSON over HTTP. Judgement: one prefix separates the API from the paths of the UI.
- The console API is defined in an OpenAPI 3.0 document in the Register directory. Tests in the Register check each response against the document. Each object schema forbids properties that it does not name. ([ADR 0017][adr-0017], [ADR 0019][adr-0019])
- Error responses are [RFC 9457][rfc9457] problem details with the type `about:blank`, with the same status codes as the operator interface: 400 for a request that is not valid, 404 for a record that does not exist, and 409 for a change that the status rules do not allow. A refusal with field errors has a different type (see below).
- An unknown API path gives 404 with a problem details body.
- The console API gives the lists of participants and the history of a participant in pages. A request gives the page and the page size. The response gives the total number of records. Judgement: the CDS pagination is for the CDR APIs. A total lets the UI show the number of pages.
- The detail of a participant is one response, with each brand and software product, and is not in pages. The key sets are one response. Judgement: the records of one legal entity are one aggregate ([ADR 0003][adr-0003]), and a key set has three keys or fewer.
- The detail of a participant gives each brand and software product in each status, in the order of the record. The fields have the names of the CDS Register schemas. An optional field with no value is not in the response, and a list is always in the response. An unknown legal entity ID gives 404, and the `detail` does not contain the ID. (See Decisions.)
- The page size is 25 by default, and 100 at most. A page after the last page gives no records and the total. A page or page size that is not valid gives 400. (See Decisions.)
- The list of data recipients is filtered by status and searched by a part of the legal entity name, with no difference for upper and lower case. It is sorted by the name, the accreditation number, the status, the number of brands or the number of software products, and then by the legal entity ID. The numbers of brands and software products include each status. (See Decisions.)
- The list of data holders is filtered by status, and searched by a part of the legal entity name or of a brand name, with no difference for upper and lower case. The list is sorted by the name, the status or the number of brands, and then by the legal entity ID. The number of brands includes each status. (See Decisions.)
- The history of a participant is sorted by the stream version, newest first. Each entry gives the stream version, the event type, the time at which the Register recorded the event, and only the fields that the table in Functions gives. The console API gives no payload and no metadata of an event. ([ADR 0011][adr-0011], [ADR 0017][adr-0017]; judgement: the metadata is trace context, which is not participant data)
- The key sets give no private key and no public key. Judgement: the JWKS endpoints give the public keys, and a private key never leaves the Register.
- A query parameter that the console API does not know is ignored. An empty value is the same as no value. Judgement: the usual HTTP behaviour, and a link with an empty field from a form still works.
- The console API has no version in its path. ([ADR 0014][adr-0014]: the UI and the API ship together)
- The detail of a data recipient gives, for the legal entity, for each brand and for each software product, the status actions that the rules permit now. The detail of a data holder gives the same for the legal entity and for each brand. Each action gives each other record that it changes, with the status before and after. The key sets give the same for each key set, with the state of each key before and after. The Register finds these with the same domain rules as the change itself. (Judgement: the status rules are in the domain ([ADR 0013][adr-0013]), and a copy in the UI can give a confirmation that is not true. See Decisions.)
- The detail of a data recipient gives, for each software product, the client certificate actions that the rules permit now: `register`, `issue` and `remove`. The detail of a data holder gives, for each brand, the server certificate actions that the rules permit now: `issue`, only for a brand with a `dataHolderBrandId` that is not REMOVED. (See Decisions.)
- The detail of a data recipient and the detail of a data holder give the version of the participant, and each key set gives its version. A version changes at each change of its record. (See Decisions.)
- Each action is a `POST` with a JSON body. The body gives the version that the operator saw, except the body of a create, of the check of a revocation and of a revocation. When the participant or the key set has a different version now, the response is 409, the `detail` says that the record changed, and the Register changes nothing. ([ADR 0003][adr-0003]; judgement: the confirmation or the form was for the old version, so its list of changes or its values can be false now)
- An action that succeeds gives 204 with no body, except a create, an add, an issue and the check of a revocation. (Judgement: the same as the operator interface. The UI gets the data of its views again.)
- A create and an add are a `POST` on the path of the list of the record. They give 201 with no body, and a `Location` header with the path of the new record in the console API, which is the start of the paths of its actions. ([RFC 9110, section 15.3.2][rfc9110-201]; judgement: the same as the operator interface)
- A change is a `POST` on the path of the record, with `/change` at the end. Its body gives the version and each field of the change, with the names and the format of the operator interface. A field that is not in the body has no value after the change. (See Decisions.)
- The body of a create gives the fields of one record, as a create in the operator interface does, with no child records and no client certificates. The body of an add gives the version and the fields of one record, with no child records and no client certificates. A body that is not one JSON object, that has a field that the action does not accept, or that is larger than 1 MiB, gives 400. ([Operator interface](#operator-interface))
- When the Register refuses a value in the body, the problem details give each field error in the extension member `errors`, as in the example of [RFC 9457, section 3][rfc9457-3]. Each field error has a `pointer` and a `detail`. The `pointer` is a JSON Pointer to the field in the request body, in the URI fragment form ([RFC 6901, section 6][rfc6901-6]), for example `#/redirectUris` or `#/endpointDetail/publicBaseUri`. The `detail` says what is wrong with the value, in one sentence that the UI can show at the field. (See Decisions.)
- These refusals have field errors: a value that does not obey the participant model (400), an identifier that a different record has (409, at the field of the identifier), a legal entity ID that a legal entity has (409, at `#/legalEntityId`), and a status in the body that breaks a status rule (409, at `#/status`). The status code is the same as in the operator interface. A refusal that names no field, for example a stale version, has no `errors`.
- A problem with field errors has a type URI that the console API OpenAPI document defines with the `errors` member, and the `title` of that type. Its `detail` gives each problem that names no field. When each problem names a field, the `detail` says that values of the request are not valid. ([RFC 9457, section 3.2][rfc9457-3-2]: a problem type defines its extension members, and `about:blank` has no meaning other than the status code. See Decisions.)
- The registration of a client certificate gives the PEM certificate in the JSON body, with the rules of the operator interface for the certificate.
- An issue gives the version and the PEM CSR (`csr`) in the JSON body, with the rules of the operator interface for the CSR, the record and the certificate. A successful issue gives 200 with a JSON body: the PEM chain of the operator interface (`certificateChain`) and the thumbprint of the certificate in the base64url format of RFC 8705 `x5t#S256` (`thumbprint`). A CSR that the Register refuses gives 400, and the `detail` names each rule that it breaks, as in the operator interface. (See Decisions.)
- The CDR CA bundle is a `GET` that gives 200 with a JSON body: the PEM bundle of the operator interface (`caBundle`). (See Decisions.)
- The list of revocations is a `GET` in pages, with the rules of the other lists. Each revocation gives the serial number, the subject as an [RFC 4514][rfc4514] string, the thumbprint, the time of the revocation and the reason. The list is sorted by the time of the revocation, newest first, and then by the serial number. (See Decisions.)
- The console API gives a serial number in uppercase hexadecimal, with two digits for each byte and no separators, as `openssl x509 -serial` gives it. (See Decisions.)
- The check of a revocation and the revocation give the PEM certificate (`certificate`) and the reason (`reason`) in the JSON body, with the rules of the operator interface. The body has no version. A check makes no change. A successful check gives 200 with the subject, the serial number, the thumbprint and the end of the validity of the first certificate. A successful revocation gives 204. A value that the Register refuses gives a field error at `#/certificate` or `#/reason` (400), and a certificate that is already revoked gives 409 with no field errors, and its `detail` gives the reason of the revocation that exists. (See Decisions.)
- The console API puts no request body and no certificate, CSR or CA bundle in telemetry. A path with a thumbprint uses its route pattern in telemetry. ([ADR 0006][adr-0006])

The console API has these paths for the creates, the adds, the changes, the status actions of brands and data holders, and the CDR certificates. The other actions are in the OpenAPI document.

| Action | Method and path | Success |
|---|---|---|
| Create a data recipient | `POST /api/data-recipients` | 201 |
| Change the details of a data recipient | `POST /api/data-recipients/{legalEntityId}/change` | 204 |
| Add a brand to a data recipient | `POST /api/data-recipients/{legalEntityId}/brands` | 201 |
| Change the details of a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/change` | 204 |
| Activate, deactivate or remove a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 |
| Add a software product to a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/software-products` | 201 |
| Change the details of a software product | `POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/change` | 204 |
| Create a data holder | `POST /api/data-holders` | 201 |
| Change the details of a data holder | `POST /api/data-holders/{legalEntityId}/change` | 204 |
| Remove a data holder | `POST /api/data-holders/{legalEntityId}/remove` | 204 |
| Add a brand to a data holder | `POST /api/data-holders/{legalEntityId}/brands` | 201 |
| Change the details of a data holder brand | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/change` | 204 |
| Activate, deactivate or remove a data holder brand | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 |
| Issue a client certificate for a software product from a CSR, and register it | `POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue` | 200 |
| Issue a server certificate for a data holder brand from a CSR | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue` | 200 |
| Get the CDR CA bundle | `GET /api/cdr-ca-bundle` | 200 |
| Get one page of the revocations of the mock CDR intermediate | `GET /api/cdr-certificates/revocations` | 200 |
| Check a revocation with no change | `POST /api/cdr-certificates/check-revocation` | 200 |
| Revoke a certificate of the mock CDR intermediate | `POST /api/cdr-certificates/revoke` | 204 |

The `{brandId}` of a data holder brand is its `dataHolderBrandId`, or its `interimId` when it has no `dataHolderBrandId`, as in the operator interface.

### Non-functional requirements

- Telemetry: the console API has the same server telemetry as the other listeners. The UI sends no telemetry. ([ADR 0014][adr-0014], [ADR 0006][adr-0006])
- Performance: on a development workstation, a list of 1,000 data recipients or data holders shows within 1 second, and the first page of a history of 1,000 events shows within 1 second. Judgement: the seed data is small, and this limit finds a page that loads all records at once.
- Tests: tests in the Register check the listener headers, the cross-origin refusal, the refusal of a `Host` that is not in the allowlist, and the console API. The conformance suite does not test the console. A small number of browser tests check the main operator journeys against the lab. ([ADR 0019][adr-0019])
- Browser tests of the CDR certificates: the operator issues a client certificate from a CSR file, saves the chain, and then revokes the certificate from the saved file through the confirmation, and the list of revocations shows it. The operator issues a server certificate for a data holder brand and saves the chain, and the history of the data holder shows the issue. The operator saves the CDR CA bundle. Each test makes its own participants, so that it runs in a lab with no conformance data, and the repository has no private key for its CSRs ([Non-functional requirements](#non-functional-requirements-1)). (See Decisions.)

## Lab-only features

These features do not exist in the production Register. Real components must not depend on them.

- The operator can load a full set of participant data from a file into a Register that has no participants. The load is separate from the Register start, and the Register start does not need a file.
- The load validates the full set against the participant model and the status rules. If one record is not valid, the load changes nothing, and it names each record and rule that fails.
- The file uses the field names of the CDS Register schemas. Judgement: the operator then sees the same names in the file and in the API responses.
- The load is a command of the Register binary, and it needs access to the Register database.
- The Register is the mock CDR intermediate CA in a running lab ([ADR 0022][adr-0022]). It issues CDR client and server certificates from CSRs, gives the CDR CA bundle, revokes CDR certificates, and signs and serves the CRL. In production, the ACCC CA issues and revokes certificates through the Participant Portal, and the Register is not a CA.

## Deployment from published artifacts

An operator deploys the Register to Kubernetes from the image and the Helm chart that CI publishes ([Register chart][register-chart]). These requirements apply to the chart and to the example seed file, not to a deployment.

### Participant seed in the chart

- The chart loads a participant seed file only when the values give the name of an existing ConfigMap that holds the file. The chart does not make the ConfigMap. With no ConfigMap, the chart loads no file, and the Register starts with no participants. ([Kubernetes: ConfigMap volume][k8s-configmap-volume]; see Lab-only features and Decisions)
- Before the Register starts, the chart loads the file as the service user, after the migrations complete. The load gets no owner credentials. ([Kubernetes: init containers][k8s-init-containers]; [ADR 0020][adr-0020])
- The load changes the Register only when the Register has no participants. After a restart, or after a change of the ConfigMap, the Register has the same participants as before the restart or the change. (See Lab-only features and Decisions)
- Only the load gets the seed file. The Register process does not get it. Judgement: the Register start does not need a file (see Lab-only features).
- The chart README gives the seed value, the ConfigMap and its file, and the grant of the application role to the service user that the load needs.

### Example seed of the four preset banks

- The repository has an example seed file with the four preset banks of the data holder: the four legal entities, each with its one brand, with the identifiers, the names and the statuses of their records in the lab seed, and no data recipients. (See Decisions)
- Each endpoint of a brand in the example seed is on an invented host of that brand on the top-level domain `.test`, and the `websiteUri` is the same as in the lab seed. A deployment changes the endpoints to the hosts of the deployment. ([RFC 6761, section 6.2][rfc6761-test]; see Participant model and Decisions)
- The example seed loads into a Register that has no participants with no error.
- A Register that the chart deploys with the example seed has the four preset brands at start-up, with their endpoints. Get Data Holder Brands Summary gives the four brands with their public base URIs ([CDS: Get Data Holder Brands Summary][cds-get-dh-brands-summary]).

## Non-functional requirements

The CDS and the ACCC publish no non-functional requirements for the Register ([CDS: Non-functional Requirements][cds-nfrs] apply to data holders). These requirements are for the lab.

- Propagation: a status change is visible on all APIs within 1 second. Data holders must react to status changes within 5 minutes ([CDS: Metadata Cache Management][cds-metadata-cache]), so the Register must not add delay.
- Performance: on a development workstation, 95% of API calls complete within 200 ms at 50 requests each second. The CDS thresholds for data holder APIs are 1000 to 1500 ms ([CDS: Non-functional Requirements][cds-nfrs]).
- Restart: participant data, statuses, signing keys, revocations and the CRL number stay the same after a restart.
- Telemetry: the Register sends traces, metrics and logs with OpenTelemetry. Telemetry contains no participant contact data and no raw URLs with identifiers. It uses route patterns.
- Secrets: the repository contains no private keys. Keys and certificates are created when the lab is set up. (The ACCC mock publishes its private keys in its repository, [CertificateManagement][accc-mock-certs].)
- Portability: a real data recipient or data holder changes between the mock Register and the production Register with configuration only: base URLs, trust anchors, client certificate, signing keys and identifiers.
- Conformance: a black-box test suite checks each requirement in this document that the CDR defines, through the APIs only ([ADR 0019][adr-0019]). The same suite can run against the ACCC mock Register. Known differences from the ACCC mock are listed in the suite as expected differences.
- Local run: the Register runs with local Docker (Compose), with its dependencies.
- Health: the Register serves a liveness probe (`GET /livez`) and a readiness probe (`GET /readyz`) on a plain HTTP listener that is separate from the CDR API listeners. Judgement: these are the probe paths of the [Kubernetes API server][k8s-health], and a separate listener keeps the probes out of the CDR APIs. The readiness probe also fails while the Register has no current CRL ([Certificate revocation list](#certificate-revocation-list)). A process that runs only the background jobs is ready when it can connect to the database ([ADR 0010][adr-0010]).

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Invalid industry error | 404 `Field/InvalidIndustry` on Brands Summary; 400 `Field/Invalid` on the other endpoints | On Brands Summary, the CDS error list and the OpenAPI specification agree. On the other endpoints they do not agree, and the ACCC clarifies 400 ([Known Issues][accc-known-issues]). |
| Page out of range | 422 `cds-all:Field/InvalidPage` | [CDS: Error Codes][cds-error-codes]. The OpenAPI specification lists no 422 for Get Data Holder Brands, but it does not give another code. |
| Client ID in client assertions | Software product ID only | The CDS will deprecate the brand ID ([CDS: Data Recipients calling the CDR Register][cds-dr-calling-register]). |
| SSA response body | The compact JWT with no JSON quotes, content type `application/json` | The CDS OpenAPI specification gives `application/json` with a string schema ([CDS: Get SSA][cds-get-ssa]), and does not say if the string has JSON quotes. The ACCC implementations use the raw body as the JWT: the test of the ACCC mock Register ([GetSoftwareStatementAssertion_Tests.cs][accc-mock-ssa-test]) and the ACCC mock data recipient ([SsaService.cs][accc-mock-dr-ssa]), which works with production. The body is thus not valid JSON. |
| SSA `scope` claim | The registered scopes of the software product, with no filter, the same for each `{industry}` | Judgement: the CDS does not say how the Register derives the claim. Data holders must ignore the scopes that they do not support ([CDS: Registration Validation][cds-registration-validation]), so a filter adds nothing. The ACCC mock does the same. |
| Sequence of the SSA path checks | Software product, then brand | Judgement: the CDS gives no sequence. The ACCC mock checks the software product first. |
| Optional SSA claims | Included when the Register has the value | Judgement: the CDS marks `legal_entity_id`, `legal_entity_name`, `sector_identifier_uri`, `tos_uri` and `policy_uri` as optional ([CDS: SSA Definition][cds-ssa-definition]). A data holder can show the legal entity name to a consumer. |
| 422 on Get SSA | Not used | The CDS OpenAPI specification gives 422 "SSA validation failed" ([CDS: Get SSA][cds-get-ssa]). The load refuses a software product that does not have a required SSA field, so the condition cannot occur. |
| SSA brand and product errors | 404 `cds-register:Field/InvalidBrand` and `Field/InvalidSoftwareProduct` | The CDS error list says 404 and the OpenAPI specification says 403 "Invalid BrandId". The CDS Known Issues give 404 as the target ([Standardise Register API Error Codes][cds-known-register-errors]), which agrees more with HTTP: the resource does not exist for this caller. |
| Revocation cascade | REVOKED gives REMOVED at once | [CDS: Status Mapping][cds-status-mapping]. |
| Brand statuses on Get Data Holder Brands | ACTIVE, INACTIVE and REMOVED | The schema gives the brand a `status` with all three values, and `updated-since` lets a data recipient poll for changes ([CDS: Get Data Holder Brands][cds-get-dh-brands]). A data recipient that has a brand in its cache can learn that the brand is now INACTIVE or REMOVED only if the API returns it. An ACCC support article says that the API lists brands "active for consumer data sharing" ([CDR Support: Brands in the CDR ecosystem][support-brands]), which is not a specification and is not clear. The ACCC mock does not filter by status. |
| `updated-since` comparison | `lastUpdated` at or after `updated-since` | Judgement: the CDS says only "updated since the specified date-time". With "after", a data recipient that polls with the last time that it saw can miss a change at the same time. A record that is returned again is harmless. The ACCC mock uses "after". |
| Order of Get Data Holder Brands | Brand ID | Judgement: the CDS gives no order. Pagination needs a stable order, and the other lists use the ID order. The ACCC mock sorts by brand name. |
| `ETag` on Get Data Holder Brands | None | The CDS OpenAPI specification gives no `If-None-Match` parameter and no `ETag` header for this API, and `updated-since` gives the incremental read. The ACCC mock adds an `ETag`. |
| Body of the 401 response | None | The CDS OpenAPI specification gives the 401 response of the secure APIs only a `WWW-Authenticate` header, and no body schema ([CDS: Get Data Holder Brands][cds-get-dh-brands]). |
| Sequence of checks on the secure APIs | Bearer token (401), then caller status (403), then the checks of the public APIs | Judgement: the CDS gives no sequence. A caller with no valid token then learns nothing about the API. |
| Caller whose software product does not exist | 403 `Authorisation/AdrStatusNotActive` with the detail REMOVED | Judgement: a token outlives no software product in practice, because the Register never deletes one. The response is the same as for a REMOVED software product, which is the nearest state. |
| Length of a data holder legal entity name | 36 characters at most | The `LegalEntityDetail` schema gives `maxLength: 36` ([CDS: Register API schemas][cds-register-schemas]), which is probably a defect in the CDS. Judgement: the Register could not return a longer name in a valid response, so the load refuses it. |
| Sequence of request checks | Version headers first, then `{industry}`. A response after a successful version negotiation has `x-v`, also when it is an error. | Judgement: the CDS does not give a sequence. The rules for `{industry}` depend on the version: deprecated versions give 404 `Field/InvalidIndustry` and current versions give 400 `Field/Invalid` ([Known Issues][accc-known-issues]). Thus the Register must know the version first. |
| Version header format | Only decimal digits, with a value of 1 or more. A value that is too large for an integer is higher than all supported versions. | The CDS requires a PositiveInteger ([CDS: Invalid Version][cds-error-400-invalid-version]). Judgement: a sign or a space is not a usual integer format, and a very large version is still a positive integer. |
| `x-fapi-interaction-id` | Return it on all APIs | Required on authenticated APIs ([CDS: HTTP Headers][cds-http-headers]). Judgement for public APIs: one behaviour is simpler. |
| ETag format | RFC 9110 | The CDS does not give a format, so the HTTP standard applies. |
| Cache-Control | `no-cache` | Judgement: see Caching, pagination, lists and CORS. |
| Deprecated versions | Not served | A request for a deprecated version gives 406 on the mock. Serve current versions only until the lab is more established. |
| Token refusal for a certificate that is not registered, or a software product that is not ACTIVE | 401 `invalid_client`, the same as for an assertion that is not valid | Judgement: RFC 6749 section 5.2 has no more specific code. One response for each cause also does not tell a caller which software products exist or which are ACTIVE. The Register log gives the cause to the operator. |
| Error codes of the token endpoint | Only the codes of RFC 6749 section 5.2. A server failure gives 500 `server_error`. | Judgement: a client library that follows the RFC knows only these codes. RFC 6749 gives `server_error` only for the authorisation endpoint, but it is the usual code. |
| Values in the OpenID Provider configuration that the Register does not use | `code_challenge_methods_supported` is empty. `response_types_supported` is `token`. `subject_types_supported` is `public`. `id_token_signing_alg_values_supported` is `PS256`. `claims_supported` has the claims of the access token. | The schema requires the fields, and the CDS says that clients can ignore three of them ([CDS: Get OpenId Provider Config][cds-get-oidc]). Judgement: the Register gives the values that are true for it. It supports no PKCE method. |
| Uniqueness of a client assertion `jti` | Unique across all software products, for 1 hour | Judgement: the CDS says only that the Register must reject a `jti` that it accepted before. An assertion is valid for 10 minutes at most, so a longer memory adds nothing. |
| Maximum lifetime of a client assertion | 10 minutes | Judgement: the CDS gives no limit. A short limit keeps the window for a stolen assertion small. |
| Signing keys storage | In the Register database. The Register creates the access token key at its first start. | Judgement: the keys must stay the same after a restart, the repository must contain no keys, and the operator must be able to start a rotation. The database is the only state that the Register keeps. The lab database is not encrypted. |
| Steps of a key change | Two operator actions: publish the next key, then activate it | Judgement: a new key must be in the JWKS before the Register signs with it, and only the operator knows when each party has the new JWKS. Two actions need no timer, and a lab user can keep the state "new key published, old key signs" to test a JWKS cache. |
| When the previous key leaves the JWKS | When the operator publishes the next key of a later key change | Judgement: an SSA is valid for 10 minutes and an access token for 5 minutes, so a later key change is after the last use of the previous key in practice. This needs no timer. If two key changes complete in less than 10 minutes, an SSA of the first key can fail verification. |
| Access token lifetime | 5 minutes | The CDS gives no lifetime. ACCC guidance ([Token Endpoint][accc-token]). |
| Status of a record under a parent that is not active | Not more active than the parent | Judgement: the CDS gives no rule for brands. See Status rules. |
| Uniqueness of identifiers | Unique across all participant records | Judgement: the CDS does not say. See Participant model. |
| Scheme of a URI field | `https` only | `URIString` permits each scheme. But all HTTP calls must use HTTPS ([CDS: Transaction Security][cds-transaction-security]), a FAPI authorisation server must refuse a redirect URI that does not use `https` ([FAPI 1.0 Advanced, section 5.2.2][fapi-as]), and a sector identifier URI must use `https` ([OIDC Core, section 8.1][oidc-pairwise]). Judgement: a data holder cannot use an endpoint or a registration with a different scheme, so the Register refuses the value when the operator gives it, not when a participant uses it. A logo, a client URI or a policy is a web page for a consumer, and one rule for each URI field is simpler. |
| Length of a URI field | 1000 characters at most, also for the URIs of a software product | The Register schemas give `maxLength: 1000` to each logo URI and each endpoint ([CDS: Register API schemas][cds-register-schemas]). Judgement: the RegistrationProperties schema gives no limit to the other URIs of a software product. One limit for each URI field is simpler, and the URIs of the seed data are much shorter. |
| Data recipient `lastUpdated` | The time of the last change to the data recipient, its brands or its software products | Judgement: the CDS says only "the date/time that the Legal Entity was last updated". A change to a brand or a software product is a change to the legal entity record that the API returns, and a data holder that polls the API then sees each change. |
| Industry on Get Data Holder Statuses | Data holders with a brand in the industry, in any brand status | The CDS is not clear, and the ACCC mock filters by industry ([RegisterStatusRepository.cs][accc-mock-dh-statuses]). Judgement for the brand status: the brands of a REMOVED data holder are REMOVED, and data recipients of that industry must see that the data holder is REMOVED. |
| Brand statuses on Get Data Holder Brands Summary | ACTIVE and INACTIVE, not REMOVED | Judgement: the CDS and the ACCC do not say, and the summary has no status field. Brands of data holders that left the CDR are not in the production list. A REMOVED brand has no endpoints to discover, and an INACTIVE brand can be only temporarily unavailable. |
| Headers that a CORS preflight permits | The requested headers | Judgement: the CDS requires that CORS protections are disabled on unauthenticated APIs ([CDS: CORS][cds-cors]). An allowlist would block a page that sends a harmless header and protects nothing. The production Register also returns the requested headers. |
| Brand with no `dataHolderBrandId` | Needs only its public details | Judgement: see Participant model. |
| Software products at a reinstatement | No change | The CDS cascades a data recipient status to its software products only "for any status other than active" ([CDS: Status Mapping][cds-status-mapping]). |
| Operator interface errors | RFC 9457 problem details | Judgement: the CDS error codes are for the CDR APIs. RFC 9457 is the HTTP standard for error details. |
| Operator authentication | One bearer token from the configuration | Judgement: the production interface is not public. A token is the simplest authentication that keeps the interface closed. mTLS needs the mock PKI. |
| Registered client certificates | The Register keeps only the SHA-256 thumbprint. The operator interface does not check the issuer. | Judgement: the token endpoint compares the full certificate through its thumbprint, which is also the RFC 8705 `x5t#S256` value. The mTLS listener accepts only certificates from the mock CDR CA, so a second issuer check adds nothing. |
| Client certificate for a REMOVED software product | Refused | Judgement: REMOVED is final ([CDS: Participant Statuses][cds-participant-statuses]), so the certificate can never be used. An INACTIVE software product can become ACTIVE, so it can get a certificate. |
| Subject of an issued client certificate | The CN must be the software product name, and the O must be the brand name, with the same characters. The certificate has the subject of the CSR. | The CDS CSR profile gives both values, and says that the values "will be validated to be correct" ([CDS: Certificate Signing Request Profile][cds-csr-profile]). Only the Register knows the names. Judgement: an exact comparison is the simplest rule, and a certificate has the subject of its CSR ([PKI DESIGN.md][pki-design]). |
| Name of more than 64 characters | The issue is refused. The `detail` gives the limit. | RFC 5280 limits a CN and an O to 64 characters ([RFC 5280, Appendix A][rfc5280-bounds]). The CDS permits a software product name and a brand name of 200 characters ([CDS: Register API schemas][cds-register-schemas]), and it does not say what to do. Judgement: a certificate with a shorter name does not follow the CSR profile, so no correct certificate is possible. |
| Registration of an issued client certificate | In the same action as the issue, with one event | In production, the ACCC issues a client certificate for one software product, and the Register accepts it ([ACCC Developer Portal: Token Endpoint][accc-token]). Two lab actions would let a certificate exist that the token endpoint refuses. One command that causes one fact is one event ([ADR 0021][adr-0021]). |
| Response of an issue | 200, with the certificate and the intermediate in PEM (`application/pem-certificate-chain`). No `Location`, and no copy in the Register. | Judgement: TLS servers and clients load the PEM chain format of [RFC 8555, section 9.1][rfc8555-chain], and a TLS server sends its chain with no root ([RFC 8446, section 4.4.2][rfc8446-cert]). The Register keeps only thumbprints (see the row "Registered client certificates"), so the issue makes no resource that a `GET` can give, and 201 does not apply ([RFC 9110, section 15.3.2][rfc9110-201]). |
| Channel of the CDR CA bundle | The operator interface | Judgement: a new component gets its trust anchors from its operator, as a participant downloads the ACCC bundles ([ACCC Developer Portal: Certificate Management][accc-cert-mgmt]). The operator uses the operator interface for all participant changes. A TLS listener needs a trust anchor first, so it cannot give the first one. |
| Format of the CDR CA bundle | The intermediate, then the root, in PEM (`application/pem-certificate-chain`) | Judgement: the order of a chain in [RFC 8555, section 9.1][rfc8555-chain], with no end-entity certificate. It is the same file that the lab gives to its relying parties. |
| Owner of a server certificate | The data holder brand | The CDS gives server certificates to data holders, for the endpoints that the data holder registers "against each of their brands" ([CDS: Certificate Management][cds-cert-mgmt], [CDS: Participant Endpoints][cds-participant-endpoints]). The ACCC says: "Server certificates are issued at a brand level to data holders" ([CDR Support: Brands in the CDR ecosystem][support-brands]), and the Participant Portal requests a server certificate for a brand ([ACCC Participant Portal User Guide, section 10.6.1][accc-portal-guide]). A data holder legal entity has no endpoints. |
| Server certificate for a data recipient | None | The CDS says that a server certificate is "Not currently required by Data Recipients" ([CDS: Certificate Management][cds-cert-mgmt]). Each data recipient endpoint (`RevocationUri`, `RecipientBaseUri`, `JwksUri`) is TLS, with a certificate from a public CA ([CDS: Participant Endpoints][cds-participant-endpoints]). A data holder authenticates to the CDR Arrangement Revocation endpoint of a data recipient with client authentication, not with mTLS ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]). The Register is not a public CA ([ADR 0022][adr-0022]). |
| Server certificate for a brand with no `dataHolderBrandId` | Refused (409) | Judgement: the sources do not say. Such a brand shares only product data ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary]), so it has no mTLS endpoint. Its identifiers do not change, so a certificate can never secure a CDR endpoint of the brand. 409 is the code of the other refusals that the state of a record causes. |
| O of a server CSR | The brand name or the legal entity name, with the same characters. Any other value is refused. | The CDS gives these two values, and says that the values "will be validated to be correct" ([CDS: Certificate Signing Request Profile][cds-csr-profile]). The subject DN "shall accurately reflect the organisation" ([ACCC CP, section 3.1.2][accc-cp]). The Participant Portal only warns, and the ACCC can then refuse the request at its review ([ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide]). Judgement: the lab has no later review, so the Register refuses at once. An exact comparison is the rule of a client certificate. |
| CN and DNS names of a server CSR, and the endpoints of the brand | Not compared | The CDS lets the data holder decide how it puts its endpoints on domains ([CDS: Certificate Management][cds-cert-mgmt]). The Register does not know each mTLS host: the authorisation server endpoints are in the OpenID Provider configuration of the data holder, not on the Register ([CDS: Participant Endpoints][cds-participant-endpoints]). The Participant Portal checks only that the CN is a host name ([ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide]), and the ACCC does not verify the ownership of a domain name ([ACCC CPS, section 3.1.6][accc-cps]). Judgement: a lab component also needs names that are not in its endpoints, for example `localhost` for access from the host. |
| Record of a server certificate issue | One `ServerCertificateIssued` event of the data holder, with the brand and the thumbprint. It sets the `lastUpdated` of the brands of the data holder. | One command that causes one fact is one event ([ADR 0021][adr-0021]). Judgement: the same fields as `ClientCertificateIssued`. The ACCC keeps a reference to the server certificates of a brand on the Register ([CDR Support: Brands in the CDR ecosystem][support-brands]), so the issue is a change to the data holder. The `lastUpdated` of a brand is the time of the last change to its data holder ([Caching, pagination, lists and CORS](#caching-pagination-lists-and-cors)), and an exception for one event type adds a rule for no gain. |
| Load into a Register that has participants | Refused | Judgement: a merge of two sets can break the status rules or reuse a REMOVED software product identifier. A reset is simple in the lab. |
| TLS 1.2 cipher suites | Only the two ECDHE-RSA AES-GCM suites | The CDS permits a suite only when FAPI 1.0 Advanced section 8.5 and BCP 195 both permit it. The production public API (`api.cdr.gov.au`) also accepts `ECDHE-RSA-AES128-SHA256` (observed 2026-10-03), which the CDS does not permit. A client that works with the lab also works with production. |
| Revocation status not known | Refuse the client certificate | Judgement: the CDS requires the check ([CDS: Certificate Validation][cds-cert-validation]) and does not say what to do when it cannot be done. To accept the certificate would let a revoked client connect when the CRL server is down. The status is not known only when the Register has no current CRL. |
| Age of a CRL that a relying party keeps | 1 minute at most. The Register uses its newest CRL for each connection. | Judgement: the CDS gives no time. A CRL is valid for 7 days, but a lab test must see a revocation quickly, and a new download for each connection would make the CRL server a bottleneck. The Register signs the CRL, so it reads its newest CRL with no download (see the row "A revocation at the Register"). |
| Identification of the certificate to revoke | The PEM certificate, or the PEM chain of an issue, in the body | The CDS says nothing. A request "shall identify the certificate to be revoked" ([ACCC CP, section 4.9.3][accc-cp]), and the Participant Portal revokes a certificate from its list of the valid certificates of the participant ([ACCC Participant Portal User Guide, sections 10.6.4 and 10.7.3][accc-portal-guide]). The Register keeps no certificates, only thumbprints, and it has no record of the certificates of the lab set-up, so it cannot get a certificate or its serial number from a thumbprint. A serial number alone does not let the Register check that the intermediate issued the certificate. Judgement: the certificate gives the serial number and a signature that the Register can check, for each certificate of the intermediate. The operator keeps each certificate, because a component cannot use one that it does not have. The `pki revoke` command took the same input. |
| Form of the revocation request | `POST /v1/cdr-certificates/revoke`, with a JSON body that has `certificate` and `reason` | Judgement: the request has two values, and the operator interface already takes JSON bodies. An action is a `POST` on a path that ends with the action, as the status actions are. The path is not under a participant, because a certificate of the lab set-up has no participant in the Register. |
| Response of a revocation | 204 | Judgement: the Register makes no resource that a `GET` gives. The CRL gives the result. |
| PEM chain in a revocation | Accepted. The Register revokes the first certificate. | Judgement: an issue gives the certificate and then the intermediate ([RFC 8555, section 9.1][rfc8555-chain]), and the operator keeps that file. If the operator gives the chain in the wrong order, the first certificate is the intermediate, which the intermediate did not sign, so the Register refuses it (400). |
| Revocation of a certificate that is already revoked | 409, no change, also with a different reason | The [PKI DESIGN.md][pki-design-revocation] permits one revocation of a certificate. The CPS says that the ACCC changes a reason to `keyCompromise` when it finds a key compromise ([ACCC CPS, section 4.9.12][accc-cps]). Judgement: the lab has no process that finds a compromise, and a second revocation that gives 204 would hide that the reason did not change. 409 is the code of the other refusals that the state of a record causes ([RFC 9110, section 15.5.10][rfc9110-409]). |
| Certificate that the intermediate did not sign | 400 | Judgement: the CRL of the intermediate can list only its own certificates ([RFC 5280, section 5][rfc5280-crl]). The fault is in the body, as for a CSR that breaks the profile. |
| Validity period and participant of a revoked certificate | Not checked | Judgement: a relying party refuses an expired certificate also with no CRL entry, so a revocation of it does no harm. The constraints of the project apply a revocation to each certificate of the intermediate, also to the certificates of the lab set-up, which have no participant in the Register. |
| Revocation and the registered client certificates | Separate. A revocation does not remove a registration. | Judgement: the mTLS listeners refuse a revoked certificate in the TLS handshake, before the token endpoint, so a registration of a revoked certificate gives no access. One command that causes one fact is one event ([ADR 0021][adr-0021]). A revocation is a fact of the CA, and a certificate of the lab set-up has no software product. The operator removes the registration with its own action. |
| Record of a revocation | A record of the CA, not an event of a participant. The history of a participant does not show it. | [ADR 0003][adr-0003] gives two aggregate types, a data recipient and a data holder, and a revocation names neither. Judgement: the key sets are records of the Register that are not event sourced, and the revocations are the same. The CRL gives each revocation with its time and its reason. |
| Times of a new CRL | At the start, at each revocation, and no later than 24 hours after the last CRL | The ACCC issues a CRL at least once every 24 hours, also with no change ([ACCC CP, section 4.9.7, Table 1][accc-cp]; [ACCC CPS, section 2.3][accc-cps]). Judgement: a CRL at each revocation lets a test see the change at once ([PKI DESIGN.md][pki-design-revocation]). A CRL at each start makes sure that the current CRL has the signature of the intermediate key that the Register reads. With a validity of 7 days, the CRL stays current for 6 days after the Register stops. |
| Test of the CRL schedule | Register tests with a controlled clock. The conformance suite checks only the current CRL. | Judgement: a black-box test of a 24-hour schedule needs a wait of 24 hours. A test with a controlled clock proves the same rule with no wait. The conformance suite can check that the current CRL is current and signed, with no wait. |
| CRL listener | A separate plain HTTP listener, with no authentication and no check of the `Host` header | Relying parties get the CRL from the HTTP URL in the certificate ([RFC 5280, section 4.2.1.13][rfc5280-crldp]), and the CRL distribution point of the production certificates is an `http` URL (observed: "CDR Intermediate CA 2025", [PKI DESIGN.md: Certificate profile][pki-design-profile]). A CRL is signed, so it needs no TLS. A TLS listener would need a revocation check of its own certificate. Judgement: the CRL is public data, so the operator token does not apply, and a DNS rebinding attack gets only public data. The probe listener is for the deployment, not for participants, and [ADR 0022][adr-0022] gives the CRL its own listener. |
| A revocation at the Register | Effective at once at the Register. Within 1 minute at other relying parties. | Judgement: the Register signs the CRL, so it can use its new CRL at once, with no wait for its copy of the CRL to expire. A conformance test can then check the revocation with no wait. The other relying parties keep a CRL for up to 1 minute (see the row "Age of a CRL that a relying party keeps"). |
| Renewal of the CRL with no revocation | A periodic background job each 10 minutes signs a new CRL when the newest CRL is 23 hours old or older | The 24 hours of the [ACCC CP, section 4.9.7, Table 1][accc-cp] less the interval of the job. A background job follows [ADR 0010][adr-0010]. Judgement: the job compares the time of day with the `thisUpdate` of the newest CRL, so the first run after a stop or a suspend of the host signs a new CRL. A timer of 24 hours does not count the time of a suspend. An interval of 10 minutes costs one database read each run. |
| Response of the CRL URL with no current CRL | 503 | Judgement: the Register cannot give the CRL for a time, and the readiness probe also fails ([RFC 9110, section 15.6.4][rfc9110-503]). A relying party then has no current CRL and refuses the certificate ([Transport security](#transport-security)). |
| Readiness with no current CRL | Not ready | Judgement: with no current CRL, each relying party refuses each CDR certificate ([Transport security](#transport-security)), so the Register cannot serve the secure APIs. |
| TLS on the health probes and the operator interface | No, plain HTTP | Judgement: they are not CDR APIs. The probes are for the container platform, and the operator interface is lab-only. |
| Health probe paths | `/livez` and `/readyz` | Judgement: the CDS and the ACCC say nothing. Kubernetes uses these paths ([Kubernetes API health endpoints][k8s-health]); `/healthz` is deprecated there. |
| Console listener | A separate listener, off by default, turned on by a setting | Judgement: [ADR 0014][adr-0014] gives a listener for each console. A production Register must not get a console without an explicit setting. |
| TLS on the console | No, plain HTTP | Judgement: the same reason as for the operator interface. The console sets no cookie. Browsers send `Sec-Fetch-Site` to `localhost` over HTTP ([Fetch Metadata][fetch-metadata]), so the cross-origin check works. |
| Confirmation of a status change | A dialog that names the cascade, and says when a change cannot be reversed | [WCAG 2.2, SC 3.3.4][wcag-3-3-4] asks for a check or a confirmation for changes to data. Judgement: a cascade changes records that the operator did not select. |
| Times in the console | The time zone of the browser, with the offset from UTC | [ADR 0012][adr-0012] permits a different time zone where a person sees the time. Judgement: an operator compares a time in the console with their own clock, and the offset lets them compare it with a time in UTC, for example in the logs. |
| Pagination of the console API | Page and page size in the request, the total in the response | Judgement: the CDS rules are for the CDR APIs, and a total is simpler for the UI. |
| Console CSP | `script-src 'self'` with no nonce and no hash, and Trusted Types required | The UI build has no inline script. A hash or nonce policy is for pages that have inline scripts ([web.dev: Strict CSP][strict-csp]). Trusted Types are in all three browser engines ([MDN compatibility data][bcd-trusted-types]). |
| Creation and change of participants in the console | In scope: the console creates, adds and changes each record that the operator interface does, and does the status actions of brands and data holders, with the same use cases | Judgement: a lab user who tests their own data recipient software must register it first, and an HTTP client with the operator token is a barrier to the most frequent set-up tasks. The console API and the operator interface stay separate adapters on the same use cases ([ADR 0014][adr-0014]). |
| Field errors in the console API | Each field error has a JSON Pointer to the field of the request body and a `detail`, in the extension member `errors` of the problem details. The UI shows each one at its field. | The format of the example of [RFC 9457, section 3][rfc9457-3], with a JSON Pointer ([RFC 6901][rfc6901-6]). A field error must show at its field ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]), and the UI has no copy of the rules, so only the Register can name the field. The domain rules already name the field of each value that they refuse. |
| Problem type of field errors | A type URI that the console API OpenAPI document defines, not `about:blank` | [RFC 9457, section 3.2][rfc9457-3-2]: a problem type definition gives its extension members, and `about:blank` means only the status code. The type also lets the UI find the field errors without a check of each member. |
| Status codes of field errors | The status code of the operator interface: 400 for a value that does not obey the participant model, 409 for an identifier in use or a status that breaks a status rule | Judgement: one rule for the status codes in the two interfaces (see the row "Refusal of a body"). The field errors are in the body, so the UI does not need a different code. |
| Identifiers of a create or an add in the console | The form shows a new random UUID for each new identifier, in a field that the operator can change. A data holder brand gets a `dataHolderBrandId`, and no `interimId`. | Judgement: the Register issues GUIDs in production ([CDS: Register API schemas][cds-register-schemas]), so a UUID is the usual value and the operator types nothing. A lab service gets its identifiers in its configuration before it starts, for example the brand ID of the mock bank, so the operator must be able to give a known value. A brand with a `dataHolderBrandId` is the usual brand. The Register refuses an identifier that a different record has (409). |
| Status of a new record in the console | A field, with ACTIVE selected first, and the status rules apply | Judgement: the operator interface permits each status in a create or an add (see the row "Status in the body of a create or an add"). With a fixed status of ACTIVE, an add under a parent that is not ACTIVE always fails, for example a software product of an INACTIVE brand. The field needs no rule in the UI: the Register refuses a status that breaks a rule, at the field of the status. |
| Child records in a create in the console | A create makes the legal entity only. The operator adds each brand and software product after the create. | The participant model does not require a brand for a legal entity, and the operator interface creates a legal entity with no brands. Judgement: each form then has the fields of one record, the add forms do the work of the child records, and each field error has a short pointer. |
| Values of a change form | The current values, and each field of the change, with the version that the form loaded. An add also sends the version. | Judgement: a change in the operator interface replaces each field ([RFC 9110, section 9.3.4][rfc9110-put]), so the form must send the fields that the operator did not change. The version stops a change that is based on an old view (see the row "Action with a stale view"). An add changes the legal entity, so it has the same check. |
| Path and method of a change in the console API | `POST` on the path of the record with `/change` at the end, with the version and the fields in the JSON body | Judgement: the same form as each other console action (see the row "Path of a certificate removal in the console API"). A `PUT` replaces the representation that a `GET` of the same path gives ([RFC 9110, section 9.3.4][rfc9110-put]), and the `GET` of a participant gives its detail with its child records, its version and its actions, not the fields of the change. A brand and a software product have no `GET`. |
| Response of a create or an add in the console API | 201 with a `Location` header and no body | [RFC 9110, section 9.3.3][rfc9110-post]: a `POST` that makes a resource sends 201 with a `Location`. Judgement: the same as the operator interface. The UI knows the identifiers, because it sent them. |
| Forms in the console | A page with its own URL for each create, add and change. A dialog only for the confirmation of an action. | Judgement: the add form of a software product has 14 fields and the add form of a data holder brand has 15, which is too much for a dialog. A page has its own URL and history entry, and Back leaves it as the operator expects. |
| Confirmation of a create, an add or a change | None | [WCAG 2.2, SC 3.3.4][wcag-3-3-4] is met when the Register checks the values and the operator can correct them, or when the change is reversible. A change is reversible by a second change. A create and an add change no other record. |
| Lists of values in a form | A text area with one value on each line for a list that the operator writes. A group of checkboxes for a fixed set. A native select for one value of a fixed set. | Judgement: a text area is one control with one label and one error, the operator can paste a list, and the keyboard and the screen reader read it line by line. A list of text fields needs add and remove buttons and a focus rule for each, and gives no better access. The checkboxes and the select show each value of the fixed set, with the keyboard support of the browser. |
| Error summary in a form | An error summary at the top of the form after each refusal, with the focus on it | WCAG 2.2 SC 3.3.1 needs only the error at the field ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]). Judgement: a software product form is longer than one screen, so an error at a field far down the page is not seen. The [GOV.UK Design System: Error summary][govuk-error-summary] shows a summary for each error, with a link to each field, and moves the focus to it. The summary also has the refusals that name no field, so the form has one place for errors. |
| Values after a refusal | The form keeps the values of the operator and the version that it loaded. It does not get the current values by itself. | Judgement: the operator does not lose their work. A form that takes a new version by itself would replace the change of a different person or tab with no decision of the operator, which the version check stops. |
| White space in a form | The console removes it at the start and at the end of each value and of each line, and ignores an empty line | Judgement: a pasted value often has white space or an empty last line. This is not a rule of the Register: a value with a space at the end is never intended. The Register still checks each value. |
| Change and add actions in the console | Shown for each record in each status | Judgement: the domain permits a change and an add in each status, and refuses only a value or a status that breaks a rule. A list of permitted changes in the console API would always give each change. |
| Cancel of a form | A link back, with no confirmation and no warning about values that the operator did not send | Judgement: a lab form is quick to fill again, and a warning before each navigation adds a rule for each route. |
| Optional fields in a form | The label says "(optional)". A required field has no mark. | [GOV.UK Design System: Question pages][govuk-question-pages]. The console API schema gives which fields are optional, so the UI copies no rule. |
| Page size of the console API | 25 by default, 100 at most | Judgement: 25 rows fit a desktop screen in the compact density, and a maximum keeps each response small. The CDS default page size is also 25. |
| Page after the last page in the console API | 200 with no records and the total | Judgement: the CDS 422 is for the CDR APIs. After a filter change, the UI can then show "no records" and go back to page 1. |
| Search of the data recipient list | A part of the legal entity name, with no difference for upper and lower case. `%` and `_` are text. | Judgement: the operator knows a part of the name. A pattern syntax would surprise the operator. |
| Sort of the data recipient list | Name (the default), accreditation number or status, in the byte order of the text, or the number of brands or of software products. Then the legal entity ID. The name sort has no difference for case. | Judgement: the byte order is the same in each database locale, and the legal entity ID keeps each page stable. The operator asked for the count sort. The counts are columns of the read model, so a count sort needs no calculation for each row ([ADR 0011][adr-0011]). |
| Counts in the data recipient list | The brands and software products in each status | Judgement: the list shows the size of each data recipient. The detail view shows each status. |
| Brands in the data holder list | The number of brands and the industries of the brands, not each brand | Judgement: one row for each legal entity stays short, also for a data holder with many brands. The detail shows each brand with its status. |
| Search of the data holder list | A part of the legal entity name or of a brand name, with the rules of the data recipient search | Judgement: the CDR APIs and the lab services name a data holder by its brand (for example, the mock bank is a brand), so the operator often knows only the brand name. |
| Sort of the data holder list | Name (the default), status or the number of brands, with the rules of the data recipient sort | Judgement: the same rules as the data recipient list. A data holder has no accreditation number and no software products. |
| Pages for the detail of a participant | None: one response | Judgement: the records of one legal entity are one aggregate and one stream, and a legal entity in the seed data has 13 software products or fewer. |
| Fields of the detail of a participant | The names of the CDS Register schemas, with `lastUpdated` for the time of the last change and `clientCertificateThumbprints` for the registered client certificates. No optional field with no value, no `null`. | Judgement: the operator sees the same names in the seed file, the CDR APIs and the console. A field that is not present is simpler in the closed schemas and the generated types than `null`. The thumbprints are not a CDS field, so the name says what the values are. |
| Fields of an event in the history | Only the fields of the table in Functions. No payload and no metadata. | The console serves only data that a domain type gives ([ADR 0011][adr-0011]), and each console API schema is closed ([ADR 0017][adr-0017]), so a raw payload cannot be in a response. Judgement: the import payload is the full record, which the detail shows. |
| Events of one action in the history | Each event is one entry, in the order of the stream version. No group and no cause. | Judgement: the events have no field that names the action that caused them. A suspension and its cascade have the same time, and the stream version gives their order. |
| Source of the history | A read model with one row for each event, not a read of the events | [ADR 0003][adr-0003]: each read API has a read model. The read model has only the fields of the table, so a payload cannot get to the console API. |
| Name of a software product in the history | The name at the time of the event | Judgement: a history shows the record as it was. The events give only the software product ID, so the name comes from the state of the stream at that event. |
| Schema of an event in the history | One closed schema for each group of event types with the same fields, with `eventType` as the discriminator | [ADR 0017][adr-0017]: the check of a response then refuses a field that the table does not give for its event type. One schema with optional fields would permit, for example, a thumbprint on a suspension. |
| Detail and history | Two pages with two URLs | Judgement: the history has its own page in the URL, as a list does. |
| Logos in the console | The logo URI as text, not the image | The UI loads no file from another origin ([OWASP ASVS 5.0, V3.6.1][asvs-v3]), and the console CSP permits images only from `'self'` and `data:`. |
| URIs of a participant in the console | Text with a copy button, not a link | Judgement: the URIs are for software (redirect URIs, API base URIs, JWKS URIs), and the lab host names of the services do not always resolve in the browser of the operator. The operator copies a URI into a different tool. |
| Times of the keys | The time at which the Register created each key | Judgement: the Register keeps no other time for a key, and the creation time shows the sequence of a key change. |
| Key material in the console | None: no private key and no public key | Judgement: the JWKS endpoints give the public keys. The console needs only the key ID to compare with a `kid`. |
| Order of the keys in a key set in the console | The key that signs, then the next key, then the previous key | Judgement: the same order as the JWKS. The operator looks first for the key that signs. |
| Path of a certificate removal in the operator interface | `DELETE` on the thumbprint under `certificates` | Judgement: the Register keeps only the thumbprint, so the thumbprint identifies the certificate. A `DELETE` on the item is the usual HTTP form for a removal ([RFC 9110, section 9.3.5][rfc9110-delete]). |
| Removal of a thumbprint that is not registered | 204, no change | Judgement: the same as a second registration, so that a script can send a removal again with no error. `DELETE` is idempotent ([RFC 9110, section 9.2.2][rfc9110-idempotent]). |
| Removal of a certificate of a REMOVED software product | Permitted | Judgement: a removal only takes trust away, so it cannot break a status rule. The registration is refused for REMOVED because the certificate could never be used. |
| When a certificate removal takes effect | At the next token request. Access tokens that exist stay valid until they expire. | The token endpoint checks the registered certificates when it issues a token ([ACCC Developer Portal: Token Endpoint][accc-token]). Judgement: an access token is valid for 5 minutes, and a check of the certificates on each secure API call adds a read for a short window. |
| Actions that need a confirmation in the console | Each action except the registration of a client certificate and the issue of a certificate. The confirmation of a certificate removal does not say that the removal cannot be reversed. | [WCAG 2.2, SC 3.3.4][wcag-3-3-4] accepts a change that is reversible, or a check of the values that the operator can correct. The operator can remove a registered certificate. The Register checks each rule of a CSR, and the operator can reverse an issue: remove the registration of a client certificate, or revoke the certificate with the chain that the issue gave. A revocation cannot be reversed. The Register does not keep the PEM file, so the operator can reverse a removal only with a copy of it: the confirmation says this. A key change is not reversible, and a status change can change records that the operator did not select (see the row "Confirmation of a status change"). |
| Path of a certificate removal in the console API | `POST` on `certificates/{thumbprint}/remove`, with the version in the JSON body | Judgement: each console action is a `POST` with the version in a JSON body (see the row "Action with a stale view"). The content of a `DELETE` has no defined meaning ([RFC 9110, section 9.3.5][rfc9110-delete]), so the version cannot go in the body of a `DELETE`. |
| Permitted certificate actions in the console API | Each software product gives `clientCertificateActions`: `register` and `issue` when the rules permit them, and `remove`, which the rules permit in each status. Each data holder brand gives `serverCertificateActions`: `issue` when the rules permit it. A revocation and the CDR CA bundle have no permitted actions. | Judgement: the same reason as the row "Cascade in the confirmation". A certificate action does not change a status, so it is not in the status actions. The rules permit a revocation and the bundle for each certificate of the intermediate, so the console API has nothing to give for them. |
| Cascade in the confirmation | The console API gives the permitted actions and the records that each one changes. The Register finds them with the domain rules on the current state, with no change. | Judgement: the rules stay in one place ([ADR 0013][adr-0013]). A separate preview request for each action gives the same data, with one more request and one more endpoint for each action. |
| Action with a stale view | The request gives the version that the operator saw. A different version gives 409, with no change. | Judgement: the event store appends only at the expected version ([ADR 0003][adr-0003]), so the check costs little. Two browser tabs can show different versions. `If-Match` with 412 ([RFC 9110, section 13.1.1][rfc9110-if-match]) is the HTTP form, but it adds two status codes that the operator interface does not use, and a `POST` to an action has no representation for the entity tag. |
| Version of a key set | An integer that is 1 when the Register creates the key set, and that each publish and each activation increments by 1, also through the operator interface. A step that the rules refuse does not change it. | Judgement: the key sets are not event sourced, so they have no stream version. The same type as the version of a participant gives one request body for each action. A value from the key IDs and their states also changes at each step, but it is a different type, and the check of a stale view then needs its own lock. |
| A key step that changes no key that exists | The confirmation still says that the step cannot be reversed. | Judgement: the operator interface has no step that removes a published key or activates a previous key again, so a publish with no previous key is also a step that cannot be reversed. |
| Result of a console action | 204, then the UI gets each view again | Judgement: the same response as the operator interface. One rule for each action is simpler than a different body for each action. |
| A check of the `Host` header on the console | An allowlist from a required setting, with no default. The lab gives `127.0.0.1:8084` and `localhost:8084`. | The console has no credential. With DNS rebinding, a web page can send a same-origin request to the console port, and `CrossOriginProtection` does not refuse it, because the `Origin` and the `Host` agree ([Vite advisory GHSA-vg6x-rcgg-rjx6][vite-dns-rebinding]). Thus any web page that the operator opens could change the Register. The check is small. Judgement: a default would be a value of one deployment, so the setting is required. 403 is the same status as the refusal of a cross-origin change. |
| Status transitions of data holders and brands | A data holder legal entity: ACTIVE to REMOVED, and REMOVED is final. A brand: the software product transitions. | The CDS gives no state diagram for them ([CDS: Participant Statuses][cds-participant-statuses]). Judgement: the CDS diagrams make REMOVED final for a software product, so REMOVED has one meaning for each record. A data holder legal entity has no INACTIVE status, so its only change is to REMOVED. A REMOVED identifier is then never in use again, the same as for a software product ([CDR Support: Removed software product][support-removed-sp]). |
| Cascade of a brand status | INACTIVE makes the ACTIVE software products INACTIVE. REMOVED makes each software product REMOVED. ACTIVE changes no software product. A REMOVED data holder makes each brand REMOVED. | Judgement: the CDS cascades only a data recipient status ([CDS: Status Mapping][cds-status-mapping]). The same cascade for a brand keeps the rule that a record is not more active than its parent, and a status change of the brand alone would break that rule. |
| A removal in the operator interface | A change to the status REMOVED. The Register deletes no record. | Judgement: the Register never issues a REMOVED software product identifier again ([CDR Support: Removed software product][support-removed-sp]), and the events of each participant stay ([ADR 0003][adr-0003]). The same rule for each identifier is simpler. |
| Removal of a data recipient legal entity | Its revocation or its surrender, which exist | The CDS gives a data recipient legal entity no REMOVED status ([CDS: Participant Statuses][cds-participant-statuses]). |
| Body of a create or an add | One record in the format of the load file, with its child records | Judgement: one format for the load and the operator interface. The operator can copy a record from the seed file, and the Register checks it with the same rules. |
| Identifiers in a create or an add | The operator gives them, with the rules of the participant model | The Register issues the identifiers in production ([CDS: Register API schemas][cds-register-schemas]). Judgement: the lab services get their identifiers in their configuration before they start, as with the load file. A GUID is the usual value. |
| Status in the body of a create or an add | Permitted, with the status rules | Judgement: the same as the load. A lab user can make a record in each status to test a data holder or a data recipient. |
| Response of a create or an add | 201 with a `Location` header and no body | [RFC 9110, section 15.3.2][rfc9110-201]. Judgement: the operator gives the identifiers, so a body adds nothing. The `Location` is also the path of the change of the record. |
| Form of a change | `PUT` with each field of the record, except the identifiers, the status, the child records and the client certificates | [RFC 9110, section 9.3.4][rfc9110-put]: the body replaces the state of the record. Judgement: each excluded field has its own action with its own rules. A merge patch ([RFC 7396][rfc7396]) needs rules for `null` and for lists, and is not simpler for a record of a few fields. |
| Identifiers in a change | They do not change | Judgement: participants keep the identifiers in their caches and their registrations, and the events name each record by its identifier. A new record is the way to get a new identifier. |
| Refusal of a body | 400 when the body does not obey the participant model. 409 when the body obeys it but breaks a status rule, or has an identifier that a different record has. | Judgement: a body that is not valid alone is a client error (400). A body that is valid alone but does not agree with the current state of the Register is a conflict with the target resource ([RFC 9110, section 15.5.10][rfc9110-409]). The status rules already give 409. |
| Uniqueness of an identifier in a create or an add | Unique across all records, also REMOVED records and records of the other kind of participant | The participant model. The Register never issues a REMOVED software product identifier again ([CDR Support: Removed software product][support-removed-sp]). |
| Size of a request body in the operator interface | 1 MiB at most | Judgement: a legal entity in the seed data is less than 20 KiB. A limit stops a body that fills the memory. |
| Brand in the history | `brandId` and `brandName`. The brand ID of a data holder brand is its `dataHolderBrandId`, or its `interimId` when it has no `dataHolderBrandId`. | Judgement: one pair of fields for each kind of brand, with the same identifier as the path of the brand in the operator interface. |
| Event of a create | `DataRecipientCreated` and `DataHolderCreated`, not the event of the load | Judgement: an event names the fact. The history then shows if the load or the operator made the record. |
| A check of the `Host` header on the operator interface | An allowlist from a required setting, with no default, as for the console. The lab gives the service name for the clients on the compose network, and `127.0.0.1` and `localhost` with the port on the host. | The operator token is public in `compose.yaml`, and the port binding to `127.0.0.1` does not stop DNS rebinding. With DNS rebinding, a web page that the operator opens can send a same-origin request with the token ([Vite advisory GHSA-vg6x-rcgg-rjx6][vite-dns-rebinding]). Judgement: the same check as the console, so that one rule protects each plain HTTP interface of the lab. A default would be a value of one deployment, so the setting is required. |
| Names of the key sets and states in the console API | The key sets are `access-token` and `ssa`, the same as in the paths of the operator interface. The states are `signing`, `next` and `previous`. | Judgement: the CDS does not name them. One name for each key set in both interfaces lets a console action use the value with no mapping. |
| Place of the certificate actions in the console | "Issue certificate" in the action row of the software product card and of the data holder brand card. The CDR CA bundle and "Revoke certificate" on a page of their own, "CDR certificates", in the side navigation. | Judgement: an issue is an action on one record, and the design system puts an action in the action row of its record ([Design system DESIGN.md][ui-design]). The bundle and a revocation are not of a participant: a revocation can be of a certificate of the lab set-up, which has no participant in the Register (see the row "Form of the revocation request"). Thus they get their own page, as the key sets do. A revocation from a registered thumbprint is not possible: the Register keeps no certificate, and a revocation needs the certificate (see the row "Identification of the certificate to revoke"). |
| Result of an issue in the console | The dialog shows the result: the thumbprint, the PEM chain in a text area that the operator cannot change, "Save as file" and "Close". No toast. | The response is the only copy of the certificate (see the row "Response of an issue"). A toast closes by itself, and the design system shows only a result that the view also shows in a toast (Design system DESIGN.md, row "Errors and toasts"; [WCAG 2.2, SC 2.2.1][wcag-2-2-1]). Judgement: the text area lets the operator see and copy the chain. The console does not read the certificate, so the Register gives the thumbprint, which the operator finds again in the software product card and in the history. |
| Close of the result of an issue | With "Close" or the Escape key, with no confirmation, also when the operator did not save the chain | Judgement: the operator can issue a new certificate from the same CSR at once, and remove the registration of a lost client certificate. A lost certificate is in no component, so it gives no access. The result tells the operator to save the chain, and a confirmation at each close adds a step to each issue. The Escape key closes a dialog in the [WAI-ARIA APG: Dialog][apg-dialog] pattern. |
| Help text of a CSR field | The subject that the Register requires, with the names of the record | [WCAG 2.2, SC 3.3.2][wcag-3-3-2]: instructions when the content needs input from the user. A CN or an O that does not agree is the most probable refusal, and the CDS gives both values ([CDS: Certificate Signing Request Profile][cds-csr-profile]). Judgement: the text gives the CDS rule and the names that the view already shows. It is not a check: the Register still checks the CSR. |
| Save of a certificate or the CA bundle in the console | The UI saves the PEM text that the Register gave as a `.pem` file, with no second request. The file name has the ID of the record. | The Register does not keep the certificate, so a second request cannot get it again. Judgement: one way to save for the three files. An ID is unique and has only characters that are safe in a file name, and a name can have other characters. The design system button and text area do the work, so the design system gets no new component while only this UI saves a file ([Design system DESIGN.md][ui-design]: the components that more than one UI uses). |
| Format of an issue and of the CA bundle in the console API | JSON, with the PEM text in a field, not `application/pem-certificate-chain` | Judgement: the console API is JSON, and its OpenAPI document and the generated types describe JSON bodies ([ADR 0017][adr-0017]). The thumbprint goes in the same response as the chain. The operator interface keeps the PEM content type for scripts. |
| Confirmation of a revocation | A check with no change first, then a confirmation that names the certificate with its subject, its serial number, its thumbprint and the end of its validity | The confirmation names its record (see the rules for the actions). The console does not read the certificate, so only the Register can name it. Judgement: the revocation applies also to the certificates of the lab set-up, for example the server certificate of the Register, and a revocation of the wrong certificate cannot be reversed with no host access. The check also shows each refusal at its field before the confirmation. The revocation does each check again, so a change between the check and the revocation is safe. The row "Cascade in the confirmation" refuses a preview request because the detail already gives its data. No view gives the data of a certificate. |
| Version in a revocation in the console API | None | Judgement: a revocation is not of a participant or a key set. The only state that it depends on is whether the certificate is revoked, and a second revocation gives 409 with no change (see the row "Revocation of a certificate that is already revoked"). Thus a stale view cannot cause a change that the operator did not confirm. |
| Errors of a CSR and of a revocation in the console | The `detail` of a refused CSR at the CSR field. Field errors at `#/certificate` and `#/reason` for a revocation. | Judgement: an issue has one field, so the `detail` is its error, as for the registration of a client certificate. A revocation has two fields, so it uses the field errors of the console API (see the row "Field errors in the console API"). |
| Revocations in the console | A list in pages, newest first, with the serial number, the subject, the thumbprint, the time and the reason | The view shows the result of each action, and a toast is never the only place ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; the rules for the actions). The CRL gives the revocations only in DER. Judgement: the subject and the thumbprint let the operator compare a revocation with its confirmation and with the registered thumbprints of a software product, whose registration the operator removes with its own action. The Register gets the certificate in each revocation request, so it can keep these values. |
| Format of a serial number in the console | Uppercase hexadecimal, with two digits for each byte and no separators | Judgement: the CDS and the ACCC give no format. The operator compares the serial number in the confirmation and in the list with the output of `openssl x509 -serial`, which has this format. The Register keeps its own format in the database: the console API changes only the format of its responses. |
| Server certificates in the detail of a data holder brand | Not shown. The history shows each issue with its thumbprint. | Judgement: the Register does not register a server certificate, and no API of the Register uses it (see the row "Record of a server certificate issue"). The event is the only record. |
| Browser tests of the CDR certificates | One journey for a client certificate issue, its save and its revocation from the saved file. One journey for a server certificate issue. One for the save of the CA bundle. Each test makes its own participants and its own CSR key, and keeps no key in the repository. | [ADR 0019][adr-0019]: browser tests check the main journeys. Judgement: a revocation needs the chain of an issue, so one journey proves the sequence. The lab of the browser tests in CI has no conformance data. The repository contains no private keys ([Non-functional requirements](#non-functional-requirements-1)). Each issue gives a new serial number, so the test runs again in the same lab. |
| Source of the seed file in the chart | The name of an existing ConfigMap that the operator makes | Judgement: the data holder chart gets its brand directory the same way ([data holder: Decisions][dataholder-decisions]). With the file in the values, a change of the values after the first load has no effect, which a reader of the values does not expect. A list in a later values file replaces the full list ([Helm: values files][helm-values]), so a deployment copies the full file to change one host. A seed preset in the chart makes the Register chart depend on the presets of the data holder. |
| Seed in the chart by default | Off | Judgement: the Register start does not need a file (see Lab-only features), and a deployment with no lab data stays possible. The data holder chart always loads its products, because each brand has a product catalogue. A Register has no default set of participants. |
| Load of the seed file by the chart | A lab-only feature in the chart | Judgement: the Register is a mock, so its chart can run a lab-only feature, as the data holder chart does for its products. With no load, an operator makes each brand by hand in the operator console. |
| Participants of the example seed | The four preset banks only, with no data recipients | Judgement: the data recipients of the lab seed are fixtures of the conformance suite, with hosts that are only in the lab. A data holder is ready when its list of software product statuses is empty ([data holder: Register statuses][dataholder-register-statuses]). |
| Hosts of the example seed | Invented hosts on `.test`, one for each brand. The `websiteUri` keeps the value of the lab seed. | `.test` is reserved "for use in testing" ([RFC 6761, section 6.2][rfc6761-test]). Judgement: a URI with a placeholder does not load, because each URI must be absolute with a host (see Participant model), and a file that loads can be tested. The website of a brand is a fixed name of the brand, not a value of one deployment. |

[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-apis
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-federation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-federation
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-archiving]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#archiving
[cds-get-oidc]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-openid-provider-config
[cds-get-jwks]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-jwks
[cds-get-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-statement-assertion-ssa
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-get-dh-brands-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands-summary
[cds-get-dh-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-statuses
[cds-get-sp-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-products-statuses
[cds-get-dr-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients-statuses
[cds-get-drs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients
[cds-openapi]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_register.json
[cds-http-headers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-headers
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[cds-error-403-adr-not-active]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-adr-status-not-active
[cds-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-errors
[cds-known-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#standardise-register-api-error-codes
[cds-pagination]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination
[cds-dr-calling-register]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-the-cdr-register
[cds-client-auth]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#client-authentication
[cds-private-key-jwt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#private-key-jwt-client-authentication
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-dcr-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#tocSregistrationproperties
[rfc3986-2]: https://www.rfc-editor.org/rfc/rfc3986#section-2
[fapi-as]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#authorization-server
[oidc-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-ciphers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ciphers
[cds-cert-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-cert-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cds-nfrs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#non-functional-requirements
[cds-register-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-endpoints
[cds-error-406-unsupported-version]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-406-header-unsupported-version
[cds-error-400-invalid-version]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-header-invalid-version
[cds-error-400-header-missing]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-header-missing
[cds-error-400-field-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid
[cds-error-5xx-unexpected]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-5xx-unexpected-error
[cds-error-invalid-industry]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-error-404-field-invalid-industry
[cds-cors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors
[cds-schema-jwk]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_schemas_tocSjwk
[cds-schema-brand-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_schemas_tocSdataholderbrandsummaryv2
[fapi-tls]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#tls-considerations
[bcp195]: https://www.rfc-editor.org/rfc/rfc9325#section-4.2
[k8s-health]: https://kubernetes.io/docs/reference/using-api/health-checks/
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[rfc9110-etag]: https://www.rfc-editor.org/rfc/rfc9110#name-etag
[rfc9110-host]: https://www.rfc-editor.org/rfc/rfc9110#section-7.2
[rfc9111-no-cache]: https://www.rfc-editor.org/rfc/rfc9111#name-no-cache-2
[rfc7517-jwks]: https://www.rfc-editor.org/rfc/rfc7517#section-5
[rfc6749-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
[rfc6750]: https://www.rfc-editor.org/rfc/rfc6750
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[readme-sources]: ../../README.md#sources-for-requirements
[accc-known-issues]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29950003
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[accc-cert-mgmt]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management
[accc-mock-ssa-test]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/API/SSA/GetSoftwareStatementAssertion_Tests.cs#L416
[accc-mock-dr-ssa]: https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.SDK/Services/Register/SsaService.cs#L46-L56
[cds-registration-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-validation
[accc-mock-token]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/TokenController.cs
[accc-mock-dh-statuses]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/RegisterStatusRepository.cs#L65-L77
[accc-brands-summary]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982
[fetch-cors]: https://fetch.spec.whatwg.org/#http-cors-protocol
[accc-mock-certs]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CertificateManagement/README.md
[research-accc-headers]: ../../docs/research/2026-09-25-accc-mock-register.md
[support-removed-sp]: https://cdr-support.zendesk.com/hc/en-us/articles/900001657906
[support-brands]: https://cdr-support.zendesk.com/hc/en-us/articles/900005881263
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[dsb-537]: https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/537
[adr-0003]: ../../docs/adr/0003-event-sourcing.md
[adr-0006]: ../../docs/adr/0006-telemetry.md
[adr-0011]: ../../docs/adr/0011-domain-data-as-typed-go.md
[adr-0012]: ../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0017]: ../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0019]: ../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[adr-0021]: ../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0022]: ../../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[adr-0010]: ../../docs/adr/0010-background-jobs-with-river.md
[pki-design]: ../../pki/DESIGN.md
[pki-design-keys]: ../../pki/DESIGN.md#keys
[rfc5280-crl]: https://www.rfc-editor.org/rfc/rfc5280#section-5
[rfc5280-crl-number]: https://www.rfc-editor.org/rfc/rfc5280#section-5.2.3
[rfc5280-crldp]: https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.13
[rfc2585-mime]: https://www.rfc-editor.org/rfc/rfc2585#section-4.2
[rfc9110-methods]: https://www.rfc-editor.org/rfc/rfc9110#section-9.1
[rfc9110-405]: https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
[rfc9110-503]: https://www.rfc-editor.org/rfc/rfc9110#name-503-service-unavailable
[pki-design-revocation]: ../../pki/DESIGN.md#revocation
[pki-design-profile]: ../../pki/DESIGN.md#certificate-profile
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[accc-cp]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf
[accc-portal-guide]: https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf
[cds-cert-mgmt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-management
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[rfc5280-bounds]: https://www.rfc-editor.org/rfc/rfc5280#appendix-A
[rfc8555-chain]: https://www.rfc-editor.org/rfc/rfc8555#section-9.1
[rfc8446-cert]: https://www.rfc-editor.org/rfc/rfc8446#section-4.4.2
[ui-design]: ../../web/packages/ui/DESIGN.md
[ui-brands]: ../../web/packages/ui/DESIGN.md#brands
[go-cop]: https://pkg.go.dev/net/http#CrossOriginProtection
[owasp-csrf]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
[rfc9700-clickjacking]: https://www.rfc-editor.org/rfc/rfc9700.html#section-4.16
[rfc9111]: https://www.rfc-editor.org/rfc/rfc9111
[vite-dns-rebinding]: https://github.com/vitejs/vite/security/advisories/GHSA-vg6x-rcgg-rjx6
[rfc9110-delete]: https://www.rfc-editor.org/rfc/rfc9110#name-delete
[rfc9110-idempotent]: https://www.rfc-editor.org/rfc/rfc9110#name-idempotent-methods
[rfc9110-if-match]: https://www.rfc-editor.org/rfc/rfc9110#name-if-match
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-put]: https://www.rfc-editor.org/rfc/rfc9110#name-put
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#name-409-conflict
[rfc7396]: https://www.rfc-editor.org/rfc/rfc7396
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[adr-0013]: ../../docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[fetch-metadata]: https://w3c.github.io/webappsec-fetch-metadata/
[strict-csp]: https://web.dev/articles/strict-csp
[bcd-trusted-types]: https://github.com/mdn/browser-compat-data/blob/main/api/TrustedTypePolicyFactory.json
[rfc9562-v4]: https://www.rfc-editor.org/rfc/rfc9562#section-5.4
[rfc6901-6]: https://www.rfc-editor.org/rfc/rfc6901#section-6
[rfc9457-3]: https://www.rfc-editor.org/rfc/rfc9457#section-3
[rfc9457-3-2]: https://www.rfc-editor.org/rfc/rfc9457#section-3.2
[rfc9110-post]: https://www.rfc-editor.org/rfc/rfc9110#name-post
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-3-3-3]: https://www.w3.org/TR/WCAG22/#error-suggestion
[govuk-error-summary]: https://design-system.service.gov.uk/components/error-summary/
[govuk-question-pages]: https://design-system.service.gov.uk/patterns/question-pages/
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[apg-dialog]: https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/
[html-download]: https://html.spec.whatwg.org/multipage/links.html#downloading-resources
[rfc4514]: https://www.rfc-editor.org/rfc/rfc4514
[k8s-configmap-volume]: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
[k8s-init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[helm-values]: https://helm.sh/docs/chart_template_guide/values_files/
[dataholder-decisions]: ../dataholder/DESIGN.md#decisions
[dataholder-register-statuses]: ../dataholder/DESIGN.md#register-statuses
[adr-0020]: ../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[register-chart]: ../../charts/register/README.md
