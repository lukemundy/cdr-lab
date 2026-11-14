import { RecordActions, type ActionDialog } from "@cdr-lab/console/record-actions";
import { Button } from "@cdr-lab/ui/button";
import { ArchiveX } from "lucide-react";

import { closeAccount, type Account, type AccountAction } from "../api/client";

function effectsOf(account: Account, action: AccountAction): string[] {
  const effects = [`${account.displayName} changes from ${account.openStatus} to CLOSED.`];
  if (action.ends.length > 0) {
    const count =
      action.ends.length === 1 ? "1 arrangement ends" : `${action.ends.length} arrangements end`;
    effects.push(`It is the last open account of the customer, so ${count}.`);
  }
  for (const end of action.ends) {
    effects.push(
      `The arrangement ${end.arrangementId} of ${end.clientName} ends. ` +
        "The data holder sends a notice to the data recipient.",
    );
  }
  return effects;
}

function closeDialog(account: Account, action: AccountAction): ActionDialog {
  const name = account.displayName;
  return {
    trigger: (
      <Button variant="danger" icon={ArchiveX}>
        Close
      </Button>
    ),
    title: `Close ${name}?`,
    name: "Close",
    effects: effectsOf(account, action),
    irreversible: true,
    result: `${name} is now CLOSED`,
  };
}

export type AccountActionsProps = {
  customerNumber: string;
  account: Account;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
};

// The create of a transaction is a form, so it is a link on the transactions of the account, not a dialog here.
export function AccountActions({ customerNumber, account, refresh }: AccountActionsProps) {
  return (
    <RecordActions
      label={`Actions of ${account.displayName}`}
      actions={account.actions.filter((action) => action.action === "close")}
      dialogOf={(action) => closeDialog(account, action)}
      run={(action) =>
        closeAccount(
          customerNumber,
          account.accountId,
          action.ends.map((end) => end.arrangementId),
        )
      }
      refresh={refresh}
    />
  );
}
