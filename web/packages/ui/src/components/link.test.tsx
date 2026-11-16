import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectVisibleFocus } from "../test/checks";
import { Pencil } from "lucide-react";

import { render } from "../test/render";
import { Link } from "./link";

// SC 2.5.8 has an exception for a link in a sentence, so these tests do not check the target size of a link.

test("a link in text passes the accessibility checks", async () => {
  const container = render(
    <p>
      See the <Link href="#history">history</Link> of the record.
    </p>,
  );

  await expectNoAxeViolations(container);
});

// Verifies: UI-0225
test("a link is underlined, so that it is different from text also without colour", () => {
  render(<Link href="#history">history</Link>);

  const style = getComputedStyle(page.getByRole("link", { name: "history" }).element());
  expect(style.textDecorationLine).toContain("underline");
});

// Verifies: UI-0201, UI-0202
test("the keyboard focuses the link with a visible focus, and Enter follows it", async () => {
  render(<Link href="#followed">history</Link>);

  await userEvent.tab();
  expectVisibleFocus(page.getByRole("link", { name: "history" }).element());
  await userEvent.keyboard("{Enter}");

  expect(window.location.hash).toBe("#followed");
  window.history.replaceState(null, "", window.location.pathname + window.location.search);
});

test("a router can give the element of the link, and the link keeps its style", () => {
  render(
    <Link
      href="#routed"
      render={({ children, ...props }) => (
        <a data-router="" {...props}>
          {children}
        </a>
      )}
    >
      history
    </Link>,
  );

  const link = page.getByRole("link", { name: "history" }).element();
  expect(link).toHaveAttribute("data-router");
  expect(getComputedStyle(link).textDecorationLine).toContain("underline");
});

// Verifies: UI-0226, UI-0230
test("the icon of a link is before its text, is not underlined, and assistive technology does not announce it", async () => {
  const container = render(
    <Link href="#change" icon={Pencil}>
      Change brand
    </Link>,
  );
  const link = page.getByRole("link", { name: "Change brand" }).element();

  const icon = link.querySelector("svg");
  expect(icon).toHaveAttribute("aria-hidden", "true");
  expect(link.firstElementChild).toBe(icon);
  expect(getComputedStyle(link).textDecorationLine).not.toContain("underline");
  expect(getComputedStyle(page.getByText("Change brand").element()).textDecorationLine).toContain(
    "underline",
  );
  await expectNoAxeViolations(container);
});
