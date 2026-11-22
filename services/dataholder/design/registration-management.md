# Mock data holder: Registration management

This area gives the management of a registration, the participant statuses that the data holder gets from the Register, and the reaction of the data holder to them. [Registration](registration.md) gives the registration request and its validation. The [authorisation server library](../../../pkg/authserver/design/README.md) gives the reaction at the PAR, authorisation and token endpoints.

## Registration management

- <a id="dh-0657"></a>**DH-0657** `GET`, `PUT` and `DELETE` at the `registration_endpoint` and `/{ClientId}` read, replace and delete a registration. ([CDS: Registration Management][cds-registration-management])
- <a id="dh-0658"></a>**DH-0658** Each registration management request has an access token as a bearer token. ([RFC 6750][rfc6750]; [CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints])
- <a id="dh-0659"></a>**DH-0659** The access token is active, has the scope `cdr:registration`, and is bound to the client certificate of the request. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]; [CDS: Holder of Key Mechanism][cds-hok])
- <a id="dh-0660"></a>**DH-0660** The access token is a token of the client in the path. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints])
- <a id="dh-0661"></a>**DH-0661** When the data holder refuses a registration management request, the response has no body. ([`cds_dcr.json`][cds-openapi-dcr]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711))

The data holder gives these refusals:

| ID | Condition | Response | Source |
|---|---|---|---|
| <a id="dh-0662"></a>**DH-0662** | The request has no bearer token | 401 with `WWW-Authenticate: Bearer` | [RFC 6750, section 3.1][rfc6750-errors]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0663"></a>**DH-0663** | The token is not known, is expired, is revoked, is not bound to the client certificate, or is a token of a deleted registration | 401 with `WWW-Authenticate: Bearer error="invalid_token"` | [RFC 6750, section 3.1][rfc6750-errors]; [RFC 8705, section 3][rfc8705-binding]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0664"></a>**DH-0664** | The path has a client ID that does not exist | 401 with `WWW-Authenticate: Bearer error="invalid_token"` | [`cds_dcr.json`][cds-openapi-dcr]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0665"></a>**DH-0665** | The token does not have the scope `cdr:registration` | 403 with `WWW-Authenticate: Bearer error="insufficient_scope", scope="cdr:registration"` | [RFC 6750, section 3.1][rfc6750-errors]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |
| <a id="dh-0666"></a>**DH-0666** | The path has the client ID of a different registration | 403 | [`cds_dcr.json`][cds-openapi-dcr]; [Judgement: 401 or 403 with no body for registration management errors](#dh-0711) |

- <a id="dh-0667"></a>**DH-0667** `GET` gives 200 with the registration, in the format of the registration response. ([CDS: RegistrationProperties][cds-schema-registration-properties])
- <a id="dh-0668"></a>**DH-0668** The `software_statement` of a `GET` response is the SSA of the last request that the data holder accepted. ([CDS: RegistrationProperties][cds-schema-registration-properties])
- <a id="dh-0669"></a>**DH-0669** A `PUT` has the same body as a registration request, with a new SSA. The same validation rules and errors apply. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0670"></a>**DH-0670** When the `software_id` of the SSA of a `PUT` is not that of the registration, the response is 400 with `invalid_software_statement`. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0671"></a>**DH-0671** A `PUT` replaces all values of the registration with the values that the request and its SSA give. The client ID and `client_id_issued_at` do not change. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0672"></a>**DH-0672** A successful `PUT` gives 200 with the registration. ([CDS: PUT Operations][cds-put-operations])
- <a id="dh-0673"></a>**DH-0673** A `DELETE` deletes the registration and gives 204. ([CDS: DCR APIs][cds-dcr-apis]; [Judgement: support for the deletion of a registration](#dh-0712))
- <a id="dh-0674"></a>**DH-0674** After a `DELETE`, client authentication with the client ID of the registration fails. ([CDS: DCR APIs][cds-dcr-apis])
- <a id="dh-0675"></a>**DH-0675** After a `DELETE`, the software product can register again. It then gets a new client ID. ([Judgement: support for the deletion of a registration](#dh-0712))
- <a id="dh-0676"></a>**DH-0676** When a software product deletes its registration, each arrangement of the client that did not end, ends at the time of the request. The end is a [revocation by the data recipient](end-of-an-arrangement.md#dh-0902), with no notice to the data recipient. ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]; [Judgement: a deletion ends the arrangements of the client as a revocation](#dh-0713))
- <a id="dh-0677"></a>**DH-0677** After a `DELETE`, the data holder accepts no token of the client. ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints])

## Participant statuses

- <a id="dh-0679"></a>**DH-0679** The data holder gets the participant statuses of the software products from [Get Software Products Statuses][cds-get-sp-statuses] of the Register, with the `{industry}` `all`. ([CDS: Metadata Cache Management][cds-metadata-cache]; [Judgement: only the statuses of software products](#dh-0714))
- <a id="dh-0680"></a>**DH-0680** The data holder gets the base URI of the public APIs of the Register from its configuration. For example, the setting `REGISTER_PUBLIC_BASE_URL` gives the base URI. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0681"></a>**DH-0681** For the connection to the public APIs of the Register, the data holder trusts only the public CAs in its configuration. ([CDS: Transaction Security][cds-transaction-security])
- <a id="dh-0682"></a>**DH-0682** The software product of a registration is the software product with the `software_id` of the registration. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="dh-0683"></a>**DH-0683** The data holder gets the statuses at an interval from its configuration. The interval is 4 minutes at most. For example, the setting `REGISTER_STATUS_INTERVAL` is `1m`. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0684"></a>**DH-0684** A status change on the Register has its effect at the data holder within 5 minutes. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0685"></a>**DH-0685** The request for the statuses has `If-None-Match` with the `ETag` of the last response. ([CDS: Metadata Cache Management][cds-metadata-cache])
- <a id="dh-0686"></a>**DH-0686** When the Register does not respond, the data holder keeps the previous status of each software product, with no time limit. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="dh-0687"></a>**DH-0687** When the response has no status, or a status that is not known, for a software product, the data holder keeps the previous status of that software product. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="dh-0688"></a>**DH-0688** When the data holder has no status for a software product, the software product gets the responses for INACTIVE. ([Judgement: INACTIVE for a software product with no status](#dh-0715))
- <a id="dh-0689"></a>**DH-0689** The readiness probe fails until the data holder gets the statuses one time. ([the readiness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0002); [Judgement: INACTIVE for a software product with no status](#dh-0715))
- <a id="dh-0690"></a>**DH-0690** When the data holder gets the status REMOVED for a software product, it ends each arrangement of each registration of that software product at once. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]; [Judgement: the clean-up at once when a software product is REMOVED](#dh-0716))
- <a id="dh-0691"></a>**DH-0691** When the data holder gets the status REMOVED for a software product, it deletes each registration of that software product. Client authentication with the client ID then fails, and the data holder accepts no token of the client. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]; [Judgement: the clean-up at once when a software product is REMOVED](#dh-0716))
- <a id="dh-0692"></a>**DH-0692** The data holder sends no notice to the data recipient for an arrangement that a REMOVED status ends. ([CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed])
- <a id="dh-0693"></a>**DH-0693** When the status of the software product of the SSA is REMOVED, the data holder refuses a registration request with 400 and the error `unapproved_software_statement`. For example, an SSA that the Register issued before the status REMOVED cannot register the software product again. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]; [Judgement: only the status REMOVED stops a registration](registration.md#dh-0705))

## Reaction to participant statuses

The authorisation server library gives the reaction at [the PAR endpoint](../../../pkg/authserver/design/authorisation-request.md#auth-0110), [the authorisation endpoint](../../../pkg/authserver/design/authorisation-request.md#auth-0111) and [the token endpoint](../../../pkg/authserver/design/authorisation-request.md#auth-0112), the [endpoints that the status does not change](../../../pkg/authserver/design/authorisation-request.md#auth-0113), and [the sequence of the status check at the PAR and token endpoints](../../../pkg/authserver/design/authorisation-request.md#auth-0114).

- <a id="dh-0694"></a>**DH-0694** The participant status of the data recipient does not change the reaction of the data holder. Only the status of the software product changes it. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: only the statuses of software products](#dh-0714))

| ID | Request or item | ACTIVE | INACTIVE | Source |
|---|---|---|---|---|
| <a id="dh-0695"></a>**DH-0695** | APIs on the resource base URI | Permitted | 403 with `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`, and the `detail` is the status of the software product | [CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: unauthorized_client, not invalid_grant, for a software product that is not ACTIVE](../../../pkg/authserver/design/authorisation-request.md#auth-0124) |
| <a id="dh-0696"></a>**DH-0696** | UserInfo endpoint | Permitted | 401 with `WWW-Authenticate: Bearer error="invalid_token"` and no body | [CDS: Data Holder Responsibilities][cds-dh-responsibilities]; [Judgement: unauthorized_client, not invalid_grant, for a software product that is not ACTIVE](../../../pkg/authserver/design/authorisation-request.md#auth-0124) |
| <a id="dh-0697"></a>**DH-0697** | Registration management | Permitted | Permitted | [CDS: Data Holder Responsibilities][cds-dh-responsibilities] |
| <a id="dh-0698"></a>**DH-0698** | Arrangements | No change | No change. Sharing starts again when the software product is ACTIVE again. | [CDS: Data Holder Responsibilities][cds-dh-responsibilities] |

- <a id="dh-0699"></a>**DH-0699** On the resource base URI and the UserInfo endpoint, the status check comes after the check of the access token, and before the other checks of the request. ([CDR Support: Error scenarios and responses][support-error-scenarios])

## Judgements

### <a id="dh-0711"></a>DH-0711 401 or 403 with no body for registration management errors

- Kind: choice
- Basis: Step 3.2. The DCR OpenAPI specification gives 401 for an "unknown or invalid Client or invalid access token" and 403 when the client "does not have permission", with no body schema ([`cds_dcr.json`][cds-openapi-dcr]). [RFC 6750, section 3.1][rfc6750-errors] gives the error values. A certificate that does not match gives `invalid_token` ([RFC 8705, section 3][rfc8705-binding]).
- Reason: The status codes follow the CDS specification, and the `WWW-Authenticate` values follow the RFCs that it uses. A response with no body agrees with the CDS specification, which has no body schema.

### <a id="dh-0712"></a>DH-0712 Support for the deletion of a registration

- Kind: choice
- Basis: Step 3.3. The CDS marks `DELETE` of a registration as optional ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]).
- Reason: A lab user and the conformance suite can then register a software product again with no reset of the lab.

### <a id="dh-0713"></a>DH-0713 A deletion ends the arrangements of the client as a revocation

- Kind: choice
- Basis: Step 2. The CDS gives no rule for the arrangements of a registration that the software product deletes ([CDS: Dynamic Client Registration endpoints][cds-dcr-endpoints]). ACCC guidance says: "After an ADR registration deletion request there should be no active consents for that ADR" ([CDR Support: Guidance on Dynamic Client Registration][support-dcr]).
- Reason: The software product asks for the deletion, so the end has the reason and the effects of a revocation by the data recipient. The data holder sends no notice, because the data recipient made the request. The end at the time of the request agrees with a revocation at the CDR arrangement revocation endpoint.

### <a id="dh-0714"></a>DH-0714 Only the statuses of software products

- Kind: choice
- Basis: Step 3.3. The CDS says that the data holder "SHOULD" use the statuses of the data recipients and of the software products ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). The reactions are the same for each data recipient status with the same software product status, because the Register changes the status of each software product when its data recipient changes ([CDS: Status Mapping][cds-status-mapping]).
- Reason: One API gives the same behaviour with one cache.

### <a id="dh-0715"></a>DH-0715 INACTIVE for a software product with no status

- Kind: choice
- Basis: Step 3.1. The CDS says to keep the previous status, and gives no rule for a software product with no previous status ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]).
- Reason: Data must not go to a software product whose status is not known. INACTIVE stops the sharing and changes no arrangement. A registration needs an SSA from the Register, so each registered software product has a status after the next request to the Register.

### <a id="dh-0716"></a>DH-0716 The clean-up at once when a software product is REMOVED

- Kind: choice
- Basis: Step 3.3. The CDS permits a batch, for example each night ([CDS: Data Holder Responsibilities][cds-dh-responsibilities]). The ACCC also permits it each time the Register cache is refreshed ([CDR Support: Status of consent when an ADR Software Product is "Removed"][support-removed]).
- Reason: One step at the time of the status is less complex than a separate batch. A lab test then does not wait for a batch.

[cds-registration-management]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-management
[cds-put-operations]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#put-operations
[cds-schema-registration-properties]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-dynamic-client-registration-api_schemas_tocSregistrationproperties
[cds-dcr-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dynamic-client-registration-endpoints
[cds-dcr-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#dcr-apis
[cds-hok]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#holder-of-key-mechanism
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-get-sp-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-products-statuses
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-openapi-dcr]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_dcr.json
[support-dcr]: https://cdr-support.zendesk.com/hc/en-us/articles/6828251436431
[support-removed]: https://cdr-support.zendesk.com/hc/en-us/articles/900001657906
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[rfc6750]: https://www.rfc-editor.org/rfc/rfc6750
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc8705-binding]: https://www.rfc-editor.org/rfc/rfc8705#section-3
