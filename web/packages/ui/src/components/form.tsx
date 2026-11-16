import {
  useEffect,
  useId,
  useRef,
  useState,
  type FormEvent,
  type MouseEvent,
  type ReactElement,
  type ReactNode,
} from "react";

import { cn } from "../lib/cn";
import { FormContext, type FormContextValue } from "../lib/form-context";
import { Button } from "./button";
import { Link } from "./link";
import type { Problem } from "./states";

// A field error of an RFC 9457 problem: a JSON Pointer to a value of the request body, in the URI fragment form.
export type FieldErrorItem = { pointer: string; detail: string };

export type FormRefusal = { problem: Problem; fieldErrors?: readonly FieldErrorItem[] | undefined };

// design/forms-and-dialogs.md, UI-0332: a form that fails with no problem details.
const unexpectedProblem: Problem = {
  title: "The change did not complete",
  detail: "The change did not complete. Refresh the page to see the current state, then try again.",
};

// The error of an item of a list or of a value in a group shows at the field of the list or of the group.
function belongsTo(pointer: string, name: string): boolean {
  const path = pointer.replace(/^#\//, "");
  return path === name || path.startsWith(`${name}/`);
}

// The first control whose value has the error, for the link of the error summary.
function controlOf(form: HTMLFormElement, pointer: string): string | undefined {
  const control = [...form.elements].find(
    (element) =>
      (element instanceof HTMLInputElement ||
        element instanceof HTMLSelectElement ||
        element instanceof HTMLTextAreaElement) &&
      element.name !== "" &&
      belongsTo(pointer, element.name),
  );
  return control?.id || undefined;
}

// A message with no target names a value that has no field in the form, so it has no link.
export type ErrorSummaryItem = { target?: string | undefined; message: string };

export type ErrorSummaryProps = {
  // The error of the form, which names no field.
  detail?: string | undefined;
  // Each field that has an error: the id of its control and the text of its error message.
  errors: readonly ErrorSummaryItem[];
};

// The fragment of the link stays out of the URL, which the router of the UI owns.
function moveFocus(event: MouseEvent<HTMLAnchorElement>, target: string) {
  const field = document.getElementById(target);
  if (!field) return;
  event.preventDefault();
  field.focus();
}

// From the GOV.UK Design System error summary, changed to the semantic tokens.
export function ErrorSummary({ detail, errors }: ErrorSummaryProps) {
  const summary = useRef<HTMLDivElement>(null);

  useEffect(() => {
    summary.current?.focus();
  }, []);

  return (
    <div
      ref={summary}
      tabIndex={-1}
      className={cn(
        "rounded-md border-2 border-danger-fg bg-danger-bg p-4 text-danger-fg",
        "focus:outline-2 focus:outline-offset-2 focus:outline-focus",
      )}
    >
      <div role="alert" className="flex flex-col gap-2">
        <h2 className="text-base font-semibold">There is a problem</h2>
        {detail && <p className="text-sm whitespace-pre-line">{detail}</p>}
        {errors.length > 0 && (
          <ul className="flex flex-col gap-1 text-sm">
            {errors.map(({ target, message }) => (
              <li key={`${target} ${message}`}>
                {target ? (
                  <Link
                    href={`#${target}`}
                    onClick={(event) => moveFocus(event, target)}
                    className="font-medium text-danger-fg hover:text-danger-fg"
                  >
                    {message}
                  </Link>
                ) : (
                  message
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

export type FormProps = {
  // The accessible name of the form.
  label: string;
  // The name of the action, for example "Save changes".
  submitLabel: string;
  // The element of the Cancel link, for example the link of a router, back to the page that opened the form.
  cancel: ReactElement;
  // Resolves with a refusal to keep the form, or with undefined after a success. The form does not check the values.
  onSubmit: (data: FormData) => Promise<FormRefusal | undefined>;
  children: ReactNode;
};

export function Form({ label, submitLabel, cancel, onSubmit, children }: FormProps) {
  const formId = useId();
  const [busy, setBusy] = useState(false);
  const running = useRef(false);
  // The count gives the summary a new key at each refusal, so that the summary gets the focus again.
  const [shown, setShown] = useState<{
    refusal: FormRefusal;
    summary: ErrorSummaryItem[];
    count: number;
  }>();
  const fieldErrors = shown?.refusal.fieldErrors ?? [];

  const context: FormContextValue = {
    idOf: (name) => `${formId}-${name}`,
    errorOf: (name) => {
      const messages = fieldErrors.filter((e) => belongsTo(e.pointer, name)).map((e) => e.detail);
      return messages.length > 0 ? messages.join(" ") : undefined;
    },
  };

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (running.current) return;
    running.current = true;
    setBusy(true);
    const form = event.currentTarget;
    const data = new FormData(form);
    let result: FormRefusal | undefined;
    try {
      result = await onSubmit(data);
    } catch {
      result = { problem: unexpectedProblem };
    }
    running.current = false;
    setBusy(false);
    const refusal = result;
    setShown(
      (previous) =>
        refusal && {
          refusal,
          summary: (refusal.fieldErrors ?? []).map((e) => ({
            target: controlOf(form, e.pointer),
            message: e.detail,
          })),
          count: (previous?.count ?? 0) + 1,
        },
    );
  }

  return (
    <FormContext value={context}>
      <form
        aria-label={label}
        noValidate
        onSubmit={(event) => void submit(event)}
        className="flex max-w-2xl flex-col gap-4"
      >
        {shown && (
          <ErrorSummary
            key={shown.count}
            detail={shown.refusal.problem.detail ?? shown.refusal.problem.title}
            errors={shown.summary}
          />
        )}
        {children}
        <output className="sr-only">{busy && `${submitLabel} in progress`}</output>
        <div className="flex items-center gap-4 pt-2">
          <Button type="submit" busy={busy}>
            {submitLabel}
          </Button>
          <Link render={cancel}>Cancel</Link>
        </div>
      </form>
    </FormContext>
  );
}
