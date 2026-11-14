package banksite

import (
	"context"
	"errors"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// ConsentWithdrawer gives app.ErrCustomerNotFound, domain.ErrArrangementNotFound for an arrangement of a different
// customer, and domain.ErrArrangementEnded.
type ConsentWithdrawer interface {
	WithdrawByConsumer(
		ctx context.Context,
		customerNumber, arrangementID string,
		method domain.WithdrawalMethod,
	) (domain.ArrangementState, error)
}

// withdrawConsent reads no body, because a withdrawal ends only the arrangement of the path.
func withdrawConsent(consents ConsentWithdrawer, s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		_, err := consents.WithdrawByConsumer(
			r.Context(), customerNumber, r.PathValue("arrangementId"), domain.WithdrawalMethodConsumerDashboard,
		)

		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, domain.ErrArrangementEnded):
			problem.Write(
				w,
				http.StatusConflict,
				"This consent changed after you opened this confirmation: it already ended. "+
					"We changed nothing.",
			)
		case errors.Is(err, domain.ErrArrangementNotFound):
			problem.Write(w, http.StatusNotFound, "You have no consent with this ID.")
		case errors.Is(err, app.ErrCustomerNotFound):
			s.endOf(r)
			s.refuse(w, r)
		default:
			problem.WriteUnexpected(w, r, err, notCompleted)
		}
	})
}
