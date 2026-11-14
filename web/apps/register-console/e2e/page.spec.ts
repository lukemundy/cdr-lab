import { expect, test, type Page } from "@playwright/test";

type Violation = { directive: string; blockedURI: string };

declare global {
  interface Window {
    violations: Violation[];
  }
}

// The init script runs before the scripts of the page, so it records each violation of the page load.
async function recordViolations(page: Page): Promise<void> {
  await page.addInitScript(() => {
    window.violations = [];
    document.addEventListener("securitypolicyviolation", (event) => {
      window.violations.push({ directive: event.effectiveDirective, blockedURI: event.blockedURI });
    });
  });
}

async function violationsOf(page: Page): Promise<Violation[]> {
  return page.evaluate(() => window.violations);
}

const pages = [
  { path: "/", heading: "Data recipients" },
  { path: "/data-holders", heading: "Data holders" },
  { path: "/signing-keys", heading: "Signing keys" },
  // The seed data recipient and data holder, whose pages show the logo URIs of the seed file.
  {
    path: "/data-recipients/c7b88455-4a14-479e-b53e-fbc4000716b3",
    heading: "Mock Budgeting Pty Ltd",
  },
  { path: "/data-holders/a81975c6-1ac9-4fe4-88fc-68f50efd27cb", heading: "Mock Holdings Ltd" },
  {
    path: "/data-recipients/c7b88455-4a14-479e-b53e-fbc4000716b3/history",
    heading: "History of Mock Budgeting Pty Ltd",
  },
  {
    path: "/data-holders/a81975c6-1ac9-4fe4-88fc-68f50efd27cb/history",
    heading: "History of Mock Holdings Ltd",
  },
  { path: "/an-unknown-route", heading: "Page not found" },
];

for (const { path, heading } of pages) {
  test(`${path} loads with no CSP or Trusted Types violation`, async ({ page }) => {
    await recordViolations(page);
    const errors: Error[] = [];
    page.on("pageerror", (error) => errors.push(error));

    const response = await page.goto(path);

    expect(response?.status()).toBe(200);
    expect(response?.headers()["content-security-policy"]).toContain(
      "require-trusted-types-for 'script'",
    );
    // The heading shows that the scripts of the UI ran under the CSP.
    await expect(page.getByRole("heading", { level: 1, name: heading })).toBeVisible();
    expect(await violationsOf(page)).toEqual([]);
    expect(errors).toEqual([]);
  });
}

test("the page enforces Trusted Types", async ({ page }) => {
  await page.goto("/");

  const refused = await page.evaluate(() => {
    try {
      document.body.innerHTML = "<p>not trusted</p>";
      return false;
    } catch (error) {
      return error instanceof TypeError;
    }
  });

  expect(refused).toBe(true);
});

test("the header and the browser tab title say CDR Register", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle("CDR Register");
  await expect(page.getByRole("banner")).toHaveText("CDR Register");
});
