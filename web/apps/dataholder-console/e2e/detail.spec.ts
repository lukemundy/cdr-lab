import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import {
  createAccount,
  createCustomer,
  createTransaction,
  uniqueAccountNumber,
  uniqueName,
} from "@cdr-lab/console/e2e/dataholder-lab";

const day = (n: number) => new Date(Date.UTC(2026, 8, 1 + n, 1)).toISOString();

// The lab browser tests make no arrangement: an arrangement needs the full consent flow of a registered client. The
// unit tests of the console API and of the UI cover the arrangements.
test("the operator finds a customer, opens the detail and opens the transactions of an account, in pages", async ({
  page,
}) => {
  const name = uniqueName();
  const customer = await createCustomer({
    firstName: `Ari${name}`,
    lastName: `Quokka${name}`,
    emailAddress: `${name}@mail.test`,
  });
  const accountNumber = uniqueAccountNumber();
  const everyday = await createAccount(customer.customerNumber, {
    displayName: `Everyday ${name}`,
    productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
    productName: "Everyday Account",
    bsb: "372100",
    accountNumber,
    creditLimit: "100.00",
  });
  await createAccount(customer.customerNumber, {
    displayName: `Saver ${name}`,
    productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
    productName: "Savings Account",
    bsb: "372100",
    accountNumber: uniqueAccountNumber(),
  });
  // 26 posted transactions, one each day, and one pending: the newest is on the first page and the oldest on the
  // second page.
  await Promise.all(
    Array.from({ length: 26 }, (_, n) =>
      createTransaction(customer.customerNumber, everyday, {
        type: "PAYMENT",
        status: "POSTED",
        description: `Coffee ${n}`,
        amount: n === 0 ? "1000.00" : "-4.50",
        reference: "",
        postingDateTime: day(n),
      }),
    ),
  );
  await createTransaction(customer.customerNumber, everyday, {
    type: "PAYMENT",
    status: "PENDING",
    description: "Lunch",
    amount: "-20.00",
    reference: "",
    executionDateTime: day(30),
  });

  await page.goto("/customers");
  await page
    .getByRole("searchbox", { name: "Search by customer number, name or email" })
    .fill(name);
  await page.getByRole("link", { name: customer.customerNumber, exact: true }).click();

  await expect(page).toHaveURL(`/customers/${customer.customerNumber}`);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(`Ari${name} Quokka${name}`);
  await page.getByRole("button", { name: "Show all accounts" }).click();
  const account = page.getByRole("region", { name: `Everyday ${name}`, exact: true });
  await expect(account.getByText("372-100", { exact: true })).toBeVisible();
  await expect(account.getByText(accountNumber, { exact: true })).toBeVisible();
  // The balances of the banking APIs: the posted amounts, and the credit limit with the pending debits.
  await expect(account.getByText("$887.50", { exact: true })).toBeVisible();
  await expect(account.getByText("$967.50", { exact: true })).toBeVisible();
  await expect(page.getByText("This customer has no arrangements")).toBeVisible();
  await expectNoAxeViolations(page);

  await account.getByRole("link", { name: `Transactions of Everyday ${name}` }).click();

  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    `Transactions of Everyday ${name}`,
  );
  const table = page.getByRole("table", { name: "Transactions" });
  await expect(table.getByRole("row").nth(1)).toContainText("Lunch");
  await expect(table.getByRole("row").nth(1)).toContainText("PENDING");
  await expect(table.getByRole("row").nth(2)).toContainText("Coffee 25");
  await expect(page.getByText("27 records")).toBeVisible();
  await expectNoAxeViolations(page);

  await page.getByRole("button", { name: "Next page" }).click();

  await expect(page).toHaveURL(
    `/customers/${customer.customerNumber}/accounts/${everyday}/transactions?page=2`,
  );
  await expect(table.getByRole("row")).toHaveCount(3);
  await expect(table.getByRole("row").nth(2)).toContainText("Coffee 0");
  await expect(table.getByRole("row").nth(2)).toContainText("$1,000.00");
});

test("the detail of a customer that does not exist says not found", async ({ page }) => {
  await page.goto("/customers/00000000");

  await expect(page.getByRole("alert")).toContainText("Not Found");
  await expect(page.getByRole("alert")).toContainText(
    "The data holder has no customer with this customer number.",
  );
});
