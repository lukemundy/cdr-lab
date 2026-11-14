import type { ReactNode } from "react";

import { cn } from "../lib/cn";

export type FieldsetProps = {
  legend: string;
  className?: string | undefined;
  children: ReactNode;
};

export function Fieldset({ legend, className, children }: FieldsetProps) {
  return (
    <fieldset className={cn("flex flex-col gap-4 rounded-md border border-border p-4", className)}>
      <legend className="px-1 text-sm font-semibold text-fg">{legend}</legend>
      {children}
    </fieldset>
  );
}
