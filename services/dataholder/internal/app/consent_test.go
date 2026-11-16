package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const consentCustomer = "00000001"

var consentTime = time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

func openAccount(id string) domain.Account {
	return domain.Account{AccountID: id, OpenStatus: domain.OpenStatusOpen}
}

// newConsents gives a customer with two open accounts and a closed account, and a second customer with one account.
func newConsents(t *testing.T) (app.Consents, *memoryStore) {
	t.Helper()

	customers := newFakeCustomerStore()
	customers.customers[consentCustomer] = customer()
	customers.customers["00000002"] = customer()
	customers.accounts[consentCustomer] = []domain.Account{
		openAccount("account-1"), openAccount("account-2"),
		{AccountID: "account-4", OpenStatus: domain.OpenStatusClosed},
	}
	customers.accounts["00000002"] = []domain.Account{openAccount("account-3")}

	store := &memoryStore{}

	// A confirmation takes an arrangement identifier and a sub.
	ids := []string{"arrangement-1", "subject-1", "arrangement-2", "subject-2"}

	return app.Consents{
		Customers: customers,
		Subjects:  memorySubjects{},
		Store:     store,
		NewID: func() string {
			id := ids[0]
			ids = ids[1:]

			return id
		},
		Now: func() time.Time { return consentTime },
	}, store
}

// memorySubjects is a SubjectStore whose key is the customer number and the sector identifier.
type memorySubjects map[[2]string]string

func (m memorySubjects) Subject(_ context.Context, customerNumber, sectorIdentifier, candidate string) (string, error) {
	key := [2]string{customerNumber, sectorIdentifier}
	if _, ok := m[key]; !ok {
		m[key] = candidate
	}

	return m[key], nil
}

type failingSubjects struct{ err error }

func (f failingSubjects) Subject(context.Context, string, string, string) (string, error) {
	return "", f.err
}

func consentRequest() domain.ConsentRequest {
	return domain.ConsentRequest{
		CustomerNumber:   consentCustomer,
		ClientID:         "client-1",
		SoftwareID:       "software-1",
		RecipientBaseURI: "https://copy.recipient.example",
		Scopes:           []string{"openid", "bank:accounts.basic:read"},
		SharingDuration:  90 * 24 * time.Hour,
		SectorIdentifier: "app.example",
	}
}

func TestTheAccountsOfAConsentAreTheOpenAccountsOfTheCustomer(t *testing.T) {
	consents, _ := newConsents(t)

	accounts, err := consents.Accounts(t.Context(), consentCustomer)

	require.NoError(t, err)
	assert.Equal(t, []domain.Account{openAccount("account-1"), openAccount("account-2")}, accounts)
}

func TestAConfirmationCreatesAnArrangementWithANewIdentifier(t *testing.T) {
	consents, store := newConsents(t)

	confirmation, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-2"})

	require.NoError(t, err)
	assert.Equal(t, "arrangement-1", confirmation.Arrangement.ArrangementID)
	assert.Equal(t, consentTime.Add(90*24*time.Hour), confirmation.Arrangement.SharingEndsAt())
	assert.Equal(t, "subject-1", confirmation.Subject)

	state, version, err := app.ArrangementType.Load(t.Context(), store, "arrangement-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), version)
	assert.Equal(t, domain.ArrangementState{Created: true, Arrangement: domain.Arrangement{
		ArrangementID:          "arrangement-1",
		CustomerNumber:         consentCustomer,
		ClientID:               "client-1",
		SoftwareID:             "software-1",
		RecipientBaseURI:       "https://copy.recipient.example",
		Scopes:                 []string{"openid", "bank:accounts.basic:read"},
		AccountIDs:             []string{"account-2"},
		ConfirmedAt:            consentTime,
		SharingDurationSeconds: 7_776_000,
	}}, state)
}

func TestAConfirmationThatBreaksARuleCreatesNoArrangement(t *testing.T) {
	tests := map[string]struct {
		selected []string
		want     error
	}{
		"no account":                      {selected: nil, want: domain.ErrNoAccountSelected},
		"account of a different customer": {selected: []string{"account-3"}, want: domain.ErrAccountNotOfCustomer},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			consents, store := newConsents(t)

			_, err := consents.Confirm(t.Context(), consentRequest(), tt.selected)

			require.ErrorIs(t, err, tt.want)
			assert.Empty(t, store.streams)
		})
	}
}

func TestAConfirmationOfACustomerThatCannotBeReadCreatesNoArrangement(t *testing.T) {
	consents, store := newConsents(t)
	request := consentRequest()
	request.CustomerNumber = "99999999"

	_, err := consents.Confirm(t.Context(), request, []string{"account-1"})

	require.ErrorIs(t, err, app.ErrCustomerNotFound)
	assert.Empty(t, store.streams)
}

func TestAStoreErrorOfAConfirmationHasItsOperation(t *testing.T) {
	consents, _ := newConsents(t)
	failure := errors.New("connection lost")
	consents.Store = failingStore{err: failure}

	_, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-1"})

	require.ErrorIs(t, err, failure)
	require.ErrorContains(t, err, "create the arrangement")
}

// Verifies: DH-0509
func TestTheSubjectOfACustomerIsTheSameForEachConsentWithOneSectorIdentifier(t *testing.T) {
	consents, _ := newConsents(t)

	first, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-1"})
	require.NoError(t, err)
	second, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-1"})
	require.NoError(t, err)

	assert.Equal(t, "subject-1", first.Subject)
	assert.Equal(t, "subject-1", second.Subject)
	assert.NotEqual(t, first.Arrangement.ArrangementID, second.Arrangement.ArrangementID)
}

// Verifies: DH-0509
func TestTheSubjectIsDifferentForADifferentCustomerOrSectorIdentifier(t *testing.T) {
	tests := map[string]func(*domain.ConsentRequest){
		"customer":          func(r *domain.ConsentRequest) { r.CustomerNumber = "00000002" },
		"sector identifier": func(r *domain.ConsentRequest) { r.SectorIdentifier = "other.example" },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			consents, _ := newConsents(t)
			first, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-1"})
			require.NoError(t, err)

			request := consentRequest()
			request.Scopes = []string{"openid"}
			change(&request)

			second, err := consents.Confirm(t.Context(), request, nil)

			require.NoError(t, err)
			assert.NotEqual(t, first.Subject, second.Subject)
		})
	}
}

func TestAConfirmationWithNoSubjectCreatesNoArrangement(t *testing.T) {
	consents, store := newConsents(t)
	failure := errors.New("connection lost")
	consents.Subjects = failingSubjects{err: failure}

	_, err := consents.Confirm(t.Context(), consentRequest(), []string{"account-1"})

	require.ErrorIs(t, err, failure)
	require.ErrorContains(t, err, "get the subject identifier")
	assert.Empty(t, store.streams)
}

type failingStore struct{ err error }

func (f failingStore) Read(context.Context, string, string) ([]eventsourcing.RecordedEvent, error) {
	return nil, nil
}

func (f failingStore) Append(context.Context, ...eventsourcing.Append) error { return f.err }
