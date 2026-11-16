# Mock data holder: Brand

This area gives the brand of a data holder instance: the brand directory, the preset brands, and the name, logo, colours and fonts that the data holder shows.

## Identity of the brand

- <a id="dh-0300"></a>**DH-0300** One instance of the data holder is one data holder brand. ([CDR Support: White Labelled brands in the CDR][support-white-label]; [CDS: Multi-Brand Support][cds-multi-brand])
- <a id="dh-0301"></a>**DH-0301** Each brand has its own product reference data endpoints, and its own Get Status and Get Outages endpoints. ([CDR Support: White Labelled brands in the CDR][support-white-label])
- <a id="dh-0302"></a>**DH-0302** The instance gets the identity of its brand from its configuration. For example, the configuration selects the brand directory and gives the brand ID. ([Judgement: the brand of an instance from its configuration](#dh-0370))
- <a id="dh-0303"></a>**DH-0303** The instance gets its base URIs from its configuration. ([Judgement: the brand of an instance from its configuration](#dh-0370))
- <a id="dh-0304"></a>**DH-0304** The brand name is the `brandName` of the brand on the Register. ([CDS: Register API schemas][cds-register-schemas])
- <a id="dh-0305"></a>**DH-0305** The brand directory gives the brand name. The data holder has no other setting for the brand name. ([Judgement: the brand directory is the one source of the brand name](#dh-0371))
- <a id="dh-0306"></a>**DH-0306** The consumer pages, the operator console and the OTP email show the brand name. ([the brand on the consumer pages](authorisation.md#dh-0836); [the brand in the console](operator-console.md#dh-1466); [the brand name as the sender name of the OTP email](consumer-authentication.md#dh-1244))
- <a id="dh-0307"></a>**DH-0307** The brand ID is the `dataHolderBrandId` of the brand on the Register. The configuration gives it. ([CDS: Register API schemas][cds-register-schemas])
- <a id="dh-0308"></a>**DH-0308** When the data holder calls a software product, the client ID of the brand is its brand ID. ([CDS: Data Holders calling Data Recipients][cds-dh-calling-dr])
- <a id="dh-0309"></a>**DH-0309** The brand has a record on the Register. Its `publicBaseUri` and its `productBaseUri` are the base URIs of the instance. ([CDS: Participant Endpoints][cds-participant-endpoints]; [CDS: Register API schemas][cds-register-schemas])
- <a id="dh-0310"></a>**DH-0310** When the brand on the Register has no `productBaseUri`, the brand has no product reference data endpoints. ([CDS: Participant Endpoints][cds-participant-endpoints])
- <a id="dh-0311"></a>**DH-0311** The `infosecBaseUri` of the brand on the Register is the issuer of the instance. A software product finds the OpenID Provider configuration with it. ([CDS: Participant Endpoints][cds-participant-endpoints])
- <a id="dh-0312"></a>**DH-0312** The brand has its own issuer. A second brand is a second instance with a different issuer. ([CDS: Multi-Brand Support][cds-multi-brand])
- <a id="dh-0313"></a>**DH-0313** The `logoUri` of the brand on the Register is the URL of the logo that the instance serves on its public base URI. For example, `<public base URI>/brand/logo.svg`. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: the logo on the public base URI](#dh-0372))

## Selection of the brand

- <a id="dh-0314"></a>**DH-0314** The configuration selects the brand with exactly one of these two settings. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023]; [Judgement: brand settings whose names start with `DATA_HOLDER_BRAND`](#dh-0373))
  - `DATA_HOLDER_BRAND`: the name of a preset brand
  - `DATA_HOLDER_BRAND_DIR`: the path of a brand directory
- <a id="dh-0315"></a>**DH-0315** The two brand settings have no default. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0316"></a>**DH-0316** The subcommands `api`, `serve` and `seed` need one brand setting. The subcommand `worker` needs no brand setting. ([a process needs only the settings that it uses](non-functional-requirements.md#dh-0010); [Judgement: brand settings whose names start with `DATA_HOLDER_BRAND`](#dh-0373))
- <a id="dh-0317"></a>**DH-0317** When both brand settings are present, or neither, a subcommand that needs one does not start. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0318"></a>**DH-0318** The preset brands are brand directories that the data holder contains. Each instance can run each preset brand with no file outside the data holder. ([Judgement: the final image contains only the binary (ADR 0018)][adr-0018]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0319"></a>**DH-0319** A lab user adds a brand with a brand directory and configuration only, with no change to the code. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0320"></a>**DH-0320** A brand directory gives only the parts of the brand that are the same in each deployment. The configuration of the deployment gives these parts. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
  - the brand ID
  - the base URIs
  - the keys and the certificates
  - the sender address of the email

## Files of a brand directory

- <a id="dh-0321"></a>**DH-0321** A brand directory has the files of the table that follows, and no other file. Only the photos and `ATTRIBUTION.md` are optional. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])

| ID | File | Required | Content | Source |
|---|---|---|---|---|
| <a id="dh-0322"></a>**DH-0322** | `brand.json` | Yes | The fields of [the brand file](#dh-0331) | [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023] |
| <a id="dh-0323"></a>**DH-0323** | `tokens.json` | Yes | The values of the brand for the semantic colour roles, the body font, the heading font and the radius of the design system. They are in the token format of the design system, with literal values only. | [what a brand can change](../../../web/packages/ui/design/tokens-and-brands.md#ui-0031); [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023] |
| <a id="dh-0324"></a>**DH-0324** | `logo.svg` | Yes | The logo, legible on the brand colour of the header of the pages and on white | [the logo on the brand colour and on white](../../../web/packages/ui/design/consumer-pages.md#ui-0363) |
| <a id="dh-0325"></a>**DH-0325** | `logo.png` | Yes | The logo of `logo.svg` as a PNG image for the email, legible on the brand colour of the header of the email | [the logo in the OTP email](consumer-authentication.md#dh-1251); [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023] |
| <a id="dh-0326"></a>**DH-0326** | `mark.svg` | Yes | A square symbol of the brand, for small sizes | [a brand can change the logo](../../../web/packages/ui/design/tokens-and-brands.md#ui-0031) |
| <a id="dh-0327"></a>**DH-0327** | `products.json` | Yes | The product catalogue of the brand, in the format of [a load file](product-reference-data.md#dh-0234) | [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023] |
| <a id="dh-0328"></a>**DH-0328** | `hero.webp` | No | The hero photo of the landing page of the online bank site, with the aspect ratio of the layout | [the hero photo on the landing page](online-bank-site.md#dh-1841); [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023] |
| <a id="dh-0329"></a>**DH-0329** | `category-<anchor>.webp` | No | The photo of a CDS product category on the online bank site, with the anchor of the category. For example, `category-term-deposits.webp`. | [CDS: Product Categories][cds-product-categories]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023] |
| <a id="dh-0330"></a>**DH-0330** | `ATTRIBUTION.md` | No | The source, the author and the licence of each photo, 16 KiB or less. The data holder does not read or serve it. | [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023] |

- <a id="dh-0331"></a>**DH-0331** The brand file `brand.json` has these fields, and all are required. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023]; [Judgement: the field name `brandName`, as on the Register](#dh-0374))
  - `brandName`: the brand name
  - `bsb`: the BSB of the accounts that a consumer opens on the online bank site
  - `layout`: the layout of the online bank site
- <a id="dh-0332"></a>**DH-0332** The `brandName` has 1 to 200 characters, and it is not only white space. ([CDS: Register API schemas][cds-register-schemas])
- <a id="dh-0333"></a>**DH-0333** The `brandName` has no control character. For example, the data holder refuses a brand name with a line break. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0334"></a>**DH-0334** The `bsb` has 6 digits, and it obeys the rule of a `bsb` in the operator interface. ([the rule of a BSB](operator-interface.md#dh-0739); [Judgement: the BSB of the brand for an account that the consumer opens](online-bank-site-accounts.md#dh-2304))
- <a id="dh-0335"></a>**DH-0335** The `layout` is one of these values. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023]; [the layouts of the online bank site](online-bank-site.md#dh-1834))
  - `meadow`
  - `ledger`
  - `horizon`
  - `updraft`
- <a id="dh-0336"></a>**DH-0336** `tokens.json` has each token of a brand of the design system. A missing token is an error. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0337"></a>**DH-0337** A catalogue with no products is valid. For example, `products.json` can be `{"products": []}`. ([Judgement: a catalogue with no products is valid](#dh-0375))

## Validation of the brand directory

- <a id="dh-0338"></a>**DH-0338** A JSON file of the brand directory has only the fields of the contract. The data holder refuses an unknown field, and a file that the contract does not name. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0339"></a>**DH-0339** The contract of the brand directory has no format version. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0340"></a>**DH-0340** The data holder ignores each entry of the directory whose name starts with `..`. For example, it ignores the `..data` entry of a ConfigMap volume. ([Kubernetes: atomic_writer.go][k8s-atomic-writer]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0341"></a>**DH-0341** The body font and the heading font of a brand are each one of a fixed set of fonts that the data holder contains. A brand with a different font is a change to the design system. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0342"></a>**DH-0342** Each colour of `tokens.json` is in the sRGB gamut. ([colour values in OKLCH](../../../web/packages/ui/design/tokens-and-brands.md#ui-0012))
- <a id="dh-0343"></a>**DH-0343** Each pair of text and background roles of a brand has a contrast of 4.5:1 or more. Each focus ring and each control boundary has a contrast of 3:1 or more against its background. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3]; [WCAG 2.2, SC 1.4.11][wcag-1-4-11]; [the text contrast of a brand](../../../web/packages/ui/design/tokens-and-brands.md#ui-0013); [the focus ring contrast of a brand](../../../web/packages/ui/design/tokens-and-brands.md#ui-0014))
- <a id="dh-0344"></a>**DH-0344** At start-up, the data holder does not check the contrast of a brand. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0345"></a>**DH-0345** Each SVG file is well-formed XML with an `svg` root element in the SVG namespace and a `viewBox`. It is 100 KiB or less. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023]; [Judgement: SVG files in the SVG namespace, and a square mark](#dh-0376))
- <a id="dh-0347"></a>**DH-0347** The `viewBox` of `mark.svg` is square. ([Judgement: SVG files in the SVG namespace, and a square mark](#dh-0376))
- <a id="dh-0348"></a>**DH-0348** `logo.png` is a valid PNG image of 100 KiB or less. It is exactly 96 pixels high and 768 pixels wide or less. ([Judgement: a `logo.png` 96 pixels high and at most 768 pixels wide](#dh-0377))
- <a id="dh-0349"></a>**DH-0349** The aspect ratio of `logo.png` is the aspect ratio of the `viewBox` of `logo.svg`, with a tolerance of 2 %. The aspect ratio is the width divided by the height. ([Judgement: a `logo.png` 96 pixels high and at most 768 pixels wide](#dh-0377))
- <a id="dh-0350"></a>**DH-0350** Each JSON file is 1 MiB or less. ([Kubernetes: ConfigMaps][k8s-configmap]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0351"></a>**DH-0351** Each photo is a WebP image that is not animated. ([RFC 9649][rfc9649]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0352"></a>**DH-0352** `hero.webp` is 256 KiB or less and 1200 to 2400 pixels wide. Its aspect ratio is the aspect ratio of the layout, with a tolerance of 2 %. ([Judgement: photo sizes that are sharp at a pixel ratio of 2](#dh-0378))
- <a id="dh-0353"></a>**DH-0353** Each category photo is 96 KiB or less and 800 to 1200 pixels wide. Its aspect ratio is 4:3, with a tolerance of 2 %. ([Judgement: photo sizes that are sharp at a pixel ratio of 2](#dh-0378))
- <a id="dh-0354"></a>**DH-0354** The data holder reads the size of each image, and the animation flag of each photo, in the header of the file before it decodes the image. ([RFC 9649][rfc9649]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0355"></a>**DH-0355** The name of a category photo has the anchor of one of the CDS product categories. The data holder refuses a different name. ([CDS: Product Categories][cds-product-categories]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0356"></a>**DH-0356** A photo of a category with no current product is valid. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0357"></a>**DH-0357** The brand directory is 1 MiB or less in all. ([Kubernetes: ConfigMaps][k8s-configmap]; [Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0358"></a>**DH-0358** The process that runs the APIs, and the product load, validate the brand directory at start-up. They check each rule of this group, and the rules of the files and of the brand file. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])
- <a id="dh-0359"></a>**DH-0359** When a rule of the brand directory fails, the process does not start. Its error names the file and the rule. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023])

## Preset brands

- <a id="dh-0346"></a>**DH-0346** The data holder contains these four preset brands. `wattle-bank` is the default preset brand. ([Judgement: the name of the preset brand for each name of its instance](#dh-0379))
  - `wattle-bank`, with the brand name "Wattle Bank"
  - `redgum-bank`
  - `harbour-bank`
  - `kestrel-bank`
- <a id="dh-0360"></a>**DH-0360** In the lab, the compose service, the host name, the database and the certificate directory of an instance have the name of its preset brand. For example, the instance of `wattle-bank` has the host name `wattle-bank`. ([Judgement: the name of the preset brand for each name of its instance](#dh-0379))
- <a id="dh-0361"></a>**DH-0361** The brand names, the product names, the email addresses and the host names of each preset brand and of the lab configuration do not contain "Mock", "Example" or a similar word. ([Judgement: realistic names in the lab](#dh-0380))
- <a id="dh-0362"></a>**DH-0362** An invented domain uses the top-level domain `.test`. For example, `wattlebank.test`. ([RFC 6761, section 6.2][rfc6761-test]; [Judgement: `.test` for an invented domain](#dh-0381))
- <a id="dh-0363"></a>**DH-0363** By default, the lab runs one preset brand: the default preset brand. ([Judgement: one preset brand by default](deployment.md#dh-1360))
- <a id="dh-0364"></a>**DH-0364** On the Register of the lab, the default preset brand has its own legal entity, with one brand. ([Judgement: realistic names in the lab](#dh-0380))
- <a id="dh-0365"></a>**DH-0365** The lab can also run the four preset brands at the same time. Each is its own instance of the same data holder image, with its own issuer. The instances differ only in their configuration. ([CDS: Multi-Brand Support][cds-multi-brand]; [the compose profile of the other preset brands](deployment.md#dh-1303))

## Judgements

### <a id="dh-0370"></a>DH-0370 The brand of an instance from its configuration

- Kind: choice
- Basis: Step 3.3. The CDS gives each brand its own issuer ([CDS: Multi-Brand Support][cds-multi-brand]), and gives no rule for how an instance gets its brand.
- Reason: The lab can then run more than one brand from one image. A brand can change its host names with no change to the data holder.

### <a id="dh-0371"></a>DH-0371 The brand directory is the one source of the brand name

- Kind: choice
- Basis: Step 3.3. [ADR 0023][adr-0023] puts the brand name in `brand.json`, and does not say whether a setting can also give it.
- Reason: A second source of the same name can disagree with the directory. The Register record of the brand has the same name.

### <a id="dh-0372"></a>DH-0372 The logo on the public base URI

- Kind: choice
- Basis: Step 3.3. The CDS makes `logoUri` mandatory for a data holder brand, and does not say who serves the logo ([CDS: Register API schemas][cds-register-schemas]).
- Reason: A data recipient can then show the logo of the brand that the consumer selects. The `logoUri` needs a fixed URL on the public base URI.

### <a id="dh-0373"></a>DH-0373 Brand settings whose names start with `DATA_HOLDER_BRAND`

- Kind: choice
- Basis: Step 3.3. [ADR 0023][adr-0023] gives two settings and does not name them.
- Reason: The names of the other brand settings start with `DATA_HOLDER_BRAND`. `seed` needs the brand name to check the `brand` of each product, also for a load file.

### <a id="dh-0374"></a>DH-0374 The field name `brandName`, as on the Register

- Kind: choice
- Basis: Step 3.3. The Register gives the brand name as `brandName` ([CDS: Register API schemas][cds-register-schemas]).
- Reason: The brand file uses the field name of the Register, as a load file uses the names of the CDS.

### <a id="dh-0375"></a>DH-0375 A catalogue with no products is valid

- Kind: choice
- Basis: Step 3.3. A brand on the Register with no `productBaseUri` has no product reference data endpoints ([CDS: Participant Endpoints][cds-participant-endpoints]).
- Reason: A user brand that tests only the authentication and the authorisation needs no products.

### <a id="dh-0376"></a>DH-0376 SVG files in the SVG namespace, and a square mark

- Kind: choice
- Basis: Step 3.3. [ADR 0023][adr-0023] gives the SVG structure, and the design system shows the mark in a row of the hero ([the row of the hero](../../../web/packages/ui/design/consumer-pages.md#ui-0365)).
- Reason: A browser does not show an SVG file with no SVG namespace as an image. A `viewBox` that is not square cannot give a square mark.

### <a id="dh-0377"></a>DH-0377 A `logo.png` 96 pixels high and at most 768 pixels wide

- Kind: choice
- Basis: Step 3.3. [ADR 0023][adr-0023] requires a fixed height, a maximum width and the aspect ratio of `logo.svg`, and does not give the numbers.
- Reason: The email shows the logo 48 CSS px high, and 96 pixels gives a sharp image on a screen with a pixel ratio of 2. A width of 768 pixels (an aspect ratio of 8) is more than each preset logo needs, and the size limit is the limit of the SVG files. The tolerance permits the rounding of the width to a whole pixel.

### <a id="dh-0378"></a>DH-0378 Photo sizes that are sharp at a pixel ratio of 2

- Kind: choice
- Basis: Step 3.3. [ADR 0023][adr-0023] requires a minimum and a maximum width for each photo, and does not give the numbers.
- Reason: The widths give a sharp photo on a screen with a pixel ratio of 2 at the size of the layouts. The limits keep a preset brand with seven photos in one ConfigMap ([Kubernetes: ConfigMaps][k8s-configmap]).

### <a id="dh-0379"></a>DH-0379 The name of the preset brand for each name of its instance

- Kind: choice
- Basis: Step 3.3. The CDS gives no rule for the names of the brands of a mock data holder.
- Reason: Each instance then has one pattern for its names. A second brand in the lab uses the same pattern.

### <a id="dh-0380"></a>DH-0380 Realistic names in the lab

- Kind: choice
- Basis: Step 3.3. Each Australian bank is an authorised deposit-taking institution and a legal entity of its own.
- Reason: A brand or a legal entity with a "Mock" or "Example" name is not a realistic bank. A client that the lab tests then sees the names that a production data holder gives.

### <a id="dh-0381"></a>DH-0381 `.test` for an invented domain

- Kind: choice
- Basis: Step 3.2. `.test` is reserved for use in testing ([RFC 6761, section 6.2][rfc6761-test]), and `.example` is also reserved ([RFC 2606, section 2][rfc2606-tlds]).
- Reason: A name with "Example" is not permitted. A realistic domain, for example one in `.com.au`, can belong to a real business.

[support-white-label]: https://cdr-support.zendesk.com/hc/en-us/articles/900003938166
[cds-dh-calling-dr]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#data-holders-calling-data-recipients
[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-multi-brand]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#multi-brand-support-separate-issuers-for-data-holder-brands
[cds-product-categories]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#product-categories
[adr-0018]: ../../../docs/adr/0018-container-images-with-a-web-ui-build-stage.md
[adr-0023]: ../../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md
[wcag-1-4-3]: https://www.w3.org/TR/WCAG22/#contrast-minimum
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
[k8s-configmap]: https://kubernetes.io/docs/concepts/configuration/configmap/
[k8s-atomic-writer]: https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/util/atomic_writer.go
[rfc9649]: https://www.rfc-editor.org/rfc/rfc9649
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[rfc2606-tlds]: https://www.rfc-editor.org/rfc/rfc2606#section-2
