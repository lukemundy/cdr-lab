package app_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	copiedBaseURI  = "https://copy.recipient.example"
	currentBaseURI = "https://current.recipient.example"
)

type fakeSender struct {
	status int
	err    error
	sent   []app.Notice
}

func (f *fakeSender) Send(_ context.Context, notice app.Notice) (int, error) {
	f.sent = append(f.sent, notice)

	return f.status, f.err
}

type fakeRegistrations map[string]domain.Registration

func (f fakeRegistrations) Registration(_ context.Context, clientID string) (domain.Registration, error) {
	registration, ok := f[clientID]
	if !ok {
		return domain.Registration{}, app.ErrRegistrationNotFound
	}

	return registration, nil
}

var noticeTime = consentTime.Add(2 * time.Minute)

// withdrawnArrangement gives the store with arrangement-1 of client-1, withdrawn by the consumer.
func withdrawnArrangement(t *testing.T) *memoryStore {
	t.Helper()

	arrangements, _, store := newArrangements(t, consentTime.Add(time.Minute))
	_, err := arrangements.WithdrawByConsumer(
		t.Context(),
		consentCustomer,
		"arrangement-1",
		domain.WithdrawalMethodAlternative,
	)
	require.NoError(t, err)

	return store
}

func newNotices(store *memoryStore, sender *fakeSender, registrations fakeRegistrations) app.Notices {
	return app.Notices{
		Store:         store,
		Registrations: registrations,
		Sender:        sender,
		Now:           func() time.Time { return noticeTime },
	}
}

func TestANoticeGoesToTheRecipientBaseURIOfTheCurrentRegistration(t *testing.T) {
	sender := &fakeSender{status: http.StatusNoContent}
	registrations := fakeRegistrations{"client-1": {ClientID: "client-1", RecipientBaseURI: currentBaseURI}}

	_, err := newNotices(withdrawnArrangement(t), sender, registrations).Attempt(t.Context(), "arrangement-1", 1)

	require.NoError(t, err)
	assert.Equal(t, []app.Notice{{RecipientBaseURI: currentBaseURI, ArrangementID: "arrangement-1"}}, sender.sent)
}

func TestANoticeGoesToTheCopiedRecipientBaseURIWhenTheClientHasNoRegistration(t *testing.T) {
	sender := &fakeSender{status: http.StatusNoContent}

	_, err := newNotices(withdrawnArrangement(t), sender, fakeRegistrations{}).Attempt(t.Context(), "arrangement-1", 1)

	require.NoError(t, err)
	assert.Equal(t, []app.Notice{{RecipientBaseURI: copiedBaseURI, ArrangementID: "arrangement-1"}}, sender.sent)
}

func TestAnAttemptHasItsNumberTimeResultAndOutcome(t *testing.T) {
	tests := map[string]struct {
		sender  fakeSender
		result  string
		outcome domain.NoticeOutcome
	}{
		"204": {
			sender:  fakeSender{status: http.StatusNoContent},
			result:  "status 204",
			outcome: domain.NoticeDelivered,
		},
		"400": {
			sender:  fakeSender{status: http.StatusBadRequest},
			result:  "status 400",
			outcome: domain.NoticeRefused,
		},
		"503": {
			sender:  fakeSender{status: http.StatusServiceUnavailable},
			result:  "status 503",
			outcome: domain.NoticeTryAgain,
		},
		"timeout": {
			sender:  fakeSender{err: app.ErrNoticeTimeout},
			result:  "timeout",
			outcome: domain.NoticeTryAgain,
		},
		"no connection": {
			sender:  fakeSender{err: app.ErrNoticeNoConnection},
			result:  "no connection",
			outcome: domain.NoticeTryAgain,
		},
		"tls": {
			sender:  fakeSender{err: app.ErrNoticeTLS},
			result:  "tls error",
			outcome: domain.NoticeTryAgain,
		},
		"other error": {
			sender: fakeSender{err: errors.New("sign: key 7 not found")}, result: "internal error",
			outcome: domain.NoticeTryAgain,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := newNotices(withdrawnArrangement(t), &tc.sender, fakeRegistrations{}).
				Attempt(t.Context(), "arrangement-1", 2)

			require.NoError(t, err)
			assert.Equal(t, app.NoticeAttemptResult{
				Attempt: 2, At: noticeTime, Result: tc.result,
				Outcome: tc.outcome,
			}, got)
		})
	}
}

func TestAnAttemptOfANoticeThatIsNotPendingSendsNothing(t *testing.T) {
	store := withdrawnArrangement(t)
	sender := &fakeSender{status: http.StatusNoContent}
	notices := newNotices(store, sender, fakeRegistrations{})
	require.NoError(t, notices.Record(t.Context(), "arrangement-1", domain.NoticeComplete,
		domain.NoticeAttempt{Attempt: 1, At: noticeTime, Result: "status 204"}))

	_, err := notices.Attempt(t.Context(), "arrangement-1", 1)

	require.ErrorIs(t, err, domain.ErrNoNoticePending)
	assert.Empty(t, sender.sent)
}

func TestARecordedNoticeHasItsStatusAndEachAttempt(t *testing.T) {
	store := withdrawnArrangement(t)
	notices := newNotices(store, &fakeSender{}, fakeRegistrations{})
	first := domain.NoticeAttempt{Attempt: 1, At: noticeTime, Result: "timeout"}
	second := domain.NoticeAttempt{Attempt: 2, At: noticeTime.Add(10 * time.Second), Result: "status 400"}

	require.NoError(t, notices.RecordAttempt(t.Context(), "arrangement-1", first))
	pending := arrangementState(t, store)
	assert.Equal(t, domain.NoticePending, pending.NoticeStatus())
	assert.Equal(t, []domain.NoticeAttempt{first}, pending.NoticeAttempts)

	require.NoError(t, notices.Record(t.Context(), "arrangement-1", domain.NoticeFailed, second))
	failed := arrangementState(t, store)
	assert.Equal(t, domain.NoticeFailed, failed.NoticeStatus())
	assert.Equal(t, []domain.NoticeAttempt{first, second}, failed.NoticeAttempts)
}
