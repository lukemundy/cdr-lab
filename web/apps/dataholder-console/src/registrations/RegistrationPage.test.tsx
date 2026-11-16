import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { ClientArrangementList, Registration } from "../api/client";
import {
  clientArrangement,
  customerDetail,
  definitionOf,
  http,
  registration,
  renderApp,
  worker,
} from "../test/app";

const clientId = "5f0c2a8e-1b7d-4c3e-9a6f-2d8e4b1c7a90";

function serve(body: Registration) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/registrations/{clientId}", ({ params, response }) => {
      requests.push(params.clientId);
      return response(200).json(body);
    }),
  );
  serveArrangements({ data: [], totalRecords: 0 });
  return requests;
}

// Each request of the arrangements of the client ID, as its search. A test can end before the request of its
// arrangements, so a test that counts the requests uses its own client ID.
function serveArrangements(body: ClientArrangementList, of = clientId) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/registrations/{clientId}/arrangements", ({ params, request, response }) => {
      if (params.clientId === of) requests.push(new URL(request.url).search);
      return response(200).json(body);
    }),
  );
  return requests;
}

const table = () => page.getByRole("table", { name: "Arrangements" });

const region = (name: string) => page.getByRole("region", { name, exact: true });
const valueOf = (regionName: string, term: string) =>
  definitionOf(region(regionName).element(), term).textContent;

describe("the client registration page", () => {
  // Verifies: DH-1431, DH-1432
  test("shows the metadata of the registration, with each URI as text with a copy button", async () => {
    const requests = serve(
      registration(
        clientId,
        { lastUpdated: "2026-10-16T01:00:00Z" },
        {
          redirect_uris: ["https://recipient.example/cb", "https://recipient.example/cb2"],
          legal_entity_name: "Recipient Pty Ltd",
        },
      ),
    );

    renderApp(`/registrations/${clientId}`);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Recipient Web" }))
      .toBeVisible();
    expect(requests).toEqual([clientId]);

    expect(valueOf("Registration", "Client ID")).toContain(clientId);
    expect(valueOf("Registration", "Software ID")).toContain("software-1");
    expect(valueOf("Registration", "Register status")).toBe("ACTIVE");
    // The tests run in Australia/Perth (vitest.config.ts).
    expect(valueOf("Registration", "Registered")).toContain("15 October 2026, 1:30:00 pm UTC+8");
    expect(valueOf("Registration", "Last changed")).toContain("16 October 2026, 9:00:00 am UTC+8");

    const redirects = definitionOf(region("Metadata").element(), "Redirect URIs");
    expect([...redirects.querySelectorAll("li")].map((item) => item.textContent)).toEqual([
      "https://recipient.example/cb",
      "https://recipient.example/cb2",
    ]);
    expect(valueOf("Metadata", "Recipient base URI")).toBe("https://recipient.example/cdr");
    expect(valueOf("Metadata", "JWKS URI")).toBe("https://recipient.example/jwks");
    expect(valueOf("Metadata", "Scopes")).toBe("openidbank:accounts.basic:read");
    expect(valueOf("Metadata", "Legal entity name")).toBe("Recipient Pty Ltd");
    expect(valueOf("Metadata", "Legal entity ID")).toBe("Not set");
    expect(valueOf("Metadata", "Brand ID")).toBe("brand-1");
    await expect
      .element(region("Metadata").getByRole("button", { name: "Copy brand ID" }))
      .toBeVisible();
    await expect
      .element(region("Metadata").getByRole("button", { name: "Copy recipient base URI" }))
      .toBeVisible();
    expect(region("Metadata").element().querySelectorAll("a")).toHaveLength(0);
    await expectNoAxeViolations(document.body);
  });

  test("a legal entity ID shows with a copy button", async () => {
    serve(registration(clientId, {}, { legal_entity_id: "legal-entity-1" }));

    renderApp(`/registrations/${clientId}`);

    await expect
      .element(region("Metadata").getByRole("button", { name: "Copy legal entity ID" }))
      .toBeVisible();
    expect(valueOf("Metadata", "Legal entity ID")).toBe("legal-entity-1");
  });

  test("a software product with no Register status shows that the status is not known", async () => {
    const noStatus: Registration = registration(clientId);
    delete noStatus.softwareProductStatus;
    serve(noStatus);

    renderApp(`/registrations/${clientId}`);

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Recipient Web" }))
      .toBeVisible();
    expect(valueOf("Registration", "Register status")).toBe("Not known");
  });

  // Verifies: DH-1433, DH-1434
  test("shows the arrangements of the client in a table, with the customer, the status and the times", async () => {
    const id = "9b3e6f10-4c2a-4d8e-b1f7-0a5c3e9d2b64";
    serve(registration(id));
    const requests = serveArrangements(
      {
        data: [
          clientArrangement("arrangement-1", {
            customerNumber: "12345678",
            firstName: "Alex",
            lastName: "Example",
          }),
          clientArrangement("arrangement-2", {
            customerNumber: "87654321",
            lastName: "Sample",
            status: "ENDED",
            sharingEndsAt: "2026-10-16T01:00:00Z",
          }),
        ],
        totalRecords: 30,
      },
      id,
    );
    worker.use(
      http.get("/api/customers/{customerNumber}", ({ params, response }) =>
        response(200).json(customerDetail(params.customerNumber)),
      ),
    );
    const router = renderApp(`/registrations/${id}`);

    const ended = table().getByRole("row").nth(2);
    // Until the response, the first row is the row of the loading state, so wait for the data first.
    await expect.element(ended.getByRole("cell", { name: "ENDED", exact: true })).toBeVisible();
    const current = table().getByRole("row").nth(1);
    await expect.element(current.getByRole("cell", { name: "CURRENT", exact: true })).toBeVisible();
    await expect.element(current.getByRole("link", { name: "Alex Example" })).toBeVisible();
    expect(ended.getByRole("cell").first().element().textContent).toContain("arrangement-2");
    expect(ended.getByRole("cell").nth(1).element().textContent).toContain("87654321");
    // The tests run in Australia/Perth (vitest.config.ts).
    expect(ended.getByRole("cell").nth(3).element().textContent).toContain(
      "15 October 2026, 1:30 pm UTC+8",
    );
    expect(ended.getByRole("cell").nth(4).element().textContent).toContain(
      "16 October 2026, 9:00 am UTC+8",
    );
    await expect
      .element(ended.getByRole("button", { name: "Copy arrangement ID of Sample" }))
      .toBeVisible();
    await expect
      .element(ended.getByRole("button", { name: "Copy customer number of Sample" }))
      .toBeVisible();
    await expect.element(page.getByText("30 records")).toBeVisible();
    await expect.element(page.getByText("Page 1 of 2")).toBeVisible();
    expect(requests).toEqual(["?pageSize=25"]);
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(window.innerWidth);
    await expectNoAxeViolations(document.body);

    await userEvent.click(ended.getByRole("link", { name: "Sample" }));

    await expect.poll(() => router.state.location.pathname).toBe("/customers/87654321");
  });

  test("gives the page of the URL to the API and keeps the next page in the URL", async () => {
    const id = "c2d8a4f6-7e1b-4a3c-9f05-6b2e8d1a4c37";
    serve(registration(id));
    const requests = serveArrangements(
      { data: [clientArrangement("arrangement-1")], totalRecords: 60 },
      id,
    );

    const router = renderApp(`/registrations/${id}?page=2`);

    await expect.element(page.getByText("Page 2 of 3")).toBeVisible();
    expect(requests).toEqual(["?page=2&pageSize=25"]);

    await userEvent.click(page.getByRole("button", { name: "Next page" }));

    await expect.poll(() => router.state.location.searchStr).toBe("?page=3");
    await expect.poll(() => requests).toContain("?page=3&pageSize=25");
  });

  test("a client with no arrangements shows that it has none", async () => {
    serve(registration(clientId));

    renderApp(`/registrations/${clientId}`);

    await expect.element(table().getByText("No arrangements on this page")).toBeVisible();
  });

  test("shows the title and the detail of the problem for an unknown registration", async () => {
    worker.use(
      http.get("/api/registrations/{clientId}", ({ response }) =>
        response(404).json({
          type: "about:blank",
          title: "Not Found",
          status: 404,
          detail: "The data holder has no registration with this client ID.",
        }),
      ),
    );

    renderApp(`/registrations/${clientId}`);

    const alert = page.getByRole("alert");
    await expect
      .element(alert.getByText("The data holder has no registration with this client ID."))
      .toBeVisible();
    await expect
      .element(page.getByRole("heading", { level: 1, name: "Client registration" }))
      .toBeVisible();
  });
});
