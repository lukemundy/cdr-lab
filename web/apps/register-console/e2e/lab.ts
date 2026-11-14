import type { Page } from "@playwright/test";

export function publicURL(path: string): string {
  const base = process.env.REGISTER_PUBLIC_URL;
  if (!base)
    throw new Error("Set REGISTER_PUBLIC_URL to the URL of the public APIs of the Register.");
  return new URL(path, base).toString();
}

// The detail page of a participant starts with each brand and software product collapsed.
export async function showAllBrands(page: Page) {
  await page.getByRole("button", { name: "Show all brands" }).click();
}
