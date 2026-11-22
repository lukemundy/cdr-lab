import { expect, test, type APIRequestContext } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";
import { publicURL, showAllBrands } from "./lab";

// Only this test changes these records of services/register/seed/example.json, and it gives back their statuses.
const recipient = { name: "Console Test Pty Ltd", id: "7c89d9d4-4f19-4bb3-90c3-b2d5b95520c4" };
const web = { name: "Console Test Web", id: "f408f354-ef53-4342-93df-5266601ff735" };
const mobile = { name: "Console Test Mobile", id: "3b9f04e2-c47b-4573-bbc5-797616237370" };

type StatusList = {
  data: { legalEntityId?: string; softwareProductId?: string; status: string }[];
};

function isStatusList(body: unknown): body is StatusList {
  return typeof body === "object" && body !== null && "data" in body && Array.isArray(body.data);
}

async function cdrStatuses(
  request: APIRequestContext,
  path: string,
): Promise<Record<string, string>> {
  const response = await request.get(publicURL(`/cdr-register/v1/all/${path}`), {
    headers: { "x-v": "3" },
  });
  expect(response.status()).toBe(200);
  const body: unknown = await response.json();
  if (!isStatusList(body)) throw new Error(`${path} gave no list of statuses`);
  return Object.fromEntries(
    body.data.map((d) => [d.legalEntityId ?? d.softwareProductId, d.status]),
  );
}

async function expectCdrStatuses(
  request: APIRequestContext,
  want: { recipient: string; web: string; mobile: string },
) {
  const recipients = await cdrStatuses(request, "data-recipients/status");
  const products = await cdrStatuses(request, "data-recipients/brands/software-products/status");
  expect({
    recipient: recipients[recipient.id],
    web: products[web.id],
    mobile: products[mobile.id],
  }).toEqual(want);
}

test("the operator suspends a data recipient through the confirmation, and then gives back its statuses", async ({
  page,
  request,
}) => {
  const violations: string[] = [];
  await page.addInitScript(() => {
    document.addEventListener("securitypolicyviolation", (event) => {
      console.error(`CSP violation: ${event.effectiveDirective}`);
    });
  });
  page.on("console", (message) => {
    if (message.text().startsWith("CSP violation")) violations.push(message.text());
  });
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  await expectCdrStatuses(request, { recipient: "ACTIVE", web: "ACTIVE", mobile: "INACTIVE" });

  await page.goto(`/data-recipients/${recipient.id}`);

  await showAllBrands(page);

  const actions = (name: string) =>
    page.getByRole("group", { name: `Status actions of ${name}`, exact: true });
  await expect(actions(recipient.name).getByRole("button")).toHaveText([
    "Suspend",
    "Revoke",
    "Surrender",
  ]);
  await expect(actions(web.name).getByRole("button")).toHaveText(["Deactivate", "Remove"]);
  await expect(actions(mobile.name).getByRole("button")).toHaveText(["Activate", "Remove"]);

  await actions(recipient.name).getByRole("button", { name: "Suspend" }).click();
  const suspension = page.getByRole("alertdialog", { name: `Suspend ${recipient.name}?` });
  await expect(suspension.getByRole("listitem")).toHaveText([
    `${recipient.name} changes from ACTIVE to SUSPENDED.`,
    `${web.name} changes from ACTIVE to INACTIVE.`,
  ]);
  await expectNoAxeViolations(page);
  await suspension.getByRole("button", { name: "Suspend", exact: true }).click();

  await expect(suspension).toBeHidden();
  await expect(
    page.getByRole("dialog", { name: `${recipient.name} is now SUSPENDED` }),
  ).toBeVisible();
  const entity = page.getByRole("region", { name: "Legal entity", exact: true });
  await expect(entity.getByText("SUSPENDED", { exact: true })).toBeVisible();
  await expect(actions(recipient.name).getByRole("button")).toHaveText([
    "Revoke",
    "Surrender",
    "Reinstate",
  ]);
  await expectCdrStatuses(request, { recipient: "SUSPENDED", web: "INACTIVE", mobile: "INACTIVE" });
  await expectNoAxeViolations(page);

  await actions(recipient.name).getByRole("button", { name: "Reinstate" }).click();
  const reinstatement = page.getByRole("alertdialog", { name: `Reinstate ${recipient.name}?` });
  await expect(reinstatement).toContainText(
    "The software products do not change. Activate each software product that must be ACTIVE.",
  );
  await reinstatement.getByRole("button", { name: "Reinstate", exact: true }).click();
  await expect(reinstatement).toBeHidden();
  await expect(entity.getByText("ACTIVE", { exact: true })).toBeVisible();

  await actions(web.name).getByRole("button", { name: "Activate" }).click();
  const activation = page.getByRole("alertdialog", { name: `Activate ${web.name}?` });
  await activation.getByRole("button", { name: "Activate", exact: true }).click();
  await expect(activation).toBeHidden();
  await expect(actions(web.name).getByRole("button")).toHaveText(["Deactivate", "Remove"]);

  await expectCdrStatuses(request, { recipient: "ACTIVE", web: "ACTIVE", mobile: "INACTIVE" });
  expect(violations).toEqual([]);
  expect(errors).toEqual([]);
});
