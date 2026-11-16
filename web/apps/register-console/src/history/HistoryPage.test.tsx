import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataHolderHistoryEntry, DataRecipientHistoryEntry } from "../api/client";
import { definitionOf, http, renderApp, worker } from "../test/app";

const id = "c7b88455-4a14-479e-b53e-fbc4000716b3";
const holderId = "a81975c6-1ac9-4fe4-88fc-68f50efd27cb";
const productId = "66b83c54-e4f4-434d-aa9b-789f64358022";
const brandId = "d7385786-d8b1-4d4b-bb80-b21616df1a4d";
const thumbprint = "fj5bZBu5UoSrAua59pRyfSczfxvH2Jo9mn-XiLUWZ_M";

const suspension: DataRecipientHistoryEntry[] = [
  {
    streamVersion: 6,
    eventType: "DataRecipientBrandDeactivated",
    recordedAt: "2026-10-23T05:00:00Z",
    brandId: "e25806f4-9384-41f6-a65d-2b12e2529e3b",
    brandName: "Mock Budget",
  },
  {
    streamVersion: 5,
    eventType: "ClientCertificateRemoved",
    recordedAt: "2026-10-23T04:00:00Z",
    softwareProductId: productId,
    softwareProductName: "Mock Budget Web",
    thumbprint,
  },
  {
    streamVersion: 4,
    eventType: "ClientCertificateRegistered",
    recordedAt: "2026-10-23T03:00:00.5Z",
    softwareProductId: productId,
    softwareProductName: "Mock Budget Web",
    thumbprint,
  },
  {
    streamVersion: 3,
    eventType: "SoftwareProductDeactivated",
    recordedAt: "2026-10-23T02:05:07.123456Z",
    softwareProductId: productId,
    softwareProductName: "Mock Budget Web",
  },
  {
    streamVersion: 2,
    eventType: "AccreditationSuspended",
    recordedAt: "2026-10-23T02:05:07.123456Z",
  },
  { streamVersion: 1, eventType: "DataRecipientImported", recordedAt: "2026-10-22T23:00:00Z" },
];

function serveRecipient(entries: DataRecipientHistoryEntry[], totalRecords = entries.length) {
  const queries: string[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) =>
      response(200).json({
        legalEntityId: id,
        legalEntityName: "Mock Budgeting Pty Ltd",
        accreditationNumber: "ADRBNK000001",
        accreditationLevel: "UNRESTRICTED",
        logoUri: "https://mock-budgeting.example/logo.png",
        status: "SUSPENDED",
        lastUpdated: "2026-10-23T03:00:00.5Z",
        version: 3,
        actions: [],
        dataRecipientBrands: [],
      }),
    ),
    http.get("/api/data-recipients/{legalEntityId}/history", ({ request, response }) => {
      queries.push(new URL(request.url).search);
      return response(200).json({ data: entries, totalRecords });
    }),
  );
  return queries;
}

function serveHolder(entries: DataHolderHistoryEntry[]) {
  worker.use(
    http.get("/api/data-holders/{legalEntityId}", ({ response }) =>
      response(200).json({
        legalEntityId: holderId,
        legalEntityName: "Mock Holdings Ltd",
        logoUri: "https://mock-holdings.example/logo.png",
        status: "ACTIVE",
        lastUpdated: "2026-10-22T23:00:00Z",
        version: 1,
        actions: [],
        dataHolderBrands: [],
      }),
    ),
    http.get("/api/data-holders/{legalEntityId}/history", ({ response }) =>
      response(200).json({ data: entries, totalRecords: entries.length }),
    ),
  );
}

const terms = (entry: Element) => [...entry.querySelectorAll("dt")].map((dt) => dt.textContent);
const timeline = () => page.getByRole("list", { name: "History" });
const entries = () => [...timeline().element().querySelectorAll(":scope > li")];

describe("the history page", () => {
  // Verifies: REG-1014, REG-1015
  test("shows each event newest first with its time, text, event type and only the fields of the table", async () => {
    const queries = serveRecipient(suspension);

    renderApp(`/data-recipients/${id}/history`);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "History of Mock Budgeting Pty Ltd" }))
      .toBeVisible();
    // The detail and the history are two requests, so the heading can show before the timeline.
    await expect.element(timeline().getByText("Loaded into the Register")).toBeVisible();
    expect(queries).toEqual(["?pageSize=25"]);
    expect(entries().map((entry) => entry.querySelector("p")?.textContent)).toEqual([
      "Brand deactivated",
      "Client certificate removed",
      "Client certificate registered",
      "Software product deactivated",
      "Accreditation suspended",
      "Loaded into the Register",
    ]);

    const [brandDeactivation, removal, certificate, deactivation, suspended] = entries();
    if (!brandDeactivation || !removal || !certificate || !deactivation || !suspended) {
      throw new Error("no entries");
    }
    expect(terms(brandDeactivation)).toEqual(["Event type", "Brand", "Brand ID"]);
    expect(definitionOf(brandDeactivation, "Brand").textContent).toBe("Mock Budget");
    expect(definitionOf(removal, "Thumbprint").textContent).toContain(thumbprint);
    expect(deactivation.querySelector("time")?.textContent).toContain(
      "23 October 2026, 10:05:07 am UTC+8",
    );
    expect(definitionOf(deactivation, "Event type").textContent).toContain(
      "SoftwareProductDeactivated",
    );
    expect(definitionOf(deactivation, "Software product").textContent).toBe("Mock Budget Web");
    expect(definitionOf(deactivation, "Software product ID").textContent).toContain(productId);
    expect(definitionOf(certificate, "Thumbprint").textContent).toContain(thumbprint);

    expect(terms(certificate)).toEqual([
      "Event type",
      "Software product",
      "Software product ID",
      "Thumbprint",
    ]);
    expect(terms(deactivation)).toEqual(["Event type", "Software product", "Software product ID"]);
    expect(terms(suspended)).toEqual(["Event type"]);
    await expect
      .element(page.getByRole("button", { name: "Copy thumbprint of event 4" }))
      .toBeVisible();

    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1038
  test("shows the issue of a client certificate with the software product and the thumbprint", async () => {
    serveRecipient([
      {
        streamVersion: 2,
        eventType: "ClientCertificateIssued",
        recordedAt: "2026-10-27T03:00:00Z",
        softwareProductId: productId,
        softwareProductName: "Mock Budget Web",
        thumbprint,
      },
      { streamVersion: 1, eventType: "DataRecipientImported", recordedAt: "2026-10-22T23:00:00Z" },
    ]);

    renderApp(`/data-recipients/${id}/history`);

    await expect.element(timeline().getByText("Client certificate issued")).toBeVisible();
    const [issue] = entries();
    if (!issue) {
      throw new Error("no entries");
    }
    expect(terms(issue)).toEqual([
      "Event type",
      "Software product",
      "Software product ID",
      "Thumbprint",
    ]);
    expect(definitionOf(issue, "Software product").textContent).toBe("Mock Budget Web");
    expect(definitionOf(issue, "Thumbprint").textContent).toContain(thumbprint);
  });

  // Verifies: REG-1055
  test("gives the page of the URL to the API and keeps the next page in the URL", async () => {
    const queries = serveRecipient(suspension.slice(0, 2), 60);

    const router = renderApp(`/data-recipients/${id}/history?page=2`);

    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();
    expect(queries).toEqual(["?page=2&pageSize=25"]);

    await userEvent.click(page.getByRole("button", { name: "Next page" }));

    await expect.poll(() => router.state.location.searchStr).toBe("?page=3");
    await expect.poll(() => queries).toContain("?page=3&pageSize=25");
  });

  test("removes a page that is not valid from the URL", async () => {
    const queries = serveRecipient(suspension);

    renderApp(`/data-recipients/${id}/history?page=zero`);

    await expect.element(timeline()).toBeVisible();
    expect(queries).toEqual(["?pageSize=25"]);
  });

  // Verifies: REG-1013
  test("the detail links to the history, and the history links back to the detail", async () => {
    serveRecipient(suspension);
    const router = renderApp(`/data-recipients/${id}`);

    await userEvent.click(page.getByRole("link", { name: "History", exact: true }));

    await expect.element(timeline()).toBeVisible();
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}/history`);
    await expect
      .element(page.getByRole("link", { name: "Data recipients" }))
      .toHaveAttribute("aria-current", "page");

    await userEvent.click(page.getByRole("link", { name: "Detail", exact: true }));

    await expect.poll(() => router.state.location.pathname).toBe(`/data-recipients/${id}`);
  });

  test("shows the history of a data holder, with the brand of a brand event", async () => {
    const holderEntries: DataHolderHistoryEntry[] = [
      {
        streamVersion: 3,
        eventType: "DataHolderBrandRemoved",
        recordedAt: "2026-10-23T01:00:00Z",
        brandId,
        brandName: "Mock Bank",
      },
      {
        streamVersion: 2,
        eventType: "DataHolderLegalEntityRemoved",
        recordedAt: "2026-10-23T01:00:00Z",
      },
      { streamVersion: 1, eventType: "DataHolderImported", recordedAt: "2026-10-22T23:00:00Z" },
    ];
    serveHolder(holderEntries);

    renderApp(`/data-holders/${holderId}/history`);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "History of Mock Holdings Ltd" }))
      .toBeVisible();
    await expect.element(timeline().getByText("Loaded into the Register")).toBeVisible();
    expect(entries().map((entry) => entry.querySelector("p")?.textContent)).toEqual([
      "Brand removed",
      "Data holder removed",
      "Loaded into the Register",
    ]);

    const [brandRemoval, removal] = entries();
    if (!brandRemoval || !removal) throw new Error("no entries");
    expect(terms(brandRemoval)).toEqual(["Event type", "Brand", "Brand ID"]);
    expect(definitionOf(brandRemoval, "Brand").textContent).toBe("Mock Bank");
    expect(definitionOf(brandRemoval, "Brand ID").textContent).toContain(brandId);
    expect(terms(removal)).toEqual(["Event type"]);
    await expect
      .element(page.getByRole("link", { name: "Data holders" }))
      .toHaveAttribute("aria-current", "page");
  });

  // Verifies: REG-1040
  test("shows the issue of a server certificate with the brand and the thumbprint", async () => {
    serveHolder([
      {
        streamVersion: 2,
        eventType: "ServerCertificateIssued",
        recordedAt: "2026-10-27T03:00:00Z",
        brandId,
        brandName: "Mock Bank",
        thumbprint,
      },
      { streamVersion: 1, eventType: "DataHolderImported", recordedAt: "2026-10-22T23:00:00Z" },
    ]);

    renderApp(`/data-holders/${holderId}/history`);

    await expect.element(timeline().getByText("Server certificate issued")).toBeVisible();
    const [issue] = entries();
    if (!issue) throw new Error("no entries");
    expect(terms(issue)).toEqual(["Event type", "Brand", "Brand ID", "Thumbprint"]);
    expect(definitionOf(issue, "Brand").textContent).toBe("Mock Bank");
    expect(definitionOf(issue, "Brand ID").textContent).toContain(brandId);
    expect(definitionOf(issue, "Thumbprint").textContent).toContain(thumbprint);
  });

  test("shows the title and the detail of the problem for an unknown legal entity", async () => {
    const notFound = {
      type: "about:blank" as const,
      title: "Not Found",
      status: 404,
      detail: "The Register has no data recipient with this legal entity ID.",
    };
    worker.use(
      http.get("/api/data-recipients/{legalEntityId}", ({ response }) =>
        response(404).json(notFound),
      ),
      http.get("/api/data-recipients/{legalEntityId}/history", ({ response }) =>
        response(404).json(notFound),
      ),
    );

    renderApp("/data-recipients/unknown-id/history");

    const alert = page.getByRole("alert");
    await expect.element(alert.getByText("Not Found")).toBeVisible();
    await expect
      .element(alert.getByText("The Register has no data recipient with this legal entity ID."))
      .toBeVisible();
    await expect.element(page.getByRole("heading", { level: 1, name: "History" })).toBeVisible();
    await expectNoAxeViolations(document.body);
  });
});
