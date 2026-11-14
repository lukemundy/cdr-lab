import { renderConsole, setUpConsoleTests } from "@cdr-lab/console/test/app";

export { definitionOf } from "@cdr-lab/console/test/app";
import { createMemoryHistory } from "@tanstack/react-router";
import { setupWorker } from "msw/browser";
import { createOpenApiHttp } from "openapi-msw";
import { page } from "vitest/browser";

import type { DataHolderSummary, DataRecipientSummary } from "../api/client";
import type { paths } from "../api/schema.gen";
import { createAppRouter } from "../router";

export const http = createOpenApiHttp<paths>();
export const worker = setupWorker();

setUpConsoleTests(worker);

export function renderApp(url: string) {
  return renderConsole(createAppRouter(createMemoryHistory({ initialEntries: [url] })));
}

// The detail page of a participant with each card open.
export async function renderDetail(url: string) {
  const router = renderApp(url);
  await page.getByRole("button", { name: "Show all brands" }).click();
  return router;
}

export function recipient(name: string, status: DataRecipientSummary["status"] = "ACTIVE") {
  const id = name.toLowerCase().replaceAll(" ", "-");
  return {
    legalEntityId: `id-${id}`,
    legalEntityName: name,
    accreditationNumber: `ADR-${id}`,
    status,
    brandCount: 1,
    softwareProductCount: 2,
  } satisfies DataRecipientSummary;
}

export function holder(name: string, fields: Partial<DataHolderSummary> = {}) {
  return {
    legalEntityId: `id-${name.toLowerCase().replaceAll(" ", "-")}`,
    legalEntityName: name,
    status: "ACTIVE",
    brandCount: 1,
    industries: ["banking"],
    ...fields,
  } satisfies DataHolderSummary;
}
