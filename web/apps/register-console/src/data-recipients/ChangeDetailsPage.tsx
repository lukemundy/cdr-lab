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
  changeDataRecipientDetails,
  getDataRecipient,
  problemOf,
  type DataRecipientDetail,
} from "../api/client";
import { textOf } from "@cdr-lab/app/form-values";
import { levels } from "./levels";
import { useChange } from "../use-change";

const route = getRouteApi("/data-recipients/$legalEntityId/change");

// The form keeps the values and the version that it loaded, also when the query gets the data recipient again.
function DetailsForm(props: { loaded: DataRecipientDetail }) {
  const [loaded] = useState(props.loaded);
  const { legalEntityId, version } = loaded;
  const send = useChange("data-recipient", legalEntityId);

  function submit(data: FormData) {
    const accreditationLevel = levels.find((l) => l.value === textOf(data, "accreditationLevel"));
    if (!accreditationLevel) throw new Error("the form has no accreditation level");
    const details = {
      legalEntityName: textOf(data, "legalEntityName"),
      accreditationNumber: textOf(data, "accreditationNumber"),
      accreditationLevel: accreditationLevel.value,
      logoUri: textOf(data, "logoUri"),
    };
    return send(
      () => changeDataRecipientDetails(legalEntityId, version, details),
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
        cancel={<RouterLink to="/data-recipients/$legalEntityId" params={{ legalEntityId }} />}
        onSubmit={submit}
      >
        <TextField label="Name" name="legalEntityName" defaultValue={loaded.legalEntityName} />
        <TextField
          label="Accreditation number"
          name="accreditationNumber"
          defaultValue={loaded.accreditationNumber}
        />
        <Select
          label="Accreditation level"
          name="accreditationLevel"
          options={levels}
          defaultValue={loaded.accreditationLevel}
        />
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
    queryKey: ["data-recipient", legalEntityId],
    queryFn: ({ signal }) => getDataRecipient(legalEntityId, signal),
  });
  return (
    <>
      <PageHeader
        title={`Change the details of ${detail.data?.legalEntityName ?? "a data recipient"}`}
        description="The Register checks each value. The legal entity ID does not change."
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && <DetailsForm key={legalEntityId} loaded={detail.data} />}
    </>
  );
}
