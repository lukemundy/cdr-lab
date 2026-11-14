import type { SoftwareProductStatusQuery } from "../api/client";
import { listSearchValidator, oneOf, type Order } from "@cdr-lab/console/list-search";

type Registration = NonNullable<SoftwareProductStatusQuery["registration"]>;

export const registrations = [
  "registered",
  "notRegistered",
] as const satisfies readonly Registration[];

const sortColumns = ["name", "status"] as const satisfies readonly NonNullable<
  SoftwareProductStatusQuery["sort"]
>[];

export type SortColumn = (typeof sortColumns)[number];

// The list has one page and no search.
export type StatusSearch = { registration?: Registration; sort?: SortColumn; order?: Order };

export function sortColumnOf(column: string): SortColumn | undefined {
  return oneOf(sortColumns, column);
}

const validateSort = listSearchValidator<never, SortColumn>([], sortColumns);

export function validateStatusSearch(raw: Record<string, unknown>): StatusSearch {
  const { sort, order } = validateSort(raw);
  const registration = oneOf(registrations, raw["registration"]);
  return {
    ...(registration && { registration }),
    ...(sort && { sort }),
    ...(order && { order }),
  };
}
