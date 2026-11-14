import { useQuery } from "@tanstack/react-query";

import { listConsents } from "../api/client";

// An action on a consent, for example a withdrawal, invalidates this key.
export const consentsKey = ["consents"];

export function useConsents() {
  return useQuery({ queryKey: consentsKey, queryFn: ({ signal }) => listConsents(signal) });
}
