import { expect, test, type Page } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import {
  createAccount,
  createCustomer,
  uniqueAccountNumber,
  uniqueName,
} from "@cdr-lab/console/e2e/dataholder-lab";

// The lab browser tests make no arrangement (see detail.spec.ts), so the close of the last open account ends no
// arrangement here. The unit tests of the console API and of the UI cover the arrangements that a close ends.
test("the operator closes an account and then the last open account of a customer", async ({
  page,
}) => {
  const name = uniqueName();
  const customer = await createCustomer({
    lastName: `Quokka${name}`,
    emailAddress: `${name}@mail.test`,
  });
  const accounts = [`Everyday ${name}`, `Saver ${name}`];
  await Promise.all(
    accounts.map((displayName) =>
      createAccount(customer.customerNumber, {
        displayName,
        productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
        productName: "Everyday Account",
        bsb: "372100",
        accountNumber: uniqueAccountNumber(),
      }),
    ),
  );

  await page.goto(`/customers/${customer.customerNumber}`);
  await page.getByRole("button", { name: "Show all accounts" }).click();

  // The second close is of the last open account.
  await closeInConsole(page, `Everyday ${name}`);
  await closeInConsole(page, `Saver ${name}`);
});

async function closeInConsole(page: Page, displayName: string) {
  const account = page.getByRole("region", { name: displayName, exact: true });
  await account.getByRole("button", { name: "Close", exact: true }).click();

  const dialog = page.getByRole("alertdialog", { name: `Close ${displayName}?` });
  await expect(dialog).toContainText(`${displayName} changes from OPEN to CLOSED.`);
  await expect(dialog).toContainText("You cannot reverse this action.");
  await expectNoAxeViolations(page);
  await dialog.getByRole("button", { name: "Close", exact: true }).click();

  await expect(dialog).toBeHidden();
  await expect(page.getByRole("dialog", { name: `${displayName} is now CLOSED` })).toBeVisible();
  await expect(account.getByText("CLOSED", { exact: true })).toBeVisible();
  await expect(account.getByRole("button", { name: "Close", exact: true })).toHaveCount(0);
}
