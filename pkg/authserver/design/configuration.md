# Authorisation server library: Configuration

This area gives the OpenID Provider configuration of each authorisation server, the JWKS and the signing keys.

## OpenID Provider configuration

- <a id="auth-0001"></a>**AUTH-0001** The OpenID Provider configuration is at `<issuer>/.well-known/openid-configuration`, and its `issuer` field is the issuer. ([OpenID Connect Discovery, section 4.3][oidc-discovery-validation])
- <a id="auth-0002"></a>**AUTH-0002** The response of the OpenID Provider configuration has `Cache-Control: no-cache`. ([RFC 9111: no-cache][rfc9111-no-cache]; [Judgement: no cache lifetime for the configuration](#auth-0018))

## Register configuration

The Register authorisation server serves Get OpenId Provider Config.

- <a id="auth-0003"></a>**AUTH-0003** The configuration of the Register authorisation server has each field that the OpenAPI schema of Get OpenId Provider Config marks as required, with these values. ([CDS: Get OpenId Provider Config][cds-get-oidc]; [CDS: Register Endpoints][cds-register-endpoints]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](#auth-0013); [Judgement: true values in the Register fields that clients can ignore](#auth-0014))
  - `issuer`: the issuer
  - `jwks_uri`: a URL on the public listener
  - `token_endpoint`: a URL on the mTLS listener
  - `grant_types_supported`: `client_credentials`
  - `token_endpoint_auth_methods_supported`: `private_key_jwt`
  - `token_endpoint_auth_signing_alg_values_supported`: `PS256`, `ES256`
  - `scopes_supported`: `cdr-register:read`
  - `tls_client_certificate_bound_access_tokens`: `true`
  - `claims_supported`: the claims of the access token, `iss`, `sub`, `aud`, `exp`, `iat`, `jti`, `client_id`, `scope` and `cnf`
  - `id_token_signing_alg_values_supported`: `PS256`
  - `subject_types_supported`: `public`
  - `response_types_supported`: `token`
  - `code_challenge_methods_supported`: an empty list

## Data holder configuration

- <a id="auth-0004"></a>**AUTH-0004** The configuration of the data holder authorisation server has each field that the CDS requires as a minimum, with these values. ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]; [CDS: Scopes and Claims][cds-scopes-and-claims]; [CDS: Request Object][cds-request-object]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](#auth-0013); [Judgement: only the CDS minimum values in acr_values_supported and claims_supported](#auth-0015); [Judgement: sharing_duration in claims_supported, as the CDS text says](#auth-0016))
  - `issuer`: the issuer
  - `jwks_uri` and `authorization_endpoint`: URLs on the public listener
  - `registration_endpoint`, `token_endpoint`, `userinfo_endpoint`, `introspection_endpoint`, `revocation_endpoint`, `pushed_authorization_request_endpoint` and `cdr_arrangement_revocation_endpoint`: URLs on the mTLS listener
  - `scopes_supported`: the scopes that the service gives
  - `claims_supported`: `sub`, `acr`, `auth_time`, `name`, `given_name`, `family_name`, `updated_at` and `sharing_duration`
  - `claims_parameter_supported`: `true`
  - `acr_values_supported`: `urn:cds.au:cdr:2`
  - `subject_types_supported`: `pairwise`
  - `grant_types_supported`: `authorization_code`, `client_credentials`, `refresh_token`
  - `response_types_supported`: `code`
  - `response_modes_supported`: `jwt`
  - `code_challenge_methods_supported`: `S256`
  - `token_endpoint_auth_methods_supported`: `private_key_jwt`
  - `token_endpoint_auth_signing_alg_values_supported` and `request_object_signing_alg_values_supported`: `PS256`, `ES256`
  - `id_token_signing_alg_values_supported` and `authorization_signing_alg_values_supported`: `PS256`
  - `tls_client_certificate_bound_access_tokens` and `require_pushed_authorization_requests`: `true`

- <a id="auth-0254"></a>**AUTH-0254** A data holder service can give a browser base URI. The `authorization_endpoint` is then on the browser base URI, and not on the issuer. When the service gives none, the `authorization_endpoint` is on the issuer. ([Judgement: a browser base URI for the authorisation endpoint](../../../services/dataholder/design/authorisation-server.md#dh-2673))
- <a id="auth-0005"></a>**AUTH-0005** The configuration of the data holder authorisation server has no field for the encryption of ID tokens or of authorisation responses. ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]; [Judgement: no encryption of ID tokens or authorisation responses](#auth-0017))

## JWKS

- <a id="auth-0006"></a>**AUTH-0006** The JWKS of the `jwks_uri` has the public key of each signing key that the service gives, and no private key. ([CDS: JSON Web Key Set endpoint][cds-jwks-endpoint])
- <a id="auth-0007"></a>**AUTH-0007** Each public key that the library publishes has the JWK members `alg`, `e`, `key_ops`, `kid`, `kty` and `n`. ([CDS: JWK][cds-schema-jwk])
- <a id="auth-0008"></a>**AUTH-0008** A service must give each signing key a different key ID. ([CDS: JSON Web Key Set endpoint][cds-jwks-endpoint])
- <a id="auth-0245"></a>**AUTH-0245** The response of the JWKS has `Cache-Control: no-cache`. ([RFC 9111: no-cache][rfc9111-no-cache]; [Judgement: no cache lifetime for the configuration](#auth-0018))
- <a id="auth-0246"></a>**AUTH-0246** When the service cannot give the signing keys, the JWKS endpoint responds with 500 and no body. ([RFC 9110, section 15.6.1][rfc9110-500])

## Signing keys

- <a id="auth-0009"></a>**AUTH-0009** The authorisation server signs with PS256 only. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms]; [Judgement: the server signs with PS256 and accepts PS256 or ES256](#auth-0013))
- <a id="auth-0010"></a>**AUTH-0010** The authorisation server signs with the first signing key that the service gives. ([Judgement: signing with the first key of the list](#auth-0019))
- <a id="auth-0011"></a>**AUTH-0011** A signing key that the library creates is an RSA key of 2048 bits. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms])
- <a id="auth-0012"></a>**AUTH-0012** A service must give only RSA signing keys of 2048 bits or longer. ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms])

## Judgements

### <a id="auth-0013"></a>AUTH-0013 The server signs with PS256 and accepts PS256 or ES256

- Kind: choice
- Basis: Step 3.3. FAPI 1.0 Advanced permits PS256 and ES256 ([section 8.6][fapi-algorithms]). Data recipients must support the two ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]). The CDS does not name the algorithms of the Register token endpoint, and it refers to FAPI 1.0 Advanced for the other endpoints.
- Reason: One algorithm for the signatures of the authorisation server is the least complex. A client can sign with each algorithm that FAPI 1.0 Advanced permits.

### <a id="auth-0014"></a>AUTH-0014 True values in the Register fields that clients can ignore

- Kind: choice
- Basis: Step 3.3. The schema requires the fields, and the CDS says that clients can ignore three of them ([CDS: Get OpenId Provider Config][cds-get-oidc]).
- Reason: The Register gives the values that are true for it. It supports no PKCE method.

### <a id="auth-0015"></a>AUTH-0015 Only the CDS minimum values in acr_values_supported and claims_supported

- Kind: choice
- Basis: Step 3.3. The CDS requires the two fields, and gives no full list of values ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]).
- Reason: The minimum values that the CDS requires are the least complex.

### <a id="auth-0016"></a>AUTH-0016 sharing_duration in claims_supported, as the CDS text says

- Kind: conflict
- Basis: Step 1. The CDS example of the configuration does not have the claim, but the CDS calls it a claim that the data holder must support ([CDS: Request Object][cds-request-object]).
- Reason: The CDS text is normative, and its example is not. The OpenID conformance suite also requires the claim for the CDR profile.

### <a id="auth-0017"></a>AUTH-0017 No encryption of ID tokens or authorisation responses

- Kind: choice
- Basis: Step 1. ID token encryption is only for the hybrid flow, which the CDS retired on 12 May 2025 ([CDS: Registration Request using JWT][cds-registration-request]). Encryption of authorisation responses is optional for a data holder ([CDS: OpenID Provider Configuration endpoint][cds-oidc-config]).
- Reason: No flow that the data holder supports needs the encryption. The data holder ignores the registration fields that it does not support ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]).

### <a id="auth-0018"></a>AUTH-0018 No cache lifetime for the configuration

- Kind: choice
- Basis: Step 3.3. The CDS gives no cache lifetime for the configuration ([CDS: Metadata Cache Management][cds-metadata-cache]).
- Reason: A client then checks for a changed document at each use, for example after a key change.

### <a id="auth-0019"></a>AUTH-0019 Signing with the first key of the list

- Kind: choice
- Basis: Step 3.2. A JWKS can hold more than one key ([RFC 7517, section 5][rfc7517-jwks]).
- Reason: Participants keep a copy of the JWKS, so a new key must be in the JWKS before the authorisation server signs with it. With a list, a service can publish the next key after the key that signs.

[oidc-discovery-validation]: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationValidation
[rfc9111-no-cache]: https://www.rfc-editor.org/rfc/rfc9111#name-no-cache-2
[rfc7517-jwks]: https://www.rfc-editor.org/rfc/rfc7517#section-5
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[support-dcr]: https://cdr-support.zendesk.com/hc/en-us/articles/6828251436431
[cds-get-oidc]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-openid-provider-config
[cds-register-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-endpoints
[cds-oidc-config]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#openid-provider-configuration-endpoint
[cds-scopes-and-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#scopes-and-claims
[cds-request-object]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-object
[cds-jwks-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#json-web-key-set-endpoint
[cds-schema-jwk]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_schemas_tocSjwk
[cds-registration-request]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-request-using-jwt
[cds-dcr-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dynamic-client-registration-endpoints
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[rfc9110-500]: https://www.rfc-editor.org/rfc/rfc9110#section-15.6.1
