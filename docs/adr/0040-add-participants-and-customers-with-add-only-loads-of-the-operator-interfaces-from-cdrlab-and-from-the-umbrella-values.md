# 0040. Add participants and customers with add-only loads of the operator interfaces, from cdrlab and from the umbrella values

- Date: 2026-11-23

## Context

A developer of data recipient software must add a software product and test customers to a running lab, in the compose lab or in a cluster deployment. A fresh deployment with the same umbrella values must give the same participants and customers again. Before version 1.0.0, a release makes no promise about the data of a deployment ([ADR 0027](0027-release-all-artifacts-with-one-version-from-a-manual-workflow-on-main.md)). Thus a fresh deployment rebuilds the seed data from the values, and the lab migrates no data.

These facts constrain the design:

- `cdrlab` uses the network interfaces of the services, as a lab operator does ([ADR 0035](0035-give-the-lab-one-cli-cdrlab-in-cli-with-an-image-that-follows-the-rules-of-a-service-image.md)). It has no access to the databases.
- The participant load of the Register is a command of the Register binary, with access to its database. It loads a file only into a Register that has no participants ([Register: no load into a Register with participants](../../services/register/design/deployment.md#reg-0901)). The umbrella chart runs it before the Register starts. After the first install, a participant that the values add has no effect. For example, a brand that an upgrade adds gets no Register record.
- The Register operator interface creates and adds one record in each request, and it has no read of a record. A 409 response can mean an identifier that is in use or a broken status rule ([Register: 409 for a conflict](../../services/register/design/operator-interface.md#reg-0856)).
- The data holder operator interface creates one customer, account or transaction in each request. The data holder makes the customer number ([data holder: the customer number](../../services/dataholder/design/consumer-authentication.md#dh-1201)). The interface has no find of a customer by its email address.
- A Kubernetes ConfigMap holds 1 MiB at most ([Kubernetes: ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)).

The real options were:

1. **A load command in each service binary**, with access to the database, as the participant load and the product load are. The chart runs it before the service starts. A load is all or nothing. `cdrlab` cannot use it, so a developer who adds records to a running lab needs a second path.
2. **`cdrlab` sends the existing requests, one for each record.** The services get no new operation. `cdrlab` must find the records that exist, but the Register has no read, and a find of a customer by email puts personal information in a URL. A failure after some requests leaves a part of the file in the service.
3. **One add-only load operation in each operator interface.** It takes a full file, adds each record that does not exist, and changes no record that exists. It is all or nothing. `cdrlab` sends the file, and the Jobs of the umbrella chart run `cdrlab` with the files of the values.

## Decision

We select option 3. The Register operator interface gets an add-only load of participants, and the data holder operator interface gets an add-only load of customers. `cdrlab` has one command for each. The commands and the Jobs use the same operations, so a record that a developer adds with `cdrlab` also goes into the values with no change.

- **Add-only.** A load adds each record of the file that the service does not have. It changes no record that exists, also when its fields differ, and it removes no record. A participant record exists when the Register has a record of the same kind with its identifier. A customer exists when the brand has a customer with its email address. A full reconcile needs reads, changes and removals that the services do not have, and a fresh deployment gives a changed record with no migration.
- **All or nothing.** The service checks the full result of the load. When one record is not valid, the load changes nothing, and the error names the record.
- **Results with no personal information.** The response gives each result in the sequence of the file. `cdrlab` writes only counts and participant identifiers to its output, because the output of a Job is a log.
- **The umbrella chart.** After each install and each upgrade, a Job runs the participant load with the participant seed file of the release, and a Job for each brand runs the customer load with the customers of that brand. A Job tries again until its service is ready. The participant seed load before the first start of the Register stays.
- **Onboarding of a software product.** A developer adds a data recipient, with its brand and software product, in the format of the load file. Then the developer gets a client certificate through the Register issue path ([ADR 0032](0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md)), which also registers the certificate. `cdrlab` has no command that does the two steps together. A data holder developer adds a data holder brand and gets a server certificate in the same way.

This decision applies to the participants of the Register and the customers of a data holder brand. It does not apply to the product catalogue, which comes from the brand. It does not change the status actions or the per-record operations of the operator interfaces.

## Consequences

- One file format for each kind of record works with `cdrlab` and in the values.
- An upgrade that adds a brand, a participant or a customer to the values adds it. An upgrade that changes or removes a record of the values changes nothing in the deployment. The change needs a fresh deployment, and the chart README must say so.
- A second run of a load with the same file changes nothing. Thus each Job can run again at each upgrade.
- The two operator interfaces get one more lab-only operation each. Real components do not use them.
- A file is one request. Thus the size of the seed of one service is limited by the body limit of its load and by the ConfigMap limit.
- The Jobs and the brands start in no fixed sequence, and the Jobs retry until each service is ready. A defect in a seed file shows as a failed Job, not as a failed install.
- A customer gets a new customer number on each fresh deployment. Tests and developers find a seed customer by its email address.
