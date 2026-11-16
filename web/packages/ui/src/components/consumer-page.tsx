import { mergeProps } from "@base-ui/react/merge-props";
import { useRender } from "@base-ui/react/use-render";
import { Lock } from "lucide-react";
import { useRef, type ReactNode } from "react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";

// The parts of the consumer pages of the data holder (DESIGN.md, Consumer pages). They use the classes of
// consumer-pages.css, so that the Go pages and a React consumer UI have one source for the look. A UI imports
// consumer.css for them.

export type ConsumerShellProps = {
  brandName: string;
  // The URL of the logo of the brand, which shows on the brand colour with the brand name as its text alternative.
  logo: string;
  // The link of the logo to the home page, for example the link element of a router.
  homeLink?: useRender.ComponentProps<"a">["render"];
  navigationLabel: string;
  navigation: ReactNode;
  footer: ReactNode;
  children: ReactNode;
};

// The shell has no "Secure" label, no hero and no step: those are the parts of the CDR authentication pages only. A
// consumer UI has the comfortable density (design/tokens-and-brands.md, UI-0029).
export function ConsumerShell({
  brandName,
  logo,
  homeLink,
  navigationLabel,
  navigation,
  footer,
  children,
}: ConsumerShellProps) {
  const main = useRef<HTMLElement>(null);
  const home = useRender({
    defaultTagName: "a",
    render: homeLink,
    props: {
      href: "/",
      className: cn("flex min-h-control items-center rounded-sm", focusRing),
      children: <img className="cp-logo max-w-none" src={logo} alt={brandName} />,
    },
  });

  return (
    <div data-density="comfortable" className="flex min-h-svh flex-col">
      <a
        href="#main-content"
        // The handler moves the focus itself, so that a router does not act on the fragment.
        onClick={(event) => {
          event.preventDefault();
          main.current?.focus();
        }}
        className={cn(
          "sr-only rounded-md bg-surface px-control-x py-2 font-medium text-fg focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-10",
          focusRing,
        )}
      >
        Skip to content
      </a>
      <header className="cp-header flex-wrap py-0">
        {home}
        <nav aria-label={navigationLabel}>
          <ul className="flex flex-wrap gap-x-4">{navigation}</ul>
        </nav>
      </header>
      <main id="main-content" ref={main} tabIndex={-1} className="cp-main pt-6 outline-none">
        {children}
      </main>
      {footer}
    </div>
  );
}

export type ConsumerNavItemProps = useRender.ComponentProps<"a">;

// The render prop takes the link element of a router, which can set aria-current itself.
export function ConsumerNavItem({ className, render, ...props }: ConsumerNavItemProps) {
  const link = useRender({
    defaultTagName: "a",
    render,
    props: mergeProps<"a">(
      {
        className: cn(
          "flex min-h-control items-center rounded-sm font-semibold text-fg-on-brand underline underline-offset-4",
          "hover:decoration-2 aria-[current=page]:decoration-2",
          focusRing,
          className,
        ),
      },
      props,
    ),
  });
  return <li>{link}</li>;
}

export function ConsumerCard({ className, children }: { className?: string; children: ReactNode }) {
  // The width is full up to the maximum of the card, also in a flex column.
  return <div className={cn("cp-card w-full", className)}>{children}</div>;
}

export type ConsumerButtonProps = useRender.ComponentProps<"a"> & {
  variant?: "primary" | "secondary" | "quiet";
};

// A link that looks like a button of the consumer pages. The render prop takes the link element of a router.
export function ConsumerButton({
  variant = "primary",
  className,
  render,
  ...props
}: ConsumerButtonProps) {
  return useRender({
    defaultTagName: "a",
    render,
    props: mergeProps<"a">(
      {
        className: cn(
          "cp-button",
          `cp-${variant}`,
          "flex items-center justify-center text-center",
          variant !== "quiet" && "no-underline",
          className,
        ),
      },
      props,
    ),
  });
}

export function PasswordStatement({ children }: { children: ReactNode }) {
  return (
    <div className="cp-note text-start">
      <Lock aria-hidden="true" className="cp-icon" />
      <p>{children}</p>
    </div>
  );
}

export function ConsumerFooter({ children }: { children: ReactNode }) {
  return <footer className="cp-footer flex flex-col gap-4">{children}</footer>;
}
