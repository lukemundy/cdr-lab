# Mock CDR Register: Operator console

This area gives the lab-only operator console of the Register: its views, the history of a participant, its presentation, its listener and security, and its non-functional requirements. The [actions](operator-console-actions.md) and the [forms](operator-console-forms.md) of the console are separate areas. The [console API](operator-console-api.md) gives the data of the views.

## Scope

- <a id="reg-1000"></a>**REG-1000** The operator console is a web UI for the lab operator. It shows the participants, their statuses, the signing keys and the CDR certificates. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="reg-1001"></a>**REG-1001** The operator console is a lab-only feature. A real component does not depend on it. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])

## Lists

- <a id="reg-1002"></a>**REG-1002** The list of the data recipients gives, for each legal entity, the name, the accreditation number, the status, the number of brands and the number of software products. ([Judgement: the counts in the data recipient list](operator-console-api.md#reg-1175))
- <a id="reg-1003"></a>**REG-1003** The lab operator can filter the list of the data recipients by status, and search it by the name of the legal entity. ([Judgement: the search of a participant list](operator-console-api.md#reg-1173))
- <a id="reg-1004"></a>**REG-1004** The list of the data holders gives, for each legal entity, the name, the legal entity ID, the status, the number of brands and the industries of its brands. ([Judgement: the brands in the data holder list](operator-console-api.md#reg-1177))
- <a id="reg-1005"></a>**REG-1005** The lab operator can filter the list of the data holders by status, and search it by the name of the legal entity or of a brand. For example, a search for a brand name finds its data holder. ([Judgement: the search of the data holder list](operator-console-api.md#reg-1176))
- <a id="reg-1006"></a>**REG-1006** Each row of a list links to the detail of its record. For example, a row of the list of the data recipients links to the detail of that data recipient. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])

## Details

- <a id="reg-1007"></a>**REG-1007** The detail of a data recipient gives the legal entity: its ID, name, accreditation number, accreditation level, logo URI, status and the time of the last change. ([the accreditation of a data recipient legal entity](participants.md#reg-0211))
- <a id="reg-1008"></a>**REG-1008** The detail of a data recipient gives each brand, with its ID, name, logo URI and status. ([the data recipient brand](participants.md#reg-0203))
- <a id="reg-1009"></a>**REG-1009** The detail of a data recipient gives each software product, with its ID, its status, its metadata and the thumbprints of its registered client certificates. ([the metadata of a software product](participants.md#reg-0212); [the client certificates of a software product](participants.md#reg-0213))
- <a id="reg-1010"></a>**REG-1010** The detail of a data holder gives the legal entity: its ID, name, logo URI, status and the time of the last change. ([the data holder legal entity](participants.md#reg-0200))
- <a id="reg-1011"></a>**REG-1011** The detail of a data holder gives each brand, with its brand ID, its interim ID or both, its name, brand group, logo URI, status, industries, endpoints and JWKS endpoint. ([the endpoints of a data holder brand](participants.md#reg-0207); [the identifiers of a data holder brand](participants.md#reg-0208))
- <a id="reg-1012"></a>**REG-1012** The detail of a data holder brand does not show its server certificates. The history shows each issue of a server certificate with its thumbprint. ([the event of a server certificate issue](certificate-management.md#reg-0630); [Judgement: no server certificates in the detail of a data holder brand](#reg-1080))
- <a id="reg-1013"></a>**REG-1013** The detail of a participant links to its history. ([Judgement: the detail and the history on two pages](#reg-1091))
- <a id="reg-1412"></a>**REG-1412** The status actions and "Register certificate" are buttons after the details of their record and before its child records. For example, the buttons of a brand are before its software products. ([the action row of a record](../../../web/packages/ui/design/components.md#ui-0246))
- <a id="reg-1413"></a>**REG-1413** Each brand and each software product is a card that starts collapsed. The title bar of the card shows the status of its record. ([the cards that start collapsed](../../../web/packages/ui/design/components.md#ui-0241); [the title bar of a collapsible card](../../../web/packages/ui/design/components.md#ui-0239))
- <a id="reg-1414"></a>**REG-1414** The heading "Brands" and each heading "Software products" have "Show all" and "Hide all". ([Show all and Hide all](../../../web/packages/ui/design/components.md#ui-0244))

## History

- <a id="reg-1014"></a>**REG-1014** The history of a participant gives each event of the stream of its legal entity, newest first. ([Judgement: event sourcing (ADR 0003)][adr-0003])
- <a id="reg-1015"></a>**REG-1015** Each entry of the history gives the time, the event type, the text of its event type and only the fields of its event type in the table of this group. The detail shows the current values of the other fields. ([Judgement: only the fields of the table in the history](operator-console-api.md#reg-1184))
- <a id="reg-1016"></a>**REG-1016** The history shows each event as one entry, in the order of the stream version. It does not group the events of one action. For example, a suspension and each software product that it makes INACTIVE are separate entries. ([Judgement: one entry for each event](#reg-1081))
- <a id="reg-1017"></a>**REG-1017** The name of a software product in the history is its name at the time of the event. For example, after a change of the name, an earlier entry shows the earlier name. ([Judgement: the name at the time of the event](#reg-1082))
- <a id="reg-1018"></a>**REG-1018** Each event type of a participant has one row in the table of this group. ([Judgement: one business fact in each event type (ADR 0021)][adr-0021])

The table gives the text and the fields of each event type in the history.

| ID | Event type | Text in the history | Fields | Source |
|---|---|---|---|---|
| <a id="reg-1019"></a>**REG-1019** | `DataRecipientImported`, `DataHolderImported` | Loaded into the Register | None | [Judgement: an event for a create](#reg-1083) |
| <a id="reg-1020"></a>**REG-1020** | `DataRecipientCreated`, `DataHolderCreated` | Created by the operator | None | [Judgement: an event for a create](#reg-1083) |
| <a id="reg-1021"></a>**REG-1021** | `DataRecipientDetailsChanged`, `DataHolderDetailsChanged` | Legal entity details changed | None | [Judgement: one business fact in each event type (ADR 0021)][adr-0021] |
| <a id="reg-1022"></a>**REG-1022** | `DataHolderLegalEntityRemoved` | Data holder removed | None | [Judgement: one business fact in each event type (ADR 0021)][adr-0021] |
| <a id="reg-1023"></a>**REG-1023** | `DataRecipientBrandAdded`, `DataHolderBrandAdded` | Brand added | Brand ID and name | [Judgement: the brand in the history](operator-console-api.md#reg-1186) |
| <a id="reg-1024"></a>**REG-1024** | `DataRecipientBrandDetailsChanged`, `DataHolderBrandDetailsChanged` | Brand details changed | Brand ID and name | [Judgement: the brand in the history](operator-console-api.md#reg-1186) |
| <a id="reg-1025"></a>**REG-1025** | `DataRecipientBrandActivated`, `DataHolderBrandActivated` | Brand activated | Brand ID and name | [Judgement: the brand in the history](operator-console-api.md#reg-1186) |
| <a id="reg-1026"></a>**REG-1026** | `DataRecipientBrandDeactivated`, `DataHolderBrandDeactivated` | Brand deactivated | Brand ID and name | [Judgement: the brand in the history](operator-console-api.md#reg-1186) |
| <a id="reg-1027"></a>**REG-1027** | `DataRecipientBrandRemoved`, `DataHolderBrandRemoved` | Brand removed | Brand ID and name | [Judgement: the brand in the history](operator-console-api.md#reg-1186) |
| <a id="reg-1028"></a>**REG-1028** | `SoftwareProductAdded` | Software product added | Software product ID and name | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1029"></a>**REG-1029** | `SoftwareProductDetailsChanged` | Software product details changed | Software product ID and name | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1030"></a>**REG-1030** | `AccreditationSuspended` | Accreditation suspended | None | [CDS: Participant Statuses][cds-participant-statuses] |
| <a id="reg-1031"></a>**REG-1031** | `AccreditationRevoked` | Accreditation revoked | None | [CDS: Participant Statuses][cds-participant-statuses] |
| <a id="reg-1032"></a>**REG-1032** | `AccreditationSurrendered` | Accreditation surrendered | None | [CDS: Participant Statuses][cds-participant-statuses] |
| <a id="reg-1033"></a>**REG-1033** | `AccreditationReinstated` | Accreditation reinstated | None | [CDS: Participant Statuses][cds-participant-statuses] |
| <a id="reg-1034"></a>**REG-1034** | `SoftwareProductActivated` | Software product activated | Software product ID and name | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1035"></a>**REG-1035** | `SoftwareProductDeactivated` | Software product deactivated | Software product ID and name | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1036"></a>**REG-1036** | `SoftwareProductRemoved` | Software product removed | Software product ID and name | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1037"></a>**REG-1037** | `ClientCertificateRegistered` | Client certificate registered | Software product ID and name, and the thumbprint | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1038"></a>**REG-1038** | `ClientCertificateIssued` | Client certificate issued | Software product ID and name, and the thumbprint | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1039"></a>**REG-1039** | `ClientCertificateRemoved` | Client certificate removed | Software product ID and name, and the thumbprint | [Judgement: the name at the time of the event](#reg-1082) |
| <a id="reg-1040"></a>**REG-1040** | `ServerCertificateIssued` | Server certificate issued | Brand ID and name, and the thumbprint | [Judgement: no server certificates in the detail of a data holder brand](#reg-1080) |

## Key sets

- <a id="reg-1041"></a>**REG-1041** The console shows the two key sets: the access token keys and the SSA keys. ([the two key sets](authorisation-server.md#reg-0465))
- <a id="reg-1042"></a>**REG-1042** Each key gives its key ID, its state and the time at which the Register created it. ([Judgement: the creation time of a key](#reg-1084))
- <a id="reg-1043"></a>**REG-1043** The state of a key is one of these. ([the two steps of a key change](authorisation-server.md#reg-0468))
  - the key that signs
  - the next key
  - the previous key, which the JWKS still publishes
- <a id="reg-1044"></a>**REG-1044** The keys of a key set are in this order: the key that signs, then the next key, then the previous key. ([Judgement: the order of the keys](#reg-1085))

## CDR certificates

- <a id="reg-1045"></a>**REG-1045** The CDR certificates page gives [the CDR CA bundle](certificate-management.md#reg-0616) in PEM. ([the CDR CA bundle for the lab operator](certificate-management.md#reg-0615); [Judgement: the Register is the CDR intermediate CA (ADR 0022)][adr-0022])
- <a id="reg-1046"></a>**REG-1046** The CDR certificates page gives the revocations of the mock CDR intermediate, newest first. Each revocation gives the serial number, the subject, the thumbprint, the time and the reason. ([Judgement: the revocations in the console](#reg-1086))
- <a id="reg-1047"></a>**REG-1047** The list of the revocations includes the certificates of the lab set-up that the lab operator revokes. For example, a revocation of the client certificate of the conformance suite is in the list. ([a revocation of a certificate of the lab set-up](certificate-revocation.md#reg-0701); [Judgement: the revocations in the console](#reg-1086))

## Presentation

- <a id="reg-1048"></a>**REG-1048** The console is for desktop browsers, with the minimum width of a console of the design system. ([the minimum width of a console](../../../web/packages/ui/design/layout.md#ui-0070); [Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="reg-1049"></a>**REG-1049** The console has the shell of a console of the design system, with the lab brand for each role and the compact density. ([the lab brand for each role that a participant brand does not give](../../../web/packages/ui/design/tokens-and-brands.md#ui-0034); [the compact density of a console](../../../web/packages/ui/design/tokens-and-brands.md#ui-0028))
- <a id="reg-1050"></a>**REG-1050** The side navigation marks the section of each page as current, also when the URL of the page is not under the URL of the section. For example, each form of a data recipient, a brand or a software product marks "Data recipients". ([Judgement: the current section in the side navigation](#reg-1087))
- <a id="reg-1051"></a>**REG-1051** The header and the title of the browser tab give the name "CDR Register". ([Judgement: the name "CDR Register"](#reg-1088))
- <a id="reg-1052"></a>**REG-1052** The filter, the sort and the page of a list are in the URL. For example, a reload of a list shows the same filter, sort and page. ([Judgement: the view in the URL](#reg-1089))
- <a id="reg-1053"></a>**REG-1053** The search of a list applies while the lab operator types, as the design system gives it. ([when the search of the filter bar applies](../../../web/packages/ui/design/components.md#ui-0271))
- <a id="reg-1054"></a>**REG-1054** A search replaces the history entry of the list. A filter, a sort and a page add a history entry. ([Judgement: the browser history of a list](#reg-1090))
- <a id="reg-1055"></a>**REG-1055** The detail and the history of a participant each have their own URL. The page of the history is in the URL. ([Judgement: the detail and the history on two pages](#reg-1091))
- <a id="reg-1056"></a>**REG-1056** The CDR certificates page has its own URL. The page of the list of the revocations is in the URL. ([Judgement: the view in the URL](#reg-1089))
- <a id="reg-1057"></a>**REG-1057** The console shows each time in the time zone of the browser, as the design system gives it. ([the time zone and the offset of a time](../../../web/packages/ui/design/content.md#ui-0147); [Judgement: the time zone of the browser](#reg-1092))
- <a id="reg-1058"></a>**REG-1058** The console loads no image from a participant. It shows a logo URI as text. ([OWASP ASVS 5.0, V3.6.1][asvs-v3]; [the Content Security Policy of the web UI](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007))
- <a id="reg-1059"></a>**REG-1059** The console shows each URI of a participant as text with a copy button, not as a link. For example, a redirect URI has a copy button. ([Judgement: the URIs of a participant as text](#reg-1093))
- <a id="reg-1060"></a>**REG-1060** The console has the accessibility level of an operator console. ([the accessibility level of an operator console](../../../web/packages/ui/design/accessibility.md#ui-0111))

## Listener and security

The console is a web UI. The [web UI library](../../../pkg/platform/consolehttp/DESIGN.md) gives the rules of its listener, its browser security, its cache headers and its routes.

- <a id="reg-1061"></a>**REG-1061** The console is on its own listener, separate from the CDR APIs and from the operator interface. ([the web UI and its API on one listener](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0001); [Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="reg-1062"></a>**REG-1062** The setting `CONSOLE_ENABLED` [turns the console on, and the console is off by default](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0002). ([Judgement: the web UI is off by default](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0018))
- <a id="reg-1063"></a>**REG-1063** The setting `CONSOLE_ALLOWED_HOSTS` is [the `Host` allowlist of the console](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0004). Its values are [the values of the `Host` header](../../../pkg/platform/hostallowlist/DESIGN.md#host-0001), with [no difference for upper and lower case](../../../pkg/platform/hostallowlist/DESIGN.md#host-0002). For example, `127.0.0.1:8084`. ([Judgement: a Host allowlist with no default](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0019))
- <a id="reg-1064"></a>**REG-1064** The console listener is plain HTTP. ([Judgement: plain HTTP for the console](#reg-1094))
- <a id="reg-1065"></a>**REG-1065** The console has no sign-in, and the browser gets no credential. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])

## Non-functional requirements

- <a id="reg-1066"></a>**REG-1066** Telemetry: the console API has the same server telemetry as the other listeners of the Register. For example, each request of the console API has a server span. ([Judgement: telemetry with OpenTelemetry (ADR 0006)][adr-0006])
- <a id="reg-1067"></a>**REG-1067** Telemetry: the UI sends no telemetry. ([Judgement: web UIs that the service embeds (ADR 0014)][adr-0014])
- <a id="reg-1068"></a>**REG-1068** Performance: on a development workstation, a list of 1,000 data recipients or 1,000 data holders shows within 1 second. ([Judgement: the performance limits of the console](#reg-1095))
- <a id="reg-1069"></a>**REG-1069** Performance: on a development workstation, the first page of a history of 1,000 events shows within 1 second. ([Judgement: the performance limits of the console](#reg-1095))

## Judgements

### <a id="reg-1080"></a>REG-1080 No server certificates in the detail of a data holder brand

- Kind: choice
- Basis: Step 3.3. [The Register does not register a server certificate](certificate-management.md#reg-0629), and no API of the Register uses it ([Judgement: the record of a server certificate issue](certificate-management.md#reg-0651)).
- Reason: The event of the issue is the only record of a server certificate. The history shows it with its thumbprint.

### <a id="reg-1081"></a>REG-1081 One entry for each event

- Kind: choice
- Basis: Step 3.3. The events have no field that names the action that caused them ([Judgement: one business fact in each event type (ADR 0021)][adr-0021]).
- Reason: A group needs a cause that the events do not have. A suspension and its cascade have the same time, and the stream version gives their order.

### <a id="reg-1082"></a>REG-1082 The name at the time of the event

- Kind: choice
- Basis: Step 3.3. The events of a software product give only its ID ([Judgement: one business fact in each event type (ADR 0021)][adr-0021]).
- Reason: A history shows the record as it was. The name comes from the state of the stream at that event.

### <a id="reg-1083"></a>REG-1083 An event for a create

- Kind: choice
- Basis: Step 3.3. An event type records one business fact ([Judgement: one business fact in each event type (ADR 0021)][adr-0021]).
- Reason: A create by the lab operator and a load are different facts. The history then shows if the load or the lab operator made the record.

### <a id="reg-1084"></a>REG-1084 The creation time of a key

- Kind: choice
- Basis: Step 3.3. The CDS gives no time for a signing key in the JWKS ([the two key sets](authorisation-server.md#reg-0465)).
- Reason: The Register keeps no other time for a key. The creation time shows the sequence of a key change.

### <a id="reg-1085"></a>REG-1085 The order of the keys

- Kind: choice
- Basis: Step 3.3. The CDS gives no order of the keys ([the two key sets](authorisation-server.md#reg-0465)).
- Reason: The order is the same as in the JWKS. The lab operator looks first for the key that signs.

### <a id="reg-1086"></a>REG-1086 The revocations in the console

- Kind: choice
- Basis: Step 3.3. The CRL gives the revocations only in DER ([the CRL in DER](certificate-revocation.md#reg-0718)). The Register gets the certificate in each revocation request ([the request body of a revocation](certificate-revocation.md#reg-0702)).
- Reason: The subject and the thumbprint let the lab operator compare a revocation with its confirmation and with the registered thumbprints of a software product. The Register can keep these values from the request.

### <a id="reg-1087"></a>REG-1087 The current section in the side navigation

- Kind: choice
- Basis: Step 3.3. The design system gives the side navigation, but not the section of a form ([the structure of a console](../../../web/packages/ui/design/layout.md#ui-0071)).
- Reason: The lab operator sees where they are in the console. A form is part of the section of its record.

### <a id="reg-1088"></a>REG-1088 The name "CDR Register"

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The lab operator manages the Register. "Console" is the name of the component, not of what the lab operator sees.

### <a id="reg-1089"></a>REG-1089 The view in the URL

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The lab operator can share a view, or go back to it, with its URL.

### <a id="reg-1090"></a>REG-1090 The browser history of a list

- Kind: choice
- Basis: Step 3.3. The search applies while the lab operator types ([when the search of the filter bar applies](../../../web/packages/ui/design/components.md#ui-0271)).
- Reason: With a replaced entry, Back does not go through each pause in the typing. A filter, a sort and a page are each one decision of the lab operator.

### <a id="reg-1091"></a>REG-1091 The detail and the history on two pages

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The history has its own page in the URL, as a list does. The lab operator can then share a page of the history or go back to it.

### <a id="reg-1092"></a>REG-1092 The time zone of the browser

- Kind: choice
- Basis: Step 3.3. [ADR 0012][adr-0012] permits a different time zone where a person sees the time.
- Reason: The lab operator compares a time in the console with their own clock. The offset lets them compare it with a time in UTC, for example in the logs.

### <a id="reg-1093"></a>REG-1093 The URIs of a participant as text

- Kind: choice
- Basis: Step 3.3. The UI loads no file from a different origin ([OWASP ASVS 5.0, V3.6.1][asvs-v3]).
- Reason: The URIs are for software, for example redirect URIs, API base URIs and JWKS URIs. The lab host names of the services do not always resolve in the browser of the lab operator. The lab operator copies a URI into a different tool.

### <a id="reg-1094"></a>REG-1094 Plain HTTP for the console

- Kind: choice
- Basis: Step 3.3. All lab ports bind to `127.0.0.1`, and the console sets no cookie ([ADR 0014][adr-0014]). Browsers send `Sec-Fetch-Site` to `localhost` over HTTP ([Fetch Metadata][fetch-metadata]).
- Reason: The operator interface is also plain HTTP. With no cookie, the console needs no `Secure` attribute. [The cross-origin check](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0005) works over HTTP.

### <a id="reg-1095"></a>REG-1095 The performance limits of the console

- Kind: choice
- Basis: Step 3.3. The CDS performance requirements are for the data holder APIs, not for a tool of the lab operator ([CDS: Non-functional Requirements][cds-nfrs]).
- Reason: The seed data is small. These limits find a page that loads all records at once.

[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[adr-0006]: ../../../docs/adr/0006-telemetry.md
[adr-0012]: ../../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0021]: ../../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0022]: ../../../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-nfrs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#non-functional-requirements
[fetch-metadata]: https://w3c.github.io/webappsec-fetch-metadata/
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
