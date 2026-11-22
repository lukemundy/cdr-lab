import { ProblemError } from "@cdr-lab/app/problem";
import { ConfirmationDialog } from "@cdr-lab/ui/confirmation-dialog";
import { ConsumerButton } from "@cdr-lab/ui/consumer-page";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { closeAccount, type Account, type AccountAction } from "../api/client";
import { useBrand } from "../brand";
import { formatAmount } from "../products/format";
import { recordSessionEnded } from "../session/session";

// The confirmation as it was when it opened, so that a refused close does not change the list under the consumer.
type Confirmation = { account: Account; action: AccountAction; lastOpen: boolean };

function effectsOf({ account, action, lastOpen }: Confirmation, brandName: string): string[] {
  const effects = [
    `${account.displayName}, account number ${account.maskedNumber}, has a balance of ` +
      `${formatAmount(account.balance.currentBalance)}.`,
    "This is a demonstration bank, so it pays out no balance.",
  ];
  if (lastOpen) {
    effects.push(
      "This is your last open account. After the close, you cannot share your data under the Consumer Data Right.",
    );
  }
  for (const end of action.ends) {
    effects.push(
      `Your consent for ${end.clientName} ends, and ${brandName} tells ${end.clientName}.`,
    );
  }
  return effects;
}

export type CloseAccountProps = {
  account: Account;
  openAccounts: number;
};

export function CloseAccount({ account, openAccounts }: CloseAccountProps) {
  const { name } = useBrand();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [opened, setOpened] = useState<Confirmation>();
  const action = account.actions.find((a) => a.action === "close");
  const current = action && { account, action, lastOpen: openAccounts === 1 };
  const shown = opened ?? current;
  if (!shown) return null;

  async function confirm(confirmation: Confirmation) {
    try {
      await closeAccount(
        confirmation.account.accountId,
        confirmation.action.ends.map((end) => end.arrangementId),
      );
    } catch (error) {
      if (!(error instanceof ProblemError)) throw error;
      recordSessionEnded(queryClient, error);
      await queryClient.invalidateQueries({ queryKey: ["accounts"] });
      return error.problem;
    }
    await queryClient.invalidateQueries({ queryKey: ["accounts"] });
    await navigate({
      to: "/accounts",
      state: { status: `You closed ${confirmation.account.displayName}.` },
    });
    return undefined;
  }

  return (
    <ConfirmationDialog
      trigger={
        <ConsumerButton variant="secondary" render={(props) => <button {...props} type="button" />}>
          Close account
        </ConsumerButton>
      }
      title={`Close ${shown.account.displayName}?`}
      action="Close account"
      effects={effectsOf(shown, name)}
      irreversible
      onConfirm={() => confirm(shown)}
      onOpenChange={(open) => setOpened(open ? current : undefined)}
    />
  );
}
