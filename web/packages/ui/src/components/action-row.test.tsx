import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { ActionRow } from "./action-row";
import { Button } from "./button";

// Verifies: UI-0246
test("an action row shows its buttons on one line, in their order", async () => {
  const container = render(
    <ActionRow>
      <Button variant="secondary">Suspend</Button>
      <Button variant="danger">Revoke</Button>
    </ActionRow>,
  );

  const suspend = page.getByRole("button", { name: "Suspend" }).element().getBoundingClientRect();
  const revoke = page.getByRole("button", { name: "Revoke" }).element().getBoundingClientRect();
  expect(revoke.left).toBeGreaterThan(suspend.right);
  expect(revoke.top).toBe(suspend.top);
  await expectNoAxeViolations(container);
});
