import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import type { ReactNode } from "react";

import { problemOf, type Consent, type DataCluster } from "../api/client";
import { usePageTitle } from "../brand";
import { formatDate } from "../products/format";
import { SignedIn, useSessionEndedBy } from "../session/session";
import { PageHeading } from "../status";

import { formatPeriod } from "./period";
import { StopSharing } from "./StopSharing";
import { useConsents } from "./useConsents";

// The consumer dashboard of rule 1.15 of the CDR Rules (services/dataholder/design/consumer-dashboard.md).
export function ConsentsPage() {
  usePageTitle("Data sharing");

  return <SignedIn>{() => <Consents />}</SignedIn>;
}

function Consents() {
  const consents = useConsents();
  const ended = useSessionEndedBy(consents.error);
  const all = consents.data?.consents ?? [];
  const current = all.filter((c) => c.status === "CURRENT");
  const past = all.filter((c) => c.status === "ENDED");

  return (
    <ConsumerCard>
      <PageHeading>Data sharing</PageHeading>
      {(consents.isPending || ended) && <LoadingState label="Loading your data sharing" />}
      {consents.isError && !ended && <ErrorState problem={problemOf(consents.error)} />}
      {consents.data && all.length === 0 && (
        <p>You do not share your data under the Consumer Data Right.</p>
      )}
      {current.length > 0 && <ConsentGroup heading="Current consents" consents={current} />}
      {past.length > 0 && <ConsentGroup heading="Ended consents" consents={past} />}
    </ConsumerCard>
  );
}

function ConsentGroup({ heading, consents }: { heading: string; consents: Consent[] }) {
  return (
    <section aria-label={heading}>
      <h2>{heading}</h2>
      <ul className="flex list-none flex-col gap-6 p-0">
        {consents.map((consent) => (
          <li key={consent.arrangementId}>
            <ConsentCard
              consent={consent}
              actions={consent.status === "CURRENT" && <StopSharing consent={consent} />}
            />
          </li>
        ))}
      </ul>
    </section>
  );
}

// actions has the controls of a current consent, for example the withdrawal.
export function ConsentCard({ consent, actions }: { consent: Consent; actions?: ReactNode }) {
  return (
    <article aria-label={consent.clientName} className="flex flex-col gap-3">
      <h3>{consent.clientName}</h3>
      <DescriptionList>
        <DescriptionItem term="Data recipient">{consent.brandName}</DescriptionItem>
        {consent.legalEntityName && (
          <DescriptionItem term="Legal entity">{consent.legalEntityName}</DescriptionItem>
        )}
        <DescriptionItem term="Consent given">{formatDate(consent.confirmedAt)}</DescriptionItem>
        <DescriptionItem term="Period">
          {formatPeriod(consent.sharingDurationSeconds)}
        </DescriptionItem>
        {consent.endedAt ? (
          <DescriptionItem term="Ended">{formatDate(consent.endedAt)}</DescriptionItem>
        ) : (
          <DescriptionItem term="Sharing ends">{formatDate(consent.sharingEndsAt)}</DescriptionItem>
        )}
      </DescriptionList>
      {consent.accounts.length > 0 && (
        <>
          <h4>Accounts</h4>
          <ul className="m-0 flex flex-col gap-1 pl-5">
            {consent.accounts.map((account) => (
              <li key={account.accountId}>
                {account.displayName}, {account.maskedNumber}
                {account.openStatus === "CLOSED" && " (closed)"}
              </li>
            ))}
          </ul>
        </>
      )}
      <h4>Data</h4>
      <ul className="m-0 flex list-none flex-col gap-3 p-0">
        {consent.dataClusters.map((cluster) => (
          <li key={cluster.name}>
            <ClusterDisclosures cluster={cluster} />
          </li>
        ))}
      </ul>
      <p>{handlingMessage(consent.brandName)}</p>
      {actions}
    </article>
  );
}

// CDS: Dashboard Standards, Data recipient handling details: the suggested text of the DSB, word for word.
function handlingMessage(brandName: string): string {
  return (
    `You should check with ${brandName} for more information on how they are handling your data, and for any ` +
    `other permissions you may have given them. See ${brandName}'s CDR policy or their Dashboard for more ` +
    "information."
  );
}

function ClusterDisclosures({ cluster }: { cluster: DataCluster }) {
  return (
    <>
      <p className="m-0 font-semibold">{cluster.name}</p>
      <p className="m-0">{cluster.permissions.join(", ")}</p>
      {cluster.firstDisclosedAt && cluster.lastDisclosedAt ? (
        <DescriptionList>
          <DescriptionItem term="First shared">
            {formatDate(cluster.firstDisclosedAt)}
          </DescriptionItem>
          <DescriptionItem term="Last shared">
            {formatDate(cluster.lastDisclosedAt)}
          </DescriptionItem>
        </DescriptionList>
      ) : (
        <p className="m-0">No data shared</p>
      )}
    </>
  );
}
