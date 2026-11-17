# Design system: Forms and dialogs

This area gives the dialogs, the form fields, the forms, the alerts and the toasts.

## Dialogs

- <a id="ui-0290"></a>**UI-0290** A dialog is modal and has a title. ([WAI-ARIA APG: Dialog][apg-dialog])
- <a id="ui-0291"></a>**UI-0291** When a dialog opens, the focus moves into the dialog. When the dialog closes, the focus goes back to the control that opened it. ([WAI-ARIA APG: Dialog][apg-dialog]; [WCAG 2.2, SC 2.4.3][wcag-2-4-3])
- <a id="ui-0292"></a>**UI-0292** When the control that opened a dialog does not exist after the action, the focus moves to an element that the UI gives. For example, the heading of the record. ([WAI-ARIA APG: Dialog][apg-dialog])
- <a id="ui-0293"></a>**UI-0293** A confirmation dialog is an alert dialog that names the action and its effects. When the action has more than one effect, the effects are a list. ([WAI-ARIA APG: Alert Dialog][apg-alertdialog]; [Judgement: an alert dialog with the first focus on cancel](#ui-0345))
- <a id="ui-0294"></a>**UI-0294** The confirm button of a confirmation dialog names the action. A cancel button closes the dialog with no change. ([WAI-ARIA APG: Alert Dialog][apg-alertdialog])
- <a id="ui-0295"></a>**UI-0295** The first focus of a confirmation dialog is on the cancel button, also when the action can be reversed. ([WAI-ARIA APG: Dialog][apg-dialog]; [Judgement: an alert dialog with the first focus on cancel](#ui-0345))
- <a id="ui-0296"></a>**UI-0296** When the action cannot be reversed, the confirmation dialog says so. ([WCAG 2.2, SC 3.3.4][wcag-3-3-4])
- <a id="ui-0297"></a>**UI-0297** The confirm button has the danger variant when the action cannot be reversed, and the primary variant when it can. The UI cannot change the variant. ([Judgement: a danger confirm button only for an action that cannot be reversed](#ui-0348))
- <a id="ui-0298"></a>**UI-0298** While the request runs, the confirm button is busy, a status message says so, and the dialog stays open. For example, the status message is "Suspend in progress". ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [Judgement: a busy button and a status message during a confirmation](#ui-0346))
- <a id="ui-0299"></a>**UI-0299** When the action fails, the confirmation dialog stays open and shows the error in an alert. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="ui-0300"></a>**UI-0300** When the action fails with no problem details, the alert says "The action did not complete" and "Refresh the page to see the current state, then try again". ([Judgement: no automatic retry for an action that fails with no problem details](#ui-0347))
- <a id="ui-0301"></a>**UI-0301** The UI never sends an action again by itself. ([Judgement: no automatic retry for an action that fails with no problem details](#ui-0347))

## Form fields

- <a id="ui-0302"></a>**UI-0302** A form field has a visible label, optional help text, and an error message that is connected to the field. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [SC 3.3.2][wcag-3-3-2])
- <a id="ui-0303"></a>**UI-0303** The error message of a field is below the label and the help text, and above the control. ([GOV.UK Design System: Error message][govuk-error]; [Judgement: the error message of a field above the control, below the help text](#ui-0354))
- <a id="ui-0304"></a>**UI-0304** The help text and the error message describe the control, in that order. ([Judgement: the error message of a field above the control, below the help text](#ui-0354))
- <a id="ui-0305"></a>**UI-0305** An error from the server for the value of a field shows as the error message of that field. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [the field errors of a problem][prob-field-errors])
- <a id="ui-0306"></a>**UI-0306** The error of an item of a list, or of a value in a group, shows at the field of the list or at the group. For example, an error at `#/redirectUris/1` shows at the field of the redirect URIs. ([the pointer of a field error][prob-pointer]; [WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="ui-0307"></a>**UI-0307** A text field, a text area, a select, a checkbox and a radio group are each in a form field. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2])
- <a id="ui-0308"></a>**UI-0308** A select is a native `select` element. ([Judgement: a native select in a form](#ui-0351))
- <a id="ui-0309"></a>**UI-0309** A select can start with no selected value. Its first option is then empty, with a text that asks the user to choose, so that the UI does not choose a value for the user. ([WCAG 2.2, SC 3.3.2][wcag-3-3-2])
- <a id="ui-0310"></a>**UI-0310** A list of values that the user writes is one text area, with one value on each line. Its help text says to give one value on each line. ([Judgement: one text area with one value on each line](#ui-0357))
- <a id="ui-0311"></a>**UI-0311** A date-time field is a native `input` element of type `datetime-local`, with a step of 1 second, in a form field. ([HTML: Local Date and Time state][html-datetime-local]; [Judgement: a native datetime-local input, sent as an instant in UTC](#ui-0352))
- <a id="ui-0312"></a>**UI-0312** A date-time field is in the time zone of the browser. Its help text says that the time is in the time zone of the browser. ([Judgement: UTC for all times (ADR 0012)][adr-0012])
- <a id="ui-0313"></a>**UI-0313** The UI sends the value of a date-time field as an instant in UTC. It converts the value with the offset of the browser at that instant. ([Judgement: UTC for all times (ADR 0012)][adr-0012]; [Judgement: a native datetime-local input, sent as an instant in UTC](#ui-0352))
- <a id="ui-0314"></a>**UI-0314** A checkbox group is a `fieldset` with a `legend`, which is the label of the group. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1]; [GOV.UK Design System: Checkboxes][govuk-checkboxes])
- <a id="ui-0315"></a>**UI-0315** A checkbox group has one help text and one error message, connected to the group. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1]; [GOV.UK Design System: Checkboxes][govuk-checkboxes])
- <a id="ui-0316"></a>**UI-0316** A link to a checkbox group goes to its first checkbox. ([GOV.UK Design System: Checkboxes][govuk-checkboxes])
- <a id="ui-0317"></a>**UI-0317** A group of related fields is a `fieldset`, with a `legend` that names the group. For example, the endpoints of a brand. ([WCAG 2.2, SC 1.3.1][wcag-1-3-1])
- <a id="ui-0318"></a>**UI-0318** A file input is a native `input` element of type `file`, with a visible label and a style on its button. ([HTML: File Upload state][html-file]; [Judgement: a native file input](#ui-0353))
- <a id="ui-0319"></a>**UI-0319** A file input opens the file chooser of the browser, and shows the name of the selected file. ([HTML: File Upload state][html-file])
- <a id="ui-0320"></a>**UI-0320** The UI gets the text of the selected file. ([Judgement: a native file input](#ui-0353))

## Forms

- <a id="ui-0321"></a>**UI-0321** A form has the fields of one record on a page, a submit button that names the action, and a "Cancel" link. ([GOV.UK Design System: Question pages][govuk-question-pages])
- <a id="ui-0322"></a>**UI-0322** The UI gives the fields of a form and its request. ([Judgement: one design system (ADR 0016)][adr-0016])
- <a id="ui-0323"></a>**UI-0323** While the request of a form runs, the submit button is busy, a status message says so, and a second submit sends nothing. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="ui-0324"></a>**UI-0324** After a refusal, the form keeps each value, and shows each field error at its field. ([WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="ui-0325"></a>**UI-0325** A form does not check the values. The server checks them. The exception is a check that the specification of the UI gives with a judgement. ([the field errors of a problem][prob-field-errors]; [Judgement: a form checks a value only when the specification of its UI gives the check](#ui-0396))
- <a id="ui-0326"></a>**UI-0326** The label of an optional field ends with "(optional)". A required field has no mark. ([GOV.UK Design System: Question pages][govuk-question-pages]; [Judgement: only optional fields have a mark](#ui-0358))
- <a id="ui-0327"></a>**UI-0327** After a refusal, the form shows the error summary at its top. The error summary is the alert of the form. ([GOV.UK Design System: Error summary][govuk-error-summary]; [Judgement: an error summary at the top of a form that gets the focus](#ui-0355))
- <a id="ui-0328"></a>**UI-0328** The error summary has the heading "There is a problem". ([GOV.UK Design System: Error summary][govuk-error-summary])
- <a id="ui-0329"></a>**UI-0329** The error summary shows the error of the form, if there is one. ([GOV.UK Design System: Error summary][govuk-error-summary]; [WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="ui-0330"></a>**UI-0330** The error summary has a list with a link to each field that has an error. The text of each link is the error message of its field. ([GOV.UK Design System: Error summary][govuk-error-summary]; [WCAG 2.2, SC 3.3.1][wcag-3-3-1])
- <a id="ui-0331"></a>**UI-0331** When the error summary appears, the focus moves to it. A link of the error summary moves the focus to its field. ([GOV.UK Design System: Error summary][govuk-error-summary]; [Judgement: an error summary at the top of a form that gets the focus](#ui-0355))
- <a id="ui-0332"></a>**UI-0332** When a form fails with no problem details, the error summary says "The change did not complete" and "Refresh the page to see the current state, then try again". ([Judgement: a form that fails with no problem details asks for a refresh](#ui-0356))

## Alerts and toasts

- <a id="ui-0333"></a>**UI-0333** An alert shows an error on the page or in a dialog. The screen reader announces the alert when it appears. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3])
- <a id="ui-0334"></a>**UI-0334** An alert stays until its cause changes. ([Judgement: errors in alerts, and only successes in toasts](#ui-0349))
- <a id="ui-0335"></a>**UI-0335** An error shows in an alert at the place of the action. An error never shows only in a toast. ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]; [Judgement: errors in alerts, and only successes in toasts](#ui-0349))
- <a id="ui-0336"></a>**UI-0336** A toast shows the result of a successful action, as a status message. Only a success shows in a toast. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [Judgement: errors in alerts, and only successes in toasts](#ui-0349))
- <a id="ui-0337"></a>**UI-0337** A toast has a close button. ([WCAG 2.2, SC 2.2.1][wcag-2-2-1])
- <a id="ui-0338"></a>**UI-0338** A toast closes by itself after 5 seconds or more, but not while it has the pointer or the focus. ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]; [Judgement: a toast closes after 5 seconds or more](#ui-0350))
- <a id="ui-0339"></a>**UI-0339** A toast is never the only place of its information. ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]; [Judgement: a toast closes after 5 seconds or more](#ui-0350))
- <a id="ui-0340"></a>**UI-0340** The close button of a toast is hidden from assistive technology until the toast has the pointer or the focus. ([Base UI: Toast][baseui-toast]; [Judgement: a toast closes after 5 seconds or more](#ui-0350))
- <a id="ui-0341"></a>**UI-0341** The F6 key moves the focus to the toasts. ([Base UI: Toast][baseui-toast])

## Judgements

### <a id="ui-0345"></a>UI-0345 An alert dialog with the first focus on cancel

- Kind: choice
- Basis: Step 3.2. The APG gives the alert dialog for a message that needs a response. It advises the focus on the least destructive action for a step that is not easily reversible ([WAI-ARIA APG: Dialog][apg-dialog]).
- Reason: The user confirms with a deliberate key press or click. One rule for all confirmations is simpler, and an action that can be reversed can also change other records, for example a suspension.

### <a id="ui-0346"></a>UI-0346 A busy button and a status message during a confirmation

- Kind: choice
- Basis: Step 3.2. A screen reader announces a status message ([WCAG 2.2, SC 4.1.3][wcag-4-1-3]).
- Reason: A screen reader does not always announce a change to the name of the focused button.

### <a id="ui-0347"></a>UI-0347 No automatic retry for an action that fails with no problem details

- Kind: choice
- Basis: Step 3.3. No source gives the response of a UI to a request that fails with no problem details.
- Reason: When the request fails, for example with a network error, the UI does not know if the change occurred. A second request can then do the change two times.

### <a id="ui-0348"></a>UI-0348 A danger confirm button only for an action that cannot be reversed

- Kind: choice
- Basis: Step 3.3. No source gives the variant of a confirm button.
- Reason: One rule in the design system makes each confirmation look the same for the same type of action.

### <a id="ui-0349"></a>UI-0349 Errors in alerts, and only successes in toasts

- Kind: choice
- Basis: Step 3.2. Content that closes by itself needs a way for the user to keep it ([WCAG 2.2, SC 2.2.1][wcag-2-2-1]).
- Reason: A toast closes by itself, and an error must stay until the user reads it and acts.

### <a id="ui-0350"></a>UI-0350 A toast closes after 5 seconds or more

- Kind: choice
- Basis: Step 3.2. 5 seconds is the Base UI default ([Base UI: Toast][baseui-toast]).
- Reason: The result is also in the view, so the toast gives no information that the user can lose. The close button is hidden until the toast has the pointer or the focus, so that the announcement does not include it.

### <a id="ui-0351"></a>UI-0351 A native select in a form

- Kind: choice
- Basis: Step 3.2. The Base UI Select and ScrollArea add a `<style>` element, which the CSP of the web UIs refuses ([Base UI: CSP Provider][baseui-csp]; [the CSP of each web UI][cnsl-csp]).
- Reason: A native select needs no popover and has the keyboard support of the browser. The filter of the filter bar is also a native select for the same reason.

### <a id="ui-0352"></a>UI-0352 A native datetime-local input, sent as an instant in UTC

- Kind: choice
- Basis: Step 3.2. The browser gives the control, its keyboard support and a value that is a valid local time or empty, in each supported browser ([HTML: Local Date and Time state][html-datetime-local]).
- Reason: A text field needs a format that the user must learn. The offset at the instant, not the offset of today, gives the correct time on the other side of a change of daylight saving. The browser gives the text of the control in the language of the browser.

### <a id="ui-0353"></a>UI-0353 A native file input

- Kind: choice
- Basis: Step 3.2. The browser gives the file chooser, its keyboard support and the name of the selected file, in each supported browser ([HTML: File Upload state][html-file]).
- Reason: A control that the UI makes itself needs a hidden input and a second focus target. The cost is that the browser gives the text of the button and of the file name in the language of the browser, not in Australian English.

### <a id="ui-0354"></a>UI-0354 The error message of a field above the control, below the help text

- Kind: choice
- Basis: Step 3.2. The [GOV.UK Design System: Error message][govuk-error] uses the same order.
- Reason: A user with a screen magnifier sees the error next to the label, also above a tall text area.

### <a id="ui-0355"></a>UI-0355 An error summary at the top of a form that gets the focus

- Kind: choice
- Basis: Step 3.2. The [GOV.UK Design System: Error summary][govuk-error-summary] shows the summary for each error and moves the focus to it.
- Reason: A form of a record can be longer than one screen, so an error message at a field far down the page is not seen. The summary is also the alert of the form, so that a form has one place for each error.

### <a id="ui-0356"></a>UI-0356 A form that fails with no problem details asks for a refresh

- Kind: choice
- Basis: Step 3.3. No source gives the response of a form to a request that fails with no problem details.
- Reason: The UI does not know if the change occurred. The text agrees with [the text of an action that fails with no problem details](#ui-0300).

### <a id="ui-0357"></a>UI-0357 One text area with one value on each line

- Kind: choice
- Basis: Step 3.3. No source gives the control of a list of values.
- Reason: One control has one label and one error message, the user can paste a list, and the keyboard and the screen reader read it line by line. A list of text fields needs add and remove buttons and a focus rule for each.

### <a id="ui-0358"></a>UI-0358 Only optional fields have a mark

- Kind: choice
- Basis: Step 3.2. The [GOV.UK Design System: Question pages][govuk-question-pages] marks the optional fields only.
- Reason: Most fields of the records are required, so the optional fields are the exception to mark.

### <a id="ui-0396"></a>UI-0396 A form checks a value only when the specification of its UI gives the check

- Kind: choice
- Basis: Step 3.3. The server gives each field error of a refusal ([the field errors of a problem][prob-field-errors]), and no source requires a check in the browser.
- Reason: The server keeps the only copy of the value rules, so a UI and its API cannot disagree. Some values never get to the server as the user typed them, for example a value that the UI does not send, or a format that the UI converts before it sends the value. A UI then checks the value in the browser, and its own specification gives the check with a judgement, so that each exception is a decision.

[adr-0012]: ../../../../docs/adr/0012-utc-for-all-times.md
[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
[cnsl-csp]: ../../../../pkg/platform/consolehttp/DESIGN.md#cnsl-0007
[prob-field-errors]: ../../../../pkg/platform/problem/DESIGN.md#prob-0003
[prob-pointer]: ../../../../pkg/platform/problem/DESIGN.md#prob-0004
[apg-alertdialog]: https://www.w3.org/WAI/ARIA/apg/patterns/alertdialog/
[apg-dialog]: https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/
[baseui-csp]: https://base-ui.com/react/utils/csp-provider
[baseui-toast]: https://base-ui.com/react/components/toast
[govuk-checkboxes]: https://design-system.service.gov.uk/components/checkboxes/
[govuk-error]: https://design-system.service.gov.uk/components/error-message/
[govuk-error-summary]: https://design-system.service.gov.uk/components/error-summary/
[govuk-question-pages]: https://design-system.service.gov.uk/patterns/question-pages/
[html-datetime-local]: https://html.spec.whatwg.org/multipage/input.html#local-date-and-time-state-(type=datetime-local)
[html-file]: https://html.spec.whatwg.org/multipage/input.html#file-upload-state-(type=file)
[wcag-1-3-1]: https://www.w3.org/TR/WCAG22/#info-and-relationships
[wcag-2-2-1]: https://www.w3.org/TR/WCAG22/#timing-adjustable
[wcag-2-4-3]: https://www.w3.org/TR/WCAG22/#focus-order
[wcag-3-3-1]: https://www.w3.org/TR/WCAG22/#error-identification
[wcag-3-3-2]: https://www.w3.org/TR/WCAG22/#labels-or-instructions
[wcag-3-3-4]: https://www.w3.org/TR/WCAG22/#error-prevention-legal-financial-data
[wcag-4-1-3]: https://www.w3.org/TR/WCAG22/#status-messages
