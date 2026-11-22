import type { RegistrationQuery } from "../api/client";
import { listSearchValidator, oneOf, type ListSearch } from "@cdr-lab/app/list-search";

const sortColumns = [
  "clientName",
  "registeredAt",
  "currentArrangementCount",
  "arrangementCount",
] as const satisfies readonly NonNullable<RegistrationQuery["sort"]>[];

export type SortColumn = (typeof sortColumns)[number];

// The API gives all the registrations in one response, with no search.
export type RegistrationSearch = Pick<ListSearch<never, SortColumn>, "sort" | "order">;

export function sortColumnOf(column: string): SortColumn | undefined {
  return oneOf(sortColumns, column);
}

const validateListSearch = listSearchValidator<never, SortColumn>([], sortColumns);

export function validateRegistrationSearch(raw: Record<string, unknown>): RegistrationSearch {
  const { sort, order } = validateListSearch(raw);
  return { ...(sort && { sort }), ...(order && { order }) };
}
