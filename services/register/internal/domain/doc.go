// Package domain contains the entities, value objects and business rules of
// the mock CDR Register.
//
// Rules:
//   - Import only the standard library. No I/O packages (os, net, database/sql,
//     log). The linter enforces this.
//   - Do not read the clock or make IDs. The app layer gives these values.
//   - Do not accept a context.Context. Domain operations do no I/O.
package domain
