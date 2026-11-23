# 0031. Sign releases with keyless cosign, and attest the provenance and an SBOM with GitHub artifact attestations

- Date: 2026-11-23
- Amended by [ADR-0039](0039-build-the-cdrlab-binaries-in-the-release-with-the-release-version-and-sign-their-checksums-with-keyless-cosign.md): the signatures and the attestations of the `cdrlab` binaries

## Context

The images and charts of the repository are public on `ghcr.io`. A team that deploys a release into its own cluster needs to verify that an artifact comes from this repository and from its workflow, and which source and build made it ([ADR 0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md)). CI has no signatures, no provenance and no SBOM.

These facts constrain the choice:

- A release adds a tag to the tested image digest, with no rebuild ([ADR 0029](0029-promote-the-tested-image-of-each-service-to-a-release-with-no-rebuild.md)). A signature on a digest stays valid for a new tag of that digest.
- A chart of a release has a new version in its `Chart.yaml`, so it has a new digest. The release packages it, and only the release can sign it.
- The release commit and the image commit of a service can be different commits. The provenance of an image describes the build that made it.
- Helm verifies only PGP provenance files with a keyring ([Helm: provenance and integrity](https://helm.sh/docs/topics/provenance/)). It does not verify a cosign signature on an OCI chart.
- Flux verifies cosign and Notation signatures of an OCI chart ([Flux: HelmChart verification](https://fluxcd.io/flux/components/source/helmcharts/)). It does not verify GitHub artifact attestations.
- A scan of a service image finds the Go modules from the build information of the binary. The npm packages of the web UIs are bundled into the binary, and the image has no npm lock file. Thus a scan of the image does not find them.

The real options were:

1. **cosign with a long-lived key.** The repository must keep the private key as a secret and replace it when it leaks. A consumer needs the public key.
2. **Keyless cosign.** The signer gets a short-lived certificate for the OIDC identity of the workflow run, and the signature goes into the public transparency log ([Sigstore: signing containers](https://docs.sigstore.dev/cosign/signing/signing_with_containers/)). No key to keep. The consumer verifies the identity of the workflow.
3. **GitHub artifact attestations only.** They give a signed SLSA provenance, at SLSA Build Level 2, and an SBOM attestation ([GitHub: artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations)). `gh attestation verify` checks them. Flux and plain `cosign verify` do not find them.
4. **PGP provenance files for the charts.** Helm verifies them at install, but they need a long-lived key, and they do not apply to images.

## Decision

The release signs each image digest and each chart digest of the release with keyless cosign. GitHub artifact attestations give the provenance and the SBOM.

- The release workflow signs. Thus the identity of each signature is the release workflow of the repository, run from `main`, and a consumer requires that identity. We sign in the release, and not at the build, because only the release can sign a chart, and one identity for all artifacts of a release is simpler for a consumer.
- The build that makes an artifact attests its provenance. CI attests each image when it publishes the image from `main`, and the release attests each chart that it packages. Thus the provenance of an image gives the image commit and the CI run that built it, not the release commit.
- CI attests an SBOM of each image in SPDX JSON when it publishes the image. The SBOM lists the Go modules of the service binary. It does not list the npm packages of the web UIs. The release notes and the README of each chart state this.
- A consumer verifies a chart with cosign on its digest, and then installs the chart by that digest. The README of each chart gives the commands that verify the signatures and the attestations.
- Only the jobs that sign or attest get the permissions to make an OIDC token and to write attestations. No job that runs for a pull request gets them.
- We use both cosign and attestations, because Flux and admission controllers verify cosign signatures, and attestations give the provenance and the SBOM.

## Consequences

- The repository keeps no signing key. A consumer verifies the workflow identity and the OIDC issuer of GitHub Actions.
- Each signature and each attestation is in a public transparency log, with the repository and the workflow. The repository is public, so this shows nothing new.
- A consumer needs `cosign` and `gh` to verify all artifacts. `helm install` alone verifies nothing.
- An image from `main` has a provenance and an SBOM, but no cosign signature. Only a release has signatures.
- The SBOM does not show the npm packages. A vulnerability scan from the SBOM misses them.
- The provenance is at SLSA Build Level 2. Level 3 needs the build in an isolated reusable workflow.
- Argo CD verifies only GnuPG signatures of git commits ([Argo CD: source integrity](https://argo-cd.readthedocs.io/en/latest/user-guide/source-integrity/)). A team that uses it verifies the artifacts in a separate step.
