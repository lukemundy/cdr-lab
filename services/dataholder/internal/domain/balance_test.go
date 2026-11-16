package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Verifies: DH-1013, DH-1014
func TestTheBalanceOfAnAccountComesFromItsTransactionsAndItsCreditLimit(t *testing.T) {
	tests := map[string]struct {
		creditLimit domain.Amount
		totals      domain.TransactionTotals
		want        domain.Balance
	}{
		"no transactions": {
			want: domain.Balance{},
		},
		"posted transactions and a pending debit": {
			totals: domain.TransactionTotals{Posted: 100_000, PendingDebits: -2_500},
			want:   domain.Balance{Current: 100_000, Available: 97_500},
		},
		"a credit limit and money owing": {
			creditLimit: 500_000,
			totals:      domain.TransactionTotals{Posted: -120_000, PendingDebits: -30_000},
			want:        domain.Balance{Current: -120_000, Available: 350_000, CreditLimit: 500_000},
		},
		"more owing than the credit limit": {
			creditLimit: 100_000,
			totals:      domain.TransactionTotals{Posted: -90_000, PendingDebits: -20_000},
			want:        domain.Balance{Current: -90_000, Available: 0, CreditLimit: 100_000},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			account := validAccount()
			account.CreditLimit = tt.creditLimit

			assert.Equal(t, tt.want, account.Balance(tt.totals))
		})
	}
}
