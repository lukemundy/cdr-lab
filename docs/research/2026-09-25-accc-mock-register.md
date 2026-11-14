# Research: ACCC CDR Mock Register (behaviour, test oracle use)

- Research date: 2026-09-25
- Purpose: Give the data to (a) write requirements for a new, independent mock CDR Register in Go, and (b) use the ACCC mock as a test oracle.
- Check the sources again before you depend on a fact; the repos change.

## Versions examined

| Repo | Version | Commit | Date |
|---|---|---|---|
| ConsumerDataRight/mock-register | v2.2.8 (latest release) | `b1bd1157befdf38d27f535b14b6fbc0a2007ac22` (main) | release 2026-07-15 |
| ConsumerDataRight/mock-solution-test-automation | v3.1.0 | `ce77f50da4eb36577ff4d44a583751fd63d0935b` (main) | 2026-07-08 |
| ConsumerDataStandardsAustralia/standards | v1.36.0 (tag; `master` also reports 1.36.0) | `master` on 2026-09-25 | - |
| mock-data-holder | v4.0.5 (not cloned; release list and compose files only) | - | 2026-07-15 |
| mock-data-recipient | v3.0.5 (not cloned; release list only) | - | 2026-07-15 |
| authorisation-server | v3.1.0 (not cloned; release list only) | - | 2026-07-15 |

All code links in this file point to this mock-register commit:
https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22

## Primary sources

- Mock Register repo: https://github.com/ConsumerDataRight/mock-register
- Mock Register README: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md
- Mock Register CHANGELOG: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CHANGELOG.md
- Certificate docs: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CertificateManagement/README.md
- Test guide: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Help/testing/HELP.md
- Container guide: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Help/container/HELP.md
- Polyglot notebook: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Help/notebooks/mock-register.ipynb
- Issues: https://github.com/ConsumerDataRight/mock-register/issues?q=is%3Aissue
- Docker Hub: https://hub.docker.com/r/consumerdataright/mock-register
- Shared test automation: https://github.com/ConsumerDataRight/mock-solution-test-automation
- CDS Register OpenAPI (1.36.0): https://github.com/ConsumerDataStandardsAustralia/standards/blob/master/docs/includes/swagger/cds_register.json
- CDS 1.36.0 archive: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/
- Org repo list: https://github.com/ConsumerDataRight

---

## 1. Licences and target versions

### 1.1 Licences

| Repo | Licence | Note |
|---|---|---|
| mock-register | MIT, "Copyright (c) 2021 Commonwealth of Australia" | [LICENSE](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/LICENSE) |
| mock-solution-test-automation | MIT, "Copyright (c) 2023 Commonwealth of Australia" | https://github.com/ConsumerDataRight/mock-solution-test-automation/blob/main/LICENSE |
| mock-data-holder | MIT | GitHub licence metadata |
| mock-data-recipient | MIT | GitHub licence metadata |
| authorisation-server | MIT | GitHub licence metadata |
| mock-data-holder-energy | MIT | Archived (read-only). Last push 2024-04-03. |
| sandbox | No licence | It has only a README and issues. It is not code. |

- The repo `mock-solution-containers` does not exist in the ConsumerDataRight org (GitHub API returns 404 on 2026-09-25).
- The docker compose files use the Microsoft SQL Server image. You must accept the Microsoft EULA to use it ([README.md L182](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md)).
- MIT lets you read and copy the behaviour. The new Go implementation is independent, so the licence has no effect on it. It has an effect only if you copy code, certificates or seed data. If you copy them, keep the MIT notice.

### 1.2 Target standards

- The README states alignment to Consumer Data Standards (CDS) v1.36.0 ([README.md L1-L15](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md)).
- The tech stack is C#, .NET 8, Ocelot (gateways), EF Core with SQL Server ([README.md L227-L233](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md)).

### 1.3 Register API versions: CDS 1.36.0 against the mock

Source for CDS columns: `cds_register.json` on `standards` master, `info.version` = `1.36.0`. Source for mock columns: [CdrApiOptions.cs L15-L23](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Models/CdrApiOptions.cs) and the controllers.

| Endpoint (CDS operationId) | Path | CDS 1.36 x-v | CDS industry enum | Mock x-v min..max | Mock x-v mandatory | Mock industry allowed |
|---|---|---|---|---|---|---|
| getRegisterOpenIdProviderConfig | `/idp/.well-known/openid-configuration` | none | - | none | no | - |
| getRegisterJwks (SSA JWKS) | `/cdr-register/v1/jwks` | none | - | none | no | - |
| getDataHolderBrands | `/cdr-register/v1/{industry}/data-holders/brands` | 3 | `IndustryEnumV2` (banking, energy, non-bank-lending, telco, all) | 2..3 | yes | v2: banking, energy, telco, all. v3: all values. |
| getDataHolderBrandsSummary | `/cdr-register/v1/{industry}/data-holders/brands/summary` | 2 | `IndustryEnumV2` | **NOT IMPLEMENTED** | - | - |
| getSoftwareStatementAssertion | `/cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa` | 4 | `IndustryEnumRecipients` (all) | 3..4 | yes | v3: banking, energy, telco, all. v4: all only. |
| getDataHolderStatuses | `/cdr-register/v1/{industry}/data-holders/status` | 2 | `IndustryEnumV2` | 1..2 | **no** (CDS says required) | v1: all, banking, energy, telco. v2: all values. |
| getSoftwareProductsStatuses | `/cdr-register/v1/{industry}/data-recipients/brands/software-products/status` | 3 | `IndustryEnumRecipients` | 2..3 | yes | v2: all, banking, energy, telco. v3: all only. |
| getDataRecipientsStatuses | `/cdr-register/v1/{industry}/data-recipients/status` | 3 | `IndustryEnumRecipients` | 2..3 | yes | v2: banking, energy, telco, all. v3: all only. |
| getDataRecipients | `/cdr-register/v1/{industry}/data-recipients` | 4 | `IndustryEnumRecipients` | 3..4 | yes | v3: banking, energy, telco, all. v4: all only. |

Facts:
- The mock maximum versions match CDS 1.36.0 for all endpoints that it implements.
- The mock also keeps one older version for each endpoint.
- The mock does not implement `GetDataHolderBrandsSummary`. No route, no controller, and no gateway route contains `brands/summary` (search of the full repo at the commit above). This is the largest gap.
- `data-holders/status` does not require `x-v`. The code comment says this is intentional "inline with v1 of the standards" ([CdrApiOptions.cs L18](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Models/CdrApiOptions.cs)). CDS 1.36 marks `x-v` required on this endpoint.
- The CDS parameter lists: `If-None-Match` is defined on summary, data-holders/status, software-products/status, data-recipients/status and data-recipients. It is not defined on data-holders/brands or SSA. The mock adds ETag to data-holders/brands too (see section 6.1).
- CDS response codes (from the OpenAPI): brands 200, 400, 401, 406. Summary 200, 304, 400, 404, 406. SSA 200, 400, 401, 403, 404, 406, 422. Status and data-recipients 200, 304, 400, 406.
- The mock returns 403 on GetDataHolderBrands when the caller is not active. CDS 1.36 does not list 403 for that endpoint.
- CDS defines the SSA 200 response as `application/json` with schema `type: string`.

---

## 2. Architecture

### 2.1 Components and ports

All components run in one container under `supervisord` ([supervisord.conf](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/supervisord.conf), [Dockerfile L43-L74](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/Dockerfile)).

| Component | Project | Port | Exposed to clients | Role |
|---|---|---|---|---|
| TLS gateway | `CDR.Register.API.Gateway.TLS` | 7000 | yes (public TLS) | Ocelot reverse proxy. It routes public endpoints. |
| mTLS gateway | `CDR.Register.API.Gateway.mTLS` | 7001 | yes (mutual TLS) | Ocelot reverse proxy. It terminates mTLS and validates the client certificate. |
| Identity provider ("Infosec") | `CDR.Register.Infosec` | 7002 | backend | OIDC discovery, JWKS, token endpoint. Custom code. It is not IdentityServer. |
| Discovery API | `CDR.Register.Discovery.API` | 7003 | backend | GetDataHolderBrands, GetDataRecipients |
| Status API | `CDR.Register.Status.API` | 7004 | backend | The three status endpoints |
| SSA API | `CDR.Register.SSA.API` | 7005 | backend | SSA JWKS and GetSoftwareStatementAssertion |
| Admin API | `CDR.Register.Admin.API` | 7006 | backend, but the docs tell users to call it | Non-standard metadata load, upsert, loopback test helpers, DB migration and seed on start |
| Repository | SQL Server (`mcr.microsoft.com/mssql/server:2022-latest`) | 1433 | separate container | All metadata, lookup tables, log tables |

Sources: [README.md L193-L225](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md), each project `appsettings.json` `Kestrel.Endpoints.Https.Url`.

Port exposure:
- The Dockerfile `EXPOSE` lists only 7000, 7001, 7006 ([Dockerfile L4-L6](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/Dockerfile)).
- `docker-compose.yml` publishes 7000-7006 ([docker-compose.yml](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/docker-compose.yml)).
- `docker-compose.IntegrationTests.yml` publishes 7000-7006 ([docker-compose.IntegrationTests.yml](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/docker-compose.IntegrationTests.yml)).
- The ecosystem compose publishes only 7000, 7001, 7006 ([docker-compose.Ecosystem.yml](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/docker-compose.Ecosystem.yml)).
- All backend services also listen on HTTPS with the TLS certificate `tls-mock-register.pfx`.

Start order:
- Admin API starts first. It runs EF migrations and loads seed data.
- Discovery, Status, SSA and Infosec wait for `https://localhost:7006/health` to return 200 ([wait-until-admin-healthy-then-start.sh](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/wait-until-admin-healthy-then-start.sh)).
- The two gateways start at once.
- Compose health check: `wget https://localhost:7000/health` must return 200. The TLS gateway forwards `/health` to Admin `/health`.

### 2.2 Gateway routes (which endpoint is where)

TLS gateway, port 7000 ([gateway-config.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.TLS/gateway-config.json)):

| Method | Upstream path on 7000 | Downstream |
|---|---|---|
| GET | `/health` | Admin 7006 |
| POST | `/admin/metadata/data-holders` | Admin 7006 (non-standard, see 3.2) |
| POST | `/admin/metadata/data-recipients` | Admin 7006 (non-standard) |
| GET | `/cdr-register/v1/{industry}/data-recipients` | Discovery 7003 |
| GET | `/cdr-register/v1/{industry}/data-recipients/status` | Status 7004 |
| GET | `/cdr-register/v1/{industry}/data-recipients/brands/software-products/status` | Status 7004 |
| GET | `/cdr-register/v1/{industry}/data-holders/status` | Status 7004 |
| GET | `/cdr-register/v1/jwks` | SSA 7005 |
| GET | `/idp/.well-known/openid-configuration` | Infosec 7002 |
| GET | `/idp/.well-known/openid-configuration/jwks` | Infosec 7002 |

mTLS gateway, port 7001 ([gateway-config.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.mTLS/gateway-config.json)):

| Method | Upstream path on 7001 | Downstream |
|---|---|---|
| POST | `/idp/connect/token` | Infosec 7002 |
| GET | `/cdr-register/v1/{industry}/data-holders/brands` | Discovery 7003 |
| GET | `/cdr-register/v1/{industry}/data-recipients/brands/{brandId}/software-products/{softwareProductId}/ssa` | SSA 7005 |

Notes:
- All routes set `DangerousAcceptAnyServerCertificateValidator: true`. The gateways do not validate backend TLS certificates.
- The TLS gateway exposes the admin upsert endpoints (data-holders, data-recipients) on the public port. It does not expose `GET/POST /admin/metadata` or `/loopback/*`. Those are only on 7006.
- The mTLS gateway has no `/health` route.

### 2.3 How TLS and mTLS terminate

- Each gateway terminates TLS in Kestrel.
- TLS gateway (7000): certificate `Certificates/tls-mock-register.pfx`, password `#M0ckRegister#` ([appsettings.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.TLS/appsettings.json)).
- mTLS gateway (7001): certificate `Certificates/register.pfx`, password `#M0ckRegister#`, root CA `Certificates/ca.pem` ([appsettings.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.mTLS/appsettings.json)).
- mTLS gateway TLS settings ([Program.cs L66-L96](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.mTLS/Program.cs)):
  - TLS 1.2 only (`SslProtocols.Tls12`).
  - `ClientCertificateMode.RequireCertificate`. The TLS handshake fails without a client certificate. The client sees a connection error, not an HTTP status. The integration tests expect `HttpRequestException` for an expired client certificate and for a server certificate used as a client certificate.
  - Cipher suites on non-Windows: `TLS_DHE_RSA_WITH_AES_128_GCM_SHA256`, `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`, `TLS_DHE_RSA_WITH_AES_256_GCM_SHA384`, `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`.
- Backend services also force TLS 1.2 ([Extensions.cs L110-L128](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs)). UNVERIFIED: whether the TLS gateway (7000) also restricts cipher suites. I did not read its `Program.cs`.
- mTLS client certificate validation ([Startup.cs L32-L66](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.mTLS/Startup.cs), [CertificateValidator.cs L23-L56](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Certificates/CertificateValidator.cs)):
  - ASP.NET certificate authentication with `ValidateCertificateUse = true`, `ValidateValidityPeriod = true`, `RevocationMode = NoCheck`.
  - A custom chain build trusts only `ca.pem` (custom root trust). Any chain status error throws `ClientCertificateException`.
  - The gateway exception handler returns 400 `text/plain` "An error occurred handling the request: ..." for `ClientCertificateException`. It returns 502 `text/plain` for other errors ([Startup.cs L77-L97](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.mTLS/Startup.cs)).
  - The validation result is cached (`AddCertificateCache`).
- Certificate forwarding to backends ([Startup.cs L103-L124](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Gateway.mTLS/Startup.cs)):
  - The gateway appends header `X-TlsClientCertThumbprint` = `X509Certificate2.Thumbprint`. This is the **SHA-1** hash, upper-case hex, 40 characters.
  - The gateway appends header `X-TlsClientCertCN` = the certificate simple name (CN).
  - The gateway appends `X-Forwarded-Host` = the original `Host`.
  - You can change the header names with config keys `Certificate:CertThumbprintNameHttpHeaderName` and `Certificate:CertCommonNameHttpHeaderName` ([Constants.cs L21-L22](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Constants.cs)).
- The backends trust these headers. They do not see the TLS client certificate. Backends read the first header value (`headers[0]`) ([Extensions.cs L350-L360](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs)).
- Security note (UNVERIFIED): the gateway uses `Headers.Append`. It does not remove a client-supplied `X-TlsClientCertThumbprint` or `X-TlsClientCertCN` header. If Ocelot forwards client headers (its default), the client value is `headers[0]`, and the backend uses it. This lets a client with any valid mock-CA certificate spoof the thumbprint and CN. Also, if you publish ports 7002-7005, a client can call a backend directly with any header values. The integration tests do this on purpose ([US50483_DynamicUrl_Tests.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/Miscellaneous/US50483_DynamicUrl_Tests.cs)). Do not copy this design into the Go Register.

### 2.4 Host names, base URLs and dynamic base paths

- Config key `PublicHostName` (for example `https://mock-register:7000`) and `SecureHostName` (for example `https://mock-register:7001`) set the host in links and in OIDC metadata ([Infosec appsettings.Release.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/appsettings.Release.json)).
- If `PublicHostName` or `SecureHostName` is empty, the services use the `X-Forwarded-Host` header, with `https://` added ([Extensions.cs L253-L277](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs)).
- The Infosec issuer is `{PublicHostName}{PathBase}/idp`. The token endpoint is `{SecureHostName}{PathBase}/idp/connect/token` ([Extensions.cs L44-L57](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs)).
- With the default Release config the issuer is `https://mock-register:7000/idp`. Clients must resolve the host name `mock-register`. The Help guide tells users to add `127.0.0.1 mock-register` to the hosts file ([Help/container/HELP.md L70-L77](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Help/container/HELP.md)).
- `BasePath` sets a static path prefix. `BasePathExpression` sets a regular expression. The regex splits the path into `PathBase` and the rest ([Extensions.cs L81-L108](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs)). The ACCC Conformance Test Suite (CTS) uses paths like `/cts/{conformanceId}/register/...`.
- If `BasePathExpression` is set, Discovery and SSA also check that the access token `iss` contains the `PathBase`. If not, they return 401 `{"errors":[{"code":"401","title":"Unauthorized","detail":"invalid_token"}]}` ([Extensions.cs L63-L79](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs), [DiscoveryController.cs L110-L119](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Controllers/DiscoveryController.cs)). The Go Register does not need this CTS feature.

### 2.5 Certificates shipped

All in [CertificateManagement/](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CertificateManagement/README.md). Each certificate has `.pfx`, `.pem`, `.key`, `.csr`, `.cnf` and an `openssl` `.cmd` script. Values below are from `openssl x509` on the `.pem` files.

| File | Password | Subject CN | Issuer | EKU | SAN | Expires (UTC) | SHA-1 thumbprint |
|---|---|---|---|---|---|---|---|
| `mtls/ca.pem` (`ca.pfx`, `ca.crt`, `ca.key`) | `#M0ckCDRCA#` | Mock CDR CA | self | - | - | 2031-04-03 | `3AF1334F5E3831EF194E6F7E9E0FAADA11D5A459` |
| `mtls/register.pem` | `#M0ckRegister#` | register.mock | Mock CDR CA | serverAuth (critical) | localhost, register.mock | 2028-06-14 | `C9D6938007937DEB339F47092F519A2A3D0CF53C` |
| `mtls/register-client.pem` | `#M0ckRegister#` | (not read) | Mock CDR CA | - | - | - | - |
| `mtls/client.pem` (Mock Data Recipient) | `#M0ckDataRecipient#` | MockDataRecipient | Mock CDR CA | clientAuth (critical) | - | 2028-06-14 | `F0E5146A51F16E236844CF0353D791F11865E405` |
| `mtls/client-additional.pem` | `#M0ckDataRecipient#` | AdditionalClientCertForTesting | Mock CDR CA | clientAuth | - | 2028-06-14 | `291FDE2AC6A3E8949305DA01F4ABE0C533A88928` |
| `mtls/client-invalid.pem` | `#M0ckDataRecipient#` | InvalidClientCertForTesting | Mock CDR CA | clientAuth | - | **2022-05-12 (expired)** | `E449D67F6C7B74D1B7B7321C1B376DC961DB0B93` |
| `mtls/server.pem` (Mock Data Holder) | `#M0ckDataHolder#` | mock-data-holder | Mock CDR CA | serverAuth | localhost, mock-data-holder, mock-data-holder-energy | 2028-08-10 | `0E707545DF43DC51E32DEA3E9C50B0E8703CD89C` |
| `tls/mock-register.pem` (also `tls-mock-register.pfx`) | `#M0ckRegister#` | mock-register | Mock CDR CA | serverAuth | mock-register, localhost | 2028-06-15 | `B9B4D3221E489EED75CF7892CC0B2B19D1F5C198` |
| `ssa/ssa.pem` | `#M0ckRegister#` | MockRegisterSSA | **self-signed** | serverAuth | mockregister.ssa | 2028-06-15 | `203A41DC1743F97212BFBF01B77895CD9F445BFB` |

Also shipped: `tls/mock-data-holder*.pfx`, `tls/mock-data-recipient.pfx`, `mtls/server-energy.pfx` (the README lists it; I did not find the file in the tree).

Facts:
- The SHA-256 thumbprint of `client.pem` (base64url, RFC 8705 form) is `fB7Ej_NQPgg3m-Zlih6VSKL2c3Oq3GIot1EZyGbelgY`. The mock does not use this value (see 4.4).
- The mTLS server certificate `register.pem` has SAN `localhost` and `register.mock`. It does not have `mock-register`. A client that verifies host names fails on `https://mock-register:7001`. The integration tests turn off server certificate validation.
- `register.pfx` has two jobs. It is the mTLS server certificate on 7001. It is also the access-token signing key in Infosec (`SigningCertificate` in [Infosec appsettings.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/appsettings.json)).
- `ssa.pfx` signs SSAs. It is self-signed. Admin also loads it for the loopback `RegisterSelfSignedJwt` helper.
- The Docker image copies `ca.crt` into the OS trust store (`update-ca-certificates`) ([Dockerfile L56-L57](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/Dockerfile)).
- The certificates expire in 2028 (CA in 2031). A Go mock that reuses them must re-issue them before 2028-06-14.
- Issue #60 reported that the Register and Data Holder server certificates had the same DN. The team fixed it in v1.3.2 ([issue #60](https://github.com/ConsumerDataRight/mock-register/issues/60)).

### 2.6 How participants get certificates

- The mock has no certificate issuance API. It does not simulate DigiCert on-boarding.
- Participants use the shipped certificates. Or they run the `openssl` `.cmd` scripts with the mock CA key (`ca.key` is in the repo) ([CertificateManagement/README.md](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CertificateManagement/README.md)).
- To bind a client certificate to a software product, you put `commonName` and `thumbprint` in the software product `certificates` array in seed data or through the Admin API. The token endpoint checks only the CN (see 4.3).
- If you change `client.pfx`, you must change the thumbprint in seed data too (README note).

---

## 3. Endpoints

### 3.1 Standard endpoints

| Method | Path | Gateway | Auth | Response |
|---|---|---|---|---|
| GET | `/idp/.well-known/openid-configuration` | 7000 | none | OIDC metadata (4.1) |
| GET | `/idp/.well-known/openid-configuration/jwks` | 7000 | none | Access-token signing JWKS (4.2) |
| POST | `/idp/connect/token` | 7001 | mTLS + `private_key_jwt` | Access token (4.3) |
| GET | `/cdr-register/v1/jwks` | 7000 | none | SSA signing JWKS (5.3) |
| GET | `/cdr-register/v1/{industry}/data-holders/brands` | 7001 | mTLS + bearer, scope `cdr-register:read`, cert-bound | Paginated brands. `x-v` 2 or 3. |
| GET | `/cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa` | 7001 | mTLS + bearer, scope `cdr-register:read`, cert-bound | SSA JWT. `x-v` 3 or 4. |
| GET | `/cdr-register/v1/{industry}/data-recipients` | 7000 | none | `x-v` 3 or 4. The same payload for both versions. |
| GET | `/cdr-register/v1/{industry}/data-recipients/status` | 7000 | none | `x-v` 2 or 3 |
| GET | `/cdr-register/v1/{industry}/data-recipients/brands/software-products/status` | 7000 | none | `x-v` 2 or 3 |
| GET | `/cdr-register/v1/{industry}/data-holders/status` | 7000 | none | `x-v` 1 or 2 |

Controller sources:
- [Discovery DiscoveryController.cs L44-L152](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Controllers/DiscoveryController.cs)
- [StatusController.cs L28-L107](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Status.API/Controllers/StatusController.cs)
- [SSAController.cs L38-L92](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.SSA.API/Controllers/SSAController.cs)
- [Infosec DiscoveryController.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/DiscoveryController.cs), [TokenController.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/TokenController.cs)

### 3.2 Non-standard admin and test endpoints

All on Admin API, port 7006, HTTPS with `tls-mock-register.pfx`. Controller: [AdminController.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/AdminController.cs), [LoopbackController.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/LoopbackController.cs).

| Method | Path | Auth | x-v | Behaviour |
|---|---|---|---|---|
| GET | `/admin/metadata` | none, always | neutral | Returns all metadata as `{ "legalEntities": [...] }`. Same shape as seed data. Nulls omitted. Integer FK ids. Ordered by `legalEntityName`. ([L77-L98](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/AdminController.cs)) |
| POST | `/admin/metadata` | none, always | neutral | Deletes **all** legal entities (cascade). Loads the body. Body = seed-data shape. 200 with empty body on success. 400 `"Database not updated"` if body has no `legalEntities`. 400 `"UnexpectedError, An error occurred loading the database."` on any exception. `Content-Type: application/json` on all results. ([L41-L75](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/AdminController.cs)) |
| POST | `/admin/metadata/data-holders` | optional (below) | `x-v: 1` mandatory | Create or update one data holder brand. Body = `DataHolderBrandModel`. 200 on success. 400 with CDS `errors` for validation. 500 with exception text on error. ([L100-L135](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/AdminController.cs)) |
| POST | `/admin/metadata/data-recipients` | optional | `x-v: 1` mandatory | Create or update one data recipient legal entity with brands and software products. Body = `LegalEntity` model. Empty body gives 400 `Field/Invalid` "Empty LegalEntity received". A software product ID that exists under a different brand gives 400 `Field/Invalid`. ([L137-L176](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/AdminController.cs), [Repository Extensions.cs L236-L274](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/Infrastructure/Extensions.cs)) |
| GET | `/loopback/MockDataRecipientJwks` | none | neutral | JWKS for the mock DR key `client.pem`. One RSA key, `alg` PS256, `key_ops` [sign, verify], `kid` = base64url(SHA-256 of JSON `{"e","kty","n"}` with Newtonsoft serialization). ([L32-L55](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/LoopbackController.cs)) |
| GET | `/loopback/MockDataRecipientClientAssertion?iss=&aud=` | none | neutral | Returns a signed client assertion (PS256, `client.key`). Default `iss`/`sub` = `LoopbackDefaultSoftwareProductId` (`6F7A1B8E-8799-48A8-9011-E3920391F713`). Default `aud` = `IdentityServerTokenUri`. 10 min expiry. Random `jti`. ([L64-L106](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/LoopbackController.cs)) |
| GET | `/loopback/RegisterSelfSignedJwt?aud=` | none | neutral | Returns a JWT signed with `SigningCertificate` (ssa.pfx), `iss`=`sub`=`cdr-register`, 10 min expiry. ([L112-L136](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Controllers/LoopbackController.cs)) |
| GET | `/health` | none | - | JSON `{"status":[{"key":"migration","value":"Healthy"},{"key":"seed-data","value":"Healthy"}]}`. Unhealthy until migration and seed finish. ([Admin Startup.cs L81-L173](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Startup.cs)) |
| GET | `/swagger/...` | none | - | `EnableSwagger: true` in each backend `appsettings.json`. UNVERIFIED: exact Swagger URL. |

Other services also map `/health` (for example [Discovery Startup.cs L72](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Startup.cs)).

Admin authentication for the two upsert endpoints ([ApiAuthorizeAttribute.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Filters/ApiAuthorizeAttribute.cs), [Admin Common/Constants.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Common/Constants.cs)):
- Config keys are `Authorization:Issuer`, `Authorization:ClientId`, `Authorization:Scope`, `Authorization:TokenScopeAttribute` (default `http://schemas.microsoft.com/ws/2008/06/identity/claims/role`).
- The README says the keys are `oidc.issuer` and `oidc.scope` ([README.md L160-L171](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md)). The README is wrong. The code uses `Authorization:*`.
- If `Authorization:Issuer` is empty (the default), there is no authentication.
- If it is set, the caller needs a bearer token with the role claim equal to `Authorization:Scope`. Failure gives 401 `{"error":"invalid_token"}` and `WWW-Authenticate: Bearer error="invalid_token"`. The integration tests call this with Azure AD client credentials (CTS only).
- `GET`/`POST /admin/metadata` never have authentication.

Admin request bodies (from [mock-register.ipynb](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Help/notebooks/mock-register.ipynb) and [Business/Model](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Business/Model)):

`POST /admin/metadata/data-holders` (`x-v: 1`):
```json
{
  "dataHolderBrandId": "bb03be60-5c46-422e-a27e-aefa0015078d",
  "brandName": "Luna",
  "brandGroup": "Luna Banking Group",
  "industries": ["banking"],
  "logoUri": "https://lunabrand/img/logo.png",
  "status": "ACTIVE",
  "endpointDetail": {
    "version": "v1",
    "publicBaseUri": "https://publicapi.lunabank",
    "productBaseUri": "https://productapi.lunabank",
    "resourceBaseUri": "https://rb-api.lunabank",
    "infosecBaseUri": "https://isb-api.lunabank",
    "extensionBaseUri": "https://eb-api.lunabank",
    "websiteUri": "https://ws-api.lunabank"
  },
  "authDetails": { "registerUType": "SIGNED-JWT", "jwksEndpoint": "https://idp.lunabank/jwks" },
  "legalEntity": {
    "legalEntityId": "667407fc-dc7e-4be5-8795-5189240bcbca",
    "legalEntityName": "Luna Bank",
    "logoUri": "https://lunabank/img/logo.png",
    "registrationNumber": "20032023",
    "registrationDate": "2023-05-22",
    "registeredCountry": "AU",
    "abn": "12345678901", "acn": "345678901", "arbn": "345678901",
    "anzsicDivision": "123",
    "organisationType": "OTHER",
    "status": "ACTIVE"
  }
}
```
Note: `authDetails` is a single object here. In the GetDataHolderBrands response it is an array.

`POST /admin/metadata/data-recipients` (`x-v: 1`):
```json
{
  "legalEntityId": "f9bde9a1-9fef-4bd9-b7ef-877369e076e3",
  "legalEntityName": "Mock Energy Company",
  "accreditationNumber": "ADR000001",
  "accreditationLevel": "UNRESTRICTED",
  "logoUri": "https://mocksoftware/img/Companylogo.png",
  "abn": "34241177887", "acn": "241177887", "arbn": "342411088",
  "organisationType": "COMPANY",
  "status": "ACTIVE",
  "dataRecipientBrands": [{
    "dataRecipientBrandId": "d50b74cf-e992-42fd-86d4-582ac9d72dbc",
    "brandName": "Mock Energy Tool",
    "logoUri": "https://mocksoftware/img/brandlogo.png",
    "status": "ACTIVE",
    "softwareProducts": [{
      "softwareProductId": "4754cf30-6cc2-4a9a-a9e1-4fb8f6d42121",
      "softwareProductName": "MyEnergyHelper",
      "softwareProductDescription": "A product to help you manage your energy costs",
      "logoUri": "https://mocksoftware/myenergyapp/img/logo.png",
      "clientUri": "https://mocksoftware/myenergyapp",
      "tosUri": "...", "policyUri": "...",
      "recipientBaseUri": "https://mock-data-recipient:9001",
      "revocationUri": "revocation",
      "redirectUris": ["consent/callback", "consent/callback2"],
      "jwksUri": "jwks",
      "scope": "openid profile bank:accounts.basic:read ... cdr:registration cdr-register:read",
      "status": "ACTIVE",
      "certificates": [{ "commonName": "MockDataRecipient", "thumbprint": "f0e5146a51f16e236844cf0353d791f11865e405" }]
    }]
  }]
}
```
- If `scope` is null, the Admin API uses config `SoftwareProductDefaultScopes` ([SoftwareScopeResolver.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Business/SoftwareScopeResolver.cs), [Admin appsettings.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/appsettings.json)).
- `redirectUris` (array) is stored as one space-separated string ([AdminMappingProfile.cs L29](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Business/AdminMappingProfile.cs)).
- The upsert sets `Brand.LastUpdated` to the current UTC time ([Repository Extensions.cs L201](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/Infrastructure/Extensions.cs)).
- UNVERIFIED: whether validators accept relative URIs such as `"jwks"`. The notebook sends them. I did not read the FluentValidation rules in [Business/Validators](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Business/Validators).

---

## 4. Auth: identity provider and token endpoint

### 4.1 OIDC discovery document

Source: [Infosec DiscoveryController.cs L21-L44](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/DiscoveryController.cs). Custom code. It is not a library.

| Field | Value |
|---|---|
| `issuer` | `{PublicHostName}{PathBase}/idp` |
| `jwks_uri` | `{PublicHostName}{PathBase}/idp/.well-known/openid-configuration/jwks` |
| `token_endpoint` | `{SecureHostName}{PathBase}/idp/connect/token` |
| `claims_supported` | `["sub"]` |
| `grant_types_supported` | `["client_credentials"]` |
| `id_token_signing_alg_values_supported` | `["PS256"]` |
| `response_types_supported` | `["token"]` |
| `scopes_supported` | `["cdr-register:read"]` |
| `code_challenge_methods_supported` | `["plain","S256"]` |
| `subject_types_supported` | `["public"]` |
| `tls_client_certificate_bound_access_tokens` | `true` |
| `token_endpoint_auth_methods_supported` | `["private_key_jwt"]` |
| `token_endpoint_auth_signing_alg_values_supported` | `["PS256"]` |

The test [US12665_IdentityServer_OIDC_Tests.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/IdentityServer/US12665_IdentityServer_OIDC_Tests.cs) asserts this document.

### 4.2 Access-token JWKS

Source: [Infosec DiscoveryController.cs L46-L84](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/DiscoveryController.cs).
- One RSA key from `register.pfx`.
- `kty` RSA, `use` sig, `alg` PS256.
- `kid` = `X509SigningCredentials.Kid` = SHA-1 thumbprint, upper-case hex.
- `x5t` = base64url SHA-1 of the DER certificate.
- `x5c` = base64 DER certificate.
- The service reads the certificate file on every request.
- Issue #13: the community asked for RFC 7638 `kid` values. The ACCC answered that the production Register does not use RFC 7638, and that clients must treat `kid` as an opaque string ([issue #13](https://github.com/ConsumerDataRight/mock-register/issues/13)).

### 4.3 Token endpoint `POST /idp/connect/token`

Sources: [TokenController.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/TokenController.cs), [TokenService.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/TokenService.cs), [ClientService.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/ClientService.cs).

Library: custom code with `System.IdentityModel.Tokens.Jwt` (`JwtSecurityTokenHandler`) and `Microsoft.IdentityModel.Tokens`. There is no OAuth server library.

Request:
- `Content-Type` must be `application/x-www-form-urlencoded` (`ValidateContentTypeFilter`). Tests expect:
  - Missing content type: 400 `{"error":"invalid_request","error_description":"Content-Type is not application/x-www-form-urlencoded"}`.
  - `application/foo`: 415 Unsupported Media Type ([US14045 L220-L221](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/IdentityServer/US14045_IdentityServer_Token_Tests.cs)).
- Form fields: `grant_type`, `client_assertion_type`, `client_assertion`, `scope`, optional `client_id`.

Validation order ([TokenController.cs L114-L171](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/TokenController.cs)). Each failure returns 400 `{"error": ..., "error_description": ...}`:

| Step | Check | `error` | `error_description` |
|---|---|---|---|
| 1 | `grant_type` present | invalid_client | `grant_type not provided` |
| 1 | `client_assertion` present | invalid_client | `client_assertion not provided` |
| 1 | `client_assertion_type` present | invalid_client | `client_assertion_type not provided` |
| 2 | `grant_type` = `client_credentials` (case-insensitive) | unsupported_grant_type | `grant_type must be client_credentials` |
| 3 | `client_assertion_type` = `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` | invalid_client | `client_assertion_type must be urn:ietf:params:oauth:client-assertion-type:jwt-bearer` |
| 4 | `scope` present and not blank | invalid_client | `empty scope` |
| 4 | every scope token = `cdr-register:read` | invalid_client | `invalid scope` |
| 5 | `client_assertion` parses as JWT | invalid_client | `Invalid client_assertion - token validation error` |
| 6 | `iss` present | invalid_client | `client_id is required` |
| 6 | if form `client_id` present: equals `iss` and `sub` (case-insensitive) | invalid_client | `Invalid client_assertion - 'sub' and 'iss' must be set to the client_id` |
| 6 | `iss` is a GUID of an existing software product | invalid_client | `invalid client_id` |
| 6 | signature, `aud`, `exp` valid (below) | invalid_client | `Invalid client_assertion - token validation error` |
| 6 | `jti` present | invalid_client | `Invalid client_assertion - 'jti' is required` |
| 6 | header `alg` is PS256 or ES256 | invalid_client | `Invalid client_assertion - Client assertion token signature algorithm must be PS256 or ES256` |
| 6 | `sub` = `iss` | invalid_client | `Invalid client_assertion - 'sub' and 'iss' must have the same value` |
| 6 | `sub` and `iss` = software product ID | invalid_client | `Invalid client_assertion - 'sub' and 'iss' must be set to the client_id` |
| 6 | `jti` not used before (per `iss`) | invalid_client | `Invalid client_assertion - 'jti' in the client assertion token must be unique` |
| 7 | mTLS CN header matches a registered certificate CN | invalid_client | `Client certificate validation failed` |

Facts:
- `client_id` semantics: `client_id` = the software product ID (GUID). The form `client_id` is optional. The code reads the client from the assertion `iss` ([TokenService.cs L41-L69](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/TokenService.cs), [ClientService.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/ClientService.cs)).
- Software product lookup ignores status. An inactive product can get a token. The resource endpoints then return 403 (see 6.4). UNVERIFIED: the exact repository query in `RegisterInfosecRepository.GetSoftwareProductAsync`.
- Client keys: the Register fetches the software product `jwksUri` on each token request. It uses `HttpClient` with server certificate validation off by default (`EnableServerCertificateValidation: false`). If fetch fails, the key list is empty and validation fails ([TokenService.cs L155-L202](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/TokenService.cs)). Option `PassUserAgent: true` adds `User-Agent: mock-register`.
- Assertion validation parameters ([TokenService.cs L221-L269](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/TokenService.cs)):
  - `ValidateIssuer = false`.
  - Valid audiences: `{PublicHostName}{PathBase}/idp` and `{SecureHostName}{PathBase}/idp/connect/token`. Comparison is case-insensitive.
  - `RequireExpirationTime`, `ValidateLifetime`, clock skew 120 s.
  - The code does not check `iat` or `nbf` beyond library defaults. It has no maximum lifetime.
- `jti` replay: the key `{iss}::{jti}` goes into `IDistributedCache` until `exp + 5 min` ([TokenService.cs L97-L110, L204-L219](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/TokenService.cs)). UNVERIFIED: which cache implementation (in-memory or SQL). `Constants.cs` names a `Register_Cache` connection string.
- The mTLS check uses **only the CN**. The code comment says "validating the thumbprint is not required" ([TokenController.cs L173-L196](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/TokenController.cs)). The CN header can be a full DN. The code extracts `CN=`. The DB value can also be a full DN. The comparison is case-insensitive.
- The scope rule rejects any scope other than exactly `cdr-register:read`. A code comment says this is a known limitation ([TokenController.cs L84-L87](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/TokenController.cs)).
- The scope errors use `invalid_client`. RFC 6749 uses `invalid_scope` for this case. This is a deviation.
- A missing `scope` is an error ("empty scope"). The default-scope code path after it cannot run.

Token response (200) ([TokenController.cs L31-L53](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Controllers/TokenController.cs)):
```json
{ "access_token": "<JWT>", "expires_in": 300, "token_type": "Bearer", "scope": "cdr-register:read" }
```
- `expires_in` comes from `AccessTokenExpiryInSeconds` (default 300).
- UNVERIFIED: exact JSON property names. The test deserializes `Access_token`, `Expires_in`, `Token_type`, `Scope`. That implies snake_case names.

### 4.4 Access token format and certificate binding

Source: [TokenService.cs L115-L153](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Infosec/Services/TokenService.cs).
- Header: `alg` PS256, `kid` = SHA-1 thumbprint of `register.pfx`, `typ` **`at+jwt`**.
- Claims: `iss` = `{PublicHostName}{PathBase}/idp`, `aud` = `cdr-register`, `client_id` = software product ID, `jti` (GUID), `scope`, `nbf`, `exp`, `iat`, and `cnf`.
- `cnf` = `{"x5t#S256": "<value of X-TlsClientCertThumbprint header>"}`.
- **Deviation:** the value is the SHA-1 upper-case hex thumbprint (for example `F0E5146A51F16E236844CF0353D791F11865E405`). RFC 8705 requires base64url of the SHA-256 hash. The integration test asserts the SHA-1 hex value ([US14045_IdentityServer_Token_Tests.cs L205-L210](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/IdentityServer/US14045_IdentityServer_Token_Tests.cs)).
- The token has no `sub` claim.

Resource-side token validation ([Extensions.cs L130-L189](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs), [MTLSHandler.cs L25-L83](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Authorization/MTLSHandler.cs), [ScopeHandler.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Authorization/ScopeHandler.cs)):
- JwtBearer. JWKS loaded once at start from `{OidcMetadataAddress}/jwks` (with retry). `ValidAudience = cdr-register`. `ValidateIssuer = false`. Lifetime checked. Clock skew `ClockSkewSeconds` (default 120).
- Scope handler: the `scope` claim must contain `cdr-register:read`.
- mTLS handler: `cnf.x5t#S256` must equal the `X-TlsClientCertThumbprint` header (exact string compare).
- Policies `DataHolderBrandsApiMultiIndustry` and `GetSSAMultiIndustry` both need the scope and mTLS ([RegisterAuthorisationPolicy.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Authorization/RegisterAuthorisationPolicy.cs)).
- If the mTLS requirement fails, the response is 401 `{"errors":[{"code":"401","title":"Unauthorized","detail":"invalid_token"}]}`. Other failures give 403 `{"errors":[{"code":"403","title":"Forbidden","detail":"You do not have permission to access this resource."}]}` ([PolicyAuthorizeAttribute.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Authorization/PolicyAuthorizeAttribute.cs)).
- UNVERIFIED: body and `WWW-Authenticate` for a missing or invalid bearer token. The attribute inherits `AuthorizeAttribute`, so the JwtBearer challenge probably returns 401 with an empty body. The tests assert only the 401 status.
- The SSA and brands handlers read `client_id` from the token. A token without a GUID `client_id` gives 400 `GeneralError/Unexpected`.

---

## 5. SSA

### 5.1 Generation

Sources: [SSAController.cs L71-L133](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.SSA.API/Controllers/SSAController.cs), [SSAService.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.SSA.API/Business/SSAService.cs), [Mapper.cs L20-L46](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.SSA.API/Business/Mapper.cs), [TokenizerService.cs L29-L62](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.SSA.API/Business/TokenizerService.cs), [CertificateService.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.SSA.API/Business/CertificateService.cs).

Order of checks:
1. Policy: bearer token, scope, cert binding (4.4).
2. Token `client_id` must be a GUID, else 400 `GeneralError/Unexpected`.
3. Token `client_id` must equal the path `softwareProductId` (case-insensitive), else **404** `urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct`, title `Invalid Software Product`, detail = the path value.
4. Status check: product, brand and participation must all be ACTIVE, else 403 `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`, title `ADR Status Is Not Active`, detail `""` ([DataRecipientStatusCheckService.cs L26-L50](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Services/DataRecipientStatusCheckService.cs)).
5. Load by `brandId` + `softwareProductId`. Not found or non-GUID gives 404 `urn:au-cds:error:cds-all:Resource/NotFound`, title `Resource Not Found`, detail `""`.

SSA JWT:
- Header: `{"alg":"PS256","kid":"<SHA-1 thumbprint of ssa.pfx, upper-case hex>","typ":"JWT"}`. The code builds it by hand.
- Signature: RSA-PSS SHA-256 with `ssa.pfx`.
- Claims:

| Claim | Source |
|---|---|
| `iss` | config `SSA:Issuer` = `cdr-register` |
| `iat` | now (Unix s) |
| `exp` | now + `SSA:ExpiryInSeconds` (600) |
| `jti` | GUID without hyphens |
| `org_id` | DR brand ID |
| `org_name` | DR brand name |
| `client_name` | software product name |
| `client_description` | software product description |
| `client_uri` | software product clientUri |
| `redirect_uris` | array (the DB string split on spaces) |
| `logo_uri` | software product logoUri |
| `tos_uri`, `policy_uri` | optional |
| `jwks_uri` | software product jwksUri |
| `revocation_uri` | software product revocationUri |
| `recipient_base_uri` | software product recipientBaseUri |
| `software_id` | software product ID |
| `software_roles` | `data-recipient-software-product` |
| `scope` | software product scope, stored value, no filter |
| `sector_identifier_uri` | optional |
| `legal_entity_id`, `legal_entity_name` | from the DR legal entity |

- `x-v` 3 and 4 give the same SSA. The scope does not change with industry.
- Required fields are validated. A missing required field throws `SsaValidationException`, which becomes 500 `GeneralError/Unexpected`.
- UNVERIFIED: the response `Content-Type`. The controller returns `Ok(string)` and has no `[Produces]`. ASP.NET normally uses the string formatter, which gives `text/plain`. With `Accept: application/json` it may give a JSON-quoted string. CDS defines `application/json` with a string schema. Test this against the running mock before you fix the Go behaviour.
- The response has `x-v` but no ETag.

### 5.2 SSA scope rules in tests

- The test `GetSSA_WithDifferentIndustry_ShouldRespondWith_DifferentScopes` only runs `industry=all`. It expects the stored scope ([GetSoftwareStatementAssertion_Tests.cs L104-L134, L399-L470](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/API/SSA/GetSoftwareStatementAssertion_Tests.cs)).

### 5.3 SSA JWKS `GET /cdr-register/v1/jwks`

- One key from `ssa.pfx`: `kty` RSA, `alg` PS256, `kid` = SHA-1 thumbprint upper-case hex, `key_ops` `["verify"]`, `n`, `e` ([CertificateService.cs L47-L64](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.SSA.API/Business/CertificateService.cs)).
- No `use`, `x5t` or `x5c` (the mapper copies them, but the RSA converter does not set them).
- CHANGELOG 2.2.4 (2026-02-25): "Fixed JWKS endpoint response to conform CDR Data Standard".
- Issue #80: the path `/cdr-register/v1/jwks` copies production (`https://api.cdr.gov.au/cdr-register/v1/jwks`). It is not the `/jwk` path that an older CDS text showed ([issue #80](https://github.com/ConsumerDataRight/mock-register/issues/80)).

---

## 6. Behaviour details

### 6.1 ETag

Source: [ETagAttribute.cs L19-L63](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Filters/ETagAttribute.cs).
- Applied to GetDataHolderBrands, GetDataRecipients and the three status endpoints. Not applied to SSA, JWKS, OIDC or token.
- Only for GET and status 200.
- Value = upper-case hex SHA-256 of the Newtonsoft serialization of the action result object. The header is `ETag: "<hex>"` (strong, quoted).
- The result includes `links.self`. So the ETag changes with host, port, path base and query string.
- `If-None-Match`: the filter trims `"` and compares the full header string. It does not parse lists, `W/` or `*`. Match gives 304 with no body. The ETag header is on the 304 too.
- The action runs in full before the ETag check. The 304 saves only bandwidth.
- Error responses do not have status 200. So they never get an ETag and never get a 304. (Derived from the code. UNVERIFIED by test.)

### 6.2 Pagination (GetDataHolderBrands only)

Sources: [DiscoveryController.cs L102-L152](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Controllers/DiscoveryController.cs), [RegisterDiscoveryRepository.cs L75-L108](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/RegisterDiscoveryRepository.cs), [Extensions.cs L196-L321](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Extensions.cs), filters `CheckPage`, `CheckPageSize`, `CheckDate`.

- Defaults: `page` = 1, `page-size` = 25.
- `page-size` must be a positive integer, else 400 `urn:au-cds:error:cds-all:Field/InvalidPageSize`, title `Invalid Page Size`, detail `Page size not a positive Integer`.
- `page-size` > 1000 gives 400 `Field/Invalid`, title `Invalid Field`, detail `Page size too large`.
- `page` must be a positive integer, else 400 `Field/Invalid`, detail `Page not a positive integer`. (It uses `Field/Invalid`, not `Field/InvalidPage`.)
- `page` > total pages (and `page` != 1) gives 400 `Field/Invalid`, detail `Page is out of range`. Page 1 of an empty set is 200.
- `updated-since`: parsed with .NET `DateTime.Parse` (invariant culture). Invalid gives 400 `urn:au-cds:error:cds-all:Field/InvalidDateTime`, title `Invalid DateTime`, detail `updated-since should be valid DateTimeString`. An empty value is ignored (200).
- Filter: `Brand.LastUpdated > updated-since` (strict).
- Sort: `brandName` ascending, then `brandId`.
- Filter by industry of the participation, unless `all`. v2 also removes non-bank-lending brands.
- **The query does not filter by brand or participation status.** The integration test builds its expected result with ACTIVE-only filters ([GetDataHolderBrandsXV3_MultiIndustry_Tests.cs L903-L917](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/API/Discovery/GetDataHolderBrandsXV3_MultiIndustry_Tests.cs)). The tests pass only because all seed brands are ACTIVE. UNVERIFIED: which behaviour production has. The Go Register must choose one and document it.
- `meta`: `totalRecords`, `totalPages`.
- `links`: `self`, `first`, `last`, `prev`, `next`. `prev` is null on page 1. `next` is null on the last page. Null values are omitted in JSON (serializer `NullValueHandling.Ignore`, [CdrJsonSerializerSettings.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Domain/Models/CdrJsonSerializerSettings.cs)).
- Links carry `page` and `page-size` always (when not null), and `updated-since` in ISO 8601 round-trip form (`o`), URL-encoded (`%3A`). Link hosts use `SecureHostName` or `X-Forwarded-Host` (see 2.4).
- `first`/`last` are omitted when `totalPages` = 0.

### 6.3 x-v and x-min-v

Sources: [CdrVersionReader.cs L28-L95](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Versioning/CdrVersionReader.cs), [ApiExceptionHandler.cs L36-L49](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Middleware/ApiExceptionHandler.cs), [ReturnXVAttribute.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Filters/ReturnXVAttribute.cs).

Rules:
1. `x-v` missing or blank and mandatory: 400 `urn:au-cds:error:cds-all:Header/Missing`, title `Missing Required Header`, detail `An API version x-v header is required, but was not specified.`
2. `x-v` missing and not mandatory (only data-holders/status): use the minimum version (1).
3. `x-v` not a positive integer: 400 `urn:au-cds:error:cds-all:Header/InvalidVersion`, title `Invalid Version`, detail `Version is not a positive Integer.`
4. `x-min-v` present and not a positive integer: 400 `Header/InvalidVersion` (same body).
5. `x-min-v` < `x-v`:
   - If `x-min-v` > max, or `x-v` < min: 406.
   - Else respond with min(`x-v`, max).
6. Otherwise (`x-min-v` absent, or `x-min-v` >= `x-v`): `x-min-v` is ignored. If `x-v` is outside min..max: 406.
7. 406 body: `urn:au-cds:error:cds-all:Header/UnsupportedVersion`, title `Unsupported Version`, detail `Requested version is lower than the minimum version or greater than maximum version.`
8. The response header `x-v` = the version that the action serves (set by `ReturnXV`).

Test cases for SSA ([GetSoftwareStatementAssertion_Tests.cs L290-L308](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/API/SSA/GetSoftwareStatementAssertion_Tests.cs)): (x-v 4, x-min-v 5) gives 200 v4. (x-v 5, x-min-v 4) gives 200 v4. (x-v 5, x-min-v 5) gives 406. (x-v 2) gives 406. (x-v foo/0/-1) gives 400.

Order of errors (derived from middleware order, UNVERIFIED by test for all cases): the version reader runs during routing. So a version error comes before authentication, industry and model validation errors.

Other rules:
- The API versioning package uses `x-v` for route selection. There is no `Accept` header versioning.
- `x-fapi-interaction-id` is not handled. The log schema has a `FapiInteractionId` column. UNVERIFIED: whether the response echoes it. I found no code that sets it.

### 6.4 Industry path parameter

Source: [CheckIndustryAttribute.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Filters/CheckIndustryAttribute.cs).
- Case-insensitive (tests send `ENERGY` and `telCO`).
- Invalid or not allowed for the version: 400 `urn:au-cds:error:cds-all:Field/Invalid`, title `Invalid Field`, detail `industry`. (CDS defines `urn:au-cds:error:cds-register:Field/InvalidIndustry`, and the mock has the constant, but it does not use it.)
- Empty industry (`/cdr-register/v1//...`) gives 404 (tests: `InlineData(null, HttpStatusCode.NotFound)`).
- The DB `IndustryType` table has rows only for banking (1), energy (2), non-bank-lending (3). `telco` (4) is valid in the path but has no data ([Repository Extensions.cs L23-L26](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/Infrastructure/Extensions.cs)).
- GetDataRecipients, GetDataRecipientsStatus and GetSoftwareProductsStatus **ignore** the industry value after validation. They return all data recipients ([DiscoveryService.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Business/DiscoveryService.cs), [RegisterStatusRepository.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/RegisterStatusRepository.cs)). v2 of the status endpoints removes non-bank-lending participations.

### 6.5 Status endpoints

Source: [RegisterStatusRepository.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/RegisterStatusRepository.cs), [StatusService.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Status.API/Business/StatusService.cs).
- DataRecipientsStatus: `data[] = { legalEntityId, status }` for every DR participation. Sort by legal entity ID string.
- SoftwareProductsStatus: `data[] = { softwareProductId, status }` for every product of every DR. Sort by product ID.
- DataHoldersStatus: `data[] = { legalEntityId, status }` for DH participations in the industry. **Only ACTIVE participations are returned.** So a DH status is always `ACTIVE` in this response. This looks like a defect. UNVERIFIED against production.
- Responses have `links.self` and an empty `meta`.
- No authentication.

### 6.6 GetDataRecipients

- Returns every DR participation, sorted by legal entity ID. Brands and products are sorted by ID ([RegisterDiscoveryRepository.cs L41-L56, L110-L124](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/RegisterDiscoveryRepository.cs)).
- Fields: `legalEntityId`, `legalEntityName`, `accreditationNumber`, `accreditationLevel` (upper-case), `logoUri`, `dataRecipientBrands[] { dataRecipientBrandId, brandName, logoUri, softwareProducts[] { softwareProductId, softwareProductName, softwareProductDescription, logoUri, status }, status }`, `status`, `lastUpdated` ([RegisterDataRecipientModel.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Business/Models/RegisterDataRecipientModel.cs)).
- v3 and v4 return the same payload. Only the allowed industry values differ.
- No pagination. `links.self` and empty `meta`.

### 6.7 GetDataHolderBrands fields

- v3 (`RegisterDataHolderBrandV2`): `dataHolderBrandId`, `brandName`, `brandGroup`, `industries` (one value, lower case), `logoUri`, `legalEntity { legalEntityId, legalEntityName, logoUri, registrationNumber, registrationDate (yyyy-MM-dd), registeredCountry, abn, acn, arbn, anzsicDivision, organisationType, status (upper case) }`, `status`, `endpointDetail { version, publicBaseUri, productBaseUri, resourceBaseUri, infosecBaseUri, extensionBaseUri, websiteUri }`, `authDetails[] { registerUType, jwksEndpoint }`, `lastUpdated`.
- v2 has no `brandGroup` and no `productBaseUri` ([Models](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Business/Models), [MappingProfile.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Business/MappingProfile.cs)).
- Brands v2/v3 also run the ADR status check on the token `client_id`. An inactive product, brand or participation gives 403 `AdrStatusNotActive` ([DiscoveryController.cs L158-L191](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Discovery.API/Controllers/DiscoveryController.cs)).

### 6.8 Error envelope

- CDS envelope: `{"errors":[{"code","title","detail","meta"?}]}` ([ResponseErrorList.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Domain/Models/ResponseErrorList.cs), [Error.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Domain/Models/Error.cs)).
- JSON is camelCase, indented, nulls omitted, enums as strings.
- Error codes and titles: [Domain/Constants.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Domain/Constants.cs).
- Unhandled exceptions: 500 `urn:au-cds:error:cds-all:GeneralError/Unexpected`, title `Unexpected Error Encountered`, detail `An unexpected exception occurred while processing the request.`
- Unknown model errors: `{"code":"400","title":"BadRequest","detail":"<key>: <message>"}` ([ModelStateErrorMiddleware.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.API.Infrastructure/Middleware/ModelStateErrorMiddleware.cs)).
- Some auth errors use non-URN codes: `{"code":"401","title":"Unauthorized","detail":"invalid_token"}` and `{"code":"403","title":"Forbidden",...}`.
- The token endpoint and admin auth use the OAuth form `{"error","error_description"}`.
- The mTLS gateway returns `text/plain` for certificate errors (400) and proxy errors (502).

### 6.9 Known deviations (summary)

From the code, issues and README:
1. No `GetDataHolderBrandsSummary` (CDS 1.36 v2).
2. `cnf.x5t#S256` holds SHA-1 upper-case hex, not base64url SHA-256 (RFC 8705).
3. The token endpoint checks the client certificate CN only, not the thumbprint.
4. Scope errors return `invalid_client`, not `invalid_scope`. Scope must be exactly `cdr-register:read`.
5. `kid` values are SHA-1 thumbprints, not RFC 7638 (issue #13; the ACCC says production also does not use RFC 7638).
6. Backends trust `X-TlsClientCertThumbprint` / `X-TlsClientCertCN` headers (spoofable, see 2.3).
7. `x-v` is optional on data-holders/status.
8. Invalid industry uses `cds-all:Field/Invalid`, not `cds-register:Field/InvalidIndustry`.
9. Invalid `page` uses `Field/Invalid`, not `Field/InvalidPage`.
10. Industry is ignored for data-recipients and DR/SP status.
11. DataHoldersStatus returns only ACTIVE participations.
12. GetDataHolderBrands does not filter by status (the tests assume ACTIVE only).
13. GetDataHolderBrands returns 403 (not in the CDS response list) and has ETag (CDS has no `If-None-Match` there).
14. The README says `oidc.issuer`; the code uses `Authorization:Issuer`.
15. The README says: "not an exact replica of the production CDR Register ... there may be slight differences" ([README.md L275-L277](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md)).
16. Issue #44 (2021): someone reported 404 instead of 401 on production for an invalid token. The ACCC could not reproduce it.

---

## 7. Seed data

### 7.1 Files and load

- Files: [Source/CDR.Register.Admin.API/Data/](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Data): `seed-data.Release.json` (84 KB), `seed-data.Development.json`, `seed-data.Container.json`.
- The files contain `//` comments (for example `"productBaseUri": null, // intentionally not provided`). Standard JSON parsers reject them. The loader uses Newtonsoft, which accepts comments. The tests strip comments first (`JsonStripComments`).
- Config: `SeedData:FilePath` and `SeedData:OverwriteExistingData` ([Admin appsettings.Release.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/appsettings.Release.json)). Release: `Data/seed-data.Release.json`, overwrite `true`. So every container start replaces the data.
- Migrations run only if connection string `Register_DBO` is set ([Admin Startup.cs L130-L183](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Admin.API/Startup.cs)).
- Load algorithm ([Repository Extensions.cs L66-L186](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/Infrastructure/Extensions.cs)): in one transaction, delete all `LegalEntity` rows (cascade), then deserialize `legalEntities` into EF entities and insert.
- Ways to load custom data ([README.md L60-L121](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/README.md)):
  1. Replace the seed file in `/app/admin/Data` and change `/app/admin/appsettings.Release.json`, then restart.
  2. `POST /admin/metadata` with the full data set.
  3. `POST /admin/metadata/data-holders` or `/data-recipients` for one record.

### 7.2 Data model (seed-data JSON, entity shape)

```
legalEntities[]
  legalEntityId (GUID), legalEntityName, logoUri, registrationNumber, registrationDate,
  registeredCountry, abn, acn, arbn, anzsicDivision, organisationTypeId (int),
  legalEntityStatusId (int), accreditationNumber, accreditationLevelId (int)
  participations[]
    participationTypeId (1=DH, 2=DR), industryId (1 banking, 2 energy, 3 NBL; null for DR), statusId
    brands[]
      brandId (GUID), brandName, brandGroup, logoUri, brandStatusId, lastUpdated
      authDetails[]      (DH only) registerUTypeId (1 = SIGNED-JWT), jwksEndpoint
      endpoint           (DH only) version, publicBaseUri, productBaseUri, resourceBaseUri,
                                   infoSecBaseUri, extensionBaseUri, websiteUri
      softwareProducts[] (DR only) softwareProductId (GUID), softwareProductName,
                                   softwareProductDescription, logoUri, sectorIdentifierUri,
                                   clientUri, tosUri, policyUri, recipientBaseUri, revocationUri,
                                   redirectUris (space-separated string), jwksUri, scope, statusId
        certificates[]  commonName, thumbprint (SHA-1 lower-case hex in seed)
```

Lookup tables ([Repository Extensions.cs L20-L60](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/Infrastructure/Extensions.cs), [Repository/Enums](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/Enums)):

| Table | Values |
|---|---|
| IndustryType | 1 banking, 2 energy, 3 non-bank-lending (telco = 4 in enum only) |
| OrganisationType | 1 SOLE_TRADER, 2 COMPANY, 3 PARTNERSHIP, 4 TRUST, 5 GOVERNMENT_ENTITY, 6 OTHER |
| ParticipationType | 1 DH, 2 DR |
| ParticipationStatus | 1 ACTIVE (both), 2 REMOVED (DH), 3 SUSPENDED (DR), 4 REVOKED (DR), 5 SURRENDERED (DR), 6 INACTIVE (both) |
| BrandStatus | 1 ACTIVE, 2 INACTIVE, 3 REMOVED |
| SoftwareProductStatus | 1 ACTIVE, 2 INACTIVE, 3 REMOVED |
| RegisterUType | 1 SIGNED-JWT |
| AccreditationLevel | 0 SPONSORED, 1 UNRESTRICTED |

Temporal tables: `AuthDetail`, `Brand`, `Endpoint`, `LegalEntity`, `Participation`, `SoftwareProduct`, `SoftwareProductCertificate` are SQL Server system-versioned tables ([RegisterDatabaseContext.cs L59-L73](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/Infrastructure/RegisterDatabaseContext.cs)). A test suite checks them ([US29068_TemporalTableTests.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/TemporalTables/US29068_TemporalTableTests.cs)).

### 7.3 Content of `seed-data.Release.json`

- 36 legal entities: 26 data holders and 10 data recipients. Each legal entity has one participation. All legal entities, participations, brands and products are ACTIVE (status 1).
- Data holders: 1 energy legal entity with 2 brands (`Mock Data Holder (Energy)`, `Cut Price Energy`), 1 non-bank-lending (`Mock Data Holder (Non-Bank Lending)`), and 24 banking legal entities (`Mock Data Holder (Banking)` and many "dummy data holder -- do not use" brands).
- Data recipients: 10 legal entities, for example `Mock Energy Company` (`MyEnergyHelper`, product `4754cf30-6cc2-4a9a-a9e1-4fb8f6d42120`), `Mock Software Company` (`MyBudgetHelper`), `FintechX` (4 products), and others.
- The test constants use brand `20C0864B-CEEF-4DE0-8944-EB0962F825EB` and product `86ECB655-9EBA-409C-9BE3-59E7ADF7080D`. The loopback default product is `6F7A1B8E-8799-48A8-9011-E3920391F713`.
- Endpoint URIs point to Docker host names (`https://mock-data-holder:8000`, `https://mock-data-holder-energy:8100`, `https://mock-data-recipient:9001`).
- The certificate record is `MockDataRecipient` / `f0e5146a51f16e236844cf0353d791f11865e405` (the shipped `client.pem`).

---

## 8. How to run it

- Image: `consumerdataright/mock-register` on Docker Hub. Tags on 2026-09-25: `latest`, `2`, `2.2`, `2.2.8`, `main` (2026-07-15), `develop` (2026-07-07), `2.2.6`, `2.2.3`, `2.2.1`, `2.2.0`, `2.1`, `2.1.0`, `1`, `1.3`, `1.3.2`. Architectures amd64 and arm64. About 11,900 pulls. No GHCR image found.
- The image is based on `mcr.microsoft.com/dotnet/aspnet:8.0` (Debian). It runs as `appuser` (uid 1000) under supervisord.
- Pin a version tag (for example `2.2.8`), and pin by digest for repeatable oracle runs.
- Database: SQL Server 2022. Host name `mssql`, user `SA`, password `Pa{}w0rd2019`, database `cdr-register` (from every `appsettings.Release.json`). The compose sets `ACCEPT_EULA` from `ACCEPT_MSSQL_EULA`, which you must set to `Y`.
- Minimum compose ([docker-compose.yml](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/docker-compose.yml)): services `mock-register` (ports 7000-7006, `ASPNETCORE_ENVIRONMENT=Release`, health check on 7000/health) and `mssql` (container name `sql1`, but the connection strings use host `mssql`, so the service name matters).
- Environment overrides (from [docker-compose.IntegrationTests.yml](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/docker-compose.IntegrationTests.yml)): `PublicHostName`, `SecureHostName`, `BasePath`, `BasePathExpression`, `Authorization__Issuer`, `Authorization__ClientId`, `Authorization__Scope`, `Certificate__CertThumbprintNameHttpHeaderName`, `Certificate__CertCommonNameHttpHeaderName`, Serilog and logging levels. ASP.NET config uses `__` for nesting.
- Host names: clients must resolve `mock-register` (issuer, token audience) and the mock DR JWKS host. On a host, add `127.0.0.1 mock-register` and `127.0.0.1 mssql` to `/etc/hosts` ([Help/testing/HELP.md](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Help/testing/HELP.md)).
- Trust: clients must trust `ca.pem` or turn off server validation.
- Logging: Serilog to console, file in `/tmp`, and SQL tables `LogEvents-*`. The request/response logger writes full request and response bodies and headers to SQL table `LogEvents-RequestResponse`. OpenTelemetry export works with standard `OTEL_EXPORTER_OTLP_*` variables (CHANGELOG 2.2.2). Application Insights is optional (2.2.7).
- Full ecosystem: [docker-compose.Ecosystem.yml](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/docker-compose.Ecosystem.yml) runs register, mock-data-holder (8000-8005), mock-data-holder-energy (8100-8105), mock-data-recipient (9001) and SQL Server. `consumerdataright/mock-data-holder-energy` comes from an archived repo.

---

## 9. Test suites

### 9.1 Mock Register integration tests

- Project: [Source/CDR.Register.IntegrationTests](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests). C#, xUnit, FluentAssertions, .NET 8.
- Run in Docker: `docker compose -f docker-compose.IntegrationTests.yml up -d --build`. The `integration-tests` Dockerfile stage runs `dotnet test --filter ${TEST_FILTER}`. The default filter is `Category!=CTSONLY`. TRX results go to `_temp/.../testresults` ([Dockerfile L25-L27](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/Dockerfile), [Help/testing/HELP.md](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Help/testing/HELP.md)).
- Size: about 200 test methods (`[Fact]` and `[Theory]` attributes) and about 890 `[InlineData]` rows.

Test areas:

| Folder / file | What it covers |
|---|---|
| `API/Discovery/*DataHolderBrands*` | v2/v3, industries, paging, `updated-since`, bad params, x-v/x-min-v, ETag, ADR status 403 |
| `API/Discovery/*DataRecipients*` | v3/v4, industries, x-v, ETag |
| `API/Status/*` | three status endpoints, versions, industries, x-v |
| `API/SSA/*` | SSA claims and signature, status 403, 401 cases, 404 cases, x-v matrix |
| `API/Update/*` | Admin upsert of data holders and data recipients, validation errors (`ExpectedErrors.cs`) |
| `IdentityServer/US12665_*` | OIDC discovery document |
| `IdentityServer/US14045_*` | token endpoint validation rules (table in 4.3), audience rules, `cnf` |
| `Gateway/US12841_Gateway_MTLS_Tests.cs` | mTLS: expired cert, server cert as client cert, wrong cert for token (401), no token (401), status changes (403) |
| `Miscellaneous/US12677_RegisterEnrolment_Tests.cs` | `GET/POST /admin/metadata` round trip, empty and invalid bodies (400) |
| `Miscellaneous/US50483_DynamicUrl_Tests.cs` | CTS dynamic base paths, spoofed certificate headers on direct backend calls |
| `TemporalTables/*` | SQL Server temporal history rows |

Structure:
- [BaseTest.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/BaseTest.cs) holds URLs, certificates, expected error JSON and SQL helpers.
- [Fixtures/TestFixture.cs](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/Fixtures/TestFixture.cs) seeds data with `POST /admin/metadata` (the seed file for `ASPNETCORE_ENVIRONMENT`). Then it runs SQL: `update softwareproduct set jwksuri = '{Admin_BaseURL}/loopback/MockDataRecipientJwks'`.
- JSON checks use `FluentAssertions.Json` `BeEquivalentTo` (order and whitespace ignored).

Base URLs are configurable ([appsettings.Release.json](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/appsettings.Release.json)): `TLS_BaseURL`, `MTLS_BaseURL`, `Admin_BaseURL`, `SSA_Downstream_BaseUrl`, `Discovery_Downstream_BaseUrl`, `Status_Downstream_BaseUrl`, `IdentityProvider_Downstream_BaseUrl`, and `ConnectionStrings:DefaultConnection`. Environment variables override them.

Can they run against another Register? Not without work. The coupling is strong:
1. **Direct SQL access.** 18 test files use the SQL connection. Tests read the expected results from SQL Server tables with EF Core (`RegisterDatabaseContext`) or raw SQL, and they change statuses with SQL `UPDATE` (`SetSoftwareProductStatusId`, `SetBrandStatusId`, `SetParticipationStatusId`, `jwksuri` patch). A Go Register would need the same SQL Server schema, or you must rewrite these helpers to use an admin API.
2. **Admin API.** The fixture needs `POST /admin/metadata` with the seed-data format.
3. **Loopback JWKS.** The token tests need `/loopback/MockDataRecipientJwks` on `Admin_BaseURL`.
4. **Implementation details asserted:** SHA-1 hex in `cnf`, exact error descriptions, `x-v` defaults, `x-min-v` rules, CN-only certificate match.
5. **Backend ports and headers:** CTS and dynamic URL tests call ports 7002-7005 directly with `X-TlsClientCertThumbprint` and `X-TlsClientCertCN` (CTS-only ones are filtered out by default).
6. **Host name:** the tests expect the OIDC `issuer` to equal `{TLS_BaseURL}/idp`.
7. **Environment:** .NET 8 SDK, Windows-oriented docs, test certificates copied into the test project.

### 9.2 mock-solution-test-automation (shared NuGet package)

- Package `ConsumerDataRight.ParticipantTooling.MockSolution.TestAutomation`, v3.1.0 (2026-07-08). MIT. 132 C# files. xUnit and Playwright.
- The README says it is "tightly coupled with CDR Mock Solutions" and "not intended to be used as a stand alone testing solution".
- The mock Data Holder and Authorisation Server test suites use it. The mock Register suite does not.
- It calls the Register token endpoint and SSA endpoint with config `REGISTER_MTLS_URL`, `REGISTER_MTLS_TOKEN_URL` ([RegisterSSAService.cs](https://github.com/ConsumerDataRight/mock-solution-test-automation/blob/ce77f50da4eb36577ff4d44a583751fd63d0935b/Source/ConsumerDataRight.ParticipantTooling.MockSolution.TestAutomation/Services/RegisterSSAService.cs), [ApiServiceDirector.cs L93](https://github.com/ConsumerDataRight/mock-solution-test-automation/blob/ce77f50da4eb36577ff4d44a583751fd63d0935b/Source/ConsumerDataRight.ParticipantTooling.MockSolution.TestAutomation/Services/ApiServiceDirector.cs)).
- It also writes to the Register database with `REGISTER_CONNECTIONSTRING`: it sets `softwareproduct.scope`, `redirecturis` and `jwksuri` with SQL ([BaseFixture.cs L68-L81](https://github.com/ConsumerDataRight/mock-solution-test-automation/blob/ce77f50da4eb36577ff4d44a583751fd63d0935b/Source/ConsumerDataRight.ParticipantTooling.MockSolution.TestAutomation/Fixtures/BaseFixture.cs), [Helpers.cs L225-L285](https://github.com/ConsumerDataRight/mock-solution-test-automation/blob/ce77f50da4eb36577ff4d44a583751fd63d0935b/Source/ConsumerDataRight.ParticipantTooling.MockSolution.TestAutomation/Helpers.cs)).
- Result: to put a Go Register under the mock Data Holder tests, the Go Register must either use the same SQL Server tables or offer an equal admin API and you must patch the helpers.

---

## 10. Activity and maintenance

- Latest release: v2.2.8 on 2026-07-15 ("Updated NuGet packages"). Earlier: 2.2.7 (2026-06-17), 2.2.6 (2026-05-13, non-bank-lending), 2.2.5 (2026-03-18), 2.2.4 (2026-02-25), 2.2.3 (2025-12-03), 2.2.2 (2025-10-15), 2.2.1 (2025-06-19), 2.2.0 (2025-03-19), 2.1.0 (2024-08-16), 2.0.0 (2024-06-12, .NET 8, Polyglot notebook) ([CHANGELOG.md](https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CHANGELOG.md)).
- Cadence: about one release every one to three months. Most releases are dependency and vulnerability updates. Feature changes follow CDS changes (non-bank-lending in 2.2.6).
- The ACCC develops in an internal Azure DevOps pipeline (`.azuredevops` folder). GitHub gets release PRs ("vX.Y.Z release", "Sync commit history"). There are no community PRs in the last 15 PRs.
- Issues: all issues are closed. The newest issue is #80 from 2023-04-20. Response on issues was active in 2021-2023.
- Related repos had releases on the same day (2026-07-15): mock-data-holder v4.0.5, mock-data-recipient v3.0.5, authorisation-server v3.1.0.
- The ACCC also runs a hosted "CDR Sandbox" built on these mocks (https://cdrsandbox.gov.au/). The `sandbox` repo is only for issues.

---

## 11. Items for the Go Register requirements and oracle use

These are conclusions from the facts above, not facts from the repo.

1. Decide per deviation (6.9) whether the Go Register copies the ACCC mock or follows CDS/RFC. Record each choice. Example: follow RFC 8705 for `cnf` (SHA-256 base64url). Then an oracle diff on `cnf` is an expected difference.
2. The oracle is good for: OIDC metadata, x-v/x-min-v matrix, error envelopes, pagination, industry validation, SSA claim set, token endpoint error cases, ETag.
3. The oracle is weak for: GetDataHolderBrandsSummary (absent), status filtering (defects), cert binding (SHA-1, CN-only, header trust).
4. Write your own black-box suite in Go. Use only HTTP, and use the shipped certificates. Give both Registers the same seed data through an admin interface. The ACCC suite cannot do this because it depends on SQL Server.
5. To change participant status in the oracle without SQL, use `POST /admin/metadata/data-recipients` with a new `status`. UNVERIFIED: whether upsert changes participation status the same way as the SQL helpers.
6. Normalise volatile fields before a diff: `jti`, `iat`, `exp`, `nbf`, signatures, `kid`, `links` hosts, `ETag` (it depends on the links), `lastUpdated` after upserts.
7. Pin the image by digest. The ACCC updates `latest` and `main` often.
8. The shipped certificates expire on 2028-06-14 (leaf) and 2031-04-03 (CA).
