import { describe, expect, test } from "vitest";

import {
  changedSearch,
  listSearchValidator,
  oneOf,
  orderOf,
  tableSort,
  validatePageSearch,
} from "./list-search";

const validate = listSearchValidator(["ACTIVE", "INACTIVE"] as const, ["name", "status"] as const);

describe("listSearchValidator", () => {
  test("keeps the valid values", () => {
    expect(
      validate({ status: "ACTIVE", q: "bank", sort: "status", order: "desc", page: "3" }),
    ).toEqual({ status: "ACTIVE", q: "bank", sort: "status", order: "desc", page: 3 });
  });

  test("removes a value that is not valid, so that an old link still shows a list", () => {
    expect(
      validate({ status: "REMOVED", q: "", sort: "secret", order: "up", page: "1.5", other: "x" }),
    ).toEqual({});
  });

  test("removes page 1, which is the default", () => {
    expect(validate({ page: "1" })).toEqual({});
  });
});

test("a change of the search goes back to page 1", () => {
  expect(changedSearch(validate, { q: "bank", page: 4 }, { status: "ACTIVE" })).toEqual({
    q: "bank",
    status: "ACTIVE",
  });
});

test("oneOf gives the value only when it is one of the values", () => {
  expect(oneOf(["a", "b"], "b")).toBe("b");
  expect(oneOf(["a", "b"], "c")).toBeUndefined();
});

test("the table sort and the order of the search are the same sort", () => {
  expect(tableSort({}, "name")).toEqual({ column: "name", direction: "ascending" });
  expect(tableSort({ sort: "status", order: "desc" }, "name")).toEqual({
    column: "status",
    direction: "descending",
  });
  expect(orderOf({ column: "name", direction: "descending" })).toBe("desc");
  expect(orderOf({ column: "name", direction: "ascending" })).toBe("asc");
});

test("validatePageSearch keeps only a page after page 1", () => {
  expect(validatePageSearch({ page: "3", q: "bank" })).toEqual({ page: 3 });
  expect(validatePageSearch({ page: "1" })).toEqual({});
  expect(validatePageSearch({ page: "2.5" })).toEqual({});
});
