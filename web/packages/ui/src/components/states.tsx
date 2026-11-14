import { LoaderCircle } from "lucide-react";

import { Alert } from "./alert";

export type Problem = { title: string; detail?: string };

export type ListStatus =
  | { state: "ready" }
  | { state: "loading" }
  | { state: "error"; problem: Problem };

export function EmptyState({ title, description }: { title: string; description?: string }) {
  return (
    <div className="flex flex-col items-center gap-1 py-8 text-center">
      <p className="font-medium text-fg">{title}</p>
      {description && <p className="text-sm text-fg-muted">{description}</p>}
    </div>
  );
}

export function LoadingState({ label = "Loading" }: { label?: string }) {
  return (
    <output className="flex items-center justify-center gap-2 py-8 text-sm text-fg-muted">
      <LoaderCircle aria-hidden="true" className="size-4 motion-safe:animate-spin" />
      {label}
    </output>
  );
}

// The title and the detail of an RFC 9457 problem details response.
export function ErrorState({ problem }: { problem: Problem }) {
  return <Alert title={problem.title}>{problem.detail && <p>{problem.detail}</p>}</Alert>;
}
