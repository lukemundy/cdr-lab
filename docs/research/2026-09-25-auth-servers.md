# Research: OAuth 2.0 / OIDC authorisation servers for cdr-lab

- Research date: 2026-09-25
- Scope: auth server for the mock CDR Register (need A) and for mock CDR Data Holders (need B).
- Target standards: Consumer Data Standards (CDS) 1.36.0; FAPI 1.0 Advanced with PAR, PKCE and JARM.

Check the sources again before you depend on a fact; products change.

## Versions examined

| Product / library | Version examined | Date of that version | Source |
|---|---|---|---|
| Keycloak | 26.7.4 | 2026-09-16 | https://github.com/keycloak/keycloak/releases |
| Ory Hydra (includes Fosite in `fosite/`) | v26.2.0 (GitHub release); changelog shows v26.3.x | 2026-03-20 | https://github.com/ory/hydra/releases , https://changelog.ory.com/ |
| Ory Fosite (standalone repo) | v0.49.0 | 2024-12-12 | https://github.com/ory/fosite/releases |
| zitadel/oidc | v3.51.6 | 2026-09-25 | https://github.com/zitadel/oidc |
| Zitadel | v4.19.1 | 2026-09-23 | https://github.com/zitadel/zitadel |
| panva/node-oidc-provider | v9.12.2 | 2026-09-05 | https://github.com/panva/node-oidc-provider |
| luikyv/go-oidc | v0.25.0 | 2026-07-23 | https://github.com/luikyv/go-oidc |
| ACCC ConsumerDataRight/authorisation-server | v3.1.0 | 2026-06-01 (changelog), release published 2026-07-15 | https://github.com/ConsumerDataRight/authorisation-server |
| ACCC ConsumerDataRight/mock-register | v2.2.8 | 2026-07-15 | https://github.com/ConsumerDataRight/mock-register |
| WSO2 Open Banking CDS Toolkit | 1.0.0 (docs site) | not stated | https://cds.ob.docs.wso2.com/en/latest/get-started/about-this-release/ |
| wso2/reference-implementation-consumerdatastandards-au | v1.0.0 | 2024-10-28 | https://github.com/wso2/reference-implementation-consumerdatastandards-au |
| Authlete | 2.2 (certified version); SaaS | n/a | https://www.authlete.com/pricing |
| Curity Identity Server | 6.6.0 / 7.1.0 (certified versions) | n/a | https://curity.io/product/plans/ |
| Connect2id server | 19.x (from blog titles) | n/a | https://connect2id.com/blog/connect2id-server-19-1 |

Repository metadata (licence, language, last push, latest release) came from the GitHub REST API (`gh api repos/<owner>/<repo>` and `/releases/latest`) on 2026-09-25.

## Primary sources

- OpenID Foundation, all certified implementations: https://openid.net/certification/all-certified-implementations/
- OpenID Foundation, FAPI1 OP Australia CDR list: https://openid.net/certification/certified-fapi1-op-australia-cdr/
- OpenID Foundation, FAPI1 Advanced OP list: https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/
- Keycloak FAPI support doc (source): https://github.com/keycloak/keycloak/blob/main/docs/guides/securing-apps/partials/oidc/fapi-support.adoc
- Keycloak default client profiles (source): https://github.com/keycloak/keycloak/blob/main/services/src/main/resources/keycloak-default-client-profiles.json
- Keycloak reverse proxy guide (source): https://github.com/keycloak/keycloak/blob/main/docs/guides/server/reverseproxy.adoc
- Keycloak sizing: https://www.keycloak.org/high-availability/multi-cluster/concepts-memory-and-cpu-sizing
- node-oidc-provider README: https://github.com/panva/node-oidc-provider
- node-oidc-provider configuration docs: https://github.com/panva/node-oidc-provider/blob/main/docs/README.md
- go-oidc README: https://github.com/luikyv/go-oidc
- zitadel/oidc README: https://github.com/zitadel/oidc
- Ory Hydra README: https://github.com/ory/hydra
- Ory Hydra OpenAPI spec: https://raw.githubusercontent.com/ory/hydra/master/spec/api.json
- Ory Hydra mTLS issue: https://github.com/ory/hydra/issues/1084
- Ory Hydra PAR discussion: https://github.com/ory/hydra/discussions/3071
- ACCC authorisation-server README: https://github.com/ConsumerDataRight/authorisation-server
- ACCC authorisation-server container help: https://github.com/ConsumerDataRight/authorisation-server/blob/main/Help/container/HELP.md
- ACCC authorisation-server changelog: https://github.com/ConsumerDataRight/authorisation-server/blob/main/CHANGELOG.md
- ACCC mock-register README: https://github.com/ConsumerDataRight/mock-register
- ACCC mock-data-holder README: https://github.com/ConsumerDataRight/mock-data-holder
- WSO2 CDS Toolkit docs: https://cds.ob.docs.wso2.com/en/latest/
- Authlete pricing: https://www.authlete.com/pricing
- Curity plans: https://curity.io/product/plans/ ; Community Edition: https://curity.io/product/community/
- Connect2id datasheet: https://connect2id.com/products/server/docs/datasheet
- Connect2id licensing: https://connect2id.com/products/server/license-and-support
- Connect2id FAPI page: https://connect2id.com/products/server/fapi

## Legend

- **yes**: the product supports the item with configuration only.
- **partial**: the product supports part of the item. The notes give the missing part.
- **ext**: you must write code in the product's extension point (plugin, SPI, hook, middleware).
- **no**: the product does not support the item. No extension point is practical.
- **UNVERIFIED**: I did not find a primary source. Do a check before you depend on it.

---

## 1. Requirements

### Need A: CDR Register auth server

- A1. `client_credentials` grant.
- A2. Client authentication with `private_key_jwt` (RFC 7523), PS256 and ES256.
- A3. The server gets client public keys from the client's remote `jwks_uri` (the software product JWKS). No manual key registration.
- A4. mTLS with certificate-bound access tokens (RFC 8705, `cnf.x5t#S256`). A private CA issues the client certificate. A proxy can terminate TLS and forward the certificate in a header.
- A5. OIDC discovery document and JWKS endpoint.
- A6. Scope `cdr-register:read`.
- A7. Create, update and remove clients from our own data (admin API or config). A status change must block token issue immediately.

### Need B: CDR Data Holder auth servers

- B1. FAPI 1.0 Advanced (final), as the CDS 1.36.0 profile gives it. Authorisation code flow only.
- B2. PAR (RFC 9126) mandatory.
- B3. PKCE S256.
- B4. JARM.
- B5. Request objects signed PS256 or ES256.
- B6. `private_key_jwt`.
- B7. mTLS certificate-bound tokens.
- B8. DCR (RFC 7591/7592). The software statement (SSA) is a JWT that the Register signs. The server must validate it against the Register JWKS.
- B9. Custom claims and parameters (`sharing_duration`, `cdr_arrangement_id`).
- B10. Refresh token lifetime tied to the sharing duration.
- B11. Token introspection and revocation.
- B12. CDR arrangement revocation endpoint.
- B13. Our own consent and authentication UI.
- B14. Our own arrangement storage.

CDS-specific behaviour (from the ACCC authorisation-server README, "Customisations" section, https://github.com/ConsumerDataRight/authorisation-server):

- The DCR request body is validated against the SSA and against the CDS. The SSA signature is verified against the Register SSA JWKS.
- PAR must extract and interpret `cdr_arrangement_id` and `sharing_duration` from the request object.
- The token endpoint must create a CDR arrangement from `sharing_duration`. The token response must include `cdr_arrangement_id`.
- The `sub` claim at UserInfo is a PPID.
- Introspection is only for refresh tokens. The introspection response must include `cdr_arrangement_id`.
- `/arrangement/revocation` is a custom CDS endpoint.

All general-purpose products below need extension code for these CDS-specific items. Only the ACCC server and the commercial CDR products do them out of the box.

---

## 2. Candidates

### 2.1 Keycloak

- Licence: Apache-2.0. Source: https://github.com/keycloak/keycloak
- Language: Java (Quarkus).
- Maintenance: very active. Release 26.7.4 on 2026-09-16. About 37,000 GitHub stars. Red Hat sponsors it (Red Hat build of Keycloak). Source: GitHub API, https://github.com/keycloak/keycloak/releases
- Certifications:
  - FAPI1 Advanced OP, Keycloak 15.0.2, certified 11-Oct-2021, several profile variants. Source: https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/
  - FAPI1 OP Australia CDR, Red Hat Keycloak 15.0.2, certified 13-May-2022. Source: https://openid.net/certification/certified-fapi1-op-australia-cdr/
  - The certified version is old (15.0.2). The current version (26.x) has no current CDR certification on the list.
- Explicit CDR statement: "Keycloak is compliant with the Australia Consumer Data Right Security Profile. If you want to apply the Australia CDR security profile, you need to use `fapi-1-advanced` profile ... If your client also applies PAR, make sure that client applies ... PKCE ... using a client profile with the `pkce-enforcer` executor." Source: https://github.com/keycloak/keycloak/blob/main/docs/guides/securing-apps/partials/oidc/fapi-support.adoc
- Built-in global client profiles (source: https://github.com/keycloak/keycloak/blob/main/services/src/main/resources/keycloak-default-client-profiles.json):
  - `fapi-1-baseline`: secure-session, pkce-enforcer, secure-client-authenticator, secure-client-uris, consent-required, full-scope-disabled.
  - `fapi-1-advanced`: secure-session, confidential-client, secure-client-authenticator, secure-client-uris, secure-request-object, secure-response-type, secure-signature-algorithm, secure-signature-algorithm-signed-jwt, consent-required, full-scope-disabled, holder-of-key-enforcer.
  - `fapi-ciba`, `fapi-2-security-profile`, `fapi-2-message-signing`, `fapi-2-dpop-security-profile`, `fapi-2-dpop-message-signing`, `oauth-2-1-for-confidential-client`, `oauth-2-1-for-public-client`, `saml-security-profile`.
  - There is no dedicated CDR profile. CDR = `fapi-1-advanced` + `pkce-enforcer` (same doc as above).
- mTLS behind a proxy: Keycloak has X.509 client certificate lookup providers for `apache`, `haproxy`, `nginx`, `traefik` (and `envoy`, UNVERIFIED in this doc version). You configure `spi-x509cert-lookup--<provider>--ssl-client-cert`. The option `trust-proxy-verification` lets the proxy do chain validation. The doc warns that a forged header can authenticate a client if the proxy does not overwrite the header. Source: https://github.com/keycloak/keycloak/blob/main/docs/guides/server/reverseproxy.adoc
- Client keys from JWKS URL: the client attribute `use.jwks.url` exists ("Use JWKS URL" in the admin console Keys tab). Source: `server-spi-private/src/main/java/org/keycloak/protocol/oidc/OIDCConfigAttributes.java` and `js/apps/admin-ui/src/clients/keys/Keys.tsx` in https://github.com/keycloak/keycloak
- DCR with software statement: I found no `software_statement` handling in the Keycloak source (GitHub code search for `software_statement` returned no match on 2026-09-25). Keycloak DCR takes a JSON body. The CDR DCR request is a signed JWT (`application/jwt`). You need a Java `ClientRegistrationProvider` / client registration policy SPI, or a proxy in front that validates the SSA and translates the request. Status: **ext**. The SPI names are from general Keycloak knowledge: UNVERIFIED for 26.x.
- CDR claims and parameters: no `cdr_arrangement_id` or `sharing_duration` in the Keycloak source (GitHub code search, 2026-09-25). Status: **ext** (Java SPIs: protocol mappers, authenticators, custom REST resource for arrangement revocation, token lifetime overrides).
- Client management: Admin REST API (create, update, delete, enable/disable clients). Realm import with JSON at start. A disabled client cannot get tokens. Tokens that were issued before stay valid until they expire (JWT access tokens). Admin REST API: general knowledge; UNVERIFIED for exact endpoint paths in 26.x.
- Custom UI: themes (FreeMarker) and Authenticator / Required Action SPIs (Java). An external consent UI is not a native pattern. Keycloak renders its own login and consent pages.
- Operational weight: "The base memory usage for a Pod including caches of Realm data and 10,000 cached sessions is 1250 MB of RAM." About 1 vCPU for each 120 client-credential grants per second. Source: https://www.keycloak.org/high-availability/multi-cluster/concepts-memory-and-cpu-sizing . Production needs an external database (PostgreSQL and others). Dev mode can use an embedded store.
- Go fit: none in code. It runs as a container. Go services talk to it with HTTP (admin API, OIDC endpoints).

| Req | Status | Notes |
|---|---|---|
| A1 client_credentials | yes | Service account clients. |
| A2 private_key_jwt PS256/ES256 | yes | Signed JWT client authenticator. `secure-signature-algorithm-signed-jwt` executor enforces FAPI algorithms. |
| A3 remote jwks_uri | yes | `use.jwks.url` client attribute. |
| A4 mTLS cert-bound via proxy header | yes | holder-of-key setting + x509cert-lookup provider. Private CA: the proxy or Keycloak truststore validates. |
| A5 discovery + JWKS | yes | Path is `/realms/<realm>/...`. The issuer contains `/realms/<realm>`. |
| A6 scope cdr-register:read | yes | Client scope. |
| A7 programmatic clients + immediate block | yes | Admin REST API; disable client. Issued tokens stay valid until expiry. |
| B1 FAPI 1 Adv | yes | `fapi-1-advanced` profile. |
| B2 PAR mandatory | yes | Per-client "require PAR" setting (general knowledge; UNVERIFIED key name). |
| B3 PKCE S256 | yes | `pkce-enforcer`. |
| B4 JARM | yes | Source: fapi-support doc and Red Hat docs. |
| B5 request object PS256/ES256 | yes | `secure-request-object`, `secure-signature-algorithm`. |
| B6 private_key_jwt | yes | |
| B7 mTLS cert-bound | yes | `holder-of-key-enforcer`. |
| B8 DCR + SSA | ext | No SSA support found. JWT request body not supported natively. Java SPI or front proxy. |
| B9 sharing_duration, cdr_arrangement_id | ext | Java SPI. |
| B10 refresh lifetime = sharing duration | ext | Java SPI. |
| B11 introspection + revocation | partial | Endpoints exist. CDS rules (refresh tokens only, `cdr_arrangement_id` in response) need ext. |
| B12 arrangement revocation | ext | Custom REST resource SPI. |
| B13 own consent UI | ext | Theme or Authenticator SPI. External UI is awkward. |
| B14 own arrangement storage | ext | Java SPI or external service call from SPI. |

### 2.2 Ory Hydra and Ory Fosite

- Licence: Apache-2.0. Language: Go. Source: https://github.com/ory/hydra
- Maintenance: Hydra is active (last push 2026-07-29; changelog shows v26.3.x releases). Source: GitHub API, https://changelog.ory.com/
- Fosite now lives inside the Hydra repository in the `fosite/` directory. The standalone `ory/fosite` repo had its last release v0.49.0 on 2024-12-12 and its last push on 2025-11-20. Source: `gh api repos/ory/hydra/contents` (directory `fosite` present), GitHub API for ory/fosite.
- Certifications: OIDC Basic, Implicit, Hybrid, Config, Dynamic OP. No FAPI certification. Source: https://github.com/ory/hydra (README), https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/ (Ory not listed).
- PAR: the Hydra public OpenAPI spec (master) lists no PAR endpoint. Public paths: `/.well-known/jwks.json`, `/.well-known/openid-configuration`, `/credentials`, `/oauth2/auth`, `/oauth2/device/auth`, `/oauth2/device/verify`, `/oauth2/register`, `/oauth2/register/{id}`, `/oauth2/revoke`, `/oauth2/sessions/logout`, `/oauth2/token`, `/userinfo`. Source: https://raw.githubusercontent.com/ory/hydra/master/spec/api.json . Fosite contains a PAR handler (`fosite/handler/par/flow_pushed_authorize.go`), but Hydra does not expose it. A 2022 maintainer reply said PAR was not planned. Source: https://github.com/ory/hydra/discussions/3071
- mTLS (RFC 8705): not supported. Issue #1084 was closed by a stale bot. On 2026-09-08 an outside contributor offered an unmerged fork implementation. Source: https://github.com/ory/hydra/issues/1084 . Fosite issue #661 is also closed (state reason "completed", but I found no merged feature). Source: https://github.com/ory/fosite/issues/661
- JARM: no evidence of support. Status **no** (UNVERIFIED that no hidden support exists).
- private_key_jwt with `jwks_uri`: supported (Fosite `JWT Profile for OAuth 2.0 Client Authentication` listed in README). Source: https://github.com/ory/hydra
- Client management: Admin API `/admin/clients`. Consent and login: Hydra redirects to your own login and consent apps (headless). This is a strong pattern for B13. Source: https://github.com/ory/hydra
- Weight: one Go binary. Needs PostgreSQL, MySQL or CockroachDB (SQLite for dev). Source: https://github.com/ory/hydra
- Enterprise: Ory sells an "Ory Enterprise License" (OEL) build. I did not check if OEL adds PAR or mTLS: UNVERIFIED. Source for OEL changelog: https://www.ory.com/docs/self-hosted/oel/oauth2/changelog/v26.3.12

| Req | Status | Notes |
|---|---|---|
| A1 | yes | |
| A2 | yes | PS256/ES256 support: UNVERIFIED. |
| A3 | yes | Client `jwks_uri`. |
| A4 | no | No RFC 8705. |
| A5 | yes | |
| A6 | yes | |
| A7 | yes | Admin API. |
| B1 | no | No FAPI certification; missing PAR, JARM, mTLS. |
| B2 | no | Not exposed. |
| B3 | yes | |
| B4 | no | |
| B5 | partial | Request objects supported; FAPI rules UNVERIFIED. |
| B6 | yes | |
| B7 | no | |
| B8 | ext | DCR exists; SSA validation needs a front proxy. |
| B9–B10 | partial/ext | Consent app can set session claims; custom lifetime UNVERIFIED. |
| B11 | yes | Standard endpoints. |
| B12 | ext | |
| B13 | yes | Headless login/consent apps. |
| B14 | ext | |

Verdict: Hydra fails A4 and most of B. Do not use. Using Fosite directly means you build PAR endpoint wiring, JARM and mTLS yourself. That is building an auth server.

### 2.3 zitadel/oidc (library) and Zitadel (product)

- zitadel/oidc licence: Apache-2.0. Language: Go. Very active (v3.51.6 on 2026-09-25). Source: GitHub API.
- Zitadel product licence: AGPL-3.0. Language: Go. Active (v4.19.1 on 2026-09-23). Source: GitHub API.
- zitadel/oidc OP feature table: Code flow yes, Client Credentials yes, Refresh yes, Discovery yes, JWT Profile (RFC 7523) yes, PKCE yes, Token Exchange yes, Device yes, **mTLS "not yet"**, Hybrid "not yet". No PAR or JARM rows. Source: https://github.com/zitadel/oidc (README "Features").
- README note: "We currently have limited availability for feature reviews." Source: https://github.com/zitadel/oidc
- Certification: OIDC RP basic and config (library). No FAPI certification. Source: https://github.com/zitadel/oidc ; https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/ (not listed)
- Zitadel product private_key_jwt: keys are uploaded to Zitadel, or Zitadel generates them. Remote `jwks_uri` for client auth: UNVERIFIED (not found in docs). Source: https://zitadel.com/docs/apis/openidoauth/authn-methods

| Req | Status | Notes |
|---|---|---|
| A1, A2, A5, A6 | yes | |
| A3 | UNVERIFIED | Library: your storage returns keys, so you can fetch remote JWKS yourself (ext). |
| A4 | no | mTLS "not yet". |
| A7 | yes (library: your storage) | |
| B1, B2, B4, B7 | no | |
| others | ext or no | |

Verdict: not suitable for A (no mTLS) or B (no PAR, JARM, mTLS, FAPI).

### 2.4 panva/node-oidc-provider

- Licence: MIT. Language: JavaScript (Node.js). Source: https://github.com/panva/node-oidc-provider
- Maintenance: active. v9.12.2 on 2026-09-05. Last push 2026-09-14. About 3,800 stars. One maintainer (Filip Skokan). README: "The only way to guarantee you get feedback from the author & sole maintainer of this module is to support the package through GitHub Sponsors." Source: GitHub API, README.
- Certifications: README says Filip Skokan certified conformance for Basic, Implicit, Hybrid, Config, Dynamic, Form Post, FAPI 1.0, FAPI CIBA and FAPI 2.0 profiles. Source: https://github.com/panva/node-oidc-provider . The OpenID FAPI1 Advanced list shows "Filip Skokan / node oidc-provider 7.4.0", certified 11-Jun-2021. Source: https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/ . Not on the Australia CDR list. Source: https://openid.net/certification/certified-fapi1-op-australia-cdr/ . The current major (v9) has no newer FAPI 1 entry that I found: UNVERIFIED.
- Specs listed in README: OAuth 2.0, OIDC Core, Discovery, RFC 8414, DCR (OIDC, RFC 7591, RFC 7592), PKCE, JAR (RFC 9101), PAR (RFC 9126), Device (RFC 8628), Revocation (RFC 7009), Introspection (RFC 7662), mTLS (RFC 8705), JWT access tokens (RFC 9068), DPoP (RFC 9449), FAPI 1.0, FAPI 2.0 and Message Signing, JARM, CIBA, back-channel logout, RP-initiated logout. Source: https://github.com/panva/node-oidc-provider
- Configuration hooks (source for all: https://github.com/panva/node-oidc-provider/blob/main/docs/README.md):
  - `features.fapi.profile`: `'1.0 Final'` gives FAPI 1.0 Part 2 Advanced behaviour. It can be a function `(ctx, client)`.
  - `features.pushedAuthorizationRequests` with `requirePushedAuthorizationRequests`.
  - `features.jwtResponseModes` (JARM).
  - `features.mTLS` with `certificateBoundAccessTokens`, `tlsClientAuth`, `selfSignedTlsClientAuth`, and `getCertificate(ctx)`. `getCertificate` returns an `X509Certificate` or PEM string. You write it, so it can read a proxy header. "Returning `undefined` causes authentication to fail wherever a certificate is required."
  - `features.registration` with `initialAccessToken`, `issueRegistrationAccessToken`, `idFactory`, `policies`. Policies are "assigned to Initial Access Tokens". CDR DCR has no initial access token, so policies do not fit directly. Use Koa middleware in front of the registration route to verify the SSA and to decode the `application/jwt` body.
  - `features.registrationManagement` (RFC 7592).
  - `features.introspection`, `features.revocation`.
  - `extraParams`: extra authorisation and PAR parameters with validator functions. Values go to `ctx.oidc.params` and interaction details. Fits `sharing_duration` and `cdr_arrangement_id`.
  - `extraTokenClaims`: extra claims in tokens.
  - `ttl`: per-artifact lifetimes. `RefreshToken` can be a callback, so it can use the sharing duration.
  - `rotateRefreshToken`.
  - `interactions.url` and `interactions.policy`: the provider redirects to your own login and consent UI. Your UI completes the interaction through the provider API.
  - `adapter`: pluggable storage. Default is in memory, "suitable for development and testing purposes only."
- Remote client `jwks_uri`: supported as standard client metadata (general knowledge; UNVERIFIED cache refresh behaviour on unknown `kid`).
- Client management: static `clients` array in config, or the adapter `Client` model backed by your own DB, or DCR. Immediate block after status change through adapter: depends on the provider client cache. UNVERIFIED.
- CDS gaps (all ext, all in JS): SSA validation and JWT request body at DCR; `cdr_arrangement_id` in the token response body; introspection only for refresh tokens with `cdr_arrangement_id`; PPID `sub`; `/arrangement/revocation` route (add a Koa route next to the provider).
- Weight: one Node process. Storage is what your adapter uses (PostgreSQL, Redis, and others). Small memory footprint (UNVERIFIED figure).
- Go fit: separate Node service in the repo. It is a second language and toolchain (npm, Node LTS, lint, tests). Not a Go library.

| Req | Status | Notes |
|---|---|---|
| A1 | yes | `client_credentials` grant. |
| A2 | yes | `private_key_jwt`; algorithms configurable (`enabledJWA`). |
| A3 | yes | Client `jwks_uri`. |
| A4 | yes | `features.mTLS` + your `getCertificate` reads the proxy header. |
| A5 | yes | |
| A6 | yes | `scopes` config. |
| A7 | yes / UNVERIFIED immediacy | Adapter or static config. Check client caching. |
| B1 | yes | `fapi.profile = '1.0 Final'`. Certified FAPI 1 Adv (7.4.0, 2021). |
| B2 | yes | |
| B3 | yes | |
| B4 | yes | |
| B5 | yes | |
| B6 | yes | |
| B7 | yes | |
| B8 | ext | Koa middleware for SSA and JWT body. |
| B9 | yes (params) / ext (token response field) | `extraParams`, `extraTokenClaims`. |
| B10 | yes | `ttl.RefreshToken` callback. |
| B11 | yes / ext for CDS rules | |
| B12 | ext | Koa route. |
| B13 | yes | `interactions.url`. |
| B14 | yes | `adapter` + your own tables for arrangements. |

### 2.5 luikyv/go-oidc (additional candidate, Go)

- Licence: MIT. Language: Go. Source: https://github.com/luikyv/go-oidc
- Maintenance: active. v0.25.0 on 2026-07-23. Last push 2026-09-07. Frequent minor releases (v0.18.0 on 2026-04-18 to v0.25.0 on 2026-07-23: 8 releases in 3 months). Pre-1.0, so API changes are likely. About 115 stars. One main maintainer: `luikyv` has 519 commits; the next human contributor has 6. Source: GitHub API (releases, contributors).
- Certifications: FAPI1 Advanced OP list shows "Luiky Vasconcelos / go-oidc >= 0.11.0", certified 10-Aug-2025. Source: https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/ . README claims Basic, Implicit, Hybrid, Config, Dynamic OP, FAPI 1.0 and FAPI 2.0. Source: https://github.com/luikyv/go-oidc . Not on the Australia CDR list. Source: https://openid.net/certification/certified-fapi1-op-australia-cdr/
- Specs in README: OIDC Core, Discovery, DCR (OIDC, RFC 7591, RFC 7592), RFC 9068, PAR, JAR, JARM, PKCE, RFC 9207, RFC 8705 mTLS and cert-bound tokens, DPoP, RAR, Resource Indicators, Token Exchange, Device, Introspection, Revocation, FAPI 1.0 Baseline and Advanced, FAPI 2.0, CIBA, OpenID Federation 1.1, OID4VCI. Source: https://github.com/luikyv/go-oidc
- Hooks (source: README):
  - `provider.New(issuer, manager, jwksFunc, opts...)`. `op.Handler()` returns an `http.Handler`. You can mount it in your own mux and add your own routes (for example `/arrangement/revocation`).
  - Storage: you pass a manager that implements the client, grant and token interfaces. `storage.NewManager` is an in-memory example.
  - `provider.WithClientCredentialsGrant()`, `WithAuthCodeGrant`, `WithRefreshTokenGrant`, `RefreshTokenRotation()`, `RefreshTokenLifetime(...)`.
  - `goidc.Grant` has `Store` (map for custom data) and `RefreshTokenExpiresAt`. Per-grant refresh lifetime from `sharing_duration`: UNVERIFIED (the field exists in `pkg/goidc/grant.go`; I did not confirm you can set it per grant).
  - `provider.WithTokenClaims(...)`, `WithIDTokenClaims`, `WithUserInfoClaims`.
  - `provider.WithMTLS(mtlsHost, func(ctx) (*x509.Certificate, error))`: you write the function, so it can read a proxy header.
  - `provider.WithJAR(...)`, `provider.WithJARM(...)` with PS256.
  - `provider.WithDCR(manager)`, `DCRInitialTokenValidator`, `DCRClientHandler` ("implement your registration policy during create and update requests"). SSA validation fits in `DCRClientHandler`. JWT (`application/jwt`) request body: UNVERIFIED; likely needs a small HTTP middleware.
  - Authentication: `goidc.AuthnPolicy` with Setup and Authenticate functions. Your code renders the UI or redirects to an external UI.
  - Client `JWKSURI` field exists (`pkg/goidc/client.go`). Source: GitHub code search, 2026-09-25.
  - Custom authorisation parameters (`sharing_duration`, `cdr_arrangement_id` inside the request object): UNVERIFIED how they reach the `AuthnPolicy`.
- Weight: in-process Go library. Storage is your own (for example PostgreSQL through your existing repository layer).
- Go fit: best of all candidates. It fits the hexagonal layout: the library is a driving adapter; the manager interfaces are implemented by driven adapters.

| Req | Status | Notes |
|---|---|---|
| A1 | yes | |
| A2 | yes | |
| A3 | yes | Client `JWKSURI`. |
| A4 | yes | `WithMTLS` + your cert function. |
| A5 | yes | |
| A6 | yes | `WithScopes`. |
| A7 | yes | Your manager reads your DB, so a status change applies on the next request. |
| B1 | yes | FAPI 1 Adv certified (>= 0.11.0). |
| B2–B7 | yes | |
| B8 | ext | `DCRClientHandler`; JWT body UNVERIFIED. |
| B9 | ext / UNVERIFIED | `Grant.Store`, `WithTokenClaims`. |
| B10 | UNVERIFIED | |
| B11 | yes / ext for CDS rules | |
| B12 | ext | Own handler on the same mux. |
| B13 | yes | `AuthnPolicy`. |
| B14 | yes | Your manager + your tables. |

Risk: single maintainer, pre-1.0, small user base. Mitigate: pin the version, run the OpenID conformance suite (it has an AU-CDR FAPI plan) in CI, and keep the wrapper thin so a swap stays cheap.

### 2.6 ACCC ConsumerDataRight/authorisation-server

- Licence: MIT. Language: C# (.NET 8). Source: https://github.com/ConsumerDataRight/authorisation-server
- Maintenance: ACCC-maintained. v3.1.0 (changelog date 2026-06-01; GitHub release 2026-07-15). Regular vulnerability patch releases (3.0.2 2025-10-15, 3.0.3 2025-12-03, 3.0.4 2025-12-10, 3.0.5 2026-03-18, 3.0.6 2026-06-17). About 8 stars. Source: https://github.com/ConsumerDataRight/authorisation-server/blob/main/CHANGELOG.md , GitHub API.
- Standards: README says it "aligns to v1.34.0 of the Consumer Data Standards", "FAPI 1.0 Migration Phase 4", passed v4.4.0 of the Conformance Test Suite for Data Recipients, and "is certified with the FAPI 1.0 Advanced Profile" (AU-CDR Adv. OP w/ Private Key, PAR, JARM). Source: README. The README says 1.34.0, not the lab target 1.36.0. Whether 1.35/1.36 changes affect the auth server: UNVERIFIED. The sibling mock-register "Aligns to v1.36.0". Source: https://github.com/ConsumerDataRight/mock-register
- Certification: "ACCC / CDR Auth Server v1.0", FAPI1 OP Australia CDR, certified 20-Feb-2023. Source: https://openid.net/certification/certified-fapi1-op-australia-cdr/
- Endpoints (source: README): discovery, `/jwks`, `/auth`, `/registration` (POST, GET, PUT, DELETE, mTLS), `/par` (mTLS, client assertion), `/token` (authorization_code, client_credentials, refresh_token), `/userinfo`, `/introspect`, `/token/revocation`, `/arrangement/revocation`.
- Components: auth server (port 8001), mTLS Gateway (Ocelot, port 8002, "Performs certificate validation"), React UI (port 3000) with OTP and account selection, an Azure Function for the Get Data Recipients discovery process, and a SQL repository "containing clients, grants and log entries". Source: README.
- Storage: Microsoft SQL Server (container `mcr.microsoft.com/mssql/server:2022-latest`, EULA). Source: https://github.com/ConsumerDataRight/authorisation-server/blob/main/Help/container/HELP.md
- Customer data: JSON seed files (`customer-seed-data.json`, `customer-seed-data-energy.json`). The UI reads them too. Source: README.
- Register dependency: DCR validates the SSA against the Register SSA JWKS. Source: README. So it works against any Register that serves the CDR Register APIs, including ours. This is also a good integration test of our Register.
- Custom consent UI and own arrangement storage: no extension points. You must fork the C# and React code. Status: **no** (without a fork).
- Operational weight: .NET service + Ocelot gateway + React UI + MSSQL. MSSQL is heavy (Microsoft recommends at least 2 GB RAM: UNVERIFIED for 2022 image).
- Go fit: none in code; runs as containers. Image: `consumerdataright/authorisation-server` on Docker Hub. Source: README.
- Need A: it has `client_credentials`, but it is a Data Holder auth server. It is not built to be the Register auth server. The Register auth is inside `ConsumerDataRight/mock-register` (Identity Provider, .NET 8, `client_credentials` + `private_key_jwt`, mTLS gateway, MSSQL). Source: https://github.com/ConsumerDataRight/mock-register . That IdP is coupled to the mock-register database. You cannot use it for our Register without adopting the whole ACCC mock-register.

| Req | Status | Notes |
|---|---|---|
| A1–A7 | partial | Not designed for the Register role. Use mock-register instead if you want ACCC code for A. |
| B1–B12 | yes | Out of the box, CDS-customised, conformance-tested. CDS 1.34.0 alignment. |
| B13 | no | Fixed React UI. Fork needed. |
| B14 | no | Fixed MSSQL schema. Fork needed. |

### 2.7 WSO2 Open Banking CDS Toolkit

- Licence: reference implementation repo Apache-2.0. Source: https://github.com/wso2/reference-implementation-consumerdatastandards-au . WSO2 production updates for Open Banking usually need a WSO2 subscription: UNVERIFIED.
- Language: Java.
- Maintenance: the reference repo had its last release v1.0.0 on 2024-10-28, last push 2026-08-06, about 4 stars. Source: GitHub API.
- Base: WSO2 Open Banking Accelerator 3.0 on WSO2 Identity Server, WSO2 API Manager and WSO2 Streaming Integrator. Needs Java, Maven and MySQL. Source: https://cds.ob.docs.wso2.com/en/latest/get-started/wso2-open-banking-cds-toolkit/ , https://github.com/wso2/reference-implementation-consumerdatastandards-au
- CDS version: the docs "About this release" page says CDS v1.8.0. That is very old (target is 1.36.0). Source: https://cds.ob.docs.wso2.com/en/latest/get-started/about-this-release/ . A newer supported version may exist: UNVERIFIED.
- Certification: "WSO2 LLC / Open Banking 3.0", FAPI1 Advanced OP, certified 13-Jun-2023. Not on the Australia CDR list. Source: https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/
- Features: claims FAPI, consent management, application onboarding (DCR). Source: https://cds.ob.docs.wso2.com/en/latest/get-started/wso2-open-banking-cds-toolkit/
- Weight: very heavy (three WSO2 products + MySQL). Build from source with Maven; no documented container path in the reference repo.
- Verdict: too heavy and the documented CDS version is years old. Do not use.

### 2.8 Authlete (commercial)

- Model: headless OAuth/OIDC engine as a service. Your front-end server calls the Authlete API. Source: https://www.authlete.com/product
- Pricing: 30-day free trial (up to 3 services, 100 clients). Business from "$999+/month". Enterprise (managed or self-managed) at custom price. FAPI is an option on all plans. Cluster regions: US, Japan, Brazil, EU (no Australia). Source: https://www.authlete.com/pricing
- Certifications: FAPI1 Advanced OP (Authlete 2.2, 11-Jun-2021) and FAPI1 OP Australia CDR (Authlete 2.2, 11-Jun-2021). Sources: https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/ , https://openid.net/certification/certified-fapi1-op-australia-cdr/
- CDS-specific features (`cdr_arrangement_id`, arrangement revocation): UNVERIFIED.
- Verdict: no permanent free tier. SaaS dependency for a lab. Cost is out of proportion. Do not use.

### 2.9 Curity Identity Server (commercial)

- Community Edition is free "for as long as you need it". Source: https://curity.io/product/community/
- FAPI 1 Baseline and Advanced and FAPI 2 are options in Standard and Enterprise editions, not in Community. Source: https://curity.io/product/plans/ (via search summary; check the page directly: UNVERIFIED detail).
- Certifications: Curity Identity Server 6.6.0 (16-Nov-2021) and 7.1.0 (24-Jun-2022), FAPI1 Advanced OP. Not on the Australia CDR list. Source: https://openid.net/certification/certified-fapi1-adv-openid-providers-profiles/
- DCR request validation example (Open Banking Brasil SSA style): https://github.com/curityio/example-dcr-request-validation/
- Verdict: the free edition lacks FAPI. Paid for B. Do not use.

### 2.10 Connect2id server (commercial)

- Licence: free "Testing, Evaluation, and Development License" in the download package, for development, testing and CI. No registration. Production needs a paid licence (search summary mentions a one-time USD 9,999 and/or USD 299–999 per month; the licence page itself says "contact sales": UNVERIFIED figures). Sources: https://connect2id.com/products/server/license-and-support , https://connect2id.com/products/server/download
- Language: Java (servlet). Storage: LDAP, SQL, DynamoDB, Redis. Source: https://connect2id.com/products/server/docs/datasheet
- Features (datasheet): PAR, JARM, JAR, DCR (RFC 7591/7592) with software statement handling, `tls_client_auth` and `self_signed_tls_client_auth`, cert-bound tokens (`cnf.x5t#S256`), `private_key_jwt` with `jwks_uri`, introspection, revocation, `client_credentials`, FAPI 1.0 Advanced. Plugin SPIs for claims sourcing, consent, token encoding, authorisation request validation and client registration interceptors. Source: https://connect2id.com/products/server/docs/datasheet . The datasheet summary also said "Australia CDR compliance": UNVERIFIED (the automated summary may be wrong).
- FAPI config guide for 11.6+: https://connect2id.com/blog/connect2id-server-11-6 ; FAPI checklist: https://connect2id.com/products/server/docs/guides/fapi-checklist
- Certification: JARM certification mentioned by vendor. Not on the FAPI1 Advanced OP list or Australia CDR list. Source: https://connect2id.com/blog/connect2id-server-11 , OpenID lists above.
- Login and consent: Connect2id uses a headless authorisation session API; your own UI drives it (general knowledge; UNVERIFIED detail).
- Verdict: technically strong. The dev licence fits a lab, but the project is not open source and a "real" component would need a paid licence later. Keep as a fallback.

### 2.11 Other CDR-certified products (commercial, for context)

From https://openid.net/certification/certified-fapi1-op-australia-cdr/ :

| Organisation | Implementation | Version | Date |
|---|---|---|---|
| ACCC | CDR Auth Server | v1.0 | 20-Feb-2023 |
| Authlete | Authlete | 2.2 | 11-Jun-2021 |
| Biza Pty Ltd | Biza HaaS, Apollo Release | — | 31-Jan-2023 |
| Cloudentity | Cloudentity | November 2021 | 01-Dec-2021 |
| Cloudentity, Inc. | Cloudentity | August 2022 | 19-Aug-2022 |
| Fiskil | Fiskil Data Provider | Sept 2025 | 16-Jul-2025 |
| IBM | IBM Security Verify | May 2022 | 18-Jul-2022 |
| IBM | IBM Security Verify Access | 10.0 | 29-Oct-2022 |
| PayPal Inc. | PayPal | December 2022 | 28-Dec-2022 |
| Ping Identity Corporation | PingFederate | 10.3 | 25-Aug-2021 |
| Red Hat | Keycloak | 15.0.2 | 13-May-2022 |

- Cloudentity is now part of SecureAuth (CDR docs: https://docs.secureauth.com/ciam/en/consumer-data-right-403790.html). Commercial SaaS.
- Ping has a CDR sandbox repo: https://github.com/pingidentity/pingidentity-cdr-sandbox (commercial product).
- Fiskil/cdr (MIT, Go) is a data recipient client toolkit, not an auth server. Source: https://github.com/Fiskil/cdr
- ConsumerDataStandardsAustralia/mock-data-holder-nodejs uses the ACCC `cdr-auth-server` component, not its own. Source: https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs

---

## 3. Comparison summary

| Candidate | Licence | Lang | Latest | FAPI1 Adv cert | AU-CDR cert | A fit | B fit | Weight |
|---|---|---|---|---|---|---|---|---|
| Keycloak | Apache-2.0 | Java | 26.7.4 (2026-09) | 15.0.2 (2021) | 15.0.2 (2022) | Full, config only | FAPI yes; CDS items need Java SPIs | Heavy: ~1.25 GB RAM + DB |
| node-oidc-provider | MIT | Node | 9.12.2 (2026-09) | 7.4.0 (2021) | no | Full | FAPI yes; CDS items via JS hooks | Light |
| go-oidc (luikyv) | MIT | Go | 0.25.0 (2026-07) | >=0.11.0 (2025) | no | Full | FAPI yes; CDS items via Go hooks | Light, in-process |
| ACCC authorisation-server | MIT | C# | 3.1.0 (2026) | via AU-CDR | v1.0 (2023) | Not its role | Full CDS (1.34.0); no own UI or storage | Medium-heavy: .NET + MSSQL |
| Ory Hydra / Fosite | Apache-2.0 | Go | 26.x | no | no | Fails A4 | Fails | Light + DB |
| zitadel/oidc, Zitadel | Apache-2.0 / AGPL-3.0 | Go | 3.51.6 / 4.19.1 | no | no | Fails A4 | Fails | Light |
| WSO2 CDS Toolkit | Apache-2.0 (+subscription?) | Java | 1.0.0 (2024) | OB 3.0 (2023) | no | Overkill | Old CDS docs | Very heavy |
| Authlete | Commercial | SaaS | — | 2.2 | 2.2 | Yes | Yes (CDS items UNVERIFIED) | SaaS, $999+/mo |
| Curity | Commercial | Java | — | 6.6/7.1 | no | Paid edition | Paid edition | Medium |
| Connect2id | Commercial (free dev licence) | Java | 19.x | no (JARM only) | no | Yes | Strong SPIs | Medium |

---

## 4. Cost of gaps

Estimates are engineering time for one experienced engineer. They are judgement, not measured.

- Keycloak for B: a Java extension module with a DCR provider (JWT body, SSA check against Register JWKS, CDS metadata rules), a PAR/authorisation hook for `sharing_duration` and `cdr_arrangement_id`, an arrangement store, token response and introspection changes, a custom REST resource for `/arrangement/revocation`, PPID mapper, and a consent Authenticator plus theme. Estimate: 3–6 weeks. You also add a Java/Maven toolchain to a Go repo, and you must track Keycloak SPI changes across majors.
- node-oidc-provider for B: Koa middleware for DCR (JWT body + SSA), config for FAPI/PAR/JARM/mTLS, `extraParams`, `ttl.RefreshToken` callback, adapter for your DB, token-response and introspection post-processing, one extra route, and an interaction app. Estimate: 1–2 weeks. You add Node to the repo.
- go-oidc for B: same shape as node-oidc-provider, in Go, with more UNVERIFIED items (JWT DCR body, custom request object params, per-grant refresh lifetime). Estimate: 1–3 weeks, plus the risk that you must patch or fork the library.
- ACCC authorisation-server for B: zero code to run. Custom UI and storage need a C# fork. Estimate for a fork that swaps UI and storage: 2–4 weeks, plus .NET and MSSQL skills.
- Hydra for A or B: you must add RFC 8705 (and for B also PAR exposure and JARM) to Hydra or a fork. That is building an auth server. Not worth it.

---

## 5. Recommendation

### Need A (Register)

1. **Keycloak** meets every A requirement with configuration only: `client_credentials`, signed-JWT client auth with `use.jwks.url`, holder-of-key tokens with certificate lookup from a proxy header, a client scope, and the Admin REST API to sync clients and disable them. Keycloak also documents CDR profile compliance. The cost is weight (about 1.25 GB RAM plus a database) and a JVM product in a Go lab.
2. **node-oidc-provider** meets every A requirement with a small amount of JS config. It is light. It adds Node to the repo.
3. **go-oidc** meets every A requirement in Go, in-process, with your own storage. Status changes apply on the next request by design. The risk is maturity (pre-1.0, one maintainer).

Opinion: pick the product for A that you will also use for B. Two auth-server stacks double the work to operate, patch and learn.

### Need B (Data Holders)

- No open-source general-purpose product does the CDS-specific items (SSA-based DCR, `cdr_arrangement_id`, `sharing_duration`, arrangement revocation, refresh-only introspection) without code.
- The only zero-code option is the **ACCC authorisation-server**. It does not let you plug in your own consent UI or arrangement storage, it needs MSSQL, and it states CDS 1.34.0. Use it as a reference and as a conformance test peer, not as the long-term base.
- **node-oidc-provider** has the best hook coverage for B13, B14, B9 and B10, and it is the most established certified library.
- **go-oidc** has similar hooks in Go. It is the best fit with the codebase, and the weakest on maturity.
- **Keycloak** is the weakest fit for B. Every CDS item becomes a Java SPI.

### Same product for A and B

- node-oidc-provider: yes. One Node service, two configurations (Register profile and Data Holder profile).
- go-oidc: yes. One `pkg/` module that wraps it, used by the Register and by each Data Holder. This matches the README rule that `pkg/` holds CDR protocol behaviour.
- Keycloak: A yes; B only with a large Java extension.

Final opinion: do a time-boxed spike (1–2 days) with go-oidc against the OpenID conformance suite AU-CDR plan. If it passes and the UNVERIFIED items (JWT DCR body, custom request object params, per-grant refresh lifetime) work, use go-oidc for both A and B. If the spike fails, use node-oidc-provider for both. Use Keycloak for A only if you need A this week with zero code and accept a second stack later.
