// The test page of csp.node.test.ts: each component in each state, under the CSP of the design system.
import "../styles.css";

import { CirclePause, Pencil, Trash2 } from "lucide-react";
import { useState } from "react";
import { createRoot } from "react-dom/client";

import { ActionRow } from "../components/action-row";
import { Alert } from "../components/alert";
import { AppShell, AppShellNavItem } from "../components/app-shell";
import { Button } from "../components/button";
import { CheckboxGroup } from "../components/checkbox-group";
import { ConfirmationDialog } from "../components/confirmation-dialog";
import {
  ConsumerButton,
  ConsumerCard,
  ConsumerFooter,
  ConsumerNavItem,
  ConsumerShell,
  PasswordStatement,
} from "../components/consumer-page";
import { CopyButton } from "../components/copy-button";
import {
  DataTable,
  dataTableColumns,
  type DataTableStatus,
  type Sort,
} from "../components/data-table";
import { DescriptionItem, DescriptionList } from "../components/description-list";
import { FileInput } from "../components/file-input";
import { Dialog, DialogClose, DialogFooter } from "../components/dialog";
import { DateTimeField } from "../components/date-time-field";
import { Fieldset } from "../components/fieldset";
import { FilterBar, FilterSelect } from "../components/filter-bar";
import { Form } from "../components/form";
import { Select, TextArea, TextField } from "../components/form-field";
import { Identifier } from "../components/identifier";
import { Link } from "../components/link";
import { PageHeader } from "../components/page-header";
import { Section } from "../components/section";
import { StatusBadge } from "../components/status-badge";
import type { ListStatus, Problem } from "../components/states";
import { Timeline, type TimelineEvent } from "../components/timeline";
import { ToastProvider, useToast } from "../components/toast";
import { LocalTime } from "../components/local-time";

const variants = ["primary", "secondary", "danger", "quiet"] as const;
const tones = ["neutral", "success", "warning", "danger", "info"] as const;

type Recipient = { id: string; name: string; status: string };

const column = dataTableColumns<Recipient>();
const columns = column.columns([
  column.accessor("name", { header: "Name", enableSorting: true }),
  column.accessor("status", {
    header: "Status",
    enableSorting: true,
    cell: ({ getValue }) => <StatusBadge>{getValue()}</StatusBadge>,
  }),
  column.display({
    id: "id",
    header: "ID",
    cell: ({ row }) => <CopyButton value={row.original.id} label={`ID of ${row.original.name}`} />,
  }),
]);

const rows: Recipient[] = [
  { id: "a8f1c3d2-6b1e-4c5a-9f0e-2d7b3c4e5f60", name: "Mock Recipient One", status: "ACTIVE" },
  { id: "b9e2d4c3-7a2f-4d6b-8e1f-3c8a4d5f6e71", name: "Mock Recipient Two", status: "SUSPENDED" },
];

const statuses = [
  { value: "", label: "All statuses" },
  { value: "ACTIVE", label: "Active" },
];

const tableStates: { label: string; rows: Recipient[]; status: DataTableStatus }[] = [
  { label: "Empty", rows: [], status: { state: "ready" } },
  { label: "Loading", rows: [], status: { state: "loading" } },
  {
    label: "Error",
    rows: [],
    status: { state: "error", problem: { title: "Service Unavailable", detail: "Try again." } },
  },
];

const events: TimelineEvent[] = [
  {
    id: "2",
    time: "2026-10-23T14:05:07.123456789Z",
    title: "Software product deactivated",
    details: (
      <DescriptionList>
        <DescriptionItem term="Software product">Mock Product</DescriptionItem>
      </DescriptionList>
    ),
  },
  { id: "1", time: "2026-10-22T01:00:00Z", title: "Loaded into the Register" },
];

const timelineStates: { label: string; status: ListStatus }[] = [
  { label: "Empty", status: { state: "ready" } },
  { label: "Loading", status: { state: "loading" } },
  {
    label: "Error",
    status: { state: "error", problem: { title: "Not Found", detail: "Try again." } },
  },
];

// The action runs for a short time, so that the test sees the busy state.
function later<T>(value: () => T): Promise<T> {
  return new Promise((resolve) => setTimeout(resolve, 300)).then(value);
}

function Feedback() {
  const toast = useToast();
  return (
    <section aria-label="Dialogs, alerts and toasts">
      <Alert title="Conflict">The data recipient changed. See its current status.</Alert>
      <p>
        <Dialog
          trigger={<Button variant="secondary">Open the dialog</Button>}
          title="Register a client certificate"
          description="A dialog with a title and a description."
        >
          <DialogFooter>
            <DialogClose>Close the dialog</DialogClose>
          </DialogFooter>
        </Dialog>
        <ConfirmationDialog
          trigger={<Button variant="danger">Suspend</Button>}
          title="Suspend Mock Recipient One?"
          action="Suspend"
          effects={["Mock Recipient One changes from ACTIVE to SUSPENDED."]}
          onConfirm={() =>
            later(() => {
              toast.show({ title: "Mock Recipient One is suspended" });
              return undefined;
            })
          }
        />
        <ConfirmationDialog
          trigger={<Button variant="danger">Revoke</Button>}
          title="Revoke Mock Recipient One?"
          action="Revoke"
          effects={[
            "Mock Recipient One changes from ACTIVE to REVOKED.",
            "Mock Product changes from ACTIVE to INACTIVE.",
          ]}
          irreversible
          onConfirm={() =>
            later((): Problem => ({ title: "Conflict", detail: "The data recipient changed." }))
          }
        />
        <ConfirmationDialog
          trigger={<Button variant="danger">Remove</Button>}
          title="Remove the certificate?"
          action="Remove"
          effects={["The software product cannot use the certificate."]}
          onConfirm={() => Promise.reject(new Error("no network"))}
        />
      </p>
    </section>
  );
}

const industries = [
  { value: "banking", label: "Banking" },
  { value: "energy", label: "Energy" },
];

const levels = [
  { value: "UNRESTRICTED", label: "Unrestricted" },
  { value: "SPONSORED", label: "Sponsored" },
];

function RecordForm() {
  return (
    <Section title="Form">
      <Form
        label="Change the details"
        submitLabel="Save changes"
        cancel={<a href="#detail">Cancel</a>}
        onSubmit={() =>
          later(() => ({
            problem: {
              title: "Values in the request are not valid",
              detail: "A value is not valid.",
            },
            fieldErrors: [
              { pointer: "#/accreditationLevel", detail: "Choose a level." },
              { pointer: "#/industries", detail: "Choose an industry." },
              { pointer: "#/endpointDetail/publicBaseUri", detail: "Enter a URI." },
            ],
          }))
        }
      >
        <TextField label="Legal entity name" name="legalEntityName" defaultValue="Mock Pty Ltd" />
        <Select label="Accreditation level" name="accreditationLevel" options={levels} />
        <CheckboxGroup label="Industries" name="industries" options={industries} />
        <Fieldset legend="Endpoints">
          <TextField label="Public base URI" name="endpointDetail/publicBaseUri" />
        </Fieldset>
      </Form>
    </Section>
  );
}

function TestPage() {
  const [sort, setSort] = useState<Sort>({ column: "name", direction: "ascending" });
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [historyPage, setHistoryPage] = useState(1);
  const [certificate, setCertificate] = useState("");

  return (
    <AppShell
      name="Design system"
      navigationLabel="Main"
      navigation={
        <>
          <AppShellNavItem href="#" current>
            Test page
          </AppShellNavItem>
          <AppShellNavItem href="#other">Other page</AppShellNavItem>
        </>
      }
    >
      <PageHeader
        title="Design system test page"
        description="Each component in each state."
        actions={<Button>Refresh</Button>}
      />
      <FilterBar label="Filter the records" searchLabel="Name" search={search} onSearch={setSearch}>
        <FilterSelect label="Status" value={status} options={statuses} onChange={setStatus} />
      </FilterBar>
      <DataTable
        label="Records"
        columns={columns}
        rows={rows}
        getRowId={(row) => row.id}
        empty={{ title: "No records" }}
        sort={sort}
        onSortChange={setSort}
        page={page}
        pageCount={3}
        rowCount={55}
        onPageChange={setPage}
      />
      {tableStates.map((state) => (
        <DataTable
          key={state.label}
          label={`${state.label} records`}
          columns={columns}
          rows={state.rows}
          getRowId={(row) => row.id}
          empty={{ title: "No records", description: "Change the filters." }}
          status={state.status}
          page={1}
          pageCount={0}
          rowCount={0}
          onPageChange={() => {}}
        />
      ))}
      <Section
        title="Legal entity"
        description="A description list in a section."
        links={
          <Link href="#change" icon={Pencil}>
            Change details
          </Link>
        }
      >
        <DescriptionList>
          <DescriptionItem term="Legal entity ID">
            <Identifier value="a8f1c3d2-6b1e-4c5a-9f0e-2d7b3c4e5f60" label="legal entity ID" />
          </DescriptionItem>
          <DescriptionItem term="BSB">
            <Identifier value="062000" display="062-000" label="BSB" />
          </DescriptionItem>
          <DescriptionItem term="Last change">
            <LocalTime value="2026-10-23T14:05:07Z" />
          </DescriptionItem>
          <DescriptionItem term="Industries" values={["banking", "energy"]} />
          <DescriptionItem term="Logo URI" />
        </DescriptionList>
        <ActionRow>
          <Button variant="secondary" icon={CirclePause}>
            Pause
          </Button>
          <Button variant="danger" icon={Trash2} busy>
            Delete
          </Button>
        </ActionRow>
        <Section title="Brands">
          <p>A nested section.</p>
        </Section>
      </Section>
      <Section title="History">
        <Timeline
          label="Events"
          events={events}
          empty={{ title: "No events" }}
          page={historyPage}
          pageCount={4}
          rowCount={80}
          onPageChange={setHistoryPage}
        />
        {timelineStates.map((state) => (
          <Timeline
            key={state.label}
            label={`${state.label} events`}
            events={[]}
            empty={{ title: "No events", description: "Change the page." }}
            status={state.status}
            page={1}
            pageCount={0}
            rowCount={0}
            onPageChange={() => {}}
          />
        ))}
      </Section>
      <Section title="Form fields">
        <TextField label="Text field" />
        <TextField label="Text field with help" description="Help text for the field." />
        <TextField label="Text field with an error" error="Enter a value." defaultValue="x" />
        <TextField label="Disabled text field" disabled defaultValue="Mock Product" />
        <DateTimeField label="Date-time field" description="In the time zone of this browser." />
        <DateTimeField
          label="Date-time field with an error"
          error="Give a time."
          defaultValue="2026-10-26T14:05:30"
        />
        <TextArea label="Disabled text area" disabled />
        <TextArea label="Text area with an error" error="The body is not one PEM certificate." />
        <FileInput label="File input with an error" error="Choose a file." onFileText={() => {}} />
        <FileInput label="Disabled file input" disabled onFileText={() => {}} />
        <CheckboxGroup
          label="Checkbox group with help and an error"
          name="groupWithError"
          options={industries}
          defaultValue={["energy"]}
          description="Help text for the group."
          error="Choose an industry."
        />
        <CheckboxGroup
          label="Disabled checkbox group"
          name="disabledGroup"
          options={industries}
          disabled
        />
        <Fieldset legend="Fieldset">
          <TextField label="Field in a fieldset" />
        </Fieldset>
        <FileInput
          label="Certificate file"
          description="Its text goes into the certificate field."
          accept=".pem,.crt,.cer"
          onFileText={setCertificate}
        />
        <TextArea
          label="Certificate"
          monospace
          value={certificate}
          onChange={(event) => setCertificate(event.target.value)}
        />
      </Section>
      <section aria-label="Buttons">
        {variants.map((variant) => (
          <p key={variant}>
            <Button variant={variant}>Save</Button>
            <Button variant={variant} disabled>
              Save
            </Button>
            <Button variant={variant} busy>
              Save
            </Button>
          </p>
        ))}
      </section>
      <section aria-label="Links">
        <p>
          See the <Link href="#history">history</Link> of the record.
        </p>
      </section>
      <section aria-label="Status badges">
        {tones.map((tone) => (
          <StatusBadge key={tone} tone={tone}>
            {tone}
          </StatusBadge>
        ))}
      </section>
      <RecordForm />
      <Feedback />
    </AppShell>
  );
}

// The consumer components, in a region of their own, because the shell has its own header and main.
function ConsumerTestPage() {
  return (
    <section aria-label="Consumer page">
      <ConsumerShell
        brandName="Mock Bank"
        logo="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='96' height='32'/%3E"
        navigationLabel="Consumer"
        navigation={<ConsumerNavItem href="#products">Products</ConsumerNavItem>}
        footer={
          <ConsumerFooter>
            <PasswordStatement>Third parties never need your password.</PasswordStatement>
            <p>Mock Bank is not a real bank.</p>
          </ConsumerFooter>
        }
      >
        <ConsumerCard>
          <h2>Consumer card</h2>
          {(["primary", "secondary", "quiet"] as const).map((variant) => (
            <ConsumerButton key={variant} variant={variant} href={`#${variant}`}>
              {variant} button link
            </ConsumerButton>
          ))}
        </ConsumerCard>
      </ConsumerShell>
    </section>
  );
}

const root = document.getElementById("root");
if (root) {
  createRoot(root).render(
    <ToastProvider>
      <TestPage />
      <ConsumerTestPage />
    </ToastProvider>,
  );
}
