import { expect, test, type APIRequestContext, type Page } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import { publicURL } from "./lab";

const keySet = (page: Page, name: string) => page.getByRole("region", { name, exact: true });
const keyID = (page: Page, set: string, state: string) =>
  keySet(page, set).getByRole("region", { name: state }).locator("code");

function isJWKS(body: unknown): body is { keys: { kid: string }[] } {
  return typeof body === "object" && body !== null && "keys" in body && Array.isArray(body.keys);
}

async function ssaJWKSKeyIDs(request: APIRequestContext): Promise<string[]> {
  const response = await request.get(publicURL("/cdr-register/v1/jwks"));
  expect(response.status()).toBe(200);
  const body: unknown = await response.json();
  if (!isJWKS(body)) throw new Error("Get JWKS gave no key set");
  return body.keys.map((k) => k.kid).toSorted();
}

test("the operator sees a key that signs in each key set", async ({ page }) => {
  await page.goto("/");
  await page
    .getByRole("navigation", { name: "Main" })
    .getByRole("link", { name: "Signing keys" })
    .click();

  await expect(page).toHaveURL("/signing-keys");
  const signing = (set: string) =>
    keySet(page, set).getByRole("region", { name: "Key that signs" });
  await Promise.all(
    ["Access token keys", "SSA keys"].flatMap((set) => [
      expect(signing(set).locator("code")).toHaveText(/\S/),
      expect(signing(set).locator("time")).toContainText(/ [ap]m UTC\+8/),
    ]),
  );
});

// The conformance key tests compare only with the state at their start, so this change does not break them.
test("the operator changes the SSA key through the two confirmations", async ({
  page,
  request,
}) => {
  await page.goto("/signing-keys");
  const actions = page.getByRole("group", { name: "Key actions of SSA keys" });
  await expect(actions.getByRole("button")).toHaveText(["Publish next key"]);
  const oldKey = (await keyID(page, "SSA keys", "Key that signs").textContent()) ?? "";
  const accessTokenKeys = await keySet(page, "Access token keys").locator("code").allTextContents();
  const previous = keyID(page, "SSA keys", "Previous key");
  const leaving = (await previous.count()) > 0 ? await previous.textContent() : undefined;

  await actions.getByRole("button", { name: "Publish next key" }).click();
  const publish = page.getByRole("alertdialog", { name: "Publish the next SSA key?" });
  await expect(publish).toContainText("The key that signs does not change.");
  if (leaving) {
    await expect(publish).toContainText(
      `Key ${leaving} leaves the JWKS, and the Register deletes it.`,
    );
  }
  await expectNoAxeViolations(page);
  await publish.getByRole("button", { name: "Publish", exact: true }).click();

  await expect(publish).toBeHidden();
  await expect(page.getByRole("dialog", { name: "The next SSA key is published" })).toBeVisible();
  const newKey = (await keyID(page, "SSA keys", "Next key").textContent()) ?? "";
  expect(newKey).not.toBe(oldKey);
  expect(await ssaJWKSKeyIDs(request)).toEqual([oldKey, newKey].toSorted());

  await actions.getByRole("button", { name: "Activate next key" }).click();
  const activate = page.getByRole("alertdialog", { name: "Activate the next SSA key?" });
  await expect(activate.getByRole("listitem")).toHaveText([
    `Key ${oldKey} becomes the previous key. The JWKS still publishes it.`,
    `Key ${newKey} starts to sign the SSAs.`,
    "A participant with an old copy of the JWKS cannot verify the new signatures.",
  ]);
  await expectNoAxeViolations(page);
  await activate.getByRole("button", { name: "Activate", exact: true }).click();

  await expect(activate).toBeHidden();
  await expect(keyID(page, "SSA keys", "Key that signs")).toHaveText(newKey);
  await expect(keyID(page, "SSA keys", "Previous key")).toHaveText(oldKey);
  await expect(actions.getByRole("button")).toHaveText(["Publish next key"]);
  expect(await ssaJWKSKeyIDs(request)).toEqual([oldKey, newKey].toSorted());
  expect(await keySet(page, "Access token keys").locator("code").allTextContents()).toEqual(
    accessTokenKeys,
  );
  await expectNoAxeViolations(page);
});
