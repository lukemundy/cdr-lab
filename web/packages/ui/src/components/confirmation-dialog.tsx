import { AlertDialog } from "@base-ui/react/alert-dialog";
import { useRef, useState, type ReactElement, type ReactNode, type RefObject } from "react";

import { useReturnFocus } from "../lib/return-focus";
import { Alert } from "./alert";
import { Button } from "./button";
import { backdropClass, descriptionClass, popupClass, titleClass } from "./dialog";
import type { Problem } from "./states";

const unexpectedProblem: Problem = {
  title: "The action did not complete",
  detail: "Refresh the page to see the current state, then try again.",
};

export type ConfirmationDialogProps = {
  // Without a trigger, the UI controls the confirmation with open and onOpenChange.
  trigger?: ReactElement;
  open?: boolean;
  title: string;
  // The name of the confirm button, for example "Suspend".
  action: string;
  effects: ReactNode[];
  irreversible?: boolean;
  // Resolve with a problem to keep the dialog open, or with undefined only when the views show the new state, so that the focus can go back to a control that still exists.
  onConfirm: () => Promise<Problem | undefined>;
  // The element that gets the focus when the trigger does not exist after the action.
  returnFocus?: RefObject<HTMLElement | null>;
  // For example, to keep a trigger that the views no longer offer until the operator closes its refused action.
  onOpenChange?: (open: boolean) => void;
};

// From the shadcn/ui registry 4.22.1, style base-nova, item alert-dialog, changed to the semantic tokens.
export function ConfirmationDialog({
  trigger,
  title,
  action,
  effects,
  irreversible = false,
  onConfirm,
  returnFocus,
  open: controlledOpen,
  onOpenChange,
}: ConfirmationDialogProps) {
  const [uncontrolledOpen, setOpen] = useState(false);
  const open = controlledOpen ?? uncontrolledOpen;
  const [wasOpen, setWasOpen] = useState(open);
  const [busy, setBusy] = useState(false);
  const [problem, setProblem] = useState<Problem | undefined>();
  const cancel = useRef<HTMLButtonElement>(null);
  // The problem of an earlier action does not show when the confirmation opens again, also when the UI opens it.
  if (open !== wasOpen) {
    setWasOpen(open);
    if (open) setProblem(undefined);
  }
  const { popupRef, rememberOpener, finalFocus } = useReturnFocus(returnFocus);

  function changeOpen(next: boolean) {
    if (busy) return;
    setOpen(next);
    onOpenChange?.(next);
  }

  async function confirm() {
    setBusy(true);
    setProblem(undefined);
    let result: Problem | undefined;
    try {
      result = await onConfirm();
    } catch {
      result = unexpectedProblem;
    }
    setBusy(false);
    setProblem(result);
    if (!result) {
      setOpen(false);
      onOpenChange?.(false);
    }
  }

  return (
    <AlertDialog.Root open={open} onOpenChange={changeOpen}>
      {trigger && <AlertDialog.Trigger render={trigger} />}
      <AlertDialog.Portal>
        <AlertDialog.Backdrop className={backdropClass} />
        <AlertDialog.Popup
          ref={popupRef}
          className={popupClass}
          initialFocus={() => {
            rememberOpener();
            return cancel.current;
          }}
          finalFocus={finalFocus}
        >
          <AlertDialog.Title className={titleClass}>{title}</AlertDialog.Title>
          <AlertDialog.Description render={<div />} className={descriptionClass}>
            {effects.length > 1 ? (
              <ul className="flex list-disc flex-col gap-1 pl-5">
                {effects.map((effect, i) => (
                  // oxlint-disable-next-line react/no-array-index-key -- the effects of one confirmation are a fixed list.
                  <li key={i}>{effect}</li>
                ))}
              </ul>
            ) : (
              <p>{effects[0]}</p>
            )}
            {irreversible && (
              <p className="pt-2 font-medium text-fg">You cannot reverse this action.</p>
            )}
          </AlertDialog.Description>
          {problem && (
            <Alert title={problem.title}>{problem.detail && <p>{problem.detail}</p>}</Alert>
          )}
          <output className="sr-only">{busy && `${action} in progress`}</output>
          <div className="flex justify-end gap-2 pt-2">
            <AlertDialog.Close ref={cancel} disabled={busy} render={<Button variant="secondary" />}>
              Cancel
            </AlertDialog.Close>
            <Button
              variant={irreversible ? "danger" : "primary"}
              busy={busy}
              onClick={() => void confirm()}
            >
              {action}
            </Button>
          </div>
        </AlertDialog.Popup>
      </AlertDialog.Portal>
    </AlertDialog.Root>
  );
}
