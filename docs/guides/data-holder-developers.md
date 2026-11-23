# Data holder developers

This guide is for a team that develops data holder software and uses the lab as the CDR Register. It gives the Register side only: the onboarding of your data holder brand, its server certificate, and the Register APIs that your software calls.

The lab has no mock data recipient yet. Thus no software in the lab calls your data holder. Use your own test client, with a client certificate of a software product that you add as in the [ADR developers guide](adr-developers.md).

The commands use the compose lab ([install guide](install.md)). They run in the lab directory, and the Register operator interface is at `http://localhost:8080` with the token `lab-operator-token`. For a cluster deployment, see [a cluster deployment](#a-cluster-deployment).

## Onboard your data holder brand

The onboarding is two commands: add the participant records, then get the server certificate of your brand ([cdrlab: the onboarding of a data holder brand](../../cli/cdrlab/design/onboarding.md#cli-0191)).

1. Write a participant load file with your legal entity and its brand. The format is the seed file of the Register ([`services/register/seed/example.json`](../../services/register/seed/example.json)). The base URIs are the URLs of your software, for example on the host. Use invented data only:

   ```json
   {"dataHolders": [{
     "legalEntityId": "3d836f62-0fe0-4e7d-9d36-eee05bd8c827",
     "legalEntityName": "Example Mutual Limited",
     "logoUri": "https://mutual.example/logo.svg",
     "status": "ACTIVE",
     "dataHolderBrands": [{
       "dataHolderBrandId": "6862797b-9bef-48fe-b3ac-61d03044063b",
       "brandName": "Example Mutual",
       "industries": ["banking"],
       "logoUri": "https://mutual.example/logo.svg",
       "status": "ACTIVE",
       "endpointDetail": {
         "version": "v1",
         "publicBaseUri": "https://localhost:9443",
         "productBaseUri": "https://localhost:9443",
         "resourceBaseUri": "https://localhost:9444",
         "infosecBaseUri": "https://localhost:9443",
         "websiteUri": "https://mutual.example"
       },
       "jwksEndpoint": "https://localhost:9443/jwks"
     }]
   }]}
   ```

2. Add the records. The command adds each record that the Register does not have, and changes no record that exists ([Register: the participant load](../../services/register/design/operator-interface.md#reg-1531)):

   ```sh
   export REGISTER_OPERATOR_TOKEN=lab-operator-token
   ./cdrlab participants add -register-url http://localhost:8080 -file example-mutual.json
   ```

3. Get the CDR server certificate of the mTLS listener of your brand. The Register issues it from a CSR, and `cdrlab` keeps the key on your host:

   ```sh
   mkdir -p mutual
   ./cdrlab cert issue -register-url http://localhost:8080 -kind server \
     -legal-entity-id 3d836f62-0fe0-4e7d-9d36-eee05bd8c827 -brand-id 6862797b-9bef-48fe-b3ac-61d03044063b \
     -cn localhost -o "Example Mutual Limited" -dns localhost -cert mutual/secure.crt -key mutual/secure.key
   ```

   Run the same command again before the certificate expires: it gets a new certificate only in the renewal window ([cdrlab: a renewal window of 30 days](../../cli/cdrlab/design/certificates.md#cli-0101)).

Your brand is now in the Register APIs:

```sh
http --verify .lab/pki/public/root.crt https://localhost:8081/cdr-register/v1/banking/data-holders/brands/summary x-v:2 |
  jq '.data[] | select(.brandName == "Example Mutual")'
```

To change or remove a record, use the operator console of the Register (`http://localhost:8080`), or make a new lab: a load does not change a record that exists.

## The trust of your software

- Your mTLS listener trusts the CDR CA bundle `.lab/pki/cdr/ca-bundle.pem` for client certificates, and serves `mutual/secure.crt`.
- The certificates of the CDR CA give the CRL URL of the Register. From the host, the CRL is at `http://localhost:8080/crl/cdr-intermediate.crl`.
- Your software trusts the public root `.lab/pki/public/root.crt` for the TLS of the Register public APIs.

## The Register APIs that a data holder calls

The public APIs of the Register need no authentication. From the host they are at `https://localhost:8081`. From a container on the compose network, they are at `https://register:8081`. The [Register APIs specification](../../services/register/design/register-apis.md) gives the versions and the rules:

```sh
root=.lab/pki/public/root.crt
http --verify $root https://localhost:8081/cdr-register/v1/banking/data-recipients x-v:4
http --verify $root https://localhost:8081/cdr-register/v1/banking/data-recipients/status x-v:3
http --verify $root https://localhost:8081/cdr-register/v1/banking/data-recipients/brands/software-products/status x-v:3
http --verify $root https://localhost:8081/cdr-register/v1/jwks
```

Your software uses the JWKS to verify the software statement assertion (SSA) of a registration ([Register: the SSA](../../services/register/design/software-statement-assertion.md#reg-0401)).

## A cluster deployment

In a cluster deployment, reach the Register operator interface with `kubectl port-forward`, and add `localhost:8080` to its Host allowlist ([operator access](../../charts/cdr-lab/README.md#operator-access)). Then run the same commands, with the token of the operator token Secret. Get the CA bundle and the public root from their Secrets:

```sh
kubectl port-forward --namespace lab service/register 8080:8080 &
export REGISTER_OPERATOR_TOKEN=$(kubectl get secret --namespace lab lab-operator-token -o jsonpath='{.data.token}' | base64 -d)
kubectl get secret --namespace lab cdr-ca-bundle -o jsonpath='{.data.ca\.crt}' | base64 -d > ca-bundle.pem
```

The base URIs of your brand must then be URLs that the clients in the cluster can reach.
