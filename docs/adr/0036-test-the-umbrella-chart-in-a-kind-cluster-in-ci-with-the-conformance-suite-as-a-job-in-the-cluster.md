# 0036. Test the umbrella chart in a kind cluster in CI, with the conformance suite as a Job in the cluster

- Date: 2026-11-23

## Context

The umbrella chart deploys the lab to a Kubernetes cluster ([ADR 0033](0033-deploy-the-lab-to-a-cluster-with-one-umbrella-chart-that-derives-each-brand-from-its-register-record.md)). CI renders the charts and checks them against the Kubernetes schemas, but it installs no chart. A defect in the order of the Jobs, the RBAC or the certificates then shows only in the cluster of a team.

These facts constrain the test:

- CI uses GitHub-hosted runners only. The standard Linux runner of a public repository has 4 vCPUs and 16 GB of memory ([GitHub: GitHub-hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)).
- kind runs a Kubernetes cluster in Docker, and it can load local images into the nodes ([kind: quick start](https://kind.sigs.k8s.io/docs/user/quick-start/)). CI already saves the images of each commit to a file.
- The conformance suite runs in the `conformance` compose service on the lab network. The data holder and the Register call back to the suite: the seed gives the host `conformance` for the JWKS of the software products of the tests. Thus the suite must be reachable from the services by that name, with a certificate for that name.

The real options were:

1. **kind, with the suite as a Job in the cluster and a Service with the name `conformance`.** The names, the seed and the certificate names stay the same as in compose.
2. **kind, with the suite on the runner.** The suite can reach the services through port-forward, but the services cannot reach the suite by the name `conformance` without more DNS and routing.
3. **k3d.** It is similar to kind, but it installs an ingress controller and a load balancer by default, which the test must turn off.

## Decision

A CI job installs the umbrella chart in a kind cluster on a standard hosted runner, and runs the conformance suite against it as a Job in the cluster (option 1).

- The job uses the images that the lab jobs of the same commit tested. It builds no other image of a service.
- The job installs the parts that the chart needs and does not contain: cert-manager with a CA issuer for the public TLS certificates, PostgreSQL with the set-up SQL files of the lab, and the mail sink. It uses a plain PostgreSQL, not an operator, because the lab SQL files are tested and need no operator.
- The suite gets its CDR client certificates from the Register, as each other part of the deployment does ([ADR 0032](0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md)).
- The job runs no Gateway. The route templates of the chart get only the render and schema checks of the chart lint.
- The compose conformance job stays. The suite then runs in two environments, and a test must pass in both.

## Consequences

- CI finds a defect of the umbrella chart, its Jobs and its certificates before a team installs it.
- CI uses more runner time for each change to the charts, the services, `pkg/`, `cli/` or the suite. The time and the memory of the job have no measurement at the date of this ADR.
- A test of the suite cannot depend on a compose-only name, port or file. The settings of the suite come from the environment of the Job.
- The set-up of the fixture certificates of the suite is different in compose and in the cluster while compose issues them offline.
- The routes and the oauth2-proxy option are not tested in CI.
