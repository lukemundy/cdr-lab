# Research: an operator console for the mock data holder

- Research date: 2026-10-25
- Scope: the records and the staff tasks of a data holder that an operator console can show or do, the operator UIs of other CDR data holder mocks, the rules for staff access to CDR data, and what the lab data holder can give a console now.
- Applies to: CDS 1.36.0, CDR Rules Compilation No. 10, and the repository on the research date.

This note does not repeat [ACCC mock Data Holder](2026-10-06-accc-mock-data-holder.md), [CDS data holder InfoSec and consent](2026-10-06-cds-data-holder-infosec-and-consent.md) (withdrawal, dashboard duties, records summary) or [Data holder: the end of data sharing](2026-10-16-data-holder-end-of-sharing.md). The architecture of the consoles is in [Web front ends](2026-10-22-web-front-ends.md) and ADRs 0014 to 0018.

## Legend

- **UNVERIFIED**: I did not find a primary source. Do a check before you depend on it.
- **JUDGEMENT**: my opinion, not a fact from a source.

## Sources

| ID | Source |
|---|---|
| R | CDR Rules, Compilation No. 10, https://www.legislation.gov.au/F2020L00094/2025-03-04/text. Series: https://www.legislation.gov.au/F2020L00094/latest/text |
| F | ACCC, Rule 9.4 reporting: https://www.cdr.gov.au/resources/forms-and-templates/rule-94-reporting , and the data holder form example of 29 January 2026: https://www.cdr.gov.au/sites/default/files/2026-01/CDR-Compliance-Example-of-9.4-reporting-form-for-Data-Holders-published-29-January-2026.xlsx |
| A | CDS 1.36.0 admin OpenAPI: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/swagger-gen/api/cds_admin.json |
| X | DSB CX Guidelines, data holder withdrawal: https://cx.dsb.gov.au/consent-management/consent-management-data-holder/withdrawal/cx-guidelines |
| O | OAIC, Privacy Safeguard Guidelines Chapter 12, v5.0: https://www.oaic.gov.au/__data/assets/pdf_file/0023/111569/b5d8cd42fc64513facd4f9a63af347b045064711.pdf , and Guide to privacy for data holders: https://www.oaic.gov.au/consumer-data-right/consumer-data-right-guidance-for-business/privacy-obligations/guide-to-privacy-for-data-holders |
| S-n | CDR Support article n: `https://cdr-support.zendesk.com/hc/en-us/articles/<n>` |

## 1. Records and reports of a data holder

### 1.1 Records (rule 9.3)

A data holder must "keep and maintain records that record and explain" these items [R 9.3(1)]:

| Para | Record | In the mock now |
|---|---|---|
| (a) | Authorisations given by CDR consumers | Arrangements |
| (b) | Amendments to or withdrawals of authorisations | Withdrawals. Amendment is out of scope. |
| (c) | Notifications of withdrawals of consents to collect, from the data recipient | Arrangement revocation by the data recipient |
| (d) | Disclosures of CDR data in response to consumer data requests | Not recorded |
| (e) | Refusals to disclose CDR data, and the rule or standard relied on | None occur (see 1.3) |
| (f), (fa) | CDR complaint data, and CDR consumer complaints | No complaints process |
| (g) | The processes to ask for authorisation and amendment, "including a video of each process" | No |

The (ca), (cb), (da) and (db) records are for energy, product data agreements and Schedule 3 elections. They do not apply to the mock.

- Each record has the date and time when it was made and, if applicable, when the event occurred [R 9.3(3)]. The data holder keeps each record for 6 years [R 9.3(5)].
- "Both DHs and ADRs are expected to record how the withdrawal was requested by the consumer", under 9.3(1)(b). The dashboard does not have to show the method. [S-4411292647823] The CX Guidelines say the same in checklist item 5CM2.00.10. [X]
- Only a refusal to disclose needs a record. A refusal to ask for an authorisation does not. [S-900001914223]

### 1.2 Copies of records for the consumer (rule 9.5)

- A consumer can ask the data holder for copies of the records of 9.3(1)(a), (b), (c), (d), (f) and (fa) that relate to them. The data holder gives them within 10 business days, with no fee. [R 9.5(1), (4), (5)]
- The ACCC has prescribed no form. The consumer can ask in any form, "e.g. contacting the data holder by telephone or attending in-person at a branch". [S-900004438886]
- JUDGEMENT: a person at the data holder finds the records of one consumer. This is a staff task, and the main rule-based reason for a record view of each customer in an operator console.

### 1.3 Reports (rule 9.4)

- Each 6 months, the data holder reports the complaint data, the number of each type of request, and the number of refusals with the rule relied on. [R 9.4(1)] The report goes through the CDR Participant Portal by 30 January and 30 July. [F]
- The form asks for complaints (Q1 to Q7), request counts (Q8 to Q10), and refusals by type: 403 for a potential threat, 404 or 422 for a blocked or suspended account, and 429 for traffic thresholds or a denial-of-service attack (Q11 to Q13). [F]
- "The return of a 403, 404, 422 and 429 HTTP error code in response to circumstances other than those set out in the above table does not constitute a refusal to disclose data". [S-4523876174863]
- The scope of the mock has no blocked accounts, no refusal for harm and no traffic thresholds ([data holder: non-goals](../../services/dataholder/design/README.md#non-goals)). A 403 `AdrStatusNotActive` is not in the table of S-4523876174863. Thus the refusal counts of the mock are always zero.
- Get Metrics is out of scope for the mock, but its schema shows the counts that a production data holder keeps, for example `AuthorisationMetricsV2` with `activeAuthorisationCount`, `revokedAuthorisationCount` and `abandonmentsByStage`. [A]

## 2. Staff and the consumer dashboard

- A consumer "may withdraw the authorisation at any time: (a) by using the data holder's consumer dashboard; or (b) by using a simple alternative method of communication to be made available by the data holder for that purpose." For (b), the data holder gives effect "as soon as practicable, and in any case within 2 business days". [R 4.25(1), (2)]
- "Data Holders must always have two forms of mechanisms available to consumers to withdraw authorisations." The alternative method "is intended to reflect existing methods of communication made available by Data Holders to its consumers." [S-900001658546]
- The notice to the data recipient follows each withdrawal, also by method (b). [R 4.26A]
- The consumer dashboard is "an online service" that the data holder gives "the CDR consumer". [R 1.15(1)] Rule 1.15(1)(d) "applies to online services, not other channels". [X, 5CM2.00.08]

JUDGEMENT:

- The dashboard is self-service for one consumer, who authenticates as the consumer. An operator console is a staff tool over all customers. A staff tool cannot satisfy rule 1.15.
- A withdrawal in the operator console is the model of method (b): a person at the data holder acts on a request from the consumer, for example a telephone call. The mock now gives `WITHDRAWN_BY_CONSUMER` for each withdrawal and records no channel.
- The mock gives effect to a withdrawal at once. This obeys rule 4.25(2), so a delay of 2 business days is not necessary.

## 3. Operator UIs of other CDR data holders

Repositories examined on 2026-10-25:

| Repository | Commit | Operator UI |
|---|---|---|
| [ConsumerDataRight/mock-data-holder](https://github.com/ConsumerDataRight/mock-data-holder) | `2d34f0d` | No. `Source/` has no UI files. The operator functions are the Manage API, with no authentication. |
| [ConsumerDataRight/mock-data-holder-energy](https://github.com/ConsumerDataRight/mock-data-holder-energy) | `8324e9c` | No. The repository is archived. Energy is now in mock-data-holder. |
| [ConsumerDataRight/authorisation-server](https://github.com/ConsumerDataRight/authorisation-server) | `d977b3e` | No. `Source/CdrAuthServer.UI` has the consumer pages, and `src/pages/Settings/` has static mock-ups of a consumer dashboard with hard-coded data and no API calls. |
| [ConsumerDataRight/mock-data-recipient](https://github.com/ConsumerDataRight/mock-data-recipient) | `b1836c1` | A data recipient UI, not a data holder UI. `Source/CDR.DataRecipient.Web/Views/` has pages for the data holders, DCR registrations (client ID, brand, status, last update) and consents (arrangement ID, client ID, scope, expiry, sharing duration). |
| [ConsumerDataStandardsAustralia/mock-data-holder-nodejs](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-nodejs), [mock-data-holder-java](https://github.com/ConsumerDataStandardsAustralia/mock-data-holder-java) | `7969d2b`, `e7e6335` | No UI files. |
| [pingidentity/pingidentity-cdr-sandbox](https://github.com/pingidentity/pingidentity-cdr-sandbox) | `33c523d` | Only the product admin consoles of the Ping products. No CDR-specific operator view. |

The ACCC Sandbox (https://cdrsandbox.gov.au/) is a hosted copy of the mocks with a participant portal. UNVERIFIED: what the portal shows. It is not open source.

No public ACCC or DSB data holder mock has an operator UI. The nearest public model is the data recipient side: the registration and consent pages of mock-data-recipient.

## 4. What an operator needs that the operator interface does not give

JUDGEMENT, with the evidence where it exists:

| Item | Evidence | Value |
|---|---|---|
| The notices to the data recipient, with each attempt, its time and its result | Rule 4.26A. The data holder records each attempt ([data holder: the record of each attempt](../../services/dataholder/design/end-of-an-arrangement.md#dh-0940)), but the operator interface gives only `noticeStatus`. | High. A data recipient developer sees why a notice failed. |
| The client registrations of software products | mock-data-recipient has a registrations page on the other side. No rule requires it. | High. Registration is the first point of failure for a new data recipient. |
| The Register status of each software product that the data holder holds, and the time of the last poll | The 5-minute limit of [CDS: Metadata Cache Management](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#metadata-cache-management) | High. It explains a 403 `AdrStatusNotActive`, and the time after a status change on the Register. |
| The channel of a withdrawal | S-4411292647823, X 5CM2.00.10 | Medium. It models rule 4.25(1)(b). |
| Disclosure records | R 9.3(1)(d), R 7.9 | Medium. It needs a new record in the data holder. |
| The loaded products | Get Products gives them | Low |
| Signing keys | One key, and no change of the key is in scope | Low |
| OTP emails | The mail sink has its own UI | Do not show. An OTP in a staff tool is a way around the authentication of the consumer. |
| Request counts, errors and latency | The telemetry backend | Do not copy. |

## 5. Staff access to CDR data

- Privacy safeguard 12 and Schedule 2 do not apply to data holders: "Privacy Safeguard 12 applies to accredited data recipients of CDR data and designated gateways for CDR data. It does not apply to data holders. However, data holders must ensure that they are adhering to their obligations under the Privacy Act 1988 ... including APP 11". [O, Chapter 12] Rule 7.11 links Schedule 2 to subsection 56EO(1) only. [R 7.11(1)]
- The privacy safeguards that apply to data holders are 1, 10, 11 and 13. [O, Guide to privacy for data holders]
- An OAIC example of misuse, for a data recipient: "if an employee of a CDR entity browses consumer statements to discover information about someone they know". [O, Chapter 12, Part A]
- I found no CDS rule about staff access or back-office tools of a data holder. UNVERIFIED (absence). APRA CPS 234 applies to ADIs, and sets information security controls for their systems. UNVERIFIED: I did not read it.
- JUDGEMENT: a production staff tool needs an identity for each staff member, views with least privilege, and a record of each read and each action. The lab data is invented and has no value, so the lab console needs a [production gap](../production-gaps.md) entry, as the Register console has.

## 6. The lab data holder now

From the repository on the research date:

- The console listener code is already shared: [`pkg/platform/consolehttp`](../../pkg/platform/consolehttp) gives the settings, the `Host` allowlist, the security headers, the cross-origin refusal, the cache headers and the SPA fallback. Some console API code is still in the Register adapter only: the page parameters, the strict JSON decode, the field-errors problem, and the 409 for a stale version. The strict decode and the 500 problem have copies in the two operator interfaces.
- The generic parts of the Register console UI (list search in the URL, router search, form values, the query client, the problem client and the test set-up) are in `web/apps/register-console`, not in a shared package. `web/packages/ui` has all the components.
- Only arrangements and client registrations are event sourced. Customers, accounts, transactions, product data and Register statuses are tables with no history and no version. The close of an account and the withdrawal of a consent already give 409 from the state of the record.
- The operator interface has no list or search of customers, and no list of arrangements, registrations or Register statuses. The notice attempts are only in the event payload.
- The validation errors of the data holder domain are rule strings, with no field. A form cannot show an error at its field without a change to the domain.
- The telemetry has no query strings now, because otelhttp v0.72.0 does not record `url.query`. No test makes sure of this. A search of customers by name or email puts personal data in the query string.

## Main findings

- No public data holder mock has an operator UI. The console requirements come from the CDR Rules and from the needs of a lab operator.
- Rule 9.5 (copies of records on request, in any form) is the rule-based reason for a record view of each customer.
- A withdrawal by staff is the alternative method of rule 4.25(1)(b). The data holder must record how the consumer asked for it, and the mock does not record this now.
- The notice attempts, the client registrations and the Register statuses are the most useful views that the operator interface does not give.
- The refusal counts of rule 9.4 are always zero in the mock, and the telemetry gives the request counts. A report has no value in the lab.
- Privacy safeguard 12 does not apply to data holders. No CDS rule covers staff tools. The production controls for staff access are a production gap.
