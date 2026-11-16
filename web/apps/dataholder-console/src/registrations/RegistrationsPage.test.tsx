import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { Registration } from "../api/client";
import { http, registration, renderApp, worker } from "../test/app";

const clientId = "5f0c2a8e-1b7d-4c3e-9a6f-2d8e4b1c7a90";
const otherId = "a3d9e6b1-7c4f-4e2a-8b5d-0f1e9c3a6b72";

// Each request of the page, as the search of its URL.
function serve(data: Registration[]) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/registrations", ({ request, response }) => {
      requests.push(new URL(request.url).search);
      return response(200).json({ data });
    }),
  );
  return requests;
}

const table = () => page.getByRole("table", { name: "Client registrations" });

describe("the client registrations page", () => {
  // Verifies: DH-1428
  test("shows each registration with its identifiers, its times and the Register status of its software product", async () => {
    const noStatus: Registration = registration(
      otherId,
      { lastUpdated: "2026-10-16T01:00:00Z" },
      { client_name: "Recipient Mobile", software_id: "software-2" },
    );
    delete noStatus.softwareProductStatus;
    serve([registration(clientId, { arrangementCount: 7, currentArrangementCount: 2 }), noStatus]);

    renderApp("/registrations");

    const first = table().getByRole("row").nth(1);
    await expect.element(first.getByRole("link", { name: clientId })).toBeVisible();
    await expect.element(first.getByRole("cell", { name: "Recipient Web" })).toBeVisible();
    await expect.element(first.getByText("software-1")).toBeVisible();
    await expect.element(first.getByText("ACTIVE", { exact: true })).toBeVisible();
    await expect.element(first.getByRole("cell", { name: "2", exact: true })).toBeVisible();
    await expect.element(first.getByRole("cell", { name: "7", exact: true })).toBeVisible();
    // The tests run in Australia/Perth (vitest.config.ts).
    expect(first.element().textContent).toContain("15 October 2026, 1:30 pm UTC+8");

    const second = table().getByRole("row").nth(2);
    await expect.element(second.getByRole("cell", { name: "Recipient Mobile" })).toBeVisible();
    expect(second.element().textContent).toContain("16 October 2026, 9:00 am UTC+8");
    await expect.element(second.getByText("Not known")).toBeVisible();
    await expect.element(page.getByText("2 records")).toBeVisible();
    await expectNoAxeViolations(document.body);
  });

  test("a URL gives the API the same sort", async () => {
    const requests = serve([registration(clientId)]);

    renderApp("/registrations?sort=arrangementCount&order=desc");

    await expect.element(table().getByRole("link", { name: clientId })).toBeVisible();
    await expect
      .element(page.getByRole("columnheader", { name: "All arrangements" }))
      .toHaveAttribute("aria-sort", "descending");
    expect(requests).toEqual(["?sort=arrangementCount&order=desc"]);
  });

  test("a value in the URL that is not valid is not sent", async () => {
    const requests = serve([registration(clientId)]);

    renderApp("/registrations?sort=softwareId&order=up&page=2&q=recipient");

    await expect.element(table().getByRole("link", { name: clientId })).toBeVisible();
    await expect
      .element(page.getByRole("columnheader", { name: "Client name" }))
      .toHaveAttribute("aria-sort", "ascending");
    expect(requests).toEqual([""]);
  });

  test.each([
    ["Current arrangements", "currentArrangementCount"],
    ["All arrangements", "arrangementCount"],
    ["Registered", "registeredAt"],
  ])("a sort by %s changes the URL and the request", async (header, sort) => {
    const requests = serve([registration(clientId)]);
    const router = renderApp("/registrations");
    await expect.element(table().getByRole("link", { name: clientId })).toBeVisible();

    await userEvent.click(page.getByRole("button", { name: header, exact: true }));

    await expect.poll(() => router.state.location.searchStr).toBe(`?sort=${sort}&order=asc`);
    await expect.poll(() => requests.at(-1)).toBe(`?sort=${sort}&order=asc`);
  });

  test("a sort by the client name turns the other way and adds a history entry", async () => {
    serve([registration(clientId)]);
    const router = renderApp("/registrations");
    await expect.element(table().getByRole("link", { name: clientId })).toBeVisible();
    const historyLength = router.history.length;

    await userEvent.click(page.getByRole("button", { name: "Client name", exact: true }));

    await expect.poll(() => router.state.location.searchStr).toBe("?sort=clientName&order=desc");
    expect(router.history.length).toBe(historyLength + 1);
  });

  test("shows that the data holder has no registrations", async () => {
    serve([]);

    renderApp("/registrations");

    await expect.element(page.getByText("No client registrations")).toBeVisible();
  });

  test("the side navigation opens the registrations, and a row links to the detail", async () => {
    serve([registration(clientId)]);
    worker.use(
      http.get("/api/customers", ({ response }) =>
        response(200).json({ data: [], totalRecords: 0 }),
      ),
      http.get("/api/registrations/{clientId}", ({ response }) =>
        response(200).json(registration(clientId)),
      ),
      http.get("/api/registrations/{clientId}/arrangements", ({ response }) =>
        response(200).json({ data: [], totalRecords: 0 }),
      ),
    );
    const router = renderApp("/customers");

    await userEvent.click(page.getByRole("link", { name: "Client registrations" }));
    await expect
      .element(page.getByRole("heading", { level: 1, name: "Client registrations" }))
      .toBeVisible();
    await userEvent.click(page.getByRole("link", { name: clientId }));

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Recipient Web" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe(`/registrations/${clientId}`);
    await expect
      .element(page.getByRole("link", { name: "Client registrations", exact: true }))
      .toHaveAttribute("aria-current", "page");
  });
});
