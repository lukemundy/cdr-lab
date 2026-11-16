import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { http, renderApp, worker } from "./test/app";

// Verifies: REG-1051
test("the header says CDR Register", async () => {
  worker.use(
    http.get("/api/data-recipients", ({ response }) =>
      response(200).json({ data: [], totalRecords: 0 }),
    ),
  );

  renderApp("/data-recipients");

  await expect.element(page.getByRole("banner")).toHaveTextContent("CDR Register");
  await expect.element(page.getByText("No data recipients match")).toBeVisible();
});

test("each item of the side navigation has an icon that assistive technology does not announce", async () => {
  worker.use(
    http.get("/api/data-recipients", ({ response }) =>
      response(200).json({ data: [], totalRecords: 0 }),
    ),
  );

  renderApp("/data-recipients");
  const navigation = page.getByRole("navigation", { name: "Main" });

  await expect.element(page.getByText("No data recipients match")).toBeVisible();

  // The exact accessible name shows that the icon adds no text to the name.
  for (const name of ["Data recipients", "Data holders", "Signing keys", "CDR certificates"]) {
    const link = navigation.getByRole("link", { name, exact: true }).element();
    expect(link.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
  }
  await expectNoAxeViolations(navigation.element());
});

test.for([
  "/create/data-recipient",
  "/data-recipients/dr-1/change",
  "/data-recipients/dr-1/add-brand",
  "/data-recipients/dr-1/brands/b-1/change",
  "/data-recipients/dr-1/brands/b-1/add-software-product",
  "/data-recipients/dr-1/software-products/sp-1/change",
])("the side navigation marks Data recipients as the section of the form page %s", async (url) => {
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) =>
      response(200).json({
        legalEntityId: "dr-1",
        legalEntityName: "Alpha Pty Ltd",
        accreditationNumber: "ADR-1",
        accreditationLevel: "UNRESTRICTED",
        logoUri: "https://alpha.example/logo.png",
        status: "ACTIVE",
        lastUpdated: "2026-10-24T02:05:07Z",
        version: 1,
        actions: [],
        dataRecipientBrands: [],
      }),
    ),
  );

  renderApp(url);
  const navigation = page.getByRole("navigation", { name: "Main" });

  await expect
    .element(navigation.getByRole("link", { name: "Data recipients" }))
    .toHaveAttribute("aria-current", "page");
  await expect
    .element(navigation.getByRole("link", { name: "Data holders" }))
    .not.toHaveAttribute("aria-current");
  await expect
    .element(navigation.getByRole("link", { name: "Signing keys" }))
    .not.toHaveAttribute("aria-current");
});

test.for([
  "/create/data-holder",
  "/data-holders/dh-1/change",
  "/data-holders/dh-1/add-brand",
  "/data-holders/dh-1/brands/dh-1-b1/change",
])("the side navigation marks Data holders as the section of the form page %s", async (url) => {
  worker.use(
    http.get("/api/data-holders/{legalEntityId}", ({ response }) =>
      response(200).json({
        legalEntityId: "dh-1",
        legalEntityName: "Delta Ltd",
        logoUri: "https://delta.example/logo.png",
        status: "ACTIVE",
        lastUpdated: "2026-10-24T02:05:07Z",
        version: 1,
        actions: [],
        dataHolderBrands: [],
      }),
    ),
  );

  renderApp(url);
  const navigation = page.getByRole("navigation", { name: "Main" });

  await expect
    .element(navigation.getByRole("link", { name: "Data holders" }))
    .toHaveAttribute("aria-current", "page");
  await expect
    .element(navigation.getByRole("link", { name: "Data recipients" }))
    .not.toHaveAttribute("aria-current");
});
