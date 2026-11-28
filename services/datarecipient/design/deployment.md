# Mock data recipient: Deployment

This area gives the preset brand in the lab, its ports, its database and its mail, the image and the chart, and the deployment from the umbrella chart.

## Lab deployment

- <a id="dr-0600"></a>**DR-0600** The lab runs the default preset brand `sorrel` with no profile. ([the default preset brand](brand.md#dr-0121))
- <a id="dr-0601"></a>**DR-0601** In the lab, the compose service, the host name, the database and the certificate directory of the instance have the name of its preset brand. For example, the compose service of `sorrel` is `sorrel`. ([the names of the instances of the data holder](../../dataholder/design/brand.md#dh-0360))
- <a id="dr-0602"></a>**DR-0602** On the host, the ports of an instance are in a block, and each listener has the same last digit as in each other service. ([Judgement: a web app listener on a data recipient (ADR 0025)][adr-0025])

  | Instance | Internal listener | Web app listener |
  |---|---|---|
  | `sorrel` | 8580 | 8585 |

- <a id="dr-0603"></a>**DR-0603** Each port binds to `127.0.0.1` only. ([the lab ports](../../dataholder/design/deployment.md#dh-1310))
- <a id="dr-0604"></a>**DR-0604** The instance has its own database, an owner user that runs the migrations, and a service user that can connect only to that database. The migrations create the application role. ([ADR 0002][adr-0002]; [ADR 0020][adr-0020])
- <a id="dr-0605"></a>**DR-0605** The instance has its own TLS certificate from the mock public CA for the web app listener. Its SAN has `localhost` and the host name of the instance. ([PKI specification](../../../pki/DESIGN.md))
- <a id="dr-0606"></a>**DR-0606** The instance sends its emails to the mail sink of the lab, from `no-reply@sorrel.test`. ([the mail settings](email.md#dr-0500); [`.test` for an invented domain](brand.md#dr-0125))
- <a id="dr-0607"></a>**DR-0607** The `WEB_APP_BASE_URI` of the instance in the lab is `https://localhost:8585`, so that the links of an email open in a browser on the host. ([the base URI of the web app](web-app.md#dr-0207))
- <a id="dr-0608"></a>**DR-0608** `mise run up` builds and starts the instance, and waits until it is healthy. `mise run reset` deletes its data. ([the health probes](non-functional-requirements.md#dr-0014))

## Image and chart

- <a id="dr-0610"></a>**DR-0610** CI publishes the data recipient image from `main` to `ghcr.io/lukemundy/cdr-lab/datarecipient`, with the gate, the tags and the inputs of the data holder image. ([the publication of the data holder image](../../dataholder/design/deployment.md#dh-1325); [the gate](../../dataholder/design/deployment.md#dh-1326); [the inputs](../../dataholder/design/deployment.md#dh-1328))
- <a id="dr-0611"></a>**DR-0611** The image is built for `linux/amd64` and `linux/arm64`. ([ADR 0030][adr-0030])
- <a id="dr-0612"></a>**DR-0612** CI publishes the chart of the data recipient from `main` as `oci://ghcr.io/lukemundy/cdr-lab/charts/datarecipient`, with the rules of the chart of the data holder. ([the publication of the data holder chart](../../dataholder/design/deployment.md#dh-1331); [the gate of the chart](../../dataholder/design/deployment.md#dh-1332))
- <a id="dr-0613"></a>**DR-0613** One image and one chart serve all brands. Brands differ only in the values of the chart and in the brand directory. ([one artifact for all brands of the data holder](../../dataholder/design/deployment.md#dh-1340))
- <a id="dr-0614"></a>**DR-0614** The chart runs one pod, and it stops the pod before it starts a new pod. ([Judgement: sessions in the memory of the process](sign-in.md#dr-0453))
- <a id="dr-0615"></a>**DR-0615** The umbrella chart deploys the default preset brand of the data recipient. ([Judgement: one umbrella chart for the lab (ADR 0033)][adr-0033])
- <a id="dr-0616"></a>**DR-0616** In a cluster deployment, the values of the chart give the mail server, the public TLS certificate and the base URI of the web app. The chart has no default for them. ([Judgement: settings with no lab default](#dr-0650))

## CI

- <a id="dr-0620"></a>**DR-0620** CI runs the browser tests of the web app against the image of the commit in the lab deployment. The tests read the emails from the mail sink. ([Judgement: browser tests that read the mail sink](#dr-0651))
- <a id="dr-0621"></a>**DR-0621** The browser tests make their own members, with an email address that is unique to the test run. ([Judgement: browser tests that read the mail sink](#dr-0651))

## Judgements

### <a id="dr-0650"></a>DR-0650 Settings with no lab default

- Kind: choice
- Basis: Step 3.3. A setting with a lab value as its default works in the lab and fails in a different deployment with no error.
- Reason: A required setting with no default fails at start-up with an error that names it. The compose file and the example values give the lab values.

### <a id="dr-0651"></a>DR-0651 Browser tests that read the mail sink

- Kind: choice
- Basis: Step 3.3. A sign-up and a password reset need a link that the data recipient sends by email ([the verification email](email.md#dr-0510)).
- Reason: A test that reads the link from the mail sink tests the full journey, with no test-only path in the data recipient. A unique address for each run lets the test find its own email.

[adr-0002]: ../../../docs/adr/0002-data-store.md
[adr-0020]: ../../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md
[adr-0025]: ../../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md
[adr-0030]: ../../../docs/adr/0030-build-each-service-image-for-linux-amd64-and-linux-arm64-with-a-cross-compile.md
[adr-0033]: ../../../docs/adr/0033-deploy-the-lab-to-a-cluster-with-one-umbrella-chart-that-derives-each-brand-from-its-register-record.md
