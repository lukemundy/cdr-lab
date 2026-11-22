package console

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// CustomerDetails gives app.ErrCustomerNotFound, and for the transactions also app.ErrAccountNotFound.
type CustomerDetails interface {
	Detail(ctx context.Context, customerNumber string) (app.CustomerDetail, error)
	TransactionPage(
		ctx context.Context,
		customerNumber, accountID string,
		page, pageSize int,
	) ([]domain.Transaction, int, error)
}

type customerDetail struct {
	CustomerNumber string        `json:"customerNumber"`
	FirstName      string        `json:"firstName,omitempty"`
	LastName       string        `json:"lastName"`
	EmailAddress   string        `json:"emailAddress"`
	LastUpdated    time.Time     `json:"lastUpdated"`
	Accounts       []account     `json:"accounts"`
	Arrangements   []arrangement `json:"arrangements"`
}

type account struct {
	AccountID        string                         `json:"accountId"`
	DisplayName      string                         `json:"displayName"`
	ProductCategory  domain.ProductCategory         `json:"productCategory"`
	ProductName      string                         `json:"productName"`
	BSB              string                         `json:"bsb"`
	AccountNumber    string                         `json:"accountNumber"`
	CreditLimit      domain.Amount                  `json:"creditLimit"`
	CurrentBalance   domain.Amount                  `json:"currentBalance"`
	AvailableBalance domain.Amount                  `json:"availableBalance"`
	OpenStatus       domain.OpenStatus              `json:"openStatus"`
	CreationDate     string                         `json:"creationDate"`
	Actions          []action[domain.AccountAction] `json:"actions"`
}

type action[A ~string] struct {
	Action A                `json:"action"`
	Ends   []arrangementEnd `json:"ends"`
}

type arrangementEnd struct {
	ArrangementID string `json:"arrangementId"`
	SoftwareID    string `json:"softwareId"`
	ClientName    string `json:"clientName"`
}

type arrangementStatus string

const (
	arrangementCurrent arrangementStatus = "CURRENT"
	arrangementEnded   arrangementStatus = "ENDED"
)

type arrangement struct {
	ArrangementID    string                  `json:"arrangementId"`
	ClientID         string                  `json:"clientId"`
	ClientRegistered bool                    `json:"clientRegistered"`
	SoftwareID       string                  `json:"softwareId"`
	ClientName       string                  `json:"clientName"`
	Scopes           []string                `json:"scopes"`
	AccountIDs       []string                `json:"accountIds"`
	ConfirmedAt      time.Time               `json:"confirmedAt"`
	SharingEndsAt    time.Time               `json:"sharingEndsAt"`
	Status           arrangementStatus       `json:"status"`
	EndedAt          *time.Time              `json:"endedAt,omitempty"`
	EndReason        domain.EndReason        `json:"endReason,omitempty"`
	WithdrawalMethod domain.WithdrawalMethod `json:"withdrawalMethod,omitempty"`
	NoticeStatus     domain.NoticeStatus     `json:"noticeStatus,omitempty"`
	// NoticeAttempts is nil, and not in the response, when the end needs no notice.
	NoticeAttempts []noticeAttempt                    `json:"noticeAttempts,omitzero"`
	Actions        []action[domain.ArrangementAction] `json:"actions"`
}

type noticeAttempt struct {
	Attempt int       `json:"attempt"`
	At      time.Time `json:"at"`
	Result  string    `json:"result"`
}

type transactionPage struct {
	Data         []domain.Transaction `json:"data"`
	TotalRecords int                  `json:"totalRecords"`
}

// A 404 detail does not repeat the identifier, because it is customer data.
func writeDetailError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, app.ErrCustomerNotFound):
		problem.Write(w, http.StatusNotFound, "The data holder has no customer with this customer number.")
	case errors.Is(err, app.ErrAccountNotFound):
		problem.Write(w, http.StatusNotFound, "The customer has no account with this account ID.")
	default:
		problem.WriteUnexpected(w, r, err, notCompleted)
	}
}

func readCustomer(details CustomerDetails, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detail, err := details.Detail(r.Context(), r.PathValue("customerNumber"))
		if err != nil {
			writeDetailError(w, r, err)
			return
		}

		at := now()
		c := detail.Customer
		body := customerDetail{
			CustomerNumber: c.CustomerNumber,
			FirstName:      c.FirstName,
			LastName:       c.LastName,
			EmailAddress:   c.EmailAddress,
			LastUpdated:    c.LastUpdated.UTC(),
			Accounts:       make([]account, 0, len(detail.Accounts)),
			Arrangements:   make([]arrangement, 0, len(detail.Arrangements)),
		}

		accounts := make([]domain.Account, 0, len(detail.Accounts))
		for _, a := range detail.Accounts {
			accounts = append(accounts, a.Account)
		}

		states := make([]domain.ArrangementState, 0, len(detail.Arrangements))
		for _, a := range detail.Arrangements {
			states = append(states, a.ArrangementState)
		}

		actions := domain.PermittedAccountActions(accounts, states, at)

		for _, a := range detail.Accounts {
			body.Accounts = append(body.Accounts, account{
				AccountID:        a.Account.AccountID,
				DisplayName:      a.Account.DisplayName,
				ProductCategory:  a.Account.ProductCategory,
				ProductName:      a.Account.ProductName,
				BSB:              a.Account.BSB,
				AccountNumber:    a.Account.AccountNumber,
				CreditLimit:      a.Balance.CreditLimit,
				CurrentBalance:   a.Balance.Current,
				AvailableBalance: a.Balance.Available,
				OpenStatus:       a.Account.OpenStatus,
				// The same value as the creationDate of the banking APIs.
				CreationDate: a.Account.CreatedAt.UTC().Format(time.DateOnly),
				Actions:      actionsOf(actions[a.Account.AccountID]),
			})
		}

		withdrawals := domain.PermittedArrangementActions(states, at)

		for _, state := range detail.Arrangements {
			view := arrangementOf(state, at)
			view.Actions = actionsOf(withdrawals[state.Arrangement.ArrangementID])
			body.Arrangements = append(body.Arrangements, view)
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}

func actionsOf[A ~string](actions []domain.Action[A]) []action[A] {
	body := make([]action[A], 0, len(actions))
	for _, a := range actions {
		ends := make([]arrangementEnd, 0, len(a.Ends))
		for _, e := range a.Ends {
			ends = append(ends, arrangementEnd{
				ArrangementID: e.Arrangement.ArrangementID,
				SoftwareID:    e.Arrangement.SoftwareID,
				ClientName:    e.Arrangement.ClientName,
			})
		}

		body = append(body, action[A]{Action: a.Action, Ends: ends})
	}

	return body
}

func arrangementOf(detail app.ArrangementDetail, now time.Time) arrangement {
	state := detail.ArrangementState
	a := state.Arrangement
	view := arrangement{
		ArrangementID:    a.ArrangementID,
		ClientID:         a.ClientID,
		ClientRegistered: detail.ClientRegistered,
		SoftwareID:       a.SoftwareID,
		ClientName:       a.ClientName,
		Scopes:           a.Scopes,
		AccountIDs:       a.AccountIDs,
		ConfirmedAt:      a.ConfirmedAt.UTC(),
		SharingEndsAt:    a.SharingEndsAt().UTC(),
		Status:           statusOf(state, now),
		EndReason:        state.EndReason,
		WithdrawalMethod: state.WithdrawalMethod,
		NoticeStatus:     state.NoticeStatus(),
	}

	if view.NoticeStatus != "" {
		view.NoticeAttempts = make([]noticeAttempt, 0, len(state.NoticeAttempts))
		for _, a := range state.NoticeAttempts {
			view.NoticeAttempts = append(
				view.NoticeAttempts,
				noticeAttempt{Attempt: a.Attempt, At: a.At.UTC(), Result: a.Result},
			)
		}
	}

	if !state.EndedAt.IsZero() {
		view.EndedAt = new(state.EndedAt.UTC())
	}

	return view
}

func statusOf(state domain.ArrangementState, now time.Time) arrangementStatus {
	if state.Ended(now) {
		return arrangementEnded
	}

	return arrangementCurrent
}

func readTransactions(details CustomerDetails) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, pageSize, err := consolehttp.ParsePage(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		found, total, err := details.TransactionPage(
			r.Context(), r.PathValue("customerNumber"), r.PathValue("accountId"), page, pageSize,
		)
		if err != nil {
			writeDetailError(w, r, err)
			return
		}

		if found == nil {
			found = []domain.Transaction{}
		}

		jsonhttp.Write(w, http.StatusOK, transactionPage{Data: found, TotalRecords: total})
	})
}
