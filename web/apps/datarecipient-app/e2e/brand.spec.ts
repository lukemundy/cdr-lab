import { expect, test, type Page } from "@playwright/test";

// The browser normalises a CSS value, so the test compares the computed value of a probe element.
async function computed(page: Page, property: "backgroundColor" | "fontFamily", value: string) {
  return page.evaluate(
    ([name, raw]) => {
      const probe = document.createElement("div");
      probe.style[name] = raw;
      document.body.append(probe);
      const result = getComputedStyle(probe)[name];
      probe.remove();
      return result;
    },
    [property, value] as const,
  );
}

test("the web app uses the page colour, the font and the mark of the brand", async ({ page }) => {
  const styles = await page.request.get("/brand/styles.css");
  expect(styles.status()).toBe(200);
  const css = await styles.text();
  const font = /--ds-font-body: ([^;]+);/.exec(css)?.[1];
  const background = /--ds-color-page: ([^;]+);/.exec(css)?.[1];
  if (!font || !background)
    throw new Error("The brand stylesheet has no body font or page colour.");

  await page.goto("/");

  const body = page.locator("body");
  await expect(body).toHaveCSS("font-family", await computed(page, "fontFamily", font));
  await expect(body).toHaveCSS(
    "background-color",
    await computed(page, "backgroundColor", background),
  );
  await expect(page.locator('link[rel="icon"]')).toHaveAttribute("href", "/brand/mark.svg");
  const mark = await page.request.get("/brand/mark.svg");
  expect(mark.headers()["content-type"]).toBe("image/svg+xml");
});
