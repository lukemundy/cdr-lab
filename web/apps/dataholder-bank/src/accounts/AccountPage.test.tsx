import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import {
  account,
  narrowWidth,
  pendingTransaction,
  presets,
  renderApp,
  serveAccounts,
  showPreset,
  transaction,
} from "../test/app";

const card = account("account-1", {
  displayName: "Low Rate Card",
  productName: "Low Rate Card",
  productCategory: "CRED_AND_CHRG_CARDS",
  creationDate: "2026-10-01",
  balance: { currentBalance: "-120.00", availableBalance: "880.00", creditLimit: "1000.00" },
});

const rows = () =>
  [...document.querySelectorAll("main li[data-transaction]")].map((row) =>
    [...row.querySelectorAll("[data-cell]")].map((cell) => cell.textContent),
  );

// Verifies: DH-2203, DH-2206, DH-2207
test("the page of an account shows its fields, its balance and its transactions", async () => {
  serveAccounts([card], {
    "account-1": [
      transaction("transaction-1", {
        description: "Corner Cafe",
        amount: "-4.50",
        postingDateTime: "2026-10-30T02:00:00Z",
        merchantName: "Corner Cafe Pty Ltd",
        merchantCategoryCode: "5814",
        reference: "Card 1234",
      }),
      pendingTransaction("transaction-2", {
        type: "TRANSFER_INCOMING",
        description: "Refund",
        amount: "12.00",
        executionDateTime: "2026-10-29T02:00:00Z",
      }),
    ],
  });

  renderApp("/accounts/account-1");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Low Rate Card" }))
    .toBeVisible();
  await expect.poll(() => document.title).toBe("Low Rate Card - Example Bank online banking");
  await expect.element(page.getByText("Credit limit:$1,000.00")).toBeVisible();
  const details = [...(document.querySelector("main dl")?.children ?? [])].map(
    (r) => r.textContent,
  );
  expect(details).toEqual([
    "Product:Low Rate Card",
    "Type:Credit and charge cards",
    "Account number:xxx-xxx xxxx5678",
    "Balance:-$120.00",
    "Available:$880.00",
    "Credit limit:$1,000.00",
    "Status:Open",
    "Opened:1 October 2026",
  ]);

  await expect.element(page.getByRole("heading", { level: 2, name: "Transactions" })).toBeVisible();
  await expect.poll(rows).toEqual([
    ["30 October 2026", "Corner Cafe", "-$4.50"],
    ["29 October 2026", "Refund", "Pending", "+$12.00"],
  ]);

  expect(page.getByText("Reference:Card 1234").query()).not.toBeVisible();
  await userEvent.click(page.getByText("Details of Corner Cafe"));
  await expect.element(page.getByText("Type:Payment")).toBeVisible();
  await expect.element(page.getByText("Reference:Card 1234")).toBeVisible();
  await expect.element(page.getByText("Merchant:Corner Cafe Pty Ltd")).toBeVisible();
  await expect.element(page.getByText("Merchant category code:5814")).toBeVisible();
});

// Verifies: DH-2205
test("the transactions have pages", async () => {
  const many = Array.from({ length: 27 }, (_, i) =>
    transaction(`transaction-${i}`, { description: `Purchase ${i + 1}` }),
  );
  serveAccounts([card], { "account-1": many });

  renderApp("/accounts/account-1");

  await expect.element(page.getByText("Purchase 1", { exact: true })).toBeVisible();
  await expect.element(page.getByText("Page 1 of 2")).toBeVisible();
  await userEvent.click(page.getByRole("link", { name: "Next page" }));
  await expect.element(page.getByText("Purchase 26")).toBeVisible();
  await expect.element(page.getByText("Page 2 of 2")).toBeVisible();
  expect(page.getByRole("link", { name: "Next page" }).query()).toBeNull();
  await expect.element(page.getByRole("link", { name: "Previous page" })).toBeVisible();
});

test("an account with no transactions says so", async () => {
  serveAccounts([card]);

  renderApp("/accounts/account-1");

  await expect
    .element(page.getByText("This account has no transactions in the last 90 days."))
    .toBeVisible();
  expect(page.getByRole("navigation", { name: "Pages of transactions" }).query()).toBeNull();
});

// Verifies: DH-2209
test("an account of a different customer, or that does not exist, is not found", async () => {
  serveAccounts([card]);

  renderApp("/accounts/account-9");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Page not found" }))
    .toBeVisible();
  expect(document.body.textContent).not.toContain("account-9");
});

describe.each(presets)("with the preset $name", (preset) => {
  test("the page of an account passes the accessibility checks and has no horizontal scroll at 320 CSS px", async () => {
    serveAccounts([card], {
      "account-1": [
        transaction("transaction-1", {
          description: "A very long description of a payment to a merchant",
          billerCode: "1234",
          billerName: "Power Co",
          crn: "98765432109876543210",
        }),
        pendingTransaction("transaction-2"),
      ],
    });

    await showPreset(preset, "/accounts/account-1");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Low Rate Card" }))
      .toBeVisible();
    await userEvent.click(
      page.getByText("Details of A very long description of a payment to a merchant"),
    );
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
    await expectNoAxeViolations(document.documentElement);
  });
});
