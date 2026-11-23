# 0027. Release all artifacts with one version, from a manual workflow on a commit of main

- Date: 2026-11-23

## Context

This ADR amends [ADR-0007](0007-conventional-commits.md). It replaces the consequence "No tool does this now": cocogitto gives the version and the changelog of a release.

CI publishes the image of each service and the Helm charts from `main`, with commit IDs as tags and the chart version `0.0.0-g<short commit>`. A team that keeps a deployment of the lab in its own cluster cannot tell from these versions what changed between two artifacts, or whether an upgrade is safe.

These facts constrain a release:

- The repository is one Go module with many components ([ADR 0004](0004-module-layout.md)). The two charts deploy the two services, and they are used together.
- The ruleset of `main` requires a pull request for each change and has no bypass actors. Thus a workflow cannot push a commit to `main`.
- Pull requests merge as one commit with the title of the pull request as the message and no body. Thus only the type, the scope and the `!` of the title reach `main`. A footer of a branch commit does not.
- cocogitto (`cog`) already checks the commit messages ([ADR 0007](0007-conventional-commits.md)), and `mise.toml` pins it. `cog bump --auto --dry-run` gives the next version from the tags and the commits, with no commit and no file change. `cog changelog <range>` gives the notes of a range before its tag exists. cog sees a `v` tag only when `cog.toml` has `tag_prefix = "v"` ([cocogitto: configuration](https://docs.cocogitto.io/reference/config.html)). While the major version is 0, cog gives a minor bump for a breaking change, and no setting changes this.

The real options were:

1. **One version for each component**, with the monorepo mode of cog. A change to one service changes only its version. A team must then find a set of versions that work together, and the repository has no test of such a set other than the commit that CI tested.
2. **One version for all artifacts.** A version identifies one commit of `main`, and all its artifacts agree. A change to one service also gives a new version of the other artifacts.

For the trigger, a release at each merge gives many versions with small changes, and a person cannot select the commit. A manual release gives versions only when a person decides that a set of changes is ready.

For the tool, git-cliff gives a configurable bump rule, but it is a second parser and a second configuration next to cog. release-please updates `CHANGELOG.md` and version files through a pull request, which is a commit to `main` for each release. semantic-release is a Node tool with npm plugins outside the pnpm workspace and its controls ([ADR 0015](0015-web-toolchain-and-npm-supply-chain-controls.md)).

For the changelog, a committed `CHANGELOG.md` needs a commit to `main` at each release. cog can make the full changelog again from the tags at any time, so the file is a copy of the tags.

## Decision

A release gives one version, `vX.Y.Z`, to all artifacts of one commit of `main`. This applies to each artifact that the repository publishes as a release. It does not apply to the publications from `main`, which keep their commit versions.

- A person starts a release by hand, with a `workflow_dispatch` workflow, for one commit of `main`. CI never starts a release.
- The release pushes no commit to any branch. The git tag `vX.Y.Z` on the release commit is the only change to the repository.
- cog in tag-only mode gives the version, from the Conventional Commits since the earlier version tag. The release never runs `cog bump` without `--dry-run`. We use cog because it already checks the commits, it needs no commit, and its bump rules are the Conventional Commits rules.
- The first version is `v0.1.0`. While the major version is 0, a breaking change gives a minor bump.
- The release makes no promise about the data of a deployment while the major version is 0. A release that needs a fresh deployment says so in its notes.
- The changelog is only in the notes of the GitHub Release. cog makes it from the Conventional Commits since the earlier version tag. The repository has no changelog file.
- The squash-merge settings stay as they are. Thus the changelog has one line for each pull request, from its title.

## Consequences

- The version tells a team that two artifacts belong together. A team pins one version for all charts and images.
- A change to one service gives a new version of all artifacts, also of the artifacts that did not change.
- The title of each pull request is the input of the changelog and of the version. A title with the wrong type or no `!` gives the wrong bump, and the fix is in the next release, because the commit on `main` cannot change.
- The repository has no offline changelog. A reader uses the GitHub Releases, or runs `cog changelog` on a clone.
- Before 1.0.0, a team can lose its data at an upgrade. The notes tell the team before the upgrade.
- The move to 1.0.0 needs a person to give the major version. The automatic bump never gives it while the major version is 0.
