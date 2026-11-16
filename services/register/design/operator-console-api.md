# Mock CDR Register: Operator console API

This area gives the console API of the operator console for the reads, the history, the key sets and the status actions. The [change API](operator-console-change-api.md) gives the creates, the adds, the changes and the CDR certificates. The console API is JSON over HTTP on the listener of the console, and only the console uses it. It obeys the [web UI library](../../../pkg/platform/consolehttp/DESIGN.md): [the `/api/` prefix](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0011), [404 for a path that is not an operation](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0012), [the page parameters](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0014), [the page sizes](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0015), [400 for a page that is not valid](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0016) and [an empty parameter is absent](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0017). The [operator console](operator-console.md) gives the UI.

## Operations

| ID | Action | Method and path | Success | Source |
|---|---|---|---|---|
| <a id="reg-1100"></a>**REG-1100** | Activate, deactivate or remove a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 | [the transitions of a brand](participants.md#reg-0225) |
| <a id="reg-1101"></a>**REG-1101** | Remove a data holder | `POST /api/data-holders/{legalEntityId}/remove` | 204 | [the transitions of a data holder legal entity](participants.md#reg-0224) |
| <a id="reg-1102"></a>**REG-1102** | Activate, deactivate or remove a data holder brand | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 | [the transitions of a brand](participants.md#reg-0225) |

- <a id="reg-1103"></a>**REG-1103** The OpenAPI document gives the path of each other read and status action. For example, the list of the data recipients is `GET /api/data-recipients`. ([Judgement: console API contracts in OpenAPI (ADR 0017)][adr-0017])
- <a id="reg-1104"></a>**REG-1104** The `{brandId}` of a data holder brand in a path is its `dataHolderBrandId`. A brand with no `dataHolderBrandId` has its `interimId` in the path. ([the operator interface](../DESIGN.md#operator-interface))

## Requests and errors

- <a id="reg-1105"></a>**REG-1105** An OpenAPI 3.0 document in the Register directory defines the console API. Tests in the Register check each response against the document. ([ADR 0017][adr-0017]; [ADR 0019][adr-0019])
- <a id="reg-1106"></a>**REG-1106** Each object schema of the OpenAPI document forbids the properties that it does not name. ([ADR 0017][adr-0017])
- <a id="reg-1107"></a>**REG-1107** The paths of the console API have no version. For example, the path of the key sets is `/api/key-sets`. ([Judgement: the UI and the API ship together (ADR 0014)][adr-0014])
- <a id="reg-1108"></a>**REG-1108** The console API ignores a query parameter that it does not know. For example, `GET /api/data-recipients?colour=red` gives the list. ([Judgement: an unknown query parameter is ignored](#reg-1170))
- <a id="reg-1109"></a>**REG-1109** An error response is a [problem details](../../../docs/spec/glossary.md#problem-details) body with [the type `about:blank`](../../../pkg/platform/problem/DESIGN.md#prob-0001), except a refusal with [the field-errors type](../../../pkg/platform/problem/DESIGN.md#prob-0002). ([RFC 9457][rfc9457])
- <a id="reg-1110"></a>**REG-1110** The status codes of an error are the codes of the operator interface: 400 for a request that is not valid, 404 for a record that does not exist, and 409 for a change that the status rules do not permit. ([the operator interface](../DESIGN.md#operator-interface); [RFC 9110, section 15.5.10][rfc9110-409])
- <a id="reg-1111"></a>**REG-1111** A legal entity ID that no legal entity has gives 404. The `detail` does not contain the ID. ([RFC 9110, section 15.5.5][rfc9110-404]; [the operator interface](../DESIGN.md#operator-interface))
- <a id="reg-1112"></a>**REG-1112** The console API gives each time in UTC. For example, `2026-10-23T01:02:03Z`. ([Judgement: UTC for all times (ADR 0012)][adr-0012])
- <a id="reg-1113"></a>**REG-1113** Telemetry: the console API puts no request body in telemetry. For example, no certificate, CSR or CA bundle is in telemetry. ([no CDR data and no personal information in telemetry](../../../pkg/platform/telemetry/DESIGN.md#tele-0004); [Judgement: telemetry with OpenTelemetry (ADR 0006)][adr-0006])
- <a id="reg-1114"></a>**REG-1114** Telemetry: a path with a thumbprint is in telemetry [as its route pattern](../../../pkg/platform/telemetry/DESIGN.md#tele-0005). For example, the thumbprint of a removed client certificate is not in a span name. ([Judgement: telemetry with OpenTelemetry (ADR 0006)][adr-0006])

## Lists

- <a id="reg-1115"></a>**REG-1115** The lists of the participants and the history of a participant are in pages. ([the page parameters](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0014))
- <a id="reg-1116"></a>**REG-1116** A response in pages gives the total number of records in all pages. ([Judgement: the total of a list](#reg-1171))
- <a id="reg-1117"></a>**REG-1117** A page after the last page gives 200 with no records and the total. For example, `page=9` of a list of 30 records gives no records and the total 30. ([Judgement: a page after the last page](#reg-1172))
- <a id="reg-1118"></a>**REG-1118** The list of the data recipients is filtered by status. ([the statuses of a data recipient legal entity](participants.md#reg-0202))
- <a id="reg-1119"></a>**REG-1119** A search of the list of the data recipients finds a part of the legal entity name, with no difference for upper and lower case. The characters `%` and `_` are text. For example, `budget` finds "Mock Budgeting Pty Ltd". ([Judgement: the search of a participant list](#reg-1173))
- <a id="reg-1120"></a>**REG-1120** The list of the data recipients is sorted by one of these values, and then by the legal entity ID. The default is the name. ([Judgement: the sorts of a participant list](#reg-1174))
  - the name
  - the accreditation number
  - the status
  - the number of brands
  - the number of software products
- <a id="reg-1121"></a>**REG-1121** A sort by a text value is in the byte order of the text. A sort by the name has no difference for upper and lower case. ([Judgement: the sorts of a participant list](#reg-1174))
- <a id="reg-1122"></a>**REG-1122** The numbers of brands and software products of a data recipient include each status. For example, a REMOVED brand is in the number of brands. ([Judgement: the counts in the data recipient list](#reg-1175))
- <a id="reg-1123"></a>**REG-1123** The list of the data holders is filtered by status. ([the statuses of a data holder legal entity](participants.md#reg-0200))
- <a id="reg-1124"></a>**REG-1124** A search of the list of the data holders finds a part of the legal entity name or of a brand name, with the rules of the data recipient search. For example, `wattle` finds the data holder "Wattle Bank Limited" by its brand "Wattle Bank". ([Judgement: the search of the data holder list](#reg-1176))
- <a id="reg-1125"></a>**REG-1125** The list of the data holders is sorted by the name, the status or the number of brands, and then by the legal entity ID. The default is the name. ([Judgement: the sorts of a participant list](#reg-1174))
- <a id="reg-1126"></a>**REG-1126** The list of the data holders gives, for each legal entity, the number of its brands and the industries of its brands, not each brand. The number of brands includes each status. ([Judgement: the brands in the data holder list](#reg-1177))

## Details

- <a id="reg-1127"></a>**REG-1127** The detail of a participant is one response, with each brand and each software product. It is not in pages. ([Judgement: one response for a detail and for the key sets](#reg-1178))
- <a id="reg-1128"></a>**REG-1128** The detail of a participant gives each brand and each software product in each status, in the order of the record. ([Judgement: one response for a detail and for the key sets](#reg-1178))
- <a id="reg-1129"></a>**REG-1129** The fields of a detail have the names of the CDS Register schemas. The time of the last change is `lastUpdated`, and the thumbprints of the registered client certificates are `clientCertificateThumbprints`. ([CDS: CDR Register API schemas][cds-register-schemas]; [Judgement: the field names of a detail](#reg-1179))
- <a id="reg-1130"></a>**REG-1130** An optional field with no value is not in a detail, and a detail has no `null`. A list is always in the detail, also when it is empty. ([Judgement: the field names of a detail](#reg-1179))
- <a id="reg-1131"></a>**REG-1131** The detail of a data recipient and the detail of a data holder give the version of the participant. The version changes at each change of the participant. ([Judgement: a version for a stale view](#reg-1180))

## Status actions

- <a id="reg-1132"></a>**REG-1132** The detail of a data recipient gives, for the legal entity, each brand and each software product, the status actions that the status rules permit for the current state of the record. ([the transitions of a data recipient legal entity](participants.md#reg-0222); [the cascades of a suspension](participants.md#reg-0227); [Judgement: the permitted actions from the domain](#reg-1181))
- <a id="reg-1133"></a>**REG-1133** The detail of a data holder gives, for the legal entity and each brand, the status actions that the status rules permit for the current state of the record. ([the transitions of a data holder legal entity](participants.md#reg-0224); [the cascade of a data holder removal](participants.md#reg-0231); [Judgement: the permitted actions from the domain](#reg-1181))
- <a id="reg-1134"></a>**REG-1134** Each permitted status action gives each other record that it changes, with the status before and after. For example, the suspension of a data recipient gives each ACTIVE software product, with ACTIVE before and INACTIVE after. ([Judgement: the permitted actions from the domain](#reg-1181))
- <a id="reg-1135"></a>**REG-1135** The Register finds the permitted actions and their changes with the same domain rules as the action itself. It changes no record to find them. ([Judgement: the permitted actions from the domain](#reg-1181))
- <a id="reg-1136"></a>**REG-1136** The detail of a data recipient gives, for each software product, `clientCertificateActions`. It has `register` and `issue` when the rules permit them, and `remove` in each status. ([Judgement: the permitted certificate actions](#reg-1182))
- <a id="reg-1137"></a>**REG-1137** The detail of a data holder gives, for each brand, `serverCertificateActions`. It has `issue` only for a brand with a `dataHolderBrandId` that is not REMOVED. ([no server certificate for a REMOVED brand](certificate-management.md#reg-0626); [no server certificate for a brand with no `dataHolderBrandId`](certificate-management.md#reg-0628); [Judgement: the permitted certificate actions](#reg-1182))
- <a id="reg-1138"></a>**REG-1138** Each action is a `POST` with a JSON body. For example, a suspension is `POST /api/data-recipients/{legalEntityId}/suspend`. ([RFC 9110, section 9.3.3][rfc9110-post]; [Judgement: console API contracts in OpenAPI (ADR 0017)][adr-0017])
- <a id="reg-1139"></a>**REG-1139** The body of an action gives the version of the participant or the key set that the lab operator saw. The body of a create, of the check of a revocation and of a revocation has no version. ([Judgement: a version for a stale view](#reg-1180))
- <a id="reg-1140"></a>**REG-1140** When the version in the body is not the version of the participant or the key set, the response is 409, and the Register changes nothing. The `detail` says that the record changed. ([Judgement: a version for a stale view](#reg-1180))
- <a id="reg-1141"></a>**REG-1141** An action that succeeds gives 204 with no body, except a create, an add, an issue and the check of a revocation. ([Judgement: no body after an action](#reg-1183))

## History

- <a id="reg-1142"></a>**REG-1142** The history of a participant is sorted by the stream version, newest first. ([Judgement: event sourcing (ADR 0003)][adr-0003])
- <a id="reg-1143"></a>**REG-1143** Each entry of the history gives the stream version, the event type, the time at which the Register recorded the event, and only the fields that [the table of the history](operator-console.md#reg-1018) gives for its event type. ([Judgement: only the fields of the table in the history](#reg-1184))
- <a id="reg-1144"></a>**REG-1144** The console API gives no payload and no metadata of an event. ([Judgement: only the fields of the table in the history](#reg-1184))
- <a id="reg-1145"></a>**REG-1145** The history has one closed schema for each group of event types with the same fields, with `eventType` as the discriminator. For example, an entry of a suspension with a thumbprint does not agree with the document. ([Judgement: one schema for each group of event types](#reg-1185))
- <a id="reg-1146"></a>**REG-1146** An entry of a brand gives `brandId` and `brandName`. The `brandId` of a data holder brand is its `dataHolderBrandId`, or its `interimId` when it has no `dataHolderBrandId`. ([Judgement: the brand in the history](#reg-1186))

## Key sets

- <a id="reg-1147"></a>**REG-1147** The key sets are one response. ([Judgement: one response for a detail and for the key sets](#reg-1178))
- <a id="reg-1148"></a>**REG-1148** The key sets give no private key and no public key. For example, a key gives only its key ID, its state and its creation time. ([Judgement: no key material in the console](#reg-1187))
- <a id="reg-1149"></a>**REG-1149** The key sets are `access-token` and `ssa`. The states of a key are `signing`, `next` and `previous`. ([Judgement: the names of the key sets and the states](#reg-1188))
- <a id="reg-1150"></a>**REG-1150** Each key set gives the actions that the rules permit for its current state, with the state of each key before and after. ([the two steps of a key change](authorisation-server.md#reg-0468); [Judgement: the permitted actions from the domain](#reg-1181))
- <a id="reg-1151"></a>**REG-1151** Each key set gives its version. The version is 1 when the Register creates the key set. ([Judgement: the version of a key set](#reg-1189))
- <a id="reg-1152"></a>**REG-1152** Each publish and each activation adds 1 to the version of the key set, also through the operator interface. A step that the rules refuse does not change the version. ([Judgement: the version of a key set](#reg-1189))

## Judgements

### <a id="reg-1170"></a>REG-1170 An unknown query parameter is ignored

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the console API, and RFC 9110 gives no rule for an unknown query parameter.
- Reason: A query parameter changes no record, so a refusal protects nothing. To read only the known parameters is the least complex.

### <a id="reg-1171"></a>REG-1171 The total of a list

- Kind: choice
- Basis: Step 3.3. The [page parameters of the web UI library](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0014) give no total.
- Reason: A total lets the UI show the number of pages.

### <a id="reg-1172"></a>REG-1172 A page after the last page

- Kind: choice
- Basis: Step 3.3. The CDS gives 422 for a page after the last page, but only for the CDR APIs ([CDS: Pagination][cds-pagination]).
- Reason: After a change of the filter, the UI can show "no records" and go back to page 1.

### <a id="reg-1173"></a>REG-1173 The search of a participant list

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The lab operator knows a part of the name. A pattern syntax surprises the lab operator.

### <a id="reg-1174"></a>REG-1174 The sorts of a participant list

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The byte order is the same in each database locale, and the legal entity ID keeps each page stable. A sort by a count lets the lab operator find the largest participants. A data holder has no accreditation number and no software products, so its list has fewer sorts.

### <a id="reg-1175"></a>REG-1175 The counts in the data recipient list

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: The list shows the size of each data recipient. The detail shows each status.

### <a id="reg-1176"></a>REG-1176 The search of the data holder list

- Kind: choice
- Basis: Step 3.3. The CDR APIs and the lab services name a data holder by its brand ([CDS: Get Data Holder Brands][cds-get-dh-brands]).
- Reason: The lab operator often knows only the brand name. For example, each preset brand is a brand name.

### <a id="reg-1177"></a>REG-1177 The brands in the data holder list

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no rule for a tool of the lab operator.
- Reason: One row for each legal entity stays short, also for a data holder with many brands. The detail shows each brand with its status.

### <a id="reg-1178"></a>REG-1178 One response for a detail and for the key sets

- Kind: choice
- Basis: Step 3.3. The records of one legal entity are one aggregate and one stream ([ADR 0003][adr-0003]).
- Reason: A legal entity in the seed data has 13 software products or fewer. A key set has three keys or fewer.

### <a id="reg-1179"></a>REG-1179 The field names of a detail

- Kind: choice
- Basis: Step 3.3. The CDS Register schemas name the fields of a participant, but not the thumbprints ([CDS: CDR Register API schemas][cds-register-schemas]).
- Reason: The lab operator sees the same names in the seed file, the CDR APIs and the console. A field that is not present is simpler than `null` in the closed schemas and the generated types. The name of the thumbprints says what the values are.

### <a id="reg-1180"></a>REG-1180 A version for a stale view

- Kind: choice
- Basis: Step 3.2. `If-Match` with 412 is the HTTP form of a check of a stale view ([RFC 9110, section 13.1.1][rfc9110-if-match]). The event store appends only at the expected version ([ADR 0003][adr-0003]).
- Reason: A confirmation or a form for an earlier version can give false changes or values, and the version check costs little. `If-Match` adds two status codes that the operator interface does not use. A `POST` to an action has no representation for an entity tag.

### <a id="reg-1181"></a>REG-1181 The permitted actions from the domain

- Kind: choice
- Basis: Step 3.3. The status rules are in the domain ([ADR 0013][adr-0013]).
- Reason: A copy of the rules in the UI can give a confirmation that is not true. A separate preview request for each action gives the same data, with one more request and one more endpoint for each action.

### <a id="reg-1182"></a>REG-1182 The permitted certificate actions

- Kind: choice
- Basis: Step 3.3. The certificate rules are in the domain ([ADR 0013][adr-0013]).
- Reason: A certificate action does not change a status, so it is not a status action. The rules permit a revocation and the CA bundle for each certificate of the intermediate, so the console API gives no permitted action for them.

### <a id="reg-1183"></a>REG-1183 No body after an action

- Kind: choice
- Basis: Step 3.3. An action of the operator interface gives 204 with no body ([the operator interface](../DESIGN.md#operator-interface)).
- Reason: One response for each action in both interfaces is the least complex. The UI gets the data of its views again after each action.

### <a id="reg-1184"></a>REG-1184 Only the fields of the table in the history

- Kind: choice
- Basis: Step 3.3. The console serves only data that a domain type gives ([ADR 0011][adr-0011]), and each console API schema is closed ([ADR 0017][adr-0017]).
- Reason: The fields of the table name the record that the event changes. The payload of a load is the full record, which the detail shows. The metadata is trace context, which is not participant data.

### <a id="reg-1185"></a>REG-1185 One schema for each group of event types

- Kind: choice
- Basis: Step 3.3. The check of a response refuses a field that its closed schema does not name ([ADR 0017][adr-0017]).
- Reason: One schema with optional fields permits, for example, a thumbprint on a suspension. A schema for each group refuses it.

### <a id="reg-1186"></a>REG-1186 The brand in the history

- Kind: choice
- Basis: Step 3.3. A data holder brand can have an interim ID and no brand ID ([the identifiers of a data holder brand](participants.md#reg-0208)).
- Reason: One pair of fields for each kind of brand is the least complex. The identifier is the same as in the path of the brand in the operator interface.

### <a id="reg-1187"></a>REG-1187 No key material in the console

- Kind: choice
- Basis: Step 3.3. The JWKS endpoints give the public keys ([CDS: Get JWKS][cds-get-jwks]).
- Reason: The console needs only the key ID to compare with a `kid`. A private key never leaves the Register.

### <a id="reg-1188"></a>REG-1188 The names of the key sets and the states

- Kind: choice
- Basis: Step 3.3. The CDS does not name the key sets or the states of a key.
- Reason: The names of the key sets are the same as in the paths of the operator interface. A console action then uses the value with no mapping.

### <a id="reg-1189"></a>REG-1189 The version of a key set

- Kind: choice
- Basis: Step 3.3. The key sets are not event sourced, so they have no stream version ([ADR 0003][adr-0003]).
- Reason: The same type as the version of a participant gives one request body for each action. A value from the key IDs and their states also changes at each step, but the check of a stale view then needs its own lock.

[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[adr-0006]: ../../../docs/adr/0006-telemetry.md
[adr-0011]: ../../../docs/adr/0011-domain-data-as-typed-go.md
[adr-0012]: ../../../docs/adr/0012-utc-for-all-times.md
[adr-0013]: ../../../docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0017]: ../../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0019]: ../../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#section-15.5.10
[rfc9110-post]: https://www.rfc-editor.org/rfc/rfc9110#section-9.3.3
[rfc9110-if-match]: https://www.rfc-editor.org/rfc/rfc9110#section-13.1.1
[cds-pagination]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-get-jwks]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-jwks
