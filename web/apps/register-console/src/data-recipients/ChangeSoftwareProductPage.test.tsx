import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientDetail, SoftwareProduct } from "../api/client";
import { definitionOf, http, renderApp, worker, renderDetail } from "../test/app";

const id = "7c89d9d4-4f19-4bb3-90c3-b2d5b95520c4";
const productId = "5d1b7a4e-3f0c-4f6b-9a8e-2b6f1c9d0e11";

const product: SoftwareProduct = {
  softwareProductId: productId,
  softwareProductName: "Console Test App",
  softwareProductDescription: "A test app",
  logoUri: "https://console-test.example/app.png",
  status: "ACTIVE",
  clientUri: "https://console-test.example",
  sectorIdentifierUri: "https://console-test.example/sector",
  redirectUris: ["https://console-test.example/cb1", "https://console-test.example/cb2"],
  jwksUri: "https://console-test.example/jwks",
  revocationUri: "https://console-test.example/revoke",
  recipientBaseUri: "https://cdr.console-test.example",
  scopes: ["openid", "profile"],
  clientCertificateThumbprints: [],
  actions: [],
  clientCertificateActions: ["register"],
};

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
      brandName: "Console Test",
      logoUri: "https://console-test.example/brand.png",
      status: "ACTIVE",
      actions: [],
      softwareProducts: [product],
    },
  ],
};

function serveDetail() {
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) => response(200).json(loaded)),
  );
}

type Result = { status: 204 } | { status: 400 | 409; body: unknown };

function serveChange(result: Result) {
  const bodies: unknown[] = [];
  worker.use(
    http.post(
      "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/change",
      async ({ request }) => {
        bodies.push(await request.json());
        if (result.status === 204) return new Response(null, { status: 204 });
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
const save = () => page.getByRole("button", { name: /^Save changes/ });
const formURL = `/data-recipients/${id}/software-products/${productId}/change`;

describe("the change of a software product", () => {
  // Verifies: REG-1421, REG-1425
  test("opens from the detail, and shows the current values, with one value of a list on each line", async () => {
    serveDetail();
    const router = await renderDetail(`/data-recipients/${id}`);

    await page.getByRole("link", { name: "Change software product Console Test App" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Change the details of Console Test App");
    expect(router.state.location.pathname).toBe(formURL);
    expect(definitionOf(document.body, "Software product ID").textContent).toContain(productId);
    await expect.element(field("Name")).toHaveValue("Console Test App");
    await expect.element(field("Description")).toHaveValue("A test app");
    await expect
      .element(field("Sector identifier URI (optional)"))
      .toHaveValue("https://console-test.example/sector");
    await expect.element(field("Terms of service URI (optional)")).toHaveValue("");
    await expect
      .element(field("Redirect URIs"))
      .toHaveValue("https://console-test.example/cb1\nhttps://console-test.example/cb2");
    await expect
      .element(field("Redirect URIs"))
      .toHaveAccessibleDescription(/one URI on each line/);
    await expect.element(field("Scopes")).toHaveValue("openid\nprofile");
    await expect.element(field("Scopes")).toHaveAccessibleDescription(/one scope on each line/);
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1432, WCON-0009
  test("sends each list with no white space and no empty line, and no optional field that has no value", async () => {
    serveDetail();
    const bodies = serveChange({ status: 204 });
    const router = renderApp(formURL);

    await userEvent.fill(
      field("Redirect URIs"),
      " https://console-test.example/new \n\n  https://console-test.example/cb2\n",
    );
    await userEvent.fill(field("Sector identifier URI (optional)"), "  ");
    await userEvent.fill(field("Policy URI (optional)"), " https://console-test.example/policy ");
    await save().click();

    await expect
      .element(page.getByRole("dialog", { name: "Details of Console Test App changed" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        version: 7,
        softwareProductName: "Console Test App",
        softwareProductDescription: "A test app",
        logoUri: "https://console-test.example/app.png",
        clientUri: "https://console-test.example",
        redirectUris: ["https://console-test.example/new", "https://console-test.example/cb2"],
        policyUri: "https://console-test.example/policy",
        jwksUri: "https://console-test.example/jwks",
        revocationUri: "https://console-test.example/revoke",
        recipientBaseUri: "https://cdr.console-test.example",
        scopes: ["openid", "profile"],
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
    for (const name of ["Console Test", "Console Test App"]) {
      expect(page.getByRole("button", { name, exact: true }).element()).toHaveAttribute(
        "aria-expanded",
        "true",
      );
    }
  });

  // Verifies: REG-1437, REG-1440
  test("shows the error of a list at the field of the list and in the error summary", async () => {
    serveDetail();
    serveChange({
      status: 400,
      body: {
        type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
        title: "Values in the request are not valid",
        status: 400,
        detail: "One or more values in the request are not valid.",
        errors: [{ pointer: "#/redirectUris", detail: "redirectUris is required" }],
      },
    });
    renderApp(formURL);

    await userEvent.fill(field("Redirect URIs"), "\n \n");
    await save().click();

    const summary = page.getByRole("alert");
    await expect
      .element(summary.getByRole("link", { name: "redirectUris is required" }))
      .toBeVisible();
    await expect
      .element(field("Redirect URIs"))
      .toHaveAccessibleDescription(/redirectUris is required/);
    await expect.element(field("Redirect URIs")).toHaveValue("\n \n");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1438
  test("shows a stale version in the error summary", async () => {
    serveDetail();
    const detail =
      "The data recipient changed after the version in the request. The Register changed nothing.";
    serveChange({
      status: 409,
      body: { type: "about:blank", title: "Conflict", status: 409, detail },
    });
    renderApp(formURL);

    await save().click();

    await expect.element(page.getByRole("alert").getByText(detail)).toBeVisible();
  });

  test("says that the software product does not exist when the data recipient has no software product with the ID", async () => {
    serveDetail();
    renderApp(`/data-recipients/${id}/software-products/unknown/change`);

    await expect
      .element(page.getByText("The data recipient has no software product with this ID."))
      .toBeVisible();
    await expect.element(page.getByRole("form")).not.toBeInTheDocument();
  });
});
