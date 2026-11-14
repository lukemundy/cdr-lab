import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { Link } from "@cdr-lab/ui/link";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Section } from "@cdr-lab/ui/section";
import { useQuery } from "@tanstack/react-query";
import { Link as RouterLink } from "@tanstack/react-router";
import type { ReactNode } from "react";

import { getOverview, problemOf, type Change, type Overview } from "../api/client";

const changeLabels: Record<Change["kind"], string> = {
  CUSTOMER_CREATED: "Customer created",
  ACCOUNT_CREATED: "Account created",
  ACCOUNT_CLOSED: "Account closed",
  ARRANGEMENT_CONFIRMED: "Arrangement confirmed",
  ARRANGEMENT_ENDED: "Arrangement ended",
  CLIENT_REGISTERED: "Client registered",
  CLIENT_UPDATED: "Registration changed",
  CLIENT_DELETED: "Registration deleted",
};

function detailOf(change: Change): ReactNode {
  if (change.productName) return change.productName;
  if (change.sharingEndsAt) {
    return (
      <>
        Sharing ends <LocalTime value={change.sharingEndsAt} />
      </>
    );
  }
  if (change.kind === "ARRANGEMENT_ENDED") {
    return change.endReason
      ? [change.endReason, change.withdrawalMethod].filter(Boolean).join(", ")
      : "The sharing period ended";
  }
  return null;
}

// Two changes can have the same time, kind and record, so each row has its position as its ID.
type ChangeRow = Change & { position: number };

const column = dataTableColumns<ChangeRow>();
const columns = column.columns([
  column.accessor("time", {
    header: "Time",
    cell: ({ getValue }) => <LocalTime value={getValue()} seconds />,
  }),
  column.accessor("kind", { header: "Change", cell: ({ getValue }) => changeLabels[getValue()] }),
  column.accessor("customerNumber", {
    header: "Customer number",
    cell: ({ getValue }) => {
      const customerNumber = getValue();
      return (
        customerNumber && (
          <Link
            render={(props) => (
              <RouterLink {...props} to="/customers/$customerNumber" params={{ customerNumber }} />
            )}
          >
            {customerNumber}
          </Link>
        )
      );
    },
  }),
  column.accessor("clientName", {
    header: "Software product",
    cell: ({ row }) => {
      const { clientId, clientName, clientRegistered } = row.original;
      if (!clientId || !clientRegistered) return clientName;
      return (
        <Link
          render={(props) => (
            <RouterLink {...props} to="/registrations/$clientId" params={{ clientId }} />
          )}
        >
          {clientName}
        </Link>
      );
    },
  }),
  column.display({ id: "detail", header: "Detail", cell: ({ row }) => detailOf(row.original) }),
]);

const plural = new Intl.PluralRules("en-AU");

function notActive(count: number) {
  return `${count} ${plural.select(count) === "one" ? "software product" : "software products"} not active`;
}

function Count({ term, value, children }: { term: string; value: number; children?: ReactNode }) {
  return (
    <div className="flex flex-col gap-1 px-4 py-3 not-first:border-l not-first:border-border">
      <dt className="text-sm font-medium text-fg-muted">{term}</dt>
      <dd className="text-3xl leading-tight font-semibold text-heading tabular-nums">{value}</dd>
      {children && <dd className="text-sm text-fg-muted">{children}</dd>}
    </div>
  );
}

function Counts({ overview }: { overview: Overview }) {
  return (
    <dl className="mb-6 grid grid-cols-4 rounded-lg border border-border bg-surface shadow-1">
      <Count term="Customers" value={overview.customerCount}>
        <Link render={(props) => <RouterLink {...props} to="/customers" />}>
          View the customers
        </Link>
      </Count>
      <Count term="Open accounts" value={overview.openAccountCount}>
        {overview.closedAccountCount} closed
      </Count>
      <Count term="Current arrangements" value={overview.currentArrangementCount}>
        {overview.endedArrangementCount} ended
      </Count>
      <Count term="Client registrations" value={overview.registrationCount}>
        <span className="flex flex-col">
          <span>{notActive(overview.notActiveSoftwareProductCount)}</span>
          <Link render={(props) => <RouterLink {...props} to="/registrations" />}>
            View the client registrations
          </Link>
        </span>
      </Count>
    </dl>
  );
}

export function OverviewPage() {
  const overview = useQuery({
    queryKey: ["overview"],
    queryFn: ({ signal }) => getOverview(signal),
  });

  let status: DataTableStatus = { state: "ready" };
  if (overview.isPending) status = { state: "loading" };
  if (overview.isError) status = { state: "error", problem: problemOf(overview.error) };

  const changes = (overview.data?.changes ?? []).map((change, position) =>
    Object.assign({ position }, change),
  );

  return (
    <>
      <PageHeader
        title="Overview"
        description="The number of records of the bank, and the newest changes. The overview shows no name, email address or account number."
      />
      {overview.data && <Counts overview={overview.data} />}
      <Section title="Newest changes">
        {/* The API gives only the newest changes, so the table has one page. */}
        <DataTable
          label="Newest changes"
          columns={columns}
          rows={changes}
          getRowId={(row) => String(row.position)}
          empty={{
            title: "No changes",
            description: "The bank has no customers and no registrations.",
          }}
          status={status}
          page={1}
          pageCount={1}
          rowCount={changes.length}
          onPageChange={() => {}}
        />
      </Section>
    </>
  );
}
