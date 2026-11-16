import { useRef, useState } from "react";
import { expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { Button } from "./button";
import { ConfirmationDialog, type ConfirmationDialogProps } from "./confirmation-dialog";
import type { Problem } from "./states";

const suspension = {
  title: "Suspend Mock Recipient One?",
  action: "Suspend",
  effects: ["Mock Recipient One changes from ACTIVE to SUSPENDED."],
};

function renderConfirmation(props: Partial<ConfirmationDialogProps> = {}) {
  const onConfirm = vi.fn<() => Promise<Problem | undefined>>().mockResolvedValue(undefined);
  render(
    <ConfirmationDialog
      trigger={<Button variant="danger">Suspend</Button>}
      onConfirm={onConfirm}
      {...suspension}
      {...props}
    />,
  );
  return {
    onConfirm,
    trigger: page.getByRole("button", { name: "Suspend", exact: true }),
    dialog: page.getByRole("alertdialog", { name: suspension.title }),
  };
}

// Verifies: UI-0293
test("a confirmation is an alert dialog that names the action and its effect, and passes the accessibility checks", async () => {
  const { trigger, dialog } = renderConfirmation();

  await trigger.click();

  await expect.element(dialog).toBeVisible();
  await expect.element(dialog).toHaveAccessibleDescription(suspension.effects[0]);
  expect(dialog.getByRole("list").query()).toBeNull();
  const confirm = dialog.getByRole("button", { name: "Suspend", exact: true });
  await expect.element(confirm).toBeVisible();
  expectTargetSize(confirm.element());
  expectTargetSize(dialog.getByRole("button", { name: "Cancel" }).element());
  await expectNoAxeViolations(document.body);
});

// Verifies: UI-0293
test("a confirmation with more than one effect gives the effects in a list", async () => {
  const effects = [
    "Mock Recipient One changes from ACTIVE to SUSPENDED.",
    "Mock Product One changes from ACTIVE to INACTIVE.",
    "Mock Product Two changes from ACTIVE to INACTIVE.",
  ];
  const { trigger, dialog } = renderConfirmation({ effects });

  await trigger.click();

  const items = dialog.getByRole("listitem").all();
  expect(items.map((item) => item.element().textContent)).toEqual(effects);
  await expectNoAxeViolations(document.body);
});

// Verifies: UI-0291, UI-0294
test("Cancel closes the confirmation with no change, and the focus goes back to the trigger", async () => {
  const { onConfirm, trigger, dialog } = renderConfirmation();

  await trigger.click();
  await dialog.getByRole("button", { name: "Cancel" }).click();

  await expect.element(dialog).not.toBeInTheDocument();
  expect(onConfirm).not.toHaveBeenCalled();
  expect(document.activeElement).toBe(trigger.element());
});

test("Escape closes the confirmation with no change", async () => {
  const { onConfirm, trigger, dialog } = renderConfirmation();

  await trigger.click();
  await userEvent.keyboard("{Escape}");

  await expect.element(dialog).not.toBeInTheDocument();
  expect(onConfirm).not.toHaveBeenCalled();
});

// Verifies: UI-0295
test("the first focus is on Cancel, also when the action can be reversed", async () => {
  const { trigger, dialog } = renderConfirmation();

  await trigger.click();

  await expect
    .poll(() => document.activeElement)
    .toBe(dialog.getByRole("button", { name: "Cancel" }).element());
});

// Verifies: UI-0295, UI-0296
test("for an action that cannot be reversed, the confirmation says so, and the first focus is on Cancel", async () => {
  const { trigger, dialog } = renderConfirmation({ irreversible: true });

  await trigger.click();

  await expect
    .element(dialog)
    .toHaveAccessibleDescription(`${suspension.effects[0]} You cannot reverse this action.`);
  await expect
    .poll(() => document.activeElement)
    .toBe(dialog.getByRole("button", { name: "Cancel" }).element());
  await expectNoAxeViolations(document.body);
});

// Verifies: UI-0201
test("the keyboard opens the confirmation, moves to the confirm button and confirms", async () => {
  const { onConfirm, trigger, dialog } = renderConfirmation();

  await userEvent.tab();
  expectVisibleFocus(trigger.element());
  await userEvent.keyboard("{Enter}");
  const cancel = dialog.getByRole("button", { name: "Cancel" });
  await expect.poll(() => document.activeElement).toBe(cancel.element());
  expectVisibleFocus(cancel.element());
  await userEvent.tab();
  expectVisibleFocus(dialog.getByRole("button", { name: "Suspend", exact: true }).element());
  await userEvent.keyboard("{Enter}");

  await expect.element(dialog).not.toBeInTheDocument();
  expect(onConfirm).toHaveBeenCalledOnce();
  expect(document.activeElement).toBe(trigger.element());
});

// Verifies: UI-0298
test("while the action runs, the confirm button is busy, a status message says so, and the confirmation stays open", async () => {
  // The test resolves the promise, so that it can see the confirmation while the action runs.
  const action = Promise.withResolvers<Problem | undefined>();
  const { trigger, dialog } = renderConfirmation({ onConfirm: () => action.promise });

  await trigger.click();
  await dialog.getByRole("button", { name: "Suspend", exact: true }).click();

  const busy = dialog.getByRole("button", { name: "Suspend (busy)" });
  await expect.element(busy).toHaveAttribute("aria-busy", "true");
  await expect.element(dialog.getByRole("status")).toHaveTextContent("Suspend in progress");
  await expect.element(dialog.getByRole("button", { name: "Cancel" })).toBeDisabled();
  await expectNoAxeViolations(document.body);
  await userEvent.keyboard("{Escape}");
  await busy.click({ force: true });
  await expect.element(dialog).toBeVisible();

  action.resolve(undefined);
  await expect.element(dialog).not.toBeInTheDocument();
  expect(document.activeElement).toBe(trigger.element());
});

// Verifies: UI-0299
test("when the action fails, the confirmation stays open and shows the problem in an alert", async () => {
  const problem = {
    title: "Conflict",
    detail: "The data recipient changed. See its current status.",
  };
  const onConfirm = vi.fn<() => Promise<Problem | undefined>>().mockResolvedValue(problem);
  const { trigger, dialog } = renderConfirmation({ onConfirm });

  await trigger.click();
  await dialog.getByRole("button", { name: "Suspend", exact: true }).click();

  const alert = dialog.getByRole("alert");
  await expect.element(alert).toMatchTextContent("Conflict");
  await expect.element(alert).toMatchTextContent("The data recipient changed.");
  await expect.element(dialog.getByRole("button", { name: "Suspend", exact: true })).toBeEnabled();
  await expect.element(dialog.getByRole("status")).toHaveTextContent("");
  await expectNoAxeViolations(document.body);
  expect(onConfirm).toHaveBeenCalledOnce();
});

// Verifies: UI-0300
test("when the action throws, the confirmation stays open and tells the operator what to do", async () => {
  const { trigger, dialog } = renderConfirmation({
    onConfirm: () =>
      Promise.reject(new TypeError("NetworkError when attempting to fetch resource.")),
  });

  await trigger.click();
  await dialog.getByRole("button", { name: "Suspend", exact: true }).click();

  const alert = dialog.getByRole("alert");
  await expect.element(alert).toMatchTextContent("The action did not complete");
  await expect
    .element(alert)
    .toMatchTextContent("Refresh the page to see the current state, then try again.");
});

test("the problem of a failed action does not show when the confirmation opens again", async () => {
  const { trigger, dialog } = renderConfirmation({
    onConfirm: () => Promise.resolve({ title: "Conflict" }),
  });

  await trigger.click();
  await dialog.getByRole("button", { name: "Suspend", exact: true }).click();
  await expect.element(dialog.getByRole("alert")).toBeVisible();
  await dialog.getByRole("button", { name: "Cancel" }).click();
  await expect.element(dialog).not.toBeInTheDocument();
  await trigger.click();

  await expect.element(dialog).toBeVisible();
  expect(dialog.getByRole("alert").query()).toBeNull();
});

function SuspendableRecord() {
  const [status, setStatus] = useState("ACTIVE");
  const heading = useRef<HTMLHeadingElement>(null);
  return (
    <>
      <h1 ref={heading} tabIndex={-1}>
        Mock Recipient One: {status}
      </h1>
      {status === "ACTIVE" && (
        <ConfirmationDialog
          trigger={<Button variant="danger">Suspend</Button>}
          returnFocus={heading}
          onConfirm={async () => {
            setStatus("SUSPENDED");
            return undefined;
          }}
          {...suspension}
        />
      )}
    </>
  );
}

// Verifies: UI-0292
test("when the trigger does not exist after the action, the focus goes to the element that the UI gives", async () => {
  render(<SuspendableRecord />);

  await page.getByRole("button", { name: "Suspend", exact: true }).click();
  await page.getByRole("alertdialog").getByRole("button", { name: "Suspend", exact: true }).click();

  await expect.element(page.getByRole("alertdialog")).not.toBeInTheDocument();
  await expect.poll(() => document.activeElement).toBe(page.getByRole("heading").element());
});

test("the confirmation tells the UI each time that it opens or closes", async () => {
  const problem = { title: "Conflict", detail: "The data recipient changed." };
  const onConfirm = vi
    .fn<() => Promise<Problem | undefined>>()
    .mockResolvedValueOnce(problem)
    .mockResolvedValue(undefined);
  const onOpenChange = vi.fn<(open: boolean) => void>();
  const { trigger, dialog } = renderConfirmation({ onConfirm, onOpenChange });

  await trigger.click();
  await dialog.getByRole("button", { name: "Cancel" }).click();
  await expect.element(dialog).not.toBeInTheDocument();
  expect(onOpenChange.mock.calls).toEqual([[true], [false]]);

  await trigger.click();
  await dialog.getByRole("button", { name: "Suspend", exact: true }).click();
  await expect.element(dialog.getByRole("alert")).toBeVisible();
  // A refused action keeps the confirmation open.
  expect(onOpenChange.mock.calls).toEqual([[true], [false], [true]]);

  await dialog.getByRole("button", { name: "Suspend", exact: true }).click();
  await expect.element(dialog).not.toBeInTheDocument();
  expect(onOpenChange.mock.calls).toEqual([[true], [false], [true], [false]]);
});

function ControlledConfirmation({ onConfirm }: { onConfirm: () => Promise<Problem | undefined> }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button onClick={() => setOpen(true)}>Continue</Button>
      <ConfirmationDialog
        {...suspension}
        open={open}
        onOpenChange={setOpen}
        onConfirm={onConfirm}
      />
    </>
  );
}

test("with no trigger, the UI opens the confirmation, and the confirmation tells the UI when it closes", async () => {
  const onConfirm = vi.fn<() => Promise<Problem | undefined>>().mockResolvedValue(undefined);
  render(<ControlledConfirmation onConfirm={onConfirm} />);
  const opener = page.getByRole("button", { name: "Continue" });
  const dialog = page.getByRole("alertdialog", { name: suspension.title });

  await opener.click();

  await expect.element(dialog).toBeVisible();
  await expect
    .poll(() => document.activeElement)
    .toBe(dialog.getByRole("button", { name: "Cancel" }).element());
  await dialog.getByRole("button", { name: "Cancel" }).click();
  await expect.element(dialog).not.toBeInTheDocument();
  expect(document.activeElement).toBe(opener.element());

  await opener.click();
  await dialog.getByRole("button", { name: "Suspend", exact: true }).click();

  await expect.element(dialog).not.toBeInTheDocument();
  expect(onConfirm).toHaveBeenCalledOnce();
});
