# Mock CDR Register: Non-functional requirements

This area gives the non-functional requirements that apply to all of the Register. They are for the lab. The CDS non-functional requirements apply to data holders, not to the Register ([CDS: Non-functional Requirements][cds-nfrs]). A non-functional requirement of one area is in that area.

## Propagation

- <a id="reg-0001"></a>**REG-0001** Propagation: when a participant status changes, each Register API gives the new status within 1 second. For example, 1 second after the lab operator makes a software product INACTIVE, Get Software Products Statuses gives INACTIVE for it. ([CDS: Metadata Cache Management][cds-metadata-cache]; [Judgement: each API gives a new status within 1 second](#reg-0020))

## Performance

- <a id="reg-0002"></a>**REG-0002** Performance: on a development workstation at 50 requests each second, 95 % of API calls complete within 200 ms. ([Judgement: a target of 200 ms, below the lowest CDS threshold](#reg-0021))

## Restart

- <a id="reg-0003"></a>**REG-0003** Restart: after a restart, the Register has the same values of these data. ([Judgement: a restart keeps the participants, keys, revocations and CRL number](#reg-0022))
  - the participants and their statuses
  - the signing keys
  - the revocations of CDR certificates
  - the CRL number

## Telemetry

The Register also obeys the [Telemetry library](../../../pkg/platform/telemetry/DESIGN.md), for example [traces, metrics and logs with OpenTelemetry](../../../pkg/platform/telemetry/DESIGN.md#tele-0001), [no CDR data and no personal information in telemetry](../../../pkg/platform/telemetry/DESIGN.md#tele-0004) and [route patterns in place of raw URLs](../../../pkg/platform/telemetry/DESIGN.md#tele-0005).

## Secrets

- <a id="reg-0004"></a>**REG-0004** Secrets: the repository contains no private key of the Register. The lab set-up or the first start of the Register creates each key. For example, the Register creates its access token signing key at its first start. ([Judgement: no private key in the repository](../../../pki/DESIGN.md#pki-0093))

## Portability

- <a id="reg-0005"></a>**REG-0005** Portability: a real data recipient or a real data holder changes from the mock Register to the production Register when it changes only these settings. ([Judgement: a real component changes the Register with configuration only](#reg-0023))
  - the base URIs
  - the CA bundles
  - the client certificate
  - the signing keys
  - the identifiers

## Conformance

- <a id="reg-0006"></a>**REG-0006** Conformance: the conformance suite checks each requirement of the Register that the CDR defines, through the APIs only. ([Judgement: the conformance suite tests only CDR conformance (ADR 0019)][adr-0019])

## Local run

- <a id="reg-0007"></a>**REG-0007** Local run: the Register runs in the lab deployment (Docker Compose), with its dependencies. ([Judgement: Docker Compose for the lab deployment](#reg-0026))

## Health

- <a id="reg-0008"></a>**REG-0008** Health: the Register serves its health probes on a plain HTTP listener that is separate from the CDR API listeners. ([the liveness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0001); [the readiness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0002); [Judgement: a separate listener for the health probes](#reg-0025))
- <a id="reg-0009"></a>**REG-0009** Health: when a process that runs the APIs has no current CRL, its readiness probe fails. For example, after the next update time of the newest CRL, `GET /readyz` gives 503. ([the readiness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0002); [the CRL check fails with no CRL that it can use](../../../pkg/cds/DESIGN.md#cdsl-0051); [Judgement: an API process with no current CRL is not ready](#reg-0024))
- <a id="reg-0010"></a>**REG-0010** Health: a process that runs only the background jobs is ready when it can connect to the database. It does not need a current CRL. ([Judgement: background jobs with River (ADR 0010)][adr-0010])

## Judgements

### <a id="reg-0020"></a>REG-0020 Each API gives a new status within 1 second

- Kind: choice
- Basis: Step 3.3. A data holder must react to a status change within 5 minutes ([CDS: Metadata Cache Management][cds-metadata-cache]). The CDS gives no time for the Register.
- Reason: A Register that adds no delay leaves all of the 5 minutes to the data holder. A test of a data holder then does not wait for the Register.

### <a id="reg-0021"></a>REG-0021 A target of 200 ms, below the lowest CDS threshold

- Kind: choice
- Basis: Step 3.3. The CDS gives no response time for the Register. Its thresholds for data holder APIs are 1,000 ms or more ([CDS: Non-functional Requirements][cds-nfrs]).
- Reason: A target below those thresholds shows a slow change early. A development workstation can send 50 requests each second.

### <a id="reg-0022"></a>REG-0022 A restart keeps the participants, keys, revocations and CRL number

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the restart of the Register. A CRL number is never used two times ([RFC 5280, section 5.2.3][rfc5280-crl-number]).
- Reason: A participant expects the same participants, statuses and keys after a restart, as at the production Register. A revocation that a restart loses makes a revoked certificate valid again.

### <a id="reg-0023"></a>REG-0023 A real component changes the Register with configuration only

- Kind: choice
- Basis: Step 3.1. A real component connects to the real CDR with a change of its configuration only ([mock component, real component](../../../docs/spec/glossary.md#mock-component-real-component)).
- Reason: A real component that the lab tests must also work with the production Register. When the mock Register differs from production only in its configuration, a test in the lab is also valid for production.

### <a id="reg-0024"></a>REG-0024 An API process with no current CRL is not ready

- Kind: choice
- Basis: Step 3.3. The Kubernetes readiness probe tells the platform when a process can serve traffic ([Kubernetes: API health endpoints][k8s-health]).
- Reason: With no current CRL, the Register refuses each client certificate, so its mTLS APIs cannot serve. A probe that fails keeps the traffic away until the Register signs a new CRL.

### <a id="reg-0025"></a>REG-0025 A separate listener for the health probes

- Kind: choice
- Basis: Step 3.3. The CDS gives no health probes.
- Reason: A separate listener keeps the probes out of the CDR APIs. A probe then needs no TLS and no client certificate.

### <a id="reg-0026"></a>REG-0026 Docker Compose for the lab deployment

- Kind: choice
- Basis: Step 3.3. The lab deployment and CI are the only environments of the lab.
- Reason: Docker Compose starts all services and their dependencies on one workstation with one command. This is the least complex deployment for the lab and for CI.

[cds-nfrs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#non-functional-requirements
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[rfc5280-crl-number]: https://www.rfc-editor.org/rfc/rfc5280#section-5.2.3
[k8s-health]: https://kubernetes.io/docs/reference/using-api/health-checks/
[adr-0010]: ../../../docs/adr/0010-background-jobs-with-river.md
[adr-0019]: ../../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
