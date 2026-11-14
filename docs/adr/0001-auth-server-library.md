# 0001. Use luikyv/go-oidc for all CDR authorisation servers

- Date: 2026-09-25

## Context

The mock Register needs an OAuth 2.0 authorisation server: `client_credentials`, `private_key_jwt` with keys from each client's remote JWKS URI, and mTLS certificate-bound access tokens (RFC 8705). Later, each mock data holder needs a FAPI 1.0 Advanced authorisation server as the CDS profiles it: PAR, PKCE, JARM, Dynamic Client Registration with an SSA, and CDS-specific parameters and endpoints.

We do not want to build our own authorisation server. We prefer an existing product, or a thin wrapper around an established library.

The research is in [Auth servers](../research/2026-09-25-auth-servers.md). In summary:

- Ory Hydra, Ory Fosite and zitadel/oidc do not support mTLS certificate-bound access tokens. They fail the Register needs.
- Keycloak meets the Register needs with configuration only. Each CDS-specific data holder feature needs a Java extension (about 3 to 6 weeks of work), and Keycloak adds a Java stack.
- No open-source product supports the CDS-specific data holder features without custom code. Only commercial products and the ACCC's own .NET server do. The ACCC server has a fixed UI and fixed storage, and it states CDS 1.34.0.
- `luikyv/go-oidc` is a Go library for OpenID Providers. It is MIT-licensed and on the OpenID FAPI 1.0 Advanced certification list (0.11.0 and later). It has hooks for storage, client authentication, DCR and mTLS. It is pre-1.0, and one person maintains it.
- `panva/node-oidc-provider` has the most established hooks, but it adds Node to the repository.
- `coreos/go-oidc` is a different kind of library. It is an OpenID Connect client (relying party) library that verifies ID tokens. It cannot be an authorisation server.

## Decision

Use `luikyv/go-oidc` for the Register authorisation server and for all data holder authorisation servers. Wrap it in a small library of our own, so that we can replace it.

A pre-1.0 library with one maintainer is acceptable for a lab environment.

## Consequences

- One authorisation server stack, in Go, for all participants.
- The features that the research did not confirm are now proved. See Verification.
- Our wrapper (`pkg/authserver`) must hide the library API from the rest of each service, because breaking changes before 1.0 are likely. The adapters that plug into go-oidc use go-oidc types, so they are in `pkg/authserver/internal`, where services cannot import them. Services use only the API that `pkg/authserver` exports, and that API has no go-oidc types.
- We run the OpenID conformance suite against our authorisation servers, because our configuration is not certified even if the library is.

## Verification

- Date: 2026-10-03
- Version: `luikyv/go-oidc` v0.25.0

The tests in [`pkg/authserver/internal/cdsrules`](../../pkg/authserver/internal/cdsrules) run each feature against an in-process FAPI 1.0 Advanced provider over mTLS: PAR, a signed request object, PKCE, `private_key_jwt` and certificate-bound tokens. Each workaround is a function in that package, and its tests fail without it.

| Feature | Result | How |
|---|---|---|
| DCR request as a JWT (`application/jwt`), as [CDS: Client Registration][cds-dcr] requires | Workaround | go-oidc accepts only JSON. `RegistrationJWT` verifies the JWT and gives go-oidc the claims as JSON. Unknown claims, for example `software_statement`, go to `ClientMeta.CustomAttributes`, where the DCR client handler can validate them. |
| Custom request object parameters (`sharing_duration`) reach our code | Workaround | go-oidc drops the request object claims that it does not know. `RequestObjectClaims` reads them for the PAR session handler, and go-oidc verifies the request object before it calls the handler. `SharingDuration` keeps the value in the session store, which go-oidc copies to the grant. |
| Refresh token lifetime for each grant | Workaround | go-oidc has one lifetime for all grants. `RefreshTokenForSharingDuration` sets the expiry for each grant, and issues no refresh token when `sharing_duration` is 0 or absent. The provider must not use `WithRefreshTokenLifetime` or `WithRefreshTokenRotation`, which replace the expiry. |
| JARM | Supported | `WithJARM`. A test verifies the signed authorisation response (`iss`, `aud`, `exp`, `code`, `state`) and exchanges its code for a certificate-bound token. |
| Registration management with a `cdr:registration` access token, as [CDS: Client Registration Management][cds-dcr-management] requires | Workaround | This gap was not in the research. go-oidc checks only the RFC 7592 registration access token. `RegistrationAccessToken` checks the access token (active, for the client, bound to the client certificate, with the scope), then gives go-oidc the registration access token of the client. |

Decision: keep `luikyv/go-oidc`. All five features work. The two DCR workarounds would also be necessary with `panva/node-oidc-provider`, because it also expects JSON registration requests and RFC 7592 registration access tokens. Upstream changes that would remove the workarounds are not offered to the go-oidc project for now.

Limits of the workarounds:

- `RequestObjectClaims` reads signed request objects only, not encrypted ones. The CDS requires signed request objects.
- The OpenID conformance suite has not run against this configuration yet (see Consequences).

[cds-dcr]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-request-using-jwt
[cds-dcr-management]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-management
