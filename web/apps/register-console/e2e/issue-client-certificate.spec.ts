import { expect, test, type APIRequestContext } from "@playwright/test";
import { randomUUID, X509Certificate } from "node:crypto";
import { readFileSync } from "node:fs";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import { firstCertificateOf, makeCSR, publicKeyOf } from "./csr";
import { showAllBrands } from "./lab";

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

// Each run makes its own data recipient and its own CSR key, so the test runs in a lab with no conformance data,
// and again in the same lab.
test("the operator issues a client certificate from a CSR file, saves the chain, and revokes the certificate from the saved file", async ({
  page,
  request,
}, testInfo) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  const legalEntityId = randomUUID();
  const brandId = randomUUID();
  const productId = randomUUID();
  const run = legalEntityId.slice(0, 8);
  const brandName = `Client Issue Test ${run}`;
  const productName = `Client Issue Test ${run} Web`;

  await expectCreated(request, "/api/data-recipients", {
    legalEntityId,
    legalEntityName: `Operator Client Issue Test ${run} Pty Ltd`,
    accreditationNumber: `ADR-ISSUE-${run}`,
    accreditationLevel: "UNRESTRICTED",
    logoUri: "https://client-issue-test.example/logo.png",
    status: "ACTIVE",
  });
  await expectCreated(request, `/api/data-recipients/${legalEntityId}/brands`, {
    version: await versionOf(request, legalEntityId),
    dataRecipientBrandId: brandId,
    brandName,
    logoUri: "https://client-issue-test.example/brand.png",
    status: "ACTIVE",
  });
  await expectCreated(
    request,
    `/api/data-recipients/${legalEntityId}/brands/${brandId}/software-products`,
    {
      version: await versionOf(request, legalEntityId),
      softwareProductId: productId,
      softwareProductName: productName,
      softwareProductDescription: "A software product of the client certificate issue test",
      logoUri: "https://client-issue-test.example/app.png",
      clientUri: "https://client-issue-test.example",
      redirectUris: ["https://client-issue-test.example/callback"],
      jwksUri: "https://client-issue-test.example/jwks",
      revocationUri: "https://client-issue-test.example/revoke",
      recipientBaseUri: "https://cdr.client-issue-test.example",
      scopes: ["openid", "profile"],
      status: "ACTIVE",
    },
  );
  const wrongName = makeCSR(testInfo.outputPath("wrong-name.csr"), {
    cn: "Other Product",
    o: brandName,
  });
  const csr = makeCSR(testInfo.outputPath("client.csr"), { cn: productName, o: brandName });

  await page.goto(`/data-recipients/${legalEntityId}`);
  await showAllBrands(page);
  const section = page.getByRole("region", { name: productName, exact: true });

  await section.getByRole("button", { name: "Issue certificate" }).click();
  const dialog = page.getByRole("dialog", {
    name: `Issue a client certificate for ${productName}`,
  });
  const csrField = dialog.getByRole("textbox", { name: "PEM CSR" });
  await csrField.fill(wrongName.text);
  await dialog.getByRole("button", { name: "Issue", exact: true }).click();

  await expect(csrField).toHaveAttribute("aria-invalid", "true");
  await expect(csrField).toHaveAccessibleDescription(
    new RegExp(`subject CN must be the software product name "${productName}"`),
  );

  await dialog.getByLabel("CSR file").setInputFiles(csr.path);
  await expect(csrField).toHaveValue(csr.text);
  await expectNoAxeViolations(page);
  await dialog.getByRole("button", { name: "Issue", exact: true }).click();

  await expect(
    dialog.getByRole("heading", { name: `Client certificate issued for ${productName}` }),
  ).toBeFocused();
  await expect(
    dialog.getByText(
      "The Register does not keep the certificate, so the console cannot show it again.",
    ),
  ).toBeVisible();
  const chain = await dialog.getByRole("textbox", { name: "PEM chain" }).inputValue();
  const issued = firstCertificateOf(chain);
  expect(issued.publicKey).toBe(publicKeyOf(csr));
  expect(issued.issuer).toContain("CN=Mock CDR Intermediate CA");
  await expect(dialog.getByText(issued.thumbprint, { exact: true })).toBeVisible();
  await expectNoAxeViolations(page);

  const download = page.waitForEvent("download");
  await dialog.getByRole("button", { name: "Save as file" }).click();
  const saved = await download;
  expect(saved.suggestedFilename()).toBe(`client-certificate-${productId}.pem`);
  expect(readFileSync(await saved.path(), "utf8")).toBe(chain);

  // The footer button comes before the close button of the dialog title.
  await dialog.getByRole("button", { name: "Close", exact: true }).first().click();

  await expect(dialog).toBeHidden();
  await expect(section.getByText(issued.thumbprint, { exact: true })).toBeVisible();

  const savedPath = testInfo.outputPath("client-certificate.pem");
  await saved.saveAs(savedPath);
  // Node gives the serial number in the format of openssl x509 -serial, which the console also uses.
  const { serialNumber } = new X509Certificate(chain);

  await page
    .getByRole("navigation", { name: "Main" })
    .getByRole("link", { name: "CDR certificates" })
    .click();
  const revoke = async () => {
    await page.getByRole("button", { name: "Revoke certificate" }).click();
    const revocation = page.getByRole("dialog", { name: "Revoke a CDR certificate" });
    await revocation.getByLabel("Certificate file").setInputFiles(savedPath);
    await expect(revocation.getByRole("textbox", { name: "PEM certificate" })).toHaveValue(chain);
    await revocation
      .getByRole("combobox", { name: "Reason" })
      .selectOption({ label: "Cessation of operation" });
    await revocation.getByRole("button", { name: "Continue" }).click();
    return revocation;
  };

  const revocation = await revoke();
  const confirmation = page.getByRole("alertdialog", { name: "Revoke this CDR certificate?" });
  await expect(confirmation).toContainText(`CN=${productName}`);
  await expect(confirmation).toContainText(serialNumber);
  await expect(confirmation).toContainText(issued.thumbprint);
  await expect(confirmation).toContainText('"Cessation of operation"');
  await expect(confirmation).toContainText("You cannot reverse this action.");
  await expectNoAxeViolations(page);
  await confirmation.getByRole("button", { name: "Revoke", exact: true }).click();

  await expect(confirmation).toBeHidden();
  await expect(revocation).toBeHidden();
  await expect(
    page.getByRole("dialog", { name: `Certificate ${serialNumber} revoked` }),
  ).toBeVisible();
  const newest = page.getByRole("table", { name: "Revocations" }).getByRole("row").nth(1);
  await expect(newest).toContainText(serialNumber);
  await expect(newest).toContainText(issued.thumbprint);
  await expect(newest).toContainText("Cessation of operation");
  await expectNoAxeViolations(page);

  // The Register keeps the revocation, so a second revocation of the same file changes nothing.
  const second = await revoke();
  await expect(second.getByRole("alert")).toContainText(
    "The certificate is already revoked, with the reason cessationOfOperation.",
  );
  await expect(confirmation).toBeHidden();
  expect(errors).toEqual([]);
});
