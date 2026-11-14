# Observed behaviour of the production CDR Register (public endpoints)

- Research date: 2026-09-25 (about 14:00 to 14:30 UTC)
- Applies to: the production Register at `https://api.cdr.gov.au`, which served CDS 1.36.0 versions on that date
- Method: read-only HTTP GET requests with httpie to the public (TLS, unauthenticated) endpoints. About 40 requests in total. The authenticated mTLS endpoints (`secure.api.cdr.gov.au`) cannot be observed without CDR credentials.

Observed behaviour changes without notice. Check again before you depend on a fact.

## 1. Infrastructure

- Responses have `x-azure-ref` and `Request-Context` headers. INFERENCE: the Register runs on Azure behind Azure Front Door, with Application Insights.
- Responses have `Strict-Transport-Security: max-age=2592000`.
- All responses, including errors, have `Access-Control-Allow-Origin: *`. This is true when the request has no `Origin` header.

## 2. OpenID Provider configuration

`GET /idp/.well-known/openid-configuration` returned exactly these fields:

```json
{
  "issuer": "https://api.cdr.gov.au/idp",
  "jwks_uri": "https://api.cdr.gov.au/idp/.well-known/openid-configuration/jwks",
  "token_endpoint": "https://secure.api.cdr.gov.au/idp/connect/token",
  "claims_supported": ["sub"],
  "code_challenge_methods_supported": ["plain", "S256"],
  "id_token_signing_alg_values_supported": ["PS256"],
  "subject_types_supported": ["public"],
  "scopes_supported": ["cdr-register:read"],
  "response_types_supported": ["token"],
  "grant_types_supported": ["client_credentials"],
  "token_endpoint_auth_methods_supported": ["private_key_jwt"],
  "tls_client_certificate_bound_access_tokens": true,
  "token_endpoint_auth_signing_alg_values_supported": ["PS256", "ES256"]
}
```

- The document has all fields that the OpenAPI schema marks as required. It does not have `request_object_signing_alg_values_supported`, which the CDS example shows.
- The token endpoint accepts PS256 and ES256 client assertions. This settles an item that the CDS text does not state.
- Headers: `Cache-Control: public, max-age=21600`. No `ETag`.

## 3. Register JWKS

`GET /cdr-register/v1/jwks`:

- Two keys. Each key has `kty` `RSA`, `alg` `PS256`, `key_ops` `["sign","verify"]` and no `use`.
- Each `kid` is 32 lower-case hex characters (for example `5f18ac8c13564d05abd034b755fe80f2`).
- Each modulus is 2048 bits (342 base64url characters).
- INFERENCE: two keys means that a key rotation is in progress, or that the Register publishes a standby key.
- Headers: `ETag` (32 upper-case hex characters, no quotes), `Cache-Control: public, max-age=3600`.

## 4. Version negotiation

Requests to `GET /cdr-register/v1/all/data-recipients/status` (current version 3):

| Request | Response |
|---|---|
| `x-v: 3` | 200, `x-v: 3` |
| `x-v: 99` | 406, `urn:au-cds:error:cds-all:Header/UnsupportedVersion`, detail "Requested version is lower than the minimum version or greater than maximum version." |
| `x-v: 99`, `x-min-v: 2` | 200, `x-v: 3` |
| `x-v: abc` | 400, `urn:au-cds:error:cds-all:Header/InvalidVersion`, detail "Version is not a positive Integer." |
| no `x-v` | 400, `urn:au-cds:error:cds-all:Header/Missing`, detail "An API version x-v header is required, but was not specified." |

- The production Register still serves deprecated versions. `GET /cdr-register/v1/banking/data-holders/brands/summary` with `x-v: 1` returned 200 with `x-v: 1`.
- Error bodies have no `meta` object.

## 5. Industry path parameter

| Request | Response |
|---|---|
| `foo/data-holders/brands/summary` (x-v 2) | 404, `urn:au-cds:error:cds-register:Field/InvalidIndustry`, title "Invalid Industry", detail "Industry URL parameter is invalid" |
| `foo/data-holders/status` (x-v 2) | 400, `urn:au-cds:error:cds-all:Field/Invalid`, title "Invalid Field" |
| `foo/data-recipients` (x-v 4) | 400, `urn:au-cds:error:cds-all:Field/Invalid` |
| `foo/data-recipients/status` (x-v 3) | 400, `urn:au-cds:error:cds-all:Field/Invalid` |
| `foo/data-recipients/brands/software-products/status` (x-v 3) | 400, `urn:au-cds:error:cds-all:Field/Invalid` |
| `banking/data-recipients` (x-v 4) | 400, `urn:au-cds:error:cds-all:Field/Invalid`, response has `x-v: 4` |
| `banking/data-recipients/status` (x-v 3) | 400, `urn:au-cds:error:cds-all:Field/Invalid`, detail "Invalid Path Parameter", response has `x-v: 3` and an `ETag` |
| `banking/data-recipients/brands/software-products/status` (x-v 3) | 400, `urn:au-cds:error:cds-all:Field/Invalid` |
| `energy`, `non-bank-lending`, `telco`, `all` on `data-holders/brands/summary` (x-v 2) | 200 |
| `all`, `energy` on `data-holders/status` (x-v 2) | 200 |

- Get Data Holder Brands Summary is different from the other endpoints: an invalid industry gives 404 `cds-register:Field/InvalidIndustry`. This agrees with the OpenAPI specification for that endpoint ("404 Industry Not Found").
- All the other public endpoints give 400 `cds-all:Field/Invalid`. This includes a valid industry value that the endpoint does not allow (for example `banking` on data recipient endpoints).
- `telco` is accepted, although no telco data holders exist.

## 6. Caching (ETag, If-None-Match, Cache-Control)

- The `ETag` value is 32 upper-case hex characters with **no quotes**, for example `ETag: D598736B65F8A0BB2A35FCA5FB460648`. RFC 9110 requires an entity tag in quotes.
- `If-None-Match` with the unquoted value returned 304. `If-None-Match` with the quoted value returned 200. The comparison is an exact string match on the unquoted value.
- INFERENCE: an HTTP client that sends RFC 9110 entity tags in quotes does not get 304 responses from the production Register.
- The same ETag was returned for the same data in repeated requests.
- `Cache-Control: public, max-age=210` on Get Data Holder Brands Summary and on the status endpoints. 210 seconds is inside the 2 to 5 minute poll period that the CDS recommends.
- A 400 response from a status endpoint also had an `ETag`.

## 7. x-fapi-interaction-id

- A request with `x-fapi-interaction-id` got a response with no `x-fapi-interaction-id` header. The production Register does not echo it.

## 8. Response data (shape and content)

Get Data Holder Brands Summary (`all`, x-v 2):

- 335 brands. 283 have `dataHolderBrandId`. 91 have `interimId`. Some brands have both. 52 brands have only `interimId`.
- Fields seen: `abn`, `acn`, `brandGroup`, `brandName`, `dataHolderBrandId`, `industries`, `interimId`, `lastUpdated`, `logoUri`, `productBaseUri`, `publicBaseUri`.
- Industry counts: `banking` 116, `energy` 84, `non-bank-lending` 135.
- `links` has only `self`, with the full URL. `meta` is `{}`.

Get Data Recipients (`all`, x-v 4):

- 38 data recipient legal entities: 30 ACTIVE, 8 SURRENDERED.
- Brands: 33 ACTIVE, 8 REMOVED.
- Software products: 187 ACTIVE, 4 INACTIVE, 75 REMOVED.
- The response includes entities that are not ACTIVE.

Get Data Holder Statuses (`all`, x-v 2):

- 115 data holder legal entities: 98 ACTIVE, 17 REMOVED. The response includes REMOVED entities. (The ACCC mock returns only ACTIVE entities.)

## 9. What this changes for the mock Register

- The error for an invalid industry depends on the endpoint (section 5).
- The Register does not echo `x-fapi-interaction-id`.
- ETags have no quotes, and `If-None-Match` needs the unquoted value.
- `Cache-Control` values are 210 seconds (summary and statuses), 3600 seconds (JWKS) and 21600 seconds (OpenID Provider configuration).
- The OpenID Provider configuration has the exact field set in section 2.
- Status and recipient lists include entities that are not ACTIVE.
- Brands Summary includes brands that have only an `interimId`.
