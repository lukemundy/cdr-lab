import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { http, product, renderApp, serveProducts, worker } from "../test/app";

// Verifies: DH-1812, DH-1814
test("the product list shows the name and the description of each product in the group of its category", async () => {
  serveProducts([
    product("loan", {
      productCategory: "RESIDENTIAL_MORTGAGES",
      name: "Home Loan",
      description: "A loan for a home.",
    }),
    product("everyday", { description: "An account for each day." }),
  ]);

  renderApp("/products");

  const groups = page.getByRole("region");
  await expect
    .element(groups.nth(0).getByRole("heading", { level: 2 }))
    .toHaveTextContent("Everyday and savings accounts");
  await expect
    .element(groups.nth(0).getByRole("heading", { level: 3, name: "Everyday Account" }))
    .toBeVisible();
  await expect.element(groups.nth(0).getByText("An account for each day.")).toBeVisible();
  await expect
    .element(groups.nth(1).getByRole("heading", { level: 2, name: "Home loans" }))
    .toBeVisible();
  await expect.element(groups.nth(1).getByText("A loan for a home.")).toBeVisible();
});

test("the group of a category has the ID of the link of the home page", async () => {
  serveProducts([product("loan", { productCategory: "RESIDENTIAL_MORTGAGES", name: "Home Loan" })]);

  renderApp("/products#residential-mortgages");

  await expect.element(page.getByRole("region", { name: "Home loans" })).toBeVisible();
  expect(document.getElementById("residential-mortgages")?.textContent).toBe("Home loans");
});

test("the product list says when the bank has no products", async () => {
  serveProducts([]);

  renderApp("/products");

  await expect.element(page.getByText("The bank has no products now.")).toBeVisible();
});

// Verifies: DH-1814
test("each product links to its detail", async () => {
  serveProducts([product("everyday")]);

  renderApp("/products");

  await expect
    .element(page.getByRole("link", { name: "Everyday Account" }))
    .toHaveAttribute("href", "/products/everyday");
});

// Verifies: DH-2640
test("the product list shows a text for a customer when the site API cannot answer", async () => {
  worker.use(http.untyped.get("/api/products", () => Response.error()));

  renderApp("/products");

  const alert = page.getByRole("alert");
  await expect.element(alert.getByText("We cannot show this information now")).toBeVisible();
  await expect.element(alert.getByText("Try again later.")).toBeVisible();
});
