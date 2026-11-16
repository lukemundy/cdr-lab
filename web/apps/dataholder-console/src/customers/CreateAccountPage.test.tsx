import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { productCategoryValues, type CustomerDetail } from "../api/client";
import { account, customerDetail, http, renderApp, worker } from "../test/app";

const number = "12345678";
const newId = "5d1f7a52-6a3b-4c1e-9f0d-2b8e4c7a9e13";

type Result = { status: 201 } | { status: 400 | 409; body: unknown };

function serveCreate(before: CustomerDetail, ...results: Result[]) {
  const bodies: unknown[] = [];
  worker.use(
    http.post("/api/customers/{customerNumber}/accounts", async ({ request }) => {
      bodies.push(await request.json());
      const result = results[Math.min(bodies.length - 1, results.length - 1)] ?? { status: 201 };
      if (result.status === 201)
        return new Response(null, {
          status: 201,
          headers: { Location: `/api/customers/${number}/accounts/${newId}` },
        });
      return new Response(JSON.stringify(result.body), {
        status: result.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }),
    http.get("/api/customers/{customerNumber}", ({ response }) =>
      response(200).json(
        bodies.length > 0 && results.at(bodies.length - 1)?.status === 201
          ? { ...before, accounts: [...before.accounts, account(newId, { displayName: "Bills" })] }
          : before,
      ),
    ),
    http.get("/api/customers", ({ response }) => response(200).json({ data: [], totalRecords: 0 })),
  );
  return bodies;
}

const fieldErrorsType =
  "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002";
const displayNameError = "displayName must have 1 to 100 characters";
const categoryError = "productCategory must be a value of BankingProductCategoryV2";
const pairError = "An account has this bsb and accountNumber.";

const refusal = {
  type: fieldErrorsType,
  title: "Values in the request are not valid",
  status: 400,
  detail: "One or more values in the request are not valid.",
  errors: [
    { pointer: "#/displayName", detail: displayNameError },
    { pointer: "#/productCategory", detail: categoryError },
  ],
};

const conflict = {
  type: fieldErrorsType,
  title: "Values in the request are not valid",
  status: 409,
  detail: "One or more values in the request are not valid.",
  errors: [{ pointer: "#/accountNumber", detail: pairError }],
};

const noAccounts = customerDetail(number, { firstName: "Sam", lastName: "Sample" });
const onlyClosed = customerDetail(number, {
  firstName: "Sam",
  lastName: "Sample",
  accounts: [account("account-1", { displayName: "Old", openStatus: "CLOSED", actions: [] })],
});

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const category = () => page.getByRole("combobox", { name: "Product category" });
const create = () => page.getByRole("button", { name: /^Create account/ });

async function fillAccount() {
  await userEvent.fill(field("Display name"), " Bills ");
  await userEvent.selectOptions(category().element(), "TRANS_AND_SAVINGS_ACCOUNTS");
  await userEvent.fill(field("Product name"), "Mock Everyday Account");
  await userEvent.fill(field("BSB"), "654321 ");
  await userEvent.fill(field("Account number"), " 87654321");
}

describe("the create of an account", () => {
  test("opens from the detail of a customer with no accounts, with empty fields", async () => {
    serveCreate(noAccounts);
    const router = renderApp(`/customers/${number}`);

    await page.getByRole("link", { name: "Create account" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Create an account for Sam Sample");
    expect(router.state.location.pathname).toBe(`/customers/${number}/create-account`);
    await expect.element(field("Display name")).toHaveValue("");
    await expect.element(category()).toHaveValue("");
    await expect.element(field("Credit limit (optional)")).toHaveValue("");
    await expect
      .element(
        page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Customers" }),
      )
      .toHaveAttribute("aria-current", "page");
    await expectNoAxeViolations(document.body);
  });

  test("offers each product category of the CDS, and no category until the operator selects one", async () => {
    serveCreate(noAccounts);
    renderApp(`/customers/${number}/create-account`);

    await expect.element(category()).toBeVisible();
    const options = [...category().element().querySelectorAll("option")].map((o) => o.value);
    expect(options).toEqual(["", ...productCategoryValues]);
    expect(options).toHaveLength(14);
  });

  test("is there for a customer whose only account is closed", async () => {
    serveCreate(onlyClosed);
    renderApp(`/customers/${number}`);

    await expect.element(page.getByRole("link", { name: "Create account" })).toBeVisible();
  });

  test("sends each field with no white space, then shows the new account with a toast", async () => {
    const bodies = serveCreate(noAccounts, { status: 201 });
    const router = renderApp(`/customers/${number}/create-account`);

    await fillAccount();
    await create().click();

    await expect.element(page.getByRole("dialog", { name: "Account Bills created" })).toBeVisible();
    expect(router.state.location.pathname).toBe(`/customers/${number}`);
    await expect.element(page.getByRole("heading", { level: 3, name: "Bills" })).toBeVisible();
    expect(bodies).toEqual([
      {
        displayName: "Bills",
        productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
        productName: "Mock Everyday Account",
        bsb: "654321",
        accountNumber: "87654321",
      },
    ]);
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  test("sends the credit limit when it has a value", async () => {
    const bodies = serveCreate(noAccounts, { status: 201 });
    renderApp(`/customers/${number}/create-account`);

    await fillAccount();
    await userEvent.fill(field("Credit limit (optional)"), " 500.00 ");
    await create().click();

    await expect.element(page.getByRole("dialog", { name: /created$/ })).toBeVisible();
    expect(bodies).toEqual([expect.objectContaining({ creditLimit: "500.00" })]);
  });

  test("shows each refused value at its field and in the summary, and keeps the values", async () => {
    const bodies = serveCreate(noAccounts, { status: 400, body: refusal });
    renderApp(`/customers/${number}/create-account`);

    await userEvent.fill(field("Product name"), "Mock");
    await create().click();

    const summary = page.getByRole("alert");
    await expect.element(summary.getByText(refusal.detail)).toBeVisible();
    await expect.element(summary.getByRole("link", { name: displayNameError })).toBeVisible();
    await expect.element(summary.getByRole("link", { name: categoryError })).toBeVisible();
    await expect.element(field("Display name")).toHaveAccessibleDescription(displayNameError);
    await expect.element(category()).toHaveAccessibleDescription(categoryError);
    await expect.element(field("Product name")).toHaveValue("Mock");
    expect(bodies).toEqual([
      { displayName: "", productCategory: "", productName: "Mock", bsb: "", accountNumber: "" },
    ]);
    await expectNoAxeViolations(document.body);

    await summary.getByRole("link", { name: categoryError }).click();
    await expect.poll(() => document.activeElement).toBe(category().element());
  });

  test("shows a BSB and account number of a different account at the account number", async () => {
    serveCreate(noAccounts, { status: 409, body: conflict });
    renderApp(`/customers/${number}/create-account`);

    await fillAccount();
    await create().click();

    await expect
      .element(page.getByRole("alert").getByRole("link", { name: pairError }))
      .toBeVisible();
    await expect.element(field("Account number")).toHaveAccessibleDescription(pairError);
    await expect.element(field("Account number")).toHaveValue(" 87654321");
    await expect.element(field("BSB")).toHaveAccessibleDescription("");
  });

  test("Cancel goes back to the detail and sends nothing", async () => {
    const bodies = serveCreate(noAccounts);
    const router = renderApp(`/customers/${number}/create-account`);

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect.element(page.getByRole("heading", { level: 1, name: "Sam Sample" })).toBeVisible();
    expect(router.state.location.pathname).toBe(`/customers/${number}`);
    expect(bodies).toEqual([]);
  });
});
