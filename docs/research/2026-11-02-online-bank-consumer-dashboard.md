# Online bank of a data holder, milestone 4: research

- Research date: 2026-11-02
- Applies to: CDS 1.36.0, CDR Rules Compilation No. 10, DSB CX Guidelines, WCAG 2.2
- File and line references are to the repository on the research date. Paths in `internal/` are in `services/dataholder/`.
- Purpose: the facts for milestone 4 of the project "A consumer uses the online bank of each data holder": "The consumer dashboard of CDR Rules rule 1.15: each consent of the customer, its details and its withdrawal, with the same effects as a withdrawal by the operator." Project acceptance criterion: "A consumer withdraws a consent on the dashboard. The data recipient then gets a notice, and the tokens of the consent stop working. The record of the withdrawal gives the dashboard as the method." Open risk of the project: "The messages that the CX Withdrawal Standards require on the dashboard."
- Related notes: [Online bank: open and close of an account](2026-11-02-online-bank-open-and-close.md), [Online bank: sign-up, sign-in and accounts](2026-11-02-online-bank-sign-in-and-accounts.md), [Data holder console](2026-10-25-data-holder-console.md), [Data holder: the end of data sharing](2026-10-16-data-holder-end-of-sharing.md), [Web front ends](2026-10-22-web-front-ends.md) (section 9.4). This note adds to them.

Check the sources again before you depend on a fact.

## Legend

- **Fact**: a source or the code says it. Each fact has a citation.
- **Option** and **Recommendation**: our analysis for the requirements step. They are not facts.
- **UNVERIFIED**: I did not find a primary source. Do a check before you depend on it.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | CDR Rules, Compilation No. 10: rules 1.10B, 1.15, 4.23, 4.24, 4.25, 4.26, 4.26A, 4.27, 7.2(8), 7.9, 9.3; Schedule 3, clauses 2.1 and 2.3 | https://www.legislation.gov.au/F2020L00094/2025-03-04/text |
| S2 | CDS 1.36.0: Dashboard Standards, Data Holder Dashboards. Source `slate/source/includes/cx_standards/dashboards.md` at tag `v1.36.0` | https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/slate/source/includes/cx_standards/dashboards.md |
| S3 | CDS 1.36.0: Withdrawal Standards. Source `slate/source/includes/cx_standards/withdrawals.md` at tag `v1.36.0` | https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/slate/source/includes/cx_standards/withdrawals.md |
| S4 | DSB CX Guidelines: data holder withdrawal (5CM2) | https://cx.dsb.gov.au/consent-management/consent-management-data-holder/withdrawal/cx-guidelines |
| S5 | DSB CX Guidelines: data holder authorisations, the dashboard (5CM1) | https://cx.dsb.gov.au/consent-management/consent-management-data-holder/authorisations/cx-guidelines |
| S6 | DSB CX Guidelines: Authorisation to disclose (3AU) | https://cx.dsb.gov.au/authorise/authorisation-to-disclose |
| S7 | CDR Support: "Consumer Dashboard - revoked and expired consents" (article 4411292647823, updated 2025-11-21) | https://cdr-support.zendesk.com/hc/en-us/articles/4411292647823 |
| S8 | CDR Support: "Dashboards" (implementation guide, article 6140622825231, updated 2025-11-25) | https://cdr-support.zendesk.com/hc/en-us/articles/6140622825231 |
| S9 | CDR Support: "Dashboards" (question, article 900001848986, updated 2024-05-09) | https://cdr-support.zendesk.com/hc/en-us/articles/900001848986 |
| S10 | CDR Support: "Data Holder Dashboard data sharing display at data cluster level" (article 4920717853327, updated 2024-07-29) | https://cdr-support.zendesk.com/hc/en-us/articles/4920717853327 |
| S11 | CDR Support: "Privacy Safeguard 5 - dates on the consumer dashboards" (article 5333587630095, updated 2025-11-21) | https://cdr-support.zendesk.com/hc/en-us/articles/5333587630095 |
| S12 | CDR Support: "Withdrawal of authorisation to disclose CDR data and notification" (article 900001658546, updated 2026-08-28) | https://cdr-support.zendesk.com/hc/en-us/articles/900001658546 |
| S13 | CDR Support: "Closed accounts in the banking and non-bank lenders sectors" (article 14519397304079, updated 2025-12-04) | https://cdr-support.zendesk.com/hc/en-us/articles/14519397304079 |
| S14 | CDR Support: "Offline customer guidance" (article 4627046759823, updated 2024-07-16) | https://cdr-support.zendesk.com/hc/en-us/articles/4627046759823 |
| S15 | WCAG 2.2: SC 3.3.4 Error Prevention, SC 4.1.3 Status Messages | https://www.w3.org/TR/WCAG22/ |

The CDR Support web pages give 403 to an automated client. I read S7 to S14 through the help centre API (`/api/v2/help_center/en-us/articles/<id>.json`).

## 1. What the repository has now

Facts:

- The data holder `DESIGN.md` puts the dashboard out of scope: "A consumer dashboard. A later change adds it to the online bank site." (`DESIGN.md:48`). The operator console "is not a consumer dashboard, and it does not satisfy rule 1.15 ... The consumer dashboard stays out of scope." (`DESIGN.md:742`)
- The consent page says "You can stop the sharing at any time. To stop it, ask {{$.Bank}}." (`internal/adapter/consumerpages/templates/consent.html:16`; `DESIGN.md:483`; Decision "Instructions to withdraw", `DESIGN.md:1171`)
- Decision "Method of the consumer to withdraw": the operator withdraws, and "Neither shows the messages of the CX Withdrawal Standards" (`DESIGN.md:1186`). Decision "Method of a withdrawal": "A field with the one value `ALTERNATIVE_METHOD` ... a consumer dashboard can add a second value later with no change to the records that exist" (`DESIGN.md:1204`). Withdrawal by the consumer: `DESIGN.md:651-655`.
- The domain already has a method field. `WithdrawalMethod` has one value, `ALTERNATIVE_METHOD` (`internal/domain/arrangement.go:30-35`). The command `WithdrawArrangement{WithdrawnAt, Method}` gives the event `ArrangementWithdrawn{withdrawnAt, method}`, and refuses an empty method (`arrangement.go:181-195`, `211-220`). Apply sets the end reason `WITHDRAWN_BY_CONSUMER` and the recorded method (`arrangement.go:250-253`). An older `ArrangementEnded` withdrawal with no method reads as `ALTERNATIVE_METHOD` (`arrangement.go:134-145`).
- `app.Arrangements.WithdrawByConsumer(ctx, customerNumber, arrangementID)` checks the customer, gives `ErrArrangementNotFound` for an arrangement of a different customer and `ErrArrangementEnded` for an ended one, and always uses `WithdrawalMethodAlternative` (`internal/app/arrangements.go:53-79`). The `end` function revokes the grant first, then records the event (`arrangements.go:184-203`).
- The notice depends on the end reason only: `NeedsNotice` is true for `WITHDRAWN_BY_CONSUMER` and `CONSUMER_NOT_ELIGIBLE` (`internal/domain/notice.go:20-22`). A projection inserts the notice job in the transaction of the event (`internal/adapter/noticejobs/noticejobs.go:45-60`). Thus a withdrawal with a different method gets the same notice with no change.
- The read model stores the method (`internal/adapter/arrangementstore/arrangementstore.go:22-40`) and gives the arrangements of a customer in the sequence of their confirmation (`ArrangementsOf`). The projection runs in the transaction of the event, so a read after a withdrawal shows the new state (rule 4.27).
- The operator interface and the console give `withdrawalMethod` (`internal/adapter/operatorapi/arrangements.go:29,44`; `DESIGN.md:735`, `752`). The console withdraw action calls the same use case (`internal/adapter/console/actions.go:21-23,73-75`).
- The arrangement records `clientName` (the `client_name` at the confirmation), `softwareId`, the scopes, the account IDs, `confirmedAt` and the sharing duration (`internal/domain/arrangement.go:53-67`; Decision `DESIGN.md:1170`). It does not record the `org_name` or the `legal_entity_name` that the consent page showed (`DESIGN.md:467`). The registration has them (`internal/domain/registration.go:70-73`).
- The data cluster language of the consent page is in Go: `domain.DataClusters(scopes)` (`internal/domain/consent.go:23-50`).
- The site API has the session, the accounts, the balance, the transactions and the close (`internal/adapter/banksite/banksite.go:57-77`). The open of an account is not on `main` yet (milestone 3 is not complete). The close gives each arrangement that ends with its `arrangementId` and `clientName` to the customer of the session (`DESIGN.md:997`).
- No code records a disclosure. No table or event has the time or the data of a response of a resource API (`grep -i disclos` finds only `internal/domain/consent.go:19`). The console note gives rule 9.3(1)(d) as "Not recorded" ([Data holder console](2026-10-25-data-holder-console.md), section 1.1). The resource routes each have one scope (`internal/adapter/httpapi/consent.go:65-91`).
- The in-process test of the close builds the real `authserver.DataHolder` verifier, the event store with its projections and the site handler in one process (`internal/root/banksiteclose_test.go:42-130`). The console browser test of a withdrawal uses an arrangement of the conformance suite, and skips when the lab has none (`web/apps/dataholder-console/e2e/withdraw.spec.ts:10-34`).

## 2. Rule 1.15: what the dashboard contains

### Facts

- A data holder that receives a consumer data request from an accredited person "on behalf of a CDR consumer" must provide "an online service" that (S1, rule 1.15(1)):
  - (a) the consumer can use "to manage authorisations to disclose CDR data";
  - (b) "contains the details of each authorisation ... specified in subrule (3)";
  - (ba) contains "any information in the data standards that is specified as information for the purposes of this rule";
  - (bb) contains "any information on the Register of Accredited Persons that is specified as information for the purposes of this rule";
  - (c) "allows the CDR consumer, at any time, to withdraw a current authorisation";
  - (d) "as part of the process of withdrawing an authorisation, displays a message, in accordance with the data standards, about the consequences of proceeding";
  - (e) "is simple and straightforward to use, and is no more complicated to use than the process for giving the authorisation";
  - (f) "is prominently displayed and readily accessible".
- Rule 1.15(3), for each authorisation (S1):
  - (a) "details of the CDR data that has been authorised to be disclosed";
  - (b) "when the CDR consumer gave the authorisation";
  - (c) "the period for which the CDR consumer gave the authorisation";
  - (d) "if the authorisation is current—when it is scheduled to expire";
  - (e) "if the authorisation is not current—when it expired";
  - (f) "information relating to CDR data that was disclosed pursuant to the authorisation (see rule 7.9)";
  - (g) a disclosure under subsection 56EN(4) of the Act (correction of data): "that fact";
  - (h) "details of each amendment".
- Rule 7.9(1): the data holder must, "as soon as practicable", update each dashboard "to indicate: (a) what CDR data was disclosed; and (b) when the CDR data was disclosed; and (c) the accredited data recipient, identified in accordance with any entry on the Register of Accredited Persons specified as being for that purpose." (S1)
- Rule 9.3(1)(d): the data holder keeps records of "disclosures of CDR data made in response to consumer data requests". (S1)
- Schedule 3, clause 2.3: in banking, the obligation applies to a request "on behalf of a CDR consumer who is eligible". (S1)
- The ACCC: rule 7.9 "requires the consumer dashboard to show what CDR data was disclosed, when the CDR data was disclosed, and the ADR it was disclosed to". It "could be met by providing notifications elsewhere on the consumer dashboard". (S9)
- The ACCC: the CX Guidelines list the disclosures "per data cluster"; "Providing a generic note may be insufficient for the purposes of PS10." (S10)
- The ACCC: the dates of a data cluster are "The date the data was first collected/disclosed" and "The final collection/disclosure date". When the end is not known, the end of the authorisation can show until the actual date is known. "PS5/10 would not be considered satisfied by listing the date(s) of the consent/authorisation period". The obligations apply "for both active and expired sharing arrangements". (S11; also S5, 5CM1.00.17 and 5CM1.00.24)
- The CDS standard for rule 1.15(1)(ba): "Data holders **MUST** advise consumers to check with the relevant data recipient for information about how their data may be handled. The precise wording of this message is at the discretion of the data holder." Suggested text: "You should check with [ADR brand/the data recipient] for more information on how they are handling your data, and for any other permissions you may have given them. See [ADR]'s CDR policy or their Dashboard for more information." (S2)
- The other row of S2 is the amendment (rule 1.15(3)(h)). The amendment is out of scope (`DESIGN.md:41`).
- The ACCC: no rule or standard requires particular status terms or a format. Show the scheduled expiry of a current consent and the time of the end of an expired one. "the CDR Rules do not require the method of withdrawal to be shown on the dashboard." (S7) "there is no obligation to display the reason for expiry" (S8)
- The ACCC: there is "no requirement to display details about the software product", but the CX Guidelines suggest the software product name. "The DSB recommends at a bare minimum to display the ADR's brand name, software product name, and legal entity. The DSB also recommends the ADR's accreditation number be displayed on DH dashboards." (S8)
- The ACCC on a consent on a single occasion: list it "As 'Expired' if the duration is 0, absent, or if duration is up to 24 hours that has elapsed". (S8)
- The ACCC: "Only ADRs are required to provide a CDR receipt, however, the DSB recommends that DHs also provide a CDR receipt." The CDS and the rules give no date format. (S8)
- CX Guidelines (not binding): tabs for active and archived arrangements (5CM1.00.08), status labels (5CM1.00.14), the accounts that the arrangement shares (5CM1.00.19), the legal entity name (5CM1.00.23), the brand name (5CM1.00.29), "Granted", "Expired", "Access period" (5CM1.00.18). (S5)
- Rule 7.2(8): "the CDR entity must make its CDR policy readily available through each online service by means of which the CDR entity ... ordinarily deals with CDR consumers." The guideline 5CM1.00.25 and 5CM1.00.26 apply it to the dashboard. (S1, S5)
- UNVERIFIED: I found no data standard or Register entry that is "specified" for rule 1.15(1)(bb) or rule 7.9(1)(c). The DSB recommendation of S8 is the only guidance about the identity of the data recipient on the dashboard.

### What the data holder has for each item

| Item | Source | Data now |
|---|---|---|
| Data authorised (1.15(3)(a)) | Scopes of the arrangement | Yes: `domain.DataClusters(scopes)` gives the cluster and permission language, which the Data Language Standards require also "when CDR data is ... reviewed or withdrawn" ([Web front ends](2026-10-22-web-front-ends.md), section 9.6) |
| Accounts of the authorisation | `AccountIDs` | Yes. A closed account shows with `openStatus` `CLOSED` (Decision "Closed accounts", `DESIGN.md:1153`) |
| When given (3)(b) | `confirmedAt` | Yes |
| Period (3)(c) | `sharingDurationSeconds` | Yes. A duration of 0 or 24 hours or less is "one occasion" (`DESIGN.md:480`) |
| Scheduled expiry (3)(d) | `SharingEndsAt()` | Yes |
| When it expired (3)(e) | `EndedAt`, or `SharingEndsAt()` in the past | Yes |
| Disclosures (3)(f), rule 7.9 | none | **No.** No record of what was disclosed or when |
| 56EN(4) disclosures (3)(g) | none | Not applicable: the data holder has no correction process |
| Amendments (3)(h) | none | Not applicable: amendment is out of scope |
| Data recipient | `clientName` at the confirmation; `clientId` | Only the software product name. The brand name (`org_name`) and the legal entity name are only on the current registration, which can change or be deleted |
| Accreditation number | none | No. The data holder does not read data recipient details from the Register |
| Consequences and handling messages (1.15(1)(d), (ba)) | static text | Not present now |
| CDR policy (rule 7.2(8)) | none | No CDR policy |

### Options for the record of disclosures

| | A. No record; a production gap | B. First and last time for each arrangement and data cluster | C. One row for each disclosure |
|---|---|---|---|
| Rule 1.15(3)(f) and 7.9 | Not met. The dashboard shows only the dates of the consent, which S11 says is not sufficient | Met, as the CX Guidelines show it (S10, S11) | Met; the dashboard reads the first and the last time of each cluster from the rows |
| Rule 9.3(1)(d) | Not met (as now) | Partly: it keeps no record of each disclosure | Met |
| Code | None | An upsert in a middleware of the resource routes after a 2xx response | An insert in the same middleware, and a query with `min` and `max` |
| Data | None | One row for each cluster of an arrangement | One row for each successful request: arrangement ID, data cluster, time |

- The production-gaps test says "A function that a project put out of scope is not a gap" (`docs/production-gaps.md`, Rules). Option A is thus a scope decision, not a gap.
- A disclosure is not a change of the state of an arrangement. A table of rows, as for transactions, fits it better than an event (ADR 0003 applies to participant state).
- The scope of a route selects the cluster (`internal/adapter/httpapi/consent.go:65-91`), with one exception: `bank:accounts.detail:read` has two cluster names, which depend on `bank:accounts.basic:read` in the same arrangement (`DESIGN.md:474-475`). The UserInfo endpoint discloses the `profile` cluster.
- JUDGEMENT: a 2xx response of a resource API or of the UserInfo endpoint is a disclosure. A response with no records, for example no transactions in the period, is also a disclosure of that cluster: the data recipient learns that there is no data.

**Recommendation: C.** It meets rules 1.15(3)(f), 7.9 and 9.3(1)(d) with one insert and one query, and the dashboard needs it. B saves little and keeps a gap in rule 9.3(1)(d). Show for each data cluster "First shared" and "Last shared", and "Not shared yet" when no row exists (the hidden draft in S8 also suggests a "no data shared yet" message).

### Options for the name of the data recipient

| | A. `clientName` only | B. Record `org_name` and `legal_entity_name` in the arrangement at the confirmation | C. Read them from the current registration |
|---|---|---|---|
| DSB minimum (S8) | Not met | Met | Met while the registration exists |
| Consistency with the consent page | Partly | The same names as the consent page showed (`DESIGN.md:467`) | Changes with the registration |
| Change | None | A field in `ArrangementCreated`: a lab reset (`docs/production-gaps.md`: the codec reads only schema version 1) | A join to the registrations; a deleted registration gives no name |

**Recommendation: B**, for the reason of the Decision "`clientName` of an arrangement" (`DESIGN.md:1170`). Use the brand name (`org_name`) in the DSB messages ("[ADR brand]"). Do not show an accreditation number: the data holder does not read it, and S8 only recommends it.

## 3. Withdrawal on the dashboard

### Facts

- Rule 4.25(1): the consumer may withdraw "(a) by using the data holder's consumer dashboard; or (b) by using a simple alternative method of communication". (S1)
- The ACCC: "Data Holders must always have two forms of mechanisms available to consumers to withdraw authorisations. One mechanism must be through the consumer dashboard, and the other may be of the Data Holder's choosing, but an alternate method must be in place." (S12) Thus the operator withdrawal stays.
- Rule 4.26(1)(b): a withdrawal on the dashboard expires the authorisation "when the authorisation was withdrawn". Rule 4.26(1)(a) gives method (b) up to 2 business days. (S1) The mock gives effect at once for both methods (`DESIGN.md:653-655`).
- Rule 4.26A: the data holder notifies the accredited person "If an authorisation ... is withdrawn or otherwise expires". (S1) The CDS exceptions are a withdrawal "initiated via the Data Recipient", the natural expiry and a status change ([Data holder: the end of data sharing](2026-10-16-data-holder-end-of-sharing.md), section 6). A withdrawal on the dashboard is not an exception.
- Rule 4.27: the data holder updates the dashboard "as soon as practicable after the information ... changes". (S1)
- Rule 9.3(1)(b): records of "withdrawals of authorisations". The ACCC: "Both DHs and ADRs are expected to record how the withdrawal was requested by the consumer." (S1, S7; S4 5CM2.00.10)
- CDS Withdrawal Standards for the data holder (S3), both **MUST**, "As part of the withdrawal process":
  - Consequences: "advise the consumer to review the consequences of withdrawal with the Data Recipient before they stop sharing their data." Suggested text (**MAY**): "You should check with [Data Recipient] before you stop sharing to understand the consequences." or "You should check with [Data Recipient] to see if your service will be impacted before you stop sharing."
  - Redundant data: "inform the consumer about the handling of redundant data and the right to delete." Suggested text (**MAY**): "CDR data is either deleted or de-identified when it is no longer required." or "[Data Recipient] will have specific policies on how to handle your data once it's no longer required."
  - The DSB gives the text as suggestions. The data holder can "use or paraphrase" it. The other rows of S3 are for data recipients, secondary users and joint accounts, which are out of scope.
- CX Guidelines for the withdrawal (S4; not binding unless a rule or standard): "Stop access", "Stop sharing" or "Stop data sharing" (5CM2.00.04); "positive friction ... to mitigate user error", for example two steps (5CM2.00.05); "a message to consumers that withdrawal was successful", shown on the dashboard at once (5CM2.00.06); CDR receipts in writing (5CM2.00.07, a MAY).
- Rule 1.15(1)(e): no more complicated than giving the authorisation. To give an authorisation, the consumer enters a customer number, an OTP and one confirmation (`DESIGN.md`, Consumer authentication and Consent).
- WCAG 2.2 SC 3.3.4: a change that cannot be reversed needs a confirmation, a check or a reversal. SC 4.1.3: a status message that assistive technology announces. (S15)

### The event (ADR 0021)

- ADR 0021: "Two facts get separate event types when a rule or a reader acts differently on them." "A field whose values all rules and readers handle in the same way is permitted, for example a final status that the service only records and shows." (`docs/adr/0021-record-one-business-fact-in-each-event-type.md`, Decision)
- With the two methods, each rule of the data holder acts in the same way: the time of the end is the time of the request, the tokens are revoked, the notice is sent (`notice.go:20-22`), and the dashboard does not show the method (S7). The method is only recorded and shown to the operator. Rule 4.26(1)(a) gives method (b) a later latest time, but the mock gives effect at once, so no rule of the mock reads the method.
- The Decision "Method of a withdrawal before the data holder recorded the method" (`DESIGN.md:1205`) made `ArrangementWithdrawn` a separate event type for the method. A second value needs no new event type and no change to the events that exist.

**Recommendation:** a second value of `WithdrawalMethod` on the existing `ArrangementWithdrawn` event, for example `CONSUMER_DASHBOARD` (the term of rule 4.25(1)(a)). No new event type. `WithdrawByConsumer` takes the method from its adapter: the operator interface and the console give `ALTERNATIVE_METHOD`, and the site gives `CONSUMER_DASHBOARD`. The OpenAPI documents of the operator interface and the console add the value. The name of the value is a new convention: propose it in the requirements PR.

### Recommendations for the withdrawal

- Reuse `app.Arrangements.WithdrawByConsumer` with the customer of the session. It already gives 404 for an arrangement of a different customer and 409 for an ended one, revokes the grant before the event, and the projection inserts the notice job. The site adds no rule.
- Two steps, as the close: the action "Stop sharing" on a current consent, then a confirmation dialog. The dialog names the app (`clientName`) and the brand name of the data recipient, says that the sharing stops at once and cannot be reversed, and that the bank tells the app. It shows the two CDS messages word for word from the DSB suggestions, with the brand name of the data recipient: "You should check with [brand] before you stop sharing to understand the consequences." and "CDR data is either deleted or de-identified when it is no longer required." JUDGEMENT: the DSB text needs no review, and a test can check it.
- After a withdrawal: a status message that names the app, the consent in the group of the ended consents, and the focus on the heading (WCAG SC 4.1.3; S4 5CM2.00.06).
- A withdrawal of an arrangement that ended between the read and the confirmation gives 409. The dialog then says that the consent changed, and the site shows the current state. The site never sends a withdrawal again by itself.
- The path can have the `arrangementId`, as in the console (`POST /api/consents/{arrangementId}/withdraw`), with no body. The rule "No path of the site API contains a customer number or a different identifier of a consumer" (`DESIGN.md`, Site API) does not forbid it: a `cdr_arrangement_id` "MUST NOT identify a consumer" (`DESIGN.md:503`), and telemetry has the route pattern only.
- No email and no CDR receipt. The data holder has no duty (S8), and the Decision "Notice email to the consumer" (`DESIGN.md:1250`) stays.

## 4. Current and past consents

### Facts

- Rule 1.15(3)(e) requires "if the authorisation is not current—when it expired". The dashboard thus shows authorisations that are not current. The PS10 dates apply also to "expired sharing arrangements" (S11).
- No rule or guidance gives how long a past authorisation stays on the dashboard (S5: not addressed). Records stay for 6 years (rule 9.3(5); [Data holder console](2026-10-25-data-holder-console.md), section 1.1).
- An arrangement with a sharing duration of 0 has `SharingEndsAt()` equal to `confirmedAt`, so it is ended at once (`arrangement.go:116-118`, `147-149`). S8 permits "Expired" for it.

**Recommendation:** show each arrangement of the customer: the current ones first, then the ended ones in a group of their own, as the accounts page does with closed accounts. For a current one show "Shares until <date>". For an ended one show the date of the end, with no reason and no method (S7, S8). Order in each group: newest confirmation first (JUDGEMENT). No time limit and no pages: a customer has few arrangements (`DESIGN.md`, Console API).

## 5. The consent page: instructions to withdraw

### Facts

- Rule 4.23(1)(f) and (g): "a statement that, at any time, the authorisation can be withdrawn" and "instructions for how the authorisation can be withdrawn". (S1)
- Rule 4.24: the data holder must not "(c) offer additional or alternative services as part of the authorisation process; (d) include or refer to other documents." (S1) The CDS friction standard forbids the same in the authentication flow (`DESIGN.md`, Separation from the CDR authentication).
- The site has "no link to the authentication and consent pages, and those pages have no link to the site" (`DESIGN.md`, Separation from the CDR authentication).
- CX Guideline 3AU 19: "data holders should provide instructions for how to review sharing arrangements". The guidelines do not require a link to the dashboard. (S6)

**Recommendation:** text with no link, because rule 4.23(1)(g) requires the instructions and rule 4.24(d) forbids a reference to a different document: "You can stop the sharing at any time. Sign in to <brand> online banking and go to Data sharing, or ask <brand>." It names both methods of rule 4.25. JUDGEMENT: a description of the dashboard is the instruction that rule 4.23(1)(g) asks for, not a different service. The OTP page and the customer number page do not change.

## 6. A customer with no open account, and the open question from milestone 3

### Facts

- The ACCC: when a consumer closes all accounts, "any existing authorisations ... will expire (see CDR Rule 4.26(1)(c)). The obligation to provide a consumer dashboard will also end (see CDR Rules 1.13 and 1.15(1), and Schedule 3, clause 2.3)." (S13)
- Rule 1.10B(1)(b) and Schedule 3, clause 2.1: eligibility needs an open account that "can be accessed online". (S1)
- The ACCC permits a data holder to show an account that cannot be shared, for "a more consistent experience with a non-CDR digital customer portal". (S8)
- When the last open account closes, each arrangement of the customer ends (`DESIGN.md`, End of eligibility). A customer with no open account thus has no current arrangement, and no withdrawal is possible.
- The open question "Closed accounts on a new consent" (`DESIGN.md:1327`) is about the account selection of the consent page: the ACCC says that a closed account that is not shared "should show ... but they must not be selectable" (S13). It is not about the dashboard.

### Options for a customer with no open account

| | A. The same dashboard, with the ended consents | B. No dashboard; a statement only |
|---|---|---|
| Rule | More than the rule requires, which S8 permits for accounts | The minimum |
| Code | No special case | A special case on the dashboard page |
| Consumer | Sees why each consent ended at the close | Loses the history of their consents |

**Recommendation: A.** No rule forbids it, it has no special case, and the accounts page already says that the customer cannot share data. The dashboard of a customer with no consent says that the customer has no consent. Keep the open question "Closed accounts on a new consent" out of milestone 4: it changes the consent page, not the dashboard. On the dashboard, show a closed account of an arrangement with its closed status.

## 7. Other items for the scope

- CDR policy (rule 7.2(8)): the site is an online service by which the bank deals with consumers, and the data holder has no CDR policy. Options: put "A CDR policy of the data holder" out of scope in `DESIGN.md` Scope, or add a static page that says that the bank is a demonstration and has no policy. **Recommendation:** out of scope. A policy is a legal document of a real entity, and the footer already says that the bank is not real.
- Prominence (rule 1.15(1)(f)): a link "Data sharing" in the header of the site when the consumer is signed in, next to the accounts link (`DESIGN.md`, Sign-in and sign-out).
- The CDR authentication does not change. A session of the site is still not a step of the CDR authentication (`DESIGN.md`, Separation from the CDR authentication).
- Offline customers: the ACCC says that a data holder "must offer a dashboard to consumers without online account access" (S14, energy sector). In the lab, each customer can sign in to the site (any password works), so each customer has a dashboard.

## 8. Telemetry

### Facts

- `DESIGN.md:1062`: no customer number, name, email address, account data, `accountId`, session ID or password in telemetry. `DESIGN.md:1019`: no span, metric or log of the site has a path value, a query string or a body, and "No log of an open or a close contains ... an arrangement ID".
- ADR 0006: no CDR data and no personally identifiable information in telemetry; route patterns, not raw URLs (`docs/adr/0006-telemetry.md`, Sensitive data).

### Analysis

- A `client_name` or an `org_name` names the software product or the brand of a business. It is not personal information on its own. With a customer of a session, "this consumer shares data with app X" is personal information about the consumer. A disclosure row with its time and data cluster is also information about the consumer.
- **Recommendation:** no arrangement ID, client name, brand name of the data recipient, data cluster of a consumer or disclosure time in a log of the dashboard or of the disclosure record. The same rule as the open and the close. The withdraw route is a route pattern in spans. A test checks it.

## 9. Tests

**Recommendations:**

- In-process test in `internal/root`, as `banksiteclose_test.go`, with no waits: create a registration, a customer with an account, and an arrangement with a grant. Call a resource API with the access token, then read the dashboard through the site API: it shows the consent, its data clusters with the first and the last disclosure time, the accounts and the end of the sharing. Withdraw it through the site API. Check that the arrangement has the end reason `WITHDRAWN_BY_CONSUMER` and the method `CONSUMER_DASHBOARD` in the read model and in the operator interface, that the real `authserver.DataHolder` verifier refuses the access token and the refresh token, and that the notice is `PENDING` (its job is in the transaction of the event).
- Site API tests: 404 for an arrangement of a different customer and for an unknown one, with no identifier in the `detail`; 409 for an ended arrangement; 403 with no session; an operator withdrawal still records `ALTERNATIVE_METHOD`.
- A domain test that `WithdrawArrangement` records each method, and an app test that the use case passes the method.
- Telemetry test: a dashboard read, a resource request and a withdrawal give no customer number, arrangement ID, client name or disclosure data in spans, metrics or logs.
- Component tests of the dashboard page and of the dialog for each preset brand at 320 CSS px, with the exact CDS message text, and the automated accessibility checks.
- Browser test against the lab: use an arrangement of the conformance suite, as `withdraw.spec.ts` of the console does. Find a customer with a current arrangement through the console API, sign in to the site with the customer number and any password, open Data sharing, stop the sharing with the confirmation, and see it in the ended group. It skips when the lab has no current arrangement. In CI, lab-web has no conformance data, so it skips there; the in-process test is the proof in CI.
- The conformance suite does not test the site (ADR 0019). Its withdrawal tests through the operator interface stay.

## 10. Changes to the `DESIGN.md`

- Scope (`DESIGN.md:35`, `48`): the dashboard moves to "In scope" in the online bank site. Out of scope: a CDR policy, CDR receipts, the accreditation number.
- Consent (`DESIGN.md:483`) and the template text: the new instructions to withdraw.
- End of an arrangement (`DESIGN.md:619-627`): a row "The consumer withdraws the consent on the dashboard | The time of the request | Yes", or a change of the operator row to "The consumer withdraws the consent (dashboard or operator)".
- Withdrawal by the consumer (`DESIGN.md:651-655`): two methods, rule 4.25(1)(a) and (b), and the two values.
- Operator interface (`DESIGN.md:735`) and Console (`DESIGN.md:742`, `752`): the second value; remove "The consumer dashboard stays out of scope".
- CDR arrangement (`DESIGN.md:502`): the arrangement records the brand name and the legal entity name of the registration at the confirmation (if B of section 2).
- A new section of the record of disclosures (if C of section 2), under Data of a consent.
- Online bank site: a new section "Consumer dashboard" with the list, the details, the withdrawal and the messages; site API rows `GET /api/consents` and `POST /api/consents/{arrangementId}/withdraw`; the NFRs for telemetry and tests; the intro (`DESIGN.md:852`) and the landing statement (`DESIGN.md:861`).
- Decisions: change "Instructions to withdraw" (`DESIGN.md:1171`), "Method of the consumer to withdraw" (`DESIGN.md:1186`) and "Method of a withdrawal" (`DESIGN.md:1204`). Add decisions for the disclosure record, the names of the data recipient, the past consents, a customer with no open account, the DSB text and the CDR policy.

## 11. Risks and decisions for the requirements PR

| Item | Detail |
|---|---|
| Record of disclosures | Rule 1.15(3)(f) and rule 7.9 need what was disclosed and when, per data cluster (S10, S11). The data holder records nothing now. Decide A, B or C of section 2. C adds a write to each successful resource request and a new table. |
| Names of the data recipient | The arrangement has only `clientName`. Recording `org_name` and `legal_entity_name` changes `ArrangementCreated`, which needs a lab reset. |
| Name of the new method value | `CONSUMER_DASHBOARD` is a new enumeration value in the operator interface, the console API and the read model. Propose it before use. |
| CDS message text | The DSB gives the text only as a suggestion (S2, S3). Recommend word for word; a paraphrase needs review. |
| Instructions on the consent page | The text names the online banking site with no link. Check it against rule 4.24(d) and the friction standard in review. |
| CDR policy | Rule 7.2(8) applies to the site. Recommend out of scope, stated in Scope. |
| Rule 1.15(1)(bb) and 7.9(1)(c) | UNVERIFIED: no specified Register entry found. |
| Milestone 3 not complete | The open of an account is not on `main`. Milestone 4 does not depend on it, but the two must not change the same lines of the site API table at the same time. |
| Browser test in CI | The journey needs a consent of the conformance suite, so it skips in CI. The in-process test is the CI proof. |

## Main findings

- Rule 1.15(3)(f) with rule 7.9 requires the dashboard to show what CDR data was disclosed, when, and to which data recipient, for each data cluster with the first and the last date. The data holder has no record of disclosures. This is the largest new item of milestone 4.
- The domain already records a withdrawal method in `ArrangementWithdrawn`. A second value `CONSUMER_DASHBOARD` meets ADR 0021, because no rule of the mock acts differently on the method. The notice and the token revocation need no change.
- The CDS requires two messages in the withdrawal (consequences, redundant data) and one on the dashboard (check with the data recipient). The DSB gives suggested text and leaves the wording to the data holder.
- Rule 1.15(3)(e) requires the time of the end of a past authorisation, so the dashboard shows past consents. The method and the reason of the end need not show.
- The ACCC says that the dashboard obligation ends with eligibility. Showing the dashboard to a customer with no open account is permitted and needs no special case.
- The consent page must give instructions to withdraw with no reference to a different document. Text that names the dashboard and the bank, with no link, meets rules 4.23(1)(g) and 4.24.
- The data holder has no CDR policy, and rule 7.2(8) applies to each online service.
