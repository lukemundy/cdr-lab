# Mock data recipient: Non-functional requirements

This area gives the requirements for telemetry, secrets, least privilege, restart and health that apply to all of the data recipient.

## Requirements

- <a id="dr-0001"></a>**DR-0001** Telemetry: a span, a metric or a log of the data recipient contains none of these secret values. ([OWASP: Logging Cheat Sheet, data to exclude][owasp-logging-exclude])
  - a password, or the hash of a password
  - the token of a verification link or a reset link, or its hash
  - a session ID
  - an access token, a refresh token, an ID token or an authorisation code
  - a client assertion, a request object, an SSA, or a private key of the data recipient
- <a id="dr-0002"></a>**DR-0002** Telemetry: a span, a metric or a log of the data recipient contains no personal information. For example, it contains no email address, no member details and no ID of a member. The client IP address is the only exception. ([no personal information in telemetry (ADR 0006)][adr-0006]; [Judgement: no member ID in telemetry](#dr-0050))
- <a id="dr-0003"></a>**DR-0003** Telemetry: the span of each operation of the API that signs a member up, signs a member in, confirms a link or resets a password has its outcome, for example `refused` or `succeeded`. It does not say if a member has the email address. ([OWASP ASVS 5.0, 16.3.1][asvs-v16]; [the same response for each address](sign-up.md#dr-0307))
- <a id="dr-0004"></a>**DR-0004** Telemetry: the data recipient sends its telemetry through the telemetry library, with the service name `datarecipient`. ([the telemetry library](../../../pkg/platform/telemetry/DESIGN.md#tele-0001))
- <a id="dr-0005"></a>**DR-0005** Restart: after a restart, the data recipient has the same members, member details and pending sign-ups. A restart ends each session. ([Judgement: sessions in the memory of the process](sign-in.md#dr-0453))
- <a id="dr-0017"></a>**DR-0017** Restart: after a restart, the data recipient has the same signing keys, registrations, consents, connections, tokens, collected data and scheduled runs. A restart ends each authorisation flow in progress. ([Judgement: sessions in the memory of the process](sign-in.md#dr-0453); [Judgement: a restart keeps the data, and not the flows or the sessions](../../dataholder/design/non-functional-requirements.md#dh-0022))
- <a id="dr-0006"></a>**DR-0006** Secrets: the repository contains no private key of the data recipient. The lab creates each key and each certificate when it is set up. ([Judgement: no private key in the repository](../../../pki/DESIGN.md#pki-0093))
- <a id="dr-0010"></a>**DR-0010** Least privilege: a process needs only the settings that it uses. For example, the subcommand `migrate` needs no brand setting, no TLS key and no SMTP setting. ([the settings of a data holder process](../../dataholder/design/non-functional-requirements.md#dh-0010))
- <a id="dr-0011"></a>**DR-0011** Least privilege: when a setting that a process uses is not present, the process does not start. Its error names the setting. ([the missing settings of a data holder process](../../dataholder/design/non-functional-requirements.md#dh-0011))
- <a id="dr-0012"></a>**DR-0012** Least privilege: the usage text of the command gives the settings of each subcommand, and each optional setting. ([the usage text of the data holder](../../dataholder/design/non-functional-requirements.md#dh-0012))
- <a id="dr-0013"></a>**DR-0013** Local run: the data recipient runs in the lab deployment (Docker Compose), with its dependencies. ([the lab deployment](deployment.md#dr-0600))
- <a id="dr-0014"></a>**DR-0014** Health: the data recipient serves its health probes on the internal listener, which is plain HTTP and separate from the web app listener. ([Judgement: a web app listener on a data recipient (ADR 0025)][adr-0025])
- <a id="dr-0018"></a>**DR-0018** Telemetry: a span, a metric or a log of the data recipient contains no CDR data, and no identifier of a consent, an arrangement, a registration or an account. For example, it contains no `cdr_arrangement_id`, no client ID of a registration, no account ID and no amount. ([no CDR data in telemetry](../../../pkg/platform/telemetry/DESIGN.md#tele-0004); [Judgement: no member ID in telemetry](#dr-0050))
- <a id="dr-0019"></a>**DR-0019** Telemetry: each call of the data recipient to the Register or to a bank has [a client span](../../../pkg/cdrclient/DESIGN.md#cdrc-0100) with [its `x-fapi-interaction-id`](../../../pkg/cdrclient/DESIGN.md#cdrc-0101), in the trace of the request or of the collection run that makes the call. ([OpenTelemetry: HTTP client spans][otel-http-client])
- <a id="dr-0020"></a>**DR-0020** Telemetry: each collection run has a span with its trigger and its outcome, for example `scheduled` and `succeeded`. The reason of a failed run is an attribute of the span. ([the reasons of a failed run](data-collection.md#dr-1016))
- <a id="dr-0016"></a>**DR-0016** Performance: on a GitHub-hosted runner of CI, with the lab PostgreSQL, the load of a consent stream of 1,000 events completes within 200 ms. ([Judgement: a load of 1,000 events within 200 ms](#dr-0052))
- <a id="dr-0015"></a>**DR-0015** Performance: in the lab deployment on a GitHub-hosted runner of CI, 95 % of sign-ins complete within 500 ms, with the hash of the password. ([Judgement: a sign-in of 500 ms with the hash](#dr-0051))

## Judgements

### <a id="dr-0050"></a>DR-0050 No member ID in telemetry

- Kind: choice
- Basis: Step 3.3. ADR 0006 permits a keyed HMAC of an identifier ([ADR 0006][adr-0006]). The ID of a member identifies a person, with the database.
- Reason: No ID is the least complex, and it needs no key. A trace ID finds the spans of one request, and the outcome of each operation gives the counts.

### <a id="dr-0051"></a>DR-0051 A sign-in of 500 ms with the hash

- Kind: choice
- Basis: Step 3.3. The CDS gives no time for the app of a data recipient. The parameters of the password hash set most of the time of a sign-in ([the password hash](sign-up.md#dr-0318)).
- Reason: 500 ms leaves time for the hash, the database and the start of the session on a shared runner. A slower sign-in shows that the parameters of the hash are too strong for the lab.

### <a id="dr-0052"></a>DR-0052 A load of 1,000 events within 200 ms

- Kind: choice
- Basis: Step 3.3. The event store loads each event of a stream, and it has no snapshot ([Judgement: event-sourced state (ADR 0003)][adr-0003]). One scheduled run each day gives approximately 365 run events in a consent of 12 months ([one scheduled run each day](data-collection.md#dr-1050)).
- Reason: 1,000 events is more than two times the events of the scheduled runs of a consent of 12 months, so it leaves space for the runs that the member starts. 200 ms leaves time for a page of the web app that loads the consent. When a test shows a slower load, a stream for each collection run is the next step.

[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[adr-0006]: ../../../docs/adr/0006-telemetry.md
[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[asvs-v16]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x25-V16-Security-Logging-and-Error-Handling.md
[owasp-logging-exclude]: https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html#data-to-exclude
[otel-http-client]: https://opentelemetry.io/docs/specs/semconv/http/http-spans/#http-client-span
