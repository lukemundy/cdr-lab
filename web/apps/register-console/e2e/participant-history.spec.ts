import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/app/e2e/checks";

// The seed file of the lab (services/register/seed/example.json).
const recipient = { name: "Mock Budgeting Pty Ltd", id: "c7b88455-4a14-479e-b53e-fbc4000716b3" };

// Other tests can add events to the stream, so the test checks only the import, which is always the first event.
test("the operator opens the history of a data recipient of the seed file from its detail", async ({
  page,
}) => {
  await page.goto(`/data-recipients/${recipient.id}`);
  await page.getByRole("link", { name: "History", exact: true }).click();

  await expect(page).toHaveURL(`/data-recipients/${recipient.id}/history`);
  await expect(
    page.getByRole("heading", { level: 1, name: `History of ${recipient.name}` }),
  ).toBeVisible();
  const history = page.getByRole("list", { name: "History" });
  const imported = history.getByRole("listitem").filter({ hasText: "DataRecipientImported" });
  await expect(imported.getByText("Loaded into the Register")).toBeVisible();
  await expect(imported.locator("time")).toContainText(
    /^\d{1,2} \w+ \d{4}, \d{1,2}:\d{2}:\d{2} [ap]m UTC\+8/,
  );

  await page.reload();
  await expect(imported.getByText("Loaded into the Register")).toBeVisible();

  await expectNoAxeViolations(page);
});
