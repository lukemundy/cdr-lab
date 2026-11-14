import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataHolderDetail } from "../api/client";
import { definitionOf, http, renderApp, worker } from "../test/app";

const id = "3a8e5c1d-9f2b-4d7a-8e6c-1b4f7a2d9c05";

const loaded: DataHolderDetail = {
  legalEntityId: id,
  legalEntityName: "Delta Ltd",
  logoUri: "https://delta.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-23T02:05:07Z",
  version: 4,
  actions: [],
  dataHolderBrands: [],
};

const changed: DataHolderDetail = { ...loaded, legalEntityName: "Delta Group Ltd", version: 5 };

type Result = { status: 204 } | { status: 400 | 409; body: unknown };

// The detail depends on the change requests, not on the count of reads, because a read of a previous
// test can arrive late.
function serve(...results: Result[]) {
  const bodies: unknown[] = [];
  worker.use(
    http.get("/api/data-holders/{legalEntityId}", ({ response }) =>
      response(200).json(bodies.length === 0 ? loaded : changed),
    ),
    http.post("/api/data-holders/{legalEntityId}/change", async ({ request }) => {
      bodies.push(await request.json());
      const result = results[Math.min(bodies.length - 1, results.length - 1)] ?? { status: 204 };
      if (result.status === 204) return new Response(null, { status: 204 });
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
    { pointer: "#/legalEntityName", detail: "legalEntityName is longer than 36 characters" },
  ],
};

const conflict = {
  type: "about:blank",
  title: "Conflict",
  status: 409,
  detail: "The participant changed after the version in the request. The Register changed nothing.",
};

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const save = () => page.getByRole("button", { name: /^Save changes/ });

describe("the change of the details of a data holder", () => {
  test("opens from the detail, and shows the current values and the legal entity ID as text", async () => {
    serve();
    const router = renderApp(`/data-holders/${id}`);

    await page
      .getByRole("region", { name: "Legal entity", exact: true })
      .getByRole("link", { name: "Change details of Delta Ltd" })
      .click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Change the details of Delta Ltd");
    expect(router.state.location.pathname).toBe(`/data-holders/${id}/change`);
    expect(definitionOf(document.body, "Legal entity ID").textContent).toContain(id);
    await expect.element(field("Name")).toHaveValue("Delta Ltd");
    await expect.element(field("Logo URI")).toHaveValue("https://delta.example/logo.png");
    await expect
      .element(page.getByRole("textbox", { name: /Legal entity ID/ }))
      .not.toBeInTheDocument();
    await expect
      .element(
        page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Data holders" }),
      )
      .toHaveAttribute("aria-current", "page");
    await expectNoAxeViolations(document.body);
  });

  test("sends each field with the loaded version and no white space, then shows the detail with a toast", async () => {
    const bodies = serve({ status: 204 });
    const router = renderApp(`/data-holders/${id}/change`);

    await userEvent.fill(field("Name"), "  Delta Group Ltd ");
    await save().click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Delta Group Ltd" }))
      .toBeVisible();
    expect(bodies).toEqual([
      { version: 4, legalEntityName: "Delta Group Ltd", logoUri: "https://delta.example/logo.png" },
    ]);
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
    await expect
      .element(page.getByRole("dialog", { name: "Details of Delta Group Ltd changed" }))
      .toBeVisible();
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  test("shows a refused value at its field and keeps the values and the version", async () => {
    const bodies = serve(
      { status: 400, body: refusedValues },
      { status: 400, body: refusedValues },
    );
    renderApp(`/data-holders/${id}/change`);

    await userEvent.fill(field("Name"), "A name that is longer than the limit of a data holder");
    await save().click();

    const summary = page.getByRole("alert");
    await expect
      .element(summary.getByRole("link", { name: "legalEntityName is longer than 36 characters" }))
      .toBeVisible();
    await expect
      .element(field("Name"))
      .toHaveAccessibleDescription("legalEntityName is longer than 36 characters");
    await expect
      .element(field("Name"))
      .toHaveValue("A name that is longer than the limit of a data holder");
    await expectNoAxeViolations(document.body);

    await save().click();

    await expect.poll(() => bodies.length).toBe(2);
    expect(bodies[1]).toMatchObject({ version: 4 });
  });

  test("shows a stale version in the error summary", async () => {
    serve({ status: 409, body: conflict });
    renderApp(`/data-holders/${id}/change`);

    await save().click();

    const summary = page.getByRole("alert");
    await expect.element(summary.getByText(conflict.detail)).toBeVisible();
    await expect.element(summary.getByRole("link")).not.toBeInTheDocument();
  });

  test("Cancel goes back to the detail and sends nothing", async () => {
    const bodies = serve({ status: 204 });
    const router = renderApp(`/data-holders/${id}/change`);

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect.element(page.getByRole("heading", { level: 1, name: "Delta Ltd" })).toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
    expect(bodies).toEqual([]);
  });
});
