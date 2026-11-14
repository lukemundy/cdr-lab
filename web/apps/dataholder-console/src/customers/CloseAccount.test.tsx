import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { CustomerDetail } from "../api/client";
import { account, arrangement, customerDetail, http, renderApp, worker } from "../test/app";

const everydayId = "6f1c1a47-3c2e-4a0b-9d55-3f8f4f1e2a10";
const saverId = "0b3e7c55-8f0e-4d6b-a0c4-7d7f4a3b9e21";

const twoOpen = customerDetail("12345678", {
  accounts: [
    account(everydayId),
    account(saverId, { displayName: "Saver", accountNumber: "87654321" }),
  ],
  arrangements: [
    arrangement("arrangement-1"),
    arrangement("arrangement-2", { clientName: "Recipient Mobile" }),
  ],
});

const lastOpen = customerDetail("12345678", {
  accounts: [
    account(everydayId, { openStatus: "CLOSED", actions: [] }),
    account(saverId, {
      displayName: "Saver",
      accountNumber: "87654321",
      actions: [
        {
          action: "close",
          ends: [
            {
              arrangementId: "arrangement-1",
              softwareId: "software-1",
              clientName: "Recipient Web",
            },
            {
              arrangementId: "arrangement-2",
              softwareId: "software-2",
              clientName: "Recipient Mobile",
            },
          ],
        },
      ],
    }),
  ],
  arrangements: twoOpen.arrangements,
});

const endsChanged = {
  type: "about:blank",
  title: "Conflict",
  status: 409,
  detail:
    "The customer changed after the confirmation, and the close now ends different arrangements. " +
    "Look at the new state of the customer. The data holder changed nothing.",
} as const;

const conflict = {
  type: "about:blank",
  title: "Conflict",
  status: 409,
  detail: "The account is closed. The data holder changed nothing.",
} as const;

// The detail depends on the close requests, not on the count of reads, because a read of a previous test can
// arrive late.
function serve(
  before: CustomerDetail,
  after: CustomerDetail,
  result: "success" | "conflict" | "ends-changed",
) {
  const requests: { accountId: string; body: unknown }[] = [];
  worker.use(
    http.get("/api/customers/{customerNumber}", ({ response }) =>
      response(200).json(requests.length === 0 ? before : after),
    ),
    http.post(
      "/api/customers/{customerNumber}/accounts/{accountId}/close",
      async ({ params, request, response }) => {
        requests.push({ accountId: params.accountId, body: await request.json() });
        if (result === "success") return response(204).empty();
        return response(409).json(result === "conflict" ? conflict : endsChanged);
      },
    ),
  );
  return requests;
}

async function renderDetail() {
  renderApp("/customers/12345678");
  await page.getByRole("button", { name: "Show all accounts" }).click();
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const actionsOf = (name: string) =>
  page.getByRole("group", { name: `Actions of ${name}`, exact: true });

describe("the close of an account", () => {
  test("an open account has a close action, and a closed account has none", async () => {
    serve(lastOpen, lastOpen, "success");

    await renderDetail();

    await expect.element(actionsOf("Saver").getByRole("button", { name: "Close" })).toBeVisible();
    expect(actionsOf("Everyday").getByRole("button").query()).toBeNull();
    await expectNoAxeViolations(document.body);
  });

  test("a confirmed close sends one request with the arrangements that it ends, and the detail then shows the account as closed", async () => {
    const requests = serve(twoOpen, lastOpen, "success");
    await renderDetail();

    await userEvent.click(actionsOf("Everyday").getByRole("button", { name: "Close" }));

    const dialog = page.getByRole("alertdialog", { name: "Close Everyday?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        "Everyday changes from OPEN to CLOSED. You cannot reverse this action.",
      );
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Close", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([{ accountId: everydayId, body: { ends: [] } }]);
    await expect
      .element(page.getByRole("dialog", { name: "Everyday is now CLOSED" }))
      .toBeVisible();
    await expect.element(region("Everyday").getByText("CLOSED", { exact: true })).toBeVisible();
    expect(actionsOf("Everyday").getByRole("button").query()).toBeNull();
  });

  test("the confirmation of the close of the last open account names each arrangement that ends and its notice", async () => {
    serve(lastOpen, lastOpen, "success");
    await renderDetail();

    await userEvent.click(actionsOf("Saver").getByRole("button", { name: "Close" }));

    await expect
      .element(page.getByRole("alertdialog", { name: "Close Saver?" }))
      .toHaveAccessibleDescription(
        "Saver changes from OPEN to CLOSED. " +
          "It is the last open account of the customer, so 2 arrangements end. " +
          "The arrangement arrangement-1 of Recipient Web ends. " +
          "The data holder sends a notice to the data recipient. " +
          "The arrangement arrangement-2 of Recipient Mobile ends. " +
          "The data holder sends a notice to the data recipient. " +
          "You cannot reverse this action.",
      );
  });

  test("a close of an account that a different tab closed shows the refusal in the dialog, and then the current state", async () => {
    const requests = serve(twoOpen, lastOpen, "conflict");
    await renderDetail();

    await userEvent.click(actionsOf("Everyday").getByRole("button", { name: "Close" }));
    const dialog = page.getByRole("alertdialog", { name: "Close Everyday?" });
    await dialog.getByRole("button", { name: "Close", exact: true }).click();

    await expect.element(dialog.getByRole("alert")).toMatchTextContent(conflict.detail);
    expect(page.getByRole("dialog", { name: /is now/ }).query()).toBeNull();
    expect(requests).toHaveLength(1);

    await dialog.getByRole("button", { name: "Cancel" }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    await expect.element(region("Everyday").getByText("CLOSED", { exact: true })).toBeVisible();
    expect(actionsOf("Everyday").getByRole("button").query()).toBeNull();
    expect(requests).toHaveLength(1);
  });

  test("the close of the last open account sends the arrangements that its confirmation named", async () => {
    const requests = serve(lastOpen, lastOpen, "success");
    await renderDetail();

    await userEvent.click(actionsOf("Saver").getByRole("button", { name: "Close" }));
    await page
      .getByRole("alertdialog", { name: "Close Saver?" })
      .getByRole("button", { name: "Close", exact: true })
      .click();

    await expect.element(page.getByRole("alertdialog")).not.toBeInTheDocument();
    expect(requests).toEqual([
      { accountId: saverId, body: { ends: ["arrangement-1", "arrangement-2"] } },
    ]);
  });

  test("a close that a different tab made the last shows the refusal, and then names the arrangements that end", async () => {
    // The other tab closed Saver, so the close of Everyday now ends the arrangements.
    const otherTabClosedSaver = customerDetail("12345678", {
      accounts: [
        account(everydayId, { actions: lastOpen.accounts[1]?.actions ?? [] }),
        account(saverId, { displayName: "Saver", openStatus: "CLOSED", actions: [] }),
      ],
      arrangements: twoOpen.arrangements,
    });
    const requests = serve(twoOpen, otherTabClosedSaver, "ends-changed");
    await renderDetail();

    await userEvent.click(actionsOf("Everyday").getByRole("button", { name: "Close" }));
    const dialog = page.getByRole("alertdialog", { name: "Close Everyday?" });
    await dialog.getByRole("button", { name: "Close", exact: true }).click();

    await expect.element(dialog.getByRole("alert")).toMatchTextContent(endsChanged.detail);
    expect(requests).toEqual([{ accountId: everydayId, body: { ends: [] } }]);

    await dialog.getByRole("button", { name: "Cancel" }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    await expect.element(region("Saver").getByText("CLOSED", { exact: true })).toBeVisible();

    await userEvent.click(actionsOf("Everyday").getByRole("button", { name: "Close" }));

    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        /It is the last open account of the customer, so 2 arrangements end\./,
      );
    expect(requests).toHaveLength(1);
  });
});
