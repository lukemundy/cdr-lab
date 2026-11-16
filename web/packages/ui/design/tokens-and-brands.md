# Design system: Tokens and brands

This area gives the design tokens, their tiers, the colour roles and their contrast, the scales and the densities, the brands, and the fonts.

## Token format and tiers

- <a id="ui-0001"></a>**UI-0001** The design system gives its tokens in the DTCG 2025.10 format, in the package [`web/packages/tokens`](../../tokens/). ([DTCG Format][dtcg-format]; [DTCG Resolver][dtcg-resolver]; [Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0002"></a>**UI-0002** A UI gets each token as a CSS custom property. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0003"></a>**UI-0003** The tokens have three tiers: primitive, semantic and component. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0004"></a>**UI-0004** A primitive token is a raw value. For example, a colour ramp or a font family. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0005"></a>**UI-0005** A component never uses a primitive token or a raw colour value. For example, a button uses the primary action role, not a step of a colour ramp. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0006"></a>**UI-0006** A semantic token is a role or a scale that components use. For example, a colour role, a font role or the spacing scale. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0007"></a>**UI-0007** Each brand gives a value for each semantic role. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0008"></a>**UI-0008** A component token is only for a value that one component needs and that no semantic role gives. ([Judgement: one design system (ADR 0016)][adr-0016])

## Colour roles and contrast

- <a id="ui-0010"></a>**UI-0010** The semantic colour roles are these. ([Judgement: one design system (ADR 0016)][adr-0016])
  - page background
  - surface
  - raised surface
  - modal backdrop
  - default text
  - muted text
  - text on a primary action
  - heading
  - link
  - brand (the header of a consumer page), and text on the brand
  - hero (the band below the header), and text on the hero
  - accent
  - border
  - strong border
  - focus ring
  - primary action, with its hover and pressed states
  - secondary action
  - success, warning, danger and information, each with a background, a border and a text value
- <a id="ui-0011"></a>**UI-0011** The modal backdrop is a dark colour with transparency. ([Judgement: the colour of the modal backdrop](#ui-0058))
- <a id="ui-0012"></a>**UI-0012** Each colour value is in OKLCH. ([CSS Color 4: OKLCH][css-oklch]; [Judgement: OKLCH colour values](#ui-0055))
- <a id="ui-0013"></a>**UI-0013** In each brand, each pair of a text role and a background role has a contrast of 4.5:1 or more. For example, the muted text on the surface. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3])
- <a id="ui-0014"></a>**UI-0014** In each brand, the focus ring and the strong border have a contrast of 3:1 or more against each background role. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11])
- <a id="ui-0015"></a>**UI-0015** The text on a primary action has a contrast of 4.5:1 or more against the primary action and its hover and pressed states, in each brand. ([WCAG 2.2, SC 1.4.3][wcag-1-4-3])
- <a id="ui-0016"></a>**UI-0016** In the console of a participant, each pair of a brand role and a lab brand role that the console shell shows has the same contrast as a pair in one brand. For example, the link of the brand on the page header band. ([text contrast](#ui-0013); [focus ring contrast](#ui-0014); [Judgement: the brand in a console](#ui-0060))
- <a id="ui-0017"></a>**UI-0017** The primary action colour has no contrast minimum against the background roles. ([WCAG 2.2, Understanding SC 1.4.11][understanding-1-4-11]; [Judgement: the contrast pairs of the primary action colour](#ui-0057))
- <a id="ui-0018"></a>**UI-0018** A link uses the link role, not the primary action colour. ([Judgement: the contrast pairs of the primary action colour](#ui-0057))
- <a id="ui-0019"></a>**UI-0019** The strong border is the boundary of a control. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11])
- <a id="ui-0020"></a>**UI-0020** The border role and the status borders are decorative, and they have no contrast minimum. For example, a divider between table rows uses the border role. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]; [Judgement: the contrast of the border role](#ui-0056))
- <a id="ui-0021"></a>**UI-0021** The accent is decoration only. No information depends on it, and it has no contrast minimum. ([WCAG 2.2, SC 1.4.1][wcag-1-4-1]; [WCAG 2.2, SC 1.4.11][wcag-1-4-11])
- <a id="ui-0022"></a>**UI-0022** The modal backdrop has no contrast minimum. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]; [Judgement: the colour of the modal backdrop](#ui-0058))

## Scales and densities

- <a id="ui-0025"></a>**UI-0025** The tokens that are not colours are these. ([Judgement: one design system (ADR 0016)][adr-0016])
  - a spacing scale on a 4 px base
  - a radius scale
  - three elevation levels
  - a type scale
  - line heights
  - font weights
  - motion durations
- <a id="ui-0026"></a>**UI-0026** The design system has two densities: compact and comfortable. ([Judgement: two densities](#ui-0066))
- <a id="ui-0027"></a>**UI-0027** A density changes only the spacing, the control heights and the table row heights. ([Judgement: two densities](#ui-0066))
- <a id="ui-0028"></a>**UI-0028** A console uses the compact density. ([Judgement: two densities](#ui-0066))
- <a id="ui-0029"></a>**UI-0029** A consumer UI uses the comfortable density, with controls 48 CSS px high. ([Judgement: two densities](#ui-0066))

## Brands

- <a id="ui-0030"></a>**UI-0030** A brand is a set of values for the brand modifier of the tokens. ([DTCG Resolver][dtcg-resolver]; [Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0031"></a>**UI-0031** A brand can change only these: the colour roles, the body font, the heading font, the radius and the logo. ([Judgement: what a brand can change](#ui-0059))
- <a id="ui-0032"></a>**UI-0032** A brand cannot change the spacing scale, the type scale or the structure of a component. ([Judgement: what a brand can change](#ui-0059))
- <a id="ui-0033"></a>**UI-0033** The lab brand is the neutral brand of the design system. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0034"></a>**UI-0034** A console uses the lab brand for each role that the brand of a participant does not give to the console. ([Judgement: the brand in a console](#ui-0060))
- <a id="ui-0035"></a>**UI-0035** The console of a participant uses these roles of the brand of that participant. ([Judgement: the brand in a console](#ui-0060))
  - brand, and text on the brand
  - hero, and text on the hero
  - accent
  - heading
  - link
  - primary action, with its hover and pressed states
  - text on a primary action
- <a id="ui-0036"></a>**UI-0036** The console of a participant uses the body font, the heading font and the radius of the lab brand. ([Judgement: the brand in a console](#ui-0060))

## Typography

- <a id="ui-0040"></a>**UI-0040** Each UI uses self-hosted fonts only. No font comes from a different origin. ([OWASP ASVS 5.0, V3.6.1][asvs-v3])
- <a id="ui-0041"></a>**UI-0041** The fonts are a fixed set, which [`fonts.json`](../../tokens/fonts.json) gives. ([Judgement: a fixed set of fonts](#ui-0065))
- <a id="ui-0042"></a>**UI-0042** A brand uses a font of the set. ([Judgement: a fixed set of fonts](#ui-0065))
- <a id="ui-0043"></a>**UI-0043** Each font of the set has a licence that permits self-hosting. Its licence text is next to its files. ([OWASP ASVS 5.0, V3.6.1][asvs-v3]; [Judgement: a fixed set of fonts](#ui-0065))
- <a id="ui-0044"></a>**UI-0044** The lab brand uses Inter for text. ([Judgement: the fonts of the lab brand](#ui-0061))
- <a id="ui-0045"></a>**UI-0045** The lab brand uses JetBrains Mono for identifiers, thumbprints and keys. ([Judgement: the fonts of the lab brand](#ui-0061))
- <a id="ui-0046"></a>**UI-0046** Amounts and numbers in tables use tabular figures. ([Judgement: tabular figures](#ui-0062))
- <a id="ui-0047"></a>**UI-0047** In a consumer UI, body text is 16 CSS px or more. ([Judgement: the smallest body text in a consumer UI](#ui-0063))
- <a id="ui-0048"></a>**UI-0048** A user can resize text to 200 % with no loss of content or function. ([WCAG 2.2, SC 1.4.4][wcag-1-4-4])
- <a id="ui-0049"></a>**UI-0049** Each heading level from h1 to h4 has a smaller font size than the level above it. ([Judgement: the heading sizes](#ui-0064))
- <a id="ui-0050"></a>**UI-0050** The headings h4, h5 and h6 have the same font size. ([Judgement: the heading sizes](#ui-0064))

## Judgements

### <a id="ui-0055"></a>UI-0055 OKLCH colour values

- Kind: choice
- Basis: Step 3.2. Lightness steps in OKLCH are even ([CSS Color 4: OKLCH][css-oklch]). Each supported browser supports `oklch()` ([MDN: oklch()][mdn-oklch]).
- Reason: Even lightness steps make the colour ramp of a new brand easier. A ramp then needs less adjustment.

### <a id="ui-0056"></a>UI-0056 The contrast of the border role

- Kind: choice
- Basis: Step 3.2. SC 1.4.11 does not apply to a border that is not necessary to identify a component or its state ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]).
- Reason: Only the boundary of a control and the focus ring need 3:1. A divider of 3:1 on each table row makes a dense console difficult to read.

### <a id="ui-0057"></a>UI-0057 The contrast pairs of the primary action colour

- Kind: choice
- Basis: Step 3.2. A button with a text label needs no boundary of 3:1, because the text identifies the component ([WCAG 2.2, Understanding SC 1.4.11][understanding-1-4-11]).
- Reason: A brand with a light primary colour, for example yellow, can then use it for its buttons. The text on the primary action keeps 4.5:1, and the link role keeps 4.5:1 against each background. The pairs of the focus ring and the strong border stay.

### <a id="ui-0058"></a>UI-0058 The colour of the modal backdrop

- Kind: choice
- Basis: Step 3.2. The backdrop is not text and not the boundary of a control ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]).
- Reason: A dimmed page shows that the dialog is modal. A semantic role lets a brand change it.

### <a id="ui-0059"></a>UI-0059 What a brand can change

- Kind: choice
- Basis: Step 3.3. [ADR 0016][adr-0016] gives one set of components for more than one brand.
- Reason: The brands must look different, but a component must work the same in each brand. An app can arrange the components differently for each brand, as [the online bank site of the data holder](../../../../services/dataholder/design/online-bank-site.md) does with its layouts.

### <a id="ui-0060"></a>UI-0060 The brand in a console

- Kind: choice
- Basis: Step 3.3. [ADR 0016][adr-0016] says that each participant has its own brand, and that the consoles have a neutral lab brand.
- Reason: The console of a participant shows more of its brand in the shell and stays a dense tool for the lab operator. The brand colour is white for some brands, so the side navigation uses the hero colour, which is dark in each preset brand. A different font changes the width of each table column, so the font stays the same.

### <a id="ui-0061"></a>UI-0061 The fonts of the lab brand

- Kind: choice
- Basis: Step 3.3. Inter and JetBrains Mono have the SIL Open Font License 1.1, which permits self-hosting.
- Reason: Inter has tabular figures ([tabular figures](#ui-0046)). In JetBrains Mono, the 0 and the O, and the 1, the l and the I, are clearly different. A lab operator needs this to read an identifier or a thumbprint.

### <a id="ui-0062"></a>UI-0062 Tabular figures

- Kind: choice
- Basis: Step 3.3. No source gives a figure style for a UI.
- Reason: Digits with one width align in columns.

### <a id="ui-0063"></a>UI-0063 The smallest body text in a consumer UI

- Kind: choice
- Basis: Step 3.3. No source gives a minimum text size.
- Reason: iOS Safari zooms into a form field with smaller text, which moves the layout.

### <a id="ui-0064"></a>UI-0064 The heading sizes

- Kind: choice
- Basis: Step 3.3. No source gives the sizes of the heading levels.
- Reason: The user sees the structure of a page that has nested sections.

### <a id="ui-0065"></a>UI-0065 A fixed set of fonts

- Kind: choice
- Basis: Step 3.2. Self-hosted fonts keep each other origin out of the CSP ([OWASP ASVS 5.0, V3.6.1][asvs-v3]).
- Reason: One list lets the build self-host each font, and a reviewer checks the licence of a font one time.

### <a id="ui-0066"></a>UI-0066 Two densities

- Kind: choice
- Basis: Step 3.3. [ADR 0016][adr-0016] says that the consoles are for desktop and data-dense, and that the consumer UIs are for mobile.
- Reason: A console shows more records on one screen, and a consumer UI has controls for touch. A density changes only sizes, so a component works the same in each density.

[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[dtcg-format]: https://www.designtokens.org/TR/2025.10/format/
[dtcg-resolver]: https://www.designtokens.org/TR/2025.10/resolver/
[css-oklch]: https://www.w3.org/TR/css-color-4/#ok-lab
[mdn-oklch]: https://developer.mozilla.org/en-US/docs/Web/CSS/color_value/oklch
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
[wcag-1-4-1]: https://www.w3.org/TR/WCAG22/#use-of-color
[wcag-1-4-3]: https://www.w3.org/TR/WCAG22/#contrast-minimum
[wcag-1-4-4]: https://www.w3.org/TR/WCAG22/#resize-text
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
[understanding-1-4-11]: https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html
