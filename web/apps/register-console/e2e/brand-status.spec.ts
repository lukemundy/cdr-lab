import { expect, test, type APIRequestContext } from "@playwright/test";
import { randomUUID } from "node:crypto";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";
import { publicURL, showAllBrands } from "./lab";

type DataRecipients = {
  data: {
    legalEntityId: string;
    dataRecipientBrands: {
      dataRecipientBrandId: string;
      status: string;
      softwareProducts: { softwareProductId: string; status: string }[];
    }[];
  }[];
};

type StatusList = { data: { softwareProductId: string; status: string }[] };

function hasData(body: unknown): body is { data: unknown[] } {
  return typeof body === "object" && body !== null && "data" in body && Array.isArray(body.data);
}

// The tests check the records that they need, not each field of the CDS schema.
const isDataRecipients = (body: unknown): body is DataRecipients => hasData(body);
const isStatusList = (body: unknown): body is StatusList => hasData(body);

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

async function versionOf(request: APIRequestContext, legalEntityId: string): Promise<number> {
  const response = await request.get(`/api/data-recipients/${legalEntityId}`);
  expect(response.status()).toBe(200);
  const body: unknown = await response.json();
  if (typeof body !== "object" || body === null || !("version" in body)) {
    throw new Error("the detail gave no version");
  }
  return Number(body.version);
}

async function expectCreated(
  request: APIRequestContext,
  path: string,
  data: Record<string, unknown>,
) {
  const response = await request.post(path, { data });
  expect(response.status(), await response.text()).toBe(201);
}

// Each run makes its own data recipient, whose name sorts after the first page of the list, so the
// test can run again.
test("the operator deactivates a brand, and the confirmation and the CDR APIs show its software products", async ({
  page,
  request,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  const legalEntityId = randomUUID();
  const brandId = randomUUID();
  const productId = randomUUID();
  const run = legalEntityId.slice(0, 8);
  const brandName = `Brand Status Test ${run}`;
  const productName = `Brand Status Test ${run} Web`;

  await expectCreated(request, "/api/data-recipients", {
    legalEntityId,
    legalEntityName: `Operator Brand Status Test ${run} Pty Ltd`,
    accreditationNumber: `ADR-BRAND-${run}`,
    accreditationLevel: "UNRESTRICTED",
    logoUri: "https://brand-status-test.example/logo.png",
    status: "ACTIVE",
  });
  await expectCreated(request, `/api/data-recipients/${legalEntityId}/brands`, {
    version: await versionOf(request, legalEntityId),
    dataRecipientBrandId: brandId,
    brandName,
    logoUri: "https://brand-status-test.example/brand.png",
    status: "ACTIVE",
  });
  await expectCreated(
    request,
    `/api/data-recipients/${legalEntityId}/brands/${brandId}/software-products`,
    {
      version: await versionOf(request, legalEntityId),
      softwareProductId: productId,
      softwareProductName: productName,
      softwareProductDescription: "A software product of the brand status test",
      logoUri: "https://brand-status-test.example/app.png",
      clientUri: "https://brand-status-test.example",
      redirectUris: ["https://brand-status-test.example/callback"],
      jwksUri: "https://brand-status-test.example/jwks",
      revocationUri: "https://brand-status-test.example/revoke",
      recipientBaseUri: "https://cdr.brand-status-test.example",
      scopes: ["openid", "profile"],
      status: "ACTIVE",
    },
  );

  await page.goto(`/data-recipients/${legalEntityId}`);

  await showAllBrands(page);

  const actions = (name: string) =>
    page.getByRole("group", { name: `Status actions of ${name}`, exact: true });
  await expect(actions(brandName).getByRole("button")).toHaveText(["Deactivate", "Remove"]);

  await actions(brandName).getByRole("button", { name: "Deactivate" }).click();
  const deactivation = page.getByRole("alertdialog", { name: `Deactivate ${brandName}?` });
  await expect(deactivation.getByRole("listitem")).toHaveText([
    `${brandName} changes from ACTIVE to INACTIVE.`,
    `${productName} changes from ACTIVE to INACTIVE.`,
  ]);
  await expectNoAxeViolations(page);
  await deactivation.getByRole("button", { name: "Deactivate", exact: true }).click();

  await expect(deactivation).toBeHidden();
  await expect(page.getByRole("dialog", { name: `${brandName} is now INACTIVE` })).toBeVisible();
  await expect(actions(brandName).getByRole("button")).toHaveText(["Activate", "Remove"]);
  const product = page.getByRole("region", { name: productName, exact: true });
  await expect(product.getByText("INACTIVE", { exact: true })).toBeVisible();
  await expect(actions(productName).getByRole("button")).toHaveText(["Remove"]);

  const recipients = await cdrGet(request, "data-recipients", "4", isDataRecipients);
  expect(recipients.data.find((d) => d.legalEntityId === legalEntityId)).toMatchObject({
    dataRecipientBrands: [
      {
        dataRecipientBrandId: brandId,
        status: "INACTIVE",
        softwareProducts: [{ softwareProductId: productId, status: "INACTIVE" }],
      },
    ],
  });
  const statuses = await cdrGet(
    request,
    "data-recipients/brands/software-products/status",
    "3",
    isStatusList,
  );
  expect(statuses.data.find((s) => s.softwareProductId === productId)?.status).toBe("INACTIVE");
  expect(errors).toEqual([]);
});
