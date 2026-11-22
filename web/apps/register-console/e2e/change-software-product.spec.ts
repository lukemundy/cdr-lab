import { expect, test, type Page } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/web-platform/e2e/checks";
import { showAllBrands } from "./lab";

// Only this test changes this record of services/register/seed/example.json, and it gives back its redirect URIs.
const recipientId = "38ec0d0d-dde5-45c2-a3e1-7db8994a50d5";
const productName = "Console Product Change Test Web";
const seededURI = "https://console-product-change-test.example/callback";
const newURIs = [
  "https://console-product-change-test.example/new",
  "https://console-product-change-test.example/second",
];

const product = (page: Page) => page.getByRole("region", { name: productName, exact: true });
const redirectURIs = (page: Page) => page.getByRole("textbox", { name: "Redirect URIs" });

async function openForm(page: Page) {
  await page.goto(`/data-recipients/${recipientId}`);
  await showAllBrands(page);
  await product(page).getByRole("link", { name: "Change software product" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    `Change the details of ${productName}`,
  );
}

async function saveRedirectURIs(page: Page, text: string) {
  await redirectURIs(page).fill(text);
  await page.getByRole("button", { name: "Save changes" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toBeFocused();
}

test("the operator changes the redirect URIs of a software product, and the detail and the history show it", async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on("pageerror", (error) => errors.push(error));

  await openForm(page);
  await expect(redirectURIs(page)).toHaveValue(seededURI);
  await redirectURIs(page).fill(" \n");
  await page.getByRole("button", { name: "Save changes" }).click();

  await expect(
    page.getByRole("alert").getByRole("link", { name: "redirectUris is required" }),
  ).toBeVisible();
  await expect(redirectURIs(page)).toHaveAccessibleDescription(/redirectUris is required/);
  await expectNoAxeViolations(page);

  await saveRedirectURIs(page, ` ${newURIs[0]}\n\n${newURIs[1]} \n`);

  await expect(
    page.getByRole("dialog", { name: `Details of ${productName} changed` }),
  ).toBeVisible();
  await Promise.all(
    newURIs.map((uri) => expect(product(page).getByText(uri, { exact: true })).toBeVisible()),
  );
  await expect(product(page).getByText(seededURI, { exact: true })).toHaveCount(0);
  await expectNoAxeViolations(page);
  await page.getByRole("link", { name: "History", exact: true }).click();
  await expect(
    page
      .getByRole("list", { name: "History" })
      .getByText("Software product details changed")
      .first(),
  ).toBeVisible();

  await openForm(page);
  await expect(redirectURIs(page)).toHaveValue(newURIs.join("\n"));
  await saveRedirectURIs(page, seededURI);
  await expect(product(page).getByText(seededURI, { exact: true })).toBeVisible();
  expect(errors).toEqual([]);
});
