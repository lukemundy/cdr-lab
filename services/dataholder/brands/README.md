# Data holder brands

A data holder instance gets its brand from a brand directory. [ADR 0023](../../../docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md) gives the format. [The brand directory library](../../../pkg/platform/branddir/DESIGN.md) gives the rules that each brand directory of the lab obeys, and [the brand area of the specification](../design/brand.md) gives the files and the rules of the data holder.

Each subdirectory here is a preset brand. The binary contains each preset, and the name of the directory is the name of the preset.

## Select a brand

Set exactly one of these settings for `dataholder api`, `dataholder serve` and `dataholder seed`:

| Setting | Value |
|---|---|
| `DATA_HOLDER_BRAND` | The name of a preset brand |
| `DATA_HOLDER_BRAND_DIR` | The path of a mounted brand directory |

The brand name of the directory must be the `brandName` of the brand on the Register. `dataholder seed -if-empty` with no file loads the catalogue of the brand.

## Add a brand

1. Make a directory with these files, and no other files:
   - `brand.json`: `{"brandName": "<the brand name>", "bsb": "<6 digits>", "layout": "<layout>"}`. The brand name has no control characters. The BSB is the BSB of the accounts that a consumer opens on the online bank site. The layout of the online bank site is `meadow`, `ledger`, `horizon` or `updraft` ([the layout of the brand file](../design/brand.md#dh-0335)).
   - `tokens.json`: a copy of a preset `tokens.json`, with the values of the brand. Each colour is OKLCH, in the sRGB gamut. The first font of each font token (`body`, `code` and `heading`) is a font of the data holder (the families in `web/packages/tokens/fonts.json`), and the other fonts are CSS generic families.
   - `logo.svg`: the logo, legible on the `color.brand` of the brand (the header of the pages) and on white (the pages of a data recipient). `mark.svg`: a square symbol, legible on `color.hero`. Each one is 100 KiB or less, with an `svg` root element in the SVG namespace and a `viewBox`.
   - `logo.png`: the logo of `logo.svg` for the email, legible on `color.brand` (the header of the email). It is 100 KiB or less, exactly 96 pixels high and 768 pixels wide or less, with the aspect ratio of the `viewBox` of `logo.svg` (a tolerance of 2 %). See [Make logo.png](#make-logopng).
   - `products.json`: the catalogue, in the format of the product load file. The `brand` of each product is the brand name or the label of a sub-brand, for example `Wattle Bank Business`.

   - Optional photos for the online bank site, as WebP images that are not animated: `hero.webp` (256 KiB or less, 1200 to 2400 pixels wide, with the aspect ratio of the layout: 4:3 for `meadow`, 21:9 for `ledger`, 3:1 for `horizon`, 1:1 for `updraft`), and `category-<anchor>.webp` for a CDS product category, for example `category-term-deposits.webp` (96 KiB or less, 800 to 1200 pixels wide, 4:3). A photo shows no logo. It can show text, but the text must not give information that the page does not also give. The site shows a shape in the brand colour in place of a photo that is not there. See [Make the photos](#make-the-photos).

   - Optional `ATTRIBUTION.md`, 16 KiB or less: the source, the author and the licence of each photo. The data holder does not read or serve it. The attribution file of each preset names each of its photos.

   Each JSON file is 1 MiB or less, and the directory is 1 MiB or less in all. A file can be a symbolic link to a file in the directory, but not to a file outside it.
2. Check the contrast of the colours. The data holder does not check it at start-up. For a preset, `mise run test:web` checks it.
3. For a preset, put the directory here. For a user brand, mount the directory into the container, read-only, and set `DATA_HOLDER_BRAND_DIR`.
4. Set the `logoUri` of the brand on the Register to `<public base URI>/brand/logo.svg`.

When a rule fails, the data holder does not start, and its error names each file and rule that fails.

## Make the photos

Any tool that crops an image and writes WebP can make the photos. The photos of the presets are [Unsplash](https://unsplash.com) photos under the [Unsplash License](https://unsplash.com/license), which permits copies and distribution. The `ATTRIBUTION.md` of each preset names each photo and its author. Each photo is a crop of the original to the size of the slot, made with ImageMagick:

```sh
magick original.jpg -auto-orient -resize 1000x750^ -gravity center -extent 1000x750 -strip -quality 74 category-term-deposits.webp
```

Lower the quality, or use the minimum width, until the file is under its limit. When a centre crop cuts the subject, change the gravity. Look at each photo before you commit it. A preset photo has a licence that permits distribution, because the repository and the images give copies of the presets to everyone.

## Make logo.png

Any tool that renders an SVG to a PNG 96 pixels high with a transparent background can make `logo.png`. The PNG files of the presets come from Chromium, through the Playwright of the web workspace. To make them, run this from `web/` after `pnpm install`, with the directory of each brand:

```sh
node --input-type=module -e '
import { chromium } from "playwright";
import { readFileSync } from "node:fs";
const dir = process.argv[1];
const svg = readFileSync(`${dir}/logo.svg`);
const [, , w, h] = svg.toString().match(/viewBox="([^"]+)"/)[1].split(/[\s,]+/).map(Number);
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: Math.round((48 * w) / h), height: 48 }, deviceScaleFactor: 2 });
await page.setContent(`<body style="margin:0"><img style="display:block;width:100vw;height:100vh" src="data:image/svg+xml;base64,${svg.toString("base64")}">`);
await page.screenshot({ path: `${dir}/logo.png`, omitBackground: true });
await browser.close();
' ../services/dataholder/brands/wattle-bank
```

Look at the PNG before you commit it.

The data holder serves `/brand/logo.svg`, `/brand/mark.svg` and `/brand/styles.css` on its public base URI, and the online bank site also serves the photos at `/brand/<file name>`. The stylesheet sets the custom properties (`--ds-*`) of the design system with the values of the brand. The operator console also serves the logo, the mark and `/brand/console.css`, which sets only [the roles of the brand that the console uses](../design/operator-console.md#dh-1486), for example `--ds-color-brand` and `--ds-color-primary`. [The roles of the brand in the console of a participant](../../../web/packages/ui/design/tokens-and-brands.md#ui-0035) gives the full list. The OTP email shows `logo.png` in a part of the email, and uses the colours of the brand in hex notation.

The lab runs the default preset brand, `wattle-bank` (`compose.yaml`).
