import { describe, expect, test } from "vitest";

import {
  additionalValue,
  formatAmount,
  formatDate,
  formatDuration,
  formatFrequency,
  formatRate,
} from "./format";

describe("a RateString", () => {
  test.each([
    ["0.0125", "1.25%"],
    ["0.05", "5.00%"],
    ["0.06125", "6.125%"],
    ["0", "0.00%"],
    ["-0.005", "-0.50%"],
  ])("%s shows as %s", (rate, want) => {
    expect(formatRate(rate)).toBe(want);
  });
});

describe("an AmountString", () => {
  test.each([
    ["1234.5", "$1,234.50"],
    ["0.00", "$0.00"],
    ["10", "$10.00"],
    ["-4.50", "-$4.50"],
  ])("%s shows as %s", (amount, want) => {
    expect(formatAmount(amount)).toBe(want);
  });
});

describe("an ISO 8601 duration", () => {
  test.each([
    ["P55D", "55 days"],
    ["P1D", "1 day"],
    ["P1M", "1 month"],
    ["P12M", "12 months"],
    ["P5Y", "5 years"],
    ["P2W", "2 weeks"],
    ["P1Y6M", "1 year and 6 months"],
    ["P1Y2M3D", "1 year, 2 months and 3 days"],
  ])("%s shows as %s", (duration, want) => {
    expect(formatDuration(duration)).toBe(want);
  });

  test.each(["PT12H", "P", "5 days", "P1.5Y", ""])(
    "%s that the site cannot put in words shows as it is",
    (value) => {
      expect(formatDuration(value)).toBe(value);
    },
  );
});

describe("a frequency", () => {
  test.each([
    ["P1M", "Each month"],
    ["P1Y", "Each year"],
    ["P1W", "Each week"],
    ["P1D", "Each day"],
    ["P3M", "Every 3 months"],
    ["P1Y6M", "Every 1 year and 6 months"],
    ["PT1H", "PT1H"],
  ])("%s shows as %s", (frequency, want) => {
    expect(formatFrequency(frequency)).toBe(want);
  });
});

// Verifies: DH-1833
test("a date shows in Australian English with no time", () => {
  expect(formatDate("2027-03-01T00:00:00+08:00")).toBe("1 March 2027");
});

describe("an additionalValue", () => {
  test.each([
    {
      kind: "feature",
      type: "INTEREST_FREE",
      value: "P55D",
      want: { term: "Period", text: "55 days" },
    },
    {
      kind: "feature",
      type: "FUNDS_AVAILABLE_AFTER",
      value: "P31D",
      want: { term: "Period", text: "31 days" },
    },
    {
      kind: "feature",
      type: "CASHBACK_OFFER",
      value: "100.00",
      want: { term: "Amount", text: "$100.00" },
    },
    { kind: "feature", type: "FREE_TXNS", value: "20", want: { term: "Number", text: "20" } },
    {
      kind: "feature",
      type: "BONUS_REWARDS",
      value: "60000",
      want: { term: "Number", text: "60,000" },
    },
    {
      kind: "feature",
      type: "DIGITAL_WALLET",
      value: "Apple Pay",
      want: { term: "Details", text: "Apple Pay" },
    },
    { kind: "feature", type: "OVERDRAFT", value: "5000", want: { term: "Details", text: "5000" } },
    { kind: "fee", type: "PERIODIC", value: "P1M", want: { term: "Charged", text: "Each month" } },
    { kind: "fee", type: "LATE_PAYMENT", value: "5", want: { term: "Days late", text: "5" } },
    {
      kind: "depositRate",
      type: "FIXED",
      value: "P12M",
      want: { term: "Period", text: "12 months" },
    },
    {
      kind: "depositRate",
      type: "BONUS",
      value: "Deposit $50",
      want: { term: "Details", text: "Deposit $50" },
    },
    {
      kind: "lendingRate",
      type: "INTRODUCTORY",
      value: "P2Y",
      want: { term: "Period", text: "2 years" },
    },
    {
      kind: "lendingRate",
      type: "PENALTY",
      value: "Late",
      want: { term: "Details", text: "Late" },
    },
  ] as const)("$kind $type $value", ({ kind, type, value, want }) => {
    expect(additionalValue(kind, type, value)).toEqual(want);
  });

  // Verifies: DH-1823
  test("a value that is not in the format of its type shows as it is", () => {
    expect(additionalValue("feature", "FREE_TXNS", "unlimited")).toEqual({
      term: "Number",
      text: "unlimited",
    });
    expect(additionalValue("feature", "CASHBACK_OFFER", "lots")).toEqual({
      term: "Amount",
      text: "lots",
    });
  });
});
