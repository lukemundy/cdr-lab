# Data holder: the end of data sharing

- Research date: 2026-10-16
- Consumer Data Standards (CDS) version: 1.36.0. It is still the current version: the newest release tag is `v1.36.0` (4 December 2025). [S2]
- CDR Rules version: Compilation No. 10, https://www.legislation.gov.au/F2020L00094/2025-03-04/text. [S3]
- Purpose: facts that the earlier notes do not have, for the requirements of the end of a consent: token revocation, CDR arrangement revocation, a withdrawal by the consumer, the notice to the data recipient, the end of the sharing period, the end of eligibility and the reaction to Register statuses. The note [CDS data holder InfoSec and consent](2026-10-06-cds-data-holder-infosec-and-consent.md) ("the InfoSec note"), sections 4.6, 5.3, 5.4, 7.5, 7.6 and 8, is still correct for this scope. This note adds to it and does not replace it.

Check the sources again before you depend on a fact; the rules and standards change.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site; the archive URL redirects to it while 1.36.0 is current) | https://consumerdatastandardsaustralia.github.io/standards/ |
| S2 | CDS source repository at the tag `v1.36.0` (file paths given for each fact) | https://github.com/ConsumerDataStandardsAustralia/standards/tree/v1.36.0 |
| S3 | Competition and Consumer (Consumer Data Right) Rules 2020, Compilation No. 10 | https://www.legislation.gov.au/F2020L00094/latest/text |
| S4 | RFC 7009, OAuth 2.0 Token Revocation | https://www.rfc-editor.org/rfc/rfc7009 |
| S5 | CDR Support Portal articles (IDs given for each fact; read through the help centre API) | https://cdr-support.zendesk.com/hc/en-us/articles/<id> |

## 1. Token revocation

- "Data Holders MUST implement a Token Revocation endpoint as described in section 2 of [RFC7009]." The endpoint "allows a Data Recipient Software Product to invalidate its tokens as required to allow for token clean up". "Revocation of Refresh Tokens and Access Tokens MUST be supported." MTLS, with client authentication, and no bearer token. [S2 `slate/source/includes/security/endpoints/_token_revocation.md`]
- The request has `token` and an optional `token_type_hint`. The server verifies that the token "was issued to the client making the revocation request". "The invalidation takes place immediately". [S4 section 2.1]
- For a refresh token, the server "SHOULD also invalidate all access tokens based on the same authorization grant". For an access token, the server "MAY revoke the respective refresh token as well". [S4 section 2.1]
- The response is 200 also for a token that is not valid: "invalid tokens do not cause an error response". The error `unsupported_token_type` is for a server that does not support the revocation of the token type. Other errors are those of RFC 6749, section 5.2. [S4 section 2.2]
- DSB: "If the ADR revokes an access token, then attempts to use the invalidated access token to make a data request, it must fail. If, however, the ADR obtains a fresh access token by using their refresh token, that is a perfectly valid data request and should succeed." [S5: 6828747049359] Thus the revocation of an access token does not end the grant.
- The token revocation endpoint is not the method to withdraw a consent. The CDR arrangement revocation endpoint is. [S5: 900003387786]
- The CDS does not say if the revocation of the refresh token of an arrangement ends the arrangement (InfoSec note, section 4.6: **UNVERIFIED**). No source found on 2026-10-16 says it.

## 2. CDR arrangement revocation by the data recipient

The InfoSec note, section 5.3, gives the request, the 204 and 422 responses and the synchronous revocation. These facts are new:

- The 422 response "MUST use the error structure defined in the Payload Conventions", with `urn:au-cds:error:cds-all:Authorisation/InvalidArrangement`. "The error detail is the CDR Arrangement ID". [S2 `slate/source/includes/security/endpoints/_cdr_arrangement_revocation.md`] [S2 `slate/source/includes/standards/_errors.md`]
- For an arrangement with `sharing_duration` 0, "the DH is expected to return an error if the revocation endpoint is called for an arrangement that has expired". [S5: 5729181240591]
- After the revocation, "The DH's solution must prevent any data disclosure against the cdr_arrangement_id." The clean-up of the refresh token "is an implementation decision for the DH". [S5: 4412301531663]
- "If consent is revoked, the access token is not guaranteed to be revoked. The expectation is that an active access token must not be honoured because the consent has been withdrawn." [S5: 6828747049359]

## 3. Responses to a data request after the end of a consent

Two sources do not agree:

- The CDS error list has 403 `urn:au-cds:error:cds-all:Authorisation/RevokedConsent`: "The consumer's consent is no longer authorised (for example revoked or expired) and the requested resource will not be provided. This error code SHOULD be supported for authenticated APIs." [S2 `slate/source/includes/standards/_errors.md`]
- The DSB: when "the consent has been withdrawn or cancelled, refresh token revoked", "a 401 Unauthorised Error with a corresponding oAuth error response should be returned by the DH". [S5: 5729181240591]
- The CDS also requires that "Consent revocation MUST also revoke associated refresh and/or access tokens" [S2 `_cdr_arrangement_revocation.md`], and FAPI 1.0 requires the RFC 6750 errors for a revoked token (the 2026-10-14 note, section 2). RFC 6750 gives 401 `invalid_token` for a token that is "expired, revoked, malformed, or invalid".
- Thus a token of a revoked consent is a revoked token. The CDS code is a SHOULD, and the CDS does not say when it applies instead of the RFC 6750 error.

## 4. The end of the sharing period

- An authorisation expires at the earliest of a list of events [S3 rule 4.26(1)]. For this scope:
  - (b) a withdrawal on the dashboard: "when the authorisation was withdrawn";
  - (a) a withdrawal by "a simple alternative method of communication" (rule 4.25(1)(b)): when the data holder gives effect to it, and at the latest "2 business days after the data holder received the communication";
  - (c) "if the CDR consumer ceases to be eligible in relation to the data holder";
  - (d) when the data holder gets the notice of the withdrawal from the data recipient;
  - (e) 12 months after the authorisation was given;
  - (f) "if the authorisation was for disclosure of CDR data on a single occasion—after the CDR data has been disclosed";
  - (g) for a specified period, "the end of" that period.
- `sharing_duration`: a value of 24 hours or less is a "one-time collection", and "a Refresh Token SHOULD be provided". A value of zero or absent is "once off access", and "only an Access Token (without a Refresh Token) will be provided". "If a Refresh Token is issued for one-time collection the Data Recipient Software Product MUST call the Data Holder's revocation endpoint after successful collection of the CDR data." The CDS does not say which revocation endpoint. [S2 `slate/source/includes/security/_request_object.md`]
- With `sharing_duration` 0, "the cdr_arrangement_id and consent are immediately expired upon creation". [S5: 5729181240591] No source says how long the data recipient can use the access token of such a consent. The access token has its own expiry (2 to 10 minutes, [S1 #access-token]).
- No source says if a data holder can give data with an access token that it issued before the end of the period, after the period ends. Rule 4.26(1)(g) says that the authorisation has then expired.

## 5. Eligibility

- A consumer is eligible when they are 18 or older (or not an individual), and "an account holder or a secondary user for an account with the data holder that is open". [S3 rule 1.10B(1)] For banking, the account must also be "set up in such a way that, at that time, it can be accessed online". [S3 Schedule 3, clause 2.1]
- ACCC: "Where a CDR consumer closes all of their accounts with a data holder (or no longer holds any accounts with the data holder that can be accessed online), they will cease to be an eligible CDR consumer in relation to that data holder. This means any existing authorisations given by the consumer to the data holder will expire (see CDR Rule 4.26(1)(c))." When the consumer closes only some accounts and stays eligible, "existing authorisations with the data holder remain in place". [S5: 14519397304079]
- An expiry for this reason is not one of the exceptions to the notice to the data recipient (section 6). Thus the data holder notifies the data recipient.

## 6. The notice to the data recipient

The InfoSec note, section 5.4, gives the endpoint, the transport, the JWT method and the retry guidance. These facts are new or more exact:

- The data holder notifies the software product "when an authorisation is withdrawn or otherwise expires", except when "The withdrawal was initiated via the Data Recipient", when "The authorisation expires at its natural expiry time", and at "Invalidation of the authorisation due to a change in the Data Recipient or Software Product status on the Register". [S2 `_cdr_arrangement_revocation.md`]
- The bearer token is a self-signed JWT. Its `iss` and `sub` are "the client ID of the bearer", which is "the 'Data Holder Brand ID' as issued by CDR Register". `jti` and `exp` are required, `iat` is optional, and the JWT is for one use. [S2 `slate/source/includes/security/_client_authentication.md`]
- `aud`: "From July 31st 2022: The Resource Path for the endpoint being accessed MUST be used." The CDS example has the full URL of the endpoint. [S2 `_client_authentication.md`] The DSB article says that the `aud` is the `RecipientBaseUri` and "not the full path of the endpoint". [S5: 6828747049359] The article is from before the date in the CDS rule, and the CDS is the first source.
- The body has `cdr_arrangement_jwt`, a signed JWT with `cdr_arrangement_id`, which "SHOULD also include" the self-signed JWT claims. The data holder "MAY additionally send" `cdr_arrangement_id` as a form parameter. [S2 `_cdr_arrangement_revocation.md`]
- The responses of the data recipient endpoint are 204 for success and 422 `InvalidArrangement` for an arrangement that "has previously been revoked", that is not of the client, or that does not exist. [S2 `_cdr_arrangement_revocation.md`] [S5: 5729181240591]
- Retry: "reasonable effort", at the discretion of the data holder. "Attempting to communicate the consent withdrawal via the CDR Arrangement Revocation endpoint by waiting a few seconds upon first failure would be suitable." After an outage, the data recipient is expected to check each consent with the token or introspection endpoint. [S5: 360004056455]
- Records: the data holder keeps records of "notifications of withdrawal of consents" and of withdrawals of authorisations. [S3 rule 9.3(1)] (InfoSec note, section 7.8.)

## 7. Withdrawal by the consumer

- Rule 4.25 permits a withdrawal on the dashboard or with "a simple alternative method of communication". For the second method, the data holder gives effect "as soon as practicable, and in any case within 2 business days". [S3 rule 4.25]
- The CX Withdrawal Standards require the data holder, "As part of the withdrawal process", to advise the consumer to review the consequences with the data recipient, and to "inform the consumer about the handling of redundant data and the right to delete". [S2 `slate/source/includes/cx_standards/withdrawals.md`] These are rules for a process with the consumer, for example the dashboard. The CDS does not say how they apply to an alternative method.

## 8. Register statuses

The InfoSec note, section 8, gives the table of data holder responsibilities and the cache rules. These facts are new or more exact:

- "CDR data MAY only be disclosed to active Data Recipients." [S2 `slate/source/includes/security/_participant_statuses.md.erb`]
- The data holder polls Get Data Recipients Statuses, Get Software Products Statuses and Get Data Recipients, and "cache these for use during requests for Consumer Data". "both the Data Recipient and Software Product statuses SHOULD be referenced". [S2 `_participant_statuses.md.erb`]
- In each row of the responsibilities table, the actions of the data holder are the same for each data recipient status with the same software product status: ACTIVE gives all actions, INACTIVE gives only the withdrawal, REMOVED gives the invalidation and the clean-up. The status mapping makes sure that a software product is not ACTIVE when its data recipient is not ACTIVE. [S2 `_participant_statuses.md.erb`]
- The error for a data request is 403 `urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive`. "The error detail SHOULD contain the current status of the ADR software product." [S2 `_errors.md`]
- UserInfo: "If an ADR is inactive and makes a request to the UserInfo endpoint, then it should receive a 401 response." [S5: 5729181240591]
- Authorisation endpoint: "When the client browser interacts with the authorise endpoint and encounters an invalid software product then the most appropriate response is to redirect back to the Software Product, which is the ADR application, providing error information." No error code is given. [S5: 5729181240591]
- Sequence of checks: "It is advisable to address coarse-grained errors first. For example, validate the ADR status before the Software Product status, or ensure the Access Token (AT) is valid before validating consent." [S5: 5729181240591]
- REMOVED: "data sharing must cease immediately", "tokens issued for any consent where the Software Product is 'Removed' must not be honoured", and the client ID "should never be re-issued or re-enabled". The invalidation and the clean-up can be "real-time, a nightly batch process, or a process run every time the CDR Register cache is refreshed". [S5: 900001657906]
- SUSPENDED (software product INACTIVE): the consents "remain valid during the period of suspension", the withdrawal must still work, and "If and when the suspension is lifted, data sharing can resume". [S5: 6828747049359]
- The CDS gives no error code for the PAR endpoint and the token endpoint for a software product that is not ACTIVE.

## 9. The authorisation server library

Background only: a library is not a source for a requirement.

- `luikyv/go-oidc` v0.25.0 has `WithTokenRevocation`. The revocation of a refresh token "always invalidates the underlying grant and related access tokens". For a JWT access token, "revocation is only effective when `WithTokenRevocationRevokeGrantOnAccessToken` is enabled, since JWTs are not stored server-side", and that option revokes the grant. Thus the library cannot revoke one JWT access token and keep its grant (section 1).
