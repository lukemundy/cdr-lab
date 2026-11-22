import type { Sort } from "@cdr-lab/ui/data-table";

const orders = ["asc", "desc"] as const;

export type Order = (typeof orders)[number];

// The URL search of a list has the names and values of the API query, without the page size.
export type ListSearch<Status extends string, Column extends string> = {
  status?: Status;
  q?: string;
  sort?: Column;
  order?: Order;
  page?: number;
};

export type ListSearchChange<Search> = { [K in keyof Search]?: Search[K] | undefined };

export const pageSize = 25;

export function oneOf<T extends string>(values: readonly T[], value: unknown): T | undefined {
  return values.find((v) => v === value);
}

// A value that is not valid is removed, so that a changed or old link still shows a list.
export function listSearchValidator<Status extends string, Column extends string>(
  statuses: readonly Status[],
  columns: readonly Column[],
) {
  return (raw: Record<string, unknown>): ListSearch<Status, Column> => {
    const search: ListSearch<Status, Column> = {};
    const status = oneOf(statuses, raw["status"]);
    const sort = oneOf(columns, raw["sort"]);
    const order = oneOf(orders, raw["order"]);
    const page = Number(raw["page"]);

    if (status) search.status = status;
    if (typeof raw["q"] === "string" && raw["q"] !== "") search.q = raw["q"];
    if (sort) search.sort = sort;
    if (order) search.order = order;
    if (Number.isInteger(page) && page > 1) search.page = page;
    return search;
  };
}

export type PageSearch = { page?: number };

// A page that is not valid is removed, so that a changed or old link still shows the page list.
export function validatePageSearch(raw: Record<string, unknown>): PageSearch {
  const page = Number(raw["page"]);
  return Number.isInteger(page) && page > 1 ? { page } : {};
}

// A change of the filter, the search or the sort goes back to page 1, and the validation gives one parameter order.
export function changedSearch<Search extends ListSearch<string, string>>(
  validate: (raw: Record<string, unknown>) => Search,
  search: Search,
  change: ListSearchChange<Search>,
): Search {
  return validate({ ...search, page: undefined, ...change });
}

export function tableSort(search: ListSearch<string, string>, defaultColumn: string): Sort {
  return {
    column: search.sort ?? defaultColumn,
    direction: search.order === "desc" ? "descending" : "ascending",
  };
}

export function orderOf(sort: Sort): Order {
  return sort.direction === "descending" ? "desc" : "asc";
}
