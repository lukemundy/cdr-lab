import { useState } from "react";
import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { DataTable, dataTableColumns, type DataTableProps, type Sort } from "./data-table";
import { StatusBadge } from "./status-badge";

type Recipient = { id: string; name: string; status: string; brands: number };

const column = dataTableColumns<Recipient>();
const columns = column.columns([
  column.accessor("name", { header: "Name", enableSorting: true }),
  column.accessor("status", {
    header: "Status",
    enableSorting: true,
    cell: ({ getValue }) => <StatusBadge>{getValue()}</StatusBadge>,
  }),
  column.accessor("brands", { header: "Brands" }),
]);

const rows: Recipient[] = [
  { id: "r1", name: "Mock Recipient One", status: "ACTIVE", brands: 2 },
  { id: "r2", name: "Mock Recipient Two", status: "SUSPENDED", brands: 1 },
];

function renderTable(props: Partial<DataTableProps<Recipient>> = {}) {
  const calls = { sorts: [] as Sort[], pages: [] as number[] };
  const container = render(
    <DataTable
      label="Data recipients"
      columns={columns}
      rows={rows}
      getRowId={(row) => row.id}
      empty={{ title: "No data recipients" }}
      sort={{ column: "name", direction: "ascending" }}
      onSortChange={(sort) => calls.sorts.push(sort)}
      page={1}
      pageCount={3}
      rowCount={55}
      onPageChange={(p) => calls.pages.push(p)}
      {...props}
    />,
  );
  return { container, calls };
}

// Verifies: UI-0260
test("the table has a name, column headers and a row for each record", async () => {
  const { container } = renderTable();

  const table = page.getByRole("table", { name: "Data recipients" });
  await expect.element(table).toBeVisible();
  expect(
    page
      .getByRole("columnheader")
      .elements()
      .map((th) => th.textContent),
  ).toEqual(["Name", "Status", "Brands"]);
  await expect
    .element(page.getByRole("row", { name: /Mock Recipient Two/ }))
    .toMatchTextContent("SUSPENDED");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0265
test("the sorted column has aria-sort, and the other columns have none", async () => {
  renderTable({ sort: { column: "status", direction: "descending" } });

  await expect
    .element(page.getByRole("columnheader", { name: "Status" }))
    .toHaveAttribute("aria-sort", "descending");
  await expect
    .element(page.getByRole("columnheader", { name: "Name" }))
    .not.toHaveAttribute("aria-sort");
});

// Verifies: UI-0265
test("a status message gives the sort", async () => {
  renderTable({ sort: { column: "name", direction: "ascending" } });

  await expect
    .element(page.getByRole("status").filter({ hasText: "Sorted" }))
    .toHaveTextContent("Sorted by Name, ascending");
});

// Verifies: UI-0264
test("a sort button sorts an unsorted column ascending, and turns a sorted column the other way", async () => {
  const { calls } = renderTable({ sort: { column: "name", direction: "ascending" } });

  await page.getByRole("button", { name: "Status" }).click();
  await page.getByRole("button", { name: "Name" }).click();

  expect(calls.sorts).toEqual([
    { column: "status", direction: "ascending" },
    { column: "name", direction: "descending" },
  ]);
});

// Verifies: UI-0263
test("a column that does not enable sorting has no sort button", async () => {
  renderTable();

  await expect.element(page.getByRole("columnheader", { name: "Brands" })).toBeVisible();
  expect(page.getByRole("button", { name: "Brands" }).elements()).toEqual([]);
});

// Verifies: UI-0201, UI-0202
test("the keyboard sorts with a visible focus", async () => {
  const { calls } = renderTable();

  await userEvent.tab();
  expectVisibleFocus(page.getByRole("button", { name: "Name" }).element());
  await userEvent.keyboard("{Enter}");

  expect(calls.sorts).toEqual([{ column: "name", direction: "descending" }]);
});

// Verifies: UI-0094
test("each sort button is 24 by 24 CSS px or more", () => {
  renderTable();

  expectTargetSize(page.getByRole("button", { name: "Name" }).element());
  expectTargetSize(page.getByRole("button", { name: "Status" }).element());
});

// Verifies: UI-0262
test("the pagination is below the table and changes the page", async () => {
  const { container, calls } = renderTable();

  const table = page.getByRole("table").element();
  const navigation = page.getByRole("navigation", { name: "Pagination" }).element();
  expect(table.compareDocumentPosition(navigation) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  await page.getByRole("button", { name: "Next page" }).click();

  expect(calls.pages).toEqual([2]);
  expect(container).toContainElement(navigation);
});

// Verifies: UI-0262
test("with no rows, the table shows the empty state", async () => {
  const { container } = renderTable({ rows: [], rowCount: 0, pageCount: 0 });

  await expect.element(page.getByRole("table")).toMatchTextContent("No data recipients");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0262
test("while it loads, the table is busy and shows the loading state", async () => {
  const { container } = renderTable({ status: { state: "loading" } });

  await expect.element(page.getByRole("table")).toHaveAttribute("aria-busy", "true");
  await expect.element(page.getByRole("status").filter({ hasText: "Loading" })).toBeVisible();
  expect(page.getByText("Mock Recipient One").elements()).toEqual([]);
  await expectNoAxeViolations(container);
});

// Verifies: UI-0262
test("after an error, the table shows the problem in the error state", async () => {
  const { container } = renderTable({
    status: { state: "error", problem: { title: "Service Unavailable", detail: "Try again." } },
  });

  await expect.element(page.getByRole("alert")).toMatchTextContent("Service Unavailable");
  await expect.element(page.getByRole("alert")).toMatchTextContent("Try again.");
  await expect.element(page.getByRole("table")).not.toHaveAttribute("aria-busy");
  await expectNoAxeViolations(container);
});

function Harness() {
  const [current, setCurrent] = useState(rows);
  const [sort, setSort] = useState<Sort>({ column: "name", direction: "ascending" });
  return (
    <>
      <button type="button" onClick={() => setCurrent(rows.toReversed().slice(0, 1))}>
        Change the rows
      </button>
      <DataTable
        label="Data recipients"
        columns={columns}
        rows={current}
        getRowId={(row) => row.id}
        empty={{ title: "No data recipients" }}
        sort={sort}
        onSortChange={setSort}
        page={1}
        pageCount={1}
        rowCount={current.length}
        onPageChange={() => {}}
      />
    </>
  );
}

test("the table shows new rows and a new sort when the app gives them", async () => {
  render(<Harness />);

  await page.getByRole("button", { name: "Change the rows" }).click();
  await page.getByRole("button", { name: "Name" }).click();

  await expect.element(page.getByText("Mock Recipient Two")).toBeVisible();
  expect(page.getByText("Mock Recipient One").elements()).toEqual([]);
  await expect
    .element(page.getByRole("columnheader", { name: "Name" }))
    .toHaveAttribute("aria-sort", "descending");
});
