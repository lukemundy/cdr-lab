# 0033. Deploy the lab to a cluster with one umbrella chart that derives each brand from its Register record

- Date: 2026-11-23

## Context

Teams that build CDR software want a mock Register and mock data holders that stay up in their development clusters. The repository has a Helm chart for the Register and a Helm chart for one data holder brand. To deploy the lab with them, a person must do these steps in sequence: make the CAs on a host, install the Register with a participant seed, issue a certificate for each brand, make the Secrets, and then install each brand.

The values of a brand release must agree with the record of the brand in the Register seed: the brand ID and the base URIs. When they do not agree, the pod becomes ready, but data recipients cannot find the brand ([data holder: example values for each preset brand](../../services/dataholder/design/deployment.md#dh-1382)). The two charts get these values from two different files, so nothing stops a difference.

The lab and CI run on Docker Compose. The services, the seed files and the certificates use the compose service names, for example `register` and `wattle-bank`.

The real options were:

1. **A guide for the two charts.** It needs no new artifact. Each team does the steps by hand, and a team can still give a brand values that disagree with its Register record.
2. **A script or a CLI that installs the two charts in sequence.** It automates the steps, but a GitOps tool, for example Argo CD or Flux, cannot use it, and the two sets of values stay separate.
3. **One umbrella chart.** One release deploys the Register and the brands. One set of values gives the participants and the brands, and the chart makes the Register seed and the brand settings from the same values. A GitOps tool deploys it as one application.

## Decision

We deploy the lab to a Kubernetes cluster with one umbrella chart (option 3). A Kubernetes cluster is a supported environment of the lab, together with compose and CI. The deployment in a cluster is a lab, not a production service.

- The values of the chart give the seed data: the participants and the brands. The chart gives each brand of the release its Register record and its data holder settings from the same entry of the values. The values give the deployment settings of a brand, for example its host names, but no field that its Register record also has. Thus a brand cannot have values that disagree with its Register record.
- The chart runs the four preset brands. The chart has the Register record of each preset brand. A brand from a mounted brand directory needs the data holder chart, with its own record in the seed.
- One release runs in one namespace, and all its parts are in that namespace. One install Job then writes the CA Secrets that all parts mount, with no RBAC across namespaces.
- The Services of the Register and of each brand have the names of the compose services, for example `register` and `wattle-bank`. Thus the seed files, the certificates and the conformance suite use the same names in compose and in a cluster. One namespace can have only one release.
- PostgreSQL is outside the chart, as for the charts of the services ([ADR 0020](0020-service-migrations-create-roles-and-the-operator-creates-users.md)). The team gives the databases and the users. The guide gives a CloudNativePG example. We do not use a PostgreSQL subchart: since 2025-08-28, the versioned Bitnami charts and images get no free updates ([bitnami/containers#83267](https://github.com/bitnami/containers/issues/83267)), and a subchart goes against the database rules of ADR 0020.
- The public TLS certificates come from cert-manager, when the values give an issuer. Otherwise, the team gives existing Secrets. The services do not know the source of a certificate. The CDR certificates come from the Register ([ADR 0032](0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md)).
- The charts of the services stay. A team that needs a different set of parts, for example a brand from a brand directory, uses them.

## Consequences

- One `helm install` with one values file deploys the Register and the brands. A GitOps tool deploys the lab as one application.
- A team cannot run two umbrella releases in one namespace. A second lab needs a second namespace.
- The umbrella chart has a copy of the Register record of each preset brand. A test must keep the copy the same as the example seed of the Register.
- The umbrella chart and the service charts can differ in their values for the same setting. The chart tests must cover the settings that the umbrella chart passes to each part.
- A team must make the databases, the database users and the operator token Secret before an install. The chart cannot do it, because PostgreSQL is outside the chart.
- The rules for an environment apply to three environments. A change that affects the deployment must work in compose, in a cluster and in CI.
