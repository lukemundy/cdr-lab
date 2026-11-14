package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func TestTheCustomerOfAnArrangementIsTheCustomerWhoGaveConsent(t *testing.T) {
	consents, store := newConsents(t)
	own := customer()
	own.FirstName = "Alex"
	consents.Customers.(*fakeCustomerStore).customers[consentCustomer] = own
	confirmation, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-1"})
	require.NoError(t, err)

	data := app.ConsentData{Arrangements: store, Customers: consents.Customers}
	got, err := data.Customer(t.Context(), confirmation.Arrangement.ArrangementID)

	require.NoError(t, err)
	assert.Equal(t, own, got)
}

func TestAnArrangementThatDoesNotExistHasNoCustomer(t *testing.T) {
	consents, store := newConsents(t)

	_, err := app.ConsentData{Arrangements: store, Customers: consents.Customers}.Customer(t.Context(), "unknown")

	assert.ErrorIs(t, err, app.ErrArrangementNotFound)
}

// consentWithAccounts gives the consent data of a confirmation of two accounts of the consent customer. The first
// selected account is created last.
func consentWithAccounts(t *testing.T) (app.ConsentData, string) {
	t.Helper()

	consents, store := newConsents(t)
	fake := consents.Customers.(*fakeCustomerStore)
	fake.accounts[consentCustomer] = []domain.Account{
		{
			AccountID:       "account-1",
			ProductCategory: domain.CategoryTransAndSavingsAccounts,
			OpenStatus:      domain.OpenStatusOpen,
		},
		{AccountID: "account-2", ProductCategory: domain.CategoryCredAndChrgCards, OpenStatus: domain.OpenStatusOpen},
		{AccountID: "account-5", ProductCategory: domain.CategoryCredAndChrgCards, OpenStatus: domain.OpenStatusOpen},
	}
	confirmation, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-2", "account-1"})
	require.NoError(t, err)

	return app.ConsentData{Arrangements: store, Customers: consents.Customers}, confirmation.Arrangement.ArrangementID
}

var (
	idOfAccount1 = domain.ResourceID("software-1", consentCustomer, "account-1")
	idOfAccount2 = domain.ResourceID("software-1", consentCustomer, "account-2")
)

func ids(accounts []app.ConsentAccount) []string {
	found := []string{}
	for _, account := range accounts {
		found = append(found, account.ID+"="+account.Account.AccountID)
	}

	return found
}

func TestTheAccountsOfAnArrangementAreInTheSequenceOfTheirCreationWithTheIdentifiersOfTheSoftwareProduct(t *testing.T) {
	data, arrangementID := consentWithAccounts(t)

	got, total, err := data.Accounts(t.Context(), arrangementID, app.AccountQuery{Page: 1, PageSize: 25})

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, []string{idOfAccount1 + "=account-1", idOfAccount2 + "=account-2"}, ids(got))
}

func TestTheAccountsOfAnArrangementMatchEachFilter(t *testing.T) {
	owned, notOwned := true, false

	tests := map[string]struct {
		query app.AccountQuery
		want  []string
	}{
		"product category": {
			query: app.AccountQuery{Category: domain.CategoryCredAndChrgCards},
			want:  []string{idOfAccount2 + "=account-2"},
		},
		"open": {
			query: app.AccountQuery{OpenStatus: domain.OpenStatusOpen},
			want:  []string{idOfAccount2 + "=account-2"},
		},
		"closed": {
			query: app.AccountQuery{OpenStatus: domain.OpenStatusClosed},
			want:  []string{idOfAccount1 + "=account-1"},
		},
		"owned": {
			query: app.AccountQuery{IsOwned: &owned},
			want:  []string{idOfAccount1 + "=account-1", idOfAccount2 + "=account-2"},
		},
		"not owned": {query: app.AccountQuery{IsOwned: &notOwned}, want: []string{}},
		"each filter": {
			query: app.AccountQuery{Category: domain.CategoryCredAndChrgCards, OpenStatus: domain.OpenStatusClosed},
			want:  []string{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			data, arrangementID := consentWithAccounts(t)
			_, err := app.Customers{
				Store: data.Customers,
				Now:   now,
			}.CloseAccount(
				t.Context(),
				consentCustomer,
				"account-1",
			)
			require.NoError(t, err)

			tt.query.Page, tt.query.PageSize = 1, 25
			got, total, err := data.Accounts(t.Context(), arrangementID, tt.query)

			require.NoError(t, err)
			assert.Equal(t, tt.want, ids(got))
			assert.Equal(t, len(tt.want), total)
		})
	}
}

func TestTheAccountsOfAnArrangementArePaginated(t *testing.T) {
	data, arrangementID := consentWithAccounts(t)

	second, total, err := data.Accounts(t.Context(), arrangementID, app.AccountQuery{Page: 2, PageSize: 1})
	require.NoError(t, err)
	outOfRange, _, err := data.Accounts(t.Context(), arrangementID, app.AccountQuery{Page: 3, PageSize: 1})
	require.NoError(t, err)

	assert.Equal(t, 2, total)
	assert.Equal(t, []string{idOfAccount2 + "=account-2"}, ids(second))
	assert.Empty(t, outOfRange)
}

func TestAnAccountOfAnArrangementIsFoundByTheIdentifierOfTheSoftwareProduct(t *testing.T) {
	data, arrangementID := consentWithAccounts(t)

	got, err := data.Account(t.Context(), arrangementID, idOfAccount2)

	require.NoError(t, err)
	assert.Equal(t, idOfAccount2, got.ID)
	assert.Equal(t, "account-2", got.Account.AccountID)
}

func TestAnAccountThatTheArrangementDoesNotGiveIsNotFound(t *testing.T) {
	data, arrangementID := consentWithAccounts(t)
	for _, accountID := range []string{
		"unknown",
		"account-1",
		domain.ResourceID("software-1", consentCustomer, "account-5"),
		domain.ResourceID("software-2", consentCustomer, "account-1"),
	} {
		_, err := data.Account(t.Context(), arrangementID, accountID)

		assert.ErrorIs(t, err, app.ErrAccountNotFound, accountID)
	}
}

func TestAnArrangementThatDoesNotExistHasNoAccounts(t *testing.T) {
	data, _ := consentWithAccounts(t)

	_, _, listErr := data.Accounts(t.Context(), "unknown", app.AccountQuery{Page: 1, PageSize: 25})
	_, accountErr := data.Account(t.Context(), "unknown", idOfAccount2)

	require.ErrorIs(t, listErr, app.ErrArrangementNotFound)
	require.ErrorIs(t, accountErr, app.ErrArrangementNotFound)
}
