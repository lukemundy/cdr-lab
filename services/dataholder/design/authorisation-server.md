# Mock data holder: Authorisation server

This area gives the choices of the data holder for its authorisation server: the OpenID Provider configuration, the JWKS, the signing keys, client authentication, the authorisation request and the subject identifier. The data holder serves the data holder authorisation server of the [authorisation server library](../../../pkg/authserver/design/README.md). The library gives the rules of the [token endpoint](../../../pkg/authserver/design/token-endpoint.md), the [authorisation request](../../../pkg/authserver/design/authorisation-request.md), the [tokens of an arrangement](../../../pkg/authserver/design/tokens-of-an-arrangement.md), introspection and token revocation. [API standards](api-standards.md) gives the endpoints.

## OpenID Provider configuration and JWKS

The library gives [the path](../../../pkg/authserver/design/configuration.md#auth-0001) and [the fields](../../../pkg/authserver/design/configuration.md#auth-0004) of the OpenID Provider configuration, and [the keys of the JWKS](../../../pkg/authserver/design/configuration.md#auth-0006).

- <a id="dh-0500"></a>**DH-0500** The issuer is the public base URI of the data holder. It is the `infosecBaseUri` of the brand on the Register. For example, the issuer of the default preset brand in the lab is `https://wattle-bank:8081`. ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]; [Judgement: the issuer on the public listener](#dh-0530))
- <a id="dh-0501"></a>**DH-0501** The issuer is an `https` URL with no query and no fragment. ([OpenID Connect Discovery, section 4.3][oidc-discovery-validation])
- <a id="dh-0502"></a>**DH-0502** The supported scopes are these. ([CDS: Scopes and Claims][cds-scopes-and-claims]; [CDS: Authorisation Scopes][cds-authorisation-scopes]; [Judgement: only the scopes of the APIs in scope](#dh-0531))
  - `openid`
  - `profile`
  - `cdr:registration`
  - `common:customer.basic:read`
  - `bank:accounts.basic:read`
  - `bank:accounts.detail:read`
  - `bank:transactions:read`
- <a id="dh-0503"></a>**DH-0503** The `jwksEndpoint` of the brand on the Register is the `jwks_uri` of the OpenID Provider configuration. ([CDS: Multi-Brand Support][cds-multi-brand]; [Judgement: one JWKS for the Register and the configuration](#dh-0532))

## Signing keys

The library gives [the signing algorithm](../../../pkg/authserver/design/configuration.md#auth-0009) and [the size of a signing key](../../../pkg/authserver/design/configuration.md#auth-0012).

- <a id="dh-0504"></a>**DH-0504** When the data holder starts with no signing key, it creates one signing key. A restart [keeps the signing keys](non-functional-requirements.md#dh-0003). ([a different key ID for each signing key](../../../pkg/authserver/design/configuration.md#auth-0008); [Judgement: a signing key that the data holder creates](#dh-0533))

## Client authentication

The library gives [the client authentication with `private_key_jwt`](../../../pkg/authserver/design/client-authentication.md#auth-0020) and the refusals of an assertion, for example [a `jti` that it accepted before](../../../pkg/authserver/design/client-authentication.md#auth-0030).

- <a id="dh-0505"></a>**DH-0505** The client ID of a software product is the client ID that the data holder issued at its registration. ([the client ID of each client](../../../pkg/authserver/design/client-authentication.md#auth-0024); [the registration response](../DESIGN.md#registration-response))
- <a id="dh-0506"></a>**DH-0506** The JWKS URI of a client is the `jwks_uri` of its registration. For example, the data holder validates [the assertion signature](../../../pkg/authserver/design/client-authentication.md#auth-0025) with the keys at that URI. ([the JWKS URI of each client](../../../pkg/authserver/design/client-authentication.md#auth-0024))

## Authorisation request

The library gives the pushed authorisation request, the request object and the authorisation endpoint.

- <a id="dh-0507"></a>**DH-0507** The authorisation request of a client uses these values of its registration. ([the grant types of the registration](../../../pkg/authserver/design/authorisation-request.md#auth-0064); [the signing algorithm of the request object](../../../pkg/authserver/design/authorisation-request.md#auth-0072); [the redirect URI of the request](../../../pkg/authserver/design/authorisation-request.md#auth-0083); [the scope of the registration](../../../pkg/authserver/design/authorisation-request.md#auth-0085))
  - `grant_types`
  - `scope`
  - `redirect_uris`
  - `request_object_signing_alg`
  - `jwks_uri`
- <a id="dh-0508"></a>**DH-0508** When the data holder refuses a request to the authorisation endpoint, [its error page](../../../pkg/authserver/design/authorisation-request.md#auth-0101) does not say which rule failed. ([RFC 6749, section 4.1.2.1][rfc6749-authorisation-errors]; [Judgement: no cause on the error page of the authorisation endpoint](#dh-0534))

## Subject identifier

The library gives the tokens of an arrangement. The data holder gives [the subject identifier of each authorisation](../../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0216), which is the `sub` of the ID token.

- <a id="dh-0509"></a>**DH-0509** The `sub` is a pairwise identifier. For each pair of customer and sector identifier, the data holder creates a random RFC 4122 version 4 UUID and keeps it. ([CDS: Sub Claim][cds-sub-claim]; [DSB: standards-maintenance issue 480][dsb-issue-480]; [OpenID Connect Core, section 8.1][oidc-core-pairwise])
- <a id="dh-0510"></a>**DH-0510** The sector identifier is the host of the `sector_identifier_uri` of the registration. ([OpenID Connect Core, section 8.1][oidc-core-pairwise])
- <a id="dh-0511"></a>**DH-0511** When the registration has no `sector_identifier_uri`, the sector identifier is the host of the `redirect_uris`. ([OpenID Connect Core, section 8.1][oidc-core-pairwise]; [Judgement: the sector identifier of a registration with no sector_identifier_uri](#dh-0535))
- <a id="dh-0512"></a>**DH-0512** When the registration has no `sector_identifier_uri` and its `redirect_uris` have more than one host, the sector identifier is the software product ID. ([Judgement: the sector identifier of a registration with no sector_identifier_uri](#dh-0535))
- <a id="dh-0513"></a>**DH-0513** The `sub` of a customer is the same in each arrangement with the same sector identifier, also after the software product registers again. ([CDR Support: ID Permanence and PPID][support-ppid])

## Judgements

### <a id="dh-0530"></a>DH-0530 The issuer on the public listener

- Kind: conflict
- Basis: Step 1. OpenID Connect Discovery requires that the `issuer` is the URL prefix of the configuration ([section 4.3][oidc-discovery-validation]), and the configuration endpoint is on the public listener ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]). The CDS example has an issuer on the mTLS host, which does not obey that rule.
- Reason: The example is not normative. The issuer is then on the public listener, and the mTLS endpoints are on a second listener.

### <a id="dh-0531"></a>DH-0531 Only the scopes of the APIs in scope

- Kind: choice
- Basis: Step 3.3. The CDS requires `openid` and `profile` ([CDS: Scopes and Claims][cds-scopes-and-claims]).
- Reason: A scope with no API gives an authorisation that discloses nothing. The data holder adds a scope with the first API that needs it.

### <a id="dh-0532"></a>DH-0532 One JWKS for the Register and the configuration

- Kind: choice
- Basis: Step 3.3. The CDS permits a `jwksEndpoint` on the Register that is different from the `jwks_uri` of the configuration ([CDS: Multi-Brand Support][cds-multi-brand]).
- Reason: One key set is the least complex.

### <a id="dh-0533"></a>DH-0533 A signing key that the data holder creates

- Kind: choice
- Basis: Step 3.3. The CDS does not say how a data holder gets its signing key ([CDS: JSON Web Key Set endpoint][cds-jwks-endpoint]).
- Reason: The key must stay the same after a restart, and [the repository contains no private key](non-functional-requirements.md#dh-0009). The lab database is not encrypted.

### <a id="dh-0534"></a>DH-0534 No cause on the error page of the authorisation endpoint

- Kind: choice
- Basis: Step 3.3. RFC 6749 says that the authorisation server informs the resource owner of the error, and it gives no content ([section 4.1.2.1][rfc6749-authorisation-errors]).
- Reason: The consumer cannot correct a fault in the request of the software product.

### <a id="dh-0535"></a>DH-0535 The sector identifier of a registration with no sector_identifier_uri

- Kind: conflict
- Basis: Step 3.2. OpenID Connect requires a `sector_identifier_uri` for redirect URIs with more than one host ([section 8.1][oidc-core-pairwise]), but the CDS marks the SSA claim as optional ([CDS: SSA Definition][cds-ssa-definition]).
- Reason: The software product ID gives one software product a `sub` that stays the same and that no other software product gets.

[cds-oidc-config]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#openid-provider-configuration-endpoint
[oidc-discovery-validation]: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationValidation
[cds-scopes-and-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#scopes-and-claims
[cds-authorisation-scopes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-scopes
[cds-multi-brand]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#multi-brand-support-separate-issuers-for-data-holder-brands
[rfc6749-authorisation-errors]: https://www.rfc-editor.org/rfc/rfc6749#section-4.1.2.1
[cds-sub-claim]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#sub-claim
[dsb-issue-480]: https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/480
[oidc-core-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[support-ppid]: https://cdr-support.zendesk.com/hc/en-us/articles/5505351487631
[cds-jwks-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#json-web-key-set-endpoint
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
