# 0024. Write specifications as identified single requirements, with each judgement next to its rules, and split them by area

- Date: 2026-11-15

## Context

This ADR amends [ADR 0013](0013-hexagonal-architecture-with-an-internal-composition-root.md). It replaces the `DESIGN.md` line of the service layout in ADR 0013. The other parts of ADR 0013 do not change.

Each component has a `DESIGN.md` that gives its requirements, with no implementation detail. Each requirement links to its primary source, or has "Judgement" and a reason. A Decisions table at the end of the file records the choices that the sources do not make. People and agents read and change these files.

The repository has five `DESIGN.md` files: two services, the PKI, the conformance suite and the design-system package of the web UIs. On 2026-11-15, a review of the five files found these facts:

- The data holder file has 53,510 words and the Register file has 26,266 words. The Decisions tables have approximately 280 and 175 rows, which are 39 % and 44 % of the words of the two files.
- A rule refers to its row with the text "(See Decisions)" (228 times in the data holder file and 71 times in the Register file). The text has no link and does not name the row. The rows are in the sequence in which they were added, not in the sequence of the rules.
- Many reasons are in two places: as "Judgement" text in the rule, and again in a row.
- 63 % of the bullets of the data holder file have two or more sentences, and many bullets have more than one rule. A test, a ticket or a different rule cannot refer to one rule.
- Approximately 25 rules are in both service files, and some copies are already different. The code of most of these rules is in `pkg/` libraries.
- The files also contain text that is not a requirement: copied CDS text, implementation detail, test plans and history.
- The data holder file is too large for an agent to read in one step. Language models find information less reliably as the context gets longer, and least reliably in the middle of the context ([Liu et al., Lost in the Middle](https://arxiv.org/abs/2307.03172)).

Other specifications and guides use these methods:

- Stable identifiers that are never used again. FAPI 1.0 keeps the number of a removed clause and marks it "withdrawn" ([FAPI 1.0 Part 2](https://openid.net/specs/openid-financial-api-part-2-1_0.html)). The Rust Reference gives a name to each rule, and tests refer to that name ([Rust Reference authoring guide](https://rust.googlesource.com/rust-lang/reference/+/refs/heads/beta/docs/authoring.md)). SQLite marks each requirement and connects tests to it ([SQLite requirements](https://www.sqlite.org/requirements.html)).
- One requirement in each statement, with the reason as a separate attribute ([INCOSE Guide to Writing Requirements](https://www.incose.org/group/requirements-working-group/)). The EARS patterns give a fixed sequence for the condition and the response of a requirement ([EARS](https://alistairmavin.com/ears/)).
- Reference text that only describes, with the reasons and the history in a different place ([Diátaxis: Reference](https://diataxis.fr/reference/)). Links in place of copied text ([Google documentation best practices](https://google.github.io/styleguide/docguide/best_practices.html)).
- Small documents. Large documents are not kept up to date and are not read ([Nygard](https://www.cognitect.com/blog/2011/11/15/documenting-architecture-decisions)).

We compared two options for the structure:

1. One file for each component, with the new rules below. A reader can search all of it, and an author need not find the correct file. The data holder file stays approximately 20,000 words after the changes, by estimate.
2. One file for each component while it is small, and an overview file plus one file for each area when it is large. A reader starts with a short overview, and an agent reads only the area that it changes. An author must find the correct file, and a rule can go into the wrong file.

## Decision

This decision applies to each specification in the repository: the `DESIGN.md` of a service, a `pkg/` library, the PKI, the conformance suite and a web package. A specification of the conformance suite gives the rules for its tests, and a specification of a design system gives the rules that each UI obeys. It does not apply to ADRs, research notes or READMEs.

**Content.** A specification describes what the component does now, or what it requires of its users, as a test or a reviewer can see it without the code. It contains no implementation detail, no test plans and no history. Git and the pull requests keep the history. It does not copy the text of a source. It states only what the component selects, limits or adds, and it links to the source for the rest. Reason: each fact then has one place, and the reader sees the choices of the lab.

**Requirements.** Each requirement is one list item or one table row, with one rule. It is in the present tense ("The page shows …"), with the condition first and the response last, in the sequence of the EARS patterns. A requirement has a source: a link to the CDS, to ACCC guidance or to a different standard, or a link to a judgement. Reason: a test, a ticket and a judgement can then refer to one rule, and a reader finds the condition of each rule in the same place.

**Judgements.** A judgement records a choice that a source does not make, or a conflict between sources and the source that wins. Each judgement gives its basis, from the sources process, and its reason in three sentences or fewer. A judgement is in the same file as its requirements, after them. The reason is written in the judgement only, and not in the requirement. We use "judgement", not "decision", so that the term is different from an architecture decision. A choice that has no alternative is not a judgement. Reason: the reason is then near its rule, and it has one place.

**Identifiers.** Each requirement and each judgement has an identifier: the component, the area and a number, for example `DH-REG-004`. An identifier never changes, also when its rule moves to a different area, and it is never used again. A removed requirement keeps its identifier in a list of withdrawn identifiers in its file. Identifiers are for links and searches only. Text refers to a rule with a link that has descriptive text, and never with an identifier alone. Only a machine reads a bare identifier, for example in the `// Verifies:` comment of a test. Reason: a reader cannot understand an identifier, but an agent can search for it, and a link with a stable target does not break when a heading changes.

**Structure.** A specification starts as one file, `DESIGN.md`. When it has more than 5,000 words, it becomes a `design/` directory. Its `README.md` gives the overview: a summary, a context diagram, a table of the areas with their files, the goals and the non-goals, and the open questions. Each area has its own file, with a summary, the requirements and the judgements. Lab-only areas, for example an operator console, have their own files. Reason: a reader gets the overview of a large component in one page, and an agent reads only the overview and one area. The limit of 5,000 words is a judgement. It keeps each file well below the size that an agent can read in one step. A change that moves a specification or a requirement also updates each link to it, also in ADRs. A link update does not change the decision of an ADR.

**Shared rules.** A rule that the code of a `pkg/` library implements is in the specification of that library, next to its code. A service that uses the library refers to that specification and gives only its differences. When code moves to `pkg/`, its requirements move in the same change. Reason: a shared rule then has one place, and it stays with its code.

**Conventions and terms.** The detailed rules for how to write a specification are in `docs/spec/conventions.md`, not in this ADR: the template, the format of identifiers and links, the words to avoid, and the form of non-functional requirements. These rules can change without a new ADR, if they obey this decision. A glossary in `docs/spec/glossary.md` gives one term for each concept, and specifications use only that term. `docs/spec/` contains only these two kinds of document.

**Checks.** A check in `mise run lint` makes sure that each identifier is unique and correct, that each requirement has a source, that each link to a judgement or a requirement goes to an identifier that exists, that no text refers to an identifier with no link, and that each file is in the word limit. It also checks the `// Verifies:` comments in tests, and it gives a warning, not a failure, for a requirement that no test verifies. The check applies only to files that use this structure. Reason: rules with no exception are kept by a tool more reliably than by a written rule.

## Consequences

- A reader finds the reason for a rule in the same file. The Decisions tables of the services become much smaller, because the copied reasons and the rows with no alternative go.
- Tests, tickets and pull requests can refer to one rule. A requirement that no test verifies shows as a warning.
- More files make each change more difficult in one way: an author must find the correct file, and a related rule can be in a different file. The overview, the links and the check reduce this risk, but do not remove it.
- An author writes more structure for each rule: an identifier, a source and sometimes a judgement. Small changes take more time than before.
- The change of the current files is large: approximately 950 bullets in the two service files. Until a file is changed, the check does not apply to it, and the two structures exist together.
- A rule that two services have, but that no library implements, stays in each service specification until its code moves to `pkg/`.
- A problem type URI of a console API that links to a heading of a `DESIGN.md` changes when the file moves to its new structure.
- ADRs that link to a specification change when the specification moves. Only their links change.
- The check does not find a rule in the wrong area, a reason that a requirement repeats, or a term that is not in the glossary. Review must find these.
