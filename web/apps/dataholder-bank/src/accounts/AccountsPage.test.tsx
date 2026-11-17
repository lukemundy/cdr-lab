import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page } from "vitest/browser";

import {
  account,
  endSession,
  http,
  narrowWidth,
  presets,
  renderApp,
  serveAccounts,
  serveBrand,
  serveSession,
  sessionEnded,
  showPreset,
  worker,
} from "../test/app";

const terms = (within: Element) =>
  [...within.querySelectorAll("dl > div")].map((row) => row.textContent ?? "");

const noSharing =
  "You cannot share your data under the Consumer Data Right until you have an open account.";

// Verifies: DH-2200, DH-2201
test("the accounts page shows the open accounts, then the closed accounts in a group", async () => {
  serveAccounts([
    account("account-1", {
      displayName: "Everyday",
      balance: { currentBalance: "-4.50", availableBalance: "995.50", creditLimit: "1000.00" },
    }),
    account("account-2", {
      displayName: "Old Saver",
      productName: "Bonus Saver",
      openStatus: "CLOSED",
      maskedNumber: "xxx-xxx xxxx4321",
      balance: { currentBalance: "0.00", availableBalance: "0.00", creditLimit: "0.00" },
    }),
    account("account-3", {
      displayName: "Card",
      productName: "Low Rate Card",
      productCategory: "CRED_AND_CHRG_CARDS",
    }),
  ]);

  renderApp("/accounts");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Your accounts" }))
    .toBeVisible();
  await expect.poll(() => document.title).toBe("Accounts - Example Bank online banking");
  const open = page.getByRole("region", { name: "Open accounts" });
  await expect.element(open.getByRole("link", { name: "Everyday" })).toBeVisible();
  const headings = [...document.querySelectorAll("main h3")].map((h) => h.textContent);
  expect(headings).toEqual(["Everyday", "Card", "Old Saver"]);
  expect(terms(open.element())).toEqual([
    "Product:Everyday Account",
    "Type:Everyday and savings accounts",
    "Account number:xxx-xxx xxxx5678",
    "Balance:-$4.50",
    "Available:$995.50",
    "Product:Low Rate Card",
    "Type:Credit and charge cards",
    "Account number:xxx-xxx xxxx5678",
    "Balance:$1,250.00",
    "Available:$1,200.00",
  ]);
  const closed = page.getByRole("region", { name: "Closed accounts" });
  await expect.element(closed.getByRole("link", { name: "Old Saver" })).toBeVisible();
  expect(terms(closed.element())).toContain("Account number:xxx-xxx xxxx4321");
  await expect
    .element(open.getByRole("link", { name: "Everyday" }))
    .toHaveAttribute("href", "/accounts/account-1");
  expect(page.getByText(noSharing).query()).toBeNull();
});

// Verifies: DH-2202
test("a customer with no account sees that they cannot share data", async () => {
  serveAccounts([]);

  renderApp("/accounts");

  await expect.element(page.getByText("You have no account with us yet.")).toBeVisible();
  await expect.element(page.getByText(noSharing)).toBeVisible();
  expect(page.getByRole("region", { name: "Closed accounts" }).query()).toBeNull();
});

// Verifies: DH-2202
test("a customer with only closed accounts sees that they cannot share data", async () => {
  serveAccounts([account("account-2", { displayName: "Old Saver", openStatus: "CLOSED" })]);

  renderApp("/accounts");

  await expect.element(page.getByText("You have no open account.")).toBeVisible();
  await expect.element(page.getByText(noSharing)).toBeVisible();
  await expect
    .element(
      page
        .getByRole("region", { name: "Closed accounts" })
        .getByRole("link", { name: "Old Saver" }),
    )
    .toBeVisible();
});

test("a failure of the site API shows the problem", async () => {
  serveSession();
  worker.use(
    http.get("/api/accounts", ({ response }) =>
      response("default").json(
        {
          type: "about:blank",
          title: "Internal Server Error",
          status: 500,
          detail: "The data holder could not complete the request.",
        },
        { status: 500 },
      ),
    ),
  );

  renderApp("/accounts");

  await expect
    .element(page.getByText("The data holder could not complete the request."))
    .toBeVisible();
});

test("with no session, the accounts page goes to the sign-in page", async () => {
  renderApp("/accounts");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
});

// Verifies: DH-2028
test("when the session ends, the accounts page goes to the sign-in page, which says so", async () => {
  serveAccounts([]);
  worker.use(
    http.get("/api/accounts", ({ response }) => {
      endSession();
      return response(403).json(sessionEnded);
    }),
  );

  renderApp("/accounts");

  await expect
    .element(page.getByRole("status").filter({ hasText: "signed you out" }))
    .toHaveTextContent("Example Bank signed you out. Sign in again.");
});

describe.each(presets)("with the preset $name", (preset) => {
  test("the accounts page passes the accessibility checks and has no horizontal scroll at 320 CSS px", async () => {
    serveAccounts([
      account("account-1", { displayName: "Everyday Transaction Account With A Long Name" }),
      account("account-2", { displayName: "Old Saver", openStatus: "CLOSED" }),
    ]);

    await showPreset(preset, "/accounts");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Your accounts" }))
      .toBeVisible();
    await expect.element(page.getByRole("link", { name: "Old Saver" })).toBeVisible();
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
    await expectNoAxeViolations(document.documentElement);
  });
});

// The dashed line of the ticket is the edge of the name, so that it never crosses the text of the card.
test("in the horizon layout, the ticket line of an account card is below the name of the account", async () => {
  await page.viewport(1280, 900);
  serveBrand({ layout: "horizon" });
  serveAccounts([account("account-1", { displayName: "Harbour Many Currencies Account" })]);

  renderApp("/accounts");

  await expect
    .element(page.getByRole("heading", { level: 3, name: "Harbour Many Currencies Account" }))
    .toBeVisible();
  const card = document.querySelector(".account-card");
  const name = card?.querySelector("h3");
  if (!card || !name) throw new Error("the page has no account card");
  expect(getComputedStyle(name).borderBlockEndStyle).toBe("dashed");
  expect(getComputedStyle(card, "::before").content).toBe("none");
});
