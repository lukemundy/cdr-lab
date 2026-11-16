# Mock data holder: Authorisation

This area gives the authorisation that a consumer gives on the authorisation page, the authorisation response, the CDR arrangement and the presentation of the consumer pages.

## Authorisation page

After the authentication, one page asks the consumer for the authorisation to disclose CDR data. The [authorisation server library](../../../pkg/authserver/design/authorisation-request.md) gives the authorisation request that starts the flow.

- <a id="dh-0800"></a>**DH-0800** After the authentication of the consumer, the data holder shows one authorisation page that asks for the authorisation to disclose CDR data. ([CDR Rules][cdr-rules], rule 4.5 and rule 4.22)
- <a id="dh-0801"></a>**DH-0801** The authorisation page shows the name of the accredited person. The name is the `legal_entity_name` of the registration. When the registration has no `legal_entity_name`, the name is the `org_name`. ([CDR Rules][cdr-rules], rule 4.23(1)(a); [CX Guidelines: Authorisation to disclose][cx-authorise]; [Judgement: the names of the data recipient on the authorisation page](#dh-0850))
- <a id="dh-0802"></a>**DH-0802** The authorisation page also shows the brand name of the data recipient (`org_name`) and the name of the software product (`client_name`) from the registration. ([CDS: SSA Definition][cds-ssa-definition]; [Judgement: the names of the data recipient on the authorisation page](#dh-0850))
- <a id="dh-0803"></a>**DH-0803** For each scope of the request in the table below, the authorisation page shows the data cluster language and the permission language of the CDS, word for word. ([CDS: Data Language Standards][cds-data-language]; [CDS: Banking Language][cds-banking-language]; [CDS: Customer Language][cds-customer-language])
- <a id="dh-0804"></a>**DH-0804** When the request has the scope `openid`, the authorisation page shows no data cluster for it. ([CDS: Data Language Standards][cds-data-language])

| ID | Scope of the request | Data cluster | Permissions | Source |
|---|---|---|---|---|
| <a id="dh-0805"></a>**DH-0805** | `common:customer.basic:read` | Name and occupation | Name; Occupation | [CDS: Customer Language][cds-customer-language] |
| <a id="dh-0806"></a>**DH-0806** | `bank:accounts.basic:read` | Account name, type and balance | Name of account; Type of account; Account balance | [CDS: Banking Language][cds-banking-language] |
| <a id="dh-0807"></a>**DH-0807** | `bank:accounts.detail:read`, with `bank:accounts.basic:read` | Account numbers and features | Account number; Interest rates; Fees; Discounts; Account terms; Account mail address | [CDS: Banking Language][cds-banking-language] |
| <a id="dh-0808"></a>**DH-0808** | `bank:accounts.detail:read`, without `bank:accounts.basic:read` | Account balance and details | Name of account; Type of account; Account balance; Account number; Interest rates; Fees; Discounts; Account terms; Account mail address | [CDS: Banking Language][cds-banking-language] |
| <a id="dh-0809"></a>**DH-0809** | `bank:transactions:read` | Transaction details | Incoming and outgoing transactions; Amounts; Dates; Descriptions of transactions; Who you have sent money to and received money from | [CDS: Banking Language][cds-banking-language] |
| <a id="dh-0810"></a>**DH-0810** | `profile` | Name | Full name and title(s) | [CDS: Customer Language][cds-customer-language] |

- <a id="dh-0811"></a>**DH-0811** When the `sharing_duration` of the request is absent, 0, or 24 hours or less, the authorisation page says that the disclosure is on one occasion. ([CDR Rules][cdr-rules], rule 4.23(1)(d); [CDS: Requesting Sharing Duration][cds-sharing-duration]; [Judgement: the end of a disclosure on one occasion](end-of-an-arrangement.md#dh-0950))
- <a id="dh-0812"></a>**DH-0812** When the `sharing_duration` is more than 24 hours, the authorisation page says that the disclosure is for a period. It shows the duration and the date on which the sharing ends. ([CDR Rules][cdr-rules], rule 4.23(1)(e); [CX Guidelines: Authorisation to disclose][cx-authorise])
- <a id="dh-0813"></a>**DH-0813** The authorisation page shows the duration in days and hours. For the longest duration, it shows one year. For example, a duration of 90,000 seconds is "1 day and 1 hour". ([Judgement: the time zone and the duration on the authorisation page](#dh-0851))
- <a id="dh-0814"></a>**DH-0814** The authorisation page shows the date on which the sharing ends in UTC, with the label "UTC". ([Judgement: the time zone and the duration on the authorisation page](#dh-0851))
- <a id="dh-0815"></a>**DH-0815** The authorisation page says that the data includes the data from before the authorisation that the data holder holds for the selected accounts. ([CDR Rules][cdr-rules], rule 4.23(1)(b); [Judgement: the period of the data on the authorisation page](#dh-0852))
- <a id="dh-0816"></a>**DH-0816** The authorisation page says that the consumer can withdraw the authorisation at any time, and how. The text is "You can stop the sharing at any time. Sign in to <brand name> online banking and go to Data sharing, or ask <brand name>". ([CDR Rules][cdr-rules], rule 4.23(1)(f) and (g), rule 4.25(1); [Judgement: the instructions to withdraw](#dh-0853))
- <a id="dh-2603"></a>**DH-2603** The authorisation page shows each data cluster as a heading and a list of its permissions. The permissions are always visible: the consumer does not open a control to read them. ([the list of groups](../../../web/packages/ui/design/consumer-pages.md#ui-0383); [Judgement: data clusters that are always visible](#dh-2611))
- <a id="dh-0817"></a>**DH-0817** The instructions to withdraw on the authorisation page have no link. ([CDR Rules][cdr-rules], rule 4.24(d); [Judgement: the instructions to withdraw](#dh-0853))
- <a id="dh-0818"></a>**DH-0818** The authorisation page asks for no information other than the account selection. It offers no other products or services, and it has no reference to other documents. ([CDR Rules][cdr-rules], rule 4.24)

## Account selection

- <a id="dh-0819"></a>**DH-0819** When the request has a scope with the prefix `bank:`, the authorisation page lists each open account of the customer, with its display name and its masked number. For example: "Everyday xxx-xxx xxxxx1234". ([CDS: Authorisation Standards][cds-authorisation-standards]; [Judgement: closed accounts are not on the authorisation page](#dh-0854))
- <a id="dh-0820"></a>**DH-0820** When the authorisation page lists accounts, no account is selected before the consumer selects it. ([CDS: Authorisation Standards][cds-authorisation-standards])
- <a id="dh-0821"></a>**DH-0821** When the authorisation page lists accounts, the consumer selects one or more accounts. The consumer cannot confirm with no account selected. ([CDS: Authorisation Standards][cds-authorisation-standards])
- <a id="dh-0822"></a>**DH-0822** When the request has no scope with the prefix `bank:`, the authorisation page has no account selection. ([CDS: Authorisation Standards][cds-authorisation-standards])
- <a id="dh-0823"></a>**DH-0823** When the customer has no open account, the authorisation page says that the customer has no open account at the brand and cannot share data. The consumer can only cancel. This rule applies to each request, also a request with no scope with the prefix `bank:`. ([CDR Rules][cdr-rules], rule 4.5(1)(c); [CX Guidelines: Authorisation to disclose][cx-authorise]; [Judgement: the authorisation page of a customer with no open account](#dh-0855))

## Confirmation and cancel

- <a id="dh-0824"></a>**DH-0824** The authorisation page has one control to confirm and one control to cancel. ([CDS: Authorisation Standards][cds-authorisation-standards])
- <a id="dh-0825"></a>**DH-0825** When the consumer cancels, the flow ends with the error `access_denied` in the [authorisation response](../../../pkg/authserver/design/authorisation-request.md#auth-0104). ([CDR Support: Error scenarios and responses][support-error-scenarios])
- <a id="dh-0826"></a>**DH-0826** When a confirmation has one of these conditions, the data holder creates no arrangement. It shows the authorisation page again with the status 400, and the flow continues. ([CDS: Authorisation Standards][cds-authorisation-standards]; [Judgement: a confirmation that breaks an account rule](#dh-0856))
  - no account, for a request with a scope with the prefix `bank:`
  - an account that the customer does not have
  - an account that is closed
  - a customer with no open account
- <a id="dh-0827"></a>**DH-0827** When the consumer confirms, the data holder creates a [CDR arrangement](#dh-0832). The flow then ends with an authorisation code in the [authorisation response](../../../pkg/authserver/design/authorisation-request.md#auth-0107). ([CDS: Authorization Code Flow][cds-code-flow])

## Authorisation response

The authorisation server library gives the authorisation response: [the redirect with a signed JWT in the `response` parameter](../../../pkg/authserver/design/authorisation-request.md#auth-0104), [the `code`, or the `error` and `error_description`](../../../pkg/authserver/design/authorisation-request.md#auth-0107), [no encryption](../../../pkg/authserver/design/authorisation-request.md#auth-0108) and [a code that is valid for 60 seconds and for one use](../../../pkg/authserver/design/authorisation-request.md#auth-0109).

## CDR arrangement

- <a id="dh-0828"></a>**DH-0828** An arrangement records these values. ([CDS: CDR Arrangement ID][cds-arrangement-id]; [CDR Rules][cdr-rules], rule 9.3(1))
  - the customer
  - the client
  - the scopes of the request
  - the account identifier of each selected account
  - the time of the confirmation
  - the sharing duration
- <a id="dh-0829"></a>**DH-0829** An arrangement records the names of the data recipient that the authorisation page showed. These are the `client_name`, the `org_name` and, when the registration has one, the `legal_entity_name` of the registration at the confirmation. ([CDR Rules][cdr-rules], rule 9.3(1)(a); [Judgement: the names of the data recipient in an arrangement](#dh-0857))
- <a id="dh-0830"></a>**DH-0830** When the registration changes after the confirmation, the names of the data recipient in the arrangement do not change. For example, a `PUT` with a new `client_name` does not change the name on the consumer dashboard. ([Judgement: the names of the data recipient in an arrangement](#dh-0857))
- <a id="dh-0831"></a>**DH-0831** The `cdr_arrangement_id` is a random RFC 4122 UUID. It does not identify the consumer. ([CDS: CDR Arrangement ID][cds-arrangement-id])
- <a id="dh-0832"></a>**DH-0832** Each confirmation creates a new arrangement. A consumer can have more than one arrangement with one software product. ([CDR Support: Concurrent consent and arrangement ID requirements][support-concurrent-consent])
- <a id="dh-0833"></a>**DH-0833** An arrangement starts at the confirmation. When the software product does not exchange the authorisation code, the arrangement stays, with no token. ([CDR Support: Failure to obtain token within Authorisation Code expiry period][support-code-expiry]; [Judgement: the start of an arrangement at the confirmation](#dh-0858))
- <a id="dh-0834"></a>**DH-0834** The sharing of an arrangement ends at the time of the confirmation plus the sharing duration. An arrangement can [end before that time](end-of-an-arrangement.md#dh-0900). ([CDS: Requesting Sharing Duration][cds-sharing-duration])

## Presentation of the consumer pages

These rules apply to each page of the authentication and of the authorisation. The [consumer authentication](consumer-authentication.md) gives the content of the authentication pages.

- <a id="dh-0835"></a>**DH-0835** Each consumer page shows the logo and the name of the brand, and uses the colours of the brand. The text alternative of the logo is the brand name. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]; [Judgement: the brand on the consumer pages](#dh-0859))
- <a id="dh-0836"></a>**DH-0836** Each consumer page names the data holder with its brand name. For example, a page that refuses a request says "<brand name> cannot continue" or "<brand name> cannot use this form". ([Judgement: the brand on the consumer pages](#dh-0859))
- <a id="dh-2600"></a>**DH-2600** Each page of the flow has a [step indicator](../../../web/packages/ui/design/consumer-pages.md#ui-0370) with 3 steps. The first authentication page is "Step 1 of 3", the OTP page is "Step 2 of 3" and the authorisation page is "Step 3 of 3". ([Judgement: three steps in the flow](#dh-2610))
- <a id="dh-2601"></a>**DH-2601** A page that refuses a request or a form has no step indicator. ([Judgement: three steps in the flow](#dh-2610))
- <a id="dh-2602"></a>**DH-2602** The [hero sentence](../../../web/packages/ui/design/consumer-pages.md#ui-0366) of each page of the flow names the data recipient with its brand name, and names the brand of the data holder. For example, the hero sentence "Acme Budget asks for your data from Wattle Bank" names both. ([the brand name of the data recipient](consumer-authentication.md#dh-1215); [Judgement: the brand on the consumer pages](#dh-0859))
- <a id="dh-0837"></a>**DH-0837** Each consumer page has a footer with the text "<brand name> is not a real bank and this is just a demonstration. CDR Lab has no affiliation with any real bank nor the ACCC, DSB, or any real CDR participant." The footer has no link. ([CDS: Common Authentication Standards][cds-cx-authentication]; [Judgement: a footer that says the pages are a demonstration](#dh-0860))
- <a id="dh-0838"></a>**DH-0838** The data holder gives the text of the footer. A brand directory cannot change or remove it. ([Judgement: a footer that says the pages are a demonstration](#dh-0860))
- <a id="dh-0839"></a>**DH-0839** Accessibility: each consumer page meets WCAG 2.2 Level AA, and the automated accessibility checks pass. ([CDS: Accessibility Standards][cds-accessibility]; [Judgement: one design system for all UIs (ADR 0016)][adr-0016])
- <a id="dh-0840"></a>**DH-0840** Accessibility: each consumer page is usable at a viewport width of 320 CSS px, with no horizontal scroll. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10])
- <a id="dh-0841"></a>**DH-0841** The consumer pages use the [design guide of the consumer pages](../../../web/packages/ui/design/consumer-pages.md) of the design system. Their styles are a stylesheet of the design system that the data holder contains, and the styles of the brand. ([Judgement: one design system for all UIs (ADR 0016)][adr-0016])
- <a id="dh-0842"></a>**DH-0842** Each consumer page works with no JavaScript, and runs no script. ([Judgement: consumer pages with no script and no content from other hosts](#dh-0861))
- <a id="dh-0843"></a>**DH-0843** Each consumer page loads styles, fonts and images only from the same origin. It loads no content from other hosts. ([Judgement: consumer pages with no script and no content from other hosts](#dh-0861))
- <a id="dh-0844"></a>**DH-0844** Each consumer page has the Content Security Policy `default-src 'none'; style-src 'self'; font-src 'self'; img-src 'self'; frame-ancestors 'none'; base-uri 'none'`, with no `form-action`. ([MDN: CSP][mdn-csp]; [Judgement: the Content Security Policy of the consumer pages](#dh-0862))
- <a id="dh-0845"></a>**DH-0845** The data holder protects each form of the consumer pages against cross-site request forgery. A form that does not come from a page of the flow, in the browser that started the flow, gives 403, and the flow continues. ([Judgement: consumer pages with no script and no content from other hosts](#dh-0861))
- <a id="dh-0846"></a>**DH-0846** The data holder serves the stylesheet at `/static/consumer.css` and the fonts at `/static/fonts/` on the public listener, with no authentication. ([Judgement: consumer pages with no script and no content from other hosts](#dh-0861); [the cache rule of a fixed file](../../../pkg/platform/asset/DESIGN.md#asset-0001))
- <a id="dh-0847"></a>**DH-0847** The data holder serves the logo, the mark and the styles of the brand on the public listener, with no authentication. They contain no CDR data. A browser [checks each file again before it uses it](../../../pkg/platform/asset/DESIGN.md#asset-0001). ([Judgement: consumer pages with no script and no content from other hosts](#dh-0861))
- <a id="dh-2090"></a>**DH-2090** The public listener serves the styles of the brand at `/brand/styles.css`, the logo at `/brand/logo.svg` and the mark at `/brand/mark.svg`. ([Judgement: the paths of the brand files](#dh-2120))
- <a id="dh-2091"></a>**DH-2091** The stylesheet of the brand sets the custom properties of the design system under the selector `:root:root`. Thus the sequence of the stylesheets on a page has no effect. ([a brand is a set of values for the brand modifier](../../../web/packages/ui/design/tokens-and-brands.md#ui-0030); [Judgement: the paths of the brand files](#dh-2120))
- <a id="dh-0848"></a>**DH-0848** A response with an SVG file of the brand does not permit scripts, also when a browser opens its URL directly. ([MDN: CSP sandbox][mdn-csp-sandbox]; [Judgement: no script in an SVG file of the brand](#dh-0863))

## Judgements

### <a id="dh-0850"></a>DH-0850 The names of the data recipient on the authorisation page

- Kind: choice
- Basis: Step 3.3. Rule 4.23(1)(a) requires the name of the accredited person ([CDR Rules][cdr-rules]), and the CX Guidelines say that this is the legal entity name ([CX Guidelines: Authorisation to disclose][cx-authorise]). The SSA definition says that `client_name` and `org_name` are for the consumer "during authorization" ([CDS: SSA Definition][cds-ssa-definition]).
- Reason: All three names meet each source. The names come from the registration, because the data holder does not read participant data from the Register.

### <a id="dh-0851"></a>DH-0851 The time zone and the duration on the authorisation page

- Kind: choice
- Basis: Step 3.3. No source gives a time zone or a format for the page. All times in the data holder are UTC ([ADR 0012][adr-0012]).
- Reason: A time zone other than UTC is only for display, and a time zone for each brand is one more setting that the lab does not need. A duration in days and hours is exact, to the hour, for each value that a software product can ask for.

### <a id="dh-0852"></a>DH-0852 The period of the data on the authorisation page

- Kind: choice
- Basis: Step 3.3. Rule 4.23(1)(b) requires the period of time to which the CDR data of the request relates ([CDR Rules][cdr-rules]). An authorisation request has no such period: the software product gives a date range only in each data request.
- Reason: The statement is correct for each request that the authorisation permits.

### <a id="dh-0853"></a>DH-0853 The instructions to withdraw

- Kind: choice
- Basis: Step 3.1. Rule 4.23(1)(g) requires instructions, and rule 4.24(d) forbids a reference to other documents in the authorisation process ([CDR Rules][cdr-rules]). The CDS also forbids other documents and services in the flow ([CDS: Common Authentication Standards][cds-cx-authentication]).
- Reason: The text names the two methods of rule 4.25(1). A description of the consumer dashboard is the instruction that rule 4.23(1)(g) asks for, not a different document or service. The online bank site also has no link to the authorisation page ([the site has no link to the consumer pages](online-bank-site-sign-in.md#dh-2059)).

### <a id="dh-0854"></a>DH-0854 Closed accounts are not on the authorisation page

- Kind: gap
- Basis: Step 2. The ACCC says that a data holder that does not share a closed account should show these accounts in the account selection step, but they must not be selectable ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). The authorisation page shows only the open accounts.
- Reason: The ACCC text is guidance ("should"), and the page shows only the accounts that the consumer can select. A lab user can still test a closed account, because an arrangement keeps an account that closes after the confirmation.

### <a id="dh-0855"></a>DH-0855 The authorisation page of a customer with no open account

- Kind: choice
- Basis: Step 3.3. Rule 4.5(1)(c) asks for an authorisation only for a request on behalf of an eligible CDR consumer, and rule 4.26(1)(c) expires an authorisation when the consumer is not eligible ([CDR Rules][cdr-rules]). The CX Guidelines show unavailable accounts, not a consumer who is not eligible ([CX Guidelines: Authorisation to disclose][cx-authorise]).
- Reason: The page uses the same pattern as a request with a scope with the prefix `bank:` and no open account.

### <a id="dh-0856"></a>DH-0856 A confirmation that breaks an account rule

- Kind: choice
- Basis: Step 3.3. The CDS requires that the consumer selects accounts, and gives no response for a form that breaks the rule ([CDS: Authorisation Standards][cds-authorisation-standards]).
- Reason: A browser sends such a form only when the consumer selects nothing or when the form is changed, and the consumer can correct the first condition. An account of a different customer gets the same response, so the response does not show which identifiers exist.

### <a id="dh-0857"></a>DH-0857 The names of the data recipient in an arrangement

- Kind: choice
- Basis: Step 3.3. Rule 9.3 requires records of each authorisation, and rule 7.9(1)(c) requires that the dashboard identifies the data recipient ([CDR Rules][cdr-rules]). The ACCC recommends the brand name, the software product name and the legal entity name of the data recipient on the dashboard of a data holder ([CDR Support: Dashboards][support-dashboards]).
- Reason: The authorisation page showed these names to the consumer. A record that changes with the registration is not a record of the authorisation. The names then stay after a `DELETE` of the registration.

### <a id="dh-0858"></a>DH-0858 The start of an arrangement at the confirmation

- Kind: choice
- Basis: Step 2. When the software product does not exchange the code, what the data holder does "is at the discretion of the DH" ([CDR Support: Failure to obtain token within Authorisation Code expiry period][support-code-expiry]).
- Reason: The arrangement then has no token, and no party can use it. A second step that removes it adds nothing.

### <a id="dh-0859"></a>DH-0859 The brand on the consumer pages

- Kind: choice
- Basis: Step 3.1. The CDS and the CX Guidelines give no rule for the brand of the data holder on these pages ([CDS: Common Authentication Standards][cds-cx-authentication]).
- Reason: A consumer expects to see the brand that they selected. The consumer can then see which brand shares the data.

### <a id="dh-0860"></a>DH-0860 A footer that says the pages are a demonstration

- Kind: choice
- Basis: Step 3.1. The CDS forbids a reference to other documents and other services in the flow ([CDS: Common Authentication Standards][cds-cx-authentication]), so the footer has no link.
- Reason: A realistic copy of a bank page must not deceive a person. A user brand directory cannot remove the footer.

### <a id="dh-0861"></a>DH-0861 Consumer pages with no script and no content from other hosts

- Kind: choice
- Basis: Step 3.2. The CDS gives no rule for scripts or content from other hosts on these pages ([CDS: Accessibility Standards][cds-accessibility]). The pages stay Go pages with no script ([ADR 0014][adr-0014]).
- Reason: The conformance suite and assistive technology can then use the pages. The pages are an authentication surface, so they load nothing from a different origin and protect each form. The styles, the fonts and the brand files contain no CDR data, so they need no authentication.

### <a id="dh-0862"></a>DH-0862 The Content Security Policy of the consumer pages

- Kind: choice
- Basis: Step 3.2. The pages need only their stylesheets, fonts and the logo, from the same origin, and no script ([MDN: CSP][mdn-csp]).
- Reason: The policy permits only these. Some browsers apply `form-action` also to the redirect after a form ([MDN: form-action][mdn-csp-form-action]), and the last form of the flow redirects to the data recipient.

### <a id="dh-0863"></a>DH-0863 No script in an SVG file of the brand

- Kind: choice
- Basis: Step 3.2. A browser runs a script in an SVG file that it opens directly ([MDN: CSP sandbox][mdn-csp-sandbox]).
- Reason: A user brand directory can contain an SVG file with a script. The sandbox stops the script also outside the pages.

### <a id="dh-2120"></a>DH-2120 The paths of the brand files

- Kind: choice
- Basis: Step 3.3. The `logoUri` of the brand on the Register needs a fixed URL on the public base URI ([the logo on the public base URI](brand.md#dh-0313)), and the CDS gives no path for the other brand files.
- Reason: The files of the brand have one prefix, and the online bank site uses the same paths. The selector has a higher specificity than the `:root` of the tokens of the design system.

### <a id="dh-2610"></a>DH-2610 Three steps in the flow

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for a step indicator in the authorisation flow ([CDS: Common Authentication Standards][cds-cx-authentication]).
- Reason: The flow always has three pages: the user identifier, the OTP and the authorisation. The consumer sees how much of the flow is left. A refusal ends the flow, so it is not a step.

### <a id="dh-2611"></a>DH-2611 Data clusters that are always visible

- Kind: choice
- Basis: Step 3.3. The CDS requires the data cluster language and the permission language word for word, but it gives no rule for their display ([CDS: Data Language Standards][cds-data-language]).
- Reason: A permission that the consumer must open to read is easy to miss.

[adr-0012]: ../../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0016]: ../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cx-authorise]: https://cx.dsb.gov.au/authorise/authorisation-to-disclose
[cds-accessibility]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#accessibility-standards
[cds-arrangement-id]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-arrangement-id
[cds-authorisation-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorisation-standards
[cds-banking-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#banking-and-non-bank-lending-language
[cds-code-flow]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#authorization-code-flow
[cds-customer-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#customer-language-common
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[cds-data-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-language-standards-common
[cds-sharing-duration]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#requesting-sharing-duration
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
[mdn-csp]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy
[mdn-csp-form-action]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/form-action
[mdn-csp-sandbox]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/sandbox
[support-code-expiry]: https://cdr-support.zendesk.com/hc/en-us/articles/360004665496
[support-concurrent-consent]: https://cdr-support.zendesk.com/hc/en-us/articles/4407560834959
[support-closed-accounts]: https://cdr-support.zendesk.com/hc/en-us/articles/14519397304079
[support-dashboards]: https://cdr-support.zendesk.com/hc/en-us/articles/6140622825231
[support-error-scenarios]: https://cdr-support.zendesk.com/hc/en-us/articles/5729181240591
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-4-10]: https://www.w3.org/TR/WCAG22/#reflow
