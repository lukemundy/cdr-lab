# 0037. Calculate the release version in the release tool, and not with cog bump

- Date: 2026-11-23

## Context

This ADR amends [ADR-0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md). It replaces these statements of ADR-0027:

- In the Context: "While the major version is 0, cog gives a minor bump for a breaking change, and no setting changes this."
- In the Decision: "cog in tag-only mode gives the version, from the Conventional Commits since the earlier version tag. The release never runs `cog bump` without `--dry-run`. We use cog because it already checks the commits, it needs no commit, and its bump rules are the Conventional Commits rules."

The release specification requires the next minor version for a breaking change while the major version is 0, and the next major version after that ([release: start and version](../../tools/release/DESIGN.md#rls-0005)). A team that uses a patch range on 0.x, for example `~0.3.0`, then gets no breaking change.

cog 7.0.0, the version that `mise.toml` pins, does not obey this rule. In a test repository, with `tag_prefix = "v"`:

- From `v0.1.0`, the commit `fix!: x` gives `v0.1.1`. A `fix` commit with the footer `BREAKING CHANGE:` also gives `v0.1.1`.
- From `v0.1.0`, the commit `feat(api)!: x` gives `v0.2.0`, which is the bump of a `feat` commit with no breaking change.
- From `v1.0.0`, the commit `fix!: y` gives `v2.0.0`.

Thus, while the major version is 0, cog ignores the breaking mark. Its configuration has no setting that changes this ([cocogitto: configuration](https://docs.cocogitto.io/reference/config.html)).

The real options were:

1. **Accept the bump of cog** and change the rule of the specification. A breaking `fix` then gives a patch version, and a team that uses a patch range gets a breaking change.
2. **Calculate the version in the release tool** from the commit messages after the earlier version tag. The rule needs only the type, the `!` of the header and the `BREAKING CHANGE` footer. The tool then has a second parser of the header, next to cog.
3. **Use a different tool**, for example git-cliff, which has a configurable bump rule. ADR-0027 did not select it, because it is a second parser and a second configuration next to cog.

## Decision

The release tool calculates the version of a release from the Conventional Commits in the first-parent history after the earlier version tag, up to the release commit, with the rules of the release specification. The release does not use `cog bump`.

cog stays the tool that checks the commit messages ([ADR-0007](0007-conventional-commits.md)), and it gives the changelog of the release notes from the same range. Thus only the bump rule moves to the release tool.

We select option 2, because it obeys the specification with the smallest change: a small part of the tool, with its tests, that already does the checks of the release.

## Consequences

- The bump rule is in the code of the release tool, with a test for each case. A change of the cog bump rule in a later version of cog does not change the version of a release.
- The release tool and cog parse the header of a commit separately. A header that cog accepts and the tool does not parse gives no bump from that commit. The git hook and the check of the pull request title use cog, so a header on `main` has the form that both parse.
- A breaking change while the major version is 0 gives the next minor version, as ADR-0027 intended.
- The move to 1.0.0 still needs a person, as ADR-0027 says.
