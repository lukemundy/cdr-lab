import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { afterEach, describe, expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { CustomerDetail } from "../api/client";
import {
  account,
  arrangement,
  customer,
  customerDetail,
  definitionOf,
  http,
  registration,
  renderApp,
  worker,
} from "../test/app";

const accountId = "6f1c1a47-3c2e-4a0b-9d55-3f8f4f1e2a10";
const savingsId = "0b3e7c55-8f0e-4d6b-a0c4-7d7f4a3b9e21";

const detail = customerDetail("12345678", {
  accounts: [
    account(accountId, {
      creditLimit: "100.00",
      currentBalance: "2500.00",
      availableBalance: "2595.50",
    }),
    account(savingsId, {
      displayName: "Saver",
      accountNumber: "87654321",
      currentBalance: "-12.30",
      openStatus: "CLOSED",
      actions: [],
    }),
  ],
  arrangements: [
    arrangement("arrangement-1", { accountIds: [accountId, savingsId] }),
    arrangement("arrangement-2", {
      clientName: "Recipient Mobile",
      clientId: "client-2",
      clientRegistered: false,
      scopes: ["openid"],
      status: "ENDED",
      endedAt: "2026-10-16T01:00:00Z",
      endReason: "WITHDRAWN_BY_CONSUMER",
      withdrawalMethod: "ALTERNATIVE_METHOD",
      noticeStatus: "FAILED",
      noticeAttempts: [
        { attempt: 1, at: "2026-10-16T01:00:01Z", result: "status 503" },
        { attempt: 2, at: "2026-10-16T01:00:11Z", result: "timeout" },
        { attempt: 3, at: "2026-10-16T01:01:11Z", result: "tls error" },
      ],
      actions: [],
    }),
    arrangement("arrangement-3", {
      clientName: "Recipient Desktop",
      status: "ENDED",
      endedAt: "2026-10-16T02:00:00Z",
      endReason: "CONSUMER_NOT_ELIGIBLE",
      noticeStatus: "PENDING",
      actions: [],
      noticeAttempts: [{ attempt: 1, at: "2026-10-16T02:00:01Z", result: "no connection" }],
    }),
  ],
});

// Each request of the page, as the customer number of its path.
function serve(body: CustomerDetail) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/customers/{customerNumber}", ({ params, response }) => {
      requests.push(params.customerNumber);
      return response(200).json(body);
    }),
  );
  return requests;
}

async function renderDetail(url: string) {
  const router = renderApp(url);
  await page.getByRole("button", { name: "Show all accounts" }).click();
  await page.getByRole("button", { name: "Show all arrangements" }).click();
  return router;
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const valueOf = (regionName: string, term: string) =>
  definitionOf(region(regionName).element(), term).textContent;

afterEach(() => {
  vi.restoreAllMocks();
});

describe("the customer page", () => {
  // Verifies: DH-1416, DH-1417, DH-1418
  test("shows the customer, each account with its balances, and each arrangement", async () => {
    const requests = serve(detail);

    await renderDetail("/customers/12345678");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Alex Example" }))
      .toBeVisible();
    expect(requests).toEqual(["12345678"]);

    expect(valueOf("Customer", "Customer number")).toContain("12345678");
    expect(valueOf("Customer", "First name")).toBe("Alex");
    expect(valueOf("Customer", "Last name")).toBe("Example");
    expect(valueOf("Customer", "Email address")).toBe("12345678@mail.test");
    // The tests run in Australia/Perth (vitest.config.ts).
    expect(valueOf("Customer", "Last changed")).toContain("15 October 2026, 1:30 pm UTC+8");

    await expect.element(region("Everyday").getByText("OPEN", { exact: true })).toBeVisible();
    expect(valueOf("Everyday", "Account ID")).toContain(accountId);
    expect(valueOf("Everyday", "Product category")).toBe("TRANS_AND_SAVINGS_ACCOUNTS");
    expect(valueOf("Everyday", "Product name")).toBe("Mock Everyday Account");
    expect(valueOf("Everyday", "BSB")).toBe("123-456");
    expect(valueOf("Everyday", "Account number")).toBe("12345678");
    expect(valueOf("Everyday", "Credit limit")).toBe("$100.00");
    expect(valueOf("Everyday", "Current balance")).toBe("$2,500.00");
    expect(valueOf("Everyday", "Available balance")).toBe("$2,595.50");
    expect(valueOf("Everyday", "Created")).toBe("15 October 2026");
    await expect.element(region("Saver").getByText("CLOSED", { exact: true })).toBeVisible();
    expect(valueOf("Saver", "Current balance")).toBe("-$12.30");

    const web = "Recipient Web";
    await expect.element(region(web).getByText("CURRENT", { exact: true })).toBeVisible();
    expect(valueOf(web, "Arrangement ID")).toContain("arrangement-1");
    expect(valueOf(web, "Client ID")).toContain("client-1");
    expect(valueOf(web, "Software ID")).toContain("software-1");
    expect(valueOf(web, "Scopes")).toBe("openidbank:accounts.basic:read");
    expect(valueOf(web, "Confirmed")).toContain("15 October 2026, 1:30 pm UTC+8");
    expect(valueOf(web, "Sharing ends")).toContain("15 October 2027, 1:30 pm UTC+8");
    const accounts = definitionOf(region(web).element(), "Accounts").querySelectorAll("li");
    expect([...accounts].map((item) => item.textContent)).toEqual([
      expect.stringContaining("Everyday"),
      expect.stringContaining("Saver"),
    ]);
    expect(accounts[1]?.textContent).toContain(savingsId);
    expect(
      [...region(web).element().querySelectorAll("dt")].map((term) => term.textContent),
    ).not.toContain("Ended");

    await expectNoAxeViolations(document.body);
  });

  // Verifies: DH-1424
  test("an arrangement that ended shows the time and the reason of its end, the method of a withdrawal, and the status of its notice", async () => {
    serve(detail);

    await renderDetail("/customers/12345678");

    const mobile = "Recipient Mobile";
    await expect.element(region(mobile).getByText("ENDED", { exact: true })).toBeVisible();
    expect(valueOf(mobile, "Ended")).toContain("16 October 2026, 9:00 am UTC+8");
    expect(valueOf(mobile, "End reason")).toBe("WITHDRAWN_BY_CONSUMER");
    expect(valueOf(mobile, "Withdrawal method")).toBe("ALTERNATIVE_METHOD");
    expect(region("Recipient Desktop").getByText("Withdrawal method").query()).toBeNull();
    expect(valueOf(mobile, "Notice status")).toBe("FAILED");
  });

  // Verifies: DH-1425
  test("an arrangement with no accounts shows None, because a consent with no bank scope has no accounts", async () => {
    serve(detail);

    await renderDetail("/customers/12345678");

    await expect.element(region("Recipient Mobile")).toBeVisible();
    expect(valueOf("Recipient Mobile", "Accounts")).toBe("None");
  });

  // Verifies: DH-1426
  test("an arrangement with a notice shows each attempt of the notice with its time and its result", async () => {
    serve(detail);

    await renderDetail("/customers/12345678");

    const attemptsOf = (regionName: string) =>
      [...definitionOf(region(regionName).element(), "Notice attempts").querySelectorAll("li")].map(
        (item) => item.textContent,
      );
    expect(attemptsOf("Recipient Mobile")).toEqual([
      expect.stringMatching(/^Attempt 1, 16 October 2026, 9:00:01 am UTC\+8.*: status 503$/),
      expect.stringMatching(/^Attempt 2, 16 October 2026, 9:00:11 am UTC\+8.*: timeout$/),
      expect.stringMatching(/^Attempt 3, 16 October 2026, 9:01:11 am UTC\+8.*: tls error$/),
    ]);
    expect(valueOf("Recipient Desktop", "Notice status")).toBe("PENDING");
    expect(attemptsOf("Recipient Desktop")).toEqual([
      expect.stringMatching(/^Attempt 1, 16 October 2026, 10:00:01 am UTC\+8.*: no connection$/),
    ]);
    expect(
      [...region("Recipient Web").element().querySelectorAll("dt")].map((term) => term.textContent),
    ).not.toContain("Notice attempts");
  });

  // Verifies: DH-1478
  test("the copy buttons of the BSB and the account number copy the digits only", async () => {
    const writeText = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
    serve(detail);

    await renderDetail("/customers/12345678");

    await region("Everyday").getByRole("button", { name: "Copy BSB of Everyday" }).click();
    expect(writeText).toHaveBeenLastCalledWith("123456");
    await region("Everyday")
      .getByRole("button", { name: "Copy account number of Everyday" })
      .click();
    expect(writeText).toHaveBeenLastCalledWith("12345678");
  });

  test("the cards of the accounts and the arrangements start collapsed", async () => {
    serve(detail);

    renderApp("/customers/12345678");

    await expect
      .element(page.getByRole("button", { name: "Everyday" }))
      .toHaveAttribute("aria-expanded", "false");
    await expect
      .element(page.getByRole("button", { name: "Recipient Web" }))
      .toHaveAttribute("aria-expanded", "false");
    await expect.element(region("Everyday").getByText("OPEN", { exact: true })).toBeVisible();
  });

  test("shows that the customer has no accounts and no arrangements", async () => {
    const noFirstName: CustomerDetail = customerDetail("12345678");
    delete noFirstName.firstName;
    serve(noFirstName);

    renderApp("/customers/12345678");

    await expect.element(page.getByRole("heading", { level: 1, name: "Example" })).toBeVisible();
    await expect.element(page.getByText("This customer has no accounts")).toBeVisible();
    await expect.element(page.getByText("This customer has no arrangements")).toBeVisible();
    expect(valueOf("Customer", "First name")).toBe("Not set");
  });

  test("shows the title and the detail of the problem for an unknown customer", async () => {
    worker.use(
      http.get("/api/customers/{customerNumber}", ({ response }) =>
        response(404).json({
          type: "about:blank",
          title: "Not Found",
          status: 404,
          detail: "The data holder has no customer with this customer number.",
        }),
      ),
    );

    renderApp("/customers/99999999");

    const alert = page.getByRole("alert");
    await expect.element(alert.getByText("Not Found")).toBeVisible();
    await expect
      .element(alert.getByText("The data holder has no customer with this customer number."))
      .toBeVisible();
    await expect.element(page.getByRole("heading", { level: 1, name: "Customer" })).toBeVisible();
    await expectNoAxeViolations(document.body);
  });

  // Verifies: DH-1423
  test("a row of the list links to the detail, and the navigation item stays current", async () => {
    serve(detail);
    worker.use(
      http.get("/api/customers", ({ response }) =>
        response(200).json({ data: [customer("12345678")], totalRecords: 1 }),
      ),
    );
    const router = renderApp("/customers");

    await userEvent.click(page.getByRole("link", { name: "12345678" }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Alex Example" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe("/customers/12345678");
    await expect
      .element(page.getByRole("link", { name: "Customers", exact: true }))
      .toHaveAttribute("aria-current", "page");
  });

  test("each account links to its transactions, and the link names the account", async () => {
    serve(detail);
    worker.use(
      http.get(
        "/api/customers/{customerNumber}/accounts/{accountId}/transactions",
        ({ response }) => response(200).json({ data: [], totalRecords: 0 }),
      ),
    );
    const router = renderApp("/customers/12345678");

    await userEvent.click(page.getByRole("link", { name: "Transactions of Saver" }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Transactions of Saver" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe(
      `/customers/12345678/accounts/${savingsId}/transactions`,
    );
  });

  // Verifies: DH-1427
  test("an arrangement links to the registration of its client only while the registration exists", async () => {
    serve(detail);
    worker.use(
      http.get("/api/registrations/{clientId}", ({ params, response }) =>
        response(200).json(registration(params.clientId)),
      ),
      http.get("/api/registrations/{clientId}/arrangements", ({ response }) =>
        response(200).json({ data: [], totalRecords: 0 }),
      ),
    );
    const router = await renderDetail("/customers/12345678");

    await expect
      .element(page.getByRole("link", { name: "Registration of Recipient Web" }))
      .toBeVisible();
    expect(
      region("Recipient Mobile").element().querySelector("a"),
      "the registration of client-2 is deleted",
    ).toBeNull();

    await userEvent.click(page.getByRole("link", { name: "Registration of Recipient Web" }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Recipient Web" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe("/registrations/client-1");
  });
});
