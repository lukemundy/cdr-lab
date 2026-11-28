# Mock data recipient: Web app

This area gives the web app of the brand: its listener, its pages and their presentation, its statements and its policies. The other areas of the web app are [sign-up](sign-up.md), [sign-in](sign-in.md), [consent](consent.md), [authorisation](authorisation.md) and [accounts and transactions](accounts-and-transactions.md).

## Scope of the web app

- <a id="dr-0200"></a>**DR-0200** The data recipient embeds and serves the web app as a React single-page application. The service is the backend for frontend of the web app. ([Judgement: web UIs as single-page applications that the service embeds (ADR 0014)][adr-0014])
- <a id="dr-0201"></a>**DR-0201** A hand-written OpenAPI document gives the contract of the API of the web app. ([Judgement: API contracts in OpenAPI (ADR 0017)][adr-0017])
- <a id="dr-0202"></a>**DR-0202** The browser gets no token and no credential other than the session cookie. ([RFC 10017, section 7.1][rfc10017-cookie-session]; [the session cookie](sign-in.md#dr-0430))

## Listener

- <a id="dr-0203"></a>**DR-0203** The web app and its API are on the web app listener, with TLS. The listener serves no other group of routes. ([Judgement: a web app listener on a data recipient (ADR 0025)][adr-0025])
- <a id="dr-0204"></a>**DR-0204** The server certificate of the web app listener is from the public CA of the deployment. ([Judgement: a web app listener on a data recipient (ADR 0025)][adr-0025])
- <a id="dr-0205"></a>**DR-0205** The web app is always on. It has no setting that turns it off. ([Judgement: the web app is the product of the data recipient](#dr-0250))
- <a id="dr-0206"></a>**DR-0206** The setting `WEB_APP_ALLOWED_HOSTS` gives the `Host` allowlist of the web app. It is required, and it has no default. ([a Host allowlist with no default](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0003))
- <a id="dr-0207"></a>**DR-0207** The setting `WEB_APP_BASE_URI` gives the base URI of the web app, for example `https://localhost:8585`. It is required, and it has no default. The data recipient makes the links of each email from it, not from a request. ([OWASP Forgot Password Cheat Sheet][owasp-forgot-password])
- <a id="dr-0208"></a>**DR-0208** The web app obeys the rules of the web UI library for these items, with the exception of [the setting that turns a web UI on](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0002). ([Judgement: the web app is the product of the data recipient](#dr-0250))
  - [the `Host` allowlist](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0004)
  - [the refusal of cross-origin changes](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0005) and [no CORS headers](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0006)
  - [the security headers](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007) and [the cache headers](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0008)
  - [the API under `/api/`](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0011), [404 for an unknown operation](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0012) and [the HTML page for each other path](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0013)
- <a id="dr-0209"></a>**DR-0209** Each error response of the API has a problem details body. A refused value of a form gives a field error with a plain-language `detail`. ([the field errors of a problem](../../../pkg/platform/problem/DESIGN.md#prob-0003); [the detail of a field error](../../../pkg/platform/problem/DESIGN.md#prob-0005); [plain language in a consumer UI](../../../web/packages/ui/design/content.md#ui-0142))
- <a id="dr-0210"></a>**DR-0210** The API refuses a JSON request body with an unknown field, and a body with a second JSON value. ([no unknown field](../../../pkg/platform/jsonhttp/DESIGN.md#body-0001); [one JSON value](../../../pkg/platform/jsonhttp/DESIGN.md#body-0002))
- <a id="dr-0211"></a>**DR-0211** The API refuses a request body that is larger than 16 KiB with 413. ([the problem of a body above the limit](../../../pkg/platform/problem/DESIGN.md#prob-0012); [Judgement: a request body of 16 KiB at most](#dr-0251))

## Pages

- <a id="dr-0212"></a>**DR-0212** The web app has these pages that need no session. ([Judgement: a landing page that says what the app does](#dr-0252))
  - the landing page
  - the sign-up page, and the page that asks the consumer to check their email
  - the page of a verification link
  - the sign-in page
  - the password reset request page, and the page of a password reset link
  - the privacy policy and the CDR policy
- <a id="dr-0213"></a>**DR-0213** The web app has these pages that need a session. ([each page with member data needs a session](sign-in.md#dr-0412))
  - the member details page
  - the home page
  - the bank selection page, the consent page and the redirection notice
  - the page of the redirect URI, and the page of each outcome of an authorisation flow
  - the page of an account
- <a id="dr-0214"></a>**DR-0214** The landing page says in one sentence that the app shows the bank accounts of the consumer in one place. It has the photo of the brand, and a primary action to sign up and a link to sign in. ([Judgement: a landing page that says what the app does](#dr-0252))
- <a id="dr-0229"></a>**DR-0229** The sign-up page, the sign-in page and the password reset pages show the access photo of the brand, when the brand has one. On a page 1280 CSS pixels wide, the photo is next to the form. On a narrow page, a crop of the photo is above the form. ([the access photo of a brand](brand.md#dr-0127); [Judgement: a portrait photo of 4:5 next to a form](brand.md#dr-0154))
- <a id="dr-0215"></a>**DR-0215** The home page shows [the connections of the member](accounts-and-transactions.md#dr-1101), or says that [the member has no connected bank](accounts-and-transactions.md#dr-1100). ([Judgement: a landing page that says what the app does](#dr-0252))

## Presentation

- <a id="dr-0216"></a>**DR-0216** The web app is mobile-first. Each page has its layout for a viewport 360 CSS pixels wide first, and it has a layout for a viewport 1280 CSS pixels wide. ([Judgement: a mobile-first web app](#dr-0253))
- <a id="dr-0217"></a>**DR-0217** Each page has no horizontal scroll at a viewport 320 CSS pixels wide. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10])
- <a id="dr-0218"></a>**DR-0218** Each page meets the accessibility target of a consumer UI. ([a consumer UI meets WCAG 2.2 Level AA](../../../web/packages/ui/design/accessibility.md#ui-0110))
- <a id="dr-0219"></a>**DR-0219** The web app uses the components and the tokens of the design system, with the colours and the fonts of the brand. ([Judgement: a design system on shadcn, Base UI and DTCG tokens (ADR 0016)][adr-0016])
- <a id="dr-0220"></a>**DR-0220** Each page has a header with the logo of the brand. The logo has the brand name as its text alternative. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="dr-0221"></a>**DR-0221** When the member is signed in, the header has a link to the home page, a link to the member details page and a control to sign out. ([Judgement: a landing page that says what the app does](#dr-0252))
- <a id="dr-0222"></a>**DR-0222** The icon of the web app in the browser is the mark of the brand. ([Judgement: a brand directory as the data holder has](brand.md#dr-0150))

## Statements and policies

- <a id="dr-0223"></a>**DR-0223** Each page has a footer. The footer has the password statement, the demonstration statement, and links to the privacy policy and to the CDR policy. ([CDR Rules][cdr-rules], rule 7.2(8); [Judgement: the password statement on each page and in each email](#dr-0254))
- <a id="dr-0224"></a>**DR-0224** The password statement says that third parties do not need the password of the consumer to get CDR data. For example: "Sorrel never needs your bank password. Never share your bank password with anyone". ([CDS: Common Authentication Standards][cds-cx-authentication]; [Judgement: the password statement on each page and in each email](#dr-0254))
- <a id="dr-0225"></a>**DR-0225** The demonstration statement says that the brand is a demonstration and not a real service, and that the consumer must give only invented details. ([Judgement: a demonstration with invented details only](#dr-0255))
- <a id="dr-0226"></a>**DR-0226** The privacy policy and the CDR policy are two different pages. The CDR policy is not part of the privacy policy. ([CDR Rules][cdr-rules], rule 7.2(2))
- <a id="dr-0227"></a>**DR-0227** The privacy policy says which personal information the brand collects, why it collects it, how a member gets access to it, and how a member makes a complaint. ([Privacy Act 1988][privacy-act], APP 1.4)
- <a id="dr-0228"></a>**DR-0228** Each policy says at its start that the brand is a demonstration and that the policy is invented text. ([Judgement: a demonstration with invented details only](#dr-0255))

## Telemetry

The web app obeys [route patterns in place of raw paths](../../../pkg/platform/telemetry/DESIGN.md#tele-0005) and [no query string in a span](../../../pkg/platform/telemetry/DESIGN.md#tele-0007) of the Telemetry library, and [the values that telemetry never contains](non-functional-requirements.md#dr-0002).

## Judgements

### <a id="dr-0250"></a>DR-0250 The web app is the product of the data recipient

- Kind: choice
- Basis: Step 3.3. The web UI library turns a web UI off by default, because a production service has no such UI ([the setting that turns a web UI on](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0002)).
- Reason: A data recipient has no function for a consumer without its app, so a setting that turns the app off has no use. The other rules of the library apply, because the risks of the browser are the same.

### <a id="dr-0251"></a>DR-0251 A request body of 16 KiB at most

- Kind: choice
- Basis: Step 3.2. The largest body of the API is the member details, with fields of 200 characters or less ([the member details](sign-up.md#dr-0320)).
- Reason: 16 KiB is more than each valid body, also with 4 bytes for each character. A smaller limit gives an attacker less memory for each request.

### <a id="dr-0252"></a>DR-0252 A landing page that says what the app does

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the pages of a data recipient before the consent.
- Reason: A consumer must know what the app does before they give personal information. The home page has no function before a bank is connected, so it says so.

### <a id="dr-0253"></a>DR-0253 A mobile-first web app

- Kind: choice
- Basis: Step 3.3. Most consumer finance apps are phone apps. The design system tests a consumer UI at a viewport 360 CSS pixels wide.
- Reason: The lab has no native app, so the web app is the phone app. A layout for a desktop viewport lets an engineer use the app beside the telemetry.

### <a id="dr-0254"></a>DR-0254 The password statement on each page and in each email

- Kind: choice
- Basis: Step 1. The CDS requires the statement "in consumer-facing interactions and communications" of data recipients ([CDS: Common Authentication Standards][cds-cx-authentication]). It does not say which interactions.
- Reason: A statement in the footer of each page and of each email is in each interaction with no list to keep. The data holder also has it in the footer of each page of its online bank site ([the password statement of the data holder](../../dataholder/design/online-bank-site.md#dh-1850)).

### <a id="dr-0255"></a>DR-0255 A demonstration with invented details only

- Kind: choice
- Basis: Step 3.1. The lab data has no value and is not protected as real personal information.
- Reason: A consumer who sees a realistic app can give real details. The statement on each page and at the start of each policy tells them not to.

[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0016]: ../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[adr-0017]: ../../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[owasp-forgot-password]: https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
[privacy-act]: https://www.legislation.gov.au/C2004A03712/latest/text
[rfc10017-cookie-session]: https://www.rfc-editor.org/rfc/rfc10017#section-7.1
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-4-10]: https://www.w3.org/TR/WCAG22/#reflow
