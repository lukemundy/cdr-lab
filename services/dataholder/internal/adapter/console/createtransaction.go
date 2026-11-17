package console

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const badTransactionBody = "The request body must be a JSON object with only the fields of the transaction."

// The console API has no GET for the Location of a transaction.
func createTransaction(creator Creator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A pointer is nil after the JSON null, which decodes into a struct with no error.
		var request *domain.TransactionInput

		err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxCreateBodyBytes), &request)
		if err != nil || request == nil {
			problem.WriteBodyRefusal(w, err, badTransactionBody)
			return
		}

		transaction, err := request.Transaction()
		if err != nil {
			writeCreateError(w, r, err)
			return
		}

		customerNumber, accountID := r.PathValue("customerNumber"), r.PathValue("accountId")

		created, err := creator.AddTransaction(r.Context(), customerNumber, accountID, transaction)
		if errors.Is(err, domain.ErrAccountClosed) {
			problem.Write(w, http.StatusConflict, "The account is closed. The data holder changed nothing.")
			return
		}

		if err != nil {
			writeCreateError(w, r, err)
			return
		}

		w.Header().Set("Location", "/api/customers/"+url.PathEscape(customerNumber)+"/accounts/"+
			url.PathEscape(accountID)+"/transactions/"+url.PathEscape(created.TransactionID))
		w.WriteHeader(http.StatusCreated)
	})
}
