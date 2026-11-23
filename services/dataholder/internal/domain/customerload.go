package domain

import (
	"errors"
	"math"
	"regexp"
	"slices"
	"strconv"
	"time"
)

type CustomerLoad struct {
	// Customers is nil when the file has no customers field.
	Customers *[]LoadCustomer `json:"customers"`
}

type LoadCustomer struct {
	FirstName    string        `json:"firstName"`
	LastName     string        `json:"lastName"`
	EmailAddress string        `json:"emailAddress"`
	Accounts     []LoadAccount `json:"accounts"`
}

// LoadAccount names a product of the brand, which gives the product name and category of the account.
type LoadAccount struct {
	ProductID     string            `json:"productId"`
	AccountNumber string            `json:"accountNumber"`
	DisplayName   *string           `json:"displayName"`
	CreditLimit   *string           `json:"creditLimit"`
	OpenStatus    *OpenStatus       `json:"openStatus"`
	Transactions  []LoadTransaction `json:"transactions"`
}

// LoadTransaction has the fields of TransactionInput, with an age in place of the three times.
type LoadTransaction struct {
	Type                 TransactionType   `json:"type"`
	Status               TransactionStatus `json:"status"`
	Description          string            `json:"description"`
	Amount               *string           `json:"amount"`
	Reference            string            `json:"reference"`
	MerchantName         string            `json:"merchantName"`
	MerchantCategoryCode string            `json:"merchantCategoryCode"`
	BillerCode           string            `json:"billerCode"`
	BillerName           string            `json:"billerName"`
	CRN                  string            `json:"crn"`
	APCANumber           string            `json:"apcaNumber"`
	ExtendedData         *ExtendedData     `json:"extendedData"`
	Age                  string            `json:"age"`
}

func (c LoadCustomer) Customer() (Customer, error) {
	customer := Customer{FirstName: c.FirstName, LastName: c.LastName, EmailAddress: c.EmailAddress}
	if err := customer.Validate(); err != nil {
		return Customer{}, err
	}

	return customer, nil
}

// Closed is true when the load must close the account after its transactions.
func (a LoadAccount) Closed() bool {
	return a.OpenStatus != nil && *a.OpenStatus == OpenStatusClosed
}

// Account takes the bsb of the brand.
func (a LoadAccount) Account(product Product, bsb string) (Account, error) {
	var faults fieldProblems

	faults.check(
		a.OpenStatus == nil || slices.Contains([]OpenStatus{OpenStatusOpen, OpenStatusClosed}, *a.OpenStatus),
		"openStatus", "openStatus must be OPEN or CLOSED",
	)

	input := AccountInput{
		DisplayName: product.Name, ProductCategory: product.ProductCategory, ProductName: product.Name, BSB: bsb,
		AccountNumber: a.AccountNumber, CreditLimit: a.CreditLimit,
	}
	if a.DisplayName != nil {
		input.DisplayName = *a.DisplayName
	}

	account, err := input.Account()
	faults.addAll(err)

	if err := faults.validationError(); err != nil {
		return Account{}, err
	}

	return account, nil
}

// Transaction gives the transaction with the rules of a create of a transaction, at the time of the load less its age.
func (t LoadTransaction) Transaction(loadedAt time.Time) (Transaction, error) {
	var faults fieldProblems

	age, err := ParseAge(t.Age)
	faults.check(
		err == nil,
		"age", "age must be an RFC 3339 duration with days, hours and minutes only, for example P3DT2H",
	)

	// The create of a transaction parses the times, so that the load has the same rules as the create.
	at := loadedAt.Add(-age).Format(time.RFC3339Nano)

	input := TransactionInput{
		Type: t.Type, Status: t.Status, Description: t.Description, Amount: t.Amount, Reference: t.Reference,
		ExecutionDateTime: &at, MerchantName: t.MerchantName, MerchantCategoryCode: t.MerchantCategoryCode,
		BillerCode: t.BillerCode, BillerName: t.BillerName, CRN: t.CRN, APCANumber: t.APCANumber,
		ExtendedData: t.ExtendedData,
	}
	if t.Status == TransactionPosted {
		input.PostingDateTime = &at
	}

	transaction, err := input.Transaction()
	faults.addAll(err)

	if err := faults.validationError(); err != nil {
		return Transaction{}, err
	}

	return transaction.InUTC(), nil
}

var (
	ErrAgeNotValid = errors.New("the age is not a duration with days, hours and minutes only")
	ageParts       = regexp.MustCompile(`^P(?:([0-9]+)D)?(?:T(?:([0-9]+)H)?(?:([0-9]+)M)?)?$`)
)

// ParseAge takes a duration of RFC 3339, appendix A, with days, hours and minutes only. A day is 24 hours.
func ParseAge(text string) (time.Duration, error) {
	parts := ageParts.FindStringSubmatch(text)
	if parts == nil || (parts[1] == "" && parts[2] == "" && parts[3] == "") || text[len(text)-1] == 'T' {
		return 0, ErrAgeNotValid
	}

	var age time.Duration

	for i, unit := range []time.Duration{24 * time.Hour, time.Hour, time.Minute} {
		if parts[i+1] == "" {
			continue
		}

		count, err := strconv.ParseInt(parts[i+1], 10, 64)
		if err != nil || count > int64((math.MaxInt64-age)/unit) {
			return 0, ErrAgeNotValid
		}

		age += time.Duration(count) * unit
	}

	return age, nil
}

// addAll takes a ValidationError or nil, as the conversions of an input give.
func (p *fieldProblems) addAll(err error) {
	invalid, _ := errors.AsType[ValidationError](err)
	p.errs = append(p.errs, invalid.Errors...)
}

// AtPosition names the record of each error by its position in a file, which is not a value of the record.
func AtPosition(position string, invalid ValidationError) []FieldError {
	errs := make([]FieldError, 0, len(invalid.Errors))
	for _, e := range invalid.Errors {
		errs = append(errs, FieldError{Field: position + "." + e.Field, Detail: position + ": " + e.Detail})
	}

	return errs
}
