package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

var (
	ErrNoAccountSelected       = errors.New("the consent has no account")
	ErrConsumerNotEligible     = errors.New("the customer has no open account")
	ErrAccountNotOfCustomer    = errors.New("the customer does not have a selected account")
	ErrArrangementExists       = errors.New("an arrangement has this identifier")
	ErrArrangementNotFound     = errors.New("no arrangement has this identifier")
	ErrArrangementEnded        = errors.New("the arrangement ended")
	ErrUnknownWithdrawalMethod = errors.New("the withdrawal has no known method")
)

// EndReason is the event that ended an arrangement before the end of its sharing.
type EndReason string

const (
	EndReasonRevokedByDataRecipient EndReason = "REVOKED_BY_DATA_RECIPIENT"
	EndReasonWithdrawnByConsumer    EndReason = "WITHDRAWN_BY_CONSUMER"
	EndReasonConsumerNotEligible    EndReason = "CONSUMER_NOT_ELIGIBLE"
	EndReasonSoftwareProductRemoved EndReason = "SOFTWARE_PRODUCT_REMOVED"
)

// WithdrawalMethod is how the consumer asked for a withdrawal (rule 9.3(1)(b) of the CDR Rules).
type WithdrawalMethod string

const (
	// WithdrawalMethodConsumerDashboard is the consumer dashboard of rule 4.25(1)(a), on the online bank site.
	WithdrawalMethodConsumerDashboard WithdrawalMethod = "CONSUMER_DASHBOARD"
	// WithdrawalMethodAlternative is the "simple alternative method of communication" of rule 4.25(1)(b): the
	// consumer asks the bank, and the operator withdraws the consent.
	WithdrawalMethodAlternative WithdrawalMethod = "ALTERNATIVE_METHOD"
)

func (m WithdrawalMethod) known() bool {
	return m == WithdrawalMethodConsumerDashboard || m == WithdrawalMethodAlternative
}

// ConsentRequest is the part of an authorisation request that an authenticated customer gives consent to.
type ConsentRequest struct {
	CustomerNumber string
	ClientID       string
	// SoftwareID is that of the registration of the client, and selects the resource identifiers.
	SoftwareID string
	// ClientName, OrgName, LegalEntityName and RecipientBaseURI are those of the registration of the client at the
	// consent.
	ClientName       string
	OrgName          string
	LegalEntityName  string
	RecipientBaseURI string
	Scopes           []string
	// SharingDuration is zero when the software product asks for none.
	SharingDuration time.Duration
	// SectorIdentifier is that of the registration of the client. It selects the sub of the customer.
	SectorIdentifier string
}

// Arrangement is a CDR arrangement: the record of one consent of a customer.
type Arrangement struct {
	ArrangementID  string `json:"arrangementId"`
	CustomerNumber string `json:"customerNumber"`
	ClientID       string `json:"clientId"`
	SoftwareID     string `json:"softwareId"`
	ClientName     string `json:"clientName"`
	// OrgName and LegalEntityName are the names of the data recipient that the consent page showed.
	OrgName         string `json:"orgName"`
	LegalEntityName string `json:"legalEntityName,omitempty"`
	// RecipientBaseURI is the target of a notice when the client has no registration.
	RecipientBaseURI string   `json:"recipientBaseUri"`
	Scopes           []string `json:"scopes"`
	// AccountIDs is empty when the scopes do not need accounts.
	AccountIDs             []string  `json:"accountIds"`
	ConfirmedAt            time.Time `json:"confirmedAt"`
	SharingDurationSeconds int64     `json:"sharingDurationSeconds"`
	// OnceOffSharingSeconds is the access token lifetime when SharingDurationSeconds is 0, because the sharing of a
	// once-off arrangement ends when its only access token expires.
	OnceOffSharingSeconds int64 `json:"onceOffSharingSeconds,omitempty"`
}

// NewArrangement gives ErrConsumerNotEligible when accounts has no open account, ErrNoAccountSelected, and
// ErrAccountNotOfCustomer for a selected identifier that is not in accounts.
func NewArrangement(
	id string,
	request ConsentRequest,
	accounts []Account,
	selected []string,
	confirmedAt time.Time,
	accessTokenLifetime time.Duration,
) (Arrangement, error) {
	if len(OpenAccounts(accounts)) == 0 {
		return Arrangement{}, ErrConsumerNotEligible
	}

	arrangement := Arrangement{
		ArrangementID:          id,
		CustomerNumber:         request.CustomerNumber,
		ClientID:               request.ClientID,
		SoftwareID:             request.SoftwareID,
		ClientName:             request.ClientName,
		OrgName:                request.OrgName,
		LegalEntityName:        request.LegalEntityName,
		RecipientBaseURI:       request.RecipientBaseURI,
		Scopes:                 request.Scopes,
		AccountIDs:             []string{},
		ConfirmedAt:            confirmedAt.UTC(),
		SharingDurationSeconds: int64(request.SharingDuration / time.Second),
	}

	if arrangement.SharingDurationSeconds == 0 {
		arrangement.OnceOffSharingSeconds = int64(accessTokenLifetime / time.Second)
	}

	if !NeedsAccounts(request.Scopes) {
		return arrangement, nil
	}

	for _, accountID := range selected {
		if !slices.ContainsFunc(accounts, func(a Account) bool { return a.AccountID == accountID }) {
			return Arrangement{}, ErrAccountNotOfCustomer
		}

		if !slices.Contains(arrangement.AccountIDs, accountID) {
			arrangement.AccountIDs = append(arrangement.AccountIDs, accountID)
		}
	}

	if len(arrangement.AccountIDs) == 0 {
		return Arrangement{}, ErrNoAccountSelected
	}

	return arrangement, nil
}

func (a Arrangement) SharingEndsAt() time.Time {
	return a.ConfirmedAt.Add(time.Duration(a.SharingDurationSeconds+a.OnceOffSharingSeconds) * time.Second)
}

type ArrangementState struct {
	Created bool
	// Arrangement is valid only when Created is true.
	Arrangement Arrangement
	// EndedAt and EndReason are zero until an event ends the arrangement before the end of its sharing.
	EndedAt   time.Time
	EndReason EndReason
	// RecordedWithdrawalMethod is zero for a withdrawal before the data holder recorded the method.
	RecordedWithdrawalMethod WithdrawalMethod
	// RecordedNotice is zero until the outcome of the notice is recorded.
	RecordedNotice NoticeStatus
	NoticeAttempts []NoticeAttempt
}

// WithdrawalMethod gives the alternative method for a withdrawal with no recorded method, because each such
// withdrawal was through the operator interface.
func (s ArrangementState) WithdrawalMethod() WithdrawalMethod {
	switch {
	case s.EndReason != EndReasonWithdrawnByConsumer:
		return ""
	case s.RecordedWithdrawalMethod != "":
		return s.RecordedWithdrawalMethod
	default:
		return WithdrawalMethodAlternative
	}
}

// EndsAt gives the time of the event that ended the arrangement, or else the end of its sharing.
func (s ArrangementState) EndsAt() time.Time {
	if !s.EndedAt.IsZero() {
		return s.EndedAt
	}

	return s.Arrangement.SharingEndsAt()
}

func (s ArrangementState) Ended(now time.Time) bool {
	return !s.EndedAt.IsZero() || !now.Before(s.Arrangement.SharingEndsAt())
}

type ArrangementCommand interface{ isArrangementCommand() }

type CreateArrangement struct {
	Arrangement Arrangement
}

func (CreateArrangement) isArrangementCommand() {}

type EndArrangement struct {
	EndedAt time.Time
	Reason  EndReason
}

func (EndArrangement) isArrangementCommand() {}

type ArrangementEvent interface{ isArrangementEvent() }

type ArrangementCreated struct {
	Arrangement Arrangement `json:"arrangement"`
}

func (ArrangementCreated) isArrangementEvent() {}

type ArrangementEnded struct {
	EndedAt time.Time `json:"endedAt"`
	Reason  EndReason `json:"reason"`
}

func (ArrangementEnded) isArrangementEvent() {}

type WithdrawArrangement struct {
	WithdrawnAt time.Time
	Method      WithdrawalMethod
}

func (WithdrawArrangement) isArrangementCommand() {}

// ArrangementWithdrawn is a separate event type so that the payloads of the ArrangementEnded withdrawals that exist,
// which have no method, do not change.
type ArrangementWithdrawn struct {
	WithdrawnAt time.Time        `json:"withdrawnAt"`
	Method      WithdrawalMethod `json:"method"`
}

func (ArrangementWithdrawn) isArrangementEvent() {}

func HandleArrangement(state ArrangementState, command ArrangementCommand) ([]ArrangementEvent, error) {
	switch c := command.(type) {
	case CreateArrangement:
		if state.Created {
			return nil, ErrArrangementExists
		}

		return []ArrangementEvent{ArrangementCreated(c)}, nil
	case EndArrangement:
		if err := canEnd(state, c.EndedAt); err != nil {
			return nil, err
		}

		return []ArrangementEvent{ArrangementEnded{EndedAt: c.EndedAt.UTC(), Reason: c.Reason}}, nil
	case WithdrawArrangement:
		if err := canEnd(state, c.WithdrawnAt); err != nil {
			return nil, err
		}

		if !c.Method.known() {
			return nil, ErrUnknownWithdrawalMethod
		}

		return []ArrangementEvent{ArrangementWithdrawn{WithdrawnAt: c.WithdrawnAt.UTC(), Method: c.Method}}, nil
	case RecordNoticeAttempt:
		return handleRecordNoticeAttempt(state, c)
	case RecordNotice:
		return handleRecordNotice(state, c)
	default:
		return nil, fmt.Errorf("the command %T is not an arrangement command", command)
	}
}

func canEnd(state ArrangementState, at time.Time) error {
	if !state.Created {
		return ErrArrangementNotFound
	}

	if state.Ended(at) {
		return ErrArrangementEnded
	}

	return nil
}

func ApplyArrangement(state ArrangementState, event ArrangementEvent) ArrangementState {
	switch e := event.(type) {
	case ArrangementCreated:
		return ArrangementState{Created: true, Arrangement: e.Arrangement}
	case ArrangementEnded:
		state.EndedAt, state.EndReason = e.EndedAt, e.Reason

		return state
	case ArrangementWithdrawn:
		state.EndedAt, state.EndReason = e.WithdrawnAt, EndReasonWithdrawnByConsumer
		state.RecordedWithdrawalMethod = e.Method

		return state
	case NoticeAttempted:
		state.NoticeAttempts = append(slices.Clone(state.NoticeAttempts), e.NoticeAttempt)

		return state
	case NoticeRecorded:
		state.RecordedNotice, state.NoticeAttempts = e.Status, e.Attempts

		return state
	default:
		return state
	}
}
