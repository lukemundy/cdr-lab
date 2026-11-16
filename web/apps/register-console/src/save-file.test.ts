import { afterEach, expect, test, vi } from "vitest";

import { saveFile } from "./save-file";
import { captureSaves } from "./test/saves";

afterEach(() => {
  vi.restoreAllMocks();
});

// Verifies: REG-1247
test("a save gives the text with no change in a file with the name", async () => {
  const saves = captureSaves();
  const text = "-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n";

  saveFile("client-certificate-sp-1.pem", text);

  expect(saves).toHaveLength(1);
  expect(saves[0]?.name).toBe("client-certificate-sp-1.pem");
  expect(await saves[0]?.text).toBe(text);
});
