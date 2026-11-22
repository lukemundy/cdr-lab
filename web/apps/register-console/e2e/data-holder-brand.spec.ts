import { expect, test, type Page } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";
import { publicURL } from "./lab";

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

type Summary = { data: { dataHolderBrandId?: string; brandName: string; industries: string[] }[] };

// The test checks the records that it needs, not each field of the CDS schema.
function isSummary(body: unknown): body is Summary {
  return typeof body === "object" && body !== null && "data" in body && Array.isArray(body.data);
}

async function fill(page: Page, values: Record<string, string>) {
  for (const [name, value] of Object.entries(values)) {
    // oxlint-disable-next-line no-await-in-loop -- a fill types into the field that has the focus, so the fills cannot run at the same time.
    await page.getByRole("textbox", { name, exact: true }).fill(value);
  }
}

// Each run makes its own data holder with the UUIDs of the forms, so the test passes again on the same lab. The name
// sorts after the seed data holders that the other tests find on the first page of the list. Get Data Holder Brands
// needs a token, which only the conformance suite can get, so a Go test in the Register checks it.
test("the operator adds a brand to a data holder and changes it, and Get Data Holder Brands Summary shows it", async ({
  page,
  request,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));

  await page.goto("/create/data-holder");
  const legalEntityId = await page.getByRole("textbox", { name: "Legal entity ID" }).inputValue();
  const run = legalEntityId.slice(0, 8);
  const holderName = `Test Brand Holder ${run} Ltd`;
  await fill(page, { Name: holderName, "Logo URI": "https://brand-test.example/logo.png" });
  await page.getByRole("button", { name: "Create data holder" }).click();
  await expect(page.getByRole("heading", { level: 1, name: holderName })).toBeFocused();

  await page.getByRole("link", { name: "Add brand" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(`Add a brand to ${holderName}`);
  const brandIdField = page.getByRole("textbox", { name: "Brand ID (optional)" });
  await expect(brandIdField).toHaveValue(uuid);
  const brandId = await brandIdField.inputValue();
  await expect(page.getByRole("textbox", { name: "Interim ID (optional)" })).toHaveValue("");
  await expect(page.getByRole("combobox", { name: "Status" })).toHaveValue("ACTIVE");
  await expectNoAxeViolations(page);

  const brandName = `Brand Test ${run}`;
  await fill(page, {
    Name: brandName,
    "Logo URI": "https://brand-test.example/brand.png",
    "Version (optional)": "1",
    "Public base URI": "https://api.brand-test.example",
    "Resource base URI (optional)": "https://mtls.brand-test.example",
    "Infosec base URI (optional)": "https://id.brand-test.example",
    "Website URI (optional)": "https://brand-test.example",
    "JWKS endpoint (optional)": "https://id.brand-test.example/jwks",
  });
  await page
    .getByRole("group", { name: "Industries" })
    .getByRole("checkbox", { name: "Banking" })
    .check();
  await page.getByRole("button", { name: "Add brand" }).click();

  await expect(page.getByRole("heading", { level: 1, name: holderName })).toBeFocused();
  await expect(page.getByRole("dialog", { name: `Brand ${brandName} added` })).toBeVisible();
  const brand = page.getByRole("region", { name: brandName, exact: true });
  await expect(brand.getByText(brandId, { exact: true })).toBeVisible();

  await brand.getByRole("link", { name: "Change brand" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    `Change the details of ${brandName}`,
  );
  const publicBaseUri = page
    .getByRole("group", { name: "Endpoints" })
    .getByRole("textbox", { name: "Public base URI", exact: true });
  await expect(publicBaseUri).toHaveValue("https://api.brand-test.example");
  await publicBaseUri.fill("");
  await page.getByRole("button", { name: "Save changes" }).click();

  const required = "endpointDetail.publicBaseUri is required";
  await expect(page.getByRole("alert").getByRole("link", { name: required })).toBeVisible();
  await expect(publicBaseUri).toHaveAccessibleDescription(required);
  await expectNoAxeViolations(page);

  const changedName = `Brand Changed ${run}`;
  await publicBaseUri.fill("https://api2.brand-test.example");
  await fill(page, { Name: changedName });
  await page
    .getByRole("group", { name: "Industries" })
    .getByRole("checkbox", { name: "Energy" })
    .check();
  await page.getByRole("button", { name: "Save changes" }).click();

  await expect(page.getByRole("heading", { level: 1, name: holderName })).toBeFocused();
  await expect(
    page.getByRole("dialog", { name: `Details of ${changedName} changed` }),
  ).toBeVisible();
  await expect(page.getByRole("region", { name: changedName, exact: true })).toBeVisible();
  await expectNoAxeViolations(page);

  const response = await request.get(
    publicURL("/cdr-register/v1/all/data-holders/brands/summary"),
    { headers: { "x-v": "2" } },
  );
  expect(response.status()).toBe(200);
  const summary: unknown = await response.json();
  if (!isSummary(summary)) throw new Error("Get Data Holder Brands Summary gave no list");
  expect(summary.data.find((b) => b.dataHolderBrandId === brandId)).toMatchObject({
    brandName: changedName,
    industries: ["banking", "energy"],
  });
  expect(errors).toEqual([]);
});
