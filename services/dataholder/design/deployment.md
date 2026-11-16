# Mock data holder: Deployment

This area gives the four preset brands in the lab, the lab-only features that no other area gives, and the publication of the data holder image and Helm chart. An operator deploys a data holder brand to Kubernetes from the image and the chart that CI publishes. The publication and chart rules apply to the artifacts and to CI, not to a deployment.

## Lab-only features

A production data holder does not have these features. Real components must not depend on them ([README: Component types][readme-component-types]).

- <a id="dh-1300"></a>**DH-1300** The online bank site is a lab-only feature. The lab turns it on for each instance that runs. A production data holder has its own web site. ([online bank site](online-bank-site.md); [README: Component types][readme-component-types])

## Preset brands in the lab

- <a id="dh-1302"></a>**DH-1302** CI runs only [the default preset brand](brand.md#dh-0363). ([Judgement: one preset brand by default](#dh-1360))
- <a id="dh-1303"></a>**DH-1303** With the compose profile `all-banks`, the lab also runs `redgum-bank`, `harbour-bank` and `kestrel-bank`, from the same data holder image. ([the four preset brands at the same time](brand.md#dh-0365); [Judgement: one preset brand by default](#dh-1360); [Judgement: the name of the profile](#dh-1361))
- <a id="dh-1304"></a>**DH-1304** With the profile `all-banks`, each instance is healthy. The Get Products response of each instance gives the product catalogue of its own preset brand. ([the four preset brands at the same time](brand.md#dh-0365))
- <a id="dh-1305"></a>**DH-1305** Each instance has its own database, and a service user that can connect only to that database. ([ADR 0002][adr-0002]; [ADR 0020][adr-0020])
- <a id="dh-1306"></a>**DH-1306** Each instance has its own TLS certificate from the mock public CA, and its own CDR server certificate. ([PKI specification](../../../pki/DESIGN.md); [a public CA certificate on the public listener](api-standards.md#dh-0461); [a CDR CA certificate on the mTLS listener](api-standards.md#dh-0463))
- <a id="dh-1307"></a>**DH-1307** The O of the CDR server certificate of an instance is the legal entity name of its brand. Its CN and a SAN are the host name of the instance. For example, O is "Wattle Bank Limited" and CN is `wattle-bank`. ([PKI specification](../../../pki/DESIGN.md); [a CDR CA certificate on the mTLS listener](api-standards.md#dh-0463))
- <a id="dh-1308"></a>**DH-1308** Each instance has its own sender address for the OTP email: `no-reply@<name>bank.test`. For example, `no-reply@wattlebank.test`. ([the sender of the OTP email](consumer-authentication.md#dh-1244); [Judgement: the names of the Register records of the preset brands](#dh-1365))
- <a id="dh-1309"></a>**DH-1309** The OTP email of each instance shows the brand of that instance. ([the sender of the OTP email](consumer-authentication.md#dh-1244); [the logo in the OTP email](consumer-authentication.md#dh-1252))
- <a id="dh-1310"></a>**DH-1310** On the host, the ports of an instance are a block of six. Each listener has the same last digit in each block. ([Judgement: one port pattern for each instance](#dh-1362))
  - 0: the health probes
  - 1: the public APIs
  - 2: the operator interface
  - 3: the secure APIs
  - 4: the operator console
  - 5: the online bank site
- <a id="dh-1311"></a>**DH-1311** The port blocks of the instances are these. ([Judgement: one port pattern for each instance](#dh-1362))
  - `wattle-bank`: 8180 to 8185
  - `redgum-bank`: 8280 to 8285
  - `harbour-bank`: 8380 to 8385
  - `kestrel-bank`: 8480 to 8485
- <a id="dh-1312"></a>**DH-1312** The databases, the users and the certificates of the four instances exist in each lab, also with no profile. A lab that turns the profile on later needs no reset. ([PostgreSQL image][postgres-image-init]; [Judgement: the set-up of all instances with no profile](#dh-1363))
- <a id="dh-1313"></a>**DH-1313** The Register seed of the lab has a record for each of the four preset brands. Each record has its own legal entity. ([Judgement: Register records of the brands that do not run](#dh-1364))
- <a id="dh-1314"></a>**DH-1314** The legal entity and the brand of each record have the status `ACTIVE`, also with no profile. ([Judgement: Register records of the brands that do not run](#dh-1364))
- <a id="dh-1315"></a>**DH-1315** Each record has the brand name of its preset brand and the base URIs of its instance. ([CDS: Register API schemas][cds-register-schemas])
- <a id="dh-1316"></a>**DH-1316** The legal entity name of each record is the brand name with "Limited". For example, "Wattle Bank Limited". ([CDS: Register API schemas][cds-register-schemas]; [Judgement: the names of the Register records of the preset brands](#dh-1365))
- <a id="dh-1317"></a>**DH-1317** The `websiteUri` of each record is `https://<name>bank.test`. For example, `https://wattlebank.test`. ([CDS: Register API schemas][cds-register-schemas]; [Judgement: the names of the Register records of the preset brands](#dh-1365))
- <a id="dh-1318"></a>**DH-1318** The lab starts with no profile. A service with no profile does not depend on a service of the profile. ([Compose: profiles][compose-profiles])
- <a id="dh-1319"></a>**DH-1319** CI checks that the compose file is valid with the profile `all-banks` on. CI does not start the profile. ([Judgement: no CI job for the other preset brands](#dh-1366))
- <a id="dh-1320"></a>**DH-1320** The conformance suite runs against the default preset brand only. ([Judgement: no CI job for the other preset brands](#dh-1366))
- <a id="dh-1321"></a>**DH-1321** After `mise run down` or `mise run reset`, no container stays of a lab service that the compose file does not define. For example, a container of a renamed service goes. ([Compose: down][compose-down]; [Judgement: the removal of orphan containers](#dh-1367))

## Publication

- <a id="dh-1325"></a>**DH-1325** CI publishes the data holder image from `main` to `ghcr.io/lukemundy/cdr-lab/dataholder`. Its tags are the full and the short commit ID. ([README: Container images][readme-conventions]; [Judgement: one pattern for the published artifacts](#dh-1370))
- <a id="dh-1326"></a>**DH-1326** CI publishes the image of a commit only after these jobs of that commit pass: lint, test, conformance, web and the data holder image. ([GitHub Actions: `needs`][gha-needs]; [Judgement: the publication gate of the image](#dh-1371))
- <a id="dh-1327"></a>**DH-1327** When a job of the gate fails or is cancelled, CI does not publish the image. When CI skips a job because its inputs did not change, CI still publishes the image. ([GitHub Actions: status check functions][gha-status-checks]; [Judgement: the publication gate of the image](#dh-1371))
- <a id="dh-1328"></a>**DH-1328** Each commit on `main` that changes an input of the image gets an image. ([Judgement: the inputs of the image job](#dh-1372))
  - the Go code of the data holder
  - the Go code of `pkg/`
  - `go.mod` and `go.sum`
  - the web workspace
  - the `Dockerfile` of the data holder
- <a id="dh-1329"></a>**DH-1329** The image job of a pull request has the same inputs as the image job of `main`. ([Judgement: the inputs of the image job](#dh-1372))
- <a id="dh-1330"></a>**DH-1330** The data holder in a published image identifies its commit in its telemetry: [`service.version` is the short commit ID of the build](../../../pkg/platform/telemetry/DESIGN.md#tele-0003). The image tag has the same short commit ID. ([`service.version` is the short commit ID](../../../pkg/platform/telemetry/DESIGN.md#tele-0003); [Judgement: one pattern for the published artifacts](#dh-1370))
- <a id="dh-1331"></a>**DH-1331** When the chart changes, CI publishes the chart from `main` as `oci://ghcr.io/lukemundy/cdr-lab/charts/dataholder`. The version is `0.0.0-g<short commit>`. ([Register chart][register-chart]; [Judgement: one pattern for the published artifacts](#dh-1370))
- <a id="dh-1332"></a>**DH-1332** CI publishes the chart of a commit only after the tests of the chart of that commit pass. ([Judgement: the publication gate of the chart](#dh-1373))

## Chart

- <a id="dh-1340"></a>**DH-1340** One image and one chart serve all brands. Brands differ only in the values of the chart and in the brand directory. ([Judgement: one artifact for all brands](#dh-1374))
- <a id="dh-1341"></a>**DH-1341** No default value of the chart is a value of one deployment. For example, the base URIs have no default. ([Judgement: no default for a value of one deployment](#dh-1375))
- <a id="dh-1342"></a>**DH-1342** When a value that each deployment must give is absent, the chart does not render. The error names the value. ([Helm: `required`][helm-required]; [Judgement: no default for a value of one deployment](#dh-1375))
- <a id="dh-1343"></a>**DH-1343** The chart README gives each value and each Secret that a deployment of a brand needs. ([Judgement: Secrets outside the chart](#dh-1376))
- <a id="dh-1344"></a>**DH-1344** The chart makes no Secrets. For example, the values give the name of an existing Secret with the TLS key of the public listener. ([Judgement: Secrets outside the chart](#dh-1376))
- <a id="dh-1345"></a>**DH-1345** The values give exactly one brand: a preset brand by its name, or the name of an existing ConfigMap that has a brand directory. ([Judgement: the brand directory of a data holder (ADR 0023)][adr-0023]; [Judgement: the brand directory as a ConfigMap](#dh-1377))
- <a id="dh-1346"></a>**DH-1346** When the values give no brand or two brands, the chart does not render. ([Judgement: the brand directory as a ConfigMap](#dh-1377))
- <a id="dh-1347"></a>**DH-1347** The data holder gets the files of the ConfigMap as one directory, not file by file. ([Kubernetes: ConfigMap volume][k8s-configmap-volume]; [Judgement: the brand directory as a ConfigMap](#dh-1377))
- <a id="dh-1348"></a>**DH-1348** Before the data holder starts, the chart applies the migrations as the owner user. ([Kubernetes: init containers][k8s-init-containers]; [ADR 0020][adr-0020])
- <a id="dh-1349"></a>**DH-1349** After the migrations complete, the chart loads the product catalogue of the brand as the service user, when the data holder has no products. The data holder starts after the load. ([the product load](product-reference-data.md); [Kubernetes: init containers][k8s-init-containers]; [Judgement: the product load in the chart](#dh-1378))
- <a id="dh-1350"></a>**DH-1350** The chart runs one pod. An upgrade stops the old pod before it starts the new pod. ([Judgement: the state that a restart keeps](non-functional-requirements.md#dh-0022))
- <a id="dh-1351"></a>**DH-1351** The chart sets the Register JWKS URL from the base URI of the public APIs of the Register and the CDS path `/cdr-register/v1/jwks`. For example, for the base URI `https://register.test`, the URL is `https://register.test/cdr-register/v1/jwks`. ([CDS: Register APIs][cds-register-apis]; [Judgement: a derived Register JWKS URL](#dh-1379))
- <a id="dh-1352"></a>**DH-1352** The [interval of the participant statuses](registration.md#dh-0683) is a value with no default. ([Judgement: no default for a value of one deployment](#dh-1375))
- <a id="dh-1353"></a>**DH-1353** The chart has no value for the [fixed OTPs](consumer-authentication.md#dh-1264). ([Judgement: no fixed OTPs in the chart](#dh-1380))
- <a id="dh-1354"></a>**DH-1354** The operator console and the online bank site are off by default. Each has a value that turns it on. ([Judgement: the lab-only UIs off in the chart](#dh-1381))
- <a id="dh-1355"></a>**DH-1355** A deployment of one brand from the published image and chart passes its readiness probe. Get Products and the OpenID Provider configuration then answer on its public listener. ([the readiness probe](../../../pkg/platform/httpserver/DESIGN.md#srv-0002); [the readiness after the first participant statuses](registration.md#dh-0689))
- <a id="dh-1356"></a>**DH-1356** The repository has example values for each of the four preset brands, one set for each brand. ([Judgement: example values for each preset brand](#dh-1382))
- <a id="dh-1357"></a>**DH-1357** Each set of example values has only the values of its brand. A deployment adds the other values. ([Judgement: example values for each preset brand](#dh-1382))
  - the preset brand
  - the brand ID
  - the base URIs of the public APIs and of the secure APIs
  - the sender address of the OTP email
- <a id="dh-1358"></a>**DH-1358** The brand ID and the base URIs of each set agree with the record of the brand in the example seed of the Register. ([Register: Deployment from published artifacts][register-deployment]; [Judgement: example values for each preset brand](#dh-1382))

## Judgements

### <a id="dh-1360"></a>DH-1360 One preset brand by default

- Kind: choice
- Basis: Step 3.3. No source gives the number of brands in a test lab.
- Reason: The lab and CI then need no more memory or time than with one brand.

### <a id="dh-1361"></a>DH-1361 The name of the profile

- Kind: choice
- Basis: Step 3.3. Compose permits any name for a profile ([Compose: profiles][compose-profiles]).
- Reason: The name `all-banks` says what the profile adds, as the names `openid` and `telemetry` do.

### <a id="dh-1362"></a>DH-1362 One port pattern for each instance

- Kind: choice
- Basis: Step 3.3. No source gives the ports of a lab.
- Reason: Ports 8280 to 8485 are not used in the lab. An operator finds each listener of each instance from one pattern.

### <a id="dh-1363"></a>DH-1363 The set-up of all instances with no profile

- Kind: choice
- Basis: Step 3.3. The PostgreSQL image runs its initialisation scripts only when the data directory is empty ([PostgreSQL image][postgres-image-init]).
- Reason: A database that only the profile creates is not there when a lab turns the profile on later. An empty database and three more certificates cost little. The grants of the application role then also do not depend on a service of the profile ([Compose: profiles][compose-profiles]).

### <a id="dh-1364"></a>DH-1364 Register records of the brands that do not run

- Kind: choice
- Basis: Step 3.3. No source gives the Register data of a test lab.
- Reason: The status of a brand on the Register is a fact about its registration, not about whether its instance runs. To a client, a brand that does not run is the same as a brand with an outage, which a data recipient must handle. A seed for each profile needs a second seed file and a second load mode.

### <a id="dh-1365"></a>DH-1365 The names of the Register records of the preset brands

- Kind: choice
- Basis: Step 3.3. The `legalEntityName` has a limit of 36 characters ([CDS: Register API schemas][cds-register-schemas]). The [invented domains](brand.md#dh-0362) use `.test`.
- Reason: An Australian bank is a public company, and its name ends with "Limited". Each name is in the limit.

### <a id="dh-1366"></a>DH-1366 No CI job for the other preset brands

- Kind: choice
- Basis: Step 3.3. No source requires a test of more than one brand.
- Reason: The conformance suite runs against the default preset brand only. A CI job that starts the profile then proves only that three more containers start, at the cost of three more instances on a runner with one node. The lab deployment proves the four instances after the merge.

### <a id="dh-1367"></a>DH-1367 The removal of orphan containers

- Kind: choice
- Basis: Step 3.2. Compose removes the containers of a service that the compose file does not define only with `--remove-orphans` ([Compose: down][compose-down]).
- Reason: A container of a renamed service can keep a port of the lab. Then the next start of the lab fails.

### <a id="dh-1370"></a>DH-1370 One pattern for the published artifacts

- Kind: choice
- Basis: Step 3.3. The repository publishes the image of each service with the full and the short commit ID as tags ([README: Container images][readme-conventions]), and the Register chart with the version `0.0.0-g<short commit>` ([Register chart][register-chart]).
- Reason: All images and charts of the repository then have one pattern. An operator finds the commit of each artifact in the same way.

### <a id="dh-1371"></a>DH-1371 The publication gate of the image

- Kind: choice
- Basis: Step 3.2. A job with `needs` is skipped when a job that it needs fails, unless its condition has a status check function ([GitHub Actions: `needs`][gha-needs], [GitHub Actions: status check functions][gha-status-checks]).
- Reason: CI skips a job only when its inputs did not change, so the result of the previous commit applies. "Tested image" then has one meaning for the images of all services. The conformance job and the image job test the same images, which one job builds for the commit.

### <a id="dh-1372"></a>DH-1372 The inputs of the image job

- Kind: choice
- Basis: Step 3.3. No source gives the path filter of a CI job.
- Reason: With only the inputs of the browser tests, a commit that changes only Go code outside the UI gets no image. One filter for pull requests and `main` is the simplest.

### <a id="dh-1373"></a>DH-1373 The publication gate of the chart

- Kind: choice
- Basis: Step 3.3. No source gives the gate of a chart publication.
- Reason: The chart does not refer to an image of its own commit, because the image tag is a required value. A failure of the tests of the service does not make the chart incorrect.

### <a id="dh-1374"></a>DH-1374 One artifact for all brands

- Kind: choice
- Basis: Step 3.3. The brand directory selects the brand ([ADR 0023][adr-0023]).
- Reason: One artifact for each brand is a second copy of the same code.

### <a id="dh-1375"></a>DH-1375 No default for a value of one deployment

- Kind: choice
- Basis: Step 3.3. Helm can refuse to render when a value is absent ([Helm: `required`][helm-required]).
- Reason: A default that is correct for one deployment is not correct for other deployments. A deployment that omits the value then starts with a value that is not correct.

### <a id="dh-1376"></a>DH-1376 Secrets outside the chart

- Kind: choice
- Basis: Step 3.3. The Register chart makes no Secrets and uses existing Secrets ([Register chart][register-chart]).
- Reason: The operator makes the keys and the certificates with the tools of the deployment, and the two charts work in the same way. The README is then the one list of the Secrets that an operator makes.

### <a id="dh-1377"></a>DH-1377 The brand directory as a ConfigMap

- Kind: choice
- Basis: Step 3.2. Helm reads only the files in the chart, so a published chart cannot make a ConfigMap from a brand directory of the repository ([Helm: chart files][helm-files]). A ConfigMap volume mount with `subPath` gets no updates ([Kubernetes: ConfigMap volume][k8s-configmap-volume]).
- Reason: A brand directory in base64 in the values is difficult to review. The operator makes the ConfigMap with `kubectl create configmap --from-file` from the directory.

### <a id="dh-1378"></a>DH-1378 The product load in the chart

- Kind: choice
- Basis: Step 3.3. No source gives the load of product reference data into a mock data holder.
- Reason: Each preset brand and each brand directory has a product catalogue, and the load changes nothing when products exist. The data holder is a mock, so its chart can run a lab-only feature, as the lab does.

### <a id="dh-1379"></a>DH-1379 A derived Register JWKS URL

- Kind: choice
- Basis: Step 1. The CDS gives the JWKS of the Register at `/cdr-register/v1/jwks` on the base URI of its public APIs ([CDS: Register APIs][cds-register-apis]).
- Reason: The chart then has one value fewer, and the path cannot be different.

### <a id="dh-1380"></a>DH-1380 No fixed OTPs in the chart

- Kind: choice
- Basis: Step 3.3. A fixed OTP is a test feature for the OpenID conformance suite in the lab ([Judgement: fixed OTPs for the OpenID conformance suite](consumer-authentication.md#dh-1288)).
- Reason: A deployment on a cluster does not run that suite. An operator cannot set by mistake a value that the chart does not have.

### <a id="dh-1381"></a>DH-1381 The lab-only UIs off in the chart

- Kind: choice
- Basis: Step 3.3. The Register chart has its console off by default ([Register chart][register-chart]).
- Reason: The operator console and the online bank site are lab-only features. A deployment turns on only the features that it uses.

### <a id="dh-1382"></a>DH-1382 Example values for each preset brand

- Kind: choice
- Basis: Step 3.3. The Register gives data recipients the endpoints of the record of a brand ([CDS: Register APIs][cds-register-apis]).
- Reason: A release with different base URIs becomes ready, because an empty list of statuses is a success, but data recipients cannot find it. The other values are values of the deployment. One set for each brand gives each release one file.

[adr-0002]: ../../../docs/adr/0002-data-store.md
[adr-0020]: ../../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[adr-0023]: ../../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md
[cds-register-apis]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-apis
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[compose-down]: https://docs.docker.com/reference/cli/docker/compose/down/
[compose-profiles]: https://docs.docker.com/reference/compose-file/profiles/
[gha-needs]: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idneeds
[gha-status-checks]: https://docs.github.com/en/actions/reference/workflows-and-actions/expressions#status-check-functions
[helm-files]: https://helm.sh/docs/chart_template_guide/accessing_files/
[helm-required]: https://helm.sh/docs/howto/charts_tips_and_tricks/#using-the-required-function
[k8s-configmap-volume]: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
[k8s-init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[postgres-image-init]: https://hub.docker.com/_/postgres
[readme-component-types]: ../../../README.md#component-types
[readme-conventions]: ../../../README.md#other-conventions
[register-chart]: ../../../charts/register/README.md
[register-deployment]: ../../register/DESIGN.md#deployment-from-published-artifacts
