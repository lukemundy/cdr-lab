package domain_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Verifies: DH-2657
func TestAnAgeHasDaysHoursAndMinutesOnly(t *testing.T) {
	valid := map[string]time.Duration{
		"P3DT2H":    74 * time.Hour,
		"P0D":       0,
		"PT45M":     45 * time.Minute,
		"P1DT1H1M":  25*time.Hour + time.Minute,
		"PT1H30M":   90 * time.Minute,
		"P90D":      90 * 24 * time.Hour,
		"P001DT02H": 26 * time.Hour,
	}
	for text, want := range valid {
		age, err := domain.ParseAge(text)

		require.NoError(t, err, text)
		assert.Equal(t, want, age, text)
	}

	for _, text := range []string{
		"", "P", "PT", "P1W", "P1Y", "P1M", "PT1S", "P1DT", "-P1D", "P-1D", "P1.5D", "p1d", "P1D ", "PT1M1H",
		"P106752D", "P99999999999999999999D",
	} {
		_, err := domain.ParseAge(text)

		assert.ErrorIs(t, err, domain.ErrAgeNotValid, text)
	}
}

func jsonNames(t reflect.Type) []string {
	names := []string{}

	for field := range t.Fields() {
		names = append(names, strings.Split(field.Tag.Get("json"), ",")[0])
	}

	slices.Sort(names)

	return names
}

// Verifies: DH-2657
func TestALoadTransactionHasTheFieldsOfACreateOfATransactionWithAnAgeInPlaceOfTheTimes(t *testing.T) {
	want := slices.DeleteFunc(jsonNames(reflect.TypeFor[domain.TransactionInput]()), func(name string) bool {
		return strings.HasSuffix(name, "DateTime")
	})
	want = append(want, "age")
	slices.Sort(want)

	assert.Equal(t, want, jsonNames(reflect.TypeFor[domain.LoadTransaction]()))
}

func validLoadTransaction() domain.LoadTransaction {
	return domain.LoadTransaction{
		Type: domain.TransactionPayment, Status: domain.TransactionPosted, Description: "Rent",
		Amount: new("-1200.00"), Age: "P3DT2H",
	}
}

var loadedAt = time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)

// Verifies: DH-2658
func TestALoadTransactionHasTheTimeOfTheLoadLessItsAge(t *testing.T) {
	at := loadedAt.Add(-74 * time.Hour)

	posted, err := validLoadTransaction().Transaction(loadedAt)

	require.NoError(t, err)
	assert.Equal(t, &at, posted.ExecutionDateTime)
	assert.Equal(t, &at, posted.PostingDateTime)
	assert.Nil(t, posted.ValueDateTime)

	input := validLoadTransaction()
	input.Status = domain.TransactionPending

	pending, err := input.Transaction(loadedAt)

	require.NoError(t, err)
	assert.Equal(t, &at, pending.ExecutionDateTime)
	assert.Nil(t, pending.PostingDateTime)
}

// Verifies: DH-2659
func TestALoadTransactionObeysTheRulesOfACreateOfATransaction(t *testing.T) {
	input := validLoadTransaction()
	input.Description = ""
	input.Amount = new("-1200.5")
	input.Age = "P1W"

	_, err := input.Transaction(loadedAt)

	assert.ElementsMatch(t, []string{"age", "amount", "description"}, fields(fieldErrors(t, err)))
	assert.NotContains(t, err.Error(), "1200.5")
	assert.NotContains(t, err.Error(), "P1W")
}

func fields(errs []domain.FieldError) []string {
	names := []string{}
	for _, e := range errs {
		names = append(names, e.Field)
	}

	return names
}

func loadProduct() domain.Product {
	return domain.Product{
		ProductID: "everyday", Name: "Mock Everyday Account", ProductCategory: domain.CategoryTransAndSavingsAccounts,
	}
}

// Verifies: DH-2653, DH-2655, DH-2656
func TestALoadAccountTakesTheNameAndTheCategoryOfItsProductAndTheBSBOfTheBrand(t *testing.T) {
	account, err := domain.LoadAccount{
		ProductID:     "everyday",
		AccountNumber: "12345678",
	}.Account(
		loadProduct(),
		"123456",
	)

	require.NoError(t, err)
	assert.Equal(t, validAccount().ProductCategory, account.ProductCategory)
	assert.Equal(t, "Mock Everyday Account", account.ProductName)
	assert.Equal(t, "Mock Everyday Account", account.DisplayName, "the displayName is the name of the product")
	assert.Equal(t, "123456", account.BSB)
	assert.Equal(t, "12345678", account.AccountNumber)
	assert.Equal(t, domain.Amount(0), account.CreditLimit)
}

// Verifies: DH-2652
func TestALoadAccountCanHaveADisplayNameAndACreditLimit(t *testing.T) {
	account, err := domain.LoadAccount{
		ProductID: "everyday", AccountNumber: "12345678", DisplayName: new("Bills"), CreditLimit: new("500.00"),
	}.Account(loadProduct(), "123456")

	require.NoError(t, err)
	assert.Equal(t, "Bills", account.DisplayName)
	assert.Equal(t, domain.Amount(50_000), account.CreditLimit)
}

// Verifies: DH-2656
func TestALoadAccountIsOpenUnlessItsOpenStatusIsClosed(t *testing.T) {
	assert.False(t, domain.LoadAccount{}.Closed())
	assert.False(t, domain.LoadAccount{OpenStatus: new(domain.OpenStatusOpen)}.Closed())
	assert.True(t, domain.LoadAccount{OpenStatus: new(domain.OpenStatusClosed)}.Closed())
}

// Verifies: DH-2659
func TestALoadAccountObeysTheRulesOfACreateOfAnAccount(t *testing.T) {
	_, err := domain.LoadAccount{
		ProductID: "everyday", AccountNumber: "12345", DisplayName: new(""), CreditLimit: new("5"),
		OpenStatus: new(domain.OpenStatus("DORMANT")),
	}.Account(loadProduct(), "123456")

	assert.ElementsMatch(t,
		[]string{"accountNumber", "displayName", "creditLimit", "openStatus"}, fields(fieldErrors(t, err)))
	assert.NotContains(t, err.Error(), "DORMANT")
	assert.NotContains(t, err.Error(), "12345 ")
}

// Verifies: DH-2651, DH-2659
func TestALoadCustomerObeysTheRulesOfACreateOfACustomer(t *testing.T) {
	customer, err := domain.LoadCustomer{
		FirstName: "Ada", LastName: "Lovelace", EmailAddress: "ada@mail.test",
	}.Customer()

	require.NoError(t, err)
	assert.Equal(t, domain.Customer{FirstName: "Ada", LastName: "Lovelace", EmailAddress: "ada@mail.test"}, customer)

	_, err = domain.LoadCustomer{LastName: "", EmailAddress: "ada at mail.test"}.Customer()

	assert.ElementsMatch(t, []string{"lastName", "emailAddress"}, fields(fieldErrors(t, err)))
}

// Verifies: DH-2650
func TestACustomerLoadFileDecodesWithTheNamesOfTheSpecification(t *testing.T) {
	var load domain.CustomerLoad

	require.NoError(t, json.Unmarshal([]byte(`{"customers": [{"lastName": "Lovelace", "emailAddress": "ada@mail.test",
		"accounts": [{"productId": "everyday", "accountNumber": "12345678", "openStatus": "CLOSED",
			"transactions": [{"type": "FEE", "status": "POSTED", "description": "Fee", "amount": "-1.00",
				"age": "P1D"}]}]}]}`), &load))

	require.NotNil(t, load.Customers)
	require.Len(t, *load.Customers, 1)
	account := (*load.Customers)[0].Accounts[0]
	assert.True(t, account.Closed())
	assert.Equal(t, "P1D", account.Transactions[0].Age)
}

// Verifies: DH-2654
func TestAPositionPrefixesTheFieldAndTheDetailOfEachError(t *testing.T) {
	err := domain.ValidationError{Errors: []domain.FieldError{
		{Field: "amount", Detail: "amount must be present"},
	}}

	prefixed := domain.AtPosition("customers[0].accounts[1].transactions[2]", err)

	assert.Equal(t, []domain.FieldError{{
		Field:  "customers[0].accounts[1].transactions[2].amount",
		Detail: "customers[0].accounts[1].transactions[2]: amount must be present",
	}}, prefixed)
}
