# Specification glossary

This glossary gives one term for each concept that two or more specifications use. Specifications use only these terms ([ADR 0024][adr-0024]).

## How to use the glossary

- Each entry gives the term and its source. When the CDS, the CDR legislation or a cited standard defines the term, the entry links to that definition and does not copy it. Otherwise, the entry is a lab term and gives a short definition.
- "Do not use" lists the synonyms that a specification must not use for the concept.
- "Not the same as" names a close concept that has its own entry.
- An entry can give one abbreviation. A specification uses the term or that abbreviation only. Requirement and judgement identifiers (for example `DH-0042`) are not abbreviations, so this rule does not apply to them.
- An entry can give a short form. A specification can use the short form when the file has no other concept with that name.
- The glossary applies to specification prose only. These do not change because of the glossary:
  - identifiers in code, API paths, field names, enumeration values and event names, for example `dataHolderBrandId` or `ALTERNATIVE_METHOD`
  - the text of a quote from a source, and the title of a link
  - UI text that the CDS consumer experience standards give, which keeps its CDS wording
- When a specification needs a shared concept that is not in the glossary, the same change adds an entry.

## Participants

### CDR consumer

- Source: [Competition and Consumer Act 2010][cdr-act], section 56AI.
- Short form: consumer.
- Not the same as: [customer](#customer), [user](#user).
- Do not use: user (for the consumer).

### CDR participant

- Source: [Competition and Consumer Act 2010][cdr-act], section 56AL: a data holder or an accredited data recipient.
- Not the same as: [participant](#participant). Use "CDR participant" only for the legal meaning.

### CDR Register

- Source: [CDS: CDR Federation][cds-federation], [CDS: Register APIs][cds-register-apis].
- Short form: Register. "Mock Register" and "production Register" name the implementation.

### Customer

- Lab term: the record of a person at a data holder, with a customer number and accounts. The CDS uses the word in [Get Customer][cds-get-customer] and [User Identifiers][cds-user-identifiers].
- Not the same as: [CDR consumer](#cdr-consumer). A consumer acts. A customer is the record of the data holder about that consumer.

### Data holder

- Source: [Competition and Consumer Act 2010][cdr-act], section 56AJ; [CDR Rules][cdr-rules], rule 1.7.
- Do not use: DH, bank (for the data holder).

### Data holder brand

- Source: [CDS: Register API schemas][cds-register-schemas], [CDS: Multi-Brand Support][cds-multi-brand].
- Short form: brand.
- Do not use: bank, the bank. A page that names the bank shows the brand name.

### Data holder brand ID

- Source: [CDS: Register API schemas][cds-register-schemas] (`dataHolderBrandId`).
- Short form: brand ID.

### Data recipient

- Source: the CDS name for an accredited data recipient ([Competition and Consumer Act 2010][cdr-act], section 56AK; [CDS: CDR Federation][cds-federation]).
- Not the same as: [software product](#software-product), [client](#client).
- Do not use: ADR. Use "accredited person" only where a CDR rule names it, for example on the authorisation page ([CDR Rules][cdr-rules], rule 4.23(1)(a)).

### Data recipient brand

- Source: [CDS: Register API schemas][cds-register-schemas].

### Lab operator

- Lab term: the person who uses the operator interfaces and operator consoles. In the data holder, the lab operator acts as the staff of the bank. In the Register, the lab operator acts as the ACCC.
- Short form: operator.
- Not the same as: [lab user](#lab-user).
- Do not use: staff, admin.

### Lab user

- Lab term: a person who runs the lab and its tests, for example an engineer who tests a data recipient.
- Not the same as: [lab operator](#lab-operator), which is a role in the lab.

### Legal entity

- Source: [CDS: Register API schemas][cds-register-schemas]. A data holder legal entity or a data recipient legal entity.

### Participant

- Source: [CDS: CDR Federation][cds-federation], [CDS: Participant Statuses][cds-participant-statuses]: a record on the Register, which is a legal entity, a brand or a software product.
- Not the same as: [CDR participant](#cdr-participant).

### Participant status

- Source: [CDS: Participant Statuses][cds-participant-statuses].
- Do not use: Register status.

### Software product

- Source: [CDS: Register API schemas][cds-register-schemas], [CDS: Software Statement Assertion][cds-ssa].
- Not the same as: [data recipient](#data-recipient), [client](#client).

### Software product ID

- Source: [CDS: Register API schemas][cds-register-schemas] (`softwareProductId`), [CDS: SSA Definition][cds-ssa-definition] (`software_id`).
- Do not use: software ID.

### User

- Lab term: a person who uses a lab UI in a browser, that is a [web UI](#web-ui) or the [consumer pages](#consumer-pages). Use it only in the specifications of the design system and the web platform package, where a rule applies to each UI that uses them.
- Not the same as: [CDR consumer](#cdr-consumer), [lab operator](#lab-operator). In a service specification, name the role.

## Consent and authorisation

The data holder asks for an authorisation. The data recipient asks for a consent. ([CDR Rules][cdr-rules], rule 1.7 "authorisation" and "consent", rule 1.10A, Division 4.4)

### Authorisation

- Source: [CDR Rules][cdr-rules], rule 1.7 ("authorisation" to disclose CDR data) and Division 4.4.
- Not the same as: [consent](#consent), [CDR arrangement](#cdr-arrangement). An authorisation is given to a data holder. A consent is given to an accredited person.
- Do not use: consent, CDR consent (for the authorisation that a data holder gets).
- The OAuth 2.0 terms "authorisation request", "authorisation code", "authorisation endpoint" and "authorisation server" are different concepts. They keep their names ([CDS: Security Endpoints][cds-security-endpoints]).

### Authorisation flow

- Lab term: the sequence of a pushed authorisation request, the consumer authentication, the authorisation page and the authorisation response, in one browser ([CDS: Authorisation Standards][cds-authorisation-standards]).
- Do not use: consent flow.

### CDR arrangement

- Source: [CDS: CDR Arrangement ID][cds-arrangement-id].
- Short form: arrangement.
- Not the same as: [authorisation](#authorisation). The arrangement is the record and the identifier (`cdr_arrangement_id`) that the CDS gives to an authorisation and the consent that it serves.
- Do not use: sharing arrangement.

### Consent

- Source: [CDR Rules][cdr-rules], rule 1.7 and rule 1.10A.
- Not the same as: [authorisation](#authorisation). Use "consent" only for the data recipient side.

### Consumer dashboard

- Source: [CDR Rules][cdr-rules], rule 1.7 and rule 1.15 (data holder); [CDS: Data Holder Dashboards][cds-dashboard-standards].
- Short form: dashboard.

### Data cluster

- Source: [CDS: Data Language Standards][cds-data-language].

### Revoke

- Source: [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation] (an arrangement), [RFC 7009][rfc7009] (a token), [RFC 5280, section 5.3.1][rfc5280-reason] (a certificate), [CDS: Participant Statuses][cds-participant-statuses] (the status `REVOKED`).
- Not the same as: [withdraw](#withdraw). A software product, a client or the Register revokes. A consumer withdraws.

### Withdraw

- Source: [CDR Rules][cdr-rules], rule 4.25 (an authorisation); [CDS: Withdrawal Standards][cds-withdrawal-standards].
- Not the same as: [revoke](#revoke).
- Do not use: revoke (for the action of the consumer).

## Authentication and APIs

### Access token

- Source: [CDS: Access Token][cds-access-token].

### Authentication

- Source: [CDS: Authentication standards][cds-cx-authentication]: the authentication of the consumer in the authorisation flow.
- Not the same as: [sign-in](#sign-in).
- Do not use: login, log in. "Login CSRF" is the name of an attack, and it stays.

### Authorisation server

- Source: [CDS: Security Endpoints][cds-security-endpoints].

### Base URI

- Source: [CDS: Participant Endpoints][cds-participant-endpoints].
- Do not use: base URL.

### Client

- Source: [RFC 6749, section 2][rfc6749-client], [RFC 7591][rfc7591-metadata]: the OAuth 2.0 client of a registration, with the client ID that the data holder issues.
- Not the same as: [software product](#software-product). One software product has one client at each data holder where it has a registration.

### Holder of key

- Source: [CDS: Holder of Key Mechanism][cds-hok], [RFC 8705, section 3][rfc8705-binding].

### Issuer

- Source: [CDS: OpenID Provider Configuration End Point][cds-oidc-config], [CDS: Multi-Brand Support][cds-multi-brand].

### JWKS

- Source: [CDS: JSON Web Key Set endpoint][cds-jwks-endpoint], [RFC 7517, section 5][rfc7517-jwks].
- Not the same as: [key set](#key-set).
- Do not use: JWK set.

### One-time password

- Source: [CDS: One-Time Password Credential Requirements][cds-otp].
- Abbreviation: OTP.

### OpenID Provider configuration

- Source: [CDS: OpenID Provider Configuration End Point][cds-oidc-config].

### Problem details

- Source: [RFC 9457][rfc9457]. The error format of the lab-only interfaces. The CDS APIs use the CDS error format ([CDS: Error Codes][cds-error-codes]).

### Refresh token

- Source: [CDS: Refresh Token][cds-refresh-token].

### Registration

- Source: [CDS: Dynamic Client Registration APIs][cds-dcr-apis]: the registration of a software product at a data holder.
- Do not use: client registration, DCR.

### Software statement assertion

- Source: [CDS: Software Statement Assertion][cds-ssa].
- Abbreviation: SSA.

## PKI

### CA bundle

- Source: [ACCC: Certificate Management][accc-cert-mgmt]. The CDR CA bundle has the root and the intermediate of the CDR CA.
- Not the same as: a trust anchor, which is the general [RFC 5280][rfc5280] term.

### CDR CA

- Source: [CDS: Certificate Trust Model][cds-trust-model]. In the lab, the mock CDR CA of the mock CDR PKI ([PKI DESIGN.md][pki-design]).
- Not the same as: [public CA](#public-ca).

### CDR certificate

- Lab term: a client certificate or a server certificate that the CDR CA issues ([CDS: Certificate Trust Model][cds-trust-model]).

### Client certificate

- Source: [CDS: Certificate Trust Model][cds-trust-model].

### CRL

- Source: [RFC 5280, section 5][rfc5280-crl]: a certificate revocation list.
- Abbreviation: CRL.

### CSR

- Source: [CDS: Certificate Signing Request Profile][cds-csr-profile]: a certificate signing request.
- Abbreviation: CSR.

### Public CA

- Source: [CDS: Participant Endpoints][cds-participant-endpoints]: the CA of the certificates of the public endpoints. In the lab, the mock public CA ([PKI DESIGN.md][pki-design]).
- Not the same as: [CDR CA](#cdr-ca).

### Server certificate

- Source: [CDS: Certificate Trust Model][cds-trust-model]: a CDR certificate for an mTLS endpoint. A "public TLS server certificate" is from the public CA.

## Lab

### Conformance suite

- Lab term: the black-box tests in `conformance/` ([conformance DESIGN.md][conformance-design]). One test of the suite is a conformance test.
- Not the same as: the OpenID conformance suite, which is a different tool.

### Console API

- Lab term: the HTTP API of an operator console, for the console only.

### Consumer pages

- Lab term: the authentication and authorisation pages of a data holder.
- Not the same as: [consumer UI](#consumer-ui).

### Consumer UI

- Lab term: a lab UI for a consumer in a browser: the consumer pages or the online bank site ([design system][ui-design]).

### Design system

- Lab term: the tokens, components and rules of the lab UIs in a browser ([design system][ui-design]).

### Health probes

- Lab term: the endpoints of a service for its liveness and readiness, on its [internal listener](#internal-listener).

### Image commit

- Lab term: for one service and one commit of `main`, the first commit with an image of that service, in the first-parent history of `main` from that commit back ([ADR 0029][adr-0029]).

### Internal listener

- Lab term: the plain HTTP listener of a service for the deployment and the lab operator ([ADR 0025][adr-0025]). The health probes, the operator interface and the operator console are on it. On the Register, the CRL is also on it.
- Not the same as: [public listener](#public-listener), which participants and consumers use.

### Key set

- Lab term: a record of the Register that holds the signing keys for one use (`access-token` or `ssa`), with a current key and an optional next key.
- Not the same as: [JWKS](#jwks), which is the endpoint that publishes the public keys.

### Lab

- Lab term: the lab deployment (Docker Compose) and its data ([README: Lab deployment][readme-lab]). "Lab-only" marks a function that production does not have.

### Lab brand

- Lab term: the neutral brand of the design system, for each role that the brand of a participant does not give ([the lab brand](../../web/packages/ui/design/tokens-and-brands.md#ui-0033)).

### Load file

- Lab term: the JSON format of the participant records that the Register loads, or of the products that the data holder loads.
- Not the same as: [seed file](#seed-file).

### Mock CDR PKI

- Lab term: the CAs and the certificate profiles of the lab ([PKI DESIGN.md][pki-design]).

### Mock component, real component

- Lab term: [README: Component types][readme-component-types].

### mTLS listener

- Lab term: the listener of a service that requires a client certificate. The secure APIs and the authorisation server endpoints that need mTLS are on it.
- Abbreviation: mTLS ([RFC 8705][rfc8705-binding]).

### Online bank site

- Lab term: the lab-only web site of a data holder brand, where a consumer signs in, sees accounts and uses the consumer dashboard.
- Short form: site.
- Do not use: bank site.

### Operator console

- Lab term: the lab-only browser UI of a service for the lab operator. It is a [web UI](#web-ui).
- Short form: console.

### Operator interface

- Lab term: the lab-only HTTP API of a service for the lab operator.
- Do not use: admin API.

### Operator token

- Lab term: the bearer token of the operator interfaces.
- Not the same as: an [access token](#access-token), which is also a bearer token ([RFC 6750][rfc6750]).

### Preset brand

- Lab term: a brand directory that the data holder contains ([brands README][brands-readme]).
- Do not use: preset bank.

### Public listener

- Lab term: the TLS listener of a service with no client authentication. The public APIs, the OpenID Provider configuration and the consumer pages are on it.
- Do not use: TLS listener.

### Release

- Lab term: the artifacts of one commit of `main` with one version, `vX.Y.Z`, and the GitHub Release with their notes ([ADR 0027][adr-0027]).
- Not the same as: a publication from `main`, which has commit versions.

### Seed file

- Lab term: a load file that the lab loads before the services start ([README: Data][readme-data]).
- Not the same as: a fixture, which is a certificate that the set-up of the lab issues for a test.

### Sign-in

- Lab term: the sign-in of a customer on the online bank site.
- Not the same as: [authentication](#authentication). The data holder keeps them separate.
- Do not use: login, log in.

### Web UI

- Lab term: a lab-only browser UI and its API on one listener, with one origin ([web UI library DESIGN.md][consolehttp-design]). The web UIs are the operator consoles, with their console APIs, and the online bank site, with its API.
- Not the same as: [operator console](#operator-console) and [online bank site](#online-bank-site), which are each one web UI.
- Not the same as: [consumer pages](#consumer-pages), which are on the public listener.

## Data

### CDR data

- Source: [Competition and Consumer Act 2010][cdr-act], section 56AI.

### Personal information

- Source: [Privacy Act 1988][privacy-act], section 6. Telemetry never contains it ([ADR 0006][adr-0006]).

[adr-0006]: ../adr/0006-telemetry.md
[adr-0024]: ../adr/0024-write-specifications-as-identified-single-requirements-split-by-area.md
[adr-0025]: ../adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[adr-0027]: ../adr/0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md
[adr-0029]: ../adr/0029-promote-the-tested-image-of-each-service-to-a-release-with-no-rebuild.md
[brands-readme]: ../../services/dataholder/brands/README.md
[conformance-design]: ../../conformance/DESIGN.md
[consolehttp-design]: ../../pkg/platform/consolehttp/DESIGN.md
[pki-design]: ../../pki/DESIGN.md
[readme-component-types]: ../../README.md#component-types
[readme-data]: ../../README.md#data
[readme-lab]: ../../README.md#lab-deployment
[ui-design]: ../../web/packages/ui/design/README.md
[accc-cert-mgmt]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management
[cdr-act]: https://www.legislation.gov.au/C2004A00109/latest/text
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[privacy-act]: https://www.legislation.gov.au/C2004A03712/latest/text
[cds-access-token]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#access-token
[cds-arrangement-id]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-id
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-authorisation-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-standards
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[cds-dashboard-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_dashboard-standards_data-holder-dashboards
[cds-data-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-language-standards-common
[cds-dcr-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dcr-apis
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[cds-federation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-federation
[cds-get-customer]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-common-api_get-customer
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[cds-jwks-endpoint]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#json-web-key-set-endpoint
[cds-multi-brand]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#multi-brand-support-separate-issuers-for-data-holder-brands
[cds-oidc-config]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#openid-provider-configuration-endpoint
[cds-otp]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#one-time-password-credential-requirements
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-refresh-token]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#refresh-token
[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-apis
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-security-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#security-endpoints
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-user-identifiers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#user-identifiers
[cds-withdrawal-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#withdrawal-standards
[rfc5280]: https://www.rfc-editor.org/rfc/rfc5280
[rfc5280-crl]: https://www.rfc-editor.org/rfc/rfc5280#section-5
[rfc5280-reason]: https://www.rfc-editor.org/rfc/rfc5280#section-5.3.1
[rfc6749-client]: https://www.rfc-editor.org/rfc/rfc6749#section-2
[rfc6750]: https://www.rfc-editor.org/rfc/rfc6750
[rfc7009]: https://www.rfc-editor.org/rfc/rfc7009
[rfc7517-jwks]: https://www.rfc-editor.org/rfc/rfc7517#section-5
[rfc7591-metadata]: https://www.rfc-editor.org/rfc/rfc7591#section-2
[rfc8705-binding]: https://www.rfc-editor.org/rfc/rfc8705#section-3
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
