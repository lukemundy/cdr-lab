# Authorisation server library: Token endpoint

This area gives the token endpoint of each authorisation server: the grants, the client credentials token, the token response and the errors. It also gives the verification of an access token for a service. The area [Tokens of an arrangement](tokens-of-an-arrangement.md) gives the rules of the authorisation code grant and the refresh token grant.

## Grants

- <a id="auth-0034"></a>**AUTH-0034** The token endpoint of the Register authorisation server accepts only `grant_type=client_credentials`. ([CDS: Data Recipients calling the CDR Register][cds-dr-calling-register])
- <a id="auth-0035"></a>**AUTH-0035** The token endpoint of the data holder authorisation server accepts the grant types `client_credentials`, `authorization_code` and `refresh_token`. ([CDS: Data Recipients calling Data Holders][cds-dr-calling-dh])
- <a id="auth-0036"></a>**AUTH-0036** When the grant type of a request is not one that the endpoint accepts, the response is 400 with `unsupported_grant_type`. ([RFC 6749, section 5.2][rfc6749-errors])
- <a id="auth-0037"></a>**AUTH-0037** When the registration of a client does not have the grant type of the request in its `grant_types`, the data holder authorisation server responds with 400 and `unauthorized_client`. ([RFC 6749, section 5.2][rfc6749-errors])

## Client credentials token

- <a id="auth-0038"></a>**AUTH-0038** A client credentials token of the data holder authorisation server has only the scope `cdr:registration`. ([CDS: Authorisation Scopes][cds-authorisation-scopes]; [Judgement: only the cdr:registration scope in a client credentials token](#auth-0060))
- <a id="auth-0039"></a>**AUTH-0039** When a client credentials request to the data holder authorisation server has no `scope`, the token gets the scope `cdr:registration`. ([RFC 6749, section 3.3][rfc6749-scope]; [Judgement: only the cdr:registration scope in a client credentials token](#auth-0060))
- <a id="auth-0040"></a>**AUTH-0040** When a client credentials request to the data holder authorisation server asks for a different scope, the response is 400 with `invalid_scope`. ([CDS: Authorisation Scopes][cds-authorisation-scopes]; [RFC 6749, section 5.2][rfc6749-errors])
- <a id="auth-0041"></a>**AUTH-0041** When the registration of a client does not have the scope `cdr:registration`, a client credentials request of the client to the data holder authorisation server gets 400 with `invalid_scope`. ([RFC 6749, section 5.2][rfc6749-errors])
- <a id="auth-0042"></a>**AUTH-0042** A client credentials token has no refresh token. ([RFC 6749, section 4.4.3][rfc6749-client-credentials])
- <a id="auth-0043"></a>**AUTH-0043** A client credentials token expires 5 minutes after it is issued. ([CDS: Access Token][cds-access-token]; [ACCC Developer Portal: Token Endpoint][accc-token]; [Judgement: an access token lifetime of 5 minutes](#auth-0059))

## Access token

- <a id="auth-0044"></a>**AUTH-0044** An access token is a JWT that the authorisation server signs with PS256. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](configuration.md#auth-0013))
- <a id="auth-0045"></a>**AUTH-0045** An access token is bound to the client certificate of the token request, as RFC 8705 specifies: `cnf.x5t#S256` is the base64url SHA-256 thumbprint of the certificate. ([CDS: Holder of Key Mechanism][cds-hok]; [RFC 8705, section 3][rfc8705-binding])

## Token response

- <a id="auth-0046"></a>**AUTH-0046** A successful response has `access_token`, `token_type` with the value `Bearer`, `expires_in` and `scope`. ([RFC 6749, section 5.1][rfc6749-token-response])
- <a id="auth-0047"></a>**AUTH-0047** Each response of the token endpoint has `Cache-Control: no-store`, also an error response. ([RFC 6749, section 5.1][rfc6749-token-response]; [RFC 6749, section 5.2][rfc6749-errors])

## Errors

- <a id="auth-0048"></a>**AUTH-0048** An error response of the token endpoint has the format of RFC 6749, section 5.2, with only the error codes of that section. ([RFC 6749, section 5.2][rfc6749-errors]; [Judgement: only the RFC 6749 error codes at the token endpoint](#auth-0057))
- <a id="auth-0049"></a>**AUTH-0049** When a server failure occurs, the token endpoint responds with 500 and `server_error`. ([Judgement: only the RFC 6749 error codes at the token endpoint](#auth-0057))
- <a id="auth-0050"></a>**AUTH-0050** An `invalid_client` error has the status 401. Each other error code has the status 400. ([RFC 6749, section 5.2][rfc6749-errors])

## Client checks

- <a id="auth-0051"></a>**AUTH-0051** The Register authorisation server issues a token only when the client certificate is one of the certificates that the service registered for the client. It compares the full certificate, not only the common name. ([ACCC Developer Portal: Token Endpoint][accc-token])
- <a id="auth-0052"></a>**AUTH-0052** The Register authorisation server issues a token only when the service gives the client as ACTIVE. ([ACCC Developer Portal: Token Endpoint][accc-token])
- <a id="auth-0053"></a>**AUTH-0053** When the client certificate is not registered or the client is not ACTIVE, the Register authorisation server responds with 401 and `invalid_client`. ([RFC 6749, section 5.2][rfc6749-errors]; [Judgement: one refusal for a client that the Register cannot trust](#auth-0058))
- <a id="auth-0054"></a>**AUTH-0054** The data holder authorisation server does not compare the client certificate with the Register data of the software product. ([CDS: Holder of Key Mechanism][cds-hok])

## Access token verification

- <a id="auth-0055"></a>**AUTH-0055** When a service verifies an access token, the token is valid only with the client certificate that it is bound to. ([CDS: Holder of Key Mechanism][cds-hok])
- <a id="auth-0056"></a>**AUTH-0056** An access token that a key of the JWKS signed stays valid until it expires, also when the authorisation server signs with a different key. ([Judgement: signing with the first key of the list](configuration.md#auth-0019))

## Judgements

### <a id="auth-0057"></a>AUTH-0057 Only the RFC 6749 error codes at the token endpoint

- Kind: choice
- Basis: Step 3.2. The CDS does not define the token errors of the Register, and RFC 6749 is the base protocol ([RFC 6749, section 5.2][rfc6749-errors]).
- Reason: A client library that follows the RFC knows only these codes. RFC 6749 gives `server_error` only for the authorisation endpoint, but it is the usual code.

### <a id="auth-0058"></a>AUTH-0058 One refusal for a client that the Register cannot trust

- Kind: choice
- Basis: Step 3.2. RFC 6749 has no code that is more specific than `invalid_client` ([RFC 6749, section 5.2][rfc6749-errors]).
- Reason: One response for each cause does not tell a caller which software products exist or which are ACTIVE. The log of the Register gives the cause to the operator.

### <a id="auth-0059"></a>AUTH-0059 An access token lifetime of 5 minutes

- Kind: choice
- Basis: Step 2. The ACCC gives 5 minutes for the Register ([ACCC Developer Portal: Token Endpoint][accc-token]). The CDS permits 2 to 10 minutes for a data holder ([CDS: Access Token][cds-access-token]). The CDS gives no lifetime for the Register, and its example of 7200 seconds is not normative.
- Reason: One lifetime for the two servers is the least complex. 5 minutes is in the range of the CDS.

### <a id="auth-0060"></a>AUTH-0060 Only the cdr:registration scope in a client credentials token

- Kind: choice
- Basis: Step 1. The CDS forbids `cdr:registration` and a CDR data scope in one token ([CDS: Authorisation Scopes][cds-authorisation-scopes]). RFC 6749 permits a default scope ([section 3.3][rfc6749-scope]).
- Reason: The client credentials grant has no consumer, so a data scope has no meaning for it.

[cds-dr-calling-register]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-the-cdr-register
[cds-dr-calling-dh]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-data-holders
[cds-authorisation-scopes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-scopes
[cds-access-token]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#access-token
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[rfc6749-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
[rfc6749-scope]: https://www.rfc-editor.org/rfc/rfc6749#section-3.3
[rfc6749-client-credentials]: https://www.rfc-editor.org/rfc/rfc6749#section-4.4.3
[rfc6749-token-response]: https://www.rfc-editor.org/rfc/rfc6749#section-5.1
[rfc8705-binding]: https://www.rfc-editor.org/rfc/rfc8705#section-3
