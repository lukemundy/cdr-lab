# CDR Register APIs: research notes

- **Research date:** 2026-09-25
- **Applies to:** Consumer Data Standards (CDS) version 1.36.0
- **Purpose:** These notes give facts, with sources, for a requirements specification for a mock CDR Register. They contain no design.

Check the sources again before you depend on a fact; the standards change.

### Primary sources

| Source | URL |
|---|---|
| CDS standards site (live, 1.36.0 at research date) | https://consumerdatastandardsaustralia.github.io/standards/ |
| CDS 1.36.0 archive | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/ |
| Register APIs section | https://consumerdatastandardsaustralia.github.io/standards/#register-apis |
| Register OpenAPI (JSON) | https://consumerdatastandardsaustralia.github.io/standards/includes/swagger/cds_register.json |
| Register OpenAPI (YAML) | https://consumerdatastandardsaustralia.github.io/standards/includes/swagger/cds_register.yaml |
| Register OpenAPI at tag v1.36.0 | https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_register.json |
| Endpoint version schedule | https://consumerdatastandardsaustralia.github.io/standards/includes/endpoint-version-schedule/ |
| DSB standards repository (tag v1.36.0) | https://github.com/ConsumerDataStandardsAustralia/standards/tree/v1.36.0 |
| DSB repository licence | https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/LICENCE.md |
| ACCC Developer Portal, Delivery Dates | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081043/Delivery+Dates |
| ACCC mock Register (reference implementation, not normative) | https://github.com/ConsumerDataRight/mock-register |

## 0. Sources and conventions

### Source abbreviations

The table below gives a short tag for each source. The notes use these tags as citations. Each path is in the DSB repository at tag `v1.36.0`. Add this prefix to a path to get its URL: `https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/`. Each "Rendered" anchor is on the page https://consumerdatastandardsaustralia.github.io/standards/.

| Tag | Source |
|---|---|
| [SWAGGER] | `docs/includes/swagger/cds_register.json` (OpenAPI 3.0.3, `info.version` = `1.36.0`). Published copy: https://consumerdatastandardsaustralia.github.io/standards/includes/swagger/cds_register.json (also `.yaml`). |
| [REG] | `slate/source/includes/_register.md.erb`. Rendered: https://consumerdatastandardsaustralia.github.io/standards/#register-apis |
| [REG-SEC] | `slate/source/includes/security/endpoints/_register.md`. Rendered: https://consumerdatastandardsaustralia.github.io/standards/#register-endpoints and #participant-endpoints |
| [EVS] | `slate/source/includes/endpoint-version-schedule/includes/_register.md`. Rendered: https://consumerdatastandardsaustralia.github.io/standards/includes/endpoint-version-schedule/ |
| [CLIENT-AUTH] | `slate/source/includes/security/_client_authentication.md`. Rendered: #client-authentication |
| [DCR] | `slate/source/includes/security/_client_registration.md.erb`. Rendered: #client-registration, #software-statement-assertion-ssa, #ssa-definition |
| [STATUS] | `slate/source/includes/security/_participant_statuses.md.erb`. Rendered: #participant-statuses, #metadata-cache-management |
| [FED] | `slate/source/includes/security/_cdr_federation.md`. Rendered: #cdr-federation |
| [CERT] | `slate/source/includes/security/_certificate_management.md`. Rendered: #certificate-management |
| [TLS] | `slate/source/includes/security/_transport_security.md`. Rendered: #transaction-security |
| [CORS] | `slate/source/includes/security/_cors.md`. Rendered: #cors |
| [HDR] | `slate/source/includes/standards/_headers.md`. Rendered: #http-headers |
| [VER] | `slate/source/includes/standards/_versioning.md`. Rendered: #versioning |
| [PAGE] | `slate/source/includes/standards/_pagination.md`. Rendered: #pagination |
| [CODES] | `slate/source/includes/standards/_response-codes.md`. Rendered: #http-response-codes |
| [ERR] | `slate/source/includes/standards/_errors.md`. Rendered: #error-codes, #cdr-register-errors |
| [TYPES] | `slate/source/includes/standards/_types.md`. Rendered: #common-field-types |
| [SCOPES] | `slate/source/includes/_scopes.md.erb`. Rendered: #admin-amp-registration |
| [KNOWN] | `slate/source/includes/known-issues.md`. Rendered: #known-issues |
| [ADMIN] | `docs/includes/swagger/cds_admin.json` (Admin APIs that the Register calls on Data Holders). |
| [RN136] / [RN135] | `slate/source/includes/releasenotes/releasenotes.1.36.0.html.md` / `...1.35.0.html.md` |
| [IMG-ADR] / [IMG-SP] | `slate/source/images/ADRStateMachine.png` / `SoftwareProductStateMachine.png` (state diagrams referenced by [STATUS]) |
| [LIC] | `LICENCE.md` in the repository root |
| [ACCC-DD] | ACCC Developer Portal "Delivery Dates": https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081043/Delivery+Dates |
| [ACCC-MOCK] | ACCC reference mock Register, `https://github.com/ConsumerDataRight/mock-register`, release v2.2.8 (2026-07-15), MIT licence. README says it aligns to CDS v1.36.0. NOT normative. I use it only to show how the ACCC interprets the standard. |

### Conventions in these notes

- "MUST", "SHOULD" and "MAY" in quotes are the normative words of the source.
- **UNVERIFIED** = I did not find a primary source for this item.
- **DISCREPANCY** = two primary sources do not agree.
- **INFERENCE** = my conclusion from the sources, not a direct statement.

### Currency check

- Tag `v1.36.0` is the newest tag and newest release in the DSB repository (release date 2025-12-04). The `master` branch head is the 1.36.0 release commit. The live site swagger has `info.version` `1.36.0`. Thus 1.36.0 is the current version at the date of research. Source: `gh api repos/ConsumerDataStandardsAustralia/standards/tags`, and the live swagger URL above.
- Archived 1.36.0 site: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/ (HTTP 200). The swagger file is not at the archive path `includes/swagger/cds_register.json` (HTTP 404). Use the GitHub tag or the live site.

### Licence

- DSB repository `LICENCE.md`: "Unless otherwise stated, all specifications and downloadable reference applications are subject to this MIT Open Licence." Copyright 2018 Data61. [LIC]
- GitHub API reports the licence as `NOASSERTION` / "Other", because the file is named `LICENCE.md` and has a preface line. The text is the MIT licence.
- ACCC mock-register: MIT. [ACCC-MOCK]

### Other OpenAPI file in the repository (do not use)

- `docs/includes/swagger/cds_register_banking_non_bank_lending.json` (and `.yaml`) has `info.version` `1.34.1` and hard-codes `/all/` into recipient paths. It is a stale artefact. The authoritative file for 1.36.0 is `cds_register.json`. [GH tree at v1.36.0]

---

## 1. Endpoint inventory

### 1.1 Base URLs

| Name | URL | Source |
|---|---|---|
| Production TLS | `https://api.cdr.gov.au` | [REG], [SWAGGER] `servers` |
| Production MTLS | `https://secure.api.cdr.gov.au` | [REG], [SWAGGER] per-operation `servers` on Get Data Holder Brands and Get SSA |

The Register has no published test or sandbox base URL in the standards. **UNVERIFIED** for any ACCC test environment URL.

### 1.2 Summary table

Paths from [SWAGGER]. Auth columns from [REG] table. Current versions from [SWAGGER] `x-version`. Supported, deprecated and obsolete versions from [EVS].

| API | Method, path | Current x-v | Other versions in [EVS] | Transport | Bearer token / scope | Caller ([REG]) |
|---|---|---|---|---|---|---|
| Get OpenId Provider Config | `GET /idp/.well-known/openid-configuration` | none (no x-v) | n/a | TLS | No | DR Brand or Software Product |
| Get JWKS | `GET /cdr-register/v1/jwks` | none (no x-v) | n/a | TLS | No | Data Holder |
| Get Data Holder Brands | `GET /cdr-register/v1/{industry}/data-holders/brands` | 3 | V1 obsolete (retired 2023-09-21); V2 deprecated in 1.35.0, retirement "Refer to ACCC" | MTLS | Yes, `cdr-register:read` | DR Brand or Software Product |
| Get Data Holder Brands Summary | `GET /cdr-register/v1/{industry}/data-holders/brands/summary` | 2 | V1 deprecated in 1.35.0, retirement "Refer to ACCC" | TLS | No | Public client |
| Get Software Statement Assertion (SSA) | `GET /cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa` | 4 | V1, V2 obsolete (retired 2023-09-21); V3 deprecated in 1.35.0, retirement "Refer to ACCC" | MTLS | Yes, `cdr-register:read` | DR Brand or Software Product |
| Get Data Holder Statuses | `GET /cdr-register/v1/{industry}/data-holders/status` | 2 | V1 deprecated in 1.35.0, retirement "Refer to ACCC" | TLS | No | Data Recipient |
| Get Software Products Statuses | `GET /cdr-register/v1/{industry}/data-recipients/brands/software-products/status` | 3 | V1 obsolete (retired 2023-09-21); V2 deprecated in 1.35.0, retirement "Refer to ACCC" | TLS | No | Data Holder Brand |
| Get Data Recipients Statuses | `GET /cdr-register/v1/{industry}/data-recipients/status` | 3 | V1 obsolete (retired 2023-09-21); V2 deprecated in 1.35.0, retirement "Refer to ACCC" | TLS | No | Data Holder Brand |
| Get Data Recipients | `GET /cdr-register/v1/{industry}/data-recipients` | 4 | V2 obsolete (retired 2023-09-21); V3 deprecated in 1.35.0, retirement "Refer to ACCC" | TLS | No | Data Holder Brand |
| Token endpoint | `POST /idp/connect/token` on `https://secure.api.cdr.gov.au` | none | n/a | MTLS | n/a (issues tokens) | DR Brand or Software Product |

Notes:

- [EVS] lists paths without the `/cdr-register/v1` prefix, for example `/{industry}/data-holders/brands`. [SWAGGER] has the full paths. The `v1` segment in the path is not the endpoint version (INFERENCE: it is the major version of the Register API; see [VER] for the same pattern in `/cds-au/v1/`).
- [EVS] lists the OIDC and JWKS endpoints as `/.well-known/openid-configuration` and `/jwks`. [SWAGGER] gives `/idp/.well-known/openid-configuration` and `/cdr-register/v1/jwks`.
- The token endpoint path is not in [SWAGGER]. It comes from the non-normative OIDC discovery example in [REG-SEC] (`"token_endpoint": "https://secure.api.cdr.gov.au/idp/connect/token"`) and the non-normative token example in [CLIENT-AUTH] (`aud: https://secure.api.cdr.gov.au/idp/connect/token`). [ACCC-MOCK] uses `POST /idp/connect/token`.
- Deprecated-version retirement dates: [EVS] says "Refer to ACCC" and links [ACCC-DD]. On [ACCC-DD] (updated "Feb 24", read 2026-09-25) the only Register rows are delivery dates for the new versions: SSA V4, Brands V3, Brands Summary V2, DH Statuses V2, SP Statuses V3, DR Statuses V3, Data Recipients V4, all "11 March 2026". [ACCC-DD] gives no retirement date for the deprecated versions. **UNVERIFIED**: the retirement date of the deprecated versions (Brands V2, Summary V1, SSA V3, DH Statuses V1, SP Statuses V2, DR Statuses V2, Data Recipients V3).
- [ACCC-MOCK] still serves the deprecated versions next to the current versions: Brands V2+V3, SSA V3+V4, DR Statuses V2+V3, SP Statuses V2+V3, DH Statuses V1+V2, Data Recipients V3+V4. It does not implement Get Data Holder Brands Summary.

### 1.3 Industry path parameter

From [SWAGGER] `components.schemas`:

| Schema | Values | Used by |
|---|---|---|
| `IndustryEnumV2` | `banking`, `energy`, `non-bank-lending`, `telco`, `all` | Brands V3, Brands Summary V2, DH Statuses V2 |
| `IndustryEnumRecipients` | `all` (only) | SSA V4, SP Statuses V3, DR Statuses V3, Data Recipients V4 |
| `IndustriesEnumV2` | `banking`, `energy`, `non-bank-lending`, `telco` | response field `industries[]` on brand objects |

- Description on data-holder endpoints: "The industry the participant is retrieving data for (Banking, etc.)". On data-recipient endpoints: "Filters the response to participants associated with an industry." [SWAGGER]
- History: in 1.34.1 all endpoints used one `IndustryEnum` = `banking|energy|telco|all`. Version 1.35.0 added `non-bank-lending` ([RN135] Decision #367) and cut recipient endpoints to `all` only (the new versions). Source: `docs/includes/swagger/cds_register.json` at tag `v1.34.1` compared with `v1.36.0`.
- Thus deprecated versions (for example SSA V3, Data Recipients V3) accept `banking|energy|telco|all` and not `non-bank-lending` (INFERENCE from the 1.34.1 swagger; [ACCC-MOCK] CheckIndustry on V3 routes agrees: BANKING, ENERGY, TELCO, ALL).
- `telco` is in the Register enum, but telco data standards are still "candidate" standards in 1.36.0 (tree has `additional/candidates/telco`). INFERENCE: the Register lists telco, the telco data APIs are not binding.
- Invalid industry: see section 5.3 (DISCREPANCY between [ERR] and [SWAGGER]).

---

## 2. Endpoint details

Common to all versioned endpoints (all except OIDC config, JWKS and token):

- Request header `x-v` (required, string, positive integer) and `x-min-v` (optional). [SWAGGER] per-operation parameters.
- Response header `x-v` (required) with the version served. [SWAGGER] `components.headers.XV`.
- Response `Content-Type: application/json`. [SWAGGER] content types.
- No `x-fapi-interaction-id` parameter or response header is declared in [SWAGGER] for any Register operation. [HDR] makes the response header mandatory for data holder APIs and says the request header is "Not required for unauthenticated calls". INFERENCE: the standard does not state that the Register must echo `x-fapi-interaction-id`. **UNVERIFIED** whether the production Register echoes it. [ACCC-MOCK] only reads it for logging.
- Error body: `ResponseErrorListV2` (section 5).

### 2.1 Get OpenId Provider Config

- `GET https://api.cdr.gov.au/idp/.well-known/openid-configuration`. operationId `getRegisterOpenIdProviderConfig`. Tag "Register Operations". [SWAGGER]
- "This endpoint does not require CORS." No auth. [SWAGGER], generated doc "This operation does not require authentication."
- Response `ResponseOpenIDProviderConfigMetadata`. Required fields [SWAGGER]: `issuer`, `jwks_uri`, `token_endpoint`, `claims_supported`, `id_token_signing_alg_values_supported`, `subject_types_supported`, `code_challenge_methods_supported`, `scopes_supported`, `response_types_supported`, `grant_types_supported`, `token_endpoint_auth_methods_supported`, `tls_client_certificate_bound_access_tokens` (boolean), `token_endpoint_auth_signing_alg_values_supported`.
- Field notes [SWAGGER]:
  - `jwks_uri`: "contains the signing key(s) used to validate access tokens issued from the CDR Register. Note that this differs from the JWKS endpoint used to validate SSAs and CDR Register client authentication."
  - `id_token_signing_alg_values_supported`, `subject_types_supported`: "Given the CDR Register does not issue ID tokens, this field can be safely ignored."
  - `code_challenge_methods_supported`: "Given the CDR Register does not support PKCE, this field can be safely ignored."
- Non-normative example [REG-SEC]:
  ```
  issuer: https://api.cdr.gov.au/idp
  jwks_uri: https://api.cdr.gov.au/idp/.well-known/openid-configuration/jwks
  token_endpoint: https://secure.api.cdr.gov.au/idp/connect/token
  claims_supported: [sub]
  id_token_signing_alg_values_supported: [PS256]
  subject_types_supported: [public]
  scopes_supported: [cdr-register:read]
  response_types_supported: [token]
  grant_types_supported: [client_credentials]
  token_endpoint_auth_methods_supported: [private_key_jwt]
  tls_client_certificate_bound_access_tokens: true
  request_object_signing_alg_values_supported: [PS256]
  ```
- DISCREPANCY: the [REG-SEC] example omits two fields that [SWAGGER] marks required (`code_challenge_methods_supported`, `token_endpoint_auth_signing_alg_values_supported`) and includes `request_object_signing_alg_values_supported`, which is not in the schema. [ACCC-MOCK] returns the schema fields: `code_challenge_methods_supported: ["plain","S256"]`, `token_endpoint_auth_signing_alg_values_supported: ["PS256"]`, and no `request_object_signing_alg_values_supported`.
- Two JWKS: the OIDC `jwks_uri` (access-token signing keys, `/idp/.well-known/openid-configuration/jwks`) is a different endpoint from Get JWKS (`/cdr-register/v1/jwks`, SSA signing and Register outbound client-auth keys). [SWAGGER] `jwks_uri` description; [ACCC-MOCK] implements both.

### 2.2 Get JWKS

- `GET https://api.cdr.gov.au/cdr-register/v1/jwks`. operationId `getRegisterJwks`. No auth, TLS, no CORS required. [SWAGGER], [REG]
- Description: "JWKS endpoint containing the public keys used by the CDR Register to validate the signature of issued SSAs and authenticate outbound calls to participants in the CDR." [SWAGGER]
- [DCR]: "CDR Register public keys are exposed on a dedicated JWKS URI endpoint." "The CDR Register JWKS endpoint is an unauthenticated endpoint, using TLS."
- Response `ResponseJWKS` = `{ keys: JWK[] }`. `JWK` required: `alg`, `e`, `key_ops` (array), `kid`, `kty`, `n`. The schema is RSA-only (`n`, `e`). [SWAGGER]
- `kid` note: "not guaranteed to be unique and additional parameters should be used to progressively identify a key within a set." [SWAGGER]. (Contrast: [TLS] JWKS section says Data Holder and Data Recipient JWK sets "MUST NOT contain multiple keys with the same kid" from 2022-07-04; that rule names DH and DR, not the Register.)
- Example JWKS in [DCR]: one RSA key, `key_ops: ["sign","verify"]`, `kid` 32 hex chars, `e: AQAB`.

### 2.3 Token endpoint (Register as OAuth authorisation server)

All from [CLIENT-AUTH] "Data Recipients calling the CDR Register" and "Private Key JWT Client Authentication", unless marked.

- `POST https://secure.api.cdr.gov.au/idp/connect/token`, `Content-Type: application/x-www-form-urlencoded`. MTLS. (Path from non-normative examples, see 1.2.)
- Form parameters:
  - `grant_type` "MUST be set to `client_credentials`".
  - `client_id`: OPTIONAL from 2023-11-13 (RFC 7521).
  - `client_assertion_type` REQUIRED = `urn:ietf:params:oauth:client-assertion-type:jwt-bearer`.
  - `client_assertion` REQUIRED.
  - `scope` OPTIONAL in the generic rule; example sends `scope=cdr-register%3Aread`.
- Client identity: "_client_id_, _iss_ and _sub_ claims MUST be set to the ID of the calling client 'Data Recipient Brand ID' OR 'Software Product ID' issued by the CDR Register." Notice: both are "currently supported"; brand ID "will be deprecated in the future and is currently retained for backwards compatibility."
- Client assertion claims (private_key_jwt):
  - `iss` REQUIRED, `sub` REQUIRED (both = client ID).
  - `aud` REQUIRED. The issuer identifier "SHOULD" be used. "the authorisation server MUST accept its Issuer Identifier, Token Endpoint URL, or the URI of the endpoint being invoked". Example: `https://secure.api.cdr.gov.au/idp/connect/token`.
  - `jti` REQUIRED; "These tokens MUST only be used once."
  - `exp` REQUIRED. `iat` OPTIONAL.
- Signing algorithm:
  - The example header has `alg: PS256`. [CLIENT-AUTH]
  - The generic rule is a JWT per OIDC Core section 9. [CLIENT-AUTH]
  - The DSB text for the Register token endpoint does not name the permitted algorithms.
  - The profile refers to FAPI 1.0 Advanced section 8.6 in other places. For example, [DCR] says "signed using algorithms specified in section 8.6". FAPI 1.0 Advanced section 8.6 permits PS256 and ES256.
  - [ACCC-MOCK] publishes `token_endpoint_auth_signing_alg_values_supported: ["PS256"]`, but it accepts PS256 or ES256 in the client assertion.
  - **UNVERIFIED** in DSB text: an explicit PS256/ES256 rule for Register client assertions.
- Client public keys: "Client public keys are obtained from the JWKS endpoints." For a Software Product this is the SSA `jwks_uri` (the Register holds it as software product metadata). [CLIENT-AUTH], [REG-SEC] JwksUri row: "ADR Brand: Used for client authentication for DRSP -> DH & Register communication and is populated in the SSA".
- Response example: `{"access_token":"eyJhbGciOiJQUz...","expires_in":7200,"token_type":"Bearer","scope":"cdr-register:read openid"}`. Non-normative. "Refresh tokens will not be provided for grant_type client_credentials".
- Token lifetime:
  - The standards give no normative lifetime for Register access tokens.
  - The "2 minutes to 10 minutes" rule applies to Data Holder access tokens only ("after the Data Holder issues it"). Source: `slate/source/includes/security/_tokens.md`, rendered anchor #access-token.
  - The non-normative example shows `expires_in: 7200`. [CLIENT-AUTH]
  - [ACCC-MOCK] uses `AccessTokenExpiryInSeconds: 300` by default.
  - **UNVERIFIED**: the production Register token lifetime.
- Certificate-bound tokens:
  - The OIDC config example has `tls_client_certificate_bound_access_tokens: true`. [REG-SEC]
  - [TLS] says: "MTLS MUST be supported as a Holder of Key (HoK) Mechanism." It also says: "resource requests MUST be validated to ensure the client certificate and access token match". The binding follows RFC 8705 section 3.
  - [ACCC-MOCK] issues a JWT access token with header `typ: at+jwt`. The token has `aud: cdr-register` and the claims `client_id`, `scope`, `jti` and `cnf: {"x5t#S256": <thumbprint>}`.
  - [ACCC-MOCK] rejects a resource call when the thumbprint of the presented certificate does not match `cnf`.
- Token error responses: the standards do not define them for the Register. INFERENCE: RFC 6749 section 5.2 format (`error`, `error_description`). [ACCC-MOCK] returns HTTP 400 with `invalid_client` for most failures and `unsupported_grant_type` for a wrong grant type.
- [ACCC-MOCK] also rejects these token requests:
  - a `scope` value other than `cdr-register:read`;
  - an assertion where `iss` is not equal to `sub`;
  - a reused `jti` (the mock keeps each `jti` until `exp` + 5 minutes);
  - a client certificate whose CN does not match a certificate registered for the software product.
- [ACCC-MOCK] allows a clock skew of 120 seconds.
- [ACCC-MOCK] accepts two audiences: the issuer, and `<secure base>/idp/connect/token`.
- [ACCC-MOCK] reads the token `client_id` as a software product ID (a GUID) at resource endpoints. Thus the ACCC mock does not support a brand ID as the client ID in practice. INFERENCE from `DiscoveryController.GetSoftwareProductIdFromAccessToken` in [ACCC-MOCK].

### 2.4 Get Data Holder Brands (x-v 3)

- `GET https://secure.api.cdr.gov.au/cdr-register/v1/{industry}/data-holders/brands`. operationId `getDataHolderBrands`. Scope `cdr-register:read`. MTLS + bearer token. [SWAGGER], [REG]
- "Allows Data Recipients to discover Data Holder Brands available in the CDR ecosystem." [SWAGGER]
- Parameters: `industry` (path, `IndustryEnumV2`), `Authorization` (header, required, RFC 6750), `x-v` (required), `x-min-v`, `updated-since` (query, DateTimeString: "returns results updated since the specified date-time"), `page` (PositiveInteger, default 1), `page-size` (PositiveInteger, default 25). [SWAGGER]
- No `If-None-Match` parameter and no `Etag` response header in [SWAGGER]. (Contrast the public endpoints.) [ACCC-MOCK] adds ETag anyway.
- Responses: 200 (`ResponseRegisterDataHolderBrandListV2`, header x-v), 400 (Missing Required Header / Invalid Version / Invalid Path Parameter), 401 (Invalid Bearer Token, `WWW-Authenticate` header), 406 (Unsupported Version). [SWAGGER]
- Response shape:
  - `data[]`: `RegisterDataHolderBrandV2`
    - required: `dataHolderBrandId` (max 36), `brandName` (max 200), `industries` (array of `IndustriesEnumV2`, uniqueItems), `logoUri`, `legalEntity`, `status` (`ACTIVE|INACTIVE|REMOVED`), `endpointDetail`, `authDetails[]`, `lastUpdated` (DateTimeString)
    - optional: `brandGroup` (ASCIIString, max 200; new in 1.36.0 "to support white labelled brands")
    - `legalEntity` (`LegalEntityDetail`): required `legalEntityId` (max 36), `legalEntityName` (max 36, see note), `logoUri`, `status` (`ACTIVE|REMOVED`); optional `registrationNumber`, `registrationDate` (DateString), `registeredCountry`, `abn` (max 11), `acn` (max 9), `arbn` (max 9), `anzsicDivision`, `organisationType` (`SOLE_TRADER|COMPANY|PARTNERSHIP|TRUST|GOVERNMENT_ENTITY|OTHER`)
    - `endpointDetail` (`RegisterDataHolderBrandServiceEndpointV2`): required `version` ("v" + major standards version, for example `v1`), `publicBaseUri`, `resourceBaseUri`, `infosecBaseUri`, `websiteUri`; optional `productBaseUri` (new in 1.36.0; "If null or not present, indicates that PRD endpoints are not available"), `extensionBaseUri`
    - `authDetails[]` (`RegisterDataHolderAuth`): `registerUType` (enum `SIGNED-JWT` only), `jwksEndpoint` (URI)
  - `links`: `LinksPaginated` (`self` required; `first`, `prev` mandatory if not first page; `next`, `last` mandatory if not last page)
  - `meta`: `MetaPaginated` (`totalRecords`, `totalPages`, both required)
- Note: `LegalEntityDetail.legalEntityName` has `maxLength: 36` in [SWAGGER] (1.34.1 too), while other names have 200. INFERENCE: probable spec defect; a real legal name can be longer than 36 chars.
- [REG] Participant Endpoints table defines what each DH base URI means (PublicBaseUri TLS, ProductBaseUri TLS, ResourceBaseUri MTLS, InfoSecBaseUri TLS, AdminBaseUri MTLS, ExtensionBaseUri TLS/MTLS). "AdminBaseUri" is registered with the Register but is not in the Get Data Holder Brands response. [REG-SEC]
- [ACCC-MOCK] behaviour: returns 403 `AdrStatusNotActive` when the calling software product (from token `client_id`) is not active; returns 400 `Field/Invalid` "Page is out of range" when page > totalPages.

### 2.5 Get Data Holder Brands Summary (x-v 2)

- `GET https://api.cdr.gov.au/cdr-register/v1/{industry}/data-holders/brands/summary`. Public, TLS, no token. Caller "Public Client". [SWAGGER], [REG]
- "Endpoint used by participants to discover public details of Data Holder Brands from the CDR Register." [SWAGGER]
- Parameters: `industry` (`IndustryEnumV2`), `x-v`, `x-min-v`, `If-None-Match` (ASCIIString). No pagination, no `updated-since`. [SWAGGER]
- Responses: 200 (headers `Etag`, `x-v`), 304 (header `Etag`), 400, 404 ("Industry Not Found", `ResponseErrorListV2`), 406. [SWAGGER]
- Response: `data[]` of `DataHolderBrandSummaryV2`; `links` (`self` only); `meta` (empty object).
  - required: `brandName`, `publicBaseUri`, `logoUri`, `industries`, `lastUpdated`
  - optional: `dataHolderBrandId`, `interimId` ("used to uniquely identify the record when dataHolderBrandId is not populated and is not to be reused"), `brandGroup` (new 1.36.0), `productBaseUri` (new 1.36.0), `abn`, `acn`, `arbn`
- INFERENCE: `dataHolderBrandId` is optional here so the Register can list brands that are not yet fully onboarded (they have only an `interimId`). The standard does not say this directly.
- [ACCC-MOCK] does not implement this endpoint.

### 2.6 Get Software Statement Assertion (SSA) (x-v 4)

- `GET https://secure.api.cdr.gov.au/cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa`. Scope `cdr-register:read`. MTLS + bearer token. [SWAGGER], [REG]
- Parameters: `industry` (`IndustryEnumRecipients` = `all`), `dataRecipientBrandId` (path), `softwareProductId` (path), `Authorization`, `x-v`, `x-min-v`. [SWAGGER]
- 200 body: JSON schema `{"type":"string"}` with media type `application/json`. The body is the compact JWS. [SWAGGER]. INFERENCE: the body is a raw JWT string, not a JSON-quoted string; the schema is ambiguous. [ACCC-MOCK] returns `Ok(ssa)` from ASP.NET (**UNVERIFIED** whether it is quoted).
- The SSA 200 `x-v` response header in [SWAGGER] has `default: "1"`, which is a schema quirk; the served version is 4.
- Responses [SWAGGER]: 200, 400 (Missing Required Header / Invalid Version / Invalid Path Parameter), 401 (Invalid Bearer Token), 403 ("Invalid BrandId"), 404 ("Invalid Software Product"), 406, 422 ("SSA validation failed").
- DISCREPANCY: [SWAGGER] says 403 = "Invalid BrandId". [ERR] says `urn:au-cds:error:cds-register:Field/InvalidBrand` is 404. [KNOWN] future target: 403 = "The ADR or the ADR's Software Product is not active", 404 = Invalid Industry / Invalid Brand Id / Invalid Software Product Id.
- [ACCC-MOCK]: 404 `cds-register:Field/InvalidSoftwareProduct` when path `softwareProductId` does not match token `client_id`; 403 `cds-all:Authorisation/AdrStatusNotActive` when the DR or software product is not active; 404 `Resource/NotFound` when no SSA found (for example brand/product mismatch).
- SSA rules: see section 3.

### 2.7 Get Data Holder Statuses (x-v 2)

- `GET https://api.cdr.gov.au/cdr-register/v1/{industry}/data-holders/status`. Public, TLS. Caller "Data Recipient". [SWAGGER], [REG]
- Parameters: `industry` (`IndustryEnumV2`), `x-v`, `x-min-v`, `If-None-Match`. [SWAGGER]
- Responses: 200 (`Etag`, `x-v`), 304, 400, 406. [SWAGGER]
- Response `DataHoldersStatusList`: `data[]` of `{ legalEntityId (max 36), status: ACTIVE|REMOVED }`; `links.self`; `meta` {}. The status is at Data Holder **legal entity** level, not brand level. [SWAGGER]

### 2.8 Get Software Products Statuses (x-v 3)

- `GET https://api.cdr.gov.au/cdr-register/v1/{industry}/data-recipients/brands/software-products/status`. Public, TLS. Caller "Data Holder Brand". [SWAGGER], [REG]
- Parameters: `industry` (`all`), `x-v`, `x-min-v`, `If-None-Match`.
- Responses: 200 (`Etag`, `x-v`), 304, 400, 406.
- Response `SoftwareProductsStatusList`: `data[]` of `{ softwareProductId (max 36), status: ACTIVE|INACTIVE|REMOVED }`; `links.self`; `meta` {}.

### 2.9 Get Data Recipients Statuses (x-v 3)

- `GET https://api.cdr.gov.au/cdr-register/v1/{industry}/data-recipients/status`. Public, TLS. Caller "Data Holder Brand".
- Parameters: `industry` (`all`), `x-v`, `x-min-v`, `If-None-Match`.
- Responses: 200 (`Etag`, `x-v`), 304, 400, 406.
- Response `DataRecipientsStatusList`: `data[]` of `{ legalEntityId (max 36), status: ACTIVE|SUSPENDED|REVOKED|SURRENDERED }`; `links.self`; `meta` {}.

### 2.10 Get Data Recipients (x-v 4)

- `GET https://api.cdr.gov.au/cdr-register/v1/{industry}/data-recipients`. Public, TLS. Caller "Data Holder Brand". Purpose: "Data Recipient, brand and product details to render Data Holder Consumer Dashboard." [REG]
- Parameters: `industry` (`all`), `x-v`, `x-min-v`, `If-None-Match`. No pagination. [SWAGGER]
- Responses: 200 (`Etag`, `x-v`), 304, 400, 406.
- Response `ResponseRegisterDataRecipientList`: `data[]` of `RegisterDataRecipient`; `links.self`; `meta` {}.
  - `RegisterDataRecipient` required: `legalEntityId`, `legalEntityName` (max 200), `accreditationNumber` (max 100), `accreditationLevel` (`UNRESTRICTED|SPONSORED`), `logoUri`, `status` (`ACTIVE|SUSPENDED|REVOKED|SURRENDERED`), `lastUpdated`. Optional `dataRecipientBrands[]`.
  - `DataRecipientBrandMetaData` required: `dataRecipientBrandId`, `brandName`, `logoUri`, `status` (`ACTIVE|INACTIVE|REMOVED`). Optional `softwareProducts[]`.
  - `SoftwareProductMetaData` required: `softwareProductId`, `softwareProductName` (max 200), `softwareProductDescription` (max 4000), `logoUri`, `status` (`ACTIVE|INACTIVE|REMOVED`).
- Note: this response does not carry redirect URIs, JWKS URI, revocation URI or recipient base URI. Those appear only in the SSA. [SWAGGER], [DCR]

---

## 3. Software Statement Assertion (SSA)

All from [DCR] unless marked.

- "The CDR Register MUST issue Software Statements to **active** Data Recipients for **active** Software Products."
- "MUST issue distinct Software Statements for each 'Software Product ID'."
- "Software Statements are issued with a **10 minute** expiry time." ([ACCC-MOCK] `SSA:ExpiryInSeconds: 600`.)
- "SSAs are signed by the CDR Register using **PS256**" (per FAPI 1.0 Advanced).
- Signing key: published at the Register Get JWKS endpoint (`/cdr-register/v1/jwks`), not the OIDC `jwks_uri`. [DCR], [SWAGGER]
- JOSE header in example: `alg: PS256`, `kid`, `typ: JWT`.

### 3.1 SSA claims

| Claim | Required | Modifiable* | Notes |
|---|---|---|---|
| `iss` | Required | | Value `cdr-register` |
| `iat` | Required | | seconds since epoch |
| `exp` | Required | | seconds since epoch |
| `jti` | Required | | unique, prevents reuse |
| `legal_entity_id` | Optional | | ADR legal entity ID |
| `legal_entity_name` | Optional | | |
| `org_id` | Required | | ADR **Brand** ID |
| `org_name` | Required | yes | Brand name |
| `client_name` | Required | yes | Software product name |
| `client_description` | Required | yes | |
| `client_uri` | Required | yes | |
| `redirect_uris` | Required | yes | array |
| `sector_identifier_uri` | Optional | yes | |
| `logo_uri` | Required | yes | |
| `tos_uri` | Optional | yes | |
| `policy_uri` | Optional | yes | |
| `jwks_uri` | Required | yes | ADR JWKS |
| `revocation_uri` | Required | yes | ADR arrangement revocation endpoint |
| `recipient_base_uri` | Required | yes | |
| `software_id` | Required | | Software Product ID; stable across versions |
| `software_roles` | Required | | `data-recipient-software-product` |
| `scope` | Required | yes | space-separated; includes `openid`, `profile`, CDS data scopes, `cdr:registration` |

\* "Modifiable" is the column in [DCR]. INFERENCE: it means the participant can change the value in the Register over time.

- Example `scope` in [DCR]: `openid profile bank:accounts.basic:read bank:accounts.detail:read bank:transactions:read bank:payees:read bank:regular_payments:read energy:... common:customer.basic:read common:customer.detail:read cdr:registration`.
- The SSA has no `aud` claim and no `sub` claim in the table or example.
- INFERENCE: the scope in the SSA reflects the industries and scopes the software product is registered for. The standard does not say how the Register derives it. **UNVERIFIED**: how the production Register computes `scope` (for example, does it include non-bank-lending scopes, and does it change by `industry` path value).

### 3.2 How Data Holders validate the SSA

- "SSA JWT signatures MUST be verified against the associated JWK published at the CDR Register JWKS endpoint." [DCR] Registration Validation
- "Data Holders MUST NOT register Data Recipient software products unless they present a verified Registration Request with a verified SSA."
- Validation "MUST be performed in accordance with [JWT]" (RFC 7519): thus `exp`, `iat`, signature.
- Registration request JWT is verified against a key from the SSA `jwks_uri`.
- "Data Holders MUST NOT allow multiple active registrations for the same `software_id`."
- From 2022-08-31, DHs "MUST ignore unsupported authorisation scopes presented in the SSA".
- Data Holders also check software product and ADR status (section 4.4).

---

## 4. Participant model and statuses

### 4.1 Entities

From [FED], [SWAGGER]:

- **Data Holder legal entity** → one or more **Data Holder Brands**. "a single designated organisation MAY be represented via the Register as multiple separate Data Holders to support multiple brands". "Where a Data Holder has multiple brands, each brand MUST have a separate issuer" (from 2022-07-04). Status (legal entity): `ACTIVE|REMOVED`. Brand status: `ACTIVE|INACTIVE|REMOVED`.
- **Data Recipient legal entity** (accredited) → **Data Recipient Brands** → **Software Products**. "a single accredited organisation is represented via the Register as a single Data Recipient and MAY be represented by multiple separate Software Products". A Software Product "MUST be registered with the Registrar" and "MAY be registered for use across one or more sectors".
  - DR legal entity status: `ACTIVE|SUSPENDED|REVOKED|SURRENDERED`. `accreditationLevel`: `UNRESTRICTED|SPONSORED`.
  - DR brand status: `ACTIVE|INACTIVE|REMOVED`.
  - Software product status: `ACTIVE|INACTIVE|REMOVED`.
- **Secondary Data Holder**: "is not registered with the Register". [FED]
- The Register is operated by the ACCC as Registrar. Functions: identity and access management of participant metadata, certificate management (issue, manage, revoke), discovery and search. [FED]
- Industries: `banking`, `energy`, `non-bank-lending`, `telco`. [SWAGGER] `IndustriesEnumV2`.
- IDs: all Register IDs are strings with `maxLength: 36` (brand, legal entity, software product). The standards do not require GUIDs. [SWAGGER]. [ACCC-MOCK] uses GUIDs. Examples in [DCR] use upper-case GUIDs.

### 4.2 State machines

From [IMG-ADR] and [IMG-SP] (images in [STATUS]):

- Data Recipient: Initial →Activate→ Active. Active →Suspend→ Suspended. Active →Revoke→ Revoked. Active →Surrender→ Surrendered. Suspended →Reactivate→ Active. Suspended →Revoke→ Revoked. Suspended →Surrender→ Surrendered. Revoked →Reactivate→ Active. Surrendered → Final.
- Software Product: Initial →Activate→ Active. Active →Deactivate→ Inactive. Inactive →Activate→ Active. Active →Remove→ Removed. Inactive →Remove→ Removed. Removed → Final.
- No state machine in the standards for DH legal entity, DH brand or DR brand. **UNVERIFIED**: allowed transitions for those.
- Note: the diagram lets Revoked return to Active, but the archive table lists Revoked as an "end state". [STATUS]

### 4.3 Cascade rules

[STATUS] "Status Mapping": when the DR status changes to non-active, software product statuses "MUST be changed accordingly":

| DR status | Cascaded Software Product status |
|---|---|
| Suspended | Inactive |
| Revoked | Removed |
| Surrendered | Removed |

- "When a Data Recipient status is Suspended, Revoked or Surrendered the Software Product status cannot be Active."
- "The CDR Registrar has the ability to change the status of a Software Product independently of the Data Recipient's accreditation status."
- "The status of a Data Recipient Brand does not impact Data Holder responsibilities due to the cascade rules". (No DR → DR brand cascade rule is stated. **UNVERIFIED**.)
- "The CDR Register MUST NOT provide an undefined status for Data Recipient or Data Recipient Software Product Statuses".
- The Register "MUST notify all Data Holders of the above changes in Data Recipient and Software Product status via the Register APIs" (that is, by the status endpoints that DHs poll).

### 4.4 Data Holder responsibilities table (drives mock test scenarios)

[STATUS]:

| DR | SP | Disclose | Facilitate consent auth | Facilitate withdrawal | Invalidate consents | Cleanup registration |
|---|---|---|---|---|---|---|
| Active | Active | Y | Y | Y | | |
| Active | Inactive | | | Y | | |
| Active | Removed | | | | Y | Y |
| Suspended | Inactive | | | Y | | |
| Suspended | Removed | | | | Y | Y |
| Revoked | Removed | | | | Y | Y |
| Surrendered | Removed | | | | Y | Y |

### 4.5 Archiving

"Dormant entries will be identified and archived automatically within the CDR Register." Dormant = end state for the past **2 years**. End states: Data Holder: Removed; Data Holder Brand: Removed; Data Recipient: Surrendered, Revoked; Data Recipient Brand: Removed; Software Product: Removed. [STATUS]. INFERENCE: archived entries no longer appear in API responses. **UNVERIFIED**.

---

## 5. Errors

### 5.1 Envelope

`ResponseErrorListV2` [SWAGGER], [ERR]:

```json
{ "errors": [ { "code": "...", "title": "...", "detail": "...", "meta": { "urn": "..." } } ] }
```

- `code`, `title`, `detail` required. `title` "MUST NOT change from occurrence to occurrence". `meta.urn` mandatory when `code` is application-specific. [SWAGGER] `ErrorV2`
- [ERR] also lists `isSecondaryDataHolderError` (not relevant to the Register; not in the Register schema).
- URN structure `urn:au-cds:error:<sub-type>:<category>/<code>`. Sub-type `cds-register` is "Reserved for CDR Register issued error codes only". [ERR]

### 5.2 Codes relevant to the Register

| HTTP | Code | When | Source |
|---|---|---|---|
| 400 | `urn:au-cds:error:cds-all:Header/Missing` | required header missing (x-v); detail = header name | [ERR] |
| 400 | `urn:au-cds:error:cds-all:Header/Invalid` | header present but invalid | [ERR] |
| 400 | `urn:au-cds:error:cds-all:Header/InvalidVersion` | x-v / x-min-v not a positive integer | [ERR] |
| 400 | `urn:au-cds:error:cds-all:Field/Missing` | missing field / query param | [ERR] |
| 400 | `urn:au-cds:error:cds-all:Field/Invalid` | invalid query/path value | [ERR] |
| 400 | `urn:au-cds:error:cds-all:Field/InvalidDateTime` | invalid `updated-since` | [ERR] |
| 400 | `urn:au-cds:error:cds-all:Field/InvalidPageSize` | `page-size` > 1000 | [ERR], [PAGE] |
| 422 | `urn:au-cds:error:cds-all:Field/InvalidPage` | page out of range; detail SHOULD be max pages | [ERR] |
| 401 | (no body schema) | invalid bearer token; `WWW-Authenticate` header per RFC 6750 | [SWAGGER] `RegisterGeneric401` |
| 403 | `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive` | ADR or its software product not active; detail SHOULD hold the status | [ERR], [KNOWN], [ACCC-MOCK] |
| 404 | `urn:au-cds:error:cds-register:Field/InvalidBrand` | bad `dataRecipientBrandId` (SSA) | [ERR] |
| 404 | `urn:au-cds:error:cds-register:Field/InvalidIndustry` | bad `industry` path | [ERR] |
| 404 | `urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct` | bad `softwareProductId` (SSA) | [ERR] |
| 404 | `urn:au-cds:error:cds-all:Resource/NotFound` | unknown URL | [ERR] |
| 406 | `urn:au-cds:error:cds-all:Header/UnsupportedVersion` | positive integer version outside supported range; detail MAY include min and max | [ERR] |
| 429 | (Retry-After header) | throttling | [CODES], [HDR] |
| 5xx | `urn:au-cds:error:cds-all:GeneralError/Unexpected`; 503 `urn:au-cds:error:cds-all:Service/Unavailable` | | [ERR] |

### 5.3 Error DISCREPANCIES (important for a mock)

1. **Page out of range**: [ERR] `Field/InvalidPage` is listed under 422. [ACCC-MOCK] returns 400 `Field/Invalid` "Page is out of range". [SWAGGER] Get Data Holder Brands lists no 422 response.
2. **Invalid industry**: [ERR] says 404 `cds-register:Field/InvalidIndustry`. [SWAGGER] lists no 404 for most endpoints (only Brands Summary has 404 "Industry Not Found" and SSA has 404 "Invalid Software Product"). [SWAGGER] 400 description "Invalid Path Parameter" suggests 400. [ACCC-MOCK] returns 400 `cds-all:Field/Invalid` with detail `industry`.
3. **Invalid brand on SSA**: [SWAGGER] 403 "Invalid BrandId"; [ERR] 404 `InvalidBrand`.
4. **SSA 422 "SSA validation failed"**: [SWAGGER] lists it; no error code is defined for it in [ERR].
5. [KNOWN] "Standardise Register API Error Codes" (Future improvements) states that "Future versions of the CDR Register API error codes are to be aligned": 400 Missing Required Header / Invalid Header / Invalid Version / Invalid Field for all endpoints; 403 not-active and 404 Invalid Industry / Invalid Brand Id / Invalid Software Product Id for SSA; 404 Invalid Industry for all others. Thus 1.36.0 officially admits that the Register error codes are not yet aligned.

---

## 6. Versioning, headers, caching, pagination

### 6.1 x-v / x-min-v negotiation

[HDR], [SWAGGER] parameter descriptions, [ERR]:

- `x-v` is mandatory, positive integer. `x-min-v` optional, positive integer.
- Respond with "the highest supported version between x-min-v and x-v".
- "If the value of x-min-v is equal to or higher than the value of x-v then the x-min-v header should be treated as absent."
- "If all versions requested are not supported then the endpoint MUST respond with a 406 Not Acceptable" → `urn:au-cds:error:cds-all:Header/UnsupportedVersion`.
- Non-integer or <= 0 → 400 `Header/InvalidVersion`. Missing x-v → 400 `Header/Missing` ("SHOULD" be supported).
- Response header `x-v` = served version, mandatory.
- INFERENCE: when `x-min-v` is absent and `x-v` is higher than the highest supported version, the standard wording gives 406 (only one version is requested). [ACCC-MOCK] message: "Requested version is lower than the minimum version or greater than maximum version." **UNVERIFIED** whether the production Register falls back to a lower version in this case.
- Endpoints without x-v: OIDC config, JWKS, token. [SWAGGER] (no parameters).
- `Accept`: if set to an unacceptable value → 406. Default `application/json`. [HDR]

### 6.2 Other headers

- `x-fapi-interaction-id`: see section 2 common notes. Not declared for Register operations.
- `Retry-After` with 429: optional. [HDR]
- Caching / compression headers: "not specified but are considered acceptable". [HDR]

### 6.3 ETag / If-None-Match

- Declared in [SWAGGER] for: Brands Summary, DH Statuses, SP Statuses, DR Statuses, Data Recipients (request `If-None-Match`, response `Etag`, 304 with `Etag`).
- Not declared for: Get Data Holder Brands (MTLS), SSA, OIDC config, JWKS.
- [STATUS]: "The CDR Register MAY cache the responses for **public** APIs hosted by the CDR Register. If caching is supported the CDR Register MUST support ETag cache validators allowing participants to detect whether the content has changed."
- [CODES]: 304 "May be returned if standard caching headers such as ETag or If-modified-since are utilised".

### 6.4 Pagination

- Only Get Data Holder Brands is paginated in [SWAGGER] (`page`, `page-size`, `LinksPaginated`, `MetaPaginated`).
- [PAGE]: defaults page 1, page-size 25; max page-size 1000 unless stated → `Field/InvalidPageSize`; `totalPages` MUST be 0 when `totalRecords` is 0; no transaction isolation needed; cursor-style links allowed.
- Other list endpoints return all rows with `links.self` and an empty `meta`.

### 6.5 CORS

- [CORS]: CORS "must be disabled (i.e. Access-Control-Allow-Origin set to *) for all unauthenticated endpoints unless specifically stated otherwise".
- OIDC config and JWKS: "This endpoint does not require CORS." [SWAGGER]
- INFERENCE: the other public Register endpoints (Summary, statuses, Data Recipients) fall under the general rule (ACAO `*`). No Register-specific statement found.

---

## 7. Transport security and certificates

- All calls HTTPS with TLS >= 1.2. [TLS]
- MTLS endpoints "MUST be configured according to the Certificate Trust Model". TLS endpoints "MUST be configured with a certificate issued by a public CA accepted by major web browsers." [REG-SEC] (stated for participant endpoints). For the Register: `api.cdr.gov.au` is TLS; `secure.api.cdr.gov.au` is MTLS. [REG]
- MTLS: client cert "MUST be issued by the CDR Certificate Authority (CA). The Server MUST NOT trust Client transport certificates issued by other authorities." Same for server certs. [TLS] (text is written for DR↔DH back-channel; [CERT] says the private chain applies to "all Register CA secured endpoints being hosted by Data Holders, Data Recipients and the Register".)
- "Endpoints ... specified as TLS, MUST NOT use MTLS." [TLS]
- Ciphers: from 2025-03-17, FAPI 1.0 Advanced section 8.5 plus only BCP195-recommended suites. [TLS]
- Register CA issues: DR client certificate (for Register APIs and DH APIs); DH server certificates for MTLS participant endpoints. [CERT]
- CSR profile [CERT]: CN = primary DNS (server) or Software Product Name (client); O = Legal Entity Name or Brand Name (server) / Brand Name (client); OU = `Consumer Data Right`; C mandatory; SHA256; RSA 2048.
- Certificate validation: validity, chain, key usage, revocation via CRL or OCSP. "The use of OCSP Stapling within the CDR ecosystem is not recommended." [CERT]
- Real CA root and intermediate certs (prod and test) are in the DSB repo: `docs/includes/register/certificates/{production,test}/ca_{root,intermediate}_{prod,test}.cer`, plus `ACCC_CPS_v1.0.docx`. [GH tree]
- [ACCC-MOCK] CertificateManagement README: "DigiCert is the Certificate Authority (CA) for the CDR". The ACCC mock uses a self-signed mock CA for both MTLS and TLS certs. [ACCC-MOCK]. Not stated in the DSB standards. 
- [ACCC-MOCK] terminates TLS in a gateway and forwards the client cert thumbprint and CN in headers (`X-TlsClientCertThumbprint`). Implementation detail only.

---

## 8. Register as a client of Data Holders (outbound)

Relevant if the mock Register must call Data Holders.

- Data Holders "MUST support either Private Key JWT Client Authentication or Self-signed JWT Client Authentication of the CDR Register" ("SHOULD" private_key_jwt). [CLIENT-AUTH]
  - private_key_jwt: DH issues a client ID to the Register out of band; "MAY be a static client ID value of `cdr-register`"; `grant_type=client_credentials`; `scope` MUST be the scope of the resource endpoint; no refresh token.
  - Self-signed JWT: client ID "MUST be set to a value of `cdr-register`"; JWT in `Authorization: Bearer`; claims `iss`, `sub` (= `cdr-register`), `aud` (for Metrics: the `<AdminBaseUri>`), `jti` (one-time), `exp`, `iat` optional. Example `alg: PS256`.
- Keys: the Register signs with keys at Get JWKS (`/cdr-register/v1/jwks`). [SWAGGER] Get JWKS description ("authenticate outbound calls to participants").
- DH Admin endpoints [ADMIN] (MTLS on AdminBaseUri):
  - `GET /admin/metrics` x-v 5, scope `admin:metrics.basic:read`, query `period`.
  - `POST /admin/register/metadata` x-v 1, scope `admin:metadata:update`, body `{"data":{"action":"REFRESH"},"meta":{}}`. "Indicate that a critical update to the metadata for Accredited Data Recipients has been made". [KNOWN] notes a response inconsistency (empty `{}` versus required `data`/`links`).
- Scopes [SCOPES]: `admin:metrics.basic:read`, `admin:metadata:update` (both "accessible ONLY to the CDR Register"), `cdr-register:read` ("allows Data Recipients to get a Software Statement Assertion and access to Data Holder metadata held by the CDR Register"), `cdr:registration`. Admin scopes "MUST never be included in the same access token as a CDR Data scope."

---

## 9. Non-functional expectations

- Participant polling [STATUS]:
  - Data Holders "MUST react to Data Recipient and associated Software Statuses changes within **5 minutes**". They poll Get Data Recipients Statuses, Get Software Products Statuses and Get Data Recipients.
  - DHs and DRs "MUST cache all other participant data they use and periodically update this cache using a slow poll."
  - Recommended cache periods: GetDataRecipientsStatus 2-5 min; GetSoftwareProductsStatus 2-5 min; GetDataRecipients (statuses) 2-5 min; GetDataRecipients (other data) 6 hours; GetDataHolderBrands (Private) 6 hours.
- If a DH gets no status or an unknown status, it "SHOULD ignore the value and use the previous status value", with "no upper bound". [STATUS]. This means a mock must never emit unknown status values (and a test feature could deliberately do so).
- The NFR section (availability, performance, throttling) of the standards is for Data Holders. I found no NFR for the Register itself in `slate/source/includes/nfrs/*`. **UNVERIFIED**: any Register availability or performance SLA (ACCC-side documents, not DSB).
- Access-token lifetime, rate limits and throttling of the Register: **UNVERIFIED**.

---

## 10. Changes in recent versions (for version support decisions)

- 1.35.0 [RN135]: Decision #367 "Updated Register endpoints to accommodate Non-Bank Lending" → new versions Brands V3, Summary V2, DH Statuses V2, SSA V4, SP Statuses V3, DR Statuses V3, Data Recipients V4; old versions deprecated. Maintenance #671 updated the Register token and openid-configuration non-normative examples.
- 1.35.1: "Register APIs: None". 
- 1.36.0 [RN136]: Decision #376 added `brandGroup` and `productBaseUri` to Get Data Holder Brands and Get Data Holder Brands Summary. INFERENCE: these are added as optional fields to the existing versions (Brands V3, Summary V2); x-version did not change.

---

## 11. Open questions and ambiguities (for the specification author)

1. Which deprecated versions must the mock serve? Retirement dates are "Refer to ACCC" and the ACCC page gives none. The ACCC mock serves both.
2. Error codes: follow [ERR] (404 cds-register codes), [SWAGGER] (400 / 403 descriptions), or the ACCC mock? [KNOWN] confirms misalignment.
3. Token endpoint: normative algorithms (PS256 only, or PS256 + ES256), access-token format (JWT or opaque), lifetime, and error format are not defined in DSB text.
4. Brand ID as client ID: standards still allow it (deprecated in future). The ACCC mock does not support it in practice.
5. OIDC discovery doc: schema and example do not agree.
6. Should the SSA 200 body be a JSON string or raw JWT text?
7. Should `x-fapi-interaction-id` be echoed?
8. ETag on Get Data Holder Brands: not declared, but allowed by generic caching rules.
9. SSA `scope` derivation from registered industries is not defined.
10. `LegalEntityDetail.legalEntityName` maxLength 36 looks like a defect.
11. DH brand / DR brand status transitions and archive visibility are not specified.
12. `telco` appears in the industry enums while telco standards are candidate only.
