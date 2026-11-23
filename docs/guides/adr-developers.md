# ADR developers

This guide is for a team that develops the software product of an accredited data recipient (ADR) and tests it against the lab. It gives the onboarding of the software product, the test customers, the sign-in of a consumer and the seed data of a cluster deployment.

The commands use the compose lab ([install guide](install.md)). They run in the lab directory. The operator interfaces are at `http://localhost:8080` (Register) and `http://localhost:8180` (Wattle Bank), with the token `lab-operator-token`. For a cluster deployment, see [a cluster deployment](#a-cluster-deployment).

## Onboard your software product

The onboarding is two commands: add the participant records, then get the client certificate of the software product. The Register registers the certificate when it issues it ([cdrlab: the onboarding of a software product](../../cli/cdrlab/design/onboarding.md#cli-0190)).

1. Write a participant load file with your data recipient, its brand and its software product. The format is the seed file of the Register ([`services/register/seed/example.json`](../../services/register/seed/example.json)). Use invented data only. For example, `example-budget.json`:

   ```json
   {"dataRecipients": [{
     "legalEntityId": "6f6c2a04-6a2f-4c2a-9b77-0c1f8c0a1b01",
     "legalEntityName": "Example Budget Pty Ltd",
     "accreditationNumber": "ADR-EXAMPLE-0001",
     "accreditationLevel": "UNRESTRICTED",
     "logoUri": "https://budget.example/logo.png",
     "status": "ACTIVE",
     "dataRecipientBrands": [{
       "dataRecipientBrandId": "1d5e3e7a-6c0b-4d43-9a8a-3f1a2c4b5d02",
       "brandName": "Example Budget",
       "logoUri": "https://budget.example/logo.png",
       "status": "ACTIVE",
       "softwareProducts": [{
         "softwareProductId": "2b7f9c1e-0d4a-4e6b-8f3c-5a6b7c8d9e03",
         "softwareProductName": "Example Budget Web",
         "softwareProductDescription": "Example Budget Web, a software product in development",
         "logoUri": "https://web.budget.example/logo.png",
         "status": "ACTIVE",
         "clientUri": "https://web.budget.example",
         "redirectUris": ["https://web.budget.example/callback"],
         "tosUri": "https://web.budget.example/terms",
         "policyUri": "https://web.budget.example/privacy",
         "jwksUri": "https://web.budget.example/jwks",
         "revocationUri": "https://web.budget.example/revoke",
         "recipientBaseUri": "https://cdr.web.budget.example",
         "scopes": ["openid", "profile", "bank:accounts.basic:read", "bank:transactions:read", "cdr:registration"]
       }]
     }]
   }]}
   ```

2. Add the records. The command adds each record that the Register does not have, and changes no record that exists ([Register: the participant load](../../services/register/design/operator-interface.md#reg-1531)):

   ```sh
   export REGISTER_OPERATOR_TOKEN=lab-operator-token
   ./cdrlab participants add -register-url http://localhost:8080 -file example-budget.json
   ```

3. Get the client certificate. `cdrlab` makes the key on your host and sends only the CSR. `-key-readable-by-all` lets the user of a container read the key:

   ```sh
   mkdir -p budget
   ./cdrlab cert issue -register-url http://localhost:8080 -kind client \
     -legal-entity-id 6f6c2a04-6a2f-4c2a-9b77-0c1f8c0a1b01 -software-product-id 2b7f9c1e-0d4a-4e6b-8f3c-5a6b7c8d9e03 \
     -cn "Example Budget Web" -o "Example Budget" -cert budget/client.crt -key budget/client.key -key-readable-by-all
   ```

Your software product is now in the Register APIs:

```sh
http --verify .lab/pki/public/root.crt https://localhost:8081/cdr-register/v1/all/data-recipients x-v:4 |
  jq '.data[] | select(.legalEntityId == "6f6c2a04-6a2f-4c2a-9b77-0c1f8c0a1b01") | .dataRecipientBrands[].softwareProducts'
```

To change or remove a record, use the operator console of the Register (`http://localhost:8080`), or make a new lab: a load does not change a record that exists.

## Run your software on the lab network

A browser on the host cannot reach the authorisation pages of a brand, because the base URIs of the compose lab are compose host names, for example `https://wattle-bank:8081`. Thus run your software and its tests as a container on the compose network `cdr-lab_default`, and drive the authorisation over HTTP from there.

The Register and the brand get your JWKS from the `jwksUri` over TLS, and they trust only the public root of the lab. Thus give the host names of your URIs a certificate from the mock public CA, and give them to the container as network aliases:

```sh
./cdrlab public-cert issue -dir .lab/pki -cert budget/web.crt -key budget/web.key -key-readable-by-all \
  web.budget.example cdr.web.budget.example
docker run --detach --name budget-web --network cdr-lab_default \
  --network-alias web.budget.example --network-alias cdr.web.budget.example \
  --volume "$PWD/budget:/budget:ro" --volume "$PWD/.lab/pki:/pki:ro" <your image>
docker run --rm --network cdr-lab_default --volume "$PWD/.lab/pki/public:/pki:ro" \
  docker.io/curlimages/curl:8.17.0 --silent --show-error --cacert /pki/root.crt https://web.budget.example/jwks
```

The last command checks that the lab can get your JWKS. In the container, your software uses these URLs and files:

| Item | Value |
|---|---|
| Register public APIs and JWKS (TLS) | `https://register:8081` |
| Register secure APIs and token endpoint (mTLS), for example Get SSA | `https://register:8083` |
| OpenID Provider configuration of Wattle Bank | `https://wattle-bank:8081/.well-known/openid-configuration` |
| Mail sink API | `http://mail:8025` |
| Client certificate and key | `/budget/client.crt`, `/budget/client.key` |
| Trust for the mTLS servers | `/pki/cdr/ca-bundle.pem` |
| Trust for the TLS servers | `/pki/public/root.crt` |

The [Register specification](../../services/register/design/README.md) and the [data holder specification](../../services/dataholder/design/README.md) give the APIs and their rules.

## Add test customers

A brand starts with no customers. Write a customer load file with invented data only ([data holder: the customer load file](../../services/dataholder/design/operator-interface.md#dh-2650)). An account names a product of the brand by its `productId`, and a transaction has an `age` in place of its times. For example, `customers.json`:

```json
{"customers": [{
  "firstName": "Jordan",
  "lastName": "Lee",
  "emailAddress": "jordan.lee@example.test",
  "accounts": [
    {"productId": "wattle-pocket", "accountNumber": "12345678",
      "transactions": [{"type": "PAYMENT", "status": "POSTED", "description": "Coffee", "amount": "-4.50", "age": "P1D"}]},
    {"productId": "canopy-rewards-card", "accountNumber": "87654321", "creditLimit": "500.00", "openStatus": "CLOSED"}
  ]
}]}
```

The product IDs of a brand are in its product API:

```sh
http --verify .lab/pki/public/root.crt https://localhost:8181/cds-au/v1/banking/products x-v:5 | jq '.data.products[].productId'
```

Add the customers. The command adds each customer whose email address the brand does not have, and gives only the counts:

```sh
export DATAHOLDER_OPERATOR_TOKEN=lab-operator-token
./cdrlab customers add -dataholder-url http://localhost:8180 -file customers.json
```

## Sign in as a customer

At the authorisation, the consumer gives the email address or the customer number of a customer. The brand then sends a one-time password (OTP) to that email address ([data holder: consumer authentication](../../services/dataholder/design/consumer-authentication.md#dh-1220)). All email of the lab goes to the mail sink. A test gets the OTP from the mail sink API, at `http://mail:8025` on the compose network or at `http://localhost:8025` on the host:

```sh
id=$(http :8025/api/v1/search query=='to:"jordan.lee@example.test"' | jq --raw-output '.messages[0].ID')
http ":8025/api/v1/message/$id" | jq --raw-output .Text | grep --only-matching --extended-regexp 'one-time password is [0-9]{6}'
```

The web page of the mail sink is at `http://localhost:8025`. An OTP is valid for 5 minutes and for one authorisation.

## A cluster deployment

In a cluster deployment, reach the operator interfaces with `kubectl port-forward`, and add the local host and port to their Host allowlists ([operator access](../../charts/cdr-lab/README.md#operator-access)). Then run the same `cdrlab` commands, with the token of the operator token Secret:

```sh
kubectl port-forward --namespace lab service/register 8080:8080 &
kubectl port-forward --namespace lab service/wattle-bank 8180:8080 &
export REGISTER_OPERATOR_TOKEN=$(kubectl get secret --namespace lab lab-operator-token -o jsonpath='{.data.token}' | base64 -d)
export DATAHOLDER_OPERATOR_TOKEN=$REGISTER_OPERATOR_TOKEN
```

With external host names and routes, a browser can do the authorisation ([routes](../../charts/cdr-lab/README.md#routes)). The host names of your software must resolve in the cluster, and its TLS certificate must chain to the root that the chart trusts (`publicTLS.rootSecret`).

### Keep the records in the values

A fresh deployment deletes all records that you added with `cdrlab`. To rebuild them, put the same records into the values of the release: `participants` takes the participant load file, and `brands.<brand>.customers` takes the customers of the customer load file ([seed data](../../charts/cdr-lab/README.md#seed-data)). The files convert to values with no change of format:

```sh
yq --output-format yaml '{"participants": .}' example-budget.json > budget-values.yaml
yq --output-format yaml '{"brands": {"wattle-bank": {"customers": .customers}}}' customers.json > customers-values.yaml
helm upgrade cdr-lab oci://ghcr.io/lukemundy/cdr-lab/charts/cdr-lab --version X.Y.Z --namespace lab \
  --values my-values.yaml --values budget-values.yaml --values customers-values.yaml --wait --timeout 10m
```

A list in a later values file replaces the full list of an earlier file. Thus give all the customers of a brand in one file. After each install and each upgrade, the seed Jobs add each record of the values that the service does not have. A customer gets a new customer number at each fresh deployment: find a seed customer by its email address.
