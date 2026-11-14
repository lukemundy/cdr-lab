import { expect, test, type APIRequestContext, type Page } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import { publicURL } from "./lab";

// Only this test changes this record of services/register/seed/example.json, and it gives back its details.
const recipient = {
  id: "1e3767b2-7bd8-435a-b0a4-e18a1be73b9a",
  name: "Console Change Test Pty Ltd",
  accreditationNumber: "ADR-MOCK-0013",
};
const changedName = "Console Change Test Group Pty Ltd";

type DataRecipients = { data: { legalEntityId: string; legalEntityName: string }[] };

function isDataRecipients(body: unknown): body is DataRecipients {
  return typeof body === "object" && body !== null && "data" in body && Array.isArray(body.data);
}

async function cdrName(request: APIRequestContext): Promise<string | undefined> {
  const response = await request.get(publicURL("/cdr-register/v1/all/data-recipients"), {
    headers: { "x-v": "4" },
  });
  expect(response.status()).toBe(200);
  const body: unknown = await response.json();
  if (!isDataRecipients(body)) throw new Error("Get Data Recipients gave no list");
  return body.data.find((d) => d.legalEntityId === recipient.id)?.legalEntityName;
}

async function changeName(page: Page, from: string, to: string) {
  await page.goto(`/data-recipients/${recipient.id}`);
  await page
    .getByRole("region", { name: "Legal entity", exact: true })
    .getByRole("link", { name: "Change details" })
    .click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(`Change the details of ${from}`);
  await page.getByRole("textbox", { name: "Name", exact: true }).fill(to);
  await page.getByRole("button", { name: "Save changes" }).click();
  await expect(page.getByRole("heading", { level: 1, name: to, exact: true })).toBeFocused();
}

test("the operator changes the name of a data recipient, and the CDR APIs and the history show it", async ({
  page,
  request,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  expect(await cdrName(request)).toBe(recipient.name);

  await page.goto(`/data-recipients/${recipient.id}/change`);
  const accreditationNumber = page.getByRole("textbox", { name: "Accreditation number" });
  await expect(accreditationNumber).toHaveValue(recipient.accreditationNumber);
  await accreditationNumber.fill("");
  await page.getByRole("button", { name: "Save changes" }).click();

  const summary = page.getByRole("alert");
  await expect(
    summary.getByRole("link", { name: "accreditationNumber is required" }),
  ).toBeVisible();
  await expect(accreditationNumber).toHaveAccessibleDescription("accreditationNumber is required");
  await expectNoAxeViolations(page);

  await changeName(page, recipient.name, changedName);

  await expect(
    page.getByRole("dialog", { name: `Details of ${changedName} changed` }),
  ).toBeVisible();
  await expectNoAxeViolations(page);
  expect(await cdrName(request)).toBe(changedName);
  await page.getByRole("link", { name: "History", exact: true }).click();
  await expect(
    page.getByRole("list", { name: "History" }).getByText("Legal entity details changed").first(),
  ).toBeVisible();

  await changeName(page, changedName, recipient.name);
  expect(await cdrName(request)).toBe(recipient.name);
  expect(errors).toEqual([]);
});
