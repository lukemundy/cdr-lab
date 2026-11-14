import tailwindcss from "@tailwindcss/vite";
import { playwright } from "@vitest/browser-playwright";
import { defineConfig } from "vitest/config";

const nodeTests = "{apps,packages}/*/src/**/*.node.test.ts";
const browserTests = "{apps,packages}/*/src/**/*.test.{ts,tsx}";
// Each app has its own project, with the Vite config of the app.
const consoleTests = "apps/{register-console,dataholder-console,dataholder-bank}/**";
const localTimeTests = "packages/ui/src/components/{local-time,date-time-field}.test.tsx";

// Each browser test has a fixed time zone, so that a time shows the same in each place where a test runs.
const testTimeZone = "Australia/Perth";
const inTimeZone = (timezoneId: string) => playwright({ contextOptions: { timezoneId } });

// A test that needs Node, for example to read files, has the name *.node.test.ts.
export default defineConfig({
  test: {
    projects: [
      {
        plugins: [tailwindcss()],
        // A dependency that Vite finds after the tests start reloads the browser and fails a test, so each run finds them first, also with a cache.
        optimizeDeps: { entries: [browserTests, `!${nodeTests}`, `!${consoleTests}`], force: true },
        test: {
          name: "browser",
          include: [browserTests],
          exclude: [nodeTests, consoleTests],
          // A wait for a response and a render can take some seconds on a busy runner.
          expect: { poll: { timeout: 5000 } },
          // WebKit is a supported browser, but CI does not test it (packages/ui/DESIGN.md, Decisions).
          browser: {
            enabled: true,
            headless: true,
            provider: inTimeZone(testTimeZone),
            instances: [
              { browser: "chromium" },
              { browser: "firefox" },
              ...["UTC", "Australia/Sydney", "America/St_Johns"].map((timezoneId) => ({
                browser: "chromium" as const,
                name: `chromium ${timezoneId}`,
                provider: inTimeZone(timezoneId),
                include: [localTimeTests],
              })),
            ],
          },
        },
      },
      "apps/register-console",
      "apps/dataholder-console",
      "apps/dataholder-bank",
      {
        test: {
          name: "node",
          include: [nodeTests],
          environment: "node",
        },
      },
    ],
  },
});
