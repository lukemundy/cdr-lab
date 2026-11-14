import { useId, type ReactNode } from "react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";
import { useFormBinding } from "../lib/form-context";

export type CheckboxOption = { value: string; label: string };

const noValues: readonly string[] = [];

export type CheckboxGroupProps = {
  // The legend of the group.
  label: string;
  name: string;
  options: readonly CheckboxOption[];
  defaultValue?: readonly string[] | undefined;
  description?: ReactNode;
  // An error from the app or from the server for the values. In a form, the group also gets the error of its name.
  error?: string | undefined;
  disabled?: boolean | undefined;
  className?: string | undefined;
};

// From the GOV.UK Design System checkboxes: the help text and the error are connected to the fieldset, and the id of
// the group is the id of its first checkbox, so that a link to the group goes to it.
export function CheckboxGroup({
  label,
  name,
  options,
  defaultValue = noValues,
  description,
  error,
  disabled,
  className,
}: CheckboxGroupProps) {
  const generated = useId();
  const binding = useFormBinding(name, error, undefined);
  const id = binding.id ?? generated;
  const descriptionId = `${id}-description`;
  const errorId = `${id}-error`;
  const describedBy = [description && descriptionId, binding.error && errorId]
    .filter(Boolean)
    .join(" ");

  return (
    <fieldset
      aria-describedby={describedBy || undefined}
      disabled={disabled}
      className={cn("flex flex-col gap-1", className)}
    >
      <legend className="mb-1 text-sm font-medium text-fg">{label}</legend>
      {description && (
        <p id={descriptionId} className="text-sm text-fg-muted">
          {description}
        </p>
      )}
      {binding.error && (
        <p id={errorId} className="text-sm font-medium text-danger-fg">
          {binding.error}
        </p>
      )}
      {options.map((option, i) => (
        <label
          key={option.value}
          className="flex min-h-6 w-fit min-w-6 items-center gap-2 text-sm text-fg has-disabled:opacity-50"
        >
          <input
            type="checkbox"
            id={i === 0 ? id : `${id}-${i}`}
            name={name}
            value={option.value}
            defaultChecked={defaultValue.includes(option.value)}
            className={cn("size-4 accent-primary disabled:cursor-not-allowed", focusRing)}
          />
          {option.label}
        </label>
      ))}
    </fieldset>
  );
}
