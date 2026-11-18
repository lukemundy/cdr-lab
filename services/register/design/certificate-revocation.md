# Mock CDR Register: Certificate revocation

This area gives the revocation of a CDR certificate, the CRL that the Register signs and serves, and the revocation check of the client certificates of the Register. In a running lab, the Register [is the only component that revokes a CDR certificate](../../../pki/DESIGN.md#pki-0048) and [the only component that signs a CRL of the intermediate](../../../pki/DESIGN.md#pki-0049). The [PKI specification](../../../pki/DESIGN.md) gives the other rules of each revocation and each CRL, and the [CA library](../../../pkg/ca/DESIGN.md) gives the revocation reasons and the CRL format.

## Revocation request

- <a id="reg-0700"></a>**REG-0700** The lab operator can revoke a certificate of the mock CDR intermediate through the operator interface. ([ACCC CP, section 4.9.3][accc-cp]; [Judgement: the Register is the CDR intermediate CA (ADR 0022)][adr-0022])
- <a id="reg-0701"></a>**REG-0701** The lab operator can also revoke a certificate that the lab set-up issued, which the Register does not know. ([Judgement: the certificate identifies the certificate to revoke](#reg-0740))
- <a id="reg-0702"></a>**REG-0702** The request body of a revocation is a JSON object with these two fields. Each field is required. ([ACCC CP, section 4.9.3][accc-cp]; [Judgement: a revocation is a JSON `POST` on a path with no participant](#reg-0741); [Judgement: the certificate identifies the certificate to revoke](#reg-0740))
  - `certificate`: one PEM certificate, or a PEM chain with the certificate first, as an issue gives it.
  - `reason`: one of [the revocation reasons of a CDR certificate](../../../pkg/ca/DESIGN.md#ca-0041), for example `keyCompromise`.
- <a id="reg-0703"></a>**REG-0703** When the `certificate` is a PEM chain, the Register revokes the first certificate. ([Judgement: the Register revokes the first certificate of a PEM chain](#reg-0743))
- <a id="reg-0704"></a>**REG-0704** When the request body breaks one of these rules, the response is 400, and the `detail` names each field that fails. ([RFC 9110, section 15.5.1][rfc9110-400])
  - The body is a JSON object with the two fields.
  - The `certificate` starts with a PEM certificate.
  - Each PEM block of the `certificate` is a certificate.
  - The `reason` is one of the revocation reasons.
- <a id="reg-0705"></a>**REG-0705** When the mock CDR intermediate did not sign the certificate, the response is 400. ([RFC 5280, section 5][rfc5280-crl]; [Judgement: 400 for a certificate of a different issuer](#reg-0745))
- <a id="reg-0706"></a>**REG-0706** When [the certificate is already revoked](../../../pki/DESIGN.md#pki-0052), the response is 409, and the Register changes nothing, also for a different reason. The `detail` gives the reason of the revocation that exists. ([Judgement: 409 for a second revocation](#reg-0744))
- <a id="reg-0707"></a>**REG-0707** The Register does not check the validity period of the certificate. For example, the lab operator can revoke an expired certificate. ([Judgement: no check of the validity period or the participant](#reg-0746))
- <a id="reg-0708"></a>**REG-0708** The Register does not compare the certificate with the participants. ([Judgement: no check of the validity period or the participant](#reg-0746))
- <a id="reg-0709"></a>**REG-0709** A revocation does not remove a registered client certificate of a software product. ([Judgement: a revocation and a registration are separate](#reg-0747))
- <a id="reg-0710"></a>**REG-0710** A revocation is not an event of a participant, and the history of a participant does not show it. ([Judgement: a revocation is a record of the CA](#reg-0748))

## CRL

- <a id="reg-0711"></a>**REG-0711** When the Register starts, it signs a new CRL. ([Judgement: a new CRL at each start](#reg-0749))
- <a id="reg-0712"></a>**REG-0712** The Register signs a new CRL no later than 24 hours after the `thisUpdate` of the current CRL, also after a suspend of the host. ([a new CRL at least once every 24 hours](../../../pki/DESIGN.md#pki-0061); [Judgement: a new CRL 24 hours after the `thisUpdate` of the current CRL](#reg-0750))
- <a id="reg-0713"></a>**REG-0713** The Register serves only its newest CRL. ([ACCC CP, section 4.9.7][accc-cp])
- <a id="reg-0714"></a>**REG-0714** A request never gets a CRL that is only partly written. ([ACCC CPS, section 4.10.2][accc-cps])
- <a id="reg-1505"></a>**REG-1505** The Register serves the CRL on [the internal listener](deployment.md#reg-1499). ([RFC 5280, section 4.2.1.13][rfc5280-crldp]; [Judgement: the CRL on the internal listener, with no authentication](#reg-0751))
- <a id="reg-0716"></a>**REG-0716** The CRL path needs no authentication. ([Judgement: the CRL on the internal listener, with no authentication](#reg-0751))
- <a id="reg-0717"></a>**REG-0717** The CRL path does not check the `Host` header. For example, a request with the IP address of the pod as its `Host` gets the CRL. ([Judgement: the CRL on the internal listener, with no authentication](#reg-0751))
- <a id="reg-0718"></a>**REG-0718** `GET /crl/cdr-intermediate.crl` gives 200 with [the current CRL in DER](../../../pkg/ca/DESIGN.md#ca-0044) and [the content type `application/pkix-crl`](../../../pki/DESIGN.md#pki-0064). ([RFC 2585, section 4.2][rfc2585-mime])
- <a id="reg-0719"></a>**REG-0719** `HEAD /crl/cdr-intermediate.crl` gives the headers of the `GET` response, with no body. ([RFC 9110, section 9.1][rfc9110-methods])
- <a id="reg-0720"></a>**REG-0720** A different method on the CRL path gives 405 with an `Allow` header. ([RFC 9110, section 15.5.6][rfc9110-405])
- <a id="reg-0721"></a>**REG-0721** A different path under `/crl/` gives 404. ([RFC 9110, section 15.5.5][rfc9110-404])
- <a id="reg-0722"></a>**REG-0722** When the Register has no current CRL, the CRL path gives 503. ([RFC 9110, section 15.6.4][rfc9110-503]; [Judgement: 503 with no current CRL](#reg-0754))
- <a id="reg-0723"></a>**REG-0723** The URL of the CRL is a required setting of the Register, with no default. ([the CRL distribution point of a CDR certificate](../../../pkg/ca/DESIGN.md#ca-0034))
- <a id="reg-0724"></a>**REG-0724** Each certificate that the Register issues gives the URL of this setting as its CRL distribution point. ([the CRL distribution point of a CDR certificate](../../../pkg/ca/DESIGN.md#ca-0034))
- <a id="reg-0725"></a>**REG-0725** When the lab operator revokes a certificate, the next TLS handshake with that certificate at the Register fails. ([Judgement: the newest CRL at each connection](#reg-0752))

## Revocation check

- <a id="reg-0726"></a>**REG-0726** The Register checks the revocation status of each client certificate on its mTLS listeners with its newest CRL of the mock CDR intermediate. ([a revocation check for an mTLS listener](../../../pkg/cds/DESIGN.md#cdsl-0046); [a failed revocation check refuses the certificate](../../../pkg/cds/DESIGN.md#cdsl-0047); [CDS: Certificate Validation][cds-cert-validation]; [Judgement: the newest CRL at each connection](#reg-0752))
- <a id="reg-0727"></a>**REG-0727** When the mock CDR intermediate did not sign the client certificate, the revocation check fails. ([CDS: Certificate Validation][cds-cert-validation]; [Judgement: the CRL of the intermediate covers each CDR client certificate](#reg-0753))
- <a id="reg-0728"></a>**REG-0728** When the newest CRL lists the client certificate, the revocation check fails. ([CDS: Certificate Validation][cds-cert-validation])
- <a id="reg-0729"></a>**REG-0729** When the Register has no current CRL, the revocation check fails. ([Judgement: refuse a certificate with a revocation status that is not known](../../../pkg/cds/DESIGN.md#cdsl-0067))

## Operator interface paths

| ID | Action | Method and path | Success | Source |
|---|---|---|---|---|
| <a id="reg-0731"></a>**REG-0731** | Revoke a certificate of the mock CDR intermediate | `POST /v1/cdr-certificates/revoke` | 204 | [Judgement: a revocation is a JSON `POST` on a path with no participant](#reg-0741); [Judgement: a revocation gives 204 with no content](#reg-0742) |

## Judgements

### <a id="reg-0740"></a>REG-0740 The certificate identifies the certificate to revoke

- Kind: choice
- Basis: Step 3.3. A request must identify the certificate to revoke ([ACCC CP, section 4.9.3][accc-cp]). The Participant Portal revokes a certificate from its list of the valid certificates of the participant ([ACCC Participant Portal User Guide, sections 10.6.4 and 10.7.3][accc-portal-guide]).
- Reason: The Register keeps only thumbprints, and it has no record of the certificates of the lab set-up, so it cannot get a certificate from a thumbprint. A serial number alone does not let the Register check that the intermediate signed the certificate. The certificate gives the serial number and a signature that the Register can check.

### <a id="reg-0741"></a>REG-0741 A revocation is a JSON `POST` on a path with no participant

- Kind: choice
- Basis: Step 3.3. The CDS defines no interface to revoke a certificate.
- Reason: The request has two values, and the operator interface already takes JSON bodies. An action is a `POST` on a path that ends with the action, as each status action is. The path is not under a participant, because a certificate of the lab set-up has no participant in the Register.

### <a id="reg-0742"></a>REG-0742 A revocation gives 204 with no content

- Kind: choice
- Basis: Step 3.2. A response with no content is 204 ([RFC 9110, section 15.3.5][rfc9110-204]).
- Reason: A revocation makes no resource that a `GET` gives. [The CRL lists the certificate with its reason](../../../pki/DESIGN.md#pki-0053), so the CRL gives the result.

### <a id="reg-0743"></a>REG-0743 The Register revokes the first certificate of a PEM chain

- Kind: choice
- Basis: Step 3.3. An issue gives the certificate and then the intermediate ([RFC 8555, section 9.1][rfc8555-chain]).
- Reason: The lab operator keeps that file, and can give it as it is. In a chain in the wrong order, the first certificate is the intermediate, which the intermediate did not sign, so the Register refuses it.

### <a id="reg-0744"></a>REG-0744 409 for a second revocation

- Kind: choice
- Basis: Step 3.3. The PKI permits [one revocation for each certificate](../../../pki/DESIGN.md#pki-0085). 409 is the code of a refusal that the state of a resource causes ([RFC 9110, section 15.5.10][rfc9110-409]).
- Reason: A second revocation that gives 204 hides that the reason did not change. The `detail` tells the lab operator the reason that applies.

### <a id="reg-0745"></a>REG-0745 400 for a certificate of a different issuer

- Kind: choice
- Basis: Step 3.2. The CRL of the intermediate can list only the certificates of the intermediate ([RFC 5280, section 5][rfc5280-crl]).
- Reason: The fault is in the request body, as for a CSR that breaks the CSR profile.

### <a id="reg-0746"></a>REG-0746 No check of the validity period or the participant

- Kind: choice
- Basis: Step 3.3. The sources do not say. The Register revokes each certificate of the intermediate ([ADR 0022][adr-0022]), also each certificate of the lab set-up.
- Reason: A relying party refuses an expired certificate also with no CRL entry, so its revocation does no harm. A certificate of the lab set-up has no participant in the Register, so a participant check refuses it.

### <a id="reg-0747"></a>REG-0747 A revocation and a registration are separate

- Kind: choice
- Basis: Step 3.3. One command that causes one fact is one event ([ADR 0021][adr-0021]).
- Reason: The mTLS listeners refuse a revoked certificate in the TLS handshake, before the token endpoint, so its registration gives no access. A revocation is a fact of the CA, and a certificate of the lab set-up has no software product. The lab operator removes the registration with its own action.

### <a id="reg-0748"></a>REG-0748 A revocation is a record of the CA

- Kind: choice
- Basis: Step 3.3. [ADR 0003][adr-0003] gives two aggregate types, a data recipient and a data holder, and a revocation names neither.
- Reason: The key sets are records of the Register that are not event sourced, and the revocations are the same. The CRL gives each revocation with its time and its reason.

### <a id="reg-0749"></a>REG-0749 A new CRL at each start

- Kind: choice
- Basis: Step 3.3. The ACCC issues a CRL at least once every 24 hours, also with no change ([ACCC CP, section 4.9.7, Table 1][accc-cp]). The sources give no rule for a start.
- Reason: A CRL at each start makes sure that the current CRL has the signature of the intermediate key that the Register reads. With [a next update 7 days after the issue](../../../pkg/ca/DESIGN.md#ca-0047), the CRL stays current for 6 days after the Register stops.

### <a id="reg-0750"></a>REG-0750 A new CRL 24 hours after the `thisUpdate` of the current CRL

- Kind: choice
- Basis: Step 3.3. The [ACCC CP, section 4.9.7, Table 1][accc-cp] gives a CRL at least once every 24 hours, with no start time. A background job follows [ADR 0010][adr-0010].
- Reason: A periodic job compares the time of day with the `thisUpdate` of the newest CRL, so its first run after a suspend of the host signs a new CRL. A timer of 24 hours does not count the time of a suspend.

### <a id="reg-0751"></a>REG-0751 The CRL on the internal listener, with no authentication

- Kind: choice
- Basis: Step 3.2. Relying parties get the CRL from the HTTP URL in the certificate ([RFC 5280, section 4.2.1.13][rfc5280-crldp]), and [a CDR certificate gives one CRL distribution point](../../../pkg/ca/DESIGN.md#ca-0034) with an `http` URL. A CRL is signed, so it needs no TLS.
- Reason: A TLS listener needs a revocation check of its own certificate. The CRL is public data, so the operator token and the `Host` allowlist do not apply, and a DNS rebinding attack gets only public data. [ADR 0025][adr-0025] puts the CRL on the internal listener, and a deployment exposes only the CRL path of that listener.

### <a id="reg-0752"></a>REG-0752 The newest CRL at each connection

- Kind: choice
- Basis: Step 2. The CDS gives no time for a revocation to take effect ([CDS: Certificate Validation][cds-cert-validation]). The relying party decides how frequently it gets new revocation data ([ACCC CP, section 4.9.6][accc-cp]). Other relying parties keep a CRL for [a maximum CRL age of 1 minute](../../../pkg/cds/DESIGN.md#cdsl-0068).
- Reason: The Register signs the CRL, so it reads its newest CRL with no download and no wait. A lab test then sees a revocation at the Register at once.

### <a id="reg-0753"></a>REG-0753 The CRL of the intermediate covers each CDR client certificate

- Kind: choice
- Basis: Step 3.3. The CDR CA has [one root CA and one issuing intermediate CA](../../../pkg/ca/DESIGN.md#ca-0001), and [the intermediate has a path length of 0](../../../pkg/ca/DESIGN.md#ca-0004).
- Reason: The mock CDR root signs only the intermediate. Thus the CRL of the intermediate gives the status of each client certificate that chains to the CDR CA bundle, and the Register has no CRL for a different issuer.

### <a id="reg-0754"></a>REG-0754 503 with no current CRL

- Kind: choice
- Basis: Step 3.2. 503 tells a client that the server cannot give the resource for a time ([RFC 9110, section 15.6.4][rfc9110-503]).
- Reason: The Register then also [fails its readiness probe](non-functional-requirements.md#reg-0009). A relying party then has no current CRL, and [refuses the certificate](../../../pkg/cds/DESIGN.md#cdsl-0051).

[accc-cp]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[accc-portal-guide]: https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf
[cds-cert-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation
[rfc2585-mime]: https://www.rfc-editor.org/rfc/rfc2585#section-4.2
[rfc5280-crl]: https://www.rfc-editor.org/rfc/rfc5280#section-5
[rfc5280-crldp]: https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.13
[rfc8555-chain]: https://www.rfc-editor.org/rfc/rfc8555#section-9.1
[rfc9110-methods]: https://www.rfc-editor.org/rfc/rfc9110#section-9.1
[rfc9110-204]: https://www.rfc-editor.org/rfc/rfc9110#name-204-no-content
[rfc9110-400]: https://www.rfc-editor.org/rfc/rfc9110#name-400-bad-request
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#name-404-not-found
[rfc9110-405]: https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#name-409-conflict
[rfc9110-503]: https://www.rfc-editor.org/rfc/rfc9110#name-503-service-unavailable
[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[adr-0010]: ../../../docs/adr/0010-background-jobs-with-river.md
[adr-0021]: ../../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0022]: ../../../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
