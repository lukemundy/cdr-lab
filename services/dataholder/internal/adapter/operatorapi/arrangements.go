package operatorapi

import (
	"context"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Arrangements gives app.ErrCustomerNotFound, and for a withdrawal also domain.ErrArrangementNotFound and
// domain.ErrArrangementEnded.
type Arrangements interface {
	OfCustomer(ctx context.Context, customerNumber string) ([]domain.ArrangementState, error)
	WithdrawByConsumer(
		ctx context.Context,
		customerNumber, arrangementID string,
		method domain.WithdrawalMethod,
	) (domain.ArrangementState, error)
}

type arrangementJSON struct {
	ArrangementID    string     `json:"arrangementId"`
	SoftwareID       string     `json:"softwareId"`
	ClientName       string     `json:"clientName"`
	Scopes           []string   `json:"scopes"`
	AccountIDs       []string   `json:"accountIds"`
	ConfirmedAt      time.Time  `json:"confirmedAt"`
	SharingEndsAt    time.Time  `json:"sharingEndsAt"`
	EndedAt          *time.Time `json:"endedAt,omitempty"`
	EndReason        string     `json:"endReason,omitempty"`
	WithdrawalMethod string     `json:"withdrawalMethod,omitempty"`
	NoticeStatus     string     `json:"noticeStatus,omitempty"`
}

func arrangementView(state domain.ArrangementState) arrangementJSON {
	a := state.Arrangement
	view := arrangementJSON{
		ArrangementID:    a.ArrangementID,
		SoftwareID:       a.SoftwareID,
		ClientName:       a.ClientName,
		Scopes:           a.Scopes,
		AccountIDs:       a.AccountIDs,
		ConfirmedAt:      a.ConfirmedAt,
		SharingEndsAt:    a.SharingEndsAt(),
		EndReason:        string(state.EndReason),
		WithdrawalMethod: string(state.WithdrawalMethod()),
		NoticeStatus:     string(state.NoticeStatus()),
	}

	if !state.EndedAt.IsZero() {
		view.EndedAt = &state.EndedAt
	}

	return view
}

func readArrangements(arrangements Arrangements) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		states, err := arrangements.OfCustomer(r.Context(), r.PathValue("customerNumber"))
		if err != nil {
			writeError(w, r, err)
			return
		}

		views := make([]arrangementJSON, 0, len(states))
		for _, state := range states {
			views = append(views, arrangementView(state))
		}

		jsonhttp.Write(w, http.StatusOK, views)
	})
}

func withdrawArrangement(arrangements Arrangements) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		withdrawn, err := arrangements.WithdrawByConsumer(
			r.Context(), r.PathValue("customerNumber"), r.PathValue("arrangementId"),
			domain.WithdrawalMethodAlternative,
		)
		if err != nil {
			writeError(w, r, err)
			return
		}

		jsonhttp.Write(w, http.StatusOK, arrangementView(withdrawn))
	})
}
