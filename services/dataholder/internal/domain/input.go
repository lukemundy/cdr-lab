package domain

import (
	"errors"
	"slices"
	"time"
)

// AccountInput has the credit limit as text, so that an error can name the field.
type AccountInput struct {
	DisplayName     string          `json:"displayName"`
	ProductCategory ProductCategory `json:"productCategory"`
	ProductName     string          `json:"productName"`
	BSB             string          `json:"bsb"`
	AccountNumber   string          `json:"accountNumber"`
	CreditLimit     *string         `json:"creditLimit"`
}

// Account gives one ValidationError with the errors of the format and of the rules of an account.
func (a AccountInput) Account() (Account, error) {
	var faults fieldProblems

	var creditLimit Amount
	if a.CreditLimit != nil {
		creditLimit = faults.amount("creditLimit", *a.CreditLimit)
	}

	account := Account{
		DisplayName: a.DisplayName, ProductCategory: a.ProductCategory, ProductName: a.ProductName, BSB: a.BSB,
		AccountNumber: a.AccountNumber, CreditLimit: creditLimit,
	}

	if err := faults.withRules(account.Validate()); err != nil {
		return Account{}, err
	}

	return account, nil
}

// TransactionInput has the amount and the times as text, so that an error can name the field.
type TransactionInput struct {
	Type                 TransactionType   `json:"type"`
	Status               TransactionStatus `json:"status"`
	Description          string            `json:"description"`
	Amount               *string           `json:"amount"`
	Reference            string            `json:"reference"`
	PostingDateTime      *string           `json:"postingDateTime"`
	ExecutionDateTime    *string           `json:"executionDateTime"`
	ValueDateTime        *string           `json:"valueDateTime"`
	MerchantName         string            `json:"merchantName"`
	MerchantCategoryCode string            `json:"merchantCategoryCode"`
	BillerCode           string            `json:"billerCode"`
	BillerName           string            `json:"billerName"`
	CRN                  string            `json:"crn"`
	APCANumber           string            `json:"apcaNumber"`
	ExtendedData         *ExtendedData     `json:"extendedData"`
}

// Transaction gives one ValidationError with the errors of the format and of the rules of a transaction.
func (t TransactionInput) Transaction() (Transaction, error) {
	var faults fieldProblems

	var amount Amount

	faults.check(t.Amount != nil, "amount", "amount must be present")

	if t.Amount != nil {
		amount = faults.amount("amount", *t.Amount)
	}

	transaction := Transaction{
		Type:                 t.Type,
		Status:               t.Status,
		Description:          t.Description,
		Amount:               amount,
		Reference:            t.Reference,
		PostingDateTime:      faults.time("postingDateTime", t.PostingDateTime),
		ExecutionDateTime:    faults.time("executionDateTime", t.ExecutionDateTime),
		ValueDateTime:        faults.time("valueDateTime", t.ValueDateTime),
		MerchantName:         t.MerchantName,
		MerchantCategoryCode: t.MerchantCategoryCode,
		BillerCode:           t.BillerCode,
		BillerName:           t.BillerName,
		CRN:                  t.CRN,
		APCANumber:           t.APCANumber,
		ExtendedData:         t.ExtendedData,
	}

	if err := faults.withRules(transaction.Validate()); err != nil {
		return Transaction{}, err
	}

	return transaction, nil
}

// withRules drops a rule error at a field with a format error, because the rule did not get the value of the field.
func (p *fieldProblems) withRules(rules error) error {
	// Validate gives only a ValidationError or nil, and nil has no errors.
	invalid, _ := errors.AsType[ValidationError](rules)

	var all fieldProblems

	for _, rule := range invalid.Errors {
		if !slices.ContainsFunc(p.errs, func(f FieldError) bool { return f.Field == rule.Field }) {
			all.errs = append(all.errs, rule)
		}
	}

	all.errs = append(all.errs, p.errs...)

	return all.validationError()
}

func (p *fieldProblems) amount(field, text string) Amount {
	amount, err := ParseAmount(text)
	p.check(err == nil, field, "%s must be a string with 1 to 16 digits, a full stop and 2 digits", field)

	return amount
}

func (p *fieldProblems) time(field string, text *string) *time.Time {
	if text == nil {
		return nil
	}

	at, err := time.Parse(time.RFC3339, *text)
	if err != nil {
		p.check(false, field, "%s must be an RFC 3339 date-time with an offset", field)
		return nil
	}

	return &at
}
