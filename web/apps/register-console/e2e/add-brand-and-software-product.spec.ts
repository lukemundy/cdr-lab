import { expect, test, type APIRequestContext, type Page } from "@playwright/test";
import { createHash, X509Certificate } from "node:crypto";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { expectNoAxeViolations } from "@cdr-lab/app/e2e/checks";
import { publicURL } from "./lab";

// The same certificate as the client certificate test. The Register permits it for more than one software product.
const pemText = readFileSync(
  fileURLToPath(new URL("fixtures/client-certificate.pem", import.meta.url)),
  "utf8",
);
const thumbprint = createHash("sha256")
  .update(new X509Certificate(pemText).raw)
  .digest("base64url");
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

type DataRecipients = {
  data: {
    legalEntityId: string;
    dataRecipientBrands: {
      dataRecipientBrandId: string;
      softwareProducts: { softwareProductId: string }[];
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

async function newIdentifier(page: Page, label: string): Promise<string> {
  const field = page.getByRole("textbox", { name: label });
  await expect(field).toHaveValue(uuid);
  await expect(page.getByRole("combobox", { name: "Status" })).toHaveValue("ACTIVE");
  return field.inputValue();
}

async function fill(page: Page, values: Record<string, string>) {
  for (const [name, value] of Object.entries(values)) {
    // oxlint-disable-next-line no-await-in-loop -- a fill types into the field that has the focus, so the fills cannot run at the same time.
    await page.getByRole("textbox", { name, exact: true }).fill(value);
  }
}

// Each run makes its own data recipient with the UUIDs of the forms, so the test passes again on the same lab. The name
// sorts after the seed data recipients that the other tests find on the first page of the list.
test("the operator creates a data recipient, adds a brand and a software product, and registers a certificate", async ({
  page,
  request,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));

  await page.goto("/create/data-recipient");
  const legalEntityId = await newIdentifier(page, "Legal entity ID");
  const run = legalEntityId.slice(0, 8);
  const recipientName = `Operator Add Test ${run} Pty Ltd`;
  await fill(page, {
    Name: recipientName,
    "Accreditation number": `ADR-ADD-${run}`,
    "Logo URI": "https://add-test.example/logo.png",
  });
  await page.getByRole("combobox", { name: "Accreditation level" }).selectOption("UNRESTRICTED");
  await page.getByRole("button", { name: "Create data recipient" }).click();
  await expect(page.getByRole("heading", { level: 1, name: recipientName })).toBeFocused();

  await page.getByRole("link", { name: "Add brand" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    `Add a brand to ${recipientName}`,
  );
  const brandId = await newIdentifier(page, "Brand ID");
  await expectNoAxeViolations(page);
  const brandName = `Add Test ${run}`;
  await fill(page, { Name: brandName, "Logo URI": "https://add-test.example/brand.png" });
  await page.getByRole("button", { name: "Add brand" }).click();

  await expect(page.getByRole("heading", { level: 1, name: recipientName })).toBeFocused();
  await expect(page.getByRole("dialog", { name: `Brand ${brandName} added` })).toBeVisible();
  const brand = page.getByRole("region", { name: brandName, exact: true });
  await expect(brand.getByText(brandId, { exact: true })).toBeVisible();

  await brand.getByRole("link", { name: "Add software product" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    `Add a software product to ${brandName}`,
  );
  const productId = await newIdentifier(page, "Software product ID");
  await expectNoAxeViolations(page);
  const productName = `Add Test ${run} Web`;
  await fill(page, {
    Name: productName,
    Description: "A software product that the browser test adds",
    "Logo URI": "https://add-test.example/app.png",
    "Client URI": "https://add-test.example",
    "Redirect URIs": "https://add-test.example/callback",
    "JWKS URI": "https://add-test.example/jwks",
    "Revocation URI": "https://add-test.example/revoke",
    "Recipient base URI": "https://cdr.add-test.example",
    Scopes: "openid\nprofile",
  });
  await page.getByRole("button", { name: "Add software product" }).click();

  await expect(page.getByRole("heading", { level: 1, name: recipientName })).toBeFocused();
  await expect(
    page.getByRole("dialog", { name: `Software product ${productName} added` }),
  ).toBeVisible();
  const product = page.getByRole("region", { name: productName, exact: true });
  await expect(product.getByText(productId, { exact: true })).toBeVisible();
  await expectNoAxeViolations(page);

  await product.getByRole("button", { name: "Register certificate" }).click();
  const registration = page.getByRole("dialog", {
    name: `Register a client certificate for ${productName}`,
  });
  await registration.getByRole("textbox", { name: "PEM certificate" }).fill(pemText);
  await registration.getByRole("button", { name: "Register", exact: true }).click();
  await expect(registration).toBeHidden();
  await expect(product.getByText(thumbprint, { exact: true })).toBeVisible();

  const recipients = await cdrGet(request, "data-recipients", "4", isDataRecipients);
  expect(recipients.data.find((d) => d.legalEntityId === legalEntityId)).toMatchObject({
    dataRecipientBrands: [
      { dataRecipientBrandId: brandId, softwareProducts: [{ softwareProductId: productId }] },
    ],
  });
  const statuses = await cdrGet(
    request,
    "data-recipients/brands/software-products/status",
    "3",
    isStatusList,
  );
  expect(statuses.data.find((s) => s.softwareProductId === productId)?.status).toBe("ACTIVE");
  expect(errors).toEqual([]);
});
