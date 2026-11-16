# Mock data holder: Consumer dashboard

This area gives the consumer dashboard on the online bank site, where a consumer sees each authorisation and withdraws it. The dashboard is the online service of rule 1.15 of the CDR Rules.

## Dashboard page

- <a id="dh-1900"></a>**DH-1900** The consumer dashboard is the page "Data sharing" of the online bank site. On it, the consumer sees each authorisation that they gave to the brand, and withdraws an authorisation. ([CDR Rules][cdr-rules], rule 1.15(1))
- <a id="dh-1901"></a>**DH-1901** The dashboard needs a session. ([the sign-in page for a page that needs a session](online-bank-site-sign-in.md#dh-2027))
- <a id="dh-1902"></a>**DH-1902** The title of the dashboard is "Data sharing - <brand name> online banking". ([the title of each page of the site](online-bank-site.md#dh-1831))
- <a id="dh-1903"></a>**DH-1903** The dashboard shows each arrangement of the customer of the session. ([CDR Rules][cdr-rules], rule 1.15(3))
- <a id="dh-1904"></a>**DH-1904** The dashboard shows the current arrangements first. Then it shows the ended arrangements in a group of their own. ([CDR Rules][cdr-rules], rule 1.15(3)(d) and (e); [CDR Support: Dashboards][support-dashboards]; [Judgement: ended authorisations on the dashboard](#dh-1950))
- <a id="dh-1905"></a>**DH-1905** In each group, the arrangement with the newest confirmation is first. ([Judgement: ended authorisations on the dashboard](#dh-1950))
- <a id="dh-1906"></a>**DH-1906** The group of the ended arrangements has no time limit and no pages. ([Judgement: ended authorisations on the dashboard](#dh-1950))
- <a id="dh-1907"></a>**DH-1907** An arrangement is ended after an [event that ends it](end-of-an-arrangement.md#dh-0900), and at and after the [end of its sharing](authorisation.md#dh-0834), as in the operator console. For example, an arrangement with a sharing duration of 0 is in the ended group at once. ([CDR Rules][cdr-rules], rule 1.15(3)(e))
- <a id="dh-1908"></a>**DH-1908** When the customer has no arrangement, the dashboard says that the customer shares no data under the Consumer Data Right. ([CDR Rules][cdr-rules], rule 1.15(3))
- <a id="dh-1909"></a>**DH-1909** A customer with no open account sees the same dashboard. ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]; [Judgement: the dashboard of a customer with no open account](#dh-1951))

## Arrangements on the dashboard

- <a id="dh-1910"></a>**DH-1910** For each arrangement, the dashboard shows the data recipient: the name of the software product (`client_name`), the brand name (`org_name`) and, when the arrangement has one, the legal entity name. ([CDR Rules][cdr-rules], rule 1.15(3) and rule 7.9(1); [the names that an arrangement records](authorisation.md#dh-0829))
- <a id="dh-1911"></a>**DH-1911** For each arrangement, the dashboard shows each data cluster of its scopes, with its permissions, in the language of the authorisation page, word for word. ([CDS: Data Language Standards][cds-data-language]; [the data language on the authorisation page](authorisation.md#dh-0803))
- <a id="dh-1912"></a>**DH-1912** For each arrangement, the dashboard shows its accounts, with the display name and the masked number. ([CDR Rules][cdr-rules], rule 1.15(3))
- <a id="dh-1913"></a>**DH-1913** A closed account of an arrangement shows its closed status. ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]; [Judgement: a closed account stays in its arrangements](banking-apis.md#dh-1160))
- <a id="dh-1914"></a>**DH-1914** For each arrangement, the dashboard shows the date of the confirmation. ([CDR Rules][cdr-rules], rule 1.15(3))
- <a id="dh-1915"></a>**DH-1915** For each arrangement, the dashboard shows the period: one occasion, or the duration. The period is the same as on the [authorisation page](authorisation.md#dh-0811). ([CDR Rules][cdr-rules], rule 1.15(3))
- <a id="dh-1916"></a>**DH-1916** For a current arrangement, the dashboard shows the date on which the sharing ends. ([CDR Rules][cdr-rules], rule 1.15(3))
- <a id="dh-1917"></a>**DH-1917** For an ended arrangement, the dashboard shows the date on which it ended. This is the time of the event that ended it, or the end of its sharing. ([CDR Rules][cdr-rules], rule 1.15(3)(e); [Judgement: ended authorisations on the dashboard](#dh-1950))
- <a id="dh-1918"></a>**DH-1918** The dashboard does not show the reason or the method of the end of an arrangement. ([CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method]; [CDR Support: Dashboards][support-dashboards]; [Judgement: ended authorisations on the dashboard](#dh-1950))
- <a id="dh-1919"></a>**DH-1919** For each data cluster, the dashboard shows the date of the first and of the last disclosure, from the [record of disclosures](banking-apis.md#dh-1110). This applies also to an ended arrangement. ([CDR Support: Privacy Safeguard 5 - dates on the consumer dashboards][support-dashboard-dates]; [CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters])
- <a id="dh-1920"></a>**DH-1920** When a data cluster has no disclosure, the dashboard shows "No data shared" for it. ([CDR Support: Data Holder Dashboard data sharing display at data cluster level][support-dashboard-clusters])
- <a id="dh-1921"></a>**DH-1921** For each arrangement, the dashboard shows this message of the CDS about the handling of the data. The brand name of the data recipient is in place of each name of the data recipient. ([CDS: Dashboard Standards][cds-dashboard-standards]; [Judgement: the suggested text of the CDS messages](#dh-1952))
  - "You should check with <data recipient brand name> for more information on how they are handling your data, and for any other permissions you may have given them. See <data recipient brand name>'s CDR policy or their Dashboard for more information."
- <a id="dh-1922"></a>**DH-1922** The dashboard shows no accreditation number. ([CDR Support: Dashboards][support-dashboards]; [Judgement: no accreditation number on the dashboard](#dh-1953))
- <a id="dh-1923"></a>**DH-1923** The dashboard shows no CDR receipt and no amendment of an authorisation. ([the non-goals of the data holder](README.md#non-goals))

## Withdrawal

- <a id="dh-1924"></a>**DH-1924** Each current arrangement has the action "Stop sharing". An ended arrangement has no action. ([CDR Rules][cdr-rules], rule 1.15(1)(c); [CX Guidelines: Withdrawal][cx-withdrawal]; [the withdrawal on the dashboard](end-of-an-arrangement.md#dh-0912))
- <a id="dh-1925"></a>**DH-1925** Before a withdrawal, the site asks the consumer to confirm it in a dialog. ([CDR Rules][cdr-rules], rule 1.15(1)(e); [WCAG 2.2, SC 3.3.4][wcag-3-3-4]; [Judgement: a confirmation of a withdrawal](#dh-1954))
- <a id="dh-1926"></a>**DH-1926** The dialog names the software product and the brand name of the data recipient. ([Judgement: a confirmation of a withdrawal](#dh-1954))
- <a id="dh-1927"></a>**DH-1927** The dialog says that the sharing stops at once, that the withdrawal cannot be reversed, and that the brand tells the data recipient. ([the effect at the time of the request](end-of-an-arrangement.md#dh-0915); [Judgement: a confirmation of a withdrawal](#dh-1954))
- <a id="dh-1928"></a>**DH-1928** The dialog has the actions "Stop sharing" and "Cancel". ([CX Guidelines: Withdrawal][cx-withdrawal]; [Judgement: a confirmation of a withdrawal](#dh-1954))
- <a id="dh-1929"></a>**DH-1929** The dialog shows these two messages of the CDS. The brand name of the data recipient is in place of the name of the data recipient. ([CDS: Withdrawal Standards][cds-withdrawal-standards]; [CDR Rules][cdr-rules], rule 1.15(1)(d); [the withdrawal messages only on the dashboard](end-of-an-arrangement.md#dh-0914); [Judgement: the suggested text of the CDS messages](#dh-1952))
  - "You should check with <data recipient brand name> before you stop sharing to understand the consequences."
  - "CDR data is either deleted or de-identified when it is no longer required."
- <a id="dh-1930"></a>**DH-1930** When the consumer confirms, the data holder withdraws the arrangement for the customer of the session, with the method `CONSUMER_DASHBOARD`. ([the method of each withdrawal](end-of-an-arrangement.md#dh-0917); [the same effects as a withdrawal by the operator](end-of-an-arrangement.md#dh-0916); [Judgement: the withdrawal uses the use case of the operator](online-bank-site-api.md#dh-1798))
- <a id="dh-1931"></a>**DH-1931** After a withdrawal, the dialog closes, and a status message names the software product. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [CX Guidelines: Withdrawal][cx-withdrawal])
- <a id="dh-1932"></a>**DH-1932** After a withdrawal, the site shows the dashboard with the focus on its heading. The arrangement is in the group of the ended arrangements. ([CDR Rules][cdr-rules], rule 4.27; [WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="dh-1933"></a>**DH-1933** When the arrangement ended between the read and the confirmation, the dialog says that the sharing changed. The site then shows the current state. ([the response to a withdrawal of an ended arrangement](online-bank-site-api.md#dh-1770))
- <a id="dh-1934"></a>**DH-1934** When a withdrawal fails for a different reason, the dialog stays open and shows the error. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="dh-1935"></a>**DH-1935** The site never sends a withdrawal again by itself. ([RFC 9110, section 9.2.2][rfc9110-idempotent])
- <a id="dh-1936"></a>**DH-1936** After a withdrawal, the data holder sends no email and no CDR receipt to the consumer. ([Judgement: no email other than the OTP email](consumer-authentication.md#dh-1287))

## Telemetry

- <a id="dh-1940"></a>**DH-1940** Telemetry: no span, metric or log of the list of the arrangements on the dashboard, of a withdrawal on the site or of the record of a disclosure contains one of these values. ([no personal information in telemetry](non-functional-requirements.md#dh-0008); [Judgement: no product ID and no data recipient name in the telemetry of the site](online-bank-site.md#dh-1894))
  - the customer number
  - an arrangement ID
  - the name of the software product
  - the brand name or the legal entity name of a data recipient
  - a data cluster
  - a time of a disclosure

## Judgements

### <a id="dh-1950"></a>DH-1950 Ended authorisations on the dashboard

- Kind: choice
- Basis: Step 2. Rule 1.15(3)(e) requires the time of the end of an authorisation that is not current ([CDR Rules][cdr-rules]). The ACCC says that the dashboard need not show the method or the reason ([CDR Support: Consumer Dashboard - revoked and expired consents][support-withdrawal-method], [CDR Support: Dashboards][support-dashboards]).
- Reason: The accounts page of the site shows closed accounts in the same way. A customer has few arrangements, so the group needs no time limit and no pages.

### <a id="dh-1951"></a>DH-1951 The dashboard of a customer with no open account

- Kind: choice
- Basis: Step 2. The ACCC says that the obligation to give a dashboard ends with the eligibility of the consumer ([CDR Support: Closed accounts in the banking and non-bank lenders sectors][support-closed-accounts]). No rule forbids a dashboard after that time.
- Reason: The same dashboard is the least complex solution. The consumer sees the authorisations that ended at the close.

### <a id="dh-1952"></a>DH-1952 The suggested text of the CDS messages

- Kind: choice
- Basis: Step 1. The CDS requires the messages, and permits the data holder to use or paraphrase the suggested text ([CDS: Dashboard Standards][cds-dashboard-standards], [CDS: Withdrawal Standards][cds-withdrawal-standards]).
- Reason: The suggested text needs no review, and a test can check it. The brand name is the name of the data recipient that the consumer knows best.

### <a id="dh-1953"></a>DH-1953 No accreditation number on the dashboard

- Kind: choice
- Basis: Step 2. The ACCC only recommends the accreditation number on the dashboard ([CDR Support: Dashboards][support-dashboards]).
- Reason: The data holder does not read participant data from the Register ([Judgement: the names of the data recipient on the authorisation page](authorisation.md#dh-0850)). The sources give no data standard and no Register entry for rule 1.15(1)(bb) or rule 7.9(1)(c).

### <a id="dh-1954"></a>DH-1954 A confirmation of a withdrawal

- Kind: choice
- Basis: Step 3.2. A withdrawal cannot be reversed ([WCAG 2.2, SC 3.3.4][wcag-3-3-4]), and the CDS messages are "part of the withdrawal process" ([CDS: Withdrawal Standards][cds-withdrawal-standards]). The CX Guidelines suggest positive friction and the words "Stop sharing" ([CX Guidelines: Withdrawal][cx-withdrawal]).
- Reason: Two steps prevent a withdrawal by error. The withdrawal stays less complicated than the authorisation, which has a customer number, an OTP and a confirmation ([CDR Rules][cdr-rules], rule 1.15(1)(e)).

[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-data-language]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-language-standards-common
[cds-dashboard-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_dashboard-standards_data-holder-dashboards
[cds-withdrawal-standards]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#withdrawal-standards
[cx-withdrawal]: https://cx.dsb.gov.au/consent-management/consent-management-data-holder/withdrawal/cx-guidelines
[support-closed-accounts]: https://cdr-support.zendesk.com/hc/en-us/articles/14519397304079
[support-dashboards]: https://cdr-support.zendesk.com/hc/en-us/articles/6140622825231
[support-withdrawal-method]: https://cdr-support.zendesk.com/hc/en-us/articles/4411292647823
[support-dashboard-dates]: https://cdr-support.zendesk.com/hc/en-us/articles/5333587630095
[support-dashboard-clusters]: https://cdr-support.zendesk.com/hc/en-us/articles/4920717853327
[rfc9110-idempotent]: https://www.rfc-editor.org/rfc/rfc9110#section-9.2.2
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
