import { expect, test } from "vitest";

import { linesOf, textOf } from "./form-values";

function formData(entries: Record<string, string>): FormData {
  const data = new FormData();
  for (const [name, value] of Object.entries(entries)) data.set(name, value);
  return data;
}

test("a text value has no white space at the start and at the end", () => {
  expect(textOf(formData({ name: "  Bank \n" }), "name")).toBe("Bank");
  expect(textOf(formData({}), "name")).toBe("");
});

test("a list of values has one value on each line, with no empty lines", () => {
  expect(
    linesOf(formData({ uris: " https://a.example \n\n https://b.example\n" }), "uris"),
  ).toEqual(["https://a.example", "https://b.example"]);
});
