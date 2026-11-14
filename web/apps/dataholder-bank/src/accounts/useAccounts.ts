import { useQuery } from "@tanstack/react-query";

import { listAccounts } from "../api/client";

export function useAccounts() {
  return useQuery({ queryKey: ["accounts"], queryFn: ({ signal }) => listAccounts(signal) });
}
