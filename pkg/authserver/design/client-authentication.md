# Authorisation server library: Client authentication

This area gives the client authentication of each endpoint that authenticates a client. These endpoints are the token endpoint of each authorisation server, and the PAR, introspection, token revocation and CDR arrangement revocation endpoints of the data holder authorisation server.

## Client assertion

- <a id="auth-0020"></a>**AUTH-0020** A client authenticates with `private_key_jwt`: the request has `client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer` and a `client_assertion`. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt]; [CDS: Data Recipients calling the CDR Register][cds-dr-calling-register])
- <a id="auth-0021"></a>**AUTH-0021** The authorisation server supports no other client authentication method. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0022"></a>**AUTH-0022** The `iss` and the `sub` of the assertion are the client ID. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0023"></a>**AUTH-0023** The `client_id` parameter is optional. When it is present, it is equal to the `iss` of the assertion. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0024"></a>**AUTH-0024** A service must give the client ID and the JWKS URI of each client. ([CDS: Client Authentication][cds-client-auth])
- <a id="auth-0025"></a>**AUTH-0025** The authorisation server validates the assertion signature with the keys at the JWKS URI of the client. ([CDS: Client Authentication][cds-client-auth]; [CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0026"></a>**AUTH-0026** The authorisation server accepts an assertion that is signed with PS256 or ES256. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms]; [Judgement: the signing algorithms](configuration.md#auth-0013))
- <a id="auth-0027"></a>**AUTH-0027** The authorisation server accepts one of these values as the `aud` of the assertion. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
  - the issuer
  - the URL of the token endpoint
  - the URL of the endpoint that the client calls
- <a id="auth-0028"></a>**AUTH-0028** The authorisation server refuses an assertion with no `jti` or no `exp`. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0029"></a>**AUTH-0029** The authorisation server refuses an assertion that is expired. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0030"></a>**AUTH-0030** The authorisation server refuses an assertion whose `jti` it accepted before. ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt])
- <a id="auth-0031"></a>**AUTH-0031** The authorisation server refuses an assertion whose `exp` is more than 10 minutes after the time of the request. ([Judgement: the maximum lifetime of a client assertion](#auth-0033))

## Refusal

- <a id="auth-0032"></a>**AUTH-0032** When the client authentication fails, the response is 401 with `invalid_client`, also when the client does not exist. ([RFC 6749, section 5.2][rfc6749-errors])

## Judgements

### <a id="auth-0033"></a>AUTH-0033 The maximum lifetime of a client assertion

- Kind: choice
- Basis: Step 3.3. The CDS requires that an assertion is used one time, and gives no lifetime ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt]).
- Reason: A short limit keeps the window for a stolen assertion small. The data holder uses the same limit for a registration request JWT.

[cds-private-key-jwt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#private-key-jwt-client-authentication
[cds-dr-calling-register]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-the-cdr-register
[cds-client-auth]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#client-authentication
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[rfc6749-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
