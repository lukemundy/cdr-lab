import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Form } from "@cdr-lab/ui/form";
import { Identifier } from "@cdr-lab/ui/identifier";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import {
  changeDataHolderBrandDetails,
  getDataHolder,
  problemOf,
  type DataHolderBrand,
} from "../api/client";
import { useChange } from "../use-change";
import { brandIdOf } from "./brand-id";
import { BrandFields, brandDetailsOf } from "./BrandFields";

const route = getRouteApi("/data-holders/$legalEntityId/brands/$brandId/change");

const noBrand = {
  title: "The brand does not exist",
  detail: "The data holder has no brand with this ID.",
};

type BrandFormProps = { legalEntityId: string; version: number; loaded: DataHolderBrand };

// The form keeps the values and the version that it loaded, also when the query gets the data holder again.
function BrandForm(props: BrandFormProps) {
  const [{ legalEntityId, version, loaded }] = useState(props);
  const send = useChange("data-holder", legalEntityId);
  // An optional value that is not set gives undefined, which the description list shows as "Not set".
  const identifier = (value: string | undefined, name: string) =>
    value && <Identifier value={value} label={`${name} of ${loaded.brandName}`} />;

  function submit(data: FormData) {
    const details = brandDetailsOf(data);
    return send(
      () => changeDataHolderBrandDetails(legalEntityId, brandIdOf(loaded), version, details),
      details.brandName,
      brandIdOf(loaded),
    );
  }

  return (
    <>
      <DescriptionList>
        <DescriptionItem term="Brand ID">
          {identifier(loaded.dataHolderBrandId, "brand ID")}
        </DescriptionItem>
        <DescriptionItem term="Interim ID">
          {identifier(loaded.interimId, "interim ID")}
        </DescriptionItem>
      </DescriptionList>
      <Form
        label="Change brand"
        submitLabel="Save changes"
        cancel={<RouterLink to="/data-holders/$legalEntityId" params={{ legalEntityId }} />}
        onSubmit={submit}
      >
        <BrandFields loaded={loaded} />
      </Form>
    </>
  );
}

export function ChangeBrandPage() {
  const { legalEntityId, brandId } = route.useParams();
  const detail = useQuery({
    queryKey: ["data-holder", legalEntityId],
    queryFn: ({ signal }) => getDataHolder(legalEntityId, signal),
  });
  const brand = detail.data?.dataHolderBrands.find((b) => brandIdOf(b) === brandId);
  return (
    <>
      <PageHeader
        title={`Change the details of ${brand?.brandName ?? "a brand"}`}
        description="The Register checks each value. The brand ID, the interim ID and the status do not change."
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && !brand && <ErrorState problem={noBrand} />}
      {detail.data && brand && (
        <BrandForm
          key={`${legalEntityId} ${brandId}`}
          legalEntityId={legalEntityId}
          version={detail.data.version}
          loaded={brand}
        />
      )}
    </>
  );
}
