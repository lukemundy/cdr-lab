import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { holder, http, renderApp, worker } from "../test/app";

// Each request of the page, as the search of its URL.
function serve(body: { data: ReturnType<typeof holder>[]; totalRecords: number }) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/data-holders", ({ request, response }) => {
      requests.push(new URL(request.url).search);
      return response(200).json(body);
    }),
  );
  return requests;
}

const table = () => page.getByRole("table", { name: "Data holders" });
const searchbox = () =>
  page.getByRole("searchbox", { name: "Search by legal entity or brand name" });

describe("the data holders page", () => {
  test("shows each data holder of the page, with its brands, industries and the total", async () => {
    serve({
      data: [
        holder("Mock Holdings Ltd", { brandCount: 3, industries: ["banking", "energy", "telco"] }),
        holder("Former Bank Ltd", { status: "REMOVED" }),
        holder("Empty Ltd", { brandCount: 0, industries: [] }),
      ],
      totalRecords: 27,
    });

    renderApp("/data-holders");

    const mock = table().getByRole("row", { name: /Mock Holdings Ltd/ });
    await expect.element(mock.getByRole("cell", { name: "banking, energy, telco" })).toBeVisible();
    await expect.element(mock.getByRole("cell", { name: "3", exact: true })).toBeVisible();
    await expect.element(table().getByText("REMOVED")).toBeVisible();
    await expect
      .element(
        table()
          .getByRole("row", { name: /Empty Ltd/ })
          .getByRole("cell", { name: "0", exact: true }),
      )
      .toBeVisible();
    await expect.element(page.getByText("27 records")).toBeVisible();
    await expect.element(page.getByText("Page 1 of 2")).toBeVisible();
    await expect
      .element(page.getByRole("button", { name: "Copy legal entity ID of Mock Holdings Ltd" }))
      .toBeVisible();
  });

  test("the navigation goes to the data holders", async () => {
    serve({ data: [], totalRecords: 0 });
    worker.use(
      http.get("/api/data-recipients", ({ response }) =>
        response(200).json({ data: [], totalRecords: 0 }),
      ),
    );
    const router = renderApp("/data-recipients");

    await userEvent.click(page.getByRole("link", { name: "Data holders" }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Data holders" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe("/data-holders");
    await expect
      .element(page.getByRole("link", { name: "Data holders" }))
      .toHaveAttribute("aria-current", "page");
    await expect.element(page.getByText("No data holders match")).toBeVisible();
  });

  test("a URL gives the API the same filter, search, sort and page", async () => {
    const requests = serve({ data: [holder("Mock Holdings Ltd")], totalRecords: 60 });

    renderApp("/data-holders?status=REMOVED&q=mutual&sort=status&order=desc&page=2");

    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();
    await expect.element(searchbox()).toHaveValue("mutual");
    await expect.element(page.getByRole("combobox", { name: "Status" })).toHaveValue("REMOVED");
    await expect
      .element(page.getByRole("columnheader", { name: "Status" }))
      .toHaveAttribute("aria-sort", "descending");
    expect(requests).toEqual([
      "?status=REMOVED&q=mutual&sort=status&order=desc&page=2&pageSize=25",
    ]);
  });

  test("a value in the URL that is not valid is not sent", async () => {
    const requests = serve({ data: [], totalRecords: 0 });

    renderApp("/data-holders?status=SUSPENDED&sort=accreditationNumber&order=up&page=one&q=");

    await expect.element(page.getByText("No data holders match")).toBeVisible();
    expect(requests).toEqual(["?pageSize=25"]);
  });

  test("a search, a filter, a sort and a page change the URL", async () => {
    const requests = serve({ data: [holder("Mock Holdings Ltd")], totalRecords: 60 });
    const router = renderApp("/data-holders?page=2");
    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();

    await userEvent.fill(searchbox().element(), "mutual");
    await expect.poll(() => router.state.location.searchStr).toBe("?q=mutual");

    await userEvent.selectOptions(
      page.getByRole("combobox", { name: "Status" }).element(),
      "ACTIVE",
    );
    await expect.poll(() => router.state.location.searchStr).toBe("?status=ACTIVE&q=mutual");

    await userEvent.click(page.getByRole("button", { name: "Status" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?status=ACTIVE&q=mutual&sort=status&order=asc");

    await userEvent.click(page.getByRole("button", { name: "Next page" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?status=ACTIVE&q=mutual&sort=status&order=asc&page=2");
    await expect
      .poll(() => requests.at(-1))
      .toBe("?status=ACTIVE&q=mutual&sort=status&order=asc&page=2&pageSize=25");
  });

  test("a typed search applies with no click and adds no history entry", async () => {
    const requests = serve({ data: [holder("Mock Holdings Ltd")], totalRecords: 1 });
    const router = renderApp("/data-holders");
    await expect.element(page.getByText("Page 1 of 1")).toBeVisible();
    const historyLength = router.history.length;

    await userEvent.type(searchbox(), "mutual");

    await expect.poll(() => router.state.location.searchStr).toBe("?q=mutual");
    // The filter bar tests prove one search for each pause. A slow machine can make a pause here.
    await expect.poll(() => requests.at(-1)).toBe("?q=mutual&pageSize=25");
    expect(router.history.length).toBe(historyLength);
    expect(document.activeElement).toBe(searchbox().element());
  });

  test("the brand count sorts in both directions", async () => {
    const requests = serve({ data: [holder("Mock Holdings Ltd")], totalRecords: 1 });
    const router = renderApp("/data-holders");
    await expect.element(page.getByText("Page 1 of 1")).toBeVisible();

    await userEvent.click(page.getByRole("button", { name: "Brands" }));
    await expect.poll(() => router.state.location.searchStr).toBe("?sort=brandCount&order=asc");
    await userEvent.click(page.getByRole("button", { name: "Brands" }));
    await expect.poll(() => requests.at(-1)).toBe("?sort=brandCount&order=desc&pageSize=25");
    await expect
      .element(page.getByRole("columnheader", { name: "Brands" }))
      .toHaveAttribute("aria-sort", "descending");
  });

  test("shows an empty state when no data holder matches", async () => {
    serve({ data: [], totalRecords: 0 });

    renderApp("/data-holders?q=unknown");

    await expect.element(page.getByText("No data holders match")).toBeVisible();
    await expect.element(page.getByText("Change the search or the status filter.")).toBeVisible();
  });

  test("shows the title and the detail of a problem from the API", async () => {
    worker.use(
      http.get("/api/data-holders", ({ response }) =>
        response(400).json({
          type: "about:blank",
          title: "Bad Request",
          status: 400,
          detail: "pageSize must be 100 or less",
        }),
      ),
    );

    renderApp("/data-holders");

    const alert = page.getByRole("alert");
    await expect.element(alert.getByText("Bad Request")).toBeVisible();
    await expect.element(alert.getByText("pageSize must be 100 or less")).toBeVisible();
  });

  test("passes the automated accessibility checks", async () => {
    serve({
      data: [holder("Mock Holdings Ltd"), holder("Former Bank Ltd", { status: "REMOVED" })],
      totalRecords: 2,
    });

    renderApp("/data-holders");
    await expect.element(table().getByRole("cell", { name: "Former Bank Ltd" })).toBeVisible();

    await expectNoAxeViolations(document.body);
  });
});
