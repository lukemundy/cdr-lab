# HTTP server library

The HTTP server library runs the HTTP listeners of a lab service, and it serves the health probes. A listener can serve more than one group of routes, each with its own security policy. A service gives the readiness check, and it puts the probes on a listener that it chooses. The specification of each service gives its readiness conditions, its listeners and their groups of routes.

## Context

```mermaid
flowchart LR
  service["Lab service<br>[External system]"]
  library["HTTP server library<br>[System]"]
  orchestrator["Container orchestrator<br>[External system]"]

  service -- "Runs its listeners and gives its readiness check<br>Go calls" --> library
  orchestrator -- "Checks the liveness and the readiness<br>HTTP" --> library

  classDef system fill:#1168bd,color:#fff,stroke:#0b4884
  classDef external fill:#999,color:#fff,stroke:#6b6b6b
  class library system
  class service,orchestrator external
```

## Library specifications

None.

## Goals

- The same health probes on each lab service.
- The same limits on slow and idle connections on each lab service.
- A stop of each lab service that completes before the container runtime stops the process.
- One security policy for each group of routes, also when groups share a listener.

## Non-goals

- The readiness conditions of a service, for example a current CRL. The specification of each service gives them.
- The listener of the health probes, and its port. The specification of each service gives them.
- The security policy of each group of routes other than the health probes. The library or the service that gives the group gives its policy.

## Health probes

- <a id="srv-0001"></a>**SRV-0001** `GET /livez` responds with 200 while the process runs. The response does not depend on the readiness check. ([Kubernetes: API health endpoints][k8s-health]; [Judgement: the probe paths of the Kubernetes API server](#srv-0003))
- <a id="srv-0002"></a>**SRV-0002** `GET /readyz` responds with 200 when the readiness check of the service passes, and with 503 when it fails. For example, a service that cannot connect to its database responds with 503. ([Kubernetes: API health endpoints][k8s-health]; [Judgement: the probe paths of the Kubernetes API server](#srv-0003))
- <a id="srv-0015"></a>**SRV-0015** A 503 response of `GET /readyz` does not give the cause of the failure. For example, the body does not contain the error of the database. ([Judgement: a readiness response with no cause](#srv-0016))

- <a id="srv-0011"></a>**SRV-0011** The health probes do not check the `Host` header. For example, a probe with the IP address of the pod as its `Host` gets the probe response. ([Kubernetes: HTTP probes][k8s-http-probes]; [Judgement: health probes with no `Host` check and no authentication](#srv-0013))
- <a id="srv-0012"></a>**SRV-0012** The health probes need no authentication. ([Kubernetes: HTTP probes][k8s-http-probes]; [Judgement: health probes with no `Host` check and no authentication](#srv-0013))

## Groups of routes

- <a id="srv-0009"></a>**SRV-0009** The security policy of a group of routes applies only to the requests for the paths of that group. For example, the `Host` allowlist of the operator interface does not apply to `GET /readyz` on the same listener. ([Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])
- <a id="srv-0010"></a>**SRV-0010** A request for a path that no group of routes of the listener serves gets 404. ([RFC 9110, section 15.5.5][rfc9110-404]; [Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])

## Connections

- <a id="srv-0004"></a>**SRV-0004** A server stops a request when the client does not send all the headers in 5 seconds. ([Judgement: limits that stop slow and idle connections](#srv-0008))
- <a id="srv-0005"></a>**SRV-0005** A server stops a request when the client does not send all of the request, with its body, in 30 seconds. ([Judgement: limits that stop slow and idle connections](#srv-0008))
- <a id="srv-0006"></a>**SRV-0006** A server stops a response that it does not write in 60 seconds from the end of the request headers. ([Judgement: limits that stop slow and idle connections](#srv-0008))
- <a id="srv-0007"></a>**SRV-0007** A server closes a keep-alive connection that has no request for 120 seconds. ([Judgement: limits that stop slow and idle connections](#srv-0008))
- <a id="srv-0014"></a>**SRV-0014** When the service stops, its servers accept no new connection, and let the open requests complete for 5 seconds at most. ([Judgement: a stop of 5 seconds at most](#srv-0017))

## Judgements

### <a id="srv-0003"></a>SRV-0003 The probe paths of the Kubernetes API server

- Kind: choice
- Basis: Step 3.2. The CDS and the ACCC say nothing about health probes. The Kubernetes API server uses `/livez` and `/readyz`, and `/healthz` is deprecated there ([Kubernetes: API health endpoints][k8s-health]).
- Reason: The paths are an industry norm. All lab services then have the same probes.

### <a id="srv-0008"></a>SRV-0008 Limits that stop slow and idle connections

- Kind: choice
- Basis: Step 3.2. The CDS and the ACCC give no connection limits. Go `net/http` has no read, write or idle limit by default, so a slow client can keep a connection open with no end ([Go: net/http Server][go-http-server]).
- Reason: The limits stop a slow request body and an idle connection from holding resources, also in a real component that copies the library. They are much longer than each request of the lab, and no response of the lab is a stream.

### <a id="srv-0013"></a>SRV-0013 Health probes with no `Host` check and no authentication

- Kind: choice
- Basis: Step 3.2. The kubelet sends an HTTP probe to the IP address of the pod, unless the probe gives a host ([Kubernetes: HTTP probes][k8s-http-probes]). The kubelet sends no credential that the service can check.
- Reason: A `Host` check then refuses the probes of a usual Kubernetes deployment. The probes give only the state of the process, so a DNS rebinding attack gets no data of value.

### <a id="srv-0016"></a>SRV-0016 A readiness response with no cause

- Kind: choice
- Basis: Step 3.2. The kubelet uses only the status code of an HTTP probe ([Kubernetes: HTTP probes][k8s-http-probes]). An error message must not show internal data to a client ([OWASP: Error Handling][owasp-error-handling]).
- Reason: The probes need no authentication, so each client can read the response. The cause can contain internal data, for example a host name of the database.

### <a id="srv-0017"></a>SRV-0017 A stop of 5 seconds at most

- Kind: choice
- Basis: Step 3.2. Docker sends SIGKILL 10 seconds after the stop signal of a Linux container, by default ([Docker: docker container stop][docker-stop]).
- Reason: An open request then usually completes. The service also has time for its other stop steps, for example the export of its telemetry, before SIGKILL.

## Withdrawn

None.

[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[docker-stop]: https://docs.docker.com/reference/cli/docker/container/stop/
[k8s-health]: https://kubernetes.io/docs/reference/using-api/health-checks/
[go-http-server]: https://pkg.go.dev/net/http#Server
[k8s-http-probes]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#http-probes
[owasp-error-handling]: https://cheatsheetseries.owasp.org/cheatsheets/Error_Handling_Cheat_Sheet.html
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5
