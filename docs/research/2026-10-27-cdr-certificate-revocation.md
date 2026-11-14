# CDR certificate revocation and the CRL

- Research date: 2026-10-27
- Consumer Data Standards (CDS) version: 1.36.0. It is still the current version: the live site says "version 1.36.0". [S1]
- Purpose: facts that the earlier notes do not have, for the requirements of the revocation of a CDR certificate by the Register and of the CRL that the Register signs and serves. The note [CDR PKI and participant lifecycle](2026-09-25-cdr-pki-and-participant-lifecycle.md) ("the PKI note"), section 1.7, is still correct for this scope. This note adds to it and does not replace it.

Check the sources again before you depend on a fact; the rules and standards change.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site) | https://consumerdatastandardsaustralia.github.io/standards/ |
| S2 | ACCC Certificate Policy v1.04 (CP) | https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf |
| S3 | ACCC Certification Practice Statement v2.0 (CPS) | https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf |
| S4 | Participant Portal User Guide v23, 14 October 2026 | https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf |
| S5 | RFC 5280 | https://www.rfc-editor.org/rfc/rfc5280 |
| S6 | RFC 9110 | https://www.rfc-editor.org/rfc/rfc9110 |

## 1. The revocation request

- The CDS gives no revocation request. It requires only the revocation check: "Status is checked through Certificate Revocation Lists (CRL) or Online Certificate Status Protocol (OCSP) responders, identified in each certificate in the chain." [S1 #certificate-validation]
- "A request to revoke a certificate shall identify the certificate to be revoked, explain the reason for revocation, and allow the request to be authenticated (eg, digitally or manually signed)." [S2 section 4.9.3]
- "Revocation requests must be authenticated. Requests to revoke a certificate may be authenticated using that certificate's associated public key, regardless of whether the private key has been compromised." [S3 section 4.9]
- Server and client certificates: "Any valid certificate provides a link from which it can be revoked." The portal shows a summary of the certificate and asks for a confirmation. [S4 sections 10.6.4 and 10.7.3] For a CTS certificate, the user also selects a reason from a list. [S4 section 11.3.4] Thus the portal identifies the certificate from its own list of the certificates that it issued.
- "ACCC will transition any revocation reason code in a CRL to 'key compromise' upon discovery of such reason". [S3 section 4.9.12]
- No source says if the Register removes a client certificate from a software product when the CA revokes it.

## 2. When the CA issues a CRL

- "CRLs shall be issued periodically, even if there are no changes to be made, to ensure timeliness of information." Routine frequency: "At least once every 24 hours". [S2 section 4.9.7, Table 1]
- "A CA shall ensure that superseded certificate status information is removed from the PKI repository upon posting of the latest certificate status information." [S2 section 4.9.7]
- "Certificate status information shall be published no later than the next scheduled update. This will facilitate the local caching of certificate status information". [S2 section 4.9.7]
- `nextUpdate`: "Indicates the date by which the next CRL will be issued." In the CRL profile, `nextUpdate` is ">= thisUpdate + CRL issuance frequency". CRL Number: "monotonically increasing integer (never repeated)". [S2 section 10, Table 17]
- "Care is taken by the CA to ensure that the public copy of the CRL is replaced automatically when it is being updated." [S3 section 4.10.2]
- "The matter of how often new revocation data should be obtained is a determination to be made by the Relying Party". When revocation information cannot be obtained, the relying party "must either reject use of the certificate, or make an informed decision to accept the risk". [S2 section 4.9.6]

## 3. RFC rules for the CRL and its server

- "Conforming CRL issuers MUST include the nextUpdate time in all CRLs." "CRL issuers SHOULD issue CRLs with a nextUpdate time equal to or later than all previous CRLs." [S5 section 5.1.2.5]
- The revocation date is "the date on which the revocation occurred". [S5 section 5.1.2.6]
- The CRL number is "a monotonically increasing sequence number for a given CRL scope and CRL issuer". "CRL issuers MUST NOT use CRLNumber values longer than 20 octets." [S5 section 5.2.3]
- "the reason code CRL entry extension SHOULD be absent instead of using the unspecified (0) reasonCode value." [S5 section 5.3.1]
- "All general-purpose servers MUST support the methods GET and HEAD." [S6 section 9.1] A 405 response "MUST generate an Allow header field". [S6 section 15.5.6]
