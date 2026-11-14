import { TextArea, TextField } from "@cdr-lab/ui/form-field";

import type { SoftwareProductChange } from "../api/client";
import { linesOf, textOf } from "@cdr-lab/console/form-values";

function uri(label: string, name: keyof SoftwareProductChange, value: string | undefined) {
  return <TextField label={label} name={name} spellCheck={false} defaultValue={value ?? ""} />;
}

// The fields that a change replaces, which are also the fields of an add except the ID and the status.
export function SoftwareProductFields({ loaded }: { loaded?: SoftwareProductChange | undefined }) {
  return (
    <>
      <TextField
        label="Name"
        name="softwareProductName"
        defaultValue={loaded?.softwareProductName}
      />
      <TextArea
        label="Description"
        name="softwareProductDescription"
        rows={3}
        defaultValue={loaded?.softwareProductDescription}
      />
      {uri("Logo URI", "logoUri", loaded?.logoUri)}
      {uri("Client URI", "clientUri", loaded?.clientUri)}
      {uri("Sector identifier URI (optional)", "sectorIdentifierUri", loaded?.sectorIdentifierUri)}
      <TextArea
        label="Redirect URIs"
        name="redirectUris"
        description="Give one URI on each line."
        spellCheck={false}
        rows={4}
        defaultValue={loaded?.redirectUris.join("\n")}
      />
      {uri("Terms of service URI (optional)", "tosUri", loaded?.tosUri)}
      {uri("Policy URI (optional)", "policyUri", loaded?.policyUri)}
      {uri("JWKS URI", "jwksUri", loaded?.jwksUri)}
      {uri("Revocation URI", "revocationUri", loaded?.revocationUri)}
      {uri("Recipient base URI", "recipientBaseUri", loaded?.recipientBaseUri)}
      <TextArea
        label="Scopes"
        name="scopes"
        description="Give one scope on each line."
        spellCheck={false}
        rows={6}
        defaultValue={loaded?.scopes.join("\n")}
      />
    </>
  );
}

// An optional field with no value is not in the request, so it has no value after a change.
export function softwareProductDetailsOf(data: FormData): SoftwareProductChange {
  const sectorIdentifierUri = textOf(data, "sectorIdentifierUri");
  const tosUri = textOf(data, "tosUri");
  const policyUri = textOf(data, "policyUri");
  return {
    softwareProductName: textOf(data, "softwareProductName"),
    softwareProductDescription: textOf(data, "softwareProductDescription"),
    logoUri: textOf(data, "logoUri"),
    clientUri: textOf(data, "clientUri"),
    ...(sectorIdentifierUri !== "" && { sectorIdentifierUri }),
    redirectUris: linesOf(data, "redirectUris"),
    ...(tosUri !== "" && { tosUri }),
    ...(policyUri !== "" && { policyUri }),
    jwksUri: textOf(data, "jwksUri"),
    revocationUri: textOf(data, "revocationUri"),
    recipientBaseUri: textOf(data, "recipientBaseUri"),
    scopes: linesOf(data, "scopes"),
  };
}
