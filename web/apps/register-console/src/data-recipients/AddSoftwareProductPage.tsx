import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Form } from "@cdr-lab/ui/form";
import { Select, TextField } from "@cdr-lab/ui/form-field";
import { Identifier } from "@cdr-lab/ui/identifier";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import {
  addSoftwareProduct,
  getDataRecipient,
  problemOf,
  type DataRecipientBrand,
} from "../api/client";
import { textOf } from "@cdr-lab/console/form-values";
import { brandAndProductStatusOf, brandAndProductStatusOptions } from "../statuses";
import { SoftwareProductFields, softwareProductDetailsOf } from "./SoftwareProductFields";
import { useSend } from "../use-change";

const route = getRouteApi("/data-recipients/$legalEntityId/brands/$brandId/add-software-product");

const noBrand = {
  title: "The brand does not exist",
  detail: "The data recipient has no brand with this ID.",
};

type ProductFormProps = { legalEntityId: string; version: number; brand: DataRecipientBrand };

// The form keeps the version that it loaded, also when the query gets the data recipient again.
function ProductForm(props: ProductFormProps) {
  const [{ legalEntityId, version, brand }] = useState(props);
  const brandId = brand.dataRecipientBrandId;
  const [softwareProductId] = useState(() => crypto.randomUUID());
  const send = useSend("data-recipient");

  function submit(data: FormData) {
    const add = {
      softwareProductId: textOf(data, "softwareProductId"),
      ...softwareProductDetailsOf(data),
      status: brandAndProductStatusOf(textOf(data, "status")),
    };
    return send(
      legalEntityId,
      () => addSoftwareProduct(legalEntityId, brandId, version, add),
      `Software product ${add.softwareProductName} added`,
      add.softwareProductId,
    );
  }

  return (
    <>
      <DescriptionList>
        <DescriptionItem term="Brand ID">
          <Identifier value={brandId} label={`brand ID of ${brand.brandName}`} />
        </DescriptionItem>
      </DescriptionList>
      <Form
        label="Add software product"
        submitLabel="Add software product"
        cancel={<RouterLink to="/data-recipients/$legalEntityId" params={{ legalEntityId }} />}
        onSubmit={submit}
      >
        <TextField
          label="Software product ID"
          name="softwareProductId"
          description="A new random UUID. Change it to give a known ID."
          spellCheck={false}
          defaultValue={softwareProductId}
        />
        <SoftwareProductFields />
        <Select
          label="Status"
          name="status"
          options={brandAndProductStatusOptions}
          defaultValue="ACTIVE"
        />
      </Form>
    </>
  );
}

export function AddSoftwareProductPage() {
  const { legalEntityId, brandId } = route.useParams();
  const detail = useQuery({
    queryKey: ["data-recipient", legalEntityId],
    queryFn: ({ signal }) => getDataRecipient(legalEntityId, signal),
  });
  const brand = detail.data?.dataRecipientBrands.find((b) => b.dataRecipientBrandId === brandId);
  return (
    <>
      <PageHeader
        title={`Add a software product to ${brand?.brandName ?? "a brand"}`}
        description="The Register checks each value. Register the client certificates after the add."
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && !brand && <ErrorState problem={noBrand} />}
      {detail.data && brand && (
        <ProductForm
          key={`${legalEntityId} ${brandId}`}
          legalEntityId={legalEntityId}
          version={detail.data.version}
          brand={brand}
        />
      )}
    </>
  );
}
