import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page } from "vitest/browser";

import {
  account,
  http,
  narrowWidth,
  presets,
  type Preset,
  product,
  renderApp,
  serveAccounts,
  serveProducts,
  serveSession,
  transaction,
  showPreset,
  worker,
} from "./test/app";
import { categoryLabels, groupByCategory } from "./products/categories";

// Each page shows the products of the preset before the checks, so that the checks do not see the loading state.
const pages = [
  {
    url: "/",
    title: "Home",
    loaded: (category: string) =>
      page
        .getByRole("main")
        .getByRole("link", { name: new RegExp(`^${category} \\d+ products?$`) }),
  },
  {
    url: "/products",
    title: "Products",
    loaded: (category: string) => page.getByRole("heading", { level: 2, name: category }),
  },
  {
    url: "/unknown",
    title: "Page not found",
    loaded: () => page.getByRole("heading", { level: 1, name: "Page not found" }),
  },
  {
    url: "/sign-up",
    title: "Sign up",
    loaded: () => page.getByRole("heading", { level: 1, name: /^Become a customer of / }),
  },
  {
    url: "/sign-in",
    title: "Sign in",
    loaded: () => page.getByRole("heading", { level: 1, name: "Sign in to online banking" }),
  },
  {
    url: "/accounts",
    title: "Accounts",
    setUp: () =>
      serveAccounts([account("account-1"), account("account-2", { openStatus: "CLOSED" })]),
    loaded: () => page.getByRole("heading", { level: 2, name: "Closed accounts" }),
  },
  {
    url: "/accounts/account-1",
    title: "Everyday",
    setUp: () => serveAccounts([account("account-1")], { "account-1": [transaction("t-1")] }),
    loaded: () => page.getByText("Corner Cafe"),
  },
  {
    url: "/customer-number",
    title: "Your customer number",
    setUp: () => serveSession(),
    loaded: () => page.getByRole("button", { name: "Copy customer number" }),
  },
];

function firstCategory(preset: Preset): string {
  const [group] = groupByCategory(preset.products);
  if (!group) throw new Error(`the preset ${preset.name} has no products`);
  return categoryLabels[group.category];
}

test("there are preset brands", () => {
  expect(presets.length).toBeGreaterThan(0);
});

describe.each(presets)("with the preset $name", (preset) => {
  test.each(pages)(
    "the $title page passes the accessibility checks and has no horizontal scroll at 320 CSS px",
    async ({ url, title, setUp, loaded }) => {
      setUp?.();
      await showPreset(preset, url);

      await expect.element(loaded(firstCategory(preset))).toBeVisible();
      await expect.poll(() => document.title).toBe(`${title} - ${preset.brandName} online banking`);
      expect(window.innerWidth).toBe(narrowWidth);
      expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
      await expectNoAxeViolations(document.documentElement);
    },
  );
});

test.each(pages)(
  "the $title page shows the logo of the brand, the link to the products, and the footer with its two statements",
  async ({ url, setUp }) => {
    setUp?.();
    serveProducts([product("everyday")]);

    renderApp(url);

    const banner = page.getByRole("banner");
    await expect
      .element(banner.getByRole("img", { name: "Example Bank" }))
      .toHaveAttribute("src", "/brand/logo.svg");
    await expect
      .element(
        page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Products" }),
      )
      .toHaveAttribute("href", "/products");
    await expect
      .element(page.getByRole("contentinfo"))
      .toHaveTextContent(
        "Third parties never need your password to get your data under the Consumer Data Right." +
          "Example Bank is not a real bank and this is just a demonstration.",
      );
  },
);

test("the pages say Bank when the data holder gives no brand", async () => {
  serveProducts([]);
  worker.use(http.untyped.get("/api/brand", () => Response.error()));

  renderApp("/");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Banking with Bank" }))
    .toBeVisible();
  await expect.element(page.getByRole("banner").getByRole("img", { name: "Bank" })).toBeVisible();
  await expect.poll(() => document.title).toBe("Home - Bank online banking");
});

test("the not found page links to the products", async () => {
  renderApp("/unknown");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Page not found" }))
    .toBeVisible();
  await expect
    .element(page.getByRole("main").getByRole("link", { name: "See all products" }))
    .toHaveAttribute("href", "/products");
});

const leftEdge = (selector: string) =>
  document.querySelector(selector)?.getBoundingClientRect().left;
const rightEdge = (selector: string) =>
  document.querySelector(selector)?.getBoundingClientRect().right;

test("the lock note and the statements of the footer line up with the content of the main area on a wide screen", async () => {
  await page.viewport(1280, 900);
  serveProducts([product("everyday")]);

  renderApp("/");

  await expect.poll(() => document.querySelector("footer > p")?.textContent).toMatch(/real bank/);
  expect(leftEdge("footer > .cp-note")).toBe(leftEdge(".home"));
  expect(leftEdge("footer > p")).toBe(leftEdge(".home"));
  expect(rightEdge("footer > .cp-note")).toBe(rightEdge(".home"));
  // The text of the note starts next to its icon, not in the middle of the wide box.
  const noteText = document.querySelector("footer > .cp-note p")?.getBoundingClientRect().left ?? 0;
  expect(noteText - (leftEdge("footer > .cp-note") ?? 0)).toBeLessThan(64);
});
