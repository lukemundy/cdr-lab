# Register chart

A Helm chart that runs the mock CDR Register on Kubernetes, on the [bjw-s common library chart](https://github.com/bjw-s-labs/helm-charts/tree/main/charts/library/common).

CI publishes the chart from `main` when the chart changes or when CI publishes the Register image, as `oci://ghcr.io/lukemundy/cdr-lab/charts/register`, version `0.0.0-g<short commit>`. Pin the exact version: a version range does not include prerelease versions. The `appVersion` of the chart is the full commit ID of the [image commit](../../docs/spec/glossary.md#image-commit) of the Register, and it is the default of `image.tag`. Thus the chart deploys a published image with no `image.tag` in the values.

```sh
helm install register oci://ghcr.io/lukemundy/cdr-lab/charts/register --version 0.0.0-g<short commit> --values my-values.yaml
```

`values.yaml` gives each value. The chart does not render when a required value is missing.

- The chart makes no Secrets. It uses existing Secrets for the database users, the TLS certificates, the mock CDR intermediate CA, the CDR CA bundle and the operator token. `values.yaml` gives the keys of each Secret.
- PostgreSQL is outside the chart. An init container runs the migrations as the owner user, and the Register connects as the service user ([ADR 0020](../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md)). The first migration creates the role `register_app`. The Register can start only after the operator grants that role to the service user. With CloudNativePG, also declare `register_app` (`login: false`) as a managed role, and give it in `inRoles` of the service user. CloudNativePG does not try a grant again when the role did not exist, and the migration accepts a role that exists.
- The chart runs one pod, and an upgrade stops the old pod before it starts the new pod.
- With no `telemetry.otlpEndpoint`, the pod sends no telemetry. With an endpoint, the Register sends 23 metric series, also when the traffic increases ([ADR 0026](../../docs/adr/0026-send-metrics-only-for-the-state-of-a-process.md)).
- The Register has three listeners, and the container and the Service give one port for each: `internal` (8080, plain HTTP), `public` (8081, TLS) and `secure` (8083, mTLS). The internal listener serves the health probes, the operator interface, the operator console and the CRL ([ADR 0025](../../docs/adr/0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md)).
- The Services are ClusterIP only. The Service of the chart keeps the full name of the release when the values add more Services, because certificates and the CRL URL give that name. The operator console is off by default. Behind a proxy that terminates TLS on port 443, give `internal.allowedHosts` the host name with no port, because HTTP clients do not send the default port in the `Host` header.
- Expose only the CRL path of the internal listener outside the cluster, for example with an ingress rule that routes only `/crl/cdr-intermediate.crl` to the port `internal`. The operator interface and the console are on the same port, and the console has no login. The chart does not make an ingress, so it cannot enforce this rule.

## Participant seed

A Register from the chart starts with no participants. To start it with participants, make a ConfigMap with a seed file in the key `participants.json`, and give its name in `seed.configMap`. The seed is a lab-only feature, and it is off by default. The requirements are in the [Register specification](../../services/register/design/deployment.md#reg-0911).

```sh
kubectl create configmap register-seed --from-file=participants.json=services/register/seed/preset-banks.json
helm install register oci://ghcr.io/lukemundy/cdr-lab/charts/register --version 0.0.0-g<short commit> --values my-values.yaml --set seed.configMap=register-seed
```

- After the migrations, an init container loads the file as the service user, only when the Register has no participants. Thus it needs the grant of `register_app` to the service user, as the Register does. A change of the ConfigMap after the first load has no effect: use the operator interface, or a new database. The load is all or nothing. The Register process does not get the file.
- [`services/register/seed/preset-banks.json`](../../services/register/seed/preset-banks.json) is an example seed with the four preset banks of the data holder and no data recipients. Each brand has one invented host on `.test`, for example `wattlebank.test`, with the ports of the data holder chart: 8081 for the public, product and infosec base URIs, the logo and the JWKS, and 8083 for the resource base URI. A deployment with one host for each brand changes only the domain of these URIs. The `websiteUri` of each brand is a fixed name of the brand. The base URIs of each brand must agree with `publicBaseURL` and `secureBaseURL` of its data holder release. The data holder chart has the values of each brand that agree with this file ([data holder chart](../dataholder/README.md#the-four-preset-brands)).
- The Register issues the CDR server certificate of a data holder brand only for a brand that it has. Thus install the Register with the seed before you make the certificates and the Secrets of the data holder releases.

`ci/test-values.yaml` has the values that the tests and `mise run lint:chart` render, and `ci/lint-values.yaml` turns on each optional part for `lint:chart`. They are not for a deployment.
