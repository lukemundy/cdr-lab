# 0041. Give the seed customers of a brand under that brand, with accounts that name a product of the brand, and transaction times relative to the load

- Date: 2026-11-23

## Context

The values of the umbrella chart give the seed data of a deployment, and no value of a brand can disagree with its Register record ([ADR 0033](0033-deploy-the-lab-to-a-cluster-with-one-umbrella-chart-that-derives-each-brand-from-its-register-record.md)). [ADR 0040](0040-add-participants-and-customers-with-add-only-loads-of-the-operator-interfaces-from-cdrlab-and-from-the-umbrella-values.md) loads the customers of a brand with an add-only load of the data holder operator interface. That load needs a file format for customers, accounts and transactions.

These facts constrain the format:

- A create of an account in the operator interface takes a free `productName`, `productCategory` and `bsb` ([data holder: create an account](../../services/dataholder/design/operator-interface.md#dh-0722)). Thus an account can name a product that the brand does not offer. The product catalogue comes from the brand, and each product has a `productId`. The brand file gives the BSB of the accounts that a consumer opens ([data holder: the BSB of the brand](../../services/dataholder/design/brand.md#dh-0334)).
- The data holder makes the customer number ([data holder: the customer number](../../services/dataholder/design/consumer-authentication.md#dh-1201)). The email address is also a user identifier, and it is unique in the brand ([data holder: user identifiers](../../services/dataholder/design/consumer-authentication.md#dh-1200)).
- When a request of Get Transactions For Account has no `oldest-time`, the data holder gives the transactions of the last 90 days ([CDS: Get Transactions For Account](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-banking-api_get-transactions-for-account)). A deployment can stay up for months.
- Helm merges maps from two values files, but a list in a later values file replaces the full list ([Helm: values files](https://helm.sh/docs/chart_template_guide/values_files/)).

The real options for the identity of a customer were:

1. **The email address.** The data holder makes the customer number, as for each other customer. A fresh deployment gives a new customer number.
2. **A customer number in the file.** The number stays the same after a fresh deployment. The file then has one more unique value to keep, and the load needs one more conflict rule.

The real options for the times of a transaction were:

1. **Absolute times**, as the create of a transaction takes them. The data is the same in each deployment, but after 90 days the transactions are not in the default response of Get Transactions For Account.
2. **Times relative to the load.** Each transaction has an age. The transactions stay recent in each new deployment, but a test cannot expect a fixed time.

## Decision

The customers of a brand are a list under that brand in the values of the umbrella chart, in the same format as the customer load file. A customer cannot go to a brand that the release does not run, because it has no place in the values outside a brand.

- **Customer.** A customer has the fields of a create of a customer, and a list of accounts. The email address identifies a seed customer (option 1). A consumer and a test sign in with the email address, so a customer number of the file adds a rule that no test needs.
- **Account.** An account names a product of the brand by its `productId`. The data holder takes the product name and the product category from that product, and the BSB from the brand. Thus an account cannot disagree with the brand. The account number is in the file, because it is unique with the BSB and a test finds an account by it. An account can be closed, with the CDS value `CLOSED` of `openStatus`.
- **Transaction.** A transaction has the fields of a create of a transaction, except the three times. It has an age instead, and the load sets the times from the time of the load less the age (option 2). The data then stays in the default window of the banking APIs in each new deployment.
- **The same customers.** Two loads of the same file give the same customers by email address, with the same fields, the same accounts by BSB and account number, and the same transactions in the same sequence and with the same ages. The customer numbers, the account IDs, the transaction IDs and the absolute times can differ.

This decision applies to the customer load file of the data holder and to the customer values of the umbrella chart. It does not change the create operations of the operator interface, which keep their fields.

## Consequences

- No value of a seed customer can name a product or a BSB that its brand does not have. A `productId` that the brand does not have makes the load fail, with an error that names the account.
- A list of customers in a later values file replaces the full list of that brand. A deployment that adds one customer in a second values file must give all customers of the brand there.
- Tests and guides refer to a seed customer by its email address, and to an account by its account number. They cannot use a customer number from an earlier deployment.
- A test that needs exact transaction times creates its transactions with the create operation, not with the seed.
- The format is a lab format. It is not a CDS schema, and a change of it before version 1.0.0 needs no migration.
