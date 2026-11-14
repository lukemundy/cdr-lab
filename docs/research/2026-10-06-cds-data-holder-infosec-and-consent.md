# CDS data holder: information security profile and consent rules for a mock banking data holder

- Research date: 2026-10-06
- Consumer Data Standards (CDS) version: 1.36.0 (released 04/12/2025). This is the current version at the research date: the live site states version 1.36.0, the 1.36.0 archive URL redirects to the live site, and the newest tag and release in the CDS repository is `v1.36.0`. [S1 #version] [S1 #changelog-and-archives] [S2]
- Security profile: FAPI 1.0 Advanced, with PAR, PKCE and JARM. [S1 #overview] [S1 #authorization-code-flow]
- CDR Rules version: Competition and Consumer (Consumer Data Right) Rules 2020, Compilation No. 10, compilation date 4 March 2025. This is the latest compilation on the Federal Register of Legislation at the research date. [S16]
- Purpose: source-cited facts for the requirements of the information security profile and the consent functions of a mock banking data holder. These notes do not design an implementation.
- Related note: [CDR PKI and participant lifecycle](2026-09-25-cdr-pki-and-participant-lifecycle.md) ("the lifecycle note"). It gives the Register side of SSAs, statuses, Get Metrics and the PKI. This note adds the data holder side only.

Check the sources again before you depend on a fact; the rules and standards change.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site and archive) | https://consumerdatastandardsaustralia.github.io/standards/ and https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/ |
| S2 | CDS source repository at tag `v1.36.0` (DCR OpenAPI specification `docs/includes/swagger/cds_dcr.json`; security profile sources in `slate/source/includes/security/`) | https://github.com/ConsumerDataStandardsAustralia/standards/tree/v1.36.0 |
| S3 | Financial-grade API Security Profile 1.0 - Part 2: Advanced | https://openid.net/specs/openid-financial-api-part-2-1_0.html |
| S4 | Financial-grade API Security Profile 1.0 - Part 1: Baseline | https://openid.net/specs/openid-financial-api-part-1-1_0.html |
| S5 | JWT Secured Authorization Response Mode for OAuth 2.0 (JARM), Final, 17 August 2025 | https://openid.net/specs/oauth-v2-jarm.html |
| S6 | RFC 9126, OAuth 2.0 Pushed Authorization Requests (PAR) | https://www.rfc-editor.org/rfc/rfc9126 |
| S7 | RFC 7636, Proof Key for Code Exchange (PKCE) | https://www.rfc-editor.org/rfc/rfc7636 |
| S8 | RFC 7591, OAuth 2.0 Dynamic Client Registration Protocol | https://www.rfc-editor.org/rfc/rfc7591 |
| S9 | RFC 7592, OAuth 2.0 Dynamic Client Registration Management Protocol | https://www.rfc-editor.org/rfc/rfc7592 |
| S10 | RFC 8705, OAuth 2.0 Mutual-TLS Client Authentication and Certificate-Bound Access Tokens | https://www.rfc-editor.org/rfc/rfc8705 |
| S11 | RFC 7662, OAuth 2.0 Token Introspection | https://www.rfc-editor.org/rfc/rfc7662 |
| S12 | RFC 7009, OAuth 2.0 Token Revocation | https://www.rfc-editor.org/rfc/rfc7009 |
| S13 | RFC 6749, The OAuth 2.0 Authorization Framework | https://www.rfc-editor.org/rfc/rfc6749 |
| S14 | OpenID Connect Core 1.0 | https://openid.net/specs/openid-connect-core-1_0.html |
| S15 | OpenID Connect Discovery 1.0 | https://openid.net/specs/openid-connect-discovery-1_0.html |
| S16 | CDR Rules 2020, Compilation No. 10 | https://www.legislation.gov.au/F2020L00094/latest/text |
| S17 | CX Guidelines (part of the CDS consumer experience material; pages are listed in section 14) | https://cx.dsb.gov.au/ |
| S18 | CDR Support Portal articles (IDs given for each fact; list in section 14) | https://cdr-support.zendesk.com/hc/en-us/articles/<id> |
| S19 | Standards maintenance issues (number given for each fact) | https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/<number> |
| S20 | ACCC Developer Portal: Get Metrics | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/32800769/Get+Metrics |
| S21 | CDS consultations and decisions (issues of the CDS repository; number given for each fact) | https://github.com/ConsumerDataStandardsAustralia/standards/issues/<number> |

Notation:
- "MUST", "SHOULD", "SHALL" and "MAY" in quotes are normative words from the source.
- **UNVERIFIED** means that no primary source confirms the fact. Do not depend on it.
- **ARCHIVED** means that the source is old. A newer source replaces it, or the fact may be out of date.
- **ADR 0001** marks the result for an item in [ADR 0001](../adr/0001-auth-server-library.md): "proved" (the Verification table of the ADR lists it), "used" (the test provider of the ADR uses it, but the table does not list it) or "not covered".
- "Non-normative" marks an example in the CDS. An example is not a requirement.

---

## 1. Endpoints that the data holder must serve

### 1.1 Endpoint table

The data holder has the role of an OpenID Provider. [S1 #data-holder] The CDS gives a fixed path only for the discovery endpoint. The data holder publishes the URL of each other endpoint in its OpenID Provider metadata. [S1 #participant-endpoints] [S1 #openid-provider-configuration-endpoint]

| Endpoint | Path rule | Transport | Client authentication | Bearer token | Discovery field |
|---|---|---|---|---|---|
| OpenID Provider configuration | `<InfoSecBaseUri>/.well-known/openid-configuration` | TLS | No | No | none |
| JWKS | From discovery | TLS | No | No | `jwks_uri` |
| Authorisation | From discovery | TLS | No | No | `authorization_endpoint` |
| Pushed authorisation (PAR) | From discovery | MTLS | Yes (`private_key_jwt`) | No | `pushed_authorization_request_endpoint` |
| Token | From discovery | MTLS | Yes (`private_key_jwt`) | No | `token_endpoint` |
| UserInfo | From discovery | MTLS | No | Yes | `userinfo_endpoint` |
| Introspection | From discovery | MTLS | Yes (`private_key_jwt`) | No | `introspection_endpoint` |
| Token revocation | From discovery | MTLS | Yes (`private_key_jwt`) | No | `revocation_endpoint` |
| CDR arrangement revocation | From discovery, `POST` | MTLS | Yes (`private_key_jwt`) | No | `cdr_arrangement_revocation_endpoint` |
| Registration: `POST /register` | From discovery | MTLS | None (the request is a signed JWT with an SSA) | No | `registration_endpoint` |
| Registration: `GET` and `PUT /register/{ClientId}` | `registration_endpoint` plus the client ID | MTLS, holder of key | Access token from the client credentials grant | Yes, scope `cdr:registration` | `registration_endpoint` |
| Registration: `DELETE /register/{ClientId}` (optional) | `registration_endpoint` plus the client ID | MTLS, holder of key | Access token from the client credentials grant | Yes, scope `cdr:registration` | `registration_endpoint` |

Sources: [S1 #openid-provider-configuration-endpoint] [S1 #json-web-key-set-endpoint] [S1 #authorisation-endpoint] [S1 #pushed-authorisation-endpoint] [S1 #token-endpoint] [S1 #userinfo-endpoint] [S1 #introspection-endpoint] [S1 #token-revocation-endpoint] [S1 #cdr-arrangement-revocation-endpoint] [S1 #dynamic-client-registration-endpoints] [S2 `slate/source/includes/security/endpoints/_registration.md`]

Notes:
- The CDS says "MUST" for each of these endpoints, except two: the authorisation endpoint has no "MUST" sentence of its own (the Authorization Code Flow needs it), and `DELETE /register/{ClientId}` is "Optional". [S1 #security-endpoints] [S1 #dynamic-client-registration-endpoints]
- The discovery, JWKS and authorisation endpoints and the DCR `POST` do "not require CORS". The DCR endpoints have no version ("Versioning is not supported for this endpoint"), so they do not use `x-v`. [S1 #openid-provider-configuration-endpoint] [S1 #cdr-dynamic-client-registration-api_register-data-recipient-oauth-client]
- All InfoSec endpoints, DCR and CDR arrangement revocation are in the "High Priority" performance tier: 95% of calls in each hour within 1000 ms. [S1 #performance-requirements]

### 1.2 Base URIs

- `InfoSecBaseUri` is the "Base URI for the OIDC Discovery endpoint only". It uses TLS. "Endpoints specified in the Discovery endpoint have the requirements detailed in the Security Endpoints section." [S1 #participant-endpoints]
- "The InfoSecBaseUri is not intended to be the BaseUri where all related security endpoints are referenced from." The data holder can put the other security endpoints on other hosts, because they have different transport rules. [S18: 4403929644559]
- Base URIs "need not be unique". One host and path can serve more than one base URI. [S18: 4403929644559]
- The Register publishes `infosecBaseUri`, `resourceBaseUri`, `adminBaseUri` and the data holder `jwksEndpoint` for each brand in Get Data Holder Brands. See the lifecycle note, section 1.3.

### 1.3 Issuer

- "Where a Data Holder has multiple brands, each brand MUST have a separate issuer." (from 4 July 2022) [S1 #multi-brand-support-separate-issuers-for-data-holder-brands]
- OpenID Connect Discovery requires: "The issuer value returned MUST be identical to the Issuer URL that was used as the prefix to /.well-known/openid-configuration". It must also be the `iss` of each ID token. [S15 section 4.3] Thus `issuer` is equal to `InfoSecBaseUri`. The CDS example does not obey this rule (see section 12, item 1).
- The issuer is the `aud` of the registration request JWT and of the request object, and an accepted `aud` of client assertions (sections 2.1, 3.2 and 4.1).

### 1.4 Required discovery metadata

"At a minimum, the Data Holder metadata MUST include" these fields. [S1 #openid-provider-configuration-endpoint]

| Field | Value rule for CDS 1.36.0 |
|---|---|
| `issuer` | Issuer identifier URL (section 1.3) |
| `authorization_endpoint`, `token_endpoint`, `userinfo_endpoint`, `jwks_uri`, `registration_endpoint` | Endpoint URLs |
| `introspection_endpoint`, `revocation_endpoint` | Endpoint URLs (from RFC 8414) |
| `pushed_authorization_request_endpoint` | PAR endpoint URL |
| `require_pushed_authorization_requests` | Boolean. The data holder must accept authorisation request data "only" through PAR (section 3.1), thus the true value is `true`. The CDS example shows `true`. |
| `cdr_arrangement_revocation_endpoint` | Endpoint URL. CDS-specific field. |
| `scopes_supported` | The supported scopes (section 3.7) |
| `claims_supported` | The supported claims (section 3.8) |
| `acr_values_supported` | The supported `acr` values (section 6.2) |
| `subject_types_supported` | The CDS requires pairwise identifiers (section 3.9). The CDS example shows `["pairwise"]`. |
| `response_types_supported` | `code` is the only permitted value now (section 3.3) |
| `response_modes_supported` | Must include the JARM mode `jwt` in practice (section 3.4). The CDS gives no list of values. |
| `grant_types_supported` | `authorization_code`, `client_credentials`, and `refresh_token` in practice. The CDS gives no list of values. |
| `code_challenge_methods_supported` | `S256` (section 3.5) |
| `token_endpoint_auth_methods_supported` | `private_key_jwt` only (section 4.1) |
| `token_endpoint_auth_signing_alg_values_supported` | `PS256`, `ES256` or both |
| `id_token_signing_alg_values_supported` | `PS256`, `ES256` or both |
| `request_object_signing_alg_values_supported` | `PS256`, `ES256` or both |
| `authorization_signing_alg_values_supported` | `PS256`, `ES256` or both. "Required if Authorization Code Flow (response_type code) is supported." |
| `tls_client_certificate_bound_access_tokens` | "Value MUST be set to true." |
| `authorization_encryption_alg_values_supported`, `authorization_encryption_enc_values_supported` | Only when the data holder supports JARM encryption. The data holder can omit them or give empty arrays (section 3.4). |
| `id_token_encryption_alg_values_supported`, `id_token_encryption_enc_values_supported` | "Required only for Data Holders supporting OIDC Hybrid Flow". The hybrid flow is retired (section 3.3), so a new data holder omits them. |

- Algorithm values: "For JWS, both clients and authorization servers shall use PS256 or ES256 algorithms" and "shall not use none". [S3 section 8.6] The data holder selects PS256, ES256 or both. Data recipients must support both. [S18: 6828251436431]
- JWKS: "Data Holders and Data Recipients JWK sets MUST NOT contain multiple keys with the same kid." [S1 #jwks-uris] "Data Holder public keys MUST only be obtained from the standard OIDC endpoint used for that purpose." [S1 #json-web-key-sets]
- The Register also holds a `jwksEndpoint` for each data holder brand. Data recipients use it to verify calls from the data holder (section 5.4). It can be a different URL from the discovery `jwks_uri`, so that outbound calls can use a different platform from the identity provider. [S1 #participant-endpoints] [S18: 6382507783183]

---

## 2. Dynamic Client Registration (DCR)

The lifecycle note, section 3.5, gives the SSA claims, the SSA lifetime (10 minutes), the SSA signature (PS256) and the Register JWKS. This section gives the data holder side.

### 2.1 Registration request JWT

- "The request MUST be presented in the format of a [RFC7519] compliant JWT." It "MUST use the HTTP POST method, using the application/jwt content type." [S1 #registration-request-using-jwt] ADR 0001: proved (workaround `RegistrationJWT`).
- "The JWT MUST be signed using algorithms specified in section 8.6 of [FAPI-1.0-Advanced]": PS256 or ES256. [S1 #registration-request-using-jwt] [S3 section 8.6]
- The data recipient signs the JWT with a key from the `jwks_uri` in the SSA (section 2.2).

Claims of the request JWT: [S1 #registration-request-using-jwt] [S2 `cds_dcr.json`]

| Claim | Requirement (CDS text) | Rule |
|---|---|---|
| `iss` | Required | The software product ID from the Register |
| `iat` | Required | Issue time |
| `exp` | Required | Expiry time |
| `jti` | Required | "Unique identifier for the JWT, used to prevent reuse of the token." |
| `aud` | Required | "Contains the DH issuer value as described in the OIDC Discovery Document." |
| `software_statement` | Required | The SSA from the Register |
| `redirect_uris` | Optional (the OpenAPI schema says mandatory) | "If used, redirect_uris MUST match or be a subset of the redirect_uris as defined in the SSA." |
| `token_endpoint_auth_method` | Required | "The only supported method will be private_key_jwt." |
| `token_endpoint_auth_signing_alg` | Required | `PS256` or `ES256` |
| `grant_types` | Required | Values: `client_credentials`, `authorization_code`, `refresh_token` |
| `response_types` | Required | Values in the schema: `code`, `code id_token`. "From May 12th 2025 Data Recipients SHALL only use response_type value code for authorisation requests." |
| `application_type` | Optional | "The only supported application type will be web." |
| `id_token_signed_response_alg` | Required | `PS256` or `ES256` |
| `id_token_encrypted_response_alg`, `id_token_encrypted_response_enc` | Conditional | "Required only if OIDC Hybrid Flow ... is registered. SHALL NOT be used for Authorization Code Flow." |
| `authorization_signed_response_alg` | Conditional | "Required if response_type of code is registered by the client." `PS256` or `ES256`. "The algorithm none is not allowed." |
| `authorization_encrypted_response_alg` | Conditional | `RSA-OAEP` or `RSA-OAEP-256`. "If unspecified, the default is that no encryption is performed." Required if the `enc` claim is present. |
| `authorization_encrypted_response_enc` | Optional | `A256GCM` or `A128CBC-HS256`. The default is `A128CBC-HS256` when the `alg` claim is present. |
| `request_object_signing_alg` | Optional (the OpenAPI schema says mandatory) | `PS256` or `ES256`. "If field not present in the request, data holders are expected to respond with an appropriate default value." |

- The CDS gives no maximum lifetime for the request JWT, and it does not say that the data holder must keep `jti` values to refuse a second use. Both are **UNVERIFIED** as requirements.

### 2.2 Validation rules

- "Validation and use of the JWT and the claims described above MUST be performed in accordance with [JWT]." [S1 #registration-validation]
- "SSA JWT signatures MUST be verified against the associated JWK published at the CDR Register JWKS endpoint." [S1 #registration-validation] The Register JWKS can have more than one key. The data holder finds the key by `kid`. [S18: 6662499263887]
- "Data Holders are required to verify the request JWT signature against an associated JWK from the jwks_uri extracted from the SSA." [S1 #registration-validation] Thus the data holder must make an outbound TLS call to a URL that it does not know before the request.
- A data holder must not refuse the first registration because its egress rules do not know the recipient host: "Giving an initial error is a non-compliant solution", and "Requiring a separate path for ADR registration is a non-compliant solution". [S18: 4403543156751]
- "Data Holders MUST NOT register Data Recipient software products unless they present a verified Registration Request with a verified SSA." [S1 #software-statement-assertion-ssa]
- "Data Holders MUST ignore unsupported authorisation scopes presented in the SSA for the creation and update of client registrations." (from 31 August 2022) [S1 #registration-validation]
- "Any fields the Data Holder does not support MUST be ignored without error." "Fields data holders do not support are not persisted." [S1 #dynamic-client-registration-endpoints] [S1 #registration-management]
- Values in the SSA have priority over values in the request: "the value of a claim in the software statement MUST take precedence". [S8 section 2]
- `sector_identifier_uri`: OpenID Connect requires that the URL uses `https` and gives a JSON array of redirect URIs, and "The values of the registered redirect_uris MUST be included in the elements of the array." [S14 section 8.1] DSB guidance: the data holder gets that document during `POST` and `PUT`, and "if the sector_identifier_uri is unreachable, POST and PUT requests to the register endpoint should fail". [S18: 6828251436431] An older article with the same procedure has the mark "UNDER REVIEW ... cannot be relied on at this time". [S18: 360004324176]
- Additional validation, for example an outbound allowlist of data recipient endpoints, "MUST NOT prevent specified response times being met". [S1 #registration-validation]
- The CDS does not say that the data holder must check the Register status of the software product at registration. The Register issues an SSA only to an active data recipient for an active software product (lifecycle note, section 2.4), and the SSA is valid for 10 minutes. The duty to refuse a registration for a product that is not ACTIVE is **UNVERIFIED** (see section 8.3).

### 2.3 What the data holder takes from the SSA, and the response

- The response to a successful `POST` is HTTP 201. It "MUST be returned to the Data Recipient conforming to Section 3.2.1 of [DCR]". [S1 #registration-response] [S2 `cds_dcr.json`]
- `client_id` is required: the "dynamically generated identifier for the Software Product issued by the Data Holder". `client_id_issued_at` is optional. [S1 #registration-response]
- "the authorisation server MUST return all registered metadata about this client, including any fields provisioned by the authorisation server itself." [S1 #registration-response]
- "The Software Statement value MUST be returned unmodified. Client metadata elements used from the software statement MUST also be returned directly as top-level client metadata values in the registration response." [S1 #registration-response]
- "Any additional claims MUST be ignored and not returned on completion of the request." [S1 #registration-response]
- Fields that come from the SSA and are in the response schema: `client_name`, `client_description`, `client_uri`, `legal_entity_id`, `legal_entity_name`, `org_id`, `org_name`, `redirect_uris`, `sector_identifier_uri`, `logo_uri`, `tos_uri`, `policy_uri`, `jwks_uri`, `revocation_uri`, `recipient_base_uri`, `software_id`, `software_roles`, `scope`. [S1 #cdr-dynamic-client-registration-api_registrationproperties_properties]
- Required fields of the response schema (`RegistrationProperties`): `client_id`, `client_name`, `client_description`, `client_uri`, `org_id`, `org_name`, `redirect_uris`, `logo_uri`, `jwks_uri`, `token_endpoint_auth_method`, `token_endpoint_auth_signing_alg`, `grant_types`, `response_types`, `id_token_signed_response_alg`, `request_object_signing_alg`, `software_statement`, `software_id`, `scope`. [S2 `cds_dcr.json`]
- The `scope` claim of the SSA is the allowlist of scopes for the client. The data holder must validate each scope in an authorisation request against it. [S18: 6828251436431]
- The response has no RFC 7592 `registration_access_token` and no `registration_client_uri` in the CDS schema. [S2 `cds_dcr.json`] [S9 section 3]
- ADR 0001: the SSA claims reach the client handler through `ClientMeta.CustomAttributes` (proved). The validation of the SSA itself is not covered.

### 2.4 Errors

- "When an error condition occurs during a registration request, the response MUST be returned to the Accredited Data Recipient conforming to Section 3.2.2 of [DCR]." [S1 #registration-errors] That section gives HTTP 400, content type `application/json`, and the members `error` (required) and `error_description` (optional). [S8 section 3.2.2]
- The CDS schema permits these `error` values only: `invalid_redirect_uri`, `invalid_client_metadata`, `invalid_software_statement`, `unapproved_software_statement`. [S1 #cdr-dynamic-client-registration-api_registrationerror_properties]

| Code | Meaning in RFC 7591 section 3.2.2 |
|---|---|
| `invalid_redirect_uri` | "The value of one or more redirection URIs is invalid." |
| `invalid_client_metadata` | "The value of one of the client metadata fields is invalid and the server has rejected this request." |
| `invalid_software_statement` | "The software statement presented is invalid." |
| `unapproved_software_statement` | "The software statement presented is not approved for use by this authorization server." |

- "Duplicate registrations are not permitted so attempts to create a registration which already exists MUST return an HTTP 400 error." [S1 #registration-errors] The non-normative example uses `invalid_software_statement` with the description "Duplicate registrations for a given software_id are not valid". No normative text gives the code for this condition.
- "For those Registration errors which do not map to Section 3.2.2 of [DCR], HTTP Response Codes in the CDR Data Standards SHOULD be used." [S1 #registration-errors]
- The CDS gives no table from conditions (bad signature, expired SSA, wrong `aud`, unknown algorithm) to error codes. The mapping is **UNVERIFIED**.
- Management errors: 401 for an "unknown or invalid Client or invalid access token", with an optional `WWW-Authenticate` header; 403 when "The client does not have permission to read, update or delete the Client"; `PUT` can give 400 with the error schema; `DELETE` can give 405 when the data holder does not support it. [S2 `cds_dcr.json`] [S9 section 2.3]

### 2.5 One registration for each software product

- "Data Holders MUST NOT allow multiple active registrations for the same software_id." [S1 #registration-validation]
- "The software_id SHOULD be used as the primary external identifier for the client to prevent duplicate client registrations." [S1 #ssa-definition]
- One software product has one registration and one client ID for each data holder brand (lifecycle note, section 3.5). Each brand has its own issuer (section 1.3).
- The client ID of a REMOVED software product "should never be re-issued or re-enabled". [S18: 900001657906]

### 2.6 Registration management

- "Data Holders MUST expose Client Registration Management endpoints". [S1 #registration-management]
- The data recipient first gets an access token with `grant_type=client_credentials` and `scope=cdr:registration`, with `private_key_jwt`. [S1 #registration-management]
- "During registration management requests, Data Holders MUST validate that the scope of access tokens provided includes cdr:registration." [S1 #dynamic-client-registration-endpoints]
- `GET`, `PUT` and `DELETE` use MTLS and holder of key: the token must match the client certificate. [S2 `slate/source/includes/security/endpoints/_registration.md`]
- "Access tokens issued for consuming CDR Data SHOULD NOT be used for client registration maintenance and vice-versa." [S1 #registration-management] The admin and registration scopes "MUST never be included in the same access token as a CDR Data scope". [S1 #authorisation-scopes]
- "Registrations MUST only be updated via Registration PUT operations." "POST and PUT operations MUST accept the SSA payload." [S1 #put-operations] [S1 #registration-management]
- A `PUT` has two uses: the client changed its details on the Register, or a new version of the SSA exists. [S1 #put-operations]
- A data recipient does not have to send a `PUT` after it changes its Register data: "they are not required to update the same information with Data Holders". The data holder must also "do a slow poll to periodically update ADR metadata from the Register". [S18: 6828251436431]
- `legal_entity_id` and `legal_entity_name` are optional in the SSA, but the DSB says that they "will always be returned". Their mutability is under review in issue 679 (open). [S18: 6828251436431] [S19: 679]
- The CDS does not cite RFC 7592. Its normative reference for registration is RFC 7591 only. [S1 #normative-references] The management endpoints follow the shape of RFC 7592 (`GET`, `PUT`, `DELETE`, 204, 401, 403, 405), but they use an OAuth access token with a scope, not the RFC 7592 registration access token. [S9 sections 2.1 to 2.3]
- Register status: the data holder must clean up the registration of a REMOVED software product. A registration of an INACTIVE software product stays. Section 8.3 gives the full table. [S1 #data-holder-responsibilities]
- ADR 0001: proved (workaround `RegistrationAccessToken`: token active, for the client, bound to the client certificate, with the scope).

---

## 3. Authorisation flow

### 3.1 PAR is the only way to send a request

- "Data Holders MUST support Pushed Authorisation Requests (PAR) via the pushed authorisation endpoint according to [PAR]." [S1 #security-profile_request-object_data-holders]
- "Authorisation request data MUST only be accepted using PAR." "Data Holders MUST reject authorisation requests containing the request parameter." [S1 #security-profile_request-object_data-holders]
- "Request Object references SHALL NOT be supported in any mode of use other than [PAR]." [S1 #request-object-submission]
- The PAR request has the signed request object in the `request` form parameter, and the client assertion. The endpoint uses MTLS. [S1 #pushed-authorisation-endpoint] [S6 section 3]
- The client authentication rules of the token endpoint apply to the PAR endpoint. [S6 section 2]
- Success is HTTP 201 with `request_uri` and `expires_in`. [S6 section 2.2] [S1 #pushed-authorisation-endpoint]
- "The Request URI MUST expire between 10 seconds and 90 seconds." [S1 #request-object-submission] RFC 9126 leaves the lifetime to the server. [S6 section 2.2] The CDS non-normative example shows `"expires_in": 3600`, which is not in the permitted range (section 12, item 1).
- "Data Holders MUST reject the reuse of request_uri values." "The Request URI is intended to be a single-use reference". [S1 #security-profile_request-object_data-holders] [S1 #request-object-submission] RFC 9126 says only "SHOULD treat request_uri values as one-time use" and permits a repeat after a browser reload. [S6 section 4] The CDS is the stricter rule.
- "An expired request_uri MUST be rejected as invalid." [S6 section 4]
- A `request_uri` that a different client presents is a failed client check. The DSB suggests 401. [S18: 5729181240591]
- The authorisation request then has only `client_id` and `request_uri` in the query string. `response_type`, `scope` and `redirect_uri` are not necessary there, and the data holder must not refuse the request because they are absent. [S18: 6382507783183] [S18: 900004240446]
- ADR 0001: used (PAR and the signed request object are in the test provider). The 10 to 90 second lifetime and the one-time use are not covered.

### 3.2 Request object

- "Request Objects MUST be signed by Data Recipient Software Products as specified in section 8.6 of [FAPI-1.0-Advanced]." [S1 #request-object]
- "Data Holders MUST require the request object to contain an exp claim that has a lifetime of no longer than 60 minutes after the nbf claim". [S1 #security-profile_request-object_data-holders] [S3 section 5.2.2, item 13]
- The server "shall require the request object to contain an nbf claim that is no longer than 60 minutes in the past". [S3 section 5.2.2, item 17] From 16 September 2022, data holders "can validate and may reject" a request object without valid `nbf` and `exp`. [S18: 6382507783183]
- The server "shall require the aud claim in the request object to be, or to be an array containing, the OP's Issuer Identifier URL". [S3 section 5.2.2, item 15] The CDS non-normative example shows the data recipient URL as `aud`, which is not correct (section 12, item 1).
- The server "shall only use the parameters included in the signed request object". [S3 section 5.2.2, item 10]
- The server "shall require the redirect_uri in the authorization request" and "shall require the value of redirect_uri to exactly match one of the pre-registered redirect URIs". [S4 section 5.2.2, items 9 and 10] With PAR, the request object has the `redirect_uri`. [S18: 900004240446]
- `openid` "MUST be present on each authentication request". [S1 #oidc-scopes] With `openid`, the server "shall require the nonce parameter". [S4 section 5.2.2.2]
- CDS-specific members of `claims`: `sharing_duration` (section 4.3) and `cdr_arrangement_id` (section 5.5). They are at the top level of `claims`, next to `id_token` and `userinfo`. [S1 #request-object] ADR 0001: `sharing_duration` proved (workaround `RequestObjectClaims`, signed request objects only). `cdr_arrangement_id` in the request object is not covered.

### 3.3 Flow and `response_type`

- "Data Holders MUST support Authorization Code Flow." "Data Holders SHALL require the value of response_type described in [RFC6749] to be code." [S1 #baseline-security-provisions]
- "Until May 12th 2025, Data Holders MAY support OIDC Hybrid Flow". "No other flows are currently supported." [S1 #authentication-flows] The future dated obligations table says: "Data Holders SHALL only support Authorization Code Flow from May 12th 2025. Data Recipients SHALL only send authorisation requests using Authorization Code Flow from this date." [S1 #future-dated-obligations]
- Thus these hybrid-flow items are retired from 12 May 2025: `response_type` `code id_token`; the ID token from the authorisation endpoint; ID token encryption (`id_token_encrypted_response_alg`, `id_token_encrypted_response_enc` and the two discovery fields); `c_hash` and `s_hash`; `response_mode` `fragment`. [S1 #oidc-hybrid-flow-requirements] [S1 #registration-request-using-jwt] [S19: 666]
- The CDS 1.36.0 text and the DCR schema still contain these items (section 12, item 2). In July 2025 a data recipient reported that it saw no responses without JARM in the previous month. [S19: 649]
- "Only Confidential Clients SHALL be supported under this profile." [S1 #oidc-client-types]
- "Data Holders MUST reject token request with an authorization code ... if it has been previously used." [S1 #access-token] [S4 section 5.2.2, item 13]
- The CDS gives no lifetime for the authorisation code. RFC 6749 recommends a maximum of 10 minutes. [S13 section 4.1.2] When the data recipient does not exchange the code in time, the consumer gave an authorisation that no party can use; what the data holder does with it "is at the discretion of the DH". [S18: 360004665496]

### 3.4 JARM and `response_mode`

- With `response_type` `code`, FAPI requires "the response_mode value jwt", and the server "shall create JWT-secured authorization responses as specified in JARM". [S3 section 5.2.2, item 2] [S3 section 5.2.2.2]
- "Data Holders MUST support [JARM] in accordance with [FAPI-1.0-Advanced] section 5.2.2.2." [S1 #authorization-code-flow]
- The response JWT has `iss` (the issuer), `aud` (the client ID), `exp`, and the response parameters (`code`, `state`, or `error`). "A maximum JWT lifetime of 10 minutes is RECOMMENDED." [S5 section 2.1]
- "The response mode jwt is a shortcut ... The default for response type code is query.jwt". [S5 section 2.3.4] The CDS does not say which of `jwt`, `query.jwt`, `fragment.jwt` and `form_post.jwt` a data holder must support. The required set is **UNVERIFIED**; `jwt` is the value in the CDS examples.
- Signing: the data recipient registers `authorization_signed_response_alg`. "If authorization_signed_response_alg is omitted, the default algorithm is PS256." [S1 #authorization-code-flow] JARM itself gives RS256 as the default, which FAPI says clients and servers "should not use". [S5 section 3] [S3 section 8.6]
- Encryption is optional for the data holder: "Data Holders MAY support Authorization Response encryption", and "it is not recommended to use encryption in this case at this time". [S1 #authorization-code-flow]
- "If authorization_encrypted_response_alg is omitted from the registration request, authorisation response encryption SHALL NOT be performed." [S1 #authorization-code-flow] The future dated obligations table gives the date: "From November 10th 2025, Data Holders SHALL NOT perform authorization response encryption unless requested by the ADR". [S1 #future-dated-obligations] This text came in CDS 1.35.0 from issue 650. Before it, a data holder could force encryption. [S19: 650]
- A data holder that does not support encryption omits the two encryption discovery fields or gives empty arrays. [S1 #authorization-code-flow] If it supports encryption, it supports at least one of `RSA-OAEP`, `RSA-OAEP-256` and at least one of `A256GCM`, `A128CBC-HS256`. [S1 #jarm-response-encryption-considerations]
- Errors also go in the JARM JWT. The `error` values are those of RFC 6749 section 4.1.2.1. Issue 649 (open) records that some data holders return other values, and suggests `access_denied` when the consumer cannot authenticate or cancels. [S19: 649] [S13 section 4.1.2.1] When the consumer cancels, the data holder "should respond to the ADR (client) with an error outcome". [S18: 5729181240591]
- ADR 0001: JARM signed responses proved (`WithJARM`: `iss`, `aud`, `exp`, `code`, `state`). JARM encryption is not covered.

### 3.5 PKCE

- "Data Holders MUST support [RFC9126] (PAR) using [PKCE] ([RFC7636]) with S256 as the code challenge method". [S1 #security-profile_request-object_data-holders] FAPI: the server "shall require PAR requests, if supported, to use PKCE (RFC7636) with S256". [S3 section 5.2.2, item 18]
- The challenge is `BASE64URL-ENCODE(SHA256(ASCII(code_verifier)))`. The data recipient sends `code_verifier` in the token request. [S7 sections 4.2 and 4.5]
- ADR 0001: used (PKCE is in the test provider).

### 3.6 ID token

- With the code flow, the token endpoint returns the only ID token: "ID Tokens MUST be signed when returned to a Data Recipient Software Product from the Token endpoint." [S1 #authorization-code-flow-requirements]
- "ID Tokens MUST be signed by Data Holders as specified in section 8.6 of [FAPI-1.0-Advanced]." [S1 #baseline-id-token-requirements]
- "ID Tokens MUST NOT contain any Personal Information (PI) claims." [S1 #baseline-id-token-requirements] Thus `name`, `given_name` and `family_name` go only in the UserInfo response.
- Encryption: the ID token encryption claims "SHALL NOT be used for Authorization Code Flow". [S1 #registration-request-using-jwt] The earlier "MUST be signed and encrypted" rule applies only to the hybrid flow. [S1 #oidc-hybrid-flow-requirements]
- `c_hash` and `s_hash`: "The following requirements apply to the OIDC Hybrid Flow". They do not apply to the code flow. [S1 #hashing-value-for-state-and-authorisation-code] FAPI requires `s_hash` only "if the response_type value code id_token is used". [S3 section 5.2.2.1]
- `acr`: "MUST contain a valid ordinal LoA value" (section 6.2). [S1 #claims]
- `auth_time`: "MUST be returned by the Data Holder in the ID Token when the Data Recipient Software Product has requested it as an essential claim". "It SHOULD NOT be returned via the UserInfo endpoint." [S1 #claims]
- `sharing_expires_at` and `refresh_token_expires_at` were ID token claims of older CDS versions. CDS 1.30.0 (24/04/2024) removed them. They are not in CDS 1.36.0. [S19: 543] A data recipient now gets the expiry of the sharing arrangement from the `exp` of the introspection response. [S1 #requesting-sharing-duration]
- `cdr_arrangement_id` is not an ID token claim. It is a field of the token response and of the introspection response (section 5.2). [S1 #cdr-arrangement-id]

### 3.7 Scopes and data clusters

- "The primary mechanism for capturing consent will be scopes and claims under [OIDC]." [S1 #consent]
- The data holder must support `openid` and `profile`, and the CDR data scopes. [S1 #oidc-scopes]

Banking and common scopes, with the data language that both parties "MUST use" (rule 8.11(1)(d) of the CDR Rules requires this standard): [S1 #authorisation-scopes] [S1 #banking-and-non-bank-lending-language] [S1 #customer-language-common] [S1 #data-language-standards-common]

| Scope | Data cluster language | Permission language |
|---|---|---|
| `bank:accounts.basic:read` | Account name, type and balance | Name of account; Type of account; Account balance |
| `bank:accounts.detail:read` | Account numbers and features | Account number; Interest rates; Fees; Discounts; Account terms; Account mail address |
| `bank:accounts.detail:read` (when the basic cluster is not shown) | Account balance and details | The basic and detail permissions together |
| `bank:transactions:read` | Transaction details | Incoming and outgoing transactions; Amounts; Dates; Descriptions of transactions; Who you have sent money to and received money from |
| `bank:regular_payments:read` | Direct debits and scheduled payments | Direct debits; Scheduled payments |
| `bank:payees:read` | Saved payees | Names and details of accounts you have saved |
| `common:customer.basic:read` (individual) | Name and occupation | Name; Occupation |
| `common:customer.detail:read` (individual) | Contact details | Phone; Email address; Mail address; Residential address |
| `common:customer.detail:read` (individual, when the basic cluster is not shown) | Name, occupation, contact details | The basic and detail permissions together |
| `common:customer.basic:read` (business) | Organisation profile | Agent name and role; Organisation name; Organisation numbers (ABN or ACN); Charity status; Establishment date; Industry; Organisation type; Country of registration |
| `common:customer.detail:read` (business) | Organisation contact details | Organisation address; Mail address; Phone number |
| `profile` | Name | Full name and title(s) |
| OIDC contact claims (optional) | Contact details | Phone number; Email address; Mail address |

- "Data language standards MUST be used when CDR data is being requested, reviewed, or access to such data is withdrawn." [S1 #data-language-standards-common]
- When a request has the detail scope without the basic scope, both parties "MUST" use the merged language (the rows "when the basic cluster is not shown"). [S1 #data-language-standards-common]
- The "detail" scopes are "effectively additional authorisation" to the basic scopes. [S1 #authorisation-scopes]
- `cdr:registration`, `admin:metrics.basic:read` and `admin:metadata:update` are not consent scopes. [S1 #authorisation-scopes]
- `offline_access` "is not relevant to the CDR standards". [S18: 900003036086]

### 3.8 Claims

- Claims that the data holder "MUST" support: `sub`, `acr`, `auth_time`, `name`, `given_name`, `family_name`, `updated_at`. [S1 #claims]
- Claims that it "MAY" support: `email`, `email_verified`, `phone_number`, `phone_number_verified`, `address`. [S1 #claims]
- "Other [OIDC] Standard Claims MUST be ignored and not authorised." [S1 #claims]
- "Data Holders SHOULD explicitly capture Claims requested by the Data Recipient." [S1 #claims]
- For a consumer that is not an individual, the profile claims are those of the authenticated user, not of the organisation. [S1 #claims]
- The UserInfo response always has `sub`. [S14 section 5.3.2] The data returned by UserInfo is CDR data: a call from a data recipient that is not active "should fail, and receive a 401 response". [S18: 5729181240591]

### 3.9 `sub` and pairwise identifiers

- "The identifier for an authenticated end-user (subject) MUST be passed in the sub claim of an ID Token and UserInfo response". [S1 #sub-claim]
- "The Data Holder MUST generate the sub value as a Pairwise Pseudonymous Identifier (PPID) as described in section 8 of [OIDC]." It "MUST be unique per customer". [S1 #sub-claim]
- "The Data Holder MUST support the sector_identifier_uri in PPID generation according to [OIDC] if this field was supplied by the client during registration." [S1 #sub-claim]
- "It is RECOMMENDED that the sub value is generated as a version 4 Universally Unique Identifier (UUID)". [S1 #sub-claim]
- OIDC rule: with a `sector_identifier_uri`, "the host component of that URL is used as the Sector Identifier". With none, the sector identifier is the host of the registered `redirect_uri`, and "If there are multiple hostnames in the registered redirect_uris, the Client MUST register a sector_identifier_uri." [S14 section 8.1]
- The algorithm must be deterministic, must give different values for different sector identifiers, and must not be reversible by other parties. One permitted method is a stored GUID for each pair of sector identifier and local account ID. [S14 section 8.1]
- The PPID identifies the authenticated user, for example a nominated representative, not the consumer that the user acts for. "The PPID value must persist (be the same) for each arrangement" of the same user and sector identifier. [S18: 5505351487631]
- Open conflict: two software products with the same `sector_identifier_uri` host get the same `sub`, but resource IDs must be different for each software product. Issue 480 is open. [S19: 480]
- ADR 0001: not covered.

---

## 4. Tokens

### 4.1 Client authentication at the token, PAR, introspection and revocation endpoints

- "Data Holders and the CDR Register MUST authenticate Data Recipient Software Products using the Private Key JWT Client Authentication method." [S1 #client-authentication]
- "the PKI Mutual TLS OAuth Client Authentication Method SHALL NOT be supported". MTLS is for transport security and holder of key only. [S1 #client-authentication]
- Assertion claims: `iss` and `sub` (the client ID), `aud`, `jti`, `exp` are required. `iat` is optional. [S1 #private-key-jwt-client-authentication]
- `aud`: "the authorisation server MUST accept its Issuer Identifier, Token Endpoint URL, or the URI of the endpoint being invoked". [S1 #private-key-jwt-client-authentication]
- `jti`: "These tokens MUST only be used once." [S1 #private-key-jwt-client-authentication]
- `client_id` form parameter: optional from 13 November 2023. [S1 #private-key-jwt-client-authentication] When it is present, the data holder checks it against the `sub` of the assertion. [S18: 6382507783183]
- `client_id` is "the ID issued to the Data Recipient Software Product by the Data Holder upon successful dynamic client registration". `grant_type` is `authorization_code`, `client_credentials` or `refresh_token`. [S1 #data-recipients-calling-data-holders]
- The data holder verifies the assertion with the keys at the `jwks_uri` of the client. "Data Recipient Software Product public keys MUST only be obtained from the URI registered with the CDR Register." [S1 #json-web-key-sets]
- The CDS gives no maximum lifetime for a client assertion. This is **UNVERIFIED** as a requirement.
- A missing or wrong `sub`: the DSB suggests 400 with `invalid_client`. [S18: 5729181240591]
- ADR 0001: used (`private_key_jwt` is in the test provider).

### 4.2 Access token

- "An Access Token MUST expire between 2 minutes to 10 minutes after the Data Holder issues it (at the discretion of the Data Holder)." [S1 #access-token]
- "The Data Holder MUST indicate the lifetime in seconds of the access token in the expires_in field". "The expiry time for issued access tokens and refresh tokens MUST be deterministic". [S1 #token-expiry]
- The server "shall only issue sender-constrained access tokens" and "shall support MTLS as mechanism for constraining the legitimate senders of access tokens". [S3 section 5.2.2, items 5 and 6]
- "[MTLS] MUST be supported as a Holder of Key (HoK) Mechanism." "resource requests MUST be validated to ensure the client certificate and access token match." [S1 #holder-of-key-mechanism]
- Binding: the token has the SHA-256 thumbprint of the client certificate, as `cnf.x5t#S256` in a JWT or as introspection data. [S10 sections 3 and 3.1] When the certificate does not match, the resource server rejects the request "using an HTTP 401 status code and the invalid_token error code". [S10 section 3]
- The CDS does not give the format of the access token (opaque or JWT). FAPI: clients "treat access tokens as opaque strings". [S4 section 5.2.2, notes]
- ADR 0001: used (certificate-bound tokens are in the test provider).

### 4.3 Refresh token and `sharing_duration`

- "Refresh Tokens MUST be supported by Data Holders in accordance with section 12 of [OIDC]." [S1 #refresh-token]
- Data holders "MUST NOT cycle refresh tokens (rotation)". [S1 #refresh-token]
- "From May 12th 2025: MUST issue Refresh Tokens with an exp value equal to the sharing duration authorised by the consumer." Before that date it was a "SHOULD". [S1 #refresh-token] [S19: 667]

`sharing_duration` rules: [S1 #requesting-sharing-duration]

| Value | Rule |
|---|---|
| A number of seconds | "The value of the sharing_duration parameter will contain the requested duration for sharing, in seconds." |
| More than one year | "a duration of one year will be assumed". One year is "365, 24 hour days (or 31,536,000 seconds)". |
| More than 0, and 24 hours or less | "one-time collection will be assumed, and a Refresh Token SHOULD be provided by the Data Holder". The data recipient "MUST call the Data Holder's revocation endpoint after successful collection". |
| 0 or absent | "once off access will be assumed and only an Access Token (without a Refresh Token) will be provided" |
| Negative | "the authorisation SHOULD fail" |

- The CDR Rules give the same limit: an authorisation expires at the latest at "the end of the period of 12 months after the authorisation was given". [S16 rule 4.26(1)(e)] A consent of a business consumer can be for 7 years, but the authorisation is for 12 months at most. [S16 rule 4.23(1), note]
- With `sharing_duration` 0, the arrangement and the consent "are immediately expired upon creation". The data holder returns an error when the data recipient tries to revoke that arrangement, and that consent cannot be amended. [S18: 5729181240591] [S18: 4886150993807]
- ADR 0001: proved (workaround `RefreshTokenForSharingDuration`: an expiry for each grant, and no refresh token when `sharing_duration` is 0 or absent; the provider must not use `WithRefreshTokenRotation`). The one-year cap and the 24-hour rule are not covered.

### 4.4 Token endpoint response

- Fields: `access_token`, `token_type`, `expires_in`, `refresh_token` (when one is issued), `id_token`, and `cdr_arrangement_id`. [S13 section 5.1] [S14 section 3.1.3.3] [S1 #cdr-arrangement-id]
- "The Data Holder MUST provide the CDR Arrangement ID as the claim cdr_arrangement_id in the Token endpoint response and Token Introspection endpoint response." [S1 #cdr-arrangement-id]
- `token_type` is "REQUIRED" in RFC 6749, and FAPI requires responses "that conform to Section 4.1.4 of RFC6749". [S13 section 5.1] [S4 section 5.2.2, item 14] The CDS examples omit it (section 12, item 1).
- A refresh response is the same, "except that it might not contain an id_token". [S14 section 12.2]
- Errors use RFC 6749 section 5.2. For a refresh token that is "invalid, expired, revoked", that section gives `invalid_grant` with HTTP 400. [S13 section 5.2] A support article gives "a 401 Unauthorised Error with a corresponding oAuth error response" for a call after the consent is withdrawn, and does not say clearly which endpoint it means (section 12, item 9). [S18: 5729181240591]
- ADR 0001: `cdr_arrangement_id` in the token response is not covered.

### 4.5 Introspection

- "Introspection of Refresh Tokens MUST be supported." "Introspection of Access Tokens and ID Tokens MUST NOT be supported." [S1 #introspection-endpoint]
- For an active token, the response "SHALL include, at least": `active`, `exp`, `scope`, `cdr_arrangement_id`. [S1 #introspection-endpoint]
- It "MAY include claims defined in Section 2.2 of [RFC7662] but username SHALL NOT be allowed". [S1 #introspection-endpoint]
- For a token that is not active, RFC 7662 applies: the response has `active` set to `false`, and the server "SHOULD NOT include any additional information". [S11 section 2.2] The DSB confirms that the CDS field list does not replace this rule. [S18: 900004618103]
- The `exp` of the response is the expiry of the sharing arrangement. [S1 #requesting-sharing-duration]
- The CDS does not say what the response is when a client sends an access token. `active: false` agrees with RFC 7662 section 2.2 ("the protected resource is not allowed to introspect this particular token"), but the CDS does not confirm it: **UNVERIFIED**.
- ADR 0001: not covered.

### 4.6 Token revocation

- "Data Holders MUST implement a Token Revocation endpoint as described in section 2 of [RFC7009]." "Revocation of Refresh Tokens and Access Tokens MUST be supported." [S1 #token-revocation-endpoint]
- The endpoint is for token clean-up. It is not the way to withdraw a consent; the CDR arrangement revocation endpoint is (section 5.3). [S1 #token-revocation-endpoint] [S18: 900003387786]
- The server checks that "the token was issued to the client making the revocation request". The response is HTTP 200 also for a token that is not valid. [S12 sections 2.1 and 2.2]
- When the token is a refresh token, the server "SHOULD also invalidate all access tokens based on the same authorization grant". [S12 section 2.1]
- After a revocation of a refresh token for a one-time collection (section 4.3), the arrangement has no usable token. The CDS does not say if the data holder then records the authorisation as expired. This is **UNVERIFIED**.
- An access token that the data recipient revoked must fail in a data request. A new access token from the refresh token is valid. [S18: 6828747049359]
- ADR 0001: not covered.

---

## 5. CDR arrangements

### 5.1 Definition

- "The CDR Arrangement ID is a unique string representing a consent arrangement between a Data Recipient Software Product and Data Holder for a given consumer." [S1 #cdr-arrangement-id]
- The CDR Rules use the word "authorisation" for the data holder side: the consumer authorises the data holder to disclose CDR data to an accredited person. [S16 rules 4.5 and 4.23] The data recipient side is the "consent".
- "Technically consents cannot be amended. Consents can be revoked and a new consent can be created, linked to a previous revoked consent via a common cdr_arrangement_id." [S18: 6828747049359]

### 5.2 `cdr_arrangement_id` rules

- It "MUST be unique to a Data Holder". It "MUST be non-guessable and MUST NOT identify a consumer". It "SHOULD be generated using an algorithm that reduces the chances of collision". [S1 #cdr-arrangement-id]
- It "MUST be static across consents within the one sharing arrangement (e.g. across consent renewal and re-authorisation)". [S1 #cdr-arrangement-id]
- The data recipient gets it from the token response or the introspection response "at any point post-consent". [S1 #obtaining-a-cdr-arrangement-id]

### 5.3 Revocation by the data recipient at the data holder endpoint

- "Data Holders and Data Recipient Software Products MUST implement a CDR Arrangement Revocation endpoint". [S1 #cdr-arrangement-revocation-endpoint]
- Request: `POST`, `application/x-www-form-urlencoded`, with the client assertion parameters and `cdr_arrangement_id`. [S1 #cdr-arrangement-revocation-endpoint]
- "Data Holders MUST only support CDR Arrangement Form Parameter method." The `cdr_arrangement_jwt` form is not for the data holder endpoint. It is for calls to the data recipient endpoint (section 5.4). [S1 #cdr-arrangement-revocation-endpoint]
- "Consent revocation MUST also revoke associated refresh and/or access tokens." [S1 #cdr-arrangement-revocation-endpoint]
- "If the cdr_arrangement_id is not related to the client making the call it MUST be rejected." [S1 #cdr-arrangement-revocation-endpoint]
- Success is "204 No Content". An arrangement ID that is not valid or not found gives "422 Unprocessable Entity", and "The server MUST respond with Invalid Consent Arrangement": `urn:au-cds:error:cds-all:Authorisation/InvalidArrangement`, in the CDS error structure. The error detail is the arrangement ID. [S1 #cdr-arrangement-revocation-endpoint] [S1 #422-unprocessable-entity-errors]
- The 422 applies to three cases: the arrangement was revoked before, it is of a different client, or it does not exist. [S18: 5729181240591]
- The revocation "must occur synchronously with returning a 204 response". [S18: 5729181240591]
- The authorisation expires when the data holder gets the notice that the consent was withdrawn. [S16 rule 4.26(1)(d)]
- After the revocation, "The DH's solution must prevent any data disclosure against the cdr_arrangement_id." [S18: 4412301531663]
- The data recipient must use this endpoint when a consent is withdrawn or expires, except when the data holder started the withdrawal, when the consent expires at its natural time, or when a Register status change of the data holder invalidates it. [S1 #cdr-arrangement-revocation-endpoint]
- ADR 0001: not covered.

### 5.4 The data holder calls the data recipient when the consumer withdraws

- "If an authorisation to disclose particular CDR data to an accredited person is withdrawn or otherwise expires, the data holder must notify the accredited person in accordance with the data standards." This is a civil penalty provision. [S16 rule 4.26A]
- The data holder "MUST use the Data Recipient Software Product's CDR Arrangement Revocation endpoint ... to notify the Data Recipient Software Product when an authorisation is withdrawn or otherwise expires". Exceptions: the data recipient started the withdrawal; the authorisation expired at its natural time; a Register status change of the data recipient or the software product invalidated it. [S1 #cdr-arrangement-revocation-endpoint]
- Path: `<RecipientBaseUri>/arrangements/revoke`. `recipient_base_uri` comes from the SSA. Transport: TLS, with a public CA certificate, not MTLS. [S1 #cdr-arrangement-revocation-endpoint] [S1 #participant-endpoints]
- Authentication: Self-signed JWT Client Authentication. The JWT is the Bearer token in the `Authorization` header. `iss` and `sub` are "the Data Holder Brand ID as issued by CDR Register". `jti` and `exp` are required, and the JWT is for one use. [S1 #self-signed-jwt-client-authentication] [S1 #data-holders-calling-data-recipients]
- `aud`: "From July 31st 2022: The Resource Path for the endpoint being accessed MUST be used." The example shows the full URL of the revocation endpoint. [S1 #self-signed-jwt-client-authentication] A support article says that `aud` is the `RecipientBaseUri` and "not the full path" (section 12, item 6).
- The data recipient verifies the JWT with the data holder keys from the `jwksEndpoint` that the Register publishes for the brand. [S1 #participant-endpoints]
- Body: "From July 31st 2022, Data Holders MUST send the cdr_arrangement_id using the CDR Arrangement JWT method." `cdr_arrangement_jwt` is a signed JWT with `cdr_arrangement_id`; it "SHOULD also include" the self-signed JWT claims (`iss`, `sub`, `aud`, `jti`, `exp`). [S1 #cdr-arrangement-revocation-endpoint]
- "Data Holders MAY additionally send a duplicate of the cdr_arrangement_id as a form parameter." From 15 November 2022, the data recipient must check that the two values are equal. [S1 #cdr-arrangement-revocation-endpoint]
- Retry: the CDS gives no rule. DSB guidance: "reasonable effort" at the discretion of the data holder; "waiting a few seconds upon first failure would be suitable"; an exhaustive retry is not "necessary or practical". [S18: 360004056455] [S19: 367] The lifecycle note, section 4.3, gives the ACCC backoff patterns.
- The data holder gives effect to the withdrawal at once. The call to the data recipient is a notice, not a condition. A withdrawal on the dashboard expires the authorisation "when the authorisation was withdrawn". [S16 rule 4.26(1)(b)]
- ADR 0001: not covered (this is an outbound call, not an authorisation server function).

### 5.5 Amendment through PAR with an existing `cdr_arrangement_id`

The first project puts amendment out of scope. These rules still apply to the data holder, because the claim can arrive in a request object:

- "Data Holders MUST support the cdr_arrangement_id claim provided in the Request Object." They "MUST treat the request under the Amending Authorisation Standards if the cdr_arrangement_id claim is provided". [S1 #specifying-an-existing-arrangement]
- "Data Holders MUST revoke existing refresh tokens and access tokens when a cdr_arrangement_id is provided in the Request Object but only after successful authorisation." [S1 #request-object-submission]
- "If the cdr_arrangement_id is not related to the consumer being authenticated it MUST be rejected." "If the cdr_arrangement_id is not recognised by to the Data Holder it MUST be rejected." [S1 #request-object-submission] The DSB suggests a redirect with the error `access_denied`. [S18: 5729181240591]
- The new consent keeps the same `cdr_arrangement_id`. [S18: 360004559675]
- The data holder invites an amendment only after a notice under rule 4.18C or 4.20S. An amendment in a different way "is of no effect". [S16 rule 4.22A]
- UI rules for an amendment: pre-select the accounts of the previous authorisation ("MUST"), and show which dataset is added or which duration changes ("MUST"). [S1 #amending-authorisation-standards]

### 5.6 How many active arrangements

- The CDS has no rule that limits a consumer to one active arrangement with one software product. Concurrent consents are permitted: a data recipient can make a new consent without an arrangement ID, and each consent then has its own arrangement ID. [S18: 4407560834959] [S18: 360004023156]
- The "one" rules are these: one active registration for each `software_id` (section 2.5); and one current set of tokens for each arrangement, because an amendment revokes the old tokens (section 5.5).

---

## 6. Consumer authentication

### 6.1 Authentication flows and their dates

- The Authentication Schedule (new in CDS 1.35.0, Decision 369) defines three approaches: Redirect to App, decoupled authentication and the fallback flow. [S1 #authentication-schedule] [S1 #changelog-and-archives]
- Redirect to App: "If a data holder provides an app, the data holder MUST implement Redirect to App". The obligation date is 10 May 2027; a data holder "MAY" implement it earlier. [S1 #authentication-schedule_redirect-to-app_data-holders] [S1 #future-dated-obligations]
- Fallback: where Redirect to App cannot be used and decoupled authentication is not supported, "data holders MUST continue to support the Redirect to Web with OTP flow". [S1 #fallback-authentication-flows]
- A mock data holder with no app thus needs only "Redirect to Web with OTP".

### 6.2 Levels of assurance

- The `acr` claim has one LoA value. [S1 #levels-of-assurance-loas]

| `acr` value | Rule |
|---|---|
| `urn:cds.au:cdr:2` | The authenticators "MUST achieve Single Factor Authentication" |
| `urn:cds.au:cdr:3` | The authenticators "MUST achieve Multi-Factor Authentication" |
| `urn:cds.au:cdr:4` | The authenticators "MUST conform with the Authentication Level AL3 rules" of the Digital ID Accreditation Data Standards |

- "READ operations SHALL only be allowed where at least an LoA of 2 has been achieved during the establishment of consent." All CDR data APIs are read operations. [S1 #levels-of-assurance-loas]
- Data holders "SHOULD implement Multi-Factor Authentication for user access to CDR data containing personal information". [S1 #baseline-credential-requirements]
- When the data recipient requests LoA 3 as essential and the data holder cannot reach it, "the request should be rejected". [S18: 6140553082255]

### 6.3 Redirect to Web with OTP

User identifier: [S1 #user-identifiers]
- The data holder "MUST request a user identifier that can uniquely identify the customer" and "that is already known by the customer".

One-time password: [S1 #one-time-password-credential-requirements]
- "The data holder MUST provide a one-time password (OTP) to the customer through an existing channel or mechanism that the customer can then enter into the redirected page."
- "The delivery mechanism for the OTP is at the discretion of the data holder, but MUST align to existing and preferred channels for the customer."
- "The provided OTP MUST only be used for authentication for CDR based sharing and MUST NOT be usable for the authorisation of other transactions or actions."
- "The provided OTP MUST be invalidated after a period of time at the discretion of the data holder." The period "SHOULD facilitate enough time for the customer to reasonably complete the authorisation process".
- "The provided OTP MUST be numeric digits and be between 4 and 6 digits in length."
- The algorithm "SHOULD incorporate a level of pseudo-randomness appropriate for the use case".
- The data holder "SHOULD implement additional controls to minimise the risk of interception of the OTP" and "of enumeration attacks via the redirect page".

Passwords and what the data holder must not ask for:
- Authenticators "MUST NOT permit Memorised Secrets ... to achieve LoA 2 (as a single factor of authentication only)". Thus a password alone is not permitted. [S1 #restricted-credentials]
- Data holders "MUST ONLY support authenticator types as permitted by the Authentication levels (AL) requirements defined in [DigitalID-Accreditation]". [S1 #baseline-credential-requirements] This note did not read that instrument: the list of permitted authenticator types is **UNVERIFIED**.
- Both parties "MUST state in consumer-facing interactions and communications that third parties do not need consumer passwords to access CDR data". [S1 #consumer-experience_common-authentication-standards]
- "Data holders MUST NOT include forgotten details links in redirect screens." [S1 #consumer-experience_redirect-to-web-with-one-time-password]
- "Data holders MUST communicate the expiry period of the OTP to the consumer in the authentication flow." [S1 #consumer-experience_redirect-to-web-with-one-time-password]
- "Data holders MUST NOT introduce unwarranted friction into the authentication process." Unwarranted friction includes requirements or requests for information "beyond normal data holder practices", offers of other services, and references to other documents. [S1 #consumer-experience_common-authentication-standards] [S16 rule 4.24]

DSB guidance on OTP: [S18: 6140553082255]
- OTPs "do not have to be sent by SMS or by phone".
- The number of retries is not limited by the CDS. "It is good practice to limit the number of failed retries". The DSB recommends that the consumer can ask for a new OTP. [S17 Redirect to Web with One Time Password]
- A number that the consumer supplies in the flow is not acceptable for delivery unless the data holder validates it against a number on file.
- The DSB recommends that the data holder does not send the OTP to an email address that is also the customer ID.
- The article says: "The standards currently do not support Security App or App2App redirect." CDS 1.35.0 and later contradict this (section 12, item 7).

Limits on a mock: the OTP rules name a channel that is "existing" and "preferred" for the customer. A mock has no real channel. No source says how a test system delivers an OTP. The rules that a mock can keep without a real channel are: numeric, 4 to 6 digits, an expiry that the flow shows, use for CDR only, no password, and no forgotten-details link.

---

## 7. Consent and authorisation UI rules for the data holder

### 7.1 Legal basis

- The data holder must ask the consumer to authorise the disclosure "in accordance with Division 4.4" and "in accordance with the data standards". [S16 rule 4.5]
- Its processes must "accord with the data standards" and, "having regard to any consumer experience guidelines developed by the Data Standards Body, be as easy to understand as practicable". [S16 rule 4.22] Thus the CX Standards bind, and the CX Guidelines are material that the data holder must consider.
- A consumer is eligible in banking only when the account "can be accessed online". [S16 Schedule 3, clause 2.1]
- The data holder can refuse to ask for an authorisation or to disclose data in the cases of rule 4.7, for example to prevent harm or abuse, or for an account that is blocked or suspended. It must tell the accredited person "in accordance with the data standards". [S16 rule 4.7]

### 7.2 What the data holder must show (rule 4.23)

When it asks for an authorisation, the data holder must give the consumer: [S16 rule 4.23(1)]

| Item | Rule |
|---|---|
| The name of the accredited person that made the request | 4.23(1)(a) |
| The period of time to which the CDR data relates | 4.23(1)(b) |
| The types of CDR data | 4.23(1)(c) |
| Whether the disclosure is on a single occasion or over a period of not more than 12 months | 4.23(1)(d) |
| For a period, what the period is | 4.23(1)(e) |
| A statement that the authorisation can be withdrawn at any time | 4.23(1)(f) |
| Instructions for how to withdraw | 4.23(1)(g) |

- The data holder must also give "any information that the Register of Accredited Persons holds in relation to the accredited person that is specified as information for the purposes of this rule". [S16 rule 4.23(2)] Which Register fields are specified is **UNVERIFIED**.
- CX Guidelines: rules 4.23(1)(a) and (2) "require data holders to refer to the accredited person's name using the legal entity name held in the Register". Data holders "should use the Brand Name of the data recipient" in authentication screens and OTP delivery, and "should not present the Software Product Name in relation to these processes". [S17 Authorisation to disclose] [S17 Redirect to Web with One Time Password]
- When `sharing_duration` is 0 or absent, or less than 24 hours, the data holder treats the request as a "single occasion" under rule 4.23(1)(d). [S17 Authorisation to disclose]

### 7.3 What the data holder must not do (rule 4.24)

The data holder must not: add requirements beyond the standards and the rules; give or ask for more information than they specify; offer other services; "include or refer to other documents". [S16 rule 4.24]

### 7.4 CX Standards for the authorisation flow

| Area | Standard | Source |
|---|---|---|
| Account selection | "Data holders MUST allow the consumer to select which of their accounts to share data from if the data request includes account-specific data and if there are multiple accounts available." The step "MAY" be omitted when no requested data is specific to an account. | [S1 #authorisation-standards] |
| Account confirm | "Data holders MUST show which accounts the data is being shared from prior to confirming authorisation if the data request includes account-specific data." | [S1 #authorisation-standards] |
| Selection functions | Search, sort, filter and other controls are a "MAY". They "MUST NOT introduce unwarranted friction". | [S1 #authorisation-standards] |
| Profile selection | "MAY", before account selection, when one identifier gives access to more than one customer profile. | [S1 #authorisation-standards] |
| Pending status | When an account needs more approvals, the data holder "MUST indicate this to the user visually and MUST provide an explanation". | [S1 #authorisation-standards] |
| Unavailable accounts | The data holder "SHOULD show these unavailable accounts in the account-selection step" and "SHOULD communicate why". | [S1 #authorisation-standards] |
| Data language | The cluster and permission language of section 3.7 is a "MUST". | [S1 #data-language-standards-common] |
| Accessibility | Both parties "MUST seek to" comply with WCAG 1.4, 2.1, 2.5, 3.1 and 3.3 in the Consent Model. | [S1 #accessibility-standards] |
| Joint account alert | "Data holders MUST alert a joint account holder where an action they are about to perform may result in the other joint account holder(s) being notified." | [S1 #consumer-experience_notification-standards] |

- CX Guidelines (not binding): the account selection step can satisfy "Account confirm" alone; with one account, the data holder can omit selection and show the account on the confirmation page; the data holder should send the consumer back to the data recipient after the final action; a user with valid credentials who cannot share data should still authenticate and then see why. [S17 Authorisation to disclose] [S17 Redirect to Web with One Time Password]

### 7.5 Withdrawal

- The consumer can withdraw "at any time" on the consumer dashboard, or with "a simple alternative method of communication". For the alternative method, the data holder must give effect "as soon as practicable, and in any case within 2 business days". [S16 rule 4.25]
- "As part of the withdrawal process, the data holder MUST advise the consumer to review the consequences of withdrawal with the Data Recipient before they stop sharing their data." [S1 #withdrawal-standards]
- "As part of the withdrawal process, the data holder MUST inform the consumer about the handling of redundant data and the right to delete." [S1 #withdrawal-standards]
- After a withdrawal, the data holder notifies the data recipient (section 5.4). [S16 rule 4.26A]

### 7.6 When an authorisation expires

- An authorisation expires at the earliest of these events: withdrawal on the dashboard (at once); withdrawal by the alternative method (when given effect, or after 2 business days); the consumer is no longer eligible; the data holder gets the notice that the consent was withdrawn (section 5.3); 12 months after the consumer gave it; for a single occasion, after the data is disclosed; for a period, the end of the period. [S16 rule 4.26(1)]
- When an accreditation is revoked or surrendered, all authorisations to that accredited person expire "when the data holder is notified of the revocation or surrender". [S16 rule 4.26(2)]

### 7.7 Consumer dashboard (list of obligations; the first project defers the dashboard)

- A banking data holder must provide a dashboard to each eligible consumer for whom it gets a consumer data request. [S16 rule 1.15(1)] [S16 Schedule 3, clause 2.3]
- The dashboard must: let the consumer manage authorisations (1.15(1)(a)); show the details of each authorisation (1.15(1)(b)); show information that the data standards and the Register specify (1.15(1)(ba), (bb)); let the consumer withdraw at any time (1.15(1)(c)); show the withdrawal message of the data standards (1.15(1)(d)); be no more complicated than the process to give the authorisation (1.15(1)(e)); be prominent and easy to find (1.15(1)(f)). [S16 rule 1.15(1)]
- Details for each authorisation: the CDR data; when it was given; its period; when it will expire or when it expired; what was disclosed (rule 7.9); each amendment (from 1 July 2024). [S16 rule 1.15(3), (3A)]
- After each disclosure, the data holder updates the dashboard with what was disclosed, when, and to which accredited data recipient. [S16 rule 7.9(1)]
- The data holder updates the dashboard "as soon as practicable" after the information changes. [S16 rule 4.27]
- CX Standards for the dashboard: show each amendment; "Data holders MUST advise consumers to check with the relevant data recipient for information about how their data may be handled." [S1 #consumer-experience_dashboard-standards_data-holder-dashboards]
- Secondary users, nominated representatives and joint accounts add more dashboard duties. [S16 rules 1.15(2A), 1.15(5), 4A.13]

### 7.8 Notifications and records

- To the data recipient: the notice of a withdrawn or expired authorisation (rule 4.26A; section 5.4).
- To account holders: notices for secondary users (rule 4.28) and joint accounts (rule 4A.14, with the CX notification standards). [S16 rules 4.28, 4A.14] [S1 #consumer-experience_notification-standards]
- CDR receipts and 90-day notifications are duties of data recipients. A receipt from a data holder is a CX Guideline only. [S1 #consumer-experience_notification-standards] [S17 Authorisation to disclose]
- Records: the data holder keeps records of authorisations, amendments and withdrawals, notices of withdrawn consents, disclosures, and each refusal to disclose with the rule or standard that it used. [S16 rule 9.3(1)]

---

## 8. Data holder reactions to Register state

The lifecycle note gives the Register side: the status values and cascades (section 2.3), the meaning of each status (section 2.4), the polling rules and cache periods (section 2.5), and the Register API paths and versions (section 3.6).

### 8.1 Register APIs that the data holder uses

| Register API | Use by the data holder | Recommended period |
|---|---|---|
| Get Data Recipients Statuses | Status of each data recipient legal entity | 2 to 5 minutes |
| Get Software Products Statuses | Status of each software product | 2 to 5 minutes |
| Get Data Recipients (statuses) | Statuses | 2 to 5 minutes |
| Get Data Recipients (other data) | Names, logos and other metadata, for example for the dashboard | 6 hours |
| Get JWKS | Keys to verify the SSA (and the Register JWT, section 9) | Not given in CDS 1.36.0 |

[S1 #metadata-cache-management] [S1 #cache-update-periods]

- All of these are TLS APIs with no authentication. The data holder needs no client certificate and no token to call the Register. See the lifecycle note, sections 1.3 and 3.1.

### 8.2 Cache rules

- "Data Holders MUST react to Data Recipient and associated Software Statuses changes within 5 minutes of the change occurring on the CDR Register." [S1 #metadata-cache-management]
- The data holder polls and "cache these for use during requests for Consumer Data". It "MUST choose a frequency of polling which ensures their systems can respond to status changes within the required timeframe". [S1 #metadata-cache-management]
- It "MUST cache all other participant data they use and periodically update this cache using a slow poll". [S1 #metadata-cache-management]
- If a status is absent or not recognised, the data holder "SHOULD ignore the value and use the previous status value". "There is no upper bound for how long previous status values should remain trusted." [S1 #data-holder-responsibilities]
- The Register can support ETag validators, so the data holder can poll with `If-None-Match`. [S1 #metadata-cache-management]

### 8.3 What to do for each status

The CDS table, with one row for each permitted pair: [S1 #data-holder-responsibilities] [S2 `slate/source/includes/security/_participant_statuses.md.erb`]

| Data recipient | Software product | Disclose CDR data | Facilitate consent authorisation | Facilitate consent withdrawal | Invalidate consents | Clean up registration |
|---|---|:-:|:-:|:-:|:-:|:-:|
| Active | Active | yes | yes | yes | | |
| Active | Inactive | | | yes | | |
| Active | Removed | | | | yes | yes |
| Suspended | Inactive | | | yes | | |
| Suspended | Removed | | | | yes | yes |
| Revoked | Removed | | | | yes | yes |
| Surrendered | Removed | | | | yes | yes |

How the columns map to endpoints. The CDS does not name each endpoint, so the last column gives the source of each row, or says that the mapping is not confirmed.

| Function | Active and Active | Software product Inactive | Software product Removed | Source or gap |
|---|---|---|---|---|
| Data requests (resource APIs, UserInfo) | Permit | Refuse | Refuse | 403 `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive` for resource APIs; "The error detail SHOULD contain the current status". [S1 #403-forbidden-errors] UserInfo: 401. [S18: 5729181240591] |
| New consents (PAR, authorisation) | Permit | Refuse | Refuse | "Facilitate Consent Authorisation" column. [S1 #data-holder-responsibilities] The error response is **UNVERIFIED**. |
| Withdrawal (arrangement revocation endpoint, dashboard) | Permit | Permit | Not applicable: consents are invalidated | "Facilitate Consent Withdrawal" column. [S1 #data-holder-responsibilities] |
| Existing consents | Stay valid | Stay valid; sharing stops and can start again | Expire ("Invalidate Consents" column) | [S18: 6828747049359] [S18: 900001657906]. For a revoked or surrendered accreditation: [S16 rule 4.26(2)] |
| Tokens | Honour | Not confirmed | "tokens ... must not be honoured" | [S18: 900001657906] For Inactive, the response of the token endpoint to a refresh request is **UNVERIFIED**. |
| Registration | Keep | Keep | Deactivate the client ID; never enable it again | [S18: 900001657906] |
| DCR `POST`, `PUT`, `GET`, `DELETE` from a product that is not Active | Permit | Not confirmed | Not confirmed | The CDS says nothing. **UNVERIFIED**. |

- For a Removed software product, "data sharing must cease immediately", but the clean-up has no time limit: "It is reasonable to execute these as batch tasks performed overnight". [S18: 900001657906] [S1 #data-holder-responsibilities]
- The data holder does not call the data recipient revocation endpoint when a Register status change invalidates an authorisation (section 5.4).
- "The status of a Data Recipient Brand does not impact Data Holder responsibilities". [S1 #data-holder-responsibilities]

---

## 9. Register calls to the data holder (Get Metrics, Metadata Update)

The first project defers these APIs. The lifecycle note, sections 3.2 and 3.3, gives the endpoints, the call frequency and the Register client. The authorisation server of the data holder must support only this:

- "Data Holders MUST support either Private Key JWT Client Authentication or Self-signed JWT Client Authentication of the CDR Register." They "SHOULD support Private Key JWT Client Authentication but MAY support Self-signed JWT Client Authentication". [S1 #cdr-register-calling-data-holders]
- Private Key JWT: the data holder issues a client ID to the Register out of band ("This MAY be a static client ID value of cdr-register"); `grant_type` is `client_credentials`; `scope` "MUST be provided and MUST be set to the scope of the resource endpoint"; the data holder issues an access token, and "a Refresh Token SHOULD NOT be included". [S1 #private-key-jwt-authentication]
- Scopes: `admin:metrics.basic:read` and `admin:metadata:update`, both "accessible ONLY to the CDR Register". They "MUST never be included in the same access token as a CDR Data scope". [S1 #authorisation-scopes]
- Self-signed JWT: no token endpoint call. The Register sends a signed JWT as the Bearer token, with client ID `cdr-register`, and `aud` "MUST" be the `<AdminBaseUri>`. The data holder rejects a `jti` that it saw before. [S1 #self-signed-jwt-client-authentication] [S1 #self-signed-jwt-authentication]
- The data holder verifies the Register JWT with the Register JWKS. See the lifecycle note, section 3.3.
- Transport: `AdminBaseUri` is MTLS, and the Register presents a CDR client certificate. [S1 #participant-endpoints]
- "This method MAY be changed by updating Data Holder registration details with the CDR Register." [S1 #cdr-register-calling-data-holders]
- Which method the production Register uses for a given data holder is **UNVERIFIED**. The ACCC page for Get Metrics does not say. [S20]
- The Register is thus a client that does not come from DCR: it has no SSA, and its keys are at the Register JWKS, not at a `jwks_uri` from an SSA.

---

## 10. Transport security for the data holder

The lifecycle note, sections 1.3 and 1.8, gives the trust model, certificate validation and the cipher rule. In summary for the data holder:

| Endpoint group | Transport | Server certificate | Client certificate |
|---|---|---|---|
| Discovery, JWKS, authorisation; public APIs (`PublicBaseUri`, `ProductBaseUri`) | TLS | Public CA "accepted by major web browsers" | None. These endpoints "MUST NOT use [MTLS]". |
| PAR, token, UserInfo, introspection, token revocation, arrangement revocation, registration; resource APIs (`ResourceBaseUri`); admin APIs (`AdminBaseUri`) | MTLS | CDR CA | CDR CA only: "The Server MUST NOT trust Client transport certificates issued by other authorities." |
| Outbound: Register APIs, data recipient `jwks_uri`, `sector_identifier_uri`, data recipient arrangement revocation endpoint | TLS | Public CA (the data holder is the client) | None |

[S1 #participant-endpoints] [S1 #use-of-mtls] [S1 #security-endpoints]

- "All HTTP calls MUST be made using HTTPS incorporating TLS >= 1.2." From 17 March 2025, "In addition to section 8.5 of [FAPI-1.0-Advanced] only cipher suites recommended in [BCP195] SHALL be permitted." [S1 #use-of-tls] [S1 #ciphers]
- Certificate validation "must check" validity, name chaining, policy and key use constraints, and revocation status "through Certificate Revocation Lists (CRL) or Online Certificate Status Protocol (OCSP) responders". [S1 #certificate-validation] "The use of OCSP Stapling within the CDR ecosystem is not recommended." [S1 #ocsp-stapling]
- A client with no certificate, an expired certificate or a certificate from a different CA: the data holder "can return a 401 Unauthorised error". [S18: 5729181240591]
- The CDS does not require that the data holder compares the client certificate with the Register data of the software product. See the lifecycle note, section 1.8. The holder-of-key check binds a token to the certificate that the client used at the token endpoint. [S10 section 3]
- CORS: "protections must be disabled (i.e. Access-Control-Allow-Origin set to *) for all unauthenticated endpoints unless specifically stated otherwise". Discovery, JWKS and authorisation are stated otherwise ("does not require CORS"). [S1 #cors] [S1 #openid-provider-configuration-endpoint]

---

## 11. Dated obligations and later changes

No CDS version later than 1.36.0 exists at the research date, so no later changes affect this scope. [S1 #changelog-and-archives] [S2]

Obligations with a date, in this scope:

| Date | Obligation | State at 2026-10-06 | Source |
|---|---|---|---|
| 4 July 2022 | Separate issuer for each brand; no two keys with one `kid` in a JWKS | In force | [S1 #multi-brand-support-separate-issuers-for-data-holder-brands] [S1 #jwks-uris] |
| 31 July 2022 | Data holders send `cdr_arrangement_jwt` to data recipients; `aud` of self-signed JWTs to data recipients is the resource path | In force | [S1 #cdr-arrangement-revocation-endpoint] [S1 #self-signed-jwt-client-authentication] |
| 31 August 2022 | Ignore unsupported scopes in the SSA | In force | [S1 #registration-validation] |
| 13 November 2023 | `client_id` parameter optional with a client assertion | In force | [S1 #private-key-jwt-client-authentication] |
| 24 April 2024 (CDS 1.30.0) | `sharing_expires_at` and `refresh_token_expires_at` removed | Removed | [S19: 543] |
| 17 March 2025 | Only BCP 195 cipher suites | In force | [S1 #ciphers] |
| 12 May 2025 | Authorization Code Flow only; the hybrid flow is retired | In force | [S1 #future-dated-obligations] |
| 12 May 2025 | Refresh token `exp` "MUST" be equal to the sharing duration | In force | [S1 #refresh-token] |
| 10 November 2025 | No JARM encryption unless the data recipient requests it | In force | [S1 #future-dated-obligations] |
| 10 May 2027 | Redirect to App, and the Authentication Flows standards for it | Future | [S1 #future-dated-obligations] |

Pending work that can change this scope:
- "Consultation Draft 374 - Minimum Baseline Authentication" is an open placeholder with no content. [S21: 374]
- Issue 679 (SSA specification, mutability of `legal_entity_id` and `legal_entity_name`), issue 649 (JARM error responses) and issue 480 (pairwise identifiers) are open. [S19: 679] [S19: 649] [S19: 480]
- The DSB consulted in May 2026 on how it develops standards. This does not change a requirement. [S21: 380]

---

## 12. Conflicts and gaps found

1. Non-normative CDS examples that do not agree with the normative text or the upstream specification:
   - Request object `aud` is the data recipient URL. FAPI requires the issuer of the data holder. The defect was reported in 2020 and is still in 1.36.0. [S1 #request-object] [S3 section 5.2.2, item 15] [S19: 156]
   - PAR response `expires_in` is 3600. The CDS limit is 10 to 90 seconds. [S1 #pushed-authorisation-endpoint] [S1 #request-object-submission]
   - Token response `expires_in` is 3600. The CDS limit for an access token is 2 to 10 minutes. [S1 #cdr-arrangement-id] [S1 #access-token]
   - Token responses have no `token_type`. RFC 6749 requires it. [S1 #cdr-arrangement-id] [S13 section 5.1]
   - Discovery: `issuer` is on the MTLS host, but the document is on the TLS host. OpenID Connect Discovery requires that the issuer is the prefix of the discovery URL. [S1 #openid-provider-configuration-endpoint] [S15 section 4.3]
   - Discovery: `grant_types_supported` has no `refresh_token` and has a CIBA grant type that the CDS does not use; `response_types_supported` has `code id_token`; `response_modes_supported` has `fragment`. [S1 #openid-provider-configuration-endpoint]
   - `code_challenge` is a padded base64 value, not the 43-character base64url value of S256. The encoded `cdr_arrangement_jwt` value has `HS256` in its header, and the labels of the two data recipient examples ("Until July 31st 2022", "From March 31st 2022") overlap. [S1 #request-object] [S7 section 4.2] [S1 #cdr-arrangement-revocation-endpoint]
2. Hybrid flow residue. The flow is retired from 12 May 2025, but CDS 1.36.0 still has the "OIDC Hybrid Flow" sections, the ID token encryption rules, `code id_token` in the DCR schema, and the ID token encryption discovery fields. [S1 #oidc-hybrid-flow] [S1 #id-token-algorithm-selection-considerations] [S2 `cds_dcr.json`] [S19: 666]
3. DCR text and DCR OpenAPI do not agree: `redirect_uris` and `request_object_signing_alg` are optional in the text and mandatory in the schema; `revocation_uri` and `recipient_base_uri` are required in the SSA definition and optional in the schema. [S1 #registration-request-using-jwt] [S1 #ssa-definition] [S2 `cds_dcr.json`]
4. DCR path. The OpenAPI server URL is `https://mtls.dh.example.com/cds-au/v1`, which gives `/cds-au/v1/register`. The security profile examples use `/register`, and the normative rule is that the path comes from `registration_endpoint`. [S2 `cds_dcr.json`] [S1 #registration-management]
5. Tokens at revocation of an arrangement. The CDS says "Consent revocation MUST also revoke associated refresh and/or access tokens". Support articles say "the access token is not guaranteed to be revoked" and that the clean-up of the refresh token "is an implementation decision". All agree that no data is disclosed after the revocation. [S1 #cdr-arrangement-revocation-endpoint] [S18: 6828747049359] [S18: 4412301531663]
6. `aud` of the data holder call to the data recipient. The CDS says the resource path of the endpoint "MUST be used" from 31 July 2022. A support article (updated 2024) says the `RecipientBaseUri`, "not the full path". A second article says that the client assertion "is sent in the body", but the CDS puts the JWT in the `Authorization` header. [S1 #self-signed-jwt-client-authentication] [S18: 6828747049359] [S18: 6382507783183]
7. Redirect to App. The OTP support article (updated 25 November 2025) says the standards "do not support Security App or App2App redirect". CDS 1.35.0 added Redirect to App, with an obligation from 10 May 2027. [S18: 6140553082255] [S1 #redirect-to-app]
8. Which name the data holder shows. Rule 4.23(1)(a) requires the name of the accredited person. The CX Guidelines say that this is the legal entity name, and also say to use the brand name and not the software product name in authentication screens. The SSA definition says that `client_name` and `org_name` are "to be presented to the end user during authorization", and marks `legal_entity_name` as optional. [S16 rule 4.23(1)(a)] [S17 Authorisation to disclose] [S17 Redirect to Web with One Time Password] [S1 #ssa-definition]
9. HTTP code after a withdrawal. A support article gives 401 with an OAuth error when "an API is invoked" after the consent is withdrawn, and does not say if it means a resource API or the token endpoint. For the token endpoint, RFC 6749 gives 400 `invalid_grant`. The same article says that upstream standards "take precedence over CDS error codes". [S18: 5729181240591] [S13 section 5.2]
10. JARM default signing algorithm: PS256 in the CDS, RS256 in JARM. The CDS version of the JARM reference is "October 2020", at a Bitbucket URL; this note used the Final version of 17 August 2025. That the two versions agree in the cited sections is **UNVERIFIED**. [S1 #authorization-code-flow] [S5 section 3] [S1 #normative-references]
11. `request_uri` one-time use: "MUST reject" in the CDS, "SHOULD" in RFC 9126, with a permitted repeat after a browser reload. [S1 #security-profile_request-object_data-holders] [S6 section 4]
12. Introspection of a token that is not active: the CDS field list reads as required in all cases. The DSB clarifies that RFC 7662 applies. [S1 #introspection-endpoint] [S18: 900004618103]
13. Pairwise identifiers and `sector_identifier_uri`: the CDS rule can give two software products the same `sub`. Issue 480 is open, and the related support article is "UNDER REVIEW". [S1 #sub-claim] [S19: 480] [S18: 360004324176]
14. `sharing_duration` of exactly 24 hours: one-time collection in the CDS ("less than or equal to 24 hours"); the CX Guidelines say "less than 24 hours". [S1 #requesting-sharing-duration] [S17 Authorisation to disclose]
15. Registration management has no upstream reference: the CDS cites RFC 7591 only, and its management endpoints are a variant of RFC 7592 with a different credential. [S1 #normative-references] [S9]
16. Old guidance that the CDS now replaces: PKCE support as optional for a data holder (2022). The DCR guidance on "Deleting ADR Client Registration" mixes the deletion of a registration with the Register statuses, and is not clear. [S18: 5000744544143] [S18: 6828251436431]
17. Gaps, with no rule in a primary source: the mapping from DCR failure conditions to the four error codes; the maximum lifetime and replay rule of the registration JWT; the maximum lifetime of a client assertion; the lifetime of an authorisation code; the JARM response modes that are required; the response to introspection of an access token; the behaviour of the token and registration endpoints for an INACTIVE software product; the effect of a refresh token revocation on the authorisation; the delivery of an OTP in a test system.

---

## 13. Open questions for requirements

- Issuer and hosts: must the mock use one issuer URL that is also the `InfoSecBaseUri` (OpenID Connect Discovery), with the MTLS endpoints on a second host, as the Register does?
- Which signing algorithms does the mock offer: PS256 only, or PS256 and ES256? Data recipients must support both.
- Does the mock support JARM encryption (optional and not recommended in the CDS)? Which JARM response modes does it accept: `jwt` only, or also `query.jwt`, `fragment.jwt` and `form_post.jwt`?
- Does the mock implement the optional `DELETE /register/{ClientId}`, or return 405?
- Which error code does the mock return for each DCR failure, and for a duplicate `software_id` (the CDS example uses `invalid_software_statement`)?
- Does the mock refuse a registration JWT `jti` that it saw before, and what maximum lifetime does it accept?
- Does the mock get the `sector_identifier_uri` document at registration and refuse the request when it cannot? The article that requires this is under review.
- How does the mock calculate `sub`: a stored UUID for each pair of user and sector identifier (recommended), or a hash? Does it use the `sector_identifier_uri` host although issue 480 is open?
- What does the mock do at the token, PAR and registration endpoints for an INACTIVE software product, and with which error? The CDS gives the answer only for data requests and for REMOVED products.
- Does the mock invalidate consents and deactivate registrations of REMOVED products at once, or in a batch? Both are permitted, but sharing must stop within 5 minutes.
- How does the mock deliver an OTP with no real channel, and which expiry period does it use?
- Which `acr` values does the mock give and advertise: `urn:cds.au:cdr:2` only, or also `urn:cds.au:cdr:3`? What does it do when LoA 3 is essential?
- Which name does the authorisation screen show: legal entity name, brand name, software product name, or more than one? Does the name come from the SSA or from Get Data Recipients?
- The first project puts amendment out of scope, but a request object can have `cdr_arrangement_id`. Does the mock refuse such a request, or treat it as a new consent under the same arrangement?
- Does a token revocation of the only refresh token of an arrangement end the authorisation? What does the mock do with an authorisation when the data recipient does not exchange the code?
- Does the first project send the notice to the data recipient (`cdr_arrangement_jwt`) when the consumer withdraws? The rule is a civil penalty provision, but the dashboard is deferred, so the mock needs a different way to start a withdrawal.
- Does the mock keep the dashboard data (rule 1.15(3), rule 7.9) from the first project, so that a later dashboard has the history?
- Which client authentication of the Register does the mock plan for Get Metrics: `private_key_jwt` (recommended) or self-signed JWT?
- Does the OpenID conformance suite (FAPI 1.0 Advanced, with PAR and JARM) run against the mock, as ADR 0001 says?

---

## 14. Source URLs for CX Guidelines pages and support articles

CX Guidelines pages used (S17), base URL https://cx.dsb.gov.au/:
`authenticate/redirect-with-one-time-password` Redirect to Web with One Time Password; `authorise/authorisation-to-disclose` Authorisation to disclose. The site says that these two pages changed on 18 March 2026 for Consultation Draft 376 (white label brand arrangements).

CDR Support Portal articles used (S18), base URL https://cdr-support.zendesk.com/hc/en-us/articles/:
6828251436431 Guidance on Dynamic Client Registration (DCR); 6382507783183 Guidance on Client Authentication; 6828747049359 Guidance on revocation or withdrawal of consent; 6140553082255 One Time Password (OTP); 4412301531663 Data Holder obligations for revoked consents; 360004056455 Data Recipient unavailability back-off pattern; 900004618103 Introspection Endpoint RFC vs Standards clarification; 360004324176 Data Holder sector_identifier_uri Support (under review); 5505351487631 ID Permanence and PPID; 4886150993807 Consent amendment with '0' sharing duration; 360004559675 Maintaining cdr_arrangement_id in amended consent; 4407560834959 Concurrent consent and arrangement ID requirements; 6662499263887 Guidance on token validation and JWKS; 900003036086 OIDC offline access scope; 360004665496 Failure to obtain token within Authorisation Code expiry period; 5729181240591 Error scenarios and responses; 360004023156 Authorisation and Consent; 900003385706 Clarification on OIDC endpoint parameters; 4403543156751 Egress policy for calls from DH to ADRs; 900004240446 Pushed Authorization Requests redirect_uri parameter; 5000744544143 Invoking authorisation code flow using PKCE claims; 900003387786 Support for token revocation endpoints; 4403929644559 Data Holder Endpoint Configuration on the CDR Register; 900001657906 Status of consent when an ADR Software Product is "Removed".

Standards maintenance issues used (S19): 156 aud in Request Object; 367 Synchronisation of consents after outages; 480 Pseudonymity of Pairwise Identifiers (open); 543 refresh_token_expires_at and sharing_expires_at claims; 649 Inconsistent JARM error responses (open); 650 Weaken JARM Encryption Requirements for ADRs; 666 Retirement of OIDC Hybrid Flow; 667 Clean up of Refresh Token requirements; 679 Update SSA specification (open).
