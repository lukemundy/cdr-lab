import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type {
  DataRecipientBrand,
  DataRecipientDetail,
  DataRecipientStatusAction,
  SoftwareProduct,
} from "./api/client";
import { http, renderDetail, worker } from "./test/app";

const id = "7c89d9d4-4f19-4bb3-90c3-b2d5b95520c4";

function product(name: string, fields: Partial<SoftwareProduct>): SoftwareProduct {
  const slug = name.toLowerCase().replaceAll(" ", "-");
  return {
    softwareProductId: `sp-${slug}`,
    softwareProductName: name,
    softwareProductDescription: "A mock software product",
    logoUri: `https://${slug}.example/logo.png`,
    status: "ACTIVE",
    clientUri: `https://${slug}.example`,
    redirectUris: [`https://${slug}.example/callback`],
    jwksUri: `https://${slug}.example/jwks`,
    revocationUri: `https://${slug}.example/revoke`,
    recipientBaseUri: `https://cdr.${slug}.example`,
    scopes: ["openid"],
    clientCertificateThumbprints: [],
    actions: [],
    clientCertificateActions: [],
    ...fields,
  };
}

const web = product("Console Test Web", {
  actions: [
    { action: "deactivate", status: "INACTIVE", irreversible: false, changes: [] },
    { action: "remove", status: "REMOVED", irreversible: true, changes: [] },
  ],
});
const mobile = product("Console Test Mobile", { status: "INACTIVE" });

const brand: DataRecipientBrand = {
  dataRecipientBrandId: "brand-console-test",
  brandName: "Console Test",
  logoUri: "https://console-test.example/logo.png",
  status: "ACTIVE",
  actions: [
    {
      action: "deactivate",
      status: "INACTIVE",
      irreversible: false,
      changes: [
        {
          softwareProductId: web.softwareProductId,
          softwareProductName: web.softwareProductName,
          from: "ACTIVE",
          to: "INACTIVE",
        },
      ],
    },
    {
      action: "remove",
      status: "REMOVED",
      irreversible: true,
      changes: [
        {
          softwareProductId: web.softwareProductId,
          softwareProductName: web.softwareProductName,
          from: "ACTIVE",
          to: "REMOVED",
        },
        {
          softwareProductId: mobile.softwareProductId,
          softwareProductName: mobile.softwareProductName,
          from: "INACTIVE",
          to: "REMOVED",
        },
      ],
    },
  ],
  softwareProducts: [web, mobile],
};

const active: DataRecipientDetail = {
  legalEntityId: id,
  legalEntityName: "Console Test Pty Ltd",
  accreditationNumber: "ADR-MOCK-0011",
  accreditationLevel: "UNRESTRICTED",
  logoUri: "https://console-test.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-23T02:05:07Z",
  version: 7,
  actions: [
    {
      action: "suspend",
      status: "SUSPENDED",
      irreversible: false,
      changes: [
        {
          softwareProductId: web.softwareProductId,
          softwareProductName: web.softwareProductName,
          from: "ACTIVE",
          to: "INACTIVE",
        },
      ],
    },
    {
      action: "revoke",
      status: "REVOKED",
      irreversible: true,
      changes: [
        {
          softwareProductId: web.softwareProductId,
          softwareProductName: web.softwareProductName,
          from: "ACTIVE",
          to: "REMOVED",
        },
        {
          softwareProductId: mobile.softwareProductId,
          softwareProductName: mobile.softwareProductName,
          from: "INACTIVE",
          to: "REMOVED",
        },
      ],
    },
  ],
  dataRecipientBrands: [brand],
};

const suspended: DataRecipientDetail = {
  ...active,
  status: "SUSPENDED",
  version: 9,
  actions: [{ action: "reinstate", status: "ACTIVE", irreversible: false, changes: [] }],
  dataRecipientBrands: [
    {
      ...brand,
      softwareProducts: [{ ...web, status: "INACTIVE", actions: [] }, mobile],
    },
  ],
};

const brandDeactivated: DataRecipientDetail = {
  ...active,
  version: 9,
  dataRecipientBrands: [
    {
      ...brand,
      status: "INACTIVE",
      actions: [
        { action: "activate", status: "ACTIVE", irreversible: false, changes: [] },
        {
          action: "remove",
          status: "REMOVED",
          irreversible: true,
          changes: [web, mobile].map((p) => ({
            softwareProductId: p.softwareProductId,
            softwareProductName: p.softwareProductName,
            from: "INACTIVE" as const,
            to: "REMOVED" as const,
          })),
        },
      ],
      softwareProducts: [{ ...web, status: "INACTIVE", actions: [] }, mobile],
    },
  ],
};

const conflict = {
  type: "about:blank",
  title: "Conflict",
  status: 409,
  detail:
    "The data recipient changed after the version in the request. The Register changed nothing.",
} as const;

// The detail depends on the action requests, not on the count of reads, because a read of a previous
// test can arrive late.
function serveDetails(before: DataRecipientDetail, after: DataRecipientDetail = before) {
  const requests: unknown[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) =>
      response(200).json(requests.length === 0 ? before : after),
    ),
  );
  return requests;
}

function serveSuspension(requests: unknown[], result: "success" | "conflict") {
  worker.use(
    http.post("/api/data-recipients/{legalEntityId}/suspend", async ({ request, response }) => {
      requests.push(await request.json());
      return result === "success" ? response(204).empty() : response(409).json(conflict);
    }),
  );
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const actionsOf = (name: string) =>
  page.getByRole("group", { name: `Status actions of ${name}`, exact: true });
const buttonNames = (name: string) =>
  actionsOf(name)
    .getByRole("button")
    .all()
    .map((button) => button.element().textContent);

describe("the status actions of a data recipient", () => {
  // Verifies: REG-1219
  test("show only the actions that the console API gives, for the legal entity, each brand and each software product", async () => {
    serveDetails(active);

    await renderDetail(`/data-recipients/${id}`);

    await expect.element(actionsOf("Console Test Pty Ltd")).toBeVisible();
    expect(buttonNames("Console Test Pty Ltd")).toEqual(["Suspend", "Revoke"]);
    expect(buttonNames("Console Test")).toEqual(["Deactivate", "Remove"]);
    expect(buttonNames("Console Test Web")).toEqual(["Deactivate", "Remove"]);
    expect(buttonNames("Console Test Mobile")).toEqual([]);
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1201, REG-1212, WCON-0015
  test("a suspension names each software product that it changes, sends the version, and shows the new state", async () => {
    const bodies = serveDetails(active, suspended);
    serveSuspension(bodies, "success");
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(
      actionsOf("Console Test Pty Ltd").getByRole("button", { name: "Suspend" }),
    );

    const dialog = page.getByRole("alertdialog", { name: "Suspend Console Test Pty Ltd?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        "Console Test Pty Ltd changes from ACTIVE to SUSPENDED. " +
          "Console Test Web changes from ACTIVE to INACTIVE.",
      );
    expect(dialog.element().textContent).not.toContain("Console Test Mobile");
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Suspend", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(bodies).toEqual([{ version: 7 }]);
    await expect
      .element(page.getByRole("dialog", { name: "Console Test Pty Ltd is now SUSPENDED" }))
      .toBeVisible();
    await expect.element(region("Legal entity").getByText("SUSPENDED")).toBeVisible();
    expect(buttonNames("Console Test Pty Ltd")).toEqual(["Reinstate"]);
    // The Suspend button does not exist now, so the focus goes into its group.
    await expect
      .poll(() => document.activeElement)
      .toBe(actionsOf("Console Test Pty Ltd").getByRole("button", { name: "Reinstate" }).element());
  });

  test("after an action that leaves no action, the focus goes to the group of the actions", async () => {
    const surrender: DataRecipientStatusAction = {
      action: "surrender",
      status: "SURRENDERED",
      irreversible: true,
      changes: [],
    };
    const surrendered: DataRecipientDetail = { ...suspended, status: "SURRENDERED", actions: [] };
    const requests = serveDetails({ ...suspended, actions: [surrender] }, surrendered);
    worker.use(
      http.post("/api/data-recipients/{legalEntityId}/surrender", async ({ request, response }) => {
        requests.push(await request.json());
        return response(204).empty();
      }),
    );
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(
      actionsOf("Console Test Pty Ltd").getByRole("button", { name: "Surrender" }),
    );
    await page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Surrender", exact: true })
      .click();

    await expect.element(page.getByRole("alertdialog")).not.toBeInTheDocument();
    await expect
      .poll(() => document.activeElement)
      .toBe(actionsOf("Console Test Pty Ltd").element());
  });

  // Verifies: REG-1213
  test("a revocation says that its software products become REMOVED and that it cannot be reversed", async () => {
    serveDetails(active);
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(
      actionsOf("Console Test Pty Ltd").getByRole("button", { name: "Revoke" }),
    );

    const dialog = page.getByRole("alertdialog", { name: "Revoke Console Test Pty Ltd?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        "Console Test Pty Ltd changes from ACTIVE to REVOKED. " +
          "Console Test Web changes from ACTIVE to REMOVED. " +
          "Console Test Mobile changes from INACTIVE to REMOVED. " +
          "You cannot reverse this action.",
      );
  });

  // Verifies: REG-1214
  test("a reinstatement says that the software products do not change", async () => {
    serveDetails(suspended);
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(
      actionsOf("Console Test Pty Ltd").getByRole("button", { name: "Reinstate" }),
    );

    await expect
      .element(page.getByRole("alertdialog", { name: "Reinstate Console Test Pty Ltd?" }))
      .toHaveAccessibleDescription(
        "Console Test Pty Ltd changes from SUSPENDED to ACTIVE. " +
          "The software products do not change. Activate each software product that must be ACTIVE.",
      );
  });

  // Verifies: WCON-0014
  test("a refused action shows the detail in the confirmation, also when the current state no longer permits the action", async () => {
    serveSuspension(serveDetails(active, suspended), "conflict");
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(
      actionsOf("Console Test Pty Ltd").getByRole("button", { name: "Suspend" }),
    );
    const dialog = page.getByRole("alertdialog", { name: "Suspend Console Test Pty Ltd?" });
    await dialog.getByRole("button", { name: "Suspend", exact: true }).click();

    await expect.element(dialog.getByRole("alert")).toMatchTextContent(conflict.detail);
    await expect
      .element(dialog.getByRole("listitem").first())
      .toHaveTextContent("Console Test Pty Ltd changes from ACTIVE to SUSPENDED.");
    expect(page.getByRole("dialog", { name: /is now/ }).query()).toBeNull();

    await dialog.getByRole("button", { name: "Cancel" }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    await expect.element(region("Legal entity").getByText("SUSPENDED")).toBeVisible();
    expect(buttonNames("Console Test Pty Ltd")).toEqual(["Reinstate"]);
    await expect
      .poll(() => document.activeElement)
      .toBe(actionsOf("Console Test Pty Ltd").getByRole("button", { name: "Reinstate" }).element());
  });

  // Verifies: REG-1202
  test("a software product action sends the version to the path of the software product", async () => {
    const requests = serveDetails(active);
    worker.use(
      http.post(
        "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/deactivate",
        async ({ params, request, response }) => {
          requests.push({ ...params, body: await request.json() });
          return response(204).empty();
        },
      ),
    );
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(
      actionsOf("Console Test Web").getByRole("button", { name: "Deactivate" }),
    );
    const dialog = page.getByRole("alertdialog", { name: "Deactivate Console Test Web?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription("Console Test Web changes from ACTIVE to INACTIVE.");
    await dialog.getByRole("button", { name: "Deactivate", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([
      { legalEntityId: id, softwareProductId: web.softwareProductId, body: { version: 7 } },
    ]);
  });

  // Verifies: REG-1202, REG-1212
  test("a brand deactivation names each software product that it changes, sends the version to the path of the brand, and shows the new state", async () => {
    const requests = serveDetails(active, brandDeactivated);
    worker.use(
      http.post(
        "/api/data-recipients/{legalEntityId}/brands/{brandId}/deactivate",
        async ({ params, request, response }) => {
          requests.push({ ...params, body: await request.json() });
          return response(204).empty();
        },
      ),
    );
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(actionsOf("Console Test").getByRole("button", { name: "Deactivate" }));
    const dialog = page.getByRole("alertdialog", { name: "Deactivate Console Test?" });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        "Console Test changes from ACTIVE to INACTIVE. " +
          "Console Test Web changes from ACTIVE to INACTIVE.",
      );
    await expectNoAxeViolations(document.body);
    await dialog.getByRole("button", { name: "Deactivate", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([
      { legalEntityId: id, brandId: brand.dataRecipientBrandId, body: { version: 7 } },
    ]);
    await expect
      .element(page.getByRole("dialog", { name: "Console Test is now INACTIVE" }))
      .toBeVisible();
    await expect.element(region("Console Test Web").getByText("INACTIVE")).toBeVisible();
    expect(buttonNames("Console Test")).toEqual(["Activate", "Remove"]);
  });

  // Verifies: REG-1213
  test("a brand removal says that its software products become REMOVED and that it cannot be reversed", async () => {
    serveDetails(active);
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(actionsOf("Console Test").getByRole("button", { name: "Remove" }));

    await expect
      .element(page.getByRole("alertdialog", { name: "Remove Console Test?" }))
      .toHaveAccessibleDescription(
        "Console Test changes from ACTIVE to REMOVED. " +
          "Console Test Web changes from ACTIVE to REMOVED. " +
          "Console Test Mobile changes from INACTIVE to REMOVED. " +
          "You cannot reverse this action.",
      );
  });
});
