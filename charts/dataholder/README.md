# Data holder chart

A Helm chart that runs one brand of the mock data holder on Kubernetes, on the [bjw-s common library chart](https://github.com/bjw-s-labs/helm-charts/tree/main/charts/library/common). One image and one chart serve all brands: a second brand is a second release with different values. The requirements are in the [data holder deployment specification](../../services/dataholder/design/deployment.md).

CI publishes the chart from `main` when a chart changes, as `oci://ghcr.io/lukemundy/cdr-lab/charts/dataholder`, version `0.0.0-g<short commit>`. Pin the exact version: a version range does not include prerelease versions.

```sh
helm install wattle-bank oci://ghcr.io/lukemundy/cdr-lab/charts/dataholder --version 0.0.0-g<short commit> --values my-values.yaml
```

## Values

The chart does not render when a required value is missing, and the error names the value. `values.yaml` gives each value with a comment.

| Value | Required | Data holder setting |
|---|---|---|
| `image.tag` | yes | The commit of the image `image.repository` (default `ghcr.io/lukemundy/cdr-lab/dataholder`). `image.digest` is optional. |
| `brand.preset` or `brand.configMap` | exactly one | `DATA_HOLDER_BRAND`, or `DATA_HOLDER_BRAND_DIR` with the ConfigMap at `/brand` |
| `brand.id` | yes | `DATA_HOLDER_BRAND_ID`: the `dataHolderBrandId` of the brand on the Register |
| `publicBaseURL`, `secureBaseURL` | yes | `PUBLIC_BASE_URL` (also the issuer), `SECURE_BASE_URL` |
| `crlMaxAge` | no (default of the data holder: `1m`) | `CRL_MAX_AGE`, `0s` or more |
| `register.publicBaseURL` | yes | `REGISTER_PUBLIC_BASE_URL`, and `REGISTER_JWKS_URL` with the CDS path `/cdr-register/v1/jwks` |
| `register.statusInterval` | yes | `REGISTER_STATUS_INTERVAL`, more than `0s` and at most `4m` |
| `mail.smtpAddress`, `mail.from` | yes | `SMTP_ADDRESS`, `MAIL_FROM` |
| `operator.allowedHosts` | yes | `OPERATOR_ALLOWED_HOSTS` |
| `console.enabled`, `console.allowedHosts` | hosts when on (default off) | `CONSOLE_ENABLED`, `CONSOLE_ALLOWED_HOSTS`, port 8084 |
| `bankSite.enabled`, `bankSite.allowedHosts` | hosts when on (default off) | `BANK_SITE_ENABLED`, `BANK_SITE_ALLOWED_HOSTS`, port 8085 |
| `telemetry.otlpEndpoint` | yes | `OTEL_EXPORTER_OTLP_ENDPOINT` |
| `database.host`, `database.name` | yes | `PGHOST`, `PGDATABASE`. `database.port` (5432) and `database.sslMode` (`require`) have defaults. |

The console and the online bank site are lab-only features. The chart has no value for the fixed one-time passwords, which are only for the OpenID conformance suite in the lab.

### The four preset brands

`examples/<preset>.yaml` has the values of each preset brand: `brand.preset`, `brand.id`, `publicBaseURL`, `secureBaseURL` and `mail.from`. Each file agrees with the record of its brand in the example seed of the Register, [`services/register/seed/preset-banks.json`](../../services/register/seed/preset-banks.json) ([Register chart](../register/README.md#participant-seed)). Give the file and the values of the deployment, for example `--values charts/dataholder/examples/wattle-bank.yaml --values my-values.yaml` from the repository root. When the deployment changes the base URLs, change them also in the seed. With a `brand.id` or base URLs that are not the same as the record of the brand on the Register, the pod becomes ready, but data recipients cannot find the brand.

## Secrets

The chart makes no Secrets. Give the name of each existing Secret in the values.

| Value | Keys | Use |
|---|---|---|
| `database.ownerSecret` | `username`, `password` | The owner user: only the `migrate` init container |
| `database.serviceSecret` | `username`, `password` | The service user: the product load and the data holder |
| `pki.publicTLSSecret` | `tls.crt`, `tls.key` | The public listener (8081) and the online bank site (8085) |
| `pki.secureTLSSecret` | `tls.crt`, `tls.key` | The secure listener (8083), a CDR server certificate |
| `pki.caBundleSecret` | `ca.crt` | The CDR CA bundle for the client certificates |
| `operator.tokenSecret` | `token` | The bearer token of the operator interface |
| `pki.publicRootSecret` (optional) | `ca.crt` | The root that the data holder trusts when it calls the Register and software products, instead of the system roots |

A brand directory in `brand.configMap` is an existing ConfigMap, for example from `kubectl create configmap <name> --from-file=<brand directory>`. A ConfigMap holds at most 1 MiB, and the data holder refuses a brand directory over 1 MiB in all. `logo.png` and the photos go into `binaryData`. A brand directory with photos is larger than the 256 KiB that a client-side `kubectl apply` permits for its `kubectl.kubernetes.io/last-applied-configuration` annotation, so use `kubectl create`, or `kubectl apply --server-side` ([Kubernetes: Server-Side Apply](https://kubernetes.io/docs/reference/using-api/server-side-apply/)).

## Operation

- PostgreSQL is outside the chart. Before the data holder starts, an init container runs the migrations as the owner user, and then a second init container loads the product catalogue of the brand as the service user when the data holder has no products ([ADR 0020](../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md)). The migrations create the role `dataholder_app`. The product load and the data holder can start only after the operator grants that role to the service user. A role is global to the PostgreSQL server, so all brands on one server share it. With CloudNativePG, also declare `dataholder_app` (`login: false`) as a managed role, and give it in `inRoles` of the service user. CloudNativePG does not try a grant again when the role did not exist, and the migration accepts a role that exists.
- The pod is ready only after the data holder gets the statuses of the software products from a Register that it trusts. A Register outage makes the pod not ready.
- The chart runs one pod, and an upgrade stops the old pod before it starts the new pod. A restart ends each authorisation flow and each session of the online bank site.
- The Services are ClusterIP only. The Service of the chart keeps the full name of the release when the values add more Services, because the certificates and the base URLs give that name. The data holder terminates TLS on the public and the secure listeners. The secure listener needs mTLS from end to end, so a proxy that terminates TLS cannot be in front of port 8083: use TCP pass-through. Behind a proxy that terminates TLS on port 443, give the `allowedHosts` values the host name with no port, because HTTP clients do not send the default port in the `Host` header.

`ci/test-values.yaml` has the values that the tests and `mise run lint:chart` render, and `ci/lint-values.yaml` turns on each optional part for `lint:chart`. They are not for a deployment.
