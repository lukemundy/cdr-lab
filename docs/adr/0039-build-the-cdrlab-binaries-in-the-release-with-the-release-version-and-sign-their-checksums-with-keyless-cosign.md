# 0039. Build the cdrlab binaries in the release, with the release version, and sign their checksums with keyless cosign

- Date: 2026-11-23

## Context

This ADR amends [ADR-0031](0031-sign-releases-with-keyless-cosign-and-attest-provenance-and-an-sbom.md). It adds the `cdrlab` binaries to the artifacts that the release signs and attests. The rules of ADR-0031 for the images and the charts do not change.

A person on a host uses a release binary of `cdrlab` or `go install` ([ADR 0035](0035-give-the-lab-one-cli-cdrlab-in-cli-with-an-image-that-follows-the-rules-of-a-service-image.md)). `cdrlab up` in the release mode pulls the images that have the version of the binary as the tag ([ADR 0038](0038-run-the-compose-lab-with-cdrlab-from-a-compose-project-in-the-binary-and-issue-its-cdr-certificates-from-the-host.md)). Thus the binary must know the version of its release.

These facts constrain the binaries:

- The release promotes the tested images with no rebuild, and `service.version` of an image stays the short commit ([ADR 0029](0029-promote-the-tested-image-of-each-service-to-a-release-with-no-rebuild.md)). An image of `main` cannot contain the version, because the build is before the tag.
- CI on `main` builds no host binaries. Thus no tested binary exists that a release can promote.
- A binary that contains third-party code needs the licence notices of that code ([ADR 0018](0018-container-images-with-a-web-ui-build-stage.md)).
- The repository is one Go module at the root, so the root tag `vX.Y.Z` is also the module version. `go install <package>@vX.Y.Z` works only when `go.mod` has no `replace` and no `exclude` directive ([Go: go install](https://go.dev/ref/mod#go-install)). A `go install` sets no `-X` flags. It gives the module version in the build information of the binary ([Go: runtime/debug.ReadBuildInfo](https://pkg.go.dev/runtime/debug#ReadBuildInfo)).
- GoReleaser makes archives, checksums and signatures from one configuration ([GoReleaser](https://goreleaser.com/)). It is a new tool and a new configuration next to the release tool. The binaries are one program for a small set of platforms.
- `cosign sign-blob` signs a file with a keyless certificate and writes a Sigstore bundle ([Sigstore: signing blobs](https://docs.sigstore.dev/cosign/signing/signing_with_blobs/)). GitHub artifact attestations attest a file by its digest ([GitHub: artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations)).

The real options for the signature of the binaries were:

1. **A cosign signature for each archive.** A consumer verifies one file, but the release has one more asset for each platform.
2. **A cosign signature for one checksum file.** A consumer verifies the checksum file, and then the checksum of the archive. The release has two more assets.

## Decision

The release workflow builds the `cdrlab` binaries from the release commit, with a `go build` for each platform and no other build tool. The binaries are the only artifacts that the release builds. We accept the rebuild for the binaries, because no tested binary exists, and the release still promotes each image.

- The binary gets the version of the release, and the commit, with `-X` flags. A build with no version flag uses the module version of its build information, when that is a release version, so that `go install …@vX.Y.Z` gives the same version. Each other build gives the version `dev`. Only the CLI gets the version. `service.version` of the images stays the short commit.
- Each archive contains the binary and the licence notices of its third-party code, as an image does (ADR 0018).
- The release signs one checksum file of the archives with keyless cosign (option 2), with the identity of ADR-0031. A consumer then verifies one signature for all platforms.
- The release attests the provenance and the SBOM of each archive, because the release is the build that makes it (ADR-0031).
- The release has no package manager entries, for example Homebrew. Release binaries and `go install` are sufficient for the users that exist.

## Consequences

- `cdrlab version` gives the version of the release, and `cdrlab up` finds the images of its release from it.
- The released binary is not a binary that CI tested. A defect of the build of one platform shows only in the release. A run of the release mode of `cdrlab up` before the publication finds a defect of the binary on Linux.
- The provenance of a binary gives the release run and the release commit. The provenance of an image gives a CI run on `main` and its image commit.
- A consumer needs `cosign` or `gh` to verify a binary. `go install` gets the source through the Go module proxy and the checksum database ([Go: module authentication](https://go.dev/ref/mod#authenticating)), with no cosign signature.
- `go.mod` must never get a `replace` or an `exclude` directive, or `go install …@vX.Y.Z` stops working.
- A binary from `go install` has no commit in its version output.
- A new platform needs a change to the build loop and to the documentation, and has no test before its first release.
