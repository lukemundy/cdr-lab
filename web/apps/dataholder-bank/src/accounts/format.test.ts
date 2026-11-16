import { expect, test } from "vitest";

import { effectiveTime, formatDay, formatSignedAmount } from "./format";

// Verifies: DH-2206
test("an amount of a transaction shows its sign", () => {
  expect(formatSignedAmount("-4.50")).toBe("-$4.50");
  expect(formatSignedAmount("1200.00")).toBe("+$1,200.00");
  expect(formatSignedAmount("0.00")).toBe("$0.00");
  expect(formatSignedAmount("12345678901234567.89")).toBe("+$12,345,678,901,234,567.89");
  expect(formatSignedAmount("not an amount")).toBe("not an amount");
});

test("a date shows as the same day in each time zone", () => {
  expect(formatDay("2026-10-01")).toBe("1 October 2026");
  expect(formatDay("October")).toBe("October");
});

test("the effective time is the posting time, else the execution time", () => {
  const base = {
    accountId: "a",
    transactionId: "t",
    isDetailAvailable: false,
    type: "PAYMENT",
    status: "POSTED",
    description: "Coffee",
    amount: "-4.50",
    reference: "",
  } as const;

  expect(
    effectiveTime({
      ...base,
      postingDateTime: "2026-10-02T00:00:00Z",
      executionDateTime: "2026-10-01T00:00:00Z",
    }),
  ).toBe("2026-10-02T00:00:00Z");
  expect(effectiveTime({ ...base, executionDateTime: "2026-10-01T00:00:00Z" })).toBe(
    "2026-10-01T00:00:00Z",
  );
});
