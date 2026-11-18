import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";

// Mock Budget Beta in services/register/seed/example.json. A REMOVED software product cannot change its status, so
// the conformance suite and the other tests do not change it.
const removedProduct = "6d0b646f-475d-4c7b-af85-0eb558f50d47";

// Verifies: DH-1436, DH-1438
test("the operator sees the Register status of each software product, and the time of the last response", async ({
  page,
}) => {
  await page.goto("/customers");
  await page
    .getByRole("navigation", { name: "Main" })
    .getByRole("link", { name: "Register statuses" })
    .click();

  await expect(page).toHaveURL("/register-statuses");
  const row = page
    .getByRole("table", { name: "Register statuses" })
    .getByRole("row")
    .filter({ hasText: removedProduct });
  await expect(row.getByText("REMOVED", { exact: true })).toBeVisible();
  // The data holder is ready only after it got the statuses, so the time is always there.
  const receivedAt = page.locator("dd", { has: page.locator("time") });
  await expect(receivedAt).toBeVisible();
  await expectNoAxeViolations(page);
});

// Verifies: DH-1440
test("the operator filters the statuses by the registration", async ({ page }) => {
  await page.goto("/register-statuses");
  const table = page.getByRole("table", { name: "Register statuses" });
  const registration = page.getByRole("combobox", { name: "Registration" });

  await registration.selectOption({ label: "Not registered" });

  await expect(page).toHaveURL("/register-statuses?registration=notRegistered");
  const removed = table.getByRole("row").filter({ hasText: removedProduct });
  await expect(removed.getByText("Not registered", { exact: true })).toBeVisible();
  await expect(removed.getByRole("link")).toHaveCount(0);

  await registration.selectOption({ label: "Registered" });

  await expect(page).toHaveURL("/register-statuses?registration=registered");
  await expect(table.getByText(removedProduct)).toHaveCount(0);
  // While the table loads, the first row has the loading state. With no registrations, it then has the empty state.
  await expect(table).not.toHaveAttribute("aria-busy");
  await expect(table.getByRole("row").nth(1)).toBeVisible();
  // A registration needs an SSA of the Register and a signed request, so this part uses the registrations that the
  // conformance suite makes (mise run conformance).
  test.skip(
    await table.getByText("No software products match").isVisible(),
    "The lab has no registrations. Run mise run conformance first.",
  );
  await expectNoAxeViolations(page);

  const link = table.getByRole("row").nth(1).getByRole("link");
  const name = await link.textContent();
  await link.click();

  await expect(page).toHaveURL(/\/registrations\/[^/]+$/u);
  await expect(page.getByRole("heading", { level: 1 })).toContainText(name ?? "");
});
