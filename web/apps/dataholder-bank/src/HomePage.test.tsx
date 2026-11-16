import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { http, product, renderApp, serveBrand, serveProducts, worker } from "./test/app";

// Verifies: DH-1805
test("the home page names the bank, says what the site has and links to the products", async () => {
  serveProducts([]);

  renderApp("/");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Banking with Example Bank" }))
    .toBeVisible();
  await expect
    .element(
      page.getByText(
        "See the products that Example Bank offers now, with their rates, fees and features.",
      ),
    )
    .toBeVisible();
  await expect
    .element(page.getByRole("main").getByRole("link", { name: "See all products" }))
    .toHaveAttribute("href", "/products");
});

// Verifies: DH-1807, DH-1808
test("the home page links to the group of each category that has a product, with the number of its products", async () => {
  serveProducts([
    product("loan", { productCategory: "RESIDENTIAL_MORTGAGES", name: "Home Loan" }),
    product("everyday"),
    product("saver", { name: "Saver" }),
  ]);

  renderApp("/");

  const categories = page.getByRole("main").getByRole("listitem");
  await expect
    .element(categories.nth(0))
    .toHaveTextContent("Everyday and savings accounts2 products");
  await expect.element(categories.nth(1)).toHaveTextContent("Home loans1 product");
  expect(categories.all()).toHaveLength(2);
  await expect
    .element(page.getByRole("link", { name: "Home loans 1 product" }))
    .toHaveAttribute("href", "/products#residential-mortgages");
});

// Verifies: DH-1809
test("the home page says that the bank shares data under the CDR only with the consent of the consumer", async () => {
  serveProducts([]);

  renderApp("/");

  await expect
    .element(page.getByRole("heading", { level: 2, name: "Sharing your data" }))
    .toBeVisible();
  await expect
    .element(
      page.getByText(
        "Example Bank shares your data under the Consumer Data Right only when you give your consent.",
      ),
    )
    .toBeVisible();
});

test("the home page shows the problem when the products do not load", async () => {
  worker.use(
    http.get("/api/products", ({ response }) =>
      response("default").json(
        {
          type: "about:blank",
          title: "Internal Server Error",
          status: 500,
          detail: "The data holder could not complete the request.",
        },
        { status: 500 },
      ),
    ),
  );

  renderApp("/");

  await expect
    .element(page.getByRole("alert"))
    .toHaveTextContent("Internal Server ErrorThe data holder could not complete the request.");
});

const photos = () => [...document.querySelectorAll<HTMLImageElement>("main img")];

// Verifies: DH-1834
test("the home page has the layout of the brand", async () => {
  serveBrand({ layout: "ledger" });
  serveProducts([]);

  renderApp("/");

  await expect
    .poll(() => document.querySelector<HTMLElement>("[data-layout]")?.dataset.layout)
    .toBe("ledger");
});

// Verifies: DH-1841, DH-1844
test("the home page shows the hero photo and the photo of each category as decoration", async () => {
  serveBrand({ photos: ["hero.webp", "category-residential-mortgages.webp"] });
  serveProducts([product("loan", { productCategory: "RESIDENTIAL_MORTGAGES", name: "Home Loan" })]);

  renderApp("/");

  await expect.element(page.getByRole("link", { name: "Home loans 1 product" })).toBeVisible();
  await expect
    .poll(() => photos().map((img) => [img.getAttribute("src"), img.getAttribute("alt")]))
    .toEqual([
      ["/brand/hero.webp", ""],
      ["/brand/category-residential-mortgages.webp", ""],
    ]);
});

// Verifies: DH-1845
test("a brand with no photos shows a shape in the brand colour in place of each photo, and no image", async () => {
  serveBrand({ photos: [] });
  serveProducts([product("everyday")]);

  renderApp("/");

  await expect
    .element(page.getByRole("link", { name: "Everyday and savings accounts 1 product" }))
    .toBeVisible();
  expect(photos().filter((img) => img.src.endsWith(".webp"))).toEqual([]);
  expect(document.querySelectorAll("main [data-photo-placeholder]")).toHaveLength(2);
});

// Verifies: DH-1840
test("the updraft layout shows the everyday and savings accounts first, then the CDS order", async () => {
  serveBrand({ layout: "updraft" });
  serveProducts([
    product("deposit", { productCategory: "TERM_DEPOSITS", name: "Term Deposit" }),
    product("loan", { productCategory: "RESIDENTIAL_MORTGAGES", name: "Home Loan" }),
    product("everyday"),
  ]);

  renderApp("/");

  const categories = page.getByRole("main").getByRole("listitem");
  await expect
    .element(categories.nth(0))
    .toHaveTextContent("Everyday and savings accounts1 product");
  await expect.element(categories.nth(1)).toHaveTextContent("Term deposits1 product");
  await expect.element(categories.nth(2)).toHaveTextContent("Home loans1 product");
});
