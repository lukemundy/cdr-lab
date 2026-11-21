# 0026. Send metrics only for the state of a process, and give the detail of each request in traces

- Date: 2026-11-21

## Context

This ADR amends [ADR-0006](0006-telemetry.md). It replaces the statement that the HTTP server metrics do not have `server.address` or `server.port`, because the services send no HTTP request metrics after this decision.

A deployment of the lab can send its telemetry to a backend that other workloads also use. Such a backend has limits on the number of active series and on the samples each second. On 2026-11-03, the Register and four data holders took a shared Grafana Cloud stack over its limit of 15,000 active series. The infrastructure metrics of the other workloads then stopped for approximately 45 minutes.

The number of series came from the instrumentation libraries: otelhttp, otelpgx and otelriver. Each of these records histograms for each route, method, status code, SQL operation or job kind. After one run of the conformance suite and the browser tests, a lab with the Register and four data holders had 16,187 series. Of these, 13,939 were HTTP server histograms. Thus the number of series increases with each route that a client calls, with each new endpoint, query or job, and with each restart, because each process start had a new `service.instance.id`. The one-shot `migrate` and `seed` processes also sent approximately 75 series each, under an instance that stops after some seconds.

The traces have the route, the status, the duration and the error of each request, query and job. A backend that supports TraceQL metrics, for example Tempo, calculates rates, error rates and latency percentiles from the spans, with any attribute. The lab records each span, unless a caller sends a `traceparent` with the sampled flag off. The lab has no alerts and no service level objectives, and its traffic is small.

The real options were:

1. **Keep the request metrics with fewer attributes and buckets.** An allowlist of instruments and attribute keys gives approximately 60 series for each process. The rate, the latency and the 5xx rate of each service are then available as metrics, at any sample rate and with the retention of the metrics backend. The number of series still changes when new error codes occur. Each new instrumentation library adds metrics until somebody adds it to the allowlist.
2. **Send metrics only for the state of the process.** The Go runtime, the connection pool and the number of active HTTP requests are state that a trace cannot give. They give 23 series for each process, and the number does not change with traffic, routes, errors or features. The request detail is available only from the traces.
3. **Keep all metrics, and limit them in the collector or the backend.** This needs work from each operator, and the lab then sends data that the backend discards. A collector of the home cluster discards all OTLP data of the lab for this reason.

## Decision

The lab services send metrics only for the state of the process. All detail of requests, queries and jobs is in the traces. This applies to each service that uses `pkg/platform/telemetry`. It does not change the traces or the logs.

- The global meter provider makes no metrics. Thus an instrumentation library that uses the global provider, also a library that a later change adds, makes no metrics. `pkg/platform/telemetry` gives the real meter provider only to the sources of process state: the Go runtime, the connection pool statistics and its own count of active HTTP server requests.
- Each metric has a fixed set of attributes with a fixed set of values. A metric that a trace can give, for example a request duration, is not a metric.
- A one-shot subcommand sends no metrics. It is short-lived, so its series show nothing about capacity.
- `service.instance.id` is the host name of the process, with a random UUID only when the host name is not available. In Kubernetes the host name is the pod name, and in Docker it is the container ID. Thus a restart of the container gives the same series.
- A new metric is a change to the specification of the telemetry library. The change gives the reason why traces cannot give the value, and changes the expected number of series where it is published.

We chose option 2 because it is the only option where the number of series stays the same as the lab grows. Option 1 gives metrics that the traces of the lab already give, and the lab has no alerts that need them. Option 3 moves the problem to each operator.

## Consequences

- The number of series is known in advance: 23 for each long-running process. A shared backend can plan for it.
- A rate, an error rate or a latency for each route needs a TraceQL metrics query, or a backend that makes span metrics. A backend that cannot query traces shows only the state of each process.
- A request that a caller sends with the sampled flag off has no span and no metric. Only the count of active requests includes it.
- Metrics cannot give a long trend of request rates, because the retention of the traces is usually shorter than that of metrics.
- An upgrade of the runtime or pool instrumentation can add a metric. A test of the exact set of metrics then fails, and somebody must accept the new series or remove them.
- A Kubernetes deployment with more than one container in a pod gives each container the same host name. Each long-running process must then be in its own pod, which the charts of the lab do.
