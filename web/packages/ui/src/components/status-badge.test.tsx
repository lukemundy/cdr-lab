import { describe, expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { StatusBadge } from "./status-badge";

describe.each(["neutral", "success", "warning", "danger", "info"] as const)(
  "the %s status badge",
  (tone) => {
    test("shows its status text and passes the accessibility checks", async () => {
      const container = render(<StatusBadge tone={tone}>SUSPENDED</StatusBadge>);

      await expect.element(page.getByText("SUSPENDED")).toBeVisible();
      await expectNoAxeViolations(container);
    });
  },
);
