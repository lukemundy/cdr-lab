import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";

import { categoryLabels } from "../src/products/categories";
import { formatRate } from "../src/products/format";

type CDSProduct = {
  productId: string;
  name: string;
  description: string;
  productCategory: keyof typeof categoryLabels;
  depositRates?: { rate: string }[];
  lendingRates?: { rate: string }[];
};

function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`Set ${name}: mise run test:lab-web sets it.`);
  return value;
}

// Node trusts the lab public CA, so these calls verify the TLS certificate of the public listener.
async function cds<T>(path: string, version: string): Promise<T> {
  const response = await fetch(
    new URL(`/cds-au/v1/banking${path}`, requiredEnv("DATAHOLDER_PUBLIC_URL")),
    {
      headers: { "x-v": version },
    },
  );
  if (!response.ok) throw new Error(`${path} gave ${response.status}: ${await response.text()}`);
  // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the CDS OpenAPI document gives the shape, and the test fails when the page does not show it.
  return ((await response.json()) as { data: T }).data;
}

test("a consumer goes from the landing page to a category and to the detail of a product, with the data of the CDS APIs", async ({
  page,
}) => {
  const { products } = await cds<{ products: CDSProduct[] }>(
    "/products?effective=CURRENT&page-size=1000",
    "5",
  );
  const withRates =
    products.find((p) => (p.depositRates ?? p.lendingRates ?? []).length > 0) ?? products[0];
  if (!withRates) throw new Error("The bank has no current product.");
  const detail = await cds<CDSProduct>(`/products/${withRates.productId}`, "7");

  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1 })).toContainText("Banking with");
  await expectNoAxeViolations(page);

  const category = categoryLabels[detail.productCategory];
  await page
    .getByRole("main")
    .getByRole("link", { name: new RegExp(`^${category}`) })
    .click();

  await expect(page).toHaveURL(/\/products#/);
  const names = await page.getByRole("main").getByRole("heading", { level: 3 }).allTextContents();
  expect(names.toSorted()).toEqual(products.map((p) => p.name).toSorted());
  await expectNoAxeViolations(page);

  await page
    .getByRole("region", { name: category })
    .getByRole("link", { name: detail.name, exact: true })
    .click();

  await expect(page.getByRole("heading", { level: 1, name: detail.name })).toBeVisible();
  await expect(page.getByText(detail.description)).toBeVisible();
  const rates = [...(detail.depositRates ?? []), ...(detail.lendingRates ?? [])];
  await Promise.all(
    rates.map(({ rate }) =>
      expect(page.getByText(formatRate(rate), { exact: true }).first()).toBeVisible(),
    ),
  );
  await expect(page).toHaveTitle(new RegExp(`^${detail.name} - .+ online banking$`));
  await expectNoAxeViolations(page);
});

test("a product that is not current shows the not found page", async ({ page }) => {
  const { products } = await cds<{ products: CDSProduct[] }>(
    "/products?effective=FUTURE&page-size=1000",
    "5",
  );
  const future = products[0];
  test.skip(future === undefined, "The bank has no future product.");

  await page.goto(`/products/${future?.productId}`);

  await expect(page.getByRole("heading", { level: 1, name: "Page not found" })).toBeVisible();
});
