// A string value keeps each decimal digit (ECMA-402, Intl.NumberFormat v3), so the UI does not convert an amount to a
// number.
const amountFormat = new Intl.NumberFormat("en-AU", {
  style: "currency",
  currency: "AUD",
  signDisplay: "negative",
});

function isAmountString(amount: string): amount is `${number}` {
  return /^-?[0-9]+\.[0-9]{2}$/.test(amount);
}

// A value that is not an AmountString shows as it is, so that the operator sees what the API gave.
export function formatAmount(amount: string): string {
  return isAmountString(amount) ? amountFormat.format(amount) : amount;
}

// A date of the API is a date in UTC, not an instant, so it is formatted in UTC.
const dateFormat = new Intl.DateTimeFormat("en-AU", {
  day: "numeric",
  month: "long",
  year: "numeric",
  timeZone: "UTC",
});

export function formatDate(date: string): string {
  return dateFormat.format(new Date(`${date}T00:00:00Z`));
}

// A value that is not 6 digits shows as it is, so that the operator sees what the API gave.
export function formatBsb(bsb: string): string {
  return /^[0-9]{6}$/.test(bsb) ? `${bsb.slice(0, 3)}-${bsb.slice(3)}` : bsb;
}
