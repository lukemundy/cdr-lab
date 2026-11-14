import { randomInt, randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";

// The journey makes its own customer, because the lab has no seed customers. A new customer has no consent, so the
// test runs also in CI, where the lab has no conformance data.
test("a new customer opens Data sharing from the header and sees that they share no data", async ({
  page,
}) => {
  await page.goto("/sign-up");
  await page.getByLabel("Last name").fill(`Quoll ${randomInt(1_000_000)}`);
  await page.getByLabel("Email address").fill(`quoll.${randomUUID()}@example.test`);
  await page.getByRole("button", { name: "Sign up" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Your customer number" })).toBeVisible();

  await page
    .getByRole("navigation", { name: "Main" })
    .getByRole("link", { name: "Data sharing" })
    .click();

  await expect(page.getByRole("heading", { level: 1, name: "Data sharing" })).toBeVisible();
  await expect(page).toHaveTitle(/^Data sharing - .+ online banking$/);
  await expect(
    page.getByText("You do not share your data under the Consumer Data Right."),
  ).toBeVisible();
  await expectNoAxeViolations(page);
});
