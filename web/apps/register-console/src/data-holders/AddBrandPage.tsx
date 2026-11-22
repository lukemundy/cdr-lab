import { Form } from "@cdr-lab/ui/form";
import { Select, TextField } from "@cdr-lab/ui/form-field";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import { addDataHolderBrand, getDataHolder, problemOf } from "../api/client";
import { textOf } from "@cdr-lab/web-platform/form-values";
import { brandAndProductStatusOf, brandAndProductStatusOptions } from "../statuses";
import { useSend } from "../use-change";
import { BrandFields, brandDetailsOf } from "./BrandFields";

const route = getRouteApi("/data-holders/$legalEntityId/add-brand");

const publicDetailsOnly = "A brand with no brand ID needs only its public details.";

// The form keeps the version that it loaded, also when the query gets the data holder again.
function BrandForm(props: { legalEntityId: string; version: number }) {
  const [{ legalEntityId, version }] = useState(props);
  const [brandId] = useState(() => crypto.randomUUID());
  const send = useSend("data-holder");

  function submit(data: FormData) {
    const dataHolderBrandId = textOf(data, "dataHolderBrandId");
    const interimId = textOf(data, "interimId");
    const add = {
      ...(dataHolderBrandId !== "" && { dataHolderBrandId }),
      ...(interimId !== "" && { interimId }),
      ...brandDetailsOf(data),
      status: brandAndProductStatusOf(textOf(data, "status")),
    };
    return send(
      legalEntityId,
      () => addDataHolderBrand(legalEntityId, version, add),
      `Brand ${add.brandName} added`,
      dataHolderBrandId || interimId,
    );
  }

  return (
    <Form
      label="Add brand"
      submitLabel="Add brand"
      cancel={<RouterLink to="/data-holders/$legalEntityId" params={{ legalEntityId }} />}
      onSubmit={submit}
    >
      <TextField
        label="Brand ID (optional)"
        name="dataHolderBrandId"
        description={`A new random UUID. Change it to give a known ID. ${publicDetailsOnly}`}
        spellCheck={false}
        defaultValue={brandId}
      />
      <TextField
        label="Interim ID (optional)"
        name="interimId"
        description={`A brand that has no brand ID needs an interim ID. ${publicDetailsOnly}`}
        spellCheck={false}
      />
      <BrandFields />
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
    queryKey: ["data-holder", legalEntityId],
    queryFn: ({ signal }) => getDataHolder(legalEntityId, signal),
  });
  return (
    <>
      <PageHeader
        title={`Add a brand to ${detail.data?.legalEntityName ?? "a data holder"}`}
        description="The Register checks each value."
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
