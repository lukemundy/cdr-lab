import { Dialog as BaseDialog } from "@base-ui/react/dialog";
import { X } from "lucide-react";
import type { ReactElement, ReactNode, RefObject } from "react";

import { cn } from "../lib/cn";
import { useReturnFocus } from "../lib/return-focus";
import { Button, type ButtonProps } from "./button";

// From the shadcn/ui registry 4.22.1, style base-nova, item dialog, changed to the semantic tokens.
export const backdropClass = cn(
  "fixed inset-0 z-50 bg-backdrop",
  "motion-safe:transition-opacity data-starting-style:opacity-0 data-ending-style:opacity-0",
);

export const popupClass = cn(
  "fixed top-1/2 left-1/2 z-50 flex w-full max-w-lg -translate-x-1/2 -translate-y-1/2 flex-col gap-4",
  "max-h-[calc(100svh-2rem)] overflow-y-auto rounded-lg border border-border bg-raised p-6 text-fg shadow-3",
  "outline-none motion-safe:transition-opacity data-starting-style:opacity-0 data-ending-style:opacity-0",
);

export const titleClass = "pr-8 text-lg font-semibold text-fg";

export const descriptionClass = "text-sm text-fg-muted";

export type DialogProps = {
  title: string;
  description?: ReactNode;
  // Without a trigger, the UI controls the dialog with open and onOpenChange.
  trigger?: ReactElement;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  // The element that gets the focus when the dialog closes and the control that opened it does not exist.
  returnFocus?: RefObject<HTMLElement | null>;
  children: ReactNode;
};

export function Dialog({
  title,
  description,
  trigger,
  open,
  onOpenChange,
  returnFocus,
  children,
}: DialogProps) {
  const { popupRef, rememberOpener, finalFocus } = useReturnFocus(returnFocus);

  return (
    <BaseDialog.Root open={open} onOpenChange={onOpenChange}>
      {trigger && <BaseDialog.Trigger render={trigger} />}
      <BaseDialog.Portal>
        <BaseDialog.Backdrop className={backdropClass} />
        <BaseDialog.Popup
          ref={popupRef}
          className={popupClass}
          initialFocus={() => {
            rememberOpener();
            return true;
          }}
          finalFocus={finalFocus}
        >
          <BaseDialog.Title className={titleClass}>{title}</BaseDialog.Title>
          {description && (
            <BaseDialog.Description className={descriptionClass}>
              {description}
            </BaseDialog.Description>
          )}
          {children}
          <BaseDialog.Close
            render={<Button variant="quiet" size="icon" className="absolute top-3 right-3" />}
            aria-label="Close"
          >
            <X aria-hidden="true" />
          </BaseDialog.Close>
        </BaseDialog.Popup>
      </BaseDialog.Portal>
    </BaseDialog.Root>
  );
}

export function DialogFooter({ children }: { children: ReactNode }) {
  return <div className="flex justify-end gap-2 pt-2">{children}</div>;
}

export function DialogClose({ variant = "secondary", ...props }: ButtonProps) {
  return <BaseDialog.Close render={<Button variant={variant} {...props} />} />;
}
