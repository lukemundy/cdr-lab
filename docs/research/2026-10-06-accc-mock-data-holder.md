# Research: ACCC CDR Mock Data Holder and related reference mocks (behaviour, seed data, test peer use)

- Research date: 2026-10-06
- Purpose: Give the data to (a) write requirements for a new, independent mock banking data holder in Go, (b) use the ACCC mock as a reference and as a test peer, and (c) know where the ACCC mock does not follow the standards.
- This note extends section 2.6 of the [Auth servers note](2026-09-25-auth-servers.md). It does not repeat that section.
- Check the sources again before you depend on a fact; the repos change.
- The upstream seed data is invented data. This note does not copy names, login identifiers or account numbers from it.

## Versions examined

| Repo | Version | Commit | Date |
|---|---|---|---|
| ConsumerDataRight/mock-data-holder | v4.0.5 (latest release, tag = `main`) | `8fdd4392c6f6072b32782dc5cde2e467a0959549` | release 2026-07-15 |
| ConsumerDataRight/authorisation-server | v3.1.0 (latest release, tag = `main`) | `c40a88a1e50451f4c3e163bb791af6aa5af2b7b0` | release 2026-07-15 |
| ConsumerDataRight/mock-data-recipient | v3.0.5 (latest release, tag = `main`) | `3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7` | release 2026-07-15 |
| ConsumerDataStandardsAustralia/mock-data-holder-nodejs | `master`; latest release v1.2.0 (2025-02-20) is an older commit | `7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05` | commit 2025-09-02 |
| ConsumerDataStandardsAustralia/testdata-cli | `master`, `package.json` version 1.8.0; latest GitHub release v1.6.2 (2024-08-30) | `c41ba785eefaaadced5a0c2b4ca66627bdd37986` | commit 2025-09-21 |
| ConsumerDataStandardsAustralia/standards | tag `v1.36.0` (OpenAPI files only) | tag | - |

Repository metadata (licence, releases, tags, last push) came from the GitHub REST API (`gh api repos/<owner>/<repo>`, `/releases`, `/tags`) on 2026-10-06. I read the code in shallow clones. I did not build or run the code. Thus each statement about run-time behaviour is derived from the source code.

Code links in this file point to these commits:

- mock-data-holder: https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549
- authorisation-server: https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0
- mock-data-recipient: https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7
- mock-data-holder-nodejs: https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05
- testdata-cli: https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986

## Primary sources

- Mock Data Holder README: https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md
- Mock Data Holder CHANGELOG: https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CHANGELOG.md
- Mock Data Holder container guide: https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Help/container/HELP.md
- Mock Data Holder certificate docs: https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CertificateManagement/README.md
- Authorisation Server README: https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/README.md
- Authorisation Server CHANGELOG: https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/CHANGELOG.md
- Mock Data Recipient README: https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/README.md
- DSB Mock Data Holder (NodeJS) README: https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/README.md
- DSB Test Data CLI README: https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986/README.md
- CDS 1.36.0 banking OpenAPI: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_banking.json
- CDS 1.36.0 common OpenAPI: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_common.json
- CDS 1.36.0 admin OpenAPI: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_admin.json
- CDS 1.36.0 archive: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/

Note about the CDS links: on 2026-10-06 the archive URL returned a short page that sends the browser to https://consumerdatastandardsaustralia.github.io/standards/. That page states "version 1.36.0 of the high level standards". I confirmed each CDS anchor and each CDS quotation in this note against that page. The links use the archive URL because it will stay at 1.36.0.

---

## 1. Licences, versions and maintenance

### 1.1 Licences

| Repo | Licence | Note |
|---|---|---|
| mock-data-holder | MIT, "Copyright (c) 2021 Commonwealth of Australia" | [LICENSE](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/LICENSE). No separate licence for the seed data. |
| authorisation-server | MIT | GitHub licence metadata, [LICENSE](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/LICENSE) |
| mock-data-recipient | MIT | GitHub licence metadata, [LICENSE](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/LICENSE) |
| mock-data-holder-nodejs | Not clear | No `LICENSE` file, and GitHub shows no licence. The README says "released under the MIT License" but links to the LICENSE file of a different repository, and also says "on a non-commercial basis" ([README.md L177-L187](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/README.md#L177-L187)). |
| testdata-cli | Not clear | No `LICENSE` file, and GitHub shows no licence. `package.json` has `"license": "MIT"` ([package.json](https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986/package.json)). The README has the same text as above ([README.md L195-L205](https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986/README.md#L195-L205)). |

- The compose files use the Microsoft SQL Server image. You must accept the Microsoft EULA ([README.md L136](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L136)).
- MIT lets us read and copy the behaviour, the certificates and the seed data, if we keep the MIT notice. The DSB data files have no clear licence. Treat them as a model only, unless the DSB clarifies the licence.

### 1.2 Stated CDS version

- The mock-data-holder README states alignment to CDS **v1.34.0**, "in particular FAPI 1.0 Migration Phase 3", and that it passed v5.2.0 of the Conformance Test Suite for Data Holders ([README.md L18-L22](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L18-L22)). The lab target is 1.36.0.
- The authorisation-server README states CDS v1.34.0 and CTS for Data Recipients v4.4.0 ([README.md L19-L20](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/README.md#L19-L20)).
- The mock-data-recipient README states CDS v1.34.0 ([README.md L16](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/README.md#L16)).
- The newest DSB data file states `"standardsVersion": "1.33.0"` ([all-data-1.33.0.json](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/load-test-data/input/1.33.0/all-data/all-data-1.33.0.json)).
- The README says: "The Mock Data Holder is provided as a development tool only. It conforms to the Consumer Data Standards." ([README.md L251-L252](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L251-L252)). Section 6 shows where it does not conform.

### 1.3 Maintenance activity

- mock-data-holder releases: v4.0.5 (2026-07-15), v4.0.3 (2025-12-10), v4.0.1 (2025-06-26), v4.0.0 (2025-04-03), v3.1.0 (2024-08-22) (GitHub releases API). The changelog dates are earlier, for example 4.0.5 on 2026-07-08 ([CHANGELOG.md L8-L10](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CHANGELOG.md#L8-L10)).
- Each release after 4.0.0 is a package update, a rebuild or a build-warning fix ([CHANGELOG.md L8-L27](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CHANGELOG.md#L8-L27)). The last change of API behaviour is 3.0.0 (2024-06-12): "Changed Banking Get Accounts API to only support version 2" and the removal of Get Metrics versions below 4 ([CHANGELOG.md L42-L51](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CHANGELOG.md#L42-L51)). 4.0.0 removed the OIDC Hybrid Flow code ([CHANGELOG.md L29-L36](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CHANGELOG.md#L29-L36)).
- authorisation-server 3.1.0 moved the UI from Create React App to Vite ([CHANGELOG.md L8-L13](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/CHANGELOG.md#L8-L13)).
- Result: the ACCC maintains the mocks for security patches. The API behaviour did not follow CDS 1.35.0 or 1.36.0.
- DSB repos: the last commit on `master` is 2025-09-02 (mock-data-holder-nodejs) and 2025-09-21 (testdata-cli). The GitHub API shows a push to mock-data-holder-nodejs on 2026-09-03. That push is not on `master`. UNVERIFIED: its branch and content.

---

## 2. Architecture of the ACCC Mock Data Holder

### 2.1 Components and ports (banking profile)

All components run in one container under `supervisord` ([supervisord.conf](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/supervisord.conf)). The image also contains the authorisation-server: the build copies the `cdr-auth-server` checkout into the image ([Dockerfile L29-L63](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L29-L63), [docker.yml L93-L107](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/.github/workflows/docker.yml#L93-L107), [README.md L51](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L51)).

| Component | Project | Port | Published by compose | Role |
|---|---|---|---|---|
| Public API | `CDR.DataHolder.Public.API` | 8000 | yes | Get Status, Get Outages (TLS, no client certificate) |
| Identity provider | `CdrAuthServer` (authorisation-server) | 8001 | yes | OIDC discovery, JWKS, `/connect/authorize`, and the backend of each InfoSec endpoint |
| mTLS gateway | `CDR.DataHolder.Shared.API.Gateway.mTLS` | 8002 | yes | Ocelot reverse proxy. It validates the client certificate and routes InfoSec, resource and admin requests |
| Banking Resource API | `CDR.DataHolder.Banking.Resource.API` | 8003 | no | Get Accounts, Get Transactions For Account |
| Admin API | `CDR.DataHolder.Admin.API` | 8004 | no | Get Metrics |
| Manage API | `CDR.DataHolder.Manage.API` | 8005 | yes | Not a CDS API. Migrations, seed data, `POST /manage/metadata` |
| Common API | `CDR.DataHolder.Common.API` | 8006 | **yes** | Get Customer |
| Auth UI | `CdrAuthServer.UI` (React, served by nginx) | 3000 | yes | Login, OTP, account selection, confirmation |
| Repository | SQL Server 2022 | 1433 | separate container | Databases `cdr-mdh` and `cdr-mdh-auth-server` |

Sources: [README.md L164-L193](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L164-L193), [docker-compose.yml L25-L47](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/DockerCompose/docker-compose.yml#L25-L47), [gateway-config.banking.json](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/gateway-config.banking.json), [Dockerfile L225-L271](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L225-L271).

Facts:
- The README says that the Resource API is "Accessed via the mTLS Gateway" ([README.md L182-L185](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L182-L185)). But the compose file publishes port 8006 (Common API) directly ([docker-compose.yml L28-L34](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/DockerCompose/docker-compose.yml#L28-L34)). See 6.3.
- Tech stack: C#, .NET 8, Ocelot, SQL Server, xUnit, Playwright ([README.md L225-L233](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L225-L233)).
- Start order: the Manage API starts first. The Admin API, the Public API, the Banking Resource API and the auth server wait until `https://localhost:8005/health` returns OK ([supervisord.conf](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/supervisord.conf), [wait-until-manage-healthy-then-start.sh](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/wait-until-manage-healthy-then-start.sh)).
- The image also starts the Energy Resource API. There is one image build stage for each industry ([Dockerfile L171-L274](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L171-L274)).
- Clients must resolve the host names `mock-data-holder`, `mock-register` and `mock-data-recipient` ([Help/container/HELP.md L86](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Help/container/HELP.md#L86)). The issuer is `https://mock-data-holder:8001` and the secure base URI is `https://mock-data-holder:8002` ([Dockerfile L228-L230](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L228-L230)).

### 2.2 mTLS gateway routes (port 8002)

Source: [gateway-config.banking.json](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/gateway-config.banking.json). Each route sets `DangerousAcceptAnyServerCertificateValidator: true`, so the gateway does not validate backend certificates.

| Method | Upstream path on 8002 | Downstream |
|---|---|---|
| GET | `/connect/authorise` (this spelling) | auth server 8001 `/connect/authorize` |
| GET | `/health/status` | auth server 8001 |
| POST | `/connect/register` | auth server 8001 |
| GET, PUT, DELETE | `/connect/register/{clientId}` | auth server 8001 |
| POST | `/connect/par`, `/connect/token`, `/connect/revocation`, `/connect/introspect`, `/connect/arrangements/revoke` | auth server 8001 |
| GET, POST | `/connect/userinfo` | auth server 8001 |
| GET | `/cds-au/v1/common/customer` | Common API 8006 |
| GET | `/cds-au/v1/{industry}/accounts` | Banking Resource API 8003 |
| GET | `/cds-au/v1/{industry}/accounts/{accountId}/transactions` | Banking Resource API 8003 |
| GET | `/cds-au/v1/admin/metrics` | Admin API 8004 |
| POST | `/cds-au/v1/admin/register/metadata` | auth server 8001 |

The gateway has no route for the Manage API, the Public API, the discovery document or the JWKS. Clients call those on ports 8005, 8000 and 8001.

### 2.3 How mTLS terminates

- Kestrel on the gateway uses TLS 1.2 only and `ClientCertificateMode.RequireCertificate` ([Program.cs L74-L87](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/Program.cs#L74-L87)).
- Certificate authentication: `ValidateCertificateUse`, `ValidateValidityPeriod`, `RevocationMode = NoCheck` ([Startup.cs L35-L66](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/Startup.cs#L35-L66)). A custom chain build trusts only the configured root CA file `Certificates/ca.crt`, also with `RevocationMode = NoCheck` ([CertificateValidator.cs L33-L54](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/Certificates/CertificateValidator.cs#L33-L54), [appsettings.json](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/appsettings.json)).
- A certificate error gives 400 `text/plain`. Other gateway errors give 502 `text/plain` ([Startup.cs L76-L96](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/Startup.cs#L76-L96)).
- The gateway **sets** (it does not append) these request headers for the backends: `X-TlsClientCertThumbprint` = `X509Certificate2.Thumbprint` (the SHA-1 hash, hexadecimal), `X-TlsClientCertCN`, `X-TlsClientCert` (the certificate, base64) and `X-Forwarded-Host` ([Startup.cs L102-L123](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Gateway.mTLS/Startup.cs#L102-L123)). Thus a client cannot change these headers through the gateway. The mock Register gateway appends the headers (see the [mock Register note](2026-09-25-accc-mock-register.md), 2.3).
- The backends do not see the TLS client certificate. They trust the headers (6.3).

### 2.4 Storage

- SQL Server, with Entity Framework Core. The banking tables are `Customer`, `Person`, `Organisation`, `Account`, `Transaction` and two log tables ([BankingDataHolderDatabaseContext.cs L23-L35](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/Infrastructure/BankingDataHolderDatabaseContext.cs#L23-L35)).
- The auth server has its own database with clients, grants, software products and a token blocklist (identifier `AddToBlacklist`) ([TokenService.cs L715-L724](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L715-L724)).
- Connection strings with the `SA` user and its password are in the Dockerfile ([Dockerfile L236-L254](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L236-L254)).
- Serilog writes logs to the console, to files in `/tmp` and to SQL tables ([Manage appsettings.Release.json L8-L38](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Manage.API/appsettings.Release.json#L8-L38)). The Admin API logs full bearer tokens at debug level ([AdminController.cs L126-L130](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Controllers/AdminController.cs#L126-L130)).

### 2.5 Connection to a Register

The data holder uses the Register in four places. Each is a configuration value.

| Use | Setting | Default in the banking image | Source |
|---|---|---|---|
| SSA signature check in DCR | `CdrAuthServer:CdrRegister:SsaJwksUri` | `https://mock-register:7000/cdr-register/v1/jwks` | [appsettings.Release.json L165-L170](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.Release.json#L165-L170), [ClientRegistrationValidator.cs L324-L342](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/ClientRegistrationValidator.cs#L324-L342) |
| Data recipient statuses | `CdrAuthServer:CdrRegister:GetDataRecipientsEndpoint` and `Version` | `https://mock-register:7000/cdr-register/v1/all/data-recipients`, `x-v: 3` | [appsettings.Release.json L165-L170](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.Release.json#L165-L170), [RegisterClientService.cs L16-L29](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/RegisterClientService.cs#L16-L29) |
| Register as an OAuth client (`client_id` `cdr-register`) | JWKS URI = the SSA JWKS URI, written by a database migration | same URL | [V003 migration L37-L45](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.Repository/Migrations/20230130011835_V003-RegisterClientSeeddata.cs#L37-L45) |
| Get Metrics with a Register self-signed JWT | `RegisterJwksUri` | `https://mock-register:7000/cdr-register/v1/jwks` | [Dockerfile L263](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L263), [AdminController.cs L171-L219](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Controllers/AdminController.cs#L171-L219) |

What it polls, and how often:
- **The data holder image polls nothing.** The auth server gets Get Data Recipients only when a caller sends `POST /cds-au/v1/admin/register/metadata` with `action` = `REFRESH` ([AdminController.cs L44-L98](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AdminController.cs#L44-L98)). It then deletes its software product table and inserts the new list ([AdminController.cs L85-L93](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AdminController.cs#L85-L93)).
- The authorisation-server repository has a separate Azure Function, `GetDataRecipients`. It runs on a timer, gets an access token as the client `cdr-register` with the scope `admin:metadata:update`, and calls the Metadata Update endpoint ([GetDataRecipientsFunction.cs L33-L34](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.GetDataRecipients/GetDataRecipientsFunction.cs#L33-L34), [local.settings.json L7-L13](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.GetDataRecipients/local.settings.json#L7-L13)). The schedule is `0-59 * * * *` ([Dockerfile.get-data-recipients L43](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/Dockerfile.get-data-recipients#L43)). UNVERIFIED: the exact period that the Azure Functions runtime gives to this expression. The mock-data-holder Dockerfile does not copy this function into the image ([Dockerfile L55-L88](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L55-L88)).
- The auth server does not call Get Data Recipients Statuses or Get Software Products Statuses. I found no reference to those endpoints in its configuration ([CdrRegisterConfiguration.cs](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Configuration/CdrRegisterConfiguration.cs)).
- A database migration seeds the software product table with 4 software products, including the product of the mock data recipient ([SoftwareProductConfiguration.cs L17-L22](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.Repository/Configuration/SoftwareProductConfiguration.cs#L17-L22)). Thus the mock ecosystem works with no refresh.
- HTTP clients do not validate server certificates by default (`EnableServerCertificateValidation: false`) ([appsettings.json L75](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L75)).

---

## 3. Endpoints

### 3.1 CDS endpoints that the mock implements

| Method | Path | Port | Version served | CDS 1.36.0 version | Auth | Source |
|---|---|---|---|---|---|---|
| GET | `/cds-au/v1/banking/accounts` | 8002 | **2 only** | 3 | mTLS, bearer, `bank:accounts.basic:read` | [ResourceController.cs L57-L111](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Controllers/ResourceController.cs#L57-L111) |
| GET | `/cds-au/v1/banking/accounts/{accountId}/transactions` | 8002 | **1 only** | 2 | mTLS, bearer, `bank:transactions:read` | [ResourceController.cs L113-L189](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Controllers/ResourceController.cs#L113-L189) |
| GET | `/cds-au/v1/common/customer` | 8002 | 1 | 1 | mTLS, bearer, `common:customer.basic:read` | [CommonController.cs L38-L72](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Common/CDR.DataHolder.Common.API/Controllers/CommonController.cs#L38-L72) |
| GET | `/cds-au/v1/discovery/status` | 8000 | 1 | 1 | none | [DiscoveryController.cs L21-L29](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Public.API/Controllers/DiscoveryController.cs#L21-L29) |
| GET | `/cds-au/v1/discovery/outages` | 8000 | 1 | 1 | none | [DiscoveryController.cs L31-L39](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Public.API/Controllers/DiscoveryController.cs#L31-L39) |
| GET | `/cds-au/v1/admin/metrics` | 8002 | 4 and 5 | 5 | bearer: a data holder access token with `admin:metrics.basic:read`, or a Register self-signed JWT | [AdminController.cs L34-L68](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Controllers/AdminController.cs#L34-L68) |
| POST | `/cds-au/v1/admin/register/metadata` | 8002 | 1 | 1 | bearer with `admin:metadata:update` | [AdminController.cs L44-L50](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AdminController.cs#L44-L50) |

The CDS versions come from the `x-version` field of each operation in the CDS 1.36.0 OpenAPI files (Primary sources).

Facts:
- Get Status and Get Outages return a static JSON file with the host name replaced ([status.json](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Public.API/Data/status.json), [outages.json](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Public.API/Data/outages.json)). The outage list has 3 fixed outages in 2021.
- Get Metrics returns a static JSON file, with `requestTime` and `links.self` replaced ([AdminController.cs L243-L248](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Controllers/AdminController.cs#L243-L248), [metrics-v5.json](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Data/metrics-v5.json)). The mock does not measure anything.
- The response models: an account has `accountId`, `creationDate`, `displayName`, `nickname`, `openStatus`, `isOwned`, `accountOwnership`, `maskedNumber`, `productCategory`, `productName` ([BankingAccountV2.cs](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Business/Models/BankingAccountV2.cs)). A transaction has the fields of [AccountTransactionModel.cs](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Business/Models/AccountTransactionModel.cs).

### 3.2 InfoSec endpoints (authorisation-server)

The Auth servers note lists these endpoints with the short paths of the README. The paths in the code are different:

| Method | Path | Called on | Source |
|---|---|---|---|
| GET | `/.well-known/openid-configuration` | 8001 | [DiscoveryController.cs L20-L21](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/DiscoveryController.cs#L20-L21) |
| GET | `/.well-known/openid-configuration/jwks` | 8001 | [JwksController.cs L29-L31](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/JwksController.cs#L29-L31) |
| POST | `/connect/par` | 8002 | [PushedAuthorisationController.cs L33-L34](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/PushedAuthorisationController.cs#L33-L34) |
| GET | `/connect/authorize` | 8001 | [AuthorisationController.cs L42-L43](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AuthorisationController.cs#L42-L43) |
| POST | `/connect/token` | 8002 | [TokenController.cs L31-L32](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/TokenController.cs#L31-L32) |
| GET, POST | `/connect/userinfo` | 8002 | [UserInfoController.cs L35-L37](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/UserInfoController.cs#L35-L37) |
| POST | `/connect/introspect` | 8002 | [IntrospectionController.cs L38-L39](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/IntrospectionController.cs#L38-L39) |
| POST | `/connect/revocation` | 8002 | [RevocationController.cs L27-L28](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/RevocationController.cs#L27-L28) |
| POST | `/connect/arrangements/revoke` | 8002 | [ArrangementRevocationController.cs L35-L36](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/ArrangementRevocationController.cs#L35-L36) |
| POST, GET, PUT, DELETE | `/connect/register`, `/connect/register/{clientId}` | 8002 | [RegistrationController.cs L38-L144](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/RegistrationController.cs#L38-L144) |

The default endpoint values of the discovery document are in [ConfigExtensions.cs L55-L64](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Extensions/ConfigExtensions.cs#L55-L64). The discovery document states `require_pushed_authorization_requests: true`, `response_modes_supported: ["jwt"]`, `request_parameter_supported: false` and `tls_client_certificate_bound_access_tokens: true` ([DiscoveryController.cs L44-L51](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/DiscoveryController.cs#L44-L51)). Defaults: `code_challenge_methods_supported` `["S256"]`, `subject_types_supported` `["pairwise"]`, `acr` `urn:cds.au:cdr:2`, `token_endpoint_auth_methods_supported` `["private_key_jwt"]` ([ConfigExtensions.cs L66-L125](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Extensions/ConfigExtensions.cs#L66-L125)).

### 3.3 CDS 1.36.0 endpoints that the mock does not implement

The mock-data-holder has only the controllers in 3.1 (search for `*Controller.cs` in `Source/` at the commit). Thus it does not implement these CDS 1.36.0 operations:

Banking (17 of 19 operations are absent):

| CDS operation | Method and path | CDS version |
|---|---|---|
| Get Bulk Balances | GET `/banking/accounts/balances` | 2 |
| Get Balances For Specific Accounts | POST `/banking/accounts/balances` | 1 |
| Get Account Balance | GET `/banking/accounts/{accountId}/balance` | 1 |
| Get Account Detail | GET `/banking/accounts/{accountId}` | 5 |
| Get Transaction Detail | GET `/banking/accounts/{accountId}/transactions/{transactionId}` | 3 |
| Get Direct Debits For Account | GET `/banking/accounts/{accountId}/direct-debits` | 1 |
| Get Bulk Direct Debits | GET `/banking/accounts/direct-debits` | 2 |
| Get Direct Debits For Specific Accounts | POST `/banking/accounts/direct-debits` | 1 |
| Get Scheduled Payments for Account | GET `/banking/accounts/{accountId}/payments/scheduled` | 2 |
| Get Scheduled Payments Bulk | GET `/banking/payments/scheduled` | 3 |
| Get Scheduled Payments For Specific Accounts | POST `/banking/payments/scheduled` | 2 |
| Get Instalment Plans for Account | GET `/banking/accounts/{accountId}/payments/plans` | 1 |
| Get Instalment Plans Bulk | GET `/banking/accounts/payments/plans` | 1 |
| Get Payees | GET `/banking/payees` | 2 |
| Get Payee Detail | GET `/banking/payees/{payeeId}` | 2 |
| **Get Products** | GET `/banking/products` | 5 |
| **Get Product Detail** | GET `/banking/products/{productId}` | 7 |

Common: Get Customer Detail (GET `/common/customer/detail`, version 2) is absent.

Admin: both CDS admin operations are present (3.1).

Source for the CDS lists: the `paths` of `cds_banking.json`, `cds_common.json` and `cds_admin.json` at the tag `v1.36.0` (Primary sources).

A request to an absent path through the gateway has no Ocelot route. UNVERIFIED: the status and body that Ocelot returns for it. CDS defines `urn:au-cds:error:cds-all:Resource/NotImplemented` for this case ([CDS error 404](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-resource-not-implemented)).

---

## 4. Seed data

### 4.1 Files

| File | Size | Use |
|---|---|---|
| [`Source/Shared/CDR.DataHolder.Manage.API/Data/seed-data.banking.json`](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Manage.API/Data/seed-data.banking.json) | 111 KB, 2696 lines | Banking customers, accounts and transactions. The Manage API loads it into SQL. |
| [`Source/CdrAuthServer/Data/customer-seed-data.json`](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Data/customer-seed-data.json) | 39 KB, 962 lines | The auth server reads it to find a customer by login identifier ([CustomerService.cs](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/CustomerService.cs)). |
| [`Source/CdrAuthServer.UI/public/customer-seed-data.json`](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.UI/public/customer-seed-data.json) | the same bytes as the file above (`cmp`) | The React UI gets this file in the browser. It uses it for the login check and for the account list. |
| `seed-data.energy.json`, `customer-seed-data-energy.json` | - | Energy profile. Not examined. |
| Test copies: `TestData/seed-data.json`, `Data/seed-data.json` in the test projects | - | Expected values for the tests |

- The README names the files `seed-data-{industry}.json` ([README.md L86-L87](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L86-L87)). The real names are `seed-data.banking.json` and `seed-data.energy.json`.
- The banking file has no `//` comments. `jq` parses it. (The mock Register seed files have comments.)
- The two customer data sets are separate files in two repositories. The auth server file has the same 21 login identifiers and the same 52 account identifiers as the data holder file, and no transactions (comparison with `jq` and `diff`). An operator who changes one file must change the other 2 copies.

### 4.2 Format and data model

Top-level keys: `Customers` and `LegalEntities`. The property names use PascalCase and are the names of the EF entity properties, not the CDS field names. The loader deserializes `Customers` directly into EF entities ([BankingDataHolderDatabaseContext.cs L89-L97](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/Infrastructure/BankingDataHolderDatabaseContext.cs#L89-L97)). There is no JSON schema and no validation step.

```
Customers[]
  CustomerID (GUID), CustomerUType ("person" | "organisation"), LoginId
  Person        PersonID, FirstName, LastName, MiddleNames, Prefix, Suffix,
                OccupationCode, OccupationCodeVersion, LastUpdateTime
  Organisation  OrganisationID, AgentFirstName, AgentLastName, AgentRole, BusinessName,
                LegalName, ShortName, ABN, ACN, IsACNCRegistered, IndustryCode,
                IndustryCodeVersion, OrganisationType, RegisteredCountry,
                EstablishmentDate, LastUpdateTime
  Accounts[]
    AccountId (string of 6 to 10 characters), CreationDate, DisplayName, NickName,
    OpenStatus, MaskedName, ProductCategory, ProductName, AccountOwnership
    Transactions[]
      TransactionId, TransactionType, Status, Description, PostingDateTime,
      ValueDateTime, ExecutionDateTime, Amount, Currency, Reference, MerchantName,
      MerchantCategoryCode, BillerCode, BillerName, CRN, ApcaNumber
LegalEntities[]   (data recipient metadata: legal entity, brands, software products)
```

One redacted example of the structure (all values replaced):

```json
{
  "Customers": [{
    "CustomerID": "00000000-0000-0000-0000-000000000000",
    "CustomerUType": "person",
    "LoginId": "<login id>",
    "Person": { "PersonID": "<guid>", "FirstName": "<x>", "LastName": "<x>", "LastUpdateTime": "<date-time>" },
    "Accounts": [{
      "AccountId": "<digits>", "DisplayName": "<x>", "MaskedName": "<x>", "OpenStatus": "OPEN",
      "ProductCategory": "TRANS_AND_SAVINGS_ACCOUNTS", "ProductName": "<x>", "AccountOwnership": "ONE_PARTY",
      "Transactions": [{ "TransactionId": "<x>", "TransactionType": "PAYMENT", "Status": "POSTED",
                         "Amount": 0.00, "PostingDateTime": "<date-time>", "Reference": "<x>" }]
    }]
  }]
}
```

Facts about the model:
- A customer is a person or an organisation, with one login identifier. An account belongs to one customer. A transaction belongs to one account. There is no joint account model: the code comment says "We don't have a concept of joint or shared accounts" ([BankingResourceRepository.cs L83-L88](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/BankingResourceRepository.cs#L83-L88)). `AccountOwnership` is only a stored string.
- The model has no balances, no account detail (rates, fees, features), no payees, no direct debits, no scheduled payments, no customer detail (addresses, phone numbers, email addresses) and no products. The entity list shows this ([Entities/](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/Entities)).
- The banking loader ignores `LegalEntities`. It reads only `Customers` ([BankingDataHolderDatabaseContext.cs L89-L97](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/Infrastructure/BankingDataHolderDatabaseContext.cs#L89-L97)). The file has 2 legal entities, each with 1 brand and 1 software product. The README says the file includes "data recipient metadata that can be obtained from the Register" ([README.md L87](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L87)).
- The internal `AccountId` and `TransactionId` are not the IDs in the API. See 6.4.

### 4.3 Content of `seed-data.banking.json`

Counts (from `jq` on the file):

| Item | Count |
|---|---|
| Customers | 21: 12 persons, 9 organisations |
| Accounts | 52. One customer has 30 accounts. 2 customers have no account. The other customers have 1 to 3 accounts. |
| Transactions | 85. One customer has 31. Most accounts have 1 transaction. |
| Account `OpenStatus` | 41 `OPEN`, 11 `CLOSED` |
| Account `ProductCategory` | `PERS_LOANS` 20, `TRANS_AND_SAVINGS_ACCOUNTS` 12, `TERM_DEPOSITS` 10, `BUSINESS_LOANS` 8, `CRED_AND_CHRG_CARDS` 1, `TRAVEL_CARDS` 1 |
| Account `AccountOwnership` | `UNKNOWN` 21, `ONE_PARTY` 16, `MANY_PARTY` 11, `TWO_PARTY` 4 |
| Transaction `TransactionType` | `TRANSFER_OUTGOING` 50, `PAYMENT` 33, `TRANSFER_INCOMING` 1, `INTEREST_PAID` 1 |
| Transaction `Status` | 84 `POSTED`, 1 `PENDING` |
| Transaction `PostingDateTime` | from 2022-03-01 to 2022-05-25 |

- The data is invented: names of persons and of organisations, occupation codes, ABNs, account names. It is a small set that someone wrote by hand. It is not generated.
- The customer with 30 accounts supports the pagination tests (for example `AC02_Get_WithPageSize5_ShouldRespondWith_200OK_Page1Of5Records`).
- Reuse: the MIT licence permits it. But the set is small, it covers only 2 of 19 banking operations, and the property names are not CDS names. Use it as a model for test scenarios (a customer with many accounts, a customer with no account, closed accounts, organisations). Do not use it as our data set.

### 4.4 How and when the data is loaded

- At the start of the Manage API: it runs the EF migrations (if the migrations connection string is set), then loads the file in `SeedData:FilePath` ([Startup.cs L95-L132](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Manage.API/Startup.cs#L95-L132)).
- It loads only when the `Customer` table is empty, unless `SeedData:OverwriteExistingData` is `true` ([Extensions.cs L38-L60](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Repository/Infrastructure/Extensions.cs#L38-L60), [BankingDataHolderDatabaseContext.cs L99-L102](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/Infrastructure/BankingDataHolderDatabaseContext.cs#L99-L102)). The Release value is `false` ([appsettings.Release.json L2-L7](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Manage.API/appsettings.Release.json#L2-L7)).
- A load is a full replacement in one database transaction: delete all transactions, customers, persons and organisations, then insert ([Extensions.cs L65-L117](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Repository/Infrastructure/Extensions.cs#L65-L117), [BankingDataHolderDatabaseContext.cs L64-L87](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/Infrastructure/BankingDataHolderDatabaseContext.cs#L64-L87)).
- The Manage `/health` endpoint is unhealthy until migration and seeding are complete. The other services wait for it (2.1).
- **Date offset does not work.** The code has an `OffsetDates` step that must move each date-time in the file from the baseline 2021-05-01 to the current date. The step calls `Regex.Replace` and discards the result, so the JSON does not change ([Extensions.cs L119-L127](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Repository/Infrastructure/Extensions.cs#L119-L127)). The Release configuration also sets `OffsetDates` to `false` ([appsettings.Release.json L5](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Manage.API/appsettings.Release.json#L5)). Thus each transaction stays in 2022. See 6.8.

### 4.5 Login identifiers and OTP

- The user identifier is the `LoginId` of a customer. The image configuration gives two login identifiers to the UI as a hint text ([Dockerfile L243](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L243)), and the UI shows them: "For testing you may use the following user names" ([LoginForm.tsx L58-L63](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.UI/src/components/LoginForm.tsx#L58-L63)).
- There is no password.
- **The OTP is the fixed value `000789`.** The UI shows a simulated SMS with this value, and the browser code compares the input with the literal ([OtpForm.tsx L26](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.UI/src/components/OtpForm.tsx#L26), [OtpForm.tsx L68-L72](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.UI/src/components/OtpForm.tsx#L68-L72)). The server has the same value as `AutoFillOtp`, only to fill the form ([appsettings.json L68-L69](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L68-L69)).
- **The server does not check the login identifier or the OTP.** The login check and the OTP check are in the browser ([LoginForm.tsx L41-L49](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.UI/src/components/LoginForm.tsx#L41-L49)). After the consent screens, the browser calls `GET /connect/authorize-callback` with the query parameters `subject_id` and `account_ids` ([useData.ts L39-L40](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.UI/src/hooks/useData.ts#L39-L40)). The server uses these values for the authorisation code grant, after it validates only the `client_id` and the `request_uri` ([AuthorisationController.cs L107-L134](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AuthorisationController.cs#L107-L134), [AuthorizeRequestValidator.cs L297-L311](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/AuthorizeRequestValidator.cs#L297-L311)). See 6.9.
- Headless mode (`CdrAuthServer:HeadlessMode`): the server skips the UI and uses one fixed user with two fixed account identifiers ([AuthorisationController.cs L66-L75](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AuthorisationController.cs#L66-L75), [HeadlessModeUser.cs](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Models/HeadlessModeUser.cs)). The default in `appsettings.json` is `true` ([appsettings.json L17](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L17)). The data holder image sets `false` ([Dockerfile L240](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L240)). The two fixed account identifiers are not in `seed-data.banking.json` (search of the file). Thus a headless consent gives a token that selects no seed account.

---

## 5. Product reference data

- **The ACCC Mock Data Holder has no product reference data.** It has no Get Products and no Get Product Detail endpoint (3.3), no product table (2.4) and no product seed file (4.1). An account has only the strings `ProductCategory` and `ProductName` (4.2).
- Get Accounts accepts these `product-category` values: `BUSINESS_LOANS`, `CRED_AND_CHRG_CARDS`, `LEASES`, `MARGIN_LOANS`, `OVERDRAFTS`, `PERS_LOANS`, `REGULATED_TRUST_ACCOUNTS`, `RESIDENTIAL_MORTGAGES`, `TERM_DEPOSITS`, `TRADE_FINANCE`, `TRANS_AND_SAVINGS_ACCOUNTS`, `TRAVEL_CARDS` ([AccountProductCategory.cs L18-L29](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Domain/ValueObjects/AccountProductCategory.cs#L18-L29)). The seed accounts use 6 of them (4.3).
- The DSB data is the only reference product data that I found. See 8.2: 55 generated products, 5 in each of 11 categories.
- I found no DSB repository that contains only example product reference data. The organisation repository list (GitHub API, 2026-10-06) has `product-comparator-demo` (MIT), a web application that reads the product APIs of real data holders. It is not a data set.

---

## 6. Deviations from the CDS and from the upstream specifications

Each item gives what the mock does, what CDS 1.36.0 says, and why it matters to us.

### 6.1 Endpoint coverage and endpoint versions

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 1 | Implements 2 of 19 banking operations and 1 of 2 customer operations (3.3). | [Banking APIs](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#banking-apis) | It is not a reference for most of our endpoints, and not for products. |
| 2 | Get Accounts serves only version 2: `[CheckXV(2, 2)]`, `[ApiVersion("2")]` ([ResourceController.cs L58-L62](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Controllers/ResourceController.cs#L58-L62)). A test expects 406 for `x-v: 3` (`AC15_Get_WithXV3_ShouldRespondWith_406NotAcceptable`). | Get Accounts is version 3 ([Get Accounts](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#get-accounts)). | A client that follows 1.36.0 and sends `x-v: 3` gets 406. Our version tests cannot use the mock as an oracle. |
| 3 | Get Transactions For Account serves only version 1 ([ResourceController.cs L114-L118](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Controllers/ResourceController.cs#L114-L118)). | Version 2 ([Get Transactions For Account](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#get-transactions-for-account)). | Same as item 2. |

### 6.2 Version negotiation and HTTP headers

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 4 | The banking, common and public APIs read only `x-v`: `HeaderApiVersionReader("x-v")`, with no version selector ([Startup.cs L77-L83](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Startup.cs#L77-L83)). The Common API and the Public API have the same configuration ([Program.cs L59-L63](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Common/CDR.DataHolder.Common.API/Program.cs#L59-L63), [Startup.cs L29-L33](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Public.API/Startup.cs#L29-L33)). They do not read `x-min-v`. Only the Admin API and the Energy API use the class that reads `x-min-v` ([ApiVersionSelector.cs L44-L117](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Versioning/ApiVersionSelector.cs#L44-L117), [Admin Startup.cs L43-L48](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Startup.cs#L43-L48)). | "The holder should respond with the highest supported version between x-min-v and x-v" ([HTTP headers](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-headers)). | `x-v: 3` with `x-min-v: 2` on Get Accounts must give version 2. The mock gives 406. |
| 5 | A missing `x-v` gives 400 `urn:au-cds:error:cds-all:Header/Missing`. A value that is not a positive integer gives 400 `Header/InvalidVersion`. A version that is not supported gives 406 `Header/UnsupportedVersion` ([ErrorResponseVersion.cs L13-L53](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Models/ErrorResponseVersion.cs#L13-L53)). | The same codes ([error codes](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes)). | Conforms. This is the same behaviour as the mock Register. |
| 6 | `x-fapi-interaction-id`: the middleware copies the request value to the response, or makes a new GUID ([InteractionIdMiddleware.cs L18-L34](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Middleware/InteractionIdMiddleware.cs#L18-L34)). It does not check that the value is a UUID: a test sends `123` and expects `123` (`AC20_Get_WithXFAPIInteractionId123_...`). The middleware is after `UseAuthentication` and `UseAuthorization` in the pipeline ([Startup.cs L208-L214](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Startup.cs#L208-L214)). UNVERIFIED by a run: a 401 response from the authorisation middleware then has no `x-fapi-interaction-id`. | "This header MUST be responded for success and error responses"; the value is an RFC 4122 UUID ([HTTP headers](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-headers)). | Our tests must not copy the `123` case. Expect a missing header on some error responses of the mock. |
| 7 | `x-fapi-auth-date` is mandatory on the 3 resource endpoints. Missing gives 400 `Header/Missing`. A value that is not in RFC 1123 format gives 400 `Header/Invalid` ([CheckAuthDateAttribute.cs L15-L25](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Filters/CheckAuthDateAttribute.cs#L15-L25)). | "Required for all resource calls (customer present and unattended)" (parameter description in `cds_banking.json`). | Conforms. |
| 8 | CORS: the mock-data-holder code has no CORS configuration (no match for `Cors` in the C# sources). Thus Get Status and Get Outages have no `Access-Control-Allow-Origin` header. The auth server gives the `AllOrigins` policy only to the JWKS endpoint ([Program.cs L175-L186](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Program.cs#L175-L186), [JwksController.cs L29](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/JwksController.cs#L29)), not to the discovery document. | "Cross-origin resource sharing (CORS) protections must be disabled (i.e. Access-Control-Allow-Origin set to *) for all unauthenticated endpoints" ([CORS](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors)). | Follow the CDS. A browser client cannot read the mock's status endpoints. |

### 6.3 Certificate binding and trust in forwarded headers

| # | The mock | Specification | Why it matters |
|---|---|---|---|
| 9 | `cnf.x5t#S256` in the access token is the value of the `X-TlsClientCertThumbprint` header ([TokenController.cs L60](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/TokenController.cs#L60), [TokenController.cs L72-L81](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/TokenController.cs#L72-L81), [TokenService.cs L582-L590](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L582-L590)). The gateway sets that header to the **SHA-1** hexadecimal thumbprint (2.3). | CDS: "[MTLS] HoK allows issued tokens to be bound to a client certificate as specified in section 3 of [MTLS]" ([Holder of Key](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism)). RFC 8705 section 3.1 gives base64url of the SHA-256 hash. | The same deviation as the mock Register. A recipient that checks `cnf` against its certificate fails. Our data holder follows RFC 8705, so the token contents differ. |
| 10 | The resource APIs compare `cnf.x5t#S256` with the request header `X-TlsClientCertThumbprint`, not with a TLS certificate ([MTLSHandler.cs L29-L67](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Authorisation/MTLSHandler.cs#L29-L67)). The compose file publishes the Common API port 8006 (2.1). UNVERIFIED by a run: a caller with an access token can call port 8006 directly and send the header value that is in the token. | "resource requests MUST be validated to ensure the client certificate and access token match" ([Holder of Key](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism)). | Do not copy the design. In our lab, no backend port must accept a certificate header from a client. |
| 11 | The token endpoint does not bind the mTLS certificate to the client. The mTLS filter checks only that the thumbprint header is present and has one value ([ValidateMtlsAttribute.cs L45-L57](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/ValidateMtlsAttribute.cs#L45-L57)). An option turns the check off for a list of paths ([ValidateMtlsAttribute.cs L33-L38](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/ValidateMtlsAttribute.cs#L33-L38)). | [Client authentication](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#client-authentication), [Transaction security](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security) | Any certificate from the mock CA works for any client. Our negative tests (wrong certificate for the client) cannot use the mock as an oracle. |
| 12 | No revocation check: `RevocationMode = NoCheck` at the gateway (2.3). The auth server has an OCSP check that is off by default ([appsettings.json L83-L86](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L83-L86), [ValidateMtlsAttribute.cs L40-L43](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/ValidateMtlsAttribute.cs#L40-L43)). | [Certificate validation](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation) | The mock accepts a revoked client certificate. |
| 13 | Outbound calls do not validate server certificates by default (2.5). The gateway accepts each backend certificate (2.2). | [Transaction security](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security) | Lab convenience only. |

### 6.4 ID permanence and the PPID `sub`

Source for the algorithm: [IdPermanenceHelper.cs L21-L51](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/IdPermanence/IdPermanenceHelper.cs#L21-L51) (IDs), [L101-L125](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/IdPermanence/IdPermanenceHelper.cs#L101-L125) (`sub`), [AesEncryptor.cs L10-L20](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/IdPermanence/AesEncryptor.cs#L10-L20). The authorisation-server has its own copy in `Source/CdrAuthServer/IdPermanence/`. The two copies are not identical text. UNVERIFIED: that they give the same output (they must, because the data holder decrypts what the auth server encrypts).

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 14 | `accountId` and `transactionId` in the API = base64 of AES(`loginId` + internal ID). The AES key = the first 24 bytes of SHA-512(`softwareProductId` + private key). The IV is 16 zero bytes. The helper changes `/` to `%2F`. The result is the same on each call (test `ACX02_GetAccountsMultipleTimes_ShouldRespondWith_SameEncryptedAccountIds`). | "IDs MUST be immutable across sessions and consents but MUST NOT be transferable across Data Recipient Software Products"; "the ID value MUST be entirely arbitrary and have no inherent meaning" ([ID permanence](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#id-permanence)). | The scope rules agree with the CDS: one ID for each software product and customer. But the ID is reversible encryption of the login identifier and the account number, with a key that is public (item 16). The ID can contain `+`, `=` and `%2F`, which a client must encode in a URL path. Our tests must treat IDs as opaque strings. |
| 15 | `sub` = base64 of AES(`loginId`) with the same key derivation ([TokenService.cs L219-L229](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L219-L229)). The key uses the software product ID. It does not use `sector_identifier_uri`, although the parameter object has it. The value is not a UUID. The data holder decrypts `sub` to get the login identifier ([TokenTransformService.cs L33-L44](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Business/Services/TokenTransformService.cs#L33-L44)). | "The Data Holder MUST generate the sub value as a Pairwise Pseudonymous Identifier (PPID)"; "MUST support the sector_identifier_uri in PPID generation ... if this field was supplied"; a version 4 UUID is RECOMMENDED ([Identifiers and subject types](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#identifiers-and-subject-types)). | Two software products with the same `sector_identifier_uri` get different `sub` values from the mock. OIDC pairwise rules give them the same value. Expect a different `sub` format from our data holder. |
| 16 | The private key of the ID permanence scheme is a literal in the repositories ([appsettings.json L21-L23](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/appsettings.json#L21-L23), [appsettings.json L11-L13](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L11-L13)). | - | Each person can decrypt each `sub` and each ID from a default deployment. |
| 17 | The access token carries the consented accounts as encrypted `account_id` claims, plus `software_id`, `sector_identifier_uri`, `cdr_arrangement_id` and a `sub` ([TokenService.cs L61-L109](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L61-L109)). The resource API reads the account list from the token, not from a consent store ([ClaimsPrincipalExtensions.cs L68-L74](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Extensions/ClaimsPrincipalExtensions.cs#L68-L74)). | CDS does not define the access token contents. | A design choice, not a deviation. It means that the two services share no consent API. |

### 6.5 Tokens, introspection and revocation

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 18 | Access token: a JWT with `aud` = `cds-au`, lifetime 300 s (default `AccessTokenExpirySeconds`) ([TokenService.cs L111-L117](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L111-L117), [ConfigExtensions.cs L146-L148](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Extensions/ConfigExtensions.cs#L146-L148)). The resource API accepts a clock skew of 1 minute ([Startup.cs L119-L127](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Startup.cs#L119-L127)). | "An Access Token MUST expire between 2 minutes to 10 minutes" ([Tokens](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#tokens)). | Conforms. |
| 19 | Refresh token: a GUID. The server issues it only when `sharing_duration` > 0, with the expiry = now + `sharing_duration` ([TokenService.cs L362-L372](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L362-L372), [TokenService.cs L420-L441](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L420-L441)). A refresh grant returns the same refresh token ([TokenService.cs L327-L328](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L327-L328)). | "MUST NOT cycle refresh tokens (rotation)"; the expiry MUST equal the sharing duration ([Refresh token](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#refresh-token)). | Conforms. A good oracle for these two rules. |
| 20 | With no `sharing_duration`, the arrangement expires at now + `AddMinutes(AccessTokenExpirySeconds)`, which is 300 minutes with the defaults ([TokenService.cs L445-L458](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L445-L458)). | - | Probably a units defect. Do not copy it. |
| 21 | Authorisation code lifetime 5 minutes ([AuthorisationController.cs L205-L216](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AuthorisationController.cs#L205-L216)). When a client sends a code that the server does not find (for example a used code), the server puts the key `{client_id}::{code}` on the blocklist, and the internal introspection then refuses each access token from that code ([TokenRequestValidator.cs L105-L108](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/TokenRequestValidator.cs#L105-L108), [IntrospectionController.cs L130-L139](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/IntrospectionController.cs#L130-L139)). `request_uri` lifetime 90 s and one use ([appsettings.json L65](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L65), [AuthorizeRequestValidator.cs L150-L160](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/AuthorizeRequestValidator.cs#L150-L160)). The request object must have `exp` and `nbf`, with `exp` at most 60 minutes after `nbf` ([RequestObjectValidator.cs L132-L154](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/RequestObjectValidator.cs#L132-L154)). | [Request object](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-object) | Reference values for our tests. |
| 22 | `POST /connect/introspect` accepts only refresh tokens. The response has `active`, `cdr_arrangement_id`, `exp` and `scope`. Each failure gives `{"active": false}` ([IntrospectionController.cs L38-L83](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/IntrospectionController.cs#L38-L83)). | "Introspection of Refresh Tokens MUST be supported. Introspection of Access Tokens and ID Tokens MUST NOT be supported." ([Introspection endpoint](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#introspection-endpoint)) | Conforms. |
| 23 | `POST /connect/introspect-internal` on port 8001 introspects **access tokens** with no authentication. The code comment says: "There is currently no auth on this endpoint" ([IntrospectionController.cs L84-L100](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/IntrospectionController.cs#L84-L100)). The resource APIs call it for each request ([AccessTokenHandler.cs L55-L94](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Authorisation/AccessTokenHandler.cs#L55-L94), [Dockerfile L232](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Dockerfile#L232)). | Same anchor as item 22. | Port 8001 is public, so each caller can introspect an access token. Our design must keep this call internal. |
| 24 | `POST /connect/revocation` returns 200 always. For a refresh token it deletes the grant. For an access token of the same client it puts the `jti` on the blocklist ([RevocationController.cs L27-L89](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/RevocationController.cs#L27-L89)). | "Revocation of Refresh Tokens and Access Tokens MUST be supported" ([Token revocation endpoint](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#token-revocation-endpoint)). | Conforms. |
| 25 | A revoked access token, an expired arrangement or an inactive software product makes the internal introspection fail. The resource API then returns **400 with no body**; the code comment says "Return 400 not 401" ([PolicyAuthorizeAttribute.cs L37-L40](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Authorisation/PolicyAuthorizeAttribute.cs#L37-L40), [DataHolderUnauthorizedResult.cs L15-L20](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Authorisation/DataHolderUnauthorizedResult.cs#L15-L20)). | 401 for a token that is not valid; 403 `Authorisation/RevokedConsent` or `Authorisation/AdrStatusNotActive` ([error codes](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes)). | Do not copy. Expected difference in each negative token test. |

### 6.6 Arrangement revocation

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 26 | `POST /connect/arrangements/revoke` with the form field `cdr_arrangement_id`: 204 on success; 400 `Field/Missing` with no ID; 422 `urn:au-cds:error:cds-all:Authorisation/InvalidArrangement` when the client has no arrangement with that ID. It deletes the refresh token grant and the arrangement grant ([ArrangementRevocationController.cs L35-L90](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/ArrangementRevocationController.cs#L35-L90)). | [CDR Arrangement Revocation endpoint](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint) | Good oracle for the status codes. |
| 27 | A client with a software product in the status `REMOVED` cannot revoke ([ArrangementRevocationController.cs L61-L78](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/ArrangementRevocationController.cs#L61-L78)). | [Data holder responsibilities](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities) | Check the CDS table before we copy the rule. |
| 28 | Revocation that the data holder starts (the consumer withdraws at the data holder): there is no dashboard. A test-only endpoint `GET /utility/dr/revoke-arrangement-jwt/{cdrArrangementId}` deletes the arrangement and calls the recipient's `{recipient_base_uri}/arrangements/revoke` with a `cdr_arrangement_jwt` ([UtilityController.cs L35-L48](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/UtilityController.cs#L35-L48), [ConsentRevocationService.cs L94-L126](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/ConsentRevocationService.cs#L94-L126)). The endpoint has no authentication attribute. | [Withdrawal standards](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#withdrawal-standards) | A model for our operator interface: one operator action that revokes an arrangement and notifies the recipient. |

### 6.7 Error codes and HTTP statuses of the resource APIs

| # | Condition | The mock | CDS 1.36.0 |
|---|---|---|---|
| 29 | The account ID cannot be decrypted, or no account has that ID | 404 `urn:au-cds:error:cds-all:Resource/NotFound`, detail "Account ID could not be found for the customer" ([ResourceController.cs L143-L164](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Controllers/ResourceController.cs#L143-L164)) | 404 `urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount`, detail = the account ID ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-authorisation-invalid-banking-account)) |
| 30 | The account exists but is not in the consent | 404 with the code `urn:au-cds:error:cds-all:Authorisation/UnavailableBankingAccount` (the namespace `cds-all` is wrong) and an empty detail ([ResourceController.cs L166-L176](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Controllers/ResourceController.cs#L166-L176), [ResponseErrorList.cs L189-L198](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Domain/Models/ResponseErrorList.cs#L189-L198)) | `urn:au-cds:error:cds-banking:Authorisation/UnavailableBankingAccount` means "temporarily unavailable" ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-404-authorisation-unavailable-banking-account)). The account that is not in the consent is the Invalid Banking Account case. |
| 31 | `page` larger than the last page | **400** `urn:au-cds:error:cds-all:Field/Invalid`, detail "Page is out of range" ([ResourceController.cs L92-L96](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Controllers/ResourceController.cs#L92-L96), [ResponseErrorList.cs L111-L115](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Domain/Models/ResponseErrorList.cs#L111-L115)). The file has a second method with `Field/InvalidPage` and the comment "This matches CDS better" ([L117-L121](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Domain/Models/ResponseErrorList.cs#L117-L121)), which the controller does not call. Get Transactions has no page range check. | **422** `urn:au-cds:error:cds-all:Field/InvalidPage` ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-422-field-invalid-page)) |
| 32 | `page` not a positive integer, or larger than 1000 | 400 `Field/Invalid` ([CheckPageAttribute.cs L18-L21](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Filters/CheckPageAttribute.cs#L18-L21)) | 400 `Field/Invalid` ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid)). CDS has no maximum of 1000 pages. |
| 33 | `page-size` not a positive integer, or larger than 1000 | 400 `Field/InvalidPageSize` ([CheckPageSizeAttribute.cs L18-L26](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Filters/CheckPageSizeAttribute.cs#L18-L26)) | 400 `Field/InvalidPageSize` for a size above 1000 ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid-page-size)). Conforms. |
| 34 | `open-status` or `product-category` not in the list | 400 with the codes `urn:au-cds:error:cds-all:Field/InvalidOpenStatus` and `Field/InvalidProductCategory` ([ResponseErrorList.cs L217-L225](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Domain/Models/ResponseErrorList.cs#L217-L225)) | These codes are not in the CDS list. The CDS code is `Field/Invalid` ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid)). |
| 35 | The token does not have the scope | The policy gives 403 `urn:au-cds:error:cds-all:Authorisation/InvalidConsent` ([PolicyAuthorizeAttribute.cs L48-L54](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Authorisation/PolicyAuthorizeAttribute.cs#L48-L54)) | 403 `Authorisation/InvalidConsent` ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-invalid-consent)). Conforms. |
| 36 | The `cnf` value does not match the header | 401 `{"errors":[{"code":"401","title":"Unauthorized","detail":"invalid_token"}]}`: the code is not a URN ([PolicyAuthorizeAttribute.cs L48-L51](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Authorisation/PolicyAuthorizeAttribute.cs#L48-L51)) | CDS error codes are URNs ([error codes](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes)). |
| 37 | Get Customer finds no customer for the login identifier | 400 with no body ([CommonController.cs L50-L67](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Common/CDR.DataHolder.Common.API/Controllers/CommonController.cs#L50-L67)) | An error response has an `errors` list ([error codes](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes)). |
| 38 | Missing, malformed or expired bearer token | 401 from the JwtBearer middleware, with a `WWW-Authenticate` header (the tests assert the header: `AC08_Get_WithExpiredAccessToken_...`). For Get Accounts only, a middleware adds the body of item 36 when the token is not a JWT ([ResourceAuthoriseErrorHandlingMiddleware.cs L19-L48](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Business/Middleware/ResourceAuthoriseErrorHandlingMiddleware.cs#L19-L48)). UNVERIFIED: the body in the other cases. | - |

The error envelope is the CDS envelope `{"errors":[{"code","title","detail"}]}`. JSON nulls are omitted ([Startup.cs L65-L69](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Startup.cs#L65-L69)).

### 6.8 Pagination, filters and data behaviour

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 39 | Defaults `page` 1 and `page-size` 25. `links` has `self`, `first`, `last`, `prev`, `next`; `prev` and `next` are null (omitted) at the ends; `first` and `last` are omitted when there are 0 pages. `meta` has `totalRecords` and `totalPages` ([Extensions.cs L19-L75](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.Business/Extensions.cs#L19-L75)). | [Pagination](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination) | Conforms. A good oracle for the link rules. |
| 40 | Accounts sort by `DisplayName`, then `AccountId`. Transactions sort by `PostingDateTime` descending, then `ExecutionDateTime` descending ([BankingResourceRepository.cs L114-L118](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/BankingResourceRepository.cs#L114-L118), [L195-L199](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/BankingResourceRepository.cs#L195-L199)). | CDS gives no sort order for accounts. | Normalise the order before a comparison. |
| 41 | `is-owned=false` gives an empty list always. `isOwned` is `true` on each account ([BankingResourceRepository.cs L83-L88](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/BankingResourceRepository.cs#L83-L88), [MappingProfile.cs L38-L44](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Business/MappingProfile.cs#L38-L44)). | [Get Accounts](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#get-accounts) | The mock has no accounts that the customer does not own. |
| 42 | Get Transactions: `newest-time` defaults to now and `oldest-time` to `newest-time` minus 90 days ([BankingResourceRepository.cs L145-L153](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/BankingResourceRepository.cs#L145-L153)). This agrees with the CDS parameter descriptions. But the seed transactions are all in 2022 and the date offset does not work (4.4). Thus, derived from the code and the data: a request with no `oldest-time` returns 0 transactions. UNVERIFIED by a run. | [Get Transactions For Account](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#get-transactions-for-account) | A recipient test against the mock must send `oldest-time`. Our seed data must have dates relative to the load time, or a fixed clock. |
| 43 | `text` matches `description` or `reference` with SQL `LIKE` ([BankingResourceRepository.cs L188-L191](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/BankingResourceRepository.cs#L188-L191)). `min-amount` and `max-amount` are inclusive. A reversed time or amount range gives 400 `Field/Invalid` ([RequestAccountTransactions.cs L46-L54](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Business/Models/RequestAccountTransactions.cs#L46-L54), [ModelStateErrorMiddleware.cs L30-L34](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Shared.API.Infrastructure/Middleware/ModelStateErrorMiddleware.cs#L30-L34)). | Parameter descriptions in `cds_banking.json` | Reference behaviour for the filters. The case sensitivity depends on the SQL Server collation (UNVERIFIED). |
| 44 | The account check for Get Transactions asks only "does an account with this ID exist", not "does it belong to this customer" ([BankingResourceRepository.cs L132-L135](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Repository/BankingResourceRepository.cs#L132-L135)). The consent check on the token claims comes after it. | - | The order of the two checks decides which 404 code a client gets (items 29 and 30). |

### 6.9 Consent UI against the CX standards

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 45 | The OTP is one fixed value for each customer and each session. It does not expire. The UI shows it in the same page as a simulated SMS (4.5). | "The data holder MUST provide a one-time password (OTP) to the customer through an existing channel or mechanism"; "The provided OTP MUST be invalidated after a period of time"; "MUST be numeric digits and be between 4 and 6 digits in length" ([OTP requirements](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#one-time-password-credential-requirements)). | Only the format (6 digits) agrees. A fixed OTP is a mock feature, not CDS behaviour. Record our own decision. |
| 46 | The server does not authenticate the user. A caller who has a valid `client_id` and `request_uri` can call `GET /connect/authorize-callback` with a `subject_id` and `account_ids` that it selects, and get an authorisation code (4.5). The server does not check that the accounts belong to the subject. | [Authentication flows](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authentication-flows) | Do not copy. But see 9: it lets a test get a consent with no browser. |
| 47 | The UI is a 4-step flow: login identifier, OTP, account selection, confirmation ([pages/](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.UI/src/pages)). I did not compare the screens with the CX standards. UNVERIFIED: conformance to the [authorisation standards](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-standards). | - | The mock is not a CX reference. |

### 6.10 Register status handling

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 48 | The auth server checks the software product status from its local table (`CheckSoftwareProductStatus`, default `true`) at the authorise step, the token endpoint, UserInfo, arrangement revocation and the internal introspection ([AuthorizeRequestValidator.cs L263-L275](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/AuthorizeRequestValidator.cs#L263-L275), [TokenRequestValidator.cs L68-L78](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/TokenRequestValidator.cs#L68-L78), [UserInfoController.cs L53-L66](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/UserInfoController.cs#L53-L66), [IntrospectionController.cs L141-L160](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/IntrospectionController.cs#L141-L160)). The error is `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive` with the description "Software product status is {status}" ([ErrorCatalogue.cs L195-L196](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Models/ErrorCatalogue.cs#L195-L196)). | [Data holder responsibilities](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities), [error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-adr-status-not-active) | It does refuse an inactive software product. A software product that is not in the table is also refused. UNVERIFIED: how `IsActive()` combines the legal entity, brand and product statuses. |
| 49 | **Within what time: no limit.** The table changes only on a Metadata Update call (2.5). The data holder image has no poller. With the separate Azure Function, the period is the timer schedule. | "Data Holders MUST react to Data Recipient and associated Software Statuses changes within 5 minutes" ([Metadata cache management](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management)). | The mock alone does not meet the 5-minute rule. Our data holder must poll. |
| 50 | The refresh uses Get Data Recipients only, with `x-v: 3` (2.5). | Get Data Recipients is version 4 in CDS 1.36.0 (see the [CDS Register APIs note](2026-09-25-cds-register-apis.md)). The CDS names 3 endpoints to poll ([Metadata cache management](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management)). | To use the mock with our Register, set `Version`, or serve version 3. |
| 51 | For a resource request with a token of an inactive software product, the resource API returns 400 with no body (item 25). The integration test with "403Forbidden" in its name expects `BadRequest` ([US12973 L509-L541](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Tests.IntegrationTests/US12973_MDH_CommonAPI_GetCustomer.cs#L509-L541)). | 403 `Authorisation/AdrStatusNotActive` ([error](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-adr-status-not-active)) | Expected difference. |
| 52 | An option `RevokeRemovedSoftwareProductConsents` (default `false`) revokes the arrangements of `REMOVED` software products at a refresh ([AdminController.cs L77-L83](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AdminController.cs#L77-L83), [CdrRegisterConfiguration.cs L11](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Configuration/CdrRegisterConfiguration.cs#L11)). | [Data holder responsibilities](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities) | By default the mock does not invalidate consents. |

### 6.11 Admin endpoints

| # | The mock | CDS 1.36.0 | Why it matters |
|---|---|---|---|
| 53 | Get Metrics accepts 2 token types: an access token of the data holder (issuer, audience `cds-au` or the admin base URI, scope `admin:metrics.basic:read`), or a JWT with `iss` = `sub` = `cdr-register`, a `jti`, and a signature that the Register JWKS verifies ([AdminController.cs L95-L219](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Controllers/AdminController.cs#L95-L219)). It does not check the `jti` for replay. Each failure uses the OAuth error `invalid_client` or `invalid_scope`. | [Get Metrics](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#get-metrics), [CDR Register calling data holders](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-calling-data-holders) | A model of the two authentication methods of the Register. The metrics values are static (3.1). |
| 54 | Metadata Update needs an access token of a known client. The only action is `REFRESH`. Success gives 200 with a text body. Failure gives 500 with a text body ([AdminController.cs L44-L98](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AdminController.cs#L44-L98)). | [Metadata Update](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-update) | UNVERIFIED: the CDS response body for this endpoint. Check the OpenAPI before we compare. |

---

## 7. Test and operator features

Ways to change data or behaviour at run time. None of them is a CDS API.

| Feature | How | Auth | Source |
|---|---|---|---|
| Replace all banking data | `POST https://localhost:8005/manage/metadata` with a body in the seed file format. It deletes all customers, accounts and transactions, then loads the body. 200 with no body on success. 400 `{"error":"Unexpected Error","detail":"An error occurred loading the database."}` on each exception. | **none** | [ManageController.cs L35-L54](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Manage.API/Controllers/ManageController.cs#L35-L54), [README.md L104-L130](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L104-L130) |
| Replace the seed file | Put a file in `/app/manage/Data`, set `SeedData:FilePath` and `SeedData:OverwriteExistingData: true`, restart | container access | [README.md L89-L102](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L89-L102) |
| Health | `GET /health` on each service. The Manage API reports migration and seeding status. | none | [Startup.cs L86-L89](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Manage.API/Startup.cs#L86-L89) |
| Trigger a revocation from the data holder | `GET /utility/dr/revoke-arrangement-jwt/{cdrArrangementId}` on the auth server (6.6) | none | [UtilityController.cs L35-L48](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/UtilityController.cs#L35-L48) |
| Refresh data recipient statuses | `POST /cds-au/v1/admin/register/metadata` (2.5) | bearer | [AdminController.cs L44-L98](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AdminController.cs#L44-L98) |
| Skip the consent UI | `CdrAuthServer:HeadlessMode`; `HeadlessModeRequiresConfirmation` adds one HTML page with Proceed and Cancel buttons (4.5) | config | [AuthorisationController.cs L66-L75](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AuthorisationController.cs#L66-L75), [L136-L197](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AuthorisationController.cs#L136-L197) |
| Fill the login form | `AutoFillCustomerId`, `AutoFillOtp` | config | [appsettings.json L68-L69](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L68-L69) |
| Other auth server flags | `AllowDuplicateRegistrations`, `SupportJarmEncryption`, `UseMtlsEndpointAliases`, `ClockSkewSeconds`, `ScopesProfile`, `AccessTokenExpirySeconds`, `RequestUriExpirySeconds`, `OverrideMtlsCheckEndpointList`, `CdrRegister:CheckSoftwareProductStatus` | config | [appsettings.json L17-L22](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L17-L22), [ConfigExtensions.cs L140-L157](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Extensions/ConfigExtensions.cs#L140-L157) |
| Metrics versions | `GetMetricsSupportedVersions` (default `4,5`) | config | [Admin Startup.cs L31-L41](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Shared/CDR.DataHolder.Admin.API/Startup.cs#L31-L41) |
| Swagger UI | The Banking Resource API adds Swagger UI at `/swagger` on its own port | none | [Startup.cs L203-L204](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Resource.API/Startup.cs#L203-L204) |

Facts:
- **There is no API to create or change one customer, one account or one transaction.** The only write operation is the full replacement. There is no `GET /manage/metadata`, although the README refers to one ([README.md L117](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L117)); the controller has only the POST action.
- The README says that the Manage API "also includes trigger points to refresh the Data Recipient, Data Recipient Status and Software Product Status from the Mock Register" ([README.md L189](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L189)). I found no such action in the Manage API. The refresh is the Metadata Update endpoint of the auth server.
- The README states: "there is currently no authentication/authorisation applied to the Manage API endpoints as these are seen to be under the control of the container owner" ([README.md L130](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/README.md#L130)). The compose file publishes port 8005.
- The replacement through `POST /manage/metadata` does not change the two customer files of the auth server and its UI (4.1). After a replacement with new customers, the login step does not know them.
- The auth server has no API to create a customer. It reads its customer file.

---

## 8. Other reference mocks

### 8.1 authorisation-server as a separate product

Section 2.6 of the Auth servers note gives the summary. More facts for a data recipient that calls it and for a data holder that uses it:

- Client authentication: `private_key_jwt`. The assertion `aud` must be the issuer or the token endpoint URL. `jti` is mandatory. The validator refuses a `jti` that is in the blocklist table, but I found no code that adds the `jti` of a client assertion to that table (search for `AddToBlacklist`). Thus, derived from the code, the server does not detect a replayed client assertion. UNVERIFIED by a run. `iss` and `sub` must be equal to the `client_id` ([ClientAssertionValidator.cs L76-L81](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/ClientAssertionValidator.cs#L76-L81), [L153-L173](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/ClientAssertionValidator.cs#L153-L173)).
- Token endpoint: `application/x-www-form-urlencoded`, `Cache-Control: no-store`, and PKCE `code_verifier` is mandatory for the authorisation code grant ([TokenController.cs L31-L41](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/TokenController.cs#L31-L41), [TokenRequestValidator.cs L99-L135](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/TokenRequestValidator.cs#L99-L135)).
- Authorisation response: JARM only. A `response_type` of `code` without `response_mode` `jwt` is an error ([RequestObjectValidator.cs L232-L253](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/RequestObjectValidator.cs#L232-L253)). The JARM JWT lives 300 s ([AuthorisationController.cs L372-L381](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/AuthorisationController.cs#L372-L381)).
- DCR: the server gets the Register JWKS from `SsaJwksUri` for each registration and checks the SSA issuer `cdr-register` and the SSA lifetime ([ClientRegistrationValidator.cs L324-L342](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Validation/ClientRegistrationValidator.cs#L324-L342)).
- A new consent with an existing `cdr_arrangement_id` keeps the arrangement ID, deletes the old refresh token and issues a new one ([TokenService.cs L381-L398](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Services/TokenService.cs#L381-L398)).
- The signing keys for ID tokens, access tokens and JARM are two committed PFX files (10).
- The repository also has a standalone image with its own TLS and mTLS gateways and its own minimal resource endpoints `resource/cds-au/v1/common/customer` and `resource/cds-au/v1/{industry}/accounts` ([ResourceController.cs L29-L30](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/ResourceController.cs#L29-L30), [L85-L86](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Controllers/ResourceController.cs#L85-L86)).
- Correction to the Auth servers note: that note gives the README alignment of 1.34.0 and marks the effect of 1.35 and 1.36 as UNVERIFIED. It is still UNVERIFIED. This note found no code change for those versions (1.3).

### 8.2 DSB mock-data-holder-nodejs and testdata-cli

- The DSB mock is a set of containers: the ACCC mock Register, DSB builds of the ACCC auth server, its UI, its gateways and the ACCC mock data recipient, plus the DSB resource server and a data loader ([README.md L5](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/README.md#L5), [docker-compose.yaml](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/docker-compose.yaml)). Thus its InfoSec behaviour is the ACCC auth server behaviour (sections 4.5, 6.3 to 6.6, 6.9).
- The resource server is NodeJS with MongoDB (dependencies `express`, `mongoose`, `@cds-au/holder-sdk`) ([test-data-server/package.json](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/test-data-server/package.json)).
- It implements more banking endpoints than the ACCC mock: accounts (list, detail, balance, bulk balances), transactions (list, detail), direct debits, scheduled payments, payees, **products and product detail**, customer and customer detail (route definitions in [test-data-server/src](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/test-data-server/src)). UNVERIFIED: the endpoint versions that it serves and its conformance.
- Seed data format: one JSON file for each CDS version, in `load-test-data/input/<version>/all-data/`. The versions are 1.24.0, 1.26.0, 1.28.0, 1.29.1, 1.31.0 and 1.33.0 ([load-test-data/input](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/load-test-data/input)). There is no 1.36.0 file.
- The file follows the "Consumer Data Right Test Data JSON Schema" of the testdata-cli ([cdr-test-data-schema.json](https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986/src/schema/cdr-test-data-schema.json)). Top-level keys: `fileVersion`, `standardsVersion`, `title`, `description`, `holders`, `clientCache`, `registerCache`. Structure:

```
holders[]
  holderId
  holder
    unauthenticated
      banking.products[]        (product reference data, CDS field names)
      energy                    (plans)
    authenticated
      customers[]
        customerId
        customer                (customerUType, person | organisation, CDS field names)
        banking
          accounts[]            account (CDS account detail), balance, transactions[]
          directDebits[], payees[], payments[]
        energy
```

- Content of `all-data-1.33.0.json` (3.3 MB, from `jq`): 1 holder; **55 banking products, 5 in each of 11 categories** (`BUSINESS_LOANS`, `CRED_AND_CHRG_CARDS`, `MARGIN_LOANS`, `OVERDRAFTS`, `PERS_LOANS`, `REGULATED_TRUST_ACCOUNTS`, `RESIDENTIAL_MORTGAGES`, `TERM_DEPOSITS`, `TRADE_FINANCE`, `TRANS_AND_SAVINGS_ACCOUNTS`, `TRAVEL_CARDS`; no `LEASES`); 3 customers, each with 10 accounts, 100 transactions, 2 direct debits, 8 payees and 12 scheduled payments.
- The data is generated, not written by hand. The testdata-cli makes it with "data factories" and the `@faker-js/faker` library ([README.md L109-L120](https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986/README.md#L109-L120), [package.json](https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986/package.json)). An options file names the factories and their options, for example 5 products for each category ([create-banking-products.json](https://github.com/ConsumerDataStandardsAustralia/testdata-cli/blob/c41ba785eefaaadced5a0c2b4ca66627bdd37986/samples/options/create-banking-products.json)). The tool also has factories for data that is not valid on purpose (`src/factories/invalid-factories`).
- Because the data uses the CDS field names and schemas, it is a closer model for our seed files than the ACCC file. But the licence is not clear (1.1), and the newest file is for CDS 1.33.0.
- UNVERIFIED: how the DSB stack maps a login identifier of the ACCC auth server to a `customerId` in its data.

### 8.3 mock-data-recipient as a test peer

Can it be a test peer for our data holder? Yes, with configuration and with our Register in front of it. What it needs:

- It is a web application at `https://localhost:9001` with pages for Get Data Holder Brands, Get SSA, Dynamic Client Registration, consent with PAR, token operations and a Swagger UI proxy for the data holder APIs. It also hosts the recipient JWKS and the CDR Arrangement Revocation endpoint ([README.md L79-L99](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/README.md#L79-L99)). It supports only the PAR flow ([README.md L113](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/README.md#L113)).
- Storage: SQL Server (`cdr-mdr` database) ([appsettings.Release.json L2-L6](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.Web/appsettings.Release.json#L2-L6)). So the lab would need a SQL Server container and the Microsoft EULA.
- Register: it needs the Register TLS base URI, mTLS base URI and OIDC discovery URI ([appsettings.Release.json L9-L14](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.Web/appsettings.Release.json#L9-L14)). It finds data holders from Get Data Holder Brands, so our Register must list our data holder brand with the correct `endpointDetail` values. It asks for the SSA with a minimum `x-v` of 3 (`getSsaMinXv`). The optional Azure Functions use `x-v: 2` for Get Data Holder Brands and `x-v: 3` for the SSA ([DiscoverDataHolders appsettings.docker.json L6-L8](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DiscoverDataHolders/appsettings.docker.json#L6-L8), [DCR appsettings.docker.json L6-L8](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DCR/appsettings.docker.json#L6-L8)). CDS 1.36.0 has versions 3 and 4 for these endpoints. UNVERIFIED: the `x-v` that the web application sends for Get Data Holder Brands.
- Identity: one software product ID and one brand ID in configuration, and the `jwksUri`, `redirectUris` and `recipientBaseUri` of the product ([appsettings.json L28-L43](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.Web/appsettings.json#L28-L43), [appsettings.Release.json L15-L19](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.Web/appsettings.Release.json#L15-L19)). Our Register seed file must have a software product with these values, or we change the configuration.
- Certificates: a client certificate (`client.pfx`), a signing certificate (`jwks.pfx`) and a TLS server certificate, each a file path with a password in configuration. The shipped client certificate is from the Mock CDR CA with the subject CN `MockDataRecipient` (`openssl x509` on [client.pem](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/CertificateManagement/mtls/client.pem)). For our lab, issue a client certificate from the lab CDR CA and replace the files. Our Register must be able to get `https://mock-data-recipient:9001/jwks` and trust its TLS certificate.
- It accepts each server certificate by default (`AcceptAnyServerCertificate: true`) ([appsettings.json L68](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.Web/appsettings.json#L68)). So it does not test our server certificates.
- It expects the ACCC mock behaviour. Expect problems where our data holder follows the specifications and the mock does not: the `cnf` format (6.3) if it checks it, endpoint versions (6.1), and error formats. UNVERIFIED: which of these the recipient code depends on. I did not read the SDK code.
- It is a browser application. Automated use needs Playwright, as the ACCC E2E tests do.

---

## 9. Use as a test oracle

Can our conformance tests also run against the ACCC Mock Data Holder? Partly. The tests would need these items:

1. **Containers:** the `consumerdataright/mock-data-holder` image and SQL Server 2022 with the EULA accepted ([docker-compose.yml L25-L47](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/DockerCompose/docker-compose.yml#L25-L47), [L93-L105](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/DockerCompose/docker-compose.yml#L93-L105)). UNVERIFIED: the image tags and digests on Docker Hub. I did not query Docker Hub.
2. **Host names:** `mock-data-holder` for the issuer and the links (2.1).
3. **Trust of client certificates:** the gateway trusts one root CA file (2.3). The shipped file is the Mock CDR CA, the same certificate as in the mock Register repository (SHA-1 `3AF1334F5E3831EF194E6F7E9E0FAADA11D5A459`, `openssl x509` on [ca.pem](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CertificateManagement/mtls/ca.pem)). Use the shipped client certificates, or mount the lab CDR root CA and set `RootCACertificate:Path`. UNVERIFIED: that the chain build accepts the lab chain with an intermediate CA.
4. **Trust of the server certificates:** the tests must trust the Mock CDR CA for ports 8000, 8001 and 8002, or replace the PFX files.
5. **Our Register:** set `CdrAuthServer__CdrRegister__SsaJwksUri` to our SSA JWKS, so that DCR accepts our SSA (2.5). The migration that creates the `cdr-register` client reads the same setting ([V003 migration L13](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.Repository/Migrations/20230130011835_V003-RegisterClientSeeddata.cs#L13)).
6. **Software product status:** our software product is not in the seeded table. Send Metadata Update as the client `cdr-register` (a client assertion signed with our SSA key) with `GetDataRecipientsEndpoint` and `Version` set for our Register, or set `CdrAuthServer__CdrRegister__CheckSoftwareProductStatus` to `false` (2.5, 6.10).
7. **The same data in both data holders:** load the ACCC file format with `POST /manage/metadata`, and load equal data into our data holder with our operator interface. The auth server customer files must also agree (7).
8. **Consent with no browser:** after PAR, call `GET /connect/authorize-callback` with `client_id`, `request_uri`, `subject_id` and `account_ids` (4.5), or use headless mode with its fixed user. UNVERIFIED by a run.

Known differences that the tests must expect:

| Area | Difference |
|---|---|
| Endpoints | Only Get Accounts, Get Transactions For Account, Get Customer, Get Status, Get Outages, Get Metrics, Metadata Update and the InfoSec endpoints exist (3) |
| Versions | Get Accounts version 2 and Get Transactions version 1, not 3 and 2; no `x-min-v` (6.1, 6.2) |
| Paths | InfoSec endpoints under `/connect/...` (3.2). Read the discovery document; do not use fixed paths. |
| Tokens | `cnf` holds SHA-1 hexadecimal; `sub` and resource IDs are base64 AES values (6.3, 6.4) |
| Errors | Items 25 and 29 to 38: 400 with no body for a revoked token, 400 for a page out of range, non-CDS codes |
| Headers | `x-fapi-interaction-id` is not validated and can be absent on a 401; no CORS headers (6.2) |
| Data | Default Get Transactions returns no data; `isOwned` is always true (6.8) |
| Status | Reacts to a Register status change only after a Metadata Update call (6.10) |

Where the mock is a good oracle: pagination links and `meta`, the transaction filters, `x-v` error cases, the 404 conditions of Get Transactions (the conditions, not the codes), refresh token rules, introspection, token revocation, arrangement revocation status codes, PAR and request object validation, DCR validation against an SSA.

Where it is weak: products, each endpoint in 3.3, certificate binding, ID formats, error codes, consent UI, Register status timing.

---

## 10. Secrets and keys in the repositories

All are test material. Do not use them outside a lab.

| Item | Location | Note |
|---|---|---|
| Mock CDR CA **private key** | `CertificateManagement/mtls/ca.key`, `ca.pfx` in [mock-data-holder](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CertificateManagement/mtls) and `ca.key` in [authorisation-server](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/CertificateManagement/mtls) | The same CA as the mock Register. Each person can issue a certificate that each ACCC mock trusts. Expires 2031-04-03. |
| Data holder mTLS server key | `CertificateManagement/mtls/server.key`, `server.pfx`; subject CN `mock-data-holder`; expires 2028-08-10 | Password `#M0ckDataHolder#` ([CertificateManagement/README.md](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/CertificateManagement/README.md)) |
| Data holder TLS server key | `CertificateManagement/tls/mock-data-holder.key`, `.pfx`; expires 2028-06-18 | The same password. Copies named `tls-mock-data-holder.pfx` are in 5 projects. |
| Auth server signing keys | [`Source/CdrAuthServer/Certificates/ps256-private.pfx`, `es256-private.pfx`](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/Certificates) | Password in [appsettings.json L55-L64](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer/appsettings.json#L55-L64). They sign access tokens, ID tokens and JARM responses. |
| Auth server TLS, mTLS and UI keys | `CertificateManagement/mtls/mtls-server.key`, `tls/tls-server.key`, `Source/CdrAuthServer.UI/nginx/ssl/authserver-ui.key` | Expire 2028-06-15 (mTLS and TLS; `openssl x509`) |
| ID permanence private key | A literal in `appsettings.json` of 3 data holder projects and of the auth server (6.4, item 16) | Decrypts each `sub`, account ID and transaction ID |
| SQL Server `SA` password | Dockerfile, compose files and `appsettings` files ([docker-compose.yml L98-L102](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/DockerCompose/docker-compose.yml#L98-L102)) | The same password in each ACCC mock |
| Register client certificate and SSA signing certificate (base64 PFX) and their password | [`local.settings.json` L14-L17](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.GetDataRecipients/local.settings.json#L14-L17) of the Azure Function | They let the function act as the Register |
| Data recipient keys | `client.key`, `jwks.key`, `mock-data-recipient.key` in [mock-data-recipient CertificateManagement](https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/CertificateManagement), and `client.pfx`, `client-additional.pfx`, `client-invalid.pfx`, `jwks.pfx` in the test projects of each repo | The test suites use them |
| DSB copies | `security/` in [mock-data-holder-nodejs](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs/blob/7969d2bd1adc01ecd25a3636dfdb9808c3cdaf05/security) has `ca.key` and server keys | UNVERIFIED: whether it is the same Mock CDR CA |

The leaf certificates expire in 2028. A lab that pins the ACCC images must plan for that date.

---

## 11. Test suites of the mocks

Names of test areas only, as candidates for our own test cases.

mock-data-holder banking integration tests ([Source/Banking/CDR.DataHolder.Banking.Tests.IntegrationTests](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Tests.IntegrationTests)), about 80 test methods and about 98 `InlineData` rows:

| File | Areas |
|---|---|
| `US12975_MDH_BankingAPI_GetAccounts.cs` | Success; page and page-size; `is-owned`; `open-status`; `product-category`; scope missing (403); `x-v` supported, not supported (406), not valid (400), missing (400); `x-fapi-auth-date` missing and not valid; `x-fapi-interaction-id` echo; no token, bad token, expired token (401 and `WWW-Authenticate`); different holder of key (401); data recipient participation not active; software product not active; consent for a subset of accounts; same encrypted account IDs on repeated calls |
| `US15350_MDH_BankingAPI_GetTransactions.cs` | Success; `oldest-time`, `newest-time`, `min-amount`, `max-amount`, `text` filters; account ID not valid (404); account not in the consent (404); the header, token, scope, holder-of-key and status cases as above |
| `US12973_MDH_CommonAPI_GetCustomer.cs` | Success for a person and for an organisation; the header, token, scope, holder-of-key and status cases as above |
| `US17613_MDH_DiscoveryAPI_GetOutages.cs`, `US17614_MDH_DiscoveryAPI_GetStatus.cs` | Success; `x-v` cases |

Unit tests: transaction filters (`AccountTransactionsTests.cs`), ID permanence (`IdPermanenceTests.cs`).

authorisation-server integration tests ([Source/CdrAuthServer.IntegrationTests/Tests](https://github.com/ConsumerDataRight/authorisation-server/blob/c40a88a1e50451f4c3e163bb791af6aa5af2b7b0/Source/CdrAuthServer.IntegrationTests/Tests)), about 157 test methods: Authorisation, OIDC configuration, OIDC JWKS, Token, UserInfo, Introspection, PAR, Registration POST, PUT, GET and DELETE, Arrangement Revocation, Token Revocation, JARM. Unit tests: the validators for authorise requests, client assertions, client registration, JWTs, PAR, request objects and token requests, and the Admin and Utility controllers. One E2E test file (`US39327_CdrAuthServer_E2ETests.cs`) drives the UI with Playwright.

Coupling (why we cannot run these suites against our data holder):
- They use the shared package `ConsumerDataRight.ParticipantTooling.MockSolution.TestAutomation` 3.1.0 ([Directory.Packages.props L8](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Directory.Packages.props#L8)). The [mock Register note](2026-09-25-accc-mock-register.md) (9.2) describes it: it writes to the SQL Server databases of the mocks.
- The status tests change the Register data with SQL (`_sqlQueryService.SetStatus`) ([US12973 L520-L521](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Tests.IntegrationTests/US12973_MDH_CommonAPI_GetCustomer.cs#L520-L521)).
- The expected data is the seed file ([TestData/seed-data.json](https://github.com/ConsumerDataRight/mock-data-holder/blob/8fdd4392c6f6072b32782dc5cde2e467a0959549/Source/Banking/CDR.DataHolder.Banking.Tests.IntegrationTests/TestData/seed-data.json)).
- They assert the deviations in section 6, for example the 406 for `x-v: 3` and the 400 for an inactive data recipient.

---

## 12. Items for the mock banking data holder requirements

These are conclusions from the facts above, not facts from the repos.

1. The ACCC mock is a narrow reference: 2 banking operations, at old versions, and no products. Most of the banking API of a first data holder has no ACCC reference. The CDS and its OpenAPI files are the only source for it.
2. No reference mock has an operator interface that creates one customer, one account or one transaction. The ACCC mock has only a full replacement with no authentication. An operator interface for single records is new design.
3. The ACCC mock keeps customers in 3 files in 2 repositories. A design with one source for customers, for both login and resource data, removes that defect.
4. No ACCC product reference data exists. The DSB data has 55 generated products in 11 categories, in CDS field names, for CDS 1.33.0, with a licence that is not clear. Seed files for products need our own data, with the DSB structure as a model.
5. Transaction dates need a decision: the ACCC seed dates are fixed in 2022 and its date offset does not work, so the default query returns nothing.
6. Decide each deviation in section 6 as "follow the CDS" or "copy the mock", and record the decision. The items with the largest effect on tests against the mock are 2 to 4 (versions), 9 (`cnf`), 14 and 15 (IDs and `sub`), 25 and 29 to 31 (errors) and 49 (status timing).
7. Login and OTP need a decision of our own. The mock has a fixed OTP that only the browser checks, and the server trusts the browser for the user and the accounts.
8. As a test peer, the ACCC mock data recipient and the ACCC data holder both need SQL Server and the Mock CDR CA or replaced certificates, and both expect old Register endpoint versions by default.
9. Normalise before a comparison with the mock: `accountId`, `transactionId`, `sub`, `cnf`, `jti`, times, `links` hosts, sort order and error codes.
