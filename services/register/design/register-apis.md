# Mock CDR Register: Register APIs

This area gives the rules that apply to all Register APIs: the endpoints, the versions and headers, the errors, caching, pagination, lists, CORS and transport security. The [CDS HTTP library](../../../pkg/cds/DESIGN.md) gives the shared rules. This area gives only the choices of the Register. The [software statement assertion](software-statement-assertion.md) and the [authorisation server](authorisation-server.md) give their own endpoints.

## Endpoints

| ID | API | Method and path | Version | Transport | Authentication | Source |
|---|---|---|---|---|---|---|
| <a id="reg-0300"></a>**REG-0300** | Get Data Holder Brands | `GET /cdr-register/v1/{industry}/data-holders/brands` | 3 | mTLS | Bearer token with `cdr-register:read` | [CDS: Get Data Holder Brands][cds-get-dh-brands]; [Judgement: only the current version of each API](#reg-0372) |
| <a id="reg-0301"></a>**REG-0301** | Get Data Holder Brands Summary | `GET /cdr-register/v1/{industry}/data-holders/brands/summary` | 2 | TLS | None | [CDS: Get Data Holder Brands Summary][cds-get-dh-brands-summary]; [Judgement: only the current version of each API](#reg-0372) |
| <a id="reg-0302"></a>**REG-0302** | Get Data Holder Statuses | `GET /cdr-register/v1/{industry}/data-holders/status` | 2 | TLS | None | [CDS: Get Data Holder Statuses][cds-get-dh-statuses]; [Judgement: only the current version of each API](#reg-0372) |
| <a id="reg-0303"></a>**REG-0303** | Get Software Products Statuses | `GET /cdr-register/v1/{industry}/data-recipients/brands/software-products/status` | 3 | TLS | None | [CDS: Get Software Products Statuses][cds-get-sp-statuses]; [Judgement: only the current version of each API](#reg-0372) |
| <a id="reg-0304"></a>**REG-0304** | Get Data Recipients Statuses | `GET /cdr-register/v1/{industry}/data-recipients/status` | 3 | TLS | None | [CDS: Get Data Recipients Statuses][cds-get-dr-statuses]; [Judgement: only the current version of each API](#reg-0372) |
| <a id="reg-0305"></a>**REG-0305** | Get Data Recipients | `GET /cdr-register/v1/{industry}/data-recipients` | 4 | TLS | None | [CDS: Get Data Recipients][cds-get-drs]; [Judgement: only the current version of each API](#reg-0372) |

## Requests and responses

- <a id="reg-0306"></a>**REG-0306** The request and response bodies obey the CDS 1.36.0 Register OpenAPI specification, except where a requirement of this specification gives a different rule. ([CDS: Register OpenAPI specification][cds-openapi])
- <a id="reg-0307"></a>**REG-0307** The TLS APIs and the mTLS APIs have different base URIs. For example, production uses `api.cdr.gov.au` and `secure.api.cdr.gov.au`. ([CDS: Register APIs][cds-register-apis])
- <a id="reg-0308"></a>**REG-0308** On a data holder API, `{industry}` accepts `banking`, `energy`, `non-bank-lending`, `telco` and `all`. ([CDS: Register API schemas][cds-register-schemas])
- <a id="reg-0309"></a>**REG-0309** On a data recipient API and on [Get SSA](software-statement-assertion.md#reg-0400), `{industry}` accepts only `all`. ([CDS: Register API schemas][cds-register-schemas])
- <a id="reg-0310"></a>**REG-0310** On Get Data Holder Statuses, an `{industry}` other than `all` gives only the data holder legal entities that have a brand in that industry. The status of the brand does not matter. ([CDS: Get Data Holder Statuses][cds-get-dh-statuses]; [Judgement: Get Data Holder Statuses filters by the industry of a brand in each status](#reg-0382))
- <a id="reg-0311"></a>**REG-0311** On Get Data Holder Statuses, `all` gives all data holder legal entities. ([CDS: Get Data Holder Statuses][cds-get-dh-statuses])
- <a id="reg-0312"></a>**REG-0312** The Register checks the version headers of a request before its `{industry}`. For example, a request with an unsupported `x-v` and an invalid `{industry}` gets 406. ([Judgement: the Register checks the version before the `{industry}`](#reg-0371))

## Versioning and headers

- <a id="reg-0313"></a>**REG-0313** An API of the [endpoint table](#reg-0300) uses [the CDS version negotiation](../../../pkg/cds/DESIGN.md#cdsl-0001), with the version of the table as its lowest and its highest version. ([Judgement: only the current version of each API](#reg-0372))
- <a id="reg-0314"></a>**REG-0314** Each response of the public APIs and the secure APIs has [the `x-fapi-interaction-id` of the request, or a new UUID](../../../pkg/cds/DESIGN.md#cdsl-0010). ([CDS: Register APIs][cds-register-apis])

## Errors

Each error response uses [the CDS error format](../../../pkg/cds/DESIGN.md#cdsl-0012), and a page error uses [the errors of CDS pagination](../../../pkg/cds/DESIGN.md#cdsl-0020). An unexpected server error gives [500 with no cause in the `detail`](../../../pkg/cds/DESIGN.md#cdsl-0015).

| ID | Condition | Status | Error code | Source |
|---|---|---|---|---|
| <a id="reg-0315"></a>**REG-0315** | An invalid `{industry}` on Get Data Holder Brands Summary | 404 | `urn:au-cds:error:cds-register:Field/InvalidIndustry` | [CDS: Invalid Industry][cds-error-invalid-industry]; [CDS: Get Data Holder Brands Summary][cds-get-dh-brands-summary]; [Judgement: 404 on the summary and 400 on the other APIs, as the ACCC says](#reg-0370) |
| <a id="reg-0316"></a>**REG-0316** | An invalid `{industry}` on each other API, also an industry that the API does not accept, for example `banking` on a data recipient API | 400 | `urn:au-cds:error:cds-all:Field/Invalid` | [CDS: Invalid Field][cds-error-400-field-invalid]; [ACCC Developer Portal: Known Issues][accc-known-issues]; [Judgement: 404 on the summary and 400 on the other APIs, as the ACCC says](#reg-0370) |
| <a id="reg-0317"></a>**REG-0317** | An `updated-since` that is not an RFC 3339 date-time | 400 | `urn:au-cds:error:cds-all:Field/InvalidDateTime` | [CDS: Error Codes][cds-error-codes] |
| <a id="reg-1494"></a>**REG-1494** | An `Accept` header that [permits no JSON response in UTF-8](../../../pkg/cds/DESIGN.md#cdsl-0071), for example `application/xml` | 406 | `urn:au-cds:error:cds-all:Header/Invalid` | [CDS: Request Headers][cds-request-headers] |
| <a id="reg-0318"></a>**REG-0318** | An unknown path | 404 | `urn:au-cds:error:cds-all:Resource/NotFound` | [CDS: Error Codes][cds-error-codes]; [the 404 for a path with no method](../../../pkg/cds/DESIGN.md#cdsl-0082) |
| <a id="reg-1495"></a>**REG-1495** | A method that the path does not support, on the public or the secure APIs, for example `POST` on Get Data Recipients. The response has [an `Allow` header](../../../pkg/cds/DESIGN.md#cdsl-0081) | 405 | `urn:au-cds:error:cds-all:GeneralError/Expected` | [CDS: HTTP Response Codes][cds-http-response-codes]; [405 for a method that the path does not support](../../../pkg/cds/DESIGN.md#cdsl-0080) |

- <a id="reg-0319"></a>**REG-0319** The Register ignores an empty `updated-since`. For example, `updated-since=` gives the same response as no `updated-since`. ([CDS: Error Codes][cds-error-codes])

## Secure APIs

The secure APIs are Get Data Holder Brands and [Get SSA](software-statement-assertion.md#reg-0400). The [authorisation server](authorisation-server.md) issues their access tokens.

- <a id="reg-0320"></a>**REG-0320** When a request to a secure API has no bearer token, the response is 401 with the header `WWW-Authenticate: Bearer` and no body. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [RFC 6750, section 3.1][rfc6750-errors])
- <a id="reg-0321"></a>**REG-0321** When the bearer token is not valid, or when it is bound to a different client certificate, the response is 401 with `WWW-Authenticate: Bearer error="invalid_token"` and no body. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [RFC 6750, section 3.1][rfc6750-errors]; [a token is valid only with its client certificate](../../../pkg/authserver/design/token-endpoint.md#auth-0055))
- <a id="reg-0322"></a>**REG-0322** When a valid token does not have the scope `cdr-register:read`, the response is 403 with `WWW-Authenticate: Bearer error="insufficient_scope"`. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [RFC 6750, section 3.1][rfc6750-errors])
- <a id="reg-0323"></a>**REG-0323** A secure API checks the status of the caller on each call. When the software product or its data recipient is not ACTIVE, the response is 403 with `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`. ([CDS: Error Codes][cds-error-403-adr-not-active])
- <a id="reg-0324"></a>**REG-0324** When the software product of the caller does not exist, the response is 403 with `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive` and the `detail` REMOVED. ([CDS: Error Codes][cds-error-403-adr-not-active]; [Judgement: a caller with no software product gets the response for REMOVED](#reg-0384))
- <a id="reg-0325"></a>**REG-0325** A secure API checks the bearer token first, then the status of the caller, then the checks of the public APIs. For example, a request with no token and an invalid `{industry}` gets 401. ([Judgement: the token first, then the status of the caller, then the public checks](#reg-0383))

## Caching

- <a id="reg-0326"></a>**REG-0326** These APIs return an `ETag` header. When `If-None-Match` matches the current `ETag`, the response is 304 with no body. ([CDS: Metadata Cache Management][cds-metadata-cache])
  - [Get JWKS](software-statement-assertion.md#reg-0401)
  - Get Data Holder Brands Summary
  - Get Data Holder Statuses
  - Get Software Products Statuses
  - Get Data Recipients Statuses
  - Get Data Recipients
- <a id="reg-0327"></a>**REG-0327** An `ETag` follows RFC 9110: the entity tag is in quotes, and `If-None-Match` uses the weak comparison. For example, `If-None-Match: W/"abc"` matches `ETag: "abc"`. ([RFC 9110: ETag][rfc9110-etag]; [Judgement: the HTTP format of an `ETag`](#reg-0373))
- <a id="reg-0328"></a>**REG-0328** An `ETag` changes only when the data of the response changes. ([RFC 9110: ETag][rfc9110-etag]; [Judgement: the HTTP format of an `ETag`](#reg-0373))
- <a id="reg-0329"></a>**REG-0329** Each response of an API of the [`ETag` list](#reg-0326) has `Cache-Control: no-cache`, as [the OpenID Provider configuration](../../../pkg/authserver/design/configuration.md#auth-0002) does. ([RFC 9111: no-cache][rfc9111-no-cache]; [Judgement: no cache lifetime for the Register APIs](#reg-0374))

## Lists

- <a id="reg-0330"></a>**REG-0330** Get Data Holder Brands supports `page` and `page-size` with [the CDS pagination rules](../../../pkg/cds/DESIGN.md#cdsl-0017). ([CDS: Get Data Holder Brands][cds-get-dh-brands])
- <a id="reg-0331"></a>**REG-0331** Get Data Holder Brands supports `updated-since`. It gives the brands with a `lastUpdated` at or after that time. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [Judgement: `updated-since` includes the same time](#reg-0375))
- <a id="reg-0332"></a>**REG-0332** Each link of a page of Get Data Holder Brands keeps [the `updated-since` of the request](../../../pkg/cds/DESIGN.md#cdsl-0029). ([CDS: Get Data Holder Brands][cds-get-dh-brands])
- <a id="reg-0333"></a>**REG-0333** Get Data Holder Brands returns each brand that has a `dataHolderBrandId`, in each status: ACTIVE, INACTIVE and REMOVED. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [Judgement: all brand statuses on Get Data Holder Brands, as the CDS says](#reg-0378))
- <a id="reg-0334"></a>**REG-0334** Get Data Holder Brands returns the brands in the order of their IDs. ([Judgement: Get Data Holder Brands in the order of the IDs](#reg-0376))
- <a id="reg-0335"></a>**REG-0335** The response of Get Data Holder Brands has no `ETag`. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [Judgement: no `ETag` on Get Data Holder Brands](#reg-0377))
- <a id="reg-0336"></a>**REG-0336** The `lastUpdated` of a brand on Get Data Holder Brands is the time of the last change to its data holder. For example, a change to a different brand of the same data holder also changes it. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [Judgement: `lastUpdated` includes the changes to the children](#reg-0380))
- <a id="reg-0337"></a>**REG-0337** A list API other than Get Data Holder Brands is not paginated. It uses [the response envelope with `links.self` and an empty `meta`](../../../pkg/cds/DESIGN.md#cdsl-0034). ([CDS: Register API schemas][cds-register-schemas])
- <a id="reg-0338"></a>**REG-0338** The status APIs and Get Data Recipients include the records in all statuses, not only ACTIVE. ([CDS: Participant Statuses][cds-participant-statuses])
- <a id="reg-0339"></a>**REG-0339** The `lastUpdated` of a data recipient on Get Data Recipients is the time of the last change to the data recipient, its brands or its software products. ([CDS: Get Data Recipients][cds-get-drs]; [Judgement: `lastUpdated` includes the changes to the children](#reg-0380))
- <a id="reg-0340"></a>**REG-0340** Get Data Holder Brands Summary includes the brands that have only an `interimId` and no `dataHolderBrandId`. ([CDS: DataHolderBrandSummaryV2][cds-schema-brand-summary]; [ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary])
- <a id="reg-0341"></a>**REG-0341** Get Data Holder Brands Summary includes the ACTIVE and the INACTIVE brands. It does not include a REMOVED brand. ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary]; [Judgement: no REMOVED brand in the summary](#reg-0379))
- <a id="reg-0342"></a>**REG-0342** On Get Data Holder Brands Summary, an `{industry}` other than `all` gives only the brands that have that industry. ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary])
- <a id="reg-0343"></a>**REG-0343** The `lastUpdated` of a brand on Get Data Holder Brands Summary is the time of the last change to its data holder. ([CDS: DataHolderBrandSummaryV2][cds-schema-brand-summary]; [Judgement: `lastUpdated` includes the changes to the children](#reg-0380))

## CORS

- <a id="reg-0344"></a>**REG-0344** The unauthenticated APIs use [the CORS rules](../../../pkg/cds/DESIGN.md#cdsl-0037), also for [a preflight request](../../../pkg/cds/DESIGN.md#cdsl-0039). ([CDS: CORS][cds-cors])
- <a id="reg-0345"></a>**REG-0345** The CORS rules of the unauthenticated APIs [expose these headers](../../../pkg/cds/DESIGN.md#cdsl-0038), so that a web page can read them. ([Fetch: CORS protocol][fetch-cors]; [Judgement: CORS exposes the version, `ETag` and interaction ID headers](#reg-0385))
  - `x-v`
  - `etag`
  - `x-fapi-interaction-id`

## Transport security

- <a id="reg-0346"></a>**REG-0346** All CDR endpoints use [the TLS versions](../../../pkg/cds/DESIGN.md#cdsl-0041) and [the TLS 1.2 cipher suites](../../../pkg/cds/DESIGN.md#cdsl-0042) of the CDS HTTP library. ([CDS: Transaction Security][cds-transaction-security])
- <a id="reg-0347"></a>**REG-0347** An mTLS endpoint presents a server certificate from the mock CDR CA. ([CDS: Certificate Trust Model][cds-cert-trust-model])
- <a id="reg-0348"></a>**REG-0348** An mTLS endpoint accepts only [client certificates that chain to the CDR CA bundle](../../../pkg/cds/DESIGN.md#cdsl-0044). The operator interface gives [the same CDR CA bundle](certificate-management.md#reg-0617). The Register also does [a revocation check of each client certificate](certificate-revocation.md#reg-0726). ([CDS: Certificate Trust Model][cds-cert-trust-model])
- <a id="reg-0349"></a>**REG-0349** A TLS endpoint presents a certificate from a mock public CA, not from the mock CDR CA. ([CDS: Participant Endpoints][cds-participant-endpoints]; [CDS: Transaction Security][cds-transaction-security])
- <a id="reg-0350"></a>**REG-0350** A TLS endpoint [does not request a client certificate](../../../pkg/cds/DESIGN.md#cdsl-0043). ([CDS: Transaction Security][cds-transaction-security])
- <a id="reg-0351"></a>**REG-0351** When a proxy terminates TLS, the Register accepts the client certificate only from that trusted proxy. A client cannot give the certificate in a request header. ([RFC 8705, section 3][rfc8705-binding]; [Judgement: the client certificate only from a trusted proxy](#reg-0386))

## Judgements

### <a id="reg-0370"></a>REG-0370 404 on the summary and 400 on the other APIs, as the ACCC says

- Kind: conflict
- Basis: Step 2. On Get Data Holder Brands Summary, the CDS error list and the OpenAPI specification agree on 404 `Field/InvalidIndustry` ([CDS: Invalid Industry][cds-error-invalid-industry]). On the other APIs, the error list gives 404 `Field/InvalidIndustry` and the OpenAPI specification gives 400. The ACCC clarifies that current versions return 400 `Field/Invalid` ([ACCC Developer Portal: Known Issues][accc-known-issues]).
- Reason: Where the CDS sources agree, the Register follows them. Where they do not agree, the ACCC guidance decides.

### <a id="reg-0371"></a>REG-0371 The Register checks the version before the `{industry}`

- Kind: choice
- Basis: Step 3.3. The CDS gives no sequence. The rules for `{industry}` depend on the version: deprecated versions give 404 `Field/InvalidIndustry`, and current versions give 400 `Field/Invalid` ([ACCC Developer Portal: Known Issues][accc-known-issues]).
- Reason: The Register must know the version before it can check `{industry}`.

### <a id="reg-0372"></a>REG-0372 Only the current version of each API

- Kind: choice
- Basis: Step 3.3. The CDS gives the versions of each Register API ([CDS: Register APIs][cds-register-apis]). The production Register also serves deprecated versions until their retirement.
- Reason: The lab components use the current versions. One version for each API is the least complex. A request for a deprecated version gets 406 from the version negotiation.

### <a id="reg-0373"></a>REG-0373 The HTTP format of an `ETag`

- Kind: choice
- Basis: Step 3.2. The CDS requires an `ETag` but does not give its format ([CDS: Metadata Cache Management][cds-metadata-cache]).
- Reason: The HTTP standard applies ([RFC 9110: ETag][rfc9110-etag]). An HTTP client then handles the `ETag` with no special code.

### <a id="reg-0374"></a>REG-0374 No cache lifetime for the Register APIs

- Kind: choice
- Basis: Step 3.3. The CDS gives no cache lifetime for the responses. A status change must reach data holders quickly ([CDS: Metadata Cache Management][cds-metadata-cache]).
- Reason: With `no-cache`, a client checks the `ETag` on each request ([RFC 9111: no-cache][rfc9111-no-cache]). This is the simplest way to make sure that the client sees each change.

### <a id="reg-0375"></a>REG-0375 `updated-since` includes the same time

- Kind: choice
- Basis: Step 3.3. The CDS says only "updated since the specified date-time" ([CDS: Get Data Holder Brands][cds-get-dh-brands]).
- Reason: With "after", a data recipient that polls with the last time that it saw can miss a change at the same time. A record that the API returns again is harmless.

### <a id="reg-0376"></a>REG-0376 Get Data Holder Brands in the order of the IDs

- Kind: choice
- Basis: Step 3.3. The CDS gives no order ([CDS: Get Data Holder Brands][cds-get-dh-brands]).
- Reason: Pagination needs a stable order. The other lists of the Register also use the order of the IDs.

### <a id="reg-0377"></a>REG-0377 No `ETag` on Get Data Holder Brands

- Kind: choice
- Basis: Step 1. The CDS OpenAPI specification gives no `If-None-Match` parameter and no `ETag` header for this API ([CDS: Get Data Holder Brands][cds-get-dh-brands]).
- Reason: `updated-since` gives the incremental read of this API. An `ETag` adds a second way for no gain.

### <a id="reg-0378"></a>REG-0378 All brand statuses on Get Data Holder Brands, as the CDS says

- Kind: conflict
- Basis: Step 1. The schema gives the brand a `status` with all three values, and `updated-since` lets a data recipient poll for changes ([CDS: Get Data Holder Brands][cds-get-dh-brands]). An ACCC support article says that the API lists the brands "active for consumer data sharing" ([CDR Support: Brands in the CDR ecosystem][support-brands]). The article is not a specification and is not clear.
- Reason: The CDS wins over an ACCC support article. A data recipient that has a brand in its cache learns that the brand is INACTIVE or REMOVED only when the API returns it.

### <a id="reg-0379"></a>REG-0379 No REMOVED brand in the summary

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC do not say which statuses the summary includes, and the summary has no status field ([CDS: DataHolderBrandSummaryV2][cds-schema-brand-summary]).
- Reason: The brands of data holders that left the CDR are not in the production list. A REMOVED brand has no endpoints to discover. An INACTIVE brand can be unavailable for a short time only.

### <a id="reg-0380"></a>REG-0380 `lastUpdated` includes the changes to the children

- Kind: choice
- Basis: Step 3.3. The CDS says only "the date/time that the Legal Entity was last updated" ([CDS: Get Data Recipients][cds-get-drs]).
- Reason: A change to a brand or a software product is a change to the legal entity record that the API returns. A data holder that polls the API then sees each change. The brands of a data holder use the same rule.

### <a id="reg-0382"></a>REG-0382 Get Data Holder Statuses filters by the industry of a brand in each status

- Kind: choice
- Basis: Step 3.3. The CDS says only that the parameter is "the industry the participant is retrieving data for" ([CDS: Get Data Holder Statuses][cds-get-dh-statuses]). Get Data Holder Brands and the summary give only the brands of the industry.
- Reason: One meaning of `{industry}` for each data holder API is the simplest. The brands of a REMOVED data holder are REMOVED, so the filter ignores the brand status, and the data recipients of that industry see that the data holder is REMOVED.

### <a id="reg-0383"></a>REG-0383 The token first, then the status of the caller, then the public checks

- Kind: choice
- Basis: Step 3.3. The CDS gives no sequence ([CDS: Get Data Holder Brands][cds-get-dh-brands]).
- Reason: A caller with no valid token then learns nothing about the API.

### <a id="reg-0384"></a>REG-0384 A caller with no software product gets the response for REMOVED

- Kind: choice
- Basis: Step 3.3. The CDS gives no response for a token whose software product does not exist ([CDS: Error Codes][cds-error-403-adr-not-active]).
- Reason: The Register never deletes a software product, so this cannot occur in practice. The response is the same as for a REMOVED software product, which is the nearest state.

### <a id="reg-0385"></a>REG-0385 CORS exposes the version, `ETag` and interaction ID headers

- Kind: choice
- Basis: Step 3.2. The CDS requires that CORS protections are disabled for these APIs ([CDS: CORS][cds-cors]). A browser hides each response header that the server does not expose ([Fetch: CORS protocol][fetch-cors]).
- Reason: A web page needs the version, the `ETag` and the interaction ID of a response.

### <a id="reg-0386"></a>REG-0386 The client certificate only from a trusted proxy

- Kind: choice
- Basis: Step 3.2. An access token is bound to the client certificate of the TLS connection ([RFC 8705, section 3][rfc8705-binding]). A request header is not authenticated.
- Reason: A client can forge a header that gives a certificate. Only the proxy that terminated the TLS connection knows the certificate of the client.

[accc-brands-summary]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982
[accc-known-issues]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29950003
[cds-cert-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-cors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors
[cds-error-400-field-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid
[cds-error-403-adr-not-active]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-adr-status-not-active
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[cds-error-invalid-industry]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-error-404-field-invalid-industry
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-get-dh-brands-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands-summary
[cds-get-dh-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-statuses
[cds-get-dr-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients-statuses
[cds-get-drs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients
[cds-get-sp-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-products-statuses
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-openapi]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_register.json
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-request-headers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-headers
[cds-http-response-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-response-codes
[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-apis
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-schema-brand-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_schemas_tocSdataholderbrandsummaryv2
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[fetch-cors]: https://fetch.spec.whatwg.org/#http-cors-protocol
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc8705-binding]: https://www.rfc-editor.org/rfc/rfc8705#section-3
[rfc9110-etag]: https://www.rfc-editor.org/rfc/rfc9110#name-etag
[rfc9111-no-cache]: https://www.rfc-editor.org/rfc/rfc9111#name-no-cache-2
[support-brands]: https://cdr-support.zendesk.com/hc/en-us/articles/900005881263
