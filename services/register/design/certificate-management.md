# Mock CDR Register: Certificate management

This area gives the issue of CDR client certificates and CDR server certificates from a CSR, and the CDR CA bundle that the Register gives. In a running lab, the Register is the mock CDR intermediate CA. When it starts, it [reads the intermediate key file](../../../pki/DESIGN.md#pki-0075). The [CA library](../../../pkg/ca/DESIGN.md) gives the CSR profiles and the certificate profiles. In production, the ACCC CA issues the certificates through the Participant Portal, and the Register is not a CA.

## Intermediate CA

- <a id="reg-0600"></a>**REG-0600** In a running lab, the Register issues CDR client certificates and CDR server certificates from a CSR, and gives the CDR CA bundle. ([ACCC Developer Portal: Certificate Management][accc-cert-mgmt]; [Judgement: the Register is the CDR intermediate CA (ADR 0022)][adr-0022])

## Client certificates

- <a id="reg-0601"></a>**REG-0601** The lab operator can get a CDR client certificate for a software product from a CSR, through the operator interface. ([CDS: Certificate Management][cds-cert-mgmt]; [Judgement: the Register is the CDR intermediate CA (ADR 0022)][adr-0022])
- <a id="reg-0602"></a>**REG-0602** The request body of an issue is one PEM CSR. When the body is not one PEM CSR, the response is 400. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="reg-0603"></a>**REG-0603** The Register never gets the private key of a certificate that it issues. ([ACCC CPS, section 3.2.1][accc-cps])
- <a id="reg-0604"></a>**REG-0604** The Register checks each CSR with [the CSR rules of the CA library](../../../pkg/ca/DESIGN.md#ca-0010), and issues the certificate with [the profile of a CDR certificate](../../../pkg/ca/DESIGN.md#ca-0025). ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="reg-0605"></a>**REG-0605** When a CSR [breaks a rule of the CSR profile](../../../pkg/ca/DESIGN.md#ca-0011), the response is 400, and the `detail` names each rule that the CSR breaks. ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="reg-0606"></a>**REG-0606** When the CN of a client CSR subject is not the software product name with the same characters, the response is 400. The `detail` names the field. ([CDS: Certificate Signing Request Profile][cds-csr-profile]; [Judgement: exact participant names in a client CSR](#reg-0640))
- <a id="reg-0607"></a>**REG-0607** When the O of a client CSR subject is not the name of the brand of the software product with the same characters, the response is 400. The `detail` names the field. ([CDS: Certificate Signing Request Profile][cds-csr-profile]; [Judgement: exact participant names in a client CSR](#reg-0640))
- <a id="reg-0608"></a>**REG-0608** When the software product name or the brand name has more than 64 characters, a client certificate issue gives 400. The `detail` gives the limit. ([RFC 5280, Appendix A][rfc5280-bounds]; [Judgement: no certificate for a name of more than 64 characters](#reg-0641))
- <a id="reg-0609"></a>**REG-0609** When the software product is REMOVED, a client certificate issue gives 409. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: no certificate for a REMOVED record](#reg-0642))
- <a id="reg-0610"></a>**REG-0610** A client certificate issue also registers the certificate for the software product, with [the rules of a registration](operator-interface.md#reg-0834). ([Judgement: a client certificate issue also registers the certificate](#reg-0643))
- <a id="reg-0611"></a>**REG-0611** After a client certificate issue, the token endpoint accepts the certificate at once. For example, the next client assertion with the certificate gets an access token. ([Judgement: a client certificate issue also registers the certificate](#reg-0643))
- <a id="reg-0612"></a>**REG-0612** The response of an issue gives the certificate and then the intermediate certificate, in PEM, with the content type `application/pem-certificate-chain`. ([RFC 8555, section 9.1][rfc8555-chain]; [Judgement: an issue gives 200 and the certificate chain with no root](#reg-0644))
- <a id="reg-0613"></a>**REG-0613** The response of an issue has no `Location` header. ([RFC 9110, section 15.3.2][rfc9110-201]; [Judgement: an issue gives 200 and the certificate chain with no root](#reg-0644))
- <a id="reg-0614"></a>**REG-0614** The Register keeps only the SHA-256 thumbprint of a certificate that it issues. Thus the lab operator cannot get the certificate again. ([Judgement: an issue gives 200 and the certificate chain with no root](#reg-0644))

## CDR CA bundle

- <a id="reg-0615"></a>**REG-0615** The lab operator can get the CDR CA bundle through the operator interface. ([ACCC Developer Portal: Certificate Management][accc-cert-mgmt]; [Judgement: the operator interface gives the CDR CA bundle](#reg-0645))
- <a id="reg-0616"></a>**REG-0616** The CDR CA bundle is the intermediate certificate and then the root certificate of the mock CDR CA, in PEM, with the content type `application/pem-certificate-chain`. ([RFC 8555, section 9.1][rfc8555-chain]; [Judgement: the CDR CA bundle is the intermediate and then the root](#reg-0646))
- <a id="reg-0617"></a>**REG-0617** The CDR CA bundle that the operator interface gives is [the CDR CA bundle that the mTLS listeners of the Register trust](../../../pkg/cds/DESIGN.md#cdsl-0045). ([CDS: Certificate Trust Model][cds-cert-trust-model])

## Server certificates

- <a id="reg-0618"></a>**REG-0618** The lab operator can get a CDR server certificate for a data holder brand from a CSR, through the operator interface. ([CDS: Certificate Management][cds-cert-mgmt]; [CDR Support: Brands in the CDR ecosystem][support-brands]; [ACCC Participant Portal User Guide, section 10.6.1][accc-portal-guide])
- <a id="reg-0619"></a>**REG-0619** The Register issues no server certificate for a data holder legal entity. ([CDR Support: Brands in the CDR ecosystem][support-brands])
- <a id="reg-0620"></a>**REG-0620** The Register issues no server certificate for a data recipient, a data recipient brand or a software product. ([CDS: Certificate Management][cds-cert-mgmt]; [Judgement: no server certificate for a data recipient](#reg-0647))
- <a id="reg-0621"></a>**REG-0621** The request body and the response of a server certificate issue are [those of a client certificate issue](#reg-0602). ([Judgement: an issue gives 200 and the certificate chain with no root](#reg-0644))
- <a id="reg-0622"></a>**REG-0622** The Register checks a server CSR with [the rules of a server CSR in the CA library](../../../pkg/ca/DESIGN.md#ca-0016). ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="reg-0623"></a>**REG-0623** When the O of a server CSR subject is not the brand name or the legal entity name with the same characters, the response is 400. The `detail` gives the two names that are permitted. ([CDS: Certificate Signing Request Profile][cds-csr-profile]; [Judgement: an exact brand or legal entity name as the O of a server CSR](#reg-0649))
- <a id="reg-0624"></a>**REG-0624** When the brand name has more than 64 characters, the O of a server CSR must be the legal entity name. If not, the response is 400, and the `detail` gives the legal entity name and the limit. ([RFC 5280, Appendix A][rfc5280-bounds]; [Judgement: no certificate for a name of more than 64 characters](#reg-0641))
- <a id="reg-0625"></a>**REG-0625** The Register does not compare the CN and the DNS names of a server CSR with the endpoints of the brand. For example, a CSR with the DNS name `localhost` gets a certificate. ([CDS: Certificate Management][cds-cert-mgmt]; [Judgement: no comparison of a server CSR with the endpoints of the brand](#reg-0650))
- <a id="reg-0626"></a>**REG-0626** When the brand is REMOVED, a server certificate issue gives 409. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: no certificate for a REMOVED record](#reg-0642))
- <a id="reg-0627"></a>**REG-0627** An INACTIVE brand can get a server certificate. ([Judgement: no certificate for a REMOVED record](#reg-0642))
- <a id="reg-0628"></a>**REG-0628** When the brand has no `dataHolderBrandId`, a server certificate issue gives 409. ([Judgement: no server certificate for a brand with no `dataHolderBrandId`](#reg-0648))
- <a id="reg-0629"></a>**REG-0629** The Register does not register a server certificate. ([Judgement: a server certificate issue is an event, not a registration](#reg-0651))
- <a id="reg-0630"></a>**REG-0630** The Register records a server certificate issue as one event of the data holder, with the brand and the thumbprint of the certificate. ([Judgement: a server certificate issue is an event, not a registration](#reg-0651))
- <a id="reg-0631"></a>**REG-0631** A server certificate issue sets the `lastUpdated` of each brand of the data holder, as each other event of the data holder does ([the `lastUpdated` of a brand](register-apis.md#reg-0336)). ([Judgement: a server certificate issue is an event, not a registration](#reg-0651))

## Operator interface paths

| ID | Action | Method and path | Success | Source |
|---|---|---|---|---|
| <a id="reg-0632"></a>**REG-0632** | Issue a client certificate for a software product from a CSR, and register it | `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue` | 200 | [Judgement: an issue gives 200 and the certificate chain with no root](#reg-0644) |
| <a id="reg-0633"></a>**REG-0633** | Get the CDR CA bundle | `GET /v1/cdr-ca-bundle` | 200 | [Judgement: the operator interface gives the CDR CA bundle](#reg-0645) |
| <a id="reg-0634"></a>**REG-0634** | Issue a server certificate for a data holder brand from a CSR | `POST /v1/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue` | 200 | [Judgement: an issue gives 200 and the certificate chain with no root](#reg-0644) |

## Judgements

### <a id="reg-0640"></a>REG-0640 Exact participant names in a client CSR

- Kind: choice
- Basis: Step 3.3. The [CDS: Certificate Signing Request Profile][cds-csr-profile] gives the software product name as the CN and the brand name as the O, and says that the values "will be validated to be correct". It does not say how. The certificate has [the subject of its CSR](../../../pkg/ca/DESIGN.md#ca-0025).
- Reason: Only the Register knows the names ([only the Register checks the participant names of a CSR](../../../pki/DESIGN.md#pki-0081)). An exact comparison is the least complex check.

### <a id="reg-0641"></a>REG-0641 No certificate for a name of more than 64 characters

- Kind: conflict
- Basis: Step 3.2. [RFC 5280, Appendix A][rfc5280-bounds] limits a CN and an O to 64 characters. The CDS permits a software product name and a brand name of 200 characters ([CDS: Register API schemas][cds-register-schemas]), and does not say what to do.
- Reason: A certificate with a shorter name does not obey the CSR profile, so no correct certificate is possible. The `detail` gives the limit, so that the lab operator can change the name. A legal entity name has 36 characters or fewer, so it can always be the O of a server certificate.

### <a id="reg-0642"></a>REG-0642 No certificate for a REMOVED record

- Kind: choice
- Basis: Step 3.3. REMOVED is a final status ([CDS: Participant Statuses][cds-participant-statuses]). The CDS does not say if a REMOVED record can get a certificate.
- Reason: A certificate of a REMOVED record can never be used. An INACTIVE record can become ACTIVE again, so it can get a certificate.

### <a id="reg-0643"></a>REG-0643 A client certificate issue also registers the certificate

- Kind: choice
- Basis: Step 3.1. In production, the ACCC issues a client certificate for one software product, and the Register accepts it ([ACCC Developer Portal: Token Endpoint][accc-token]).
- Reason: Two lab actions let a certificate exist that the token endpoint refuses. One command that causes one fact is one event ([ADR 0021][adr-0021]).

### <a id="reg-0644"></a>REG-0644 An issue gives 200 and the certificate chain with no root

- Kind: choice
- Basis: Step 3.2. TLS servers and clients load the PEM chain format of [RFC 8555, section 9.1][rfc8555-chain], and a TLS server sends its chain with no root ([RFC 8446, section 4.4.2][rfc8446-cert]).
- Reason: The certificate and the intermediate are the file that a TLS server or client needs. The Register keeps only thumbprints, so an issue makes no resource that a `GET` can give, and 201 does not apply ([RFC 9110, section 15.3.2][rfc9110-201]).

### <a id="reg-0645"></a>REG-0645 The operator interface gives the CDR CA bundle

- Kind: choice
- Basis: Step 3.3. A participant downloads the ACCC CA bundles from its operator channel ([ACCC Developer Portal: Certificate Management][accc-cert-mgmt]). The CDS gives no API for the bundle.
- Reason: The lab operator uses the operator interface for all participant changes. A TLS listener needs a trust anchor first, so it cannot give the first trust anchor.

### <a id="reg-0646"></a>REG-0646 The CDR CA bundle is the intermediate and then the root

- Kind: choice
- Basis: Step 3.2. [RFC 8555, section 9.1][rfc8555-chain] gives the order of a chain: each certificate before its issuer.
- Reason: The bundle is the same chain with no end-entity certificate. It is the same file that the lab gives to its relying parties.

### <a id="reg-0647"></a>REG-0647 No server certificate for a data recipient

- Kind: choice
- Basis: Step 3.1. The CDS says that data recipients do not need a server certificate now ([CDS: Certificate Management][cds-cert-mgmt]). Each data recipient endpoint is a TLS endpoint with a certificate from a public CA ([CDS: Participant Endpoints][cds-participant-endpoints]), and a data holder calls the [CDR Arrangement Revocation endpoint][cds-arrangement-revocation] with client authentication, not with mTLS.
- Reason: No endpoint of a data recipient uses a CDR server certificate. The Register is not a public CA ([ADR 0022][adr-0022]).

### <a id="reg-0648"></a>REG-0648 No server certificate for a brand with no `dataHolderBrandId`

- Kind: choice
- Basis: Step 3.3. The sources do not say. A brand with no `dataHolderBrandId` shares only product data ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary]), so it has no mTLS endpoint.
- Reason: The identifiers of a brand do not change, so a certificate can never secure a CDR endpoint of the brand. 409 is the code of the other refusals that the state of a record causes ([RFC 9110, section 15.5.10][rfc9110-409]).

### <a id="reg-0649"></a>REG-0649 An exact brand or legal entity name as the O of a server CSR

- Kind: choice
- Basis: Step 3.3. The [CDS: Certificate Signing Request Profile][cds-csr-profile] gives the brand name or the legal entity name as the O, and the subject DN must reflect the organisation ([ACCC CP, section 3.1.2][accc-cp]). The Participant Portal only warns, and the ACCC can refuse the request at its review ([ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide]).
- Reason: The lab has no later review, so the Register refuses at once. An exact comparison is also the rule of a client CSR.

### <a id="reg-0650"></a>REG-0650 No comparison of a server CSR with the endpoints of the brand

- Kind: choice
- Basis: Step 3.3. The data holder decides how it puts its endpoints on domains ([CDS: Certificate Management][cds-cert-mgmt]). The authorisation server endpoints are in the OpenID Provider configuration of the data holder, not on the Register ([CDS: Participant Endpoints][cds-participant-endpoints]). The Participant Portal checks only that the CN is a host name ([ACCC Participant Portal User Guide, section 10.6.2][accc-portal-guide]), and the ACCC does not verify the ownership of a domain name ([ACCC CPS, section 3.1.6][accc-cps]).
- Reason: The Register does not know each mTLS host of a brand. A lab component also needs names that are not in its endpoints, for example `localhost` for access from the host.

### <a id="reg-0651"></a>REG-0651 A server certificate issue is an event, not a registration

- Kind: choice
- Basis: Step 3.3. The ACCC keeps a reference to the server certificates of a brand on the Register ([CDR Support: Brands in the CDR ecosystem][support-brands]). One command that causes one fact is one event ([ADR 0021][adr-0021]).
- Reason: No API of the Register uses a server certificate, so a registration adds nothing. The issue is a change to the data holder, and the `lastUpdated` of a brand is the time of the last change to its data holder. An exception for one event type adds a rule for no gain.

[accc-cert-mgmt]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[accc-brands-summary]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982
[accc-cp]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[accc-portal-guide]: https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf
[support-brands]: https://cdr-support.zendesk.com/hc/en-us/articles/900005881263
[cds-cert-mgmt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-management
[cds-cert-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[rfc5280-bounds]: https://www.rfc-editor.org/rfc/rfc5280#appendix-A
[rfc8446-cert]: https://www.rfc-editor.org/rfc/rfc8446#section-4.4.2
[rfc8555-chain]: https://www.rfc-editor.org/rfc/rfc8555#section-9.1
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#name-409-conflict
[adr-0021]: ../../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0022]: ../../../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
