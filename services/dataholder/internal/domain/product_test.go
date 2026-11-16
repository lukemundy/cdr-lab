package domain_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// product gives a product that obeys all rules, with extra fields from fields.
func product(fields string) []byte {
	if fields != "" {
		fields = ", " + fields
	}

	return fmt.Appendf(nil, `{
		"productId": "bonus-saver",
		"lastUpdated": "2026-08-03T11:00:00+10:00",
		"productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
		"name": "Bonus Saver",
		"description": "A savings account.",
		"brand": "Mock Bank",
		"isTailored": false%s
	}`, fields)
}

func TestParseProductGivesTheFieldsOfTheProduct(t *testing.T) {
	p, err := domain.ParseProduct(product(`
		"effectiveFrom": "2026-01-01T00:00:00Z",
		"effectiveTo": "2027-01-01T00:00:00Z",
		"features": [{"featureType": "NPP_PAYID"}, {"featureType": "OTHER", "additionalInfo": "Name each account"}],
		"fees": [
			{"name": "Monthly fee", "feeType": "PERIODIC", "feeMethodUType": "fixedAmount",
			 "fixedAmount": {"amount": "5.00"}, "additionalValue": "P1M"},
			{"name": "Foreign fee", "feeType": "TRANSACTION", "feeMethodUType": "rateBased",
			 "rateBased": {"rateType": "TRANSACTION", "rate": "0.03"}},
			{"name": "Break cost", "feeType": "EXIT", "feeMethodUType": "variable",
			 "variable": {"feeMaximum": "300.00"}}
		],
		"depositRates": [{"depositRateType": "VARIABLE", "rate": "0.0155", "applicationType": "PERIODIC",
			"applicationFrequency": "P1M"}],
		"lendingRates": [{"lendingRateType": "FIXED", "rate": "0.0799", "comparisonRate": "0.092",
			"applicationType": "MATURITY", "repaymentType": "PRINCIPAL_AND_INTEREST", "loanPurpose": "UNCONSTRAINED",
			"additionalValue": "P5Y"}]
	`))

	require.NoError(t, err)
	assert.Equal(t, "bonus-saver", p.ProductID)
	assert.Equal(t, domain.CategoryTransAndSavingsAccounts, p.ProductCategory)
	assert.Equal(t, "Bonus Saver", p.Name)
	assert.Equal(t, "Mock Bank", p.Brand)
	assert.True(t, p.LastUpdated.Equal(time.Date(2026, 8, 3, 1, 0, 0, 0, time.UTC)))
	require.NotNil(t, p.EffectiveFrom)
	assert.True(t, p.EffectiveFrom.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	require.NotNil(t, p.EffectiveTo)
	assert.Equal(t, []domain.Feature{
		{FeatureType: domain.FeatureNppPayid},
		{FeatureType: domain.FeatureOther, AdditionalInfo: "Name each account"},
	}, p.Features)
	require.Len(t, p.Fees, 3)
	assert.Equal(t, &domain.FeeAmount{Amount: "5.00"}, p.Fees[0].FixedAmount)
	assert.Equal(t, &domain.FeeRate{RateType: domain.FeeRateTransaction, Rate: "0.03"}, p.Fees[1].RateBased)
	assert.Equal(t, &domain.FeeRange{FeeMaximum: "300.00"}, p.Fees[2].Variable)
	require.Len(t, p.DepositRates, 1)
	assert.Equal(t, "0.0155", p.DepositRates[0].Rate)
	require.Len(t, p.LendingRates, 1)
	assert.Equal(t, domain.LendingRateFixed, p.LendingRates[0].LendingRateType)
	assert.Equal(t, "0.092", p.LendingRates[0].ComparisonRate)
}

func TestParseProductGivesNoEffectiveDatesWhenTheProductHasNone(t *testing.T) {
	p, err := domain.ParseProduct(product(""))

	require.NoError(t, err)
	assert.Nil(t, p.EffectiveFrom)
	assert.Nil(t, p.EffectiveTo)
}

// Verifies: DH-0209
func TestParseProductAcceptsEachProductCategory(t *testing.T) {
	for _, category := range []string{
		"BUSINESS_LOANS", "BUY_NOW_PAY_LATER", "CRED_AND_CHRG_CARDS", "LEASES", "MARGIN_LOANS", "OVERDRAFTS",
		"PERS_LOANS", "REGULATED_TRUST_ACCOUNTS", "RESIDENTIAL_MORTGAGES", "TERM_DEPOSITS", "TRADE_FINANCE",
		"TRANS_AND_SAVINGS_ACCOUNTS", "TRAVEL_CARDS",
	} {
		t.Run(category, func(t *testing.T) {
			_, err := domain.ParseProduct(fmt.Appendf(nil, `{"productId": "a", "lastUpdated": "2026-08-03T01:00:00Z",
				"productCategory": %q, "name": "n", "description": "d", "brand": "b", "isTailored": true}`, category))

			assert.NoError(t, err)
		})
	}
}

// Verifies: DH-0231
func TestParseProductRefusesAFieldThatTheDataHolderDoesNotHold(t *testing.T) {
	tests := map[string]string{
		"field that the CDS does not have":    `"productName": "Bonus Saver"`,
		"CDS field that is not supported yet": `"constraints": [{"constraintType": "MIN_BALANCE"}]`,
		"field of a component":                `"fees": [{"name": "n", "feeType": "EXIT", "feeCapp": "1.00"}]`,
	}
	for name, fields := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.ParseProduct(product(fields))

			assert.ErrorContains(t, err, "unknown field")
		})
	}
}

func TestParseProductRefusesAValueOfTheWrongType(t *testing.T) {
	tests := map[string]string{
		"text for a boolean":                  `"isTailored": "no"`,
		"effectiveFrom that is not date-time": `"effectiveFrom": "2026-01-01"`,
		"object for a list":                   `"features": {"featureType": "NPP_PAYID"}`,
	}
	for name, fields := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.ParseProduct(product(fields))

			assert.Error(t, err)
		})
	}
}

// Verifies: DH-0216
func TestParseProductNamesTheRuleThatAProductBreaks(t *testing.T) {
	const required = `"lastUpdated": "2026-08-03T01:00:00Z", "productCategory": "TERM_DEPOSITS", "name": "n",
		"description": "d", "brand": "b"`

	fee := func(fields string) []byte {
		return product(`"fees": [{"name": "Fee", "feeType": "EXIT", "feeMethodUType": "fixedAmount",
			"fixedAmount": {"amount": "1.00"}}, {` + fields + `}]`)
	}
	lendingRate := func(fields string) []byte {
		return product(`"lendingRates": [{` + fields + `}]`)
	}

	tests := map[string]struct {
		data     []byte
		wantRule string
	}{
		"no productId": {data: []byte(`{` + required + `}`), wantRule: "productId is mandatory"},
		"productId that is not ASCII": {
			data: []byte(
				`{"productId": "épargne", ` + required + `}`,
			),
			wantRule: "productId has a character that is not ASCII",
		},
		"no lastUpdated": {
			data: []byte(`{"productId": "a", "productCategory": "TERM_DEPOSITS", "name": "n", "description": "d",
				"brand": "b"}`),
			wantRule: "lastUpdated is mandatory",
		},
		"no name": {
			data: []byte(`{"productId": "a", "lastUpdated": "2026-08-03T01:00:00Z", "productCategory": "TERM_DEPOSITS",
				"description": "d", "brand": "b"}`),
			wantRule: "name is mandatory",
		},
		"name of more than 100 characters": {
			data: []byte(`{"productId": "a", "lastUpdated": "2026-08-03T01:00:00Z", "productCategory": "TERM_DEPOSITS",
				"name": "` + strings.Repeat("é", 101) + `", "description": "d", "brand": "b"}`),
			wantRule: "name has more than 100 characters",
		},
		"no description": {
			data: []byte(`{"productId": "a", "lastUpdated": "2026-08-03T01:00:00Z", "productCategory": "TERM_DEPOSITS",
				"name": "n", "brand": "b"}`),
			wantRule: "description is mandatory",
		},
		"no brand": {
			data: []byte(`{"productId": "a", "lastUpdated": "2026-08-03T01:00:00Z", "productCategory": "TERM_DEPOSITS",
				"name": "n", "description": "d"}`),
			wantRule: "brand is mandatory",
		},
		"productCategory that is not in the enumeration": {
			data: []byte(`{"productId": "a", "lastUpdated": "2026-08-03T01:00:00Z", "productCategory": "SHARES",
				"name": "n", "description": "d", "brand": "b"}`),
			wantRule: `productCategory "SHARES" is not a value of the CDS enumeration`,
		},
		"featureType that is not in the enumeration": {
			data:     product(`"features": [{"featureType": "NPP_PAYID"}, {"featureType": "FREE_COFFEE"}]`),
			wantRule: `features[1].featureType "FREE_COFFEE" is not a value of the CDS enumeration`,
		},
		"feature OTHER with no additionalInfo": {
			data:     product(`"features": [{"featureType": "OTHER"}]`),
			wantRule: "features[0]: additionalInfo is mandatory when the type is OTHER",
		},
		"fee with no name": {
			data:     fee(`"feeType": "EXIT", "feeMethodUType": "variable", "variable": {}`),
			wantRule: "fees[1].name is mandatory",
		},
		"feeType that is not in the enumeration": {
			data:     fee(`"name": "n", "feeType": "TIP", "feeMethodUType": "variable", "variable": {}`),
			wantRule: `fees[1].feeType "TIP" is not a value of the CDS enumeration`,
		},
		"fee OTHER with no additionalInfo": {
			data:     fee(`"name": "n", "feeType": "OTHER", "feeMethodUType": "variable", "variable": {}`),
			wantRule: "fees[1]: additionalInfo is mandatory when the type is OTHER",
		},
		"feeMethodUType that is not in the enumeration": {
			data:     fee(`"name": "n", "feeType": "EXIT", "feeMethodUType": "free"`),
			wantRule: `fees[1].feeMethodUType "free" is not a value of the CDS enumeration`,
		},
		"fee with no object for its feeMethodUType": {
			data: fee(`"name": "n", "feeType": "EXIT", "feeMethodUType": "rateBased",
				"fixedAmount": {"amount": "5.00"}`),
			wantRule: "fees[1]: rateBased is mandatory when feeMethodUType is rateBased",
		},
		"fee with a null object for its feeMethodUType": {
			data:     fee(`"name": "n", "feeType": "EXIT", "feeMethodUType": "variable", "variable": null`),
			wantRule: "fees[1]: variable is mandatory when feeMethodUType is variable",
		},
		"fee amount that is not an AmountString": {
			data: fee(`"name": "n", "feeType": "EXIT", "feeMethodUType": "fixedAmount",
				"fixedAmount": {"amount": "5"}`),
			wantRule: "fees[1].fixedAmount.amount is not an AmountString",
		},
		"fee rate type that is not in the enumeration": {
			data: fee(`"name": "n", "feeType": "EXIT", "feeMethodUType": "rateBased",
				"rateBased": {"rateType": "FLAT", "rate": "0.01"}`),
			wantRule: `fees[1].rateBased.rateType "FLAT" is not a value of the CDS enumeration`,
		},
		"fee rate that is not a RateString": {
			data: fee(`"name": "n", "feeType": "EXIT", "feeMethodUType": "rateBased",
				"rateBased": {"rateType": "BALANCE", "rate": "3%"}`),
			wantRule: "fees[1].rateBased.rate is not a RateString",
		},
		"fee maximum that is not an AmountString": {
			data: fee(`"name": "n", "feeType": "EXIT", "feeMethodUType": "variable",
				"variable": {"feeMaximum": "$300"}`),
			wantRule: "fees[1].variable.feeMaximum is not an AmountString",
		},
		"depositRateType that is not in the enumeration": {
			data: product(
				`"depositRates": [{"depositRateType": "HIGH", "rate": "0.01", "applicationType": "MATURITY"}]`,
			),
			wantRule: `depositRates[0].depositRateType "HIGH" is not a value of the CDS enumeration`,
		},
		"deposit rate that is not a RateString": {
			data: product(`"depositRates": [{"depositRateType": "VARIABLE", "rate": "1.55%",
				"applicationType": "MATURITY"}]`),
			wantRule: "depositRates[0].rate is not a RateString",
		},
		"deposit rate with no applicationType": {
			data:     product(`"depositRates": [{"depositRateType": "VARIABLE", "rate": "0.01"}]`),
			wantRule: `depositRates[0].applicationType "" is not a value of the CDS enumeration`,
		},
		"periodic deposit rate with no applicationFrequency": {
			data: product(
				`"depositRates": [{"depositRateType": "VARIABLE", "rate": "0.01", "applicationType": "PERIODIC"}]`,
			),
			wantRule: "depositRates[0]: applicationFrequency is mandatory when applicationType is PERIODIC",
		},
		"lendingRateType that is not in the enumeration": {
			data: lendingRate(`"lendingRateType": "LOW", "rate": "0.06", "applicationType": "MATURITY",
				"repaymentType": "INTEREST_ONLY", "loanPurpose": "INVESTMENT"`),
			wantRule: `lendingRates[0].lendingRateType "LOW" is not a value of the CDS enumeration`,
		},
		"comparison rate that is not a RateString": {
			data: lendingRate(`"lendingRateType": "FIXED", "rate": "0.06", "comparisonRate": "six",
				"applicationType": "MATURITY", "repaymentType": "INTEREST_ONLY", "loanPurpose": "INVESTMENT"`),
			wantRule: "lendingRates[0].comparisonRate is not a RateString",
		},
		"periodic lending rate with no applicationFrequency": {
			data: lendingRate(`"lendingRateType": "FIXED", "rate": "0.06", "applicationType": "PERIODIC",
				"repaymentType": "INTEREST_ONLY", "loanPurpose": "INVESTMENT"`),
			wantRule: "lendingRates[0]: applicationFrequency is mandatory when applicationType is PERIODIC",
		},
		"repaymentType that is not in the enumeration": {
			data: lendingRate(`"lendingRateType": "FIXED", "rate": "0.06", "applicationType": "MATURITY",
				"repaymentType": "NEVER", "loanPurpose": "INVESTMENT"`),
			wantRule: `lendingRates[0].repaymentType "NEVER" is not a value of the CDS enumeration`,
		},
		"lending rate with no loanPurpose": {
			data: lendingRate(`"lendingRateType": "FIXED", "rate": "0.06", "applicationType": "MATURITY",
				"repaymentType": "INTEREST_ONLY"`),
			wantRule: `lendingRates[0].loanPurpose "" is not a value of the CDS enumeration`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.ParseProduct(tc.data)

			assert.ErrorContains(t, err, tc.wantRule)
		})
	}
}

func TestParseProductNamesEachRuleThatOneProductBreaks(t *testing.T) {
	_, err := domain.ParseProduct(product(`
		"features": [{"featureType": "OTHER"}],
		"fees": [{"name": "Fee", "feeType": "EXIT", "feeMethodUType": "fixedAmount"}],
		"depositRates": [{"depositRateType": "VARIABLE", "rate": "0.01", "applicationType": "PERIODIC"}]`))

	require.Error(t, err)
	require.ErrorContains(t, err, "features[0]: additionalInfo is mandatory when the type is OTHER")
	require.ErrorContains(t, err, "fees[0]: fixedAmount is mandatory when feeMethodUType is fixedAmount")
	require.ErrorContains(t, err, "depositRates[0]: applicationFrequency is mandatory")
}

func TestProductIDGivesTheIDOfAProductThatIsNotValid(t *testing.T) {
	assert.Equal(t, "card", domain.ProductID([]byte(`{"productId": "card", "productCategory": "SHARES"}`)))
	assert.Empty(t, domain.ProductID([]byte(`{"name": "Card"}`)))
	assert.Empty(t, domain.ProductID([]byte(`not JSON`)))
}

func TestDuplicateProductIDsGivesEachIDOneTime(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, domain.DuplicateProductIDs([]string{"a", "b", "a", "c", "b", "a"}))
	assert.Empty(t, domain.DuplicateProductIDs([]string{"a", "b"}))
}

// Verifies: DH-0209
func TestParseProductCategoryAcceptsOnlyAValueOfTheCDSEnumeration(t *testing.T) {
	category, ok := domain.ParseProductCategory("TERM_DEPOSITS")
	assert.True(t, ok)
	assert.Equal(t, domain.CategoryTermDeposits, category)

	for _, value := range []string{"", "term_deposits", "SAVINGS"} {
		_, ok := domain.ParseProductCategory(value)
		assert.False(t, ok, value)
	}
}

// Verifies: DH-0218
func TestAProductNameOf100CharactersIsValid(t *testing.T) {
	_, err := domain.ParseProduct([]byte(`{"productId": "a", "lastUpdated": "2026-08-03T01:00:00Z",
		"productCategory": "TERM_DEPOSITS", "name": "` + strings.Repeat("é", 100) + `", "description": "d",
		"brand": "b"}`))

	require.NoError(t, err)
}

// The CDS puts these four categories under "Deposit Products", and each other category under "Lending Products".
func TestOnlyTheCDSDepositProductCategoriesAreDeposits(t *testing.T) {
	deposits := []domain.ProductCategory{
		domain.CategoryRegulatedTrustAccounts, domain.CategoryTermDeposits, domain.CategoryTransAndSavingsAccounts,
		domain.CategoryTravelCards,
	}
	assert.ElementsMatch(t, deposits, domain.DepositCategories())

	for _, value := range domain.ProductEnumerations()["ProductCategory"] {
		category := domain.ProductCategory(value)
		assert.Equal(t, slices.Contains(deposits, category), category.IsDeposit(), value)
	}
}

func TestDepositCategoriesGivesACopy(t *testing.T) {
	domain.DepositCategories()[0] = "CHANGED"

	assert.NotContains(t, domain.DepositCategories(), domain.ProductCategory("CHANGED"))
}
