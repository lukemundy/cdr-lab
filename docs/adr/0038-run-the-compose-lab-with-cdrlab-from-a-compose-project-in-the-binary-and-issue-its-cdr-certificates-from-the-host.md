# 0038. Run the compose lab with cdrlab, from a compose project in the binary, and issue its CDR certificates from the host

- Date: 2026-11-23

## Context

The compose lab starts from a checkout of the repository. The mise tasks run the `pki` command with `go run` to make the CAs and the certificates, and then run Docker Compose. Thus the host needs Go, mise, openssl and a checkout. A team that wants only to run the lab from a release cannot do this with Docker alone.

[ADR 0032](0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md) requires that each CDR certificate of a deployment comes from the Register, also in compose. [ADR 0035](0035-give-the-lab-one-cli-cdrlab-in-cli-with-an-image-that-follows-the-rules-of-a-service-image.md) gives one CLI, `cdrlab`, and a release binary for a person on a host.

These facts constrain the design:

- The compose file mounts files that are in the repository: the seed file of the Register and the set-up SQL files of PostgreSQL. A release user has no checkout.
- The data holders read their CDR server certificates when they start. The Register issues a certificate only while it runs. Thus a data holder can start only after the Register runs and its certificate exists.
- The compose services read their certificates and keys from bind mounts of `.lab/pki`. A container that writes into a bind mount writes as its own user, for example the non-root user of a distroless image. The host user then cannot always change or delete the files, and with rootless Podman the user IDs are mapped.
- The Register publishes its internal listener, with the operator interface, on `127.0.0.1` of the host.
- Compose gives the images of a release only through the `image` keys of the compose file. The images of a release have the tag `vX.Y.Z` ([ADR 0029](0029-promote-the-tested-image-of-each-service-to-a-release-with-no-rebuild.md)).
- A Go binary can contain files with `go:embed` ([Go: embed](https://pkg.go.dev/embed)).

The real options to get the compose project on a release host were:

1. **The binary contains the compose project of its release commit.** The binary and the compose project cannot have different versions. The binary is larger, and `go install` also gives the project.
2. **A release asset with the compose project.** The binary downloads it for its version. This needs a network call to GitHub and one more asset to sign.

The real options to issue the CDR certificates of compose were:

1. **A one-shot compose service that runs the `cdrlab` image** after the Register is healthy, as in a cluster. The files then belong to the user of the container.
2. **`cdrlab` on the host** calls the operator interface of the Register on its host port, between a start of the Register and a start of the other services. The files belong to the host user. `cdrlab up` then starts the lab in two steps.

## Decision

`cdrlab up`, `cdrlab down` and `cdrlab reset` run the compose lab. The mise lab tasks call these commands. Compose stays the deployment for local development.

- `cdrlab up` has two modes. The **source mode** uses the compose project of a checkout, and Compose builds the images. The mise tasks use it. The **release mode** uses the compose project that the binary contains (option 1), and Compose pulls the images with the version of the binary as the tag. We select option 1, because a binary and its compose project then always agree, with no network call other than the image pulls.
- The release mode writes the compose project into the lab directory, and it changes only a lab that it wrote with the same version. A version 0.x makes no promise about the data of an earlier version ([ADR 0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md)), so a lab of a different version needs a reset.
- `cdrlab` on the host makes the CAs, issues the public TLS certificates from the mock public CA, and gets each CDR certificate through the Register issue path, with the file output (option 2). We select option 2, because the host user then owns all files of `.lab/pki`, in each container runtime, and compose needs no `cdrlab` image.
- The compose project gives each value of the lab that `cdrlab` needs: the certificates, their names and identifiers, the URL of the operator interface on the host and the operator token. `cdrlab` has no default for a value of one deployment.
- `cdrlab` runs Docker Compose as a separate program. It does not contain Compose.

This decision does not apply to a cluster deployment, which keeps the Jobs of ADR 0032. It does not make the conformance suite and the OpenID conformance suite run from a release: they run the Go tests of a checkout.

## Consequences

- A host with Docker and Compose runs the lab from a release, with no Go, Node, mise, openssl or checkout.
- Compose and a cluster use the same `cdrlab` commands and the same Register issue path. Only the output is different.
- A client certificate issue also registers the certificate ([Register: the issue registers the certificate](../../services/register/design/certificate-management.md#reg-0610)). A test that needs a client certificate that the Register does not accept must remove the registration first.
- `docker compose up` alone does not start a complete lab, because the data holders need the certificates that `cdrlab up` gets. The documentation uses `cdrlab up`.
- Each `cdrlab up` checks each certificate and renews it when it is near its expiry. A lab that runs for more than 13 months renews only at a `cdrlab up`.
- A change of the compose project, the seed file or the SQL files goes into the next release binary. A release user cannot change the seed file of the binary, and uses the operator interfaces to add participants.
- The binary is larger by the size of the compose project and the seed file.
- An upgrade of the binary needs `cdrlab reset` before `cdrlab up`, also when the release notes say that no fresh deployment is necessary.
