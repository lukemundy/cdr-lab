import type { AnyRouter } from "@tanstack/react-router";
import { render } from "@cdr-lab/ui/test/render";
import { afterEach, beforeAll, beforeEach } from "vitest";
import { page } from "vitest/browser";

import { AppShell } from "../shell";
import { createQueryClient } from "../query";

// Each console makes its own MSW worker, because the types of a second copy of msw do not agree.
export type MockWorker = {
  start(options: { quiet: boolean }): Promise<unknown>;
  resetHandlers(): void;
};

export function setUpConsoleTests(worker: MockWorker) {
  // No worker.stop() after the file: in Firefox, it also stops the mocks of the next test file.
  beforeAll(async () => {
    await worker.start({ quiet: true });
  });
  // A console is for desktop browsers, at 1280 CSS px or more (the Presentation of the console in each DESIGN.md).
  beforeEach(async () => {
    await page.viewport(1280, 800);
  });
  afterEach(() => worker.resetHandlers());
}

export function renderAppShell<Router extends AnyRouter>(router: Router): Router {
  render(<AppShell router={router} queryClient={createQueryClient()} />);
  return router;
}

export function definitionOf(container: Element, term: string): HTMLElement {
  const dt = [...container.querySelectorAll("dt")].find((element) => element.textContent === term);
  const dd = dt?.nextElementSibling;
  if (!(dd instanceof HTMLElement) || dd.tagName !== "DD") throw new Error(`no value for ${term}`);
  return dd;
}
