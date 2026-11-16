# Authorisation server library: Authorisation request

This area gives the authorisation flow of the data holder authorisation server: the pushed authorisation request, the request object, the authorisation endpoint and the authorisation response. It also gives the status checks of the software product at these endpoints and at the token endpoint. The service gives the consumer pages and the status of each client.

## Pushed authorisation request

- <a id="auth-0061"></a>**AUTH-0061** A client starts an authorisation flow with `POST` to the PAR endpoint. The body is a form with the client assertion parameters and a `request` parameter, which is a signed request object. ([CDS: Pushed Authorisation endpoint][cds-par-endpoint]; [RFC 9126, section 3][rfc9126-request-object])
- <a id="auth-0062"></a>**AUTH-0062** The authorisation server uses only the parameters in the request object. It ignores each other authorisation parameter of the form. ([FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 10)
- <a id="auth-0063"></a>**AUTH-0063** The authorisation server refuses a pushed authorisation request with a `request_uri` parameter. ([RFC 9126, section 2.1][rfc9126-request])
- <a id="auth-0064"></a>**AUTH-0064** When the registration of the client does not have the grant type `authorization_code`, the response is 400 with `invalid_grant`. ([Judgement: the error codes of the PAR endpoint](#auth-0115))
- <a id="auth-0065"></a>**AUTH-0065** A successful request gives 201 with `request_uri`, `expires_in` and `Cache-Control: no-store`. ([RFC 9126, section 2.2][rfc9126-response])
- <a id="auth-0066"></a>**AUTH-0066** The `request_uri` expires 60 seconds after the authorisation server issues it, and `expires_in` is 60. ([CDS: Request Object Submission][cds-request-object-submission]; [Judgement: the lifetime of a request_uri](#auth-0116))
- <a id="auth-0067"></a>**AUTH-0067** The `request_uri` has a random part that a client cannot guess. ([RFC 9126, section 7.1][rfc9126-guessing])
- <a id="auth-0068"></a>**AUTH-0068** An error response has the format of RFC 9126, section 2.3. A failed client authentication gets [the refusal of the client authentication](client-authentication.md#auth-0032). ([RFC 9126, section 2.3][rfc9126-errors])
- <a id="auth-0069"></a>**AUTH-0069** Each other error gets 400 with `error` and `error_description`. ([RFC 9126, section 2.3][rfc9126-errors]; [Judgement: the error codes of the PAR endpoint](#auth-0115))
- <a id="auth-0070"></a>**AUTH-0070** For a 400 response, the `error_description` names the claim or the rule that fails. It does not repeat a value of the request object. ([RFC 9126, section 2.3][rfc9126-errors]; [Judgement: the error codes of the PAR endpoint](#auth-0115))
- <a id="auth-0071"></a>**AUTH-0071** For a 401 response, the `error_description` does not give the cause. ([RFC 9126, section 2.3][rfc9126-errors])

## Request object

- <a id="auth-0072"></a>**AUTH-0072** The request object is a JWT that the client signs with the `request_object_signing_alg` of its registration. ([CDS: Request Object][cds-request-object])
- <a id="auth-0073"></a>**AUTH-0073** The authorisation server accepts a request object that is signed with PS256 or ES256. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms]; [Judgement: the signing algorithms](configuration.md#auth-0013))
- <a id="auth-0074"></a>**AUTH-0074** The authorisation server verifies the signature of the request object with the keys at the JWKS URI of the client. ([CDS: Request Object][cds-request-object])
- <a id="auth-0075"></a>**AUTH-0075** The `iss` and the `client_id` of the request object are the client ID of the authenticated client. ([FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 15)
- <a id="auth-0076"></a>**AUTH-0076** The `aud` of the request object is the issuer, or a list that contains the issuer. ([FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 15)
- <a id="auth-0077"></a>**AUTH-0077** The request object has an `nbf` and an `exp`. ([CDS: Request Object, Data Holders][cds-request-object-dh]; [FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], items 13 and 17)
- <a id="auth-0078"></a>**AUTH-0078** The `nbf` is not more than 60 minutes before the time of the request. ([FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 17)
- <a id="auth-0079"></a>**AUTH-0079** The `exp` is not more than 60 minutes after the `nbf`. ([CDS: Request Object, Data Holders][cds-request-object-dh]; [FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 13)
- <a id="auth-0080"></a>**AUTH-0080** The authorisation server refuses a request object that is expired. ([FAPI 1.0 Advanced, section 5.2.2][fapi-advanced-server], item 13)
- <a id="auth-0081"></a>**AUTH-0081** The `response_type` is `code`. ([CDS: Baseline Security Provisions][cds-baseline-security])
- <a id="auth-0082"></a>**AUTH-0082** The `response_mode` is `jwt`. ([FAPI 1.0 Advanced, section 5.2.2.2][fapi-advanced-jarm]; [Judgement: only the jwt response mode](#auth-0117))
- <a id="auth-0083"></a>**AUTH-0083** The `redirect_uri` is present, and it is equal to one of the `redirect_uris` of the registration. ([FAPI 1.0 Baseline, section 5.2.2][fapi-baseline-server], items 9 and 10)
- <a id="auth-0084"></a>**AUTH-0084** The `scope` contains `openid`. ([CDS: Scopes and Claims][cds-scopes-and-claims])
- <a id="auth-0085"></a>**AUTH-0085** When a scope of the request object is not in the `scope` of the registration, the response is 400 with `invalid_scope`. ([CDR Support: Guidance on Dynamic Client Registration][support-dcr])
- <a id="auth-0086"></a>**AUTH-0086** When the `scope` of the request object contains `cdr:registration`, the response is 400 with `invalid_scope`. ([CDS: Authorisation Scopes][cds-authorisation-scopes])
- <a id="auth-0087"></a>**AUTH-0087** The `nonce` is present. The `state` is optional. ([FAPI 1.0 Baseline, section 5.2.2.2][fapi-baseline-nonce])
- <a id="auth-0088"></a>**AUTH-0088** The `code_challenge` is present, and the `code_challenge_method` is `S256`. ([CDS: Request Object, Data Holders][cds-request-object-dh]; [RFC 7636, section 4.2][rfc7636-challenge])
- <a id="auth-0089"></a>**AUTH-0089** When `claims.sharing_duration` is present, it is an integer of seconds that is 0 or more. The authorisation server refuses a negative value and a value that is not an integer. ([CDS: Requesting Sharing Duration][cds-sharing-duration])
- <a id="auth-0090"></a>**AUTH-0090** For a `claims.sharing_duration` of more than one year, the authorisation server uses one year (31,536,000 seconds). ([CDS: Requesting Sharing Duration][cds-sharing-duration])
- <a id="auth-0091"></a>**AUTH-0091** The authorisation server refuses a request object with `claims.cdr_arrangement_id`. ([CDS: Specifying an existing arrangement][cds-existing-arrangement]; [Judgement: no amendment of an arrangement](#auth-0120))
- <a id="auth-0092"></a>**AUTH-0092** When `claims.id_token.acr` is essential and gives values, one of the values is `urn:cds.au:cdr:2`. If not, the authorisation server refuses the request. ([CDS: Levels of Assurance][cds-loa]; [CDR Support: One Time Password (OTP)][support-otp]; [Judgement: one level of assurance](#auth-0121))
- <a id="auth-0093"></a>**AUTH-0093** The authorisation server ignores each other member of `claims`, and each other parameter that this area does not name. ([CDS: Claims][cds-claims])
- <a id="auth-0094"></a>**AUTH-0094** These faults give 400 with `invalid_request_object`. ([Judgement: the error codes of the PAR endpoint](#auth-0115))
  - a fault of the JWT: its signature, `iss`, `client_id`, `aud`, `nbf` or `exp`
  - a fault of a member of `claims`
- <a id="auth-0095"></a>**AUTH-0095** These faults give 400 with `invalid_request`. ([Judgement: the error codes of the PAR endpoint](#auth-0115))
  - a fault of `response_type`, `response_mode`, `redirect_uri`, `nonce`, `code_challenge` or `code_challenge_method`
  - a `scope` with no `openid`

## Authorisation endpoint

- <a id="auth-0096"></a>**AUTH-0096** The client sends the browser of the consumer to the authorisation endpoint with `client_id` and `request_uri` in the query. The authorisation server needs no other parameter, and it ignores each other parameter. ([RFC 9126, section 4][rfc9126-authorisation]; [CDR Support: Pushed Authorization Requests redirect_uri parameter][support-par-redirect-uri])
- <a id="auth-0097"></a>**AUTH-0097** The authorisation server refuses a request with a `request` parameter, and a request with no `request_uri`. ([CDS: Request Object, Data Holders][cds-request-object-dh])
- <a id="auth-0098"></a>**AUTH-0098** The authorisation server accepts a `request_uri` one time only. It refuses a `request_uri` that it does not know, that is expired, or that it accepted before. ([CDS: Request Object, Data Holders][cds-request-object-dh]; [RFC 9126, section 4][rfc9126-authorisation]; [Judgement: one use of a request_uri at the first visit](#auth-0118))
- <a id="auth-0099"></a>**AUTH-0099** The authorisation server refuses a `request_uri` of a different client with 401. ([CDR Support: Error scenarios and responses][support-error-scenarios])
- <a id="auth-0100"></a>**AUTH-0100** When the authorisation server refuses a request to the authorisation endpoint, it does not redirect the browser. ([RFC 6749, section 4.1.2.1][rfc6749-authorisation-errors]; [Judgement: no redirect for a refused authorisation request](#auth-0119))
- <a id="auth-0101"></a>**AUTH-0101** For a refusal with no redirect, the service shows its error page with the status 400. The status is 401 for a `request_uri` of a different client and for a client ID that is not registered. ([CDR Support: Error scenarios and responses][support-error-scenarios]; [Judgement: no redirect for a refused authorisation request](#auth-0119))
- <a id="auth-0102"></a>**AUTH-0102** An authorisation flow that ended does not continue. The authorisation server refuses a request for one of its pages [with no redirect](#auth-0100). ([Judgement: no redirect for a refused authorisation request](#auth-0119))
- <a id="auth-0103"></a>**AUTH-0103** An authorisation flow ends 10 minutes after the authorisation request, when the consumer does not complete it. The authorisation server then refuses each request of the flow [with no redirect](#auth-0100). ([Judgement: the time limit of an authorisation flow](#auth-0122))

## Authorisation response

- <a id="auth-0104"></a>**AUTH-0104** At the end of an authorisation flow, the authorisation server redirects the browser to the `redirect_uri` of the request object, with one query parameter, `response`. ([CDS: Authorization Code Flow][cds-code-flow]; [JARM, section 2.3.1][jarm-query]; [Judgement: only the jwt response mode](#auth-0117))
- <a id="auth-0105"></a>**AUTH-0105** The `response` is a JWT that the authorisation server signs with PS256. ([JARM, section 2.1][jarm-response]; [Judgement: the signing algorithms](configuration.md#auth-0013))
- <a id="auth-0106"></a>**AUTH-0106** The JWT has `iss` (the issuer), `aud` (the client ID) and `exp`, which is 10 minutes after the time of the response. ([JARM, section 2.1][jarm-response]; [Judgement: the lifetime of an authorisation code and of the authorisation response](#auth-0123))
- <a id="auth-0107"></a>**AUTH-0107** When the consumer gives the authorisation, the JWT has `code`. For an error, it has `error` and `error_description`. It has `state` when the request object has one. ([JARM, section 2.1][jarm-response])
- <a id="auth-0108"></a>**AUTH-0108** The authorisation response is not encrypted. ([CDS: Authorization Code Flow][cds-code-flow]; [Judgement: no encryption of ID tokens or authorisation responses](configuration.md#auth-0017))
- <a id="auth-0109"></a>**AUTH-0109** The authorisation code is valid for 60 seconds and for one use. ([RFC 6749, section 4.1.2][rfc6749-code]; [Judgement: the lifetime of an authorisation code and of the authorisation response](#auth-0123))

## Status of the software product

The service gives the authorisation server the status of the software product of each client: ACTIVE or not ACTIVE.

- <a id="auth-0110"></a>**AUTH-0110** When the software product is not ACTIVE, the PAR endpoint responds with 400 and `unauthorized_client`. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: the error for a software product that is not ACTIVE](#auth-0124))
- <a id="auth-0111"></a>**AUTH-0111** When the software product is not ACTIVE at a request to the authorisation endpoint or to a page of the flow, the flow ends with the error `unauthorized_client` in [the authorisation response](#auth-0104). The consumer pages do not get the request. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Error scenarios and responses][support-error-scenarios]; [Judgement: the error for a software product that is not ACTIVE](#auth-0124))
- <a id="auth-0112"></a>**AUTH-0112** When the software product is not ACTIVE, the token endpoint responds to the `authorization_code` and `refresh_token` grants with 400 and `unauthorized_client`. The refresh token does not change. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [RFC 6749, section 5.2][rfc6749-errors]; [Judgement: the error for a software product that is not ACTIVE](#auth-0124))
- <a id="auth-0113"></a>**AUTH-0113** The status of the software product does not change the `client_credentials` grant, introspection, token revocation and CDR arrangement revocation. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="auth-0114"></a>**AUTH-0114** At the PAR and token endpoints, the status check comes after the client authentication. At these endpoints and at the authorisation endpoint, other checks of the request can come before the status check. ([Judgement: the sequence of the status check](#auth-0125))

## Judgements

### <a id="auth-0115"></a>AUTH-0115 The error codes of the PAR endpoint

- Kind: choice
- Basis: Step 3.3. The CDS gives no codes. RFC 9126 permits each code of RFC 6749, sections 5.2 and 4.1.2.1 ([RFC 9126, section 2.3][rfc9126-errors]), and OpenID Connect adds `invalid_request_object` ([OpenID Connect Core, section 3.1.2.6][oidc-core-auth-error]).
- Reason: The codes of the underlying OpenID Provider library are the least complex. A client reads the `error_description` to find its fault.

### <a id="auth-0116"></a>AUTH-0116 The lifetime of a request_uri

- Kind: choice
- Basis: Step 3.3. The CDS permits 10 to 90 seconds ([CDS: Request Object Submission][cds-request-object-submission]). The CDS example has 3600, and it is not normative.
- Reason: The client redirects the browser at once, and 60 seconds permits a slow browser.

### <a id="auth-0117"></a>AUTH-0117 Only the jwt response mode

- Kind: choice
- Basis: Step 3.3. FAPI 1.0 Advanced requires `jwt` with the response type `code` ([section 5.2.2.2][fapi-advanced-jarm]). For that response type, `jwt` gives the response in the query ([JARM, section 2.3.4][jarm-jwt]). The CDS does not name the other JARM modes.
- Reason: One mode is the least complex.

### <a id="auth-0118"></a>AUTH-0118 One use of a request_uri at the first visit

- Kind: conflict
- Basis: Step 1. The CDS requires that the data holder "MUST reject the reuse of request_uri values" ([CDS: Request Object, Data Holders][cds-request-object-dh]). FAPI recommends one use at the completion of the flow, and the OpenID conformance suite tests that recommendation.
- Reason: The CDS wins. Thus a second visit with a `request_uri` gets the error page, also before the consumer completes the flow.

### <a id="auth-0119"></a>AUTH-0119 No redirect for a refused authorisation request

- Kind: choice
- Basis: Step 3.2. RFC 6749 forbids a redirect when the client or the redirect URI is not valid ([section 4.1.2.1][rfc6749-authorisation-errors]).
- Reason: Without a valid `request_uri`, the authorisation server has no request. Thus it has no redirect URI that it can trust.

### <a id="auth-0120"></a>AUTH-0120 No amendment of an arrangement

- Kind: gap
- Basis: Step 1. The CDS requires support for `cdr_arrangement_id` in a request object, for the amendment of a consent ([CDS: Specifying an existing arrangement][cds-existing-arrangement]).
- Reason: Amendment is out of scope. A refusal before the consumer authenticates is better than a new arrangement that the client did not ask for.

### <a id="auth-0121"></a>AUTH-0121 One level of assurance

- Kind: choice
- Basis: Step 2. An OTP is one factor ([CDS: Levels of Assurance][cds-loa]). When a data holder cannot reach an essential level of assurance, "the request should be rejected" ([CDR Support: One Time Password (OTP)][support-otp]).
- Reason: All CDR data APIs are read operations, which need `urn:cds.au:cdr:2`. Thus the authorisation server supports only that value, and refuses a request that needs a different value as essential.

### <a id="auth-0122"></a>AUTH-0122 The time limit of an authorisation flow

- Kind: choice
- Basis: Step 3.3. The CDS gives no limit.
- Reason: 10 minutes gives time for two OTPs, and the limit removes flows that a consumer left. After the limit, the authorisation server cannot be sure that the browser is that of the consumer, so it does not redirect.

### <a id="auth-0123"></a>AUTH-0123 The lifetime of an authorisation code and of the authorisation response

- Kind: choice
- Basis: Step 3.2. The CDS gives no lifetime. RFC 6749 recommends 10 minutes at most for a code ([section 4.1.2][rfc6749-code]), and JARM recommends 10 minutes at most for the response ([section 2.1][jarm-response]).
- Reason: A client exchanges the code at once, so 60 seconds is sufficient. The 10 minutes of the response is the value of the underlying OpenID Provider library, which is the least complex.

### <a id="auth-0124"></a>AUTH-0124 The error for a software product that is not ACTIVE

- Kind: choice
- Basis: Step 3.2. The CDS says only that the data holder does not "Facilitate Consent Authorisation" or disclose data ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). The DSB says that the authorisation endpoint redirects with an error, and gives no code ([CDR Support: Error scenarios and responses][support-error-scenarios]). RFC 6749 gives `unauthorized_client` for a client that is not permitted to do the request ([section 4.1.2.1][rfc6749-authorisation-errors], [section 5.2][rfc6749-errors]).
- Reason: `invalid_grant` at the token endpoint would tell the client that the authorisation ended. A software product that is not ACTIVE keeps its arrangements.

### <a id="auth-0125"></a>AUTH-0125 The sequence of the status check

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no sequence for these endpoints. The ACCC sequence is for an access token ([CDR Support: Error scenarios and responses][support-error-scenarios]).
- Reason: Each sequence refuses the request, and the Register API gives the status to each person, so the sequence discloses nothing. The underlying OpenID Provider library does its checks of a request before the status check. A different sequence would need a second client authentication.

[cds-par-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pushed-authorisation-endpoint
[cds-request-object]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-object
[cds-request-object-dh]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#security-profile_request-object_data-holders
[cds-request-object-submission]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-object-submission
[cds-existing-arrangement]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#specifying-an-existing-arrangement
[cds-baseline-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#baseline-security-provisions
[cds-code-flow]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorization-code-flow
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
[cds-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#claims
[cds-loa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#levels-of-assurance-loas
[cds-scopes-and-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#scopes-and-claims
[cds-authorisation-scopes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-scopes
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[support-dcr]: https://cdr-support.zendesk.com/hc/en-us/articles/6828251436431
[support-otp]: https://cdr-support.zendesk.com/hc/en-us/articles/6140553082255
[support-par-redirect-uri]: https://cdr-support.zendesk.com/hc/en-us/articles/900004240446
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[fapi-advanced-server]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#authorization-server
[fapi-advanced-jarm]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#jarm
[fapi-baseline-server]: https://openid.net/specs/openid-financial-api-part-1-1_0.html#authorization-server
[fapi-baseline-nonce]: https://openid.net/specs/openid-financial-api-part-1-1_0.html#client-requesting-openid-scope
[jarm-response]: https://openid.net/specs/oauth-v2-jarm.html#section-2.1
[jarm-query]: https://openid.net/specs/oauth-v2-jarm.html#section-2.3.1
[jarm-jwt]: https://openid.net/specs/oauth-v2-jarm.html#section-2.3.4
[oidc-core-auth-error]: https://openid.net/specs/openid-connect-core-1_0.html#AuthError
[rfc6749-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
[rfc6749-code]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2
[rfc6749-authorisation-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2.1
[rfc7636-challenge]: https://www.rfc-editor.org/rfc/rfc7636#section-4.2
[rfc9126-request]: https://www.rfc-editor.org/rfc/rfc9126#section-2.1
[rfc9126-response]: https://www.rfc-editor.org/rfc/rfc9126#section-2.2
[rfc9126-errors]: https://www.rfc-editor.org/rfc/rfc9126#section-2.3
[rfc9126-request-object]: https://www.rfc-editor.org/rfc/rfc9126#section-3
[rfc9126-authorisation]: https://www.rfc-editor.org/rfc/rfc9126#section-4
[rfc9126-guessing]: https://www.rfc-editor.org/rfc/rfc9126#section-7.1
