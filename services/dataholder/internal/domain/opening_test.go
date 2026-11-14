package domain_test

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func TestEachCategoryIsOneKindOfProduct(t *testing.T) {
	all := slices.Concat(domain.DepositCategories(), domain.RevolvingCategories(), domain.LoanCategories())

	assert.ElementsMatch(t, domain.ProductEnumerations()["ProductCategory"], categoryTexts(all))
	assert.ElementsMatch(t, []domain.ProductCategory{
		domain.CategoryBuyNowPayLater, domain.CategoryCredAndChrgCards, domain.CategoryMarginLoans,
		domain.CategoryOverdrafts,
	}, domain.RevolvingCategories())
	assert.ElementsMatch(t, []domain.ProductCategory{
		domain.CategoryBusinessLoans, domain.CategoryLeases, domain.CategoryPersLoans,
		domain.CategoryResidentialMortgages, domain.CategoryTradeFinance,
	}, domain.LoanCategories())
}

func categoryTexts(categories []domain.ProductCategory) []string {
	texts := make([]string, 0, len(categories))
	for _, c := range categories {
		texts = append(texts, string(c))
	}

	return texts
}

func TestAnOpeningGivesTheValueOfTheKindOfTheProductInCents(t *testing.T) {
	tests := map[string]struct {
		category domain.ProductCategory
		input    domain.OpeningInput
		want     domain.Opening
	}{
		"deposit": {category: domain.CategoryTermDeposits, want: domain.Opening{}},
		"revolving": {
			category: domain.CategoryCredAndChrgCards, input: domain.OpeningInput{CreditLimit: new("5000.00")},
			want: domain.Opening{CreditLimit: 500_000},
		},
		"loan": {
			category: domain.CategoryResidentialMortgages, input: domain.OpeningInput{LoanAmount: new("650000.50")},
			want: domain.Opening{LoanAmount: 65_000_050},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.input.Opening(tt.category)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAnOpeningValueThatIsAbsentNotAnAmountOrNotMoreThanZeroIsAnErrorOfItsField(t *testing.T) {
	tests := map[string]struct {
		category domain.ProductCategory
		input    domain.OpeningInput
		want     domain.FieldError
	}{
		"absent credit limit": {
			category: domain.CategoryOverdrafts,
			want:     domain.FieldError{Field: "creditLimit", Detail: "creditLimit must be present"},
		},
		"credit limit that is not an amount": {
			category: domain.CategoryOverdrafts, input: domain.OpeningInput{CreditLimit: new("5,000")},
			want: domain.FieldError{
				Field:  "creditLimit",
				Detail: "creditLimit must be a string with 1 to 16 digits, a full stop and 2 digits",
			},
		},
		"zero credit limit": {
			category: domain.CategoryOverdrafts, input: domain.OpeningInput{CreditLimit: new("0.00")},
			want: domain.FieldError{Field: "creditLimit", Detail: "creditLimit must be more than 0.00"},
		},
		"absent loan amount": {
			category: domain.CategoryPersLoans,
			want:     domain.FieldError{Field: "loanAmount", Detail: "loanAmount must be present"},
		},
		"negative loan amount": {
			category: domain.CategoryPersLoans, input: domain.OpeningInput{LoanAmount: new("-1.00")},
			want: domain.FieldError{Field: "loanAmount", Detail: "loanAmount must be more than 0.00"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := tt.input.Opening(tt.category)

			var invalid domain.ValidationError
			require.ErrorAs(t, err, &invalid)
			assert.Equal(t, []domain.FieldError{tt.want}, invalid.Errors)
		})
	}
}

func TestAnOpeningValueThatTheKindOfTheProductHasNotIsAnErrorOfItsField(t *testing.T) {
	tests := map[string]struct {
		category domain.ProductCategory
		input    domain.OpeningInput
		field    string
	}{
		"credit limit of a deposit": {
			category: domain.CategoryTransAndSavingsAccounts, input: domain.OpeningInput{CreditLimit: new("1.00")},
			field: "creditLimit",
		},
		"loan amount of a deposit": {
			category: domain.CategoryTravelCards, input: domain.OpeningInput{LoanAmount: new("1.00")},
			field: "loanAmount",
		},
		"loan amount of a card": {
			category: domain.CategoryCredAndChrgCards,
			input:    domain.OpeningInput{CreditLimit: new("1.00"), LoanAmount: new("1.00")}, field: "loanAmount",
		},
		"credit limit of a loan": {
			category: domain.CategoryBusinessLoans,
			input:    domain.OpeningInput{CreditLimit: new("1.00"), LoanAmount: new("1.00")}, field: "creditLimit",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := tt.input.Opening(tt.category)

			var extra domain.ValueNotForProductError
			require.ErrorAs(t, err, &extra)
			assert.Equal(t, tt.field, extra.Field)
		})
	}
}

func TestADrawdownIsAPostedDebitOfTheLoanAmountAtTheTimeOfTheOpen(t *testing.T) {
	at := time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC)

	got := domain.Drawdown(65_000_050, at)

	assert.Equal(t, domain.Transaction{
		Type: domain.TransactionOther, Status: domain.TransactionPosted, Description: "Loan drawdown",
		Amount: -65_000_050, PostingDateTime: &at, ExecutionDateTime: &at,
	}, got)
	require.NoError(t, got.Validate())
}
