import { renderAppShell, setUpConsoleTests } from "@cdr-lab/web-platform/test/shell";

export { definitionOf } from "@cdr-lab/web-platform/test/shell";
import { createMemoryHistory } from "@tanstack/react-router";
import { setupWorker } from "msw/browser";
import { createOpenApiHttp } from "openapi-msw";

import type {
  Account,
  Arrangement,
  ClientArrangement,
  CustomerDetail,
  CustomerSummary,
  Registration,
  RegistrationMetadata,
  Transaction,
} from "../api/client";
import type { paths } from "../api/schema.gen";
import { createAppRouter } from "../router";

export const http = createOpenApiHttp<paths>();
// The brand is an initial handler, so that a test can give a different one with worker.use.
export const worker = setupWorker(
  http.get("/api/brand", ({ response }) => response(200).json({ brandName: "Example Bank" })),
);

setUpConsoleTests(worker);

export function renderApp(url: string) {
  return renderAppShell(createAppRouter(createMemoryHistory({ initialEntries: [url] })));
}

// All customer data in the tests is invented.
export function customer(number: string, fields: Partial<CustomerSummary> = {}) {
  return {
    customerNumber: number,
    firstName: "Alex",
    lastName: "Example",
    emailAddress: `${number}@mail.test`,
    lastUpdated: "2026-10-15T05:30:00Z",
    openAccountCount: 1,
    currentArrangementCount: 0,
    ...fields,
  } satisfies CustomerSummary;
}

export function account(id: string, fields: Partial<Account> = {}) {
  return {
    accountId: id,
    displayName: "Everyday",
    productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
    productName: "Mock Everyday Account",
    bsb: "123456",
    accountNumber: "12345678",
    creditLimit: "0.00",
    currentBalance: "0.00",
    availableBalance: "0.00",
    openStatus: "OPEN",
    creationDate: "2026-10-15",
    actions: [
      { action: "close", ends: [] },
      { action: "create-transaction", ends: [] },
    ],
    ...fields,
  } satisfies Account;
}

export function arrangement(id: string, fields: Partial<Arrangement> = {}) {
  const { softwareId = "software-1", clientName = "Recipient Web" } = fields;
  return {
    arrangementId: id,
    clientId: "client-1",
    clientRegistered: true,
    softwareId,
    clientName,
    scopes: ["openid", "bank:accounts.basic:read"],
    accountIds: [],
    confirmedAt: "2026-10-15T05:30:00Z",
    sharingEndsAt: "2027-10-15T05:30:00Z",
    status: "CURRENT",
    actions: [{ action: "withdraw", ends: [{ arrangementId: id, softwareId, clientName }] }],
    ...fields,
  } satisfies Arrangement;
}

export function customerDetail(number: string, fields: Partial<CustomerDetail> = {}) {
  return {
    customerNumber: number,
    firstName: "Alex",
    lastName: "Example",
    emailAddress: `${number}@mail.test`,
    lastUpdated: "2026-10-15T05:30:00Z",
    accounts: [],
    arrangements: [],
    ...fields,
  } satisfies CustomerDetail;
}

export function transaction(id: string, fields: Partial<Transaction> = {}) {
  return {
    transactionId: id,
    type: "PAYMENT",
    status: "POSTED",
    description: "Coffee",
    amount: "-4.50",
    reference: "",
    postingDateTime: "2026-10-15T05:30:00Z",
    ...fields,
  } satisfies Transaction;
}

export function registration(
  clientId: string,
  fields: Partial<Omit<Registration, "metadata">> = {},
  metadata: Partial<RegistrationMetadata> = {},
) {
  return {
    registeredAt: "2026-10-15T05:30:00Z",
    lastUpdated: "2026-10-15T05:30:00Z",
    softwareProductStatus: "ACTIVE",
    arrangementCount: 0,
    currentArrangementCount: 0,
    ...fields,
    metadata: {
      client_id: clientId,
      client_id_issued_at: 1792042200,
      client_name: "Recipient Web",
      client_description: "A data recipient for tests",
      client_uri: "https://recipient.example",
      org_id: "brand-1",
      org_name: "Recipient",
      redirect_uris: ["https://recipient.example/cb"],
      logo_uri: "https://recipient.example/logo.png",
      jwks_uri: "https://recipient.example/jwks",
      revocation_uri: "https://recipient.example/revoke",
      recipient_base_uri: "https://recipient.example/cdr",
      token_endpoint_auth_method: "private_key_jwt",
      token_endpoint_auth_signing_alg: "PS256",
      grant_types: ["client_credentials", "authorization_code"],
      response_types: ["code"],
      application_type: "web",
      id_token_signed_response_alg: "PS256",
      authorization_signed_response_alg: "PS256",
      request_object_signing_alg: "PS256",
      software_id: "software-1",
      software_roles: "data-recipient-software-product",
      scope: "openid bank:accounts.basic:read",
      ...metadata,
    },
  } satisfies Registration;
}

export function clientArrangement(id: string, fields: Partial<ClientArrangement> = {}) {
  return {
    arrangementId: id,
    customerNumber: "12345678",
    lastName: "Example",
    status: "CURRENT",
    confirmedAt: "2026-10-15T05:30:00Z",
    sharingEndsAt: "2027-10-15T05:30:00Z",
    ...fields,
  } satisfies ClientArrangement;
}
