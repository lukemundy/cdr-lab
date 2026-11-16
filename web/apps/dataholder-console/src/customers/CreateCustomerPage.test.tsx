import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { customer, customerDetail, http, renderApp, worker } from "../test/app";

const number = "00047291";

type Result = { status: 201 } | { status: 400; body: unknown };

function serveCreate(...results: Result[]) {
  const bodies: unknown[] = [];
  worker.use(
    http.post("/api/customers", async ({ request }) => {
      bodies.push(await request.json());
      const result = results[Math.min(bodies.length - 1, results.length - 1)] ?? { status: 201 };
      if (result.status === 201)
        return new Response(null, {
          status: 201,
          headers: { Location: `/api/customers/${number}` },
        });
      return new Response(JSON.stringify(result.body), {
        status: result.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }),
    http.get("/api/customers/{customerNumber}", ({ response }) =>
      response(200).json(customerDetail(number, { firstName: "Sam", lastName: "Sample" })),
    ),
    http.get("/api/customers", ({ response }) =>
      response(200).json({ data: [customer("12345678")], totalRecords: 1 }),
    ),
  );
  return bodies;
}

const lastNameError = "lastName must have 1 to 100 characters";
const emailError = "emailAddress must be one email address";

const refusal = {
  type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
  title: "Values in the request are not valid",
  status: 400,
  detail: "One or more values in the request are not valid.",
  errors: [
    { pointer: "#/lastName", detail: lastNameError },
    { pointer: "#/emailAddress", detail: emailError },
  ],
};

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const create = () => page.getByRole("button", { name: /^Create customer/ });

describe("the create of a customer", () => {
  test("opens from the list, with empty fields", async () => {
    serveCreate();
    const router = renderApp("/customers");

    await page.getByRole("link", { name: "Create customer" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Create a customer");
    expect(router.state.location.pathname).toBe("/create/customer");
    await expect.element(field("First name (optional)")).toHaveValue("");
    await expect.element(field("Last name")).toHaveValue("");
    await expect.element(field("Email address")).toHaveValue("");
    await expect
      .element(
        page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Customers" }),
      )
      .toHaveAttribute("aria-current", "page");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: DH-1453, DH-1455
  test("sends each field with no white space, then shows the detail with a toast", async () => {
    const bodies = serveCreate({ status: 201 });
    const router = renderApp("/create/customer");

    await userEvent.fill(field("First name (optional)"), " Sam ");
    await userEvent.fill(field("Last name"), " Sample");
    await userEvent.fill(field("Email address"), "sam@mail.test ");
    await create().click();

    await expect.element(page.getByRole("heading", { level: 1, name: "Sam Sample" })).toBeVisible();
    expect(bodies).toEqual([
      { firstName: "Sam", lastName: "Sample", emailAddress: "sam@mail.test" },
    ]);
    expect(router.state.location.pathname).toBe(`/customers/${number}`);
    await expect
      .element(page.getByRole("dialog", { name: `Customer ${number} created` }))
      .toBeVisible();
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  // Verifies: DH-1453
  test("sends no first name when it has no value", async () => {
    const bodies = serveCreate({ status: 201 });
    renderApp("/create/customer");

    await userEvent.fill(field("First name (optional)"), "  ");
    await userEvent.fill(field("Last name"), "Sample");
    await userEvent.fill(field("Email address"), "sam@mail.test");
    await create().click();

    await expect.element(page.getByRole("dialog", { name: /created$/ })).toBeVisible();
    expect(bodies).toEqual([{ lastName: "Sample", emailAddress: "sam@mail.test" }]);
  });

  // Verifies: DH-1453
  test("shows each refused value at its field and in the summary, and keeps the values", async () => {
    const bodies = serveCreate({ status: 400, body: refusal });
    renderApp("/create/customer");

    await userEvent.fill(field("First name (optional)"), "Sam");
    await userEvent.fill(field("Email address"), " sam at mail.test");
    await create().click();

    const summary = page.getByRole("alert");
    await expect.element(summary.getByText(refusal.detail)).toBeVisible();
    await expect.element(summary.getByRole("link", { name: lastNameError })).toBeVisible();
    await expect.element(summary.getByRole("link", { name: emailError })).toBeVisible();
    await expect.element(field("Last name")).toHaveAccessibleDescription(lastNameError);
    await expect
      .element(field("Email address"))
      .toHaveAccessibleDescription(
        `The data holder sends each one-time password of the customer to this address. ${emailError}`,
      );
    await expect.element(field("First name (optional)")).toHaveValue("Sam");
    await expect.element(field("Email address")).toHaveValue(" sam at mail.test");
    expect(bodies).toHaveLength(1);
    await expectNoAxeViolations(document.body);

    await summary.getByRole("link", { name: emailError }).click();
    await expect.poll(() => document.activeElement).toBe(field("Email address").element());
  });

  // Verifies: DH-1453
  test("Cancel goes back to the list and sends nothing", async () => {
    const bodies = serveCreate();
    const router = renderApp("/create/customer");

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect.element(page.getByRole("heading", { level: 1, name: "Customers" })).toBeVisible();
    expect(router.state.location.pathname).toBe("/customers");
    expect(bodies).toEqual([]);
  });
});
