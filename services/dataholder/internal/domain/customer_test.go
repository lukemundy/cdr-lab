package domain_test

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func validCustomer() domain.Customer {
	return domain.Customer{FirstName: "Alex", LastName: "Example", EmailAddress: "alex@example.com"}
}

func validAccount() domain.Account {
	return domain.Account{
		DisplayName:     "Everyday",
		ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName:     "Mock Everyday Account",
		BSB:             "123456",
		AccountNumber:   "12345678",
	}
}

func fieldErrors(t *testing.T, err error) []domain.FieldError {
	t.Helper()

	var invalid domain.ValidationError
	require.ErrorAs(t, err, &invalid)

	return invalid.Errors
}

func TestACustomerNumberHas8DecimalDigits(t *testing.T) {
	tests := map[string]struct {
		random uint64
		field  string
		want   string
	}{
		"zero":                  {random: 0, want: "00000000"},
		"leading zeros":         {random: 42, want: "00000042"},
		"largest number":        {random: 99_999_999, want: "99999999"},
		"first number above":    {random: 100_000_000, want: "00000000"},
		"largest random value":  {random: math.MaxUint64, want: "09551615"},
		"number with 8 digits":  {random: 12_345_678, want: "12345678"},
		"number above 8 digits": {random: 9_912_345_678, want: "12345678"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.NewCustomerNumber(tt.random))
		})
	}
}

func TestAnAccountNumberHas8DecimalDigits(t *testing.T) {
	tests := map[string]struct {
		random uint64
		want   string
	}{
		"zero":                  {random: 0, want: "00000000"},
		"leading zeros":         {random: 42, want: "00000042"},
		"largest number":        {random: 99_999_999, want: "99999999"},
		"number above 8 digits": {random: 9_912_345_678, want: "12345678"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.NewAccountNumber(tt.random))
		})
	}
}

func TestValidBSBAcceptsOnly6Digits(t *testing.T) {
	assert.True(t, domain.ValidBSB("372100"))

	for _, value := range []string{"", "37210", "3721000", "372-100", "37210a", "３７２１００"} {
		assert.False(t, domain.ValidBSB(value), value)
	}
}

func TestAValidCustomerHasNoError(t *testing.T) {
	require.NoError(t, validCustomer().Validate())

	withoutFirstName := validCustomer()
	withoutFirstName.FirstName = ""
	require.NoError(t, withoutFirstName.Validate())
}

func TestACustomerNamesEachRuleThatItBreaks(t *testing.T) {
	tests := map[string]struct {
		change func(*domain.Customer)
		field  string
		want   string
	}{
		"no last name": {
			change: func(c *domain.Customer) { c.LastName = "" },
			field:  "lastName",
			want:   "lastName must have 1 to 100 characters",
		},
		"long last name": {
			change: func(c *domain.Customer) { c.LastName = strings.Repeat("é", 101) },
			field:  "lastName",
			want:   "lastName must have 1 to 100 characters",
		},
		"long first name": {
			change: func(c *domain.Customer) { c.FirstName = strings.Repeat("a", 101) },
			field:  "firstName",
			want:   "firstName must have 100 characters at most",
		},
		"no email address": {
			change: func(c *domain.Customer) { c.EmailAddress = "" },
			field:  "emailAddress",
			want:   "emailAddress must be one email address",
		},
		"email address with no domain": {
			change: func(c *domain.Customer) { c.EmailAddress = "alex" },
			field:  "emailAddress",
			want:   "emailAddress must be one email address",
		},
		"two email addresses": {
			change: func(c *domain.Customer) { c.EmailAddress = "alex@example.com, sam@example.com" },
			field:  "emailAddress",
			want:   "emailAddress must be one email address",
		},
		"email address with a display name": {
			change: func(c *domain.Customer) { c.EmailAddress = "Alex <alex@example.com>" },
			field:  "emailAddress",
			want:   "emailAddress must be one email address",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			customer := validCustomer()
			tt.change(&customer)

			assert.Equal(
				t,
				[]domain.FieldError{{Field: tt.field, Detail: tt.want}},
				fieldErrors(t, customer.Validate()),
			)
		})
	}
}

func TestACustomerWithMoreThanOneFaultNamesEachRule(t *testing.T) {
	err := domain.Customer{}.Validate()

	assert.Equal(t, []domain.FieldError{
		{Field: "lastName", Detail: "lastName must have 1 to 100 characters"},
		{Field: "emailAddress", Detail: "emailAddress must be one email address"},
	}, fieldErrors(t, err))
	assert.EqualError(t, err, "lastName must have 1 to 100 characters; emailAddress must be one email address",
		"the operator interface gives each rule in one detail")
}

func TestALastNameOf100CharactersThatAreNotASCIIIsValid(t *testing.T) {
	customer := validCustomer()
	customer.LastName = strings.Repeat("é", 100)

	require.NoError(t, customer.Validate())
}

func TestAValidAccountHasNoError(t *testing.T) {
	require.NoError(t, validAccount().Validate())

	for _, number := range []string{"123456", "1234567890"} {
		account := validAccount()
		account.AccountNumber = number
		require.NoError(t, account.Validate(), number)
	}
}

func TestAnAccountNamesEachRuleThatItBreaks(t *testing.T) {
	tests := map[string]struct {
		change func(*domain.Account)
		field  string
		want   string
	}{
		"no display name": {
			change: func(a *domain.Account) { a.DisplayName = "" },
			field:  "displayName",
			want:   "displayName must have 1 to 100 characters",
		},
		"long product name": {
			change: func(a *domain.Account) { a.ProductName = strings.Repeat("a", 101) },
			field:  "productName",
			want:   "productName must have 1 to 100 characters",
		},
		"unknown product category": {
			change: func(a *domain.Account) { a.ProductCategory = "SAVINGS" },
			field:  "productCategory",
			want:   "productCategory must be a value of BankingProductCategoryV2",
		},
		"short bsb": {
			change: func(a *domain.Account) { a.BSB = "12345" },
			field:  "bsb",
			want:   "bsb must have 6 digits",
		},
		"bsb with a hyphen": {
			change: func(a *domain.Account) { a.BSB = "123-456" },
			field:  "bsb",
			want:   "bsb must have 6 digits",
		},
		"short account number": {
			change: func(a *domain.Account) { a.AccountNumber = "12345" },
			field:  "accountNumber",
			want:   "accountNumber must have 6 to 10 digits",
		},
		"long account number": {
			change: func(a *domain.Account) { a.AccountNumber = "12345678901" },
			field:  "accountNumber",
			want:   "accountNumber must have 6 to 10 digits",
		},
		"account number with a letter": {
			change: func(a *domain.Account) { a.AccountNumber = "12345a78" },
			field:  "accountNumber",
			want:   "accountNumber must have 6 to 10 digits",
		},
		"negative credit limit": {
			change: func(a *domain.Account) { a.CreditLimit = -1 },
			field:  "creditLimit",
			want:   "creditLimit must be zero or more",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			account := validAccount()
			tt.change(&account)

			assert.Equal(t, []domain.FieldError{{Field: tt.field, Detail: tt.want}}, fieldErrors(t, account.Validate()))
		})
	}
}

func TestAValidationErrorDoesNotShowTheValueOfAName(t *testing.T) {
	customer := validCustomer()
	customer.LastName = strings.Repeat("Z", 101)
	customer.EmailAddress = "not-an-address"

	message := customer.Validate().Error()

	assert.NotContains(t, message, "ZZZ")
	assert.NotContains(t, message, "not-an-address")
}

// Verifies: DH-1015
func TestTheMaskedNumberShowsOnlyTheLast4Digits(t *testing.T) {
	tests := map[string]string{
		"12345678":   "xxx-xxx xxxx5678",
		"123456":     "xxx-xxx xx3456",
		"1234567890": "xxx-xxx xxxxxx7890",
	}
	for accountNumber, want := range tests {
		got := domain.Account{BSB: "123456", AccountNumber: accountNumber}.MaskedNumber()

		assert.Equal(t, want, got)
	}
}

func TestAnOpenAccountCloses(t *testing.T) {
	account := validAccount()
	account.OpenStatus = domain.OpenStatusOpen

	closed, err := account.Close()

	require.NoError(t, err)
	assert.Equal(t, domain.OpenStatusClosed, closed.OpenStatus)
	assert.Equal(t, domain.OpenStatusOpen, account.OpenStatus, "Close does not change the account that it gets")
}

// Verifies: DH-1007
func TestAClosedAccountDoesNotCloseAgain(t *testing.T) {
	account := validAccount()
	account.OpenStatus = domain.OpenStatusClosed

	_, err := account.Close()

	require.ErrorIs(t, err, domain.ErrAccountClosed)
}

func TestOnlyAnOpenAccountAcceptsATransaction(t *testing.T) {
	account := validAccount()
	account.OpenStatus = domain.OpenStatusOpen

	require.NoError(t, account.AcceptTransaction())

	account.OpenStatus = domain.OpenStatusClosed

	require.ErrorIs(t, account.AcceptTransaction(), domain.ErrAccountClosed)
}

func TestOpenAccountsAreTheAccountsThatAreNotClosedInTheirSequence(t *testing.T) {
	accounts := []domain.Account{
		{AccountID: "a", OpenStatus: domain.OpenStatusOpen},
		{AccountID: "b", OpenStatus: domain.OpenStatusClosed},
		{AccountID: "c", OpenStatus: domain.OpenStatusOpen},
	}

	assert.Equal(t, []domain.Account{accounts[0], accounts[2]}, domain.OpenAccounts(accounts))
	assert.Empty(t, domain.OpenAccounts(accounts[1:2]))
}
