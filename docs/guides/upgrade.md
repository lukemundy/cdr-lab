# Upgrade the lab

This guide gives the upgrade of a compose lab and of a cluster deployment to a new release.

Before version 1.0.0, a release makes no promise about the data of a deployment. A new version can need a fresh deployment: new databases and new participants ([ADR 0028](../adr/0028-find-the-need-for-a-fresh-deployment-from-the-migrations-and-the-event-golden-files.md)). The lab data has no value, so a fresh deployment is quick and safe.

## Read the release notes

The release notes of each version are on its GitHub Release. Read them before each upgrade:

1. **Needs a fresh deployment** is the first section. It says "No" when the data of the earlier version stays usable. It says "Yes" and lists the changed files when a migration or a stored event of the earlier version changed ([release: the fresh deployment section](../../tools/release/DESIGN.md#rls-0064)). The section compares with the version just before. When you skip a version, read the section of each version between.
2. **Changelog** gives each change. A change with `!` is a breaking change, for example of an API or a value of the chart.
3. **Images** and **Charts** give the digests that you verify before a cluster upgrade ([verify a release](../../charts/cdr-lab/README.md#verify-a-release)).

## The compose lab

`cdrlab up` starts only a lab of its own version. A binary of a different version refuses the lab directory, and tells you to reset it ([cdrlab: the release mode changes only a lab of its own version](../../cli/cdrlab/design/compose-lab.md#cli-0145)). Thus each version change of the compose lab is a fresh deployment. In the lab directory:

```sh
./cdrlab reset
```

Then get the binary of the new version ([install guide](install.md#the-compose-lab-from-a-release)), and start the lab again:

```sh
./cdrlab up
```

`cdrlab reset` deletes the data, the CAs and the files that `cdrlab up` wrote. Run the onboarding commands again after the start, for example the commands of the [ADR developers guide](adr-developers.md). Keep your load files to make this quick.

From a checkout, `mise run reset` and then `mise run up` do the same.

## A cluster deployment

When the release notes say "No", upgrade the release in place, after you verify the chart:

```sh
helm upgrade cdr-lab oci://ghcr.io/lukemundy/cdr-lab/charts/cdr-lab --version X.Y.Z \
  --namespace lab --values my-values.yaml --wait --timeout 10m
```

The upgrade runs the Jobs of the chart again ([upgrade](../../charts/cdr-lab/README.md#upgrade)). The seed Jobs add only the new seed records of the values, for example the participants and the customers that you added to the values. They change no record that exists, and they remove no record ([umbrella chart: only the new records after an upgrade](../../charts/cdr-lab/DESIGN.md#lab-0135)). To change or remove a seed record, make a fresh deployment.

When the release notes say "Yes", make a fresh deployment with the steps of the chart README ([fresh deployment](../../charts/cdr-lab/README.md#fresh-deployment)):

1. Uninstall the release.
2. Delete the Secrets that the Jobs made, so that the new install makes a new mock CDR CA.
3. Make new, empty databases for the Register and each brand, with the same users.
4. Install the new version with the same values.

The seed Jobs then add the participants and the customers of the values again. The customer numbers and the account IDs change: find a seed customer by its email address, and an account by its account number. The records that you added with `cdrlab` and not in the values are gone. To keep them, put them in the values ([ADR developers guide](adr-developers.md#keep-the-records-in-the-values)).
