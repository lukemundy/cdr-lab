import { readFile } from "node:fs/promises";
import { expect, test } from "vitest";

const tokensCSS = await readFile(new URL("../dist/tokens.css", import.meta.url), "utf8");

// DESIGN.md (Tokens): a density changes only spacing, control heights and table row heights.
test("the comfortable density sets only the density tokens, under its own selector", () => {
  const block = /\[data-density="comfortable"\] \{([^}]*)\}/.exec(tokensCSS)?.[1] ?? "";
  const names = [...block.matchAll(/(--ds-[\w-]+):/g)].map((m) => m[1] ?? "");
  expect(names.toSorted((a, b) => a.localeCompare(b))).toEqual([
    "--ds-size-control",
    "--ds-size-table-row",
    "--ds-space-cell-x",
    "--ds-space-cell-y",
    "--ds-space-control-x",
  ]);
  expect(block).toContain("--ds-size-control: 48px;");
});
