import type { components } from "../api/schema.gen";

type Schemas = components["schemas"];

// A label in plain language for each value of each enumeration of a product, from the meaning that the CDS gives
// (CDS: Product & Account Components). The Record types make the build fail when a value has no label.

export const featureLabels: Record<Schemas["FeatureType"], string> = {
  ADDITIONAL_CARDS: "Extra cards",
  BALANCE_TRANSFERS: "Balance transfers",
  BILL_PAYMENT: "Bill payment service",
  BONUS_REWARDS: "Bonus reward points",
  CARD_ACCESS: "Card",
  CASHBACK_OFFER: "Cashback offer",
  COMPLEMENTARY_PRODUCT_DISCOUNTS: "Discounts on other products",
  EXTRA_DOWN_PAYMENT: "Larger first payment",
  DIGITAL_BANKING: "Online and app banking",
  DIGITAL_WALLET: "Digital wallet",
  DONATE_INTEREST: "Give your interest to a charity",
  EXTRA_REPAYMENTS: "Extra repayments at no cost",
  FRAUD_PROTECTION: "Fraud protection",
  FREE_TXNS: "Free transactions each month",
  FREE_TXNS_ALLOWANCE: "Free transaction fees each month",
  FUNDS_AVAILABLE_AFTER: "Deposits available after a time",
  GUARANTOR: "A guarantor can support your application",
  INSTALMENT_PLAN: "Pay for purchases in instalments",
  INSURANCE: "Insurance",
  INTEREST_FREE: "Interest-free period on purchases",
  INTEREST_FREE_TRANSFERS: "Interest-free period on balance transfers",
  LOYALTY_PROGRAM: "Loyalty points program",
  NOTIFICATIONS: "Alerts",
  NPP_ENABLED: "Fast payments to your BSB and account number",
  NPP_PAYID: "Receive payments to a PayID",
  OFFSET: "Offset account",
  OTHER: "Other feature",
  OVERDRAFT: "Overdraft available",
  REDRAW: "Redraw extra repayments",
  RELATIONSHIP_MANAGEMENT: "Relationship manager",
  UNLIMITED_TXNS: "Unlimited free transactions",
};

export const feeLabels: Record<Schemas["FeeType"], string> = {
  CASH_ADVANCE: "Cash advance fee",
  DEPOSIT: "Deposit fee",
  DISHONOUR: "Dishonour fee",
  ENQUIRY: "Enquiry fee",
  EVENT: "Service fee",
  EXIT: "Account closing fee",
  LATE_PAYMENT: "Late payment fee",
  OTHER: "Other fee",
  PAYMENT: "Payment fee",
  PERIODIC: "Regular fee",
  PURCHASE: "Purchase fee",
  REPLACEMENT: "Replacement fee",
  TRANSACTION: "Transaction fee",
  UPFRONT: "Set-up fee",
  UPFRONT_PER_PLAN: "Fee for each instalment plan",
  VARIATION: "Change fee",
  WITHDRAWAL: "Withdrawal fee",
};

export const feeMethodLabels: Record<Schemas["FeeMethodUType"], string> = {
  fixedAmount: "A fixed amount",
  rateBased: "A percentage",
  variable: "An amount in a range",
};

export const feeRateLabels: Record<Schemas["FeeRateType"], string> = {
  BALANCE: "the balance",
  INTEREST_ACCRUED: "the interest",
  TRANSACTION: "the transaction amount",
};

export const depositRateLabels: Record<Schemas["DepositRateType"], string> = {
  BONUS: "Bonus rate",
  BUNDLE_BONUS: "Bonus rate with a package",
  FIXED: "Fixed rate",
  FLOATING: "Floating rate",
  INTRODUCTORY: "Introductory bonus rate",
  MARKET_LINKED: "Market-linked rate",
  VARIABLE: "Variable rate",
};

export const lendingRateLabels: Record<Schemas["LendingRateType"], string> = {
  BALANCE_TRANSFER: "Balance transfer rate",
  BUNDLE_DISCOUNT_FIXED: "Package discount on the fixed rate",
  BUNDLE_DISCOUNT_VARIABLE: "Package discount on the variable rate",
  CASH_ADVANCE: "Cash advance rate",
  DISCOUNT: "Discount",
  FIXED: "Fixed rate",
  FLOATING: "Floating rate",
  INTRODUCTORY: "Introductory discount",
  MARKET_LINKED: "Market-linked rate",
  PENALTY: "Penalty rate",
  PURCHASE: "Purchase rate",
  VARIABLE: "Variable rate",
};

export const applicationLabels: Record<Schemas["ApplicationType"], string> = {
  MATURITY: "At the end of the term",
  PERIODIC: "Regularly",
  UPFRONT: "At the start",
};

export const repaymentLabels: Record<Schemas["RepaymentType"], string> = {
  INTEREST_ONLY: "Interest only",
  OTHER: "Other",
  PRINCIPAL_AND_INTEREST: "Principal and interest",
  UNCONSTRAINED: "Any",
};

export const loanPurposeLabels: Record<Schemas["LoanPurpose"], string> = {
  INVESTMENT: "Investment",
  OTHER: "Other",
  OWNER_OCCUPIED: "A home you live in",
  UNCONSTRAINED: "Any",
};
