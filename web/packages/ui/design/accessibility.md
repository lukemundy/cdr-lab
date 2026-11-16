# Design system: Accessibility

This area gives the accessibility level of each type of UI, the automated accessibility checks and the accessible names of repeated links.

## Accessibility levels

- <a id="ui-0110"></a>**UI-0110** A consumer UI meets WCAG 2.2 Level AA. ([CDS: Accessibility Standards][cds-accessibility]; [AHRC guidelines][ahrc]; [Judgement: two accessibility levels (ADR 0016)][adr-0016])
- <a id="ui-0111"></a>**UI-0111** An operator console has no accessibility target other than the accessibility of the components and the automated checks. ([Judgement: two accessibility levels (ADR 0016)][adr-0016])

## Automated checks

- <a id="ui-0112"></a>**UI-0112** The automated checks run axe-core with these tags. ([axe-core API][axe-api]; [Judgement: the automated checks use all WCAG Level A and AA tags](#ui-0120))
  - `wcag2a`
  - `wcag2aa`
  - `wcag21a`
  - `wcag21aa`
  - `wcag22aa`
- <a id="ui-0113"></a>**UI-0113** A violation that an automated check finds is a failure. ([Judgement: two accessibility levels (ADR 0016)][adr-0016])
- <a id="ui-0114"></a>**UI-0114** Each component in each state, and each page of each UI, passes the automated checks. ([Judgement: two accessibility levels (ADR 0016)][adr-0016])
- <a id="ui-0115"></a>**UI-0115** An exception to an automated check has a comment with its reason, at the place of the exception. ([Judgement: each exception to an automated check has a comment with its reason](#ui-0121))

## Accessible names

- <a id="ui-0116"></a>**UI-0116** When a page has the same link text for more than one record, the accessible name of each link also names its record. Only assistive technology gets the added text. For example, "Change brand" for each brand of a data recipient. ([WCAG 2.2, SC 2.4.4][wcag-2-4-4]; [Judgement: the accessible name of a repeated link names its record](#ui-0122))

## Judgements

### <a id="ui-0120"></a>UI-0120 The automated checks use all WCAG Level A and AA tags

- Kind: choice
- Basis: Step 3.2. axe-core runs only the rules of the tags that it gets ([axe-core API][axe-api]).
- Reason: WCAG 2.2 Level AA includes the criteria of WCAG 2.0 and 2.1, so the checks need each tag. The two levels of [ADR 0016][adr-0016] still give automated checks for the consoles.

### <a id="ui-0121"></a>UI-0121 Each exception to an automated check has a comment with its reason

- Kind: choice
- Basis: Step 3.3. No source gives a rule for an exception to an automated check.
- Reason: A reviewer sees why the check does not apply at that place.

### <a id="ui-0122"></a>UI-0122 The accessible name of a repeated link names its record

- Kind: choice
- Basis: Step 3.2. A heading is not a context of a link that SC 2.4.4 accepts ([WCAG 2.2, SC 2.4.4][wcag-2-4-4]).
- Reason: A screen-reader user who lists the links of a page can tell each link from the others.

[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[cds-accessibility]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#accessibility-standards
[ahrc]: https://humanrights.gov.au/our-work/disability-rights/publications/guidelines-equal-access-digital-goods-and-services
[axe-api]: https://github.com/dequelabs/axe-core/blob/develop/doc/API.md
[wcag-2-4-4]: https://www.w3.org/TR/WCAG22/#link-purpose-in-context
