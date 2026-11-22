import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/app/e2e/checks";
import { requiredEnv } from "@cdr-lab/app/e2e/dataholder-lab";

type ConsoleCustomer = { customerNumber: string; currentArrangementCount: number };
type ConsoleArrangement = { arrangementId: string; clientName: string; actions: unknown[] };

async function consoleGet<T>(path: string): Promise<T> {
  const response = await fetch(new URL(path, requiredEnv("DATAHOLDER_INTERNAL_URL")));
  if (!response.ok) throw new Error(`The console API gave ${response.status} for ${path}`);
  // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the console OpenAPI document gives the shape, and the test fails when the site does not show it.
  return (await response.json()) as T;
}

// A consent needs the full consent flow of a registered software product, so this test withdraws a consent that the
// conformance suite makes, and it skips in CI, where the lab has no conformance data.
async function consentToWithdraw() {
  const customers = await consoleGet<{ data: ConsoleCustomer[] }>(
    "/api/customers?sort=lastUpdated&order=desc&pageSize=100",
  );
  const withConsents = customers.data.filter((c) => c.currentArrangementCount > 0);
  const details = await Promise.all(
    withConsents.map(({ customerNumber }) =>
      consoleGet<{ customerNumber: string; arrangements: ConsoleArrangement[] }>(
        `/api/customers/${customerNumber}`,
      ),
    ),
  );
  for (const { customerNumber, arrangements } of details) {
    const arrangement = arrangements.find((a) => a.actions.length > 0);
    if (arrangement) return { customerNumber, ...arrangement };
  }
  return undefined;
}

test("a customer stops the sharing of a consent on Data sharing, and sees it in the ended consents", async ({
  page,
}) => {
  const found = await consentToWithdraw();
  test.skip(!found, "The lab has no current consent. Run mise run conformance first.");
  if (!found) return;
  const { customerNumber, clientName } = found;

  await page.goto("/sign-in");
  await page.getByLabel("Email address or customer number").fill(customerNumber);
  await page.getByLabel("Password").fill("any password at all");
  await page.getByRole("button", { name: "Sign in" }).click();
  await page
    .getByRole("navigation", { name: "Main" })
    .getByRole("link", { name: "Data sharing" })
    .click();

  const current = page.getByRole("region", { name: "Current consents" });
  // A customer of the conformance suite can have more than one consent of one app.
  const card = current.getByRole("article", { name: clientName, exact: true }).first();
  await card.getByRole("button", { name: "Stop sharing" }).click();

  const dialog = page.getByRole("alertdialog", { name: `Stop sharing with ${clientName}?` });
  await expect(dialog).toContainText("before you stop sharing to understand the consequences.");
  await expect(dialog).toContainText(
    "CDR data is either deleted or de-identified when it is no longer required.",
  );
  await expectNoAxeViolations(page);
  await dialog.getByRole("button", { name: "Stop sharing" }).click();

  await expect(dialog).toBeHidden();
  await expect(page.getByText(`You stopped sharing your data with ${clientName}.`)).toBeVisible();
  await expect(
    page
      .getByRole("region", { name: "Ended consents" })
      .getByRole("article", { name: clientName, exact: true })
      .first(),
  ).toBeVisible();

  const after = await consoleGet<{
    arrangements: (ConsoleArrangement & { withdrawalMethod?: string })[];
  }>(`/api/customers/${customerNumber}`);
  const withdrawn = after.arrangements.filter((a) => a.withdrawalMethod === "CONSUMER_DASHBOARD");
  expect(withdrawn.length, "the withdrawal has the method CONSUMER_DASHBOARD").toBeGreaterThan(0);
});
