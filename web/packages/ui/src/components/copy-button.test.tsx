import { afterEach, expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { CopyButton } from "./copy-button";

const id = "a8f1c3d2-6b1e-4c5a-9f0e-2d7b3c4e5f60";

afterEach(() => {
  vi.restoreAllMocks();
  vi.useRealTimers();
  Reflect.deleteProperty(navigator, "clipboard");
});

function renderCopyButton() {
  const container = render(<CopyButton value={id} label="legal entity ID" />);
  return { container, button: page.getByRole("button", { name: "Copy legal entity ID" }) };
}

test("a click copies the value and announces the result", async () => {
  const writeText = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
  const { container, button } = renderCopyButton();

  await button.click();

  expect(writeText).toHaveBeenCalledWith(id);
  await expect.element(page.getByRole("status")).toHaveTextContent("Copied");
  await expectNoAxeViolations(container);
});

test("the keyboard focuses the button with a visible focus, and Enter copies", async () => {
  const writeText = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
  const { button } = renderCopyButton();

  await userEvent.tab();
  expectVisibleFocus(button.element());
  await userEvent.keyboard("{Enter}");

  expect(writeText).toHaveBeenCalledWith(id);
});

test("when the browser refuses the copy, the button tells the user to copy the text", async () => {
  vi.spyOn(navigator.clipboard, "writeText").mockRejectedValue(
    new DOMException("", "NotAllowedError"),
  );
  const { container, button } = renderCopyButton();

  await button.click();

  await expect
    .element(page.getByRole("status"))
    .toHaveTextContent("Could not copy. Select the text and copy it.");
  await expectNoAxeViolations(container);
});

// A page from a host name other than localhost over plain HTTP is not a secure context, so it has no clipboard.
test("when the page has no clipboard, the button tells the user to copy the text", async () => {
  Object.defineProperty(navigator, "clipboard", { value: undefined, configurable: true });
  const { button } = renderCopyButton();

  await button.click();

  await expect
    .element(page.getByRole("status"))
    .toHaveTextContent("Could not copy. Select the text and copy it.");
});

test("the result goes after 3 seconds", async () => {
  vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
  const { button } = renderCopyButton();
  await button.click();
  await expect.element(page.getByRole("status")).toHaveTextContent("Copied");

  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  await button.click();
  await expect.poll(() => page.getByRole("status").element().textContent).toBe("Copied");
  await vi.advanceTimersByTimeAsync(3000);
  vi.useRealTimers();

  await expect.element(page.getByRole("status")).toHaveTextContent("");
});

test("the button is 24 by 24 CSS px or more, and passes the accessibility checks before a copy", async () => {
  const { container, button } = renderCopyButton();

  expectTargetSize(button.element());
  await expectNoAxeViolations(container);
});
