import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { afterEach, describe, expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import type {
  DataHolderBrand,
  DataHolderDetail,
  DataRecipientDetail,
  SoftwareProduct,
} from "./api/client";
import { http, renderDetail, worker } from "./test/app";
import { captureSaves } from "./test/saves";

const recipientId = "2a7f4c1e-8d3b-4e6a-9c5f-1b0d7e3a6c42";
const holderId = "6c1e9a3f-2b7d-4c8e-a5f0-3d9b1e7c4a26";
const thumbprint = "fj5bZBu5UoSrAua59pRyfSczfxvH2Jo9mn-XiLUWZ_M";
const csrText =
  "-----BEGIN CERTIFICATE REQUEST-----\nMIIBszCCAVmgAwIBAgIBATAKBggqhkjOPQQDAjAA\n-----END CERTIFICATE REQUEST-----\n";
const chain =
  "-----BEGIN CERTIFICATE-----\nMIIBleaf\n-----END CERTIFICATE-----\n" +
  "-----BEGIN CERTIFICATE-----\nMIIBintermediate\n-----END CERTIFICATE-----\n";

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
    clientCertificateActions: ["register", "issue", "remove"],
    ...fields,
  };
}

const web = product("Issue Test Web", {});
const retired = product("Issue Test Retired", {
  status: "REMOVED",
  clientCertificateActions: ["remove"],
});

function recipientWith(...products: SoftwareProduct[]): DataRecipientDetail {
  return {
    legalEntityId: recipientId,
    legalEntityName: "Issue Test Pty Ltd",
    accreditationNumber: "ADR-ISSUE-0001",
    accreditationLevel: "UNRESTRICTED",
    logoUri: "https://issue-test.example/logo.png",
    status: "ACTIVE",
    lastUpdated: "2026-10-27T02:05:07Z",
    version: 6,
    actions: [],
    dataRecipientBrands: [
      {
        dataRecipientBrandId: "brand-issue-test",
        brandName: "Issue Test",
        logoUri: "https://issue-test.example/logo.png",
        status: "ACTIVE",
        actions: [],
        softwareProducts: products,
      },
    ],
  };
}

function holderBrand(name: string, fields: Partial<DataHolderBrand>): DataHolderBrand {
  return {
    brandName: name,
    industries: ["banking"],
    logoUri: "https://issue-holder.example/logo.png",
    status: "ACTIVE",
    actions: [],
    serverCertificateActions: [],
    endpointDetail: { publicBaseUri: "https://api.issue-holder.example" },
    ...fields,
  };
}

const bank = holderBrand("Issue Bank", {
  dataHolderBrandId: "dhb-issue-bank",
  serverCertificateActions: ["issue"],
});
const products = holderBrand("Issue Products", { interimId: "interim-issue-products" });

const holder: DataHolderDetail = {
  legalEntityId: holderId,
  legalEntityName: "Issue Holdings Ltd",
  logoUri: "https://issue-holder.example/logo.png",
  status: "ACTIVE",
  lastUpdated: "2026-10-27T02:05:07Z",
  version: 3,
  actions: [],
  dataHolderBrands: [bank, products],
};

// The detail depends on the issue requests, not on the count of reads, because a read of a previous test can
// arrive late.
function serveRecipient(before: DataRecipientDetail, after: DataRecipientDetail = before) {
  const requests: unknown[] = [];
  worker.use(
    http.get("/api/data-recipients/{legalEntityId}", ({ response }) =>
      response(200).json(requests.length === 0 ? before : after),
    ),
  );
  return requests;
}

type Reply = { status: 200 } | { status: 400 | 409; detail: string };

function problem(status: 400 | 409, detail: string) {
  return {
    type: "about:blank" as const,
    title: status === 400 ? "Bad Request" : "Conflict",
    status,
    detail,
  };
}

function serveClientIssue(requests: unknown[], reply: Reply) {
  worker.use(
    http.post(
      "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue",
      async ({ params, request, response }) => {
        requests.push({ ...params, body: await request.json() });
        if (reply.status === 200)
          return response(200).json({ certificateChain: chain, thumbprint });
        return response(reply.status).json(problem(reply.status, reply.detail));
      },
    ),
  );
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const clientDialog = () =>
  page.getByRole("dialog", { name: "Issue a client certificate for Issue Test Web" });
const csrField = () => clientDialog().getByRole("textbox", { name: "PEM CSR" });

async function openClientIssue() {
  await userEvent.click(
    region("Issue Test Web").getByRole("button", { name: "Issue certificate" }),
  );
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe("the issue of a client certificate", () => {
  // Verifies: REG-1229
  test("only a software product that the rules permit offers the issue", async () => {
    serveRecipient(recipientWith(web, retired));
    await renderDetail(`/data-recipients/${recipientId}`);

    await expect
      .element(region("Issue Test Web").getByRole("button", { name: "Issue certificate" }))
      .toBeVisible();
    expect(
      region("Issue Test Retired").getByRole("button", { name: "Issue certificate" }).query(),
    ).toBeNull();
  });

  // Verifies: REG-1233, REG-1234
  test("the help text gives the subject that the Register requires and says that it needs no private key", async () => {
    serveRecipient(recipientWith(web));
    await renderDetail(`/data-recipients/${recipientId}`);

    await openClientIssue();

    await expect
      .element(csrField())
      .toHaveAccessibleDescription(
        "Paste one PEM CSR, or choose a file. The Register needs only the CSR, and never the private key. " +
          'The CN must be the software product name "Issue Test Web", and the O must be the brand name "Issue Test".',
      );
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1206, REG-1238, REG-1239
  test("an issue sends the version and the CSR, and the result shows the chain once with a save as a file", async () => {
    const requests = serveRecipient(
      recipientWith(web),
      recipientWith({ ...web, clientCertificateThumbprints: [thumbprint] }),
    );
    serveClientIssue(requests, { status: 200 });
    const saves = captureSaves();
    await renderDetail(`/data-recipients/${recipientId}`);

    await openClientIssue();
    await userEvent.upload(
      clientDialog().getByLabelText("CSR file"),
      new File([csrText], "client.csr", { type: "application/pkcs10" }),
    );
    await expect.element(csrField()).toHaveValue(csrText);
    await clientDialog().getByRole("button", { name: "Issue", exact: true }).click();

    const heading = clientDialog().getByRole("heading", {
      name: "Client certificate issued for Issue Test Web",
    });
    await expect.element(heading).toHaveFocus();
    expect(requests).toEqual([
      {
        legalEntityId: recipientId,
        softwareProductId: web.softwareProductId,
        body: { version: 6, csr: csrText },
      },
    ]);
    const chainField = clientDialog().getByRole("textbox", { name: "PEM chain" });
    await expect.element(chainField).toHaveValue(chain);
    await expect.element(chainField).toHaveAttribute("readonly");
    await expect.element(clientDialog().getByText(thumbprint, { exact: true })).toBeVisible();
    await expect
      .element(
        clientDialog().getByText(
          "The Register registered the certificate for Issue Test Web. The token endpoint accepts it now.",
        ),
      )
      .toBeVisible();
    await expect
      .element(
        clientDialog().getByText(
          "The Register does not keep the certificate, so the console cannot show it again. Save the chain before you close this dialog.",
        ),
      )
      .toBeVisible();
    await expectNoAxeViolations(document.body);

    await clientDialog().getByRole("button", { name: "Save as file" }).click();

    expect(saves.map((s) => s.name)).toEqual([`client-certificate-${web.softwareProductId}.pem`]);
    expect(await saves[0]?.text).toBe(chain);
    await expect.element(heading).toBeVisible();

    await userEvent.keyboard("{Escape}");

    await expect.element(clientDialog()).not.toBeInTheDocument();
    await expect
      .element(region("Issue Test Web").getByText(thumbprint, { exact: true }))
      .toBeVisible();
  });

  // Verifies: REG-1236
  test("a refused CSR shows the detail of the 400 as the error of the field", async () => {
    const detail =
      'The CSR breaks these rules: subject CN must be the software product name "Issue Test Web".';
    serveClientIssue(serveRecipient(recipientWith(web)), { status: 400, detail });
    await renderDetail(`/data-recipients/${recipientId}`);

    await openClientIssue();
    await userEvent.fill(csrField(), csrText);
    await clientDialog().getByRole("button", { name: "Issue", exact: true }).click();

    await expect.element(csrField()).toHaveAttribute("aria-invalid", "true");
    await expect.element(csrField()).toHaveAccessibleDescription(expect.stringContaining(detail));
    await expect.element(csrField()).toHaveValue(csrText);
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1237, WCON-0015
  test("a different refusal shows in an alert in the dialog, and the views show the current state", async () => {
    const detail =
      "The participant changed after the version in the request. The Register changed nothing.";
    const removed = {
      ...web,
      status: "REMOVED" as const,
      clientCertificateActions: ["remove" as const],
    };
    serveClientIssue(serveRecipient(recipientWith(web), recipientWith(removed)), {
      status: 409,
      detail,
    });
    await renderDetail(`/data-recipients/${recipientId}`);

    await openClientIssue();
    await userEvent.fill(csrField(), csrText);
    await clientDialog().getByRole("button", { name: "Issue", exact: true }).click();

    await expect.element(clientDialog().getByRole("alert")).toMatchTextContent(detail);

    await clientDialog().getByRole("button", { name: "Cancel" }).click();

    await expect.element(clientDialog()).not.toBeInTheDocument();
    await expect.element(region("Issue Test Web").getByText("REMOVED")).toBeVisible();
    expect(
      region("Issue Test Web").getByRole("button", { name: "Issue certificate" }).query(),
    ).toBeNull();
  });
});

describe("the issue of a server certificate", () => {
  // Verifies: REG-1207, REG-1229, REG-1244
  test("only a brand that the rules permit offers the issue, and the result names the server certificate", async () => {
    const requests: unknown[] = [];
    worker.use(
      http.get("/api/data-holders/{legalEntityId}", ({ response }) => response(200).json(holder)),
      http.post(
        "/api/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue",
        async ({ params, request, response }) => {
          requests.push({ ...params, body: await request.json() });
          return response(200).json({ certificateChain: chain, thumbprint });
        },
      ),
    );
    const saves = captureSaves();
    await renderDetail(`/data-holders/${holderId}`);

    expect(
      region("Issue Products").getByRole("button", { name: "Issue certificate" }).query(),
    ).toBeNull();
    await userEvent.click(region("Issue Bank").getByRole("button", { name: "Issue certificate" }));
    const dialog = page.getByRole("dialog", {
      name: "Issue a server certificate for Issue Bank",
    });
    const field = dialog.getByRole("textbox", { name: "PEM CSR" });
    await expect
      .element(field)
      .toHaveAccessibleDescription(
        expect.stringContaining(
          'The CN must be the primary DNS name, and the O must be the brand name "Issue Bank" or the legal entity name "Issue Holdings Ltd".',
        ),
      );
    await userEvent.fill(field, csrText);
    await dialog.getByRole("button", { name: "Issue", exact: true }).click();

    await expect
      .element(dialog.getByRole("heading", { name: "Server certificate issued for Issue Bank" }))
      .toHaveFocus();
    expect(requests).toEqual([
      { legalEntityId: holderId, brandId: "dhb-issue-bank", body: { version: 3, csr: csrText } },
    ]);
    await expect
      .element(
        dialog.getByText(
          "The Register does not register a server certificate. A TLS client that trusts the CDR CA bundle accepts it.",
        ),
      )
      .toBeVisible();
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Save as file" }).click();

    expect(saves.map((s) => s.name)).toEqual(["server-certificate-dhb-issue-bank.pem"]);
    expect(await saves[0]?.text).toBe(chain);

    // The footer button comes before the close button of the dialog title.
    await dialog.getByRole("button", { name: "Close", exact: true }).first().click();

    await expect.element(dialog).not.toBeInTheDocument();
  });
});
