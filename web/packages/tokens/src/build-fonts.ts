// Only the latin and latin-ext subsets of the normal style, so that a service that embeds them stays small.
import { copyFile, mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { fonts } from "./fonts.ts";

const subsets = ["latin", "latin-ext"];
const fontsDir = fileURLToPath(new URL("../dist/fonts/", import.meta.url));

async function copyFaces(name: string) {
  const dir = dirname(fileURLToPath(import.meta.resolve(`${name}/package.json`)));
  const family = name.split("/")[1] ?? name;
  const css = await readFile(join(dir, "index.css"), "utf8");
  const faces = subsets.map((subset) => {
    const face = new RegExp(
      String.raw`/\* ${family}-${subset}-wght-normal \*/\n@font-face \{[^}]*\}`,
    ).exec(css)?.[0];
    const file = face && /url\(\.\/files\/([^)]+)\)/.exec(face)?.[1];
    if (!face || !file) throw new Error(`${name} has no ${subset} font face`);
    return { face: face.replace("./files/", "./fonts/"), file };
  });
  await Promise.all([
    ...faces.map(({ file }) => copyFile(join(dir, "files", file), join(fontsDir, file))),
    // The SIL Open Font License 1.1 requires its text with each copy of the font.
    copyFile(join(dir, "LICENSE"), join(fontsDir, `${family}-LICENSE.txt`)),
  ]);
  return faces.map(({ face }) => face);
}

await mkdir(fontsDir, { recursive: true });
const faces = await Promise.all(fonts.map(({ package: name }) => copyFaces(name)));
await writeFile(join(fontsDir, "../fonts.css"), `${faces.flat().join("\n\n")}\n`);
