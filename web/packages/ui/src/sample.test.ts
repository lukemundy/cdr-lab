import { expect, test } from "vitest";

// jsdom and happy-dom do no layout, so this placeholder passes only in a real browser.
test("the browser lays out an element", () => {
  const box = document.createElement("div");
  box.style.width = "120px";
  document.body.append(box);

  expect(box.getBoundingClientRect().width).toBe(120);
});
