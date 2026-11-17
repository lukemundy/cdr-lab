import { expect, test, type Page } from "@playwright/test";

import { definitionOf, expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import {
  closeAccount,
  createAccount,
  createCustomer,
  uniqueAccountNumber,
  uniqueName,
} from "@cdr-lab/console/e2e/dataholder-lab";

// playwright.config.ts runs this file after the other tests of the data holder, so that no other test changes the
// counts while it runs.

async function countOf(page: Page, term: string): Promise<number> {
  return Number(await definitionOf(page, term).textContent());
}

async function counts(page: Page) {
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: "Overview" })).toBeVisible();
  await expect(page.locator("dl dt").first()).toBeVisible();
  return {
    customers: await countOf(page, "Customers"),
    openAccounts: await countOf(page, "Open accounts"),
    closed: Number(
      /^(\d+) closed$/.exec((await page.getByText(/^\d+ closed$/).textContent()) ?? "")?.[1],
    ),
  };
}

const account = () => ({
  displayName: "Everyday",
  productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
  productName: "Lab Everyday Account",
  bsb: "372100",
  accountNumber: uniqueAccountNumber(),
});

// Verifies: DH-1404
test("each count of the overview agrees with the records that the operator interface creates", async ({
  page,
}) => {
  const before = await counts(page);

  const customer = await createCustomer({
    lastName: uniqueName(),
    emailAddress: `${uniqueName()}@mail.test`,
  });
  const closing = await createAccount(customer.customerNumber, account());
  await createAccount(customer.customerNumber, account());
  await closeAccount(customer.customerNumber, closing);

  const after = await counts(page);
  expect(after).toEqual({
    customers: before.customers + 1,
    openAccounts: before.openAccounts + 1,
    closed: before.closed + 1,
  });

  // The newest changes are the changes of this test, newest first, with no name or email address.
  const rows = page.getByRole("table", { name: "Newest changes" }).getByRole("row");
  await expect(rows.nth(1)).toContainText("Account closed");
  await expect(rows.nth(1).getByRole("link", { name: customer.customerNumber })).toBeVisible();
  await expect(page.getByRole("table", { name: "Newest changes" })).not.toContainText(
    customer.lastName,
  );
  await expectNoAxeViolations(page);
});
