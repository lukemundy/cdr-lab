import { CheckboxGroup } from "@cdr-lab/ui/checkbox-group";
import { Fieldset } from "@cdr-lab/ui/fieldset";
import { TextField } from "@cdr-lab/ui/form-field";

import type { DataHolderBrandChange, Industry } from "../api/client";
import { textOf } from "@cdr-lab/app/form-values";

const industryOptions = [
  { value: "banking", label: "Banking" },
  { value: "energy", label: "Energy" },
  { value: "non-bank-lending", label: "Non-bank lending" },
  { value: "telco", label: "Telco" },
] as const satisfies readonly { value: Industry; label: string }[];

type Endpoint = keyof DataHolderBrandChange["endpointDetail"];

function uri(label: string, name: string, value: string | undefined) {
  return <TextField label={label} name={name} spellCheck={false} defaultValue={value ?? ""} />;
}

function endpoint(label: string, name: Endpoint, loaded: DataHolderBrandChange | undefined) {
  return uri(label, `endpointDetail/${name}`, loaded?.endpointDetail[name]);
}

// The fields that a change replaces, which are also the fields of an add except the identifiers and the status. The
// name of each endpoint field is the JSON Pointer path of its value, so that its field error shows at the field.
export function BrandFields({ loaded }: { loaded?: DataHolderBrandChange | undefined }) {
  return (
    <>
      <TextField label="Name" name="brandName" defaultValue={loaded?.brandName} />
      <TextField
        label="Brand group (optional)"
        name="brandGroup"
        defaultValue={loaded?.brandGroup}
      />
      <CheckboxGroup
        label="Industries"
        name="industries"
        options={industryOptions}
        defaultValue={loaded?.industries}
      />
      {uri("Logo URI", "logoUri", loaded?.logoUri)}
      <Fieldset legend="Endpoints">
        <TextField
          label="Version (optional)"
          name="endpointDetail/version"
          defaultValue={loaded?.endpointDetail.version}
        />
        {endpoint("Public base URI", "publicBaseUri", loaded)}
        {endpoint("Product base URI (optional)", "productBaseUri", loaded)}
        {endpoint("Resource base URI (optional)", "resourceBaseUri", loaded)}
        {endpoint("Infosec base URI (optional)", "infosecBaseUri", loaded)}
        {endpoint("Extension base URI (optional)", "extensionBaseUri", loaded)}
        {endpoint("Website URI (optional)", "websiteUri", loaded)}
      </Fieldset>
      {uri("JWKS endpoint (optional)", "jwksEndpoint", loaded?.jwksEndpoint)}
    </>
  );
}

const optionalEndpoints = [
  "version",
  "productBaseUri",
  "resourceBaseUri",
  "infosecBaseUri",
  "extensionBaseUri",
  "websiteUri",
] as const satisfies readonly Endpoint[];

// An optional field with no value is not in the request, so it has no value after a change.
export function brandDetailsOf(data: FormData): DataHolderBrandChange {
  const checked = data.getAll("industries");
  const brandGroup = textOf(data, "brandGroup");
  const jwksEndpoint = textOf(data, "jwksEndpoint");
  const endpointDetail: DataHolderBrandChange["endpointDetail"] = {
    publicBaseUri: textOf(data, "endpointDetail/publicBaseUri"),
  };
  for (const key of optionalEndpoints) {
    const value = textOf(data, `endpointDetail/${key}`);
    if (value !== "") endpointDetail[key] = value;
  }
  return {
    brandName: textOf(data, "brandName"),
    ...(brandGroup !== "" && { brandGroup }),
    industries: industryOptions.map((o) => o.value).filter((i) => checked.includes(i)),
    logoUri: textOf(data, "logoUri"),
    endpointDetail,
    ...(jwksEndpoint !== "" && { jwksEndpoint }),
  };
}
