# 0024. Write specifications as identified single requirements, with each judgement next to its rules, and split them by area

- Date: 2026-11-15

## Context

This ADR amends [ADR 0013](0013-hexagonal-architecture-with-an-internal-composition-root.md). ADR 0013 puts a `DESIGN.md` in the layout of each service. This ADR replaces that line. The other parts of ADR 0013 do not change.

Each component has a `DESIGN.md`. It gives the requirements of the component, with no implementation detail. Each requirement links to its primary source, or it has "Judgement" and a reason. A Decisions table at the end of the file records the choices that the sources do not make. People and agents read and change these files.

The repository has five `DESIGN.md` files:

- the Register
- the data holder
- the PKI
- the conformance suite
- the design system of the web UIs

On 2026-11-15, a review of the five files found these facts:

- The data holder file has 53,510 words. The Register file has 26,266 words.
- The Decisions tables have approximately 280 and 175 rows. These rows are 39 % and 44 % of the words in the two files.
- A rule refers to its row with the text "(See Decisions)". This text is in the data holder file 228 times and in the Register file 71 times. The text has no link and does not name the row.
- The rows are in the sequence in which we added them, not in the sequence of the rules.
- Many reasons are in two places: in the rule, and again in a row.
- In the data holder file, 63 % of the bullets have two or more sentences. Many bullets have more than one rule. Thus a test or a ticket cannot refer to one rule.
- Approximately 25 rules are in both service files, and some of the copies are already different. The code of most of these rules is in `pkg/` libraries.
- The data holder file has six open questions. Three of them already have an answer in the file.
- The files also contain text that is not a requirement: copied CDS text, implementation detail, test plans and history.
- An agent cannot read the data holder file in one step. Language models find information less reliably in a long context, and least reliably in the middle of it ([Liu et al., Lost in the Middle](https://arxiv.org/abs/2307.03172)).

Other specifications and guides use these methods:

- **Stable identifiers.** FAPI 1.0 keeps the number of a removed clause and marks it "withdrawn" ([FAPI 1.0 Part 2](https://openid.net/specs/openid-financial-api-part-2-1_0.html)). The Rust Reference gives a name to each rule, and tests refer to that name ([Rust Reference authoring guide](https://rust.googlesource.com/rust-lang/reference/+/refs/heads/beta/docs/authoring.md)). SQLite marks each requirement and connects its tests to it ([SQLite requirements](https://www.sqlite.org/requirements.html)).
- **One requirement in each statement.** The reason is a separate attribute ([INCOSE Guide to Writing Requirements](https://www.incose.org/group/requirements-working-group/)). The EARS patterns give a fixed sequence: the condition first, then the response ([EARS](https://alistairmavin.com/ears/)).
- **Reference text only describes.** The reasons and the history are in a different place ([Diátaxis: Reference](https://diataxis.fr/reference/)). A document links to a source and does not copy it ([Google documentation best practices](https://google.github.io/styleguide/docguide/best_practices.html)).
- **No open questions in an accepted document.** A Rust RFC must resolve its unresolved questions before the feature becomes stable ([Rust RFC template](https://github.com/rust-lang/rfcs/blob/master/0000-template.md)).
- **A context diagram.** The C4 system context diagram shows a system, its users and the systems that it connects to ([C4 model: System context diagram](https://c4model.com/diagrams/system-context)).
- **Small documents.** People do not read or update large documents ([Nygard](https://www.cognitect.com/blog/2011/11/15/documenting-architecture-decisions)).

We compared two structures:

1. **One file for each component, with the new rules below.** A reader can search all of the file. An author does not have to find the correct file. By estimate, the data holder file is still approximately 20,000 words after the changes.
2. **One file while the component is small, then an overview and one file for each area.** A reader starts with a short overview. An agent reads only the area that it changes. But an author must find the correct file, and a rule can go into the wrong file.

## Decision

### Scope

This decision applies to each specification in the repository:

- a service
- a `pkg/` library
- the PKI
- the conformance suite, which gives the rules for its tests
- a web package, for example a design system, which gives the rules that each UI obeys

It does not apply to ADRs, research notes or READMEs.

### Content

- A specification states what the component does. A test or a reviewer can check each statement without the code.
- The specification of a library or a design system also states what its users must do.
- A specification has no implementation detail, no test plans and no history. Git and the pull requests keep the history.
- A specification does not copy the text of a source. It states only what the component selects, limits or adds. For the rest, it links to the source.
- A specification has no open questions. A question with no answer is a ticket.

Reason: each fact has one place, and the reader sees only the choices of the lab.

### Requirements

- Each requirement is one list item or one table row, with one rule.
- A requirement is in the present tense, for example "The page shows …".
- The condition comes first and the response comes last, in the sequence of the EARS patterns.
- Each requirement has a source. The source is a link to the CDS, to ACCC guidance or to a different standard, or a link to a judgement.

Reason: a test, a ticket and a judgement can refer to one rule. The reader finds the condition of each rule in the same position.

### Judgements

- A judgement records one of these:
  - a choice that the sources do not make
  - a conflict between two sources, and the source that we follow
  - a source requirement that the component does not meet
- Each judgement gives its basis in the sources process, and its reason in three sentences or fewer.
- A judgement is in the same file as its requirements, after them.
- The reason is in the judgement only. The requirement links to the judgement.
- A choice with no alternative is not a judgement.
- We use the term "judgement", not "decision". Thus a judgement is not confused with an architecture decision.

Reason: the reason is near its rule, and it has one place.

### Identifiers

- Each requirement and each judgement has an identifier: a code for the component and a number, for example `DH-0042`.
- An identifier never changes, and we never use it again.
- A rule that moves to a different area in the same component keeps its identifier. The file shows the area.
- A rule that moves to a different component gets a new identifier. This occurs when its code moves to `pkg/`.
- The withdrawn list of the file keeps each removed identifier. For a rule that moved, the list links to its new identifier.
- Identifiers are for links and searches only. Text refers to a rule with a link that has descriptive text, not with the identifier alone.
- A bare identifier is only for a machine, for example in the `// Verifies:` comment of a test.

Reason: a reader cannot understand an identifier, but an agent can search for it. A link to a stable identifier does not break when a heading changes. An identifier with no area cannot become incorrect when a rule moves in its component.

### Structure

- A specification starts as one file, `DESIGN.md`.
- When the file has more than 5,000 words, it becomes a `design/` directory. The limit is a judgement. It keeps each file well below the size that an agent can read in one step.
- The `README.md` of the directory is the overview. It has:
  - a summary
  - a context diagram: the component, its users and the external systems that it connects to (C4 level 1), with no internal parts
  - a table of the areas and their files
  - a table of the `pkg/` specifications that apply to the component
  - the goals and the non-goals
- Each area has its own file, with a summary, the requirements and the judgements.
- An area is one interface of the component, for example an API, a web UI or an operator interface. Or it is one capability that has its own section in the primary source, for example consent. An area uses the name of the CDS section when there is one.
- A lab-only area, for example an operator console, has its own file.
- The table of areas is the full list. A new area adds a row to it.
- A change that moves a requirement also updates each link to it. This includes the links in ADRs. A link update does not change the decision of an ADR.

Reason: a reader gets the overview of a large component on one page. An agent reads only the overview and one area.

### Shared rules

- A rule that the code of a `pkg/` library implements is in the specification of that library.
- The overview of a service lists the library specifications that apply to it. The service specification gives only its differences from them.
- When code moves to `pkg/`, its requirements move in the same change.

Reason: a shared rule has one place, and it stays with its code. The list tells a reader which rules apply, without the code.

### Conventions and terms

- `docs/spec/conventions.md` gives the detailed rules for how to write a specification. Examples are the template, the format of identifiers, links and diagrams, the words to avoid, and the form of non-functional requirements.
- These rules can change without a new ADR, if they obey this decision.
- `docs/spec/glossary.md` gives one term for each concept. Specifications use only that term.
- `docs/spec/` contains only these two documents.

### Checks

A check in `mise run lint` makes sure that:

- each identifier is unique and has the correct format
- each requirement has a source
- each link to a requirement or a judgement goes to an identifier that exists
- no text refers to an identifier without a link
- each file is in the word limit
- the table of library specifications agrees with the `pkg/` imports of the component
- each `// Verifies:` comment in a test refers to an identifier that exists

The check gives a warning, not a failure, for a requirement that no test verifies. It applies only to files that use this structure.

Reason: a tool keeps a rule with no exceptions more reliably than a written rule.

## Consequences

- A reader finds the reason for a rule in the same file. The Decisions tables become much smaller, because the copied reasons and the rows with no alternative go.
- Tests, tickets and pull requests can refer to one rule. A requirement that no test verifies shows as a warning.
- More files make some changes more difficult. An author must find the correct file, and a related rule can be in a different file. The overview, the links and the check reduce this risk, but do not remove it.
- An author writes more structure for each rule: an identifier, a source and sometimes a judgement. Small changes take more time than before.
- The change of the current files is large: approximately 950 bullets in the two service files. Until a file changes, the check does not apply to it. Thus the old and the new structure exist together for a time.
- A rule that two services have, but that no library implements, stays in each service specification until its code moves to `pkg/`.
- The open questions in the current files become judgements or tickets.
- A problem type URI of a console API links to a heading of a `DESIGN.md`. This URI changes when the file moves to the new structure.
- When a specification moves, the ADRs that link to it change. Only their links change.
- The check does not find a rule in the wrong area, a reason that a requirement repeats, or a term that is not in the glossary. Review must find these.
