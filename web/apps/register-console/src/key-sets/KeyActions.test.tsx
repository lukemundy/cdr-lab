import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { KeySet, KeySetList } from "../api/client";
import { http, renderApp, worker } from "../test/app";

const created = "2026-10-23T01:00:00Z";

const accessTokenKeys: KeySet = {
  keySet: "access-token",
  version: 1,
  keys: [{ keyId: "at-1", state: "signing", createdAt: created }],
  actions: [{ action: "publish-next", irreversible: true, changes: [] }],
};

const ssaAfterChange: KeySet = {
  keySet: "ssa",
  version: 3,
  keys: [
    { keyId: "ssa-2", state: "signing", createdAt: created },
    { keyId: "ssa-1", state: "previous", createdAt: created },
  ],
  actions: [
    { action: "publish-next", irreversible: true, changes: [{ keyId: "ssa-1", from: "previous" }] },
  ],
};

const ssaPublished: KeySet = {
  keySet: "ssa",
  version: 4,
  keys: [
    { keyId: "ssa-2", state: "signing", createdAt: created },
    { keyId: "ssa-3", state: "next", createdAt: created },
  ],
  actions: [
    {
      action: "activate-next",
      irreversible: true,
      changes: [
        { keyId: "ssa-2", from: "signing", to: "previous" },
        { keyId: "ssa-3", from: "next", to: "signing" },
      ],
    },
  ],
};

const ssaActivated: KeySet = {
  keySet: "ssa",
  version: 5,
  keys: [
    { keyId: "ssa-3", state: "signing", createdAt: created },
    { keyId: "ssa-2", state: "previous", createdAt: created },
  ],
  actions: [
    { action: "publish-next", irreversible: true, changes: [{ keyId: "ssa-2", from: "previous" }] },
  ],
};

const conflict = {
  type: "about:blank",
  title: "Conflict",
  status: 409,
  detail: "The key set changed after the version in the request. The Register changed nothing.",
} as const;

// The key sets depend on the step requests, not on the count of reads, because a read of a
// previous test can arrive late.
function serveKeySets(before: KeySet, after: KeySet = before) {
  const requests: unknown[] = [];
  worker.use(
    http.get("/api/key-sets", ({ response }) => {
      const body: KeySetList = { data: [accessTokenKeys, requests.length === 0 ? before : after] };
      return response(200).json(body);
    }),
  );
  return requests;
}

function serveStep(
  requests: unknown[],
  action: "publish-next" | "activate-next",
  result: "success" | "conflict",
) {
  worker.use(
    http.post(`/api/key-sets/{keySet}/${action}`, async ({ params, request, response }) => {
      requests.push({ ...params, body: await request.json() });
      return result === "success" ? response(204).empty() : response(409).json(conflict);
    }),
  );
}

const actionsOf = (name: string) => page.getByRole("group", { name: `Key actions of ${name}` });
const buttonNames = (name: string) =>
  actionsOf(name)
    .getByRole("button")
    .all()
    .map((button) => button.element().textContent);
const key = (set: string, state: string) =>
  page.getByRole("region", { name: set, exact: true }).getByRole("region", { name: state });

describe("the key actions of a key set", () => {
  test("offer only the step that the console API gives for each key set", async () => {
    serveKeySets(ssaPublished);

    renderApp("/signing-keys");

    await expect.element(actionsOf("SSA keys")).toBeVisible();
    expect(buttonNames("Access token keys")).toEqual(["Publish next key"]);
    expect(buttonNames("SSA keys")).toEqual(["Activate next key"]);
    await expectNoAxeViolations(document.body);
  });

  test("a publish names the previous key that leaves the JWKS, sends the version, and shows the new state", async () => {
    const requests = serveKeySets(ssaAfterChange, ssaPublished);
    serveStep(requests, "publish-next", "success");
    renderApp("/signing-keys");

    await userEvent.click(actionsOf("SSA keys").getByRole("button", { name: "Publish next key" }));

    const dialog = page.getByRole("alertdialog", { name: "Publish the next SSA key?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        "The JWKS of the SSA keys publishes a new next key. The key that signs does not change. " +
          "Key ssa-1 leaves the JWKS, and the Register deletes it. " +
          "You cannot reverse this action.",
      );
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Publish", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([{ keySet: "ssa", body: { version: 3 } }]);
    await expect
      .element(page.getByRole("dialog", { name: "The next SSA key is published" }))
      .toBeVisible();
    await expect.element(key("SSA keys", "Next key").getByText("ssa-3")).toBeVisible();
    expect(buttonNames("SSA keys")).toEqual(["Activate next key"]);
    await expect
      .poll(() => document.activeElement)
      .toBe(actionsOf("SSA keys").getByRole("button", { name: "Activate next key" }).element());
  });

  test("a publish with no previous key names no key that leaves the JWKS", async () => {
    serveKeySets(ssaAfterChange);
    renderApp("/signing-keys");

    await userEvent.click(
      actionsOf("Access token keys").getByRole("button", { name: "Publish next key" }),
    );

    await expect
      .element(page.getByRole("alertdialog", { name: "Publish the next access token key?" }))
      .toHaveAccessibleDescription(
        "The JWKS of the access token keys publishes a new next key. The key that signs does not change. " +
          "You cannot reverse this action.",
      );
  });

  test("an activation names the key that starts to sign and the key that becomes the previous key", async () => {
    const requests = serveKeySets(ssaPublished, ssaActivated);
    serveStep(requests, "activate-next", "success");
    renderApp("/signing-keys");

    await userEvent.click(actionsOf("SSA keys").getByRole("button", { name: "Activate next key" }));

    const dialog = page.getByRole("alertdialog", { name: "Activate the next SSA key?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        "Key ssa-2 becomes the previous key. The JWKS still publishes it. " +
          "Key ssa-3 starts to sign the SSAs. " +
          "A participant with an old copy of the JWKS cannot verify the new signatures. " +
          "You cannot reverse this action.",
      );
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Activate", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([{ keySet: "ssa", body: { version: 4 } }]);
    await expect
      .element(page.getByRole("dialog", { name: "The next SSA key now signs" }))
      .toBeVisible();
    await expect.element(key("SSA keys", "Key that signs").getByText("ssa-3")).toBeVisible();
    await expect.element(key("SSA keys", "Previous key").getByText("ssa-2")).toBeVisible();
  });

  test("a refused step shows the detail in the confirmation, also when the current state no longer permits the step", async () => {
    serveStep(serveKeySets(ssaPublished, ssaActivated), "activate-next", "conflict");
    renderApp("/signing-keys");

    await userEvent.click(actionsOf("SSA keys").getByRole("button", { name: "Activate next key" }));
    const dialog = page.getByRole("alertdialog", { name: "Activate the next SSA key?" });
    await dialog.getByRole("button", { name: "Activate", exact: true }).click();

    await expect.element(dialog.getByRole("alert")).toMatchTextContent(conflict.detail);
    expect(page.getByRole("dialog", { name: /now signs/ }).query()).toBeNull();

    await dialog.getByRole("button", { name: "Cancel" }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(buttonNames("SSA keys")).toEqual(["Publish next key"]);
    await expect
      .poll(() => document.activeElement)
      .toBe(actionsOf("SSA keys").getByRole("button", { name: "Publish next key" }).element());
  });
});
