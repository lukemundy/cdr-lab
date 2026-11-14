import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import { uniqueName } from "@cdr-lab/console/e2e/dataholder-lab";

const emailError = "emailAddress must be one email address";

// The form makes the customer, so the test needs no data in the lab. All data is invented.
test("the operator creates a customer, and the data holder refuses a value first", async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));
  const name = uniqueName();
  const lastName = `Quokka${name}`;

  await page.goto("/customers");
  await page.getByRole("link", { name: "Create customer" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Create a customer");
  await expect(page).toHaveURL("/create/customer");
  await expectNoAxeViolations(page);

  await page.getByRole("textbox", { name: "First name (optional)" }).fill("Alex");
  await page.getByRole("textbox", { name: "Last name" }).fill(lastName);
  const email = page.getByRole("textbox", { name: "Email address" });
  await email.fill(`${name} at mail.test`);
  await page.getByRole("button", { name: "Create customer" }).click();

  await expect(page.getByRole("alert").getByRole("link", { name: emailError })).toBeVisible();
  await expect(email).toHaveAccessibleDescription(new RegExp(`${emailError}$`));
  await expect(email).toHaveValue(`${name} at mail.test`);
  await expect(page.getByRole("textbox", { name: "Last name" })).toHaveValue(lastName);
  await expectNoAxeViolations(page);

  await email.fill(`${name}@mail.test`);
  await page.getByRole("button", { name: "Create customer" }).click();

  await expect(page.getByRole("heading", { level: 1, name: `Alex ${lastName}` })).toBeFocused();
  await expect(page).toHaveURL(/\/customers\/[0-9]{8}$/);
  const customerNumber = new URL(page.url()).pathname.split("/").at(-1);
  await expect(
    page.getByRole("dialog", { name: `Customer ${customerNumber} created` }),
  ).toBeVisible();
  await expect(page.getByText(`${name}@mail.test`)).toBeVisible();
  await expectNoAxeViolations(page);
  expect(errors).toEqual([]);
});
