# Data recipient, milestone 2: connect a bank research

- Research date: 2026-11-29
- Applies to: CDS 1.36.0, FAPI 1.0 Advanced, JARM, CDR Rules Compilation No. 10 (F2025C00572), Competition and Consumer Act 2010 Part IVD Compilation No. 167 (C2026C00405)
- File and line references are to the repository on the research date.
- Purpose: the facts for milestone 2 of the project "A consumer can connect their bank accounts to a mock data recipient in the lab" (P-CDR-13): "A consumer connects one bank and sees the accounts and the transactions that they shared." The project risks for this milestone: the host of the JWKS and of the redirect URI, the election to delete under rule 4.16, and the growth of the event stream of a consent. The requirements PR resolves them.
- Related notes: [Data recipient: sign-up and sign-in](2026-11-28-data-recipient-sign-up.md), [CDS data holder InfoSec and consent](2026-10-06-cds-data-holder-infosec-and-consent.md), [Data holder: the end of data sharing](2026-10-16-data-holder-end-of-sharing.md), [CDR server certificates](2026-10-27-cdr-server-certificates.md). This note adds the data recipient side.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Analysis**, **Option** and **Recommendation**: our interpretation for the requirements step. They are not facts and not legal advice.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | CDS 1.36.0: Security Profile, Standards (headers, pagination), Non-functional Requirements, Endpoint Version Schedule, CX Standards, `cds_banking.json` (git tag `v1.36.0`, `slate/source/includes/`) | https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/ |
| S2 | FAPI 1.0 Advanced (Part 2), sections 5.2.3, 5.2.3.2 and 8.6 | https://openid.net/specs/openid-financial-api-part-2-1_0.html |
| S3 | JWT Secured Authorization Response Mode for OAuth 2.0 (JARM), section 4.4 | https://openid.net/specs/oauth-v2-jarm.html |
| S4 | CDR Rules: rules 1.8, 1.14, 1.18, 4.3, 4.4, 4.11, 4.12, 4.14, 4.16, 4.17, 4.18, 4.20, 7.4, 7.12, 7.13, 9.3, 9.4, Schedule 2 clause 2.2 | https://www.legislation.gov.au/F2020L00094/latest/text |
| S5 | OAIC, CDR Privacy Safeguard Guidelines, Chapter C: Consent, consultation draft, September 2022 | https://www.oaic.gov.au/__data/assets/pdf_file/0018/21636/Chapter-C-Consent.pdf |
| S6 | Competition and Consumer Act 2010, section 56EO (privacy safeguard 12) | https://www.legislation.gov.au/C2004A00109/latest/text |
| S7 | RFC 6761, section 6.3 | https://www.rfc-editor.org/rfc/rfc6761#section-6.3 |

## 1. What the repository has now

### Facts: the data recipient service

- `services/datarecipient/` has only `design/`. No code, compose service, seed record, certificate entry, JWKS or CDR endpoint exists.
- The planned listeners are internal 8580 and web app 8585 ([DR-0602](../../services/datarecipient/design/deployment.md#dr-0602)). The web app listener has a certificate from the mock public CA with the SANs `localhost` and `sorrel` ([DR-0605](../../services/datarecipient/design/deployment.md#dr-0605)). `WEB_APP_BASE_URI` is `https://localhost:8585` ([DR-0607](../../services/datarecipient/design/deployment.md#dr-0607)).
- [ADR 0025](../adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md) (lines 50 and 52) gives the web app its own listener, so that the CDR endpoints of a data recipient are on a different listener. No port is assigned to that listener yet.
- The session cookie has `SameSite=Lax` for the return from the data holder ([DR-0455](../../services/datarecipient/design/sign-in.md#dr-0455)).

### Facts: the Register

- A software product has redirect URIs, a sector identifier URI, a JWKS URI, a revocation URI, a recipient base URI and scopes ([REG-0212](../../services/register/design/participants.md#reg-0212)). The scopes must contain `openid` and `profile` ([REG-1496](../../services/register/design/participants.md#reg-1496)). Each URI is absolute `https` with a host ([REG-0215](../../services/register/design/participants.md#reg-0215)).
- A legal entity has an accreditation number ([REG-0211](../../services/register/design/participants.md#reg-0211)).
- Get SSA is x-v 4, mTLS, with a token that has `cdr-register:read` ([REG-0400](../../services/register/design/software-statement-assertion.md#reg-0400)). `{industry}` accepts only `all` ([REG-0309](../../services/register/design/register-apis.md#reg-0309)). The SSA expires 10 minutes after issue ([REG-0412](../../services/register/design/software-statement-assertion.md#reg-0412)).
- Get Data Holder Brands is x-v 3, mTLS, paginated, with `updated-since` and no ETag ([REG-0300](../../services/register/design/register-apis.md#reg-0300), [REG-0330](../../services/register/design/register-apis.md#reg-0330), [REG-0335](../../services/register/design/register-apis.md#reg-0335)). The secure APIs give 403 `AdrStatusNotActive` for a software product that is not active ([REG-0323](../../services/register/design/register-apis.md#reg-0323)).
- The Register token endpoint is `POST /idp/connect/token`, mTLS, `private_key_jwt`, `client_credentials` only, with a token of 5 minutes ([REG-0452](../../services/register/design/authorisation-server.md#reg-0452)). The client ID is the Software Product ID ([REG-0453](../../services/register/design/authorisation-server.md#reg-0453)). The Register gets the JWKS of the software product over HTTPS (`pkg/authserver/jwks.go:12-17`), and trusts only the mock public CA (`compose.yaml:86-87`).
- The Register issues a client certificate from a CSR at `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue` ([REG-0632](../../services/register/design/certificate-management.md#reg-0632)). The CN is the software product name and the O is the brand name ([REG-0606](../../services/register/design/certificate-management.md#reg-0606), [REG-0607](../../services/register/design/certificate-management.md#reg-0607)). The issue also registers the certificate ([REG-0610](../../services/register/design/certificate-management.md#reg-0610)). The Register issues no server certificate to a data recipient ([REG-0620](../../services/register/design/certificate-management.md#reg-0620)).
- `cdrlab cert issue -kind client` writes the certificate and key to files or a Secret (`cli/cdrlab/cmd/cdrlab/main.go:45-52`). `lab/certificates.json:123-131` has a `client` entry.
- `cdrlab participants add` posts a load file to `POST /v1/participant-loads`, add-only and all or nothing (`cli/cdrlab/cmd/cdrlab/main.go:387-414`, [REG-1531](../../services/register/design/operator-interface.md#reg-1531)).
- The seed data recipients have hosts on `.example`, for example `https://web.mock-budget.example/jwks` (`services/register/seed/example.json:330`). The project description says `example.com`. The test participants use `https://conformance:8443/jwks` and `https://conformance:8444` (`example.json:749`, `:1035`).

### Facts: the data holder

- The issuer is the `infosecBaseUri` of the brand, for example `https://wattle-bank:8081` ([DH-0500](../../services/dataholder/design/authorisation-server.md#dh-0500)). The discovery document puts `/jwks` and `/authorize` on the issuer, and the other endpoints on the secure base `https://wattle-bank:8083` (`pkg/authserver/dataholder.go:467-481`). A client reads the paths from discovery ([DH-0423](../../services/dataholder/design/api-standards.md#dh-0423)).
- The registration checks the SSA with `REGISTER_JWKS_URL` ([DH-0612](../../services/dataholder/design/registration.md#dh-0612)). It gets the `jwks_uri` of the SSA over HTTPS, with trust in the public CAs only ([DH-0622](../../services/dataholder/design/registration.md#dh-0622), [DH-0623](../../services/dataholder/design/registration.md#dh-0623)). The `redirect_uris` must be in the SSA ([DH-0631](../../services/dataholder/design/registration.md#dh-0631)). The holder keeps one registration for each `software_id` ([DH-0632](../../services/dataholder/design/registration.md#dh-0632)) and gives no `registration_access_token` ([DH-0647](../../services/dataholder/design/registration.md#dh-0647)).
- The request object needs `nonce`, PKCE S256 and `response_mode=jwt`. A `cdr_arrangement_id` claim is refused ([AUTH-0082](../../pkg/authserver/design/authorisation-request.md#auth-0082), [AUTH-0088](../../pkg/authserver/design/authorisation-request.md#auth-0088), [AUTH-0091](../../pkg/authserver/design/authorisation-request.md#auth-0091)). A `request_uri` lives 60 seconds ([AUTH-0066](../../pkg/authserver/design/authorisation-request.md#auth-0066)).
- The JARM response is signed with PS256 and is not encrypted ([AUTH-0108](../../pkg/authserver/design/authorisation-request.md#auth-0108)). The code is valid for 60 seconds, one use.
- The token response has `id_token` and `cdr_arrangement_id`. The access token lives 5 minutes. A refresh token is issued when the sharing duration is more than 0. It expires at the end of sharing and does not rotate ([AUTH-0205](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0205) to [AUTH-0211](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0211)). The ID token is PS256, not encrypted, with `acr` `urn:cds.au:cdr:2` and no personal information ([AUTH-0214](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0214) to [AUTH-0217](../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0217)).
- The `sub` is pairwise for each sector identifier. With no `sector_identifier_uri`, the sector identifier is the host of the redirect URIs ([DH-0509](../../services/dataholder/design/authorisation-server.md#dh-0509), [DH-0511](../../services/dataholder/design/authorisation-server.md#dh-0511)).
- The scopes of the holder are `openid`, `profile`, `cdr:registration`, `common:customer.basic:read`, `bank:accounts.basic:read`, `bank:accounts.detail:read` and `bank:transactions:read` ([DH-0502](../../services/dataholder/design/authorisation-server.md#dh-0502)).
- The banking APIs are Get Accounts v3, Get Account Detail v5, Get Balance v1, Get Bulk Balances v2, Get Balances For Specific Accounts v1, Get Transactions For Account v2 and Get Transaction Detail v3 (`services/dataholder/design/api-standards.md:13-22`). `x-fapi-auth-date` is required ([DH-0434](../../services/dataholder/design/api-standards.md#dh-0434)). Transactions are in descending order with a default window of 90 days ([DH-1083](../../services/dataholder/design/banking-apis.md#dh-1083) to [DH-1085](../../services/dataholder/design/banking-apis.md#dh-1085)). Each ID is a UUIDv8 for each software product ([DH-1040](../../services/dataholder/design/banking-apis.md#dh-1040)).
- The consent page of the holder selects the accounts. Only open accounts can be selected, none is pre-selected, and one is required ([DH-0819](../../services/dataholder/design/authorisation.md#dh-0819) to [DH-0823](../../services/dataholder/design/authorisation.md#dh-0823)). Each confirmation makes a new arrangement ([DH-0832](../../services/dataholder/design/authorisation.md#dh-0832)).
- The holder polls the software product statuses each `REGISTER_STATUS_INTERVAL` (5 seconds in `compose.yaml:46`). It uses only the software product status ([DH-0683](../../services/dataholder/design/registration-management.md#dh-0683), [DH-0694](../../services/dataholder/design/registration-management.md#dh-0694)).
- The holder sends the notice of the end of an arrangement to `<recipient_base_uri>/arrangements/revoke` over TLS, with trust in the public CAs only ([DH-0925](../../services/dataholder/design/end-of-an-arrangement.md#dh-0925), [DH-0927](../../services/dataholder/design/end-of-an-arrangement.md#dh-0927)). The bearer JWT is PS256 with the key of the holder JWKS, and `aud` is the endpoint URL ([DH-0928](../../services/dataholder/design/end-of-an-arrangement.md#dh-0928) to [DH-0931](../../services/dataholder/design/end-of-an-arrangement.md#dh-0931)). The body has only `cdr_arrangement_jwt` ([DH-0932](../../services/dataholder/design/end-of-an-arrangement.md#dh-0932)). A 404 fails with no new attempt ([DH-0935](../../services/dataholder/design/end-of-an-arrangement.md#dh-0935)).

### Facts: code that a client can use

- `pkg/` has no client for the CDR: no mTLS HTTP client, no x-fapi headers, no DCR client, no JARM or ID token validation and no client pagination. The client code is in the conformance suite only.
- The conformance suite acts as a data recipient:
  - `conformance/internal/labtest/recipient.go:71` `ServeJWKS` serves `/jwks`, `/sector-identifier.json` and `/arrangements/revoke` over TLS with a certificate from the public CA.
  - `recipient.go:148` `NewAssertion`, `:178` `RequestToken`.
  - `conformance/internal/labtest/dataholder.go:113` `SoftwareStatement`, `:142` `RegistrationClaims`, `:165` `PostRegistration`.
  - `conformance/dataholder/par_test.go:40-80`, `authorise_test.go:150-200` (JARM) and `consent_tokens_test.go:97-130` are in a test package and are not exported.
- Reusable server parts: `pkg/authserver/signingkey.go:15` `NewSigningKey`, `pkg/ca/csr.go:144` `NewCSR`, `pkg/cds/tls.go:29` `MutualTLSConfig`, `pkg/cds/certificate.go:64` `WatchCertificateFiles`, `pkg/platform/jobs` (River) and `pkg/eventsourcing`.
- `Load` reads the whole stream and applies each event (`pkg/eventsourcing/eventsourcing.go:55-80`, `pkg/eventsourcing/eventstore/eventstore.go:58-75`). The library has no snapshots. ADR 0003 lists "snapshot" as a synonym that the repository does not use for state (line 49).
- [ADR 0010](../adr/0010-background-jobs-with-river.md) (lines 41-42): a projection inserts a job in the append transaction, and the worker appends an outcome event.

### Facts: networks

- Compose has no `networks:` and no `aliases:`. Services reach each other by service name on the default network. The suite runs with `--use-aliases`, so the services reach it at `conformance` (`mise.toml:210-211`).
- The host reaches each service at `localhost:<host port>`. The host port of the public listener of `wattle-bank` is 8181, and the container port is 8081 (`compose.yaml:157-161`).
- Thus the discovery document of `wattle-bank` gives `https://wattle-bank:8081/authorize`, and a browser on the host can resolve neither the name nor that port. No browser test in the repository opens the authorisation page of a bank. The OpenID conformance suite uses its own browser in its container (`compose.yaml:154-155`).
- The umbrella chart makes the base URIs of a brand from its host names. With no host name, the base URI is the Service name and the port ([LAB-0031](../../charts/cdr-lab/DESIGN.md#lab-0031)). With a host name, the port is 443 ([LAB-0032](../../charts/cdr-lab/DESIGN.md#lab-0032)), and each pod must reach each external host name ([LAB-0037](../../charts/cdr-lab/DESIGN.md#lab-0037)). The chart makes the Register record of a bank from the same host names ([LAB-0018](../../charts/cdr-lab/DESIGN.md#lab-0018)).
- The release of the cluster job in CI has no routes and no external host names ([LAB-0114](../../charts/cdr-lab/DESIGN.md#lab-0114)).

## 2. The data recipient side of the CDS

### Facts: participant set-up

- The Register issues an SSA only to an active data recipient for an active software product. The SSA is signed with PS256. The data holder verifies it against the Register JWKS (S1, Client Registration: Software Statement Assertion, Registration Validation).
- The SSA has `org_id`, `org_name`, `client_name`, `redirect_uris`, `jwks_uri`, `recipient_base_uri`, `software_roles` and `scope`, with others. `openid` and `profile` are in the SSA scope (S1, SSA Definition).
- Get Software Statement Assertion is `GET /{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa`, version 4 since CDS 1.35.0 (S1, Endpoint Version Schedule: Register).
- The client certificate from the Register CA secures the calls to the Register and to the data holders. A data recipient needs no CDR server certificate (S1, Certificate Management: Issued by the Register CA for Data Recipients).
- The JWKS endpoint of the software product is TLS, with no client authentication. Its key set must not have two keys with the same `kid` (S1, JSON Web Key Set endpoint).
- The data recipient endpoints that are TLS must not use mTLS (S1, Transport Security: Use of MTLS). The CDS does not say which CA issues their TLS server certificate. The data recipient endpoints are the JWKS endpoint and `<RecipientBaseUri>/arrangements/revoke` (S1, Endpoint Version Schedule: Data Recipients).
- Data holders and data recipients "MUST cache all other participant data they use and periodically update this cache using a slow poll". The recommended period for Get Data Holder Brands is 6 hours (S1, Participant Statuses: Cache update periods).

### Facts: Dynamic Client Registration

- The registration request is a JWT that the software product signs. `iss` is the Software Product ID and `aud` is the issuer of the data holder (S1, Registration Request using JWT).
- Required claims include `token_endpoint_auth_method` (`private_key_jwt` only), `token_endpoint_auth_signing_alg`, `grant_types`, `response_types`, `id_token_signed_response_alg` and `software_statement`. `authorization_signed_response_alg` is required for `response_type` `code`. `redirect_uris` is optional, and must match or be a subset of the SSA value (S1, Registration Request using JWT).
- From 12 May 2025, a data recipient uses only `response_type` `code` (S1, Registration Request using JWT). `id_token_encrypted_response_alg` and `_enc` "SHALL NOT be used for Authorization Code Flow" (same).
- The data holder verifies the request JWT with a key from the `jwks_uri` of the SSA (S1, Registration Validation). Thus the data holder must reach the `jwks_uri` at each registration.
- GET, PUT and DELETE `/register/{clientID}` need a client credentials token with the scope `cdr:registration`. A PUT is the only update. It is used when the registration details on the Register change, or when a new SSA version is released (S1, Registration Management, PUT Operations). Access tokens for CDR data "SHOULD NOT" be used for registration management (same).

### Facts: client authentication

- The data holder and the Register authenticate the software product with `private_key_jwt`. `iss` and `sub` are the client ID, and the claims `aud`, `jti` and `exp` are required. Each assertion is single use. The `client_assertion_type` is `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` (S1, Private Key JWT Client Authentication).
- At the data holder, the client ID is the ID that the data holder issued at registration. `grant_type` is `authorization_code`, `client_credentials` or `refresh_token` (S1, Data Recipients calling Data Holders).
- At the Register, the client ID is the Software Product ID or the Data Recipient Brand ID (S1, Data Recipients calling the CDR Register).
- The software product authenticates a data holder that calls it with the self-signed JWT method. The JWT is in the `Authorization` header as a bearer token. `aud` must be the resource path of the endpoint, and the endpoint must also accept the `<RecipientBaseUri>`. Each JWT is single use (S1, Self-signed JWT Client Authentication).
- The CDR Arrangement Revocation endpoint of the software product accepts the "CDR Arrangement JWT" method and must not refuse the form parameter. When both are present, they must be identical. A `cdr_arrangement_id` that does not belong to the caller is refused with 422 (S1, CDR Arrangement Revocation End Point).
- Each back-channel call between a software product and a data holder uses mTLS, with certificates of the CDR CA only (S1, Transport Security: Use of MTLS). Tokens are bound to the client certificate (S1, Holder of Key Mechanism; S2, 5.2.3 item 1).

### Facts: the authorisation

- The software product "SHALL only use Authorization Code Flow" from 12 May 2025. It "MUST use [RFC9126] (PAR) with [PKCE]" and `S256` (S1, Authentication Flows: Data Recipient Software Products).
- The request object is signed with PS256 or ES256 (S2, 8.6). It has `nbf`, and `exp` no more than 60 minutes after `nbf` (S1, Request Object: Data Recipient Software Products). `aud` is the issuer of the data holder (S2, 5.2.3 item 10). All parameters are in the signed request object (S2, 5.2.3 item 8).
- Each `request_uri` is used once. A data holder makes it expire between 10 and 90 seconds (S1, Request Object Submission).
- `sharing_duration` is a claim in the request object, in seconds. A value of more than one year is one year (31,536,000 seconds). A value of 24 hours or less is a one-time collection. With zero or no value, the data holder issues no refresh token. A refresh token that is issued for a one-time collection must be revoked after the collection (S1, Requesting Sharing Duration).
- A new consent provides `cdr_arrangement_id` in the request object only to amend an existing arrangement (S1, Specifying an existing arrangement). The software product gets the `cdr_arrangement_id` from the token response or from token introspection (S1, Identifiers: CDR Arrangement ID).
- The `scope` has `openid`. `profile` is optional (S1, Scopes and Claims). The non-normative request object example asks for the `acr` claim with `urn:cds.au:cdr:3` as essential (S1, Request Object).
- The software product "MUST support [JARM]". It requests a signing algorithm from the `authorization_signing_alg_values_supported` of the data holder. With no `authorization_signed_response_alg`, the default is PS256. Encryption is optional and is "not recommended" (S1, Authorization Code Flow).
- The client verifies the authorisation response as JARM section 4.4 specifies (S2, 5.2.3.2): the `iss`, the `aud`, the `exp` and the signature with the JWKS of the data holder (S3, section 4.4).
- The authorisation code is not reused (S1, Authentication Flows).
- The software product "SHOULD record the following information each time an authorisation flow is executed: username (consumer’s ID at the Data Recipient Software Product), timestamp, IP, consent scopes and duration" (S1, Authentication Flows).
- The redirect URIs for an app and for a web site are separate (S1, Redirect to App). The mock has a web app only.

### Facts: tokens

- The ID token from the token endpoint is signed, with no encryption in the authorisation code flow. It has no personal information claims (S1, Tokens: ID Token).
- The `sub` is a pairwise pseudonymous identifier for the consumer at the data holder (S1, Claims).
- An access token expires 2 to 10 minutes after the data holder issues it. `expires_in` gives the lifetime (S1, Access Token, Token Expiry).
- The data holder does not rotate refresh tokens. The `exp` of a refresh token is equal to the sharing duration (S1, Refresh Token).
- The software product gets the expiry of the sharing arrangement from token introspection, in `exp` (S1, Requesting Sharing Duration).

### Facts: resource calls

- `x-v` is mandatory. `x-min-v` is optional. The data holder gives the highest version between the two (S1, HTTP Headers).
- `x-fapi-auth-date` is "Required for all resource calls (customer present and unattended)". It is "the time when the customer last logged in to the Data Recipient Software Product" (S1, HTTP Headers).
- `x-fapi-customer-ip-address` and `x-cds-client-headers` are present only when the customer is logged in. The presence of `x-fapi-customer-ip-address` makes the call customer present (S1, HTTP Headers; S1, NFRs: Definitions).
- `x-fapi-interaction-id` is optional on a request. The data holder plays it back, or makes a new one (S1, HTTP Headers).
- The default page size is 25 and the maximum is 1000. The response has `links` and `meta` with `totalRecords` and `totalPages` (S1, Pagination).
- With no `oldest-time`, Get Transactions For Account gives 90 days before `newest-time` (S1, `cds_banking.json`, Get Transactions For Account).
- The versions for CDS 1.36.0 from 9 November 2026: Get Accounts v3, Get Account Detail v5, Get Bulk Balances v2, Get Transactions For Account v2, Get Transaction Detail v3. Get Accounts v2, Get Account Detail v4, Get Bulk Balances v1, Get Transactions For Account v1 and Get Transaction Detail v2 retire on 7 December 2026 (S1, Endpoint Version Schedule: Banking).

### Facts: non-functional rules for a data recipient

- A session is "the life span of a unique Access Token" (S1, NFRs: Definitions).
- Unattended traffic in low traffic periods has a threshold of "20 sessions per day, per customer, per Data Recipient Software Product", "100 total calls per session" and "5TPS per session". In high traffic periods (6am to midnight) the data holder gives best effort only (S1, Traffic Thresholds; NFRs: Definitions).
- Data recipients "should schedule unattended calls to avoid high traffic periods" and avoid short bursts (S1, Data Recipient Requirements).
- Customer present traffic has a threshold of 10 TPS for each customer and 50 TPS for each software product (S1, Traffic Thresholds).

### Facts: participant statuses

- The CDS tells data holders, not data recipients, to react to the statuses of data recipients and software products within 5 minutes (S1, Participant Statuses).
- A data recipient whose software product is not active cannot get an SSA (S1, Software Statement Assertion). The CDS gives no rule for how a data recipient reacts to the status of a data holder brand.

### Recommendations

- Request the scopes `openid`, `bank:accounts.basic:read`, `bank:transactions:read`, and `bank:accounts.detail:read` only if the app shows the account number or the features. Do not request `profile`, because the member details already give the name, and rule 1.8 limits the collection to what is reasonably needed.
- Send `x-fapi-auth-date` from the time of the last sign-in of the member. Send `x-fapi-customer-ip-address` and `x-cds-client-headers` only on a collection that a request of the member starts.
- Send a new `x-fapi-interaction-id` on each call, and record it in the span of the call. The acceptance criterion of the project needs it in the trace.
- Use the current versions only, with `x-v` and no `x-min-v`. A data holder that a team brings to the lab can be on a different version, so a 406 is an error that the app shows, not a fallback.
- Validate the ID token from the token endpoint: the signature, `iss`, `aud`, `exp`, `nonce` and `acr`. Keep the `sub` only as an identifier of the consumer at that data holder.

## 3. Consent rules for a collection consent and a use consent

### Facts

- Rule 4.3(2): the accredited person may ask the consumer to give "(a) a collection consent for the accredited person to collect their CDR data from the CDR participant; and (b) a use consent for the accredited person to use that CDR data" (S4). Note 1 to rule 4.3(2A): the requests "may be bundled together".
- Rule 4.3(4): "The request ceases to be ***valid*** if the collection consent is withdrawn." (S4)
- Rule 4.4(3): the accredited person makes a consumer data request "(a) using the data holder’s accredited person request service; and (b) in accordance with the data standards." Note 1 to rule 4.4(2): an accredited person "might need to make regular consumer data requests over a period of time". (S4)
- Rule 4.11(1), for a collection consent: (a) the consumer selects or agrees to "the particular types of CDR data"; (aa) for a use consent, "the specific uses of collected data"; (b) the period, which is "(iii) a single occasion; or (iv) a specified period of time"; (c) "seek the CDR consumer’s express consent"; (e) "allow the CDR consumer to make an election in relation to deletion of redundant data in accordance with rule 4.16." Note 1: "An accredited person cannot infer consent, or seek to rely on an implied consent." (S4)
- Rule 4.11(3) requires this information when the consumer gives the consents in scope (S4):
  - (a) "its name";
  - (b) "its accreditation number";
  - (c) how the collection and the use comply with the data minimisation principle, with, for a collection consent, "an explanation of why that collection is reasonably needed, and relates to a time period that is no longer than is reasonably needed";
  - (f) the facts about an OSP, if the data goes to an OSP;
  - (g) "a statement that, at any time, the consent can be withdrawn";
  - (h) "(i) a statement, in accordance with rule 4.17, regarding the accredited person’s intended treatment of redundant data; (ii) a statement outlining the CDR consumer’s right to elect that their redundant data be deleted; (iii) instructions for how the election can be made".
- Rule 4.12(1): the period is not "more than 12 months". Rule 4.12(2): no consent unless the collection and the use "would comply with the data minimisation principle". (S4)
- Rule 1.8(1): a collection complies if the accredited person does not seek "(a) more CDR data than is reasonably needed; or (b) CDR data that relates to a longer time period than is reasonably needed". (S4)
- Rule 4.14(1): a consent expires at the earliest of the events of the rule, for example "(d) at the end of the period the CDR consumer consented to in accordance with rule 4.11". Rule 4.14(3): when a data holder notifies the withdrawal of an authorisation under rule 4.26A, "the collection consent expires when the accredited person receives that notification." (S4)
- Rule 4.16(1): the consumer may elect "that the collected data, and any CDR data directly or indirectly derived from it, be deleted when it becomes redundant data". Rule 4.16(2): by writing or on the consumer dashboard. (S4)
- Rule 4.16(3): "This rule does not apply if the accredited person: (a) has a general policy of deleting redundant data; and (b) when seeking the consent, informs the CDR consumer that their CDR data will be deleted when it becomes redundant data." (S4)
- Rule 4.17(1): the accredited person states whether its general policy is "(a) deleting the redundant data; or (b) de-identifying the redundant data; or (c) deciding, when the CDR data becomes redundant data, whether to delete it or de-identify it." Rule 4.17(2) adds statements only for (b) and (c). (S4)
- Rule 4.18: a CDR receipt "as soon as practicable after the CDR consumer: (a) gives the accredited person a collection consent". It is a civil penalty provision. P-CDR-14 delivers it. (S4)
- Rule 4.20(1)(b): the notification applies when "90 days have elapsed" since the consent, an amendment, the last use of the dashboard or the last notification. P-CDR-14 delivers it. (S4)
- Rule 1.14(3)(e)(ii): the consumer dashboard shows, for a consent over a period, "how often data has been, and is expected to be, collected or disclosed over that period". Rule 7.4(1): the dashboard shows what data was collected, when, and from which participant, "as soon as practicable". P-CDR-14 delivers the dashboard. (S4)
- Rule 9.3(2): the records of an accredited data recipient include "(a) all consents", "(b) amendments to or withdrawals of consents", "(c) notifications of withdrawals of authorisations received from data holders", "(e) collections of CDR data under these rules", "(f) elections to delete and withdrawals of those elections", "(g) the use of CDR data" and "(h) the processes by which the accredited data recipient asks CDR consumers for their consent ... including a video of each process". Rule 9.3(3): "Each record ... must include the date and time when the record was made and, if applicable, the date and time when the event described by the record occurred." (S4)
- Rule 9.3(2)(e) says "collections", not "each collection". No rule gives the granularity of a record of collection.
- CDS CX standards (S1):
  - "Consent: Redirection": "Data recipients MUST notify consumers of redirection prior to authentication."
  - "Data Language Standards: Language to be used": data recipients "MUST use data language standards to describe data clusters and permissions in consumer-facing interactions".
  - "Detailed scope requests": when a consent has the detailed cluster and not the basic cluster, the data recipient uses the merged language that the table marks with ‡.
  - "Authentication: Passwords": the statement that third parties do not need consumer passwords.
  - The accessibility standards apply to "all aspects of the Consent Model".
- The banking data language (S1, Data Language Standards: Banking and Non-Bank Lending Language):

  | Data cluster | Permissions | Scope |
  |---|---|---|
  | Account name, type and balance | Name of account; Type of account; Account balance | `bank:accounts.basic:read` |
  | Account numbers and features | Account number; Interest rates; Fees; Discounts; Account terms; Account mail address | `bank:accounts.detail:read` |
  | Account balance and details (‡, detailed without basic) | The permissions of both rows above | `bank:accounts.detail:read` |
  | Transaction details | Incoming and outgoing transactions; Amounts; Dates; Descriptions of transactions; Who you have sent money to and received money from | `bank:transactions:read` |

- The CX authorisation standards make account selection a step at the data holder, not at the data recipient (S1, Authorisation Standards: Account selection).
- An OAIC consultation draft of the Privacy Safeguard Guidelines, Chapter C (September 2022), says: "Where the accredited person will or may de-identify redundant data, the accredited person must also: allow the consumer to elect for their redundant data to be deleted". It also says "Where an accredited person does not have a general policy of deleting redundant data ... [it] must allow consumers to elect to have their redundant data deleted prior to the final withdrawal step" (S5). The final text of the current version was not checked, because the OAIC site refused the automated request.

### Analysis: the election to delete (project risk)

- Rule 4.11(1)(e) and rule 4.11(3)(h)(ii) and (iii) refer to the election "in accordance with rule 4.16". Rule 4.16(3) says that rule 4.16 does not apply when the policy is to delete and the consent states it. In our interpretation, the brand then has no election to offer, and no instructions to give. The OAIC draft agrees: it ties the election to a policy that "will or may de-identify".
- The consent screen still must give the statement of rule 4.17(1)(a), which is also the information of rule 4.16(3)(b).
- Rule 9.3(2)(f) records elections. With no election, the brand records none. The record of the consent (rule 9.3(2)(a)) holds the statement that the consumer saw.
- Rule 9.4(2)(f)(ii) asks each report for "the proportion of CDR consumers who ... had exercised the election to delete". With no election, the proportion is not applicable. The lab makes no reports.

### Recommendations

- Show no election on the consent screen. Show the statement of rule 4.17(1)(a): the brand deletes the CDR data when it becomes redundant. Record a judgement with rule 4.16(3) as its basis. P-CDR-14 then also has no election on the dashboard (rule 1.14(1)(e)), with the same judgement.
- The consent screen has, in this order: the name and the accreditation number of the brand, the data clusters with the CDS language, the purpose (the use) and why the data and the period are needed, the duration, the statement on redundant data, the statement that the consumer can withdraw at any time, the password statement, and one express action, for example "I agree". Then a redirection notice that names the bank, before the browser goes to the bank.
- The consumer selects the clusters and the duration (rule 4.11(1)(a)(i) and (b)(i)). "Account name, type and balance" is required, because the app cannot show an account without it. "Transaction details" is optional. The duration is a list of choices up to 12 months, with no pre-selected value. The project acceptance criterion asks for this choice.
- The accreditation number is the number of the legal entity on the Register ([REG-0211](../../services/register/design/participants.md#reg-0211)). The app gets it from Get Data Recipients, or from the configuration. It must not be in the brand directory as a second copy.
- The record of each consent holds the clusters and the scopes, the duration, the uses, the version of the statement on redundant data, the time, and the member ID, as CDS Authentication Flows recommends. The client IP address goes into the record table, not into an event, because it is personal information.
- A video of the consent process (rule 9.3(2)(h)) is a record of the operator, not a function of the software. The specification needs no requirement for it.

## 4. The lab topology

### Facts

- Each fetch of a data recipient URI comes from a service, and the browser follows only the redirect URI:

  | URI | Who uses it | From where |
  |---|---|---|
  | `redirect_uris` | The browser of the consumer. The data holder only compares the string. | The host |
  | `jwks_uri` | The data holder at each registration, the Register at each token request | The lab network |
  | `recipient_base_uri` | The data holder, for the notice of the end of an arrangement | The lab network |
  | `logoUri` of a data holder brand | The browser, if the app links to it | The host |
  | `authorization_endpoint` of the data holder | The browser | The host |

- The SSA gives each of these as a separate field (S1, SSA Definition). The CDS has no rule that the `jwks_uri`, the `recipient_base_uri` and the `redirect_uris` share a host.
- The data holder and the Register trust only the public CAs for the JWKS and the notice ([DH-0623](../../services/dataholder/design/registration.md#dh-0623), [DH-0927](../../services/dataholder/design/end-of-an-arrangement.md#dh-0927), `compose.yaml:86-87`). Thus the CDR listener of the data recipient needs a certificate from the mock public CA.
- RFC 6761 section 6.3 reserves `.localhost` for the loopback address (S7). Browser support for subdomains of `localhost` and their resolution in a container were not checked.
- The bank logos in Get Data Holder Brands are on the issuer host, for example `https://wattle-bank:8081/brand/logo.svg` (`services/register/seed/example.json:6`).

### Analysis

- For the data recipient, a split is enough: a redirect URI on `localhost:8585`, and a JWKS and a recipient base URI on the service name, for example `https://sorrel:8581`. The data holder needs no change for these fields.
- The browser also needs the authorisation page of the bank, and the bank gives it on a host that the browser cannot reach. This blocks the milestone in Compose, also with a perfect data recipient. It is a change in the data holder or in the lab, not only in the data recipient.
- In a cluster with routes, the external host names work from the browser and from the pods ([LAB-0037](../../charts/cdr-lab/DESIGN.md#lab-0037)). One host name then serves both, and the problem is only in Compose.

### Options for Compose: the authorisation page of the bank

| Option | What changes | Cost |
|---|---|---|
| A. Host names that resolve on both sides, for example `wattle-bank.localhost`, with the same port in the container and on the host | Each listener of each service, the seed URIs, the certificate SANs, the Host allowlists | A change of each port and URI in the lab. A spike must prove that browsers and the Go resolver in a container resolve the names. |
| B. A data holder setting for the base URI of its browser pages, for example `https://localhost:8181`, used for `authorization_endpoint` and the consent pages | The data holder discovery and its settings. The issuer stays. | The browser of the OpenID conformance suite is on the lab network and cannot follow `localhost`. That profile needs a different value. |
| C. Entries in `/etc/hosts` on the host, and unique container ports | The host of each person | A manual step with root on each host, outside `cdrlab`. |
| D. A browser on the lab network, for example a container with noVNC | A new lab service | A poor experience for a person, and the tests do not change. |

### Recommendations

- Data recipient: two listeners with TLS. The web app listener has the redirect URI, for example `https://localhost:8585/connect/callback` in Compose. A new public listener (port 8581, after the ADR 0025 judgement of [DR-0602](../../services/datarecipient/design/deployment.md#dr-0602)) has `/jwks` and the recipient base URI. Its certificate is from the mock public CA, with the SAN `sorrel`.
- The Register record of `sorrel` goes into the seed, with these URIs, and its client certificate goes into `lab/certificates.json`. The umbrella chart makes the record from the host names of the two listeners, as it does for a bank ([LAB-0018](../../charts/cdr-lab/DESIGN.md#lab-0018)).
- The bank authorisation page: a time-boxed spike of option A first, because it gives one host name for each listener in Compose and in a cluster. Option B if the spike fails. Both change the data holder, so the requirements PR needs a decision of the repository owner, and the data holder specification changes in the same milestone.
- Bank logos: the data recipient gets each logo from the lab network and serves it from its own origin, with a cache. This also keeps the Content Security Policy of the web app at `'self'`.
- Milestone 3: two brands with redirect URIs on `localhost` have the same sector identifier, so a bank gives them the same `sub` ([DH-0511](../../services/dataholder/design/authorisation-server.md#dh-0511)). A `sector_identifier_uri` for each brand, or a host name for each brand, keeps them separate.

## 5. The growth of the event stream of a consent

### Facts

- Rule 9.3(2)(e) requires records of "collections of CDR data", and rule 9.3(3) requires the time of each record and of its event (S4).
- A session is the life of one access token. Unattended traffic has a threshold of 20 sessions each day for each customer and software product, and 100 calls in each session (S1, Traffic Thresholds). A scheduled collection each hour needs 24 sessions each day, so it is above the threshold.
- Unattended calls should be in the low traffic period, from midnight to 6am (S1, Data Recipient Requirements; NFRs: Definitions).
- The consumer dashboard shows "how often data has been, and is expected to be, collected" (S4, rule 1.14(3)(e)(ii)). P-CDR-14 delivers it.
- `Load` reads the whole stream (section 1). [ADR 0021](../adr/0021-record-one-business-fact-in-each-event-type.md) (line 33): one business fact for each event type.
- The project constraint: the events hold no personal information and no CDR data.

### Options

| Option | Events in a 12-month consent | Change |
|---|---|---|
| A. One stream for each consent, with one event for each collection run, and a schedule of one run each day in the low traffic period | Approximately 365, and more for each run that the member starts | None |
| B. A stream for each collection, and the consent stream only for the lifecycle | Consent: a few. Collection: 2 or 3 each | A second aggregate. The worker reads the consent state before it starts, so a withdrawal can occur during a run. |
| C. Snapshots in `pkg/eventsourcing` | Any | A new library function, and a change to ADR 0003 |

### Recommendations

- Option A. One event for each collection run, not for each API call. Two event types, for example `DataCollected` and `CollectionFailed` (ADR 0021). The payload has the data clusters, the time window, the counts and the `x-fapi-interaction-id` values, and no account IDs, amounts or descriptions.
- One scheduled run each day in the low traffic period, with a random offset for each consent, so that runs do not come in bursts. A run that the member starts is customer present, with a minimum interval, for example 15 minutes.
- A test that loads a stream of 1,000 events proves that the load stays fast. Option B is the next step if a test shows a problem.
- The first run collects the default window of Get Transactions For Account (90 days). Later runs collect from the newest transaction time of the last run, with an overlap, and remove duplicates by the transaction ID, which does not change for the software product ([DH-1040](../../services/dataholder/design/banking-apis.md#dh-1040)).

## 6. Storage, deletion and telemetry

### Facts

- Schedule 2, clause 2.2, control requirement (2)(a): "Encryption methods are utilised to secure CDR data at rest by encrypting file systems, end-user devices, portable storage media and backup media. Cryptographic keys are securely stored, backed-up and retained." (S4)
- Section 56EO(2) of the Act: when the CDR entity "no longer needs any of that CDR data" (the redundant data), it "must take the steps specified in the consumer data rules to destroy the redundant data or to ensure that the redundant data is de-identified" (S6).
- Rule 7.13: with no de-identification, "The step is to apply the CDR data deletion process to the redundant data." Rule 1.18: "(a) delete, to the extent reasonably practicable, that CDR data and any copies of that CDR data; (b) make a record to evidence the deletion" (S4).
- Rule 4.14(1)(d): the consent expires at the end of its period (S4). The CDS requires the data recipient to call the arrangement revocation endpoint of the data holder when a consent is withdrawn or expires, except at its natural expiry (S1, CDR Arrangement Revocation End Point).
- The refresh token is bound to the client certificate (S1, Holder of Key Mechanism). It is valid for the full sharing duration and does not rotate (S1, Refresh Token).
- [ADR 0006](../adr/0006-telemetry.md) (lines 37-40): no CDR data and no personal information in telemetry, also mocked data. Record route patterns, not raw URLs, because paths can contain account IDs.
- `docs/production-gaps.md` has entries for signing keys in the database with no encryption, for no TLS and no encryption at rest in PostgreSQL, and for the codes and refresh tokens of the data holder in plain text. It has no entry for the tokens or the collected data of a data recipient.

### Recommendations

- Tokens: the refresh token, the access token and the ID token go into a table, not into events. Plain text is a lab shortcut with a new production gap entry, because the certificate binding limits the use of a stolen token. An encryption with a key from a setting is the alternative, with more settings and no benefit for invented data.
- Collected data: accounts, balances and transactions go into tables with the consent ID as a key, so that one statement deletes them. Make the existing entry for encryption at rest name the CDR data of a data recipient.
- Deletion: when the consent expires, a job deletes the collected data and appends an event that records the deletion (rule 1.18(b)). The job revokes the refresh token at the holder when the consent ends before its natural expiry. In milestone 2, only the natural expiry occurs.
- Telemetry: spans for each call to the Register and to the banks, with the route pattern, the status, the `x-v` and the `x-fapi-interaction-id`. No account IDs, `sub`, tokens, `cdr_arrangement_id` values in URLs, amounts or names. The `x-fapi-interaction-id` is a random UUID and is not CDR data.
- The notice of the end of an arrangement: P-CDR-14 delivers the endpoint. Until then, a notice gets a 404 and fails with no new attempt ([DH-0935](../../services/dataholder/design/end-of-an-arrangement.md#dh-0935)). The next scheduled run then gets an error from the token endpoint. The run records `CollectionFailed`, and the app tells the member to connect again.

## 7. Risks and decisions for the requirements PR

| Item | Options | Recommendation |
|---|---|---|
| The browser on the host cannot open the authorisation page of a bank in Compose, because discovery gives `https://wattle-bank:8081/authorize` | A. Names that resolve on both sides, with equal ports. B. A browser base URI setting in the data holder. C. `/etc/hosts`. D. A browser on the lab network. | A decision of the repository owner. A time-boxed spike of A, then B. The data holder specification changes in the same milestone. |
| The host of the JWKS and of the redirect URI (project risk) | One host name for both. Separate URIs for each side. | Separate URIs in Compose: the redirect URI on `localhost:8585`, the JWKS and the recipient base URI on `https://sorrel:8581`. One external host name for each listener in a cluster. |
| A listener for the CDR endpoints of the data recipient | The web app listener. A new public listener. | A new public listener with TLS from the mock public CA, after the judgement of ADR 0025. |
| The election to delete (project risk) | Offer and record it. Offer nothing, because rule 4.16(3) applies. | No election. The consent states that the brand deletes redundant data (rule 4.17(1)(a)). A judgement on rule 4.16(3), with the OAIC draft as support. P-CDR-14 uses the same judgement for the dashboard. |
| The growth of the event stream (project risk) | One stream with a daily schedule. A stream for each collection. Snapshots. | One stream, one event for each run, one scheduled run each day in the low traffic period. A test with 1,000 events. An hourly schedule is above the CDS threshold of 20 unattended sessions each day. |
| The scopes and the clusters | Basic accounts, transactions, detailed accounts, `profile`, customer | `bank:accounts.basic:read` (required) and `bank:transactions:read` (optional). No `profile` and no customer scope, because of rule 1.8. |
| The duration | Free entry. A list of choices. | A list of choices up to 12 months, with no pre-selected value. |
| The accreditation number | From the Register. From the brand directory. | From the Register legal entity. |
| Client code for the CDR | Copy from the conformance tests. A new `pkg/` library. | A `pkg/` library for the client parts that the suite and the data recipient share (assertions, DCR, PAR, JARM, ID token, mTLS client). Decide its scope in the requirements. |
| Tokens and collected data at rest | Plain text with a production gap. Encryption with a key from a setting. | Plain text with a new production gap entry. Collected data in tables with the consent ID as a key. |
| The notice from a bank before P-CDR-14 | No endpoint (404). A minimal endpoint now. | No endpoint. A failed run tells the member to connect again. |
| Bank logos | A link to the bank host. A copy from the data recipient origin. | A copy from the data recipient origin, because the host browser cannot reach the bank host in Compose. |

## Main findings

- In Compose, a browser on the host cannot open the authorisation page of a lab bank, because discovery gives the service name and the container port. This blocks the milestone, and the fix is in the data holder or in the lab names.
- The SSA gives the redirect URI, the JWKS URI and the recipient base URI as separate fields. The browser uses only the redirect URI, so a data recipient can put the other two on the service name.
- The data holder and the Register trust only the public CAs for the JWKS and the notice. The CDR listener of the data recipient needs a certificate from the mock public CA.
- Rule 4.16(3) removes the election to delete when the brand has a general policy to delete and says so. An OAIC consultation draft ties the election to de-identification.
- The CDS threshold for unattended traffic is 20 sessions each day for each customer. An hourly collection is above it, and a daily run gives approximately 365 events in a 12-month consent.
- `pkg/eventsourcing` loads the whole stream and has no snapshots.
- `pkg/` has no CDR client code. The client parts are in the conformance tests and are not exported.
- `x-fapi-auth-date` is required on each resource call, also unattended. Only a customer present call has `x-fapi-customer-ip-address`.
- The seed data recipients are on `.example`, not `example.com`.
