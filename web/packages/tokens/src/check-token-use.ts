import { glob, readFile } from "node:fs/promises";

const patterns = [
  /--ds-primitive-[\w-]*/g,
  /(?<![\w&#/-])#(?:[\da-f]{8}|[\da-f]{6}|[\da-f]{3,4})(?![\w-])/gi,
  /\b(?:rgba?|hsla?|hwb|lab|lch|oklab|oklch|color|color-mix|light-dark)\(/gi,
];

export interface Violation {
  line: number;
  text: string;
}

export function findViolations(source: string): Violation[] {
  return source
    .split("\n")
    .flatMap((text, i) =>
      patterns.flatMap((pattern) =>
        [...text.matchAll(pattern)].map((m) => ({ line: i + 1, text: m[0] })),
      ),
    );
}

// Only the tokens package has raw values, and a test can compare a computed colour.
const sources = ["{apps,packages}/*/src/**/*.{css,html,ts,tsx}"];
const excluded = ["packages/tokens/**", "**/*.test.{ts,tsx}"];

if (import.meta.main) {
  let failed = false;
  for await (const file of glob(sources, { exclude: excluded })) {
    for (const { line, text } of findViolations(await readFile(file, "utf8"))) {
      console.error(
        `${file}:${line}: ${text}: use a semantic token, not a primitive token or a raw colour`,
      );
      failed = true;
    }
  }
  if (failed) process.exit(1);
}
