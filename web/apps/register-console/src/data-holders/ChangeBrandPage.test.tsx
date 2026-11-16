import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataHolderBrand, DataHolderDetail } from "../api/client";
import { definitionOf, http, renderApp, worker } from "../test/app";

const id = "2c4e6a8b-0d1f-4a3c-8e5b-7d9f1b3c5e70";
const brandId = "9a1b3c5d-7e9f-4b2d-8c4e-6f8a0b2c4d61";
const interimId = "interim-products";

const bank: DataHolderBrand = {
  dataHolderBrandId: brandId,
  interimId: "interim-bank",
  brandName: "Delta Bank",
  brandGroup: "Delta Group",
  industries: ["banking", "non-bank-lending"],
  logoUri: "https://delta.example/brand.png",
  status: "INACTIVE",
  actions: [],
  serverCertificateActions: [],
  endpointDetail: {
    version: "1",
    publicBaseUri: "https://api.delta.example",
    productBaseUri: "https://api.delta.example/products",
    resourceBaseUri: "https://mtls.delta.example",
    infosecBaseUri: "https://id.delta.example",
    websiteUri: "https://delta.example",
  },
  jwksEndpoint: "https://id.delta.example/jwks",
};

const products: DataHolderBrand = {
  interimId,
  brandName: "Delta Products",
  industries: ["telco"],
  logoUri: "https://delta.example/products.png",
  status: "ACTIVE",
  actions: [],
  serverCertificateActions: [],
  endpointDetail: { publicBaseUri: "https://products.delta.example" },
};

const loaded: DataHolderDetail = {
  legalEntityId: id,
  legalEntityName: "Delta Ltd",
  logoUri: "https://delta.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-24T02:05:07Z",
  version: 5,
  actions: [],
  dataHolderBrands: [bank, products],
};

type Result = { status: 204 } | { status: 400 | 409; body: unknown };

function serve(result: Result = { status: 204 }) {
  const requests: { brandId: string; body: unknown }[] = [];
  worker.use(
    http.get("/api/data-holders/{legalEntityId}", ({ response }) => response(200).json(loaded)),
    http.post(
      "/api/data-holders/{legalEntityId}/brands/{brandId}/change",
      async ({ params, request }) => {
        requests.push({ brandId: params.brandId, body: await request.json() });
        if (result.status === 204) return new Response(null, { status: 204 });
        return new Response(JSON.stringify(result.body), {
          status: result.status,
          headers: { "Content-Type": "application/problem+json" },
        });
      },
    ),
  );
  return requests;
}

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const industries = () => page.getByRole("group", { name: "Industries" });
const save = () => page.getByRole("button", { name: /^Save changes/ });

describe("the change of a data holder brand", () => {
  test("opens from the detail, and shows the current values and the identifiers as text", async () => {
    serve();
    const router = renderApp(`/data-holders/${id}`);

    await page
      .getByRole("region", { name: "Delta Bank", exact: true })
      .getByRole("link", { name: "Change brand Delta Bank" })
      .click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Change the details of Delta Bank");
    expect(router.state.location.pathname).toBe(`/data-holders/${id}/brands/${brandId}/change`);
    expect(definitionOf(document.body, "Brand ID").textContent).toContain(brandId);
    expect(definitionOf(document.body, "Interim ID").textContent).toContain("interim-bank");
    await expect.element(page.getByRole("textbox", { name: /Brand ID/ })).not.toBeInTheDocument();
    await expect.element(page.getByRole("combobox", { name: "Status" })).not.toBeInTheDocument();
    await expect.element(field("Name")).toHaveValue("Delta Bank");
    await expect.element(field("Brand group (optional)")).toHaveValue("Delta Group");
    await expect.element(industries().getByRole("checkbox", { name: "Banking" })).toBeChecked();
    await expect
      .element(industries().getByRole("checkbox", { name: "Non-bank lending" }))
      .toBeChecked();
    await expect.element(industries().getByRole("checkbox", { name: "Energy" })).not.toBeChecked();
    await expect
      .element(field("Product base URI (optional)"))
      .toHaveValue("https://api.delta.example/products");
    await expect.element(field("Extension base URI (optional)")).toHaveValue("");
    await expect
      .element(field("JWKS endpoint (optional)"))
      .toHaveValue("https://id.delta.example/jwks");
    await expect
      .element(
        page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Data holders" }),
      )
      .toHaveAttribute("aria-current", "page");
    await expectNoAxeViolations(document.body);
  });

  test("sends each field with the loaded version, and an optional field with no value as no field", async () => {
    const requests = serve();
    const router = renderApp(`/data-holders/${id}/brands/${brandId}/change`);

    await userEvent.fill(field("Name"), " Delta Banking ");
    await userEvent.clear(field("Brand group (optional)"));
    await userEvent.clear(field("Product base URI (optional)"));
    await industries().getByRole("checkbox", { name: "Non-bank lending" }).click();
    await industries().getByRole("checkbox", { name: "Energy" }).click();
    await save().click();

    await expect
      .element(page.getByRole("dialog", { name: "Details of Delta Banking changed" }))
      .toBeVisible();
    expect(requests).toEqual([
      {
        brandId,
        body: {
          version: 5,
          brandName: "Delta Banking",
          industries: ["banking", "energy"],
          logoUri: "https://delta.example/brand.png",
          endpointDetail: {
            version: "1",
            publicBaseUri: "https://api.delta.example",
            resourceBaseUri: "https://mtls.delta.example",
            infosecBaseUri: "https://id.delta.example",
            websiteUri: "https://delta.example",
          },
          jwksEndpoint: "https://id.delta.example/jwks",
        },
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
  });

  test("a brand with only an interim ID has the interim ID in its URL", async () => {
    const requests = serve();
    renderApp(`/data-holders/${id}`);

    await page
      .getByRole("region", { name: "Delta Products", exact: true })
      .getByRole("link", { name: "Change brand Delta Products" })
      .click();
    await expect.element(field("Name")).toHaveValue("Delta Products");
    expect(definitionOf(document.body, "Brand ID").textContent).toBe("Not set");
    await save().click();

    await expect.poll(() => requests.map((r) => r.brandId)).toEqual([interimId]);
    expect(requests[0]?.body).toEqual({
      version: 5,
      brandName: "Delta Products",
      industries: ["telco"],
      logoUri: "https://delta.example/products.png",
      endpointDetail: { publicBaseUri: "https://products.delta.example" },
    });
  });

  test("shows a refused endpoint at its field and a refused industry at the group", async () => {
    serve({
      status: 400,
      body: {
        type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
        title: "Values in the request are not valid",
        status: 400,
        detail: "One or more values in the request are not valid.",
        errors: [
          {
            pointer: "#/endpointDetail/websiteUri",
            detail: "endpointDetail.websiteUri is required",
          },
          { pointer: "#/industries", detail: "industries is required" },
        ],
      },
    });
    renderApp(`/data-holders/${id}/brands/${brandId}/change`);

    await userEvent.clear(field("Website URI (optional)"));
    await save().click();

    await expect
      .element(field("Website URI (optional)"))
      .toHaveAccessibleDescription("endpointDetail.websiteUri is required");
    await expect.element(industries()).toHaveAccessibleDescription("industries is required");
    await expectNoAxeViolations(document.body);
  });

  test("says that the brand does not exist when the data holder has no brand with the ID", async () => {
    serve();
    renderApp(`/data-holders/${id}/brands/brand-unknown/change`);

    await expect
      .element(page.getByText("The data holder has no brand with this ID."))
      .toBeVisible();
    await expect.element(page.getByRole("form")).not.toBeInTheDocument();
  });

  test("Cancel goes back to the detail and sends nothing", async () => {
    const requests = serve();
    const router = renderApp(`/data-holders/${id}/brands/${brandId}/change`);

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect.element(page.getByRole("heading", { level: 1, name: "Delta Ltd" })).toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
    expect(requests).toEqual([]);
  });
});
