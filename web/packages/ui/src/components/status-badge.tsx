import { cva, type VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "../lib/cn";

// From the shadcn/ui registry 4.22.1, style base-nova, item badge, changed to the status roles.
// The text gives the status, and the tone only supports it (WCAG 2.2, SC 1.4.1).
const statusBadgeVariants = cva(
  "inline-flex w-fit shrink-0 items-center rounded-full border px-2 text-xs leading-normal font-medium whitespace-nowrap",
  {
    variants: {
      tone: {
        neutral: "border-border bg-secondary text-fg",
        success: "border-success-border bg-success-bg text-success-fg",
        warning: "border-warning-border bg-warning-bg text-warning-fg",
        danger: "border-danger-border bg-danger-bg text-danger-fg",
        info: "border-info-border bg-info-bg text-info-fg",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

export type StatusBadgeProps = ComponentProps<"span"> & VariantProps<typeof statusBadgeVariants>;

export function StatusBadge({ className, tone, ...props }: StatusBadgeProps) {
  return <span className={cn(statusBadgeVariants({ tone }), className)} {...props} />;
}
