import { describe, expect, test } from "vitest";
import { findViolations } from "./check-token-use.ts";

describe("findViolations", () => {
  test.each([
    ["a primitive token in CSS", ".a { color: var(--ds-primitive-color-blue-600); }"],
    [
      "a primitive token in a Tailwind class",
      '<p className="text-(--ds-primitive-color-blue-600)">',
    ],
    ["a hex colour", ".a { color: #1a2b3c; }"],
    ["a short hex colour in a Tailwind class", '<p className="bg-[#fff]">'],
    ["an oklch() colour", ".a { color: oklch(50% 0.1 260); }"],
    ["an rgb() colour", ".a { border-color: rgb(0 0 0 / 0.1); }"],
    ["an hsl() colour", ".a { color: hsla(0, 0%, 0%, 0.5); }"],
    ["a color-mix() colour", ".a { color: color-mix(in oklch, var(--ds-color-primary), white); }"],
  ])("finds %s", (_, source) => {
    expect(findViolations(source)).toHaveLength(1);
  });

  test.each([
    ["a semantic token in CSS", ".a { color: var(--ds-color-fg); }"],
    ["a semantic Tailwind class", '<p className="bg-surface text-fg-muted border-border-strong">'],
    ["a link to a fragment", '<a href="#main">Skip to content</a>'],
    ["a colour keyword", ".a { background: transparent; color: currentColor; }"],
  ])("accepts %s", (_, source) => {
    expect(findViolations(source)).toEqual([]);
  });

  test("gives the line of each violation", () => {
    expect(findViolations(".a {\n  color: #000;\n}")).toEqual([{ line: 2, text: "#000" }]);
  });
});
