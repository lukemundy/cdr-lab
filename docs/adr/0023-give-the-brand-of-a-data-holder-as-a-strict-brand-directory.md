# 0023. Give the brand of a data holder instance as a strict brand directory, with the preset brands in the binary

- Date: 2026-11-01

## Context

One instance of the data holder is one data holder brand. A brand needs a name, a logo, a small mark, the colours, the fonts and the radius of the design system, a layout and photos for the online bank site, and a product catalogue. Before this decision, the instance got only its brand name from a setting. The same binary could not run as a different bank, and a new brand needed a change to the code.

The format of the brand is a contract with users. A user adds a brand with files and configuration only, and the data holder must read the same files after later changes. Some parts of a brand are not the same in each deployment: the brand ID that the Register gives, the base URIs, the keys and certificates, and the sender address of the email.

These facts limit the options:

- The final image contains only the binary ([ADR 0018](0018-container-images-with-a-web-ui-build-stage.md)). Each file that a service reads at run time is in the binary (`go:embed`) or comes from a mount.
- The design system gives a brand as a context of the `brand` modifier of the DTCG tokens, with the token names of its brands. Its brand files use aliases to primitive tokens ([ADR 0016](0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md)). Only the web build has the primitive tokens.
- The CDS gives `brandName` a maximum length of 200, and makes `logoUri` mandatory for a data holder brand ([CDS: Register API schemas](https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#cdr-register-api-schemas)).
- A browser does not run a script in an SVG file that a page shows with `img` ([MDN: SVG as an image](https://developer.mozilla.org/en-US/docs/Web/SVG/Guides/SVG_as_an_image)). It runs the script when a user opens the URL of the file directly, unless the response has the CSP `sandbox` directive ([MDN: CSP sandbox](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/sandbox)).
- An email cannot load the logo from the public base URI when the reader has not permitted remote images ([Gmail Help: Turn images on or off](https://support.google.com/mail/answer/145919)). Some email clients do not show SVG, and all show PNG ([caniemail: SVG](https://www.caniemail.com/features/image-svg/), [caniemail: PNG](https://www.caniemail.com/features/image-png/)). The Go standard library decodes PNG, but it has no SVG rasteriser.
- A Kubernetes ConfigMap is at most 1 MiB, and the key of each file has only alphanumeric characters, `-`, `_` and `.`, so a ConfigMap cannot hold a subdirectory ([Kubernetes: ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)). A ConfigMap volume adds the entries `..data` and `..<timestamp>`, and each file is a symbolic link into them ([Kubernetes: atomic_writer.go](https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/util/atomic_writer.go)).

- Two sites with the same layout and different colours look like the same bank. A photo is approximately ten times the size of a logo. Each current browser shows WebP ([MDN: Image file type and format guide](https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/Image_types#webp_image)). The Go standard library has no WebP decoder, but the size of a WebP image is in its header ([RFC 9649](https://www.rfc-editor.org/rfc/rfc9649)).

The options for the preset brands, which the data holder contains:

- **Presets in the binary, and a mounted directory for a user brand.** The image runs each preset with no other file, and ADR 0018 stays true. One loader reads the two forms. The cost is two settings.
- **Presets copied into the image at a fixed path.** One setting, but the final image then contains files other than the binary, and the path in the image becomes a contract.
- **Presets mounted from the repository.** The image alone cannot run a preset, and a deployment from a published image or chart has no copy of the repository.

The options for the tokens:

- **Aliases, as in the design system.** A brand directory then needs the primitive tokens, which the data holder does not have at run time.
- **Literal values with the token names of the design system.** The data holder can convert the file to CSS with no web build. The values are longer to write.

The options for the layout of the online bank site:

- **A field of the brand that selects one of a fixed set of layouts in the site.** A user brand can use each layout, and a preset obeys the same contract. A new layout is a change to the site.
- **A layout for each preset in the code.** Each preset can have any layout, but a user brand cannot have one, and the presets then do not obey the contract of a user directory.
- **The tokens only (colours, fonts, radius).** No new field, but the sites then have the same structure, and differ only in colour and shape.

The options for change of the format:

- **A format version field.** It permits a later change that removes or renames a field, but it is a new convention, and one data holder reads only the presets that it contains and the directories of its users.
- **No version, and a strict check.** An error in a user directory shows at start-up. A new file or a new field is then a change of the contract.

## Decision

The data holder gets its brand from a brand directory. This decision applies to the data holder service only. It does not apply to the brand of the other participants, or to the brands of the design system.

- **Selection.** The configuration selects the brand with exactly one of two settings: the name of a preset brand, or the path of a mounted brand directory. Neither has a default, because a default is the brand of one deployment. The preset brands are brand directories in the source of the data holder, which the binary embeds, so the image runs each preset with no other file. One loader reads the two forms, so a preset obeys the same contract as a user directory.
- **Scope of the directory.** The directory holds only the parts of the brand that are the same in each deployment. The brand ID, the base URIs, the keys, the certificates and the sender address stay in the configuration of the deployment.
- **Files.** The directory has exactly these files: `brand.json` (the `brandName`, as the Register gives it, with no control character, because the name goes into the header of an email, and the `bsb` of the accounts that a consumer opens on the online bank site, with 6 digits, the rule of a BSB of an account), `tokens.json`, `logo.svg`, `logo.png` (the same logo for email), `mark.svg` and `products.json` (the format of the product load file). It can also have photos for the online bank site, which are optional: `hero.webp` and one `category-<anchor>.webp` for each CDS product category, with the anchor of the category in the site (for example `category-term-deposits.webp`). It can also have `ATTRIBUTION.md`, which is optional and credits the authors of the photos. The data holder does not read or serve it, because the credit is for the people who copy the directory, not for the pages. It has no other file and no subdirectory, because a ConfigMap cannot hold one, but the data holder ignores each entry whose name starts with `..`. Those are the names that a ConfigMap volume adds, and a file of the brand cannot have such a name. A symbolic link to a file in the directory is valid, because a ConfigMap volume gives each file as a link, but a link out of the directory is not valid. Each JSON file is 1 MiB or less, the size of a ConfigMap, each SVG and PNG file is 100 KiB or less, `hero.webp` is 256 KiB or less, each category photo is 96 KiB or less, and `ATTRIBUTION.md` is 16 KiB or less. The full directory is 1 MiB or less, so that it fits in one ConfigMap. The data holder checks the sizes before it reads a file.
- **Layout and photos.** `brand.json` has a required `layout`, one of a fixed set of layouts of the online bank site. A layout changes the shapes and the arrangement of the pages, not their content, their order or their accessibility. A photo is WebP, not animated, with the aspect ratio of its use: for the hero, the aspect ratio of the layout; for a category, 4:3. Each has a minimum and a maximum width, so that the photo is sharp and small. The photos are decorative: a photo can show text, for example a sign or a screen, but it carries no meaning that the page does not also give in text, so the directory gives no alt text.
- **No version, and a strict check.** The format has no version field. The data holder refuses an unknown field and an unknown file. A change that adds a file or a field is a change of this contract.
- **Tokens.** `tokens.json` has each token of a brand of the design system, with the same names, as DTCG 2025.10 literal values. An alias is not valid. Each colour is OKLCH in the sRGB gamut, as the design system requires ([ADR 0016](0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md)). A missing token is an error, so the colours of a brand never come from a different brand. The tokens have a body font and a heading font. The first font of a font token is one of a fixed set of fonts that the data holder contains, and the other fonts are CSS generic families. Thus a brand needs no font file and no licence check. A new font is a change to the data holder and to the design system.
- **Validation at start-up.** The process that serves the APIs, and the product load, validate the full directory at start-up: each file, field and colour, the fonts, the SVG structure, the PNG image with a fixed height, a maximum width and the aspect ratio of `logo.svg`, each WebP image with its size and aspect ratio (each checked in the header of the file before the image is decoded), and each product with the product rules of the data holder. The process does not start when a rule fails, and one error names each file and rule that fails. The contrast of the colours is not a start-up check. A test of the design system checks the contrast of each preset brand, with the list of pairs of the design system, so that the list has one source.
- **Assets.** The data holder serves the logo, the mark and a stylesheet of the brand on its public base URI, so that the `logoUri` of the brand on the Register can point to the logo. The stylesheet sets the custom properties of the design system with the values of the brand. Each value in it comes from a number, a fixed unit or a fixed font name, so a directory cannot add other CSS. The data holder does not change an SVG file, and does not make a PNG from an SVG: the author of the brand gives both files. Each SVG response has a CSP with `sandbox` and `nosniff`, so a script in a user SVG file does not run, also when a user opens its URL directly. The online bank site serves the photos of the brand from its own origin, and gives the list of the photos that the brand has, so that it never asks for a photo that is not there.

## Consequences

- An operator runs a preset with one setting, and a user brand with one setting and one mount. A new preset is a new directory in the source and a new build.
- A user brand needs no change to the code, but it can use only the fonts and the layouts that the data holder contains. A new layout, or a change of the name of a layout, is a change to the site and of the contract.
- The photos are optional, so a brand with no photo starts, and its site shows a shape in the brand colour in place of each photo. The data holder does not check what a photo shows: a photo with text or a logo is a duty of the author of the brand.
- The photos of the presets make the source and the binary larger, by approximately 0.5 MiB for each preset. A preset directory with its photos still fits in one ConfigMap.
- A mistake in a user directory stops the start, with the file and the rule. The cost is that a directory with an extra file, for example a note, also stops the start.
- A user directory with a low contrast starts. Only the presets have a contrast test. A user brand can thus make the pages of its instance fail WCAG 2.2 AA.
- The pages show the logo on the header colour of the brand, data recipients show it on white, and the pages show the mark on the hero colour. No check proves that the files are legible there, also for a preset. As for the contrast of a user brand, this is a duty of the author of the brand.
- A new token of the design system is a change of the contract. Each brand directory, also each user directory, must then add the token, or the data holder does not start. A test of the data holder compares the token names of the contract with a brand of the design system, so the two do not drift.
- A later file or a later optional field is a change of the contract that the data holder of that time must accept. An older data holder refuses the new directory. `logo.png` was such a change: a directory without it does not start. The required `bsb` of `brand.json`, the required `layout` and the heading font token were also such changes. A BSB is a fact of the bank, as its name is, so it is in the directory and not in the configuration of each deployment.
- The author of a brand keeps two logo files the same. The aspect ratio check finds a PNG of a different shape, but not a PNG with different content.
- `tokens.json` has literal values, so a brand directory repeats values that the design system gives as primitive tokens.
- A user brand in a Kubernetes deployment fits in one ConfigMap, and needs no other change to the loader.
- The data holder does not sanitise SVG files. A page that puts a user SVG inline, and not in `img`, can run its script. The sandbox protects only the URL of the file.
