# 0028. Find the need for a fresh deployment from the migrations and the event golden files

- Date: 2026-11-23

## Context

This ADR amends [ADR-0003](0003-event-sourcing.md). It replaces the consequence "A change to the payload of an event needs a new schema version and code that reads the old versions". While the major version is 0, such a change needs a fresh deployment instead.

A release makes no promise about the data of a deployment while the major version is 0 ([ADR 0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md)). A team that upgrades a long-lived deployment must still know before the upgrade when the new version cannot use the data of the earlier version. The team then makes a fresh deployment: a new database and new participants.

Two kinds of change make the data of an earlier version unusable:

- **A changed, renamed or deleted migration.** goose records each applied migration by its version, and does not run an applied migration again ([ADR 0002](0002-data-store.md)). A database of the earlier version then has a schema that the new migrations do not give. A new migration file applies to an existing database, and needs no fresh deployment.
- **A changed event payload.** The JSON codec of `pkg/eventsourcing` reads only schema version 1, and it refuses a payload field that the Go type does not have. A field that is renamed, removed or of a different type makes the stored events of the earlier version unreadable. A new event type does not change the stored events.

The payload Go types are in the domain files of each service, together with the domain logic. A path check on those files flags each change of domain logic, also when no payload changes.

The real options were:

1. **A commit footer**, for example `Lab-Reset: required`. The squash merge keeps only the title of the pull request ([ADR 0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md)), so the footer reaches `main` only when the person who merges types it. The signal depends on a person.
2. **`!` in the title of the pull request.** It reaches `main`. But `!` also marks API changes that keep the data, and a person can forget it. It is a mixed signal.
3. **The diff of the release.** The changed migration files are in the diff. A payload change is in the diff only when a file in the repository records the encoded payloads. A test can keep such a file correct: it encodes one sample of each registered event type and compares the result with the file. The codec gives the list of registered event types, so a test can find an event type with no sample.

## Decision

A release needs a fresh deployment when the diff from the earlier version tag to the release commit has one of these changes:

- a migration file that the earlier version had is changed, renamed or deleted
- an entry of an event golden file that the earlier version had is changed or removed

We find the need from the diff only, because the diff does not depend on a person. A new migration file and a new event type do not count.

The release notes start with a fixed section, "Needs a fresh deployment", that gives the result and lists the files.

Each service that keeps events has an **event golden file**: a committed file with one encoded sample of each event type that the service registers. A test of the service obeys these rules:

- It gets the registered event types from the codec, and fails when a registered event type has no sample.
- It encodes each sample with the real codec of the service, and fails when the result is not the same as the file.
- It writes the file again only when a person asks for it, for example with a flag.

The file uses sample JSON, not a schema of field names and types, because sample JSON is simpler and shows the real encoding.

This decision applies while the release makes no data promise. A later ADR that gives a data promise replaces it.

## Consequences

- A team reads one section of the notes to know whether an upgrade keeps its data.
- Each new event type needs a sample, and each change of a payload changes the golden file in the same pull request. A reviewer sees the payload change in the diff of that file.
- A change of a migration that is already on `main` makes the next release need a fresh deployment. An author who wants to keep the data adds a new migration instead.
- The check finds a change of the shape of a payload, not of its meaning. For example, a field that keeps its name and type but changes its units is not found. The person who starts the release can add such a change to the notes, because the notes of a GitHub Release stay editable.
- No code reads an earlier payload. A payload change before 1.0.0 costs a fresh deployment, not a schema version.
