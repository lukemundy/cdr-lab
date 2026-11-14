import type { Transaction } from "../api/client";

// The formats of the account data for a consumer, in Australian English (packages/ui/DESIGN.md, Content). A value that
// is not in its format shows as it is, so that the page never hides data.

const locale = "en-AU";

// A string value keeps each decimal digit (ECMA-402, Intl.NumberFormat v3), so the page does not convert an amount to
// a number.
const signed = new Intl.NumberFormat(locale, {
  style: "currency",
  currency: "AUD",
  signDisplay: "exceptZero",
});
// A CDS DateString is a date, not an instant, so it is formatted in UTC.
const day = new Intl.DateTimeFormat(locale, {
  day: "numeric",
  month: "long",
  year: "numeric",
  timeZone: "UTC",
});

function isAmountString(amount: string): amount is `${number}` {
  return /^-?\d+\.\d{2}$/.test(amount);
}

export function formatSignedAmount(amount: string): string {
  return isAmountString(amount) ? signed.format(amount) : amount;
}

export function formatDay(date: string): string {
  return /^\d{4}-\d{2}-\d{2}$/.test(date) ? day.format(new Date(`${date}T00:00:00Z`)) : date;
}

// The effective time of the CDS: the posting time of a posted transaction, and the execution time of a pending one.
export function effectiveTime(transaction: Transaction): string | undefined {
  return transaction.postingDateTime ?? transaction.executionDateTime;
}
