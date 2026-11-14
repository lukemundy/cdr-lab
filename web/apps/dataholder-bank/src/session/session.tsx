import { ProblemError } from "@cdr-lab/console/problem";
import { LoadingState } from "@cdr-lab/ui/states";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Navigate } from "@tanstack/react-router";
import { type ReactNode, useEffect } from "react";

import { getSession, type SessionCustomer, sessionEndedType } from "../api/client";

export const sessionKey = ["session"];

// The site API gives 403 with no valid session, so the query does not try again (createQueryClient).
export function useSession() {
  return useQuery({ queryKey: sessionKey, queryFn: ({ signal }) => getSession(signal) });
}

// The data holder gives a problem type of its own when the session of the browser ended.
export function isSessionEnded(error: unknown): boolean {
  return (
    error instanceof ProblemError &&
    "type" in error.problem &&
    error.problem.type === sessionEndedType
  );
}

// A page with the data of a customer needs a session. With no valid session, the site goes to the sign-in page.
export function SignedIn({ children }: { children: (customer: SessionCustomer) => ReactNode }) {
  const session = useSession();

  if (session.isPending) return <LoadingState label="Loading" />;
  if (session.isError) return <Navigate to="/sign-in" replace />;
  return children(session.data);
}

const forbidden = 403;

// A 403 for the data of a customer means that the session ended after SignedIn read it. The site then reads the
// session again, and SignedIn goes to the sign-in page. It gives true while the session is read again.
export function useSessionEndedBy(...errors: unknown[]): boolean {
  const queryClient = useQueryClient();
  const ended = errors.some((error) => error instanceof ProblemError && error.status === forbidden);

  useEffect(() => {
    if (ended) void queryClient.invalidateQueries({ queryKey: sessionKey });
  }, [ended, queryClient]);

  return ended;
}
