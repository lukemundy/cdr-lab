import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { Arrangement, CustomerDetail } from "../api/client";
import { account, arrangement, customerDetail, http, renderApp, worker } from "../test/app";

const accountId = "6f1c1a47-3c2e-4a0b-9d55-3f8f4f1e2a10";
const webId = "8f4d7a52-3c1e-4b6a-9d2f-1e0c5b7a9d31";

const ended = {
  status: "ENDED",
  endedAt: "2026-10-16T01:00:00Z",
  endReason: "WITHDRAWN_BY_CONSUMER",
  withdrawalMethod: "ALTERNATIVE_METHOD",
  noticeStatus: "PENDING",
  noticeAttempts: [],
  actions: [],
} satisfies Partial<Arrangement>;

const current = customerDetail("12345678", {
  accounts: [account(accountId)],
  arrangements: [
    arrangement(webId),
    arrangement("arrangement-2", { clientName: "Recipient Mobile", ...ended }),
  ],
});

const withdrawn = customerDetail("12345678", {
  accounts: current.accounts,
  arrangements: [arrangement(webId, ended), current.arrangements[1] ?? arrangement("unused")],
});

const conflict = {
  type: "about:blank",
  title: "Conflict",
  status: 409,
  detail: "The arrangement ended. The data holder changed nothing.",
} as const;

// The detail depends on the withdraw requests, not on the count of reads, because a read of a previous test can
// arrive late.
function serve(before: CustomerDetail, after: CustomerDetail, result: "success" | "conflict") {
  const requests: { arrangementId: string; body: string }[] = [];
  worker.use(
    http.get("/api/customers/{customerNumber}", ({ response }) =>
      response(200).json(requests.length === 0 ? before : after),
    ),
    http.post(
      "/api/customers/{customerNumber}/arrangements/{arrangementId}/withdraw",
      async ({ params, request, response }) => {
        const body = await new Response(request.body).text();
        requests.push({ arrangementId: params.arrangementId, body });
        if (result === "success") return response(204).empty();
        return response(409).json(conflict);
      },
    ),
  );
  return requests;
}

async function renderDetail() {
  renderApp("/customers/12345678");
  await page.getByRole("button", { name: "Show all arrangements" }).click();
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const actionsOf = (name: string) =>
  page.getByRole("group", { name: `Actions of ${name}`, exact: true });

describe("the withdrawal of a consent", () => {
  // Verifies: DH-1447
  test("an arrangement that did not end has a withdraw action, and an arrangement that ended has none", async () => {
    serve(current, current, "success");

    await renderDetail();

    await expect
      .element(actionsOf("Recipient Web").getByRole("button", { name: "Withdraw" }))
      .toBeVisible();
    expect(actionsOf("Recipient Mobile").getByRole("button").query()).toBeNull();
    await expectNoAxeViolations(document.body);
  });

  test("a confirmed withdrawal sends one request with no body, and the detail then shows the end with the method", async () => {
    const requests = serve(current, withdrawn, "success");
    await renderDetail();

    await userEvent.click(actionsOf("Recipient Web").getByRole("button", { name: "Withdraw" }));

    const dialog = page.getByRole("alertdialog", {
      name: "Withdraw the consent of Recipient Web?",
    });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        `The arrangement ${webId} of Recipient Web ends. ` +
          "The data holder sends a notice to the data recipient. You cannot reverse this action.",
      );
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Withdraw", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([{ arrangementId: webId, body: "" }]);
    await expect
      .element(page.getByRole("dialog", { name: "The arrangement of Recipient Web is now ENDED" }))
      .toBeVisible();
    const web = region("Recipient Web");
    await expect.element(web.getByText("ENDED", { exact: true })).toBeVisible();
    await expect.element(web.getByText("ALTERNATIVE_METHOD", { exact: true })).toBeVisible();
    expect(actionsOf("Recipient Web").getByRole("button").query()).toBeNull();
  });

  // Verifies: DH-1448, DH-1449
  test("a withdrawal of an arrangement that a different tab withdrew shows the refusal in the dialog, and then the current state", async () => {
    const requests = serve(current, withdrawn, "conflict");
    await renderDetail();

    await userEvent.click(actionsOf("Recipient Web").getByRole("button", { name: "Withdraw" }));
    const dialog = page.getByRole("alertdialog", {
      name: "Withdraw the consent of Recipient Web?",
    });
    await dialog.getByRole("button", { name: "Withdraw", exact: true }).click();

    await expect.element(dialog.getByRole("alert")).toMatchTextContent(conflict.detail);
    expect(page.getByRole("dialog", { name: /is now/ }).query()).toBeNull();
    expect(requests).toHaveLength(1);

    await dialog.getByRole("button", { name: "Cancel" }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    await expect.element(region("Recipient Web").getByText("ENDED", { exact: true })).toBeVisible();
    expect(actionsOf("Recipient Web").getByRole("button").query()).toBeNull();
    expect(requests).toHaveLength(1);
  });
});
