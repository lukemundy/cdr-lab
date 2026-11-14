import { cn } from "../lib/cn";
import { CopyButton } from "./copy-button";

export type IdentifierProps = {
  value: string;
  // The shown text, for example a BSB as "062-000". The copy button still copies `value`.
  display?: string;
  // The name of the value, for example "legal entity ID". The copy button name is "Copy <label>".
  label: string;
  className?: string;
};

export function Identifier({ value, display = value, label, className }: IdentifierProps) {
  return (
    <span className={cn("inline-flex max-w-full min-w-0 items-center gap-1", className)}>
      <code className="min-w-0 font-mono text-xs wrap-break-word">{display}</code>
      <CopyButton value={value} label={label} className="shrink-0" />
    </span>
  );
}
