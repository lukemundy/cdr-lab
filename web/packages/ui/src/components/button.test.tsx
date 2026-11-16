import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { Pencil } from "lucide-react";

import { render } from "../test/render";
import { Button } from "./button";

describe.each(["primary", "secondary", "danger", "quiet"] as const)("the %s button", (variant) => {
  // Verifies: UI-0220, UI-0221
  test("passes the accessibility checks when enabled, disabled and busy", async () => {
    const container = render(
      <>
        <Button variant={variant}>Save</Button>
        <Button variant={variant} disabled>
          Save
        </Button>
        <Button variant={variant} busy>
          Save
        </Button>
      </>,
    );

    await expectNoAxeViolations(container);
  });

  // Verifies: UI-0094
  test("is 24 by 24 CSS px or more, also with only an icon", () => {
    render(
      <>
        <Button variant={variant}>OK</Button>
        <Button variant={variant} size="icon" aria-label="Refresh">
          <svg aria-hidden="true" />
        </Button>
      </>,
    );

    expectTargetSize(page.getByRole("button", { name: "OK" }).element());
    expectTargetSize(page.getByRole("button", { name: "Refresh" }).element());
  });
});

// Verifies: UI-0201, UI-0202
test("the keyboard focuses the button with a visible focus, and Enter and Space activate it", async () => {
  let clicks = 0;
  render(<Button onClick={() => clicks++}>Save</Button>);

  await userEvent.tab();
  expectVisibleFocus(page.getByRole("button", { name: "Save" }).element());
  await userEvent.keyboard("{Enter}");
  await userEvent.keyboard(" ");

  expect(clicks).toBe(2);
});

// Verifies: UI-0221
test("a disabled button is disabled for assistive technology", async () => {
  render(<Button disabled>Save</Button>);

  await expect.element(page.getByRole("button", { name: "Save" })).toBeDisabled();
});

// Verifies: UI-0221
test("a busy button says that it is busy, keeps its focus and does not act", async () => {
  let clicks = 0;
  render(
    <Button busy onClick={() => clicks++}>
      Save
    </Button>,
  );
  const button = page.getByRole("button", { name: "Save (busy)" });

  await expect.element(button).toHaveAttribute("aria-disabled", "true");
  await expect.element(button).toHaveAttribute("aria-busy", "true");
  await userEvent.tab();
  expectVisibleFocus(button.element());
  await userEvent.keyboard("{Enter}");

  expect(clicks).toBe(0);
});

test("a button does not submit a form unless its type is submit", async () => {
  let submits = 0;
  render(
    <form
      onSubmit={(event) => {
        event.preventDefault();
        submits++;
      }}
    >
      <Button>Other</Button>
      <Button type="submit">Send</Button>
    </form>,
  );

  await page.getByRole("button", { name: "Other" }).click();
  await page.getByRole("button", { name: "Send" }).click();

  expect(submits).toBe(1);
});

// Verifies: UI-0222, UI-0230
test("the icon of a button is before its text, and assistive technology does not announce it", async () => {
  const container = render(<Button icon={Pencil}>Change</Button>);
  const button = page.getByRole("button", { name: "Change" });

  const icon = button.element().querySelector("svg");
  expect(icon).toHaveAttribute("aria-hidden", "true");
  expect(button.element().firstElementChild).toBe(icon);
  await expectNoAxeViolations(container);
});

// Verifies: UI-0222
test("a busy button shows the busy indicator in place of its icon", () => {
  render(
    <Button icon={Pencil} busy>
      Change
    </Button>,
  );

  const icons = page
    .getByRole("button", { name: "Change (busy)" })
    .element()
    .querySelectorAll("svg");
  expect(icons).toHaveLength(1);
  expect(icons[0]).toHaveClass("motion-safe:animate-spin");
});

// Verifies: UI-0223
test("an enabled button shows the pointer cursor", () => {
  render(<Button>Save</Button>);

  expect(getComputedStyle(page.getByRole("button", { name: "Save" }).element()).cursor).toBe(
    "pointer",
  );
});
