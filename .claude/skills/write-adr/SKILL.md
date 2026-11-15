---
name: write-adr
description: >
  Write a new architecture decision record (ADR) in docs/adr/, or review a draft ADR. Use any time you are asked to
  create an ADR, to record a decision, or to move rules or conventions from another document into an ADR. Also use
  before you edit an existing ADR.
---

An ADR records one decision: the context that made it necessary, what we decided, and what results from it. It is written for a future reader who was not there. If you do nothing else, make sure that reader can understand why we made the decision and what it costs.

## Before you write

Decide if the decision needs an ADR. Write one when the decision:

- affects structure, dependencies, interfaces, construction techniques or non-functional characteristics, or
- is difficult or expensive to reverse, or
- sets a pattern that more than one component must follow, or
- had two or more real options.

Do not write one for a change that a commit message explains, or to describe what a component does. That detail goes in a `DESIGN.md` or stays in the code.

An ADR can record a decision that was made informally in the past. Write it as the formal decision. Do not give the history of when or where the decision was first made.

If the facts are not sufficient to write the context or the reasons, ask the user. Do not invent a reason. A reason that you reconstruct must be named as such when you report your work.

## Process

1. Read all the ADRs in `docs/adr/`. Use their terms and do not contradict them.
2. Copy `docs/adr/0000-adr-template.md` to `docs/adr/NNNN-short-title.md`, with the next number that is not used. Numbers are never used again.
3. Write the three sections: Context, Decision, Consequences. Do not add a status field. On `main` means accepted. Anything else is a draft.
4. Check the draft against the rules and the anti-patterns below.
5. Check each fact, link and version in the draft against its source. Each reference must exist.
6. If the new ADR changes or adds to an earlier ADR, decide whether it amends or supersedes the earlier ADR:
   - **Amend** when the change is minor: it changes or adds to one part, and most of the earlier decision stays true.
   - **Supersede** when the change is significant: it reverses the main decision, or little of the earlier decision stays true, or the earlier ADR already has two amendments. The new ADR then states the complete decision, also the parts that stay the same, so that a reader needs only the new ADR.
7. Say so in the Context of the new ADR, with a link to the earlier one. For an amendment, name the statements of the earlier ADR that the new ADR replaces, if any. Then add one of these lines to the earlier ADR, directly below its date and below the lines of earlier amendments. Make no other change to the earlier ADR.
   - `- Amended by [ADR-NNNN](NNNN-short-title.md): <the part that changed>`
   - `- Superseded by [ADR-NNNN](NNNN-short-title.md)`

   A superseded ADR keeps its `Amended by` lines.

## Sections

**Title.** `# NNNN. <decision as an instruction>`, for example "Use PostgreSQL, with goose for schema changes". The title states the decision, not the problem.

**Date.** `- Date: YYYY-MM-DD`, the date of the ADR.

**Context.** The problem, and the facts and forces that make a decision necessary.

- Use neutral language: facts, not arguments for the option that we selected.
- Give only the context that drives this decision. Leave out context that the decision does not depend on, for example CDR-specific background, the component types or the lab deployment, when the decision is orthogonal to them.
- If options were compared, give each real option here, with its advantages, its disadvantages and the evidence. Link each fact directly to its primary source.

**Decision.** What we do, and why.

- Full sentences, active voice.
- State the scope of the decision here: what it applies to, and what it does not apply to.
- Record the decision and its reasons, not how the code implements it.
- Cover the main points. Leave out low-priority detail. Do not make the common case into a rule: if one binary for each service is usual but not required, do not require it.
- Do not copy a list that configuration already contains (for example the valid commit scopes). State the rule that the list must obey.
- Refer to kinds of components and where they go, not to specific current or future components. Leave a kind that is not settled for a later ADR.
- Leave adjacent topics out. They get their own ADRs.

**Consequences.** What results from the decision: what becomes easier, what becomes more difficult, and what we must now do or accept.

- Include the negative consequences, and the effects that come later.
- Include the effects on the people who operate and maintain the result.
- A consequence is a result. It is not a new rule (that goes in Decision), and it is not an instruction to a person. Write "The git hook is optional. Without it, the first check is in CI", not "Each person must install the git hook".
- A consequence helps the reader understand the effects of the decision. It does not add scope or requirements to the decision. If a consequence adds work or a requirement, put that in the Decision, or write the consequence again as an effect.

## Rules for this repository

- One decision for each ADR. One or two pages.
- Never reference or link `README.md`. It is not a primary source.
- An ADR can reference only ADRs with a lower number. A past decision cannot depend on a future one. The `Amended by` and `Superseded by` lines are the only exceptions.
- An ADR is not a living document and not a design document. Do not add API descriptions or implementation notes to an ADR after it is on `main`.
- A change that moves the target of a link in an ADR updates the link. A link update does not change the decision, so it needs no amendment.
- Do not use a new convention (a new section, a status field, an ID scheme, a link format) before the user approves it.
- Write in ASD-STE100 Simplified Technical English, with inclusive language.
- The repo `README.md` _may_ link to an ADR in this form: `See [ADR-0004](docs/adr/0004-module-layout.md).`

## Anti-patterns

| Anti-pattern | How to recognise it | Remedy |
|---|---|---|
| Fairy tale | Only advantages. The justification is shallow. | Give the disadvantages of the selected option too. |
| Sales pitch | Adjectives and claims with no evidence. | Remove them, or give the evidence and its source. |
| Free lunch | No consequences, or only harmless ones. | State the costs, and the consequences that come later. |
| Dummy alternative | An option that was never viable, to make the selected option look better. | Give only the options that were real. If there was only one, say so. |
| Sprint | One option, and only short-term effects. | Find the real alternatives. Give the mid-term and long-term effects. |
| Tunnel vision | Only the local context. Operators, maintainers and users are ignored. | Give the effects on each group that the decision touches. |
| Maze | The content does not agree with the title, or goes off the topic. | Move the material that is off the topic out, or to a different ADR. |
| Blueprint or policy in disguise | A cookbook of commands with no reasons. | Give the context and the reason for each rule. Remove the rules that have none. |
| Mega-ADR | Many pages, with a detailed design. | Move the design to a `DESIGN.md` or leave it in the code. |
| Implementation detail | Describes how the code works. | Record why, not how. |
| Glorified change log | An ADR for each change. | Use the commit message. |
| Living design document | An ADR on `main` that changes when the design changes. | Leave it. A changed decision needs a new ADR that amends or supersedes it. |
| Missing rationale | A decision with no reason. | Give the context and the reason. Ask the user if you do not know them. |
| Pseudo-accuracy | Weighted scores or numbers that look exact but have no source. | Remove them, or give the measurement and how it was made. |
| Invented rationale | Reasons that the author guessed, or references that do not exist. | Check each fact and each link. Ask the user for the reasons. |

## Report

When you report the ADR to the user, list in a separate section:

- Each reason in the Context that you reconstructed and the user did not give.
- Each place where the ADR does not agree with the current code or with a different document.
- Each decision that the user must make.

## Sources

The section guidance and the anti-patterns come from these sources:

- Michael Nygard, [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions), 2011. The Context, Decision, Consequences structure, and the rule to give all consequences.
- Olaf Zimmermann, [How to create ADRs, and how not to](https://ozimmer.ch/practices/2023/04/03/ADRCreation.html), 2023. The anti-patterns from Fairy tale to Mega-ADR, and Pseudo-accuracy.
- Olaf Zimmermann, [How to review ADRs, and how not to](https://ozimmer.ch/practices/2023/04/05/ADRReview.html), 2023. The questions that a reviewer asks.
- Government Digital Service, [The GDS Way: Documenting architecture decisions](https://gds-way.digital.cabinet-office.gov.uk/standards/architecture-decisions.html). An ADR on `main` is accepted, the title states the decision, and a superseded ADR has a mark with a link.
- Amazon Web Services, [AWS Prescriptive Guidance: ADR process](https://docs.aws.amazon.com/prescriptive-guidance/latest/architectural-decision-records/adr-process.html) and [Best practices](https://docs.aws.amazon.com/prescriptive-guidance/latest/architectural-decision-records/best-practices.html). Record the reason and not the implementation, and do not change an accepted ADR.
- Microsoft, [Azure Well-Architected Framework: Maintain an architecture decision record](https://learn.microsoft.com/en-us/azure/well-architected/architect-role/architecture-decision-record). When a decision is significant, the Missing rationale anti-pattern, and the rule that an ADR is not a design guide.
- Spotify Engineering, [When Should I Write an Architecture Decision Record](https://engineering.atspotify.com/2020/04/when-should-i-write-an-architecture-decision-record), 2020. An ADR can record a decision that was made in the past.
- Red Hat, [Architecture decision records](https://www.redhat.com/en/blog/architecture-decision-records). The Glorified change log anti-pattern, and alternatives in the Context.
- Equal Experts, [Accelerating Architectural Decision Records (ADRs) with Generative AI](https://www.equalexperts.com/blog/our-thinking/accelerating-architectural-decision-records-adrs-with-generative-ai/). The Invented rationale anti-pattern.
- Nat Pryce, [adr-tools](https://github.com/npryce/adr-tools), `adr new` help text ([source](https://github.com/npryce/adr-tools/blob/master/src/adr-new)). A new ADR can link to an earlier ADR with a link and a reverse link, for example "Amends" and "Amended by".
- Netresearch, [nr-llm: Architecture Decision Records](https://docs.typo3.org/p/netresearch/nr-llm/0.32/en-us/Adr/Index.html). An amended ADR stays current, and its mark names the part that changed.
