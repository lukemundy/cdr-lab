import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataHolderDetail } from "../api/client";
import { definitionOf, holder, http, renderApp, worker, renderDetail } from "../test/app";

const id = "8f2b4c8e-0a43-4d4f-9b2f-6a7d3e1c2b10";

const detail: DataHolderDetail = {
  legalEntityId: id,
  legalEntityName: "Mock Holdings Ltd",
  logoUri: "https://mock-holdings.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-22T23:59:59Z",
  version: 1,
  actions: [],
  dataHolderBrands: [
    {
      dataHolderBrandId: "d7385786-d8b1-4d4b-bb80-b21616df1a4d",
      interimId: "1b2c3d4e-0000-4000-8000-000000000001",
      brandName: "Mock Bank",
      brandGroup: "Mock Group",
      industries: ["banking", "non-bank-lending"],
      logoUri: "https://mock-bank.example/logo.png",
      status: "ACTIVE",
      actions: [],
      serverCertificateActions: [],
      endpointDetail: {
        version: "1",
        publicBaseUri: "https://api.mock-bank.example",
        productBaseUri: "https://api.mock-bank.example/products",
        resourceBaseUri: "https://mtls.mock-bank.example",
        infosecBaseUri: "https://id.mock-bank.example",
        extensionBaseUri: "https://ext.mock-bank.example",
        websiteUri: "https://mock-bank.example",
      },
      jwksEndpoint: "https://id.mock-bank.example/jwks",
    },
    {
      interimId: "d9530e60-cdc4-408c-a6f4-aba39e4158b6",
      brandName: "Mock Mutual",
      industries: ["banking"],
      logoUri: "https://mock-mutual.example/logo.png",
      status: "INACTIVE",
      actions: [],
      serverCertificateActions: [],
      endpointDetail: { publicBaseUri: "https://prd.mock-mutual.example" },
    },
  ],
};

function serve(body: DataHolderDetail) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/data-holders/{legalEntityId}", ({ params, response }) => {
      requests.push(params.legalEntityId);
      return response(200).json(body);
    }),
  );
  return requests;
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const valueOf = (regionName: string, term: string) =>
  definitionOf(region(regionName).element(), term).textContent;

describe("the data holder page", () => {
  test("shows the legal entity and each field of each brand", async () => {
    const requests = serve(detail);

    await renderDetail(`/data-holders/${id}`);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Mock Holdings Ltd" }))
      .toBeVisible();
    expect(requests).toEqual([id]);

    expect(valueOf("Legal entity", "Legal entity ID")).toContain(id);
    expect(valueOf("Legal entity", "Logo URI")).toContain(detail.logoUri);
    expect(valueOf("Legal entity", "Status")).toBe("ACTIVE");
    // The tests run in Australia/Perth (vitest.config.ts).
    const changed = definitionOf(region("Legal entity").element(), "Last changed");
    expect(changed.textContent).toContain("23 October 2026, 7:59 am UTC+8");
    expect(changed.querySelector("time")?.dateTime).toBe("2026-10-22T23:59:59.000Z");

    await expect
      .element(region("Mock Bank").getByRole("heading", { level: 3 }))
      .toHaveTextContent("Mock Bank");
    const bank = [
      ["Brand ID", "d7385786-d8b1-4d4b-bb80-b21616df1a4d"],
      ["Interim ID", "1b2c3d4e-0000-4000-8000-000000000001"],
      ["Brand group", "Mock Group"],
      ["Industries", "banking, non-bank-lending"],
      ["Logo URI", "https://mock-bank.example/logo.png"],
      ["Endpoint version", "1"],
      ["Public base URI", "https://api.mock-bank.example"],
      ["Product base URI", "https://api.mock-bank.example/products"],
      ["Resource base URI", "https://mtls.mock-bank.example"],
      ["Infosec base URI", "https://id.mock-bank.example"],
      ["Extension base URI", "https://ext.mock-bank.example"],
      ["Website URI", "https://mock-bank.example"],
      ["JWKS endpoint", "https://id.mock-bank.example/jwks"],
    ] as const;
    expect(Object.fromEntries(bank.map(([term]) => [term, valueOf("Mock Bank", term)]))).toEqual(
      Object.fromEntries(bank),
    );
    await expect
      .element(page.getByRole("button", { name: "Copy resource base URI of Mock Bank" }))
      .toBeVisible();
    expect(document.querySelectorAll("img")).toHaveLength(0);

    await expectNoAxeViolations(document.body);
  });

  test("a brand with only an interim ID shows Not set for the brand ID and the secure endpoints", async () => {
    serve(detail);

    await renderDetail(`/data-holders/${id}`);

    await expect.element(region("Mock Mutual")).toBeVisible();
    expect(valueOf("Mock Mutual", "Interim ID")).toContain("d9530e60-cdc4-408c-a6f4-aba39e4158b6");
    expect(valueOf("Mock Mutual", "Public base URI")).toContain("https://prd.mock-mutual.example");
    // The status of a brand is in its title bar, so it shows also when the card is collapsed.
    await expect
      .element(region("Mock Mutual").getByText("INACTIVE", { exact: true }))
      .toBeVisible();
    const notSet = [
      "Brand ID",
      "Brand group",
      "Endpoint version",
      "Product base URI",
      "Resource base URI",
      "Infosec base URI",
      "Extension base URI",
      "Website URI",
      "JWKS endpoint",
    ];
    expect(Object.fromEntries(notSet.map((term) => [term, valueOf("Mock Mutual", term)]))).toEqual(
      Object.fromEntries(notSet.map((term) => [term, "Not set"])),
    );
  });

  test("a data holder with no brands shows an empty state", async () => {
    serve({ ...detail, dataHolderBrands: [] });

    await renderDetail(`/data-holders/${id}`);

    await expect
      .element(region("Brands").getByText("This data holder has no brands"))
      .toBeVisible();
  });

  test("shows the title and the detail of the problem for an unknown legal entity", async () => {
    worker.use(
      http.get("/api/data-holders/{legalEntityId}", ({ response }) =>
        response(404).json({
          type: "about:blank",
          title: "Not Found",
          status: 404,
          detail: "The Register has no data holder with this legal entity ID.",
        }),
      ),
    );

    renderApp("/data-holders/unknown-id");

    const alert = page.getByRole("alert");
    await expect.element(alert.getByText("Not Found")).toBeVisible();
    await expect
      .element(alert.getByText("The Register has no data holder with this legal entity ID."))
      .toBeVisible();
    await expect
      .element(page.getByRole("heading", { level: 1, name: "Data holder" }))
      .toBeVisible();
    await expectNoAxeViolations(document.body);
  });

  test("a row of the list links to the detail, and the navigation item stays current", async () => {
    serve(detail);
    worker.use(
      http.get("/api/data-holders", ({ response }) =>
        response(200).json({
          data: [{ ...holder("Mock Holdings Ltd"), legalEntityId: id }],
          totalRecords: 1,
        }),
      ),
    );
    const router = renderApp("/data-holders");

    await userEvent.click(page.getByRole("link", { name: "Mock Holdings Ltd" }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Mock Holdings Ltd" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
    await expect
      .element(page.getByRole("link", { name: "Data holders" }))
      .toHaveAttribute("aria-current", "page");
  });
});
