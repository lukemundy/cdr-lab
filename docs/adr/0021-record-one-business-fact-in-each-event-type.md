# 0021. Record one business fact in each event type

- Date: 2026-10-26

## Context

[ADR 0003](0003-event-sourcing.md) defines an event as "a fact that occurred, with a past-tense name". It does not say when two facts share one event type and when they get separate event types. This ADR amends ADR 0003 with that rule.

Facts of the same kind can have different rules and different data. For example, an arrangement ends for one of several causes ([data holder specification: End of an arrangement](../../services/dataholder/design/end-of-an-arrangement.md)). The data holder sends a notice to the data recipient for some causes only, and only a withdrawal by the consumer has a withdrawal method.

There are three options for facts of the same kind:

| | A generic event with a field that selects the fact | One event type with a union payload | One event type for each fact |
|---|---|---|---|
| Example | `ArrangementEnded{reason}` | `ArrangementEnded{reason: {type, ...}}`, where each `type` has its own fields | `ArrangementWithdrawn`, `ArrangementRevoked` |
| Data of one fact | A field that is valid only for some values of a different field | In its variant only | In its event type only |
| Code | The existing JSON codec | A marshal and unmarshal for the union, which select the variant from a field in the payload | The existing JSON codec, which selects the Go type from the event type in the envelope |
| A change to one fact | Changes the payload of each fact of the kind | Changes the payload of each fact of the kind | Changes one event type |
| A reader that wants each fact of the kind | Reads one event type | Reads one event type | Reads each event type of the kind |

Practitioner guidance does not support the generic event:

- Dudycz calls it "State Obsession": "we flattened business information". He writes that it is "better to replace the generic event with more specific ones, even if they seem very similar at the design time", because "it's important not to think about the state change but about what exactly has happened" ([Dudycz: State Obsession](https://event-driven.io/en/state-obsession/)).
- O'Sullivan gives a test. Split a change into separate events when different readers react to different values, for example `AccountClosed` and `AccountActivated` in place of a status change. Merge changes that readers always handle together. A reader that must examine the payload to know whether it cares shows that "you've bled some logic from your domain into your listeners" ([O'Sullivan: Event granularity](https://dev.to/barryosull/event-granularity-modelling-events-in-event-driven-applications-e50)).
- Öberg: "Having a specific event name is absolutely crucial though, for reporting purposes" ([Kurrent forum: Generic events vs specific events](https://discuss.kurrent.io/t/generic-events-vs-specific-events/1670)).

The guidance also names the opposite fault. Dudycz calls one event for each changed field "Property Sourcing": "The events themselves tell us nothing about the operation that performed them" ([Dudycz: Property Sourcing](https://event-driven.io/en/property-sourcing)). Young: "do not use the word “And” in an event name; use two events instead" ([Young: General Versioning Concerns](https://leanpub.com/read/esversioning/leanpub-auto-general-versioning-concerns)).

These sources are practitioner writing, not standards. None of them compares a union payload with separate event types.

## Decision

- An event type records one business fact. Its name says what occurred, not which state changed.
- Two facts get separate event types when a rule or a reader acts differently on them. An event type does not have a field that selects between such facts. This includes a union payload: its variants are separate facts. The event store already selects the Go type of an event from its event type, so a union adds a second selection in the payload and the code for it.
- Each field of an event type applies to each event of that type.
- A field whose values all rules and readers handle in the same way is permitted, for example a final status that the service only records and shows.
- Facts that one command causes, and that the rules always handle together, are one event. One event for each changed field is not permitted, and an event name does not contain "And".

This decision applies to each event type.

## Consequences

- A rule that acts on one fact reads its event type. It does not examine a field of the payload to find the fact.
- The number of event types increases with the number of facts. A reader that wants each fact of one kind, for example each end of an arrangement, reads more than one event type.
- `ArrangementEnded` does not follow this decision: its `reason` selects between causes that have different rules.
- Go has no check that a type switch handles each event type. A new event type that apply does not handle is a fault that only a test or a review finds.
