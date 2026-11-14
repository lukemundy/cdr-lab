import { useRef, useState } from "react";
import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { Button } from "./button";
import { Dialog, DialogClose, DialogFooter } from "./dialog";

function renderDialog() {
  render(
    <Dialog
      trigger={<Button>Register a certificate</Button>}
      title="Register a client certificate"
      description="The certificate of a software product."
    >
      <label>
        Certificate <input />
      </label>
      <DialogFooter>
        <DialogClose>Cancel</DialogClose>
      </DialogFooter>
    </Dialog>,
  );
  return {
    trigger: page.getByRole("button", { name: "Register a certificate" }),
    dialog: page.getByRole("dialog", { name: "Register a client certificate" }),
  };
}

test("the trigger opens a modal dialog with its title and description, and the focus goes into it", async () => {
  const { trigger, dialog } = renderDialog();

  await trigger.click();

  await expect.element(dialog).toBeVisible();
  await expect
    .element(dialog)
    .toHaveAccessibleDescription("The certificate of a software product.");
  expect(dialog.element().contains(document.activeElement)).toBe(true);
  await expectNoAxeViolations(document.body);
});

test("the keyboard opens the dialog, Tab stays in it, and Escape closes it and gives the focus back to the trigger", async () => {
  const { trigger, dialog } = renderDialog();

  await userEvent.tab();
  expectVisibleFocus(trigger.element());
  await userEvent.keyboard("{Enter}");
  await expect.element(dialog).toBeVisible();
  for (let i = 0; i < 5; i++) {
    // oxlint-disable-next-line no-await-in-loop -- each Tab key press starts from the focus that the previous one gave.
    await userEvent.tab();
    // The focus trap of Base UI moves the focus from a guard element at the edge of the dialog, after the key press.
    // oxlint-disable-next-line no-await-in-loop -- the check is for this Tab key press.
    await expect.poll(() => dialog.element().contains(document.activeElement)).toBe(true);
  }
  await userEvent.keyboard("{Escape}");

  await expect.element(dialog).not.toBeInTheDocument();
  expect(document.activeElement).toBe(trigger.element());
});

test("the close button and a close control in the dialog close it", async () => {
  const { trigger, dialog } = renderDialog();

  await trigger.click();
  const close = page.getByRole("button", { name: "Close", exact: true });
  expectTargetSize(close.element());
  await close.click();
  await expect.element(dialog).not.toBeInTheDocument();

  await trigger.click();
  await page.getByRole("button", { name: "Cancel" }).click();
  await expect.element(dialog).not.toBeInTheDocument();
  expect(document.activeElement).toBe(trigger.element());
});

function RemovableRecord() {
  const [open, setOpen] = useState(false);
  const [removed, setRemoved] = useState(false);
  const heading = useRef<HTMLHeadingElement>(null);
  return (
    <>
      <h1 ref={heading} tabIndex={-1}>
        Mock Recipient One
      </h1>
      {!removed && <Button onClick={() => setOpen(true)}>Remove</Button>}
      <Dialog title="Remove the record" open={open} onOpenChange={setOpen} returnFocus={heading}>
        <Button
          onClick={() => {
            setRemoved(true);
            setOpen(false);
          }}
        >
          Remove now
        </Button>
      </Dialog>
    </>
  );
}

test("when the control that opened the dialog does not exist at the close, the focus goes to the element that the UI gives", async () => {
  render(<RemovableRecord />);

  await page.getByRole("button", { name: "Remove", exact: true }).click();
  await page.getByRole("button", { name: "Remove now" }).click();

  await expect.element(page.getByRole("dialog")).not.toBeInTheDocument();
  await expect.poll(() => document.activeElement).toBe(page.getByRole("heading").element());
});
