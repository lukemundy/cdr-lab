// The stylesheet of the app (main.tsx), so that the checks see the pages as a browser shows them.
import "@cdr-lab/ui/consumer.css";
import "../site.css";

import { ConsoleApp } from "@cdr-lab/console/app";
import { createQueryClient } from "@cdr-lab/console/query";
import { render } from "@cdr-lab/ui/test/render";
import { createMemoryHistory } from "@tanstack/react-router";
import { setupWorker } from "msw/browser";
import { createOpenApiHttp } from "openapi-msw";
import { afterEach, beforeAll, beforeEach, expect } from "vitest";
import { page } from "vitest/browser";

import {
  type Account,
  type Brand,
  type Consent,
  type ProductDetail,
  type SessionCustomer,
  sessionEndedType,
  type Transaction,
} from "../api/client";
import type { paths } from "../api/schema.gen";
import { createAppRouter } from "../router";

export const http = createOpenApiHttp<paths>();

export function brandOf(brandName: string, fields: Partial<Brand> = {}) {
  return {
    brandName,
    demonstrationNotice: `${brandName} is not a real bank and this is just a demonstration.`,
    layout: "meadow",
    photos: [],
    ...fields,
  } satisfies Brand;
}

export function serveBrand(fields: Partial<Brand>) {
  worker.use(
    http.get("/api/brand", ({ response }) => response(200).json(brandOf("Example Bank", fields))),
  );
}

export const noSession = {
  type: "about:blank" as const,
  title: "Forbidden",
  status: 403,
  detail: "Sign in to online banking to see this page.",
};

export const sessionEnded = {
  type: sessionEndedType,
  title: "The session ended" as const,
  status: 403,
  detail: "The bank signed you out. Sign in again.",
};

// The customer of the tests is invented.
export const customer = { customerNumber: "00047291", firstName: "Zebracorn", lastName: "Quoll" };

// The response of a request after the end of the session expires the session cookie, so each later read of the
// session gets the 403 of a request with no session (services/dataholder/design/online-bank-site-api.md, DH-1744).
export function endSession() {
  worker.use(http.get("/api/session", ({ response }) => response(403).json(noSession)));
}

export function serveSession(signedIn: SessionCustomer = customer) {
  worker.use(http.get("/api/session", ({ response }) => response(200).json(signedIn)));
}

// The brand is an initial handler, so that a test can give a different one with worker.use.
export const worker = setupWorker(
  http.get("/api/brand", ({ response }) => response(200).json(brandOf("Example Bank"))),
  http.get("/api/session", ({ response }) => response(403).json(noSession)),
);

// services/dataholder/design/online-bank-site.md (Accessibility): each page is usable at 320 CSS px with no horizontal scroll.
export const narrowWidth = 320;

// No worker.stop() after the file: in Firefox, it also stops the mocks of the next test file.
beforeAll(async () => {
  await worker.start({ quiet: true });
});
beforeEach(async () => {
  await page.viewport(narrowWidth, 640);
});
afterEach(() => {
  worker.resetHandlers();
  for (const link of document.head.querySelectorAll("link[data-brand]")) link.remove();
});

// Each term of the description lists in within, with its value, as "term:value".
export function termsOf(within: ParentNode): string[] {
  return [...within.querySelectorAll("dl > div")].map(
    (row) => `${row.querySelector("dt")?.textContent}:${row.querySelector("dd")?.textContent}`,
  );
}

export function renderApp(url: string) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [url] }));
  render(<ConsoleApp router={router} queryClient={createQueryClient()} />);
  return router;
}

// The site API gives each product with its detail fields, and a 404 for an unknown product.
export function serveProducts(products: ProductDetail[]) {
  worker.use(
    http.get("/api/products", ({ response }) => response(200).json({ products })),
    http.get("/api/products/{productId}", ({ params, response }) => {
      const found = products.find((p) => p.productId === params.productId);
      return found
        ? response(200).json(found)
        : response(404).json({
            type: "about:blank",
            title: "Not Found",
            status: 404,
            detail: "The bank has no current product with this ID.",
          });
    }),
  );
}

// All product data in the tests is invented.
export function product(productId: string, fields: Partial<ProductDetail> = {}) {
  return {
    productId,
    name: "Everyday Account",
    description: "An account for everyday banking.",
    productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
    brand: "Example Bank",
    isTailored: false,
    lastUpdated: "2026-10-01T00:00:00Z",
    ...fields,
  } satisfies ProductDetail;
}

const notFound = {
  type: "about:blank",
  title: "Not Found",
  status: 404,
  detail: "You have no account with this ID.",
} as const;

// The site API gives a session, the accounts of its customer, and the transactions of each account in pages of
// pageSize. An account that is not in the list gives 404.
export function serveAccounts(
  accounts: Account[],
  transactions: Record<string, Transaction[]> = {},
  pageSize = 25,
) {
  const find = (accountId: string) => accounts.find((a) => a.accountId === accountId);
  serveSession();
  worker.use(
    http.get("/api/accounts", ({ response }) => response(200).json({ accounts })),
    http.get("/api/accounts/{accountId}/balance", ({ params, response }) => {
      const found = find(params.accountId);
      return found ? response(200).json(found.balance) : response(404).json(notFound);
    }),
    http.get("/api/accounts/{accountId}/transactions", ({ params, query, response }) => {
      if (!find(params.accountId)) return response(404).json(notFound);
      const all = transactions[params.accountId] ?? [];
      const requested = Number(query.get("page") ?? "1");
      return response(200).json({
        transactions: all.slice((requested - 1) * pageSize, requested * pageSize),
        totalRecords: all.length,
        totalPages: Math.ceil(all.length / pageSize),
      });
    }),
  );
}

// All account data in the tests is invented.
export function account(accountId: string, fields: Partial<Account> = {}) {
  return {
    accountId,
    creationDate: "2026-10-01",
    displayName: "Everyday",
    maskedNumber: "xxx-xxx xxxx5678",
    openStatus: "OPEN",
    productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
    productName: "Everyday Account",
    balance: { currentBalance: "1250.00", availableBalance: "1200.00", creditLimit: "0.00" },
    // The site API gives an open account the close action, and a closed account no action.
    actions: fields.openStatus === "CLOSED" ? [] : [{ action: "close", ends: [] }],
    ...fields,
  } satisfies Account;
}

// The site API gives a session and the consents of its customer.
export function serveConsents(consents: Consent[]) {
  serveSession();
  worker.use(http.get("/api/consents", ({ response }) => response(200).json({ consents })));
}

// All consent data in the tests is invented. The times are at midday in UTC, so that each date is the same in each
// time zone of a test browser.
export function consent(arrangementId: string, fields: Partial<Consent> = {}) {
  return {
    arrangementId,
    clientName: "Budget Planner",
    brandName: "Planner Co",
    legalEntityName: "Planner Holdings Pty Ltd",
    status: "CURRENT",
    confirmedAt: "2026-10-01T12:00:00Z",
    sharingDurationSeconds: 90 * 24 * 3600,
    sharingEndsAt: "2026-12-30T12:00:00Z",
    accounts: [
      {
        accountId: "account-1",
        creationDate: "2026-09-01",
        displayName: "Everyday",
        maskedNumber: "xxx-xxx xxxx5678",
        openStatus: "OPEN",
        productCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
        productName: "Everyday Account",
      },
    ],
    dataClusters: [
      {
        name: "Account name, type and balance",
        permissions: ["Name of account", "Type of account", "Account balance"],
        firstDisclosedAt: "2026-10-02T12:00:00Z",
        lastDisclosedAt: "2026-10-30T12:00:00Z",
      },
      { name: "Transaction details", permissions: ["Amounts", "Dates"] },
    ],
    ...fields,
  } satisfies Consent;
}

export function transaction(transactionId: string, fields: Partial<Transaction> = {}) {
  return {
    accountId: "account-1",
    transactionId,
    isDetailAvailable: false,
    type: "PAYMENT",
    status: "POSTED",
    description: "Corner Cafe",
    amount: "-4.50",
    reference: "",
    postingDateTime: "2026-10-30T02:00:00Z",
    executionDateTime: "2026-10-30T02:00:00Z",
    ...fields,
  } satisfies Transaction;
}

// A pending transaction has an execution time and no posting time.
export function pendingTransaction(transactionId: string, fields: Partial<Transaction> = {}) {
  const { postingDateTime: _posted, ...pending } = transaction(transactionId, {
    status: "PENDING",
    ...fields,
  });
  return pending satisfies Transaction;
}

// The brand stylesheet, the logo, the brand name and the catalogue of each preset brand, as the data holder
// serves them (services/dataholder/internal/adapter/consumerpages/golden_test.go writes the stylesheets).
const brandStyles = import.meta.glob<string>(
  "../../../../../services/dataholder/internal/adapter/consumerpages/testdata/brands/*.css",
  { query: "?url", import: "default", eager: true },
);
const images = import.meta.glob<string>("../../../../../services/dataholder/brands/*/logo.svg", {
  query: "?url",
  import: "default",
  eager: true,
});
const brandFiles = import.meta.glob<{ brandName: string; layout: Brand["layout"] }>(
  "../../../../../services/dataholder/brands/*/brand.json",
  { import: "default", eager: true },
);
const photoFiles = import.meta.glob<string>("../../../../../services/dataholder/brands/*/*.webp", {
  query: "?url",
  import: "default",
  eager: true,
});
const catalogues = import.meta.glob<{ products: ProductDetail[] }>(
  "../../../../../services/dataholder/brands/*/products.json",
  { import: "default", eager: true },
);

const presetDir = (name: string) => `../../../../../services/dataholder/brands/${name}`;

function required<T>(value: T | undefined, what: string): T {
  if (value === undefined) throw new Error(`the preset has no ${what}`);
  return value;
}

export const presets = Object.entries(brandStyles).map(([path, styles]) => {
  const name =
    path
      .split("/")
      .at(-1)
      ?.replace(/\.css$/, "") ?? path;
  const brandFile = required(brandFiles[`${presetDir(name)}/brand.json`], `${name}/brand.json`);
  const photos = Object.fromEntries(
    Object.entries(photoFiles)
      .filter(([file]) => file.startsWith(`${presetDir(name)}/`))
      .map(([file, url]) => [file.split("/").at(-1) ?? file, url]),
  );
  return {
    name,
    styles,
    logo: required(images[`${presetDir(name)}/logo.svg`], `${name}/logo.svg`),
    brandName: brandFile.brandName,
    layout: brandFile.layout,
    photos,
    products: required(catalogues[`${presetDir(name)}/products.json`], `${name}/products.json`)
      .products,
  };
});

export type Preset = (typeof presets)[number];

function loaded(image: HTMLImageElement): Promise<unknown> {
  return new Promise((resolve, reject) => {
    image.addEventListener("load", resolve, { once: true });
    image.addEventListener("error", reject, { once: true });
  });
}

// The test server has no /brand/logo.svg, so the first load of the logo fails. Firefox can send that error event after
// the change of the source, so the change starts again after an error.
function showLogo(image: HTMLImageElement, url: string, attempts = 3): Promise<unknown> {
  const done = loaded(image);
  image.src = url;
  return done.catch((error: unknown) => {
    if (attempts <= 1) throw error;
    return showLogo(image, url, attempts - 1);
  });
}

// showPreset serves the brand and the products of the preset, and loads its stylesheet. The logo of the page has the
// URL of the data holder, so the test changes it to the URL of the test server. (The browser keeps an image for its URL,
// so one URL cannot give the logo of each preset.)
export async function showPreset(preset: Preset, url: string) {
  worker.use(
    http.get("/api/brand", ({ response }) =>
      response(200).json(
        brandOf(preset.brandName, { layout: preset.layout, photos: Object.keys(preset.photos) }),
      ),
    ),
    // The data holder serves the photos at /brand/<file name>, and the test server has them at a different URL.
    http.untyped.get("/brand/:file", ({ params }) => {
      const photo = preset.photos[String(params.file)];
      return photo ? fetch(photo) : undefined;
    }),
  );
  serveProducts(preset.products);
  await new Promise((resolve, reject) => {
    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = preset.styles;
    link.dataset.brand = "";
    link.addEventListener("load", resolve);
    link.addEventListener("error", reject);
    document.head.append(link);
  });
  renderApp(url);
  await expect
    .poll(() => document.querySelector("footer")?.textContent ?? "")
    .toContain(preset.brandName);
  const logos = [...document.querySelectorAll<HTMLImageElement>('img[src="/brand/logo.svg"]')];
  await Promise.all(logos.map((logo) => showLogo(logo, preset.logo)));
  await document.fonts.ready;
}
