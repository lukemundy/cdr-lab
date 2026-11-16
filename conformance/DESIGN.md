# Conformance suite: design

This document gives the requirements for the conformance suite in `conformance/`. Each change to the suite must obey them. It does not describe the implementation.

## Purpose

The conformance suite checks that the deployed components of the lab obey the Consumer Data Right (CDR). When a test of the suite fails, a component does not conform.

## Sources

The rules follow from [ADR 0019][adr-0019]. "Judgement" marks a rule that has no other source, and gives the reason.

## Scope

In scope:

- The black-box tests of the deployed Register and data holder, one Go package for each component: `conformance/register` and `conformance/dataholder`. The Register serves the CRL of the mock CDR PKI ([ADR 0022][adr-0022]), so `conformance/register` has its tests.
- The pinned CDS OpenAPI specifications in the `spec/` directory of each package.
- The shared test code in `conformance/internal/labtest`.

Out of scope:

- The OpenID conformance run in `conformance/openid` (build tag `openid`, `mise run openid-conformance`). It runs the test plan of the OpenID Foundation, and it does not run in CI ([README][readme]).
- The interfaces that the CDR does not define: the operator interfaces, the health probes and the operator consoles. The Go tests of each service test them, and the browser tests test the UIs ([ADR 0019][adr-0019]).

## Subject of a test

- A test checks only behaviour that the CDR defines: the CDS, the CDR Rules, and the security profile and PKI that the CDS gives. ([ADR 0019][adr-0019]) The reason: a failure then always means that a component does not conform.
- A test can call an interface that the CDR does not define to set up a state, for example a status change on the Register or a new customer of the data holder. The test does not check that interface: its subject is a CDR behaviour. ([ADR 0019][adr-0019])
- A requirement of an interface that the CDR does not define gets a Go test in its service, which `mise run test` runs with no part of the lab deployment other than PostgreSQL. ([ADR 0019][adr-0019])

## Black-box tests

- A test calls the deployed components only through their network interfaces, as a CDR participant or a consumer does. It does not read a database, and it does not import the code of a service. Judgement: the suite then tests the deployment, and it can run against a different implementation of a component. The Go `internal` rule makes an import of service code impossible.
- A test acts as the participant does: it finds the base URIs of a brand through the Register, gets each SSA from the Register, and acts as the browser of the consumer in an authorisation flow ([the conformance suite acts as a participant](../services/dataholder/design/non-functional-requirements.md#dh-0015)).
- A test checks each response of a CDS API against the pinned CDS OpenAPI specification of the target CDS version, in addition to the checks of the test. Judgement: a field, a type or an enumeration that does not agree with the CDS then fails in each test that gets it ([ADR 0011][adr-0011]).
- The pinned specification files are copies of the CDS files. Do not edit them. The `SOURCE.md` file next to them gives their source and checksum. Judgement: an edit makes the check agree with the component, not with the CDS.

## Environment

- The suite runs in the `conformance` compose service on the lab network, with `mise run conformance`. It also serves the JWKS and the endpoints of the software products that it acts as. Judgement: the components then call the suite as they call a participant.
- A test gets each setting from an environment variable, which `compose.yaml` gives. A test fails, and does not skip, when a variable is not set. Judgement: a skip hides a test that did not run, so CI can pass with no check.
- The test files have the build tag `conformance`, so that `mise run test` does not run them. Judgement: `mise run test` needs only PostgreSQL ([README: Development][readme-development]).

## Runs and data

- The suite passes in one run after a new lab: `mise run reset && mise run up && mise run conformance`. A test can change the seed data in a way that cannot be reversed, so a second run with no reset can fail. That is acceptable. Judgement: a new lab is quick to make, and its data has no value ([AGENTS.md][agents]).
- No requirement, acceptance criterion or test needs a second run of the suite. Judgement: the suite is the slowest check in CI ([ADR 0019][adr-0019]), and a second run doubles that time.
- One test proves a full sequence, for example register, delete and register again. When a step of the sequence fails, the test stops, because the next steps depend on it. Judgement: the alternative is a sequence across tests or across runs, which `go test` does not keep in order.
- Each test package has its own participants in the seed files and its own listen ports, so that `go test` runs the packages at the same time. Do not make the packages run one after the other (`-p 1`). Judgement: the packages are independent, and a sequential run makes the suite slower.
- A test that changes a state uses data that no other test uses, for example its own participant, customer or account number. Judgement: the tests of a package can then run in any sequence.
- All customer, account and transaction data in the tests is invented. ([README: Data][readme-data]) The CDR data and personal information rules of [ADR 0006][adr-0006] apply also to mock data.

## Time

- The suite has no long waits. A fixed wait is at most 2 seconds. A poll stops as soon as its condition is true, and its limit is the time that the requirement or the lab configuration gives, for example the status interval of the data holder in `compose.yaml`. Judgement: a long wait makes each run of the suite slower.
- A time rule with a longer period, for example the life of an authorisation code or the 10 minutes of an authorisation flow, is proved in-process by a Go test that controls the time, in the service or in its library in `pkg/`. The suite does not check it. Judgement: as above.

## Exceptions

| Test | Exception | Reason |
|---|---|---|
| The data holder tests of the end of eligibility, the removal of a software product and the notices (`eligibility_test.go`, `removal_test.go`, `notice_test.go`) | They read the arrangements of a customer from the operator interface to see the end reason of an arrangement and the outcome of a notice. | The subject is a CDR behaviour, but no CDR API shows these values to the test. In production, the consumer dashboard of the data holder shows them. |
| The data holder test of the logo (`brand_test.go`) | It checks the `Content-Security-Policy` and `X-Content-Type-Options` headers of the logo, which the CDR does not define. | The CDS makes the `logoUri` of a brand mandatory ([CDS: Register API schemas][cds-register-schemas]), and a data recipient opens that URL. Only the URL that the Register gives proves that the deployed public listener serves the logo with the headers that stop a script in it ([data holder specification][dataholder-design]). |
| The data holder test of the stylesheets and the font of the authentication page (`brand_test.go`) | It checks that the stylesheets of the page and one font that they refer to load from the origin of the page. The CDR does not define the styles of the page. | The data holder pages work with no styles, so only a request to the deployed public listener shows that the image has the build of the styles and the fonts ([data holder specification][dataholder-design]). |
| The Register tests of Get SSA (`ssa_test.go`) | A response with the status 200 is not checked against the specification. | The response body is a JWT with no JSON quotes, which the string schema of the specification cannot validate ([the body of the SSA response][register-ssa-body]). |
| The Register test of an unknown path (`status_test.go`) | The response is not checked against the specification. | The specification has no such path. |

[adr-0006]: ../docs/adr/0006-telemetry.md
[cds-register-schemas]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas
[adr-0011]: ../docs/adr/0011-domain-data-as-typed-go.md
[adr-0019]: ../docs/adr/0019-limit-the-conformance-suite-to-cdr-conformance.md
[adr-0022]: ../docs/adr/0022-run-the-mock-cdr-intermediate-ca-in-the-register.md
[readme]: ../README.md
[readme-development]: ../README.md#development
[agents]: ../AGENTS.md
[readme-data]: ../README.md#data
[register-ssa-body]: ../services/register/design/software-statement-assertion.md#reg-0414
[dataholder-design]: ../services/dataholder/design/README.md
