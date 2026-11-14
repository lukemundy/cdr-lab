import { Field } from "@base-ui/react/field";
import type { ComponentProps, ReactNode } from "react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";
import { useFormBinding } from "../lib/form-context";

// From the shadcn/ui registry 4.22.1, style base-nova, items input and textarea, changed to the semantic tokens.
export const controlClasses = cn(
  "w-full min-w-0 rounded-md border border-border-strong bg-surface px-control-x text-sm text-fg",
  "aria-invalid:border-danger-fg disabled:cursor-not-allowed disabled:opacity-50",
  focusRing,
);

export type FormFieldProps = {
  label: string;
  description?: ReactNode;
  // An error from the app or from the server for the value. The field is invalid while it is set. In a form, the
  // field also gets the error of its name from the server.
  error?: string | undefined;
  disabled?: boolean | undefined;
  className?: string | undefined;
  // One Base UI Field control, for example TextField, TextArea or FileInput.
  children: ReactNode;
};

export function FormField({
  label,
  description,
  error,
  disabled,
  className,
  children,
}: FormFieldProps) {
  const invalid = Boolean(error);
  return (
    <Field.Root
      invalid={invalid}
      disabled={disabled}
      className={cn("flex flex-col gap-1", className)}
    >
      <Field.Label className="text-sm font-medium text-fg">{label}</Field.Label>
      {description && (
        <Field.Description className="text-sm text-fg-muted">{description}</Field.Description>
      )}
      <Field.Error match={invalid} className="text-sm font-medium text-danger-fg">
        {error}
      </Field.Error>
      {children}
    </Field.Root>
  );
}

type FieldTextProps = Omit<FormFieldProps, "children">;

export type TextFieldProps = FieldTextProps &
  Omit<ComponentProps<"input">, "type" | "className" | "disabled"> & {
    type?: "text" | "email" | "url" | "password";
  };

export function TextField({
  label,
  description,
  error,
  disabled,
  className,
  type = "text",
  id,
  ...props
}: TextFieldProps) {
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
        type={type}
        className={cn(controlClasses, "h-control")}
        id={binding.id}
        {...props}
      />
    </FormField>
  );
}

export type TextAreaProps = FieldTextProps &
  Omit<ComponentProps<"textarea">, "className" | "disabled"> & {
    monospace?: boolean;
  };

export function TextArea({
  label,
  description,
  error,
  disabled,
  className,
  monospace = false,
  rows = 6,
  id,
  ...props
}: TextAreaProps) {
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
        id={binding.id}
        render={
          <textarea
            rows={rows}
            className={cn(controlClasses, "py-2", monospace && "font-mono text-xs")}
            {...props}
          />
        }
      />
    </FormField>
  );
}

export type SelectOption = { value: string; label: string };

export type SelectProps = FieldTextProps &
  Omit<ComponentProps<"select">, "className" | "disabled" | "children" | "multiple"> & {
    options: readonly SelectOption[];
    // The text of an empty first option, so that the user and not the UI chooses the value.
    placeholder?: string | undefined;
  };

export function Select({
  label,
  description,
  error,
  disabled,
  className,
  options,
  placeholder,
  id,
  ...props
}: SelectProps) {
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
        id={binding.id}
        render={
          <select className={cn(controlClasses, "h-control")} {...props}>
            {placeholder !== undefined && <option value="">{placeholder}</option>}
            {options.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        }
      />
    </FormField>
  );
}
