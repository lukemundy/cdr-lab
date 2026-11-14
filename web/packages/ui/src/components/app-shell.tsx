import { mergeProps } from "@base-ui/react/merge-props";
import { useRender } from "@base-ui/react/use-render";
import type { LucideIcon } from "lucide-react";
import { useRef, type ReactNode } from "react";

import { cn } from "../lib/cn";
import { focusRing, focusRingOnHero } from "../lib/focus";

export type AppShellProps = {
  name: string;
  // The URL of the logo of a brand. The header then shows the logo, with the name as its text alternative, on the brand
  // colour.
  logo?: string;
  // The text after the logo, for example the kind of UI.
  label?: string;
  navigationLabel: string;
  navigation: ReactNode;
  children: ReactNode;
};

export function AppShell({
  name,
  logo,
  label,
  navigationLabel,
  navigation,
  children,
}: AppShellProps) {
  const main = useRef<HTMLElement>(null);

  return (
    <div className="grid min-h-svh grid-cols-[16rem_1fr] grid-rows-[auto_1fr] bg-shell-page">
      <a
        href="#main-content"
        // The handler moves the focus itself, so that a router does not act on the fragment.
        onClick={(event) => {
          event.preventDefault();
          main.current?.focus();
        }}
        className={cn(
          "sr-only rounded-md bg-surface px-control-x py-2 text-sm font-medium text-fg focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-10",
          focusRing,
        )}
      >
        Skip to content
      </a>
      {logo === undefined ? (
        <header className="col-span-2 flex h-12 items-center border-b-3 border-accent bg-surface px-4">
          <p className="text-base font-semibold text-fg">{name}</p>
        </header>
      ) : (
        <header className="col-span-2 flex h-12 items-center gap-3 border-b-3 border-accent bg-brand px-4 text-fg-on-brand">
          <img src={logo} alt={name} className="block h-8 w-auto" />
          {label && <p className="text-sm font-medium">{label}</p>}
        </header>
      )}
      <nav aria-label={navigationLabel} className="bg-hero p-2 text-fg-on-hero">
        <ul className="flex flex-col gap-1">{navigation}</ul>
      </nav>
      <main
        id="main-content"
        ref={main}
        tabIndex={-1}
        className="min-w-0 bg-shell-page p-6 outline-none"
      >
        {children}
      </main>
    </div>
  );
}

export type AppShellNavItemProps = useRender.ComponentProps<"a"> & {
  current?: boolean;
  icon?: LucideIcon;
};

// The render prop takes the link element of a router, which can set aria-current itself.
export function AppShellNavItem({
  className,
  render,
  current = false,
  icon: Icon,
  children,
  ...props
}: AppShellNavItemProps) {
  const link = useRender({
    defaultTagName: "a",
    render,
    props: mergeProps<"a">(
      {
        className: cn(
          "flex min-h-control items-center gap-2 rounded-md border-l-3 border-transparent px-control-x text-sm",
          "text-fg-on-hero hover:bg-shell-nav-hover",
          // The bar is in the text colour, so that the current item does not depend on the accent.
          "aria-[current=page]:border-fg-on-hero aria-[current=page]:bg-shell-nav-current aria-[current=page]:font-semibold",
          focusRingOnHero,
          className,
        ),
        ...(current && { "aria-current": "page" as const }),
        children: (
          <>
            {Icon && <Icon aria-hidden="true" className="size-4 shrink-0" />}
            {children}
          </>
        ),
      },
      props,
    ),
  });
  return <li>{link}</li>;
}
