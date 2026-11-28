# Data recipient brands

A data recipient instance gets its brand from a brand directory, in the format of the data holder. [The brand directory library](../../../pkg/platform/branddir/DESIGN.md) gives the rules that each brand directory of the lab obeys, and [the brand area of the specification](../design/brand.md) gives the files and the rules of the data recipient.

Each subdirectory here is a preset brand. The binary contains each preset, and the name of the directory is the name of the preset. `sorrel` is the default preset brand.

## Add a brand

1. Make a directory with these files, and no other files:
   - `brand.json`: `{"brandName": "<the brand name>"}`. The brand name has no control characters, and it is the `brandName` of the brand on the Register.
   - `tokens.json`: a copy of a preset `tokens.json`, with the values of the brand. Each colour is OKLCH, in the sRGB gamut. The first font of each font token is a font of the design system (the families in `web/packages/tokens/fonts.json`), and the other fonts are CSS generic families.
   - `logo.svg`: the logo, legible on the `color.brand` of the brand (the header) and on white. `mark.svg`: a square symbol, legible on `color.hero`. Each one is 100 KiB or less, with an `svg` root element in the SVG namespace and a `viewBox`.
   - `logo.png`: the logo of `logo.svg` for the emails, legible on `color.brand`. It is 100 KiB or less, exactly 96 pixels high and 768 pixels wide or less, with the aspect ratio of the `viewBox` of `logo.svg` (a tolerance of 2 %). The data holder brands README tells [how to make it](../../dataholder/brands/README.md#make-logopng).
   - Optional photos, as WebP images that are not animated:
     - `hero.webp`: the landing page. 256 KiB or less, 1200 to 2400 pixels wide, 4:3.
     - `access.webp`: the sign-up, sign-in and password reset pages. 160 KiB or less, 800 to 1600 pixels wide, 4:5.

     The data holder brands README tells [how to make the photos](../../dataholder/brands/README.md#make-the-photos).
   - Optional `ATTRIBUTION.md`, 16 KiB or less: the source, the author and the licence of each photo. The data recipient does not read or serve it.

   Each JSON file is 1 MiB or less, and the directory is 1 MiB or less in all.
2. Check the contrast of the colours. The data recipient does not check it at start-up. For a preset, `mise run test:web` checks it.
3. For a preset, put the directory here.

When a rule fails, the data recipient does not start, and its error names each file and rule that fails.
