import { Field } from "@base-ui/react/field";
import { useState, type ChangeEvent, type ComponentProps } from "react";

import { cn } from "../lib/cn";
import { controlClasses, FormField, type FormFieldProps } from "./form-field";

const readError = "Could not read the file. Choose the file again.";

export type FileInputProps = Omit<FormFieldProps, "children"> &
  Pick<ComponentProps<"input">, "accept" | "name"> & {
    // The text of the selected file, decoded as UTF-8.
    onFileText: (text: string) => void;
  };

export function FileInput({
  label,
  description,
  error,
  disabled,
  className,
  accept,
  name,
  onFileText,
}: FileInputProps) {
  const [failed, setFailed] = useState(false);

  async function change(event: ChangeEvent<HTMLInputElement>) {
    const file = event.currentTarget.files?.[0];
    setFailed(false);
    if (!file) return;
    try {
      onFileText(await file.text());
    } catch {
      setFailed(true);
    }
  }

  return (
    <FormField
      label={label}
      description={description}
      error={failed ? readError : error}
      disabled={disabled}
      className={className}
    >
      <Field.Control
        type="file"
        accept={accept}
        name={name}
        onChange={(event) => void change(event)}
        className={cn(
          controlClasses,
          "h-control cursor-pointer py-1 text-fg-muted",
          "file:mr-3 file:h-full file:cursor-pointer file:rounded-sm file:border file:border-border-strong",
          "file:bg-secondary file:px-2 file:text-sm file:font-medium file:text-fg",
        )}
      />
    </FormField>
  );
}
