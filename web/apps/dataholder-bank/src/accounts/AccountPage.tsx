import { ProblemError } from "@cdr-lab/console/problem";
import { ConsumerButton, ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link } from "@tanstack/react-router";

import {
  getBalance,
  listTransactions,
  problemOf,
  type Account,
  type Balance,
  type Transaction,
} from "../api/client";
import { usePageTitle } from "../brand";
import { NotFoundPage } from "../NotFoundPage";
import { categoryLabels } from "../products/categories";
import { formatAmount, formatDate } from "../products/format";
import { SignedIn, useSessionEndedBy } from "../session/session";
import { PageHeading } from "../status";
import { Term, Terms } from "../terms";
import { CloseAccount } from "./CloseAccount";
import { effectiveTime, formatDay, formatSignedAmount } from "./format";
import { openStatusLabels, transactionStatusLabels, transactionTypeLabels } from "./labels";
import { useAccounts } from "./useAccounts";

const route = getRouteApi("/accounts/$accountId");
const notFound = 404;

const isNotFound = (error: unknown) => error instanceof ProblemError && error.status === notFound;

export function AccountPage() {
  return <SignedIn>{() => <AccountOfCustomer />}</SignedIn>;
}

function AccountOfCustomer() {
  const { accountId } = route.useParams();
  const accounts = useAccounts();
  const balance = useQuery({
    queryKey: ["accounts", accountId, "balance"],
    queryFn: ({ signal }) => getBalance(accountId, signal),
  });
  const account = accounts.data?.accounts.find((a) => a.accountId === accountId);
  const openAccounts = accounts.data?.accounts.filter((a) => a.openStatus === "OPEN").length ?? 0;
  const ended = useSessionEndedBy(accounts.error, balance.error);

  if ((accounts.data && !account) || isNotFound(balance.error)) return <NotFoundPage />;
  if (ended) return <LoadingState label="Loading the account" />;

  return (
    <ConsumerCard>
      {(accounts.isPending || balance.isPending) && <LoadingState label="Loading the account" />}
      {accounts.isError && <ErrorState problem={problemOf(accounts.error)} />}
      {balance.isError && <ErrorState problem={problemOf(balance.error)} />}
      {account && balance.data && (
        <Detail account={account} balance={balance.data} openAccounts={openAccounts} />
      )}
    </ConsumerCard>
  );
}

function Detail({
  account,
  balance,
  openAccounts,
}: {
  account: Account;
  balance: Balance;
  openAccounts: number;
}) {
  usePageTitle(account.displayName);

  return (
    <>
      <PageHeading>{account.displayName}</PageHeading>
      <div className="account-card account-summary" data-status={account.openStatus}>
        <img className="account-mark" src="/brand/mark.svg" alt="" />
        <Terms>
          <Term term="Product">{account.productName}</Term>
          <Term term="Type">{categoryLabels[account.productCategory]}</Term>
          <Term term="Account number">{account.maskedNumber}</Term>
          <Term term="Balance">{formatAmount(balance.currentBalance)}</Term>
          <Term term="Available">{formatAmount(balance.availableBalance)}</Term>
          <Term term="Credit limit">{formatAmount(balance.creditLimit)}</Term>
          <Term term="Status">{openStatusLabels[account.openStatus]}</Term>
          <Term term="Opened">{formatDay(account.creationDate)}</Term>
        </Terms>
      </div>
      <CloseAccount account={account} openAccounts={openAccounts} />
      <Transactions accountId={account.accountId} />
    </>
  );
}

function Transactions({ accountId }: { accountId: string }) {
  const { page = 1 } = route.useSearch();
  const list = useQuery({
    queryKey: ["accounts", accountId, "transactions", page],
    queryFn: ({ signal }) => listTransactions(accountId, page, signal),
    placeholderData: keepPreviousData,
  });
  const ended = useSessionEndedBy(list.error);

  return (
    <section aria-labelledby="transactions">
      <h2 id="transactions">Transactions</h2>
      {/* The default period of Get Transactions For Account (CDS: Get Transactions For Account). */}
      <p className="text-fg-muted">The last 90 days, newest first.</p>
      {list.isPending && <LoadingState label="Loading the transactions" />}
      {list.isError && !ended && <ErrorState problem={problemOf(list.error)} />}
      {list.data?.totalRecords === 0 && (
        <p>This account has no transactions in the last 90 days.</p>
      )}
      {list.data && list.data.transactions.length > 0 && (
        <>
          <ul className="flex list-none flex-col gap-3 p-0">
            {list.data.transactions.map((transaction) => (
              <TransactionItem key={transaction.transactionId} transaction={transaction} />
            ))}
          </ul>
          <Pages page={page} pageCount={list.data.totalPages} accountId={accountId} />
        </>
      )}
    </section>
  );
}

function TransactionItem({ transaction }: { transaction: Transaction }) {
  const time = effectiveTime(transaction);
  const pending = transaction.status === "PENDING";

  return (
    <li data-transaction className="border-b border-border pb-3">
      <div className="flex flex-wrap items-baseline gap-x-3">
        <span data-cell className="text-fg-muted">
          {time && formatDate(time)}
        </span>
        <span data-cell className="min-w-0 grow wrap-break-word font-semibold">
          {transaction.description}
        </span>
        {pending && (
          <span data-cell className="rounded-sm border border-border px-1 text-sm">
            {transactionStatusLabels.PENDING}
          </span>
        )}
        <span data-cell className="whitespace-nowrap tabular-nums">
          {formatSignedAmount(transaction.amount)}
        </span>
      </div>
      <details>
        <summary className="inline-flex min-h-control cursor-pointer items-center text-link underline underline-offset-2">
          Details<span className="sr-only"> of {transaction.description}</span>
        </summary>
        <Terms>
          <OtherFields transaction={transaction} />
        </Terms>
      </details>
    </li>
  );
}

// The fields that are not in the row, in the sequence of BankingTransactionV2. Each field with no value is left out.
function OtherFields({ transaction: t }: { transaction: Transaction }) {
  return (
    <>
      <Term term="Type">{transactionTypeLabels[t.type]}</Term>
      <Term term="Status">{transactionStatusLabels[t.status]}</Term>
      {t.postingDateTime && <Term term="Posted">{formatDate(t.postingDateTime)}</Term>}
      {t.valueDateTime && <Term term="Value date">{formatDate(t.valueDateTime)}</Term>}
      {t.executionDateTime && <Term term="Made">{formatDate(t.executionDateTime)}</Term>}
      {t.reference && <Term term="Reference">{t.reference}</Term>}
      {t.merchantName && <Term term="Merchant">{t.merchantName}</Term>}
      {t.merchantCategoryCode && (
        <Term term="Merchant category code">{t.merchantCategoryCode}</Term>
      )}
      {t.billerCode && <Term term="Biller code">{t.billerCode}</Term>}
      {t.billerName && <Term term="Biller">{t.billerName}</Term>}
      {t.crn && (
        <Term term="Customer reference number">
          <span className="break-all">{t.crn}</span>
        </Term>
      )}
      {t.apcaNumber && <Term term="APCA number">{t.apcaNumber}</Term>}
    </>
  );
}

function Pages({
  page,
  pageCount,
  accountId,
}: {
  page: number;
  pageCount: number;
  accountId: string;
}) {
  if (pageCount <= 1) return null;

  const link = (to: number, label: string) => (
    <ConsumerButton
      variant="secondary"
      render={(props) => (
        <Link {...props} to="/accounts/$accountId" params={{ accountId }} search={{ page: to }} />
      )}
    >
      {label}
    </ConsumerButton>
  );

  return (
    <nav aria-label="Pages of transactions" className="flex flex-wrap items-center gap-3 pt-4">
      {page > 1 && link(page - 1, "Previous page")}
      <span className="tabular-nums">
        Page {page} of {pageCount}
      </span>
      {page < pageCount && link(page + 1, "Next page")}
    </nav>
  );
}
