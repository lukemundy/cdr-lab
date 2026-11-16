import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientDetail } from "../api/client";
import { definitionOf, http, renderApp, worker } from "../test/app";

const id = "7c89d9d4-4f19-4bb3-90c3-b2d5b95520c4";
const brandId = "0f3a8c2e-1d4b-4e5f-8a9b-7c6d5e4f3a21";

const loaded: DataRecipientDetail = {
  legalEntityId: id,
  legalEntityName: "Console Test Pty Ltd",
  accreditationNumber: "ADR-MOCK-0011",
  accreditationLevel: "UNRESTRICTED",
  logoUri: "https://console-test.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-23T02:05:07Z",
  version: 4,
  actions: [],
  dataRecipientBrands: [
    {
      dataRecipientBrandId: brandId,
      brandName: "Console Test",
      logoUri: "https://console-test.example/brand.png",
      status: "INACTIVE",
      actions: [],
      softwareProducts: [],
    },
  ],
};

function serveChange(status: 204 | 400) {
  const bodies: unknown[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) => response(200).json(loaded)),
    http.post(
      "/api/data-recipients/{legalEntityId}/brands/{brandId}/change",
      async ({ request }) => {
        bodies.push(await request.json());
        if (status === 204) return new Response(null, { status: 204 });
        return new Response(
          JSON.stringify({
            type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
            title: "Values in the request are not valid",
            status: 400,
            detail: "One or more values in the request are not valid.",
            errors: [{ pointer: "#/brandName", detail: "brandName is required" }],
          }),
          { status: 400, headers: { "Content-Type": "application/problem+json" } },
        );
      },
    ),
  );
  return bodies;
}

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const save = () => page.getByRole("button", { name: /^Save changes/ });

describe("the change of a data recipient brand", () => {
  // Verifies: REG-1421, REG-1422
  test("opens from the detail, and shows the current values and the brand ID as text", async () => {
    serveChange(204);
    const router = renderApp(`/data-recipients/${id}`);

    await page.getByRole("link", { name: "Change brand Console Test" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Change the details of Console Test");
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}/brands/${brandId}/change`);
    expect(definitionOf(document.body, "Brand ID").textContent).toContain(brandId);
    await expect.element(field("Name")).toHaveValue("Console Test");
    await expect.element(field("Logo URI")).toHaveValue("https://console-test.example/brand.png");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1433, REG-1434, REG-1445, REG-1447
  test("sends each field with the loaded version, then shows the detail with a toast and the brand open", async () => {
    const bodies = serveChange(204);
    const router = renderApp(`/data-recipients/${id}/brands/${brandId}/change`);

    await userEvent.fill(field("Name"), " Console Test Budgeting ");
    await save().click();

    await expect
      .element(page.getByRole("dialog", { name: "Details of Console Test Budgeting changed" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        version: 4,
        brandName: "Console Test Budgeting",
        logoUri: "https://console-test.example/brand.png",
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    await expect
      .element(page.getByRole("button", { name: "Console Test", exact: true }))
      .toHaveAttribute("aria-expanded", "true");
  });

  // Verifies: REG-1436, REG-1440
  test("shows a refused value at its field and in the error summary", async () => {
    serveChange(400);
    renderApp(`/data-recipients/${id}/brands/${brandId}/change`);

    await userEvent.fill(field("Name"), "");
    await save().click();

    await expect
      .element(page.getByRole("alert").getByRole("link", { name: "brandName is required" }))
      .toBeVisible();
    await expect.element(field("Name")).toHaveAccessibleDescription("brandName is required");
  });

  test("says that the brand does not exist when the data recipient has no brand with the ID", async () => {
    serveChange(204);
    renderApp(`/data-recipients/${id}/brands/unknown/change`);

    await expect
      .element(page.getByText("The data recipient has no brand with this ID."))
      .toBeVisible();
    await expect.element(page.getByRole("form")).not.toBeInTheDocument();
  });
});
