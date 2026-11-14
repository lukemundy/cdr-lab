# 0007. Use Conventional Commits for commit messages

- Date: 2026-10-04

## Context

People, agents and automated tools write the commits of this repository. Without a rule, each one writes messages in a different form, and the history is hard to read and to search.

The repository has many components ([ADR 0004](0004-module-layout.md)). A reader of the history wants to know which part of the repository a commit changes, and if the change is a breaking change.

A rule that only a document gives is not obeyed reliably. A tool must check it.

## Decision

Commit messages obey [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/#specification):

```
<type>[(<scope>)][!]: <description>

[body]

[footers]
```

- Types: `feat` and `fix` (from the specification), and `build`, `chore`, `ci`, `docs`, `perf`, `refactor`, `style`, `test`, `revert`.
- Scopes: each valid scope must be defined in `cog.toml`. The checks reject a scope that is not defined. A change to the full repository has no scope.
- Breaking changes: add `!` before the colon, or a `BREAKING CHANGE:` footer.
- [Cocogitto](https://docs.cocogitto.io) (`cog`) checks the messages, in three places:
  - A `commit-msg` git hook rejects a message that does not obey the rules. `mise run hooks` installs it.
  - CI checks all the commits of the branch again.
  - CI checks the title of each pull request, because the title becomes the commit message on `main` when the pull request merges as one commit.
- The checks ignore merge commits.

## Consequences

- The git hook is optional. For a person who does not install it, the first check of a commit message is in CI.
- A new scope needs a change to `cog.toml` before its first use.
- A message that does not obey the rules, in any commit of a branch, fails CI. The fix is to rewrite the commit.
- The history can give a changelog or a version number later. No tool does this now.
