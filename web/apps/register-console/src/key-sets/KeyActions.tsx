import { Button } from "@cdr-lab/ui/button";
import { ConfirmationDialog } from "@cdr-lab/ui/confirmation-dialog";
import { useToast } from "@cdr-lab/ui/toast";
import { Plus, Power, type LucideIcon } from "lucide-react";
import { useRef, useState } from "react";

import { ProblemError } from "@cdr-lab/console/problem";
import { changeKey, type KeyAction, type KeySet, type KeyState } from "../api/client";

export type KeySetNames = { title: string; noun: string; signs: string };

const steps = {
  "publish-next": {
    button: "Publish next key",
    icon: Plus,
    action: "Publish",
    title: (noun: string) => `Publish the next ${noun} key?`,
    done: (noun: string) => `The next ${noun} key is published`,
  },
  "activate-next": {
    button: "Activate next key",
    icon: Power,
    action: "Activate",
    title: (noun: string) => `Activate the next ${noun} key?`,
    done: (noun: string) => `The next ${noun} key now signs`,
  },
} as const satisfies Record<
  KeyAction["action"],
  {
    button: string;
    icon: LucideIcon;
    action: string;
    title: (noun: string) => string;
    done: (noun: string) => string;
  }
>;

const becomes = {
  signing: (id: string, names: KeySetNames) => `Key ${id} starts to sign ${names.signs}.`,
  previous: (id: string) => `Key ${id} becomes the previous key. The JWKS still publishes it.`,
  next: (id: string) => `Key ${id} becomes the next key.`,
} as const satisfies Record<KeyState, (id: string, names: KeySetNames) => string>;

function effectOf({ keyId, to }: KeyAction["changes"][number], names: KeySetNames): string {
  if (!to) return `Key ${keyId} leaves the JWKS, and the Register deletes it.`;
  return becomes[to](keyId, names);
}

function effectsOf(action: KeyAction, names: KeySetNames): string[] {
  const changes = action.changes.map((change) => effectOf(change, names));
  if (action.action === "publish-next") {
    return [
      `The JWKS of the ${names.noun} keys publishes a new next key.`,
      "The key that signs does not change.",
      ...changes,
    ];
  }
  return [
    ...changes,
    "A participant with an old copy of the JWKS cannot verify the new signatures.",
  ];
}

export type KeyActionsProps = {
  keySet: KeySet;
  names: KeySetNames;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
};

export function KeyActions({ keySet, names, refresh }: KeyActionsProps) {
  const toast = useToast();
  const group = useRef<HTMLFieldSetElement>(null);
  // The refresh after a refused step can remove the step, so its dialog stays until the operator closes it.
  const [open, setOpen] = useState<KeyAction>();
  const shown = [...keySet.actions];
  if (open && !shown.some((a) => a.action === open.action)) shown.push(open);

  async function confirm(action: KeyAction) {
    try {
      await changeKey(keySet.keySet, action.action, keySet.version);
    } catch (error) {
      if (!(error instanceof ProblemError)) throw error;
      await refresh();
      return error.problem;
    }
    await refresh();
    toast.show({ title: steps[action.action].done(names.noun) });
    return undefined;
  }

  return (
    <fieldset ref={group} tabIndex={-1} className="flex flex-wrap gap-2 outline-none">
      <legend className="sr-only">Key actions of {names.title}</legend>
      {shown.map((action) => (
        <ConfirmationDialog
          key={action.action}
          trigger={
            <Button
              variant={action.irreversible ? "danger" : "secondary"}
              icon={steps[action.action].icon}
            >
              {steps[action.action].button}
            </Button>
          }
          title={steps[action.action].title(names.noun)}
          action={steps[action.action].action}
          effects={effectsOf(action, names)}
          irreversible={action.irreversible}
          onConfirm={() => confirm(action)}
          onOpenChange={(isOpen) => setOpen(isOpen ? action : undefined)}
          returnFocus={group}
        />
      ))}
    </fieldset>
  );
}
