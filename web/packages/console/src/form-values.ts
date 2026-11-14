// The console removes the white space at the start and at the end of each value.
export function textOf(data: FormData, name: string): string {
  const value = data.get(name);
  return typeof value === "string" ? value.trim() : "";
}

export function linesOf(data: FormData, name: string): string[] {
  return textOf(data, name)
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line !== "");
}
