import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page } from "vitest/browser";

import {
  account,
  http,
  narrowWidth,
  presets,
  product,
  renderApp,
  serveAccounts,
  serveProducts,
  serveSession,
  sessionEnded,
  showPreset,
  worker,
} from "../test/app";

const saver = product("saver", { name: "Bonus Saver" });
const opened = account("account-new", {
  displayName: "Bonus Saver",
  productName: "Bonus Saver",
  balance: { currentBalance: "0.00", availableBalance: "0.00", creditLimit: "0.00" },
});

// serveOpen opens the account opened, and keeps the body of each request.
function serveOpen(bodies: unknown[]) {
  worker.use(
    http.post("/api/accounts", async ({ request, response }) => {
      bodies.push(await request.json());
      return response(201).json(opened);
    }),
  );
}

// Verifies: DH-2217
test("the open page names the product and its category, and needs a confirmation", async () => {
  serveSession();
  serveProducts([saver]);

  renderApp("/products/saver/open");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Open Bonus Saver" }))
    .toBeVisible();
  await expect.poll(() => document.title).toBe("Open Bonus Saver - Example Bank online banking");
  await expect.element(page.getByText("Everyday and savings accounts")).toBeVisible();
  await expect
    .element(page.getByText("Example Bank opens an account of this product in your name."))
    .toBeVisible();
  await expect.element(page.getByRole("button", { name: "Open account" })).toBeVisible();
  await expect
    .element(page.getByRole("link", { name: "Cancel" }))
    .toHaveAttribute("href", "/products/saver");
});

// Verifies: DH-2233
test("an open of a deposit product sends the product ID only, then shows the new account with a status message and the focus on its heading", async () => {
  const bodies: unknown[] = [];
  serveProducts([saver]);
  serveAccounts([opened]);
  serveOpen(bodies);

  renderApp("/products/saver/open");
  await page.getByRole("button", { name: "Open account" }).click();

  const heading = page.getByRole("heading", { level: 1, name: "Bonus Saver" });
  await expect.element(heading).toHaveFocus();
  await expect
    .element(page.getByRole("status").filter({ hasText: "Your Bonus Saver account is open." }))
    .toBeVisible();
  expect(bodies).toEqual([{ productId: "saver" }]);
  await expectNoAxeViolations(document.documentElement);
});

test("a refusal of the open shows the error of the bank", async () => {
  serveSession();
  serveProducts([saver]);
  worker.use(
    http.post("/api/accounts", ({ response }) =>
      response(400).json({
        type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
        title: "Values in the request are not valid",
        status: 400,
        detail: "One or more values in the request are not valid.",
        errors: [
          {
            pointer: "#/productId",
            detail:
              "The bank has no current product with this ID. Select a product from the list of products.",
          },
        ],
      }),
    ),
  );

  renderApp("/products/saver/open");
  await page.getByRole("button", { name: "Open account" }).click();

  await expect
    .element(page.getByText("The bank has no current product with this ID.", { exact: false }))
    .toBeVisible();
  await expect
    .element(page.getByRole("heading", { level: 1, name: "Open Bonus Saver" }))
    .toBeVisible();
});

test("an open after the end of the session goes to the sign-in page", async () => {
  serveSession();
  serveProducts([saver]);
  worker.use(
    http.post("/api/accounts", ({ response }) => {
      worker.use(
        http.get("/api/session", ({ response: session }) => session(403).json(sessionEnded)),
      );
      return response(403).json(sessionEnded);
    }),
  );

  renderApp("/products/saver/open");
  await page.getByRole("button", { name: "Open account" }).click();

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
});

// Verifies: DH-2216
test("the open page needs a session", async () => {
  serveProducts([saver]);

  renderApp("/products/saver/open");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
});

const card = product("card", { name: "Glide Card", productCategory: "CRED_AND_CHRG_CARDS" });
const loan = product("loan", { name: "Bloom Personal Loan", productCategory: "PERS_LOANS" });

// Verifies: DH-2219
test("the open page of a lending product says that the bank approves the account immediately", async () => {
  serveSession();
  serveProducts([card]);

  renderApp("/products/card/open");

  await expect
    .element(
      page.getByText("Example Bank is a demonstration bank, so it does no credit check.", {
        exact: false,
      }),
    )
    .toBeVisible();
  await expect.element(page.getByLabelText("Credit limit")).toBeVisible();
});

test("the open page of a deposit product asks for no value", async () => {
  serveSession();
  serveProducts([saver]);

  renderApp("/products/saver/open");

  await expect.element(page.getByRole("button", { name: "Open account" })).toBeVisible();
  expect(page.getByRole("textbox").query()).toBeNull();
  expect(page.getByText("does no credit check", { exact: false }).query()).toBeNull();
});

test.each([
  {
    item: card,
    label: "Credit limit",
    typed: "$5,000.50",
    want: { productId: "card", creditLimit: "5000.50" },
  },
  {
    item: loan,
    label: "Loan amount",
    typed: "20,000",
    want: { productId: "loan", loanAmount: "20000.00" },
  },
])(
  "an open of $item.name sends the $label as an AmountString",
  async ({ item, label, typed, want }) => {
    const bodies: unknown[] = [];
    serveProducts([item]);
    serveAccounts([opened]);
    serveOpen(bodies);

    renderApp(`/products/${item.productId}/open`);
    await page.getByLabelText(label).fill(typed);
    await page.getByRole("button", { name: "Open account" }).click();

    await expect
      .element(page.getByRole("status").filter({ hasText: `Your ${item.name} account is open.` }))
      .toBeVisible();
    expect(bodies).toEqual([want]);
  },
);

// Verifies: DH-2221
test("text that is not an amount gives a field error and sends no open", async () => {
  const bodies: unknown[] = [];
  serveSession();
  serveProducts([card]);
  serveOpen(bodies);

  renderApp("/products/card/open");
  await page.getByLabelText("Credit limit").fill("five thousand");
  await page.getByRole("button", { name: "Open account" }).click();

  await expect
    .element(
      page.getByRole("link", {
        name: "Enter the credit limit in dollars, for example 5,000 or 5,000.50.",
      }),
    )
    .toBeVisible();
  await expect.element(page.getByLabelText("Credit limit")).toHaveAttribute("aria-invalid", "true");
  expect(bodies).toEqual([]);
});

// Verifies: DH-2222
test("a refusal of the amount shows the error of the bank at the field", async () => {
  serveSession();
  serveProducts([loan]);
  worker.use(
    http.post("/api/accounts", ({ response }) =>
      response(400).json({
        type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
        title: "Values in the request are not valid",
        status: 400,
        detail: "One or more values in the request are not valid.",
        errors: [{ pointer: "#/loanAmount", detail: "Enter a loan amount of more than $0.00." }],
      }),
    ),
  );

  renderApp("/products/loan/open");
  await page.getByLabelText("Loan amount").fill("0");
  await page.getByRole("button", { name: "Open account" }).click();

  await expect
    .element(page.getByRole("link", { name: "Enter a loan amount of more than $0.00." }))
    .toBeVisible();
  await expect.element(page.getByLabelText("Loan amount")).toHaveAttribute("aria-invalid", "true");
});

test("the open page of a product that is not current shows the not found page", async () => {
  serveSession();
  serveProducts([]);

  renderApp("/products/ended-saver/open");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Page not found" }))
    .toBeVisible();
});

describe.each(presets)("with the preset $name", (preset) => {
  test.each(["TRANS_AND_SAVINGS_ACCOUNTS", "CRED_AND_CHRG_CARDS", "RESIDENTIAL_MORTGAGES"])(
    "the open page of %s passes the accessibility checks and has no horizontal scroll at 320 CSS px",
    async (category) => {
      const item = preset.products.find((p) => p.productCategory === category);
      if (!item) throw new Error(`the preset has no product of ${category}`);
      serveSession();

      await showPreset(preset, `/products/${item.productId}/open`);

      await expect
        .element(page.getByRole("heading", { level: 1, name: `Open ${item.name}` }))
        .toBeVisible();
      expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
      await expectNoAxeViolations(document.documentElement);
    },
  );
});
