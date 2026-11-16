# Mock CDR Register: Participants

This area gives the participant model of the Register: the participants, their fields and their statuses, the status rules and the cascades of a status change.

## Participants

Each row gives one kind of participant record, its parent and its status values.

| ID | Record | Parent | Status values | Source |
|---|---|---|---|---|
| <a id="reg-0200"></a>**REG-0200** | Data holder legal entity | None | ACTIVE, REMOVED | [CDS: Register API schemas][cds-register-schemas]; [CDS: CDR Federation][cds-federation] |
| <a id="reg-0201"></a>**REG-0201** | Data holder brand | Data holder legal entity | ACTIVE, INACTIVE, REMOVED | [CDS: Register API schemas][cds-register-schemas]; [CDS: CDR Federation][cds-federation] |
| <a id="reg-0202"></a>**REG-0202** | Data recipient legal entity | None | ACTIVE, SUSPENDED, REVOKED, SURRENDERED | [CDS: Register API schemas][cds-register-schemas]; [CDS: CDR Federation][cds-federation] |
| <a id="reg-0203"></a>**REG-0203** | Data recipient brand | Data recipient legal entity | ACTIVE, INACTIVE, REMOVED | [CDS: Register API schemas][cds-register-schemas]; [CDS: CDR Federation][cds-federation] |
| <a id="reg-0204"></a>**REG-0204** | Software product | Data recipient brand | ACTIVE, INACTIVE, REMOVED | [CDS: Register API schemas][cds-register-schemas]; [CDS: CDR Federation][cds-federation] |

## Fields

The Register checks the fields of a participant record each time that it gets one, for example from the load or from the [operator interface](../DESIGN.md#operator-interface).

- <a id="reg-0205"></a>**REG-0205** The name of a data holder legal entity has 36 characters or fewer. When a name is longer, the Register refuses it. For example, the load refuses a data holder with a name of 37 characters. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: a data holder name of 36 characters at most](#reg-0260))
- <a id="reg-0206"></a>**REG-0206** A data holder brand has one or more industries. Each industry is one of `banking`, `energy`, `non-bank-lending` and `telco`. ([CDS: Get Data Holder Brands][cds-get-dh-brands])
- <a id="reg-0207"></a>**REG-0207** A data holder brand has its endpoints, its JWKS endpoint and an optional brand group. The endpoints are the public, product, resource, infosec, admin and extension base URIs and the website URI. ([CDS: Participant Endpoints][cds-participant-endpoints]; [CDS: Get Data Holder Brands][cds-get-dh-brands])
- <a id="reg-0208"></a>**REG-0208** A data holder brand has a `dataHolderBrandId`, an `interimId` or both. ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary])
- <a id="reg-0209"></a>**REG-0209** A data holder brand that shares only product data has no `dataHolderBrandId`. ([ACCC Developer Portal: Get Data Holder Brands Summary][accc-brands-summary])
- <a id="reg-0210"></a>**REG-0210** A data holder brand with no `dataHolderBrandId` needs only its public details: name, logo, industries, status and public base URI. For example, the load accepts such a brand with no resource base URI. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: only the public details for a brand of product data](#reg-0261))
- <a id="reg-0211"></a>**REG-0211** A data recipient legal entity has an accreditation number and an accreditation level. The level is `UNRESTRICTED` or `SPONSORED`. ([CDS: Get Data Recipients][cds-get-drs])
- <a id="reg-0212"></a>**REG-0212** A software product has the metadata that its SSA contains. ([CDS: SSA Definition][cds-ssa-definition])
  - name and description
  - client URI and logo URI
  - redirect URIs and sector identifier URI
  - terms of service URI and policy URI
  - JWKS URI, revocation URI and recipient base URI
  - scopes
- <a id="reg-0213"></a>**REG-0213** A software product has a list of the client certificates that the Register trusts for it. ([ACCC Developer Portal: Token Endpoint][accc-token])
- <a id="reg-0214"></a>**REG-0214** When a CDS Register schema gives a text field a `maxLength`, the field has that limit in characters. The Register refuses a longer value. ([CDS: Register API schemas][cds-register-schemas])
- <a id="reg-0215"></a>**REG-0215** Each URI field of a participant is an absolute URI with the `https` scheme and a host. The Register refuses a value that is not. ([CDS: Common field types][cds-common-field-types]; [CDS: RegistrationProperties schema][cds-dcr-schemas]; [Judgement: only `https` in a URI field](#reg-0262))
  - each logo URI
  - each endpoint and the JWKS endpoint of a data holder brand
  - each URI of a software product, also each redirect URI
- <a id="reg-0216"></a>**REG-0216** Each URI field of a participant has 1000 characters or fewer. The Register refuses a longer value. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: 1000 characters for each URI field](#reg-0263))
- <a id="reg-0217"></a>**REG-0217** A URI field has only the characters of a URI. The Register refuses a value with a different character, for example a space. ([CDS: Common field types][cds-common-field-types]; [RFC 3986, section 2][rfc3986-2])
- <a id="reg-0218"></a>**REG-0218** An optional URI field with no value is valid. For example, a software product with no policy URI is valid. ([CDS: SSA Definition][cds-ssa-definition])

## Identifiers

- <a id="reg-0219"></a>**REG-0219** Each identifier of a participant record is a string of 36 characters or fewer. ([CDS: Register API schemas][cds-register-schemas])
- <a id="reg-0220"></a>**REG-0220** The Register issues each identifier as a GUID. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: GUIDs for identifiers](#reg-0264))
- <a id="reg-0221"></a>**REG-0221** Each identifier is unique across all participant records, not only across the records of one kind. For example, no software product has the identifier of a data holder brand. ([Judgement: one identifier for each record](#reg-0265))

## Status transitions

- <a id="reg-0222"></a>**REG-0222** A data recipient legal entity changes status only with these transitions. ([CDS: Participant Statuses][cds-participant-statuses])
  - ACTIVE to SUSPENDED, REVOKED or SURRENDERED
  - SUSPENDED to ACTIVE, REVOKED or SURRENDERED
  - REVOKED to ACTIVE
  - SURRENDERED is final
- <a id="reg-0223"></a>**REG-0223** A software product changes status only with these transitions. ([CDS: Participant Statuses][cds-participant-statuses])
  - ACTIVE to INACTIVE or REMOVED
  - INACTIVE to ACTIVE or REMOVED
  - REMOVED is final
- <a id="reg-0224"></a>**REG-0224** A data holder legal entity changes status only from ACTIVE to REMOVED. REMOVED is final. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: the transitions of a data holder legal entity](#reg-0266))
- <a id="reg-0225"></a>**REG-0225** A data holder brand and a data recipient brand change status only with the transitions of a software product. ([CDS: Participant Statuses][cds-participant-statuses]; [Judgement: the transitions of a brand](#reg-0267))
- <a id="reg-0226"></a>**REG-0226** The Register never returns a status value that is not in the [participant records](#reg-0200). ([CDS: Participant Statuses][cds-participant-statuses])

## Cascades

- <a id="reg-0227"></a>**REG-0227** When a data recipient legal entity becomes SUSPENDED, each of its ACTIVE software products becomes INACTIVE at the same time. ([CDS: Status Mapping][cds-status-mapping])
- <a id="reg-0228"></a>**REG-0228** When a data recipient legal entity becomes REVOKED or SURRENDERED, each of its software products that is not REMOVED becomes REMOVED at the same time. ([CDS: Status Mapping][cds-status-mapping])
- <a id="reg-0229"></a>**REG-0229** When a data recipient legal entity is SUSPENDED, REVOKED or SURRENDERED, none of its software products is ACTIVE. ([CDS: Status Mapping][cds-status-mapping])
- <a id="reg-0230"></a>**REG-0230** A software product can change status with no change to its data recipient. For example, the lab operator makes one software product INACTIVE, and its data recipient stays ACTIVE. ([CDS: Data Holder Responsibilities][cds-dh-responsibilities])
- <a id="reg-0231"></a>**REG-0231** When a data holder legal entity becomes REMOVED, each of its brands that is not REMOVED becomes REMOVED at the same time. ([Judgement: the cascade of a brand status](#reg-0268))
- <a id="reg-0232"></a>**REG-0232** When a data recipient brand becomes INACTIVE, each of its ACTIVE software products becomes INACTIVE at the same time. ([CDS: Status Mapping][cds-status-mapping]; [Judgement: the cascade of a brand status](#reg-0268))
- <a id="reg-0233"></a>**REG-0233** When a data recipient brand becomes REMOVED, each of its software products that is not REMOVED becomes REMOVED at the same time. ([CDS: Status Mapping][cds-status-mapping]; [Judgement: the cascade of a brand status](#reg-0268))
- <a id="reg-0234"></a>**REG-0234** When a data recipient brand becomes ACTIVE, its software products do not change. ([Judgement: the cascade of a brand status](#reg-0268))
- <a id="reg-0235"></a>**REG-0235** A record is never more active than its parent. ([Judgement: a record is not more active than its parent](#reg-0269))
  - A software product of an INACTIVE data recipient brand is INACTIVE or REMOVED.
  - A software product of a REMOVED data recipient brand is REMOVED.
  - A brand of a REMOVED data holder legal entity is REMOVED.

## Removal

- <a id="reg-0236"></a>**REG-0236** The Register never issues the identifier of a REMOVED software product again. ([CDR Support: Removed software product][support-removed-sp])
- <a id="reg-0237"></a>**REG-0237** The Register deletes no participant record. A removal is a change to the status REMOVED, and the identifier stays in use. ([Judgement: a removal is a change of status](#reg-0270))

## Judgements

### <a id="reg-0260"></a>REG-0260 A data holder name of 36 characters at most

- Kind: choice
- Basis: Step 1. The `LegalEntityDetail` schema of Get Data Holder Brands gives `maxLength: 36` ([CDS: Register API schemas][cds-register-schemas]). This is shorter than the limit of the other name fields, and is possibly a defect in the CDS.
- Reason: The Register cannot give a longer name in a valid response. Thus it refuses the name when it gets it.

### <a id="reg-0261"></a>REG-0261 Only the public details for a brand of product data

- Kind: choice
- Basis: Step 1. Get Data Holder Brands gives the secure endpoints of a brand, and requires `dataHolderBrandId` ([CDS: Register API schemas][cds-register-schemas]).
- Reason: A brand that shares only product data has no secure endpoints. Thus it needs no value for them.

### <a id="reg-0262"></a>REG-0262 Only `https` in a URI field

- Kind: choice
- Basis: Step 3.2. `URIString` permits each scheme ([CDS: Common field types][cds-common-field-types]). But all HTTP calls use HTTPS ([CDS: Transaction Security][cds-transaction-security]), a FAPI authorisation server refuses a redirect URI with a different scheme ([FAPI 1.0 Advanced, section 5.2.2][fapi-as]), and a sector identifier URI uses `https` ([OIDC Core, section 8.1][oidc-pairwise]).
- Reason: A data holder cannot use an endpoint or a registration with a different scheme, so the Register refuses the value when the lab operator gives it. A logo, a client URI and a policy are web pages for a consumer. One rule for each URI field is simpler.

### <a id="reg-0263"></a>REG-0263 1000 characters for each URI field

- Kind: choice
- Basis: Step 3.3. The Register schemas give `maxLength: 1000` to each logo URI and each endpoint ([CDS: Register API schemas][cds-register-schemas]). The RegistrationProperties schema gives no limit to the other URIs of a software product ([CDS: RegistrationProperties schema][cds-dcr-schemas]).
- Reason: One limit for each URI field is simpler. The URIs of the seed data are much shorter.

### <a id="reg-0264"></a>REG-0264 GUIDs for identifiers

- Kind: choice
- Basis: Step 3.3. The CDS gives a string of 36 characters or fewer, and no format ([CDS: Register API schemas][cds-register-schemas]).
- Reason: A GUID is unique with no coordination. It has 36 characters, so it is in the limit.

### <a id="reg-0265"></a>REG-0265 One identifier for each record

- Kind: choice
- Basis: Step 3.3. The CDS does not say if an identifier is unique across the kinds of records ([CDS: Register API schemas][cds-register-schemas]).
- Reason: GUIDs are unique in all cases, so the rule costs nothing. It prevents a mix-up between records.

### <a id="reg-0266"></a>REG-0266 The transitions of a data holder legal entity

- Kind: choice
- Basis: Step 3.3. The CDS gives no state diagram for data holders ([CDS: Participant Statuses][cds-participant-statuses]).
- Reason: The CDS diagrams make REMOVED final for a software product, so REMOVED has one meaning for each record. A data holder legal entity has no INACTIVE status, so its only change is to REMOVED. A REMOVED identifier is then never in use again, the same as for a software product ([CDR Support: Removed software product][support-removed-sp]).

### <a id="reg-0267"></a>REG-0267 The transitions of a brand

- Kind: choice
- Basis: Step 3.3. The CDS gives no state diagram for brands ([CDS: Participant Statuses][cds-participant-statuses]).
- Reason: A brand has the same three status values as a software product. The same transitions give the same value the same meaning for each record.

### <a id="reg-0268"></a>REG-0268 The cascade of a brand status

- Kind: choice
- Basis: Step 3.3. The CDS cascades only a data recipient status ([CDS: Status Mapping][cds-status-mapping]).
- Reason: The same cascade for a brand and for a data holder keeps the rule that a record is not more active than its parent. A status change of the brand alone breaks that rule.

### <a id="reg-0269"></a>REG-0269 A record is not more active than its parent

- Kind: choice
- Basis: Step 3.1. The CDS gives no rule for brands ([CDS: Participant Statuses][cds-participant-statuses]).
- Reason: A brand that is not active cannot share data. Thus its children cannot share data either.

### <a id="reg-0270"></a>REG-0270 A removal is a change of status

- Kind: choice
- Basis: Step 3.3. The Register never issues the identifier of a REMOVED software product again ([CDR Support: Removed software product][support-removed-sp]), and the events of each participant stay ([ADR 0003][adr-0003]).
- Reason: The same rule for each identifier is simpler. A record that stays also keeps its history.

[accc-brands-summary]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-dcr-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#tocSregistrationproperties
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-federation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-federation
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-get-drs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[fapi-as]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#authorization-server
[oidc-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[rfc3986-2]: https://www.rfc-editor.org/rfc/rfc3986#section-2
[support-removed-sp]: https://cdr-support.zendesk.com/hc/en-us/articles/900001657906
