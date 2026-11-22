import { randomInt } from "node:crypto";

export function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`Set ${name}: mise run test:lab-web sets it.`);
  return value;
}

export type Customer = {
  customerNumber: string;
  firstName?: string;
  lastName: string;
  emailAddress: string;
};

// The operator interface of the data holder of the console under test, or of a different instance.
async function operatorPost(
  path: string,
  body: unknown,
  status = 201,
  operatorURL = requiredEnv("DATAHOLDER_INTERNAL_URL"),
): Promise<unknown> {
  const response = await fetch(new URL(path, operatorURL), {
    method: "POST",
    headers: {
      Authorization: `Bearer ${requiredEnv("DATAHOLDER_OPERATOR_TOKEN")}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (response.status !== status) {
    throw new Error(`The operator interface refused the record: ${await response.text()}`);
  }
  return response.json();
}

function stringField(body: unknown, name: string): string {
  const value: unknown = typeof body === "object" && body !== null && Reflect.get(body, name);
  if (typeof value !== "string") throw new Error(`The operator interface gave no ${name}.`);
  return value;
}

// The lab has no seed customers, so each test creates its own with the operator interface. All data is invented.
export async function createCustomer(
  customer: Omit<Customer, "customerNumber">,
  operatorURL?: string,
): Promise<Customer> {
  const body = await operatorPost("/v1/customers", customer, 201, operatorURL);
  return { ...customer, customerNumber: stringField(body, "customerNumber") };
}

export type Account = {
  displayName: string;
  productCategory: string;
  productName: string;
  bsb: string;
  accountNumber: string;
  creditLimit?: string;
};

// Gives the accountId.
export async function createAccount(
  customerNumber: string,
  account: Account,
  operatorURL?: string,
): Promise<string> {
  const body = await operatorPost(
    `/v1/customers/${customerNumber}/accounts`,
    account,
    201,
    operatorURL,
  );
  return stringField(body, "accountId");
}

export type Transaction = {
  type: string;
  status: "POSTED" | "PENDING";
  description: string;
  amount: string;
  reference: string;
  postingDateTime?: string;
  executionDateTime?: string;
};

export async function createTransaction(
  customerNumber: string,
  accountId: string,
  transaction: Transaction,
): Promise<void> {
  await operatorPost(
    `/v1/customers/${customerNumber}/accounts/${accountId}/transactions`,
    transaction,
  );
}

export async function closeAccount(customerNumber: string, accountId: string): Promise<void> {
  await operatorPost(`/v1/customers/${customerNumber}/accounts/${accountId}/close`, undefined, 200);
}

// An account number that no other test run has, with the 8 digits of an account number of a bank.
export function uniqueAccountNumber(): string {
  return String(randomInt(10_000_000, 100_000_000));
}

// A name that no other test run has, with letters only, as in a name of a person.
export function uniqueName(): string {
  const letters = "abcdefghijklmnopqrstuvwxyz";
  return Array.from({ length: 10 }, () => letters[randomInt(letters.length)]).join("");
}
