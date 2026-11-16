# Mock CDR Register: Authorisation server

This area gives the choices of the Register for its authorisation server: the OpenID Provider configuration, the token endpoint, the access tokens and the signing keys. The Register serves the Register authorisation server of the [authorisation server library](../../../pkg/authserver/design/README.md). This area gives only the choices and the additions of the Register.

## Endpoints

| ID | Endpoint | Method and path | Transport | Authentication | Source |
|---|---|---|---|---|---|
| <a id="reg-0450"></a>**REG-0450** | Get OpenId Provider Config | `GET /idp/.well-known/openid-configuration` | TLS | None | [CDS: Get OpenId Provider Config][cds-get-oidc] |
| <a id="reg-0451"></a>**REG-0451** | Access token signing keys | The `jwks_uri` of the OpenID Provider configuration | TLS | None | [CDS: Get OpenId Provider Config][cds-get-oidc] |
| <a id="reg-0452"></a>**REG-0452** | Token endpoint | `POST /idp/connect/token` | mTLS | `private_key_jwt` | [CDS: Register Endpoints][cds-register-endpoints] |

## Configuration and tokens

The library gives [the fields and values of the Register configuration](../../../pkg/authserver/design/configuration.md#auth-0003), [only the client credentials grant](../../../pkg/authserver/design/token-endpoint.md#auth-0034) with [`private_key_jwt` client authentication](../../../pkg/authserver/design/client-authentication.md#auth-0020), and an access token that is [bound to the client certificate](../../../pkg/authserver/design/token-endpoint.md#auth-0045), [expires 5 minutes after it is issued](../../../pkg/authserver/design/token-endpoint.md#auth-0043) and has [no refresh token](../../../pkg/authserver/design/token-endpoint.md#auth-0042).

- <a id="reg-0453"></a>**REG-0453** The client ID of a software product is its software product ID. The Register refuses a client assertion with a data recipient brand ID as its client ID. ([CDS: Data Recipients calling the CDR Register][cds-dr-calling-register]; [Judgement: the software product ID as the client ID](#reg-0480))
- <a id="reg-0454"></a>**REG-0454** The Register gives the JWKS URI of the software product to [validate the client assertion signature](../../../pkg/authserver/design/client-authentication.md#auth-0025). ([CDS: Data Recipients calling the CDR Register][cds-dr-calling-register])
- <a id="reg-0455"></a>**REG-0455** The Register keeps the `jti` of each client assertion that it accepted for 1 hour, across all software products. In that time, it [refuses the same `jti` again](../../../pkg/authserver/design/client-authentication.md#auth-0030). ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt]; [Judgement: the time that the Register keeps a `jti`](#reg-0481))
- <a id="reg-0456"></a>**REG-0456** The Register gives [the registered client certificates](participants.md#reg-0213) of the software product as the [certificates of the client](../../../pkg/authserver/design/token-endpoint.md#auth-0051). ([ACCC Developer Portal: Token Endpoint][accc-token])
- <a id="reg-0457"></a>**REG-0457** The Register gives the client as ACTIVE only when [the software product is ACTIVE](../../../pkg/authserver/design/token-endpoint.md#auth-0052). ([ACCC Developer Portal: Token Endpoint][accc-token])

## Errors

The library gives [the format and the codes of RFC 6749, section 5.2](../../../pkg/authserver/design/token-endpoint.md#auth-0048) for the token endpoint errors. A client that the Register cannot authenticate gets [401 `invalid_client`](../../../pkg/authserver/design/client-authentication.md#auth-0032), as does [a client that the Register cannot trust](../../../pkg/authserver/design/token-endpoint.md#auth-0053).

## Signing keys

- <a id="reg-0465"></a>**REG-0465** The Register has two key sets. One key set signs access tokens, and the `jwks_uri` of the OpenID Provider configuration gives it. A different key set signs SSAs, and [Get JWKS](software-statement-assertion.md#reg-0401) gives it. ([CDS: Get OpenId Provider Config][cds-get-oidc]; [CDS: Get JWKS][cds-get-jwks])
- <a id="reg-0466"></a>**REG-0466** Each key of the two key sets is an [RSA key of 2048 bits or longer](../../../pkg/authserver/design/configuration.md#auth-0012) for PS256. Each published JWK has [the members that the CDS requires](../../../pkg/authserver/design/configuration.md#auth-0007). ([CDS: Get JWKS][cds-get-jwks])
- <a id="reg-0467"></a>**REG-0467** The lab operator can change the key of each key set through the [operator interface](operator-interface.md#reg-0846) or the [operator console](operator-console-actions.md#reg-1210). During a key change, the JWKS gives the old and the new key. ([RFC 7517: JWK Set][rfc7517-jwks]; [Judgement: the steps of a key change](#reg-0482))
- <a id="reg-0468"></a>**REG-0468** A key change has two steps. In the first step, the JWKS gives the next key, and the Register still signs with the old key. In the second step, the Register signs with the next key, and the JWKS still gives the old key as the previous key. ([Judgement: the steps of a key change](#reg-0482))
- <a id="reg-0469"></a>**REG-0469** The previous key leaves the JWKS when the lab operator does the first step of a later key change of the same key set. ([Judgement: when the previous key leaves the JWKS](#reg-0483))
- <a id="reg-0470"></a>**REG-0470** An SSA that the old key signed stays valid until it expires, as does [an access token](../../../pkg/authserver/design/token-endpoint.md#auth-0056). ([CDS: SSA][cds-ssa])
- <a id="reg-0471"></a>**REG-0471** A key change of one key set does not change the other key set. ([CDS: Get OpenId Provider Config][cds-get-oidc])

## Judgements

### <a id="reg-0480"></a>REG-0480 The software product ID as the client ID

- Kind: choice
- Basis: Step 3.3. The CDS permits the software product ID or the brand ID as the client ID. It says that the brand ID is deprecated ([CDS: Data Recipients calling the CDR Register][cds-dr-calling-register]).
- Reason: One client ID for each software product is the least complex. The CDS gives the software product ID as the identifier that stays.

### <a id="reg-0481"></a>REG-0481 The time that the Register keeps a `jti`

- Kind: choice
- Basis: Step 3.3. The CDS says only that the Register refuses a `jti` that it accepted before ([CDS: Private Key JWT Client Authentication][cds-private-key-jwt]).
- Reason: A client assertion is valid for [10 minutes at most](../../../pkg/authserver/design/client-authentication.md#auth-0031). Thus a longer memory adds nothing.

### <a id="reg-0482"></a>REG-0482 The steps of a key change

- Kind: choice
- Basis: Step 3.3. Participants cache the JWKS, so the Register publishes a new key before it signs with it. A JWKS can hold more than one key ([RFC 7517: JWK Set][rfc7517-jwks]).
- Reason: Only the lab operator knows when each party has the new JWKS. Two operator actions need no timer. A lab user can keep the state "next key published, old key signs" to test a JWKS cache.

### <a id="reg-0483"></a>REG-0483 When the previous key leaves the JWKS

- Kind: choice
- Basis: Step 3.3. An SSA is valid for 10 minutes ([CDS: SSA][cds-ssa]), and an access token for 5 minutes.
- Reason: A later key change is after the last use of the previous key in practice, and this needs no timer. When two key changes complete in less than 10 minutes, an SSA of the first key can fail its verification.

[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[cds-dr-calling-register]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-the-cdr-register
[cds-get-jwks]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-jwks
[cds-get-oidc]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-openid-provider-config
[cds-private-key-jwt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#private-key-jwt-client-authentication
[cds-register-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-endpoints
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[rfc7517-jwks]: https://www.rfc-editor.org/rfc/rfc7517#section-5
