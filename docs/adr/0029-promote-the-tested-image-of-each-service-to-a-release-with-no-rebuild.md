# 0029. Promote the tested image of each service to a release, with no rebuild

- Date: 2026-11-23
- Amended by [ADR-0035](0035-give-the-lab-one-cli-cdrlab-in-cli-with-an-image-that-follows-the-rules-of-a-service-image.md): the image of the lab CLI

## Context

A release gives one version to the artifacts of one commit of `main` ([ADR 0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md)). The images are the artifacts that run, so a team needs to know that the image of a release is the image that CI tested.

These facts constrain the images of a release:

- CI builds the image of a service only when an input of that image changes. The inputs are a list of paths in the path filter of CI. CI tests the image in the lab jobs, and publishes it from `main` only after its gate jobs pass, with the full and the short commit ID as tags. Thus most commits of `main` have no image of each service.
- The build sets the commit in the binary, and `service.version` in telemetry is the short commit ([ADR 0006](0006-telemetry.md)). An image that CI built before the release cannot contain the version of the release.
- A new tag on an existing manifest in a registry keeps its digest. Signatures and attestations refer to a digest, so they stay valid for the new tag.
- The charts require `image.tag`, and they do not use the `appVersion` of the chart. Thus the `appVersion` of a chart does not select an image.

The real options were:

1. **Rebuild the images at the release, with the version in the binary.** Telemetry then gives the version. But the released image is not the tested image, and a base image or a dependency can change between the two builds.
2. **Release the image of the release commit only.** The release then needs an image for each service at each release commit, but CI builds an image only when an input changes.
3. **Release the last tested image of each service.** For each service, the image of the last commit that had an image build is the image that a build of the release commit gives, because CI skipped the build only when no input of the image changed. A failed or cancelled publish leaves an older image as the last image, so the release needs checks.

## Decision

A release uses, for each service, the image that CI tested and published from `main`. The release builds no image.

- The **image commit** of a service is the first commit with an image of that service, in the first-parent history of `main` from the release commit back.
- The release refuses a commit whose CI run on `main` did not pass.
- For each service, the release refuses when the image commit is not the last commit, at or before the release commit, that changes an input of the image. This check stops the release of an old image after a failed or cancelled publish.
- One file in the repository gives the inputs of the image of each service. The path filter of CI and the release read that file, so that the two lists cannot differ.
- The release adds the tag `vX.Y.Z` to the digest of the image commit. The digest does not change.
- The release notes give the image commit of each service, because the services can have different image commits.
- `service.version` stays the short commit of the build. The release notes and the image tags map the version to the commit.

The charts select the image of their own version:

- The default of `image.tag` is the `appVersion` of the chart. A value of `image.tag` still overrides it. This replaces the chart rule that `image.tag` is required.
- A chart of a release has the `appVersion` `vX.Y.Z`, the tag that the release puts on the images.
- A chart from `main` has the full commit ID of the image commit of its service as its `appVersion`, from the same rule. Thus the default also gives an image that exists.

## Consequences

- The digest of each released image is the digest that CI tested. A team can compare the digest with the publication from `main`.
- Telemetry gives the commit, not the version. An operator uses the release notes to find the version of a commit.
- A release is not possible when the last publish of an image failed, until CI on `main` publishes that image. A run again of the failed workflow, or a new commit that changes an input of the image, does this.
- A commit that starts no CI run, for example a change to `renovate.json` only, cannot be a release commit. The release of the commit before it is possible.
- A change to the CI workflow or to `mise.toml` is an input of each image, so it gives a new image of each service.
- The chart from `main` must wait for the image of its own commit, when the commit changes an input of the image.
- A team that installs a chart with no `image.tag` gets the image that belongs to the chart.
