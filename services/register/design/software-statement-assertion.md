# Mock CDR Register: Software statement assertion

This area gives Get SSA, the claims and the signature of an SSA, and Get JWKS, which gives the keys that sign each SSA. The [Register APIs](register-apis.md) give the rules that apply to all Register APIs, also to these two. The [authorisation server](authorisation-server.md#reg-0465) gives the key sets and the key changes.

## Endpoints

| ID | API | Method and path | Version | Transport | Authentication | Source |
|---|---|---|---|---|---|---|
| <a id="reg-0400"></a>**REG-0400** | Get SSA | `GET /cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa` | 4 | mTLS | Bearer token with `cdr-register:read` | [CDS: Get SSA][cds-get-ssa]; [Judgement: only the current version of each API](register-apis.md#reg-0372) |
| <a id="reg-0401"></a>**REG-0401** | Get JWKS: the SSA signing keys and the outbound signing keys | `GET /cdr-register/v1/jwks` | None | TLS | None | [CDS: Get JWKS][cds-get-jwks] |

## Issue

- <a id="reg-0402"></a>**REG-0402** The Register issues an SSA only for an ACTIVE software product of an ACTIVE data recipient. For other callers, the [status check of the secure APIs](register-apis.md#reg-0323) refuses the request. ([CDS: SSA][cds-ssa])
- <a id="reg-0403"></a>**REG-0403** When the path `softwareProductId` is not the software product of the caller, the response is 404 with `urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct`. ([CDS: CDR Register Errors][cds-register-errors]; [Judgement: 404 for a brand or a software product of a different caller](#reg-0424))
- <a id="reg-0404"></a>**REG-0404** When the path `dataRecipientBrandId` is not the brand of the caller, the response is 404 with `urn:au-cds:error:cds-register:Field/InvalidBrand`. ([CDS: CDR Register Errors][cds-register-errors]; [Judgement: 404 for a brand or a software product of a different caller](#reg-0424))
- <a id="reg-0405"></a>**REG-0405** The Register checks the path software product before the path brand. For example, a request with a different software product and a different brand gets `Field/InvalidSoftwareProduct`. ([Judgement: the Register checks the software product before the brand](#reg-0422))
- <a id="reg-0406"></a>**REG-0406** The `detail` of a path error is the path value. ([CDS: CDR Register Errors][cds-register-errors])

## Claims

- <a id="reg-0407"></a>**REG-0407** The SSA contains each claim that the SSA definition requires. ([CDS: SSA Definition][cds-ssa-definition])
- <a id="reg-0408"></a>**REG-0408** The SSA has these values. ([CDS: SSA Definition][cds-ssa-definition])
  - `iss` is `cdr-register`
  - `org_id` is the data recipient brand ID
  - `software_id` is the software product ID
- <a id="reg-0409"></a>**REG-0409** The SSA has `legal_entity_id`, `legal_entity_name` and each optional URI that the software product has. For example, an SSA has `policy_uri` only when the software product has a policy URI. ([CDS: SSA Definition][cds-ssa-definition]; [Judgement: the SSA has each optional claim that has a value](#reg-0423))
- <a id="reg-0410"></a>**REG-0410** The SSA has a `jti` that the Register does not use again. ([CDS: SSA Definition][cds-ssa-definition])
- <a id="reg-0411"></a>**REG-0411** The `scope` claim contains the scopes registered for the software product, with a space between them. The claim is the same for each `{industry}`. ([CDS: SSA Definition][cds-ssa-definition]; [Judgement: the `scope` claim has each registered scope, with no industry filter](#reg-0421))
- <a id="reg-0412"></a>**REG-0412** The SSA expires 10 minutes after the Register issues it. ([CDS: SSA][cds-ssa])

## Signature and response

- <a id="reg-0413"></a>**REG-0413** The Register signs the SSA with PS256, with a key that Get JWKS gives. ([CDS: SSA][cds-ssa]; [CDS: Get JWKS][cds-get-jwks])
- <a id="reg-0414"></a>**REG-0414** The body of a Get SSA response is the SSA as a compact JWT with no JSON quotes. The content type is `application/json`. ([CDS: Get SSA][cds-get-ssa]; [RFC 7519, section 3][rfc7519-compact]; [Judgement: the SSA response is a compact JWT with no JSON quotes](#reg-0420))
- <a id="reg-0415"></a>**REG-0415** The response of Get SSA has no `ETag`. ([CDS: Get SSA][cds-get-ssa])

## Judgements

### <a id="reg-0420"></a>REG-0420 The SSA response is a compact JWT with no JSON quotes

- Kind: conflict
- Basis: Step 3.2. The CDS OpenAPI response of Get SSA is `application/json` with a string schema ([CDS: Get SSA][cds-get-ssa]). A JSON text of a string has quotes ([RFC 8259, section 7][rfc8259-strings]). An SSA is a JWT, and a JWT is a compact string with no quotes ([RFC 7519, section 3][rfc7519-compact]).
- Reason: The Register sends the compact JWT, which a data recipient puts into its registration request with no change. The content type stays as the CDS gives it.

### <a id="reg-0421"></a>REG-0421 The `scope` claim has each registered scope, with no industry filter

- Kind: choice
- Basis: Step 3.3. The CDS does not say how the Register derives the claim ([CDS: SSA Definition][cds-ssa-definition]). A data holder ignores the scopes that it does not support ([CDS: Registration Validation][cds-registration-validation]).
- Reason: A filter by `{industry}` adds nothing, because the data holder ignores the other scopes. The registered scopes with no filter are the simplest.

### <a id="reg-0422"></a>REG-0422 The Register checks the software product before the brand

- Kind: choice
- Basis: Step 3.3. The CDS gives no sequence for the path checks ([CDS: CDR Register Errors][cds-register-errors]).
- Reason: The access token identifies the software product of the caller, so its check comes first. Each fixed sequence obeys the CDS.

### <a id="reg-0423"></a>REG-0423 The SSA has each optional claim that has a value

- Kind: choice
- Basis: Step 3.3. The CDS marks `legal_entity_id`, `legal_entity_name`, `sector_identifier_uri`, `tos_uri` and `policy_uri` as optional ([CDS: SSA Definition][cds-ssa-definition]).
- Reason: A data holder can show the legal entity name to a consumer. A claim with a value gives the data holder more information at no cost.

### <a id="reg-0424"></a>REG-0424 404 for a brand or a software product of a different caller

- Kind: conflict
- Basis: Step 1. The CDS error list gives 404, and the OpenAPI specification gives 403 "Invalid BrandId" ([CDS: CDR Register Errors][cds-register-errors]). The CDS Known Issues give 404 as the target ([CDS Known Issues: Standardise Register API Error Codes][cds-known-register-errors]).
- Reason: 404 agrees more with HTTP: the resource does not exist for this caller.

[cds-get-jwks]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-jwks
[cds-get-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-statement-assertion-ssa
[cds-known-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#standardise-register-api-error-codes
[cds-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-errors
[cds-registration-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-validation
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[rfc7519-compact]: https://www.rfc-editor.org/rfc/rfc7519#section-3
[rfc8259-strings]: https://www.rfc-editor.org/rfc/rfc8259#section-7
