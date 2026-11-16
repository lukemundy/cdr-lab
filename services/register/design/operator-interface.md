# Mock CDR Register: Operator interface

This area gives the lab-only HTTP API that creates and changes participants, registers client certificates and changes participant statuses. In production, the CDR Registrar keeps the participant data and statuses on the Register ([CDS: Participant Statuses][cds-participant-statuses]). The interface of the production Register is not public, so the lab has its own interface. Real components must not depend on it. The [certificate management](certificate-management.md) and [certificate revocation](certificate-revocation.md) areas give the certificate issue and revocation actions of the operator interface.

## Status actions

- <a id="reg-0800"></a>**REG-0800** The lab operator changes the participants and their statuses through the operator interface. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: a lab-only operator interface](#reg-0880))
- <a id="reg-0801"></a>**REG-0801** The lab operator can suspend, revoke, surrender and reinstate a data recipient legal entity. ([CDS: Participant Statuses][cds-participant-statuses])
- <a id="reg-0802"></a>**REG-0802** The lab operator can activate, deactivate and remove a software product. ([CDS: Participant Statuses][cds-participant-statuses])
- <a id="reg-0803"></a>**REG-0803** The lab operator can activate, deactivate and remove a data recipient brand and a data holder brand. ([CDS: Participant Statuses][cds-participant-statuses])
- <a id="reg-0804"></a>**REG-0804** The lab operator can remove a data holder legal entity. ([CDS: Participant Statuses][cds-participant-statuses])
- <a id="reg-0805"></a>**REG-0805** Each status action obeys [the status transitions](participants.md#reg-0222) and [the cascades](participants.md#reg-0227). For example, a suspension of a data recipient makes its ACTIVE software products INACTIVE. ([CDS: Status Mapping][cds-status-mapping])
- <a id="reg-0806"></a>**REG-0806** When the lab operator reinstates a data recipient to ACTIVE, the status of its software products does not change. The lab operator activates each software product. ([CDS: Status Mapping][cds-status-mapping])
- <a id="reg-0807"></a>**REG-0807** A removal is a status action that sets the status REMOVED. The Register deletes no record. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: a removal is a change of status](participants.md#reg-0270))
- <a id="reg-0808"></a>**REG-0808** A data recipient legal entity has no remove action, because it has no REMOVED status. Its revocation or its surrender is its removal. ([CDS: Participant Statuses][cds-participant-statuses])

## Creates, adds and changes

- <a id="reg-0809"></a>**REG-0809** The lab operator can create a data recipient and a data holder. The request body is one legal entity in [the format of the load file](deployment.md#reg-0905), with its brands and, for a data recipient, their software products. ([Judgement: the format of the load file for a create or an add](#reg-0888))
- <a id="reg-0810"></a>**REG-0810** When a legal entity with the same identifier exists, a create gives 409. ([RFC 9110, section 15.5.10][rfc9110-409])
- <a id="reg-0811"></a>**REG-0811** The lab operator can add a brand to a legal entity, and a software product to a data recipient brand. The request body is one brand or one software product in the format of the load file. ([Judgement: the format of the load file for a create or an add](#reg-0888))
- <a id="reg-0812"></a>**REG-0812** The body of an add of a data recipient brand can have the software products of the brand. ([Judgement: the format of the load file for a create or an add](#reg-0888))
- <a id="reg-0813"></a>**REG-0813** The body of a create or an add gives the identifier of each new record, with [the rules of an identifier](participants.md#reg-0219). ([CDS: Register API schemas][cds-register-schemas]; [Judgement: the lab operator gives the identifiers](#reg-0889))
- <a id="reg-0814"></a>**REG-0814** The body of a create or an add can give each status to each record, with the status rules. For example, the lab operator can add a software product that is INACTIVE. ([Judgement: a status in a create or an add](#reg-0890))
- <a id="reg-0815"></a>**REG-0815** A create and an add give 201 with no body. ([RFC 9110, section 15.3.2][rfc9110-201]; [Judgement: the response of a create or an add](#reg-0891))
- <a id="reg-0816"></a>**REG-0816** The `Location` header of a create or an add gives the path of the new record in the operator interface. For example, a create of a data holder gives `/v1/data-holders/{legalEntityId}`. ([RFC 9110, section 15.3.2][rfc9110-201]; [Judgement: the response of a create or an add](#reg-0891))
- <a id="reg-0817"></a>**REG-0817** A change is a `PUT` on the path of the record. A change gives 204. ([RFC 9110, section 9.3.4][rfc9110-put]; [Judgement: a change replaces each field of the record](#reg-0892))
- <a id="reg-0818"></a>**REG-0818** The body of a change gives each field of the record in the format of the load file, except the identifiers, the status, the brands, the software products and the client certificates. ([Judgement: a change replaces each field of the record](#reg-0892))
- <a id="reg-0819"></a>**REG-0819** When a field is not in the body of a change, the field has no value after the change. ([RFC 9110, section 9.3.4][rfc9110-put]; [Judgement: a change replaces each field of the record](#reg-0892))
- <a id="reg-0820"></a>**REG-0820** The identifiers of a record do not change. ([Judgement: identifiers that do not change](#reg-0893))
- <a id="reg-0821"></a>**REG-0821** The path of a data holder brand has its `dataHolderBrandId`. When the brand has no `dataHolderBrandId`, the path has its `interimId`. ([CDS: Register API schemas][cds-register-schemas])
- <a id="reg-0822"></a>**REG-0822** After a create, an add or a change, the participant obeys the participant model and the status rules, the same as after a load. ([Judgement: the format of the load file for a create or an add](#reg-0888))
- <a id="reg-0823"></a>**REG-0823** For a create, an add or a change, the Register checks the full legal entity, with its brands and software products. For example, an add of an ACTIVE software product to an INACTIVE brand gives 409. ([Judgement: the format of the load file for a create or an add](#reg-0888))

## Visibility of a change

- <a id="reg-0824"></a>**REG-0824** A create, an add and a change obey [the propagation time of a status change](non-functional-requirements.md#reg-0001) on the Register APIs. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="reg-0825"></a>**REG-0825** After a create or an add, a new data holder brand is on Get Data Holder Brands Summary. ([CDS: Get Data Holder Brands Summary][cds-get-dh-brands-summary])
- <a id="reg-0826"></a>**REG-0826** After a create or an add, a new data holder brand with a `dataHolderBrandId` is on Get Data Holder Brands. ([CDS: Register APIs][cds-register-apis])
- <a id="reg-0827"></a>**REG-0827** After a create, a new data holder legal entity is on Get Data Holder Statuses. ([CDS: Register APIs][cds-register-apis])
- <a id="reg-0828"></a>**REG-0828** After a create, a new data recipient is on Get Data Recipients and Get Data Recipients Statuses. ([CDS: Register APIs][cds-register-apis])
- <a id="reg-0829"></a>**REG-0829** After a create or an add, each new software product is on Get Software Products Statuses. ([CDS: Register APIs][cds-register-apis])
- <a id="reg-0830"></a>**REG-0830** After a change of a software product, the next SSA of the software product has the changed metadata. ([CDS: Get SSA][cds-get-ssa])
- <a id="reg-0831"></a>**REG-0831** After a change of the JWKS URI of a software product, the token endpoint uses the changed JWKS URI for the next client assertion. ([CDS: Register APIs][cds-register-apis])
- <a id="reg-0832"></a>**REG-0832** A create, an add and a change set [the `lastUpdated` of the legal entity](register-apis.md#reg-0339). ([CDS: Register API schemas][cds-register-schemas])
- <a id="reg-0833"></a>**REG-0833** A create, an add and a change change [the `ETag`](register-apis.md#reg-0328) of each API whose response data changes. ([RFC 9110, section 8.8.3][rfc9110-etag])

## Client certificates

- <a id="reg-0834"></a>**REG-0834** The lab operator can register a client certificate for a software product, and remove a registered client certificate. ([ACCC Developer Portal: Token Endpoint][accc-token])
- <a id="reg-0835"></a>**REG-0835** The request body of a registration is one PEM certificate. When the body is not one PEM certificate, the response is 400. ([RFC 7468][rfc7468])
- <a id="reg-0836"></a>**REG-0836** The Register keeps the SHA-256 thumbprint of a registered certificate, not the certificate. ([RFC 8705, section 3.1][rfc8705-x5t]; [Judgement: only the thumbprint of a registered certificate](#reg-0882))
- <a id="reg-0837"></a>**REG-0837** The Register does not check the issuer of a certificate at its registration. ([Judgement: only the thumbprint of a registered certificate](#reg-0882))
- <a id="reg-0838"></a>**REG-0838** When the certificate is already registered for the software product, a registration changes nothing and gives 204. ([Judgement: a registration and a removal that are idempotent](#reg-0885))
- <a id="reg-0839"></a>**REG-0839** When the software product is REMOVED, a registration gives 409. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: no certificate for a REMOVED record](certificate-management.md#reg-0642))
- <a id="reg-0840"></a>**REG-0840** The path of a removal gives the thumbprint of the certificate in the base64url format of the RFC 8705 `x5t#S256` value. ([RFC 8705, section 3.1][rfc8705-x5t]; [Judgement: a removal as a `DELETE` on the thumbprint](#reg-0884))
- <a id="reg-0841"></a>**REG-0841** When the thumbprint is not a base64url SHA-256 value, a removal gives 400. ([RFC 8705, section 3.1][rfc8705-x5t])
- <a id="reg-0842"></a>**REG-0842** When the thumbprint is not registered for the software product, a removal changes nothing and gives 204. ([RFC 9110, section 9.2.2][rfc9110-idempotent]; [Judgement: a registration and a removal that are idempotent](#reg-0885))
- <a id="reg-0843"></a>**REG-0843** The lab operator can remove a client certificate in each status of the software product, also REMOVED. ([Judgement: a removal in each status](#reg-0886))
- <a id="reg-0844"></a>**REG-0844** After a removal, the token endpoint refuses the certificate. ([ACCC Developer Portal: Token Endpoint][accc-token]; [Judgement: a removal takes effect at the next token request](#reg-0887))
- <a id="reg-0845"></a>**REG-0845** An access token that the Register issued before a removal stays valid until it expires. ([Judgement: a removal takes effect at the next token request](#reg-0887))

## Key sets

- <a id="reg-0846"></a>**REG-0846** The lab operator can publish the next key of a key set, and then activate it. The [two steps of a key change](authorisation-server.md#reg-0468) apply. ([Judgement: the steps of a key change](authorisation-server.md#reg-0482))
- <a id="reg-0847"></a>**REG-0847** When the key set has a next key, a publish gives 409. ([RFC 9110, section 15.5.10][rfc9110-409])
- <a id="reg-0848"></a>**REG-0848** When the key set has no next key, an activation gives 409. ([RFC 9110, section 15.5.10][rfc9110-409])
- <a id="reg-0849"></a>**REG-0849** When the key set is not `access-token` or `ssa`, the response is 404. ([RFC 9110, section 15.5.5][rfc9110-404])

## Listener and access

The operator interface obeys the [operator interface library](../../../pkg/platform/operatorhttp/DESIGN.md): [each request has the operator token](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0001), [a `Host` allowlist with no default](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0005) and [no cross-origin requests](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0007).

- <a id="reg-0850"></a>**REG-0850** The operator interface is on a listener that is separate from the CDR APIs. ([Judgement: a lab-only operator interface](#reg-0880))
- <a id="reg-0851"></a>**REG-0851** The listener of the operator interface uses plain HTTP, with no TLS. ([Judgement: plain HTTP for the operator interface](#reg-0883))

## Bodies and errors

- <a id="reg-0852"></a>**REG-0852** When a request body is not one JSON object, the response is 400. For example, a body with [a JSON value after its first JSON value](../../../pkg/platform/jsonhttp/DESIGN.md#body-0002) gets 400. ([RFC 8259, section 2][rfc8259-2]; [Judgement: 400 or 409 for a refused body](#reg-0894))
- <a id="reg-0853"></a>**REG-0853** When a request body has [a field that the action does not accept](../../../pkg/platform/jsonhttp/DESIGN.md#body-0001), the response is 400. ([Judgement: 400 or 409 for a refused body](#reg-0894))
- <a id="reg-0854"></a>**REG-0854** When a request body is larger than 1 MiB, the response is 400. ([Judgement: a body limit of 1 MiB](#reg-0895))
- <a id="reg-0855"></a>**REG-0855** When a request body does not obey the participant model, the response is 400. For example, a URI with the `http` scheme gets 400. ([Judgement: 400 or 409 for a refused body](#reg-0894))
- <a id="reg-0856"></a>**REG-0856** When a request body obeys the participant model but breaks a status rule, the response is 409. For example, an ACTIVE software product of an INACTIVE brand gets 409. ([RFC 9110, section 15.5.10][rfc9110-409]; [Judgement: 400 or 409 for a refused body](#reg-0894))
- <a id="reg-0857"></a>**REG-0857** When a request body is not valid, the `detail` names each record and rule that fails. ([RFC 9457, section 3.1.4][rfc9457-detail])
- <a id="reg-0858"></a>**REG-0858** When an identifier in the body is the identifier of a different record, the response is 409. This applies also when that record is REMOVED or is a different kind of participant. ([an identifier is unique across all participant records](participants.md#reg-0221); [Judgement: 400 or 409 for a refused body](#reg-0894))
- <a id="reg-0859"></a>**REG-0859** When an identifier in the body is in use, the `detail` gives the identifier. ([RFC 9457, section 3.1.4][rfc9457-detail])
- <a id="reg-0860"></a>**REG-0860** When the status rules do not permit a status action, the response is 409, and the `detail` names the rule. ([RFC 9110, section 15.5.10][rfc9110-409])
- <a id="reg-0861"></a>**REG-0861** When the record of an action does not exist, the response is 404. ([RFC 9110, section 15.5.5][rfc9110-404])
- <a id="reg-0862"></a>**REG-0862** An error response is a [problem details](../../../docs/spec/glossary.md#problem-details) body with [the type `about:blank`](../../../pkg/platform/problem/DESIGN.md#prob-0001). ([RFC 9457][rfc9457]; [Judgement: problem details for the operator interface](#reg-0881))

## Operator interface paths

| ID | Action | Method and path | Success | Source |
|---|---|---|---|---|
| <a id="reg-0863"></a>**REG-0863** | Suspend, revoke, surrender or reinstate a data recipient | `POST /v1/data-recipients/{legalEntityId}/{suspend,revoke,surrender,reinstate}` | 204 | [Judgement: a lab-only operator interface](#reg-0880) |
| <a id="reg-0864"></a>**REG-0864** | Activate, deactivate or remove a software product | `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/{activate,deactivate,remove}` | 204 | [Judgement: a lab-only operator interface](#reg-0880) |
| <a id="reg-0865"></a>**REG-0865** | Register a client certificate for a software product | `POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates` | 204 | [Judgement: a lab-only operator interface](#reg-0880) |
| <a id="reg-0866"></a>**REG-0866** | Remove a client certificate of a software product | `DELETE /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/{thumbprint}` | 204 | [Judgement: a removal as a `DELETE` on the thumbprint](#reg-0884) |
| <a id="reg-0867"></a>**REG-0867** | Create a data recipient | `POST /v1/data-recipients` | 201 | [Judgement: the response of a create or an add](#reg-0891) |
| <a id="reg-0868"></a>**REG-0868** | Change the details of a data recipient | `PUT /v1/data-recipients/{legalEntityId}` | 204 | [Judgement: a change replaces each field of the record](#reg-0892) |
| <a id="reg-0869"></a>**REG-0869** | Add a brand to a data recipient | `POST /v1/data-recipients/{legalEntityId}/brands` | 201 | [Judgement: the response of a create or an add](#reg-0891) |
| <a id="reg-0870"></a>**REG-0870** | Change the details of a data recipient brand | `PUT /v1/data-recipients/{legalEntityId}/brands/{brandId}` | 204 | [Judgement: a change replaces each field of the record](#reg-0892) |
| <a id="reg-0871"></a>**REG-0871** | Activate, deactivate or remove a data recipient brand | `POST /v1/data-recipients/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 | [Judgement: a lab-only operator interface](#reg-0880) |
| <a id="reg-0872"></a>**REG-0872** | Add a software product to a data recipient brand | `POST /v1/data-recipients/{legalEntityId}/brands/{brandId}/software-products` | 201 | [Judgement: the response of a create or an add](#reg-0891) |
| <a id="reg-0873"></a>**REG-0873** | Change the details of a software product | `PUT /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}` | 204 | [Judgement: a change replaces each field of the record](#reg-0892) |
| <a id="reg-0874"></a>**REG-0874** | Create a data holder | `POST /v1/data-holders` | 201 | [Judgement: the response of a create or an add](#reg-0891) |
| <a id="reg-0875"></a>**REG-0875** | Change the details of a data holder | `PUT /v1/data-holders/{legalEntityId}` | 204 | [Judgement: a change replaces each field of the record](#reg-0892) |
| <a id="reg-0876"></a>**REG-0876** | Remove a data holder | `POST /v1/data-holders/{legalEntityId}/remove` | 204 | [Judgement: a lab-only operator interface](#reg-0880) |
| <a id="reg-0877"></a>**REG-0877** | Add a brand to a data holder | `POST /v1/data-holders/{legalEntityId}/brands` | 201 | [Judgement: the response of a create or an add](#reg-0891) |
| <a id="reg-0878"></a>**REG-0878** | Change the details of a data holder brand | `PUT /v1/data-holders/{legalEntityId}/brands/{brandId}` | 204 | [Judgement: a change replaces each field of the record](#reg-0892) |
| <a id="reg-0879"></a>**REG-0879** | Activate, deactivate or remove a data holder brand | `POST /v1/data-holders/{legalEntityId}/brands/{brandId}/{activate,deactivate,remove}` | 204 | [Judgement: a lab-only operator interface](#reg-0880) |
| <a id="reg-0896"></a>**REG-0896** | Publish the next key of a key set, or activate it. `{keySet}` is `access-token` or `ssa`. | `POST /v1/key-sets/{keySet}/{publish-next,activate-next}` | 204 | [Judgement: a lab-only operator interface](#reg-0880) |

## Judgements

### <a id="reg-0880"></a>REG-0880 A lab-only operator interface

- Kind: choice
- Basis: Step 3.3. The CDR Registrar keeps the participant data and statuses on the Register ([CDS: Participant Statuses][cds-participant-statuses]). The interface of the production Register for this is not public.
- Reason: The lab needs a way to change participants that a script can use. A separate listener keeps the lab-only actions out of the CDR APIs.

### <a id="reg-0881"></a>REG-0881 Problem details for the operator interface

- Kind: choice
- Basis: Step 3.2. The CDS error codes are for the CDR APIs ([CDS: Error Codes][cds-error-codes]). [RFC 9457][rfc9457] is the HTTP standard for error details.
- Reason: The operator interface is not a CDR API, so the HTTP standard applies.

### <a id="reg-0882"></a>REG-0882 Only the thumbprint of a registered certificate

- Kind: choice
- Basis: Step 3.2. The token endpoint compares the client certificate through its SHA-256 thumbprint, which is also the `x5t#S256` value of [RFC 8705, section 3.1][rfc8705-x5t].
- Reason: The thumbprint identifies the full certificate, so the Register needs no copy. The mTLS listener accepts only certificates from the mock CDR CA, so an issuer check at the registration adds nothing.

### <a id="reg-0883"></a>REG-0883 Plain HTTP for the operator interface

- Kind: choice
- Basis: Step 3.3. The CDS transport rules apply to the CDR APIs ([CDS: Transaction Security][cds-transaction-security]). The operator interface is not a CDR API.
- Reason: The operator interface is not a CDR API, and it is lab-only.

### <a id="reg-0884"></a>REG-0884 A removal as a `DELETE` on the thumbprint

- Kind: choice
- Basis: Step 3.2. A `DELETE` on the item is the usual HTTP form of a removal ([RFC 9110, section 9.3.5][rfc9110-delete]).
- Reason: The Register keeps only the thumbprint, so the thumbprint identifies the certificate.

### <a id="reg-0885"></a>REG-0885 A registration and a removal that are idempotent

- Kind: choice
- Basis: Step 3.2. `DELETE` is idempotent ([RFC 9110, section 9.2.2][rfc9110-idempotent]).
- Reason: A script can send a registration or a removal again with no error. A second registration and a removal of a thumbprint that is not registered then have the same result: 204 and no change.

### <a id="reg-0886"></a>REG-0886 A removal in each status

- Kind: choice
- Basis: Step 3.3. The CDS does not say if a removal of a certificate depends on the status of the software product.
- Reason: A removal only takes trust away, so it cannot break a status rule. A registration is refused for REMOVED, because the certificate can never be used.

### <a id="reg-0887"></a>REG-0887 A removal takes effect at the next token request

- Kind: choice
- Basis: Step 2. The token endpoint checks the registered certificates when it issues a token ([ACCC Developer Portal: Token Endpoint][accc-token]).
- Reason: An access token is valid for 5 minutes. A check of the certificates at each secure API call adds a database read for a short window.

### <a id="reg-0888"></a>REG-0888 The format of the load file for a create or an add

- Kind: choice
- Basis: Step 3.3. The CDS gives no API to create participants.
- Reason: One format for the load and the operator interface is the least complex. The lab operator can copy a record from the seed file, and the Register checks it with the same rules.

### <a id="reg-0889"></a>REG-0889 The lab operator gives the identifiers

- Kind: choice
- Basis: Step 3.3. In production, the Register issues the identifiers ([CDS: Register API schemas][cds-register-schemas]).
- Reason: The lab services get their identifiers in their configuration before they start, as with the load file. A GUID is the usual value.

### <a id="reg-0890"></a>REG-0890 A status in a create or an add

- Kind: choice
- Basis: Step 3.3. The CDS gives no API to create participants.
- Reason: The load permits each status, so a create and an add do too. A lab user can then make a record in each status to test a data holder or a data recipient.

### <a id="reg-0891"></a>REG-0891 The response of a create or an add

- Kind: choice
- Basis: Step 3.2. A `POST` that makes a resource gives 201 with a `Location` header ([RFC 9110, section 15.3.2][rfc9110-201]).
- Reason: The lab operator gives the identifiers, so a body adds nothing. The `Location` is also the path of a change of the record.

### <a id="reg-0892"></a>REG-0892 A change replaces each field of the record

- Kind: choice
- Basis: Step 3.2. The body of a `PUT` replaces the state of the resource ([RFC 9110, section 9.3.4][rfc9110-put]).
- Reason: Each field that a change excludes has its own action with its own rules. A merge patch ([RFC 7396][rfc7396]) needs rules for `null` and for lists, and is not simpler for a record of few fields.

### <a id="reg-0893"></a>REG-0893 Identifiers that do not change

- Kind: choice
- Basis: Step 3.3. The CDS does not say if an identifier can change.
- Reason: Participants keep the identifiers in their caches and their registrations, and the events name each record by its identifier. A new record is the way to get a new identifier.

### <a id="reg-0894"></a>REG-0894 400 or 409 for a refused body

- Kind: choice
- Basis: Step 3.2. A conflict with the current state of the target resource is 409 ([RFC 9110, section 15.5.10][rfc9110-409]).
- Reason: A body that is not valid alone is a client error (400). A body that is valid alone but does not agree with the current state of the Register is a conflict. The status rules already give 409.

### <a id="reg-0895"></a>REG-0895 A body limit of 1 MiB

- Kind: choice
- Basis: Step 3.2. [RFC 9110, section 15.5.14][rfc9110-413] gives 413 for content that is larger than the server accepts.
- Reason: A legal entity in the seed data is less than 20 KiB, and a limit stops a body that fills the memory. 400 is the status of each other body that the operator interface refuses, so a script needs one status for a refused body.

[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-apis
[cds-get-dh-brands-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands-summary
[cds-get-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-statement-assertion-ssa
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[rfc7396]: https://www.rfc-editor.org/rfc/rfc7396
[rfc7468]: https://www.rfc-editor.org/rfc/rfc7468
[rfc8259-2]: https://www.rfc-editor.org/rfc/rfc8259#section-2
[rfc8705-x5t]: https://www.rfc-editor.org/rfc/rfc8705#section-3.1
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-404]: https://www.rfc-editor.org/rfc/rfc9110#name-404-not-found
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#name-409-conflict
[rfc9110-413]: https://www.rfc-editor.org/rfc/rfc9110#name-413-content-too-large
[rfc9110-delete]: https://www.rfc-editor.org/rfc/rfc9110#name-delete
[rfc9110-etag]: https://www.rfc-editor.org/rfc/rfc9110#name-etag
[rfc9110-idempotent]: https://www.rfc-editor.org/rfc/rfc9110#name-idempotent-methods
[rfc9110-put]: https://www.rfc-editor.org/rfc/rfc9110#name-put
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[rfc9457-detail]: https://www.rfc-editor.org/rfc/rfc9457#section-3.1.4
