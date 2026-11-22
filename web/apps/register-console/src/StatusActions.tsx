import { Button } from "@cdr-lab/ui/button";
import {
  Ban,
  CirclePause,
  CirclePlay,
  LogOut,
  Power,
  PowerOff,
  Trash2,
  type LucideIcon,
} from "lucide-react";
import { RecordActions, type ActionDialog } from "@cdr-lab/web-platform/record-actions";

import {
  type DataHolderBrandStatusAction,
  type DataHolderStatusAction,
  type DataRecipientBrandStatusAction,
  type DataRecipientStatusAction,
  type SoftwareProductStatusAction,
} from "./api/client";

type StatusAction =
  | DataRecipientStatusAction
  | DataRecipientBrandStatusAction
  | SoftwareProductStatusAction
  | DataHolderStatusAction
  | DataHolderBrandStatusAction;

type StatusChange = StatusAction["changes"][number];

function nameOf(change: StatusChange): string {
  return "softwareProductName" in change ? change.softwareProductName : change.brandName;
}

const names = {
  suspend: "Suspend",
  revoke: "Revoke",
  surrender: "Surrender",
  reinstate: "Reinstate",
  activate: "Activate",
  deactivate: "Deactivate",
  remove: "Remove",
} as const satisfies Record<StatusAction["action"], string>;

const icons = {
  suspend: CirclePause,
  revoke: Ban,
  surrender: LogOut,
  reinstate: CirclePlay,
  activate: Power,
  deactivate: PowerOff,
  remove: Trash2,
} as const satisfies Record<StatusAction["action"], LucideIcon>;

function effectsOf(record: { name: string; status: string }, action: StatusAction): string[] {
  const changes: readonly StatusChange[] = action.changes;
  const effects = [
    `${record.name} changes from ${record.status} to ${action.status}.`,
    ...changes.map((change) => `${nameOf(change)} changes from ${change.from} to ${change.to}.`),
  ];
  if (action.action === "reinstate") {
    effects.push(
      "The software products do not change. Activate each software product that must be ACTIVE.",
    );
  }
  return effects;
}

function dialogOf(record: { name: string; status: string }, action: StatusAction): ActionDialog {
  return {
    trigger: (
      <Button variant={action.irreversible ? "danger" : "secondary"} icon={icons[action.action]}>
        {names[action.action]}
      </Button>
    ),
    title: `${names[action.action]} ${record.name}?`,
    name: names[action.action],
    effects: effectsOf(record, action),
    irreversible: action.irreversible,
    result: `${record.name} is now ${action.status}`,
  };
}

export type StatusActionsProps<A extends StatusAction> = {
  record: { name: string; status: string };
  actions: A[];
  // Throws a ProblemError when the Register refuses the action.
  run: (action: A) => Promise<void>;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
};

export function StatusActions<A extends StatusAction>({
  record,
  actions,
  run,
  refresh,
}: StatusActionsProps<A>) {
  return (
    <RecordActions
      label={`Status actions of ${record.name}`}
      actions={actions}
      dialogOf={(action) => dialogOf(record, action)}
      run={run}
      refresh={refresh}
    />
  );
}
