import { fieldErrorsType } from "@cdr-lab/console/problem";
import { expectTypeOf, test } from "vitest";

import { sessionEndedType } from "./client";
import type { components } from "./schema.gen";

// The type check fails when the OpenAPI document and the shared constant have different values.
test("the field-errors type of the OpenAPI document is the shared field-errors type", () => {
  expectTypeOf<components["schemas"]["FieldErrorsProblem"]["type"]>().toEqualTypeOf<
    typeof fieldErrorsType
  >();
});

test("the session-ended type of the OpenAPI document is the session-ended type of the client", () => {
  expectTypeOf<components["schemas"]["SessionEndedProblem"]["type"]>().toEqualTypeOf<
    typeof sessionEndedType
  >();
});
