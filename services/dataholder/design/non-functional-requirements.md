# Mock data holder: Non-functional requirements

This area gives the non-functional requirements that apply to all of the data holder. They are for the lab. A non-functional requirement of one area is in that area.

## Performance

- <a id="dh-0001"></a>**DH-0001** Performance: on a development workstation at 50 requests each second, 95 % of API calls complete within 200 ms. ([CDS: Performance Requirements][cds-performance]; [Judgement: a response time target for the lab](#dh-0020))
- <a id="dh-0002"></a>**DH-0002** Performance: when the traffic is above a CDS traffic threshold, the data holder does not throttle or refuse a request. For example, the data holder gives no 429 response. ([CDS: Traffic Thresholds][cds-traffic-thresholds]; [Judgement: no throttle at the traffic thresholds](#dh-0021))

## Restart

- <a id="dh-0003"></a>**DH-0003** Restart: after a restart, the data holder has the same values of these data. ([Judgement: the state that a restart keeps](#dh-0022))
  - the products
  - the registrations
  - the signing keys
  - the customers, the accounts and the transactions
  - the arrangements and their ends
  - the records of disclosures
  - the notices, with their attempts and their outcomes
  - the subject identifiers and the resource identifiers
  - the refresh tokens
  - the revoked access tokens
- <a id="dh-0004"></a>**DH-0004** Restart: when the data holder restarts during an authorisation flow, the flow can fail. The consumer then starts a new authorisation flow. ([Judgement: the state that a restart keeps](#dh-0022))
- <a id="dh-0005"></a>**DH-0005** Restart: a restart ends each session of the online bank site. The consumer then signs in again. ([Judgement: the state that a restart keeps](#dh-0022); [the session of the online bank site](../DESIGN.md#session))
- <a id="dh-0006"></a>**DH-0006** Restart: after a restart, the data holder gets the statuses of the software products from the Register again. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [the get of the participant statuses](registration.md#dh-0679))

## Telemetry

The data holder also obeys the [Telemetry library](../../../pkg/platform/telemetry/DESIGN.md), for example [traces, metrics and logs with OpenTelemetry](../../../pkg/platform/telemetry/DESIGN.md#tele-0001) and [route patterns in place of raw URLs](../../../pkg/platform/telemetry/DESIGN.md#tele-0005).

- <a id="dh-0007"></a>**DH-0007** Telemetry: a span, a metric or a log of the data holder contains none of these secret values. ([OWASP: Logging Cheat Sheet, data to exclude][owasp-logging-exclude])
  - an SSA
  - a request JWT, a client assertion or a request object
  - a `request_uri` or an authorisation code
  - an access token, a refresh token or an ID token
  - a self-signed JWT or a `cdr_arrangement_jwt`
  - a session ID of the online bank site
  - a password
- <a id="dh-0008"></a>**DH-0008** Telemetry: a span, a metric or a log of the data holder contains none of these values. For example, a log can contain the `x-fapi-interaction-id`, because it is not in the list. ([no CDR data and no personal information in telemetry](../../../pkg/platform/telemetry/DESIGN.md#tele-0004))
  - a customer number
  - a name or an email address
  - a one-time password
  - a `sub`
  - account data or transaction data
  - an `accountId` or a `transactionId`

## Secrets

- <a id="dh-0009"></a>**DH-0009** Secrets: the repository contains no private key of the data holder. The lab creates each key and each certificate when it is set up. ([Judgement: no private key in the repository](../../../pki/DESIGN.md#pki-0093))

## Least privilege

- <a id="dh-0010"></a>**DH-0010** Least privilege: a process that runs only the APIs, or only the workers, needs only the settings that it uses. For example, a process that runs only the workers needs no TLS key of the APIs, no operator token and no SMTP setting. ([Judgement: background jobs with River (ADR 0010)][adr-0010]; [Judgement: only the settings that a process uses](#dh-0023))
- <a id="dh-0011"></a>**DH-0011** Least privilege: when a setting that a process uses is not present, the process does not start. Its error names the setting. ([Judgement: only the settings that a process uses](#dh-0023))
- <a id="dh-0012"></a>**DH-0012** Least privilege: the usage text of the command gives the settings of each subcommand. ([Judgement: only the settings that a process uses](#dh-0023))

## Portability

- <a id="dh-0013"></a>**DH-0013** Portability: a real client changes from the mock data holder to a production data holder when it changes only these settings. ([Judgement: a real client changes the data holder with configuration only](#dh-0026))
  - the base URIs
  - the CA bundles
  - the client certificate
  - the signing keys
  - the identifiers

## Conformance

- <a id="dh-0014"></a>**DH-0014** Conformance: the conformance suite checks each requirement of the data holder that the CDR defines, through the APIs only. ([Judgement: the conformance suite tests only CDR conformance (ADR 0019)][adr-0019])
- <a id="dh-0015"></a>**DH-0015** Conformance: the conformance suite acts as a participant, in these ways. It does not depend on the name of a preset brand. ([Judgement: the conformance suite tests only CDR conformance (ADR 0019)][adr-0019]; [Judgement: a conformance suite that acts as a participant](#dh-0024))
  - It gets the base URIs and the brand name of the brand from the Register.
  - It gets each SSA from the Register.
  - It acts as the browser of the consumer: it reads each consumer page and sends each form.
  - It gets the tokens of an arrangement through the authorisation flow.
  - It creates customer data with the operator interface.
  - It serves the CDR arrangement revocation endpoint of a software product, and gets each notice there.
  - It changes the statuses of software products with the operator interface of the Register.

## Local run

- <a id="dh-0016"></a>**DH-0016** Local run: the data holder runs in the lab deployment (Docker Compose), with its dependencies. ([Judgement: Docker Compose for the lab deployment](#dh-0027))

## Health

- <a id="dh-0017"></a>**DH-0017** Health: the data holder serves its health probes on a plain HTTP listener that is separate from the CDR API listeners. ([the liveness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0001); [the readiness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0002); [Judgement: a separate listener for the health probes](#dh-0025))

## Judgements

### <a id="dh-0020"></a>DH-0020 A response time target for the lab

- Kind: choice
- Basis: Step 3.3. For 95 % of calls each hour, the CDS thresholds are 1,000 ms or more ([CDS: Performance Requirements][cds-performance]). The CDS gives them for production, not for a lab ([CDS: Non-functional Requirements][cds-nfrs]).
- Reason: A target below the CDS thresholds shows a slow change before it can go above them. A development workstation can send 50 requests each second.

### <a id="dh-0021"></a>DH-0021 No throttle at the traffic thresholds

- Kind: choice
- Basis: Step 3.3. The CDS permits a data holder to throttle or refuse calls above the traffic thresholds, and does not require it ([CDS: Traffic Thresholds][cds-traffic-thresholds]).
- Reason: The lab has no such load. A 429 response would make the conformance tests slow and not reliable.

### <a id="dh-0022"></a>DH-0022 The state that a restart keeps

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the restart of a data holder.
- Reason: A data recipient expects its registrations, arrangements and tokens after a restart, as at a production data holder. An authorisation flow and a session of the online bank site are short, and the consumer can start them again. Memory for them is the least complex.

### <a id="dh-0023"></a>DH-0023 Only the settings that a process uses

- Kind: choice
- Basis: Step 3.2. The principle of least privilege applies to the secrets of a process.
- Reason: Each deployment that has a secret that it does not use is one more location where the secret can be disclosed. A process that stops at a missing setting, and the usage text, tell the operator which settings each process needs.

### <a id="dh-0024"></a>DH-0024 A conformance suite that acts as a participant

- Kind: choice
- Basis: Step 3.3. [ADR 0019][adr-0019] limits the suite to CDR conformance, but does not say how the suite finds a data holder and its data.
- Reason: A participant has only the Register and the CDR APIs, so the suite then proves what a participant sees. The suite runs against each brand with no change, also when a catalogue changes.

### <a id="dh-0025"></a>DH-0025 A separate listener for the health probes

- Kind: choice
- Basis: Step 3.3. The CDS gives no health probes.
- Reason: A separate listener keeps the probes out of the CDR APIs. A probe then needs no TLS and no client certificate.

### <a id="dh-0026"></a>DH-0026 A real client changes the data holder with configuration only

- Kind: choice
- Basis: Step 3.1. A real component connects to the real CDR with a change of its configuration only ([mock component, real component](../../../docs/spec/glossary.md#mock-component-real-component)).
- Reason: A real component that the lab tests must also work with production data holders. When the mock data holder differs from production only in its configuration, a test in the lab is also valid for production.

### <a id="dh-0027"></a>DH-0027 Docker Compose for the lab deployment

- Kind: choice
- Basis: Step 3.3. The lab deployment and CI are the only environments of the lab.
- Reason: Docker Compose starts all services and their dependencies on one workstation with one command. This is the least complex deployment for the lab and for CI.

[cds-performance]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#performance-requirements
[cds-nfrs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#non-functional-requirements
[cds-traffic-thresholds]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#traffic-thresholds
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[owasp-logging-exclude]: https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html#data-to-exclude
[adr-0010]: ../../../docs/adr/0010-background-jobs-with-river.md
[adr-0019]: ../../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
