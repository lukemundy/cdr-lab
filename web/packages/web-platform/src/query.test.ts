import { expect, test } from "vitest";

import { ProblemError } from "./problem";
import { createQueryClient } from "./query";

const retry = createQueryClient().getDefaultOptions().queries?.retry;

function retries(failures: number, error: Error): boolean {
  if (typeof retry !== "function") throw new Error("retry is not a function");
  return retry(failures, error);
}

// Verifies: WCON-0010
test("a server or network failure is tried again two times", () => {
  const serverError = new ProblemError({ title: "Internal Server Error" }, 500);

  expect(retries(0, serverError)).toBe(true);
  expect(retries(1, new TypeError("Failed to fetch"))).toBe(true);
  expect(retries(2, serverError)).toBe(false);
});

// Verifies: WCON-0011
test("a request that the API refused is not tried again", () => {
  expect(retries(0, new ProblemError({ title: "Not Found" }, 404))).toBe(false);
});
