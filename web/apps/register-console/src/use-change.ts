import type { FormRefusal } from "@cdr-lab/ui/form";
import { useToast } from "@cdr-lab/ui/toast";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { refusalOf } from "@cdr-lab/web-platform/problem";

export type Participant = "data-recipient" | "data-holder";

const detailPaths = {
  "data-recipient": "/data-recipients/$legalEntityId",
  "data-holder": "/data-holders/$legalEntityId",
} as const;

export function useSend(participant: Participant) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const toast = useToast();

  const refresh = (legalEntityId: string) =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: [participant, legalEntityId] }),
      queryClient.invalidateQueries({ queryKey: [`${participant}-history`, legalEntityId] }),
      queryClient.invalidateQueries({ queryKey: [`${participant}s`] }),
    ]);

  return async function send(
    legalEntityId: string,
    request: () => Promise<void>,
    result: string,
    open?: string,
  ): Promise<FormRefusal | undefined> {
    try {
      await request();
    } catch (error) {
      await refresh(legalEntityId);
      return refusalOf(error);
    }
    await refresh(legalEntityId);
    toast.show({ title: result });
    await navigate({
      to: detailPaths[participant],
      params: { legalEntityId },
      state: { focusTitle: true, ...(open && { open }) },
    });
    return undefined;
  };
}

export function useChange(participant: Participant, legalEntityId: string) {
  const send = useSend(participant);
  return (change: () => Promise<void>, name: string, open?: string) =>
    send(legalEntityId, change, `Details of ${name} changed`, open);
}
