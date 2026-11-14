import { Check, Copy } from "lucide-react";
import { useEffect, useState } from "react";

import { cn } from "../lib/cn";
import { Button } from "./button";

const messages = {
  copied: "Copied",
  failed: "Could not copy. Select the text and copy it.",
};

type Result = { message: keyof typeof messages };

export type CopyButtonProps = {
  value: string;
  // The name of the value, for example "legal entity ID". The button name is "Copy <label>".
  label: string;
  className?: string;
};

export function CopyButton({ value, label, className }: CopyButtonProps) {
  // Each copy gives a new object, so that the timer starts again.
  const [result, setResult] = useState<Result | null>(null);

  useEffect(() => {
    if (!result) return undefined;
    const timer = setTimeout(() => setResult(null), 3000);
    return () => clearTimeout(timer);
  }, [result]);

  async function copy() {
    try {
      // navigator.clipboard is undefined outside a secure context, so this also throws there.
      await navigator.clipboard.writeText(value);
      setResult({ message: "copied" });
    } catch {
      setResult({ message: "failed" });
    }
  }

  return (
    <span className={cn("inline-flex items-center gap-2", className)}>
      <Button variant="quiet" size="icon" aria-label={`Copy ${label}`} onClick={() => void copy()}>
        {result?.message === "copied" ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
      </Button>
      <output className="text-xs text-fg-muted">{result && messages[result.message]}</output>
    </span>
  );
}
