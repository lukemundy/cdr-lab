# Mock data recipient: Consent

This area gives the selection of a bank, the consent page, the redirection notice, the record of each consent and the end of a consent. A member gives two consents together: a collection consent and a use consent. The [authorisation](authorisation.md) at the bank follows the consent.

## Selection of a bank

- <a id="dr-0800"></a>**DR-0800** The home page has an action to connect a bank. It opens the bank selection page. ([Judgement: a landing page that says what the app does](web-app.md#dr-0252))
- <a id="dr-0801"></a>**DR-0801** The bank selection page lists [each bank that the web app offers](participation.md#dr-0719), in the alphabetical sequence of the brand names. Each item has the logo and the brand name of the bank. ([CDS: Get Data Holder Brands][cds-get-dh-brands])
- <a id="dr-0802"></a>**DR-0802** When the member selects a bank, the web app shows the consent page for that bank. ([CDR Rules][cdr-rules], rule 4.11(1))
- <a id="dr-0803"></a>**DR-0803** Each page and each operation of the API that connects a bank needs a session and the member details. ([the member details before each other page](sign-up.md#dr-0321))

## Content of the consent page

The consent page gives the information of [CDR Rules][cdr-rules], rule 4.11(3), for the consents of the brand. The brand has no outsourced service provider, no sponsor, no fee, no direct marketing and no de-identification, so rule 4.11(3)(ca), (d), (da), (e), (f) and (i) give no information.

- <a id="dr-0804"></a>**DR-0804** The consent page shows the brand name of the data recipient, and the accreditation number from [the Register](participation.md#dr-0715). ([CDR Rules][cdr-rules], rule 4.11(3)(a) and (b))
- <a id="dr-0805"></a>**DR-0805** The consent page shows the logo and the brand name of the selected bank, and says that the brand asks to collect the data from that bank. ([CDR Rules][cdr-rules], rule 4.11(1)(a))
- <a id="dr-0806"></a>**DR-0806** For each data cluster in the table below, the consent page shows the data cluster language and the permission language of the CDS, word for word. ([CDS: Data Language Standards][cds-data-language]; [the data language on the authorisation page of the data holder](../../dataholder/design/authorisation.md#dh-0803))

| ID | Data cluster | Permissions | Scope | Selection | Source |
|---|---|---|---|---|---|
| <a id="dr-0807"></a>**DR-0807** | Account name, type and balance | Name of account; Type of account; Account balance | `bank:accounts.basic:read` | Required. The page has no control to remove it. | [CDS: Banking Language][cds-banking-language]; [Judgement: the account cluster is required and the transaction cluster is optional](#dr-0851) |
| <a id="dr-0808"></a>**DR-0808** | Transaction details | Incoming and outgoing transactions; Amounts; Dates; Descriptions of transactions; Who you have sent money to and received money from | `bank:transactions:read` | Optional. A checkbox that is not selected when the page opens. | [CDS: Banking Language][cds-banking-language]; [Judgement: the account cluster is required and the transaction cluster is optional](#dr-0851) |

- <a id="dr-0809"></a>**DR-0809** The consent page asks for no other data cluster. ([CDR Rules][cdr-rules], rule 1.8 and rule 4.12(2); [Judgement: the account cluster is required and the transaction cluster is optional](#dr-0851))
- <a id="dr-0810"></a>**DR-0810** The consent page states the use of the data: the brand shows the accounts, the balances and the transactions to the member in the web app. It states no other use. ([CDR Rules][cdr-rules], rule 4.11(1)(aa)(ii); [Judgement: a collection consent and a use consent on one page](#dr-0850))
- <a id="dr-0811"></a>**DR-0811** For each data cluster, the consent page says why the brand needs it. For example: "We need your account names, types and balances to show your accounts". ([CDR Rules][cdr-rules], rule 4.11(3)(c))
- <a id="dr-0812"></a>**DR-0812** The consent page says why the brand needs the sharing period that the member selects, and that the first collection gets the transactions of up to 90 days before the connection. ([CDR Rules][cdr-rules], rule 4.11(3)(c)(i); [the window of the first collection](data-collection.md#dr-1011))
- <a id="dr-0813"></a>**DR-0813** The consent page says that the brand collects the data one time each day, and also when the member asks for new data. ([CDR Rules][cdr-rules], rule 1.14(3)(e)(ii); [the schedule of the collection runs](data-collection.md#dr-1000))
- <a id="dr-0814"></a>**DR-0814** The consent page says that the member can withdraw the consent at any time, and how. ([CDR Rules][cdr-rules], rule 4.11(3)(g); [Judgement: a withdrawal only at the bank until the dashboard](#dr-0854))
- <a id="dr-0815"></a>**DR-0815** The consent page says that the brand deletes the data when it no longer needs it, and that it does not de-identify the data. ([CDR Rules][cdr-rules], rule 4.11(3)(h)(i) and rule 4.17(1)(a); [Judgement: no election to delete, because the brand deletes all redundant data](#dr-0853))
- <a id="dr-0816"></a>**DR-0816** The consent page has no election to delete redundant data. ([CDR Rules][cdr-rules], rule 4.16(3); [Judgement: no election to delete, because the brand deletes all redundant data](#dr-0853))
- <a id="dr-0817"></a>**DR-0817** The text of the consent page is in plain language, and it meets the accessibility target of a consumer UI. ([CDS: Consent Standards][cds-cx-consent]; [plain language in a consumer UI](../../../web/packages/ui/design/content.md#ui-0142); [the accessibility target](web-app.md#dr-0218))

## Sharing period

- <a id="dr-0818"></a>**DR-0818** The consent page asks the member to select one sharing period from these values. No value is selected when the page opens. ([CDR Rules][cdr-rules], rule 4.11(1)(b)(i) and rule 4.12(1); [Judgement: four sharing periods up to 12 months, with none selected](#dr-0852))
  - 1 month (30 days)
  - 3 months (90 days)
  - 6 months (180 days)
  - 12 months (365 days)
- <a id="dr-0819"></a>**DR-0819** When the member selects a sharing period, the consent page shows the date on which the consent expires. ([CDR Rules][cdr-rules], rule 4.11(1)(b); [a date in the time zone of the browser](../../../web/packages/ui/design/content.md#ui-0147))
- <a id="dr-0820"></a>**DR-0820** The `sharing_duration` of the authorisation request is the selected period in seconds. For example, 12 months is 31,536,000 seconds. ([CDS: Requesting Sharing Duration][cds-sharing-duration])

## Express consent

- <a id="dr-0821"></a>**DR-0821** The consent page has one primary action, "I agree", and a control to cancel. No control of the page is selected for the member, except the required data cluster. ([CDR Rules][cdr-rules], rule 4.11(1)(c) and Note 1)
- <a id="dr-0822"></a>**DR-0822** When the member agrees with no sharing period selected, the page shows a field error at the sharing period, and the data recipient records no consent. ([CDR Rules][cdr-rules], rule 4.11(1)(b); [the field errors of a problem](../../../pkg/platform/problem/DESIGN.md#prob-0003))
- <a id="dr-0823"></a>**DR-0823** When the member agrees, the data recipient records [the consent](#dr-0831). When the member cancels, it records nothing, and the web app shows the home page. ([CDR Rules][cdr-rules], rule 9.3(2)(a))

## Redirection notice

- <a id="dr-0824"></a>**DR-0824** After the member agrees, the web app shows a redirection notice before the browser goes to the bank. ([CDS: Consent Standards][cds-cx-consent]; [Judgement: a redirection notice on its own page](#dr-0855))
- <a id="dr-0825"></a>**DR-0825** The redirection notice names the bank. It says that the member confirms their identity at the bank with a one-time password, and that the bank does not ask for a password. ([CDS: Consent Standards][cds-cx-consent]; [CDS: Common Authentication Standards][cds-cx-authentication])
- <a id="dr-0826"></a>**DR-0826** The redirection notice has one primary action, "Continue to <bank brand name>", and a control to cancel. The action [starts the authorisation flow](authorisation.md#dr-0900). ([CDS: Consent Standards][cds-cx-consent]; [Judgement: a redirection notice on its own page](#dr-0855))

## Record of a consent

- <a id="dr-0827"></a>**DR-0827** The data recipient records each consent and its life as events in one event stream for each consent. ([Judgement: event-sourced state (ADR 0003)][adr-0003]; [Judgement: one stream for each consent, with one event for each collection run](#dr-0856))
- <a id="dr-0828"></a>**DR-0828** Each event type records one fact. ([Judgement: one business fact for each event type (ADR 0021)][adr-0021])
- <a id="dr-0829"></a>**DR-0829** An event contains only identifiers, times, codes, counts and the choices of the consent. It contains no personal information and no CDR data. For example, it contains no name, email address, IP address, account ID, amount, description or token. ([CDR Rules][cdr-rules], rule 9.3 and rule 9.4; [Judgement: one stream for each consent, with one event for each collection run](#dr-0856))
- <a id="dr-0830"></a>**DR-0830** Each event has the time when the data recipient recorded it. When the fact occurred at a different time, the event also has that time. ([CDR Rules][cdr-rules], rule 9.3(3))
- <a id="dr-0831"></a>**DR-0831** The stream of a consent records these facts. ([CDR Rules][cdr-rules], rule 9.3(2)(a), (c), (e) and rule 1.18(b); [Judgement: one business fact for each event type (ADR 0021)][adr-0021])
  - the member gave the consent: the consent ID, the member ID, the brand ID of the bank, the scopes, the sharing period and the version of the consent text
  - the bank gave the authorisation: the `cdr_arrangement_id` and the time at which the consent expires
  - the authorisation flow ended with no authorisation: the error code
  - [a collection run succeeded](data-collection.md#dr-1015), or [a collection run failed](data-collection.md#dr-1016)
  - the consent expired, or the consent was replaced
  - the data recipient deleted the collected data and the tokens of the consent
- <a id="dr-0832"></a>**DR-0832** Each version of the text of the consent page has an identifier, and the repository keeps the text of each version. ([CDR Rules][cdr-rules], rule 9.3(2)(h))
- <a id="dr-0833"></a>**DR-0833** For each authorisation flow, the data recipient records the member, the time, the client IP address, the scopes and the sharing period in a table, not in an event. ([CDS: Authentication Flows][cds-authentication-flows]; [no personal information in an event](#dr-0829))

## End of a consent

- <a id="dr-0834"></a>**DR-0834** The collection consent and the use consent have the same period. They expire together at the end of the sharing period, from the time of the authorisation. ([CDR Rules][cdr-rules], rule 4.14(1)(d); [Judgement: a collection consent and a use consent on one page](#dr-0850))
- <a id="dr-0835"></a>**DR-0835** Within 1 hour after a consent expires, the data recipient deletes the collected data and the tokens of the consent, and records the deletion. ([CDR Rules][cdr-rules], rule 1.18 and rule 7.13; [Competition and Consumer Act 2010][cdr-act], section 56EO(2))
- <a id="dr-0836"></a>**DR-0836** When a consent expires at the end of its sharing period, the data recipient does not call the CDR Arrangement Revocation endpoint of the bank. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation])
- <a id="dr-0837"></a>**DR-0837** A consent with no authorisation 30 minutes after the member agreed gives no connection and no collection. The data recipient then refuses the callback of its flow. ([Judgement: an authorisation flow of 30 minutes at most](authorisation.md#dr-0951))

## Judgements

### <a id="dr-0850"></a>DR-0850 A collection consent and a use consent on one page

- Kind: choice
- Basis: Step 1. An accredited person can ask for a collection consent and a use consent, and can bundle the requests ([CDR Rules][cdr-rules], rule 4.3(2) and rule 4.3(2A), Note 1).
- Reason: The data has one use, so one page with one action is the least complex. The two consents have the same data, the same period and the same record.

### <a id="dr-0851"></a>DR-0851 The account cluster is required and the transaction cluster is optional

- Kind: choice
- Basis: Step 1. The accredited person can let the consumer select the types of CDR data, or ask for agreement to the types that it presents ([CDR Rules][cdr-rules], rule 4.11(1)(a)). It asks for no more data than it needs ([CDR Rules][cdr-rules], rule 1.8).
- Reason: The app cannot show an account without its name, type and balance, so that cluster is required. The app works without the transactions, so the member selects them. The app shows no account number, no feature and no customer details, so it asks for no detailed account scope, no customer scope and no `profile` scope.

### <a id="dr-0852"></a>DR-0852 Four sharing periods up to 12 months, with none selected

- Kind: choice
- Basis: Step 1. The consumer selects the period, or agrees to a period that the accredited person presents, and the period is 12 months or less ([CDR Rules][cdr-rules], rule 4.11(1)(b) and rule 4.12(1)). The CDS limits a sharing duration to 31,536,000 seconds ([CDS: Requesting Sharing Duration][cds-sharing-duration]).
- Reason: A short list is easier to use than a free entry. 365 days is 12 months or less from each start date, and it is the CDS limit. No value is selected, so the member makes an active choice, and the lab can test each value.

### <a id="dr-0853"></a>DR-0853 No election to delete, because the brand deletes all redundant data

- Kind: choice
- Basis: Step 1. The election to delete and its instructions are "in accordance with rule 4.16" ([CDR Rules][cdr-rules], rule 4.11(1)(e) and rule 4.11(3)(h)). Rule 4.16 does not apply when the accredited person has a general policy to delete redundant data and says so when it asks for the consent ([CDR Rules][cdr-rules], rule 4.16(3)).
- Reason: Each brand of the data recipient deletes all redundant data, and the consent page says so. Thus the page has no election, the data recipient records no election, and the later dashboard has no election.

### <a id="dr-0854"></a>DR-0854 A withdrawal only at the bank until the dashboard

- Kind: gap
- Basis: Step 1. A consumer can withdraw a consent through the dashboard of the accredited person or in writing ([CDR Rules][cdr-rules], rule 4.13(1)). The consumer can also withdraw the authorisation at the data holder ([the dashboard of the data holder](../../dataholder/design/consumer-dashboard.md)).
- Reason: The dashboard and the withdrawal of the data recipient come in a later project. Until then, the consent page says that the member stops the sharing on the dashboard of the bank, and a later collection run finds that [the arrangement ended](data-collection.md#dr-1020).

### <a id="dr-0855"></a>DR-0855 A redirection notice on its own page

- Kind: choice
- Basis: Step 1. Data recipients notify consumers of the redirection before the authentication ([CDS: Consent Standards][cds-cx-consent]). A `request_uri` is valid for 10 to 90 seconds ([CDS: Request Object Submission][cds-request-object-submission]).
- Reason: A page of its own makes the notice clear after a long consent page. The data recipient sends the pushed authorisation request only when the member continues, so the `request_uri` does not expire while the member reads.

### <a id="dr-0856"></a>DR-0856 One stream for each consent, with one event for each collection run

- Kind: choice
- Basis: Step 1. The accredited data recipient keeps records of the consents, of their amendments and withdrawals, and of the collections of CDR data, with the time of each record and of its event ([CDR Rules][cdr-rules], rule 9.3(2) and rule 9.3(3)). The records stay after the deletion of the redundant data.
- Reason: One event for each run, not for each call, and one scheduled run each day give approximately 365 events in a consent of 12 months. A load of that stream completes within the limit of [the load of a stream](non-functional-requirements.md#dr-0016) with no snapshot. Events with no personal information and no CDR data can stay for the records after a deletion.

[adr-0003]: ../../../docs/adr/0003-event-sourcing.md
[adr-0021]: ../../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[cdr-act]: https://www.legislation.gov.au/C2004A00109/latest/text
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-authentication-flows]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authentication-flows
[cds-banking-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#banking-and-non-bank-lending-language
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[cds-cx-consent]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_consent-standards
[cds-data-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-language-standards-common
[cds-get-dh-brands]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands
[cds-request-object-submission]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#request-object-submission
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
