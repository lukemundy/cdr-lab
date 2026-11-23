# cdrlab: Compose lab

This area gives the start, the stop and the reset of the compose lab with `cdrlab`, in the source mode and in the release mode.

## Compose lab

`cdrlab up` has a source mode for a checkout of the repository, and a release mode for a host with no checkout ([ADR 0038][adr-0038]).

- <a id="cli-0134"></a>**CLI-0134** The lab directory is the current directory. An input gives a different directory. ([Judgement: the current directory as the lab directory](#cli-0173))
- <a id="cli-0135"></a>**CLI-0135** The CAs and the certificates of the compose lab are in `.lab/pki` of the lab directory. ([Judgement: the current directory as the lab directory](#cli-0173))
- <a id="cli-0136"></a>**CLI-0136** `cdrlab` runs Docker Compose as `docker compose`. When the environment variable `CDRLAB_COMPOSE` has a value, `cdrlab` runs that command. For example, `docker-compose`. ([Judgement: the Compose command from the environment](#cli-0174))
- <a id="cli-0137"></a>**CLI-0137** An input selects the source mode. In the source mode, `cdrlab up` uses the compose project of the lab directory, and Compose builds the images. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0138"></a>**CLI-0138** In the source mode, an input makes Compose use the images that are there, with no build. For example, the images that CI loads. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0139"></a>**CLI-0139** The binary contains the compose project of its commit: the compose file, the certificate list, the seed file of the Register and the set-up SQL files of the lab PostgreSQL. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0140"></a>**CLI-0140** In the release mode, `cdrlab up` writes the compose project of the binary into the lab directory, and records the version of the binary there. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0141"></a>**CLI-0141** In the release mode, Compose pulls the image of each service from `ghcr.io/lukemundy/cdr-lab/<service>`, with the version of the binary as the tag. ([release: the version tag of each image](../../../tools/release/DESIGN.md#rls-0030); [Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0142"></a>**CLI-0142** In the release mode, an input gives a different image repository prefix, and an input gives a different image tag. For example, `ghcr.io/lukemundy/cdr-lab/rehearsal` and `rehearsal-<run ID>`. ([release: the images of a rehearsal](../../../tools/release/DESIGN.md#rls-0014))
- <a id="cli-0143"></a>**CLI-0143** When the version of the binary is not a [release version](../../../tools/release/DESIGN.md#rls-0004) and no input gives an image tag, the release mode changes nothing, and it fails. The error tells the user to use the source mode or a release binary. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0144"></a>**CLI-0144** When the lab directory has a file of the compose project of the binary and no version record of `cdrlab`, the release mode changes nothing, and it fails. For example, the lab directory is a checkout of the repository, or it has a seed file of the user at the path of the seed file of the project. ([Judgement: the release mode changes only a lab of its own version](#cli-0175))
- <a id="cli-0145"></a>**CLI-0145** When the version record of the lab directory is not the version of the binary, the release mode changes nothing, and it fails. The error gives the two versions, and tells the user to run `cdrlab reset`. ([Judgement: the release mode changes only a lab of its own version](#cli-0175))

## Start of the compose lab

- <a id="cli-0146"></a>**CLI-0146** `cdrlab up` does these steps in this sequence, and it stops at the first step that fails. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
  1. In the release mode, it writes the compose project.
  2. It runs [the CA set-up](certificates.md#cli-0119), with `.lab/pki` as the output.
  3. It runs [the public certificate command](certificates.md#cli-0130) for each public TLS certificate of the certificate list.
  4. It starts the Register and the services that the Register depends on, and waits until the Register is healthy.
  5. It runs [the certificate command](certificates.md#cli-0030), with files as the output, for each CDR certificate of the certificate list.
  6. It starts the other services, and waits until each service is healthy or has completed.
- <a id="cli-0147"></a>**CLI-0147** The certificate list gives the URL of the Register operator interface from the host, and the operator token. For each certificate, it gives the inputs of its command and its files. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0193"></a>**CLI-0193** The certificate list is the JSON file `lab/certificates.json` of the compose project. When a path of the list is not in the lab directory, `cdrlab up` changes nothing, and it fails. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0194"></a>**CLI-0194** The keys of each CDR certificate of the list are the flag names of [the certificate command](certificates.md#cli-0030), so that a different tool can give the entries to that command. For example, the key `legal-entity-id` gives `-legal-entity-id`. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0148"></a>**CLI-0148** `cdrlab up` gets each certificate of the certificate list, also a certificate of a service that it does not start. For example, the certificates of the three other preset brands. ([Judgement: each certificate of the list at each start](#cli-0176))
- <a id="cli-0149"></a>**CLI-0149** Any user can read each key file of the certificate list. ([PKI: lab keys that any user can read](../../../pki/DESIGN.md#pki-0077))
- <a id="cli-0150"></a>**CLI-0150** When the inputs name services, `cdrlab up` starts only those services and their dependencies in step 6. ([Compose: up][compose-up])
- <a id="cli-0151"></a>**CLI-0151** An input also starts the three other preset brands (the compose profile `all-banks`). ([Judgement: one preset brand by default](../../../services/dataholder/design/deployment.md#dh-1360))
- <a id="cli-0152"></a>**CLI-0152** `cdrlab up` starts the telemetry backend, unless an input turns it off. With telemetry off, the services export no telemetry. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0153"></a>**CLI-0153** When a wait of step 4 or step 6 takes more than 60 seconds, `cdrlab up` fails, and the error names each service that is not healthy. An input gives a different time. ([Compose: up][compose-up])

## Stop and reset of the compose lab

- <a id="cli-0154"></a>**CLI-0154** `cdrlab down` stops and removes the containers of each compose profile. It keeps the volumes and `.lab/pki`. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])
- <a id="cli-0155"></a>**CLI-0155** `cdrlab reset` stops and removes the containers and the volumes of each compose profile, and deletes `.lab/pki`. ([PKI: a lab reset deletes the CAs and the Register database](../../../pki/DESIGN.md#pki-0065))
- <a id="cli-0156"></a>**CLI-0156** In the release mode, `cdrlab reset` also deletes the compose project and the version record that `cdrlab up` wrote. It deletes no other file. ([Judgement: the release mode changes only a lab of its own version](#cli-0175))
- <a id="cli-0157"></a>**CLI-0157** When the lab directory has no lab, `cdrlab down` and `cdrlab reset` change nothing, and exit with the status 0. ([Judgement: the CA set-up changes no CA that exists](certificates.md#cli-0100))

## Judgements

### <a id="cli-0173"></a>CLI-0173 The current directory as the lab directory

- Kind: choice
- Basis: Step 3.2. Compose finds the compose file in the current directory ([Compose: `docker compose`][compose-cli]). The mise tasks run in the root of the repository.
- Reason: The same rule then gives the root of a checkout in the source mode, and the directory of the user in the release mode. A fixed directory of the user is a value that differs on each operating system.

### <a id="cli-0174"></a>CLI-0174 The Compose command from the environment

- Kind: choice
- Basis: Step 3.2. Docker installs Compose as the plugin `docker compose` ([Docker: install Compose][compose-install]). `mise.toml` pins the program `docker-compose`, and a Podman host uses it with the Docker API of Podman.
- Reason: A host with only Docker needs no setting. A developer gets the pinned version with one variable, which mise sets.

### <a id="cli-0175"></a>CLI-0175 The release mode changes only a lab of its own version

- Kind: choice
- Basis: Step 3.3. A release makes no promise about the data of a deployment while the major version is 0 ([ADR 0027][adr-0027]).
- Reason: The images of a different version can fail on the data of the lab, and the cause is then not clear. A refusal with a clear error costs a reset, and the lab data has no value. The record also stops a write over a checkout or another compose project. The check is for each file of the project, not only the compose file, because the release mode must not write over any file that it did not write.

### <a id="cli-0176"></a>CLI-0176 Each certificate of the list at each start

- Kind: choice
- Basis: Step 3.3. A start with a compose profile can come after an earlier start with no profile.
- Reason: Then each service has its certificate when a later start turns on its profile. Each start also renews each certificate in its renewal window, so a lab that stays up for more than 13 months renews at its next start.

[adr-0027]: ../../../docs/adr/0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md
[adr-0038]: ../../../docs/adr/0038-run-the-compose-lab-with-cdrlab-from-a-compose-project-in-the-binary-and-issue-its-cdr-certificates-from-the-host.md
[compose-cli]: https://docs.docker.com/reference/cli/docker/compose/
[compose-install]: https://docs.docker.com/compose/install/
[compose-up]: https://docs.docker.com/reference/cli/docker/compose/up/
