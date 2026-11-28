# Mock data recipient: Deployment

This area gives the preset brand in the lab, its ports, its database and its mail, the image and the chart, and the deployment from the umbrella chart.

## Lab deployment

- <a id="dr-0600"></a>**DR-0600** The lab runs the default preset brand `sorrel` with no profile. ([the default preset brand](brand.md#dr-0121))
- <a id="dr-0601"></a>**DR-0601** In the lab, the compose service, the host name, the database and the certificate directory of the instance have the name of its preset brand. For example, the compose service of `sorrel` is `sorrel`. ([the names of the instances of the data holder](../../dataholder/design/brand.md#dh-0360))
- <a id="dr-0602"></a>**DR-0602** On the host, the ports of an instance are in a block, and each listener has the same last digit as in each other service. ([Judgement: a web app listener on a data recipient (ADR 0025)][adr-0025])

  | Instance | Internal listener | Public listener | Web app listener |
  |---|---|---|---|
  | `sorrel` | 8580 | 8581 | 8585 |

- <a id="dr-0637"></a>**DR-0637** In the container, each listener has the port that the Register gives to the listener with the same last digit: 8080 for the internal listener, 8081 for the public listener and 8085 for the web app listener. For example, the JWKS URI on the lab network is `https://sorrel:8081/jwks`, and the host reaches the JWKS at `https://localhost:8581/jwks`. ([the ports of the Register](../../register/design/deployment.md#reg-1500); [Judgement: a web app listener on a data recipient (ADR 0025)][adr-0025])
- <a id="dr-0603"></a>**DR-0603** Each port binds to `127.0.0.1` only. ([the lab ports](../../dataholder/design/deployment.md#dh-1310))
- <a id="dr-0604"></a>**DR-0604** The instance has its own database, an owner user that runs the migrations, and a service user that can connect only to that database. The migrations create the application role. ([ADR 0002][adr-0002]; [ADR 0020][adr-0020])
- <a id="dr-0605"></a>**DR-0605** The instance has its own TLS certificate from the mock public CA for the web app listener and the public listener. Its SAN has `localhost` and the host name of the instance. ([PKI specification](../../../pki/DESIGN.md); [the certificate of the public listener](participation.md#dr-0706))
- <a id="dr-0606"></a>**DR-0606** The instance sends its emails to the mail sink of the lab, from `no-reply@sorrel.test`. ([the mail settings](email.md#dr-0500); [`.test` for an invented domain](brand.md#dr-0125))
- <a id="dr-0607"></a>**DR-0607** The `WEB_APP_BASE_URI` of the instance in the lab is `https://localhost:8585`, so that the links of an email open in a browser on the host. ([the base URI of the web app](web-app.md#dr-0207))
- <a id="dr-0608"></a>**DR-0608** `mise run up` builds and starts the instance, and waits until it is healthy. `mise run reset` deletes its data. ([the health probes](non-functional-requirements.md#dr-0014))

## Participant in the lab

- <a id="dr-0630"></a>**DR-0630** The Register seed of the lab has a record of `sorrel`: a legal entity, one brand with the brand ID of the instance, and one software product. Each has the status `ACTIVE`. ([CDS: Register API schemas][cds-register-schemas]; [the Register records of the data holder](../../dataholder/design/deployment.md#dh-1313))
- <a id="dr-0631"></a>**DR-0631** The legal entity of `sorrel` has the name "Sorrel Pty Ltd", an invented accreditation number and the accreditation level `UNRESTRICTED`. ([the accreditation of a data recipient](../../register/design/participants.md#reg-0211); [names with no "Mock"](brand.md#dr-0124))
- <a id="dr-0632"></a>**DR-0632** The software product of `sorrel` in the lab has these URIs. ([CDS: SSA Definition][cds-ssa-definition]; [Judgement: the redirect URI on the web app, and the JWKS on the public listener](participation.md#dr-0750); [Judgement: the CDR policy as the policy and the terms of the software product](#dr-0652))

  | Field | URI |
  |---|---|
  | `redirectUris` | `https://localhost:8585/connect/callback` |
  | `jwksUri` | `https://sorrel:8081/jwks` |
  | `recipientBaseUri` | `https://sorrel:8081` |
  | `revocationUri` | `https://sorrel:8081/arrangements/revoke` |
  | `clientUri` | `https://localhost:8585` |
  | `policyUri` and `tosUri` | the URI of the CDR policy page of the web app |
  | `logoUri` of the legal entity, the brand and the software product | the URI of the logo that the web app serves |

- <a id="dr-0633"></a>**DR-0633** The scopes of the software product of `sorrel` are `openid`, `profile`, `bank:accounts.basic:read`, `bank:transactions:read` and `cdr:registration`. ([the scopes that the Register requires](../../register/design/participants.md#reg-1496); [Judgement: the account cluster is required and the transaction cluster is optional](consent.md#dr-0851))
- <a id="dr-0634"></a>**DR-0634** `cdrlab up` gets a client certificate for the software product of `sorrel` from the Register before the instance starts. The certificate list of the lab has its entry. ([Judgement: CDR certificates from the Register (ADR 0032)][adr-0032]; [Judgement: CDR certificates issued from the host (ADR 0038)][adr-0038])
- <a id="dr-0635"></a>**DR-0635** In the lab, `REGISTER_CACHE_INTERVAL` is 1 minute, so that a bank that a lab user adds shows in the web app within 1 minute. ([Judgement: a cache interval of the Register in the configuration](participation.md#dr-0753))
- <a id="dr-0636"></a>**DR-0636** The instance calls the Register and each bank at their compose service names, on the lab network. ([the names of the instances of the data holder](../../dataholder/design/brand.md#dh-0360))

## Image and chart

- <a id="dr-0610"></a>**DR-0610** CI publishes the data recipient image from `main` to `ghcr.io/lukemundy/cdr-lab/datarecipient`, with the gate, the tags and the inputs of the data holder image. ([the publication of the data holder image](../../dataholder/design/deployment.md#dh-1325); [the gate](../../dataholder/design/deployment.md#dh-1326); [the inputs](../../dataholder/design/deployment.md#dh-1328))
- <a id="dr-0611"></a>**DR-0611** The image is built for `linux/amd64` and `linux/arm64`. ([ADR 0030][adr-0030])
- <a id="dr-0612"></a>**DR-0612** CI publishes the chart of the data recipient from `main` as `oci://ghcr.io/lukemundy/cdr-lab/charts/datarecipient`, with the rules of the chart of the data holder. ([the publication of the data holder chart](../../dataholder/design/deployment.md#dh-1331); [the gate of the chart](../../dataholder/design/deployment.md#dh-1332))
- <a id="dr-0613"></a>**DR-0613** One image and one chart serve all brands. Brands differ only in the values of the chart and in the brand directory. ([one artifact for all brands of the data holder](../../dataholder/design/deployment.md#dh-1340))
- <a id="dr-0614"></a>**DR-0614** The chart runs one pod, and it stops the pod before it starts a new pod. ([Judgement: sessions in the memory of the process](sign-in.md#dr-0453))
- <a id="dr-0615"></a>**DR-0615** The umbrella chart deploys the default preset brand of the data recipient. ([Judgement: one umbrella chart for the lab (ADR 0033)][adr-0033])
- <a id="dr-0616"></a>**DR-0616** In a cluster deployment, the values of the chart give the mail server, the public TLS certificates, the base URI of the web app, the base URI of the public listener, the software product ID and the cache interval of the Register. The chart has no default for them. ([Judgement: settings with no lab default](#dr-0650))
- <a id="dr-0617"></a>**DR-0617** The umbrella chart has the Register record of the default preset brand of the data recipient, and makes its URIs from the host names of the web app listener and of the public listener. ([the Register record of the data recipient in the umbrella chart](../../../charts/cdr-lab/DESIGN.md#lab-0136))
- <a id="dr-0618"></a>**DR-0618** In a cluster deployment, each TLS listener of the data recipient has one external host name, which the browser and the pods both reach. ([the pods reach each external host name](../../../charts/cdr-lab/DESIGN.md#lab-0037); [Judgement: separate host names on port 443 (ADR 0034)][adr-0034])

## CI

- <a id="dr-0620"></a>**DR-0620** CI runs the browser tests of the web app against the image of the commit in the lab deployment. The tests read the emails from the mail sink. ([Judgement: browser tests that read the mail sink](#dr-0651))
- <a id="dr-0621"></a>**DR-0621** The browser tests make their own members, with an email address that is unique to the test run. ([Judgement: browser tests that read the mail sink](#dr-0651))

## Judgements

### <a id="dr-0650"></a>DR-0650 Settings with no lab default

- Kind: choice
- Basis: Step 3.3. A setting with a lab value as its default works in the lab and fails in a different deployment with no error.
- Reason: A required setting with no default fails at start-up with an error that names it. The compose file and the example values give the lab values.

### <a id="dr-0652"></a>DR-0652 The CDR policy as the policy and the terms of the software product

- Kind: choice
- Basis: Step 3.3. The SSA has a `policy_uri` and a `tos_uri` ([CDS: SSA Definition][cds-ssa-definition]). The web app has a CDR policy and a privacy policy, and no terms of service ([the pages of the web app](web-app.md#dr-0212)).
- Reason: The CDR policy is the document that tells a consumer how the brand treats CDR data. A terms page for a demonstration adds text with no use.

### <a id="dr-0651"></a>DR-0651 Browser tests that read the mail sink

- Kind: choice
- Basis: Step 3.3. A sign-up and a password reset need a link that the data recipient sends by email ([the verification email](email.md#dr-0510)).
- Reason: A test that reads the link from the mail sink tests the full journey, with no test-only path in the data recipient. A unique address for each run lets the test find its own email.

[adr-0002]: ../../../docs/adr/0002-data-store.md
[adr-0020]: ../../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[adr-0030]: ../../../docs/adr/0030-build-each-service-image-for-linux-amd64-and-linux-arm64-with-a-cross-compile.md
[adr-0032]: ../../../docs/adr/0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md
[adr-0033]: ../../../docs/adr/0033-deploy-the-lab-to-a-cluster-with-one-umbrella-chart-that-derives-each-brand-from-its-register-record.md
[adr-0034]: ../../../docs/adr/0034-expose-the-tls-listeners-of-a-cluster-deployment-with-optional-gateway-api-tlsroutes-in-passthrough-mode.md
[adr-0038]: ../../../docs/adr/0038-run-the-compose-lab-with-cdrlab-from-a-compose-project-in-the-binary-and-issue-its-cdr-certificates-from-the-host.md
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[cds-ssa-definition]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#ssa-definition
