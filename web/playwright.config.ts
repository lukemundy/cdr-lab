import { defineConfig } from "@playwright/test";

// The journeys of each UI against the running lab deployment. mise run test:lab-web sets the URLs.
function requiredURL(name: string): string {
  const url = process.env[name];
  if (!url) {
    throw new Error(`Set ${name} to the URL of the UI in the running lab.`);
  }
  return url;
}

const dataholderConsole = {
  browserName: "chromium",
  baseURL: requiredURL("DATAHOLDER_INTERNAL_URL"),
  timezoneId: "Australia/Perth",
} as const;

export default defineConfig({
  testDir: "apps",
  testMatch: "*/e2e/**/*.spec.ts",
  forbidOnly: true,
  reporter: "list",
  projects: [
    {
      name: "register-console",
      testMatch: "register-console/e2e/**/*.spec.ts",
      // A fixed time zone, so that the times show the same on each computer.
      use: {
        browserName: "chromium",
        baseURL: requiredURL("REGISTER_INTERNAL_URL"),
        timezoneId: "Australia/Perth",
      },
    },
    {
      name: "dataholder-console",
      testMatch: "dataholder-console/e2e/**/*.spec.ts",
      testIgnore: "dataholder-console/e2e/**/*.serial.spec.ts",
      use: dataholderConsole,
    },
    {
      // These tests count the records of the data holder, so they run after each test that creates records there.
      name: "dataholder-console-serial",
      testMatch: "dataholder-console/e2e/**/*.serial.spec.ts",
      dependencies: ["dataholder-console", "dataholder-bank"],
      use: dataholderConsole,
    },
    {
      name: "dataholder-bank",
      testMatch: "dataholder-bank/e2e/**/*.spec.ts",
      use: {
        browserName: "chromium",
        baseURL: requiredURL("BANK_SITE_URL"),
        timezoneId: "Australia/Perth",
        viewport: { width: 360, height: 800 },
        // Chromium does not read NODE_EXTRA_CA_CERTS, so it cannot verify the certificate of the lab public CA. The
        // calls from Node to the CDS APIs verify it.
        ignoreHTTPSErrors: true,
      },
    },
    {
      name: "datarecipient-app",
      testMatch: "datarecipient-app/e2e/**/*.spec.ts",
      // The web app is mobile-first (services/datarecipient/design/web-app.md).
      use: {
        browserName: "chromium",
        baseURL: requiredURL("DATARECIPIENT_WEB_APP_URL"),
        timezoneId: "Australia/Perth",
        viewport: { width: 360, height: 800 },
        // As for the online bank site: Chromium cannot verify the certificate of the lab public CA.
        ignoreHTTPSErrors: true,
      },
    },
  ],
});
