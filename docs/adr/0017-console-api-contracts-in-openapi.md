# 0017. Define each console API in a hand-written OpenAPI document, and generate the TypeScript types from it

- Date: 2026-10-22
- Superseded in part by [ADR-0019](0019-limit-the-conformance-suite-to-cdr-conformance.md)

## Context

Each operator console calls a console API in its service ([ADR 0014](0014-web-uis-as-embedded-react-single-page-applications.md)). The API is in Go and the UI is in TypeScript. If the two do not agree on a field, a route or an error, the UI fails only at runtime.

The conformance suite already validates CDS responses against the pinned CDS OpenAPI specifications with `getkin/kin-openapi`. These specifications are OpenAPI 3.0.3. A service does not validate domain data with a schema library. The check against a specification is in the conformance tests ([ADR 0011](0011-domain-data-as-typed-go.md)). The operator interfaces return errors as RFC 9457 problem details.

We compared four options:

| | Hand-written OpenAPI | OpenAPI generated from Go | TypeScript types written by hand | Consumer-driven contracts (Pact) |
|---|---|---|---|---|
| Source of truth | One YAML document for each API | The Go handlers | None | The contracts that the UI tests write |
| Drift found by | The conformance suite (Go side) and the TypeScript compiler (UI side) | The TypeScript compiler | End-to-end tests only | The provider verification |
| Cost | A document to keep correct by hand | A code-first HTTP framework, for example huma ([huma](https://github.com/danielgtaylor/huma)), which changes the HTTP adapter pattern of each service | Least tooling | A broker or a file exchange, and a Go verifier ([Pact JS](https://github.com/pact-foundation/pact-js)) |

TypeScript generators for OpenAPI:

- `openapi-typescript` generates types only, and `openapi-fetch` is a small typed client that uses them ([openapi-ts](https://openapi-ts.dev/introduction)). `openapi-msw` gives MSW test handlers the same types ([openapi-msw](https://github.com/christoph-fricke/openapi-msw)). `openapi-typescript` 7.13.0 fails with TypeScript 7 ([openapi-typescript#2841](https://github.com/openapi-ts/openapi-typescript/issues/2841)).
- orval generates clients, query hooks and mocks, and works with TypeScript 7. It has 25 direct dependencies, and it depends on esbuild, which has an install script ([orval](https://github.com/orval-labs/orval)).

## Decision

This decision applies to the console APIs. It does not apply to the CDS APIs, which have the CDS specifications, or to the operator interfaces.

- Each console API has one OpenAPI 3.0 document, written by hand, in the directory of the service that serves it. The service does not read the document at runtime.
- The conformance suite validates the responses of each console API against its document, with the same code that it uses for the CDS specifications.
- Each object schema in a console API document forbids properties that it does not name (`additionalProperties: false`). Thus the validation finds a field that the Go adapter adds and the document does not have. The CDS specifications permit more properties, but a console API has one client, which ships with it.
- `openapi-typescript` generates the TypeScript types of each document into the UI that uses it. The generated file is committed. CI generates it again and fails when the result is different. Thus the generator does not run at each install, and a change to the document shows in the diff of the UI.
- The UI calls the API with `openapi-fetch`, and its tests use `openapi-msw` handlers. The reason for these tools: they generate types only, with few dependencies, and the client, the test handlers and the document use the same types.
- Errors from a console API are RFC 9457 problem details, as the operator interfaces give.

## Consequences

- A change to a console API changes three things in one pull request: the document, the Go adapter and the UI. The conformance suite finds a Go adapter that does not agree with the document. The TypeScript compiler finds a UI that does not agree with it.
- A person keeps the document correct by hand. A new field needs a change to the document before the conformance suite accepts it.
- The conformance check needs the lab deployment, so a mismatch on the Go side shows in the conformance run, not in the unit tests.
- The generator uses the TypeScript 6 alias until it supports TypeScript 7 ([ADR 0015](0015-web-toolchain-and-npm-supply-chain-controls.md)).
- `openapi-msw` gives the test handlers types, but it does not validate their bodies at runtime. A handler with a body that does not agree with the document can still compile if its TypeScript type is wider than the schema.
