# Design system: Consumer pages

This area gives the design guide of the consumer pages: the visual pattern of each part of a page. The specification of the service that has the pages gives the content of each page. A brand changes only [what a brand can change](tokens-and-brands.md#ui-0031).

## Header and hero

- <a id="ui-0360"></a>**UI-0360** The header of a consumer page has the logo of the brand on the brand colour, at the start, 32 CSS px high. The brand name is the text alternative of the logo. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="ui-0361"></a>**UI-0361** The header can have a "Secure" label with a lock icon, at the end. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0362"></a>**UI-0362** The accent colour is a line of 4 CSS px at the top of the header. ([the accent is decoration only](tokens-and-brands.md#ui-0021); [Judgement: the colour roles of the consumer pages](#ui-0393))
- <a id="ui-0363"></a>**UI-0363** The logo of a brand is legible on its brand colour and on white, because the Register gives the same file to data recipients. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="ui-0364"></a>**UI-0364** A page with a data recipient has a hero: a band in the hero colour below the header. A page with no data recipient has no hero. ([Judgement: the colour roles of the consumer pages](#ui-0393))
- <a id="ui-0365"></a>**UI-0365** The hero has a row in the centre with the initial of the data recipient, an arrow and the mark of the brand. Assistive technology does not announce the row. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="ui-0366"></a>**UI-0366** Below the row, the hero has one sentence in the centre. ([Judgement: one design system (ADR 0016)][adr-0016])

## Card

- <a id="ui-0367"></a>**UI-0367** A consumer page has one raised card for its task, at most 440 CSS px wide, in the centre. ([one column in a consumer UI](layout.md#ui-0087); [Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0368"></a>**UI-0368** The card overlaps the lower edge of the hero by 24 CSS px. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0369"></a>**UI-0369** The card has the large radius and two shadow layers. ([Judgement: the radius of the consumer pages](#ui-0394))
- <a id="ui-0370"></a>**UI-0370** A step indicator is the text "Step n of m" in small upper-case text above the heading, and a thin bar that shows the progress. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0371"></a>**UI-0371** The text of the step indicator gives the step. The bar is decoration, and assistive technology does not announce it. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="ui-0372"></a>**UI-0372** A consumer page has one h1 for its task: 24 CSS px, bold, letter spacing -0.015 em, in the heading colour. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; [Judgement: the colour roles of the consumer pages](#ui-0393))
- <a id="ui-0373"></a>**UI-0373** The sections of the card use h2 and h3. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])

## Controls

- <a id="ui-0374"></a>**UI-0374** The label of a text field is above the field. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2])
- <a id="ui-0375"></a>**UI-0375** A text field is the height of a control, with the strong border and the small radius. ([the control size](layout.md#ui-0093); [Judgement: the radius of the consumer pages](#ui-0394))
- <a id="ui-0376"></a>**UI-0376** The text of a text field is 18 CSS px, so that iOS Safari does not zoom. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0377"></a>**UI-0377** A field for a code, for example a one-time password, has wide letter spacing. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0378"></a>**UI-0378** A button is full width and the height of a control, with the medium radius. ([the control size](layout.md#ui-0093); [Judgement: the radius of the consumer pages](#ui-0394))
- <a id="ui-0379"></a>**UI-0379** The primary action has the primary colour. A secondary action has the strong border. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0380"></a>**UI-0380** A quiet action looks like a link, in the link colour. For example, "Cancel" is a quiet action. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0381"></a>**UI-0381** A choice card is a native checkbox and its label in a card with the strong border and the medium radius, one card for each choice. For example, each account of an account selection. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]; [Judgement: the radius of the consumer pages](#ui-0394))
- <a id="ui-0382"></a>**UI-0382** A selected choice card has a border of 2 CSS px in the heading colour. ([Judgement: the colour roles of the consumer pages](#ui-0393))

## Content parts

- <a id="ui-0383"></a>**UI-0383** A list of groups shows each group as a heading and a list of its items, with a divider between groups. For example, the data clusters and their permissions. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0384"></a>**UI-0384** A note is below the card, in the information colours, with a lock icon. For example, a statement about passwords is a note. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0385"></a>**UI-0385** The footer has small muted text, in the centre. ([Judgement: one design system (ADR 0016)][adr-0016])

## Motion and focus

- <a id="ui-0386"></a>**UI-0386** Hover and press have a transition of 100 ms, with `cubic-bezier(0.2, 0, 0, 1)`. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0387"></a>**UI-0387** A pressed button moves 1 CSS px. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0388"></a>**UI-0388** A consumer page has no page transitions. When the user asks for reduced motion, nothing on the page moves. ([WCAG 2.2, SC 2.3.3][wcag-2-3-3])
- <a id="ui-0389"></a>**UI-0389** The focus indicator of a consumer page is an outline of 3 CSS px in the focus colour, with an offset of 2 CSS px. ([WCAG 2.2, SC 2.4.7][wcag-2-4-7]; [Judgement: the focus indicator](components.md#ui-0278))

## Stylesheet

- <a id="ui-0390"></a>**UI-0390** The consumer pages stylesheet uses only the semantic custom properties of the tokens. ([Judgement: the stylesheet of the consumer pages](#ui-0395))

## Judgements

### <a id="ui-0393"></a>UI-0393 The colour roles of the consumer pages

- Kind: choice
- Basis: Step 3.3. No source gives the colour roles of a consumer page.
- Reason: The preset brands use a different colour for the header, the band below it, the headings and the links than for the primary action. The roles for the heading, the link, the brand, the hero and the accent let a brand give each colour. The accent is decoration only, so it has no contrast pair.

### <a id="ui-0394"></a>UI-0394 The radius of the consumer pages

- Kind: choice
- Basis: Step 3.3. No source gives the radius of each part of a consumer page.
- Reason: The radius scale of a brand gives each shape, for example a pill button with a medium radius of 9999 px. A component token for the radius is one more token in each brand directory.

### <a id="ui-0395"></a>UI-0395 The stylesheet of the consumer pages

- Kind: choice
- Basis: Step 3.3. The consumer pages are server pages with no script ([ADR 0014][adr-0014]).
- Reason: A Tailwind CSS stylesheet needs a scan of the page templates, so the web build then depends on the source of a service. A stylesheet with only the semantic custom properties takes each brand, and the token check also applies to it.

[adr-0014]: ../../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
[wcag-2-3-3]: https://www.w3.org/TR/WCAG22/#animation-from-interactions
[wcag-2-4-7]: https://www.w3.org/TR/WCAG22/#focus-visible
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
