# 0003. Use event sourcing for participant state

- Date: 2026-09-29
- Amended by [ADR-0021](0021-record-one-business-fact-in-each-event-type.md): the design of event types

## Context

The Register keeps participants whose statuses change over time. The status rules are state machines with cascades: for example, a data recipient suspension makes its ACTIVE software products INACTIVE at the same time ([Register DESIGN.md: Status rules](../../services/register/DESIGN.md#status-rules)). The Register must also never issue a REMOVED software product identifier again.

Later services have stronger needs. Data holders and accredited data recipients must keep records of each consumer authorisation or consent, and of each amendment and withdrawal, including how the consumer asked for the withdrawal. They must keep these records for 6 years ([CDR Rules, rule 9.3](https://www.legislation.gov.au/F2020L00094/latest/text)). A history of changes is a primary requirement there, not a secondary one.

We compared two options for the Register:

| | Status columns, with a history table | Event sourcing |
|---|---|---|
| Current state | Direct read of a row | A read model, which a projection keeps up to date |
| History | A second write, which code can forget | The only write. The history is the data. |
| Status rules | In the domain, then a row update | In the domain. The events are the result. |
| `updated-since` and `ETag` | Timestamps that each write must update | From the global position of the events |
| Removed identifiers | A rule on each insert | The stream stays, so the identifier stays |
| Complexity | Low | Higher: projections, event schema versions, new terms |

For the Register data only, status columns are sufficient. We chose event sourcing because the consent records of later services need it, and one pattern for all participant state is simpler than two. The maintainer also knows the pattern and wants to learn more about it.

Many event sourcing terms have synonyms, and teams use them in different ways. Thus this decision also sets the terms.

## Decision

### Terms

Use these terms in code, comments, documents and tickets. Do not use the synonyms in the last column. The functions handle and apply are verbs: "handle a command", "apply an event".

| Term | Meaning | Do not use |
|---|---|---|
| Event | A fact that occurred, with a past-tense name | domain event |
| Event type | The stable string name of an event in storage | event name, kind |
| Payload | The data of an event. The domain defines it. | body, data |
| Envelope | The stored record around a payload | record, message |
| Metadata | Data in the envelope about the event that is not domain data. Now, this is only the W3C trace context. | headers, context |
| Aggregate | One consistency boundary, with one stream | aggregate root, entity |
| Aggregate type | The definition of one kind of aggregate: its stream type, initial state, handle, apply, and the codec of its events | decider, aggregate class |
| Stream | The ordered events of one aggregate | event stream, aggregate stream |
| Stream type | The stored name of an aggregate type | category |
| Stream version | The position of an event in its stream, starting at 1 | revision, sequence number |
| Expected version | The stream version that the caller read. The append fails if the stream version is different. | — |
| Global position | The order of an event across all streams | offset, checkpoint, sequence |
| Command | A request to change an aggregate, with an imperative name. The aggregate can reject it. | intent, action |
| State | The value that apply makes from the events of a stream | aggregate state, snapshot |
| Handle | A pure function: (state, command) → events or an error | decide |
| Apply | A pure function: (state, event) → state | evolve, when, mutate |
| Load | To read a stream and fold it with apply | rehydrate, replay |
| Execute | To load, handle, and append with the expected version | dispatch, process, command handler |
| Projection | Code that applies events to a read model, in the append transaction | denormaliser, view builder |
| Read model | The tables that a projection writes and that queries read | view, query model |

Add a term to this table only when the code uses it.

### Model

- Handle and apply come from the Decider pattern ([Chassaing: Functional Event Sourcing Decider](https://thinkbeforecoding.com/post/2021/12/17/functional-event-sourcing-decider)), where they are `decide` and `evolve`. They are pure. Thus they go in `domain`, which does no I/O and does not read the clock.
- An aggregate is the smallest unit that contains all the rules of one change. The status cascades change a data recipient legal entity and its software products together. Thus the Register has two aggregate types:
  - Data recipient: a legal entity, its brands and their software products.
  - Data holder: a legal entity and its brands.

  The stream ID is the legal entity ID. A command for a brand or a software product needs the legal entity ID, or a read model that gives it.

### Code

- The generic code is in `pkg/`, because more than one service uses it. It contains no code that is specific to one service.
  - `pkg/eventsourcing`: the aggregate type, load, execute, the `Store` port, the envelope types and a JSON codec. It imports only the standard library, so that `app` can use it. The linter enforces this.
  - `pkg/eventsourcing/eventstore`: the PostgreSQL `Store`.
- The domain does not import `eventsourcing`. The app layer makes each aggregate type from the domain types and functions.

### Storage

- One `events` table ([ADR 0002](0002-data-store.md)). A unique constraint on (stream type, stream ID, stream version) gives optimistic concurrency. When two appends use the same expected version, one of them fails with a version conflict.
- One append can add events to more than one stream, in one transaction.
- Payloads are JSON (`jsonb`), with the event type and a schema version in the envelope. All payloads are schema version 1 now.
- The metadata contains the W3C trace context of the append, so that an event links to the trace that caused it.
- The global position is an identity column. It has gaps, because "PostgreSQL sequence objects cannot be used to obtain gapless sequences" ([PostgreSQL: Sequence Manipulation Functions](https://www.postgresql.org/docs/18/functions-sequence.html)). It is also not in commit order: a transaction that gets a lower position can commit after one that gets a higher position.
- Projections run in the append transaction. A read model is then correct as soon as the append commits, which meets the propagation requirement ([a status change at once on each Register API](../../services/register/design/non-functional-requirements.md#reg-0001)). It also avoids the commit-order problem above, because no reader follows the global position.

## Consequences

- Each change to participant state is an event. The history is complete, and a read model can be deleted and made again from the events.
- Each read API needs a read model and a projection. The first read API adds the projection mechanism.
- A consumer that follows the global position outside the append transaction, for example a River job, must allow for positions that commit out of order. Decide how when the first such consumer exists.
- A change to the payload of an event needs a new schema version and code that reads the old versions.
- Payloads contain participant data, and later CDR data. Never put payloads in logs, metrics or traces. Database spans already contain no query arguments ([ADR 0002](0002-data-store.md)).
- New code uses the terms above. A review rejects synonyms.
