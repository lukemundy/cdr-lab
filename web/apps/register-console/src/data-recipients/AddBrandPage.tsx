import { Form } from "@cdr-lab/ui/form";
import { Select, TextField } from "@cdr-lab/ui/form-field";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import { addDataRecipientBrand, getDataRecipient, problemOf } from "../api/client";
import { textOf } from "@cdr-lab/app/form-values";
import { brandAndProductStatusOf, brandAndProductStatusOptions } from "../statuses";
import { useSend } from "../use-change";

const route = getRouteApi("/data-recipients/$legalEntityId/add-brand");

// The form keeps the version that it loaded, also when the query gets the data recipient again.
function BrandForm(props: { legalEntityId: string; version: number }) {
  const [{ legalEntityId, version }] = useState(props);
  const [brandId] = useState(() => crypto.randomUUID());
  const send = useSend("data-recipient");

  function submit(data: FormData) {
    const add = {
      dataRecipientBrandId: textOf(data, "dataRecipientBrandId"),
      brandName: textOf(data, "brandName"),
      logoUri: textOf(data, "logoUri"),
      status: brandAndProductStatusOf(textOf(data, "status")),
    };
    return send(
      legalEntityId,
      () => addDataRecipientBrand(legalEntityId, version, add),
      `Brand ${add.brandName} added`,
      add.dataRecipientBrandId,
    );
  }

  return (
    <Form
      label="Add brand"
      submitLabel="Add brand"
      cancel={<RouterLink to="/data-recipients/$legalEntityId" params={{ legalEntityId }} />}
      onSubmit={submit}
    >
      <TextField
        label="Brand ID"
        name="dataRecipientBrandId"
        description="A new random UUID. Change it to give a known ID."
        spellCheck={false}
        defaultValue={brandId}
      />
      <TextField label="Name" name="brandName" />
      <TextField label="Logo URI" name="logoUri" spellCheck={false} />
      <Select
        label="Status"
        name="status"
        options={brandAndProductStatusOptions}
        defaultValue="ACTIVE"
      />
    </Form>
  );
}

export function AddBrandPage() {
  const { legalEntityId } = route.useParams();
  const detail = useQuery({
    queryKey: ["data-recipient", legalEntityId],
    queryFn: ({ signal }) => getDataRecipient(legalEntityId, signal),
  });
  return (
    <>
      <PageHeader
        title={`Add a brand to ${detail.data?.legalEntityName ?? "a data recipient"}`}
        description="The Register checks each value. Add the software products after the add."
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && (
        <BrandForm
          key={legalEntityId}
          legalEntityId={legalEntityId}
          version={detail.data.version}
        />
      )}
    </>
  );
}
