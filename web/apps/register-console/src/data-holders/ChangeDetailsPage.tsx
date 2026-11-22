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
  changeDataHolderDetails,
  getDataHolder,
  problemOf,
  type DataHolderDetail,
} from "../api/client";
import { textOf } from "@cdr-lab/web-platform/form-values";
import { useChange } from "../use-change";

const route = getRouteApi("/data-holders/$legalEntityId/change");

// The form keeps the values and the version that it loaded, also when the query gets the data holder again.
function DetailsForm(props: { loaded: DataHolderDetail }) {
  const [loaded] = useState(props.loaded);
  const { legalEntityId, version } = loaded;
  const send = useChange("data-holder", legalEntityId);

  function submit(data: FormData) {
    const details = {
      legalEntityName: textOf(data, "legalEntityName"),
      logoUri: textOf(data, "logoUri"),
    };
    return send(
      () => changeDataHolderDetails(legalEntityId, version, details),
      details.legalEntityName,
    );
  }

  return (
    <>
      <DescriptionList>
        <DescriptionItem term="Legal entity ID">
          <Identifier
            value={legalEntityId}
            label={`legal entity ID of ${loaded.legalEntityName}`}
          />
        </DescriptionItem>
      </DescriptionList>
      <Form
        label="Change details"
        submitLabel="Save changes"
        cancel={<RouterLink to="/data-holders/$legalEntityId" params={{ legalEntityId }} />}
        onSubmit={submit}
      >
        <TextField label="Name" name="legalEntityName" defaultValue={loaded.legalEntityName} />
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

export function ChangeDetailsPage() {
  const { legalEntityId } = route.useParams();
  const detail = useQuery({
    queryKey: ["data-holder", legalEntityId],
    queryFn: ({ signal }) => getDataHolder(legalEntityId, signal),
  });
  return (
    <>
      <PageHeader
        title={`Change the details of ${detail.data?.legalEntityName ?? "a data holder"}`}
        description="The Register checks each value. The legal entity ID does not change."
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && <DetailsForm key={legalEntityId} loaded={detail.data} />}
    </>
  );
}
