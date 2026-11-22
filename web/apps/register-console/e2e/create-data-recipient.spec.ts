import { expect, test, type APIRequestContext } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/app/e2e/checks";
import { publicURL } from "./lab";

// Mock Budgeting Pty Ltd in services/register/seed/example.json. The refused create does not change it.
const seededID = "c7b88455-4a14-479e-b53e-fbc4000716b3";
const idNotUnique = "legalEntityId is not unique: a different record has it";

type DataRecipients = { data: { legalEntityId: string; legalEntityName: string }[] };

function isDataRecipients(body: unknown): body is DataRecipients {
  return typeof body === "object" && body !== null && "data" in body && Array.isArray(body.data);
}

async function cdrName(request: APIRequestContext, id: string): Promise<string | undefined> {
  const response = await request.get(publicURL("/cdr-register/v1/all/data-recipients"), {
    headers: { "x-v": "4" },
  });
  expect(response.status()).toBe(200);
  const body: unknown = await response.json();
  if (!isDataRecipients(body)) throw new Error("Get Data Recipients gave no list");
  return body.data.find((d) => d.legalEntityId === id)?.legalEntityName;
}

// Each run makes a new data recipient with the UUID of the form, so the test passes again on the same lab. The name
// sorts after the seed data recipients that the other tests find on the first page of the list.
test("the operator creates a data recipient, and the CDR APIs and the history show it", async ({
  page,
  request,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));

  await page.goto("/data-recipients");
  await page.getByRole("link", { name: "Create data recipient" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Create a data recipient");
  const id = page.getByRole("textbox", { name: "Legal entity ID" });
  const generated = await id.inputValue();
  expect(generated).toMatch(
    /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
  );
  await expect(page.getByRole("combobox", { name: "Accreditation level" })).toHaveValue("");
  await expect(page.getByRole("combobox", { name: "Status" })).toHaveValue("ACTIVE");
  await expectNoAxeViolations(page);

  const name = `Operator Create Test ${generated.slice(0, 8)} Pty Ltd`;
  await page.getByRole("textbox", { name: "Name", exact: true }).fill(name);
  await page
    .getByRole("textbox", { name: "Accreditation number" })
    .fill(`ADR-CREATE-${generated.slice(0, 8)}`);
  await page.getByRole("combobox", { name: "Accreditation level" }).selectOption("SPONSORED");
  await page
    .getByRole("textbox", { name: "Logo URI" })
    .fill("https://create-test.example/logo.png");

  await id.fill(seededID);
  await page.getByRole("button", { name: "Create data recipient" }).click();

  await expect(page.getByRole("alert").getByRole("link", { name: idNotUnique })).toBeVisible();
  await expect(id).toHaveAccessibleDescription(new RegExp(`${idNotUnique}$`));
  await expectNoAxeViolations(page);

  await id.fill(generated);
  await page.getByRole("button", { name: "Create data recipient" }).click();

  await expect(page.getByRole("heading", { level: 1, name, exact: true })).toBeFocused();
  await expect(page).toHaveURL(`/data-recipients/${generated}`);
  await expect(page.getByRole("dialog", { name: `Data recipient ${name} created` })).toBeVisible();
  await expectNoAxeViolations(page);
  expect(await cdrName(request, generated)).toBe(name);
  expect(await cdrName(request, seededID)).toBe("Mock Budgeting Pty Ltd");

  await page.getByRole("link", { name: "History", exact: true }).click();
  await expect(
    page.getByRole("list", { name: "History" }).getByText("Created by the operator"),
  ).toBeVisible();
  expect(errors).toEqual([]);
});
