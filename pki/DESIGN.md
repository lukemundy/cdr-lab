# Mock CDR PKI

The mock CDR PKI gives the lab a CDR CA and a separate public CA. They have the structure and the certificate profiles of production. In production, the ACCC operates the CDR CA, and TLS endpoints use certificates from public CAs. A real component that trusts the mock CAs then needs only a change of its CA bundles and certificates to work with production. The `pki` command makes the CAs and issues the certificates of the lab at set-up. In a running lab, the Register is the CDR intermediate CA ([ADR 0022][adr-0022]).

## Context

```mermaid
flowchart LR
  labuser["Lab user<br>[Person]"]
  pki["Mock CDR PKI<br>[System]"]
  register["Mock CDR Register<br>[External system]"]
  services["Lab services and tests<br>[External system]"]

  labuser -- "Makes the CAs and issues certificates<br>pki command" --> pki
  pki -- "Gives the intermediate certificate and key<br>Files" --> register
  pki -- "Gives the CA bundles, certificates and keys<br>Files" --> services
  services -- "Get the CRL of the intermediate<br>HTTP" --> register

  classDef person fill:#08427b,color:#fff,stroke:#052e56
  classDef system fill:#1168bd,color:#fff,stroke:#0b4884
  classDef external fill:#999,color:#fff,stroke:#6b6b6b
  class labuser person
  class pki system
  class register,services external
```

## Library specifications

| Library | Specification |
|---|---|
| `pkg/ca` | None. |

## Goals

- A mock CDR CA: a root CA and one issuing intermediate CA.
- A mock public CA for TLS endpoints.
- The issue of CDR certificates from CSRs, and of public TLS server certificates for lab hosts. The `pki` command issues at the set-up of the lab. In a running lab, the Register issues CDR client and server certificates as the intermediate CA ([ADR 0022][adr-0022]).
- The revocation of CDR certificates, and a CRL of the CDR intermediate. In a running lab, the Register revokes CDR certificates, and signs and serves the CRL ([ADR 0022][adr-0022]). The `pki` command does neither.

## Non-goals

- OCSP ([a CRL and no OCSP](#pki-0084)).
- The revocation of CA certificates and of public TLS server certificates.
- A second intermediate or a second trust chain. The dual trust chain period of production ends when the certificates of the earlier chain expire, by 1 November 2026 ([ACCC: Certificate FAQ][accc-cert-faq]).
- The trust chain of the ACCC Conformance Test Suite (CTS).
- The workflow of the Participant Portal: approval, renewal reminders and agreements.

## CA structure

- <a id="pki-0001"></a>**PKI-0001** The CDR CA has one root CA and one issuing intermediate CA. ([ACCC: Certificate Management][accc-cert-mgmt])
- <a id="pki-0002"></a>**PKI-0002** The intermediate issues each CDR certificate. ([ACCC: Certificate Management][accc-cert-mgmt])
- <a id="pki-0003"></a>**PKI-0003** The root key and the intermediate key are RSA 4096 keys. ([CP section 6.1.5, Table 5][accc-cp]; observed on 2026-10-02: the key of "CDR Intermediate CA 2025" is RSA 4096)
- <a id="pki-0004"></a>**PKI-0004** Each CA signs with SHA-256 with RSA. ([CP section 6.1.5, Table 5][accc-cp])
- <a id="pki-0005"></a>**PKI-0005** The root certificate and the intermediate certificate are CA certificates with the key usages `keyCertSign` and `cRLSign`. ([Judgement: the profiles of the production certificates and CRL](#pki-0079))
- <a id="pki-0006"></a>**PKI-0006** The intermediate certificate has a path length of 0. ([Judgement: the profiles of the production certificates and CRL](#pki-0079))
- <a id="pki-0007"></a>**PKI-0007** The root certificate is valid for 20 years. ([CP section 6.3.2, Table 6][accc-cp])
- <a id="pki-0008"></a>**PKI-0008** The intermediate certificate is valid for 10 years. ([CP section 6.3.2, Table 6][accc-cp]; [Judgement: the profiles of the production certificates and CRL](#pki-0079))
- <a id="pki-0009"></a>**PKI-0009** The subject of the CDR root and of the CDR intermediate has the O `cdr-lab` and the OU `Consumer Data Right`. ([Judgement: mock CA names that are not from the ACCC](#pki-0078))
- <a id="pki-0010"></a>**PKI-0010** The CDR CA certificates have no ACCC name and no ACCC certificate policy OID. ([Judgement: mock CA names that are not from the ACCC](#pki-0078))
- <a id="pki-0011"></a>**PKI-0011** The public CA is separate from the CDR CA, and no certificate of one CA chains to the other. ([CDS: Participant Endpoints][cds-participant-endpoints]; observed on 2026-10-02: `api.cdr.gov.au` has a certificate from a public CA)
- <a id="pki-0012"></a>**PKI-0012** A TLS endpoint uses a certificate from the public CA, not a CDR certificate. ([CDS: Participant Endpoints][cds-participant-endpoints])
- <a id="pki-0013"></a>**PKI-0013** Relying parties trust the CDR CA bundle, which has the intermediate and the root. ([ACCC: Certificate Management][accc-cert-mgmt]; observed on 2026-10-02: `secure.api.cdr.gov.au` sends only its leaf certificate, so a client that trusts only the root cannot verify it)

## CSR profile

The CSR profile is the [CDS: Certificate Signing Request Profile][cds-csr-profile]. These rules give what the CDR CA checks.

- <a id="pki-0014"></a>**PKI-0014** The CDR CA issues a certificate only from a CSR that obeys each rule of this group. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="pki-0015"></a>**PKI-0015** When a CSR breaks a rule, the CA refuses the CSR and names each rule that it breaks. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="pki-0016"></a>**PKI-0016** The key of the CSR is an RSA 2048 key. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="pki-0017"></a>**PKI-0017** The signature algorithm of the CSR is SHA-256 with RSA. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="pki-0018"></a>**PKI-0018** The CSR signature is valid with the key in the CSR (proof of possession). ([CPS section 3.2.1][accc-cps])
- <a id="pki-0019"></a>**PKI-0019** The subject of the CSR has these attributes. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
  - CN
  - O
  - OU, with the value `Consumer Data Right`
  - C, with a two-letter country code
- <a id="pki-0020"></a>**PKI-0020** In a server CSR, the CN is the primary DNS name. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="pki-0021"></a>**PKI-0021** The subject alternative names of a server CSR are DNS names only. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="pki-0022"></a>**PKI-0022** The CN of a server CSR is a host name: only letters, digits, hyphens and dots, with no wildcard, and not an IP address. ([ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide]; [RFC 1123, section 2.1][rfc1123-names])
- <a id="pki-0023"></a>**PKI-0023** Each DNS name of a server CSR obeys the host name rules of the CN. ([Judgement: the host name rules for each name in a server CSR](#pki-0080))
- <a id="pki-0024"></a>**PKI-0024** Each label of a host name has 1 to 63 characters, and does not start or end with a hyphen. ([RFC 1123, section 2.1][rfc1123-names]; [RFC 1035, section 2.3.4][rfc1035-size])
- <a id="pki-0025"></a>**PKI-0025** A host name has 253 characters or fewer. ([RFC 1035, section 2.3.4][rfc1035-size])
- <a id="pki-0026"></a>**PKI-0026** The CN of a server CSR has 64 characters or fewer. ([RFC 5280, Appendix A][rfc5280-bounds])
- <a id="pki-0027"></a>**PKI-0027** A host name with no dot is valid, for example `wattle-bank`. ([Judgement: the host name rules for each name in a server CSR](#pki-0080))
- <a id="pki-0028"></a>**PKI-0028** The `pki` command does not check the O of a server CSR. ([Judgement: only the Register checks the participant names of a CSR](#pki-0081))
- <a id="pki-0029"></a>**PKI-0029** A client CSR has no subject alternative names. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="pki-0030"></a>**PKI-0030** The `pki` command does not check the CN and the O of a client CSR. ([Judgement: only the Register checks the participant names of a CSR](#pki-0081))

## Certificate profile

A CDR certificate copies the profile of the production server certificate of `secure.api.cdr.gov.au`, from "CDR Intermediate CA 2025".

- <a id="pki-0031"></a>**PKI-0031** The subject of a CDR certificate is the subject of its CSR. ([Judgement: the profiles of the production certificates and CRL](#pki-0079))
- <a id="pki-0032"></a>**PKI-0032** A CDR certificate is valid for 13 months from its issue. ([ACCC: Certificate FAQ][accc-cert-faq]; [Judgement: the profiles of the production certificates and CRL](#pki-0079))
- <a id="pki-0033"></a>**PKI-0033** A certificate is never valid after its issuer expires. ([CP section 6.3.2][accc-cp])
- <a id="pki-0034"></a>**PKI-0034** A CDR certificate has the key usage `digitalSignature` in a critical extension. ([Judgement: the profiles of the production certificates and CRL](#pki-0079))
- <a id="pki-0035"></a>**PKI-0035** The basic constraints of a CDR certificate say that it is not a CA. ([Judgement: the profiles of the production certificates and CRL](#pki-0079))
- <a id="pki-0036"></a>**PKI-0036** A server certificate has the extended key usage `serverAuth` only. ([CP section 6.1.7][accc-cp]; [Judgement: one extended key usage for each certificate type](#pki-0082))
- <a id="pki-0037"></a>**PKI-0037** A client certificate has the extended key usage `clientAuth` only. ([CP section 6.1.7][accc-cp]; [Judgement: one extended key usage for each certificate type](#pki-0082))
- <a id="pki-0038"></a>**PKI-0038** A CDR certificate never has the extended key usage `anyExtendedKeyUsage`. ([CP section 6.1.7][accc-cp])
- <a id="pki-0039"></a>**PKI-0039** A server certificate has a subject alternative name for its CN and for each DNS name in the CSR. ([CP section 10, Table 14][accc-cp]; observed on 2026-10-02: the SAN contains the CN)
- <a id="pki-0040"></a>**PKI-0040** A CDR certificate gives one CRL distribution point, which is the URL of the CRL of its issuer. ([RFC 5280, section 4.2.1.13][rfc5280-crldp]; observed on 2026-10-02: the production certificate gives the CRL URL of "CDR Intermediate CA 2025")
- <a id="pki-0041"></a>**PKI-0041** In the lab, the CRL URL is on the CRL listener of the Register in the lab network: `http://register`, the port of the CRL listener in `compose.yaml`, and the path `/crl/cdr-intermediate.crl`. ([Judgement: the Register service name in the CRL URL](#pki-0091))
- <a id="pki-0042"></a>**PKI-0042** A CDR certificate gives no OCSP URL and no CA issuers URL. ([Judgement: a CRL and no OCSP](#pki-0084))
- <a id="pki-0043"></a>**PKI-0043** The serial number of a certificate is random, with more than 64 bits. ([CP section 10, Table 14][accc-cp]; [CA/Browser Forum Baseline Requirements, section 7.1][cabf-br])
- <a id="pki-0044"></a>**PKI-0044** The CA does not check that the subject DN of a certificate is unique. ([CPS section 3.1][accc-cps]; [Judgement: no check of unique subject names](#pki-0092))
- <a id="pki-0045"></a>**PKI-0045** A public TLS server certificate for a lab host has the validity, the key usage and the extended key usage of a CDR server certificate. ([CA/Browser Forum Baseline Requirements, section 6.3.2][cabf-br]; [Judgement: the CDR server profile for a public TLS server certificate](#pki-0083))
- <a id="pki-0046"></a>**PKI-0046** The subject of a public TLS server certificate has the CN only. ([Judgement: the CDR server profile for a public TLS server certificate](#pki-0083))
- <a id="pki-0047"></a>**PKI-0047** Each name of a public TLS server certificate obeys the host name rules of a server CSR. ([Judgement: the CDR server profile for a public TLS server certificate](#pki-0083))

## Revocation

The [Register DESIGN.md][register-design-crl] gives the revocation request, the CRL listener and the other times of a new CRL. The rules of this group apply to each revocation and each CRL.

- <a id="pki-0048"></a>**PKI-0048** The Register is the only component that revokes a CDR certificate. ([ADR 0022][adr-0022])
- <a id="pki-0049"></a>**PKI-0049** The Register is the only component that signs a CRL of the intermediate. ([ADR 0022][adr-0022])
- <a id="pki-0050"></a>**PKI-0050** The Register keeps the only list of revocations. ([ADR 0022][adr-0022])
- <a id="pki-0051"></a>**PKI-0051** The lab operator can revoke a CDR certificate with one of these reasons. ([RFC 5280, section 5.3.1][rfc5280-reason]; [CPS section 4.9][accc-cps]: the CA does not support suspension)
  - `unspecified`
  - `keyCompromise`
  - `affiliationChanged`
  - `superseded`
  - `cessationOfOperation`
  - `privilegeWithdrawn`
- <a id="pki-0052"></a>**PKI-0052** When the lab operator revokes a certificate that is revoked, the CA refuses the revocation. ([Judgement: one revocation for each certificate](#pki-0085))
- <a id="pki-0053"></a>**PKI-0053** When a revocation completes, the CRL lists the certificate with its reason. ([Judgement: a new CRL at each revocation](#pki-0086))
- <a id="pki-0054"></a>**PKI-0054** The revocation date of a CRL entry is the time of the revocation. ([RFC 5280, section 5.1.2.6][rfc5280-revoked])
- <a id="pki-0055"></a>**PKI-0055** A CRL entry with the reason `unspecified` has no reason code extension. ([RFC 5280, section 5.3.1][rfc5280-reason])
- <a id="pki-0056"></a>**PKI-0056** The CRL is a V2 CRL in DER. ([CP section 10, Table 17][accc-cp]; observed on 2026-10-02: the production CRL of "CDR Intermediate CA 2025")
- <a id="pki-0057"></a>**PKI-0057** The CDR intermediate signs the CRL with SHA-256 with RSA. ([CP section 10, Table 17][accc-cp])
- <a id="pki-0058"></a>**PKI-0058** The CRL has an authority key identifier and a CRL number. ([CP section 10, Table 17][accc-cp]; [RFC 5280, section 5.2.3][rfc5280-crl-number])
- <a id="pki-0059"></a>**PKI-0059** Each new CRL has a higher CRL number than each CRL before it, also after a restart of the Register. ([RFC 5280, section 5.2.3][rfc5280-crl-number])
- <a id="pki-0060"></a>**PKI-0060** The next update of a CRL is 7 days after its issue. ([Judgement: a CRL validity of 7 days](#pki-0087))
- <a id="pki-0061"></a>**PKI-0061** The Register issues a new CRL at least once every 24 hours, also when no revocation occurs. ([CP section 4.9.7, Table 1][accc-cp]; [CPS section 2.3][accc-cps])
- <a id="pki-0062"></a>**PKI-0062** The Register issues a new CRL at each revocation. ([Judgement: a new CRL at each revocation](#pki-0086))
- <a id="pki-0063"></a>**PKI-0063** A revoked certificate stays on the CRL, also after the certificate expires. ([Judgement: revoked certificates stay on the CRL](#pki-0088))
- <a id="pki-0064"></a>**PKI-0064** The Register serves the CRL over HTTP, with the content type `application/pkix-crl`, at the URL in each CDR certificate. ([RFC 2585, section 4.2][rfc2585-mime])
- <a id="pki-0065"></a>**PKI-0065** A lab reset deletes the CAs and the Register database, and thus each revocation and each CRL. ([Judgement: a lab reset makes a new CRL issuer](#pki-0089))
- <a id="pki-0066"></a>**PKI-0066** After a lab reset, the new intermediate has no revocations, and its CRL numbers start again. ([Judgement: a lab reset makes a new CRL issuer](#pki-0089))
- <a id="pki-0067"></a>**PKI-0067** The `pki` command does not revoke certificates, sign CRLs or keep a list of revocations, and it has no `revoke` command and no `crl` command. ([ADR 0022][adr-0022])
- <a id="pki-0068"></a>**PKI-0068** The set-up issues the fixture `conformance/revoked.crt` before the lab starts, as each other fixture, and does not revoke it. ([Judgement: each conformance test revokes the fixture](#pki-0090))
- <a id="pki-0069"></a>**PKI-0069** When a conformance test needs a revoked certificate, the test revokes the fixture through the Register operator interface with the reason `keyCompromise`. ([Judgement: each conformance test revokes the fixture](#pki-0090))
- <a id="pki-0070"></a>**PKI-0070** When the fixture is already revoked (409), the conformance test accepts the response. ([Judgement: each conformance test revokes the fixture](#pki-0090))

## Keys

- <a id="pki-0071"></a>**PKI-0071** Secrets: the repository contains no private keys. ([Judgement: no private key in the repository](#pki-0093))
- <a id="pki-0072"></a>**PKI-0072** The lab creates the CA keys when it is set up. ([Judgement: no private key in the repository](#pki-0093))
- <a id="pki-0073"></a>**PKI-0073** Only the owner can read the key file of a root CA. ([Judgement: only the owner reads a root key](#pki-0094))
- <a id="pki-0074"></a>**PKI-0074** The deployment sets who can read the intermediate key file, and in the lab any user can read it. ([Judgement: lab keys that any user can read](#pki-0095))
- <a id="pki-0075"></a>**PKI-0075** When the Register starts, it reads the intermediate key file, so that it can issue certificates, revoke certificates and sign CRLs. ([ADR 0022][adr-0022])
- <a id="pki-0076"></a>**PKI-0076** The Register never reads a root key. ([ADR 0022][adr-0022])
- <a id="pki-0077"></a>**PKI-0077** Any user can read the private key of a certificate that the lab issues for its own services and tests. ([Judgement: lab keys that any user can read](#pki-0095))

## Judgements

### <a id="pki-0078"></a>PKI-0078 Mock CA names that are not from the ACCC

- Kind: choice
- Basis: Step 3.1. The ACCC issues the production CDR certificates under its certificate policy ([CP][accc-cp]).
- Reason: The mock certificates are not from the ACCC. They must not look as if they are.

### <a id="pki-0079"></a>PKI-0079 The profiles of the production certificates and CRL

- Kind: choice
- Basis: Step 3.2. On 2026-10-02, we observed "CDR Root CA", "CDR Intermediate CA 2025", the server certificate of `secure.api.cdr.gov.au` and the CRL of "CDR Intermediate CA 2025". The [CP section 6.3.2, Table 6][accc-cp] gives only maximums, for example 15 years for an intermediate.
- Reason: A mock component copies the behaviour of its production equivalent ([README: Component types][readme-component-types]). A relying party that works with the mock certificates then works with the production certificates.

### <a id="pki-0080"></a>PKI-0080 The host name rules for each name in a server CSR

- Kind: choice
- Basis: Step 3.2. The [ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide] gives the host name rules for the CN only. The [CDS: Certificate Signing Request Profile][cds-csr-profile] calls the subject alternative names secondary DNS names, and [RFC 1123, section 2.1][rfc1123-names] gives the syntax of a host name.
- Reason: Each name is a DNS name, so the same rules apply to each name. The rules require no dot, so a lab host name is valid.

### <a id="pki-0081"></a>PKI-0081 Only the Register checks the participant names of a CSR

- Kind: choice
- Basis: Step 3.3. The [CDS: Certificate Signing Request Profile][cds-csr-profile] gives the CN and the O of a client certificate, and the O of a server certificate, from the names of the participant.
- Reason: Only the Register knows the software products, the brands and the legal entities. The Register checks the names when it issues a certificate ([Register DESIGN.md: Operator interface][register-design]).

### <a id="pki-0082"></a>PKI-0082 One extended key usage for each certificate type

- Kind: choice
- Basis: Step 3.2. The [CP section 6.1.7][accc-cp] requires an extended key usage on an end-entity certificate. On 2026-10-02, the production server certificate of `secure.api.cdr.gov.au` had one extended key usage.
- Reason: One usage for each certificate type is the narrowest profile that the CP permits. Production uses it.

### <a id="pki-0083"></a>PKI-0083 The CDR server profile for a public TLS server certificate

- Kind: choice
- Basis: Step 3.3. The [CA/Browser Forum Baseline Requirements, section 6.3.2][cabf-br] give a maximum validity of 398 days.
- Reason: One profile for each lab server certificate is the least complex. A validity of 13 months is less than 398 days.

### <a id="pki-0084"></a>PKI-0084 A CRL and no OCSP

- Kind: choice
- Basis: Step 3.3. The [CDS: Certificate Validation][cds-cert-validation] requires a relying party to check revocation through the CRL or OCSP that each certificate gives. Production certificates give both.
- Reason: A relying party that uses the CRL then works in production. A CRL is a signed file, but OCSP needs a responder service. Relying parties load the CA bundle, so a certificate needs no CA issuers URL.

### <a id="pki-0085"></a>PKI-0085 One revocation for each certificate

- Kind: choice
- Basis: Step 3.3. The [CPS section 4.9.12][accc-cps] lets the ACCC change a reason to `keyCompromise` when it finds a key compromise.
- Reason: The lab has no process that finds a key compromise. One revocation for each certificate is the least complex rule.

### <a id="pki-0086"></a>PKI-0086 A new CRL at each revocation

- Kind: choice
- Basis: Step 3.3. Production can take up to 24 hours to publish a revocation ([CPS section 4.9][accc-cps]).
- Reason: A lab test must see a revocation at once. The [Register DESIGN.md][register-design-crl] gives the other times of a new CRL.

### <a id="pki-0087"></a>PKI-0087 A CRL validity of 7 days

- Kind: choice
- Basis: Step 3.2. On 2026-10-02, the next update of the production CRL of "CDR Intermediate CA 2025" was 7 days after its issue.
- Reason: The CRL copies production. With a new CRL at least once every 24 hours, the current CRL is valid for 6 days or more while the Register runs.

### <a id="pki-0088"></a>PKI-0088 Revoked certificates stay on the CRL

- Kind: gap
- Basis: Step 2. Production removes an entry from the CRL after the certificate expires ([CPS section 4.10.1][accc-cps]).
- Reason: A lab certificate is valid for 13 months. A lab reset deletes the CA and its CRL.

### <a id="pki-0089"></a>PKI-0089 A lab reset makes a new CRL issuer

- Kind: choice
- Basis: Step 3.2. A CRL number is unique for one CRL scope and one CRL issuer ([RFC 5280, section 5.2.3][rfc5280-crl-number]).
- Reason: A new intermediate has a new key and a new authority key identifier, so it is a new CRL issuer. Its CRL numbers can start again. The certificates of the earlier lab do not chain to the new CA bundle, so they need no revocation.

### <a id="pki-0090"></a>PKI-0090 Each conformance test revokes the fixture

- Kind: choice
- Basis: Step 3.3. The CDS defines no interface to revoke a certificate, and the Register keeps the only list of revocations ([ADR 0022][adr-0022]). A test can use an interface that the CDR does not define to set up a state ([conformance DESIGN.md: Subject of a test][conformance-design-subject]).
- Reason: In the lab, the data holder loads the CRL at each check ([data holder DESIGN.md: Transport security][dataholder-design-transport]), so it sees the revocation at once. The lab start then has no step after the Register starts.

### <a id="pki-0091"></a>PKI-0091 The Register service name in the CRL URL

- Kind: choice
- Basis: Step 3.3. A certificate can give only one useful CRL URL, and production gives one ([RFC 5280, section 4.2.1.13][rfc5280-crldp]).
- Reason: All lab participants, also the conformance tests, run in the lab network. The Register serves the CRL ([ADR 0022][adr-0022]).

### <a id="pki-0092"></a>PKI-0092 No check of unique subject names

- Kind: gap
- Basis: Step 2. The [CPS section 3.1][accc-cps] requires a unique subject DN.
- Reason: A renewal in production keeps the CN. A strict check stops a renewal in the lab.

### <a id="pki-0093"></a>PKI-0093 No private key in the repository

- Kind: choice
- Basis: Step 3.2. [CWE-321][cwe-321] lists a hard-coded cryptographic key as a weakness.
- Reason: The repository is public, so a key in it protects nothing. The lab creates each key at set-up.

### <a id="pki-0094"></a>PKI-0094 Only the owner reads a root key

- Kind: choice
- Basis: Step 3.2. [CWE-732][cwe-732] lists a permission that is too wide on a critical resource as a weakness.
- Reason: The root keys sign the CA certificates that all lab participants trust. No lab component reads them.

### <a id="pki-0095"></a>PKI-0095 Lab keys that any user can read

- Kind: choice
- Basis: Step 3.3. Lab containers read the keys as a different user ([production gaps][production-gaps]).
- Reason: The intermediate key signs only lab certificates. The key of a lab certificate protects only a throwaway lab identity. The lab data has no value.

## Withdrawn

None.

[readme-component-types]: ../README.md#component-types
[adr-0022]: ../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[production-gaps]: ../docs/production-gaps.md
[register-design]: ../services/register/DESIGN.md#operator-interface
[register-design-crl]: ../services/register/DESIGN.md#certificate-revocation-list
[conformance-design-subject]: ../conformance/DESIGN.md#subject-of-a-test
[dataholder-design-transport]: ../services/dataholder/DESIGN.md#transport-security
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cds-cert-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[accc-cert-mgmt]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management
[accc-cert-faq]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/374538246/Frequently+Asked+Questions+FAQ
[accc-cp]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[accc-portal-guide]: https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf
[rfc5280-crldp]: https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.13
[rfc5280-bounds]: https://www.rfc-editor.org/rfc/rfc5280#appendix-A
[rfc5280-reason]: https://www.rfc-editor.org/rfc/rfc5280#section-5.3.1
[rfc5280-revoked]: https://www.rfc-editor.org/rfc/rfc5280#section-5.1.2.6
[rfc5280-crl-number]: https://www.rfc-editor.org/rfc/rfc5280#section-5.2.3
[rfc1123-names]: https://www.rfc-editor.org/rfc/rfc1123#section-2.1
[rfc1035-size]: https://www.rfc-editor.org/rfc/rfc1035#section-2.3.4
[rfc2585-mime]: https://www.rfc-editor.org/rfc/rfc2585#section-4.2
[cabf-br]: https://cabforum.org/working-groups/server/baseline-requirements/requirements/
[cwe-321]: https://cwe.mitre.org/data/definitions/321.html
[cwe-732]: https://cwe.mitre.org/data/definitions/732.html
