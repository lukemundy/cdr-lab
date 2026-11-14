import { afterEach, expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { Button } from "./button";
import { ToastProvider, useToast } from "./toast";

afterEach(() => {
  vi.useRealTimers();
});

function Example() {
  const toast = useToast();
  return (
    <Button
      onClick={() =>
        toast.show({
          title: "Mock Recipient One is suspended",
          description: "2 software products are inactive.",
        })
      }
    >
      Suspend
    </Button>
  );
}

function renderToast() {
  render(
    <ToastProvider>
      <Example />
    </ToastProvider>,
  );
  return {
    button: page.getByRole("button", { name: "Suspend" }),
    region: page.getByRole("region", { name: "Notifications" }),
    toast: page.getByRole("dialog", { name: "Mock Recipient One is suspended" }),
  };
}

test("a toast shows in a live region, so that it is announced as a status message", async () => {
  const { button, region, toast } = renderToast();

  await expect.element(region).toHaveAttribute("aria-live", "polite");
  await button.click();

  await expect.element(region.getByRole("dialog")).toBeVisible();
  await expect.element(toast).toHaveAccessibleDescription("2 software products are inactive.");
  await expectNoAxeViolations(document.body);
});

// Base UI hides the close button from assistive technology until the toast has the pointer or the focus.
test("the close button of a toast closes it", async () => {
  const { button, toast } = renderToast();

  await button.click();
  await toast.hover();
  const close = toast.getByRole("button", { name: "Close notification" });
  await expect.element(close).toBeVisible();
  expectTargetSize(close.element());
  await close.click();

  await expect.element(toast).not.toBeInTheDocument();
});

test("a toast closes by itself after 5 seconds", async () => {
  const { button, toast } = renderToast();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });

  await button.click();
  await expect.element(toast).toBeVisible();
  await vi.advanceTimersByTimeAsync(4900);
  await expect.element(toast).toBeVisible();
  await vi.advanceTimersByTimeAsync(100);
  vi.useRealTimers();

  await expect.element(toast).not.toBeInTheDocument();
});

test("a toast does not close while it has the pointer", async () => {
  const { button, toast } = renderToast();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });

  await button.click();
  await toast.hover();
  await vi.advanceTimersByTimeAsync(10_000);
  await expect.element(toast).toBeVisible();
  await button.hover();
  await vi.advanceTimersByTimeAsync(5000);
  vi.useRealTimers();

  await expect.element(toast).not.toBeInTheDocument();
});

test("the keyboard moves the focus to the toast with F6, the toast does not close while it has the focus, and Enter on its close button closes it", async () => {
  const { button, toast } = renderToast();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });

  await userEvent.tab();
  await userEvent.keyboard("{Enter}");
  await expect.element(toast).toBeVisible();
  await userEvent.keyboard("{F6}");
  await userEvent.tab();
  expect(toast.element().contains(document.activeElement)).toBe(true);
  await vi.advanceTimersByTimeAsync(10_000);
  await expect.element(toast).toBeVisible();
  vi.useRealTimers();

  const close = toast.getByRole("button", { name: "Close notification" });
  while (document.activeElement !== close.element()) {
    // oxlint-disable-next-line no-await-in-loop -- each Tab key press starts from the focus that the previous one gave.
    await userEvent.tab();
  }
  expectVisibleFocus(close.element());
  await userEvent.keyboard("{Enter}");

  await expect.element(toast).not.toBeInTheDocument();
  expect(button.element()).toBeTruthy();
});
