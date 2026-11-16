import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { KeySetList } from "../api/client";
import { http, renderApp, worker } from "../test/app";

const keyChange: KeySetList = {
  data: [
    {
      keySet: "access-token",
      version: 3,
      actions: [],
      keys: [
        { keyId: "at-2", state: "signing", createdAt: "2026-10-23T02:05:07.123Z" },
        { keyId: "at-1", state: "previous", createdAt: "2026-10-22T23:59:59Z" },
      ],
    },
    {
      keySet: "ssa",
      version: 2,
      actions: [],
      keys: [
        { keyId: "ssa-1", state: "signing", createdAt: "2026-10-22T01:00:00Z" },
        { keyId: "ssa-2", state: "next", createdAt: "2026-10-23T14:30:00Z" },
      ],
    },
  ],
};

function serve(body: KeySetList) {
  worker.use(http.get("/api/key-sets", ({ response }) => response(200).json(body)));
}

const keySet = (name: string) => page.getByRole("region", { name, exact: true });
const key = (set: string, state: string) => keySet(set).getByRole("region", { name: state });

describe("the signing keys page", () => {
  // Verifies: REG-1042, REG-1057
  test("shows the key ID and the creation time of each key in the time zone of the browser, under its state", async () => {
    serve(keyChange);

    renderApp("/signing-keys");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Signing keys" }))
      .toBeVisible();
    const signing = key("Access token keys", "Key that signs");
    await expect.element(signing.getByText("at-2", { exact: true })).toBeVisible();
    // The tests run in Australia/Perth (vitest.config.ts).
    const signingTime = signing.element().querySelector("time");
    expect(signingTime?.textContent).toContain("23 October 2026, 10:05:07 am UTC+8");
    expect(signingTime?.dateTime).toBe("2026-10-23T02:05:07.123Z");
    await expect
      .element(key("Access token keys", "Previous key").getByText("at-1", { exact: true }))
      .toBeVisible();
    await expect
      .element(key("SSA keys", "Key that signs").getByText("ssa-1", { exact: true }))
      .toBeVisible();
    const next = key("SSA keys", "Next key");
    await expect.element(next.getByText("ssa-2", { exact: true })).toBeVisible();
    expect(next.element().querySelector("time")?.textContent).toContain(
      "23 October 2026, 10:30:00 pm UTC+8",
    );
    await expect
      .element(page.getByRole("button", { name: "Copy key ID of the next SSA key" }))
      .toBeVisible();
    await expect
      .element(page.getByRole("button", { name: "Copy key ID of the access token key that signs" }))
      .toBeVisible();
  });

  test("shows only the keys that a key set has", async () => {
    serve({
      data: [
        {
          keySet: "access-token",
          version: 1,
          actions: [],
          keys: [{ keyId: "at-1", state: "signing", createdAt: "2026-10-23T00:00:00Z" }],
        },
        { keySet: "ssa", version: 1, actions: [], keys: [] },
      ],
    });

    renderApp("/signing-keys");

    await expect
      .element(key("Access token keys", "Key that signs").getByText("at-1", { exact: true }))
      .toBeVisible();
    await expect.element(keySet("Access token keys").getByText("Next key")).not.toBeInTheDocument();
    await expect.element(keySet("SSA keys").getByText("This key set has no keys")).toBeVisible();
  });

  test("shows the title and the detail of a problem from the API", async () => {
    worker.use(
      http.get("/api/key-sets", ({ response }) =>
        response("default").json(
          {
            type: "about:blank",
            title: "Internal Server Error",
            status: 500,
            detail: "The Register could not complete the request.",
          },
          { status: 500 },
        ),
      ),
    );

    renderApp("/signing-keys");

    const alert = page.getByRole("alert");
    await expect.element(alert.getByText("Internal Server Error")).toBeVisible();
    await expect
      .element(alert.getByText("The Register could not complete the request."))
      .toBeVisible();
  });

  test("the navigation marks the page, and goes to it from another page", async () => {
    serve(keyChange);
    worker.use(
      http.get("/api/data-recipients", ({ response }) =>
        response(200).json({ data: [], totalRecords: 0 }),
      ),
    );
    const router = renderApp("/data-recipients");

    await userEvent.click(page.getByRole("link", { name: "Signing keys" }));

    await expect.poll(() => router.state.location.pathname).toBe("/signing-keys");
    await expect
      .element(page.getByRole("link", { name: "Signing keys" }))
      .toHaveAttribute("aria-current", "page");
    await expect
      .element(key("SSA keys", "Next key").getByText("ssa-2", { exact: true }))
      .toBeVisible();
  });

  // Verifies: REG-1060
  test("passes the automated accessibility checks", async () => {
    serve(keyChange);

    renderApp("/signing-keys");
    await expect
      .element(key("SSA keys", "Next key").getByText("ssa-2", { exact: true }))
      .toBeVisible();

    await expectNoAxeViolations(document.body);
  });
});
