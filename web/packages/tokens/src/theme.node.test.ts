import { readFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { compile } from "tailwindcss";
import { describe, expect, test } from "vitest";
import { findViolations } from "./check-token-use.ts";
import { shellMixCSS, shellMixes } from "./shell.ts";
import { colorRoles, loadTokens } from "./tokens.ts";

const packageDir = dirname(dirname(fileURLToPath(import.meta.url)));
const tailwindDir = dirname(fileURLToPath(import.meta.resolve("tailwindcss/package.json")));

const theme = await readFile(join(packageDir, "src/theme.css"), "utf8");
const tokensCSS = await readFile(join(packageDir, "dist/tokens.css"), "utf8");
const { resolver, permutations } = await loadTokens();

async function build(classes: string[]) {
  const compiler = await compile(
    `@import "tailwindcss/theme.css" layer(theme);\n@import "tailwindcss/utilities.css" layer(utilities);\n${theme}`,
    {
      base: packageDir,
      loadStylesheet: async (id) => {
        const path = join(tailwindDir, id.replace(/^tailwindcss\//, ""));
        return { path, base: dirname(path), content: await readFile(path, "utf8") };
      },
    },
  );
  return compiler.build(classes);
}

describe("theme.css", () => {
  test.each(permutations.flatMap((input) => colorRoles(resolver.apply(input))))(
    "gives a Tailwind colour for %s",
    (id) => {
      const name = id.replace(/^color\./, "");
      expect(theme).toContain(`--color-${name}: var(--ds-color-${name});`);
    },
  );

  // The formatter can break a long declaration over more than one line.
  const flatTheme = theme.replaceAll(/\s+/g, " ").replaceAll("( ", "(").replaceAll(" )", ")");

  test.each(shellMixes)("gives the shell colour $id", (mix) => {
    expect(flatTheme).toContain(shellMixCSS(mix));
  });

  test("refers only to variables that tokens.css defines", () => {
    const used = new Set([...theme.matchAll(/var\((--ds-[\w-]+)\)/g)].map((m) => m[1]));
    expect(used.size).toBeGreaterThan(0);
    for (const name of used) expect(tokensCSS).toContain(`${name}:`);
  });

  test("a semantic utility uses the semantic variable", async () => {
    const css = await build([
      "bg-surface",
      "text-fg-muted",
      "border-border-strong",
      "p-4",
      "h-control",
      "rounded-md",
    ]);
    expect(css).toContain("background-color: var(--ds-color-surface)");
    expect(css).toContain("color: var(--ds-color-fg-muted)");
    expect(css).toContain("border-color: var(--ds-color-border-strong)");
    expect(css).toContain("calc(var(--ds-space-1) * 4)");
    expect(css).toContain("height: var(--ds-size-control)");
    expect(css).toContain("border-radius: var(--ds-radius-md)");
  });

  test("the default Tailwind palette and shadows give no utilities", async () => {
    const css = await build(["bg-red-500", "text-white", "bg-black", "shadow-sm", "shadow-lg"]);
    expect(css).not.toMatch(/\.(bg-red-500|text-white|bg-black|shadow-sm|shadow-lg)\b/);
  });

  // The @property rules of Tailwind have raw initial values, so only the utilities count.
  test("the utilities have no raw colour value", async () => {
    const css = await build([
      "bg-page",
      "text-fg",
      "border-border",
      "shadow-1",
      "shadow-2",
      "shadow-3",
      "font-sans",
      "font-mono",
      "font-heading",
    ]);
    const utilities = css.slice(css.indexOf("@layer utilities"), css.indexOf("@property"));
    expect(utilities).toContain("var(--ds-elevation-3)");
    expect(utilities).toContain("var(--ds-font-heading)");
    expect(findViolations(utilities)).toEqual([]);
  });
});
