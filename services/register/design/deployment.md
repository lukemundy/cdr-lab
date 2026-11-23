# Mock CDR Register: Deployment

This area gives the listeners of the Register and their ports, the lab-only participant load, the participant seed of the Helm chart, the example seed of the four preset brands, and the publication of the image and the chart. The production Register has no participant load. Real components must not depend on it. The rules of the chart and of the example seed apply to the chart and to the file, not to a deployment.

## Listeners

- <a id="reg-1498"></a>**REG-1498** The Register has three listeners. ([Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])
  - the internal listener, with plain HTTP
  - the public listener, with TLS and no client authentication
  - the mTLS listener, with TLS and a required client certificate
- <a id="reg-1499"></a>**REG-1499** The internal listener serves these groups of routes, [each with its own security policy](../../../pkg/platform/httpserver/DESIGN.md#srv-0009). ([Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])
  - the health probes
  - the operator interface
  - the operator console, when it is on
  - the CRL
- <a id="reg-1500"></a>**REG-1500** In the lab, the internal listener is on port 8080, the public listener on port 8081 and the mTLS listener on port 8083. ([Judgement: each listener keeps the last digit of its port](#reg-1504))
- <a id="reg-1501"></a>**REG-1501** The setting `INTERNAL_ALLOWED_HOSTS` is the `Host` allowlist of the operator interface and of the console. Its values are [the values of the `Host` header](../../../pkg/platform/hostallowlist/DESIGN.md#host-0001). For example, `127.0.0.1:8080`. ([Judgement: a Host allowlist with no default](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0009); [Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])
- <a id="reg-1509"></a>**REG-1509** In a process that serves the operator interface, the setting `INTERNAL_ALLOWED_HOSTS` is required, and it has no default. ([Judgement: a Host allowlist with no default](../../../pkg/platform/operatorhttp/DESIGN.md#opi-0009))
- <a id="reg-1502"></a>**REG-1502** A process that runs only the background jobs serves only the health probes on the internal listener. ([Judgement: background jobs with River (ADR 0010)][adr-0010]; [Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])
- <a id="reg-1503"></a>**REG-1503** The chart gives one container port and one Service port for each listener. ([Judgement: listeners by TLS mode and audience (ADR 0025)][adr-0025])
- <a id="reg-1506"></a>**REG-1506** The chart README says that a deployment exposes only the CRL path of the internal listener outside the cluster. For example, an ingress rule routes only `/crl/cdr-intermediate.crl` to the internal listener. ([Judgement: the CRL on the internal listener, with no authentication](certificate-revocation.md#reg-0751))

## Participant load

- <a id="reg-0900"></a>**REG-0900** The lab operator can load a full set of participant data from a file into a Register that has no participants. ([Judgement: no load into a Register with participants](#reg-0940))
- <a id="reg-0901"></a>**REG-0901** When the Register has participants, the load refuses the file and changes nothing. ([Judgement: no load into a Register with participants](#reg-0940))
- <a id="reg-0902"></a>**REG-0902** The load is separate from the start of the Register. The start of the Register does not need a file. ([Judgement: the seed in the chart is off by default](#reg-0943))
- <a id="reg-0903"></a>**REG-0903** The load checks the full set against [the participant records](participants.md#reg-0200), [the status transitions](participants.md#reg-0222) and [the cascades](participants.md#reg-0227). ([CDS: Participant Statuses][cds-participant-statuses])
- <a id="reg-0904"></a>**REG-0904** When one record is not valid, the load changes nothing, and it names each record and rule that fails. ([Judgement: no load into a Register with participants](#reg-0940))
- <a id="reg-0905"></a>**REG-0905** The load file uses the field names of the CDS Register schemas. For example, a software product has the field `softwareProductId`. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: the load file uses the CDS field names](#reg-0941))
- <a id="reg-0906"></a>**REG-0906** The load is a command of the Register binary, and it needs access to the Register database. ([Judgement: no load into a Register with participants](#reg-0940))

## Publication

- <a id="reg-0910"></a>**REG-0910** An operator deploys the Register to Kubernetes from the image and the Helm chart that CI publishes from `main`, or from the image and the chart of a [release](../../../tools/release/DESIGN.md#rls-0009). ([Register chart][register-chart]; [README: Container images][readme-conventions]; [Judgement: one version from a manual workflow on main (ADR 0027)][adr-0027])
- <a id="reg-1510"></a>**REG-1510** Each image of the Register that CI publishes is an OCI image index with an image for `linux/amd64` and an image for `linux/arm64`. ([OCI: image index][oci-image-index]; [Judgement: images for linux/amd64 and linux/arm64 (ADR 0030)][adr-0030])
- <a id="reg-1511"></a>**REG-1511** Integrity: the `linux/amd64` image of the index has the same configuration digest as the image that the lab jobs of its commit tested. ([Judgement: images for linux/amd64 and linux/arm64 (ADR 0030)][adr-0030])
- <a id="reg-1512"></a>**REG-1512** The `appVersion` of a chart from `main` is the full commit ID of the [image commit](../../../docs/spec/glossary.md#image-commit) of the Register at that commit. ([Helm: charts][helm-charts]; [Judgement: promote the tested image with no rebuild (ADR 0029)][adr-0029])
- <a id="reg-1513"></a>**REG-1513** CI publishes the chart from `main` only when the image commit of the Register is the last commit that changes an input of the image. For example, when the commit changes an input of the image, the chart waits until CI publishes the image of that commit. ([Judgement: promote the tested image with no rebuild (ADR 0029)][adr-0029])

## Chart

- <a id="reg-1514"></a>**REG-1514** When the values give no image tag, the chart uses its `appVersion` as the tag of the Register image. A value of the image tag overrides the `appVersion`. ([Helm: charts][helm-charts]; [Judgement: promote the tested image with no rebuild (ADR 0029)][adr-0029])

## Participant seed in the chart

- <a id="reg-0911"></a>**REG-0911** The chart loads a participant seed file only when the values give the name of an existing ConfigMap that holds the file. ([Kubernetes: ConfigMap volume][k8s-configmap-volume]; [Judgement: a ConfigMap as the source of the seed](#reg-0942))
- <a id="reg-0912"></a>**REG-0912** The chart does not make the ConfigMap of the seed file. ([Judgement: a ConfigMap as the source of the seed](#reg-0942))
- <a id="reg-0913"></a>**REG-0913** When the values give no ConfigMap, the chart loads no file, and the Register starts with no participants. ([Judgement: the seed in the chart is off by default](#reg-0943))
- <a id="reg-0914"></a>**REG-0914** The chart loads the file before the Register starts, after the migrations complete. ([Kubernetes: init containers][k8s-init-containers]; [Judgement: the chart loads the seed file before the Register starts](#reg-0944))
- <a id="reg-0915"></a>**REG-0915** The load of the chart connects to the database as the service user, and gets no owner credentials. ([Judgement: service migrations create roles and the operator creates users (ADR 0020)][adr-0020])
- <a id="reg-0916"></a>**REG-0916** The load of the chart changes the Register only when the Register has no participants. When the Register has participants, the load changes nothing and does not fail. ([Judgement: no load into a Register with participants](#reg-0940))
- <a id="reg-0917"></a>**REG-0917** After a restart, or after a change of the ConfigMap, the Register has the same participants as before the restart or the change. ([Judgement: a ConfigMap as the source of the seed](#reg-0942))
- <a id="reg-0918"></a>**REG-0918** Only the load gets the seed file. The Register process does not get it. ([Judgement: the seed in the chart is off by default](#reg-0943))
- <a id="reg-0919"></a>**REG-0919** The chart README gives the seed value, the ConfigMap and its file, and the grant of the application role to the service user that the load needs. ([Judgement: service migrations create roles and the operator creates users (ADR 0020)][adr-0020])

## Example seed of the four preset brands

A deployment changes the endpoints of the example seed to the hosts of the deployment.

- <a id="reg-0920"></a>**REG-0920** The repository has an example seed file with the four preset brands of the data holder: the four legal entities, each with its one brand. ([Judgement: the example seed has only the four preset brands](#reg-0945))
- <a id="reg-0921"></a>**REG-0921** Each record of the example seed has the identifiers, the names and the status of the same record in the lab seed. ([Judgement: the example seed has only the four preset brands](#reg-0945))
- <a id="reg-0922"></a>**REG-0922** The example seed has no data recipients. ([Judgement: the example seed has only the four preset brands](#reg-0945))
- <a id="reg-0923"></a>**REG-0923** Each endpoint of a brand in the example seed is on an invented host of that brand on the top-level domain `.test`. ([RFC 6761, section 6.2][rfc6761-test]; [Judgement: invented `.test` hosts for the endpoints of the example seed](#reg-0946))
- <a id="reg-0924"></a>**REG-0924** The `websiteUri` of each brand in the example seed is the same as in the lab seed. ([Judgement: invented `.test` hosts for the endpoints of the example seed](#reg-0946))
- <a id="reg-0925"></a>**REG-0925** The example seed loads into a Register that has no participants with no error. ([Judgement: invented `.test` hosts for the endpoints of the example seed](#reg-0946))
- <a id="reg-0926"></a>**REG-0926** When the chart deploys a Register with the example seed, Get Data Holder Brands Summary gives the four preset brands at start-up, with their public base URIs. ([CDS: Get Data Holder Brands Summary][cds-get-dh-brands-summary])

## Judgements

### <a id="reg-1504"></a>REG-1504 Each listener keeps the last digit of its port

- Kind: choice
- Basis: Step 3.3. No source gives the ports of a lab. [ADR 0025][adr-0025] gives the listeners.
- Reason: Each listener then has the same last digit in the Register and in each data holder instance. An operator finds each listener from one pattern.

### <a id="reg-0940"></a>REG-0940 No load into a Register with participants

- Kind: choice
- Basis: Step 3.3. The CDS gives no participant load.
- Reason: A merge of two sets can break the status rules or use a REMOVED software product identifier again. A reset is simple in the lab.

### <a id="reg-0941"></a>REG-0941 The load file uses the CDS field names

- Kind: choice
- Basis: Step 3.3. The CDS gives no format for a participant file. The CDS Register schemas give the field names of the API responses ([CDS: Register API schemas][cds-register-schemas]).
- Reason: The lab operator then sees the same names in the file and in the API responses.

### <a id="reg-0942"></a>REG-0942 A ConfigMap as the source of the seed

- Kind: choice
- Basis: Step 3.3. The data holder chart gets [its brand directory from a ConfigMap](../../dataholder/design/deployment.md#dh-1345).
- Reason: With the file in the values, a change of the values after the first load has no effect, which a reader of the values does not expect. A list in a later values file replaces the full list ([Helm: values files][helm-values]), so a deployment must copy the full file to change one host. A seed preset in the chart makes the Register chart depend on the presets of the data holder.

### <a id="reg-0943"></a>REG-0943 The seed in the chart is off by default

- Kind: choice
- Basis: Step 3.3. A Register has no default set of participants.
- Reason: The start of the Register does not need a file, and a deployment with no lab data stays possible. The data holder chart always loads its products, because each brand has a product catalogue.

### <a id="reg-0944"></a>REG-0944 The chart loads the seed file before the Register starts

- Kind: choice
- Basis: Step 3.3. The participant load is lab-only. The data holder chart runs [its product load](../../dataholder/design/deployment.md#dh-1349).
- Reason: The Register is a mock, so its chart can run a lab-only feature. With no load, an operator makes each brand by hand in the operator console.

### <a id="reg-0945"></a>REG-0945 The example seed has only the four preset brands

- Kind: choice
- Basis: Step 3.3. A data holder is ready when its list of software product statuses is empty ([data holder: readiness after the first participant statuses](../../dataholder/design/registration-management.md#dh-0689)).
- Reason: The data recipients of the lab seed are fixtures of the conformance suite, with hosts that are only in the lab. A Register with only the four preset brands is enough to start the data holders.

### <a id="reg-0946"></a>REG-0946 Invented `.test` hosts for the endpoints of the example seed

- Kind: choice
- Basis: Step 3.2. `.test` is reserved for use in testing ([RFC 6761, section 6.2][rfc6761-test]).
- Reason: A URI with a placeholder does not load, because each URI must be absolute with a host ([each URI is absolute with a host](participants.md#reg-0215)), and a file that loads can be tested. The website of a brand is a fixed name of the brand, not a value of one deployment.

[cds-participant-statuses]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-statuses
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-get-dh-brands-summary]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api_get-data-holder-brands-summary
[k8s-configmap-volume]: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
[k8s-init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[helm-values]: https://helm.sh/docs/chart_template_guide/values_files/
[rfc6761-test]: https://www.rfc-editor.org/rfc/rfc6761#section-6.2
[register-chart]: ../../../charts/register/README.md
[readme-conventions]: ../../../README.md#other-conventions
[adr-0010]: ../../../docs/adr/0010-background-jobs-with-river.md
[adr-0020]: ../../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[adr-0027]: ../../../docs/adr/0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md
[adr-0029]: ../../../docs/adr/0029-promote-the-tested-image-of-each-service-to-a-release-with-no-rebuild.md
[adr-0030]: ../../../docs/adr/0030-build-each-service-image-for-linux-amd64-and-linux-arm64-with-a-cross-compile.md
[helm-charts]: https://helm.sh/docs/topics/charts/
[oci-image-index]: https://github.com/opencontainers/image-spec/blob/main/image-index.md
