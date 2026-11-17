import { readdir, readFile } from "node:fs/promises";
import { join } from "node:path";

import { expect, test } from "vitest";

const appsDir = join(import.meta.dirname, "../../../apps");
const apps = await readdir(appsDir);

// Verifies: UI-0141
test.each(apps)("the page of %s has lang en-AU", async (app) => {
  const page = await readFile(join(appsDir, app, "index.html"), "utf8");

  expect(page).toMatch(/<html lang="en-AU"[ >]/);
});
