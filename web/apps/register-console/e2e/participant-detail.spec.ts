import { expect, test, type Locator } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import { showAllBrands } from "./lab";

// The seed file of the lab (services/register/seed/example.json).
const recipient = { name: "Mock Budgeting Pty Ltd", id: "c7b88455-4a14-479e-b53e-fbc4000716b3" };
const holder = { name: "Mock Holdings Ltd", id: "a81975c6-1ac9-4fe4-88fc-68f50efd27cb" };

const definition = (region: Locator, term: string) =>
  region
    .locator("dt", { hasText: new RegExp(`^${term}$`) })
    .locator("xpath=following-sibling::dd[1]");

test("the operator opens a data recipient of the seed file from the list", async ({ page }) => {
  const origins = new Set<string>();
  page.on("request", (request) => origins.add(new URL(request.url()).origin));
  await page.goto("/data-recipients");
  const consoleOrigin = new URL(page.url()).origin;
  await page
    .getByRole("table", { name: "Data recipients" })
    .getByRole("link", { name: recipient.name, exact: true })
    .click();

  await expect(page).toHaveURL(`/data-recipients/${recipient.id}`);
  await expect(page.getByRole("heading", { level: 1, name: recipient.name })).toBeVisible();
  await showAllBrands(page);
  const entity = page.getByRole("region", { name: "Legal entity", exact: true });
  await expect(definition(entity, "Accreditation number")).toHaveText("ADR-MOCK-0001");
  await expect(definition(entity, "Logo URI")).toHaveText(
    "https://mock-budgeting.example/logo.png",
  );
  const product = page.getByRole("region", { name: "Mock Budget Web", exact: true });
  await expect(definition(product, "Sector identifier URI")).toHaveText("Not set");
  await expect(page.locator("img")).toHaveCount(0);
  expect([...origins]).toEqual([consoleOrigin]);

  await expectNoAxeViolations(page);
});

test("the operator opens a data holder of the seed file from the list", async ({ page }) => {
  await page.goto("/data-holders");
  await page
    .getByRole("table", { name: "Data holders" })
    .getByRole("link", { name: holder.name, exact: true })
    .click();

  await expect(page).toHaveURL(`/data-holders/${holder.id}`);
  await expect(page.getByRole("heading", { level: 1, name: holder.name })).toBeVisible();
  await showAllBrands(page);
  const energy = page.getByRole("region", { name: "Mock Energy", exact: true });
  await expect(definition(energy, "Resource base URI")).toHaveText(
    "https://mtls.mock-energy.example",
  );
  const mutual = page.getByRole("region", { name: "Mock Mutual", exact: true });
  await expect(definition(mutual, "Interim ID")).toHaveText("d9530e60-cdc4-408c-a6f4-aba39e4158b6");
  await Promise.all(
    ["Brand ID", "Resource base URI", "Infosec base URI", "JWKS endpoint"].map((term) =>
      expect(definition(mutual, term)).toHaveText("Not set"),
    ),
  );

  await expectNoAxeViolations(page);
});

test("an unknown legal entity shows the problem from the Register", async ({ page }) => {
  await page.goto("/data-recipients/an-unknown-legal-entity");

  const alert = page.getByRole("alert");
  await expect(alert.getByText("Not Found")).toBeVisible();
  await expect(
    alert.getByText("The Register has no data recipient with this legal entity ID."),
  ).toBeVisible();
});
