import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataHolderBrand, DataHolderBrandStatusAction, DataHolderDetail } from "../api/client";
import { http, worker, renderDetail } from "../test/app";

const id = "3e0f6a3c-1d55-4b57-9a54-7c2a4f3b9e21";

const deactivate: DataHolderBrandStatusAction = {
  action: "deactivate",
  status: "INACTIVE",
  irreversible: false,
  changes: [],
};
const remove: DataHolderBrandStatusAction = {
  action: "remove",
  status: "REMOVED",
  irreversible: true,
  changes: [],
};

const bankId = "brand-console-bank";
const productsId = "interim-console-products";

const bank: DataHolderBrand = {
  dataHolderBrandId: bankId,
  brandName: "Console Bank",
  industries: ["banking"],
  logoUri: "https://console-bank.example/logo.png",
  status: "ACTIVE",
  actions: [deactivate, remove],
  serverCertificateActions: [],
  endpointDetail: { publicBaseUri: "https://api.console-bank.example" },
};

const products: DataHolderBrand = {
  interimId: productsId,
  brandName: "Console Products",
  industries: ["banking"],
  logoUri: "https://console-products.example/logo.png",
  status: "INACTIVE",
  actions: [{ action: "activate", status: "ACTIVE", irreversible: false, changes: [] }, remove],
  serverCertificateActions: [],
  endpointDetail: { publicBaseUri: "https://api.console-products.example" },
};

const active: DataHolderDetail = {
  legalEntityId: id,
  legalEntityName: "Console Holdings Ltd",
  logoUri: "https://console-holdings.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-24T02:05:07Z",
  version: 6,
  actions: [
    {
      action: "remove",
      status: "REMOVED",
      irreversible: true,
      changes: [
        {
          dataHolderBrandId: bankId,
          brandName: bank.brandName,
          from: "ACTIVE",
          to: "REMOVED",
        },
        {
          interimId: productsId,
          brandName: products.brandName,
          from: "INACTIVE",
          to: "REMOVED",
        },
      ],
    },
  ],
  dataHolderBrands: [bank, products],
};

const removed: DataHolderDetail = {
  ...active,
  status: "REMOVED",
  version: 9,
  actions: [],
  dataHolderBrands: [
    { ...bank, status: "REMOVED", actions: [] },
    { ...products, status: "REMOVED", actions: [] },
  ],
};

// The detail depends on the action requests, not on the count of reads, because a read of a previous
// test can arrive late.
function serveDetails(before: DataHolderDetail, after: DataHolderDetail = before) {
  const requests: unknown[] = [];
  worker.use(
    http.get("/api/data-holders/{legalEntityId}", ({ response }) =>
      response(200).json(requests.length === 0 ? before : after),
    ),
  );
  return requests;
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const actionsOf = (name: string) =>
  page.getByRole("group", { name: `Status actions of ${name}`, exact: true });
const buttonNames = (name: string) =>
  actionsOf(name)
    .getByRole("button")
    .all()
    .map((button) => button.element().textContent);

describe("the status actions of a data holder", () => {
  test("show only the actions that the console API gives, for the legal entity and each brand", async () => {
    serveDetails(active);

    await renderDetail(`/data-holders/${id}`);

    await expect.element(actionsOf("Console Holdings Ltd")).toBeVisible();
    expect(buttonNames("Console Holdings Ltd")).toEqual(["Remove"]);
    expect(buttonNames("Console Bank")).toEqual(["Deactivate", "Remove"]);
    expect(buttonNames("Console Products")).toEqual(["Activate", "Remove"]);
    await expectNoAxeViolations(document.body);
  });

  test("a removal names each brand that it removes, says that it cannot be reversed, sends the version, and shows the new state", async () => {
    const requests = serveDetails(active, removed);
    worker.use(
      http.post(
        "/api/data-holders/{legalEntityId}/remove",
        async ({ params, request, response }) => {
          requests.push({ ...params, body: await request.json() });
          return response(204).empty();
        },
      ),
    );
    await renderDetail(`/data-holders/${id}`);

    await userEvent.click(
      actionsOf("Console Holdings Ltd").getByRole("button", { name: "Remove" }),
    );

    const dialog = page.getByRole("alertdialog", { name: "Remove Console Holdings Ltd?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        "Console Holdings Ltd changes from ACTIVE to REMOVED. " +
          "Console Bank changes from ACTIVE to REMOVED. " +
          "Console Products changes from INACTIVE to REMOVED. " +
          "You cannot reverse this action.",
      );
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Remove", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([{ legalEntityId: id, body: { version: 6 } }]);
    await expect
      .element(page.getByRole("dialog", { name: "Console Holdings Ltd is now REMOVED" }))
      .toBeVisible();
    await expect.element(region("Legal entity").getByText("REMOVED")).toBeVisible();
    await expect.element(region("Console Bank").getByText("REMOVED")).toBeVisible();
    expect(buttonNames("Console Holdings Ltd")).toEqual([]);
    expect(buttonNames("Console Bank")).toEqual([]);
  });

  test("a brand action sends the version to the path of the brand, with the interim ID when the brand has no brand ID", async () => {
    const requests = serveDetails(active);
    for (const action of ["deactivate", "remove"] as const) {
      worker.use(
        http.post(
          `/api/data-holders/{legalEntityId}/brands/{brandId}/${action}`,
          async ({ params, request, response }) => {
            requests.push({ ...params, action, body: await request.json() });
            return response(204).empty();
          },
        ),
      );
    }
    await renderDetail(`/data-holders/${id}`);

    await userEvent.click(actionsOf("Console Bank").getByRole("button", { name: "Deactivate" }));
    const deactivation = page.getByRole("alertdialog", { name: "Deactivate Console Bank?" });
    await expect
      .element(deactivation)
      .toHaveAccessibleDescription("Console Bank changes from ACTIVE to INACTIVE.");
    await deactivation.getByRole("button", { name: "Deactivate", exact: true }).click();
    await expect.element(deactivation).not.toBeInTheDocument();

    await userEvent.click(actionsOf("Console Products").getByRole("button", { name: "Remove" }));
    const removal = page.getByRole("alertdialog", { name: "Remove Console Products?" });
    await expect
      .element(removal)
      .toHaveAccessibleDescription(
        "Console Products changes from INACTIVE to REMOVED. You cannot reverse this action.",
      );
    await removal.getByRole("button", { name: "Remove", exact: true }).click();
    await expect.element(removal).not.toBeInTheDocument();

    expect(requests).toEqual([
      {
        legalEntityId: id,
        brandId: bankId,
        action: "deactivate",
        body: { version: 6 },
      },
      {
        legalEntityId: id,
        brandId: productsId,
        action: "remove",
        body: { version: 6 },
      },
    ]);
  });
});
