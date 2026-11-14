import type { ProductCategory } from "../api/client";
import { loanProductCategoryValues, revolvingProductCategoryValues } from "../api/schema.gen";

// The value that the consumer gives at the open of a lending product, with the name of its field in the site API
// (services/dataholder/DESIGN.md, Opening of an account). A deposit product has none.
export type OpenValue = { name: "creditLimit" | "loanAmount"; label: string; description: string };

const creditLimit: OpenValue = {
  name: "creditLimit",
  label: "Credit limit",
  description: "The most that you can owe on this account, in dollars, for example 5,000.",
};
const loanAmount: OpenValue = {
  name: "loanAmount",
  label: "Loan amount",
  description:
    "The bank pays out this amount when the account opens, and you then owe it. In dollars, for example 20,000.",
};

export function openValueOf(category: ProductCategory): OpenValue | undefined {
  if ((revolvingProductCategoryValues as readonly ProductCategory[]).includes(category))
    return creditLimit;
  if ((loanProductCategoryValues as readonly ProductCategory[]).includes(category))
    return loanAmount;
  return undefined;
}

// Dollars with an optional "$", commas between groups of 3 digits, and at most 2 digits of cents.
const dollars = /^\$?\s*(\d{1,3}(?:,\d{3})+|\d+)(?:\.(\d{1,2}))?$/;

// The site API takes a CDS AmountString only, so the page converts the usual forms of an amount. The data holder
// checks the rules of the value.
export function amountStringOf(text: string): string | undefined {
  const match = dollars.exec(text.trim());
  if (!match?.[1]) return undefined;
  const whole = match[1].replaceAll(",", "").replace(/^0+(?=\d)/, "");
  const cents = (match[2] ?? "").padEnd(2, "0");
  return `${whole}.${cents}`;
}
