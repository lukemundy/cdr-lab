import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";

// Mock Holdings Ltd in services/register/seed/example.json. The refused create does not change it.
const seededID = "a81975c6-1ac9-4fe4-88fc-68f50efd27cb";
const idNotUnique = "legalEntityId is not unique: a different record has it";

// Each run makes a new data holder with the UUID of the form, so the test passes again on the same lab. The names sort
// after the seed data holders that other tests find on the first page, and fit the limit of 36 characters.
test("the operator creates a data holder and changes its details, and the history shows them", async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));

  await page.goto("/data-holders");
  await page.getByRole("link", { name: "Create data holder" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Create a data holder");
  const id = page.getByRole("textbox", { name: "Legal entity ID" });
  const generated = await id.inputValue();
  expect(generated).toMatch(
    /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
  );
  await expect(page.getByRole("combobox", { name: "Status" })).toHaveValue("ACTIVE");
  await expectNoAxeViolations(page);

  const name = `Test Create Holder ${generated.slice(0, 8)} Ltd`;
  await page.getByRole("textbox", { name: "Name", exact: true }).fill(name);
  await page
    .getByRole("textbox", { name: "Logo URI" })
    .fill("https://create-holder-test.example/logo.png");

  await id.fill(seededID);
  await page.getByRole("button", { name: "Create data holder" }).click();

  await expect(page.getByRole("alert").getByRole("link", { name: idNotUnique })).toBeVisible();
  await expect(id).toHaveAccessibleDescription(new RegExp(`${idNotUnique}$`));
  await expectNoAxeViolations(page);

  await id.fill(generated);
  await page.getByRole("button", { name: "Create data holder" }).click();

  await expect(page.getByRole("heading", { level: 1, name, exact: true })).toBeFocused();
  await expect(page).toHaveURL(`/data-holders/${generated}`);
  await expect(page.getByRole("dialog", { name: `Data holder ${name} created` })).toBeVisible();
  await expectNoAxeViolations(page);

  const changedName = `Test Changed Holder ${generated.slice(0, 8)} Ltd`;
  await page
    .getByRole("region", { name: "Legal entity", exact: true })
    .getByRole("link", { name: "Change details" })
    .click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(`Change the details of ${name}`);
  const nameField = page.getByRole("textbox", { name: "Name", exact: true });
  await expect(nameField).toHaveValue(name);
  await nameField.fill(changedName);
  await page.getByRole("button", { name: "Save changes" }).click();

  await expect(
    page.getByRole("heading", { level: 1, name: changedName, exact: true }),
  ).toBeFocused();
  await expect(page).toHaveURL(`/data-holders/${generated}`);
  await expect(
    page.getByRole("dialog", { name: `Details of ${changedName} changed` }),
  ).toBeVisible();
  await expectNoAxeViolations(page);

  await page.getByRole("link", { name: "History", exact: true }).click();
  const history = page.getByRole("list", { name: "History" });
  await expect(history.getByText("Legal entity details changed")).toBeVisible();
  await expect(history.getByText("Created by the operator")).toBeVisible();
  expect(errors).toEqual([]);
});
