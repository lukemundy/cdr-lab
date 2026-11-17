import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { Account } from "../api/client";
import {
  account,
  endSession,
  http,
  narrowWidth,
  noSession,
  presets,
  renderApp,
  serveAccounts,
  sessionEnded,
  showPreset,
  worker,
} from "../test/app";

// All account and consent data in the tests is invented.
const budgetApp = { arrangementId: "arrangement-1", clientName: "Budget Buddy" };
const taxApp = { arrangementId: "arrangement-2", clientName: "Tax Helper" };

const lastAccount = () =>
  account("account-1", {
    displayName: "Everyday",
    maskedNumber: "xxx-xxx xxxx5678",
    balance: { currentBalance: "12.50", availableBalance: "12.50", creditLimit: "0.00" },
    actions: [{ action: "close", ends: [budgetApp, taxApp] }],
  });

// serveCloses gives each close to respond, which can change the accounts that the site API serves.
function serveCloses(
  accounts: Account[],
  respond: (accountId: string, ends: string[]) => Response,
) {
  serveAccounts(accounts);
  const requests: { accountId: string; ends: string[] }[] = [];
  worker.use(
    http.get("/api/accounts", ({ response }) => response(200).json({ accounts })),
    http.post("/api/accounts/{accountId}/close", async ({ params, request }) => {
      const { ends } = await request.json();
      requests.push({ accountId: params.accountId, ends });
      return respond(params.accountId, ends);
    }),
  );
  return requests;
}

const closeOf = (accounts: Account[]) => (accountId: string) => {
  const index = accounts.findIndex((a) => a.accountId === accountId);
  accounts[index] = { ...accounts[index]!, openStatus: "CLOSED", actions: [] };
  return new Response(null, { status: 204 });
};

const problem = (status: number, detail: string) =>
  Response.json({ type: "about:blank", title: "Refused", status, detail }, { status });

const dialog = () => page.getByRole("alertdialog");

// Verifies: DH-2240
test("a closed account has no close action", async () => {
  serveAccounts([account("account-2", { displayName: "Old Saver", openStatus: "CLOSED" })]);

  renderApp("/accounts/account-2");

  await expect.element(page.getByRole("heading", { level: 1, name: "Old Saver" })).toBeVisible();
  expect(page.getByRole("button", { name: "Close account" }).query()).toBeNull();
});

// Verifies: DH-2241, DH-2242
test("the confirmation of an account that is not the last names the account and its balance, and no consent", async () => {
  serveAccounts([
    account("account-1", {
      balance: { currentBalance: "-4.50", availableBalance: "995.50", creditLimit: "1000.00" },
    }),
    account("account-2", { displayName: "Saver" }),
  ]);

  renderApp("/accounts/account-1");
  await userEvent.click(page.getByRole("button", { name: "Close account" }));

  await expect.element(dialog().getByRole("heading", { name: "Close Everyday?" })).toBeVisible();
  const text = dialog().element().textContent;
  expect(text).toContain("Everyday, account number xxx-xxx xxxx5678, has a balance of -$4.50.");
  expect(text).toContain("This is a demonstration bank, so it pays out no balance.");
  expect(text).toContain("You cannot reverse this action.");
  expect(text).not.toContain("Consumer Data Right");
  await expectNoAxeViolations(document.documentElement);
});

// Verifies: DH-2243, DH-2248
test("the close of the last open account names each app whose consent ends, then shows the closed account", async () => {
  const accounts = [lastAccount()];
  const requests = serveCloses(accounts, closeOf(accounts));

  renderApp("/accounts/account-1");
  await userEvent.click(page.getByRole("button", { name: "Close account" }));

  const text = dialog().element().textContent;
  expect(text).toContain(
    "This is your last open account. After the close, you cannot share your data under the Consumer Data Right.",
  );
  expect(text).toContain(
    "Your consent for Budget Buddy ends, and Example Bank tells Budget Buddy.",
  );
  expect(text).toContain("Your consent for Tax Helper ends, and Example Bank tells Tax Helper.");

  await userEvent.click(dialog().getByRole("button", { name: "Close account" }));

  const heading = page.getByRole("heading", { level: 1, name: "Your accounts" });
  await expect.element(heading).toBeVisible();
  await expect.element(page.getByText("You closed Everyday.")).toBeVisible();
  await expect.element(heading).toHaveFocus();
  const closed = page.getByRole("region", { name: "Closed accounts" });
  await expect.element(closed.getByRole("link", { name: "Everyday" })).toBeVisible();
  await expect
    .element(
      page.getByText(
        "You cannot share your data under the Consumer Data Right until you have an open account.",
      ),
    )
    .toBeVisible();
  expect(requests).toEqual([{ accountId: "account-1", ends: ["arrangement-1", "arrangement-2"] }]);
});

// Verifies: DH-2245
test("a close that now ends different consents keeps the dialog as it was, says that the accounts changed, and loads the new state", async () => {
  const accounts = [lastAccount()];
  const requests = serveCloses(accounts, () => {
    accounts[0] = {
      ...accounts[0]!,
      actions: [
        {
          action: "close",
          ends: [budgetApp, taxApp, { arrangementId: "arrangement-3", clientName: "Late App" }],
        },
      ],
    };
    return problem(
      409,
      "Your accounts or your consents changed after you opened this confirmation.",
    );
  });

  renderApp("/accounts/account-1");
  await userEvent.click(page.getByRole("button", { name: "Close account" }));
  await userEvent.click(dialog().getByRole("button", { name: "Close account" }));

  await expect
    .element(
      dialog().getByText(
        "Your accounts or your consents changed after you opened this confirmation.",
      ),
    )
    .toBeVisible();
  expect(dialog().element().textContent).not.toContain("Late App");

  await userEvent.click(dialog().getByRole("button", { name: "Cancel" }));
  await expect.element(dialog()).not.toBeInTheDocument();
  await userEvent.click(page.getByRole("button", { name: "Close account" }));
  await expect
    .element(dialog().getByText("Your consent for Late App ends, and Example Bank tells Late App."))
    .toBeVisible();
  expect(requests).toHaveLength(1);
});

// Verifies: DH-2249
test("a close that fails keeps the dialog open with the error, and the site does not send it again", async () => {
  const requests = serveCloses([lastAccount()], () =>
    problem(500, "The data holder could not complete the request."),
  );

  renderApp("/accounts/account-1");
  await userEvent.click(page.getByRole("button", { name: "Close account" }));
  await userEvent.click(dialog().getByRole("button", { name: "Close account" }));

  await expect
    .element(dialog().getByText("The data holder could not complete the request."))
    .toBeVisible();
  expect(requests).toHaveLength(1);
});

// Verifies: DH-2028
test("a close after the end of the session goes to the sign-in page, which says so", async () => {
  serveCloses([lastAccount()], () => {
    endSession();
    worker.use(http.get("/api/accounts", ({ response }) => response(403).json(noSession)));
    return Response.json(sessionEnded, { status: 403 });
  });

  renderApp("/accounts/account-1");
  await userEvent.click(page.getByRole("button", { name: "Close account" }));
  await userEvent.click(dialog().getByRole("button", { name: "Close account" }));

  await expect
    .element(page.getByRole("status").filter({ hasText: "signed you out" }))
    .toHaveTextContent("Example Bank signed you out. Sign in again.");
});

describe.each(presets)("with the preset $name", (preset) => {
  test("the confirmation of a close passes the accessibility checks and has no horizontal scroll at 320 CSS px", async () => {
    serveAccounts([lastAccount()]);

    await showPreset(preset, "/accounts/account-1");
    await userEvent.click(page.getByRole("button", { name: "Close account" }));

    await expect.element(dialog()).toBeVisible();
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
    await expectNoAxeViolations(document.documentElement);
  });
});
