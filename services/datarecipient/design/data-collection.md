# Mock data recipient: Data collection

This area gives the collection runs of a connection: when they occur, the calls to the bank, the storage of the collected data, the record of each run and the failure of a run. A collection run is one set of calls to the bank that gets the current data of one connection.

## Schedule

- <a id="dr-1000"></a>**DR-1000** Each current connection has one scheduled collection run each day, in the low traffic period. The low traffic period is from 00:00 to 06:00 in the time zone `Australia/Sydney`. ([CDS: Data Recipient Requirements][cds-dr-requirements]; [CDS: NFRs, Definitions][cds-definitions]; [Judgement: one scheduled run each day in the low traffic period](#dr-1050))
- <a id="dr-1001"></a>**DR-1001** Each connection has its own time for its scheduled run, at a random offset in the low traffic period. The time does not change for the life of the connection. ([CDS: Data Recipient Requirements][cds-dr-requirements]; [Judgement: one scheduled run each day in the low traffic period](#dr-1050))
- <a id="dr-1002"></a>**DR-1002** A scheduled run is unattended. Its calls have no `x-fapi-customer-ip-address` and no `x-cds-client-headers`. ([the headers of a call with no customer present](../../../pkg/cdrclient/DESIGN.md#cdrc-0093))
- <a id="dr-1003"></a>**DR-1003** After a new connection, the data recipient starts the first collection run at once. The first run is customer present. ([CDS: NFRs, Definitions][cds-definitions])
- <a id="dr-1004"></a>**DR-1004** The page of a connection has an action to get new data. The action starts a collection run that is customer present. ([CDS: NFRs, Definitions][cds-definitions]; [Judgement: a run that the member starts, at most each 15 minutes](#dr-1051))
- <a id="dr-1005"></a>**DR-1005** When a run of the connection that is customer present started less than 15 minutes before, the data recipient starts no run, and the page says the time from which the member can get new data. ([Judgement: a run that the member starts, at most each 15 minutes](#dr-1051))
- <a id="dr-1006"></a>**DR-1006** The calls of a run that is customer present have `x-fapi-customer-ip-address` and `x-cds-client-headers` from the request of the member. The `x-cds-client-headers` has only [the headers that the library permits](../../../pkg/cdrclient/DESIGN.md#cdrc-0093). For example, it never has the session cookie. ([CDS: HTTP Headers][cds-http-headers])
- <a id="dr-1007"></a>**DR-1007** Each call of a run has `x-fapi-auth-date` with the time of the last sign-in of the member. The data recipient keeps that time with the member, so that a restart does not change it. ([CDS: HTTP Headers][cds-http-headers])
- <a id="dr-1008"></a>**DR-1008** A connection has one run at a time. When a run of the connection is in progress, a request for a new run starts no run. ([CDS: Data Recipient Requirements][cds-dr-requirements])

## Calls of a run

- <a id="dr-1009"></a>**DR-1009** A run starts with [a refresh of the access token](../../../pkg/cdrclient/DESIGN.md#cdrc-0085), except the first run, which uses the access token of the token response. ([CDS: Refresh Token][cds-refresh-token]; [CDS: Token Expiry][cds-token-expiry])
- <a id="dr-1010"></a>**DR-1010** A run calls these endpoints of the bank, each with the page size 1000. It calls Get Transactions For Account for each account only when the consent has `bank:transactions:read`. ([CDS: Endpoint Version Schedule][cds-version-schedule]; [CDS: Pagination][cds-pagination]; [one version for each endpoint](../../../pkg/cdrclient/DESIGN.md#cdrc-0090))

  | Endpoint | `x-v` |
  |---|---|
  | [Get Accounts][cds-get-accounts] | 3 |
  | [Get Bulk Balances][cds-get-bulk-balances] | 2 |
  | [Get Transactions For Account][cds-get-transactions] | 2 |

- <a id="dr-1011"></a>**DR-1011** The first run of a connection calls Get Transactions For Account with no `oldest-time`, so that the bank gives its default window of 90 days. ([CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: a first window of 90 days, and later windows with an overlap of 7 days](#dr-1052))
- <a id="dr-1012"></a>**DR-1012** A later run calls Get Transactions For Account with an `oldest-time` 7 days before the start of the last successful run. ([CDS: Get Transactions For Account][cds-get-transactions]; [Judgement: a first window of 90 days, and later windows with an overlap of 7 days](#dr-1052))
- <a id="dr-1013"></a>**DR-1013** Performance: an unattended run makes 100 calls or fewer with each access token, and 5 calls each second or fewer. When the run needs more calls, it refreshes the access token before the next call. ([CDS: Traffic Thresholds][cds-traffic-thresholds])

## Storage of the collected data

- <a id="dr-1014"></a>**DR-1014** The data recipient keeps the collected accounts, balances and transactions in tables, with the consent ID as a key, not in the events. ([no CDR data in an event](consent.md#dr-0829); [CDR Rules][cdr-rules], rule 1.18)
- <a id="dr-1017"></a>**DR-1017** After a successful run, the stored accounts and balances of the connection are those of the run. The data recipient deletes the stored data of each account that the response of Get Accounts does not have. ([CDR Rules][cdr-rules], rule 1.18 and rule 7.13)
- <a id="dr-1018"></a>**DR-1018** After a successful run, the stored transactions of an account in the window of the run are those of the run. The data recipient keeps the stored transactions before the window. ([Judgement: a first window of 90 days, and later windows with an overlap of 7 days](#dr-1052))
- <a id="dr-1019"></a>**DR-1019** A run that fails changes no stored data. ([Judgement: a failed run keeps the data and waits for the next run](#dr-1054))

## Record of a run

- <a id="dr-1015"></a>**DR-1015** The stream of the consent records each successful run with its start time, its end time, its trigger, and the number of accounts and of transactions that it got. The trigger is one of `first`, `scheduled` and `member`. ([CDR Rules][cdr-rules], rule 9.3(2)(e) and rule 9.3(3); [Judgement: one stream for each consent, with one event for each collection run](consent.md#dr-0856))
- <a id="dr-1016"></a>**DR-1016** The stream of the consent records each failed run with its start time, its end time, its trigger and one of these reasons. ([CDR Rules][cdr-rules], rule 9.3(2)(e); [Judgement: one business fact for each event type (ADR 0021)][adr-0021])
  - `arrangement_ended`: [the arrangement ended](#dr-1020)
  - `software_product_not_active`: the bank refused the software product with `unauthorized_client` or `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`
  - `bank_unavailable`: the bank did not answer, or it gave a status of 500 or more
  - `response_not_valid`: a response was not valid for the CDS

## Failure of a run

- <a id="dr-1020"></a>**DR-1020** The arrangement ended when the token endpoint gives `invalid_grant`, or when a resource call gives 403 with `urn:au-cds:error:cds-all:Authorisation/RevokedConsent`. The connection then needs attention. ([CDS: Error Codes][cds-error-codes]; [the grant ended](../../../pkg/cdrclient/DESIGN.md#cdrc-0086); [Judgement: a connection that needs attention until the dashboard](#dr-1053))
- <a id="dr-1021"></a>**DR-1021** A connection that needs attention has no more scheduled runs, and its page has no action to get new data. ([Judgement: a connection that needs attention until the dashboard](#dr-1053))
- <a id="dr-1022"></a>**DR-1022** The home page shows a connection that needs attention with a message that the bank stopped the sharing, and the offer to [connect the bank again](authorisation.md#dr-0919). ([Judgement: a connection that needs attention until the dashboard](#dr-1053))
- <a id="dr-1023"></a>**DR-1023** After a run fails for a different reason, the data recipient makes no new attempt before the next scheduled run. The member can start a run. ([CDS: Data Recipient Requirements][cds-dr-requirements]; [Judgement: a failed run keeps the data and waits for the next run](#dr-1054))
- <a id="dr-1024"></a>**DR-1024** When a run that the member started fails, the page of the connection says that the brand could not get new data from the bank, and shows the data of the last successful run. ([Judgement: a failed run keeps the data and waits for the next run](#dr-1054))
- <a id="dr-1025"></a>**DR-1025** No collection run starts for a consent that expired or that was replaced. ([CDR Rules][cdr-rules], rule 4.14(1))

## Judgements

### <a id="dr-1050"></a>DR-1050 One scheduled run each day in the low traffic period

- Kind: choice
- Basis: Step 1. Data recipients schedule unattended calls out of the high traffic period, which is from 6am to midnight, and avoid short bursts ([CDS: Data Recipient Requirements][cds-dr-requirements]; [CDS: NFRs, Definitions][cds-definitions]). Unattended traffic has a threshold of 20 sessions each day for each customer and software product ([CDS: Traffic Thresholds][cds-traffic-thresholds]). The CDS gives no time zone for the period.
- Reason: One run each day is far below the threshold, and it gives a record of approximately 365 runs in a consent of 12 months. Most lab users and banks are in the eastern states, so the period is in Sydney time. A random time for each connection stops a burst at midnight.

### <a id="dr-1051"></a>DR-1051 A run that the member starts, at most each 15 minutes

- Kind: choice
- Basis: Step 3.3. The CDS gives customer present traffic no limit of sessions each day ([CDS: Traffic Thresholds][cds-traffic-thresholds]), and asks data recipients to minimise the traffic to data holders ([CDS: Data Recipient Requirements][cds-dr-requirements]).
- Reason: A member who made a payment wants to see it, so the app lets the member ask for new data. A limit of one run each 15 minutes keeps the traffic and the stream small, and a lab user does not wait long.

### <a id="dr-1052"></a>DR-1052 A first window of 90 days, and later windows with an overlap of 7 days

- Kind: choice
- Basis: Step 3.2. With no `oldest-time`, Get Transactions For Account gives the 90 days before `newest-time` ([CDS: Get Transactions For Account][cds-get-transactions]). A pending transaction can change or become posted after the last run. The collection relates to a period that is not more than the use needs ([CDR Rules][cdr-rules], rule 1.8(1)(b)).
- Reason: The default window of the bank is enough for a view of recent spending, so the first run asks for no more. An overlap of 7 days gets each change of a recent transaction again, and a replacement of the window removes a pending transaction that is not in the new response.

### <a id="dr-1053"></a>DR-1053 A connection that needs attention until the dashboard

- Kind: gap
- Basis: Step 1. When a data holder notifies the withdrawal of an authorisation, the collection consent expires ([CDR Rules][cdr-rules], rule 4.14(3)). The data recipient has no endpoint for the notice ([no endpoint for the end of an arrangement](participation.md#dr-0751)).
- Reason: An error at the token endpoint or a revoked consent shows that the arrangement ended, but it is not the notice of the rule. The data recipient stops the collection, keeps the consent, and offers the member a new connection. With the endpoint, the notice makes the consent expire.

### <a id="dr-1054"></a>DR-1054 A failed run keeps the data and waits for the next run

- Kind: choice
- Basis: Step 1. Data recipients design their services to be resilient when a data holder refuses a call ([CDS: Data Recipient Requirements][cds-dr-requirements]).
- Reason: The data of the last successful run is still the best data that the member has. A new attempt at once adds traffic to a bank that has a problem, and the member can start a run.

[adr-0021]: ../../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-definitions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#definitions
[cds-dr-requirements]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-recipient-requirements
[cds-error-codes]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#error-codes
[cds-get-accounts]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-accounts
[cds-get-bulk-balances]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-bulk-balances
[cds-get-transactions]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account
[cds-http-headers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#http-headers
[cds-pagination]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#pagination
[cds-refresh-token]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#refresh-token
[cds-token-expiry]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#token-expiry
[cds-traffic-thresholds]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#traffic-thresholds
[cds-version-schedule]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#endpoint-version-schedule
