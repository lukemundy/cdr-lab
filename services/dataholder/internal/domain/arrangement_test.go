package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var confirmedAt = time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

func ownAccounts() []domain.Account {
	return []domain.Account{
		{AccountID: "account-1", OpenStatus: domain.OpenStatusOpen},
		{AccountID: "account-2", OpenStatus: domain.OpenStatusOpen},
	}
}

func consentRequest(scopes ...string) domain.ConsentRequest {
	return domain.ConsentRequest{
		CustomerNumber:   "00000001",
		ClientID:         "client-1",
		SoftwareID:       "software-1",
		ClientName:       "Recipient Web",
		OrgName:          "Recipient Brand",
		LegalEntityName:  "Recipient Pty Ltd",
		RecipientBaseURI: "https://recipient.example",
		Scopes:           scopes,
		SharingDuration:  90 * 24 * time.Hour,
	}
}

func TestAnArrangementRecordsTheConsent(t *testing.T) {
	request := consentRequest("openid", "bank:accounts.basic:read")

	got, err := domain.NewArrangement("arrangement-1", request, ownAccounts(), []string{"account-2"}, confirmedAt)

	require.NoError(t, err)
	assert.Equal(t, domain.Arrangement{
		ArrangementID:          "arrangement-1",
		CustomerNumber:         "00000001",
		ClientID:               "client-1",
		SoftwareID:             "software-1",
		ClientName:             "Recipient Web",
		OrgName:                "Recipient Brand",
		LegalEntityName:        "Recipient Pty Ltd",
		RecipientBaseURI:       "https://recipient.example",
		Scopes:                 []string{"openid", "bank:accounts.basic:read"},
		AccountIDs:             []string{"account-2"},
		ConfirmedAt:            confirmedAt,
		SharingDurationSeconds: 7_776_000,
	}, got)
	assert.Equal(t, time.Date(2027, 1, 9, 9, 30, 0, 0, time.UTC), got.SharingEndsAt())
}

func TestAnArrangementHasItsTimesInUTC(t *testing.T) {
	perth := time.FixedZone("AWST", 8*60*60)

	got, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "profile"), ownAccounts(), nil, confirmedAt.In(perth),
	)

	require.NoError(t, err)
	assert.Equal(t, time.UTC, got.ConfirmedAt.Location())
	assert.Equal(t, time.UTC, got.SharingEndsAt().Location())
}

func TestAnArrangementHasEachSelectedAccountOneTime(t *testing.T) {
	got, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "bank:transactions:read"), ownAccounts(),
		[]string{"account-2", "account-1", "account-2"}, confirmedAt,
	)

	require.NoError(t, err)
	assert.Equal(t, []string{"account-2", "account-1"}, got.AccountIDs)
}

func TestAnArrangementWithABankScopeNeedsAnAccount(t *testing.T) {
	_, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "bank:accounts.basic:read"), ownAccounts(), nil, confirmedAt,
	)

	assert.ErrorIs(t, err, domain.ErrNoAccountSelected)
}

func TestAnArrangementRefusesAnAccountThatTheCustomerDoesNotHave(t *testing.T) {
	_, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "bank:accounts.basic:read"), ownAccounts(),
		[]string{"account-1", "account-of-a-different-customer"}, confirmedAt,
	)

	assert.ErrorIs(t, err, domain.ErrAccountNotOfCustomer)
}

func TestACustomerWithNoOpenAccountCannotGiveAConsent(t *testing.T) {
	for _, scopes := range [][]string{{"openid", "profile"}, {"openid", "bank:accounts.basic:read"}} {
		_, err := domain.NewArrangement("arrangement-1", consentRequest(scopes...), nil, nil, confirmedAt)

		assert.ErrorIs(t, err, domain.ErrConsumerNotEligible, scopes)
	}
}

func TestAnArrangementWithNoBankScopeHasNoAccounts(t *testing.T) {
	got, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "profile"), ownAccounts(), []string{"account-1"}, confirmedAt,
	)

	require.NoError(t, err)
	assert.Empty(t, got.AccountIDs)
}

func TestAnArrangementIsCreatedOneTime(t *testing.T) {
	arrangement := domain.Arrangement{ArrangementID: "arrangement-1"}

	events, err := domain.HandleArrangement(
		domain.ArrangementState{}, domain.CreateArrangement{Arrangement: arrangement},
	)

	require.NoError(t, err)
	require.Equal(t, []domain.ArrangementEvent{domain.ArrangementCreated{Arrangement: arrangement}}, events)

	state := domain.ApplyArrangement(domain.ArrangementState{}, events[0])
	assert.Equal(t, domain.ArrangementState{Created: true, Arrangement: arrangement}, state)

	_, err = domain.HandleArrangement(state, domain.CreateArrangement{Arrangement: arrangement})
	assert.ErrorIs(t, err, domain.ErrArrangementExists)
}

func createdState(t *testing.T, sharingDuration time.Duration) domain.ArrangementState {
	t.Helper()

	request := consentRequest("openid", "profile")
	request.SharingDuration = sharingDuration
	arrangement, err := domain.NewArrangement("arrangement-1", request, ownAccounts(), nil, confirmedAt)
	require.NoError(t, err)

	return domain.ApplyArrangement(domain.ArrangementState{}, domain.ArrangementCreated{Arrangement: arrangement})
}

func end(at time.Time) domain.EndArrangement {
	return domain.EndArrangement{EndedAt: at, Reason: domain.EndReasonRevokedByDataRecipient}
}

func TestAnArrangementEndsAtTheTimeAndForTheReasonOfTheCommand(t *testing.T) {
	state := createdState(t, time.Hour)
	endedAt := confirmedAt.Add(time.Minute)

	events, err := domain.HandleArrangement(state, end(endedAt))

	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementEvent{domain.ArrangementEnded{
		EndedAt: endedAt, Reason: domain.EndReasonRevokedByDataRecipient,
	}}, events)

	ended := domain.ApplyArrangement(state, events[0])
	assert.True(t, ended.Ended(endedAt))
	assert.Equal(t, endedAt, ended.EndedAt)
	assert.Equal(t, domain.EndReasonRevokedByDataRecipient, ended.EndReason)
}

func TestAnArrangementEndsAtTheEndOfItsSharing(t *testing.T) {
	state := createdState(t, time.Hour)

	assert.False(t, state.Ended(confirmedAt.Add(time.Hour-time.Second)))
	assert.True(t, state.Ended(confirmedAt.Add(time.Hour)))
	assert.True(t, createdState(t, 0).Ended(confirmedAt), "a sharing duration of 0 ends at once")
}

func TestAnArrangementEndsAtItsEndOrElseAtTheEndOfItsSharing(t *testing.T) {
	state := createdState(t, time.Hour)
	assert.Equal(t, confirmedAt.Add(time.Hour), state.EndsAt())

	endedAt := confirmedAt.Add(time.Minute)
	events, err := domain.HandleArrangement(state, end(endedAt))
	require.NoError(t, err)

	assert.Equal(t, endedAt, domain.ApplyArrangement(state, events[0]).EndsAt())
}

func TestAnArrangementThatEndedDoesNotEndAgain(t *testing.T) {
	state := createdState(t, time.Hour)
	events, err := domain.HandleArrangement(state, end(confirmedAt.Add(time.Minute)))
	require.NoError(t, err)

	ended := domain.ApplyArrangement(state, events[0])

	_, err = domain.HandleArrangement(ended, end(confirmedAt.Add(2*time.Minute)))
	require.ErrorIs(t, err, domain.ErrArrangementEnded)

	_, err = domain.HandleArrangement(state, end(confirmedAt.Add(time.Hour)))
	require.ErrorIs(t, err, domain.ErrArrangementEnded, "the sharing ended")
}

func TestAnArrangementThatDoesNotExistCannotEnd(t *testing.T) {
	_, err := domain.HandleArrangement(domain.ArrangementState{}, end(confirmedAt))

	require.ErrorIs(t, err, domain.ErrArrangementNotFound)
}

func TestAnArrangementEndHasItsTimeInUTC(t *testing.T) {
	events, err := domain.HandleArrangement(
		createdState(t, time.Hour),
		end(confirmedAt.In(time.FixedZone("AWST", 8*60*60))),
	)

	require.NoError(t, err)
	assert.Equal(t, time.UTC, events[0].(domain.ArrangementEnded).EndedAt.Location())
}

func withdraw(at time.Time) domain.WithdrawArrangement {
	return domain.WithdrawArrangement{WithdrawnAt: at, Method: domain.WithdrawalMethodAlternative}
}

func TestAWithdrawalEndsTheArrangementAndRecordsItsMethod(t *testing.T) {
	for _, method := range []domain.WithdrawalMethod{
		domain.WithdrawalMethodConsumerDashboard, domain.WithdrawalMethodAlternative,
	} {
		t.Run(string(method), func(t *testing.T) {
			state := createdState(t, time.Hour)
			withdrawnAt := confirmedAt.Add(time.Minute).In(time.FixedZone("AWST", 8*60*60))

			events, err := domain.HandleArrangement(
				state,
				domain.WithdrawArrangement{WithdrawnAt: withdrawnAt, Method: method},
			)

			require.NoError(t, err)
			assert.Equal(t, []domain.ArrangementEvent{domain.ArrangementWithdrawn{
				WithdrawnAt: withdrawnAt.UTC(), Method: method,
			}}, events)

			withdrawn := domain.ApplyArrangement(state, events[0])
			assert.True(t, withdrawn.Ended(withdrawnAt))
			assert.Equal(t, withdrawnAt.UTC(), withdrawn.EndedAt)
			assert.Equal(t, domain.EndReasonWithdrawnByConsumer, withdrawn.EndReason)
			assert.Equal(t, method, withdrawn.WithdrawalMethod())
		})
	}
}

func TestAWithdrawalHasTheRulesOfAnEndAndNeedsAKnownMethod(t *testing.T) {
	state := createdState(t, time.Hour)

	_, err := domain.HandleArrangement(domain.ArrangementState{}, withdraw(confirmedAt))
	require.ErrorIs(t, err, domain.ErrArrangementNotFound)

	_, err = domain.HandleArrangement(state, withdraw(confirmedAt.Add(time.Hour)))
	require.ErrorIs(t, err, domain.ErrArrangementEnded, "the sharing ended")

	events, err := domain.HandleArrangement(state, end(confirmedAt.Add(time.Minute)))
	require.NoError(t, err)
	_, err = domain.HandleArrangement(domain.ApplyArrangement(state, events[0]), withdraw(confirmedAt.Add(time.Minute)))
	require.ErrorIs(t, err, domain.ErrArrangementEnded)

	for _, method := range []domain.WithdrawalMethod{"", "EMAIL"} {
		_, err = domain.HandleArrangement(
			state,
			domain.WithdrawArrangement{WithdrawnAt: confirmedAt.Add(time.Minute), Method: method},
		)
		require.ErrorIs(t, err, domain.ErrUnknownWithdrawalMethod, "method %q", method)
	}
}

// Each withdrawal before the record of the method was through the operator interface.
func TestAWithdrawalBeforeTheRecordOfTheMethodHasTheAlternativeMethod(t *testing.T) {
	withdrawn := domain.ApplyArrangement(createdState(t, time.Hour), domain.ArrangementEnded{
		EndedAt: confirmedAt.Add(time.Minute), Reason: domain.EndReasonWithdrawnByConsumer,
	})

	assert.Equal(t, domain.WithdrawalMethodAlternative, withdrawn.WithdrawalMethod())
}

func TestOnlyAWithdrawalHasAMethod(t *testing.T) {
	assert.Empty(t, createdState(t, time.Hour).WithdrawalMethod())

	for _, reason := range []domain.EndReason{
		domain.EndReasonRevokedByDataRecipient, domain.EndReasonConsumerNotEligible,
		domain.EndReasonSoftwareProductRemoved,
	} {
		ended := domain.ApplyArrangement(createdState(t, time.Hour), domain.ArrangementEnded{
			EndedAt: confirmedAt.Add(time.Minute), Reason: reason,
		})
		assert.Empty(t, ended.WithdrawalMethod(), reason)
	}
}
