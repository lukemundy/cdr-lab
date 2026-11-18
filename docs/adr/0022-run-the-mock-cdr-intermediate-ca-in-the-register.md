# 0022. Run the mock CDR intermediate CA in the Register

- Date: 2026-10-27
- Amended by [ADR-0025](0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md): the listener of the CRL

## Context

In production, the CDR CA issues a participant certificate from a CSR that the participant gives. The CA does not get the private key ([ACCC CPS v2.0, section 3.2.1][accc-cps]). The Register accepts a client certificate only when the certificate was issued to the software product ([ACCC Developer Portal: Token Endpoint][accc-token]).

The lab has a mock CDR CA: a root and an issuing intermediate ([PKI DESIGN.md](../../pki/DESIGN.md)). Its state is files on the host. These files are the CA certificates, the CA keys, and a list of revocations with the CRL number. The `pki` command makes the CAs. It also issues, revokes and makes CRLs. A static file server serves the CRL. Thus each issue and each revocation needs a person or a script on the host, with a copy of the repository and read access to the intermediate key.

Each CRL is valid for 7 days. Only `pki crl` and `pki revoke` make a new CRL, and the lab set-up runs `pki crl`. The Register and the data holder refuse a CRL after its next update. They then refuse each client certificate, because the CDS requires a revocation check ([CDS: Certificate Validation][cds-cert-validation]) and the Register refuses a certificate when it cannot do the check ([the revocation check fails with no current CRL](../../services/register/design/certificate-revocation.md#reg-0729)). Thus, 7 days after the last host action, all mTLS clients fail.

The lab operator must be able to add and remove components in a running lab with no host access, no restart and no reset. The operator works through the Register operator interface, with scripts and agents, and through the console. Thus a process that runs all the time must issue certificates, keep the revocations and sign a new CRL before the current one expires. That process must have the intermediate key.

The set-up issues the certificates of the lab services and tests before those services start. A service cannot issue its own server certificate. Thus a part of the CA work stays before the start of the lab.

The options were:

1. **The Register runs the intermediate CA.** It issues from a CSR, keeps the revocations in its database, and signs and serves the CRL.
   - Advantages: one component to deploy. No new API and no new authentication between services. The Register checks a CSR against the participant data and registers the client certificate in one action.
   - Disadvantages: the production Register is not a CA. A security problem in the Register also exposes the intermediate key. The Register gets a second responsibility, and the CRL gets a new URL.
2. **A new PKI service runs the intermediate CA**, with an internal API that the Register calls.
   - Advantages: the CA is separate from the Register, as in production. The key is not in the Register. The CRL can keep its URL.
   - Disadvantages: a new service, an internal API between two services, and authentication for that API. A participant that deploys the Register must also deploy the PKI service. The issue and the registration of a client certificate are in two services, and a failure between the two steps needs a rule.
3. **The Register signs with the CA files of the host**, on a mount that it can write. The `pki` command also writes the same files.
   - Advantages: the least work for the issue.
   - Disadvantages: two writers of one CA state. A revocation from the host and one from the Register can both write the list of revocations. The CRL renewal still needs a design.

## Decision

The Register runs the mock CDR intermediate CA in a running lab. It issues CDR certificates from CSRs. It keeps the revocations of the intermediate. It signs a new CRL before the current CRL expires, and it serves the CRL over HTTP. The reasons:

- A process that runs all the time must hold the key, so that the CRL stays current with no host action. The Register runs all the time already.
- Only the Register knows the software products and the participants. It must check a client CSR against them and register the client certificate. In production, the issue and the registration are one ACCC process.
- A participant that deploys the Register for its own tests gets the CA with it.

The Register reads the intermediate certificate and the intermediate key from files when it starts. The key is not in the Register database, and no interface of the Register gives it. The Register never gets the root key.

The Register keeps the only list of revocations of the intermediate, in its database. After this decision, no other component revokes a CDR certificate or signs a CRL of the intermediate.

The `pki` command keeps the set-up: it makes the CAs, and it issues the certificates of the lab services and tests before the lab starts. The Register and the `pki` command share the code that builds and checks certificates, CSRs and CRLs, in a `pkg/` library ([ADR 0004](0004-module-layout.md)).

This decision applies to the mock CDR intermediate CA only. It does not apply to the mock CDR root, or to the mock public CA: the Register is not a public CA, and a component gets its public certificate from the CA of its deployment. The issue and the revocation are lab-only features of the mock Register. Real components must not depend on them.

## Consequences

- An operator issues and revokes CDR certificates through the Register operator interface and console, with no host access.
- The CRL stays current while the Register runs. When the Register stops, the last CRL stays valid until its next update, up to 7 days.
- A security problem in the Register can expose the intermediate key. Then a person can issue certificates that each lab participant trusts. This risk is acceptable only because the lab data has no value. The key in a file, with no key management service or hardware security module, is a production gap.
- The Register container reads the key file as a user that is not the owner of the file on the host. Docker and Podman map container users to host users differently, so in the lab any user can read the intermediate key file.
- The CRL URL in each new CDR certificate changes to the Register. Certificates of an earlier lab give the old URL, so a lab reset is necessary after the change.
- The revocation of a fixture certificate needs a running Register, so a test that needs a revoked certificate revokes it through the Register.
- The Register has more code and one more listener. A change to the CA rules is now a change to the Register, and the PKI design and the Register design both give rules for the same CA.
- A later decision to move the CA to a separate service needs a new API between the services, but the shared `pkg/` code stays.

[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[cds-cert-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation
