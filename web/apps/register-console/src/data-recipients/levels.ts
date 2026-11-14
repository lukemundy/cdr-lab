import type { AccreditationLevel } from "../api/client";

export const levels: { value: AccreditationLevel; label: string }[] = [
  { value: "UNRESTRICTED", label: "UNRESTRICTED" },
  { value: "SPONSORED", label: "SPONSORED" },
];
