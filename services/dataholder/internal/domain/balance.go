package domain

// TransactionTotals has the sum of the posted amounts, and the sum of the pending amounts that are negative.
type TransactionTotals struct {
	Posted        Amount
	PendingDebits Amount
}

type Balance struct {
	Current     Amount
	Available   Amount
	CreditLimit Amount
}

func (a Account) Balance(totals TransactionTotals) Balance {
	return Balance{
		Current:     totals.Posted,
		Available:   max(a.CreditLimit+totals.Posted+totals.PendingDebits, 0),
		CreditLimit: a.CreditLimit,
	}
}
