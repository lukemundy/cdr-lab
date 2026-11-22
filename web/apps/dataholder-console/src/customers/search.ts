import type { CustomerQuery } from "../api/client";
import {
  listSearchValidator,
  oneOf,
  pageSize,
  type ListSearch,
} from "@cdr-lab/web-platform/list-search";

const sortColumns = [
  "lastName",
  "customerNumber",
  "lastUpdated",
  "openAccountCount",
  "currentArrangementCount",
] as const satisfies readonly NonNullable<CustomerQuery["sort"]>[];

export type SortColumn = (typeof sortColumns)[number];

export type CustomerSearch = ListSearch<never, SortColumn>;

export function sortColumnOf(column: string): SortColumn | undefined {
  return oneOf(sortColumns, column);
}

export const validateCustomerSearch = listSearchValidator<never, SortColumn>([], sortColumns);

export function apiQuery(search: CustomerSearch): CustomerQuery {
  return { ...search, pageSize };
}
