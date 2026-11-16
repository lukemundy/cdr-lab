# Mock CDR Register: design

The [overview of the specification](design/README.md) gives the summary, the context, the areas, the goals and the non-goals. Each area file in `design/` gives the rules of its area.

## Operator console

### Functions

The console also does these actions of the operator interface, with the same rules ([the operator interface](design/operator-interface.md#reg-0800)):

- Create a data recipient or a data holder: the legal entity, with no brands. Add a brand to a legal entity, and add a software product to a data recipient brand. (See Decisions.)
- Change the details of a legal entity, a brand or a software product: the fields that a change in the operator interface replaces.

Rules for a create, an add and a change:

- Each create, add and change is a form on its own page, with its own URL. It is not a dialog. A URL of a form is never the URL of a record, so that a record with each ID, for example `new`, keeps its detail. (Judgement: the records have many fields. See Decisions.)
- The list of data recipients has the action "Create data recipient", and the list of data holders has "Create data holder". The detail of a participant has the actions "Change details" and "Add brand" for the legal entity, "Change brand" for each brand, "Add software product" for each data recipient brand, and "Change software product" for each software product. The console shows these actions for each record in each status. (See Decisions.) Each of these actions is a link directly after the heading of its record, and "Add software product" is at the heading "Software products" of its brand. The status actions and "Register certificate" are buttons after the details of their record and before its child records. ([Design system DESIGN.md][ui-design]) Each brand and each software product is a card that starts collapsed, with its status in its title bar. "Brands" and each "Software products" heading have "Show all" and "Hide all". After a form, the card of the record that the form changed or added is open. ([Design system DESIGN.md][ui-design])
- A create and an add give each field of one record in the format of the load file, with the identifiers and the status, and with no child records and no client certificates. ([Operator interface](design/operator-interface.md#reg-0800))
- For each identifier of a create or an add, the form shows a new random UUID ([RFC 9562, section 5.4][rfc9562-v4]) in a field that the operator can change. The operator can keep it, or give a known identifier, for example the brand ID in the configuration of a lab service. A data holder brand gets a UUID in its `dataHolderBrandId` field, and its `interimId` field is empty. (See Decisions.)
- The status of a create or an add is a field. Its values are the status values of the record of the [participant records](design/participants.md#reg-0200), and ACTIVE is selected first. The status rules apply: for example, a software product of an INACTIVE brand must be INACTIVE or REMOVED. (See Decisions.)
- A change form shows the current value of each field, and shows the identifiers of the record as text that the operator cannot change. The console sends each field of the change, as a change in the operator interface does: an optional field with no value has no value after the change. ([Operator interface](design/operator-interface.md#reg-0800))
- A change and an add send the version of the participant that the form loaded. A create has no version. A stale version gives 409, and the Register changes nothing (see the row "Action with a stale view" in Decisions).
- A choice of one value from a fixed set (the accreditation level, the status) is a native select. A choice from a fixed set of more than one value (the industries of a data holder brand) is a group of checkboxes. A list of values that the operator writes (the redirect URIs, the scopes) is a text area with one value on each line. The endpoints of a data holder brand are a group of fields with a legend. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; see Decisions)
- In a create, a select of a value that is required and that has no usual value (the accreditation level) has no selected value, so that the operator chooses it. (Judgement: a value that the console selects can go to the Register with no decision of the operator.)
- The label of each field that the console API schema does not require says "(optional)". The help text of the identifier fields of a data holder brand says that a brand with no `dataHolderBrandId` needs only its public details. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2]; [GOV.UK Design System: Question pages][govuk-question-pages] marks the optional fields, not the required fields)
- The console removes the white space at the start and at the end of each value and of each line of a list, and ignores an empty line. It sends an optional field that has no value as no field. (Judgement: a value that the operator pastes often has white space, and a space at the end of a URI or an identifier is never intended.)
- The console does not check the values. The Register checks them with the rules of the participant model and the status rules. The UI has no copy of the rules. (Judgement: one set of rules, in the domain ([ADR 0013][adr-0013]).)
- When the Register refuses a value of a field, the console shows the `detail` of the field error as the error message of that field. The error of an item of a list shows at the field of the list. A refusal that names no field shows in the error summary of the form. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1], [SC 3.3.3][wcag-3-3-3])
- After a refusal, the form shows an error summary at its top, and the focus moves to the summary. The summary gives the `detail` of the problem, and a link to each field that has an error. ([GOV.UK Design System: Error summary][govuk-error-summary]; see the error summary in the [design system DESIGN.md][ui-design] and Decisions)
- After a refusal, the form keeps each value that the operator gave, and the version that it loaded. When the version is stale, the operator opens the form again to see the current values. (Judgement: the operator does not lose their work, and the console never replaces the change of a different person or tab with no decision of the operator.)
- While the request runs, the submit button is busy, and the form sends no second request. The console never sends a create, an add or a change again by itself.
- After a success, the console goes to the detail of the participant and shows the result in a toast. The focus moves to the heading of the detail. ([WCAG 2.2, SC 2.4.3][wcag-2-4-3], [SC 4.1.3][wcag-4-1-3]; judgement: the form page no longer exists)
- Each form has a "Cancel" link back to the page that opened it. A cancel sends no request, and the console does not ask the operator to confirm it. (See Decisions.)

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Creation and change of participants in the console | In scope: the console creates, adds and changes each record that the operator interface does, and does the status actions of brands and data holders, with the same use cases | Judgement: a lab user who tests their own data recipient software must register it first, and an HTTP client with the operator token is a barrier to the most frequent set-up tasks. The console API and the operator interface stay separate adapters on the same use cases ([ADR 0014][adr-0014]). |
| Identifiers of a create or an add in the console | The form shows a new random UUID for each new identifier, in a field that the operator can change. A data holder brand gets a `dataHolderBrandId`, and no `interimId`. | Judgement: the Register issues GUIDs in production ([CDS: Register API schemas][cds-register-schemas]), so a UUID is the usual value and the operator types nothing. A lab service gets its identifiers in its configuration before it starts, for example the brand ID of the mock bank, so the operator must be able to give a known value. A brand with a `dataHolderBrandId` is the usual brand. The Register refuses an identifier that a different record has (409). |
| Status of a new record in the console | A field, with ACTIVE selected first, and the status rules apply | Judgement: the operator interface permits each status in a create or an add (see the row "Status in the body of a create or an add"). With a fixed status of ACTIVE, an add under a parent that is not ACTIVE always fails, for example a software product of an INACTIVE brand. The field needs no rule in the UI: the Register refuses a status that breaks a rule, at the field of the status. |
| Child records in a create in the console | A create makes the legal entity only. The operator adds each brand and software product after the create. | The participant model does not require a brand for a legal entity, and the operator interface creates a legal entity with no brands. Judgement: each form then has the fields of one record, the add forms do the work of the child records, and each field error has a short pointer. |
| Values of a change form | The current values, and each field of the change, with the version that the form loaded. An add also sends the version. | Judgement: a change in the operator interface replaces each field ([RFC 9110, section 9.3.4][rfc9110-put]), so the form must send the fields that the operator did not change. The version stops a change that is based on an old view (see the row "Action with a stale view"). An add changes the legal entity, so it has the same check. |
| Forms in the console | A page with its own URL for each create, add and change. A dialog only for the confirmation of an action. | Judgement: the add form of a software product has 14 fields and the add form of a data holder brand has 15, which is too much for a dialog. A page has its own URL and history entry, and Back leaves it as the operator expects. |
| Confirmation of a create, an add or a change | None | [WCAG 2.2, SC 3.3.4][wcag-3-3-4] is met when the Register checks the values and the operator can correct them, or when the change is reversible. A change is reversible by a second change. A create and an add change no other record. |
| Lists of values in a form | A text area with one value on each line for a list that the operator writes. A group of checkboxes for a fixed set. A native select for one value of a fixed set. | Judgement: a text area is one control with one label and one error, the operator can paste a list, and the keyboard and the screen reader read it line by line. A list of text fields needs add and remove buttons and a focus rule for each, and gives no better access. The checkboxes and the select show each value of the fixed set, with the keyboard support of the browser. |
| Error summary in a form | An error summary at the top of the form after each refusal, with the focus on it | WCAG 2.2 SC 3.3.1 needs only the error at the field ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]). Judgement: a software product form is longer than one screen, so an error at a field far down the page is not seen. The [GOV.UK Design System: Error summary][govuk-error-summary] shows a summary for each error, with a link to each field, and moves the focus to it. The summary also has the refusals that name no field, so the form has one place for errors. |
| Values after a refusal | The form keeps the values of the operator and the version that it loaded. It does not get the current values by itself. | Judgement: the operator does not lose their work. A form that takes a new version by itself would replace the change of a different person or tab with no decision of the operator, which the version check stops. |
| White space in a form | The console removes it at the start and at the end of each value and of each line, and ignores an empty line | Judgement: a pasted value often has white space or an empty last line. This is not a rule of the Register: a value with a space at the end is never intended. The Register still checks each value. |
| Change and add actions in the console | Shown for each record in each status | Judgement: the domain permits a change and an add in each status, and refuses only a value or a status that breaks a rule. A list of permitted changes in the console API would always give each change. |
| Cancel of a form | A link back, with no confirmation and no warning about values that the operator did not send | Judgement: a lab form is quick to fill again, and a warning before each navigation adds a rule for each route. |
| Optional fields in a form | The label says "(optional)". A required field has no mark. | [GOV.UK Design System: Question pages][govuk-question-pages]. The console API schema gives which fields are optional, so the UI copies no rule. |

[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-apis
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-federation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-federation
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-status-mapping]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#status-mapping
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-metadata-cache]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management
[cds-archiving]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#archiving
[cds-get-oidc]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-openid-provider-config
[cds-get-jwks]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-jwks
[cds-get-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-statement-assertion-ssa
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-get-dh-brands-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands-summary
[cds-get-dh-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-statuses
[cds-get-sp-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-software-products-statuses
[cds-get-dr-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients-statuses
[cds-get-drs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-recipients
[cds-openapi]: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/docs/includes/swagger/cds_register.json
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[cds-error-403-adr-not-active]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-403-authorisation-adr-status-not-active
[cds-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-errors
[cds-known-register-errors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#standardise-register-api-error-codes
[cds-dr-calling-register]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipients-calling-the-cdr-register
[cds-ssa]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#software-statement-assertion-ssa
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-dcr-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#tocSregistrationproperties
[rfc3986-2]: https://www.rfc-editor.org/rfc/rfc3986#section-2
[fapi-as]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#authorization-server
[oidc-pairwise]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[cds-transaction-security]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#transaction-security
[cds-cert-trust-model]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-trust-model
[cds-cert-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-validation
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cds-nfrs]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#non-functional-requirements
[cds-register-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-endpoints
[cds-error-400-field-invalid]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-400-field-invalid
[cds-error-invalid-industry]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#register-error-404-field-invalid-industry
[cds-cors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors
[cds-schema-brand-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_schemas_tocSdataholderbrandsummaryv2
[rfc9110-etag]: https://www.rfc-editor.org/rfc/rfc9110#name-etag
[rfc9111-no-cache]: https://www.rfc-editor.org/rfc/rfc9111#name-no-cache-2
[rfc7517-jwks]: https://www.rfc-editor.org/rfc/rfc7517#section-5
[rfc6750-errors]: https://www.rfc-editor.org/rfc/rfc6750#section-3.1
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
[readme-sources]: ../../README.md#sources-for-requirements
[accc-known-issues]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29950003
[accc-token]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/30081139
[accc-cert-mgmt]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/360415310/Certificate+Management
[accc-mock-ssa-test]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.IntegrationTests/API/SSA/GetSoftwareStatementAssertion_Tests.cs#L416
[accc-mock-dr-ssa]: https://github.com/ConsumerDataRight/mock-data-recipient/blob/3421e0dfa7a8cca59e17dee0d7778ac1c56e15d7/Source/CDR.DataRecipient.SDK/Services/Register/SsaService.cs#L46-L56
[cds-registration-validation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#registration-validation
[accc-mock-dh-statuses]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/Source/CDR.Register.Repository/RegisterStatusRepository.cs#L65-L77
[accc-brands-summary]: https://consumerdataright.atlassian.net/wiki/spaces/DP/pages/29851982
[fetch-cors]: https://fetch.spec.whatwg.org/#http-cors-protocol
[accc-mock-certs]: https://github.com/ConsumerDataRight/mock-register/blob/b1bd1157befdf38d27f535b14b6fbc0a2007ac22/CertificateManagement/README.md
[research-accc-headers]: ../../docs/research/2026-09-25-accc-mock-register.md
[support-removed-sp]: https://cdr-support.zendesk.com/hc/en-us/articles/900001657906
[support-brands]: https://cdr-support.zendesk.com/hc/en-us/articles/900005881263
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[dsb-537]: https://github.com/ConsumerDataStandardsAustralia/standards-maintenance/issues/537
[adr-0003]: ../../docs/adr/0003-event-sourcing.md
[adr-0006]: ../../docs/adr/0006-telemetry.md
[adr-0011]: ../../docs/adr/0011-domain-data-as-typed-go.md
[adr-0012]: ../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0017]: ../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0019]: ../../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[adr-0021]: ../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[adr-0022]: ../../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[adr-0010]: ../../docs/adr/0010-background-jobs-with-river.md
[pki-design]: ../../pki/DESIGN.md
[pki-design-keys]: ../../pki/DESIGN.md#keys
[rfc5280-crl]: https://www.rfc-editor.org/rfc/rfc5280#section-5
[rfc5280-crl-number]: https://www.rfc-editor.org/rfc/rfc5280#section-5.2.3
[rfc5280-crldp]: https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.13
[rfc2585-mime]: https://www.rfc-editor.org/rfc/rfc2585#section-4.2
[rfc9110-methods]: https://www.rfc-editor.org/rfc/rfc9110#section-9.1
[rfc9110-405]: https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
[rfc9110-503]: https://www.rfc-editor.org/rfc/rfc9110#name-503-service-unavailable
[pki-design-revocation]: ../../pki/DESIGN.md#revocation
[accc-cps]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certification-Practice-Statement-v2.0-published-27-August-2025.pdf
[accc-cp]: https://www.cdr.gov.au/sites/default/files/2025-08/ACCC-Certificate-Policy-v1.04-published-27-August-2025.pdf
[accc-portal-guide]: https://www.cdr.gov.au/sites/default/files/2026-10/Participant%20Portal%20User%20Guide_-v23-20261014_091444.pdf
[cds-cert-mgmt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-management
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[rfc5280-bounds]: https://www.rfc-editor.org/rfc/rfc5280#appendix-A
[rfc8555-chain]: https://www.rfc-editor.org/rfc/rfc8555#section-9.1
[rfc8446-cert]: https://www.rfc-editor.org/rfc/rfc8446#section-4.4.2
[ui-design]: ../../web/packages/ui/DESIGN.md
[ui-brands]: ../../web/packages/ui/DESIGN.md#brands
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
[rfc9110-delete]: https://www.rfc-editor.org/rfc/rfc9110#name-delete
[rfc9110-idempotent]: https://www.rfc-editor.org/rfc/rfc9110#name-idempotent-methods
[rfc9110-if-match]: https://www.rfc-editor.org/rfc/rfc9110#name-if-match
[rfc9110-201]: https://www.rfc-editor.org/rfc/rfc9110#name-201-created
[rfc9110-put]: https://www.rfc-editor.org/rfc/rfc9110#name-put
[rfc9110-409]: https://www.rfc-editor.org/rfc/rfc9110#name-409-conflict
[rfc7396]: https://www.rfc-editor.org/rfc/rfc7396
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[adr-0013]: ../../docs/adr/0013-hexagonal-architecture-with-an-internal-composition-root.md
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[fetch-metadata]: https://w3c.github.io/webappsec-fetch-metadata/
[rfc9562-v4]: https://www.rfc-editor.org/rfc/rfc9562#section-5.4
[rfc9110-post]: https://www.rfc-editor.org/rfc/rfc9110#name-post
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-3-3-3]: https://www.w3.org/TR/WCAG22/#error-suggestion
[govuk-error-summary]: https://design-system.service.gov.uk/components/error-summary/
[govuk-question-pages]: https://design-system.service.gov.uk/patterns/question-pages/
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[apg-dialog]: https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/
[html-download]: https://html.spec.whatwg.org/multipage/links.html#downloading-resources
[rfc4514]: https://www.rfc-editor.org/rfc/rfc4514
[k8s-configmap-volume]: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
[k8s-init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[helm-values]: https://helm.sh/docs/chart_template_guide/values_files/
[dataholder-brand-configmap]: ../dataholder/design/deployment.md#dh-1345
[dataholder-register-statuses]: ../dataholder/design/registration.md#dh-0689
[adr-0020]: ../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[register-chart]: ../../charts/register/README.md
