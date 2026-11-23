# Install the lab

This guide gives the three ways to start the lab:

- [The compose lab from a release](#the-compose-lab-from-a-release): one host with Docker, and no checkout of the repository.
- [The compose lab from a checkout](#the-compose-lab-from-a-checkout): for a contributor to the repository.
- [A cluster deployment](#a-cluster-deployment): the umbrella chart in a namespace of a Kubernetes cluster, for a team.

The lab is a simulation. It is not for production use, and all its data must be invented. [docs/production-gaps.md](../production-gaps.md) lists the shortcuts of the lab.

## The compose lab from a release

You need Docker with Compose (`docker compose`), on Linux or macOS, `amd64` or `arm64`. On Windows, use WSL 2. To use a different Compose command, set `CDRLAB_COMPOSE`, for example `CDRLAB_COMPOSE=docker-compose`.

Each release `vX.Y.Z` has a `cdrlab` archive for each platform, a checksum file and its signature. The release notes of each version are the source of these commands. They need [gh](https://cli.github.com) and [cosign](https://docs.sigstore.dev/cosign/system_config/installation/). In an empty directory:

```sh
version=vX.Y.Z platform=linux-amd64
gh release download "$version" --repo lukemundy/cdr-lab \
  --pattern "cdrlab-$version-$platform.tar.gz" --pattern "cdrlab-$version-checksums.txt*"
cosign verify-blob "cdrlab-$version-checksums.txt" --bundle "cdrlab-$version-checksums.txt.sigstore.json" \
  --certificate-identity https://github.com/lukemundy/cdr-lab/.github/workflows/release.yml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
sha256sum --check --ignore-missing "cdrlab-$version-checksums.txt"  # macOS: shasum -a 256 --check --ignore-missing
gh attestation verify "cdrlab-$version-$platform.tar.gz" --repo lukemundy/cdr-lab \
  --signer-workflow lukemundy/cdr-lab/.github/workflows/release.yml --source-ref refs/heads/main
tar -xzf "cdrlab-$version-$platform.tar.gz" cdrlab
```

Then start the lab. The lab directory is the current directory: `cdrlab up` writes the compose project and the certificates of the lab in it, and pulls the images of its version from `ghcr.io/lukemundy/cdr-lab`:

```sh
./cdrlab up
http --verify .lab/pki/public/root.crt https://localhost:8181/cds-au/v1/banking/products x-v:5
```

The ports are in the [README](../../README.md#lab-deployment). `./cdrlab -h` gives each command and its flags. For example, `-all-banks` also starts the three other preset brands, and `-no-telemetry` starts no telemetry backend.

To stop the lab and keep its data, run `./cdrlab down`. To delete its data, its CAs and the files that `cdrlab up` wrote, run `./cdrlab reset`. `cdrlab up` starts only a lab of its own version. To change the version, see the [upgrade guide](upgrade.md#the-compose-lab).

With Go, `go install github.com/lukemundy/cdr-lab/cli/cdrlab/cmd/cdrlab@vX.Y.Z` builds the same `cdrlab` from the source, with no signature.

## The compose lab from a checkout

The [README](../../README.md#development) gives the tools and the commands. In short:

```sh
mise install
mise run up
```

`mise run up` builds the images from the checkout. `mise run reset` deletes the data and the CAs.

## A cluster deployment

The umbrella chart deploys the Register and a set of the preset brands to one namespace. The [chart README](../../charts/cdr-lab/README.md) gives each value and each step. This section gives the sequence.

1. Make the prerequisites in the namespace ([before the install](../../charts/cdr-lab/README.md#before-the-install)):
   - a PostgreSQL database for the Register and for each brand, each with an owner user and a service user, as Secrets. The chart README gives an example with [CloudNativePG](../../charts/cdr-lab/README.md#postgresql-with-cloudnativepg).
   - the operator token Secret.
   - the [public TLS certificates](../../charts/cdr-lab/README.md#public-tls-certificates): a cert-manager issuer, or Secrets that you make.
   - a mail server for the one-time password emails, for example Mailpit.
   - optional: a Gateway for external host names ([routes](../../charts/cdr-lab/README.md#routes)).
2. Write a values file. The [install section](../../charts/cdr-lab/README.md#install) of the chart README gives an example, and `values.yaml` gives each value. Add the seed data of your release to the values ([seed data](../../charts/cdr-lab/README.md#seed-data)).
3. Verify the chart and its images by digest ([verify a release](../../charts/cdr-lab/README.md#verify-a-release)).
4. Install the chart:

   ```sh
   helm install cdr-lab oci://ghcr.io/lukemundy/cdr-lab/charts/cdr-lab --version X.Y.Z \
     --namespace lab --values my-values.yaml --wait --timeout 10m
   kubectl get pods,jobs --namespace lab
   ```

5. Check the install ([verify an install](../../charts/cdr-lab/README.md#verify-an-install)).
6. Get access to the operator interfaces with `kubectl port-forward` ([operator access](../../charts/cdr-lab/README.md#operator-access)).

### A cluster from a checkout

From a checkout, `mise run cluster:up` makes a [kind](https://kind.sigs.k8s.io) cluster and installs the chart of the checkout, as the cluster job of CI does. It uses the images of the lab: the comment of the task in `mise.toml` gives the commands that make them. `mise run cluster:conformance` runs the conformance suite in it, and `mise run cluster:down` deletes it.
