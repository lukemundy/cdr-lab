import { ActionRow } from "@cdr-lab/ui/action-row";
import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Identifier } from "@cdr-lab/ui/identifier";
import { Link } from "@cdr-lab/ui/link";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Section } from "@cdr-lab/ui/section";
import { EmptyState, ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { StatusBadge, type StatusBadgeProps } from "@cdr-lab/ui/status-badge";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink, useLocation } from "@tanstack/react-router";
import { AppWindow, Plus, ReceiptText } from "lucide-react";

import {
  getCustomer,
  problemOf,
  type Account,
  type Arrangement,
  type CustomerDetail,
} from "../api/client";
import { formatAmount, formatBsb, formatDate } from "../format";
import { AccountActions } from "./AccountActions";
import { ArrangementActions } from "./ArrangementActions";

const route = getRouteApi("/customers/$customerNumber");

type Tone = StatusBadgeProps["tone"];

const accountTones = { OPEN: "success", CLOSED: "neutral" } as const satisfies Record<
  Account["openStatus"],
  Tone
>;

export const arrangementTones = { CURRENT: "success", ENDED: "neutral" } as const satisfies Record<
  Arrangement["status"],
  Tone
>;

export function customerQuery(customerNumber: string) {
  return {
    queryKey: ["customer", customerNumber],
    queryFn: ({ signal }: { signal: AbortSignal }) => getCustomer(customerNumber, signal),
  };
}

export function nameOf({ firstName, lastName }: { firstName?: string; lastName: string }) {
  return firstName ? `${firstName} ${lastName}` : lastName;
}

function Customer({ detail }: { detail: CustomerDetail }) {
  return (
    <Section card title="Customer">
      <DescriptionList>
        <DescriptionItem term="Customer number">
          <Identifier
            value={detail.customerNumber}
            label={`customer number of ${nameOf(detail)}`}
          />
        </DescriptionItem>
        <DescriptionItem term="First name">{detail.firstName}</DescriptionItem>
        <DescriptionItem term="Last name">{detail.lastName}</DescriptionItem>
        <DescriptionItem term="Email address">{detail.emailAddress}</DescriptionItem>
        <DescriptionItem term="Last changed">
          <LocalTime value={detail.lastUpdated} />
        </DescriptionItem>
      </DescriptionList>
    </Section>
  );
}

type AccountCardProps = {
  account: Account;
  customerNumber: string;
  refresh: () => Promise<unknown>;
};

function AccountCard({ account, customerNumber, refresh }: AccountCardProps) {
  const of = account.displayName;

  return (
    <Section
      card
      collapsible
      title={of}
      status={
        <StatusBadge tone={accountTones[account.openStatus]}>{account.openStatus}</StatusBadge>
      }
      links={
        <Link
          icon={ReceiptText}
          render={(props) => (
            <RouterLink
              {...props}
              to="/customers/$customerNumber/accounts/$accountId/transactions"
              params={{ customerNumber, accountId: account.accountId }}
            />
          )}
        >
          Transactions<span className="sr-only"> of {of}</span>
        </Link>
      }
    >
      <DescriptionList>
        <DescriptionItem term="Account ID">
          <Identifier value={account.accountId} label={`account ID of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Product category">{account.productCategory}</DescriptionItem>
        <DescriptionItem term="Product name">{account.productName}</DescriptionItem>
        <DescriptionItem term="BSB">
          <Identifier value={account.bsb} display={formatBsb(account.bsb)} label={`BSB of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Account number">
          <Identifier value={account.accountNumber} label={`account number of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Credit limit">{formatAmount(account.creditLimit)}</DescriptionItem>
        <DescriptionItem term="Current balance">
          {formatAmount(account.currentBalance)}
        </DescriptionItem>
        <DescriptionItem term="Available balance">
          {formatAmount(account.availableBalance)}
        </DescriptionItem>
        <DescriptionItem term="Created">{formatDate(account.creationDate)}</DescriptionItem>
      </DescriptionList>
      <ActionRow>
        <AccountActions customerNumber={customerNumber} account={account} refresh={refresh} />
      </ActionRow>
    </Section>
  );
}

type ArrangementCardProps = {
  arrangement: Arrangement;
  accounts: readonly Account[];
  customerNumber: string;
  refresh: () => Promise<unknown>;
};

function ArrangementCard({ arrangement, accounts, customerNumber, refresh }: ArrangementCardProps) {
  const of = arrangement.clientName;
  const accountName = (id: string) => accounts.find((a) => a.accountId === id)?.displayName;

  return (
    <Section
      card
      collapsible
      title={of}
      status={
        <StatusBadge tone={arrangementTones[arrangement.status]}>{arrangement.status}</StatusBadge>
      }
      links={
        arrangement.clientRegistered && (
          <Link
            icon={AppWindow}
            render={(props) => (
              <RouterLink
                {...props}
                to="/registrations/$clientId"
                params={{ clientId: arrangement.clientId }}
              />
            )}
          >
            Registration<span className="sr-only"> of {of}</span>
          </Link>
        )
      }
    >
      <DescriptionList>
        <DescriptionItem term="Arrangement ID">
          <Identifier value={arrangement.arrangementId} label={`arrangement ID of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Client ID">
          <Identifier value={arrangement.clientId} label={`client ID of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Software ID">
          <Identifier value={arrangement.softwareId} label={`software ID of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Scopes" values={arrangement.scopes} />
        <DescriptionItem
          term="Accounts"
          empty="None"
          values={arrangement.accountIds.map((id) => (
            <span key={id} className="flex flex-col">
              {accountName(id)}
              <Identifier value={id} label={`account ID of ${accountName(id) ?? id} in ${of}`} />
            </span>
          ))}
        />
        <DescriptionItem term="Confirmed">
          <LocalTime value={arrangement.confirmedAt} />
        </DescriptionItem>
        <DescriptionItem term="Sharing ends">
          <LocalTime value={arrangement.sharingEndsAt} />
        </DescriptionItem>
        {arrangement.endedAt && (
          <DescriptionItem term="Ended">
            <LocalTime value={arrangement.endedAt} />
          </DescriptionItem>
        )}
        {arrangement.endReason && (
          <DescriptionItem term="End reason">{arrangement.endReason}</DescriptionItem>
        )}
        {arrangement.withdrawalMethod && (
          <DescriptionItem term="Withdrawal method">{arrangement.withdrawalMethod}</DescriptionItem>
        )}
        {arrangement.noticeStatus && (
          <DescriptionItem term="Notice status">{arrangement.noticeStatus}</DescriptionItem>
        )}
        {arrangement.noticeAttempts && (
          <DescriptionItem
            term="Notice attempts"
            values={arrangement.noticeAttempts.map((a) => (
              <span key={a.attempt}>
                Attempt {a.attempt}, <LocalTime value={a.at} seconds />: {a.result}
              </span>
            ))}
          />
        )}
      </DescriptionList>
      <ActionRow>
        <ArrangementActions
          customerNumber={customerNumber}
          arrangement={arrangement}
          refresh={refresh}
        />
      </ActionRow>
    </Section>
  );
}

export function CustomerPage() {
  const { customerNumber } = route.useParams();
  const detail = useQuery(customerQuery(customerNumber));
  const focusTitle = useLocation({ select: (location) => location.state.focusTitle ?? false });
  const queryClient = useQueryClient();
  // The list of the customers has the count of the open accounts and of the current arrangements.
  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ["customer", customerNumber] }),
      queryClient.invalidateQueries({ queryKey: ["customers"] }),
    ]);

  return (
    <>
      <PageHeader
        title={detail.data ? nameOf(detail.data) : "Customer"}
        description="A customer of the brand, with each account and each arrangement."
        focusTitle={focusTitle}
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && (
        <>
          <Customer detail={detail.data} />
          <Section
            title="Accounts"
            showAll="accounts"
            links={
              <Link
                icon={Plus}
                render={(props) => (
                  <RouterLink
                    {...props}
                    to="/customers/$customerNumber/create-account"
                    params={{ customerNumber }}
                  />
                )}
              >
                Create account
              </Link>
            }
          >
            {detail.data.accounts.length === 0 && (
              <EmptyState title="This customer has no accounts" />
            )}
            {detail.data.accounts.map((account) => (
              <AccountCard
                key={account.accountId}
                account={account}
                customerNumber={customerNumber}
                refresh={refresh}
              />
            ))}
          </Section>
          <Section title="Arrangements" showAll="arrangements">
            {detail.data.arrangements.length === 0 && (
              <EmptyState title="This customer has no arrangements" />
            )}
            {detail.data.arrangements.map((arrangement) => (
              <ArrangementCard
                key={arrangement.arrangementId}
                arrangement={arrangement}
                accounts={detail.data.accounts}
                customerNumber={customerNumber}
                refresh={refresh}
              />
            ))}
          </Section>
        </>
      )}
    </>
  );
}
