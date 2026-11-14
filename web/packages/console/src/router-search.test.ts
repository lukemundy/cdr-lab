import { expect, test } from "vitest";

import { parseSearch, stringifySearch } from "./router-search";

test("the search has the names and values of the API query", () => {
  expect(parseSearch("?q=a+b&page=2")).toEqual({ q: "a b", page: "2" });
  expect(stringifySearch({ q: "a b", page: 2 })).toBe("?q=a+b&page=2");
});

test("a search with no value has no query string", () => {
  expect(stringifySearch({ q: "", page: undefined, open: true })).toBe("");
});
