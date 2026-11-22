import {
  consoleReadFailure,
  ProblemError,
  problemOf as problemOfService,
} from "@cdr-lab/web-platform/problem";
import type { Problem } from "@cdr-lab/ui/states";
import createClient from "openapi-fetch";

import type { components, paths } from "./schema.gen";

export type DataRecipientSummary = components["schemas"]["DataRecipientSummary"];
export type DataRecipientStatus = components["schemas"]["DataRecipientStatus"];
export type KeySetList = components["schemas"]["KeySetList"];
export type KeySet = components["schemas"]["KeySet"];
export type KeyState = components["schemas"]["KeyState"];
export type KeyAction = components["schemas"]["KeyAction"];
export type DataRecipientQuery = NonNullable<
  paths["/api/data-recipients"]["get"]["parameters"]["query"]
>;

export type DataHolderSummary = components["schemas"]["DataHolderSummary"];
export type DataHolderStatus = components["schemas"]["DataHolderStatus"];
export type DataHolderQuery = NonNullable<paths["/api/data-holders"]["get"]["parameters"]["query"]>;

export type DataRecipientDetail = components["schemas"]["DataRecipientDetail"];
export type DataRecipientBrand = components["schemas"]["DataRecipientBrand"];
export type SoftwareProduct = components["schemas"]["SoftwareProduct"];
export type DataRecipientStatusAction = components["schemas"]["DataRecipientStatusAction"];
export type DataRecipientBrandStatusAction =
  components["schemas"]["DataRecipientBrandStatusAction"];
export type SoftwareProductStatusAction = components["schemas"]["SoftwareProductStatusAction"];
export type DataRecipientDetailsChange = Omit<
  components["schemas"]["DataRecipientDetailsChangeRequest"],
  "version"
>;
export type DataRecipientBrandChange = Omit<
  components["schemas"]["DataRecipientBrandChangeRequest"],
  "version"
>;
export type SoftwareProductChange = Omit<
  components["schemas"]["SoftwareProductChangeRequest"],
  "version"
>;
export type DataRecipientBrandAdd = Omit<
  components["schemas"]["DataRecipientBrandAddRequest"],
  "version"
>;
export type SoftwareProductAdd = Omit<
  components["schemas"]["SoftwareProductAddRequest"],
  "version"
>;
type DataRecipientCreateRequest = components["schemas"]["DataRecipientCreateRequest"];
// The console does not check the values, so a request can have no accreditation level, which the Register refuses.
export type DataRecipientCreate = Omit<DataRecipientCreateRequest, "accreditationLevel"> & {
  accreditationLevel?: AccreditationLevel | undefined;
};
export type AccreditationLevel = components["schemas"]["AccreditationLevel"];
export type DataHolderDetail = components["schemas"]["DataHolderDetail"];
export type DataHolderStatusAction = components["schemas"]["DataHolderStatusAction"];
export type DataHolderBrandStatusAction = components["schemas"]["DataHolderBrandStatusAction"];
export type DataHolderCreate = components["schemas"]["DataHolderCreateRequest"];
export type DataHolderDetailsChange = Omit<
  components["schemas"]["DataHolderDetailsChangeRequest"],
  "version"
>;
export type DataHolderBrand = components["schemas"]["DataHolderBrand"];
export type DataHolderBrandAdd = Omit<
  components["schemas"]["DataHolderBrandAddRequest"],
  "version"
>;
export type DataHolderBrandChange = Omit<
  components["schemas"]["DataHolderBrandChangeRequest"],
  "version"
>;
export type Industry = components["schemas"]["Industry"];
export type CertificateIssue = components["schemas"]["CertificateIssueResponse"];
export type RevocationRequest = components["schemas"]["RevocationRequest"];
export type RevocationReason = components["schemas"]["RevocationReason"];
export type RevocableCertificate = components["schemas"]["RevocableCertificate"];
export type Revocation = components["schemas"]["Revocation"];
export type RevocationQuery = NonNullable<
  paths["/api/cdr-certificates/revocations"]["get"]["parameters"]["query"]
>;
export type Status = components["schemas"]["Status"];
export type DataRecipientHistoryEntry = components["schemas"]["DataRecipientHistoryEntry"];
export type DataHolderHistoryEntry = components["schemas"]["DataHolderHistoryEntry"];
export type HistoryQuery = NonNullable<
  paths["/api/data-recipients/{legalEntityId}/history"]["get"]["parameters"]["query"]
>;

const api = createClient<paths>();

const readFailure = consoleReadFailure("the Register");

export function problemOf(error: unknown): Problem {
  return problemOfService(error, readFailure);
}

export async function listDataRecipients(query: DataRecipientQuery, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/data-recipients", {
    params: { query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listDataHolders(query: DataHolderQuery, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/data-holders", {
    params: { query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listKeySets(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/key-sets", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getDataRecipient(legalEntityId: string, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/data-recipients/{legalEntityId}", {
    params: { path: { legalEntityId } },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getDataHolder(legalEntityId: string, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/data-holders/{legalEntityId}", {
    params: { path: { legalEntityId } },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getDataRecipientHistory(
  legalEntityId: string,
  query: HistoryQuery,
  signal: AbortSignal,
) {
  const { data, error, response } = await api.GET("/api/data-recipients/{legalEntityId}/history", {
    params: { path: { legalEntityId }, query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function getDataHolderHistory(
  legalEntityId: string,
  query: HistoryQuery,
  signal: AbortSignal,
) {
  const { data, error, response } = await api.GET("/api/data-holders/{legalEntityId}/history", {
    params: { path: { legalEntityId }, query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function changeDataRecipientStatus(
  legalEntityId: string,
  action: DataRecipientStatusAction["action"],
  version: number,
) {
  const { error, response } = await api.POST(`/api/data-recipients/{legalEntityId}/${action}`, {
    params: { path: { legalEntityId } },
    body: { version },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function changeDataRecipientBrandStatus(
  legalEntityId: string,
  brandId: string,
  action: DataRecipientBrandStatusAction["action"],
  version: number,
) {
  const { error, response } = await api.POST(
    `/api/data-recipients/{legalEntityId}/brands/{brandId}/${action}`,
    { params: { path: { legalEntityId, brandId } }, body: { version } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function changeSoftwareProductStatus(
  legalEntityId: string,
  softwareProductId: string,
  action: SoftwareProductStatusAction["action"],
  version: number,
) {
  const { error, response } = await api.POST(
    `/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/${action}`,
    { params: { path: { legalEntityId, softwareProductId } }, body: { version } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function changeDataHolderStatus(
  legalEntityId: string,
  action: DataHolderStatusAction["action"],
  version: number,
) {
  const { error, response } = await api.POST(`/api/data-holders/{legalEntityId}/${action}`, {
    params: { path: { legalEntityId } },
    body: { version },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function changeDataHolderBrandStatus(
  legalEntityId: string,
  brandId: string,
  action: DataHolderBrandStatusAction["action"],
  version: number,
) {
  const { error, response } = await api.POST(
    `/api/data-holders/{legalEntityId}/brands/{brandId}/${action}`,
    { params: { path: { legalEntityId, brandId } }, body: { version } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function registerClientCertificate(
  legalEntityId: string,
  softwareProductId: string,
  version: number,
  certificate: string,
) {
  const { error, response } = await api.POST(
    "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates",
    { params: { path: { legalEntityId, softwareProductId } }, body: { version, certificate } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function issueClientCertificate(
  legalEntityId: string,
  softwareProductId: string,
  version: number,
  csr: string,
) {
  const { data, error, response } = await api.POST(
    "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue",
    { params: { path: { legalEntityId, softwareProductId } }, body: { version, csr } },
  );
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function issueServerCertificate(
  legalEntityId: string,
  brandId: string,
  version: number,
  csr: string,
) {
  const { data, error, response } = await api.POST(
    "/api/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue",
    { params: { path: { legalEntityId, brandId } }, body: { version, csr } },
  );
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function removeClientCertificate(
  legalEntityId: string,
  softwareProductId: string,
  thumbprint: string,
  version: number,
) {
  const { error, response } = await api.POST(
    "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/{thumbprint}/remove",
    { params: { path: { legalEntityId, softwareProductId, thumbprint } }, body: { version } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function changeKey(
  keySet: KeySet["keySet"],
  action: KeyAction["action"],
  version: number,
) {
  const { error, response } = await api.POST(`/api/key-sets/{keySet}/${action}`, {
    params: { path: { keySet } },
    body: { version },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function changeDataRecipientDetails(
  legalEntityId: string,
  version: number,
  details: DataRecipientDetailsChange,
) {
  const { error, response } = await api.POST("/api/data-recipients/{legalEntityId}/change", {
    params: { path: { legalEntityId } },
    body: { version, ...details },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function changeDataRecipientBrandDetails(
  legalEntityId: string,
  brandId: string,
  version: number,
  details: DataRecipientBrandChange,
) {
  const { error, response } = await api.POST(
    "/api/data-recipients/{legalEntityId}/brands/{brandId}/change",
    { params: { path: { legalEntityId, brandId } }, body: { version, ...details } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function changeSoftwareProductDetails(
  legalEntityId: string,
  softwareProductId: string,
  version: number,
  details: SoftwareProductChange,
) {
  const { error, response } = await api.POST(
    "/api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/change",
    { params: { path: { legalEntityId, softwareProductId } }, body: { version, ...details } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function createDataRecipient(create: DataRecipientCreate) {
  const { error, response } = await api.POST("/api/data-recipients", {
    // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the Register checks the value, not the UI.
    body: create as DataRecipientCreateRequest,
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function addDataRecipientBrand(
  legalEntityId: string,
  version: number,
  add: DataRecipientBrandAdd,
) {
  const { error, response } = await api.POST("/api/data-recipients/{legalEntityId}/brands", {
    params: { path: { legalEntityId } },
    body: { version, ...add },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function addSoftwareProduct(
  legalEntityId: string,
  brandId: string,
  version: number,
  add: SoftwareProductAdd,
) {
  const { error, response } = await api.POST(
    "/api/data-recipients/{legalEntityId}/brands/{brandId}/software-products",
    { params: { path: { legalEntityId, brandId } }, body: { version, ...add } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function createDataHolder(create: DataHolderCreate) {
  const { error, response } = await api.POST("/api/data-holders", { body: create });
  if (error) throw new ProblemError(error, response.status);
}

export async function changeDataHolderDetails(
  legalEntityId: string,
  version: number,
  details: DataHolderDetailsChange,
) {
  const { error, response } = await api.POST("/api/data-holders/{legalEntityId}/change", {
    params: { path: { legalEntityId } },
    body: { version, ...details },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function addDataHolderBrand(
  legalEntityId: string,
  version: number,
  add: DataHolderBrandAdd,
) {
  const { error, response } = await api.POST("/api/data-holders/{legalEntityId}/brands", {
    params: { path: { legalEntityId } },
    body: { version, ...add },
  });
  if (error) throw new ProblemError(error, response.status);
}

export async function changeDataHolderBrandDetails(
  legalEntityId: string,
  brandId: string,
  version: number,
  details: DataHolderBrandChange,
) {
  const { error, response } = await api.POST(
    "/api/data-holders/{legalEntityId}/brands/{brandId}/change",
    { params: { path: { legalEntityId, brandId } }, body: { version, ...details } },
  );
  if (error) throw new ProblemError(error, response.status);
}

export async function getCDRCABundle(signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/cdr-ca-bundle", { signal });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function listRevocations(query: RevocationQuery, signal: AbortSignal) {
  const { data, error, response } = await api.GET("/api/cdr-certificates/revocations", {
    params: { query },
    signal,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

// The console does not check the values, so a request can have no reason, which the Register refuses.
export type RevocationValues = { certificate: string; reason?: RevocationReason | undefined };

export async function checkRevocation(values: RevocationValues) {
  const { data, error, response } = await api.POST("/api/cdr-certificates/check-revocation", {
    // oxlint-disable-next-line typescript/no-unsafe-type-assertion -- the Register checks the value, not the UI.
    body: values as RevocationRequest,
  });
  if (error) throw new ProblemError(error, response.status);
  return data;
}

export async function revokeCertificate(request: RevocationRequest) {
  const { error, response } = await api.POST("/api/cdr-certificates/revoke", { body: request });
  if (error) throw new ProblemError(error, response.status);
}
