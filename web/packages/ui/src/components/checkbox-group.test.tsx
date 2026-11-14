import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { CheckboxGroup } from "./checkbox-group";
import { Form } from "./form";

const industries = [
  { value: "banking", label: "Banking" },
  { value: "energy", label: "Energy" },
  { value: "telco", label: "Telco" },
];

const group = () => page.getByRole("group", { name: "Industries" });
const checkbox = (name: string) => page.getByRole("checkbox", { name });

test("a checkbox group is a fieldset with a legend, and a checkbox for each option", async () => {
  const container = render(
    <CheckboxGroup label="Industries" name="industries" options={industries} />,
  );

  await expect.element(group()).toBeVisible();
  expect(group().element().tagName).toBe("FIELDSET");
  await expect.element(page.getByText("Industries", { exact: true })).toBeVisible();
  await expect.element(group().getByRole("checkbox")).toHaveLength(3);
  await expect.element(checkbox("Banking")).not.toBeChecked();
  await expect.element(group()).toHaveAccessibleDescription("");
  await expectNoAxeViolations(container);
});

test("the group has one help text and one error message, and the screen reader reads them for the group", async () => {
  const container = render(
    <CheckboxGroup
      label="Industries"
      name="industries"
      options={industries}
      description="Select each industry of the brand."
      error="industries is required"
    />,
  );

  await expect.element(page.getByText("Select each industry of the brand.")).toBeVisible();
  await expect.element(page.getByText("industries is required")).toBeVisible();
  await expect
    .element(group())
    .toHaveAccessibleDescription("Select each industry of the brand. industries is required");
  await expect.element(checkbox("Banking")).toHaveAccessibleDescription("");
  await expectNoAxeViolations(container);
});

test("the default values are checked, and a click on a label changes its checkbox", async () => {
  render(
    <CheckboxGroup
      label="Industries"
      name="industries"
      options={industries}
      defaultValue={["energy", "telco"]}
    />,
  );

  await expect.element(checkbox("Banking")).not.toBeChecked();
  await expect.element(checkbox("Energy")).toBeChecked();

  await page.getByText("Banking", { exact: true }).click();
  await page.getByText("Telco", { exact: true }).click();

  await expect.element(checkbox("Banking")).toBeChecked();
  await expect.element(checkbox("Telco")).not.toBeChecked();
});

test("the keyboard reaches each checkbox with a visible focus, and Space changes it", async () => {
  render(<CheckboxGroup label="Industries" name="industries" options={industries} />);

  await userEvent.tab();
  expectVisibleFocus(checkbox("Banking").element());
  await userEvent.keyboard(" ");
  await expect.element(checkbox("Banking")).toBeChecked();
  await userEvent.tab();
  expectVisibleFocus(checkbox("Energy").element());
});

test("each checkbox with its label is a target of 24 by 24 CSS px or more", () => {
  render(<CheckboxGroup label="Industries" name="industries" options={industries} />);

  for (const { label } of industries) {
    const target = checkbox(label).element().closest("label");
    if (!target) throw new Error(`no label for ${label}`);
    expectTargetSize(target);
  }
});

test("a disabled group disables each checkbox", async () => {
  const container = render(
    <CheckboxGroup label="Industries" name="industries" options={industries} disabled />,
  );

  await expect.element(checkbox("Banking")).toBeDisabled();
  await expect.element(checkbox("Telco")).toBeDisabled();
  await expectNoAxeViolations(container);
});

test("in a form, the group gives each checked value, and shows the error of a value in the group", async () => {
  let got: FormData | undefined;
  const container = render(
    <Form
      label="Add brand"
      submitLabel="Add brand"
      cancel={<a href="#detail">Cancel</a>}
      onSubmit={(data) => {
        got = data;
        return Promise.resolve({
          problem: { title: "Values in the request are not valid" },
          fieldErrors: [
            { pointer: "#/industries", detail: 'industry "insurance" is not valid' },
            { pointer: "#/industries/1", detail: 'industry "telco" is not unique' },
          ],
        });
      }}
    >
      <CheckboxGroup
        label="Industries"
        name="industries"
        options={industries}
        defaultValue={["energy"]}
      />
    </Form>,
  );

  await page.getByText("Telco", { exact: true }).click();
  await page.getByRole("button", { name: "Add brand" }).click();

  await expect.poll(() => got?.getAll("industries")).toEqual(["energy", "telco"]);
  await expect
    .element(group())
    .toHaveAccessibleDescription(
      'industry "insurance" is not valid industry "telco" is not unique',
    );
  await expect.element(checkbox("Telco")).toBeChecked();
  await expectNoAxeViolations(container);

  await page
    .getByRole("alert")
    .getByRole("link", { name: 'industry "insurance" is not valid' })
    .click();

  await expect.poll(() => document.activeElement).toBe(checkbox("Banking").element());
});
