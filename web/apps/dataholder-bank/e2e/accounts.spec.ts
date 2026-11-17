import { readdir, readFile } from "node:fs/promises";
import { randomInt, randomUUID } from "node:crypto";

import { expect, test, type Page } from "@playwright/test";

import { definitionOf, expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import { requiredEnv } from "@cdr-lab/console/e2e/dataholder-lab";

import { loanProductCategoryValues, revolvingProductCategoryValues } from "../src/api/schema.gen";

type CDSProduct = { productId: string; name: string; productCategory: string };

const brandsDir = new URL("../../../../services/dataholder/brands/", import.meta.url);

// Node trusts the lab public CA, so the call verifies the TLS certificate of the public listener.
async function currentProducts(): Promise<CDSProduct[]> {
  const response = await fetch(
    new URL(
      "/cds-au/v1/banking/products?effective=CURRENT&page-size=1000",
      requiredEnv("DATAHOLDER_PUBLIC_URL"),
    ),
    { headers: { "x-v": "5" } },
  );
  if (!response.ok) throw new Error(`Get Products gave ${response.status}`);
  // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the CDS OpenAPI document gives the shape, and the test fails when the page does not show it.
  return ((await response.json()) as { data: { products: CDSProduct[] } }).data.products;
}

// The bsb of the preset brand directory whose brand name the site shows.
async function bsbOfBrand(brandName: string): Promise<string> {
  const presets = await readdir(brandsDir, { withFileTypes: true });
  const brands = await Promise.all(
    presets
      .filter((entry) => entry.isDirectory())
      .map(async (entry): Promise<unknown> =>
        JSON.parse(await readFile(new URL(`${entry.name}/brand.json`, brandsDir), "utf8")),
      ),
  );
  const brand = brands.find(
    (b) => typeof b === "object" && b !== null && Reflect.get(b, "brandName") === brandName,
  );
  if (typeof brand !== "object" || brand === null)
    throw new Error(`No preset brand has the name ${brandName}.`);
  return String(Reflect.get(brand, "bsb"));
}

// The site gives no BSB, so the test reads the account with the operator interface.
async function accountsOfCustomer(
  customerNumber: string,
): Promise<{ bsb: string; accountNumber: string }[]> {
  const response = await fetch(
    new URL(`/v1/customers/${customerNumber}`, requiredEnv("DATAHOLDER_OPERATOR_URL")),
    { headers: { Authorization: `Bearer ${requiredEnv("DATAHOLDER_OPERATOR_TOKEN")}` } },
  );
  if (!response.ok) throw new Error(`The operator interface gave ${response.status}`);
  // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the operator interface gives the fields of an account.
  const body = (await response.json()) as { accounts: { bsb: string; accountNumber: string }[] };
  return body.accounts;
}

async function signUp(page: Page): Promise<string> {
  await page.goto("/sign-up");
  await page.getByLabel("Last name").fill(`Quoll ${randomInt(1_000_000)}`);
  await page.getByLabel("Email address").fill(`quoll.${randomUUID()}@example.test`);
  await page.getByRole("button", { name: "Sign up" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Your customer number" })).toBeVisible();
  const customerNumber = await page.getByText(/^\d{8}$/).textContent();
  if (!customerNumber) throw new Error("The page shows no customer number.");
  return customerNumber;
}

const isIn = (categories: readonly string[]) => (product: CDSProduct) =>
  categories.includes(product.productCategory);

// The journey makes its own customer, because the lab has no seed customers, and gives no consent.
// Verifies: DH-2215, DH-2240
test("a new customer opens an account of a deposit product, and the customer closes their last open account after the confirmation", async ({
  page,
}) => {
  const products = await currentProducts();
  const deposit = products.find((p) => p.productCategory === "TRANS_AND_SAVINGS_ACCOUNTS");
  if (!deposit) throw new Error("The bank has no current deposit product.");

  const customerNumber = await signUp(page);

  await page.goto(`/products/${deposit.productId}`);
  await page.getByRole("link", { name: "Open an account" }).click();
  await expect(page.getByRole("heading", { level: 1, name: `Open ${deposit.name}` })).toBeVisible();
  await expectNoAxeViolations(page);

  await page.getByRole("button", { name: "Open account" }).click();
  await expect(page.getByRole("heading", { level: 1, name: deposit.name })).toBeFocused();
  await expect(page.getByRole("status").filter({ hasText: "account is open" })).toHaveText(
    `Your ${deposit.name} account is open.`,
  );
  await expect(definitionOf(page, "Product")).toHaveText(deposit.name);
  await expectNoAxeViolations(page);

  const brand = await fetch(new URL("/api/brand", requiredEnv("BANK_SITE_URL")));
  // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the site API document gives the shape.
  const { brandName } = (await brand.json()) as { brandName: string };
  const accounts = await accountsOfCustomer(customerNumber);
  expect(accounts.map((a) => a.bsb)).toEqual([await bsbOfBrand(brandName)]);
  const lastDigits = accounts[0]?.accountNumber.slice(-4);

  const accountUrl = page.url();
  await page.getByRole("button", { name: "Close account" }).click();

  const dialog = page.getByRole("alertdialog");
  await expect(dialog).toContainText(
    `${deposit.name}, account number xxx-xxx xxxx${lastDigits}, has a balance of $0.00.`,
  );
  await expect(dialog).toContainText(
    "This is your last open account. After the close, you cannot share your data under the Consumer Data Right.",
  );
  await expectNoAxeViolations(page);
  await dialog.getByRole("button", { name: "Close account" }).click();

  await expect(page.getByRole("heading", { level: 1, name: "Your accounts" })).toBeFocused();
  await expect(page.getByText(`You closed ${deposit.name}.`)).toBeVisible();
  await expect(
    page.getByRole("region", { name: "Closed accounts" }).getByRole("link", { name: deposit.name }),
  ).toBeVisible();
  await expect(
    page.getByText(
      "You cannot share your data under the Consumer Data Right until you have an open account.",
    ),
  ).toBeVisible();
  await expectNoAxeViolations(page);

  await page.goto(accountUrl);
  await expect(definitionOf(page, "Status")).toHaveText("Closed");
  await expect(page.getByRole("button", { name: "Close account" })).toHaveCount(0);
});

// Verifies: DH-2225, DH-2229
test("a new customer opens revolving credit with a credit limit and a loan that owes the loan amount", async ({
  page,
}) => {
  const products = await currentProducts();
  const revolving = products.find(isIn(revolvingProductCategoryValues));
  const loan = products.find(isIn(loanProductCategoryValues));
  if (!revolving || !loan)
    throw new Error("The bank has no current revolving credit or loan product.");

  await signUp(page);

  await page.goto(`/products/${revolving.productId}`);
  await page.getByRole("link", { name: "Open an account" }).click();
  await page.getByLabel("Credit limit").fill("$5,000");
  await expectNoAxeViolations(page);
  await page.getByRole("button", { name: "Open account" }).click();
  await expect(page.getByRole("heading", { level: 1, name: revolving.name })).toBeFocused();
  await expect(definitionOf(page, "Balance")).toHaveText("$0.00");
  await expect(definitionOf(page, "Available")).toHaveText("$5,000.00");
  await expect(definitionOf(page, "Credit limit")).toHaveText("$5,000.00");

  await page.goto(`/products/${loan.productId}`);
  await page.getByRole("link", { name: "Open an account" }).click();
  await page.getByLabel("Loan amount").fill("20,000.50");
  await page.getByRole("button", { name: "Open account" }).click();
  await expect(page.getByRole("heading", { level: 1, name: loan.name })).toBeFocused();
  await expect(definitionOf(page, "Balance")).toHaveText("-$20,000.50");
  await expect(definitionOf(page, "Available")).toHaveText("$0.00");
  await expect(definitionOf(page, "Credit limit")).toHaveText("$0.00");
  await expect(page.getByText("Loan drawdown", { exact: true })).toBeVisible();
  await expectNoAxeViolations(page);
});
