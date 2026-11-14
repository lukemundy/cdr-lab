# 0011. Model domain data as typed Go, with its validation in the domain

- Date: 2026-10-18

## Context

The services keep and serve data that the CDS defines, for example products, accounts and transactions. The CDS gives each type in an OpenAPI specification, with its fields, its enumerations and its required fields. The conformance tests already check each response against the pinned CDS OpenAPI specification.

The `domain` package contains the entities and the business rules, and imports only the standard library ([ADR 0005](0005-service-architecture.md)).

There are two real options for data that the CDS defines:

| | Validate against the OpenAPI schema, and keep the JSON as given | Typed Go in `domain`, with the validation there |
|---|---|---|
| Code | Little code for each type. The schema gives the validation. | A Go type and a validation for each CDS type and each enumeration that the service uses. |
| Dependencies | An OpenAPI library in the service. | None |
| Business rules | The domain still needs Go types for each field that a rule uses, so part of each type is in two places. | One type, which the rules use. |
| Calculations and changes to one record | The service must parse the JSON each time. | Direct, on the type. |
| What the service serves | The JSON that it got, also fields that no rule checks. | Only fields that the type has. |

## Decision

- Domain data is Go types in the `domain` package of each service. The validation of the data is in `domain`, with the other business rules.
- A service does not validate domain data with an OpenAPI or JSON schema library. The check against the CDS OpenAPI specification is in the conformance tests.
- A service does not store or serve JSON that it does not parse into a domain type.
- A `jsonb` column can hold a typed domain struct. When a query must read a value in that column, the value moves to its own columns.
- To limit the code, a type has only the optional CDS fields and components that the service uses now. For each enumeration that a type uses, the type has the full CDS list of values.

## Consequences

- Each CDS type that a service uses needs Go code and tests. This is more code than a schema check.
- A new CDS version can need a change to the Go types, in addition to the change to the pinned specification.
- A field that the CDS defines but that the service does not have yet is not in the responses, until a change adds it.
- A rule, a calculation or an operator change to one record uses the Go type, and needs no JSON parser.
- The domain tests check the rules, not the agreement with the CDS. A type that does not agree with the CDS fails only in the conformance tests.
