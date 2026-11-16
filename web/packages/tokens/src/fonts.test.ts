import { beforeAll, expect, test } from "vitest";
import fontsURL from "../dist/fonts.css?url";
import tokensURL from "../dist/tokens.css?url";
import { fonts } from "./fonts.ts";

beforeAll(async () => {
  await Promise.all(
    [tokensURL, fontsURL].map(
      (href) =>
        new Promise((resolve, reject) => {
          const link = document.createElement("link");
          link.rel = "stylesheet";
          link.href = href;
          link.addEventListener("load", resolve);
          link.addEventListener("error", reject);
          document.head.append(link);
        }),
    ),
  );
});

function width(text: string, font: string, numeric = "normal") {
  const span = document.createElement("span");
  span.textContent = text;
  span.style.font = font;
  span.style.fontVariantNumeric = numeric;
  document.body.append(span);
  const w = span.getBoundingClientRect().width;
  span.remove();
  return w;
}

test.each(fonts.map(({ family }) => family))(
  "the %s font loads from this origin",
  async (family) => {
    const faces = await document.fonts.load(`16px "${family}"`, "0123456789 ĀĒĪŌŪ");
    expect(faces.length).toBeGreaterThan(0);
    for (const face of faces) expect(face.status).toBe("loaded");
  },
);

// Verifies: UI-0046
test("the body font has tabular figures", async () => {
  await document.fonts.load('16px "Inter Variable"');
  const font = "16px var(--ds-font-body)";
  expect(width("1111", font)).not.toBeCloseTo(width("0000", font), 1);
  expect(width("1111", font, "tabular-nums")).toBeCloseTo(width("0000", font, "tabular-nums"), 1);
});
