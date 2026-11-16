# Mock data holder: API standards

This area gives the rules that apply to all CDR APIs of the data holder: the endpoints and their listeners, the versions and headers, the errors, CORS and transport security. The [CDS HTTP library](../../../pkg/cds/DESIGN.md) gives the shared rules. This area gives only the choices of the data holder.

## Endpoints

The first table gives the CDR APIs of the data holder. The second table gives the security endpoints. The [product reference data](product-reference-data.md), the [authorisation server](authorisation-server.md), the [registration](registration.md) and the [banking APIs](banking-apis.md) give the behaviour of each endpoint.

| ID | API | Method and path | Version | Base URI | Listener | Authentication | Source |
|---|---|---|---|---|---|---|---|
| <a id="dh-0400"></a>**DH-0400** | Get Products | `GET /cds-au/v1/banking/products` | 5 | Product | Public | None | [CDS: Get Products][cds-get-products]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0401"></a>**DH-0401** | Get Product Detail | `GET /cds-au/v1/banking/products/{productId}` | 7 | Product | Public | None | [CDS: Get Product Detail][cds-get-product-detail]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0402"></a>**DH-0402** | Get Status | `GET /cds-au/v1/discovery/status` | 1 | Public | Public | None | [CDS: Get Status][cds-get-status] |
| <a id="dh-0403"></a>**DH-0403** | Get Outages | `GET /cds-au/v1/discovery/outages` | 1 | Public | Public | None | [CDS: Get Outages][cds-get-outages] |
| <a id="dh-0404"></a>**DH-0404** | Get Customer | `GET /cds-au/v1/common/customer` | 1 | Resource | mTLS | Bearer token with `common:customer.basic:read`, holder of key | [CDS: Get Customer][cds-get-customer]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0405"></a>**DH-0405** | Get Accounts | `GET /cds-au/v1/banking/accounts` | 3 | Resource | mTLS | Bearer token with `bank:accounts.basic:read`, holder of key | [CDS: Get Accounts][cds-get-accounts]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0406"></a>**DH-0406** | Get Account Detail | `GET /cds-au/v1/banking/accounts/{accountId}` | 5 | Resource | mTLS | Bearer token with `bank:accounts.detail:read`, holder of key | [CDS: Get Account Detail][cds-get-account-detail]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0407"></a>**DH-0407** | Get Account Balance | `GET /cds-au/v1/banking/accounts/{accountId}/balance` | 1 | Resource | mTLS | Bearer token with `bank:accounts.basic:read`, holder of key | [CDS: Get Account Balance][cds-get-account-balance]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0408"></a>**DH-0408** | Get Bulk Balances | `GET /cds-au/v1/banking/accounts/balances` | 2 | Resource | mTLS | Bearer token with `bank:accounts.basic:read`, holder of key | [CDS: Get Bulk Balances][cds-get-bulk-balances]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0409"></a>**DH-0409** | Get Balances For Specific Accounts | `POST /cds-au/v1/banking/accounts/balances` | 1 | Resource | mTLS | Bearer token with `bank:accounts.basic:read`, holder of key | [CDS: Get Balances For Specific Accounts][cds-get-balances-specific]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0410"></a>**DH-0410** | Get Transactions For Account | `GET /cds-au/v1/banking/accounts/{accountId}/transactions` | 2 | Resource | mTLS | Bearer token with `bank:transactions:read`, holder of key | [CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |
| <a id="dh-0411"></a>**DH-0411** | Get Transaction Detail | `GET /cds-au/v1/banking/accounts/{accountId}/transactions/{transactionId}` | 3 | Resource | mTLS | Bearer token with `bank:transactions:read`, holder of key | [CDS: Get Transaction Detail][cds-get-transaction-detail]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480) |

| ID | Endpoint | Method and URL | Listener | Authentication | Source |
|---|---|---|---|---|---|
| <a id="dh-0412"></a>**DH-0412** | OpenID Provider configuration | `GET <issuer>/.well-known/openid-configuration` | Public | None | [CDS: OpenID Provider Configuration endpoint][cds-oidc-config] |
| <a id="dh-0413"></a>**DH-0413** | JWKS | `GET`, `jwks_uri` | Public | None | [CDS: JSON Web Key Set endpoint][cds-jwks-endpoint] |
| <a id="dh-0414"></a>**DH-0414** | Token | `POST`, `token_endpoint` | mTLS | `private_key_jwt` | [CDS: Token endpoint][cds-token-endpoint] |
| <a id="dh-0415"></a>**DH-0415** | PAR | `POST`, `pushed_authorization_request_endpoint` | mTLS | `private_key_jwt` | [CDS: PAR endpoint][cds-par-endpoint] |
| <a id="dh-0416"></a>**DH-0416** | Authorisation | `GET`, `authorization_endpoint` | Public | None. The consumer authenticates in the authorisation flow. | [CDS: Authorisation endpoint][cds-authorisation-endpoint] |
| <a id="dh-0417"></a>**DH-0417** | Introspection | `POST`, `introspection_endpoint` | mTLS | `private_key_jwt` | [CDS: Introspection endpoint][cds-introspection-endpoint] |
| <a id="dh-0418"></a>**DH-0418** | Token revocation | `POST`, `revocation_endpoint` | mTLS | `private_key_jwt` | [CDS: Token Revocation endpoint][cds-token-revocation] |
| <a id="dh-0419"></a>**DH-0419** | CDR arrangement revocation | `POST`, `cdr_arrangement_revocation_endpoint` | mTLS | `private_key_jwt` | [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation] |
| <a id="dh-0420"></a>**DH-0420** | UserInfo | `GET` or `POST`, `userinfo_endpoint` | mTLS | Bearer token, holder of key | [CDS: UserInfo endpoint][cds-userinfo-endpoint] |
| <a id="dh-0421"></a>**DH-0421** | Registration | `POST`, `registration_endpoint` | mTLS | None. The request is a signed JWT with an SSA. | [CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints] |
| <a id="dh-0422"></a>**DH-0422** | Registration management | `GET`, `PUT` or `DELETE`, `registration_endpoint` and `/{ClientId}` | mTLS | Bearer token with `cdr:registration`, holder of key | [CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints] |

- <a id="dh-0423"></a>**DH-0423** Only the OpenID Provider configuration has a fixed path. A client gets the URL of each other security endpoint from the OpenID Provider configuration. For example, the token endpoint is at the `token_endpoint` of the configuration. ([CDS: Participant Endpoints][cds-participant-endpoints]; [Judgement: no fixed path for the security endpoints](#dh-0481))
- <a id="dh-0424"></a>**DH-0424** The request and response bodies of each CDR API obey the CDS 1.36.0 OpenAPI specifications, except where a different requirement of the data holder limits them. For example, the response of Get Accounts obeys `cds_banking.json`. ([CDS OpenAPI: banking][cds-openapi-banking]; [CDS OpenAPI: common][cds-openapi-common]; [CDS OpenAPI: DCR][cds-openapi-dcr])
- <a id="dh-0425"></a>**DH-0425** The path of each API follows the CDS URI structure, after the base URI that the brand has on the Register. For example, Get Accounts is at the resource base URI and `/cds-au/v1/banking/accounts`. ([CDS: URI Structure][cds-uri-structure]; [CDR Support: Product Base URI][support-product-base-uri])
- <a id="dh-0426"></a>**DH-0426** The product base URI is the same as the public base URI. ([CDS: Participant Endpoints][cds-participant-endpoints]; [Judgement: one base URI for the public endpoints](#dh-0482))
- <a id="dh-0427"></a>**DH-0427** The resource base URI is the base URI of the mTLS listener. It is the `resourceBaseUri` of the brand on the Register. ([CDS: Participant Endpoints][cds-participant-endpoints])
- <a id="dh-0428"></a>**DH-0428** When a request on the resource base URI is for a CDS banking API that is out of scope, or for Get Customer Detail, the response is 404 with `urn:au-cds:error:cds-all:Resource/NotImplemented`. For example, Get Payees gives this response. ([CDS: Resource Not Implemented][cds-error-404-not-implemented]; [Judgement: no implementation of the other banking APIs](#dh-0483))
- <a id="dh-0429"></a>**DH-0429** The data holder does not check the token of a request for an API that is out of scope. A request with no token and a request with a token get the same response. ([Judgement: no token check for an API that is out of scope](#dh-0484))

## Versioning and headers

- <a id="dh-0430"></a>**DH-0430** The security endpoints have no version. They do not use `x-v` and `x-min-v`, and the version rules of this group do not apply to them. ([CDS: DCR APIs][cds-dcr-apis])
- <a id="dh-0431"></a>**DH-0431** Each API of the first table of Endpoints uses [the CDS version negotiation](../../../pkg/cds/DESIGN.md#cdsl-0001). Its lowest and its highest version are the version of the table. For example, a request for version 4 of Get Products gives [406 with `Header/UnsupportedVersion`](../../../pkg/cds/DESIGN.md#cdsl-0008). ([CDS: HTTP Headers][cds-http-headers]; [Judgement: only the CDS 1.36.0 version of each API](#dh-0480))
- <a id="dh-0432"></a>**DH-0432** When a request has a fault in its version headers and in its path or query parameters, the response gives the version header error. ([Judgement: the version headers before the parameters](#dh-0485))
- <a id="dh-0433"></a>**DH-0433** Each response of the public listener and the mTLS listener has [the `x-fapi-interaction-id` of the request, or a new UUID](../../../pkg/cds/DESIGN.md#cdsl-0010). ([CDS: HTTP Headers][cds-http-headers])
- <a id="dh-0434"></a>**DH-0434** When a request on the resource base URI has no `x-fapi-auth-date`, the response is 400 with `urn:au-cds:error:cds-all:Header/Missing`, and the `detail` is `x-fapi-auth-date`. ([CDS: HTTP Headers][cds-http-headers]; [Judgement: a required x-fapi-auth-date](#dh-0486))
- <a id="dh-0435"></a>**DH-0435** When the `x-fapi-auth-date` of a request on the resource base URI is not an HTTP-date, the response is 400 with `urn:au-cds:error:cds-all:Header/Invalid`, and the `detail` is `x-fapi-auth-date`. ([CDS: HTTP Headers][cds-http-headers]; [CDR Support: Format for x-fapi-auth-date][support-auth-date])
- <a id="dh-0436"></a>**DH-0436** The data holder accepts an `x-fapi-customer-ip-address` with an IPv4 or an IPv6 address. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource])
- <a id="dh-0437"></a>**DH-0437** When the `x-fapi-customer-ip-address` of a request on the resource base URI is not an IP address, the response is 400 with `urn:au-cds:error:cds-all:Header/Invalid`, and the `detail` is `x-fapi-customer-ip-address`. ([CDS: Invalid Header][cds-error-400-header-invalid])
- <a id="dh-0438"></a>**DH-0438** The data holder does not validate `x-cds-client-headers`. For example, a value that is not Base64 does not change the response. ([CDR Support: DH requirement for x-cds-client-headers][support-client-headers])
- <a id="dh-0439"></a>**DH-0439** Each response of an API on the resource base URI has a `Date` header. ([FAPI 1.0 Baseline, section 6.2.1][fapi-baseline-resource])

## Errors

- <a id="dh-0440"></a>**DH-0440** The rules of this group apply to each request to a CDR API, except to the security endpoints and the operator interface. These have their own error formats. For example, the token endpoint uses [the format of RFC 6749](../../../pkg/authserver/design/token-endpoint.md#auth-0048). ([the error format of the operator interface](operator-interface.md#dh-0733); [the error format of the registration](registration.md#dh-0648))
- <a id="dh-0441"></a>**DH-0441** When a query parameter has a value that is not valid, the response is 400 with `urn:au-cds:error:cds-all:Field/Invalid`, and the `detail` is the name of the parameter. For example, an `effective` or a `product-category` that is not in its enumeration, and [a `page` that is not a positive integer](../../../pkg/cds/DESIGN.md#cdsl-0020). ([CDS: Invalid Field][cds-error-400-field-invalid])
- <a id="dh-0442"></a>**DH-0442** When an `updated-since`, an `oldest-time` or a `newest-time` is not an RFC 3339 date-time with an offset, the response is 400 with `urn:au-cds:error:cds-all:Field/InvalidDateTime`, and the `detail` is the name of the parameter. ([CDS: Invalid Date][cds-error-400-invalid-date-time]; [CDS: Common Field Types][cds-common-field-types])
- <a id="dh-0443"></a>**DH-0443** A query parameter with an empty value is the same as a parameter that is absent. For example, `?brand=` gives the same products as no `brand`. ([an empty page parameter is absent](../../../pkg/cds/DESIGN.md#cdsl-0022); [Judgement: an empty query parameter is absent](#dh-0487))
- <a id="dh-0444"></a>**DH-0444** When the brand does not have the `productId` of the path, the response is 404 with `urn:au-cds:error:cds-all:Resource/Invalid`, and the `detail` is the `productId`. ([CDS: Invalid Resource][cds-error-404-resource-invalid])
- <a id="dh-0445"></a>**DH-0445** When the software product cannot get the `accountId` of the path with the token, the response is 404 with `urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount`, and the `detail` is the `accountId`. For example, an `accountId` that does not exist, an `accountId` of a different customer or software product, and an `accountId` of an account that is not in the arrangement. ([CDS: Invalid Banking Account][cds-error-404-invalid-banking-account]; [Judgement: one error for an account that the token does not give](#dh-0488))
- <a id="dh-0446"></a>**DH-0446** When the software product cannot get an `accountId` of the request body with the token, the response is 422 with `urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount`. The response has one error for each such `accountId`, with the `accountId` as the `detail`. ([CDS: Invalid Banking Account][cds-error-422-invalid-banking-account]; [Judgement: one error for an account that the token does not give](#dh-0488))
- <a id="dh-0447"></a>**DH-0447** When the response has an `InvalidBankingAccount` error for an `accountId` of the request body, it has no data. ([CDR Support: Error scenarios and responses][support-error-scenarios])
- <a id="dh-0448"></a>**DH-0448** When the `transactionId` of the path is not a transaction of the account, the response is 404 with `urn:au-cds:error:cds-all:Resource/Invalid`, and the `detail` is the `transactionId`. A transaction with no detail gets the same response. ([CDS: Invalid Resource][cds-error-404-resource-invalid]; [CDR Support: Error scenarios and responses][support-error-scenarios])
- <a id="dh-0449"></a>**DH-0449** When a request body is not a JSON object with the structure of its schema, the response is 400 with `urn:au-cds:error:cds-all:Field/Invalid`. For example, a body that is a JSON list. ([CDS: Invalid Field][cds-error-400-field-invalid])
- <a id="dh-0450"></a>**DH-0450** When a required field of a request body is absent, the response is 400 with `urn:au-cds:error:cds-all:Field/Missing`. ([CDS: Missing Required Field][cds-error-400-field-missing])
- <a id="dh-0451"></a>**DH-0451** The `detail` of an error of a request body is the name of the field. When the body is not JSON, the `detail` is `data`. For example, a body with no `accountIds` gives the `detail` `data.accountIds`. ([CDS: Invalid Field][cds-error-400-field-invalid]; [Judgement: data as the detail of a body that is not JSON](#dh-0489))
- <a id="dh-0452"></a>**DH-0452** The data holder ignores a field of a request body that the schema does not give. ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]; [Judgement: no refusal of a field that the schema does not give](#dh-0490))
- <a id="dh-0453"></a>**DH-0453** When a request body is more than 64 KiB, the response is 413 with `urn:au-cds:error:cds-all:GeneralError/Expected`. ([RFC 9110, section 15.5.14][rfc9110-413]; [Judgement: a request body of 64 KiB at most](#dh-0491); [Judgement: GeneralError/Expected for 413 and 415](#dh-0492))
- <a id="dh-0454"></a>**DH-0454** When a `POST` request has a `Content-Type` that is not `application/json`, the response is 415 with `urn:au-cds:error:cds-all:GeneralError/Expected`. ([CDS: HTTP Headers][cds-http-headers]; [CDS: HTTP Response Codes][cds-http-response-codes]; [Judgement: GeneralError/Expected for 413 and 415](#dh-0492))
- <a id="dh-0455"></a>**DH-0455** When the path of a request is not known, the response is 404 with `urn:au-cds:error:cds-all:Resource/NotFound`. ([CDS: Resource Not Found][cds-error-404-not-found])

## CORS

- <a id="dh-0456"></a>**DH-0456** Each endpoint on the public listener uses [the CORS rules](../../../pkg/cds/DESIGN.md#cdsl-0037), also for [a preflight request](../../../pkg/cds/DESIGN.md#cdsl-0039), except the authorisation endpoint and the pages of its flow. For example, Get Products, the OpenID Provider configuration and the JWKS use the CORS rules. ([CDS: CORS][cds-cors]; [Judgement: CORS on the OpenID Provider configuration and the JWKS](#dh-0493))
- <a id="dh-0457"></a>**DH-0457** The CORS rules of the public listener [expose these headers](../../../pkg/cds/DESIGN.md#cdsl-0038), so that a web page can read them. ([CDR Support: CORS support][support-cors])
  - `x-v`
  - `Retry-After`
  - `x-fapi-interaction-id`
- <a id="dh-0458"></a>**DH-0458** The endpoints on the mTLS listener do not permit cross-origin requests. For example, their responses have no `Access-Control-Allow-Origin` header. ([Judgement: no CORS on the mTLS listener](#dh-0494))
- <a id="dh-0459"></a>**DH-0459** The authorisation endpoint and the pages of its flow do not permit cross-origin requests. ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]; [Judgement: no CORS on the authorisation pages](#dh-0495))

## Transport security

- <a id="dh-0460"></a>**DH-0460** Each CDR listener uses [the TLS versions](../../../pkg/cds/DESIGN.md#cdsl-0041) and [the TLS 1.2 cipher suites](../../../pkg/cds/DESIGN.md#cdsl-0042) of the CDS HTTP library. ([CDS: Transaction Security][cds-transaction-security])
- <a id="dh-0461"></a>**DH-0461** The public listener presents a server certificate from the public CA, not from the CDR CA. ([CDS: Participant Endpoints][cds-participant-endpoints]; [a TLS endpoint uses a certificate from the public CA](../../../pki/DESIGN.md#pki-0012))
- <a id="dh-0462"></a>**DH-0462** The mTLS listener is separate from the public listener. ([CDS: Participant Endpoints][cds-participant-endpoints]; [Judgement: the issuer on the public listener](authorisation-server.md#dh-0530))
- <a id="dh-0463"></a>**DH-0463** The mTLS listener presents a server certificate from the CDR CA. ([CDS: Certificate Trust Model][cds-cert-trust-model])
- <a id="dh-0464"></a>**DH-0464** The mTLS listener accepts only [a client certificate that chains to the CDR CA bundle](../../../pkg/cds/DESIGN.md#cdsl-0044) of the data holder configuration. ([CDS: Certificate Trust Model][cds-cert-trust-model]; [the CDR CA bundle of an mTLS listener](../../../pkg/cds/DESIGN.md#cdsl-0045))
- <a id="dh-0465"></a>**DH-0465** When a client gives no client certificate, the mTLS listener ends the TLS handshake. ([CDS: Transaction Security][cds-transaction-security]; [Judgement: the TLS findings of the OpenID conformance suite](#dh-0497))
- <a id="dh-0466"></a>**DH-0466** The data holder checks the revocation status of each client certificate with [the CRL check of the CDS HTTP library](../../../pkg/cds/DESIGN.md#cdsl-0048). It [refuses the certificate when the check fails](../../../pkg/cds/DESIGN.md#cdsl-0047). ([a revocation check for an mTLS listener](../../../pkg/cds/DESIGN.md#cdsl-0046))
- <a id="dh-0467"></a>**DH-0467** The [maximum CRL age](../../../pkg/cds/DESIGN.md#cdsl-0052) is an optional setting of the data holder. When the setting is absent, the maximum age is [the default of 1 minute](../../../pkg/cds/DESIGN.md#cdsl-0053). ([Judgement: a maximum CRL age of 1 minute](../../../pkg/cds/DESIGN.md#cdsl-0068))
- <a id="dh-0468"></a>**DH-0468** In the lab, the maximum CRL age is 0. The data holder then [loads the CRL at each check](../../../pkg/cds/DESIGN.md#cdsl-0054). ([Judgement: a CRL load at each check in the lab](#dh-0496))
- <a id="dh-0469"></a>**DH-0469** The data holder gets these values from its configuration. ([the CDR CA bundle of an mTLS listener](../../../pkg/cds/DESIGN.md#cdsl-0045); [a real client changes the data holder with configuration only](non-functional-requirements.md#dh-0013))
  - the base URI of the mTLS listener
  - the server certificate and key of the mTLS listener
  - the CA bundle for client certificates

## Judgements

### <a id="dh-0480"></a>DH-0480 Only the CDS 1.36.0 version of each API

- Kind: choice
- Basis: Step 3.3. Data holders must implement the versions of CDS 1.36.0 by 9 November 2026, and they can retire the earlier versions from 7 December 2026 ([CDS: Future Dated Obligations][cds-future-dated]). They can retire version 4 of Get Products and version 6 of Get Product Detail from 10 August 2026 ([CDS: Get Products][cds-get-products], [CDS: Get Product Detail][cds-get-product-detail]).
- Reason: The lab has no data recipient that needs an earlier version. The earlier versions are only on pages that the CDS calls obsolete.

### <a id="dh-0481"></a>DH-0481 No fixed path for the security endpoints

- Kind: choice
- Basis: Step 3.2. The CDS gives a path only for the OpenID Provider configuration ([CDS: Participant Endpoints][cds-participant-endpoints]). The DCR OpenAPI specification has the server URL `https://mtls.dh.example.com/cds-au/v1`, and the examples of the security profile use `/register` ([CDS OpenAPI: DCR][cds-openapi-dcr]).
- Reason: A client that reads the URLs from the configuration works with each path.

### <a id="dh-0482"></a>DH-0482 One base URI for the public endpoints

- Kind: choice
- Basis: Step 3.3. The CDS permits a product base URI that is the same as the public base URI ([CDS: Participant Endpoints][cds-participant-endpoints]).
- Reason: One base URI for all public endpoints is the least complex.

### <a id="dh-0483"></a>DH-0483 No implementation of the other banking APIs

- Kind: choice
- Basis: Step 3.3. The CDS says that `Resource/NotImplemented` "SHOULD be supported for unimplemented APIs" ([CDS: Resource Not Implemented][cds-error-404-not-implemented]).
- Reason: Payees, direct debits, scheduled payments and instalment plans add data types and no new behaviour. Get Customer Detail needs addresses and phone numbers, and the lab needs no more personal information.

### <a id="dh-0484"></a>DH-0484 No token check for an API that is out of scope

- Kind: choice
- Basis: Step 3.3. An API that the data holder does not implement has no scope to check.
- Reason: The response tells a caller with no token only that the API is not implemented. A token check before the response adds code with no benefit.

### <a id="dh-0485"></a>DH-0485 The version headers before the parameters

- Kind: choice
- Basis: Step 3.3. The CDS gives no sequence for the checks of a request ([CDS: HTTP Headers][cds-http-headers]).
- Reason: One sequence for all CDR APIs is the least complex.

### <a id="dh-0486"></a>DH-0486 A required x-fapi-auth-date

- Kind: conflict
- Basis: Step 1. The CDS text says that the header is "Required for all resource calls" ([CDS: HTTP Headers][cds-http-headers]). The OpenAPI specification marks it as not required. A DSB article says that its use is "at the discretion of the DH", but that data holders "must validate that the mandatory headers are provided" ([CDR Support: FAPI header obligations for Data Holders][support-fapi-headers]).
- Reason: The CDS text is the first source.

### <a id="dh-0487"></a>DH-0487 An empty query parameter is absent

- Kind: choice
- Basis: Step 3.3. The CDS does not specify a query parameter with an empty value ([CDS: Pagination][cds-pagination]).
- Reason: Each parameter of Get Products is optional, and a client that builds a query from a form can send a name with no value. The CDS HTTP library gives [the same rule for `page` and `page-size`](../../../pkg/cds/DESIGN.md#cdsl-0063).

### <a id="dh-0488"></a>DH-0488 One error for an account that the token does not give

- Kind: choice
- Basis: Step 3.3. The CDS permits `InvalidBankingAccount` or `UnavailableBankingAccount` for an account that is not in the arrangement ([CDS: Invalid Banking Account][cds-error-404-invalid-banking-account]).
- Reason: For this token, the account is never available. One response for an account that does not exist and an account that is not in the arrangement does not show which accounts exist.

### <a id="dh-0489"></a>DH-0489 data as the detail of a body that is not JSON

- Kind: choice
- Basis: Step 3.3. The CDS `detail` of a body error is the name of the field ([CDS: Invalid Field][cds-error-400-field-invalid]).
- Reason: A body that is not JSON has no field. `data` is the top-level field of each request schema.

### <a id="dh-0490"></a>DH-0490 No refusal of a field that the schema does not give

- Kind: choice
- Basis: Step 3.2. The request schemas do not forbid other properties, and `meta` is an open object ([CDS: Get Balances For Specific Accounts][cds-get-balances-specific]).
- Reason: A refusal refuses a request that the schema accepts.

### <a id="dh-0491"></a>DH-0491 A request body of 64 KiB at most

- Kind: choice
- Basis: Step 3.2. The CDS gives no limit. 413 is the HTTP status for a body that is too large ([RFC 9110, section 15.5.14][rfc9110-413]).
- Reason: 1000 `accountId` values are approximately 40 KB. The other request bodies of the data holder have the same limit.

### <a id="dh-0492"></a>DH-0492 GeneralError/Expected for 413 and 415

- Kind: choice
- Basis: Step 3.3. The CDS gives 415 for a media type that is not supported, and no error code for 413 or 415 ([CDS: HTTP Response Codes][cds-http-response-codes]).
- Reason: `GeneralError/Expected` is the CDS code for an expected error with no more specific code ([CDS: Error Codes][cds-error-codes]).

### <a id="dh-0493"></a>DH-0493 CORS on the OpenID Provider configuration and the JWKS

- Kind: choice
- Basis: Step 3.3. The CDS says that these endpoints do "not require CORS", which does not forbid the header ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]).
- Reason: One behaviour for all endpoints on the public listener is the least complex.

### <a id="dh-0494"></a>DH-0494 No CORS on the mTLS listener

- Kind: choice
- Basis: Step 3.2. The CDS applies CORS to the APIs with no authentication ([CDS: CORS][cds-cors]).
- Reason: The endpoints on the mTLS listener are authenticated endpoints. A web page in a browser has no CDR client certificate.

### <a id="dh-0495"></a>DH-0495 No CORS on the authorisation pages

- Kind: choice
- Basis: Step 3.2. The CDS says that the authorisation endpoint does "not require CORS" ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]).
- Reason: The authorisation endpoint and the pages of its flow are pages for a consumer and not an API. A different web site must not read them.

### <a id="dh-0496"></a>DH-0496 A CRL load at each check in the lab

- Kind: choice
- Basis: Step 3.3. The [default maximum CRL age](../../../pkg/cds/DESIGN.md#cdsl-0068) delays a revocation by up to 1 minute.
- Reason: A conformance test revokes a certificate and then checks the refusal at once, because the suite does not wait ([conformance DESIGN.md][conformance-design]). The CRL server of the lab is in the lab network, so a load at each check costs little.

### <a id="dh-0497"></a>DH-0497 The TLS findings of the OpenID conformance suite

- Kind: conflict
- Basis: Step 1. The OpenID conformance suite probes TLS 1.2 on the resource endpoint with no client certificate, and it reports the closed connection as a failure of the conditions `DisallowInsecureCipher` and `CheckForBCP195InsecureFAPICiphers`. The CDS requires a client certificate on this endpoint ([CDS: Transaction Security][cds-transaction-security]).
- Reason: The CDS is the first source, so the data holder accepts these two findings. A handshake with a cipher suite that the CDS does not permit also fails.

[cds-get-products]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-products
[cds-get-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-product-detail
[cds-get-status]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-status
[cds-get-outages]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-outages
[cds-get-customer]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-customer
[cds-get-accounts]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-accounts
[cds-get-account-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-detail
[cds-get-account-balance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-account-balance
[cds-get-bulk-balances]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-bulk-balances
[cds-get-balances-specific]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-balances-for-specific-accounts
[cds-get-transactions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account
[cds-get-transaction-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transaction-detail
[cds-oidc-config]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#openid-provider-configuration-endpoint
[cds-jwks-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#json-web-key-set-endpoint
[cds-token-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#token-endpoint
[cds-par-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pushed-authorisation-endpoint
[cds-authorisation-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-endpoint
[cds-introspection-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#introspection-endpoint
[cds-token-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#token-revocation-endpoint
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-userinfo-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#userinfo-endpoint
[cds-dcr-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dynamic-client-registration-endpoints
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-openapi-banking]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_banking.json
[cds-openapi-common]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_common.json
[cds-openapi-dcr]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_dcr.json
[cds-uri-structure]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#uri-structure
[support-product-base-uri]: https://cdr-support.zendesk.com/hc/en-us/articles/16505555811855
[cds-error-404-not-implemented]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-resource-not-implemented
[cds-dcr-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dcr-apis
[cds-http-headers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-headers
[support-auth-date]: https://cdr-support.zendesk.com/hc/en-us/articles/4405773134863
[fapi-baseline-resource]: https://openid.net/specs/openid-financial-api-part-1-1_0.html#protected-resources-provisions
[cds-error-400-header-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-header-invalid
[support-client-headers]: https://cdr-support.zendesk.com/hc/en-us/articles/4404953530767
[cds-error-400-field-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid
[cds-error-400-invalid-date-time]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid-date-time
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-error-404-resource-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-resource-invalid
[cds-error-404-invalid-banking-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-authorisation-invalid-banking-account
[cds-error-422-invalid-banking-account]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-422-authorisation-invalid-banking-account
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[cds-error-400-field-missing]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-missing
[rfc9110-413]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.14
[cds-http-response-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-response-codes
[cds-error-404-not-found]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-resource-not-found
[cds-cors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors
[support-cors]: https://cdr-support.zendesk.com/hc/en-us/articles/900003054706
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-cert-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-future-dated]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#future-dated-obligations
[support-fapi-headers]: https://cdr-support.zendesk.com/hc/en-us/articles/4410664891919
[cds-pagination]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[conformance-design]: ../../conformance/DESIGN.md
