import { ProblemError, problemOf as problemOfService } from "@cdr-lab/web-platform/problem";
import type { Problem } from "@cdr-lab/ui/states";
import createClient from "openapi-fetch";

import type { components, paths } from "./schema.gen";

export type Brand = components["schemas"]["Brand"];
export type Product = components["schemas"]["Product"];
export type ProductList = components["schemas"]["ProductList"];
export type ProductDetail = components["schemas"]["ProductDetail"];
export type ProductCategory = components["schemas"]["ProductCategory"];
export type SessionCustomer = components["schemas"]["SessionCustomer"];
export type SignUp = components["schemas"]["SignUpRequest"];
export type Account = components["schemas"]["Account"];
export type AccountAction = components["schemas"]["AccountAction"];
export type Balance = components["schemas"]["Balance"];
export type Transaction = components["schemas"]["Transaction"];
export type TransactionList = components["schemas"]["TransactionList"];
export type Consent = components["schemas"]["Consent"];
export type ConsentAccount = components["schemas"]["ConsentAccount"];
export type DataCluster = components["schemas"]["DataCluster"];
export type OpenAccountRequest = components["schemas"]["OpenAccountRequest"];
export { productCategoryValues } from "./schema.gen";

const api = createClient<paths>();

export const sessionEndedType =
  "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0010" as const;

const readFailure: Problem = {
  title: "We cannot show this information now",
  detail: "Try again later.",
};

export function problemOf(error: unknown): Problem {
  return problemOfService(error, readFailure);
}

export async function getBrand(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/brand", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getProduct(productId: string, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/products/{productId}", {
    params: { path: { productId } },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listProducts(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/products", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function signUp(body: SignUp) {
  const { data, error, response } = await api.POST("/api/customers", { body });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

// The body has no password: the data holder refuses it, and the consumer can type a real password from habit.
// identifier is the email address or the customer number.
export async function signIn(identifier: string) {
  const { error, response } = await api.POST("/api/session", { body: { identifier } });
  if (error) throw new ProblemError(error, response.status);
}

export async function getSession(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/session", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function signOut() {
  const { error, response } = await api.DELETE("/api/session");
  if (error) throw new ProblemError(error, response.status);
}

export async function listAccounts(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/accounts", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getBalance(accountId: string, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/accounts/{accountId}/balance", {
    params: { path: { accountId } },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listTransactions(accountId: string, page: number, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/accounts/{accountId}/transactions", {
    params: { path: { accountId }, query: { page } },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

// ends has the arrangementId of each consent that the confirmation named.
export async function closeAccount(accountId: string, ends: string[]) {
  const { error, response } = await api.POST("/api/accounts/{accountId}/close", {
    params: { path: { accountId } },
    body: { ends },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function openAccount(body: OpenAccountRequest) {
  const { data, error, response } = await api.POST("/api/accounts", { body });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function withdrawConsent(arrangementId: string) {
  const { error, response } = await api.POST("/api/consents/{arrangementId}/withdraw", {
    params: { path: { arrangementId } },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function listConsents(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/consents", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}
