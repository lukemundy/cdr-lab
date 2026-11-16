# Design system: Components

This area gives the components that more than one UI uses, their states, and the rules for all components. The [forms and dialogs](forms-and-dialogs.md) give the components of a form and of a dialog. The [consumer pages](consumer-pages.md) give the parts of a consumer page.

## Component set

- <a id="ui-0200"></a>**UI-0200** The design system has each component that more than one UI uses. A UI does not make its own version of a component that the design system has. ([Judgement: one design system (ADR 0016)][adr-0016])

## Rules for all components

- <a id="ui-0201"></a>**UI-0201** Each component works with a keyboard only. ([WCAG 2.2, SC 2.1.1][wcag-2-1-1])
- <a id="ui-0202"></a>**UI-0202** Each component has a visible focus indicator. ([WCAG 2.2, SC 2.4.7][wcag-2-4-7])
- <a id="ui-0203"></a>**UI-0203** The focus indicator is an outline of 2 CSS px in the focus ring role, with an offset. ([WCAG 2.2, SC 2.4.7][wcag-2-4-7]; [Judgement: the focus indicator](#ui-0278))
- <a id="ui-0204"></a>**UI-0204** No component depends on a drag. ([WCAG 2.2, SC 2.5.7][wcag-2-5-7])
- <a id="ui-0205"></a>**UI-0205** No component uses only colour to give information. ([WCAG 2.2, SC 1.4.1][wcag-1-4-1])
- <a id="ui-0206"></a>**UI-0206** When the user asks for reduced motion (`prefers-reduced-motion`), a component with motion uses no motion. ([WCAG 2.2, SC 2.3.3][wcag-2-3-3]; [Judgement: no motion for reduced motion](#ui-0279))
- <a id="ui-0207"></a>**UI-0207** Each input, and each control that has a border, uses the strong border. ([the boundary of a control](tokens-and-brands.md#ui-0019); [WCAG 2.2, SC 1.4.11][wcag-1-4-11])
- <a id="ui-0208"></a>**UI-0208** A component renders no `<style>` element and no `<script>` element. ([the CSP of each web UI][cnsl-csp])
- <a id="ui-0209"></a>**UI-0209** A component sets no `style` attribute from a string, and assigns no HTML string. ([the CSP of each web UI][cnsl-csp])
- <a id="ui-0210"></a>**UI-0210** A component can set a style through the CSSOM, because the CSP does not control it. For example, the React `style` prop sets a style through the CSSOM. ([MDN: style-src][mdn-style-src])
- <a id="ui-0211"></a>**UI-0211** Each component works with no CSP violation and no Trusted Types violation under the CSP of each web UI. ([the CSP of each web UI][cnsl-csp])

## Shell and page header

- <a id="ui-0212"></a>**UI-0212** The app shell has a side navigation, a header and a main area, with [the colours of the shell of a console](layout.md#ui-0073). ([the structure of a console](layout.md#ui-0071))
- <a id="ui-0213"></a>**UI-0213** The side navigation marks the current page. ([Judgement: the structure of a console](layout.md#ui-0107))
- <a id="ui-0214"></a>**UI-0214** An item of the side navigation can have an icon. ([Judgement: the icons of links and buttons](#ui-0284))
- <a id="ui-0215"></a>**UI-0215** The header of the app shell has the name of the UI as text, or a logo with the name as its text alternative. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1])
- <a id="ui-0216"></a>**UI-0216** The header of the app shell can have a label on the brand colour. For example, "Operator console". ([Judgement: the structure of a console](layout.md#ui-0107))
- <a id="ui-0217"></a>**UI-0217** The app shell has a "skip to content" link. ([WCAG 2.2, SC 2.4.1][wcag-2-4-1])
- <a id="ui-0218"></a>**UI-0218** The page header has the title, an optional description and actions, in [the page header band](layout.md#ui-0079). ([Judgement: the structure of a console](layout.md#ui-0107))
- <a id="ui-0219"></a>**UI-0219** When the UI asks for it, the focus moves to the title of the page header when the page shows. For example, the focus moves to the title after a form that sent the user to the page. ([WCAG 2.2, SC 2.4.3][wcag-2-4-3])

## Buttons and links

- <a id="ui-0220"></a>**UI-0220** A button has one of these variants. ([Judgement: one design system (ADR 0016)][adr-0016])
  - primary
  - secondary
  - danger
  - quiet
- <a id="ui-0221"></a>**UI-0221** A button can be disabled or busy. The screen reader announces the disabled state and the busy state. ([WCAG 2.2, SC 4.1.2][wcag-4-1-2])
- <a id="ui-0222"></a>**UI-0222** A button can have an icon before its text. On a busy button, the busy indicator replaces the icon. ([Judgement: the icons of links and buttons](#ui-0284))
- <a id="ui-0223"></a>**UI-0223** An enabled button shows the pointer cursor. This includes the button of a collapsible card. ([Judgement: the cursor of a button](#ui-0283))
- <a id="ui-0224"></a>**UI-0224** The button of a card has a hover background. ([Judgement: the cursor of a button](#ui-0283))
- <a id="ui-0225"></a>**UI-0225** A link is visibly different from text, also without colour. ([WCAG 2.2, SC 1.4.1][wcag-1-4-1])
- <a id="ui-0226"></a>**UI-0226** A link can have an icon before its text. Only the text of the link is underlined. ([Judgement: the icons of links and buttons](#ui-0284))
- <a id="ui-0227"></a>**UI-0227** A link never looks like a button. ([Judgement: the place of the links and the buttons of a record](#ui-0280))
- <a id="ui-0228"></a>**UI-0228** A link or a button of an action has an icon before its text, and keeps its text. ([Judgement: the icons of links and buttons](#ui-0284))
- <a id="ui-0229"></a>**UI-0229** Each type of action has one icon in all UIs. The UI gives the icon. ([Judgement: the icons of links and buttons](#ui-0284))
- <a id="ui-0230"></a>**UI-0230** Assistive technology does not announce the icon of a link, a button or an item of the side navigation. ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]; [Judgement: the icons of links and buttons](#ui-0284))

## Records

- <a id="ui-0231"></a>**UI-0231** A section is a region of a page with a heading, optional links directly after the heading, and an optional description. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; [SC 2.4.6][wcag-2-4-6])
- <a id="ui-0232"></a>**UI-0232** The links of a section go to a page of the record of the section. For example, a link goes to the form of the record. ([Judgement: the place of the links and the buttons of a record](#ui-0280))
- <a id="ui-0233"></a>**UI-0233** The heading level of a section follows the structure of the page. Sections can nest. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0234"></a>**UI-0234** A section of one record is a card. A card has a border, and a title bar with the heading and the links. ([GOV.UK Design System: Summary list, summary cards][govuk-summary-card]; [Judgement: the separation of the sections of a page](#ui-0281))
- <a id="ui-0235"></a>**UI-0235** A card in a card has a different background. ([Judgement: the separation of the sections of a page](#ui-0281))
- <a id="ui-0236"></a>**UI-0236** A section of a group of records has no card. Its heading has more space above it. ([Judgement: the separation of the sections of a page](#ui-0281))
- <a id="ui-0237"></a>**UI-0237** A page has at most two levels of cards. ([Judgement: the separation of the sections of a page](#ui-0281))
- <a id="ui-0238"></a>**UI-0238** A card can collapse to its title bar. The heading of the card then has a button with the expanded state. ([WAI-ARIA APG: Disclosure][apg-disclosure]; [Judgement: the collapse of cards](#ui-0282))
- <a id="ui-0239"></a>**UI-0239** The title bar of a collapsible card shows the status of its record and its links. ([Judgement: the collapse of cards](#ui-0282))
- <a id="ui-0240"></a>**UI-0240** The collapsed content of a card is `hidden="until-found"`, so that a find in the page opens the card. ([HTML: the hidden attribute][html-hidden])
- <a id="ui-0241"></a>**UI-0241** The UI gives the cards that start collapsed. The card of the record of the page never collapses. ([Judgement: the collapse of cards](#ui-0282))
- <a id="ui-0242"></a>**UI-0242** When a form sends the user back to the page, the card of the record that the form changed or added is open. The cards around that card are also open. ([Judgement: the collapse of cards](#ui-0282))
- <a id="ui-0243"></a>**UI-0243** A reload of the page does not keep the open state of the cards. ([Judgement: the collapse of cards](#ui-0282))
- <a id="ui-0244"></a>**UI-0244** A section of a group can have "Show all" and "Hide all" buttons at its heading. They open or close each card in the group, also the cards in a nested group. ([Judgement: the collapse of cards](#ui-0282))
- <a id="ui-0245"></a>**UI-0245** The accessible names of "Show all" and "Hide all" say what they open or close. For example, "Show all software products of Mock Budget". ([WCAG 2.2, SC 2.4.6][wcag-2-4-6])
- <a id="ui-0246"></a>**UI-0246** The buttons that act on one record are in one action row, after the details of the record and before its child records. ([Judgement: the place of the links and the buttons of a record](#ui-0280))
- <a id="ui-0247"></a>**UI-0247** A description list gives the key-value pairs of the detail of a record as an HTML description list (`dl`), so that each value is connected to its term. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0248"></a>**UI-0248** A value of a description list can be a list of values. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0249"></a>**UI-0249** An empty value of a description list shows the text "Not set". ([Judgement: an empty value in a description list](#ui-0288))
- <a id="ui-0250"></a>**UI-0250** When no value is a normal state, an empty list of values can show a different text. For example, "None". ([Judgement: an empty value in a description list](#ui-0288))
- <a id="ui-0251"></a>**UI-0251** A long value of a description list wraps, so that the page has no horizontal scroll. For example, a URI or a thumbprint wraps. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10])
- <a id="ui-0252"></a>**UI-0252** A timeline shows the events of a record as an ordered list (`ol`), newest first. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0253"></a>**UI-0253** Each event of a timeline has its time in a `time` element, with the ISO 8601 UTC value in `datetime`. ([HTML: the time element][html-time])
- <a id="ui-0254"></a>**UI-0254** Each event of a timeline has a title and optional details. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0255"></a>**UI-0255** A timeline has pagination below the list, and [the empty, loading and error states](#ui-0275). ([Judgement: the pagination controls](#ui-0286))

## Identifiers

- <a id="ui-0256"></a>**UI-0256** A copy button copies an identifier, and announces the result. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="ui-0257"></a>**UI-0257** The identifier component shows an identifier as [identifiers in full](content.md#ui-0162) give it. ([Judgement: the identifier as a component](#ui-0289))
- <a id="ui-0258"></a>**UI-0258** The identifier component can show the identifier in a usual format. The copy button then copies the identifier with no format. For example, a BSB shows in two groups of three digits. ([Judgement: the identifier as a component](#ui-0289))
- <a id="ui-0259"></a>**UI-0259** An identifier wraps in a description list. A table column can keep it on one line. ([WCAG 2.2, SC 1.4.10][wcag-1-4-10])

## Lists

- <a id="ui-0260"></a>**UI-0260** A data table has column headers. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0261"></a>**UI-0261** A data table has row selection only when an action needs it. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0262"></a>**UI-0262** A data table has pagination below the table, and [the empty, loading and error states](#ui-0275). ([Judgement: the pagination controls](#ui-0286))
- <a id="ui-0263"></a>**UI-0263** A column of a data table sorts only when the server can sort it. ([WAI-ARIA APG: Sortable Table][apg-sortable-table]; [Judgement: the sort of the data table](#ui-0287))
- <a id="ui-0264"></a>**UI-0264** A sort button sorts an unsorted column ascending, and turns a sorted column the other way. ([WAI-ARIA APG: Sortable Table][apg-sortable-table])
- <a id="ui-0265"></a>**UI-0265** The sorted column has `aria-sort`. A status message also gives the sort. ([WAI-ARIA APG: Sortable Table][apg-sortable-table]; [WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [Judgement: the sort of the data table](#ui-0287))
- <a id="ui-0266"></a>**UI-0266** Pagination has a Previous button and a Next button. It has no list of page numbers and no page-size control. ([Judgement: the pagination controls](#ui-0286))
- <a id="ui-0267"></a>**UI-0267** Pagination shows the text "Page X of Y" and the number of records, each as a status message. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [Judgement: the pagination controls](#ui-0286))
- <a id="ui-0268"></a>**UI-0268** The status text of a status badge is always visible. Colour only supports the text. ([WCAG 2.2, SC 1.4.1][wcag-1-4-1])
- <a id="ui-0269"></a>**UI-0269** A filter bar has an optional search field and the filters of a list. Each filter has a visible label. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2])
- <a id="ui-0270"></a>**UI-0270** A filter is a native `select` element. ([Judgement: the search of the filter bar](#ui-0285))
- <a id="ui-0271"></a>**UI-0271** The search applies 300 ms after the last change of the text, or at once on the Enter key. ([Judgement: the search of the filter bar](#ui-0285))
- <a id="ui-0272"></a>**UI-0272** When the search field has text, it shows a "Clear search" button. The button empties the field, applies the empty search at once and puts the focus in the field. ([Judgement: the search of the filter bar](#ui-0285))
- <a id="ui-0273"></a>**UI-0273** The focus stays in the search field while the list changes. ([WCAG 2.2, SC 3.2.2][wcag-3-2-2]; [Judgement: the search of the filter bar](#ui-0285))
- <a id="ui-0274"></a>**UI-0274** A filter applies when its value changes. ([WCAG 2.2, SC 3.2.2][wcag-3-2-2]; [Judgement: the search of the filter bar](#ui-0285))

## States

- <a id="ui-0275"></a>**UI-0275** The empty state, the loading state and the error state each have one pattern, which each view uses. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0276"></a>**UI-0276** The error state shows the `title` and the `detail` of a problem details response. ([RFC 9457, section 3.1][rfc9457-3-1])

## Judgements

### <a id="ui-0278"></a>UI-0278 The focus indicator

- Kind: choice
- Basis: Step 3.2. Forced-colours mode removes a box shadow, but it keeps an outline ([MDN: forced-colors][mdn-forced-colors]).
- Reason: An outline keeps the focus visible in forced-colours mode. A consumer uses a touch screen or a small screen more often, so [the consumer pages](consumer-pages.md#ui-0389) use a wider outline.

### <a id="ui-0279"></a>UI-0279 No motion for reduced motion

- Kind: choice
- Basis: Step 3.2. SC 2.3.3 is a Level AAA criterion ([WCAG 2.2, SC 2.3.3][wcag-2-3-3]).
- Reason: The motion durations are tokens, so this rule costs nothing.

### <a id="ui-0280"></a>UI-0280 The place of the links and the buttons of a record

- Kind: choice
- Basis: Step 3.3. No source gives the place of the controls of a record.
- Reason: Links and buttons in one row looked the same. In two places, the user sees which controls go to a page and which change the record, and reads the details before an action. A link with the look of a button announces "link", but does not activate with the Space key.

### <a id="ui-0281"></a>UI-0281 The separation of the sections of a page

- Kind: choice
- Basis: Step 3.2. A border around a record groups its content ([Nielsen Norman Group: Common region][nng-common-region]).
- Reason: Sections with only headings and equal space did not show where a record starts and ends. More than two levels of cards add borders that give no more information.

### <a id="ui-0282"></a>UI-0282 The collapse of cards

- Kind: choice
- Basis: Step 3.2. The disclosure pattern puts the button in the heading ([WAI-ARIA APG: Disclosure][apg-disclosure]).
- Reason: A record with many child records gives a long page, and the lab operator usually wants one record, so the status in the title bar keeps the most important fact visible. The links stay outside the button, because a link in a `summary` is a control in a control. "Show all" and "Hide all" are buttons, because they change the view and do not act on a record.

### <a id="ui-0283"></a>UI-0283 The cursor of a button

- Kind: choice
- Basis: Step 3.3. The browser default for a button is the arrow cursor.
- Reason: With the arrow cursor, the user could not see where to click to open a card. A link already shows the pointer cursor.

### <a id="ui-0284"></a>UI-0284 The icons of links and buttons

- Kind: choice
- Basis: Step 3.2. A decorative image needs no text alternative ([WCAG 2.2, SC 1.1.1][wcag-1-1-1]).
- Reason: An icon helps the user find a control again. An icon for an action, for example a revocation, has no meaning that all users know, so the text gives the meaning.

### <a id="ui-0285"></a>UI-0285 The search of the filter bar

- Kind: choice
- Basis: Step 3.2. A change of the list is a change of content, not of context ([WCAG 2.2, SC 3.2.2][wcag-3-2-2]), and the status message of the pagination announces the result ([SC 4.1.3][wcag-4-1-3]).
- Reason: The delay gives one search for each pause in the typing, not one for each key. The clear button replaces the cancel button of the browser, which only some browsers show. A native select needs no popover and has the keyboard support of the browser.

### <a id="ui-0286"></a>UI-0286 The pagination controls

- Kind: choice
- Basis: Step 3.3. No source gives the pagination controls of a list.
- Reason: The consoles show short lists, and two buttons are the simplest control that works with a keyboard. The status messages announce each change of page and the result of each search or filter ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]).

### <a id="ui-0287"></a>UI-0287 The sort of the data table

- Kind: choice
- Basis: Step 3.2. The sort pattern of the [WAI-ARIA APG: Sortable Table][apg-sortable-table].
- Reason: Screen readers do not always announce a change of `aria-sort`, so a status message also gives the sort. The list always has a sort from the server, so a column has no unsorted state after a sort.

### <a id="ui-0288"></a>UI-0288 An empty value in a description list

- Kind: choice
- Basis: Step 3.3. No source gives the text of an empty value.
- Reason: The term stays in the list, so that the user sees that a field has no value, and an empty value is silent for a screen-reader user. "Not set" says that a value is missing, which is not correct for a list that is empty by design. For example, the accounts of an authorisation with no `bank:` scope.

### <a id="ui-0289"></a>UI-0289 The identifier as a component

- Kind: choice
- Basis: Step 3.3. No source gives the components of a design system.
- Reason: Each list and each detail view shows identifiers. The design system gives one version, not one in each view.

[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[cnsl-csp]: ../../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007
[apg-disclosure]: https://www.w3.org/WAI/ARIA/apg/patterns/disclosure/
[apg-sortable-table]: https://www.w3.org/WAI/ARIA/apg/patterns/table/examples/sortable-table/
[govuk-summary-card]: https://design-system.service.gov.uk/components/summary-list/#summary-cards
[html-hidden]: https://html.spec.whatwg.org/multipage/interaction.html#the-hidden-attribute
[html-time]: https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-time-element
[mdn-forced-colors]: https://developer.mozilla.org/en-US/docs/Web/CSS/@media/forced-colors
[mdn-style-src]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/style-src
[nng-common-region]: https://www.nngroup.com/articles/common-region/
[rfc9457-3-1]: https://www.rfc-editor.org/rfc/rfc9457#section-3.1
[wcag-1-1-1]: https://www.w3.org/TR/WCAG22/#non-text-content
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-1-4-1]: https://www.w3.org/TR/WCAG22/#use-of-color
[wcag-1-4-10]: https://www.w3.org/TR/WCAG22/#reflow
[wcag-1-4-11]: https://www.w3.org/TR/WCAG22/#non-text-contrast
[wcag-2-1-1]: https://www.w3.org/TR/WCAG22/#keyboard
[wcag-2-3-3]: https://www.w3.org/TR/WCAG22/#animation-from-interactions
[wcag-2-4-1]: https://www.w3.org/TR/WCAG22/#bypass-blocks
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-2-4-6]: https://www.w3.org/TR/WCAG22/#headings-and-labels
[wcag-2-4-7]: https://www.w3.org/TR/WCAG22/#focus-visible
[wcag-2-5-7]: https://www.w3.org/TR/WCAG22/#dragging-movements
[wcag-3-2-2]: https://www.w3.org/TR/WCAG22/#on-input
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-4-1-2]: https://www.w3.org/TR/WCAG22/#name-role-value
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
