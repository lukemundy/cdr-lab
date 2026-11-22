import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { FilterBar, FilterSelect } from "@cdr-lab/ui/filter-bar";
import { Identifier } from "@cdr-lab/ui/identifier";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Link } from "@cdr-lab/ui/link";
import { StatusBadge } from "@cdr-lab/ui/status-badge";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { listDataRecipients, problemOf, type DataRecipientSummary } from "../api/client";
import { dataRecipientTones } from "../statuses";
import {
  changedSearch,
  orderOf,
  pageSize,
  tableSort,
  type ListSearchChange,
} from "@cdr-lab/web-platform/list-search";
import {
  apiQuery,
  sortColumnOf,
  statuses,
  validateDataRecipientSearch,
  type DataRecipientSearch,
} from "./search";

const route = getRouteApi("/data-recipients");

const column = dataTableColumns<DataRecipientSummary>();
const columns = column.columns([
  column.accessor("legalEntityName", {
    header: "Name",
    enableSorting: true,
    cell: ({ row }) => (
      <Link
        render={(props) => (
          <RouterLink
            {...props}
            to="/data-recipients/$legalEntityId"
            params={{ legalEntityId: row.original.legalEntityId }}
          />
        )}
      >
        {row.original.legalEntityName}
      </Link>
    ),
  }),
  column.accessor("legalEntityId", {
    header: "Legal entity ID",
    cell: ({ row }) => (
      <Identifier
        value={row.original.legalEntityId}
        label={`legal entity ID of ${row.original.legalEntityName}`}
        className="whitespace-nowrap"
      />
    ),
  }),
  column.accessor("accreditationNumber", { header: "Accreditation number", enableSorting: true }),
  column.accessor("status", {
    header: "Status",
    enableSorting: true,
    cell: ({ getValue }) => (
      <StatusBadge tone={dataRecipientTones[getValue()]}>{getValue()}</StatusBadge>
    ),
  }),
  column.accessor("brandCount", { header: "Brands", enableSorting: true }),
  column.accessor("softwareProductCount", { header: "Software products", enableSorting: true }),
]);

const statusOptions = [
  { value: "", label: "All statuses" },
  ...statuses.map((status) => ({ value: status, label: status })),
];

export function DataRecipientsPage() {
  // The router keeps the search parameters of the URL that the route does not validate, so they are removed here.
  const search = validateDataRecipientSearch(route.useSearch());
  const navigate = route.useNavigate();
  const list = useQuery({
    queryKey: ["data-recipients", search],
    queryFn: ({ signal }) => listDataRecipients(apiQuery(search), signal),
    placeholderData: keepPreviousData,
  });

  function update(change: ListSearchChange<DataRecipientSearch>, { replace = false } = {}) {
    void navigate({ search: changedSearch(validateDataRecipientSearch, search, change), replace });
  }

  let status: DataTableStatus = { state: "ready" };
  if (list.isPending) status = { state: "loading" };
  if (list.isError) status = { state: "error", problem: problemOf(list.error) };

  const total = list.data?.totalRecords ?? 0;

  return (
    <>
      <PageHeader
        title="Data recipients"
        description="The data recipient legal entities of the Register."
        actions={
          <Link
            icon={Plus}
            render={(props) => <RouterLink {...props} to="/create/data-recipient" />}
          >
            Create data recipient
          </Link>
        }
      />
      <FilterBar
        label="Data recipients"
        searchLabel="Search by name"
        search={search.q ?? ""}
        // A search applies while the operator types, so it replaces the history entry and does not add one for each pause.
        onSearch={(q) => update({ q: q || undefined }, { replace: true })}
      >
        <FilterSelect
          label="Status"
          value={search.status ?? ""}
          options={statusOptions}
          onChange={(value) => update({ status: statuses.find((s) => s === value) })}
        />
      </FilterBar>
      <DataTable
        label="Data recipients"
        columns={columns}
        rows={list.data?.data ?? []}
        getRowId={(row) => row.legalEntityId}
        empty={{
          title: "No data recipients match",
          description: "Change the search or the status filter.",
        }}
        status={status}
        sort={tableSort(search, "legalEntityName")}
        onSortChange={(sort) => update({ sort: sortColumnOf(sort.column), order: orderOf(sort) })}
        page={search.page ?? 1}
        pageCount={Math.ceil(total / pageSize)}
        rowCount={total}
        onPageChange={(page) => void navigate({ search: { ...search, page } })}
      />
    </>
  );
}
