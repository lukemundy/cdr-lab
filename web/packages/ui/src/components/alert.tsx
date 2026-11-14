import { CircleAlert } from "lucide-react";
import type { ReactNode } from "react";

import { cn } from "../lib/cn";

export type AlertProps = {
  title: string;
  children?: ReactNode;
  className?: string;
};

// From the shadcn/ui registry 4.22.1, style base-nova, item alert, changed to the semantic tokens.
export function Alert({ title, children, className }: AlertProps) {
  return (
    <div
      role="alert"
      className={cn(
        "flex items-start gap-2 rounded-md border border-danger-border bg-danger-bg p-4 text-danger-fg",
        className,
      )}
    >
      <CircleAlert aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
      <div className="flex flex-col gap-1">
        <p className="font-medium">{title}</p>
        {children && <div className="text-sm">{children}</div>}
      </div>
    </div>
  );
}
