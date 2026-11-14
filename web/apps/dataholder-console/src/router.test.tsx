import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { http, renderApp, worker } from "./test/app";

function serveNoCustomers() {
  worker.use(
    http.get("/api/customers", ({ response }) => response(200).json({ data: [], totalRecords: 0 })),
  );
}

test("the header gives the logo of the brand and the label, and the tab title gives the brand name", async () => {
  serveNoCustomers();

  renderApp("/customers");

  const banner = page.getByRole("banner");
  await expect
    .element(banner.getByRole("img", { name: "Example Bank" }))
    .toHaveAttribute("src", "/brand/logo.svg");
  await expect.element(banner).toHaveTextContent("Operator console");
  await expect.poll(() => document.title).toBe("Example Bank");
});

test("the header says Data holder when the data holder gives no brand name", async () => {
  serveNoCustomers();
  worker.use(http.untyped.get("/api/brand", () => Response.error()));

  renderApp("/customers");

  await expect.element(page.getByText("No customers match")).toBeVisible();
  await expect
    .element(page.getByRole("banner").getByRole("img", { name: "Data holder" }))
    .toBeVisible();
});

test("the item of the side navigation is current, and assistive technology does not announce its icon", async () => {
  serveNoCustomers();

  renderApp("/customers");
  const navigation = page.getByRole("navigation", { name: "Main" });
  const link = navigation.getByRole("link", { name: "Customers", exact: true });

  await expect.element(link).toHaveAttribute("aria-current", "page");
  expect(link.element().querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
  await expectNoAxeViolations(navigation.element());
});
