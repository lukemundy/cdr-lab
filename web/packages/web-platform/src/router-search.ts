// The URL search is plain URLSearchParams, with the same names and values as the console API query.
export function parseSearch(search: string): Record<string, string> {
  return Object.fromEntries(new URLSearchParams(search));
}

export function stringifySearch(search: Record<string, unknown>): string {
  const params = new URLSearchParams();
  for (const [name, value] of Object.entries(search)) {
    if ((typeof value === "string" && value !== "") || typeof value === "number") {
      params.set(name, String(value));
    }
  }
  const text = params.toString();
  return text ? `?${text}` : "";
}
