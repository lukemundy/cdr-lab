import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { Fieldset } from "./fieldset";
import { Form } from "./form";
import { TextField } from "./form-field";

test("a fieldset is a group of fields with a legend that names the group", async () => {
  const container = render(
    <Fieldset legend="Endpoints">
      <TextField label="Public base URI" />
      <TextField label="Website URI" />
    </Fieldset>,
  );
  const group = page.getByRole("group", { name: "Endpoints" });

  await expect.element(group).toBeVisible();
  expect(group.element().tagName).toBe("FIELDSET");
  await expect.element(page.getByText("Endpoints", { exact: true })).toBeVisible();
  await expect.element(group.getByRole("textbox")).toHaveLength(2);
  await expectNoAxeViolations(container);
});

test("in a form, each field of the fieldset shows the error of its nested value", async () => {
  const container = render(
    <Form
      label="Add brand"
      submitLabel="Add brand"
      cancel={<a href="#detail">Cancel</a>}
      onSubmit={() =>
        Promise.resolve({
          problem: { title: "Values in the request are not valid" },
          fieldErrors: [
            {
              pointer: "#/endpointDetail/publicBaseUri",
              detail: "endpointDetail.publicBaseUri is required",
            },
          ],
        })
      }
    >
      <Fieldset legend="Endpoints">
        <TextField label="Public base URI" name="endpointDetail/publicBaseUri" />
        <TextField label="Website URI" name="endpointDetail/websiteUri" />
      </Fieldset>
    </Form>,
  );

  await page.getByRole("button", { name: "Add brand" }).click();

  const group = page.getByRole("group", { name: "Endpoints" });
  await expect
    .element(group.getByRole("textbox", { name: "Public base URI" }))
    .toHaveAccessibleDescription("endpointDetail.publicBaseUri is required");
  await expect
    .element(group.getByRole("textbox", { name: "Website URI" }))
    .not.toHaveAttribute("aria-invalid");
  await expectNoAxeViolations(container);
});
