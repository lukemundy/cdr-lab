import { fieldErrorsType } from "@cdr-lab/web-platform/problem";
import { expectTypeOf, test } from "vitest";

import type { components } from "./schema.gen";

// The type check fails when the OpenAPI document and the shared constant have different values.
test("the field-errors type of the OpenAPI document is the shared field-errors type", () => {
  expectTypeOf<components["schemas"]["FieldErrorsProblem"]["type"]>().toEqualTypeOf<
    typeof fieldErrorsType
  >();
});
