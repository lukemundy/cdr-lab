import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "./button";

export type PaginationProps = {
  // The first page is 1.
  page: number;
  pageCount: number;
  rowCount: number;
  onPageChange: (page: number) => void;
};

const plural = new Intl.PluralRules("en-AU");
const numberFormat = new Intl.NumberFormat("en-AU");

export function Pagination({ page, pageCount, rowCount, onPageChange }: PaginationProps) {
  const lastPage = Math.max(pageCount, 1);

  return (
    <nav
      aria-label="Pagination"
      className="flex items-center justify-between gap-4 pt-4 text-sm text-fg-muted"
    >
      {/* A status message, so that a screen reader announces the result of a new search or filter. */}
      <output className="tabular-nums">
        {numberFormat.format(rowCount)} {plural.select(rowCount) === "one" ? "record" : "records"}
      </output>
      <div className="flex items-center gap-2">
        <Button variant="secondary" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>
          <ChevronLeft aria-hidden="true" />
          Previous<span className="sr-only"> page</span>
        </Button>
        <output className="tabular-nums">
          Page {page} of {lastPage}
        </output>
        <Button
          variant="secondary"
          disabled={page >= lastPage}
          onClick={() => onPageChange(page + 1)}
        >
          Next<span className="sr-only"> page</span>
          <ChevronRight aria-hidden="true" />
        </Button>
      </div>
    </nav>
  );
}
