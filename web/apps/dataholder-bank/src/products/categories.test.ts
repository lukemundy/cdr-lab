import { expect, test } from "vitest";

import type { Product } from "../api/client";
import { categoryAnchor, groupByCategory, productCount } from "./categories";

function product(
  productId: string,
  name: string,
  productCategory: Product["productCategory"],
): Product {
  return {
    productId,
    name,
    productCategory,
    description: "A product.",
    brand: "Wattle Bank",
    isTailored: false,
    lastUpdated: "2026-10-01T00:00:00Z",
  };
}

// Verifies: DH-1812
test("the groups are in the order of the categories of the CDS, with only the categories that have a product", () => {
  const groups = groupByCategory([
    product("loan", "Home Loan", "RESIDENTIAL_MORTGAGES"),
    product("saver", "Saver", "TRANS_AND_SAVINGS_ACCOUNTS"),
    product("term", "Term Deposit", "TERM_DEPOSITS"),
    product("card", "Card", "CRED_AND_CHRG_CARDS"),
  ]);

  expect(groups.map((group) => group.category)).toEqual([
    "TERM_DEPOSITS",
    "TRANS_AND_SAVINGS_ACCOUNTS",
    "CRED_AND_CHRG_CARDS",
    "RESIDENTIAL_MORTGAGES",
  ]);
});

// Verifies: DH-1813
test("in a group, the products are in the order of their names", () => {
  const [group] = groupByCategory([
    product("c", "zip Saver", "TRANS_AND_SAVINGS_ACCOUNTS"),
    product("a", "Everyday", "TRANS_AND_SAVINGS_ACCOUNTS"),
    product("b", "Bonus Saver", "TRANS_AND_SAVINGS_ACCOUNTS"),
  ]);

  expect(group?.products.map((p) => p.name)).toEqual(["Bonus Saver", "Everyday", "zip Saver"]);
});

test("no product gives no group", () => {
  expect(groupByCategory([])).toEqual([]);
});

test("the anchor of a category is its value in lower case with hyphens", () => {
  expect(categoryAnchor("TRANS_AND_SAVINGS_ACCOUNTS")).toBe("trans-and-savings-accounts");
});

test("the count of products is a phrase", () => {
  expect(productCount(1)).toBe("1 product");
  expect(productCount(3)).toBe("3 products");
});
