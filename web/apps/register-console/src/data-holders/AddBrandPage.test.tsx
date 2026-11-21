import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataHolderDetail } from "../api/client";
import { http, renderApp, worker } from "../test/app";

const id = "4e7a1c2b-8d3f-4a6e-9b1c-2f5d8e7a3c10";
const brandId = "6b2d9e4f-1a3c-4e5b-8d7f-0c9a2b4e6d81";
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

const loaded: DataHolderDetail = {
  legalEntityId: id,
  legalEntityName: "Delta Ltd",
  logoUri: "https://delta.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-24T02:05:07Z",
  version: 3,
  actions: [],
  dataHolderBrands: [],
};

type Result = { status: 201 } | { status: 400 | 409; body: unknown };

function serve(result: Result = { status: 201 }) {
  const bodies: unknown[] = [];
  worker.use(
    http.get("/api/data-holders/{legalEntityId}", ({ response }) => response(200).json(loaded)),
    http.post("/api/data-holders/{legalEntityId}/brands", async ({ request }) => {
      bodies.push(await request.json());
      if (result.status === 201)
        return new Response(null, {
          status: 201,
          headers: { Location: `/api/data-holders/${id}/brands/${brandId}` },
        });
      return new Response(JSON.stringify(result.body), {
        status: result.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }),
  );
  return bodies;
}

const fieldErrorsType =
  "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002";

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const status = () => page.getByRole("combobox", { name: "Status" });
const industries = () => page.getByRole("group", { name: "Industries" });
const endpoints = () => page.getByRole("group", { name: "Endpoints" });
const add = () => page.getByRole("button", { name: /^Add brand/ });
const formURL = `/data-holders/${id}/add-brand`;

describe("the add of a data holder brand", () => {
  // Verifies: REG-1417, REG-1418, REG-1419
  test("opens from the detail, with a new UUID as the brand ID, no interim ID and the status ACTIVE", async () => {
    serve();
    const router = renderApp(`/data-holders/${id}`);

    await page.getByRole("link", { name: "Add brand" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Add a brand to Delta Ltd");
    expect(router.state.location.pathname).toBe(formURL);
    await expect.element(field("Brand ID (optional)")).toHaveValue(expect.stringMatching(uuid));
    await expect.element(field("Interim ID (optional)")).toHaveValue("");
    await expect
      .element(field("Brand ID (optional)"))
      .toHaveAccessibleDescription(/A brand with no brand ID needs only its public details/);
    await expect
      .element(field("Interim ID (optional)"))
      .toHaveAccessibleDescription(/A brand with no brand ID needs only its public details/);
    await expect.element(status()).toHaveValue("ACTIVE");
    await expect.element(industries().getByRole("checkbox")).toHaveLength(4);
    await expect
      .element(endpoints().getByRole("textbox", { name: "Public base URI", exact: true }))
      .toHaveValue("");
    await expect
      .element(
        page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Data holders" }),
      )
      .toHaveAttribute("aria-current", "page");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1432, REG-1434, WCON-0008
  test("sends the version, each field with no white space and no optional field with no value, then shows the detail", async () => {
    const bodies = serve();
    const router = renderApp(formURL);

    await userEvent.fill(field("Brand ID (optional)"), ` ${brandId} `);
    await userEvent.fill(field("Name"), " Delta Bank ");
    await userEvent.fill(field("Logo URI"), "https://delta.example/brand.png ");
    await industries().getByRole("checkbox", { name: "Energy" }).click();
    await industries().getByRole("checkbox", { name: "Banking" }).click();
    await userEvent.fill(field("Version (optional)"), "1");
    await userEvent.fill(field("Public base URI"), " https://api.delta.example");
    await userEvent.fill(field("Resource base URI (optional)"), "https://mtls.delta.example");
    await userEvent.fill(field("Infosec base URI (optional)"), "https://id.delta.example");
    await userEvent.fill(field("Website URI (optional)"), "https://delta.example");
    await userEvent.fill(field("JWKS endpoint (optional)"), "https://id.delta.example/jwks");
    await userEvent.selectOptions(status(), "INACTIVE");
    await add().click();

    await expect
      .element(page.getByRole("dialog", { name: "Brand Delta Bank added" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        version: 3,
        dataHolderBrandId: brandId,
        brandName: "Delta Bank",
        industries: ["banking", "energy"],
        logoUri: "https://delta.example/brand.png",
        status: "INACTIVE",
        endpointDetail: {
          version: "1",
          publicBaseUri: "https://api.delta.example",
          resourceBaseUri: "https://mtls.delta.example",
          infosecBaseUri: "https://id.delta.example",
          websiteUri: "https://delta.example",
        },
        jwksEndpoint: "https://id.delta.example/jwks",
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  test("sends a brand with only an interim ID and its public details", async () => {
    const bodies = serve();
    renderApp(formURL);

    await userEvent.clear(field("Brand ID (optional)"));
    await userEvent.fill(field("Interim ID (optional)"), "interim-1");
    await userEvent.fill(field("Name"), "Delta Products");
    await userEvent.fill(field("Logo URI"), "https://delta.example/products.png");
    await industries().getByRole("checkbox", { name: "Telco" }).click();
    await userEvent.fill(field("Public base URI"), "https://products.delta.example");
    await add().click();

    await expect
      .poll(() => bodies)
      .toEqual([
        {
          version: 3,
          interimId: "interim-1",
          brandName: "Delta Products",
          industries: ["telco"],
          logoUri: "https://delta.example/products.png",
          status: "ACTIVE",
          endpointDetail: { publicBaseUri: "https://products.delta.example" },
        },
      ]);
  });

  // Verifies: REG-1436, REG-1437, REG-1441
  test("shows a refused endpoint at its field and a refused industry at the group, and keeps the values", async () => {
    serve({
      status: 400,
      body: {
        type: fieldErrorsType,
        title: "Values in the request are not valid",
        status: 400,
        detail: "One or more values in the request are not valid.",
        errors: [
          {
            pointer: "#/endpointDetail/publicBaseUri",
            detail: "endpointDetail.publicBaseUri is required",
          },
          { pointer: "#/industries", detail: "industries is required" },
        ],
      },
    });
    renderApp(formURL);

    await userEvent.fill(field("Name"), "Delta Bank");
    await add().click();

    const summary = page.getByRole("alert");
    await expect
      .element(summary.getByRole("link", { name: "endpointDetail.publicBaseUri is required" }))
      .toBeVisible();
    await expect
      .element(endpoints().getByRole("textbox", { name: "Public base URI", exact: true }))
      .toHaveAccessibleDescription("endpointDetail.publicBaseUri is required");
    await expect.element(industries()).toHaveAccessibleDescription("industries is required");
    await expect.element(field("Name")).toHaveValue("Delta Bank");
    await expectNoAxeViolations(document.body);

    await summary.getByRole("link", { name: "industries is required" }).click();

    await expect
      .poll(() => document.activeElement)
      .toBe(industries().getByRole("checkbox", { name: "Banking" }).element());
  });

  // Verifies: REG-1436
  test("shows an interim ID in use at its field", async () => {
    const notUnique = "interimId is not unique: a different record has it";
    serve({
      status: 409,
      body: {
        type: fieldErrorsType,
        title: "Values in the request are not valid",
        status: 409,
        detail: "One or more values in the request are not valid.",
        errors: [{ pointer: "#/interimId", detail: notUnique }],
      },
    });
    renderApp(formURL);

    await userEvent.fill(field("Interim ID (optional)"), "interim-1");
    await add().click();

    await expect
      .element(field("Interim ID (optional)"))
      .toHaveAccessibleDescription(new RegExp(`${notUnique}$`));
    await expect.element(field("Interim ID (optional)")).toHaveValue("interim-1");
  });

  // Verifies: REG-1448, REG-1449
  test("Cancel goes back to the detail and sends nothing", async () => {
    const bodies = serve();
    const router = renderApp(formURL);

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect.element(page.getByRole("heading", { level: 1, name: "Delta Ltd" })).toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
    expect(bodies).toEqual([]);
  });
});
