import { RecordActions, type ActionDialog } from "@cdr-lab/web-platform/record-actions";
import { Button } from "@cdr-lab/ui/button";
import { Ban } from "lucide-react";

import { withdrawArrangement, type Arrangement, type ArrangementAction } from "../api/client";

function withdrawDialog(arrangement: Arrangement, action: ArrangementAction): ActionDialog {
  const of = arrangement.clientName;
  return {
    trigger: (
      <Button variant="danger" icon={Ban}>
        Withdraw
      </Button>
    ),
    title: `Withdraw the consent of ${of}?`,
    name: "Withdraw",
    effects: action.ends.map(
      (end) =>
        `The arrangement ${end.arrangementId} of ${end.clientName} ends. ` +
        "The data holder sends a notice to the data recipient.",
    ),
    irreversible: true,
    result: `The arrangement of ${of} is now ENDED`,
  };
}

export type ArrangementActionsProps = {
  customerNumber: string;
  arrangement: Arrangement;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
};

export function ArrangementActions({
  customerNumber,
  arrangement,
  refresh,
}: ArrangementActionsProps) {
  return (
    <RecordActions
      label={`Actions of ${arrangement.clientName}`}
      actions={arrangement.actions}
      dialogOf={(action) => withdrawDialog(arrangement, action)}
      run={() => withdrawArrangement(customerNumber, arrangement.arrangementId)}
      refresh={refresh}
    />
  );
}
