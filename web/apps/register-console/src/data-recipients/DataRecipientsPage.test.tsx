import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { http, recipient, renderApp, worker } from "../test/app";

// Each request of the page, as the search of its URL.
function serve(body: { data: ReturnType<typeof recipient>[]; totalRecords: number }) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/data-recipients", ({ request, response }) => {
      requests.push(new URL(request.url).search);
      return response(200).json(body);
    }),
  );
  return requests;
}

const table = () => page.getByRole("table", { name: "Data recipients" });

describe("the data recipients page", () => {
  // Verifies: REG-1002
  test("shows each data recipient of the page, with its counts and the total", async () => {
    serve({
      data: [recipient("Mock Budgeting Pty Ltd"), recipient("Other Pty Ltd", "SUSPENDED")],
      totalRecords: 27,
    });

    renderApp("/data-recipients");

    await expect
      .element(table().getByRole("cell", { name: "Mock Budgeting Pty Ltd" }))
      .toBeVisible();
    await expect.element(table().getByText("SUSPENDED")).toBeVisible();
    await expect.element(page.getByText("27 records")).toBeVisible();
    await expect.element(page.getByText("Page 1 of 2")).toBeVisible();
    await expect
      .element(page.getByRole("button", { name: "Copy legal entity ID of Mock Budgeting Pty Ltd" }))
      .toBeVisible();
  });

  test("/ goes to the data recipients", async () => {
    serve({ data: [], totalRecords: 0 });

    const router = renderApp("/");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Data recipients" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe("/data-recipients");
    // The request ends before the next test starts.
    await expect.element(page.getByText("No data recipients match")).toBeVisible();
  });

  // Verifies: REG-1052
  test("a URL gives the API the same filter, search, sort and page", async () => {
    const requests = serve({ data: [recipient("Mock Budgeting Pty Ltd")], totalRecords: 60 });

    renderApp("/data-recipients?status=SUSPENDED&q=mock&sort=status&order=desc&page=2");

    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();
    await expect
      .element(page.getByRole("searchbox", { name: "Search by name" }))
      .toHaveValue("mock");
    await expect.element(page.getByRole("combobox", { name: "Status" })).toHaveValue("SUSPENDED");
    await expect
      .element(page.getByRole("columnheader", { name: "Status" }))
      .toHaveAttribute("aria-sort", "descending");
    expect(requests).toEqual([
      "?status=SUSPENDED&q=mock&sort=status&order=desc&page=2&pageSize=25",
    ]);
  });

  test("a value in the URL that is not valid is not sent", async () => {
    const requests = serve({ data: [], totalRecords: 0 });

    renderApp("/data-recipients?status=DELETED&sort=lastUpdated&order=up&page=one&q=");

    await expect.element(page.getByText("No data recipients match")).toBeVisible();
    expect(requests).toEqual(["?pageSize=25"]);
  });

  test("a search, a filter, a sort and a page change the URL", async () => {
    const requests = serve({ data: [recipient("Mock Budgeting Pty Ltd")], totalRecords: 60 });
    const router = renderApp("/data-recipients?page=2");
    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();

    await userEvent.fill(
      page.getByRole("searchbox", { name: "Search by name" }).element(),
      "budget",
    );
    await expect.poll(() => router.state.location.searchStr).toBe("?q=budget");

    await userEvent.selectOptions(
      page.getByRole("combobox", { name: "Status" }).element(),
      "REVOKED",
    );
    await expect.poll(() => router.state.location.searchStr).toBe("?status=REVOKED&q=budget");

    await userEvent.click(page.getByRole("button", { name: "Accreditation number" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?status=REVOKED&q=budget&sort=accreditationNumber&order=asc");

    await userEvent.click(page.getByRole("button", { name: "Next page" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?status=REVOKED&q=budget&sort=accreditationNumber&order=asc&page=2");
    await expect
      .poll(() => requests.at(-1))
      .toBe("?status=REVOKED&q=budget&sort=accreditationNumber&order=asc&page=2&pageSize=25");
  });

  // Verifies: REG-1053, REG-1054
  test("a typed search applies with no click and adds no history entry for each pause", async () => {
    const requests = serve({ data: [recipient("Mock Budgeting Pty Ltd")], totalRecords: 1 });
    const router = renderApp("/data-recipients");
    await expect.element(page.getByText("Page 1 of 1")).toBeVisible();
    const historyLength = router.history.length;
    const searchbox = page.getByRole("searchbox", { name: "Search by name" });

    await userEvent.type(searchbox, "budget");

    await expect.poll(() => router.state.location.searchStr).toBe("?q=budget");
    // The filter bar tests prove one search for each pause. A slow machine can make a pause here.
    await expect.poll(() => requests.at(-1)).toBe("?q=budget&pageSize=25");
    expect(router.history.length).toBe(historyLength);
    expect(document.activeElement).toBe(searchbox.element());
    await expect.element(page.getByRole("button", { name: "Search" })).not.toBeInTheDocument();
  });

  test("the clear button shows the list with no search and puts the focus in the field", async () => {
    const requests = serve({ data: [recipient("Mock Budgeting Pty Ltd")], totalRecords: 1 });
    const router = renderApp("/data-recipients?q=budget");
    const searchbox = page.getByRole("searchbox", { name: "Search by name" });
    await expect.element(searchbox).toHaveValue("budget");

    await userEvent.click(page.getByRole("button", { name: "Clear search" }));

    await expect.poll(() => router.state.location.searchStr).toBe("");
    await expect.poll(() => requests.at(-1)).toBe("?pageSize=25");
    await expect.element(searchbox).toHaveValue("");
    expect(document.activeElement).toBe(searchbox.element());
  });

  test("back and forward show the search of the URL in the field", async () => {
    serve({ data: [recipient("Mock Budgeting Pty Ltd")], totalRecords: 1 });
    const router = renderApp("/data-recipients?q=mock");
    const searchbox = page.getByRole("searchbox", { name: "Search by name" });
    await expect.element(searchbox).toHaveValue("mock");

    await router.navigate({ to: "/data-recipients", search: { q: "other" } });
    await expect.element(searchbox).toHaveValue("other");
    router.history.back();
    await expect.element(searchbox).toHaveValue("mock");
    router.history.forward();
    await expect.element(searchbox).toHaveValue("other");
  });

  test("the brand and software product counts sort in both directions", async () => {
    const requests = serve({ data: [recipient("Mock Budgeting Pty Ltd")], totalRecords: 1 });
    const router = renderApp("/data-recipients");
    await expect.element(page.getByText("Page 1 of 1")).toBeVisible();

    await userEvent.click(page.getByRole("button", { name: "Brands" }));
    await expect.poll(() => router.state.location.searchStr).toBe("?sort=brandCount&order=asc");
    await expect
      .element(page.getByRole("columnheader", { name: "Brands" }))
      .toHaveAttribute("aria-sort", "ascending");
    await userEvent.click(page.getByRole("button", { name: "Brands" }));
    await expect.poll(() => router.state.location.searchStr).toBe("?sort=brandCount&order=desc");

    await userEvent.click(page.getByRole("button", { name: "Software products" }));
    await expect
      .poll(() => requests.at(-1))
      .toBe("?sort=softwareProductCount&order=asc&pageSize=25");
  });

  test("shows an empty state when no data recipient matches", async () => {
    serve({ data: [], totalRecords: 0 });

    renderApp("/data-recipients?q=unknown");

    await expect.element(page.getByText("No data recipients match")).toBeVisible();
    await expect.element(page.getByText("Change the search or the status filter.")).toBeVisible();
  });

  test("shows the title and the detail of a problem from the API", async () => {
    worker.use(
      http.get("/api/data-recipients", ({ response }) =>
        response(400).json({
          type: "about:blank",
          title: "Bad Request",
          status: 400,
          detail: "pageSize must be 100 or less",
        }),
      ),
    );

    renderApp("/data-recipients");

    const alert = page.getByRole("alert");
    await expect.element(alert.getByText("Bad Request")).toBeVisible();
    await expect.element(alert.getByText("pageSize must be 100 or less")).toBeVisible();
  });

  test("shows a problem when the API cannot answer", async () => {
    worker.use(http.untyped.get("/api/data-recipients", () => Response.error()));

    renderApp("/data-recipients");

    await expect
      .element(page.getByRole("alert").getByText("The console cannot get data from the Register"))
      .toBeVisible();
  });

  // Verifies: REG-1060
  test("passes the automated accessibility checks", async () => {
    serve({
      data: [recipient("Mock Budgeting Pty Ltd"), recipient("Other Pty Ltd", "REVOKED")],
      totalRecords: 2,
    });

    renderApp("/data-recipients");
    await expect.element(table().getByRole("cell", { name: "Other Pty Ltd" })).toBeVisible();

    await expectNoAxeViolations(document.body);
  });
});

describe("a page that does not exist", () => {
  test("says so and links to the data recipients", async () => {
    renderApp("/unknown");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Page not found" }))
      .toBeVisible();
    await expect
      .element(page.getByRole("link", { name: "Go to the data recipients" }))
      .toHaveAttribute("href", "/data-recipients");
  });
});
