import type { DataRecipientQuery, DataRecipientStatus } from "../api/client";
import { listSearchValidator, oneOf, pageSize, type ListSearch } from "@cdr-lab/app/list-search";

export const statuses = [
  "ACTIVE",
  "SUSPENDED",
  "REVOKED",
  "SURRENDERED",
] as const satisfies readonly DataRecipientStatus[];

const sortColumns = [
  "legalEntityName",
  "accreditationNumber",
  "status",
  "brandCount",
  "softwareProductCount",
] as const satisfies readonly NonNullable<DataRecipientQuery["sort"]>[];

export type SortColumn = (typeof sortColumns)[number];

export type DataRecipientSearch = ListSearch<DataRecipientStatus, SortColumn>;

export function sortColumnOf(column: string): SortColumn | undefined {
  return oneOf(sortColumns, column);
}

export const validateDataRecipientSearch = listSearchValidator(statuses, sortColumns);

export function apiQuery(search: DataRecipientSearch): DataRecipientQuery {
  return { ...search, pageSize };
}
