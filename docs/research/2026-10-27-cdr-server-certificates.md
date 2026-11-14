# CDR server certificates

- Research date: 2026-10-27
- Consumer Data Standards (CDS) version: 1.36.0. It is still the current version: the live site says "These standards represent version 1.36.0". [S1]
- Purpose: facts that the earlier notes do not have, for the requirements of the issue of a CDR server certificate by the Register: the record that owns a server certificate, the endpoints that need one, and the checks of a server CSR. The note [CDR PKI and participant lifecycle](2026-09-25-cdr-pki-and-participant-lifecycle.md) ("the PKI note"), sections 1.3, 1.4 and 1.6, is still correct for this scope. This note adds to it and does not replace it.

Check the sources again before you depend on a fact; the rules and standards change.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site; the archive URL redirects to it while 1.36.0 is current) | https://consumerdatastandardsaustralia.github.io/standards/ |
| S2 | CDR Support Portal: Brands in the CDR ecosystem (900005881263, updated 10 March 2026; read through the help centre API) | https://cdr-support.zendesk.com/hc/en-us/articles/900005881263 |
| S3 | Participant Portal User Guide v23, 14 October 2026 | https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf |
| S4 | ACCC Certificate Policy v1.04 (CP) | https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf |
| S5 | ACCC Certification Practice Statement v2.0 (CPS) | https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf |

S3 replaces the guide of 28 August 2026 that the PKI note cites. That URL now gives 404.

## 1. The record that owns a server certificate

- CDS, Certificate Management, "Issued by the Register for Data Holders": "Server Certificate(s)". "Certificate is issued to a FQDN. Secures the endpoints as detailed in Participant endpoints." "It will be up to the DH on how these endpoints are segregated. They may all be on the one domain (so only one certificate required) or could be separated." [S1 #certificate-management]
- CDS, Participant Endpoints: "Participants will be required to register base URIs against each of their brands". [S1 #participant-endpoints]
- "Server certificates are issued at a brand level to data holders. SAN certificates used to protect endpoints spanning multiple domains can also be issued by request and reference to this information is stored in the Register." [S2, section "Certificates"]
- "Server certificates are required for a Data Holder." "To manage Production certificates for a brand, please follow the instructions below": the user selects the brand, then the brand participation, then "Production details". [S3 sections 10.6 and 10.6.1]
- "The ACCC is maintaining the existing certificate management approach where they are applied ONLY at the Brand level." [S3 section 10.5, "Maintain Public Key Infrastructure (PKI) certificates", directly before section 10.6]

Thus each source that names a record gives the data holder brand. No source gives a server certificate to a data holder legal entity, a data recipient or a software product.

## 2. Data recipients and server certificates

- CDS, Certificate Management, "Issued by the Register CA for Data Recipients": the client certificate "Secures the following: Consuming Register APIs. Consuming Data Holder APIs." For "Server Certificate(s)": "Certificate is issued to a FQDN." "Not currently required by Data Recipients." [S1 #certificate-management]
- CDS, Participant Endpoints, the ADR brand rows: `RevocationUri` is TLS, "Used for consent withdrawal notification from a Data Holder and is populated in the SSA". `RecipientBaseUri` is TLS, "This MUST be the base to provide reference to Data Recipient Endpoints." `JwksUri` is TLS. [S1 #participant-endpoints]
- CDS, CDR Arrangement Revocation endpoint, "Data Recipient hosted endpoint": "Data Recipient Software Products MUST expose their CDR Arrangement Revocation endpoint under their recipient_base_uri published in their Software Statement Assertion." "Data Holders must be authenticated when they call this endpoint according to the guidance in the Client Authentication section." [S1 #cdr-arrangement-revocation-endpoint]
- "Endpoints specified as TLS MUST be configured with a certificate issued by a public CA accepted by major web browsers." [S1 #participant-endpoints]

Thus in CDS 1.36.0 no data recipient endpoint is MTLS. A data holder authenticates to the data recipient with a signed JWT, not with a CDR client certificate.

## 3. The checks of a server CSR

CDS CSR profile, server column: CN "Primary DNS Name e.g. api1.test.entity.com"; SAN, optional, "Secondary DNS Name(s)"; O, mandatory, "Legal Entity Name or Brand Name". The change history gives the O change: "- Brand Name + Legal Entity Name or Brand Name" (CDS 1.35.1, Decision 377). "Note: optional values, if provided, will be validated to be correct." [S1 #certificate-signing-request-profile]

The Participant Portal checks a server CSR when the user decodes it. [S3 section 10.6.2]

- "These validations include mandatory CSR attributes, Fully Qualified Domain Name checks, and cryptographic requirements".
- "Errors will prevent a form from being submitted." "Warning messages will still allow submission, but may alert you to a problem with a field".
- The example screen gives these messages:
  - Error: "Common Name (CN) is not valid. A valid FQDN contains only letters (a-z, A-Z), numbers (0-9), hyphens and dots; no wildcards, IP addresses, URLs, paths, spaces, unsupported characters, or invalid DNS lengths."
  - Error: "Organisation (OU) is not valid. Must be Consumer Data Right"
  - Error: "The country code value is invalid."
  - Warning: "Organisation (O) does not match the registered Brand Name or Legal Entity (LE) Name. Submission is allowed but the request may be rejected on subsequent review."
  - The review box: "Warnings do not prevent submission. Requests submitted with warnings only will be subjected to review and may be rejected".
- The portal shows the CN and the SANs from the CSR: "Certificate common name and any SANs are decoded from the CSR and displayed." [S3 section 10.6.1]
- No source found on 2026-10-27 says that the portal or the ACCC compares the CN or the SANs with the registered endpoints of the brand.

The CP and the CPS:

- "Subject DNs shall accurately reflect the organisation with which the Subject is affiliated." For end-entity equipment, the name can be "a fully qualified domain name (www.example.com)". [S4 section 3.1.2]
- "Subject Alternative Name may be used, if marked non-critical; section 10 lists the accepted contents (email address, UPN, FQDN)". [S4 section 3.1.1]
- "ACCC does not verify or determine whether a Subscriber has Intellectual Property Rights in the name appearing in a DN or to arbitrate, mediate, or otherwise resolve any dispute concerning the ownership of any domain name". [S5 section 3.1.6]
- A device certificate application gives a "service name (eg, DNS Fully Qualified Domain Name, IP address ...) sufficient to uniquely identify the Subject". [S5 section 3.2.3.1] The CPS thus permits an IP address for a device in general, but the portal refuses it for a CDR server CSR.

## 4. Conflicts and gaps

1. The O of a server CSR: the CDS says that values "will be validated to be correct" [S1]. The portal gives only a warning, and the ACCC can refuse the request later on review [S3].
2. The CDS and the portal do not say if the check of the O is exact, or if it ignores case or spaces.
3. The CDS gives the CN as a "Primary DNS Name". The portal says that the CN is a valid FQDN, but its rule does not require a dot. No source gives a rule for each SAN; the portal example shows only the CN rule.
4. A brand that shares only product data has no `dataHolderBrandId`, only an `interimId` ([ACCC Developer Portal: Get Data Holder Brands Summary](https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982)). It has no MTLS endpoints. No source says if the ACCC issues a server certificate to such a brand.
