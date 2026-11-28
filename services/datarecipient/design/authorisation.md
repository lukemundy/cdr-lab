# Mock data recipient: Authorisation

This area gives the authorisation flow of the data recipient at a bank, after [the consent](consent.md): the pushed authorisation request, the redirect of the browser, the callback, the tokens and the connection that the flow makes. The [CDR client library](../../../pkg/cdrclient/DESIGN.md) gives the protocol of each call.

## Start of the flow

- <a id="dr-0900"></a>**DR-0900** When the member continues from the redirection notice, the data recipient [pushes an authorisation request](../../../pkg/cdrclient/DESIGN.md#cdrc-0060) to the bank. It then redirects the browser to [the authorisation endpoint](../../../pkg/cdrclient/DESIGN.md#cdrc-0064) with 303. ([CDS: Authentication Flows][cds-authentication-flows]; [RFC 9126, section 4][rfc9126-authorise])
- <a id="dr-0901"></a>**DR-0901** The data recipient gets the endpoints of the bank from the OpenID Provider configuration at the `infosecBaseUri` of the brand on the Register. ([CDS: Participant Endpoints][cds-participant-endpoints]; [the endpoints from the configuration](../../../pkg/cdrclient/DESIGN.md#cdrc-0022))
- <a id="dr-0902"></a>**DR-0902** The `scope` of the request object is `openid` and the scopes of the data clusters that the member agreed to. It has no other scope. ([CDS: Scopes and Claims][cds-scopes-and-claims]; [Judgement: the account cluster is required and the transaction cluster is optional](consent.md#dr-0851))
- <a id="dr-0903"></a>**DR-0903** The `redirect_uri` of the request object is [the redirect URI of the software product](participation.md#dr-0702). ([CDS: Request Object, Data Recipient Software Products][cds-request-object-dr])
- <a id="dr-0904"></a>**DR-0904** The data recipient keeps the `state`, the `nonce`, the code verifier, the consent and the session of each flow on the server. The browser gets none of them, except the `state` in the authorisation response. ([RFC 9700, section 4.7.1][rfc9700-csrf]; [the browser gets no token](web-app.md#dr-0202))
- <a id="dr-0905"></a>**DR-0905** When the pushed authorisation request fails, the web app shows a page that says that the bank cannot be connected now. The stream of the consent records [an authorisation flow with no authorisation](consent.md#dr-0831). ([RFC 9126, section 2.3][rfc9126-errors])

## Callback

- <a id="dr-0906"></a>**DR-0906** The page of the redirect URI sends the `response` parameter to the API with `POST`, and then removes the query from the address of the page. ([JARM, section 2.3.1][jarm-query]; [Judgement: the callback page posts the response to the API](#dr-0950))
- <a id="dr-0907"></a>**DR-0907** The data recipient accepts a callback only in the session that started its flow. When the session is different or ended, the web app shows the sign-in page, and the data recipient does not complete the flow. ([RFC 9700, section 4.7.1][rfc9700-csrf]; [each page with member data needs a session](sign-in.md#dr-0412))
- <a id="dr-0908"></a>**DR-0908** The data recipient [validates the authorisation response](../../../pkg/cdrclient/DESIGN.md#cdrc-0070) and [its `state`](../../../pkg/cdrclient/DESIGN.md#cdrc-0071). When the validation fails, it completes no flow and records no authorisation. ([FAPI 1.0 Advanced, section 5.2.3.2][fapi-advanced-client-jarm])
- <a id="dr-0909"></a>**DR-0909** The data recipient accepts each flow one time only. A second callback of a flow gets the same page as a callback that is not valid. ([RFC 6749, section 10.5][rfc6749-code-reuse])
- <a id="dr-0910"></a>**DR-0910** When a valid response has `error`, the stream of the consent records the error code. The web app shows a page that says that the bank did not authorise the sharing, with an action to go to the home page. ([JARM, section 2.1][jarm-response]; [Judgement: a page for each outcome of the flow](#dr-0952))
- <a id="dr-0911"></a>**DR-0911** When the error is `access_denied`, the page says that the member cancelled at the bank, and that the brand collects no data. ([CDR Support: Error scenarios and responses][support-error-scenarios]; [Judgement: a page for each outcome of the flow](#dr-0952))

## Tokens

- <a id="dr-0912"></a>**DR-0912** When the response has a code, the data recipient [exchanges the code](../../../pkg/cdrclient/DESIGN.md#cdrc-0080) and [validates the ID token](../../../pkg/cdrclient/DESIGN.md#cdrc-0083). ([CDS: Tokens][cds-tokens])
- <a id="dr-0913"></a>**DR-0913** The data recipient keeps the access token, the refresh token, the expiry time of each, and the `cdr_arrangement_id` of the consent in a table, not in an event. It does not keep the ID token or the `sub`. ([Judgement: tokens in a table, not in the events](#dr-0953); [CDR Rules][cdr-rules], rule 1.8)
- <a id="dr-0914"></a>**DR-0914** The expiry time of the consent is the time of the token response and the `sharing_duration` of the request. ([CDS: Requesting Sharing Duration][cds-sharing-duration]; [CDR Rules][cdr-rules], rule 4.14(1)(d))
- <a id="dr-0915"></a>**DR-0915** When the token exchange or the validation of the ID token fails, the stream of the consent records an authorisation flow with no authorisation. The web app shows a page that says that the connection failed, and that the member can try again. ([CDS: Tokens][cds-tokens]; [Judgement: a page for each outcome of the flow](#dr-0952))

## Connection

- <a id="dr-0916"></a>**DR-0916** A consent with an authorisation is a connection. After the token response, the data recipient starts [the first collection run](data-collection.md#dr-1003) of the connection. ([CDR Rules][cdr-rules], rule 4.4)
- <a id="dr-0917"></a>**DR-0917** After the token response, the web app shows the home page with the new connection, and says that the brand is getting the data. ([Judgement: a page for each outcome of the flow](#dr-0952))
- <a id="dr-0918"></a>**DR-0918** A member can have more than one connection, also to the same bank. Each connection has its own consent and its own arrangement. ([the new arrangement for each authorisation](../../dataholder/design/authorisation.md#dh-0832))
- <a id="dr-0919"></a>**DR-0919** When a connection [needs attention](data-collection.md#dr-1020), the web app offers to connect the bank again. The new connection starts with a new consent page. ([Judgement: a new connection replaces a connection that needs attention](#dr-0954))
- <a id="dr-0920"></a>**DR-0920** When a new connection to a bank completes from that offer, the data recipient ends the consent of the connection that needed attention, with the reason "replaced". It [revokes the arrangement of that consent](../../../pkg/cdrclient/DESIGN.md#cdrc-0088) at the bank, and then deletes the collected data and the tokens of that consent. A 422 from the bank is a success. ([CDR Rules][cdr-rules], rule 1.18 and rule 4.18AA; [CDS: CDR Arrangement Revocation endpoint][cds-arrangement-revocation]; [Judgement: a new connection replaces a connection that needs attention](#dr-0954))

## Judgements

### <a id="dr-0950"></a>DR-0950 The callback page posts the response to the API

- Kind: choice
- Basis: Step 3.2. The bank redirects the browser to the redirect URI with the `response` in the query ([JARM, section 2.3.1][jarm-query]). The web app is a single-page application, and each change is not a `GET` request ([the requests that are not `GET`](sign-in.md#dr-0411)).
- Reason: A `POST` keeps the change out of a `GET`, and the API gets the response in a body that no log of a URL records. The page removes the query, so the response is not in the history of the browser.

### <a id="dr-0951"></a>DR-0951 An authorisation flow of 30 minutes at most

- Kind: choice
- Basis: Step 3.3. A flow of the data holder ends 10 minutes after the request ([the end of a flow of the data holder](../../../pkg/authserver/design/authorisation-request.md#auth-0103)), and its authorisation response is valid for 10 minutes ([the response lifetime](../../../pkg/authserver/design/authorisation-request.md#auth-0106)).
- Reason: 30 minutes covers the consent, the flow at the bank and the response. The data recipient can then free each flow that the member left.

### <a id="dr-0952"></a>DR-0952 A page for each outcome of the flow

- Kind: choice
- Basis: Step 3.3. The CDS gives no page of a data recipient after the authorisation.
- Reason: The member must know if the bank shares the data. A page for each outcome tells the member what occurred and what to do next.

### <a id="dr-0953"></a>DR-0953 Tokens in a table, not in the events

- Kind: choice
- Basis: Step 3.2. A token gives access to CDR data until it expires. The events stay for the records after the data recipient deletes the data ([one stream for each consent](consent.md#dr-0856)).
- Reason: A table lets the data recipient delete each token at the end of the consent. The `sub` and the ID token are not needed to collect or show the data, so the data recipient does not keep them.

### <a id="dr-0954"></a>DR-0954 A new connection replaces a connection that needs attention

- Kind: choice
- Basis: Step 3.3. The data holder does not support the amendment of an authorisation ([data holder non-goals](../../dataholder/design/README.md#non-goals)). A connection that needs attention has no arrangement that the data recipient can use.
- Reason: The member connects again to see current data, so the old consent has no use. An end of the old consent and a deletion of its data stops two copies of the same accounts.

[cds-arrangement-revocation]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-revocation-endpoint
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-authentication-flows]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authentication-flows
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-request-object-dr]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#security-profile_request-object_data-recipient-software-products
[cds-scopes-and-claims]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#scopes-and-claims
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
[cds-tokens]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#tokens
[fapi-advanced-client-jarm]: https://openid.net/specs/openid-financial-api-part-2-1_0.html#jarm-response-mode
[jarm-query]: https://openid.net/specs/oauth-v2-jarm.html#section-2.3.1
[jarm-response]: https://openid.net/specs/oauth-v2-jarm.html#name-the-jwt-response-document
[rfc6749-code-reuse]: https://www.rfc-editor.org/rfc/rfc6749#section-10.5
[rfc9126-authorise]: https://www.rfc-editor.org/rfc/rfc9126#section-4
[rfc9126-errors]: https://www.rfc-editor.org/rfc/rfc9126#section-2.3
[rfc9700-csrf]: https://www.rfc-editor.org/rfc/rfc9700#section-4.7.1
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
