import { expect, test, type APIRequestContext } from "@playwright/test";
import { randomUUID } from "node:crypto";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";
import { publicURL, showAllBrands } from "./lab";

type Summary = { data: { dataHolderBrandId?: string }[] };
type Statuses = { data: { legalEntityId: string; status: string }[] };

function hasData(body: unknown): body is { data: unknown[] } {
  return typeof body === "object" && body !== null && "data" in body && Array.isArray(body.data);
}

// The tests check the records that they need, not each field of the CDS schema.
const isSummary = (body: unknown): body is Summary => hasData(body);
const isStatuses = (body: unknown): body is Statuses => hasData(body);

async function cdrGet<T>(
  request: APIRequestContext,
  path: string,
  version: string,
  is: (body: unknown) => body is T,
): Promise<T> {
  const response = await request.get(publicURL(`/cdr-register/v1/all/${path}`), {
    headers: { "x-v": version },
  });
  expect(response.status()).toBe(200);
  const body: unknown = await response.json();
  if (!is(body)) throw new Error(`${path} gave no list`);
  return body;
}

async function brandIdsOfSummary(request: APIRequestContext) {
  const summary = await cdrGet(request, "data-holders/brands/summary", "2", isSummary);
  return summary.data.map((b) => b.dataHolderBrandId);
}

// Each run makes its own data holder and brand through the console API, with a name that sorts after the seed data
// holders that the other tests find on the first page of the list, so the test passes again on the same lab. Get Data
// Holder Brands needs a token, which only the conformance suite can get, so a Go test in the Register checks it.
test("the operator deactivates a data holder brand and removes the data holder, and the CDR APIs show it", async ({
  page,
  request,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  const legalEntityId = randomUUID();
  const brandId = randomUUID();
  const run = legalEntityId.slice(0, 8);
  const holderName = `Test Status Holder ${run} Ltd`;
  const brandName = `Status Test ${run}`;

  const created = await request.post("/api/data-holders", {
    data: {
      legalEntityId,
      legalEntityName: holderName,
      logoUri: "https://status-test.example/logo.png",
      status: "ACTIVE",
    },
  });
  expect(created.status(), await created.text()).toBe(201);
  const added = await request.post(`/api/data-holders/${legalEntityId}/brands`, {
    data: {
      version: 1,
      dataHolderBrandId: brandId,
      brandName,
      industries: ["banking"],
      logoUri: "https://status-test.example/brand.png",
      status: "ACTIVE",
      endpointDetail: {
        version: "1",
        publicBaseUri: "https://api.status-test.example",
        resourceBaseUri: "https://mtls.status-test.example",
        infosecBaseUri: "https://id.status-test.example",
        websiteUri: "https://status-test.example",
      },
      jwksEndpoint: "https://id.status-test.example/jwks",
    },
  });
  expect(added.status(), await added.text()).toBe(201);

  await page.goto(`/data-holders/${legalEntityId}`);

  await showAllBrands(page);

  const actions = (name: string) =>
    page.getByRole("group", { name: `Status actions of ${name}`, exact: true });
  const brand = page.getByRole("region", { name: brandName, exact: true });
  await expect(actions(holderName).getByRole("button")).toHaveText(["Remove"]);
  await expect(actions(brandName).getByRole("button")).toHaveText(["Deactivate", "Remove"]);

  await actions(brandName).getByRole("button", { name: "Deactivate" }).click();
  const deactivation = page.getByRole("alertdialog", { name: `Deactivate ${brandName}?` });
  await expect(deactivation).toHaveAccessibleDescription(
    `${brandName} changes from ACTIVE to INACTIVE.`,
  );
  await deactivation.getByRole("button", { name: "Deactivate", exact: true }).click();

  await expect(deactivation).toBeHidden();
  await expect(page.getByRole("dialog", { name: `${brandName} is now INACTIVE` })).toBeVisible();
  await expect(brand.getByText("INACTIVE", { exact: true })).toBeVisible();
  await expect(actions(brandName).getByRole("button")).toHaveText(["Activate", "Remove"]);
  // Get Data Holder Brands Summary includes INACTIVE brands.
  expect(await brandIdsOfSummary(request)).toContain(brandId);

  await actions(holderName).getByRole("button", { name: "Remove" }).click();
  const removal = page.getByRole("alertdialog", { name: `Remove ${holderName}?` });
  await expect(removal.getByRole("listitem")).toHaveText([
    `${holderName} changes from ACTIVE to REMOVED.`,
    `${brandName} changes from INACTIVE to REMOVED.`,
  ]);
  await expect(removal).toContainText("You cannot reverse this action.");
  await expectNoAxeViolations(page);
  await removal.getByRole("button", { name: "Remove", exact: true }).click();

  await expect(removal).toBeHidden();
  await expect(page.getByRole("dialog", { name: `${holderName} is now REMOVED` })).toBeVisible();
  const legalEntity = page.getByRole("region", { name: "Legal entity", exact: true });
  await expect(legalEntity.getByText("REMOVED", { exact: true })).toBeVisible();
  await expect(brand.getByText("REMOVED", { exact: true })).toBeVisible();
  await expect(actions(holderName).getByRole("button")).toHaveCount(0);
  await expect(actions(brandName).getByRole("button")).toHaveCount(0);

  const statuses = await cdrGet(request, "data-holders/status", "2", isStatuses);
  expect(statuses.data.find((s) => s.legalEntityId === legalEntityId)?.status).toBe("REMOVED");
  expect(await brandIdsOfSummary(request)).not.toContain(brandId);
  expect(errors).toEqual([]);
});
