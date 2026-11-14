import type { StatusBadgeProps } from "@cdr-lab/ui/status-badge";

import type { DataHolderStatus, DataRecipientStatus, Status } from "./api/client";

type Tone = StatusBadgeProps["tone"];

export const dataRecipientTones = {
  ACTIVE: "success",
  SUSPENDED: "warning",
  REVOKED: "danger",
  SURRENDERED: "neutral",
} as const satisfies Record<DataRecipientStatus, Tone>;

export const dataHolderTones = {
  ACTIVE: "success",
  REMOVED: "neutral",
} as const satisfies Record<DataHolderStatus, Tone>;

export const brandAndProductTones = {
  ACTIVE: "success",
  INACTIVE: "warning",
  REMOVED: "neutral",
} as const satisfies Record<Status, Tone>;

const brandAndProductStatuses = [
  "ACTIVE",
  "INACTIVE",
  "REMOVED",
] as const satisfies readonly Status[];

export const brandAndProductStatusOptions = brandAndProductStatuses.map((status) => ({
  value: status,
  label: status,
}));

// The form sends the value of its status select, which is one of the options.
export function brandAndProductStatusOf(value: string): Status {
  const status = brandAndProductStatuses.find((s) => s === value);
  if (!status) throw new Error("the form has no status");
  return status;
}
