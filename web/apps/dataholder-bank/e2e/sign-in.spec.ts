import { randomInt, randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import {
  createAccount,
  createTransaction,
  uniqueAccountNumber,
} from "@cdr-lab/console/e2e/dataholder-lab";

// The journey makes its own customer, account and transaction, with invented values, because the lab has no seed
// customers.
test("a consumer signs up, signs out, signs in with any password, and sees the account that the operator opens", async ({
  page,
  context,
}) => {
  const lastName = `Quoll ${randomInt(1_000_000)}`;
  // An email address identifies one customer of the bank.
  const emailAddress = `quoll.${randomUUID()}@example.test`;

  await page.goto("/");
  await page
    .getByRole("navigation", { name: "Main" })
    .getByRole("link", { name: "Sign in" })
    .click();
  await page.getByRole("link", { name: "Sign up" }).click();
  await expect(page).toHaveTitle(/^Sign up - .+ online banking$/);
  await expectNoAxeViolations(page);

  await page.getByLabel("First name (optional)").fill("Zebracorn");
  await page.getByLabel("Last name").fill(lastName);
  await page.getByLabel("Email address").fill(emailAddress);
  await page.getByRole("button", { name: "Sign up" }).click();

  await expect(page.getByRole("heading", { level: 1, name: "Your customer number" })).toBeVisible();
  const customerNumber = await page.getByText(/^\d{8}$/).textContent();
  if (!customerNumber) throw new Error("The page shows no customer number.");
  await expectNoAxeViolations(page);

  const [cookie] = await context.cookies();
  expect(cookie?.name).toMatch(/^__Host-bank-session-[0-9a-f-]{36}$/);
  expect(cookie).toMatchObject({ httpOnly: true, secure: true, sameSite: "Strict", expires: -1 });

  await page.getByRole("button", { name: "Sign out" }).click();
  await expect(
    page.getByRole("heading", { level: 1, name: "Sign in to online banking" }),
  ).toBeVisible();
  await expect(page).toHaveTitle(/^Sign in - .+ online banking$/);
  expect(await context.cookies()).toHaveLength(0);
  await expectNoAxeViolations(page);

  await page.getByLabel("Email address or customer number").fill(customerNumber);
  await page.getByLabel("Password").fill("any password at all");
  await page.getByRole("button", { name: "Sign in" }).click();

  await expect(page.getByRole("heading", { level: 1, name: "Your accounts" })).toBeVisible();
  await expect(page.getByText("You have no account with us yet.")).toBeVisible();
  await expectNoAxeViolations(page);

  const accountNumber = uniqueAccountNumber();
  const accountId = await createAccount(customerNumber, {
    displayName: "Everyday",
    productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
    productName: "Everyday Account",
    bsb: "062000",
    accountNumber,
  });
  await createTransaction(customerNumber, accountId, {
    type: "PAYMENT",
    status: "POSTED",
    description: "Corner Cafe",
    amount: "-4.50",
    reference: "Card",
    // An hour ago, so that the time is in the default period of the transactions also with a small clock skew.
    postingDateTime: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
  });

  await page.reload();
  const open = page.getByRole("region", { name: "Open accounts" });
  await expect(
    open.getByText(`Account number:xxx-xxx xxxx${accountNumber.slice(-4)}`),
  ).toBeVisible();
  await expect(open.getByText("Balance:-$4.50")).toBeVisible();
  await open.getByRole("link", { name: "Everyday" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Everyday" })).toBeVisible();
  await expect(page.getByRole("listitem").filter({ hasText: "Corner Cafe" })).toContainText(
    "-$4.50",
  );
  await expectNoAxeViolations(page);

  await page.goto("/customer-number");
  await expect(page.getByText(customerNumber, { exact: true })).toBeVisible();

  await page.getByRole("button", { name: "Sign out" }).click();
  await expect(page).toHaveURL(/\/sign-in$/);
  await page.goto("/customer-number");
  await expect(
    page.getByRole("heading", { level: 1, name: "Sign in to online banking" }),
  ).toBeVisible();

  await page.getByLabel("Email address or customer number").fill(emailAddress.toUpperCase());
  await page.getByLabel("Password").fill("any password at all");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Your accounts" })).toBeVisible();
  await page.goto("/customer-number");
  await expect(page.getByText(customerNumber, { exact: true })).toBeVisible();
});

test("a sign-in with a customer number that does not exist is refused", async ({ page }) => {
  await page.goto("/sign-in");
  await page.getByLabel("Email address or customer number").fill("00000000");
  await page.getByLabel("Password").fill("any");
  await page.getByRole("button", { name: "Sign in" }).click();

  // The lab has no customer 00000000, because a random number of 8 digits is rarely 0.
  await expect(page.getByRole("alert")).toContainText(
    "We could not sign you in. Check your email address or customer number.",
  );
});
