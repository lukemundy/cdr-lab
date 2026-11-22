import { ConfirmationDialog } from "@cdr-lab/ui/confirmation-dialog";
import { useToast } from "@cdr-lab/ui/toast";
import { useRef, useState, type ReactElement, type ReactNode } from "react";

import { ProblemError } from "./problem";

export type ActionDialog = {
  trigger: ReactElement;
  title: string;
  // The name of the confirm button.
  name: string;
  effects: ReactNode[];
  irreversible: boolean;
  // The title of the toast after a success.
  result: string;
};

export type RecordActionsProps<A extends { action: string }> = {
  // The accessible name of the group of the actions.
  label: string;
  // Only the actions that the console API permits now.
  actions: A[];
  dialogOf: (action: A) => ActionDialog;
  // Throws a ProblemError when the service refuses the action.
  run: (action: A) => Promise<void>;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
};

export function RecordActions<A extends { action: string }>({
  label,
  actions,
  dialogOf,
  run,
  refresh,
}: RecordActionsProps<A>) {
  const toast = useToast();
  const group = useRef<HTMLFieldSetElement>(null);
  // The refresh after a refused action can remove the action, so its dialog stays, as it was when it opened, until the operator closes it.
  const [open, setOpen] = useState<{ action: A; dialog: ActionDialog }>();
  const shown = actions.map((action) => ({ action, dialog: dialogOf(action) }));
  if (open && !actions.some((a) => a.action === open.action.action)) shown.push(open);

  async function confirm(action: A, dialog: ActionDialog) {
    try {
      await run(action);
    } catch (error) {
      if (!(error instanceof ProblemError)) throw error;
      await refresh();
      return error.problem;
    }
    await refresh();
    toast.show({ title: dialog.result });
    return undefined;
  }

  return (
    <fieldset ref={group} tabIndex={-1} className="flex flex-wrap gap-2 outline-none">
      <legend className="sr-only">{label}</legend>
      {shown.map(({ action, dialog }) => (
        <ConfirmationDialog
          key={action.action}
          trigger={dialog.trigger}
          title={dialog.title}
          action={dialog.name}
          effects={dialog.effects}
          irreversible={dialog.irreversible}
          onConfirm={() => confirm(action, dialog)}
          onOpenChange={(isOpen) => setOpen(isOpen ? { action, dialog } : undefined)}
          returnFocus={group}
        />
      ))}
    </fieldset>
  );
}
