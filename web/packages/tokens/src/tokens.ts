import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import {
  calculatePermutations,
  defineConfig,
  parse,
  type ColorValueNormalized,
  type TokenNormalizedSet,
} from "@terrazzo/parser";
import Color from "colorjs.io";

export const resolverURL = new URL("../tokens/resolver.json", import.meta.url);

export const cssVariablePrefix = "--ds-";

export function cssVariable(id: string) {
  return `${cssVariablePrefix}${id.replaceAll(".", "-")}`;
}

export function isPrimitive(id: string) {
  return id.startsWith("primitive.");
}

export function colorRoles(tokens: TokenNormalizedSet) {
  return Object.values(tokens)
    .filter((token) => token.$type === "color" && !isPrimitive(token.id))
    .map((token) => token.id);
}

export function toColor(value: ColorValueNormalized) {
  const [a = 0, b = 0, c = 0] = value.components;
  return new Color(value.colorSpace, [a ?? 0, b ?? 0, c ?? 0], value.alpha);
}

export async function loadTokens() {
  const config = defineConfig(
    { tokens: [fileURLToPath(resolverURL)] },
    { cwd: new URL("../", import.meta.url) },
  );
  const { resolver } = await parse(
    [{ src: await readFile(resolverURL, "utf8"), filename: resolverURL }],
    {
      config,
    },
  );
  const modifiers = Object.entries(resolver.source.modifiers ?? {}).map(
    ([name, modifier]): [string, string[]] => [name, Object.keys(modifier.contexts)],
  );
  return { resolver, permutations: calculatePermutations(modifiers) };
}
