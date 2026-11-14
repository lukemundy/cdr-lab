import { mergeProps } from "@base-ui/react/merge-props";
import { useRender } from "@base-ui/react/use-render";
import type { LucideIcon } from "lucide-react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";

const underline = "underline underline-offset-2 hover:decoration-2";
const linkClasses = cn("rounded-sm text-link", focusRing);

export type LinkProps = useRender.ComponentProps<"a"> & { icon?: LucideIcon };

// The render prop takes the link element of a router, for example TanStack Router.
export function Link({ className, render, icon: Icon, children, ...props }: LinkProps) {
  return useRender({
    defaultTagName: "a",
    render,
    props: mergeProps<"a">(
      {
        className: cn(linkClasses, Icon ? "inline-flex items-center gap-1" : underline, className),
        children: Icon ? (
          <>
            <Icon aria-hidden="true" className="size-4 shrink-0" />
            <span className={underline}>{children}</span>
          </>
        ) : (
          children
        ),
      },
      props,
    ),
  });
}
