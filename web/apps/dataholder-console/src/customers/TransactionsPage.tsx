import { pageSize } from "@cdr-lab/web-platform/list-search";
import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { Identifier } from "@cdr-lab/ui/identifier";
import { Link } from "@cdr-lab/ui/link";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { StatusBadge } from "@cdr-lab/ui/status-badge";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink, useLocation } from "@tanstack/react-router";
import { Plus } from "lucide-react";
import type { ReactNode } from "react";

import { listTransactions, problemOf, type Transaction } from "../api/client";
import { formatAmount } from "../format";
import { customerQuery } from "./CustomerPage";

const route = getRouteApi("/customers/$customerNumber/accounts/$accountId/transactions");

// The fields that are not in a column, in the sequence of the operator interface. Each field with no value is left out.
function otherFields(t: Transaction): [string, ReactNode][] {
  const identifier = (name: string, value: string) => (
    <Identifier value={value} label={`${name} of ${t.description}`} className="whitespace-nowrap" />
  );
  const fields: [string, ReactNode][] = [
    ["Transaction ID", identifier("transaction ID", t.transactionId)],
  ];
  const add = (term: string, value: ReactNode) => {
    if (value !== undefined && value !== "") fields.push([term, value]);
  };
  const addIdentifier = (term: string, name: string, value: string | undefined) => {
    if (value) fields.push([term, identifier(name, value)]);
  };
  add("Reference", t.reference);
  if (t.postingDateTime && t.executionDateTime) {
    add("Executed", <LocalTime value={t.executionDateTime} />);
  }
  if (t.valueDateTime) add("Value date", <LocalTime value={t.valueDateTime} />);
  add("Merchant name", t.merchantName);
  add("Merchant category code", t.merchantCategoryCode);
  addIdentifier("Biller code", "biller code", t.billerCode);
  add("Biller name", t.billerName);
  addIdentifier("CRN", "CRN", t.crn);
  addIdentifier("APCA number", "APCA number", t.apcaNumber);
  const extended = t.extendedData;
  if (extended) {
    add("Payer", extended.payer);
    add("Payee", extended.payee);
    add("Extended description", extended.nppPayload.extendedDescription);
    add("NPP service", extended.nppPayload.service);
    add("NPP service version", extended.nppPayload.serviceVersion);
  }
  return fields;
}

const column = dataTableColumns<Transaction>();
const columns = column.columns([
  column.display({
    id: "time",
    header: "Time",
    // A posted transaction has its posting time, and a pending transaction its execution time.
    cell: ({ row }) => {
      const time = row.original.postingDateTime ?? row.original.executionDateTime;
      return time && <LocalTime value={time} />;
    },
  }),
  column.accessor("description", { header: "Description" }),
  column.accessor("type", { header: "Type" }),
  column.accessor("status", {
    header: "Status",
    cell: ({ getValue }) => (
      <StatusBadge tone={getValue() === "PENDING" ? "warning" : "neutral"}>
        {getValue()}
      </StatusBadge>
    ),
  }),
  column.accessor("amount", {
    header: "Amount",
    cell: ({ getValue }) => <span className="whitespace-nowrap">{formatAmount(getValue())}</span>,
  }),
  column.display({
    id: "details",
    header: "Details",
    cell: ({ row }) => (
      <dl className="flex flex-col gap-0.5">
        {otherFields(row.original).map(([term, value]) => (
          <div key={term} className="flex flex-wrap gap-x-2">
            <dt className="text-fg-muted">{term}</dt>
            <dd className="min-w-0 wrap-break-word">{value}</dd>
          </div>
        ))}
      </dl>
    ),
  }),
]);

export function TransactionsPage() {
  const { customerNumber, accountId } = route.useParams();
  const search = route.useSearch();
  const navigate = route.useNavigate();
  const detail = useQuery(customerQuery(customerNumber));
  const list = useQuery({
    queryKey: ["transactions", customerNumber, accountId, search],
    queryFn: ({ signal }) =>
      listTransactions(customerNumber, accountId, { ...search, pageSize }, signal),
    placeholderData: keepPreviousData,
  });

  let status: DataTableStatus = { state: "ready" };
  if (list.isPending) status = { state: "loading" };
  if (list.isError) status = { state: "error", problem: problemOf(list.error) };

  const focusTitle = useLocation({ select: (location) => location.state.focusTitle ?? false });
  const account = detail.data?.accounts.find((a) => a.accountId === accountId);
  const name = account?.displayName;
  const total = list.data?.totalRecords ?? 0;
  const canCreate = account?.actions.some((a) => a.action === "create-transaction") ?? false;

  return (
    <>
      <PageHeader
        title={name ? `Transactions of ${name}` : "Transactions"}
        description="The transactions of the account, newest first."
        focusTitle={focusTitle}
        actions={
          <>
            <Link
              render={(props) => (
                <RouterLink
                  {...props}
                  to="/customers/$customerNumber"
                  params={{ customerNumber }}
                />
              )}
            >
              Customer
            </Link>
            {canCreate && (
              <Link
                icon={Plus}
                render={(props) => (
                  <RouterLink
                    {...props}
                    to="/customers/$customerNumber/accounts/$accountId/create-transaction"
                    params={{ customerNumber, accountId }}
                  />
                )}
              >
                Create transaction
              </Link>
            )}
          </>
        }
      />
      <DataTable
        label="Transactions"
        columns={columns}
        rows={list.data?.data ?? []}
        getRowId={(row) => row.transactionId}
        empty={{ title: "No transactions on this page", description: "Go to the first page." }}
        status={status}
        page={search.page ?? 1}
        pageCount={Math.ceil(total / pageSize)}
        rowCount={total}
        onPageChange={(page) => void navigate({ search: { ...search, page } })}
      />
    </>
  );
}
