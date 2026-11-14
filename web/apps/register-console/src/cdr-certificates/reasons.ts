import type { RevocationReason } from "../api/client";

// The reasons of the operator interface in words (Register DESIGN.md, Operator console).
export const reasonWords = {
  unspecified: "Unspecified",
  keyCompromise: "Key compromise",
  affiliationChanged: "Affiliation changed",
  superseded: "Superseded",
  cessationOfOperation: "Cessation of operation",
  privilegeWithdrawn: "Privilege withdrawn",
} as const satisfies Record<RevocationReason, string>;
