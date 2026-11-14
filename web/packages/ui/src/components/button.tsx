import { Button as BaseButton } from "@base-ui/react/button";
import { cva, type VariantProps } from "class-variance-authority";
import { LoaderCircle, type LucideIcon } from "lucide-react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";

// From the shadcn/ui registry 4.22.1, style base-nova, item button, changed to the semantic tokens.
const buttonVariants = cva(
  [
    "inline-flex h-control shrink-0 items-center justify-center gap-2 rounded-md border px-control-x",
    "cursor-pointer text-sm font-medium whitespace-nowrap select-none motion-safe:transition-colors",
    "data-disabled:cursor-not-allowed data-disabled:not-aria-busy:opacity-50",
    "[&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
    focusRing,
  ],
  {
    variants: {
      variant: {
        primary:
          "border-transparent bg-primary text-fg-on-primary not-data-disabled:hover:bg-primary-hover not-data-disabled:active:bg-primary-pressed",
        secondary: "border-border-strong bg-secondary text-fg not-data-disabled:hover:bg-surface",
        danger:
          "border-border-strong bg-danger-bg text-danger-fg not-data-disabled:hover:border-danger-fg",
        quiet: "border-transparent bg-transparent text-fg not-data-disabled:hover:bg-secondary",
      },
      size: {
        default: "",
        icon: "w-control px-0",
      },
    },
    defaultVariants: { variant: "primary", size: "default" },
  },
);

export type ButtonProps = BaseButton.Props &
  VariantProps<typeof buttonVariants> & {
    // A busy button stays in the tab sequence, so that the focus does not move while it works.
    busy?: boolean;
    icon?: LucideIcon;
  };

export function Button({
  className,
  variant,
  size,
  busy = false,
  icon: Icon,
  disabled = false,
  children,
  ...props
}: ButtonProps) {
  return (
    <BaseButton
      className={cn(buttonVariants({ variant, size }), className)}
      disabled={disabled || busy}
      focusableWhenDisabled={busy}
      aria-busy={busy || undefined}
      {...props}
    >
      {busy ? (
        <LoaderCircle aria-hidden="true" className="motion-safe:animate-spin" />
      ) : (
        Icon && <Icon aria-hidden="true" />
      )}
      {children}
      {busy && <span className="sr-only">(busy)</span>}
    </BaseButton>
  );
}
