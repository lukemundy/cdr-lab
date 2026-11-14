import { Form } from "@cdr-lab/ui/form";
import { Select, TextField } from "@cdr-lab/ui/form-field";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import { createDataHolder } from "../api/client";
import { textOf } from "@cdr-lab/console/form-values";
import { useSend } from "../use-change";
import { statuses } from "./search";

const statusOptions = statuses.map((status) => ({ value: status, label: status }));

export function CreateDataHolderPage() {
  const [legalEntityId] = useState(() => crypto.randomUUID());
  const send = useSend("data-holder");

  function submit(data: FormData) {
    const status = statuses.find((s) => s === textOf(data, "status"));
    if (!status) throw new Error("the form has no status");
    const create = {
      legalEntityId: textOf(data, "legalEntityId"),
      legalEntityName: textOf(data, "legalEntityName"),
      logoUri: textOf(data, "logoUri"),
      status,
    };
    return send(
      create.legalEntityId,
      () => createDataHolder(create),
      `Data holder ${create.legalEntityName} created`,
    );
  }

  return (
    <>
      <PageHeader
        title="Create a data holder"
        description="The Register checks each value. Add the brands after the create."
      />
      <Form
        label="Create data holder"
        submitLabel="Create data holder"
        cancel={<RouterLink to="/data-holders" />}
        onSubmit={submit}
      >
        <TextField
          label="Legal entity ID"
          name="legalEntityId"
          description="A new random UUID. Change it to give a known ID."
          spellCheck={false}
          defaultValue={legalEntityId}
        />
        <TextField label="Name" name="legalEntityName" />
        <TextField label="Logo URI" name="logoUri" spellCheck={false} />
        <Select label="Status" name="status" options={statusOptions} defaultValue="ACTIVE" />
      </Form>
    </>
  );
}
