# Mock data holder: Online bank site

This area gives the lab-only online bank site of the brand: its pages and their presentation. The site has a landing page, the current products, sign-up, sign-in and sign-out, the accounts of the consumer, the opening and the close of an account, and the consumer dashboard. The other areas of the site are [sign-in](online-bank-site-sign-in.md), [accounts](online-bank-site-accounts.md), the [consumer dashboard](consumer-dashboard.md) and the [site API](online-bank-site-api.md).

## Scope of the site

- <a id="dh-1800"></a>**DH-1800** The online bank site is a lab-only feature. A real component does not depend on it, and it is not the web site of a production data holder. ([Judgement: web UIs as single-page applications that the service embeds (ADR 0014)][adr-0014]; [mock and real components][readme-component-types])
- <a id="dh-1801"></a>**DH-1801** The data holder embeds and serves the site as a React single-page application. ([Judgement: web UIs as single-page applications that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1802"></a>**DH-1802** A hand-written OpenAPI document gives the contract of the [site API](../DESIGN.md#site-api). ([Judgement: API contracts in OpenAPI (ADR 0017)][adr-0017])
- <a id="dh-1803"></a>**DH-1803** The site is not part of the authorisation flow. The authentication and authorisation pages stay on the public listener. ([the separation from the CDR authentication](../DESIGN.md#separation-from-the-cdr-authentication))
- <a id="dh-1804"></a>**DH-1804** The site does not change the authentication and authorisation pages. ([the separation from the CDR authentication](../DESIGN.md#separation-from-the-cdr-authentication); [the brand on the consumer pages](authorisation.md#dh-0836))

## Landing page

- <a id="dh-1805"></a>**DH-1805** The landing page has a heading that names the brand, one sentence about the brand, and a link to the product list. ([Judgement: one text of the landing page for each brand](#dh-1880))
- <a id="dh-1806"></a>**DH-1806** The data holder gives the text of the landing page, with the brand name in it. A brand directory does not change it. ([Judgement: one text of the landing page for each brand](#dh-1880))
- <a id="dh-1807"></a>**DH-1807** The landing page has one link for each product category that has a current product. The link has the name of the category and the number of its current products. ([CDS: Product Categories][cds-product-categories]; [Judgement: the product categories on the landing page](#dh-1881))
- <a id="dh-1808"></a>**DH-1808** Each link of a product category goes to the group of that category in the product list. ([Judgement: the product categories on the landing page](#dh-1881))
- <a id="dh-1809"></a>**DH-1809** The landing page has a short statement that the brand shares the data of a consumer under the Consumer Data Right only with the authorisation of the consumer. ([Judgement: a statement about CDR sharing on the landing page](#dh-1882))
- <a id="dh-1810"></a>**DH-1810** The text of the site makes no claim about rates, awards or customers. It copies no text of a real bank. ([Judgement: no claim and no copied text on the site](#dh-1883); [Judgement: a footer that says the pages are a demonstration](authorisation.md#dh-0860))

## Products

- <a id="dh-1811"></a>**DH-1811** The product list shows each current product of the brand. It shows no future product, and no product whose `effectiveTo` is in the past. ([CDS: Get Products][cds-get-products]; [a current product](product-reference-data.md#dh-0219); [Judgement: the products that the site shows](../DESIGN.md#decisions))
- <a id="dh-1812"></a>**DH-1812** The product list puts the products in groups by `productCategory`, in the order of the categories in the CDS. Only a category with a current product has a group. ([CDS: Product Categories][cds-product-categories])
- <a id="dh-1813"></a>**DH-1813** In a group, the products are in the order of their names. ([Judgement: the order of the products in a group](#dh-1884))
- <a id="dh-1814"></a>**DH-1814** For each product, the product list shows its name and its description, and links to the detail of the product. ([CDS: BankingProductV6][cds-schema-product]; [Judgement: the name and the description in the product list](#dh-1885))
- <a id="dh-1815"></a>**DH-1815** The detail of a product shows each of these fields that the data holder holds for the product. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
  - the name, the category and the description
  - `isTailored`
  - `effectiveFrom` and `effectiveTo`, when present
  - each deposit rate and each lending rate
  - each fee and each feature
  - each field of a rate, a fee or a feature that has a value
- <a id="dh-1816"></a>**DH-1816** When `isTailored` is true, the detail of the product says that the rates and the fees can be different for the consumer. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
- <a id="dh-1817"></a>**DH-1817** When a product is not current, or does not exist, its detail is the "not found" page of the site. ([Judgement: the products that the site shows](../DESIGN.md#decisions))
- <a id="dh-1818"></a>**DH-1818** The site shows the same products and the same values as Get Products and Get Product Detail of the brand at the same time. ([the products of the site API](../DESIGN.md#site-api); [Judgement: the source of the products of the site](../DESIGN.md#decisions))
- <a id="dh-1819"></a>**DH-1819** The site shows a label in plain language for each value of a CDS enumeration. The label gives the meaning that the CDS gives for the value. For example, the site shows a label for `TRANS_AND_SAVINGS_ACCOUNTS` and for `MONTHLY_ACCOUNT_FEE`. ([CDS: BankingProductDetailV7][cds-schema-product-detail]; [CDS: Product Categories][cds-product-categories]; [design system: Content][ui-content]; [Judgement: the labels of the CDS enumerations](../DESIGN.md#decisions))
- <a id="dh-1820"></a>**DH-1820** The site shows each RateString as a percentage. For example, `0.0125` shows as "1.25%". ([CDS: Common Field Types][cds-common-field-types]; [design system: Content][ui-content])
- <a id="dh-1821"></a>**DH-1821** The site shows each AmountString as an amount. ([CDS: Common Field Types][cds-common-field-types]; [design system: Content][ui-content])
- <a id="dh-1822"></a>**DH-1822** The site shows each ISO 8601 duration in words. For example, `P1M` shows as "each month" for an `applicationFrequency`, and `P55D` shows as "55 days". ([CDS: Common Field Types][cds-common-field-types]; [design system: Content][ui-content])
- <a id="dh-1823"></a>**DH-1823** The site shows an `additionalValue` in the format that the CDS gives for its type. When the type gives no format, the site shows the value as text. ([CDS: BankingProductDetailV7][cds-schema-product-detail])
- <a id="dh-1824"></a>**DH-1824** When the site does not know a path, it shows its "not found" page. ([Judgement: routes in the browser](#dh-1886))

## Presentation

- <a id="dh-1825"></a>**DH-1825** The site uses the tokens and the components of the design system, with the colours, the fonts and the radius of the brand. ([design system: Brands][ui-brands]; [Judgement: one design system (ADR 0016)][adr-0016])
- <a id="dh-1826"></a>**DH-1826** The site has a shell of its own, with a header, a main area and a footer. It does not use the shell of the operator console. ([design system: Brands][ui-brands])
- <a id="dh-1827"></a>**DH-1827** Headings use the heading font of the brand. The other text uses the body font. ([design system: Brands][ui-brands])
- <a id="dh-1828"></a>**DH-1828** Each page shows the logo of the brand on the brand colour, with the brand name as its text alternative. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="dh-1829"></a>**DH-1829** Each page has a link to the product list. ([design system: Brands][ui-brands])
- <a id="dh-1830"></a>**DH-1830** The pages of the site have no "Secure" label, no hero with the data recipient and no step indicator. ([Judgement: a site that does not look like the consumer pages](#dh-1890); [the separation from the CDR authentication](../DESIGN.md#separation-from-the-cdr-authentication))
- <a id="dh-1831"></a>**DH-1831** The title of each page is the name of the page, then "<brand name> online banking". For example, "Data sharing - <brand name> online banking". ([Judgement: a site that does not look like the consumer pages](#dh-1890))
- <a id="dh-1832"></a>**DH-1832** Until the site API gives the brand, and when it fails, the pages say "Bank" in place of the brand name. ([Judgement: "Bank" before the brand name is known](#dh-1893))
- <a id="dh-1833"></a>**DH-1833** The site shows dates and amounts with the rules of the design system: Australian English and the locale `en-AU`. ([design system: Content][ui-content])

## Layouts

Each brand has one layout. The table gives the shapes of each layout.

- <a id="dh-1834"></a>**DH-1834** Each brand has one of the four layouts in the table below. The `layout` of the brand directory selects it. ([the layout of a brand directory](brand.md#dh-0335); [Judgement: a layout for each brand](#dh-1887))
- <a id="dh-1835"></a>**DH-1835** A layout changes the arrangement and the shapes of the landing page, of the account cards and of the other pages. ([Judgement: a layout for each brand](#dh-1887))
- <a id="dh-1836"></a>**DH-1836** A layout does not change the text, the content, the order of the content in the product list or the accessibility of a page. ([Judgement: a layout for each brand](#dh-1887))

| ID | Layout | Hero photo | Product categories on the landing page | Account cards | Source |
|---|---|---|---|---|---|
| <a id="dh-1837"></a>**DH-1837** | `meadow` | 4:3, in a frame with round corners, with the text on a panel in the brand colour that overlaps it | Cards with a photo and a count | Cards in the brand colour | [Judgement: a layout for each brand](#dh-1887) |
| <a id="dh-1838"></a>**DH-1838** | `ledger` | 21:9, full width, with the text on a panel that overlaps its lower edge | A ruled list with a square photo, the name and the count | Rows in the hero colour, with the balances in columns | [Judgement: a layout for each brand](#dh-1887) |
| <a id="dh-1839"></a>**DH-1839** | `horizon` | 3:1, full width, above a band in the hero colour with the text | Cards with a 3:2 photo and a count | Cards in the hero colour, like a ticket | [Judgement: a layout for each brand](#dh-1887) |
| <a id="dh-1840"></a>**DH-1840** | `updraft` | 1:1, in a panel in the hero colour with the text, cut at an angle | A grid of tiles with a photo and a label on a solid panel. The first tile is large. It is `TRANS_AND_SAVINGS_ACCOUNTS` when the brand has a current product in it. The other tiles are in the order of the CDS. | Cards in the hero colour | [Judgement: a layout for each brand](#dh-1887) |

## Photos and colours

- <a id="dh-1841"></a>**DH-1841** The landing page shows the hero photo. Each link of a product category on the landing page shows the photo of its category. ([the photos of a brand directory](brand.md#dh-0328))
- <a id="dh-1842"></a>**DH-1842** The product list shows the photo of each category at its group. The detail of a product shows the photo of its category. ([the photo of a category](brand.md#dh-0329))
- <a id="dh-1843"></a>**DH-1843** The sign-up and sign-in pages show the hero photo. ([the photos of a brand directory](brand.md#dh-0328))
- <a id="dh-1844"></a>**DH-1844** Each photo is decoration, with an empty text alternative. Text on the page gives the meaning of each photo. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]; [Judgement: an empty text alternative for the photos](#dh-1888))
- <a id="dh-1845"></a>**DH-1845** When the brand has no photo for a place, the page shows a shape in the brand colour with the mark. The page never shows a broken image. ([the photos of a brand directory](brand.md#dh-0328))
- <a id="dh-1846"></a>**DH-1846** No text is on a photo. A text in front of a photo is on a solid panel, so that its contrast does not depend on the photo. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3])
- <a id="dh-1847"></a>**DH-1847** The accent colour of a brand is only for shapes, not for text. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3]; [Judgement: the accent colour only for shapes](#dh-1889))
- <a id="dh-1848"></a>**DH-1848** For each preset brand, each pair of colours that a layout uses for text, or for the boundary of a control, is in the contrast test of the design system. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3]; [WCAG 2.2, SC 1.4.11][wcag-1-4-11]; [Judgement: the accent colour only for shapes](#dh-1889))

## Footer

- <a id="dh-1849"></a>**DH-1849** Each page has a footer with two statements. The password statement is first, and the text of the footer of the consumer pages is second. ([Judgement: the password statement in the footer](#dh-1891); [the footer of the consumer pages](authorisation.md#dh-0837))
- <a id="dh-1850"></a>**DH-1850** The first statement of the footer says that third parties do not need the password of the consumer to get CDR data. ([CDS: Common Authentication Standards][cds-cx-authentication]; [Judgement: the password statement in the footer](#dh-1891))
- <a id="dh-1851"></a>**DH-1851** The second statement of the footer is the text of the [footer of the consumer pages](authorisation.md#dh-0837). The site API gives this text. A brand directory cannot change or remove it. ([the data holder gives the text of the footer](authorisation.md#dh-0838))
- <a id="dh-1852"></a>**DH-1852** The footer has no link. ([design system: Consumer pages][ui-consumer-pages]; [Judgement: no CDR policy on the site](#dh-1892))

## Accessibility

- <a id="dh-1853"></a>**DH-1853** The site meets WCAG 2.2 Level AA, with the brand of each preset brand. ([Judgement: one design system (ADR 0016)][adr-0016]; [design system: Accessibility][ui-accessibility])
- <a id="dh-1854"></a>**DH-1854** The automated accessibility checks pass on each page, with the brand of each preset brand. ([Judgement: one design system (ADR 0016)][adr-0016]; [design system: Accessibility][ui-accessibility])
- <a id="dh-1855"></a>**DH-1855** Each page is designed for a viewport width of 360 CSS px. ([design system: Layout][ui-layout])
- <a id="dh-1856"></a>**DH-1856** Each page is usable at a viewport width of 320 CSS px, with no horizontal scroll. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10]; [design system: Layout][ui-layout])

## Non-functional requirements

The site API also obeys [route patterns in place of raw paths](../../../pkg/platform/telemetry/DESIGN.md#tele-0005) and [no query string in a span](../../../pkg/platform/telemetry/DESIGN.md#tele-0007) of the Telemetry library, and the [secret values that telemetry never contains](non-functional-requirements.md#dh-0007), for example the session ID and the password. The [consumer dashboard](consumer-dashboard.md#dh-1940) gives the telemetry of the dashboard.

- <a id="dh-1857"></a>**DH-1857** Telemetry: the site API has the same HTTP server telemetry as the other listeners of the data holder. ([traces, metrics and logs with OpenTelemetry](../../../pkg/platform/telemetry/DESIGN.md#tele-0001); [Judgement: OpenTelemetry for traces, metrics and logs (ADR 0006)][adr-0006])
- <a id="dh-1858"></a>**DH-1858** Telemetry: no span, metric or log of the site contains a request body or a cookie value. ([no secret values in telemetry](non-functional-requirements.md#dh-0007); [no personal information in telemetry](non-functional-requirements.md#dh-0008))
- <a id="dh-1859"></a>**DH-1859** Telemetry: no log of the site API contains a query string or a path value. ([no personal information in telemetry](non-functional-requirements.md#dh-0008); [Judgement: OpenTelemetry for traces, metrics and logs (ADR 0006)][adr-0006])
- <a id="dh-1860"></a>**DH-1860** Telemetry: no log of an open or a close of an account contains the product ID or an arrangement ID. ([no personal information in telemetry](non-functional-requirements.md#dh-0008); [Judgement: no product ID and no data recipient name in the telemetry of the site](#dh-1894))
- <a id="dh-1861"></a>**DH-1861** Telemetry: the web UI of the site sends no telemetry. ([Judgement: web UIs as single-page applications that the service embeds (ADR 0014)][adr-0014])
- <a id="dh-1862"></a>**DH-1862** Performance: on a development workstation, the landing page and the product list of the largest preset catalogue show within 1 second. ([Judgement: the performance limits of the site](#dh-1895))
- <a id="dh-1863"></a>**DH-1863** Performance: on a development workstation, the first page of 10,000 transactions of an account shows within 1 second. ([Judgement: the performance limits of the site](#dh-1895))

## Judgements

### <a id="dh-1880"></a>DH-1880 One text of the landing page for each brand

- Kind: choice
- Basis: Step 3.3. The brand directory has no field for the text of the landing page ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023]).
- Reason: A new field changes the contract of the brand directory, and the site needs no text for each brand. The site API gives the brand name and the text of the footer, so the footer has one source, as on the consumer pages.

### <a id="dh-1881"></a>DH-1881 The product categories on the landing page

- Kind: choice
- Basis: Step 3.3. The CDS gives the product categories, but no rule for a web site of a data holder ([CDS: Product Categories][cds-product-categories]).
- Reason: A consumer looks for a type of product first.

### <a id="dh-1882"></a>DH-1882 A statement about CDR sharing on the landing page

- Kind: choice
- Basis: Step 3.1. The consumer dashboard is the online service of the data holder for CDR sharing ([CDR Rules][cdr-rules], rule 1.15).
- Reason: The site is where a consumer can learn about the CDR sharing of the brand. The consumer dashboard is on the site.

### <a id="dh-1883"></a>DH-1883 No claim and no copied text on the site

- Kind: choice
- Basis: Step 3.1. The sources give no rule for the text of a mock data holder.
- Reason: The brand is invented. The text of the site must not deceive a person.

### <a id="dh-1884"></a>DH-1884 The order of the products in a group

- Kind: choice
- Basis: Step 3.3. The CDS gives the order of the categories, not the order of the products in a category ([CDS: Product Categories][cds-product-categories]).
- Reason: A consumer finds a product by its name.

### <a id="dh-1885"></a>DH-1885 The name and the description in the product list

- Kind: choice
- Basis: Step 3.3. `name` and `description` are required fields of a product ([CDS: BankingProductV6][cds-schema-product]).
- Reason: The name and the description are the required fields that a consumer reads. The other fields are on the detail of the product.

### <a id="dh-1886"></a>DH-1886 Routes in the browser

- Kind: choice
- Basis: Step 3.3. Each web UI routes in the browser ([ADR 0014][adr-0014]).
- Reason: The site routes in the browser, as the consoles do. Thus the site, not the server, shows the page for a path that it does not know.

### <a id="dh-1887"></a>DH-1887 A layout for each brand

- Kind: choice
- Basis: Step 3.3. The sources give no rule for the look of a web site of a data holder.
- Reason: With the logo and the brand name hidden, a person must see which brand a page is from. Colours alone do not do this.

### <a id="dh-1888"></a>DH-1888 An empty text alternative for the photos

- Kind: choice
- Basis: Step 3.2. A decoration has an empty text alternative ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]).
- Reason: Each photo is decoration, and the text of the page gives its meaning. A field for the text in the brand directory changes its contract and gives no information.

### <a id="dh-1889"></a>DH-1889 The accent colour only for shapes

- Kind: choice
- Basis: Step 3.2. Text needs a contrast of 4.5:1 ([WCAG 2.2, SC 1.4.3][wcag-1-4-3]).
- Reason: The accent of some preset brands has a contrast lower than 4.5:1 on white or on the hero colour. The contrast test of the preset brands checks only the pairs that the layouts use for text.

### <a id="dh-1890"></a>DH-1890 A site that does not look like the consumer pages

- Kind: choice
- Basis: Step 3.1. The CDS forbids a reference to a different service in the authentication flow ([CDS: Common Authentication Standards][cds-cx-authentication]), so the difference is on the site.
- Reason: The consumer must not confuse the site with the CDR authentication. The sign-in of the site asks for a password, and the CDR authentication never does.

### <a id="dh-1891"></a>DH-1891 The password statement in the footer

- Kind: choice
- Basis: Step 3.3. The CDS requires the statement "in consumer-facing interactions and communications" ([CDS: Common Authentication Standards][cds-cx-authentication]), and the site is a consumer-facing interaction of the data holder.
- Reason: The footer is on each page, and a consumer can open each page first.

### <a id="dh-1892"></a>DH-1892 No CDR policy on the site

- Kind: gap
- Basis: Step 1. Rule 7.2(8) requires that the CDR policy is readily available through each online service of the data holder, and so through the online bank site ([CDR Rules][cdr-rules]).
- Reason: A CDR policy is a legal document of a real entity. The footer says that the brand is not real.

### <a id="dh-1893"></a>DH-1893 "Bank" before the brand name is known

- Kind: choice
- Basis: Step 3.3. The site API is the only source of data for the web UI ([ADR 0014][adr-0014]).
- Reason: The site uses the same rule as the operator console, which says "Data holder". "Bank" is true for each brand.

### <a id="dh-1894"></a>DH-1894 No product ID and no data recipient name in the telemetry of the site

- Kind: choice
- Basis: Step 3.1. The telemetry contains no personal information ([ADR 0006][adr-0006]).
- Reason: With the session, a product ID tells which product a consumer has. The name of a data recipient tells with which app a consumer shares data.

### <a id="dh-1895"></a>DH-1895 The performance limits of the site

- Kind: choice
- Basis: Step 3.3. The CDS performance requirements are for the CDR APIs, not for a web site ([the response time target of the APIs](non-functional-requirements.md#dh-0001)).
- Reason: The site uses the same limits as the operator consoles.

[readme-component-types]: ../../../README.md#component-types
[adr-0006]: ../../../docs/adr/0006-telemetry.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0016]: ../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[adr-0017]: ../../../docs/adr/0017-console-api-contracts-in-openapi.md
[adr-0023]: ../../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md
[cdr-rules]: https://www.legislation.gov.au/F2020L00094/latest/text
[cds-get-products]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-products
[cds-schema-product]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductv6
[cds-schema-product-detail]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_schemas_tocSbankingproductdetailv7
[cds-common-field-types]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[cds-cx-authentication]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#consumer-experience_common-authentication-standards
[ui-brands]: ../../../web/packages/ui/DESIGN.md#brands
[ui-consumer-pages]: ../../../web/packages/ui/DESIGN.md#consumer-pages
[ui-content]: ../../../web/packages/ui/DESIGN.md#content
[ui-accessibility]: ../../../web/packages/ui/DESIGN.md#accessibility
[ui-layout]: ../../../web/packages/ui/DESIGN.md#layout
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-4-3]: https://www.w3.org/TR/WCAG22/#contrast-minimum
[wcag-1-4-10]: https://www.w3.org/TR/WCAG22/#reflow
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
