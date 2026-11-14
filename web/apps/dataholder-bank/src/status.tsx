import { useRouterState } from "@tanstack/react-router";
import { useEffect, useRef, type ReactNode } from "react";

declare module "@tanstack/react-router" {
  interface HistoryState {
    // The result of the action that went to the page, for example a close of an account.
    status?: string;
  }
}

// PageHeading takes the focus and shows the status message after an action goes to the page (WCAG 2.2, SC 4.1.3).
export function PageHeading({ children }: { children: ReactNode }) {
  const status = useRouterState({ select: (state) => state.location.state.status });
  const heading = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    if (status) heading.current?.focus();
  }, [status]);

  return (
    <>
      <h1 ref={heading} tabIndex={-1} className="outline-none">
        {children}
      </h1>
      {status && <output className="block font-semibold">{status}</output>}
    </>
  );
}
