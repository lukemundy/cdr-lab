import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Form } from "@cdr-lab/ui/form";
import { Identifier } from "@cdr-lab/ui/identifier";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import {
  changeSoftwareProductDetails,
  getDataRecipient,
  problemOf,
  type SoftwareProduct,
} from "../api/client";
import { SoftwareProductFields, softwareProductDetailsOf } from "./SoftwareProductFields";
import { useChange } from "../use-change";

const route = getRouteApi(
  "/data-recipients/$legalEntityId/software-products/$softwareProductId/change",
);

const noSoftwareProduct = {
  title: "The software product does not exist",
  detail: "The data recipient has no software product with this ID.",
};

type ProductFormProps = { legalEntityId: string; version: number; loaded: SoftwareProduct };

// The form keeps the values and the version that it loaded, also when the query gets the data recipient again.
function ProductForm(props: ProductFormProps) {
  const [{ legalEntityId, version, loaded }] = useState(props);
  const softwareProductId = loaded.softwareProductId;
  const send = useChange("data-recipient", legalEntityId);

  function submit(data: FormData) {
    const details = softwareProductDetailsOf(data);
    return send(
      () => changeSoftwareProductDetails(legalEntityId, softwareProductId, version, details),
      details.softwareProductName,
      softwareProductId,
    );
  }

  return (
    <>
      <DescriptionList>
        <DescriptionItem term="Software product ID">
          <Identifier
            value={softwareProductId}
            label={`software product ID of ${loaded.softwareProductName}`}
          />
        </DescriptionItem>
      </DescriptionList>
      <Form
        label="Change software product"
        submitLabel="Save changes"
        cancel={<RouterLink to="/data-recipients/$legalEntityId" params={{ legalEntityId }} />}
        onSubmit={submit}
      >
        <SoftwareProductFields loaded={loaded} />
      </Form>
    </>
  );
}

export function ChangeSoftwareProductPage() {
  const { legalEntityId, softwareProductId } = route.useParams();
  const detail = useQuery({
    queryKey: ["data-recipient", legalEntityId],
    queryFn: ({ signal }) => getDataRecipient(legalEntityId, signal),
  });
  const product = detail.data?.dataRecipientBrands
    .flatMap((b) => b.softwareProducts)
    .find((p) => p.softwareProductId === softwareProductId);
  return (
    <>
      <PageHeader
        title={`Change the details of ${product?.softwareProductName ?? "a software product"}`}
        description="The Register checks each value. The software product ID, the status and the client certificates do not change."
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && !product && <ErrorState problem={noSoftwareProduct} />}
      {detail.data && product && (
        <ProductForm
          key={`${legalEntityId} ${softwareProductId}`}
          legalEntityId={legalEntityId}
          version={detail.data.version}
          loaded={product}
        />
      )}
    </>
  );
}
