import type { RevocationReason } from "../api/client";

// The reasons of the operator interface in words (services/register/design/operator-console-actions.md).
export const reasonWords = {
  unspecified: "Unspecified",
  keyCompromise: "Key compromise",
  affiliationChanged: "Affiliation changed",
  superseded: "Superseded",
  cessationOfOperation: "Cessation of operation",
  privilegeWithdrawn: "Privilege withdrawn",
} as const satisfies Record<RevocationReason, string>;
