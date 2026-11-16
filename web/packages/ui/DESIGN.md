# Design system: design

This document gives the specification and the requirements for the design system of the lab web UIs: the design tokens, the components and the rules for their use. It does not describe the implementation.

## Purpose

All lab web UIs use one design system, so that they look and act the same and a change to a rule has one place. The UIs are operator consoles (desktop) and consumer UIs (mobile). Each mock data holder and each data recipient is a different brand, so the design system gives one set of components for more than one brand. The decisions are in [ADR 0016][adr-0016].

## Sources

The requirements follow the process in [README: Sources for requirements][readme-sources]. The CDS is the first source for the consumer UIs. For the parts that the CDS does not cover, the sources are WCAG 2.2, the Australian Government Style Manual and the platform guidelines that each requirement gives. "Judgement" marks a requirement that no source states, and gives the reason.

Background research is in [Web front ends][research].

## Scope

In scope:

- The design tokens and the brands.
- The components that more than one UI uses, and their states.
- The layout rules for consoles and for consumer UIs.
- The accessibility level and the automated checks.
- The content rules: language, dates, times and amounts.

Out of scope for now:

- A dark theme. The tokens permit it later ([ADR 0016][adr-0016]).
- The CX rules of each CDR flow (consent, authorisation, withdrawal). The `DESIGN.md` of the service that has the flow gives them. This document gives the components that those flows use.
- A component catalogue ([ADR 0016][adr-0016]).

## Users

| User | Device | Density | Accessibility level |
|---|---|---|---|
| Operator console | Desktop | Compact | Console |
| Consumer UI | Mobile first | Comfortable | Consumer |

## Tokens

- The tokens are in the package [`web/packages/tokens`][tokens-package], in the DTCG 2025.10 format ([DTCG Format][dtcg-format], [DTCG Resolver][dtcg-resolver]). The build gives CSS custom properties, which the Tailwind theme uses. ([ADR 0016][adr-0016])
- The tokens have three tiers:
  - Primitive: raw values, for example colour ramps and font families. A component never uses a primitive token or a raw colour value. A check in CI finds each use in the apps and the shared packages.
  - Semantic: roles and scales that components use, for example the colour roles, the font roles and the spacing scale. Each brand gives a value for each role.
  - Component: only for a value that one component needs and that no role gives.
- The semantic colour roles are: page background, surface, raised surface, modal backdrop, default text, muted text, text on a primary action, heading, link, brand (the header of a consumer page) and text on the brand, hero (the band below the header) and text on the hero, accent (decoration only), border, strong border, focus ring, primary action (and its hover and pressed states), secondary action, and the status roles success, warning, danger and information, each with a background, a border and a text value. A new role needs a change to this list.
- Colour values are in OKLCH. Judgement: lightness steps in OKLCH are even, so a ramp for a new brand needs less adjustment ([CSS Color 4: OKLCH][css-oklch]).
- Each pair of text and background roles has a contrast of 4.5:1 or more, and each focus ring and control boundary has 3:1 or more against its background, in each brand. The console of a participant uses some roles of its brand and the lab brand for the rest (see Brands), so the test also checks each pair of a brand role and a lab role that the console shell shows, for each brand. The primary action colour is not a text colour and not a control boundary: a link uses the link role, and a primary button has a text label (see Decisions). ([WCAG 2.2, SC 1.4.3][wcag-1-4-3], [SC 1.4.11][wcag-1-4-11]) A test checks each pair for each brand.
- The strong border is the control boundary. The border role and the status borders are decorative: they are not necessary to identify a component or its state, so SC 1.4.11 does not apply to them, and they need no minimum contrast. ([WCAG 2.2, SC 1.4.11][wcag-1-4-11])
- The non-colour tokens are: a spacing scale on a 4 px base, a radius scale, three elevation levels, a type scale, line heights, font weights and motion durations.
- Two densities, compact and comfortable, change only spacing, control heights and table row heights. The consoles use compact. A consumer UI uses comfortable, with controls 48 CSS px high.

## Brands

- A brand is a set of values for the brand modifier of the tokens. It can change the colour roles, the body font, the heading font, the radius and the logo. It cannot change the spacing scale, the type scale or the component structure. Judgement: the brands must look different, but a component must work the same in each brand. An app can arrange the components differently for each brand, as the online bank site of the data holder does with its layouts ([data holder DESIGN.md: Online bank site][dataholder-site]).
- The lab brand is neutral. The consoles use it for each role that the brand of a participant does not give.
- The console of a participant uses these roles of the brand of that participant: brand and text on the brand, hero and text on the hero, accent, heading, link, primary action (and its hover and pressed states) and text on a primary action. The other roles, the body font, the heading font and the radius stay in the lab brand. (See Decisions)
- The Register console has no brand of a participant. It uses the lab brand for each role.
- Each mock participant that has a consumer UI has its own brand. A new participant UI adds a brand in the same change.

## Typography

- The UIs use self-hosted fonts only. No font comes from another origin. ([OWASP ASVS 5.0, V3.6.1][asvs-v3])
- The fonts are a fixed set, which [`fonts.json`][fonts-list] gives. A brand uses a font of the set. A new font has a licence that permits self-hosting, and its licence text is next to its files.
- The lab brand uses Inter, with JetBrains Mono for identifiers, thumbprints and keys. Amounts and numbers in tables use tabular figures. Judgement: digits with one width align in columns.
- Body text in a consumer UI is 16 CSS px or more. Judgement: iOS Safari zooms into a form field with smaller text.
- Text can be resized to 200 % with no loss of content or function. ([WCAG 2.2, SC 1.4.4][wcag-1-4-4])
- Each heading level from h1 to h4 has a smaller font size than the level above it. h4, h5 and h6 have the same size. Judgement: the user sees the structure of a page that has nested sections.

## Layout

- A console is usable at a viewport width of 1280 CSS px and more. A console has a side navigation, a header, and a main area that starts with the page header.
- The shell of a console shows its brand in three places only: the header, the side navigation and the page header. Tables, forms, cards and dialogs have the same layout in each brand. (See Decisions)
  - The header is on the brand colour, with a bottom rule of 3 CSS px in the accent colour.
  - The side navigation is on the hero colour, with the text on the hero. The current item has a lighter step of the hero colour, a semibold text and a bar of 3 CSS px in the text on the hero, so that the current item does not depend on the accent, which is decoration only. The focus ring in the side navigation is in the text on the hero.
  - The page header is a band with the full width of the main area, on 10 % of the accent in the surface colour, with a bottom border. The title is in the heading colour.
  - The main area is on 3 % of the hero in the page colour. Cards stay on the surface colour.
- A consumer UI is designed for a width of 360 CSS px first, and reflows with no horizontal scroll at 320 CSS px. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10]; judgement for 360 px: a common phone width)
- A consumer UI uses one column, and the primary action of a step is at the end of the content. A sticky action bar does not hide a focused element. ([WCAG 2.2, SC 2.4.11][wcag-2-4-11])
- A consumer UI respects the safe areas of the device (`env(safe-area-inset-*)`) and uses the small viewport height (`svh`) for full-height layouts. ([MDN: env()][mdn-env], [MDN: length][mdn-length])
- The UIs never prevent zoom. ([WCAG 2.2, SC 1.4.4][wcag-1-4-4])
- In a consumer UI, each control is 44 by 44 CSS px or more. In a console, each control is 24 by 24 CSS px or more. ([WCAG 2.2, SC 2.5.8][wcag-2-5-8], [Apple: Accessibility][apple-hig]; judgement: 44 px on touch screens, the WCAG minimum on desktop)

## Components

The design system has each component that more than one UI uses. A UI does not make its own version of a component that the design system has.

First set, for the consoles:

| Component | Requirements |
|---|---|
| App shell | Side navigation with the current page marked and an optional icon for each item, header, main area, with the colours of the shell of a console (see Layout). The header has the name as text, or a logo with the name as its text alternative and an optional label on the brand colour. A "skip to content" link. Assistive technology does not announce the icons. |
| Page header | A band with the title, an optional description and actions (see Layout). When the UI asks for it, for example after a form that sent the user to the page, the focus moves to the title when the page shows. ([WCAG 2.2, SC 2.4.3][wcag-2-4-3]) |
| Button | Primary, secondary, danger and quiet variants. A disabled or busy state that the screen reader announces. An optional icon before the text. On a busy button, the busy indicator replaces the icon. An enabled button, also the button of a collapsible card, shows the pointer cursor, and the button of a card has a hover background. |
| Link | Visibly different from text, also without colour. ([WCAG 2.2, SC 1.4.1][wcag-1-4-1]) An optional icon before the text. Only the text is underlined. |
| Data table | Column headers, sort with the sort state announced, row selection only when an action needs it, an empty state, a loading state and an error state. Pagination below the table. |
| Status badge | The status text is always visible. Colour only supports the text. ([WCAG 2.2, SC 1.4.1][wcag-1-4-1]) |
| Filter bar | An optional search field with a clear button, and filters for a list. Each filter has a visible label. |
| Section | A region of a page with a heading, optional links directly after the heading, and an optional description. The links go to a page of the record of the section, for example a form. The heading level follows the structure of the page, and sections can nest. A section of one record is a card: a border, and a title bar with the heading and the links. A card in a card has a different background. A section of a group of records has no card. A card can collapse to its title bar: the heading has a button with the expanded state, and the title bar also shows the status of the record and its links. The collapsed content is `hidden="until-found"`, so that a find in the page opens the card. A section of a group can have "Show all" and "Hide all" buttons, which open or close each card in it, also the cards in a nested group. Their accessible names say what they open, for example "Show all software products of Mock Budget". ([WAI-ARIA APG: Disclosure][apg-disclosure], [HTML: the hidden attribute][html-hidden]) ([WCAG 2.2, SC 1.3.1][wcag-1-3-1], [SC 2.4.6][wcag-2-4-6]; [GOV.UK Design System: Summary list, summary cards][govuk-summary-card]) |
| Action row | The buttons that act on one record, in one row after the details of the record and before its child records. |
| Description list | Key-value pairs for the detail of a record, as an HTML description list (`dl`), so that each value is connected to its term ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]). A value can be a list of values. An empty value shows the text "Not set". An empty list of values can give a different text, for example "None", when no values is a normal state. A long value, for example a URI or a thumbprint, wraps, so that the page has no horizontal scroll. |
| Timeline | The events of a record as an ordered list (`ol`), newest first. Each event has its time in a `time` element with the ISO 8601 UTC value in `datetime`, a title and optional details. Pagination below the list, and the empty, loading and error states. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1], [HTML: the time element][html-time]) |
| Dialog | A modal dialog with a title. Focus moves into the dialog and back to the control that opened it. When that control does not exist after the action, focus moves to an element that the UI gives, for example the heading of the record. ([WAI-ARIA APG: Dialog][apg-dialog]) |
| Confirmation dialog | An alert dialog that names the action and its effects, as a list when there is more than one effect. Its confirm button names the action, and a cancel button closes it with no change. The first focus is on the cancel button. When the action cannot be reversed, the dialog says so. While the request runs, the confirm button is busy, a status message says so, and the dialog stays open. When the action fails, the dialog stays open and shows the error in an alert. ([WAI-ARIA APG: Alert Dialog][apg-alertdialog], [WAI-ARIA APG: Dialog][apg-dialog]) |
| Form field | A visible label, optional help text, and an error message that is connected to the field. An error from the server for the value of a field shows as the error message of that field. A field error from the server has a JSON Pointer to a value of the request body: the error of an item of a list, or of a value in a group, shows at the field of the list or at the group. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1], [SC 3.3.2][wcag-3-3-2]) |
| Text field, text area, select, checkbox, radio group | In the form field. A select is a native `select` element. A select can start with no selected value: its first option is empty and has a text that asks the user to choose, so that the UI does not choose a value for the user. A text area for a list of values has help text that says to give one value on each line. |
| Date-time field | A native date-time control in the form field, in the time zone of the browser, with seconds. The UI sends the value as an instant in UTC. The help text says that the time is in the time zone of the browser. ([ADR 0012][adr-0012]) |
| Checkbox group | A group of checkboxes in a `fieldset` with a `legend`, which is the label of the group. The group has one help text and one error message, connected to the group. A link to the group goes to its first checkbox. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1], [SC 3.3.1][wcag-3-3-1]; [GOV.UK Design System: Checkboxes][govuk-checkboxes]) |
| Fieldset | A group of related fields in a `fieldset`, with a `legend` that names the group, for example the endpoints of a brand. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]) |
| Form | The fields of one record on a page, with a submit button that names the action and a "Cancel" link. The UI gives the fields and the request. While the request runs, the submit button is busy, a status message says so, and a second submit sends nothing. After a refusal, the form keeps each value, shows each field error at its field, and shows the error summary. The form does not check the values itself: the server checks them. A label of an optional field ends with "(optional)", and a required field has no mark. ([GOV.UK Design System: Question pages][govuk-question-pages]) |
| Error summary | The alert of a form, at the top of the form, with the heading "There is a problem". It shows the error of the form, if there is one, and a list with a link to each field that has an error, with the text of the error message of the field. The focus moves to it when it appears. A link moves the focus to its field. ([GOV.UK Design System: Error summary][govuk-error-summary], [WCAG 2.2, SC 3.3.1][wcag-3-3-1]) |
| File input | A control with a visible label that opens the file chooser of the browser, and shows the name of the selected file. The UI gets the text of the file. For a certificate, the text goes into a text area, where the user can see and change it. |
| Alert | An error on the page or in a dialog, announced when it appears. It stays until the cause changes. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]) |
| Toast | The result of a successful action, announced as a status message. It has a close button. It closes by itself after 5 seconds or more, but not while it has the pointer or the focus. A toast is never the only place of its information, and an error never shows only in a toast. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3], [SC 2.2.1][wcag-2-2-1]) |
| Empty, loading and error states | One pattern for each, used by each view. The error state shows the `title` and `detail` of a problem details response. |
| Copy button | Copies an identifier, and announces the result. |
| Identifier | An identifier in full, in the monospace font, with a copy button (see Content). It can show the identifier in a usual format, for example a BSB in two groups of three digits, and the copy button then copies the identifier with no format. It wraps in a description list. A table column can keep it on one line. |

The consumer pages of the data holder have the components of the next section. The first React consumer UI adds them to this package as components, with the same look.

## Consumer pages

This is the design guide of the consumer UIs. The authentication and consent pages of the data holder follow it now, and a later consumer UI follows it too. The rules for the content of each page are in the `DESIGN.md` of the service.

| Part | Rules |
|---|---|
| Header | The logo of the brand on the brand colour, at the start, 32 CSS px high, with the brand name as its text alternative. A "Secure" label with a lock icon at the end. The accent colour is a line of 4 CSS px at the top. The logo of a brand is legible on its brand colour and on white, because the Register gives the same file to data recipients. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]) |
| Hero | A band in the hero colour below the header. A row in the centre has the initial of the data recipient, an arrow and the mark of the brand, which assistive technology does not announce. Below the row, one sentence in the centre names the data recipient and the bank. A page with no data recipient has no hero. |
| Card | One raised card for the task of the page, at most 440 CSS px wide, in the centre. The card overlaps the lower edge of the hero by 24 CSS px. It has the large radius and two shadow layers. |
| Step | "Step n of 3" in small upper-case text above the heading, and a thin bar that shows the progress. The text gives the step, and the bar is decoration. |
| Heading | One h1 for the task, 24 CSS px, bold, letter spacing -0.015 em, in the heading colour. Sections of the card use h2 and h3. |
| Text field | The label above the field. The field is the height of a control, with the strong border and the small radius, and its text is 18 CSS px, so that iOS Safari does not zoom. A one-time password is one field with `autocomplete="one-time-code"` and wide letter spacing, not one field for each digit. ([WCAG 2.2, SC 3.3.8][wcag-3-3-8]) |
| Buttons | Full width, the height of a control, the medium radius. The primary action has the primary colour. A secondary action has a strong border. A quiet action, for example "Cancel", looks like a link, in the link colour. |
| Account selection | A native checkbox and its label in a card with the strong border, one card for each account. A selected card has a border of 2 CSS px in the heading colour. |
| Data clusters | Each cluster as a heading and a list of its permissions, always visible, with a divider between clusters. Judgement: a permission that the consumer must open to read is easy to miss. |
| Password statement | Below the card, in the information colours, with a lock icon. |
| Footer | The text of the service, in small muted text, in the centre. It has no link. |
| Motion | 100 ms for hover and press, with `cubic-bezier(0.2, 0, 0, 1)`, and a move of 1 CSS px when a button is pressed. No page transitions. No movement when the user asks for reduced motion. ([WCAG 2.2, SC 2.3.3][wcag-2-3-3]) |
| Focus | An outline of 3 CSS px in the focus colour, with an offset of 2 CSS px. |

A brand changes the colours, the fonts, the radius and the logo. Each other part of the design is the same in each brand (see Brands).

Rules for all components:

- Each component works with a keyboard only, with a visible focus indicator. ([WCAG 2.2, SC 2.1.1][wcag-2-1-1], [SC 2.4.7][wcag-2-4-7])
- No component depends on a drag. ([WCAG 2.2, SC 2.5.7][wcag-2-5-7])
- No component uses only colour to give information. ([WCAG 2.2, SC 1.4.1][wcag-1-4-1])
- A component with motion uses no motion when the user asks for reduced motion (`prefers-reduced-motion`). ([WCAG 2.2, SC 2.3.3][wcag-2-3-3]; judgement: the AAA criterion costs nothing with tokens for motion)
- Each input and each control that has a border uses the strong border, so that its boundary has 3:1 or more ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]). The border role is only for decorative lines, for example dividers.
- A component renders no `<style>` or `<script>` element, sets no `style` attribute from a string, and assigns no HTML string, so that the UIs can use a strict CSP with Trusted Types. A style that a script sets through the CSSOM (for example the React `style` prop) is permitted, because the CSP does not control it ([MDN: style-src][mdn-style-src]). Judgement: [the CSP of each web UI](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007) requires this.
- Each component works with no CSP or Trusted Types violation under [the CSP of each web UI](../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007). A test page with each component in each state checks this.

## Accessibility

- A consumer UI targets WCAG 2.2 Level AA. ([ADR 0016][adr-0016]; [AHRC guidelines][ahrc])
- An operator console has no target beyond what the components give, but the automated checks must pass. ([ADR 0016][adr-0016])
- The automated checks run axe-core with the tags `wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa` and `wcag22aa`. A violation fails the test. ([axe-core API][axe-api]) The checks run in the component tests and in the browser tests of each UI. For the consumer pages of the data holder, they run on each page and state, with each preset brand, at a viewport width of 320 CSS px.
- An exception to a check needs a comment with the reason, at the place of the exception.
- There are no manual screen-reader audits. ([ADR 0016][adr-0016])
- When a page has the same link for more than one record, for example "Change brand", the accessible name of each link also names its record, in text that only assistive technology gets. ([WCAG 2.2, SC 2.4.4][wcag-2-4-4]: a heading is not a context of a link that the criterion accepts)

## Content

- The language is Australian English. Each page has `lang="en-AU"`. ([WCAG 2.2, SC 3.1.1][wcag-3-1-1])
- Text in a consumer UI is plain language, at an Australian year 7 reading level. ([Style Manual: Literacy and access][style-literacy])
- Dates in text are "22 October 2026". Dates in tables can be "22 Oct 2026". A date never has a two-digit year. ([Style Manual: Dates and time][style-dates])
- Times are "2:05 pm". A time shows in the time zone of the browser, with its offset from UTC at that instant, for example "2:05 pm UTC+8" or "1:59 am UTC+10". A time in UTC has the label "UTC". ([Style Manual: Dates and time][style-dates], [ADR 0012][adr-0012]) No server, cookie or account keeps the time zone.
- A screen reader gets the offset in words, for example "UTC plus 8 hours". (See Decisions.)
- A time in a timeline has seconds, for example "2:05:07 pm UTC+8". Judgement: more than one event can occur in one minute.
- Amounts are "$1,234.50", with 2 decimal digits. A negative amount has a minus sign in a consumer UI. ([Style Manual: Currency][style-currency])
- The UIs format dates, times and amounts with `Intl` and the locale `en-AU`, at display only. A UI does not calculate with a formatted value. ([ECMA-402][ecma-402]; [ADR 0012][adr-0012])
- An identifier is shown in full, in the monospace font, with a copy button. Judgement: an operator copies identifiers into other tools.

## Browser support

- The UIs support the current and previous major versions of Chrome, Edge, Firefox and Safari, and Safari on iOS. Judgement: the lab has no users on old browsers, and each feature that the UIs need is in these versions.
- The UIs use no feature that is not in Safari, or they include a polyfill for it. Example: `Temporal` is not in Safari ([MDN: Temporal][mdn-temporal]).

## Non-functional requirements

- The design system has no dependency that runs an install script ([ADR 0015][adr-0015]).
- The tests of a component check its behaviour by role and by label, its states, and the automated accessibility checks, in a real browser ([ADR 0015][adr-0015]).
- The tests run in a fixed browser time zone, so that a time shows the same on each computer. The tests of the time components also run in UTC, in a time zone with daylight saving and in a time zone with a negative offset.
- A change to a token or a component shows in each UI at its next build. A UI does not copy tokens or components.

## Decisions

| Question | Decision | Reason |
|---|---|---|
| Colour space of the tokens | OKLCH | Judgement: even lightness steps make brand ramps easier. All supported browsers support it ([MDN: oklch()][mdn-oklch]). |
| Place of the links and the buttons of a record | Links that go to a page of the record are directly after its heading. Buttons that act on the record are in an action row after its details. A link never looks like a button. | Judgement: the owner asked for it. Links and buttons in one row looked the same, and a text link looked less important than the buttons next to it. In two places, the user sees which controls go to a page and which change the record, and reads the details, for example the status, before an action. A link with the look of a button announces "link" but looks like a button, and does not activate with the Space key. |
| Separation of the sections of a page | A card for each record, with at most two levels of cards. A section of a group of records is a heading with more space above it, with no card. | Judgement: the owner asked for more separation. Sections with only headings and equal space did not show where a record starts and ends. A border around a record groups its content ([Nielsen Norman Group: Common region][nng-common-region]). More than two levels of cards add borders that do not give more information. |
| Collapse of cards | A card of a child record, for example a brand on the page of its data recipient, starts collapsed when the page shows the records of a participant. The card of the record of the page never collapses. A form that sends the user back to the page opens the card of the record that it changed or added, and the cards around it. The open state is not kept after a reload. | Judgement: the owner asked for it. A participant with many brands and software products gives a long page, and the operator usually wants one record. The status in the title bar keeps the most important fact visible. The card of the changed record opens so that the operator sees the change. The disclosure pattern puts the button in the heading, so that the links stay outside the button: a link in a `summary` is a control in a control. Show all and Hide all are buttons at the heading of a group, because they change the view and do not act on a record. The signing keys do not collapse: each key has two fields, and its key ID is the information that the operator looks for. |
| Cursor of a button | The pointer cursor | Judgement: the owner asked for it. The browser default for a button is the arrow cursor, and the user could not see where to click to open a card. A link already shows the pointer cursor. |
| Icons of links and buttons | A link or a button of an action has an icon before its text. Each type of action has one icon in all UIs, which the UI gives, and assistive technology does not announce it. The text stays. | Judgement: the owner asked for it. An icon helps the user find a control again, but an icon for an action such as a revocation has no meaning that all users know, so the text gives the meaning. A decorative image needs no text alternative ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]). |
| Control size | 44 by 44 CSS px in consumer UIs, 24 by 24 CSS px in consoles | WCAG 2.2 SC 2.5.8 gives 24 px as the AA minimum. Apple gives 44 pt as the default size for touch ([Apple: Accessibility][apple-hig]). |
| Smallest body text in a consumer UI | 16 CSS px | Judgement: iOS Safari zooms into form fields with smaller text, which moves the layout. |
| Accessibility checks | axe-core, WCAG 2.0, 2.1 and 2.2 A and AA tags, a violation fails | axe-core needs each tag ([axe-core API][axe-api]). The two-level bar of [ADR 0016][adr-0016] still gives automated checks for the consoles. |
| Browser support | Current and previous major versions of the evergreen browsers | Judgement: see Browser support. |
| What a brand can change | Colours, body font, heading font, radius and logo | Judgement: see Brands. |
| Amounts in consoles | Accounting format is not used | Judgement: one format for all UIs is simpler, and the operator consoles show few amounts. |
| Contrast of the border role | No minimum. Only control boundaries (the strong border) and the focus ring need 3:1. | WCAG 2.2 SC 1.4.11 does not apply to a border that is not necessary to identify a component or its state. A 3:1 divider on each table row makes a dense console difficult to read. |
| Monospace font of the lab brand | JetBrains Mono (SIL Open Font License 1.1) | Judgement: its 0 and O, and its 1, l and I, are clearly different, which an operator needs to read an identifier or a thumbprint. The licence permits self-hosting. |
| Source of the shadcn/ui components | Copied by hand from the shadcn/ui registry, and changed to the semantic tokens. The repository has no `components.json`, and the shadcn CLI does not run. A comment in each file gives the registry item, style and version. | Judgement: the registry source uses the shadcn token names and `dark:` variants, which the tokens do not have, so each file needs changes. The CLI runs outside the controls of the workspace lock file ([ADR 0015][adr-0015]). An upstream fix thus needs a manual compare with the registry item ([shadcn/ui: Registry][shadcn-registry]). |
| Focus indicator | A 2 px outline in the focus ring role, with an offset. 3 px on the consumer pages. | Judgement: forced-colours mode removes a box shadow, but it keeps an outline ([MDN: forced-colors][mdn-forced-colors]). A consumer uses a touch screen or a small screen more often, and the wider outline is easier to see. |
| Brand in a console | The shell roles of the brand (brand and the text on it, hero and the text on it, accent and heading), and its link and primary action roles. The lab brand for the rest, and the lab font and radius. | Judgement: the operator asked for consoles that show more of the brand and are less white and flat, but stay dense staff tools. The brand colour is white for some brands, so the side navigation uses the hero colour, which is dark in each preset. A different font changes the width of each table column, so the font stays the same. |
| Mixed colours of the shell | Component tokens of the design system, which mix two roles: the hover and the current item of the side navigation, the page header band and the page of the main area | Judgement: a brand then gets them with no new token, and the contrast test checks each one. The mix ratios keep 4.5:1 for the link of each preset brand on the band and on the page. |
| Focus ring in the side navigation | The text on the hero, not the focus role | The focus role has less than 3:1 on the hero colour of some preset brands ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]). The text on the hero has 4.5:1 or more on the hero in each brand. |
| Contrast pairs of the primary action colour | The primary action colour has no text pair and no 3:1 pair against the backgrounds. The text on it has 4.5:1. A link uses the link role, which has 4.5:1 against each background. | A button with a text label needs no boundary of 3:1, because the text identifies the component ([WCAG 2.2, Understanding SC 1.4.11][understanding-1-4-11]). Judgement: a brand with a light primary colour, for example yellow, can then use it for its buttons, as the bank does. The pairs of the focus ring and the strong border stay. |
| Colour roles of the consumer pages | Heading, link, brand and text on the brand, hero and text on the hero, and accent | Judgement: the brands of the banks use a different colour for the header, the band below it, the headings and the links than for the primary action. The accent is decoration only, so it has no contrast pair. |
| Radius of the consumer pages | The card uses the large radius, a button and an account card the medium radius, and a text field the small radius. No component token for the radius. | Judgement: the scale of a brand gives each shape, for example a pill button with a medium radius of 9999 px. A component token for the radius is one more token in each brand directory. |
| Browsers in the component tests | Chromium and Firefox | Judgement: WebKit needs system libraries on each CI run ([Playwright: Browsers][playwright-browsers]), which costs CI time. Safari is still a supported browser (see Browser support), but CI does not test it. |
| When the search of the filter bar applies | 300 ms after the last change of the text, or at once on the Enter key. A "Clear search" button shows when the field has text: it empties the field, applies the empty search at once and puts the focus in the field. The focus stays in the field while the list changes. A filter applies when its value changes. | Judgement: the operator asked for a search with no click. The delay gives one search for each pause in the typing, not one for each key. A change of the list is a change of content, not of context, and the status message of the pagination announces the result ([WCAG 2.2, SC 3.2.2][wcag-3-2-2], [SC 4.1.3][wcag-4-1-3]). The clear button replaces the cancel button of the browser, which only some browsers show. A filter is a native select, which needs no popover and has the keyboard support of the browser. |
| Pagination controls | Previous and Next buttons, and the text "Page X of Y" and the number of records, each as a status message. No list of page numbers and no page-size control. | Judgement: the consoles show short lists, and two buttons are the simplest control that works with a keyboard. The status messages announce each change of page and the result of each search or filter ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]). |
| Sort of the data table | A column sorts only when the server can sort it. A sort button sorts an unsorted column ascending and turns a sorted column the other way. The sorted column has `aria-sort`, and a status message also gives the sort. | The sort pattern of the [WAI-ARIA APG: Sortable Table][apg-sortable-table]. Judgement: screen readers do not always announce a change of `aria-sort`, and the list always has a sort from the server, so there is no unsorted state. |
| Empty value in a description list | The text "Not set". An empty list of values can give its own text, for example "None". | Judgement: the term stays in the list, so that the user sees that a field has no value, for example an optional SSA field. An empty value is silent for a screen-reader user. "Not set" says that a value is missing, which is not correct for a list that is empty by design, for example the accounts of a consent with no `bank:` scope. |
| Seconds in a timeline | Each time has seconds | Judgement: an operator action and its cascade, or two operator actions, can occur in one minute. |
| Format of a date with a time | The date in full, a comma, the time and the offset, for example "23 October 2026, 2:05:07 pm UTC+8" | Judgement: the Style Manual gives the date and the time, but not one form for the two together ([Style Manual: Dates and time][style-dates]). The UI puts the parts of `Intl` together itself, because the text between the parts changes from one ICU version to the next, for example "at" or a comma. |
| Time zone of a time | The time zone of the browser, shown as the offset from UTC at that instant ("UTC+8", "UTC+10:30", "UTC-3:30"), not as a time zone name or abbreviation. "UTC" when the offset is zero. | Judgement: an offset is not ambiguous and an operator can compare it directly with a time in UTC, for example in Grafana, the logs or an API response. The abbreviations of `Intl` for `en-AU` give "AWST" only in some places, and "GMT+8" in many others, for example Singapore, and some abbreviations have more than one meaning, for example "CST". A name such as "Australia/Sydney" does not show if daylight saving applies. The offset changes with daylight saving, so the two times of a repeated hour are different. The browser gives the offset, so the UI needs no time zone data. The offset is "UTC" and a sign, not "+08:00", because "UTC+8" is the usual form for a person, and the "UTC" label stays for the UTC time zone. |
| Offset for a screen reader | The visible offset is hidden from assistive technology, and visually hidden text gives it in words, for example "UTC minus 3 hours 30 minutes". The hidden text cannot be selected, so a copy of the time has the visible text only. | Judgement: a screen reader can read "-" as "dash", or not at all at its default punctuation level, which gives the wrong offset. Words give one reading in each screen reader. |
| A time value that is not valid | The UI shows the value as text, with no `time` element | Judgement: the API contract gives each time as a `date-time`, so only a defect gives a different value. One raw value is better for the user than a page that stops. |
| Pattern of the confirmation dialog | The alert dialog pattern, with the first focus on the cancel button, also when the action can be reversed | The APG gives the alert dialog for a message that needs a response, and advises the focus on the least destructive action for a step that is not easily reversible ([WAI-ARIA APG: Dialog][apg-dialog]). Judgement: the operator confirms with a deliberate key press or click. One rule for all confirmations is simpler, and a reversible action can also change other records, for example a suspension. |
| Busy state of a confirmation | The busy confirm button, and also a status message, for example "Suspend in progress" | Judgement: a screen reader does not always announce a change to the name of the focused button. A status message is announced ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]). |
| An action that fails with no problem details | The dialog stays open, and an alert says "The action did not complete" and "Refresh the page to see the current state, then try again." | Judgement: when the request fails, for example with a network error, the UI does not know if the change occurred. The UI never sends the action again by itself ([Register DESIGN.md: Operator console][register-console]). |
| Variant of the confirm button | Danger when the action cannot be reversed, primary when it can. The UI cannot change it. | Judgement: one rule in the design system, so that each confirmation looks the same for the same type of action. |
| Colour of the modal backdrop | A semantic role, the modal backdrop, which is a dark colour with transparency | Judgement: a dimmed page shows that the dialog is modal. The role is not text and not a control boundary, so it has no contrast minimum ([WCAG 2.2, SC 1.4.11][wcag-1-4-11]). A brand can change it. |
| Errors and toasts | An error shows in an alert at the place of the action. Only a success shows in a toast. | Judgement: a toast closes by itself, and an error must stay until the user reads it and acts ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]). |
| Time of a toast | 5 seconds or more, with a pause while it has the pointer or the focus | Judgement: the result is also in the view, so the toast gives no information that the user can lose. 5 seconds is the Base UI default ([Base UI: Toast][baseui-toast]). The close button is hidden from assistive technology until the toast has the pointer or the focus, as Base UI does, so that the announcement does not include it. F6 moves the focus to the toasts. |
| Select in a form | A native `select` element, not the Base UI Select | The Base UI Select and ScrollArea add a `<style>` element, which the CSP of the consoles refuses ([Base UI: CSP Provider][baseui-csp]). A native select needs no popover and has the keyboard support of the browser. Judgement: the same reason as for the select of the filter bar. |
| Input of a time | A native `input` element of type `datetime-local` with a step of 1 second. The UI converts its value, which has no offset, to UTC with the offset of the browser at that instant. | Judgement: the browser gives the control, its keyboard support and a value that is a valid local time or empty, in each supported browser ([HTML: Local Date and Time state][html-datetime-local]). A text field needs a format that the user must learn. The offset at the instant, and not the offset of today, gives the correct time on the other side of a change of daylight saving. The browser gives the text of the control in the language of the browser, as for the file input. |
| Certificate file | The UI puts the text of the file in a text area, and sends the text | Judgement: the user sees the value before the action, the request has one value, and the server is the only place that reads the certificate. |
| File input | A native `input` element of type `file`, with a style on its button. The browser gives the text of the button and of the file name, in the language of the browser. | Judgement: the browser gives the file chooser, its keyboard support and the name of the selected file, in each supported browser ([HTML: File Upload state][html-file]). A control that the UI makes itself needs a hidden input and a second focus target. The text in the language of the browser is the cost: the UI cannot make it Australian English. |
| Place of the error message of a field | Below the label and the help text, and above the control | Judgement: a user with a screen magnifier sees the error next to the label, also above a tall text area. The [GOV.UK Design System: Error message][govuk-error] uses the same order. The control is described by the help text and the error, in that order. |
| Errors of a form | An error summary at the top of the form, with the focus on it, and an error message at each field | Judgement: a form of a record can be longer than one screen, so an error message at a field far down the page is not seen. The [GOV.UK Design System: Error summary][govuk-error-summary] shows the summary for each error and moves the focus to it. The summary is also the alert of the form, so that a form has one place for each error. |
| A form that fails with no problem details | The error summary says "The change did not complete" and "Refresh the page to see the current state, then try again." | Judgement: the same rule as for an action that fails with no problem details. The UI does not know if the change occurred. |
| List of values that the user writes | A text area with one value on each line, not a list of text fields | Judgement: one control has one label and one error message, the user can paste a list, and the keyboard and the screen reader read it line by line. A list of text fields needs add and remove buttons and a focus rule for each. |
| A required field | No mark on a required field. "(optional)" at the end of the label of an optional field. | [GOV.UK Design System: Question pages][govuk-question-pages]. Judgement: most fields of the records are required, so the optional fields are the exception to mark. |
| Identifier as a component | One component for the identifier rule of Content | Judgement: each list and each detail view shows identifiers, and the design system gives one version, not one in each view. |
| Styles of the data holder consumer pages | A plain stylesheet of this package, which uses only the semantic custom properties of the tokens. The build puts the token and font stylesheets before it. The Go templates use its class names. No Tailwind CSS. | Judgement: the pages are Go pages with no script ([ADR 0014][adr-0014]). Tailwind needs a scan of the Go templates, so the web build then depends on the source of a service, and a class change in a template needs a web build. The token check of CI also applies to the plain stylesheet. |
| Accessibility checks of the data holder consumer pages | A Go test writes each page and state, and the brand stylesheet of each preset brand, to golden files. A browser test of this package runs the checks on them. | Judgement: a browser test of the real flow needs a pushed authorisation request with mutual TLS and a signed request object. The golden files are the output of the templates, and the Go test fails when they are not current. |

[adr-0012]: ../../../docs/adr/0012-utc-for-all-times.md
[adr-0014]: ../../../docs/adr/0014-web-uis-as-embedded-react-single-page-applications.md
[adr-0015]: ../../../docs/adr/0015-web-toolchain-and-npm-supply-chain-controls.md
[adr-0016]: ../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[tokens-package]: ../tokens/
[fonts-list]: ../tokens/fonts.json
[dataholder-site]: ../../../services/dataholder/design/online-bank-site.md
[readme-sources]: ../../../README.md#sources-for-requirements
[research]: ../../../docs/research/2026-10-22-web-front-ends.md
[register-console]: ../../../services/register/DESIGN.md#operator-console
[dtcg-format]: https://www.designtokens.org/TR/2025.10/format/
[dtcg-resolver]: https://www.designtokens.org/TR/2025.10/resolver/
[css-oklch]: https://www.w3.org/TR/css-color-4/#ok-lab
[mdn-style-src]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/style-src
[mdn-forced-colors]: https://developer.mozilla.org/en-US/docs/Web/CSS/@media/forced-colors
[playwright-browsers]: https://playwright.dev/docs/browsers#install-system-dependencies
[shadcn-registry]: https://ui.shadcn.com/docs/registry
[mdn-oklch]: https://developer.mozilla.org/en-US/docs/Web/CSS/color_value/oklch
[mdn-env]: https://developer.mozilla.org/en-US/docs/Web/CSS/env
[mdn-length]: https://developer.mozilla.org/en-US/docs/Web/CSS/length
[mdn-temporal]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Temporal
[asvs-v3]: https://github.com/OWASP/ASVS/blob/v5.0.0_release/5.0/en/0x12-V3-Web-Frontend-Security.md
[apple-hig]: https://developer.apple.com/design/human-interface-guidelines/accessibility
[apg-sortable-table]: https://www.w3.org/WAI/ARIA/apg/patterns/table/examples/sortable-table/
[apg-dialog]: https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/
[apg-alertdialog]: https://www.w3.org/WAI/ARIA/apg/patterns/alertdialog/
[baseui-toast]: https://base-ui.com/react/components/toast
[baseui-csp]: https://base-ui.com/react/utils/csp-provider
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[axe-api]: https://github.com/dequelabs/axe-core/blob/develop/doc/API.md
[ahrc]: https://humanrights.gov.au/our-work/disability-rights/publications/guidelines-equal-access-digital-goods-and-services
[ecma-402]: https://tc39.es/ecma402/
[html-datetime-local]: https://html.spec.whatwg.org/multipage/input.html#local-date-and-time-state-(type=datetime-local)
[html-file]: https://html.spec.whatwg.org/multipage/input.html#file-upload-state-(type=file)
[govuk-error]: https://design-system.service.gov.uk/components/error-message/
[html-time]: https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-time-element
[style-literacy]: https://www.stylemanual.gov.au/accessible-and-inclusive-content/literacy-and-access
[style-dates]: https://www.stylemanual.gov.au/grammar-punctuation-and-conventions/numbers-and-measurements/dates-and-time
[style-currency]: https://www.stylemanual.gov.au/grammar-punctuation-and-conventions/numbers-and-measurements/currency
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[nng-common-region]: https://www.nngroup.com/articles/common-region/
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-4-1]: https://www.w3.org/TR/WCAG22/#use-of-color
[wcag-2-4-4]: https://www.w3.org/TR/WCAG22/#link-purpose-in-context
[wcag-1-4-3]: https://www.w3.org/TR/WCAG22/#contrast-minimum
[wcag-1-4-4]: https://www.w3.org/TR/WCAG22/#resize-text
[wcag-1-4-10]: https://www.w3.org/TR/WCAG22/#reflow
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
[understanding-1-4-11]: https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html
[wcag-2-1-1]: https://www.w3.org/TR/WCAG22/#keyboard
[wcag-2-3-3]: https://www.w3.org/TR/WCAG22/#animation-from-interactions
[wcag-3-3-8]: https://www.w3.org/TR/WCAG22/#accessible-authentication-minimum
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-2-4-6]: https://www.w3.org/TR/WCAG22/#headings-and-labels
[wcag-2-4-7]: https://www.w3.org/TR/WCAG22/#focus-visible
[wcag-2-4-11]: https://www.w3.org/TR/WCAG22/#focus-not-obscured-minimum
[wcag-2-5-7]: https://www.w3.org/TR/WCAG22/#dragging-movements
[wcag-2-5-8]: https://www.w3.org/TR/WCAG22/#target-size-minimum
[wcag-3-1-1]: https://www.w3.org/TR/WCAG22/#language-of-page
[wcag-3-2-2]: https://www.w3.org/TR/WCAG22/#on-input
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
[govuk-checkboxes]: https://design-system.service.gov.uk/components/checkboxes/
[apg-disclosure]: https://www.w3.org/WAI/ARIA/apg/patterns/disclosure/
[html-hidden]: https://html.spec.whatwg.org/multipage/interaction.html#the-hidden-attribute
[govuk-summary-card]: https://design-system.service.gov.uk/components/summary-list/#summary-cards
[govuk-error-summary]: https://design-system.service.gov.uk/components/error-summary/
[govuk-question-pages]: https://design-system.service.gov.uk/patterns/question-pages/
