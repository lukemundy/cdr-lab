# Authorisation server library: Tokens of an arrangement

This area gives the tokens of a CDR arrangement at the data holder authorisation server: the authorisation code grant, the refresh token grant, the token response and the ID token. It also gives the introspection, token revocation and CDR arrangement revocation endpoints. Each of these endpoints authenticates the client as the area [Client authentication](client-authentication.md) gives, and a failed client authentication gives [401 with `invalid_client`](client-authentication.md#auth-0032).

## Authorisation code grant

- <a id="auth-0201"></a>**AUTH-0201** The data holder authorisation server accepts `grant_type=authorization_code` with `code`, `redirect_uri` and `code_verifier`. The `redirect_uri` is equal to that of the request object, and the `code_verifier` matches the `code_challenge`. ([RFC 6749, section 4.1.3][rfc6749-token-request]; [RFC 7636, section 4.6][rfc7636-verify])
- <a id="auth-0202"></a>**AUTH-0202** The response is 400 with `invalid_grant` for each of these conditions. ([CDS: Access Token][cds-access-token]; [RFC 6749, section 4.1.2][rfc6749-code]; [RFC 6749, section 5.2][rfc6749-errors])
  - The code is not known.
  - The code is expired.
  - The code is of a different client.
  - The code was used before.
  - The `redirect_uri` does not match.
  - The `code_verifier` does not match.
- <a id="auth-0203"></a>**AUTH-0203** A second use of a code also revokes the tokens of the first use. ([RFC 6749, section 4.1.2][rfc6749-code])
- <a id="auth-0204"></a>**AUTH-0204** When the server refuses a token request for a code that it knows, the code ends. Each later request with the code gives 400 with `invalid_grant`. ([Judgement: a refused token request ends the code](#auth-0232))

## Token response of an arrangement

- <a id="auth-0205"></a>**AUTH-0205** A response to the authorisation code grant or the refresh token grant also has `id_token` and `cdr_arrangement_id`, in addition to the members of [a successful token response](token-endpoint.md#auth-0046). ([CDS: CDR Arrangement ID][cds-arrangement-id]; [OpenID Connect Core, section 3.1.3.3][oidc-core-token-response])
- <a id="auth-0206"></a>**AUTH-0206** The response has a `refresh_token` only when the sharing duration is more than 0. ([CDS: Requesting Sharing Duration][cds-sharing-duration]: a refresh token "SHOULD be provided" for a one-time collection)
- <a id="auth-0207"></a>**AUTH-0207** The access token has the scopes that the consumer approved. It expires 5 minutes after it is issued, as a client credentials token does. ([CDS: Access Token][cds-access-token]; [Judgement: an access token lifetime of 5 minutes](token-endpoint.md#auth-0059))
- <a id="auth-0208"></a>**AUTH-0208** When the sharing ends less than 5 minutes after the token is issued, the access token expires when the sharing ends. The `expires_in` shows this expiry. This rule applies also to the sharing duration 0, for which a service must give the end of the sharing. ([CDR Rules, rule 4.26(1)(f) and (g)][cdr-rules]; [Judgement: the access token expires at the end of the sharing](#auth-0233))
- <a id="auth-0209"></a>**AUTH-0209** The refresh token expires at the end of the sharing of the arrangement. ([CDS: Refresh Token][cds-refresh-token]: "an exp value equal to the sharing duration authorised by the consumer")

## Refresh token grant

- <a id="auth-0210"></a>**AUTH-0210** The data holder authorisation server accepts `grant_type=refresh_token` with `refresh_token`. The response has a new access token, an ID token and the `cdr_arrangement_id`. ([OpenID Connect Core, section 12.2][oidc-core-refresh-response]; [Judgement: an ID token in a refresh response](#auth-0234))
- <a id="auth-0211"></a>**AUTH-0211** The server does not change the refresh token. When the response has a `refresh_token`, it is the value of the request. ([CDS: Refresh Token][cds-refresh-token]: "MUST NOT cycle refresh tokens")
- <a id="auth-0212"></a>**AUTH-0212** The response is 400 with `invalid_grant` for a refresh token with one of these conditions. ([RFC 6749, section 5.2][rfc6749-errors])
  - It is not known.
  - It is expired.
  - It is revoked.
  - It is of a different client.
- <a id="auth-0213"></a>**AUTH-0213** The server does not compare the client certificate of a refresh request with the certificate of an earlier token request. The new access token is bound to the client certificate of the refresh request. ([RFC 8705, section 7.1][rfc8705-refresh]; [Judgement: a refresh request can use a different client certificate](#auth-0235))

## ID token

- <a id="auth-0214"></a>**AUTH-0214** The server signs the ID token with PS256. It does not encrypt the ID token. ([CDS: Baseline ID Token requirements][cds-id-token]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](configuration.md#auth-0013); [Judgement: no encryption of ID tokens or authorisation responses](configuration.md#auth-0017))
- <a id="auth-0215"></a>**AUTH-0215** The ID token has `iss`, `sub`, `aud`, `exp`, `iat`, `auth_time`, `acr` and `nonce`. The `acr` is `urn:cds.au:cdr:2`. ([OpenID Connect Core, section 2][oidc-core-id-token]; [CDS: Claims][cds-claims])
- <a id="auth-0216"></a>**AUTH-0216** A service must give the subject identifier of each authorisation. The `sub` of the ID token is that value. ([CDS: Sub Claim][cds-sub-claim])
- <a id="auth-0217"></a>**AUTH-0217** The ID token has no claim with personal information, and no `sharing_expires_at` or `refresh_token_expires_at`. ([CDS: Baseline ID Token requirements][cds-id-token]: "MUST NOT contain any Personal Information (PI) claims")

## Introspection

- <a id="auth-0218"></a>**AUTH-0218** The introspection endpoint accepts `POST` with the client assertion parameters and `token`. ([CDS: Introspection endpoint][cds-introspection-endpoint]; [RFC 7662, section 2.1][rfc7662-request])
- <a id="auth-0219"></a>**AUTH-0219** For a refresh token of the client that is not expired and not revoked, the response is 200 with `active` set to `true`, `exp`, `scope` and `cdr_arrangement_id`. The `exp` is the end of the sharing of the arrangement. The response has no `username`. ([CDS: Introspection endpoint][cds-introspection-endpoint]; [CDS: Requesting Sharing Duration][cds-sharing-duration])
- <a id="auth-0220"></a>**AUTH-0220** For each other token, the response is 200 with only `active` set to `false`. For example, an access token, an ID token, a refresh token of a different client and a value that the server does not know. ([CDS: Introspection endpoint][cds-introspection-endpoint]: introspection of access tokens and ID tokens "MUST NOT be supported"; [RFC 7662, section 2.2][rfc7662-response]; [Judgement: active false for a token that is not a refresh token of the client](#auth-0236))
- <a id="auth-0221"></a>**AUTH-0221** A request with no `token` gives 400 with `invalid_request`. ([RFC 7662, section 2.3][rfc7662-errors])

## Token revocation

- <a id="auth-0222"></a>**AUTH-0222** The token revocation endpoint accepts `POST` with the client assertion parameters, `token` and an optional `token_type_hint`. The server finds the token without the hint, and ignores the hint. ([RFC 7009, section 2.1][rfc7009-request])
- <a id="auth-0223"></a>**AUTH-0223** For a refresh token of the client, the server revokes the refresh token and each access token of its grant. The arrangement does not end. ([RFC 7009, section 2.1][rfc7009-request]; [Judgement: no end of the arrangement at a token revocation](#auth-0237))
- <a id="auth-0224"></a>**AUTH-0224** For an access token of the client, the server revokes only that access token. A new access token from the refresh token is valid. This rule applies also to a client credentials token. ([CDS: Token Revocation endpoint][cds-token-revocation]: "Revocation of Refresh Tokens and Access Tokens MUST be supported"; [CDR Support: Guidance on revocation or withdrawal of consent][support-revocation-guidance])
- <a id="auth-0225"></a>**AUTH-0225** For a token of a different client, the response is 400 with `unauthorized_client`, and the token does not change. ([RFC 7009, section 2.1][rfc7009-request]: "the request is refused"; [Judgement: unauthorized_client for a token of a different client](#auth-0238))
- <a id="auth-0226"></a>**AUTH-0226** For each other value, the response is 200 with no body, and nothing changes. For example, a token that is not known, an expired or revoked token, and an ID token. ([RFC 7009, section 2.2][rfc7009-response])
- <a id="auth-0227"></a>**AUTH-0227** A request with no `token` gives 400 with `invalid_request`. ([RFC 7009, section 2.2.1][rfc7009-errors]; [RFC 6749, section 5.2][rfc6749-errors])

## CDR arrangement revocation

- <a id="auth-0228"></a>**AUTH-0228** The CDR arrangement revocation endpoint accepts `POST` with the client assertion parameters and `cdr_arrangement_id`. The server ignores `cdr_arrangement_jwt`. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]: "Data Holders MUST only support CDR Arrangement Form Parameter method")
- <a id="auth-0229"></a>**AUTH-0229** For an arrangement of the client that did not end, a service must end the arrangement and revoke its tokens before the revocation returns. The response is then 204 with no body. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [CDR Support: Error scenarios and responses][support-error-scenarios]: the revocation "must occur synchronously")
- <a id="auth-0230"></a>**AUTH-0230** When the service gives that the arrangement is not known, is of a different client or ended, the response is 422 with `urn:au-cds:error:cds-all:Authorisation/InvalidArrangement`. The `detail` is the `cdr_arrangement_id` of the request. ([CDS: Invalid Consent Arrangement][cds-error-422-invalid-arrangement]; [CDR Support: Error scenarios and responses][support-error-scenarios])
- <a id="auth-0231"></a>**AUTH-0231** A request with no `cdr_arrangement_id`, or with an empty value, gives 400 with `urn:au-cds:error:cds-all:Field/Missing`. The `detail` is `cdr_arrangement_id`. ([CDS: Missing Required Field][cds-error-400-field-missing]; [Judgement: Field/Missing for a revocation with no cdr_arrangement_id](#auth-0239))

## Judgements

### <a id="auth-0232"></a>AUTH-0232 A refused token request ends the code

- Kind: choice
- Basis: Step 3.3. The CDS requires only that the server refuses a used code ([CDS: Access Token][cds-access-token]). RFC 6749 recommends revocation for a second use ([section 4.1.2][rfc6749-code]). RFC 7636 gives only the error for a `code_verifier` that does not match ([section 4.6][rfc7636-verify]). The ACCC leaves the response to the data holder ([CDR Support: Failure to obtain token within Authorisation Code expiry period][support-code-expiry]).
- Reason: No source requires a second attempt, and one attempt is the least complex. A valid code with a `code_verifier` that does not match can show that a different party got the code.

### <a id="auth-0233"></a>AUTH-0233 The access token expires at the end of the sharing

- Kind: choice
- Basis: Step 3.3. An authorisation for a period expires at the end of the period, and an authorisation for a single occasion expires after the disclosure ([CDR Rules, rule 4.26(1)(f) and (g)][cdr-rules]). A sharing duration of 0 gives an access token only ([CDS: Requesting Sharing Duration][cds-sharing-duration]).
- Reason: An expiry that the token and `expires_in` show is the simplest rule for a data recipient and for the server. The service gives the end of the sharing also for the sharing duration 0, so that no access token stays valid after its arrangement ends. For example, a data holder gives the confirmation plus 5 minutes.

### <a id="auth-0234"></a>AUTH-0234 An ID token in a refresh response

- Kind: choice
- Basis: Step 3.3. OpenID Connect permits a refresh response with no ID token ([OpenID Connect Core, section 12.2][oidc-core-refresh-response]).
- Reason: A software product then always gets the same members.

### <a id="auth-0235"></a>AUTH-0235 A refresh request can use a different client certificate

- Kind: choice
- Basis: Step 3.1. The CDS requires only that the client certificate and the access token match on a resource request ([CDS: Holder of Key Mechanism][cds-hok]). FAPI 1.0 Advanced requires only sender-constrained access tokens ([section 5.2.2][fapi-advanced-server], item 5). The client authentication binds the refresh token to the client ([RFC 8705, section 7.1][rfc8705-refresh]), and a client uses its refresh token after it replaces its certificate ([RFC 8705, section 6.3][rfc8705-expiration]).
- Reason: A sharing duration can be one year. An authorisation must not end because a software product replaces its certificate.

### <a id="auth-0236"></a>AUTH-0236 active false for a token that is not a refresh token of the client

- Kind: choice
- Basis: Step 2. The CDS forbids introspection of access tokens and ID tokens, and gives no response for it ([CDS: Introspection endpoint][cds-introspection-endpoint]). The DSB says that RFC 7662 applies ([CDR Support: Introspection Endpoint RFC vs Standards clarification][support-introspection]), and RFC 7662 gives this response for a token that the client must not see ([section 2.2][rfc7662-response]).
- Reason: The response follows the DSB clarification. It does not tell a client which tokens exist.

### <a id="auth-0237"></a>AUTH-0237 No end of the arrangement at a token revocation

- Kind: choice
- Basis: Step 2. The token revocation endpoint is for "token clean up" ([CDS: Token Revocation endpoint][cds-token-revocation]). The CDR arrangement revocation endpoint withdraws an authorisation ([CDR Support: Support for token revocation endpoints][support-token-revocation]).
- Reason: The arrangement then has no token, so the data holder discloses no data. A software product that wants to end the arrangement calls the CDR arrangement revocation endpoint.

### <a id="auth-0238"></a>AUTH-0238 unauthorized_client for a token of a different client

- Kind: choice
- Basis: Step 3.2. RFC 7009 says that "the request is refused" and gives no error code ([section 2.1][rfc7009-request]). The error format is that of RFC 6749 ([RFC 7009, section 2.2.1][rfc7009-errors]).
- Reason: `unauthorized_client` is the RFC 6749 code for a client that is not permitted to do the request ([RFC 6749, section 5.2][rfc6749-errors]).

### <a id="auth-0239"></a>AUTH-0239 Field/Missing for a revocation with no cdr_arrangement_id

- Kind: choice
- Basis: Step 3.3. The CDS gives 422 for an identifier that is not valid or not found, and no response for a request with no identifier ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]). The CDS gives `Field/Missing` for a required field that is absent ([CDS: Missing Required Field][cds-error-400-field-missing]).
- Reason: The rule is the same as for a required field on the other CDS APIs of the data holder ([a required field that is absent](../../../services/dataholder/design/api-standards.md#dh-0450)).

[cds-access-token]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#access-token
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[cds-arrangement-id]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-id
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
[cds-refresh-token]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#refresh-token
[cds-id-token]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#baseline-id-token-requirements
[cds-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#claims
[cds-sub-claim]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#sub-claim
[cds-introspection-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#introspection-endpoint
[cds-token-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#token-revocation-endpoint
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-error-422-invalid-arrangement]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-422-authorisation-invalid-arrangement
[cds-error-400-field-missing]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-missing
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[support-code-expiry]: https://cdr-support.zendesk.com/hc/en-us/articles/360004665496
[support-introspection]: https://cdr-support.zendesk.com/hc/en-us/articles/900004618103
[support-revocation-guidance]: https://cdr-support.zendesk.com/hc/en-us/articles/6828747049359
[support-token-revocation]: https://cdr-support.zendesk.com/hc/en-us/articles/900003387786
[fapi-advanced-server]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#authorization-server
[oidc-core-id-token]: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
[oidc-core-token-response]: https://openid.net/specs/openid-connect-core-1_0.html#TokenResponse
[oidc-core-refresh-response]: https://openid.net/specs/openid-connect-core-1_0.html#RefreshTokenResponse
[rfc6749-code]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2
[rfc6749-token-request]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.3
[rfc6749-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
[rfc7636-verify]: https://www.rfc-editor.org/rfc/rfc7636#section-4.6
[rfc7662-request]: https://www.rfc-editor.org/rfc/rfc7662#section-2.1
[rfc7662-response]: https://www.rfc-editor.org/rfc/rfc7662#section-2.2
[rfc7662-errors]: https://www.rfc-editor.org/rfc/rfc7662#section-2.3
[rfc7009-request]: https://www.rfc-editor.org/rfc/rfc7009#section-2.1
[rfc7009-response]: https://www.rfc-editor.org/rfc/rfc7009#section-2.2
[rfc7009-errors]: https://www.rfc-editor.org/rfc/rfc7009#section-2.2.1
[rfc8705-refresh]: https://www.rfc-editor.org/rfc/rfc8705#section-7.1
[rfc8705-expiration]: https://www.rfc-editor.org/rfc/rfc8705#section-6.3
