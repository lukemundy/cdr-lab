import { useState } from "react";
import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { Select, TextArea, TextField } from "./form-field";

test("a text field has a visible label, and the screen reader reads the label and the help text", async () => {
  const container = render(
    <TextField label="Software product name" description="The name in the SSA." />,
  );
  const field = page.getByRole("textbox", { name: "Software product name" });

  await expect.element(page.getByText("Software product name", { exact: true })).toBeVisible();
  await expect.element(page.getByText("The name in the SSA.")).toBeVisible();
  await expect.element(field).toHaveAccessibleDescription("The name in the SSA.");
  await expect.element(field).not.toHaveAttribute("aria-invalid");
  await expectNoAxeViolations(container);
});

test("a password field hides its value, and a password manager can fill it", async () => {
  const container = render(
    <TextField label="Password" name="password" type="password" autoComplete="current-password" />,
  );
  const field = page.getByLabelText("Password");

  await expect.element(field).toHaveAttribute("type", "password");
  await expect.element(field).toHaveAttribute("autocomplete", "current-password");
  await expectNoAxeViolations(container);
});

test("a text field with no help text has no description", async () => {
  render(<TextField label="Name" />);

  await expect.element(page.getByRole("textbox", { name: "Name" })).toHaveAccessibleDescription("");
});

test("an error shows at the field, and the field is invalid and described by the error", async () => {
  const container = render(
    <TextField label="Name" description="The name in the SSA." error="Enter a name." />,
  );
  const field = page.getByRole("textbox", { name: "Name" });

  await expect.element(page.getByText("Enter a name.")).toBeVisible();
  await expect.element(field).toHaveAttribute("aria-invalid", "true");
  await expect.element(field).toHaveAccessibleDescription("The name in the SSA. Enter a name.");
  await expectNoAxeViolations(container);
});

test("the error shows above the control, after the label and the help text", () => {
  render(<TextField label="Name" description="Help" error="Enter a name." />);
  const elements = [
    page.getByText("Name", { exact: true }).element(),
    page.getByText("Help", { exact: true }).element(),
    page.getByText("Enter a name.", { exact: true }).element(),
    page.getByRole("textbox", { name: "Name" }).element(),
  ];

  for (let i = 1; i < elements.length; i++) {
    const [before, after] = [elements[i - 1], elements[i]];
    if (!before || !after) throw new Error("no element");
    expect(before.compareDocumentPosition(after)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);
  }
});

// The error of the server comes after the user typed the value, so the error prop changes on a field that exists.
function ServerErrorForm() {
  const [error, setError] = useState<string | undefined>();
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        setError("The body is not one PEM certificate.");
      }}
    >
      <TextArea label="Certificate" error={error} onChange={() => setError(undefined)} />
      <button type="submit">Register</button>
    </form>
  );
}

test("an error from the server shows at the field, and goes when the user changes the value", async () => {
  const container = render(<ServerErrorForm />);
  const field = page.getByRole("textbox", { name: "Certificate" });

  await userEvent.type(field, "not a certificate");
  await page.getByRole("button", { name: "Register" }).click();

  await expect.element(field).toHaveAttribute("aria-invalid", "true");
  await expect.element(field).toHaveAccessibleDescription("The body is not one PEM certificate.");
  await expectNoAxeViolations(container);

  await userEvent.type(field, "x");

  await expect.element(field).not.toHaveAttribute("aria-invalid");
  await expect.element(field).toHaveAccessibleDescription("");
  await expect
    .element(page.getByText("The body is not one PEM certificate."))
    .not.toBeInTheDocument();
});

test("a text area is a multi-line text box with its label, help text and error", async () => {
  const container = render(
    <TextArea
      label="Certificate"
      description="One PEM certificate."
      error="The body is not one PEM certificate."
      monospace
    />,
  );
  const field = page.getByRole("textbox", { name: "Certificate" });

  expect(field.element().tagName).toBe("TEXTAREA");
  await expect
    .element(field)
    .toHaveAccessibleDescription("One PEM certificate. The body is not one PEM certificate.");
  expect(getComputedStyle(field.element()).fontFamily).toMatch(/JetBrains Mono/);
  await expectNoAxeViolations(container);
});

test("a text area keeps line breaks, and Enter adds a line", async () => {
  render(<TextArea label="Certificate" />);
  const field = page.getByRole("textbox", { name: "Certificate" });

  await userEvent.type(field, "first{Enter}second");

  await expect.element(field).toHaveValue("first\nsecond");
});

test("a disabled field is not in the tab sequence, and passes the accessibility checks", async () => {
  const container = render(
    <>
      <TextField label="Name" disabled defaultValue="Mock Product" />
      <TextArea label="Certificate" disabled />
    </>,
  );

  await expect.element(page.getByRole("textbox", { name: "Name" })).toBeDisabled();
  await expect.element(page.getByRole("textbox", { name: "Certificate" })).toBeDisabled();
  await userEvent.tab();
  expect(document.activeElement).toBe(document.body);
  await expectNoAxeViolations(container);
});

test("the keyboard reaches each field with a visible focus", async () => {
  render(
    <>
      <TextField label="Name" />
      <TextArea label="Certificate" />
    </>,
  );

  await userEvent.tab();
  expectVisibleFocus(page.getByRole("textbox", { name: "Name" }).element());
  await userEvent.keyboard("Mock");
  await expect.element(page.getByRole("textbox", { name: "Name" })).toHaveValue("Mock");
  await userEvent.tab();
  expectVisibleFocus(page.getByRole("textbox", { name: "Certificate" }).element());
});

test("each field is 24 by 24 CSS px or more", () => {
  render(
    <>
      <TextField label="Name" />
      <TextArea label="Certificate" />
    </>,
  );

  expectTargetSize(page.getByRole("textbox", { name: "Name" }).element());
  expectTargetSize(page.getByRole("textbox", { name: "Certificate" }).element());
});

const levels = [
  { value: "UNRESTRICTED", label: "Unrestricted" },
  { value: "SPONSORED", label: "Sponsored" },
];

test("a select is a native select with its label, help text, options and error", async () => {
  const container = render(
    <Select
      label="Accreditation level"
      description="The level of the accreditation."
      error="Choose a level."
      options={levels}
      defaultValue="SPONSORED"
    />,
  );
  const select = page.getByRole("combobox", { name: "Accreditation level" });

  expect(select.element().tagName).toBe("SELECT");
  await expect.element(select).toHaveValue("SPONSORED");
  await expect.element(select).toHaveAttribute("aria-invalid", "true");
  await expect
    .element(select)
    .toHaveAccessibleDescription("The level of the accreditation. Choose a level.");
  await expect.element(select.getByRole("option")).toHaveLength(2);
  await expectNoAxeViolations(container);
});

test("the keyboard changes the value of a select, with a visible focus", async () => {
  render(<Select label="Accreditation level" options={levels} defaultValue="UNRESTRICTED" />);
  const select = page.getByRole("combobox", { name: "Accreditation level" });

  await userEvent.tab();
  expectVisibleFocus(select.element());
  expectTargetSize(select.element());
  await userEvent.selectOptions(select, "Sponsored");

  await expect.element(select).toHaveValue("SPONSORED");
});

test("a select with a placeholder has no selected value until the user chooses one", async () => {
  const container = render(
    <form aria-label="Create">
      <Select
        label="Accreditation level"
        name="accreditationLevel"
        placeholder="Choose a level"
        options={levels}
      />
    </form>,
  );
  const select = page.getByRole("combobox", { name: "Accreditation level" });

  await expect.element(select).toHaveValue("");
  await expect.element(select).toHaveDisplayValue("Choose a level");
  await expect.element(select.getByRole("option")).toHaveLength(3);
  await expectNoAxeViolations(container);

  await userEvent.selectOptions(select, "Sponsored");

  await expect.element(select).toHaveValue("SPONSORED");
});
