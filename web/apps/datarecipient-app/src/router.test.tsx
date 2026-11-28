import { renderAppShell } from "@cdr-lab/web-platform/test/shell";
import { createMemoryHistory } from "@tanstack/react-router";
import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { createAppRouter } from "./router";

test("the app renders the main area of the page", async () => {
  renderAppShell(createAppRouter(createMemoryHistory({ initialEntries: ["/"] })));

  await expect.element(page.getByRole("main")).toBeInTheDocument();
});
