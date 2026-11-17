# Mock data holder: Registration

This area gives the registration of a software product with an SSA, registration management, and the reaction of the data holder to participant statuses. The [authorisation server library](../../../pkg/authserver/design/README.md) gives the client authentication and the reaction at the PAR, authorisation and token endpoints.

## Registration request

- <a id="dh-0600"></a>**DH-0600** A software product registers with `POST` to the `registration_endpoint`. The body is one compact JWT, with the content type `application/jwt`. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0601"></a>**DH-0601** When the content type of a registration request is not `application/jwt`, the data holder responds with 415 and no body. ([CDS: HTTP Response Codes][cds-http-response-codes]; [Judgement: no body for a 415 response](#dh-0700))
- <a id="dh-0602"></a>**DH-0602** The data holder accepts a request JWT that is signed with PS256 or ES256. ([CDS: Registration Request using JWT][cds-registration-request]; [FAPI 1.0 Advanced, section 8.6][fapi-algorithms])
- <a id="dh-0603"></a>**DH-0603** A request JWT has these claims. ([CDS: Registration Request using JWT][cds-registration-request])
  - `iss`, `iat`, `exp`, `jti` and `aud`
  - `token_endpoint_auth_method` and `token_endpoint_auth_signing_alg`
  - `grant_types` and `response_types`
  - `id_token_signed_response_alg` and `authorization_signed_response_alg`
  - `software_statement`
- <a id="dh-0604"></a>**DH-0604** The `aud` of a request JWT is the issuer, or a list that contains the issuer. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0605"></a>**DH-0605** The `iss` of a request JWT is the `software_id` of the SSA. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0606"></a>**DH-0606** The data holder refuses a request JWT that is expired. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0607"></a>**DH-0607** The data holder refuses a request JWT whose `exp` is more than 10 minutes after the time of the request. ([Judgement: a request JWT of 10 minutes at most, with no reuse for 1 hour](#dh-0701))
- <a id="dh-0608"></a>**DH-0608** The data holder refuses a request JWT whose `jti` it accepted in the last hour, from any software product. For example, a second `POST` with the same request JWT gets `invalid_client_metadata`. ([CDS: Registration Request using JWT][cds-registration-request]; [Judgement: a request JWT of 10 minutes at most, with no reuse for 1 hour](#dh-0701))
- <a id="dh-0609"></a>**DH-0609** The data holder refuses a registration request body that is larger than 64 KiB. ([Judgement: a registration request of 64 KiB at most](#dh-0702))

## Validation of the SSA

- <a id="dh-0610"></a>**DH-0610** The data holder verifies the SSA signature with the key of the Register JWKS that has the `kid` of the SSA. ([CDS: Registration Validation][cds-registration-validation]; [CDS: Software Statement Assertion][cds-ssa])
- <a id="dh-0611"></a>**DH-0611** The SSA algorithm is PS256. ([CDS: Software Statement Assertion][cds-ssa])
- <a id="dh-0612"></a>**DH-0612** The data holder gets the Register JWKS from a URL in its configuration. For example, the setting `REGISTER_JWKS_URL` gives the URL. ([CDS: Registration Validation][cds-registration-validation])
- <a id="dh-0613"></a>**DH-0613** For the connection to the Register JWKS, the data holder trusts only the public CAs in its configuration. ([CDS: Transaction Security][cds-transaction-security])
- <a id="dh-0614"></a>**DH-0614** When the SSA has a `kid` that the data holder does not have, the data holder gets the Register JWKS again before it refuses the SSA. It gets the JWKS again one time each second at most. ([Judgement: a new get of the Register JWKS for an unknown key ID](#dh-0703))
- <a id="dh-0615"></a>**DH-0615** The `iss` of the SSA is `cdr-register`. ([CDS: SSA Definition][cds-ssa-definition])
- <a id="dh-0616"></a>**DH-0616** The data holder refuses an SSA that is expired. ([CDS: SSA Definition][cds-ssa-definition])
- <a id="dh-0617"></a>**DH-0617** The SSA has each claim that the SSA definition marks as required. For example, the data holder refuses an SSA with no `revocation_uri`. ([CDS: SSA Definition][cds-ssa-definition])
- <a id="dh-0618"></a>**DH-0618** The `software_roles` of the SSA is `data-recipient-software-product`. ([CDS: SSA Definition][cds-ssa-definition])
- <a id="dh-0619"></a>**DH-0619** The `jwks_uri` of the SSA is an `https` URL. When it is not, the data holder refuses the SSA. ([CDS: Transaction Security][cds-transaction-security]; [CDR Support: Egress policy for calls from DH to ADRs][support-egress])
- <a id="dh-0620"></a>**DH-0620** The data holder refuses an SSA whose `jti` it accepted before, until the SSA expires, with 400 and `invalid_software_statement`. For example, a `PUT` with the SSA of the registration gets this refusal. ([CDS: SSA Definition][cds-ssa-definition]; [Judgement: one use of each SSA, until it expires](#dh-0704))
- <a id="dh-0621"></a>**DH-0621** At registration, the data holder refuses only a software product whose status is REMOVED. For example, a software product that is INACTIVE, or that has no status at the data holder, can register. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: only the status REMOVED stops a registration](#dh-0705))

## Validation of the request

- <a id="dh-0622"></a>**DH-0622** The data holder verifies the signature of the request JWT with the keys at the `jwks_uri` of the SSA. ([CDS: Registration Validation][cds-registration-validation])
- <a id="dh-0623"></a>**DH-0623** The data holder connects to the `jwks_uri` on any host, and trusts only the public CAs in its configuration for that connection. It has no allowlist of data recipient hosts. ([CDR Support: Egress policy for calls from DH to ADRs][support-egress]; [CDS: Transaction Security][cds-transaction-security])
- <a id="dh-0624"></a>**DH-0624** The `token_endpoint_auth_method` is `private_key_jwt`. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0625"></a>**DH-0625** The `token_endpoint_auth_signing_alg` and the `request_object_signing_alg` are `PS256` or `ES256`. ([CDS: Registration Request using JWT][cds-registration-request]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](../../../pkg/authserver/design/configuration.md#auth-0013))
- <a id="dh-0626"></a>**DH-0626** The `id_token_signed_response_alg` and the `authorization_signed_response_alg` are `PS256`. ([CDS: Registration Request using JWT][cds-registration-request]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](../../../pkg/authserver/design/configuration.md#auth-0013))
- <a id="dh-0627"></a>**DH-0627** When the request has an `application_type`, it is `web`. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0628"></a>**DH-0628** Each value of `grant_types` is `client_credentials`, `authorization_code` or `refresh_token`. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0629"></a>**DH-0629** The `response_types` contains `code`. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0630"></a>**DH-0630** The data holder registers only the response type `code`, also when the request has `code id_token`. ([CDS: Registration Request using JWT][cds-registration-request]; [Judgement: only code in response_types](#dh-0706))
- <a id="dh-0631"></a>**DH-0631** When the request has `redirect_uris`, each value is in the `redirect_uris` of the SSA. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0632"></a>**DH-0632** The data holder permits one registration for each `software_id`. ([CDS: Registration Validation][cds-registration-validation])

## Registration values

- <a id="dh-0633"></a>**DH-0633** The data holder issues a new client ID for each registration. The client ID is a random RFC 4122 UUID, and it is not the `software_id`. ([CDS: Registration Response][cds-registration-response])
- <a id="dh-0634"></a>**DH-0634** The data holder does not issue the client ID of a deleted registration again. ([CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed])
- <a id="dh-0635"></a>**DH-0635** The registration takes these values from the SSA. ([CDS: Registration Response][cds-registration-response]; [RFC 7591, section 2.3][rfc7591-software-statement])
  - `client_name`, `client_description`, `client_uri` and `logo_uri`
  - `org_id` and `org_name`
  - `redirect_uris`, `jwks_uri`, `revocation_uri` and `recipient_base_uri`
  - `software_id`, `software_roles` and `scope`
  - `legal_entity_id`, `legal_entity_name`, `sector_identifier_uri`, `tos_uri` and `policy_uri`, when the SSA has them
- <a id="dh-0636"></a>**DH-0636** A claim of the request with the name of a value from the SSA has no effect, except `redirect_uris`. For example, a `client_name` in the request does not change the `client_name` of the registration. ([RFC 7591, section 2.3][rfc7591-software-statement])
- <a id="dh-0637"></a>**DH-0637** The data holder does not get the document at the `sector_identifier_uri`. ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]; [Judgement: no check of the sector identifier document](#dh-0707))
- <a id="dh-0638"></a>**DH-0638** The `scope` of the registration has the scopes of the SSA that the data holder supports, in the sequence of the SSA. ([CDS: Registration Validation][cds-registration-validation])
- <a id="dh-0639"></a>**DH-0639** The `redirect_uris` of the registration are those of the request. When the request has none, they are those of the SSA. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0640"></a>**DH-0640** When the request has no `request_object_signing_alg`, the registration has `PS256`. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0641"></a>**DH-0641** When the request has no `application_type`, the registration has `web`. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dh-0642"></a>**DH-0642** The data holder ignores each claim of the request that it does not support, with no error. For example, it ignores the claims for the encryption of ID tokens and of authorisation responses. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]; [Judgement: no encryption of ID tokens or authorisation responses](../../../pkg/authserver/design/configuration.md#auth-0017))

## Registration response

- <a id="dh-0643"></a>**DH-0643** A successful registration gives 201 with the registration as `application/json`. ([CDS: Registration Response][cds-registration-response]; [RFC 7591, section 3.2.1][rfc7591-response])
- <a id="dh-0644"></a>**DH-0644** The response has `client_id`, `client_id_issued_at` and each value of the registration. ([CDS: Registration Response][cds-registration-response])
- <a id="dh-0645"></a>**DH-0645** The `software_statement` of the response is the SSA of the request, with no change. ([CDS: Registration Response][cds-registration-response])
- <a id="dh-0646"></a>**DH-0646** The response has no claim that the data holder ignored. ([CDS: Registration Response][cds-registration-response])
- <a id="dh-0647"></a>**DH-0647** The response has no `registration_access_token` and no `registration_client_uri`. ([CDS: RegistrationProperties][cds-schema-registration-properties])

## Registration errors

- <a id="dh-0648"></a>**DH-0648** When a registration request is not valid, the data holder responds with 400 and an `application/json` body with the members `error` and `error_description`. ([CDS: Registration Errors][cds-registration-errors]; [RFC 7591, section 3.2.2][rfc7591-errors])
- <a id="dh-0649"></a>**DH-0649** The `error_description` names the claim or the rule that fails. For example, "the SSA has no revocation_uri". ([CDS: Registration Errors][cds-registration-errors])

The data holder uses these `error` values:

| ID | Condition | `error` | Source |
|---|---|---|---|
| <a id="dh-0650"></a>**DH-0650** | The SSA is absent or is not a JWT, its signature is not valid, its `kid` is not in the Register JWKS, it is expired, or one of its claims is absent or not valid | `invalid_software_statement` | [CDS: Registration Errors][cds-registration-errors]; [Judgement: invalid_software_statement for a fault in the SSA and for a duplicate](#dh-0708) |
| <a id="dh-0651"></a>**DH-0651** | The `software_id` of the SSA has a registration | `invalid_software_statement` | [CDS: Registration Errors][cds-registration-errors]; [Judgement: invalid_software_statement for a fault in the SSA and for a duplicate](#dh-0708) |
| <a id="dh-0652"></a>**DH-0652** | A value of `redirect_uris` is not in the SSA | `invalid_redirect_uri` | [CDS: Registration Errors][cds-registration-errors]; [Judgement: invalid_software_statement for a fault in the SSA and for a duplicate](#dh-0708) |
| <a id="dh-0653"></a>**DH-0653** | The body is not a JWT or is larger than 64 KiB, the request JWT has a signature, an `aud`, an `iss`, an `exp`, a `jti` or a client metadata value that is not valid, or a required claim is absent | `invalid_client_metadata` | [CDS: Registration Errors][cds-registration-errors]; [Judgement: invalid_software_statement for a fault in the SSA and for a duplicate](#dh-0708) |
| <a id="dh-0654"></a>**DH-0654** | The data holder cannot get the keys at the `jwks_uri` of the SSA | `invalid_client_metadata` | [CDS: Registration Errors][cds-registration-errors]; [Judgement: invalid_software_statement for a fault in the SSA and for a duplicate](#dh-0708) |

- <a id="dh-0655"></a>**DH-0655** The data holder uses the error `unapproved_software_statement` only for [a software product that is REMOVED](#dh-0693). ([RFC 7591, section 3.2.2][rfc7591-errors]; [Judgement: unapproved_software_statement only for a REMOVED software product](#dh-0709))
- <a id="dh-0656"></a>**DH-0656** When the data holder cannot get the Register JWKS, it responds with 500 and the error `server_error`. ([RFC 7591, section 3.2.2][rfc7591-errors]; [Judgement: server_error when the Register JWKS is not available](#dh-0710))

## Registration management

- <a id="dh-0657"></a>**DH-0657** `GET`, `PUT` and `DELETE` at the `registration_endpoint` and `/{ClientId}` read, replace and delete a registration. ([CDS: Registration Management][cds-registration-management])
- <a id="dh-0658"></a>**DH-0658** Each registration management request has an access token as a bearer token. ([RFC 6750][rfc6750]; [CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints])
- <a id="dh-0659"></a>**DH-0659** The access token is active, has the scope `cdr:registration`, and is bound to the client certificate of the request. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]; [CDS: Holder of Key Mechanism][cds-hok])
- <a id="dh-0660"></a>**DH-0660** The access token is a token of the client in the path. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints])
- <a id="dh-0661"></a>**DH-0661** When the data holder refuses a registration management request, the response has no body. ([`cds_dcr.json`][cds-openapi-dcr]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711))

The data holder gives these refusals:

| ID | Condition | Response | Source |
|---|---|---|---|
| <a id="dh-0662"></a>**DH-0662** | The request has no bearer token | 401 with `WWW-Authenticate: Bearer` | [RFC 6750, section 3.1][rfc6750-errors]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0663"></a>**DH-0663** | The token is not known, is expired, is revoked, is not bound to the client certificate, or is a token of a deleted registration | 401 with `WWW-Authenticate: Bearer error="invalid_token"` | [RFC 6750, section 3.1][rfc6750-errors]; [RFC 8705, section 3][rfc8705-binding]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0664"></a>**DH-0664** | The path has a client ID that does not exist | 401 with `WWW-Authenticate: Bearer error="invalid_token"` | [`cds_dcr.json`][cds-openapi-dcr]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0665"></a>**DH-0665** | The token does not have the scope `cdr:registration` | 403 with `WWW-Authenticate: Bearer error="insufficient_scope", scope="cdr:registration"` | [RFC 6750, section 3.1][rfc6750-errors]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0666"></a>**DH-0666** | The path has the client ID of a different registration | 403 | [`cds_dcr.json`][cds-openapi-dcr]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |

- <a id="dh-0667"></a>**DH-0667** `GET` gives 200 with the registration, in the format of the registration response. ([CDS: RegistrationProperties][cds-schema-registration-properties])
- <a id="dh-0668"></a>**DH-0668** The `software_statement` of a `GET` response is the SSA of the last request that the data holder accepted. ([CDS: RegistrationProperties][cds-schema-registration-properties])
- <a id="dh-0669"></a>**DH-0669** A `PUT` has the same body as a registration request, with a new SSA. The same validation rules and errors apply. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0670"></a>**DH-0670** When the `software_id` of the SSA of a `PUT` is not that of the registration, the response is 400 with `invalid_software_statement`. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0671"></a>**DH-0671** A `PUT` replaces all values of the registration with the values that the request and its SSA give. The client ID and `client_id_issued_at` do not change. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0672"></a>**DH-0672** A successful `PUT` gives 200 with the registration. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0673"></a>**DH-0673** A `DELETE` deletes the registration and gives 204. ([CDS: DCR APIs][cds-dcr-apis]; [Judgement: support for the deletion of a registration](#dh-0712))
- <a id="dh-0674"></a>**DH-0674** After a `DELETE`, client authentication with the client ID of the registration fails. ([CDS: DCR APIs][cds-dcr-apis])
- <a id="dh-0675"></a>**DH-0675** After a `DELETE`, the software product can register again. It then gets a new client ID. ([Judgement: support for the deletion of a registration](#dh-0712))
- <a id="dh-0676"></a>**DH-0676** When a software product deletes its registration, each arrangement of the client that did not end, ends at the time of the request. The end is a [revocation by the data recipient](end-of-an-arrangement.md#dh-0902), with no notice to the data recipient. ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]; [Judgement: a deletion ends the arrangements of the client as a revocation](#dh-0713))
- <a id="dh-0677"></a>**DH-0677** After a `DELETE`, the data holder accepts no token of the client. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints])

## Participant statuses

- <a id="dh-0679"></a>**DH-0679** The data holder gets the participant statuses of the software products from [Get Software Products Statuses][cds-get-sp-statuses] of the Register, with the `{industry}` `all`. ([CDS: Metadata Cache Management][cds-metadata-cache]; [Judgement: only the statuses of software products](#dh-0714))
- <a id="dh-0680"></a>**DH-0680** The data holder gets the base URI of the public APIs of the Register from its configuration. For example, the setting `REGISTER_PUBLIC_BASE_URL` gives the base URI. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0681"></a>**DH-0681** For the connection to the public APIs of the Register, the data holder trusts only the public CAs in its configuration. ([CDS: Transaction Security][cds-transaction-security])
- <a id="dh-0682"></a>**DH-0682** The software product of a registration is the software product with the `software_id` of the registration. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="dh-0683"></a>**DH-0683** The data holder gets the statuses at an interval from its configuration. The interval is 4 minutes at most. For example, the setting `REGISTER_STATUS_INTERVAL` is `1m`. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0684"></a>**DH-0684** A status change on the Register has its effect at the data holder within 5 minutes. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0685"></a>**DH-0685** The request for the statuses has `If-None-Match` with the `ETag` of the last response. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0686"></a>**DH-0686** When the Register does not respond, the data holder keeps the previous status of each software product, with no time limit. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="dh-0687"></a>**DH-0687** When the response has no status, or a status that is not known, for a software product, the data holder keeps the previous status of that software product. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="dh-0688"></a>**DH-0688** When the data holder has no status for a software product, the software product gets the responses for INACTIVE. ([Judgement: INACTIVE for a software product with no status](#dh-0715))
- <a id="dh-0689"></a>**DH-0689** The readiness probe fails until the data holder gets the statuses one time. ([the readiness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0002); [Judgement: INACTIVE for a software product with no status](#dh-0715))
- <a id="dh-0690"></a>**DH-0690** When the data holder gets the status REMOVED for a software product, it ends each arrangement of each registration of that software product at once. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]; [Judgement: the clean-up at once when a software product is REMOVED](#dh-0716))
- <a id="dh-0691"></a>**DH-0691** When the data holder gets the status REMOVED for a software product, it deletes each registration of that software product. Client authentication with the client ID then fails, and the data holder accepts no token of the client. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]; [Judgement: the clean-up at once when a software product is REMOVED](#dh-0716))
- <a id="dh-0692"></a>**DH-0692** The data holder sends no notice to the data recipient for an arrangement that a REMOVED status ends. ([CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed])
- <a id="dh-0693"></a>**DH-0693** When the status of the software product of the SSA is REMOVED, the data holder refuses a registration request with 400 and the error `unapproved_software_statement`. For example, an SSA that the Register issued before the status REMOVED cannot register the software product again. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]; [Judgement: only the status REMOVED stops a registration](#dh-0705))

## Reaction to participant statuses

The authorisation server library gives the reaction at [the PAR endpoint](../../../pkg/authserver/design/authorisation-request.md#auth-0110), [the authorisation endpoint](../../../pkg/authserver/design/authorisation-request.md#auth-0111) and [the token endpoint](../../../pkg/authserver/design/authorisation-request.md#auth-0112), the [endpoints that the status does not change](../../../pkg/authserver/design/authorisation-request.md#auth-0113), and [the sequence of the status check at the PAR and token endpoints](../../../pkg/authserver/design/authorisation-request.md#auth-0114).

- <a id="dh-0694"></a>**DH-0694** The participant status of the data recipient does not change the reaction of the data holder. Only the status of the software product changes it. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: only the statuses of software products](#dh-0714))

| ID | Request or item | ACTIVE | INACTIVE | Source |
|---|---|---|---|---|
| <a id="dh-0695"></a>**DH-0695** | APIs on the resource base URI | Permitted | 403 with `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`, and the `detail` is the status of the software product | [CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: unauthorized_client, not invalid_grant, for a software product that is not ACTIVE](../../../pkg/authserver/design/authorisation-request.md#auth-0124) |
| <a id="dh-0696"></a>**DH-0696** | UserInfo endpoint | Permitted | 401 with `WWW-Authenticate: Bearer error="invalid_token"` and no body | [CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: unauthorized_client, not invalid_grant, for a software product that is not ACTIVE](../../../pkg/authserver/design/authorisation-request.md#auth-0124) |
| <a id="dh-0697"></a>**DH-0697** | Registration management | Permitted | Permitted | [CDS: Data Holder Responsibilities][cds-dh-responsibilities] |
| <a id="dh-0698"></a>**DH-0698** | Arrangements | No change | No change. Sharing starts again when the software product is ACTIVE again. | [CDS: Data Holder Responsibilities][cds-dh-responsibilities] |

- <a id="dh-0699"></a>**DH-0699** On the resource base URI and the UserInfo endpoint, the status check comes after the check of the access token, and before the other checks of the request. ([CDR Support: Error scenarios and responses][support-error-scenarios])

## Judgements

### <a id="dh-0700"></a>DH-0700 No body for a 415 response

- Kind: choice
- Basis: Step 3.3. The CDS gives the status code for a media type that is not supported, and no body for it ([CDS: HTTP Response Codes][cds-http-response-codes]).
- Reason: No registration error code of RFC 7591 fits the condition. The status code is sufficient for a client.

### <a id="dh-0701"></a>DH-0701 A request JWT of 10 minutes at most, with no reuse for 1 hour

- Kind: choice
- Basis: Step 3.3. The CDS says only that the `jti` prevents the reuse of the token, and gives no lifetime ([CDS: Registration Request using JWT][cds-registration-request]).
- Reason: A short limit keeps the window for a stolen request small. A request is not valid for longer than 10 minutes, so a memory of 1 hour adds nothing. The [authorisation server library uses the same limit for a client assertion](../../../pkg/authserver/design/client-authentication.md#auth-0033).

### <a id="dh-0702"></a>DH-0702 A registration request of 64 KiB at most

- Kind: choice
- Basis: Step 3.3. The CDS gives no size limit for a registration request ([CDS: Registration Request using JWT][cds-registration-request]).
- Reason: An SSA and the client metadata are a few kilobytes. A limit protects the data holder from a large body.

### <a id="dh-0703"></a>DH-0703 A new get of the Register JWKS for an unknown key ID

- Kind: choice
- Basis: Step 3.3. The CDS requires the check of the SSA signature with the Register JWKS, and gives no cache rule for it ([CDS: Registration Validation][cds-registration-validation]).
- Reason: The Register can change its key, and a registration must then not need a restart of the data holder. The limit stops an SSA with an invented `kid` from making a call to the Register for each request. The Register publishes a new key before it signs with it, so a JWKS from after that time has the key.

### <a id="dh-0704"></a>DH-0704 One use of each SSA, until it expires

- Kind: choice
- Basis: Step 1. The SSA definition says that the `jti` is "used to prevent reuse of the SSA", and gives no time limit and no error code for a reuse ([CDS: SSA Definition][cds-ssa-definition]).
- Reason: The data holder refuses an expired SSA, so it keeps each `jti` only until the SSA expires. A software product gets a new SSA from the Register for each registration request, also for a `PUT`. With one use, each registration request shows a recent approval of the Register.

### <a id="dh-0705"></a>DH-0705 Only the status REMOVED stops a registration

- Kind: choice
- Basis: Step 2. The CDS says that the statuses "SHOULD be referenced" for registration management, and gives no response for a registration request ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). ACCC guidance says that the status REMOVED is final, and that the client ID of a REMOVED software product "cannot be reactivated" ([CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]).
- Reason: An SSA stays valid after the Register changes the status, so only a check at the data holder stops a new registration of a REMOVED software product. An INACTIVE software product can become ACTIVE again, and the [reaction to participant statuses](#dh-0694) stops its sharing. Because no registration can occur after REMOVED, the clean-up at the change to REMOVED is sufficient, and the data holder needs no other clean-up.

### <a id="dh-0706"></a>DH-0706 Only code in response_types

- Kind: choice
- Basis: Step 3.2. The CDS schema lists `code id_token`, but data recipients use only `code` ([CDS: Registration Request using JWT][cds-registration-request]). An authorisation server can replace a requested value ([RFC 7591, section 2][rfc7591-metadata]).
- Reason: A data recipient that sends the two values can still register. The response shows what is registered.

### <a id="dh-0707"></a>DH-0707 No check of the sector identifier document

- Kind: gap
- Basis: Step 2. ACCC guidance says that a data holder gets the document at the `sector_identifier_uri` at `POST` and `PUT`, and refuses the request when it cannot get it ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]). OpenID Connect requires that the document contains each registered redirect URI ([OpenID Connect Core, section 8.1][oidc-core-pairwise]).
- Reason: The data holder registers the `sector_identifier_uri` as the SSA gives it, and uses only its host for the pairwise `sub`. An older DSB article with the same procedure is under review, but that does not withdraw the guidance.

### <a id="dh-0708"></a>DH-0708 invalid_software_statement for a fault in the SSA and for a duplicate

- Kind: choice
- Basis: Step 3.2. The CDS gives the four codes of RFC 7591, and no table of conditions ([CDS: Registration Errors][cds-registration-errors]). For a duplicate registration, the CDS requires 400 and gives no code. Its example, which is not normative, uses `invalid_software_statement`.
- Reason: A fault in the SSA gives `invalid_software_statement`, and so does a duplicate, because the `software_id` comes from the SSA. A redirect URI that the SSA does not have gives `invalid_redirect_uri`. Each other fault of the request gives `invalid_client_metadata`.

### <a id="dh-0709"></a>DH-0709 unapproved_software_statement only for a REMOVED software product

- Kind: choice
- Basis: Step 3.2. The CDS lists the four codes of RFC 7591, and gives no condition for `unapproved_software_statement` ([CDS: Registration Errors][cds-registration-errors]). RFC 7591 gives it for an SSA that the authorisation server does not approve ([RFC 7591, section 3.2.2][rfc7591-errors]).
- Reason: The SSA of a REMOVED software product has a valid signature and valid claims, so `invalid_software_statement` is not correct. The data holder does not approve it, because the status REMOVED is final. The data holder approves each other valid SSA of the Register.

### <a id="dh-0710"></a>DH-0710 server_error when the Register JWKS is not available

- Kind: choice
- Basis: Step 3.2. RFC 7591 gives no error code for a fault that is not in the request ([RFC 7591, section 3.2.2][rfc7591-errors]).
- Reason: The cause is not the request, so a 400 code would be wrong. `server_error` with 500 is the usual OAuth 2.0 response for a fault of the server.

### <a id="dh-0711"></a>DH-0711 401 or 403 with no body for registration management errors

- Kind: choice
- Basis: Step 3.2. The DCR OpenAPI specification gives 401 for an "unknown or invalid Client or invalid access token" and 403 when the client "does not have permission", with no body schema ([`cds_dcr.json`][cds-openapi-dcr]). [RFC 6750, section 3.1][rfc6750-errors] gives the error values. A certificate that does not match gives `invalid_token` ([RFC 8705, section 3][rfc8705-binding]).
- Reason: The status codes follow the CDS specification, and the `WWW-Authenticate` values follow the RFCs that it uses. A response with no body agrees with the CDS specification, which has no body schema.

### <a id="dh-0712"></a>DH-0712 Support for the deletion of a registration

- Kind: choice
- Basis: Step 3.3. The CDS marks `DELETE` of a registration as optional ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]).
- Reason: A lab user and the conformance suite can then register a software product again with no reset of the lab.

### <a id="dh-0713"></a>DH-0713 A deletion ends the arrangements of the client as a revocation

- Kind: choice
- Basis: Step 2. The CDS gives no rule for the arrangements of a registration that the software product deletes ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]). ACCC guidance says: "After an ADR registration deletion request there should be no active consents for that ADR" ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]).
- Reason: The software product asks for the deletion, so the end has the reason and the effects of a revocation by the data recipient. The data holder sends no notice, because the data recipient made the request. The end at the time of the request agrees with a revocation at the CDR arrangement revocation endpoint.

### <a id="dh-0714"></a>DH-0714 Only the statuses of software products

- Kind: choice
- Basis: Step 3.3. The CDS says that the data holder "SHOULD" use the statuses of the data recipients and of the software products ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). The reactions are the same for each data recipient status with the same software product status, because the Register changes the status of each software product when its data recipient changes ([CDS: Status Mapping][cds-status-mapping]).
- Reason: One API gives the same behaviour with one cache.

### <a id="dh-0715"></a>DH-0715 INACTIVE for a software product with no status

- Kind: choice
- Basis: Step 3.1. The CDS says to keep the previous status, and gives no rule for a software product with no previous status ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]).
- Reason: Data must not go to a software product whose status is not known. INACTIVE stops the sharing and changes no arrangement. A registration needs an SSA from the Register, so each registered software product has a status after the next request to the Register.

### <a id="dh-0716"></a>DH-0716 The clean-up at once when a software product is REMOVED

- Kind: choice
- Basis: Step 3.3. The CDS permits a batch, for example each night ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). The ACCC also permits it each time the Register cache is refreshed ([CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]).
- Reason: One step at the time of the status is less complex than a separate batch. A lab test then does not wait for a batch.

[cds-registration-request]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-request-using-jwt
[cds-registration-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-validation
[cds-registration-response]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-response
[cds-registration-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-errors
[cds-registration-management]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-management
[cds-put-operations]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#put-operations
[cds-schema-registration-properties]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-dynamic-client-registration-api_schemas_tocSregistrationproperties
[cds-http-response-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-response-codes
[cds-dcr-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dynamic-client-registration-endpoints
[cds-dcr-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dcr-apis
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-get-sp-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-products-statuses
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-openapi-dcr]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_dcr.json
[support-dcr]: https://cdr-support.zendesk.com/hc/en-us/articles/6828251436431
[support-egress]: https://cdr-support.zendesk.com/hc/en-us/articles/4403543156751
[support-removed]: https://cdr-support.zendesk.com/hc/en-us/articles/900001657906
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[oidc-core-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[rfc6750]: https://www.rfc-editor.org/rfc/rfc6750
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc7591-metadata]: https://www.rfc-editor.org/rfc/rfc7591#section-2
[rfc7591-software-statement]: https://www.rfc-editor.org/rfc/rfc7591#section-2.3
[rfc7591-response]: https://www.rfc-editor.org/rfc/rfc7591#section-3.2.1
[rfc7591-errors]: https://www.rfc-editor.org/rfc/rfc7591#section-3.2.2
[rfc8705-binding]: https://www.rfc-editor.org/rfc/rfc8705#section-3
