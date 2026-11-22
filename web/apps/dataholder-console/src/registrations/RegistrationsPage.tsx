import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { Identifier } from "@cdr-lab/ui/identifier";
import { Link } from "@cdr-lab/ui/link";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { StatusBadge } from "@cdr-lab/ui/status-badge";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";

import { listRegistrations, problemOf, type Registration } from "../api/client";
import { statusTones } from "../statuses/tones";
import { changedSearch, orderOf, tableSort } from "@cdr-lab/web-platform/list-search";
import { sortColumnOf, validateRegistrationSearch } from "./search";

const route = getRouteApi("/registrations");

const column = dataTableColumns<Registration>();
const columns = column.columns([
  column.accessor("metadata.client_id", {
    header: "Client ID",
    cell: ({ getValue }) => (
      <Link
        className="font-mono text-xs"
        render={(props) => (
          <RouterLink {...props} to="/registrations/$clientId" params={{ clientId: getValue() }} />
        )}
      >
        {getValue()}
      </Link>
    ),
  }),
  column.accessor("metadata.software_id", {
    header: "Software ID",
    cell: ({ getValue, row }) => (
      <Identifier value={getValue()} label={`software ID of ${row.original.metadata.client_id}`} />
    ),
  }),
  column.accessor("metadata.client_name", {
    id: "clientName",
    header: "Client name",
    enableSorting: true,
  }),
  column.accessor("registeredAt", {
    header: "Registered",
    enableSorting: true,
    cell: ({ getValue }) => <LocalTime value={getValue()} />,
  }),
  column.accessor("lastUpdated", {
    header: "Last changed",
    cell: ({ getValue }) => <LocalTime value={getValue()} />,
  }),
  column.accessor("softwareProductStatus", {
    header: "Register status",
    cell: ({ getValue }) => {
      const status = getValue();
      return status ? (
        <StatusBadge tone={statusTones[status]}>{status}</StatusBadge>
      ) : (
        <span className="text-fg-muted">Not known</span>
      );
    },
  }),
  column.accessor("currentArrangementCount", {
    header: "Current arrangements",
    enableSorting: true,
  }),
  column.accessor("arrangementCount", { header: "All arrangements", enableSorting: true }),
]);

export function RegistrationsPage() {
  // The router keeps the search parameters of the URL that the route does not validate, so they are removed here.
  const search = validateRegistrationSearch(route.useSearch());
  const navigate = route.useNavigate();
  const list = useQuery({
    queryKey: ["registrations", search],
    queryFn: ({ signal }) => listRegistrations(search, signal),
    placeholderData: keepPreviousData,
  });

  let status: DataTableStatus = { state: "ready" };
  if (list.isPending) status = { state: "loading" };
  if (list.isError) status = { state: "error", problem: problemOf(list.error) };

  const rows = list.data?.data ?? [];

  return (
    <>
      <PageHeader
        title="Client registrations"
        description="The registration of each software product with the data holder, and the Register status of the software product."
      />
      {/* The API gives all the registrations in one response, so the table has one page. */}
      <DataTable
        label="Client registrations"
        columns={columns}
        rows={rows}
        getRowId={(row) => row.metadata.client_id}
        empty={{
          title: "No client registrations",
          description: "A software product registers with the registration endpoint.",
        }}
        status={status}
        sort={tableSort(search, "clientName")}
        onSortChange={(sort) =>
          void navigate({
            search: changedSearch(validateRegistrationSearch, search, {
              sort: sortColumnOf(sort.column),
              order: orderOf(sort),
            }),
          })
        }
        page={1}
        pageCount={1}
        rowCount={rows.length}
        onPageChange={() => {}}
      />
    </>
  );
}
