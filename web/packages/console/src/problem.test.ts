import { describe, expect, test } from "vitest";

import { hasFieldErrors, ProblemError, problemOf, refusalOf } from "./problem";

const fieldErrorsType = "https://service.example/DESIGN.md#console-api";
const fieldErrorsProblem = {
  type: fieldErrorsType,
  title: "Values in the request are not valid",
  status: 400,
  detail: "One or more values in the request are not valid.",
  errors: [{ pointer: "#/name", detail: "name is required" }],
};

describe("ProblemError", () => {
  test("gives the field errors of the problem type of the service", () => {
    const error = new ProblemError(fieldErrorsProblem, 400, fieldErrorsType);

    expect(error.status).toBe(400);
    expect(error.message).toBe("Values in the request are not valid");
    expect(error.fieldErrors).toEqual(fieldErrorsProblem.errors);
  });

  test("gives no field errors for a different problem type", () => {
    const error = new ProblemError(fieldErrorsProblem, 400, "https://other.example/");

    expect(error.fieldErrors).toEqual([]);
  });
});

test("hasFieldErrors checks the problem type", () => {
  expect(hasFieldErrors(fieldErrorsProblem, fieldErrorsType)).toBe(true);
  expect(hasFieldErrors({ title: "Conflict" }, fieldErrorsType)).toBe(false);
});

test("refusalOf gives the problem and the field errors to the form", () => {
  const error = new ProblemError(fieldErrorsProblem, 400, fieldErrorsType);

  expect(refusalOf(error)).toEqual({
    problem: fieldErrorsProblem,
    fieldErrors: fieldErrorsProblem.errors,
  });
  const failure = new TypeError("Failed to fetch");
  expect(() => refusalOf(failure)).toThrow(failure);
});

test("problemOf names the service when the console gets no problem", () => {
  const conflict = { title: "Conflict", detail: "The account is closed." };

  expect(problemOf(new ProblemError(conflict, 409, fieldErrorsType), "the data holder")).toBe(
    conflict,
  );
  expect(problemOf(new TypeError("Failed to fetch"), "the data holder")).toEqual({
    title: "The console cannot get data from the data holder",
    detail: "Make sure that the data holder runs, and then try again.",
  });
});
