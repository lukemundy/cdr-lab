import type { components } from "../api/schema.gen";

type Schemas = components["schemas"];

// A label in plain language for each value of each enumeration of an account and a transaction, from the meaning that
// the CDS gives (CDS: BankingAccountV3, BankingTransactionV2). The Record types make the build fail when a value has no
// label.

export const openStatusLabels: Record<Schemas["OpenStatus"], string> = {
  OPEN: "Open",
  CLOSED: "Closed",
};

export const transactionTypeLabels: Record<Schemas["TransactionType"], string> = {
  DIRECT_DEBIT: "Direct debit",
  FEE: "Fee",
  INTEREST_CHARGED: "Interest charged",
  INTEREST_PAID: "Interest paid",
  OTHER: "Other",
  PAYMENT: "Payment",
  TRANSFER_INCOMING: "Money in",
  TRANSFER_OUTGOING: "Money out",
};

export const transactionStatusLabels: Record<Schemas["TransactionStatus"], string> = {
  PENDING: "Pending",
  POSTED: "Completed",
};
