import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Form } from "@cdr-lab/ui/form";
import { TextField } from "@cdr-lab/ui/form-field";
import { Identifier } from "@cdr-lab/ui/identifier";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";
import { useState } from "react";

import {
  changeDataRecipientBrandDetails,
  getDataRecipient,
  problemOf,
  type DataRecipientBrand,
} from "../api/client";
import { textOf } from "@cdr-lab/web-platform/form-values";
import { useChange } from "../use-change";

const route = getRouteApi("/data-recipients/$legalEntityId/brands/$brandId/change");

const noBrand = {
  title: "The brand does not exist",
  detail: "The data recipient has no brand with this ID.",
};

type BrandFormProps = { legalEntityId: string; version: number; loaded: DataRecipientBrand };

// The form keeps the values and the version that it loaded, also when the query gets the data recipient again.
function BrandForm(props: BrandFormProps) {
  const [{ legalEntityId, version, loaded }] = useState(props);
  const brandId = loaded.dataRecipientBrandId;
  const send = useChange("data-recipient", legalEntityId);

  function submit(data: FormData) {
    const details = { brandName: textOf(data, "brandName"), logoUri: textOf(data, "logoUri") };
    return send(
      () => changeDataRecipientBrandDetails(legalEntityId, brandId, version, details),
      details.brandName,
      brandId,
    );
  }

  return (
    <>
      <DescriptionList>
        <DescriptionItem term="Brand ID">
          <Identifier value={brandId} label={`brand ID of ${loaded.brandName}`} />
        </DescriptionItem>
      </DescriptionList>
      <Form
        label="Change brand"
        submitLabel="Save changes"
        cancel={<RouterLink to="/data-recipients/$legalEntityId" params={{ legalEntityId }} />}
        onSubmit={submit}
      >
        <TextField label="Name" name="brandName" defaultValue={loaded.brandName} />
        <TextField
          label="Logo URI"
          name="logoUri"
          spellCheck={false}
          defaultValue={loaded.logoUri}
        />
      </Form>
    </>
  );
}

export function ChangeBrandPage() {
  const { legalEntityId, brandId } = route.useParams();
  const detail = useQuery({
    queryKey: ["data-recipient", legalEntityId],
    queryFn: ({ signal }) => getDataRecipient(legalEntityId, signal),
  });
  const brand = detail.data?.dataRecipientBrands.find((b) => b.dataRecipientBrandId === brandId);
  return (
    <>
      <PageHeader
        title={`Change the details of ${brand?.brandName ?? "a brand"}`}
        description="The Register checks each value. The brand ID, the status and the software products do not change."
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
