import { ProblemError } from "@cdr-lab/console/problem";
import { ConfirmationDialog } from "@cdr-lab/ui/confirmation-dialog";
import { ConsumerButton } from "@cdr-lab/ui/consumer-page";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { withdrawConsent, type Consent } from "../api/client";
import { useBrand } from "../brand";
import { consentsKey } from "./useConsents";

// The last two effects are the DSB text of CDS: Withdrawal Standards (Consequences and Redundant data), word for word.
function effectsOf({ clientName, brandName }: Consent, bankName: string): string[] {
  return [
    `${bankName} stops sharing your data with ${clientName} of ${brandName} at once.`,
    `${bankName} tells ${brandName} that you stopped the sharing.`,
    `You should check with ${brandName} before you stop sharing to understand the consequences.`,
    "CDR data is either deleted or de-identified when it is no longer required.",
  ];
}

export function StopSharing({ consent }: { consent: Consent }) {
  const { name } = useBrand();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  // A refused withdrawal loads the consents when the dialog closes, so that the dialog stays while the consumer reads
  // the refusal.
  const [refused, setRefused] = useState(false);

  async function confirm() {
    try {
      await withdrawConsent(consent.arrangementId);
    } catch (error) {
      if (!(error instanceof ProblemError)) throw error;
      setRefused(true);
      return error.problem;
    }
    await queryClient.invalidateQueries({ queryKey: consentsKey });
    await navigate({
      to: "/data-sharing",
      state: { status: `You stopped sharing your data with ${consent.clientName}.` },
    });
    return undefined;
  }

  function changeOpen(open: boolean) {
    if (open || !refused) return;
    setRefused(false);
    void queryClient.invalidateQueries({ queryKey: consentsKey });
  }

  return (
    <ConfirmationDialog
      trigger={
        <ConsumerButton variant="secondary" render={(props) => <button {...props} type="button" />}>
          Stop sharing
        </ConsumerButton>
      }
      title={`Stop sharing with ${consent.clientName}?`}
      action="Stop sharing"
      effects={effectsOf(consent, name)}
      irreversible
      onConfirm={confirm}
      onOpenChange={changeOpen}
    />
  );
}
