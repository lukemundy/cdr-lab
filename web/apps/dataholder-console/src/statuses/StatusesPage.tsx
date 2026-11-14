import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { FilterBar, FilterSelect } from "@cdr-lab/ui/filter-bar";
import { Identifier } from "@cdr-lab/ui/identifier";
import { Link } from "@cdr-lab/ui/link";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { StatusBadge } from "@cdr-lab/ui/status-badge";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";

import { listSoftwareProductStatuses, problemOf, type SoftwareProductStatus } from "../api/client";
import {
  changedSearch,
  orderOf,
  tableSort,
  type ListSearchChange,
} from "@cdr-lab/console/list-search";
import { registrations, sortColumnOf, validateStatusSearch, type StatusSearch } from "./search";
import { statusTones } from "./tones";

const route = getRouteApi("/register-statuses");

const column = dataTableColumns<SoftwareProductStatus>();
const columns = column.columns([
  column.accessor("name", {
    header: "Name",
    enableSorting: true,
    cell: ({ row }) => {
      const { clientId, name } = row.original;
      return clientId ? (
        <Link
          render={(props) => (
            <RouterLink {...props} to="/registrations/$clientId" params={{ clientId }} />
          )}
        >
          {name}
        </Link>
      ) : (
        <span className="text-fg-muted">Not registered</span>
      );
    },
  }),
  column.accessor("softwareProductId", {
    header: "Software product ID",
    cell: ({ getValue }) => (
      <Identifier value={getValue()} label={`software product ID of ${getValue()}`} />
    ),
  }),
  column.accessor("status", {
    header: "Status",
    enableSorting: true,
    cell: ({ getValue }) => <StatusBadge tone={statusTones[getValue()]}>{getValue()}</StatusBadge>,
  }),
]);

const registrationLabels = { registered: "Registered", notRegistered: "Not registered" } as const;

const registrationOptions = [
  { value: "", label: "All" },
  ...registrations.map((value) => ({ value, label: registrationLabels[value] })),
];

export function StatusesPage() {
  // The router keeps the search parameters of the URL that the route does not validate, so they are removed here.
  const search = validateStatusSearch(route.useSearch());
  const navigate = route.useNavigate();
  const list = useQuery({
    queryKey: ["software-product-statuses", search],
    queryFn: ({ signal }) => listSoftwareProductStatuses(search, signal),
    placeholderData: keepPreviousData,
  });

  function update(change: ListSearchChange<StatusSearch>) {
    void navigate({ search: changedSearch(validateStatusSearch, search, change) });
  }

  let status: DataTableStatus = { state: "ready" };
  if (list.isPending) status = { state: "loading" };
  if (list.isError) status = { state: "error", problem: problemOf(list.error) };

  const rows = list.data?.data ?? [];
  const receivedAt = list.data?.receivedAt;

  let empty: { title: string; description?: string } = { title: "No statuses" };
  if (receivedAt && search.registration) {
    empty = { title: "No software products match", description: "Change the registration filter." };
  } else if (receivedAt) {
    empty = { title: "No statuses", description: "The Register gave no software products." };
  }

  return (
    <>
      <PageHeader
        title="Register statuses"
        description="The status of each software product on the Register, as the data holder has it. The data holder gets the statuses at an interval, so a change on the Register shows here up to 5 minutes later."
      />
      {list.data && (
        <DescriptionList>
          <DescriptionItem term="Last got from the Register">
            {receivedAt ? (
              <LocalTime value={receivedAt} seconds />
            ) : (
              "The data holder has not got the statuses from the Register."
            )}
          </DescriptionItem>
        </DescriptionList>
      )}
      <FilterBar label="Register statuses">
        <FilterSelect
          label="Registration"
          value={search.registration ?? ""}
          options={registrationOptions}
          onChange={(value) => update({ registration: registrations.find((r) => r === value) })}
        />
      </FilterBar>
      {/* The API gives all the statuses in one response, so the table has one page. */}
      <DataTable
        label="Register statuses"
        columns={columns}
        rows={rows}
        getRowId={(row) => row.softwareProductId}
        empty={empty}
        status={status}
        sort={tableSort(search, "name")}
        onSortChange={(sort) => update({ sort: sortColumnOf(sort.column), order: orderOf(sort) })}
        page={1}
        pageCount={1}
        rowCount={rows.length}
        onPageChange={() => {}}
      />
    </>
  );
}
