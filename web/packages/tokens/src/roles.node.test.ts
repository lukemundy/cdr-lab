import { readFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { expect, test } from "vitest";
import { findViolations } from "./check-token-use.ts";
import { colorRoles, loadTokens } from "./tokens.ts";

const packageDir = dirname(dirname(fileURLToPath(import.meta.url)));
const page = await readFile(join(packageDir, "dist/roles.html"), "utf8");
const { resolver, permutations } = await loadTokens();

test("roles.html has no script", () => {
  expect(page).not.toMatch(/<script|\son\w+=|javascript:/i);
});

test("roles.html loads only the token CSS", () => {
  const links = [...page.matchAll(/<link [^>]*href="([^"]+)"/g)].map((m) => m[1]);
  expect(links).toHaveLength(2);
  expect(new Set(links)).toEqual(new Set(["tokens.css", "fonts.css"]));
  expect(page).not.toMatch(/(src|href)="(https?:)?\/\//);
  expect(page).not.toMatch(/\sstyle=/);
});

test("roles.html takes each colour from a token", () => {
  expect(findViolations(page)).toEqual([]);
});

test.each(permutations.flatMap((input) => colorRoles(resolver.apply(input))))(
  "roles.html shows %s",
  (id) => {
    expect(page).toContain(`var(--ds-${id.replaceAll(".", "-")})`);
  },
);
