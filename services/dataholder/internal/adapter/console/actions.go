package console

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// AccountCloser gives app.ErrEndsChanged, app.ErrCustomerNotFound, app.ErrAccountNotFound and
// domain.ErrAccountClosed.
type AccountCloser interface {
	CloseAccountThatEnds(ctx context.Context, customerNumber, accountID string, ends []string) (domain.Account, error)
}

// ArrangementWithdrawer gives app.ErrCustomerNotFound, domain.ErrArrangementNotFound and domain.ErrArrangementEnded.
type ArrangementWithdrawer interface {
	WithdrawByConsumer(
		ctx context.Context,
		customerNumber, arrangementID string,
		method domain.WithdrawalMethod,
	) (domain.ArrangementState, error)
}

const (
	// A customer has few arrangements, and each ID is short.
	maxActionBodyBytes = 1 << 14
	badCloseBody       = "The request body must be a JSON object with only ends, the list of the IDs of the " +
		"arrangements that the confirmation of the close named."
	endsChanged = "The customer changed after the confirmation, and the close now ends different arrangements. " +
		"Look at the new state of the customer. The data holder changed nothing."
)

func closeAccount(accounts AccountCloser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ends, ok := decodeEnds(http.MaxBytesReader(w, r.Body, maxActionBodyBytes))
		if !ok {
			problem.Write(w, http.StatusBadRequest, badCloseBody)
			return
		}

		_, err := accounts.CloseAccountThatEnds(
			r.Context(), r.PathValue("customerNumber"), r.PathValue("accountId"), ends,
		)

		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, app.ErrEndsChanged):
			problem.Write(w, http.StatusConflict, endsChanged)
		case errors.Is(err, domain.ErrAccountClosed):
			problem.Write(w, http.StatusConflict, "The account is closed. The data holder changed nothing.")
		default:
			writeDetailError(w, r, err)
		}
	})
}

func decodeEnds(body io.Reader) ([]string, bool) {
	var request struct {
		Ends *[]string `json:"ends"`
	}

	if jsonhttp.Decode(body, &request) != nil || request.Ends == nil {
		return nil, false
	}

	return *request.Ends, true
}

// A withdrawal has no body, because it ends only the arrangement of the path.
func withdrawArrangement(arrangements ArrangementWithdrawer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := arrangements.WithdrawByConsumer(
			r.Context(), r.PathValue("customerNumber"), r.PathValue("arrangementId"),
			domain.WithdrawalMethodAlternative,
		)

		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, domain.ErrArrangementEnded):
			problem.Write(w, http.StatusConflict, "The arrangement ended. The data holder changed nothing.")
		case errors.Is(err, domain.ErrArrangementNotFound):
			problem.Write(w, http.StatusNotFound, "The customer has no arrangement with this arrangement ID.")
		default:
			writeDetailError(w, r, err)
		}
	})
}
