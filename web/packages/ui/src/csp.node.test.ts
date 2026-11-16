import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import tailwindcss from "@tailwindcss/vite";
import { chromium, firefox, type Browser, type BrowserType, type Page } from "playwright";
import { build, preview, type PreviewServer } from "vite";
import { afterAll, beforeAll, describe, expect, test } from "vitest";

// The CSP of each web UI (pkg/platform/consolehttp/DESIGN.md, CNSL-0007).
const csp = [
  "default-src 'self'",
  "script-src 'self'",
  "style-src 'self'",
  "img-src 'self' data:",
  "font-src 'self'",
  "connect-src 'self'",
  "object-src 'none'",
  "base-uri 'none'",
  "form-action 'self'",
  "frame-ancestors 'none'",
  "require-trusted-types-for 'script'",
].join("; ");

const pem = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n";

const root = join(import.meta.dirname, "test-page");
// The browser tests run at the same time on the CI runner, so a step can take some seconds there.
const stepTimeout = 10_000;
const testTimeout = 30_000;
let outDir = "";
let server: PreviewServer | undefined;
let url = "";

// A production build, because the Vite development server adds CSS in style elements, which the CSP refuses.
beforeAll(async () => {
  outDir = await mkdtemp(join(tmpdir(), "ui-test-page-"));
  await build({
    root,
    configFile: false,
    // The test page is one chunk, and Base UI has "use client" directives: the build warns about both.
    logLevel: "error",
    plugins: [tailwindcss()],
    build: {
      outDir,
      emptyOutDir: true,
      // As the console build: no data: URIs, which the CSP permits for images only.
      assetsInlineLimit: 0,
      modulePreload: { polyfill: false },
    },
  });
  server = await preview({
    root,
    configFile: false,
    logLevel: "warn",
    build: { outDir },
    preview: { host: "127.0.0.1", port: 0, headers: { "Content-Security-Policy": csp } },
  });
  url = server.resolvedUrls?.local[0] ?? "";
}, 60_000);

afterAll(async () => {
  await server?.close();
  if (outDir) await rm(outDir, { recursive: true, force: true });
});

async function violationsOf(page: Page): Promise<unknown> {
  return page.evaluate(() => Reflect.get(window, "violations"));
}

// The init script runs before the scripts of the page, so it records each violation from the start of the load.
async function open(browser: Browser): Promise<{ page: Page; errors: string[] }> {
  const page = await browser.newPage();
  // A step that fails names itself before the test timeout.
  page.setDefaultTimeout(stepTimeout);
  await page.addInitScript(() => {
    const violations: string[] = [];
    Object.assign(window, { violations });
    document.addEventListener("securitypolicyviolation", (event) => {
      violations.push(`${event.effectiveDirective} ${event.blockedURI}`);
    });
  });
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  const response = await page.goto(url);
  expect(response?.headers()["content-security-policy"]).toBe(csp);
  try {
    await page.getByRole("heading", { name: "Design system test page" }).waitFor();
  } catch {
    throw new Error(
      `the test page did not show: ${JSON.stringify({ errors, violations: await violationsOf(page) })}`,
    );
  }
  return { page, errors };
}

describe.each([
  ["Chromium", chromium],
  ["Firefox", firefox],
])("in %s", (_name, browserType: BrowserType) => {
  let browser: Browser | undefined;

  beforeAll(async () => {
    browser = await browserType.launch();
  });

  afterAll(async () => {
    await browser?.close();
  });

  // Verifies: UI-0211
  test(
    "each component loads and works with no CSP or Trusted Types violation",
    async () => {
      if (!browser) throw new Error("no browser");
      const { page, errors } = await open(browser);

      await page.evaluate(() => document.fonts.ready);
      await page.keyboard.press("Tab");
      await page.keyboard.press("Enter");
      for (let i = 0; i < 30; i++) {
        // oxlint-disable-next-line no-await-in-loop -- each Tab key press starts from the focus that the previous one gave.
        await page.keyboard.press("Tab");
      }
      await page.getByRole("searchbox", { name: "Name" }).fill("Mock");
      await page.keyboard.press("Enter");
      await page.getByRole("combobox", { name: "Status" }).selectOption("ACTIVE");
      await page
        .getByRole("table", { name: "Records", exact: true })
        .getByRole("button", { name: "Name" })
        .click();
      await page.getByRole("status").filter({ hasText: "Sorted by Name, descending" }).waitFor();
      await page.getByRole("button", { name: "Next page" }).first().click();
      await page.getByRole("status").filter({ hasText: "Page 2 of 3" }).waitFor();
      const history = page.getByRole("region", { name: "History" });
      await history.getByRole("button", { name: "Next page" }).first().click();
      await history.getByRole("status").filter({ hasText: "Page 2 of 4" }).waitFor();
      await page.getByRole("button", { name: "Copy ID of Mock Recipient One" }).click();
      // A copy can fail in a headless browser, but each result shows a message.
      await page
        .getByRole("status")
        .filter({ hasText: /^(Copied|Could not copy)/ })
        .waitFor();

      await page.getByRole("textbox", { name: "Text field", exact: true }).fill("Mock");
      // The component test checks the keyboard, so this step only selects the file.
      await page.getByLabel("Certificate file").setInputFiles({
        name: "client.crt",
        mimeType: "text/plain",
        buffer: Buffer.from(pem),
      });
      const certificate = page.getByRole("textbox", { name: "Certificate", exact: true });
      await expect.poll(() => certificate.inputValue()).toBe(pem);

      const form = page.getByRole("form", { name: "Change the details" });
      await form.getByRole("checkbox", { name: "Energy" }).check();
      await form.getByRole("button", { name: "Save changes" }).click();
      await form.getByRole("status").filter({ hasText: "Save changes in progress" }).waitFor();
      const summary = form.getByRole("alert").filter({ hasText: "There is a problem" });
      await summary.getByRole("link", { name: "Choose a level." }).click();
      await expect
        .poll(() =>
          form
            .getByRole("combobox", { name: "Accreditation level" })
            .evaluate((e) => e === document.activeElement),
        )
        .toBe(true);
      await summary.getByRole("link", { name: "Choose an industry." }).click();
      await expect
        .poll(() =>
          form
            .getByRole("group", { name: "Industries" })
            .getByRole("checkbox", { name: "Banking" })
            .evaluate((e) => e === document.activeElement),
        )
        .toBe(true);
      await summary.getByRole("link", { name: "Enter a URI." }).click();
      await expect
        .poll(() =>
          form
            .getByRole("group", { name: "Endpoints" })
            .getByRole("textbox", { name: "Public base URI" })
            .evaluate((e) => e === document.activeElement),
        )
        .toBe(true);

      await page.getByRole("button", { name: "Open the dialog" }).click();
      await page.getByRole("dialog", { name: "Register a client certificate" }).waitFor();
      await page.getByRole("button", { name: "Close the dialog" }).click();
      await page.getByRole("dialog").waitFor({ state: "detached" });

      await page.getByRole("button", { name: "Suspend", exact: true }).click();
      const suspension = page.getByRole("alertdialog", { name: "Suspend Mock Recipient One?" });
      await suspension.getByRole("button", { name: "Suspend", exact: true }).click();
      await suspension.getByRole("status").filter({ hasText: "Suspend in progress" }).waitFor();
      await suspension.waitFor({ state: "detached" });
      await page.getByRole("dialog", { name: "Mock Recipient One is suspended" }).waitFor();

      await page.getByRole("button", { name: "Revoke", exact: true }).click();
      const revocation = page.getByRole("alertdialog", { name: "Revoke Mock Recipient One?" });
      await revocation.getByRole("listitem").first().waitFor();
      await revocation.getByRole("button", { name: "Revoke", exact: true }).click();
      await revocation.getByRole("alert").filter({ hasText: "Conflict" }).waitFor();
      await revocation.getByRole("button", { name: "Cancel" }).click();
      await revocation.waitFor({ state: "detached" });

      await page.getByRole("button", { name: "Remove", exact: true }).click();
      const removal = page.getByRole("alertdialog", { name: "Remove the certificate?" });
      await removal.getByRole("button", { name: "Remove", exact: true }).click();
      await removal.getByRole("alert").filter({ hasText: "The action did not complete" }).waitFor();
      await page.keyboard.press("Escape");
      await removal.waitFor({ state: "detached" });

      const consumer = page.getByRole("region", { name: "Consumer page" });
      await consumer.getByRole("link", { name: "Skip to content" }).focus();
      await page.keyboard.press("Enter");
      await consumer.getByRole("main").and(page.locator(":focus")).waitFor();

      expect(await violationsOf(page)).toEqual([]);
      expect(errors).toEqual([]);
    },
    testTimeout,
  );

  // Verifies: UI-0211
  test(
    "the page enforces Trusted Types",
    async () => {
      if (!browser) throw new Error("no browser");
      const { page } = await open(browser);

      const refused = await page.evaluate(() => {
        try {
          document.body.innerHTML = "<p>not trusted</p>";
          return false;
        } catch (error) {
          return error instanceof TypeError;
        }
      });

      expect(refused).toBe(true);
    },
    testTimeout,
  );
});
