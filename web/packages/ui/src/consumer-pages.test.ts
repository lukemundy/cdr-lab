import { beforeAll, describe, expect, test } from "vitest";
import { page as browserPage } from "vitest/browser";
import consumerURL from "../dist/consumer-pages/consumer.css?url";
import { expectNoAxeViolations } from "./test/checks.ts";

// The golden files of the data holder: each consumer page, and the brand stylesheet of each preset brand
// (services/dataholder/internal/adapter/consumerpages/golden_test.go).
const pages = import.meta.glob<string>(
  "../../../../services/dataholder/internal/adapter/consumerpages/testdata/pages/*.html",
  { query: "?raw", import: "default", eager: true },
);
const brandStyles = import.meta.glob<string>(
  "../../../../services/dataholder/internal/adapter/consumerpages/testdata/brands/*.css",
  { query: "?url", import: "default", eager: true },
);
const images = import.meta.glob<string>(
  "../../../../services/dataholder/brands/*/{logo,mark}.svg",
  {
    query: "?url",
    import: "default",
    eager: true,
  },
);

// DESIGN.md (Layout): a consumer UI reflows with no horizontal scroll at 320 CSS px.
const narrowWidth = 320;

const baseName = (path: string) =>
  path
    .split("/")
    .at(-1)
    ?.replace(/\.\w+$/, "") ?? path;

const presets = Object.entries(brandStyles).map(([path, styles]) => {
  const name = baseName(path);
  const image = (file: string) => {
    const url = images[`../../../../services/dataholder/brands/${name}/${file}`];
    if (!url) throw new Error(`the preset ${name} has no ${file}`);
    return url;
  };
  return { name, styles, logo: image("logo.svg"), mark: image("mark.svg") };
});

const goldenPages = Object.entries(pages).map(([path, html]) => ({ name: baseName(path), html }));

function stylesheet(href: string) {
  return new Promise((resolve, reject) => {
    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = href;
    link.addEventListener("load", resolve);
    link.addEventListener("error", reject);
    document.head.append(link);
  });
}

// show puts the page in the document, as the data holder serves it, with the URLs of the test server.
async function show(html: string, preset: (typeof presets)[number]) {
  const golden = new DOMParser().parseFromString(html, "text/html");
  for (const link of document.head.querySelectorAll("link[rel=stylesheet]")) link.remove();
  document.documentElement.lang = golden.documentElement.lang;
  document.title = golden.title;
  document.body.replaceWith(document.adoptNode(golden.body));
  for (const image of document.querySelectorAll<HTMLImageElement>('img[src^="/brand/"]')) {
    image.src = image.getAttribute("src") === "/brand/mark.svg" ? preset.mark : preset.logo;
  }
  await Promise.all([stylesheet(consumerURL), stylesheet(preset.styles)]);
  await Promise.all([...document.images].map((image) => image.decode()));
  await document.fonts.ready;
}

beforeAll(async () => {
  await browserPage.viewport(narrowWidth, 640);
});

test("there are golden pages and preset brands", () => {
  expect(goldenPages.length).toBeGreaterThan(0);
  expect(presets.length).toBeGreaterThan(0);
});

describe.each(presets)("preset $name", (preset) => {
  test.each(goldenPages)("the $name page passes the accessibility checks", async ({ html }) => {
    await show(html, preset);

    await expectNoAxeViolations(document.documentElement);
  });

  test.each(goldenPages)(
    "the $name page has no horizontal scroll at 320 CSS px",
    async ({ html }) => {
      await show(html, preset);

      expect(window.innerWidth).toBe(narrowWidth);
      expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
    },
  );

  test("the pages use the font of the brand", async () => {
    await show(goldenPages[0]?.html ?? "", preset);

    const [family = ""] = getComputedStyle(document.body).fontFamily.split(",");
    const faces = await document.fonts.load(`16px ${family}`);
    expect(faces.length).toBeGreaterThan(0);
    for (const face of faces) expect(face.status).toBe("loaded");
  });
});
