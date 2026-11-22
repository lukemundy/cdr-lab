import { expect, test, type Page } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/app/e2e/checks";
import {
  createAccount,
  createCustomer,
  uniqueAccountNumber,
  uniqueName,
} from "@cdr-lab/app/e2e/dataholder-lab";

// The brandName of the Wattle Bank brand in services/register/seed/example.json, which compose.yaml gives.
const brandName = "Wattle Bank";

// Verifies: DH-1466, DH-1468, DH-1470, DH-1472, DH-1485
test("the console gives the brand in the shell and the tab", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByRole("heading", { level: 1, name: "Overview" })).toBeVisible();
  const banner = page.getByRole("banner");
  const logo = banner.getByRole("img", { name: brandName });
  await expect(logo).toHaveAttribute("src", "/brand/logo.svg");
  // The CSP of the console permits the logo, so the browser loads it.
  await expect
    .poll(() => logo.evaluate((image: HTMLImageElement) => image.complete && image.naturalWidth))
    .toBeGreaterThan(0);
  await expect(banner).toHaveText("Operator console");
  await expect(page).toHaveTitle(brandName);

  const styles = await page.request.get("/brand/console.css");
  expect(styles.status()).toBe(200);
  const css = await styles.text();
  await expect(banner).toHaveCSS("background-color", await computedColour(page, css, "brand"));
  const navigation = page.getByRole("navigation", { name: "Main" });
  await expect(navigation).toHaveCSS("background-color", await computedColour(page, css, "hero"));
  await expect(navigation).not.toHaveCSS("background-color", "rgb(255, 255, 255)");

  const icon = page.locator('link[rel="icon"]');
  await expect(icon).toHaveAttribute("href", "/brand/mark.svg");
  const mark = await page.request.get("/brand/mark.svg");
  expect(mark.status()).toBe(200);
  expect(mark.headers()["content-type"]).toBe("image/svg+xml");

  await expectNoAxeViolations(page);
});

// Verifies: DH-1415
test("the operator finds a customer by a part of each field, with no difference for case", async ({
  page,
}) => {
  const name = uniqueName();
  const found = await createCustomer({
    firstName: `Ari${name}`,
    lastName: `Quokka${name}`,
    emailAddress: `${name}.found@mail.test`,
  });
  await createCustomer({ lastName: `Wombat${name}`, emailAddress: `${name}.other@mail.test` });
  await page.goto("/customers");
  const searchbox = page.getByRole("searchbox", {
    name: "Search by customer number, name or email",
  });
  const table = page.getByRole("table", { name: "Customers" });

  // A part of the customer number can also be in the numbers of the customers of other tests.
  const partOfNumber = found.customerNumber.slice(2, 7);
  async function search(text: string, { onlyFound = true } = {}) {
    await searchbox.fill(text);

    await expect(page).toHaveURL(`/customers?q=${encodeURIComponent(text)}`);
    await expect(
      table.getByRole("cell", { name: found.customerNumber, exact: true }),
    ).toBeVisible();
    if (onlyFound) await expect(table.getByRole("row")).toHaveCount(2);
  }

  await search(partOfNumber, { onlyFound: false });
  await search(`ARI${name}`.toUpperCase());
  await search(`kka${name}`);
  await search(`${name}.FOUND@`);

  await page.reload();

  await expect(searchbox).toHaveValue(`${name}.FOUND@`);
  await expect(table.getByRole("cell", { name: found.emailAddress, exact: true })).toBeVisible();
  await expect(
    table.getByRole("cell", { name: `Quokka${name}, Ari${name}`, exact: true }),
  ).toBeVisible();
});

// Verifies: WCON-0001
test("the search, the sort and the page of the list stay the same after a reload and in a link", async ({
  page,
  browser,
}) => {
  const name = uniqueName();
  const created = await Promise.all(
    Array.from({ length: 26 }, (_, i) =>
      createCustomer({ lastName: `Numbat${name}`, emailAddress: `${name}.${i}@mail.test` }),
    ),
  );
  const numbers = created.map((customer) => customer.customerNumber);
  // The last page of a sort by the customer number, newest first, has the lowest number.
  const lowest = numbers.toSorted()[0]!;
  const table = page.getByRole("table", { name: "Customers" });

  await page.goto(`/customers?q=${name}`);
  await table.getByRole("button", { name: "Customer number" }).click();
  await table.getByRole("button", { name: "Customer number" }).click();
  await page.getByRole("button", { name: "Next page" }).click();

  const url = `/customers?q=${name}&sort=customerNumber&order=desc&page=2`;
  await expect(page).toHaveURL(url);
  await expect(page.getByText("Page 2 of 2")).toBeVisible();
  await expect(table.getByRole("row")).toHaveCount(2);
  await expect(table.getByRole("cell", { name: lowest, exact: true })).toBeVisible();

  await page.reload();

  await expect(page.getByText("Page 2 of 2")).toBeVisible();
  await expect(table.getByRole("columnheader", { name: "Customer number" })).toHaveAttribute(
    "aria-sort",
    "descending",
  );
  await expect(table.getByRole("cell", { name: lowest, exact: true })).toBeVisible();

  const linked = await browser.newPage();
  await linked.goto(new URL(url, page.url()).toString());

  const linkedTable = linked.getByRole("table", { name: "Customers" });
  await expect(
    linked.getByRole("searchbox", { name: "Search by customer number, name or email" }),
  ).toHaveValue(name);
  await expect(linked.getByText("Page 2 of 2")).toBeVisible();
  await expect(linkedTable.getByRole("cell", { name: lowest, exact: true })).toBeVisible();
  await linked.close();
});

test("the operator sorts the customers by the number of open accounts, in both directions", async ({
  page,
}) => {
  const name = uniqueName();
  const created = await Promise.all(
    [0, 1, 2].map((i) =>
      createCustomer({ lastName: `Bilby${name}`, emailAddress: `${name}.${i}@mail.test` }),
    ),
  );
  // The customer at index i gets i open accounts.
  await Promise.all(
    created.flatMap((customer, i) =>
      Array.from({ length: i }, () =>
        createAccount(customer.customerNumber, {
          displayName: "Everyday",
          productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
          productName: "Everyday Account",
          bsb: "372100",
          accountNumber: uniqueAccountNumber(),
        }),
      ),
    ),
  );
  const numbers = created.map((customer) => customer.customerNumber);
  const table = page.getByRole("table", { name: "Customers" });
  const rows = () => table.locator("tbody tr td:first-child").allTextContents();

  await page.goto(`/customers?q=${name}`);
  await expect(table.getByRole("row")).toHaveCount(4);

  await table.getByRole("button", { name: "Open accounts" }).click();
  await expect(page).toHaveURL(`/customers?q=${name}&sort=openAccountCount&order=asc`);
  await expect(table.getByRole("columnheader", { name: "Open accounts" })).toHaveAttribute(
    "aria-sort",
    "ascending",
  );
  await expect.poll(rows).toEqual(numbers);

  await table.getByRole("button", { name: "Open accounts" }).click();
  await expect(page).toHaveURL(`/customers?q=${name}&sort=openAccountCount&order=desc`);
  await expect.poll(rows).toEqual(numbers.toReversed());
});

// Verifies: DH-1479
test("the list of customers passes the automated accessibility checks", async ({ page }) => {
  const name = uniqueName();
  await createCustomer({
    firstName: "Ari",
    lastName: `Quokka${name}`,
    emailAddress: `${name}@mail.test`,
  });

  await page.goto(`/customers?q=${name}`);
  await expect(
    page.getByRole("table", { name: "Customers" }).getByRole("cell", { name: `${name}@mail.test` }),
  ).toBeVisible();

  await expectNoAxeViolations(page);
});

// The computed colour of a role of console.css, as the browser gives it.
async function computedColour(page: Page, css: string, role: string): Promise<string> {
  const value = new RegExp(`--ds-color-${role}: ([^;]+);`).exec(css)?.[1];
  expect(value, role).toBeDefined();
  return page.evaluate((colour) => {
    const probe = document.createElement("div");
    probe.style.backgroundColor = colour;
    document.body.append(probe);
    const computed = getComputedStyle(probe).backgroundColor;
    probe.remove();
    return computed;
  }, value ?? "");
}
