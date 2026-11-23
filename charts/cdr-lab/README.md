# Umbrella chart

A Helm chart that deploys the lab to one namespace of a Kubernetes cluster with one release: the mock CDR Register and a set of the four preset brands of the mock data holder. It makes the mock CDR CA when it is absent, gets the CDR certificates from the Register and renews them, and makes the Register seed from its values. The requirements are in the [umbrella chart specification](DESIGN.md), and the decisions are in [ADR 0033](../../docs/adr/0033-deploy-the-lab-to-a-cluster-with-one-umbrella-chart-that-derives-each-brand-from-its-register-record.md) and [ADR 0034](../../docs/adr/0034-expose-the-tls-listeners-of-a-cluster-deployment-with-optional-gateway-api-tlsroutes-in-passthrough-mode.md).

A namespace holds one release of the chart. The Services have the fixed names `register` and the name of each brand, for example `wattle-bank`, and the Secrets have fixed names. For a second lab, use a second namespace.

The chart uses the [bjw-s common library chart](https://github.com/bjw-s-labs/helm-charts/tree/main/charts/library/common), as the [Register chart](../register/README.md) and the [data holder chart](../dataholder/README.md) do. It does not use those two charts as subcharts: a subchart gets its values only from the values files, but the chart must give each brand the brand ID and the base URIs of its Register record. Thus the chart builds the containers of the Register and of each brand itself, and a test fails when their variables differ from the service charts.

A brand from a brand directory needs the data holder chart, with its own record in the Register seed.

## Before the install

The chart makes no PostgreSQL, no mail server, no cert-manager and no Gateway. Make these in the namespace of the release before the install:

1. **The databases and the users.** A database for the Register and for each brand, each with an owner user that has `CREATEROLE` and a service user. The migrations create the roles `register_app` and `dataholder_app`, and the operator grants them to the service users ([ADR 0020](../../docs/adr/0020-service-migrations-create-roles-and-the-operator-creates-users.md)). Give each user as a Secret with the keys `username` and `password`. See [PostgreSQL with CloudNativePG](#postgresql-with-cloudnativepg).
2. **The operator token.** A Secret with the token in the key `token`, for example `kubectl create secret generic lab-operator-token --from-literal=token=$(openssl rand -hex 16)`. The Register, each brand and each Job of the chart use it.
3. **The public TLS certificates.** Give a cert-manager issuer in `publicTLS.issuerRef`, or make the Secrets yourself. See [Public TLS certificates](#public-tls-certificates).
4. **A mail server** for the one-time password emails of each brand, for example [Mailpit](https://mailpit.axllent.org). A brand sends no email until a consumer authorises.

Each pod must reach the CRL URL and each base URI that has an external host name. With external host names, the cluster must resolve them, and the pods must reach the Gateway (see [Routes](#routes)).

## Install

`values.yaml` gives each value with a comment. The chart does not render when a required value is missing, and the error names the value. The values of a brand have no field of its Register record: the chart gives the brand ID, the names, the statuses and the base URIs from one source, so a brand cannot disagree with its record. The values file below runs the Register and the four preset brands, with no external host names:

```yaml
operator:
  tokenSecret: lab-operator-token
database:
  host: pg-rw
publicTLS:
  issuerRef: {name: lab-ca, kind: Issuer}
  rootSecret: public-root
register:
  database: {name: register, ownerSecret: register-owner, serviceSecret: register-service}
dataHolders:
  registerStatusInterval: 5s
brands:
  wattle-bank:
    database: {name: wattle_bank, ownerSecret: wattle-bank-owner, serviceSecret: wattle-bank-service}
    mail: {smtpAddress: "mail:1025", from: no-reply@wattlebank.test}
  redgum-bank:
    database: {name: redgum_bank, ownerSecret: redgum-bank-owner, serviceSecret: redgum-bank-service}
    mail: {smtpAddress: "mail:1025", from: no-reply@redgumbank.test}
  harbour-bank:
    database: {name: harbour_bank, ownerSecret: harbour-bank-owner, serviceSecret: harbour-bank-service}
    mail: {smtpAddress: "mail:1025", from: no-reply@harbourbank.test}
  kestrel-bank:
    database: {name: kestrel_bank, ownerSecret: kestrel-bank-owner, serviceSecret: kestrel-bank-service}
    mail: {smtpAddress: "mail:1025", from: no-reply@kestrelbank.test}
```

A release `vX.Y.Z` publishes the chart as `oci://ghcr.io/lukemundy/cdr-lab/charts/cdr-lab`, version `X.Y.Z`, with the `appVersion` `vX.Y.Z`. The default tag of each image (Register, data holder and `cdrlab`) is the `appVersion`. After the cluster job of CI passes on `main`, CI also publishes the chart with the version `0.0.0-g<short commit>`. The default tag of each image in that chart is the [image commit](../../docs/spec/glossary.md#image-commit) of the image. From a checkout, give the image tags, for example the commit of a published image:

```sh
helm install cdr-lab oci://ghcr.io/lukemundy/cdr-lab/charts/cdr-lab --version X.Y.Z \
  --namespace lab --values my-values.yaml --wait --timeout 10m
helm install cdr-lab charts/cdr-lab --namespace lab --values my-values.yaml \
  --set images.register.tag=<commit>,images.dataholder.tag=<commit>,images.cdrlab.tag=<commit> --wait --timeout 10m
```

At the install:

1. The CA Job runs `cdrlab ca init`. It makes the mock CDR CA when its two Secrets are absent, and changes nothing when they exist. The root key goes to no Secret.
2. The Register starts when the CA Secrets and its public TLS Secret exist. Its init containers run the migrations and load the seed, only into a Register with no participants. The Register issues the certificate of its own mTLS listener.
3. The certificate Job of each brand runs `cdrlab cert issue`. It tries again until the Register is ready, and fails after 10 minutes.
4. Each brand starts when its CDR server certificate Secret exists.

A pod with a Secret that is absent does not start, so no step waits for a manual action. A Job that fails has the state `Failed`, and `helm install --wait` stops with an error. Use `--timeout 10m` or more, because the certificate Jobs try for 10 minutes. `kubectl logs job/<name>` gives the cause.

The CronJob `<release>-<brand>-renew` of each brand runs the same command one time each day (`certificates.schedule`). It gets a new certificate when the current one expires within 30 days, or when the CA or the names change. Each service reads its certificate files again within 5 minutes, with no restart.

### Verify an install

```sh
kubectl get pods,jobs --namespace lab
kubectl run check --rm -it --restart=Never --namespace lab --image=docker.io/library/alpine:3.22 -- \
  sh -c 'apk add -q curl && curl -sk -H "x-v: 2" https://register:8081/cdr-register/v1/banking/data-holders/brands/summary'
```

The response gives each brand of the release with its base URIs. Give `-k` only for this check: a client must trust the public root.

## Secrets

| Secret | Writer | Content |
|---|---|---|
| `cdr-ca-intermediate` | CA Job | `kubernetes.io/tls`: the certificate and the key of the mock CDR intermediate CA. Only the Register pod mounts it. |
| `cdr-ca-bundle` | CA Job | `ca.crt`: the intermediate and the root certificate. |
| `<brand>-secure-tls` | certificate Job and CronJob of the brand | `kubernetes.io/tls`: the CDR server certificate of the mTLS listener of the brand, and `ca.crt`. |
| `register-public-tls`, `<brand>-public-tls` | cert-manager, or the team | `kubernetes.io/tls`: the certificate of the public listener. |

The Secrets that the Jobs write stay after an uninstall. Thus an install with the same values and the same databases gives the Register the same CA. To give your own mock CDR CA, make `cdr-ca-intermediate` and `cdr-ca-bundle` before the install. The CA Job does not change a CA that exists, and it fails when only one of the two Secrets exists.

The ServiceAccount of each Job and CronJob can get, update, patch and delete only the Secrets that it writes, and create Secrets. The Register and the brands have no API access.

## Public TLS certificates

With `publicTLS.issuerRef`, the chart makes a cert-manager Certificate for each public listener, with its Secret. The DNS names are the external host name and the Service name of the listener, and also the host name of the online bank site of a brand when the site is on. cert-manager issues no CDR certificate: the Register issues them.

Without an issuer, make a `kubernetes.io/tls` Secret with the names of the table above for the Register and each brand, with the same DNS names.

When the public certificates come from a private CA, give a Secret with only that CA certificate in the key `ca.crt` in `publicTLS.rootSecret`. The Register and each brand then trust that root instead of the system roots, for example when a brand calls the Register at `https://register:8081`. Do not give the Secret of the CA itself: each pod mounts the full Secret, with its key. For example, with a namespaced cert-manager CA:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata: {name: selfsigned}
spec: {selfSigned: {}}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata: {name: public-root-ca}
spec:
  isCA: true
  commonName: cdr-lab public root
  secretName: public-root-ca
  privateKey: {algorithm: ECDSA, size: 256}
  issuerRef: {name: selfsigned, kind: Issuer}
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata: {name: lab-ca}
spec: {ca: {secretName: public-root-ca}}
```

```sh
kubectl create secret generic public-root \
  --from-literal=ca.crt="$(kubectl get secret public-root-ca -o jsonpath='{.data.ca\.crt}' | base64 -d)"
```

## PostgreSQL with CloudNativePG

A [CloudNativePG](https://cloudnative-pg.io) cluster in the namespace, with a managed role for each user and a `Database` for each service. Declare the roles `register_app` and `dataholder_app` (`login: false`) as managed roles too, and give them in `inRoles` of the service users: CloudNativePG does not try a grant again when the role did not exist, and the migrations accept a role that exists. Add one pair of users, Secrets and a `Database` for each brand of the release:

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata: {name: pg}
spec:
  instances: 1
  storage: {size: 1Gi}
  managed:
    roles:
      - {name: register_app, ensure: present, login: false}
      - {name: dataholder_app, ensure: present, login: false}
      - {name: register_owner, ensure: present, login: true, createrole: true, passwordSecret: {name: register-owner}}
      - {name: register, ensure: present, login: true, inRoles: [register_app], passwordSecret: {name: register-service}}
      - {name: wattle_bank_owner, ensure: present, login: true, createrole: true, passwordSecret: {name: wattle-bank-owner}}
      - {name: wattle_bank, ensure: present, login: true, inRoles: [dataholder_app], passwordSecret: {name: wattle-bank-service}}
---
apiVersion: postgresql.cnpg.io/v1
kind: Database
metadata: {name: register}
spec: {name: register, owner: register_owner, cluster: {name: pg}}
---
apiVersion: postgresql.cnpg.io/v1
kind: Database
metadata: {name: wattle-bank}
spec: {name: wattle_bank, owner: wattle_bank_owner, cluster: {name: pg}}
```

```sh
for user in register-owner:register_owner register-service:register \
  wattle-bank-owner:wattle_bank_owner wattle-bank-service:wattle_bank; do
  kubectl create secret generic "${user%%:*}" --type=kubernetes.io/basic-auth \
    --from-literal=username="${user#*:}" --from-literal=password="$(openssl rand -hex 16)"
done
```

The host of the values is the read-write Service of the cluster, `pg-rw`. CloudNativePG serves TLS, so the default `database.sslMode` `require` works.

## Host names and base URIs

Each TLS listener has an external host name and a port: the public and the mTLS listener of the Register and of each brand, and the online bank site of a brand when it is on. With no host name, a listener has its Service name and its port, for example `https://wattle-bank:8083`. With a host name and no port, the port is 443 and the base URI has no port, for example `https://mtls.wattle.lab.example`. The base URIs of the Register records, the settings of the data holders, the certificates and the Host allowlists come from these values.

The brands always call the Register at `https://register:8081`. The CRL URL in each CDR certificate is `http://<public host of the Register>/crl/cdr-intermediate.crl`, or `http://register:8080/crl/cdr-intermediate.crl` with no host name. Each pod and each client must reach it.

## Routes

The routes are off by default, and then the chart makes ClusterIP Services only. With `routes.enabled`, each TLS listener must have a host name, and the chart makes:

- one TLSRoute for each TLS listener, with its host name, for the parent Gateways of `routes.tls.parentRefs`. The Gateway listener must have `tls.mode: Passthrough`, because the services terminate TLS and the mTLS listeners need the client certificate.
- one HTTPRoute for the exact path `/crl/cdr-intermediate.crl` on the public host name of the Register, for the parent Gateways of `routes.crl.parentRefs`. It goes to the internal listener of the Register.

The chart makes no Gateway, no Ingress and no other route to an internal listener. CI renders the routes and checks them against the Gateway API schemas, but runs no Gateway.

A Gateway with one listener on port 443 routes all TLSRoutes by SNI, because each listener has its own host name:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata: {name: lab, namespace: gateway}
spec:
  gatewayClassName: <class>
  listeners:
    - name: tls
      protocol: TLS
      port: 443
      hostname: "*.lab.example"
      tls: {mode: Passthrough}
      allowedRoutes: {namespaces: {from: All}, kinds: [{kind: TLSRoute}]}
    - name: http
      protocol: HTTP
      port: 80
      hostname: register.lab.example
      allowedRoutes: {namespaces: {from: All}, kinds: [{kind: HTTPRoute}]}
```

```yaml
routes:
  enabled: true
  tls: {parentRefs: [{name: lab, namespace: gateway, sectionName: tls}]}
  crl: {parentRefs: [{name: lab, namespace: gateway, sectionName: http}]}
register:
  public: {host: register.lab.example}
  secure: {host: mtls.register.lab.example}
brands:
  wattle-bank:
    public: {host: wattle.lab.example}
    secure: {host: mtls.wattle.lab.example}
```

When the platform gives a port for each listener instead, give the port of the Gateway listener in the values, for example `secure: {host: wattle.lab.example, port: 8443}`, and refer to that listener with `sectionName` or `port` in `routes.tls.parentRefs`. The base URI is then `https://wattle.lab.example:8443`.

With passthrough, the services see the address of the Gateway as the client address ([production gaps](../../docs/production-gaps.md)).

## Operator access

The operator interface and the operator console are on the internal listener (port 8080) of each service, which has no route. Reach them with `kubectl port-forward`, so that Kubernetes RBAC controls who can reach them:

```sh
kubectl port-forward --namespace lab service/register 8080:8080
kubectl port-forward --namespace lab service/wattle-bank 8180:8080
```

A port-forward sends the Host header `localhost:<local port>`. Add it to the Host allowlist of the service, for example `register.internal.allowedHosts: [localhost:8080]` and `brands.wattle-bank.internal.allowedHosts: [localhost:8180]`. Turn on a console with `register.console.enabled` or `brands.<brand>.console.enabled`. The bearer token of the operator interface is in the token Secret.

This Role permits a port-forward in the namespace. Bind it to the operators only: each person with it can use the operator interface and see all the invented consumer data.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata: {name: lab-operator, namespace: lab}
rules:
  - apiGroups: [""]
    resources: [pods, services]
    verbs: [get, list]
  - apiGroups: [""]
    resources: [pods/portforward]
    verbs: [create]
```

For a browser URL to a console, an option is [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview) in the namespace with the upstream `http://register:8080/` (the Host header `register:8080` is in the allowlist), behind your own route with TLS. Use it for the console only. oauth2-proxy cannot send the static operator token, so the operator interface stays on port-forward.

## Upgrade

```sh
helm upgrade cdr-lab oci://ghcr.io/lukemundy/cdr-lab/charts/cdr-lab --version X.Y.Z \
  --namespace lab --values my-values.yaml --wait --timeout 10m
```

The name of each Job has the revision of the release, for example `cdr-lab-ca-r2`. Thus each upgrade runs the Jobs again, also when their image changes, and Helm deletes the Jobs of the earlier revision. The Jobs change nothing when their Secrets are current. The Register and each brand run one pod, and an upgrade stops the old pod before it starts the new pod. Data stays across an upgrade only when the release notes say so: before version 1.0.0, a release can need a fresh deployment.

## Fresh deployment

A fresh deployment deletes all data and the mock CDR CA:

1. `helm uninstall cdr-lab --namespace lab`
2. Delete the Secrets that the Jobs made: `kubectl delete secret --namespace lab cdr-ca-intermediate cdr-ca-bundle <brand>-secure-tls...`
3. Make new, empty databases for the Register and each brand, with the same users.
4. Install the chart again.

## Verify a release

The release workflow on `main` signs the chart and each image with keyless cosign, and attests the provenance of the chart. Verify the chart by the digest from the release notes, as for the [Register chart](../register/README.md#verify-a-release), with `charts/cdr-lab` as the chart name, and verify each of the three images that it deploys.

`ci/test-values.yaml` has the values that the tests and `mise run lint:chart` render, and `ci/lint-values.yaml` turns on each optional part for `lint:chart`. They are not for a deployment.
