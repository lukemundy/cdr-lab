import { expect, test } from "@playwright/test";
import { createHash, X509Certificate } from "node:crypto";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";
import { showAllBrands } from "./lab";

// Only this test changes this record of services/register/seed/example.json, and it removes the certificate again.
const recipient = {
  name: "Console Certificate Test Pty Ltd",
  id: "816bfdc3-72b2-4a5c-8cf1-6bc4541a9777",
};
const product = "Console Certificate Test Web";

// A self-signed certificate with no private key in the repository. The Register does not check the issuer.
const certificateFile = fileURLToPath(new URL("fixtures/client-certificate.pem", import.meta.url));
const pemText = readFileSync(certificateFile, "utf8");
const thumbprint = createHash("sha256")
  .update(new X509Certificate(pemText).raw)
  .digest("base64url");

test("the operator registers a client certificate from a file, and then removes it through the confirmation", async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  await page.goto(`/data-recipients/${recipient.id}`);
  await showAllBrands(page);
  const section = page.getByRole("region", { name: product, exact: true });
  const remove = section.getByRole("button", {
    name: `Remove the client certificate ${thumbprint}`,
  });

  await section.getByRole("button", { name: "Register certificate" }).click();
  const registration = page.getByRole("dialog", {
    name: `Register a client certificate for ${product}`,
  });
  await registration.getByLabel("Certificate file").setInputFiles(certificateFile);
  await expect(registration.getByRole("textbox", { name: "PEM certificate" })).toHaveValue(pemText);
  await expectNoAxeViolations(page);
  await registration.getByRole("button", { name: "Register", exact: true }).click();

  await expect(registration).toBeHidden();
  await expect(
    page.getByRole("dialog", { name: `Client certificate registered for ${product}` }),
  ).toBeVisible();
  await expect(section.getByText(thumbprint, { exact: true })).toBeVisible();

  await remove.click();
  const removal = page.getByRole("alertdialog", {
    name: `Remove the client certificate of ${product}?`,
  });
  await expect(removal).toContainText(thumbprint);
  await expect(removal).toContainText(product);
  await expectNoAxeViolations(page);
  await removal.getByRole("button", { name: "Remove", exact: true }).click();

  await expect(removal).toBeHidden();
  await expect(section.getByText(thumbprint, { exact: true })).toBeHidden();

  await page.getByRole("link", { name: "History", exact: true }).click();
  const history = page.getByRole("list", { name: "History" });
  await expect(
    history.getByRole("listitem").filter({ hasText: "ClientCertificateRemoved" }).first(),
  ).toContainText(thumbprint);
  await expect(history.getByText("Client certificate removed").first()).toBeVisible();
  await expectNoAxeViolations(page);
  expect(errors).toEqual([]);
});
