import type { components } from "../api/schema.gen";

type Schemas = components["schemas"];

// The formats of CDS values for a consumer (CDS: Common Field Types), in Australian English (packages/ui/DESIGN.md,
// Content). A value that is not in its format shows as it is, so that the page never hides data.

const locale = "en-AU";

const percent = new Intl.NumberFormat(locale, {
  style: "percent",
  minimumFractionDigits: 2,
  maximumFractionDigits: 4,
});
const currency = new Intl.NumberFormat(locale, { style: "currency", currency: "AUD" });
const integer = new Intl.NumberFormat(locale);
const date = new Intl.DateTimeFormat(locale, { day: "numeric", month: "long", year: "numeric" });
const list = new Intl.ListFormat(locale, { type: "conjunction" });

const decimal = /^-?\d+(\.\d+)?$/;
const count = /^\d+$/;
// ISO 8601 durations with no time part, for example P1Y6M or P55D.
const duration = /^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?$/;
const units = ["year", "month", "week", "day"] as const;

export function formatRate(rate: string): string {
  return decimal.test(rate) ? percent.format(Number(rate)) : rate;
}

export function formatAmount(amount: string): string {
  return decimal.test(amount) ? currency.format(Number(amount)) : amount;
}

// The date of a date-time, in the time zone of the browser.
export function formatDate(dateTime: string): string {
  return date.format(new Date(dateTime));
}

function durationParts(value: string): string[] | undefined {
  const match = duration.exec(value);
  if (!match) return undefined;
  const parts = units.flatMap((unit, index) => {
    const amount = match[index + 1];
    if (amount === undefined) return [];
    return [`${amount} ${unit}${amount === "1" ? "" : "s"}`];
  });
  return parts.length > 0 ? parts : undefined;
}

export function formatDuration(value: string): string {
  const parts = durationParts(value);
  return parts ? list.format(parts) : value;
}

// The frequency of an applicationFrequency or of a periodic fee, for example "Each month" for P1M.
export function formatFrequency(value: string): string {
  const parts = durationParts(value);
  if (!parts) return value;
  const [only] = parts;
  if (parts.length === 1 && only?.startsWith("1 ")) return `Each ${only.slice(2)}`;
  return `Every ${list.format(parts)}`;
}

type Format = "duration" | "frequency" | "amount" | "count" | "days" | "text";

const terms: Record<Format, string> = {
  duration: "Period",
  frequency: "Charged",
  amount: "Amount",
  count: "Number",
  days: "Days late",
  text: "Details",
};

const asText = (value: string) => value;

const formatters: Record<Format, (value: string) => string> = {
  duration: formatDuration,
  frequency: formatFrequency,
  amount: formatAmount,
  count: (value) => (count.test(value) ? integer.format(Number(value)) : value),
  days: asText,
  text: asText,
};

// The format of the additionalValue of each type that the CDS gives a format for. Each other type is text.
const formats = {
  feature: {
    ADDITIONAL_CARDS: "count",
    BONUS_REWARDS: "count",
    CASHBACK_OFFER: "amount",
    FREE_TXNS: "count",
    FREE_TXNS_ALLOWANCE: "amount",
    FUNDS_AVAILABLE_AFTER: "duration",
    INTEREST_FREE: "duration",
    INTEREST_FREE_TRANSFERS: "duration",
  },
  fee: { LATE_PAYMENT: "days", PERIODIC: "frequency" },
  depositRate: { FIXED: "duration", INTRODUCTORY: "duration" },
  lendingRate: { FIXED: "duration", INTRODUCTORY: "duration" },
} satisfies {
  feature: Partial<Record<Schemas["FeatureType"], Format>>;
  fee: Partial<Record<Schemas["FeeType"], Format>>;
  depositRate: Partial<Record<Schemas["DepositRateType"], Format>>;
  lendingRate: Partial<Record<Schemas["LendingRateType"], Format>>;
};

type Types = {
  feature: Schemas["FeatureType"];
  fee: Schemas["FeeType"];
  depositRate: Schemas["DepositRateType"];
  lendingRate: Schemas["LendingRateType"];
};

export function additionalValue<Kind extends keyof Types>(
  kind: Kind,
  type: Types[Kind],
  value: string,
): { term: string; text: string } {
  const byType: Partial<Record<string, Format>> = formats[kind];
  const format = byType[type] ?? "text";
  return { term: terms[format], text: formatters[format](value) };
}
