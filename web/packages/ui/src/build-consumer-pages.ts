// The stylesheet and the fonts of the consumer pages of the data holder, which its image copies and embeds.
import { cp, mkdir, readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";

const tokensDist = new URL("../../tokens/dist/", import.meta.url);
const outDir = new URL("../dist/consumer-pages/", import.meta.url);

const parts = await Promise.all(
  [
    new URL("tokens.css", tokensDist),
    new URL("fonts.css", tokensDist),
    new URL("consumer-pages.css", import.meta.url),
  ].map((file) => readFile(file, "utf8")),
);

await mkdir(outDir, { recursive: true });
// fonts.css refers to ./fonts/, which is next to consumer.css.
await cp(fileURLToPath(new URL("fonts/", tokensDist)), fileURLToPath(new URL("fonts/", outDir)), {
  recursive: true,
});
await writeFile(new URL("consumer.css", outDir), parts.join("\n"));
