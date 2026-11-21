import { describe, expect, test } from "vitest";

import { fieldErrorsType, hasFieldErrors, ProblemError, problemOf, refusalOf } from "./problem";

const fieldErrorsProblem = {
  type: fieldErrorsType,
  title: "Values in the request are not valid",
  status: 400,
  detail: "One or more values in the request are not valid.",
  errors: [{ pointer: "#/name", detail: "name is required" }],
};

// Verifies: WCON-0013
test("the field-errors type is the identifier of its rule in the problem specification", () => {
  expect(fieldErrorsType).toBe(
    "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
  );
});

describe("ProblemError", () => {
  // Verifies: WCON-0013
  test("gives the field errors of the field-errors type", () => {
    const error = new ProblemError(fieldErrorsProblem, 400);

    expect(error.status).toBe(400);
    expect(error.message).toBe("Values in the request are not valid");
    expect(error.fieldErrors).toEqual(fieldErrorsProblem.errors);
  });

  // Verifies: WCON-0013
  test("gives no field errors for a different problem type", () => {
    const otherProblem = { ...fieldErrorsProblem, type: "https://other.example/" };
    const error = new ProblemError(otherProblem, 400);

    expect(error.fieldErrors).toEqual([]);
  });
});

// Verifies: WCON-0013
test("hasFieldErrors checks the problem type", () => {
  expect(hasFieldErrors(fieldErrorsProblem)).toBe(true);
  expect(hasFieldErrors({ title: "Conflict" })).toBe(false);
});

// Verifies: WCON-0013
test("refusalOf gives the problem and the field errors to the form", () => {
  const error = new ProblemError(fieldErrorsProblem, 400);

  expect(refusalOf(error)).toEqual({
    problem: fieldErrorsProblem,
    fieldErrors: fieldErrorsProblem.errors,
  });
  const failure = new TypeError("Failed to fetch");
  expect(() => refusalOf(failure)).toThrow(failure);
});

// Verifies: WCON-0012
test("problemOf names the service when the console gets no problem", () => {
  const conflict = { title: "Conflict", detail: "The account is closed." };

  expect(problemOf(new ProblemError(conflict, 409), "the data holder")).toBe(conflict);
  expect(problemOf(new TypeError("Failed to fetch"), "the data holder")).toEqual({
    title: "The console cannot get data from the data holder",
    detail: "Make sure that the data holder runs, and then try again.",
  });
});
