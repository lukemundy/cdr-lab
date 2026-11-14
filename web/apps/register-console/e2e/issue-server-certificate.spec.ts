import { expect, test } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import { firstCertificateOf, makeCSR, publicKeyOf } from "./csr";
import { showAllBrands } from "./lab";

const site = "https://server-issue-test.example";

// A brand with a dataHolderBrandId has secure endpoints. A brand with only an interim ID shares only product data.
function brand(name: string, ids: { dataHolderBrandId: string } | { interimId: string }) {
  const secure =
    "dataHolderBrandId" in ids
      ? {
          resourceBaseUri: "https://mtls.server-issue-test.example",
          infosecBaseUri: "https://id.server-issue-test.example",
          websiteUri: site,
        }
      : {};
  return {
    ...ids,
    brandName: name,
    industries: ["banking"],
    logoUri: `${site}/brand.png`,
    status: "ACTIVE",
    endpointDetail: {
      version: "1",
      publicBaseUri: "https://api.server-issue-test.example",
      ...secure,
    },
    ...("dataHolderBrandId" in ids
      ? { jwksEndpoint: "https://id.server-issue-test.example/jwks" }
      : {}),
  };
}

// Each run makes its own data holder and its own CSR key, so the test runs in a lab with no conformance data, and
// again in the same lab.
test("the operator issues a server certificate for a data holder brand, saves the chain, and the history shows the issue", async ({
  page,
  request,
}, testInfo) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  const legalEntityId = randomUUID();
  const brandId = randomUUID();
  const run = legalEntityId.slice(0, 8);
  const holderName = `Test Server Issue ${run} Ltd`;
  const brandName = `Server Issue Test ${run}`;
  const productBrandName = `Server Issue Products ${run}`;

  const created = await request.post("/api/data-holders", {
    data: {
      legalEntityId,
      legalEntityName: holderName,
      logoUri: `${site}/logo.png`,
      status: "ACTIVE",
    },
  });
  expect(created.status(), await created.text()).toBe(201);
  for (const [version, data] of [
    [1, brand(brandName, { dataHolderBrandId: brandId })],
    [2, brand(productBrandName, { interimId: randomUUID() })],
  ] as const) {
    // oxlint-disable-next-line no-await-in-loop -- each add needs the version that the add before it made.
    const added = await request.post(`/api/data-holders/${legalEntityId}/brands`, {
      data: { version, ...data },
    });
    // oxlint-disable-next-line no-await-in-loop -- the same reason.
    expect(added.status(), await added.text()).toBe(201);
  }
  const host = `mtls.${run}.server-issue-test.example`;
  const csr = makeCSR(testInfo.outputPath("server.csr"), { cn: host, o: brandName }, [
    host,
    "localhost",
  ]);

  await page.goto(`/data-holders/${legalEntityId}`);
  await showAllBrands(page);

  await expect(
    page
      .getByRole("region", { name: productBrandName, exact: true })
      .getByRole("button", { name: "Issue certificate" }),
  ).toHaveCount(0);
  await page
    .getByRole("region", { name: brandName, exact: true })
    .getByRole("button", { name: "Issue certificate" })
    .click();
  const dialog = page.getByRole("dialog", { name: `Issue a server certificate for ${brandName}` });
  await dialog.getByLabel("CSR file").setInputFiles(csr.path);
  await expect(dialog.getByRole("textbox", { name: "PEM CSR" })).toHaveValue(csr.text);
  await expectNoAxeViolations(page);
  await dialog.getByRole("button", { name: "Issue", exact: true }).click();

  await expect(
    dialog.getByRole("heading", { name: `Server certificate issued for ${brandName}` }),
  ).toBeFocused();
  const chain = await dialog.getByRole("textbox", { name: "PEM chain" }).inputValue();
  const issued = firstCertificateOf(chain);
  expect(issued.publicKey).toBe(publicKeyOf(csr));
  expect(issued.issuer).toContain("CN=Mock CDR Intermediate CA");
  await expect(dialog.getByText(issued.thumbprint, { exact: true })).toBeVisible();
  await expectNoAxeViolations(page);

  const download = page.waitForEvent("download");
  await dialog.getByRole("button", { name: "Save as file" }).click();
  const saved = await download;
  expect(saved.suggestedFilename()).toBe(`server-certificate-${brandId}.pem`);
  expect(readFileSync(await saved.path(), "utf8")).toBe(chain);

  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();

  await page.getByRole("link", { name: "History", exact: true }).click();
  const history = page.getByRole("list", { name: "History" });
  await expect(
    history.getByRole("listitem").filter({ hasText: "ServerCertificateIssued" }).first(),
  ).toContainText(issued.thumbprint);
  await expect(history.getByText("Server certificate issued").first()).toBeVisible();
  await expectNoAxeViolations(page);
  expect(errors).toEqual([]);
});
