# Mock data recipient: Brand

This area gives the identity of the brand of an instance, the selection of the brand, the files of a brand directory and their validation, and the preset brands.

## Identity of the brand

- <a id="dr-0100"></a>**DR-0100** One instance of the data recipient is one data recipient brand. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0101"></a>**DR-0101** The instance gets the identity of its brand from its configuration. For example, the configuration selects the brand directory and gives the brand ID. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0102"></a>**DR-0102** The brand ID is the `dataRecipientBrandId` of the brand on the Register. The configuration gives it. ([CDS: Register API schemas][cds-register-schemas])
- <a id="dr-0103"></a>**DR-0103** The brand ID is a value that a cookie name can contain. For example, a UUID is valid. The instance does not start with a brand ID that is not valid. ([RFC 6265, section 4.1.1][rfc6265-syntax]; [the session cookie name](sign-in.md#dr-0430))
- <a id="dr-0104"></a>**DR-0104** The brand name is the `brandName` of the brand on the Register. ([CDS: Register API schemas][cds-register-schemas])
- <a id="dr-0105"></a>**DR-0105** The brand directory gives the brand name. The data recipient has no other setting for the brand name. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0106"></a>**DR-0106** The web app and each email show the brand name and the logo of the brand. ([the header of each page](web-app.md#dr-0220); [the header of each email](email.md#dr-0504))

## Selection of the brand

- <a id="dr-0107"></a>**DR-0107** The configuration selects the brand with exactly one of these two settings. ([Judgement: brand settings whose names start with `DATA_RECIPIENT_BRAND`](#dr-0151))
  - `DATA_RECIPIENT_BRAND`: the name of a preset brand
  - `DATA_RECIPIENT_BRAND_DIR`: the path of a brand directory
- <a id="dr-0108"></a>**DR-0108** The two brand settings have no default. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0109"></a>**DR-0109** The subcommands `api` and `serve` need one brand setting. The subcommand `migrate` needs no brand setting. ([a process needs only the settings that it uses](non-functional-requirements.md#dr-0010))
- <a id="dr-0110"></a>**DR-0110** When both brand settings are present, or neither, a subcommand that needs one does not start. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0111"></a>**DR-0111** The preset brands are brand directories that the data recipient contains. Each instance can run each preset brand with no file outside the data recipient. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0112"></a>**DR-0112** A lab user adds a brand with a brand directory and configuration only, with no change to the code. ([Judgement: a brand directory as the data holder has](#dr-0150))

## Files of a brand directory

- <a id="dr-0113"></a>**DR-0113** A brand directory has these files, and no other file. Only the photos and `ATTRIBUTION.md` are optional. ([Judgement: a brand directory as the data holder has](#dr-0150))
  - `brand.json`: the brand file
  - `tokens.json`: the design tokens of the brand
  - `logo.svg`: the logo, for the header of the web app
  - `mark.svg`: a square symbol of the brand, for the icon of the web app
  - `logo.png`: the logo, for the header of each email
  - `hero.webp`: the photo of the landing page
  - `access.webp`: the photo of the sign-up, sign-in and password reset pages
  - `ATTRIBUTION.md`: the source, the author and the licence of each photo
- <a id="dr-0114"></a>**DR-0114** The brand file `brand.json` has only the field `brandName`, and the field is required. ([CDS: Register API schemas][cds-register-schemas])
- <a id="dr-0115"></a>**DR-0115** The data recipient does not read or serve `ATTRIBUTION.md`. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0116"></a>**DR-0116** The logo is legible on the header colour of the brand and on white. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11])

## Validation of the brand directory

- <a id="dr-0117"></a>**DR-0117** A brand directory obeys each rule of the brand directory library, and the contrast rules of the design system. These are the groups of rules. ([Judgement: a brand directory as the data holder has](#dr-0150))
  - [the files of each brand directory](../../../pkg/platform/branddir/DESIGN.md#brand-0001) and [the rules of the `brandName`](../../../pkg/platform/branddir/DESIGN.md#brand-0006)
  - [the validation of a brand directory](../../../pkg/platform/branddir/DESIGN.md#brand-0009)
  - [the text contrast of a brand](../../../web/packages/ui/design/tokens-and-brands.md#ui-0013) and [the focus ring contrast of a brand](../../../web/packages/ui/design/tokens-and-brands.md#ui-0014)
- <a id="dr-0118"></a>**DR-0118** `hero.webp` is 256 KiB or less and 1200 to 2400 pixels wide. Its aspect ratio is 4:3, with a tolerance of 2 %. ([Judgement: a hero photo of 4:3 for a phone screen](#dr-0152))
- <a id="dr-0127"></a>**DR-0127** `access.webp` is 160 KiB or less and 800 to 1600 pixels wide. Its aspect ratio is 4:5, with a tolerance of 2 %. ([Judgement: a portrait photo of 4:5 next to a form](#dr-0154))
- <a id="dr-0119"></a>**DR-0119** At start-up, the data recipient does not check the contrast of a brand. A check in the tests of the web workspace checks each preset brand. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0120"></a>**DR-0120** Each process that needs a brand setting validates the brand directory at start-up. When a rule fails, the process does not start. Its error names the file and the rule. ([Judgement: a brand directory as the data holder has](#dr-0150))

## Preset brands

- <a id="dr-0121"></a>**DR-0121** The data recipient contains the preset brand `sorrel`, with the brand name "Sorrel". It is the default preset brand. ([Judgement: a brand directory as the data holder has](#dr-0150))
- <a id="dr-0122"></a>**DR-0122** The look of `sorrel` is light, playful and warm. Its primary colours are a leaf green and a rose, on a cream background. Its photos show gardens, markets and people outdoors. ([Judgement: two brands that are different in each part of their look](#dr-0153))
- <a id="dr-0123"></a>**DR-0123** The primary colours of each preset brand are not the primary colours of a preset data holder brand. For example, `sorrel` has no yellow, red or blue primary colour. ([Judgement: two brands that are different in each part of their look](#dr-0153))
- <a id="dr-0124"></a>**DR-0124** The brand names, the email addresses and the host names of each preset brand and of the lab configuration do not contain "Mock", "Example" or a similar word. ([the realistic names of the data holder](../../dataholder/design/brand.md#dh-0361))
- <a id="dr-0125"></a>**DR-0125** An invented domain uses the top-level domain `.test`. For example, `sorrel.test`. ([RFC 6761, section 6.2][rfc6761-test])
- <a id="dr-0126"></a>**DR-0126** Each photo of a preset brand has a licence that permits copies and distribution. `ATTRIBUTION.md` names each photo, its author and its licence. ([Unsplash License][unsplash-license])

## Judgements

### <a id="dr-0150"></a>DR-0150 A brand directory as the data holder has

- Kind: choice
- Basis: Step 3.3. The CDS gives the brand name and the logo of a data recipient brand on the Register ([CDS: Register API schemas][cds-register-schemas]), and says nothing about how the software gets them. The data holder gets its brand from a strict brand directory ([ADR 0023][adr-0023]).
- Reason: One format of brand directory for the two kinds of instance is the least complex for a lab user. The data recipient has no product catalogue, no BSB, no layout and no category photos, so its directory has fewer files.

### <a id="dr-0151"></a>DR-0151 Brand settings whose names start with `DATA_RECIPIENT_BRAND`

- Kind: choice
- Basis: Step 3.3. The data holder uses `DATA_HOLDER_BRAND` and `DATA_HOLDER_BRAND_DIR` ([the brand settings of the data holder](../../dataholder/design/brand.md#dh-0314)).
- Reason: The same names, with the name of the component, let a lab user configure each kind of instance in the same way.

### <a id="dr-0152"></a>DR-0152 A hero photo of 4:3 for a phone screen

- Kind: choice
- Basis: Step 3.3. The web app is mobile-first ([the web app is mobile-first](web-app.md#dr-0216)). The data holder gives the sizes of its photos for a pixel ratio of 2 ([photo sizes of the data holder](../../dataholder/design/brand.md#dh-0378)).
- Reason: A 4:3 photo fills the width of a phone screen and leaves space for the text and the actions of the landing page. One aspect ratio is the least complex.

### <a id="dr-0153"></a>DR-0153 Two brands that are different in each part of their look

- Kind: choice
- Basis: Step 3.3. The authorisation page of a data holder shows the brand of the data holder and the brand name of the data recipient together ([the data recipient on the authorisation page](../../dataholder/design/authorisation.md#dh-0802)).
- Reason: Brands that are different in each part show a part of the web app that ignores the brand. Colours that are not the colours of a bank help the consumer see which brand is which.

### <a id="dr-0154"></a>DR-0154 A portrait photo of 4:5 next to a form

- Kind: choice
- Basis: Step 3.3. The pages of the membership have a form, and the web app has a layout 1280 CSS pixels wide ([the web app is mobile-first](web-app.md#dr-0216)). The data holder gives the sizes of its photos for a pixel ratio of 2 ([photo sizes of the data holder](../../dataholder/design/brand.md#dh-0378)).
- Reason: A portrait photo fills the side of a wide page next to a narrow form, and a crop of it gives a band above the form on a phone. A width of 800 to 1600 pixels is sharp at a pixel ratio of 2 in a panel up to 800 CSS pixels wide. The size limit keeps a brand with two photos far below the limit of the directory.

[adr-0023]: ../../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[rfc6265-syntax]: https://www.rfc-editor.org/rfc/rfc6265#section-4.1.1
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[unsplash-license]: https://unsplash.com/license
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
