# Production gaps

This file lists the shortcuts that are acceptable in the lab and that a production service could not operate with. It is not a plan, and its entries are not tickets.

## Rules

- An entry is a choice that we made because it is easy, and that a production service could not operate with. A function that a project put out of scope is not a gap.
- Add the entry in the change that takes the shortcut. Remove the entry in the change that closes the gap.
- One line for each gap. The file has no history, no priority and no status.
- A gap in `pkg/` is more important than a gap in a mock service. A mock is never for production, but the real components use the `pkg/` libraries as they are.

## Shared libraries (`pkg/`)

| Shortcut | Why it is acceptable in the lab | What a production service needs | Where |
|---|---|---|---|
| A data holder authorisation server keeps its authorisation sessions (pushed requests, and flows in progress) in the memory of the process, with a limit of 10,000 sessions. | The lab runs one instance of each data holder, and a flow that ends at a restart is acceptable there. | A store that all instances share and that stays after a restart, with a limit for each client. | `pkg/authserver/sessions.go` |
| The trace, metric and log providers shut down in sequence with one time limit for all three. | The lab telemetry backend is local and fast, so each provider completes in the limit. | A limit for each provider, or a shutdown of the three at the same time, so that a slow export of one signal does not stop the flush of the others. | `pkg/platform/telemetry`, and `telemetryShutdownTimeout` in each `cmd` |
| The revocation check of an mTLS listener checks only the client certificate, not each certificate in the chain that the CDS requires ([CDS: Certificate Validation](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation)). | The lab never revokes a CA certificate ([PKI: Non-goals](../pki/DESIGN.md#non-goals)), and the mock CDR intermediate signs each lab client certificate. | A check of the revocation status of each certificate in the chain, through the CRL or the OCSP responder that each certificate identifies. | `pkg/cds` revocation check, and the revocation check of the Register |
| The event codec reads only schema version 1 and refuses payload fields that the Go type does not have. No code reads earlier payloads, so a change to an event payload needs a lab reset. | Lab data lasts only until the next reset or CI run, so no stored event has an earlier payload. | Rules for changes to payloads that keep the stored events readable, for example a weak schema or schema versions with code that reads the earlier versions, so that also an older build can read the events of a newer build after a rollback. | `pkg/eventsourcing` JSON codec |

## All services

| Shortcut | Why it is acceptable in the lab | What a production service needs | Where |
|---|---|---|---|
| The operator interface has one static bearer token for all operators, over plain HTTP. The token is public in `compose.yaml`. | The lab has one operator, and its data has no value. The port binds to `127.0.0.1` only, and the interface refuses each request whose `Host` header is not in `OPERATOR_ALLOWED_HOSTS`, so that a web page cannot use DNS rebinding to reach it. | TLS, an identity for each operator, and a record of each operator action. | `pkg/platform/operatorhttp`, `adapter/operatorapi` of each service |
| The operator console has no login and gives the browser no credential, over plain HTTP. Any process on the host can use it. The console of the data holder shows consumer data: names, email addresses and customer numbers, which are also the user identifiers of the consumers. | The console is off by default, the lab has one operator, and all consumer data of the lab is invented. The console refuses each request whose `Host` header is not in `CONSOLE_ALLOWED_HOSTS`, so that a web page cannot use DNS rebinding to reach it. The `127.0.0.1` binding and the refusal of changes from other origins do not stop a web page alone, because DNS rebinding goes around both. | TLS, a login with an identity and a session for each operator, and a record of each operator action ([ADR 0014](adr/0014-web-uis-as-embedded-react-single-page-applications.md)). | `pkg/platform/consolehttp`, `adapter/console` of each service |
| The private signing keys are in the service database with no encryption. | The repository must contain no keys, and the lab keys protect no real data. | Keys in a key management service or a hardware security module, where the service cannot read the private part. | `signing_keys` table of each service |
| The connection to PostgreSQL has no TLS, and the database has no encryption at rest. | The database is on the lab network of one host, and all its data is invented. | TLS to the database with server verification, and encryption at rest. | `PGSSLMODE` in `compose.yaml`. The Register chart can use TLS, but it cannot give the CA to verify the server. |
| No endpoint has a rate limit. | The lab has no traffic that the services do not expect. | Limits for each client and each endpoint. For a data holder, the CDS traffic thresholds give the minimum. | all HTTP listeners |

## Register

| Shortcut | Why it is acceptable in the lab | What a production service needs | Where |
|---|---|---|---|
| The Register reads the private key of the mock CDR intermediate CA from a file, with no key management service or hardware security module. In the lab, any user on the host can read the key file. | The key signs only lab certificates, and the lab data has no value ([ADR 0022](adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md)). `mise run reset` deletes the key. | A CA key in a key management service or a hardware security module, where the service cannot read the private part. | `CDR_INTERMEDIATE_KEY_FILE` in `compose.yaml`, `pki/internal/labdir` |
| The Register chart runs one pod and stops it before it starts the new pod, because the Register is the issuing CA. Each upgrade or restart stops the Register. | The lab has no users that a short stop affects. | More than one instance with one CRL signer, so that an upgrade keeps the Register available. | `charts/register` |

## Data holder

| Shortcut | Why it is acceptable in the lab | What a production service needs | Where |
|---|---|---|---|
| The data holder chart runs one pod and stops it before it starts the new pod, because the authorisation sessions and the online bank site sessions are in the memory of the process. Each upgrade or restart stops the brand, and ends each authorisation flow and each session. | The lab has no users that a short stop affects. | Shared session stores, so that more than one instance serves the brand during an upgrade. | `charts/dataholder` |
| The data holder sends email over SMTP with no TLS and no authentication. | The SMTP server is the mail sink on the lab network. | TLS to the mail server, with authentication. | `adapter/smtpmail` of the data holder |
| The authorisation codes and the refresh tokens are in the database in plain text. | The lab tokens give access to invented data only. The authorisation server library reads the refresh token back at the code exchange, so a hash from the start is not possible. | Encryption of the two values with a key that the database does not have, or a hash of each value after its last read. | `grants` table of the data holder |
| The data holder never removes a grant, also after its code and its refresh token expire. | A reset of the lab removes all grants. | A removal of each grant that no token request can use. | `adapter/grantstore` of the data holder |
| The limits of the one-time password (3 sent, 3 incorrect entries) are for one authorisation flow only. A client can start many flows for one customer. | Only registered lab software products can start a flow, and the mail sink delivers no email. | Limits for each customer across flows, for the emails and for the incorrect entries. | `domain.Authentication` of the data holder |
| The online bank site signs in a consumer with the email address or the customer number only. Its password field accepts any value, and the UI does not send the password. A session has no inactivity timeout. | The site is off by default, all customer data of the lab is invented, and the sign-in page says that any password works. The CDR authentication does not use the site, and it still asks for the email address or the customer number, and a one-time password. | Real authentication of the consumer: a credential that the bank verifies, multi-factor authentication, limits on failed sign-ins, and an inactivity timeout. | `adapter/banksite` of the data holder |
| The online bank site keeps its sessions in the memory of the process, with a limit of 10,000 sessions. A restart ends each session. | The lab runs one instance of each data holder, and a consumer who signs in again after a restart is acceptable there. | A session store that all instances share and that stays after a restart, with a limit for each customer. | `adapter/banksite` of the data holder |
| The online bank sites, the consoles and the CDR pages of the lab share the host `localhost`. Cookies are not isolated by port, so the browser sends the session cookie of each site to each lab service on that host. | All lab ports bind to `127.0.0.1`, each other lab service ignores the cookie, and all customer data of the lab is invented. The cookie name has the brand ID, so the site of each bank reads only its own cookie. | A host of its own for each site, so that the browser sends the session cookie only to that site. | `compose.yaml` |
