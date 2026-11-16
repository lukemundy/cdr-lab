import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import {
  customer,
  http,
  renderApp,
  serveAccounts,
  serveSession,
  sessionEnded,
  worker,
} from "../test/app";

// The golden page of the data holder: the customer number page of the CDR authentication
// (services/dataholder/internal/adapter/consumerpages/golden_test.go).
const cdrPages = import.meta.glob<string>(
  "../../../../../services/dataholder/internal/adapter/consumerpages/testdata/pages/customer-number.html",
  { query: "?raw", import: "default", eager: true },
);

function fieldErrors(errors: { pointer: string; detail: string }[], status = 400) {
  return {
    type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002" as const,
    title: "Values in the request are not valid" as const,
    status,
    detail: "One or more values in the request are not valid.",
    errors,
  };
}

test("a sign-up sends the values, starts a session and shows the new customer number with the CDR text", async () => {
  let sent: unknown;
  worker.use(
    http.post("/api/customers", async ({ request, response }) => {
      sent = await request.json();
      serveSession();
      return response(201).json(customer);
    }),
  );

  renderApp("/sign-up");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Become a customer of Example Bank" }))
    .toBeVisible();
  await expect
    .element(page.getByText("Give only invented details.", { exact: false }))
    .toBeVisible();
  await expect
    .element(page.getByLabelText("Email address"))
    .toHaveAccessibleDescription(
      "Use an invented address, for example name@example.test. The bank sends each one-time password to this address.",
    );
  await userEvent.fill(page.getByLabelText("First name (optional)"), "Zebracorn");
  await userEvent.fill(page.getByLabelText("Last name"), "Quoll");
  await userEvent.fill(page.getByLabelText("Email address"), "quoll@example.test");
  await userEvent.click(page.getByRole("button", { name: "Sign up" }));

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Your customer number" }))
    .toBeVisible();
  expect(sent).toEqual({
    firstName: "Zebracorn",
    lastName: "Quoll",
    emailAddress: "quoll@example.test",
  });
  await expect.element(page.getByText("00047291", { exact: true })).toBeVisible();
  await expect.element(page.getByRole("button", { name: "Copy customer number" })).toBeVisible();
  await expect
    .element(
      page.getByText(
        "You sign in to Example Bank online banking with this customer number or with your email address.",
      ),
    )
    .toBeVisible();
  await expect
    .element(
      page.getByText(
        "When an app asks for your Example Bank data under the Consumer Data Right, Example Bank asks for your " +
          "email address or this customer number. Example Bank then sends a one-time password to your email address.",
      ),
    )
    .toBeVisible();
  await expect
    .poll(() => document.title)
    .toBe("Your customer number - Example Bank online banking");
});

test("a sign-up that the bank refuses shows each error at its field and in the error summary", async () => {
  worker.use(
    http.post("/api/customers", ({ response }) =>
      response(400).json(
        fieldErrors([
          { pointer: "#/lastName", detail: "Enter a last name of 1 to 100 characters." },
        ]),
      ),
    ),
  );

  renderApp("/sign-up");
  await userEvent.click(page.getByRole("button", { name: "Sign up" }));

  await expect
    .element(page.getByRole("alert"))
    .toHaveTextContent(
      "There is a problemOne or more values in the request are not valid.Enter a last name of 1 to 100 characters.",
    );
  await expect
    .element(page.getByLabelText("Last name"))
    .toHaveAccessibleDescription("Enter a last name of 1 to 100 characters.");
});

test("a sign-up with an email address in use shows the error at the email address", async () => {
  const inUse =
    "A customer of the bank has this email address. Sign in, or enter a different email address.";
  worker.use(
    http.post("/api/customers", ({ response }) =>
      response(409).json(fieldErrors([{ pointer: "#/emailAddress", detail: inUse }], 409)),
    ),
  );

  renderApp("/sign-up");
  await userEvent.fill(page.getByLabelText("Last name"), "Quoll");
  await userEvent.fill(page.getByLabelText("Email address"), "quoll@example.test");
  await userEvent.click(page.getByRole("button", { name: "Sign up" }));

  // The field also has its own description, before the error.
  await expect
    .element(page.getByLabelText("Email address"))
    .toHaveAccessibleDescription(
      /A customer of the bank has this email address\. Sign in, or enter a different email address\.$/,
    );
});

test("the sign-in page has the note of the demonstration bank, and fields that a password manager can fill", async () => {
  renderApp("/sign-in");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
  await expect.poll(() => document.title).toBe("Sign in - Example Bank online banking");
  await expect
    .element(
      page.getByText(
        "This is Example Bank online banking. This is a demonstration bank: any password works here, and you " +
          "must not use a real password. When an app asks for your Example Bank data under the Consumer Data Right, " +
          "you do not sign in here. Example Bank then asks only for your email address or customer number and a " +
          "one-time password, and never for your password.",
      ),
    )
    .toBeVisible();
  const identifier = page.getByLabelText("Email address or customer number");
  await expect.element(identifier).toHaveAttribute("autocomplete", "username");
  await expect.element(identifier).not.toHaveAttribute("inputmode");
  const password = page.getByLabelText("Password");
  await expect.element(password).toHaveAttribute("type", "password");
  await expect.element(password).toHaveAttribute("autocomplete", "current-password");
});

test("a sign-in sends only the identifier, and then the site shows the accounts with a sign-out control", async () => {
  let sent: unknown;
  worker.use(
    http.post("/api/session", async ({ request, response }) => {
      sent = await request.json();
      serveAccounts([]);
      return response(204).empty();
    }),
  );

  renderApp("/sign-in");
  await userEvent.fill(
    page.getByLabelText("Email address or customer number"),
    "Quoll@Example.test",
  );
  await userEvent.fill(page.getByLabelText("Password"), "correct horse battery staple");
  await userEvent.click(page.getByRole("button", { name: "Sign in" }));

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Your accounts" }))
    .toBeVisible();
  expect(sent).toEqual({ identifier: "Quoll@Example.test" });
  const navigation = page.getByRole("navigation", { name: "Main" });
  await expect.element(navigation.getByRole("button", { name: "Sign out" })).toBeVisible();
  await expect
    .element(navigation.getByRole("link", { name: "Accounts" }))
    .toHaveAttribute("href", "/accounts");
  await expect
    .element(navigation.getByRole("link", { name: "Data sharing" }))
    .toHaveAttribute("href", "/data-sharing");
});

test("a sign-in with no password sends nothing and says that any password works", async () => {
  let sent = false;
  worker.use(
    http.post("/api/session", ({ response }) => {
      sent = true;
      return response(204).empty();
    }),
  );

  renderApp("/sign-in");
  await userEvent.fill(page.getByLabelText("Email address or customer number"), "00047291");
  await userEvent.click(page.getByRole("button", { name: "Sign in" }));

  await expect
    .element(page.getByLabelText("Password"))
    .toHaveAccessibleDescription("Enter a password. Any password works.");
  expect(sent).toBe(false);
});

test("a sign-in with an identifier that no customer has shows the error at the identifier", async () => {
  const notSignedIn = "We could not sign you in. Check your email address or customer number.";
  worker.use(
    http.post("/api/session", ({ response }) =>
      response(400).json(fieldErrors([{ pointer: "#/identifier", detail: notSignedIn }])),
    ),
  );

  renderApp("/sign-in");
  await userEvent.fill(page.getByLabelText("Email address or customer number"), "00000000");
  await userEvent.fill(page.getByLabelText("Password"), "any");
  await userEvent.click(page.getByRole("button", { name: "Sign in" }));

  await expect
    .element(page.getByLabelText("Email address or customer number"))
    .toHaveAccessibleDescription(notSignedIn);
});

test("a page with the data of a customer goes to the sign-in page with no session", async () => {
  renderApp("/customer-number");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
  await expect
    .element(page.getByText("Example Bank signed you out.", { exact: false }))
    .not.toBeInTheDocument();
  await expect
    .element(page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Sign in" }))
    .toHaveAttribute("href", "/sign-in");
});

test("when the session ended, the sign-in page says that the bank signed the consumer out", async () => {
  worker.use(http.get("/api/session", ({ response }) => response(403).json(sessionEnded)));

  renderApp("/customer-number");

  await expect
    .element(page.getByRole("status").filter({ hasText: "signed you out" }))
    .toHaveTextContent("Example Bank signed you out. Sign in again.");
});

test("a sign-out ends the session and shows the sign-in page", async () => {
  serveSession();
  let signedOut = false;
  worker.use(
    http.delete("/api/session", ({ response }) => {
      signedOut = true;
      worker.resetHandlers();
      return response(204).empty();
    }),
  );

  renderApp("/");
  await userEvent.click(page.getByRole("button", { name: "Sign out" }));

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
  expect(signedOut).toBe(true);
  await expect
    .element(page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Sign in" }))
    .toBeVisible();
});

test("the sign-in page of the site and the customer number page of the CDR authentication are different", async () => {
  const [cdrPage] = Object.values(cdrPages);
  if (!cdrPage) throw new Error("the data holder has no golden customer number page");
  const cdr = new DOMParser().parseFromString(cdrPage, "text/html");

  renderApp("/sign-in");
  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
  await expect.poll(() => document.title).toBe("Sign in - Example Bank online banking");

  expect(cdr.title).not.toContain("Sign in");
  expect(cdr.title).not.toBe(document.title);
  expect(cdr.querySelector("h1")?.textContent).not.toBe(document.querySelector("h1")?.textContent);
  expect(cdr.querySelectorAll('input[type="password"]')).toHaveLength(0);
  expect(document.querySelectorAll('input[type="password"]')).toHaveLength(1);
});
