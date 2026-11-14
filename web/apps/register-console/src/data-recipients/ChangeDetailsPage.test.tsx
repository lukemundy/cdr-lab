import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientDetail } from "../api/client";
import { definitionOf, http, renderApp, worker } from "../test/app";

const id = "7c89d9d4-4f19-4bb3-90c3-b2d5b95520c4";

const loaded: DataRecipientDetail = {
  legalEntityId: id,
  legalEntityName: "Console Test Pty Ltd",
  accreditationNumber: "ADR-MOCK-0011",
  accreditationLevel: "UNRESTRICTED",
  logoUri: "https://console-test.example/logo.png",
  status: "SUSPENDED",
  lastUpdated: "2026-10-23T02:05:07Z",
  version: 7,
  actions: [],
  dataRecipientBrands: [],
};

const changed: DataRecipientDetail = {
  ...loaded,
  legalEntityName: "Console Test Group Pty Ltd",
  accreditationLevel: "SPONSORED",
  version: 8,
};

type Result = { status: 204 } | { status: 400 | 409; body: unknown } | { status: 500 };

// The detail depends on the change requests, not on the count of reads, because a read of a previous
// test can arrive late.
function serve(...results: Result[]) {
  const bodies: unknown[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) =>
      response(200).json(bodies.length === 0 ? loaded : changed),
    ),
    http.post("/api/data-recipients/{legalEntityId}/change", async ({ request }) => {
      bodies.push(await request.json());
      const result = results[Math.min(bodies.length - 1, results.length - 1)] ?? { status: 204 };
      if (result.status === 204) return new Response(null, { status: 204 });
      if (result.status === 500) return new Response("not JSON", { status: 500 });
      return new Response(JSON.stringify(result.body), {
        status: result.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }),
  );
  return bodies;
}

const refusedValues = {
  type: "https://github.com/lukemundy/cdr-lab/blob/main/services/register/DESIGN.md#console-api",
  title: "Values in the request are not valid",
  status: 400,
  detail: "One or more values in the request are not valid.",
  errors: [
    { pointer: "#/accreditationNumber", detail: "accreditationNumber is required" },
    { pointer: "#/logoUri", detail: "logoUri is longer than 1000 characters" },
  ],
};

const conflict = {
  type: "about:blank",
  title: "Conflict",
  status: 409,
  detail:
    "The data recipient changed after the version in the request. The Register changed nothing.",
};

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const level = () => page.getByRole("combobox", { name: "Accreditation level" });
const save = () => page.getByRole("button", { name: /^Save changes/ });

describe("the change of the details of a data recipient", () => {
  test("opens from the detail, and shows the current values and the legal entity ID as text", async () => {
    serve();
    const router = renderApp(`/data-recipients/${id}`);

    await page.getByRole("link", { name: "Change details of Console Test Pty Ltd" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Change the details of Console Test Pty Ltd");
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}/change`);
    expect(definitionOf(document.body, "Legal entity ID").textContent).toContain(id);
    await expect.element(field("Name")).toHaveValue("Console Test Pty Ltd");
    await expect.element(field("Accreditation number")).toHaveValue("ADR-MOCK-0011");
    await expect.element(level()).toHaveValue("UNRESTRICTED");
    await expect.element(field("Logo URI")).toHaveValue("https://console-test.example/logo.png");
    await expect
      .element(page.getByRole("textbox", { name: /Legal entity ID/ }))
      .not.toBeInTheDocument();
    await expectNoAxeViolations(document.body);
  });

  test("sends each field with the loaded version and no white space, then shows the detail with a toast", async () => {
    const bodies = serve({ status: 204 });
    const router = renderApp(`/data-recipients/${id}/change`);

    await userEvent.fill(field("Name"), "  Console Test Group Pty Ltd ");
    await userEvent.selectOptions(level(), "SPONSORED");
    await save().click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Console Test Group Pty Ltd" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        version: 7,
        legalEntityName: "Console Test Group Pty Ltd",
        accreditationNumber: "ADR-MOCK-0011",
        accreditationLevel: "SPONSORED",
        logoUri: "https://console-test.example/logo.png",
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    await expect
      .element(page.getByRole("dialog", { name: "Details of Console Test Group Pty Ltd changed" }))
      .toBeVisible();
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  test("shows each refused value at its field and in the error summary, and keeps the values and the version", async () => {
    // The Register gives a new version after the refusal, for example after a change in a different tab.
    const bodies = serve(
      { status: 400, body: refusedValues },
      { status: 400, body: refusedValues },
    );
    renderApp(`/data-recipients/${id}/change`);

    await userEvent.fill(field("Accreditation number"), " ");
    await userEvent.fill(field("Logo URI"), "https://console-test.example/new.png");
    await save().click();

    const summary = page.getByRole("alert");
    await expect
      .element(summary.getByRole("heading", { name: "There is a problem" }))
      .toBeVisible();
    await expect.element(summary.getByRole("link")).toHaveLength(2);
    expect(document.activeElement).toBe(summary.element().parentElement);
    await expect
      .element(field("Accreditation number"))
      .toHaveAccessibleDescription("accreditationNumber is required");
    await expect
      .element(field("Logo URI"))
      .toHaveAccessibleDescription("logoUri is longer than 1000 characters");
    await expect.element(field("Logo URI")).toHaveValue("https://console-test.example/new.png");
    await expect.element(field("Name")).not.toHaveAttribute("aria-invalid");
    await expectNoAxeViolations(document.body);

    await save().click();

    await expect.poll(() => bodies.length).toBe(2);
    expect(bodies[1]).toMatchObject({ version: 7, accreditationNumber: "" });
  });

  test("shows a stale version in the error summary, and changes no field", async () => {
    serve({ status: 409, body: conflict });
    renderApp(`/data-recipients/${id}/change`);

    await save().click();

    const summary = page.getByRole("alert");
    await expect.element(summary.getByText(conflict.detail)).toBeVisible();
    await expect.element(summary.getByRole("link")).not.toBeInTheDocument();
    await expect.element(field("Name")).not.toHaveAttribute("aria-invalid");
  });

  test("shows a failure with no problem details in the error summary", async () => {
    serve({ status: 500 });
    renderApp(`/data-recipients/${id}/change`);

    await save().click();

    await expect
      .element(page.getByRole("alert").getByText(/^The change did not complete\./))
      .toBeVisible();
  });

  test("Cancel goes back to the detail and sends nothing", async () => {
    const bodies = serve({ status: 204 });
    const router = renderApp(`/data-recipients/${id}/change`);

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Console Test Pty Ltd" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    expect(bodies).toEqual([]);
  });
});
