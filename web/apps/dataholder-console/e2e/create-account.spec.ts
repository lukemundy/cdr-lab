import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import {
  createAccount,
  createCustomer,
  uniqueAccountNumber,
  uniqueName,
} from "@cdr-lab/console/e2e/dataholder-lab";

const pairError = "An account has this bsb and accountNumber.";

// The test makes each record, so it needs no data in the lab. All data is invented.
test("the operator creates an account, and the data holder refuses the pair of a different account first", async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  const name = uniqueName();
  const customer = await createCustomer({
    firstName: "Alex",
    lastName: `Quokka${name}`,
    emailAddress: `${name}@mail.test`,
  });
  const other = await createCustomer({
    lastName: `Numbat${name}`,
    emailAddress: `n${name}@mail.test`,
  });
  const takenNumber = uniqueAccountNumber();
  await createAccount(other.customerNumber, {
    displayName: `Taken ${name}`,
    productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
    productName: "Everyday Account",
    bsb: "372100",
    accountNumber: takenNumber,
  });

  await page.goto(`/customers/${customer.customerNumber}`);
  await expect(page.getByText("This customer has no accounts")).toBeVisible();
  await page.getByRole("link", { name: "Create account" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    `Create an account for Alex Quokka${name}`,
  );
  await expect(page).toHaveURL(`/customers/${customer.customerNumber}/create-account`);
  await expectNoAxeViolations(page);

  const displayName = `Bills ${name}`;
  await page.getByRole("textbox", { name: "Display name" }).fill(displayName);
  await page.getByRole("combobox", { name: "Product category" }).selectOption("TRAVEL_CARDS");
  await page.getByRole("textbox", { name: "Product name" }).fill("Travel Card");
  await page.getByRole("textbox", { name: "BSB" }).fill("372100");
  const accountNumber = page.getByRole("textbox", { name: "Account number" });
  await accountNumber.fill(takenNumber);
  await page.getByRole("button", { name: "Create account" }).click();

  await expect(page.getByRole("alert").getByRole("link", { name: pairError })).toBeVisible();
  await expect(accountNumber).toHaveAccessibleDescription(pairError);
  await expectNoAxeViolations(page);

  const newNumber = uniqueAccountNumber();
  await accountNumber.fill(newNumber);
  await page.getByRole("button", { name: "Create account" }).click();

  await expect(page.getByRole("heading", { level: 1, name: `Alex Quokka${name}` })).toBeFocused();
  await expect(page).toHaveURL(`/customers/${customer.customerNumber}`);
  await expect(page.getByRole("dialog", { name: `Account ${displayName} created` })).toBeVisible();
  await page.getByRole("button", { name: "Show all accounts" }).click();
  const account = page.getByRole("region", { name: displayName, exact: true });
  await expect(account.getByText("TRAVEL_CARDS")).toBeVisible();
  await expect(account.getByText(newNumber)).toBeVisible();
  await expect(account.getByText("OPEN")).toBeVisible();
  await expectNoAxeViolations(page);
  expect(errors).toEqual([]);
});
