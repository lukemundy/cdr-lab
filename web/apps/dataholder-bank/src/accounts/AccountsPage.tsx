import { ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { Link } from "@tanstack/react-router";

import { problemOf, type Account } from "../api/client";
import { usePageTitle } from "../brand";
import { categoryLabels } from "../products/categories";
import { formatAmount } from "../products/format";
import { SignedIn, useSessionEndedBy } from "../session/session";
import { PageHeading } from "../status";
import { Term, Terms } from "../terms";
import { useAccounts } from "./useAccounts";

export function AccountsPage() {
  usePageTitle("Accounts");

  return <SignedIn>{() => <Accounts />}</SignedIn>;
}

function Accounts() {
  const accounts = useAccounts();
  const ended = useSessionEndedBy(accounts.error);
  const all = accounts.data?.accounts ?? [];
  const open = all.filter((a) => a.openStatus === "OPEN");
  const closed = all.filter((a) => a.openStatus === "CLOSED");

  return (
    <ConsumerCard>
      <PageHeading>Your accounts</PageHeading>
      {(accounts.isPending || ended) && <LoadingState label="Loading your accounts" />}
      {accounts.isError && !ended && <ErrorState problem={problemOf(accounts.error)} />}
      {accounts.data && open.length === 0 && (
        <>
          {/* CDR Rules, rule 1.10B(1)(b): a customer with no open account is not eligible to share. */}
          <p>
            {all.length === 0 ? "You have no account with us yet." : "You have no open account."}
          </p>
          <p>
            You cannot share your data under the Consumer Data Right until you have an open account.
          </p>
        </>
      )}
      {open.length > 0 && <AccountGroup heading="Open accounts" accounts={open} />}
      {closed.length > 0 && <AccountGroup heading="Closed accounts" accounts={closed} />}
    </ConsumerCard>
  );
}

function AccountGroup({ heading, accounts }: { heading: string; accounts: Account[] }) {
  return (
    <section aria-label={heading}>
      <h2>{heading}</h2>
      <ul className="account-list">
        {accounts.map((account) => (
          <li key={account.accountId} className="account-card" data-status={account.openStatus}>
            <img className="account-mark" src="/brand/mark.svg" alt="" />
            <h3>
              <Link
                to="/accounts/$accountId"
                params={{ accountId: account.accountId }}
                className="inline-flex min-h-control items-center rounded-sm text-link underline underline-offset-2 hover:decoration-2 focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-focus"
              >
                {account.displayName}
              </Link>
            </h3>
            <AccountTerms account={account} />
          </li>
        ))}
      </ul>
    </section>
  );
}

export function AccountTerms({ account }: { account: Account }) {
  return (
    <Terms>
      <Term term="Product">{account.productName}</Term>
      <Term term="Type">{categoryLabels[account.productCategory]}</Term>
      <Term term="Account number">{account.maskedNumber}</Term>
      <Term term="Balance">{formatAmount(account.balance.currentBalance)}</Term>
      <Term term="Available">{formatAmount(account.balance.availableBalance)}</Term>
    </Terms>
  );
}
