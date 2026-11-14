# 0006. Use OpenTelemetry for traces, metrics and logs

- Date: 2026-10-04

## Context

Every service needs telemetry. If each service handles telemetry in its own way, the signals are different in each service, and each service must solve the same problems again. We need one consistent way for services to make, protect and export telemetry.

Telemetry must not contain sensitive data. A telemetry backend has different access controls and retention from the service that owns the data. In a CDR lab, the sensitive data is CDR data and personally identifiable information. All the data in the lab is mocked, but the rule applies to it as it does to real data.

## Decision

### Signals and export

- Services use OpenTelemetry for traces, metrics and logs, with OTLP export and the standard `OTEL_*` environment variables.
- `pkg/platform/telemetry` sets up the SDK. It also contains the other telemetry code and helpers that services share.
- A service sets up telemetry at the start of `main`, with the service name as a constant. `OTEL_SERVICE_NAME` and `OTEL_RESOURCE_ATTRIBUTES` cannot change the service name.
- A service exports a signal only when `OTEL_EXPORTER_OTLP_ENDPOINT`, the endpoint for the signal, or `OTEL_<SIGNAL>_EXPORTER` is set. If none is set, the service drops the signal with no errors, and writes one log line that tells you. This is different from the OpenTelemetry default (OTLP to `localhost`), which gives an error for each export.
- Metrics export each 10 seconds, not each 60 seconds (the OpenTelemetry default). `OTEL_METRIC_EXPORT_INTERVAL` can change this.
- Logs go to two destinations. The logs on stdout should be easy for a person to read. The logs that go to OTLP are structured for a telemetry backend to ingest.
- A telemetry failure never stops a service or changes its exit code. Telemetry errors go to stdout.
- Services propagate W3C trace context only, not baggage. Baggage from callers is not trusted, and outbound calls would send it to third parties. `OTEL_PROPAGATORS` can change this.

### Instrumentation

- Instrument centrally: in HTTP middleware, database driver hooks, job queue middleware and the log bridge. Business code contains no bespoke spans or metrics, and `domain` contains no telemetry.
- Each HTTP server and each one-shot subcommand (for example `migrate`) uses the instrumentation in `pkg/platform/telemetry`.
- A one-shot subcommand has one CLI span, as the OpenTelemetry [CLI span convention](https://github.com/open-telemetry/semantic-conventions/blob/v1.44.0/model/cli/spans.yaml) (development stability) gives. The span name is `<executable> <subcommand>`, for example `register seed`, which the convention permits.
- A database span is never a root span. The database instrumentation makes a span only when a parent span exists.
- Use OpenTelemetry semantic conventions for attribute names, also in log attributes.
- HTTP server metrics do not have `server.address` or `server.port`. These attributes have high cardinality, and such attributes do not belong on metrics. Spans still have them.
- The telemetry resource identifies the build: `service.version` is the short commit hash, `vcs.ref.head.revision` is the full commit hash, and `app.build.time` and `app.build.id` (the CI run ID, or `local`) come from `pkg/platform/buildinfo`. The build sets these values with `-ldflags -X`. A commit hash ends in `-dirty` when the working tree has changes that are not committed.

### Sensitive data

- CDR data and personally identifiable information must never appear in logs, metrics or traces. This includes mocked data.
- If you must record an identifier (for example an email address), record a keyed HMAC-SHA256 of it.
- URL paths and query strings can contain sensitive data (for example account IDs). Thus record route patterns and not raw URLs. `pkg/platform/telemetry` replaces `url.path` on each span with `http.route`, or with `REDACTED` when no route matches.
- Database spans contain the SQL text with placeholders, but no query arguments. `pkg/platform/database` removes the messages of query errors from spans ([ADR 0002](0002-data-store.md)).
- Sensitive values must never be on a command line. Thus a CLI span can have `process.command_args`.

## Consequences

- All services give the same signals, and one change to `pkg/platform` changes them all.
- Business code cannot add a span or a metric for one operation. If central instrumentation does not show enough, improve the central instrumentation.
- Database queries that run with no parent span make no spans.
- A service with no OTLP configuration sends nothing, and gives no error. To find this condition, look for the log line at startup.
- Redaction makes spans less useful for diagnosis: a span shows the route and the SQLSTATE, not the URL or the error message. The caller of a query still gets the full error, and must not log it as it is.
- A service that records the HMAC of an identifier needs a secret key, and the same key gives the same value in each service that must correlate the identifier.
- A change of telemetry backend needs a change to the `OTEL_*` configuration only.
