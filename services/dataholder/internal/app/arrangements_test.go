package app_test

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type memoryGrants struct {
	revoked []string
	err     error
}

func (m *memoryGrants) RevokeArrangement(_ context.Context, arrangementID string) error {
	if m.err != nil {
		return m.err
	}

	m.revoked = append(m.revoked, arrangementID)

	return nil
}

// newArrangements gives the arrangement arrangement-1 of client-1, with a sharing duration of one hour.
func newArrangements(t *testing.T, now time.Time) (app.Arrangements, *memoryGrants, *memoryStore) {
	t.Helper()

	consents, store := newConsents(t)
	request := consentRequest()
	request.SharingDuration = time.Hour
	_, err := consents.Confirm(t.Context(), request, []string{"account-1"})
	require.NoError(t, err)

	grants := &memoryGrants{}

	return app.Arrangements{
		Store:     store,
		Grants:    grants,
		Customers: consents.Customers,
		Reader:    memoryReader{store: store},
		Now:       func() time.Time { return now },
	}, grants, store
}

// memoryReader gives the arrangements of the store in the sequence of their confirmation.
type memoryReader struct{ store *memoryStore }

func (m memoryReader) ArrangementsOf(ctx context.Context, customerNumber string) ([]domain.ArrangementState, error) {
	return m.where(ctx, func(a domain.Arrangement) bool { return a.CustomerNumber == customerNumber })
}

func (m memoryReader) ArrangementsOfSoftwareProduct(
	ctx context.Context,
	softwareID string,
) ([]domain.ArrangementState, error) {
	return m.where(ctx, func(a domain.Arrangement) bool { return a.SoftwareID == softwareID })
}

func (m memoryReader) where(
	ctx context.Context,
	keep func(domain.Arrangement) bool,
) ([]domain.ArrangementState, error) {
	var states []domain.ArrangementState

	for key := range m.store.streams {
		id, found := strings.CutPrefix(key, app.ArrangementType.StreamType+"/")
		if !found {
			continue
		}

		state, _, err := app.ArrangementType.Load(ctx, m.store, id)
		if err != nil {
			return nil, err
		}

		if keep(state.Arrangement) {
			states = append(states, state)
		}
	}

	slices.SortFunc(states, func(a, b domain.ArrangementState) int {
		return cmp.Or(
			a.Arrangement.ConfirmedAt.Compare(b.Arrangement.ConfirmedAt),
			cmp.Compare(a.Arrangement.ArrangementID, b.Arrangement.ArrangementID),
		)
	})

	return states, nil
}

func arrangementState(t *testing.T, store *memoryStore) domain.ArrangementState {
	t.Helper()

	state, _, err := app.ArrangementType.Load(t.Context(), store, "arrangement-1")
	require.NoError(t, err)

	return state
}

// Verifies: DH-0902, DH-0908
func TestARevocationByTheDataRecipientEndsTheArrangementAndRevokesItsGrant(t *testing.T) {
	now := consentTime.Add(time.Minute)
	arrangements, grants, store := newArrangements(t, now)

	require.NoError(t, arrangements.RevokeByDataRecipient(t.Context(), "client-1", "arrangement-1"))

	assert.Equal(t, []string{"arrangement-1"}, grants.revoked)

	state := arrangementState(t, store)
	assert.Equal(t, now, state.EndedAt)
	assert.Equal(t, domain.EndReasonRevokedByDataRecipient, state.EndReason)
}

func TestARevocationThatTheDataRecipientCannotMakeGivesInvalidArrangementAndChangesNothing(t *testing.T) {
	tests := map[string]struct {
		clientID, arrangementID string
		after                   time.Duration
	}{
		"arrangement that does not exist":   {"client-1", "arrangement-9", time.Minute},
		"arrangement of a different client": {"client-2", "arrangement-1", time.Minute},
		"arrangement whose sharing ended":   {"client-1", "arrangement-1", time.Hour},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			arrangements, grants, store := newArrangements(t, consentTime.Add(tc.after))

			err := arrangements.RevokeByDataRecipient(t.Context(), tc.clientID, tc.arrangementID)

			require.ErrorIs(t, err, app.ErrInvalidArrangement)
			assert.Empty(t, grants.revoked)
			assert.True(t, arrangementState(t, store).EndedAt.IsZero())
		})
	}
}

func TestASecondRevocationGivesInvalidArrangement(t *testing.T) {
	arrangements, _, _ := newArrangements(t, consentTime.Add(time.Minute))
	require.NoError(t, arrangements.RevokeByDataRecipient(t.Context(), "client-1", "arrangement-1"))

	err := arrangements.RevokeByDataRecipient(t.Context(), "client-1", "arrangement-1")

	require.ErrorIs(t, err, app.ErrInvalidArrangement)
}

func TestAnArrangementDoesNotEndWhenItsGrantIsNotRevoked(t *testing.T) {
	arrangements, grants, store := newArrangements(t, consentTime.Add(time.Minute))
	grants.err = errors.New("database failure")

	err := arrangements.RevokeByDataRecipient(t.Context(), "client-1", "arrangement-1")

	require.Error(t, err)
	require.NotErrorIs(t, err, app.ErrInvalidArrangement)
	assert.True(t, arrangementState(t, store).EndedAt.IsZero(), "a new request can end it")
}

// Verifies: DH-0903, DH-0908, DH-0917
func TestAWithdrawalByTheConsumerEndsTheArrangementWithTheMethodOfTheAdapterAndRevokesItsGrant(t *testing.T) {
	for _, method := range []domain.WithdrawalMethod{
		domain.WithdrawalMethodConsumerDashboard, domain.WithdrawalMethodAlternative,
	} {
		t.Run(string(method), func(t *testing.T) {
			now := consentTime.Add(time.Minute)
			arrangements, grants, store := newArrangements(t, now)

			withdrawn, err := arrangements.WithdrawByConsumer(t.Context(), consentCustomer, "arrangement-1", method)

			require.NoError(t, err)
			assert.Equal(t, []string{"arrangement-1"}, grants.revoked)
			assert.Equal(t, arrangementState(t, store), withdrawn)
			assert.Equal(t, now, withdrawn.EndedAt)
			assert.Equal(t, domain.EndReasonWithdrawnByConsumer, withdrawn.EndReason)
			assert.Equal(t, method, withdrawn.RecordedWithdrawalMethod)
		})
	}
}

// The stored events of a withdrawal before the record of the method.
func TestAStoredWithdrawalWithNoMethodHasTheAlternativeMethod(t *testing.T) {
	recorded := []eventsourcing.RecordedEvent{
		{
			EventType: "ArrangementCreated", SchemaVersion: 1, StreamVersion: 1,
			Payload: []byte(`{"arrangement": {"arrangementId": "arrangement-1", "customerNumber": "00000001",
				"clientId": "client-1", "softwareId": "software-1", "clientName": "Recipient Web",
				"recipientBaseUri": "https://recipient.example", "scopes": ["openid"], "accountIds": [],
				"confirmedAt": "2026-10-17T01:00:00Z", "sharingDurationSeconds": 3600}}`),
		},
		{
			EventType: "ArrangementEnded", SchemaVersion: 1, StreamVersion: 2,
			Payload: []byte(`{"endedAt": "2026-10-17T01:30:00Z", "reason": "WITHDRAWN_BY_CONSUMER"}`),
		},
	}

	state, _, err := app.ArrangementType.Fold("arrangement-1", recorded)

	require.NoError(t, err)
	assert.Equal(t, domain.EndReasonWithdrawnByConsumer, state.EndReason)
	assert.Equal(t, domain.WithdrawalMethodAlternative, state.WithdrawalMethod())
}

func TestAWithdrawalThatTheOperatorCannotMakeGivesItsErrorAndChangesNothing(t *testing.T) {
	tests := map[string]struct {
		customerNumber, arrangementID string
		after                         time.Duration
		want                          error
	}{
		"customer that does not exist": {"00000009", "arrangement-1", time.Minute, app.ErrCustomerNotFound},
		"arrangement that does not exist": {
			consentCustomer,
			"arrangement-9",
			time.Minute,
			domain.ErrArrangementNotFound,
		},
		"arrangement of a different customer": {
			"00000002",
			"arrangement-1",
			time.Minute,
			domain.ErrArrangementNotFound,
		},
		"arrangement whose sharing ended": {
			consentCustomer,
			"arrangement-1",
			time.Hour,
			domain.ErrArrangementEnded,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			arrangements, grants, store := newArrangements(t, consentTime.Add(tc.after))

			_, err := arrangements.WithdrawByConsumer(
				t.Context(),
				tc.customerNumber,
				tc.arrangementID,
				domain.WithdrawalMethodAlternative,
			)

			require.ErrorIs(t, err, tc.want)
			assert.Empty(t, grants.revoked)
			assert.True(t, arrangementState(t, store).EndedAt.IsZero())
		})
	}
}

func TestASecondWithdrawalGivesArrangementEnded(t *testing.T) {
	arrangements, _, _ := newArrangements(t, consentTime.Add(time.Minute))
	_, err := arrangements.WithdrawByConsumer(
		t.Context(),
		consentCustomer,
		"arrangement-1",
		domain.WithdrawalMethodAlternative,
	)
	require.NoError(t, err)

	_, err = arrangements.WithdrawByConsumer(
		t.Context(),
		consentCustomer,
		"arrangement-1",
		domain.WithdrawalMethodAlternative,
	)

	require.ErrorIs(t, err, domain.ErrArrangementEnded)
}

func TestTheArrangementsOfACustomerNeedTheCustomer(t *testing.T) {
	arrangements, _, store := newArrangements(t, consentTime.Add(time.Minute))

	own, err := arrangements.OfCustomer(t.Context(), consentCustomer)
	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementState{arrangementState(t, store)}, own)

	_, err = arrangements.OfCustomer(t.Context(), "00000009")
	require.ErrorIs(t, err, app.ErrCustomerNotFound)
}

// addArrangement adds an arrangement with no account and a sharing duration of 90 days.
func addArrangement(t *testing.T, store *memoryStore, id, customerNumber, clientID string) {
	t.Helper()

	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, id, domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: id, CustomerNumber: customerNumber, ClientID: clientID,
			Scopes: []string{"openid", "profile"}, AccountIDs: []string{},
			ConfirmedAt: consentTime, SharingDurationSeconds: int64(90 * 24 * time.Hour / time.Second),
		},
	}))
}

func stateOf(t *testing.T, store *memoryStore, id string) domain.ArrangementState {
	t.Helper()

	state, _, err := app.ArrangementType.Load(t.Context(), store, id)
	require.NoError(t, err)

	return state
}

// Verifies: DH-0904, DH-0918, DH-0921
func TestTheEndForACustomerWhoIsNotEligibleEndsEachArrangementThatDidNotEndAtTheTimeOfTheClose(t *testing.T) {
	closedAt := consentTime.Add(10 * time.Minute)
	// The time of the request does not change the time of the end, so that a later attempt records the close.
	arrangements, grants, store := newArrangements(t, closedAt.Add(time.Minute))
	addArrangement(t, store, "arrangement-2", consentCustomer, "client-2")
	addArrangement(t, store, "arrangement-3", consentCustomer, "client-1")
	addArrangement(t, store, "arrangement-4", "00000002", "client-1")
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-3", domain.EndArrangement{
		EndedAt: consentTime.Add(time.Minute), Reason: domain.EndReasonWithdrawnByConsumer,
	}))

	require.NoError(t, arrangements.EndForIneligibleCustomer(t.Context(), consentCustomer, closedAt))

	assert.Equal(t, []string{"arrangement-1", "arrangement-2"}, grants.revoked)

	for _, id := range []string{"arrangement-1", "arrangement-2"} {
		state := stateOf(t, store, id)
		assert.Equal(t, closedAt, state.EndedAt, id)
		assert.Equal(t, domain.EndReasonConsumerNotEligible, state.EndReason, id)
	}

	assert.Equal(t, domain.EndReasonWithdrawnByConsumer, stateOf(t, store, "arrangement-3").EndReason)
	assert.True(t, stateOf(t, store, "arrangement-4").EndedAt.IsZero(), "a different customer")
}

func TestTheEndForACustomerWhoIsNotEligibleIsSafeToRepeat(t *testing.T) {
	closedAt := consentTime.Add(10 * time.Minute)
	arrangements, grants, store := newArrangements(t, closedAt)
	require.NoError(t, arrangements.EndForIneligibleCustomer(t.Context(), consentCustomer, closedAt))

	events := len(store.streams[app.ArrangementType.StreamType+"/arrangement-1"])

	require.NoError(t, arrangements.EndForIneligibleCustomer(t.Context(), consentCustomer, closedAt))

	assert.Equal(t, []string{"arrangement-1"}, grants.revoked)
	assert.Len(t, store.streams[app.ArrangementType.StreamType+"/arrangement-1"], events)
}

func TestTheEndForACustomerWhoIsNotEligibleDoesNotChangeAnArrangementWhoseSharingEndedBeforeTheClose(t *testing.T) {
	closedAt := consentTime.Add(2 * time.Hour)
	arrangements, grants, store := newArrangements(t, closedAt)

	require.NoError(t, arrangements.EndForIneligibleCustomer(t.Context(), consentCustomer, closedAt))

	assert.Empty(t, grants.revoked)
	assert.Empty(t, arrangementState(t, store).EndReason)
}

func TestTheEndForACustomerWhoIsNotEligibleGivesTheErrorOfAnEnd(t *testing.T) {
	arrangements, grants, _ := newArrangements(t, consentTime.Add(time.Minute))
	grants.err = errors.New("database failure")

	err := arrangements.EndForIneligibleCustomer(t.Context(), consentCustomer, consentTime.Add(time.Minute))

	require.ErrorIs(t, err, grants.err)
}

func registerSoftwareProduct(t *testing.T, store *memoryStore, softwareID, clientID string) {
	t.Helper()

	require.NoError(t, app.ClientRegistrationType.Execute(t.Context(), store, softwareID, domain.RegisterClient{
		Registration: domain.Registration{ClientID: clientID, SoftwareID: softwareID},
	}))
}

func TestTheRemovalOfASoftwareProductEndsItsArrangementsAndDeletesItsRegistration(t *testing.T) {
	removedAt := consentTime.Add(10 * time.Minute)
	arrangements, grants, store := newArrangements(t, removedAt.Add(time.Minute))
	registerSoftwareProduct(t, store, "software-1", "client-1")
	registerSoftwareProduct(t, store, "software-2", "client-2")
	addArrangement(t, store, "arrangement-2", consentCustomer, "client-2")

	require.NoError(t, arrangements.RemoveSoftwareProduct(t.Context(), "software-1", removedAt))

	removed := stateOf(t, store, "arrangement-1")
	assert.Equal(t, domain.EndReasonSoftwareProductRemoved, removed.EndReason)
	assert.Equal(t, removedAt, removed.EndedAt, "the end has the time at which the data holder got the status")
	assert.Equal(t, []string{"arrangement-1"}, grants.revoked)
	assert.True(t, stateOf(t, store, "arrangement-2").EndedAt.IsZero(), "a different software product")

	registration, _, err := app.ClientRegistrationType.Load(t.Context(), store, "software-1")
	require.NoError(t, err)
	assert.False(t, registration.Registered)

	other, _, err := app.ClientRegistrationType.Load(t.Context(), store, "software-2")
	require.NoError(t, err)
	assert.True(t, other.Registered)
}

func TestTheRemovalOfASoftwareProductIsSafeToRepeat(t *testing.T) {
	removedAt := consentTime.Add(10 * time.Minute)
	arrangements, grants, store := newArrangements(t, removedAt)
	registerSoftwareProduct(t, store, "software-1", "client-1")
	require.NoError(t, arrangements.RemoveSoftwareProduct(t.Context(), "software-1", removedAt))

	require.NoError(t, arrangements.RemoveSoftwareProduct(t.Context(), "software-1", removedAt))

	assert.Equal(t, []string{"arrangement-1"}, grants.revoked)
}

func TestTheRemovalOfASoftwareProductWithNoRegistrationEndsItsArrangements(t *testing.T) {
	removedAt := consentTime.Add(10 * time.Minute)
	arrangements, _, store := newArrangements(t, removedAt)

	require.NoError(t, arrangements.RemoveSoftwareProduct(t.Context(), "software-1", removedAt))

	assert.Equal(t, domain.EndReasonSoftwareProductRemoved, stateOf(t, store, "arrangement-1").EndReason)
}
