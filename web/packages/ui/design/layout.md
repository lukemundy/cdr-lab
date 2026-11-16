# Design system: Layout

This area gives the layout of a console and its shell, the layout of a consumer UI, the size of a control, and the browsers that the UIs support.

## Console layout

- <a id="ui-0070"></a>**UI-0070** A console is usable at a viewport width of 1280 CSS px and more. ([Judgement: a minimum console width of 1280 CSS px](#ui-0100))
- <a id="ui-0071"></a>**UI-0071** A console has a side navigation, a header and a main area. ([Judgement: the same side navigation, header and main area in each console](#ui-0107))
- <a id="ui-0072"></a>**UI-0072** The main area of a console starts with the page header. ([Judgement: the same side navigation, header and main area in each console](#ui-0107))
- <a id="ui-0073"></a>**UI-0073** The shell of a console shows its brand only in the header, the side navigation and the page header. ([Judgement: the participant brand in the shell, with the lab fonts and radius](tokens-and-brands.md#ui-0060))
- <a id="ui-0074"></a>**UI-0074** Tables, forms, cards and dialogs have the same layout in each brand. ([Judgement: a brand changes only colours, fonts, radius and logo](tokens-and-brands.md#ui-0059))
- <a id="ui-0075"></a>**UI-0075** The header of a console is on the brand colour, with a bottom rule of 3 CSS px in the accent colour. ([Judgement: the participant brand in the shell, with the lab fonts and radius](tokens-and-brands.md#ui-0060))
- <a id="ui-0076"></a>**UI-0076** The side navigation is on the hero colour, with the text on the hero. ([Judgement: the participant brand in the shell, with the lab fonts and radius](tokens-and-brands.md#ui-0060))
- <a id="ui-0077"></a>**UI-0077** The current item of the side navigation has a lighter step of the hero colour, a semibold text and a bar of 3 CSS px in the text on the hero. It does not depend on the accent. ([the accent is decoration only](tokens-and-brands.md#ui-0021); [Judgement: shell colours that mix two roles, with no new token](#ui-0102))
- <a id="ui-0078"></a>**UI-0078** The focus ring in the side navigation is in the text on the hero, not in the focus ring role. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]; [Judgement: the focus ring in the side navigation uses the text on the hero](#ui-0101))
- <a id="ui-0079"></a>**UI-0079** The page header is a band with the full width of the main area, with a bottom border. Its colour is 10 % of the accent in the surface colour. ([Judgement: shell colours that mix two roles, with no new token](#ui-0102))
- <a id="ui-0080"></a>**UI-0080** The title of the page header is in the heading colour. ([Judgement: the participant brand in the shell, with the lab fonts and radius](tokens-and-brands.md#ui-0060))
- <a id="ui-0081"></a>**UI-0081** The main area is on 3 % of the hero in the page colour. ([Judgement: shell colours that mix two roles, with no new token](#ui-0102))
- <a id="ui-0082"></a>**UI-0082** A card in the main area stays on the surface colour. ([Judgement: shell colours that mix two roles, with no new token](#ui-0102))
- <a id="ui-0083"></a>**UI-0083** Each mixed colour of the shell is a component token that mixes two roles. The mixed colours are the hover and the current item of the side navigation, the page header band and the page of the main area. ([Judgement: shell colours that mix two roles, with no new token](#ui-0102))

## Consumer UI layout

- <a id="ui-0085"></a>**UI-0085** A consumer UI is designed for a viewport width of 360 CSS px first. ([Judgement: a design width of 360 CSS px for a consumer UI](#ui-0103))
- <a id="ui-0086"></a>**UI-0086** A consumer UI reflows with no horizontal scroll at a viewport width of 320 CSS px. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10])
- <a id="ui-0087"></a>**UI-0087** A consumer UI has one column. ([Judgement: one column in a consumer UI](#ui-0104))
- <a id="ui-0088"></a>**UI-0088** The primary action of a step is at the end of the content. ([Judgement: one column in a consumer UI](#ui-0104))
- <a id="ui-0089"></a>**UI-0089** A sticky action bar does not hide a focused element. ([WCAG 2.2, SC 2.4.11][wcag-2-4-11])
- <a id="ui-0090"></a>**UI-0090** A consumer UI keeps its content in the safe areas of the device (`env(safe-area-inset-*)`). ([MDN: env()][mdn-env])
- <a id="ui-0091"></a>**UI-0091** A full-height layout of a consumer UI uses the small viewport height (`svh`). ([MDN: length][mdn-length])
- <a id="ui-0092"></a>**UI-0092** No UI prevents zoom. ([WCAG 2.2, SC 1.4.4][wcag-1-4-4])

## Control size

- <a id="ui-0093"></a>**UI-0093** In a consumer UI, each control is 44 by 44 CSS px or more. ([Apple: Accessibility][apple-hig]; [Judgement: controls of 44 CSS px in a consumer UI and 24 CSS px in a console](#ui-0105))
- <a id="ui-0094"></a>**UI-0094** In a console, each control is 24 by 24 CSS px or more. ([WCAG 2.2, SC 2.5.8][wcag-2-5-8]; [Judgement: controls of 44 CSS px in a consumer UI and 24 CSS px in a console](#ui-0105))

## Browser support

- <a id="ui-0096"></a>**UI-0096** The UIs support the current and the previous major version of each of these browsers. ([Judgement: the current and the previous major version of each main browser](#ui-0106))
  - Chrome
  - Edge
  - Firefox
  - Safari
  - Safari on iOS
- <a id="ui-0097"></a>**UI-0097** A UI uses a feature that Safari does not have only with a polyfill. For example, Safari does not have `Temporal`. ([MDN: Temporal][mdn-temporal]; [Judgement: the current and the previous major version of each main browser](#ui-0106))

## Judgements

### <a id="ui-0100"></a>UI-0100 A minimum console width of 1280 CSS px

- Kind: choice
- Basis: Step 3.3. [ADR 0016][adr-0016] says that the consoles are for desktop and data-dense.
- Reason: 1280 CSS px is a common width of a desktop screen.

### <a id="ui-0101"></a>UI-0101 The focus ring in the side navigation uses the text on the hero

- Kind: choice
- Basis: Step 3.2. The focus ring role has less than 3:1 on the hero colour of some preset brands ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]).
- Reason: The text on the hero has 4.5:1 or more on the hero in each brand. Thus the focus ring is visible in each brand.

### <a id="ui-0102"></a>UI-0102 Shell colours that mix two roles, with no new token

- Kind: choice
- Basis: Step 3.3. No source gives the colours of a console shell.
- Reason: A brand gets the mixed colours with no new token, and the contrast check includes each one. The mix ratios keep 4.5:1 for the link of each preset brand on the band and on the page.

### <a id="ui-0103"></a>UI-0103 A design width of 360 CSS px for a consumer UI

- Kind: choice
- Basis: Step 3.3. WCAG gives 320 CSS px for reflow ([WCAG 2.2, SC 1.4.10][wcag-1-4-10]), but no design width.
- Reason: 360 CSS px is a common phone width.

### <a id="ui-0104"></a>UI-0104 One column in a consumer UI

- Kind: choice
- Basis: Step 3.3. No source gives the columns of a consumer UI.
- Reason: A phone width has space for one column only. The consumer reads the content of a step before its primary action.

### <a id="ui-0105"></a>UI-0105 Controls of 44 CSS px in a consumer UI and 24 CSS px in a console

- Kind: choice
- Basis: Step 3.2. WCAG 2.2 gives 24 by 24 CSS px as the Level AA minimum ([WCAG 2.2, SC 2.5.8][wcag-2-5-8]). Apple gives 44 pt as the default size for touch ([Apple: Accessibility][apple-hig]).
- Reason: A consumer UI is used on a touch screen, so it uses 44 px. A console is used on desktop, so the WCAG minimum is sufficient.

### <a id="ui-0106"></a>UI-0106 The current and the previous major version of each main browser

- Kind: choice
- Basis: Step 3.3. No source gives the browsers of a lab UI.
- Reason: The lab has no users on old browsers. Each feature that the UIs need is in these versions.

### <a id="ui-0107"></a>UI-0107 The same side navigation, header and main area in each console

- Kind: choice
- Basis: Step 3.3. No source gives the structure of a console.
- Reason: Each console has the same places for its navigation, its brand and the title of a page. A lab operator who uses more than one console then finds each part in the same place.

[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[apple-hig]: https://developer.apple.com/design/human-interface-guidelines/accessibility
[mdn-env]: https://developer.mozilla.org/en-US/docs/Web/CSS/env
[mdn-length]: https://developer.mozilla.org/en-US/docs/Web/CSS/length
[mdn-temporal]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Temporal
[wcag-1-4-4]: https://www.w3.org/TR/WCAG22/#resize-text
[wcag-1-4-10]: https://www.w3.org/TR/WCAG22/#reflow
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
[wcag-2-4-11]: https://www.w3.org/TR/WCAG22/#focus-not-obscured-minimum
[wcag-2-5-8]: https://www.w3.org/TR/WCAG22/#target-size-minimum
