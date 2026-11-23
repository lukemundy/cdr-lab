# The lab in CI

This guide is for the CI of a different repository, for example the tests of a software product against the lab. It gives the compose lab on a hosted runner, the pin of a release version, the onboarding in the job, and a kind cluster as an option.

## The compose lab on a hosted runner

A GitHub-hosted Ubuntu runner has Docker with Compose and the GitHub CLI, so `cdrlab up` needs no other tool. The images of a release are public on `ghcr.io`, so the job needs no registry login.

The job below starts the compose lab of a pinned release, with no telemetry backend, onboards a software product, runs the tests and then deletes the lab. The files `example-budget.json` and `customers.json` are in your repository, in the formats of the [ADR developers guide](adr-developers.md):

```yaml
jobs:
  lab-tests:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    env:
      CDRLAB_VERSION: vX.Y.Z
      GH_TOKEN: ${{ github.token }}
      REGISTER_OPERATOR_TOKEN: lab-operator-token
      DATAHOLDER_OPERATOR_TOKEN: lab-operator-token
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1

      - name: Download and verify cdrlab
        run: |
          archive="cdrlab-$CDRLAB_VERSION-linux-amd64.tar.gz"
          gh release download "$CDRLAB_VERSION" --repo lukemundy/cdr-lab --pattern "$archive"
          gh attestation verify "$archive" --repo lukemundy/cdr-lab \
            --signer-workflow lukemundy/cdr-lab/.github/workflows/release.yml --source-ref refs/heads/main
          tar -xzf "$archive" cdrlab

      - name: Start the lab
        run: ./cdrlab up -no-telemetry -wait-timeout 3m

      - name: Onboard the software product and the test customers
        run: |
          ./cdrlab participants add -register-url http://localhost:8080 -file example-budget.json
          mkdir -p budget
          ./cdrlab cert issue -register-url http://localhost:8080 -kind client \
            -legal-entity-id 6f6c2a04-6a2f-4c2a-9b77-0c1f8c0a1b01 -software-product-id 2b7f9c1e-0d4a-4e6b-8f3c-5a6b7c8d9e03 \
            -cn "Example Budget Web" -o "Example Budget Pty Ltd" -cert budget/client.crt -key budget/client.key \
            -key-readable-by-all
          ./cdrlab customers add -dataholder-url http://localhost:8180 -file customers.json

      - name: Test
        run: ./run-my-tests.sh

      - name: Show the logs of the lab
        if: failure()
        run: docker compose --profile '*' logs --no-color --tail 200

      - name: Delete the lab
        if: always()
        run: ./cdrlab reset
```

Notes:

- `gh attestation verify` checks the provenance of the archive: the release workflow of `lukemundy/cdr-lab` built it on `main`. The [install guide](install.md#the-compose-lab-from-a-release) also gives the cosign check of the checksum file, which needs cosign on the runner.
- `-no-telemetry` starts no telemetry backend, and the services export no telemetry ([cdrlab: telemetry off](../../cli/cdrlab/design/compose-lab.md#cli-0152)). A cold runner pulls each image, so `-wait-timeout 3m` gives more than the default 60 seconds ([cdrlab: the wait](../../cli/cdrlab/design/compose-lab.md#cli-0153)).
- The lab directory is the working directory of the job. `cdrlab up` refuses a directory that has a file of its compose project and no record of `cdrlab`, for example a `compose.yaml` of your repository at the same path ([cdrlab: the release mode changes only a lab of its own version](../../cli/cdrlab/design/compose-lab.md#cli-0144)). Then give a different directory with `-dir`.
- A test that does an authorisation must run on the compose network, in a container, as the [ADR developers guide](adr-developers.md#run-your-software-on-the-lab-network) says. The tests on the runner can use the ports on `localhost` ([README: lab deployment](../../README.md#lab-deployment)).
- Each CI run starts a new lab, so the job needs no fresh deployment.

## Pin a release version

Give the version in one place, for example `CDRLAB_VERSION`. The binary starts the images of its own version, so one value pins the whole lab. To change the version, read the release notes of the new version first ([upgrade guide](upgrade.md)), then change the value in a pull request, so that its CI runs the tests against the new version.

## A kind cluster as an option

When your tests need the cluster deployment, make a [kind](https://kind.sigs.k8s.io) cluster in the job, for example with the `helm/kind-action` action. Then do the steps of a [cluster deployment](install.md#a-cluster-deployment): the prerequisites, the values with your seed data, and `helm install` of the chart version that you pin. The cluster job of this repository does the same steps (`mise run cluster:up`). A kind cluster takes more time and more memory than the compose lab, so use it only for a test that needs Kubernetes.
