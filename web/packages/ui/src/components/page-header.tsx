import { useEffect, useRef, type ReactNode } from "react";

export type PageHeaderProps = {
  title: string;
  description?: string;
  actions?: ReactNode;
  // Moves the focus to the title when the page shows, for example after a form that sent the user to this page.
  focusTitle?: boolean;
};

export function PageHeader({ title, description, actions, focusTitle = false }: PageHeaderProps) {
  const heading = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    if (focusTitle) heading.current?.focus();
  }, [focusTitle]);

  return (
    // The band has the full width of the main area, so it cancels the padding of the main area.
    <header className="-mx-6 -mt-6 mb-6 flex items-end justify-between gap-4 border-b border-border bg-shell-band px-6 py-4">
      <div className="flex flex-col gap-1">
        <h1
          ref={heading}
          tabIndex={focusTitle ? -1 : undefined}
          className="text-2xl leading-tight font-semibold text-heading outline-none"
        >
          {title}
        </h1>
        {description && <p className="text-sm text-fg-muted">{description}</p>}
      </div>
      {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
    </header>
  );
}
