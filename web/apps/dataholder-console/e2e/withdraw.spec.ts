import { expect, test } from "@playwright/test";
import createClient from "openapi-fetch";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import type { paths } from "../src/api/schema.gen";
import { requiredEnv } from "@cdr-lab/console/e2e/dataholder-lab";

const api = createClient<paths>({ baseUrl: requiredEnv("DATAHOLDER_INTERNAL_URL") });

// An arrangement needs the full consent flow of a registered client, so this test withdraws an arrangement that the
// conformance suite makes (mise run conformance). The conformance suite changed its customers last.
async function arrangementToWithdraw() {
  const { data: customers } = await api.GET("/api/customers", {
    params: { query: { sort: "lastUpdated", order: "desc", pageSize: 100 } },
  });
  const withArrangements = (customers?.data ?? []).filter((c) => c.currentArrangementCount > 0);
  const details = await Promise.all(
    withArrangements.map(({ customerNumber }) =>
      api.GET("/api/customers/{customerNumber}", { params: { path: { customerNumber } } }),
    ),
  );

  for (const { data: detail } of details) {
    const arrangement = detail?.arrangements.find((a) => a.actions.length > 0);
    if (detail && arrangement) return { customerNumber: detail.customerNumber, ...arrangement };
  }
  return undefined;
}

test("the operator withdraws a consent for a customer", async ({ page }) => {
  const found = await arrangementToWithdraw();
  test.skip(!found, "The lab has no current arrangement. Run mise run conformance first.");
  if (!found) return;
  const { customerNumber, arrangementId, clientName } = found;

  await page.goto(`/customers/${customerNumber}`);
  await page.getByRole("button", { name: "Show all arrangements" }).click();
  const card = page.getByRole("region", { name: clientName, exact: true }).filter({
    hasText: arrangementId,
  });
  await card.getByRole("button", { name: "Withdraw", exact: true }).click();

  const dialog = page.getByRole("alertdialog", { name: `Withdraw the consent of ${clientName}?` });
  await expect(dialog).toContainText(`The arrangement ${arrangementId} of ${clientName} ends.`);
  await expect(dialog).toContainText("The data holder sends a notice to the data recipient.");
  await expect(dialog).toContainText("You cannot reverse this action.");
  await expectNoAxeViolations(page);
  await dialog.getByRole("button", { name: "Withdraw", exact: true }).click();

  await expect(dialog).toBeHidden();
  await expect(
    page.getByRole("dialog", { name: `The arrangement of ${clientName} is now ENDED` }),
  ).toBeVisible();
  await expect(card.getByText("ENDED", { exact: true })).toBeVisible();
  await expect(card.getByText("WITHDRAWN_BY_CONSUMER", { exact: true })).toBeVisible();
  await expect(card.getByText("ALTERNATIVE_METHOD", { exact: true })).toBeVisible();
  await expect(card.getByText("Notice status")).toBeVisible();
  await expect(card.getByRole("button", { name: "Withdraw", exact: true })).toHaveCount(0);
});
