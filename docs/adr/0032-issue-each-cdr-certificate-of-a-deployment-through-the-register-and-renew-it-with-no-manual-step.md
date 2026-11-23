# 0032. Issue each CDR certificate of a deployment through the Register, and renew it with no manual step

- Date: 2026-11-23

## Context

This ADR amends [ADR-0022](0022-run-the-mock-cdr-intermediate-ca-in-the-register.md). It replaces the statement "The `pki` command keeps the set-up: it makes the CAs, and it issues the certificates of the lab services and tests before the lab starts". The set-up still makes the CAs, but the CDR certificates of the services and the tests come from the Register.

A team that keeps the lab in a Kubernetes cluster for many months needs CDR certificates that stay valid with no person on a host. These facts constrain the design:

- A CDR certificate is valid for 13 months ([CA library: the validity of a CDR certificate](../../pkg/ca/DESIGN.md#ca-0026)). A deployment that lives longer than that needs a renewal.
- The set-up issues the CDR certificates of the lab services offline, with the intermediate key on the host. The Register then does not check the names of the CSR, and it records no event of the issue.
- The Register issue path checks the CSR against the participant, registers a client certificate, and records the issue of a server certificate ([Register: certificate management](../../services/register/design/certificate-management.md#reg-0600)). The Register never gets the private key of a certificate that it issues ([Register: no private key](../../services/register/design/certificate-management.md#reg-0603)).
- The Register needs a CDR server certificate for its own mTLS listener, and it holds the intermediate key.
- The services load their certificate files one time, at start. A changed file needs a restart.
- Kubernetes updates the files of a mounted Secret after the Secret changes, except for a `subPath` mount ([Kubernetes: Secrets as files](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod)).

The real options were:

1. **cert-manager with its CA issuer.** cert-manager renews the certificates. But the CA issuer needs the intermediate key in a second Secret, in the namespace of the issuer ([cert-manager: CA](https://cert-manager.io/docs/configuration/ca/)). It does not do the checks, the events or the client registration of the Register, and compose has no cert-manager.
2. **A cert-manager external issuer that calls the Register.** It keeps the Register checks. It needs a new controller, its CRDs, cluster-scoped RBAC and a change to the approval of cert-manager in each cluster ([cert-manager: external issuers](https://cert-manager.io/docs/contributing/external-issuers/), [cert-manager: CertificateRequest approval](https://cert-manager.io/docs/usage/certificaterequest/)). Compose still needs a different path.
3. **The data holder requests its own certificate from the Register.** It needs no Job. But the data holder then holds the operator token of the Register, it cannot start while the Register is down, and a data holder that is not a lab component cannot use the path.
4. **A command that sends a CSR to the existing Register issue path and writes the result.** A Job runs it at install and a CronJob runs it before expiry. It needs a Kubernetes Role for the Secrets that it writes, and a reload of the certificate in each service.

## Decision

In each deployment of the lab, each CDR certificate of a lab service or a lab test comes from the Register. No other component issues a CDR certificate. We use option 4, because it keeps the checks and the events of the Register, it needs no new component in the cluster, and compose can use the same command with files as its output.

- One command of the lab CLI makes a key and a CSR, sends the CSR to the existing issue path of the Register operator interface, and writes the certificate and its key. In a cluster, it writes a Secret. The key stays where the certificate is used. The command does nothing when the current certificate is valid for more than its renewal window and chains to the current CDR CA bundle.
- In a cluster, a Job runs the command after an install, and a CronJob runs it each day. Thus a certificate is replaced before it expires with no manual step.
- The Register issues the CDR server certificate of its own mTLS listener from its intermediate, at start and before expiry. Its key is only in the memory of the process. This removes the need for a certificate of the Register before the Register runs.
- Each listener of a service uses a renewed certificate file with no restart. The shared TLS code of `pkg/cds` reads the files again when they change. We do not use a tool that restarts a pod when a Secret changes, because the clusters of the teams do not all have one.
- In a cluster, an install Job makes the mock CDR CA and writes the intermediate Secret and the CDR CA bundle Secret, only when they are absent. It keeps no copy of the root key, because no component needs the root key after the set-up. A deployment can give its own CA Secrets with the same names, and the Job then changes nothing.
- cert-manager does not issue CDR certificates. It can issue the public TLS certificates, which do not come from the mock CDR CA.

The issue, the renewal and the reload of CDR certificates are lab-only functions of mock components, as the issue and the revocation of ADR 0022 are. Real components must not depend on them. A real component that reads its certificate files gets the reload of `pkg/cds` with no dependency on the Register.

## Consequences

- The Register checks the names of each CDR certificate of the lab, and the history of the Register records each server certificate issue. A daily CronJob records nothing when no renewal is necessary.
- A cluster needs no CRD, no controller and no cluster-scoped change for CDR certificates. Each Job needs a Role that permits it to read and write Secrets in its namespace. Kubernetes cannot limit the creation of a Secret by name, so the Role permits the creation of any Secret in the namespace.
- The Jobs need the operator token of the Register. The token is a Secret in the namespace.
- A failed renewal leaves the current certificate in use. With a renewal window of 30 days, a failure has many daily runs before the certificate expires. The failure shows only as a failed Job, so a team that does not watch its Jobs finds it late.
- The CDR CA Secrets stay after an uninstall of the chart, because the chart does not own them. A new install reuses them. A new CA needs a delete of the Secrets and a fresh deployment.
- Each restart of the Register makes a new key and a new certificate for its mTLS listener.
- The compose set-up also uses the Register issue path. The data holders of compose then can start only after the Register is healthy and their certificates exist.
- The services and `pkg/cds` have more code: a watch of the certificate files, and the self-issue of the Register.
