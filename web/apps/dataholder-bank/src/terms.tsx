import type { ReactNode } from "react";

export function Terms({ children }: { children: ReactNode }) {
  return <dl className="m-0 flex flex-col gap-1">{children}</dl>;
}

export function Term({ term, children }: { term: string; children: ReactNode }) {
  return (
    <div className="flex flex-wrap gap-x-2">
      <dt className="text-fg-muted">{term}:</dt>
      <dd className="m-0">{children}</dd>
    </div>
  );
}
