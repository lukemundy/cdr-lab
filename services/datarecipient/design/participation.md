# Mock data recipient: Participation

This area gives the work of the data recipient as a participant of the CDR: its software product, its public listener and its JWKS, its signing keys and its client certificate, the calls to the Register, the list of banks and their logos, and the registration with each bank. The [CDR client library](../../../pkg/cdrclient/DESIGN.md) gives the protocol of each call.

## Software product

- <a id="dr-0700"></a>**DR-0700** The instance is one software product of its brand on the Register. The setting `SOFTWARE_PRODUCT_ID` gives its software product ID. The setting is required, and it has no default. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: settings with no lab default](deployment.md#dr-0650))
- <a id="dr-0701"></a>**DR-0701** The data recipient gets the base URIs of the Register from the settings `REGISTER_PUBLIC_BASE_URL` and `REGISTER_SECURE_BASE_URL`. Both are required, and they have no default. ([the base URI of the Register at the data holder](../../dataholder/design/registration-management.md#dh-0680); [Judgement: settings with no lab default](deployment.md#dr-0650))
- <a id="dr-0702"></a>**DR-0702** The redirect URI of the software product is the base URI of the web app and the path `/connect/callback`. For example, `https://localhost:8585/connect/callback`. ([CDS: SSA Definition][cds-ssa-definition]; [Judgement: the redirect URI on the web app, and the JWKS on the public listener](#dr-0750))
- <a id="dr-0703"></a>**DR-0703** The JWKS URI of the software product is the base URI of the public listener and the path `/jwks`. The recipient base URI is the base URI of the public listener. ([CDS: SSA Definition][cds-ssa-definition]; [Judgement: the redirect URI on the web app, and the JWKS on the public listener](#dr-0750))

## Public listener

- <a id="dr-0704"></a>**DR-0704** The data recipient has a public listener with TLS and no client authentication. It serves only the JWKS. ([CDS: Use of MTLS][cds-use-of-mtls]; [Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025]; [Judgement: no endpoint for the end of an arrangement before the dashboard](#dr-0751))
- <a id="dr-0705"></a>**DR-0705** The public listener has [the TLS profile of the CDS library](../../../pkg/cds/DESIGN.md#cdsl-0041) and [requests no client certificate](../../../pkg/cds/DESIGN.md#cdsl-0043). ([CDS: Use of MTLS][cds-use-of-mtls])
- <a id="dr-0706"></a>**DR-0706** The server certificate of the public listener is from the public CA of the deployment. It gets its certificate from a certificate file and a key file, and [uses a renewed certificate with no restart](../../../pkg/cds/DESIGN.md#cdsl-0089). ([CDS: Participant Endpoints][cds-participant-endpoints]; [the trust of a data holder for a JWKS](../../dataholder/design/registration.md#dh-0623))
- <a id="dr-0707"></a>**DR-0707** A request for a path other than `/jwks` on the public listener gets 404 with no body. ([Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])
- <a id="dr-0708"></a>**DR-0708** The public listener has no `Host` allowlist and no session cookie. ([CDS: JSON Web Key Set endpoint][cds-jwks-endpoint]; [Judgement: a web app listener on a data recipient (ADR 0025)][adr-0025])

## Signing keys and JWKS

- <a id="dr-0709"></a>**DR-0709** The JWKS gives [the public keys of the signing keys](../../../pkg/cdrclient/DESIGN.md#cdrc-0012), with `Cache-Control: no-cache`. ([CDS: JSON Web Key Set endpoint][cds-jwks-endpoint]; [Judgement: no cache lifetime for the configuration](../../../pkg/authserver/design/configuration.md#auth-0018))
- <a id="dr-0710"></a>**DR-0710** When the data recipient starts with no signing key, it creates one RSA signing key of 2048 bits. A restart keeps the signing keys. ([the size of a signing key](../../../pkg/cdrclient/DESIGN.md#cdrc-0011); [Judgement: a signing key that the data holder creates](../../dataholder/design/authorisation-server.md#dh-0533))
- <a id="dr-0711"></a>**DR-0711** The data recipient signs each client assertion, registration JWT and request object with [the first signing key](../../../pkg/cdrclient/DESIGN.md#cdrc-0010). ([FAPI 1.0 Advanced, section 8.6][fapi-algorithms])

## Client certificate

- <a id="dr-0712"></a>**DR-0712** The data recipient calls the mTLS endpoints of the Register and of each bank with a client certificate that the Register issued for its software product. The settings `CLIENT_TLS_CERT_FILE` and `CLIENT_TLS_KEY_FILE` give the files. ([CDS: Certificate Management][cds-cert-mgmt]; [Judgement: CDR certificates from the Register (ADR 0032)][adr-0032])
- <a id="dr-0713"></a>**DR-0713** The data recipient trusts only the CDR CA bundle of the setting `CDR_CA_BUNDLE_FILE` for an mTLS call, and only the public CAs of the deployment for each other TLS call. ([trust for an mTLS call](../../../pkg/cdrclient/DESIGN.md#cdrc-0001); [trust for a TLS call](../../../pkg/cdrclient/DESIGN.md#cdrc-0002))
- <a id="dr-0714"></a>**DR-0714** The data recipient has no CDR server certificate. ([CDS: Certificate Management][cds-cert-mgmt])

## Register

- <a id="dr-0715"></a>**DR-0715** The data recipient gets the accreditation number of its brand from Get Data Recipients. It is the accreditation number of the legal entity that has the brand ID of the instance. ([CDS: Get Data Recipients][cds-get-drs]; [Judgement: the accreditation number from the Register](#dr-0752))
- <a id="dr-0716"></a>**DR-0716** The data recipient gets the list of banks from [Get Data Holder Brands](../../../pkg/cdrclient/DESIGN.md#cdrc-0033) at start-up, and again each interval of the setting `REGISTER_CACHE_INTERVAL`. The setting is required, and it has no default. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: a cache interval of the Register in the configuration](#dr-0753))
- <a id="dr-0717"></a>**DR-0717** The data recipient gets the accreditation number again at each interval of the list of banks. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: a cache interval of the Register in the configuration](#dr-0753))
- <a id="dr-0718"></a>**DR-0718** When a call to the Register fails, the data recipient keeps the last list of banks and the last accreditation number, and tries again after one minute. ([CDS: Data Recipient Requirements][cds-dr-requirements])
- <a id="dr-0719"></a>**DR-0719** The web app offers only a bank whose brand status is `ACTIVE` and whose industries contain `banking`. ([CDS: Get Data Holder Brands][cds-get-dh-brands]; [Judgement: only active banks](#dr-0754))
- <a id="dr-0720"></a>**DR-0720** The data recipient is not ready until it has a list of banks and an accreditation number. ([the readiness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0002))

## Logos of the banks

- <a id="dr-0721"></a>**DR-0721** For each bank of the list, the data recipient gets the logo from the `logoUri` of the brand, and keeps a copy. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: a copy of each bank logo on the origin of the web app](#dr-0755))
- <a id="dr-0722"></a>**DR-0722** The web app serves each copy from its own origin. The browser gets no URL of a bank for a logo. ([Judgement: a copy of each bank logo on the origin of the web app](#dr-0755))
- <a id="dr-0723"></a>**DR-0723** The data recipient keeps a logo only when it is an SVG or a PNG of 256 KiB or less. When the logo does not obey, the web app shows the initial of the brand name in its place. ([Judgement: a copy of each bank logo on the origin of the web app](#dr-0755))
- <a id="dr-0724"></a>**DR-0724** The response of a logo copy has its media type, `X-Content-Type-Options: nosniff` and `Content-Security-Policy: default-src 'none'; style-src 'unsafe-inline'; sandbox`. ([OWASP: File Upload Cheat Sheet][owasp-file-upload]; [Judgement: a copy of each bank logo on the origin of the web app](#dr-0755))

## Registration with a bank

- <a id="dr-0725"></a>**DR-0725** Before the first authorisation flow at a bank, the data recipient [registers its software product](../../../pkg/cdrclient/DESIGN.md#cdrc-0040) with that bank, with a new SSA. ([CDS: Registration Request using JWT][cds-registration-request]; [CDS: Get Software Statement Assertion][cds-get-ssa])
- <a id="dr-0726"></a>**DR-0726** The data recipient keeps the client ID of each registration, with the brand ID of the bank. It uses the registration for each later flow at that bank. ([CDS: Registration Response][cds-registration-response])
- <a id="dr-0727"></a>**DR-0727** Two flows that start at the same time at a bank with no registration give one registration. ([the one registration for each software product](../../dataholder/design/registration.md#dh-0632))
- <a id="dr-0728"></a>**DR-0728** The registration request gives the redirect URI of the software product as its only redirect URI. ([CDS: Registration Request using JWT][cds-registration-request])
- <a id="dr-0729"></a>**DR-0729** When a registration fails, the web app says that the bank cannot be connected now, and the member can try again later. The data recipient keeps no registration, and the stream of the consent records an authorisation flow with no authorisation. ([CDS: Registration Errors][cds-registration-errors]; [the facts of a consent](consent.md#dr-0831))

## Judgements

### <a id="dr-0750"></a>DR-0750 The redirect URI on the web app, and the JWKS on the public listener

- Kind: choice
- Basis: Step 3.3. The SSA gives the redirect URIs, the JWKS URI and the recipient base URI as separate values ([CDS: SSA Definition][cds-ssa-definition]). The browser follows only the redirect URI. A data holder gets the JWKS, and later sends the notices to the recipient base URI.
- Reason: The redirect URI goes to the web app, which has the session of the member. The JWKS and the notices are on a listener with no cookie and no catch-all route, as ADR 0025 requires, so each URI can be on a host that its client reaches.

### <a id="dr-0751"></a>DR-0751 No endpoint for the end of an arrangement before the dashboard

- Kind: gap
- Basis: Step 1. A software product serves the CDR Arrangement Revocation endpoint at its recipient base URI ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]). The data holder sends no new attempt after a 404 ([the failed notice of the data holder](../../dataholder/design/end-of-an-arrangement.md#dh-0935)).
- Reason: The endpoint, the withdrawal and the dashboard are [non-goals of the data recipient](README.md#non-goals), because they make one function together. Without the endpoint, a collection run finds an arrangement that ended ([a connection that needs attention](data-collection.md#dr-1020)).

### <a id="dr-0752"></a>DR-0752 The accreditation number from the Register

- Kind: choice
- Basis: Step 3.3. The Register gives the accreditation number of each data recipient legal entity ([CDS: Get Data Recipients][cds-get-drs]). The consent page shows it ([CDR Rules][cdr-rules], rule 4.11(3)(b)).
- Reason: One source of the number cannot disagree with the Register. A setting or a field in the brand directory is a second copy that a lab user must keep equal.

### <a id="dr-0753"></a>DR-0753 A cache interval of the Register in the configuration

- Kind: choice
- Basis: Step 1. Data recipients keep a copy of the participant data that they use, and update it with a slow poll. The CDS recommends 6 hours for Get Data Holder Brands and for the other data of Get Data Recipients ([CDS: Participant Statuses][cds-participant-statuses]).
- Reason: A lab user who adds a bank wants to see it in minutes. A setting with no default lets the lab use a short interval and a different deployment use the CDS value.

### <a id="dr-0754"></a>DR-0754 Only active banks

- Kind: choice
- Basis: Step 3.1. Get Data Holder Brands gives the status and the industries of each brand ([CDS: Get Data Holder Brands][cds-get-dh-brands]). A brand that is not active is not a brand that a consumer can connect.
- Reason: A bank that cannot complete a flow gives the member an error after the consent. The web app does not offer it.

### <a id="dr-0755"></a>DR-0755 A copy of each bank logo on the origin of the web app

- Kind: choice
- Basis: Step 3.2. The CDS gives the `logoUri` of each brand and no rule for its use ([CDS: Register API schemas][cds-register-schemas]). In the lab, the logo is on a bank host that a browser on the host cannot reach.
- Reason: A copy shows the logo in each deployment, and keeps the Content Security Policy of the web app at its own origin. An SVG can contain a script, so the copy has a policy that stops each script when a browser opens it directly.

[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[adr-0032]: ../../../docs/adr/0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-cert-mgmt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-management
[cds-dr-requirements]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipient-requirements
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-get-drs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients
[cds-get-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-statement-assertion-ssa
[cds-jwks-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#json-web-key-set-endpoint
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-registration-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-errors
[cds-registration-request]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-request-using-jwt
[cds-registration-response]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-response
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-use-of-mtls]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#use-of-mtls
[fapi-algorithms]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#algorithm-considerations
[owasp-file-upload]: https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html
