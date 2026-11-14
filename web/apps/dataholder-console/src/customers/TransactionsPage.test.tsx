import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { afterEach, describe, expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { Transaction, TransactionList } from "../api/client";
import {
  account,
  customerDetail,
  definitionOf,
  http,
  renderApp,
  transaction,
  worker,
} from "../test/app";

const accountId = "6f1c1a47-3c2e-4a0b-9d55-3f8f4f1e2a10";
const url = `/customers/12345678/accounts/${accountId}/transactions`;

// Each request of the transactions, as the search of its URL.
function serve(body: TransactionList) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/customers/{customerNumber}", ({ response }) =>
      response(200).json(customerDetail("12345678", { accounts: [account(accountId)] })),
    ),
    http.get(
      "/api/customers/{customerNumber}/accounts/{accountId}/transactions",
      ({ params, request, response }) => {
        requests.push(
          `${params.customerNumber} ${params.accountId} ${new URL(request.url).search}`,
        );
        return response(200).json(body);
      },
    ),
  );
  return requests;
}

const table = () => page.getByRole("table", { name: "Transactions" });

const pending: Transaction = transaction("transaction-1", {
  status: "PENDING",
  type: "TRANSFER_INCOMING",
  description: "Salary",
  amount: "2500.00",
  executionDateTime: "2026-10-15T01:00:00Z",
  billerCode: "12345",
  billerName: "Example Biller",
  crn: "987654",
  apcaNumber: "123456",
});
delete pending.postingDateTime;

afterEach(() => {
  vi.restoreAllMocks();
});

describe("the transactions page", () => {
  test("shows each transaction of the page with its fields, and the total", async () => {
    const requests = serve({
      data: [
        transaction("transaction-2", {
          description: "Rent",
          amount: "-1200.00",
          reference: "Unit 4",
          valueDateTime: "2026-10-16T00:00:00Z",
          merchantName: "Example Agents",
          merchantCategoryCode: "6513",
          extendedData: {
            payee: "Example Agents",
            extensionUType: "nppPayload",
            nppPayload: { extendedDescription: "October", service: "X2P1", serviceVersion: "03" },
          },
        }),
        pending,
      ],
      totalRecords: 30,
    });

    renderApp(url);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Transactions of Everyday" }))
      .toBeVisible();
    const rent = table().getByRole("row").nth(1);
    // The tests run in Australia/Perth (vitest.config.ts).
    // Until the response, the first row is the row of the loading state, so wait for the data first.
    await expect.element(rent.getByRole("cell", { name: "Rent", exact: true })).toBeVisible();
    expect(rent.getByRole("cell").first().element().textContent).toContain(
      "15 October 2026, 1:30 pm UTC+8",
    );
    await expect.element(rent.getByRole("cell", { name: "PAYMENT" })).toBeVisible();
    await expect.element(rent.getByRole("cell", { name: "POSTED" })).toBeVisible();
    await expect.element(rent.getByRole("cell", { name: "-$1,200.00" })).toBeVisible();
    const rentValue = (term: string) => definitionOf(rent.element(), term).textContent;
    expect(rentValue("Transaction ID")).toContain("transaction-2");
    expect(rentValue("Reference")).toBe("Unit 4");
    expect(rentValue("Value date")).toContain("16 October 2026, 8:00 am UTC+8");
    expect(rentValue("Merchant name")).toBe("Example Agents");
    expect(rentValue("Merchant category code")).toBe("6513");
    expect(rentValue("Payee")).toBe("Example Agents");
    expect(rentValue("Extended description")).toBe("October");
    expect(rentValue("NPP service")).toBe("X2P1");
    expect(rentValue("NPP service version")).toBe("03");
    expect(() => definitionOf(rent.element(), "Biller code")).toThrow("no value for Biller code");

    const salary = table().getByRole("row").nth(2);
    await expect.element(salary.getByRole("cell").first()).toBeVisible();
    expect(salary.getByRole("cell").first().element().textContent).toContain(
      "15 October 2026, 9:00 am UTC+8",
    );
    await expect.element(salary.getByRole("cell", { name: "PENDING" })).toBeVisible();
    await expect.element(salary.getByRole("cell", { name: "$2,500.00" })).toBeVisible();
    const salaryValue = (term: string) => definitionOf(salary.element(), term).textContent;
    expect(salaryValue("Biller code")).toBe("12345");
    expect(salaryValue("Biller name")).toBe("Example Biller");
    expect(salaryValue("CRN")).toBe("987654");
    expect(salaryValue("APCA number")).toBe("123456");
    expect(() => definitionOf(salary.element(), "Reference")).toThrow("no value for Reference");

    await expect.element(page.getByText("30 records")).toBeVisible();
    await expect.element(page.getByText("Page 1 of 2")).toBeVisible();
    expect(requests).toEqual([`12345678 ${accountId} ?pageSize=25`]);
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(window.innerWidth);
    await expectNoAxeViolations(document.body);
  });

  test("the biller code, the CRN and the APCA number are identifiers that name the transaction", async () => {
    const writeText = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
    serve({ data: [pending], totalRecords: 1 });

    renderApp(url);

    const salary = table().getByRole("row").nth(1);
    const copy = (name: string) =>
      salary.getByRole("button", { name: `Copy ${name} of Salary` }).click();
    await copy("biller code");
    expect(writeText).toHaveBeenLastCalledWith("12345");
    await copy("CRN");
    expect(writeText).toHaveBeenLastCalledWith("987654");
    await copy("APCA number");
    expect(writeText).toHaveBeenLastCalledWith("123456");
  });

  test("gives the page of the URL to the API and keeps the next page in the URL", async () => {
    const requests = serve({ data: [transaction("transaction-1")], totalRecords: 60 });

    const router = renderApp(`${url}?page=2`);

    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();
    expect(requests).toEqual([`12345678 ${accountId} ?page=2&pageSize=25`]);

    await userEvent.click(page.getByRole("button", { name: "Next page" }));

    await expect.poll(() => router.state.location.searchStr).toBe("?page=3");
    await expect.poll(() => requests).toContain(`12345678 ${accountId} ?page=3&pageSize=25`);
  });

  test("an account with no transactions says so", async () => {
    serve({ data: [], totalRecords: 0 });

    renderApp(url);

    await expect.element(page.getByText("No transactions on this page")).toBeVisible();
  });

  test("links to the detail of the customer", async () => {
    serve({ data: [], totalRecords: 0 });
    const router = renderApp(url);

    await userEvent.click(page.getByRole("link", { name: "Customer", exact: true }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Alex Example" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe("/customers/12345678");
  });

  test("shows the title and the detail of the problem for an unknown account", async () => {
    serve({ data: [], totalRecords: 0 });
    worker.use(
      http.get(
        "/api/customers/{customerNumber}/accounts/{accountId}/transactions",
        ({ response }) =>
          response(404).json({
            type: "about:blank",
            title: "Not Found",
            status: 404,
            detail: "The customer has no account with this account ID.",
          }),
      ),
    );

    renderApp(`/customers/12345678/accounts/unknown/transactions`);

    const alert = page.getByRole("alert");
    await expect
      .element(alert.getByText("The customer has no account with this account ID."))
      .toBeVisible();
    await expect
      .element(page.getByRole("heading", { level: 1, name: "Transactions" }))
      .toBeVisible();
  });
});
