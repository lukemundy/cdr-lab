import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { Transaction } from "../api/client";
import { account, customerDetail, http, renderApp, transaction, worker } from "../test/app";

const number = "12345678";
const accountId = "6f1c1a47-3c2e-4a0b-9d55-3f8f4f1e2a10";
const transactionsUrl = `/customers/${number}/accounts/${accountId}/transactions`;
const createUrl = `/customers/${number}/accounts/${accountId}/create-transaction`;

type Result = { status: 201 } | { status: 400 | 409; body: unknown };

// The list has the new transaction after a create.
function serve(results: Result[] = [], openStatus: "OPEN" | "CLOSED" = "OPEN") {
  const bodies: unknown[] = [];
  const listed: Transaction[] = [];
  worker.use(
    http.get("/api/customers/{customerNumber}", ({ response }) =>
      response(200).json(
        customerDetail(number, {
          accounts: [account(accountId, openStatus === "OPEN" ? {} : { openStatus, actions: [] })],
        }),
      ),
    ),
    http.get("/api/customers/{customerNumber}/accounts/{accountId}/transactions", ({ response }) =>
      response(200).json({ data: listed, totalRecords: listed.length }),
    ),
    http.post(
      "/api/customers/{customerNumber}/accounts/{accountId}/transactions",
      async ({ request }) => {
        bodies.push(await request.json());
        const result = results[bodies.length - 1] ?? { status: 201 };
        if (result.status === 201) {
          listed.push(transaction("transaction-new", { description: "Groceries" }));
          return new Response(null, {
            status: 201,
            headers: {
              Location: `/api/customers/${number}/accounts/${accountId}/transactions/new`,
            },
          });
        }
        return new Response(JSON.stringify(result.body), {
          status: result.status,
          headers: { "Content-Type": "application/problem+json" },
        });
      },
    ),
  );
  return bodies;
}

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const select = (name: string) => page.getByRole("combobox", { name, exact: true });
const time = (name: string) => page.getByLabelText(name, { exact: true });
const create = () => page.getByRole("button", { name: /^Create transaction/ });

const serviceError = "extendedData.nppPayload.service must be a value of the CDS enumeration";
const payerError = "extendedData.payer must be present for a positive amount";

const refusal = {
  type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
  title: "Values in the request are not valid",
  status: 400,
  detail: "One or more values in the request are not valid.",
  errors: [
    { pointer: "#/extendedData/nppPayload/service", detail: serviceError },
    { pointer: "#/extendedData/payer", detail: payerError },
  ],
};

async function fillRequired() {
  await userEvent.selectOptions(select("Type"), "PAYMENT");
  await userEvent.selectOptions(select("Status"), "POSTED");
  await userEvent.fill(field("Description"), " Groceries ");
  await userEvent.fill(field("Amount"), "-87.65 ");
}

describe("the create of a transaction", () => {
  test("opens from the transactions of an open account, with no value chosen", async () => {
    serve();
    const router = renderApp(transactionsUrl);

    await page.getByRole("link", { name: "Create transaction" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Create a transaction");
    await expect.element(page.getByText("For Everyday.", { exact: false })).toBeVisible();
    expect(router.state.location.pathname).toBe(createUrl);
    await expect.element(select("Type")).toHaveValue("");
    await expect.element(select("Status")).toHaveValue("");
    await expect.element(field("Description")).toHaveValue("");
    await expect.element(time("Posting time (optional)")).toHaveValue("");
    await expectNoAxeViolations(document.body);
  });

  test("the type and the status are the values of the CDS", async () => {
    serve();
    renderApp(createUrl);

    const options = (name: string) =>
      [...select(name).element().querySelectorAll("option")].map((o) => o.value);

    await expect.element(select("Type")).toBeVisible();
    expect(options("Type")).toEqual([
      "",
      "DIRECT_DEBIT",
      "FEE",
      "INTEREST_CHARGED",
      "INTEREST_PAID",
      "OTHER",
      "PAYMENT",
      "TRANSFER_INCOMING",
      "TRANSFER_OUTGOING",
    ]);
    expect(options("Status")).toEqual(["", "PENDING", "POSTED"]);
    expect(options("Extension type")).toEqual(["", "nppPayload"]);
    expect(options("NPP service")).toEqual(["", "X2P1", "IFTI", "BSCT", "CATSCT"]);
  });

  test("the transactions of a closed account have no create", async () => {
    serve([], "CLOSED");
    renderApp(transactionsUrl);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Transactions of Everyday" }))
      .toBeVisible();
    await expect.element(page.getByRole("link", { name: "Customer" })).toBeVisible();
    await expect
      .element(page.getByRole("link", { name: "Create transaction" }))
      .not.toBeInTheDocument();
  });

  test("the detail of the customer shows only the close as an action of the account", async () => {
    serve();
    renderApp(`/customers/${number}`);
    await page.getByRole("button", { name: "Show all accounts" }).click();

    const actions = page.getByRole("group", { name: "Actions of Everyday" });
    await expect.element(actions.getByRole("button", { name: "Close" })).toBeVisible();
    await expect.element(actions.getByRole("button")).toHaveLength(1);
  });

  test("sends each value with no white space and each time in UTC, then shows the list with a toast", async () => {
    const bodies = serve([{ status: 201 }]);
    const router = renderApp(createUrl);

    await fillRequired();
    // The browser of the tests is in Australia/Perth, UTC+8.
    await userEvent.fill(time("Posting time (optional)"), "2026-10-26T14:05:30");
    await userEvent.fill(field("Merchant name (optional)"), " Example Mart");
    await create().click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Transactions of Everyday" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        type: "PAYMENT",
        status: "POSTED",
        description: "Groceries",
        amount: "-87.65",
        postingDateTime: "2026-10-26T06:05:30.000Z",
        merchantName: "Example Mart",
      },
    ]);
    expect(router.state.location.pathname).toBe(transactionsUrl);
    await expect
      .element(page.getByRole("table", { name: "Transactions" }).getByText("Groceries"))
      .toBeVisible();
    await expect
      .element(page.getByRole("dialog", { name: "Transaction Groceries created" }))
      .toBeVisible();
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  test("sends no type and no status when the operator chose none, so that the data holder refuses them", async () => {
    const bodies = serve([{ status: 400, body: { ...refusal, errors: [] } }]);
    renderApp(createUrl);

    await create().click();

    await expect.element(page.getByRole("alert")).toBeVisible();
    expect(bodies).toEqual([{ description: "", amount: "" }]);
  });

  test("sends the extended data when one of its fields has a value", async () => {
    const bodies = serve([{ status: 201 }]);
    renderApp(createUrl);

    await fillRequired();
    await userEvent.fill(field("Payee (optional)"), "Example Mart ");
    await userEvent.selectOptions(select("Extension type"), "nppPayload");
    await userEvent.selectOptions(select("NPP service"), "X2P1");
    await create().click();

    await expect.element(page.getByRole("dialog", { name: /created$/ })).toBeVisible();
    expect(bodies).toEqual([
      {
        type: "PAYMENT",
        status: "POSTED",
        description: "Groceries",
        amount: "-87.65",
        extendedData: {
          payee: "Example Mart",
          extensionUType: "nppPayload",
          nppPayload: { extendedDescription: "", service: "X2P1", serviceVersion: "" },
        },
      },
    ]);
  });

  test("shows a refused value of the extended data at its field and in the summary, and keeps the values", async () => {
    const bodies = serve([{ status: 400, body: refusal }]);
    renderApp(createUrl);

    await fillRequired();
    await userEvent.fill(field("Payee (optional)"), "Example Mart");
    await create().click();

    const summary = page.getByRole("alert");
    await expect.element(summary.getByText(refusal.detail)).toBeVisible();
    await expect.element(summary.getByRole("link", { name: serviceError })).toBeVisible();
    await expect.element(select("NPP service")).toHaveAccessibleDescription(serviceError);
    await expect
      .element(field("Payer (optional)"))
      .toHaveAccessibleDescription(`For a positive amount. ${payerError}`);
    await expect.element(field("Payee (optional)")).toHaveValue("Example Mart");
    await expect.element(select("Type")).toHaveValue("PAYMENT");
    expect(bodies).toHaveLength(1);
    await expectNoAxeViolations(document.body);

    await summary.getByRole("link", { name: serviceError }).click();
    await expect.poll(() => document.activeElement).toBe(select("NPP service").element());
  });

  test("shows the refusal of a closed account in the summary", async () => {
    const detail = "The account is closed. The data holder changed nothing.";
    serve([{ status: 409, body: { type: "about:blank", title: "Conflict", status: 409, detail } }]);
    renderApp(createUrl);

    await fillRequired();
    await create().click();

    await expect.element(page.getByRole("alert").getByText(detail)).toBeVisible();
  });

  test("Cancel goes back to the transactions and sends nothing", async () => {
    const bodies = serve();
    const router = renderApp(createUrl);

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Transactions of Everyday" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe(transactionsUrl);
    expect(bodies).toEqual([]);
  });
});
