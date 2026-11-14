package domain

import "time"

// OpeningInput has the values of an open on the online bank site as text, so that an error can name the field.
type OpeningInput struct {
	CreditLimit *string `json:"creditLimit"`
	LoanAmount  *string `json:"loanAmount"`
}

// Opening has a CreditLimit only for revolving credit, and a LoanAmount only for a loan.
type Opening struct {
	CreditLimit Amount
	LoanAmount  Amount
}

// ValueNotForProductError has no value of the request, because the text can go to telemetry.
type ValueNotForProductError struct {
	Field string
}

func (e ValueNotForProductError) Error() string {
	return e.Field + " is not a value of a product of this category"
}

// Opening gives a ValueNotForProductError for a value that the category does not have, and a ValidationError for the
// value that it has.
func (o OpeningInput) Opening(category ProductCategory) (Opening, error) {
	if o.CreditLimit != nil && !category.IsRevolving() {
		return Opening{}, ValueNotForProductError{Field: "creditLimit"}
	}

	if o.LoanAmount != nil && !category.IsLoan() {
		return Opening{}, ValueNotForProductError{Field: "loanAmount"}
	}

	var faults fieldProblems

	var opening Opening

	switch {
	case category.IsRevolving():
		opening.CreditLimit = faults.positiveAmount("creditLimit", o.CreditLimit)
	case category.IsLoan():
		opening.LoanAmount = faults.positiveAmount("loanAmount", o.LoanAmount)
	}

	if err := faults.validationError(); err != nil {
		return Opening{}, err
	}

	return opening, nil
}

func (p *fieldProblems) positiveAmount(field string, text *string) Amount {
	if text == nil {
		p.check(false, field, "%s must be present", field)
		return 0
	}

	before := len(p.errs)

	amount := p.amount(field, *text)
	if len(p.errs) == before {
		p.check(amount > 0, field, "%s must be more than 0.00", field)
	}

	return amount
}

func Drawdown(loanAmount Amount, at time.Time) Transaction {
	return Transaction{
		Type: TransactionOther, Status: TransactionPosted, Description: "Loan drawdown", Amount: -loanAmount,
		PostingDateTime: &at, ExecutionDateTime: &at,
	}
}
