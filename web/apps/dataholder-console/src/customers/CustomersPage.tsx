import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { FilterBar } from "@cdr-lab/ui/filter-bar";
import { Link } from "@cdr-lab/ui/link";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { listCustomers, problemOf, type CustomerSummary } from "../api/client";
import {
  changedSearch,
  orderOf,
  pageSize,
  tableSort,
  type ListSearchChange,
} from "@cdr-lab/console/list-search";
import { apiQuery, sortColumnOf, validateCustomerSearch, type CustomerSearch } from "./search";

const route = getRouteApi("/customers");

// The name shows the last name first, because the list sorts by the last name.
function nameOf({ firstName, lastName }: CustomerSummary): string {
  return firstName ? `${lastName}, ${firstName}` : lastName;
}

const column = dataTableColumns<CustomerSummary>();
const columns = column.columns([
  column.accessor("customerNumber", {
    header: "Customer number",
    enableSorting: true,
    cell: ({ getValue }) => (
      <Link
        className="whitespace-nowrap"
        render={(props) => (
          <RouterLink
            {...props}
            to="/customers/$customerNumber"
            params={{ customerNumber: getValue() }}
          />
        )}
      >
        {getValue()}
      </Link>
    ),
  }),
  column.accessor("lastName", {
    header: "Name",
    enableSorting: true,
    cell: ({ row }) => nameOf(row.original),
  }),
  column.accessor("emailAddress", { header: "Email address" }),
  column.accessor("openAccountCount", { header: "Open accounts", enableSorting: true }),
  column.accessor("currentArrangementCount", {
    header: "Current arrangements",
    enableSorting: true,
  }),
  column.accessor("lastUpdated", {
    header: "Last changed",
    enableSorting: true,
    cell: ({ getValue }) => <LocalTime value={getValue()} />,
  }),
]);

export function CustomersPage() {
  // The router keeps the search parameters of the URL that the route does not validate, so they are removed here.
  const search = validateCustomerSearch(route.useSearch());
  const navigate = route.useNavigate();
  const list = useQuery({
    queryKey: ["customers", search],
    queryFn: ({ signal }) => listCustomers(apiQuery(search), signal),
    placeholderData: keepPreviousData,
  });

  function update(change: ListSearchChange<CustomerSearch>, { replace = false } = {}) {
    void navigate({ search: changedSearch(validateCustomerSearch, search, change), replace });
  }

  let status: DataTableStatus = { state: "ready" };
  if (list.isPending) status = { state: "loading" };
  if (list.isError) status = { state: "error", problem: problemOf(list.error) };

  const total = list.data?.totalRecords ?? 0;

  return (
    <>
      <PageHeader
        title="Customers"
        description="The customers of the brand."
        actions={
          <Link icon={Plus} render={(props) => <RouterLink {...props} to="/create/customer" />}>
            Create customer
          </Link>
        }
      />
      <FilterBar
        label="Customers"
        searchLabel="Search by customer number, name or email"
        search={search.q ?? ""}
        // A search applies while the operator types, so it replaces the history entry and does not add one for each pause.
        onSearch={(q) => update({ q: q || undefined }, { replace: true })}
      />
      <DataTable
        label="Customers"
        columns={columns}
        rows={list.data?.data ?? []}
        getRowId={(row) => row.customerNumber}
        empty={{
          title: "No customers match",
          description: "Change the search, or create a customer.",
        }}
        status={status}
        sort={tableSort(search, "lastName")}
        onSortChange={(sort) => update({ sort: sortColumnOf(sort.column), order: orderOf(sort) })}
        page={search.page ?? 1}
        pageCount={Math.ceil(total / pageSize)}
        rowCount={total}
        onPageChange={(page) => void navigate({ search: { ...search, page } })}
      />
    </>
  );
}
