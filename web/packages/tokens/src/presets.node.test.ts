import { readdir, readFile } from "node:fs/promises";
import Color from "colorjs.io";
import { describe, expect, test } from "vitest";
import { consoleShellPairs, contrastPairs } from "./contrast.ts";
import { consoleColor, participantConsoleRoles } from "./shell.ts";
import { loadTokens, toColor } from "./tokens.ts";

// The preset brands of the data holder (services/dataholder/brands) and of the data recipient
// (services/datarecipient/brands). Each one has the colour roles of the design system, and the specification of each
// service requires the same contrast as for the brands here. Only the data holder has a console.
const services = [
  { service: "dataholder", hasConsole: true },
  { service: "datarecipient", hasConsole: false },
];

const presets = (
  await Promise.all(
    services.map(async ({ service, hasConsole }) => {
      const presetsURL = new URL(`../../../../services/${service}/brands/`, import.meta.url);
      return Promise.all(
        (await readdir(presetsURL, { withFileTypes: true }))
          .filter((entry) => entry.isDirectory())
          .map(async ({ name }) => {
            const tokens: unknown = JSON.parse(
              await readFile(new URL(`${name}/tokens.json`, presetsURL), "utf8"),
            );
            return { service, name, hasConsole, tokens };
          }),
      );
    }),
  )
).flat();

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

// A brand directory has only literal OKLCH values (ADR 0023).
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

test.each(services)("$service has a preset brand", ({ service }) => {
  expect(presets.filter((preset) => preset.service === service).length).toBeGreaterThan(0);
});

// The console of a preset uses the lab brand for each role that the brand does not give to its console.
const { resolver, permutations } = await loadTokens();
const lab = resolver.apply(permutations[0] ?? {});

function labColor(id: string) {
  const token = lab[id];
  if (token?.$type !== "color") throw new Error(`${id} is not a colour token of the lab brand`);
  return toColor(token.$value);
}

describe.each(presets)("$service preset $name", ({ tokens, hasConsole }) => {
  test.each(contrastPairs)("$fg on $bg has a contrast of $min:1 or more", ({ fg, bg, min }) => {
    expect(colorOf(tokens, fg).contrast(colorOf(tokens, bg), "WCAG21")).toBeGreaterThanOrEqual(min);
  });

  const consoleRole = (id: string) =>
    participantConsoleRoles.includes(id) ? colorOf(tokens, id) : labColor(id);

  test.skipIf(!hasConsole).each(consoleShellPairs)(
    "console: $fg on $bg has a contrast of $min:1 or more",
    ({ fg, bg, min }) => {
      expect(
        consoleColor(fg, consoleRole).contrast(consoleColor(bg, consoleRole), "WCAG21"),
      ).toBeGreaterThanOrEqual(min);
    },
  );
});
