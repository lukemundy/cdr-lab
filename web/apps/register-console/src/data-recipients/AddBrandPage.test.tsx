import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientDetail } from "../api/client";
import { http, renderApp, worker } from "../test/app";

const id = "7c89d9d4-4f19-4bb3-90c3-b2d5b95520c4";
const brandId = "0b6f4b55-2d0e-4c55-9a43-5b8e1f3b7a10";
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

const loaded: DataRecipientDetail = {
  legalEntityId: id,
  legalEntityName: "Console Test Pty Ltd",
  accreditationNumber: "ADR-MOCK-0011",
  accreditationLevel: "UNRESTRICTED",
  logoUri: "https://console-test.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-23T02:05:07Z",
  version: 7,
  actions: [],
  dataRecipientBrands: [],
};

type Result = { status: 201 } | { status: 400 | 409; body: unknown };

function serve(result: Result = { status: 201 }) {
  const bodies: unknown[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) => response(200).json(loaded)),
    http.post("/api/data-recipients/{legalEntityId}/brands", async ({ request }) => {
      bodies.push(await request.json());
      if (result.status === 201)
        return new Response(null, {
          status: 201,
          headers: { Location: `/api/data-recipients/${id}/brands/${brandId}` },
        });
      return new Response(JSON.stringify(result.body), {
        status: result.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }),
  );
  return bodies;
}

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const status = () => page.getByRole("combobox", { name: "Status" });
const add = () => page.getByRole("button", { name: /^Add brand/ });
const formURL = `/data-recipients/${id}/add-brand`;

describe("the add of a data recipient brand", () => {
  // Verifies: REG-1417, REG-1419
  test("opens from the detail, with a new UUID and the status ACTIVE", async () => {
    serve();
    const router = renderApp(`/data-recipients/${id}`);

    await page.getByRole("link", { name: "Add brand" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Add a brand to Console Test Pty Ltd");
    expect(router.state.location.pathname).toBe(formURL);
    await expect.element(field("Brand ID")).toHaveValue(expect.stringMatching(uuid));
    await expect.element(status()).toHaveValue("ACTIVE");
    await expect.element(status().getByRole("option")).toHaveLength(3);
    await expect.element(field("Name")).toHaveValue("");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1430, REG-1434, REG-1445
  test("sends the version that it loaded and each field with no white space, then shows the detail with a toast", async () => {
    const bodies = serve();
    const router = renderApp(formURL);

    await userEvent.fill(field("Brand ID"), ` ${brandId} `);
    await userEvent.fill(field("Name"), " Console Brand ");
    await userEvent.fill(field("Logo URI"), "https://console-test.example/brand.png ");
    await userEvent.selectOptions(status(), "INACTIVE");
    await add().click();

    await expect
      .element(page.getByRole("dialog", { name: "Brand Console Brand added" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        version: 7,
        dataRecipientBrandId: brandId,
        brandName: "Console Brand",
        logoUri: "https://console-test.example/brand.png",
        status: "INACTIVE",
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  // Verifies: REG-1436, REG-1441
  test("shows a brand ID in use at its field, and keeps the values", async () => {
    const notUnique = "dataRecipientBrandId is not unique: a different record has it";
    serve({
      status: 409,
      body: {
        type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
        title: "Values in the request are not valid",
        status: 409,
        detail: "One or more values in the request are not valid.",
        errors: [{ pointer: "#/dataRecipientBrandId", detail: notUnique }],
      },
    });
    renderApp(formURL);

    await userEvent.fill(field("Brand ID"), brandId);
    await userEvent.fill(field("Name"), "Console Brand");
    await add().click();

    await expect
      .element(page.getByRole("alert").getByRole("link", { name: notUnique }))
      .toBeVisible();
    await expect
      .element(field("Brand ID"))
      .toHaveAccessibleDescription(new RegExp(`${notUnique}$`));
    await expect.element(field("Brand ID")).toHaveValue(brandId);
    await expect.element(field("Name")).toHaveValue("Console Brand");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1438
  test("shows a stale version in the error summary", async () => {
    const detail =
      "The data recipient changed after the version in the request. The Register changed nothing.";
    serve({ status: 409, body: { type: "about:blank", title: "Conflict", status: 409, detail } });
    renderApp(formURL);

    await add().click();

    await expect.element(page.getByRole("alert").getByText(detail)).toBeVisible();
  });

  // Verifies: REG-1448, REG-1449
  test("Cancel goes back to the detail and sends nothing", async () => {
    const bodies = serve();
    const router = renderApp(formURL);

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Console Test Pty Ltd" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    expect(bodies).toEqual([]);
  });
});
