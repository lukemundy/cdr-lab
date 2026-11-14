import { expect, test } from "vitest";

import { formatAmount, formatBsb, formatDate } from "./format";

test("an amount has its sign, the dollar sign, the group separators and 2 decimal digits", () => {
  expect(formatAmount("2500.00")).toBe("$2,500.00");
  expect(formatAmount("-4.50")).toBe("-$4.50");
  expect(formatAmount("0.00")).toBe("$0.00");
  expect(formatAmount("-0.00")).toBe("$0.00");
});

test("an amount keeps each digit, also above the precision of a number", () => {
  expect(formatAmount("9999999999999999.99")).toBe("$9,999,999,999,999,999.99");
});

test("a date is the day, the month in full and the year, with no change of the day for the time zone", () => {
  // The tests run in Australia/Perth (vitest.config.ts), and the API gives the date in UTC.
  expect(formatDate("2026-10-15")).toBe("15 October 2026");
  expect(formatDate("2026-01-01")).toBe("1 January 2026");
});

test("a value that is not an amount shows as it is", () => {
  expect(formatAmount("1e3")).toBe("1e3");
});

test("a BSB is two groups of three digits with a hyphen", () => {
  expect(formatBsb("062000")).toBe("062-000");
});

test("a value that is not a BSB shows as it is", () => {
  expect(formatBsb("06200")).toBe("06200");
  expect(formatBsb("062-000")).toBe("062-000");
});
