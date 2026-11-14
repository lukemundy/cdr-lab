import { X } from "lucide-react";
import {
  useEffect,
  useEffectEvent,
  useId,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";
import { Button } from "./button";

// From the shadcn/ui registry 4.22.1, style base-nova, items input and native-select, changed to the semantic tokens.
const fieldClasses = cn(
  "h-control rounded-md border border-border-strong bg-surface px-control-x text-sm text-fg",
  focusRing,
);

// DESIGN.md, Decisions: when the search of the filter bar applies.
const searchDelayMs = 300;

type SearchProps =
  | { searchLabel: string; search: string; onSearch: (search: string) => void }
  | { searchLabel?: never; search?: never; onSearch?: never };

export type FilterBarProps = SearchProps & {
  label: string;
  children?: ReactNode;
};

export function FilterBar({ label, searchLabel, search = "", onSearch, children }: FilterBarProps) {
  const searchId = useId();
  const field = useRef<HTMLInputElement>(null);
  const [value, setValue] = useState(search);
  // The search that the field applied last, or that it got last from the parent.
  const [applied, setApplied] = useState(search);
  const [previousSearch, setPreviousSearch] = useState(search);

  // A new search that this field did not apply, for example from the URL after Back, replaces the text of the field.
  if (search !== previousSearch) {
    setPreviousSearch(search);
    if (search !== applied) {
      setApplied(search);
      setValue(search);
    }
  }

  function apply(next: string) {
    const trimmed = next.trim();
    if (trimmed === applied) return;
    setApplied(trimmed);
    onSearch?.(trimmed);
  }

  // An effect event, so that the timer calls the onSearch of the latest render, and a new render does not restart the timer.
  const applyLater = useEffectEvent(apply);

  useEffect(() => {
    const timer =
      value.trim() === applied ? undefined : setTimeout(() => applyLater(value), searchDelayMs);
    return () => clearTimeout(timer);
  }, [value, applied]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    apply(value);
  }

  function clear() {
    setValue("");
    apply("");
    field.current?.focus();
  }

  return (
    <form
      // oxlint-disable-next-line jsx-a11y/prefer-tag-over-role -- the Playwright role engine does not give the search element its role, so a test cannot find it by role.
      role="search"
      aria-label={label}
      onSubmit={submit}
      className="flex flex-wrap items-end gap-4 pb-4"
    >
      {searchLabel !== undefined && (
        <div className="flex flex-col gap-1">
          <label htmlFor={searchId} className="text-sm font-medium text-fg">
            {searchLabel}
          </label>
          <div className="relative">
            <input
              ref={field}
              id={searchId}
              name="search"
              type="search"
              value={value}
              onChange={(event) => setValue(event.target.value)}
              // The clear button replaces the cancel button of the browser.
              className={cn(
                fieldClasses,
                "w-64 pr-control [&::-webkit-search-cancel-button]:appearance-none",
              )}
            />
            {value !== "" && (
              <Button
                type="button"
                variant="quiet"
                size="icon"
                aria-label="Clear search"
                onClick={clear}
                className="absolute inset-y-0 right-0"
              >
                <X aria-hidden="true" />
              </Button>
            )}
          </div>
        </div>
      )}
      {children}
    </form>
  );
}

export type FilterSelectProps = {
  label: string;
  value: string;
  options: readonly { value: string; label: string }[];
  onChange: (value: string) => void;
};

export function FilterSelect({ label, value, options, onChange }: FilterSelectProps) {
  const id = useId();

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="text-sm font-medium text-fg">
        {label}
      </label>
      <select
        id={id}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={fieldClasses}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
}
