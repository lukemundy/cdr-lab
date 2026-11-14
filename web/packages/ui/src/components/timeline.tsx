import type { ReactNode } from "react";

import { Pagination, type PaginationProps } from "./pagination";
import { EmptyState, ErrorState, LoadingState, type ListStatus } from "./states";
import { LocalTime } from "./local-time";

export type TimelineEvent = {
  id: string;
  // An ISO 8601 time.
  time: string;
  title: string;
  details?: ReactNode;
};

// The timeline does not sort: the app gives the events newest first.
export type TimelineProps = PaginationProps & {
  label: string;
  events: TimelineEvent[];
  empty: { title: string; description?: string };
  status?: ListStatus;
};

const ready: ListStatus = { state: "ready" };

export function Timeline({ label, events, empty, status = ready, ...pagination }: TimelineProps) {
  function body() {
    if (status.state === "loading") return <LoadingState />;
    if (status.state === "error") return <ErrorState problem={status.problem} />;
    if (events.length === 0) return <EmptyState {...empty} />;
    return (
      <ol aria-label={label} className="flex flex-col">
        {events.map((event) => (
          <li
            key={event.id}
            className="relative flex flex-col gap-1 border-l border-border pb-4 pl-4 before:absolute before:top-1.5 before:-left-1 before:size-2 before:rounded-full before:bg-border-strong last:pb-0"
          >
            <span className="text-sm text-fg-muted tabular-nums">
              <LocalTime value={event.time} seconds />
            </span>
            <p className="font-medium text-fg">{event.title}</p>
            {event.details && <div className="text-sm">{event.details}</div>}
          </li>
        ))}
      </ol>
    );
  }

  return (
    <div>
      {body()}
      <Pagination {...pagination} />
    </div>
  );
}
