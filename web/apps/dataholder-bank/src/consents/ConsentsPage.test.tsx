import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page } from "vitest/browser";

import {
  consent,
  http,
  narrowWidth,
  presets,
  renderApp,
  serveConsents,
  serveSession,
  sessionEnded,
  showPreset,
  worker,
} from "../test/app";

const terms = (within: Element) =>
  [...within.querySelectorAll("dl > div")].map((row) => row.textContent ?? "");

// CDS: Dashboard Standards, Data recipient handling details, with the brand name of the data recipient.
const handling = (brandName: string) =>
  `You should check with ${brandName} for more information on how they are handling your data, and for any ` +
  `other permissions you may have given them. See ${brandName}'s CDR policy or their Dashboard for more ` +
  "information.";

test("the dashboard shows the current consents, then the ended consents in a group", async () => {
  serveConsents([
    consent("arrangement-3", { clientName: "Tax Helper", brandName: "Tax Co" }),
    consent("arrangement-2", {
      clientName: "Old Budget",
      status: "ENDED",
      endedAt: "2026-10-20T12:00:00Z",
    }),
    consent("arrangement-1", { clientName: "Budget Planner" }),
  ]);

  renderApp("/data-sharing");

  await expect.element(page.getByRole("heading", { level: 1, name: "Data sharing" })).toBeVisible();
  await expect.poll(() => document.title).toBe("Data sharing - Example Bank online banking");
  const current = page.getByRole("region", { name: "Current consents" });
  await expect.element(current.getByRole("heading", { name: "Tax Helper" })).toBeVisible();
  const headings = [...document.querySelectorAll("main h3")].map((h) => h.textContent);
  expect(headings).toEqual(["Tax Helper", "Budget Planner", "Old Budget"]);
  const ended = page.getByRole("region", { name: "Ended consents" });
  await expect.element(ended.getByRole("heading", { name: "Old Budget" })).toBeVisible();
  expect(terms(ended.element())).toContain("Ended:20 October 2026");
});

test("a consent shows its data recipient, dates, accounts, data clusters and disclosures", async () => {
  serveConsents([
    consent("arrangement-1", {
      accounts: [
        {
          accountId: "account-1",
          creationDate: "2026-09-01",
          displayName: "Everyday",
          maskedNumber: "xxx-xxx xxxx5678",
          openStatus: "OPEN",
          productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
          productName: "Everyday Account",
        },
        {
          accountId: "account-2",
          creationDate: "2026-09-01",
          displayName: "Old Saver",
          maskedNumber: "xxx-xxx xxxx4321",
          openStatus: "CLOSED",
          productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
          productName: "Saver",
        },
      ],
    }),
  ]);

  renderApp("/data-sharing");

  const card = page.getByRole("article", { name: "Budget Planner" });
  await expect.element(card).toBeVisible();
  expect(terms(card.element())).toEqual([
    "Data recipient:Planner Co",
    "Legal entity:Planner Holdings Pty Ltd",
    "Consent given:1 October 2026",
    "Period:90 days",
    "Sharing ends:30 December 2026",
    "First shared:2 October 2026",
    "Last shared:30 October 2026",
  ]);
  const accounts = [...card.element().querySelectorAll("h4 + ul > li")].map((li) => li.textContent);
  expect(accounts[0]).toBe("Everyday, xxx-xxx xxxx5678");
  expect(accounts[1]).toBe("Old Saver, xxx-xxx xxxx4321 (closed)");
  await expect.element(card.getByText("Account name, type and balance")).toBeVisible();
  await expect
    .element(card.getByText("Name of account, Type of account, Account balance"))
    .toBeVisible();
  await expect.element(card.getByText("No data shared")).toBeVisible();
  await expect.element(card.getByText(handling("Planner Co"), { exact: true })).toBeVisible();
});

test("a consent with no legal entity name, no account and one occasion shows only what it has", async () => {
  const { legalEntityName: _none, ...noLegalEntity } = consent("arrangement-1", {
    sharingDurationSeconds: 0,
    status: "ENDED",
    endedAt: "2026-10-01T12:00:00Z",
    accounts: [],
    dataClusters: [{ name: "Name", permissions: ["Full name and title(s)"] }],
  });
  serveConsents([noLegalEntity]);

  renderApp("/data-sharing");

  const card = page.getByRole("article", { name: "Budget Planner" });
  await expect.element(card).toBeVisible();
  expect(terms(card.element())).toEqual([
    "Data recipient:Planner Co",
    "Consent given:1 October 2026",
    "Period:One occasion",
    "Ended:1 October 2026",
  ]);
  expect(card.getByRole("heading", { name: "Accounts" }).query()).toBeNull();
});

test("a customer with no consent sees that they share no data", async () => {
  serveConsents([]);

  renderApp("/data-sharing");

  await expect
    .element(page.getByText("You do not share your data under the Consumer Data Right."))
    .toBeVisible();
  expect(page.getByRole("region", { name: "Current consents" }).query()).toBeNull();
});

test("a failure of the site API shows the problem", async () => {
  serveSession();
  worker.use(
    http.get("/api/consents", ({ response }) =>
      response("default").json(
        {
          type: "about:blank",
          title: "Internal Server Error",
          status: 500,
          detail: "The data holder could not complete the request.",
        },
        { status: 500 },
      ),
    ),
  );

  renderApp("/data-sharing");

  await expect
    .element(page.getByText("The data holder could not complete the request."))
    .toBeVisible();
});

test("with no session, the dashboard goes to the sign-in page", async () => {
  renderApp("/data-sharing");

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Sign in to online banking" }))
    .toBeVisible();
});

test("when the session ends, the dashboard goes to the sign-in page, which says so", async () => {
  serveConsents([]);
  worker.use(
    http.get("/api/consents", ({ response }) => {
      worker.use(http.get("/api/session", ({ response: ended }) => ended(403).json(sessionEnded)));
      return response(403).json(sessionEnded);
    }),
  );

  renderApp("/data-sharing");

  await expect
    .element(page.getByRole("status").filter({ hasText: "signed you out" }))
    .toHaveTextContent("Example Bank signed you out. Sign in again.");
});

describe.each(presets)("with the preset $name", (preset) => {
  test("the dashboard passes the accessibility checks and has no horizontal scroll at 320 CSS px", async () => {
    serveConsents([
      consent("arrangement-2", { clientName: "Budget Planner With A Very Long Software Name" }),
      consent("arrangement-1", { status: "ENDED", endedAt: "2026-10-20T12:00:00Z" }),
    ]);

    await showPreset(preset, "/data-sharing");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Data sharing" }))
      .toBeVisible();
    await expect.element(page.getByRole("region", { name: "Ended consents" })).toBeVisible();
    await expect.element(page.getByText(handling("Planner Co")).first()).toBeVisible();
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
    await expectNoAxeViolations(document.documentElement);
  });
});
