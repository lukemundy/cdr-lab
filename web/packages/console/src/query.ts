import { QueryClient } from "@tanstack/react-query";

import { ProblemError } from "./problem";

const maxRetries = 2;

// A request that the API refused gives the same answer again, so only a server or network failure is tried again.
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: (failures, error) =>
          failures < maxRetries && !(error instanceof ProblemError && error.status < 500),
      },
    },
  });
}
