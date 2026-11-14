package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func validAccountInput() domain.AccountInput {
	return domain.AccountInput{
		DisplayName: "Everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName: "Mock Everyday Account", BSB: "123456", AccountNumber: "12345678",
	}
}

func validTransactionInput() domain.TransactionInput {
	return domain.TransactionInput{
		Type: domain.TransactionPayment, Status: domain.TransactionPosted, Description: "Rent",
		Amount: new("-1200.00"), PostingDateTime: new("2026-10-15T09:30:00+08:00"),
	}
}

func TestAnAccountInputWithNoCreditLimitHasACreditLimitOfZero(t *testing.T) {
	account, err := validAccountInput().Account()

	require.NoError(t, err)
	assert.Equal(t, validAccount(), account)
}

func TestAnAccountInputGivesItsCreditLimitInCents(t *testing.T) {
	input := validAccountInput()
	input.CreditLimit = new("1500.00")

	account, err := input.Account()

	require.NoError(t, err)
	assert.Equal(t, domain.Amount(150_000), account.CreditLimit)
}

func TestACreditLimitWithNoValidFormatIsAnErrorOfItsField(t *testing.T) {
	input := validAccountInput()
	input.CreditLimit = new("1500.5")

	_, err := input.Account()

	assert.Equal(t, []domain.FieldError{{
		Field: "creditLimit", Detail: "creditLimit must be a string with 1 to 16 digits, a full stop and 2 digits",
	}}, fieldErrors(t, err))
	assert.NotContains(t, err.Error(), "1500.5")
}

func TestAnAccountInputGivesTheFormatErrorsAndTheRuleErrorsTogether(t *testing.T) {
	input := validAccountInput()
	input.DisplayName = ""
	input.CreditLimit = new("1500.5")

	_, err := input.Account()

	assert.Equal(t, []domain.FieldError{
		{Field: "displayName", Detail: "displayName must have 1 to 100 characters"},
		{Field: "creditLimit", Detail: "creditLimit must be a string with 1 to 16 digits, a full stop and 2 digits"},
	}, fieldErrors(t, err))
}

func TestATransactionInputGivesTheAmountInCentsAndEachTime(t *testing.T) {
	input := validTransactionInput()
	input.ValueDateTime = new("2026-10-15T00:00:00Z")

	transaction, err := input.Transaction()

	require.NoError(t, err)
	assert.Equal(t, domain.Amount(-120_000), transaction.Amount)
	assert.True(t, time.Date(2026, 10, 15, 1, 30, 0, 0, time.UTC).Equal(*transaction.PostingDateTime))
	assert.True(t, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC).Equal(*transaction.ValueDateTime))
	assert.Nil(t, transaction.ExecutionDateTime)
}

func TestATransactionInputGivesAnErrorAtEachFieldWithNoValidFormat(t *testing.T) {
	noAmount := validTransactionInput()
	noAmount.Amount = nil

	badValues := validTransactionInput()
	badValues.Amount = new("-1200")
	badValues.PostingDateTime = new("2026-10-15T09:30:00")
	badValues.ExecutionDateTime = new("yesterday")
	badValues.ValueDateTime = new("2026-10-15")

	tests := map[string]struct {
		input domain.TransactionInput
		want  []domain.FieldError
	}{
		"no amount": {
			input: noAmount,
			want:  []domain.FieldError{{Field: "amount", Detail: "amount must be present"}},
		},
		// The rule that a POSTED transaction has a postingDateTime gives no second error at the field.
		"amount and times": {
			input: badValues,
			want: []domain.FieldError{
				{Field: "amount", Detail: "amount must be a string with 1 to 16 digits, a full stop and 2 digits"},
				{Field: "postingDateTime", Detail: "postingDateTime must be an RFC 3339 date-time with an offset"},
				{Field: "executionDateTime", Detail: "executionDateTime must be an RFC 3339 date-time with an offset"},
				{Field: "valueDateTime", Detail: "valueDateTime must be an RFC 3339 date-time with an offset"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := tt.input.Transaction()

			assert.Equal(t, tt.want, fieldErrors(t, err))
			assert.NotContains(t, err.Error(), "-1200")
			assert.NotContains(t, err.Error(), "2026-10-15")
		})
	}
}

func TestATransactionInputGivesTheFormatErrorsAndTheRuleErrorsTogether(t *testing.T) {
	input := validTransactionInput()
	input.Description = ""
	input.Amount = new("-1200")

	_, err := input.Transaction()

	assert.Equal(t, []domain.FieldError{
		{Field: "description", Detail: "description must have 1 to 500 characters"},
		{Field: "amount", Detail: "amount must be a string with 1 to 16 digits, a full stop and 2 digits"},
	}, fieldErrors(t, err))
}
