import { describe, expect, test } from "vitest";

import { productCategoryValues } from "../api/client";
import { amountStringOf, openValueOf } from "./opening";

describe("the value of an open", () => {
  test.each([
    ["CRED_AND_CHRG_CARDS", "creditLimit", "Credit limit"],
    ["OVERDRAFTS", "creditLimit", "Credit limit"],
    ["MARGIN_LOANS", "creditLimit", "Credit limit"],
    ["BUY_NOW_PAY_LATER", "creditLimit", "Credit limit"],
    ["PERS_LOANS", "loanAmount", "Loan amount"],
    ["BUSINESS_LOANS", "loanAmount", "Loan amount"],
    ["RESIDENTIAL_MORTGAGES", "loanAmount", "Loan amount"],
    ["LEASES", "loanAmount", "Loan amount"],
    ["TRADE_FINANCE", "loanAmount", "Loan amount"],
  ] as const)("of %s is %s, with the label %s", (category, name, label) => {
    expect(openValueOf(category)).toMatchObject({ name, label });
  });

  test.each([
    "TRANS_AND_SAVINGS_ACCOUNTS",
    "TERM_DEPOSITS",
    "TRAVEL_CARDS",
    "REGULATED_TRUST_ACCOUNTS",
  ] as const)("of the deposit product %s is none", (category) => {
    expect(openValueOf(category)).toBeUndefined();
  });

  test("of each lending category is one of the site API", () => {
    const lending = productCategoryValues.filter((c) => openValueOf(c) !== undefined);
    expect(lending).toHaveLength(9);
  });
});

describe("an amount that a person types", () => {
  test.each([
    ["5000", "5000.00"],
    ["5,000", "5000.00"],
    ["$5,000.50", "5000.50"],
    ["$ 5000.5", "5000.50"],
    [" 1,234,567.89 ", "1234567.89"],
    ["0", "0.00"],
    ["007", "7.00"],
    ["$0.99", "0.99"],
  ])("%s is the AmountString %s", (text, want) => {
    expect(amountStringOf(text)).toBe(want);
  });

  test.each([
    "",
    "five thousand",
    "5,00",
    "50,00.00",
    "5000.123",
    "-5000",
    "5.000,00",
    "$",
    "1e5",
    ".5",
  ])("%j is not an amount", (text) => {
    expect(amountStringOf(text)).toBeUndefined();
  });
});
