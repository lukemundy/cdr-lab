# HTTP server library

The HTTP server library runs the HTTP listeners of a lab service, and it serves the health probes. A service gives the readiness check, and it puts the probes on a listener that it chooses. The specification of each service gives its readiness conditions and its listeners.

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

## Non-goals

- The readiness conditions of a service, for example a current CRL. The specification of each service gives them.
- The listener of the health probes, and its port. The specification of each service gives them.

## Health probes

- <a id="srv-0001"></a>**SRV-0001** `GET /livez` responds with 200 while the process runs. The response does not depend on the readiness check. ([Kubernetes: API health endpoints][k8s-health]; [Judgement: the probe paths of the Kubernetes API server](#srv-0003))
- <a id="srv-0002"></a>**SRV-0002** `GET /readyz` responds with 200 when the readiness check of the service passes, and with 503 when it fails. For example, a service that cannot connect to its database responds with 503. ([Kubernetes: API health endpoints][k8s-health]; [Judgement: the probe paths of the Kubernetes API server](#srv-0003))

## Judgements

### <a id="srv-0003"></a>SRV-0003 The probe paths of the Kubernetes API server

- Kind: choice
- Basis: Step 3.2. The CDS and the ACCC say nothing about health probes. The Kubernetes API server uses `/livez` and `/readyz`, and `/healthz` is deprecated there ([Kubernetes: API health endpoints][k8s-health]).
- Reason: The paths are an industry norm. All lab services then have the same probes.

## Withdrawn

None.

[k8s-health]: https://kubernetes.io/docs/reference/using-api/health-checks/
