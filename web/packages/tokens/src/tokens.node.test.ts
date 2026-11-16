import { describe, expect, test } from "vitest";
import { consoleShellPairs, contrastPairs } from "./contrast.ts";
import { consoleColor } from "./shell.ts";
import { colorRoles, loadTokens, toColor } from "./tokens.ts";

// The semantic colour roles (web/packages/ui/design/tokens-and-brands.md, UI-0010). A new role needs a change to that
// list first.
const designRoles = [
  "color.page",
  "color.surface",
  "color.raised",
  "color.backdrop",
  "color.fg",
  "color.fg-muted",
  "color.fg-on-primary",
  "color.heading",
  "color.link",
  "color.brand",
  "color.fg-on-brand",
  "color.hero",
  "color.fg-on-hero",
  "color.accent",
  "color.border",
  "color.border-strong",
  "color.focus",
  "color.primary",
  "color.primary-hover",
  "color.primary-pressed",
  "color.secondary",
  ...["success", "warning", "danger", "info"].flatMap((status) => [
    `color.${status}-bg`,
    `color.${status}-border`,
    `color.${status}-fg`,
  ]),
];

const { resolver, permutations } = await loadTokens();

// Verifies: UI-0010
test("the semantic colour roles are the roles of the specification", () => {
  for (const input of permutations) {
    expect(new Set(colorRoles(resolver.apply(input)))).toEqual(new Set(designRoles));
  }
});

describe.each(permutations)("permutation %o", (input) => {
  const tokens = resolver.apply(input);
  const color = (id: string) => {
    const token = tokens[id];
    if (token?.$type !== "color") throw new Error(`${id} is not a colour token`);
    return toColor(token.$value);
  };

  // Verifies: UI-0013, UI-0014, UI-0015
  test.each(contrastPairs)("$fg on $bg has a contrast of $min:1 or more", ({ fg, bg, min }) => {
    expect(color(fg).contrast(color(bg), "WCAG21")).toBeGreaterThanOrEqual(min);
  });

  // Verifies: UI-0016
  test.each(consoleShellPairs)(
    "console shell: $fg on $bg has a contrast of $min:1 or more",
    ({ fg, bg, min }) => {
      expect(
        consoleColor(fg, color).contrast(consoleColor(bg, color), "WCAG21"),
      ).toBeGreaterThanOrEqual(min);
    },
  );

  // The contrast is correct only when the browser shows the colour as written.
  test.each(colorRoles(tokens))("%s is in the sRGB gamut", (id) => {
    expect(color(id).inGamut("srgb")).toBe(true);
  });
});
