import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientDetail } from "../api/client";
import { definitionOf, http, renderApp, worker, renderDetail } from "../test/app";

const id = "7c89d9d4-4f19-4bb3-90c3-b2d5b95520c4";
const productId = "5d1b7a4e-3f0c-4f6b-9a8e-2b6f1c9d0e11";
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
  dataRecipientBrands: [
    {
      dataRecipientBrandId: "b-1",
      brandName: "Console Brand",
      logoUri: "https://console-test.example/brand.png",
      status: "INACTIVE",
      actions: [],
      softwareProducts: [],
    },
  ],
};

type Result = { status: 201 } | { status: 400 | 409; body: unknown };

function serve(result: Result = { status: 201 }) {
  const bodies: unknown[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) => response(200).json(loaded)),
    http.post(
      "/api/data-recipients/{legalEntityId}/brands/{brandId}/software-products",
      async ({ request }) => {
        bodies.push(await request.json());
        if (result.status === 201)
          return new Response(null, {
            status: 201,
            headers: { Location: `/api/data-recipients/${id}/software-products/${productId}` },
          });
        return new Response(JSON.stringify(result.body), {
          status: result.status,
          headers: { "Content-Type": "application/problem+json" },
        });
      },
    ),
  );
  return bodies;
}

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const status = () => page.getByRole("combobox", { name: "Status" });
const add = () => page.getByRole("button", { name: /^Add software product/ });
const formURL = `/data-recipients/${id}/brands/b-1/add-software-product`;

async function fillRequiredFields() {
  await userEvent.fill(field("Name"), " Console App ");
  await userEvent.fill(field("Description"), "A test app");
  await userEvent.fill(field("Logo URI"), "https://console-test.example/app.png");
  await userEvent.fill(field("Client URI"), "https://console-test.example");
  await userEvent.fill(
    field("Redirect URIs"),
    " https://console-test.example/cb1 \n\n https://console-test.example/cb2",
  );
  await userEvent.fill(field("JWKS URI"), "https://console-test.example/jwks");
  await userEvent.fill(field("Revocation URI"), "https://console-test.example/revoke");
  await userEvent.fill(field("Recipient base URI"), "https://cdr.console-test.example");
  await userEvent.fill(field("Scopes"), "openid\nprofile\n");
}

describe("the add of a software product", () => {
  // Verifies: REG-1417, REG-1419
  test("opens from the brand, with a new UUID, empty fields and the status ACTIVE", async () => {
    serve();
    const router = await renderDetail(`/data-recipients/${id}`);

    await page.getByRole("link", { name: "Add software product to Console Brand" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Add a software product to Console Brand");
    expect(router.state.location.pathname).toBe(formURL);
    expect(definitionOf(document.body, "Brand ID").textContent).toContain("b-1");
    await expect.element(field("Software product ID")).toHaveValue(expect.stringMatching(uuid));
    await expect.element(status()).toHaveValue("ACTIVE");
    await expect.element(status().getByRole("option")).toHaveLength(3);
    await expect.element(field("Name")).toHaveValue("");
    await expect.element(field("Redirect URIs")).toHaveValue("");
    await expect
      .element(field("Redirect URIs"))
      .toHaveAccessibleDescription(/one URI on each line/);
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1432, REG-1434, WCON-0009
  test("sends the version, each list with no empty line, and no optional field that has no value", async () => {
    const bodies = serve();
    const router = renderApp(formURL);

    await userEvent.fill(field("Software product ID"), ` ${productId} `);
    await fillRequiredFields();
    await userEvent.fill(field("Policy URI (optional)"), " https://console-test.example/policy ");
    await userEvent.selectOptions(status(), "INACTIVE");
    await add().click();

    await expect
      .element(page.getByRole("dialog", { name: "Software product Console App added" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        version: 7,
        softwareProductId: productId,
        softwareProductName: "Console App",
        softwareProductDescription: "A test app",
        logoUri: "https://console-test.example/app.png",
        clientUri: "https://console-test.example",
        redirectUris: ["https://console-test.example/cb1", "https://console-test.example/cb2"],
        policyUri: "https://console-test.example/policy",
        jwksUri: "https://console-test.example/jwks",
        revocationUri: "https://console-test.example/revoke",
        recipientBaseUri: "https://cdr.console-test.example",
        scopes: ["openid", "profile"],
        status: "INACTIVE",
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
  });

  // Verifies: REG-1420, REG-1436
  test("shows a status that breaks a status rule at the status field", async () => {
    const rule =
      "status is ACTIVE, but data recipient brand b-1 is INACTIVE: a software product of an INACTIVE brand must be INACTIVE or REMOVED";
    serve({
      status: 409,
      body: {
        type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
        title: "Values in the request are not valid",
        status: 409,
        detail: "One or more values in the request are not valid.",
        errors: [{ pointer: "#/status", detail: rule }],
      },
    });
    renderApp(formURL);

    await fillRequiredFields();
    await add().click();

    await expect.element(page.getByRole("alert").getByRole("link", { name: rule })).toBeVisible();
    await expect.element(status()).toHaveAccessibleDescription(rule);
    await expect.element(status()).toHaveValue("ACTIVE");
    await expect.element(field("Name")).toHaveValue(" Console App ");
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

  test("says that the brand does not exist when the data recipient has no brand with the ID", async () => {
    serve();
    renderApp(`/data-recipients/${id}/brands/unknown/add-software-product`);

    await expect
      .element(page.getByText("The data recipient has no brand with this ID."))
      .toBeVisible();
    await expect.element(page.getByRole("form")).not.toBeInTheDocument();
  });
});
