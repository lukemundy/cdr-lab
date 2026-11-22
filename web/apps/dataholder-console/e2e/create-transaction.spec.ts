import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";
import {
  closeAccount,
  createAccount,
  createCustomer,
  uniqueAccountNumber,
  uniqueName,
} from "@cdr-lab/web-platform/e2e/dataholder-lab";

const payerError = "extendedData.payer must be present for a positive amount";

// The test makes its customer and accounts with the operator interface, so it needs no other data in the lab. All
// data is invented.
test("the operator creates a transaction for an open account, and the data holder refuses a value of the extended data first", async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  const name = uniqueName();
  const customer = await createCustomer({
    lastName: `Quokka${name}`,
    emailAddress: `${name}@mail.test`,
  });
  const [open, closed] = await Promise.all(
    [`Everyday ${name}`, `Saver ${name}`].map((displayName) =>
      createAccount(customer.customerNumber, {
        displayName,
        productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
        productName: "Everyday Account",
        bsb: "372100",
        accountNumber: uniqueAccountNumber(),
      }),
    ),
  );
  if (!open || !closed) throw new Error("the operator interface made no account");
  await closeAccount(customer.customerNumber, closed);
  const transactions = `/customers/${customer.customerNumber}/accounts/${open}/transactions`;

  await page.goto(`/customers/${customer.customerNumber}/accounts/${closed}/transactions`);
  await expect(
    page.getByRole("heading", { level: 1, name: `Transactions of Saver ${name}` }),
  ).toBeVisible();
  await expect(page.getByRole("link", { name: "Create transaction" })).toHaveCount(0);

  await page.goto(transactions);
  await page.getByRole("link", { name: "Create transaction" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Create a transaction");
  await expect(page).toHaveURL(
    `/customers/${customer.customerNumber}/accounts/${open}/create-transaction`,
  );
  await expectNoAxeViolations(page);

  await page.getByRole("combobox", { name: "Type", exact: true }).selectOption("TRANSFER_INCOMING");
  await page.getByRole("combobox", { name: "Status", exact: true }).selectOption("POSTED");
  await page.getByRole("textbox", { name: "Description", exact: true }).fill(`Refund ${name}`);
  await page.getByRole("textbox", { name: "Amount", exact: true }).fill("12.50");
  // The browser is in Australia/Perth, UTC+8.
  await page.getByLabel("Posting time (optional)").fill("2026-10-26T14:05:30");
  await page.getByRole("textbox", { name: "Payee (optional)" }).fill("Example Mart");
  await page.getByRole("combobox", { name: "Extension type" }).selectOption("nppPayload");
  await page.getByRole("combobox", { name: "NPP service" }).selectOption("X2P1");
  await page.getByRole("textbox", { name: "NPP service version" }).fill("03");
  await page.getByRole("textbox", { name: "Extended description" }).fill("Refund of an order");
  await page.getByRole("button", { name: "Create transaction" }).click();

  const payer = page.getByRole("textbox", { name: "Payer (optional)" });
  await expect(page.getByRole("alert").getByRole("link", { name: payerError })).toBeVisible();
  await expect(payer).toHaveAccessibleDescription(new RegExp(`${payerError}$`));
  await expect(page.getByRole("textbox", { name: "Payee (optional)" })).toHaveValue("Example Mart");
  await expectNoAxeViolations(page);

  await payer.fill("Example Mart");
  await page.getByRole("button", { name: "Create transaction" }).click();

  await expect(
    page.getByRole("heading", { level: 1, name: `Transactions of Everyday ${name}` }),
  ).toBeFocused();
  await expect(page).toHaveURL(transactions);
  await expect(
    page.getByRole("dialog", { name: `Transaction Refund ${name} created` }),
  ).toBeVisible();
  const row = page.getByRole("row", { name: new RegExp(`Refund ${name}`) });
  await expect(row).toContainText("TRANSFER_INCOMING");
  await expect(row).toContainText("$12.50");
  await expect(row).toContainText("Refund of an order");
  await expect(row.locator("time").first()).toHaveAttribute("datetime", "2026-10-26T06:05:30.000Z");
  await expectNoAxeViolations(page);
  expect(errors).toEqual([]);
});
