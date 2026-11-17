import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { Consent } from "../api/client";
import {
  consent,
  endSession,
  http,
  narrowWidth,
  noSession,
  presets,
  renderApp,
  serveConsents,
  sessionEnded,
  showPreset,
  worker,
} from "../test/app";

// All consent data in the tests is invented.

// serveWithdrawals gives each withdrawal to respond, which can change the consents that the site API serves.
function serveWithdrawals(consents: Consent[], respond: (arrangementId: string) => Response) {
  serveConsents(consents);
  const requests: string[] = [];
  worker.use(
    http.get("/api/consents", ({ response }) => response(200).json({ consents })),
    http.post("/api/consents/{arrangementId}/withdraw", ({ params }) => {
      requests.push(params.arrangementId);
      return respond(params.arrangementId);
    }),
  );
  return requests;
}

const end = (consents: Consent[]) => (arrangementId: string) => {
  const index = consents.findIndex((c) => c.arrangementId === arrangementId);
  consents[index] = { ...consents[index]!, status: "ENDED", endedAt: "2026-11-02T12:00:00Z" };
  return new Response(null, { status: 204 });
};

const problem = (status: number, detail: string) =>
  Response.json({ type: "about:blank", title: "Refused", status, detail }, { status });

const dialog = () => page.getByRole("alertdialog");
const stopSharing = () => page.getByRole("button", { name: "Stop sharing" });

// Verifies: DH-1924
test("an ended consent has no action to stop the sharing", async () => {
  serveConsents([consent("arrangement-1", { status: "ENDED", endedAt: "2026-10-20T12:00:00Z" })]);

  renderApp("/data-sharing");

  await expect.element(page.getByRole("region", { name: "Ended consents" })).toBeVisible();
  expect(stopSharing().query()).toBeNull();
});

// Verifies: DH-1926, DH-1929
test("the confirmation names the app and the data recipient, and shows the messages of the CDS word for word", async () => {
  serveConsents([
    consent("arrangement-1", { clientName: "Budget Planner", brandName: "Planner Co" }),
  ]);

  renderApp("/data-sharing");
  await userEvent.click(stopSharing());

  await expect
    .element(dialog().getByRole("heading", { name: "Stop sharing with Budget Planner?" }))
    .toBeVisible();
  const text = dialog().element().textContent;
  expect(text).toContain(
    "Example Bank stops sharing your data with Budget Planner of Planner Co at once.",
  );
  expect(text).toContain("Example Bank tells Planner Co that you stopped the sharing.");
  // CDS: Withdrawal Standards, Withdrawing authorisation: Consequences and Redundant data.
  expect(text).toContain(
    "You should check with Planner Co before you stop sharing to understand the consequences.",
  );
  expect(text).toContain(
    "CDR data is either deleted or de-identified when it is no longer required.",
  );
  expect(text).toContain("You cannot reverse this action.");
  await expect.element(dialog().getByRole("button", { name: "Cancel" })).toBeVisible();
  await expectNoAxeViolations(document.documentElement);
});

// Verifies: DH-1925
test("a cancel sends no withdrawal", async () => {
  const requests = serveWithdrawals(
    [consent("arrangement-1")],
    () => new Response(null, { status: 204 }),
  );

  renderApp("/data-sharing");
  await userEvent.click(stopSharing());
  await userEvent.click(dialog().getByRole("button", { name: "Cancel" }));

  await expect.element(dialog()).not.toBeInTheDocument();
  expect(requests).toEqual([]);
});

// Verifies: DH-1931, DH-1932
test("a withdrawal says that the sharing stopped, and shows the consent in the ended group", async () => {
  const consents = [consent("arrangement-1", { clientName: "Budget Planner" })];
  const requests = serveWithdrawals(consents, end(consents));

  renderApp("/data-sharing");
  await userEvent.click(stopSharing());
  await userEvent.click(dialog().getByRole("button", { name: "Stop sharing" }));

  const heading = page.getByRole("heading", { level: 1, name: "Data sharing" });
  await expect
    .element(page.getByText("You stopped sharing your data with Budget Planner."))
    .toBeVisible();
  await expect.element(heading).toHaveFocus();
  const ended = page.getByRole("region", { name: "Ended consents" });
  await expect.element(ended.getByRole("article", { name: "Budget Planner" })).toBeVisible();
  expect(page.getByRole("region", { name: "Current consents" }).query()).toBeNull();
  expect(requests).toEqual(["arrangement-1"]);
});

// Verifies: DH-1933
test("a withdrawal of a consent that ended keeps the dialog, says that the consent changed, then shows the current state", async () => {
  const consents = [consent("arrangement-1", { clientName: "Budget Planner" })];
  const requests = serveWithdrawals(consents, (arrangementId) => {
    end(consents)(arrangementId);
    return problem(
      409,
      "This consent changed after you opened this confirmation: it already ended. We changed nothing.",
    );
  });

  renderApp("/data-sharing");
  await userEvent.click(stopSharing());
  await userEvent.click(dialog().getByRole("button", { name: "Stop sharing" }));

  await expect
    .element(
      dialog().getByText("This consent changed after you opened this confirmation", {
        exact: false,
      }),
    )
    .toBeVisible();
  await userEvent.click(dialog().getByRole("button", { name: "Cancel" }));

  await expect.element(dialog()).not.toBeInTheDocument();
  const ended = page.getByRole("region", { name: "Ended consents" });
  await expect.element(ended.getByRole("article", { name: "Budget Planner" })).toBeVisible();
  expect(requests).toEqual(["arrangement-1"]);
});

// Verifies: DH-1934, DH-1935
test("a withdrawal that fails keeps the dialog open with the error, and the site does not send it again", async () => {
  const requests = serveWithdrawals([consent("arrangement-1")], () =>
    problem(500, "The data holder could not complete the request."),
  );

  renderApp("/data-sharing");
  await userEvent.click(stopSharing());
  await userEvent.click(dialog().getByRole("button", { name: "Stop sharing" }));

  await expect
    .element(dialog().getByText("The data holder could not complete the request."))
    .toBeVisible();
  expect(requests).toEqual(["arrangement-1"]);
});

// Verifies: DH-2028
test("a withdrawal after the end of the session goes to the sign-in page, which says so", async () => {
  serveWithdrawals([consent("arrangement-1")], () => {
    endSession();
    worker.use(http.get("/api/consents", ({ response }) => response(403).json(noSession)));
    return Response.json(sessionEnded, { status: 403 });
  });

  renderApp("/data-sharing");
  await userEvent.click(stopSharing());
  await userEvent.click(dialog().getByRole("button", { name: "Stop sharing" }));
  await expect
    .element(dialog().getByText("The bank signed you out.", { exact: false }))
    .toBeVisible();
  await userEvent.click(dialog().getByRole("button", { name: "Cancel" }));

  await expect
    .element(page.getByRole("status").filter({ hasText: "signed you out" }))
    .toHaveTextContent("Example Bank signed you out. Sign in again.");
});

describe.each(presets)("with the preset $name", (preset) => {
  // Verifies: DH-1854, DH-1856
  test("the confirmation of a withdrawal passes the accessibility checks and has no horizontal scroll at 320 CSS px", async () => {
    serveConsents([consent("arrangement-1")]);

    await showPreset(preset, "/data-sharing");
    await userEvent.click(stopSharing());

    await expect.element(dialog()).toBeVisible();
    expect(dialog().element().textContent).toContain(
      "You should check with Planner Co before you stop sharing to understand the consequences.",
    );
    expect(dialog().element().textContent).toContain(
      "CDR data is either deleted or de-identified when it is no longer required.",
    );
    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(narrowWidth);
    await expectNoAxeViolations(document.documentElement);
  });
});
