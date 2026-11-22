import { pageSize } from "@cdr-lab/web-platform/list-search";
import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Identifier } from "@cdr-lab/ui/identifier";
import { Link } from "@cdr-lab/ui/link";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Section } from "@cdr-lab/ui/section";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { StatusBadge } from "@cdr-lab/ui/status-badge";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";

import {
  getRegistration,
  listRegistrationArrangements,
  problemOf,
  type ClientArrangement,
  type Registration,
} from "../api/client";
import { arrangementTones, nameOf } from "../customers/CustomerPage";
import { statusTones } from "../statuses/tones";

const route = getRouteApi("/registrations/$clientId");

function identifier(value: string | undefined, label: string) {
  return value ? <Identifier value={value} label={label} /> : undefined;
}

function Summary({ registration }: { registration: Registration }) {
  const { metadata, softwareProductStatus: status } = registration;

  return (
    <Section card title="Registration">
      <DescriptionList>
        <DescriptionItem term="Client ID">
          <Identifier value={metadata.client_id} label="client ID" />
        </DescriptionItem>
        <DescriptionItem term="Software ID">
          <Identifier value={metadata.software_id} label="software ID" />
        </DescriptionItem>
        <DescriptionItem term="Register status">
          {status ? (
            <StatusBadge tone={statusTones[status]}>{status}</StatusBadge>
          ) : (
            <span className="text-fg-muted">Not known</span>
          )}
        </DescriptionItem>
        <DescriptionItem term="Registered">
          <LocalTime value={registration.registeredAt} seconds />
        </DescriptionItem>
        <DescriptionItem term="Last changed">
          <LocalTime value={registration.lastUpdated} seconds />
        </DescriptionItem>
      </DescriptionList>
    </Section>
  );
}

function Metadata({ metadata: m }: { metadata: Registration["metadata"] }) {
  return (
    <Section card title="Metadata">
      <DescriptionList>
        <DescriptionItem term="Client name">{m.client_name}</DescriptionItem>
        <DescriptionItem term="Client description">{m.client_description}</DescriptionItem>
        <DescriptionItem term="Client ID issued">
          <LocalTime value={new Date(m.client_id_issued_at * 1000).toISOString()} seconds />
        </DescriptionItem>
        <DescriptionItem term="Brand ID">{identifier(m.org_id, "brand ID")}</DescriptionItem>
        <DescriptionItem term="Brand name">{m.org_name}</DescriptionItem>
        <DescriptionItem term="Legal entity ID">
          {identifier(m.legal_entity_id, "legal entity ID")}
        </DescriptionItem>
        <DescriptionItem term="Legal entity name">{m.legal_entity_name}</DescriptionItem>
        <DescriptionItem
          term="Redirect URIs"
          values={m.redirect_uris.map((value) => identifier(value, "redirect URI"))}
        />
        <DescriptionItem term="Recipient base URI">
          {identifier(m.recipient_base_uri, "recipient base URI")}
        </DescriptionItem>
        <DescriptionItem term="JWKS URI">{identifier(m.jwks_uri, "JWKS URI")}</DescriptionItem>
        <DescriptionItem term="Revocation URI">
          {identifier(m.revocation_uri, "revocation URI")}
        </DescriptionItem>
        <DescriptionItem term="Client URI">
          {identifier(m.client_uri, "client URI")}
        </DescriptionItem>
        <DescriptionItem term="Logo URI">{identifier(m.logo_uri, "logo URI")}</DescriptionItem>
        <DescriptionItem term="Sector identifier URI">
          {identifier(m.sector_identifier_uri, "sector identifier URI")}
        </DescriptionItem>
        <DescriptionItem term="Terms of service URI">
          {identifier(m.tos_uri, "terms of service URI")}
        </DescriptionItem>
        <DescriptionItem term="Policy URI">
          {identifier(m.policy_uri, "policy URI")}
        </DescriptionItem>
        <DescriptionItem term="Scopes" values={m.scope.split(" ").filter(Boolean)} />
        <DescriptionItem term="Grant types" values={m.grant_types} />
        <DescriptionItem term="Response types" values={m.response_types} />
        <DescriptionItem term="Application type">{m.application_type}</DescriptionItem>
        <DescriptionItem term="Software roles">{m.software_roles}</DescriptionItem>
        <DescriptionItem term="Token endpoint authentication method">
          {m.token_endpoint_auth_method}
        </DescriptionItem>
        <DescriptionItem term="Token endpoint signing algorithm">
          {m.token_endpoint_auth_signing_alg}
        </DescriptionItem>
        <DescriptionItem term="Request object signing algorithm">
          {m.request_object_signing_alg}
        </DescriptionItem>
        <DescriptionItem term="ID token signing algorithm">
          {m.id_token_signed_response_alg}
        </DescriptionItem>
        <DescriptionItem term="Authorisation response signing algorithm">
          {m.authorization_signed_response_alg}
        </DescriptionItem>
      </DescriptionList>
    </Section>
  );
}

const column = dataTableColumns<ClientArrangement>();
const columns = column.columns([
  column.accessor("arrangementId", {
    header: "Arrangement ID",
    cell: ({ row, getValue }) => (
      <Identifier value={getValue()} label={`arrangement ID of ${nameOf(row.original)}`} />
    ),
  }),
  column.display({
    id: "customer",
    header: "Customer",
    cell: ({ row }) => {
      const { customerNumber } = row.original;
      const name = nameOf(row.original);
      return (
        <span className="flex flex-col items-start gap-1">
          <Link
            render={(props) => (
              <RouterLink {...props} to="/customers/$customerNumber" params={{ customerNumber }} />
            )}
          >
            {name}
          </Link>
          <Identifier
            value={customerNumber}
            label={`customer number of ${name}`}
            className="whitespace-nowrap"
          />
        </span>
      );
    },
  }),
  column.accessor("status", {
    header: "Status",
    cell: ({ getValue }) => (
      <StatusBadge tone={arrangementTones[getValue()]}>{getValue()}</StatusBadge>
    ),
  }),
  column.accessor("confirmedAt", {
    header: "Confirmed",
    cell: ({ getValue }) => <LocalTime value={getValue()} />,
  }),
  column.accessor("sharingEndsAt", {
    header: "Sharing ends",
    cell: ({ getValue }) => <LocalTime value={getValue()} />,
  }),
]);

function Arrangements({ clientId }: { clientId: string }) {
  const search = route.useSearch();
  const navigate = route.useNavigate();
  const list = useQuery({
    queryKey: ["registration-arrangements", clientId, search],
    queryFn: ({ signal }) =>
      listRegistrationArrangements(clientId, { ...search, pageSize }, signal),
    placeholderData: keepPreviousData,
  });

  let status: DataTableStatus = { state: "ready" };
  if (list.isPending) status = { state: "loading" };
  if (list.isError) status = { state: "error", problem: problemOf(list.error) };
  const total = list.data?.totalRecords ?? 0;

  return (
    <Section title="Arrangements" description="The arrangements of the client, newest first.">
      <DataTable
        label="Arrangements"
        columns={columns}
        rows={list.data?.data ?? []}
        getRowId={(row) => row.arrangementId}
        empty={{ title: "No arrangements on this page", description: "Go to the first page." }}
        status={status}
        page={search.page ?? 1}
        pageCount={Math.ceil(total / pageSize)}
        rowCount={total}
        onPageChange={(page) => void navigate({ search: { ...search, page } })}
      />
    </Section>
  );
}

export function RegistrationPage() {
  const { clientId } = route.useParams();
  const registration = useQuery({
    queryKey: ["registration", clientId],
    queryFn: ({ signal }) => getRegistration(clientId, signal),
  });

  return (
    <>
      <PageHeader
        title={registration.data?.metadata.client_name ?? "Client registration"}
        description="The registration of a software product with the data holder, with its metadata and each arrangement of its client."
      />
      {registration.isPending && <LoadingState />}
      {registration.isError && <ErrorState problem={problemOf(registration.error)} />}
      {registration.data && (
        <>
          <Summary registration={registration.data} />
          <Metadata metadata={registration.data.metadata} />
          <Arrangements clientId={clientId} />
        </>
      )}
    </>
  );
}
