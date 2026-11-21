import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { CustomerSummary } from "../api/client";
import { customer, http, renderApp, worker } from "../test/app";

// Each request of the page, as the search of its URL.
function serve(body: { data: CustomerSummary[]; totalRecords: number }) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/customers", ({ request, response }) => {
      requests.push(new URL(request.url).search);
      return response(200).json(body);
    }),
  );
  return requests;
}

const table = () => page.getByRole("table", { name: "Customers" });
const searchbox = () =>
  page.getByRole("searchbox", { name: "Search by customer number, name or email" });

describe("the customers page", () => {
  // Verifies: DH-1413
  test("shows each customer of the page, with the counts and the total", async () => {
    serve({
      data: [
        customer("12345678", { openAccountCount: 2, currentArrangementCount: 3 }),
        {
          customerNumber: "87654321",
          lastName: "Sample",
          emailAddress: "87654321@mail.test",
          lastUpdated: "2026-10-15T05:30:00Z",
          openAccountCount: 0,
          currentArrangementCount: 0,
        },
      ],
      totalRecords: 27,
    });

    renderApp("/customers");

    const first = table().getByRole("row").nth(1);
    await expect.element(first.getByRole("cell", { name: "12345678" })).toBeVisible();
    await expect.element(first.getByRole("cell", { name: "Example, Alex" })).toBeVisible();
    await expect.element(first.getByRole("cell", { name: "12345678@mail.test" })).toBeVisible();
    await expect.element(first.getByRole("cell", { name: "2", exact: true })).toBeVisible();
    await expect.element(first.getByRole("cell", { name: "3", exact: true })).toBeVisible();
    await expect.element(table().getByRole("cell", { name: "Sample", exact: true })).toBeVisible();
    await expect.element(page.getByText("27 records")).toBeVisible();
    await expect.element(page.getByText("Page 1 of 2")).toBeVisible();
  });

  test("a URL gives the API the same search, sort and page", async () => {
    const requests = serve({ data: [customer("12345678")], totalRecords: 60 });

    renderApp("/customers?q=exam&sort=lastUpdated&order=desc&page=2");

    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();
    await expect.element(searchbox()).toHaveValue("exam");
    await expect
      .element(page.getByRole("columnheader", { name: "Last changed" }))
      .toHaveAttribute("aria-sort", "descending");
    expect(requests).toEqual(["?q=exam&sort=lastUpdated&order=desc&page=2&pageSize=25"]);
  });

  test("a value in the URL that is not valid is not sent", async () => {
    const requests = serve({ data: [], totalRecords: 0 });

    renderApp("/customers?status=OPEN&sort=emailAddress&order=up&page=one&q=");

    await expect.element(page.getByText("No customers match")).toBeVisible();
    expect(requests).toEqual(["?pageSize=25"]);
  });

  // Verifies: DH-1474, WCON-0001
  test("a search, a sort and a page change the URL, and a search adds no history entry", async () => {
    const requests = serve({ data: [customer("12345678")], totalRecords: 60 });
    const router = renderApp("/customers?page=2");
    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();
    const historyLength = router.history.length;

    await userEvent.fill(searchbox().element(), "alex@");
    await expect.poll(() => router.state.location.searchStr).toBe("?q=alex%40");
    expect(router.history.length).toBe(historyLength);

    await userEvent.click(page.getByRole("button", { name: "Customer number" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?q=alex%40&sort=customerNumber&order=asc");

    await userEvent.click(page.getByRole("button", { name: "Next page" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?q=alex%40&sort=customerNumber&order=asc&page=2");
    await expect
      .poll(() => requests.at(-1))
      .toBe("?q=alex%40&sort=customerNumber&order=asc&page=2&pageSize=25");
    expect(router.history.length).toBe(historyLength + 2);
  });

  test("the name sorts by the last name", async () => {
    const requests = serve({ data: [customer("12345678")], totalRecords: 1 });
    const router = renderApp("/customers?sort=customerNumber");
    await expect.element(page.getByText("Page 1 of 1")).toBeVisible();

    await userEvent.click(page.getByRole("button", { name: "Name" }));

    await expect.poll(() => router.state.location.searchStr).toBe("?sort=lastName&order=asc");
    await expect.poll(() => requests.at(-1)).toBe("?sort=lastName&order=asc&pageSize=25");
  });

  test("the open account and current arrangement counts sort in both directions", async () => {
    const requests = serve({ data: [customer("12345678")], totalRecords: 1 });
    const router = renderApp("/customers");
    await expect.element(page.getByText("Page 1 of 1")).toBeVisible();

    await userEvent.click(page.getByRole("button", { name: "Open accounts" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?sort=openAccountCount&order=asc");
    await expect
      .element(page.getByRole("columnheader", { name: "Open accounts" }))
      .toHaveAttribute("aria-sort", "ascending");
    await userEvent.click(page.getByRole("button", { name: "Open accounts" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?sort=openAccountCount&order=desc");

    await userEvent.click(page.getByRole("button", { name: "Current arrangements" }));
    await expect
      .poll(() => requests.at(-1))
      .toBe("?sort=currentArrangementCount&order=asc&pageSize=25");
  });

  test("shows a problem when the API cannot answer", async () => {
    worker.use(http.untyped.get("/api/customers", () => Response.error()));

    renderApp("/customers");

    await expect
      .element(
        page.getByRole("alert").getByText("The console cannot get data from the data holder"),
      )
      .toBeVisible();
  });

  // Verifies: DH-1479
  test("passes the automated accessibility checks", async () => {
    serve({ data: [customer("12345678"), customer("87654321")], totalRecords: 2 });

    renderApp("/customers");
    await expect.element(table().getByRole("cell", { name: "87654321" })).toBeVisible();

    await expectNoAxeViolations(document.body);
  });
});

describe("a page that does not exist", () => {
  test("says so and links to the customers", async () => {
    renderApp("/unknown");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Page not found" }))
      .toBeVisible();
    await expect
      .element(page.getByRole("link", { name: "Go to the customers" }))
      .toHaveAttribute("href", "/customers");
  });
});
