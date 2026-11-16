import { expect, test } from "@playwright/test";

import { expectNoAxeViolations } from "@cdr-lab/console/e2e/checks";
import {
  createAccount,
  createCustomer,
  uniqueAccountNumber,
  uniqueName,
} from "@cdr-lab/console/e2e/dataholder-lab";

// The four preset banks of compose.yaml, each with its own block of ports. The lab runs only wattle-bank by default;
// mise run up --all-banks starts the others. A bank that does not run is skipped.
const presets = [
  { name: "wattle-bank", block: 81 },
  { name: "redgum-bank", block: 82 },
  { name: "harbour-bank", block: 83 },
  { name: "kestrel-bank", block: 84 },
];

const navigationColours = new Map<string, string>();

async function runs(url: string): Promise<boolean> {
  try {
    return (await fetch(url)).ok;
  } catch {
    return false;
  }
}

for (const { name, block } of presets) {
  const consoleURL = `http://localhost:${block}84`;
  const operatorURL = `http://localhost:${block}82`;

  // Verifies: DH-1479
  test(`each page of the console of ${name} passes the automated accessibility checks`, async ({
    page,
  }) => {
    test.skip(!(await runs(`${consoleURL}/api/brand`)), `${name} does not run in this lab`);

    const customer = await createCustomer(
      { lastName: uniqueName(), emailAddress: `${uniqueName()}@mail.test` },
      operatorURL,
    );
    const accountId = await createAccount(
      customer.customerNumber,
      {
        displayName: "Everyday",
        productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
        productName: "Lab Everyday Account",
        bsb: "372100",
        accountNumber: uniqueAccountNumber(),
      },
      operatorURL,
    );

    // One tab shows the pages one after the other.
    /* oxlint-disable no-await-in-loop */
    for (const path of [
      "/",
      "/customers",
      `/customers/${customer.customerNumber}`,
      `/customers/${customer.customerNumber}/accounts/${accountId}/transactions`,
      "/registrations",
      "/register-statuses",
      `/create/customer`,
    ]) {
      await page.goto(consoleURL + path);
      await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
      await expectNoAxeViolations(page);
    }
    /* oxlint-enable no-await-in-loop */

    const navigation = page.getByRole("navigation", { name: "Main" });
    const colour = await navigation.evaluate(
      (element) => getComputedStyle(element).backgroundColor,
    );
    expect(colour).not.toBe("rgb(255, 255, 255)");
    navigationColours.set(name, colour);
  });
}

// With the logo hidden, the colour of the side navigation tells the banks apart.
// Verifies: DH-1470
test("the side navigation of each running bank has a colour of its own", () => {
  test.skip(navigationColours.size < 2, "fewer than two banks run in this lab");
  expect(new Set(navigationColours.values()).size).toBe(navigationColours.size);
});
