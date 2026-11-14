import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";

// A registration needs an SSA of the Register and a signed request, so this test uses the registrations that the
// conformance suite makes (mise run conformance).
test("the operator opens the list of the client registrations and the detail of a registration", async ({
  page,
}) => {
  await page.goto("/customers");
  await page
    .getByRole("navigation", { name: "Main" })
    .getByRole("link", { name: "Client registrations" })
    .click();

  await expect(page).toHaveURL("/registrations");
  const table = page.getByRole("table", { name: "Client registrations" });
  // With no registrations, the first row has the empty state.
  await expect(table.getByRole("row").nth(1)).toBeVisible();
  test.skip(
    await page.getByText("No client registrations").isVisible(),
    "The lab has no registrations. Run mise run conformance first.",
  );
  await expectNoAxeViolations(page);

  // The client with the most arrangements is first, so that the detail shows arrangements when the lab has them.
  await page.getByRole("button", { name: "All arrangements", exact: true }).click();
  await page.getByRole("button", { name: "All arrangements", exact: true }).click();
  await expect(page).toHaveURL("/registrations?sort=arrangementCount&order=desc");
  await expect(page.getByRole("columnheader", { name: "All arrangements" })).toHaveAttribute(
    "aria-sort",
    "descending",
  );

  // The table keeps the rows of the previous sort until the response of the new sort arrives.
  await expect
    .poll(async () => {
      const counts = (await table.locator("tbody tr td:last-child").allTextContents()).map(Number);
      return counts.every((count, i) => i === 0 || counts[i - 1]! >= count);
    })
    .toBe(true);

  const first = table.getByRole("row").nth(1);
  const arrangementCount = Number(await first.getByRole("cell").last().textContent());
  const link = first.getByRole("link");
  const clientId = await link.textContent();
  await link.click();

  await expect(page).toHaveURL(`/registrations/${clientId}`);
  const arrangements = page.getByRole("table", { name: "Arrangements" });
  if (arrangementCount === 0) {
    await expect(arrangements.getByText("No arrangements on this page")).toBeVisible();
  } else {
    // The first page has 25 arrangements or less, and each one links to its customer.
    await expect(arrangements.getByRole("link")).toHaveCount(Math.min(arrangementCount, 25));
    const records = `${arrangementCount} ${arrangementCount === 1 ? "record" : "records"}`;
    await expect(page.getByText(records, { exact: true })).toBeVisible();
  }
  const metadata = page.getByRole("region", { name: "Metadata", exact: true });
  await expect(metadata.getByRole("button", { name: "Copy recipient base URI" })).toBeVisible();
  await expect(metadata.getByRole("button", { name: "Copy JWKS URI" })).toBeVisible();
  await expect(metadata.getByRole("button", { name: "Copy redirect URI" }).first()).toBeVisible();
  await expect(metadata.getByRole("link")).toHaveCount(0);
  await expectNoAxeViolations(page);
});

test("the detail of a registration that does not exist says not found", async ({ page }) => {
  await page.goto("/registrations/00000000-0000-4000-8000-000000000000");

  await expect(page.getByRole("alert")).toContainText(
    "The data holder has no registration with this client ID.",
  );
});
