# Mock CDR Register: Deployment

This area gives the lab-only participant load, the participant seed of the Helm chart, the example seed of the four preset brands, and the publication of the image and the chart. The production Register has no participant load. Real components must not depend on it. The rules of the chart and of the example seed apply to the chart and to the file, not to a deployment.

## Participant load

- <a id="reg-0900"></a>**REG-0900** The lab operator can load a full set of participant data from a file into a Register that has no participants. ([Judgement: no load into a Register with participants](#reg-0940))
- <a id="reg-0901"></a>**REG-0901** When the Register has participants, the load refuses the file and changes nothing. ([Judgement: no load into a Register with participants](#reg-0940))
- <a id="reg-0902"></a>**REG-0902** The load is separate from the start of the Register. The start of the Register does not need a file. ([Judgement: the seed in the chart is off by default](#reg-0943))
- <a id="reg-0903"></a>**REG-0903** The load checks the full set against [the participant records](participants.md#reg-0200), [the status transitions](participants.md#reg-0222) and [the cascades](participants.md#reg-0227). ([CDS: Participant Statuses][cds-participant-statuses])
- <a id="reg-0904"></a>**REG-0904** When one record is not valid, the load changes nothing, and it names each record and rule that fails. ([Judgement: no load into a Register with participants](#reg-0940))
- <a id="reg-0905"></a>**REG-0905** The load file uses the field names of the CDS Register schemas. For example, a software product has the field `softwareProductId`. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: the load file uses the CDS field names](#reg-0941))
- <a id="reg-0906"></a>**REG-0906** The load is a command of the Register binary, and it needs access to the Register database. ([Judgement: no load into a Register with participants](#reg-0940))

## Publication

- <a id="reg-0910"></a>**REG-0910** An operator deploys the Register to Kubernetes from the image and the Helm chart that CI publishes from `main`. ([Register chart][register-chart]; [README: Container images][readme-conventions])

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
- Basis: Step 3.3. A data holder is ready when its list of software product statuses is empty ([data holder: readiness after the first participant statuses](../../dataholder/design/registration.md#dh-0689)).
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
[adr-0020]: ../../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
