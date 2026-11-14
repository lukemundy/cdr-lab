# Mock CDR PKI: design

This document gives the specification and the requirements for the mock CDR PKI. It does not describe the implementation.

## Purpose

In production, the ACCC operates the CDR certificate authority. It issues the certificates that participants use for mTLS. TLS endpoints use certificates from public CAs. The lab cannot use either. The mock PKI gives the lab a CDR CA and a public CA with the same structure and certificate profiles as production. A real component that trusts the mock CAs then needs only a change to its CA bundles and certificates to work with production.

## Sources

The requirements follow the process in [README: Sources for requirements][readme-sources]. Each requirement links to its source. "Judgement" marks a requirement that the sources do not state, and gives the reason. "Observed" marks a fact from a production certificate. The observations are from 2026-10-02.

Primary sources:

- Consumer Data Standards (CDS) 1.36.0: [Certificate Trust Model][cds-trust-model], [Certificate Signing Request Profile][cds-csr-profile], [Certificate Validation][cds-cert-validation], [Participant Endpoints][cds-participant-endpoints].
- ACCC Developer Portal: [Certificate Management][accc-cert-mgmt], [Certificate FAQ][accc-cert-faq]. ACCC [Participant Portal User Guide][accc-portal-guide].
- [ACCC Certificate Policy v1.04][accc-cp] (CP) and [ACCC Certification Practice Statement v2.0][accc-cps] (CPS).

## Scope

In scope:

- A mock CDR CA: a root CA and one issuing intermediate CA.
- A mock public CA for TLS endpoints.
- Issue of CDR certificates from CSRs, and of public TLS server certificates for lab hosts. The `pki` command issues at the set-up of the lab. In a running lab, the Register issues CDR client and server certificates as the intermediate CA ([ADR 0022][adr-0022]).
- Revocation of CDR certificates, and a CRL of the CDR intermediate. In a running lab, the Register revokes CDR certificates, and signs and serves the CRL ([ADR 0022][adr-0022]). The `pki` command does neither.

Out of scope for now:

- OCSP. See Decisions.
- Revocation of CA certificates and of public TLS server certificates.
- The production dual trust chain period. See Decisions.
- The ACCC Conformance Test Suite (CTS) trust chain.
- The Participant Portal workflow: approval, renewal reminders and agreements.

## CA structure

- The CDR CA has a root CA and an issuing intermediate CA. The intermediate issues all participant certificates. ([ACCC: Certificate Management][accc-cert-mgmt])
- The root key is RSA 4096. The intermediate key is RSA 4096. Signatures use SHA-256. ([CP section 6.1.5, Table 5][accc-cp]; observed: the production "CDR Intermediate CA 2025" key is RSA 4096)
- Both CA certificates are CA certificates with key usage `keyCertSign` and `cRLSign`. The intermediate has a path length of 0. (Observed: "CDR Root CA" and "CDR Intermediate CA 2025")
- The root is valid for 20 years. The intermediate is valid for 10 years. (CP maximums are 20 and 15 years, [CP section 6.3.2, Table 6][accc-cp]; observed: "CDR Intermediate CA 2025" is valid for 10 years)
- The mock CA names have the OU `Consumer Data Right`, but not the ACCC name or the ACCC certificate policy OID. Judgement: the mock certificates are not from the ACCC, and they must not look as if they are.
- The public CA is separate from the CDR CA. No certificate from one chains to the other. TLS endpoints use a certificate from a public CA, and they must not use CDR certificates. ([CDS: Participant Endpoints][cds-participant-endpoints]; observed: `api.cdr.gov.au` has a public CA certificate)
- Relying parties trust a CDR CA bundle: the root and the intermediate. The ACCC publishes its trust chains as bundles ([ACCC: Certificate Management][accc-cert-mgmt]). Observed: the production mTLS endpoint `secure.api.cdr.gov.au` sends only its leaf certificate, so a client that trusts only the root cannot verify it.

## CSR profile

The CDR CA issues a certificate only from a CSR that follows the CDS profile ([CDS: Certificate Signing Request Profile][cds-csr-profile]). If a CSR breaks a rule, the CA refuses it and names each rule that it breaks.

- The key is RSA 2048. The signature algorithm is SHA-256 with RSA.
- The CSR signature is valid with the key in the CSR (proof of possession). ([CPS section 3.2.1][accc-cps])
- The subject has CN, O, OU and C. OU is `Consumer Data Right`. C is a two-letter country code.
- Server certificate: the CN is the primary DNS name. The subject alternative names are DNS names only.
- The CN and each DNS name of a server CSR are host names: only letters, digits, hyphens and dots, with no wildcard, and not an IP address. Each label has 1 to 63 characters and does not start or end with a hyphen, and the name has 253 characters or fewer. The CN has 64 characters or fewer. ([ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide], [RFC 1123, section 2.1][rfc1123-names], [RFC 1035, section 2.3.4][rfc1035-size], [RFC 5280, Appendix A][rfc5280-bounds]; see Decisions.)
- The CDS gives the server O as the legal entity name or the brand name. The `pki` command does not check the O. The Register checks it when it issues a server certificate, because only the Register knows the brands ([Register DESIGN.md][register-design]).
- Client certificate: the CSR has no subject alternative names. The CDS gives the client CN as the software product name, and the O as the brand name. The `pki` command does not check the names. The Register checks them when it issues a client certificate, because only the Register knows the software products ([Register DESIGN.md][register-design]).

## Certificate profile

CDR certificates copy the production certificate profile. (Observed: the `secure.api.cdr.gov.au` certificate from "CDR Intermediate CA 2025")

- The subject is the subject of the CSR.
- Valid for 13 months from issue. The [ACCC FAQ][accc-cert-faq] gives 13 months for the old trust chain, and the observed certificate from the current chain also has 13 months.
- A certificate is never valid after its issuer expires. ([CP section 6.3.2][accc-cp])
- Key usage `digitalSignature` (critical). Basic constraints: not a CA.
- Extended key usage: `serverAuth` for a server certificate, `clientAuth` for a client certificate. The CP requires an extended key usage on end-entity certificates and prohibits `anyExtendedKeyUsage` ([CP section 6.1.7][accc-cp]). Judgement: one usage for each certificate type, the same as the observed server certificate.
- A server certificate has a subject alternative name for its CN and for each DNS name in the CSR. ([CP section 10, Table 14][accc-cp]: the SAN is "always present"; observed: the SAN contains the CN)
- A CDR certificate gives one CRL distribution point: the URL of the CRL of its issuer. ([RFC 5280, section 4.2.1.13][rfc5280-crldp]; observed: the production certificate gives the CRL URL of "CDR Intermediate CA 2025") It gives no OCSP URL and no CA issuers URL, because the PKI has no OCSP responder and relying parties load the CA bundle.
- The serial number is random, with more than 64 bits. ([CP section 10, Table 14][accc-cp]: the serial number is unique; the [CA/Browser Forum Baseline Requirements, section 7.1][cabf-br] require 64 random bits)

Public TLS server certificates for lab hosts have the same validity, key usage and extended key usage as a CDR server certificate. Their subject is the CN only, and each of their names follows the host name rule of a server CSR. Judgement: 13 months is less than the 398-day maximum of the [CA/Browser Forum Baseline Requirements, section 6.3.2][cabf-br].

## Revocation

The Register is the only component that revokes a CDR certificate and signs a CRL of the intermediate, and it keeps the only list of revocations ([ADR 0022][adr-0022]). The [Register DESIGN.md][register-design-crl] gives the revocation request, the CRL listener and the times of a new CRL. The rules below apply to each revocation and each CRL.

- The lab operator can revoke a CDR certificate with one of these RFC 5280 reasons: `unspecified`, `keyCompromise`, `affiliationChanged`, `superseded`, `cessationOfOperation` or `privilegeWithdrawn`. The other reasons are for CA certificates, attribute certificates or suspension, and the CA does not support suspension ([CPS section 4.9][accc-cps]). ([RFC 5280, section 5.3.1][rfc5280-reason])
- A certificate that is revoked cannot be revoked again.
- When a revocation completes, the CRL lists the certificate with its reason. Judgement: production can take up to 24 hours ([CPS section 4.9][accc-cps]), but a lab test must see the change at once.
- The revocation date of a CRL entry is the time of the revocation. ([RFC 5280, section 5.1.2.6][rfc5280-revoked]) An entry with the reason `unspecified` has no reason code extension. ([RFC 5280, section 5.3.1][rfc5280-reason]: the extension "SHOULD be absent instead of using the unspecified (0) reasonCode value")
- The CRL is a V2 CRL in DER, signed by the CDR intermediate with SHA-256 and RSA. It has an authority key identifier and a CRL number. Each new CRL has a higher CRL number, also after a restart of the Register, and a CRL number is never used two times. ([CP section 10, Table 17][accc-cp]; [RFC 5280, section 5.2.3][rfc5280-crl-number]; observed: the production CRL of "CDR Intermediate CA 2025")
- The next update of a CRL is 7 days after its issue. (Observed: the production CRL of "CDR Intermediate CA 2025")
- A new CRL is issued at least once every 24 hours, also when no revocation occurred, and also at each revocation. ([CP section 4.9.7, Table 1][accc-cp]: routine CRLs "at least once every 24 hours"; [CPS section 2.3][accc-cps]) Thus the current CRL is valid for 6 days or more while the Register runs.
- A revoked certificate stays on the CRL. Judgement: production removes it after the certificate expires ([CPS section 4.10.1][accc-cps]), but a lab certificate is valid for 13 months, and a lab reset deletes the CA and its CRL.
- The Register serves the CRL over HTTP, with the content type `application/pkix-crl`, at the URL in each CDR certificate. ([RFC 2585, section 4.2][rfc2585-mime])
- A lab reset deletes the CAs and the Register database, and thus each revocation and each CRL. The new intermediate has no revocations, and its CRL numbers start again. The certificates of the earlier lab do not chain to the new CA bundle, so they need no revocation. Judgement: a new intermediate has a new key and a new authority key identifier, so it is a new CRL issuer ([RFC 5280, section 5.2.3][rfc5280-crl-number]: the CRL number is "for a given CRL scope and CRL issuer").

### Set-up

- The `pki` command does not revoke certificates or sign CRLs, and it keeps no list of revocations. It has no `revoke` command and no `crl` command. ([ADR 0022][adr-0022])
- The set-up issues the fixture `conformance/revoked.crt` before the lab starts, as each other fixture, and does not revoke it. Each conformance test that needs a revoked certificate revokes the fixture through the Register operator interface with the reason `keyCompromise`, and accepts a fixture that is already revoked (409). Judgement: the CDS defines no interface to revoke a certificate, so a test uses an interface that the CDR does not define to set up a state ([conformance DESIGN.md: Subject of a test][conformance-design-subject]). The lab start then has no step after the Register starts.

## Keys

- The repository contains no private keys. The lab creates the CA keys when it is set up. (Register DESIGN.md: non-functional requirements)
- Only the owner can read the key file of a root CA. Judgement: these keys sign the CA certificates that all lab participants trust, and no lab component reads them.
- The deployment decides who can read the intermediate key file. In the lab, any user can read it. Judgement: the Register container reads the key as another user, the key signs only lab certificates, and the lab data has no value.
- The Register reads the intermediate key file when it starts, so that it can issue and revoke certificates and sign CRLs in a running lab ([ADR 0022][adr-0022]). It never reads the root key.
- Any user can read the private key of a certificate that the lab issues for its own services and tests. Judgement: lab containers read the key as another user, and the key protects only a throwaway lab identity.

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Copy the dual trust chain period? | No. One root and one intermediate. | Production needs both chains only until the old certificates expire, by 1 November 2026 ([ACCC: Certificate FAQ][accc-cert-faq]). The lasting requirement is that relying parties load a CA bundle, not one certificate. A second intermediate can then be added with no change to the relying parties. |
| CRL and OCSP? | CRL only. No OCSP. | The CDS requires relying parties to check revocation "through CRL or OCSP, as given in each certificate" ([CDS: Certificate Validation][cds-cert-validation]). Production certificates give both, so a relying party that uses the CRL works in production. A CRL is a signed file. OCSP needs a responder service. |
| CRL validity | 7 days | The same as the production CRL. The Register issues a new CRL at least once every 24 hours, so a CRL is replaced 6 days or more before its next update. |
| Times of a new CRL | At least once every 24 hours, and at each revocation | The routine CRL frequency of the [CP section 4.9.7, Table 1][accc-cp]. Judgement: a CRL at each revocation lets a lab test see the change at once. The [Register DESIGN.md][register-design-crl] gives the other times. |
| Revocation of the set-up fixture | By each conformance test that uses it, through the Register, not by the lab start | [ADR 0022][adr-0022]: the Register keeps the only list of revocations. Judgement: in the lab, the data holder loads the CRL at each check ([data holder DESIGN.md: Transport security][dataholder-design-transport]), so it sees the revocation at once, and the lab start needs no step after the Register starts. |
| Host name in the CRL URL | The service name of the Register in the lab network, on the CRL listener of the Register (`http://register`, with the port of the CRL listener in `compose.yaml`, and the path `/crl/cdr-intermediate.crl`) | Judgement: all lab participants, including the conformance tests, run in the lab network. A certificate can give only one useful CRL URL, and production gives one. The Register serves the CRL ([ADR 0022][adr-0022]), so the static CRL server of the lab is not necessary. |
| Mock CA names | `O=cdr-lab, OU=Consumer Data Right`, no ACCC name or policy OID | Judgement: see CA structure. |
| Client certificate CN and O | Checked only by the Register, when it issues | Only the Register knows the software products and the brands. See CSR profile. |
| Server certificate O | Checked only by the Register, when it issues | Only the Register knows the brands and the legal entities. See CSR profile. |
| Names in a server CSR | Host names only, for the CN and for each DNS name | The Participant Portal refuses a CN that is not a host name: "A valid FQDN contains only letters (a-z, A-Z), numbers (0-9), hyphens and dots; no wildcards, IP addresses, URLs, paths, spaces, unsupported characters, or invalid DNS lengths" ([ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide]). Judgement: the portal gives no rule for the subject alternative names, and the CDS calls them "Secondary DNS Name(s)" ([CDS: Certificate Signing Request Profile][cds-csr-profile]), so the same rule applies to each name. The rule does not require a dot, so a lab host name such as `wattle-bank` is valid. |
| Unique subject names | Not checked | The [CPS section 3.1][accc-cps] requires a unique subject DN. Judgement: a renewal in production keeps the CN, so a strict check would stop renewal in the lab. |

[readme-sources]: ../README.md#sources-for-requirements
[adr-0022]: ../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[register-design]: ../services/register/DESIGN.md#operator-interface
[register-design-crl]: ../services/register/DESIGN.md#certificate-revocation-list
[readme-lab]: ../README.md#lab-deployment
[conformance-design-subject]: ../conformance/DESIGN.md#subject-of-a-test
[dataholder-design-transport]: ../services/dataholder/DESIGN.md#transport-security
[cds-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cds-cert-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[accc-cert-mgmt]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management
[accc-cert-faq]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/374538246/Frequently+Asked+Questions+FAQ
[accc-cp]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[rfc5280-crldp]: https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.13
[rfc5280-bounds]: https://www.rfc-editor.org/rfc/rfc5280#appendix-A
[rfc1123-names]: https://www.rfc-editor.org/rfc/rfc1123#section-2.1
[rfc1035-size]: https://www.rfc-editor.org/rfc/rfc1035#section-2.3.4
[accc-portal-guide]: https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf
[rfc5280-reason]: https://www.rfc-editor.org/rfc/rfc5280#section-5.3.1
[rfc5280-revoked]: https://www.rfc-editor.org/rfc/rfc5280#section-5.1.2.6
[rfc5280-crl-number]: https://www.rfc-editor.org/rfc/rfc5280#section-5.2.3
[rfc2585-mime]: https://www.rfc-editor.org/rfc/rfc2585#section-4.2
[cabf-br]: https://cabforum.org/working-groups/server/baseline-requirements/requirements/
