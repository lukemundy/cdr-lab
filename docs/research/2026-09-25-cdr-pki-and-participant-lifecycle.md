# CDR Register and PKI: operational rules for a mock Register

- Research date: 2026-09-25
- Consumer Data Standards (CDS) version: 1.36.0 (released 04/12/2025). This is the current version at the research date.
- CDR Rules version: Competition and Consumer (Consumer Data Right) Rules 2020, Compilation No. 10, compilation date 4 March 2025 (includes the 2025 Measures No. 1 amendments). This is the latest compilation on the Federal Register of Legislation at the research date.
- Purpose: source-cited facts for requirements (functional and non-functional) for a mock CDR Register and a mock CDR PKI. These notes do not design an implementation.

Check the sources again before you depend on a fact; the rules and standards change.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site and archive) | https://consumerdatastandardsaustralia.github.io/standards/ and https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/ |
| S2 | CDS source repository (Register OpenAPI spec, CA certificates, CPS v1.0) | https://github.com/ConsumerDataStandardsAustralia/standards (files: `slate/source/includes/swagger/cds_register.json`, `slate/source/includes/register/certificates/`) |
| S3 | CDR Register Design Reference v1.5.0 (ARCHIVED; the CDS now holds the design) | https://consumerdatastandardsaustralia.github.io/register/ |
| S4 | ACCC Certificate Policy (CP) v1.04, 27 August 2025 | https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf |
| S5 | ACCC Certification Practice Statement (CPS) v2.0, 27 August 2025 | https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf |
| S6 | ACCC CPS v1.0 (14 May 2020, superseded; governs certificates from the old intermediate) | `slate/source/includes/register/certificates/documentation/ACCC_CPS_v1.0.docx` in S2; superseded versions: https://www.cdr.gov.au/resources/agreements/digital-certificate-agreements-superseded-versions |
| S7 | ACCC Developer Portal: Certificate Management | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management |
| S8 | ACCC Developer Portal: Certificate FAQ | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/374538246/Frequently+Asked+Questions+FAQ |
| S9 | Standards maintenance issue 695 (trust chain migration) | https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/695 |
| S10 | ACCC Developer Portal: Register API pages | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29982721/Register+API (child pages are listed in section 9) |
| S11 | cdr.gov.au on-boarding pages | https://www.cdr.gov.au/for-providers/on-boarding-for-data-recipients and https://www.cdr.gov.au/for-providers/on-boarding-for-data-holders |
| S12 | CDR participant on-boarding guide v3, January 2026 | https://www.cdr.gov.au/sites/default/files/2026-01/CDR-participant-on-boarding-guide-v3-published-30-January-2026.pdf |
| S13 | Participant Portal User Guide v23, 28 August 2026 | https://www.cdr.gov.au/sites/default/files/2026-09/Participant%20Portal%20User%20Guide_-v23-20260828_141801.pdf |
| S14 | CDR Rules 2020, Compilation No. 10 | https://www.legislation.gov.au/F2020L00094/latest/text |
| S15 | CDR Support Portal articles (IDs given for each fact) | https://cdr-support.zendesk.com/hc/en-us/articles/<id> |
| S16 | ACCC mock-register (v2.2.8, 15 July 2026, aligned to CDS 1.36.0) | https://github.com/ConsumerDataRight/mock-register |
| S17 | cdr.gov.au rollout pages | https://www.cdr.gov.au/rollout and https://www.cdr.gov.au/rollout/cdr-non-bank-lenders-sector |
| S18 | Treasury: CDR expansion to telecommunications | https://treasury.gov.au/consultation/c2022-315575 |
| S19 | Standards maintenance issue 537 (periodic polling model) | https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/537 |
| S20 | Digital certificate agreements page | https://www.cdr.gov.au/resources/agreements/digital-certificate-agreements |

Notation:
- "MUST", "SHOULD" and "MAY" in quotes are normative words from the source.
- **UNVERIFIED** means that no primary source confirms the fact. Do not depend on it.
- **ARCHIVED** means that the source is old. A newer source replaces it, or the fact may be out of date.

---

## 1. CDR PKI and certificates

### 1.1 Who operates the CA

- The ACCC, as CDR Registrar, issues certificates to participants. DigiCert is the Certificate Authority. [S7] [S15: 900005826963]
- The ACCC calls these certificates "Register CA certificates" or "CDR certificates". They chain to the "CDR Root CA". [S7] [S15: 900005826963]
- The CDS says: "The CDR utilises a private certificate trust chain for all Register CA secured endpoints being hosted by Data Holders, Data Recipients and the Register." [S1 #certificate-trust-model]
- Since CDS 1.35.1 (08/10/2025), the CDS does not list the CA certificates. The ACCC holds the operational detail at S7. [S1 #certificate-trust-model] [S9]

### 1.2 Trust chains (production and test)

Production has two trust chains with one shared root. [S7] [S8]

| Chain | Root | Intermediate (issuing CA) | Issues certificates |
|---|---|---|---|
| Trust Chain 1 (PKI Platform 8, "PKI8") | CDR Root CA | CDR Banking Intermediate CA | Until 30 September 2025 |
| Trust Chain 2 (DigiCert ONE, "DC1") | CDR Root CA | CDR Intermediate CA 2025 | From 1 October 2025 |

- From 1 October 2025, "CDR participants MUST support certificates issued from both trust chains". [S7]
- Old-chain certificates stay valid until they expire. The ACCC expects all old-chain certificates to expire by 1 November 2026. After that date, participants need only Trust Chain 2. [S7] [S8] [S5 section 1.1A]
- Participants cannot create or renew PKI8 certificates from 1 October 2025. [S8] [S9]
- The ACCC uses the Register's own client certificate to find data holders that do not trust the new chain. [S9]

Revocation endpoints for each intermediate: [S8] [S9]

| | PKI8 (CDR Banking Intermediate CA) | DC1 (CDR Intermediate CA 2025) |
|---|---|---|
| OCSP | http://pki-ocsp.digicert.com | http://ocsp.one.au.digicert.com |
| CRL | http://pki-crl.symauth.com/ca_f85d06b7582eaa18d92979bfa3382474/LatestCRL.crl | http://crl.one.au.digicert.com/CDRIntermediateCA2025.crl |

CA certificate details. I decoded these certificates with `openssl` from the files in S2 and S7:

| Certificate | Subject DN | Key | Signature | Validity | Extensions |
|---|---|---|---|---|---|
| CDR Root CA (production) | C=AU, O=Australian Competition and Consumer Commission, OU=Consumer Data Right, CN=CDR Root CA | RSA 4096 | sha256WithRSAEncryption | 2019-10-17 to 2039-10-16 | basicConstraints critical CA:TRUE; keyUsage critical keyCertSign, cRLSign |
| CDR Banking Intermediate CA (production, PKI8) | C=AU, O=Australian Competition and Consumer Commission, OU=Consumer Data Right, CN=CDR Banking Intermediate CA | RSA 2048 | sha256WithRSAEncryption | 2019-10-17 to 2034-10-16 | basicConstraints critical CA:TRUE, pathlen:0; keyUsage critical keyCertSign, cRLSign; certificatePolicies 1.2.36.1.400.1 (CPS: https://CDR.gov.au/resources); CRL DP http://crl.au.digicert.com/CA/ACCC_CDR_RootCA.crl |
| CDR Intermediate CA 2025 (production, DC1) | C=AU, O=Australian Competition and Consumer Commission, OU=Consumer Data Right, CN=CDR Intermediate CA 2025 | RSA 4096 | sha256WithRSAEncryption | 2025-05-27 to 2035-05-26 | basicConstraints critical CA:TRUE, pathlen:0; keyUsage digitalSignature, keyCertSign, cRLSign; certificatePolicies 1.2.36.1.400.1; CRL DP http://crl.au.digicert.com/CA/ACCC_CDR_RootCA.crl |
| CDR CTS Root CA (Conformance Test Suite) | C=AU, O=Australian Competition and Consumer Commission, OU=Consumer Data Right, OU=For TEST Purposes ONLY, CN=CDR CTS Root CA | RSA 4096 | sha256WithRSAEncryption | 2023-01-27 to 2043-01-27 | CA:TRUE; keyCertSign, cRLSign |
| CDR CTS Intermediate CA | C=AU, O=Australian Competition and Consumer Commission, OU=Consumer Data Right, OU=For TEST Purposes ONLY, CN=CDR CTS Intermediate CA | RSA 2048 | sha256WithRSAEncryption | 2023-01-27 to 2038-01-27 | CA:TRUE; keyCertSign, cRLSign; CRL DP http://ca.cdr.gov.au/crl/cts.crl |
| CDR TEST Root CA (older test chain in S2) | C=AU, O=Australian Competition and Consumer Commission, OU=Consumer Data Right, OU=For TEST Purposes ONLY, CN=CDR TEST Root CA | RSA 4096 | sha256WithRSAEncryption | 2019-10-17 to 2039-10-16 | CA:TRUE |
| CDR TEST Banking Intermediate CA (older test chain in S2) | as above, CN=CDR TEST Banking Intermediate CA | RSA 2048 | sha256WithRSAEncryption | 2019-10-17 to 2034-10-16 | CA:TRUE, pathlen:0; CRL DP http://crl.au.digicert.com/CA/ACCC_CDR_TEST_RootCA.crl |

Notes:
- The CTS chain is separate. "CTS utilizes its own separate trust chain and cannot be used to transact in the production environment." [S7]
- The standards repository still holds the older "CDR TEST" chain. S7 now publishes the "CDR CTS" chain for testing. The relation between the two test chains is **UNVERIFIED**.
- The certificate policy OID 1.2.36.1.400.1 comes from the intermediate certificates. [S2 decoded]

### 1.3 Which endpoints use CDR CA certificates and which use public CA certificates

- "Endpoints specified as MTLS MUST be configured according to the Certificate Trust Model ... Endpoints specified as TLS MUST be configured with a certificate issued by a public CA accepted by major web browsers." [S1 #participant-endpoints]
- For MTLS: "The presented Client transport certificate MUST be issued by the CDR Certificate Authority (CA). The Server MUST NOT trust Client transport certificates issued by other authorities." The same rule applies to the Server transport certificate. [S1 #transaction-security]
- Public endpoints and TLS endpoints "MUST NOT use MTLS". [S1 #transaction-security]
- "For TLS connections, the 'CDR certificates' provisioned for mTLS connections should not be used." [S15: 900005826963]

Register endpoints: [S1 #register-apis] [S10]

| Register API | Caller | Transport | Auth | Host (production) |
|---|---|---|---|---|
| Get OpenId Provider Config | DR brand or software product | TLS | none | api.cdr.gov.au |
| Get JWKS | Data holder | TLS | none | api.cdr.gov.au |
| Token endpoint (`/idp/connect/token`) | Data recipient | MTLS | private_key_jwt | secure.api.cdr.gov.au |
| Get Data Holder Brands | DR brand or software product | MTLS | Bearer | secure.api.cdr.gov.au |
| Get Software Statement Assertion (SSA) | DR brand or software product | MTLS | Bearer | secure.api.cdr.gov.au |
| Get Data Holder Brands Summary | Public client | TLS | none | api.cdr.gov.au |
| Get Data Holder Statuses | Data recipient | TLS | none | api.cdr.gov.au |
| Get Software Products Statuses | Data holder brand | TLS | none | api.cdr.gov.au |
| Get Data Recipients Statuses | Data holder brand | TLS | none | api.cdr.gov.au |
| Get Data Recipients | Data holder brand | TLS | none | api.cdr.gov.au |

- Production base URLs: TLS `https://api.cdr.gov.au`, MTLS `https://secure.api.cdr.gov.au`. [S1 #register-apis]
- The Register has a CDR server certificate for Get Data Holder Brands and Get SSA. The Register has a CDR client certificate to call data holder Get Metrics. [S9, ACCC comment]
- The TLS certificate of `api.cdr.gov.au` (public CA or CDR CA) is **UNVERIFIED**. The CDS rule for TLS endpoints (public CA) suggests a public CA.

Participant endpoints: [S1 #participant-endpoints]

| Base URI | DH brand | ADR brand | Transport |
|---|:-:|:-:|---|
| PublicBaseUri | yes | | TLS |
| ProductBaseUri | yes | | TLS ("MAY be the same as the publicBaseUri") |
| ResourceBaseUri | yes | | MTLS |
| InfoSecBaseUri | yes | | TLS (OIDC discovery only; endpoints in discovery follow the Security Endpoints section) |
| AdminBaseUri | yes | | MTLS ("admin endpoints called by the CDR Register") |
| ExtensionBaseUri | yes | | TLS for public, MTLS for authenticated (optional) |
| RevocationUri | | yes | TLS (in the SSA) |
| RecipientBaseUri | | yes | TLS |
| JwksUri | yes | yes | TLS |

Certificates that the Register CA issues: [S1 #certificate-management]
- Data holders: Server certificate(s). "Certificate is issued to a FQDN." The data holder decides how to split endpoints across domains.
- Data recipients: Client certificate. It secures "Consuming Register APIs" and "Consuming Data Holder APIs". A server certificate is "Not currently required by Data Recipients."
- Data holder server certificates are issued at brand level. SAN certificates for more than one domain are available on request. [S15: 900005881263]

### 1.4 CSR profile (participant certificates)

The CDS gives this profile. [S1 #certificate-signing-request-profile]

| CSR field | Required | Server | Client |
|---|---|---|---|
| Common Name (CN) | Mandatory | Primary DNS name, e.g. `api1.test.entity.com` | Software Product Name |
| SAN | Optional | Secondary DNS names, e.g. `api2.test.entity.com` | N/A |
| Organization (O) | Mandatory | Legal Entity Name or Brand Name (changed in CDS 1.35.1) | Brand Name |
| Organizational Unit (OU) | Mandatory | `Consumer Data Right` | `Consumer Data Right` |
| Country (C) | Mandatory | e.g. `AU` | e.g. `AU` |
| State (ST) | Optional | e.g. `New South Wales` | same |
| Locality (L) | Optional | e.g. `Sydney` | same |
| Email Address | Optional | participant email | participant email |
| Signature Algorithm | Mandatory | `SHA256` | `SHA256` |
| Key Algorithm | Mandatory | `RSA` | `RSA` |
| Key Size | Mandatory | `2048` | `2048` |

- "optional values, if provided, will be validated to be correct." [S1 #certificate-signing-request-profile]
- The RA verifies proof of possession. It checks the signature on the PKCS #10 CSR with the public key in the CSR. [S5 section 3.2.1] [S6]
- The RA checks that the Subject DN is unique for end-entity certificates. [S5 section 3.1] [S6]

### 1.5 Certificate profile limits in the ACCC Certificate Policy

Algorithm and key size (CP Table 5): [S4 section 6.1.5]

| | Root CA | Sub-CA | Subscriber certificate |
|---|---|---|---|
| Digest | SHA-256 or 512 | SHA-256 or 512 | SHA-256 or 512 |
| ECC | P-256 or 384 | P-256 or 384 | P-256 |
| RSA | 4096 | 2048 or 4096 | 2048 or 4096 |

- The CP allows ECC P-256 for subscribers. The CDS CSR profile requires RSA 2048 and SHA256. The CDS profile is the stricter rule for participants. [S1] [S4]

Device Authentication certificate (CP Table 14): [S4 section 10]
- Version V3. Serial number must be unique.
- Issuer signature: sha256WithRSAEncryption or ecdsa-with-SHA256.
- Subject: unique X.500 DN; `cn={ Host URL | Host IP Address | Host Name }`.
- Key: RSA 2048 or ECC 256-bit prime.
- Key Usage: critical; digitalSignature; keyEncipherment optional.
- Extended Key Usage: non-critical; "As per Section 10.9".
- Basic Constraints: critical; end entity; no path length.
- Certificate Policies: non-critical; issuer CP OID.
- Authority Information Access: caIssuers URL and OCSP responder URL.
- CRL Distribution Points: HTTP and/or LDAP URI.
- Subject Alternative Name: non-critical; "always present"; Host URL, IP address or host name.

Extended Key Usage (CP Table 20): [S4 section 10.9]
- End-entity certificates "shall always" have the EKU extension. The EKU "shall not contain anyExtendedKeyUsage". [S4 section 6.1.7]
- The row "Device Authentication, Web Server" lists `id-kp-serverAuth {1 3 6 1 5 5 7 3 1}` and `id-kp-clientAuth {1.3.6.1.5.5.7.3.2}` as required EKUs. All other EKUs are prohibited. (The PDF text layout makes this row hard to read. Check the PDF.)
- The row "Subscriber, Group, Role: Authentication" lists `id-kp-clientAuth` (and smartCardLogon, id-pkinit-KPClientAuth) as required.
- The EKU values in real CDR participant leaf certificates (for example, serverAuth only, clientAuth only, or both) are **UNVERIFIED**. I did not find a real leaf certificate.

Validity periods:
- ACCC FAQ: old-system certificates expire "13 months from date of issuance". [S8]
- The 13-month period for DC1 certificates is **UNVERIFIED**. The FAQ says it about the old system. The expected end of the old chain (1 November 2026, 13 months after 30 September 2025) agrees with a 13-month period. [S8]
- CP maximum: "Server or Device Identity or Signature" 3 years (private key and certificate). Root CA 20 years. Sub CA 15 years. OCSP responder 10 or 15 years. [S4 section 6.3.2 Table 6]
- A CA must not issue a certificate that is valid after the CA certificate expires. [S4 section 6.3.2] [S5 section 5.6]

### 1.6 How participants get certificates

- Participants request certificates in the CDR Participant Portal (https://portal.cdr.gov.au). [S12 steps 5 and 9] [S13 section 10.5]
- Only a primary business contact, primary IT contact or authorised IT contact can request, revoke and renew certificates. [S13 section 10.5]
- Request steps (production): [S13 section 10.5.1]
  1. The data recipient selects "Client certificate". The data holder selects "Server certificate".
  2. The user pastes the CSR. The portal decodes and validates the CSR.
  3. The user enters the certificate common name and an email address. The user agrees to the certificate management policies.
  4. The request goes to the ACCC Technical Operations team. The status shows "submitted".
- Certificate approval emails and renewal reminder emails go to the email address in the request. [S13 section 10.5.1]
- Participants can request certificates for "a brand or software product". [S13 section 10.5.1]
- The ACCC takes "3 – 4 days" to issue a production certificate. [S12 step 9] The indicative table says 1–5 business days for the participant and 3–5 business days for the ACCC. [S12 Table 2]
- Data holders must give their production endpoints before they request a production certificate. [S12 step 9]
- Participants must accept the Subscriber Agreement and the Relying Party Agreement before the Registrar issues certificates. Data recipients must also accept the Trade Mark Licence Agreement. [S11] [S12 step 2]
- The portal note says: "The ACCC is maintaining the existing certificate management approach where they are applied ONLY at the Brand level." [S13 section 10.5] This note conflicts with the CDS client CN (Software Product Name) and with S10 (the Register checks certificates "issued to the software product"). See open questions.

### 1.7 Renewal, re-key and revocation

- Participants can renew a production certificate in the portal. The renew option becomes available 30 days before the certificate expires. The renewal keeps the common name and email address. [S13 section 10.5.5]
- The ACCC notifies participants to renew CTS certificates 30 days before expiry. [S12 step 5]
- A renewed certificate follows the same process as the first issue. [S5 section 4.6]
- Participants can revoke a certificate in the portal. They select a revocation reason. The portal "sends revocation data and reflects the change status in near real time". [S13 section 10.5.3]
- The CA does not support certificate suspension. [S6 section 4.9.13] [S5 section 4.9]
- The ACCC Registrar revokes an ADR's certificates when the ADR is revoked or surrenders. The Registrar reissues certificates when an ADR goes from revoked back to active. [S3 "Status Changes"] (ARCHIVED design, but no newer source contradicts it.)
- For third-party revocation requests, ACCC staff "begin investigating the request within 24 hours after receipt". [S5 section 4.9.3] [S6]
- "ACCC maintains a continuous 24/7 ability to internally respond to any high priority revocation requests." [S6 section 4.9.3]

CRL and OCSP service levels:

| Item | Value | Source |
|---|---|---|
| CRL for end-entity certificates | "issued at least once per day" | S5 section 2.3; S6 |
| CRL for CA certificates | at least every 12 months, and within 18 hours if a CA certificate is revoked | S5 section 2.3 |
| CRL after key compromise | new CRL within 18 hours after notice | S6 section 4.9.12; S15: 900005826963 |
| Maximum time from revocation request to information available to relying parties | "no greater than twenty-four (24) hours" | S6 section 4.9.8; S5 |
| OCSP response time | "no later than six seconds after the request is received" | S6 section 4.9.9 |
| OCSP update, subscriber certificates | "at least every four days" | S6 section 4.10.1; S5 section 4.10.1 |
| OCSP update after revocation | "Within 2 hours of notification" | S15: 900005826963 |
| CRL entry lifetime | serial stays on the CRL until one more CRL after the certificate expires | S5 section 4.10.1 |
| Status service availability | "available 24x7 without interruption" | S6 section 4.10.2 |
| OCSP protocol | RFC 5019 and/or RFC 6960; GET method supported; responder never says "good" for a certificate that was not issued | S6 section 4.9.9 and 4.9.10 |
| Time accuracy | times in certificates, CRLs and status entries are accurate to within 3 minutes | S6 section 6.8 |

- The CA OCSP responder certificate has EKU `id-kp-OCSPSigning` (critical) and `id-pkix-ocsp-nocheck`. [S4 Table 16]
- The CRL is V2. It has a CRL Number that always increases. It has a reason code for key compromise or CA compromise. [S4 Table 17]
- "The use of OCSP Stapling within the CDR ecosystem is not recommended." [S1 #certificate-management]

### 1.8 Certificate validation by participants (relying parties)

The CDS says certificate validation "must check": [S1 #certificate-validation]
1. Certificate validity. The private key signature links to the presented certificate. The certificate is valid and not revoked.
2. Issuer-to-subject name chaining.
3. Policy and key use constraints (for example CA and CRL signing, digital signing, client and server authentication).
4. Revocation status through CRL or OCSP, as given in each certificate in the chain.

Other facts:
- "OCSP is the preferred mechanism for checking the validity of certificates used within the ecosystem." [S15: 900005826963]
- Data sharing cannot proceed if revocation checking is not possible, "either in 'real time' or cached mode". [S15: 900005826963, citing S3]
- If the CA is not available, participants use cached OCSP statuses. When cached statuses expire, data sharing stops. [S3 "Certificate Authority Unavailable"]
- "the current Register design does not require client organisation metadata as published on the GetDataRecipients endpoint to be validated against the certificate in the validation process." [S15: 900005826963]

TLS rules:
- "All HTTP calls MUST be made using HTTPS incorporating TLS >= 1.2." [S1 #transaction-security]
- From 17 March 2025, only cipher suites that BCP195 recommends are permitted, in addition to FAPI 1.0 Advanced section 8.5. Before that date, only `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256` and `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384` were permitted. [S1 #ciphers]
- MTLS "MUST be supported as a Holder of Key (HoK) Mechanism" (RFC 8705). Resource servers must check that the client certificate and the access token match. [S1 #holder-of-key-mechanism]

### 1.9 The ACCC mock CA and test certificates (mock-register)

S16 shows how the ACCC mocks the PKI. [S16 `CertificateManagement/README.md`]
- A self-signed "Mock CDR CA" replaces DigiCert. Subject: `C=AU, ST=ACT, L=Canberra, O=ACCC, OU=CDR, CN=Mock CDR CA`. RSA 4096. SHA-256. Valid 2021-04-05 to 2031-04-03. No intermediate.
- The mock CA issues all mock mTLS and TLS certificates. The mock uses self-signed (mock CA) certificates for TLS endpoints too, not a public CA.
- Mock leaf certificates (decoded): RSA 2048; sha256WithRSAEncryption; about 5 years validity (for example 2023-06-15 to 2028-06-14); basicConstraints CA:FALSE; keyUsage critical digitalSignature, nonRepudiation, keyEncipherment; EKU critical with one value.

| Mock certificate | Subject CN | EKU | SAN |
|---|---|---|---|
| `mtls/register.pem` (Register mTLS server) | register.mock | serverAuth | DNS:localhost, DNS:register.mock |
| `mtls/register-client.pem` (Register client, for Get Metrics) | (not decoded) | (not decoded) | |
| `mtls/server.pem` (mock data holder server) | mock-data-holder | serverAuth | DNS:localhost, DNS:mock-data-holder, DNS:mock-data-holder-energy |
| `mtls/client.pem` (mock data recipient client) | MockDataRecipient | clientAuth | none |
| `tls/mock-register.pem` (Register TLS) | mock-register | serverAuth | DNS:mock-register, DNS:localhost |
| `ssa/ssa.pem` (SSA signing key; self-signed) | MockRegisterSSA | serverAuth | DNS:mockregister.ssa |

- The mock client certificate does not follow the CDS CSR profile (CN is not a software product name; O is `ACCC`). [S16 decoded]
- The repository publishes the private keys and PFX passwords (for example `#M0ckCDRCA#`). These are test secrets only. [S16]
- The mock Register stores `certificates: [{commonName, thumbprint}]` for each software product in seed data. [S16 `seed-data.Development.json`]
- The mock mTLS gateway puts the client certificate thumbprint and common name into HTTP headers for downstream services. [S16 `CDR.Register.API.Gateway.mTLS/Startup.cs`]
- The mock token endpoint accepts the client when a registered certificate of the software product has the same common name. It does not check the thumbprint ("validating the thumbprint is not required"). [S16 `CDR.Register.Infosec/Controllers/TokenController.cs`]
- The real Register "will verify the client certificate used in the request against the list of client certificates that have been issued to the software product". [S10 Token Endpoint page]

---

## 2. Participant lifecycle on the Register

### 2.1 Legal basis

- The ACCC is the Accreditation Registrar. It maintains the Register of Accredited Persons and an associated database. [S14 rules 5.24, 5.25] [S12 section 1.3]
- The Register of Accredited Persons must contain, for each accredited person: name; accreditation number; addresses for service; level of accreditation; sponsor or affiliates; CDR representatives (name, ABN, address); conditions; revocation (fact and date); suspension (fact and period); end of suspension (fact and date); surrender (fact and date); each brand name; links to website and CDR policy. [S14 rule 5.24]
- The associated database holds what the Registrar needs to process requests "in accordance with these rules and the data standards". This includes a list of data holders, their brand names, website and CDR policy links, and the URI of the product data request service. [S14 rule 5.25(1)]
- Participants must give information when the Registrar asks, and must tell the Registrar "as soon as practicable" when their information is out of date. Civil penalty up to $250,000 for a body corporate. [S14 rule 5.25(2)–(5)]
- The Registrar must enter information from the Accreditor "as soon as practicable". [S14 rule 5.26]
- The Registrar must make the rule 5.24 information and the data holder list and brands public. [S14 rule 5.27] The public register is at https://www.cdr.gov.au/find-a-provider. [S15: 900005881263]
- Other Registrar functions: communicate Register information to data holders and accredited persons "in accordance with these rules and the data standards"; maintain "the security, integrity and stability of the Register and associated database", including testing. [S14 rule 5.30]
- The Registrar may automate processes, including decisions. [S14 rule 5.32]

### 2.2 On-boarding steps

The ACCC manages on-boarding. On-boarding starts after data holder registration or data recipient accreditation. [S11] [S12 section 1.2]

| # | Step | Participant time | ACCC time |
|---|---|---|---|
| 1 | Receive on-boarding information | N/A | 1–3 business days |
| 2 | Accept PKI certificate agreements (Subscriber, Relying Party; plus Trade Mark Licence for data recipients) | 1–5 business days | N/A |
| 3 | Enter participation details in the Participant Portal | 1–5 business days | N/A |
| 4 | Enrol in CTS (test environment details) | 1–14 business days | N/A |
| 5 | CSR for test (CTS) certificate | 1 business day | 1 business day |
| 6 | Complete CTS testing | 1–30 business days | 1–30 business days |
| 7 | Give production technical details | 1–5 business days | N/A |
| 8 | CSR for production certificate | 1–5 business days | 3–5 business days |
| 9 | Confirm production environment and readiness (attestation) | 1–14 business days | N/A |
| 10 | Registrar approval and activation on the Register | N/A | 5–14 business days |
| | Total | 3 weeks – 5 months | |

[S12 Table 2]

Details:
- Only the legal authority contact can accept the agreements in the portal. [S12 step 2]
- Production details for accredited data recipients: participation details, brand details, certificate request, authentication details, software product details, software product authentication details, software product endpoints. [S12 step 8]
- Production details for data holders: participation details, brand details, certificate request, authentication details, endpoints. [S12 step 8]
- Data holder brand participation fields: participation type; CDR sector; data sharing purpose (Product Reference Data or Consumer Data Sharing); logo URI; website URL; Public Base URI; Product Base URI. [S13 section 10.2.1]
- Data holder legal entity and brand fields: legal entity website URL, logo URI, CDR policy URL; brand name, description, type, brand group (optional, for white label), participation type, logo URL, website URL, CDR policy URL. [S12 Table 3]
- Data holder endpoints: Public base URI (required), Product base URI, Resource base URI (required), Info sec base URI (required), Admin base URI (required), Extension base URI. [S13 section 10.3.3]
- Data recipient fields: brand participation (participation type, logo URI, website URL, CDR policy URL); brand (name, description, type); software product (name, description). [S12 Table 4]
- Software products for a CDR representative must include the full name of that representative. [S12 step 3]
- Only unrestricted data recipients can create software products. Only a primary IT contact or authorised IT contact can add or edit software products. [S13 section 10.4]
- The SSA claims show what the Register holds for each software product: client name and description, client URI, logo URI, redirect URIs, sector identifier URI, ToS URI, policy URI, JWKS URI, revocation URI, recipient base URI, scope. [S1 #ssa-definition] The mock Register seed data has the same fields. [S16]
- Data holders can run "Validate Production Details" in the portal. It checks authentication details, endpoints, certificates, metadata and supporting URIs. Errors block the activation request. Warnings need a reason. This option is only for data holder brand activations. [S13 section 10.3.4]
- "Brand activation is not automatic." The ACCC assesses and approves it. [S13 section 10.3.4]
- Participants should allow at least 5 business days before the proposed activation date. [S12 step 10]
- A data holder is "immediately discoverable in the ecosystem" when the ACCC makes it active. [S12 step 10]
- After activation, data recipients do Dynamic Client Registration when they are ready. [S12 step 10]
- Changes to production endpoints after activation go through ACCC Technical Operations (CDRTechnicalOperations@accc.gov.au), not self-service. [S13 section 10.3.3]
- For an active brand, a change to the brand name or description needs a "Change Register" request that the ACCC approves. [S13 section 10.1.2]
- A data holder brand that is published only for product reference data has the portal status "PRD Active". A brand that is not published has status "New". [S13 section 10.2.1]
- A PRD-only brand has no `dataHolderBrandId` in Get Data Holder Brands Summary. It has an `interimId` instead. [S10 Get Data Holder Brands Summary] (Known issue: V2 does not return `interimId`. [S10 Known Issues])

### 2.3 Status models

Status enums in the Register API (CDS 1.36.0). [S2 `cds_register.json`]

| Entity | Status values | Returned by |
|---|---|---|
| Data recipient legal entity | ACTIVE, SUSPENDED, REVOKED, SURRENDERED | Get Data Recipients Statuses; Get Data Recipients |
| Data recipient brand | ACTIVE, INACTIVE, REMOVED | Get Data Recipients |
| Software product | ACTIVE, INACTIVE, REMOVED | Get Software Products Statuses; Get Data Recipients |
| Data holder legal entity | ACTIVE, REMOVED | Get Data Holder Statuses (by `legalEntityId`); Get Data Holder Brands (`legalEntity.status`) |
| Data holder brand | ACTIVE, INACTIVE, REMOVED | Get Data Holder Brands |
| Data recipient accreditation level | UNRESTRICTED, SPONSORED | Get Data Recipients |

Rules for the status model:
- "The status of participants is maintained in the CDR Register by the CDR Registrar according to determination of the CDR Accreditor or the participant themselves." [S1 #participant-statuses]
- The CDS shows the state machines as images only (ADRStateMachine.png, SoftwareProductStateMachine.png). [S1 #participant-statuses]
- Cascade when the Registrar changes a data recipient status to a value other than active: [S1 #status-mapping]

| Data recipient status | Cascaded software product status |
|---|---|
| Suspended | Inactive |
| Revoked | Removed |
| Surrendered | Removed |

- "When a Data Recipient status is Suspended, Revoked or Surrendered the Software Product status cannot be Active." [S1]
- "The CDR Register MUST NOT provide an undefined status for Data Recipient or Data Recipient Software Product Statuses." [S1]
- If a data holder gets no status or an unknown status, it "SHOULD ignore the value and use the previous status value". "There is no upper bound for how long previous status values should remain trusted." [S1]
- The Registrar can change a software product status independently of the data recipient status. [S1 #data-holder-responsibilities]
- A "Removed" software product ID is final. The Register never re-issues it. A reinstated product gets a new `softwareProductId`. [S15: 900001657906]
- A surrender, revocation or suspension "takes effect when the fact ... is included in the Register of Accredited Persons". [S14 rule 5.22]

Collection arrangement cascade (provider and principal): [S3 "Accreditation Status Discovery"] (ARCHIVED; not in CDS 1.36.0)

| Provider status | Principal status | Principal software product status |
|---|---|---|
| Active | Active | Active |
| Suspended | Active | Inactive |
| Revoked | Active | Inactive |
| Surrendered | Active | Inactive |
| Any | Suspended | Inactive |
| Any | Revoked | Removed |
| Any | Surrendered | Removed |

Status change events from the archived Register design: [S3 "Status Changes"] (ARCHIVED)
- ADR active: the Registrar issues certificates. Status set to active. The ADR can register software products with data holders.
- ADR suspended: status "suspended"; all software products "inactive".
- ADR reactivated from suspension: status "active". Software products are reactivated "on a case by case basis" only.
- ADR revoked: the Registrar revokes certificates. Status "revoked". Software products "inactive" first, then "removed" after the appeals timeframe.
- ADR reactivated from revoked: the Registrar reissues certificates. Status "active".
- ADR surrendered: the Registrar revokes certificates. Status "surrendered". Software products "removed".
- New software product: status "active".
- Software product deactivated: status "inactive". Reactivated: "active". Removed: "removed".
- Data holder status changes: "Rules regarding Data Holder off-boarding and revocation have not yet been provided."

Conflict: CDS 1.36.0 cascades Revoked to Removed at once. The archived design says Inactive, then Removed after appeals. Standards maintenance issue 431 closed in December 2022 with no change to the standards. [S15: 4414895941391] The real Register behaviour at revocation is **UNVERIFIED**.

Who changes statuses:
- The CDR Accreditor (the ACCC) decides accreditation surrender, suspension and revocation. [S14 rule 5.17]
- Surrender: the accredited person applies in writing. The Accreditor "must, in writing, accept that surrender". [S14 rule 5.17 item 1]
- Suspension and revocation: the Accreditor must first inform the person and give "a reasonable opportunity to be heard". [S14 rules 5.18, 5.20]
- Urgent suspension: the Accreditor can suspend first and hear the person after. [S14 rule 5.21]
- A suspension can end at a specified date or event. The Accreditor can extend or remove it. [S14 rule 5.19]
- The Accreditor must notify the Registrar "as soon as practicable" after accreditation, conditions, surrender, suspension, extension, end of suspension or revocation. [S14 rule 5.15]
- For sponsored accreditation, the Accreditor may suspend or revoke if the sponsorship ends or the sponsor is suspended or revoked. [S14 rule 5.17 item 11]

Temporary Registrar actions:
- The Registrar may stop the Register being used to make consumer data requests to a data holder for up to 10 days, to keep "the security, integrity and stability" of the Register. It may change the data holder's information in the database to do this. [S14 rule 5.33]
- The Registrar may direct an accredited person not to make requests, or a data holder not to respond, for up to 10 days. [S14 rule 5.34]

Archiving:
- "Dormant entries will be identified and archived automatically within the CDR Register." A dormant entry has been in an end state for 2 years. [S1 #archiving]
- End states: data holder Removed; data holder brand Removed; data recipient Surrendered or Revoked; data recipient brand Removed; software product Removed. [S1 #archiving]
- Archived entries do not show in Register API output by default. The archived design had a query parameter `include-archived=true`. [S3] (ARCHIVED. The CDS 1.36.0 OpenAPI spec has no `include-archived` parameter. [S2])
- If a parent entity is archived, all child entities are also archived. [S3] (ARCHIVED)

### 2.4 What each status means for access

Data holder obligations by status: [S1 #data-holder-responsibilities]

| DR status | SP status | Disclose CDR data | Facilitate consent authorisation | Facilitate consent withdrawal | Invalidate consents | Clean up registration |
|---|---|:-:|:-:|:-:|:-:|:-:|
| Active | Active | yes | yes | yes | | |
| Active | Inactive | | | yes | | |
| Active | Removed | | | | yes | yes |
| Suspended | Inactive | | | yes | | |
| Suspended | Removed | | | | yes | yes |
| Revoked | Removed | | | | yes | yes |
| Surrendered | Removed | | | | yes | yes |

- "CDR data MAY only be disclosed to active Data Recipients." Data holders "MUST cease disclosing" when the Registrar suspends or revokes, or the recipient surrenders. [S1 #participant-statuses]
- "The CDR Register MUST notify all Data Holders of the above changes ... via the Register APIs." [S1 #participant-statuses]
- "The status of a Data Recipient Brand does not impact Data Holder responsibilities." [S1]
- Invalidation of consents and clean-up are bulk operations. "It is reasonable to execute these as batch tasks performed overnight." [S1]
- When a software product is Removed, data sharing "must cease immediately". Tokens "must not be honoured". Clean-up means the client ID is de-activated and cannot be re-activated. [S15: 900001657906]
- Suspended ADR: consents stay valid. Sharing stops. Consumers can still withdraw consent. Sharing can resume after the suspension ends. [S15: 900003471766, citing S14 rule 5.23(3)]
- Revoked or surrendered: all consents expire when the revocation or surrender takes effect. [S14 rule 4.14(6)] Data holder authorisations expire "when the data holder is notified of the revocation or surrender". [S14 rule 4.26(2)]
- The accredited person must not collect CDR data after revocation or surrender, or while suspended. [S14 rule 5.23(3)(a)]

Register access by status:
- SSA: "The CDR Register MUST issue Software Statements to active Data Recipients for active Software Products." [S1 #software-statement-assertion-ssa] "The CDR Register will only issue SSAs for Software Products which are currently in the Active state." [S3]
- Access token: "The data recipient software product must be in an ACTIVE state in order to be issued an access token by the Register." [S10 Token Endpoint]
- The mock Register checks for SSA: software product ACTIVE, brand ACTIVE and legal entity (participation) ACTIVE. It returns HTTP 403 with code `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`. [S16 `DataRecipientStatusCheckService.cs`, `SSAController.cs`, `Constants.cs`]
- Whether the real Register also requires brand ACTIVE for the SSA, and which HTTP status and error code it returns, is **UNVERIFIED**. The CDS lists 403 "Invalid BrandId", 404 "Invalid Software Product" and 422 "SSA validation failed" for the SSA endpoint. [S1 #cdr-register-api_get-software-statement-assertion-ssa]
- The Get Data Holder Brands API lists data holder brands "active for consumer data sharing". [S15: 900005881263] Whether it returns INACTIVE and REMOVED brands is **UNVERIFIED**.
- When a data holder brand is removed, ADRs can no longer collect data from it. Use consents continue. [S15: 5173137026831]

### 2.5 Required reactions by other participants

- Data holders "MUST react to Data Recipient and associated Software Statuses changes within 5 minutes of the change occurring on the CDR Register." [S1 #metadata-cache-management]
- Data holders poll Get Data Recipients Statuses, Get Software Products Statuses and Get Data Recipients. They "MUST choose a frequency of polling which ensures their systems can respond to status changes within the required timeframe." [S1]
- Data holders and data recipients "MUST cache all other participant data they use and periodically update this cache using a slow poll." [S1]
- The data holder does not need to poll brand status every 2–5 minutes. [S15: 4411345019791]

Recommended cache periods: [S1 #cache-update-periods]

| API | Type | Period |
|---|---|---|
| GetDataRecipientsStatus | Public | 2–5 minutes |
| GetSoftwareProductsStatus | Public | 2–5 minutes |
| GetDataRecipients (statuses) | Public | 2–5 minutes |
| GetDataRecipients (other data) | Public | 6 hours |
| GetDataHolderBrands | Private | 6 hours |

- JWKS cache: "upper bound is constrained to 15 minutes", and the timeout should be configurable. [S3 "JWKS Cache Management"] (ARCHIVED; not in CDS 1.36.0.)

---

## 3. Register and data holder interactions

### 3.1 Register responsibilities (summary)

The Register: [S3 "Register"] [S1 #register-apis]
1. Lets data recipients discover data holder brands and their endpoints.
2. Lets data holders discover data recipients, brands and software products.
3. Issues SSAs for Dynamic Client Registration.
4. Publishes its public JWKS so that participants trust SSAs and Register-signed JWTs.
5. Collects metrics from data holders with Get Metrics.
6. Publishes statuses of data recipients, software products and data holders.

- Register APIs "are not implemented by Data Holders and Data Recipients." Statuses "are provided authoritatively by the Register". [S1 #register-apis]
- Data holders do not call any authenticated Register API. [S15: 360004662036] (ARCHIVED article, but the CDS 1.36.0 API table agrees: all data-holder-facing APIs are TLS with no auth. [S1])

### 3.2 Metadata Update (Register calls the data holder)

- Endpoint: `POST <AdminBaseUri>/cds-au/v1/admin/register/metadata`. MTLS. Version 1. [S1 #cdr-admin-api_metadata-update]
- Body: `{"data": {"action": "REFRESH"}, "meta": {}}`. `REFRESH` is the only action. It "requires the data holder to call the ACCC to refresh meta data as soon as practicable". [S1]
- Responses: 200 (empty object; the spec shows `null` in error, see known issue), 400 (Invalid Field, Missing Required Field, Invalid Version), 406 (Unsupported Version). [S1] [S1 #known-issues]
- Scope `admin:metadata:update`: "accessible ONLY to the CDR Register". [S1 #authorisation-scopes]
- The ACCC does not use this endpoint. DSB said: "At present the ACCC has not implemented this trigger to broadcast ADR updates." [S19] The archived design says the ACCC "will consider" it "for a future iteration". [S3 "Cache Refresh Metadata Request"]
- Scenarios for a future refresh request: new data holder brand active; new software product active; data holder brand removed; software product deactivated or removed; ADR surrendered, suspended or revoked. [S3] (ARCHIVED)

### 3.3 Get Metrics (Register pulls from the data holder)

- Endpoint: `GET <AdminBaseUri>/cds-au/v1/admin/metrics`. MTLS. Current version 5. Query parameter `period` = `CURRENT`, `HISTORIC` or `ALL` (default `ALL`). [S1 #cdr-admin-api_get-metrics]
- "This endpoint allows the ACCC to obtain operational statistics from the Data Holder (at the Data Holder Brand level)." [S1]
- Scope `admin:metrics.basic:read`: "accessible ONLY to the CDR Register". [S1 #authorisation-scopes]
- AEMO, the AER and the Victorian department are not required to implement Get Metrics. [S1]
- Call frequency: "The Consumer Data Right Register will call DHs once per day, around 5:00AM AEST, every day. Adhoc calls may be required however this would be infrequent and would not occur en masse." [S15: 6140435155471] [S10 Get Metrics page: "typically call these APIs daily at 5:00 AM AEST"]
- The ACCC does not call Get Metrics before a data holder is active for consumer data sharing. [S15: 15490454471695]
- White label providers must give separate Get Metrics endpoints for each brand. [S15: 900005182383]
- If the data holder is unavailable, "The CDR Register will back-fill metrics once the Data Holder becomes available." [S3 "Data Holder Unavailable"] (ARCHIVED)
- The ACCC publishes metrics on the CDR Performance dashboard: https://www.cdr.gov.au/performance. [S15: 900005182383]

How the Register authenticates to data holders: [S1 #cdr-register-calling-data-holders]
- Data holders "MUST support either Private Key JWT Client Authentication or Self-signed JWT Client Authentication of the CDR Register". They "SHOULD" support Private Key JWT.
- Private Key JWT: the data holder issues a client ID to the Register out of band (it "MAY" be `cdr-register`). The Register uses `client_credentials` grant with the scope of the resource endpoint. No refresh token "SHOULD" be included.
- Self-signed JWT: client ID "MUST" be `cdr-register`. The Register sends a JWT as a Bearer token. Example claims: `iss`=`sub`=`cdr-register`, `aud` = AdminBaseUri, `iat`, `exp` (5 minutes after `iat` in the example), `jti`. Algorithm PS256. Data holders must reject reuse of a `jti`.
- The data holder validates the Register JWT with the Register JWKS. [S10 Get JWKS]
- The Register presents its CDR client certificate on the MTLS connection. [S9]
- The data holder can change the method by updating its registration details with the Register. [S1]

### 3.4 How data holders discover and refresh ADR metadata

- Data holders poll the public (TLS, unauthenticated) Register APIs. There is no push. [S1 #metadata-cache-management] [S19]
- Get Data Recipients is "optimised for the presentation of ADR and its associated information on the Data Holder dashboard". DSB recommends a 6-hour poll. [S15: 4411345019791]
- `{industry}` for data-recipient APIs must be `all`. Other values return 400. [S10]
- Data holders get the SSA from the ADR at registration. They verify its signature against the Register JWKS. They verify the registration request JWT against the JWKS at the SSA `jwks_uri`. [S1 #registration-validation]

### 3.5 Dynamic Client Registration depends on the SSA

- "Data Recipients MUST present a valid SSA when undertaking dynamic client registration." [S1 #software-statement-assertion-ssa]
- "Data Holders MUST NOT register Data Recipient software products unless they present a verified Registration Request with a verified SSA." [S1]
- The Register issues a separate SSA for each software product ID. [S1]
- SSA lifetime: "10 minute expiry time". [S1] An ADR can use one SSA for many registrations while it is valid. [S10 SSA page]
- SSA signature: PS256. [S1]
- `iss` = `cdr-register`. [S1 #ssa-definition]
- SSA claims (required unless marked optional): `iss`, `iat`, `exp`, `jti`, `legal_entity_id` (optional), `legal_entity_name` (optional), `org_id` (data recipient brand ID), `org_name`, `client_name`, `client_description`, `client_uri`, `redirect_uris`, `sector_identifier_uri` (optional), `logo_uri`, `tos_uri` (optional), `policy_uri` (optional), `jwks_uri`, `revocation_uri`, `recipient_base_uri`, `software_id` (software product ID), `software_roles` (`data-recipient-software-product`), `scope`. [S1 #ssa-definition]
- `legal_entity_id` and `legal_entity_name` are immutable once provided (open issue 679 may change this). [S15: 4411672618895]
- Data holders "MUST NOT allow multiple active registrations for the same software_id". [S1 #registration-validation]
- One software product has one registration per data holder brand. The client ID is unique per brand. [S15: 900005881263]
- Get SSA path: `GET /cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa`. Version 4. MTLS. Bearer token. [S2] [S10]
- The Register token endpoint: [S1 #register-endpoints] [S1 #data-recipients-calling-the-cdr-register] [S10 Token Endpoint]
  - `grant_type=client_credentials`; `private_key_jwt`; `scope=cdr-register:read`.
  - `client_id`, `iss` and `sub` = data recipient brand ID or software product ID. Brand ID "will be deprecated in the future".
  - Signed with PS256. Keys at the software product `jwks_uri`.
  - Access token is bound to the client certificate (`tls_client_certificate_bound_access_tokens: true`).
  - Access token lifetime: 5 minutes. No refresh token. [S10] (The CDS example shows `expires_in: 7200`, but the example is non-normative.)
  - The mock Register uses a 300-second access token and a 600-second SSA. [S16 appsettings]
- Register OIDC discovery (CDS example): `issuer` `https://api.cdr.gov.au/idp`; `jwks_uri` `https://api.cdr.gov.au/idp/.well-known/openid-configuration/jwks`; `token_endpoint` `https://secure.api.cdr.gov.au/idp/connect/token`; `scopes_supported` `["cdr-register:read"]`; `grant_types_supported` `["client_credentials"]`; `token_endpoint_auth_methods_supported` `["private_key_jwt"]`; `id_token_signing_alg_values_supported` and `request_object_signing_alg_values_supported` `["PS256"]`. [S1 #register-endpoints]
- The Register JWKS for SSA verification is at `https://api.cdr.gov.au/cdr-register/v1/jwks`. [S10 Get JWKS] [S10 Resolved Issues]

### 3.6 Register API details for a mock

| API | Path | Version (x-v) | Caching | Notes |
|---|---|---|---|---|
| Get Data Holder Brands | `/cdr-register/v1/{industry}/data-holders/brands` | 3 | none in spec | `updated-since`, `page` (default 1), `page-size` (default 25) |
| Get Data Holder Brands Summary | `/cdr-register/v1/{industry}/data-holders/brands/summary` | 2 | ETag, 304 | 404 Industry Not Found |
| Get SSA | `/cdr-register/v1/{industry}/data-recipients/brands/{dataRecipientBrandId}/software-products/{softwareProductId}/ssa` | 4 | none | industry `all` |
| Get Data Holder Statuses | `/cdr-register/v1/{industry}/data-holders/status` | 2 | ETag, 304 | |
| Get Software Products Statuses | `/cdr-register/v1/{industry}/data-recipients/brands/software-products/status` | 3 | ETag, 304 | industry `all` |
| Get Data Recipients Statuses | `/cdr-register/v1/{industry}/data-recipients/status` | 3 | ETag, 304 | industry `all` |
| Get Data Recipients | `/cdr-register/v1/{industry}/data-recipients` | 4 | ETag, 304 | industry `all` |
| Get JWKS | `/cdr-register/v1/jwks` | none | | |
| Get OpenId Provider Config | `/idp/.well-known/openid-configuration` | none | | |

[S2 `cds_register.json`] [S10]

- Headers: `x-v` mandatory; `x-min-v` optional; the server responds with the highest supported version; 406 if none. [S1]
- The ACCC deprecated older versions on 11 March 2026. It plans to decommission them in "Q2 2027 (Oct-Dec)". [S10 Change History]
- Errors: `urn:au-cds:error:cds-register:Field/InvalidBrand`, `Field/InvalidIndustry`, `Field/InvalidSoftwareProduct`. [S1 #cdr-register-errors] For an invalid industry, current versions return 400 with `urn:au-cds:error:cds-all:Field/Invalid`. Deprecated versions return 404 with `Field/InvalidIndustry`. [S10 Known Issues]
- CORS: `Access-Control-Allow-Origin: *` on all unauthenticated endpoints (since 14 August 2023). [S10 Resolved Issues]

---

## 4. Non-functional requirements relevant to the Register and its callers

### 4.1 What the CDS says about the Register itself

- The CDS NFRs (availability, performance, traffic thresholds) apply to data holders. They do not set targets for the Register. [S1 #non-functional-requirements] [S15: 6140435155471]
- No published availability target, response-time target or rate limit for the Register was found. Register SLOs are **UNVERIFIED**.
- "The CDR Register MAY cache the responses for public APIs ... If caching is supported the CDR Register MUST support ETag cache validators." [S1 #metadata-cache-management]
- The archived design says: "All public APIs are hosted on a CDN supporting ETag cache validators." Clients can rely on 304 Not Modified. [S3 "ETag Support"] (ARCHIVED)

### 4.2 Propagation and latency

- Status changes: data holders must react within 5 minutes of the change on the Register. [S1]
- A status change on the Register is the legal trigger. A surrender, revocation or suspension takes effect when it is on the Register. [S14 rule 5.22]
- Recommended poll: 2–5 minutes for statuses; 6 hours for other data. [S1]
- The Register must publish status changes through the Register APIs. [S1]
- The ACCC portal says certificate revocation status shows "in near real time" in the portal. [S13 section 10.5.3]
- CRL latency: up to 24 hours from request to relying parties. OCSP after revocation: within 2 hours. [S6] [S15: 900005826963]

### 4.3 Outages and backoff

- The Register uses a CDN for its public APIs. During Register outages, participants use the CDN copies. If the CDN is down, participants use their caches. [S3 "CDN Hosted Endpoints"] (ARCHIVED)
- If the Register is unavailable: "It is not expected that the status of an ADR or Software Product will change when the CDR Register or CDN is unavailable". Data holders continue until ACCC Operations tells them to stop. Data recipients cannot get SSAs, so they cannot register. [S3 "CDR Register Unavailable"] (ARCHIVED)
- Long-term backoff: exponential; wait = 2^retries × 100 ms; maximum wait 1 hour; total 7 days. [S3 "Backoff Patterns"] [S15: 900004394866]
- Short-term backoff: one retry each minute for 10 minutes. [S3] [S15: 900004394866]
- Planned outage norms for the Register (notice period, maintenance windows) are **UNVERIFIED**. The CDS rule for data holders is at least one week notice for planned outages, published with Get Outages. [S1 #availability-requirements]

### 4.4 Data holder NFRs that callers of a mock data holder will expect (context only)

- Availability: 99.5% per month; planned outages excluded; one week notice. [S1 #availability-requirements]
- Performance: 95% of calls each hour within the threshold. High priority (InfoSec endpoints including DCR) 1000 ms. Unauthenticated 1500 ms. Unattended 4000 ms. Large payload 6000 ms. [S1 #performance-requirements]
- DCR must meet the NFR performance times. Extra checks (for example outbound allowlisting of recipient endpoints) must not prevent this. [S1 #registration-validation]
- Traffic thresholds: public traffic 300 TPS total. Secure traffic 150–450 peak TPS by number of active authorisations. [S1 #traffic-thresholds]

---

## 5. Industries and sectors

### 5.1 Designated and live sectors (at 2026-09-25)

| Sector | Designated | Status | Source |
|---|---|---|---|
| Banking | yes | Live since July 2020 | S17 rollout |
| Energy | yes | Product data since October 2022; consumer data since November 2022 | S17 rollout |
| Non-bank lending | 21 November 2022 | Product data from 13 July 2026 (live); consumer data from 9 November 2026 (initial providers) and 10 May 2027 (large providers) | S17 NBL page |
| Telecommunications | Consumer Data Right (Telecommunications Sector) Designation 2022 | Paused. No rules made. Treasury planned a strategic assessment towards the end of 2024. | S18 |

- In the NBL sector the rules do not apply to "complex requests" (secondary users, nominated representatives, joint or partnership accounts). [S17 NBL page]
- The current (2026) status of the telecommunications pause is **UNVERIFIED** beyond S18.

### 5.2 How the Register models industry

- `IndustryEnumV2` (path parameter for data holder APIs): `banking`, `energy`, `non-bank-lending`, `telco`, `all`. [S2 `cds_register.json`]
- `IndustriesEnumV2` (the `industries` array on a data holder brand): `banking`, `energy`, `non-bank-lending`, `telco`. [S2]
- `IndustryEnumRecipients` (path parameter for data recipient APIs and SSA): `all` only. [S2]
- Data recipients and software products are not specific to a sector. `{industry}` "must be set to all". Other values return 400. [S10]
- For data holder APIs, `{industry}` filters brands by sector. The ACCC portal documents `all`, `banking`, `energy` and `non-bank-lending`. It does not document `telco`. [S10]
- A data holder brand has an `industries` array. A brand can belong to more than one industry. [S2] Data holders may use one brand for many sectors or separate brands for each sector. [S15: 900005881263]
- Participants set the CDR sector for each brand participation in the portal. [S13 section 10.2.1]
- The mock Register added non-bank lending on 6 February 2026 (migration `AddNonBankLendingIndustry`). [S16]
- CDS 1.36.0 adds `brandGroup` to data holder brands for white label arrangements (Decision 376). [S1 #register-apis] [S15: 900005881263]
- CDS 1.35.0 adds `productBaseUri` to data holder brands. [S1 #register-apis]

---

## 6. ACCC mock-register behaviour (reference for comparison)

- Components: TLS gateway (:7000), mTLS gateway (:7001), identity provider (:7002), Discovery API (:7003), Status API (:7004), SSA API (:7005), Admin API (:7006). [S16 README]
- The Admin API is not part of the Register standard. It loads and exports metadata (`GET/POST /admin/metadata`, `POST /admin/metadata/data-holders`, `POST /admin/metadata/data-recipients`). It has no authentication unless `oidc.issuer` is set. [S16 README]
- Seed data fields for data holders: legal entity (ID, name, logo, ABN, ACN, ARBN, registration details, ANZSIC division, organisation type, status), participations (type, industry, status), brands (ID, name, brand group, logo, status, last updated, auth details `registerUTypeId` and `jwksEndpoint`, endpoints `publicBaseUri`, `productBaseUri`, `resourceBaseUri`, `infoSecBaseUri`, `extensionBaseUri`, `websiteUri`). [S16 seed-data]
- Seed data fields for software products: ID, name, description, logo URI, sector identifier URI, client URI, ToS URI, policy URI, recipient base URI, revocation URI, redirect URIs, JWKS URI, scope, status, certificates (common name, thumbprint). [S16 seed-data]
- The mock data holder seed data does not store `adminBaseUri`. [S16 seed-data] So the mock Register does not appear to call Get Metrics or Metadata Update. This is **UNVERIFIED** (I did not read all mock source).

---

## 7. Conflicts and gaps found

1. Revocation cascade: CDS says Revoked → software products Removed. The archived design says Inactive, then Removed after appeals. [S1] [S3] [S15: 4414895941391]
2. Certificate level: the CDS client CSR CN is the software product name, and the Register checks certificates "issued to the software product" [S1] [S10]. The portal guide says certificates apply "ONLY at the Brand level" [S13].
3. Access token lifetime: CDS example 7200 s (non-normative); ACCC portal 5 minutes. [S1] [S10]
4. Register SSA error for a non-active product: CDS lists 403 "Invalid BrandId", 404, 422. The mock returns 403 with `AdrStatusNotActive`. The real behaviour is **UNVERIFIED**. [S1] [S16]
5. Test PKI: S2 holds "CDR TEST Root CA" and "CDR TEST Banking Intermediate CA". S7 publishes "CDR CTS Root CA" and "CDR CTS Intermediate CA". [S2] [S7]
6. Metadata Update is in the standards but the ACCC has not implemented the Register trigger. [S1] [S19]
7. Participant leaf certificate EKU and exact validity for DC1 certificates are not published in a source that I could read. [S4] [S8]

## 8. Open questions for requirements

- Must the mock PKI copy the two-chain (PKI8 and DC1) period, or only the current chain (DC1)? The dual-chain period ends about 1 November 2026.
- Must the mock PKI run OCSP and CRL services? Participants must check revocation, and the CDS prefers OCSP.
- Should the mock issue leaf certificates with 13-month validity, and with serverAuth, clientAuth or both EKUs?
- Should the mock bind client certificates to software products (CDS and S10) or to brands (S13)?
- Should the mock Register return the CDS error codes or copy the ACCC mock (403 `AdrStatusNotActive`)?
- Should the mock Register implement Get Metrics collection (daily at 05:00 AEST) and the Metadata Update trigger (not used in production)?
- Should the mock support the `telco` industry value? It is in the enum but the sector is paused.
- Which revocation cascade (immediate Removed, or Inactive then Removed) should the mock use?
- What availability, latency and rate-limit targets should the mock have? No production Register SLOs are published.

## 9. Source URLs for ACCC Developer Portal pages (S10)

| Page | URL |
|---|---|
| Register API | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29982721/Register+API |
| Token Endpoint | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139 |
| Get Data Holder Brands | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081113 |
| Get Data Holder Brands Summary | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982 |
| Get Data Holder Statuses | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29949976 |
| Get Data Recipients | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29949983 |
| Get Data Recipients Statuses | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851989 |
| Get Software Products Statuses | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081126 |
| Get SSA | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30015535 |
| Get JWKS | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30015542 |
| Get OpenId Provider Config | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30015552 |
| Get Metrics | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/32800769/Get+Metrics |
| Known Issues | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29950003 |
| Resolved Issues | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29949959 |
| Delivery Dates | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081043 |
| Change History | https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/42827777 |

CDR Support Portal articles used (S15), base URL https://cdr-support.zendesk.com/hc/en-us/articles/:
900005826963 Certificate Validation; 900005881263 Brands in the CDR ecosystem; 4411345019791 Data Holder cache update periods; 900001657906 Status of consent when an ADR Software Product is "Removed"; 900003471766 Withdrawing vs invalidating consent; 4414895941391 [ARCHIVE] Register revoked ADR with Inactive product; 5173137026831 Data holder brand removed; 4411672618895 ADR legal entity name modification; 6140435155471 Guidance on metrics; 900005182383 CDR Register Metrics Collection; 15490454471695 NBL Get Metrics expectations; 900004394866 Backoff patterns; 360004662036 [ARCHIVE] Data Holder Authentication with Register; 7313594692367 Certificates and cipher suites (older cipher list; the CDS BCP195 rule replaces it).
