import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { afterEach, describe, expect, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { RevocableCertificate, Revocation } from "../api/client";
import { http, renderApp, worker } from "../test/app";
import { captureSaves } from "../test/saves";

const caBundle =
  "-----BEGIN CERTIFICATE-----\nMIIBintermediate\n-----END CERTIFICATE-----\n" +
  "-----BEGIN CERTIFICATE-----\nMIIBroot\n-----END CERTIFICATE-----\n";
const pemText =
  "-----BEGIN CERTIFICATE-----\nMIIBszCCAVmgAwIBAgIBATAKBggqhkjOPQQDAjAA\n-----END CERTIFICATE-----\n";
const thumbprint = "fj5bZBu5UoSrAua59pRyfSczfxvH2Jo9mn-XiLUWZ_M";

const checked: RevocableCertificate = {
  serialNumber: "0A1B2C",
  subject: "CN=Revoke Test Web,OU=Consumer Data Right,O=Revoke Test,C=AU",
  thumbprint,
  notAfter: "2027-10-27T01:02:03Z",
};

const older: Revocation = {
  serialNumber: "77",
  subject: "CN=Old Web,O=Old,C=AU",
  thumbprint: "old-thumbprint",
  revokedAt: "2026-10-20T01:02:03Z",
  reason: "superseded",
};

const revoked: Revocation = {
  serialNumber: checked.serialNumber,
  subject: checked.subject,
  thumbprint,
  revokedAt: "2026-10-27T03:04:05Z",
  reason: "keyCompromise",
};

const fieldErrorsType =
  "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002" as const;

type Requests = { checks: unknown[]; revocations: unknown[]; pages: (string | null)[] };

// The list depends on the revocation requests, not on the count of reads, because a read of a previous test can
// arrive late.
function serve(after: Revocation[] = [revoked, older]): Requests {
  const requests: Requests = { checks: [], revocations: [], pages: [] };
  worker.use(
    http.get("/api/cdr-ca-bundle", ({ response }) => response(200).json({ caBundle })),
    http.get("/api/cdr-certificates/revocations", ({ request, response }) => {
      requests.pages.push(new URL(request.url).searchParams.get("page"));
      const data = requests.revocations.length === 0 ? [older] : after;
      return response(200).json({ data, totalRecords: data.length });
    }),
  );
  return requests;
}

function serveCheck(requests: Requests, result: "success" | "field errors" | "revoked") {
  worker.use(
    http.post("/api/cdr-certificates/check-revocation", async ({ request, response }) => {
      requests.checks.push(await request.json());
      if (result === "success") return response(200).json(checked);
      if (result === "revoked") {
        return response(409).json({
          type: "about:blank",
          title: "Conflict",
          status: 409,
          detail:
            "The certificate is already revoked, with the reason keyCompromise. The Register changed nothing.",
        });
      }
      return response(400).json({
        type: fieldErrorsType,
        title: "Values in the request are not valid",
        status: 400,
        detail: "One or more values in the request are not valid.",
        errors: [
          {
            pointer: "#/certificate",
            detail: "The mock CDR intermediate did not sign the certificate.",
          },
          { pointer: "#/reason", detail: "The reason is required." },
        ],
      });
    }),
  );
}

function serveRevoke(requests: Requests, result: "success" | "revoked") {
  worker.use(
    http.post("/api/cdr-certificates/revoke", async ({ request, response }) => {
      const body = await request.json();
      if (result === "revoked") {
        return response(409).json({
          type: "about:blank",
          title: "Conflict",
          status: 409,
          detail:
            "The certificate is already revoked, with the reason superseded. The Register changed nothing.",
        });
      }
      requests.revocations.push(body);
      return response(204).empty();
    }),
  );
}

const revokeDialog = () => page.getByRole("dialog", { name: "Revoke a CDR certificate" });
const certificateField = () => revokeDialog().getByRole("textbox", { name: "PEM certificate" });
const reasonField = () => revokeDialog().getByRole("combobox", { name: "Reason" });
const confirmation = () => page.getByRole("alertdialog", { name: "Revoke this CDR certificate?" });
const table = () => page.getByRole("table", { name: "Revocations" });

async function openRevocation() {
  await page.getByRole("button", { name: "Revoke certificate" }).click();
  await expect.element(revokeDialog()).toBeVisible();
}

async function giveRevocation(reason = "Key compromise") {
  await userEvent.upload(
    revokeDialog().getByLabelText("Certificate file"),
    new File([pemText], "client-certificate.pem", { type: "application/pem-certificate-chain" }),
  );
  await expect.element(certificateField()).toHaveValue(pemText);
  await userEvent.selectOptions(reasonField(), reason);
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe("the CDR certificates page", () => {
  // Verifies: REG-1230, REG-1231
  test("the side navigation marks the section, and the page shows the CA bundle and the revocations", async () => {
    serve();

    renderApp("/cdr-certificates");

    const navigation = page.getByRole("navigation", { name: "Main" });
    await expect
      .element(navigation.getByRole("link", { name: "CDR certificates" }))
      .toHaveAttribute("aria-current", "page");
    const bundle = page.getByRole("textbox", { name: "PEM CA bundle" });
    await expect.element(bundle).toHaveValue(caBundle);
    await expect.element(bundle).toHaveAttribute("readonly");
    const row = table().getByRole("row").nth(1);
    await expect.element(row.getByText("77", { exact: true })).toBeVisible();
    await expect.element(row.getByText("CN=Old Web,O=Old,C=AU", { exact: true })).toBeVisible();
    await expect.element(row.getByText("old-thumbprint", { exact: true })).toBeVisible();
    await expect.element(row.getByText("Superseded", { exact: true })).toBeVisible();
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1208, REG-1247, REG-1248
  test("Save as file saves the CA bundle as cdr-ca-bundle.pem with the text of the Register", async () => {
    serve();
    const saves = captureSaves();
    renderApp("/cdr-certificates");

    await page.getByRole("button", { name: "Save as file" }).click();

    expect(saves.map((s) => s.name)).toEqual(["cdr-ca-bundle.pem"]);
    expect(await saves[0]?.text).toBe(caBundle);
  });

  test("the page of the list is in the URL", async () => {
    const requests = serve();

    renderApp("/cdr-certificates?page=2");

    await expect.element(table()).toBeVisible();
    await vi.waitFor(() => expect(requests.pages).toContain("2"));
  });
});

describe("the revocation of a CDR certificate", () => {
  // Verifies: REG-1250, REG-1251
  test("the dialog has a certificate field, a file input and a reason with no selected value", async () => {
    serve();
    renderApp("/cdr-certificates");

    await openRevocation();

    await expect
      .element(certificateField())
      .toHaveAccessibleDescription(
        "Paste one PEM certificate, or choose the file of an issue. The Register revokes the first certificate.",
      );
    await expect.element(reasonField()).toHaveValue("");
    const options = reasonField().element().querySelectorAll("option");
    expect([...options].map((o) => o.textContent)).toEqual([
      "Choose a reason",
      "Unspecified",
      "Key compromise",
      "Affiliation changed",
      "Superseded",
      "Cessation of operation",
      "Privilege withdrawn",
    ]);
    await expectNoAxeViolations(document.body);
  });

  // Verifies: REG-1209, REG-1253, REG-1256
  test("Continue checks the request, the confirmation names the certificate, and Revoke revokes it", async () => {
    const requests = serve();
    serveCheck(requests, "success");
    serveRevoke(requests, "success");
    renderApp("/cdr-certificates?page=2");

    await openRevocation();
    await giveRevocation();
    await revokeDialog().getByRole("button", { name: "Continue" }).click();

    await expect.element(confirmation()).toBeVisible();
    expect(requests.checks).toEqual([{ certificate: pemText, reason: "keyCompromise" }]);
    expect(requests.revocations).toEqual([]);
    const text = confirmation().element().textContent;
    expect(text).toContain(checked.subject);
    expect(text).toContain(checked.serialNumber);
    expect(text).toContain(thumbprint);
    expect(text).toContain('"Key compromise"');
    expect(text).toContain("its validity ends at");
    expect(text).toContain(
      "The Register refuses it at once, and other relying parties refuse it within 1 minute.",
    );
    expect(text).toContain(
      "A registration of the certificate for a software product does not change.",
    );
    expect(text).toContain("You cannot reverse this action.");
    await expect.element(confirmation().getByRole("button", { name: "Cancel" })).toHaveFocus();
    await expectNoAxeViolations(document.body);

    await confirmation().getByRole("button", { name: "Revoke", exact: true }).click();

    await expect.element(confirmation()).not.toBeInTheDocument();
    await expect.element(revokeDialog()).not.toBeInTheDocument();
    expect(requests.revocations).toEqual([{ certificate: pemText, reason: "keyCompromise" }]);
    await expect
      .element(page.getByRole("dialog", { name: `Certificate ${checked.serialNumber} revoked` }))
      .toBeVisible();
    await expect
      .element(table().getByRole("row").nth(1).getByText("Key compromise", { exact: true }))
      .toBeVisible();
    await expect
      .element(table().getByRole("row").nth(1).getByText(thumbprint, { exact: true }))
      .toBeVisible();
    await vi.waitFor(() => expect(requests.pages.at(-1)).toBeNull());
  });

  test("Cancel in the confirmation goes back to the dialog with the values, and revokes nothing", async () => {
    const requests = serve();
    serveCheck(requests, "success");
    serveRevoke(requests, "success");
    renderApp("/cdr-certificates");

    await openRevocation();
    await giveRevocation("Superseded");
    await revokeDialog().getByRole("button", { name: "Continue" }).click();
    await confirmation().getByRole("button", { name: "Cancel" }).click();

    await expect.element(confirmation()).not.toBeInTheDocument();
    await expect.element(certificateField()).toHaveValue(pemText);
    await expect.element(reasonField()).toHaveValue("superseded");
    await expect.element(revokeDialog().getByRole("button", { name: "Continue" })).toHaveFocus();
    expect(requests.revocations).toEqual([]);
  });

  // Verifies: REG-1254
  test("the field errors of a refused check show at their fields and in the error summary", async () => {
    const requests = serve();
    serveCheck(requests, "field errors");
    renderApp("/cdr-certificates");

    await openRevocation();
    await userEvent.fill(certificateField(), pemText);
    await revokeDialog().getByRole("button", { name: "Continue" }).click();

    const summary = revokeDialog().getByRole("alert");
    await expect.element(summary).toBeVisible();
    expect(summary.element().textContent).toContain("There is a problem");
    await expect
      .element(
        summary.getByRole("link", {
          name: "The mock CDR intermediate did not sign the certificate.",
        }),
      )
      .toBeVisible();
    await expect.element(certificateField()).toHaveAttribute("aria-invalid", "true");
    await expect
      .element(certificateField())
      .toHaveAccessibleDescription(
        expect.stringContaining("The mock CDR intermediate did not sign the certificate."),
      );
    await expect.element(reasonField()).toHaveAttribute("aria-invalid", "true");
    expect(requests.checks).toEqual([{ certificate: pemText }]);
    expect(confirmation().query()).toBeNull();
    await expectNoAxeViolations(document.body);

    await summary.getByRole("link", { name: "The reason is required." }).click();
    await expect.element(reasonField()).toHaveFocus();
  });

  // Verifies: REG-1255
  test("a certificate that is already revoked shows the detail in the error summary, with no field error", async () => {
    const requests = serve();
    serveCheck(requests, "revoked");
    renderApp("/cdr-certificates");

    await openRevocation();
    await giveRevocation();
    await revokeDialog().getByRole("button", { name: "Continue" }).click();

    await expect.element(revokeDialog().getByRole("alert")).toBeVisible();
    expect(revokeDialog().getByRole("alert").element().textContent).toContain(
      "already revoked, with the reason keyCompromise",
    );
    await expect.element(certificateField()).not.toHaveAttribute("aria-invalid", "true");
    expect(confirmation().query()).toBeNull();
  });

  // Verifies: WCON-0014
  test("a revocation that the Register refuses after the check shows the problem in the confirmation", async () => {
    const requests = serve();
    serveCheck(requests, "success");
    serveRevoke(requests, "revoked");
    renderApp("/cdr-certificates");

    await openRevocation();
    await giveRevocation();
    await revokeDialog().getByRole("button", { name: "Continue" }).click();
    await confirmation().getByRole("button", { name: "Revoke", exact: true }).click();

    await expect.element(confirmation().getByRole("alert")).toBeVisible();
    expect(confirmation().getByRole("alert").element().textContent).toContain(
      "already revoked, with the reason superseded",
    );
  });
});
