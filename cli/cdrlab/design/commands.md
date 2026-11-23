# cdrlab: Commands

This area gives the commands of `cdrlab`, its version, its access to the Kubernetes API and the `cdrlab` image.

## Commands

| ID | Command | Function | Source |
|---|---|---|---|
| <a id="cli-0104"></a>**CLI-0104** | `cdrlab version` | Gives the version and the commit of the build. | [Judgement: the release version only in the CLI (ADR 0039)][adr-0039] |
| <a id="cli-0105"></a>**CLI-0105** | `cdrlab ca init` | Makes the mock CDR CA, with Secrets or a directory as the output. With a directory, it also makes the mock public CA. | [Judgement: one CLI for the lab (ADR 0035)][adr-0035] |
| <a id="cli-0106"></a>**CLI-0106** | `cdrlab cert issue` | Gets a CDR certificate through the Register, with a Secret or files as the output. | [Judgement: CDR certificates from the Register (ADR 0032)][adr-0032] |
| <a id="cli-0107"></a>**CLI-0107** | `cdrlab public-cert issue` | Issues a public TLS server certificate from the mock public CA in a directory, with files as the output. | [Judgement: public TLS certificates from the files of the public CA](certificates.md#cli-0170) |
| <a id="cli-0108"></a>**CLI-0108** | `cdrlab up` | Makes the CAs and the certificates of the compose lab, and starts the compose lab. | [Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038] |
| <a id="cli-0109"></a>**CLI-0109** | `cdrlab down` | Stops the compose lab, and keeps its data. | [Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038] |
| <a id="cli-0110"></a>**CLI-0110** | `cdrlab reset` | Stops the compose lab, and deletes its data, its CAs and its certificates. | [Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038] |
| <a id="cli-0177"></a>**CLI-0177** | `cdrlab participants add` | Adds the participants of a load file to the Register, and changes no participant that exists. | [Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040] |
| <a id="cli-0178"></a>**CLI-0178** | `cdrlab customers add` | Adds the customers of a customer load file to one data holder brand, and changes no customer that exists. | [Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040] |

- <a id="cli-0111"></a>**CLI-0111** The repository has no CLI for the lab other than `cdrlab`. The repository tools in `tools/` are not CLIs for the lab. ([Judgement: one CLI for the lab (ADR 0035)][adr-0035])
- <a id="cli-0112"></a>**CLI-0112** `cdrlab` has no command that revokes a certificate, signs a CRL or keeps a list of revocations. ([Judgement: the Register is the CDR intermediate CA (ADR 0022)][adr-0022])
- <a id="cli-0113"></a>**CLI-0113** A command that has files as the output makes 0 calls to the Kubernetes API. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])

## Version

- <a id="cli-0114"></a>**CLI-0114** The version command gives the version, the full commit ID and the build time. For example: `cdrlab v0.3.1, commit 0123…cdef, built 2026-12-01T03:00:00Z`. ([Judgement: the release version only in the CLI (ADR 0039)][adr-0039])
- <a id="cli-0115"></a>**CLI-0115** When the build sets a version, the command gives that version. For example, a release binary gives the version of its release. ([Judgement: the release version only in the CLI (ADR 0039)][adr-0039])
- <a id="cli-0116"></a>**CLI-0116** When the build sets no version, and the Go module version of the binary is a [release version](../../../tools/release/DESIGN.md#rls-0004), the command gives the module version. For example, `go install github.com/lukemundy/cdr-lab/cli/cdrlab/cmd/cdrlab@v0.3.1` gives `v0.3.1`. ([Go: runtime/debug.ReadBuildInfo][go-buildinfo]; [Judgement: the release version only in the CLI (ADR 0039)][adr-0039])
- <a id="cli-0117"></a>**CLI-0117** When the version has neither source, it is `dev`. For example, a pseudo-version of `go install …@main` and a `go run` in a checkout give `dev`. ([Judgement: the release version only in the CLI (ADR 0039)][adr-0039])
- <a id="cli-0118"></a>**CLI-0118** When the build gives no commit, the output gives `unknown` as the commit. ([Judgement: the release version only in the CLI (ADR 0039)][adr-0039])

## Kubernetes access

- <a id="cli-0001"></a>**CLI-0001** In a pod, the commands use the ServiceAccount of the pod and its namespace. ([Kubernetes: access the API from a pod][k8s-api-from-pod])
- <a id="cli-0002"></a>**CLI-0002** The commands read and write Secrets only in the namespace of the pod. ([Judgement: CDR certificates from the Register (ADR 0032)][adr-0032])
- <a id="cli-0003"></a>**CLI-0003** When a command fails, it exits with a status that is not 0, and it writes the cause to its error output. ([POSIX: exit status][posix-exit])

## Image

- <a id="cli-0050"></a>**CLI-0050** CI publishes the `cdrlab` image from `main` to `ghcr.io/lukemundy/cdr-lab/cdrlab`, with the rules of the image of a service. ([data holder: the publication of the image](../../../services/dataholder/design/deployment.md#dh-1325); [Judgement: the image of the lab CLI (ADR 0035)][adr-0035])
  - The tags are the full and the short commit ID.
  - The image is an OCI image index for `linux/amd64` and `linux/arm64`.
  - CI publishes it only after its gate jobs pass, and for each commit that changes an input of the image.
- <a id="cli-0051"></a>**CLI-0051** The inputs of the image are the Go code of `cli/cdrlab`, the Go code of `pkg/`, `go.mod`, `go.sum` and the `Dockerfile` of `cdrlab`. ([Judgement: the image job runs for each input of the image](../../../services/dataholder/design/deployment.md#dh-1372))
- <a id="cli-0052"></a>**CLI-0052** The image has a distroless base image, the `cdrlab` binary and the licence notices of its third-party code in `/licenses/`. ([Judgement: the image of the lab CLI (ADR 0035)][adr-0035])
- <a id="cli-0053"></a>**CLI-0053** The image runs as a user that is not root. ([Judgement: the image of the lab CLI (ADR 0035)][adr-0035])

## Judgements

None.

[adr-0022]: ../../../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[adr-0032]: ../../../docs/adr/0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md
[adr-0035]: ../../../docs/adr/0035-give-the-lab-one-cli-cdrlab-in-cli-with-an-image-that-follows-the-rules-of-a-service-image.md
[adr-0038]: ../../../docs/adr/0038-run-the-compose-lab-with-cdrlab-from-a-compose-project-in-the-binary-and-issue-its-cdr-certificates-from-the-host.md
[adr-0039]: ../../../docs/adr/0039-build-the-cdrlab-binaries-in-the-release-with-the-release-version-and-sign-their-checksums-with-keyless-cosign.md
[adr-0040]: ../../../docs/adr/0040-add-participants-and-customers-with-add-only-loads-of-the-operator-interfaces-from-cdrlab-and-from-the-umbrella-values.md
[go-buildinfo]: https://pkg.go.dev/runtime/debug#ReadBuildInfo
[k8s-api-from-pod]: https://kubernetes.io/docs/tasks/run-application/access-api-from-pod/
[posix-exit]: https://pubs.opengroup.org/onlinepubs/9799919799/utilities/V3_chap02.html#tag_19_08_02
