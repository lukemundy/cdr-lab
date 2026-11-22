import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { ProblemError } from "@cdr-lab/web-platform/problem";
import { ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { Pagination } from "@cdr-lab/ui/pagination";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi } from "@tanstack/react-router";

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
        <DescriptionList>
          <DescriptionItem term="Product">{account.productName}</DescriptionItem>
          <DescriptionItem term="Type">{categoryLabels[account.productCategory]}</DescriptionItem>
          <DescriptionItem term="Account number">{account.maskedNumber}</DescriptionItem>
          <DescriptionItem term="Balance">{formatAmount(balance.currentBalance)}</DescriptionItem>
          <DescriptionItem term="Available">
            {formatAmount(balance.availableBalance)}
          </DescriptionItem>
          <DescriptionItem term="Credit limit">{formatAmount(balance.creditLimit)}</DescriptionItem>
          <DescriptionItem term="Status">{openStatusLabels[account.openStatus]}</DescriptionItem>
          <DescriptionItem term="Opened">{formatDay(account.creationDate)}</DescriptionItem>
        </DescriptionList>
      </div>
      <CloseAccount account={account} openAccounts={openAccounts} />
      <Transactions accountId={account.accountId} />
    </>
  );
}

function Transactions({ accountId }: { accountId: string }) {
  const { page = 1 } = route.useSearch();
  const navigate = route.useNavigate();
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
          <Pagination
            page={page}
            pageCount={list.data.totalPages}
            rowCount={list.data.totalRecords}
            onPageChange={(to) => void navigate({ search: { page: to } })}
          />
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
        <DescriptionList>
          <OtherFields transaction={transaction} />
        </DescriptionList>
      </details>
    </li>
  );
}

// The fields that are not in the row, in the sequence of BankingTransactionV2. Each field with no value is left out.
function OtherFields({ transaction: t }: { transaction: Transaction }) {
  return (
    <>
      <DescriptionItem term="Type">{transactionTypeLabels[t.type]}</DescriptionItem>
      <DescriptionItem term="Status">{transactionStatusLabels[t.status]}</DescriptionItem>
      {t.postingDateTime && (
        <DescriptionItem term="Posted">{formatDate(t.postingDateTime)}</DescriptionItem>
      )}
      {t.valueDateTime && (
        <DescriptionItem term="Value date">{formatDate(t.valueDateTime)}</DescriptionItem>
      )}
      {t.executionDateTime && (
        <DescriptionItem term="Made">{formatDate(t.executionDateTime)}</DescriptionItem>
      )}
      {t.reference && <DescriptionItem term="Reference">{t.reference}</DescriptionItem>}
      {t.merchantName && <DescriptionItem term="Merchant">{t.merchantName}</DescriptionItem>}
      {t.merchantCategoryCode && (
        <DescriptionItem term="Merchant category code">{t.merchantCategoryCode}</DescriptionItem>
      )}
      {t.billerCode && <DescriptionItem term="Biller code">{t.billerCode}</DescriptionItem>}
      {t.billerName && <DescriptionItem term="Biller">{t.billerName}</DescriptionItem>}
      {t.crn && (
        <DescriptionItem term="Customer reference number">
          <span className="break-all">{t.crn}</span>
        </DescriptionItem>
      )}
      {t.apcaNumber && <DescriptionItem term="APCA number">{t.apcaNumber}</DescriptionItem>}
    </>
  );
}
