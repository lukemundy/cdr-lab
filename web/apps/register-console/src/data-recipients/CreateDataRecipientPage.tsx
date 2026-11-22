import { Form } from "@cdr-lab/ui/form";
import { Select, TextField } from "@cdr-lab/ui/form-field";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import { createDataRecipient } from "../api/client";
import { textOf } from "@cdr-lab/app/form-values";
import { levels } from "./levels";
import { statuses } from "./search";
import { useSend } from "../use-change";

const statusOptions = statuses.map((status) => ({ value: status, label: status }));

export function CreateDataRecipientPage() {
  const [legalEntityId] = useState(() => crypto.randomUUID());
  const send = useSend("data-recipient");

  function submit(data: FormData) {
    const status = statuses.find((s) => s === textOf(data, "status"));
    if (!status) throw new Error("the form has no status");
    const create = {
      legalEntityId: textOf(data, "legalEntityId"),
      legalEntityName: textOf(data, "legalEntityName"),
      accreditationNumber: textOf(data, "accreditationNumber"),
      accreditationLevel: levels.find((l) => l.value === textOf(data, "accreditationLevel"))?.value,
      logoUri: textOf(data, "logoUri"),
      status,
    };
    return send(
      create.legalEntityId,
      () => createDataRecipient(create),
      `Data recipient ${create.legalEntityName} created`,
    );
  }

  return (
    <>
      <PageHeader
        title="Create a data recipient"
        description="The Register checks each value. Add the brands and the software products after the create."
      />
      <Form
        label="Create data recipient"
        submitLabel="Create data recipient"
        cancel={<RouterLink to="/data-recipients" />}
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
        <TextField label="Accreditation number" name="accreditationNumber" />
        <Select
          label="Accreditation level"
          name="accreditationLevel"
          placeholder="Choose a level"
          options={levels}
          defaultValue=""
        />
        <TextField label="Logo URI" name="logoUri" spellCheck={false} />
        <Select label="Status" name="status" options={statusOptions} defaultValue="ACTIVE" />
      </Form>
    </>
  );
}
