# Mock CDR Register: Operator console change API

This area gives the console API of the operator console for the creates, the adds, the changes and the CDR certificates. The [console API](operator-console-api.md) gives the reads, the history, the key sets, the status actions and the rules of all requests. The [forms](operator-console-forms.md) and the [actions](operator-console-actions.md) give the UI.

## Operations

| ID | Action | Method and path | Success | Source |
|---|---|---|---|---|
| <a id="reg-1300"></a>**REG-1300** | Create a data recipient | `POST /api/data-recipients` | 201 | [Judgement: the response of a create or an add](#reg-1382) |
| <a id="reg-1301"></a>**REG-1301** | Change the details of a data recipient | `POST /api/data-recipients/{legalEntityId}/change` | 204 | [Judgement: a `POST` for a change](#reg-1381) |
| <a id="reg-1302"></a>**REG-1302** | Add a brand to a data recipient | `POST /api/data-recipients/{legalEntityId}/brands` | 201 | [Judgement: the response of a create or an add](#reg-1382) |
| <a id="reg-1303"></a>**REG-1303** | Change the details of a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/change` | 204 | [Judgement: a `POST` for a change](#reg-1381) |
| <a id="reg-1304"></a>**REG-1304** | Add a software product to a data recipient brand | `POST /api/data-recipients/{legalEntityId}/brands/{brandId}/software-products` | 201 | [Judgement: the response of a create or an add](#reg-1382) |
| <a id="reg-1305"></a>**REG-1305** | Change the details of a software product | `POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/change` | 204 | [Judgement: a `POST` for a change](#reg-1381) |
| <a id="reg-1306"></a>**REG-1306** | Create a data holder | `POST /api/data-holders` | 201 | [Judgement: the response of a create or an add](#reg-1382) |
| <a id="reg-1307"></a>**REG-1307** | Change the details of a data holder | `POST /api/data-holders/{legalEntityId}/change` | 204 | [Judgement: a `POST` for a change](#reg-1381) |
| <a id="reg-1308"></a>**REG-1308** | Add a brand to a data holder | `POST /api/data-holders/{legalEntityId}/brands` | 201 | [Judgement: the response of a create or an add](#reg-1382) |
| <a id="reg-1309"></a>**REG-1309** | Change the details of a data holder brand | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/change` | 204 | [Judgement: a `POST` for a change](#reg-1381) |
| <a id="reg-1310"></a>**REG-1310** | Remove a client certificate of a software product | `POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/{thumbprint}/remove` | 204 | [Judgement: the path of a certificate removal](#reg-1383) |
| <a id="reg-1311"></a>**REG-1311** | Issue a client certificate for a software product from a CSR, and register it | `POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue` | 200 | [the issue of a client certificate](certificate-management.md#reg-0632) |
| <a id="reg-1312"></a>**REG-1312** | Issue a server certificate for a data holder brand from a CSR | `POST /api/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue` | 200 | [the issue of a server certificate](certificate-management.md#reg-0634) |
| <a id="reg-1313"></a>**REG-1313** | Get the CDR CA bundle | `GET /api/cdr-ca-bundle` | 200 | [the CDR CA bundle in the operator interface](certificate-management.md#reg-0633) |
| <a id="reg-1314"></a>**REG-1314** | Get one page of the revocations of the mock CDR intermediate | `GET /api/cdr-certificates/revocations` | 200 | [Judgement: the revocations in the console](operator-console.md#reg-1086) |
| <a id="reg-1315"></a>**REG-1315** | Check a revocation with no change | `POST /api/cdr-certificates/check-revocation` | 200 | [Judgement: the confirmation of a revocation](operator-console-actions.md#reg-1292) |
| <a id="reg-1316"></a>**REG-1316** | Revoke a certificate of the mock CDR intermediate | `POST /api/cdr-certificates/revoke` | 204 | [the revocation in the operator interface](certificate-revocation.md#reg-0731) |

## Creates, adds and changes

- <a id="reg-1320"></a>**REG-1320** A create and an add are a `POST` on the path of the list of the record. For example, an add of a data holder brand is a `POST` on `/api/data-holders/{legalEntityId}/brands`. ([RFC 9110, section 9.3.3][rfc9110-post]; [Judgement: the response of a create or an add](#reg-1382))
- <a id="reg-1321"></a>**REG-1321** A create and an add give 201 with no body. The `Location` header gives the path of the new record in the console API, which is the start of the paths of its actions. ([RFC 9110, section 15.3.2][rfc9110-201]; [Judgement: the response of a create or an add](#reg-1382))
- <a id="reg-1322"></a>**REG-1322** A change is a `POST` on the path of the record, with `/change` at the end. ([Judgement: a `POST` for a change](#reg-1381))
- <a id="reg-1323"></a>**REG-1323** The body of a change gives the version and each field of the change, with the names and the format of the operator interface. ([the body of a change in the operator interface](operator-interface.md#reg-0818); [Judgement: a `POST` for a change](#reg-1381))
- <a id="reg-1324"></a>**REG-1324** A field that is not in the body of a change has no value after the change. For example, a change with no `policyUri` removes the policy URI of the software product. ([a field that is not in the body of a change](operator-interface.md#reg-0819))
- <a id="reg-1325"></a>**REG-1325** The body of a create gives the fields of one record, as a create in the operator interface does, with no child records and no client certificates. ([the body of a create in the operator interface](operator-interface.md#reg-0809))
- <a id="reg-1326"></a>**REG-1326** The body of an add gives the version and the fields of one record, with no child records and no client certificates. ([the body of an add in the operator interface](operator-interface.md#reg-0811); [Judgement: a version for a stale view](operator-console-api.md#reg-1180))
- <a id="reg-1327"></a>**REG-1327** The response is 400 for a body of a create, an add or a change that breaks one of these rules. ([400 for a body that is not one JSON object](operator-interface.md#reg-0852); [400 for a body larger than 1 MiB](operator-interface.md#reg-0854); [a field that the operation does not accept](../../../pkg/platform/jsonhttp/DESIGN.md#body-0001))
  - The body is one JSON object.
  - Each field of the body is a field that the action accepts.
  - The body is 1 MiB or smaller.
- <a id="reg-1328"></a>**REG-1328** When the Register refuses a value in the body, the problem details give [the field errors of the problem details library](../../../pkg/platform/problem/DESIGN.md#prob-0003). Each field error has [a JSON Pointer to the field](../../../pkg/platform/problem/DESIGN.md#prob-0004), for example `#/redirectUris`, and [a `detail` that the UI can show at the field](../../../pkg/platform/problem/DESIGN.md#prob-0005). ([RFC 9457, section 3][rfc9457-3])
- <a id="reg-1329"></a>**REG-1329** These refusals have field errors, with the status code of the operator interface. ([400 for a value that breaks the participant model](operator-interface.md#reg-0855); [409 for an identifier in use](operator-interface.md#reg-0858); [409 for a body that breaks a status rule](operator-interface.md#reg-0856); [Judgement: the status codes of field errors](#reg-1380))
  - A value that does not obey the participant model: 400, at the field of the value.
  - An identifier that a different record has: 409, at the field of the identifier.
  - A legal entity ID that a legal entity has: 409, at `#/legalEntityId`.
  - A status in the body that breaks a status rule: 409, at `#/status`.
- <a id="reg-1330"></a>**REG-1330** A refusal that names no field has [no `errors`](../../../pkg/platform/problem/DESIGN.md#prob-0007). For example, a stale version has no `errors`. ([RFC 9457, section 3][rfc9457-3])
- <a id="reg-1331"></a>**REG-1331** The `detail` of a problem with field errors gives each problem that names no field. When each problem names a field, the `detail` is [the general `detail` of field errors](../../../pkg/platform/problem/DESIGN.md#prob-0006). ([RFC 9457, section 3][rfc9457-3])

## CDR certificates

- <a id="reg-1332"></a>**REG-1332** The registration of a client certificate gives the PEM certificate in the JSON body, with the rules of the operator interface for the certificate. ([the request body of a registration](operator-interface.md#reg-0835))
- <a id="reg-1333"></a>**REG-1333** An issue gives the version and the PEM CSR (`csr`) in the JSON body. The rules of the operator interface for the CSR, the record and the certificate apply. ([the request body of an issue](certificate-management.md#reg-0602); [the CSR rules of the Register](certificate-management.md#reg-0604))
- <a id="reg-1334"></a>**REG-1334** A successful issue gives 200 with a JSON body. The body gives the PEM chain of the operator interface (`certificateChain`) and the thumbprint of the certificate (`thumbprint`). ([the response of an issue](certificate-management.md#reg-0612); [Judgement: JSON for an issue and for the CA bundle](#reg-1384))
- <a id="reg-1335"></a>**REG-1335** The `thumbprint` of an issue is in the base64url format of the RFC 8705 `x5t#S256`. ([RFC 8705, section 3.1][rfc8705-x5t])
- <a id="reg-1336"></a>**REG-1336** When the Register refuses a CSR, the response is 400, and the `detail` names each rule that the CSR breaks, as in the operator interface. ([a CSR that breaks the CSR profile](certificate-management.md#reg-0605); [Judgement: the errors of a CSR and of a revocation](operator-console-actions.md#reg-1291))
- <a id="reg-1337"></a>**REG-1337** The CDR CA bundle gives 200 with a JSON body. The body gives the PEM bundle of the operator interface (`caBundle`). ([the CDR CA bundle](certificate-management.md#reg-0616); [Judgement: JSON for an issue and for the CA bundle](#reg-1384))
- <a id="reg-1338"></a>**REG-1338** The list of the revocations is in pages, with the rules of the other lists. ([the lists in pages](operator-console-api.md#reg-1115); [the total of a list](operator-console-api.md#reg-1116); [a page after the last page](operator-console-api.md#reg-1117))
- <a id="reg-1339"></a>**REG-1339** Each revocation gives the serial number, the subject as an [RFC 4514][rfc4514] string, the thumbprint, the time of the revocation and the reason. ([Judgement: the revocations in the console](operator-console.md#reg-1086))
- <a id="reg-1340"></a>**REG-1340** The list of the revocations is sorted by the time of the revocation, newest first, and then by the serial number. ([Judgement: the revocations in the console](operator-console.md#reg-1086))
- <a id="reg-1341"></a>**REG-1341** The console API gives a serial number in uppercase hexadecimal, with two digits for each byte and no separators. For example, `01A2FF`, as `openssl x509 -serial` gives it. ([Judgement: the format of a serial number](#reg-1386))
- <a id="reg-1342"></a>**REG-1342** The check of a revocation and the revocation give the PEM certificate (`certificate`) and the reason (`reason`) in the JSON body, with the rules of the operator interface. The body has [no version](operator-console-api.md#reg-1139). ([the request body of a revocation](certificate-revocation.md#reg-0702); [Judgement: no version in a revocation](#reg-1385))
- <a id="reg-1343"></a>**REG-1343** The check of a revocation makes no change. A successful check gives 200 with the subject, the serial number, the thumbprint and the end of the validity of the first certificate. ([Judgement: the confirmation of a revocation](operator-console-actions.md#reg-1292))
- <a id="reg-1344"></a>**REG-1344** A successful revocation gives 204. ([Judgement: the response of a revocation](certificate-revocation.md#reg-0742))
- <a id="reg-1345"></a>**REG-1345** When the Register refuses a value of a check or a revocation, the response is 400 with a field error at `#/certificate` or `#/reason`. ([Judgement: the errors of a CSR and of a revocation](operator-console-actions.md#reg-1291))
- <a id="reg-1346"></a>**REG-1346** When the certificate is already revoked, the check and the revocation give 409 with no field errors. The `detail` gives the reason of the revocation that exists. ([a second revocation gives 409](certificate-revocation.md#reg-0706))

## Judgements

### <a id="reg-1380"></a>REG-1380 The status codes of field errors

- Kind: choice
- Basis: Step 3.3. The operator interface gives [400 for a value that breaks the participant model](operator-interface.md#reg-0855), [409 for an identifier in use](operator-interface.md#reg-0858) and [409 for a body that breaks a status rule](operator-interface.md#reg-0856) ([Judgement: 400 or 409 for a refused body](operator-interface.md#reg-0894)).
- Reason: One rule for the status codes in the two interfaces is the least complex. The field errors are in the body, so the UI does not need a different code.

### <a id="reg-1381"></a>REG-1381 A `POST` for a change

- Kind: choice
- Basis: Step 3.2. A `PUT` replaces the representation that a `GET` of the same path gives ([RFC 9110, section 9.3.4][rfc9110-put]).
- Reason: The `GET` of a participant gives its detail with its child records, its version and its actions, not the fields of the change. A brand and a software product have no `GET`. A `POST` with the version in the body is the same form as each other console action.

### <a id="reg-1382"></a>REG-1382 The response of a create or an add

- Kind: choice
- Basis: Step 3.2. A `POST` that makes a resource gives 201 with a `Location` ([RFC 9110, section 9.3.3][rfc9110-post]).
- Reason: The response is the same as in the operator interface. The UI knows the identifiers, because it sent them, so a body adds nothing.

### <a id="reg-1383"></a>REG-1383 The path of a certificate removal

- Kind: choice
- Basis: Step 3.2. The content of a `DELETE` has no defined meaning ([RFC 9110, section 9.3.5][rfc9110-delete]).
- Reason: Each console action is a `POST` with the version in a JSON body ([Judgement: a version for a stale view](operator-console-api.md#reg-1180)). The version cannot go in the body of a `DELETE`.

### <a id="reg-1384"></a>REG-1384 JSON for an issue and for the CA bundle

- Kind: choice
- Basis: Step 3.3. The console API is JSON, and its OpenAPI document and the generated types describe JSON bodies ([ADR 0017][adr-0017]).
- Reason: The thumbprint goes in the same response as the chain. The operator interface keeps the content type `application/pem-certificate-chain` for scripts.

### <a id="reg-1385"></a>REG-1385 No version in a revocation

- Kind: choice
- Basis: Step 3.3. A second revocation gives 409 with no change ([Judgement: 409 for a second revocation](certificate-revocation.md#reg-0744)).
- Reason: A revocation is not of a participant or a key set. The only state that it depends on is whether the certificate is revoked. Thus a stale view cannot cause a change that the lab operator did not confirm.

### <a id="reg-1386"></a>REG-1386 The format of a serial number

- Kind: choice
- Basis: Step 3.3. The CDS and the ACCC give no format for a serial number.
- Reason: The lab operator compares the serial number in the confirmation and in the list with the output of `openssl x509 -serial`, which has this format. The Register keeps its own format in the database, and the console API changes only the format of its responses.

[adr-0017]: ../../../docs/adr/0017-console-api-contracts-in-openapi.md
[rfc4514]: https://www.rfc-editor.org/rfc/rfc4514
[rfc8705-x5t]: https://www.rfc-editor.org/rfc/rfc8705#section-3.1
[rfc9110-post]: https://www.rfc-editor.org/rfc/rfc9110#section-9.3.3
[rfc9110-put]: https://www.rfc-editor.org/rfc/rfc9110#section-9.3.4
[rfc9110-delete]: https://www.rfc-editor.org/rfc/rfc9110#section-9.3.5
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#section-15.3.2
[rfc9457-3]: https://www.rfc-editor.org/rfc/rfc9457#section-3
