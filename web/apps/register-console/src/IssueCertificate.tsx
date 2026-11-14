import { Alert } from "@cdr-lab/ui/alert";
import { Button } from "@cdr-lab/ui/button";
import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Dialog, DialogClose, DialogFooter } from "@cdr-lab/ui/dialog";
import { FileInput } from "@cdr-lab/ui/file-input";
import { TextArea } from "@cdr-lab/ui/form-field";
import { Identifier } from "@cdr-lab/ui/identifier";
import type { Problem } from "@cdr-lab/ui/states";
import { Download, FileBadge } from "lucide-react";
import { useEffect, useRef, useState, type FormEvent } from "react";

import { ProblemError } from "@cdr-lab/console/problem";

import type { CertificateIssue } from "./api/client";
import { saveFile } from "./save-file";

const unexpectedProblem: Problem = {
  title: "The action did not complete",
  detail: "Refresh the page to see the current state, then try again.",
};

const kinds = {
  client: {
    name: "client certificate",
    Name: "Client certificate",
    registration: (record: string) =>
      `The Register registered the certificate for ${record}. The token endpoint accepts it now.`,
  },
  server: {
    name: "server certificate",
    Name: "Server certificate",
    registration: () =>
      "The Register does not register a server certificate. A TLS client that trusts the CDR CA bundle accepts it.",
  },
};

export type IssueCertificateProps = {
  kind: keyof typeof kinds;
  recordName: string;
  // The subject that the Register requires, with the names of the record.
  subject: string;
  fileName: string;
  permitted: boolean;
  // Throws a ProblemError when the Register refuses the CSR.
  run: (csr: string) => Promise<CertificateIssue>;
  // Resolves when the views show the current state.
  refresh: () => Promise<unknown>;
};

export function IssueCertificate({
  kind,
  recordName,
  subject,
  fileName,
  permitted,
  run,
  refresh,
}: IssueCertificateProps) {
  const text = kinds[kind];
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [csr, setCSR] = useState("");
  const [fieldError, setFieldError] = useState<string>();
  const [problem, setProblem] = useState<Problem>();
  const [result, setResult] = useState<CertificateIssue>();
  const resultHeading = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    if (result) resultHeading.current?.focus();
  }, [result]);

  // The refresh after an issue can remove the action, so its dialog stays until the operator closes it.
  if (!permitted && !open) return null;

  function changeOpen(next: boolean) {
    if (busy) return;
    if (next) {
      setCSR("");
      setFieldError(undefined);
      setProblem(undefined);
      setResult(undefined);
    }
    setOpen(next);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setFieldError(undefined);
    setProblem(undefined);
    let issued: CertificateIssue;
    try {
      issued = await run(csr);
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
    setResult(issued);
    setBusy(false);
    await refresh();
  }

  return (
    <Dialog
      title={`Issue a ${text.name} for ${recordName}`}
      description="The mock CDR intermediate CA issues the certificate from a CSR."
      trigger={
        <Button variant="secondary" icon={FileBadge}>
          Issue certificate
        </Button>
      }
      open={open}
      onOpenChange={changeOpen}
    >
      {result ? (
        <div className="flex flex-col gap-4">
          <h3 ref={resultHeading} tabIndex={-1} className="font-semibold outline-none">
            {text.Name} issued for {recordName}
          </h3>
          <p>{text.registration(recordName)}</p>
          <DescriptionList>
            <DescriptionItem term="Thumbprint">
              <Identifier value={result.thumbprint} label={`thumbprint of the ${text.name}`} />
            </DescriptionItem>
          </DescriptionList>
          <TextArea
            label="PEM chain"
            description="The certificate and then the intermediate certificate."
            monospace
            rows={10}
            readOnly
            spellCheck={false}
            value={result.certificateChain}
          />
          <p className="font-medium">
            The Register does not keep the certificate, so the console cannot show it again. Save
            the chain before you close this dialog.
          </p>
          <DialogFooter>
            <Button
              variant="secondary"
              icon={Download}
              onClick={() => saveFile(fileName, result.certificateChain)}
            >
              Save as file
            </Button>
            <DialogClose variant="primary">Close</DialogClose>
          </DialogFooter>
        </div>
      ) : (
        <form noValidate onSubmit={(event) => void submit(event)} className="flex flex-col gap-4">
          <TextArea
            label="PEM CSR"
            description={`Paste one PEM CSR, or choose a file. The Register needs only the CSR, and never the private key. ${subject}`}
            name="csr"
            monospace
            rows={10}
            spellCheck={false}
            value={csr}
            onChange={(event) => setCSR(event.currentTarget.value)}
            error={fieldError}
          />
          <FileInput
            label="CSR file"
            description="The text of the file goes into the PEM CSR field."
            accept=".pem,.csr,.req"
            onFileText={setCSR}
          />
          {problem && (
            <Alert title={problem.title}>{problem.detail && <p>{problem.detail}</p>}</Alert>
          )}
          <DialogFooter>
            <DialogClose disabled={busy}>Cancel</DialogClose>
            <Button type="submit" busy={busy}>
              Issue
            </Button>
          </DialogFooter>
        </form>
      )}
    </Dialog>
  );
}
