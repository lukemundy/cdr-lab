import { playwright } from "@vitest/browser-playwright";
import { defineConfig, mergeConfig } from "vitest/config";

import app from "./vite.config";

// The tests use the Vite config of the app, so that the components run with the React Compiler, as in the build.
export default mergeConfig(
  app,
  defineConfig({
    // Each run finds the dependencies before the tests start, for the reason in the workspace config.
    optimizeDeps: { entries: ["src/**/*.test.tsx"], force: true },
    test: {
      name: "datarecipient-app",
      include: ["src/**/*.test.{ts,tsx}"],
      // A wait for a response and a render can take some seconds on a busy runner.
      expect: { poll: { timeout: 5000 } },
      browser: {
        enabled: true,
        headless: true,
        // A fixed time zone, as in the workspace config.
        provider: playwright({ contextOptions: { timezoneId: "Australia/Perth" } }),
        instances: [{ browser: "chromium" }, { browser: "firefox" }],
      },
    },
  }),
);
