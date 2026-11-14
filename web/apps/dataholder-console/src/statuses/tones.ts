import type { StatusBadgeProps } from "@cdr-lab/ui/status-badge";

import type { SoftwareProductStatus } from "../api/client";

export const statusTones = {
  ACTIVE: "success",
  INACTIVE: "warning",
  REMOVED: "neutral",
} as const satisfies Record<SoftwareProductStatus["status"], StatusBadgeProps["tone"]>;
