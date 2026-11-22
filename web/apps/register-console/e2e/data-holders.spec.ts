import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";

// The seed brand "Mock Mutual" is not in this legal entity name, so a match shows that the search includes brand names.
const seeded = "Mock Holdings Ltd";

test("the operator finds a data holder of the seed file by a brand name, and a reload shows the same list", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByRole("link", { name: "Data holders" }).click();
  await expect(page).toHaveURL("/data-holders");

  const searchbox = page.getByRole("searchbox", { name: "Search by legal entity or brand name" });
  await searchbox.fill("mutual");
  await expect(page).toHaveURL("/data-holders?q=mutual");
  await page.getByRole("combobox", { name: "Status" }).selectOption("ACTIVE");

  await expect(page).toHaveURL("/data-holders?status=ACTIVE&q=mutual");
  const table = page.getByRole("table", { name: "Data holders" });
  await expect(table.getByRole("cell", { name: seeded, exact: true })).toBeVisible();
  await expect(table.getByRole("row")).toHaveCount(2);

  await page.reload();

  await expect(table.getByRole("cell", { name: seeded, exact: true })).toBeVisible();
  await expect(searchbox).toHaveValue("mutual");
  await expect(page.getByRole("combobox", { name: "Status" })).toHaveValue("ACTIVE");
});

test("the list of data holders passes the automated accessibility checks", async ({ page }) => {
  await page.goto("/data-holders");
  await expect(
    page
      .getByRole("table", { name: "Data holders" })
      .getByRole("cell", { name: seeded, exact: true }),
  ).toBeVisible();

  await expectNoAxeViolations(page);
});
