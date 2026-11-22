import { expect, test } from "@playwright/test";
import { X509Certificate } from "node:crypto";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { expectNoAxeViolations } from "@cdr-lab/app/e2e/checks";

// A self-signed certificate with no private key in the repository.
const selfSignedFile = fileURLToPath(new URL("fixtures/client-certificate.pem", import.meta.url));

function certificatesOf(bundle: string): X509Certificate[] {
  return (
    bundle.match(/-----BEGIN CERTIFICATE-----[\s\S]+?-----END CERTIFICATE-----\n?/g) ?? []
  ).map((pem) => new X509Certificate(pem));
}

test("the operator saves the CDR CA bundle, the intermediate and then the root", async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  await page.goto("/cdr-certificates");

  const field = page.getByRole("textbox", { name: "PEM CA bundle" });
  await expect(field).toHaveValue(/BEGIN CERTIFICATE/);
  const bundle = await field.inputValue();
  const [intermediate, root] = certificatesOf(bundle);
  expect(certificatesOf(bundle)).toHaveLength(2);
  expect(intermediate?.subject).toContain("CN=Mock CDR Intermediate CA");
  expect(root && intermediate?.verify(root.publicKey)).toBe(true);
  expect(root && root.verify(root.publicKey)).toBe(true);
  await expectNoAxeViolations(page);

  const download = page.waitForEvent("download");
  await page.getByRole("button", { name: "Save as file" }).click();
  const saved = await download;
  expect(saved.suggestedFilename()).toBe("cdr-ca-bundle.pem");
  expect(readFileSync(await saved.path(), "utf8")).toBe(bundle);
  expect(errors).toEqual([]);
});

test("the Register refuses to revoke a certificate that the intermediate did not sign", async ({
  page,
}) => {
  await page.goto("/cdr-certificates");

  await page.getByRole("button", { name: "Revoke certificate" }).click();
  const revocation = page.getByRole("dialog", { name: "Revoke a CDR certificate" });
  await revocation.getByLabel("Certificate file").setInputFiles(selfSignedFile);
  await revocation
    .getByRole("combobox", { name: "Reason" })
    .selectOption({ label: "Key compromise" });
  await revocation.getByRole("button", { name: "Continue" }).click();

  const field = revocation.getByRole("textbox", { name: "PEM certificate" });
  await expect(field).toHaveAttribute("aria-invalid", "true");
  await expect(field).toHaveAccessibleDescription(
    /The mock CDR intermediate did not sign the certificate\./,
  );
  await expect(revocation.getByRole("alert")).toContainText("There is a problem");
  await expect(page.getByRole("alertdialog")).toHaveCount(0);
  await expectNoAxeViolations(page);
});
