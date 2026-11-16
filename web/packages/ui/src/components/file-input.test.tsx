import { useState } from "react";
import { afterEach, expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { FileInput } from "./file-input";
import { TextArea } from "./form-field";

const pem = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n";

afterEach(() => {
  vi.restoreAllMocks();
});

function certificateFile(): File {
  return new File([pem], "client.crt", { type: "application/x-x509-ca-cert" });
}

function CertificateForm() {
  const [certificate, setCertificate] = useState("");
  return (
    <>
      <FileInput
        label="Certificate file"
        description="A PEM file. Its text goes into the certificate field."
        accept=".pem,.crt,.cer"
        onFileText={setCertificate}
      />
      <TextArea
        label="Certificate"
        value={certificate}
        onChange={(event) => setCertificate(event.target.value)}
      />
    </>
  );
}

// Verifies: UI-0318
test("a file input is a file control with a visible label and help text", async () => {
  const container = render(<CertificateForm />);
  const input = page.getByLabelText("Certificate file");

  await expect.element(page.getByText("Certificate file", { exact: true })).toBeVisible();
  await expect.element(input).toHaveAttribute("type", "file");
  await expect.element(input).toHaveAttribute("accept", ".pem,.crt,.cer");
  await expect
    .element(input)
    .toHaveAccessibleDescription("A PEM file. Its text goes into the certificate field.");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0320
test("the text of the selected file goes into the text area, where the user can change it", async () => {
  const container = render(<CertificateForm />);
  const input = page.getByLabelText("Certificate file");
  const certificate = page.getByRole("textbox", { name: "Certificate" });

  await userEvent.upload(input, certificateFile());

  // The browser shows the file name outside the DOM, so the test checks the selected file.
  const element = input.element();
  if (!(element instanceof HTMLInputElement)) throw new Error("not an input");
  expect(element.files?.[0]?.name).toBe("client.crt");
  await expect.element(certificate).toHaveValue(pem);
  await userEvent.type(certificate, "x");
  await expect.element(certificate).toHaveValue(`${pem}x`);
  await expectNoAxeViolations(container);
});

test("a file that the browser cannot read shows an error at the field", async () => {
  const texts: string[] = [];
  render(<FileInput label="Certificate file" onFileText={(text) => texts.push(text)} />);
  const input = page.getByLabelText("Certificate file");
  // The upload gives the page a new File object, so the test makes each read fail.
  vi.spyOn(Blob.prototype, "text").mockRejectedValue(new DOMException("", "NotReadableError"));

  await userEvent.upload(input, certificateFile());

  await expect.element(input).toHaveAttribute("aria-invalid", "true");
  await expect
    .element(input)
    .toHaveAccessibleDescription("Could not read the file. Choose the file again.");
  expect(texts).toEqual([]);
});

// Verifies: UI-0302
test("an error that the app gives shows at the field", async () => {
  const container = render(
    <FileInput label="Certificate file" error="Choose a file." onFileText={() => {}} />,
  );
  const input = page.getByLabelText("Certificate file");

  await expect.element(input).toHaveAttribute("aria-invalid", "true");
  await expect.element(input).toHaveAccessibleDescription("Choose a file.");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0201, UI-0202
test("the keyboard reaches the file input with a visible focus", async () => {
  render(<FileInput label="Certificate file" onFileText={() => {}} />);

  await userEvent.tab();

  expectVisibleFocus(page.getByLabelText("Certificate file").element());
});

// A file chooser opens outside the page, so the test checks the activation that opens it, and stops the chooser.
// Verifies: UI-0319
test("Space and Enter activate the file input, which opens the file chooser", async () => {
  render(<FileInput label="Certificate file" onFileText={() => {}} />);
  const input = page.getByLabelText("Certificate file").element();
  let activations = 0;
  input.addEventListener("click", (event) => {
    event.preventDefault();
    activations++;
  });

  await userEvent.tab();
  await userEvent.keyboard(" ");
  await userEvent.keyboard("{Enter}");

  expect(activations).toBe(2);
});

test("a disabled file input passes the accessibility checks", async () => {
  const container = render(<FileInput label="Certificate file" disabled onFileText={() => {}} />);

  await expect.element(page.getByLabelText("Certificate file")).toBeDisabled();
  await expectNoAxeViolations(container);
});

// Verifies: UI-0094
test("the file input is 24 by 24 CSS px or more", () => {
  render(<FileInput label="Certificate file" onFileText={() => {}} />);

  expectTargetSize(page.getByLabelText("Certificate file").element());
});
