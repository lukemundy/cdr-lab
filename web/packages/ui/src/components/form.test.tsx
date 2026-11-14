import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { Form, type FormRefusal } from "./form";
import { Select, TextField } from "./form-field";

const levels = [
  { value: "UNRESTRICTED", label: "Unrestricted" },
  { value: "SPONSORED", label: "Sponsored" },
];

type Submit = (data: FormData) => Promise<FormRefusal | undefined>;

function renderForm(onSubmit: Submit) {
  return render(
    <Form
      label="Change details"
      submitLabel="Save changes"
      cancel={<a href="#detail">Cancel</a>}
      onSubmit={onSubmit}
    >
      <TextField label="Legal entity name" name="legalEntityName" defaultValue="Mock Pty Ltd" />
      <TextField label="Logo URI (optional)" name="logoUri" />
      <Select
        label="Accreditation level"
        name="accreditationLevel"
        options={levels}
        defaultValue="SPONSORED"
      />
    </Form>,
  );
}

const refusal: FormRefusal = {
  problem: {
    title: "Values in the request are not valid",
    detail: "One or more values are not valid.",
  },
  fieldErrors: [
    { pointer: "#/legalEntityName", detail: "legalEntityName is longer than 200 characters" },
    {
      pointer: "#/accreditationLevel",
      detail: 'accreditationLevel "X" is not UNRESTRICTED or SPONSORED',
    },
  ],
};

test("a form has its fields, a submit button that names the action and a Cancel link", async () => {
  const container = renderForm(() => Promise.resolve(undefined));

  await expect.element(page.getByRole("form", { name: "Change details" })).toBeVisible();
  await expect.element(page.getByRole("button", { name: "Save changes" })).toBeVisible();
  await expect
    .element(page.getByRole("link", { name: "Cancel" }))
    .toHaveAttribute("href", "#detail");
  await expect
    .element(page.getByRole("combobox", { name: "Accreditation level" }))
    .toHaveValue("SPONSORED");
  await expect.element(page.getByRole("alert")).not.toBeInTheDocument();
  await expectNoAxeViolations(container);
});

test("the form gives the values to the submit, and the server checks them", async () => {
  let got: FormData | undefined;
  renderForm((data) => {
    got = data;
    return Promise.resolve(undefined);
  });

  await userEvent.fill(page.getByRole("textbox", { name: "Logo URI (optional)" }), "not a URI");
  await page.getByRole("button", { name: "Save changes" }).click();

  await expect.poll(() => got?.get("logoUri")).toBe("not a URI");
  expect(got?.get("legalEntityName")).toBe("Mock Pty Ltd");
  expect(got?.get("accreditationLevel")).toBe("SPONSORED");
});

test("while the request runs, the submit button is busy, a status says so, and a second submit sends nothing", async () => {
  let calls = 0;
  let finish: ((refusal: FormRefusal | undefined) => void) | undefined;
  renderForm(() => {
    calls++;
    return new Promise((resolve) => {
      finish = resolve;
    });
  });
  const submit = page.getByRole("button", { name: /^Save changes/ });

  await submit.click();

  await expect.element(submit).toHaveAttribute("aria-busy", "true");
  await expect.element(page.getByRole("status")).toHaveTextContent("Save changes in progress");
  await userEvent.click(page.getByRole("textbox", { name: "Legal entity name" }));
  await userEvent.keyboard("{Enter}");
  expect(calls).toBe(1);

  finish?.(undefined);

  await expect.element(submit).not.toHaveAttribute("aria-busy");
  await expect.element(page.getByRole("status")).toHaveTextContent("");
});

test("after a refusal, the error summary has the focus, and each field shows its error and keeps its value", async () => {
  const container = renderForm(() => Promise.resolve(refusal));
  const name = page.getByRole("textbox", { name: "Legal entity name" });

  await userEvent.fill(name, "A changed name");
  await page.getByRole("button", { name: "Save changes" }).click();

  const summary = page.getByRole("alert");
  await expect.element(summary.getByRole("heading", { name: "There is a problem" })).toBeVisible();
  await expect.element(summary.getByText("One or more values are not valid.")).toBeVisible();
  await expect.element(summary.getByRole("link")).toHaveLength(2);
  await expect
    .element(summary.getByRole("link").first())
    .toHaveTextContent("legalEntityName is longer than 200 characters");
  expect(summary.element().parentElement).toBe(document.activeElement);

  await expect.element(name).toHaveValue("A changed name");
  await expect.element(name).toHaveAttribute("aria-invalid", "true");
  await expect
    .element(name)
    .toHaveAccessibleDescription("legalEntityName is longer than 200 characters");
  await expect
    .element(page.getByRole("combobox", { name: "Accreditation level" }))
    .toHaveAccessibleDescription('accreditationLevel "X" is not UNRESTRICTED or SPONSORED');
  await expect
    .element(page.getByRole("textbox", { name: "Logo URI (optional)" }))
    .not.toHaveAttribute("aria-invalid");
  await expectNoAxeViolations(container);
});

test("the error summary is at the top of the form", async () => {
  renderForm(() => Promise.resolve(refusal));

  await page.getByRole("button", { name: "Save changes" }).click();

  const form = page.getByRole("form").element();
  await expect.element(page.getByRole("alert")).toBeVisible();
  expect(form.firstElementChild?.contains(page.getByRole("alert").element())).toBe(true);
});

test("a link of the error summary moves the focus to its field", async () => {
  renderForm(() => Promise.resolve(refusal));
  await page.getByRole("button", { name: "Save changes" }).click();
  const links = page.getByRole("alert").getByRole("link");

  await expect.element(links.first()).toBeVisible();
  await userEvent.tab();
  expectVisibleFocus(links.first().element());
  await userEvent.keyboard("{Enter}");

  expect(document.activeElement).toBe(
    page.getByRole("textbox", { name: "Legal entity name" }).element(),
  );

  await links.nth(1).click();

  expect(document.activeElement).toBe(
    page.getByRole("combobox", { name: "Accreditation level" }).element(),
  );
});

test("a refusal that names no field shows only in the error summary, and the focus moves to it again", async () => {
  const conflict: FormRefusal = {
    problem: {
      title: "Conflict",
      detail: "The data recipient changed after the version in the request.",
    },
  };
  renderForm(() => Promise.resolve(conflict));
  const submit = page.getByRole("button", { name: /^Save changes/ });

  await submit.click();

  const summary = page.getByRole("alert");
  await expect
    .element(summary.getByText("The data recipient changed after the version in the request."))
    .toBeVisible();
  await expect.element(summary.getByRole("link")).not.toBeInTheDocument();
  await expect
    .element(page.getByRole("textbox", { name: "Legal entity name" }))
    .not.toHaveAttribute("aria-invalid");

  await submit.click();

  await expect.poll(() => summary.element().parentElement === document.activeElement).toBe(true);
});

test("a submit that fails with no problem details says that the change did not complete", async () => {
  renderForm(() => Promise.reject(new Error("no network")));

  await page.getByRole("button", { name: "Save changes" }).click();

  await expect
    .element(
      page
        .getByRole("alert")
        .getByText(
          "The change did not complete. Refresh the page to see the current state, then try again.",
        ),
    )
    .toBeVisible();
});

test("a field error of a list item or of a nested value shows at the field of the list or of the value", async () => {
  render(
    <Form
      label="Change brand"
      submitLabel="Save changes"
      cancel={<a href="#detail">Cancel</a>}
      onSubmit={() =>
        Promise.resolve({
          problem: { title: "Values in the request are not valid" },
          fieldErrors: [
            { pointer: "#/redirectUris/1", detail: "redirectUris has a value that is not valid" },
            {
              pointer: "#/endpointDetail/publicBaseUri",
              detail: "endpointDetail.publicBaseUri is required",
            },
          ],
        })
      }
    >
      <TextField label="Redirect URIs" name="redirectUris" />
      <TextField label="Public base URI" name="endpointDetail/publicBaseUri" />
    </Form>,
  );

  await page.getByRole("button", { name: "Save changes" }).click();

  await expect
    .element(page.getByRole("textbox", { name: "Redirect URIs" }))
    .toHaveAccessibleDescription("redirectUris has a value that is not valid");
  await expect
    .element(page.getByRole("textbox", { name: "Public base URI" }))
    .toHaveAccessibleDescription("endpointDetail.publicBaseUri is required");
});
