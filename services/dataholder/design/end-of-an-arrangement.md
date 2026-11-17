# Mock data holder: End of an arrangement

This area gives the end of a CDR arrangement: CDR arrangement revocation, withdrawal by the consumer, the end of the sharing period, the end of eligibility, the deletion of a registration, and the notice to the data recipient.

## End of an arrangement

- <a id="dh-0900"></a>**DH-0900** An arrangement ends at the earliest of the events in the table below. The table also gives the time of the end, and if the data holder sends a [notice to the data recipient](#dh-0924). ([CDR Rules][cdr-rules], rule 4.26(1) and (2); [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation])

| ID | Event | Time of the end | Notice to the data recipient | Source |
|---|---|---|---|---|
| <a id="dh-0901"></a>**DH-0901** | The [sharing of the arrangement ends](authorisation.md#dh-0834) | The end of the sharing | No | [CDR Rules][cdr-rules], rule 4.26(1)(d) and (e) |
| <a id="dh-0902"></a>**DH-0902** | The software product [revokes the arrangement](../../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0229) | The time of the request | No | [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation] |
| <a id="dh-0903"></a>**DH-0903** | The consumer [withdraws the authorisation](#dh-0912) on the consumer dashboard, or the operator withdraws it for the consumer | The time of the request | Yes | [CDR Rules][cdr-rules], rule 4.26(1)(a) and (b), rule 4.26A |
| <a id="dh-0904"></a>**DH-0904** | The customer has [no open account](#dh-0918) | The close of the last open account | Yes | [CDR Rules][cdr-rules], rule 4.26(1)(c), rule 4.26A; [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation] |
| <a id="dh-2618"></a>**DH-2618** | The software product [deletes its registration](registration.md#dh-0676) | The time of the request | No | [CDR Support: Guidance on Dynamic Client Registration][support-dcr] |
| <a id="dh-0905"></a>**DH-0905** | The software product becomes REMOVED on the Register ([the reaction to the status REMOVED](registration.md#dh-0690)) | The time at which the data holder gets the status | No | [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [CDS: Data Holder Responsibilities][cds-dh-responsibilities] |

- <a id="dh-0906"></a>**DH-0906** The sharing duration of an arrangement is one year at most. Thus the sharing of an arrangement always ends before the limit of 12 months of rule 4.26(1)(e). ([CDS: Requesting Sharing Duration][cds-sharing-duration]; [CDR Rules][cdr-rules], rule 4.26(1)(e))
- <a id="dh-0907"></a>**DH-0907** An arrangement that ended does not start again. ([CDR Rules][cdr-rules], rule 4.26(1))
- <a id="dh-0908"></a>**DH-0908** When an arrangement ends, the data holder revokes its refresh token and its access tokens at the same time. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation])
- <a id="dh-0909"></a>**DH-0909** When an arrangement ends before the end of its sharing, the data holder records the time and the event of the end. ([CDR Rules][cdr-rules], rule 9.3(1))
- <a id="dh-0910"></a>**DH-0910** An arrangement for a disclosure on one occasion ends at the end of its sharing, or when the software product revokes it. The data holder does not find when the data has been disclosed. ([CDR Rules][cdr-rules], rule 4.26(1)(f); [CDS: Requesting Sharing Duration][cds-sharing-duration]; [Judgement: a disclosure on one occasion ends with its sharing, or at a revocation](#dh-0950))

## CDR arrangement revocation

The authorisation server library gives [the CDR arrangement revocation endpoint](../../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0228). At a revocation, the data holder [ends the arrangement and revokes its tokens before the response](../../../pkg/authserver/design/tokens-of-an-arrangement.md#auth-0229).

- <a id="dh-0911"></a>**DH-0911** An arrangement with the sharing duration 0 is current until the [end of its sharing](authorisation.md#dh-0834). Until then, each event of the [end table](#dh-0900) ends it and [revokes its access token](#dh-0908). For example, a revocation 1 minute after the confirmation gives 204. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [CDR Rules][cdr-rules], rule 4.26(1)(f); [Judgement: an arrangement with the sharing duration 0 ends when its access token expires](#dh-2619))

## Withdrawal by the consumer

- <a id="dh-0912"></a>**DH-0912** The consumer can withdraw an authorisation on the [consumer dashboard](consumer-dashboard.md#dh-1924) of the online bank site. ([CDR Rules][cdr-rules], rule 4.25(1)(a); [CDR Support: Withdrawal of authorisation to disclose CDR data and notification][support-withdrawal-notification])
- <a id="dh-0913"></a>**DH-0913** When the consumer asks the brand to withdraw an authorisation, the operator withdraws it with the [operator interface](operator-interface.md#dh-0727) or the [operator console](operator-console-create.md#dh-1443). This is the simple alternative method of communication. ([CDR Rules][cdr-rules], rule 4.25(1)(b); [CDR Support: Withdrawal of authorisation to disclose CDR data and notification][support-withdrawal-notification])
- <a id="dh-0914"></a>**DH-0914** Only the consumer dashboard shows the messages of the CDS Withdrawal Standards. The operator interface and the operator console do not show them. ([CDS: Withdrawal Standards][cds-withdrawal-standards]; [Judgement: the withdrawal messages only on the consumer dashboard](#dh-0951))
- <a id="dh-0915"></a>**DH-0915** A withdrawal has its effect at the time of the request: the request of the consumer on the dashboard, or the request of the operator. ([CDR Rules][cdr-rules], rule 4.25(2), rule 4.26(1)(a) and (b))
- <a id="dh-0916"></a>**DH-0916** The two methods of withdrawal have the same effects: the [end of the arrangement](#dh-0903), the [revocation of its tokens](#dh-0908) and the [notice to the data recipient](#dh-0924). ([CDR Rules][cdr-rules], rule 4.25, rule 4.26A)
- <a id="dh-0917"></a>**DH-0917** The data holder records the method of each withdrawal. A withdrawal on the consumer dashboard has the method `CONSUMER_DASHBOARD`. A withdrawal through the operator interface or the operator console has the method `ALTERNATIVE_METHOD`. ([CDR Rules][cdr-rules], rule 9.3(1)(b) and rule 4.25(1); [CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method])

## End of eligibility

- <a id="dh-0918"></a>**DH-0918** When the operator, or the consumer on the online bank site, closes the last open account of a customer, each arrangement of the customer that did not end, ends. This includes an arrangement that has no account. ([CDR Rules][cdr-rules], rule 4.26(1)(c); [CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts])
- <a id="dh-0920"></a>**DH-0920** When the customer has a different open account, the close of an account does not end an arrangement. ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts])
- <a id="dh-0921"></a>**DH-0921** The arrangements end before the response to the close. Their time of the end is the time of the close. ([CDR Rules][cdr-rules], rule 4.26(1)(c))
- <a id="dh-0922"></a>**DH-0922** When two closes of the last two open accounts of a customer occur at the same time, the arrangements of the customer end. ([CDR Rules][cdr-rules], rule 4.26(1)(c))
- <a id="dh-0923"></a>**DH-0923** When the data holder stops after the close and before the arrangements end, the arrangements end after the next start, 1 minute or more after the close. Their time of the end is the time of the close. ([CDR Rules][cdr-rules], rule 4.26(1)(c); [Judgement: the arrangements end after the next start when a stop interrupts a close](#dh-0952))

## Notice to the data recipient

- <a id="dh-0924"></a>**DH-0924** For each event of the [end table](#dh-0900) with a notice, the data holder notifies the software product of the arrangement. ([CDR Rules][cdr-rules], rule 4.26A; [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation])
- <a id="dh-0925"></a>**DH-0925** The notice is a `POST` to `<recipient_base_uri>/arrangements/revoke`. The `recipient_base_uri` is the value of the current registration of the client of the arrangement. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [Judgement: the current recipient_base_uri, or the value at the confirmation](#dh-0953))
- <a id="dh-0926"></a>**DH-0926** When the client has no registration, the `recipient_base_uri` is the value of the registration at the confirmation. For example, the software product deleted its registration with `DELETE`. ([Judgement: the current recipient_base_uri, or the value at the confirmation](#dh-0953))
- <a id="dh-0927"></a>**DH-0927** The connection of a notice is TLS. The data holder trusts only the public CAs in its configuration for it, and it connects to any host, as for a `jwks_uri`. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [CDR Support: Egress policy for calls from DH to ADRs][support-egress])
- <a id="dh-0928"></a>**DH-0928** The request has `Authorization: Bearer` with a self-signed JWT. The data holder signs it with PS256 and a key of its JWKS. ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt])
- <a id="dh-0929"></a>**DH-0929** The bearer JWT has `iss` and `sub` with the data holder brand ID, a `jti` that is a new random RFC 4122 UUID, and `iat`. ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt])
- <a id="dh-0930"></a>**DH-0930** The `aud` of the bearer JWT is the URL of the endpoint. For example: `https://recipient.example/arrangements/revoke`. ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt]; [Judgement: the resource path in the aud of a notice, as the CDS says](#dh-0954))
- <a id="dh-0931"></a>**DH-0931** The `exp` of the bearer JWT is 5 minutes after its `iat`. ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt]; [Judgement: a lifetime of 5 minutes for the JWTs of a notice](#dh-0955))
- <a id="dh-0932"></a>**DH-0932** The body of a notice is a form with only `cdr_arrangement_jwt`. It is a JWT with `cdr_arrangement_id` and the claims of the bearer JWT, with a different `jti`, and the data holder signs it in the same way. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [Judgement: only cdr_arrangement_jwt in the body of a notice](#dh-0956))
- <a id="dh-0933"></a>**DH-0933** When the response is 204 or 422, the notice is complete. ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [Judgement: a notice is complete at 204 or 422](#dh-0957))
- <a id="dh-0934"></a>**DH-0934** The data holder makes a new attempt after one of these results. ([Judgement: a notice is complete at 204 or 422](#dh-0957))
  - a 5xx response
  - no response within 10 seconds
  - no connection
- <a id="dh-0935"></a>**DH-0935** When the response has a different status, the notice fails, with no new attempt. For example, a 401 response. ([Judgement: a notice is complete at 204 or 422](#dh-0957))
- <a id="dh-0936"></a>**DH-0936** The data holder makes a maximum of 3 attempts. The second attempt is 10 seconds after the first, and the third is 60 seconds after the second. After the third attempt, the notice fails. ([CDR Support: Data Recipient unavailability back-off pattern][support-backoff]; [Judgement: 3 attempts at most for a notice, also across a restart](#dh-0958))
- <a id="dh-0937"></a>**DH-0937** The arrangement ends before the first attempt of its notice. The notice does not change the end. ([CDR Rules][cdr-rules], rule 4.26(1))
- <a id="dh-0938"></a>**DH-0938** When the data holder stops while a notice is not complete and did not fail, the notice continues its attempts after the next start. The total stays at 3 attempts. ([Judgement: 3 attempts at most for a notice, also across a restart](#dh-0958))
- <a id="dh-0939"></a>**DH-0939** The data holder records each notice, its outcome and the time of the outcome. ([CDR Rules][cdr-rules], rule 9.3(1); [Judgement: a record of each notice and attempt, with no response values](#dh-0959))
- <a id="dh-0940"></a>**DH-0940** The data holder records the time and the result of each attempt when the attempt occurs, also while the notice is pending. ([Judgement: a record of each attempt when it occurs](#dh-0960))
- <a id="dh-0941"></a>**DH-0941** The result of an attempt is one of these. It contains no value of the response. ([CDR Rules][cdr-rules], rule 9.3(1); [Judgement: a record of each notice and attempt, with no response values](#dh-0959))
  - the status of the response
  - `timeout`
  - `no connection`
  - `tls error`
  - `internal error`, when the data holder cannot make the request

## Judgements

### <a id="dh-0950"></a>DH-0950 A disclosure on one occasion ends with its sharing, or at a revocation

- Kind: conflict
- Basis: Step 1. Rule 4.26(1)(f) expires the authorisation after the CDR data has been disclosed ([CDR Rules][cdr-rules]). The CDS requires that the data recipient calls the revocation endpoint after a successful collection, for a sharing duration of 24 hours or less ([CDS: Requesting Sharing Duration][cds-sharing-duration]).
- Reason: A data holder cannot know when a data recipient has all the data that it needs. The CDS rule gives the data holder a signal, and the end of the sharing is the limit.

### <a id="dh-2619"></a>DH-2619 An arrangement with the sharing duration 0 ends when its access token expires

- Kind: choice
- Basis: Step 3.1. Rule 4.26(1)(f) ends an authorisation for a disclosure on a single occasion after the disclosure ([CDR Rules][cdr-rules]). The CDS gives only an access token for the sharing duration 0 ([CDS: Requesting Sharing Duration][cds-sharing-duration]). A revocation of the arrangement must revoke its tokens ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]).
- Reason: The data holder cannot find when the data recipient has all the data, and the data recipient collects data only with its access token. An arrangement that ends at the confirmation keeps a valid token that no revocation can stop. When the arrangement and its token end at the same time, each end stops the token.

### <a id="dh-0951"></a>DH-0951 The withdrawal messages only on the consumer dashboard

- Kind: choice
- Basis: Step 3.3. The CDS Withdrawal Standards are for the withdrawal process with the consumer ([CDS: Withdrawal Standards][cds-withdrawal-standards]).
- Reason: In the lab, the operator acts for the brand. The operator interface and the operator console are not for the consumer.

### <a id="dh-0952"></a>DH-0952 The arrangements end after the next start when a stop interrupts a close

- Kind: choice
- Basis: Step 3.3. Rule 4.26(1)(c) gives no rule for a data holder that stops between the close and the end ([CDR Rules][cdr-rules]).
- Reason: A stop must not leave the arrangements of a consumer who is not eligible, because a second close is refused. The delay of 1 minute lets the request end the arrangements first.

### <a id="dh-0953"></a>DH-0953 The current recipient_base_uri, or the value at the confirmation

- Kind: choice
- Basis: Step 3.3. The CDS names the `recipient_base_uri` of the registration ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]). A software product changes it with a `PUT` after a change on the Register.
- Reason: The current value follows that change. The value at the confirmation gives a target after a `DELETE` of the registration.

### <a id="dh-0954"></a>DH-0954 The resource path in the aud of a notice, as the CDS says

- Kind: conflict
- Basis: Step 1. The CDS requires the resource path of the endpoint in the `aud` ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt]). An ACCC article says that the `aud` is the `RecipientBaseUri` ([CDR Support: Guidance on revocation or withdrawal of consent][support-revocation-guidance]).
- Reason: The CDS is the first source.

### <a id="dh-0955"></a>DH-0955 A lifetime of 5 minutes for the JWTs of a notice

- Kind: choice
- Basis: Step 3.3. The CDS requires `exp` and gives no lifetime ([CDS: Self-signed JWT Client Authentication][cds-self-signed-jwt]).
- Reason: Each attempt has new JWTs, so a short lifetime is sufficient. The CDS example has 5 minutes.

### <a id="dh-0956"></a>DH-0956 Only cdr_arrangement_jwt in the body of a notice

- Kind: choice
- Basis: Step 3.3. The CDS requires the JWT method, and the duplicate form parameter `cdr_arrangement_id` is optional ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]).
- Reason: One form of the value is the least complex, and each data recipient must accept it.

### <a id="dh-0957"></a>DH-0957 A notice is complete at 204 or 422

- Kind: choice
- Basis: Step 3.3. The CDS gives 204 for success, and 422 for an arrangement that the data recipient does not have or that it revoked ([CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]).
- Reason: In both cases, the data recipient knows that the arrangement ended. A new attempt can help only when the cause is temporary.

### <a id="dh-0958"></a>DH-0958 3 attempts at most for a notice, also across a restart

- Kind: choice
- Basis: Step 2. The ACCC asks for a reasonable effort, for example a wait of a few seconds after the first failure, and not an exhaustive retry ([CDR Support: Data Recipient unavailability back-off pattern][support-backoff]).
- Reason: The notice is a civil penalty provision ([CDR Rules][cdr-rules], rule 4.26A), so a restart must not lose it. A restart gives no new attempts, so that a data holder that restarts again and again does not send a notice with no limit.

### <a id="dh-0959"></a>DH-0959 A record of each notice and attempt, with no response values

- Kind: choice
- Basis: Step 3.3. Rule 9.3(1) requires a record of each notice, and gives no detail ([CDR Rules][cdr-rules]).
- Reason: The attempts show what the data holder did when a notice fails. A result has no value of the response, because the response comes from a different participant.

### <a id="dh-0960"></a>DH-0960 A record of each attempt when it occurs

- Kind: choice
- Basis: Step 3.3. No source gives a time for the record of an attempt ([CDR Rules][cdr-rules], rule 9.3(1)).
- Reason: The operator and the developer of the data recipient must see the attempts of a notice that is in progress, and a restart must not lose them. A record at the outcome only shows no attempt until the third attempt, which is more than 70 seconds after the first.

[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cds-dh-responsibilities]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holder-responsibilities
[cds-self-signed-jwt]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#self-signed-jwt-client-authentication
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
[cds-withdrawal-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#withdrawal-standards
[support-backoff]: https://cdr-support.zendesk.com/hc/en-us/articles/360004056455
[support-closed-accounts]: https://cdr-support.zendesk.com/hc/en-us/articles/14519397304079
[support-dcr]: https://cdr-support.zendesk.com/hc/en-us/articles/6828251436431
[support-egress]: https://cdr-support.zendesk.com/hc/en-us/articles/4403543156751
[support-revocation-guidance]: https://cdr-support.zendesk.com/hc/en-us/articles/6828747049359
[support-withdrawal-method]: https://cdr-support.zendesk.com/hc/en-us/articles/4411292647823
[support-withdrawal-notification]: https://cdr-support.zendesk.com/hc/en-us/articles/900001658546
