import { ProblemError, problemOf as problemOfService } from "@cdr-lab/console/problem";
import type { Problem } from "@cdr-lab/ui/states";
import createClient from "openapi-fetch";

import type { components, paths } from "./schema.gen";

export type Overview = components["schemas"]["Overview"];
export type Change = components["schemas"]["Change"];
export type CustomerSummary = components["schemas"]["CustomerSummary"];
export type CustomerDetail = components["schemas"]["CustomerDetail"];
export type Account = components["schemas"]["Account"];
export type AccountAction = components["schemas"]["AccountAction"];
export type Arrangement = components["schemas"]["Arrangement"];
export type ArrangementAction = components["schemas"]["ArrangementAction"];
export type Transaction = components["schemas"]["Transaction"];
export type TransactionList = components["schemas"]["TransactionList"];
export type SoftwareProductStatusList = components["schemas"]["SoftwareProductStatusList"];
export type SoftwareProductStatus = components["schemas"]["SoftwareProductStatus"];
export type Registration = components["schemas"]["Registration"];
export type ClientArrangement = components["schemas"]["ClientArrangement"];
export type ClientArrangementList = components["schemas"]["ClientArrangementList"];
export type RegistrationMetadata = components["schemas"]["RegistrationMetadata"];
export type CustomerQuery = NonNullable<paths["/api/customers"]["get"]["parameters"]["query"]>;
export type SoftwareProductStatusQuery = NonNullable<
  paths["/api/software-product-statuses"]["get"]["parameters"]["query"]
>;
export type RegistrationQuery = NonNullable<
  paths["/api/registrations"]["get"]["parameters"]["query"]
>;
export type CustomerCreate = components["schemas"]["CustomerCreateRequest"];
export type AccountCreate = components["schemas"]["AccountCreateRequest"];
export {
  extensionUTypeValues,
  nPPServiceValues,
  productCategoryValues,
  transactionStatusValues,
  transactionTypeValues,
} from "./schema.gen";
type TransactionCreateRequest = components["schemas"]["TransactionCreateRequest"];
type ExtendedDataCreate = NonNullable<TransactionCreateRequest["extendedData"]>;
// A select that the operator did not choose gives no value, and the data holder refuses it, not the UI.
export type TransactionCreate = Omit<
  TransactionCreateRequest,
  "type" | "status" | "extendedData"
> & {
  type?: TransactionType | undefined;
  status?: TransactionStatus | undefined;
  extendedData?:
    | (Omit<ExtendedDataCreate, "extensionUType" | "nppPayload"> & {
        extensionUType?: ExtensionUType | undefined;
        nppPayload: Omit<ExtendedDataCreate["nppPayload"], "service"> & {
          service?: NPPService | undefined;
        };
      })
    | undefined;
};
export type TransactionType = components["schemas"]["TransactionType"];
export type TransactionStatus = components["schemas"]["TransactionStatus"];
export type NPPService = components["schemas"]["NPPService"];
export type ExtensionUType = components["schemas"]["ExtensionUType"];

const api = createClient<paths>();

export function problemOf(error: unknown): Problem {
  return problemOfService(error, "the data holder");
}

export async function getBrand(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/brand", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getOverview(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/overview", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listCustomers(query: CustomerQuery, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/customers", {
    params: { query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

// Gives the customer number of the new customer, from the path of the Location header.
export async function createCustomer(create: CustomerCreate): Promise<string> {
  const { error, response } = await api.POST("/api/customers", { body: create });
  if (error) throw new ProblemError(error, response.status);
  const customerNumber = /^\/api\/customers\/([^/]+)$/.exec(
    response.headers.get("Location") ?? "",
  )?.[1];
  if (!customerNumber) throw new Error("the data holder gave no Location of the new customer");
  return decodeURIComponent(customerNumber);
}

export async function createAccount(customerNumber: string, create: AccountCreate) {
  const { error, response } = await api.POST("/api/customers/{customerNumber}/accounts", {
    params: { path: { customerNumber } },
    body: create,
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function getCustomer(customerNumber: string, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/customers/{customerNumber}", {
    params: { path: { customerNumber } },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

// ends are the IDs of the arrangements that the confirmation named.
export async function closeAccount(customerNumber: string, accountId: string, ends: string[]) {
  const { error, response } = await api.POST(
    "/api/customers/{customerNumber}/accounts/{accountId}/close",
    { params: { path: { customerNumber, accountId } }, body: { ends } },
  );
  if (error) throw new ProblemError(error, response.status);
}

// A withdrawal has no body: it ends only the arrangement of the path.
export async function withdrawArrangement(customerNumber: string, arrangementId: string) {
  const { error, response } = await api.POST(
    "/api/customers/{customerNumber}/arrangements/{arrangementId}/withdraw",
    { params: { path: { customerNumber, arrangementId } } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function listTransactions(
  customerNumber: string,
  accountId: string,
  query: { page?: number; pageSize: number },
  signal: AbortSignal,
) {
  const { data, error, response } = await api.GET(
    "/api/customers/{customerNumber}/accounts/{accountId}/transactions",
    { params: { path: { customerNumber, accountId }, query }, signal },
  );
  if (error) throw new ProblemError(error, response.status);
  return data;
}

// The console API has no detail of a transaction, so the UI does not need its Location.
export async function createTransaction(
  customerNumber: string,
  accountId: string,
  create: TransactionCreate,
) {
  const { error, response } = await api.POST(
    "/api/customers/{customerNumber}/accounts/{accountId}/transactions",
    {
      params: { path: { customerNumber, accountId } },
      // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the data holder checks the value, not the UI.
      body: create as TransactionCreateRequest,
    },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function listSoftwareProductStatuses(
  query: SoftwareProductStatusQuery,
  signal: AbortSignal,
) {
  const { data, error, response } = await api.GET("/api/software-product-statuses", {
    params: { query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listRegistrations(query: RegistrationQuery, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/registrations", {
    params: { query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getRegistration(clientId: string, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/registrations/{clientId}", {
    params: { path: { clientId } },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listRegistrationArrangements(
  clientId: string,
  query: { page?: number; pageSize: number },
  signal: AbortSignal,
) {
  const { data, error, response } = await api.GET("/api/registrations/{clientId}/arrangements", {
    params: { path: { clientId }, query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}
