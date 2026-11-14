import {
  createColumnHelper,
  rowSortingFeature,
  tableFeatures,
  useTable,
  type RowData,
  type TableOptions,
} from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";
import { Pagination, type PaginationProps } from "./pagination";
import { EmptyState, ErrorState, LoadingState, type ListStatus } from "./states";

const features = tableFeatures({ rowSortingFeature });

// A column can sort only when its definition sets enableSorting, because the server must support the sort.
export function dataTableColumns<Row extends RowData>() {
  return createColumnHelper<typeof features, Row>();
}

export type Sort = { column: string; direction: "ascending" | "descending" };

export type DataTableStatus = ListStatus;

// The sort and the page are controlled: the app keeps them, for example in the URL, and the callbacks ask for a change.
export type DataTableProps<Row extends RowData> = PaginationProps & {
  label: string;
  columns: TableOptions<typeof features, Row>["columns"];
  rows: Row[];
  getRowId: (row: Row) => string;
  empty: { title: string; description?: string };
  status?: DataTableStatus;
  sort?: Sort;
  onSortChange?: (sort: Sort) => void;
};

const ready: DataTableStatus = { state: "ready" };

export function DataTable<Row extends RowData>({
  label,
  columns,
  rows,
  getRowId,
  empty,
  status = ready,
  sort,
  onSortChange,
  ...pagination
}: DataTableProps<Row>) {
  const table = useTable({
    features,
    columns,
    data: rows,
    getRowId,
    manualSorting: true,
    defaultColumn: { enableSorting: false },
    state: { sorting: sort ? [{ id: sort.column, desc: sort.direction === "descending" }] : [] },
  });

  const headers = table.getHeaderGroups().flatMap((group) => group.headers);
  const sortedHeader = headers.find((header) => header.column.id === sort?.column);
  const tableRows = table.getRowModel().rows;

  function body() {
    if (status.state === "loading") return <LoadingState />;
    if (status.state === "error") return <ErrorState problem={status.problem} />;
    if (tableRows.length === 0) return <EmptyState {...empty} />;
    return null;
  }
  const message = body();

  return (
    <div>
      <table
        aria-busy={status.state === "loading" || undefined}
        className="w-full border-collapse text-sm tabular-nums"
      >
        <caption className="sr-only">{label}</caption>
        <thead>
          <tr className="border-b border-border-strong">
            {headers.map((header) => {
              const direction = sort?.column === header.column.id ? sort.direction : undefined;
              return (
                <th
                  key={header.id}
                  scope="col"
                  aria-sort={direction}
                  className="h-table-row px-cell-x text-left font-medium text-fg-muted"
                >
                  {header.column.getCanSort() ? (
                    <button
                      type="button"
                      onClick={() =>
                        onSortChange?.({
                          column: header.column.id,
                          direction: direction === "ascending" ? "descending" : "ascending",
                        })
                      }
                      className={cn(
                        "-mx-1 inline-flex min-h-6 items-center gap-1 rounded-sm px-1 hover:text-fg",
                        focusRing,
                      )}
                    >
                      <table.FlexRender header={header} />
                      <SortIcon direction={direction} />
                    </button>
                  ) : (
                    <table.FlexRender header={header} />
                  )}
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody>
          {message ? (
            <tr>
              <td colSpan={headers.length} className="p-cell-x">
                {message}
              </td>
            </tr>
          ) : (
            tableRows.map((row) => (
              <tr key={row.id} className="border-b border-border">
                {row.getAllCells().map((cell) => (
                  // oxlint-disable-next-line jsx-a11y/control-has-associated-label -- a cell is not a control, and its content comes from the column definition.
                  <td key={cell.id} className="h-table-row px-cell-x py-cell-y text-fg">
                    <table.FlexRender cell={cell} />
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
      <output className="sr-only">
        {sort &&
          sortedHeader &&
          `Sorted by ${headerText(sortedHeader.column.columnDef.header, sort.column)}, ${sort.direction}`}
      </output>
      <Pagination {...pagination} />
    </div>
  );
}

function headerText(header: unknown, fallback: string): string {
  return typeof header === "string" ? header : fallback;
}

function SortIcon({ direction }: { direction: Sort["direction"] | undefined }) {
  if (direction === "ascending") return <ArrowUp aria-hidden="true" className="size-4" />;
  if (direction === "descending") return <ArrowDown aria-hidden="true" className="size-4" />;
  return <ArrowUpDown aria-hidden="true" className="size-4 opacity-60" />;
}
