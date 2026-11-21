# Authorisation server library: Client authentication

This area gives the client authentication of each endpoint that authenticates a client. These endpoints are the token endpoint of each authorisation server, and the PAR, introspection, token revocation and CDR arrangement revocation endpoints of the data holder authorisation server.

## Client assertion

- <a id="auth-0020"></a>**AUTH-0020** A client authenticates with `private_key_jwt`: the request has `client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer` and a `client_assertion`. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt]; [CDS: Data Recipients calling the CDR Register][cds-dr-calling-register])
- <a id="auth-0021"></a>**AUTH-0021** The authorisation server supports no other client authentication method. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0022"></a>**AUTH-0022** The `iss` and the `sub` of the assertion are the client ID. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0023"></a>**AUTH-0023** The `client_id` parameter is optional. When it is present, it is equal to the `iss` of the assertion. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0024"></a>**AUTH-0024** A service must give the client ID and the JWKS URI of each client. ([CDS: Client Authentication][cds-client-auth])
- <a id="auth-0025"></a>**AUTH-0025** The authorisation server validates the assertion signature with the keys at the JWKS URI of the client. ([CDS: Client Authentication][cds-client-auth]; [CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0026"></a>**AUTH-0026** The authorisation server accepts an assertion that is signed with PS256 or ES256. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](configuration.md#auth-0013))
- <a id="auth-0240"></a>**AUTH-0240** When a service gives the registered `token_endpoint_auth_signing_alg` of a client, the authorisation server refuses an assertion of that client with a different algorithm. For example, a client that registered PS256 cannot authenticate with ES256. ([OpenID Connect Dynamic Client Registration 1.0, section 2][oidc-dcr-metadata])
- <a id="auth-0027"></a>**AUTH-0027** The authorisation server accepts one of these values as the `aud` of the assertion. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
  - the issuer
  - the URL of the token endpoint
  - the URL of the endpoint that the client calls
- <a id="auth-0028"></a>**AUTH-0028** The authorisation server refuses an assertion with no `jti` or no `exp`. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0029"></a>**AUTH-0029** The authorisation server refuses an assertion that is expired, after the [clock leeway](#auth-0251). ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0030"></a>**AUTH-0030** The authorisation server refuses an assertion whose `jti` it accepted before. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0250"></a>**AUTH-0250** A service must keep each `jti` that the authorisation server accepted for longer than the JWT can be valid, plus the [clock leeway](#auth-0251). This is more than 10 minutes and 30 seconds for a client assertion, and more than 60 minutes and 30 seconds for a [request object](authorisation-request.md#auth-0253). For example, the Register keeps each `jti` for 1 hour, and the data holder for 2 hours. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt]; [Judgement: a maximum client assertion lifetime of 10 minutes](#auth-0033); [Judgement: a clock leeway of 30 seconds](#auth-0252))
- <a id="auth-0031"></a>**AUTH-0031** The authorisation server refuses an assertion whose `exp` is more than 10 minutes after the time of the request. ([Judgement: a maximum client assertion lifetime of 10 minutes](#auth-0033))

## Clock leeway

- <a id="auth-0251"></a>**AUTH-0251** The authorisation server accepts a clock difference of 30 seconds or less in the time claims of a client assertion and of a request object. It accepts an `iat` or an `nbf` that is 30 seconds or less after the time of the request, and an `exp` that is 30 seconds or less before it. For example, it accepts an assertion with an `iat` 20 seconds in the future, and it refuses one with an `iat` 60 seconds in the future. ([RFC 7519, section 4.1.4][rfc7519-exp]; [Judgement: a clock leeway of 30 seconds](#auth-0252))

## JWKS of a client

- <a id="auth-0242"></a>**AUTH-0242** When the authorisation server gets the JWKS of a client, it does not follow a redirect. For example, a JWKS URI that responds with 302 gives no keys. ([Judgement: limits on the request for the JWKS of a client](#auth-0244))
- <a id="auth-0243"></a>**AUTH-0243** Performance: the authorisation server stops a request for the JWKS of a client after 5 seconds. ([Judgement: limits on the request for the JWKS of a client](#auth-0244))

## Refusal

- <a id="auth-0032"></a>**AUTH-0032** When the client authentication fails, the response is 401 with `invalid_client`, also when the client does not exist. ([RFC 6749, section 5.2][rfc6749-errors]; [Judgement: 401 for invalid_client, not the 400 of the ACCC guidance](#auth-0241))
- <a id="auth-0248"></a>**AUTH-0248** Telemetry: the log of a refused request has only the error code and the error description. For example, it has no client ID and no claim of the assertion. ([no CDR data or personal information in telemetry](../../platform/telemetry/DESIGN.md#tele-0004))

## Judgements

### <a id="auth-0033"></a>AUTH-0033 A maximum client assertion lifetime of 10 minutes

- Kind: choice
- Basis: Step 3.3. The CDS requires that an assertion is used one time, and gives no lifetime ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt]).
- Reason: A short limit keeps the window for a stolen assertion small. The data holder uses the same limit for a registration request JWT.

### <a id="auth-0252"></a>AUTH-0252 A clock leeway of 30 seconds

- Kind: choice
- Basis: Step 3.2. The CDS and FAPI 1.0 Advanced give no leeway for clock differences. RFC 7519 permits "some small leeway, usually no more than a few minutes, to account for clock skew" ([RFC 7519, section 4.1.4][rfc7519-exp]). As an industry benchmark, the FAPI 2.0 Security Profile requires a server to accept an `iat` or an `nbf` 10 seconds or less in the future, and to refuse one more than 60 seconds in the future. It says: "Some ecosystems have found that the value of 30 seconds is needed to fully eliminate clock skew issues" ([FAPI 2.0 Security Profile, section 5.3.2.1][fapi2-clock]).
- Reason: The lab exists to test real clients, and their clocks are not exact. With no leeway, a correct client whose clock is 1 second ahead fails. 30 seconds is in the range of the benchmark. One value for `iat`, `nbf` and `exp` is the least complex. The value is small against the lifetime of an assertion or a request object, so it does not make the window for a stolen JWT much larger. The negative time tests of the OpenID FAPI 1.0 Advanced test plan are 5 minutes or more from the time of the server, so the leeway does not change their results.

### <a id="auth-0241"></a>AUTH-0241 401 for invalid_client, not the 400 of the ACCC guidance

- Kind: conflict
- Basis: Step 2. ACCC guidance says: "When client authentication occurs against the token endpoint and the sub claim on the client assertion is missing or does not match the client, then the most appropriate response is to return 400: invalid_client in the body." ([CDR Support: Error scenarios and responses][support-error-scenarios]). RFC 6749 permits 401: "The authorization server MAY return an HTTP 401 (Unauthorized) status code to indicate which HTTP authentication schemes are supported" ([RFC 6749, section 5.2][rfc6749-errors]).
- Reason: The ACCC text says "most appropriate", so it is not a requirement. The authorisation server library gives 401 for each `invalid_client`, and one status code for each failure of the client authentication is the least complex. A client that follows RFC 6749 accepts the two status codes.

### <a id="auth-0244"></a>AUTH-0244 Limits on the request for the JWKS of a client

- Kind: choice
- Basis: Step 3.2. A server that gets a URL from a client must not follow a redirect ([OWASP: Server-Side Request Forgery Prevention][owasp-ssrf]).
- Reason: A redirect can send the request to a host that the registration does not name. A JWKS URI that does not reply must not hold a request of a client with no end.

[cds-private-key-jwt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#private-key-jwt-client-authentication
[cds-dr-calling-register]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-the-cdr-register
[cds-client-auth]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#client-authentication
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[rfc6749-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[oidc-dcr-metadata]: https://openid.net/specs/openid-connect-registration-1_0.html#ClientMetadata
[rfc7519-exp]: https://www.rfc-editor.org/rfc/rfc7519#section-4.1.4
[fapi2-clock]: https://openid.net/specs/fapi-security-profile-2_0-final.html#section-5.3.2.1
[owasp-ssrf]: https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html
