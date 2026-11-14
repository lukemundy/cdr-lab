import { Toast as BaseToast } from "@base-ui/react/toast";
import { CircleCheck, X } from "lucide-react";
import type { ReactNode } from "react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";
import { Button } from "./button";

const timeout = 5000;

export type ToastMessage = {
  title: string;
  description?: string;
};

export function useToast() {
  const manager = BaseToast.useToastManager();
  return {
    show({ title, description }: ToastMessage) {
      manager.add({ title, description });
    },
  };
}

function ToastList() {
  const { toasts } = BaseToast.useToastManager();
  return toasts.map((toast) => (
    <BaseToast.Root
      key={toast.id}
      toast={toast}
      className={cn(
        "pointer-events-auto rounded-lg border border-border bg-raised text-fg shadow-3",
        "data-limited:hidden motion-safe:transition-opacity data-starting-style:opacity-0 data-ending-style:opacity-0",
        focusRing,
      )}
    >
      <BaseToast.Content className="flex items-start gap-3 p-4">
        <CircleCheck aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-success-fg" />
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <BaseToast.Title className="text-sm font-medium" />
          <BaseToast.Description className="text-sm text-fg-muted" />
        </div>
        <BaseToast.Close
          aria-label="Close notification"
          render={<Button variant="quiet" size="icon" className="-my-1" />}
        >
          <X aria-hidden="true" />
        </BaseToast.Close>
      </BaseToast.Content>
    </BaseToast.Root>
  ));
}

// From the shadcn/ui registry 4.22.1, style base-nova, item toast, changed to the semantic tokens and to a list with no stack animation.
export function ToastProvider({ children }: { children: ReactNode }) {
  return (
    <BaseToast.Provider timeout={timeout}>
      {children}
      <BaseToast.Portal>
        <BaseToast.Viewport className="pointer-events-none fixed right-4 bottom-4 z-50 flex w-96 max-w-[calc(100vw-2rem)] flex-col gap-2 outline-none">
          <ToastList />
        </BaseToast.Viewport>
      </BaseToast.Portal>
    </BaseToast.Provider>
  );
}
