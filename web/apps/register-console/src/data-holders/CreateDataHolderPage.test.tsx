import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataHolderDetail } from "../api/client";
import { holder, http, renderApp, worker } from "../test/app";

const id = "5d1c9a2e-7b3f-4e8a-9c21-0f6b4d2e8a31";
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

const created: DataHolderDetail = {
  legalEntityId: id,
  legalEntityName: "Delta Ltd",
  logoUri: "https://delta.example/logo.png",
  status: "REMOVED",
  lastUpdated: "2026-10-24T02:05:07Z",
  version: 1,
  actions: [],
  dataHolderBrands: [],
};

type Result = { status: 201 } | { status: 400 | 409; body: unknown };

function serveCreate(...results: Result[]) {
  const bodies: unknown[] = [];
  worker.use(
    http.post("/api/data-holders", async ({ request }) => {
      bodies.push(await request.json());
      const result = results[Math.min(bodies.length - 1, results.length - 1)] ?? { status: 201 };
      if (result.status === 201)
        return new Response(null, {
          status: 201,
          headers: { Location: `/api/data-holders/${id}` },
        });
      return new Response(JSON.stringify(result.body), {
        status: result.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }),
    http.get("/api/data-holders/{legalEntityId}", ({ response }) => response(200).json(created)),
    http.get("/api/data-holders", ({ response }) =>
      response(200).json({ data: [holder("Alpha Ltd")], totalRecords: 1 }),
    ),
  );
  return bodies;
}

const idNotUnique = "legalEntityId is not unique: a different record has it";

const refusal = {
  type: "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
  title: "Values in the request are not valid",
  status: 409,
  detail: "One or more values in the request are not valid.",
  errors: [{ pointer: "#/legalEntityId", detail: idNotUnique }],
};

const field = (name: string) => page.getByRole("textbox", { name, exact: true });
const status = () => page.getByRole("combobox", { name: "Status" });
const create = () => page.getByRole("button", { name: /^Create data holder/ });

async function fillRequiredFields() {
  await userEvent.fill(field("Name"), " Delta Ltd ");
  await userEvent.fill(field("Logo URI"), "https://delta.example/logo.png ");
}

describe("the create of a data holder", () => {
  // Verifies: REG-1407, REG-1417, REG-1419
  test("opens from the list, with a new UUID and the status ACTIVE", async () => {
    serveCreate();
    const router = renderApp("/data-holders");

    await page.getByRole("link", { name: "Create data holder" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1 }))
      .toHaveTextContent("Create a data holder");
    expect(router.state.location.pathname).toBe("/create/data-holder");
    await expect.element(field("Legal entity ID")).toHaveValue(expect.stringMatching(uuid));
    await expect.element(status()).toHaveValue("ACTIVE");
    await expect.element(status().getByRole("option")).toHaveLength(2);
    await expect.element(field("Name")).toHaveValue("");
    await expect
      .element(
        page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Data holders" }),
      )
      .toHaveAttribute("aria-current", "page");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1430, REG-1434, REG-1445
  test("sends each field with no white space and no version, then shows the detail with a toast", async () => {
    const bodies = serveCreate({ status: 201 });
    const router = renderApp("/create/data-holder");

    await userEvent.fill(field("Legal entity ID"), ` ${id} `);
    await fillRequiredFields();
    await userEvent.selectOptions(status(), "REMOVED");
    await create().click();

    await expect.element(page.getByRole("heading", { level: 1, name: "Delta Ltd" })).toBeVisible();
    expect(bodies).toEqual([
      {
        legalEntityId: id,
        legalEntityName: "Delta Ltd",
        logoUri: "https://delta.example/logo.png",
        status: "REMOVED",
      },
    ]);
    expect(router.state.location.pathname).toBe(`/data-holders/${id}`);
    await expect
      .element(page.getByRole("dialog", { name: "Data holder Delta Ltd created" }))
      .toBeVisible();
    await expect
      .poll(() => document.activeElement)
      .toBe(page.getByRole("heading", { level: 1 }).element());
  });

  // Verifies: REG-1436, REG-1441
  test("shows a legal entity ID in use at its field, and keeps the values", async () => {
    const bodies = serveCreate({ status: 409, body: refusal });
    renderApp("/create/data-holder");

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
    await expect.element(field("Name")).toHaveValue(" Delta Ltd ");
    expect(bodies).toHaveLength(1);
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1448, REG-1449
  test("Cancel goes back to the list and sends nothing", async () => {
    const bodies = serveCreate();
    const router = renderApp("/create/data-holder");

    await page.getByRole("link", { name: "Cancel" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Data holders" }))
      .toBeVisible();
    expect(router.state.location.pathname).toBe("/data-holders");
    expect(bodies).toEqual([]);
  });
});
