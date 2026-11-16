import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { expect, test } from "vitest";
import { page } from "vitest/browser";

import type { Overview } from "../api/client";
import { definitionOf, http, renderApp, worker } from "../test/app";

const overview: Overview = {
  customerCount: 128,
  openAccountCount: 214,
  closedAccountCount: 19,
  currentArrangementCount: 37,
  endedArrangementCount: 112,
  registrationCount: 4,
  notActiveSoftwareProductCount: 1,
  changes: [
    {
      time: "2026-11-03T06:41:07Z",
      kind: "ARRANGEMENT_CONFIRMED",
      customerNumber: "00000001",
      clientId: "client-1",
      clientName: "Recipient Web",
      clientRegistered: true,
      sharingEndsAt: "2027-11-03T06:41:07Z",
    },
    {
      time: "2026-11-03T06:40:07Z",
      kind: "ARRANGEMENT_ENDED",
      customerNumber: "00000002",
      clientId: "client-2",
      clientName: "Recipient Beta",
      clientRegistered: false,
      endReason: "WITHDRAWN_BY_CONSUMER",
      withdrawalMethod: "CONSUMER_DASHBOARD",
    },
    {
      time: "2026-11-03T06:39:07Z",
      kind: "ACCOUNT_CREATED",
      customerNumber: "00000002",
      productName: "Wattle Pocket",
    },
    {
      time: "2026-11-03T06:38:07Z",
      kind: "CLIENT_DELETED",
      clientId: "client-2",
      clientName: "Recipient Beta",
      clientRegistered: false,
    },
  ],
};

function serve(body: Overview) {
  worker.use(http.get("/api/overview", ({ response }) => response(200).json(body)));
}

const changes = () => page.getByRole("table", { name: "Newest changes" });
const rowOf = (time: string) => changes().getByRole("row").filter({ hasText: time });

// Verifies: DH-1403
test("the first page of the console is the overview, and its navigation item is current", async () => {
  serve(overview);

  renderApp("/");

  await expect.element(page.getByRole("heading", { level: 1, name: "Overview" })).toBeVisible();
  const navigation = page.getByRole("navigation", { name: "Main" });
  await expect
    .element(navigation.getByRole("link", { name: "Overview", exact: true }))
    .toHaveAttribute("aria-current", "page");
  expect(navigation.getByRole("link").first().element().textContent).toBe("Overview");
});

// Verifies: DH-1404, DH-1405
test("the overview gives each count, with the counts that give it context and the links to the lists", async () => {
  serve(overview);

  renderApp("/");

  await expect.element(page.getByText("214", { exact: true })).toBeVisible();
  expect(definitionOf(document.body, "Customers").textContent).toBe("128");
  expect(definitionOf(document.body, "Open accounts").textContent).toBe("214");
  expect(definitionOf(document.body, "Current arrangements").textContent).toBe("37");
  expect(definitionOf(document.body, "Client registrations").textContent).toBe("4");
  await expect.element(page.getByText("19 closed")).toBeVisible();
  await expect.element(page.getByText("112 ended")).toBeVisible();
  await expect.element(page.getByText("1 software product not active")).toBeVisible();
  await expect
    .element(page.getByRole("link", { name: "View the customers" }))
    .toHaveAttribute("href", "/customers");
  await expect
    .element(page.getByRole("link", { name: "View the client registrations" }))
    .toHaveAttribute("href", "/registrations");
  await expectNoAxeViolations(document.body);
});

// Verifies: DH-1408, DH-1409, DH-1410, DH-1411
test("the newest changes give the time, the change, the customer number, the software product and the detail", async () => {
  serve(overview);

  renderApp("/");

  // The tests run in Australia/Perth (vitest.config.ts).
  const confirmed = rowOf("2:41:07 pm UTC+8");
  await expect.element(confirmed.getByText("Arrangement confirmed")).toBeVisible();
  await expect
    .element(confirmed.getByRole("link", { name: "00000001" }))
    .toHaveAttribute("href", "/customers/00000001");
  await expect
    .element(confirmed.getByRole("link", { name: "Recipient Web" }))
    .toHaveAttribute("href", "/registrations/client-1");
  const sharingEnds = confirmed.getByRole("cell").filter({ hasText: "Sharing ends" });
  expect(sharingEnds.element().textContent?.replaceAll(/\s/g, " ")).toContain(
    "Sharing ends 3 November 2027",
  );

  const ended = rowOf("2:40:07 pm UTC+8");
  await expect.element(ended.getByText("Arrangement ended")).toBeVisible();
  // A registration that does not exist now has no link.
  await expect.element(ended.getByText("Recipient Beta")).toBeVisible();
  await expect.element(ended.getByRole("link", { name: "Recipient Beta" })).not.toBeInTheDocument();
  await expect.element(ended.getByText("WITHDRAWN_BY_CONSUMER, CONSUMER_DASHBOARD")).toBeVisible();

  await expect.element(rowOf("2:39:07 pm UTC+8").getByText("Wattle Pocket")).toBeVisible();
  await expect.element(rowOf("2:38:07 pm UTC+8").getByText("Registration deleted")).toBeVisible();
});

test("an overview with no changes says so", async () => {
  serve({ ...overview, changes: [] });

  renderApp("/");

  await expect.element(page.getByText("No changes")).toBeVisible();
});

test("an overview that fails shows the problem", async () => {
  worker.use(
    http.untyped.get("/api/overview", () =>
      Response.json(
        {
          type: "about:blank",
          title: "Internal Server Error",
          status: 500,
          detail: "The data holder could not complete the request.",
        },
        { status: 500, headers: { "Content-Type": "application/problem+json" } },
      ),
    ),
  );

  renderApp("/");

  await expect
    .element(page.getByText("The data holder could not complete the request."))
    .toBeVisible();
});
