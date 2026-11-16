import { readdir, readFile } from "node:fs/promises";
import Color from "colorjs.io";
import { describe, expect, test } from "vitest";
import { consoleShellPairs, contrastPairs } from "./contrast.ts";
import { consoleColor, participantConsoleRoles } from "./shell.ts";
import { loadTokens, toColor } from "./tokens.ts";

// The preset brands of the data holder (services/dataholder/brands). Each one has the colour roles of the design
// system, and services/dataholder/design/brand.md requires the same contrast as for the brands here.
const presetsURL = new URL("../../../../services/dataholder/brands/", import.meta.url);

const presets = await Promise.all(
  (await readdir(presetsURL, { withFileTypes: true }))
    .filter((entry) => entry.isDirectory())
    .map(async ({ name }) => {
      const tokens: unknown = JSON.parse(
        await readFile(new URL(`${name}/tokens.json`, presetsURL), "utf8"),
      );
      return { name, tokens };
    }),
);

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

// The data holder accepts only literal OKLCH values (ADR 0023).
function colorOf(tokens: unknown, id: string) {
  const group = isRecord(tokens) ? tokens.color : undefined;
  const token = isRecord(group) ? group[id.replace(/^color\./, "")] : undefined;
  const value = isRecord(token) ? token.$value : undefined;
  if (!isRecord(value) || value.colorSpace !== "oklch" || !Array.isArray(value.components)) {
    throw new Error(`${id} is not a literal OKLCH colour of the preset`);
  }
  const [l = 0, c = 0, h = 0] = value.components.map(Number);
  return new Color("oklch", [l, c, h], typeof value.alpha === "number" ? value.alpha : 1);
}

test("the data holder has a preset brand", () => {
  expect(presets.length).toBeGreaterThan(0);
});

// The console of a preset uses the lab brand for each role that the brand does not give to its console.
const { resolver, permutations } = await loadTokens();
const lab = resolver.apply(permutations[0] ?? {});

function labColor(id: string) {
  const token = lab[id];
  if (token?.$type !== "color") throw new Error(`${id} is not a colour token of the lab brand`);
  return toColor(token.$value);
}

describe.each(presets)("preset $name", ({ tokens }) => {
  test.each(contrastPairs)("$fg on $bg has a contrast of $min:1 or more", ({ fg, bg, min }) => {
    expect(colorOf(tokens, fg).contrast(colorOf(tokens, bg), "WCAG21")).toBeGreaterThanOrEqual(min);
  });

  const consoleRole = (id: string) =>
    participantConsoleRoles.includes(id) ? colorOf(tokens, id) : labColor(id);

  test.each(consoleShellPairs)(
    "console: $fg on $bg has a contrast of $min:1 or more",
    ({ fg, bg, min }) => {
      expect(
        consoleColor(fg, consoleRole).contrast(consoleColor(bg, consoleRole), "WCAG21"),
      ).toBeGreaterThanOrEqual(min);
    },
  );
});
