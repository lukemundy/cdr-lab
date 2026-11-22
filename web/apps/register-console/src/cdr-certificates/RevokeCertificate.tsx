import { Button } from "@cdr-lab/ui/button";
import { ConfirmationDialog } from "@cdr-lab/ui/confirmation-dialog";
import { Dialog, DialogClose, DialogFooter } from "@cdr-lab/ui/dialog";
import { FileInput } from "@cdr-lab/ui/file-input";
import { ErrorSummary, type FormRefusal } from "@cdr-lab/ui/form";
import { Select, TextArea } from "@cdr-lab/ui/form-field";
import { LocalTime } from "@cdr-lab/ui/local-time";
import type { Problem } from "@cdr-lab/ui/states";
import { useToast } from "@cdr-lab/ui/toast";
import { Ban } from "lucide-react";
import { useId, useState, type FormEvent } from "react";

import { ProblemError, refusalOf } from "@cdr-lab/web-platform/problem";

import type {
  RevocableCertificate,
  RevocationReason,
  RevocationRequest,
  RevocationValues,
} from "../api/client";
import { reasonWords } from "./reasons";

const unexpectedProblem: Problem = {
  title: "The action did not complete",
  detail: "Refresh the page to see the current state, then try again.",
};

const reasons = Object.entries(reasonWords).map(([value, label]) => ({ value, label }));

function reasonOf(value: string): RevocationReason | undefined {
  return Object.keys(reasonWords).find((reason): reason is RevocationReason => reason === value);
}

export type RevokeCertificateProps = {
  // Throws a ProblemError when the Register refuses the request.
  check: (values: RevocationValues) => Promise<RevocableCertificate>;
  revoke: (request: RevocationRequest) => Promise<void>;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
  // Resolves when the list of revocations shows its first page.
  showNewest: () => Promise<unknown>;
};

type Checked = { certificate: RevocableCertificate; request: RevocationRequest };

export function RevokeCertificate({ check, revoke, refresh, showNewest }: RevokeCertificateProps) {
  const toast = useToast();
  const id = useId();
  const certificateId = `${id}-certificate`;
  const reasonId = `${id}-reason`;
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [certificate, setCertificate] = useState("");
  const [reason, setReason] = useState("");
  // The count gives the summary a new key at each refusal, so that the summary gets the focus again.
  const [refusal, setRefusal] = useState<{ value: FormRefusal; count: number }>();
  const [checked, setChecked] = useState<Checked>();

  const errorOf = (field: string) => {
    const messages = (refusal?.value.fieldErrors ?? [])
      .filter((e) => e.pointer === `#/${field}`)
      .map((e) => e.detail);
    return messages.length > 0 ? messages.join(" ") : undefined;
  };
  const targets: Record<string, string> = {
    "#/certificate": certificateId,
    "#/reason": reasonId,
  };

  function changeOpen(next: boolean) {
    if (busy) return;
    if (next) {
      setCertificate("");
      setReason("");
      setRefusal(undefined);
      setChecked(undefined);
    }
    setOpen(next);
  }

  function refuse(value: FormRefusal) {
    setRefusal((previous) => ({ value, count: (previous?.count ?? 0) + 1 }));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) return;
    setBusy(true);
    const values: RevocationValues = { certificate, reason: reasonOf(reason) };
    let result: RevocableCertificate;
    try {
      result = await check(values);
    } catch (error) {
      await refresh();
      refuse(error instanceof ProblemError ? refusalOf(error) : { problem: unexpectedProblem });
      setBusy(false);
      return;
    }
    setRefusal(undefined);
    setBusy(false);
    // The check refuses a request with no reason, so a successful check has a reason.
    if (values.reason) {
      setChecked({ certificate: result, request: { certificate, reason: values.reason } });
    }
  }

  async function confirm(): Promise<Problem | undefined> {
    if (!checked) return undefined;
    try {
      await revoke(checked.request);
    } catch (error) {
      await refresh();
      return refusalOf(error).problem;
    }
    // The list goes to its first page first, so that the refresh gets only that page.
    await showNewest();
    await refresh();
    setOpen(false);
    toast.show({ title: `Certificate ${checked.certificate.serialNumber} revoked` });
    return undefined;
  }

  return (
    <Dialog
      title="Revoke a CDR certificate"
      description="The mock CDR intermediate revokes the certificate, and its CRL lists it."
      trigger={
        <Button variant="secondary" icon={Ban}>
          Revoke certificate
        </Button>
      }
      open={open}
      onOpenChange={changeOpen}
    >
      <form
        aria-label="Revoke a CDR certificate"
        noValidate
        onSubmit={(event) => void submit(event)}
        className="flex flex-col gap-4"
      >
        {refusal && (
          <ErrorSummary
            key={refusal.count}
            detail={refusal.value.problem.detail ?? refusal.value.problem.title}
            errors={(refusal.value.fieldErrors ?? []).map((e) => ({
              target: targets[e.pointer],
              message: e.detail,
            }))}
          />
        )}
        <TextArea
          id={certificateId}
          label="PEM certificate"
          description="Paste one PEM certificate, or choose the file of an issue. The Register revokes the first certificate."
          name="certificate"
          monospace
          rows={10}
          spellCheck={false}
          value={certificate}
          onChange={(event) => setCertificate(event.currentTarget.value)}
          error={errorOf("certificate")}
        />
        <FileInput
          label="Certificate file"
          description="The text of the file goes into the PEM certificate field."
          accept=".pem,.crt,.cer"
          onFileText={setCertificate}
        />
        <Select
          id={reasonId}
          label="Reason"
          name="reason"
          placeholder="Choose a reason"
          options={reasons}
          value={reason}
          onChange={(event) => setReason(event.currentTarget.value)}
          error={errorOf("reason")}
        />
        <output className="sr-only">{busy && "Continue in progress"}</output>
        <DialogFooter>
          <DialogClose disabled={busy}>Cancel</DialogClose>
          <Button type="submit" busy={busy}>
            Continue
          </Button>
        </DialogFooter>
      </form>
      {checked && (
        <ConfirmationDialog
          open
          onOpenChange={(next) => {
            if (!next) setChecked(undefined);
          }}
          title="Revoke this CDR certificate?"
          action="Revoke"
          irreversible
          effects={[
            <>
              The Register revokes the certificate{" "}
              <span className="font-medium text-fg">{checked.certificate.subject}</span>, with the
              serial number{" "}
              <code className="font-mono text-xs">{checked.certificate.serialNumber}</code>, for the
              reason &quot;{reasonWords[checked.request.reason]}&quot;.
            </>,
            <>
              The thumbprint of the certificate is{" "}
              <code className="font-mono text-xs break-all">{checked.certificate.thumbprint}</code>,
              and its validity ends at <LocalTime value={checked.certificate.notAfter} />.
            </>,
            "The CRL lists the certificate. The Register refuses it at once, and other relying parties refuse it within 1 minute.",
            "A registration of the certificate for a software product does not change.",
          ]}
          onConfirm={confirm}
        />
      )}
    </Dialog>
  );
}
