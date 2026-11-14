import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page } from "vitest/browser";

import { productCategoryValues } from "../api/schema.gen";
import {
  http,
  narrowWidth,
  presets,
  product,
  renderApp,
  serveProducts,
  showPreset,
  worker,
} from "../test/app";

const terms = () =>
  [...document.querySelectorAll("main dl > div")].map((row) => row.textContent ?? "");

test("the detail shows each field of a deposit product in plain language", async () => {
  serveProducts([
    product("saver", {
      name: "Bonus Saver",
      description: "A savings account with a bonus rate.",
      effectiveFrom: "2026-01-01T00:00:00+08:00",
      depositRates: [
        {
          depositRateType: "VARIABLE",
          rate: "0.0125",
          applicationType: "PERIODIC",
          applicationFrequency: "P1M",
        },
        {
          depositRateType: "BONUS",
          rate: "0.035",
          applicationType: "PERIODIC",
          applicationFrequency: "P1M",
          additionalValue: "Deposit $50 each month",
          additionalInfo: "Make no withdrawal in the month.",
        },
      ],
      fees: [
        {
          name: "Monthly fee",
          feeType: "PERIODIC",
          feeMethodUType: "fixedAmount",
          fixedAmount: { amount: "5.00" },
          additionalValue: "P1M",
        },
        {
          name: "Overseas fee",
          feeType: "TRANSACTION",
          feeMethodUType: "rateBased",
          rateBased: { rateType: "TRANSACTION", rate: "0.03" },
        },
      ],
      features: [
        { featureType: "NPP_PAYID" },
        { featureType: "DIGITAL_WALLET", additionalValue: "Apple Pay" },
        { featureType: "OTHER", additionalInfo: "Round up your purchases." },
      ],
    }),
  ]);

  renderApp("/products/saver");

  await expect.element(page.getByRole("heading", { level: 1, name: "Bonus Saver" })).toBeVisible();
  await expect.poll(() => document.title).toBe("Bonus Saver - Example Bank online banking");
  await expect.element(page.getByText("Everyday and savings accounts")).toBeVisible();
  await expect.element(page.getByText("A savings account with a bonus rate.")).toBeVisible();
  const rates = page.getByRole("region", { name: "Interest rates" });
  await expect
    .element(rates.getByRole("heading", { level: 3, name: "Variable rate" }))
    .toBeVisible();
  await expect.element(rates.getByRole("heading", { level: 3, name: "Bonus rate" })).toBeVisible();
  const fees = page.getByRole("region", { name: "Fees" });
  await expect.element(fees.getByRole("heading", { level: 3, name: "Monthly fee" })).toBeVisible();
  const features = page.getByRole("region", { name: "Features" });
  await expect
    .element(features.getByRole("heading", { level: 3, name: "Receive payments to a PayID" }))
    .toBeVisible();
  await expect
    .element(features.getByRole("heading", { level: 3, name: "Other feature" }))
    .toBeVisible();

  expect(terms()).toEqual([
    "Available from:1 January 2026",
    "Rate:1.25%",
    "Interest is paid:Regularly",
    "How often:Each month",
    "Rate:3.50%",
    "Interest is paid:Regularly",
    "How often:Each month",
    "Details:Deposit $50 each month",
    "More information:Make no withdrawal in the month.",
    "Type:Regular fee",
    "Worked out as:A fixed amount",
    "Amount:$5.00",
    "Charged:Each month",
    "Type:Transaction fee",
    "Worked out as:A percentage",
    "Rate:3.00% of the transaction amount",
    "Details:Apple Pay",
    "More information:Round up your purchases.",
  ]);
  expect(page.getByText("The rates and fees can be different for you.").query()).toBeNull();
  await expect
    .element(page.getByRole("link", { name: "Open an account" }))
    .toHaveAttribute("href", "/products/saver/open");
});

test("the detail shows each field of a lending product, and the note of a tailored product", async () => {
  serveProducts([
    product("home", {
      name: "Home Loan",
      productCategory: "RESIDENTIAL_MORTGAGES",
      isTailored: true,
      effectiveTo: "2027-06-30T00:00:00+08:00",
      lendingRates: [
        {
          lendingRateType: "FIXED",
          rate: "0.0599",
          comparisonRate: "0.0612",
          applicationType: "PERIODIC",
          applicationFrequency: "P1M",
          repaymentType: "PRINCIPAL_AND_INTEREST",
          loanPurpose: "OWNER_OCCUPIED",
          additionalValue: "P2Y",
        },
      ],
      fees: [
        {
          name: "Break fee",
          feeType: "EXIT",
          feeMethodUType: "variable",
          variable: { feeMinimum: "100.00", feeMaximum: "5000.00" },
        },
      ],
    }),
  ]);

  renderApp("/products/home");

  await expect
    .element(page.getByText("The rates and fees can be different for you."))
    .toBeVisible();
  await expect.element(page.getByRole("heading", { level: 3, name: "Fixed rate" })).toBeVisible();
  expect(terms()).toEqual([
    "Available until:30 June 2027",
    "Rate:5.99%",
    "Comparison rate:6.12%",
    "Interest is charged:Regularly",
    "How often:Each month",
    "Repayments:Principal and interest",
    "Loan purpose:A home you live in",
    "Period:2 years",
    "Type:Account closing fee",
    "Worked out as:An amount in a range",
    "Minimum:$100.00",
    "Maximum:$5,000.00",
  ]);
  await expect.element(page.getByRole("link", { name: "Open an account" })).toBeVisible();
});

test.each(productCategoryValues)(
  "a product of %s has the action to open an account",
  async (productCategory) => {
    serveProducts([product("one", { name: "One", productCategory })]);

    renderApp("/products/one");

    await expect
      .element(page.getByRole("link", { name: "Open an account" }))
      .toHaveAttribute("href", "/products/one/open");
  },
);

test("a product that is not current shows the not found page", async () => {
  serveProducts([]);

  renderApp("/products/ended-saver");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Page not found" }))
    .toBeVisible();
  await expect.poll(() => document.title).toBe("Page not found - Example Bank online banking");
});

test("the detail shows the problem when the product does not load", async () => {
  worker.use(
    http.get("/api/products/{productId}", ({ response }) =>
      response("default").json(
        {
          type: "about:blank",
          title: "Internal Server Error",
          status: 500,
          detail: "The data holder could not complete the request.",
        },
        { status: 500 },
      ),
    ),
  );

  renderApp("/products/saver");

  await expect
    .element(page.getByRole("alert"))
    .toHaveTextContent("Internal Server ErrorThe data holder could not complete the request.");
});

describe.each(presets)("with the preset $name", (preset) => {
  test.each(preset.products.map(({ productId, name }) => ({ productId, name })))(
    "the detail of $name passes the accessibility checks and has no horizontal scroll at 320 CSS px",
    async ({ productId, name }) => {
      await showPreset(preset, `/products/${productId}`);

      await expect.element(page.getByRole("heading", { level: 1, name })).toBeVisible();
      expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
      await expectNoAxeViolations(document.documentElement);
    },
  );
});
