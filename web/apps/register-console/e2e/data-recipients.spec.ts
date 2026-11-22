import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";

// The seed file of the lab (services/register/seed/example.json). The conformance suite does not change its status.
const seeded = "Mock Budgeting Pty Ltd";

test("the operator finds a data recipient of the seed file, and a reload shows the same list", async ({
  page,
}) => {
  await page.goto("/");
  await expect(page).toHaveURL("/data-recipients");

  await page.getByRole("searchbox", { name: "Search by name" }).fill("mock budget");
  await expect(page).toHaveURL("/data-recipients?q=mock+budget");
  await page.getByRole("combobox", { name: "Status" }).selectOption("ACTIVE");

  await expect(page).toHaveURL("/data-recipients?status=ACTIVE&q=mock+budget");
  const table = page.getByRole("table", { name: "Data recipients" });
  await expect(table.getByRole("cell", { name: seeded, exact: true })).toBeVisible();
  await expect(table.getByRole("row")).toHaveCount(2);

  await page.reload();

  await expect(table.getByRole("cell", { name: seeded, exact: true })).toBeVisible();
  await expect(page.getByRole("searchbox", { name: "Search by name" })).toHaveValue("mock budget");
  await expect(page.getByRole("combobox", { name: "Status" })).toHaveValue("ACTIVE");
});

test("the list of data recipients passes the automated accessibility checks", async ({ page }) => {
  await page.goto("/data-recipients");
  await expect(
    page
      .getByRole("table", { name: "Data recipients" })
      .getByRole("cell", { name: seeded, exact: true }),
  ).toBeVisible();

  await expectNoAxeViolations(page);
});

test("the operator sorts the data recipients by the number of software products, in both directions", async ({
  page,
}) => {
  await page.goto("/data-recipients");
  const table = page.getByRole("table", { name: "Data recipients" });
  await expect(table.getByRole("cell", { name: seeded, exact: true })).toBeVisible();
  // A poll waits for the rows of the new sort.
  const sorted = (direction: 1 | -1) => async () => {
    const counts = (await table.locator("tbody tr td:last-child").allTextContents()).map(Number);
    return (
      counts.length > 0 && counts.every((n, i) => i === 0 || (n - counts[i - 1]!) * direction >= 0)
    );
  };

  await table.getByRole("button", { name: "Software products" }).click();
  await expect(page).toHaveURL("/data-recipients?sort=softwareProductCount&order=asc");
  await expect(table.getByRole("columnheader", { name: "Software products" })).toHaveAttribute(
    "aria-sort",
    "ascending",
  );
  await expect.poll(sorted(1)).toBe(true);

  await table.getByRole("button", { name: "Software products" }).click();
  await expect(page).toHaveURL("/data-recipients?sort=softwareProductCount&order=desc");
  await expect(table.getByRole("columnheader", { name: "Software products" })).toHaveAttribute(
    "aria-sort",
    "descending",
  );
  await expect.poll(sorted(-1)).toBe(true);
});
