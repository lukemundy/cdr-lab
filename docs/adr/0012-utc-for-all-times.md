# 0012. Use UTC for all times, and a different time zone only for display

- Date: 2026-10-18

## Context

The services calculate with times, keep them and send them to other participants, for example the expiry of a token, the end of a sharing period and the time of an event.

The CDS requires that each date-time in an API is "offset relative to UTC" ([CDS: Common Field Types](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#common-field-types)). PostgreSQL keeps a `timestamptz` as an instant, and gives it in the time zone of the session ([PostgreSQL: Date/Time Types](https://www.postgresql.org/docs/18/datatype-datetime.html)). A Go `time.Time` keeps a location, and two values for the same instant in different locations format differently.

Some times are for a person, for example the end date of a sharing period on a consent page. A person can expect their local time there.

## Decision

- All times in a service are UTC: in `domain`, in events, in storage, in calculations and in the APIs. An adapter that gets a time from the database, a library or a different participant gives it to `app` in UTC.
- A conversion to a different time zone occurs only in an adapter, at the point where a person sees the time: a page, an email or an operator response.
- A time that a person sees in UTC has the label "UTC".

This decision applies to services and to `pkg/` libraries.

## Consequences

- A calculation has no time zone rules and no daylight saving changes, and a test of it needs no time zone.
- Two values for the same instant format the same, in each service and in each test.
- A person sees UTC until an adapter converts the time. A local time for a person needs a time zone from a setting or from the person, in that adapter.
- A rule that depends on a local date, for example "today" for a brand, needs its time zone as an input of the rule. UTC does not give it.
