import { ActionRow } from "@cdr-lab/ui/action-row";
import { Button } from "@cdr-lab/ui/button";
import { DataTable, dataTableColumns, type DataTableStatus } from "@cdr-lab/ui/data-table";
import { TextArea } from "@cdr-lab/ui/form-field";
import { Identifier } from "@cdr-lab/ui/identifier";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Section } from "@cdr-lab/ui/section";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { keepPreviousData, useQuery, useQueryClient } from "@tanstack/react-query";
import { getRouteApi } from "@tanstack/react-router";
import { Download } from "lucide-react";

import { pageSize, validatePageSearch } from "@cdr-lab/console/list-search";

import {
  checkRevocation,
  getCDRCABundle,
  listRevocations,
  problemOf,
  revokeCertificate,
  type Revocation,
} from "../api/client";
import { saveFile } from "../save-file";
import { reasonWords } from "./reasons";
import { RevokeCertificate } from "./RevokeCertificate";

const column = dataTableColumns<Revocation>();

const columns = column.columns([
  column.accessor("serialNumber", {
    header: "Serial number",
    cell: ({ getValue }) => <Identifier value={getValue()} label="serial number" />,
  }),
  column.accessor("subject", { header: "Subject" }),
  column.accessor("thumbprint", {
    header: "Thumbprint",
    cell: ({ getValue }) => <Identifier value={getValue()} label="thumbprint" />,
  }),
  column.accessor("revokedAt", {
    header: "Revoked",
    cell: ({ getValue }) => <LocalTime value={getValue()} seconds />,
  }),
  column.accessor("reason", {
    header: "Reason",
    cell: ({ getValue }) => reasonWords[getValue()],
  }),
]);

const route = getRouteApi("/cdr-certificates");

function CABundle() {
  const bundle = useQuery({
    queryKey: ["cdr-ca-bundle"],
    queryFn: ({ signal }) => getCDRCABundle(signal),
  });

  return (
    <Section
      card
      title="CDR CA bundle"
      description="The intermediate and then the root certificate of the mock CDR CA. A component that uses CDR certificates trusts this bundle."
    >
      {bundle.isPending && <LoadingState />}
      {bundle.isError && <ErrorState problem={problemOf(bundle.error)} />}
      {bundle.data && (
        <>
          <TextArea
            label="PEM CA bundle"
            monospace
            rows={10}
            readOnly
            spellCheck={false}
            value={bundle.data.caBundle}
          />
          <ActionRow>
            <Button
              variant="secondary"
              icon={Download}
              onClick={() => saveFile("cdr-ca-bundle.pem", bundle.data.caBundle)}
            >
              Save as file
            </Button>
          </ActionRow>
        </>
      )}
    </Section>
  );
}

export function CDRCertificatesPage() {
  const search = validatePageSearch(route.useSearch());
  const navigate = route.useNavigate();
  const queryClient = useQueryClient();
  const list = useQuery({
    queryKey: ["cdr-revocations", search],
    queryFn: ({ signal }) => listRevocations({ ...search, pageSize }, signal),
    placeholderData: keepPreviousData,
  });
  const total = list.data?.totalRecords ?? 0;
  let status: DataTableStatus = { state: "ready" };
  if (list.isPending) status = { state: "loading" };
  if (list.isError) status = { state: "error", problem: problemOf(list.error) };

  return (
    <>
      <PageHeader
        title="CDR certificates"
        description="The CA bundle and the revocations of the mock CDR intermediate CA, which issues the CDR client and server certificates."
      />
      <CABundle />
      <ActionRow>
        <RevokeCertificate
          check={checkRevocation}
          revoke={revokeCertificate}
          refresh={() => queryClient.invalidateQueries({ queryKey: ["cdr-revocations"] })}
          showNewest={() => navigate({ search: {} })}
        />
      </ActionRow>
      <Section
        title="Revocations"
        description="Each revocation of the mock CDR intermediate, newest first."
      >
        <DataTable
          label="Revocations"
          columns={columns}
          rows={list.data?.data ?? []}
          getRowId={(row) => row.serialNumber}
          empty={{
            title: "No revocations on this page",
            description:
              "The mock CDR intermediate revoked no certificates, or the page is after the last page.",
          }}
          status={status}
          page={search.page ?? 1}
          pageCount={Math.ceil(total / pageSize)}
          rowCount={total}
          onPageChange={(page) => void navigate({ search: { page } })}
        />
      </Section>
    </>
  );
}
