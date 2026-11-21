import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { DataRecipientDetail, SoftwareProduct } from "../api/client";
import { http, worker, renderDetail } from "../test/app";

const id = "5f3b0f0e-9c7d-4f5e-8a61-2d4c6b8e0a13";
const thumbprint = "fj5bZBu5UoSrAua59pRyfSczfxvH2Jo9mn-XiLUWZ_M";
const pemText =
  "-----BEGIN CERTIFICATE-----\nMIIBszCCAVmgAwIBAgIBATAKBggqhkjOPQQDAjAA\n-----END CERTIFICATE-----\n";

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
    clientCertificateActions: ["register", "remove"],
    ...fields,
  };
}

const web = product("Certificate Test Web", {});
const retired = product("Certificate Test Retired", {
  status: "REMOVED",
  clientCertificateThumbprints: [thumbprint],
  clientCertificateActions: ["remove"],
});

function detailWith(...products: SoftwareProduct[]): DataRecipientDetail {
  return {
    legalEntityId: id,
    legalEntityName: "Console Certificate Test Pty Ltd",
    accreditationNumber: "ADR-MOCK-0012",
    accreditationLevel: "UNRESTRICTED",
    logoUri: "https://console-certificate-test.example/logo.png",
    status: "ACTIVE",
    lastUpdated: "2026-10-23T02:05:07Z",
    version: 4,
    actions: [],
    dataRecipientBrands: [
      {
        dataRecipientBrandId: "brand-certificate-test",
        brandName: "Certificate Test",
        logoUri: "https://console-certificate-test.example/logo.png",
        status: "ACTIVE",
        actions: [],
        softwareProducts: products,
      },
    ],
  };
}

const registered = detailWith({ ...web, clientCertificateThumbprints: [thumbprint] }, retired);

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

function problem(status: 400 | 409, detail: string) {
  return {
    type: "about:blank" as const,
    title: status === 400 ? "Bad Request" : "Conflict",
    status,
    detail,
  };
}

function serveRegistration(
  requests: unknown[],
  reply: { status: 204 } | { status: 400 | 409; detail: string },
) {
  worker.use(
    http.post(
      "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates",
      async ({ params, request, response }) => {
        requests.push({ ...params, body: await request.json() });
        if (reply.status === 204) return response(204).empty();
        return response(reply.status).json(problem(reply.status, reply.detail));
      },
    ),
  );
}

const region = (name: string) => page.getByRole("region", { name, exact: true });
const registrationDialog = () =>
  page.getByRole("dialog", { name: "Register a client certificate for Certificate Test Web" });
const certificateField = () =>
  registrationDialog().getByRole("textbox", { name: "PEM certificate" });

async function openRegistration() {
  await userEvent.click(
    region("Certificate Test Web").getByRole("button", { name: "Register certificate" }),
  );
}

describe("the client certificates of a software product", () => {
  test("a REMOVED software product offers the removal and no registration", async () => {
    serveDetails(detailWith(web, retired));
    await renderDetail(`/data-recipients/${id}`);

    await expect.element(region("Certificate Test Retired")).toBeVisible();
    expect(
      region("Certificate Test Retired")
        .getByRole("button", { name: "Register certificate" })
        .query(),
    ).toBeNull();
    await expect
      .element(
        region("Certificate Test Retired").getByRole("button", {
          name: `Remove the client certificate ${thumbprint}`,
        }),
      )
      .toBeVisible();
    await expect
      .element(region("Certificate Test Web").getByRole("button", { name: "Register certificate" }))
      .toBeVisible();
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1204, REG-1211
  test("a pasted certificate is registered with the version, with no confirmation, and the thumbprint shows", async () => {
    const requests = serveDetails(detailWith(web, retired), registered);
    serveRegistration(requests, { status: 204 });
    await renderDetail(`/data-recipients/${id}`);

    await openRegistration();
    await expect.element(certificateField()).toHaveFocus();
    await expectNoAxeViolations(document.body);
    await userEvent.fill(certificateField(), pemText);
    await registrationDialog().getByRole("button", { name: "Register", exact: true }).click();

    await expect.element(registrationDialog()).not.toBeInTheDocument();
    expect(requests).toEqual([
      {
        legalEntityId: id,
        softwareProductId: web.softwareProductId,
        body: { version: 4, certificate: pemText },
      },
    ]);
    await expect
      .element(
        page.getByRole("dialog", {
          name: "Client certificate registered for Certificate Test Web",
        }),
      )
      .toBeVisible();
    await expect
      .element(region("Certificate Test Web").getByText(thumbprint, { exact: true }))
      .toBeVisible();
  });

  // Verifies: REG-1225
  test("the text of a selected file shows in the field before the registration", async () => {
    serveDetails(detailWith(web));
    await renderDetail(`/data-recipients/${id}`);

    await openRegistration();
    await userEvent.upload(
      registrationDialog().getByLabelText("Certificate file"),
      new File([pemText], "client.pem", { type: "application/x-pem-file" }),
    );

    await expect.element(certificateField()).toHaveValue(pemText);
  });

  // Verifies: REG-1221
  test("text that is not one PEM certificate shows the detail of the 400 as the error of the field", async () => {
    const detail = "The certificate must be one PEM certificate, with no other text.";
    serveRegistration(serveDetails(detailWith(web)), { status: 400, detail });
    await renderDetail(`/data-recipients/${id}`);

    await openRegistration();
    await userEvent.fill(certificateField(), "not a certificate");
    await registrationDialog().getByRole("button", { name: "Register", exact: true }).click();

    await expect.element(certificateField()).toHaveAttribute("aria-invalid", "true");
    await expect
      .element(certificateField())
      .toHaveAccessibleDescription(expect.stringContaining(detail));
    await expect.element(certificateField()).toHaveValue("not a certificate");
    await expectNoAxeViolations(document.body);
  });

  // Verifies: WCON-0014, WCON-0015
  test("a refused registration shows the detail in the dialog and the current state, also when the registration is no longer permitted", async () => {
    const detail =
      "The data recipient changed after the version in the request. The Register changed nothing.";
    const removed = {
      ...web,
      status: "REMOVED" as const,
      clientCertificateActions: ["remove" as const],
    };
    serveRegistration(serveDetails(detailWith(web), detailWith(removed)), { status: 409, detail });
    await renderDetail(`/data-recipients/${id}`);

    await openRegistration();
    await userEvent.fill(certificateField(), pemText);
    await registrationDialog().getByRole("button", { name: "Register", exact: true }).click();

    await expect.element(registrationDialog().getByRole("alert")).toMatchTextContent(detail);

    await registrationDialog().getByRole("button", { name: "Cancel" }).click();

    await expect.element(registrationDialog()).not.toBeInTheDocument();
    await expect.element(region("Certificate Test Web").getByText("REMOVED")).toBeVisible();
    expect(
      region("Certificate Test Web").getByRole("button", { name: "Register certificate" }).query(),
    ).toBeNull();
  });

  // Verifies: REG-1205, REG-1211
  test("the confirmation of a removal names the thumbprint and the software product, and the removal sends the version", async () => {
    const requests = serveDetails(registered, detailWith(web, retired));
    worker.use(
      http.post(
        "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/{thumbprint}/remove",
        async ({ params, request, response }) => {
          requests.push({ ...params, body: await request.json() });
          return response(204).empty();
        },
      ),
    );
    await renderDetail(`/data-recipients/${id}`);

    await userEvent.click(
      region("Certificate Test Web").getByRole("button", {
        name: `Remove the client certificate ${thumbprint}`,
      }),
    );

    const dialog = page.getByRole("alertdialog", {
      name: "Remove the client certificate of Certificate Test Web?",
    });
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(
        `The token endpoint refuses the client certificate ${thumbprint} for Certificate Test Web at the next token request. ` +
          "Access tokens that exist stay valid until they expire. " +
          "To trust the certificate again, register its PEM certificate again.",
      );
    await expectNoAxeViolations(document.body);

    await dialog.getByRole("button", { name: "Remove", exact: true }).click();

    await expect.element(dialog).not.toBeInTheDocument();
    expect(requests).toEqual([
      {
        legalEntityId: id,
        softwareProductId: web.softwareProductId,
        thumbprint,
        body: { version: 4 },
      },
    ]);
    await expect
      .element(
        page.getByRole("dialog", { name: "Client certificate removed from Certificate Test Web" }),
      )
      .toBeVisible();
    expect(
      region("Certificate Test Web").getByText(thumbprint, { exact: true }).query(),
    ).toBeNull();
  });
});
