import { defineConfig } from "@terrazzo/cli";
import css from "@terrazzo/plugin-css";
import rolesPlugin from "./src/roles-plugin.ts";
import { cssVariable } from "./src/tokens.ts";

export default defineConfig({
  tokens: ["./tokens/resolver.json"],
  outDir: "./dist/",
  plugins: [
    css({
      filename: "tokens.css",
      variableName: (token) => cssVariable(token.id),
      // A second brand or density adds a permutation with its own selector. A consumer page sets data-density.
      permutations: [
        { input: {}, prepare: (contents) => `:root {\n${contents}\n}` },
        {
          input: { density: "comfortable" },
          // Only the tokens of the density modifier, so that the selector does not set the colours of a brand again.
          include: ["size.*", "space.control-x", "space.cell-x", "space.cell-y"],
          prepare: (contents) => `[data-density="comfortable"] {\n${contents}\n}`,
        },
      ],
    }),
    rolesPlugin(),
  ],
  lint: {
    rules: {
      // Colour values are in OKLCH (web/packages/ui/design/tokens-and-brands.md, UI-0012).
      "core/colorspace": ["error", { colorSpace: "oklch" }],
      // A colour outside sRGB changes in the browser, so its contrast test is not valid.
      "core/max-gamut": ["error", { gamut: "srgb" }],
    },
  },
});
