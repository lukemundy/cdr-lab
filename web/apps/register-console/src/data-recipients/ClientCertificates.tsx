import { Alert } from "@cdr-lab/ui/alert";
import { Button } from "@cdr-lab/ui/button";
import { ConfirmationDialog } from "@cdr-lab/ui/confirmation-dialog";
import { Dialog, DialogClose, DialogFooter } from "@cdr-lab/ui/dialog";
import { FileInput } from "@cdr-lab/ui/file-input";
import { TextArea } from "@cdr-lab/ui/form-field";
import { Identifier } from "@cdr-lab/ui/identifier";
import type { Problem } from "@cdr-lab/ui/states";
import { useToast } from "@cdr-lab/ui/toast";
import { FileKey, Trash2 } from "lucide-react";
import { useRef, useState, type FormEvent } from "react";

import { ProblemError } from "@cdr-lab/console/problem";

const unexpectedProblem: Problem = {
  title: "The action did not complete",
  detail: "Refresh the page to see the current state, then try again.",
};

export type RegisterCertificateProps = {
  productName: string;
  permitted: boolean;
  // Throws a ProblemError when the Register refuses the certificate.
  run: (certificate: string) => Promise<void>;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
};

export function RegisterCertificate({
  productName,
  permitted,
  run,
  refresh,
}: RegisterCertificateProps) {
  const toast = useToast();
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [certificate, setCertificate] = useState("");
  const [fieldError, setFieldError] = useState<string>();
  const [problem, setProblem] = useState<Problem>();

  // The refresh after a refused registration can remove the action, so its dialog stays until the operator closes it.
  if (!permitted && !open) return null;

  function changeOpen(next: boolean) {
    if (busy) return;
    if (next) {
      setCertificate("");
      setFieldError(undefined);
      setProblem(undefined);
    }
    setOpen(next);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setFieldError(undefined);
    setProblem(undefined);
    try {
      await run(certificate);
    } catch (error) {
      if (error instanceof ProblemError && error.status === 400) {
        setFieldError(error.problem.detail ?? error.problem.title);
      } else {
        await refresh();
        setProblem(error instanceof ProblemError ? error.problem : unexpectedProblem);
      }
      setBusy(false);
      return;
    }
    await refresh();
    setBusy(false);
    setOpen(false);
    toast.show({ title: `Client certificate registered for ${productName}` });
  }

  return (
    <Dialog
      title={`Register a client certificate for ${productName}`}
      description="The token endpoint then issues tokens to a client that uses this certificate."
      trigger={
        <Button variant="secondary" icon={FileKey}>
          Register certificate
        </Button>
      }
      open={open}
      onOpenChange={changeOpen}
    >
      <form noValidate onSubmit={(event) => void submit(event)} className="flex flex-col gap-4">
        <TextArea
          label="PEM certificate"
          description="Paste one PEM certificate, or choose a file. The Register checks the certificate."
          name="certificate"
          monospace
          rows={10}
          spellCheck={false}
          value={certificate}
          onChange={(event) => setCertificate(event.currentTarget.value)}
          error={fieldError}
        />
        <FileInput
          label="Certificate file"
          description="The text of the file goes into the PEM certificate field."
          accept=".pem,.crt,.cer"
          onFileText={setCertificate}
        />
        {problem && (
          <Alert title={problem.title}>{problem.detail && <p>{problem.detail}</p>}</Alert>
        )}
        <DialogFooter>
          <DialogClose disabled={busy}>Cancel</DialogClose>
          <Button type="submit" busy={busy}>
            Register
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

export type CertificateListProps = {
  productName: string;
  thumbprints: string[];
  removable: boolean;
  // Throws a ProblemError when the Register refuses the removal.
  remove: (thumbprint: string) => Promise<void>;
  refresh: () => Promise<unknown>;
};

export function CertificateList({
  productName,
  thumbprints,
  removable,
  remove,
  refresh,
}: CertificateListProps) {
  const toast = useToast();
  const list = useRef<HTMLUListElement>(null);
  // The refresh after a refused removal can remove the thumbprint, so its dialog stays until the operator closes it.
  const [open, setOpen] = useState<string>();
  const shown = open && !thumbprints.includes(open) ? [...thumbprints, open] : thumbprints;

  if (shown.length === 0) return <span className="text-fg-muted">Not set</span>;

  async function confirm(thumbprint: string) {
    try {
      await remove(thumbprint);
    } catch (error) {
      if (!(error instanceof ProblemError)) throw error;
      await refresh();
      return error.problem;
    }
    await refresh();
    toast.show({ title: `Client certificate removed from ${productName}` });
    return undefined;
  }

  return (
    <ul ref={list} tabIndex={-1} className="flex flex-col gap-1 outline-none">
      {shown.map((thumbprint) => (
        <li key={thumbprint} className="flex items-center gap-2">
          <Identifier
            value={thumbprint}
            label={`client certificate thumbprint of ${productName}`}
          />
          {removable && (
            <ConfirmationDialog
              trigger={
                <Button
                  variant="secondary"
                  icon={Trash2}
                  aria-label={`Remove the client certificate ${thumbprint}`}
                >
                  Remove
                </Button>
              }
              title={`Remove the client certificate of ${productName}?`}
              action="Remove"
              effects={[
                `The token endpoint refuses the client certificate ${thumbprint} for ${productName} at the next token request.`,
                "Access tokens that exist stay valid until they expire.",
                "To trust the certificate again, register its PEM certificate again.",
              ]}
              onConfirm={() => confirm(thumbprint)}
              onOpenChange={(isOpen) => setOpen(isOpen ? thumbprint : undefined)}
              returnFocus={list}
            />
          )}
        </li>
      ))}
    </ul>
  );
}
