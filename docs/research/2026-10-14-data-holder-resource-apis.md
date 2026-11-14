# Data holder resource APIs: additions for the APIs of a consent

- Research date: 2026-10-14
- Consumer Data Standards (CDS) version: 1.36.0. It is still the current version: the newest release tag is `v1.36.0` (4 December 2025), and no decision that is later than Decision 376 is closed. [S1] [S2]
- Purpose: facts that the notes of 2026-10-06 do not have, for the requirements of the customer, account, balance and transaction APIs and of the UserInfo endpoint. The notes [CDS banking and common APIs](2026-10-06-cds-banking-and-common-apis.md) and [CDS data holder InfoSec and consent](2026-10-06-cds-data-holder-infosec-and-consent.md) are still correct for this scope. This note adds to them and does not replace them.

Check the sources again before you depend on a fact; the rules and standards change.

## Primary sources

| ID | Source | URL |
|---|---|---|
| S1 | Consumer Data Standards 1.36.0 (live site; the archive URL redirects to it while 1.36.0 is current) | https://consumerdatastandardsaustralia.github.io/standards/ |
| S2 | CDS source repository: releases and issues | https://github.com/ConsumerDataStandardsAustralia/standards |
| S3 | FAPI 1.0 Part 1: Baseline (Final) | https://openid.net/specs/openid-financial-api-part-1-1_0.html |
| S4 | FAPI 1.0 Part 2: Advanced (Final) | https://openid.net/specs/openid-financial-api-part-2-1_0.html |
| S5 | OpenID Connect Core 1.0 | https://openid.net/specs/openid-connect-core-1_0.html |
| S6 | CDR Support Portal articles (IDs given for each fact; read through the help centre API) | https://cdr-support.zendesk.com/hc/en-us/articles/<id> |

## 1. ID permanence and the pairwise identifier

The 2026-10-06 note could not read article 5505351487631. The help centre API returned it on 2026-10-14 (last update 10 October 2025). [S6: 5505351487631]

- "ID Permanence and Pairwise Identifiers are two different things." The PPID is the `sub`. ID permanence "pertains to the resource identifiers", for example the account ID.
- The PPID is different for each user and for each sector identifier.
- Account IDs and other resource IDs are different for each consumer and for each **software product**. They "must persist (be the same) for each arrangement created in the context of the combination of detail specified".
- Thus two software products with the same sector identifier get the same `sub` and different account IDs for one customer.

## 2. Resource server rules of FAPI 1.0

FAPI 1.0 Advanced section 6.2.1 requires the provisions of Baseline section 6.2.1 and RFC 8705. [S4 section 6.2.1] The resource server: [S3 section 6.2.1]

- "shall accept access tokens in the HTTP header as in Section 2.1 of OAuth 2.0 Bearer Token Usage", and "shall not accept access tokens in the query parameters";
- "shall verify that the access token is neither expired nor revoked";
- "shall verify that the scope associated with the access token authorizes access to the resource it is representing";
- "shall only return the resource identified by the combination of the entity implicit in the access and the granted scope and otherwise return errors as in Section 3.1 of RFC6750";
- "shall send the server date in HTTP Date header";
- "shall set the response header x-fapi-interaction-id" to the request value or a new UUID, and "shall log the value of x-fapi-interaction-id in the log entry";
- "shall not reject requests with a x-fapi-customer-ip-address header containing a valid IPv4 or IPv6 address".

The note of section 6.2.1 says that the resource server "can use OAuth Token Introspection" to find the entity and the scope. Advanced section 8.2 says that the protected resources "shall not accept a bearer access token" and accept only sender-constrained tokens. [S4 section 8.2]

## 3. UserInfo endpoint

- "Data Holders MUST support a UserInfo endpoint." The endpoint uses MTLS and a bearer token, with no client authentication. The requirements are those of OpenID Connect Core section 5.3. [S1 #userinfo-endpoint]
- "Data Holders MUST support the profile scope as described in section 5.4 of [OIDC]." The claims that "MUST be supported" are `sub`, `acr`, `auth_time`, `name`, `given_name`, `family_name` and `updated_at`. `auth_time` "SHOULD NOT be returned via the UserInfo endpoint". `email`, `email_verified`, `phone_number`, `phone_number_verified` and `address` "MAY be supported". [S1 #scopes-and-claims]
- The list of configuration fields says that `userinfo_endpoint` is necessary "only if OIDC Hybrid Flow is supported". [S1 #openid-provider-configuration-endpoint] This does not agree with the MUST in the endpoint section. The hybrid flow is retired from 12 May 2025.
- The UserInfo response is JSON, unless the client registered a signed or encrypted response. [S5 section 5.3.2]
- The UserInfo endpoint returns CDR data. A call from a data recipient that is not active "should fail, and receive a 401 response". [S6: 5729181240591]

## 4. Header guidance titles

The 2026-10-06 note cites these articles. Their titles, for links: 4405773134863 "Format for x-fapi-auth-date"; 4404953530767 "DH requirement for x-cds-client-headers"; 4410664891919 "FAPI header obligations for Data Holders"; 14519397304079 "Closed accounts in the banking and non-bank lenders sectors". [S6]

## 5. Open items

- The library `luikyv/go-oidc` v0.25.0 has `Provider.Introspect`, which gives the token information and the grant of an access token. This note does not check if it detects a revoked grant. Background only: a library is not a source for a requirement.
