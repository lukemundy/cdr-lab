import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientDetail } from "../api/client";
import { http, recipient, renderApp, worker } from "../test/app";

const id = "0b6f4b55-2d0e-4c55-9a43-5b8e1f3b7a10";
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

const created: DataRecipientDetail = {
  legalEntityId: id,
  legalEntityName: "Gamma Pty Ltd",
  accreditationNumber: "ADR-G",
  accreditationLevel: "SPONSORED",
  logoUri: "https://gamma.example/logo.png",
  status: "SUSPENDED",
  lastUpdated: "2026-10-24T02:05:07Z",
  version: 1,
  actions: [],
  dataRecipientBrands: [],
};

type Result = { status: 201 } | { status: 400 | 409; body: unknown };

function serveCreate(...results: Result[]) {
  const bodies: unknown[] = [];
  worker.use(
    http.post("/api/data-recipients", async ({ request }) => {
      bodies.push(await request.json());
      const result = results[Math.min(bodies.length - 1, results.length - 1)] ?? { status: 201 };
      if (result.status === 201)
        return new Response(null, {
          status: 201,
          headers: { Location: `/api/data-recipients/${id}` },
        });
      return new Response(JSON.stringify(result.body), {
        status: result.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }),
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) => response(200).json(created)),
    http.get("/api/data-recipients", ({ response }) =>
      response(200).json({ data: [recipient("Alpha")], totalRecords: 1 }),
    ),
  );
  return bodies;
}

const refusal = (pointer: string, detail: string, status: 400 | 409) => ({
  type: "https://github.com/lukemundy/cdr-lab/blob/main/services/register/DESIGN.md#console-api",
  title: "Values in the request are not valid",
  status,
  detail: "One or more values in the request are not valid.",
  errors: [{ pointer, detail }],
});

const idNotUnique = "legalEntityId is not unique: a different record has it";
const noLevel = 'accreditationLevel "" is not UNRESTRICTED or SPONSORED';

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const level = () => page.getByRole("combobox", { name: "Accreditation level" });
const status = () => page.getByRole("combobox", { name: "Status" });
const create = () => page.getByRole("button", { name: /^Create data recipient/ });

async function fillRequiredFields() {
  await userEvent.fill(field("Name"), " Gamma Pty Ltd ");
  await userEvent.fill(field("Accreditation number"), "ADR-G");
  await userEvent.fill(field("Logo URI"), "https://gamma.example/logo.png ");
}

describe("the create of a data recipient", () => {
  test("opens from the list, with a new UUID, no accreditation level and the status ACTIVE", async () => {
    serveCreate();
    const router = renderApp("/data-recipients");

    await page.getByRole("link", { name: "Create data recipient" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Create a data recipient");
    expect(router.state.location.pathname).toBe("/create/data-recipient");
    await expect.element(field("Legal entity ID")).toHaveValue(expect.stringMatching(uuid));
    await expect.element(level()).toHaveValue("");
    await expect.element(status()).toHaveValue("ACTIVE");
    await expect.element(status().getByRole("option")).toHaveLength(4);
    await expect.element(field("Name")).toHaveValue("");
    await expectNoAxeViolations(document.body);
  });

  test("gives a different UUID each time that it opens", async () => {
    serveCreate();
    renderApp("/create/data-recipient");
    await expect.element(field("Legal entity ID")).toHaveValue(expect.stringMatching(uuid));
    const first = field("Legal entity ID").element();
    if (!(first instanceof HTMLInputElement)) throw new Error("no input");
    const firstValue = first.value;

    await page.getByRole("link", { name: "Cancel" }).click();
    await page.getByRole("link", { name: "Create data recipient" }).click();

    await expect.element(field("Legal entity ID")).toHaveValue(expect.stringMatching(uuid));
    await expect.element(field("Legal entity ID")).not.toHaveValue(firstValue);
  });

  test("sends each field with no white space and no version, then shows the detail with a toast", async () => {
    const bodies = serveCreate({ status: 201 });
    const router = renderApp("/create/data-recipient");

    await userEvent.fill(field("Legal entity ID"), ` ${id} `);
    await fillRequiredFields();
    await userEvent.selectOptions(level(), "SPONSORED");
    await userEvent.selectOptions(status(), "SUSPENDED");
    await create().click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Gamma Pty Ltd" }))
      .toBeVisible();
    expect(bodies).toEqual([
      {
        legalEntityId: id,
        legalEntityName: "Gamma Pty Ltd",
        accreditationNumber: "ADR-G",
        accreditationLevel: "SPONSORED",
        logoUri: "https://gamma.example/logo.png",
        status: "SUSPENDED",
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-recipients/${id}`);
    await expect
      .element(page.getByRole("dialog", { name: "Data recipient Gamma Pty Ltd created" }))
      .toBeVisible();
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  test("sends no accreditation level when the operator chose none, so that the Register refuses it at its field", async () => {
    const bodies = serveCreate({
      status: 400,
      body: refusal("#/accreditationLevel", noLevel, 400),
    });
    renderApp("/create/data-recipient");

    await fillRequiredFields();
    await create().click();

    await expect.element(level()).toHaveAccessibleDescription(noLevel);
    expect(bodies[0]).not.toHaveProperty("accreditationLevel");
  });

  test("shows a legal entity ID in use at its field, and keeps the values", async () => {
    const bodies = serveCreate({ status: 409, body: refusal("#/legalEntityId", idNotUnique, 409) });
    renderApp("/create/data-recipient");

    await userEvent.fill(field("Legal entity ID"), id);
    await fillRequiredFields();
    await create().click();

    const summary = page.getByRole("alert");
    await expect.element(summary.getByRole("link", { name: idNotUnique })).toBeVisible();
    await expect
      .element(field("Legal entity ID"))
      .toHaveAccessibleDescription(
        `A new random UUID. Change it to give a known ID. ${idNotUnique}`,
      );
    await expect.element(field("Legal entity ID")).toHaveValue(id);
    await expect.element(field("Name")).toHaveValue(" Gamma Pty Ltd ");
    expect(bodies).toHaveLength(1);
    await expectNoAxeViolations(document.body);
  });

  test("Cancel goes back to the list and sends nothing", async () => {
    const bodies = serveCreate();
    const router = renderApp("/create/data-recipient");

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Data recipients" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe("/data-recipients");
    expect(bodies).toEqual([]);
  });
});
