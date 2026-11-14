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

test("the site uses the font and the primary colour of the brand stylesheet", async ({ page }) => {
  const styles = await page.request.get("/brand/styles.css");
  expect(styles.status()).toBe(200);
  const css = await styles.text();
  const font = /--ds-font-body: ([^;]+);/.exec(css)?.[1];
  const primary = /--ds-color-primary: ([^;]+);/.exec(css)?.[1];
  if (!font || !primary)
    throw new Error("The brand stylesheet has no body font or primary colour.");

  await page.goto("/sign-up");
  await expect(page.locator("body")).toHaveCSS(
    "font-family",
    await computed(page, "fontFamily", font),
  );
  await expect(page.getByRole("button", { name: "Sign up" })).toHaveCSS(
    "background-color",
    await computed(page, "backgroundColor", primary),
  );
});
