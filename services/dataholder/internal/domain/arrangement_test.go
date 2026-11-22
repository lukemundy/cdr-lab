package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var confirmedAt = time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

const accessTokenLifetime = 5 * time.Minute

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

// Verifies: DH-0828, DH-0829
func TestAnArrangementRecordsTheConsent(t *testing.T) {
	request := consentRequest("openid", "bank:accounts.basic:read")

	got, err := domain.NewArrangement(
		"arrangement-1", request, ownAccounts(), []string{"account-2"}, confirmedAt, accessTokenLifetime,
	)

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
		accessTokenLifetime,
	)

	require.NoError(t, err)
	assert.Equal(t, time.UTC, got.ConfirmedAt.Location())
	assert.Equal(t, time.UTC, got.SharingEndsAt().Location())
}

func TestAnArrangementHasEachSelectedAccountOneTime(t *testing.T) {
	got, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "bank:transactions:read"), ownAccounts(),
		[]string{"account-2", "account-1", "account-2"}, confirmedAt, accessTokenLifetime,
	)

	require.NoError(t, err)
	assert.Equal(t, []string{"account-2", "account-1"}, got.AccountIDs)
}

// Verifies: DH-0826
func TestAnArrangementWithABankScopeNeedsAnAccount(t *testing.T) {
	_, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "bank:accounts.basic:read"), ownAccounts(), nil, confirmedAt,
		accessTokenLifetime,
	)

	assert.ErrorIs(t, err, domain.ErrNoAccountSelected)
}

// Verifies: DH-0826
func TestAnArrangementRefusesAnAccountThatTheCustomerDoesNotHave(t *testing.T) {
	_, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "bank:accounts.basic:read"), ownAccounts(),
		[]string{"account-1", "account-of-a-different-customer"}, confirmedAt, accessTokenLifetime,
	)

	assert.ErrorIs(t, err, domain.ErrAccountNotOfCustomer)
}

// Verifies: DH-0826
func TestACustomerWithNoOpenAccountCannotGiveAConsent(t *testing.T) {
	for _, scopes := range [][]string{{"openid", "profile"}, {"openid", "bank:accounts.basic:read"}} {
		_, err := domain.NewArrangement(
			"arrangement-1", consentRequest(scopes...), nil, nil, confirmedAt, accessTokenLifetime,
		)

		assert.ErrorIs(t, err, domain.ErrConsumerNotEligible, scopes)
	}
}

// Verifies: DH-0822
func TestAnArrangementWithNoBankScopeHasNoAccounts(t *testing.T) {
	got, err := domain.NewArrangement(
		"arrangement-1", consentRequest("openid", "profile"), ownAccounts(), []string{"account-1"}, confirmedAt,
		accessTokenLifetime,
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
	arrangement, err := domain.NewArrangement(
		"arrangement-1", request, ownAccounts(), nil, confirmedAt, accessTokenLifetime,
	)
	require.NoError(t, err)

	return domain.ApplyArrangement(domain.ArrangementState{}, domain.ArrangementCreated{Arrangement: arrangement})
}

func end(at time.Time) domain.RevokeArrangement {
	return domain.RevokeArrangement{RevokedAt: at}
}

// endCause is one event of the end table, with the state that the event gives.
type endCause struct {
	command func(at time.Time) domain.ArrangementCommand
	event   func(at time.Time) domain.ArrangementEvent
	reason  domain.EndReason
	method  domain.WithdrawalMethod
	notice  domain.NoticeStatus
}

func endCauses() map[string]endCause {
	return map[string]endCause{
		"revocation": {
			command: func(at time.Time) domain.ArrangementCommand { return domain.RevokeArrangement{RevokedAt: at} },
			event:   func(at time.Time) domain.ArrangementEvent { return domain.ArrangementRevoked{RevokedAt: at} },
			reason:  domain.EndReasonRevokedByDataRecipient,
		},
		"withdrawal on the consumer dashboard": {
			command: func(at time.Time) domain.ArrangementCommand {
				return domain.WithdrawArrangement{WithdrawnAt: at, Method: domain.WithdrawalMethodConsumerDashboard}
			},
			event: func(at time.Time) domain.ArrangementEvent {
				return domain.ArrangementWithdrawn{WithdrawnAt: at, Method: domain.WithdrawalMethodConsumerDashboard}
			},
			reason: domain.EndReasonWithdrawnByConsumer,
			method: domain.WithdrawalMethodConsumerDashboard,
			notice: domain.NoticePending,
		},
		"withdrawal by the alternative method": {
			command: func(at time.Time) domain.ArrangementCommand { return withdraw(at) },
			event: func(at time.Time) domain.ArrangementEvent {
				return domain.ArrangementWithdrawn{WithdrawnAt: at, Method: domain.WithdrawalMethodAlternative}
			},
			reason: domain.EndReasonWithdrawnByConsumer,
			method: domain.WithdrawalMethodAlternative,
			notice: domain.NoticePending,
		},
		"ineligibility": {
			command: func(at time.Time) domain.ArrangementCommand {
				return domain.EndArrangementForIneligibility{EndedAt: at}
			},
			event: func(at time.Time) domain.ArrangementEvent {
				return domain.ArrangementEndedForIneligibility{EndedAt: at}
			},
			reason: domain.EndReasonConsumerNotEligible,
			notice: domain.NoticePending,
		},
		"removal": {
			command: func(at time.Time) domain.ArrangementCommand {
				return domain.EndArrangementForRemoval{EndedAt: at}
			},
			event: func(at time.Time) domain.ArrangementEvent {
				return domain.ArrangementEndedForRemoval{EndedAt: at}
			},
			reason: domain.EndReasonSoftwareProductRemoved,
		},
		"deletion": {
			command: func(at time.Time) domain.ArrangementCommand {
				return domain.EndArrangementForDeletion{EndedAt: at}
			},
			event: func(at time.Time) domain.ArrangementEvent {
				return domain.ArrangementEndedForDeletion{EndedAt: at}
			},
			reason: domain.EndReasonRevokedByDataRecipient,
		},
	}
}

// Verifies: DH-0900, DH-0902, DH-0903, DH-0904, DH-0905, DH-0909, DH-0917, DH-2618
func TestEachCauseOfAnEndHasItsOwnEventAndGivesItsStateInUTC(t *testing.T) {
	for name, cause := range endCauses() {
		t.Run(name, func(t *testing.T) {
			state := createdState(t, time.Hour)
			endedAt := confirmedAt.Add(time.Minute)

			events, err := domain.HandleArrangement(
				state, cause.command(endedAt.In(time.FixedZone("AWST", 8*60*60))),
			)

			require.NoError(t, err)
			assert.Equal(t, []domain.ArrangementEvent{cause.event(endedAt)}, events, "the event has the time in UTC")

			ended := domain.ApplyArrangement(state, events[0])
			assert.True(t, ended.Ended(endedAt))
			assert.Equal(t, endedAt, ended.EndedAt)
			assert.Equal(t, endedAt, ended.EndsAt())
			assert.Equal(t, cause.reason, ended.EndReason)
			assert.Equal(t, cause.method, ended.WithdrawalMethod)
			assert.Equal(t, cause.notice, ended.NoticeStatus())
		})
	}
}

// Verifies: DH-0834, DH-0901
func TestAnArrangementEndsAtTheEndOfItsSharing(t *testing.T) {
	state := createdState(t, time.Hour)

	assert.False(t, state.Ended(confirmedAt.Add(time.Hour-time.Second)))
	assert.True(t, state.Ended(confirmedAt.Add(time.Hour)))
}

// Verifies: DH-0834, DH-0911
func TestTheSharingOfAnArrangementWithNoSharingDurationEndsWhenItsAccessTokenExpires(t *testing.T) {
	state := createdState(t, 0)

	assert.Equal(t, confirmedAt.Add(accessTokenLifetime), state.Arrangement.SharingEndsAt())
	assert.False(t, state.Ended(confirmedAt), "the arrangement is current until its access token expires")
	assert.False(t, state.Ended(confirmedAt.Add(accessTokenLifetime-time.Second)))
	assert.True(t, state.Ended(confirmedAt.Add(accessTokenLifetime)))
	assert.Equal(t, confirmedAt.Add(time.Hour), createdState(t, time.Hour).Arrangement.SharingEndsAt(),
		"the access token lifetime does not change a sharing duration")
}

// Verifies: DH-0911
func TestAnArrangementWithNoSharingDurationCanEndBeforeItsAccessTokenExpires(t *testing.T) {
	state := createdState(t, 0)
	endedAt := confirmedAt.Add(time.Minute)

	events, err := domain.HandleArrangement(state, end(endedAt))

	require.NoError(t, err)
	assert.Equal(t, endedAt, domain.ApplyArrangement(state, events[0]).EndsAt())
}

func TestAnArrangementThatDidNotEndEndsAtTheEndOfItsSharing(t *testing.T) {
	assert.Equal(t, confirmedAt.Add(time.Hour), createdState(t, time.Hour).EndsAt())
}

// Verifies: DH-0907
func TestAnArrangementThatEndedDoesNotEndAgainForAnyCause(t *testing.T) {
	state := createdState(t, time.Hour)

	for firstName, first := range endCauses() {
		ended := domain.ApplyArrangement(state, first.event(confirmedAt.Add(time.Minute)))

		for secondName, second := range endCauses() {
			_, err := domain.HandleArrangement(ended, second.command(confirmedAt.Add(2*time.Minute)))
			require.ErrorIs(t, err, domain.ErrArrangementEnded, "%s after %s", secondName, firstName)
		}
	}

	for name, cause := range endCauses() {
		_, err := domain.HandleArrangement(state, cause.command(confirmedAt.Add(time.Hour)))
		require.ErrorIs(t, err, domain.ErrArrangementEnded, "%s after the end of the sharing", name)
	}
}

func TestAnArrangementThatDoesNotExistCannotEnd(t *testing.T) {
	for name, cause := range endCauses() {
		_, err := domain.HandleArrangement(domain.ArrangementState{}, cause.command(confirmedAt))

		require.ErrorIs(t, err, domain.ErrArrangementNotFound, name)
	}
}

func withdraw(at time.Time) domain.WithdrawArrangement {
	return domain.WithdrawArrangement{WithdrawnAt: at, Method: domain.WithdrawalMethodAlternative}
}

// Verifies: DH-0917
func TestAWithdrawalNeedsAKnownMethod(t *testing.T) {
	for _, method := range []domain.WithdrawalMethod{"", "EMAIL"} {
		_, err := domain.HandleArrangement(
			createdState(t, time.Hour),
			domain.WithdrawArrangement{WithdrawnAt: confirmedAt.Add(time.Minute), Method: method},
		)
		require.ErrorIs(t, err, domain.ErrUnknownWithdrawalMethod, "method %q", method)
	}
}

func TestAnArrangementThatDidNotEndHasNoEndReasonOrWithdrawalMethod(t *testing.T) {
	state := createdState(t, time.Hour)

	assert.Empty(t, state.EndReason)
	assert.Empty(t, state.WithdrawalMethod)
}
