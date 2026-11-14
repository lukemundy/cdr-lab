# 0009. Serve client registration outside the OpenID Provider library

- Date: 2026-10-09

## Context

A data holder authorisation server uses `luikyv/go-oidc` ([ADR 0001](0001-auth-server-library.md)). That ADR proved that the library can serve Dynamic Client Registration as the CDS profiles it, with two workarounds: one changes the request JWT into the JSON that the library accepts, and one changes the `cdr:registration` access token into the RFC 7592 registration access token that the library checks.

The CDS registration rules are different from RFC 7591 and RFC 7592 in more places than the two that ADR 0001 tested ([CDS: Client Registration](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#client-registration)):

- The request has a Software Statement Assertion (SSA). The data holder verifies it with the keys of the Register, and then verifies the request with the keys at the `jwks_uri` of the SSA. The values of the SSA have priority over the values of the request.
- The response has the SSA values as top-level members and `client_id_issued_at`. The CDS schema has no `registration_access_token` and no `registration_client_uri` ([`cds_dcr.json`](https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_dcr.json)).
- The errors use four codes, of which three apply to a data holder that approves each valid SSA: `invalid_software_statement`, `invalid_redirect_uri` and `invalid_client_metadata`.

The library at v0.25.0 does this ([`internal/dcr/util.go`](https://github.com/luikyv/go-oidc/blob/v0.25.0/internal/dcr/util.go), [`internal/dcr/model.go`](https://github.com/luikyv/go-oidc/blob/v0.25.0/internal/dcr/model.go)):

- It has no SSA. A client handler gets the request claims after the library validates them.
- It changes each error of the client handler into `invalid_client_metadata`.
- Its response always has `registration_client_uri`, and has `registration_access_token` at creation. It has no `client_id_issued_at`.
- It creates the client ID, and it calls a store that we supply to save, read and delete a client.

We compared two options:

| | Through the library | Outside the library |
|---|---|---|
| SSA rules, SSA priority, scope filter | In our code, before the library gets the request | In our code |
| Error codes | A workaround that changes the error response of the library | In our code |
| Response | A workaround that removes two members and adds one | In our code |
| Client ID and storage | The library creates the ID and calls our store | In our code |
| Workarounds that depend on the behaviour of the library | Four: the request format, the access token, the error codes and the response | None |
| Registration rules of RFC 7591 that the library checks | Used, but the CDS rules replace most of them | Not used |
| Later features of the library for registration | Available | Not available |

With the first option, our code validates the full request, then changes the request for the library, and then changes the response of the library. The library adds a client ID and a call to our store.

## Decision

A data holder serves the registration endpoints (create, read, update and delete) with its own code, and not with the OpenID Provider library. The rules of a registration are business rules of the service, with the same structure as its other business rules ([ADR 0005](0005-service-architecture.md)).

The library gets the registered clients through a port, to authenticate them and to serve each other endpoint. This decision does not change ADR 0001 for those endpoints: the token endpoint, client authentication, PAR, the authorisation flow and JARM stay in the library.

The registration management endpoints verify the `cdr:registration` access token with the library, because the library issues that token.

The two registration workarounds of ADR 0001 are removed, because no code uses them.

## Consequences

- Each registration rule of the service specification is in one place, and a test can check it with no HTTP server and no library.
- A change to the registration code of the library, before or after its version 1.0, does not affect a data holder.
- We maintain the registration endpoints. A defect in the validation of client metadata is ours to find, and a security fix of the library for registration does not reach us.
- The library and our code must agree on what a client is. The port that gives the library its clients maps each registration to the client type of the library, and a registration value that the library does not accept fails at client authentication, not at registration.
- The Verification table of ADR 0001 lists two workarounds that no longer exist. That table is correct for the date and the version that it gives.
- If a later version of the library supports the CDS registration rules, a new ADR can move the endpoints to it.
