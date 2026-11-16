import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientBrand, DataRecipientDetail, SoftwareProduct } from "../api/client";
import { definitionOf, http, recipient, renderApp, worker, renderDetail } from "../test/app";

const id = "c7b88455-4a14-479e-b53e-fbc4000716b3";
const thumbprint = "fj5bZBu5UoSrAua59pRyfSczfxvH2Jo9mn-XiLUWZ_M";
const longRedirectUri = `https://web.mock-budget.example/${"callback/".repeat(20)}`;

const web: SoftwareProduct = {
  softwareProductId: "66b83c54-e4f4-434d-aa9b-789f64358022",
  softwareProductName: "Mock Budget Web",
  softwareProductDescription: "A mock software product",
  logoUri: "https://web.mock-budget.example/logo.png",
  status: "ACTIVE",
  clientUri: "https://web.mock-budget.example",
  sectorIdentifierUri: "https://web.mock-budget.example/sector",
  redirectUris: ["https://web.mock-budget.example/callback", longRedirectUri],
  tosUri: "https://web.mock-budget.example/terms",
  policyUri: "https://web.mock-budget.example/privacy",
  jwksUri: "https://web.mock-budget.example/jwks",
  revocationUri: "https://web.mock-budget.example/revoke",
  recipientBaseUri: "https://cdr.web.mock-budget.example",
  scopes: ["openid", "bank:accounts.basic:read"],
  clientCertificateThumbprints: [thumbprint],
  actions: [],
  clientCertificateActions: [],
};

const mobile: SoftwareProduct = {
  softwareProductId: "c9c2cde8-e247-4565-b341-376b029738d9",
  softwareProductName: "Mock Budget Mobile",
  softwareProductDescription: "A mock software product",
  logoUri: "https://mobile.mock-budget.example/logo.png",
  status: "INACTIVE",
  clientUri: "https://mobile.mock-budget.example",
  redirectUris: ["https://mobile.mock-budget.example/callback"],
  jwksUri: "https://mobile.mock-budget.example/jwks",
  revocationUri: "https://mobile.mock-budget.example/revoke",
  recipientBaseUri: "https://cdr.mobile.mock-budget.example",
  scopes: ["openid"],
  clientCertificateThumbprints: [],
  actions: [],
  clientCertificateActions: [],
};

const detail: DataRecipientDetail = {
  legalEntityId: id,
  legalEntityName: "Mock Budgeting Pty Ltd",
  accreditationNumber: "ADRBNK000001",
  accreditationLevel: "UNRESTRICTED",
  logoUri: "https://mock-budgeting.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-23T02:05:07.123456Z",
  version: 1,
  actions: [],
  dataRecipientBrands: [
    {
      dataRecipientBrandId: "e25806f4-9384-41f6-a65d-2b12e2529e3b",
      brandName: "Mock Budget",
      logoUri: "https://mock-budget.example/logo.png",
      status: "ACTIVE",
      actions: [],
      softwareProducts: [web, mobile],
    },
    {
      dataRecipientBrandId: "0b1e4b5a-8f55-4a8e-9f0e-3c8a4d2b1f00",
      brandName: "Mock Saver",
      logoUri: "https://mock-saver.example/logo.png",
      status: "REMOVED",
      actions: [],
      softwareProducts: [],
    },
  ],
};

function serve(body: DataRecipientDetail) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ params, response }) => {
      requests.push(params.legalEntityId);
      return response(200).json(body);
    }),
  );
  return requests;
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const valueOf = (regionName: string, term: string) =>
  definitionOf(region(regionName).element(), term).textContent;

describe("the data recipient page", () => {
  // Verifies: REG-1007, REG-1008, REG-1009
  test("shows the legal entity, each brand and each software product", async () => {
    const requests = serve(detail);

    await renderDetail(`/data-recipients/${id}`);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Mock Budgeting Pty Ltd" }))
      .toBeVisible();
    expect(requests).toEqual([id]);

    const entity = "Legal entity";
    expect(valueOf(entity, "Legal entity ID")).toContain(id);
    expect(valueOf(entity, "Accreditation number")).toContain("ADRBNK000001");
    expect(valueOf(entity, "Accreditation level")).toContain("UNRESTRICTED");
    expect(valueOf(entity, "Status")).toContain("ACTIVE");
    expect(valueOf(entity, "Logo URI")).toContain("https://mock-budgeting.example/logo.png");
    // The tests run in Australia/Perth (vitest.config.ts).
    const changed = definitionOf(region(entity).element(), "Last changed");
    expect(changed.textContent).toContain("23 October 2026, 10:05 am UTC+8");
    expect(changed.querySelector("time")?.dateTime).toBe("2026-10-23T02:05:07.123Z");

    const budget = "Mock Budget";
    await expect
      .element(region(budget).getByRole("heading", { level: 3 }))
      .toHaveTextContent("Mock Budget");
    expect(valueOf(budget, "Brand ID")).toContain("e25806f4-9384-41f6-a65d-2b12e2529e3b");

    const product = "Mock Budget Web";
    await expect
      .element(region(product).getByRole("heading", { level: 5 }))
      .toHaveTextContent("Mock Budget Web");
    expect(valueOf(product, "Software product ID")).toContain(web.softwareProductId);
    expect(valueOf(product, "Sector identifier URI")).toContain(
      "https://web.mock-budget.example/sector",
    );
    expect(
      definitionOf(region(product).element(), "Redirect URIs").querySelectorAll("li"),
    ).toHaveLength(2);
    expect(valueOf(product, "Scopes")).toContain("bank:accounts.basic:read");
    expect(valueOf(product, "Client certificate thumbprints")).toContain(thumbprint);
    await expect
      .element(
        page.getByRole("button", { name: "Copy client certificate thumbprint of Mock Budget Web" }),
      )
      .toBeVisible();

    await expect.element(region("Mock Saver").getByText("REMOVED")).toBeVisible();
    await expect
      .element(region("Mock Saver").getByText("This brand has no software products"))
      .toBeVisible();

    await expectNoAxeViolations(document.body);
  });

  test("each link is at the heading of its record, and the buttons are after the details of the record and before its child records", async () => {
    const deactivate: DataRecipientBrand["actions"][number] = {
      action: "deactivate",
      status: "INACTIVE",
      changes: [],
      irreversible: false,
    };
    const [budget, saver] = detail.dataRecipientBrands;
    serve({
      ...detail,
      dataRecipientBrands: [{ ...budget!, actions: [deactivate] }, saver!],
    });

    await renderDetail(`/data-recipients/${id}`);

    const brand = region("Mock Budget");
    const heading = brand.getByRole("heading", { level: 3, name: "Mock Budget" });
    const change = brand.getByRole("link", { name: "Change brand Mock Budget" });
    await expect.element(change).toBeVisible();
    expect(change.element().getBoundingClientRect().top).toBeLessThan(
      heading.element().getBoundingClientRect().bottom,
    );

    const button = brand.getByRole("button", { name: "Deactivate" }).element();
    const details = brand.element().querySelector("dl")!;
    const products = brand.getByRole("heading", { level: 4, name: "Software products" }).element();
    expect(details.compareDocumentPosition(button) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(
      button.compareDocumentPosition(products) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    await expect
      .element(
        region("Software products")
          .first()
          .getByRole("link", { name: "Add software product to Mock Budget" }),
      )
      .toBeVisible();
  });

  test("shows each identifier, URI and thumbprint in full, in the monospace font, with a copy button", async () => {
    serve(detail);

    await renderDetail(`/data-recipients/${id}`);

    const product = region("Mock Budget Web");
    await expect.element(product.getByText(longRedirectUri)).toBeVisible();
    for (const [value, label] of [
      [web.softwareProductId, "software product ID of Mock Budget Web"],
      [web.jwksUri, "JWKS URI of Mock Budget Web"],
      [longRedirectUri, "redirect URI of Mock Budget Web"],
      [thumbprint, "client certificate thumbprint of Mock Budget Web"],
      [detail.logoUri, "logo URI of Mock Budgeting Pty Ltd"],
    ] as const) {
      const code = page.getByText(value, { exact: true }).element();
      expect(code.tagName).toBe("CODE");
      expect(getComputedStyle(code).fontFamily).toMatch(/mono/i);
      expect(page.getByRole("button", { name: `Copy ${label}` }).elements()).not.toHaveLength(0);
    }
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(window.innerWidth);
  });

  // Verifies: REG-1058
  test("shows the logos as URIs and loads no image", async () => {
    serve(detail);

    await renderDetail(`/data-recipients/${id}`);

    await expect.element(page.getByRole("heading", { level: 1 })).toBeVisible();
    expect(document.querySelectorAll("img")).toHaveLength(0);
  });

  test("shows Not set for an optional field with no value and for an empty list", async () => {
    serve(detail);

    await renderDetail(`/data-recipients/${id}`);

    const product = "Mock Budget Mobile";
    await expect.element(region(product)).toBeVisible();
    const notSet = [
      "Sector identifier URI",
      "Terms of service URI",
      "Policy URI",
      "Client certificate thumbprints",
    ];
    expect(Object.fromEntries(notSet.map((term) => [term, valueOf(product, term)]))).toEqual(
      Object.fromEntries(notSet.map((term) => [term, "Not set"])),
    );
  });

  test("shows the title and the detail of the problem for an unknown legal entity", async () => {
    worker.use(
      http.get("/api/data-recipients/{legalEntityId}", ({ response }) =>
        response(404).json({
          type: "about:blank",
          title: "Not Found",
          status: 404,
          detail: "The Register has no data recipient with this legal entity ID.",
        }),
      ),
    );

    renderApp("/data-recipients/unknown-id");

    const alert = page.getByRole("alert");
    await expect.element(alert.getByText("Not Found")).toBeVisible();
    await expect
      .element(alert.getByText("The Register has no data recipient with this legal entity ID."))
      .toBeVisible();
    await expect
      .element(page.getByRole("heading", { level: 1, name: "Data recipient" }))
      .toBeVisible();
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1006
  test("a row of the list links to the detail, and the navigation item stays current", async () => {
    serve(detail);
    worker.use(
      http.get("/api/data-recipients", ({ response }) =>
        response(200).json({
          data: [{ ...recipient("Mock Budgeting Pty Ltd"), legalEntityId: id }],
          totalRecords: 1,
        }),
      ),
    );
    const router = renderApp("/data-recipients");

    await userEvent.click(page.getByRole("link", { name: "Mock Budgeting Pty Ltd" }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Mock Budgeting Pty Ltd" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    await expect
      .element(page.getByRole("link", { name: "Data recipients" }))
      .toHaveAttribute("aria-current", "page");
  });
});
