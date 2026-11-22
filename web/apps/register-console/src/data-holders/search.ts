import type { DataHolderQuery, DataHolderStatus } from "../api/client";
import { listSearchValidator, oneOf, pageSize, type ListSearch } from "@cdr-lab/app/list-search";

export const statuses = ["ACTIVE", "REMOVED"] as const satisfies readonly DataHolderStatus[];

const sortColumns = [
  "legalEntityName",
  "status",
  "brandCount",
] as const satisfies readonly NonNullable<DataHolderQuery["sort"]>[];

export type SortColumn = (typeof sortColumns)[number];

export type DataHolderSearch = ListSearch<DataHolderStatus, SortColumn>;

export function sortColumnOf(column: string): SortColumn | undefined {
  return oneOf(sortColumns, column);
}

export const validateDataHolderSearch = listSearchValidator(statuses, sortColumns);

export function apiQuery(search: DataHolderSearch): DataHolderQuery {
  return { ...search, pageSize };
}
