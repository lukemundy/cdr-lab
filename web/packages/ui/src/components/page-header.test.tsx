import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { Button } from "./button";
import { PageHeader } from "./page-header";

// Verifies: UI-0218
test("the page header gives the page title as the first heading, with its description and actions", async () => {
  const container = render(
    <PageHeader
      title="Data recipients"
      description="The accredited data recipients of the Register."
      actions={<Button>Refresh</Button>}
    />,
  );

  await expect
    .element(page.getByRole("heading", { level: 1, name: "Data recipients" }))
    .toBeVisible();
  await expect
    .element(page.getByText("The accredited data recipients of the Register."))
    .toBeVisible();
  await expect.element(page.getByRole("button", { name: "Refresh" })).toBeVisible();
  await expectNoAxeViolations(container);
});

test("the page header needs only a title", async () => {
  const container = render(<PageHeader title="Data holders" />);

  await expect.element(page.getByRole("heading", { level: 1, name: "Data holders" })).toBeVisible();
  await expectNoAxeViolations(container);
});

// Verifies: UI-0219
test("a page header can take the focus to its title when the page shows", async () => {
  const container = render(<PageHeader title="Mock Pty Ltd" focusTitle />);
  const title = page.getByRole("heading", { level: 1, name: "Mock Pty Ltd" });

  await expect.poll(() => document.activeElement).toBe(title.element());
  await expectNoAxeViolations(container);
});

test("a page header does not take the focus by default", async () => {
  render(<PageHeader title="Mock Pty Ltd" />);

  await expect.element(page.getByRole("heading", { level: 1 })).toBeVisible();
  expect(document.activeElement).toBe(document.body);
});

// Verifies: UI-0079, UI-0080
test("the page header is a band on the shell band colour, with the title in the heading colour", async () => {
  const container = render(<PageHeader title="Customers" />);

  const band = page.getByRole("heading", { level: 1 }).element().closest("header");
  expect(band?.className).toContain("bg-shell-band");
  expect(page.getByRole("heading", { level: 1 }).element().className).toContain("text-heading");
  await expectNoAxeViolations(container);
});
