# Mock data holder: Online bank site sign-in

This area gives the sign-up, the sign-in, the sign-out and the session of the online bank site, its separation from the authentication of a consumer, and the listener of the site.

## Sign-up

- <a id="dh-2000"></a>**DH-2000** The sign-up page asks for the first name, the last name and the email address, and for nothing else. The first name is optional. ([Judgement: the sign-up is a create of a customer of the operator](#dh-2100))
- <a id="dh-2001"></a>**DH-2001** The sign-up page asks for no password and no identity document. The data holder makes no identity check at a sign-up. ([Judgement: the sign-up is a create of a customer of the operator](#dh-2100))
- <a id="dh-2002"></a>**DH-2002** A sign-up creates the customer with the rules of [a create of a customer in the operator interface](operator-interface.md#dh-0720), for each value. For example, [the rule of a `lastName`](operator-interface.md#dh-0734) applies. ([Judgement: the sign-up is a create of a customer of the operator](#dh-2100))
- <a id="dh-2003"></a>**DH-2003** A sign-up gives the new customer a customer number with the rules of the customer numbers. ([the customer number](consumer-authentication.md#dh-1201); [Judgement: the sign-up is a create of a customer of the operator](#dh-2100))
- <a id="dh-2004"></a>**DH-2004** The operator interface and the operator console show a customer of a sign-up in the same way as a customer of the operator. The data holder keeps no record of the channel of a sign-up. ([Judgement: the sign-up is a create of a customer of the operator](#dh-2100))
- <a id="dh-2005"></a>**DH-2005** The sign-up page says that the data holder is a demonstration, and that the consumer must give only invented details. ([Judgement: no rule for the domain of the email address](#dh-2102))
- <a id="dh-2006"></a>**DH-2006** The sign-up page gives an example email address on the reserved domain `.test`. For example, `name@example.test`. ([RFC 6761, section 6.2][rfc6761-test]; [Judgement: no rule for the domain of the email address](#dh-2102))
- <a id="dh-2007"></a>**DH-2007** The data holder does not refuse an email address on a domain other than `.test`. ([Judgement: no rule for the domain of the email address](#dh-2102))
- <a id="dh-2008"></a>**DH-2008** When the data holder refuses a value of a sign-up, the page shows the error at its field and in an error summary. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="dh-2009"></a>**DH-2009** The text of each error of a sign-up is in plain language, and the site API gives it. For example, the text for a last name that is not valid is "Enter a last name of 1 to 100 characters". ([plain language in a consumer UI](../../../web/packages/ui/design/content.md#ui-0142); [the plain-language detail of a field error](online-bank-site-api.md#dh-1736))
- <a id="dh-2010"></a>**DH-2010** When a different customer of the brand has the email address, the page shows the refusal at the email address. The text says that the address is in use, and that the consumer can sign in or give a different address. ([the uniqueness of the email address](operator-interface.md#dh-0736))
- <a id="dh-2011"></a>**DH-2011** A successful sign-up starts a session for the new customer. ([Judgement: a session after a sign-up](#dh-2103))
- <a id="dh-2012"></a>**DH-2012** After a successful sign-up, the site shows the new customer number on a page of its own, with a copy button. ([Judgement: a session after a sign-up](#dh-2103))
- <a id="dh-2013"></a>**DH-2013** The customer number page says that the consumer signs in to the site with this number or with the email address. ([CDS: User Identifiers][cds-user-identifiers])
- <a id="dh-2014"></a>**DH-2014** The customer number page says that an app that gets the data of the consumer under the Consumer Data Right asks for the email address or this number. It also says that the app then asks for a one-time password that the brand sends to the email address. ([CDS: User Identifiers][cds-user-identifiers]; [the OTP by email](consumer-authentication.md#dh-1220))
- <a id="dh-2015"></a>**DH-2015** A new customer of a sign-up has no account. The data holder opens no account at a sign-up. For example, the consumer then [opens an account on the site](online-bank-site-accounts.md#dh-2210), or the operator opens one. ([Judgement: no account at a sign-up](#dh-2101))

## Sign-in and sign-out

- <a id="dh-2016"></a>**DH-2016** The sign-in page asks for the email address or the customer number in one field, and for a password. ([the user identifiers](consumer-authentication.md#dh-1200); [WCAG 2.2, SC 3.3.8][wcag-3-3-8])
- <a id="dh-2017"></a>**DH-2017** The identifier field of the sign-in page has `autocomplete="username"`. ([WCAG 2.2, SC 1.3.5][wcag-1-3-5])
- <a id="dh-2018"></a>**DH-2018** The password field of the sign-in page has `type="password"` and `autocomplete="current-password"`. ([WCAG 2.2, SC 1.3.5][wcag-1-3-5])
- <a id="dh-2019"></a>**DH-2019** The consumer can paste into each field of the sign-in page, and a password manager can fill each field. ([WCAG 2.2, SC 3.3.8][wcag-3-3-8])
- <a id="dh-2020"></a>**DH-2020** The sign-in page accepts any password. ([Judgement: the page does not send the password](#dh-2104))
- <a id="dh-2021"></a>**DH-2021** When the password field is empty, the sign-in page refuses the sign-in and says that any password works. It sends no request. ([Judgement: the page does not send the password](#dh-2104))
- <a id="dh-2022"></a>**DH-2022** The sign-in page does not send the password to the data holder. It sends only the user identifier. ([Judgement: the page does not send the password](#dh-2104))
- <a id="dh-2023"></a>**DH-2023** When the user identifier is the email address or the customer number of a customer, the data holder starts a session for that customer. ([the user identifiers](consumer-authentication.md#dh-1200))
- <a id="dh-2024"></a>**DH-2024** When no customer has the user identifier, the data holder refuses the sign-in. The page says "We could not sign you in. Check your email address or customer number". ([Judgement: the refusal of an unknown user identifier](#dh-2105))
- <a id="dh-2025"></a>**DH-2025** The refusal of an unknown user identifier is the same for an email address and a customer number. Thus the sign-in shows which email addresses and customer numbers exist. ([Judgement: the refusal of an unknown user identifier](#dh-2105))
- <a id="dh-2026"></a>**DH-2026** Each page that shows the data of a customer needs a session. For example, the accounts page needs a session. ([Judgement: one session cookie for a consumer UI with a sign-in (ADR 0014)][adr-0014])
- <a id="dh-2027"></a>**DH-2027** When the consumer opens a page that needs a session, and the browser has no valid session, the site shows the sign-in page. ([Judgement: one session cookie for a consumer UI with a sign-in (ADR 0014)][adr-0014])
- <a id="dh-2028"></a>**DH-2028** When a session ended, the sign-in page says that the data holder signed the consumer out. ([the session-ended type](../../../pkg/platform/problem/DESIGN.md#prob-0010); [the response to an ended session](online-bank-site-api.md#dh-1744))
- <a id="dh-2029"></a>**DH-2029** Each page of the site shows whether the consumer is signed in. ([CDR Rules][cdr-rules], rule 1.15(1)(f))
- <a id="dh-2030"></a>**DH-2030** When the consumer is signed in, the header of each page has these items. ([CDR Rules][cdr-rules], rule 1.15(1)(f): "prominently displayed and readily accessible")
  - a link to [the accounts page](online-bank-site-accounts.md#dh-2200)
  - a link "Data sharing" to [the consumer dashboard](consumer-dashboard.md#dh-1900)
  - a control to sign out
- <a id="dh-2031"></a>**DH-2031** After a successful sign-in, the site shows the accounts page. After a sign-out, the site shows the sign-in page. ([Judgement: the pages after a sign-in and a sign-out](#dh-2117))
- <a id="dh-2032"></a>**DH-2032** A sign-out ends the session. ([OWASP ASVS 5.0, 7.4.1][asvs-v7])

## Session

- <a id="dh-2033"></a>**DH-2033** The data holder keeps each session of the site on the server. ([RFC 10017, section 7.1][rfc10017-cookie-session]; [Judgement: one session cookie for a consumer UI with a sign-in (ADR 0014)][adr-0014])
- <a id="dh-2034"></a>**DH-2034** The session cookie has only the session ID. The browser gets no token and no customer number. ([Judgement: one session cookie for a consumer UI with a sign-in (ADR 0014)][adr-0014])
- <a id="dh-2035"></a>**DH-2035** The session ID is a random value from a cryptographic random source, with at least 128 bits of entropy. ([OWASP ASVS 5.0, 7.2.3][asvs-v7])
- <a id="dh-2036"></a>**DH-2036** Each sign-in and each sign-up starts a new session with a new session ID. ([OWASP ASVS 5.0, 7.2.4][asvs-v7]; [OWASP Session Management Cheat Sheet][owasp-session])
- <a id="dh-2037"></a>**DH-2037** When the request of a sign-in or a sign-up has a session cookie, the session of that cookie ends. ([OWASP Session Management Cheat Sheet][owasp-session])
- <a id="dh-2038"></a>**DH-2038** A session ends 24 hours after its start. ([OWASP ASVS 5.0, 7.1.1 and 7.1.2][asvs-v7]; [WCAG 2.2, SC 2.2.1][wcag-2-2-1]; [Judgement: no inactivity timeout](#dh-2107))
- <a id="dh-2039"></a>**DH-2039** A session has no inactivity timeout. ([Judgement: no inactivity timeout](#dh-2107))
- <a id="dh-2040"></a>**DH-2040** A customer can have more than one session at the same time, with no limit. ([Judgement: no inactivity timeout](#dh-2107))
- <a id="dh-2041"></a>**DH-2041** The data holder keeps at most 10,000 sessions that did not end. ([Judgement: sessions in the memory of the process](#dh-2106))
- <a id="dh-2042"></a>**DH-2042** When the data holder has 10,000 sessions that did not end, it refuses a sign-up with 503, and it creates no customer. ([Judgement: sessions in the memory of the process](#dh-2106))
- <a id="dh-2043"></a>**DH-2043** When the data holder has 10,000 sessions that did not end, it refuses a sign-in with 503. When the request of the sign-in has a valid session, the data holder ends that session and starts the new session. ([Judgement: sessions in the memory of the process](#dh-2106))
- <a id="dh-2044"></a>**DH-2044** After a session ends, its session ID gives no access. ([OWASP ASVS 5.0, 7.4.1][asvs-v7])
- <a id="dh-2045"></a>**DH-2045** When the data holder no longer has the customer of a session, the session is not valid. The data holder ends it at the next request of the session, also at a read of the session. ([Judgement: the session of a customer that the data holder does not have](#dh-2112))
- <a id="dh-2046"></a>**DH-2046** The name of the session cookie is `__Host-bank-session-<DATA_HOLDER_BRAND_ID>`, with the `dataHolderBrandId` of the instance. ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]; [Judgement: the name of the session cookie](#dh-2108))
- <a id="dh-2047"></a>**DH-2047** The session cookie has these attributes. ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]; [OWASP Session Management Cheat Sheet][owasp-session])
  - `Secure`
  - `HttpOnly`
  - `SameSite=Strict`
  - `Path=/`
- <a id="dh-2048"></a>**DH-2048** The session cookie has no `Domain`, no `Max-Age` and no `Expires`. Thus the browser removes the cookie when it closes. ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]; [Judgement: the name of the session cookie](#dh-2108))
- <a id="dh-2049"></a>**DH-2049** The response to a sign-out expires the session cookie with `Max-Age=0`. ([OWASP ASVS 5.0, 7.4.1][asvs-v7])
- <a id="dh-2050"></a>**DH-2050** The response to a sign-out has no `Clear-Site-Data` header. ([MDN: Clear-Site-Data][mdn-clear-site-data]; [Judgement: no `Clear-Site-Data` at a sign-out](#dh-2109))
- <a id="dh-2051"></a>**DH-2051** The only protection of the site against cross-site request forgery is [the refusal of cross-origin changes](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0005). The site has no CSRF token. ([OWASP CSRF Prevention Cheat Sheet][owasp-csrf]; [Judgement: CSRF protection by the origin check](#dh-2110))
- <a id="dh-2052"></a>**DH-2052** The sign-up, the sign-in and the sign-out are not `GET` requests. Thus the refusal of cross-origin changes also applies to them. ([OWASP CSRF Prevention Cheat Sheet][owasp-csrf]; [Judgement: CSRF protection by the origin check](#dh-2110))

## Separation from the CDR authentication

The sign-in of the site asks for a password, and the CDR authentication does not. These rules keep the two apart, so that the consumer does not confuse them.

- <a id="dh-2053"></a>**DH-2053** The sign-in page shows a note above the form. ([CDS: Common Authentication Standards][cds-cx-authentication]; [Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2054"></a>**DH-2054** The style of the note on the sign-in page is different from the style of the lock note of the consumer pages. ([Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2055"></a>**DH-2055** The note on the sign-in page says: "This is <brand name> online banking. This is a demonstration bank: any password works here, and you must not use a real password. When an app asks for your <brand name> data under the Consumer Data Right, you do not sign in here. <brand name> then asks only for your email address or customer number and a one-time password, and never for your password". ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Passwords; [Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2056"></a>**DH-2056** The title of the sign-in page is "Sign in - <brand name> online banking". ([the titles of the site](online-bank-site.md#dh-1831); [Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2057"></a>**DH-2057** The heading of the sign-in page is "Sign in to online banking". ([Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2058"></a>**DH-2058** The site and the consumer pages are on different origins. ([a listener of its own for the site](#dh-2063); [Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2059"></a>**DH-2059** The site has no link to the consumer pages. ([Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2060"></a>**DH-2060** The consumer pages have no link to the site. ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Friction)
- <a id="dh-2061"></a>**DH-2061** The consumer pages are the same when the browser has a session of the site. For example, they keep their titles and their headings. ([Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))
- <a id="dh-2062"></a>**DH-2062** A session of the site does not replace a step of the CDR authentication, and a CDR authentication does not start a session of the site. ([Judgement: the separation of the sign-in from the CDR authentication](#dh-2111))

## Listener

The site is a web UI. The [web UI library specification](../../../pkg/platform/consolehttp/DESIGN.md) gives the rules of its listener, its browser security, its cache headers and its routes. These rules are the choices of the site.

- <a id="dh-2063"></a>**DH-2063** The site is on a listener of its own. It is not on the public listener or on the listener of the operator console. ([Judgement: a listener of its own for the site](#dh-2113))
- <a id="dh-2064"></a>**DH-2064** The setting `BANK_SITE_ENABLED` [turns the listener of the site on](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0002). ([Judgement: the names of the site settings](#dh-2114))
- <a id="dh-2065"></a>**DH-2065** The setting `BANK_SITE_ALLOWED_HOSTS` is [the `Host` allowlist](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0003) of the site, with values in the format of the console allowlist. For example, `localhost:8185`. ([Judgement: the names of the site settings](#dh-2114))
- <a id="dh-2066"></a>**DH-2066** The listener of the site uses TLS, with the certificate and the key of the public listener. ([Judgement: TLS on the site](#dh-2115))
- <a id="dh-2067"></a>**DH-2067** The listener of the site uses [the TLS versions and cipher suites of the CDR listeners](api-standards.md#dh-0460). ([Judgement: TLS on the site](#dh-2115))
- <a id="dh-2068"></a>**DH-2068** The listener of the site does not request a client certificate. ([Judgement: TLS on the site](#dh-2115))
- <a id="dh-2069"></a>**DH-2069** The responses of the site have no `Strict-Transport-Security` header. ([RFC 6797, Appendix B][rfc6797-b]; [Judgement: no HSTS on the site](#dh-2116))
- <a id="dh-2070"></a>**DH-2070** The site sets one cookie: [the session cookie](#dh-2046). ([Judgement: one session cookie for a consumer UI with a sign-in (ADR 0014)][adr-0014])
- <a id="dh-2071"></a>**DH-2071** The listener of the site serves the styles, the logo and the mark of the brand at [the paths of the public listener](authorisation.md#dh-2090). ([Judgement: brand files at fixed paths under `/brand/`](authorisation.md#dh-2120))
- <a id="dh-2072"></a>**DH-2072** The listener of the site serves each photo of the brand at `/brand/<file name>`. For example, `/brand/hero.webp`. ([Judgement: brand files at fixed paths under `/brand/`](authorisation.md#dh-2120))
- <a id="dh-2073"></a>**DH-2073** The listener of the site serves the fonts of the fixed set of the data holder. ([the fixed set of fonts](brand.md#dh-0341))
- <a id="dh-2074"></a>**DH-2074** The brand files, the photos and the fonts on the listener of the site have [the `Host` allowlist](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0004) and [the response headers](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007) of the site. ([Judgement: a strict CSP with Trusted Types](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0020))
- <a id="dh-2075"></a>**DH-2075** The data holder starts the listener of the site only in the process that runs the APIs. ([Judgement: a listener of its own for the site](#dh-2113))
- <a id="dh-2076"></a>**DH-2076** A process that runs only the workers needs no setting of the site. ([a process needs only the settings that it uses](non-functional-requirements.md#dh-0010))

## Judgements

### <a id="dh-2100"></a>DH-2100 The sign-up is a create of a customer of the operator

- Kind: choice
- Basis: Step 3.3. The CDR Rules require records of authorisations and their withdrawals (rule 9.3), and no rule requires a record of how a person became a customer ([CDR Rules][cdr-rules]).
- Reason: One use case then gives the rules of the operator to the sign-up by construction. A field needs a rule or a reader that uses it ([ADR 0021][adr-0021]), and nothing uses the channel.

### <a id="dh-2101"></a>DH-2101 No account at a sign-up

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the online web site of a data holder.
- Reason: The consumer selects the product, as at a real bank. Thus a new customer of the site cannot give an authorisation until the customer has an open account ([CDR Rules][cdr-rules], rule 1.10B(1)(b)).

### <a id="dh-2102"></a>DH-2102 No rule for the domain of the email address

- Kind: choice
- Basis: Step 3.3. [RFC 6761, section 6.2][rfc6761-test] reserves `.test` for tests, and the operator interface has no rule for the domain.
- Reason: The mail sink of the lab delivers no email. A rule only for the site makes the sign-up different from the create of the operator, so only the text of the page asks for an invented address.

### <a id="dh-2103"></a>DH-2103 A session after a sign-up

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the online web site of a data holder.
- Reason: A sign-up that starts a session is one step less for the consumer. The page with the new customer number shows first, so the consumer still sees the number.

### <a id="dh-2104"></a>DH-2104 The page does not send the password

- Kind: choice
- Basis: Step 3.3. The site is lab-only and checks no password ([production gaps][production-gaps]).
- Reason: A person can type a real password from habit. When the page does not send it, no password goes to the data holder, its database or its telemetry. An API that refuses a `password` field proves this, and an API that ignores the field cannot.

### <a id="dh-2105"></a>DH-2105 The refusal of an unknown user identifier

- Kind: choice
- Basis: Step 3.3. The CDR authentication hides which user identifiers exist ([CDS: One-Time Password Credential Requirements][cds-otp]; [no email for an unknown user identifier](consumer-authentication.md#dh-1223)), and the CDS gives no rule for the site.
- Reason: With any password, each sign-in with the user identifier only shows which email addresses and customer numbers exist, so a refusal that hides them protects nothing. The site is lab-only and its data is invented. A text in plain language helps the consumer more.

### <a id="dh-2106"></a>DH-2106 Sessions in the memory of the process

- Kind: choice
- Basis: Step 3.2. A session on the server ends at a sign-out with no revocation list ([OWASP ASVS 5.0, 7.4.1][asvs-v7]), and [RFC 10017, section 7.1][rfc10017-cookie-session] permits a cookie session for an application on one domain with no OAuth.
- Reason: The lab runs one instance of each brand, so a store in memory is the simplest, with the limit of the authorisation sessions. A full store refuses a new session, so that many sign-ins cannot end the sessions of other consumers. The [production gaps][production-gaps] give the shared store that production needs.

### <a id="dh-2107"></a>DH-2107 No inactivity timeout

- Kind: gap
- Basis: Step 3.2. [OWASP ASVS 5.0, 7.3.1][asvs-v7] requires an inactivity timeout. A time limit that is longer than 20 hours needs no warning or extension ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]).
- Reason: The lab has no value to protect, and a short timeout only stops a demonstration. The [production gaps][production-gaps] record the sign-in of the site.

### <a id="dh-2108"></a>DH-2108 The name of the session cookie

- Kind: choice
- Basis: Step 3.2. Cookies are not isolated by port ([RFC 6265, section 8.5][rfc6265-8-5]), so in the lab the browser sends the cookie of each site to the consumer pages, the consoles and the other sites. The `__Host-` prefix is the recommendation for a session ID ([OWASP Session Management Cheat Sheet][owasp-session]).
- Reason: A name that the consumer pages do not use keeps the two apart. The brand ID in the name stops a sign-in at one brand from replacing the session of a different brand in the same browser. With no expiry, the server gives the lifetime.

### <a id="dh-2109"></a>DH-2109 No `Clear-Site-Data` at a sign-out

- Kind: choice
- Basis: Step 3.2. `Clear-Site-Data: "cookies"` affects the entire registered domain ([MDN: Clear-Site-Data][mdn-clear-site-data]).
- Reason: In the lab, all UIs are on `localhost`. The header then also removes the cookie of a CDR authentication flow in progress.

### <a id="dh-2110"></a>DH-2110 CSRF protection by the origin check

- Kind: choice
- Basis: Step 3.2. Fetch Metadata is the primary protection against CSRF for modern browsers, with the check of `Origin` as the fallback ([OWASP CSRF Prevention Cheat Sheet][owasp-csrf]). [Go: CrossOriginProtection][go-crossoriginprotection] does both, and the design system supports modern browsers only ([the supported browsers](../../../web/packages/ui/design/layout.md#ui-0096)).
- Reason: The check also applies to the sign-in, so login CSRF needs no session before the sign-in. A different lab port is a different origin, so the check also refuses a change from a different lab UI.

### <a id="dh-2111"></a>DH-2111 The separation of the sign-in from the CDR authentication

- Kind: choice
- Basis: Step 3.1. The friction standard forbids a reference to a different service in the authentication flow, and the CDS gives the phrasing of the password statement to the data holder ([CDS: Common Authentication Standards][cds-cx-authentication]: Authentication: Friction, Authentication: Passwords).
- Reason: A consumer must not learn to give a password when a data recipient starts the CDR authentication, so the note is on the site and the consumer pages do not change. The data recipient starts the authorisation flow, so the site has no link to it. The CDR authentication always asks for the user identifier and the one-time password, and it never uses a session of the site.

### <a id="dh-2112"></a>DH-2112 The session of a customer that the data holder does not have

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the online web site of a data holder.
- Reason: The account routes already refuse a session whose customer the data holder does not have. One rule for each route is the simplest.

### <a id="dh-2113"></a>DH-2113 A listener of its own for the site

- Kind: choice
- Basis: Step 3.3. The site is lab-only, and the CDS gives the rules of the public listener ([CDS: CORS][cds-cors]).
- Reason: On the public listener, the site needs exceptions to the CORS rule and to the not-found rule of the CDR APIs. A lab-only function then goes into the listener of the CDR endpoints. The port of the site is in [the port block of an instance](deployment.md#dh-1310).

### <a id="dh-2114"></a>DH-2114 The names of the site settings

- Kind: choice
- Basis: Step 3.3. The [web UI library](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0002) gives a setting that turns a web UI on and an allowlist setting, and does not name them.
- Reason: The names follow `CONSOLE_ENABLED` and `CONSOLE_ALLOWED_HOSTS`. `BANK_SITE` says what the listener serves, and it is different from the CDR names "public" and "secure".

### <a id="dh-2115"></a>DH-2115 TLS on the site

- Kind: choice
- Basis: Step 3.2. The `__Host-` prefix of the session cookie requires the `Secure` attribute ([RFC 6265bis, section 4.1.3.2][rfc6265bis-host]).
- Reason: A real online bank uses HTTPS. The public certificate of each instance already has the host name `localhost`, so the PKI does not change. One TLS configuration for all TLS listeners is the simplest.

### <a id="dh-2116"></a>DH-2116 No HSTS on the site

- Kind: choice
- Basis: Step 3.2. A browser applies an HSTS policy to each port of the host ([RFC 6797, Appendix B][rfc6797-b]).
- Reason: In the lab, all listeners are on `localhost`, and the consoles and the operator interfaces use plain HTTP. An HSTS policy from the site then stops them in the browser.

### <a id="dh-2117"></a>DH-2117 The pages after a sign-in and a sign-out

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the online web site of a data holder.
- Reason: A consumer signs in to see the accounts, so the accounts page is the first page. After a sign-out, the sign-in page lets the consumer sign in again at once.

[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0021]: ../../../docs/adr/0021-record-one-business-fact-in-each-event-type.md
[asvs-v7]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x16-V7-Session-Management.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-cors]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cors
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[cds-otp]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#one-time-password-credential-requirements
[cds-user-identifiers]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#user-identifiers
[go-crossoriginprotection]: https://pkg.go.dev/net/http#CrossOriginProtection
[mdn-clear-site-data]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Clear-Site-Data
[owasp-csrf]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
[owasp-session]: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
[production-gaps]: ../../../docs/production-gaps.md
[rfc10017-cookie-session]: https://www.rfc-editor.org/rfc/rfc10017#section-7.1
[rfc6265-8-5]: https://www.rfc-editor.org/rfc/rfc6265#section-8.5
[rfc6265bis-host]: https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-rfc6265bis#section-4.1.3.2
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[rfc6797-b]: https://www.rfc-editor.org/rfc/rfc6797#appendix-B
[wcag-1-3-5]: https://www.w3.org/TR/WCAG22/#identify-input-purpose
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-8]: https://www.w3.org/TR/WCAG22/#accessible-authentication-minimum
