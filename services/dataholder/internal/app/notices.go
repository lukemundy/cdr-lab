package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// A NoticeSender gives one of these errors when it gets no response. Each text is the result of the attempt.
var (
	ErrNoticeTimeout      = errors.New("timeout")
	ErrNoticeNoConnection = errors.New("no connection")
	ErrNoticeTLS          = errors.New("tls error")
)

// resultInternalError hides an error whose text can contain data, for example a database error.
const resultInternalError = "internal error"

type Notice struct {
	RecipientBaseURI string
	ArrangementID    string
}

type NoticeSender interface {
	// Send gives the status of the response.
	Send(ctx context.Context, notice Notice) (int, error)
}

type NoticeAttemptResult struct {
	domain.NoticeAttempt

	Outcome domain.NoticeOutcome
}

type Notices struct {
	Store         eventsourcing.Store
	Registrations RegistrationReader
	Sender        NoticeSender
	Now           func() time.Time
}

// Attempt gives domain.ErrNoNoticePending, and sends nothing, when the arrangement has no pending notice.
func (n Notices) Attempt(ctx context.Context, arrangementID string, attempt int) (NoticeAttemptResult, error) {
	state, _, err := ArrangementType.Load(ctx, n.Store, arrangementID)
	if err != nil {
		return n.internalError(ctx, attempt, err), nil
	}

	if state.NoticeStatus() != domain.NoticePending {
		return NoticeAttemptResult{}, domain.ErrNoNoticePending
	}

	baseURI, err := n.recipientBaseURI(ctx, state.Arrangement)
	if err != nil {
		return n.internalError(ctx, attempt, err), nil
	}

	status, err := n.Sender.Send(ctx, Notice{RecipientBaseURI: baseURI, ArrangementID: arrangementID})
	if err != nil {
		if known := noResponse(err); known != nil {
			return n.attempted(attempt, known.Error(), domain.NoticeTryAgain), nil
		}

		return n.internalError(ctx, attempt, err), nil
	}

	return n.attempted(attempt, "status "+strconv.Itoa(status), domain.NoticeResponseOutcome(status)), nil
}

func noResponse(err error) error {
	for _, known := range []error{ErrNoticeTimeout, ErrNoticeNoConnection, ErrNoticeTLS} {
		if errors.Is(err, known) {
			return known
		}
	}

	return nil
}

// recipientBaseURI follows a change of the registration, and is the value of the consent for a client with none.
func (n Notices) recipientBaseURI(ctx context.Context, arrangement domain.Arrangement) (string, error) {
	registration, err := n.Registrations.Registration(ctx, arrangement.ClientID)
	if errors.Is(err, ErrRegistrationNotFound) {
		return arrangement.RecipientBaseURI, nil
	}

	if err != nil {
		return "", fmt.Errorf("read the registration: %w", err)
	}

	return registration.RecipientBaseURI, nil
}

func (n Notices) attempted(attempt int, result string, outcome domain.NoticeOutcome) NoticeAttemptResult {
	return NoticeAttemptResult{
		Attempt: attempt, At: n.Now().UTC(), Result: result,
		Outcome: outcome,
	}
}

// internalError is an attempt to try again. The message of err is not logged, because database errors can contain
// query arguments.
func (n Notices) internalError(ctx context.Context, attempt int, err error) NoticeAttemptResult {
	slog.ErrorContext(ctx, "notice attempt failed", "error.type", fmt.Sprintf("%T", err))

	return n.attempted(attempt, resultInternalError, domain.NoticeTryAgain)
}

func (n Notices) RecordAttempt(ctx context.Context, arrangementID string, attempt domain.NoticeAttempt) error {
	err := ArrangementType.Execute(ctx, n.Store, arrangementID, domain.RecordNoticeAttempt{Attempt: attempt})
	if err != nil {
		return fmt.Errorf("record the notice attempt: %w", err)
	}

	return nil
}

func (n Notices) Record(
	ctx context.Context,
	arrangementID string,
	status domain.NoticeStatus,
	last domain.NoticeAttempt,
) error {
	err := ArrangementType.Execute(ctx, n.Store, arrangementID, domain.RecordNotice{
		RecordedAt: n.Now(), Status: status, Attempt: last,
	})
	if err != nil {
		return fmt.Errorf("record the notice: %w", err)
	}

	return nil
}
