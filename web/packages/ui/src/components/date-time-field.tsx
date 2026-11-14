import { Field } from "@base-ui/react/field";
import type { ComponentProps } from "react";

import { cn } from "../lib/cn";
import { useFormBinding } from "../lib/form-context";
import { controlClasses, FormField, type FormFieldProps } from "./form-field";

export type DateTimeFieldProps = Omit<FormFieldProps, "children"> &
  Omit<ComponentProps<"input">, "type" | "className" | "disabled" | "step">;

// A native date-time control in the time zone of the browser, with seconds. utcOf gives its value in UTC.
export function DateTimeField({
  label,
  description,
  error,
  disabled,
  className,
  id,
  ...props
}: DateTimeFieldProps) {
  const binding = useFormBinding(props.name, error, id);
  return (
    <FormField
      label={label}
      description={description}
      error={binding.error}
      disabled={disabled}
      className={className}
    >
      <Field.Control
        type="datetime-local"
        step={1}
        className={cn(controlClasses, "h-control w-auto")}
        id={binding.id}
        {...props}
      />
    </FormField>
  );
}

// A value with no offset is a local time (ECMA-262), and a value that is not a time stays as it is for the server to refuse.
export function utcOf(local: string): string {
  if (local === "") return "";
  const date = new Date(local);
  return Number.isNaN(date.getTime()) ? local : date.toISOString();
}
