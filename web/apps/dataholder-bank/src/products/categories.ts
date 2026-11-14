import type { Product, ProductCategory } from "../api/client";
import { productCategoryValues } from "../api/client";

// A label in plain language for each category, from the meaning that the CDS gives (CDS: Product Categories). The
// Record type makes the build fail when a category has no label.
export const categoryLabels: Record<ProductCategory, string> = {
  REGULATED_TRUST_ACCOUNTS: "Trust accounts for regulated businesses",
  TERM_DEPOSITS: "Term deposits",
  TRANS_AND_SAVINGS_ACCOUNTS: "Everyday and savings accounts",
  TRAVEL_CARDS: "Travel money cards",
  BUSINESS_LOANS: "Business loans",
  BUY_NOW_PAY_LATER: "Buy now, pay later",
  CRED_AND_CHRG_CARDS: "Credit and charge cards",
  LEASES: "Leases",
  MARGIN_LOANS: "Margin loans",
  OVERDRAFTS: "Overdrafts",
  PERS_LOANS: "Personal and car loans",
  RESIDENTIAL_MORTGAGES: "Home loans",
  TRADE_FINANCE: "Trade finance",
};

// The ID of the group of a category in the product list, for example "trans-and-savings-accounts".
export function categoryAnchor(category: ProductCategory): string {
  return category.toLowerCase().replaceAll("_", "-");
}

// The file name of the photo of a category in the brand directory, for example "category-term-deposits.webp".
export function categoryPhoto(category: ProductCategory): string {
  return `category-${categoryAnchor(category)}.webp`;
}

export type CategoryGroup = { category: ProductCategory; products: Product[] };

const byName = new Intl.Collator("en-AU");

// The groups are in the order of the categories in the CDS, which the site API document gives, and only a category
// with a product has a group. In a group, the products are in the order of their names.
export function groupByCategory(products: readonly Product[]): CategoryGroup[] {
  return productCategoryValues.flatMap((category) => {
    const inCategory = products
      .filter((product) => product.productCategory === category)
      .toSorted((a, b) => byName.compare(a.name, b.name));
    return inCategory.length > 0 ? [{ category, products: inCategory }] : [];
  });
}

export function productCount(count: number): string {
  return count === 1 ? "1 product" : `${count} products`;
}
