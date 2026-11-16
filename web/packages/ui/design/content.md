# Design system: Content

This area gives the language of the UIs and the format of dates, times, amounts and identifiers.

## Language

- <a id="ui-0140"></a>**UI-0140** The language of each UI is Australian English. ([Style Manual][style-manual])
- <a id="ui-0141"></a>**UI-0141** Each page has `lang="en-AU"`. ([WCAG 2.2, SC 3.1.1][wcag-3-1-1])
- <a id="ui-0142"></a>**UI-0142** Text in a consumer UI is plain language, at an Australian year 7 reading level. ([Style Manual: Literacy and access][style-literacy])

## Dates and times

- <a id="ui-0143"></a>**UI-0143** A date in text has the form "22 October 2026". ([Style Manual: Dates and time][style-dates])
- <a id="ui-0144"></a>**UI-0144** A date in a table can have the form "22 Oct 2026". ([Style Manual: Dates and time][style-dates])
- <a id="ui-0145"></a>**UI-0145** A date never has a two-digit year. ([Style Manual: Dates and time][style-dates])
- <a id="ui-0146"></a>**UI-0146** A time has the form "2:05 pm". ([Style Manual: Dates and time][style-dates])
- <a id="ui-0147"></a>**UI-0147** A time shows in the time zone of the browser, with its offset from UTC at that instant. For example, "2:05 pm UTC+8" or "1:59 am UTC+10". ([Judgement: UTC for all times (ADR 0012)][adr-0012]; [Judgement: the time zone of a time](#ui-0165))
- <a id="ui-0148"></a>**UI-0148** The offset is "UTC", a sign and the hours, with the minutes when they are not zero. It is not a time zone name or an abbreviation. For example, "UTC+10:30" or "UTC-3:30". ([Judgement: the time zone of a time](#ui-0165))
- <a id="ui-0149"></a>**UI-0149** When the offset is zero, the time has the label "UTC". ([Judgement: the time zone of a time](#ui-0165))
- <a id="ui-0150"></a>**UI-0150** No server, cookie or account keeps the time zone of a user. ([Judgement: the time zone of a time](#ui-0165))
- <a id="ui-0151"></a>**UI-0151** A screen reader gets the offset in words. For example, "UTC plus 8 hours" or "UTC minus 3 hours 30 minutes". ([Judgement: the offset for a screen reader](#ui-0166))
- <a id="ui-0152"></a>**UI-0152** The visible offset is hidden from assistive technology. ([Judgement: the offset for a screen reader](#ui-0166))
- <a id="ui-0153"></a>**UI-0153** A copy of a time has the visible text only. ([Judgement: the offset for a screen reader](#ui-0166))
- <a id="ui-0154"></a>**UI-0154** A time in a timeline has seconds. For example, "2:05:07 pm UTC+8". ([Judgement: seconds in a timeline](#ui-0167))
- <a id="ui-0155"></a>**UI-0155** A date with a time has the date in full, a comma, the time and the offset. For example, "23 October 2026, 2:05:07 pm UTC+8". ([Style Manual: Dates and time][style-dates]; [Judgement: the format of a date with a time](#ui-0168))
- <a id="ui-0156"></a>**UI-0156** When a time value is not valid, the UI shows the value as text, with no `time` element. ([HTML: the time element][html-time]; [Judgement: a time value that is not valid](#ui-0169))

## Amounts

- <a id="ui-0157"></a>**UI-0157** An amount has the form "$1,234.50", with 2 decimal digits. ([Style Manual: Currency][style-currency])
- <a id="ui-0158"></a>**UI-0158** In a consumer UI, a negative amount has a minus sign. For example, "-$12.00". ([Style Manual: Currency][style-currency])
- <a id="ui-0159"></a>**UI-0159** A console does not use the accounting format for an amount. For example, a console does not show "($12.00)". ([Judgement: amounts in a console](#ui-0170))

## Formats

- <a id="ui-0160"></a>**UI-0160** A UI formats dates, times and amounts with `Intl` and the locale `en-AU`. ([ECMA-402][ecma-402]; [Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0161"></a>**UI-0161** A UI formats a value only for display. It does not calculate with a formatted value. ([Judgement: UTC for all times (ADR 0012)][adr-0012])

## Identifiers

- <a id="ui-0162"></a>**UI-0162** An identifier shows in full, in the monospace font, with a copy button. ([Judgement: identifiers in full](#ui-0171))

## Judgements

### <a id="ui-0165"></a>UI-0165 The time zone of a time

- Kind: choice
- Basis: Step 3.3. [ADR 0012][adr-0012] permits a different time zone where a person sees the time. The browser gives the offset at each instant, so the UI needs no time zone data.
- Reason: An offset is not ambiguous, and a lab operator can compare it directly with a time in UTC, for example in the logs. The abbreviations of `Intl` for `en-AU` are not the same in each place, and some have more than one meaning, for example "CST". The offset changes with daylight saving, so the two times of a repeated hour are different.

### <a id="ui-0166"></a>UI-0166 The offset for a screen reader

- Kind: choice
- Basis: Step 3.3. No source gives how a screen reader reads an offset.
- Reason: A screen reader can read "-" as "dash", or not at all at its default punctuation level, which gives the wrong offset. Words give one reading in each screen reader.

### <a id="ui-0167"></a>UI-0167 Seconds in a timeline

- Kind: choice
- Basis: Step 3.3. No source gives the precision of a time in a timeline.
- Reason: More than one event can occur in one minute. For example, an operator action and its cascade.

### <a id="ui-0168"></a>UI-0168 The format of a date with a time

- Kind: choice
- Basis: Step 3.3. The Style Manual gives the date and the time, but not one form for the two together ([Style Manual: Dates and time][style-dates]).
- Reason: The text between the parts of `Intl` changes from one ICU version to the next, for example "at" or a comma. Thus the UI puts the parts together itself.

### <a id="ui-0169"></a>UI-0169 A time value that is not valid

- Kind: choice
- Basis: Step 3.3. The API contract gives each time as a `date-time`, so only a defect gives a different value.
- Reason: One raw value is better for the user than a page that stops.

### <a id="ui-0170"></a>UI-0170 Amounts in a console

- Kind: choice
- Basis: Step 3.3. The Style Manual gives no accounting format ([Style Manual: Currency][style-currency]).
- Reason: One format for all UIs is simpler. The operator consoles show few amounts.

### <a id="ui-0171"></a>UI-0171 Identifiers in full

- Kind: choice
- Basis: Step 3.3. No source gives how a lab UI shows an identifier.
- Reason: A lab operator copies identifiers into other tools.

[adr-0012]: ../../../../docs/adr/0012-utc-for-all-times.md
[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[style-manual]: https://www.stylemanual.gov.au/
[style-literacy]: https://www.stylemanual.gov.au/accessible-and-inclusive-content/literacy-and-access
[style-dates]: https://www.stylemanual.gov.au/grammar-punctuation-and-conventions/numbers-and-measurements/dates-and-time
[style-currency]: https://www.stylemanual.gov.au/grammar-punctuation-and-conventions/numbers-and-measurements/currency
[wcag-3-1-1]: https://www.w3.org/TR/WCAG22/#language-of-page
[html-time]: https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-time-element
[ecma-402]: https://tc39.es/ecma402/
