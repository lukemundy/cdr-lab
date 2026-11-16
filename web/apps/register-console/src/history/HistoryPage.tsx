import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Identifier } from "@cdr-lab/ui/identifier";
import { Link } from "@cdr-lab/ui/link";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Timeline, type TimelineEvent } from "@cdr-lab/ui/timeline";
import type { ListStatus } from "@cdr-lab/ui/states";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import type { ReactNode } from "react";

import {
  getDataHolder,
  getDataHolderHistory,
  getDataRecipient,
  getDataRecipientHistory,
  problemOf,
  type DataHolderHistoryEntry,
  type DataRecipientHistoryEntry,
} from "../api/client";
import { pageSize, validatePageSearch, type PageSearch } from "@cdr-lab/console/list-search";

type HistoryEntry = DataRecipientHistoryEntry | DataHolderHistoryEntry;

// The texts of the table of the history in services/register/design/operator-console.md.
const eventTexts = {
  DataRecipientImported: "Loaded into the Register",
  DataHolderImported: "Loaded into the Register",
  DataRecipientCreated: "Created by the operator",
  DataRecipientDetailsChanged: "Legal entity details changed",
  DataRecipientBrandAdded: "Brand added",
  DataRecipientBrandDetailsChanged: "Brand details changed",
  DataRecipientBrandActivated: "Brand activated",
  DataRecipientBrandDeactivated: "Brand deactivated",
  DataRecipientBrandRemoved: "Brand removed",
  SoftwareProductAdded: "Software product added",
  SoftwareProductDetailsChanged: "Software product details changed",
  DataHolderCreated: "Created by the operator",
  DataHolderDetailsChanged: "Legal entity details changed",
  DataHolderLegalEntityRemoved: "Data holder removed",
  DataHolderBrandAdded: "Brand added",
  DataHolderBrandDetailsChanged: "Brand details changed",
  DataHolderBrandActivated: "Brand activated",
  DataHolderBrandDeactivated: "Brand deactivated",
  DataHolderBrandRemoved: "Brand removed",
  AccreditationSuspended: "Accreditation suspended",
  AccreditationRevoked: "Accreditation revoked",
  AccreditationSurrendered: "Accreditation surrendered",
  AccreditationReinstated: "Accreditation reinstated",
  SoftwareProductActivated: "Software product activated",
  SoftwareProductDeactivated: "Software product deactivated",
  SoftwareProductRemoved: "Software product removed",
  ClientCertificateRegistered: "Client certificate registered",
  ClientCertificateIssued: "Client certificate issued",
  ClientCertificateRemoved: "Client certificate removed",
  ServerCertificateIssued: "Server certificate issued",
} as const satisfies Record<HistoryEntry["eventType"], string>;

function Details({ entry }: { entry: HistoryEntry }) {
  return (
    <DescriptionList>
      <DescriptionItem term="Event type">
        <Identifier value={entry.eventType} label={`event type of event ${entry.streamVersion}`} />
      </DescriptionItem>
      {"softwareProductId" in entry && (
        <>
          <DescriptionItem term="Software product">{entry.softwareProductName}</DescriptionItem>
          <DescriptionItem term="Software product ID">
            <Identifier
              value={entry.softwareProductId}
              label={`software product ID of event ${entry.streamVersion}`}
            />
          </DescriptionItem>
        </>
      )}
      {"brandId" in entry && (
        <>
          <DescriptionItem term="Brand">{entry.brandName}</DescriptionItem>
          <DescriptionItem term="Brand ID">
            <Identifier value={entry.brandId} label={`brand ID of event ${entry.streamVersion}`} />
          </DescriptionItem>
        </>
      )}
      {"thumbprint" in entry && (
        <DescriptionItem term="Thumbprint">
          <Identifier
            value={entry.thumbprint}
            label={`thumbprint of event ${entry.streamVersion}`}
          />
        </DescriptionItem>
      )}
    </DescriptionList>
  );
}

function timelineEvent(entry: HistoryEntry): TimelineEvent {
  return {
    id: String(entry.streamVersion),
    time: entry.recordedAt,
    title: eventTexts[entry.eventType],
    details: <Details entry={entry} />,
  };
}

function statusOf(query: { isPending: boolean; isError: boolean; error: unknown }): ListStatus {
  if (query.isPending) return { state: "loading" };
  if (query.isError) return { state: "error", problem: problemOf(query.error) };
  return { state: "ready" };
}

type HistoryViewProps = {
  name: string | undefined;
  detailLink: ReactNode;
  entries: readonly HistoryEntry[];
  total: number;
  status: ListStatus;
  search: PageSearch;
  onPageChange: (page: number) => void;
};

function HistoryView({
  name,
  detailLink,
  entries,
  total,
  status,
  search,
  onPageChange,
}: HistoryViewProps) {
  return (
    <>
      <PageHeader
        title={name ? `History of ${name}` : "History"}
        description="Each event of the legal entity, newest first. The events of one action have the same time."
        actions={detailLink}
      />
      <Timeline
        label="History"
        events={entries.map(timelineEvent)}
        empty={{ title: "No events on this page", description: "Go to the first page." }}
        status={status}
        page={search.page ?? 1}
        pageCount={Math.ceil(total / pageSize)}
        rowCount={total}
        onPageChange={onPageChange}
      />
    </>
  );
}

const dataRecipientRoute = getRouteApi("/data-recipients/$legalEntityId/history");

export function DataRecipientHistoryPage() {
  const { legalEntityId } = dataRecipientRoute.useParams();
  const search = validatePageSearch(dataRecipientRoute.useSearch());
  const navigate = dataRecipientRoute.useNavigate();
  const detail = useQuery({
    queryKey: ["data-recipient", legalEntityId],
    queryFn: ({ signal }) => getDataRecipient(legalEntityId, signal),
  });
  const history = useQuery({
    queryKey: ["data-recipient-history", legalEntityId, search],
    queryFn: ({ signal }) =>
      getDataRecipientHistory(legalEntityId, { ...search, pageSize }, signal),
    placeholderData: keepPreviousData,
  });

  return (
    <HistoryView
      name={detail.data?.legalEntityName}
      detailLink={
        <Link
          render={(props) => (
            <RouterLink
              {...props}
              to="/data-recipients/$legalEntityId"
              params={{ legalEntityId }}
            />
          )}
        >
          Detail
        </Link>
      }
      entries={history.data?.data ?? []}
      total={history.data?.totalRecords ?? 0}
      status={statusOf(history)}
      search={search}
      onPageChange={(page) => void navigate({ search: { ...search, page } })}
    />
  );
}

const dataHolderRoute = getRouteApi("/data-holders/$legalEntityId/history");

export function DataHolderHistoryPage() {
  const { legalEntityId } = dataHolderRoute.useParams();
  const search = validatePageSearch(dataHolderRoute.useSearch());
  const navigate = dataHolderRoute.useNavigate();
  const detail = useQuery({
    queryKey: ["data-holder", legalEntityId],
    queryFn: ({ signal }) => getDataHolder(legalEntityId, signal),
  });
  const history = useQuery({
    queryKey: ["data-holder-history", legalEntityId, search],
    queryFn: ({ signal }) => getDataHolderHistory(legalEntityId, { ...search, pageSize }, signal),
    placeholderData: keepPreviousData,
  });

  return (
    <HistoryView
      name={detail.data?.legalEntityName}
      detailLink={
        <Link
          render={(props) => (
            <RouterLink {...props} to="/data-holders/$legalEntityId" params={{ legalEntityId }} />
          )}
        >
          Detail
        </Link>
      }
      entries={history.data?.data ?? []}
      total={history.data?.totalRecords ?? 0}
      status={statusOf(history)}
      search={search}
      onPageChange={(page) => void navigate({ search: { ...search, page } })}
    />
  );
}
