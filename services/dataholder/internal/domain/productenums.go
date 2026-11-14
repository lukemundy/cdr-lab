package domain

import "slices"

// The enumerations of the CDS banking product schemas, with all their values.

type ProductCategory string

const (
	CategoryBusinessLoans  ProductCategory = "BUSINESS_LOANS"
	CategoryBuyNowPayLater ProductCategory = "BUY_NOW_PAY_LATER"
	//nolint:gosec // G101: "Cred" is from the CDS name for credit and charge cards, not a credential.
	CategoryCredAndChrgCards        ProductCategory = "CRED_AND_CHRG_CARDS"
	CategoryLeases                  ProductCategory = "LEASES"
	CategoryMarginLoans             ProductCategory = "MARGIN_LOANS"
	CategoryOverdrafts              ProductCategory = "OVERDRAFTS"
	CategoryPersLoans               ProductCategory = "PERS_LOANS"
	CategoryRegulatedTrustAccounts  ProductCategory = "REGULATED_TRUST_ACCOUNTS"
	CategoryResidentialMortgages    ProductCategory = "RESIDENTIAL_MORTGAGES"
	CategoryTermDeposits            ProductCategory = "TERM_DEPOSITS"
	CategoryTradeFinance            ProductCategory = "TRADE_FINANCE"
	CategoryTransAndSavingsAccounts ProductCategory = "TRANS_AND_SAVINGS_ACCOUNTS"
	CategoryTravelCards             ProductCategory = "TRAVEL_CARDS"
)

var productCategories = []ProductCategory{
	CategoryBusinessLoans,
	CategoryBuyNowPayLater,
	CategoryCredAndChrgCards,
	CategoryLeases,
	CategoryMarginLoans,
	CategoryOverdrafts,
	CategoryPersLoans,
	CategoryRegulatedTrustAccounts,
	CategoryResidentialMortgages,
	CategoryTermDeposits,
	CategoryTradeFinance,
	CategoryTransAndSavingsAccounts,
	CategoryTravelCards,
}

// depositCategories are the "Deposit Products" of the CDS Product Categories. Each other category is a lending product:
// revolving credit with a credit limit, or a loan.
var depositCategories = []ProductCategory{
	CategoryRegulatedTrustAccounts, CategoryTermDeposits, CategoryTransAndSavingsAccounts, CategoryTravelCards,
}

var revolvingCategories = []ProductCategory{
	CategoryBuyNowPayLater, CategoryCredAndChrgCards, CategoryMarginLoans, CategoryOverdrafts,
}

// DepositCategories gives a copy.
func DepositCategories() []ProductCategory {
	return slices.Clone(depositCategories)
}

// RevolvingCategories gives a copy.
func RevolvingCategories() []ProductCategory {
	return slices.Clone(revolvingCategories)
}

// LoanCategories gives the lending categories that are not revolving credit.
func LoanCategories() []ProductCategory {
	return slices.DeleteFunc(slices.Clone(productCategories), func(c ProductCategory) bool {
		return c.IsDeposit() || c.IsRevolving()
	})
}

func (c ProductCategory) IsDeposit() bool {
	return slices.Contains(depositCategories, c)
}

func (c ProductCategory) IsRevolving() bool {
	return slices.Contains(revolvingCategories, c)
}

func (c ProductCategory) IsLoan() bool {
	return slices.Contains(LoanCategories(), c)
}

func ParseProductCategory(s string) (ProductCategory, bool) {
	c := ProductCategory(s)
	return c, slices.Contains(productCategories, c)
}

type FeatureType string

const (
	FeatureAdditionalCards               FeatureType = "ADDITIONAL_CARDS"
	FeatureBalanceTransfers              FeatureType = "BALANCE_TRANSFERS"
	FeatureBillPayment                   FeatureType = "BILL_PAYMENT"
	FeatureBonusRewards                  FeatureType = "BONUS_REWARDS"
	FeatureCardAccess                    FeatureType = "CARD_ACCESS"
	FeatureCashbackOffer                 FeatureType = "CASHBACK_OFFER"
	FeatureComplementaryProductDiscounts FeatureType = "COMPLEMENTARY_PRODUCT_DISCOUNTS"
	FeatureExtraDownPayment              FeatureType = "EXTRA_DOWN_PAYMENT"
	FeatureDigitalBanking                FeatureType = "DIGITAL_BANKING"
	FeatureDigitalWallet                 FeatureType = "DIGITAL_WALLET"
	FeatureDonateInterest                FeatureType = "DONATE_INTEREST"
	FeatureExtraRepayments               FeatureType = "EXTRA_REPAYMENTS"
	FeatureFraudProtection               FeatureType = "FRAUD_PROTECTION"
	FeatureFreeTxns                      FeatureType = "FREE_TXNS"
	FeatureFreeTxnsAllowance             FeatureType = "FREE_TXNS_ALLOWANCE"
	FeatureFundsAvailableAfter           FeatureType = "FUNDS_AVAILABLE_AFTER"
	FeatureGuarantor                     FeatureType = "GUARANTOR"
	FeatureInstalmentPlan                FeatureType = "INSTALMENT_PLAN"
	FeatureInsurance                     FeatureType = "INSURANCE"
	FeatureInterestFree                  FeatureType = "INTEREST_FREE"
	FeatureInterestFreeTransfers         FeatureType = "INTEREST_FREE_TRANSFERS"
	FeatureLoyaltyProgram                FeatureType = "LOYALTY_PROGRAM"
	FeatureNotifications                 FeatureType = "NOTIFICATIONS"
	FeatureNppEnabled                    FeatureType = "NPP_ENABLED"
	FeatureNppPayid                      FeatureType = "NPP_PAYID"
	FeatureOffset                        FeatureType = "OFFSET"
	FeatureOther                         FeatureType = "OTHER"
	FeatureOverdraft                     FeatureType = "OVERDRAFT"
	FeatureRedraw                        FeatureType = "REDRAW"
	FeatureRelationshipManagement        FeatureType = "RELATIONSHIP_MANAGEMENT"
	FeatureUnlimitedTxns                 FeatureType = "UNLIMITED_TXNS"
)

var featureTypes = []FeatureType{
	FeatureAdditionalCards,
	FeatureBalanceTransfers,
	FeatureBillPayment,
	FeatureBonusRewards,
	FeatureCardAccess,
	FeatureCashbackOffer,
	FeatureComplementaryProductDiscounts,
	FeatureExtraDownPayment,
	FeatureDigitalBanking,
	FeatureDigitalWallet,
	FeatureDonateInterest,
	FeatureExtraRepayments,
	FeatureFraudProtection,
	FeatureFreeTxns,
	FeatureFreeTxnsAllowance,
	FeatureFundsAvailableAfter,
	FeatureGuarantor,
	FeatureInstalmentPlan,
	FeatureInsurance,
	FeatureInterestFree,
	FeatureInterestFreeTransfers,
	FeatureLoyaltyProgram,
	FeatureNotifications,
	FeatureNppEnabled,
	FeatureNppPayid,
	FeatureOffset,
	FeatureOther,
	FeatureOverdraft,
	FeatureRedraw,
	FeatureRelationshipManagement,
	FeatureUnlimitedTxns,
}

type FeeType string

const (
	FeeTypeCashAdvance    FeeType = "CASH_ADVANCE"
	FeeTypeDeposit        FeeType = "DEPOSIT"
	FeeTypeDishonour      FeeType = "DISHONOUR"
	FeeTypeEnquiry        FeeType = "ENQUIRY"
	FeeTypeEvent          FeeType = "EVENT"
	FeeTypeExit           FeeType = "EXIT"
	FeeTypeLatePayment    FeeType = "LATE_PAYMENT"
	FeeTypeOther          FeeType = "OTHER"
	FeeTypePayment        FeeType = "PAYMENT"
	FeeTypePeriodic       FeeType = "PERIODIC"
	FeeTypePurchase       FeeType = "PURCHASE"
	FeeTypeReplacement    FeeType = "REPLACEMENT"
	FeeTypeTransaction    FeeType = "TRANSACTION"
	FeeTypeUpfront        FeeType = "UPFRONT"
	FeeTypeUpfrontPerPlan FeeType = "UPFRONT_PER_PLAN"
	FeeTypeVariation      FeeType = "VARIATION"
	FeeTypeWithdrawal     FeeType = "WITHDRAWAL"
)

var feeTypes = []FeeType{
	FeeTypeCashAdvance,
	FeeTypeDeposit,
	FeeTypeDishonour,
	FeeTypeEnquiry,
	FeeTypeEvent,
	FeeTypeExit,
	FeeTypeLatePayment,
	FeeTypeOther,
	FeeTypePayment,
	FeeTypePeriodic,
	FeeTypePurchase,
	FeeTypeReplacement,
	FeeTypeTransaction,
	FeeTypeUpfront,
	FeeTypeUpfrontPerPlan,
	FeeTypeVariation,
	FeeTypeWithdrawal,
}

type FeeMethodUType string

const (
	FeeFixedAmount FeeMethodUType = "fixedAmount"
	FeeRateBased   FeeMethodUType = "rateBased"
	FeeVariable    FeeMethodUType = "variable"
)

var feeMethodUTypes = []FeeMethodUType{
	FeeFixedAmount, FeeRateBased, FeeVariable,
}

type FeeRateType string

const (
	FeeRateBalance         FeeRateType = "BALANCE"
	FeeRateInterestAccrued FeeRateType = "INTEREST_ACCRUED"
	FeeRateTransaction     FeeRateType = "TRANSACTION"
)

var feeRateTypes = []FeeRateType{
	FeeRateBalance, FeeRateInterestAccrued, FeeRateTransaction,
}

type DepositRateType string

const (
	DepositRateBonus        DepositRateType = "BONUS"
	DepositRateBundleBonus  DepositRateType = "BUNDLE_BONUS"
	DepositRateFixed        DepositRateType = "FIXED"
	DepositRateFloating     DepositRateType = "FLOATING"
	DepositRateIntroductory DepositRateType = "INTRODUCTORY"
	DepositRateMarketLinked DepositRateType = "MARKET_LINKED"
	DepositRateVariable     DepositRateType = "VARIABLE"
)

var depositRateTypes = []DepositRateType{
	DepositRateBonus,
	DepositRateBundleBonus,
	DepositRateFixed,
	DepositRateFloating,
	DepositRateIntroductory,
	DepositRateMarketLinked,
	DepositRateVariable,
}

type LendingRateType string

const (
	LendingRateBalanceTransfer        LendingRateType = "BALANCE_TRANSFER"
	LendingRateBundleDiscountFixed    LendingRateType = "BUNDLE_DISCOUNT_FIXED"
	LendingRateBundleDiscountVariable LendingRateType = "BUNDLE_DISCOUNT_VARIABLE"
	LendingRateCashAdvance            LendingRateType = "CASH_ADVANCE"
	LendingRateDiscount               LendingRateType = "DISCOUNT"
	LendingRateFixed                  LendingRateType = "FIXED"
	LendingRateFloating               LendingRateType = "FLOATING"
	LendingRateIntroductory           LendingRateType = "INTRODUCTORY"
	LendingRateMarketLinked           LendingRateType = "MARKET_LINKED"
	LendingRatePenalty                LendingRateType = "PENALTY"
	LendingRatePurchase               LendingRateType = "PURCHASE"
	LendingRateVariable               LendingRateType = "VARIABLE"
)

var lendingRateTypes = []LendingRateType{
	LendingRateBalanceTransfer,
	LendingRateBundleDiscountFixed,
	LendingRateBundleDiscountVariable,
	LendingRateCashAdvance,
	LendingRateDiscount,
	LendingRateFixed,
	LendingRateFloating,
	LendingRateIntroductory,
	LendingRateMarketLinked,
	LendingRatePenalty,
	LendingRatePurchase,
	LendingRateVariable,
}

type ApplicationType string

const (
	ApplicationMaturity ApplicationType = "MATURITY"
	ApplicationPeriodic ApplicationType = "PERIODIC"
	ApplicationUpfront  ApplicationType = "UPFRONT"
)

var applicationTypes = []ApplicationType{
	ApplicationMaturity, ApplicationPeriodic, ApplicationUpfront,
}

type RepaymentType string

const (
	RepaymentInterestOnly         RepaymentType = "INTEREST_ONLY"
	RepaymentOther                RepaymentType = "OTHER"
	RepaymentPrincipalAndInterest RepaymentType = "PRINCIPAL_AND_INTEREST"
	RepaymentUnconstrained        RepaymentType = "UNCONSTRAINED"
)

var repaymentTypes = []RepaymentType{
	RepaymentInterestOnly, RepaymentOther, RepaymentPrincipalAndInterest, RepaymentUnconstrained,
}

type LoanPurpose string

const (
	PurposeInvestment    LoanPurpose = "INVESTMENT"
	PurposeOther         LoanPurpose = "OTHER"
	PurposeOwnerOccupied LoanPurpose = "OWNER_OCCUPIED"
	PurposeUnconstrained LoanPurpose = "UNCONSTRAINED"
)

var loanPurposes = []LoanPurpose{
	PurposeInvestment, PurposeOther, PurposeOwnerOccupied, PurposeUnconstrained,
}

// ProductEnumerations gives the values of each enumeration of a product, by the name of its type.
func ProductEnumerations() map[string][]string {
	return map[string][]string{
		"ProductCategory": values(productCategories),
		"FeatureType":     values(featureTypes),
		"FeeType":         values(feeTypes),
		"FeeMethodUType":  values(feeMethodUTypes),
		"FeeRateType":     values(feeRateTypes),
		"DepositRateType": values(depositRateTypes),
		"LendingRateType": values(lendingRateTypes),
		"ApplicationType": values(applicationTypes),
		"RepaymentType":   values(repaymentTypes),
		"LoanPurpose":     values(loanPurposes),
	}
}

func values[T ~string](list []T) []string {
	strs := make([]string, 0, len(list))
	for _, v := range list {
		strs = append(strs, string(v))
	}

	return strs
}
