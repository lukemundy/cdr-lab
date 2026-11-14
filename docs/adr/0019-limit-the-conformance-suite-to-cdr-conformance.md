# 0019. Limit the conformance suite to CDR conformance, and test the other interfaces in their service

- Date: 2026-10-23

## Context

The conformance suite in `conformance/` is a set of black-box tests. It calls the deployed components in the lab deployment through their APIs, and checks each response against the pinned CDS OpenAPI specifications. It needs the full lab deployment, so it is the slowest check in CI, and it can pass only one time after the lab loads the seed files.

The services also have interfaces that the CDR does not define: the operator interfaces, the health probes, and the operator consoles with their console APIs ([ADR 0014](0014-web-uis-as-embedded-react-single-page-applications.md)). The suite tests some of these interfaces now, for example the operator interfaces of the Register and the data holder. [ADR 0017](0017-console-api-contracts-in-openapi.md) gives the suite one more task: to validate the responses of each console API against its OpenAPI document.

When the suite also tests these interfaces, these problems occur:

- A failure in the suite does not tell if a component is not conformant to the CDR, or if a lab-only interface has a defect.
- Each change to a lab-only interface must run the full lab deployment in CI to be tested.
- A mismatch between a Go adapter and its document shows only in the conformance run, not in `mise run test` ([ADR 0017](0017-console-api-contracts-in-openapi.md), Consequences).

The other option is to test these interfaces with Go tests in the service. The tests run the HTTP handlers of the service in the process, with the real adapters and, where necessary, the lab PostgreSQL that `mise run test` already uses. These tests do not prove the configuration of the deployment, for example a port or a setting in `compose.yaml`. The browser tests of a UI run against the lab deployment ([ADR 0015](0015-web-toolchain-and-npm-supply-chain-controls.md)), so they prove the deployment of the UIs.

This ADR supersedes the part of ADR 0017 that gives the console API check to the conformance suite. The other parts of ADR 0017 do not change.

## Decision

- The conformance suite tests only the behaviour that the CDR defines: the CDS, the CDR Rules, and the security profile and PKI that the CDS gives. The reason: a failure in the suite then always means that a component is not conformant.
- The suite can call an interface that the CDR does not define to set up a state for a test, for example a status change through an operator interface. It does not test that interface.
- Tests in the service test each interface that the CDR does not define. They run the HTTP handlers in the process, with `mise run test`, and need no other part of the lab deployment.
- For a console API, these tests validate each response against the OpenAPI document of the API, with the same validation code that the suite uses for the CDS specifications. The code is thus in `pkg/`, not in `conformance/`. The listener headers and the refusal of cross-origin requests of a console are also tested there.
- The browser tests of a UI check the main user journeys against the lab deployment ([ADR 0015](0015-web-toolchain-and-npm-supply-chain-controls.md)). They are not part of the conformance suite.

## Consequences

- A change to a lab-only interface does not need the conformance job in CI. CI can run the conformance job only for changes to the components and to the suite.
- A Go adapter that does not agree with its OpenAPI document fails in `mise run test`, without the lab deployment.
- The tests in the service do not prove the configuration of the deployment. A console that the deployment does not turn on, or a wrong port, shows only in the browser tests against the lab.
- The suite already has tests of the operator interfaces and the health probes. Until a change moves them, the suite does not obey this decision.
- The validation code moves from `conformance/` to `pkg/`, so the suite and the services use one copy.
